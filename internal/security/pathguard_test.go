package security_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rootArgNames are the identifiers that name a confinement root. A path built by
// joining one of these to an author-written value has to go through
// security.ResolveWorkdirPath / ResolveSpecPath instead, or the resulting path is
// whatever the spec asked for — including `../` out of the tree, or through a
// symlinked ancestor the program under test planted.
var rootArgNames = map[string]bool{
	"workdir": true,
	"specDir": true,
}

// allowedJoins lists the joins that legitimately bypass the resolver, by
// "<package dir>:<function>". Each entry is a claim that the second argument is
// NOT author-written, and it has to stay small: this list is the whole reason a
// reviewer can trust the rule.
var allowedJoins = map[string]string{
	"internal/runner/cmd:EnsureSandboxHome": "joins a package constant (SandboxHomeDirName), not spec input",
	"internal/runner/cmd:resolveDir":        "run.cwd is deliberately unconfined -- an absolute cwd is a documented capability, so confining the relative form would only be inconsistent",
	"internal/fixture:Write":                "fixture.from READS the author's own tree, which is the point of fixtures_dir pointing above the spec directory; absolute sources are documented",
}

// TestWorkdirJoinsGoThroughTheResolver is the structural half of the confinement
// rule. Every path-taking feature -- assert.file, fixtures, redirects, snapshots,
// screenshots, changes -- resolves author-written paths through one helper, and
// the reason is that they each grew their own join once and each inherited the
// same hole: an ancestor component that is a symlink out of the tree (#430).
//
// forbidigo cannot express this, because its patterns match the function name
// and never the arguments. So the rule is checked here, where the arguments are
// visible: a filepath.Join whose first argument is a confinement root is a
// finding unless it is listed above with a reason.
func TestWorkdirJoinsGoThroughTheResolver(t *testing.T) {
	t.Parallel()
	root := repoRootDir(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			// .claude can hold git worktrees checked out inside the repo, which
			// would scan a second copy of every package.
			case ".git", ".claude", "dist", "node_modules", "public", "testdata", "website", "site":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		rel, rerr := filepath.Rel(root, filepath.Dir(path))
		if rerr != nil {
			return rerr
		}
		pkgDir := filepath.ToSlash(rel)

		var fn string
		ast.Inspect(f, func(n ast.Node) bool {
			if decl, ok := n.(*ast.FuncDecl); ok {
				fn = decl.Name.Name
			}
			call, ok := n.(*ast.CallExpr)
			if !ok || !isFilepathJoin(call.Fun) || len(call.Args) < 2 {
				return true
			}
			first, ok := call.Args[0].(*ast.Ident)
			if !ok || !rootArgNames[first.Name] {
				return true
			}
			key := pkgDir + ":" + fn
			if _, allowed := allowedJoins[key]; allowed {
				return true
			}
			pos := fset.Position(call.Pos())
			t.Errorf("%s:%d: %s joins %q directly.\n"+
				"Author-written paths must resolve through security.ResolveWorkdirPath or ResolveSpecPath, "+
				"which reject `../` and an ancestor symlinked out of the tree.\n"+
				"If this second argument is not author-written, add %q to allowedJoins with the reason.",
				filepath.ToSlash(pos.Filename), pos.Line, fn, first.Name, key)
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// TestAllowedJoinsAreAllStillReal keeps the escape list honest: an entry whose
// join has been rewritten or deleted must not linger, or the list slowly stops
// describing the code and starts excusing whatever lands on the same name.
func TestAllowedJoinsAreAllStillReal(t *testing.T) {
	t.Parallel()
	root := repoRootDir(t)
	fset := token.NewFileSet()
	seen := map[string]bool{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "dist", "node_modules", "public", "testdata", "website", "site":
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		rel, rerr := filepath.Rel(root, filepath.Dir(path))
		if rerr != nil {
			return rerr
		}
		pkgDir := filepath.ToSlash(rel)
		var fn string
		ast.Inspect(f, func(n ast.Node) bool {
			if decl, ok := n.(*ast.FuncDecl); ok {
				fn = decl.Name.Name
			}
			call, ok := n.(*ast.CallExpr)
			if !ok || !isFilepathJoin(call.Fun) || len(call.Args) < 2 {
				return true
			}
			if first, ok := call.Args[0].(*ast.Ident); ok && rootArgNames[first.Name] {
				seen[pkgDir+":"+fn] = true
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	for key, why := range allowedJoins {
		if !seen[key] {
			t.Errorf("allowedJoins lists %q (%s) but no such join exists any more; remove the entry", key, why)
		}
	}
}

// isFilepathJoin reports whether fun is the filepath.Join selector.
func isFilepathJoin(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Join" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "filepath"
}

// repoRootDir walks up to the directory holding go.mod.
func repoRootDir(t *testing.T) string {
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
			t.Fatal("go.mod not found above the security package")
		}
		dir = parent
	}
}
