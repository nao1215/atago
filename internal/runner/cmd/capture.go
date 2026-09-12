package cmd

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/diag"
)

// captureDrainGrace bounds how long Run waits for a capture pipe to reach EOF
// after the command has exited.
//
// It exists because owning the pipes means owning what os/exec was doing for us:
// Cmd.WaitDelay's force-close applies only to pipes os/exec created itself, so
// with atago's own pipes a lingering grandchild would block the drain forever.
// The value mirrors the WaitDelay configureCancellation sets, so a run step that
// backgrounds a long-running program fails the same way and after the same wait
// as it did before (#343, #344).
const captureDrainGrace = 2 * time.Second

// earlyEOFMargin is how far ahead of the command's exit a stream's EOF must be
// before atago calls it early.
//
// EOF and process exit are observed by two different goroutines — the drain's
// io.Copy returning, and Wait returning — so on a loaded CI runner they can land
// a few milliseconds apart in either order for a perfectly ordinary command. 50ms
// is comfortably above that scheduling jitter and far below the gap the case this
// exists for produces (a daemonizing tool holds its stdout closed for the rest of
// its run). A smaller margin would put a phantom note on ordinary failures, which
// is worse than saying nothing.
const earlyEOFMargin = 50 * time.Millisecond

// errDrainDeadline reports that a capture pipe was still open captureDrainGrace
// after the command exited, which happens when something the command started
// inherited the pipe and outlived it. It stands in for exec.ErrWaitDelay, which
// os/exec no longer produces for these streams now that atago owns them: the
// condition and the advice are identical, so captureFailure treats them alike
// and the message a user sees is unchanged.
var errDrainDeadline = diag.CaptureFailed.Errorf("the capture pipe was still open 2s after the command exited")

// capture owns one of the command's output streams end to end: the pipe handed
// to the child, the goroutine draining it, and the bytes it collected.
//
// Owning the pipe (rather than assigning a strings.Builder and letting os/exec
// create the pipe and copier internally) is what makes EOF observable at all.
// os/exec would only ever hand back the final byte count, so a child that wrote
// to a handle that was not our pipe and a child that printed nothing produced the
// identical observation: EOF, zero bytes, nil error (#344).
type capture struct {
	r, w *os.File
	buf  strings.Builder
	done chan struct{}

	// endedAt is when the drain loop finished, eof reports whether it finished at
	// a real end-of-file, and err holds a read failure. They are written by the
	// drain goroutine and are safe to read only after done is closed.
	endedAt time.Time
	eof     bool
	err     error
}

// newCapture allocates the pipe for one stream.
func newCapture() (*capture, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	return &capture{r: r, w: w, done: make(chan struct{})}, nil
}

// writer is the end handed to the child as its stdout or stderr.
func (c *capture) writer() *os.File { return c.w }

// releaseWriter closes the parent's copy of the write end. It MUST be called
// right after Start: until it is, atago itself holds the write end open and the
// read side would never see EOF at all.
func (c *capture) releaseWriter() {
	if c.w != nil {
		_ = c.w.Close()
		c.w = nil
	}
}

// drain reads the stream to EOF in the background.
func (c *capture) drain() {
	go func() {
		defer close(c.done)
		_, err := io.Copy(&c.buf, c.r)
		c.endedAt = time.Now()
		// io.Copy reports EOF as a nil error. Anything else means the loop ended
		// for a reason other than the stream ending — including forceClose below,
		// which is how Run stops a drain that would otherwise never finish.
		c.eof = err == nil
		c.err = err
	}()
}

// waitUntil reports whether the drain finished before deadline.
func (c *capture) waitUntil(deadline time.Time) bool {
	select {
	case <-c.done:
		return true
	case <-time.After(time.Until(deadline)):
		return false
	}
}

// forceClose ends a drain that will not finish on its own, by closing the read
// end under it. The blocked read returns immediately, so the goroutine exits and
// its buffer becomes safe to read — which is why Run always joins after this
// rather than reading the buffer out from under a live goroutine.
func (c *capture) forceClose() { _ = c.r.Close() }

// join blocks until the drain goroutine has returned.
func (c *capture) join() { <-c.done }

// close releases both ends. Safe to call more than once and after forceClose.
func (c *capture) close() {
	c.releaseWriter()
	_ = c.r.Close()
}

// String returns the captured bytes. Only valid once the drain has been joined.
func (c *capture) String() string { return c.buf.String() }

// earlyBy reports how far ahead of exitedAt this stream reached EOF, and whether
// that gap is large enough to be meaningful. A drain that did not end at a real
// EOF (a forced close, a read error) reports nothing: the stream did not end, it
// was cut off, and that is a different fact already reported as a capture
// failure.
func (c *capture) earlyBy(exitedAt time.Time) (time.Duration, bool) {
	if !c.eof || c.endedAt.IsZero() {
		return 0, false
	}
	d := exitedAt.Sub(c.endedAt)
	if d < earlyEOFMargin {
		return 0, false
	}
	return d, true
}

// earlyEOF builds the Result's per-stream observation: which streams ended
// meaningfully before the command did. It returns nil when neither did, so an
// ordinary command carries no map at all and a failing assertion on it grows no
// note.
func earlyEOF(stdout, stderr *capture, exitedAt time.Time) map[string]time.Duration {
	var out map[string]time.Duration
	for name, c := range map[string]*capture{"stdout": stdout, "stderr": stderr} {
		if d, ok := c.earlyBy(exitedAt); ok {
			if out == nil {
				out = make(map[string]time.Duration, 2)
			}
			out[name] = d
		}
	}
	return out
}
