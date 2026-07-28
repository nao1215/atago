package ptyrun

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// readStep is one scripted outcome of a Read: some bytes, or an error, in the
// order the drain loop will see them.
type readStep struct {
	data []byte
	err  error
}

func bytesStep(s string) readStep { return readStep{data: []byte(s)} }
func errStep(err error) readStep  { return readStep{err: err} }

// fakePTY is a scripted terminal master: each Read returns the next step, and
// once the script is exhausted it returns end forever (or blocks until
// closeTerm when end is nil). It lets driveSession's drain loop be driven
// deterministically — including down error paths a real pty cannot be made to
// take on demand.
type fakePTY struct {
	mu     sync.Mutex
	script []readStep
	// end terminates the read loop once the script runs out. Nil means the
	// terminal simply has nothing more to say until it is closed.
	end error
	// closed reports that closeTerm was called; a Read after that fails the way
	// a read on a closed handle does.
	closed bool
	writes []byte
}

func (f *fakePTY) Read(p []byte) (int, error) {
	for {
		f.mu.Lock()
		if f.closed {
			f.mu.Unlock()
			return 0, io.ErrClosedPipe
		}
		if len(f.script) > 0 {
			step := f.script[0]
			f.script = f.script[1:]
			f.mu.Unlock()
			if step.err != nil {
				return 0, step.err
			}
			return copy(p, step.data), nil
		}
		end := f.end
		f.mu.Unlock()
		if end != nil {
			return 0, end
		}
		// Nothing left and no terminating error: block until closed rather than
		// spinning, mirroring a live terminal with nothing to say.
		time.Sleep(pollInterval)
	}
}

func (f *fakePTY) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes = append(f.writes, p...)
	return len(p), nil
}

func (f *fakePTY) close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
}

// fakeSession wires a fakePTY into a ptyProcess whose child has already exited
// with code, which is the state driveSession's finish path always runs in.
func fakeSession(f *fakePTY, code int) ptyProcess {
	exit := make(chan int, 1)
	exit <- code
	return ptyProcess{
		rw:        f,
		exit:      exit,
		kill:      func() {},
		closeTerm: f.close,
		dir:       "/tmp/fake",
	}
}

// TestDriveSession_ReadErrorIsNotSilence is the #345 regression. Every read
// error used to end the drain loop identically, so a transcript lost to a real
// read failure was indistinguishable from a session that simply ended — and
// every screen/transcript assertion afterwards ran against a partial transcript
// atago believed was complete. A lost transcript must be a hard error naming the
// pty step, not a confidently wrong assertion.
func TestDriveSession_ReadErrorIsNotSilence(t *testing.T) {
	t.Parallel()
	f := &fakePTY{script: []readStep{bytesStep("hello")}, end: syscall.EBADF}
	p := &spec.PTY{Command: "mytool"}

	res, ef, err := driveSession(context.Background(), p, fakeSession(f, 0))
	if err == nil {
		t.Fatalf("a lost transcript must be an error; got res=%+v ef=%+v", res, ef)
	}
	for _, want := range []string{"pty", "mytool", "incomplete"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
	// The contract, pinned so it cannot silently change: no Result reaches the
	// caller, so no assertion can run against the partial transcript. The bytes
	// read before the failure are named in the message instead.
	if res != nil {
		t.Errorf("a lost transcript must not hand back a Result: %+v", res)
	}
	if ef != nil {
		t.Errorf("a lost transcript is not an expect failure: %+v", ef)
	}
	if !strings.Contains(err.Error(), "5 bytes") {
		t.Errorf("error should say how much was read before the failure: %v", err)
	}
}

// TestDriveSession_NormalEndsStayNormal proves the classification did not turn
// an ordinary end of session into an error: a pty master returns EOF or, on
// POSIX, EIO once the child is gone, and both are how every green session ends.
func TestDriveSession_NormalEndsStayNormal(t *testing.T) {
	t.Parallel()
	for name, endErr := range map[string]error{
		"eof": io.EOF,
		"eio": syscall.EIO,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := &fakePTY{script: []readStep{bytesStep("hello")}, end: endErr}
			res, ef, err := driveSession(context.Background(), &spec.PTY{
				Command: "mytool",
				Session: []spec.PTYAction{{Expect: "hello"}},
			}, fakeSession(f, 0))
			if err != nil {
				t.Fatalf("a normal end must not error: %v", err)
			}
			if ef != nil {
				t.Fatalf("expect should have matched: %+v", ef)
			}
			if string(res.Stdout) != "hello" {
				t.Errorf("transcript = %q, want %q", res.Stdout, "hello")
			}
			if res.ExitCode != 0 {
				t.Errorf("exit code = %d, want 0", res.ExitCode)
			}
		})
	}
}

// TestDriveSession_DeliberateCloseIsNotALostTranscript covers the path every
// session takes: finish closes the terminal itself and then waits for the drain
// goroutine, whose pending Read fails precisely because of that close. Treating
// atago's own close as a lost transcript would turn every green pty step red.
func TestDriveSession_DeliberateCloseIsNotALostTranscript(t *testing.T) {
	t.Parallel()
	// No terminating error: the fake blocks until closeTerm, so the read that
	// ends the loop is the one atago's own close interrupts.
	f := &fakePTY{script: []readStep{bytesStep("ready")}}
	res, ef, err := driveSession(context.Background(), &spec.PTY{
		Command: "mytool",
		Session: []spec.PTYAction{{Expect: "ready"}},
	}, fakeSession(f, 0))
	if err != nil {
		t.Fatalf("a deliberate close must not read as a lost transcript: %v", err)
	}
	if ef != nil {
		t.Fatalf("expect should have matched: %+v", ef)
	}
	if string(res.Stdout) != "ready" {
		t.Errorf("transcript = %q, want %q", res.Stdout, "ready")
	}
}

// TestDriveSession_EINTRIsRetried proves an interrupted read is resumed rather
// than mistaken for the end of the session: a signal arriving mid-read must not
// truncate the transcript, silently or loudly.
func TestDriveSession_EINTRIsRetried(t *testing.T) {
	t.Parallel()
	// The interrupt lands BETWEEN the two chunks, so the tail is delivered only
	// if the drain loop survives it.
	f := &fakePTY{
		script: []readStep{bytesStep("be"), errStep(syscall.EINTR), bytesStep("fore")},
		end:    io.EOF,
	}
	res, _, err := driveSession(context.Background(), &spec.PTY{
		Command: "mytool",
		Session: []spec.PTYAction{{Expect: "before"}},
	}, fakeSession(f, 0))
	if err != nil {
		t.Fatalf("EINTR must be retried, not treated as an end: %v", err)
	}
	if string(res.Stdout) != "before" {
		t.Errorf("transcript = %q, want %q (EINTR truncated it)", res.Stdout, "before")
	}
}

// TestDriveSession_LostTranscriptOutranksAnExpectFailure: when the transcript is
// incomplete, the expect that failed against it may have failed only because the
// bytes are missing. Reporting it as an assertion failure would blame the spec
// for atago's own lost data, so the hard error wins.
func TestDriveSession_LostTranscriptOutranksAnExpectFailure(t *testing.T) {
	t.Parallel()
	f := &fakePTY{script: []readStep{bytesStep("partial")}, end: syscall.EBADF}
	res, ef, err := driveSession(context.Background(), &spec.PTY{
		Command: "mytool",
		Timeout: "200ms",
		Session: []spec.PTYAction{{Expect: "never-arrives"}},
	}, fakeSession(f, 0))
	if err == nil {
		t.Fatalf("a lost transcript must outrank the expect failure; got res=%+v ef=%+v", res, ef)
	}
	if !errors.Is(err, syscall.EBADF) {
		t.Errorf("error should wrap the underlying read failure: %v", err)
	}
}
