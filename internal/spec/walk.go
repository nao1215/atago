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

// CollectServiceVars folds every ${name} reference in a service's fields
// (command, cwd, env values) into set. Shared by explain and manifest so their
// "variables used" summaries agree on a scenario's services.
func CollectServiceVars(set map[string]bool, svc *Service) {
	CollectVars(set, svc.Command, svc.Cwd)
	for _, v := range svc.Env {
		CollectVars(set, v)
	}
}

// CollectStepVars folds every ${name} reference in a step's fields into set.
// It is the single source of truth for which fields of each step kind carry
// variables, so the explain summary and the manifest never disagree about a
// step's referenced variables (they previously drifted: explain under-reported
// run env, http body_file/body_to/form/files, and cdp upload/download).
func CollectStepVars(set map[string]bool, step *Step) {
	switch step.Kind() {
	case StepFixture:
		CollectVars(set, step.Fixture.File, step.Fixture.Content, step.Fixture.Symlink)
	case StepService:
		CollectVars(set, step.Service.Command, step.Service.Cwd)
	case StepRun:
		r := step.Run
		CollectVars(set, r.Command, r.Cwd, r.Stdin.Inline, r.Stdin.File)
		for _, v := range r.Env {
			CollectVars(set, v)
		}
	case StepHTTP:
		h := step.HTTP
		CollectVars(set, h.Path, h.Body, h.BodyFile, h.BodyTo)
		for _, v := range h.Form {
			CollectVars(set, v)
		}
		for _, f := range h.Files {
			CollectVars(set, f.Path)
		}
	case StepQuery:
		CollectVars(set, step.Query.SQL)
	case StepGRPC:
		CollectVars(set, step.GRPC.Method)
	case StepCDP:
		for _, a := range step.CDP.Actions {
			collectCDPActionVars(set, a)
		}
	case StepPTY:
		pt := step.PTY
		CollectVars(set, pt.Command, pt.Cwd)
		for _, v := range pt.Env {
			CollectVars(set, v)
		}
		for _, a := range pt.Session {
			if a.Send != nil && a.Send.Text != nil {
				CollectVars(set, *a.Send.Text)
			}
			CollectVars(set, a.Expect)
		}
	case StepSignal:
		CollectVars(set, step.Signal.Service)
	}
}

// collectCDPActionVars folds the ${name} references of one browser action into
// set, covering every field the action set can carry.
func collectCDPActionVars(set map[string]bool, a CDPAction) {
	CollectVars(set, a.Navigate, a.WaitVisible, a.WaitHidden, a.Click, a.Check, a.Uncheck, a.Text, a.Eval)
	if a.SendKeys != nil {
		CollectVars(set, a.SendKeys.Selector, a.SendKeys.Value)
	}
	if a.Press != nil {
		CollectVars(set, a.Press.Selector, a.Press.Key)
	}
	if a.Select != nil {
		CollectVars(set, a.Select.Selector, a.Select.Value)
	}
	if a.Screenshot != nil {
		CollectVars(set, a.Screenshot.Path, a.Screenshot.Selector)
	}
	if a.Attribute != nil {
		CollectVars(set, a.Attribute.Selector, a.Attribute.Name)
	}
	if a.Upload != nil {
		CollectVars(set, a.Upload.Selector, a.Upload.File)
	}
	if a.Download != nil {
		CollectVars(set, a.Download.Click, a.Download.Dir)
	}
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
		fc.JSON = walkJSONAssert(a.File.JSON, visit)
		c.File = &fc
	}
	c.Header = walkHeaderMatch(a.Header, visit)
	if a.Image != nil {
		ic := *a.Image
		ic.Path = visit(a.Image.Path)
		ic.SimilarTo = visit(a.Image.SimilarTo)
		c.Image = &ic
	}
	// The rendered-screen matcher is a stream matcher like stdout/stderr.
	c.Screen = walkStream(a.Screen, visit)
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
	c.JSON = walkJSONAssert(s.JSON, visit)
	c.YAML = walkJSONAssert(s.YAML, visit)
	return &c
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
