package engine

import (
	"strings"
	"testing"
)

// TestResolveTimeout proves the five-level precedence chain (#17): step >
// runner > defaults.run > suite > built-in 60s, with an explicit "0" at any
// level stopping the walk (the documented escape hatch).
func TestResolveTimeout(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, step, runner, defaults, suite string
		wantValue, wantSource               string
	}{
		{"step wins over all", "1s", "2s", "3s", "4s", "1s", "run.timeout"},
		{"runner beats defaults and suite", "", "2s", "3s", "4s", "2s", "runner.timeout"},
		{"defaults beats suite", "", "", "3s", "4s", "3s", "defaults.run.timeout"},
		{"suite beats built-in", "", "", "", "4s", "4s", "suite.timeout"},
		{"built-in when nothing set", "", "", "", "", "60s", "built-in 60s default timeout"},
		{"explicit step 0 disables", "0", "2s", "3s", "4s", "0", "run.timeout"},
		{"explicit suite 0s disables the built-in", "", "", "", "0s", "0s", "suite.timeout"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v, src := resolveTimeout(tc.step, tc.runner, tc.defaults, tc.suite)
			if v != tc.wantValue || src != tc.wantSource {
				t.Errorf("resolveTimeout(%q,%q,%q,%q) = (%q,%q), want (%q,%q)",
					tc.step, tc.runner, tc.defaults, tc.suite, v, src, tc.wantValue, tc.wantSource)
			}
		})
	}
}

// timeoutHint digs the exit_code check hint out of the first failing step so
// the tests can assert which level the timeout-kill message names.
func timeoutHint(t *testing.T, res *SuiteResult) string {
	t.Helper()
	for _, sc := range res.Scenarios {
		for _, st := range sc.Steps {
			for _, c := range st.Checks {
				if c != nil && !c.OK && strings.Contains(c.Hint, "was killed") {
					return c.Hint
				}
			}
		}
	}
	t.Fatalf("no timeout-kill hint found in result: %+v", res.Scenarios)
	return ""
}

// TestEngine_SuiteTimeoutBoundsSteps proves suite.timeout bounds a step with
// no timeout of its own (#17) and the failure hint names suite.timeout.
func TestEngine_SuiteTimeoutBoundsSteps(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
  timeout: 150ms
scenarios:
  - name: suite timeout kills the sleeping step
    steps:
      - run: {shell: true, command: `+sleepCmd(5)+`}
      - assert:
          exit_code: 0
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	if hint := timeoutHint(t, res); !strings.Contains(hint, "suite.timeout") {
		t.Errorf("hint = %q, want it to name suite.timeout", hint)
	}
}

// TestEngine_DefaultsRunTimeoutApplies proves defaults.run.timeout bounds a
// bare step (#17) and the hint names defaults.run.timeout — even though the
// loader no longer string-merges it into the step.
func TestEngine_DefaultsRunTimeoutApplies(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
  timeout: 10s
defaults:
  run:
    timeout: 150ms
scenarios:
  - name: defaults timeout kills the sleeping step
    steps:
      - run: {shell: true, command: `+sleepCmd(5)+`}
      - assert:
          exit_code: 0
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	if hint := timeoutHint(t, res); !strings.Contains(hint, "defaults.run.timeout") {
		t.Errorf("hint = %q, want it to name defaults.run.timeout", hint)
	}
}

// TestEngine_RunnerTimeoutBeatsDefaultsRun proves the runner-common timeout
// outranks defaults.run.timeout in the chain (#17) — the level the loader
// string-merge used to invert.
func TestEngine_RunnerTimeoutBeatsDefaultsRun(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
runners:
  slow:
    type: cmd
    timeout: 10s
defaults:
  run:
    timeout: 50ms
scenarios:
  - name: runner timeout wins so the short sleep survives
    steps:
      - run: {shell: true, runner: slow, command: `+sleepCmd(1)+`}
      - assert:
          exit_code: 0
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed (runner 10s must beat defaults 50ms): %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_StepTimeoutZeroDisables proves `timeout: "0"` opts a step out of
// every timeout level, including a short suite.timeout (#17).
func TestEngine_StepTimeoutZeroDisables(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
  timeout: 150ms
scenarios:
  - name: step zero disables the suite bound
    steps:
      - run: {shell: true, timeout: "0", command: `+sleepCmd(1)+`}
      - assert:
          exit_code: 0
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed (timeout 0 must disable the 150ms suite bound): %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_StepTimeoutHintNamesRunTimeout proves a step-authored timeout
// kill still says run.timeout (#17).
func TestEngine_StepTimeoutHintNamesRunTimeout(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: step timeout kills the sleeping step
    steps:
      - run: {shell: true, timeout: 150ms, command: `+sleepCmd(5)+`}
      - assert:
          exit_code: 0
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	if hint := timeoutHint(t, res); !strings.Contains(hint, "run.timeout") {
		t.Errorf("hint = %q, want it to name run.timeout", hint)
	}
}
