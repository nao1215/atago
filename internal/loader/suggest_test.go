package loader

import (
	"slices"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/yaml"
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
// TestLoadBytes_CollectionShapeHint pins the hint for a list or a mapping
// written where one value belongs. The decoder's own message names a Go type
// and the outermost struct field ("Spec.Scenarios of type string"), which is
// true and useless: it names neither the key the author wrote nor the shape it
// wants, so `matches: [a, b]` — the natural mistake next to `contains:`, which
// does take a list — read as a problem with `scenarios`.
func TestLoadBytes_CollectionShapeHint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		src       string
		wantHints []string
	}{
		{
			name:      "block list on a single-pattern matcher",
			src:       "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout:\n            matches:\n              - \"^h\"\n              - \"i$\"\n",
			wantHints: []string{`"matches" takes a single value, not a list`, "one assert per pattern", `"contains"`},
		},
		{
			name:      "flow list on a single-pattern matcher",
			src:       "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert: {stdout: {matches: [\"^h\", \"i$\"]}}\n",
			wantHints: []string{`"matches" takes a single value, not a list`},
		},
		{
			name:      "mapping under a key that takes text",
			src:       "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command:\n            bin: echo\n",
			wantHints: []string{`"command" takes a single value, not a mapping`},
		},
		{
			name:      "a list where a duration belongs",
			src:       "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          timeout:\n            - 1s\n            - 2s\n",
			wantHints: []string{`"timeout" takes a single value, not a list`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if err == nil {
				t.Fatal("LoadBytes() error = nil, want a decode error")
			}
			for _, want := range tt.wantHints {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %v, want it to contain %q", err, want)
				}
			}
			// The Go type name is the decoder's, and it stays in the message;
			// what must not happen is the hint blaming the outer field.
			if strings.Contains(err.Error(), `hint: "scenarios"`) {
				t.Errorf("hint blames the outermost field instead of the key at the marker: %v", err)
			}
		})
	}
}

// TestSuggestShape_Passthrough proves the hint is attached only to the error
// it explains: an error at a list index or at the document, and a scalar of
// the wrong kind, come back untouched rather than gaining a guess.
func TestSuggestShape_Passthrough(t *testing.T) {
	t.Parallel()
	for _, e := range []*yaml.Error{
		{Msg: "m", Path: "", Want: "a string", Got: yaml.SequenceNode},
		{Msg: "m", Path: "scenarios[0]", Want: "a string", Got: yaml.SequenceNode},
		{Msg: "m", Path: "suite.timeout", Want: "a string", Got: yaml.ScalarNode},
		{Msg: "m", Path: "suite.name", Want: "a list", Got: yaml.MappingNode},
	} {
		if got := suggestShape("msg", e); got != "msg" {
			t.Errorf("suggestShape(%+v) = %q, want it unchanged", e, got)
		}
	}
}

// TestLastKey pins how the key a hint names is read off an error path.
func TestLastKey(t *testing.T) {
	t.Parallel()
	for path, want := range map[string]string{
		"":                                  "",
		"suite":                             "suite",
		"scenarios[0].steps[1].run.command": "command",
		"scenarios[0]":                      "",
		`env."A.B"`:                         "A.B",
	} {
		if got := lastKey(path); got != want {
			t.Errorf("lastKey(%q) = %q, want %q", path, got, want)
		}
	}
}

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

// TestClosestField_TieIsDeterministic pins that a typo equally close to
// several fields gets the same hint in every process. The vocabulary is
// collected into a map, and a hint picked in map order changed from run to
// run ("lt", then "fg", then "gt" for the same spec), which `--ci` promises
// not to do. A tie now goes to the alphabetically first field.
func TestClosestField_TieIsDeterministic(t *testing.T) {
	t.Parallel()
	vocab := fieldVocabulary()
	if !slices.IsSorted(vocab) {
		t.Fatal("the field vocabulary is not sorted, so a tie is resolved in map order")
	}
	const typo = "zz"
	best, want := 3, ""
	ties := 0
	for _, name := range vocab {
		switch d := editDistance(typo, name); {
		case d < best:
			best, want, ties = d, name, 1
		case d == best:
			ties++
		}
	}
	if ties < 2 {
		t.Fatalf("%q is not a tie (%d candidate at distance %d); pick a typo several fields are equally close to", typo, ties, best)
	}
	if got, _ := closestField(typo); got != want {
		t.Errorf("closestField(%q) = %q, want the alphabetically first of the %d tied fields, %q", typo, got, ties, want)
	}
}

// TestSuggestUnknownField_Passthrough proves a distant unknown field yields no
// "did you mean" guess.
func TestSuggestUnknownField_Passthrough(t *testing.T) {
	t.Parallel()
	if got := suggestUnknownField(`unknown field "zzqqxxww"`, "zzqqxxww"); got != `unknown field "zzqqxxww"` {
		t.Errorf("suggestUnknownField added a wild guess: %q", got)
	}
}

// TestSourcePos_UnknownAndNil exercises the locator's defensive branches: a nil
// receiver, a source built from unparseable bytes, and lookups that name no
// node all report the zero position.
func TestSourcePos_UnknownAndNil(t *testing.T) {
	t.Parallel()
	var nilSrc *Source
	if l, c := nilSrc.SuitePos(); l != 0 || c != 0 {
		t.Errorf("nil source SuitePos = %d:%d, want zero", l, c)
	}
	broken := newSource([]byte("a: b: c: ["))
	if broken == nil {
		t.Fatal("newSource returned nil")
	}
	if l, _ := broken.SuitePos(); l != 0 {
		t.Errorf("broken-source SuitePos line = %d, want zero", l)
	}
	valid := newSource([]byte("suite:\n  name: x\n"))
	if l, _ := valid.RunnerPos("nope"); l != 0 {
		t.Errorf("missing runner line = %d, want zero", l)
	}
	if l, _ := valid.ScenarioPos(3); l != 0 {
		t.Errorf("missing scenario line = %d, want zero", l)
	}
	if l, _ := valid.StepPos(0, 0); l != 0 {
		t.Errorf("missing step line = %d, want zero", l)
	}
}

// TestSuitePos_Fallback proves SuitePos falls back to the `suite:` mapping when
// the suite has no name key (the primary `$.suite.name` lookup misses).
func TestSuitePos_Fallback(t *testing.T) {
	t.Parallel()
	// suite has no name, so the lookup falls back to the suite mapping, which
	// starts at its first key: the `timeout: 1s` line (line 3).
	src := newSource([]byte("version: \"1\"\nsuite:\n  timeout: 1s\n"))
	line, _ := src.SuitePos()
	if line == 0 {
		t.Error("SuitePos() fell through to zero; expected the suite mapping line via fallback")
	}
	if line != 3 {
		t.Errorf("SuitePos() fallback line = %d, want 3", line)
	}
}
