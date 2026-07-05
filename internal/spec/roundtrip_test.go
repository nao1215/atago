package spec

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/goccy/go-yaml"
)

// Three spec fields decode several authoring shapes through a custom
// UnmarshalYAML (a scalar or one of a few mappings). Marshaling them back must
// emit a shape the same unmarshaler accepts, or a loaded spec cannot be written
// out and re-read — the round-trip a future canonicalizer/formatter would rely
// on. These tests pin that marshal/unmarshal symmetry per type; the default
// struct marshal breaks it (it writes fields like `equals`/`inline`/`text` the
// unmarshaler rejects or silently drops).

func TestExitCode_YAMLRoundTrip(t *testing.T) {
	t.Parallel()
	cases := map[string]ExitCode{
		"equals zero":     {Equals: intp(0)},
		"equals nonzero":  {Equals: intp(2)},
		"equals negative": {Equals: intp(-1)},
		"not":             {Not: intp(1)},
		"in pair":         {In: []int{0, 2}},
		"in single":       {In: []int{127}},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := marshalReload(t, in)
			if !reflect.DeepEqual(in, got) {
				t.Errorf("exit_code round-trip:\n in  = %+v\n got = %+v", in, got)
			}
		})
	}
}

// TestExitCode_YAMLRoundTrip_Property fuzzes the integer forms with a fixed seed:
// any equals/not/in built from arbitrary ints must survive marshal→unmarshal.
func TestExitCode_YAMLRoundTrip_Property(t *testing.T) {
	t.Parallel()
	f := func(n int, form uint8) bool {
		var in ExitCode
		switch form % 3 {
		case 0:
			in.Equals = &n
		case 1:
			in.Not = &n
		default:
			in.In = []int{n, n + 1}
		}
		return reflect.DeepEqual(in, marshalReloadNoT(in))
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 500, Rand: rand.New(rand.NewSource(1))}); err != nil {
		t.Error(err)
	}
}

func TestStdin_YAMLRoundTrip(t *testing.T) {
	t.Parallel()
	cases := map[string]Stdin{
		"inline plain":    {Inline: "hello"},
		"inline empty":    {Inline: ""},
		"inline unicode":  {Inline: "日本語 \t emoji 💥 line1\nline2"},
		"inline yamlmeta": {Inline: "- : # * & [x]"},
		"file":            {File: "input.txt", mapped: true},
		"base64":          {Base64: "aGVsbG8=", mapped: true},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := marshalReload(t, in)
			if !reflect.DeepEqual(in, got) {
				t.Errorf("stdin round-trip:\n in  = %+v\n got = %+v", in, got)
			}
		})
	}
}

func TestPTYSend_YAMLRoundTrip(t *testing.T) {
	t.Parallel()
	// A lone carriage return is deliberately not exercised: YAML normalizes bare
	// CRs in scalars, so no marshaler could round-trip one — that key press is
	// authored as {key: enter}, which is covered below.
	cases := map[string]PTYSend{
		"text plain":     {Text: strp("yes")},
		"text empty":     {Text: strp("")},
		"text multiline": {Text: strp("line1\nline2")},
		"text escape":    {Text: strp("\x01\x1b[A")},
		"key enter":      {Key: "enter"},
		"key ctrl-c":     {Key: "ctrl-c"},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := marshalReload(t, in)
			if !reflect.DeepEqual(in, got) {
				t.Errorf("send round-trip:\n in  = %+v\n got = %+v", in, got)
			}
		})
	}
}

// marshalReload marshals v to YAML and decodes it back into a fresh T, failing
// the test on any encode/decode error. It is how each round-trip case checks
// that a value survives a write-then-read cycle.
func marshalReload[T any](t *testing.T, v T) T {
	t.Helper()
	b, err := yaml.Marshal(v)
	if err != nil {
		t.Fatalf("marshal %+v: %v", v, err)
	}
	var got T
	if err := yaml.Unmarshal(b, &got); err != nil {
		t.Fatalf("reload %q: %v", b, err)
	}
	return got
}

// marshalReloadNoT is the property-test variant: it returns the zero value on
// any error so quick.Check reports the offending input via a false result.
func marshalReloadNoT[T any](v T) T {
	var got T
	b, err := yaml.Marshal(v)
	if err != nil {
		return got
	}
	_ = yaml.Unmarshal(b, &got)
	return got
}
