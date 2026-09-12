package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// TestEngine_ServiceReadyFileCapturedAndUsed covers the services feature end to
// end through the engine: a background service publishes a value to a ready file, the engine
// captures it into ${addr}, and a later step reads the value back.
func TestEngine_ServiceReadyFileCapturedAndUsed(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: capture and use a service value
    services:
      - name: pub
        shell: true
        command: '`+publishEchoIdle("ok-127.0.0.1:5000", "ready.txt", "serving", 5)+`'
        ready:
          file: ready.txt
          store: addr
          timeout: 5s
    steps:
      - run: {shell: true, command: "echo ${addr}"}
      - assert:
          stdout: {contains: "ok-127.0.0.1:5000"}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}
}

// TestEngine_ServiceStartFailureErrorsScenario verifies a service whose
// readiness never arrives turns the scenario into an error (steps do not run).
func TestEngine_ServiceStartFailureErrorsScenario(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: service never ready
    services:
      - name: stuck
        shell: true
        command: '`+sleepCmd(5)+`'
        ready:
          file: never.txt
          timeout: 150ms
    steps:
      - run: {shell: true, command: echo should-not-run}
      - assert:
          stdout: {contains: should-not-run}
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	if len(res.Scenarios) != 1 || res.Scenarios[0].Status != StatusError {
		t.Fatalf("scenario status = %+v, want error", res.Scenarios)
	}
	// The failing run/assert steps must not have executed.
	for _, sr := range res.Scenarios[0].Steps {
		if sr.Run != nil {
			t.Errorf("a step ran despite the service failing to become ready")
		}
	}
}

// TestEngine_LeadingFixturesApplyBeforeServices: a real daemon reads its config
// file at startup, so the leading run of fixture steps must land in the workdir
// before services start. The service here fails fast unless the fixture exists
// when it boots.
func TestEngine_LeadingFixturesApplyBeforeServices(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: service reads an authored config at startup
    services:
      - name: configured
        shell: true
        # Copy the authored config to prove it was readable at startup, then
        # publish readiness and idle like a real daemon.
        command: '`+copyThenIdle("app.conf", "seen.conf", 5)+`'
        ready:
          file: seen.conf
          timeout: 5s
    steps:
      - fixture:
          file: app.conf
          content: "listen=127.0.0.1:0"
      - run:
          shell: true
          command: `+catCmd()+` seen.conf
      - assert:
          exit_code: 0
          stdout:
            contains: listen=127.0.0.1:0
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_ServiceWorkdirVarExpansion confirms a service command can reference
// ${workdir} (seeded before services start).
func TestEngine_ServiceWorkdirVarExpansion(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: service writes into workdir via ${workdir}
    services:
      - name: w
        shell: true
        command: '`+publishEchoIdle("done", `"${workdir}/ready.txt"`, "serving", 5)+`'
        ready:
          file: ready.txt
          store: marker
    steps:
      - run: {shell: true, command: "echo ${marker}"}
      - assert:
          stdout: {contains: done}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}
}

// TestEngine_SuiteOnlyKindInScenarioStepNamesTheRule covers the backstop for a
// spec built through the API rather than parsed. `service:` and `mock_server:`
// are suite.setup-only kinds; the loader turns them away with ATG-2106, so this
// path is reachable only when a caller assembles the spec in Go.
//
// It used to report "step has no recognized action", which is wrong twice over:
// atago recognizes both kinds perfectly well, and the message sent the reader
// looking for a typo in a step that has none. The kind is named instead, along
// with where it does belong.
func TestEngine_SuiteOnlyKindInScenarioStepNamesTheRule(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		step spec.Step
		want string
	}{
		{"service", spec.Step{Service: &spec.Service{Name: "worker", Command: "sleep 1"}}, "service steps are only allowed in suite.setup"},
		{"mock_server", spec.Step{MockServer: &spec.MockServer{Name: "api"}}, "mock_server steps are only allowed in suite.setup"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &spec.Spec{
				Version:   "1",
				Suite:     spec.Suite{Name: "api-built"},
				Scenarios: []spec.Scenario{{Name: "misplaced", Steps: []spec.Step{tt.step}}},
			}
			res := New().Run(context.Background(), s, "api.atago.yaml")
			sc := res.Scenarios[0]
			if sc.Status != StatusError {
				t.Fatalf("status = %s, want error", sc.Status)
			}
			got := sc.Steps[0].ErrMsg
			if !strings.Contains(got, tt.want) {
				t.Errorf("ErrMsg = %q, want it to contain %q", got, tt.want)
			}
			if strings.Contains(got, "no recognized action") {
				t.Errorf("ErrMsg claims atago does not recognize a kind it does: %q", got)
			}
		})
	}
}
