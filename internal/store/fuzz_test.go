package store

import (
	"testing"
	"testing/quick"

	"github.com/nao1215/atago/internal/spec"
)

// FuzzExpand asserts the two robust invariants of ${name} expansion (issue #46):
//   - it never panics on arbitrary input, and
//   - expanding against an *empty* store never grows the string: with no known
//     variables the only edits are leaving live refs verbatim and collapsing a
//     $${name} escape to ${name} (which removes one byte).
func FuzzExpand(f *testing.F) {
	for _, seed := range []string{
		"", "plain text", "${a}", "$${a}", "${a}-${b}", "pid $$", "$$$${x}",
		"${", "}${", "${1bad}", "unicode ${café}", "${a}${a}${a}",
	} {
		f.Add(seed)
	}
	empty := New()
	f.Fuzz(func(t *testing.T, in string) {
		out := empty.Expand(in) // must not panic
		if len(out) > len(in) {
			t.Fatalf("empty-store Expand grew the string: %q (%d) -> %q (%d)", in, len(in), out, len(out))
		}
		// Populated store must not panic either, on the same input.
		s := New()
		s.Set("a", "AAA")
		s.Set("b", "BBB")
		_ = s.Expand(in)
	})
}

// TestExpand_Properties checks the interpolation contract over randomly
// generated identifiers and values via testing/quick (issue #46): a live
// reference expands to its value, an escaped reference stays literal, and an
// unknown reference is left verbatim.
func TestExpand_Properties(t *testing.T) {
	t.Parallel()

	// ident produces a valid ${name} identifier (matches the ${name} grammar).
	ident := func(n uint) string {
		const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
		const rest = alpha + "0123456789"
		name := string(alpha[n%uint(len(alpha))])
		n /= uint(len(alpha))
		for i := 0; i < int(n%6); i++ {
			name += string(rest[n%uint(len(rest))])
			n /= uint(len(rest))
		}
		return name
	}

	liveExpandsToValue := func(nameSeed uint, value string) bool {
		name := ident(nameSeed)
		s := New()
		s.Set(name, value)
		return s.Expand("${"+name+"}") == value
	}
	if err := quick.Check(liveExpandsToValue, nil); err != nil {
		t.Errorf("live reference did not expand to its value: %v", err)
	}

	escapedStaysLiteral := func(nameSeed uint, value string) bool {
		name := ident(nameSeed)
		s := New()
		s.Set(name, value) // even though it is set, the escape must win
		return s.Expand("$${"+name+"}") == "${"+name+"}"
	}
	if err := quick.Check(escapedStaysLiteral, nil); err != nil {
		t.Errorf("escaped reference did not stay literal: %v", err)
	}

	unknownStaysVerbatim := func(nameSeed uint) bool {
		name := ident(nameSeed)
		in := "before ${" + name + "} after"
		// The store is empty, so the reference is unknown and must be preserved.
		out := New().Expand(in)
		return out == in && len(spec.VarRefs(out)) == 1
	}
	if err := quick.Check(unknownStaysVerbatim, nil); err != nil {
		t.Errorf("unknown reference was not left verbatim: %v", err)
	}
}
