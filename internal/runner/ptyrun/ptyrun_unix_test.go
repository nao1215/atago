//go:build !windows

package ptyrun

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

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
// hang: waitDrain must return even when something else still holds the terminal
// open, so that a stuck reader fails a run instead of wedging it.
//
// creack/pty performs its ioctls through (*os.File).Fd(), which hands back a
// blocking descriptor, and a read(2) already parked in the kernel is immune to
// Close — the descriptor is only really closed once that read returns. So while
// any slave handle survived (a descendant that escaped the process-group kill,
// or atago's own tty in the start-failure path) waitDrain blocked forever, and
// internal/engine and internal/runner/ptyrun both died on the 5m test timeout.
//
// Where the poller owns the master, closing it ends the read outright, and the
// test holds that to the stronger standard: it must not take the backstop to
// get there.
func TestWaitDrain_ClosingTheMasterEndsTheDrain(t *testing.T) {
	t.Parallel()

	master, tty, err := OpenTerminal(24, 80)
	if err != nil {
		t.Fatalf("OpenTerminal: %v", err)
	}
	// The same question OpenTerminal asks: does closing the master interrupt a
	// read in flight, or does only the backstop end the wait?
	pollerOwned := master.SetReadDeadline(time.Time{}) == nil
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
	started := time.Now()
	go func() {
		defer close(done)
		term.waitDrain(func() { _ = master.Close() }, 0)
	}()

	select {
	case <-done:
	case <-time.After(closeGrace + 30*time.Second):
		t.Fatal("waitDrain never returned after the master was closed while the terminal stayed open")
	}
	if elapsed := time.Since(started); pollerOwned && elapsed >= closeGrace {
		t.Errorf("waitDrain took %s: closing a poller-owned master must end the read outright, not wait out the %s backstop", elapsed, closeGrace)
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

// TestOpenTerminal_PairsTheMasterWithItsOwnSlave pins the invariant behind
// #385: the two files OpenTerminal returns must be the two ends of ONE
// terminal. They were not always — creack/pty's Linux ptsname hands the ioctl
// its answer buffer as a bare uintptr, and a goroutine stack that moves before
// the syscall leaves the index reading 0, so atago opened /dev/pts/0 (somebody
// else's terminal) and paired it with a master that had no slave at all. The
// child's output then went to that stranger's terminal and the step reported a
// clean exit 0 with an empty transcript.
//
// A byte written to the tty has to come out of the master; nothing else proves
// the two halves are connected. The loop is there because a wrong pair was
// never the common case: it is cheap enough to run a batch, and a batch is what
// makes a still-broken pairing likely to show up here rather than in a pty step.
func TestOpenTerminal_PairsTheMasterWithItsOwnSlave(t *testing.T) {
	t.Parallel()

	for i := range 100 {
		master, tty, err := OpenTerminal(24, 80)
		if err != nil {
			t.Fatalf("iteration %d: OpenTerminal: %v", i, err)
		}
		if _, werr := tty.Write([]byte("atago\n")); werr != nil {
			t.Fatalf("iteration %d: write to the terminal: %v", i, werr)
		}
		buf := make([]byte, 64)
		n, rerr := master.Read(buf)
		if rerr != nil {
			t.Fatalf("iteration %d: read the master: %v", i, rerr)
		}
		if got := string(buf[:n]); !strings.Contains(got, "atago") {
			t.Fatalf("iteration %d: the master read %q, want the bytes written to its own slave: "+
				"the pair is not a pair", i, got)
		}
		_ = tty.Close()
		_ = master.Close()
	}
}

// TestSetTerminalSize_KeepsTheMasterUnderThePoller pins the other half of the
// same rule: sizing a terminal must not cost it its poller. creack/pty's
// Setsize reaches the descriptor through (*os.File).Fd(), which is documented
// to return it in blocking mode — so a mid-session resize (#379) used to clear
// O_NONBLOCK on the master and leave the drain parked inside read(2), where
// closing the terminal can no longer reach it. atago's ioctls go through
// ControlFD, which borrows the descriptor and leaves the file as it found it.
func TestSetTerminalSize_KeepsTheMasterUnderThePoller(t *testing.T) {
	t.Parallel()

	master, tty, err := OpenTerminal(24, 80)
	if err != nil {
		t.Fatalf("OpenTerminal: %v", err)
	}
	defer func() { _ = tty.Close() }()
	defer func() { _ = master.Close() }()

	// Whether the poller owns this master at all is a platform fact; only where
	// it does is there anything for the resize to take away.
	if err := master.SetReadDeadline(time.Time{}); err != nil {
		t.Skipf("this platform's pty master is not poller-owned: %v", err)
	}
	if err := setTerminalSize(master, 40, 120); err != nil {
		t.Fatalf("setTerminalSize: %v", err)
	}
	if err := master.SetReadDeadline(time.Time{}); err != nil {
		t.Errorf("after a resize the master is no longer poller-owned (%v): "+
			"closing it can no longer interrupt the drain's pending read", err)
	}
	if !nonblocking(t, master) {
		t.Error("after a resize the master lost O_NONBLOCK: its reads park in read(2), " +
			"where Close cannot reach them")
	}
}

// nonblocking reports whether f's descriptor still carries O_NONBLOCK.
func nonblocking(t *testing.T, f *os.File) bool {
	t.Helper()
	on := false
	if err := ControlFD(f, func(fd int) error {
		flags, ferr := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
		if ferr != nil {
			return ferr
		}
		on = flags&unix.O_NONBLOCK != 0
		return nil
	}); err != nil {
		t.Fatalf("reading the descriptor's flags: %v", err)
	}
	return on
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
