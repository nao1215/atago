package record

import (
	"os"

	"golang.org/x/term"
)

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
