package assert

import (
	"strings"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkScreen evaluates a `screen` assertion (#27) against the rendered
// terminal screen a pty step produced. The matchers are the stream family;
// on failure the screen is shown in a bordered block so its width is
// unambiguous, and the full text flows to --artifacts-dir as a sidecar next
// to the raw transcript.
func checkScreen(sa *spec.StreamAssert, res *runner.Result, env Env) *CheckResult {
	if res == nil || !res.IsPTY {
		return &CheckResult{Desc: "assert screen", Hint: "no pty step has run in this scenario yet (screen asserts render a pty step's terminal)"}
	}
	cr := checkStream("screen", sa, res.Screen, true, env)
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
// visible in failure output.
func borderedScreen(screen string) string {
	lines := strings.Split(screen, "\n")
	width := 0
	for _, l := range lines {
		if len(l) > width {
			width = len(l)
		}
	}
	var b strings.Builder
	bar := "+" + strings.Repeat("-", width+2) + "+"
	b.WriteString(bar + "\n")
	for _, l := range lines {
		b.WriteString("| " + l + strings.Repeat(" ", width-len(l)) + " |\n")
	}
	b.WriteString(bar)
	return b.String()
}
