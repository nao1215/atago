package ptyrun

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

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
