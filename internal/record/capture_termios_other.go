//go:build !windows && !linux

package record

import "golang.org/x/sys/unix"

// ioctlGetTermios is the request that reads a terminal's attributes on Darwin
// and the BSDs, used to check the pty's ECHO and ICANON flags while recording (#69).
const ioctlGetTermios = unix.TIOCGETA
