package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	shellwords "github.com/mattn/go-shellwords"

	"github.com/nao1215/atago/internal/record"
)

// TestPOSIXJoin_RoundTrip proves posixJoin's output re-tokenizes to the same
// argv through go-shellwords — the tokenizer the cmd runner uses on POSIX
// (#30 follow-up). Boundaries must survive spaces, quotes, the metacharacters
// go-shellwords treats as separators, and backslashes.
func TestPOSIXJoin_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"mytool", "convert", "input.txt"},
		{"mytool", "--msg", "hello world"},
		{"tool", "it's"},
		{"tool", `she said "hi"`},
		{"tool", "spaced arg", "--flag=with space"},
		{"grep", "foo|bar"},
		{"tool", "a>b", "c<d", "e;f", "g&h"},
		{"tool", `C:\tmp\file`},
		{"tool", "$HOME", "*.go"},
		{"tool", ""}, // an empty argument must survive as a distinct entry
	}
	for _, argv := range cases {
		joined := posixJoin(argv)
		got, err := shellwords.Parse(joined)
		if err != nil {
			t.Errorf("%v -> %q does not tokenize: %v", argv, joined, err)
			continue
		}
		if !reflect.DeepEqual(got, argv) {
			t.Errorf("round-trip changed argv: %v -> %q -> %v", argv, joined, got)
		}
	}
}

// TestRecordPTY_ExistingOutExitsEarly proves `record --pty --out FILE` refuses
// an existing --out BEFORE driving any session (#69 follow-up): the check now
// fires first, so it returns ExitConfig without opening a pty (which would hang
// waiting on stdin here) and leaves the existing file untouched.
func TestRecordPTY_ExistingOutExitsEarly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out := filepath.Join(dir, "generated.atago.yaml")
	if err := os.WriteFile(out, []byte("precious"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := recordCmd([]string{"--pty", "--out", out, "--", "echo", "hi"}, &stdout, &stderr)
	if code != ExitConfig {
		t.Fatalf("exit = %d, want %d", code, ExitConfig)
	}
	if !strings.Contains(stderr.String(), "use --force") {
		t.Errorf("stderr = %q, want mention of --force", stderr.String())
	}
	if b, _ := os.ReadFile(out); string(b) != "precious" {
		t.Errorf("existing file was modified: %q", b)
	}
}

// TestPTYExitCode covers the --pty timeout verdict (#194): a clean session is
// ExitOK and silent, while a timed-out one prints a specific, actionable
// message (the elapsed bound plus the --timeout hint) and returns ExitExec so a
// wedged recording never reports success. A non-positive timeout reports the
// default bound, matching the value CapturePTY actually applied.
func TestPTYExitCode(t *testing.T) {
	t.Parallel()

	var ok bytes.Buffer
	if code := ptyExitCode(false, 30*time.Second, &ok); code != ExitOK {
		t.Errorf("clean session exit = %d, want %d", code, ExitOK)
	}
	if ok.Len() != 0 {
		t.Errorf("clean session wrote to stderr: %q", ok.String())
	}

	var timedOut bytes.Buffer
	if code := ptyExitCode(true, 2*time.Second, &timedOut); code != ExitExec {
		t.Errorf("timed-out session exit = %d, want %d", code, ExitExec)
	}
	msg := timedOut.String()
	for _, want := range []string{"did not exit within 2s", "--timeout"} {
		if !strings.Contains(msg, want) {
			t.Errorf("timeout message %q missing %q", msg, want)
		}
	}

	var defaulted bytes.Buffer
	ptyExitCode(true, 0, &defaulted)
	if want := "did not exit within " + record.DefaultCaptureTimeout.String(); !strings.Contains(defaulted.String(), want) {
		t.Errorf("non-positive timeout message %q missing %q", defaulted.String(), want)
	}
}

// TestWindowsJoin_Boundaries proves windowsJoin leaves a whitespace-free path
// verbatim (the e2e-windows regression this fixes: single quotes had been
// treated as part of the filename) while quoting whitespace-bearing tokens.
func TestWindowsJoin_Boundaries(t *testing.T) {
	t.Parallel()
	if got := windowsJoin([]string{`C:\Users\runner\atago.exe`, "version"}); got != `C:\Users\runner\atago.exe version` {
		t.Errorf("plain Windows path must pass through: %q", got)
	}
	if got := windowsJoin([]string{`C:\Program Files\tool.exe`, "run"}); got != `"C:\Program Files\tool.exe" run` {
		t.Errorf("spaced path must be double-quoted verbatim (backslashes literal): %q", got)
	}
}
