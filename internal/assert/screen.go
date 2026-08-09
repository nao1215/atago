package assert

import (
	"strings"
	"unicode/utf8"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkScreen evaluates a `screen` assertion (#27) against the rendered
// terminal screen a pty step produced. The matchers are the stream family;
// on failure the screen is shown in a bordered block so its width is
// unambiguous, and the full text flows to --artifacts-dir as a sidecar next
// to the raw transcript.
func checkScreen(sa *spec.ScreenAssert, res *runner.Result, env Env) *CheckResult {
	if res == nil || !res.IsPTY {
		return &CheckResult{Desc: "assert screen", Hint: "no pty step has run in this scenario yet (screen asserts render a pty step's terminal)"}
	}
	// An assert may be attributes ALONE ("the error line is red"), in which case
	// there is no stream matcher to run.
	cr := &CheckResult{OK: true, Desc: "assert screen"}
	if len(sa.SetMatchers()) > 0 {
		cr = checkStream("screen", &sa.StreamAssert, res.Screen, true, env)
	}
	if cr.OK && len(sa.Attrs) > 0 {
		cr = checkScreenAttrs(sa.Attrs, res)
	}
	if cr.OK {
		return cr
	}
	cr.Actual = borderedScreen(string(res.Screen))
	if cr.ArtifactKind == "" {
		cr.ArtifactKind = "screen"
		cr.ArtifactActual = res.Screen
	}
	return cr
}

// borderedScreen frames the rendered screen so trailing spaces and width are
// visible in failure output. Width and padding are measured in runes, not bytes:
// a pty/TUI screen routinely contains box-drawing characters (─│┌, 3 bytes each)
// and CJK text, and a byte-based measure would pad those rows short and produce a
// ragged right border in exactly the screens this assertion exists to check.
func borderedScreen(screen string) string {
	lines := strings.Split(screen, "\n")
	width := 0
	for _, l := range lines {
		if n := utf8.RuneCountInString(l); n > width {
			width = n
		}
	}
	var b strings.Builder
	bar := "+" + strings.Repeat("-", width+2) + "+"
	b.WriteString(bar + "\n")
	for _, l := range lines {
		b.WriteString("| " + l + strings.Repeat(" ", width-utf8.RuneCountInString(l)) + " |\n")
	}
	b.WriteString(bar)
	return b.String()
}
