package assert

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// seedTree builds a small deterministic tree for the walk/snapshot tests.
func seedTree(t *testing.T) string {
	t.Helper()
	wd := t.TempDir()
	mustWrite := func(rel, content string) {
		t.Helper()
		p := filepath.Join(wd, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("site/content/posts/hello.md", "hello\n")
	mustWrite("site/hugo.toml", "baseURL = 'x'\n")
	mustWrite("site/debug.log", "noise\n")
	mustWrite("site/.git/config", "gitstuff\n")
	return wd
}

// TestWalkTree_DeterministicAndIgnored proves ordering, /-separated paths,
// basename globs, and subtree pruning (#25).
func TestWalkTree_DeterministicAndIgnored(t *testing.T) {
	t.Parallel()
	wd := seedTree(t)
	entries, err := walkTree(filepath.Join(wd, "site"), []string{"*.log", ".git/**"})
	if err != nil {
		t.Fatalf("walkTree: %v", err)
	}
	var rels []string
	for _, e := range entries {
		rels = append(rels, e.kind+" "+e.rel)
	}
	got := strings.Join(rels, "\n")
	want := strings.Join([]string{
		"dir content",
		"dir content/posts",
		"file content/posts/hello.md",
		"file hugo.toml",
	}, "\n")
	if got != want {
		t.Errorf("walk = \n%s\nwant\n%s", got, want)
	}
}

// TestWalkTree_SymlinkRecordedNotTraversed proves symlinks appear as `link`
// entries with their target and their subtree is never walked (#25).
func TestWalkTree_SymlinkRecordedNotTraversed(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on Windows")
	}
	wd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wd, "real"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wd, "real", "f.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(wd, "alias")); err != nil {
		t.Fatal(err)
	}
	entries, err := walkTree(wd, nil)
	if err != nil {
		t.Fatalf("walkTree: %v", err)
	}
	var lines []string
	for _, e := range entries {
		lines = append(lines, e.manifestLine())
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "link alias -> real") {
		t.Errorf("manifest missing the link entry:\n%s", joined)
	}
	if strings.Contains(joined, "alias/f.txt") {
		t.Errorf("symlink was traversed:\n%s", joined)
	}
}

// TestCheckDir_RecursiveMatchers proves nested contains, file-only counts,
// and deep globs (#25).
func TestCheckDir_RecursiveMatchers(t *testing.T) {
	t.Parallel()
	wd := seedTree(t)
	env := Env{Workdir: wd}

	pass := &spec.DirAssert{
		Path:      "site",
		Recursive: true,
		Ignore:    []string{"*.log", ".git/**"},
		Contains:  []string{"content/posts/hello.md"},
		Count:     intp(2), // hello.md + hugo.toml; dirs and ignored files excluded
		Glob:      "*.md",
	}
	if cr := checkDir(pass, env); !cr.OK {
		t.Fatalf("recursive matchers failed: %+v", cr)
	}

	fail := &spec.DirAssert{Path: "site", Recursive: true, Contains: []string{"content/missing.md"}}
	if cr := checkDir(fail, env); cr.OK {
		t.Fatal("missing nested path passed")
	}
}

// TestManifestDiff_PathsWithSpaces proves the diff parser follows the line
// grammar instead of splitting on whitespace, so spaced filenames survive.
func TestManifestDiff_PathsWithSpaces(t *testing.T) {
	t.Parallel()
	expected := "file My File.txt sha256:aaa\nlink my link -> some target\n"
	actual := "file My File.txt sha256:bbb\ndir new dir\n"
	got := manifestDiff(expected, actual)
	for _, want := range []string{
		"changed: My File.txt",
		"added:   dir new dir",
		"removed: link my link -> some target",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("diff = %q, want %q", got, want)
		}
	}
}

// TestManifestLine_EscapesControlBytes is a regression: a filesystem name
// carrying a newline (legal on POSIX) must render as ONE manifest line, or a
// single such entry produces the same manifest text as a structurally different
// multi-entry tree and falsely matches its golden. The escape must also be
// unambiguous, so a name containing a literal backslash-n does not collide with
// a name containing a real newline.
func TestManifestLine_EscapesControlBytes(t *testing.T) {
	t.Parallel()
	newlineDir := treeEntry{rel: "a\ndir b", kind: "dir"}.manifestLine()
	if strings.ContainsAny(newlineDir, "\n\r") {
		t.Errorf("manifestLine leaked a raw control byte, breaking one-line-per-entry: %q", newlineDir)
	}
	// The single newline-named entry must not equal the two-entry manifest it
	// previously collided with.
	twoEntries := treeEntry{rel: "a", kind: "dir"}.manifestLine() + "\n" +
		treeEntry{rel: "dir b", kind: "dir"}.manifestLine()
	if newlineDir == twoEntries {
		t.Errorf("newline-named entry still collides with a two-entry tree: %q", newlineDir)
	}
	// A literal backslash-n name and a real-newline name must stay distinct.
	if literal := (treeEntry{rel: `a\ndir b`, kind: "dir"}).manifestLine(); literal == newlineDir {
		t.Errorf("literal backslash-n name collides with the real-newline name: %q", literal)
	}
	// Ordinary names are untouched, so existing goldens are unaffected.
	if got := (treeEntry{rel: "content/posts/hello.md", kind: "file", hash: "abc"}).manifestLine(); got != "file content/posts/hello.md sha256:abc" {
		t.Errorf("ordinary name changed: %q", got)
	}
}

// TestCheckDir_SnapshotNewlineNameNoFalseMatch proves the escape end to end: a
// tree of two dirs and a tree of one dir whose name embeds a newline must not
// match each other's golden. POSIX-only — Windows forbids newlines in names.
func TestCheckDir_SnapshotNewlineNameNoFalseMatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("newlines are not legal in Windows filenames")
	}
	t.Parallel()
	specDir := t.TempDir()

	// Golden recorded from a tree of two sibling dirs.
	twoDirs := t.TempDir()
	if err := os.MkdirAll(filepath.Join(twoDirs, "root", "a"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(twoDirs, "root", "dir b"), 0o750); err != nil {
		t.Fatal(err)
	}
	d := &spec.DirAssert{Path: "root", Snapshot: "tree_golden"}
	if cr := checkDir(d, Env{Workdir: twoDirs, SpecDir: specDir, UpdateSnapshots: true}); !cr.OK {
		t.Fatalf("update failed: %+v", cr)
	}

	// A different tree: one dir whose name is "a<newline>dir b".
	oneDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(oneDir, "root", "a\ndir b"), 0o750); err != nil {
		t.Fatal(err)
	}
	cr := checkDir(d, Env{Workdir: oneDir, SpecDir: specDir})
	if cr.OK {
		t.Error("a one-entry tree with a newline in its name falsely matched a two-dir golden")
	}
}

// TestCheckDir_SnapshotRoundTrip proves record → green compare → mutation
// diff naming exactly the changed path (#25).
func TestCheckDir_SnapshotRoundTrip(t *testing.T) {
	t.Parallel()
	wd := seedTree(t)
	specDir := t.TempDir()
	d := &spec.DirAssert{Path: "site", Snapshot: "site_tree", Ignore: []string{"*.log", ".git/**"}}

	// Record.
	if cr := checkDir(d, Env{Workdir: wd, SpecDir: specDir, UpdateSnapshots: true}); !cr.OK {
		t.Fatalf("update failed: %+v", cr)
	}
	// Green compare.
	if cr := checkDir(d, Env{Workdir: wd, SpecDir: specDir}); !cr.OK {
		t.Fatalf("compare after update failed: %+v", cr)
	}
	// Mutate one file and add another: the diff names exactly those paths.
	if err := os.WriteFile(filepath.Join(wd, "site", "hugo.toml"), []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wd, "site", "new.txt"), []byte("n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cr := checkDir(d, Env{Workdir: wd, SpecDir: specDir})
	if cr.OK {
		t.Fatal("mutated tree still matched the snapshot")
	}
	if !strings.Contains(cr.Actual, "changed: hugo.toml") {
		t.Errorf("Actual = %q, want changed: hugo.toml", cr.Actual)
	}
	if !strings.Contains(cr.Actual, "added:   file new.txt") {
		t.Errorf("Actual = %q, want the added file named", cr.Actual)
	}
	if strings.Contains(cr.Actual, "hello.md") {
		t.Errorf("Actual = %q, must not name unchanged paths", cr.Actual)
	}
}
