package assert

import (
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

// FuzzChangesGlob attacks the changes-assert glob layer (#70) with arbitrary
// entry patterns and observed paths:
//   - matchesAny/patternMatchesAny must never panic, whatever the pattern
//     (doublestar's ErrBadPattern is swallowed by design);
//   - self-match law: an entry in which firstUnescapedGlobMeta finds no
//     metacharacter (and that carries no backslash escapes) is, per the
//     globMetaNote contract, a literal filename — so it must match itself.
//     A violation means a literal changes entry fails with a baffling message
//     and NO explanatory note (the exact footgun globMetaNote exists to catch);
//   - suggestion law: when globMetaNote fires, the escaped spelling it prints
//     (escapeGlobMeta) must actually match the literal filename it was derived
//     from — a suggestion that itself fails to match is worse than none.
func FuzzChangesGlob(f *testing.F) {
	f.Add("out.txt", "out.txt")
	f.Add("site/**", "site/a/b.html")
	f.Add("data[1].json", "data[1].json")
	f.Add(`literal\*star`, "literal*star")
	f.Add("{a,b}.txt", "{a,b}.txt")
	f.Add("a\\", "a\\")
	f.Add(`\0*`, "0")
	f.Fuzz(func(t *testing.T, pat, path string) {
		_ = matchesAny([]string{pat}, path)        // must not panic
		_ = patternMatchesAny(pat, []string{path}) // must not panic
		_ = globMetaNote(pat)                      // must not panic

		if meta := firstUnescapedGlobMeta(pat); meta == 0 {
			if !strings.Contains(pat, `\`) && pat != "" {
				ok, err := doublestar.Match(pat, pat)
				if err == nil && !ok {
					t.Fatalf("entry %q has no metacharacter per firstUnescapedGlobMeta, yet does not match itself as a literal path", pat)
				}
			}
		} else {
			esc := escapeGlobMeta(pat)
			ok, err := doublestar.Match(esc, pat)
			if err != nil || !ok {
				t.Fatalf("globMetaNote suggests %q for literal %q, but it does not match (ok=%v err=%v)", esc, pat, ok, err)
			}
		}
	})
}
