//go:build !windows

package ptyrun

import (
	"context"
	"os"
	"strings"
	"testing"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// TestRun_FastExitOutputNotLost is a regression test for a drain race: a child
// that writes and exits immediately could lose its output when the master was
// closed before the reader goroutine drained the pty buffer (seen as a flaky
// examples/pty.atago.yaml failure under coverage instrumentation). Repeating
// the fast-exit case many times makes the lost-output window reliably visible.
func TestRun_FastExitOutputNotLost(t *testing.T) {
	t.Parallel()

	shell := true
	for i := range 200 {
		p := &spec.PTY{
			Shell:   &shell,
			Command: "if [ -t 0 ]; then echo is-a-tty; else echo is-a-pipe; fi",
		}
		res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if ef != nil {
			t.Fatalf("iteration %d: unexpected expect failure: %+v", i, ef)
		}
		if res.ExitCode != 0 {
			t.Fatalf("iteration %d: exit code = %d, want 0 (stdout %q)", i, res.ExitCode, res.Stdout)
		}
		if !strings.Contains(string(res.Stdout), "is-a-tty") {
			t.Fatalf("iteration %d: transcript lost the child's output: %q", i, res.Stdout)
		}
	}
}

// TestRun_NoShellFastExitOutputNotLost covers the exact CrossPlatformE2E flake
// shape from the scheduled macOS job: a no-session pty step running the real
// `echo` binary exited 0 while its captured stdout came back empty. Repeating
// the no-shell case keeps the regression guard aligned with the self-hosted
// record spec that exercises `record --pty -- echo done`.
func TestRun_NoShellFastExitOutputNotLost(t *testing.T) {
	t.Parallel()

	for i := range 200 {
		p := &spec.PTY{Command: "echo done"}
		res, ef, err := Run(context.Background(), p, t.TempDir(), nil)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if ef != nil {
			t.Fatalf("iteration %d: unexpected expect failure: %+v", i, ef)
		}
		if res.ExitCode != 0 {
			t.Fatalf("iteration %d: exit code = %d, want 0 (stdout %q)", i, res.ExitCode, res.Stdout)
		}
		if !strings.Contains(string(res.Stdout), "done") {
			t.Fatalf("iteration %d: transcript lost the child's output: %q", i, res.Stdout)
		}
	}
}

// TestResolveCwd covers cwd resolution for a pty step: empty stays at the
// workdir, an absolute path is used verbatim, and a relative path nests inside
// the workdir — matching the cmd runner's rule so a pty and a run step agree on
// where a relative cwd points.
func TestResolveCwd(t *testing.T) {
	t.Parallel()
	const wd = "/work"
	cases := []struct{ cwd, want string }{
		{"", wd},
		{"sub", "/work" + string(os.PathSeparator) + "sub"},
		{"/abs", "/abs"},
	}
	for _, c := range cases {
		if got := runnercmd.ResolveDir(wd, c.cwd); got != c.want {
			t.Errorf("ResolveDir(%q, %q) = %q, want %q", wd, c.cwd, got, c.want)
		}
	}
}
