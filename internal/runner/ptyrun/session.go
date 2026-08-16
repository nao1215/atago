package ptyrun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// pollInterval is how often an expect re-checks the accumulated transcript.
const pollInterval = 10 * time.Millisecond

// drainGrace bounds how long finish waits for the reader to hit EOF before
// closing the terminal: an orphaned grandchild that inherited the pty can hold
// it open indefinitely, and its output is not worth hanging for.
const drainGrace = 2 * time.Second

// ptyProcess is the platform-specific pty surface driveSession drives. The
// POSIX runner builds one over creack/pty and the Windows runner over ConPTY,
// then both hand it to driveSession — so the expect/send loop, transcript
// accumulation, and Result shaping live in exactly one place (#8; Windows
// follow-up to #78).
type ptyProcess struct {
	// rw is the terminal master: reads yield the child's output (terminal echo
	// included, ANSI intact), writes deliver sends to the child.
	rw io.ReadWriter
	// trans is an optional already-running terminal drain. The POSIX runner
	// starts it before the child starts to close fast-exit races; other
	// backends can leave it nil and let driveSession start the same drain when
	// the backend cannot start it any earlier.
	trans *transcriptDrain
	// exit receives the child's observed exit code exactly once, when the
	// process is reaped. It must be buffered (cap 1) so the reaper never blocks
	// waiting for a receiver that a kill path drains later.
	exit <-chan int
	// kill force-terminates the whole process tree (a timed-out or aborted
	// session must not leak a running child or its descendants).
	kill func()
	// closeTerm releases the terminal master so the read goroutine unblocks and
	// finish can snapshot a complete transcript.
	closeTerm func()
	// releaseTTY drops the runner's own handle on the terminal's slave side, and
	// must be called only once the child has been reaped. Until it runs, the
	// terminal cannot reach its last close — which is what keeps a fast-exiting
	// child's final bytes readable on a platform that discards them there (see
	// the POSIX runner). It may be nil for a backend with no such handle, and it
	// must be safe to call more than once.
	releaseTTY func()
	// resize changes the terminal size mid-session (#379). The platform decides
	// how the child learns about it — a POSIX TIOCSWINSZ makes the kernel send
	// SIGWINCH to the foreground process group, and ConPTY notifies its client
	// directly — so driveSession only has to say when.
	resize func(rows, cols int) error
	// dir is the resolved workdir the child ran in, surfaced as Result.Workdir.
	dir string
	// env is the environment the pty child was started with, reused for
	// mid-session host commands (#380) so a helper cannot quietly step outside
	// the isolation the step asked for.
	env []string
}

// driveSession runs the platform-neutral half of a pty step: it drains the
// transcript, walks the expect/send/expect_screen session in order, and shapes
// the Result. Every terminal- and process-specific detail is behind proc, so
// POSIX and Windows share this exact control flow. A never-satisfied session
// wait is returned as an ExpectFailure (reported like a failed assertion); only
// "could not start/drive the terminal" conditions are hard errors.
func driveSession(ctx context.Context, p *spec.PTY, proc ptyProcess) (*runner.Result, *ExpectFailure, error) {
	expects, err := compileSession(p.Session)
	if err != nil {
		return nil, nil, diag.InternalError.Errorf("pty: invalid expect regexp: %w", err)
	}

	budget := sessionTimeout(p)
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	term := proc.trans
	if term == nil {
		term = startTranscriptDrain(proc.rw, p)
	}
	d := &sessionDriver{p: p, proc: proc, term: term, start: time.Now()}

	// Drive the session in order. expect polls the transcript; expect_screen
	// polls the rendered screen; send writes to the terminal; an empty send
	// transmits EOF (^D).
	for i, a := range p.Session {
		var out *sessionOutcome
		switch {
		case expects[i] != nil:
			out = d.waitExpect(ctx, expects[i], a.Expect)
		case a.ExpectScreen != nil:
			out = d.waitExpectScreen(ctx, a.ExpectScreen)
		case a.Exec != nil:
			out = d.runExec(ctx, i, a.Exec)
		case a.Resize != nil:
			out = d.applyResize(i, a.Resize)
		case a.Send != nil:
			out = d.send(i, a.Send)
		}
		if out != nil {
			return out.res, out.ef, out.err
		}
	}

	// Session complete: wait for the child to exit within the budget.
	select {
	case code := <-proc.exit:
		out := d.finish(false, code, nil)
		return out.res, out.ef, out.err
	case <-ctx.Done():
		// A parent cancellation is a hard error; a session-budget timeout is a
		// normal timed-out result (#30).
		if errors.Is(ctx.Err(), context.Canceled) {
			out := d.canceled(ctx)
			return out.res, out.ef, out.err
		}
		out := d.abort(nil)
		return out.res, out.ef, out.err
	}
}

// sessionOutcome is how a sessionDriver step ends the session: the triple
// driveSession returns. A nil *sessionOutcome from a step handler means the
// session continues with the next action.
type sessionOutcome struct {
	res *runner.Result
	ef  *ExpectFailure
	err error
}

// sessionDriver holds the state one pty session's actions share: the process
// handles, the transcript drain, and the match offset that makes each expect
// wait for its NEXT occurrence.
type sessionDriver struct {
	p     *spec.PTY
	proc  ptyProcess
	term  *transcriptDrain
	start time.Time
	// matchOffset is the byte index just past the previously matched expect:
	// each expect scans only transcript[matchOffset:], so a pattern that recurs
	// (any shell prompt) waits for its NEXT occurrence instead of matching the
	// stale earlier one.
	matchOffset int
}

// finish shapes the session's Result after the child has been reaped.
func (d *sessionDriver) finish(timedOut bool, code int, ef *ExpectFailure) *sessionOutcome {
	// The child is reaped by now (see the note further down), so the runner's
	// own slave handle has done its job: it kept the terminal from reaching
	// its last close while output was still unread. Drop it here, because
	// until it goes the reader has no EOF to end on and waitDrain would
	// spend its whole grace waiting for one.
	if d.proc.releaseTTY != nil {
		d.proc.releaseTTY()
	}
	// Drain before closing: a fast-exiting child's final output may still sit
	// in the pty buffer, and closing the master discards it. Once the last
	// handle is gone the reader hits EOF and readDone closes on its own;
	// drainGrace bounds the wait in case a descendant kept the terminal open.
	d.term.waitDrain(d.proc.closeTerm, drainGrace)
	tr := d.term.snapshot()
	// A transcript atago knows is incomplete must not become a Result: an
	// expect_screen compared against missing bytes is a confidently wrong
	// verdict, and it would read as the spec's fault rather than atago's. So
	// this outranks even a pending ExpectFailure — that expect may have failed
	// only because the bytes never arrived.
	//
	// Every caller of finish has already reaped the child (the session loop
	// receives proc.exit; abort and failHard kill and wait), so an end-of-
	// session read error accepted here is by construction one that arrived
	// after the child exited. That is why the reader can classify on the error
	// alone without racing the exit.
	rerr := d.term.readError()
	if rerr != nil {
		return &sessionOutcome{err: diag.CaptureFailed.Errorf("pty %q: the terminal transcript is incomplete — reading the terminal failed after %d bytes: %w",
			d.p.Command, len(tr), rerr)}
	}
	screenTextStr, screenCells := renderScreenCells(tr, d.p, d.term.snapshotResizes())
	screenText := []byte(screenTextStr)
	res := &runner.Result{
		Command:  d.p.Command,
		Stdout:   tr,
		Duration: time.Since(d.start),
		Workdir:  d.proc.dir,
		TimedOut: timedOut,
		IsPTY:    true,
		// The rendered screen (#27) is derived from the same bytes, so screen
		// asserts and transcript asserts never disagree about what happened.
		// Replaying through the recorded size changes (#379) keeps that true
		// for a session that resized: each part of the frame is drawn under
		// the size it was produced under.
		Screen: screenText,
		// The same frame with its colors and attributes, for `attrs:` (#382).
		ScreenCells: screenCells,
	}
	if timedOut {
		res.ExitCode = -1
	} else {
		res.ExitCode = code
	}
	return &sessionOutcome{res: res, ef: ef}
}

// abort kills the tree and reaps it, then finishes as timed out.
func (d *sessionDriver) abort(ef *ExpectFailure) *sessionOutcome {
	d.proc.kill()
	<-d.proc.exit
	return d.finish(true, -1, ef)
}

// failHard cleans up (kill, reap, close, drain) before surfacing a hard
// error, so a failed terminal write never leaks the child or goroutines.
func (d *sessionDriver) failHard(err error) *sessionOutcome {
	d.proc.kill()
	<-d.proc.exit
	if d.proc.releaseTTY != nil {
		d.proc.releaseTTY()
	}
	d.term.waitDrain(d.proc.closeTerm, 0)
	return &sessionOutcome{err: err}
}

// canceled surfaces a parent-context cancellation (Ctrl-C / suite cancel) as a
// hard execution error, so the engine stops the scenario instead of asserting
// against a killed terminal — mirroring the cmd runner's cancel/timeout split
// (#30).
func (d *sessionDriver) canceled(ctx context.Context) *sessionOutcome {
	return d.failHard(diag.RunInterrupted.Errorf("pty %q canceled: %w", d.p.Command, ctx.Err()))
}

// waitExpect polls the transcript past the previous match until re matches,
// the session budget runs out, or the run is canceled. nil means matched.
func (d *sessionDriver) waitExpect(ctx context.Context, re *regexp.Regexp, pattern string) *sessionOutcome {
	matched := false
	scannedTo := -1 // transcript length at the last scan; -1 forces one
	for {
		if n := d.term.curLen(); n != scannedTo {
			tail, m := d.term.tailFrom(d.matchOffset)
			scannedTo = m
			if loc := re.FindIndex(tail); loc != nil {
				d.matchOffset += loc[1]
				matched = true
				break
			}
		}
		select {
		case <-ctx.Done():
			// One last check: bytes may have landed in the final poll
			// window before the deadline fired.
			tail, _ := d.term.tailFrom(d.matchOffset)
			if loc := re.FindIndex(tail); loc != nil {
				d.matchOffset += loc[1]
				matched = true
			}
		case <-time.After(pollInterval):
			continue
		}
		break
	}
	if !matched {
		// A parent-context cancellation is an execution error that must stop
		// the scenario; only a genuine session-budget timeout
		// (DeadlineExceeded) becomes an ExpectFailure.
		if errors.Is(ctx.Err(), context.Canceled) {
			return d.canceled(ctx)
		}
		return d.abort(&ExpectFailure{Pattern: pattern, Transcript: string(d.term.snapshot())})
	}
	return nil
}

// waitExpectScreen polls the rendered screen until the matcher holds (and, with
// stable_for, keeps holding), its wait times out, or the run is canceled. nil
// means matched.
func (d *sessionDriver) waitExpectScreen(ctx context.Context, es *spec.PTYExpectScreen) *sessionOutcome {
	waitCtx, cancelWait := sessionWaitContext(ctx, es.Timeout)
	defer cancelWait()
	var matched bool
	stable := &stability{need: parsePositiveDuration(es.StableFor)}
	scannedTo := -1 // transcript length at the last render; -1 forces one
	for {
		if n := d.term.curLen(); n != scannedTo {
			scannedTo = n
			if matched = stable.observe(d.checkScreen(es).OK); matched {
				break
			}
		}
		select {
		case <-waitCtx.Done():
			// One last render: output may have landed in the final poll
			// window before the deadline fired.
			matched = stable.observe(d.checkScreen(es).OK)
		case <-time.After(pollInterval):
			// No new output this tick — but an already-matching screen may
			// have just held still long enough to satisfy stable_for.
			if !stable.heldWithoutRedraw() {
				continue
			}
			matched = true
		}
		break
	}
	if matched {
		return nil
	}
	// A parent-context cancellation is an execution error that must stop
	// the scenario; only a genuine wait timeout becomes a failed check.
	if errors.Is(waitCtx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return d.canceled(ctx)
	}
	cr := d.checkScreen(es)
	if cr.OK && stable.need > 0 {
		screen, _ := d.term.currentScreen()
		cr = &assert.CheckResult{
			Desc:           fmt.Sprintf("pty expect_screen stable for %s", stable.need),
			Expected:       fmt.Sprintf("rendered screen to keep satisfying the matcher for %s", stable.need),
			Actual:         string(screen),
			Hint:           fmt.Sprintf("the rendered screen matched, but not continuously for %s before the timeout elapsed", stable.need),
			ArtifactKind:   "screen",
			ArtifactActual: screen,
		}
	}
	return d.abort(&ExpectFailure{Check: cr})
}

// checkScreen runs the expect_screen matcher against the current rendered
// screen.
func (d *sessionDriver) checkScreen(es *spec.PTYExpectScreen) *assert.CheckResult {
	screen, cells := d.term.currentScreen()
	return checkRenderedScreen(es, screen, cells)
}

// stability tracks an expect_screen stable_for window: the matcher must hold
// continuously for `need` before the wait passes, absorbing redraw churn
// without a blind sleep. A zero need degrades to "matched once".
type stability struct {
	need  time.Duration
	since time.Time // when the screen started matching; zero while not matching
}

// observe feeds one matcher verdict and reports whether the wait is satisfied:
// immediately for a plain match, or once the screen has matched continuously
// for the window.
func (s *stability) observe(ok bool) bool {
	if !ok {
		s.since = time.Time{}
		return false
	}
	if s.need <= 0 {
		return true
	}
	if s.since.IsZero() {
		s.since = time.Now()
	}
	return time.Since(s.since) >= s.need
}

// heldWithoutRedraw reports whether a screen that already matched has now been
// still (no new output to re-render) for the rest of the window.
func (s *stability) heldWithoutRedraw() bool {
	return s.need > 0 && !s.since.IsZero() && time.Since(s.since) >= s.need
}

// runExec runs a mid-session host command. Blocking here is the contract:
// after this returns, the change the command made exists, so whatever the
// session waits for next is waiting on the program noticing it rather than on
// a race (#380).
func (d *sessionDriver) runExec(ctx context.Context, i int, e *spec.PTYExec) *sessionOutcome {
	if xerr := runSessionExec(ctx, e, d.proc.dir, d.proc.env); xerr != nil {
		return d.failHard(fmt.Errorf("pty %q: session[%d]: %w", d.p.Command, i, xerr))
	}
	return nil
}

// applyResize changes the terminal size mid-session (#379). The boundary is
// recorded before changing the size, so the replay that builds the rendered
// screen splits the transcript at the same point the real terminal did.
func (d *sessionDriver) applyResize(i int, r *spec.PTYResize) *sessionOutcome {
	d.term.markResize(r.Rows, r.Cols)
	if rerr := d.proc.resize(r.Rows, r.Cols); rerr != nil {
		return d.failHard(diag.PTYFailed.Errorf("pty %q: session[%d] resize to %dx%d: %w",
			d.p.Command, i, r.Rows, r.Cols, rerr))
	}
	return nil
}

// send writes one send payload to the terminal, first refusing a paste or a
// mouse event the program never asked to receive.
func (d *sessionDriver) send(i int, s *spec.PTYSend) *sessionOutcome {
	// A paste is only a paste if the program asked for one. Without the
	// mode, the markers arrive as ordinary characters and the failure
	// surfaces somewhere far away — as a REPL executing a pasted block
	// line by line, or as "[200~" typed into a prompt — so refuse here,
	// where the mistake is (#378).
	if s.Paste != nil && !d.term.modeEnabled(decsetBracketedPaste) {
		return d.failHard(diag.TerminalModeMismatch.Errorf("pty %q: session[%d] sends a paste, but the program has not enabled bracketed paste "+
			"(it never wrote ESC [?2004h, or turned the mode back off). "+
			"Programs that do not distinguish a paste from typing take a plain send instead; "+
			"if this one does enable the mode, wait for it with an expect or expect_screen before pasting",
			d.p.Command, i))
	}
	// A mouse event only means something to a program that asked to be
	// told about the mouse, and in the encoding atago speaks (#381).
	if s.Mouse != nil {
		if merr := checkMouseMode(d.term, d.p.Command, i); merr != nil {
			return d.failHard(merr)
		}
	}
	// Bytes resolves named keys to their xterm sequences, wraps a paste
	// in its markers, and keeps the historical rule that an empty
	// verbatim send transmits EOF (^D).
	if _, werr := d.term.write(s.Bytes()); werr != nil {
		return d.failHard(diag.PTYFailed.Errorf("pty: send: %w", werr))
	}
	return nil
}

// isSessionEnd reports whether a terminal read error is how a session ends
// rather than how a transcript is lost (#345).
//
// Three things end a pty session normally. io.EOF is the generic one. syscall.EIO
// is what a POSIX master returns once the last slave handle is gone — the child
// exited — and on Windows conpty.Read maps the equivalent ConPTY conditions
// (ERROR_BROKEN_PIPE, ERROR_HANDLE_EOF, ERROR_PIPE_NOT_CONNECTED) to io.EOF, so
// this list stays platform-neutral. fs.ErrClosed covers atago's own
// proc.closeTerm: finish closes the master and then waits for the drain
// goroutine, whose pending read fails precisely because of that close.
//
// Everything else is a read that failed for a reason atago did not cause and
// cannot account for, which means the transcript is short by an unknown amount.
// The service runner already holds this line — its syncBuffer prefixes a
// truncated log with how many bytes it dropped — and truncation must not be
// silent in the runner whose whole output is one accumulated stream.
func isSessionEnd(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.EIO) ||
		errors.Is(err, fs.ErrClosed) ||
		errors.Is(err, io.ErrClosedPipe)
}

func sessionWaitContext(parent context.Context, timeout string) (context.Context, context.CancelFunc) {
	d := parsePositiveDuration(timeout)
	if d <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, d)
}

func parsePositiveDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0
	}
	return d
}

func checkRenderedScreen(es *spec.PTYExpectScreen, screen []byte, cells [][]runner.ScreenCell) *assert.CheckResult {
	return assert.Check(&spec.Assert{Screen: &es.ScreenAssert}, &runner.Result{
		IsPTY:       true,
		Screen:      screen,
		ScreenCells: cells,
	}, assert.Env{})
}
