package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/fixture"
	"github.com/nao1215/atago/internal/runner"
	mockrunner "github.com/nao1215/atago/internal/runner/mock"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

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
func (e *Engine) newSuiteRuntime(s *spec.Spec) (*suiteRuntime, error) {
	if len(s.Suite.Setup) == 0 && len(s.Suite.Teardown) == 0 && len(s.Suite.Env) == 0 {
		return nil, nil
	}
	dir, err := os.MkdirTemp("", "atago-suite-")
	if err != nil {
		return nil, fmt.Errorf("could not create suite dir: %w", err)
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
	rt.set("suitedir", dir)
	return rt, nil
}

// runSuiteSteps executes one suite-level block (setup or teardown) in order.
// Setup (stopOnFailure=true) aborts at the first failed step — every scenario
// is then errored by the caller. Teardown (stopOnFailure=false) always runs
// every step: cleanups of independent resources must not shadow each other.
// The returned bool reports whether every step succeeded.
func (e *Engine) runSuiteSteps(ctx context.Context, steps []spec.Step, rt *suiteRuntime, rc runConfig, stopOnFailure bool) ([]StepResult, bool) {
	var out []StepResult
	var current *runner.Result
	ok := true

	for i := range steps {
		step := &steps[i]
		sr := StepResult{Index: i, Kind: step.Kind()}
		failed := false

		if ctx.Err() != nil {
			sr.ErrMsg = fmt.Sprintf("run canceled: %v", ctx.Err())
			out = append(out, sr)
			return out, false
		}

		switch step.Kind() {
		case spec.StepFixture:
			if err := fixture.Write(expandFixture(rt.st, step.Fixture), rt.dir, rc.specDir); err != nil {
				sr.ErrMsg = err.Error()
				failed = true
			}
		case spec.StepRun:
			run := mergeScenarioEnv(rt.env, expandRun(rt.st, step.Run), rt.st)
			r, untilChecks, err := e.runStep(ctx, run, rt.st, rt.dir, rc.specDir, rc, rt.sshConns, nil) // suite setup/teardown steps carry no changes assert
			if err != nil {
				sr.ErrMsg = err.Error()
				failed = true
				break
			}
			current = r
			sr.Run = maskResult(rc.masker, r)
			if len(untilChecks) > 0 {
				sr.Checks = untilChecks
				if !assert.AllOK(untilChecks) {
					failed = true
				}
			}
		case spec.StepStore:
			val, err := extractValue(expandStore(rt.st, step.Store), current, rt.dir)
			if err != nil {
				sr.ErrMsg = err.Error()
				failed = true
			} else {
				rt.set(step.Store.Name, val)
			}
		case spec.StepAssert:
			crs := assert.CheckAll(expandAssert(rt.st, step.Assert), current, assert.Env{
				Workdir:         rt.dir,
				SpecDir:         rc.specDir,
				UpdateSnapshots: e.UpdateSnapshots,
				Secrets:         rc.masker.MaskBytes,
				Scrub:           rc.scrubber.Apply,
				MockRecords: func(name string) ([]mockrunner.Record, bool) {
					for _, m := range rt.mocks {
						if m.Name() == name {
							return m.Records(), true
						}
					}
					return nil, false
				},
			})
			sr.Checks = crs
			if !assert.AllOK(crs) {
				failed = true
			}
		case spec.StepService:
			proc, captured, err := servicerunner.Start(ctx, expandService(rt.st, rt.env, step.Service), rt.dir)
			if proc != nil {
				rt.services = append(rt.services, proc)
			}
			if err != nil {
				sr.ErrMsg = err.Error()
				failed = true
				break
			}
			if step.Service.Ready != nil && step.Service.Ready.Store != "" {
				rt.set(step.Service.Ready.Store, captured)
			}
		case spec.StepMockServer:
			ms, err := mockrunner.Start(ctx, step.MockServer, rc.specDir)
			if err != nil {
				sr.ErrMsg = err.Error()
				failed = true
				break
			}
			rt.mocks = append(rt.mocks, ms)
			rt.set(ms.Name()+".url", ms.URL())
			rt.set(ms.Name()+".port", ms.Port())
		default:
			sr.ErrMsg = fmt.Sprintf("%s steps are not allowed at suite level", step.Kind())
			failed = true
		}

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
	out, _ := e.runSuiteSteps(tctx, s.Suite.Teardown, rt, rc, false)
	return out
}
