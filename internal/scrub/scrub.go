// Package scrub applies user-declared regex→placeholder rewrites to captured
// CLI output before snapshot comparison (#137). Where internal/security's Masker
// redacts known secret *values* and internal/snapshot's built-in normalizers
// fold a fixed set of volatile forms (ANSI, UUID, timestamp, port, path), a
// Scrubber handles the open set of volatile *patterns* only the spec author
// knows about: auto-increment IDs, request identifiers, custom timestamps. This
// is atago's declarative "output determinization layer" — the piece that turns a
// snapshot that flakes on every run into a stable golden.
package scrub

import (
	"fmt"
	"regexp"

	"github.com/nao1215/atago/internal/spec"
)

// Scrubber applies an ordered list of compiled rewrite rules. A nil Scrubber is
// a valid no-op, so callers may hold a possibly-nil value and always call Apply.
type Scrubber struct {
	rules []compiledRule
}

type compiledRule struct {
	re          *regexp.Regexp
	placeholder []byte
}

// New compiles the rules in order. It returns (nil, nil) for an empty rule set —
// a nil Scrubber is a safe no-op — and an error naming the offending index and
// pattern when a rule's regexp does not compile.
func New(rules []spec.ScrubRule) (*Scrubber, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	compiled := make([]compiledRule, 0, len(rules))
	for i, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("scrub[%d]: invalid pattern %q: %w", i, r.Pattern, err)
		}
		compiled = append(compiled, compiledRule{re: re, placeholder: []byte(r.Placeholder)})
	}
	return &Scrubber{rules: compiled}, nil
}

// Apply rewrites b through every rule in order and returns the result. Rules
// apply top-to-bottom: a later rule sees the output of the earlier ones. The
// placeholder is inserted literally — a `$1` in it is NOT a regexp expansion
// reference — so a placeholder can safely contain any bytes. A nil/empty
// Scrubber returns b unchanged.
func (s *Scrubber) Apply(b []byte) []byte {
	if s == nil || len(s.rules) == 0 {
		return b
	}
	for _, r := range s.rules {
		b = r.re.ReplaceAllLiteral(b, r.placeholder)
	}
	return b
}
