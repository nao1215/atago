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
func CapturePTY(command string, shell bool, in, out *os.File) (PTYRecording, error) {
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

	code := cpty.Wait(context.Background())
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
	return rec, nil
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
