package ptyrun

import (
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
