//go:build windows

package conpty

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

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

// TestClose_ReleasesAPendingRead is the #406 regression, and the one thing the
// package could not previously do: end a read that is already parked in the
// kernel.
//
// Read issues a plain synchronous ReadFile, and CloseHandle does not abort one
// that is already in flight — the reader stays blocked until the write end goes
// away on its own, which a surviving conhost or an escaped descendant can put
// off indefinitely. In the recorder that showed up twice: every timed-out
// capture had to wait out its whole drain grace, and the reader goroutine leaked
// for the life of the process holding the pipe handle open. Close now cancels
// pending I/O on the handle first.
//
// The child is a command that produces nothing and outlives the test, so the
// read under way is genuinely parked with no data coming: exactly the state the
// bug needed.
func TestClose_ReleasesAPendingRead(t *testing.T) {
	// Repeated because the bug it guards was intermittent in CI: whether a read
	// is parked at close time depended on timing, so one green iteration proves
	// less than a handful. Each iteration is a fresh console and costs well
	// under a second.
	for i := range 5 {
		cp, err := Start(`cmd /S /C "ping -n 30 127.0.0.1 > NUL"`, "", nil, 24, 80)
		if err != nil {
			t.Fatalf("iteration %d: Start: %v", i, err)
		}

		// Drain first: the console emits its initial screen setup, and Close has
		// to interrupt a read waiting for MORE, not one about to return bytes
		// that are already buffered.
		readErr := make(chan error, 1)
		reading := make(chan struct{})
		go func() {
			buf := make([]byte, 4096)
			close(reading)
			for {
				if _, rerr := cp.Read(buf); rerr != nil {
					readErr <- rerr
					return
				}
			}
		}()
		<-reading
		time.Sleep(300 * time.Millisecond) // the setup output drains; the read parks

		start := time.Now()
		_ = cp.Close()
		select {
		case rerr := <-readErr:
			// os.ErrClosed is what classifyReadError maps a canceled read to,
			// and what the recorder already treats as a normal end of session.
			if !errors.Is(rerr, os.ErrClosed) && !errors.Is(rerr, io.EOF) {
				t.Errorf("iteration %d: read ended with %v, want os.ErrClosed (canceled) or io.EOF", i, rerr)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("iteration %d: a read parked in the kernel was still blocked %s after Close: "+
				"CloseHandle alone does not abort it, so the reader leaks and the recorder waits out its drain grace",
				i, time.Since(start))
		}
	}
}
