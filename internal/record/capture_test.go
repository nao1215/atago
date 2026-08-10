package record

import (
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

// drainTeardownGrace bounds the teardown wait in capturePipes. It is generous —
// the drain ends as soon as the write end closes — and exists only so a future
// regression fails one test by name rather than wedging the package.
const drainTeardownGrace = 30 * time.Second

// capturePipes builds the two pipes a CapturePTY call needs — an input pipe the
// child reads from and an output pipe standing in for the developer's screen —
// drains the output so a terminal write never blocks on a full buffer, and
// registers a teardown that closes them in the ONE order that terminates.
//
// The order is the whole point (#406). The drain goroutine sits parked in a
// read on outR, and on Windows os.Pipe hands back synchronous handles: Close
// cannot interrupt a read already in the kernel, so (*File).Close waits for the
// in-flight read to finish before it returns. Closing outR first therefore
// deadlocks — the read it is waiting for can only end when outW closes, which is
// the statement that never runs — and that is exactly how this file wedged the
// whole package on the 5-minute test timeout in CI. Closing the WRITE end first
// gives the drain EOF; waiting for the drain to exit releases the last reference
// to outR; only then can outR close. POSIX forgives the wrong order because its
// pipe fds are poller-owned and Close unblocks a parked read.
func capturePipes(t *testing.T) (in *os.File, out *os.File) {
	t.Helper()
	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("input pipe: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		_ = inR.Close()
		_ = inW.Close()
		t.Fatalf("output pipe: %v", err)
	}
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		_, _ = io.Copy(io.Discard, outR)
	}()
	t.Cleanup(func() {
		_ = outW.Close() // the drain now reaches EOF
		_ = inR.Close()
		_ = inW.Close()
		select {
		case <-drained: // and has released its reference to outR
			_ = outR.Close()
		case <-time.After(drainTeardownGrace):
			// Never block here. A drain that will not end means outR.Close()
			// would wait forever on Windows, and a teardown that hangs takes the
			// whole PACKAGE down on the go test timeout — which is how one stuck
			// recorder erased the results of every other test in this package.
			// Leak the handle and fail this test by name instead.
			t.Errorf("the output drain did not finish %s after its write end closed: "+
				"something still holds the pipe open, and closing the read end would hang", drainTeardownGrace)
		}
	})
	return inR, outW
}

// TestCapturePTY_RecordsOutputAndExit drives the whole capture path with a
// self-exiting command and no interactive input: start the child in a real pty
// (POSIX) / ConPTY (Windows), drain its output into the recording, and reap its
// exit code. It runs on every OS, so it is the automated backstop for the
// otherwise interactive, human-driven `record --pty` — including the Windows
// ConPTY capture, which has no other test.
func TestCapturePTY_RecordsOutputAndExit(t *testing.T) {
	inR, outW := capturePipes(t) // an empty input pipe: the command needs no keystrokes

	// shell:true → `sh -c` / `cmd /S /C`, so echo is a builtin on both shells.
	rec, err := CapturePTY("echo capture-marker", true, inR, outW, 30*time.Second)
	if err != nil {
		t.Fatalf("CapturePTY() error = %v", err)
	}
	if rec.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", rec.ExitCode)
	}
	var out strings.Builder
	for _, seg := range rec.Segments {
		out.Write(seg.Output)
	}
	if !strings.Contains(out.String(), "capture-marker") {
		t.Errorf("recording missing the child's output: %q", out.String())
	}
}

// TestCapturePTY_StartFailureDoesNotHang is a regression test for a recorder
// hang: `atago record --pty -- <not-a-command>` printed its banner and then
// waited forever. The child never started, so nothing ever wrote to the
// terminal, and the recorder waited on an output reader whose read could not
// end while the recorder itself still held the terminal's slave open — closing
// the master does not interrupt a read already parked in the kernel.
func TestCapturePTY_StartFailureDoesNotHang(t *testing.T) {
	inR, outW := capturePipes(t)

	const missing = "atago-no-such-binary-2a5f1c"
	errCh := make(chan error, 1)
	go func() {
		_, cerr := CapturePTY(missing, false, inR, outW, 30*time.Second)
		errCh <- cerr
	}()

	select {
	case cerr := <-errCh:
		if cerr == nil {
			t.Fatal("CapturePTY() error = nil, want the start failure")
		}
		if !strings.Contains(cerr.Error(), missing) {
			t.Errorf("CapturePTY() error = %v, want it to name the command %q", cerr, missing)
		}
	case <-time.After(60 * time.Second):
		t.Fatal("CapturePTY() hung after the child failed to start")
	}
}

// TestCapturePTY_NoShellFastExitOutputNotLost covers the same fast-exit path as
// the macOS CrossPlatformE2E flake, but in the recorder: a no-shell `echo`
// binary that exits immediately must still leave its output in the recording.
func TestCapturePTY_NoShellFastExitOutputNotLost(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell:false echo is a POSIX path; windows coverage lives in the ConPTY tests")
	}

	for i := range 100 {
		inR, inW, err := os.Pipe()
		if err != nil {
			t.Fatalf("iteration %d: input pipe: %v", i, err)
		}
		outR, outW, err := os.Pipe()
		if err != nil {
			_ = inR.Close()
			_ = inW.Close()
			t.Fatalf("iteration %d: output pipe: %v", i, err)
		}
		done := make(chan struct{})
		go func() {
			defer close(done)
			_, _ = io.Copy(io.Discard, outR)
		}()

		rec, err := CapturePTY("echo capture-marker", false, inR, outW, 30*time.Second)
		_ = outW.Close()
		_ = inR.Close()
		_ = inW.Close()
		<-done
		_ = outR.Close()

		if err != nil {
			t.Fatalf("iteration %d: CapturePTY() error = %v", i, err)
		}
		if rec.ExitCode != 0 {
			t.Fatalf("iteration %d: exit code = %d, want 0", i, rec.ExitCode)
		}
		var out strings.Builder
		for _, seg := range rec.Segments {
			out.Write(seg.Output)
		}
		if !strings.Contains(out.String(), "capture-marker") {
			t.Fatalf("iteration %d: recording missing the child's output: %q", i, out.String())
		}
	}
}

// TestCapturePTY_TimesOutOnNonExitingChild is the cross-platform backstop for
// #194: a program that never exits must not hang the recorder forever. It runs
// a child that ignores stdin and outlives the timeout, and asserts CapturePTY
// returns within a bound carrying ErrCaptureTimeout instead of blocking. It
// exercises both the POSIX pty process-group kill and the Windows ConPTY tree
// kill, since that timeout path has no other automated coverage.
func TestCapturePTY_TimesOutOnNonExitingChild(t *testing.T) {
	inR, outW := capturePipes(t) // the child ignores input; the pipe just gives it a stdin

	// A command that ignores stdin and runs far longer than the timeout, so the
	// only way CapturePTY returns is the timeout firing — not the child exiting.
	// `sleep 60` on POSIX; `ping -n 60 127.0.0.1` is the closest builtin-free
	// Windows equivalent (it runs ~60s and never reads stdin).
	command := "sleep 60"
	if runtime.GOOS == "windows" {
		command = "ping -n 60 127.0.0.1"
	}

	const timeout = 500 * time.Millisecond
	done := make(chan struct{})
	var rec PTYRecording
	var capErr error
	start := time.Now()
	go func() {
		defer close(done)
		rec, capErr = CapturePTY(command, true, inR, outW, timeout)
	}()

	// Generously bound the whole call: the timeout plus teardown (tree kill,
	// drain grace) must complete well under this, or the recorder is hanging.
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatalf("CapturePTY did not return within 15s — it hung on a non-exiting child")
	}

	if !errors.Is(capErr, ErrCaptureTimeout) {
		t.Fatalf("CapturePTY() error = %v, want ErrCaptureTimeout", capErr)
	}
	if elapsed := time.Since(start); elapsed < timeout {
		t.Errorf("returned in %v, before the %v timeout could have elapsed", elapsed, timeout)
	}
	// The killed child never exits cleanly, so the recorded code is non-zero.
	if rec.ExitCode == 0 {
		t.Errorf("exit code = 0, want a non-zero code for a killed child")
	}
}
