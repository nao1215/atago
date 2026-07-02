package assert

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

func intp(i int) *int         { return &i }
func strp(s string) *string   { return &s }
func boolp(b bool) *bool      { return &b }
func f64p(f float64) *float64 { return &f }

func TestCheck_ExitCode(t *testing.T) {
	t.Parallel()
	res := &runner.Result{ExitCode: 2}
	tests := []struct {
		name   string
		a      *spec.Assert
		wantOK bool
	}{
		{"equals match", &spec.Assert{ExitCode: &spec.ExitCode{Equals: intp(2)}}, true},
		{"equals mismatch", &spec.Assert{ExitCode: &spec.ExitCode{Equals: intp(0)}}, false},
		{"not match", &spec.Assert{ExitCode: &spec.ExitCode{Not: intp(0)}}, true},
		{"not mismatch", &spec.Assert{ExitCode: &spec.ExitCode{Not: intp(2)}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Check(tt.a, res, Env{}); got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

func TestCheck_Stream(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("Alice and Bob\n"), Stderr: []byte("")}
	tests := []struct {
		name   string
		a      *spec.Assert
		wantOK bool
	}{
		{"contains hit", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice"}}}, true},
		{"contains miss", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Carol"}}}, false},
		{"not_contains hit", &spec.Assert{Stdout: &spec.StreamAssert{NotContains: spec.StringList{"Carol"}}}, true},
		{"not_contains miss", &spec.Assert{Stdout: &spec.StreamAssert{NotContains: spec.StringList{"Alice"}}}, false},
		{"not_contains on a line", &spec.Assert{Stdout: &spec.StreamAssert{Line: intp(1), NotContains: spec.StringList{"Carol"}}}, true},
		{"contains list all present", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice", "Bob"}}}, true},
		{"contains list one missing", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice", "Carol"}}}, false},
		{"not_contains list all absent", &spec.Assert{Stdout: &spec.StreamAssert{NotContains: spec.StringList{"Carol", "Dave"}}}, true},
		{"not_contains list one present", &spec.Assert{Stdout: &spec.StreamAssert{NotContains: spec.StringList{"Carol", "Bob"}}}, false},
		{"matches hit", &spec.Assert{Stdout: &spec.StreamAssert{Matches: strp("A.+e")}}, true},
		{"equals trailing-newline tolerant", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("Alice and Bob")}}, true},
		{"not_equals differs", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("someone else")}}, true},
		{"not_equals matches fails", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("Alice and Bob")}}, false},
		{"not_equals newline tolerant", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("Alice and Bob\n")}}, false},
		{"not_equals on a line", &spec.Assert{Stdout: &spec.StreamAssert{Line: intp(1), NotEquals: strp("nope")}}, true},
		{"stderr empty", &spec.Assert{Stderr: &spec.StreamAssert{Empty: boolp(true)}}, true},
		{"stdout not empty", &spec.Assert{Stdout: &spec.StreamAssert{Empty: boolp(false)}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Check(tt.a, res, Env{}); got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_Stream_CRLF proves `equals` treats CRLF output (cmd.exe on Windows)
// like LF output, so the same spec asserts the same behavior on every OS. Line
// endings are an OS artifact, not observable CLI behavior.
func TestCheck_Stream_CRLF(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("Alice and Bob\r\nsecond line\r\n")}
	tests := []struct {
		name   string
		a      *spec.Assert
		wantOK bool
	}{
		{"equals folds CRLF", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("Alice and Bob\nsecond line")}}, true},
		{"equals still exact per line", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("Alice and Bob\nsecond")}}, false},
		{"not_equals folds CRLF", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("Alice and Bob\nsecond line")}}, false},
		{"line selector strips CR", &spec.Assert{Stdout: &spec.StreamAssert{Line: intp(2), Equals: strp("second line")}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Check(tt.a, res, Env{}); got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_ContainsList_FailureNamesElement verifies an array contains /
// not_contains failure identifies which element failed, and that a single-element
// list keeps the original (no "element N of M") failure phrasing.
func TestCheck_ContainsList_FailureNamesElement(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("Alice and Bob\n")}

	// contains: the second element is missing → hint names element 2 of 2.
	cr := Check(&spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice", "Carol"}}}, res, Env{})
	if cr.OK {
		t.Fatal("expected contains list to fail")
	}
	if !strings.Contains(cr.Hint, `"Carol"`) || !strings.Contains(cr.Hint, "element 2 of 2") {
		t.Errorf("contains hint = %q, want it to name \"Carol\" and element 2 of 2", cr.Hint)
	}

	// not_contains: a present element → hint names it and marks it unexpected.
	cr = Check(&spec.Assert{Stdout: &spec.StreamAssert{NotContains: spec.StringList{"Carol", "Bob"}}}, res, Env{})
	if cr.OK {
		t.Fatal("expected not_contains list to fail")
	}
	if !strings.Contains(cr.Hint, `"Bob"`) || !strings.Contains(cr.Hint, "unexpectedly present") {
		t.Errorf("not_contains hint = %q, want it to name \"Bob\" as unexpectedly present", cr.Hint)
	}

	// single-element list keeps the pre-list phrasing (no "element N of M").
	cr = Check(&spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Carol"}}}, res, Env{})
	if cr.OK || strings.Contains(cr.Hint, "element") {
		t.Errorf("single-element hint = %q, want a plain failure with no element index", cr.Hint)
	}
}

// TestCheckAll_MultiTarget verifies one assert with several targets yields one
// result per target, in SetTargets order, and that AllOK aggregates them.
func TestCheckAll_MultiTarget(t *testing.T) {
	t.Parallel()
	res := &runner.Result{ExitCode: 0, Stdout: []byte("hello\n")}

	a := &spec.Assert{
		ExitCode: &spec.ExitCode{Equals: intp(0)},
		Stdout:   &spec.StreamAssert{Contains: spec.StringList{"hello"}},
	}
	got := CheckAll(a, res, Env{})
	if len(got) != 2 {
		t.Fatalf("CheckAll returned %d results, want 2", len(got))
	}
	if !AllOK(got) {
		t.Errorf("AllOK = false, want true (%+v)", got)
	}

	// One failing target makes AllOK false; the other still reports its own pass.
	a.Stdout = &spec.StreamAssert{Contains: spec.StringList{"nope"}}
	got = CheckAll(a, res, Env{})
	if len(got) != 2 {
		t.Fatalf("CheckAll returned %d results, want 2", len(got))
	}
	if AllOK(got) {
		t.Error("AllOK = true, want false when one target fails")
	}
	if !got[0].OK {
		t.Errorf("exit_code target should pass, got %+v", got[0])
	}
	if got[1].OK {
		t.Errorf("stdout target should fail, got %+v", got[1])
	}
}

func TestCheck_HTTP(t *testing.T) {
	t.Parallel()
	hdr := http.Header{}
	hdr.Set("Content-Type", "application/json; charset=utf-8")
	res := &runner.Result{IsHTTP: true, StatusCode: 200, Header: hdr, Body: []byte(`{"id":7}`)}
	tests := []struct {
		name   string
		a      *spec.Assert
		wantOK bool
	}{
		{"status hit", &spec.Assert{Status: intp(200)}, true},
		{"status miss", &spec.Assert{Status: intp(404)}, false},
		{"header equals hit", &spec.Assert{Header: &spec.HeaderMatch{Name: "Content-Type", Equals: strp("application/json; charset=utf-8")}}, true},
		{"header contains hit (case-insensitive name)", &spec.Assert{Header: &spec.HeaderMatch{Name: "content-type", Contains: strp("application/json")}}, true},
		{"header contains miss", &spec.Assert{Header: &spec.HeaderMatch{Name: "Content-Type", Contains: strp("text/html")}}, false},
		{"body json hit", &spec.Assert{Body: &spec.StreamAssert{JSON: &spec.JSONAssert{Path: "$.id", Equals: 7}}}, true},
		{"body contains hit", &spec.Assert{Body: &spec.StreamAssert{Contains: spec.StringList{`"id":7`}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Check(tt.a, res, Env{}); got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

func TestCheck_HTTP_NoResponse(t *testing.T) {
	t.Parallel()
	// A cmd result (IsHTTP=false) must not satisfy HTTP assertions.
	cmdRes := &runner.Result{Stdout: []byte("hi")}
	for _, a := range []*spec.Assert{
		{Status: intp(200)},
		{Header: &spec.HeaderMatch{Name: "X", Equals: strp("y")}},
	} {
		if got := Check(a, cmdRes, Env{}); got.OK {
			t.Errorf("HTTP assertion unexpectedly passed against a non-HTTP result: %+v", a)
		}
	}
}

// TestCheck_Line covers the 1-based line selector (spec.md §16.2), modeled on
// sqly's ShellSpec `The line N should equal '...'` output-format tests.
func TestCheck_Line(t *testing.T) {
	t.Parallel()
	// A pretty-printed JSON array and a CSV body, both with a trailing newline.
	jsonOut := "[\n  {\"id\":1},\n  {\"id\":2}\n]\n"
	csvOut := "user_name,identifier\nbooker12,1\n"
	tests := []struct {
		name   string
		out    string
		s      *spec.StreamAssert
		wantOK bool
	}{
		{"line 1 equals bracket", jsonOut, &spec.StreamAssert{Line: intp(1), Equals: strp("[")}, true},
		{"line 1 mismatch", jsonOut, &spec.StreamAssert{Line: intp(1), Equals: strp("{")}, false},
		{"last line equals close bracket", jsonOut, &spec.StreamAssert{Line: intp(4), Equals: strp("]")}, true},
		{"line contains", jsonOut, &spec.StreamAssert{Line: intp(2), Contains: spec.StringList{"\"id\":1"}}, true},
		{"line matches", jsonOut, &spec.StreamAssert{Line: intp(3), Matches: strp(`"id":2`)}, true},
		{"csv header line", csvOut, &spec.StreamAssert{Line: intp(1), Equals: strp("user_name,identifier")}, true},
		{"csv data line", csvOut, &spec.StreamAssert{Line: intp(2), Equals: strp("booker12,1")}, true},
		{"out of range fails", csvOut, &spec.StreamAssert{Line: intp(3), Equals: strp("x")}, false},
		{"trailing newline ignored", "only\n", &spec.StreamAssert{Line: intp(2), Equals: strp("")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: []byte(tt.out)}
			got := Check(&spec.Assert{Stdout: tt.s}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

func TestCheck_JSON(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte(`{"name":"Alice","items":[{"id":1},{"id":2}]}`)}
	tests := []struct {
		name   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		{"equals string", &spec.JSONAssert{Path: "$.name", Equals: "Alice"}, true},
		{"equals number", &spec.JSONAssert{Path: "$.items[0].id", Equals: 1}, true},
		{"equals number mismatch", &spec.JSONAssert{Path: "$.items[0].id", Equals: 9}, false},
		{"length", &spec.JSONAssert{Path: "$.items", Length: intp(2)}, true},
		{"matches", &spec.JSONAssert{Path: "$.name", Matches: strp("A.+")}, true},
		{"no match path", &spec.JSONAssert{Path: "$.missing", Equals: "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: tt.j}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_JSON_Compare covers the numeric bound matchers gt/gte/lt/lte, which
// let a spec pin a non-deterministic-but-bounded metric (a count, a coverage
// figure) without an exact equals. A non-numeric selected node must fail.
func TestCheck_JSON_Compare(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte(`{"count":3,"rate":0.5,"num_str":"7","name":"Alice"}`)}
	tests := []struct {
		name   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		{"gt hit", &spec.JSONAssert{Path: "$.count", Gt: f64p(2)}, true},
		{"gt equal is not gt", &spec.JSONAssert{Path: "$.count", Gt: f64p(3)}, false},
		{"gt miss", &spec.JSONAssert{Path: "$.count", Gt: f64p(5)}, false},
		{"gte equal", &spec.JSONAssert{Path: "$.count", Gte: f64p(3)}, true},
		{"gte miss", &spec.JSONAssert{Path: "$.count", Gte: f64p(4)}, false},
		{"lt hit", &spec.JSONAssert{Path: "$.count", Lt: f64p(4)}, true},
		{"lt equal is not lt", &spec.JSONAssert{Path: "$.count", Lt: f64p(3)}, false},
		{"lte equal", &spec.JSONAssert{Path: "$.count", Lte: f64p(3)}, true},
		{"float rate lt", &spec.JSONAssert{Path: "$.rate", Lt: f64p(1)}, true},
		{"numeric string compares", &spec.JSONAssert{Path: "$.num_str", Gte: f64p(7)}, true},
		{"non-numeric fails gt", &spec.JSONAssert{Path: "$.name", Gt: f64p(0)}, false},
		{"no match path fails", &spec.JSONAssert{Path: "$.missing", Gt: f64p(0)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: tt.j}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_JSON_EqualsStructural verifies json.equals compares nested
// objects/arrays structurally and is insensitive to map key ordering (#40),
// rather than relying on fmt.Sprintf("%v") of the decoded Go values.
func TestCheck_JSON_EqualsStructural(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte(`{"user":{"name":"Alice","age":30},"tags":["a","b"]}`)}
	tests := []struct {
		name   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		// Map key order differs from the JSON source; must still match.
		{"object key order", &spec.JSONAssert{Path: "$.user", Equals: map[string]any{"age": 30, "name": "Alice"}}, true},
		{"object value mismatch", &spec.JSONAssert{Path: "$.user", Equals: map[string]any{"name": "Bob", "age": 30}}, false},
		{"object extra key", &spec.JSONAssert{Path: "$.user", Equals: map[string]any{"name": "Alice"}}, false},
		{"array equal", &spec.JSONAssert{Path: "$.tags", Equals: []any{"a", "b"}}, true},
		{"array order matters", &spec.JSONAssert{Path: "$.tags", Equals: []any{"b", "a"}}, false},
		// Numeric normalization still holds inside nested structures.
		{"nested numeric normalization", &spec.JSONAssert{Path: "$.user", Equals: map[string]any{"name": "Alice", "age": 30.0}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: tt.j}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// Regression: empty stdout is not valid JSON. oj.Parse returns (nil, nil) for
// empty/whitespace input, which previously produced a misleading "no match"
// message. Found by dogfooding `gup list --json` (empty when no tools exist).
func TestCheck_JSON_EmptyStdout(t *testing.T) {
	t.Parallel()
	for _, data := range []string{"", "   ", "\n\t "} {
		res := &runner.Result{Stdout: []byte(data)}
		got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: &spec.JSONAssert{Path: "$.x", Equals: 1}}}, res, Env{})
		if got.OK {
			t.Errorf("empty stdout %q unexpectedly passed JSON assertion", data)
		}
		if !strings.Contains(got.Hint, "empty") {
			t.Errorf("hint = %q, want it to mention 'empty'", got.Hint)
		}
	}
}

// Issue #32: json length and stream empty matchers were only tested on their
// passing direction. Cover the failure and value-type branches.
func TestCheck_JSON_Length(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte(`{"arr":[1,2,3],"str":"hello","obj":{"a":1,"b":2},"num":42}`)}
	tests := []struct {
		name     string
		j        *spec.JSONAssert
		wantOK   bool
		wantHint string
	}{
		{"array length match", &spec.JSONAssert{Path: "$.arr", Length: intp(3)}, true, ""},
		{"array length mismatch", &spec.JSONAssert{Path: "$.arr", Length: intp(2)}, false, "length at"},
		{"string length match", &spec.JSONAssert{Path: "$.str", Length: intp(5)}, true, ""},
		{"string length mismatch", &spec.JSONAssert{Path: "$.str", Length: intp(4)}, false, "length at"},
		{"object length match", &spec.JSONAssert{Path: "$.obj", Length: intp(2)}, true, ""},
		{"object length mismatch", &spec.JSONAssert{Path: "$.obj", Length: intp(1)}, false, "length at"},
		{"number has no length", &spec.JSONAssert{Path: "$.num", Length: intp(2)}, false, "no length"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: tt.j}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Fatalf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
			if tt.wantHint != "" && !strings.Contains(got.Hint, tt.wantHint) {
				t.Errorf("hint = %q, want it to contain %q", got.Hint, tt.wantHint)
			}
		})
	}
}

func TestCheck_Stream_Empty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		data   []byte
		empty  bool
		wantOK bool
	}{
		{"empty:true on empty passes", []byte("  \n "), true, true},
		{"empty:true on non-empty fails", []byte("output"), true, false},
		{"empty:false on non-empty passes", []byte("output"), false, true},
		{"empty:false on empty fails", []byte("   "), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: tt.data}
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{Empty: boolp(tt.empty)}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// Regression for issue #9: the yaml stream matcher was schema-valid but
// unimplemented at runtime (fell through to "matcher not supported yet").
func TestCheck_YAML(t *testing.T) {
	t.Parallel()
	doc := "name: alice\nitems:\n  - id: 1\n  - id: 2\n"
	res := &runner.Result{Stdout: []byte(doc)}
	tests := []struct {
		name   string
		y      *spec.JSONAssert
		wantOK bool
	}{
		{"equals string", &spec.JSONAssert{Path: "$.name", Equals: "alice"}, true},
		{"equals number", &spec.JSONAssert{Path: "$.items[0].id", Equals: 1}, true},
		{"equals mismatch", &spec.JSONAssert{Path: "$.name", Equals: "bob"}, false},
		{"length", &spec.JSONAssert{Path: "$.items", Length: intp(2)}, true},
		{"matches", &spec.JSONAssert{Path: "$.name", Matches: strp("a.+")}, true},
		{"no match path", &spec.JSONAssert{Path: "$.missing", Equals: "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{YAML: tt.y}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
			if !tt.wantOK && strings.Contains(got.Hint, "not supported") {
				t.Errorf("yaml matcher reported as unsupported: %q", got.Hint)
			}
		})
	}
}

func TestCheck_YAML_Empty(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("   ")}
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{YAML: &spec.JSONAssert{Path: "$.x", Equals: 1}}}, res, Env{})
	if got.OK {
		t.Fatal("empty stdout unexpectedly passed YAML assertion")
	}
	if !strings.Contains(got.Hint, "empty") {
		t.Errorf("hint = %q, want it to mention 'empty'", got.Hint)
	}
}

func TestCheck_File(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "out.json"), []byte(`{"name":"Alice"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		f      *spec.FileAssert
		wantOK bool
	}{
		{"exists true", &spec.FileAssert{Path: "out.json", Exists: boolp(true)}, true},
		{"missing exists false", &spec.FileAssert{Path: "nope.txt", Exists: boolp(false)}, true},
		{"missing exists true fails", &spec.FileAssert{Path: "nope.txt", Exists: boolp(true)}, false},
		{"contains", &spec.FileAssert{Path: "out.json", Contains: spec.StringList{"Alice"}}, true},
		{"contains list all present", &spec.FileAssert{Path: "out.json", Contains: spec.StringList{"name", "Alice"}}, true},
		{"contains list one missing", &spec.FileAssert{Path: "out.json", Contains: spec.StringList{"Alice", "Carol"}}, false},
		{"not_contains hit", &spec.FileAssert{Path: "out.json", NotContains: spec.StringList{"Carol"}}, true},
		{"not_contains list all absent", &spec.FileAssert{Path: "out.json", NotContains: spec.StringList{"Carol", "Dave"}}, true},
		{"not_contains list one present", &spec.FileAssert{Path: "out.json", NotContains: spec.StringList{"Carol", "Alice"}}, false},
		{"not_contains miss", &spec.FileAssert{Path: "out.json", NotContains: spec.StringList{"Alice"}}, false},
		{"executable false for a 0600 file", &spec.FileAssert{Path: "out.json", Executable: boolp(false)}, true},
		{"executable true mismatch", &spec.FileAssert{Path: "out.json", Executable: boolp(true)}, false},
		{"json", &spec.FileAssert{Path: "out.json", JSON: &spec.JSONAssert{Path: "$.name", Equals: "Alice"}}, true},
		// A workdir-relative path must not escape the scenario workdir.
		{"parent traversal rejected", &spec.FileAssert{Path: "../escape.txt", Exists: boolp(true)}, false},
		{"deep traversal rejected", &spec.FileAssert{Path: "sub/../../escape.txt", Contains: spec.StringList{"x"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{File: tt.f}, nil, Env{Workdir: dir})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_File_TraversalRejected proves a file assertion whose path escapes the
// scenario workdir fails with a containment error naming the field, even when the
// escaping target actually exists on disk.
func TestCheck_File_TraversalRejected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workdir := filepath.Join(root, "scn")
	if err := os.Mkdir(workdir, 0o750); err != nil {
		t.Fatal(err)
	}
	// A real file one level above the workdir: containment, not existence, must
	// drive the rejection.
	if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("top"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := Check(&spec.Assert{File: &spec.FileAssert{Path: "../secret.txt", Contains: spec.StringList{"top"}}}, nil, Env{Workdir: workdir})
	if got.OK {
		t.Fatalf("path escaping the workdir was accepted")
	}
	if !strings.Contains(got.Hint, "escapes the scenario workdir") {
		t.Errorf("hint %q should explain the containment failure", got.Hint)
	}
}

// TestCheck_Snapshot_TraversalRejected proves a relative snapshot path may not
// escape the spec directory.
func TestCheck_Snapshot_TraversalRejected(t *testing.T) {
	t.Parallel()
	specDir := t.TempDir()
	res := &runner.Result{Stdout: []byte("hello\n")}
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{Snapshot: "../escape.snap"}}, res, Env{SpecDir: specDir, Workdir: specDir})
	if got.OK {
		t.Fatalf("snapshot path escaping the spec directory was accepted")
	}
	if !strings.Contains(got.Hint, "escapes the spec directory") {
		t.Errorf("hint %q should explain the containment failure", got.Hint)
	}
}

// TestCheck_File_ExistsUnreadable verifies that a stat error other than
// "not exist" (issue #39) is surfaced as an error-style failure rather than
// being collapsed into exists=false.
func TestCheck_File_ExistsUnreadable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows ignores POSIX chmod bits, so a path cannot be made unreadable this way")
	}
	if os.Getuid() == 0 {
		t.Skip("root bypasses directory permission bits")
	}
	dir := t.TempDir()
	// A file inside a directory with no execute (search) permission cannot be
	// stat-ed: os.Stat returns a permission error, not os.ErrNotExist.
	sub := filepath.Join(dir, "locked")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sub, "secret.txt")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sub, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(sub, 0o700) })

	// exists:false must NOT pass just because the path is unreadable.
	got := Check(&spec.Assert{File: &spec.FileAssert{Path: target, Exists: boolp(false)}}, nil, Env{Workdir: dir})
	if got.OK {
		t.Fatalf("unreadable path reported as absent; want error-style failure. hint=%s", got.Hint)
	}
	// exists:true must also fail with the real error, not a plain mismatch.
	got = Check(&spec.Assert{File: &spec.FileAssert{Path: target, Exists: boolp(true)}}, nil, Env{Workdir: dir})
	if got.OK {
		t.Fatalf("unreadable path with exists:true should fail; hint=%s", got.Hint)
	}
}
