package scrub

import (
	"bytes"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// TestNew_CompilesAndApplies proves an ordered rule set rewrites every match to
// its literal placeholder, applying rules top-to-bottom.
func TestNew_CompilesAndApplies(t *testing.T) {
	t.Parallel()
	s, err := New([]spec.ScrubRule{
		{Pattern: `req-\d+`, Placeholder: "<REQ>"},
		{Pattern: `id=\d+`, Placeholder: "id=<ID>"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	in := []byte("start req-42 mid id=1007 end req-9")
	got := string(s.Apply(in))
	want := "start <REQ> mid id=<ID> end <REQ>"
	if got != want {
		t.Errorf("Apply = %q, want %q", got, want)
	}
}

// TestNew_InvalidPatternErrors proves a malformed regexp is reported with its
// index and the offending pattern, so a spec author can fix it.
func TestNew_InvalidPatternErrors(t *testing.T) {
	t.Parallel()
	_, err := New([]spec.ScrubRule{
		{Pattern: `ok`, Placeholder: "x"},
		{Pattern: `(`, Placeholder: "y"},
	})
	if err == nil {
		t.Fatal("New accepted an invalid regexp, want error")
	}
	for _, want := range []string{"scrub[1]", "("} {
		if !bytes.Contains([]byte(err.Error()), []byte(want)) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
}

// TestApply_LiteralPlaceholder proves a `$`-bearing placeholder is inserted
// literally, not treated as a regexp expansion template ($1 must survive).
func TestApply_LiteralPlaceholder(t *testing.T) {
	t.Parallel()
	s, err := New([]spec.ScrubRule{{Pattern: `(foo)`, Placeholder: "$1<X>"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := string(s.Apply([]byte("foo")))
	if got != "$1<X>" {
		t.Errorf("Apply = %q, want literal %q (no $1 expansion)", got, "$1<X>")
	}
}

// TestApply_EmptyPlaceholderDeletes proves an empty placeholder deletes matches.
func TestApply_EmptyPlaceholderDeletes(t *testing.T) {
	t.Parallel()
	s, err := New([]spec.ScrubRule{{Pattern: `\s+trailing`, Placeholder: ""}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := string(s.Apply([]byte("keep   trailing"))); got != "keep" {
		t.Errorf("Apply = %q, want %q", got, "keep")
	}
}

// TestNilAndEmpty_Passthrough proves a nil Scrubber and an empty rule set are
// safe no-ops that return the input unchanged, so callers can hold a possibly-nil
// scrubber and always call Apply.
func TestNilAndEmpty_Passthrough(t *testing.T) {
	t.Parallel()
	var nilS *Scrubber
	in := []byte("unchanged 123")
	if got := nilS.Apply(in); !bytes.Equal(got, in) {
		t.Errorf("nil Apply = %q, want %q", got, in)
	}
	empty, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil): %v", err)
	}
	if empty != nil {
		t.Errorf("New(nil) = %v, want nil scrubber for an empty rule set", empty)
	}
	if got := empty.Apply(in); !bytes.Equal(got, in) {
		t.Errorf("empty Apply = %q, want %q", got, in)
	}
}
