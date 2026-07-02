package loader

import (
	"strings"
	"testing"
)

// TestApplyDefaults_RunAndScenarioEnv proves defaults.run and
// defaults.scenario.env expand into the concrete model: shell fills when unset, scalar
// fields fill only when unset, and env shallow-merges with the step/scenario
// value winning per key.
func TestApplyDefaults_RunAndScenarioEnv(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  run:
    shell: true
    timeout: 5s
    env:
      A: from-default
      B: from-default
  scenario:
    env:
      HOME: ${workdir}/home
      SHARED: default
scenarios:
  - name: inherits and overrides
    env:
      SHARED: own
    steps:
      - run:
          command: echo hi
          timeout: 9s
          env:
            B: own
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	sc := s.Scenarios[0]
	if sc.Env["HOME"] != "${workdir}/home" {
		t.Errorf("scenario env HOME = %q, want default applied", sc.Env["HOME"])
	}
	if sc.Env["SHARED"] != "own" {
		t.Errorf("scenario env SHARED = %q, want scenario value to win", sc.Env["SHARED"])
	}
	run := sc.Steps[0].Run
	if !run.ShellEnabled() {
		t.Errorf("run.shell = false, want default applied")
	}
	if run.Timeout != "9s" {
		t.Errorf("run.timeout = %q, want explicit 9s to win", run.Timeout)
	}
	if run.Env["A"] != "from-default" {
		t.Errorf("run.env[A] = %q, want default applied", run.Env["A"])
	}
	if run.Env["B"] != "own" {
		t.Errorf("run.env[B] = %q, want step value to win", run.Env["B"])
	}
}

// TestApplyDefaults_ExplicitShellFalseWins proves an authored `shell: false`
// beats a defaulted `shell: true` — the documented "an explicitly authored
// value always wins" rule holds for booleans too (Shell is a *bool so unset
// and false stay distinct).
func TestApplyDefaults_ExplicitShellFalseWins(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  run:
    shell: true
  service:
    shell: true
scenarios:
  - name: opts out per element
    services:
      - name: mock
        shell: false
        command: ./mock
    steps:
      - run:
          shell: false
          command: echo hi
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if s.Scenarios[0].Steps[0].Run.ShellEnabled() {
		t.Errorf("run.shell = true, want the authored false to win over the default")
	}
	if s.Scenarios[0].Services[0].ShellEnabled() {
		t.Errorf("service.shell = true, want the authored false to win over the default")
	}
}

// TestApplyDefaults_Service proves defaults.service fills shell/env and copies a
// whole ready probe into a service that declares none, while a service with its
// own ready keeps it.
func TestApplyDefaults_Service(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  service:
    shell: true
    ready:
      file: ready.txt
      store: addr
      timeout: 5s
scenarios:
  - name: uses default ready
    services:
      - name: mock
        command: ./mock --ready-file ready.txt
    steps:
      - run:
          command: echo ${addr}
  - name: keeps its own ready
    services:
      - name: mock
        command: ./mock
        ready:
          port: 127.0.0.1:9000
    steps:
      - run:
          command: echo hi
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	svc := s.Scenarios[0].Services[0]
	if !svc.ShellEnabled() {
		t.Errorf("service.shell = false, want default applied")
	}
	if svc.Ready == nil || svc.Ready.File != "ready.txt" || svc.Ready.Store != "addr" {
		t.Fatalf("service.ready = %+v, want default probe copied in", svc.Ready)
	}
	own := s.Scenarios[1].Services[0]
	if own.Ready == nil || own.Ready.Port != "127.0.0.1:9000" || own.Ready.File != "" {
		t.Errorf("service.ready = %+v, want the service's own probe kept", own.Ready)
	}
}

// TestApplyDefaults_MatrixInstancesInherit proves defaults are applied to every
// concrete scenario produced by matrix expansion (applyDefaults runs after
// expandMatrix).
func TestApplyDefaults_MatrixInstancesInherit(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  run:
    shell: true
scenarios:
  - name: "greets ${who}"
    matrix:
      - { who: Alice }
      - { who: Bob }
    steps:
      - run:
          command: echo ${who}
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if len(s.Scenarios) != 2 {
		t.Fatalf("got %d scenarios, want 2 matrix instances", len(s.Scenarios))
	}
	for _, sc := range s.Scenarios {
		if !sc.Steps[0].Run.ShellEnabled() {
			t.Errorf("scenario %q: run.shell = false, want default applied", sc.Name)
		}
	}
}

// TestValidateDefaults_RejectsUnsupportedFields proves fields the loader would
// silently ignore are load-time errors instead of no-ops.
func TestValidateDefaults_RejectsUnsupportedFields(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"defaults.run.command":     "run:\n    command: echo hi\n",
		"defaults.run.retry":       "run:\n    retry:\n      times: 2\n      until:\n        exit_code: 0\n",
		"defaults.service.name":    "service:\n    name: mock\n",
		"defaults.service.command": "service:\n    command: ./mock\n",
	}
	for want, frag := range cases {
		src := "version: \"1\"\nsuite:\n  name: sample\ndefaults:\n  " + frag +
			"scenarios:\n  - name: ok\n    steps:\n      - run:\n          command: echo hi\n"
		_, err := LoadBytes("sample.atago.yaml", []byte(src))
		if err == nil {
			t.Errorf("%s: expected a validation error, got nil", want)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%s: error = %v, want it to mention %q", want, err, want)
		}
	}
}

// TestApplyDefaults_UnknownKeyRejected proves an unknown key under defaults is a
// load error (strict decode), matching the merge-rule contract.
func TestApplyDefaults_UnknownKeyRejected(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  bogus: true
scenarios:
  - name: ok
    steps:
      - run:
          command: echo hi
`
	if _, err := LoadBytes("sample.atago.yaml", []byte(src)); err == nil {
		t.Fatal("expected an error for an unknown defaults key, got nil")
	}
}

// TestApplyDefaults_AbsentIsNoOp proves a spec without defaults is unchanged
// (backward compatibility).
func TestApplyDefaults_AbsentIsNoOp(t *testing.T) {
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
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if s.Defaults != nil {
		t.Errorf("Defaults = %+v, want nil", s.Defaults)
	}
	if s.Scenarios[0].Steps[0].Run.ShellEnabled() {
		t.Errorf("run.shell = true, want unchanged false without defaults")
	}
}
