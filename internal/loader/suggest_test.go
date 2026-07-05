package loader

import (
	"strings"
	"testing"
)

// TestLoadBytes_UnknownFieldSuggestion proves the first-five-minutes typo gets
// a did-you-mean hint, and a real field at the wrong nesting gets an
// indentation hint instead of a misleading suggestion.
func TestLoadBytes_UnknownFieldSuggestion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		src      string
		wantHint string
	}{
		{
			name:     "misspelled step key",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - asserts:\n          exit_code: 0",
			wantHint: `did you mean "assert"?`,
		},
		{
			name:     "misspelled stream target",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdut: {contains: hi}",
			wantHint: `did you mean "stdout"?`,
		},
		{
			name:     "valid field at the wrong nesting",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n        command: echo hi",
			wantHint: "check the indentation and nesting",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if err == nil {
				t.Fatal("LoadBytes() error = nil, want an unknown-field error")
			}
			if !strings.Contains(err.Error(), tt.wantHint) {
				t.Errorf("error = %q, want hint %q", err.Error(), tt.wantHint)
			}
		})
	}
}

// TestLoadBytes_BareScalarMatcherHint proves the single most common first-spec
// mistake — a bare scalar where a matcher mapping is required — earns a hint
// showing the accepted shape, while a legitimate scalar `body:` on an http step
// stays error-free.
func TestLoadBytes_BareScalarMatcherHint(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout: hello"
	_, err := LoadBytes("t.atago.yaml", []byte(src))
	if err == nil {
		t.Fatal("LoadBytes() error = nil, want a scalar-where-mapping error")
	}
	want := `stdout must set a matcher mapping, e.g. stdout: {contains: "..."}`
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want hint %q", err.Error(), want)
	}

	// An http step's body is a real scalar field: no error, no hint.
	ok := "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://127.0.0.1:1\nscenarios:\n  - name: a\n    steps:\n      - http:\n          runner: api\n          method: POST\n          path: /\n          body: hello"
	if _, err := LoadBytes("t.atago.yaml", []byte(ok)); err != nil {
		t.Errorf("http body scalar should load cleanly, got %v", err)
	}
}

// TestSuggest_NoWildGuess proves a token far from every field name gets no
// suggestion — a wrong hint is worse than none.
func TestSuggest_NoWildGuess(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - zzqqxx:\n          exit_code: 0"
	_, err := LoadBytes("t.atago.yaml", []byte(src))
	if err == nil {
		t.Fatal("LoadBytes() error = nil, want an unknown-field error")
	}
	if strings.Contains(err.Error(), "did you mean") {
		t.Errorf("error = %q; a distant token must not get a guess", err.Error())
	}
}

// TestEditDistance pins the Levenshtein helper that powers "did you mean X?".
// An off-by-one here would push a real typo just outside the suggestion window
// (or admit a wild guess), so the boundary values are load-bearing.
func TestEditDistance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"a", "", 1},
		{"kitten", "sitting", 3},
		{"stdout", "stdut", 1},
		{"assert", "asserts", 1},
		{"flaw", "lawn", 2},
	}
	for _, tt := range tests {
		if got := editDistance(tt.a, tt.b); got != tt.want {
			t.Errorf("editDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

// TestClosestField checks the suggestion picks the nearest field and declines to
// guess when nothing is close — and never suggests the same name back.
func TestClosestField(t *testing.T) {
	t.Parallel()
	tests := []struct {
		typo     string
		wantOK   bool
		wantName string
	}{
		{"stdut", true, "stdout"},
		{"asserts", true, "assert"},
		{"commnd", true, "command"},
		{"zzqqxxww", false, ""},
	}
	for _, tt := range tests {
		got, ok := closestField(tt.typo)
		if ok != tt.wantOK {
			t.Errorf("closestField(%q) ok = %v, want %v (got %q)", tt.typo, ok, tt.wantOK, got)
			continue
		}
		if ok && got != tt.wantName {
			t.Errorf("closestField(%q) = %q, want %q", tt.typo, got, tt.wantName)
		}
	}

	// A token that exactly matches an existing field must not suggest itself.
	if got, ok := closestField("command"); ok && got == "command" {
		t.Errorf("closestField(exact field) suggested itself: %q", got)
	}
}

// TestSuggestScalarMatcher_Passthrough proves the hint helpers leave unrelated
// messages untouched — a spurious hint is worse than none.
func TestSuggestScalarMatcher_Passthrough(t *testing.T) {
	t.Parallel()
	// No "used where mapping is expected" phrase: returned verbatim.
	if got := suggestScalarMatcher("some unrelated error"); got != "some unrelated error" {
		t.Errorf("suggestScalarMatcher passthrough = %q", got)
	}
	// The phrase is present but no stream target on the marked line: no hint.
	msg := "string was used where mapping is expected\n>  3 | foo: bar"
	if got := suggestScalarMatcher(msg); got != msg {
		t.Errorf("suggestScalarMatcher with no stream target added a hint: %q", got)
	}
	// A real stream target on the marked line gets the shape hint.
	msg2 := "string was used where mapping is expected\n>  9 |     stdout: hi"
	if got := suggestScalarMatcher(msg2); got == msg2 {
		t.Errorf("suggestScalarMatcher missed a stream target: %q", got)
	}
}

// TestSuggestUnknownField_Passthrough covers the non-field-name branch.
func TestSuggestUnknownField_Passthrough(t *testing.T) {
	t.Parallel()
	if got := suggestUnknownField("no field marker here"); got != "no field marker here" {
		t.Errorf("suggestUnknownField passthrough = %q", got)
	}
	// A distant unknown field yields no "did you mean" guess.
	if got := suggestUnknownField(`unknown field "zzqqxxww"`); got != `unknown field "zzqqxxww"` {
		t.Errorf("suggestUnknownField added a wild guess: %q", got)
	}
}

// TestSourcePos_UnknownAndNil exercises Source.pos's defensive branches: a nil
// receiver, a source built from unparseable bytes, a malformed path expression,
// and a path that resolves to no node — all must report the zero Position.
func TestSourcePos_UnknownAndNil(t *testing.T) {
	t.Parallel()

	// nil receiver / nil file.
	var nilSrc *Source
	if p := nilSrc.pos("$.suite"); p.Line != 0 {
		t.Errorf("nil source pos = %+v, want zero", p)
	}

	// Unparseable YAML: newSource yields a fileless Source that answers unknown.
	broken := newSource([]byte("a: b: c: ["))
	if broken == nil {
		t.Fatal("newSource returned nil")
	}
	if p := broken.pos("$.suite"); p.Line != 0 {
		t.Errorf("broken-source pos = %+v, want zero", p)
	}

	valid := newSource([]byte("suite:\n  name: x\n"))
	// A malformed path string resolves to unknown rather than erroring out.
	if p := valid.pos("$.["); p.Line != 0 {
		t.Errorf("malformed path pos = %+v, want zero", p)
	}
	// A path that names no node resolves to unknown.
	if p := valid.pos("$.nope.deeper"); p.Line != 0 {
		t.Errorf("missing-node pos = %+v, want zero", p)
	}
}

// TestSuitePos_Fallback proves SuitePos falls back to the `suite:` mapping when
// the suite has no name key (the primary `$.suite.name` lookup misses).
func TestSuitePos_Fallback(t *testing.T) {
	t.Parallel()
	// suite has no name — $.suite.name misses, fallback to $.suite, which (like
	// every mapping-node lookup here) resolves to the mapping's first field value
	// token: the `timeout: 1s` line (line 3).
	src := newSource([]byte("version: \"1\"\nsuite:\n  timeout: 1s\n"))
	line, _ := src.SuitePos()
	if line == 0 {
		t.Error("SuitePos() fell through to zero; expected the suite mapping line via fallback")
	}
	if line != 3 {
		t.Errorf("SuitePos() fallback line = %d, want 3", line)
	}
}
