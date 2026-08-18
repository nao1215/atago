//go:build windows

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIsCmdExe pins which names select cmd.exe's calling convention. Windows
// paths are case-insensitive and the extension is optional, so both spellings
// count in any case; anything else is treated as a shell that takes -c.
func TestIsCmdExe(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		path string
		want bool
	}{
		{`C:\Windows\System32\cmd.exe`, true},
		{`C:\Windows\System32\CMD.EXE`, true},
		{"cmd", true},
		{"cmd.exe", true},
		{`C:\Program Files\Git\bin\bash.exe`, false},
		{"bash", false},
		{"pwsh.exe", false},
		{`C:\tools\cmd.bat`, false}, // PATHEXT would resolve a bare `cmd` to this
	} {
		if got := isCmdExe(tt.path); got != tt.want {
			t.Errorf("isCmdExe(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// TestShellArgs_PerInterpreter covers the two calling conventions: cmd.exe
// takes the command after /c, everything else after -c.
func TestShellArgs_PerInterpreter(t *testing.T) {
	t.Parallel()
	if got := shellArgs(`C:\Windows\System32\cmd.exe`, "echo hi"); len(got) != 2 || got[0] != "/c" || got[1] != "echo hi" {
		t.Errorf("shellArgs(cmd.exe) = %v, want [/c echo hi]", got)
	}
	if got := shellArgs(`C:\Program Files\Git\bin\bash.exe`, "echo hi"); len(got) != 2 || got[0] != "-c" || got[1] != "echo hi" {
		t.Errorf("shellArgs(bash.exe) = %v, want [-c echo hi]", got)
	}
}

// TestShellCmdLine_QuotesTheInterpreter is the ConPTY half. The command line
// CreateProcess receives has to name the interpreter with its own quoting —
// %SystemRoot% under "Program Files" is not hypothetical on a customized host —
// while cmd.exe's `/S /C "<command>"` contract still passes the command through
// verbatim, so an embedded double quote survives.
func TestShellCmdLine_QuotesTheInterpreter(t *testing.T) {
	t.Parallel()
	got := shellCmdLine(`C:\Program Files\cmd.exe`, `echo {"id":7}`)
	want := `"C:\Program Files\cmd.exe" /S /C "echo {"id":7}"`
	if got != want {
		t.Errorf("shellCmdLine(cmd.exe) = %q, want %q", got, want)
	}
	// A non-cmd shell parses MSVCRT quoting, so the command is escaped as one
	// argument rather than handed over raw.
	got = shellCmdLine(`C:\Git\bash.exe`, `echo "hi there"`)
	if !strings.HasPrefix(got, `C:\Git\bash.exe -c `) {
		t.Errorf("shellCmdLine(bash.exe) = %q, want it to start with the interpreter and -c", got)
	}
	if strings.Contains(got, `/S /C`) {
		t.Errorf("shellCmdLine(bash.exe) = %q, want no cmd.exe switches", got)
	}
}

// TestShellPath_ResolvesAbsolutely is the Windows half of the anti-hijack
// contract: with no override the interpreter comes from %SystemRoot%, never
// from a bare name os/exec would look up on the PATH that atago has already
// prepended the program under test to.
func TestShellPath_ResolvesAbsolutely(t *testing.T) {
	t.Setenv("ATAGO_SHELL", "")
	got := shellPath()
	if !filepath.IsAbs(got) {
		t.Fatalf("shellPath() = %q, want an absolute path", got)
	}
	if strings.ToLower(filepath.Base(got)) != "cmd.exe" {
		t.Errorf("shellPath() = %q, want cmd.exe", got)
	}
	if _, err := os.Stat(got); err != nil {
		t.Errorf("shellPath() = %q, which does not exist: %v", got, err)
	}
}

// TestShellPath_SystemRootWinsOverComspec pins the precedence: COMSPEC is an
// ordinary inherited variable, %SystemRoot%\System32\cmd.exe is where Windows
// itself installs the interpreter, so the fixed location is preferred.
func TestShellPath_SystemRootWinsOverComspec(t *testing.T) {
	t.Setenv("ATAGO_SHELL", "")
	t.Setenv("COMSPEC", `C:\definitely\not\here\cmd.exe`)
	got := shellPath()
	if strings.EqualFold(got, `C:\definitely\not\here\cmd.exe`) {
		t.Fatal("shellPath() took a COMSPEC that does not exist")
	}
	root := os.Getenv("SystemRoot")
	if root == "" {
		t.Skip("no SystemRoot on this host to compare against")
	}
	if want := filepath.Join(root, "System32", "cmd.exe"); !strings.EqualFold(got, want) {
		t.Errorf("shellPath() = %q, want %q", got, want)
	}
}
