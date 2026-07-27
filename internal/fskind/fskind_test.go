package fskind

import (
	"io/fs"
	"testing"
)

// TestNameAndToken pins the two vocabularies against each other: every kind has
// a prose name for a failure sentence and a single-word token for the tree
// manifest, and the token never carries a space (the manifest grammar puts the
// path in the next field).
func TestNameAndToken(t *testing.T) {
	t.Parallel()
	cases := []struct {
		mode  fs.FileMode
		name  string
		token string
	}{
		{0, "regular file", "file"},
		{0o755, "regular file", "file"},
		{fs.ModeDir, "directory", "dir"},
		{fs.ModeSymlink, "symlink", "link"},
		{fs.ModeNamedPipe, "named pipe", "fifo"},
		{fs.ModeSocket, "socket", "socket"},
		{fs.ModeDevice, "device", "device"},
		{fs.ModeDevice | fs.ModeCharDevice, "device", "device"},
		{fs.ModeIrregular, "non-regular entry", "irregular"},
	}
	for _, tt := range cases {
		if got := Name(tt.mode); got != tt.name {
			t.Errorf("Name(%v) = %q, want %q", tt.mode, got, tt.name)
		}
		if got := Token(tt.mode); got != tt.token {
			t.Errorf("Token(%v) = %q, want %q", tt.mode, got, tt.token)
		}
		for _, r := range Token(tt.mode) {
			if r == ' ' {
				t.Errorf("Token(%v) = %q contains a space", tt.mode, Token(tt.mode))
			}
		}
	}
}

// TestOpenable pins the one question the read paths ask: only a regular file may
// be opened. A pipe blocks until a writer appears, and no step timeout covers an
// assertion, so opening one hangs a run to death.
func TestOpenable(t *testing.T) {
	t.Parallel()
	if !Openable(0) || !Openable(0o600) {
		t.Error("a regular file must be openable")
	}
	for _, mode := range []fs.FileMode{fs.ModeDir, fs.ModeSymlink, fs.ModeNamedPipe, fs.ModeSocket, fs.ModeDevice, fs.ModeIrregular} {
		if Openable(mode) {
			t.Errorf("Openable(%v) = true, want false", mode)
		}
	}
}
