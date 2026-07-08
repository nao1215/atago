package snapshot

import (
	"bytes"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/security"
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
		// A bare ESC and an incomplete CSI: removing a complete OSC that sits
		// between them can splice the leftover ESC onto the following bytes and
		// form a fresh CSI escape, which a single-pass CSI-then-OSC strip would
		// leave behind. Fold to a fixed point so idempotence still holds.
		"\x1b", "\x1b[0",
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

// TestNormalize_IdempotentAcrossOSCSplice pins the specific splice that the
// random fuzz above targets: a stray ESC directly before a complete OSC
// sequence whose removal joins the ESC to the bytes that followed the OSC,
// forming a valid CSI escape. Stripping CSI before OSC in a single pass leaves
// that fresh escape in the output, so a second Normalize changes it again and a
// raw ESC leaks into the golden — exactly what the CSI normalizer exists to
// prevent. The fixed-point strip must remove it in one Normalize call.
func TestNormalize_IdempotentAcrossOSCSplice(t *testing.T) {
	t.Parallel()
	cases := []string{
		"a\x1b\x1b]8;;http://x\x1b\\[31mb",
		"\x1b\x1b]0;title\x07[m",
		"link \x1b\x1b]8;;u\x1b\\[31mred\x1b[0m\n",
	}
	for _, in := range cases {
		once := Normalize([]byte(in), Options{})
		twice := Normalize(once, Options{})
		if !bytes.Equal(once, twice) {
			t.Errorf("Normalize not idempotent\n in:    %q\n once:  %q\n twice: %q", in, once, twice)
		}
		if bytes.ContainsRune(once, '\x1b') {
			t.Errorf("Normalize(%q) = %q still contains a raw ESC byte", in, once)
		}
	}
}

// TestNormalize_SecretMaskedAfterCRLFFold proves a declared secret cannot leak
// into a golden when the captured output uses CRLF line endings. Masking runs on
// the raw bytes, but the CRLF fold that follows can reconstruct a multi-line
// secret whose stored value uses LF; without a second masking pass the raw
// credential lands in the committed golden.
func TestNormalize_SecretMaskedAfterCRLFFold(t *testing.T) {
	t.Parallel()
	m := security.NewMasker([]string{"lineA\nlineB"})
	opt := Options{Secrets: m.MaskBytes}
	got := string(Normalize([]byte("lineA\r\nlineB\r\n"), opt))
	if strings.Contains(got, "lineA\nlineB") {
		t.Errorf("secret leaked into normalized output: %q", got)
	}
	if got != "***\n" {
		t.Errorf("Normalize masked CRLF output = %q, want %q", got, "***\n")
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

// TestNormalize_Scrub proves the user Scrub hook rewrites volatile patterns the
// built-in normalizers do not know about (an auto-increment id), and that it runs
// BEFORE the built-ins: a scrub whose placeholder overlaps a built-in target is
// itself left alone by the built-in pass (the placeholder is literal text).
func TestNormalize_Scrub(t *testing.T) {
	t.Parallel()
	reID := regexp.MustCompile(`id=\d+`)
	scrub := func(b []byte) []byte { return reID.ReplaceAllLiteral(b, []byte("id=<ID>")) }
	opt := Options{Scrub: scrub}

	// The auto-increment id is scrubbed; the built-in timestamp masker still runs
	// afterwards on the rest of the line.
	in := "row id=4093 created 2026-06-30T09:00:00Z"
	want := "row id=<ID> created <timestamp>"
	if got := string(Normalize([]byte(in), opt)); got != want {
		t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
	}

	// Secrets run before scrub: a scrub rule sees the already-masked text.
	opt2 := Options{
		Secrets: func(b []byte) []byte { return []byte(strings.ReplaceAll(string(b), "topsecret", "***")) },
		Scrub:   func(b []byte) []byte { return regexp.MustCompile(`\*\*\*`).ReplaceAllLiteral(b, []byte("<REDACTED>")) },
	}
	if got := string(Normalize([]byte("token=topsecret"), opt2)); got != "token=<REDACTED>" {
		t.Errorf("secrets-then-scrub = %q, want %q", got, "token=<REDACTED>")
	}
}

// TestNormalize_ScrubAppliesAfterCRLFFold proves a line-anchored scrub rule
// fires on CRLF output. Before the fix the scrub ran only on the raw bytes,
// where a (?m)^id=\d+$ rule sees "id=42\r" and never matches, so the volatile id
// reached the golden and every later run failed. The rule must also match once
// CRLF is folded to LF.
func TestNormalize_ScrubAppliesAfterCRLFFold(t *testing.T) {
	t.Parallel()
	reID := regexp.MustCompile(`(?m)^id=\d+$`)
	opt := Options{Scrub: func(b []byte) []byte { return reID.ReplaceAllLiteral(b, []byte("id=<ID>")) }}
	got := string(Normalize([]byte("id=42\r\n"), opt))
	if got != "id=<ID>\n" {
		t.Errorf("Normalize(CRLF) = %q, want %q", got, "id=<ID>\n")
	}
	// Idempotent: the folded, scrubbed golden re-normalizes to itself.
	if again := string(Normalize([]byte(got), opt)); again != got {
		t.Errorf("Normalize not idempotent: %q -> %q", got, again)
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

// TestNormalize_PathMaskingBoundaries is a regression for the naive substring
// masking of the home and workdir prefixes: a masked prefix must only replace a
// whole path component, or it corrupts an unrelated sibling path.
func TestNormalize_PathMaskingBoundaries(t *testing.T) {
	t.Run("home does not swallow a longer sibling", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("os.UserHomeDir ignores $HOME on Windows, so the POSIX home path cannot be forced")
		}
		t.Setenv("HOME", "/home/nao")
		// /home/naoki is a different user; only the exact home dir is masked.
		got := string(Normalize([]byte("/home/naoki/project and /home/nao/x"), Options{}))
		if strings.Contains(got, "~ki") {
			t.Errorf("home masking corrupted a sibling path: %q", got)
		}
		if !strings.Contains(got, "~/x") {
			t.Errorf("home dir was not masked: %q", got)
		}
	})
	t.Run("root home does not turn every slash into tilde", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("os.UserHomeDir ignores $HOME on Windows, so a root home cannot be forced")
		}
		t.Setenv("HOME", "/")
		got := string(Normalize([]byte("path /usr/local/bin"), Options{}))
		if got != "path /usr/local/bin" {
			t.Errorf("HOME=/ mangled unrelated paths: %q", got)
		}
	})
	t.Run("workdir does not swallow a punctuation-suffixed sibling", func(t *testing.T) {
		// '+' is a legal filename byte, so /tmp/run1+cache is a different directory
		// and must be left intact; only a separator, whitespace, or end of token is
		// a masking boundary.
		got := string(Normalize([]byte("/tmp/run1/x and /tmp/run1+cache/y"), Options{Workdir: "/tmp/run1"}))
		if !strings.Contains(got, "<workdir>/x") {
			t.Errorf("workdir was not masked: %q", got)
		}
		if !strings.Contains(got, "/tmp/run1+cache/y") {
			t.Errorf("workdir masking corrupted a punctuation-suffixed sibling: %q", got)
		}
	})
	t.Run("workdir does not swallow a prefix-sibling temp dir", func(t *testing.T) {
		got := string(Normalize([]byte("a /tmp/run1/x b /tmp/run10/y"), Options{Workdir: "/tmp/run1"}))
		if !strings.Contains(got, "<workdir>/x") {
			t.Errorf("workdir was not masked: %q", got)
		}
		if !strings.Contains(got, "/tmp/run10/y") {
			t.Errorf("workdir masking corrupted a prefix-sibling: %q", got)
		}
	})
}
