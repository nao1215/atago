package spec

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestVarRefs(t *testing.T) {
	t.Parallel()
	got := VarRefs("${a}/x/${b_2}-${a}")
	sort.Strings(got)
	if want := []string{"a", "a", "b_2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("VarRefs = %v, want %v", got, want)
	}
	if VarRefs("no vars here") != nil {
		t.Error("VarRefs on plain text should be nil")
	}
	// Issue #37: an escaped $${name} is literal text, not a live reference, so
	// it must not be collected (the linter would otherwise flag it as undefined).
	if got := VarRefs("$${literal}"); got != nil {
		t.Errorf("VarRefs on escaped ref = %v, want nil", got)
	}
	if got := VarRefs("$${skip} but ${real}"); !reflect.DeepEqual(got, []string{"real"}) {
		t.Errorf("VarRefs mixed escape = %v, want [real]", got)
	}
	// Issue #24: namespaced built-ins use dotted names (${<mock>.url},
	// ${<mock>.port}). VarRefs must stay in lockstep with store.varRef and
	// collect them, or manifest/explain silently drop every dotted reference.
	if got := VarRefs("base_url: ${api.url} port ${api.port}"); !reflect.DeepEqual(got, []string{"api.url", "api.port"}) {
		t.Errorf("VarRefs dotted = %v, want [api.url api.port]", got)
	}
	if got := VarRefs("${env:HOME}"); !reflect.DeepEqual(got, []string{"env:HOME"}) {
		t.Errorf("VarRefs env = %v, want [env:HOME]", got)
	}
}

func sp(s string) *string { return &s }

// TestWalkAssertStrings_CollectAndExpand verifies the walker both records (with
// an identity visit) and substitutes (with a mutating visit) across every
// interpolatable field, and that it returns a copy without mutating the input.
func TestWalkAssertStrings_CollectAndExpand(t *testing.T) {
	t.Parallel()
	emptyList := StringList{}
	a := &Assert{
		Stdout:  &StreamAssert{Contains: StringList{"${a}"}},
		Rows:    &StreamAssert{JSON: &JSONAssert{Path: "$.${b}", Equals: "${c}"}},
		Message: &StreamAssert{Equals: sp("${d}")},
		Value:   &StreamAssert{YAML: &JSONAssert{Path: "$.x", Matches: sp("${e}")}},
		File:    &FileAssert{Path: "${f}", Contains: StringList{"${g}"}},
		Header:  &HeaderMatch{Name: "X", Equals: sp("${h}"), Matches: sp("${r}")},
		Image:   &ImageAssert{Path: "${i}", SimilarTo: "${j}"},
		Screen:  &StreamAssert{Contains: StringList{"${k}"}},
		Dir:     &DirAssert{Path: "${l}", Contains: []string{"${m}"}, NotContains: []string{"${n}"}, Glob: "${o}", Ignore: []string{"${p}"}},
		PDF:     &PDFAssert{Path: "${q}", Metadata: map[string]string{"title": "${s}"}, Text: &StreamAssert{Contains: StringList{"${t}"}}},
		Mock:    &MockAssert{Name: "api", Path: "${u}", Header: &HeaderMatch{Name: "Y", Contains: sp("${v}")}, Body: &StreamAssert{Contains: StringList{"${w}"}}},
		Changes: &ChangesAssert{Created: &StringList{"${x}"}, Modified: &emptyList},
	}

	// Collect: identity visit records every reference.
	seen := map[string]bool{}
	WalkAssertStrings(a, func(s string) string {
		for _, n := range VarRefs(s) {
			seen[n] = true
		}
		return s
	})
	for _, want := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x"} {
		if !seen[want] {
			t.Errorf("collect missed ${%s}; got %v", want, seen)
		}
	}

	// Expand: mutating visit substitutes, and the input is not mutated.
	out := WalkAssertStrings(a, func(s string) string { return strings.ReplaceAll(s, "${a}", "A") })
	if out.Stdout.Contains[0] != "A" {
		t.Errorf("expand stdout.contains = %q, want A", out.Stdout.Contains[0])
	}
	if a.Stdout.Contains[0] != "${a}" {
		t.Error("WalkAssertStrings mutated its input")
	}

	// The changes lists keep the nil-vs-empty distinction the exhaustive-set
	// semantics depend on: a visited empty list stays non-nil-and-empty, and an
	// omitted (nil) list stays nil.
	if out.Changes.Modified == nil || len(*out.Changes.Modified) != 0 {
		t.Errorf("changes.modified = %v, want a non-nil empty list", out.Changes.Modified)
	}
	if out.Changes.Deleted != nil {
		t.Errorf("changes.deleted = %v, want nil (omitted stays unconstrained)", out.Changes.Deleted)
	}
}

func TestWalkJSONValueStrings(t *testing.T) {
	t.Parallel()
	in := map[string]any{"s": "${x}", "n": 1, "arr": []any{"${y}", 2}}
	out, ok := WalkJSONValueStrings(in, func(s string) string { return strings.ToUpper(s) }).(map[string]any)
	if !ok {
		t.Fatalf("result is not a map")
	}
	if out["s"] != "${X}" {
		t.Errorf("string leaf = %v", out["s"])
	}
	if out["n"] != 1 {
		t.Errorf("non-string mutated: %v", out["n"])
	}
	arr, ok := out["arr"].([]any)
	if !ok || arr[0] != "${Y}" || arr[1] != 2 {
		t.Errorf("array = %v", out["arr"])
	}
	// Input not mutated.
	if in["s"] != "${x}" {
		t.Error("WalkJSONValueStrings mutated its input")
	}
}
