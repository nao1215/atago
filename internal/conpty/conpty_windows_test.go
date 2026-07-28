//go:build windows

package conpty

import (
	"errors"
	"io"
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

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

// TestReadErrorClassification pins the mapping Read applies to a ReadFile
// failure (#345). Collapsing every failure into os.ErrClosed left the pty
// runner unable to tell the end of a session from a lost transcript, so a
// truncated transcript produced a confidently wrong screen assertion. The
// end-of-session codes must read as io.EOF, atago's own close must read as
// os.ErrClosed, and anything else must be neither — a genuine failure the
// caller has to report.
func TestReadErrorClassification(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		err     error
		wantEOF bool
		wantClo bool
	}{
		{"broken pipe is the end of a session", windows.ERROR_BROKEN_PIPE, true, false},
		{"handle eof is the end of a session", windows.ERROR_HANDLE_EOF, true, false},
		{"pipe not connected is the end of a session", windows.ERROR_PIPE_NOT_CONNECTED, true, false},
		{"invalid handle is our own close", windows.ERROR_INVALID_HANDLE, false, true},
		{"aborted is our own close", windows.ERROR_OPERATION_ABORTED, false, true},
		{"access denied is a real failure", windows.ERROR_ACCESS_DENIED, false, false},
		{"not enough memory is a real failure", windows.ERROR_NOT_ENOUGH_MEMORY, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := classifyReadError(c.err)
			if errors.Is(got, io.EOF) != c.wantEOF {
				t.Errorf("errors.Is(%v, io.EOF) = %v, want %v", got, !c.wantEOF, c.wantEOF)
			}
			if errors.Is(got, os.ErrClosed) != c.wantClo {
				t.Errorf("errors.Is(%v, os.ErrClosed) = %v, want %v", got, !c.wantClo, c.wantClo)
			}
			// A real failure must still carry its cause so the report can name it.
			if !c.wantEOF && !c.wantClo && !errors.Is(got, c.err) {
				t.Errorf("a real failure lost its cause: %v does not wrap %v", got, c.err)
			}
		})
	}
}
