package loader

import (
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/nao1215/atago/internal/spec"
)

// verbatimSpec wraps one step in a minimal spec.
func verbatimSpec(step string) string {
	return "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    steps:\n      - " + step + "\n"
}

// TestVerbatim_TextFieldsKeepTheSourceText pins the repair across the shapes
// YAML retypes: a trailing-zero decimal, leading zeros, hex, a case-varied
// boolean. Each was silently rewritten — `contains: 1.20` asserted the
// substring "1.2", which matches output that never contained 1.20 — so each is
// checked where it can do damage: an asserted literal, a command, an exported
// environment value, and fixture content.
func TestVerbatim_TextFieldsKeepTheSourceText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
		want string
		got  func(*spec.Spec) string
	}{
		{
			name: "assert equals keeps a trailing zero",
			src:  verbatimSpec("run: {command: echo}\n      - assert: {stdout: {equals: 1.20}}"),
			want: "1.20",
			got:  func(s *spec.Spec) string { return *s.Scenarios[0].Steps[1].Assert.Stdout.Equals },
		},
		{
			name: "assert contains keeps a trailing zero",
			src:  verbatimSpec("run: {command: echo}\n      - assert: {stdout: {contains: 1.20}}"),
			want: "1.20",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[1].Assert.Stdout.Contains[0] },
		},
		{
			name: "assert contains keeps boolean spelling",
			src:  verbatimSpec("run: {command: echo}\n      - assert: {stdout: {contains: TRUE}}"),
			want: "TRUE",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[1].Assert.Stdout.Contains[0] },
		},
		{
			name: "assert matches keeps a hex-looking pattern",
			src:  verbatimSpec("run: {command: echo}\n      - assert: {stdout: {matches: 0x10}}"),
			want: "0x10",
			got:  func(s *spec.Spec) string { return *s.Scenarios[0].Steps[1].Assert.Stdout.Matches },
		},
		{
			name: "file equals keeps leading zeros",
			src:  verbatimSpec("run: {command: echo}\n      - assert: {file: {path: f, equals: 007}}"),
			want: "007",
			got:  func(s *spec.Spec) string { return *s.Scenarios[0].Steps[1].Assert.File.Equals },
		},
		{
			name: "command keeps its spelling",
			src:  verbatimSpec("run: {command: true}"),
			want: "true",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[0].Run.Command },
		},
		{
			name: "step env value keeps leading zeros",
			src:  verbatimSpec("run: {command: echo, env: {V: 007}}"),
			want: "007",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[0].Run.Env["V"] },
		},
		{
			name: "fixture content keeps a trailing zero",
			src:  verbatimSpec("fixture: {file: f.txt, content: 1.10}"),
			want: "1.10",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[0].Fixture.Content },
		},
		{
			name: "inline stdin keeps its spelling",
			src:  verbatimSpec("run: {command: cat, stdin: 0x10}"),
			want: "0x10",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[0].Run.Stdin.Inline },
		},
		{
			name: "scenario env value keeps a trailing zero",
			src:  "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: a\n    env: {VERSION: 1.20}\n    steps:\n      - run: {command: echo}\n",
			want: "1.20",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Env["VERSION"] },
		},
		{
			name: "suite env value keeps a trailing zero",
			src:  "version: \"1\"\nsuite:\n  name: s\n  env: {VERSION: 1.20}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n",
			want: "1.20",
			got:  func(s *spec.Spec) string { return s.Suite.Env["VERSION"] },
		},
		{
			name: "an env name keeps its case",
			src:  verbatimSpec("run: {command: echo, env: {TRUE: x}}"),
			want: "x",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[0].Run.Env["TRUE"] },
		},
		{
			name: "a matrix row value keeps a trailing zero",
			src:  "version: \"1\"\nsuite:\n  name: s\nscenarios:\n  - name: 'a ${v}'\n    matrix: [{v: 1.20}]\n    steps:\n      - run: {command: 'echo ${v}'}\n",
			want: "a 1.20|1.20",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Name + "|" + s.Scenarios[0].Vars["v"] },
		},
		{
			name: "a pty send keeps leading zeros",
			src:  verbatimSpec("pty: {command: cat, session: [{send: 007}]}"),
			want: "007",
			got:  func(s *spec.Spec) string { return *s.Scenarios[0].Steps[0].PTY.Session[0].Send.Text },
		},
		{
			name: "a defaults run env value keeps leading zeros",
			src:  "version: \"1\"\nsuite:\n  name: s\ndefaults:\n  run: {env: {V: 007}}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n",
			want: "007",
			got:  func(s *spec.Spec) string { return s.Scenarios[0].Steps[0].Run.Env["V"] },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if got := tt.got(s); got != tt.want {
				t.Errorf("value = %q, want %q (the text the spec author wrote)", got, tt.want)
			}
		})
	}
}

// TestVerbatim_TypedFieldsKeepTheirTyping is the other half of the contract.
// The repair is driven by the target's Go type, so a field that is genuinely a
// number, a boolean, or a JSON value must be untouched: `exit_code: 007` is the
// number 7, and `json: {equals: true}` compares against the JSON boolean, which
// is documented not to equal the string "true".
func TestVerbatim_TypedFieldsKeepTheirTyping(t *testing.T) {
	t.Parallel()
	t.Run("exit_code stays a number", func(t *testing.T) {
		t.Parallel()
		s, err := LoadBytes("t.atago.yaml", []byte(verbatimSpec("run: {command: echo}\n      - assert: {exit_code: 007}")))
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		ec := s.Scenarios[0].Steps[1].Assert.ExitCode
		if ec == nil || ec.Equals == nil || *ec.Equals != 7 {
			t.Fatalf("exit_code = %#v, want the number 7", ec)
		}
	})
	t.Run("a json equals keeps its YAML type", func(t *testing.T) {
		t.Parallel()
		s, err := LoadBytes("t.atago.yaml", []byte(verbatimSpec("run: {command: echo}\n      - assert: {stdout: {json: {path: \"$.ok\", equals: true}}}")))
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		got := s.Scenarios[0].Steps[1].Assert.Stdout.JSON[0].Equals
		if b, ok := got.(bool); !ok || !b {
			t.Fatalf("json equals = %#v (%T), want the boolean true", got, got)
		}
	})
	t.Run("a null command is still missing", func(t *testing.T) {
		t.Parallel()
		_, err := LoadBytes("t.atago.yaml", []byte(verbatimSpec("run: {command: null}")))
		if err == nil || !strings.Contains(err.Error(), "command is required") {
			t.Fatalf("error = %v, want command is required (null is unset, not the text \"null\")", err)
		}
	})
	t.Run("a quoted value is unchanged", func(t *testing.T) {
		t.Parallel()
		s, err := LoadBytes("t.atago.yaml", []byte(verbatimSpec("run: {command: echo}\n      - assert: {stdout: {equals: \"1.20\"}}")))
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if got := *s.Scenarios[0].Steps[1].Assert.Stdout.Equals; got != "1.20" {
			t.Errorf("value = %q, want 1.20", got)
		}
	})
	t.Run("a binary payload still decodes", func(t *testing.T) {
		t.Parallel()
		// `atago record` writes !!binary when a captured stream is not valid
		// UTF-8; an explicit tag says the author asked for something other than
		// the source text, so the repair must not touch it.
		s, err := LoadBytes("t.atago.yaml", []byte(verbatimSpec("run: {command: echo}\n      - assert: {stdout: {contains: !!binary \"MTgwMw==\"}}")))
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if got := s.Scenarios[0].Steps[1].Assert.Stdout.Contains[0]; got != "1803" {
			t.Errorf("value = %q, want the decoded 1803", got)
		}
	})
}

// TestVerbatim_EveryCustomDecodeIsCovered walks the spec model for types that
// decode themselves. Such a type ignores its yaml tags, so the repair walk has
// no way to reach the text it holds unless customDecodes says how — and a type
// missing from there fails silently, which is the failure mode this whole
// change exists to remove. Every entry must also say why it is shaped the way
// it is, so "nobody looked at this one" cannot pass as a decision.
func TestVerbatim_EveryCustomDecodeIsCovered(t *testing.T) {
	t.Parallel()
	nodeUnmarshaler := reflect.TypeFor[interface{ UnmarshalYAML(ast.Node) error }]()
	funcUnmarshaler := reflect.TypeFor[interface {
		UnmarshalYAML(func(any) error) error
	}]()

	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if seen[t] {
			return
		}
		seen[t] = true
		switch t.Kind() {
		case reflect.Struct:
			for i := 0; i < t.NumField(); i++ {
				walk(t.Field(i).Type)
			}
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Pointer:
			walk(t.Elem())
		default:
		}
	}
	walk(reflect.TypeFor[spec.Spec]())

	for typ := range seen {
		if !reflect.PointerTo(typ).Implements(nodeUnmarshaler) && !reflect.PointerTo(typ).Implements(funcUnmarshaler) {
			continue
		}
		entry, ok := customDecodes[typ]
		if !ok {
			t.Errorf("%s decodes itself but has no customDecodes entry; the verbatim walk cannot reach its text, so a retyped scalar there would be restored nowhere", typ)
			continue
		}
		if entry.why == "" {
			t.Errorf("customDecodes[%s] states no reason; record how this type's text is reached", typ)
		}
	}
	for typ := range customDecodes {
		if !seen[typ] {
			t.Errorf("customDecodes lists %s, which the spec model no longer reaches; remove the entry", typ)
		}
	}
}
