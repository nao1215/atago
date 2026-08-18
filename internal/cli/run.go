package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/diag"
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
	// snapshotsUpdated is how many golden files the run rewrote under
	// --update-snapshots, read from the engine after the run. It reaches the
	// report from the recorder the writes go through rather than from the
	// results, which cannot see the writes a teardown, a suite lifecycle block,
	// or a non-surviving repeat/retry iteration made.
	snapshotsUpdated int
	rerunFailed      bool
	// rerunTargets are the spec paths the run was pointed at BEFORE
	// --rerun-failed narrowed them to the ones holding recorded failures. The
	// ledger warning needs the difference: an entry under a spec that was never
	// a target went unexecuted because of the user's own targeting, not because
	// the scenario disappeared.
	rerunTargets []string
	allowFlaky   bool
	allowXPass   bool
	profile      string
	verbose      bool
	ci           bool
	stdout       io.Writer
	stderr       io.Writer
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
	eng.AllowFlaky = opts.allowFlaky
	eng.AllowXPass = opts.allowXPass
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
		opts.rerunTargets = append([]string(nil), paths...)
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
	// Build the binary under test before anything runs (#393). It happens after
	// the signal context exists so Ctrl-C interrupts a long build too, and
	// before the first scenario because every scenario would otherwise be
	// testing a stale binary or none at all.
	scratch, scratchErr := os.MkdirTemp("", "atago-subject-")
	if scratchErr != nil {
		fmt.Fprintf(stderr, "%s: %v\n", label, scratchErr)
		return ExitInternal
	}
	defer func() { _ = os.RemoveAll(scratch) }()
	built, buildErr := buildSubjects(ctx, paths, opts.profile, scratch, stderr)
	if buildErr != nil {
		fmt.Fprintf(stderr, "%s: %v\n", label, buildErr)
		// A manifest that will not load is spec content, not an execution
		// failure, and it has to exit the same way it did before the build phase
		// existed — discovering the manifest earlier must not change what the
		// same broken file reports.
		var lerr *loader.Error
		if errors.As(buildErr, &lerr) {
			return exitForLoadError(lerr)
		}
		return ExitExec
	}
	if err := exposeSubjects(built); err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", label, err)
		return ExitInternal
	}

	start := time.Now()
	suiteResults, loadErrs := runSpecs(ctx, eng, paths)
	elapsed := time.Since(start)
	// Read the rewrite count from the engine that did the writing: a run that
	// ends red still has to report the goldens it replaced along the way.
	opts.snapshotsUpdated = eng.SnapshotsUpdated()

	return finishRun(ctx, opts, suiteResults, loadErrs, progress, elapsed)
}

// formatAlternatives renders the report formats as the "console|json|…" list
// the flag help and usage line show. Derived from report.AllFormats — the list
// used to be spelled here by hand and a format added to the report package
// would have been invisible in --help.
func formatAlternatives() string {
	return strings.Join(report.FormatNames(), "|")
}

// formatProse renders the report formats as the "console, json, …, or tap"
// prose the unknown-format diagnostic uses, from the same list.
func formatProse() string {
	names := report.FormatNames()
	if len(names) == 1 {
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + ", or " + names[len(names)-1]
}

// parseRunFlags parses and validates `atago run`'s flags into a runOptions. The
// bool return is true when parsing already decided the outcome (a --help, a bad
// flag, an unknown --report, no matching spec files, or a failed bounds check),
// in which case the int is the exit code to return immediately; otherwise it is
// ExitOK and the caller proceeds with the returned options.
func parseRunFlags(label string, args []string, stdout, stderr io.Writer) (*runOptions, int, bool) {
	fs := flag.NewFlagSet(label, flag.ContinueOnError)
	fs.SetOutput(stderr)
	reportFmt := fs.String("report", "console", "report format: "+formatAlternatives())
	updateSnapshots := fs.Bool("update-snapshots", false, "create or overwrite snapshot files instead of comparing")
	ci := fs.Bool("ci", false, "CI-safe defaults: deterministic, no color (sets NO_COLOR), secret masking")
	parallel := fs.Int("parallel", runtime.NumCPU(), "number of scenarios to run concurrently; scenarios are isolated, each in its own temp dir")
	failFast := fs.Bool("fail-fast", false, "stop scheduling new scenarios after the first outcome that fails the run (a failure, an error, an XPASS, or a flake unless allowed)")
	var filter csvFlag
	fs.Var(&filter, "filter", "run only scenarios whose name contains any of these comma-separated substrings (repeatable; OR semantics like --tag)")
	var tag csvFlag
	fs.Var(&tag, "tag", "run only scenarios with any of these tags, matched exactly (comma-separated and repeatable; OR semantics)")
	var skipTag csvFlag
	fs.Var(&skipTag, "skip-tag", "skip scenarios with any of these tags, matched exactly (comma-separated and repeatable)")
	artifactsDir := fs.String("artifacts-dir", "", "write deterministic failure artifacts (actual/expected payloads) under DIR for review tooling")
	rerunFailed := fs.Bool("rerun-failed", false, "run only the scenarios that failed on the previous run (recorded in .atago/last-failed.json)")
	repeat := fs.Int("repeat", 0, "run each selected scenario N times to surface flakiness; any failing iteration fails the run")
	retryFailed := fs.Int("retry-failed", 0, "retry failed scenarios up to N times; a recovered scenario is reported as flaky and still fails the run unless --allow-flaky")
	profile := fs.String("profile", "", "build the subject with the named `profile` from the directory manifest (e.g. a coverage-instrumented build)")
	allowFlaky := fs.Bool("allow-flaky", false, "exit 0 when the only problem is flakiness; for a suite whose instability is known and accepted")
	allowXPass := fs.Bool("allow-xpass", false, "exit 0 when an expect_fail scenario passed (XPASS); by default a fixed known bug fails the run so the spec gets promoted")
	verbose := fs.Bool("verbose", false, "trace every scenario as it finishes: commands, exit codes, captured output, and per-assertion verdicts — for passing scenarios too")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "Usage: atago run [--report "+formatAlternatives()+"] [--update-snapshots] [--parallel N] [--fail-fast] [--filter S] [--tag T] [--skip-tag T] [--rerun-failed] [--repeat N] [--retry-failed N] [--allow-flaky] [--allow-xpass] [--profile NAME] [--artifacts-dir DIR] [--verbose] [--ci] <path | dir>...\n  (directories are searched recursively)\n")
		fs.PrintDefaults()
	}
	operands, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			replayFlagOutput(err, stderr)
			return nil, ExitOK, true
		}
		reportFlagError(label, err, stderr)
		return nil, ExitConfig, true
	}
	if *ci {
		// Force deterministic, color-free output. Secret masking is always on.
		//nolint:forbidigo // Process-wide on purpose, and set once during flag parsing:
		// --ci has to reach the color decision in every package, before any scenario runs.
		_ = os.Setenv("NO_COLOR", "1")
	}

	format := report.Format(*reportFmt)
	if !format.Valid() {
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.BadOptionValue.Annotate(fmt.Sprintf("unknown --report %q (want %s)", *reportFmt, formatProse())))
		return nil, ExitConfig, true
	}

	paths, exitCode, ok := specTargets(label, operands, stderr)
	if !ok {
		return nil, exitCode, true
	}

	// --repeat and --retry-failed answer opposite questions (does it flake? /
	// keep CI green despite flakes) and would fight over the attempt loop. --repeat
	// only ACTIVATES at > 1 (a value < 2 is a documented no-op), so --repeat 1
	// changes nothing and must not be rejected alongside --retry-failed.
	if *repeat > 1 && *retryFailed > 0 {
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.OptionsExclusive.Annotate("--repeat and --retry-failed are mutually exclusive (repeat detects flakiness, retry-failed tolerates it)"))
		return nil, ExitConfig, true
	}
	if *repeat < 0 || *retryFailed < 0 {
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.OptionOutOfRange.Annotate("--repeat and --retry-failed must be >= 0"))
		return nil, ExitConfig, true
	}
	// A negative --parallel is a typo, not a request: the engine would clamp it to
	// sequential and exit 0, silently ignoring the mistake. Reject it with the same
	// config error as --repeat/--retry-failed for consistent bounds checking. Zero
	// is left to mean the default (like repeat/retry allow 0).
	if *parallel < 0 {
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.OptionOutOfRange.Annotate("--parallel must be >= 0"))
		return nil, ExitConfig, true
	}
	if strings.TrimSpace(*artifactsDir) != "" {
		// Fail fast if the artifacts dir cannot be used. An existing regular file at
		// the path, or a directory that cannot be created, would otherwise make
		// every artifact write fail silently, leaving the user to believe a run
		// produced no reviewable failures when in fact none could be written.
		if err := ensureArtifactsDir(*artifactsDir); err != nil {
			fmt.Fprintf(stderr, "%s: %s\n", label, diag.OutputNotWritable.Annotate(fmt.Sprintf("--artifacts-dir %q is not usable: %v", *artifactsDir, err)))
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
		allowFlaky:      *allowFlaky,
		allowXPass:      *allowXPass,
		profile:         *profile,
		verbose:         *verbose,
		ci:              *ci,
		stdout:          stdout,
		stderr:          stderr,
	}, ExitOK, false
}

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
		if eng.FailFast && suiteFailed(suiteResults[i], eng.AllowFlaky, eng.AllowXPass) {
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
