package diag

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCode_String(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		code Code
		want string
	}{
		"four digits":       {code: Code(2201), want: "ATG2201"},
		"trailing zeros":    {code: Code(2010), want: "ATG2010"},
		"highest family":    {code: Code(6999), want: "ATG6999"},
		"padded to four":    {code: Code(2001), want: "ATG2001"},
		"exit-4 family":     {code: Code(4100), want: "ATG4100"},
		"unregistered code": {code: Code(2999), want: "ATG2999"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := tt.code.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCode_ExitCode(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		code Code
		want int
	}{
		"spec error":    {code: SpecEmpty, want: 2},
		"config error":  {code: Code(3001), want: 3},
		"exec error":    {code: Code(4999), want: 4},
		"internal":      {code: Code(5000), want: 5},
		"security":      {code: Code(6500), want: 6},
		"top of family": {code: Code(2999), want: 2},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := tt.code.ExitCode(); got != tt.want {
				t.Errorf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCode_Annotate(t *testing.T) {
	t.Parallel()
	got := RequiredKey.Annotate(`scenarios[0].name is required`)
	want := `ATG2201: scenarios[0].name is required`
	if got != want {
		t.Errorf("Annotate() = %q, want %q", got, want)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		in     string
		want   Code
		wantOK bool
	}{
		"registered code":     {in: "ATG2201", want: RequiredKey, wantOK: true},
		"surrounding spaces":  {in: "  ATG2201 ", want: RequiredKey, wantOK: true},
		"lowercase prefix":    {in: "atg2201", want: RequiredKey, wantOK: true},
		"mixed case prefix":   {in: "Atg2201", want: RequiredKey, wantOK: true},
		"well-formed unknown": {in: "ATG2999", wantOK: false},
		"unassigned family":   {in: "ATG1000", wantOK: false},
		"no prefix":           {in: "2201", wantOK: false},
		"not a number":        {in: "ATGabcd", wantOK: false},
		"empty":               {in: "", wantOK: false},
		"prefix only":         {in: "ATG", wantOK: false},
		"shorter than prefix": {in: "AT", wantOK: false},
		"trailing text":       {in: "ATG2201x", wantOK: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := Parse(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("Parse(%q) ok = %v, want %v", tt.in, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCodes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		in   string
		want []Code
	}{
		"one code in a sentence": {
			in:   "spec.yaml: ATG2201: suite.name is required",
			want: []Code{RequiredKey},
		},
		"several, in order of appearance": {
			in:   "ATG2501 and then ATG2201",
			want: []Code{DuplicateName, RequiredKey},
		},
		"repeats collapse": {
			in:   "ATG2201 ATG2201 ATG2201",
			want: []Code{RequiredKey},
		},
		"unregistered numbers are not codes": {
			in:   "ATG2999 ATG2201",
			want: []Code{RequiredKey},
		},
		"a longer number is not a code": {
			in:   "ATG22015",
			want: nil,
		},
		"no codes": {
			in:   "everything passed",
			want: nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := Codes(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("Codes() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Codes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()
	e, ok := Lookup(RequiredKey)
	if !ok {
		t.Fatal("Lookup(RequiredKey) reported the code as unregistered")
	}
	if e.Code != RequiredKey || e.Name != "RequiredKey" {
		t.Errorf("Lookup() = %+v, want the RequiredKey entry", e)
	}
	if _, ok := Lookup(Code(2999)); ok {
		t.Error("Lookup(ATG2999) found an entry for an unassigned code")
	}
}

// TestAll_Ordered checks the reference is rendered in code order regardless of
// the order the entries happen to be declared in.
func TestAll_Ordered(t *testing.T) {
	t.Parallel()
	all := All()
	if len(all) == 0 {
		t.Fatal("no diagnostics are registered")
	}
	for i := 1; i < len(all); i++ {
		if all[i-1].Code >= all[i].Code {
			t.Fatalf("All() is not ascending at %d: %s then %s", i, all[i-1].Code, all[i].Code)
		}
	}
}

// TestAll_EntriesAreComplete restates at the registry level what register
// enforces per call, so a future entry added by some other path cannot land
// half-documented.
func TestAll_EntriesAreComplete(t *testing.T) {
	t.Parallel()
	for _, e := range All() {
		if e.Name == "" || e.Summary == "" || e.Detail == "" || e.Fix == "" || e.Since == "" {
			t.Errorf("%s (%s) is missing published text: %+v", e.Code, e.Name, e)
		}
		if fam := e.Code.ExitCode(); fam < 2 || fam > 6 {
			t.Errorf("%s is in family %d, outside the documented ATG2xxx-ATG6xxx range", e.Code, fam)
		}
	}
}

// TestRegister_RejectsBadRegistrations pins the guarantees register makes. They
// are what lets the rest of the codebase treat "a code exists" and "a code is
// documented" as the same statement.
func TestRegister_RejectsBadRegistrations(t *testing.T) {
	t.Parallel()
	full := Entry{Summary: "s", Detail: "d", Fix: "f", Since: "v0.21.0"}
	tests := map[string]struct {
		number int
		name   string
		entry  Entry
	}{
		"below the range":    {number: 1999, name: "X", entry: full},
		"above the range":    {number: 7000, name: "X", entry: full},
		"reserved family":    {number: 1200, name: "X", entry: full},
		"no name":            {number: 2900, name: "", entry: full},
		"no summary":         {number: 2900, name: "X", entry: Entry{Detail: "d", Fix: "f", Since: "v"}},
		"no detail":          {number: 2900, name: "X", entry: Entry{Summary: "s", Fix: "f", Since: "v"}},
		"no fix":             {number: 2900, name: "X", entry: Entry{Summary: "s", Detail: "d", Since: "v"}},
		"no since":           {number: 2900, name: "X", entry: Entry{Summary: "s", Detail: "d", Fix: "f"}},
		"already registered": {number: int(RequiredKey), name: "X", entry: full},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recover() == nil {
					t.Errorf("register(%d, %q, ...) was accepted, want a panic", tt.number, tt.name)
				}
			}()
			register(tt.number, tt.name, tt.entry)
		})
	}
}

// TestRegistry_NamesMatchIdentifiers is the guard against the one duplication
// in this package: register takes the identifier's name as a string so the
// registry can report it, and nothing but this test stops the two from
// drifting. It also proves every code in the package is registered rather than
// constructed some other way, by requiring each Code-typed package variable to
// come from a register call.
func TestRegistry_NamesMatchIdentifiers(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read the diag package directory: %v", err)
	}
	fset := token.NewFileSet()
	declared := map[string]int{}
	for _, de := range entries {
		path := de.Name()
		if de.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vs.Names {
					if i >= len(vs.Values) {
						continue
					}
					call, ok := vs.Values[i].(*ast.CallExpr)
					if !ok {
						continue
					}
					fn, ok := call.Fun.(*ast.Ident)
					if !ok || fn.Name != "register" {
						continue
					}
					if len(call.Args) < 2 {
						t.Fatalf("%s: register(%s) takes a number and a name", filepath.Base(path), name.Name)
					}
					lit, ok := call.Args[1].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						t.Fatalf("%s: register's name argument for %s is not a string literal", filepath.Base(path), name.Name)
					}
					got, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatalf("%s: unquote the name argument for %s: %v", filepath.Base(path), name.Name, err)
					}
					if got != name.Name {
						t.Errorf("%s: %s is registered under the name %q; the identifier and the registered name must match, or the reference names something that does not exist", filepath.Base(path), name.Name, got)
					}
					num, ok := call.Args[0].(*ast.BasicLit)
					if !ok || num.Kind != token.INT {
						t.Fatalf("%s: register's code argument for %s is not a literal number", filepath.Base(path), name.Name)
					}
					n, err := strconv.Atoi(num.Value)
					if err != nil {
						t.Fatalf("%s: parse the code argument for %s: %v", filepath.Base(path), name.Name, err)
					}
					declared[name.Name] = n
				}
			}
		}
	}

	if len(declared) != len(All()) {
		t.Errorf("the package declares %d register calls but %d diagnostics are registered", len(declared), len(All()))
	}
	for _, e := range All() {
		n, ok := declared[e.Name]
		if !ok {
			t.Errorf("%s is registered as %q, which is not a package variable in diag", e.Code, e.Name)
			continue
		}
		if Code(n) != e.Code {
			t.Errorf("%s is bound to the identifier %s, which registers %d", e.Code, e.Name, n)
		}
	}
}
