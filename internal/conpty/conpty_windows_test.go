//go:build windows

package conpty

import "testing"

// TestCommandLine covers how a command becomes the single command line ConPTY
// hands to CreateProcess: a shell command reuses cmd.exe's `/S /C "<command>"`
// contract verbatim, and a shell-free command is tokenized with the cmd runner's
// splitter and re-escaped so the C runtime re-parses it to the same argv (plain
// words stay bare, a path with spaces gets quoted).
func TestCommandLine(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		command string
		shell   bool
		want    string
	}{
		{"shell wraps in cmd /S /C", `echo hi & echo bye`, true, `cmd /S /C "echo hi & echo bye"`},
		{"plain words stay bare", `tool --flag value`, false, `tool --flag value`},
		{"quoted path with spaces re-quotes", `"C:\Program Files\t.exe" run`, false, `"C:\Program Files\t.exe" run`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := CommandLine(c.command, c.shell)
			if err != nil {
				t.Fatalf("CommandLine(%q, %v): unexpected error: %v", c.command, c.shell, err)
			}
			if got != c.want {
				t.Errorf("CommandLine(%q, %v) = %q, want %q", c.command, c.shell, got, c.want)
			}
		})
	}
}
