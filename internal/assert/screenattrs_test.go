package assert

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// cellsFor builds a one-row screen from a string, applying style to the cells
// covering want (its first occurrence) and leaving the rest default.
func cellsFor(text, want string, style func(*runner.ScreenCell)) *runner.Result {
	runes := []rune(text)
	row := make([]runner.ScreenCell, len(runes))
	start := strings.Index(text, want)
	for i, r := range runes {
		row[i] = runner.ScreenCell{Rune: r, FG: spec.DefaultScreenColor, BG: spec.DefaultScreenColor + 1}
	}
	if start >= 0 && want != "" {
		from := len([]rune(text[:start]))
		for i := from; i < from+len([]rune(want)); i++ {
			style(&row[i])
		}
	}
	return &runner.Result{IsPTY: true, Screen: []byte(text), ScreenCells: [][]runner.ScreenCell{row}}
}

// TestCheckScreenAttrs covers the matcher's contract (#382): colors and
// attributes, `default` as a first-class answer, and the failures that must be
// reported as failures rather than passed over.
func TestCheckScreenAttrs(t *testing.T) {
	t.Parallel()
	red := func(c *runner.ScreenCell) { c.FG = 1; c.Bold = true }

	tests := map[string]struct {
		res      *runner.Result
		attr     spec.ScreenAttr
		wantOK   bool
		wantHint string
	}{
		"color and attribute both hold": {
			res:    cellsFor("ERROR: nope", "ERROR", red),
			attr:   spec.ScreenAttr{Text: "ERROR", FG: "red", Bold: boolp(true)},
			wantOK: true,
		},
		"wrong color is reported with what was found": {
			res:      cellsFor("ERROR: nope", "ERROR", red),
			attr:     spec.ScreenAttr{Text: "ERROR", FG: "green"},
			wantHint: "fg=red (wanted green)",
		},
		// `bold: false` is a claim, not the absence of one.
		"negative attribute claim": {
			res:      cellsFor("ERROR: nope", "ERROR", red),
			attr:     spec.ScreenAttr{Text: "ERROR", Bold: boolp(false)},
			wantHint: "bold=true (wanted false)",
		},
		// The point of `fg: default`: proving --no-color really did leave the
		// frame uncolored.
		"default color matches uncolored text": {
			res:    cellsFor("plain text", "plain", func(*runner.ScreenCell) {}),
			attr:   spec.ScreenAttr{Text: "plain", FG: "default"},
			wantOK: true,
		},
		"text that is not on screen says so": {
			res:      cellsFor("ERROR: nope", "ERROR", red),
			attr:     spec.ScreenAttr{Text: "MISSING", FG: "red"},
			wantHint: "does not appear on the screen",
		},
		"a row that does not exist says so": {
			res:      cellsFor("ERROR", "ERROR", red),
			attr:     spec.ScreenAttr{Text: "ERROR", Row: 9, FG: "red"},
			wantHint: "row 9 does not exist",
		},
		"row restriction is honored": {
			res:    cellsFor("ERROR", "ERROR", red),
			attr:   spec.ScreenAttr{Text: "ERROR", Row: 1, FG: "red"},
			wantOK: true,
		},
		"256-palette index": {
			res:    cellsFor("styled", "styled", func(c *runner.ScreenCell) { c.FG = 203 }),
			attr:   spec.ScreenAttr{Text: "styled", FG: "203"},
			wantOK: true,
		},
		"reverse video, how a TUI draws the selected row": {
			res:    cellsFor("> README.md", "README.md", func(c *runner.ScreenCell) { c.Reverse = true }),
			attr:   spec.ScreenAttr{Text: "README.md", Reverse: boolp(true)},
			wantOK: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cr := checkScreenAttrs([]spec.ScreenAttr{tt.attr}, tt.res)
			if cr.OK != tt.wantOK {
				t.Fatalf("OK = %v, want %v (hint %q)", cr.OK, tt.wantOK, cr.Hint)
			}
			if tt.wantHint != "" && !strings.Contains(cr.Hint, tt.wantHint) {
				t.Errorf("hint = %q, want it to contain %q", cr.Hint, tt.wantHint)
			}
		})
	}
}

// TestCheckScreenAttrs_BoldBrightensTheColor pins the terminal rule that would
// otherwise make the obvious spec fail. A terminal draws bold text in a color's
// BRIGHT variant, and the emulator mirrors it, so `SGR 31` on bold text comes
// back as index 9 rather than 1. An author who wrote `fg: red` beside
// `bold: true` meant the 31 their program emitted, so the bright twin satisfies
// it — while asking for `bright-red` stays exact.
func TestCheckScreenAttrs_BoldBrightensTheColor(t *testing.T) {
	t.Parallel()
	boldRed := cellsFor("ERROR", "ERROR", func(c *runner.ScreenCell) { c.FG = 9; c.Bold = true })
	if cr := checkScreenAttrs([]spec.ScreenAttr{{Text: "ERROR", FG: "red", Bold: boolp(true)}}, boldRed); !cr.OK {
		t.Errorf("bold red should satisfy fg: red: %s", cr.Hint)
	}
	if cr := checkScreenAttrs([]spec.ScreenAttr{{Text: "ERROR", FG: "bright-red"}}, boldRed); !cr.OK {
		t.Errorf("the bright name should also match: %s", cr.Hint)
	}
	// The widening runs one way only: dim text is not bright text.
	dimRed := cellsFor("ERROR", "ERROR", func(c *runner.ScreenCell) { c.FG = 1 })
	if cr := checkScreenAttrs([]spec.ScreenAttr{{Text: "ERROR", FG: "bright-red"}}, dimRed); cr.OK {
		t.Error("plain red should not satisfy fg: bright-red")
	}
}

// TestCheckScreenAttrs_EveryCellMustMatch is the case that makes the matcher
// worth having: styling that stops halfway through a word is a real bug (a
// reset landing mid-token), and accepting it would let exactly that through.
func TestCheckScreenAttrs_EveryCellMustMatch(t *testing.T) {
	t.Parallel()
	res := cellsFor("ERROR", "ERR", func(c *runner.ScreenCell) { c.FG = 1 })
	cr := checkScreenAttrs([]spec.ScreenAttr{{Text: "ERROR", FG: "red"}}, res)
	if cr.OK {
		t.Error("half-red text should not satisfy a red claim")
	}
}

// TestCheckScreenAttrs_AnyOccurrenceSatisfies pins the position-free rule: with
// several occurrences, one that matches is enough, so a styling claim does not
// break every time the layout shifts.
func TestCheckScreenAttrs_AnyOccurrenceSatisfies(t *testing.T) {
	t.Parallel()
	// Two "ok"s: the first plain, the second green.
	row := []runner.ScreenCell{}
	for _, r := range "ok ok" {
		row = append(row, runner.ScreenCell{Rune: r, FG: spec.DefaultScreenColor})
	}
	row[3].FG, row[4].FG = 2, 2
	res := &runner.Result{IsPTY: true, Screen: []byte("ok ok"), ScreenCells: [][]runner.ScreenCell{row}}

	if cr := checkScreenAttrs([]spec.ScreenAttr{{Text: "ok", FG: "green"}}, res); !cr.OK {
		t.Errorf("a later matching occurrence should satisfy the entry: %s", cr.Hint)
	}
}
