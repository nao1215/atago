package ptyrun

import (
	"os"
	"strconv"
	"syscall"

	"github.com/nao1215/atago/internal/diag"
	"golang.org/x/sys/unix"
)

// openTerminalPair opens a pseudo-terminal pair the way any POSIX program does:
// open the multiplexer, ask the kernel which slave belongs to this master,
// unlock that slave, open it. It exists — instead of a call to creack/pty's
// Open, which does the same four steps — because of how creack/pty passes the
// answer buffer to the first ioctl.
//
// Its ptsname declares the index as a local variable and hands the ioctl
// `uintptr(unsafe.Pointer(&n))`, then passes that uintptr down through two
// ordinary function calls before the syscall happens. A uintptr is a number, not
// a pointer: it neither keeps the variable alive nor pins it. The Go runtime
// moves goroutine stacks — it copies them on growth and can shrink them while
// scanning for the garbage collector — so if the stack moves anywhere in that
// window, the kernel writes the index into the abandoned stack and the variable
// still reads its zero value. creack/pty then reports the slave as
// "/dev/pts/0", which under Linux's flat namespace is some other program's
// terminal, and hands atago a "pair" whose two halves are not connected.
//
// The consequences are exactly the flake in #385, seen roughly once per 10^5
// pty steps under load and observed directly on Linux 7.0 with the master's
// real index at 10 and the slave at /dev/pts/0, still held open by a desktop
// daemon: the child writes its output into that stranger's terminal, the master
// atago drains never gets a slave at all (so its reads answer EAGAIN forever
// rather than ending with EIO), the session-end grace elapses, atago closes the
// master, and the step reports a clean exit 0 with an empty transcript.
//
// Doing the ioctl here keeps the conversion inside the syscall call expression,
// which is the one form the unsafe.Pointer rules guarantee — x/sys/unix marks
// those wrappers //go:uintptrescapes, so the argument is forced to the heap and
// stays put for the duration of the call. The index is then always the kernel's
// answer, never a zero left behind by a stack that moved.
//
// macOS is not affected and still goes through creack/pty: its ptsname asks for
// the name into a heap-allocated slice, and Go's heap does not move.
func openTerminalPair() (master, tty *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = master.Close()
		}
	}()

	var index int
	if ierr := ControlFD(master, func(fd int) error {
		var gerr error
		index, gerr = unix.IoctlGetInt(fd, unix.TIOCGPTN)
		return gerr
	}); ierr != nil {
		return nil, nil, diag.PTYFailed.Errorf("asking the kernel which terminal /dev/ptmx just handed out: %w", ierr)
	}
	// TIOCSPTLCK with a zero clears the lock the kernel puts on a freshly
	// allocated slave; until it is cleared, opening the slave fails with EIO.
	if ierr := ControlFD(master, func(fd int) error {
		return unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0)
	}); ierr != nil {
		return nil, nil, diag.PTYFailed.Errorf("unlocking terminal /dev/pts/%d: %w", index, ierr)
	}

	name := "/dev/pts/" + strconv.Itoa(index)
	tty, err = os.OpenFile(name, os.O_RDWR|syscall.O_NOCTTY, 0) //nolint:gosec // the path is /dev/pts/ plus an integer the kernel just gave us
	if err != nil {
		return nil, nil, err
	}
	return master, tty, nil
}
