//go:build !windows

package ptyrun

import (
	"os"

	"golang.org/x/sys/unix"
)

// ControlFD runs fn with f's raw descriptor and returns fn's error, without
// handing the descriptor over to the caller the way (*os.File).Fd() does.
//
// The difference matters for every terminal atago opens. Fd() is documented to
// return a descriptor in blocking mode: os.File takes the descriptor out of the
// runtime poller and clears O_NONBLOCK on it, permanently and for every later
// read. A pty master that has been through Fd() therefore parks its reads inside
// read(2), where Close cannot reach them — which is the hang OpenTerminal exists
// to prevent. SyscallConn's Control hands the descriptor to fn for the duration
// of the call and leaves the file exactly as it found it, so an ioctl (a window
// size, a termios query) no longer costs the terminal its poller.
func ControlFD(f *os.File, fn func(fd int) error) error {
	sc, err := f.SyscallConn()
	if err != nil {
		return err
	}
	var inner error
	if cerr := sc.Control(func(fd uintptr) { inner = fn(int(fd)) }); cerr != nil {
		return cerr
	}
	return inner
}

// setTerminalSize resizes the terminal behind master. TIOCSWINSZ on the master
// is the whole mechanism: the kernel records the new size AND sends SIGWINCH to
// the terminal's foreground process group, so the child learns about a
// mid-session resize exactly as it would from a real window change (#379).
func setTerminalSize(master *os.File, rows, cols uint16) error {
	return ControlFD(master, func(fd int) error {
		return unix.IoctlSetWinsize(fd, unix.TIOCSWINSZ, &unix.Winsize{Row: rows, Col: cols})
	})
}

// terminalEchoMode asks the terminal whether its line discipline echoes typed
// input right now. tty is the slave side, whose attributes are the ones the
// program sets and the line discipline obeys; atago keeps its own handle on it
// until the child is reaped, which covers every send. A terminal that will not
// answer is echoModeUnknown rather than a failed step: the driver still has
// its older, timing-based way to tell.
func terminalEchoMode(tty *os.File) echoMode {
	mode := echoModeUnknown
	_ = ControlFD(tty, func(fd int) error {
		t, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
		if err != nil {
			return err
		}
		if t.Lflag&unix.ECHO != 0 {
			mode = echoModeOn
		} else {
			mode = echoModeOff
		}
		return nil
	})
	return mode
}
