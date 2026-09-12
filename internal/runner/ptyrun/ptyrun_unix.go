//go:build !windows

package ptyrun

import (
	"context"
	"math"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/runner"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// Run executes p.Command inside a POSIX pseudo-terminal in workdir and drives
// the expect/send session against it via the shared driveSession core. The
// terminal comes from OpenTerminal and cleanup kills the whole process group
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
		return nil, nil, diag.CommandNotStarted.Errorf("pty: start %q: %w", p.Command, err)
	}
	// releaseTTY drops atago's own slave handle. It is deferred as a backstop and
	// called explicitly at the points below, so it has to tolerate both.
	var ttyOnce sync.Once
	releaseTTY := func() { ttyOnce.Do(func() { _ = tty.Close() }) }
	defer releaseTTY()
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	// Start draining before the child starts, so the reader is in place for a
	// command that prints and exits at once. startTranscriptDrain does not return
	// until its reader goroutine has been scheduled and has reached its read call
	// — one statement short of a promise that the read is posted, but as close as
	// user space gets.
	term := startTranscriptDrain(master, p)
	if err := cmd.Start(); err != nil {
		// Nothing was started, so nothing can be waiting to be read: drop the
		// slave handle now, or the master read has no reason to end and waiting
		// for the drain would hang instead of surfacing the start failure.
		releaseTTY()
		term.waitDrain(func() { _ = master.Close() }, 0)
		return nil, nil, diag.CommandNotStarted.Errorf("pty: start %q: %w", p.Command, err)
	}
	// atago's slave handle deliberately stays open until the child has been
	// reaped (driveSession's finish calls releaseTTY). Handing the terminal to
	// the child and closing this handle immediately would leave the child's own
	// descriptors as the last ones, and on macOS the pty discards whatever it
	// still holds the moment that last handle closes — a read already parked in
	// the kernel returns EOF with no bytes, so no amount of draining earlier can
	// win that race. Keeping one handle here means the last close happens after
	// atago has read, on a schedule atago controls.

	// Reap exactly once, from one place: probing a zombie with signal 0 keeps
	// succeeding, so liveness must come from Wait itself. The buffered channel
	// lets the reaper deliver the code even when a kill path drains it later.
	exitCh := make(chan int, 1)
	// The exit code goes through the cmd runner's shared mapping, so a signaled
	// child reports 128+signal here exactly as it does through a run: step. The
	// abort paths (session budget, parent cancel) resolve to -1 before this value
	// is ever consumed, which is what keeps a timeout kill from looking like 137.
	go func() { exitCh <- runnercmd.ExitCode(cmd.Wait()) }()

	proc := ptyProcess{
		rw:    master,
		trans: term,
		exit:  exitCh,
		// Negative pid signals the whole process group created by Setsid.
		kill:       func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) },
		closeTerm:  func() { _ = master.Close() },
		releaseTTY: releaseTTY,
		// setTerminalSize's TIOCSWINSZ on the master is the whole mechanism: the
		// kernel records the new size AND sends SIGWINCH to the terminal's
		// foreground process group, so the child learns about it exactly as it
		// would from a real window change (#379).
		resize: func(rows, cols int) error {
			if rows < 1 || cols < 1 || rows > math.MaxUint16 || cols > math.MaxUint16 {
				return diag.PTYFailed.Errorf("size %dx%d is out of range for a terminal", rows, cols)
			}
			return setTerminalSize(master, uint16(rows), uint16(cols))
		},
		dir: cmd.Dir,
		env: env,
	}
	return driveSession(ctx, p, proc)
}

// OpenTerminal opens a pseudo-terminal pair sized rows×cols and hands back a
// master whose reads the runtime poller owns, so that closing it interrupts a
// read already in flight — where the platform lets the poller own it at all.
// Every POSIX pty atago opens — the pty runner here and the interactive
// recorder — goes through it, because getting this wrong hangs the process
// rather than failing it.
//
// The pair itself comes from openTerminalPair, which on Linux is atago's own
// four syscalls rather than creack/pty's: creack/pty can silently report the
// wrong slave there, and a master paired with a stranger's terminal loses the
// whole transcript (#385, and the reasoning in terminal_linux.go).
//
// The poller part is the other reason this helper exists. An ioctl reached
// through (*os.File).Fd() — which is how creack/pty sizes a terminal — costs
// the file its poller: Fd() is documented to return a blocking descriptor, so
// os.File clears O_NONBLOCK and hands the descriptor over. Reads then park
// inside read(2), where Close cannot reach them — Go can only mark the file
// closed and defer the real close(2) until the read returns on its own. A master
// read ends on its own when the last slave handle goes away, so any surviving
// handle (a descendant that escaped the process-group kill) left the drain
// blocked forever. Every ioctl atago performs therefore goes through ControlFD
// instead, and adoptMasterReads restores O_NONBLOCK on a master that arrived
// blocking, so that Close unblocks a pending read with fs.ErrClosed — which
// isSessionEnd already reads as a normal end of session.
func OpenTerminal(rows, cols uint16) (master, tty *os.File, err error) {
	master, tty, err = openTerminalPair()
	if err != nil {
		return nil, nil, err
	}
	closeBoth := func() {
		_ = tty.Close()
		_ = master.Close()
	}
	if err := setTerminalSize(master, rows, cols); err != nil {
		closeBoth()
		return nil, nil, err
	}
	if err := adoptMasterReads(master); err != nil {
		closeBoth()
		return nil, nil, err
	}
	return master, tty, nil
}

// adoptMasterReads puts the master's reads back under the runtime poller when
// the poller is in a position to take them, and leaves them blocking when it is
// not.
//
// That condition is not cosmetic. A non-blocking descriptor the poller does not
// own surfaces EAGAIN straight to the caller, so every read fails instead of
// waiting and the drain dies on the first one. Which regime a pty master lands
// in is decided by how it was opened: os.OpenFile registers the descriptor with
// the poller (atago's own Linux path), while creack/pty on macOS wraps a raw
// descriptor with os.NewFile, which does not. SetReadDeadline
// reports that ownership directly — a descriptor the poller does not hold
// answers os.ErrNoDeadline — so ask it instead of guessing from the platform.
//
// Where the answer is no, closing the master cannot interrupt a read already in
// flight, and waitDrain's bounded join is what keeps that from wedging the run.
func adoptMasterReads(master *os.File) error {
	// The zero time means "no deadline": this sets nothing and clears nothing,
	// it only asks the question.
	if err := master.SetReadDeadline(time.Time{}); err != nil {
		return nil //nolint:nilerr // a master the poller does not own is a platform fact, not a failure to open a terminal
	}
	sc, err := master.SyscallConn()
	if err != nil {
		return err
	}
	var setErr error
	if ctlErr := sc.Control(func(fd uintptr) { setErr = syscall.SetNonblock(int(fd), true) }); ctlErr != nil {
		return ctlErr
	}
	return setErr
}
