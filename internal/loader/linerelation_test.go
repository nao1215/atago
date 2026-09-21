package loader

import (
	"strings"
	"testing"
)

// relationSpec wraps one stdout assertion in the smallest spec that carries it.
func relationSpec(assert string) string {
	return "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert: {stdout: " + assert + "}\n"
}

// TestLoadBytes_LineRelationsAreWholeStreamMatchers: permutation_of and
// subset_of pin every line, so combining them with another matcher would state
// two contracts where the stronger one already decides (#658). They follow
// equals, which the loader has always refused to combine.
func TestLoadBytes_LineRelationsAreWholeStreamMatchers(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		assert string
		want   string
	}{
		"permutation_of alone":          {"{permutation_of: \"a\\nb\"}", ""},
		"subset_of alone":               {"{subset_of: \"a\\nb\"}", ""},
		"permutation_of with contains":  {"{permutation_of: \"a\", contains: a}", "cannot be combined"},
		"subset_of with equals":         {"{subset_of: \"a\", equals: a}", "cannot be combined"},
		"both relations together":       {"{subset_of: \"a\", permutation_of: \"a\"}", "cannot be combined"},
		"permutation_of with a count":   {"{permutation_of: \"a\", count: 1}", "cannot be combined"},
		"empty permutation_of refused":  {"{permutation_of: \"\"}", "compares against no lines"},
		"empty subset_of refused":       {"{subset_of: \"\"}", "compares against no lines"},
		"relation with a line selector": {"{line: 1, permutation_of: \"a\"}", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("r.atago.yaml", []byte(relationSpec(c.assert)))
			switch {
			case c.want == "" && err != nil:
				t.Errorf("LoadBytes() error = %v, want the assertion accepted", err)
			case c.want != "" && err == nil:
				t.Errorf("LoadBytes() error = nil, want it refused with %q", c.want)
			case c.want != "" && err != nil && !strings.Contains(err.Error(), c.want):
				t.Errorf("error = %q, want it to mention %q", err, c.want)
			}
		})
	}
}

// TestLoadBytes_NoMatcherDiagnosticNamesTheRelations keeps the ways out of the
// "must set at least one matcher" error complete: an author who reached for a
// relation and put it in the wrong place has to see it listed.
func TestLoadBytes_NoMatcherDiagnosticNamesTheRelations(t *testing.T) {
	t.Parallel()
	_, err := LoadBytes("r.atago.yaml", []byte(relationSpec("{}")))
	if err == nil {
		t.Fatal("LoadBytes() error = nil, want an empty stream assertion refused")
	}
	for _, want := range []string{"permutation_of", "subset_of"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to name %q as a matcher", err, want)
		}
	}
}
