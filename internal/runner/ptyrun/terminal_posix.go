//go:build !windows && !linux

package ptyrun

import (
	"os"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

// openTerminalPair opens a pseudo-terminal pair. Everywhere but Linux this is
// creack/pty's Open verbatim: the pointer hazard that made Linux need its own
// copy (see terminal_linux.go) is not present on these platforms, where the
// slave's name is read into a heap allocation rather than a stack variable.
func openTerminalPair() (master, tty *os.File, err error) {
	return pty.Open()
}

// ioctlGetTermios is the request that reads a terminal's attributes on Darwin
// and the BSDs.
const ioctlGetTermios = unix.TIOCGETA
