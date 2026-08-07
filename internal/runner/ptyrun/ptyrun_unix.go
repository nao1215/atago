//go:build !windows

package ptyrun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
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
	// A fresh session and controlling terminal: the child owns the pty, and
	// cleanup can still kill the whole tree by the negative process-group pid.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}

	rows, cols := uint16(defaultRows), uint16(defaultCols)
	if p.Rows > 0 && p.Rows < 1<<16 {
		rows = uint16(p.Rows)
	}
	if p.Cols > 0 && p.Cols < 1<<16 {
		cols = uint16(p.Cols)
	}
	master, tty, err := OpenTerminal(rows, cols)
	if err != nil {
		return nil, nil, fmt.Errorf("pty: start %q: %w", p.Command, err)
	}
	defer func() { _ = tty.Close() }()
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	// Post the first Read before the child starts: macOS can otherwise let a
	// no-shell fast-exit command print and disappear before the drain exists.
	term := startTranscriptDrain(master, p)
	// Yield once so the drain goroutine can enter its blocking Read before this
	// goroutine launches a child that may write and exit immediately.
	runtime.Gosched()
	if err := cmd.Start(); err != nil {
		// Drop atago's own slave handle first: while the parent still holds the
		// terminal open the master read has no reason to end, and waiting for the
		// drain would hang instead of surfacing the start failure.
		_ = tty.Close()
		term.waitDrain(func() { _ = master.Close() }, 0)
		return nil, nil, fmt.Errorf("pty: start %q: %w", p.Command, err)
	}
	_ = tty.Close()

	// Reap exactly once, from one place: probing a zombie with signal 0 keeps
	// succeeding, so liveness must come from Wait itself. The buffered channel
	// lets the reaper deliver the code even when a kill path drains it later.
	exitCh := make(chan int, 1)
	go func() { exitCh <- waitExitCode(cmd.Wait()) }()

	proc := ptyProcess{
		rw:    master,
		trans: term,
		exit:  exitCh,
		// Negative pid signals the whole process group created by Setsid.
		kill:      func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) },
		closeTerm: func() { _ = master.Close() },
		dir:       cmd.Dir,
	}
	return driveSession(ctx, p, proc)
}

// OpenTerminal opens a pseudo-terminal pair sized rows×cols and hands back a
// master whose reads the runtime poller owns, so that closing it interrupts a
// read already in flight. Every POSIX pty atago opens — the pty runner here and
// the interactive recorder — goes through it, because getting this wrong hangs
// the process rather than failing it.
//
// That last part is the whole reason this helper exists. creack/pty runs its
// ioctls through (*os.File).Fd(), which is documented to return a blocking
// descriptor: os.File hands the descriptor over to the caller and takes it back
// out of the poller. Reads then park inside read(2), where Close cannot reach
// them — Go can only mark the file closed and defer the real close(2) until the
// read returns on its own. A master read ends on its own when the last slave
// handle goes away, so any surviving handle (a descendant that escaped the
// process-group kill) left the drain blocked forever. Restoring O_NONBLOCK puts
// reads back under the poller, where Close unblocks them with fs.ErrClosed —
// which isSessionEnd already reads as a normal end of session.
func OpenTerminal(rows, cols uint16) (master, tty *os.File, err error) {
	master, tty, err = pty.Open()
	if err != nil {
		return nil, nil, err
	}
	closeBoth := func() {
		_ = tty.Close()
		_ = master.Close()
	}
	if err := pty.Setsize(master, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
		closeBoth()
		return nil, nil, err
	}
	sc, err := master.SyscallConn()
	if err != nil {
		closeBoth()
		return nil, nil, err
	}
	var setErr error
	if ctlErr := sc.Control(func(fd uintptr) { setErr = syscall.SetNonblock(int(fd), true) }); ctlErr != nil {
		closeBoth()
		return nil, nil, ctlErr
	}
	if setErr != nil {
		closeBoth()
		return nil, nil, setErr
	}
	return master, tty, nil
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
