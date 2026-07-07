package loader

import (
	"strings"
	"testing"
)

// TestLoadBytes_ContainsScalarAndList proves `contains` / `not_contains` accept
// both a scalar string (backward compatible) and a list of strings, decoding to
// a one- or many-element StringList.
func TestLoadBytes_ContainsScalarAndList(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            contains: hi
      - assert:
          stdout:
            contains: [hi, there]
      - assert:
          file:
            path: out.txt
            not_contains: [boom, kaboom]
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	steps := s.Scenarios[0].Steps
	if got := steps[1].Assert.Stdout.Contains; len(got) != 1 || got[0] != "hi" {
		t.Errorf("scalar contains decoded as %v, want [hi]", got)
	}
	if got := steps[2].Assert.Stdout.Contains; len(got) != 2 || got[0] != "hi" || got[1] != "there" {
		t.Errorf("list contains decoded as %v, want [hi there]", got)
	}
	if got := steps[3].Assert.File.NotContains; len(got) != 2 || got[1] != "kaboom" {
		t.Errorf("list not_contains decoded as %v, want [boom kaboom]", got)
	}
}

// TestLoadBytes_MultiTargetAssert proves one assert may set several targets and
// that each target is still shape-validated.
func TestLoadBytes_MultiTargetAssert(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          exit_code: 0
          stdout:
            contains: hi
          file:
            path: out.txt
            exists: true
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	a := s.Scenarios[0].Steps[1].Assert
	if len(a.SetTargets()) != 3 {
		t.Fatalf("expected 3 targets, got %v", a.SetTargets())
	}

	// A bad shape on one of several targets is still reported.
	bad := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          exit_code: 0
          stdout:
            contains: hi
            equals: also
`
	if _, err := LoadBytes("sample.atago.yaml", []byte(bad)); err == nil {
		t.Fatal("expected a validation error for stdout setting two matchers, got nil")
	}
}

// TestLoadBytes_ContainsEmptyListRejected proves an explicitly-empty list is a
// validation error rather than a trivially-passing matcher.
func TestLoadBytes_ContainsEmptyListRejected(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            contains: []
`
	_, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err == nil || !strings.Contains(err.Error(), "contains must not be empty") {
		t.Fatalf("LoadBytes() error = %v, want a 'contains must not be empty' validation error", err)
	}
}

// TestLoadBytes_EmptyStringElementRejected proves an empty-string element in a
// contains/not_contains matcher is a validation error: `contains ""` is an
// always-true no-op and `not_contains ""` can never pass, so both are authoring
// mistakes worth catching at load time rather than at run time.
func TestLoadBytes_EmptyStringElementRejected(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"contains", "not_contains"} {
		src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            ` + key + `: ""
`
		_, err := LoadBytes("sample.atago.yaml", []byte(src))
		if err == nil || !strings.Contains(err.Error(), "empty string") {
			t.Fatalf("%s: LoadBytes() error = %v, want an 'empty string' validation error", key, err)
		}
	}
}

// TestLoadBytes_EmptyRegexpRejected proves an empty matches/not_matches pattern
// is a validation error, mirroring the empty-string contains/not_contains case:
// an empty regexp matches everything, so `matches: ""` is an always-true no-op
// and `not_matches: ""` can never pass. Covered for the stream, json, and
// header matchers that expose a regexp field.
func TestLoadBytes_EmptyRegexpRejected(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		matcher string
	}{
		{"stream matches", `            matches: ""`},
		{"stream not_matches", `            not_matches: ""`},
		{"json matches", "            json:\n              path: $.x\n              matches: \"\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
` + tc.matcher + "\n"
			_, err := LoadBytes("sample.atago.yaml", []byte(src))
			if err == nil || !strings.Contains(err.Error(), "empty regexp") {
				t.Fatalf("LoadBytes() error = %v, want an 'empty regexp' validation error", err)
			}
		})
	}
}

// TestLoadBytes_EmptyMatchingNotMatchesRejected proves that a not_matches pattern
// that matches the empty string (e.g. "z*") is a load-time validation error:
// such a pattern matches at position 0 of every input, so not_matches can never
// pass — the same trap as an empty pattern. The identical pattern under matches
// is legitimate (it matches everything, which the author may intend) and still
// loads, and a pattern that requires at least one character (e.g. "z+") loads too.
func TestLoadBytes_EmptyMatchingNotMatchesRejected(t *testing.T) {
	t.Parallel()
	build := func(matcher string) []byte {
		return []byte(`
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
` + matcher + "\n")
	}
	t.Run("not_matches empty-matching rejected", func(t *testing.T) {
		t.Parallel()
		_, err := LoadBytes("sample.atago.yaml", build(`            not_matches: "z*"`))
		if err == nil || !strings.Contains(err.Error(), "matches the empty string") {
			t.Fatalf("LoadBytes() error = %v, want an empty-string not_matches error", err)
		}
	})
	t.Run("matches empty-matching still loads", func(t *testing.T) {
		t.Parallel()
		if _, err := LoadBytes("sample.atago.yaml", build(`            matches: "z*"`)); err != nil {
			t.Fatalf("LoadBytes() error = %v, want nil (an empty-matching matches is legitimate)", err)
		}
	})
	t.Run("not_matches requiring a character still loads", func(t *testing.T) {
		t.Parallel()
		if _, err := LoadBytes("sample.atago.yaml", build(`            not_matches: "z+"`)); err != nil {
			t.Fatalf("LoadBytes() error = %v, want nil (z+ needs at least one character)", err)
		}
	})
}

// TestLoadBytes_NegativeJSONLengthRejected proves a negative json/yaml length is
// a validation error: no array, object, or string has a negative length, so the
// matcher could never pass and the mistake is caught at load time.
func TestLoadBytes_NegativeJSONLengthRejected(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            json:
              path: $.items
              length: -1
`
	_, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err == nil || !strings.Contains(err.Error(), "length must be >= 0") {
		t.Fatalf("LoadBytes() error = %v, want a 'length must be >= 0' validation error", err)
	}
}

// TestLoadBytes_ShellMetacharWithoutShell proves a command carrying shell syntax
// is rejected when shell is not enabled, with a fix-forward hint. Each metachar
// in the guarded set is exercised.
func TestLoadBytes_ShellMetacharWithoutShell(t *testing.T) {
	t.Parallel()
	for _, cmd := range []string{
		"echo hi > out.txt",
		"echo hi >> out.txt",
		"cat < in.txt",
		"echo hi | wc -l",
		"a || b",
		"a && b",
		"echo hi; echo bye",
		"echo $(date)",
		"echo `date`",
	} {
		src := "" +
			"version: \"1\"\n" +
			"suite:\n  name: sample\n" +
			"scenarios:\n  - name: ok\n    steps:\n" +
			"      - run: {command: \"" + cmd + "\"}\n"
		_, err := LoadBytes("sample.atago.yaml", []byte(src))
		if err == nil {
			t.Errorf("command %q: expected a validation error, got nil", cmd)
			continue
		}
		msg := err.Error()
		if !strings.Contains(msg, "shell is not enabled") ||
			!strings.Contains(msg, "shell: true") ||
			!strings.Contains(msg, "stdout_to") {
			t.Errorf("command %q: error %q missing the expected guidance", cmd, msg)
		}
	}
}

// TestLoadBytes_ShellMetacharAllowedWhenShellOrQuoted proves the guard does not
// fire when shell is enabled, and does not false-positive on a quoted metachar.
func TestLoadBytes_ShellMetacharAllowedWhenShellOrQuoted(t *testing.T) {
	t.Parallel()
	cases := []string{
		"      - run: {shell: true, command: \"echo hi > out.txt\"}\n",
		"      - run: {command: \"grep '>>' in.txt\"}\n",
	}
	for _, step := range cases {
		src := "" +
			"version: \"1\"\n" +
			"suite:\n  name: sample\n" +
			"scenarios:\n  - name: ok\n    steps:\n" + step
		if _, err := LoadBytes("sample.atago.yaml", []byte(src)); err != nil {
			t.Errorf("step %q: unexpected error %v", step, err)
		}
	}
}

// TestLoadBytes_StdoutToStderrTo proves the structured redirect keys decode on a
// non-shell run step.
func TestLoadBytes_StdoutToStderrTo(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run:
          command: echo hi
          stdout_to: out.txt
          stderr_to: err.txt
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	run := s.Scenarios[0].Steps[0].Run
	if run.StdoutTo != "out.txt" || run.StderrTo != "err.txt" {
		t.Errorf("stdout_to/stderr_to decoded as %q/%q, want out.txt/err.txt", run.StdoutTo, run.StderrTo)
	}
}

// TestLoadBytes_MockServerRejectedInScenario proves a mock_server step is
// rejected outside suite.setup, like a service step. It is a suite-setup-only
// action, but validateStep (the scenario steps/teardown path) had no case for
// it, so it was silently accepted and never started.
func TestLoadBytes_MockServerRejectedInScenario(t *testing.T) {
	t.Parallel()
	specs := map[string]string{
		"steps": `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - mock_server:
          name: api
          routes:
            - {method: GET, path: /}
`,
		"teardown": `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
    teardown:
      - mock_server:
          name: api
          routes:
            - {method: GET, path: /}
`,
	}
	for block, src := range specs {
		_, err := LoadBytes("sample.atago.yaml", []byte(src))
		if err == nil {
			t.Fatalf("%s: LoadBytes() error = nil, want a 'mock_server steps are only allowed in suite.setup' error", block)
		}
		if !strings.Contains(err.Error(), "mock_server") || !strings.Contains(err.Error(), "suite.setup") {
			t.Fatalf("%s: LoadBytes() error = %v, want it to mention mock_server and suite.setup", block, err)
		}
	}
}

// TestLoadBytes_ExitCodeUnknownKeyRejected proves the exit_code mapping form is
// decoded strictly: an unknown key is an authoring typo and must be rejected,
// matching the loader's yaml.Strict() decode for the rest of the document.
func TestLoadBytes_ExitCodeUnknownKeyRejected(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          exit_code: {not: 0, bogus: 5}
`
	_, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err == nil {
		t.Fatalf("LoadBytes() error = nil, want an unknown-field error for exit_code.bogus")
	}
}

// TestLoadBytes_DirGlobValidated proves a malformed dir.glob pattern is rejected
// at load time, like the changes globs and dir.ignore patterns already are,
// instead of only misbehaving at check time.
func TestLoadBytes_DirGlobValidated(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
      - assert:
          dir:
            path: out
            glob: "a["
`
	_, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err == nil {
		t.Fatalf("LoadBytes() error = nil, want a 'not a valid glob' error for dir.glob")
	}
	if !strings.Contains(err.Error(), "glob") {
		t.Fatalf("LoadBytes() error = %v, want it to mention the invalid glob", err)
	}
}

// TestLoadBytes_NegativeDurationRejected proves a negative timeout is rejected at
// load time. A negative wall-clock bound is never meaningful — the same rule
// validatePTY and validateSignal already enforce — and a negative step timeout
// produces an already-expired context that kills the step immediately.
func TestLoadBytes_NegativeDurationRejected(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"suite.timeout": `
version: "1"
suite:
  name: sample
  timeout: "-5s"
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
`,
		"run.timeout": `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi, timeout: "-5s"}
`,
		"defaults.run.timeout": `
version: "1"
suite:
  name: sample
defaults:
  run:
    timeout: "-5s"
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
`,
	}
	for name, src := range cases {
		_, err := LoadBytes("sample.atago.yaml", []byte(src))
		if err == nil {
			t.Fatalf("%s: LoadBytes() error = nil, want a negative-duration validation error", name)
		}
	}
}
