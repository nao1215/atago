package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/report"
)

// runCmd implements `atago run`. label is the command name to name in error
// messages ("atago run", or "atago snapshot update" when snapshotCmd delegates
// here), so a diagnostic identifies the command the user actually invoked.
func runCmd(label string, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(label, flag.ContinueOnError)
	fs.SetOutput(stderr)
	reportFmt := fs.String("report", "console", "report format: console|json|junit|gha|tap")
	updateSnapshots := fs.Bool("update-snapshots", false, "create or overwrite snapshot files instead of comparing")
	ci := fs.Bool("ci", false, "CI-safe defaults: deterministic, no color (sets NO_COLOR), secret masking")
	parallel := fs.Int("parallel", runtime.NumCPU(), "number of scenarios to run concurrently; scenarios are isolated, each in its own temp dir")
	failFast := fs.Bool("fail-fast", false, "stop scheduling new scenarios after the first failure")
	var filter csvFlag
	fs.Var(&filter, "filter", "run only scenarios whose name contains any of these comma-separated substrings (repeatable; OR semantics like --tag)")
	var tag csvFlag
	fs.Var(&tag, "tag", "run only scenarios with any of these tags (comma-separated and repeatable; OR semantics)")
	var skipTag csvFlag
	fs.Var(&skipTag, "skip-tag", "skip scenarios with any of these tags (comma-separated and repeatable)")
	artifactsDir := fs.String("artifacts-dir", "", "write deterministic failure artifacts (actual/expected payloads) under DIR for review tooling")
	rerunFailed := fs.Bool("rerun-failed", false, "run only the scenarios that failed on the previous run (recorded in .atago/last-failed.json)")
	repeat := fs.Int("repeat", 0, "run each selected scenario N times to surface flakiness; any failing iteration fails the run")
	retryFailed := fs.Int("retry-failed", 0, "retry failed scenarios up to N times; recovered scenarios are reported as flaky, never hidden")
	verbose := fs.Bool("verbose", false, "trace every scenario as it finishes: commands, exit codes, captured output, and per-assertion verdicts — for passing scenarios too")
	fs.Usage = func() {
		fmt.Fprint(stderr, "Usage: atago run [--report console|json|junit|gha|tap] [--update-snapshots] [--parallel N] [--fail-fast] [--filter S] [--tag T] [--skip-tag T] [--rerun-failed] [--repeat N] [--retry-failed N] [--artifacts-dir DIR] [--verbose] [--ci] <path | dir>...\n  (directories are searched recursively)\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitConfig
	}
	if *ci {
		// Force deterministic, color-free output. Secret masking is always on.
		_ = os.Setenv("NO_COLOR", "1")
	}

	format := report.Format(*reportFmt)
	if !format.Valid() {
		fmt.Fprintf(stderr, label+": unknown --report %q (want console, json, junit, gha, or tap)\n", *reportFmt)
		return ExitConfig
	}

	targets := fs.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}

	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, label+": %v\n", err)
		return ExitConfig
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, label+": no *.atago.yaml files found")
		return ExitConfig
	}

	// --repeat and --retry-failed answer opposite questions (does it flake? /
	// keep CI green despite flakes) and would fight over the attempt loop. --repeat
	// only ACTIVATES at > 1 (a value < 2 is a documented no-op), so --repeat 1
	// changes nothing and must not be rejected alongside --retry-failed.
	if *repeat > 1 && *retryFailed > 0 {
		fmt.Fprintln(stderr, label+": --repeat and --retry-failed are mutually exclusive (repeat detects flakiness, retry-failed tolerates it)")
		return ExitConfig
	}
	if *repeat < 0 || *retryFailed < 0 {
		fmt.Fprintln(stderr, label+": --repeat and --retry-failed must be >= 0")
		return ExitConfig
	}
	// A negative --parallel is a typo, not a request: the engine would clamp it to
	// sequential and exit 0, silently ignoring the mistake. Reject it with the same
	// config error as --repeat/--retry-failed for consistent bounds checking. Zero
	// is left to mean the default (like repeat/retry allow 0).
	if *parallel < 0 {
		fmt.Fprintln(stderr, label+": --parallel must be >= 0")
		return ExitConfig
	}

	eng := engine.New()
	eng.UpdateSnapshots = *updateSnapshots
	eng.Parallel = *parallel
	eng.FailFast = *failFast
	eng.Repeat = *repeat
	eng.RetryFailed = *retryFailed
	eng.FilterNames = filter
	eng.Tags = tag
	eng.SkipTags = skipTag
	if strings.TrimSpace(*artifactsDir) != "" {
		eng.Artifacts = artifact.NewDir(*artifactsDir)
	}

	// --rerun-failed restricts this run to the scenarios recorded as failing on
	// the previous run (#64). It intersects the recorded spec paths with the
	// collected targets so the usual path semantics still apply, and installs an
	// identity selector so only the recorded scenarios execute. With nothing
	// recorded there is nothing to rerun, which is reported and treated as success.
	//
	// rerunPreserved holds recorded failures this rerun did NOT execute, which must
	// be carried back into the saved ledger below. A rerun skips a recorded failure
	// two ways: its spec is outside this run's targets (a narrowed
	// `--rerun-failed a.atago.yaml` when b.atago.yaml also had recorded failures),
	// or an active --filter/--tag/--skip-tag excludes it. Either way the scenario
	// was not re-verified and is still failing, so overwriting the ledger with only
	// what ran would forget still-failing work and could greenlight the loop. It is
	// computed after the run from what actually executed, which covers both cases
	// uniformly (tag exclusion cannot be predicted from the ledger, which stores
	// only names). recordedFailures carries the loaded ledger to that computation.
	var rerunPreserved []failedEntry
	var recordedFailures []failedEntry
	if *rerunFailed {
		state, lerr := loadRerunState()
		if lerr != nil {
			fmt.Fprintf(stderr, label+": cannot read %s: %v\n", rerunStatePath(), lerr)
			return ExitConfig
		}
		// Absolutize the recorded spec paths and the run targets so a spec matches
		// regardless of how its path is spelled between the recording run and the
		// rerun (relative vs absolute). Without this, a rerun addressed by an
		// equivalent-but-different spelling finds "nothing" and silently greenlights
		// while the failures are still real.
		for i := range state.Failed {
			state.Failed[i].SpecPath = absClean(state.Failed[i].SpecPath)
		}
		for i := range paths {
			paths[i] = absClean(paths[i])
		}
		sel := state.selectSet()
		if len(sel) == 0 {
			fmt.Fprintln(stderr, label+": no previously failed scenarios recorded; nothing to rerun")
			return ExitOK
		}
		paths = intersectPaths(paths, state.specPaths())
		if len(paths) == 0 {
			fmt.Fprintln(stderr, label+": no previously failed scenarios under the given targets")
			return ExitOK
		}
		recordedFailures = state.Failed
		eng.Select = sel
	}

	// In console mode, stream a live dot per scenario as it finishes, so a run
	// visibly zips along. JSON output stays pure (no dots on stdout).
	// --verbose (#6) replaces the dots with a full per-scenario trace; with a
	// machine report the trace goes to stderr so stdout stays machine-readable.
	var progress *report.Progress
	switch {
	case *verbose && format == report.FormatConsole:
		eng.OnScenario = report.NewVerbose(stdout).Scenario
	case *verbose:
		eng.OnScenario = report.NewVerbose(stderr).Scenario
	case format == report.FormatConsole:
		progress = report.NewProgress(stdout)
		eng.OnScenario = progress.Scenario
	}

	// Cancel the whole run on Ctrl-C / SIGTERM. NotifyContext restores the default
	// signal disposition on the second signal, so an unresponsive run can still be
	// force-killed. The context threads into every scenario and runner so an
	// interrupt stops scheduling new work and unwinds in-flight cleanup promptly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// With --parallel > 1, run suites concurrently too, sharing one global
	// semaphore so the TOTAL number of in-flight scenarios across every suite is
	// capped at N. This parallelizes both many-small-suites and few-large-suites
	// runs. Results are reassembled in input order for a deterministic report.
	if *parallel > 1 {
		eng.Sem = make(chan struct{}, *parallel)
	}
	start := time.Now()
	suiteResults, loadErrs := runSpecs(ctx, eng, paths)
	elapsed := time.Since(start)

	results := make([]*engine.SuiteResult, 0, len(paths))
	exit := ExitOK
	loadFailures := 0
	for i := range paths {
		if loadErrs[i] != nil {
			fmt.Fprintf(stderr, "%v\n", loadErrs[i])
			exit = worseExit(exit, exitForLoadError(loadErrs[i]))
			loadFailures++
			continue
		}
		// A nil result with no load error is a spec fail-fast (or an interrupt)
		// skipped before running: it contributes no scenarios, so omit it.
		if suiteResults[i] == nil {
			continue
		}
		results = append(results, suiteResults[i])
		exit = worseExit(exit, exitForSuite(suiteResults[i]))
	}
	if progress != nil {
		progress.Done()
	}

	// Scenarios that actually executed. A Select can exclude every scenario in a
	// loaded suite — most importantly a --rerun-failed whose recorded scenario
	// names no longer exist in the specs (renamed or removed while still broken).
	ranScenarios := 0
	for _, r := range results {
		ranScenarios += len(r.Scenarios)
	}
	// Carry back recorded failures this rerun did not execute (spec outside the
	// targets, or excluded by an active --filter/--tag/--skip-tag). They were not
	// re-verified and are still failing, so they stay in the ledger; only scenarios
	// that actually ran are re-decided below (recorded again if still failing,
	// dropped if fixed).
	if *rerunFailed {
		executed := make(map[string]bool)
		for _, r := range results {
			for _, sc := range r.Scenarios {
				executed[engine.ScenarioID(r.SpecPath, sc.Name)] = true
			}
		}
		for _, e := range recordedFailures {
			if !executed[engine.ScenarioID(e.SpecPath, e.Scenario)] {
				rerunPreserved = append(rerunPreserved, e)
			}
		}
	}
	// A --rerun-failed run that matched NOTHING verified nothing, yet the recorded
	// failures are still real: greenlighting it and clearing the state would
	// silently forget still-failing work. Warn loudly, keep the state, and do not
	// exit green. Require at least one suite to have LOADED, so this stays about a
	// scenario-name mismatch — when every spec fails to parse, the load errors
	// (already printed) are the real story, not a "renamed or removed" scenario.
	// An active --filter/--tag/--skip-tag is excluded here: when the user's own
	// selection is why nothing ran, blaming a rename/removal is wrong (and
	// contradicts the selection warning below). The excluded failures are still
	// preserved into the ledger via rerunPreserved above, so no work is lost.
	selectionActive := len(filter) > 0 || len(tag) > 0 || len(skipTag) > 0
	rerunMatchedNothing := *rerunFailed && !selectionActive && len(results) > 0 && ranScenarios == 0 && ctx.Err() == nil
	if rerunMatchedNothing {
		fmt.Fprintln(stderr, label+": warning: no recorded failing scenarios matched the current specs (renamed or removed?); the recorded failures were kept, not cleared")
		exit = worseExit(exit, ExitConfig)
	}

	// Record this run's failing scenarios for a later `--rerun-failed` (#64). The
	// state reflects exactly the scenarios that ran: failures are recorded and a
	// fully-green run clears the file. It is only rewritten when at least one suite
	// loaded, so a run where every spec failed to parse leaves prior state intact;
	// and a --rerun-failed that matched no scenario must NOT clear the file, or the
	// still-failing work it could not map would be forgotten. A narrowed
	// --rerun-failed carries back the recorded failures for specs outside its
	// target (rerunPreserved), which it did not re-verify, so a partial rerun
	// cannot silently drop still-failing work elsewhere in the ledger. Writing is
	// best-effort — a read-only checkout must not fail the run — so a write error
	// is a warning, not a fatal exit.
	if len(results) > 0 && !rerunMatchedNothing {
		if err := saveRerunState(append(collectFailures(results), rerunPreserved...)); err != nil {
			fmt.Fprintf(stderr, label+": could not update %s: %v\n", rerunStatePath(), err)
		}
	}

	// Every spec failed to load, or an interrupt skipped every suite before it
	// produced a result. Don't print a misleading "0 scenarios" report — but a run
	// that was interrupted before completing must never exit 0.
	if len(results) == 0 {
		if ctx.Err() != nil {
			fmt.Fprintln(stderr, label+": interrupted")
			return worseExit(exit, ExitExec)
		}
		return exit
	}

	// A selection that matches nothing still exits 0 (nothing ran, nothing
	// failed), but stay loud about it: a typo'd --filter/--tag in CI would
	// otherwise greenlight silently.
	if len(filter) > 0 || len(tag) > 0 || len(skipTag) > 0 {
		total := 0
		for _, r := range results {
			total += len(r.Scenarios)
		}
		if total == 0 && ctx.Err() == nil {
			var sel []string
			if len(filter) > 0 {
				sel = append(sel, fmt.Sprintf("--filter %q", strings.Join(filter, ",")))
			}
			if len(tag) > 0 {
				sel = append(sel, fmt.Sprintf("--tag %q", strings.Join(tag, ",")))
			}
			if len(skipTag) > 0 {
				sel = append(sel, fmt.Sprintf("--skip-tag %q", strings.Join(skipTag, ",")))
			}
			fmt.Fprintf(stderr, label+": warning: no scenarios matched %s (name matching is a case-sensitive substring)\n", strings.Join(sel, " "))
		}
	}

	if err := report.Render(stdout, format, results, report.WithLoadFailures(loadFailures), report.WithElapsed(elapsed)); err != nil {
		fmt.Fprintf(stderr, label+": failed to write report: %v\n", err)
		return worseExit(exit, ExitInternal)
	}
	// An interrupted run never reports success, even in the rare case where the
	// signal landed between scenarios and every scheduled one was skipped: the run
	// did not complete, so the exit code is at least an execution error (4).
	if ctx.Err() != nil {
		fmt.Fprintln(stderr, label+": interrupted")
		exit = worseExit(exit, ExitExec)
	}
	return exit
}

// runSpecs loads and executes every spec in paths under ctx, returning each
// suite's result (nil where the load failed) and the matching load error. It is
// the context-aware core of `atago run`: cancelling ctx (Ctrl-C / SIGTERM) stops
// scheduling new specs and propagates into each engine.Run so in-flight
// scenarios unwind promptly. Results stay in input order for a deterministic
// report regardless of concurrency.
func runSpecs(ctx context.Context, eng *engine.Engine, paths []string) ([]*engine.SuiteResult, []error) {
	suiteResults := make([]*engine.SuiteResult, len(paths))
	loadErrs := make([]error, len(paths))
	// failStop threads --fail-fast ACROSS spec files. The engine's own fail-fast
	// stops scenarios only within one suite; without this a failing first spec
	// would still let every later spec run. Once a suite fails, no new spec is
	// scheduled (specs already in flight under --parallel still finish).
	var failStop atomic.Bool
	runOne := func(i int, p string) {
		s, lerr := loader.Load(p)
		if lerr != nil {
			loadErrs[i] = lerr
			return
		}
		suiteResults[i] = eng.Run(ctx, s, p)
		if eng.FailFast && suiteFailed(suiteResults[i]) {
			failStop.Store(true)
		}
	}
	if eng.Sem != nil {
		// A fixed worker pool rather than one goroutine per spec: with a goroutine
		// per path, every spec is loaded the moment the run starts, so an interrupt
		// can no longer prevent any of that work. Workers are capped at the shared
		// scenario semaphore's capacity — more concurrently active suites than
		// semaphore slots cannot add throughput (each running scenario holds a
		// slot), they only front-load spec parsing. On Ctrl-C the dispatch loop
		// stops feeding, so the remaining specs are never loaded, let alone run.
		workers := cap(eng.Sem)
		if workers > len(paths) {
			workers = len(paths)
		}
		jobs := make(chan int)
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					runOne(i, paths[i])
				}
			}()
		}
		for i := range paths {
			if ctx.Err() != nil || failStop.Load() {
				break
			}
			select {
			case jobs <- i:
			case <-ctx.Done():
			}
		}
		close(jobs)
		wg.Wait()
	} else {
		for i, p := range paths {
			// Stop launching new suites once interrupted or --fail-fast has tripped;
			// the already-canceled ctx still flows into eng.Run so a partially-run
			// suite reports cleanly.
			if ctx.Err() != nil || failStop.Load() {
				break
			}
			runOne(i, p)
		}
	}
	return suiteResults, loadErrs
}

// suiteFailed reports whether a completed suite counts as a failure for
// --fail-fast: a failed or errored verdict, or a security-policy violation.
func suiteFailed(res *engine.SuiteResult) bool {
	return res != nil && (res.Status == engine.StatusFailed || res.Status == engine.StatusError || res.SecurityViolation)
}

// collectSpecFiles resolves run targets into a deduplicated list of spec files.
// A target may be a spec file or a directory; a directory is always searched
// recursively for *.atago.yaml files. atago targets every kind of CLI, so it
// avoids the Go-specific "dir/..." glob — a plain directory is enough. A trailing
// "..." is tolerated for convenience but is no longer required.
func collectSpecFiles(targets []string) ([]string, error) {
	seen := make(map[string]bool)
	var out []string
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}

	for _, t := range targets {
		t = strings.TrimSuffix(t, "...")
		t = strings.TrimSuffix(t, string(os.PathSeparator))
		if t == "" {
			t = "."
		}

		info, err := os.Stat(t)
		if err != nil {
			return nil, fmt.Errorf("cannot access %q: %w", t, err)
		}
		if !info.IsDir() {
			add(filepath.Clean(t))
			continue
		}
		if err := walkSpecDir(t, add); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// walkSpecDir recursively collects every *.atago.yaml file under dir.
func walkSpecDir(dir string, add func(string)) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if isSpecFile(path) {
			add(filepath.Clean(path))
		}
		return nil
	})
}

func isSpecFile(p string) bool {
	return strings.HasSuffix(p, ".atago.yaml") || strings.HasSuffix(p, ".atago.yml")
}

// csvFlag is a repeatable flag whose every occurrence is split on commas and
// accumulated, giving OR semantics across both forms: `--filter a,b` and
// `--filter a --filter b` both select names containing "a" or "b". This fixes
// the old single-string --filter, which treated a comma list as one literal
// substring and silently kept only the last of repeated flags (#119).
type csvFlag []string

func (c *csvFlag) String() string { return strings.Join(*c, ",") }

func (c *csvFlag) Set(v string) error {
	*c = append(*c, splitCSV(v)...)
	return nil
}

// splitCSV splits a comma-separated flag value into trimmed, non-empty tokens.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// exitForLoadError maps a spec load failure to an exit code. Both YAML-syntax
// (KindParse) and schema/semantic validation (KindValidation) errors are
// spec-content errors and exit 2; exit 3 is reserved for
// CLI-invocation problems (unknown command, bad flag, no files) handled by the
// caller. This is why a `db` runner missing its `dsn` exits 2, not 3 — the
// README documents 3 as CLI-invocation config, not spec content (issue #21).
func exitForLoadError(err error) int {
	// Both parse and validation load errors are spec-content errors → ExitParse.
	return ExitParse
}

func exitForSuite(res *engine.SuiteResult) int {
	// A security policy violation (e.g. a denied network host) takes precedence
	// over the generic execution-error code.
	if res.SecurityViolation {
		return ExitSecurity
	}
	switch res.Status {
	case engine.StatusPassed, engine.StatusSkipped, engine.StatusFlaky:
		return ExitOK
	case engine.StatusFailed:
		return ExitFailures
	case engine.StatusError:
		return ExitExec
	default:
		return ExitInternal
	}
}

// worseExit returns the more severe of two exit codes, preferring failure codes
// over success but treating exec/parse errors as most severe.
func worseExit(a, b int) int {
	severity := func(code int) int {
		switch code {
		case ExitOK:
			return 0
		case ExitFailures:
			return 1
		case ExitConfig:
			return 2
		case ExitParse:
			return 3
		case ExitExec:
			return 4
		case ExitSecurity:
			return 6
		default:
			return 5
		}
	}
	if severity(b) > severity(a) {
		return b
	}
	return a
}
