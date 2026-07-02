//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
	"time"
)

// configureCancellation makes a cancelled command kill its whole process group.
// Setpgid puts the child (and everything it spawns) in a new group whose id is
// the child's pid; on cancel we signal -pid, so an orphaned grandchild — the
// classic `sh -c "sleep 30"` case — dies too and releases the captured pipes.
// WaitDelay bounds how long Wait lingers for I/O after that, as a safety net.
func configureCancellation(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		// Negative pid targets the process group. Ignore ESRCH (already gone).
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 2 * time.Second
}
