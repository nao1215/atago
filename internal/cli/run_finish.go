package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/report"
)

func finishRun(ctx context.Context, opts *runOptions, suiteResults []*engine.SuiteResult, loadErrs []error, progress *report.Progress, elapsed time.Duration) int {
	if len(suiteResults) != len(opts.paths) || len(loadErrs) != len(opts.paths) {
		return failIncomplete(opts, progress)
	}
	results, exit, loadFailures, ok := collectSuiteExits(opts, suiteResults, loadErrs)
	if !ok {
		return failIncomplete(opts, progress)
	}
	if progress != nil {
		progress.Done()
	}

	exit = worseExit(exit, settleRerunLedger(ctx, opts, results, unreachedSpecs(opts.paths, suiteResults, loadErrs)))

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

	exit = worseExit(exit, emptySelectionExit(ctx, opts, results))

	if err := report.Render(opts.stdout, opts.format, results, report.WithLoadFailures(loadFailures), report.WithElapsed(elapsed), report.WithAllowFlaky(opts.allowFlaky), report.WithAllowXPass(opts.allowXPass)); err != nil {
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

// failIncomplete reports a result/path bookkeeping mismatch — a bug in atago,
// never in the spec — and returns the internal-error exit code.
func failIncomplete(opts *runOptions, progress *report.Progress) int {
	if progress != nil {
		progress.Done()
	}
	fmt.Fprintln(opts.stderr, opts.label+": internal error: incomplete run results")
	return ExitInternal
}

// collectSuiteExits pairs each spec path with its load error or suite result,
// printing load failures and folding every outcome into one exit code. ok is
// false when the slices run dry early (a bookkeeping bug the caller reports).
func collectSuiteExits(opts *runOptions, suiteResults []*engine.SuiteResult, loadErrs []error) (results []*engine.SuiteResult, exit, loadFailures int, ok bool) {
	results = make([]*engine.SuiteResult, 0, len(opts.paths))
	exit = ExitOK
	remainingResults := suiteResults
	remainingLoadErrs := loadErrs
	for range opts.paths {
		loadErr, nextLoadErrs, ok := shiftSlice(remainingLoadErrs)
		if !ok {
			return nil, 0, 0, false
		}
		suiteResult, nextResults, ok := shiftSlice(remainingResults)
		if !ok {
			return nil, 0, 0, false
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
		exit = worseExit(exit, exitForSuite(suiteResult, opts.allowFlaky, opts.allowXPass))
	}
	return results, exit, loadFailures, true
}

// settleRerunLedger updates the last-failed ledger for a later `--rerun-failed`
// (#64) and returns the exit contribution of a --rerun-failed that matched
// nothing. The preservation invariants live with the ledger primitives in
// rerun.go.
// unreachedSpecs are the spec paths the run never loaded: --fail-fast stopped
// scheduling after an earlier spec turned the run red, or an interrupt ended the
// run first. They produced neither a suite result nor a load error, so the run
// has nothing to say about the scenarios inside them.
func unreachedSpecs(paths []string, suiteResults []*engine.SuiteResult, loadErrs []error) []string {
	var out []string
	for i, p := range paths {
		if loadErrs[i] == nil && suiteResults[i] == nil {
			out = append(out, p)
		}
	}
	return out
}

func settleRerunLedger(ctx context.Context, opts *runOptions, results []*engine.SuiteResult, unreached []string) int {
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
	// contradicts the empty-selection warning). The excluded failures are still
	// preserved into the ledger via rerunPreserved, so no work is lost.
	if opts.rerunFailed && !opts.selectionActive() && len(results) > 0 && ranScenarios == 0 && ctx.Err() == nil {
		fmt.Fprintf(opts.stderr, "%s: %s\n", opts.label, diag.RerunNothingMatched.Annotate("no recorded failing scenarios matched the current specs (renamed or removed?); the recorded failures were kept, not cleared"))
		return ExitConfig
	}

	// The ledger is left untouched when no suite loaded (prior state stays
	// intact); the matched-nothing case above returned before reaching here so
	// its unmapped failures survive.
	if len(results) > 0 {
		// A --rerun-failed that matched only SOME of the recorded failures is the
		// quiet version of the case above: the run reports fewer scenarios than
		// were recorded, which reads as "the others are fixed". Say so before the
		// ledger is rewritten, while the prior entries are still readable.
		if opts.rerunFailed && !opts.selectionActive() {
			warnUnmatchedRerunEntries(opts.label, opts.stderr, results, unreached)
		}
		updateRerunLedger(opts.label, opts.stderr, results, opts.allowXPass)
	}
	return ExitOK
}

// emptySelectionExit handles a selection that matches nothing: interactively
// this still exits 0 (nothing ran, nothing failed) but stays loud; under --ci
// it is a hard config error so a typo'd --filter/--tag/--skip-tag cannot
// silently disable the whole suite in a pipeline forever.
func emptySelectionExit(ctx context.Context, opts *runOptions, results []*engine.SuiteResult) int {
	if !opts.selectionActive() {
		return ExitOK
	}
	total := 0
	for _, r := range results {
		total += len(r.Scenarios)
	}
	// total == 0 here can only mean the selectors excluded every scenario, never
	// that the specs were empty: the loader rejects a spec with no scenarios, and
	// a selected-but-skipped scenario (os gate, skip step) still appears in
	// res.Scenarios. So this is precisely the "selectors filtered everything"
	// case the task must fail on, not "the specs had nothing to run".
	if total > 0 || ctx.Err() != nil {
		return ExitOK
	}
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
		fmt.Fprintf(opts.stderr, "%s: %s\n", opts.label, diag.EmptySelection.Annotate(fmt.Sprintf("no scenarios matched %s under --ci; refusing to exit 0 (an empty selection would silently disable the suite). %s. Run `atago list` to see available scenarios and tags.", strings.Join(sel, " "), note)))
		return ExitConfig
	}
	hint := note
	if tagActive {
		hint += "; run `atago list` to see the available tags"
	}
	fmt.Fprintf(opts.stderr, opts.label+": warning: no scenarios matched %s (%s)\n", strings.Join(sel, " "), hint)
	return ExitOK
}

func shiftSlice[T any](values []T) (T, []T, bool) {
	var zero T
	if len(values) == 0 {
		return zero, nil, false
	}
	return values[0], values[1:], true
}

// suiteFailed reports whether a completed suite counts as a failure for
// --fail-fast across spec files: a security-policy violation, or any scenario
// whose outcome turns the run red. The suite's own Status is not enough — a
// flaky recovery and an XPASS both rank green there while failing the run — so
// the scenarios are consulted with the same predicate the exit code uses.
func suiteFailed(res *engine.SuiteResult, allowFlaky, allowXPass bool) bool {
	if res == nil {
		return false
	}
	if res.SecurityViolation {
		return true
	}
	for _, sc := range res.Scenarios {
		if engine.FailsRun(sc.Status, allowFlaky, allowXPass) {
			return true
		}
	}
	return false
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

// exitForSuite maps a suite result to an exit code. allowFlaky decides the one
// status whose severity is a policy rather than a fact: a scenario that needed a
// retry to pass, or that passed on some iterations and failed on others, has not
// been shown to work, so a suite holding one fails. --allow-flaky is for a suite
// with instability the caller already knows about and accepts — atago's own pty
// scenarios lose keystrokes when their sessions are starved of CPU — and it says
// so at the command line instead of every run quietly going green.
func exitForSuite(res *engine.SuiteResult, allowFlaky, allowXPass bool) int {
	// A security policy violation (e.g. a denied network host) takes precedence
	// over the generic execution-error code.
	if res.SecurityViolation {
		return ExitSecurity
	}
	switch res.Status {
	case engine.StatusPassed, engine.StatusSkipped, engine.StatusFlaky, engine.StatusXFail, engine.StatusXPass:
		// A flaky scenario and an XPASS do not raise the suite's status —
		// worseStatus ranks both alongside passed, so a suite of one flake and nine
		// passes still reports passed — which is why the scenarios are what decide
		// here. A flake is a scenario whose answer depended on how many times it
		// ran, and an XPASS is a fixed bug nobody promoted yet; both fail the run
		// unless the caller accepted them at the command line. engine.FailsRun is
		// the same predicate --fail-fast consults, so the flag stops for exactly
		// the outcomes that decide this exit code.
		for _, sc := range res.Scenarios {
			if engine.FailsRun(sc.Status, allowFlaky, allowXPass) {
				return ExitFailures
			}
		}
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
