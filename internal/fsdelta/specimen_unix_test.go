//go:build !windows

package fsdelta

import (
	"context"
	"net"
	"os"
	"syscall"
	"testing"
)

// mkfifoAt plants a named pipe. The delta must record it without ever opening
// it: opening a FIFO for reading blocks until a writer appears, which would hang
// the scan rather than fail it.
func mkfifoAt(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := syscall.Mkfifo(path, uint32(mode)); err != nil {
		t.Fatal(err)
	}
	// Mkfifo applies the umask, so state the mode explicitly.
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// socketAt plants a unix domain socket by binding a listener to the path. The
// listener is closed on cleanup, which is also what unlinks the socket.
func socketAt(t *testing.T, path string) {
	t.Helper()
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
}
