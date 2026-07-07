//go:build windows

package ptyrun

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// TestWindowsCommandLine covers how a pty command becomes the single command
// line ConPTY hands to CreateProcess: a shell step reuses cmd.exe's `/S /C
// "<command>"` contract verbatim, and a shell-free step is tokenized with the
// cmd runner's splitter and re-escaped so the C runtime re-parses it to the same
// argv (plain words stay bare, a path with spaces gets quoted).
func TestWindowsCommandLine(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command string
		shell   bool
		want    string
	}{
		{"shell wraps in cmd /S /C", `echo hi & echo bye`, true, `cmd /S /C "echo hi & echo bye"`},
		{"plain words stay bare", `tool --flag value`, false, `tool --flag value`},
		{"quoted path with spaces re-quotes", `"C:\Program Files\t.exe" run`, false, `"C:\Program Files\t.exe" run`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := windowsCommandLine(c.command, c.shell)
			if err != nil {
				t.Fatalf("windowsCommandLine(%q, %v): unexpected error: %v", c.command, c.shell, err)
			}
			if got != c.want {
				t.Errorf("windowsCommandLine(%q, %v) = %q, want %q", c.command, c.shell, got, c.want)
			}
		})
	}
}

// TestRun_Windows_CapturesOutputAndExit exercises the real ConPTY path on the
// Windows CI runner: a fast-exiting shell command runs inside a pseudo-console,
// its output is captured through the transcript reader, an expect matches it,
// and the child's exit code surfaces as 0. This is the self-contained proof that
// pty steps are no longer POSIX-only.
func TestRun_Windows_CapturesOutputAndExit(t *testing.T) {
	t.Parallel()
	shell := true
	p := &spec.PTY{
		Shell:   &shell,
		Command: "echo hello-conpty",
		Session: []spec.PTYAction{{Expect: "hello-conpty"}},
	}
	res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if ef != nil {
		t.Fatalf("Run: unexpected expect failure: %+v", ef)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (transcript %q)", res.ExitCode, res.Stdout)
	}
	if !strings.Contains(string(res.Stdout), "hello-conpty") {
		t.Errorf("transcript missing the child's output: %q", res.Stdout)
	}
	if !strings.Contains(string(res.Screen), "hello-conpty") {
		t.Errorf("rendered screen missing the child's output: %q", res.Screen)
	}
}

// TestRun_Windows_ExpectTimeoutAborts exercises the abort path: an expect that
// never matches must fail within the session budget (reported as an
// ExpectFailure, not a hard error) and the whole process tree must be killed via
// taskkill, so a long-running child never leaks past the step.
func TestRun_Windows_ExpectTimeoutAborts(t *testing.T) {
	t.Parallel()
	shell := true
	start := time.Now()
	p := &spec.PTY{
		Shell:   &shell,
		Command: "ping -n 20 127.0.0.1",
		Timeout: "2s",
		Session: []spec.PTYAction{{Expect: "a-pattern-that-never-appears"}},
	}
	res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
	if err != nil {
		t.Fatalf("Run: unexpected hard error: %v", err)
	}
	if ef == nil {
		t.Fatalf("expected an ExpectFailure for the never-matching expect, got none (transcript %q)", res.Stdout)
	}
	if !res.TimedOut {
		t.Errorf("result should be marked timed out")
	}
	// The 2s budget must dominate; ping -n 20 would otherwise run ~19s.
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Errorf("session ran %v, want it bounded near the 2s budget (kill did not take)", elapsed)
	}
}
