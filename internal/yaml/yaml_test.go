package yaml

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
)

// value parses src and returns the generic value of its first document.
func value(t *testing.T, src string) any {
	t.Helper()
	f, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse(%q): %v", src, err)
	}
	v, err := Value(f.First())
	if err != nil {
		t.Fatalf("Value(%q): %v", src, err)
	}
	return v
}

func TestParseValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
		want any
	}{
		{"block mapping", "a: 1\nb: x\n", map[string]any{"a": int64(1), "b": "x"}},
		{"nested", "a:\n  b:\n    - 1\n    - two\n", map[string]any{"a": map[string]any{"b": []any{int64(1), "two"}}}},
		{"compact sequence under key", "a:\n- x\n- y\nb: z\n", map[string]any{"a": []any{"x", "y"}, "b": "z"}},
		{"mapping in sequence", "- a: 1\n  b: 2\n- c: 3\n", []any{map[string]any{"a": int64(1), "b": int64(2)}, map[string]any{"c": int64(3)}}},
		{"nested sequence", "- - a\n  - b\n- c\n", []any{[]any{"a", "b"}, "c"}},
		{"flow", "{a: [1, 2, {b: c}], d: e}", map[string]any{"a": []any{int64(1), int64(2), map[string]any{"b": "c"}}, "d": "e"}},
		{"flow over lines with comments", "[a, # one\n  b,\n  c]", []any{"a", "b", "c"}},
		{"flow pair in sequence", "[a: b, c]", []any{map[string]any{"a": "b"}, "c"}},
		{"flow key without value", "{a, b: 1}", map[string]any{"a": nil, "b": int64(1)}},
		{"null forms", "a:\nb: ~\nc: null\nd: !!null x", map[string]any{"a": nil, "b": nil, "c": nil, "d": nil}},
		{"bools", "[true, True, FALSE, yes, on]", []any{true, true, false, "yes", "on"}},
		{"ints", "[0, -3, +4, 0x1f, 0o17, 017, 0b101, 1_000]", []any{int64(0), int64(-3), int64(4), int64(31), int64(15), int64(15), int64(5), int64(1000)}},
		{"words that are not ints", "[09, 0X1, 0x, _1, 1e]", []any{"09", "0X1", "0x", "_1", "1e"}},
		{"floats", "[1.5, -0.25, 1e3, .5, 1., .inf, -.Inf]", []any{1.5, -0.25, 1000.0, 0.5, 1.0, math.Inf(1), math.Inf(-1)}},
		{"large ints", "[18446744073709551615, 18446744073709551616]", []any{uint64(math.MaxUint64), "18446744073709551616"}},
		{"plain keeps its text", "v: 1.20\n", map[string]any{"v": 1.2}},
		{"quoted is a string", "['1', \"true\", 'it''s']", []any{"1", "true", "it's"}},
		{"double quoted escapes", `"a\tb\n\x41\u00e9\U0001F600\\\""`, "a\tb\nAé😀\\\""},
		{"double quoted folding", "\"one\n  two\n\n  three\"", "one two\nthree"},
		{"escaped break joins", "\"one\\\n  two\"", "onetwo"},
		{"escaped blank survives a fold", "\"a\\t\n b\"", "a\t b"},
		{"plain over lines", "a: one\n  two\n\n  three\nb: x", map[string]any{"a": "one two\nthree", "b": "x"}},
		{"comment ends plain", "a: b # c\n", map[string]any{"a": "b"}},
		{"hash without space is text", "a: b#c", map[string]any{"a": "b#c"}},
		{"url is not a key", "a: http://x:8080/y", map[string]any{"a": "http://x:8080/y"}},
		{"literal clip", "a: |\n  x\n   y\n\nb: 1", map[string]any{"a": "x\n y\n", "b": int64(1)}},
		{"literal strip", "a: |-\n  x\n\n", map[string]any{"a": "x"}},
		{"literal keep", "a: |+\n  x\n\n", map[string]any{"a": "x\n\n"}},
		{"literal keeps trailing blanks", "a: |-\n  x  \n", map[string]any{"a": "x  "}},
		{"literal indent indicator", "a: |2\n    x\n  y\n", map[string]any{"a": "  x\ny\n"}},
		{"folded", "a: >\n  one\n  two\n\n  three\n    more\n  end\n", map[string]any{"a": "one two\nthree\n  more\nend\n"}},
		{"literal at end without break", "a: |\n  x", map[string]any{"a": "x"}},
		{"anchor and alias", "a: &x {b: 1}\nc: *x\n", map[string]any{"a": map[string]any{"b": int64(1)}, "c": map[string]any{"b": int64(1)}}},
		{"anchor on its own line", "a: &x\n  b: 1\nc: *x\n", map[string]any{"a": map[string]any{"b": int64(1)}, "c": map[string]any{"b": int64(1)}}},
		{"anchored scalar in flow", "[&x a, *x]", []any{"a", "a"}},
		{"alias as key", "a: &k key\n*k : v\n", map[string]any{"a": "key", "key": "v"}},
		{"merge key", "base: &b {a: 1, b: 2}\nx:\n  <<: *b\n  b: 3\n", map[string]any{"base": map[string]any{"a": int64(1), "b": int64(2)}, "x": map[string]any{"a": int64(1), "b": int64(3)}}},
		{"merge list", "a: &a {x: 1}\nb: &b {x: 2, y: 2}\nc:\n  <<: [*a, *b]\n", map[string]any{"a": map[string]any{"x": int64(1)}, "b": map[string]any{"x": int64(2), "y": int64(2)}, "c": map[string]any{"x": int64(1), "y": int64(2)}}},
		{"tags", "[!!str 1, !!int '2', !!float 3, !!bool 'true', !Ref x, ! 12]", []any{"1", int64(2), 3.0, true, "x", "12"}},
		{"binary", "!!binary aGk=", "hi"},
		{"tag on a key", "!!str a: b", map[string]any{"a": "b"}},
		{"explicit key", "? a\n: b\n", map[string]any{"a": "b"}},
		{"document markers", "%YAML 1.2\n---\na: 1\n...\n", map[string]any{"a": int64(1)}},
		{"first document with content", "---\n# nothing\n---\nb\n---\nc\n", "b"},
		{"crlf", "a: 1\r\nb: |\r\n  x\r\n", map[string]any{"a": int64(1), "b": "x\n"}},
		{"lone cr", "a: 1\rb: 2", map[string]any{"a": int64(1), "b": int64(2)}},
		{"bom", "\ufeffa: 1", map[string]any{"a": int64(1)}},
		{"key with spaces and quotes", "\"a b\": 1\n'c': 2", map[string]any{"a b": int64(1), "c": int64(2)}},
		{"empty flow collections", "a: []\nb: {}", map[string]any{"a": []any{}, "b": map[string]any{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := value(t, tt.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("value(%q)\n got  %#v\n want %#v", tt.src, got, tt.want)
			}
		})
	}
}

func TestParseNaN(t *testing.T) {
	t.Parallel()
	if f, ok := value(t, ".nan").(float64); !ok || !math.IsNaN(f) {
		t.Errorf(".nan = %#v, want NaN", f)
	}
}

func TestParseEmpty(t *testing.T) {
	t.Parallel()
	for _, src := range []string{"", "   \n", "# only a comment\n", "---\n", "---\n...\n"} {
		f, err := Parse([]byte(src))
		if err != nil {
			t.Errorf("Parse(%q): %v", src, err)
			continue
		}
		if f.First() != nil {
			t.Errorf("Parse(%q).First() = %#v, want nil", src, f.First())
		}
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src       string
		line, col int
		msg       string
	}{
		{"a: b: c", 1, 4, "mapping value is not allowed"},
		{"a: 1\n  b: 2", 2, 3, "a line that continues a value cannot hold a key"},
		{"a:\n  b:\n    c: 1\n   d: 2", 4, 4, "unexpected indentation"},
		{"a: 1\na: 2", 2, 1, `mapping key "a" is written twice (first at line 1)`},
		{"a: [1, 2", 1, 8, "never closed"},
		{"a: \"x", 1, 4, "never closed"},
		{"a: - b", 1, 4, "sequence cannot start on the line of its key"},
		{"a:\n\tb: 1", 2, 1, "tab character cannot indent"},
		{"a: *nope", 1, 4, "alias *nope names no anchor"},
		{"a: 1\n- b", 2, 1, "sequence item stands where a key was expected"},
		{"a: \"x\" y", 1, 8, "unexpected text after the quoted value"},
		{"a: [x] y", 1, 8, "unexpected text after the flow collection"},
		{"%YAML 2.0\n---\na", 1, 1, "only YAML 1 is read"},
		{"%YAML 1.2\na", 1, 1, "has to be followed by \"---\""},
		{"a: \"\\q\"", 1, 5, "unknown escape"},
		{"a: @x", 1, 4, "cannot start a plain value"},
		{"- a\nb: 1", 2, 1, "unexpected text after the document"},
		{"<<: 1", 1, 1, "merge key (<<) takes a mapping"},
		{"a: & x", 1, 4, "anchor needs a name"},
	}
	for _, tt := range tests {
		_, err := Parse([]byte(tt.src))
		var e *Error
		if !errors.As(err, &e) {
			t.Errorf("Parse(%q) error = %v, want *Error", tt.src, err)
			continue
		}
		if e.Line != tt.line || e.Col != tt.col || !strings.Contains(e.Msg, tt.msg) {
			t.Errorf("Parse(%q) = [%d:%d] %s, want [%d:%d] containing %q", tt.src, e.Line, e.Col, e.Msg, tt.line, tt.col, tt.msg)
		}
	}
}

func TestNodePositions(t *testing.T) {
	t.Parallel()
	src := "suite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - &s\n        assert: {}\n"
	f, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	root := f.First()
	check := func(what string, n *Node, line, col int) {
		t.Helper()
		if n == nil || n.Line != line || n.Col != col {
			t.Errorf("%s at %v, want %d:%d", what, n, line, col)
		}
	}
	check("suite.name", root.Get("suite").Get("name"), 2, 9)
	check("scenario", root.Get("scenarios").Index(0), 4, 5)
	check("scenario name", root.Get("scenarios").Index(0).Get("name"), 4, 11)
	check("step", root.Get("scenarios").Index(0).Get("steps").Index(0), 6, 9)
	check("flow mapping", root.Get("scenarios").Index(0).Get("steps").Index(0).Get("run"), 6, 14)
	check("anchored step starts at its anchor", root.Get("scenarios").Index(0).Get("steps").Index(1), 7, 9)
	if root.Get("nope") != nil || root.Index(0) != nil || root.Get("scenarios").Index(9) != nil {
		t.Error("a lookup that names nothing returned a node")
	}
}

func TestFormatError(t *testing.T) {
	t.Parallel()
	src := []byte("a: 1\nb: 2\nc: 3\nd: e: f\ne: 5\nf: 6\ng: 7\nh: 8\n")
	_, err := Parse(src)
	want := `[4:4] a mapping value is not allowed here; quote the value if it contains ": ", or start a nested mapping on the next line
   1 | a: 1
   2 | b: 2
   3 | c: 3
>  4 | d: e: f
          ^
   5 | e: 5
   6 | f: 6
   7 | g: 7`
	if got := FormatError(err, src); got != want {
		t.Errorf("FormatError =\n%s\nwant\n%s", got, want)
	}
	if got := FormatError(errors.New("plain"), src); got != "plain" {
		t.Errorf("FormatError(plain) = %q", got)
	}
}

type inner struct {
	C string `yaml:"c"`
}

type outer struct {
	A     string            `yaml:"a"`
	N     int               `yaml:"n"`
	U     uint8             `yaml:"u"`
	F     float64           `yaml:"f"`
	B     *bool             `yaml:"b"`
	L     []string          `yaml:"l"`
	M     map[string]string `yaml:"m"`
	Any   any               `yaml:"any"`
	In    inner             `yaml:",inline"`
	Ptr   *inner            `yaml:"ptr"`
	Skip  string            `yaml:"-"`
	Lower string
}

func TestDecodeStruct(t *testing.T) {
	t.Parallel()
	src := "a: 1.20\nn: 0x10\nu: 7\nf: 2\nb: true\nl: [x, 007]\nm: {k: 1.0}\nany: 1.5\nc: inline\nptr: {c: p}\nlower: y\n"
	f, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	var o outer
	if err := Decode(f.First(), &o, true); err != nil {
		t.Fatal(err)
	}
	tr := true
	want := outer{A: "1.20", N: 16, U: 7, F: 2, B: &tr, L: []string{"x", "007"}, M: map[string]string{"k": "1.0"}, Any: 1.5, In: inner{C: "inline"}, Ptr: &inner{C: "p"}, Lower: "y"}
	if !reflect.DeepEqual(o, want) {
		t.Errorf("decoded\n%#v\nwant\n%#v", o, want)
	}
}

func TestDecodeErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		src  string
		want Error
	}{
		{"a: [1]", Error{Line: 1, Col: 4, Path: "a", Want: "a string", Got: SequenceNode}},
		{"l: x", Error{Line: 1, Col: 4, Path: "l", Want: "a list", Got: ScalarNode}},
		{"ptr: [1]", Error{Line: 1, Col: 6, Path: "ptr", Want: "a mapping", Got: SequenceNode}},
		{"n: 1.5", Error{Line: 1, Col: 4, Path: "n"}},
		{"u: -1", Error{Line: 1, Col: 4, Path: "u"}},
		{"u: 300", Error{Line: 1, Col: 4, Path: "u"}},
		{"b: yes", Error{Line: 1, Col: 4, Path: "b"}},
		{"f: x", Error{Line: 1, Col: 4, Path: "f"}},
		{"l: [a, {b: c}]", Error{Line: 1, Col: 8, Path: "l[1]", Want: "a string", Got: MappingNode}},
		// The first problem in the order written is the one reported.
		{"a: x\nzz: 1\nn: q", Error{Line: 2, Col: 1, Key: "zz"}},
		{"n: q\nzz: 1", Error{Line: 1, Col: 4, Path: "n"}},
	}
	for _, tt := range tests {
		f, err := Parse([]byte(tt.src))
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.src, err)
		}
		var o outer
		err = Decode(f.First(), &o, true)
		var e *Error
		if !errors.As(err, &e) {
			t.Errorf("Decode(%q) = %v, want *Error", tt.src, err)
			continue
		}
		got := Error{Line: e.Line, Col: e.Col, Path: e.Path, Key: e.Key, Want: e.Want, Got: e.Got}
		if tt.want.Key != "" {
			got.Path = ""
		}
		if got != tt.want {
			t.Errorf("Decode(%q) = %+v (%s), want %+v", tt.src, got, e.Msg, tt.want)
		}
	}
}

func TestDecodeNotStrict(t *testing.T) {
	t.Parallel()
	var o outer
	if err := Unmarshal([]byte("zz: 1\na: x"), &o); err != nil || o.A != "x" {
		t.Errorf("Unmarshal = %v, %+v; want the unknown key left out", err, o)
	}
}

func TestDecodeNull(t *testing.T) {
	t.Parallel()
	o := outer{A: "keep", Ptr: &inner{}}
	if err := Unmarshal([]byte("a: null\nptr: ~"), &o); err != nil {
		t.Fatal(err)
	}
	if o.A != "" || o.Ptr != nil {
		t.Errorf("null = %+v, want an empty string and a nil pointer", o)
	}
	if err := Decode(nil, &o, true); err != nil {
		t.Errorf("Decode(nil) = %v, want nil", err)
	}
	if err := Decode(nil, o, true); err == nil {
		t.Error("Decode into a non-pointer = nil error")
	}
}

// shape is a type that reads itself, the way a spec field that takes more than
// one shape does.
type shape struct{ got string }

func (s *shape) UnmarshalYAML(n *Node) error {
	if n.Kind == MappingNode {
		return errors.New("no mappings here")
	}
	s.got = n.Kind.String() + ":" + n.Value
	return nil
}

func TestDecodeUnmarshaler(t *testing.T) {
	t.Parallel()
	var v struct {
		S shape  `yaml:"s"`
		P *shape `yaml:"p"`
	}
	if err := Unmarshal([]byte("s: 1.20\np: x"), &v); err != nil {
		t.Fatal(err)
	}
	if v.S.got != "scalar:1.20" || v.P == nil || v.P.got != "scalar:x" {
		t.Errorf("unmarshaler got %+v %+v", v.S, v.P)
	}
	err := Unmarshal([]byte("s:\n  k: v"), &v)
	var e *Error
	if !errors.As(err, &e) || e.Line != 2 || e.Col != 3 || e.Msg != "no mappings here" {
		t.Errorf("unmarshaler error = %v, want it placed at the node", err)
	}
}

func TestDecodeBinary(t *testing.T) {
	t.Parallel()
	var v struct {
		S string `yaml:"s"`
	}
	if err := Unmarshal([]byte("s: !!binary |\n  aGVs\n  bG8=\n"), &v); err != nil || v.S != "hello" {
		t.Errorf("binary = %q, %v", v.S, err)
	}
	if err := Unmarshal([]byte("s: !!binary '%%%'"), &v); err == nil {
		t.Error("invalid base64 decoded without an error")
	}
}

// TestAliasExpansionIsBounded holds a document of nested aliases, which stands
// for far more values than it has lines, to an error rather than letting it
// take every byte of memory there is.
func TestAliasExpansionIsBounded(t *testing.T) {
	t.Parallel()
	var b strings.Builder
	b.WriteString("a0: &a0 [x, x, x, x, x, x, x, x, x, x]\n")
	for i := 1; i < 10; i++ {
		fmt.Fprintf(&b, "a%d: &a%d [", i, i)
		for j := range 10 {
			if j > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "*a%d", i-1)
		}
		b.WriteString("]\n")
	}
	var v any
	err := Unmarshal([]byte(b.String()), &v)
	if err == nil || !strings.Contains(err.Error(), ErrTooLarge.Error()) {
		t.Errorf("alias bomb decoded with error %v, want %v", err, ErrTooLarge)
	}
}

func TestDepthIsBounded(t *testing.T) {
	t.Parallel()
	src := strings.Repeat("[", MaxDepth+1) + strings.Repeat("]", MaxDepth+1)
	if _, err := Parse([]byte(src)); err == nil || !strings.Contains(err.Error(), "nests deeper") {
		t.Errorf("deep nesting = %v, want the depth error", err)
	}
	var b strings.Builder
	for i := range MaxDepth + 1 {
		b.WriteString(strings.Repeat(" ", i))
		b.WriteString("- \n")
	}
	if _, err := Parse([]byte(strings.ReplaceAll(b.String(), "- \n", "-\n"))); err != nil && !strings.Contains(err.Error(), "nests deeper") {
		t.Errorf("deep block nesting = %v", err)
	}
}

// TestQuoteRoundTrips is the law `atago record` relies on: whatever the string,
// Quote(s) written after a key reads back as s.
func TestQuoteRoundTrips(t *testing.T) {
	t.Parallel()
	cases := []string{
		"", " ", "plain", "two words", "1.20", "007", "true", "null", "~", "-", "- x", "-x", "?", "? x",
		":", "a: b", "a #b", "a#b", "#", "&a", "*a", "!t", "|", ">", "'", `"`, "%", "@", "`", "[", "{", ",",
		"---", "...", "a\tb", "line\nbreak", "cr\r", "esc\x1b", "del\x7f", "\u0085", "\u2028", "\ufeff",
		"back\\slash", "日本語", "trailing ", " leading", ".inf", "0x10", "1e3", "http://x:1/y",
	}
	r := rand.New(rand.NewPCG(1, 2))
	alphabet := []rune("ab :#-?'\"\\\n\t{}[],&*!|>%@`~.01e\u00e9\u3042")
	for range 2000 {
		n := r.IntN(8)
		rs := make([]rune, n)
		for i := range rs {
			rs[i] = alphabet[r.IntN(len(alphabet))]
		}
		cases = append(cases, string(rs))
	}
	for _, s := range cases {
		q := Quote(s)
		if strings.ContainsAny(q, "\n\r") {
			t.Errorf("Quote(%q) = %q spans lines", s, q)
		}
		var v map[string]string
		if err := Unmarshal([]byte("v: "+q+"\n"), &v); err != nil {
			t.Errorf("Quote(%q) = %s does not parse: %v", s, q, err)
			continue
		}
		if v["v"] != s {
			t.Errorf("Quote(%q) = %s reads back as %q", s, q, v["v"])
		}
	}
	if Quote("plain") != "plain" || Quote("two words") != "two words" {
		t.Error("a plain word was quoted")
	}
}

func TestMarshalRoundTrips(t *testing.T) {
	t.Parallel()
	in := map[string]any{
		"s":     []any{"a", "1.20", "", "x: y", int64(3), -2.5, true, nil, map[string]any{}, []any{}},
		"m":     map[string]any{"k": map[string]any{"n": []any{[]any{"deep"}}}},
		"inf":   math.Inf(1),
		"whole": 2.0,
	}
	out, err := Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var back any
	if err := Unmarshal(out, &back); err != nil {
		t.Fatalf("Unmarshal(Marshal): %v\n%s", err, out)
	}
	if !reflect.DeepEqual(back, in) {
		t.Errorf("round trip\n%s\ngot  %#v\nwant %#v", out, back, in)
	}
}

type marshalOrder struct {
	Z   string   `yaml:"z"`
	A   string   `yaml:"a,omitempty"`
	L   []string `yaml:"l,omitempty"`
	In  inner    `yaml:",inline"`
	Ptr *inner   `yaml:"ptr,omitempty"`
}

func TestMarshalStruct(t *testing.T) {
	t.Parallel()
	out, err := Marshal([]any{marshalOrder{Z: "z", L: []string{"x"}, In: inner{C: "c"}}, MapSlice{{"b", 1}, {"a", nil}}})
	if err != nil {
		t.Fatal(err)
	}
	want := "- z: z\n  l:\n    - x\n  c: c\n- b: 1\n  a: null\n"
	if string(out) != want {
		t.Errorf("Marshal =\n%s\nwant\n%s", out, want)
	}
	if _, err := Marshal(make(chan int)); err == nil {
		t.Error("Marshal(chan) = nil error")
	}
}
