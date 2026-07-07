package engine

import (
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// The expand* helpers apply ${name} variable substitution to the
// user-controllable string fields of a step before it executes. They return
// shallow copies so the original spec is never mutated.

func expandRun(st *store.Store, r *spec.Run) *spec.Run {
	c := *r
	c.Command = st.Expand(r.Command)
	c.Cwd = st.Expand(r.Cwd)
	// Stdin: inline text and the file path are ${name}-expanded; base64 is
	// deliberately not (binary payloads must stay byte-exact, mirroring the
	// fixture.base64 rule) (#18).
	c.Stdin.Inline = st.Expand(r.Stdin.Inline)
	c.Stdin.File = st.Expand(r.Stdin.File)
	// The stdout_to/stderr_to redirect targets are workdir-relative paths, just
	// like the assert paths that later read them; expand them so a per-matrix-row
	// or store-derived filename resolves instead of being written verbatim.
	c.StdoutTo = st.Expand(r.StdoutTo)
	c.StderrTo = st.Expand(r.StderrTo)
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
// matrix-bound variables. Scenario env is layered under the service's
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
		// The log-regexp readiness probe must be expanded too: a probe like
		// `log: "listening on ${workdir}/sock"` is compiled verbatim by the
		// service runner, so leaving ${workdir} unexpanded means it can never
		// match and the service always hits its readiness timeout.
		rc.Log = st.Expand(svc.Ready.Log)
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
// — the declarative value-binding that lets a login response's
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
// and JSON request body so a grpc call can reference stored values.
func expandGRPC(st *store.Store, g *spec.GRPC) *spec.GRPC {
	c := *g
	c.Method = st.Expand(g.Method)
	c.Header = st.ExpandMap(g.Header)
	c.JSON = spec.WalkJSONValueStrings(g.JSON, st.Expand)
	return &c
}

// expandCDP applies ${name} substitution to a cdp step's action arguments so a
// browser flow can reference stored values.
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
		if a.Upload != nil {
			// Upload.File is a workdir-relative path; expand it (and the selector)
			// so a ${workdir}/store-derived file is uploaded, matching what
			// CollectStepVars reports as a live variable use.
			up := *a.Upload
			up.Selector = st.Expand(up.Selector)
			up.File = st.Expand(up.File)
			a.Upload = &up
		}
		if a.Download != nil {
			// Download.Dir is a workdir-relative capture directory; expand it (and
			// the click selector) for the same reason as Upload above.
			dl := *a.Download
			dl.Click = st.Expand(dl.Click)
			dl.Dir = st.Expand(dl.Dir)
			a.Download = &dl
		}
		out.Actions[i] = a
	}
	return &out
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
