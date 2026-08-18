//go:build windows

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
// A shell that is NOT cmd.exe (an ATAGO_SHELL pointing at Git Bash, say) is
// left alone: it parses its command line with the MSVCRT rules Go already
// writes, so overriding the escaping would corrupt the command instead of
// preserving it.
func ConfigureShell(c *exec.Cmd, command string) {
	if !isCmdExe(c.Path) {
		return
	}
	c.SysProcAttr = &syscall.SysProcAttr{CmdLine: shellCmdLine(c.Path, command)}
}

// ShellCmdLine returns the raw Windows command line that runs command through
// the shell, for the callers that must hand CreateProcess a single string
// rather than an argv — the ConPTY path used by `pty:` steps and
// `atago record --pty`. Keeping it here is what makes a pty session and a run
// step resolve the SAME shell.
func ShellCmdLine(command string) string { return shellCmdLine(shellPath(), command) }

// shellCmdLine composes the command line for an explicit shell binary.
func shellCmdLine(shell, command string) string {
	if isCmdExe(shell) {
		return syscall.EscapeArg(shell) + ` /S /C "` + command + `"`
	}
	return syscall.EscapeArg(shell) + " -c " + syscall.EscapeArg(command)
}

// shellArgs returns the argv that hands command to shell. cmd.exe takes it
// after /c; everything else atago can be pointed at — a POSIX shell shipped by
// Git for Windows or MSYS2, or PowerShell, for which -c is an accepted
// abbreviation of -Command — takes it after -c.
func shellArgs(shell, command string) []string {
	if isCmdExe(shell) {
		return []string{"/c", command}
	}
	return []string{"-c", command}
}

// isCmdExe reports whether path names the Windows command interpreter, which is
// what decides between the two calling conventions above. Windows paths are
// case-insensitive and the extension is optional, so both spellings count.
func isCmdExe(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == "cmd.exe" || base == "cmd"
}

// shellPath returns an absolute path to the shell used for `shell: true`.
//
// It deliberately does NOT trust PATH, for the same reason the POSIX build does
// not: atago prepends the *program under test*'s directory to its own PATH so a
// spec can invoke the binary by bare name, and a CLI that ships something named
// cmd.exe would then become the harness's shell — changing redirect semantics
// and exit codes for every `shell: true` step in the suite. Passing the bare
// name "cmd" to os/exec resolved exactly that way. So a fixed system location
// wins, mirroring the POSIX build's absolute /bin/sh, and ATAGO_SHELL is the
// documented override on both platforms rather than on POSIX alone.
//
// COMSPEC is consulted after %SystemRoot%\System32\cmd.exe because it is the
// weaker guarantee of the two: it is an ordinary inherited variable, whereas
// SystemRoot is the location Windows itself installs the interpreter at.
func shellPath() string {
	if s := os.Getenv("ATAGO_SHELL"); s != "" {
		return s
	}
	if root := os.Getenv("SystemRoot"); root != "" {
		if p := filepath.Join(root, "System32", "cmd.exe"); isRegularFile(p) {
			return p
		}
	}
	if c := os.Getenv("COMSPEC"); filepath.IsAbs(c) && isRegularFile(c) {
		return c
	}
	if p, err := exec.LookPath("cmd"); err == nil {
		return p
	}
	return "cmd"
}

// isRegularFile reports whether path exists and is a file, so a directory named
// cmd.exe cannot be selected as the interpreter.
func isRegularFile(path string) bool {
	//nolint:gosec // G703: the path is %SystemRoot%\System32\cmd.exe, COMSPEC, or the operator's own ATAGO_SHELL — machine configuration, not spec input — and Stat only reads metadata.
	fi, err := os.Stat(path)
	return err == nil && fi.Mode().IsRegular()
}
