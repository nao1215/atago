package assert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkChanges evaluates the workdir-delta assertion target (#70): the exact
// set of files the immediately preceding run/pty step created, modified, and
// deleted. Each set field is EXHAUSTIVE in both directions — every observed
// path must be matched by an entry, and every entry must match at least one
// observed path — so `modified: []` asserts "modified nothing". An omitted
// (nil) field is unconstrained. Paths are compared with forward slashes, so the
// same spec passes on Windows.
func checkChanges(c *spec.ChangesAssert, res *runner.Result) *CheckResult {
	desc := "assert changes"
	if res == nil || res.Changes == nil {
		return &CheckResult{
			Desc: desc,
			Hint: "no workdir delta was recorded; a changes assert must immediately follow a run/pty step",
		}
	}
	d := res.Changes

	var problems []string
	checkCategory("created", c.Created, d.Created, &problems)
	checkCategory("modified", c.Modified, d.Modified, &problems)
	checkCategory("deleted", c.Deleted, d.Deleted, &problems)

	if len(problems) == 0 {
		return pass(desc)
	}
	sort.Strings(problems)
	return &CheckResult{
		Desc:     desc,
		Expected: describeChangesExpected(c),
		Actual:   describeChangesActual(d),
		Hint:     strings.Join(problems, "; "),
	}
}

// checkCategory enforces the exhaustive-set semantics for one category. A nil
// entries pointer leaves the category unconstrained; a non-nil (possibly empty)
// list requires an exact, bidirectional cover.
func checkCategory(name string, entries *spec.StringList, observed []string, problems *[]string) {
	if entries == nil {
		return
	}
	// Normalize a single leading "./": observed paths are workdir-relative
	// without it, so "./out.txt" must match the observed "out.txt". Safe for
	// globs like "site/**".
	pats := make([]string, len(*entries))
	for i, p := range *entries {
		pats[i] = strings.TrimPrefix(p, "./")
	}
	for _, obs := range observed {
		if !matchesAny(pats, obs) {
			*problems = append(*problems, fmt.Sprintf("unexpected %s file %q (no entry matches it)", name, obs))
		}
	}
	for _, pat := range pats {
		if !patternMatchesAny(pat, observed) {
			msg := fmt.Sprintf("%s entry %q matched no file the step %s", name, pat, name)
			if note := globMetaNote(pat); note != "" {
				msg += "; " + note
			}
			*problems = append(*problems, msg)
		}
	}
}

// globMetaNote explains a common footgun: a changes entry is a doublestar glob,
// so an unescaped `[ ] * ?` is a metacharacter, not a literal filename byte. When
// such an entry matches nothing the failure is otherwise baffling (the Expected
// and Actual read identically), so we point at the first metacharacter and show
// the escaped spelling that would match a literal filename. It returns "" when
// the entry has no unescaped metacharacter.
func globMetaNote(pat string) string {
	meta := firstUnescapedGlobMeta(pat)
	if meta == 0 {
		return ""
	}
	return fmt.Sprintf(`note: %q is a glob metacharacter — write "%s" to match a literal filename`, string(meta), escapeGlobMeta(pat))
}

// firstUnescapedGlobMeta returns the first unescaped glob metacharacter in pat,
// or 0 when there is none. A backslash escapes the following byte, so an entry
// the author already escaped correctly (e.g. `a\{1\}.txt`) reports none and gets
// no note. The set matches doublestar's metacharacters: the `{ }` brace
// alternation must be here too, or a literal file named `{1}.txt` can never be
// matched by its own name and the failure stays baffling with no note.
func firstUnescapedGlobMeta(pat string) byte {
	for i := 0; i < len(pat); i++ {
		if pat[i] == '\\' {
			i++ // skip the escaped byte
			continue
		}
		switch pat[i] {
		case '[', ']', '*', '?', '{', '}':
			return pat[i]
		}
	}
	return 0
}

// escapeGlobMeta returns a doublestar pattern that matches pat as a literal
// filename by backslash-escaping every byte doublestar treats specially —
// including the backslash itself. Escaping the backslash matters: without it a
// pat that already contains one (e.g. `\0*`) would produce `\0\*`, which matches
// the 2-byte name `0*`, not the literal `\0*` the author typed. A suggestion
// that fails to match the very name it was derived from is worse than none.
func escapeGlobMeta(pat string) string {
	var b strings.Builder
	for i := 0; i < len(pat); i++ {
		switch pat[i] {
		case '\\', '[', ']', '*', '?', '{', '}':
			b.WriteByte('\\')
		}
		b.WriteByte(pat[i])
	}
	return b.String()
}

// matchesAny reports whether p matches at least one pattern (exact path, a
// single-level `*` glob, or a `**` doublestar that crosses `/`), always
// /-separated. A backslash escapes a literal metacharacter (`\[`, `\?`, `\*`).
func matchesAny(pats []string, p string) bool {
	for _, pat := range pats {
		if ok, _ := doublestar.Match(pat, p); ok {
			return true
		}
	}
	return false
}

// patternMatchesAny reports whether pat matches at least one observed path.
func patternMatchesAny(pat string, paths []string) bool {
	for _, p := range paths {
		if ok, _ := doublestar.Match(pat, p); ok {
			return true
		}
	}
	return false
}

// describeChangesExpected renders the asserted delta for the failure block.
func describeChangesExpected(c *spec.ChangesAssert) string {
	var parts []string
	if c.Created != nil {
		parts = append(parts, "created "+bracket(*c.Created))
	}
	if c.Modified != nil {
		parts = append(parts, "modified "+bracket(*c.Modified))
	}
	if c.Deleted != nil {
		parts = append(parts, "deleted "+bracket(*c.Deleted))
	}
	return strings.Join(parts, ", ")
}

// describeChangesActual renders the observed delta for the failure block.
func describeChangesActual(d *fsdelta.Delta) string {
	return fmt.Sprintf("created %s, modified %s, deleted %s",
		bracket(d.Created), bracket(d.Modified), bracket(d.Deleted))
}

// bracket renders a path list as "[a, b]" (or "[]" when empty).
func bracket(items []string) string {
	return "[" + strings.Join(items, ", ") + "]"
}
