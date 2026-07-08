//go:build !windows

package ptyrun

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/creack/pty"

	"github.com/nao1215/atago/internal/runner"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// Run executes p.Command inside a POSIX pseudo-terminal in workdir and drives
// the expect/send session against it via the shared driveSession core. The
// terminal is a creack/pty master and cleanup kills the whole process group
// (Setsid), so a timed-out or aborted session never leaks the child tree.
func Run(ctx context.Context, p *spec.PTY, workdir string, env []string) (*runner.Result, *ExpectFailure, error) {
	name, args, err := runnercmd.CommandLine(p.Command, p.Shell != nil && *p.Shell)
	if err != nil {
		return nil, nil, err
	}

	// CommandContext binds the child to the parent context (Ctrl-C / suite
	// cancel); Cancel is overridden so that cancellation kills the whole process
	// group (Setsid), not only the direct child. The session-budget timeout is
	// driveSession's job — it kills via proc.kill on abort.
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // the spec author's declared command is the subject under test
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.Dir = workdir
	if p.Cwd != "" {
		cmd.Dir = runnercmd.ResolveDir(workdir, p.Cwd)
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

	// Reap exactly once, from one place: probing a zombie with signal 0 keeps
	// succeeding, so liveness must come from Wait itself. The buffered channel
	// lets the reaper deliver the code even when a kill path drains it later.
	exitCh := make(chan int, 1)
	go func() { exitCh <- waitExitCode(cmd.Wait()) }()

	proc := ptyProcess{
		rw:   master,
		exit: exitCh,
		// Negative pid signals the whole process group created by Setsid.
		kill:      func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) },
		closeTerm: func() { _ = master.Close() },
		dir:       cmd.Dir,
	}
	return driveSession(ctx, p, proc)
}

// waitExitCode maps a cmd.Wait() error to the observed exit code, mirroring the
// cmd runner: nil is a clean 0, an ExitError carries the process's own code, and
// any other failure to reap is -1.
func waitExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
