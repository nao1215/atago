//go:build windows

package cmd

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// ConfigureShell hands cmd.exe the raw command line for a `shell: true`
// command. Go's default argv-to-command-line escaping follows the MSVCRT
// quoting rules, but cmd.exe never unescapes them: an embedded double quote
// would reach the command as a literal \" and corrupt it (a spec printing JSON
// was the reproducer). `/S /C "<command>"` is cmd's documented contract: strip
// exactly the first and last quote and run everything between verbatim — the
// command behaves as if typed at the prompt.
//
// When ATAGO_SHELL selects a POSIX shell (Git Bash's sh) the command runs via
// `<sh> -c <command>` instead, and that shell unquotes MSVCRT-escaped argv like
// any mingw program — so Go's default escaping is already correct and this hack
// must be skipped, exactly as on POSIX.
func ConfigureShell(c *exec.Cmd, command string) {
	if !windowsUsesCmdExe(runtime.GOOS, os.Getenv("ATAGO_SHELL")) {
		return
	}
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: `cmd /S /C "` + command + `"`}
}
