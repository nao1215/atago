package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/engine"
)

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
}

// WithLoadFailures records how many spec files failed to load for this run, so
// the summary can reflect them instead of silently omitting them (#120).
func WithLoadFailures(n int) Option {
	return func(o *renderOptions) { o.loadFailures = n }
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
