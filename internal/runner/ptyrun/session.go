package ptyrun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

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
	// dir is the resolved workdir the child ran in, surfaced as Result.Workdir.
	dir string
}

// driveSession runs the platform-neutral half of a pty step: it drains the
// transcript, walks the expect/send session in order, and shapes the Result.
// Every terminal- and process-specific detail is behind proc, so POSIX and
// Windows share this exact control flow. A never-matching expect is returned as
// an ExpectFailure (reported like a failed assertion); only "could not
// start/drive the terminal" conditions are hard errors.
func driveSession(ctx context.Context, p *spec.PTY, proc ptyProcess) (*runner.Result, *ExpectFailure, error) {
	expects, err := compileSession(p.Session)
	if err != nil {
		return nil, nil, fmt.Errorf("pty: invalid expect regexp: %w", err)
	}

	budget := sessionTimeout(p)
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	start := time.Now()

	// Transcript accumulator: one goroutine drains the master so the child never
	// blocks on a full terminal buffer. Reads end when the child exits (EOF/EIO)
	// or the master is closed.
	var mu sync.Mutex
	var transcript []byte
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		buf := make([]byte, 4096)
		for {
			n, rerr := proc.rw.Read(buf)
			if n > 0 {
				mu.Lock()
				transcript = append(transcript, buf[:n]...)
				mu.Unlock()
			}
			if rerr != nil {
				return
			}
		}
	}()
	snapshot := func() []byte {
		mu.Lock()
		defer mu.Unlock()
		return append([]byte(nil), transcript...)
	}
	// tailFrom copies only transcript[from:] under the lock and reports the
	// transcript's current length; curLen reports the length alone. Together
	// they let a pending expect skip the poll entirely when nothing new
	// arrived and copy only the bytes it can still match — a TUI redrawing at
	// full speed otherwise costs an O(transcript) allocation every 10ms per
	// expect, pure garbage churn on bytes before matchOffset that no later
	// expect may ever see again.
	tailFrom := func(from int) ([]byte, int) {
		mu.Lock()
		defer mu.Unlock()
		if from > len(transcript) {
			from = len(transcript)
		}
		return append([]byte(nil), transcript[from:]...), len(transcript)
	}
	curLen := func() int {
		mu.Lock()
		defer mu.Unlock()
		return len(transcript)
	}

	finish := func(timedOut bool, code int, ef *ExpectFailure) (*runner.Result, *ExpectFailure, error) {
		// Drain before closing: a fast-exiting child's final output may still sit
		// in the pty buffer, and closing the master discards it. Once the last
		// handle is gone the reader hits EOF and readDone closes on its own;
		// drainGrace bounds the wait in case a descendant kept the terminal open.
		select {
		case <-readDone:
		case <-time.After(drainGrace):
		}
		proc.closeTerm()
		<-readDone
		tr := snapshot()
		res := &runner.Result{
			Command:  p.Command,
			Stdout:   tr,
			Duration: time.Since(start),
			Workdir:  proc.dir,
			TimedOut: timedOut,
			IsPTY:    true,
			// The rendered screen (#27) is derived from the same bytes, so screen
			// asserts and transcript asserts never disagree about what happened.
			Screen: []byte(RenderScreen(tr, p)),
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
		proc.closeTerm()
		<-readDone
		return nil, nil, err
	}

	// canceledResult surfaces a parent-context cancellation (Ctrl-C / suite
	// cancel) as a hard execution error, so the engine stops the scenario instead
	// of asserting against a killed terminal — mirroring the cmd runner's
	// cancel/timeout split (#30).
	canceledResult := func() (*runner.Result, *ExpectFailure, error) {
		return failHard(fmt.Errorf("pty %q canceled: %w", p.Command, ctx.Err()))
	}

	// Drive the session in order. expect polls the transcript; send writes to the
	// terminal; an empty send transmits EOF (^D).
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
		if a.Send != nil {
			// Bytes resolves named keys to their xterm sequences and keeps the
			// historical rule that an empty verbatim send transmits EOF (^D).
			if _, werr := proc.rw.Write(a.Send.Bytes()); werr != nil {
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
