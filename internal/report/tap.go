package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// writeTAP emits a Test Anything Protocol (TAP version 13) stream, the format
// several CI ecosystems and TAP consumers expect. One
// `ok`/`not ok` line per scenario across every suite, numbered from 1; failures
// and errors carry a YAML diagnostic block, and skips use the `# SKIP` directive.
// Rendered by Render (FormatTAP).
func writeTAP(w io.Writer, results []*engine.SuiteResult) error {
	var b strings.Builder
	total := 0
	for _, res := range results {
		total += len(res.Scenarios)
		// A suite that errored before any scenario ran (#7) still contributes a
		// failing point, so the plan is never a bare "1..0" for a non-zero exit.
		if suiteErroredWithoutScenarios(res) {
			total += len(suiteFailurePoints(res))
		}
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
			case engine.StatusFlaky:
				// A scenario that failed and then passed on retry (#29) is green
				// for the verdict, matching the exit code, console, gha, and
				// junit. Emit a passing point so a TAP consumer agrees, and keep
				// the recovery visible in the diagnostic rather than hiding it.
				fmt.Fprintf(&b, "ok %d - %s\n", n, name)
				writeTAPDiagnostic(&b, fmt.Sprintf("flaky: passed after %d attempts", sc.Attempts), detailText(sc))
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
		if suiteErroredWithoutScenarios(res) {
			for _, p := range suiteFailurePoints(res) {
				n++
				fmt.Fprintf(&b, "not ok %d - %s\n", n, tapDescription(res.Suite, p.name))
				writeTAPDiagnostic(&b, p.message, p.body)
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
			// The literal block is copied verbatim into the reader's YAML, so a
			// raw control byte from captured output (an ANSI escape, a bell)
			// would make the diagnostic unparseable. Tab survives; the rest
			// becomes U+FFFD.
			fmt.Fprintf(b, "    %s\n", sanitizeControlBytes(line))
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
	// A captured control byte (ANSI escape, bell, DEL) has no place on a TAP
	// line; fold it so the ok/not-ok description and any SKIP directive stay
	// clean. Newlines are already flattened above, so none survive here.
	s = sanitizeControlBytes(s)
	return strings.TrimSpace(s)
}
