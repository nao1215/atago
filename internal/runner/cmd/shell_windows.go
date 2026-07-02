//go:build windows

package cmd

import (
	"os/exec"
	"syscall"
)

// ConfigureShell hands cmd.exe the raw command line for a `shell: true`
// command. Go's default argv-to-command-line escaping follows the MSVCRT
// quoting rules, but cmd.exe never unescapes them: an embedded double quote
// would reach the command as a literal \" and corrupt it (a spec printing JSON
// was the reproducer). `/S /C "<command>"` is cmd's documented contract: strip
// exactly the first and last quote and run everything between verbatim — the
// command behaves as if typed at the prompt.
func ConfigureShell(c *exec.Cmd, command string) {
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: `cmd /S /C "` + command + `"`}
}
