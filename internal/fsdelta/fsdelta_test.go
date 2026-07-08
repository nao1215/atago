package fsdelta

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"
	"time"
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

// TestScan_MissingRootReturnsEmpty covers the os.IsNotExist branch: scanning a
// root that does not exist yields an empty snapshot and no error, so a step that
// never created its workdir does not turn a `changes:` assertion into an engine
// error.
func TestScan_MissingRootReturnsEmpty(t *testing.T) {
	t.Parallel()
	snap, err := Scan(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("Scan(missing) error = %v, want nil", err)
	}
	if len(snap) != 0 {
		t.Errorf("Scan(missing) = %v, want empty snapshot", snap)
	}
}

// TestScan_SkipsSymlinks proves symlinks are not tracked (only regular files
// are), so a symlink cannot be mistaken for a created/modified content file.
func TestScan_SkipsSymlinks(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is restricted on Windows")
	}
	root := t.TempDir()
	target := filepath.Join(root, "real.txt")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}
	snap, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if _, ok := snap["link.txt"]; ok {
		t.Errorf("symlink should not be tracked, snapshot = %v", snap)
	}
	if _, ok := snap["real.txt"]; !ok {
		t.Errorf("regular file should be tracked, snapshot = %v", snap)
	}
}

// TestDiff_Antisymmetry is a metamorphic law: swapping the two snapshots turns
// every Created into a Deleted and vice versa, while the Modified set is
// unchanged (only its membership flips direction, not the paths). A break here
// would mean the delta reports the wrong verb for a change.
func TestDiff_Antisymmetry(t *testing.T) {
	t.Parallel()
	pre := Snapshot{"keep": "h", "changed": "old", "onlypre": "x"}
	post := Snapshot{"keep": "h", "changed": "new", "onlypost": "y"}

	fwd := Diff(pre, post)
	rev := Diff(post, pre)

	if !reflect.DeepEqual(fwd.Created, rev.Deleted) {
		t.Errorf("forward Created %v != reverse Deleted %v", fwd.Created, rev.Deleted)
	}
	if !reflect.DeepEqual(fwd.Deleted, rev.Created) {
		t.Errorf("forward Deleted %v != reverse Created %v", fwd.Deleted, rev.Created)
	}
	fm, rm := append([]string(nil), fwd.Modified...), append([]string(nil), rev.Modified...)
	sort.Strings(fm)
	sort.Strings(rm)
	if !reflect.DeepEqual(fm, rm) {
		t.Errorf("Modified set differs by direction: %v vs %v", fm, rm)
	}
	if len(fwd.Modified) != 1 || fwd.Modified[0] != "changed" {
		t.Errorf("Modified = %v, want [changed]", fwd.Modified)
	}
}

// TestDiff_RenameIsDeletePlusCreate pins the documented v1 semantics: a
// content-preserving rename is reported as a delete of the old path and a
// create of the new one, never as a single move.
func TestDiff_RenameIsDeletePlusCreate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "a.txt", "same-bytes")
	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}
	if err := os.Rename(filepath.Join(root, "a.txt"), filepath.Join(root, "b.txt")); err != nil {
		t.Fatal(err)
	}
	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}
	got := Diff(pre, post)
	want := Delta{Created: []string{"b.txt"}, Deleted: []string{"a.txt"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rename Diff() = %+v, want %+v", got, want)
	}
}

// TestDiff_DeleteRecreateIdenticalIsNoChange proves the delta is content-based,
// not timestamp-based: deleting a file and recreating it with identical bytes
// reports no change, so a step that rewrites a file to the same content does not
// churn a `changes:` assertion.
func TestDiff_DeleteRecreateIdenticalIsNoChange(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "f.txt", "content")
	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "f.txt")); err != nil {
		t.Fatal(err)
	}
	write(t, root, "f.txt", "content")
	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}
	if d := Diff(pre, post); len(d.Created)+len(d.Modified)+len(d.Deleted) != 0 {
		t.Errorf("delete+recreate identical Diff() = %+v, want all empty", d)
	}
}

// TestScan_EmptyFileAndTruncation covers the zero-byte boundary: an empty file
// is a tracked create (distinct from a deleted file, which is absent), and
// truncating a file to empty is a modification.
func TestScan_EmptyFileAndTruncation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "shrink.txt", "not empty yet")
	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}
	write(t, root, "empty.txt", "")
	write(t, root, "shrink.txt", "")
	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}
	got := Diff(pre, post)
	want := Delta{Created: []string{"empty.txt"}, Modified: []string{"shrink.txt"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("empty-file Diff() = %+v, want %+v", got, want)
	}
}

// TestScan_UnreadableFileStillTracked proves a created regular file that cannot
// be read (mode 000) is not silently dropped: it is recorded so `created: []`
// cannot pass for a step that planted an unreadable file, and its deletion is
// reported too. An unreadable file present in both snapshots is unchanged.
func TestScan_UnreadableFileStillTracked(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Unix file modes; chmod 000 does not deny reads the same way on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses mode 000 read permission")
	}
	root := t.TempDir()
	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}
	locked := filepath.Join(root, "locked.txt")
	if err := os.WriteFile(locked, []byte("secret"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o600) })
	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}
	got := Diff(pre, post)
	if want := []string{"locked.txt"}; !reflect.DeepEqual(got.Created, want) {
		t.Errorf("created = %v, want %v (an unreadable created file must be reported)", got.Created, want)
	}
	// Deletion of an unreadable file is reported as the mirror.
	if rev := Diff(post, pre); !reflect.DeepEqual(rev.Deleted, []string{"locked.txt"}) {
		t.Errorf("deleted = %v, want [locked.txt]", rev.Deleted)
	}
	// Unreadable in both snapshots: unchanged, reported nowhere.
	if stable := Diff(post, post); len(stable.Created)+len(stable.Modified)+len(stable.Deleted) != 0 {
		t.Errorf("unreadable-in-both should be unchanged, got %+v", stable)
	}
}

// TestDiff_FileReplacedByDirectoryAndReverse covers a name changing kind: a file
// replaced by a directory of the same name deletes the file and creates the
// directory's contents, and the reverse deletes the contents and creates the
// file. Directories themselves are never tracked, only the regular files.
func TestDiff_FileReplacedByDirectoryAndReverse(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "foo", "i am a file")
	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "foo")); err != nil {
		t.Fatal(err)
	}
	write(t, root, "foo/bar.txt", "now a dir")
	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}
	got := Diff(pre, post)
	want := Delta{Created: []string{"foo/bar.txt"}, Deleted: []string{"foo"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("file->dir Diff() = %+v, want %+v", got, want)
	}
	// The reverse must be the mirror image.
	rev := Diff(post, pre)
	wantRev := Delta{Created: []string{"foo"}, Deleted: []string{"foo/bar.txt"}}
	if !reflect.DeepEqual(rev, wantRev) {
		t.Errorf("dir->file Diff() = %+v, want %+v", rev, wantRev)
	}
}

// TestScan_BrokenAndCyclicSymlinksNoCrash proves Scan never follows a symlink
// into a crash or infinite loop: a dangling link and a self-referential cyclic
// link are both skipped (only regular files are tracked), leaving the real file.
func TestScan_BrokenAndCyclicSymlinksNoCrash(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is restricted on Windows")
	}
	root := t.TempDir()
	write(t, root, "real.txt", "data")
	if err := os.Symlink(filepath.Join(root, "missing-target"), filepath.Join(root, "broken")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "loop"), filepath.Join(root, "loop")); err != nil {
		t.Fatal(err)
	}
	snap, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan with broken/cyclic symlinks: %v", err)
	}
	if _, ok := snap["real.txt"]; !ok {
		t.Errorf("regular file should be tracked, snapshot = %v", keys(snap))
	}
	for _, link := range []string{"broken", "loop"} {
		if _, ok := snap[link]; ok {
			t.Errorf("symlink %q should not be tracked, snapshot = %v", link, keys(snap))
		}
	}
}

// TestDiff_MetadataOnlyChangeNotDetected pins that the delta is purely
// content-based: a chmod-only or mtime-only change leaves the hash unchanged and
// is reported nowhere, so `changes:` does not flag a file whose bytes are equal.
func TestDiff_MetadataOnlyChangeNotDetected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, root, "perm.txt", "bytes")
	write(t, root, "time.txt", "bytes")
	pre, err := Scan(root)
	if err != nil {
		t.Fatalf("pre scan: %v", err)
	}
	if err := os.Chmod(filepath.Join(root, "perm.txt"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(filepath.Join(root, "time.txt"), old, old); err != nil {
		t.Fatal(err)
	}
	post, err := Scan(root)
	if err != nil {
		t.Fatalf("post scan: %v", err)
	}
	if d := Diff(pre, post); len(d.Created)+len(d.Modified)+len(d.Deleted) != 0 {
		t.Errorf("metadata-only Diff() = %+v, want all empty (content unchanged)", d)
	}
}

// TestDiff_EmptySnapshots covers the trivial edges: two empty snapshots produce
// an all-empty delta, and a scan-against-nothing reports everything created.
func TestDiff_EmptySnapshots(t *testing.T) {
	t.Parallel()
	empty := Snapshot{}
	if d := Diff(empty, empty); len(d.Created)+len(d.Modified)+len(d.Deleted) != 0 {
		t.Errorf("Diff(empty,empty) = %+v, want all empty", d)
	}
	post := Snapshot{"a": "1", "b": "2"}
	d := Diff(empty, post)
	want := []string{"a", "b"}
	if !reflect.DeepEqual(d.Created, want) || len(d.Modified) != 0 || len(d.Deleted) != 0 {
		t.Errorf("Diff(empty, post) = %+v, want Created %v", d, want)
	}
}
