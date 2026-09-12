//go:build windows

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

// TestRun_CancelKillsTheWholeTree is the Windows counterpart of the POSIX
// process-group kill. A cancelled `shell: true` step must take its
// grandchildren with it: killing only the spawned cmd.exe left everything it
// had started running, holding the captured stdout pipe open until WaitDelay
// force-closed it — so the step reported a truncated stream AND leaked live
// processes into the rest of the run.
//
// The step starts a NESTED cmd.exe appending to beat.txt in a tight loop.
// `echo ... >>` reopens the file every iteration, so its size is an unbuffered
// witness of whether the tree is still alive: once Run has returned, the file
// must stop growing.
func TestRun_CancelKillsTheWholeTree(t *testing.T) {
	t.Parallel()
	// Not t.TempDir(): if the tree DOES survive, its open handles make TempDir's
	// cleanup fatal, which would report the bug as an unrelated cleanup failure
	// instead of the assertion below.
	//nolint:usetesting // Deliberate, for the reason above.
	wd, err := os.MkdirTemp("", "atago-tree-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(wd) })

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// The loop appends thousands of times a second, so this is ample for the
		// "did not grow" check below to have something real to compare against.
		time.Sleep(time.Second)
		cancel()
	}()
	// The iteration count only has to outlast the test; the loop is never
	// expected to finish on its own.
	const command = `cmd /c for /L %i in (1,1,100000000) do echo tick>>beat.txt`
	if _, err := New().Run(ctx, &spec.Run{Shell: spec.Bool(true), Command: command}, wd); err == nil {
		t.Fatal("Run() error = nil; a parent-cancel kill must be reported as an error")
	}

	beat := filepath.Join(wd, "beat.txt")
	before := beatSize(t, beat)
	if before == 0 {
		t.Fatal("beat.txt is empty: the nested loop never ran, so this test proves nothing about the kill")
	}
	// A survivor appends continuously, so any grace at all exposes it.
	time.Sleep(time.Second)
	if after := beatSize(t, beat); after != before {
		t.Errorf("beat.txt grew from %d to %d bytes after Run returned: the grandchild survived the cancel", before, after)
	}
}

// beatSize returns the size of the witness file, treating "not created yet" as
// zero so the caller's own emptiness check produces the diagnostic.
func beatSize(t *testing.T, path string) int64 {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}
