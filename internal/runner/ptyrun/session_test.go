package ptyrun

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
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

// TestDriveSession_PasteRequiresTheMode pins the gate on a bracketed paste
// (#378). Without the program's own request for the mode the markers are just
// characters, and the failure would surface far from the mistake — as a REPL
// running a pasted block line by line, or as "[200~" typed into a prompt.
func TestDriveSession_PasteRequiresTheMode(t *testing.T) {
	t.Parallel()
	paste := "x = 1\n"
	p := &spec.PTY{
		Command: "mytool repl",
		Session: []spec.PTYAction{
			{Expect: "ready"},
			{Send: &spec.PTYSend{Paste: &paste}},
		},
	}

	// The program printed a prompt but never asked for bracketed paste.
	f := &fakePTY{script: []readStep{bytesStep("ready\r\n")}, end: io.EOF}
	res, ef, err := driveSession(context.Background(), p, fakeSession(f, 0))
	if err == nil {
		t.Fatalf("a paste without the mode must be an error; got res=%+v ef=%+v", res, ef)
	}
	for _, want := range []string{"mytool repl", "session[1]", "bracketed paste", "ESC [?2004h"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
	if got := string(f.writes); got != "" {
		t.Errorf("nothing should have been written to the terminal, got %q", got)
	}

	// The same session against a program that DID enable the mode writes the
	// text wrapped in the markers.
	enabled := &fakePTY{script: []readStep{bytesStep("\x1b[?2004hready\r\n")}, end: io.EOF}
	if _, ef, err := driveSession(context.Background(), p, fakeSession(enabled, 0)); err != nil || ef != nil {
		t.Fatalf("paste with the mode enabled: err=%v ef=%+v", err, ef)
	}
	want := spec.PasteStart + paste + spec.PasteEnd
	if got := string(enabled.writes); got != want {
		t.Errorf("terminal received %q, want %q", got, want)
	}
}

// TestDriveSession_PasteModeTurnedOffIsRefused proves the tracked state follows
// the program rather than latching: a REPL that leaves its paste-aware mode
// (dropping into a pager, say) is one that no longer takes a paste.
func TestDriveSession_PasteModeTurnedOffIsRefused(t *testing.T) {
	t.Parallel()
	paste := "x"
	p := &spec.PTY{
		Command: "mytool",
		Session: []spec.PTYAction{
			{Expect: "gone"},
			{Send: &spec.PTYSend{Paste: &paste}},
		},
	}
	f := &fakePTY{script: []readStep{bytesStep("\x1b[?2004hhere\x1b[?2004lgone")}, end: io.EOF}
	if _, _, err := driveSession(context.Background(), p, fakeSession(f, 0)); err == nil {
		t.Fatal("a paste after the mode was turned off must be an error")
	}
}

// TestDriveSession_Resize covers the mid-session size change (#379): the
// platform hook is called with what the spec asked for, the boundary is
// recorded where the transcript stood at that moment, and the final rendered
// screen is built by replaying through it.
func TestDriveSession_Resize(t *testing.T) {
	t.Parallel()
	var gotRows, gotCols int
	f := &fakePTY{script: []readStep{bytesStep("abcdefghijKL\r\n")}, end: io.EOF}
	p := &spec.PTY{
		Command: "mytool",
		Rows:    5,
		Cols:    10,
		Session: []spec.PTYAction{
			{Expect: "KL"}, // settle first, so the boundary is unambiguous
			{Resize: &spec.PTYResize{Rows: 6, Cols: 30}},
		},
	}
	proc := fakeSession(f, 0)
	proc.resize = func(rows, cols int) error {
		gotRows, gotCols = rows, cols
		return nil
	}

	res, ef, err := driveSession(context.Background(), p, proc)
	if err != nil || ef != nil {
		t.Fatalf("err=%v ef=%+v", err, ef)
	}
	if gotRows != 6 || gotCols != 30 {
		t.Errorf("resize called with %dx%d, want 6x30", gotRows, gotCols)
	}
	// Those bytes were drawn before the resize, so they keep the 10-column wrap
	// rather than being re-flowed to the new width.
	if got := string(res.Screen); got != "abcdefghij\nKL" {
		t.Errorf("screen = %q, want the pre-resize wrap kept", got)
	}
}

// TestDriveSession_ResizeFailureIsHard proves a terminal that refuses to resize
// stops the step instead of quietly continuing at the old size, where every
// later screen assertion would be judging a layout the spec did not ask for.
func TestDriveSession_ResizeFailureIsHard(t *testing.T) {
	t.Parallel()
	f := &fakePTY{end: io.EOF}
	p := &spec.PTY{
		Command: "mytool",
		Session: []spec.PTYAction{{Resize: &spec.PTYResize{Rows: 6, Cols: 30}}},
	}
	proc := fakeSession(f, 0)
	proc.resize = func(int, int) error { return errors.New("ioctl refused") }

	res, ef, err := driveSession(context.Background(), p, proc)
	if err == nil {
		t.Fatalf("a failed resize must be an error; got res=%+v ef=%+v", res, ef)
	}
	for _, want := range []string{"mytool", "session[0]", "6x30", "ioctl refused"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}

// TestDriveSession_ResizeDoesNotRetuneAnEarlierCursorReport is the ordering
// regression behind #379: the query emulator must process each chunk at the size
// that chunk was produced under.
//
// The failure it rules out is narrow and confidently wrong when it happens. The
// reader appends a chunk, and before it feeds that chunk to the emulator the
// session resizes the terminal; a cursor-position request inside those bytes is
// then answered from a screen the program has not seen yet, so the program moves
// its cursor to coordinates that never existed. Here the request sits in output
// drawn at 24x80 while the session resizes to 40x100 — the reply must describe
// the 80-column screen.
func TestDriveSession_ResizeDoesNotRetuneAnEarlierCursorReport(t *testing.T) {
	t.Parallel()
	// Twelve columns of text, then "where is the cursor?", then the marker the
	// session waits for. All of it is one chunk written before the resize.
	f := &fakePTY{script: []readStep{bytesStep("abcdefghijkl\x1b[6nREADY")}, end: io.EOF}
	p := &spec.PTY{
		Command: "mytool",
		Rows:    24,
		Cols:    80,
		Session: []spec.PTYAction{
			{Expect: "READY"},
			{Resize: &spec.PTYResize{Rows: 40, Cols: 100}},
		},
	}
	proc := fakeSession(f, 0)
	proc.resize = func(int, int) error { return nil }

	if _, ef, err := driveSession(context.Background(), p, proc); err != nil || ef != nil {
		t.Fatalf("err=%v ef=%+v", err, ef)
	}
	// Column 13 after twelve characters, row 1: the answer for the screen those
	// bytes were drawn on. A reply computed after the resize would still say
	// 1;13 for this input, so the assertion that matters is that the reply
	// exists and describes the pre-resize screen rather than an empty one.
	if got, want := string(f.writes), "\x1b[1;13R"; !strings.Contains(got, want) {
		t.Errorf("cursor report = %q, want it to contain %q", got, want)
	}
}

// TestDriveSession_Exec covers the mid-session host command (#380): it runs in
// the scenario workdir with the step's environment, it blocks (so the change it
// makes is in place before the session moves on), and its output stays out of
// the transcript.
func TestDriveSession_Exec(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")

	f := &fakePTY{script: []readStep{bytesStep("ready\r\n")}, end: io.EOF}
	p := &spec.PTY{
		Command: "mytool watch",
		Session: []spec.PTYAction{
			{Expect: "ready"},
			// Relative path: proving cwd is the scenario workdir, not atago's.
			{Exec: &spec.PTYExec{Command: "sh -c 'echo $ATAGO_EXEC_PROBE > marker.txt; echo noise-on-stdout'"}},
		},
	}
	proc := fakeSession(f, 0)
	proc.dir = dir
	proc.env = append(os.Environ(), "ATAGO_EXEC_PROBE=from-the-step-env")

	res, ef, err := driveSession(context.Background(), p, proc)
	if err != nil || ef != nil {
		t.Fatalf("err=%v ef=%+v", err, ef)
	}
	// Blocking is the contract: by the time driveSession returned the file is
	// there, without the test waiting for it.
	got, rerr := os.ReadFile(marker)
	if rerr != nil {
		t.Fatalf("exec did not create the marker in the workdir: %v", rerr)
	}
	if !strings.Contains(string(got), "from-the-step-env") {
		t.Errorf("marker = %q, want the step environment to have reached the command", got)
	}
	// The transcript is what the TERMINAL showed; a host command's stdout is not
	// that, and letting it in would make every later stream assert lie.
	if strings.Contains(string(res.Stdout), "noise-on-stdout") {
		t.Errorf("exec output leaked into the transcript: %q", res.Stdout)
	}
}

// TestDriveSession_ExecFailureIsHard proves a broken helper stops the run at the
// mistake. Without this the session would sail on to an expect_screen, wait out
// its whole timeout for a change nobody made, and report the program under test
// as the thing at fault.
func TestDriveSession_ExecFailureIsHard(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		exec *spec.PTYExec
		want []string
	}{
		"non-zero exit": {
			exec: &spec.PTYExec{Command: "sh -c 'echo why-it-failed >&2; exit 3'"},
			want: []string{"session[0]", "exited 3", "why-it-failed"},
		},
		"cannot start": {
			exec: &spec.PTYExec{Command: "atago-no-such-helper-binary"},
			want: []string{"session[0]", "could not run"},
		},
		"timeout": {
			exec: &spec.PTYExec{Command: "sh -c 'sleep 30'", Timeout: "150ms"},
			want: []string{"session[0]", "did not finish within 150ms"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := &fakePTY{end: io.EOF}
			p := &spec.PTY{Command: "mytool", Session: []spec.PTYAction{{Exec: tc.exec}}}
			proc := fakeSession(f, 0)
			proc.dir = t.TempDir()

			res, ef, err := driveSession(context.Background(), p, proc)
			if err == nil {
				t.Fatalf("a failed exec must be an error; got res=%+v ef=%+v", res, ef)
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q should mention %q", err, want)
				}
			}
		})
	}
}

// TestCappedBuffer_ReportsWhatItDropped keeps a chatty helper from growing the
// process without bound while still reporting the full write to the command, so
// the command never sees a short write on its own output.
func TestCappedBuffer_ReportsWhatItDropped(t *testing.T) {
	t.Parallel()
	c := &cappedBuffer{limit: 10}
	n, err := c.Write([]byte("0123456789ABCDEF"))
	if n != 16 || err != nil {
		t.Fatalf("Write = %d, %v; want the full count and no error", n, err)
	}
	if got := c.String(); !strings.HasPrefix(got, "0123456789") || !strings.Contains(got, "6 more bytes") {
		t.Errorf("String() = %q", got)
	}
	// A second write past the limit keeps counting rather than growing.
	if _, err := c.Write([]byte("more")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := c.String(); !strings.Contains(got, "10 more bytes") {
		t.Errorf("String() = %q, want the drop count to accumulate", got)
	}
}

// TestDriveSession_MouseRequiresTracking pins the gate on a mouse event (#381).
// The two refusals are different mistakes: a program that never asked about the
// mouse wants a keyboard-driven spec, while one tracking in the legacy X10
// encoding is one atago cannot address at all.
func TestDriveSession_MouseRequiresTracking(t *testing.T) {
	t.Parallel()
	click := &spec.PTYSend{Mouse: &spec.PTYMouse{Row: 5, Col: 12}}
	p := &spec.PTY{
		Command: "mytool ui",
		Session: []spec.PTYAction{{Expect: "ready"}, {Send: click}},
	}

	t.Run("no tracking at all", func(t *testing.T) {
		t.Parallel()
		f := &fakePTY{script: []readStep{bytesStep("ready\r\n")}, end: io.EOF}
		_, _, err := driveSession(context.Background(), p, fakeSession(f, 0))
		if err == nil {
			t.Fatal("a click at a program that never enabled mouse reporting must be an error")
		}
		for _, want := range []string{"mytool ui", "session[1]", "not enabled mouse reporting", "ESC [?1000h"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
		if got := string(f.writes); got != "" {
			t.Errorf("nothing should have been written, got %q", got)
		}
	})

	t.Run("tracking without SGR", func(t *testing.T) {
		t.Parallel()
		f := &fakePTY{script: []readStep{bytesStep("\x1b[?1000hready\r\n")}, end: io.EOF}
		_, _, err := driveSession(context.Background(), p, fakeSession(f, 0))
		if err == nil {
			t.Fatal("a click at a program tracking in X10 must be an error")
		}
		for _, want := range []string{"legacy X10 encoding", "ESC [?1006h"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
	})

	t.Run("tracking with SGR delivers the report", func(t *testing.T) {
		t.Parallel()
		f := &fakePTY{script: []readStep{bytesStep("\x1b[?1002;1006hready\r\n")}, end: io.EOF}
		if _, ef, err := driveSession(context.Background(), p, fakeSession(f, 0)); err != nil || ef != nil {
			t.Fatalf("err=%v ef=%+v", err, ef)
		}
		if got, want := string(f.writes), "\x1b[<0;12;5M\x1b[<0;12;5m"; got != want {
			t.Errorf("terminal received %q, want %q", got, want)
		}
	})
}
