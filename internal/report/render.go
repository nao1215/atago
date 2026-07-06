package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/engine"
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

// Option configures an optional aspect of a Render call. It keeps the common
// three-argument form unchanged while letting callers pass extra run-level
// context (e.g. spec-load failures) without a signature churn.
type Option func(*renderOptions)

type renderOptions struct {
	// loadFailures is the number of spec files that failed to load (parse/schema
	// errors) before any scenario could run. Such files contribute to no suite in
	// results, so the console summary reports them separately and reads FAILED
	// rather than a misleading PASSED that contradicts the non-zero exit code (#120).
	loadFailures int
	// elapsed, when set, is the run's real wall-clock time. The console summary
	// prefers it over the sum of per-suite durations, which overcounts when
	// --parallel runs suites concurrently (4 one-second suites in parallel finish
	// in ~1s, not 4s).
	elapsed    time.Duration
	hasElapsed bool
}

// WithLoadFailures records how many spec files failed to load for this run, so
// the summary can reflect them instead of silently omitting them (#120).
func WithLoadFailures(n int) Option {
	return func(o *renderOptions) { o.loadFailures = n }
}

// WithElapsed supplies the run's real wall-clock duration so the console summary
// reports it instead of summing per-suite durations, which overcounts under
// concurrent (--parallel) suites.
func WithElapsed(d time.Duration) Option {
	return func(o *renderOptions) { o.elapsed = d; o.hasElapsed = true }
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
		var agg engine.Counts
		var total int
		var dur time.Duration
		hardFail := false
		for _, res := range results {
			for i := range res.Scenarios {
				writeDetail(&b, color, res.Suite, &res.Scenarios[i])
			}
			writeSuiteDetail(&b, color, res)
			writeRepeatRates(&b, color, res)
			writeFlaky(&b, color, res)
			c := res.Counts()
			agg.Passed += c.Passed
			agg.Failed += c.Failed
			agg.Errored += c.Errored
			agg.Skipped += c.Skipped
			agg.Flaky += c.Flaky
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
		writeSummary(&b, color, agg, total, dur, hardFail, o.loadFailures)
		_, err := io.WriteString(w, b.String())
		return err
	case FormatJSON:
		doc := jsonDocument{SchemaVersion: jsonSchemaVersion, Suites: make([]jsonReport, 0, len(results))}
		for _, res := range results {
			doc.Suites = append(doc.Suites, buildJSON(res))
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(doc)
	case FormatJUnit:
		return writeJUnit(w, buildJUnit(results))
	case FormatGHA:
		return writeGHA(w, results)
	case FormatTAP:
		return writeTAP(w, results)
	default:
		return fmt.Errorf("unknown report format %q", f)
	}
}
