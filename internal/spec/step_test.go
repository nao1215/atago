package spec

import (
	"reflect"
	"testing"

	"github.com/goccy/go-yaml"
)

func intp(i int) *int       { return &i }
func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

func TestStep_SetKeysAndKind(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		step Step
		want StepKind
	}{
		{"fixture", Step{Fixture: &Fixture{}}, StepFixture},
		{"run", Step{Run: &Run{}}, StepRun},
		{"http", Step{HTTP: &HTTP{}}, StepHTTP},
		{"query", Step{Query: &Query{}}, StepQuery},
		{"grpc", Step{GRPC: &GRPC{}}, StepGRPC},
		{"cdp", Step{CDP: &CDP{}}, StepCDP},
		{"assert", Step{Assert: &Assert{}}, StepAssert},
		{"store", Step{Store: &Store{}}, StepStore},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if keys := tc.step.SetKeys(); len(keys) != 1 || keys[0] != tc.want {
				t.Errorf("SetKeys = %v, want [%s]", keys, tc.want)
			}
			if got := tc.step.Kind(); got != tc.want {
				t.Errorf("Kind = %q, want %q", got, tc.want)
			}
		})
	}

	// None set and more-than-one set both collapse to StepNone.
	empty := Step{}
	if got := empty.Kind(); got != StepNone {
		t.Errorf("empty step Kind = %q, want none", got)
	}
	two := Step{Run: &Run{}, Assert: &Assert{}}
	if keys := two.SetKeys(); len(keys) != 2 {
		t.Errorf("two-action SetKeys = %v, want 2 keys", keys)
	}
	if got := two.Kind(); got != StepNone {
		t.Errorf("two-action Kind = %q, want none", got)
	}
}

func TestAssert_SetTargets(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a    Assert
		want AssertTarget
	}{
		{"exit_code", Assert{ExitCode: &ExitCode{}}, AssertExitCode},
		{"stdout", Assert{Stdout: &StreamAssert{}}, AssertStdout},
		{"stderr", Assert{Stderr: &StreamAssert{}}, AssertStderr},
		{"file", Assert{File: &FileAssert{}}, AssertFile},
		{"status", Assert{Status: intp(200)}, AssertStatus},
		{"header", Assert{Header: &HeaderMatch{}}, AssertHeader},
		{"body", Assert{Body: &StreamAssert{}}, AssertBody},
		{"rows", Assert{Rows: &StreamAssert{}}, AssertRows},
		{"grpc_status", Assert{GRPCStatus: intp(0)}, AssertGRPCStatus},
		{"message", Assert{Message: &StreamAssert{}}, AssertMessage},
		{"value", Assert{Value: &StreamAssert{}}, AssertValue},
		{"image", Assert{Image: &ImageAssert{}}, AssertImage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.a.SetTargets(); len(got) != 1 || got[0] != tc.want {
				t.Errorf("SetTargets = %v, want [%s]", got, tc.want)
			}
		})
	}
	emptyA := Assert{}
	if got := emptyA.SetTargets(); len(got) != 0 {
		t.Errorf("empty assert targets = %v, want none", got)
	}
}

func TestStreamAssert_SetMatchers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		s    StreamAssert
		want string
	}{
		{"empty", StreamAssert{Empty: boolp(true)}, "empty"},
		{"contains", StreamAssert{Contains: StringList{"x"}}, "contains"},
		{"not_contains", StreamAssert{NotContains: StringList{"x"}}, "not_contains"},
		{"matches", StreamAssert{Matches: strp("x")}, "matches"},
		{"equals", StreamAssert{Equals: strp("x")}, "equals"},
		{"not_equals", StreamAssert{NotEquals: strp("x")}, "not_equals"},
		{"json", StreamAssert{JSON: JSONChecks{{}}}, "json"},
		{"yaml", StreamAssert{YAML: JSONChecks{{}}}, "yaml"},
		{"snapshot", StreamAssert{Snapshot: "s.snap"}, "snapshot"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.s.SetMatchers(); !reflect.DeepEqual(got, []string{tc.want}) {
				t.Errorf("SetMatchers = %v, want [%s]", got, tc.want)
			}
		})
	}
	emptyS := StreamAssert{}
	if got := emptyS.SetMatchers(); len(got) != 0 {
		t.Errorf("empty stream matchers = %v, want none", got)
	}
}

// TestJSONChecks_UnmarshalYAML covers the #156 scalar-or-sequence decoding: a
// single mapping decodes to a one-element list, a sequence to its elements, and
// an empty sequence is rejected. It is decoded through a StreamAssert's json:
// field so it exercises the real YAML the loader sees.
func TestJSONChecks_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	t.Run("single mapping", func(t *testing.T) {
		var s StreamAssert
		if err := yaml.Unmarshal([]byte("json:\n  path: $.a\n  equals: 1\n"), &s); err != nil {
			t.Fatalf("unmarshal single: %v", err)
		}
		if len(s.JSON) != 1 || s.JSON[0].Path != "$.a" {
			t.Fatalf("single mapping = %+v, want one check at $.a", s.JSON)
		}
	})

	t.Run("list of mappings", func(t *testing.T) {
		var s StreamAssert
		src := "json:\n  - {path: $.a, equals: 1}\n  - {path: $.b, equals: 2}\n"
		if err := yaml.Unmarshal([]byte(src), &s); err != nil {
			t.Fatalf("unmarshal list: %v", err)
		}
		if len(s.JSON) != 2 || s.JSON[0].Path != "$.a" || s.JSON[1].Path != "$.b" {
			t.Fatalf("list = %+v, want two checks at $.a and $.b", s.JSON)
		}
	})

	t.Run("empty list rejected", func(t *testing.T) {
		var s StreamAssert
		if err := yaml.Unmarshal([]byte("json: []\n"), &s); err == nil {
			t.Fatal("empty json list should be rejected")
		}
	})

	t.Run("unknown key inside a check is rejected", func(t *testing.T) {
		var s StreamAssert
		if err := yaml.Unmarshal([]byte("json:\n  path: $.a\n  equalss: 1\n"), &s); err == nil {
			t.Fatal("a typo'd key inside a json check should be rejected (strict)")
		}
	})
}

// TestJSONChecks_MarshalRoundTrip proves a single check round-trips to the
// mapping form (not a one-item list) and a multi-check list round-trips to a
// sequence, so a loaded json/yaml matcher re-marshals to the shape it was
// written in.
func TestJSONChecks_MarshalRoundTrip(t *testing.T) {
	t.Parallel()
	for _, src := range []string{
		"json:\n  path: $.a\n  equals: 1\n",
		"json:\n  - path: $.a\n    equals: 1\n  - path: $.b\n    equals: 2\n",
	} {
		var s StreamAssert
		if err := yaml.Unmarshal([]byte(src), &s); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		out, err := yaml.Marshal(&s)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var back StreamAssert
		if err := yaml.Unmarshal(out, &back); err != nil {
			t.Fatalf("re-unmarshal %q: %v", out, err)
		}
		if !reflect.DeepEqual(s.JSON, back.JSON) {
			t.Errorf("round-trip mismatch: %+v -> %s -> %+v", s.JSON, out, back.JSON)
		}
	}
}

func TestExitCode_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	t.Run("scalar int", func(t *testing.T) {
		t.Parallel()
		var e ExitCode
		if err := yaml.Unmarshal([]byte("0"), &e); err != nil {
			t.Fatal(err)
		}
		if e.Equals == nil || *e.Equals != 0 || e.Not != nil {
			t.Errorf("scalar decode = %+v, want Equals=0", e)
		}
	})
	t.Run("quoted int scalar", func(t *testing.T) {
		t.Parallel()
		// Regression: a YAML-quoted integer (exit_code: "0") must decode as the
		// integer 0, not fail with the "must be an integer, got \"0\"" message.
		for _, in := range []string{`"0"`, `'2'`} {
			var e ExitCode
			if err := yaml.Unmarshal([]byte(in), &e); err != nil {
				t.Fatalf("quoted %s: %v", in, err)
			}
			if e.Equals == nil || e.Not != nil || e.In != nil {
				t.Errorf("quoted %s decode = %+v, want Equals set", in, e)
			}
		}
	})
	t.Run("not map", func(t *testing.T) {
		t.Parallel()
		var e ExitCode
		if err := yaml.Unmarshal([]byte("not: 2"), &e); err != nil {
			t.Fatal(err)
		}
		if e.Not == nil || *e.Not != 2 || e.Equals != nil {
			t.Errorf("not decode = %+v, want Not=2", e)
		}
	})
	t.Run("invalid shape is an error", func(t *testing.T) {
		t.Parallel()
		var e ExitCode
		if err := yaml.Unmarshal([]byte("[1,2]"), &e); err == nil {
			t.Error("expected error decoding a sequence into exit_code")
		}
	})
}
