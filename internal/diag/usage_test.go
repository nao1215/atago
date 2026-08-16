package diag

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// familyOf maps a package to the exit code the diagnostics it raises must
// belong to. A package that raises coded errors has to appear here, so adding
// one is a deliberate statement about which family it reports in rather than
// something that happens by accident.
//
// The point is the thousands digit: ATG2103 promises the process exits 2, and
// that promise is only true if the package raising it exits 2. A loader
// reporting an ATG4xxx would make the code lie about what the reader is
// looking at.
var familyOf = map[string]int{
	"internal/loader": 2,
}

// TestCodes_AreReferenced is the guard against a code that exists only in the
// reference. Every registered diagnostic must be raised from somewhere in
// non-test code: publishing a code atago cannot produce sends the reader
// looking for a cause that is not there, and leaves dead entries behind when a
// check is deleted.
func TestCodes_AreReferenced(t *testing.T) {
	t.Parallel()
	used := usedCodes(t)
	for _, e := range All() {
		if _, ok := used[e.Name]; !ok {
			t.Errorf("%s (diag.%s) is registered and documented but never raised; either use it or remove its entry", e.Code, e.Name)
		}
	}
}

// TestCodes_MatchTheirPackageFamily checks the thousands digit against the
// exit code the raising package produces, so a code cannot promise one exit
// status while its caller returns another.
func TestCodes_MatchTheirPackageFamily(t *testing.T) {
	t.Parallel()
	byName := map[string]Entry{}
	for _, e := range All() {
		byName[e.Name] = e
	}
	for name, pkgs := range usedCodes(t) {
		e := byName[name]
		for _, pkg := range pkgs {
			want, ok := familyOf[pkg]
			if !ok {
				t.Errorf("%s raises diagnostics but is not listed in familyOf; add it with the exit code it reports", pkg)
				continue
			}
			if got := e.Code.ExitCode(); got != want {
				t.Errorf("%s raises %s (diag.%s), which is family %d, but %s reports exit %d", pkg, e.Code, e.Name, got, pkg, want)
			}
		}
	}
}

// usedCodes reports every registered diagnostic referenced from non-test Go
// files outside this package, mapped to the slash-separated package
// directories that reference it. References to the package's other exported
// names — the Code type, Lookup, All — are not codes and are skipped.
func usedCodes(t *testing.T) map[string][]string {
	t.Helper()
	root := repoRoot(t)
	registered := map[string]bool{}
	for _, e := range All() {
		registered[e.Name] = true
	}
	used := map[string][]string{}
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "dist", "node_modules", "public", "testdata", "diag":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		if !importsDiag(file) {
			return nil
		}
		rel, rerr := filepath.Rel(root, filepath.Dir(path))
		if rerr != nil {
			return rerr
		}
		pkg := filepath.ToSlash(rel)
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "diag" {
				return true
			}
			name := sel.Sel.Name
			if !registered[name] {
				return true
			}
			if !slices.Contains(used[name], pkg) {
				used[name] = append(used[name], pkg)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk the repository: %v", err)
	}
	return used
}

// importsDiag reports whether the file imports this package, so a local
// variable that happens to be called diag in an unrelated file cannot be read
// as a reference to it.
func importsDiag(file *ast.File) bool {
	for _, imp := range file.Imports {
		if imp.Path != nil && strings.HasSuffix(strings.Trim(imp.Path.Value, `"`), "/internal/diag") {
			return true
		}
	}
	return false
}

// repoRoot walks up from this package to the directory holding go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the diag package")
		}
		dir = parent
	}
}
