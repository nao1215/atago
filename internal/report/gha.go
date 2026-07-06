package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// writeGHA emits GitHub Actions workflow-command annotations, so
// failures surface inline in the Actions UI. One `::error::` line per failed or
// errored scenario, plus a final `::notice::` summary. Rendered by Render
// (FormatGHA).
func writeGHA(w io.Writer, results []*engine.SuiteResult) error {
	var b strings.Builder
	var agg engine.Counts
	var total int
	for _, res := range results {
		for i := range res.Scenarios {
			sc := &res.Scenarios[i]
			switch sc.Status {
			case engine.StatusFailed:
				fmt.Fprintf(&b, "::error title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData(firstFailureMessage(sc)+" — "+oneLine(detailText(sc))))
			case engine.StatusError:
				fmt.Fprintf(&b, "::error title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData(firstErrorMessage(sc)))
			case engine.StatusFlaky:
				// Green for the job, loud in the annotations (#29, #138).
				fmt.Fprintf(&b, "::warning title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+sc.Name), ghaEscapeData(flakyMessage(sc)))
			}
		}
		// A suite that errored before any scenario ran (#7) surfaces its cause as
		// an error annotation, so the Actions UI is never silent for a non-zero
		// exit that produced no scenario rows.
		if suiteErroredWithoutScenarios(res) {
			for _, p := range suiteFailurePoints(res) {
				fmt.Fprintf(&b, "::error title=%s::%s\n",
					ghaEscapeProp(res.Suite+" / "+p.name), ghaEscapeData(p.message))
			}
		}
		c := res.Counts()
		agg.Passed += c.Passed
		agg.Failed += c.Failed
		agg.Errored += c.Errored
		agg.Skipped += c.Skipped
		agg.Flaky += c.Flaky
		total += len(res.Scenarios)
	}
	flaky := ""
	if agg.Flaky > 0 {
		flaky = fmt.Sprintf(", %d flaky", agg.Flaky)
	}
	fmt.Fprintf(&b, "::notice title=atago::%s\n", ghaEscapeData(fmt.Sprintf(
		"%d scenarios: %d passed, %d failed, %d errored, %d skipped%s",
		total, agg.Passed, agg.Failed, agg.Errored, agg.Skipped, flaky)))
	_, err := io.WriteString(w, b.String())
	return err
}

func oneLine(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\n", "; ")
}

// ghaEscapeData escapes a workflow-command message body per the GitHub spec.
// Beyond the required %/CR/LF encoding it folds any other raw control byte from
// captured output (an ANSI escape, a bell) to U+FFFD, so an annotation never
// carries a byte the Actions UI would mangle.
func ghaEscapeData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return sanitizeControlBytes(s)
}

// ghaEscapeProp escapes a workflow-command property value (stricter than data).
func ghaEscapeProp(s string) string {
	s = ghaEscapeData(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}
