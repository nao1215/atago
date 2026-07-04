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
	pats := []string(*entries)
	for _, obs := range observed {
		if !matchesAny(pats, obs) {
			*problems = append(*problems, fmt.Sprintf("unexpected %s file %q (no entry matches it)", name, obs))
		}
	}
	for _, pat := range pats {
		if !patternMatchesAny(pat, observed) {
			*problems = append(*problems, fmt.Sprintf("%s entry %q matched no file the step %s", name, pat, name))
		}
	}
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
