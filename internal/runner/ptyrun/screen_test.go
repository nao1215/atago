package ptyrun

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// RenderScreen renders a transcript's final screen as plain text. It is
// test-only sugar over renderScreenCells, which production calls directly.
func RenderScreen(transcript []byte, p *spec.PTY) string {
	return renderScreenResized(transcript, p, nil)
}

// renderScreenResized is RenderScreen with mid-session resizes (#379).
func renderScreenResized(transcript []byte, p *spec.PTY, resizes []screenResize) string {
	text, _ := renderScreenCells(transcript, p, resizes)
	return text
}

// TestRenderScreen_OverwriteAndErase proves the emulator's whole value (#27):
// a line overwritten with \r shows only its FINAL text on the screen, while
// the raw transcript would contain both versions.
func TestRenderScreen_OverwriteAndErase(t *testing.T) {
	t.Parallel()
	transcript := []byte("loading...\rdone.     \r\nnext\r\n")
	got := RenderScreen(transcript, &spec.PTY{})
	if !strings.Contains(got, "done.") {
		t.Errorf("screen = %q, want the final text", got)
	}
	if strings.Contains(got, "loading") {
		t.Errorf("screen = %q, must not contain the overwritten text", got)
	}
	if !strings.Contains(got, "next") {
		t.Errorf("screen = %q, want the second line", got)
	}
}

// TestRenderScreen_CursorMovementAndClear covers cursor addressing and
// screen-clear sequences.
func TestRenderScreen_CursorMovementAndClear(t *testing.T) {
	t.Parallel()
	// Draw garbage, clear the screen (ED 2), home the cursor, draw the menu.
	transcript := []byte("garbage everywhere\r\n\x1b[2J\x1b[HMain Menu\r\n> Settings\r\n")
	got := RenderScreen(transcript, &spec.PTY{Rows: 10, Cols: 40})
	lines := strings.Split(got, "\n")
	if len(lines) < 2 || lines[0] != "Main Menu" || lines[1] != "> Settings" {
		t.Errorf("screen = %q, want the cleared redraw only", got)
	}
	if strings.Contains(got, "garbage") {
		t.Errorf("screen = %q, must not contain pre-clear content", got)
	}
}

// TestRenderScreen_WideCharactersOccupyTwoColumns is the #432 regression: a wide
// character (CJK, emoji) takes two terminal columns, so cursor addressing after
// it and wrapping at the right margin depend on the emulator modeling that width.
// The previous emulator stored one column per rune, so a label positioned with
// absolute cursor addressing after a Japanese string landed two columns early and
// a wide row did not wrap where the terminal wraps it. Each case below is what a
// real terminal shows.
func TestRenderScreen_WideCharactersOccupyTwoColumns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		transcript string
		rows, cols int
		want       string
	}{
		{
			// "日本語" is three characters and SIX columns, so a label the program
			// positions at column 7 (\x1b[1;7H) sits flush against it, not after a
			// three-column gap.
			name:       "cursor addressing after a wide string is not shifted",
			transcript: "\x1b[2J\x1b[1;1H日本語\x1b[1;7H[OK]",
			rows:       3, cols: 20,
			want: "日本語[OK]",
		},
		{
			// Overwriting the SECOND column of a wide character blanks it: writing
			// `x` at column 2, inside "日", leaves the terminal showing " x" where
			// "日" was, then the rest. A one-column-per-rune model would instead
			// replace the wrong character.
			name:       "overwriting a wide character's second column blanks it",
			transcript: "\x1b[2J\x1b[1;1H日本語\x1b[1;2Hx",
			rows:       2, cols: 20,
			want: " x本語",
		},
		{
			// "ab🎉" is columns 1-2 (ab) and 3-4 (the emoji), so text positioned at
			// column 5 follows it with no gap.
			name:       "an emoji is two columns wide",
			transcript: "\x1b[2J\x1b[1;1Hab🎉\x1b[1;5Hcd",
			rows:       2, cols: 20,
			want: "ab🎉cd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderScreen([]byte(tt.transcript), &spec.PTY{Rows: tt.rows, Cols: tt.cols})
			if got != tt.want {
				t.Errorf("RenderScreen(%q) = %q, want %q", tt.transcript, got, tt.want)
			}
		})
	}
}

// TestRenderScreen_PreservesGraphemeClusters is the #437 regression: a grapheme
// cluster is one cell on the terminal and must render whole. Keeping only the
// leading rune returned the base emoji alone for a ZWJ sequence, so a screen
// text match on the real cluster could never succeed. A precomposed accent
// (U+00E9) is a single rune and was never affected; a DECOMPOSED
// base-plus-combining-mark is reduced to its base by the emulator upstream,
// which is out of atago's hands, so it is not asserted here.
func TestRenderScreen_PreservesGraphemeClusters(t *testing.T) {
	t.Parallel()
	// A woman-technologist ZWJ sequence: woman + ZWJ + laptop, one cluster over
	// one cell.
	const dev = "\U0001F469\u200d\U0001F4BB"
	got := RenderScreen([]byte(dev+"\r\n"), &spec.PTY{Rows: 3, Cols: 20})
	if got != dev {
		t.Errorf("RenderScreen(%q) = %q, want the whole cluster %q", dev, got, dev)
	}
}

// TestRenderScreen_WideCharacterAutowrap is the #503 regression. A real
// terminal arms pending wrap once a wide character has filled the last columns,
// so the next character starts the following row, and it wraps a wide character
// that no longer fits rather than dropping it. The emulator arms pending wrap
// only from the LAST column, which a wide character reaches from the
// second-to-last — so `X` overwrote the second half of `本` (blanking its first
// half) and `語` fell off the end entirely.
func TestRenderScreen_WideCharacterAutowrap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		transcript string
		rows, cols int
		want       string
	}{
		{
			name:       "a character after a wide one filled the row starts the next row",
			transcript: "日本X",
			rows:       5, cols: 4,
			want: "日本\nX",
		},
		{
			name:       "a wide character that no longer fits wraps instead of vanishing",
			transcript: "日本語",
			rows:       5, cols: 5,
			want: "日本\n語",
		},
		{
			name:       "a wide character after a narrow one filled the row wraps too",
			transcript: "abcd日",
			rows:       5, cols: 4,
			want: "abcd\n日",
		},
		{
			// The wrap must not fire twice: `日本` fills row 1 exactly, and the
			// program's own newline is what moves to row 2.
			name:       "an explicit newline after a filled row does not skip one",
			transcript: "日本\r\nX",
			rows:       5, cols: 4,
			want: "日本\nX",
		},
		{
			// Cursor addressing cancels a pending wrap, the way a terminal does.
			name:       "cursor addressing after a filled row cancels the pending wrap",
			transcript: "日本\x1b[1;1HX",
			rows:       5, cols: 4,
			want: "X 本",
		},
		{
			// With autowrap off (DECAWM reset) a terminal keeps overwriting the
			// last column instead of wrapping, so the correction must stay out.
			name:       "autowrap off keeps the emulator's overwrite behavior",
			transcript: "\x1b[?7l日本X",
			rows:       5, cols: 4,
			want: "日 X",
		},
		{
			// Narrow autowrap was never broken; it must keep working alongside.
			name:       "narrow characters still wrap at the last column",
			transcript: "abcdX",
			rows:       5, cols: 4,
			want: "abcd\nX",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderScreen([]byte(tt.transcript), &spec.PTY{Rows: tt.rows, Cols: tt.cols})
			if got != tt.want {
				t.Errorf("RenderScreen(%q) = %q, want %q", tt.transcript, got, tt.want)
			}
		})
	}
}

// TestRenderScreen_PreservesCombiningMarks is the #505 regression: a decomposed
// grapheme (an ASCII base plus a combining mark) must render as the program
// wrote it. The emulator commits an ASCII character to the screen the moment it
// arrives, so the mark that follows can never join it and was dropped — a
// `screen:` assertion saw `e` where the program wrote `e´`.
func TestRenderScreen_PreservesCombiningMarks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		transcript string
		rows, cols int
		want       string
	}{
		{
			name:       "a decomposed accent keeps its mark",
			transcript: "e\u0301",
			rows:       3, cols: 10,
			want: "e\u0301",
		},
		{
			name:       "the mark does not swallow the character after it",
			transcript: "e\u0301X",
			rows:       3, cols: 10,
			want: "e\u0301X",
		},
		{
			name:       "two marks on one base are both kept",
			transcript: "e\u0301\u0308",
			rows:       3, cols: 10,
			want: "e\u0301\u0308",
		},
		{
			name:       "a mark on the last column still lands on its base",
			transcript: "abce\u0301",
			rows:       3, cols: 4,
			want: "abce\u0301",
		},
		{
			// A precomposed accent and a non-ASCII base were never affected.
			name:       "a precomposed accent is unchanged",
			transcript: "\u00e9",
			rows:       3, cols: 10,
			want: "\u00e9",
		},
		{
			name:       "a wide base keeps its mark",
			transcript: "\u304b\u3099",
			rows:       3, cols: 10,
			want: "\u304b\u3099",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RenderScreen([]byte(tt.transcript), &spec.PTY{Rows: tt.rows, Cols: tt.cols})
			if got != tt.want {
				t.Errorf("RenderScreen(%q) = %q, want %q", tt.transcript, got, tt.want)
			}
		})
	}
}

// TestRenderScreen_TrailingNormalization proves per-line trailing whitespace
// and trailing blank rows are stripped so snapshots stay stable.
func TestRenderScreen_TrailingNormalization(t *testing.T) {
	t.Parallel()
	got := RenderScreen([]byte("only line   \r\n"), &spec.PTY{Rows: 24, Cols: 80})
	if got != "only line" {
		t.Errorf("screen = %q, want %q (trailing spaces and blank rows stripped)", got, "only line")
	}
}

// TestRenderScreen_NegativeCSIParamDoesNotCrash guards the fix for the crash
// FuzzRenderScreen found: a negative CSI DCH ("\x1b[-10P") or ICH ("\x1b[-5@")
// parameter drove vt10x's deleteChars/insertBlanks into a slice-bounds panic.
// A real terminal treats "-" as an invalid parameter byte and ignores the whole
// sequence; RenderScreen now drops it, so the surrounding text still renders.
func TestRenderScreen_NegativeCSIParamDoesNotCrash(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name       string
		transcript string
	}{
		{"DCH", "before\x1b[-10Pafter"},
		{"ICH", "before\x1b[-5@after"},
		{"large negative DCH", "x\x1b[-99999999Py"},
	} {
		got := RenderScreen([]byte(tc.transcript), &spec.PTY{Rows: 5, Cols: 40})
		if !strings.Contains(got, "before") && !strings.Contains(got, "x") {
			t.Errorf("%s: screen = %q, want the surrounding text preserved", tc.name, got)
		}
	}
}

// TestRenderScreen_PanicContainment guards the recover backstop: vt10x's
// parser re-enters CSI across an invalid-rune run, so "ESC \xf3 [ -1 P"
// reaches deleteChars with a negative parameter even though the byte-level
// pre-filter sees no ESC-'[' pair (found by FuzzRenderScreen). The emulator's
// panic must stay contained — the process survives and everything drawn before
// the malformed sequence still renders.
func TestRenderScreen_PanicContainment(t *testing.T) {
	t.Parallel()
	got := RenderScreen([]byte("drawn\r\n\x1b\xf3[-1P"), &spec.PTY{Rows: 5, Cols: 40})
	if !strings.Contains(got, "drawn") {
		t.Errorf("screen = %q, want the pre-panic content preserved", got)
	}
}

// TestRenderScreen_DroppedSequenceStaysStateNeutral guards the second bug
// FuzzRenderScreen found: dropping a malformed CSI must not re-open the
// unterminated CSI before it. In the raw stream the second ESC aborts the
// first sequence's 16-digit parameter buffer; a naive sanitizer that copies
// the aborted prefix verbatim but deletes the aborting sequence leaves that
// buffer OPEN in its output, and the ordinary text 'Z' that follows becomes
// the final byte of a quadrillion-step CBT — hanging the emulator for hours.
func TestRenderScreen_DroppedSequenceStaysStateNeutral(t *testing.T) {
	t.Parallel()
	got := RenderScreen([]byte("00\x1b[2\x89\xd40000000000000\x8300\x1b[2\x89\xd4\x82tZ"), &spec.PTY{Rows: 1, Cols: 140})
	if !strings.Contains(got, "Z") {
		t.Errorf("screen = %q, want the trailing text rendered", got)
	}
}

// TestRenderScreen_HugeRepeatCountIsClamped guards the hang FuzzRenderScreen
// found: vt10x executes CBT/CHT one tab stop at a time, so an adversarial
// "CSI 80111111110 Z" spins for minutes. The sanitizer clamps oversized digit
// runs, keeping the sequence's effect while bounding its cost.
func TestRenderScreen_HugeRepeatCountIsClamped(t *testing.T) {
	t.Parallel()
	done := make(chan string, 1)
	go func() {
		done <- RenderScreen([]byte("x\x1b[80111111110Zy"), &spec.PTY{Rows: 5, Cols: 40})
	}()
	select {
	case got := <-done:
		// The clamped CBT legitimately tabs the cursor back to column 0, so
		// 'y' overwrites 'x' — exactly what a real terminal shows. The point
		// is completing at all, with the trailing text rendered.
		if !strings.Contains(got, "y") {
			t.Errorf("screen = %q, want the trailing text rendered", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("RenderScreen still hangs on a huge CSI repeat count")
	}
}

// TestRenderScreen_KeepsValidCSI proves the negative-param guard leaves
// well-formed CSI sequences (and their effects) untouched.
func TestRenderScreen_KeepsValidCSI(t *testing.T) {
	t.Parallel()
	// A normal SGR-colored line plus a positive DCH must still render.
	got := RenderScreen([]byte("\x1b[31mred\x1b[0m\r\n"), &spec.PTY{Rows: 5, Cols: 20})
	if got != "red" {
		t.Errorf("screen = %q, want %q", got, "red")
	}
}

// TestRenderScreen_TruncatesAtCols proves output wider than the terminal
// wraps at cols, mirroring a real terminal.
func TestRenderScreen_TruncatesAtCols(t *testing.T) {
	t.Parallel()
	got := RenderScreen([]byte(strings.Repeat("x", 15)+"\r\n"), &spec.PTY{Rows: 5, Cols: 10})
	lines := strings.Split(got, "\n")
	if len(lines) < 2 || len(lines[0]) != 10 || len(lines[1]) != 5 {
		t.Errorf("screen = %q, want a 10-col wrap then 5 leftover chars", got)
	}
}

// TestRenderScreenResized proves the rendered screen follows a mid-session
// resize (#379): each part of the transcript is drawn under the size it was
// actually produced under, so a frame written at 10 columns is not re-flowed by
// a later widening — and one written after the widening is not truncated to the
// old width.
func TestRenderScreenResized(t *testing.T) {
	t.Parallel()
	// "abcdefghij" fills a 10-column row; at 10 columns the trailing "XY" wraps
	// onto the next line, at 30 it stays on one.
	transcript := []byte("abcdefghijXY\r\n")
	narrow := RenderScreen(transcript, &spec.PTY{Rows: 5, Cols: 10})
	if narrow != "abcdefghij\nXY" {
		t.Fatalf("baseline at 10 cols = %q", narrow)
	}

	// The resize lands BEFORE those bytes, so they are drawn at 30 columns.
	got := renderScreenResized(transcript, &spec.PTY{Rows: 5, Cols: 10},
		[]screenResize{{offset: 0, rows: 5, cols: 30}})
	if got != "abcdefghijXY" {
		t.Errorf("after widening = %q, want one unwrapped line", got)
	}

	// The resize lands AFTER them, so they keep the wrap they were drawn with.
	got = renderScreenResized(transcript, &spec.PTY{Rows: 5, Cols: 10},
		[]screenResize{{offset: len(transcript), rows: 5, cols: 30}})
	if got != "abcdefghij\nXY" {
		t.Errorf("resize after the output = %q, want the original wrap kept", got)
	}
}

// TestRenderScreenResized_OffsetsAreClamped keeps a recorded offset that no
// longer indexes the transcript (a shorter snapshot, an offset past the end)
// from panicking the render, since a screen assert runs on whatever bytes have
// arrived rather than on a finished transcript.
func TestRenderScreenResized_OffsetsAreClamped(t *testing.T) {
	t.Parallel()
	transcript := []byte("hello\r\n")
	for _, sizes := range [][]screenResize{
		{{offset: -5, rows: 4, cols: 20}},
		{{offset: 999, rows: 4, cols: 20}},
		// Out of order: the second offset precedes the first.
		{{offset: 5, rows: 4, cols: 20}, {offset: 1, rows: 6, cols: 30}},
	} {
		got := renderScreenResized(transcript, &spec.PTY{Rows: 4, Cols: 20}, sizes)
		if !strings.Contains(got, "hello") {
			t.Errorf("resizes %+v lost the output: %q", sizes, got)
		}
	}
}

// TestRenderScreen_NoResizesIsUnchanged pins that adding the resize path did not
// move the ordinary case: with no size changes the render is byte-identical to
// what the one-segment version produced.
func TestRenderScreen_NoResizesIsUnchanged(t *testing.T) {
	t.Parallel()
	transcript := []byte("loading...\rdone.     \r\n")
	if got, want := renderScreenResized(transcript, &spec.PTY{Rows: 5, Cols: 20}, nil),
		RenderScreen(transcript, &spec.PTY{Rows: 5, Cols: 20}); got != want {
		t.Errorf("resized render = %q, plain render = %q", got, want)
	}
}

// TestSanitizeTranscriptMarks_CutsLandOnUnitBoundaries is the regression for
// the hang FuzzRenderScreen found while #379 was being written: the resized
// render first cut the RAW transcript and sanitized each piece, so a cut inside
// "\x1b[80111111110Z" hid the sequence from the repeat-count clamp — the pieces
// passed through as a truncated CSI plus ordinary digits, vt10x's stateful
// parser put them back together, and the render spun for minutes on a
// quadrillion-step tab.
//
// Translating the offset instead keeps the clamp in force, and lands the cut on
// a boundary between whole units so no sequence is ever split.
func TestSanitizeTranscriptMarks_CutsLandOnUnitBoundaries(t *testing.T) {
	t.Parallel()
	raw := []byte("x\x1b[80111111110Zy")
	// A mark in the middle of the sequence.
	sanitized, cuts := sanitizeTranscriptMarks(raw, []int{len(raw) / 2})
	if len(cuts) != 1 {
		t.Fatalf("got %d translated offsets, want 1", len(cuts))
	}
	// The clamp still applied: the absurd repeat count is not in the output.
	if bytes.Contains(sanitized, []byte("80111111110")) {
		t.Errorf("the repeat count survived sanitizing: %q", sanitized)
	}
	// Neither piece ends mid-sequence: writing them in order is the same as
	// writing the whole buffer.
	head, tail := sanitized[:cuts[0]], sanitized[cuts[0]:]
	if bytes.Contains(head, []byte{0x1b}) && !bytes.HasSuffix(head, []byte("Z")) {
		t.Errorf("head %q ends inside a sequence", head)
	}
	if got := string(append(append([]byte(nil), head...), tail...)); got != string(sanitized) {
		t.Errorf("pieces do not reassemble: %q", got)
	}

	// The render through that cut must finish promptly rather than hang.
	done := make(chan string, 1)
	go func() {
		done <- renderScreenResized(raw, &spec.PTY{Rows: 5, Cols: 40},
			[]screenResize{{offset: len(raw) / 2, rows: 5, cols: 40}})
	}()
	select {
	case got := <-done:
		// The clamped back-tab still runs, so what remains on screen is whatever
		// the clamped sequence produced — the same thing the unsplit render
		// produces. Equality is the real claim: cutting the transcript for a
		// resize must not change what the frame says.
		if want := RenderScreen(raw, &spec.PTY{Rows: 5, Cols: 40}); got != want {
			t.Errorf("split render = %q, unsplit render = %q", got, want)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("renderScreenResized hangs when a resize falls inside a CSI sequence")
	}
}

// TestSanitizeTranscriptMarks_TranslatesEveryOffset covers the plain
// bookkeeping: one translated offset per mark, ascending, always inside the
// sanitized buffer — including the no-ESC fast path and offsets past the end.
func TestSanitizeTranscriptMarks_TranslatesEveryOffset(t *testing.T) {
	t.Parallel()
	for _, raw := range [][]byte{
		[]byte("plain text with no escapes"),
		[]byte("a\x1b[31mred\x1b[0mb"),
		[]byte("\x1b[-5@dropped\x1b[2J"),
	} {
		marks := []int{0, 1, len(raw) / 2, len(raw), len(raw) + 100}
		sanitized, cuts := sanitizeTranscriptMarks(raw, marks)
		if len(cuts) != len(marks) {
			t.Fatalf("%q: got %d offsets for %d marks", raw, len(cuts), len(marks))
		}
		for i, c := range cuts {
			if c < 0 || c > len(sanitized) {
				t.Errorf("%q: offset %d = %d, outside the sanitized buffer (len %d)", raw, i, c, len(sanitized))
			}
			if i > 0 && c < cuts[i-1] {
				t.Errorf("%q: offsets not ascending: %v", raw, cuts)
			}
		}
	}
}

// TestRenderScreenCells_TracksColorsAndAttributes proves the cell render picks
// up what the plain-text render throws away (#382), and — the part that matters
// most — that the two views agree about which row a character is on. A row-number
// mismatch would make every attribute failure point at the wrong line.
func TestRenderScreenCells_TracksColorsAndAttributes(t *testing.T) {
	t.Parallel()
	transcript := []byte("\x1b[1;31mERROR\x1b[0m ok\r\n\x1b[7mSELECTED\x1b[0m\r\n")
	text, cells := renderScreenCells(transcript, &spec.PTY{Rows: 5, Cols: 40}, nil)

	if text != "ERROR ok\nSELECTED" {
		t.Fatalf("text = %q", text)
	}
	if len(cells) != 2 {
		t.Fatalf("got %d cell rows, want 2 (text has %d lines)", len(cells), strings.Count(text, "\n")+1)
	}
	// Row 1: ERROR is bold red, " ok" is neither. The emulator stores the color
	// the program emitted — index 1 (red) — with bold as its own attribute, rather
	// than folding bold into a bright index. A terminal still DRAWS bold red in the
	// bright variant, and the assertion layer accounts for that (colorMatches
	// widens `red` toward `bright-red` under bold); the stored color stays faithful
	// to the SGR code.
	for i, r := range "ERROR" {
		c := cells[0][i]
		if c.Content != string(r) || c.FG != 1 || !c.Bold {
			t.Errorf("cell (1,%d) = %+v, want %q bold red", i+1, c, r)
		}
	}
	for i := len("ERROR"); i < len("ERROR ok"); i++ {
		if c := cells[0][i]; c.Bold || c.FG != runner.DefaultColor {
			t.Errorf("cell (1,%d) = %+v, want unstyled", i+1, c)
		}
	}
	// Row 2: reverse video, which is how a TUI draws the selected row.
	for i := range "SELECTED" {
		if c := cells[1][i]; !c.Reverse {
			t.Errorf("cell (2,%d) = %+v, want reverse", i+1, c)
		}
	}

	// The alignment law: the text is exactly the cells' content, row for row.
	for y, line := range strings.Split(text, "\n") {
		var b strings.Builder
		for _, c := range cells[y] {
			b.WriteString(c.Content)
		}
		if got := strings.TrimRight(b.String(), " \t"); got != line {
			t.Errorf("row %d: text %q, cells %q", y+1, line, got)
		}
	}
}

// TestRenderScreenCells_DefaultColorsAreDistinguishable is what makes a
// `--no-color` contract assertable: text the program never colored has to come
// back as the terminal's own color, not as some arbitrary index.
func TestRenderScreenCells_DefaultColorsAreDistinguishable(t *testing.T) {
	t.Parallel()
	_, cells := renderScreenCells([]byte("plain\r\n"), &spec.PTY{Rows: 3, Cols: 20}, nil)
	if len(cells) == 0 {
		t.Fatal("no cells rendered")
	}
	for i := range "plain" {
		c := cells[0][i]
		if c.FG != spec.DefaultScreenColor && c.FG != spec.DefaultScreenColor+1 {
			t.Errorf("cell %d fg = %d, want a default color", i, c.FG)
		}
	}
}
