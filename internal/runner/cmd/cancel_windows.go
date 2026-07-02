//go:build windows

package cmd

import (
	"os/exec"
	"time"
)

// configureCancellation on Windows relies on CommandContext's default kill of the
// spawned process plus WaitDelay to force-close captured pipes if a child still
// holds them. Windows has no POSIX process groups; job objects would be the
// equivalent but are out of scope for this black-box runner.
func configureCancellation(cmd *exec.Cmd) {
	cmd.WaitDelay = 2 * time.Second
}
