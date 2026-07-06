//go:build windows

package cmd

import "os/exec"

// exitCodeFor returns the process exit code verbatim. Windows has no POSIX
// signals, so there is no 128+signal convention to apply.
func exitCodeFor(err *exec.ExitError) int {
	return err.ExitCode()
}
