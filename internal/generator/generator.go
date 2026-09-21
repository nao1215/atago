// Package generator produces the values a scenario's `forall:` block binds to its
// variables (#656) — the generator half of property-based testing, ported from
// the Gleam library metamon (https://github.com/nao1215/metamon).
//
// Everything here is deterministic. A generated value is not a random value:
// it is a value derived from a seed the spec itself carries, so the same spec
// produces the same inputs on every machine, on every run, and in every report,
// and a failing instance can be re-run by name. That is also why this package
// carries its own PRNG rather than calling math/rand, which this repository's
// linter refuses outright: atago's output is a contract, and a test runner that
// cannot reproduce its own inputs cannot tell a real failure from its noise.
package generator

import (
	"sort"
	"strconv"
	"strings"
)

// Kind names one generator. The set is deliberately small and CLI-shaped: a
// generated value is substituted into a command line, an environment variable,
// stdin, or a fixture, so what matters is the shapes an argument can take —
// numbers, digits, letters, printable text, non-ASCII text, a fixed set of
// choices — rather than metamon's full algebra of generators over Gleam types.
type Kind string

// The generator kinds a spec may name.
const (
	// Int is an integer in [Min, Max], rendered in base 10.
	Int Kind = "int"
	// Bool is "true" or "false".
	Bool Kind = "bool"
	// Digits is a string of 0-9 whose length is in [Min, Max].
	Digits Kind = "digits"
	// Alpha is a string of a-zA-Z.
	Alpha Kind = "alpha"
	// Alphanumeric is a string of a-zA-Z0-9.
	Alphanumeric Kind = "alphanumeric"
	// ASCII is a string of printable ASCII, minus the characters that would
	// change how a command line is read rather than what it says: the quotes,
	// the backslash, the backtick, and the dollar sign that starts a `${name}`
	// reference. A generator whose values could re-quote the command under test
	// would report tokenizer errors instead of the program's behavior.
	ASCII Kind = "ascii"
	// Unicode is a string of non-ASCII letters — Latin-1, Greek, Cyrillic, kana
	// and CJK — for the "does this CLI handle text it did not write" question.
	Unicode Kind = "unicode"
	// OneOf picks from the fixed set of values the spec lists.
	OneOf Kind = "one_of"
)

// namedKinds is the single source of truth for what a spec's `type:` may name.
// The loader validates against it, its error message lists it, and a drift test
// compares it against the schema's enum. OneOf is deliberately not in it: a
// choice generator is spelled by listing the choices (`{one_of: [a, b]}`), so
// naming it as a type would be a second spelling of the same thing.
var namedKinds = []Kind{Int, Bool, Digits, Alpha, Alphanumeric, ASCII, Unicode}

// Kinds returns every generator kind a `type:` may name, sorted.
func Kinds() []string {
	out := make([]string, len(namedKinds))
	for i, k := range namedKinds {
		out[i] = string(k)
	}
	sort.Strings(out)
	return out
}

// Known reports whether name is a generator kind a `type:` may name.
func Known(name string) bool {
	for _, k := range namedKinds {
		if string(k) == name {
			return true
		}
	}
	return false
}

// Bounded reports whether a kind reads Min/Max at all. Bool and OneOf do not:
// their value space is fixed by the kind, so a range on them is an authoring
// mistake rather than a narrowing.
func Bounded(k Kind) bool {
	return k != Bool && k != OneOf
}

// Defaults returns the Min/Max a kind uses when the spec states none: an
// integer range for Int, a length range for the string kinds. They are small on
// purpose — every generated row is a scenario that spawns a process, so the
// defaults are sized for "a handful of shapes", not for a fuzzing campaign.
func Defaults(k Kind) (minimum, maximum int) {
	if k == Int {
		return 0, 100
	}
	return 1, 8
}

// Var is one bound variable: the name a scenario references as ${name}, the
// kind that produces its values, and the bounds or choices that shape them.
type Var struct {
	Name   string
	Kind   Kind
	Min    int
	Max    int
	Values []string
}

// Rows generates up to runs distinct variable bindings, in a fixed order.
//
// Boundary values come first, the way metamon consumes a generator's edges
// before it samples: the ends of an integer range, the shortest and the longest
// string, both booleans, every `one_of` choice. Those are where a CLI breaks —
// the empty argument, the one that is too long, the flag value nobody tries —
// so a suite that runs ten instances spends the first of them there.
//
// Rows are deduplicated, which makes runs an upper bound rather than a promise:
// a `one_of` over two choices expands to two scenarios however many runs are
// asked for, instead of to twenty scenarios that are eighteen copies. When the
// space is larger than runs, generation stops as soon as runs distinct rows
// exist; when it is smaller, it stops when drawing has stopped finding anything
// new.
func Rows(seed uint64, runs int, vars []Var) []map[string]string {
	if runs <= 0 || len(vars) == 0 {
		return nil
	}
	r := &rng{state: seed}
	seen := make(map[string]bool, runs)
	rows := make([]map[string]string, 0, runs)
	// A draw that repeats an earlier row costs one attempt, so a small space
	// (two booleans, three choices) is exhausted rather than retried forever.
	// The budget is generous enough that a large space fills runs on the first
	// pass and cheap enough that an exhausted one stops in microseconds.
	for attempt := 0; len(rows) < runs && attempt < runs*32; attempt++ {
		row := make(map[string]string, len(vars))
		for i := range vars {
			row[vars[i].Name] = value(r, &vars[i], attempt)
		}
		key := rowKey(row)
		if seen[key] {
			continue
		}
		seen[key] = true
		rows = append(rows, row)
	}
	return rows
}

// value returns one variable's value for the given attempt: its attempt-th edge
// while it has one, and a sampled value afterwards.
func value(r *rng, v *Var, attempt int) string {
	if e := edges(r, v); attempt < len(e) {
		return e[attempt]
	}
	return sample(r, v)
}

// edges returns the boundary values for a variable, in the order they are tried.
// The string edges still draw their characters from r, so asking for them
// advances the stream exactly as a sample would.
func edges(r *rng, v *Var) []string {
	switch v.Kind {
	case Bool:
		return []string{"false", "true"}
	case OneOf:
		return v.Values
	case Int:
		if v.Min == v.Max {
			return []string{strconv.Itoa(v.Min)}
		}
		return []string{strconv.Itoa(v.Min), strconv.Itoa(v.Max)}
	case Digits, Alpha, Alphanumeric, ASCII, Unicode:
		if v.Min == v.Max {
			return []string{text(r, v.Kind, v.Min)}
		}
		return []string{text(r, v.Kind, v.Min), text(r, v.Kind, v.Max)}
	}
	return nil
}

// sample draws one value uniformly from the variable's space.
func sample(r *rng, v *Var) string {
	switch v.Kind {
	case Bool:
		if r.intn(2) == 0 {
			return "false"
		}
		return "true"
	case OneOf:
		if len(v.Values) == 0 {
			return ""
		}
		return v.Values[r.intn(len(v.Values))]
	case Int:
		return strconv.Itoa(v.Min + r.intn(v.Max-v.Min+1))
	case Digits, Alpha, Alphanumeric, ASCII, Unicode:
		return text(r, v.Kind, v.Min+r.intn(v.Max-v.Min+1))
	}
	return ""
}

// The alphabets the string kinds draw from. ASCII is printable ASCII minus
// `'`, `"`, `\`, backtick and `$` — see the Kind doc comment for why.
const (
	digitChars = "0123456789"
	alphaChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	asciiChars = " !#%&()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_" +
		"abcdefghijklmnopqrstuvwxyz{|}~"
)

// unicodeRanges are the non-ASCII code point ranges the unicode kind draws
// from: Latin-1 letters, Greek, Cyrillic, hiragana, katakana, and common CJK.
// They are all printable, single-scalar and assigned, so a generated value is
// text a terminal and a report can show rather than an unassigned code point or
// half of a surrogate pair.
var unicodeRanges = [][2]rune{
	{0x00C0, 0x00FF}, // Latin-1 letters
	{0x0391, 0x03C9}, // Greek
	{0x0410, 0x044F}, // Cyrillic
	{0x3041, 0x3093}, // hiragana
	{0x30A1, 0x30F3}, // katakana
	{0x4E00, 0x4EFF}, // CJK unified ideographs (start of the block)
}

// text builds a string of n characters of the given kind.
func text(r *rng, k Kind, n int) string {
	if n <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		switch k {
		case Digits:
			b.WriteByte(digitChars[r.intn(len(digitChars))])
		case Alpha:
			b.WriteByte(alphaChars[r.intn(len(alphaChars))])
		case Alphanumeric:
			pool := alphaChars + digitChars
			b.WriteByte(pool[r.intn(len(pool))])
		case ASCII:
			b.WriteByte(asciiChars[r.intn(len(asciiChars))])
		case Unicode:
			rg := unicodeRanges[r.intn(len(unicodeRanges))]
			span := int(rg[1]-rg[0]) + 1
			b.WriteRune(rg[0] + rune(r.intn(span))) //nolint:gosec // the draw is below span, a range width inside the BMP, so the sum is a valid rune
		case Int, Bool, OneOf:
			return ""
		}
	}
	return b.String()
}

// rowKey renders a row as a comparable string so two draws that bound the same
// values are recognized as the same row. The separators cannot appear in a
// variable name, so no two distinct rows can collide on one key.
func rowKey(row map[string]string) string {
	names := make([]string, 0, len(row))
	for k := range row {
		names = append(names, k)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		b.WriteString(n)
		b.WriteByte('\x00')
		b.WriteString(row[n])
		b.WriteByte('\x01')
	}
	return b.String()
}

// Seed derives the stream a scenario generates from: the spec's own `seed:`
// mixed with the text that identifies the scenario. Two scenarios in one suite
// therefore never draw the same values from the default seed — which would make
// a two-scenario suite test one input twice — while each of them keeps drawing
// the same values until someone edits the spec.
func Seed(userSeed int, parts ...string) uint64 {
	// FNV-1a over the identifying text, then one round of the PRNG's mixer so
	// neighboring scenario names do not produce neighboring streams.
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	h := uint64(offset64)
	for _, p := range parts {
		for i := 0; i < len(p); i++ {
			h ^= uint64(p[i])
			h *= prime64
		}
		h ^= '\x00'
		h *= prime64
	}
	h ^= uint64(userSeed) //nolint:gosec // a spec's seed is an identifier, not a quantity; wraparound is fine
	r := &rng{state: h}
	return r.next()
}

// rng is splitmix64: a 64-bit state, one addition and three xor-shift-multiply
// rounds per draw. It is here rather than in the standard library because
// atago's values have to be identical on every OS, architecture and Go release
// — a generated input that changed under the user's feet would change what a
// spec asserts — and because this repository's linter forbids math/rand for
// exactly that reason.
type rng struct {
	state uint64
}

// next returns the next 64 bits of the stream.
func (r *rng) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// intn returns a value in [0, n). n must be positive.
func (r *rng) intn(n int) int {
	if n <= 1 {
		return 0
	}
	// Modulo of a 63-bit draw. The bias this leaves is on the order of n/2^63
	// for the ranges a spec can ask for (a string length, a small integer
	// range, a list index), which is far below anything a test could observe,
	// and a rejection loop would cost determinism nothing but clarity.
	return int((r.next() >> 1) % uint64(n)) //nolint:gosec // n > 1, and the shift keeps the value below 2^63
}
