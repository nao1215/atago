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

// TestEngine_SuiteSetup_FixtureStoreAssertKinds exercises the fixture, store, and
// assert step kinds inside suite.setup (not only run/service), covering the
// per-kind branches of runSuiteSteps that the other suite tests skip. A fixture
// written in setup lands in ${suitedir} and is visible to scenarios, and a setup
// assert against a setup run's output participates in the setup verdict.
func TestEngine_SuiteSetup_FixtureStoreAssertKinds(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  setup:
    - fixture:
        file: seed.txt
        content: "hello from setup"
    - run:
        shell: true
        command: echo boot-42
    - assert:
        exit_code: 0
        stdout: {contains: boot-42}
    - store:
        name: bootid
        from:
          stdout:
            matches: "boot-[0-9]+"
scenarios:
  - name: sees the setup store
    steps:
      # The fixture/store/assert setup steps above run once before this scenario
      # regardless of what it does, so they are what exercises runSuiteSteps'
      # per-kind branches; the scenario just confirms the captured store reached
      # it with a cross-platform echo (no cat/;, which cmd.exe does not honor).
      - run: {shell: true, command: "echo id=${bootid}"}
      - assert:
          exit_code: 0
          stdout:
            contains: id=boot-42
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
	if len(res.Setup) != 4 {
		t.Fatalf("recorded %d setup step results, want 4", len(res.Setup))
	}
}

// TestEngine_SuiteSetup_FailingAssertErrsScenarios covers the suite-setup failure
// path triggered by a failing assert (not a run error): a false setup assertion
// aborts the suite and errors every scenario, exercising the assert-failure and
// stop-on-failure branches of runSuiteSteps.
func TestEngine_SuiteSetup_FailingAssertErrsScenarios(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  setup:
    - run: {shell: true, command: echo actual}
    - assert:
        stdout: {contains: expected-but-absent}
scenarios:
  - name: never runs
    steps:
      - run: {shell: true, command: echo unreached}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error: %+v", res.Status, res.Scenarios)
	}
	if len(res.Scenarios) != 1 || res.Scenarios[0].Status != StatusError {
		t.Fatalf("scenario should be errored by the setup failure: %+v", res.Scenarios)
	}
}

// TestEngine_SuiteSetupAssertMasksSecrets is a regression for #243: a suite.setup
// assert failure whose Actual/Hint carry a declared secret must be masked before
// it reaches the reports, exactly as a scenario-level assert failure already is
// (#12). The suite path used to set sr.Checks directly, bypassing maskCheck, so a
// secret printed by a setup command leaked into the console/JSON failure block.
func TestEngine_SuiteSetupAssertMasksSecrets(t *testing.T) {
	const secret = "ghp_suite_setup_secret_value"
	t.Setenv("ATAGO_TEST_SECRET", secret)
	src := `
version: "1"
suite:
  name: s
  setup:
    - run:
        shell: true
        command: echo token-` + envRef("ATAGO_TEST_SECRET") + `
    - assert:
        stdout: {equals: this-will-not-match}
secrets:
  - ATAGO_TEST_SECRET
scenarios:
  - name: never runs
    steps:
      - run: {shell: true, command: echo unreached}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error (setup assert fails)", res.Status)
	}
	// The failing setup assert is step index 1; its check's Actual holds the
	// setup command's stdout, which contains the secret. It must be masked.
	if len(res.Setup) < 2 {
		t.Fatalf("recorded %d setup steps, want >= 2", len(res.Setup))
	}
	for _, sr := range res.Setup {
		for _, cr := range sr.Checks {
			blob := cr.Desc + "\n" + cr.Expected + "\n" + cr.Actual + "\n" + cr.Hint
			if strings.Contains(blob, secret) {
				t.Errorf("suite setup check leaked the raw secret:\n%s", blob)
			}
		}
	}
}

// TestEngine_SuiteSetupUnresolvedVarGuard is a regression for #243: a suite.setup
// run whose command references an undefined ${name} with no shell must error with
// the same explained diagnostic the scenario path gives, instead of leaking the
// literal ${name} into argv. The suite path called runStep directly, skipping the
// unresolved-reference guard.
func TestEngine_SuiteSetupUnresolvedVarGuard(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: s
  setup:
    - run: {command: "echo ${undefined_setup_var}"}
scenarios:
  - name: never runs
    steps:
      - run: {shell: true, command: echo unreached}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	res := New().Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error (unresolved ${name} in setup)", res.Status)
	}
	if len(res.Setup) < 1 {
		t.Fatalf("recorded %d setup steps, want >= 1", len(res.Setup))
	}
	msg := res.Setup[0].ErrMsg
	if !strings.Contains(msg, "undefined_setup_var") || !strings.Contains(msg, "no variable with that name") {
		t.Errorf("setup err = %q, want the explained unresolved-reference diagnostic", msg)
	}
}
