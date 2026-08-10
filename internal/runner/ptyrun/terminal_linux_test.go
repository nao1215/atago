package ptyrun

import (
	"strconv"
	"testing"

	"golang.org/x/sys/unix"
)

// TestOpenTerminalPair_SlaveIsTheOneTheKernelNamed states the Linux half of
// #385 in the kernel's own terms: the slave atago opens must be the slave the
// master was given, and the only authority on which one that is is TIOCGPTN on
// the master.
//
// creack/pty asks the same question but lets the answer be written into a stack
// variable it no longer keeps alive across the call, so a stack that moved left
// the index reading 0 and atago opened /dev/pts/0 — a terminal that belongs to
// whatever program happens to hold it. Comparing the opened name against a
// freshly read index says that in one line, and unlike the transcript-level
// guards it does not need thousands of spawns to have an opinion.
func TestOpenTerminalPair_SlaveIsTheOneTheKernelNamed(t *testing.T) {
	t.Parallel()

	for i := range 200 {
		master, tty, err := openTerminalPair()
		if err != nil {
			t.Fatalf("iteration %d: openTerminalPair: %v", i, err)
		}
		var index int
		if ierr := ControlFD(master, func(fd int) error {
			var gerr error
			index, gerr = unix.IoctlGetInt(fd, unix.TIOCGPTN)
			return gerr
		}); ierr != nil {
			t.Fatalf("iteration %d: TIOCGPTN on the master: %v", i, ierr)
		}
		if want := "/dev/pts/" + strconv.Itoa(index); tty.Name() != want {
			t.Fatalf("iteration %d: opened %q as the slave, but the master's terminal is %q",
				i, tty.Name(), want)
		}
		_ = tty.Close()
		_ = master.Close()
	}
}
