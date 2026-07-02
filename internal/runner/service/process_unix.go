//go:build !windows

package service

import (
	"context"
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

func (p *processCmd) terminate() { _ = signalGroup(p.cmd, syscall.SIGTERM) }
func (p *processCmd) kill()      { _ = signalGroup(p.cmd, syscall.SIGKILL) }

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
