package cmd

import (
	"errors"
	"os/exec"
)

// ExitCode maps a finished process's wait error to the exit code atago reports:
// 0 for a clean exit, the process's own code for an ordinary failure, the
// platform's convention for a process terminated by a signal (128+signal on
// POSIX, see exitCodeFor), and -1 when the process could not be reaped at all.
//
// Every runner that reports an exit code goes through here — the cmd runner, the
// pty runner, and the interactive recorder — so the same program driven two ways
// reports the same number and a recorded session generates a spec a run: step
// can satisfy.
//
// Callers must resolve their own timeout and parent-cancel paths BEFORE calling.
// Those kill the child with a signal, and mapping that kill to 137 would erase
// the "timed out and was killed" diagnosis and turn a hang into a plausible
// exit code an author could assert; -1 stays the sentinel for "there is no exit
// status to report".
func ExitCode(waitErr error) int {
	if waitErr == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		return exitCodeFor(exitErr)
	}
	return -1
}
