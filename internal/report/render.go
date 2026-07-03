package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/engine"
)

// Render writes one or more suite results in the requested format. Console
// renders each suite in turn; JSON emits one stable top-level document
// ({"schema_version","suites":[...]}) regardless of suite count (#43).
func Render(w io.Writer, f Format, results []*engine.SuiteResult) error {
	switch f {
	case FormatConsole:
		var b strings.Builder
		color := isTTY(w)
		var agg engine.Counts
		var total int
		var dur time.Duration
		for _, res := range results {
			for i := range res.Scenarios {
				writeDetail(&b, color, res.Suite, &res.Scenarios[i])
			}
			writeSuiteDetail(&b, color, res)
			c := res.Counts()
			agg.Passed += c.Passed
			agg.Failed += c.Failed
			agg.Errored += c.Errored
			agg.Skipped += c.Skipped
			total += len(res.Scenarios)
			dur += res.Duration
		}
		writeSummary(&b, color, agg, total, dur)
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
