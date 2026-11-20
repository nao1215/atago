package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// errSharingViolation stands in for the transient Windows error a reader gets
// while the writer is mid-rename or an antivirus scanner holds the file open.
var errSharingViolation = errors.New("the process cannot access the file because it is being used by another process")

// stubReadyFileReader swaps readReadyFile for the duration of the test. Each
// call consumes the next scripted result; the last one repeats once the script
// runs out.
func stubReadyFileReader(t *testing.T, results ...func() ([]byte, error)) *int {
	t.Helper()
	var mu sync.Mutex
	calls := 0
	orig := readReadyFile
	readReadyFile = func(string) ([]byte, error) {
		mu.Lock()
		i := calls
		calls++
		mu.Unlock()
		if i >= len(results) {
			i = len(results) - 1
		}
		return results[i]()
	}
	t.Cleanup(func() { readReadyFile = orig })
	return &calls
}

func ok(s string) func() ([]byte, error)    { return func() ([]byte, error) { return []byte(s), nil } }
func fail(err error) func() ([]byte, error) { return func() ([]byte, error) { return nil, err } }

// readyFileFixture writes a non-empty file so the existence/size check passes
// and only the stubbed read decides readiness. The Proc never exits.
func readyFileFixture(t *testing.T) (*Proc, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "addr.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	return &Proc{name: "addr", done: make(chan struct{})}, path
}

// TestWaitFile_RetriesTransientReadError reproduces the windows-11-arm failure
// "ATG4004: read ready file ...: being used by another process": the file
// exists and is non-empty, but the first reads hit a sharing violation. The
// wait must keep retrying within ready.timeout instead of failing on the
// first read error.
func TestWaitFile_RetriesTransientReadError(t *testing.T) {
	p, path := readyFileFixture(t)
	calls := stubReadyFileReader(t,
		fail(errSharingViolation),
		fail(errSharingViolation),
		ok("127.0.0.1:54321\r\n"),
	)
	got, err := p.waitFile(context.Background(), path, "addr", 5*time.Second)
	if err != nil {
		t.Fatalf("waitFile() error = %v, want success after transient read errors", err)
	}
	if got != "127.0.0.1:54321" {
		t.Errorf("waitFile() = %q, want 127.0.0.1:54321", got)
	}
	if *calls != 3 {
		t.Errorf("read calls = %d, want 3", *calls)
	}
}

// TestWaitFile_WaitsPastBlankContent covers a reader that catches the file
// before its content is flushed: whitespace-only content is not a value to
// bind, so the wait continues until real content appears.
func TestWaitFile_WaitsPastBlankContent(t *testing.T) {
	p, path := readyFileFixture(t)
	stubReadyFileReader(t, ok(""), ok(" \r\n"), ok("127.0.0.1:1\n"))
	got, err := p.waitFile(context.Background(), path, "addr", 5*time.Second)
	if err != nil {
		t.Fatalf("waitFile() error = %v", err)
	}
	if got != "127.0.0.1:1" {
		t.Errorf("waitFile() = %q, want 127.0.0.1:1", got)
	}
}

// TestWaitFile_PersistentReadErrorTimesOut keeps a read error that never
// clears: the wait ends at ready.timeout, and the error names the last read
// failure so the cause is not hidden behind a bare timeout.
func TestWaitFile_PersistentReadErrorTimesOut(t *testing.T) {
	p, path := readyFileFixture(t)
	stubReadyFileReader(t, fail(errSharingViolation))
	_, err := p.waitFile(context.Background(), path, "addr", 100*time.Millisecond)
	if err == nil {
		t.Fatal("waitFile() error = nil, want a readiness timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error %q should report the timeout", err)
	}
	if !strings.Contains(err.Error(), "being used by another process") {
		t.Errorf("error %q should carry the last read error", err)
	}
}
