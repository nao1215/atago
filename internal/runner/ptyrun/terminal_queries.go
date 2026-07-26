package ptyrun

import (
	"io"

	"github.com/hinshun/vt10x"

	"github.com/nao1215/atago/internal/spec"
)

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// vt102DA1 is a conservative primary device-attributes reply. atago's screen
// emulator is xterm-ish enough for ordinary TUIs, but it does not implement a
// modern terminal's full feature matrix; replying as a plain VT102-class device
// avoids promising capabilities we do not emulate while still satisfying apps
// that only need a well-formed DA1 response.
const vt102DA1 = "\x1b[?6c"

// terminalQueries mirrors the live pty stream through a terminal emulator so
// CPR/DSR requests can be answered against the current cursor position, and
// supplements vt10x with the missing DA1/DECID replies Yazi expects at startup.
type terminalQueries struct {
	term vt10x.Terminal
	da1  da1Scanner
	w    io.Writer
}

func newTerminalQueries(p *spec.PTY, w io.Writer) *terminalQueries {
	rows, cols := defaultRows, defaultCols
	if p.Rows > 0 {
		rows = p.Rows
	}
	if p.Cols > 0 {
		cols = p.Cols
	}
	return &terminalQueries{
		term: vt10x.New(
			vt10x.WithSize(cols, rows),
			// vt10x already emits DSR/CPR and OSC color replies from the live
			// terminal state; we add only the DA1 gap alongside it.
			vt10x.WithWriter(w),
		),
		w: w,
	}
}

func (t *terminalQueries) consume(chunk []byte) {
	writeTranscript(t.term, chunk)
	for range t.da1.consume(chunk) {
		_, _ = t.w.Write([]byte(vt102DA1))
	}
}

// da1Scanner incrementally recognizes the two legacy identify-terminal probes:
// ESC Z (DECID) and CSI c / CSI 0 c (DA1). It keeps state across read chunks
// because pty reads may split an escape sequence arbitrarily.
type da1Scanner struct {
	state  da1State
	csiBuf []byte
}

type da1State uint8

const (
	da1Normal da1State = iota
	da1ESC
	da1CSI
)

func (s *da1Scanner) consume(chunk []byte) []struct{} {
	var matched []struct{}
	for _, b := range chunk {
		switch s.state {
		case da1Normal:
			if b == 0x1b {
				s.state = da1ESC
			}
		case da1ESC:
			switch b {
			case '[':
				s.state = da1CSI
				s.csiBuf = s.csiBuf[:0]
			case 'Z':
				matched = append(matched, struct{}{})
				s.state = da1Normal
			case 0x1b:
				// A new ESC restarts escape parsing.
				s.state = da1ESC
			default:
				s.state = da1Normal
			}
		case da1CSI:
			switch {
			case b == 0x1b:
				s.state = da1ESC
			case b < 0x20 || b == 0x7f:
				// Control bytes inside a half-read probe abort it for our narrow
				// DA1 detection; the live emulator still sees the original bytes.
				s.state = da1Normal
			case b >= 0x40 && b <= 0x7e:
				if b == 'c' && isDA1Request(s.csiBuf) {
					matched = append(matched, struct{}{})
				}
				s.state = da1Normal
				s.csiBuf = s.csiBuf[:0]
			default:
				if len(s.csiBuf) < 32 {
					s.csiBuf = append(s.csiBuf, b)
				} else {
					s.state = da1Normal
					s.csiBuf = s.csiBuf[:0]
				}
			}
		}
	}
	return matched
}

func isDA1Request(body []byte) bool {
	if len(body) == 0 {
		return true // CSI c
	}
	if len(body) == 1 && body[0] == '0' {
		return true // CSI 0 c
	}
	return false
}
