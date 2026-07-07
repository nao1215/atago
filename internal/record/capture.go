package record

import (
	"errors"
	"os"
	"time"

	"golang.org/x/term"
)

// DefaultCaptureTimeout bounds a `record --pty` session when the caller passes a
// non-positive timeout: an interactive program that never exits (a server, or a
// prompt whose quit keystroke was lost to a read-readiness race) must fail
// loudly instead of hanging the recorder forever (#194). It mirrors the pty:
// spec step's default so the CLI and the spec step agree on "how long is too
// long".
const DefaultCaptureTimeout = 30 * time.Second

// ErrCaptureTimeout is returned by CapturePTY — together with the transcript
// captured so far — when the recorded program does not exit within the timeout.
// The child process tree is killed before it is returned, so the caller can
// still write the partial spec and nothing is left running (#194).
var ErrCaptureTimeout = errors.New("record --pty: the program did not exit within the timeout")

// resolveCaptureTimeout applies DefaultCaptureTimeout when the caller passes a
// non-positive duration, matching sessionTimeout in the pty: runner.
func resolveCaptureTimeout(d time.Duration) time.Duration {
	if d <= 0 {
		return DefaultCaptureTimeout
	}
	return d
}

// terminalSize returns the invoking terminal's rows/cols, or the pty default
// (24x80) when out is not a terminal. golang.org/x/term.GetSize reads a POSIX
// terminal and a Windows console alike, so one helper serves both the POSIX and
// ConPTY capture backends.
func terminalSize(out *os.File) (rows, cols int) {
	if w, h, err := term.GetSize(int(out.Fd())); err == nil && w > 0 && h > 0 {
		return h, w
	}
	return 24, 80
}
