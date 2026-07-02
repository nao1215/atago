// Package store holds scenario variables and performs ${name} expansion
// . Expansion is deliberately simple substitution — atago is not a
// programming language.
package store

import (
	"regexp"
)

// varRef matches an optional escaping `$` (group 1) followed by a `${name}`
// reference (group 2). When the leading `$` is present — i.e. the source was
// `$${name}` — the match is a literal escape that renders as `${name}` without
// expansion. A bare `$$` not followed by `{name}` (a
// shell PID, a doubled currency sign) does not match and is left untouched.
var varRef = regexp.MustCompile(`(\$?)\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// Store is a per-scenario variable map.
type Store struct {
	vars map[string]string
}

// New returns an empty Store.
func New() *Store { return &Store{vars: make(map[string]string)} }

// Set assigns a variable.
func (s *Store) Set(name, value string) { s.vars[name] = value }

// Get returns a variable and whether it was set.
func (s *Store) Get(name string) (string, bool) {
	v, ok := s.vars[name]
	return v, ok
}

// Expand replaces ${name} references with stored values. Unknown references are
// left verbatim so they surface as obvious failures rather than empty strings.
func (s *Store) Expand(in string) string {
	if s == nil || !varRef.MatchString(in) {
		return in
	}
	return varRef.ReplaceAllStringFunc(in, func(m string) string {
		sub := varRef.FindStringSubmatch(m)
		escaped, name := sub[1], sub[2]
		if escaped != "" {
			// $${name} → literal ${name}, never expanded.
			return "${" + name + "}"
		}
		if v, ok := s.vars[name]; ok {
			return v
		}
		return m
	})
}

// Unresolved returns the names of ${name} references in in that no stored
// variable resolves. Escaped $${name} literals are not reported — the author
// explicitly asked for literal text. Callers use this to turn a reference that
// nothing could ever expand into an explained failure instead of passing the
// literal text on.
func (s *Store) Unresolved(in string) []string {
	var names []string
	for _, m := range varRef.FindAllStringSubmatch(in, -1) {
		escaped, name := m[1], m[2]
		if escaped != "" {
			continue
		}
		if _, ok := s.vars[name]; !ok {
			names = append(names, name)
		}
	}
	return names
}

// ExpandMap returns a copy of m with all values expanded.
func (s *Store) ExpandMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return m
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = s.Expand(v)
	}
	return out
}
