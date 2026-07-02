package spec

import "regexp"

// varRef matches an optional escaping `$` (group 1) followed by a `${name}` or
// `${env:NAME}` interpolation reference (group 2, including the optional
// `env:` prefix). It is the single source of truth for what a variable
// reference looks like, shared by the engine's expansion and explain
// (issue #23). It must stay in lockstep with
// store.varRef, including the `$${name}` literal-escape handling (issue #37).
var varRef = regexp.MustCompile(`(\$?)\$\{((?:env:)?[a-zA-Z_][a-zA-Z0-9_]*)\}`)

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
	if a.Header != nil {
		hc := *a.Header
		hc.Contains = walkPtr(a.Header.Contains, visit)
		hc.Equals = walkPtr(a.Header.Equals, visit)
		c.Header = &hc
	}
	if a.Image != nil {
		ic := *a.Image
		ic.Path = visit(a.Image.Path)
		ic.SimilarTo = visit(a.Image.SimilarTo)
		c.Image = &ic
	}
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
