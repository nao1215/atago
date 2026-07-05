package engine

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/loader"
)

func runSpec(t *testing.T, src string) *SuiteResult {
	t.Helper()
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return New().Run(context.Background(), s, "t.atago.yaml")
}

// The helpers below keep the test specs runnable on every OS. Specs run their
// commands with `shell: true`, which maps to /bin/sh on POSIX and cmd.exe on
// Windows; echo/exit are builtins of both shells, and the two genuinely
// different utilities get a per-OS spelling here.

// sleepCmd returns a silent shell command that blocks roughly sec seconds.
// Windows has no sleep; ping -n waits ~1s between its echoes and >nul keeps it
// quiet.
func sleepCmd(sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("ping -n %d 127.0.0.1 >nul", sec+1)
	}
	return fmt.Sprintf("sleep %d", sec)
}

// catCmd is the shell builtin/utility that prints a file's content.
func catCmd() string {
	if runtime.GOOS == "windows" {
		return "type"
	}
	return "cat"
}

// copyThenIdle copies src to dst (failing if src is missing), then blocks — a
// stand-in for a daemon that reads its config file at startup.
func copyThenIdle(src, dst string, sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("copy %s %s& %s", src, dst, sleepCmd(sec))
	}
	return fmt.Sprintf("cp %s %s; %s", src, dst, sleepCmd(sec))
}

// envRef is the shell's expansion syntax for the named environment variable.
func envRef(name string) string {
	if runtime.GOOS == "windows" {
		return "%" + name + "%"
	}
	return "$" + name
}

// echoThenIdle prints msg, then blocks — a stand-in for a chatty service.
func echoThenIdle(msg string, sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("echo %s& %s", msg, sleepCmd(sec))
	}
	return fmt.Sprintf("echo %s; %s", msg, sleepCmd(sec))
}

// publishEchoIdle writes content to file, prints logline, then blocks — a
// stand-in for a server that publishes its readiness and keeps serving.
func publishEchoIdle(content, file, logline string, sec int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("echo %s >%s& echo %s& %s", content, file, logline, sleepCmd(sec))
	}
	return fmt.Sprintf("echo %s > %s; echo %s; %s", content, file, logline, sleepCmd(sec))
}

func TestEngine_PassingScenario(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: echo
    steps:
      - run: {shell: true, command: echo hello}
      - assert: {exit_code: 0}
      - assert:
          stdout: {contains: hello}
      - assert:
          stderr: {empty: true}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}
}

func TestEngine_FailingAssertion(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: exit-1-passes
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
}

func TestEngine_FixtureThenRun(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: read fixture
    steps:
      - fixture:
          file: greeting.txt
          content: "hi there"
      - run: {shell: true, command: `+catCmd()+` greeting.txt}
      - assert:
          stdout: {contains: "hi there"}
      - assert:
          file: {path: greeting.txt, exists: true}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_CmdRunnerCwdAndTimeout proves a cmd runner's common cwd/timeout
// fields reach the local command instead of being silently
// dropped: cwd relocates the step, timeout bounds it, and the step's own value
// still wins over the runner's.
func TestEngine_CmdRunnerCwdAndTimeout(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
runners:
  local:
    type: cmd
    cwd: sub
  slow:
    type: cmd
    timeout: 150ms
scenarios:
  - name: runner cwd relocates the step
    steps:
      - fixture:
          file: sub/here.txt
          content: found-me
      - run: {shell: true, runner: local, command: `+catCmd()+` here.txt}
      - assert:
          stdout: {contains: found-me}
  - name: runner timeout bounds the step
    steps:
      - run: {shell: true, runner: slow, command: `+sleepCmd(5)+`}
      - assert:
          exit_code: {not: 0}
  - name: step timeout wins over the runner timeout
    steps:
      - run: {shell: true, runner: slow, timeout: 5s, command: echo quick}
      - assert:
          exit_code: 0
          stdout: {contains: quick}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_EnvInterpolation proves ${env:NAME} resolves from the host
// environment (t.Setenv forbids t.Parallel): a set variable flows into a
// command, and an unset one on a shell-less run errors naming the variable
// instead of leaking the literal reference into argv.
func TestEngine_EnvInterpolation(t *testing.T) {
	t.Setenv("ATAGO_TEST_GREETING", "hello-env")
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: env ref expands
    steps:
      - run: {shell: true, command: "echo ${env:ATAGO_TEST_GREETING}"}
      - assert:
          stdout: {contains: hello-env}
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}

	res = runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: unset env ref errors
    steps:
      - run: {command: "echo ${env:ATAGO_TEST_DEFINITELY_UNSET}"}
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
	if msg := res.Scenarios[0].Steps[0].ErrMsg; !strings.Contains(msg, "environment variable ATAGO_TEST_DEFINITELY_UNSET is not set") {
		t.Errorf("ErrMsg = %q, want it to name the unset environment variable", msg)
	}
}

// TestEngine_TeardownAlwaysRuns proves the teardown contract: teardown steps
// run after a pass, after an assertion failure, and after an execution error;
// they share the scenario store (a store-captured value flows into cleanup);
// and a teardown failure is reported without changing the scenario's verdict.
func TestEngine_TeardownAlwaysRuns(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: teardown runs after a pass and sees stored values
    steps:
      - run: {shell: true, command: echo resource-42}
      - store:
          name: rid
          from:
            stdout:
              matches: "resource-[0-9]+"
      - assert:
          stdout: {contains: resource-42}
    teardown:
      - run: {shell: true, command: "echo deleting ${rid}"}
      - assert:
          stdout: {contains: deleting resource-42}

  - name: teardown runs after a failed assertion
    steps:
      - run: {shell: true, command: echo hello}
      - assert:
          stdout: {contains: goodbye}
    teardown:
      - run: {shell: true, command: echo cleanup-ran}
      - assert:
          stdout: {contains: cleanup-ran}

  - name: teardown runs after an execution error
    steps:
      - run: {command: definitely-not-a-real-binary-xyz}
    teardown:
      - run: {shell: true, command: echo cleanup-after-error}
      - assert:
          stdout: {contains: cleanup-after-error}
`)
	scs := res.Scenarios
	if scs[0].Status != StatusPassed {
		t.Errorf("scenario[0] status = %s, want passed: %+v", scs[0].Status, scs[0].Teardown)
	}
	if scs[1].Status != StatusFailed {
		t.Errorf("scenario[1] status = %s, want failed (teardown must not rescue it)", scs[1].Status)
	}
	if scs[2].Status != StatusError {
		t.Errorf("scenario[2] status = %s, want error", scs[2].Status)
	}
	for i, sc := range scs {
		if len(sc.Teardown) != 2 {
			t.Errorf("scenario[%d] ran %d teardown steps, want 2", i, len(sc.Teardown))
		}
		if sc.TeardownFailed() {
			t.Errorf("scenario[%d] teardown failed unexpectedly: %+v", i, sc.Teardown)
		}
	}
}

// TestEngine_TeardownFailureDoesNotFlipVerdict proves a failing teardown is
// recorded (TeardownFailed) while the scenario keeps its Steps-decided verdict,
// and that every teardown step still runs after an earlier teardown failure.
func TestEngine_TeardownFailureDoesNotFlipVerdict(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: passes with a failing teardown
    steps:
      - run: {shell: true, command: echo ok}
      - assert:
          stdout: {contains: ok}
    teardown:
      - run: {command: definitely-not-a-real-binary-xyz}
      - run: {shell: true, command: echo second-cleanup}
      - assert:
          stdout: {contains: second-cleanup}
`)
	sc := res.Scenarios[0]
	if sc.Status != StatusPassed {
		t.Fatalf("status = %s, want passed (teardown failure must not flip the verdict)", sc.Status)
	}
	if res.Status != StatusPassed {
		t.Fatalf("suite status = %s, want passed", res.Status)
	}
	if !sc.TeardownFailed() {
		t.Error("TeardownFailed() = false, want true (first teardown step errored)")
	}
	if len(sc.Teardown) != 3 {
		t.Fatalf("ran %d teardown steps, want all 3 despite the first failing", len(sc.Teardown))
	}
	if sc.Teardown[0].ErrMsg == "" {
		t.Error("teardown[0].ErrMsg empty, want the execution error recorded")
	}
	if !assert.AllOK(sc.Teardown[2].Checks) {
		t.Errorf("teardown[2] assert should pass, got %+v", sc.Teardown[2].Checks)
	}
}

func TestEngine_ExecutionError(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: missing binary
    steps:
      - run: {command: definitely-not-a-real-binary-xyz}
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
}

func TestEngine_ParallelKeepsOrderAndRunsAll(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: p
scenarios:
  - name: one
    steps: [{run: {shell: true, command: echo one}}, {assert: {stdout: {contains: one}}}]
  - name: two
    steps: [{run: {shell: true, command: echo two}}, {assert: {stdout: {contains: two}}}]
  - name: three
    steps: [{run: {shell: true, command: echo three}}, {assert: {stdout: {contains: three}}}]
  - name: four
    steps: [{run: {shell: true, command: echo four}}, {assert: {stdout: {contains: four}}}]
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Parallel = 4
	res := eng.Run(context.Background(), s, "t.atago.yaml")

	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed", res.Status)
	}
	want := []string{"one", "two", "three", "four"}
	for i, w := range want {
		if res.Scenarios[i].Name != w {
			t.Errorf("scenario[%d] = %q, want %q (order not preserved)", i, res.Scenarios[i].Name, w)
		}
	}
}

// parityScenarios builds a spec with n scenarios cycling passed/failed/skipped so
// the parallel scheduler has real mixed outcomes to interleave.
func parityScenarios(n int) string {
	var b strings.Builder
	b.WriteString("version: \"1\"\nsuite:\n  name: par\nscenarios:\n")
	for i := range n {
		fmt.Fprintf(&b, "  - name: sc-%03d\n", i)
		switch i % 3 {
		case 0: // passes
			b.WriteString("    steps:\n      - run: {command: \"true\"}\n      - assert: {exit_code: 0}\n")
		case 1: // fails: false exits 1, asserted 0
			b.WriteString("    steps:\n      - run: {command: \"false\"}\n      - assert: {exit_code: 0}\n")
		default: // skipped: an only-probe that never passes
			b.WriteString("    only: {command: \"false\"}\n    steps:\n      - run: {command: \"true\"}\n")
		}
	}
	return b.String()
}

// TestEngine_ParallelMatchesSerial is a metamorphic parity check: --parallel N
// must not change the outcome. For a spec with mixed passed/failed/skipped
// scenarios, every per-scenario verdict and the definition order must be
// identical to a serial (--parallel 1) run, at every worker count. The scheduler
// shares results/done across workers, so a concurrency defect would surface as a
// changed status, a dropped scenario, or reordering; run under -race in CI it
// also guards the shared state against data races.
func TestEngine_ParallelMatchesSerial(t *testing.T) {
	t.Parallel()
	s, err := loader.LoadBytes("par.atago.yaml", []byte(parityScenarios(18)))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	runAt := func(workers int) []ScenarioResult {
		e := New()
		e.Parallel = workers
		return e.Run(context.Background(), s, "par.atago.yaml").Scenarios
	}

	serial := runAt(1)
	for _, workers := range []int{2, 4, 8} {
		for iter := range 4 {
			got := runAt(workers)
			if len(got) != len(serial) {
				t.Fatalf("parallel=%d iter=%d: %d scenarios, serial had %d", workers, iter, len(got), len(serial))
			}
			for i := range serial {
				if got[i].Name != serial[i].Name {
					t.Fatalf("parallel=%d iter=%d: order changed at %d: %q vs serial %q",
						workers, iter, i, got[i].Name, serial[i].Name)
				}
				if got[i].Status != serial[i].Status {
					t.Fatalf("parallel=%d iter=%d: %s status = %s, serial = %s",
						workers, iter, got[i].Name, got[i].Status, serial[i].Status)
				}
			}
		}
	}
}

func TestEngine_FailFastSkipsRemaining(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: ff
scenarios:
  - name: boom
    steps: [{run: {shell: true, command: "exit 1"}}, {assert: {exit_code: 0}}]
  - name: never1
    steps: [{run: {shell: true, command: echo hi}}, {assert: {exit_code: 0}}]
  - name: never2
    steps: [{run: {shell: true, command: echo hi}}, {assert: {exit_code: 0}}]
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Parallel = 1 // deterministic: the first failure stops the rest
	eng.FailFast = true
	res := eng.Run(context.Background(), s, "t.atago.yaml")

	if res.Scenarios[0].Status != StatusFailed {
		t.Errorf("scenario[0] = %s, want failed", res.Scenarios[0].Status)
	}
	for _, i := range []int{1, 2} {
		if res.Scenarios[i].Status != StatusSkipped {
			t.Errorf("scenario[%d] = %s, want skipped (fail-fast)", i, res.Scenarios[i].Status)
		}
	}
}

// Regression for issue #30: when the parent context is cancelled mid-scenario
// (Ctrl-C / suite cancel), the engine step loop must stop rather than keep
// executing steps and evaluating assertions against a killed result. The
// scenario ends in error and not all of its steps run.
func TestEngine_ContextCancelStopsSteps(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: cancel
scenarios:
  - name: long
    steps:
      - run: {shell: true, command: ` + sleepCmd(5) + `}
      - assert: {exit_code: 0}
      - assert: {exit_code: 0}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	res := New().Run(ctx, s, "t.atago.yaml")

	if res.Scenarios[0].Status != StatusError {
		t.Fatalf("status = %s, want error after cancellation", res.Scenarios[0].Status)
	}
	// The sleep step is killed and the trailing assert steps must not all run.
	passedAsserts := 0
	for _, st := range res.Scenarios[0].Steps {
		for _, ck := range st.Checks {
			if ck != nil && ck.OK {
				passedAsserts++
			}
		}
	}
	if passedAsserts == 2 {
		t.Error("both trailing assertions ran; the loop did not stop on cancellation")
	}
}

// Regression for issue #26: fail-fast under --parallel > 1 (the concurrent
// cancellation path) was untested. With one early failure and Parallel >= 4,
// scheduling must stop after the first failure so some scenarios are Skipped,
// and the run must stay race-clean under `go test -race`.
func TestEngine_FailFastParallel(t *testing.T) {
	t.Parallel()
	var b strings.Builder
	b.WriteString("version: \"1\"\nsuite:\n  name: ffp\nscenarios:\n")
	// The first scenario fails fast (a short sleep then a false-exit assertion);
	// the rest sleep long enough that fail-fast can cancel their scheduling.
	b.WriteString("  - name: boom\n    steps: [{run: {shell: true, command: \"exit 1\"}}, {assert: {exit_code: 0}}]\n")
	for i := 0; i < 19; i++ {
		fmt.Fprintf(&b, "  - name: slow-%d\n    steps: [{run: {shell: true, command: %s}}, {assert: {exit_code: 0}}]\n", i, sleepCmd(2))
	}
	s, err := loader.LoadBytes("t.atago.yaml", []byte(b.String()))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Parallel = 4
	eng.FailFast = true
	res := eng.Run(context.Background(), s, "t.atago.yaml")

	if res.Status != StatusFailed {
		t.Fatalf("suite status = %s, want failed", res.Status)
	}
	var failed, skipped, other int
	for _, sc := range res.Scenarios {
		switch sc.Status {
		case StatusFailed:
			failed++
		case StatusSkipped:
			skipped++
		default:
			other++
		}
	}
	if failed == 0 {
		t.Error("expected at least one failed scenario")
	}
	if skipped == 0 {
		t.Errorf("expected some scenarios to be skipped by fail-fast, got failed=%d skipped=%d other=%d", failed, skipped, other)
	}
}

func TestEngine_Selection(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sel
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: echo a}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: echo b}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast]
    steps: [{run: {shell: true, command: echo g}}, {assert: {exit_code: 0}}]
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	tests := []struct {
		name              string
		configure         func(*Engine)
		wantScenarioNames []string
	}{
		{"filter substring", func(e *Engine) { e.FilterName = "bet" }, []string{"beta"}},
		{"tag include", func(e *Engine) { e.Tags = []string{"fast"} }, []string{"alpha", "gamma"}},
		{"skip tag", func(e *Engine) { e.SkipTags = []string{"slow"} }, []string{"alpha", "gamma"}},
		{"no selection runs all", func(e *Engine) {}, []string{"alpha", "beta", "gamma"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			eng := New()
			tt.configure(eng)
			res := eng.Run(context.Background(), s, "t.atago.yaml")
			var got []string
			for _, sc := range res.Scenarios {
				got = append(got, sc.Name)
			}
			if len(got) != len(tt.wantScenarioNames) {
				t.Fatalf("selected %v, want %v", got, tt.wantScenarioNames)
			}
			for i, w := range tt.wantScenarioNames {
				if got[i] != w {
					t.Errorf("selected[%d] = %q, want %q (%v)", i, got[i], w, got)
				}
			}
		})
	}
}

func TestEngine_SkipOnlyByOS(t *testing.T) {
	t.Parallel()
	otherOS := "darwin"
	if runtime.GOOS == "darwin" {
		otherOS = "linux"
	}
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: only-other-os
    only:
      os: `+otherOS+`
    steps:
      - run: {command: echo nope}
`)
	if res.Scenarios[0].Status != StatusSkipped {
		t.Fatalf("status = %s, want skipped", res.Scenarios[0].Status)
	}
}
