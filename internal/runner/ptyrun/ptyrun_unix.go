//go:build !windows

package ptyrun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"

	"github.com/nao1215/atago/internal/runner"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// pollInterval is how often an expect re-checks the accumulated transcript.
const pollInterval = 10 * time.Millisecond

// drainGrace bounds how long finish waits for the reader to hit EIO before
// closing the master: an orphaned grandchild that inherited the slave can hold
// the terminal open indefinitely, and its output is not worth hanging for.
const drainGrace = 2 * time.Second

// Run executes p.Command inside a pseudo-terminal in workdir, drives the
// expect/send session, waits for the process to exit within the session
// budget, and returns the transcript as the result's stdout. A never-matching
// expect is returned as an ExpectFailure (reported like a failed assertion);
// only "could not start/drive the terminal" conditions are hard errors.
func Run(ctx context.Context, p *spec.PTY, workdir string, env []string) (*runner.Result, *ExpectFailure, error) {
	expects, err := compileSession(p.Session)
	if err != nil {
		return nil, nil, fmt.Errorf("pty: invalid expect regexp: %w", err)
	}

	name, args, err := runnercmd.CommandLine(p.Command, p.Shell != nil && *p.Shell)
	if err != nil {
		return nil, nil, err
	}

	budget := sessionTimeout(p)
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	// CommandContext gives noctx its context; Cancel is overridden below so a
	// timeout kills the whole process group (Setsid), not only the child.
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // the spec author's declared command is the subject under test
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.Dir = workdir
	if p.Cwd != "" {
		cmd.Dir = resolveCwd(workdir, p.Cwd)
	}
	cmd.Env = env
	// A fresh process group so cleanup kills the whole tree, mirroring the cmd
	// runner's teardown discipline.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	rows, cols := uint16(defaultRows), uint16(defaultCols)
	if p.Rows > 0 && p.Rows < 1<<16 {
		rows = uint16(p.Rows)
	}
	if p.Cols > 0 && p.Cols < 1<<16 {
		cols = uint16(p.Cols)
	}
	master, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: rows, Cols: cols})
	if err != nil {
		return nil, nil, fmt.Errorf("pty: start %q: %w", p.Command, err)
	}

	start := time.Now()

	// Reap exactly once, from one place: probing a zombie with signal 0 keeps
	// succeeding, so liveness must come from Wait itself.
	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	// Transcript accumulator: one goroutine drains the master so the child
	// never blocks on a full terminal buffer. Reads end when the child exits
	// (EIO) or the master is closed.
	var mu sync.Mutex
	var transcript []byte
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		buf := make([]byte, 4096)
		for {
			n, rerr := master.Read(buf)
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

	kill := func() {
		// Negative pid signals the whole process group created by Setsid.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	finish := func(timedOut bool, waitErr error, ef *ExpectFailure) (*runner.Result, *ExpectFailure, error) {
		// Drain before closing: a fast-exiting child's final output may still
		// sit in the pty buffer, and closing the master discards it. Once the
		// last slave fd is gone the reader hits EIO and readDone closes on its
		// own; drainGrace bounds the wait in case a descendant kept the slave.
		select {
		case <-readDone:
		case <-time.After(drainGrace):
		}
		_ = master.Close()
		<-readDone
		transcript := snapshot()
		res := &runner.Result{
			Command:  p.Command,
			Stdout:   transcript,
			Duration: time.Since(start),
			Workdir:  cmd.Dir,
			TimedOut: timedOut,
			IsPTY:    true,
			// The rendered screen (#27) is derived from the same bytes, so
			// screen asserts and transcript asserts never disagree about what
			// happened.
			Screen: []byte(RenderScreen(transcript, p)),
		}
		switch {
		case timedOut:
			res.ExitCode = -1
		case waitErr == nil:
			res.ExitCode = 0
		default:
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				res.ExitCode = exitErr.ExitCode()
			} else {
				res.ExitCode = -1
			}
		}
		return res, ef, nil
	}

	// abort kills the tree and reaps it, then finishes as timed out.
	abort := func(ef *ExpectFailure) (*runner.Result, *ExpectFailure, error) {
		kill()
		waitErr := <-waitCh
		_ = waitErr
		return finish(true, nil, ef)
	}

	// failHard cleans up (kill, reap, close, drain) before surfacing a hard
	// error, so a failed terminal write never leaks the child or goroutines.
	failHard := func(err error) (*runner.Result, *ExpectFailure, error) {
		kill()
		<-waitCh
		_ = master.Close()
		<-readDone
		return nil, nil, err
	}

	// Drive the session in order. expect polls the transcript; send writes to
	// the terminal; an empty send transmits EOF (^D).
	//
	// matchOffset is the byte index just past the previously matched expect: each
	// expect scans only transcript[matchOffset:], so a pattern that recurs (any
	// shell prompt) waits for its NEXT occurrence instead of matching the stale
	// earlier one. Without it, a second `expect "PROMPT> "` matches the first
	// prompt already in the buffer and the session races ahead — a false pass for
	// exactly the interactive flows this feature exists for.
	matchOffset := 0
	for i, a := range p.Session {
		if expects[i] != nil {
			matched := false
			for {
				if loc := expects[i].FindIndex(snapshot()[matchOffset:]); loc != nil {
					matchOffset += loc[1]
					matched = true
					break
				}
				select {
				case <-ctx.Done():
					// One last check: bytes may have landed in the final poll
					// window before the deadline fired.
					if loc := expects[i].FindIndex(snapshot()[matchOffset:]); loc != nil {
						matchOffset += loc[1]
						matched = true
					}
				case <-time.After(pollInterval):
					continue
				}
				break
			}
			if !matched {
				// A parent-context cancellation (Ctrl-C / suite cancel) is an
				// execution error that must stop the scenario, not an expect
				// failure the engine asserts against — mirroring the cmd runner's
				// cancel/timeout split (#30). Only a genuine session-budget timeout
				// (DeadlineExceeded) becomes an ExpectFailure.
				if errors.Is(ctx.Err(), context.Canceled) {
					return failHard(fmt.Errorf("pty %q cancelled: %w", p.Command, ctx.Err()))
				}
				return abort(&ExpectFailure{Pattern: a.Expect, Transcript: string(snapshot())})
			}
			continue
		}
		if a.Send != nil {
			// Bytes resolves named keys to their xterm sequences and keeps the
			// historical rule that an empty verbatim send transmits EOF (^D,
			// delivered as VEOF on an empty input line) (#26).
			if _, werr := master.Write(a.Send.Bytes()); werr != nil {
				return failHard(fmt.Errorf("pty: send: %w", werr))
			}
		}
	}

	// Session complete: wait for the child to exit within the budget.
	select {
	case waitErr := <-waitCh:
		return finish(false, waitErr, nil)
	case <-ctx.Done():
		// A parent cancellation is a hard error; a session-budget timeout is a
		// normal timed-out result (#30).
		if errors.Is(ctx.Err(), context.Canceled) {
			return failHard(fmt.Errorf("pty %q cancelled: %w", p.Command, ctx.Err()))
		}
		return abort(nil)
	}
}

// resolveCwd resolves a step cwd against the scenario workdir like the cmd
// runner does: relative paths nest inside the workdir.
func resolveCwd(workdir, cwd string) string {
	if cwd == "" {
		return workdir
	}
	if len(cwd) > 0 && cwd[0] == '/' {
		return cwd
	}
	return workdir + string(os.PathSeparator) + cwd
}
