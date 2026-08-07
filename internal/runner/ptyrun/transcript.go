package ptyrun

import (
	"errors"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// transcriptDrain owns the live terminal reader and the accumulated transcript.
// Starting it before the child starts closes the fast-exit window where a real
// command can print and exit before atago posts its first Read on the master.
type transcriptDrain struct {
	p  *spec.PTY
	rw io.ReadWriter

	writeMu sync.Mutex

	mu          sync.Mutex
	transcript  []byte
	readErr     error
	readDone    chan struct{}
	screenLen   int
	screenCache []byte
}

func startTranscriptDrain(rw io.ReadWriter, p *spec.PTY) *transcriptDrain {
	t := &transcriptDrain{
		p:         p,
		rw:        rw,
		readDone:  make(chan struct{}),
		screenLen: -1,
	}
	queries := newTerminalQueries(p, writerFunc(t.write))
	go func() {
		defer close(t.readDone)
		buf := make([]byte, 4096)
		for {
			n, rerr := t.rw.Read(buf)
			if n > 0 {
				t.mu.Lock()
				t.transcript = append(t.transcript, buf[:n]...)
				t.mu.Unlock()
				queries.consume(buf[:n])
			}
			if rerr == nil {
				continue
			}
			// An interrupted read is not an end: resume it, or a signal arriving
			// mid-read truncates the transcript.
			if errors.Is(rerr, syscall.EINTR) {
				continue
			}
			if !isSessionEnd(rerr) {
				t.mu.Lock()
				t.readErr = rerr
				t.mu.Unlock()
			}
			return
		}
	}()
	return t
}

func (t *transcriptDrain) write(b []byte) (int, error) {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return t.rw.Write(b)
}

func (t *transcriptDrain) snapshot() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]byte(nil), t.transcript...)
}

// tailFrom copies only transcript[from:] under the lock and reports the
// transcript's current length; curLen reports the length alone. Together they
// let a pending expect skip the poll entirely when nothing new arrived and copy
// only the bytes it can still match.
func (t *transcriptDrain) tailFrom(from int) ([]byte, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if from > len(t.transcript) {
		from = len(t.transcript)
	}
	return append([]byte(nil), t.transcript[from:]...), len(t.transcript)
}

func (t *transcriptDrain) curLen() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.transcript)
}

func (t *transcriptDrain) currentScreen() []byte {
	t.mu.Lock()
	n := len(t.transcript)
	if n == t.screenLen {
		screen := append([]byte(nil), t.screenCache...)
		t.mu.Unlock()
		return screen
	}
	snap := append([]byte(nil), t.transcript...)
	t.mu.Unlock()

	rendered := []byte(RenderScreen(snap, t.p))

	t.mu.Lock()
	if len(t.transcript) == n {
		t.screenCache = append(t.screenCache[:0], rendered...)
		t.screenLen = n
	}
	t.mu.Unlock()
	return rendered
}

func (t *transcriptDrain) readError() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.readErr
}

// closeGrace bounds the join after the terminal is closed. Closing it is what
// ends the pending read, so this is a backstop rather than the normal path: a
// reader that stays blocked anyway must not wedge the whole run — a pty drain
// blocked past a close is how two packages once died on the 5m test timeout.
const closeGrace = 10 * time.Second

func (t *transcriptDrain) waitDrain(closeTerm func(), grace time.Duration) {
	if grace > 0 {
		select {
		case <-t.readDone:
		case <-time.After(grace):
		}
	}
	closeTerm()
	select {
	case <-t.readDone:
	case <-time.After(closeGrace):
	}
}
