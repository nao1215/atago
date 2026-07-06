package assert

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/runner/mock"
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

// TestCheck_ExitCode_TimedOut proves a timed-out command's failure says so
// instead of presenting the synthetic -1 as a normal exit code.
func TestCheck_ExitCode_TimedOut(t *testing.T) {
	t.Parallel()
	res := &runner.Result{ExitCode: -1, TimedOut: true, Duration: 200 * time.Millisecond}
	got := Check(&spec.Assert{ExitCode: &spec.ExitCode{Equals: intp(0)}}, res, Env{})
	if got.OK {
		t.Fatal("OK = true, want a failure for a timed-out command")
	}
	if !strings.Contains(got.Actual, "timed out after 200ms") {
		t.Errorf("Actual = %q, want it to mention the timeout", got.Actual)
	}
	if !strings.Contains(got.Hint, "run.timeout") {
		t.Errorf("Hint = %q, want it to name run.timeout", got.Hint)
	}
}

// TestCheck_ExitCode_In proves membership in the accepted set (#19): a listed
// code passes, an unlisted one fails with the set spelled out.
func TestCheck_ExitCode_In(t *testing.T) {
	t.Parallel()
	in := &spec.ExitCode{In: []int{0, 2}}
	if got := Check(&spec.Assert{ExitCode: in}, &runner.Result{ExitCode: 2}, Env{}); !got.OK {
		t.Errorf("exit code 2 against in [0, 2]: OK = false, want pass: %+v", got)
	}
	got := Check(&spec.Assert{ExitCode: in}, &runner.Result{ExitCode: 1}, Env{})
	if got.OK {
		t.Fatal("exit code 1 against in [0, 2]: OK = true, want failure")
	}
	if !strings.Contains(got.Expected, "exit code in [0, 2]") {
		t.Errorf("Expected = %q, want it to list the accepted set", got.Expected)
	}
	if !strings.Contains(got.Hint, "[0, 2]") {
		t.Errorf("Hint = %q, want it to list the accepted set", got.Hint)
	}
}

// TestCheck_ExitCode_InTimedOut proves a timeout kill keeps the timeout hint
// under the in matcher instead of presenting the synthetic -1 as a plain
// mismatch (#19).
func TestCheck_ExitCode_InTimedOut(t *testing.T) {
	t.Parallel()
	res := &runner.Result{ExitCode: -1, TimedOut: true, Duration: 200 * time.Millisecond}
	got := Check(&spec.Assert{ExitCode: &spec.ExitCode{In: []int{0, 2}}}, res, Env{})
	if got.OK {
		t.Fatal("OK = true, want a failure for a timed-out command")
	}
	if !strings.Contains(got.Actual, "timed out after 200ms") {
		t.Errorf("Actual = %q, want it to mention the timeout", got.Actual)
	}
	if !strings.Contains(got.Hint, "run.timeout") {
		t.Errorf("Hint = %q, want the timeout hint, not a bare mismatch", got.Hint)
	}
}

// TestCheck_ExitCode_TimedOutNamesSource proves the hint names the level that
// supplied the timeout when the engine's resolver recorded one (#17).
func TestCheck_ExitCode_TimedOutNamesSource(t *testing.T) {
	t.Parallel()
	res := &runner.Result{ExitCode: -1, TimedOut: true, TimeoutSource: "suite.timeout", Duration: 200 * time.Millisecond}
	got := Check(&spec.Assert{ExitCode: &spec.ExitCode{Equals: intp(0)}}, res, Env{})
	if got.OK {
		t.Fatal("OK = true, want a failure for a timed-out command")
	}
	if !strings.Contains(got.Hint, "suite.timeout") {
		t.Errorf("Hint = %q, want it to name suite.timeout", got.Hint)
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
		{"not_matches absent pattern passes", &spec.Assert{Stdout: &spec.StreamAssert{NotMatches: strp("(?i)error|warn")}}, true},
		{"not_matches present pattern fails", &spec.Assert{Stdout: &spec.StreamAssert{NotMatches: strp("A.+e")}}, false},
		{"not_matches on a line", &spec.Assert{Stdout: &spec.StreamAssert{Line: intp(1), NotMatches: strp("Carol")}}, true},
		{"equals trailing-newline tolerant", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("Alice and Bob")}}, true},
		{"not_equals differs", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("someone else")}}, true},
		{"not_equals matches fails", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("Alice and Bob")}}, false},
		{"not_equals newline tolerant", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("Alice and Bob\n")}}, false},
		{"not_equals on a line", &spec.Assert{Stdout: &spec.StreamAssert{Line: intp(1), NotEquals: strp("nope")}}, true},
		{"stderr empty", &spec.Assert{Stderr: &spec.StreamAssert{Empty: boolp(true)}}, true},
		{"stdout not empty", &spec.Assert{Stdout: &spec.StreamAssert{Empty: boolp(false)}}, true},
		// Text matchers compose (AND): every one that is set must hold.
		{"contains and not_contains both hold", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice"}, NotContains: spec.StringList{"Carol"}}}, true},
		{"contains holds but not_contains fails", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice"}, NotContains: spec.StringList{"Bob"}}}, false},
		{"contains fails while not_contains holds", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Carol"}, NotContains: spec.StringList{"Dave"}}}, false},
		{"matches and not_matches both hold", &spec.Assert{Stdout: &spec.StreamAssert{Matches: strp("A.+e"), NotMatches: strp("(?i)error")}}, true},
		{"contains not_contains and matches all hold", &spec.Assert{Stdout: &spec.StreamAssert{Contains: spec.StringList{"Alice"}, NotContains: spec.StringList{"Carol"}, Matches: strp("Bob")}}, true},
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

// TestCheck_Stream_EqualsTrailingNewline is a regression: `equals`/`not_equals`
// tolerate a single phantom trailing newline (the one most commands emit, which
// a YAML block scalar also carries), not an arbitrary run of trailing blank
// lines. Trimming every trailing newline made `equals "hello"` pass against
// output with extra blank lines, and `not_equals` report a false "equal", and
// it disagreed with the line: matcher, which drops exactly one trailing newline
// so a deliberate trailing blank line stays addressable (see splitLines).
func TestCheck_Stream_EqualsTrailingNewline(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		stdout string
		a      *spec.Assert
		wantOK bool
	}{
		{"single trailing newline tolerated", "hello\n", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("hello")}}, true},
		{"no newline either side", "hello", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("hello")}}, true},
		{"want carries the newline", "hello", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("hello\n")}}, true},
		{"extra trailing blank lines are not equal", "hello\n\n\n", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("hello")}}, false},
		{"a deliberate trailing blank line is distinguishable", "hello\n\n", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("hello\n")}}, false},
		{"empty output is not equal to blank lines", "\n\n", &spec.Assert{Stdout: &spec.StreamAssert{Equals: strp("")}}, false},
		{"not_equals sees the extra blank lines", "hello\n\n\n", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("hello")}}, true},
		{"not_equals stays tolerant of one newline", "hello\n", &spec.Assert{Stdout: &spec.StreamAssert{NotEquals: strp("hello")}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: []byte(tc.stdout)}
			if got := Check(tc.a, res, Env{}); got.OK != tc.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tc.wantOK, got.Hint)
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

// TestCheck_Line covers the 1-based line selector, modeled on
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
		// Regression: a deliberate trailing blank line stays addressable — only the
		// single phantom final newline is dropped, not every trailing newline.
		{"trailing blank line addressable", "hello\n\n", &spec.StreamAssert{Line: intp(2), Empty: boolp(true)}, true},
		{"content after blank preserved", "a\n\nb\n", &spec.StreamAssert{Line: intp(3), Equals: strp("b")}, true},
		// A bare newline is one blank line, not zero (emptiness is judged before
		// the single-newline trim).
		{"bare newline is one blank line", "\n", &spec.StreamAssert{Line: intp(1), Empty: boolp(true)}, true},
		{"bare newline has no line 2", "\n", &spec.StreamAssert{Line: intp(2), Empty: boolp(true)}, false},
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

// TestCheck_JSON_NumericStringStrict is a regression for the toFloat bug where
// fmt.Sscanf("%g") accepted a numeric PREFIX and ignored trailing bytes, so
// "1.2.3" parsed as 1.2 and "3abc" as 3. That made two different version strings
// compare equal and let a non-numeric field silently satisfy a numeric matcher.
// A numeric-string coercion must require the WHOLE string to be a valid number.
func TestCheck_JSON_NumericStringStrict(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte(`{"ver":"1.2.3","other":"1.2.9","mixed":"3abc","spaced":"3 4","clean":"7"}`)}
	tests := []struct {
		name   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		// Two distinct version strings must NOT be equal (both used to coerce to 1.2).
		{"distinct versions not equal", &spec.JSONAssert{Path: "$.ver", Equals: "1.2.9"}, false},
		// Same version string still equals itself via string fallback.
		{"same version equals", &spec.JSONAssert{Path: "$.ver", Equals: "1.2.3"}, true},
		// A version string is not a number, so a numeric compare must fail, not coerce to 1.2.
		{"version not numeric for gt", &spec.JSONAssert{Path: "$.ver", Gt: f64p(1)}, false},
		// "3abc" is not numeric.
		{"mixed not numeric", &spec.JSONAssert{Path: "$.mixed", Gte: f64p(3)}, false},
		{"mixed not equal to 3", &spec.JSONAssert{Path: "$.mixed", Equals: 3}, false},
		// "3 4" is not a single number.
		{"spaced not numeric", &spec.JSONAssert{Path: "$.spaced", Lt: f64p(4)}, false},
		// A genuinely clean numeric string still coerces.
		{"clean numeric string compares", &spec.JSONAssert{Path: "$.clean", Gte: f64p(7)}, true},
		{"clean numeric string equals number", &spec.JSONAssert{Path: "$.clean", Equals: 7}, true},
		// Two distinct strings that merely parse to the same float must NOT be
		// equal under exact `equals` (string-vs-string is byte-exact).
		{"7 not equal to string 7.0", &spec.JSONAssert{Path: "$.clean", Equals: "7.0"}, false},
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

// TestCheck_YAML_UnsignedInteger is a regression (CodeRabbit): goccy/go-yaml
// decodes an integer that overflows int64 as uint64, so the numeric matchers
// must recognize unsigned kinds — otherwise a large value falls back to string
// comparison and fails gt/gte/lt/lte and equals.
func TestCheck_YAML_UnsignedInteger(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("big: 18446744073709551615\n")} // math.MaxUint64
	tests := []struct {
		name   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		{"gt below", &spec.JSONAssert{Path: "$.big", Gt: f64p(1e19)}, true},
		{"lt above", &spec.JSONAssert{Path: "$.big", Lt: f64p(2e19)}, true},
		{"gte equal-ish", &spec.JSONAssert{Path: "$.big", Gte: f64p(1e19)}, true},
		{"not gt above", &spec.JSONAssert{Path: "$.big", Gt: f64p(2e19)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{YAML: tt.j}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_JSON_LargeInteger is a regression for numeric matchers on integers
// that exceed float64's 53-bit exact range. oj decodes an integer beyond int64
// as json.Number; both cases previously misbehaved: equals compared through
// float64 and reported two distinct ids equal, and gt/gte/lt/lte rejected a
// json.Number as "not numeric".
func TestCheck_JSON_LargeInteger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		data   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		// int64 values one apart, both beyond 2^53: must NOT be equal.
		{"distinct int64 not equal", `{"id":9007199254740993}`, &spec.JSONAssert{Path: "$.id", Equals: int64(9007199254740992)}, false},
		{"same int64 equal", `{"id":9007199254740993}`, &spec.JSONAssert{Path: "$.id", Equals: int64(9007199254740993)}, true},
		// A value beyond int64 decodes as json.Number; comparisons must work.
		{"json.Number gt", `{"n":10000000000000000000}`, &spec.JSONAssert{Path: "$.n", Gt: f64p(1)}, true},
		{"json.Number lt", `{"n":10000000000000000000}`, &spec.JSONAssert{Path: "$.n", Lt: f64p(2e19)}, true},
		// Distinct integers beyond uint64, one apart: equals stays exact.
		{"distinct huge not equal", `{"n":100000000000000000001}`, &spec.JSONAssert{Path: "$.n", Equals: "100000000000000000000"}, false},
		{"same huge equal", `{"n":100000000000000000000}`, &spec.JSONAssert{Path: "$.n", Equals: "100000000000000000000"}, true},
		// A json.Number is a real number, so equals against a numeric spec value
		// compares numerically, not lexically: whitespace around the expected
		// digits does not defeat the match.
		{"huge equals ignores surrounding space", `{"n":100000000000000000000}`, &spec.JSONAssert{Path: "$.n", Equals: " 100000000000000000000 "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: []byte(tt.data)}
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: tt.j}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_JSON_MatchesWholeNumberFloat is a regression: a JSON whole-number
// float (1000000.0) was rendered with %v as "1e+06", so a matches pattern
// written against the digits ("^1000000$") failed. Numeric nodes must render
// without scientific notation.
func TestCheck_JSON_MatchesWholeNumberFloat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		data   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		{"whole float matches digits", `{"n":1000000.0}`, &spec.JSONAssert{Path: "$.n", Matches: strp("^1000000$")}, true},
		{"whole float not scientific", `{"n":1000000.0}`, &spec.JSONAssert{Path: "$.n", Matches: strp("e\\+")}, false},
		{"integer matches digits", `{"n":9007199254740993}`, &spec.JSONAssert{Path: "$.n", Matches: strp("^9007199254740993$")}, true},
		{"fractional float still matches", `{"n":0.0001}`, &spec.JSONAssert{Path: "$.n", Matches: strp("^0.0001$")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: []byte(tt.data)}
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
	// A multi-byte string's length is its character count, not its byte count:
	// "café" is 4 (not 5). Verified separately so the shared table stays ASCII.
	unicodeRes := &runner.Result{Stdout: []byte(`{"name":"café","emoji":"a→b"}`)}
	for _, tc := range []struct {
		path string
		n    int
		ok   bool
	}{
		{"$.name", 4, true},
		{"$.name", 5, false}, // byte count would wrongly pass here
		{"$.emoji", 3, true}, // a, →, b
	} {
		got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: &spec.JSONAssert{Path: tc.path, Length: intp(tc.n)}}}, unicodeRes, Env{})
		if got.OK != tc.ok {
			t.Errorf("unicode length %s == %d: OK = %v, want %v (%s)", tc.path, tc.n, got.OK, tc.ok, got.Hint)
		}
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

// TestCheck_File_SymlinkEscapeRejected proves that lexical containment is not
// enough: a file assertion target that is a symlink inside the workdir but points
// at a host file outside it must not be read through (issue #16). The untrusted
// program under test could plant such a link to disclose /etc/passwd et al. into
// the report; checkFile must refuse to follow it.
func TestCheck_File_SymlinkEscapeRejected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	workdir := t.TempDir()
	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("top-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A link planted inside the workdir (so lexical containment passes) that
	// resolves to the out-of-root secret.
	if err := os.Symlink(secret, filepath.Join(workdir, "leak.txt")); err != nil {
		t.Fatal(err)
	}
	got := Check(&spec.Assert{File: &spec.FileAssert{Path: "leak.txt", Contains: spec.StringList{"top-secret"}}}, nil, Env{Workdir: workdir})
	if got.OK {
		t.Fatalf("read through a workdir symlink pointing outside the root")
	}
	if strings.Contains(got.Actual, "top-secret") || strings.Contains(got.Hint, "top-secret") || strings.Contains(string(got.ArtifactActual), "top-secret") {
		t.Errorf("the out-of-root secret leaked into the result: %+v", got)
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

// TestBorderedScreen_MultibyteAligns is a regression test for a width bug in the
// screen-assert failure box: it measured line width and padding in bytes, so a
// rendered TUI screen containing box-drawing characters (─│┌, 3 bytes each) or
// CJK text produced a ragged right border — exactly the screens atago's pty/TUI
// assertions exist to check. Every framed content row must have the same rune
// width so the closing "|" column lines up.
func TestBorderedScreen_MultibyteAligns(t *testing.T) {
	t.Parallel()
	// An ASCII line and a multibyte line of different byte lengths but knowable
	// rune widths. Byte-based padding would give these rows different rune widths.
	screen := "abc\n日本\n┌──┐"
	out := borderedScreen(screen)

	lines := strings.Split(out, "\n")
	if len(lines) < 3 {
		t.Fatalf("bordered output has too few lines:\n%s", out)
	}
	// The top bar sets the box width; every subsequent row (content rows and the
	// bottom bar) must match it rune-for-rune.
	want := utf8.RuneCountInString(lines[0])
	for i, l := range lines {
		if got := utf8.RuneCountInString(l); got != want {
			t.Errorf("row %d rune width = %d, want %d (ragged border)\nrow: %q\nfull:\n%s", i, got, want, l, out)
		}
	}
}

// ---------------------------------------------------------------------------
// pdf.go: inflate, readPDFString, unescapePDFByte, metadataActual, parsePDF
// ---------------------------------------------------------------------------

// TestInflate_RoundTripAndError exercises the FlateDecode success path (a real
// zlib stream) and the raw-stream error path (a non-Flate stream returns an
// error so parsePDF keeps the raw bytes).
func TestInflate_RoundTripAndError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	payload := []byte("BT (Compressed body) Tj ET")
	if _, err := zw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	// Prepend stray CR/LF so we exercise the TrimLeft in inflate.
	compressed := append([]byte("\r\n"), buf.Bytes()...)
	out, err := inflate(compressed)
	if err != nil {
		t.Fatalf("inflate valid zlib: %v", err)
	}
	if !bytes.Equal(out, payload) {
		t.Errorf("inflate roundtrip = %q, want %q", out, payload)
	}

	if _, err := inflate([]byte("this is not zlib data at all")); err == nil {
		t.Error("inflate on raw (non-zlib) bytes must return an error so the caller keeps raw bytes")
	}
	// Empty input must not panic and must error.
	if _, err := inflate(nil); err == nil {
		t.Error("inflate(nil) should error, not silently succeed")
	}
}

// TestReadPDFString_Escapes covers unescapePDFByte (all mapped escapes plus the
// default passthrough) and balanced/nested/unbalanced parentheses.
func TestReadPDFString_Escapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		lit  string
		want string
	}{
		{"plain", `(hello)`, "hello"},
		{"newline-escape", `(a\nb)`, "a\nb"},
		{"tab-escape", `(a\tb)`, "a\tb"},
		{"cr-escape", `(a\rb)`, "a\rb"},
		{"escaped-paren", `(a\)b)`, "a)b"},
		{"escaped-open-paren", `(a\(b)`, "a(b"},
		{"escaped-backslash", `(a\\b)`, `a\b`},
		{"default-passthrough", `(a\zb)`, "azb"}, // \z is not special -> literal z
		{"nested-balanced", `(a (b) c)`, "a (b) c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := decodePDFString([]byte(tt.lit)); got != tt.want {
				t.Errorf("decodePDFString(%q) = %q, want %q", tt.lit, got, tt.want)
			}
		})
	}
}

// TestReadPDFString_Boundaries feeds out-of-range / non-paren openings and a
// trailing backslash so the function never panics.
func TestReadPDFString_Boundaries(t *testing.T) {
	t.Parallel()
	data := []byte(`(abc)`)
	if _, ok := readPDFString(data, -1); ok {
		t.Error("negative open index must return false")
	}
	if _, ok := readPDFString(data, len(data)); ok {
		t.Error("open index at len must return false")
	}
	if _, ok := readPDFString([]byte("xyz"), 0); ok {
		t.Error("a non-'(' opening byte must return false")
	}
	// Trailing backslash with no following byte: must not panic.
	if _, ok := readPDFString([]byte(`(ab\`), 0); !ok {
		t.Error("unterminated string is read leniently and returns true")
	}
	// Unbalanced open paren: lenient parser returns what it has.
	if s, ok := readPDFString([]byte(`(unclosed`), 0); !ok || s != "unclosed" {
		t.Errorf("unbalanced literal = %q ok=%v", s, ok)
	}
}

// TestParsePDF_Compressed proves parsePDF extracts text from a FlateDecode
// content stream, exercising the inflate branch inside parsePDF.
func TestParsePDF_Compressed(t *testing.T) {
	t.Parallel()
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	if _, err := zw.Write([]byte("BT (Zlib hidden text) Tj ET")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.5\n1 0 obj\n<< /Type /Page >>\nendobj\n4 0 obj\n<< /Length 99 /Filter /FlateDecode >>\nstream\n")
	pdf.Write(zbuf.Bytes())
	pdf.WriteString("\nendstream\nendobj\n%%EOF\n")

	doc := parsePDF(pdf.Bytes())
	if !strings.Contains(doc.text, "Zlib hidden text") {
		t.Errorf("compressed text not extracted: %q", doc.text)
	}
}

// TestParsePDF_Malformed feeds garbage / truncated bodies (after the header,
// which checkPDF requires) and asserts parsePDF never panics and returns a
// zero-ish doc.
func TestParsePDF_Malformed(t *testing.T) {
	t.Parallel()
	inputs := [][]byte{
		[]byte("%PDF-1.4\n"),
		[]byte("%PDF-1.4\nstream\n"),                    // truncated stream, no endstream
		[]byte("%PDF-1.4\nstream\n\xff\xfe\nendstream"), // binary garbage stream
		[]byte("%PDF-1.4\n/Type /Page /Title ("),        // dangling metadata literal
		append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte{0x00}, 512)...),
	}
	for i, in := range inputs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("parsePDF panicked on input %d: %v", i, r)
				}
			}()
			_ = parsePDF(in)
		}()
	}
}

// TestCheckPDF_PageBoundaries hits the min/max off-by-one edges and the
// metadataActual "field not present" branch.
func TestCheckPDF_PageBoundaries(t *testing.T) {
	t.Parallel()
	wd := writePDF(t, minimalPDF) // 1 page
	// min == pages passes, min == pages+1 fails.
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", MinPages: ptrInt(1)}); !cr.OK {
		t.Errorf("min 1 on a 1-page PDF should pass: %+v", cr)
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", MinPages: ptrInt(2)}); cr.OK {
		t.Error("min 2 on a 1-page PDF must fail")
	}
	// max == pages passes, max == pages-1 fails.
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", MaxPages: ptrInt(1)}); !cr.OK {
		t.Errorf("max 1 on a 1-page PDF should pass: %+v", cr)
	}
	if cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", MaxPages: ptrInt(0)}); cr.OK {
		t.Error("max 0 on a 1-page PDF must fail")
	}
	// A metadata field absent from the doc reports "field not present".
	cr := checkPDFOK(t, wd, &spec.PDFAssert{Path: "doc.pdf", Metadata: map[string]string{"subject": "anything"}})
	if cr.OK {
		t.Fatal("absent metadata field must fail")
	}
	if !strings.Contains(cr.Actual, "field not present") {
		t.Errorf("Actual = %q, want it to say the field is not present", cr.Actual)
	}
}

func TestMetadataActual(t *testing.T) {
	t.Parallel()
	if got := metadataActual("", false); got != "field not present" {
		t.Errorf("metadataActual(_, false) = %q", got)
	}
	if got := metadataActual("value", true); got != "value" {
		t.Errorf("metadataActual present = %q", got)
	}
}

// ---------------------------------------------------------------------------
// screen.go: checkScreen, borderedScreen
// ---------------------------------------------------------------------------

func TestCheckScreen_NoPTY(t *testing.T) {
	t.Parallel()
	sa := &spec.StreamAssert{Contains: spec.StringList{"x"}}
	if cr := checkScreen(sa, nil, Env{}); cr.OK {
		t.Error("screen assert with no result should not pass")
	}
	if cr := checkScreen(sa, &runner.Result{IsPTY: false}, Env{}); cr.OK {
		t.Error("screen assert with a non-pty result should not pass")
	}
}

func TestCheckScreen_PassAndFail(t *testing.T) {
	t.Parallel()
	res := &runner.Result{IsPTY: true, Screen: []byte("hello\nworld")}
	if cr := checkScreen(&spec.StreamAssert{Contains: spec.StringList{"world"}}, res, Env{}); !cr.OK {
		t.Errorf("matching screen should pass: %+v", cr)
	}
	fail := checkScreen(&spec.StreamAssert{Contains: spec.StringList{"absent"}}, res, Env{})
	if fail.OK {
		t.Fatal("non-matching screen should fail")
	}
	if !strings.Contains(fail.Actual, "+---") || !strings.Contains(fail.Actual, "| hello") {
		t.Errorf("failure Actual should be a bordered screen, got:\n%s", fail.Actual)
	}
	if fail.ArtifactKind != "screen" {
		t.Errorf("ArtifactKind = %q, want screen", fail.ArtifactKind)
	}
	if !bytes.Equal(fail.ArtifactActual, res.Screen) {
		t.Error("ArtifactActual should carry the raw screen bytes")
	}
}

func TestBorderedScreen(t *testing.T) {
	t.Parallel()
	// Ragged lines: the border width follows the widest line and every row is
	// right-padded to it, so trailing whitespace stays visible.
	out := borderedScreen("ab\nabcd\n")
	lines := strings.Split(out, "\n")
	width := len(lines[0])
	for _, l := range lines {
		if len(l) != width {
			t.Errorf("uneven border row %q (len %d != %d)", l, len(l), width)
		}
	}
	// Single empty screen must not panic and still frames a zero-width box.
	_ = borderedScreen("")
}

// ---------------------------------------------------------------------------
// jsonmatch.go: toFloat, jsonMatches, single, opSymbol, applyJSONMatch, checkYAML
// ---------------------------------------------------------------------------

func TestToFloat_AllKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		in     any
		wantOK bool
		want   float64
	}{
		{"int", int(3), true, 3},
		{"int8", int8(-3), true, -3},
		{"int16", int16(300), true, 300},
		{"int32", int32(-9), true, -9},
		{"int64", int64(1 << 40), true, float64(int64(1) << 40)},
		{"uint", uint(7), true, 7},
		{"uint8", uint8(255), true, 255},
		{"uint16", uint16(65535), true, 65535},
		{"uint32", uint32(1), true, 1},
		{"uint64-huge", uint64(1) << 62, true, float64(uint64(1) << 62)},
		{"float32", float32(1.5), true, 1.5},
		{"float64", float64(2.25), true, 2.25},
		{"numeric-string", "42", true, 42},
		{"trimmed-string", "  3.5  ", true, 3.5},
		{"non-numeric-string", "abc", false, 0},
		{"version-string-not-a-float", "1.2.3", false, 0}, // #: full-consume guard
		{"trailing-junk", "3abc", false, 0},
		{"bool-not-numeric", true, false, 0},
		{"nil-not-numeric", nil, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat(tt.in)
			if ok != tt.wantOK {
				t.Fatalf("toFloat(%v) ok = %v, want %v", tt.in, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("toFloat(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestOpSymbol(t *testing.T) {
	t.Parallel()
	for op, sym := range map[string]string{"gt": ">", "gte": ">=", "lt": "<", "lte": "<=", "weird": "weird"} {
		if got := opSymbol(op); got != sym {
			t.Errorf("opSymbol(%q) = %q, want %q", op, got, sym)
		}
	}
}

// TestJSONCompare_Metamorphic proves gt and lt can never both pass for the same
// value/threshold, and that a non-numeric selected value fails cleanly.
func TestJSONCompare_Metamorphic(t *testing.T) {
	t.Parallel()
	data := []byte(`{"n": 5, "s": "hello"}`)
	gt := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.n", Gt: f64p(5)})
	lt := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.n", Lt: f64p(5)})
	if gt.OK && lt.OK {
		t.Error("gt 5 and lt 5 must not both pass for value 5")
	}
	// gte 5 and lte 5 both pass (boundary) — that is correct, not contradictory.
	gte := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.n", Gte: f64p(5)})
	lte := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.n", Lte: f64p(5)})
	if !gte.OK || !lte.OK {
		t.Error("gte 5 and lte 5 should both pass at the boundary")
	}
	// Non-numeric node cannot be compared.
	if cr := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.s", Gt: f64p(0)}); cr.OK {
		t.Error("gt on a string value must fail")
	}
}

func TestJSONMatches_Errors(t *testing.T) {
	t.Parallel()
	data := []byte(`{"v": "abc123"}`)
	if cr := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.v", Matches: strp(`\d+`)}); !cr.OK {
		t.Errorf("matches digits should pass: %+v", cr)
	}
	if cr := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.v", Matches: strp(`^zzz$`)}); cr.OK {
		t.Error("non-matching pattern should fail")
	}
	if cr := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.v", Matches: strp(`(`)}); cr.OK {
		t.Error("invalid regexp should fail, not panic")
	}
}

func TestSingle_ZeroAndMultiple(t *testing.T) {
	t.Parallel()
	data := []byte(`{"a": [1, 2, 3]}`)
	// Zero matches.
	if cr := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.missing", Equals: 1}); cr.OK {
		t.Error("no-match path should fail via single()")
	}
	// Multiple matches (wildcard) → single() rejects.
	cr := checkJSON("d", "stdout", data, &spec.JSONAssert{Path: "$.a[*]", Equals: 1})
	if cr.OK {
		t.Fatal("multiple matches should fail single()")
	}
	if !strings.Contains(cr.Actual, "matches") {
		t.Errorf("Actual = %q, want it to report the match count", cr.Actual)
	}
}

func TestApplyJSONMatch_NoMatcherAndBadPath(t *testing.T) {
	t.Parallel()
	if cr := checkJSON("d", "stdout", []byte(`{"a":1}`), &spec.JSONAssert{Path: "$.a"}); cr.OK {
		t.Error("a matcher with no operator must fail")
	}
	if cr := checkJSON("d", "stdout", []byte(`{"a":1}`), &spec.JSONAssert{Path: "$[", Equals: 1}); cr.OK {
		t.Error("an invalid JSON path must fail, not panic")
	}
}

func TestCheckYAML(t *testing.T) {
	t.Parallel()
	yamlData := []byte("name: Ada\nage: 42\n")
	if cr := checkYAML("d", "stdout", yamlData, &spec.JSONAssert{Path: "$.name", Equals: "Ada"}); !cr.OK {
		t.Errorf("yaml equals should pass: %+v", cr)
	}
	if cr := checkYAML("d", "stdout", []byte("   \n"), &spec.JSONAssert{Path: "$.name", Equals: "Ada"}); cr.OK {
		t.Error("empty YAML should fail")
	}
	if cr := checkYAML("d", "stdout", []byte("a: [b, c"), &spec.JSONAssert{Path: "$.a", Length: intp(2)}); cr.OK {
		t.Error("malformed YAML should fail, not panic")
	}
}

// ---------------------------------------------------------------------------
// dir.go: dirStatActual error branch, count boundaries, glob error
// ---------------------------------------------------------------------------

func TestCheckDir_CountBoundaries(t *testing.T) {
	t.Parallel()
	wd := makeTree(t) // site has 3 direct entries
	// min == n passes, min == n+1 fails.
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", MinCount: ptrInt(3)}); !cr.OK {
		t.Errorf("min 3 should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", MinCount: ptrInt(4)}); cr.OK {
		t.Error("min 4 must fail on 3 entries")
	}
	// max == n passes, max == n-1 fails.
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", MaxCount: ptrInt(3)}); !cr.OK {
		t.Errorf("max 3 should pass: %+v", cr)
	}
	if cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", MaxCount: ptrInt(2)}); cr.OK {
		t.Error("max 2 must fail on 3 entries")
	}
}

// TestCheckDir_CountOnMissing hits the dirStatActual error branch: a count
// constraint on a non-existent directory (no exists field) reports the stat
// error rather than "not a directory".
func TestCheckDir_CountOnMissing(t *testing.T) {
	t.Parallel()
	wd := makeTree(t)
	cr := checkDirOK(t, wd, &spec.DirAssert{Path: "does-not-exist", Count: ptrInt(0)})
	if cr.OK {
		t.Fatal("count on a missing dir must fail")
	}
	if cr.Actual == "not a directory" || cr.Actual == "" {
		t.Errorf("Actual = %q, want the underlying stat error", cr.Actual)
	}
}

func TestCheckDir_GlobInvalid(t *testing.T) {
	t.Parallel()
	wd := makeTree(t)
	cr := checkDirOK(t, wd, &spec.DirAssert{Path: "site", Glob: "["})
	if cr.OK {
		t.Error("an invalid glob pattern must fail, not match")
	}
	if !strings.Contains(cr.Hint, "invalid glob") {
		t.Errorf("Hint = %q, want it to flag the invalid glob", cr.Hint)
	}
}

func TestDirStatActual(t *testing.T) {
	t.Parallel()
	if got := dirStatActual(nil, os.ErrNotExist); got != os.ErrNotExist.Error() {
		t.Errorf("err branch = %q", got)
	}
	// A regular file's FileInfo → "not a directory".
	f := filepath.Join(t.TempDir(), "x")
	if err := os.WriteFile(f, []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirStatActual(info, nil); got != "not a directory" {
		t.Errorf("not-a-dir branch = %q", got)
	}
}

// ---------------------------------------------------------------------------
// file.go: not_contains, executable, exists:false, default, readFile error
// ---------------------------------------------------------------------------

func writeWorkFile(t *testing.T, name, body string, mode os.FileMode) string {
	t.Helper()
	wd := t.TempDir()
	if err := os.WriteFile(filepath.Join(wd, name), []byte(body), mode); err != nil {
		t.Fatal(err)
	}
	return wd
}

func checkFileOK(t *testing.T, wd string, f *spec.FileAssert) *CheckResult {
	t.Helper()
	return Check(&spec.Assert{File: f}, nil, Env{Workdir: wd})
}

func TestCheckFile_ExistsFalseAndDefault(t *testing.T) {
	t.Parallel()
	wd := writeWorkFile(t, "a.txt", "hi", 0o600)
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "missing.txt", Exists: boolp(false)}); !cr.OK {
		t.Errorf("exists:false on a missing file should pass: %+v", cr)
	}
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "a.txt", Exists: boolp(false)}); cr.OK {
		t.Error("exists:false on a present file must fail")
	}
	// No matcher set → hint.
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "a.txt"}); cr.OK {
		t.Error("a file assertion with no field must fail")
	}
}

func TestCheckFile_NotContains(t *testing.T) {
	t.Parallel()
	wd := writeWorkFile(t, "a.txt", "hello world", 0o600)
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "a.txt", NotContains: spec.StringList{"absent"}}); !cr.OK {
		t.Errorf("not_contains absent should pass: %+v", cr)
	}
	present := checkFileOK(t, wd, &spec.FileAssert{Path: "a.txt", NotContains: spec.StringList{"world"}})
	if present.OK {
		t.Error("not_contains present should fail")
	}
	if present.ArtifactKind != "file" {
		t.Errorf("ArtifactKind = %q, want file", present.ArtifactKind)
	}
	// Reading a missing file for a contains check surfaces a read error.
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "missing.txt", Contains: spec.StringList{"x"}}); cr.OK {
		t.Error("contains on a missing file must fail to read")
	}
}

func TestCheckFile_Executable(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Windows ignores POSIX executable bits, so a 0755 file is not observably executable")
	}
	wd := writeWorkFile(t, "run.sh", "#!/bin/sh\n", 0o755)
	if err := os.WriteFile(filepath.Join(wd, "plain.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "run.sh", Executable: boolp(true)}); !cr.OK {
		t.Errorf("executable:true on a 0755 file should pass: %+v", cr)
	}
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "plain.txt", Executable: boolp(false)}); !cr.OK {
		t.Errorf("executable:false on a 0600 file should pass: %+v", cr)
	}
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "plain.txt", Executable: boolp(true)}); cr.OK {
		t.Error("executable:true on a non-exec file must fail")
	}
	// Stat error on a missing file.
	if cr := checkFileOK(t, wd, &spec.FileAssert{Path: "missing", Executable: boolp(true)}); cr.OK {
		t.Error("executable on a missing file must fail")
	}
}

func TestExistenceAndExecutabilityWords(t *testing.T) {
	t.Parallel()
	if existence(true) != "exist" || existence(false) != "not exist" {
		t.Error("existence phrasing wrong")
	}
	if executability(true) != "be" || executability(false) != "not be" {
		t.Error("executability phrasing wrong")
	}
}

// ---------------------------------------------------------------------------
// http.go: grpc_status, dbRows/grpcMessage/cdpValue, header value matchers
// ---------------------------------------------------------------------------

func TestCheckGRPCStatus(t *testing.T) {
	t.Parallel()
	if cr := checkGRPCStatus(intp(0), nil); cr.OK {
		t.Error("grpc_status with no result must fail")
	}
	if cr := checkGRPCStatus(intp(0), &runner.Result{IsGRPC: false}); cr.OK {
		t.Error("grpc_status without a gRPC call must fail")
	}
	res := &runner.Result{IsGRPC: true, GRPCStatus: 5}
	if cr := checkGRPCStatus(intp(5), res); !cr.OK {
		t.Errorf("matching grpc status should pass: %+v", cr)
	}
	if cr := checkGRPCStatus(intp(0), res); cr.OK {
		t.Error("mismatched grpc status must fail")
	}
}

func TestCapturedBytesHelpers(t *testing.T) {
	t.Parallel()
	if dbRows(nil) != nil || grpcMessage(nil) != nil || cdpValue(nil) != nil || httpBody(nil) != nil {
		t.Error("captured-bytes helpers should return nil for a nil result")
	}
	res := &runner.Result{
		RowsJSON:    []byte(`[{"id":1}]`),
		MessageJSON: []byte(`{"ok":true}`),
		CDPValue:    []byte("value"),
		Body:        []byte("body"),
	}
	if string(dbRows(res)) != `[{"id":1}]` {
		t.Error("dbRows mismatch")
	}
	if string(grpcMessage(res)) != `{"ok":true}` {
		t.Error("grpcMessage mismatch")
	}
	if string(cdpValue(res)) != "value" {
		t.Error("cdpValue mismatch")
	}
	if string(httpBody(res)) != "body" {
		t.Error("httpBody mismatch")
	}
}

func TestCheckHeaderValue_AllMatchers(t *testing.T) {
	t.Parallel()
	h := func(m *spec.HeaderMatch) *CheckResult {
		return checkHeaderValue(m, "application/json; charset=utf-8", "response")
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Equals: strp("application/json; charset=utf-8")}); !cr.OK {
		t.Errorf("equals should pass: %+v", cr)
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Equals: strp("text/plain")}); cr.OK {
		t.Error("equals mismatch must fail")
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Contains: strp("json")}); !cr.OK {
		t.Errorf("contains should pass: %+v", cr)
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Contains: strp("xml")}); cr.OK {
		t.Error("contains mismatch must fail")
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Matches: strp(`^application/`)}); !cr.OK {
		t.Errorf("matches should pass: %+v", cr)
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Matches: strp(`^text/`)}); cr.OK {
		t.Error("matches mismatch must fail")
	}
	if cr := h(&spec.HeaderMatch{Name: "Content-Type", Matches: strp(`(`)}); cr.OK {
		t.Error("invalid regexp must fail, not panic")
	}
	// No matcher set → hint.
	if cr := h(&spec.HeaderMatch{Name: "Content-Type"}); cr.OK {
		t.Error("a header match with no matcher must fail")
	}
}

// ---------------------------------------------------------------------------
// mock.go: checkMock, mockFilterLabel, plural, summarizeRecords
// ---------------------------------------------------------------------------

func mockEnv(records map[string][]mock.Record) Env {
	return Env{MockRecords: func(name string) ([]mock.Record, bool) {
		r, ok := records[name]
		return r, ok
	}}
}

func TestCheckMock_NoServerAndNilProvider(t *testing.T) {
	t.Parallel()
	if cr := checkMock(&spec.MockAssert{Name: "api"}, Env{}); cr.OK {
		t.Error("mock assert with no provider must fail")
	}
	env := mockEnv(map[string][]mock.Record{})
	if cr := checkMock(&spec.MockAssert{Name: "unknown"}, env); cr.OK {
		t.Error("mock assert for an unknown server must fail")
	}
}

func TestCheckMock_CountAndFilter(t *testing.T) {
	t.Parallel()
	records := []mock.Record{
		{Method: "GET", Path: "/a", Status: 200},
		{Method: "POST", Path: "/a", Status: 201},
		{Method: "GET", Path: "/b", Status: 200},
	}
	env := mockEnv(map[string][]mock.Record{"api": records})

	// Exact count over all routes.
	if cr := checkMock(&spec.MockAssert{Name: "api", Count: intp(3)}, env); !cr.OK {
		t.Errorf("count 3 (any route) should pass: %+v", cr)
	}
	// Wrong count reports the summary of recorded requests.
	wrong := checkMock(&spec.MockAssert{Name: "api", Count: intp(2)}, env)
	if wrong.OK {
		t.Fatal("count 2 must fail")
	}
	if !strings.Contains(wrong.Actual, "GET /a -> 200") {
		t.Errorf("Actual should summarize recorded requests, got:\n%s", wrong.Actual)
	}
	// Filter by method+path.
	if cr := checkMock(&spec.MockAssert{Name: "api", Method: "get", Path: "/a", Count: intp(1)}, env); !cr.OK {
		t.Errorf("GET /a count 1 should pass: %+v", cr)
	}
	// Without count: at least one match required.
	if cr := checkMock(&spec.MockAssert{Name: "api", Path: "/b"}, env); !cr.OK {
		t.Errorf("path /b (>=1) should pass: %+v", cr)
	}
	if cr := checkMock(&spec.MockAssert{Name: "api", Path: "/zzz"}, env); cr.OK {
		t.Error("no request for /zzz must fail")
	}
}

func TestCheckMock_HeaderAndBody(t *testing.T) {
	t.Parallel()
	rec := mock.Record{Method: "POST", Path: "/ingest", Status: 200, Body: []byte(`{"k":"v"}`)}
	rec.Header = map[string][]string{"X-Token": {"secret"}}
	env := mockEnv(map[string][]mock.Record{"api": {rec}})

	pass := checkMock(&spec.MockAssert{
		Name:   "api",
		Header: &spec.HeaderMatch{Name: "X-Token", Equals: strp("secret")},
		Body:   &spec.StreamAssert{Contains: spec.StringList{`"k":"v"`}},
	}, env)
	if !pass.OK {
		t.Errorf("matching header+body should pass: %+v", pass)
	}
	// Header mismatch surfaces as failure.
	if cr := checkMock(&spec.MockAssert{Name: "api", Header: &spec.HeaderMatch{Name: "X-Token", Equals: strp("nope")}}, env); cr.OK {
		t.Error("header mismatch must fail")
	}
	// Body mismatch surfaces as failure.
	if cr := checkMock(&spec.MockAssert{Name: "api", Body: &spec.StreamAssert{Contains: spec.StringList{"absent"}}}, env); cr.OK {
		t.Error("body mismatch must fail")
	}
}

func TestMockFilterLabelAndPlural(t *testing.T) {
	t.Parallel()
	cases := []struct {
		m    *spec.MockAssert
		want string
	}{
		{&spec.MockAssert{}, "(any route)"},
		{&spec.MockAssert{Path: "/x"}, "for /x"},
		{&spec.MockAssert{Method: "get"}, "for GET"},
		{&spec.MockAssert{Method: "post", Path: "/y"}, "for POST /y"},
	}
	for _, c := range cases {
		if got := mockFilterLabel(c.m); got != c.want {
			t.Errorf("mockFilterLabel = %q, want %q", got, c.want)
		}
	}
	if plural("request", 1) != "request" || plural("request", 0) != "requests" || plural("request", 2) != "requests" {
		t.Error("plural wrong")
	}
	if summarizeRecords(nil) != "  (no requests recorded)" {
		t.Error("empty summarizeRecords wrong")
	}
}

// ---------------------------------------------------------------------------
// changes.go: describeChangesExpected with all three categories
// ---------------------------------------------------------------------------

func TestDescribeChangesExpected(t *testing.T) {
	t.Parallel()
	c := &spec.ChangesAssert{Created: list("a"), Modified: list("b"), Deleted: list("c")}
	got := describeChangesExpected(c)
	for _, want := range []string{"created [a]", "modified [b]", "deleted [c]"} {
		if !strings.Contains(got, want) {
			t.Errorf("describeChangesExpected = %q, want it to contain %q", got, want)
		}
	}
	if describeChangesExpected(&spec.ChangesAssert{}) != "" {
		t.Error("an empty changes assert should describe as empty string")
	}
}

// ---------------------------------------------------------------------------
// image.go: isSVG, format on garbage, similar_to on undecodable data
// ---------------------------------------------------------------------------

func TestIsSVG(t *testing.T) {
	t.Parallel()
	if !isSVG([]byte(`<?xml version="1.0"?>` + "\n" + `<svg xmlns="http://www.w3.org/2000/svg"></svg>`)) {
		t.Error("an SVG with an XML declaration should be detected")
	}
	if !isSVG([]byte("<svg></svg>")) {
		t.Error("a bare <svg should be detected")
	}
	if isSVG([]byte("<html><body></body></html>")) {
		t.Error("HTML must not be detected as SVG")
	}
	// A large blob whose "<svg" would sit past the 1024-byte scan window.
	big := append(bytes.Repeat([]byte(" "), 2000), []byte("<svg")...)
	if isSVG(big) {
		t.Error("an <svg past the 1024-byte head window must not be detected")
	}
}

func TestCheckImage_FormatUnknown(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	writeImage(t, wd, "junk.png", []byte("not an image at all"))
	cr := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "junk.png", Format: "png"}}, nil, Env{Workdir: wd})
	if cr.OK {
		t.Fatal("garbage bytes should not match format png")
	}
	if !strings.Contains(cr.Actual, "unknown") {
		t.Errorf("Actual = %q, want it to say unknown", cr.Actual)
	}
}

// TestCheckImage_SimilarTo_Undecodable proves an undecodable actual image fails
// cleanly (no panic) and still attaches image artifacts.
func TestCheckImage_SimilarTo_Undecodable(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	baseline := makePNG(t, 2, 2, color.RGBA{R: 255, A: 255})
	writeImage(t, wd, "base.png", baseline)
	writeImage(t, wd, "actual.png", []byte("\x89PNG\r\n\x1a\ntruncated"))
	cr := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "actual.png", SimilarTo: "base.png"}},
		nil, Env{Workdir: wd, SpecDir: wd})
	if cr.OK {
		t.Fatal("an undecodable actual image must fail similar_to")
	}
	if cr.ArtifactKind != "image" {
		t.Errorf("ArtifactKind = %q, want image", cr.ArtifactKind)
	}
}

func TestMeanPixelDiff_IdenticalAndDifferent(t *testing.T) {
	t.Parallel()
	red := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	blue := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			red.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
			blue.SetNRGBA(x, y, color.NRGBA{B: 255, A: 255})
		}
	}
	if d := meanPixelDiff(red, red); d != 0 {
		t.Errorf("identical images diff = %v, want 0", d)
	}
	d := meanPixelDiff(red, blue)
	if d <= 0 || d > 1 {
		t.Errorf("different images diff = %v, want in (0,1]", d)
	}
	// Heatmap of two different images encodes to a valid PNG the same size.
	hm := renderDiffHeatmap(red, blue)
	img, err := png.Decode(bytes.NewReader(hm))
	if err != nil {
		t.Fatalf("heatmap not a valid PNG: %v", err)
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Errorf("heatmap size = %v, want 4x4", img.Bounds())
	}
}

// ---------------------------------------------------------------------------
// stream.go: excerpt truncation, line selector out of range, splitLines
// ---------------------------------------------------------------------------

func TestExcerpt_Truncation(t *testing.T) {
	t.Parallel()
	short := "small"
	if excerpt(short) != short {
		t.Error("short strings pass through unchanged")
	}
	long := strings.Repeat("a", excerptLimit+50)
	got := excerpt(long)
	if !strings.HasSuffix(got, "... (truncated)") {
		t.Error("long strings get a truncation marker")
	}
	if len(got) >= len(long) {
		t.Error("truncated excerpt should be shorter than the input")
	}
}

func TestCheckStream_LineOutOfRange(t *testing.T) {
	t.Parallel()
	sa := &spec.StreamAssert{Line: intp(9), Contains: spec.StringList{"x"}}
	cr := checkStream("stdout", sa, []byte("only one line\n"), true, Env{})
	if cr.OK {
		t.Fatal("selecting line 9 of a 1-line stream must fail")
	}
	if !strings.Contains(cr.Hint, "out of range") {
		t.Errorf("Hint = %q, want an out-of-range message", cr.Hint)
	}
	// A deliberate trailing blank line stays addressable.
	if _, ok := selectLine("a\n\n", 2); !ok {
		t.Error("a deliberate trailing blank line should be addressable as line 2")
	}
	if got := countLines(""); got != 0 {
		t.Errorf("countLines(empty) = %d, want 0", got)
	}
	if got := countLines("\n"); got != 1 {
		t.Errorf("countLines(single newline) = %d, want 1", got)
	}
}

// TestStreamMatchers_Metamorphic proves a matcher and its negation never both
// pass on the same data (contains/not_contains, matches/not_matches,
// equals/not_equals).
func TestStreamMatchers_Metamorphic(t *testing.T) {
	t.Parallel()
	data := []byte("The quick brown fox\n")
	pairs := []struct {
		name string
		pos  *spec.StreamAssert
		neg  *spec.StreamAssert
	}{
		{"contains", &spec.StreamAssert{Contains: spec.StringList{"quick"}}, &spec.StreamAssert{NotContains: spec.StringList{"quick"}}},
		{"contains-absent", &spec.StreamAssert{Contains: spec.StringList{"zzz"}}, &spec.StreamAssert{NotContains: spec.StringList{"zzz"}}},
		{"matches", &spec.StreamAssert{Matches: strp(`fox`)}, &spec.StreamAssert{NotMatches: strp(`fox`)}},
		{"equals", &spec.StreamAssert{Equals: strp("The quick brown fox\n")}, &spec.StreamAssert{NotEquals: strp("The quick brown fox\n")}},
	}
	for _, p := range pairs {
		pos := checkStream("stdout", p.pos, data, true, Env{})
		neg := checkStream("stdout", p.neg, data, true, Env{})
		if pos.OK == neg.OK {
			t.Errorf("%s: a matcher and its negation returned the same verdict (both %v)", p.name, pos.OK)
		}
	}
}

// ---------------------------------------------------------------------------
// tree.go: hashFile error, walkTree error, ignoredPath patterns
// ---------------------------------------------------------------------------

func TestHashFile_Error(t *testing.T) {
	t.Parallel()
	if _, err := hashFile(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("hashFile on a missing path must error")
	}
}

func TestWalkTree_MissingDir(t *testing.T) {
	t.Parallel()
	if _, err := walkTree(filepath.Join(t.TempDir(), "absent"), nil); err == nil {
		t.Error("walkTree on a missing dir must error")
	}
}

func TestIgnoredPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rel    string
		ignore []string
		want   bool
	}{
		{"node_modules", []string{"node_modules/**"}, true},
		{"node_modules/pkg/x.js", []string{"node_modules/**"}, true},
		{"src/main.go", []string{"node_modules/**"}, false},
		{"a/b/c.log", []string{"*.log"}, true}, // basename match for /-less pattern
		{"a/b/c.txt", []string{"*.log"}, false},
		{"dist/app.js", []string{"dist/*.js"}, true}, // full-path path.Match
		{"deep/dist/app.js", []string{"dist/*.js"}, false},
	}
	for _, tt := range tests {
		if got := ignoredPath(tt.rel, tt.ignore); got != tt.want {
			t.Errorf("ignoredPath(%q, %v) = %v, want %v", tt.rel, tt.ignore, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// assert.go: CheckAll no targets, checkTarget default, streamBytes stderr,
// checkExitCode no result / default
// ---------------------------------------------------------------------------

func TestCheckAll_NoTargets(t *testing.T) {
	t.Parallel()
	results := CheckAll(&spec.Assert{}, nil, Env{})
	if len(results) != 1 || results[0].OK {
		t.Errorf("an assert with no targets should yield one failing result, got %+v", results)
	}
}

func TestStreamBytes_Stderr(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("OUT"), Stderr: []byte("ERR")}
	if string(streamBytes(res, "stdout")) != "OUT" {
		t.Error("streamBytes stdout mismatch")
	}
	if string(streamBytes(res, "stderr")) != "ERR" {
		t.Error("streamBytes stderr mismatch")
	}
	if streamBytes(nil, "stdout") != nil {
		t.Error("streamBytes(nil) should be nil")
	}
}

func TestCheckExitCode_NoResultAndDefault(t *testing.T) {
	t.Parallel()
	if cr := checkExitCode(&spec.ExitCode{Equals: intp(0)}, nil); cr.OK {
		t.Error("exit_code with no result must fail")
	}
	// An exit_code with no operator set falls to the default hint.
	if cr := checkExitCode(&spec.ExitCode{}, &runner.Result{ExitCode: 0}); cr.OK {
		t.Error("exit_code with no operator must fail")
	}
}

// TestCheckTarget_AllFamilies drives every remaining target family through the
// public Check entry point so the checkTarget switch is exercised end-to-end.
func TestCheckTarget_AllFamilies(t *testing.T) {
	t.Parallel()
	// DB rows.
	db := &runner.Result{IsDB: true, RowsJSON: []byte(`[{"id":1}]`)}
	if cr := Check(&spec.Assert{Rows: &spec.StreamAssert{JSON: &spec.JSONAssert{Path: "$[0].id", Equals: 1}}}, db, Env{}); !cr.OK {
		t.Errorf("rows target should pass: %+v", cr)
	}
	// gRPC status + message.
	g := &runner.Result{IsGRPC: true, GRPCStatus: 0, MessageJSON: []byte(`{"ok":true}`)}
	if cr := Check(&spec.Assert{GRPCStatus: intp(0)}, g, Env{}); !cr.OK {
		t.Errorf("grpc_status target should pass: %+v", cr)
	}
	if cr := Check(&spec.Assert{Message: &spec.StreamAssert{Contains: spec.StringList{`"ok":true`}}}, g, Env{}); !cr.OK {
		t.Errorf("message target should pass: %+v", cr)
	}
	// CDP value.
	cdp := &runner.Result{IsCDP: true, CDPValue: []byte("Ready")}
	if cr := Check(&spec.Assert{Value: &spec.StreamAssert{Equals: strp("Ready")}}, cdp, Env{}); !cr.OK {
		t.Errorf("value target should pass: %+v", cr)
	}
	// Screen.
	pty := &runner.Result{IsPTY: true, Screen: []byte("menu")}
	if cr := Check(&spec.Assert{Screen: &spec.StreamAssert{Contains: spec.StringList{"menu"}}}, pty, Env{}); !cr.OK {
		t.Errorf("screen target should pass: %+v", cr)
	}
	// Duration.
	dur := &runner.Result{Duration: 5 * time.Millisecond}
	if cr := Check(&spec.Assert{Duration: &spec.DurationAssert{LT: "1s"}}, dur, Env{}); !cr.OK {
		t.Errorf("duration target should pass: %+v", cr)
	}
	// Changes.
	ch := &runner.Result{Changes: &fsdelta.Delta{Created: []string{"out.txt"}}}
	if cr := Check(&spec.Assert{Changes: &spec.ChangesAssert{Created: list("out.txt")}}, ch, Env{}); !cr.OK {
		t.Errorf("changes target should pass: %+v", cr)
	}
	// Image + Dir + PDF + Mock via Check so their checkTarget cases run.
	wd := t.TempDir()
	writeImage(t, wd, "p.png", makePNG(t, 3, 3, color.RGBA{A: 255}))
	if cr := Check(&spec.Assert{Image: &spec.ImageAssert{Path: "p.png", Format: "png"}}, nil, Env{Workdir: wd}); !cr.OK {
		t.Errorf("image target should pass: %+v", cr)
	}
	if cr := Check(&spec.Assert{Dir: &spec.DirAssert{Path: ".", Exists: boolp(true)}}, nil, Env{Workdir: wd}); !cr.OK {
		t.Errorf("dir target should pass: %+v", cr)
	}
	pwd := writePDF(t, minimalPDF)
	if cr := Check(&spec.Assert{PDF: &spec.PDFAssert{Path: "doc.pdf", Pages: ptrInt(1)}}, nil, Env{Workdir: pwd}); !cr.OK {
		t.Errorf("pdf target should pass: %+v", cr)
	}
	menv := mockEnv(map[string][]mock.Record{"api": {{Method: "GET", Path: "/", Status: 200}}})
	if cr := Check(&spec.Assert{Mock: &spec.MockAssert{Name: "api"}}, nil, menv); !cr.OK {
		t.Errorf("mock target should pass: %+v", cr)
	}
}

// TestCheckDirRecursive_FailureBranches hits the recursive not_contains,
// file-count off-by-one, and glob-no-match failure branches (#25). Counts in
// recursive mode are over FILES only, so a directory does not inflate the count.
func TestCheckDirRecursive_FailureBranches(t *testing.T) {
	t.Parallel()
	wd := makeTree(t) // site: index.html, about.html, assets/app.css -> 3 files, 1 dir
	rec := func(d *spec.DirAssert) *CheckResult {
		d.Recursive = true
		return checkDirOK(t, wd, d)
	}
	// not_contains a path that is present -> fail.
	if cr := rec(&spec.DirAssert{Path: "site", NotContains: []string{"index.html"}}); cr.OK {
		t.Error("recursive not_contains of a present path must fail")
	}
	// Exactly 3 files (dirs excluded) passes; 4 fails; 2 fails.
	if cr := rec(&spec.DirAssert{Path: "site", Count: ptrInt(3)}); !cr.OK {
		t.Errorf("recursive count 3 files should pass: %+v", cr)
	}
	if cr := rec(&spec.DirAssert{Path: "site", Count: ptrInt(4)}); cr.OK {
		t.Error("recursive count 4 must fail (only 3 files, dir not counted)")
	}
	if cr := rec(&spec.DirAssert{Path: "site", MinCount: ptrInt(4)}); cr.OK {
		t.Error("recursive min 4 must fail")
	}
	if cr := rec(&spec.DirAssert{Path: "site", MaxCount: ptrInt(2)}); cr.OK {
		t.Error("recursive max 2 must fail")
	}
	// glob matching nothing in the tree -> fail.
	if cr := rec(&spec.DirAssert{Path: "site", Glob: "*.pdf"}); cr.OK {
		t.Error("recursive glob *.pdf must fail (no match)")
	}
	// glob matching a nested basename -> pass.
	if cr := rec(&spec.DirAssert{Path: "site", Glob: "*.css"}); !cr.OK {
		t.Errorf("recursive glob *.css should match assets/app.css: %+v", cr)
	}
}

// ---------------------------------------------------------------------------
// snapshot.go: missing, update, mismatch, path escape
// ---------------------------------------------------------------------------

func TestCheckSnapshot_Lifecycle(t *testing.T) {
	t.Parallel()
	specDir := t.TempDir()
	env := Env{SpecDir: specDir, Workdir: specDir}

	// Missing snapshot.
	miss := checkSnapshot("d", "stdout", "snap.txt", []byte("hello\n"), env)
	if miss.OK || !strings.Contains(miss.Actual, "missing") {
		t.Errorf("missing snapshot should fail with 'missing': %+v", miss)
	}

	// Update writes it.
	up := checkSnapshot("d", "stdout", "snap.txt", []byte("hello\n"), Env{SpecDir: specDir, Workdir: specDir, UpdateSnapshots: true})
	if !up.OK {
		t.Fatalf("update should pass: %+v", up)
	}
	if _, err := os.Stat(filepath.Join(specDir, "snap.txt")); err != nil {
		t.Fatalf("snapshot file not written: %v", err)
	}

	// Now a match passes.
	if cr := checkSnapshot("d", "stdout", "snap.txt", []byte("hello\n"), env); !cr.OK {
		t.Errorf("matching snapshot should pass: %+v", cr)
	}
	// And a mismatch fails with artifacts.
	mismatch := checkSnapshot("d", "stdout", "snap.txt", []byte("changed\n"), env)
	if mismatch.OK {
		t.Fatal("changed data should mismatch the snapshot")
	}
	if mismatch.ArtifactKind != "snapshot" {
		t.Errorf("ArtifactKind = %q, want snapshot", mismatch.ArtifactKind)
	}

	// A snapshot path escaping the spec dir is rejected.
	if cr := checkSnapshot("d", "stdout", "../escape.txt", []byte("x"), env); cr.OK {
		t.Error("a snapshot path escaping the spec dir must be rejected")
	}
}

// ---------------------------------------------------------------------------
// Fuzz-style smoke: feed detectImageFormat and parsePDF a batch of adversarial
// byte patterns and assert they never panic.
// ---------------------------------------------------------------------------

func TestParsersNeverPanicOnGarbage(t *testing.T) {
	t.Parallel()
	seeds := [][]byte{
		nil, {}, {0x00}, {0xff, 0xd8}, {'B', 'M'},
		[]byte("RIFF"), []byte("GIF8"), []byte("<svg"),
		bytes.Repeat([]byte{0xff}, 100),
		[]byte("\x89PNG\r\n\x1a\n"),
	}
	for i, s := range seeds {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on seed %d: %v", i, r)
				}
			}()
			_ = detectImageFormat(s)
			_ = isBMP(s)
			_ = isAVIF(s)
			_ = isSVG(s)
			_ = parsePDF(append([]byte("%PDF-1.4\n"), s...))
		}()
	}
}

// TestCheck_JSON_CompareLargeIntExact is a regression: the numeric bound
// matchers (gt/gte/lt/lte) must compare integers beyond 2^53 exactly, like
// equals already does, instead of collapsing distinct values to the same float64.
func TestCheck_JSON_CompareLargeIntExact(t *testing.T) {
	t.Parallel()
	// 9007199254740993 = 2^53 + 1; float64 rounds it down to 2^53.
	res := &runner.Result{Stdout: []byte(`{"big":9007199254740993}`)}
	tests := []struct {
		name   string
		j      *spec.JSONAssert
		wantOK bool
	}{
		{"gt its lower neighbor", &spec.JSONAssert{Path: "$.big", Gt: f64p(9007199254740992)}, true},
		{"not lte its lower neighbor", &spec.JSONAssert{Path: "$.big", Lte: f64p(9007199254740992)}, false},
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

// TestCheck_JSON_BoolNotEqualStringSpelling is a regression: a JSON boolean must
// not equal the string spelling of its value — `true` (bool) and "true" (string)
// are different JSON types. The old fmt.Sprintf fallback compared their printed
// forms and reported a false pass.
func TestCheck_JSON_BoolNotEqualStringSpelling(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte(`{"b":true}`)}
	// bool true must NOT equal the string "true".
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: &spec.JSONAssert{Path: "$.b", Equals: "true"}}}, res, Env{})
	if got.OK {
		t.Errorf("JSON boolean true wrongly equals the string \"true\"")
	}
	// bool true must still equal the boolean true.
	got = Check(&spec.Assert{Stdout: &spec.StreamAssert{JSON: &spec.JSONAssert{Path: "$.b", Equals: true}}}, res, Env{})
	if !got.OK {
		t.Errorf("JSON boolean true did not equal true (%s)", got.Hint)
	}
}
