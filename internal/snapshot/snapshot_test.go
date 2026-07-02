package snapshot

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		opt  Options
		want string
	}{
		{"ansi stripped", "\x1b[32mok\x1b[0m", Options{}, "ok"},
		{"uuid masked", "id=550e8400-e29b-41d4-a716-446655440000", Options{}, "id=<uuid>"},
		{"timestamp masked", "at 2026-06-30T09:00:00Z done", Options{}, "at <timestamp> done"},
		{"port masked", "listening on 127.0.0.1:54321", Options{}, "listening on 127.0.0.1:<port>"},
		{"workdir masked", "wrote /tmp/atago-xyz/out.txt", Options{Workdir: "/tmp/atago-xyz"}, "wrote <workdir>/out.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := string(Normalize([]byte(tt.in), tt.opt)); got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCompareAndUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots", "out.txt")

	// Missing snapshot.
	ok, _, _, err := Compare(path, []byte("hello\n"), Options{})
	if ok || !errors.Is(err, ErrMissing) {
		t.Fatalf("Compare(missing) = ok %v err %v, want ErrMissing", ok, err)
	}

	// Update writes the normalized content and creates parent dirs.
	if err := Update(path, []byte("\x1b[1mhello\x1b[0m\n"), Options{}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	stored, _ := os.ReadFile(path)
	if string(stored) != "hello\n" {
		t.Errorf("stored = %q, want normalized %q", stored, "hello\n")
	}

	// Compare now matches (actual gets normalized too).
	ok, _, _, err = Compare(path, []byte("\x1b[31mhello\x1b[0m\n"), Options{})
	if err != nil || !ok {
		t.Fatalf("Compare(match) = ok %v err %v, want ok", ok, err)
	}

	// A real difference fails.
	ok, exp, act, err := Compare(path, []byte("goodbye\n"), Options{})
	if err != nil || ok {
		t.Fatalf("Compare(diff) = ok %v err %v, want not ok", ok, err)
	}
	if !strings.Contains(exp, "hello") || !strings.Contains(act, "goodbye") {
		t.Errorf("expected/actual = %q/%q", exp, act)
	}
}

// TestNormalize_CRLF: line endings are an OS artifact — a snapshot recorded on
// POSIX must match cmd.exe (CRLF) output on Windows.
func TestNormalize_CRLF(t *testing.T) {
	t.Parallel()
	got := string(Normalize([]byte("hello\r\nworld\r\n"), Options{}))
	if got != "hello\nworld\n" {
		t.Errorf("Normalize CRLF = %q, want LF-only", got)
	}
}
