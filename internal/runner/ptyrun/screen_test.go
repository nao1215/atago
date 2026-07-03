package ptyrun

import (
	"strings"
	"testing"

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
