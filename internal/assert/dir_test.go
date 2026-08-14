package assert

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

func ptrBool(b bool) *bool { return &b }
func ptrInt(i int) *int    { return &i }

// makeTree builds a small directory tree under a temp workdir and returns the
// workdir. Layout: site/ { index.html, about.html, assets/app.css }.
func makeTree(t *testing.T) string {
	t.Helper()
	wd := t.TempDir()
	site := filepath.Join(wd, "site")
	if err := os.MkdirAll(filepath.Join(site, "assets"), 0o750); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"index.html", "about.html"} {
		if err := os.WriteFile(filepath.Join(site, f), []byte("<html>"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(site, "assets", "app.css"), []byte("body{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	return wd
}

func checkDirOK(t *testing.T, wd string, d *spec.DirAssert) *CheckResult {
	t.Helper()
	return Check(&spec.Assert{Dir: d}, nil, Env{Workdir: wd})
}

func TestCheckDir_Exists(t *testing.T) {
	wd := makeTree(t)
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Exists: ptrBool(true)}); !cr.OK {
		t.Errorf("exists:true should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "missing", Exists: ptrBool(false)}); !cr.OK {
		t.Errorf("exists:false on missing dir should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "missing", Exists: ptrBool(true)}); cr.OK {
		t.Error("exists:true on missing dir should fail")
	}
}

func TestCheckDir_ChildrenAndCounts(t *testing.T) {
	wd := makeTree(t)
	// site has 3 direct entries: index.html, about.html, assets.
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Count: ptrInt(3)}); !cr.OK {
		t.Errorf("count 3 should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Count: ptrInt(2)}); cr.OK {
		t.Error("count 2 should fail")
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", MinCount: ptrInt(1), MaxCount: ptrInt(5)}); !cr.OK {
		t.Errorf("count in [1,5] should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Contains: []string{"index.html", "assets/app.css"}}); !cr.OK {
		t.Errorf("nested contains should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Contains: []string{"nope.html"}}); cr.OK {
		t.Error("missing child should fail contains")
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", NotContains: []string{"secret.txt"}}); !cr.OK {
		t.Errorf("absent forbidden child should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", NotContains: []string{"index.html"}}); cr.OK {
		t.Error("present forbidden child should fail")
	}
}

func TestCheckDir_Glob(t *testing.T) {
	wd := makeTree(t)
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Glob: "*.html"}); !cr.OK {
		t.Errorf("glob *.html should match: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Glob: "*.pdf"}); cr.OK {
		t.Error("glob *.pdf should not match")
	}
}

func TestCheckDir_PathConfinement(t *testing.T) {
	wd := makeTree(t)
	// The directory path itself may not escape the workdir.
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "../escape", Exists: ptrBool(true)}); cr.OK {
		t.Error("path escaping the workdir must be rejected")
	}
	// A child path may not escape the directory via traversal.
	cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Contains: []string{"../../etc/passwd"}})
	if cr.OK {
		t.Error("child path traversal must be rejected")
	}
}

// TestCheckDir_SymlinkedRootMayNotEscapeTheWorkdir pins the confinement this
// assertion documents. The path check is lexical, so it cannot see through a
// symlink the program under test planted: `path:` pointing at a link out of the
// workdir passed the check, and os.ReadDir then followed it and listed a
// directory the spec has no business reading. Resolving the link before the
// containment test closes that, and an in-workdir link stays allowed.
func TestCheckDir_SymlinkedRootMayNotEscapeTheWorkdir(t *testing.T) {
	wd := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("s"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(wd, "escape")); err != nil {
		t.Skipf("symlinks unsupported on this platform: %v", err)
	}
	for _, d := range []*spec.DirAssert{
		{Path: "escape", Contains: []string{"secret.txt"}},
		{Path: "escape", Recursive: true, Contains: []string{"secret.txt"}},
		{Path: "escape", Count: ptrInt(1)},
	} {
		if cr := checkDirOK(t, wd, d); cr.OK {
			t.Errorf("reading %+v through a symlink out of the workdir must be rejected", d)
		}
	}
}

// TestCheckDir_DanglingSymlinkedRootJudgedByItsTarget keeps the containment rule
// from depending on whether a link's target happens to exist. A dangling link
// out of the workdir resolves to nothing, so `exists: false` would otherwise
// pass — and passing tells the spec author that the external path is absent,
// which is a question about the filesystem outside the workdir. A dangling link
// that stays inside (a `latest ->` pointing at a release not built yet) is
// ordinary and must still answer.
//
// It runs the same pair twice: once against the workdir as given, and once
// against the same directory reached through a symlink. The second spelling is
// what CI runs on macOS, where /var and /tmp sit behind /private — a dangling
// link's declared target keeps the unresolved spelling, so comparing it only
// against the RESOLVED root refused a link that never left the workdir.
func TestCheckDir_DanglingSymlinkedRootJudgedByItsTarget(t *testing.T) {
	for _, tc := range []struct {
		name         string
		symlinkedDir bool
	}{
		{name: "workdir as given"},
		{name: "workdir reached through a symlink", symlinkedDir: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			wd := filepath.Join(base, "real")
			if err := os.Mkdir(wd, 0o750); err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(t.TempDir(), "absent")
			if err := os.Symlink(outside, filepath.Join(wd, "escape")); err != nil {
				t.Skipf("symlinks unsupported on this platform: %v", err)
			}
			if err := os.Symlink("releases/v3", filepath.Join(wd, "latest")); err != nil {
				t.Fatal(err)
			}
			// A chain: only its far end says where the path goes, so judging the
			// first hop alone would accept this one — "hop" points at an in-workdir
			// name, and that name is the link out.
			if err := os.Symlink("escape", filepath.Join(wd, "hop")); err != nil {
				t.Fatal(err)
			}
			if tc.symlinkedDir {
				alias := filepath.Join(base, "alias")
				if err := os.Symlink("real", alias); err != nil {
					t.Fatal(err)
				}
				wd = alias
			}

			if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "escape", Exists: ptrBool(false)}); cr.OK {
				t.Error("a dangling link out of the workdir must be refused, not answered")
			}
			if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "hop", Exists: ptrBool(false)}); cr.OK {
				t.Error("a dangling chain whose far end leaves the workdir must be refused too")
			}
			if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "latest", Exists: ptrBool(false)}); !cr.OK {
				t.Errorf("a dangling link inside the workdir must still answer: %+v", cr)
			}
		})
	}
}

// TestCheckDir_BrokenSymlinkMembership pins that contains/not_contains judge
// membership by the directory entry, not by whether the link target resolves. A
// dangling symlink is a real dirent, so not_contains must FAIL (the file was
// left behind) and contains must PASS. Using os.Stat here would follow the link,
// see IsNotExist, and wrongly report the entry absent — a dangerous false PASS
// for not_contains.
func TestCheckDir_BrokenSymlinkMembership(t *testing.T) {
	wd := t.TempDir()
	dir := filepath.Join(wd, "out")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	// A symlink whose target does not exist: the dirent "planted" is present, but
	// os.Stat on it returns IsNotExist.
	if err := os.Symlink(filepath.Join(dir, "nonexistent-target"), filepath.Join(dir, "planted")); err != nil {
		t.Skipf("symlinks unsupported on this platform: %v", err)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "out", NotContains: []string{"planted"}}); cr.OK {
		t.Error("not_contains must FAIL for a present (broken-symlink) entry")
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "out", Contains: []string{"planted"}}); !cr.OK {
		t.Errorf("contains must PASS for a present (broken-symlink) entry: %+v", cr)
	}
}

func TestCheckDir_NotADirectory(t *testing.T) {
	wd := makeTree(t)
	if err := os.WriteFile(filepath.Join(wd, "afile"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "afile", Count: ptrInt(0)}); cr.OK {
		t.Error("a regular file is not a directory; count constraint should fail")
	}
}

// TestCheckDir_FailureShowsListing pins the diagnostic contract for directory
// assertions: a failure must show what the directory actually holds, the way a
// file assertion shows the file's content and a changes assertion shows the
// observed delta. Reporting only "missing" or "3 entries" forces the reader to
// re-run the generator by hand to learn what it produced.
func TestCheckDir_FailureShowsListing(t *testing.T) {
	wd := makeTree(t)
	cases := []struct {
		name string
		dir  *spec.DirAssert
		want []string
	}{
		{
			name: "contains names the entries that are there",
			dir:  &spec.DirAssert{Path: "site", Contains: []string{"style.css"}},
			want: []string{"about.html", "assets/", "index.html"},
		},
		{
			name: "not_contains shows the listing too",
			dir:  &spec.DirAssert{Path: "site", NotContains: []string{"index.html"}},
			want: []string{"about.html", "assets/", "index.html"},
		},
		{
			name: "count shows which entries were counted",
			dir:  &spec.DirAssert{Path: "site", Count: ptrInt(9)},
			want: []string{"about.html", "assets/", "index.html"},
		},
		{
			name: "glob shows the entries it tried to match",
			dir:  &spec.DirAssert{Path: "site", Glob: "*.js"},
			want: []string{"about.html", "assets/", "index.html"},
		},
		{
			name: "recursive contains shows the walked tree",
			dir:  &spec.DirAssert{Path: "site", Recursive: true, Contains: []string{"nope.txt"}},
			want: []string{"assets/app.css", "index.html"},
		},
		{
			name: "recursive glob shows the walked tree",
			dir:  &spec.DirAssert{Path: "site", Recursive: true, Glob: "*.js"},
			want: []string{"assets/app.css", "index.html"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cr := checkDirOK(t, wd, tc.dir)
			if cr.OK {
				t.Fatalf("assertion should have failed: %+v", cr)
			}
			for _, want := range tc.want {
				if !strings.Contains(cr.Actual, want) {
					t.Errorf("Actual = %q, want it to list %q", cr.Actual, want)
				}
			}
		})
	}
}

// TestCheckDir_ListingIsBounded keeps a huge directory from flooding the
// failure block: the listing is capped and says how many entries it elided.
func TestCheckDir_ListingIsBounded(t *testing.T) {
	wd := t.TempDir()
	big := filepath.Join(wd, "big")
	if err := os.MkdirAll(big, 0o750); err != nil {
		t.Fatal(err)
	}
	for i := range 200 {
		if err := os.WriteFile(filepath.Join(big, fmt.Sprintf("f%03d.txt", i)), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cr := checkDirOK(t, wd, &spec.DirAssert{Path: "big", Contains: []string{"nope"}})
	if cr.OK {
		t.Fatal("assertion should have failed")
	}
	if n := strings.Count(cr.Actual, "\n"); n > dirListingLimit+3 {
		t.Errorf("Actual has %d lines, want the listing capped near %d", n, dirListingLimit)
	}
	if !strings.Contains(cr.Actual, "more") {
		t.Errorf("Actual = %q, want it to say how many entries were elided", cr.Actual)
	}
}
