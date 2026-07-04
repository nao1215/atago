package fsdelta

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDiff_CreatedModifiedDeletedUntouched exercises every category plus nested
// directories, and proves an untouched file appears in none of them (#70).
func TestDiff_CreatedModifiedDeletedUntouched(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "keep.txt", "same")
	write(t, root, "change.txt", "before")
	write(t, root, "gone.txt", "bye")
	write(t, root, "nested/old.txt", "old")

	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}

	// keep.txt untouched; change.txt modified; gone.txt & nested/old.txt deleted;
	// nested/new.css created.
	write(t, root, "change.txt", "after")
	if err := os.Remove(filepath.Join(root, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "nested", "old.txt")); err != nil {
		t.Fatal(err)
	}
	write(t, root, "nested/new.css", "body{}")

	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}

	got := Diff(pre, post)
	want := Delta{
		Created:  []string{"nested/new.css"},
		Modified: []string{"change.txt"},
		Deleted:  []string{"gone.txt", "nested/old.txt"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff() = %+v, want %+v", got, want)
	}
}

// TestScan_ForwardSlashKeys proves nested paths are keyed with forward slashes
// on every OS, so a spec's globs compare in one space (#70).
func TestScan_ForwardSlashKeys(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "a/b/c.txt", "x")
	snap, err := Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if _, ok := snap["a/b/c.txt"]; !ok {
		t.Errorf("key = %v, want a/b/c.txt (forward slashes)", keys(snap))
	}
}

// TestScan_SkipsDirectoriesOnly proves directories are not tracked as files: an
// empty directory produces no snapshot entry (a rename is delete+create) (#70).
func TestScan_SkipsDirectoriesOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "emptydir"), 0o755); err != nil {
		t.Fatal(err)
	}
	snap, err := Scan(root)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(snap) != 0 {
		t.Errorf("snapshot = %v, want empty (directories are not files)", keys(snap))
	}
}

func keys(m Snapshot) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
