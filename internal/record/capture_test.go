package record

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/runner/ptyrun"
	"github.com/nao1215/atago/internal/spec"
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

// capturePipesMarker is capturePipes plus a synchronization point: its drain
// scans the output stream for marker and closes the returned channel the first
// time it appears. A test that must type only AFTER the child has changed the
// terminal's state (raw mode, echo off) waits on the channel instead of
// sleeping, so the echo-state tag the recorder attaches to the input is
// deterministic. It also returns the input pipe's write end so the test can
// type. The teardown order is the same one capturePipes documents (#406).
func capturePipesMarker(t *testing.T, marker string) (in, inWrite, out *os.File, seen <-chan struct{}) {
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
	found := make(chan struct{})
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		var window []byte
		buf := make([]byte, 4096)
		notified := false
		for {
			n, rerr := outR.Read(buf)
			if n > 0 && !notified {
				window = append(window, buf[:n]...)
				if strings.Contains(string(window), marker) {
					notified = true
					close(found)
				}
			}
			if rerr != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = outW.Close() // the drain now reaches EOF
		_ = inR.Close()
		_ = inW.Close()
		select {
		case <-drained:
			_ = outR.Close()
		case <-time.After(drainTeardownGrace):
			t.Errorf("the output drain did not finish %s after its write end closed: "+
				"something still holds the pipe open, and closing the read end would hang", drainTeardownGrace)
		}
	})
	return inR, inW, outW, found
}

// waitMarker bounds the wait for a capturePipesMarker channel so a child that
// never prints its marker fails the test by name instead of wedging it.
func waitMarker(t *testing.T, seen <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-seen:
	case <-time.After(15 * time.Second):
		t.Fatalf("child never printed its %s marker", what)
	}
}

// inputSegments collects the input bursts of a recording in order.
func inputSegments(rec PTYRecording) []PTYSegment {
	var in []PTYSegment
	for _, seg := range rec.Segments {
		if seg.Input != nil {
			in = append(in, seg)
		}
	}
	return in
}

// TestCapturePTY_RawModeKeystrokeIsNotSecret is a regression test for recorded
// TUI sessions: a full-screen program's raw mode (fzf, vim, htop) clears ECHO
// and ICANON together, and tagging input by ECHO alone turned every keystroke
// of a recorded TUI session into an ${env:ATAGO_SECRET_n} placeholder — a spec
// that replays nothing. Only a canonical (line-mode) prompt with echo off is a
// password prompt; a raw-mode keystroke must be recorded literally.
func TestCapturePTY_RawModeKeystrokeIsNotSecret(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stty and termios echo state are POSIX; a ConPTY exposes no echo state at all")
	}
	inR, inW, outW, ready := capturePipesMarker(t, "READY")

	// A TUI stand-in: enter raw-ish mode (echo and canonical input both off),
	// announce READY, read one keystroke, restore the terminal, and exit.
	const cmd = "stty -icanon -echo min 1 time 0; printf READY; head -c 1 >/dev/null; stty sane; echo BYE"
	recCh := make(chan PTYRecording, 1)
	errCh := make(chan error, 1)
	go func() {
		rec, err := CapturePTY(cmd, true, inR, outW, 30*time.Second)
		recCh <- rec
		errCh <- err
	}()

	waitMarker(t, ready, "raw-mode")
	if _, err := inW.WriteString("j"); err != nil {
		t.Fatalf("typing into the recorder: %v", err)
	}

	rec := <-recCh
	if err := <-errCh; err != nil {
		t.Fatalf("CapturePTY() error = %v", err)
	}
	inputs := inputSegments(rec)
	if len(inputs) != 1 {
		t.Fatalf("input segments = %d, want 1 (%+v)", len(inputs), rec.Segments)
	}
	if inputs[0].EchoOff {
		t.Errorf("a raw-mode (TUI) keystroke was tagged as secret input")
	}
	generated, err := GeneratePTY(rec, Options{SuiteName: "tui"})
	if err != nil {
		t.Fatalf("GeneratePTY: %v", err)
	}
	if strings.Contains(string(generated), "ATAGO_SECRET") {
		t.Errorf("a raw-mode (TUI) keystroke was rendered as a secret placeholder:\n%s", generated)
	}
	if !strings.Contains(string(generated), `- send: "j"`) {
		t.Errorf("the recorded keystroke should replay literally:\n%s", generated)
	}
}

// TestCapturePTY_CanonicalEchoOffInputIsSecret pins the password half of the
// same distinction: a canonical prompt that disables echo (read -s, sudo, ssh)
// must still mask what was typed, before and after the raw-mode fix.
func TestCapturePTY_CanonicalEchoOffInputIsSecret(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stty and termios echo state are POSIX; a ConPTY exposes no echo state at all")
	}
	inR, inW, outW, ready := capturePipesMarker(t, "Password:")

	const cmd = "stty -echo; printf 'Password: '; read pw; stty echo; echo OK"
	recCh := make(chan PTYRecording, 1)
	errCh := make(chan error, 1)
	go func() {
		rec, err := CapturePTY(cmd, true, inR, outW, 30*time.Second)
		recCh <- rec
		errCh <- err
	}()

	waitMarker(t, ready, "prompt")
	const secret = "hunter2-super-secret"
	if _, err := inW.WriteString(secret + "\n"); err != nil {
		t.Fatalf("typing into the recorder: %v", err)
	}

	rec := <-recCh
	if err := <-errCh; err != nil {
		t.Fatalf("CapturePTY() error = %v", err)
	}
	inputs := inputSegments(rec)
	if len(inputs) != 1 {
		t.Fatalf("input segments = %d, want 1 (%+v)", len(inputs), rec.Segments)
	}
	if !inputs[0].EchoOff {
		t.Errorf("a canonical echo-off (password) burst was not tagged as secret")
	}
	generated, err := GeneratePTY(rec, Options{SuiteName: "login"})
	if err != nil {
		t.Fatalf("GeneratePTY: %v", err)
	}
	if strings.Contains(string(generated), secret) {
		t.Fatalf("SECRET LEAKED into the generated spec:\n%s", generated)
	}
	if !strings.Contains(string(generated), "${env:ATAGO_SECRET_1}") {
		t.Errorf("expected an ${env:...} placeholder for the password:\n%s", generated)
	}
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

// TestExitCode_EveryRunnerAgrees is the cross-runner fence for what an exit code
// MEANS in atago: one table of small programs driven through every runner atago
// has — the cmd runner behind `run:`, the pty runner behind `pty:`, and the
// interactive recorder behind `record --pty` — asserting all three report the
// same code for the same program. A runner that invents its own conversion (a
// signaled child reported as Go's -1 rather than the POSIX 128+signal) fails
// here instead of shipping a second dialect of "exit code", which is how the pty
// runner and the recorder came to disagree with `run:` in the first place.
//
// The timeout row is the other half of the contract: -1 stays the sentinel for
// "atago killed it and there is no exit status to report", so the signal a
// timeout kill delivers must NOT be mapped to 137.
func TestExitCode_EveryRunnerAgrees(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX signals only; ConPTY has no 128+signal convention to agree on")
	}

	cases := []struct {
		name    string
		command string
		timeout time.Duration
		want    int
	}{
		{name: "clean exit", command: "exit 0", want: 0},
		{name: "ordinary failure", command: "exit 7", want: 7},
		{name: "self SIGTERM", command: "kill -TERM $$", want: 143},
		{name: "self SIGINT", command: "kill -INT $$", want: 130},
		{name: "killed by timeout", command: "sleep 60", timeout: 500 * time.Millisecond, want: -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			timeout := c.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			run := &spec.Run{Shell: spec.Bool(true), Command: c.command}
			if c.timeout > 0 {
				run.Timeout = c.timeout.String()
			}
			cmdRes, err := runnercmd.New().Run(context.Background(), run, t.TempDir())
			if err != nil {
				t.Fatalf("run runner: %v", err)
			}
			if cmdRes.ExitCode != c.want {
				t.Errorf("run: exit code = %d, want %d", cmdRes.ExitCode, c.want)
			}

			p := &spec.PTY{Shell: spec.Bool(true), Command: c.command, Timeout: timeout.String()}
			ptyRes, ef, err := ptyrun.Run(context.Background(), p, t.TempDir(), nil)
			if err != nil {
				t.Fatalf("pty runner: %v", err)
			}
			if ef != nil {
				t.Fatalf("pty runner: unexpected expect failure: %+v", ef)
			}
			if ptyRes.ExitCode != c.want {
				t.Errorf("pty: exit code = %d, want %d (the same program through run: reports %d)", ptyRes.ExitCode, c.want, cmdRes.ExitCode)
			}

			inR, outW := capturePipes(t)
			rec, err := CapturePTY(c.command, true, inR, outW, timeout)
			if err != nil && !errors.Is(err, ErrCaptureTimeout) {
				t.Fatalf("recorder: %v", err)
			}
			if rec.ExitCode != c.want {
				t.Errorf("record --pty: exit code = %d, want %d — a recorded session generates a spec a run: step must be able to satisfy", rec.ExitCode, c.want)
			}
		})
	}
}
