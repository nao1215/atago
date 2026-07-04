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

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
	"golang.org/x/term"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
)

// captureDrainGrace bounds how long capture waits for the pty's final output to
// drain after the child exits before closing the master.
const captureDrainGrace = 500 * time.Millisecond

// CapturePTY runs command inside a pseudo-terminal wired to in/out (the
// developer's terminal), puts in into raw mode, forwards keystrokes and output
// until the program exits, and returns the recorded session (#69). It is
// POSIX-only; the Windows build returns a clear error.
func CapturePTY(command string, shell bool, in, out *os.File) (PTYRecording, error) {
	name, args, err := runnercmd.CommandLine(command, shell)
	if err != nil {
		return PTYRecording{}, err
	}

	// The session runs until the child exits under the developer's own hand, so
	// there is no timeout to bound — Background gives the interactive session no
	// deadline while still satisfying the context-aware exec contract.
	cmd := exec.CommandContext(context.Background(), name, args...) //nolint:gosec // recording the user's declared command is the purpose
	cmd.Env = os.Environ()
	// A fresh session so the child owns the pty and a stray descendant does not
	// keep it open past the child's exit.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	rows, cols := terminalSize(out)
	master, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}) //nolint:gosec // geometry is bounded by terminalSize
	if err != nil {
		return PTYRecording{}, fmt.Errorf("record --pty: start %q: %w", command, err)
	}
	defer func() { _ = master.Close() }()

	rec := PTYRecording{Command: command, Shell: shell, Rows: rows, Cols: cols}
	var mu sync.Mutex

	// Raw mode on the invoking terminal so keystrokes (including control keys and
	// arrows) reach the child unbuffered and unechoed by the outer terminal —
	// the inner pty does its own echo. Restored before we return.
	if oldState, rerr := term.MakeRaw(int(in.Fd())); rerr == nil {
		defer func() { _ = term.Restore(int(in.Fd()), oldState) }()
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
				echoOff := echoDisabled(master)
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

	// Output: child → developer's screen, recorded verbatim (ANSI intact).
	outDone := make(chan struct{})
	go func() {
		defer close(outDone)
		buf := make([]byte, 4096)
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

	waitErr := cmd.Wait()
	// Drain the pty's final bytes before closing the master, then stop the
	// output reader. A bounded grace keeps a lingering descendant from hanging us.
	select {
	case <-outDone:
	case <-time.After(captureDrainGrace):
	}
	_ = master.Close()
	<-outDone

	mu.Lock()
	rec.ExitCode = exitCode(waitErr)
	mu.Unlock()
	return rec, nil
}

// terminalSize returns the invoking terminal's rows/cols, or the pty default
// (24x80) when out is not a terminal.
func terminalSize(out *os.File) (rows, cols int) {
	if r, c, err := pty.Getsize(out); err == nil && r > 0 && c > 0 {
		return r, c
	}
	return 24, 80
}

// echoDisabled reports whether the pty's terminal echo is currently off — the
// signal of a password prompt, whose typed bytes must not be recorded (#69).
func echoDisabled(master *os.File) bool {
	t, err := unix.IoctlGetTermios(int(master.Fd()), ioctlGetTermios)
	if err != nil {
		return false
	}
	return t.Lflag&unix.ECHO == 0
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
