package spec

import (
	"github.com/nao1215/atago/internal/yaml"
)

// Forall makes a scenario a template over GENERATED inputs (#656), the way
// Matrix makes it a template over authored ones: the loader expands it into one
// concrete scenario per generated row, with the row's values seeded as ${name}
// variables. It is the property-based-testing half of the Gleam library metamon
// (https://github.com/nao1215/metamon) — "for every input of this shape, the
// CLI still holds up" — brought to a spec file.
//
// The values are derived from Seed, never from the clock or the OS: the same
// spec produces the same inputs everywhere, every instance carries its own
// values in its name, and a failure names the input that caused it. Asking for
// different inputs is an edit to the spec (Seed), not a re-run.
type Forall struct {
	// Vars binds each variable name to the generator that produces its values.
	Vars map[string]Generator `yaml:"vars"`
	// Runs is how many instances to generate, and an upper bound rather than a
	// promise: rows are deduplicated, so a generator over three choices expands
	// to three scenarios however large Runs is. Zero means the default (10).
	Runs int `yaml:"runs,omitempty"`
	// Seed picks the stream. Two specs that differ only in Seed generate
	// different inputs; the same spec generates the same inputs forever.
	Seed int `yaml:"seed,omitempty"`
}

// Generator describes one variable's values: a kind (with an optional range),
// or a fixed set of choices. The kinds themselves live in internal/generator,
// which owns what each one produces; this type is only what the YAML carries.
type Generator struct {
	// Type names the kind: int, bool, digits, alpha, alphanumeric, ascii,
	// unicode. Empty when OneOf is used instead.
	Type string `yaml:"type,omitempty"`
	// Min and Max bound an integer generator's value or a string generator's
	// length. Nil means the kind's own default.
	Min *int `yaml:"min,omitempty"`
	Max *int `yaml:"max,omitempty"`
	// OneOf is the fixed set of values to pick from, the spelling of a
	// generator whose space is a list rather than a shape.
	OneOf []string `yaml:"one_of,omitempty"`
	// Examples are values always tried, ahead of the generated ones (#661,
	// metamon's with_examples): the input that broke the tool once, which is
	// never a boundary of any generator and which nothing else would reach.
	// They are used as written, so one may sit outside the generator's range.
	Examples []string `yaml:"examples,omitempty"`
}

// UnmarshalYAML decodes a generator as the scalar shorthand (`s: ascii`, the
// kind with its default range) or as a mapping (`{type: ascii, min: 0, max: 32}`
// / `{one_of: [json, yaml]}`). It decodes from the AST node so a shape error
// carries the offending value's [line:col] for the loader's excerpt-and-caret
// formatter, and so unknown keys are rejected here — a custom unmarshaler
// bypasses the loader's strict decode, which is what would otherwise let a
// misspelled `mins:` through in silence.
func (g *Generator) UnmarshalYAML(node *yaml.Node) error {
	fail := failf
	if node.Kind == yaml.ScalarNode {
		return yaml.Decode(node, &g.Type, true)
	}
	if node.Kind != yaml.MappingNode {
		return fail("a forall variable must be a generator name or a mapping ({type: ..., min: ..., max: ...} or {one_of: [...]})")
	}
	for _, pair := range node.Pairs {
		k := pair.Key.Value
		if k == "one_of" || k == "examples" {
			// The values are text, read as written: `one_of: [1.20]` offers
			// the four characters 1.20, not the float they spell.
			if err := g.setList(k, pair.Value, fail); err != nil {
				return err
			}
			continue
		}
		v, err := yaml.Value(pair.Value)
		if err != nil {
			return err
		}
		if err := g.setKey(k, v, fail); err != nil {
			return err
		}
	}
	return nil
}

// setList assigns one_of or examples, which are lists of values.
func (g *Generator) setList(k string, n *yaml.Node, fail func(string, ...any) error) error {
	if n == nil || n.Kind != yaml.SequenceNode {
		return fail("forall generator: %s must be a list of values", k)
	}
	// An authored empty list is kept as an empty (non-nil) slice so the loader
	// can tell "chooses from nothing" — a generator that can produce no value —
	// from a generator that never mentioned the key at all.
	values := make([]string, 0, len(n.Items))
	for _, item := range n.Items {
		if item != nil && item.Kind != yaml.ScalarNode {
			return fail("forall generator: %s must be a list of values", k)
		}
		var s string
		if err := yaml.Decode(item, &s, true); err != nil {
			return err
		}
		values = append(values, s)
	}
	if k == "examples" {
		g.Examples = values
	} else {
		g.OneOf = values
	}
	return nil
}

// setKey assigns one mapping key, reporting the shapes each one accepts.
func (g *Generator) setKey(k string, v any, fail func(string, ...any) error) error {
	switch k {
	case "type":
		s, ok := v.(string)
		if !ok {
			return fail("forall generator: type must be a generator name")
		}
		g.Type = s
	case "min", "max":
		// asInt (pty_send.go) reads every numeric shape a generic decode
		// hands back, so `{type: int, min: -50}` goes through the same reader
		// as `times:`.
		i, ok := asInt(v)
		if !ok {
			return fail("forall generator: %s must be an integer", k)
		}
		if k == "min" {
			g.Min = &i
		} else {
			g.Max = &i
		}
	default:
		return fail("forall generator: unknown key %q (accepted: type, min, max, one_of, examples)", k)
	}
	return nil
}

// MarshalYAML emits the shape the author could have written: the scalar
// shorthand when the generator is a bare kind, the mapping otherwise. The
// default struct marshal would write a `type: ""` for a one_of generator, which
// the unmarshaler then reads as a kind nobody named.
func (g Generator) MarshalYAML() (any, error) {
	if g.Type != "" && g.Min == nil && g.Max == nil && len(g.OneOf) == 0 && len(g.Examples) == 0 {
		return g.Type, nil
	}
	m := make(map[string]any, 4)
	if g.Type != "" {
		m["type"] = g.Type
	}
	if g.Min != nil {
		m["min"] = *g.Min
	}
	if g.Max != nil {
		m["max"] = *g.Max
	}
	if len(g.OneOf) > 0 {
		m["one_of"] = g.OneOf
	}
	if len(g.Examples) > 0 {
		m["examples"] = g.Examples
	}
	return m, nil
}
