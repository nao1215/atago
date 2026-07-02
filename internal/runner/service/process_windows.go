//go:build windows

package service

import (
	"context"
	"os/exec"
	"time"
)

// processCmd wraps an *exec.Cmd. Windows has no POSIX process groups or signals,
// so termination is a hard kill of the spawned process.
type processCmd struct {
	cmd *exec.Cmd
}

func newProcessCmd(ctx context.Context, name string, args []string) *processCmd {
	c := exec.CommandContext(ctx, name, args...) //nolint:gosec // executing user-declared service commands is the purpose of atago
	c.WaitDelay = 5 * time.Second
	return &processCmd{cmd: c}
}

func (p *processCmd) terminate() {
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}

func (p *processCmd) kill() {
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}
