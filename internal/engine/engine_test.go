package engine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
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

// TestProbeSucceeds_BoundedByTimeout is a regression: a skip/only probe command
// must be time-bounded so a hanging probe (e.g. `sleep 9999`) cannot stall the
// sequential selection phase and, with it, the whole run.
func TestProbeSucceeds_BoundedByTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX sleep as the hanging probe")
	}
	t.Parallel()
	e := New()
	e.probeTimeout = 100 * time.Millisecond
	start := time.Now()
	ok := e.probeSucceeds(context.Background(), "sleep 30")
	elapsed := time.Since(start)
	if ok {
		t.Error("a hanging probe timed out and must not count as succeeded")
	}
	if elapsed > 2*time.Second {
		t.Errorf("probeSucceeds ran %s; a probe must be bounded by the probe timeout", elapsed)
	}
}

// TestEngine_UnresolvedVarErrorsForNamedCmdRunner is a regression: the
// unresolved-${name} guard fires for any local (non-ssh) run, not just an
// unnamed one. A named cmd runner executes the command as argv the same as the
// default runner, so a typo'd ${name} would otherwise leak into argv and run a
// garbled command instead of erroring with the reference named. Only an ssh
// runner (remote, where a remote shell may expand it) is exempt.
func TestEngine_UnresolvedVarErrorsForNamedCmdRunner(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
runners:
  local:
    type: cmd
scenarios:
  - name: a typo in a named cmd runner errors instead of running literally
    steps:
      - run:
          runner: local
          command: "echo ${typo}"
`)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error (unresolved ${typo} must not run literally through a named cmd runner): %+v",
			res.Status, res.Scenarios)
	}
	msg := res.Scenarios[0].Steps[0].ErrMsg
	if !strings.Contains(msg, "typo") || !strings.Contains(msg, "no variable with that name") {
		t.Errorf("error = %q, want it to name the unresolved ${typo}", msg)
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
		{"filter substring", func(e *Engine) { e.FilterNames = []string{"bet"} }, []string{"beta"}},
		{"filter OR selects multiple", func(e *Engine) { e.FilterNames = []string{"alph", "gam"} }, []string{"alpha", "gamma"}},
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

// TestEngine_SelectionSetAlgebra pins the set semantics of --filter/--tag/
// --skip-tag beyond the single-condition cases in TestEngine_Selection: tags
// are OR-ed (a scenario matches if it carries any listed tag), filter and tag
// compose as intersection (not union), and a skip-tag drops a scenario even when
// a tag it also carries was included — skip is applied last and wins. An empty
// selection (every scenario skipped) is a valid, non-error outcome.
func TestEngine_SelectionSetAlgebra(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sel
scenarios:
  - name: login
    tags: [fast, auth]
    steps: [{run: {shell: true, command: echo x}}, {assert: {exit_code: 0}}]
  - name: logout
    tags: [fast, auth]
    steps: [{run: {shell: true, command: echo x}}, {assert: {exit_code: 0}}]
  - name: signup
    tags: [slow, auth]
    steps: [{run: {shell: true, command: echo x}}, {assert: {exit_code: 0}}]
  - name: search
    tags: [fast, core]
    steps: [{run: {shell: true, command: echo x}}, {assert: {exit_code: 0}}]
  - name: billing
    tags: [slow, core, fast]
    steps: [{run: {shell: true, command: echo x}}, {assert: {exit_code: 0}}]
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	tests := []struct {
		name      string
		configure func(*Engine)
		want      []string
	}{
		{"tag matches any listed tag", func(e *Engine) { e.Tags = []string{"auth"} },
			[]string{"login", "logout", "signup"}},
		{"multiple tags union", func(e *Engine) { e.Tags = []string{"fast", "slow"} },
			[]string{"login", "logout", "signup", "search", "billing"}},
		{"skip-tag wins over an included tag", func(e *Engine) { e.Tags = []string{"fast"}; e.SkipTags = []string{"slow"} },
			[]string{"login", "logout", "search"}},
		{"filter intersects tag", func(e *Engine) { e.FilterNames = []string{"log"}; e.Tags = []string{"fast"} },
			[]string{"login", "logout"}},
		{"skip-tag alone", func(e *Engine) { e.SkipTags = []string{"core"} },
			[]string{"login", "logout", "signup"}},
		{"skip-tag can empty the selection", func(e *Engine) { e.Tags = []string{"core"}; e.SkipTags = []string{"fast"} },
			nil},
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
			if len(got) != len(tt.want) {
				t.Fatalf("selected %v, want %v", got, tt.want)
			}
			for i, w := range tt.want {
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

// TestExtractValue_AllBranches exercises every source branch of extractValue,
// asserting both the "wrong family / no step yet" guards and the happy path for
// each. These branches (body/header/rows/message/value) were previously
// unexercised, and each guard is a data-integrity gate: capturing from the wrong
// result family would store a stale or empty value under a variable name a later
// step trusts.
func TestExtractValue_AllBranches(t *testing.T) {
	t.Parallel()

	jsonSel := &spec.StreamAssert{JSON: &spec.JSONAssert{Path: "$.v"}}
	body := []byte(`{"v":"hit"}`)

	httpRes := &runner.Result{IsHTTP: true, Body: body, Header: http.Header{"X-Token": {"abc"}}}
	dbRes := &runner.Result{IsDB: true, RowsJSON: body}
	grpcRes := &runner.Result{IsGRPC: true, MessageJSON: body}
	cdpRes := &runner.Result{IsCDP: true, CDPValue: body}

	tests := []struct {
		name    string
		from    *spec.StoreFrom
		current *runner.Result
		want    string
		wantErr string // substring; "" means expect success
	}{
		{"nil from", nil, httpRes, "", "'from' is required"},
		{"empty from", &spec.StoreFrom{}, httpRes, "", "must set stdout"},

		{"stdout nil result", &spec.StoreFrom{Stdout: jsonSel}, nil, "", "no command has run"},

		{"body ok", &spec.StoreFrom{Body: jsonSel}, httpRes, "hit", ""},
		{"body wrong family", &spec.StoreFrom{Body: jsonSel}, dbRes, "", "no HTTP request"},
		{"body nil result", &spec.StoreFrom{Body: jsonSel}, nil, "", "no HTTP request"},

		{"header ok", &spec.StoreFrom{Header: "X-Token"}, httpRes, "abc", ""},
		{"header missing", &spec.StoreFrom{Header: "X-Absent"}, httpRes, "", "has no"},
		{"header wrong family", &spec.StoreFrom{Header: "X-Token"}, dbRes, "", "no HTTP request"},

		{"rows ok", &spec.StoreFrom{Rows: jsonSel}, dbRes, "hit", ""},
		{"rows wrong family", &spec.StoreFrom{Rows: jsonSel}, httpRes, "", "no query has run"},

		{"message ok", &spec.StoreFrom{Message: jsonSel}, grpcRes, "hit", ""},
		{"message wrong family", &spec.StoreFrom{Message: jsonSel}, httpRes, "", "no gRPC call"},

		{"value ok", &spec.StoreFrom{Value: jsonSel}, cdpRes, "hit", ""},
		{"value wrong family", &spec.StoreFrom{Value: jsonSel}, httpRes, "", "no cdp step"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sp := &spec.Store{Name: "x", From: tt.from}
			got, err := extractValue(sp, tt.current, "")
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got value %q", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestExtractValue_FileNeedsJSONSelector proves a file source without a json
// selector is a clean error, not a nil-deref.
func TestExtractValue_FileNeedsJSONSelector(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f.json"), []byte(`{"v":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	sp := &spec.Store{Name: "x", From: &spec.StoreFrom{File: &spec.FileAssert{Path: "f.json"}}}
	_, err := extractValue(sp, &runner.Result{}, dir)
	if err == nil || !strings.Contains(err.Error(), "needs a json or text selector") {
		t.Fatalf("want json/text-selector error, got %v", err)
	}
}

// TestExtractStream_NoSelector proves a stream source that sets neither json,
// matches, nor trim is a clean error.
func TestExtractStream_NoSelector(t *testing.T) {
	t.Parallel()
	_, err := extractStream(&spec.StreamAssert{}, []byte("data"))
	if err == nil || !strings.Contains(err.Error(), "json, matches, or trim") {
		t.Fatalf("want selector error, got %v", err)
	}
}

// TestJSONValue_NullSelectsError is a regression for the bug where a JSON path
// selecting a null value stored the Go literal string "<nil>" instead of erroring
// — a Go-ism leaking into a user-visible variable that silently masked "the field
// was null".
func TestJSONValue_NullSelectsError(t *testing.T) {
	t.Parallel()
	got, err := jsonValue([]byte(`{"token":null}`), "$.token")
	if err == nil {
		t.Fatalf("null capture returned %q, want an error", got)
	}
	if strings.Contains(got, "<nil>") {
		t.Fatalf("leaked Go nil literal: %q", got)
	}
	if !strings.Contains(err.Error(), "null") {
		t.Errorf("error %q should mention the null value", err)
	}
}

// TestJSONValue_Edges covers the remaining jsonValue paths: invalid JSON, invalid
// path syntax, a path selecting a container (array/object), and multibyte scalars.
func TestJSONValue_Edges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, data, path, want string
		wantErr                string
	}{
		{"invalid json", `{`, "$.x", "", "invalid JSON"},
		{"invalid path", `{"x":1}`, "$[", "", "invalid JSON path"},
		{"multibyte scalar", `{"名前":"太郎"}`, "$.名前", "太郎", ""},
		{"number int", `{"x":42}`, "$.x", "42", ""},
		{"bool", `{"x":true}`, "$.x", "true", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := jsonValue([]byte(tt.data), tt.path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got %q err %v, want %q", got, err, tt.want)
			}
		})
	}
}

// TestRegexValue_Edges covers invalid regexp, a multibyte capture, and the
// documented "no capture group -> whole match" fallback.
func TestRegexValue_Edges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, data, pattern, want string
		wantErr                   string
	}{
		{"invalid regexp", "x", `(`, "", "invalid regexp"},
		{"multibyte capture", "名前=太郎さん", `名前=(\p{Han}+)`, "太郎", ""},
		{"no group whole match", "abc123", `[a-z]+`, "abc", ""},
		{"first group only", "a=1;b=2", `(\w)=(\d)`, "a", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := regexValue([]byte(tt.data), tt.pattern)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got %q err %v, want %q", got, err, tt.want)
			}
		})
	}
}

// TestMergedEnv_Precedence proves own wins per key, base-only keys survive, and
// the empty-base fast path returns own unchanged without mutating either map.
// An inverted precedence here would silently let a suite-wide default override a
// scenario's explicit value.
func TestMergedEnv_Precedence(t *testing.T) {
	t.Parallel()
	base := map[string]string{"A": "base", "B": "baseOnly"}
	own := map[string]string{"A": "own", "C": "ownOnly"}
	got := mergedEnv(base, own)
	if got["A"] != "own" {
		t.Errorf("own must win: A=%q", got["A"])
	}
	if got["B"] != "baseOnly" || got["C"] != "ownOnly" {
		t.Errorf("keys lost: %+v", got)
	}
	// Inputs untouched.
	if base["A"] != "base" || own["A"] != "own" {
		t.Errorf("inputs mutated: base=%v own=%v", base, own)
	}
	// Empty base returns the exact own map (documented no-alloc fast path).
	if out := mergedEnv(nil, own); len(out) != len(own) {
		t.Errorf("empty base should return own, got %+v", out)
	}
}

// TestExpandHTTP_Fields proves ${name} substitution reaches every
// user-controllable HTTP field (path, header values, JSON body, raw body,
// body_file, body_to, form values, file part paths) and leaves the input HTTP
// spec untouched (expandHTTP returns a copy).
func TestExpandHTTP_Fields(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("id", "42")
	st.Set("tok", "secret")
	st.Set("workdir", "/wd")

	in := &spec.HTTP{
		Path:     "/users/${id}",
		Header:   map[string]string{"Authorization": "Bearer ${tok}"},
		JSON:     map[string]any{"name": "u${id}"},
		Body:     "raw-${id}",
		BodyFile: "${workdir}/in.bin",
		BodyTo:   "${workdir}/out.bin",
		Form:     map[string]string{"q": "${tok}"},
		Files:    []spec.FilePart{{Path: "${workdir}/a.png"}},
	}
	out := expandHTTP(st, in)

	if out.Path != "/users/42" {
		t.Errorf("path = %q", out.Path)
	}
	if out.Header["Authorization"] != "Bearer secret" {
		t.Errorf("header = %q", out.Header["Authorization"])
	}
	if m, _ := out.JSON.(map[string]any); m["name"] != "u42" {
		t.Errorf("json = %v", out.JSON)
	}
	if out.Body != "raw-42" {
		t.Errorf("body = %q", out.Body)
	}
	if out.BodyFile != "/wd/in.bin" || out.BodyTo != "/wd/out.bin" {
		t.Errorf("body_file/body_to = %q/%q", out.BodyFile, out.BodyTo)
	}
	if out.Form["q"] != "secret" {
		t.Errorf("form = %q", out.Form["q"])
	}
	if out.Files[0].Path != "/wd/a.png" {
		t.Errorf("file part = %q", out.Files[0].Path)
	}
	// Input untouched.
	if in.Path != "/users/${id}" || in.Files[0].Path != "${workdir}/a.png" {
		t.Errorf("input mutated: %+v", in)
	}
}

// TestExpandService_EnvPrecedence proves a background service's own env overrides
// the scenario env on a key conflict, and both are ${name}-expanded. An inverted
// precedence would let a scenario-wide default clobber a service's explicit
// credential.
func TestExpandService_EnvPrecedence(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("home", "/wd")
	svc := &spec.Service{
		Name:    "db",
		Command: "${home}/server",
		Cwd:     "${home}",
		Env:     map[string]string{"MODE": "service", "PORT": "${home}"},
	}
	scenarioEnv := map[string]string{"MODE": "scenario", "SHARED": "${home}/x"}
	out := expandService(st, scenarioEnv, svc)

	if out.Command != "/wd/server" || out.Cwd != "/wd" {
		t.Errorf("command/cwd = %q/%q", out.Command, out.Cwd)
	}
	if out.Env["MODE"] != "service" {
		t.Errorf("service env must win: MODE=%q", out.Env["MODE"])
	}
	if out.Env["SHARED"] != "/wd/x" {
		t.Errorf("scenario env not expanded: SHARED=%q", out.Env["SHARED"])
	}
	if out.Env["PORT"] != "/wd" {
		t.Errorf("service env not expanded: PORT=%q", out.Env["PORT"])
	}
}

// TestNormalizeHostAndAllowedHosts covers normalizeHost's URL vs bare-host
// branches and how allowedHosts threads them out of a spec. A full URL must
// collapse to its host[:port]; a bare host or host:port passes through verbatim
// so it can match the runner's url.Host comparison.
func TestNormalizeHostAndAllowedHosts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		entry, want string
	}{
		{"https://example.com", "example.com"},
		{"https://example.com:8443/path", "example.com:8443"},
		{"example.com", "example.com"},
		{"example.com:9000", "example.com:9000"},
	}
	for _, tt := range tests {
		if got := normalizeHost(tt.entry); got != tt.want {
			t.Errorf("normalizeHost(%q) = %q, want %q", tt.entry, got, tt.want)
		}
	}

	// No permissions block -> nil (no restriction).
	if got := allowedHosts(&spec.Spec{}); got != nil {
		t.Errorf("allowedHosts with no permissions = %v, want nil", got)
	}
	s := &spec.Spec{Permissions: &spec.Permissions{Network: &spec.NetworkPolicy{
		Allow: []string{"https://api.example.com", "localhost:8080"},
	}}}
	got := allowedHosts(s)
	if len(got) != 2 || got[0] != "api.example.com" || got[1] != "localhost:8080" {
		t.Errorf("allowedHosts = %v", got)
	}
}

// TestServiceLogged proves the duplicate-guard used by the service-log artifact
// writer: a name already recorded reports true, an unrecorded name false. A false
// negative here would write a service's log (and its masked secrets) twice.
func TestServiceLogged(t *testing.T) {
	t.Parallel()
	out := &ScenarioResult{ServiceLogs: []ServiceLog{{Name: "api"}, {Name: "db"}}}
	if !serviceLogged(out, "api") || !serviceLogged(out, "db") {
		t.Error("recorded services should report logged")
	}
	if serviceLogged(out, "cache") {
		t.Error("unrecorded service should report not logged")
	}
	if serviceLogged(&ScenarioResult{}, "api") {
		t.Error("empty result should report not logged")
	}
}

// TestSkipReason covers every gate of skipReason: OS-only/skip, env-only/skip,
// and the command probe. Each gate is a scenario-selection decision; an inverted
// gate would run a scenario that should be skipped (or vice versa).
func TestSkipReason(t *testing.T) {
	// Not parallel: env-based gates use t.Setenv.
	e := New()
	ctx := context.Background()
	otherOS := "plan9-nonexistent"

	// only.OS pointing at a non-host OS -> skip.
	if r, skip := e.skipReason(ctx, &spec.Scenario{Only: &spec.Condition{OS: otherOS}}); !skip || !strings.Contains(r, "only on os") {
		t.Errorf("only.OS mismatch: reason=%q skip=%v", r, skip)
	}
	// skip.OS matching the host -> skip.
	if r, skip := e.skipReason(ctx, &spec.Scenario{Skip: &spec.Condition{OS: runtime.GOOS}}); !skip || !strings.Contains(r, "skip on os") {
		t.Errorf("skip.OS match: reason=%q skip=%v", r, skip)
	}
	// only.OS matching the host -> not skipped.
	if _, skip := e.skipReason(ctx, &spec.Scenario{Only: &spec.Condition{OS: runtime.GOOS}}); skip {
		t.Error("only.OS matching host should not skip")
	}

	// only.Env unset -> skip; set -> run.
	if r, skip := e.skipReason(ctx, &spec.Scenario{Only: &spec.Condition{Env: "ATAGO_UNSET_XYZ"}}); !skip || !strings.Contains(r, "only when env") {
		t.Errorf("only.Env unset: reason=%q skip=%v", r, skip)
	}
	t.Setenv("ATAGO_SET_XYZ", "1")
	if _, skip := e.skipReason(ctx, &spec.Scenario{Only: &spec.Condition{Env: "ATAGO_SET_XYZ"}}); skip {
		t.Error("only.Env set should not skip")
	}
	// skip.Env set -> skip.
	if r, skip := e.skipReason(ctx, &spec.Scenario{Skip: &spec.Condition{Env: "ATAGO_SET_XYZ"}}); !skip || !strings.Contains(r, "skip when env") {
		t.Errorf("skip.Env set: reason=%q skip=%v", r, skip)
	}

	// Command probes (POSIX shells only; cmd.exe lacks true/false builtins).
	if runtime.GOOS != "windows" {
		// only.Command that fails -> skip.
		if r, skip := e.skipReason(ctx, &spec.Scenario{Only: &spec.Condition{Command: "exit 1"}}); !skip || !strings.Contains(r, "only when command") {
			t.Errorf("only.Command fail: reason=%q skip=%v", r, skip)
		}
		// skip.Command that succeeds -> skip.
		if r, skip := e.skipReason(ctx, &spec.Scenario{Skip: &spec.Condition{Command: "exit 0"}}); !skip || !strings.Contains(r, "skip when command") {
			t.Errorf("skip.Command success: reason=%q skip=%v", r, skip)
		}
		// No gates -> run.
		if _, skip := e.skipReason(ctx, &spec.Scenario{}); skip {
			t.Error("scenario with no gates should not skip")
		}
	}
}

// TestSuiteSetupFailure covers each branch of the setup-failure summary: empty
// setup, a step ErrMsg, a failing check, and the no-detail fallback.
func TestSuiteSetupFailure(t *testing.T) {
	t.Parallel()
	if got := suiteSetupFailure(nil); got != suiteSetupLabel+" failed" {
		t.Errorf("empty setup = %q", got)
	}
	errMsg := suiteSetupFailure([]StepResult{{Index: 2, Kind: spec.StepRun, ErrMsg: "boom"}})
	if !strings.Contains(errMsg, "step 2") || !strings.Contains(errMsg, "boom") {
		t.Errorf("errmsg branch = %q", errMsg)
	}
	checkFail := suiteSetupFailure([]StepResult{{Index: 1, Kind: spec.StepAssert, Checks: []*assert.CheckResult{{OK: false, Desc: "want 200"}}}})
	if !strings.Contains(checkFail, "want 200") {
		t.Errorf("check branch = %q", checkFail)
	}
	fallback := suiteSetupFailure([]StepResult{{Index: 0, Kind: spec.StepFixture}})
	if !strings.Contains(fallback, "step 0") || strings.Contains(fallback, ":") {
		t.Errorf("fallback branch = %q", fallback)
	}
}

// TestSuiteResultCounts tallies each status exactly once, proving flaky is a
// distinct bucket (not folded into passed or failed) — the report's headline
// numbers depend on it.
func TestSuiteResultCounts(t *testing.T) {
	t.Parallel()
	s := &SuiteResult{Scenarios: []ScenarioResult{
		{Status: StatusPassed}, {Status: StatusPassed},
		{Status: StatusFailed},
		{Status: StatusSkipped},
		{Status: StatusError},
		{Status: StatusFlaky},
	}}
	c := s.Counts()
	if c != (Counts{Passed: 2, Failed: 1, Skipped: 1, Errored: 1, Flaky: 1}) {
		t.Errorf("counts = %+v", c)
	}
}

// TestScenarioID proves the identity key is stable and distinguishes same-named
// scenarios across different spec paths (the --rerun-failed / --select key).
func TestScenarioID(t *testing.T) {
	t.Parallel()
	a := ScenarioID("a.yaml", "login")
	b := ScenarioID("b.yaml", "login")
	if a == b {
		t.Error("same name in different specs must not collide")
	}
	if a != ScenarioID("a.yaml", "login") {
		t.Error("ScenarioID must be deterministic")
	}
}

// TestEngineMatches covers each scenario-filter gate: name substring, required
// tags, excluded tags, and the explicit --select set. A gate that reads inverted
// would run the wrong subset of scenarios.
func TestEngineMatches(t *testing.T) {
	t.Parallel()
	sc := &spec.Scenario{Name: "login flow", Tags: []string{"smoke", "auth"}}

	if !New().matches(sc, "s.yaml") {
		t.Error("no filters should match everything")
	}

	byName := New()
	byName.FilterNames = []string{"login"}
	if !byName.matches(sc, "s.yaml") {
		t.Error("name substring should match")
	}
	byName.FilterNames = []string{"logout"}
	if byName.matches(sc, "s.yaml") {
		t.Error("non-matching name substring should reject")
	}
	byName.FilterNames = []string{"logout", "login"}
	if !byName.matches(sc, "s.yaml") {
		t.Error("OR of substrings should match when any matches")
	}

	byTag := New()
	byTag.Tags = []string{"smoke"}
	if !byTag.matches(sc, "s.yaml") {
		t.Error("required tag present should match")
	}
	byTag.Tags = []string{"nightly"}
	if byTag.matches(sc, "s.yaml") {
		t.Error("required tag absent should reject")
	}

	bySkipTag := New()
	bySkipTag.SkipTags = []string{"auth"}
	if bySkipTag.matches(sc, "s.yaml") {
		t.Error("excluded tag present should reject")
	}

	bySelect := New()
	bySelect.Select = map[string]bool{ScenarioID("s.yaml", "login flow"): true}
	if !bySelect.matches(sc, "s.yaml") {
		t.Error("selected id should match")
	}
	if bySelect.matches(sc, "other.yaml") {
		t.Error("unselected id (different path) should reject")
	}
}

// TestResolveHTTPConfig covers the named-runner resolution branches: the
// no-runner default (built-in 60s), an unknown runner, a wrong-typed runner, a
// runner whose base_url is ${name}-expanded, and an invalid runner timeout.
func TestResolveHTTPConfig(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("host", "example.com")

	rc := runConfig{runners: map[string]spec.Runner{
		"api":    {Type: "http", BaseURL: "https://${host}", Timeout: "5s"},
		"dbconn": {Type: "db", DSN: "sqlite:x"},
		"bad":    {Type: "http", Timeout: "not-a-duration"},
	}}

	// No runner: default timeout applies, no base_url.
	cfg, err := resolveHTTPConfig(&spec.HTTP{}, st, rc)
	if err != nil || cfg.Timeout <= 0 {
		t.Fatalf("default: cfg=%+v err=%v", cfg, err)
	}

	// Named http runner: base_url expanded, its timeout used.
	cfg, err = resolveHTTPConfig(&spec.HTTP{Runner: "api"}, st, rc)
	if err != nil || cfg.BaseURL != "https://example.com" {
		t.Fatalf("api: cfg=%+v err=%v", cfg, err)
	}

	// Unknown runner.
	if _, err := resolveHTTPConfig(&spec.HTTP{Runner: "ghost"}, st, rc); err == nil || !strings.Contains(err.Error(), "unknown runner") {
		t.Fatalf("unknown runner err = %v", err)
	}
	// Wrong type.
	if _, err := resolveHTTPConfig(&spec.HTTP{Runner: "dbconn"}, st, rc); err == nil || !strings.Contains(err.Error(), "not an http runner") {
		t.Fatalf("wrong-type err = %v", err)
	}
	// Invalid timeout.
	if _, err := resolveHTTPConfig(&spec.HTTP{Runner: "bad"}, st, rc); err == nil || !strings.Contains(err.Error(), "invalid timeout") {
		t.Fatalf("invalid-timeout err = %v", err)
	}
}

// TestRunHTTP_UnknownRunnerErrors proves runHTTP surfaces a config error (no
// request attempted) and does not flag it as a network-policy violation.
func TestRunHTTP_UnknownRunnerErrors(t *testing.T) {
	t.Parallel()
	e := New()
	rc := runConfig{runners: map[string]spec.Runner{}, masker: security.NewMasker(nil)}
	_, secViolation, err := e.runHTTP(context.Background(), &spec.HTTP{Runner: "ghost"}, store.New(), rc, t.TempDir())
	if err == nil || secViolation {
		t.Fatalf("want plain config error, got err=%v secViolation=%v", err, secViolation)
	}
}

// TestMaskCheck_ArtifactPayloads proves the durable artifact payloads (#48) are
// masked before they can reach a sidecar file — a secret in ArtifactActual /
// ArtifactExpected must not survive maskCheck. A nil "no expected payload" must
// stay nil rather than becoming an empty masked slice.
func TestMaskCheck_ArtifactPayloads(t *testing.T) {
	t.Parallel()
	m := security.NewMasker([]string{"s3cret"})
	cr := &assert.CheckResult{
		ArtifactActual:   []byte("token=s3cret in body"),
		ArtifactExpected: nil,
	}
	maskCheck(m, cr)
	if strings.Contains(string(cr.ArtifactActual), "s3cret") {
		t.Errorf("ArtifactActual leaked secret: %q", cr.ArtifactActual)
	}
	if cr.ArtifactExpected != nil {
		t.Errorf("nil ArtifactExpected became %q; must stay nil", cr.ArtifactExpected)
	}
	// nil CheckResult and empty masker are both no-ops (no panic).
	maskCheck(m, nil)
	maskCheck(security.NewMasker(nil), &assert.CheckResult{Desc: "s3cret stays"})
}

// TestWriteArtifacts covers the sidecar writer: it writes actual+expected text
// and each binary blob for a failed check, records their paths, and is a no-op
// for passing checks, empty ArtifactKind, or a nil Artifacts dir.
func TestWriteArtifacts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	e := New()
	e.Artifacts = artifact.NewDir(root)

	cr := &assert.CheckResult{
		OK:               false,
		ArtifactKind:     "stdout",
		ArtifactActual:   []byte("actual output"),
		ArtifactExpected: []byte("expected output"),
		ArtifactBlobs:    []assert.ArtifactBlob{{Role: "diff", Ext: "png", Data: []byte("PNGDATA")}, {Role: "empty", Ext: "bin", Data: nil}},
	}
	e.writeArtifacts(cr, "t.atago.yaml", "scn", 0, 1)
	if len(cr.ArtifactFiles) != 3 { // actual, expected, diff (empty blob skipped)
		t.Fatalf("ArtifactFiles = %d, want 3: %+v", len(cr.ArtifactFiles), cr.ArtifactFiles)
	}
	for _, f := range cr.ArtifactFiles {
		if _, err := os.Stat(filepath.Join(root, f.Path)); err != nil {
			t.Errorf("artifact %q not written: %v", f.Path, err)
		}
	}

	// No-op guards.
	passing := &assert.CheckResult{OK: true, ArtifactKind: "stdout", ArtifactActual: []byte("x")}
	e.writeArtifacts(passing, "t.atago.yaml", "scn", 0, 0)
	if len(passing.ArtifactFiles) != 0 {
		t.Error("passing check should write nothing")
	}
	noKind := &assert.CheckResult{OK: false, ArtifactActual: []byte("x")}
	e.writeArtifacts(noKind, "t.atago.yaml", "scn", 0, 0)
	if len(noKind.ArtifactFiles) != 0 {
		t.Error("empty ArtifactKind should write nothing")
	}
	// Nil Artifacts dir: no panic, no write.
	(&Engine{}).writeArtifacts(cr, "t.atago.yaml", "scn", 0, 0)
}

// TestEngine_LeadingFixtureFailureErrorsScenario proves a leading fixture that
// cannot be written (its `from` source is missing) errors the scenario before any
// service or run step executes — the applyLeadingFixtures failure path. The later
// run step must never run.
func TestEngine_LeadingFixtureFailureErrorsScenario(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: leading fixture fails
    steps:
      - fixture:
          file: config.txt
          from: this-source-does-not-exist.txt
      - run:
          shell: true
          command: echo unreached
`)
	sc := res.Scenarios[0]
	if sc.Status != StatusError {
		t.Fatalf("status = %s, want error: %+v", sc.Status, sc.Steps)
	}
	for _, st := range sc.Steps {
		if st.Run != nil && strings.Contains(string(st.Run.Stdout), "unreached") {
			t.Error("run step executed despite the leading fixture failure")
		}
	}
}

type bughuntCloser struct{ closed bool }

func (f *bughuntCloser) Close() error { f.closed = true; return nil }

// TestResolveConn covers the shared per-runner connection resolver used by the
// db/grpc/ssh/browser step helpers: cache hit (no re-open), unknown runner,
// wrong-type runner, invalid timeout, and an open error. A cache miss that
// re-opened on every step would leak connections; a wrong-type acceptance would
// run a query against, say, an ssh runner.
func TestResolveConn(t *testing.T) {
	t.Parallel()
	rc := runConfig{runners: map[string]spec.Runner{
		"db":   {Type: "db", Timeout: "5s"},
		"bad":  {Type: "db", Timeout: "not-a-duration"},
		"http": {Type: "http"},
	}}
	conns := map[string]*bughuntCloser{}
	opens := 0
	open := func(spec.Runner, time.Duration) (*bughuntCloser, error) {
		opens++
		return &bughuntCloser{}, nil
	}

	c1, err := resolveConn("db", "query step", "db", rc, conns, false, open)
	if err != nil || c1 == nil {
		t.Fatalf("first resolve: %v", err)
	}
	c2, err := resolveConn("db", "query step", "db", rc, conns, false, open)
	if err != nil || c1 != c2 || opens != 1 {
		t.Fatalf("cache miss: opens=%d c1==c2=%v", opens, c1 == c2)
	}

	if _, err := resolveConn("ghost", "query step", "db", rc, conns, false, open); err == nil || !strings.Contains(err.Error(), "unknown runner") {
		t.Fatalf("unknown runner err = %v", err)
	}
	if _, err := resolveConn("http", "query step", "db", rc, conns, false, open); err == nil || !strings.Contains(err.Error(), "not a db runner") {
		t.Fatalf("wrong-type err = %v", err)
	}
	// defaultBounded=false so the raw (invalid) runner timeout is parsed and fails.
	if _, err := resolveConn("bad", "query step", "db", rc, conns, false, open); err == nil || !strings.Contains(err.Error(), "invalid timeout") {
		t.Fatalf("invalid-timeout err = %v", err)
	}
	openErr := func(spec.Runner, time.Duration) (*bughuntCloser, error) { return nil, errors.New("boom") }
	if _, err := resolveConn("db", "query step", "db", rc, map[string]*bughuntCloser{}, true, openErr); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("open error = %v", err)
	}
}
