package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/fixture"
	"github.com/nao1215/atago/internal/runner"
	mockrunner "github.com/nao1215/atago/internal/runner/mock"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// suiteArtifactScope names where a suite setup/teardown step's artifacts belong.
// Those steps run once per suite rather than per scenario, so they carry the
// phase label as their scenario name, the -1 index that marks "not a scenario",
// and no attempt of their own.
func suiteArtifactScope(rc runConfig, label string) artifact.Scenario {
	return artifact.Scenario{SpecPath: rc.specPath, Name: label, Index: -1}
}

// suiteSetupLabel names the phase of a scenario error caused by a failed
// suite.setup step, mirroring the "service setup" labeling for pre-step
// failures (#7).
const suiteSetupLabel = "suite setup"

// suiteRuntime carries the once-per-suite state created by suite.setup (#7):
// the ${suitedir} scratch directory, the suite store (builtins, ${suitedir},
// setup-captured values), the suite-wide background services, and the variable
// snapshot seeded into every scenario's store.
type suiteRuntime struct {
	dir      string
	st       *store.Store
	services []*servicerunner.Proc
	// mocks are the suite-wide stub HTTP servers started by suite.setup
	// mock_server steps (#24); their ${<name>.url}/${<name>.port} vars flow to
	// every scenario, and scenario `mock:` asserts can read their records.
	mocks    []*mockrunner.Server
	env      map[string]string // raw suite.env; expanded per use
	vars     map[string]string // seeded into every scenario store
	sshConns map[string]*sshrunner.Runner
}

// set records a suite variable in both the suite store (for later setup steps)
// and the snapshot every scenario receives.
func (rt *suiteRuntime) set(name, value string) {
	rt.st.Set(name, value)
	rt.vars[name] = value
}

// stop tears the suite runtime down: services in LIFO order (after suite
// teardown has run — the caller sequences that), ssh connections, and the
// scratch directory.
func (rt *suiteRuntime) stop() {
	for i := len(rt.services) - 1; i >= 0; i-- {
		rt.services[i].Stop()
	}
	for i := len(rt.mocks) - 1; i >= 0; i-- {
		rt.mocks[i].Stop()
	}
	for _, c := range rt.sshConns {
		_ = c.Close()
	}
	if rt.dir != "" {
		_ = os.RemoveAll(rt.dir)
	}
}

// newSuiteRuntime prepares the suite scratch dir and store. It returns nil
// when the spec declares no suite-level blocks, so the common case pays
// nothing.
func (e *Engine) newSuiteRuntime(s *spec.Spec, specDir, fixturesDir string) (*suiteRuntime, error) {
	if len(s.Suite.Setup) == 0 && len(s.Suite.Teardown) == 0 && len(s.Suite.Env) == 0 {
		return nil, nil
	}
	dir, err := os.MkdirTemp("", "atago-suite-")
	if err != nil {
		return nil, diag.SandboxSetupFailed.Errorf("could not create suite dir: %w", err)
	}
	rt := &suiteRuntime{
		dir:      dir,
		st:       store.New(),
		env:      s.Suite.Env,
		vars:     map[string]string{},
		sshConns: map[string]*sshrunner.Runner{},
	}
	for k, v := range e.builtins {
		rt.st.Set(k, v)
	}
	rt.set(store.BuiltinSuitedir, dir)
	// Suite setup reads the same input directories scenarios do (#394).
	rt.set(store.BuiltinSpecdir, absPath(specDir))
	if fixturesDir != "" {
		rt.set(store.BuiltinFixtures, fixturesDir)
	}
	return rt, nil
}

// runSuiteSteps executes one suite-level block (setup or teardown) in order.
// Setup (stopOnFailure=true) aborts at the first failed step — every scenario
// is then errored by the caller. Teardown (stopOnFailure=false) always runs
// every step: cleanups of independent resources must not shadow each other.
// The returned bool reports whether every step succeeded.
func (e *Engine) runSuiteSteps(ctx context.Context, steps []spec.Step, rt *suiteRuntime, rc runConfig, stopOnFailure bool, label string) ([]StepResult, bool) {
	var out []StepResult
	// rc is this block's own copy. A suite block is its own snapshot writer —
	// it belongs to no scenario, and naming the two blocks separately keeps a
	// clash between them naming both sides — set the same way runScenario names
	// a scenario, so every Env built below it (asserts and `until` polls alike)
	// carries it.
	rc.snapshotWriter = rc.specPath + " / " + label
	x := &suiteStepper{e: e, rt: rt, rc: rc, label: label}
	ok := true

	for i := range steps {
		step := &steps[i]
		sr := StepResult{Index: i, Kind: step.Kind()}

		if ctx.Err() != nil {
			sr.ErrMsg = fmt.Sprintf("run canceled: %v", ctx.Err())
			out = append(out, sr)
			return out, false
		}

		failed := x.exec(ctx, step, i, &sr)

		sr.ErrMsg = rc.masker.Mask(sr.ErrMsg)
		out = append(out, sr)
		if failed {
			ok = false
			if stopOnFailure {
				return out, false
			}
		}
	}
	return out, ok
}

// suiteStepper executes one suite-level block's steps, threading the latest
// run result between them the way scenario steps do (store/assert read the
// most recent run).
type suiteStepper struct {
	e       *Engine
	rt      *suiteRuntime
	rc      runConfig
	label   string
	current *runner.Result
}

// exec dispatches one suite step to its kind's executor and reports failure.
func (x *suiteStepper) exec(ctx context.Context, step *spec.Step, i int, sr *StepResult) (failed bool) {
	switch step.Kind() {
	case spec.StepFixture:
		if err := fixture.Write(expandFixture(x.rt.st, step.Fixture), x.rt.dir, x.rc.specDir); err != nil {
			sr.ErrMsg = err.Error()
			return true
		}
		return false
	case spec.StepRun:
		return x.execRun(ctx, step, i, sr)
	case spec.StepStore:
		val, err := extractValue(expandStore(x.rt.st, step.Store), x.current, x.rt.dir)
		if err != nil {
			sr.ErrMsg = err.Error()
			return true
		}
		x.rt.set(step.Store.Name, val)
		return false
	case spec.StepAssert:
		return x.execAssert(step, i, sr)
	case spec.StepService:
		return x.execService(ctx, step, sr)
	case spec.StepMockServer:
		return x.execMockServer(ctx, step, sr)
	default:
		sr.ErrMsg = fmt.Sprintf("%s steps are not allowed at suite level", step.Kind())
		return true
	}
}

// execRun runs one suite-level command and folds its retry `until` checks.
func (x *suiteStepper) execRun(ctx context.Context, step *spec.Step, i int, sr *StepResult) bool {
	// Same unresolved/leaked-${name} guard the scenario path enforces, so a
	// typo in suite.setup errors with the explained diagnostic instead of
	// leaking the literal reference into argv (#243).
	if msg := runRefGuard(x.rt.st, step.Run, x.rc.runners); msg != "" {
		sr.ErrMsg = msg
		return true
	}
	run := mergeScenarioEnv(resolvableEnv(x.rt.st, x.rt.env), expandRun(x.rt.st, step.Run), x.rt.st)
	r, untilChecks, err := x.e.runStep(ctx, run, x.rt.st, x.rt.dir, x.rc.specDir, x.rc, x.rt.sshConns, nil) // suite setup/teardown steps carry no changes assert
	if err != nil {
		sr.ErrMsg = err.Error()
		return true
	}
	x.current = r
	sr.Run = maskResult(x.rc.masker, r)
	if len(untilChecks) > 0 {
		// recordChecks masks secrets in the check payloads and writes any
		// --artifacts-dir sidecars, exactly as the scenario path does — the
		// suite path used to set sr.Checks raw and leak both (#243).
		x.e.recordChecks(x.rc.masker, untilChecks, suiteArtifactScope(x.rc, x.label), i)
		sr.Checks = untilChecks
		if !assert.AllOK(untilChecks) {
			return true
		}
	}
	return false
}

// execAssert checks one suite-level assert against the latest run result.
func (x *suiteStepper) execAssert(step *spec.Step, i int, sr *StepResult) bool {
	env := x.e.assertEnv(x.rc, x.rt.dir, x.rc.specDir)
	env.MockRecords = func(name string) ([]mockrunner.Record, bool) {
		for _, m := range x.rt.mocks {
			if m.Name() == name {
				return m.Records(), true
			}
		}
		return nil, false
	}
	crs := assert.CheckAll(expandAssert(x.rt.st, step.Assert), x.current, env)
	x.e.recordChecks(x.rc.masker, crs, suiteArtifactScope(x.rc, x.label), i)
	sr.Checks = crs
	return !assert.AllOK(crs)
}

// execService starts one suite-level background service and records it for
// LIFO shutdown even when startup fails partway.
func (x *suiteStepper) execService(ctx context.Context, step *spec.Step, sr *StepResult) bool {
	proc, captured, err := servicerunner.Start(ctx, expandService(x.rt.st, resolvableEnv(x.rt.st, x.rt.env), step.Service), x.rt.dir)
	if proc != nil {
		x.rt.services = append(x.rt.services, proc)
	}
	if err != nil {
		sr.ErrMsg = err.Error()
		return true
	}
	if step.Service.Ready != nil && step.Service.Ready.Store != "" {
		x.rt.set(step.Service.Ready.Store, captured)
	}
	return false
}

// execMockServer starts one suite-level mock server and exposes its address.
func (x *suiteStepper) execMockServer(ctx context.Context, step *spec.Step, sr *StepResult) bool {
	ms, err := mockrunner.Start(ctx, step.MockServer, x.rc.specDir)
	if err != nil {
		sr.ErrMsg = err.Error()
		return true
	}
	x.rt.mocks = append(x.rt.mocks, ms)
	x.rt.set(ms.Name()+".url", ms.URL())
	x.rt.set(ms.Name()+".port", ms.Port())
	return false
}

// suiteSetupFailure summarizes a failed setup block for the per-scenario error.
func suiteSetupFailure(setup []StepResult) string {
	if len(setup) == 0 {
		return suiteSetupLabel + " failed"
	}
	last := setup[len(setup)-1]
	if last.ErrMsg != "" {
		return fmt.Sprintf("%s failed at step %d (%s): %s", suiteSetupLabel, last.Index, last.Kind, last.ErrMsg)
	}
	for _, ck := range last.Checks {
		if ck != nil && !ck.OK {
			return fmt.Sprintf("%s failed at step %d (%s): %s", suiteSetupLabel, last.Index, last.Kind, ck.Desc)
		}
	}
	return fmt.Sprintf("%s failed at step %d (%s)", suiteSetupLabel, last.Index, last.Kind)
}

// runSuiteTeardown executes suite.teardown on a context that survives an
// interrupt (bounded, like scenario teardown) while suite services are still
// up. Failures are recorded but never change the suite verdict.
func (e *Engine) runSuiteTeardown(ctx context.Context, s *spec.Spec, rt *suiteRuntime, rc runConfig) []StepResult {
	if rt == nil || len(s.Suite.Teardown) == 0 {
		return nil
	}
	tctx := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		tctx, cancel = context.WithTimeout(context.Background(), teardownInterruptTimeout)
		defer cancel()
	}
	out, _ := e.runSuiteSteps(tctx, s.Suite.Teardown, rt, rc, false, "suite teardown")
	return out
}
