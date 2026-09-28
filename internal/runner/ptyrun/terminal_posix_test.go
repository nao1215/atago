//go:build !windows && !linux

package ptyrun

import "golang.org/x/sys/unix"

// ioctlSetTermios is the request that writes a terminal's attributes on Darwin
// and the BSDs; only tests change them.
const ioctlSetTermios = unix.TIOCSETA
