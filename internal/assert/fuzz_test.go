package assert

import (
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// FuzzCheckJSON feeds arbitrary document bytes and an arbitrary JSONPath to the
// JSON matcher (issue #46). The matcher must never panic and must always return
// a diagnosis (invalid JSON, invalid path, no match, or a comparison) rather
// than crashing on adversarial input.
func FuzzCheckJSON(f *testing.F) {
	seeds := []struct {
		data, path string
	}{
		{`{"a":1}`, "$.a"},
		{`[1,2,3]`, "$[0]"},
		{"", "$"},
		{"not json", "$.x"},
		{`{"a":{"b":[{"c":1}]}}`, "$.a.b[0].c"},
		{`{"a":1}`, "$["},
	}
	for _, s := range seeds {
		f.Add([]byte(s.data), s.path)
	}
	f.Fuzz(func(t *testing.T, data []byte, path string) {
		j := &spec.JSONAssert{Path: path, Equals: "x"}
		if cr := checkJSON("fuzz", "fuzz", data, j); cr == nil {
			t.Fatal("checkJSON returned nil CheckResult")
		}
	})
}

// FuzzValuesEqual asserts that structural comparison of decoded JSON values is
// reflexive and total: a parsed value always equals itself, and comparing any
// two parsed values never panics (issue #46, #40). It parses through the same
// panic-recovering parseJSON the matcher uses, so an ojg parser panic on
// malformed input is a skipped case here, not a fuzz failure.
func FuzzValuesEqual(f *testing.F) {
	for _, s := range []string{`1`, `"s"`, `[1,2]`, `{"a":1,"b":[2,3]}`, `null`, `true`, `1.5`} {
		f.Add([]byte(s), []byte(`{"a":1}`))
	}
	f.Fuzz(func(t *testing.T, a, b []byte) {
		va, errA := parseJSON(a)
		vb, errB := parseJSON(b)
		if errA != nil || errB != nil {
			return
		}
		if !valuesEqual(va, va) {
			t.Fatalf("valuesEqual is not reflexive for %q", a)
		}
		// Comparing two arbitrary values must be total (never panics); the
		// boolean result itself is not constrained here.
		_ = valuesEqual(va, vb)
	})
}
