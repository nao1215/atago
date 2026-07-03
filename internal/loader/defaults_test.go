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

// TestApplyDefaults_ClearEnvPassEnv proves defaults.run.clear_env /
// defaults.run.pass_env layer into steps under the authored-value-wins rule,
// and the same for defaults.service (#16).
func TestApplyDefaults_ClearEnvPassEnv(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  run:
    clear_env: true
    pass_env: [PATH, HOME]
  service:
    clear_env: true
    pass_env: [PATH]
scenarios:
  - name: hermetic by default
    services:
      - name: mock
        command: ./mock
    steps:
      - run:
          command: echo hi
      - run:
          command: echo bye
          clear_env: false
      - run:
          command: echo own
          pass_env: [LANG]
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	sc := s.Scenarios[0]
	if !sc.Steps[0].Run.ClearEnvEnabled() {
		t.Error("step 0: clear_env default not applied")
	}
	if got := sc.Steps[0].Run.PassEnv; len(got) != 2 || got[0] != "PATH" || got[1] != "HOME" {
		t.Errorf("step 0: pass_env = %v, want default [PATH HOME]", got)
	}
	if sc.Steps[1].Run.ClearEnvEnabled() {
		t.Error("step 1: authored clear_env: false must win over the default")
	}
	if got := sc.Steps[1].Run.PassEnv; len(got) != 0 {
		t.Errorf("step 1: pass_env = %v, want none (pass_env is not inherited when clear_env is off)", got)
	}
	if got := sc.Steps[2].Run.PassEnv; len(got) != 1 || got[0] != "LANG" {
		t.Errorf("step 2: pass_env = %v, want the authored [LANG] to win", got)
	}
	if !sc.Services[0].ClearEnvEnabled() {
		t.Error("service: clear_env default not applied")
	}
	if got := sc.Services[0].PassEnv; len(got) != 1 || got[0] != "PATH" {
		t.Errorf("service: pass_env = %v, want default [PATH]", got)
	}
}

// TestValidate_PassEnvRequiresClearEnv proves pass_env without clear_env: true
// is a load-time validation error with a positioned message (#16).
func TestValidate_PassEnvRequiresClearEnv(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, src, wantPos string
	}{
		{
			name: "run step",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
          pass_env: [PATH]
`,
			wantPos: `scenario "s".steps[0].run`,
		},
		{
			name: "scenario service",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    services:
      - name: mock
        command: ./mock
        pass_env: [PATH]
    steps:
      - run:
          command: echo hi
`,
			wantPos: `service "mock"`,
		},
		{
			name: "pty step",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - pty:
          command: cat
          pass_env: [PATH]
`,
			wantPos: `scenario "s".steps[0].pty`,
		},
		{
			name: "suite setup service",
			src: `
version: "1"
suite:
  name: sample
  setup:
    - service:
        name: shared
        command: ./mock
        pass_env: [PATH]
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
`,
			wantPos: "suite.setup[0].service",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("sample.atago.yaml", []byte(tc.src))
			if err == nil {
				t.Fatal("LoadBytes() = nil error, want a pass_env-requires-clear_env validation error")
			}
			if !strings.Contains(err.Error(), "pass_env") || !strings.Contains(err.Error(), "clear_env") {
				t.Errorf("error = %q, want it to mention pass_env requiring clear_env", err)
			}
			if !strings.Contains(err.Error(), tc.wantPos) {
				t.Errorf("error = %q, want position %q", err, tc.wantPos)
			}
		})
	}
}

// TestValidate_PassEnvEmptyEntryRejected proves an empty variable name inside
// pass_env is rejected at load time (#16).
func TestValidate_PassEnvEmptyEntryRejected(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
          clear_env: true
          pass_env: ["PATH", ""]
`
	_, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err == nil {
		t.Fatal("LoadBytes() = nil error, want an empty-pass_env-entry validation error")
	}
	if !strings.Contains(err.Error(), "pass_env") {
		t.Errorf("error = %q, want it to mention pass_env", err)
	}
}

// TestValidate_StdinShapes proves the stdin one-of rule and base64 validation
// fire at load time with positioned messages (#18).
func TestValidate_StdinShapes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, src, wantMsg string
	}{
		{
			name: "empty mapping",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: cat
          stdin: {}
`,
			wantMsg: "exactly one of file/base64",
		},
		{
			name: "both file and base64",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: cat
          stdin: {file: in.txt, base64: AAEC}
`,
			wantMsg: "exactly one of file/base64",
		},
		{
			name: "invalid base64",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: cat
          stdin: {base64: "not!!base64"}
`,
			wantMsg: "not valid base64",
		},
		{
			name: "unknown stdin key",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: cat
          stdin: {fil: in.txt}
`,
			wantMsg: "accepted: file, base64",
		},
		{
			name: "defaults.run.stdin rejected",
			src: `
version: "1"
suite:
  name: sample
defaults:
  run:
    stdin: "shared input"
scenarios:
  - name: s
    steps:
      - run:
          command: cat
`,
			wantMsg: "defaults.run.stdin is not supported",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("sample.atago.yaml", []byte(tc.src))
			if err == nil {
				t.Fatal("LoadBytes() = nil error, want a stdin validation error")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error = %q, want substring %q", err, tc.wantMsg)
			}
		})
	}
}

// TestValidate_ExitCodeIn proves the exit_code in-set rules (#19): one-of
// violations, an empty set, duplicate values, and non-int entries all fail at
// load time.
func TestValidate_ExitCodeIn(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, src, wantMsg string
	}{
		{
			name: "not and in together",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: {not: 1, in: [0, 2]}
`,
			wantMsg: "exactly one of",
		},
		{
			name: "empty in list",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: {in: []}
`,
			wantMsg: "at least one accepted exit code",
		},
		{
			name: "duplicate values",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: {in: [0, 2, 0]}
`,
			wantMsg: "more than once",
		},
		{
			name: "non-int entry",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: {in: [0, ok]}
`,
			wantMsg: "exit_code",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("sample.atago.yaml", []byte(tc.src))
			if err == nil {
				t.Fatal("LoadBytes() = nil error, want an exit_code validation error")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error = %q, want substring %q", err, tc.wantMsg)
			}
		})
	}
}

// TestValidate_MockServers proves the mock-server rules (#24): duplicate
// names, route payload one-of, unknown mock names in asserts (listing the
// declared ones), and count-0-with-matchers contradictions all fail at load.
func TestValidate_MockServers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, src, wantMsg string
	}{
		{
			name: "duplicate mock names",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    mock_servers:
      - name: api
        routes: [{method: GET, path: /}]
      - name: api
        routes: [{method: GET, path: /}]
    steps:
      - run: {command: echo hi}
`,
			wantMsg: `duplicate mock server name "api"`,
		},
		{
			name: "route with two payload sources",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    mock_servers:
      - name: api
        routes:
          - method: GET
            path: /
            body: inline
            body_file: canned.json
    steps:
      - run: {command: echo hi}
`,
			wantMsg: "at most one of json/body/body_file",
		},
		{
			name: "unknown mock name in assert lists declared",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    mock_servers:
      - name: api
        routes: [{method: GET, path: /}]
    steps:
      - run: {command: echo hi}
      - assert:
          mock: {name: apo, count: 1}
`,
			wantMsg: `"apo" is not a declared mock server (declared: api)`,
		},
		{
			name: "count zero with body matcher",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    mock_servers:
      - name: api
        routes: [{method: GET, path: /}]
    steps:
      - run: {command: echo hi}
      - assert:
          mock:
            name: api
            count: 0
            body: {contains: x}
`,
			wantMsg: "count: 0 cannot be combined",
		},
		{
			name: "relative route path",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    mock_servers:
      - name: api
        routes: [{method: GET, path: v1/x}]
    steps:
      - run: {command: echo hi}
`,
			wantMsg: `must start with "/"`,
		},
		{
			name: "suite mock is a legal assert target",
			src: `
version: "1"
suite:
  name: sample
  setup:
    - mock_server:
        name: shared
        routes: [{method: GET, path: /}]
scenarios:
  - name: s
    steps:
      - run: {command: echo hi}
      - assert:
          mock: {name: shared, count: 0}
`,
			wantMsg: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("sample.atago.yaml", []byte(tc.src))
			if tc.wantMsg == "" {
				if err != nil {
					t.Fatalf("LoadBytes() error = %v, want clean load", err)
				}
				return
			}
			if err == nil {
				t.Fatal("LoadBytes() = nil error, want a mock validation error")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error = %q, want substring %q", err, tc.wantMsg)
			}
		})
	}
}

// TestValidate_SignalStep proves the signal-step rules (#23): the target must
// be a declared service (with the declared names listed), the signal name
// must be in the accepted set, and wait.timeout must parse.
func TestValidate_SignalStep(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, src, wantMsg string
	}{
		{
			name: "unknown service lists declared names",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    services:
      - name: web
        command: ./web
      - name: db
        command: ./db
    steps:
      - signal:
          service: cache
          signal: TERM
`,
			wantMsg: `signal.service "cache" is not a declared service (declared: db, web)`,
		},
		{
			name: "invalid signal name",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    services:
      - name: web
        command: ./web
    steps:
      - signal:
          service: web
          signal: STOP
`,
			wantMsg: "not an accepted signal",
		},
		{
			name: "bad wait timeout",
			src: `
version: "1"
suite:
  name: sample
scenarios:
  - name: s
    services:
      - name: web
        command: ./web
    steps:
      - signal:
          service: web
          signal: TERM
          wait:
            timeout: soon
`,
			wantMsg: "signal.wait.timeout",
		},
		{
			name: "suite service is a legal target",
			src: `
version: "1"
suite:
  name: sample
  setup:
    - service:
        name: shared
        command: ./shared
scenarios:
  - name: s
    steps:
      - signal:
          service: shared
          signal: SIGHUP
      - run:
          command: echo ok
`,
			wantMsg: "", // loads cleanly
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("sample.atago.yaml", []byte(tc.src))
			if tc.wantMsg == "" {
				if err != nil {
					t.Fatalf("LoadBytes() error = %v, want clean load", err)
				}
				return
			}
			if err == nil {
				t.Fatal("LoadBytes() = nil error, want a signal validation error")
			}
			if !strings.Contains(err.Error(), tc.wantMsg) {
				t.Errorf("error = %q, want substring %q", err, tc.wantMsg)
			}
		})
	}
}

// TestValidate_SuiteTimeoutInvalid proves an unparsable suite.timeout is a
// load-time validation error (exit 2) naming the field (#17).
func TestValidate_SuiteTimeoutInvalid(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
  timeout: fast
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
`
	_, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err == nil {
		t.Fatal("LoadBytes() = nil error, want a suite.timeout validation error")
	}
	if !strings.Contains(err.Error(), "suite.timeout") {
		t.Errorf("error = %q, want it to mention suite.timeout", err)
	}
}

// TestApplyDefaults_DoesNotMergeTimeout proves defaults.run.timeout is left to
// the engine's precedence resolver instead of being string-merged into steps
// (#17): the merged step keeps an empty timeout so the resolver can tell the
// levels apart.
func TestApplyDefaults_DoesNotMergeTimeout(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  run:
    timeout: 5s
scenarios:
  - name: s
    steps:
      - run:
          command: echo hi
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if got := s.Scenarios[0].Steps[0].Run.Timeout; got != "" {
		t.Errorf("run.timeout = %q, want empty (the engine resolves defaults.run.timeout itself)", got)
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
