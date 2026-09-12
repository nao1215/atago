package report

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/plural"
)

// flakyMessage renders the one-line reason a scenario is flaky, in a form that
// fits whichever knob produced the instability (#138): a --repeat run reports
// its flake rate ("2/10 iterations failed"), while a --retry-failed recovery
// reports its attempt count ("passed after 2 attempts"). Every format (tap, gha,
// junit) shares this so the wording never drifts between them.
func flakyMessage(sc *engine.ScenarioResult) string {
	if len(sc.Iterations) > 0 {
		total := len(sc.Iterations)
		return fmt.Sprintf("flaky: %d/%d iterations failed", total-sc.PassedIterations(), total)
	}
	return fmt.Sprintf("flaky: passed after %d attempts", sc.Attempts)
}

// expectFailSummary is the one-line "why is this a known bug" text every format
// puts next to an XFAIL or XPASS: the reason, and the issue when there is one.
// Shared for the same reason flakyMessage is — two lines describing the same
// scenario must not word it differently.
func expectFailSummary(sc *engine.ScenarioResult) string {
	if sc.ExpectFail == nil {
		return "expected failure"
	}
	if sc.ExpectFail.Issue != "" {
		return sc.ExpectFail.Reason + " (" + sc.ExpectFail.Issue + ")"
	}
	return sc.ExpectFail.Reason
}

// xpassMessage says what an XPASS means and what to do, in one line, for the
// formats that have room for one.
func xpassMessage(sc *engine.ScenarioResult) string {
	return "xpass: the known bug is fixed — drop expect_fail: and move this scenario into the guarded suite (" +
		expectFailSummary(sc) + ")"
}

// flakySuffix renders the ", N flaky" tail of a summary tally. Flaky scenarios
// (#29) are green for the verdict but never hidden, and the suffix appears only
// when non-zero so steady-state output is unchanged. Console and gha share this
// for the same reason they share flakyMessage: the wording must not drift
// between two lines describing the same run.
func flakySuffix(c engine.Counts) string {
	if c.Flaky == 0 {
		return ""
	}
	return fmt.Sprintf(", %d flaky", c.Flaky)
}

// expectFailSuffix names the expected-failure tallies in the summary line
// (#395). They are appended rather than folded into passed/failed because they
// answer a different question — "how many known bugs are still known, and how
// many are fixed" — and a reader who sees neither number cannot tell a suite
// documenting three live bugs from one documenting none.
func expectFailSuffix(c engine.Counts) string {
	out := ""
	if c.XFail > 0 {
		out += fmt.Sprintf(", %d xfail", c.XFail)
	}
	if c.XPass > 0 {
		out += fmt.Sprintf(", %d xpass", c.XPass)
	}
	return out
}

// Option configures an optional aspect of a Render call. It keeps the common
// three-argument form unchanged while letting callers pass extra run-level
// context (e.g. spec-load failures) without a signature churn.
type Option func(*renderOptions)

type renderOptions struct {
	// loadFailures are the spec files that failed to load (parse/schema errors)
	// before any scenario could run. Such files contribute to no suite in
	// results, so every format reports them separately and reads FAILED rather
	// than a misleading PASSED that contradicts the non-zero exit code (#120).
	loadFailures []LoadFailure
	// elapsed, when set, is the run's real wall-clock time. The console summary
	// prefers it over the sum of per-suite durations, which overcounts when
	// --parallel runs suites concurrently (4 one-second suites in parallel finish
	// in ~1s, not 4s).
	elapsed    time.Duration
	hasElapsed bool
	// allowFlaky mirrors --allow-flaky. A flaky scenario fails the run, so the
	// summary must read FAILED for one, or the headline contradicts the exit code
	// the same way a load failure used to. When the caller accepts the
	// instability, both go back to green together.
	allowFlaky bool
	// allowXPass mirrors --allow-xpass, for the same reason allowFlaky exists:
	// an XPASS fails the run by default, so the summary must read FAILED unless
	// the caller accepted it.
	allowXPass bool
	// snapshotsUpdated is how many golden files the run rewrote under
	// --update-snapshots, as counted by the engine's write recorder.
	snapshotsUpdated int
}

// LoadFailure is one spec file the run was given and could not read: the path
// as written, and the loader's diagnostic.
//
// The path and message travel with the report rather than only the count,
// because a machine format has to name the file to be worth reading. A count
// tells a dashboard that something is wrong; the path tells it what.
type LoadFailure struct {
	SpecPath string
	Message  string
}

// WithLoadFailures records the spec files that failed to load for this run, so
// every format reflects them instead of silently omitting them (#120).
func WithLoadFailures(fails ...LoadFailure) Option {
	return func(o *renderOptions) { o.loadFailures = fails }
}

// WithSnapshotsUpdated records how many golden files this run rewrote under
// --update-snapshots. Rewriting the committed goldens is the one passing
// outcome a reviewer has to be told about: a job carrying --update-snapshots by
// accident rewrites every expected result to whatever the code currently does
// and still reports green.
//
// The count comes from the engine's write recorder rather than from the
// reported results, because the results cannot answer the question. A walk over
// them missed a scenario's teardown and both suite lifecycle blocks, never saw
// the repeat/retry iterations that are not the surviving result — so a run that
// went red after rewriting a golden reported no rewrite at all — and counted one
// per matrix row where one file was written.
func WithSnapshotsUpdated(n int) Option {
	return func(o *renderOptions) { o.snapshotsUpdated = n }
}

// snapshotSuffix names the snapshot rewrites in a summary line, in the shape
// the flaky and load-failure tails already use.
func snapshotSuffix(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(", %s updated", plural.Count(n, "snapshot", "snapshots"))
}

// loadFailureSuffix renders the ", N specs failed to load" tail the console and
// gha summaries share. A spec that never parsed ran no scenario, so it cannot be
// folded into the passed/failed tally without lying about how much was tested —
// it gets its own count, in one wording shared by both summaries.
func loadFailureSuffix(n int) string {
	if n == 0 {
		return ""
	}
	specPlural := "specs"
	if n == 1 {
		specPlural = "spec"
	}
	return fmt.Sprintf(", %d %s failed to load", n, specPlural)
}

// WithElapsed supplies the run's real wall-clock duration so the console summary
// reports it instead of summing per-suite durations, which overcounts under
// concurrent (--parallel) suites.
func WithElapsed(d time.Duration) Option {
	return func(o *renderOptions) { o.elapsed = d; o.hasElapsed = true }
}

// WithAllowFlaky records that the caller accepts flakiness for this run, so the
// summary reads PASSED for a flaky-only run and matches its zero exit code.
func WithAllowFlaky(allow bool) Option {
	return func(o *renderOptions) { o.allowFlaky = allow }
}

// WithAllowXPass records that the caller accepts an expect_fail scenario that
// passed, so the summary reads PASSED for an xpass-only run and matches its
// zero exit code.
func WithAllowXPass(allow bool) Option {
	return func(o *renderOptions) { o.allowXPass = allow }
}

// Render writes one or more suite results in the requested format. Console
// renders each suite in turn; JSON emits one stable top-level document
// ({"schema_version","suites":[...]}) regardless of suite count (#43).
func Render(w io.Writer, f Format, results []*engine.SuiteResult, opts ...Option) error {
	var o renderOptions
	for _, opt := range opts {
		opt(&o)
	}
	switch f {
	case FormatConsole:
		var b strings.Builder
		color := isTTY(w)
		agg := engine.SumCounts(results)
		var total int
		var dur time.Duration
		hardFail := false
		for _, res := range results {
			for i := range res.Scenarios {
				writeDetail(&b, color, res.Suite, res.SpecPath, &res.Scenarios[i])
			}
			writeSuiteDetail(&b, color, res)
			writeRepeatRates(&b, color, res)
			writeFlaky(&b, color, res)
			total += len(res.Scenarios)
			dur += res.Duration
			// A suite that errored before producing any scenario row (#7) has
			// zero counts; force the summary verdict to FAILED regardless.
			if suiteErroredWithoutScenarios(res) {
				hardFail = true
			}
		}
		if o.hasElapsed {
			dur = o.elapsed
		}
		writeSummary(&b, color, agg, total, dur, hardFail, len(o.loadFailures), o.allowFlaky, o.allowXPass, o.snapshotsUpdated)
		_, err := io.WriteString(w, b.String())
		return err
	case FormatJSON:
		doc := jsonDocument{SchemaVersion: jsonSchemaVersion, Suites: make([]jsonReport, 0, len(results)), SnapshotsUpdated: o.snapshotsUpdated}
		for _, res := range results {
			doc.Suites = append(doc.Suites, buildJSON(res, o.allowXPass))
		}
		for _, lf := range o.loadFailures {
			// Forward slashes, like every other spec_path in the document: two
			// fields naming files must not disagree about the separator on Windows.
			doc.LoadFailures = append(doc.LoadFailures, jsonLoadFailure{SpecPath: filepath.ToSlash(lf.SpecPath), Error: lf.Message})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(doc)
	case FormatJUnit:
		return writeJUnit(w, buildJUnit(results, o.allowXPass, o.loadFailures))
	case FormatGHA:
		return writeGHA(w, results, o.allowXPass, o.loadFailures, o.snapshotsUpdated)
	case FormatTAP:
		return writeTAP(w, results, o.loadFailures, o.snapshotsUpdated)
	default:
		return fmt.Errorf("unknown report format %q", f)
	}
}
