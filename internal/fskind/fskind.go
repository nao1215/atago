// Package fskind names what a filesystem entry is. atago tells a spec author
// which kind of thing sits at a path in three unrelated places — a failed dir
// assertion, a refused read, a tree snapshot manifest — and the three used to
// spell the vocabulary separately, so a name could drift while the reader
// assumed one word meant one thing everywhere.
package fskind

import "io/fs"

// Name is the prose name of an entry, for a sentence a spec author reads:
// "%q exists but is a named pipe".
func Name(mode fs.FileMode) string {
	switch {
	case mode.IsRegular():
		return "regular file"
	case mode&fs.ModeDir != 0:
		return "directory"
	case mode&fs.ModeSymlink != 0:
		return "symlink"
	case mode&fs.ModeNamedPipe != 0:
		return "named pipe"
	case mode&fs.ModeSocket != 0:
		return "socket"
	case mode&fs.ModeDevice != 0:
		return "device"
	default:
		return "non-regular entry"
	}
}

// Token is the single-word name a tree snapshot manifest uses as its first
// field. It stays one word (no spaces) because the manifest grammar is
// "<kind> <path>[...]" with one entry per line, so a two-word kind would read as
// part of the path.
func Token(mode fs.FileMode) string {
	switch {
	case mode.IsRegular():
		return "file"
	case mode&fs.ModeDir != 0:
		return "dir"
	case mode&fs.ModeSymlink != 0:
		return "link"
	case mode&fs.ModeNamedPipe != 0:
		return "fifo"
	case mode&fs.ModeSocket != 0:
		return "socket"
	case mode&fs.ModeDevice != 0:
		return "device"
	default:
		return "irregular"
	}
}

// Openable reports whether reading the entry's bytes is a bounded operation.
// Opening a named pipe for reading blocks until another process writes to it,
// and a device node can block or return the host's data, so a path atago walks
// or asserts on must be checked before it is opened: no step timeout covers the
// assertion phase, and a blocked assertion hangs the run to death.
func Openable(mode fs.FileMode) bool {
	return mode.IsRegular()
}
