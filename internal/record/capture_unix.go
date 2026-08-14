//go:build !windows

// Interactive pty recording (#69): run a command in a real pseudo-terminal
// wired to the developer's own terminal, forward keystrokes and output
// unchanged, and record the session so it can be reconstructed as an
// expect/send spec. This is the POSIX capture half; the pure transcript→session
// generation lives in pty.go.
package record

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/runner/ptyrun"
)

// captureDrainGrace bounds how long capture waits for the pty's final output to
// drain after the child exits before closing the master.
const captureDrainGrace = 500 * time.Millisecond

// CapturePTY runs command inside a pseudo-terminal wired to in/out (the
// developer's terminal), puts in into raw mode, forwards keystrokes and output
// until the program exits, and returns the recorded session (#69). This is the
// POSIX build; capture_windows.go implements the same contract over a ConPTY.
//
// timeout bounds the wait for the child to exit: a program that never exits (a
// server, or a prompt whose quit keystroke was lost) would otherwise hang the
// recorder forever. When it elapses the whole child process tree is killed and
// the transcript captured so far is returned with ErrCaptureTimeout (#194). A
// non-positive timeout falls back to DefaultCaptureTimeout.
func CapturePTY(command string, shell bool, in, out *os.File, timeout time.Duration) (PTYRecording, error) {
	name, args, err := runnercmd.CommandLine(command, shell)
	if err != nil {
		return PTYRecording{}, err
	}

	// Background: the child is bounded by the timeout below (a killed process
	// group), not by context cancellation, so a plain Background satisfies the
	// context-aware exec contract without a redundant second deadline.
	cmd := exec.CommandContext(context.Background(), name, args...) //nolint:gosec // recording the user's declared command is the purpose
	cmd.Env = os.Environ()
	// A fresh session and controlling terminal so the child owns the pty and a
	// stray descendant does not keep it open past the child's exit.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}

	rows, cols := terminalSize(out)
	master, tty, err := ptyrun.OpenTerminal(uint16(rows), uint16(cols)) //nolint:gosec // geometry is bounded by terminalSize
	if err != nil {
		return PTYRecording{}, fmt.Errorf("record --pty: start %q: %w", command, err)
	}
	// releaseTTY drops the recorder's own slave handle. It stays open until the
	// child has been reaped: on macOS the terminal discards whatever it still
	// holds the moment its last slave handle closes, so a command that prints and
	// exits at once would lose its final bytes even to a read already parked in
	// the kernel. Holding one handle here puts that last close after the drain,
	// on a schedule the recorder controls.
	var ttyOnce sync.Once
	releaseTTY := func() { ttyOnce.Do(func() { _ = tty.Close() }) }
	defer releaseTTY()
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	defer func() { _ = master.Close() }()

	rec := PTYRecording{Command: command, Shell: shell, Rows: rows, Cols: cols}
	var mu sync.Mutex

	// Raw mode on the invoking terminal so keystrokes (including control keys and
	// arrows) reach the child unbuffered and unechoed by the outer terminal —
	// the inner pty does its own echo. Restored before we return.
	if oldState, rerr := term.MakeRaw(int(in.Fd())); rerr == nil {
		defer func() { _ = term.Restore(int(in.Fd()), oldState) }()
	}

	// Output: child → developer's screen, recorded verbatim (ANSI intact).
	//
	// reading is closed when the goroutine has nothing left to do but read, and
	// waiting for it below is what puts this reader ahead of the child: creating
	// a goroutine only makes it runnable, so the runtime.Gosched that used to
	// stand here guaranteed nothing about a no-shell fast-exit command that
	// prints and disappears. The pty runner's drain takes the same handoff.
	outDone := make(chan struct{})
	reading := make(chan struct{})
	go func() {
		defer close(outDone)
		buf := make([]byte, 4096)
		close(reading)
		for {
			n, rerr := master.Read(buf)
			if n > 0 {
				mu.Lock()
				rec.AppendOutput(buf[:n])
				mu.Unlock()
				_, _ = out.Write(buf[:n])
			}
			if rerr != nil {
				return
			}
		}
	}()
	<-reading

	if err := cmd.Start(); err != nil {
		// Nothing started, so nothing is waiting to be read: drop the slave handle
		// before waiting, or the master read has no reason to end and the wait
		// below would hang instead of reporting that the child never started.
		releaseTTY()
		_ = master.Close()
		waitDrained(outDone)
		return PTYRecording{}, fmt.Errorf("record --pty: start %q: %w", command, err)
	}

	// Input: developer keystrokes → child. Each Read is one burst; the pty's
	// current ECHO state tags it as a secret (echo off) or not. This goroutine
	// blocks on in.Read and is abandoned when the one-shot process exits — the
	// child's exit is what ends the recording, not end-of-input.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := in.Read(buf)
			if n > 0 {
				echoOff := secretPromptActive(master)
				mu.Lock()
				rec.AppendInput(buf[:n], echoOff)
				mu.Unlock()
				if _, werr := master.Write(buf[:n]); werr != nil {
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	// Reap the child from one goroutine so the wait can be raced against the
	// timeout. The buffered channel lets the timeout path deliver the code even
	// after it kills the tree — the reaper never blocks on a drained receiver.
	waitCh := make(chan int, 1)
	go func() { waitCh <- exitCode(cmd.Wait()) }()

	timedOut := false
	var code int
	select {
	case code = <-waitCh:
	case <-time.After(resolveCaptureTimeout(timeout)):
		timedOut = true
		// Kill the whole process group (Setsid) so a never-exiting child and any
		// descendant are torn down, not just the direct child, then reap it.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		code = <-waitCh
	}

	// The child is reaped, so the handle held since before it started has done
	// its job; drop it now, or the reader has no end-of-terminal to reach and the
	// drain below would spend its whole grace waiting for one.
	releaseTTY()
	// Drain the pty's final bytes before closing the master, then stop the
	// output reader. A bounded grace keeps a lingering descendant from hanging us.
	select {
	case <-outDone:
	case <-time.After(captureDrainGrace):
	}
	_ = master.Close()
	waitDrained(outDone)

	mu.Lock()
	rec.ExitCode = code
	mu.Unlock()
	if timedOut {
		return rec, ErrCaptureTimeout
	}
	return rec, nil
}

// secretPromptActive reports whether the pty is in the termios state of a
// password prompt — echo off while canonical (line) input stays on, the state
// read -s, sudo, and ssh put a terminal in — whose typed bytes must not be
// recorded (#69). ECHO alone is not the signal: a full-screen TUI's raw mode
// (fzf, vim, htop) clears ECHO and ICANON together, and treating that as
// secret turned every keystroke of a recorded TUI session into an
// ${env:ATAGO_SECRET_n} placeholder — a spec that replays nothing. Asking
// through ControlFD rather than master.Fd() is what keeps this poll harmless:
// Fd() would take the master out of the runtime poller and leave its reads
// blocking in read(2), where closing the terminal can no longer interrupt
// them — and this runs on every keystroke burst.
func secretPromptActive(master *os.File) bool {
	secret := false
	if err := ptyrun.ControlFD(master, func(fd int) error {
		t, terr := unix.IoctlGetTermios(fd, ioctlGetTermios)
		if terr != nil {
			return terr
		}
		secret = t.Lflag&unix.ECHO == 0 && t.Lflag&unix.ICANON != 0
		return nil
	}); err != nil {
		return false
	}
	return secret
}

// exitCode extracts a process exit code from cmd.Wait's error (0 on success,
// the signal-derived code otherwise, -1 when it cannot be determined).
func exitCode(waitErr error) int {
	if waitErr == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
