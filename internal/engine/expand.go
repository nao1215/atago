package engine

import (
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// The expand* helpers apply ${name} variable substitution to the
// user-controllable string fields of a step before it executes. They return
// copies so the original spec is never mutated. Each delegates to the step
// kind's Walk*Strings walker in internal/spec: the walker is the ONE list of
// what the kind expands, shared with the variable summaries and the security
// notes, so the engine cannot expand a field the summaries do not know about —
// the drift that let run's env, http's body_file/body_to/form/files, and cdp's
// upload/download go uncounted while the engine expanded them.

func expandRun(st *store.Store, r *spec.Run) *spec.Run {
	return spec.WalkRunStrings(r, st.Expand)
}

// mergeScenarioEnv layers the scenario-level env under a run step's own env so
// every command in a scenario shares a base environment (e.g. an isolated
// HOME/GOBIN) without repeating it on each step. The step's own env wins on key
// conflicts. Scenario env values are ${name}-expanded too. The passed run is
// already expanded and owned by the caller, so it is mutated in place.
func mergeScenarioEnv(scenarioEnv map[string]string, r *spec.Run, st *store.Store) *spec.Run {
	if len(scenarioEnv) == 0 {
		return r
	}
	merged := make(map[string]string, len(scenarioEnv)+len(r.Env))
	for k, v := range st.ExpandMap(scenarioEnv) {
		merged[k] = v
	}
	for k, v := range r.Env { // step env overrides scenario env
		merged[k] = v
	}
	r.Env = merged
	return r
}

// expandService applies ${name} substitution to a background service's fields
// (via the shared walker, which includes the readiness probes: a log regexp
// like `listening on ${workdir}/sock` is compiled verbatim by the service
// runner, so an unexpanded reference could never match). Scenario env is then
// layered under the service's own env, mirroring run steps.
func expandService(st *store.Store, scenarioEnv map[string]string, svc *spec.Service) *spec.Service {
	c := spec.WalkServiceStrings(svc, st.Expand)
	merged := make(map[string]string, len(scenarioEnv)+len(c.Env))
	for k, v := range st.ExpandMap(scenarioEnv) {
		merged[k] = v
	}
	for k, v := range c.Env { // service env overrides scenario env
		merged[k] = v
	}
	c.Env = merged
	return c
}

func expandFixture(st *store.Store, f *spec.Fixture) *spec.Fixture {
	return spec.WalkFixtureStrings(f, st.Expand)
}

func expandStore(st *store.Store, s *spec.Store) *spec.Store {
	return spec.WalkStoreStrings(s, st.Expand)
}

// expandAssert applies ${name} substitution to every interpolatable string in an
// assertion via the shared spec.WalkAssertStrings walker (issue #23).
func expandAssert(st *store.Store, a *spec.Assert) *spec.Assert {
	return spec.WalkAssertStrings(a, st.Expand)
}

// expandHTTP applies ${name} substitution to an http step's request fields —
// the declarative value-binding that lets a login response's token flow into a
// later authenticated request.
func expandHTTP(st *store.Store, h *spec.HTTP) *spec.HTTP {
	return spec.WalkHTTPStrings(h, st.Expand)
}

// expandHTTP's gRPC counterpart.
func expandGRPC(st *store.Store, g *spec.GRPC) *spec.GRPC {
	return spec.WalkGRPCStrings(g, st.Expand)
}

// expandCDP applies ${name} substitution to a cdp step's action arguments so a
// browser flow can reference stored values.
func expandCDP(st *store.Store, c *spec.CDP) *spec.CDP {
	return spec.WalkCDPStrings(c, st.Expand)
}

// mergedEnv layers base beneath own (own wins per key) without mutating either.
// It returns own unchanged when base is empty, so the common no-suite-env case
// allocates nothing.
func mergedEnv(base, own map[string]string) map[string]string {
	if len(base) == 0 {
		return own
	}
	out := make(map[string]string, len(base)+len(own))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range own {
		out[k] = v
	}
	return out
}
