package ptyrun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/assert"
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
	// resize changes the terminal size mid-session (#379). The platform decides
	// how the child learns about it — a POSIX TIOCSWINSZ makes the kernel send
	// SIGWINCH to the foreground process group, and ConPTY notifies its client
	// directly — so driveSession only has to say when.
	resize func(rows, cols int) error
	// dir is the resolved workdir the child ran in, surfaced as Result.Workdir.
	dir string
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
		return nil, nil, fmt.Errorf("pty: invalid expect regexp: %w", err)
	}

	budget := sessionTimeout(p)
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	start := time.Now()
	term := proc.trans
	if term == nil {
		term = startTranscriptDrain(proc.rw, p)
	}
	writeTerm := term.write
	snapshot := term.snapshot
	tailFrom := term.tailFrom
	curLen := term.curLen
	currentScreen := term.currentScreen

	finish := func(timedOut bool, code int, ef *ExpectFailure) (*runner.Result, *ExpectFailure, error) {
		// Drain before closing: a fast-exiting child's final output may still sit
		// in the pty buffer, and closing the master discards it. Once the last
		// handle is gone the reader hits EOF and readDone closes on its own;
		// drainGrace bounds the wait in case a descendant kept the terminal open.
		term.waitDrain(proc.closeTerm, drainGrace)
		tr := snapshot()
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
		rerr := term.readError()
		if rerr != nil {
			return nil, nil, fmt.Errorf(
				"pty %q: the terminal transcript is incomplete — reading the terminal failed after %d bytes: %w",
				p.Command, len(tr), rerr)
		}
		res := &runner.Result{
			Command:  p.Command,
			Stdout:   tr,
			Duration: time.Since(start),
			Workdir:  proc.dir,
			TimedOut: timedOut,
			IsPTY:    true,
			// The rendered screen (#27) is derived from the same bytes, so screen
			// asserts and transcript asserts never disagree about what happened.
			// Replaying through the recorded size changes (#379) keeps that true
			// for a session that resized: each part of the frame is drawn under
			// the size it was produced under.
			Screen: []byte(renderScreenResized(tr, p, term.snapshotResizes())),
		}
		if timedOut {
			res.ExitCode = -1
		} else {
			res.ExitCode = code
		}
		return res, ef, nil
	}

	// abort kills the tree and reaps it, then finishes as timed out.
	abort := func(ef *ExpectFailure) (*runner.Result, *ExpectFailure, error) {
		proc.kill()
		<-proc.exit
		return finish(true, -1, ef)
	}

	// failHard cleans up (kill, reap, close, drain) before surfacing a hard
	// error, so a failed terminal write never leaks the child or goroutines.
	failHard := func(err error) (*runner.Result, *ExpectFailure, error) {
		proc.kill()
		<-proc.exit
		term.waitDrain(proc.closeTerm, 0)
		return nil, nil, err
	}

	// canceledResult surfaces a parent-context cancellation (Ctrl-C / suite
	// cancel) as a hard execution error, so the engine stops the scenario instead
	// of asserting against a killed terminal — mirroring the cmd runner's
	// cancel/timeout split (#30).
	canceledResult := func() (*runner.Result, *ExpectFailure, error) {
		return failHard(fmt.Errorf("pty %q canceled: %w", p.Command, ctx.Err()))
	}

	// Drive the session in order. expect polls the transcript; expect_screen
	// polls the rendered screen; send writes to the terminal; an empty send
	// transmits EOF (^D).
	//
	// matchOffset is the byte index just past the previously matched expect: each
	// expect scans only transcript[matchOffset:], so a pattern that recurs (any
	// shell prompt) waits for its NEXT occurrence instead of matching the stale
	// earlier one.
	matchOffset := 0
	for i, a := range p.Session {
		if expects[i] != nil {
			matched := false
			scannedTo := -1 // transcript length at the last scan; -1 forces one
			for {
				if n := curLen(); n != scannedTo {
					tail, m := tailFrom(matchOffset)
					scannedTo = m
					if loc := expects[i].FindIndex(tail); loc != nil {
						matchOffset += loc[1]
						matched = true
						break
					}
				}
				select {
				case <-ctx.Done():
					// One last check: bytes may have landed in the final poll
					// window before the deadline fired.
					tail, _ := tailFrom(matchOffset)
					if loc := expects[i].FindIndex(tail); loc != nil {
						matchOffset += loc[1]
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
					return canceledResult()
				}
				return abort(&ExpectFailure{Pattern: a.Expect, Transcript: string(snapshot())})
			}
			continue
		}
		if a.ExpectScreen != nil {
			waitCtx, cancelWait := sessionWaitContext(ctx, a.ExpectScreen.Timeout)
			matched := false
			var stableSince time.Time
			stableFor := parsePositiveDuration(a.ExpectScreen.StableFor)
			scannedTo := -1 // transcript length at the last render; -1 forces one
			for {
				if n := curLen(); n != scannedTo {
					scannedTo = n
					screen := currentScreen()
					cr := checkRenderedScreen(a.ExpectScreen, screen)
					if cr.OK {
						if stableFor <= 0 {
							matched = true
							break
						}
						if stableSince.IsZero() {
							stableSince = time.Now()
						}
						if time.Since(stableSince) >= stableFor {
							matched = true
							break
						}
					} else {
						stableSince = time.Time{}
					}
				}
				select {
				case <-waitCtx.Done():
					screen := currentScreen()
					cr := checkRenderedScreen(a.ExpectScreen, screen)
					if cr.OK {
						if stableFor <= 0 {
							matched = true
						} else {
							if stableSince.IsZero() {
								stableSince = time.Now()
							}
							if time.Since(stableSince) >= stableFor {
								matched = true
							}
						}
					}
				case <-time.After(pollInterval):
					if !stableSince.IsZero() && stableFor > 0 && time.Since(stableSince) >= stableFor {
						matched = true
					}
					if matched {
						break
					}
					continue
				}
				break
			}
			cancelWait()
			if !matched {
				// A parent-context cancellation is an execution error that must stop
				// the scenario; only a genuine wait timeout becomes a failed check.
				if errors.Is(waitCtx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
					return canceledResult()
				}
				screen := currentScreen()
				cr := checkRenderedScreen(a.ExpectScreen, screen)
				if cr.OK && stableFor > 0 {
					cr = &assert.CheckResult{
						Desc:           fmt.Sprintf("pty expect_screen stable for %s", stableFor),
						Expected:       fmt.Sprintf("rendered screen to keep satisfying the matcher for %s", stableFor),
						Actual:         string(screen),
						Hint:           fmt.Sprintf("the rendered screen matched, but not continuously for %s before the timeout elapsed", stableFor),
						ArtifactKind:   "screen",
						ArtifactActual: screen,
					}
				}
				return abort(&ExpectFailure{Check: cr})
			}
			continue
		}
		if a.Resize != nil {
			// Record the boundary before changing the size, so the replay that
			// builds the rendered screen splits the transcript at the same point
			// the real terminal did (#379).
			term.markResize(a.Resize.Rows, a.Resize.Cols)
			if rerr := proc.resize(a.Resize.Rows, a.Resize.Cols); rerr != nil {
				return failHard(fmt.Errorf("pty %q: session[%d] resize to %dx%d: %w",
					p.Command, i, a.Resize.Rows, a.Resize.Cols, rerr))
			}
			continue
		}
		if a.Send != nil {
			// A paste is only a paste if the program asked for one. Without the
			// mode, the markers arrive as ordinary characters and the failure
			// surfaces somewhere far away — as a REPL executing a pasted block
			// line by line, or as "[200~" typed into a prompt — so refuse here,
			// where the mistake is (#378).
			if a.Send.Paste != nil && !term.modeEnabled(decsetBracketedPaste) {
				return failHard(fmt.Errorf(
					"pty %q: session[%d] sends a paste, but the program has not enabled bracketed paste "+
						"(it never wrote ESC [?2004h, or turned the mode back off). "+
						"Programs that do not distinguish a paste from typing take a plain send instead; "+
						"if this one does enable the mode, wait for it with an expect or expect_screen before pasting",
					p.Command, i))
			}
			// Bytes resolves named keys to their xterm sequences, wraps a paste
			// in its markers, and keeps the historical rule that an empty
			// verbatim send transmits EOF (^D).
			if _, werr := writeTerm(a.Send.Bytes()); werr != nil {
				return failHard(fmt.Errorf("pty: send: %w", werr))
			}
		}
	}

	// Session complete: wait for the child to exit within the budget.
	select {
	case code := <-proc.exit:
		return finish(false, code, nil)
	case <-ctx.Done():
		// A parent cancellation is a hard error; a session-budget timeout is a
		// normal timed-out result (#30).
		if errors.Is(ctx.Err(), context.Canceled) {
			return canceledResult()
		}
		return abort(nil)
	}
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

func checkRenderedScreen(es *spec.PTYExpectScreen, screen []byte) *assert.CheckResult {
	return assert.Check(&spec.Assert{Screen: &es.StreamAssert}, &runner.Result{
		IsPTY:  true,
		Screen: screen,
	}, assert.Env{})
}
