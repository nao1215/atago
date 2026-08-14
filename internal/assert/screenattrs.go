package assert

import (
	"fmt"

	"github.com/rivo/uniseg"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkScreenAttrs evaluates the attribute entries of a screen assertion (#382):
// each entry must find at least one occurrence of its text whose every cell
// carries the demanded colors and styling.
//
// Position-free by default is a deliberate choice. A styling claim pinned to a
// coordinate breaks every time the layout shifts, which would make the feature
// cost more than it pays; an author who cares about the position says so with
// `row`. When several occurrences exist, the closest one is reported on failure,
// because "found it, drawn like this instead" is the message that leads to a fix.
func checkScreenAttrs(attrs []spec.ScreenAttr, res *runner.Result) *CheckResult {
	for i := range attrs {
		a := &attrs[i]
		if cr := checkOneScreenAttr(a, res); !cr.OK {
			return cr
		}
	}
	return &CheckResult{OK: true, Desc: "assert screen attrs"}
}

func checkOneScreenAttr(a *spec.ScreenAttr, res *runner.Result) *CheckResult {
	fail := func(hint, actual string) *CheckResult {
		return &CheckResult{
			Desc:           "assert screen shows " + a.Describe(),
			Expected:       a.Describe(),
			Actual:         actual,
			Hint:           hint,
			ArtifactKind:   "screen",
			ArtifactActual: res.Screen,
		}
	}

	rows := res.ScreenCells
	if a.Row > 0 {
		if a.Row > len(rows) {
			return fail(
				fmt.Sprintf("the screen has %d row(s), so row %d does not exist", len(rows), a.Row),
				borderedScreen(string(res.Screen)))
		}
		rows = rows[a.Row-1 : a.Row]
	}

	// Match by grapheme cluster, not by rune: the query text and the screen cells
	// each hold one cluster per cell, so a ZWJ emoji or a base-plus-combining
	// sequence lines up cell-for-cluster (#437).
	text := splitGraphemes(a.Text)
	if len(text) == 0 {
		return fail("the entry names no text to check", borderedScreen(string(res.Screen)))
	}

	found := false
	var nearest string
	for _, row := range rows {
		for start := 0; start+len(text) <= len(row); start++ {
			if !graphemesMatch(row[start:start+len(text)], text) {
				continue
			}
			found = true
			if mismatch := firstAttrMismatch(a, row[start:start+len(text)]); mismatch == "" {
				return &CheckResult{OK: true, Desc: "assert screen shows " + a.Describe()}
			} else if nearest == "" {
				nearest = mismatch
			}
		}
	}

	if !found {
		where := "on the screen"
		if a.Row > 0 {
			where = fmt.Sprintf("on row %d", a.Row)
		}
		return fail(
			fmt.Sprintf("the text %q does not appear %s, so there is nothing to check the styling of", a.Text, where),
			borderedScreen(string(res.Screen)))
	}
	return fail(
		fmt.Sprintf("%q is on screen but drawn differently: %s", a.Text, nearest),
		borderedScreen(string(res.Screen)))
}

// graphemesMatch reports whether the cells spell exactly want, one cluster per
// cell.
func graphemesMatch(cells []runner.ScreenCell, want []string) bool {
	for i, g := range want {
		if cells[i].Content != g {
			return false
		}
	}
	return true
}

// splitGraphemes breaks s into grapheme clusters, the same unit a screen cell
// holds, so a query naming a ZWJ emoji or a combining sequence matches one cell
// rather than being compared rune-by-rune against a single cell.
func splitGraphemes(s string) []string {
	var out []string
	state := -1
	for len(s) > 0 {
		var cluster string
		cluster, s, _, state = uniseg.FirstGraphemeClusterInString(s, state)
		out = append(out, cluster)
	}
	return out
}

// firstAttrMismatch returns a description of the first way the cells fail the
// entry, or "" when every cell satisfies it. EVERY cell must match: half-styled
// text is a real bug (a color reset landing mid-word), and accepting it would
// let exactly that through.
func firstAttrMismatch(a *spec.ScreenAttr, cells []runner.ScreenCell) string {
	if a.FG != "" {
		if want, ok := spec.ParseScreenColor(a.FG); ok {
			for _, c := range cells {
				if !colorMatches(c.FG, want, c.Bold) {
					return fmt.Sprintf("fg=%s (wanted %s)", spec.DescribeScreenColor(c.FG), a.FG)
				}
			}
		}
	}
	if a.BG != "" {
		if want, ok := spec.ParseScreenColor(a.BG); ok {
			for _, c := range cells {
				// Background is never brightened by bold; that rule is a
				// foreground one.
				if !colorMatches(c.BG, want, false) {
					return fmt.Sprintf("bg=%s (wanted %s)", spec.DescribeScreenColor(c.BG), a.BG)
				}
			}
		}
	}
	for _, want := range []struct {
		flag *bool
		name string
		get  func(runner.ScreenCell) bool
	}{
		{a.Bold, "bold", func(c runner.ScreenCell) bool { return c.Bold }},
		{a.Italic, "italic", func(c runner.ScreenCell) bool { return c.Italic }},
		{a.Underline, "underline", func(c runner.ScreenCell) bool { return c.Underline }},
		{a.Reverse, "reverse", func(c runner.ScreenCell) bool { return c.Reverse }},
		{a.Blink, "blink", func(c runner.ScreenCell) bool { return c.Blink }},
	} {
		if want.flag == nil {
			continue
		}
		for _, c := range cells {
			if got := want.get(c); got != *want.flag {
				return fmt.Sprintf("%s=%v (wanted %v)", want.name, got, *want.flag)
			}
		}
	}
	return ""
}

// colorMatches compares an observed color to a wanted one.
//
// `default` matches both of the emulator's two default values (foreground and
// background), so a spec never has to know which slot it is asking about.
//
// Bold brightens: a terminal draws bold text in a color's bright variant, and
// the emulator mirrors that, so SGR 31 on bold text comes back as index 9 rather
// than 1. An author who wrote `fg: red` next to `bold: true` meant the 31 their
// program emitted, so the bright twin of an ANSI color satisfies it. Asking for
// `bright-red` specifically is still exact — the widening only runs from the
// dim name toward the bright one.
func colorMatches(got, want uint32, bold bool) bool {
	if want == spec.DefaultScreenColor {
		return got == spec.DefaultScreenColor || got == spec.DefaultScreenColor+1
	}
	if got == want {
		return true
	}
	return bold && want < 8 && got == want+8
}
