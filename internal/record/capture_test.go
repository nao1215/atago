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

// TestCapturePTY_RecordsOutputAndExit drives the whole capture path with a
// self-exiting command and no interactive input: start the child in a real pty
// (POSIX) / ConPTY (Windows), drain its output into the recording, and reap its
// exit code. It runs on every OS, so it is the automated backstop for the
// otherwise interactive, human-driven `record --pty` — including the Windows
// ConPTY capture, which has no other test.
func TestCapturePTY_RecordsOutputAndExit(t *testing.T) {
	inR, inW, err := os.Pipe() // an empty input pipe: the command needs no keystrokes
	if err != nil {
		t.Fatalf("input pipe: %v", err)
	}
	defer func() { _ = inR.Close(); _ = inW.Close() }()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("output pipe: %v", err)
	}
	defer func() { _ = outR.Close(); _ = outW.Close() }()
	// Drain the output pipe so the terminal's writes never block on a full buffer.
	go func() { _, _ = io.Copy(io.Discard, outR) }()

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
	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("input pipe: %v", err)
	}
	defer func() { _ = inR.Close(); _ = inW.Close() }()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("output pipe: %v", err)
	}
	defer func() { _ = outR.Close(); _ = outW.Close() }()
	go func() { _, _ = io.Copy(io.Discard, outR) }()

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
	inR, inW, err := os.Pipe() // the child ignores input; the pipe just gives it a stdin
	if err != nil {
		t.Fatalf("input pipe: %v", err)
	}
	defer func() { _ = inR.Close(); _ = inW.Close() }()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("output pipe: %v", err)
	}
	defer func() { _ = outR.Close(); _ = outW.Close() }()
	go func() { _, _ = io.Copy(io.Discard, outR) }()

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
