package ptyrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	"github.com/nao1215/atago/internal/spec"
)

// maxExecOutputBytes bounds what is kept from a mid-session host command's
// output. It is only ever read to explain a failure, so a tail is enough — and
// a command that floods must not grow the process's memory while a TUI is being
// driven.
const maxExecOutputBytes = 64 << 10

// execWaitDelay bounds how long Run may keep waiting on inherited output pipes
// after the command itself has been killed. See the comment at its use.
const execWaitDelay = 2 * time.Second

// runSessionExec runs one host command to completion while the program under
// test keeps running (#380).
//
// Every failure here is a hard error rather than an assertion failure. The
// command is scaffolding — it makes the change the session is about to watch
// the program react to — so a broken one must stop the run at the mistake. The
// alternative is worse than useless: the expect_screen that follows would wait
// out its whole timeout for a change nobody made, and report the program as the
// thing at fault.
func runSessionExec(ctx context.Context, e *spec.PTYExec, dir string, env []string) error {
	name, args, err := runnercmd.CommandLine(e.Command, e.ShellEnabled())
	if err != nil {
		return err
	}

	// The session budget bounds this too: whichever runs out first ends it, so a
	// helper can never outlive the session it belongs to.
	timeout := spec.DefaultPTYExecTimeout
	if e.Timeout != "" {
		if d, perr := time.ParseDuration(e.Timeout); perr == nil && d > 0 {
			timeout = d
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, name, args...) //nolint:gosec // the spec author's declared command is the point
	cmd.Dir = dir
	cmd.Env = env
	// Killing the command on timeout is not enough to get Run back: a helper run
	// through the shell leaves grandchildren holding the output pipes open, and
	// Run waits for those to close. `sh -c 'sleep 30'` bounded at 150ms took the
	// full thirty seconds until this was added — the whole point of the timeout
	// is that a stuck helper does not hold the session, so cap the grace after
	// the kill rather than waiting on an orphan.
	cmd.WaitDelay = execWaitDelay
	// Captured, not streamed: this output is not what the TERMINAL showed, so it
	// has no business in the transcript. It exists to explain a failure.
	var out cappedBuffer
	out.limit = maxExecOutputBytes
	cmd.Stdout = &out
	cmd.Stderr = &out

	start := time.Now()
	runErr := cmd.Run()
	if runErr == nil {
		return nil
	}

	detail := out.String()
	switch {
	case errors.Is(runCtx.Err(), context.DeadlineExceeded):
		return fmt.Errorf("exec %q did not finish within %s%s", e.Command, timeout, execOutputSuffix(detail))
	case errors.Is(ctx.Err(), context.Canceled):
		return fmt.Errorf("exec %q was canceled after %s: %w", e.Command, time.Since(start).Round(time.Millisecond), ctx.Err())
	}
	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return fmt.Errorf("exec %q exited %d, so the change the session waits for was not made%s",
			e.Command, exitErr.ExitCode(), execOutputSuffix(detail))
	}
	return fmt.Errorf("exec %q could not run: %w%s", e.Command, runErr, execOutputSuffix(detail))
}

func execOutputSuffix(detail string) string {
	if strings.TrimSpace(detail) == "" {
		return " (it printed nothing)"
	}
	return "\noutput:\n" + detail
}

// cappedBuffer keeps at most limit bytes and says how many it dropped, so a
// failure message stays useful without letting a chatty command grow without
// bound.
type cappedBuffer struct {
	buf     bytes.Buffer
	limit   int
	dropped int
}

// Write always reports the full write: a short count would make the command see
// a broken pipe on its own output, which is not what happened — atago simply
// stopped keeping the excess.
func (c *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	switch room := c.limit - c.buf.Len(); {
	case room <= 0:
		c.dropped += n
	case n > room:
		c.dropped += n - room
		c.buf.Write(p[:room])
	default:
		c.buf.Write(p)
	}
	return n, nil
}

func (c *cappedBuffer) String() string {
	if c.dropped == 0 {
		return c.buf.String()
	}
	return fmt.Sprintf("%s\n... and %d more bytes", c.buf.String(), c.dropped)
}
