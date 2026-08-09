package ptyrun

import (
	"errors"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/runner"
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
	cellsCache  [][]runner.ScreenCell
	// modes records the DEC private modes the program has requested, read off
	// its own output (#378). A terminal feature the program never asked for is
	// one whose input atago must not fabricate.
	modes map[int]bool
	// resizes records where in the transcript each mid-session size change took
	// effect (#379), so the replay that renders the screen applies the same
	// sizes at the same points the live terminal did.
	resizes []screenResize
}

func startTranscriptDrain(rw io.ReadWriter, p *spec.PTY) *transcriptDrain {
	t := &transcriptDrain{
		p:         p,
		rw:        rw,
		readDone:  make(chan struct{}),
		screenLen: -1,
		modes:     map[int]bool{},
	}
	queries := newTerminalQueries(p, writerFunc(t.write))
	var modeScan decsetScanner
	go func() {
		defer close(t.readDone)
		buf := make([]byte, 4096)
		// applied counts the size changes already handed to the query emulator.
		// The emulator is resized HERE rather than in markResize so that a chunk
		// is always fed to the emulator at the size that chunk was produced
		// under: a resize recorded while this chunk was in flight belongs after
		// it, and one recorded before it is applied below, before consume (#379).
		applied := 0
		for {
			n, rerr := t.rw.Read(buf)
			if n > 0 {
				// Scan before taking the lock: the scanner is owned by this
				// goroutine, so only the resulting transitions need guarding.
				transitions := modeScan.consume(buf[:n])
				t.mu.Lock()
				// Read the pending resizes BEFORE appending: each was recorded
				// at a transcript offset at or before this chunk's start, so it
				// takes effect ahead of these bytes.
				pending := append([]screenResize(nil), t.resizes[applied:]...)
				t.transcript = append(t.transcript, buf[:n]...)
				for _, m := range transitions {
					t.modes[m.Param] = m.Enabled
				}
				t.mu.Unlock()
				for _, r := range pending {
					queries.resize(r.rows, r.cols)
				}
				applied += len(pending)
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

// currentScreen renders the live screen, returning both the plain text and the
// cells behind it (#382) so a mid-session wait can ask about styling too.
func (t *transcriptDrain) currentScreen() ([]byte, [][]runner.ScreenCell) {
	t.mu.Lock()
	n := len(t.transcript)
	if n == t.screenLen {
		screen := append([]byte(nil), t.screenCache...)
		cells := t.cellsCache
		t.mu.Unlock()
		return screen, cells
	}
	snap := append([]byte(nil), t.transcript...)
	sizes := append([]screenResize(nil), t.resizes...)
	t.mu.Unlock()

	text, cells := renderScreenCells(snap, t.p, sizes)
	rendered := []byte(text)

	t.mu.Lock()
	// Re-check the resize count as well as the length: markResize invalidates
	// the cache by clearing screenLen, and a resize with no new bytes after it
	// would otherwise let this stale render be cached back in.
	if len(t.transcript) == n && len(t.resizes) == len(sizes) {
		t.screenCache = append(t.screenCache[:0], rendered...)
		t.cellsCache = cells
		t.screenLen = n
	}
	t.mu.Unlock()
	return rendered, cells
}

// snapshotResizes copies the recorded size changes for a final render.
func (t *transcriptDrain) snapshotResizes() []screenResize {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]screenResize(nil), t.resizes...)
}

// markResize records that the terminal is about to become rows×cols, anchored
// at the transcript length right now (#379). Taking the length under the same
// lock the reader appends under is what makes the anchor meaningful: everything
// already read is attributed to the old size, and everything read afterwards to
// the new one. Output the program had already emitted but that has not been
// read yet lands after the anchor — which is exactly what a real terminal does
// with bytes still in flight when the window changes, and why the documented
// authoring rule is to settle the screen before resizing.
func (t *transcriptDrain) markResize(rows, cols int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resizes = append(t.resizes, screenResize{offset: len(t.transcript), rows: rows, cols: cols})
	// The cached screen was rendered under the old sizes.
	t.screenLen = -1
	// The live query emulator is NOT resized here. Resizing it from this
	// goroutine could land between the reader appending a chunk and feeding that
	// same chunk to the emulator, so bytes produced at the old size would be
	// replayed at the new one — and a cursor-position request inside them would
	// be answered with coordinates from a screen the program has not seen yet.
	// The reader applies the change instead, in order with the chunks.
}

// modeEnabled reports whether the program currently has a DEC private mode on,
// as of every byte read so far (#378).
func (t *transcriptDrain) modeEnabled(param int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.modes[param]
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
