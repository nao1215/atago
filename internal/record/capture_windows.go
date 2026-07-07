//go:build windows

// Interactive pty recording on Windows (#69). This is the ConPTY capture half:
// it runs the command in a pseudo console wired to the developer's own console,
// forwards keystrokes and output unchanged, and records the session so it can be
// reconstructed as an expect/send spec — the Windows counterpart of the POSIX
// capture in capture_unix.go. The transcript→session generation in pty.go is
// shared across platforms.
//
// One limitation versus POSIX: a ConPTY exposes no per-child echo state, so
// record --pty cannot auto-detect a password prompt to mask its bytes. Recorded
// input is never marked secret on Windows; convert a password send to a
// ${env:...} placeholder by hand.
package record

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/term"

	"github.com/nao1215/atago/internal/conpty"
)

// captureDrainGrace bounds how long capture waits for the pseudo console's final
// output to drain after the child exits before closing it.
const captureDrainGrace = 500 * time.Millisecond

// CapturePTY runs command inside a ConPTY wired to in/out (the developer's
// console), puts in into raw mode, forwards keystrokes and output until the
// program exits, and returns the recorded session (#69). Needs Windows 10 (1809)
// or later for the ConPTY API.
//
// timeout bounds the wait for the child to exit: a program that never exits (a
// server, or a prompt whose quit keystroke was lost) would otherwise hang the
// recorder forever. When it elapses the whole child process tree is killed and
// the transcript captured so far is returned with ErrCaptureTimeout (#194). A
// non-positive timeout falls back to DefaultCaptureTimeout.
func CapturePTY(command string, shell bool, in, out *os.File, timeout time.Duration) (PTYRecording, error) {
	if !conpty.IsAvailable() {
		return PTYRecording{}, fmt.Errorf("record --pty needs ConPTY, which requires Windows 10 version 1809 or later")
	}

	cmdLine, err := conpty.CommandLine(command, shell)
	if err != nil {
		return PTYRecording{}, err
	}

	rows, cols := terminalSize(out)

	// Enable VT output on the developer's console so the ConPTY's ANSI stream
	// renders as intended; restored on return. Raw mode on the input console
	// (below) enables VT input so keys reach the child unbuffered.
	restoreOut := enableVTOutput(out)
	defer restoreOut()

	cpty, err := conpty.Start(cmdLine, "", os.Environ(), rows, cols)
	if err != nil {
		return PTYRecording{}, fmt.Errorf("record --pty: start %q: %w", command, err)
	}
	defer func() { _ = cpty.Close() }()

	rec := PTYRecording{Command: command, Shell: shell, Rows: rows, Cols: cols}
	var mu sync.Mutex

	if oldState, rerr := term.MakeRaw(int(in.Fd())); rerr == nil {
		defer func() { _ = term.Restore(int(in.Fd()), oldState) }()
	}

	// Input: developer keystrokes → child. A ConPTY has no observable echo state,
	// so input is always recorded as non-secret (echoOff=false). This goroutine
	// blocks on in.Read and is abandoned when the one-shot process exits — the
	// child's exit ends the recording, not end-of-input.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := in.Read(buf)
			if n > 0 {
				mu.Lock()
				rec.AppendInput(buf[:n], false)
				mu.Unlock()
				if _, werr := cpty.Write(buf[:n]); werr != nil {
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
			n, rerr := cpty.Read(buf)
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

	// Bound the wait: cpty.Wait returns early when ctx fires, letting a
	// never-exiting child be torn down instead of hanging the recorder.
	ctx, cancel := context.WithTimeout(context.Background(), resolveCaptureTimeout(timeout))
	defer cancel()
	code := cpty.Wait(ctx)
	timedOut := false
	if ctx.Err() != nil {
		// The wait unblocked because the deadline fired, not because the child
		// exited. Kill the whole tree (Windows has no process groups, so taskkill
		// /T walks descendants) so nothing is left running, then record a timeout.
		timedOut = true
		killTree(cpty.Pid())
	}
	// Drain the pseudo console's final bytes before closing it, then stop the
	// output reader. A bounded grace keeps a lingering descendant from hanging us.
	select {
	case <-outDone:
	case <-time.After(captureDrainGrace):
	}
	_ = cpty.Close()
	<-outDone

	mu.Lock()
	rec.ExitCode = code
	mu.Unlock()
	if timedOut {
		return rec, ErrCaptureTimeout
	}
	return rec, nil
}

// killTree force-terminates the child and every descendant. Windows has no
// process groups, so taskkill /T walks the process tree — the closest analog to
// the POSIX capture's kill of the whole Setsid group, so a timed-out session
// never leaks a running child (#194).
func killTree(pid int) {
	// A fire-and-forget teardown; Background is the honest context here.
	_ = exec.CommandContext(context.Background(), "taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run() //nolint:gosec // fixed argv, pid from our own child
}

// enableVTOutput turns on ENABLE_VIRTUAL_TERMINAL_PROCESSING for the developer's
// output console so the ConPTY's ANSI stream renders, and returns a func that
// restores the previous mode. When out is not a console it is a no-op.
func enableVTOutput(out *os.File) func() {
	h := windows.Handle(out.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return func() {}
	}
	if err := windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		return func() {}
	}
	return func() { _ = windows.SetConsoleMode(h, mode) }
}
