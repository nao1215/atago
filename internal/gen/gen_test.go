package gen

import (
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// alphaVar is the shorthand every test below starts from.
func alphaVar(minimum, maximum int) Var {
	return Var{Name: "s", Kind: Alpha, Min: minimum, Max: maximum}
}

// TestRows_Deterministic is the property the whole package exists for: the same
// seed produces the same values, on every run and every machine. A generated
// input that changed under the author's feet would change what the spec asserts,
// and a failure nobody can reproduce is worse than no test.
func TestRows_Deterministic(t *testing.T) {
	t.Parallel()
	vars := []Var{alphaVar(1, 8), {Name: "n", Kind: Int, Min: -5, Max: 5}}
	first := Rows(42, 12, vars)
	second := Rows(42, 12, vars)
	if len(first) != len(second) {
		t.Fatalf("row counts differ: %d vs %d", len(first), len(second))
	}
	for i := range first {
		for k, v := range first[i] {
			if second[i][k] != v {
				t.Errorf("row %d var %q = %q on the second call, want %q", i, k, second[i][k], v)
			}
		}
	}
	if same(first, Rows(43, 12, vars)) {
		t.Error("a different seed produced the same rows; the seed has to pick the stream")
	}
}

// same reports whether two row lists are identical.
func same(a, b []map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		for k, v := range a[i] {
			if b[i][k] != v {
				return false
			}
		}
	}
	return true
}

// TestRows_BoundaryValuesComeFirst pins the edges-first order ported from
// metamon: the ends of the space are what a CLI breaks on, so a suite that runs
// ten instances must spend the first of them there rather than reaching an
// empty argument by luck on the seventh.
func TestRows_BoundaryValuesComeFirst(t *testing.T) {
	t.Parallel()
	t.Run("int range ends", func(t *testing.T) {
		t.Parallel()
		rows := Rows(1, 5, []Var{{Name: "n", Kind: Int, Min: -3, Max: 9}})
		if rows[0]["n"] != "-3" || rows[1]["n"] != "9" {
			t.Errorf("first rows = %q, %q; want the range ends -3 and 9", rows[0]["n"], rows[1]["n"])
		}
	})
	t.Run("string length ends", func(t *testing.T) {
		t.Parallel()
		rows := Rows(1, 5, []Var{alphaVar(0, 6)})
		if len(rows[0]["s"]) != 0 || len(rows[1]["s"]) != 6 {
			t.Errorf("first lengths = %d, %d; want the empty string and the longest one", len(rows[0]["s"]), len(rows[1]["s"]))
		}
	})
	t.Run("both booleans", func(t *testing.T) {
		t.Parallel()
		rows := Rows(1, 5, []Var{{Name: "b", Kind: Bool}})
		if len(rows) != 2 || rows[0]["b"] != "false" || rows[1]["b"] != "true" {
			t.Errorf("rows = %v; want exactly false then true", rows)
		}
	})
	t.Run("every choice", func(t *testing.T) {
		t.Parallel()
		rows := Rows(1, 9, []Var{{Name: "f", Kind: OneOf, Values: []string{"json", "yaml", "text"}}})
		if len(rows) != 3 {
			t.Fatalf("rows = %v; want one per choice", rows)
		}
		for i, want := range []string{"json", "yaml", "text"} {
			if rows[i]["f"] != want {
				t.Errorf("row %d = %q, want %q", i, rows[i]["f"], want)
			}
		}
	})
}

// TestRows_AreDistinct: runs is an upper bound, not a promise. Twenty instances
// of a two-value generator would be eighteen copies, each of them a process
// start and a line of report.
func TestRows_AreDistinct(t *testing.T) {
	t.Parallel()
	rows := Rows(7, 20, []Var{{Name: "f", Kind: OneOf, Values: []string{"a", "b"}}, {Name: "b", Kind: Bool}})
	if len(rows) != 4 {
		t.Errorf("rows = %d, want the 4 distinct combinations of two choices and two booleans: %v", len(rows), rows)
	}
	seen := map[string]bool{}
	for _, row := range rows {
		key := row["f"] + "/" + row["b"]
		if seen[key] {
			t.Errorf("duplicate row %q", key)
		}
		seen[key] = true
	}
}

// TestRows_StayInBounds checks every value against the space it was asked for,
// over the whole run rather than on the first row.
func TestRows_StayInBounds(t *testing.T) {
	t.Parallel()
	rows := Rows(3, 40, []Var{
		{Name: "n", Kind: Int, Min: 10, Max: 12},
		{Name: "d", Kind: Digits, Min: 2, Max: 4},
		{Name: "a", Kind: Alphanumeric, Min: 1, Max: 5},
	})
	for _, row := range rows {
		n, err := strconv.Atoi(row["n"])
		if err != nil || n < 10 || n > 12 {
			t.Errorf("n = %q, want an integer in [10, 12]", row["n"])
		}
		if l := len(row["d"]); l < 2 || l > 4 {
			t.Errorf("d = %q, want 2-4 digits", row["d"])
		}
		if strings.Trim(row["d"], digitChars) != "" {
			t.Errorf("d = %q, want digits only", row["d"])
		}
		if l := len(row["a"]); l < 1 || l > 5 {
			t.Errorf("a = %q, want 1-5 characters", row["a"])
		}
		if strings.Trim(row["a"], alphaChars+digitChars) != "" {
			t.Errorf("a = %q, want letters and digits only", row["a"])
		}
	}
}

// TestASCII_CannotRequoteACommand: a generated value is substituted into a
// command line before that line is split into argv, and into every other field
// as text. A value carrying a quote, a backslash or a `${` would change how the
// command is READ rather than what it says, and the failure would be a
// tokenizer error rather than the program's behavior.
func TestASCII_CannotRequoteACommand(t *testing.T) {
	t.Parallel()
	rows := Rows(11, 60, []Var{{Name: "s", Kind: ASCII, Min: 1, Max: 40}})
	for _, row := range rows {
		if strings.ContainsAny(row["s"], "'\"\\`$") {
			t.Errorf("ascii value %q carries a quoting character", row["s"])
		}
		for _, r := range row["s"] {
			if r < 0x20 || r > 0x7E {
				t.Errorf("ascii value %q carries a non-printable character %U", row["s"], r)
			}
		}
	}
}

// TestUnicode_IsPrintableNonASCII: the unicode kind exists for the "does this
// CLI handle text it did not write" question, so its values have to be valid
// UTF-8 a report and a terminal can show — not unassigned code points and not
// ASCII the other kinds already cover.
func TestUnicode_IsPrintableNonASCII(t *testing.T) {
	t.Parallel()
	rows := Rows(5, 20, []Var{{Name: "u", Kind: Unicode, Min: 1, Max: 6}})
	for _, row := range rows {
		v := row["u"]
		if !utf8.ValidString(v) {
			t.Errorf("unicode value %q is not valid UTF-8", v)
		}
		for _, r := range v {
			if r < 0x80 {
				t.Errorf("unicode value %q carries the ASCII rune %q", v, r)
			}
			if !unicode.IsPrint(r) {
				t.Errorf("unicode value %q carries a non-printable rune %U", v, r)
			}
		}
		if n := utf8.RuneCountInString(v); n < 1 || n > 6 {
			t.Errorf("unicode value %q has %d runes, want 1-6 (the bound is a rune count)", v, n)
		}
	}
}

// TestRows_GoldenStream pins actual generated values. The PRNG, the alphabets
// and the order edges are consumed in are all part of what a spec means: change
// any of them and every forall spec in the world starts testing different
// inputs, and every committed document that names an instance goes stale. This
// test is what makes that a deliberate edit.
func TestRows_GoldenStream(t *testing.T) {
	t.Parallel()
	rows := Rows(Seed(0, "suite", "scenario"), 4, []Var{
		{Name: "n", Kind: Int, Min: 0, Max: 100},
		{Name: "s", Kind: Alpha, Min: 1, Max: 4},
	})
	want := []map[string]string{
		{"n": "0", "s": "a"},
		{"n": "100", "s": "oYBq"},
		{"n": "50", "s": "Dj"},
		{"n": "16", "s": "Lotf"},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows = %v, want %d rows", rows, len(want))
	}
	for i := range want {
		for k, v := range want[i] {
			if rows[i][k] != v {
				t.Errorf("row %d %q = %q, want %q (a changed stream retests every forall spec on different inputs)", i, k, rows[i][k], v)
			}
		}
	}
}

// TestSeed_SeparatesScenarios: two scenarios in one suite must not test the
// same generated inputs, or a file with three scenarios is one scenario run
// three times.
func TestSeed_SeparatesScenarios(t *testing.T) {
	t.Parallel()
	a := Rows(Seed(0, "suite", "first"), 5, []Var{alphaVar(4, 4)})
	b := Rows(Seed(0, "suite", "second"), 5, []Var{alphaVar(4, 4)})
	if same(a, b) {
		t.Error("two scenario names produced the same values; the scenario has to be part of the seed")
	}
	if !same(a, Rows(Seed(0, "suite", "first"), 5, []Var{alphaVar(4, 4)})) {
		t.Error("the same scenario produced different values")
	}
	if same(a, Rows(Seed(1, "suite", "first"), 5, []Var{alphaVar(4, 4)})) {
		t.Error("the spec's own seed did not change the values")
	}
}

// TestRows_NothingToDo covers the two ways a caller can ask for no work.
func TestRows_NothingToDo(t *testing.T) {
	t.Parallel()
	if rows := Rows(1, 0, []Var{alphaVar(1, 2)}); rows != nil {
		t.Errorf("rows = %v, want none for runs = 0", rows)
	}
	if rows := Rows(1, 5, nil); rows != nil {
		t.Errorf("rows = %v, want none with no variables", rows)
	}
}

// TestKinds_AreTheNamedOnes: Kinds is what the loader validates against and what
// its error message lists, so it must name every kind a `type:` accepts and
// nothing else — one_of is spelled by listing the choices.
func TestKinds_AreTheNamedOnes(t *testing.T) {
	t.Parallel()
	want := []string{"alpha", "alphanumeric", "ascii", "bool", "digits", "int", "unicode"}
	got := Kinds()
	if len(got) != len(want) {
		t.Fatalf("Kinds() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Kinds()[%d] = %q, want %q (sorted)", i, got[i], want[i])
		}
		if !Known(want[i]) {
			t.Errorf("Known(%q) = false", want[i])
		}
	}
	for _, name := range []string{"one_of", "string", "", "INT"} {
		if Known(name) {
			t.Errorf("Known(%q) = true, want false", name)
		}
	}
}

// TestBoundedAndDefaults pins which kinds read a range and what they default to,
// since the loader refuses min/max on the kinds that do not.
func TestBoundedAndDefaults(t *testing.T) {
	t.Parallel()
	for _, k := range []Kind{Int, Digits, Alpha, Alphanumeric, ASCII, Unicode} {
		if !Bounded(k) {
			t.Errorf("Bounded(%q) = false, want true", k)
		}
	}
	for _, k := range []Kind{Bool, OneOf} {
		if Bounded(k) {
			t.Errorf("Bounded(%q) = true, want false", k)
		}
	}
	if low, high := Defaults(Int); low != 0 || high != 100 {
		t.Errorf("Defaults(int) = %d, %d; want 0, 100", low, high)
	}
	if low, high := Defaults(Alpha); low != 1 || high != 8 {
		t.Errorf("Defaults(alpha) = %d, %d; want 1, 8", low, high)
	}
}
