//go:build !windows

package service

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// processCmd wraps an *exec.Cmd whose children are placed in their own process
// group so the whole tree can be signaled at once — a service started with
// `shell: true` may fork further children that a bare process kill would orphan.
type processCmd struct {
	cmd *exec.Cmd
}

func newProcessCmd(ctx context.Context, name string, args []string) *processCmd {
	c := exec.CommandContext(ctx, name, args...) //nolint:gosec // executing user-declared service commands is the purpose of atago
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// On scenario-context cancellation, signal the whole group (not just the
	// leader) and force-kill if it lingers past WaitDelay.
	c.Cancel = func() error { return signalGroup(c, syscall.SIGTERM) }
	c.WaitDelay = 5 * time.Second
	return &processCmd{cmd: c}
}

// started is a no-op on POSIX: the process group is established at spawn time by
// Setpgid, so there is nothing to wire up after Start (the Windows build assigns
// a job object here instead).
func (p *processCmd) started() error { return nil }

func (p *processCmd) terminate() { _ = signalGroup(p.cmd, syscall.SIGTERM) }
func (p *processCmd) kill()      { _ = signalGroup(p.cmd, syscall.SIGKILL) }

// namedSignals maps the signal names a `signal:` step accepts (#23) to their
// POSIX numbers. The loader validates against the same set, so an unknown
// name here means the spec bypassed validation.
var namedSignals = map[string]syscall.Signal{
	"TERM": syscall.SIGTERM,
	"INT":  syscall.SIGINT,
	"HUP":  syscall.SIGHUP,
	"USR1": syscall.SIGUSR1,
	"USR2": syscall.SIGUSR2,
	"KILL": syscall.SIGKILL,
}

// signalByName delivers the named signal to the whole process group (#23),
// consistent with terminate/kill.
func (p *processCmd) signalByName(name string) error {
	sig, ok := namedSignals[name]
	if !ok {
		return fmt.Errorf("unknown signal %q (accepted: TERM, INT, HUP, USR1, USR2, KILL)", name)
	}
	return signalGroup(p.cmd, sig)
}

// signalGroup sends sig to the command's process group, falling back to the
// leader process if the group id cannot be resolved.
func signalGroup(c *exec.Cmd, sig syscall.Signal) error {
	if c.Process == nil {
		return nil
	}
	if pgid, err := syscall.Getpgid(c.Process.Pid); err == nil {
		return syscall.Kill(-pgid, sig)
	}
	return c.Process.Signal(sig)
}
