package engine

import (
	"runtime"
	"strings"
	"testing"
)

// TestEngine_MatrixRunsEachRow checks that a matrix scenario expands and each
// instance resolves its own bound variables.
func TestEngine_MatrixRunsEachRow(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: m
scenarios:
  - name: "echoes ${word}"
    matrix:
      - { word: alpha }
      - { word: beta }
    steps:
      - run:
          shell: true
          command: echo ${word}
      - assert:
          stdout:
            contains: ${word}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
	if len(res.Scenarios) != 2 {
		t.Fatalf("scenarios = %d, want 2", len(res.Scenarios))
	}
	if res.Scenarios[0].Name != "echoes alpha" || res.Scenarios[1].Name != "echoes beta" {
		t.Errorf("names = %q, %q", res.Scenarios[0].Name, res.Scenarios[1].Name)
	}
}

// TestEngine_RetryUntilPasses checks that retry stops once until is satisfied.
// The marker file makes the second attempt produce different output.
func TestEngine_RetryUntilPasses(t *testing.T) {
	t.Parallel()
	// First attempt creates a marker and reports waiting; the second sees the
	// marker and reports ready.
	probe := "if [ -f marker ]; then echo ready; else touch marker; echo waiting; fi"
	if runtime.GOOS == "windows" {
		probe = "if exist marker (echo ready) else (echo waiting& type nul >marker)"
	}
	res := runSpec(t, `
version: "1"
suite:
  name: r
scenarios:
  - name: polls until ready
    steps:
      - run:
          shell: true
          command: "`+probe+`"
          retry:
            times: 5
            interval: 5ms
            until:
              stdout: {contains: ready}
      - assert:
          stdout: {contains: ready}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_RetryUntilFails marks the scenario failed when until never holds.
func TestEngine_RetryUntilFails(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: r
scenarios:
  - name: never ready
    steps:
      - run:
          shell: true
          command: echo waiting
          retry:
            times: 3
            interval: 1ms
            until:
              stdout: {contains: ready}
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_UnresolvedVarWithoutShellErrors: a ${name} nothing defines, in a
// local run without shell: true, is an explained step error — no shell could
// ever expand it, so passing the literal text to argv is always a typo.
func TestEngine_UnresolvedVarWithoutShellErrors(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: typo in a variable name
    steps:
      - run: {command: "echo ${no_such_var}"}
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	msg := res.Scenarios[0].Steps[0].ErrMsg
	if !strings.Contains(msg, "${no_such_var}") || !strings.Contains(msg, "shell") {
		t.Errorf("error should name the reference and mention the shell option, got %q", msg)
	}
}

// TestEngine_UnresolvedVarWithShellIsLeftToTheShell: with shell: true the same
// reference is legitimate shell syntax and must not be rejected.
func TestEngine_UnresolvedVarWithShellIsLeftToTheShell(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: shell variable expansion
    steps:
      - run: {shell: true, command: "echo ${PATH}"}
      - assert: {exit_code: 0}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_UnresolvedVarInCwdErrors: a ${name} nothing defines in a run step's
// cwd is an explained error naming the reference, not the misleading
// "executable not found" the child raises when it cannot start in a literal
// "${name}" directory. Go sets cmd.Dir verbatim, so no shell ever expands it.
func TestEngine_UnresolvedVarInCwdErrors(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: v
scenarios:
  - name: typo in cwd
    steps:
      - run: {command: "pwd", cwd: "${no_such_dir}"}
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	msg := res.Scenarios[0].Steps[0].ErrMsg
	if !strings.Contains(msg, "${no_such_dir}") || !strings.Contains(msg, "cwd") {
		t.Errorf("error should name the cwd reference, got %q", msg)
	}
	if strings.Contains(msg, "fork/exec") {
		t.Errorf("error should not be the misleading exec failure, got %q", msg)
	}
}

// TestEngine_SkipByCommand exercises the probe-command skip predicate.
func TestEngine_SkipByCommand(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: c
scenarios:
  - name: skipped when probe succeeds
    skip:
      command: "exit 0"
    steps:
      - run: {shell: true, command: echo nope}
  - name: only when probe succeeds
    only:
      command: "exit 0"
    steps:
      - run: {shell: true, command: echo yes}
      - assert: {stdout: {contains: yes}}
  - name: only skipped when probe fails
    only:
      command: "exit 1"
    steps:
      - run: {shell: true, command: echo nope}
`)
	if res.Scenarios[0].Status != StatusSkipped {
		t.Errorf("scenario[0] = %s, want skipped (skip command succeeded)", res.Scenarios[0].Status)
	}
	if res.Scenarios[1].Status != StatusPassed {
		t.Errorf("scenario[1] = %s, want passed (only command succeeded)", res.Scenarios[1].Status)
	}
	if res.Scenarios[2].Status != StatusSkipped {
		t.Errorf("scenario[2] = %s, want skipped (only command failed)", res.Scenarios[2].Status)
	}
}
