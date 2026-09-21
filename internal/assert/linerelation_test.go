package assert

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// TestCheck_PermutationOf covers the relation a CLI needs when the content of
// the output is the contract and its order is not: the same lines, in any
// order. It is an equality over multisets, so a line printed twice has to be
// there twice.
func TestCheck_PermutationOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		stdout string
		want   string
		wantOK bool
	}{
		{"same lines reordered", "b\na\nc\n", "a\nb\nc", true},
		{"identical output", "a\nb\n", "a\nb", true},
		{"trailing newline is not a line", "a\nb\n", "a\nb\n", true},
		{"a missing line fails", "a\nb\n", "a\nb\nc", false},
		{"an extra line fails", "a\nb\nc\n", "a\nb", false},
		{"a duplicate is not the same as one", "a\na\n", "a", false},
		{"duplicates match when both have them", "a\nb\na\n", "a\na\nb", true},
		{"a blank line inside counts", "a\n\nb\n", "a\nb", false},
		{"leading space is part of the line", "  a\nb\n", "a\nb", false},
		{"CRLF output folds like every text matcher", "b\r\na\r\n", "a\nb", true},
		{"a different line fails", "a\nx\n", "a\nb", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: []byte(tt.stdout)}
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{PermutationOf: strp(tt.want)}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_SubsetOf covers the relation a filtering flag has to satisfy: every
// line of the output was already in the reference. It is a multiset subset, so
// output that duplicated a line the reference has once is not a subset —
// printing something twice is a bug a set comparison would hide.
func TestCheck_SubsetOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		stdout string
		want   string
		wantOK bool
	}{
		{"a filtered listing", "b\n", "a\nb\nc", true},
		{"the whole listing is a subset of itself", "a\nb\nc\n", "a\nb\nc", true},
		{"empty output is a subset of anything", "", "a\nb", true},
		{"a line the reference never had fails", "b\nz\n", "a\nb\nc", false},
		{"a duplicated line needs two in the reference", "b\nb\n", "a\nb\nc", false},
		{"a duplicated line passes when the reference has two", "b\nb\n", "b\nb\nc", true},
		{"reordering does not matter", "c\na\n", "a\nb\nc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res := &runner.Result{Stdout: []byte(tt.stdout)}
			got := Check(&spec.Assert{Stdout: &spec.StreamAssert{SubsetOf: strp(tt.want)}}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_LineRelation_FailureNamesTheDifference: a relation that only says
// "no" leaves the reader diffing two blocks of text by eye, which is the whole
// reason the order-insensitive comparison was needed in the first place. The
// hint has to name the lines that are missing and the ones that are extra.
func TestCheck_LineRelation_FailureNamesTheDifference(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("alpha\ndelta\n")}
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{PermutationOf: strp("alpha\nbeta")}}, res, Env{})
	if got.OK {
		t.Fatal("OK = true, want a failure")
	}
	if !strings.Contains(got.Hint, "beta") {
		t.Errorf("Hint = %q, want the missing line named", got.Hint)
	}
	if !strings.Contains(got.Hint, "delta") {
		t.Errorf("Hint = %q, want the unexpected line named", got.Hint)
	}
	subset := Check(&spec.Assert{Stdout: &spec.StreamAssert{SubsetOf: strp("alpha\nbeta")}}, res, Env{})
	if subset.OK {
		t.Fatal("OK = true, want a failure")
	}
	if !strings.Contains(subset.Hint, "delta") {
		t.Errorf("Hint = %q, want the line that was not in the reference named", subset.Hint)
	}
	if strings.Contains(subset.Hint, "beta") {
		t.Errorf("Hint = %q, must not report a reference line as missing: a subset may leave lines out", subset.Hint)
	}
}

// TestCheck_LineRelation_NoCommand keeps the relations in line with every other
// stream matcher: with no command run, the failure says so rather than
// comparing against nothing.
func TestCheck_LineRelation_NoCommand(t *testing.T) {
	t.Parallel()
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{PermutationOf: strp("a")}}, nil, Env{})
	if got.OK {
		t.Fatal("OK = true, want a failure with no command")
	}
}
