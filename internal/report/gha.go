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
			}
		}
		c := res.Counts()
		agg.Passed += c.Passed
		agg.Failed += c.Failed
		agg.Errored += c.Errored
		agg.Skipped += c.Skipped
		total += len(res.Scenarios)
	}
	fmt.Fprintf(&b, "::notice title=atago::%s\n", ghaEscapeData(fmt.Sprintf(
		"%d scenarios: %d passed, %d failed, %d errored, %d skipped",
		total, agg.Passed, agg.Failed, agg.Errored, agg.Skipped)))
	_, err := io.WriteString(w, b.String())
	return err
}

func oneLine(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\n", "; ")
}

// ghaEscapeData escapes a workflow-command message body per the GitHub spec.
func ghaEscapeData(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}

// ghaEscapeProp escapes a workflow-command property value (stricter than data).
func ghaEscapeProp(s string) string {
	s = ghaEscapeData(s)
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}
