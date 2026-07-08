//go:build windows

package ptyrun

import (
	"context"
	"errors"
	"fmt"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"os/exec"
	"strconv"

	"github.com/nao1215/atago/internal/conpty"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// Run executes p.Command inside a Windows pseudo-console (ConPTY) and drives the
// expect/send session against it via the shared driveSession core, so pty specs
// that were POSIX-only now run on Windows too (follow-up to #78). The loader
// still accepts pty steps on every platform; here they execute instead of
// returning the old "unsupported" error. ConPTY needs Windows 10 (1809) or
// later — an older host gets one clear execution error (exit 4).
func Run(ctx context.Context, p *spec.PTY, workdir string, env []string) (*runner.Result, *ExpectFailure, error) {
	if !conpty.IsAvailable() {
		return nil, nil, errors.New("pty steps need ConPTY, which requires Windows 10 version 1809 or later (gate the scenario with `skip: {os: windows}` for older hosts)")
	}

	cmdLine, err := conpty.CommandLine(p.Command, p.Shell != nil && *p.Shell)
	if err != nil {
		return nil, nil, err
	}

	dir := workdir
	if p.Cwd != "" {
		dir = runnercmd.ResolveDir(workdir, p.Cwd)
	}

	rows, cols := defaultRows, defaultCols
	if p.Rows > 0 {
		rows = p.Rows
	}
	if p.Cols > 0 {
		cols = p.Cols
	}

	// env is whatever the engine resolved (clear_env/pass_env already applied):
	// nil inherits the parent's environment; a non-nil slice starts the child
	// from exactly that set.
	cpty, err := conpty.Start(cmdLine, dir, env, rows, cols)
	if err != nil {
		return nil, nil, fmt.Errorf("pty: start %q: %w", p.Command, err)
	}

	pid := cpty.Pid()

	// Reap in one place, mirroring the POSIX runner. cpty.Wait blocks on the
	// process handle and returns its exit code; a parent-context cancel that
	// unblocks it early, or an unreadable code, maps to -1. On a normal run the
	// parent context stays alive, so Wait blocks until the child exits; on
	// abort/timeout driveSession kills the tree, which lets Wait return. The
	// buffered channel lets a kill path drain the code later.
	exitCh := make(chan int, 1)
	go func() { exitCh <- cpty.Wait(ctx) }()

	proc := ptyProcess{
		rw:        cpty,
		exit:      exitCh,
		kill:      func() { killTree(pid) },
		closeTerm: func() { _ = cpty.Close() },
		dir:       dir,
	}
	return driveSession(ctx, p, proc)
}

// killTree force-terminates the child and every descendant. Windows has no
// process groups, so taskkill /T walks the process tree — the closest analog to
// the POSIX runner's kill of the whole Setsid group, so a timed-out or aborted
// pty session never leaks a running child.
func killTree(pid int) {
	// A fire-and-forget teardown; Background is the honest context here.
	_ = exec.CommandContext(context.Background(), "taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run() //nolint:gosec // fixed argv, pid from our own child
}
