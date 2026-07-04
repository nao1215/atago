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
