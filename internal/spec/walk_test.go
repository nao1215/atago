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
}

func sp(s string) *string { return &s }

// TestWalkAssertStrings_CollectAndExpand verifies the walker both records (with
// an identity visit) and substitutes (with a mutating visit) across every
// interpolatable field, and that it returns a copy without mutating the input.
func TestWalkAssertStrings_CollectAndExpand(t *testing.T) {
	t.Parallel()
	a := &Assert{
		Stdout:  &StreamAssert{Contains: StringList{"${a}"}},
		Rows:    &StreamAssert{JSON: &JSONAssert{Path: "$.${b}", Equals: "${c}"}},
		Message: &StreamAssert{Equals: sp("${d}")},
		Value:   &StreamAssert{YAML: &JSONAssert{Path: "$.x", Matches: sp("${e}")}},
		File:    &FileAssert{Path: "${f}", Contains: StringList{"${g}"}},
		Header:  &HeaderMatch{Name: "X", Equals: sp("${h}")},
		Image:   &ImageAssert{Path: "${i}", SimilarTo: "${j}"},
	}

	// Collect: identity visit records every reference.
	seen := map[string]bool{}
	WalkAssertStrings(a, func(s string) string {
		for _, n := range VarRefs(s) {
			seen[n] = true
		}
		return s
	})
	for _, want := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
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
