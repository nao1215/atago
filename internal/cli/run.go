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

// runOptions is the validated result of parsing `atago run`'s flags: the plain
// configuration the run pipeline needs, with every ExitConfig-worthy check
// already done by parseRunFlags. Keeping it a plain struct lets finishRun's
// exit-code invariants be exercised by a unit test instead of only through full
// CLI invocations (#246).
type runOptions struct {
	label           string
	format          report.Format
	paths           []string
	updateSnapshots bool
	parallel        int
	failFast        bool
	repeat          int
	retryFailed     int
	filter          csvFlag
	tag             csvFlag
	skipTag         csvFlag
	artifactsDir    string
	rerunFailed     bool
	verbose         bool
	ci              bool
	stdout          io.Writer
	stderr          io.Writer
}

// selectionActive reports whether the user narrowed the run with a name/tag
// selector — the shared guard for the rerun-matched-nothing and
// selection-matched-nothing warnings.
func (o *runOptions) selectionActive() bool {
	return len(o.filter) > 0 || len(o.tag) > 0 || len(o.skipTag) > 0
}

// runCmd implements `atago run`. label is the command name to name in error
// messages ("atago run", or "atago snapshot update" when snapshotCmd delegates
// here), so a diagnostic identifies the command the user actually invoked. It is
// the ~40-line pipeline between parseRunFlags (flag parse + validation) and
// finishRun (post-run bookkeeping and the final exit code) (#246).
func runCmd(label string, args []string, stdout, stderr io.Writer) int {
	opts, exit, done := parseRunFlags(label, args, stdout, stderr)
	if done {
		return exit
	}

	eng := engine.New()
	eng.UpdateSnapshots = opts.updateSnapshots
	eng.Parallel = opts.parallel
	eng.FailFast = opts.failFast
	eng.Repeat = opts.repeat
	eng.RetryFailed = opts.retryFailed
	eng.FilterNames = opts.filter
	eng.Tags = opts.tag
	eng.SkipTags = opts.skipTag
	if strings.TrimSpace(opts.artifactsDir) != "" {
		eng.Artifacts = artifact.NewDir(opts.artifactsDir)
	}

	paths := opts.paths
	// --rerun-failed restricts this run to the scenarios recorded as failing on
	// the previous run (#64); the selection and canonicalization invariants live
	// with the ledger primitives in rerun.go.
	if opts.rerunFailed {
		narrowed, exitNow, done := applyRerunSelection(label, stderr, paths, eng)
		if done {
			return exitNow
		}
		paths = narrowed
	}
	opts.paths = paths

	// In console mode, stream a live dot per scenario as it finishes, so a run
	// visibly zips along. JSON output stays pure (no dots on stdout).
	// --verbose (#6) replaces the dots with a full per-scenario trace; with a
	// machine report the trace goes to stderr so stdout stays machine-readable.
	var progress *report.Progress
	switch {
	case opts.verbose && opts.format == report.FormatConsole:
		eng.OnScenario = report.NewVerbose(stdout).Scenario
	case opts.verbose:
		eng.OnScenario = report.NewVerbose(stderr).Scenario
	case opts.format == report.FormatConsole:
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
	if opts.parallel > 1 {
		eng.Sem = make(chan struct{}, opts.parallel)
	}
	start := time.Now()
	suiteResults, loadErrs := runSpecs(ctx, eng, paths)
	elapsed := time.Since(start)

	return finishRun(opts, suiteResults, loadErrs, progress, elapsed, ctx)
}

// parseRunFlags parses and validates `atago run`'s flags into a runOptions. The
// bool return is true when parsing already decided the outcome (a --help, a bad
// flag, an unknown --report, no matching spec files, or a failed bounds check),
// in which case the int is the exit code to return immediately; otherwise it is
// ExitOK and the caller proceeds with the returned options.
func parseRunFlags(label string, args []string, stdout, stderr io.Writer) (*runOptions, int, bool) {
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
	fs.Var(&tag, "tag", "run only scenarios with any of these tags, matched exactly (comma-separated and repeatable; OR semantics)")
	var skipTag csvFlag
	fs.Var(&skipTag, "skip-tag", "skip scenarios with any of these tags, matched exactly (comma-separated and repeatable)")
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
			return nil, ExitOK, true
		}
		return nil, ExitConfig, true
	}
	if *ci {
		// Force deterministic, color-free output. Secret masking is always on.
		_ = os.Setenv("NO_COLOR", "1")
	}

	format := report.Format(*reportFmt)
	if !format.Valid() {
		fmt.Fprintf(stderr, label+": unknown --report %q (want console, json, junit, gha, or tap)\n", *reportFmt)
		return nil, ExitConfig, true
	}

	targets := fs.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}

	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, label+": %v\n", err)
		return nil, ExitConfig, true
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, label+": no *.atago.yaml (or *.atago.yml) files found")
		return nil, ExitConfig, true
	}

	// --repeat and --retry-failed answer opposite questions (does it flake? /
	// keep CI green despite flakes) and would fight over the attempt loop. --repeat
	// only ACTIVATES at > 1 (a value < 2 is a documented no-op), so --repeat 1
	// changes nothing and must not be rejected alongside --retry-failed.
	if *repeat > 1 && *retryFailed > 0 {
		fmt.Fprintln(stderr, label+": --repeat and --retry-failed are mutually exclusive (repeat detects flakiness, retry-failed tolerates it)")
		return nil, ExitConfig, true
	}
	if *repeat < 0 || *retryFailed < 0 {
		fmt.Fprintln(stderr, label+": --repeat and --retry-failed must be >= 0")
		return nil, ExitConfig, true
	}
	// A negative --parallel is a typo, not a request: the engine would clamp it to
	// sequential and exit 0, silently ignoring the mistake. Reject it with the same
	// config error as --repeat/--retry-failed for consistent bounds checking. Zero
	// is left to mean the default (like repeat/retry allow 0).
	if *parallel < 0 {
		fmt.Fprintln(stderr, label+": --parallel must be >= 0")
		return nil, ExitConfig, true
	}
	if strings.TrimSpace(*artifactsDir) != "" {
		// Fail fast if the artifacts dir cannot be used. An existing regular file at
		// the path, or a directory that cannot be created, would otherwise make
		// every artifact write fail silently, leaving the user to believe a run
		// produced no reviewable failures when in fact none could be written.
		if err := ensureArtifactsDir(*artifactsDir); err != nil {
			fmt.Fprintf(stderr, label+": --artifacts-dir %q is not usable: %v\n", *artifactsDir, err)
			return nil, ExitConfig, true
		}
	}
	return &runOptions{
		label:           label,
		format:          format,
		paths:           paths,
		updateSnapshots: *updateSnapshots,
		parallel:        *parallel,
		failFast:        *failFast,
		repeat:          *repeat,
		retryFailed:     *retryFailed,
		filter:          filter,
		tag:             tag,
		skipTag:         skipTag,
		artifactsDir:    *artifactsDir,
		rerunFailed:     *rerunFailed,
		verbose:         *verbose,
		ci:              *ci,
		stdout:          stdout,
		stderr:          stderr,
	}, ExitOK, false
}

func finishRun(opts *runOptions, suiteResults []*engine.SuiteResult, loadErrs []error, progress *report.Progress, elapsed time.Duration, ctx context.Context) int {
	failIncomplete := func() int {
		if progress != nil {
			progress.Done()
		}
		fmt.Fprintln(opts.stderr, opts.label+": internal error: incomplete run results")
		return ExitInternal
	}

	if len(suiteResults) != len(opts.paths) || len(loadErrs) != len(opts.paths) {
		return failIncomplete()
	}

	results := make([]*engine.SuiteResult, 0, len(opts.paths))
	exit := ExitOK
	loadFailures := 0
	remainingResults := suiteResults
	remainingLoadErrs := loadErrs
	for range opts.paths {
		loadErr, nextLoadErrs, ok := shiftSlice(remainingLoadErrs)
		if !ok {
			return failIncomplete()
		}
		suiteResult, nextResults, ok := shiftSlice(remainingResults)
		if !ok {
			return failIncomplete()
		}
		remainingLoadErrs = nextLoadErrs
		remainingResults = nextResults
		if loadErr != nil {
			fmt.Fprintf(opts.stderr, "%v\n", loadErr)
			exit = worseExit(exit, exitForLoadError(loadErr))
			loadFailures++
			continue
		}
		// A nil result with no load error is a spec fail-fast (or an interrupt)
		// skipped before running: it contributes no scenarios, so omit it.
		if suiteResult == nil {
			continue
		}
		results = append(results, suiteResult)
		exit = worseExit(exit, exitForSuite(suiteResult))
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
	rerunMatchedNothing := opts.rerunFailed && !opts.selectionActive() && len(results) > 0 && ranScenarios == 0 && ctx.Err() == nil
	if rerunMatchedNothing {
		fmt.Fprintln(opts.stderr, opts.label+": warning: no recorded failing scenarios matched the current specs (renamed or removed?); the recorded failures were kept, not cleared")
		exit = worseExit(exit, ExitConfig)
	}

	// Update the last-failed ledger for a later `--rerun-failed` (#64); the
	// preservation invariants live with the ledger primitives in rerun.go. The
	// ledger is left untouched when no suite loaded (prior state stays intact)
	// and when a --rerun-failed matched nothing (its unmapped failures must
	// survive).
	if len(results) > 0 && !rerunMatchedNothing {
		updateRerunLedger(opts.label, opts.stderr, results, ranScenarios)
	}

	// Every spec failed to load, or an interrupt skipped every suite before it
	// produced a result. Don't print a misleading "0 scenarios" report — but a run
	// that was interrupted before completing must never exit 0.
	if len(results) == 0 {
		if ctx.Err() != nil {
			fmt.Fprintln(opts.stderr, opts.label+": interrupted")
			return worseExit(exit, ExitExec)
		}
		return exit
	}

	// A selection that matches nothing: interactively this still exits 0 (nothing
	// ran, nothing failed) but stays loud; under --ci it is a hard config error so
	// a typo'd --filter/--tag/--skip-tag cannot silently disable the whole suite in
	// a pipeline forever.
	if opts.selectionActive() {
		total := 0
		for _, r := range results {
			total += len(r.Scenarios)
		}
		// total == 0 here can only mean the selectors excluded every scenario, never
		// that the specs were empty: the loader rejects a spec with no scenarios, and
		// a selected-but-skipped scenario (os gate, skip step) still appears in
		// res.Scenarios. So this is precisely the "selectors filtered everything"
		// case the task must fail on, not "the specs had nothing to run".
		if total == 0 && ctx.Err() == nil {
			var sel []string
			if len(opts.filter) > 0 {
				sel = append(sel, fmt.Sprintf("--filter %q", strings.Join(opts.filter, ",")))
			}
			if len(opts.tag) > 0 {
				sel = append(sel, fmt.Sprintf("--tag %q", strings.Join(opts.tag, ",")))
			}
			if len(opts.skipTag) > 0 {
				sel = append(sel, fmt.Sprintf("--skip-tag %q", strings.Join(opts.skipTag, ",")))
			}
			tagActive := len(opts.tag) > 0 || len(opts.skipTag) > 0
			// The note is selector-aware: --filter matches names by case-sensitive
			// substring, but --tag/--skip-tag compare tags for EXACT equality
			// (engine.hasAnyTag uses ==). A single "substring" note for tags would send
			// users fixing the wrong thing, so name each selector's real rule.
			note := selectorNoMatchNote(len(opts.filter) > 0, tagActive)
			if opts.ci {
				fmt.Fprintf(opts.stderr, opts.label+": no scenarios matched %s under --ci; refusing to exit 0 (an empty selection would silently disable the suite). %s. Run `atago list` to see available scenarios and tags.\n", strings.Join(sel, " "), note)
				exit = worseExit(exit, ExitConfig)
			} else {
				hint := note
				if tagActive {
					hint += "; run `atago list` to see the available tags"
				}
				fmt.Fprintf(opts.stderr, opts.label+": warning: no scenarios matched %s (%s)\n", strings.Join(sel, " "), hint)
			}
		}
	}

	if err := report.Render(opts.stdout, opts.format, results, report.WithLoadFailures(loadFailures), report.WithElapsed(elapsed)); err != nil {
		fmt.Fprintf(opts.stderr, opts.label+": failed to write report: %v\n", err)
		return worseExit(exit, ExitInternal)
	}
	// An interrupted run never reports success, even in the rare case where the
	// signal landed between scenarios and every scheduled one was skipped: the run
	// did not complete, so the exit code is at least an execution error (4).
	if ctx.Err() != nil {
		fmt.Fprintln(opts.stderr, opts.label+": interrupted")
		exit = worseExit(exit, ExitExec)
	}
	return exit
}

func shiftSlice[T any](values []T) (T, []T, bool) {
	var zero T
	if len(values) == 0 {
		return zero, nil, false
	}
	return values[0], values[1:], true
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

// ensureArtifactsDir verifies --artifacts-dir names a usable directory, creating
// it when absent. It returns an error when the path exists as a non-directory or
// cannot be created, so run can report the problem up front instead of letting
// every later artifact write fail silently.
func ensureArtifactsDir(dir string) error {
	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("exists but is not a directory")
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(dir, 0o750)
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

// selectorNoMatchNote explains, per active selector, why a selection came up
// empty. --filter matches scenario NAMES by case-sensitive substring, whereas
// --tag/--skip-tag compare TAGS for exact equality (engine.hasAnyTag uses ==);
// conflating the two rules would point users at the wrong fix.
func selectorNoMatchNote(filterActive, tagActive bool) string {
	var parts []string
	if filterActive {
		parts = append(parts, "--filter matches scenario names by case-sensitive substring")
	}
	if tagActive {
		parts = append(parts, "--tag/--skip-tag match tags exactly")
	}
	return strings.Join(parts, "; ")
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
