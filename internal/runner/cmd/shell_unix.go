//go:build !windows

package cmd

import "os/exec"

// ConfigureShell is a no-op on POSIX: `sh -c <command>` passes the command as
// one argv element with no re-quoting, so the argv from CommandLine is already
// exact.
func ConfigureShell(_ *exec.Cmd, _ string) {}
