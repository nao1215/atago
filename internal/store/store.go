// Package store holds scenario variables and performs ${name} expansion.
// Expansion is deliberately simple substitution — atago is not a
// programming language.
package store

import (
	"os"
	"regexp"
)

// varRef matches an optional escaping `$` (group 1) followed by a `${name}` or
// `${env:NAME}` reference (group 2 carries the name including the optional
// `env:` prefix). When the leading `$` is present — i.e. the source was
// `$${name}` — the match is a literal escape that renders as `${name}` without
// expansion. A bare `$$` not followed by `{name}` (a shell PID, a doubled
// currency sign) does not match and is left untouched.
// Dotted segments (${api.url}) exist for namespaced built-ins like the mock
// server's url/port seeds (#24); a name never starts or ends with a dot.
var varRef = regexp.MustCompile(`(\$?)\$\{((?:env:)?[a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z0-9_]+)*)\}`)

// envPrefix marks a reference resolved from the host environment instead of
// the scenario store: `${env:HOME}` expands to os.Getenv("HOME"). It exists
// for fields no shell ever touches (an http runner's base_url or headers, a
// db dsn, ssh credentials), where injecting a CI-provided value previously
// required a three-step shell/store dance.
const envPrefix = "env:"

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

// resolve looks up one reference name: an `env:`-prefixed name resolves from
// the host environment (set-but-empty expands to ""), anything else from the
// scenario store. Env names never fall back to store variables or vice versa.
func (s *Store) resolve(name string) (string, bool) {
	if envName, isEnv := cutEnvPrefix(name); isEnv {
		return os.LookupEnv(envName)
	}
	v, ok := s.vars[name]
	return v, ok
}

// cutEnvPrefix splits the `env:` marker off a reference name.
func cutEnvPrefix(name string) (string, bool) {
	if len(name) > len(envPrefix) && name[:len(envPrefix)] == envPrefix {
		return name[len(envPrefix):], true
	}
	return name, false
}

// Expand replaces ${name} references with stored values and ${env:NAME}
// references with host environment variables. Unknown references are left
// verbatim so they surface as obvious failures rather than empty strings.
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
		if v, ok := s.resolve(name); ok {
			return v
		}
		return m
	})
}

// ExpandDetectingLeaks expands in like Expand, and additionally reports the
// names of any ${...} references that a *substituted value* carries into the
// output. Expansion is deliberately single-pass (see Expand): a reference that
// a store/matrix value contains is never re-examined, so it survives verbatim
// into the result — and, for a no-shell run.command, into argv, where nothing
// will ever expand it. Reporting these lets the run guard refuse the garbled
// command instead of leaking the literal text (#249).
//
// Only references introduced by a substituted value are reported. A reference
// the author wrote directly in in is not (an unresolved one is already the
// pre-expansion guard's job, and a resolved one expands normally), and an
// escaped $${name} — in the input or inside a substituted value — is a
// deliberate literal, never a leak.
func (s *Store) ExpandDetectingLeaks(in string) (string, []string) {
	if s == nil || !varRef.MatchString(in) {
		return in, nil
	}
	var leaked []string
	out := varRef.ReplaceAllStringFunc(in, func(m string) string {
		sub := varRef.FindStringSubmatch(m)
		escaped, name := sub[1], sub[2]
		if escaped != "" {
			return "${" + name + "}"
		}
		v, ok := s.resolve(name)
		if !ok {
			return m // authored unresolved reference: pre-expansion guard's domain
		}
		// The substituted value is inserted verbatim (single-pass); any live
		// reference it carries will never be expanded and would leak into argv.
		// An escaped $${x} inside the value is a deliberate literal, so skip it.
		for _, vm := range varRef.FindAllStringSubmatch(v, -1) {
			if vm[1] != "" {
				continue
			}
			leaked = append(leaked, vm[2])
		}
		return v
	})
	return out, leaked
}

// Escape rewrites text so that Expand returns it verbatim: it prefixes an extra
// `$` onto exactly the references Expand acts on — a live `${name}` becomes the
// literal `$${name}`, and an already-escaped `$${name}` becomes `$$${name}` —
// while a `${` that Expand ignores (one not followed by a valid name, e.g.
// `${1}` or `${}`) is left untouched, because Expand would already pass it
// through. It is the exact inverse Expand relies on, so raw observed text (a
// recorded command, output anchor, or typed pty input) can be embedded in a spec
// without being re-expanded at replay. A blind `${`→`$${` rewrite is wrong: the
// expander only unescapes `$${<valid-name>}`, so `$${1}` would never round-trip
// back to the observed `${1}`.
func Escape(s string) string {
	return varRef.ReplaceAllStringFunc(s, func(m string) string { return "$" + m })
}

// Unresolved returns the names of ${name} references in in that no stored
// variable resolves, and of ${env:NAME} references whose environment variable
// is not set. Escaped $${name} literals are not reported — the author
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
		if _, ok := s.resolve(name); !ok {
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
