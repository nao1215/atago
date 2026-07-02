package assert

import (
	"fmt"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// firstMissing returns the first substring in subs that is NOT present in got,
// its 0-based index, and true; it returns ok=false when every element is
// present. It backs the `contains` matcher's "all must be present" semantics
// while still letting the caller report exactly which element failed.
func firstMissing(got string, subs spec.StringList) (sub string, idx int, ok bool) {
	for i, s := range subs {
		if !strings.Contains(got, s) {
			return s, i, true
		}
	}
	return "", 0, false
}

// firstPresent returns the first substring in subs that IS present in got, its
// 0-based index, and true; it returns ok=false when every element is absent. It
// backs the `not_contains` matcher's "all must be absent" semantics.
func firstPresent(got string, subs spec.StringList) (sub string, idx int, ok bool) {
	for i, s := range subs {
		if strings.Contains(got, s) {
			return s, i, true
		}
	}
	return "", 0, false
}

// quoteList renders a StringList as a space-separated list of quoted elements,
// e.g. `"a", "b"`, for the multi-element failure description.
func quoteList(subs spec.StringList) string {
	parts := make([]string, len(subs))
	for i, s := range subs {
		parts[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(parts, ", ")
}

// fileContainsDesc renders the file assertion label, mirroring containsDesc but
// with the file path in the subject.
func fileContainsDesc(pathLabel string, subs spec.StringList, want bool) string {
	if len(subs) == 1 {
		if want {
			return fmt.Sprintf("assert file %q contains %q", pathLabel, subs[0])
		}
		return fmt.Sprintf("assert file %q does not contain %q", pathLabel, subs[0])
	}
	if want {
		return fmt.Sprintf("assert file %q contains all of %s", pathLabel, quoteList(subs))
	}
	return fmt.Sprintf("assert file %q contains none of %s", pathLabel, quoteList(subs))
}

// elementLabel returns "" for a single-element matcher (so the failure text is
// byte-identical to the pre-list format) and " (element N of M)" for a list, so
// a failure over an array says which element failed.
func elementLabel(idx, n int) string {
	if n <= 1 {
		return ""
	}
	return fmt.Sprintf(" (element %d of %d)", idx+1, n)
}
