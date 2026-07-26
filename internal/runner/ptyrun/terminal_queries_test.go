package ptyrun

import (
	"bytes"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

func TestDA1Scanner_AcrossChunks(t *testing.T) {
	var s da1Scanner
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
	var s da1Scanner
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
