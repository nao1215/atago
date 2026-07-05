package engine

import (
	"context"
	"time"

	"github.com/nao1215/atago/internal/fixture"
	mockrunner "github.com/nao1215/atago/internal/runner/mock"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// initStore seeds the scenario's variable store and resolves its effective env.
// Order matters: built-ins first, then suite-level values (#7), ${workdir}, and
// finally the matrix row so the most specific values win. suite.env layers
// beneath the scenario's own env (the scenario wins per key).
func (x *scenarioRun) initStore() {
	st := store.New()
	for k, v := range x.e.builtins {
		st.Set(k, v)
	}
	// Suite-level values (#7) come before the scenario's own: ${suitedir} and
	// suite.setup captures are shared context every scenario may reference.
	for k, v := range x.rc.suiteVars {
		st.Set(k, v)
	}
	// ${workdir} is the absolute path of this scenario's isolated temp dir, so
	// specs can build absolute env paths (e.g. HOME=${workdir}/home,
	// GOBIN=${workdir}/gobin) that child toolchains require.
	st.Set("workdir", x.workdir)
	// Seed matrix row variables so ${name} references in commands, env, and
	// assertions resolve to this instance's values.
	for k, v := range x.sc.Vars {
		st.Set(k, v)
	}
	x.st = st
	x.scEnv = mergedEnv(x.rc.suiteEnv, x.sc.Env)
}

// startMocks starts the scenario's mock servers (#24) before the leading
// fixtures and services, so fixture contents and service commands/env can
// reference ${<mock>.url}. Each binds an ephemeral loopback port and seeds
// ${<name>.url} / ${<name>.port} into the store; they stop LIFO with the
// scenario. It reports false (after recording the error) if any fails to start.
func (x *scenarioRun) startMocks(ctx context.Context) bool {
	for i := range x.sc.MockServers {
		ms, err := mockrunner.Start(ctx, &x.sc.MockServers[i], x.specDir)
		if err != nil {
			x.out.Status = StatusError
			x.out.Steps = append(x.out.Steps, StepResult{Kind: spec.StepNone, Setup: true, ErrMsg: x.masker.Mask(err.Error())})
			x.out.Duration = time.Since(x.start)
			return false
		}
		x.mocks = append(x.mocks, ms)
		x.st.Set(ms.Name()+".url", ms.URL())
		x.st.Set(ms.Name()+".port", ms.Port())
	}
	return true
}

// stopMocks stops the scenario's mock servers LIFO.
func (x *scenarioRun) stopMocks() {
	for i := len(x.mocks) - 1; i >= 0; i-- {
		x.mocks[i].Stop()
	}
}

// applyLeadingFixtures writes the uninterrupted prefix of fixture steps BEFORE
// services start, so a background server can consume authored input (its config
// file, seed data) the way a real daemon does. Fixtures after the first
// non-fixture step keep their in-order, after-services timing, so a scenario can
// still simulate files appearing while the service runs. It returns the number
// of leading fixtures applied and false if one failed (already recorded).
func (x *scenarioRun) applyLeadingFixtures() (int, bool) {
	n := 0
	for n < len(x.sc.Steps) && x.sc.Steps[n].Kind() == spec.StepFixture {
		fx := x.sc.Steps[n].Fixture
		sr := StepResult{Index: n, Kind: spec.StepFixture}
		if err := fixture.Write(expandFixture(x.st, fx), x.workdir, x.specDir); err != nil {
			sr.ErrMsg = err.Error()
			x.out.Steps = append(x.out.Steps, sr)
			x.out.Status = StatusError
			x.out.Duration = time.Since(x.start)
			return n, false
		}
		x.out.Steps = append(x.out.Steps, sr)
		n++
	}
	return n, true
}

// startServices starts the scenario's background services after the store is
// seeded (so their commands can reference ${workdir} and matrix vars) and after
// the leading fixtures, but before any other step runs. They are stopped LIFO
// when the scenario ends, however it ends. It reports false if one fails to
// become ready, preserving the failed service's log as a durable artifact (#51).
func (x *scenarioRun) startServices(ctx context.Context) bool {
	for i := range x.sc.Services {
		proc, captured, err := servicerunner.Start(ctx, expandService(x.st, x.scEnv, &x.sc.Services[i]), x.workdir)
		if err != nil {
			x.out.Status = StatusError
			// Preserve the failed service's log (and any peers already started) as a
			// durable artifact before tearing down, so a readiness failure stays
			// inspectable (#51). proc is the stopped-but-readable failed service.
			if proc != nil {
				x.services = append(x.services, proc)
			}
			x.e.writeServiceLogs(&x.out, x.masker, x.services, x.rc.specPath, x.sc.Name, x.idx)
			// The readiness error can embed the service's raw output, so mask secrets
			// before it reaches a report (issue #12).
			x.out.Steps = append(x.out.Steps, StepResult{Kind: spec.StepNone, Setup: true, ErrMsg: x.masker.Mask(err.Error())})
			x.out.Duration = time.Since(x.start)
			return false
		}
		x.services = append(x.services, proc)
		if x.sc.Services[i].Ready != nil && x.sc.Services[i].Ready.Store != "" {
			x.st.Set(x.sc.Services[i].Ready.Store, captured)
		}
	}
	return true
}

// stopServices stops the scenario's background services LIFO.
func (x *scenarioRun) stopServices() {
	for i := len(x.services) - 1; i >= 0; i-- {
		x.services[i].Stop()
	}
}

// closeConns closes the lazily-opened per-scenario connections (db, ssh, grpc,
// browser), each opened on first use and closed once at scenario end.
func (x *scenarioRun) closeConns() {
	for _, c := range x.dbConns {
		_ = c.Close()
	}
	for _, c := range x.sshConns {
		_ = c.Close()
	}
	for _, c := range x.grpcConns {
		_ = c.Close()
	}
	for _, c := range x.browserConns {
		_ = c.Close()
	}
}

// mockRecords resolves a mock server's recorded requests for the `mock:`
// assertion target (#24): the scenario's own mocks first, then suite-wide ones,
// mirroring the store's scenario-over-suite precedence.
func (x *scenarioRun) mockRecords(name string) ([]mockrunner.Record, bool) {
	for _, m := range x.mocks {
		if m.Name() == name {
			return m.Records(), true
		}
	}
	for _, m := range x.rc.suiteMocks {
		if m.Name() == name {
			return m.Records(), true
		}
	}
	return nil, false
}
