package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// writeTAP emits a Test Anything Protocol (TAP version 13) stream, the format
// several CI ecosystems and TAP consumers expect (ADR-0024). One
// `ok`/`not ok` line per scenario across every suite, numbered from 1; failures
// and errors carry a YAML diagnostic block, and skips use the `# SKIP` directive.
// Rendered by Render (FormatTAP).
func writeTAP(w io.Writer, results []*engine.SuiteResult) error {
	var b strings.Builder
	total := 0
	for _, res := range results {
		total += len(res.Scenarios)
	}
	b.WriteString("TAP version 13\n")
	fmt.Fprintf(&b, "1..%d\n", total)

	n := 0
	for _, res := range results {
		for i := range res.Scenarios {
			sc := &res.Scenarios[i]
			n++
			name := tapDescription(res.Suite, sc.Name)
			switch sc.Status {
			case engine.StatusPassed:
				fmt.Fprintf(&b, "ok %d - %s\n", n, name)
			case engine.StatusSkipped:
				fmt.Fprintf(&b, "ok %d - %s # SKIP %s\n", n, name, tapInline(sc.SkipReason))
			case engine.StatusFailed:
				fmt.Fprintf(&b, "not ok %d - %s\n", n, name)
				writeTAPDiagnostic(&b, firstFailureMessage(sc), detailText(sc))
			case engine.StatusError:
				fmt.Fprintf(&b, "not ok %d - %s\n", n, name)
				writeTAPDiagnostic(&b, firstErrorMessage(sc), detailText(sc))
			default:
				fmt.Fprintf(&b, "not ok %d - %s\n", n, name)
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// writeTAPDiagnostic emits the YAML block a TAP 13 consumer reads for a failed
// point. The message and body are indented and any `...` terminator inside the
// body is defanged so it cannot close the block early.
func writeTAPDiagnostic(b *strings.Builder, message, body string) {
	b.WriteString("  ---\n")
	fmt.Fprintf(b, "  message: %q\n", tapInline(message))
	if body != "" {
		b.WriteString("  data: |\n")
		for _, line := range strings.Split(body, "\n") {
			fmt.Fprintf(b, "    %s\n", line)
		}
	}
	b.WriteString("  ...\n")
}

// tapDescription builds a test point description, escaping the `#` that would
// otherwise start a TAP directive.
func tapDescription(suite, name string) string {
	return tapInline(suite + " / " + name)
}

// tapInline flattens a string to a single line and escapes `#` so it cannot be
// misread as the start of a TAP directive.
func tapInline(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "#", "\\#")
	return strings.TrimSpace(s)
}
