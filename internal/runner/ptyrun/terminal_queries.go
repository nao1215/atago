package ptyrun

import (
	"fmt"
	"io"
	"sync"

	"github.com/hinshun/vt10x"

	"github.com/nao1215/atago/internal/spec"
)

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// writeQueryTerminal feeds a chunk to the query emulator, containing any panic
// from its escape parser so a malformed sequence during a live session cannot
// take down the run. This emulator only tracks state to answer device-attribute
// and cursor-position probes; the screen a `screen:` assertion reads is rendered
// separately by renderScreenCells.
func writeQueryTerminal(term vt10x.Terminal, chunk []byte) {
	defer func() { _ = recover() }()
	_, _ = term.Write(chunk)
}

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
	// mu guards term. consume runs on the drain goroutine while resize is
	// called from the session goroutine (#379), so the emulator has two writers
	// even though each is single-threaded on its own.
	mu     sync.Mutex
	term   vt10x.Terminal
	probes probeScanner
	// sanitize holds an incomplete trailing escape between chunks so a CSI split
	// across reads is bounded as one sequence instead of reassembling inside
	// vt10x (#438). Only feed touches it, under mu.
	sanitize streamSanitizer
	w        io.Writer
	// graphics is set when the step emulates a kitty-graphics terminal. It
	// answers the graphics query and records images, and its presence turns on
	// the cell-size reply.
	graphics *kittyGraphics
	// encodeGraphicsReply, when set, rewrites a graphics reply before it is
	// written. The Windows console host that forwards graphics drops an APC
	// string written to it as is, so there the reply goes out as key presses.
	encodeGraphicsReply func([]byte) []byte
}

// replyEncoder is implemented by a terminal whose host needs some replies
// rewritten to reach the program (the Windows pseudo console).
type replyEncoder interface {
	EncodeReply(p []byte) []byte
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
			// terminal state; we add the DA1 reply (and, with graphics, the
			// cell-size reply) alongside it.
			vt10x.WithWriter(w),
		),
		w:        w,
		graphics: graphicsFor(p),
	}
}

func graphicsFor(p *spec.PTY) *kittyGraphics {
	if !p.KittyGraphics() {
		return nil
	}
	return newKittyGraphics()
}

// resize keeps the query emulator the same size as the real terminal (#379).
// Its whole job is answering cursor-position reports, and a stale size makes
// those answers wrong for every program that asks after a resize.
func (t *terminalQueries) resize(rows, cols int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.term.Resize(cols, rows)
}

// consume feeds one chunk of program output to the emulator and answers the
// probes in it. Answers go out in the order the probes were asked, the way a
// real terminal answers: the emulator is fed only up to each probe before that
// probe's own reply is written. Feeding the whole chunk first would let the
// emulator's replies (a DSR status report, a cursor position) overtake the
// replies written here, and a program that reads up to the DSR report as its
// end-of-answers marker would never see a DA1 or graphics reply sent after it.
func (t *terminalQueries) consume(chunk []byte) {
	start := 0
	for i, b := range chunk {
		q, probed := t.probes.step(b)
		graphicsDone := t.graphics != nil && t.graphics.step(b)
		if !probed && !graphicsDone {
			continue
		}
		t.feed(chunk[start : i+1])
		start = i + 1
		if graphicsDone {
			for _, r := range t.graphics.takeReplies() {
				if t.encodeGraphicsReply != nil {
					r = t.encodeGraphicsReply(r)
				}
				_, _ = t.w.Write(r)
			}
		}
		if probed {
			t.answer(q)
		}
	}
	t.feed(chunk[start:])
}

// feed writes bytes to the emulator. The emulator sees the sanitized stream
// (counts clamped, malformed and split sequences bounded); the scanners see the
// raw bytes because they must still recognize a well-formed probe wherever it
// lands.
func (t *terminalQueries) feed(b []byte) {
	if len(b) == 0 {
		return
	}
	t.mu.Lock()
	writeQueryTerminal(t.term, t.sanitize.feed(b))
	t.mu.Unlock()
}

func (t *terminalQueries) answer(q queryKind) {
	switch q {
	case queryDA1:
		_, _ = t.w.Write([]byte(vt102DA1))
	case queryCellSize:
		if t.graphics != nil {
			_, _ = fmt.Fprintf(t.w, "\x1b[6;%d;%dt", graphicsCellHeight, graphicsCellWidth)
		}
	}
}

// queryKind is a terminal probe the scanner recognized.
type queryKind uint8

const (
	// queryDA1 is ESC Z (DECID) or CSI c / CSI 0 c.
	queryDA1 queryKind = iota
	// queryCellSize is CSI 16 t, "report the cell size in pixels".
	queryCellSize
)

// probeScanner incrementally recognizes the two legacy identify-terminal probes,
// ESC Z (DECID) and CSI c / CSI 0 c (DA1), and the cell-size probe CSI 16 t. It
// keeps state across read chunks because pty reads may split an escape sequence
// arbitrarily.
type probeScanner struct {
	state  da1State
	csiBuf []byte
}

type da1State uint8

const (
	da1Normal da1State = iota
	da1ESC
	da1CSI
)

// consume scans a whole chunk and returns every probe in it, in order.
func (s *probeScanner) consume(chunk []byte) []queryKind {
	var matched []queryKind
	for _, b := range chunk {
		if q, ok := s.step(b); ok {
			matched = append(matched, q)
		}
	}
	return matched
}

// step advances the scanner by one byte and reports a probe that byte
// completed.
func (s *probeScanner) step(b byte) (queryKind, bool) {
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
			s.state = da1Normal
			return queryDA1, true
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
			// detection; the live emulator still sees the original bytes.
			s.state = da1Normal
		case b >= 0x40 && b <= 0x7e:
			body := string(s.csiBuf)
			s.state = da1Normal
			s.csiBuf = s.csiBuf[:0]
			if b == 'c' && isDA1Request([]byte(body)) {
				return queryDA1, true
			}
			if b == 't' && body == "16" {
				return queryCellSize, true
			}
		default:
			if len(s.csiBuf) < 32 {
				s.csiBuf = append(s.csiBuf, b)
			} else {
				s.state = da1Normal
				s.csiBuf = s.csiBuf[:0]
			}
		}
	}
	return 0, false
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
