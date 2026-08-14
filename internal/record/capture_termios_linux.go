//go:build linux

package record

import "golang.org/x/sys/unix"

// ioctlGetTermios is the request that reads a terminal's attributes on Linux,
// used to check the pty's ECHO and ICANON flags while recording (#69).
const ioctlGetTermios = unix.TCGETS
