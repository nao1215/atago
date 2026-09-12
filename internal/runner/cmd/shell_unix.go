//go:build !windows

package cmd

import (
	"os"
	"os/exec"
)

// ConfigureShell is a no-op on POSIX: `sh -c <command>` passes the command as
// one argv element with no re-quoting, so the argv from CommandLine is already
// exact.
func ConfigureShell(_ *exec.Cmd, _ string) {}

// shellArgs returns the argv that hands command to shell. Every POSIX shell
// takes it after -c.
func shellArgs(_, command string) []string { return []string{"-c", command} }

// shellPath returns an absolute path to the POSIX shell used for `shell: true`.
//
// It deliberately does NOT trust PATH: atago sets up PATH for the *program
// under test*, and a CLI may legitimately ship its own `sh` applet (e.g.
// mimixbox). If the harness resolved its shell through that PATH, the program
// under test would hijack atago's shell — changing pipe/redirect semantics and
// exit codes. So we prefer a fixed system location (mirroring ShellSpec's
// absolute `--shell /bin/sh`). The ATAGO_SHELL env var allows an explicit
// override; an absolute /bin/sh is the default; only as a last resort do we
// fall back to a PATH lookup.
func shellPath() string {
	if s := os.Getenv("ATAGO_SHELL"); s != "" {
		return s
	}
	if _, err := os.Stat("/bin/sh"); err == nil {
		return "/bin/sh"
	}
	if p, err := exec.LookPath("sh"); err == nil {
		return p
	}
	return "/bin/sh"
}
