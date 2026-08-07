//go:build !windows

package ptyrun

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// TestRun_FastExitOutputNotLost is a regression test for a drain race: a child
// that writes and exits immediately could lose its output when the master was
// closed before the reader goroutine drained the pty buffer (seen as a flaky
// examples/pty.atago.yaml failure under coverage instrumentation). Repeating
// the fast-exit case many times makes the lost-output window reliably visible.
func TestRun_FastExitOutputNotLost(t *testing.T) {
	t.Parallel()

	shell := true
	for i := range 200 {
		p := &spec.PTY{
			Shell:   &shell,
			Command: "if [ -t 0 ]; then echo is-a-tty; else echo is-a-pipe; fi",
		}
		res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if ef != nil {
			t.Fatalf("iteration %d: unexpected expect failure: %+v", i, ef)
		}
		if res.ExitCode != 0 {
			t.Fatalf("iteration %d: exit code = %d, want 0 (stdout %q)", i, res.ExitCode, res.Stdout)
		}
		if !strings.Contains(string(res.Stdout), "is-a-tty") {
			t.Fatalf("iteration %d: transcript lost the child's output: %q", i, res.Stdout)
		}
	}
}

// TestRun_NoShellFastExitOutputNotLost covers the exact CrossPlatformE2E flake
// shape from the scheduled macOS job: a no-session pty step running the real
// `echo` binary exited 0 while its captured stdout came back empty. Repeating
// the no-shell case keeps the regression guard aligned with the self-hosted
// record spec that exercises `record --pty -- echo done`.
func TestRun_NoShellFastExitOutputNotLost(t *testing.T) {
	t.Parallel()

	for i := range 200 {
		p := &spec.PTY{Command: "echo done"}
		res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if ef != nil {
			t.Fatalf("iteration %d: unexpected expect failure: %+v", i, ef)
		}
		if res.ExitCode != 0 {
			t.Fatalf("iteration %d: exit code = %d, want 0 (stdout %q)", i, res.ExitCode, res.Stdout)
		}
		if !strings.Contains(string(res.Stdout), "done") {
			t.Fatalf("iteration %d: transcript lost the child's output: %q", i, res.Stdout)
		}
	}
}

// TestWaitDrain_ClosingTheMasterEndsTheDrain is a regression test for a CI
// hang: closing the pty master must interrupt the reader even when something
// else still holds the terminal open.
//
// creack/pty performs its ioctls through (*os.File).Fd(), which hands back a
// blocking descriptor, and a read(2) already parked in the kernel is immune to
// Close — the descriptor is only really closed once that read returns. So while
// any slave handle survived (a descendant that escaped the process-group kill,
// or atago's own tty in the start-failure path) waitDrain blocked forever, and
// internal/engine and internal/runner/ptyrun both died on the 5m test timeout.
func TestWaitDrain_ClosingTheMasterEndsTheDrain(t *testing.T) {
	t.Parallel()

	master, tty, err := OpenTerminal(24, 80)
	if err != nil {
		t.Fatalf("OpenTerminal: %v", err)
	}
	// Deliberately keep the slave open: this is what makes the master read
	// block instead of ending with EIO.
	defer func() { _ = tty.Close() }()

	term := startTranscriptDrain(master, &spec.PTY{})
	// Park the reader inside read(2) first: a drain closed before its first read
	// ends on the closed descriptor and would prove nothing. Seeing the written
	// bytes land in the transcript means the goroutine consumed them and went
	// back for more.
	if _, werr := tty.Write([]byte("parked\n")); werr != nil {
		t.Fatalf("write to the terminal: %v", werr)
	}
	deadline := time.Now().Add(10 * time.Second)
	for term.curLen() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the drain never read the bytes written to the terminal")
		}
		time.Sleep(time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		defer close(done)
		term.waitDrain(func() { _ = master.Close() }, 0)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("waitDrain never returned after the master was closed while the terminal stayed open")
	}
	// Closing the terminal is how atago ends the session, so the drain must not
	// report it as a lost transcript.
	if rerr := term.readError(); rerr != nil {
		t.Errorf("readError() = %v, want nil: atago's own close is a session end, not a read failure", rerr)
	}
}

// TestRun_StartFailureDoesNotHang covers the other half of the same hang: when
// the child cannot be started at all, Run must surface that error instead of
// waiting on a drain whose read can never end — atago still held the slave.
func TestRun_StartFailureDoesNotHang(t *testing.T) {
	t.Parallel()

	p := &spec.PTY{Command: "atago-no-such-binary-2a5f1c"}
	type outcome struct{ err error }
	ch := make(chan outcome, 1)
	go func() {
		_, _, err := Run(context.Background(), p, t.TempDir(), nil)
		ch <- outcome{err}
	}()

	select {
	case got := <-ch:
		if got.err == nil {
			t.Fatal("Run() error = nil, want the start failure")
		}
		if !strings.Contains(got.err.Error(), p.Command) {
			t.Errorf("Run() error = %v, want it to name the command %q", got.err, p.Command)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Run() hung after the child failed to start")
	}
}

// TestResolveCwd covers cwd resolution for a pty step: empty stays at the
// workdir, an absolute path is used verbatim, and a relative path nests inside
// the workdir — matching the cmd runner's rule so a pty and a run step agree on
// where a relative cwd points.
func TestResolveCwd(t *testing.T) {
	t.Parallel()
	const wd = "/work"
	cases := []struct{ cwd, want string }{
		{"", wd},
		{"sub", "/work" + string(os.PathSeparator) + "sub"},
		{"/abs", "/abs"},
	}
	for _, c := range cases {
		if got := runnercmd.ResolveDir(wd, c.cwd); got != c.want {
			t.Errorf("ResolveDir(%q, %q) = %q, want %q", wd, c.cwd, got, c.want)
		}
	}
}
