package assert

import (
	"strings"

	"github.com/mattn/go-runewidth"

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
// visible in failure output. Width and padding are measured in DISPLAY COLUMNS,
// which is what the terminal the box is read in actually draws: a pty/TUI screen
// routinely carries box-drawing characters (─│┌, 3 bytes and 1 column each), CJK
// text and emoji (1 rune and 2 columns each), and measuring in bytes or in runes
// pads those rows to the wrong place — a right border that wanders, in exactly
// the screens this assertion exists to check.
func borderedScreen(screen string) string {
	lines := strings.Split(screen, "\n")
	width := 0
	for _, l := range lines {
		if n := displayWidth(l); n > width {
			width = n
		}
	}
	var b strings.Builder
	bar := "+" + strings.Repeat("-", width+2) + "+"
	b.WriteString(bar + "\n")
	for _, l := range lines {
		b.WriteString("| " + l + strings.Repeat(" ", width-displayWidth(l)) + " |\n")
	}
	b.WriteString(bar)
	return b.String()
}

// screenWidth measures display columns with EastAsianWidth off, deliberately
// rather than by taking the package default.
//
// go-runewidth reads the host locale at init and turns that flag ON under a CJK
// locale, which would make the width of a failure box depend on the developer's
// LANG — the same screen framed one way on a Japanese laptop and another in CI.
// A report atago prints has to read the same everywhere. Off is also the right
// answer: the flag governs the AMBIGUOUS-width characters, which is where box
// drawing lives (─│┌), and every terminal a TUI is drawn in gives those one
// column. CJK and emoji are Wide, not ambiguous, so they still count two.
var screenWidth = &runewidth.Condition{StrictEmojiNeutral: true}

// displayWidth reports how many terminal columns s occupies. A wide character
// (CJK, emoji) takes two, a combining mark takes none, and everything else takes
// one — the same table a terminal lays a line out with, so a box drawn to this
// measure closes where the text ends.
func displayWidth(s string) int {
	return screenWidth.StringWidth(s)
}
