package spec

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// Stdin is a run step's standard-input source (#18). It accepts either the
// historical scalar string (inline text) or a mapping with exactly one of
// `file` (a workdir-relative, ${name}-expanded, workdir-confined path whose
// bytes are fed to the child) or `base64` (binary bytes, validated at load
// time; no ${name} expansion, mirroring fixture.base64). The loader enforces
// the one-of rule.
type Stdin struct {
	// Inline is the scalar form, fed to stdin verbatim.
	Inline string
	// File names a workdir-relative file whose bytes become stdin.
	File string
	// Base64 carries binary stdin as base64.
	Base64 string

	// mapped records that the author used the mapping form, so the validator
	// can reject an empty mapping ({}), which is otherwise indistinguishable
	// from "no stdin".
	mapped bool
}

// IsZero reports whether no stdin was authored at all.
func (s Stdin) IsZero() bool {
	return s.Inline == "" && s.File == "" && s.Base64 == "" && !s.mapped
}

// IsMapping reports whether the author used the {file/base64} mapping form.
func (s Stdin) IsMapping() bool { return s.mapped }

// UnmarshalYAML decodes stdin as a scalar string or a {file}/{base64} mapping.
// It decodes from the AST node so escapes like "\x1b" in the inline form stay
// resolved by goccy's parser AND every shape error carries the offending
// value's [line:col] for the loader's excerpt-and-caret formatter. Unknown
// mapping keys are rejected here (a custom unmarshaler bypasses the loader's
// strict-decode), with the accepted shapes spelled out.
func (s *Stdin) UnmarshalYAML(node ast.Node) error {
	fail := func(format string, args ...any) error {
		return &yaml.SyntaxError{Message: fmt.Sprintf(format, args...), Token: node.GetToken()}
	}
	var one string
	if err := yaml.NodeToValue(node, &one); err == nil {
		s.Inline = one
		return nil
	}
	var raw map[string]any
	if err := yaml.NodeToValue(node, &raw); err != nil {
		return fail("stdin must be a string, {file: path}, or {base64: data}")
	}
	s.mapped = true
	for k, v := range raw {
		str, ok := v.(string)
		if !ok {
			return fail("stdin.%s must be a string", k)
		}
		switch k {
		case "file":
			s.File = str
		case "base64":
			s.Base64 = str
		default:
			return fail("stdin: unknown key %q (accepted: file, base64)", k)
		}
	}
	return nil
}

// MarshalYAML emits the scalar inline form or the {file}/{base64} mapping the
// author used, mirroring UnmarshalYAML so a loaded stdin round-trips. The
// default struct marshal would instead write an `inline:` key the unmarshaler
// rejects.
func (s Stdin) MarshalYAML() (any, error) {
	if s.mapped {
		m := make(map[string]string, 1)
		if s.File != "" {
			m["file"] = s.File
		}
		if s.Base64 != "" {
			m["base64"] = s.Base64
		}
		return m, nil
	}
	return s.Inline, nil
}
