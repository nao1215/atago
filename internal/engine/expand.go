package engine

import (
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// The expand* helpers apply ${name} variable substitution (spec.md §18) to the
// user-controllable string fields of a step before it executes. They return
// shallow copies so the original spec is never mutated.

func expandRun(st *store.Store, r *spec.Run) *spec.Run {
	c := *r
	c.Command = st.Expand(r.Command)
	c.Cwd = st.Expand(r.Cwd)
	c.Stdin = st.Expand(r.Stdin)
	c.Env = st.ExpandMap(r.Env)
	return &c
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

// expandService applies ${name} substitution to a background service's command,
// cwd, env, and readiness file/port so a service can reference ${workdir} and
// matrix-bound variables (ADR-0031). Scenario env is layered under the service's
// own env, mirroring run steps.
func expandService(st *store.Store, scenarioEnv map[string]string, svc *spec.Service) *spec.Service {
	c := *svc
	c.Command = st.Expand(svc.Command)
	c.Cwd = st.Expand(svc.Cwd)
	merged := make(map[string]string, len(scenarioEnv)+len(svc.Env))
	for k, v := range st.ExpandMap(scenarioEnv) {
		merged[k] = v
	}
	for k, v := range st.ExpandMap(svc.Env) { // service env overrides scenario env
		merged[k] = v
	}
	c.Env = merged
	if svc.Ready != nil {
		rc := *svc.Ready
		rc.File = st.Expand(svc.Ready.File)
		rc.Port = st.Expand(svc.Ready.Port)
		c.Ready = &rc
	}
	return &c
}

func expandFixture(st *store.Store, f *spec.Fixture) *spec.Fixture {
	c := *f
	c.File = st.Expand(f.File)
	c.Content = st.Expand(f.Content)
	c.From = st.Expand(f.From)
	c.Symlink = st.Expand(f.Symlink)
	return &c
}

func expandStore(st *store.Store, s *spec.Store) *spec.Store {
	if s.From == nil || s.From.File == nil {
		return s
	}
	c := *s
	fromCopy := *s.From
	fileCopy := *s.From.File
	fileCopy.Path = st.Expand(s.From.File.Path)
	fromCopy.File = &fileCopy
	c.From = &fromCopy
	return &c
}

// expandAssert applies ${name} substitution to every interpolatable string in an
// assertion via the shared spec.WalkAssertStrings walker (issue #23).
func expandAssert(st *store.Store, a *spec.Assert) *spec.Assert {
	return spec.WalkAssertStrings(a, st.Expand)
}

// expandHTTP applies ${name} substitution to an http step's path, header values,
// and JSON body so requests can reference stored values from earlier steps
// (spec.md §18) — the declarative value-binding that lets a login response's
// token flow into a later authenticated request.
func expandHTTP(st *store.Store, h *spec.HTTP) *spec.HTTP {
	c := *h
	c.Path = st.Expand(h.Path)
	c.Header = st.ExpandMap(h.Header)
	c.JSON = spec.WalkJSONValueStrings(h.JSON, st.Expand)
	c.Body = st.Expand(h.Body)
	c.BodyFile = st.Expand(h.BodyFile)
	c.BodyTo = st.Expand(h.BodyTo)
	c.Form = st.ExpandMap(h.Form)
	if len(h.Files) > 0 {
		c.Files = make([]spec.FilePart, len(h.Files))
		for i, f := range h.Files {
			f.Path = st.Expand(f.Path)
			c.Files[i] = f
		}
	}
	return &c
}

// expandHTTP's gRPC counterpart: expand ${name} in the method, header values,
// and JSON request body so a grpc call can reference stored values (ADR-0028).
func expandGRPC(st *store.Store, g *spec.GRPC) *spec.GRPC {
	c := *g
	c.Method = st.Expand(g.Method)
	c.Header = st.ExpandMap(g.Header)
	c.JSON = spec.WalkJSONValueStrings(g.JSON, st.Expand)
	return &c
}

// expandCDP applies ${name} substitution to a cdp step's action arguments so a
// browser flow can reference stored values (ADR-0029).
func expandCDP(st *store.Store, c *spec.CDP) *spec.CDP {
	out := *c
	out.Actions = make([]spec.CDPAction, len(c.Actions))
	for i, a := range c.Actions {
		a.Navigate = st.Expand(a.Navigate)
		a.WaitVisible = st.Expand(a.WaitVisible)
		a.WaitHidden = st.Expand(a.WaitHidden)
		a.Click = st.Expand(a.Click)
		a.Check = st.Expand(a.Check)
		a.Uncheck = st.Expand(a.Uncheck)
		a.Text = st.Expand(a.Text)
		a.Eval = st.Expand(a.Eval)
		if a.SendKeys != nil {
			sk := *a.SendKeys
			sk.Selector = st.Expand(sk.Selector)
			sk.Value = st.Expand(sk.Value)
			a.SendKeys = &sk
		}
		if a.Press != nil {
			p := *a.Press
			p.Selector = st.Expand(p.Selector)
			p.Key = st.Expand(p.Key)
			a.Press = &p
		}
		if a.Select != nil {
			s := *a.Select
			s.Selector = st.Expand(s.Selector)
			s.Value = st.Expand(s.Value)
			a.Select = &s
		}
		if a.Screenshot != nil {
			s := *a.Screenshot
			s.Path = st.Expand(s.Path)
			s.Selector = st.Expand(s.Selector)
			a.Screenshot = &s
		}
		if a.Attribute != nil {
			at := *a.Attribute
			at.Selector = st.Expand(at.Selector)
			at.Name = st.Expand(at.Name)
			a.Attribute = &at
		}
		out.Actions[i] = a
	}
	return &out
}
