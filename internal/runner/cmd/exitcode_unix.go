//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
)

// exitCodeFor returns the exit code atago reports for a finished process. A
// process terminated by a signal has no exit code of its own; POSIX shells
// report it as 128+signal (130 for SIGINT, 137 for SIGKILL, 143 for SIGTERM),
// and atago mirrors that so a spec can assert a CLI's signal-handling contract.
//
// Go's ExitError.ExitCode() returns -1 for a signaled process, which collides
// with atago's own timeout/cancel sentinel; the caller resolves timeout and
// parent-cancel before reaching here, so a signal at this point is always the
// program's own termination.
func exitCodeFor(err *exec.ExitError) int {
	if ws, ok := err.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return err.ExitCode()
}
