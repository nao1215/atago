//go:build windows

package cmd

import (
	"context"
	"os/exec"
	"strconv"
	"time"
)

// configureCancellation makes a cancelled command tear down its whole process
// tree, the Windows counterpart of the POSIX build's process-group kill.
//
// Windows has no process groups, so `taskkill /T` is the equivalent: it walks
// the LIVE tree at kill time, which is race-free with respect to a descendant
// spawned moments earlier. It is the same mechanism the pty runner
// (ptyrun.killTree) and the background service runner already use, so all three
// runners now leave nothing behind when a step is cancelled or times out.
//
// Before this, CommandContext's default cancel killed only the spawned leader:
// a `shell: true` step that started anything else left it running, holding the
// captured stdout pipe open until WaitDelay force-closed it — so the step
// reported a truncated stream and the survivor leaked into the rest of the run.
// WaitDelay stays as the backstop for a descendant taskkill cannot reach.
func configureCancellation(ctx context.Context, cmd *exec.Cmd) {
	// The teardown runs precisely because ctx is done, so it keeps ctx's VALUES
	// and drops its cancellation: handing taskkill an already-cancelled context
	// would kill the killer before it reached the tree.
	teardown := context.WithoutCancel(ctx)
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			killTree(teardown, cmd.Process.Pid)
		}
		// The tree is going away either way and os/exec reports the outcome
		// through Wait; a taskkill exit status adds nothing a caller can act on.
		return nil
	}
	cmd.WaitDelay = 2 * time.Second
}

// killTree force-terminates pid and every descendant. It mirrors
// ptyrun.killTree; both are fire-and-forget, because the only failure worth
// distinguishing — the process is already gone — needs no handling.
func killTree(ctx context.Context, pid int) {
	_ = exec.CommandContext(ctx, "taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run() //nolint:gosec // fixed argv, pid from our own child
}
