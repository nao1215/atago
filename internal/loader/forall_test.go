package loader

import (
	"strings"
	"testing"
)

// forallSpec wraps a forall block in the smallest spec that carries it.
func forallSpec(block string) string {
	return "version: \"1\"\nsuite:\n  name: forall\nscenarios:\n  - name: a\n    forall:\n" +
		block + "    steps:\n      - run: {command: echo}\n"
}

// TestLoadBytes_ForallExpands is the shape of the whole feature: a forall block
// is gone by the time loading finishes, and what is left is ordinary scenarios
// carrying concrete values — which is why nothing downstream (the engine, the
// reports, explain, doc, list, the manifest) needed to learn about it.
func TestLoadBytes_ForallExpands(t *testing.T) {
	t.Parallel()
	s, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {s: {type: alpha, min: 2, max: 4}}\n      runs: 5\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if len(s.Scenarios) != 5 {
		t.Fatalf("scenarios = %d, want 5 generated instances", len(s.Scenarios))
	}
	for i, sc := range s.Scenarios {
		if sc.Forall != nil {
			t.Errorf("instance %d still carries a forall block", i)
		}
		if sc.Matrix != nil {
			t.Errorf("instance %d still carries the generated rows", i)
		}
		v, bound := sc.Vars["s"]
		if !bound {
			t.Fatalf("instance %d bound no value: %+v", i, sc.Vars)
		}
		if len(v) < 2 || len(v) > 4 {
			t.Errorf("instance %d value %q is outside the declared length range", i, v)
		}
		if !strings.Contains(sc.Name, v) {
			t.Errorf("instance %d name %q does not carry its value %q, so a failure cannot name its input", i, sc.Name, v)
		}
	}
}

// TestLoadBytes_ForallIsReproducible: loading the same bytes twice has to
// produce the same inputs, or a failing instance could not be re-run — and a
// spec whose seed differs has to produce different ones, which is the only way
// an author asks for a new draw.
func TestLoadBytes_ForallIsReproducible(t *testing.T) {
	t.Parallel()
	src := forallSpec("      vars: {s: alpha}\n")
	first, err := LoadBytes("f.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	second, err := LoadBytes("f.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	for i := range first.Scenarios {
		if first.Scenarios[i].Vars["s"] != second.Scenarios[i].Vars["s"] {
			t.Fatalf("instance %d differs between loads: %q vs %q", i, first.Scenarios[i].Vars["s"], second.Scenarios[i].Vars["s"])
		}
	}
	seeded, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {s: alpha}\n      seed: 9\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	identical := true
	for i := range first.Scenarios {
		if first.Scenarios[i].Vars["s"] != seeded.Scenarios[i].Vars["s"] {
			identical = false
			break
		}
	}
	if identical {
		t.Error("changing seed: produced the same inputs")
	}
}

// TestLoadBytes_ForallDefaults pins the two defaults an author gets for free:
// ten instances, and the kind's own range.
func TestLoadBytes_ForallDefaults(t *testing.T) {
	t.Parallel()
	s, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {s: digits}\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if len(s.Scenarios) != defaultForallRuns {
		t.Errorf("scenarios = %d, want the default %d", len(s.Scenarios), defaultForallRuns)
	}
	for _, sc := range s.Scenarios {
		if l := len(sc.Vars["s"]); l < 1 || l > 8 {
			t.Errorf("value %q is outside the default 1-8 length range", sc.Vars["s"])
		}
	}
}

// TestLoadBytes_ForallOneOfIsVerbatim: a choice generator's values are
// substituted into the scenario as text, so `007` means those three characters.
// YAML would have typed it as the integer 7 and the CLI under test would have
// received a different argument than the spec shows.
func TestLoadBytes_ForallOneOfIsVerbatim(t *testing.T) {
	t.Parallel()
	s, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {v: {one_of: [007, 1.20, \"x\"]}}\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	got := map[string]bool{}
	for _, sc := range s.Scenarios {
		got[sc.Vars["v"]] = true
	}
	for _, want := range []string{"007", "1.20", "x"} {
		if !got[want] {
			t.Errorf("no instance bound %q; got %v", want, got)
		}
	}
	if len(s.Scenarios) != 3 {
		t.Errorf("scenarios = %d, want one per choice", len(s.Scenarios))
	}
}

// TestLoadBytes_ForallNameTemplate: a name that references the generated
// variables is substituted per instance, exactly as a matrix name is.
func TestLoadBytes_ForallNameTemplate(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: forall\nscenarios:\n  - name: \"parses ${v}\"\n    forall:\n" +
		"      vars: {v: {one_of: [json, yaml]}}\n    steps:\n      - run: {command: echo}\n"
	s, err := LoadBytes("f.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if len(s.Scenarios) != 2 || s.Scenarios[0].Name != "parses json" || s.Scenarios[1].Name != "parses yaml" {
		t.Errorf("names = %q, %q; want the template filled per instance", s.Scenarios[0].Name, s.Scenarios[1].Name)
	}
}

// TestLoadBytes_ForallRefused walks the authoring mistakes a generator block can
// carry. Each message has to name what is wrong, because a generated input is
// the one part of a spec the author cannot read off the file.
func TestLoadBytes_ForallRefused(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		block string
		want  string
	}{
		"unknown kind":          {"      vars: {s: strings}\n", "is not a generator"},
		"no kind at all":        {"      vars: {s: {min: 1}}\n", "must name a generator type"},
		"range on bool":         {"      vars: {b: {type: bool, max: 3}}\n", "no range to narrow"},
		"one_of with a type":    {"      vars: {s: {type: alpha, one_of: [a]}}\n", "one_of together with type"},
		"duplicate choice":      {"      vars: {s: {one_of: [a, a]}}\n", "twice"},
		"empty range":           {"      vars: {s: {type: alpha, min: 5, max: 2}}\n", "greater than max"},
		"negative length":       {"      vars: {s: {type: alpha, min: -1}}\n", "string LENGTH"},
		"length over the limit": {"      vars: {s: {type: ascii, max: 9000}}\n", "at most 4096"},
		"too many runs":         {"      vars: {s: alpha}\n      runs: 5000\n", "at most 1000"},
		"negative runs":         {"      vars: {s: alpha}\n      runs: -1\n", "must not be negative"},
		"reserved name":         {"      vars: {workdir: alpha}\n", "shadows a built-in variable"},
		"unknown key":           {"      vars: {s: {type: alpha, mins: 1}}\n", "unknown key"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("f.atago.yaml", []byte(forallSpec(c.block)))
			if err == nil {
				t.Fatalf("LoadBytes() error = nil, want the block refused")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error = %q, want it to mention %q", err, c.want)
			}
		})
	}
}

// TestLoadBytes_ForallAndMatrix: the two are alternatives. Layering them would
// have to mean a product, which nobody asked for and which would make the
// instance count of a spec impossible to read off it.
func TestLoadBytes_ForallAndMatrix(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: forall\nscenarios:\n  - name: a\n    matrix: [{x: 1}]\n    forall:\n" +
		"      vars: {s: alpha}\n    steps:\n      - run: {command: echo}\n"
	_, err := LoadBytes("f.atago.yaml", []byte(src))
	if err == nil || !strings.Contains(err.Error(), "states its inputs or generates them") {
		t.Errorf("error = %v, want both-blocks refused", err)
	}
}

// TestLoadBytes_ForallNameReferences: a name that names SOME of the generated
// variables collapses the instances that differ in the others onto one name.
// The author cannot see that coming — the values are not in the file — so the
// refusal has to name the variables the template left out.
func TestLoadBytes_ForallNameReferences(t *testing.T) {
	t.Parallel()
	t.Run("partial reference", func(t *testing.T) {
		t.Parallel()
		src := "version: \"1\"\nsuite:\n  name: forall\nscenarios:\n  - name: \"checks ${a}\"\n    forall:\n" +
			"      vars: {a: alpha, b: digits}\n    steps:\n      - run: {command: echo}\n"
		_, err := LoadBytes("f.atago.yaml", []byte(src))
		if err == nil || !strings.Contains(err.Error(), "${b}") {
			t.Errorf("error = %v, want the omitted variable named", err)
		}
	})
	t.Run("undeclared reference", func(t *testing.T) {
		t.Parallel()
		src := "version: \"1\"\nsuite:\n  name: forall\nscenarios:\n  - name: \"checks ${typo}\"\n    forall:\n" +
			"      vars: {a: alpha}\n    steps:\n      - run: {command: echo}\n"
		_, err := LoadBytes("f.atago.yaml", []byte(src))
		if err == nil || !strings.Contains(err.Error(), "which forall does not generate") {
			t.Errorf("error = %v, want the unknown reference reported", err)
		}
	})
}

// TestLoadBytes_ForallScalarShorthand: `s: ascii` is the whole opt-in, and it
// has to mean the same thing as spelling the mapping out.
func TestLoadBytes_ForallScalarShorthand(t *testing.T) {
	t.Parallel()
	short, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {s: ascii}\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	long, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {s: {type: ascii}}\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	for i := range short.Scenarios {
		if short.Scenarios[i].Vars["s"] != long.Scenarios[i].Vars["s"] {
			t.Errorf("instance %d: shorthand %q, mapping %q", i, short.Scenarios[i].Vars["s"], long.Scenarios[i].Vars["s"])
		}
	}
}

// TestLoadBytes_ForallExamples pins the must-try values (#661): they are the
// first instances, in the order written, and they are used as written — an
// example outside the generator's own range is the point, not a mistake.
func TestLoadBytes_ForallExamples(t *testing.T) {
	t.Parallel()
	s, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars:\n        s: {type: alpha, min: 3, max: 6, examples: [\"\", \"--\", \"a b\"]}\n      runs: 5\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if len(s.Scenarios) != 5 {
		t.Fatalf("scenarios = %d, want 5", len(s.Scenarios))
	}
	for i, want := range []string{"", "--", "a b"} {
		if got := s.Scenarios[i].Vars["s"]; got != want {
			t.Errorf("instance %d bound %q, want the example %q", i, got, want)
		}
	}
	for _, sc := range s.Scenarios[3:] {
		if l := len(sc.Vars["s"]); l < 3 || l > 6 {
			t.Errorf("generated value %q is outside the declared range", sc.Vars["s"])
		}
	}
}

// TestLoadBytes_ForallExamplesAreVerbatim: an example is substituted into the
// scenario as text, so `007` means those three characters, exactly as a one_of
// value does.
func TestLoadBytes_ForallExamplesAreVerbatim(t *testing.T) {
	t.Parallel()
	s, err := LoadBytes("f.atago.yaml", []byte(forallSpec("      vars: {v: {type: digits, examples: [007, 1.20]}}\n")))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	for i, want := range []string{"007", "1.20"} {
		if got := s.Scenarios[i].Vars["v"]; got != want {
			t.Errorf("instance %d bound %q, want %q", i, got, want)
		}
	}
}

// TestLoadBytes_ForallExamplesRefused walks the mistakes an examples list can
// carry. The count rule is the one worth a message: more examples than runs
// would either drop the last ones or run more scenarios than the block says.
func TestLoadBytes_ForallExamplesRefused(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		block string
		want  string
	}{
		"more examples than runs": {"      vars: {s: {type: alpha, examples: [a, b, c]}}\n      runs: 2\n", "raise runs to at least 3"},
		"duplicate example":       {"      vars: {s: {type: alpha, examples: [a, a]}}\n", "twice"},
		"examples with one_of":    {"      vars: {s: {one_of: [a, b], examples: [c]}}\n", "examples together with one_of"},
		"examples not a list":     {"      vars: {s: {type: alpha, examples: nope}}\n", "must be a list"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("f.atago.yaml", []byte(forallSpec(c.block)))
			if err == nil {
				t.Fatalf("LoadBytes() error = nil, want the block refused")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error = %q, want it to mention %q", err, c.want)
			}
		})
	}
}

// TestLoadBytes_ForallExamplesWithinDefaultRuns: the count rule reads the
// default when the block states no runs, so the message quotes the number the
// loader would actually have used.
func TestLoadBytes_ForallExamplesWithinDefaultRuns(t *testing.T) {
	t.Parallel()
	many := make([]string, defaultForallRuns+1)
	for i := range many {
		many[i] = string(rune('a' + i))
	}
	block := "      vars: {s: {type: alpha, examples: [" + strings.Join(many, ", ") + "]}}\n"
	_, err := LoadBytes("f.atago.yaml", []byte(forallSpec(block)))
	if err == nil || !strings.Contains(err.Error(), "forall.runs is 10") {
		t.Errorf("error = %v, want the default runs quoted", err)
	}
}
