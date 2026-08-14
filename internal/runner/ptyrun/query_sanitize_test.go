package ptyrun

import (
	"strings"
	"testing"
)

// TestStreamSanitizer_SplitOversizedCSI is the core #438 case: an oversized
// count-based CSI split across two chunks must not reach the emulator as two
// halves that reassemble into a quadrillion-step sequence. The sanitizer holds
// the incomplete first half and clamps the whole thing once it is complete.
func TestStreamSanitizer_SplitOversizedCSI(t *testing.T) {
	t.Parallel()
	var s streamSanitizer

	// First chunk ends mid-CSI (no final byte): nothing is safe to emit yet.
	if got := s.feed([]byte("\x1b[8011")); len(got) != 0 {
		t.Fatalf("first half emitted %q, want nothing held back", got)
	}
	// Second chunk completes the CBT. The reassembled 80111111110 count is clamped
	// to four digits before it reaches the emulator.
	got := string(s.feed([]byte("1111110Z")))
	if got != "\x1b[8011Z" {
		t.Fatalf("completed sequence = %q, want the clamped %q", got, "\x1b[8011Z")
	}
}

// TestStreamSanitizer_HoldsAndCompletes covers the ordinary split: a clean CSI
// broken across chunks is emitted intact once its final byte arrives, and plain
// text on either side passes straight through.
func TestStreamSanitizer_HoldsAndCompletes(t *testing.T) {
	t.Parallel()
	var s streamSanitizer

	if got := string(s.feed([]byte("ab\x1b[1"))); got != "ab" {
		t.Fatalf("first chunk = %q, want the plain prefix %q", got, "ab")
	}
	if got := string(s.feed([]byte(";31mcd"))); got != "\x1b[1;31mcd" {
		t.Fatalf("second chunk = %q, want the completed sequence plus text", got)
	}
}

// TestStreamSanitizer_DropsNeverEndingEscape pins the buffer bound: an escape
// that never terminates must not grow the carry without limit; past the cap it
// is dropped rather than held or fed on.
func TestStreamSanitizer_DropsNeverEndingEscape(t *testing.T) {
	t.Parallel()
	var s streamSanitizer

	// An OSC with no terminator, longer than the carry cap.
	huge := "\x1b]0;" + strings.Repeat("x", maxQueryCarry+16)
	if got := s.feed([]byte(huge)); len(got) != 0 {
		t.Fatalf("an incomplete over-long escape emitted %q, want nothing", got)
	}
	// The next ordinary text is emitted normally — the dropped escape did not
	// wedge the sanitizer.
	if got := string(s.feed([]byte("hello"))); got != "hello" {
		t.Fatalf("after dropping the over-long escape, got %q, want %q", got, "hello")
	}
}

// TestIncompleteEscapeStart pins where the boundary between complete units and a
// trailing incomplete escape falls, for the shapes the streaming split depends
// on.
func TestIncompleteEscapeStart(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		in   string
		want int
	}{
		{"no escape ends on a boundary", "plain text", len("plain text")},
		{"complete CSI then boundary", "\x1b[1;31mX", len("\x1b[1;31mX")},
		{"trailing lone ESC is incomplete", "ab\x1b", 2},
		{"trailing CSI without final is incomplete", "ab\x1b[12;", 2},
		{"trailing OSC without terminator is incomplete", "ab\x1b]0;title", 2},
		{"OSC closed by BEL is complete", "\x1b]0;t\x07", len("\x1b]0;t\x07")},
		{"OSC closed by ST is complete", "\x1b]0;t\x1b\\", len("\x1b]0;t\x1b\\")},
		// An ESC that is not the start of ST aborts the string; the CSI after it is
		// a complete unit, so nothing is held.
		{"OSC aborted by a new escape ends at that escape", "\x1b]x\x1bA\x1b[6n", len("\x1b]x\x1bA\x1b[6n")},
		{"OSC aborted then trailing incomplete CSI", "\x1b]x\x1bA\x1b[6", len("\x1b]x\x1bA")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := incompleteEscapeStart([]byte(tc.in)); got != tc.want {
				t.Errorf("incompleteEscapeStart(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
