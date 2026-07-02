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
