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
	"syscall"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/report"
)

// runCmd implements `atago run`.
func runCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	reportFmt := fs.String("report", "console", "report format: console|json|junit|gha|tap")
	updateSnapshots := fs.Bool("update-snapshots", false, "create or overwrite snapshot files instead of comparing")
	ci := fs.Bool("ci", false, "CI-safe defaults: deterministic, no color (sets NO_COLOR), secret masking")
	parallel := fs.Int("parallel", runtime.NumCPU(), "number of scenarios to run concurrently; scenarios are isolated, each in its own temp dir")
	failFast := fs.Bool("fail-fast", false, "stop scheduling new scenarios after the first failure")
	filter := fs.String("filter", "", "run only scenarios whose name contains this substring")
	tag := fs.String("tag", "", "run only scenarios with any of these comma-separated tags")
	skipTag := fs.String("skip-tag", "", "skip scenarios with any of these comma-separated tags")
	artifactsDir := fs.String("artifacts-dir", "", "write deterministic failure artifacts (actual/expected payloads) under DIR for review tooling")
	rerunFailed := fs.Bool("rerun-failed", false, "run only the scenarios that failed on the previous run (recorded in .atago/last-failed.json)")
	fs.Usage = func() {
		fmt.Fprint(stderr, "Usage: atago run [--report console|json|junit|gha|tap] [--update-snapshots] [--parallel N] [--fail-fast] [--filter S] [--tag T] [--skip-tag T] [--rerun-failed] [--artifacts-dir DIR] [--ci] <path | dir>...\n  (directories are searched recursively)\n")
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
		fmt.Fprintf(stderr, "atago run: unknown --report %q (want console, json, junit, gha, or tap)\n", *reportFmt)
		return ExitConfig
	}

	targets := fs.Args()
	if len(targets) == 0 {
		targets = []string{"."}
	}

	paths, err := collectSpecFiles(targets)
	if err != nil {
		fmt.Fprintf(stderr, "atago run: %v\n", err)
		return ExitConfig
	}
	if len(paths) == 0 {
		fmt.Fprintln(stderr, "atago run: no *.atago.yaml files found")
		return ExitConfig
	}

	eng := engine.New()
	eng.UpdateSnapshots = *updateSnapshots
	eng.Parallel = *parallel
	eng.FailFast = *failFast
	eng.FilterName = *filter
	eng.Tags = splitCSV(*tag)
	eng.SkipTags = splitCSV(*skipTag)
	if strings.TrimSpace(*artifactsDir) != "" {
		eng.Artifacts = artifact.NewDir(*artifactsDir)
	}

	// --rerun-failed restricts this run to the scenarios recorded as failing on
	// the previous run (#64). It intersects the recorded spec paths with the
	// collected targets so the usual path semantics still apply, and installs an
	// identity selector so only the recorded scenarios execute. With nothing
	// recorded there is nothing to rerun, which is reported and treated as success.
	if *rerunFailed {
		state, lerr := loadRerunState()
		if lerr != nil {
			fmt.Fprintf(stderr, "atago run: cannot read %s: %v\n", rerunStatePath(), lerr)
			return ExitConfig
		}
		sel := state.selectSet()
		if len(sel) == 0 {
			fmt.Fprintln(stderr, "atago run: no previously failed scenarios recorded; nothing to rerun")
			return ExitOK
		}
		paths = intersectPaths(paths, state.specPaths())
		if len(paths) == 0 {
			fmt.Fprintln(stderr, "atago run: no previously failed scenarios under the given targets")
			return ExitOK
		}
		eng.Select = sel
	}

	// In console mode, stream a live dot per scenario as it finishes, so a run
	// visibly zips along. JSON output stays pure (no dots on stdout).
	var progress *report.Progress
	if format == report.FormatConsole {
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
	suiteResults, loadErrs := runSpecs(ctx, eng, paths)

	results := make([]*engine.SuiteResult, 0, len(paths))
	exit := ExitOK
	for i := range paths {
		if loadErrs[i] != nil {
			fmt.Fprintf(stderr, "%v\n", loadErrs[i])
			exit = worseExit(exit, exitForLoadError(loadErrs[i]))
			continue
		}
		results = append(results, suiteResults[i])
		exit = worseExit(exit, exitForSuite(suiteResults[i]))
	}
	if progress != nil {
		progress.Done()
	}

	// Record this run's failing scenarios for a later `--rerun-failed` (#64). The
	// state reflects exactly the scenarios that ran: failures are recorded and a
	// fully-green run clears the file. It is only rewritten when at least one suite
	// loaded, so a run where every spec failed to parse leaves prior state intact.
	// Writing is best-effort — a read-only checkout must not fail the run — so a
	// write error is a warning, not a fatal exit.
	if len(results) > 0 {
		if err := saveRerunState(collectFailures(results)); err != nil {
			fmt.Fprintf(stderr, "atago run: could not update %s: %v\n", rerunStatePath(), err)
		}
	}

	// Every spec failed to load; the errors are already on stderr. Don't print a
	// misleading "0 scenarios" report.
	if len(results) == 0 {
		return exit
	}

	// A selection that matches nothing still exits 0 (nothing ran, nothing
	// failed), but stay loud about it: a typo'd --filter/--tag in CI would
	// otherwise greenlight silently.
	if *filter != "" || *tag != "" || *skipTag != "" {
		total := 0
		for _, r := range results {
			total += len(r.Scenarios)
		}
		if total == 0 && ctx.Err() == nil {
			var sel []string
			if *filter != "" {
				sel = append(sel, fmt.Sprintf("--filter %q", *filter))
			}
			if *tag != "" {
				sel = append(sel, fmt.Sprintf("--tag %q", *tag))
			}
			if *skipTag != "" {
				sel = append(sel, fmt.Sprintf("--skip-tag %q", *skipTag))
			}
			fmt.Fprintf(stderr, "atago run: warning: no scenarios matched %s (name matching is a case-sensitive substring)\n", strings.Join(sel, " "))
		}
	}

	if err := report.Render(stdout, format, results); err != nil {
		fmt.Fprintf(stderr, "atago run: failed to write report: %v\n", err)
		return worseExit(exit, ExitInternal)
	}
	// An interrupted run never reports success, even in the rare case where the
	// signal landed between scenarios and every scheduled one was skipped: the run
	// did not complete, so the exit code is at least an execution error (4).
	if ctx.Err() != nil {
		fmt.Fprintln(stderr, "atago run: interrupted")
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
	runOne := func(i int, p string) {
		s, lerr := loader.Load(p)
		if lerr != nil {
			loadErrs[i] = lerr
			return
		}
		suiteResults[i] = eng.Run(ctx, s, p)
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
			if ctx.Err() != nil {
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
			// Stop launching new suites once interrupted; the already-cancelled ctx
			// still flows into eng.Run so a partially-run suite reports cleanly.
			if ctx.Err() != nil {
				break
			}
			runOne(i, p)
		}
	}
	return suiteResults, loadErrs
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
	case engine.StatusPassed, engine.StatusSkipped:
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
