package snapshot

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNormalize_Idempotent proves normalizing already-normalized output is a
// no-op: Normalize(Normalize(x)) == Normalize(x). A committed golden is the
// normalized form, and Compare re-normalizes the actual before matching, so a
// non-idempotent rule would make a golden fail to match its own source output —
// a spurious failure on every snapshot assertion. Random splices of the exact
// fragments each rule targets (ANSI/OSC, UUID, timestamp, loopback port, the
// workdir) exercise the interactions between rules.
func TestNormalize_Idempotent(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewSource(1))
	frags := []string{
		"\x1b[32m", "\x1b[0m", "\x1b[?25l", "\x1b[?1049h", "\x1b[38:2:1:2:3m",
		"\x1b]0;title\x07", "\x1b]8;;http://x\x1b\\",
		"550e8400-e29b-41d4-a716-446655440000",
		"2026-06-30T09:00:00Z", "2026-06-30 09:00:00", "2026-06-30T09:00:00.123+09:00",
		"127.0.0.1:54321", "localhost:8080", "[::1]:22", "0.0.0.0:1",
		"/tmp/atago-xyz", "plain text ", "\r\n", "\n", "$", "{", "}", "abc",
		"127.0.0.1:", ":", "-", "T", "Z",
	}
	opt := Options{Workdir: "/tmp/atago-xyz"}
	for iter := range 10000 {
		var in []byte
		for n := rng.Intn(8); n > 0; n-- {
			in = append(in, frags[rng.Intn(len(frags))]...)
		}
		once := Normalize(in, opt)
		if twice := Normalize(once, opt); !bytes.Equal(once, twice) {
			t.Fatalf("iter %d: Normalize not idempotent\n in:    %q\n once:  %q\n twice: %q", iter, in, once, twice)
		}
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		opt  Options
		want string
	}{
		{"ansi stripped", "\x1b[32mok\x1b[0m", Options{}, "ok"},
		// A spinner/TUI emits private-mode CSI (cursor hide/show, alt-screen) and
		// colon-subparam SGR; none of it may reach the golden.
		{"csi private-mode stripped", "loading\x1b[?25ldone\x1b[?25h", Options{}, "loadingdone"},
		{"csi alt-screen stripped", "\x1b[?1049hui\x1b[?1049l", Options{}, "ui"},
		{"csi colon sgr stripped", "\x1b[38:2:255:0:0mred\x1b[0m", Options{}, "red"},
		{"osc title stripped", "\x1b]0;my title\x07text", Options{}, "text"},
		{"uuid masked", "id=550e8400-e29b-41d4-a716-446655440000", Options{}, "id=<uuid>"},
		{"timestamp masked", "at 2026-06-30T09:00:00Z done", Options{}, "at <timestamp> done"},
		{"port masked", "listening on 127.0.0.1:54321", Options{}, "listening on 127.0.0.1:<port>"},
		{"single-digit port masked", "127.0.0.1:0 bound", Options{}, "127.0.0.1:<port> bound"},
		{"six-digit port fully masked", "127.0.0.1:123456", Options{}, "127.0.0.1:<port>"},
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

// TestCompare_CRLFGolden is a regression: a golden checked out with CRLF (git
// autocrlf, a CRLF editor) must still match LF-folded actual output, since the
// actual side is CRLF-folded during normalization.
func TestCompare_CRLFGolden(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "golden.txt")
	if err := os.WriteFile(path, []byte("hello\r\nworld\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ok, _, _, err := Compare(path, []byte("hello\nworld\n"), Options{})
	if err != nil {
		t.Fatalf("Compare error = %v", err)
	}
	if !ok {
		t.Error("a CRLF golden did not match LF-folded actual; CRLF must be folded on both sides")
	}
}
