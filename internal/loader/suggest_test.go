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
