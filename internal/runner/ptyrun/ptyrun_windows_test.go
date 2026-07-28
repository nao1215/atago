//go:build windows

package ptyrun

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

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

// TestRun_Windows_CaptureSurvivesConcurrentConPTYSessions puts #339's leading
// hypothesis under test.
//
// On windows-latest, one scenario of the portable self-hosted subset failed with
// a nested `atago run` that exited 0 while its stdout came back EMPTY — output
// the child certainly produced, since a run that exits 0 always prints a PASSED
// summary. It reproduced once in fourteen runs and never on Linux or macOS. The
// subset runs that spec alongside pty_portable.atago.yaml at the default
// parallelism, and attaching a pseudoconsole is a process-global operation on
// Windows, so a child spawned while a ConPTY is being set up misrouting its
// handles is the plausible mechanism. Nothing in the CI log proved the two
// overlapped, so this test forces the overlap that CI only stumbles into.
//
// What it asserts is the invariant the flake broke: a command that writes to
// stdout and exits 0 must never come back with an empty capture. Since #344 the
// runner can also say WHY a stream is empty — EOF on a capture pipe cannot
// precede the child's exit unless the child's stdout was not atago's pipe — so
// an early-EOF observation here would be direct evidence for the hypothesis
// rather than another unreadable empty string.
//
// A green run does not prove the hypothesis wrong (the flake is rare and this
// forces only tens of overlaps, not thousands). It is a cheap standing guard on
// the exact interaction, and if it ever fails it hands over the diagnosis.
func TestRun_Windows_CaptureSurvivesConcurrentConPTYSessions(t *testing.T) {
	t.Parallel()

	const (
		ptyWorkers = 3  // concurrent ConPTY sessions being set up and torn down
		captures   = 40 // captured child processes racing against them
	)
	shell := true
	stop := make(chan struct{})
	var ptyWG sync.WaitGroup

	// Churn pseudo-consoles for the whole life of the capture loop: the
	// suspected window is the setup/teardown of a ConPTY, not a session at rest.
	for range ptyWorkers {
		ptyWG.Add(1)
		go func() {
			defer ptyWG.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_, _, _ = Run(context.Background(), &spec.PTY{
					Shell:   &shell,
					Command: "echo churn",
					Timeout: "10s",
					Session: []spec.PTYAction{{Expect: "churn"}},
				}, t.TempDir(), nil)
			}
		}()
	}

	runner := runnercmd.New()
	wd := t.TempDir()
	for i := range captures {
		res, err := runner.Run(context.Background(), &spec.Run{
			Command: `cmd /c echo produced`,
		}, wd)
		if err != nil {
			close(stop)
			ptyWG.Wait()
			t.Fatalf("capture %d: Run() error = %v", i, err)
		}
		if res.ExitCode != 0 {
			t.Errorf("capture %d: exit code = %d, want 0", i, res.ExitCode)
		}
		if !strings.Contains(string(res.Stdout), "produced") {
			// This is the #339 observation. EarlyEOF, when present, says the
			// child's stdout was closed or was never connected to atago's pipe.
			t.Errorf("capture %d: stdout = %q, want it to contain %q (early EOF: %v)",
				i, res.Stdout, "produced", res.EarlyEOF)
		}
		if len(res.EarlyEOF) != 0 {
			t.Errorf("capture %d: a stream ended before the command exited: %v", i, res.EarlyEOF)
		}
	}

	close(stop)
	ptyWG.Wait()
}
