package loader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// regexpSurface records what a spec field carrying a user-supplied regexp is
// validated for at load time, keyed by the function that compiles it.
//
// The empty-matching check ("q*" where "q+" was meant) was written once, for
// the field that motivated it, and nobody had a list of the others: the same
// pattern silently produced nonsense under a count bound, in a scrub rule, and
// in a store capture. This table is that list, and the walk below fails when a
// new place compiles a spec-supplied pattern without being added to it.
type regexpSurface struct {
	// field names the spec key the pattern comes from.
	field string
	// rejectsEmptyMatching says a pattern matching the empty string is refused
	// at load. When set, refuse must load and be rejected.
	rejectsEmptyMatching bool
	// refuse is a spec whose only fault is an empty-matching pattern in this
	// field. Required when rejectsEmptyMatching is set.
	refuse string
	// reason explains why an empty-matching pattern is allowed here.
	reason string
}

// regexpSpec builds a minimal spec around one step.
func regexpSpec(step string) string {
	return "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    steps:\n      - " + step + "\n"
}

// regexpSurfaces maps "package.function" — every function in internal/ that
// compiles a pattern coming from a spec — to the decision for that field.
var regexpSurfaces = map[string]regexpSurface{
	"loader.validateRegexp": {
		field:                "assert stdout/stderr/body/rows/message/value/screen matches and not_matches",
		rejectsEmptyMatching: true,
		refuse:               regexpSpec("run: {command: echo}") + "      - assert: {stdout: {not_matches: \"z*\"}}\n",
	},
	"loader.matchesEmpty": {
		field:  "the shared predicate itself",
		reason: "it answers the question rather than asking it",
	},
	"loader.validateStoreSelector": {
		field:                "store from.<source>.matches",
		rejectsEmptyMatching: true,
		refuse:               regexpSpec("run: {command: echo}") + "      - store: {name: v, from: {stdout: {matches: \"[0-9]*\"}}}\n",
	},
	"loader.validatePTY": {
		field: "pty session expect",
		reason: "an expect waits for a pattern to appear in the transcript; one that matches the empty string " +
			"is satisfied immediately, which is a slow-motion no-op rather than an assertion that can never pass, " +
			"and the session budget still bounds it",
	},
	"loader.validateReady": {
		field:  "service ready.log",
		reason: "a readiness probe that matches the empty string is satisfied by the first byte of output, which is a weak probe rather than an impossible one",
	},
	"assert.streamMatches": {
		field:  "assert matches at check time",
		reason: "the loader already refused what cannot be satisfied; this compiles the surviving pattern",
	},
	"assert.streamNotMatches": {
		field:  "assert not_matches at check time",
		reason: "validated at load by validateRegexp",
	},
	"assert.countPattern": {
		field:  "assert matches under a count bound at check time",
		reason: "validated at load by validateStreamCount, which refuses an empty-matching pattern there",
	},
	"assert.checkHeaderValue": {
		field:  "assert header matches at check time",
		reason: "validated at load by validateRegexp",
	},
	"assert.jsonMatches": {
		field:  "assert json/yaml matches at check time",
		reason: "validated at load by validateRegexp",
	},
	"engine.regexValue": {
		field:  "store capture at run time",
		reason: "validated at load by validateStoreSelector",
	},
	"scrub.New": {
		field:                "top-level scrub rule patterns",
		rejectsEmptyMatching: true,
		refuse:               "version: \"1\"\nsuite: {name: s}\nscrub:\n  - {pattern: \"[0-9]*\", placeholder: \"<ID>\"}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n",
	},
	"service.waitReady": {
		field:  "service ready.log at run time",
		reason: "validated at load by validateReady",
	},
	"ptyrun.compileSession": {
		field:  "pty expect at run time",
		reason: "validated at load by validatePTY",
	},
}

// TestRegexpSurfaces_EveryCompileSiteIsDecided walks internal/ for calls that
// compile a pattern the caller did not write literally — that is, one a spec
// author supplied — and requires each to appear in regexpSurfaces. A new
// regexp-taking field therefore cannot ship without someone deciding which
// validations it gets, which is the gap this class kept falling into.
func TestRegexpSurfaces_EveryCompileSiteIsDecided(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "internal")
	fset := token.NewFileSet()
	found := map[string]bool{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		pkg := file.Name.Name
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			ast.Inspect(fn.Body, func(inner ast.Node) bool {
				call, ok := inner.(*ast.CallExpr)
				if !ok || !isRegexpCompile(call) || len(call.Args) != 1 {
					return true
				}
				if _, literal := call.Args[0].(*ast.BasicLit); literal {
					// atago's own pattern, fixed in the source: not a spec field.
					return true
				}
				found[pkg+"."+fn.Name.Name] = true
				return true
			})
			return false
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/: %v", err)
	}
	if len(found) == 0 {
		t.Fatal("found no regexp compile sites; the walk is not looking where it thinks it is")
	}

	for site := range found {
		surface, ok := regexpSurfaces[site]
		switch {
		case !ok:
			t.Errorf("%s compiles a spec-supplied regexp and is not in regexpSurfaces; "+
				"decide whether an empty-matching pattern is refused for that field, and record it", site)
		case surface.rejectsEmptyMatching && surface.refuse == "":
			t.Errorf("%s claims to refuse empty-matching patterns without a spec proving it", site)
		case !surface.rejectsEmptyMatching && surface.reason == "":
			t.Errorf("%s accepts empty-matching patterns without a stated reason", site)
		}
	}
	for site := range regexpSurfaces {
		if !found[site] {
			t.Errorf("regexpSurfaces lists %s, which no longer compiles a spec-supplied regexp", site)
		}
	}
}

// TestRegexpSurfaces_RefusalsHold loads each refusal spec the table declares, so
// a claimed refusal cannot rot into a comment.
func TestRegexpSurfaces_RefusalsHold(t *testing.T) {
	t.Parallel()
	for site, surface := range regexpSurfaces {
		if !surface.rejectsEmptyMatching {
			continue
		}
		t.Run(site, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(surface.refuse))
			if err == nil {
				t.Fatalf("%s (%s): LoadBytes() error = nil, want the empty-matching pattern refused", site, surface.field)
			}
			if !strings.Contains(err.Error(), "matches the empty string") {
				t.Errorf("%s: error = %q, want it to name the empty-string match", site, err)
			}
		})
	}
}

// isRegexpCompile reports whether the call is regexp.Compile or
// regexp.MustCompile.
func isRegexpCompile(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "regexp" && (sel.Sel.Name == "Compile" || sel.Sel.Name == "MustCompile")
}
