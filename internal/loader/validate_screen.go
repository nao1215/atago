package loader

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// validateScreen checks a rendered-screen assertion: the stream matchers it
// shares with stdout/stderr, plus the attribute entries (#382).
func validateScreen(add func(string, ...any), where string, sa *spec.ScreenAssert) {
	if sa == nil {
		return
	}
	// A screen assert may consist of attribute entries ALONE — "the error line is
	// red" is a complete claim — so the stream half is only checked when the
	// author wrote one.
	// A count bound also belongs to the stream half, so an attrs-only assert that
	// carries one must still be checked — otherwise the bound is silently
	// ignored, which is worse than rejecting it.
	if len(sa.SetMatchers()) > 0 || sa.HasCount() || len(sa.Attrs) == 0 {
		validateStream(add, where, &sa.StreamAssert)
	}
	for i := range sa.Attrs {
		validateScreenAttr(add, fmt.Sprintf("%s.attrs[%d]", where, i), &sa.Attrs[i])
	}
}

func validateScreenAttr(add func(string, ...any), where string, a *spec.ScreenAttr) {
	if a.Text == "" {
		add("%s.text is required (the substring whose cells are checked)", where)
	}
	if !a.HasAttribute() {
		add("%s needs at least one of fg/bg/bold/italic/underline/reverse/blink; naming text alone asserts nothing", where)
	}
	if a.Row < 0 {
		add("%s.row must be a 1-based screen row", where)
	}
	validateScreenColor(add, where+".fg", a.FG)
	validateScreenColor(add, where+".bg", a.BG)
}

// validateScreenColor accepts an ANSI name, a 256-palette index, or `default`.
func validateScreenColor(add func(string, ...any), where, value string) {
	if value == "" {
		return
	}
	if spec.ValidScreenColor(value) {
		return
	}
	// A number outside the palette is the most likely mistake, so say so
	// specifically rather than listing every name.
	if n, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
		add("%s %d is outside the 256-color palette (0-255)", where, n)
		return
	}
	add("%s %q is not a color (use %s)", where, value, spec.ScreenColorNames())
}
