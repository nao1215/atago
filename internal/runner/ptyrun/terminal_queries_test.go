package ptyrun

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

func TestDA1Scanner_AcrossChunks(t *testing.T) {
	var s probeScanner
	if got := len(s.consume([]byte("\x1b["))); got != 0 {
		t.Fatalf("first chunk matches = %d, want 0", got)
	}
	if got := len(s.consume([]byte("0"))); got != 0 {
		t.Fatalf("second chunk matches = %d, want 0", got)
	}
	if got := len(s.consume([]byte("c"))); got != 1 {
		t.Fatalf("final chunk matches = %d, want 1", got)
	}
}

func TestDA1Scanner_DECID(t *testing.T) {
	var s probeScanner
	if got := len(s.consume([]byte("\x1bZ"))); got != 1 {
		t.Fatalf("DECID matches = %d, want 1", got)
	}
}

func TestTerminalQueries_ReplyToDA1AndDSR(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40}, &out)

	q.consume([]byte("abc\x1b[5n\x1b[6n\x1b[c"))

	want := "\x1b[0n\x1b[1;4R" + vt102DA1
	if got := out.String(); got != want {
		t.Fatalf("replies = %q, want %q", got, want)
	}
}

// TestTerminalQueries_OversizedCSIStillAnswersCPR is the #438 hang regression: a
// program that emits an enormous count-based CSI just before a cursor-position
// request must not stall the query terminal on that count — the sanitizer clamps
// it, so the emulator reaches the `\x1b[6n` and answers it. Running under the
// test's own deadline is the hang check.
func TestTerminalQueries_OversizedCSIStillAnswersCPR(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40}, &out)

	// CBT (backward tab) is one of vt10x's loop-per-count handlers; an
	// eight-billion count would spin it for minutes unclamped.
	q.consume([]byte("hi\x1b[99999999Z\x1b[6n"))

	if got := out.String(); !strings.HasSuffix(got, "R") || !strings.HasPrefix(got, "\x1b[") {
		t.Fatalf("no cursor-position reply after an oversized CSI: %q", got)
	}
}

// TestTerminalQueries_SplitOversizedCSIStillAnswersCPR is the split variant: the
// oversized CSI is delivered in two chunks, so a per-chunk sanitizer would let
// the two halves reassemble inside the emulator and spin. The streaming
// sanitizer holds the first half, and the cursor-position request in the second
// chunk is still answered.
func TestTerminalQueries_SplitOversizedCSIStillAnswersCPR(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40}, &out)

	q.consume([]byte("hi\x1b[8011"))     // ends mid-CSI
	q.consume([]byte("1111110Z\x1b[6n")) // completes the oversized CBT, then a CPR

	if got := out.String(); !strings.HasSuffix(got, "R") {
		t.Fatalf("no cursor-position reply after a split oversized CSI: %q", got)
	}
}

// TestTerminalQueries_MalformedCSIStillAnswersCPR pins that a malformed sequence
// does not abort the rest of the chunk: a later CPR request is still answered.
func TestTerminalQueries_MalformedCSIStillAnswersCPR(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40}, &out)

	// A negative parameter is malformed; a conformant terminal ignores the
	// sequence, and the CPR after it must still be answered.
	q.consume([]byte("hi\x1b[-5X\x1b[6n"))

	if got := out.String(); !strings.HasSuffix(got, "R") {
		t.Fatalf("no cursor-position reply after a malformed CSI: %q", got)
	}
}

// TestTerminalQueries_AbortedStringStillAnswersCPR pins that an OSC/string
// sequence aborted by a new escape (an ESC not followed by ST) does not swallow
// the rest of the stream: the CPR request after it is still answered. Holding the
// aborted string in the carry would have stranded the query.
func TestTerminalQueries_AbortedStringStillAnswersCPR(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40}, &out)

	// OSC "x" aborted by ESC A, then a cursor-position request.
	q.consume([]byte("\x1b]x\x1bA\x1b[6n"))

	if got := out.String(); !strings.HasSuffix(got, "R") {
		t.Fatalf("no cursor-position reply after an aborted string sequence: %q", got)
	}
}

// TestTerminalQueries_RepliesFollowRequestOrder pins that an answer written by
// atago (DA1) is not overtaken by one the emulator writes (DSR) for a probe
// asked after it. Programs read up to the DSR reply as their end-of-answers
// marker, so an overtaken reply is a reply they never see.
func TestTerminalQueries_RepliesFollowRequestOrder(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40}, &out)

	q.consume([]byte("\x1b[c\x1b[5n"))

	if got, want := out.String(), vt102DA1+"\x1b[0n"; got != want {
		t.Fatalf("replies = %q, want %q", got, want)
	}
}

func TestTerminalQueries_CellSizeOnlyWithGraphics(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		graphics string
		want     string
	}{
		{"", "\x1b[0n"},
		{spec.PTYGraphicsKitty, "\x1b[6;20;10t\x1b[0n"},
	} {
		var out bytes.Buffer
		q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40, Graphics: tc.graphics}, &out)
		q.consume([]byte("\x1b[16t\x1b[5n"))
		if got := out.String(); got != tc.want {
			t.Errorf("graphics %q: replies = %q, want %q", tc.graphics, got, tc.want)
		}
	}
}

// TestTerminalQueries_KittyQueryInOrder is the startup probe ratatui-image and
// similar libraries send: graphics query, DA1, cell size, then DSR as the end
// marker. Every answer must precede the DSR reply.
func TestTerminalQueries_KittyQueryInOrder(t *testing.T) {
	var out bytes.Buffer
	q := newTerminalQueries(&spec.PTY{Rows: 10, Cols: 40, Graphics: spec.PTYGraphicsKitty}, &out)

	probe := "\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[c\x1b[16t\x1b[5n"
	// Byte-at-a-time delivery is the worst case for a stateful scanner.
	for i := range len(probe) {
		q.consume([]byte{probe[i]})
	}

	want := "\x1b_Gi=31;OK\x1b\\" + vt102DA1 + "\x1b[6;20;10t" + "\x1b[0n"
	if got := out.String(); got != want {
		t.Fatalf("replies = %q, want %q", got, want)
	}
}
