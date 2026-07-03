package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// The suite.setup block (#7) exists so the bootstrap shell scripts real
// migrations could not shed (build a helper, start a shared peer, warm a
// cache) become spec YAML: an ordered list of steps run ONCE before any
// scenario, in a suite-scoped scratch dir (${suitedir}), where a `service:`
// step starts a suite-wide background process at that exact point in the
// sequence. suite.teardown always runs after the last scenario; suite
// services stop last (LIFO).

// TestEngine_SuiteSetup_RunsOnceAndSharesStore proves setup runs exactly once
// (not per scenario), its captured stores and ${suitedir} are visible to every
// scenario, and suite.env is layered into scenario commands.
func TestEngine_SuiteSetup_RunsOnceAndSharesStore(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  env:
    SHARED_FLAG: from-suite
  setup:
    - run:
        shell: true
        command: "echo setup >> ${suitedir}/count.txt"
    - run:
        shell: true
        command: echo bootstrap-token-99
    - store:
        name: token
        from:
          stdout:
            matches: "bootstrap-token-[0-9]+"
scenarios:
  - name: first sees the shared store and env
    steps:
      - run: {shell: true, command: "echo got ${token} flag=` + envRef("SHARED_FLAG") + `"}
      - assert:
          exit_code: 0
          stdout:
            contains:
              - got bootstrap-token-99
              - flag=from-suite
  - name: second sees them too
    steps:
      - run: {shell: true, command: "echo again ${token}"}
      - assert:
          stdout: {contains: again bootstrap-token-99}
  - name: setup ran exactly once
    steps:
      - run: {shell: true, command: "wc -l < ${suitedir}/count.txt"}
      - assert:
          exit_code: 0
          stdout:
            contains: "1"
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Parallel = 4 // setup must still run once even with parallel scenarios
	res := eng.Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
	if len(res.Setup) != 3 {
		t.Fatalf("recorded %d setup step results, want 3", len(res.Setup))
	}
}

// TestEngine_SuiteService_StartsOnceReadyStoreShared proves a `service:` setup
// step starts a suite-wide process whose ready.store value reaches every
// scenario, ordered relative to surrounding run steps (the service command
// uses a file the preceding run step created).
func TestEngine_SuiteService_StartsOnceReadyStoreShared(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  setup:
    - service:
        name: peer
        shell: true
        command: '` + publishEchoIdle("addr-127.0.0.1:7777", "${suitedir}/ready.txt", "serving", 30) + `'
        ready:
          file: "${suitedir}/ready.txt"
          store: addr
          timeout: 5s
scenarios:
  - name: one
    steps:
      - run: {shell: true, command: "echo dial ${addr}"}
      - assert:
          stdout: {contains: "dial addr-127.0.0.1:7777"}
  - name: two
    steps:
      - run: {shell: true, command: "echo redial ${addr}"}
      - assert:
          stdout: {contains: "redial addr-127.0.0.1:7777"}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_SuiteSetupFailure_ErrsEveryScenario proves a failing setup step
// marks every scenario errored with the suite-setup phase label, and no
// scenario step ever runs.
func TestEngine_SuiteSetupFailure_ErrsEveryScenario(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  setup:
    - run: {command: definitely-not-a-real-binary-xyz}
scenarios:
  - name: never runs
    steps:
      - run: {shell: true, command: echo unreached}
  - name: never runs either
    steps:
      - run: {shell: true, command: echo unreached}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	if len(res.Scenarios) != 2 {
		t.Fatalf("got %d scenario results, want 2", len(res.Scenarios))
	}
	for i, sc := range res.Scenarios {
		if sc.Status != StatusError {
			t.Errorf("scenario[%d] status = %s, want error", i, sc.Status)
		}
		if len(sc.Steps) != 1 || !sc.Steps[0].Setup {
			t.Errorf("scenario[%d] should carry a single setup-phase step result, got %+v", i, sc.Steps)
		}
		if !strings.Contains(sc.Steps[0].ErrMsg, "suite setup") {
			t.Errorf("scenario[%d] ErrMsg = %q, want it to name the suite setup phase", i, sc.Steps[0].ErrMsg)
		}
	}
}

// TestEngine_SuiteTeardown_AlwaysRunsAndServicesOutliveIt proves suite
// teardown runs after the last scenario even when scenarios fail, can still
// talk to suite services (they stop last), and its failures never change the
// suite verdict.
func TestEngine_SuiteTeardown_AlwaysRunsAndServicesOutliveIt(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  setup:
    - service:
        name: peer
        shell: true
        command: '` + publishEchoIdle("up", "${suitedir}/ready.txt", "serving", 30) + `'
        ready:
          file: "${suitedir}/ready.txt"
          timeout: 5s
  teardown:
    - run:
        shell: true
        command: "` + catCmd() + ` ${suitedir}/ready.txt && echo teardown-saw-service > ${suitedir}/td.txt"
    - run: {command: definitely-not-a-real-binary-xyz}
scenarios:
  - name: fails on purpose
    steps:
      - run: {shell: true, command: echo hello}
      - assert:
          stdout: {contains: goodbye}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	// The scenario failed; a failing suite-teardown step must not upgrade or
	// rescue that verdict.
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	if len(res.Teardown) != 2 {
		t.Fatalf("recorded %d suite teardown steps, want 2 (all run despite the failure)", len(res.Teardown))
	}
	if res.Teardown[0].ErrMsg != "" {
		t.Errorf("teardown[0] should have reached the still-running service: %+v", res.Teardown[0])
	}
	if res.Teardown[1].ErrMsg == "" {
		t.Error("teardown[1] should record its execution error")
	}
}

// TestEngine_SuiteSetup_AbsentIsNoOp proves specs without the block behave
// exactly as before (backward compatibility).
func TestEngine_SuiteSetup_AbsentIsNoOp(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: plain
    steps:
      - run: {shell: true, command: echo plain}
      - assert:
          stdout: {contains: plain}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}
	if len(res.Setup) != 0 || len(res.Teardown) != 0 {
		t.Errorf("no-block spec recorded setup/teardown results: %+v / %+v", res.Setup, res.Teardown)
	}
}
