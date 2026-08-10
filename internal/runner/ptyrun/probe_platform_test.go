//go:build !windows

package ptyrun

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"
)

// TestProbe_MasterAfterLastSlaveClose is a temporary diagnostic, not a contract:
// it reports whether a read on the master still returns bytes the slave wrote
// before the last slave handle closed. It never fails, so both answers reach the
// log on every platform.
func TestProbe_MasterAfterLastSlaveClose(t *testing.T) {
	master, tty, err := OpenTerminal(24, 80)
	if err != nil {
		t.Fatalf("open terminal: %v", err)
	}
	defer func() { _ = master.Close() }()

	if _, werr := tty.WriteString("done\n"); werr != nil {
		t.Fatalf("write from the slave: %v", werr)
	}
	// Give the line discipline time to make the bytes readable, then drop the
	// last slave handle before any read is posted — the fast-exit shape.
	time.Sleep(50 * time.Millisecond)
	if cerr := tty.Close(); cerr != nil {
		t.Fatalf("close the slave: %v", cerr)
	}

	buf := make([]byte, 4096)
	var got []byte
	var readErr error
	for range 5 {
		n, rerr := master.Read(buf)
		if n > 0 {
			got = append(got, buf[:n]...)
		}
		if rerr != nil {
			readErr = rerr
			break
		}
	}
	t.Logf("PROBE after-last-slave-close: bytes=%q err=%v eof=%v", got, readErr, errors.Is(readErr, io.EOF))
}

// TestProbe_MasterReadPostedBeforeClose is the control: the same write and close
// with a read already parked in the kernel. If this one keeps the bytes and the
// probe above loses them, the loss window is the unposted read.
func TestProbe_MasterReadPostedBeforeClose(t *testing.T) {
	master, tty, err := OpenTerminal(24, 80)
	if err != nil {
		t.Fatalf("open terminal: %v", err)
	}
	defer func() { _ = master.Close() }()

	type readResult struct {
		data []byte
		err  error
	}
	done := make(chan readResult, 1)
	go func() {
		buf := make([]byte, 4096)
		var got []byte
		var rerr error
		for range 5 {
			n, e := master.Read(buf)
			if n > 0 {
				got = append(got, buf[:n]...)
			}
			if e != nil {
				rerr = e
				break
			}
		}
		done <- readResult{got, rerr}
	}()
	// Let the reader park in read(2) before anything is written.
	time.Sleep(100 * time.Millisecond)

	if _, werr := tty.WriteString("done\n"); werr != nil {
		t.Fatalf("write from the slave: %v", werr)
	}
	if cerr := tty.Close(); cerr != nil {
		t.Fatalf("close the slave: %v", cerr)
	}

	select {
	case r := <-done:
		t.Logf("PROBE posted-read: bytes=%q err=%v", r.data, r.err)
	case <-time.After(5 * time.Second):
		t.Log("PROBE posted-read: the read never ended")
	}
}

// TestProbe_MasterPollerOwnership records which regime this platform's master
// lands in, since that decides whether Close can interrupt a parked read.
func TestProbe_MasterPollerOwnership(t *testing.T) {
	master, tty, err := OpenTerminal(24, 80)
	if err != nil {
		t.Fatalf("open terminal: %v", err)
	}
	defer func() { _ = tty.Close() }()
	defer func() { _ = master.Close() }()

	derr := master.SetReadDeadline(time.Time{})
	t.Logf("PROBE poller-owned: setReadDeadlineErr=%v noDeadline=%v", derr, errors.Is(derr, os.ErrNoDeadline))
}
