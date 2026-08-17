package spec

import "regexp"

// varRef matches an optional escaping `$` (group 1) followed by a `${name}` or
// `${env:NAME}` interpolation reference (group 2, including the optional
// `env:` prefix). It is the single source of truth for what a variable
// reference looks like, shared by the engine's expansion and explain
// (issue #23). It must stay in lockstep with
// store.varRef, including the `$${name}` literal-escape handling (issue #37)
// and the dotted segments of namespaced built-ins like a mock server's
// `${<name>.url}`/`${<name>.port}` (issue #24) — without the `\.`-segment
// branch, VarRefs silently dropped every dotted reference, so `atago manifest`
// and explain under-reported which variables a scenario actually uses.
var varRef = regexp.MustCompile(`(\$?)\$\{((?:env:)?[a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z0-9_]+)*)\}`)

// VarRefs returns the variable names referenced by live ${name} occurrences in
// s. Escaped `$${name}` references are literal text, not live references, and
// are intentionally excluded (issue #37).
func VarRefs(s string) []string {
	matches := varRef.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return nil
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		if m[1] != "" {
			continue // escaped $${name}: literal, not a live reference
		}
		names = append(names, m[2])
	}
	if len(names) == 0 {
		return nil
	}
	return names
}

// CollectVars adds the variable names referenced by ${name} interpolation in
// each field to set. It is the shared accumulator behind the variable-usage
// summaries that explain and manifest each build over a scenario's fields.
func CollectVars(set map[string]bool, fields ...string) {
	for _, f := range fields {
		for _, name := range VarRefs(f) {
			set[name] = true
		}
	}
}

// collecting returns a walker visit that records each ${name} reference into
// set and leaves the string unchanged. Passing it to the same Walk*Strings
// walker the engine expands with is what keeps collection and expansion in
// exact lockstep, field for field.
func collecting(set map[string]bool) func(string) string {
	return func(s string) string {
		CollectVars(set, s)
		return s
	}
}

// CollectServiceVars folds every ${name} reference the engine would expand in a
// service into set — the walker is the field list, shared with expandService.
func CollectServiceVars(set map[string]bool, svc *Service) {
	WalkServiceStrings(svc, collecting(set))
}

// StepRunner returns the runner a step names, or "" for a kind that carries no
// runner. It is the one definition of that mapping: the variable walk and the
// security summary both need to reach a step's runner definition, and spelling
// the switch twice is how they would come to disagree about which steps have
// one.
func StepRunner(step *Step) string {
	switch step.Kind() {
	case StepRun:
		return step.Run.Runner
	case StepHTTP:
		return step.HTTP.Runner
	case StepQuery:
		return step.Query.Runner
	case StepGRPC:
		return step.GRPC.Runner
	case StepCDP:
		return step.CDP.Runner
	default:
		return ""
	}
}

// RunnerVarFields returns the runner fields the engine ${name}-expands at use
// time — a cmd runner's cwd, an http base_url, a db dsn, a grpc target, and an
// ssh runner's address and credentials. It is shared by the variable walk and
// the security summary so both read one list of what a runner interpolates.
func RunnerVarFields(r Runner) []string {
	return []string{r.Cwd, r.BaseURL, r.DSN, r.Target, r.Host, r.User, r.Password, r.KeyFile, r.KnownHosts}
}

// CollectRunnerVars folds every ${name} reference in a runner definition into
// set. A runner is expanded when a step uses it, so a `${env:DB_PASSWORD}` in a
// dsn is a genuine dependency of every scenario that queries through it.
func CollectRunnerVars(set map[string]bool, r Runner) {
	CollectVars(set, RunnerVarFields(r)...)
}

// CollectStepVars folds every ${name} reference the engine would expand in a
// step into set, including those in the definition of the runner it names.
// Each kind is read through the same Walk*Strings walker the engine expands
// with (internal/spec/stepstrings.go), so the field lists cannot drift apart —
// they had, repeatedly: explain under-reported run env, http
// body_file/body_to/form/files, cdp upload/download, and pty paste/exec, and
// the whole assert kind was uncounted, each because this function kept its own
// copy of what a kind expands. The retry `until:` assertion is the one
// addition on top of the walkers: the poll loop expands it per attempt through
// WalkAssertStrings, so its references are just as live.
func CollectStepVars(set map[string]bool, step *Step, runners map[string]Runner) {
	if name := StepRunner(step); name != "" {
		if r, ok := runners[name]; ok {
			CollectRunnerVars(set, r)
		}
	}
	collect := collecting(set)
	switch step.Kind() {
	case StepFixture:
		WalkFixtureStrings(step.Fixture, collect)
	case StepService:
		CollectServiceVars(set, step.Service)
	case StepRun:
		WalkRunStrings(step.Run, collect)
		collectRetryVars(set, step.Run.Retry)
	case StepHTTP:
		WalkHTTPStrings(step.HTTP, collect)
		collectRetryVars(set, step.HTTP.Retry)
	case StepQuery:
		WalkQueryStrings(step.Query, collect)
	case StepGRPC:
		WalkGRPCStrings(step.GRPC, collect)
	case StepCDP:
		WalkCDPStrings(step.CDP, collect)
	case StepPTY:
		WalkPTYStrings(step.PTY, collect)
	case StepSignal:
		WalkSignalStrings(step.Signal, collect)
	case StepAssert:
		WalkAssertStrings(step.Assert, collect)
	case StepStore:
		WalkStoreStrings(step.Store, collect)
	}
}

// collectRetryVars folds the ${name} references of a retry's until assertion
// into set. The engine expands it before each evaluation (pollUntil ->
// expandAssert), so a reference there is as live as one in the step's own
// fields; run and http share the retry shape and so share this walk.
func collectRetryVars(set map[string]bool, r *Retry) {
	if r == nil || r.Until == nil {
		return
	}
	WalkAssertStrings(r.Until, func(s string) string {
		CollectVars(set, s)
		return s
	})
}

// WalkPTYExpectScreenStrings returns a copy of a session-local expect_screen
// matcher with visit applied to every interpolatable matcher string. The timing
// knobs (timeout/stable_for) are durations, not variable-bearing text.
func WalkPTYExpectScreenStrings(es *PTYExpectScreen, visit func(string) string) *PTYExpectScreen {
	if es == nil {
		return nil
	}
	c := *es
	// Reuse the assert walker so the matcher surface stays in exact lockstep
	// with top-level `screen:` strings.
	a := WalkAssertStrings(&Assert{Screen: &es.ScreenAssert}, visit)
	if a.Screen != nil {
		c.ScreenAssert = *a.Screen
	}
	return &c
}

// WalkAssertStrings returns a deep copy of a in which visit has been applied to
// every interpolatable (${name}) string field — the single field list the
// engine's expansion needs (issue #23). Pass a substituting visit (e.g.
// store.Expand) to expand. Returns nil for a nil assert.
func WalkAssertStrings(a *Assert, visit func(string) string) *Assert {
	if a == nil {
		return nil
	}
	c := *a
	c.Stdout = walkStream(a.Stdout, visit)
	c.Stderr = walkStream(a.Stderr, visit)
	c.Body = walkStream(a.Body, visit)
	c.Rows = walkStream(a.Rows, visit)
	c.Message = walkStream(a.Message, visit)
	c.Value = walkStream(a.Value, visit)
	if a.File != nil {
		fc := *a.File
		fc.Path = visit(a.File.Path)
		fc.Contains = walkList(a.File.Contains, visit)
		fc.NotContains = walkList(a.File.NotContains, visit)
		fc.JSON = walkJSONChecks(a.File.JSON, visit)
		c.File = &fc
	}
	c.Header = walkHeaderMatch(a.Header, visit)
	if a.Image != nil {
		ic := *a.Image
		ic.Path = visit(a.Image.Path)
		ic.SimilarTo = visit(a.Image.SimilarTo)
		c.Image = &ic
	}
	// The rendered-screen target carries the stream matchers plus the attribute
	// entries; both hold author-written text that takes ${name} expansion.
	if a.Screen != nil {
		sc := *a.Screen
		if walked := walkStream(&a.Screen.StreamAssert, visit); walked != nil {
			sc.StreamAssert = *walked
		}
		if len(a.Screen.Attrs) > 0 {
			sc.Attrs = make([]ScreenAttr, len(a.Screen.Attrs))
			for i, at := range a.Screen.Attrs {
				ac := at
				ac.Text = visit(at.Text)
				sc.Attrs[i] = ac
			}
		}
		c.Screen = &sc
	}
	if a.Dir != nil {
		dc := *a.Dir
		dc.Path = visit(a.Dir.Path)
		dc.Contains = walkStrings(a.Dir.Contains, visit)
		dc.NotContains = walkStrings(a.Dir.NotContains, visit)
		dc.Glob = visit(a.Dir.Glob)
		dc.Ignore = walkStrings(a.Dir.Ignore, visit)
		c.Dir = &dc
	}
	if a.PDF != nil {
		pc := *a.PDF
		pc.Path = visit(a.PDF.Path)
		if len(a.PDF.Metadata) > 0 {
			m := make(map[string]string, len(a.PDF.Metadata))
			for k, v := range a.PDF.Metadata {
				m[k] = visit(v)
			}
			pc.Metadata = m
		}
		pc.Text = walkStream(a.PDF.Text, visit)
		c.PDF = &pc
	}
	if a.Mock != nil {
		mc := *a.Mock
		// Name stays untouched: it references a mock server declared in the
		// spec, checked against the declared set at load time.
		mc.Path = visit(a.Mock.Path)
		mc.Header = walkHeaderMatch(a.Mock.Header, visit)
		mc.Body = walkStream(a.Mock.Body, visit)
		c.Mock = &mc
	}
	if a.Changes != nil {
		cc := *a.Changes
		cc.Created = walkListPtr(a.Changes.Created, visit)
		cc.Modified = walkListPtr(a.Changes.Modified, visit)
		cc.Deleted = walkListPtr(a.Changes.Deleted, visit)
		c.Changes = &cc
	}
	return &c
}

// walkHeaderMatch returns a copy of a header matcher with visit applied to its
// matcher arguments (the header name is a protocol identifier, not user text).
func walkHeaderMatch(h *HeaderMatch, visit func(string) string) *HeaderMatch {
	if h == nil {
		return nil
	}
	c := *h
	c.Contains = walkPtr(h.Contains, visit)
	c.Equals = walkPtr(h.Equals, visit)
	c.Matches = walkPtr(h.Matches, visit)
	return &c
}

// walkStream returns a copy of a stream matcher with visit applied to its text
// fields and json/yaml matcher payloads. Empty/Line/Snapshot carry no
// interpolatable user text.
func walkStream(s *StreamAssert, visit func(string) string) *StreamAssert {
	if s == nil {
		return nil
	}
	c := *s
	c.Contains = walkList(s.Contains, visit)
	c.NotContains = walkList(s.NotContains, visit)
	c.Matches = walkPtr(s.Matches, visit)
	c.NotMatches = walkPtr(s.NotMatches, visit)
	c.Equals = walkPtr(s.Equals, visit)
	c.NotEquals = walkPtr(s.NotEquals, visit)
	c.JSON = walkJSONChecks(s.JSON, visit)
	c.YAML = walkJSONChecks(s.YAML, visit)
	return &c
}

// walkJSONChecks returns a copy of a json/yaml matcher list (#156) with visit
// applied to each check's path, regex, and expected value.
func walkJSONChecks(list JSONChecks, visit func(string) string) JSONChecks {
	if list == nil {
		return nil
	}
	out := make(JSONChecks, len(list))
	for i := range list {
		out[i] = *walkJSONAssert(&list[i], visit)
	}
	return out
}

// walkJSONAssert returns a copy of a json/yaml matcher with visit applied to its
// path, regex, and expected value.
func walkJSONAssert(j *JSONAssert, visit func(string) string) *JSONAssert {
	if j == nil {
		return nil
	}
	c := *j
	c.Path = visit(j.Path)
	c.Matches = walkPtr(j.Matches, visit)
	c.Equals = WalkJSONValueStrings(j.Equals, visit)
	return &c
}

// WalkJSONValueStrings returns a copy of a decoded JSON value (as produced for
// http/grpc request bodies or a json matcher's expected value) with visit
// applied to every string within it, leaving non-string scalars untouched.
func WalkJSONValueStrings(v any, visit func(string) string) any {
	switch t := v.(type) {
	case string:
		return visit(t)
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = WalkJSONValueStrings(val, visit)
		}
		return m
	case map[any]any:
		m := make(map[any]any, len(t))
		for k, val := range t {
			m[k] = WalkJSONValueStrings(val, visit)
		}
		return m
	case []any:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = WalkJSONValueStrings(val, visit)
		}
		return s
	default:
		return v
	}
}

// walkStrings applies visit to each element of a plain []string field (dir
// contains/not_contains/ignore), returning a new slice, or nil for nil input.
func walkStrings(l []string, visit func(string) string) []string {
	if l == nil {
		return nil
	}
	out := make([]string, len(l))
	for i, v := range l {
		out[i] = visit(v)
	}
	return out
}

// walkListPtr applies visit through a *StringList (the changes entry lists),
// preserving the nil-vs-empty distinction the exhaustive-set semantics depend
// on: nil stays nil (unconstrained) and an empty list stays empty (asserts the
// category changed nothing).
func walkListPtr(l *StringList, visit func(string) string) *StringList {
	if l == nil {
		return nil
	}
	out := walkList(*l, visit)
	if out == nil {
		out = StringList{}
	}
	return &out
}

// walkList applies visit to each element of a StringList, returning a new list,
// or nil when the input is nil (an unset contains/not_contains matcher).
func walkList(l StringList, visit func(string) string) StringList {
	if l == nil {
		return nil
	}
	out := make(StringList, len(l))
	for i, v := range l {
		out[i] = visit(v)
	}
	return out
}

// walkPtr applies visit to a *string field, returning a new pointer, or nil when
// the input is nil.
func walkPtr(p *string, visit func(string) string) *string {
	if p == nil {
		return nil
	}
	v := visit(*p)
	return &v
}
