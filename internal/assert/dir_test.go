package assert

import (
	"os"
	"path/filepath"
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

func TestCheckDir_NotADirectory(t *testing.T) {
	wd := makeTree(t)
	if err := os.WriteFile(filepath.Join(wd, "afile"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "afile", Count: ptrInt(0)}); cr.OK {
		t.Error("a regular file is not a directory; count constraint should fail")
	}
}
