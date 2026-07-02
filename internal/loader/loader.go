// Package loader reads a atago YAML file and turns it into a validated
// *spec.Spec. Validation happens in layers: YAML parse, then
// schema/semantic checks. Errors carry the file path and, for parse failures,
// the line/column reported by goccy/go-yaml.
package loader

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/nao1215/atago/internal/spec"
)

// Kind classifies why loading failed, so callers can map it to an exit code
// (exit 2 = spec parse error).
type Kind int

const (
	// KindParse is a YAML syntax or decode error.
	KindParse Kind = iota
	// KindValidation is a schema or semantic error in an otherwise-parseable file.
	KindValidation
)

// Error is a loader failure annotated with the source path and kind.
type Error struct {
	Path string
	Kind Kind
	Msg  string
}

func (e *Error) Error() string {
	if e.Path == "" {
		return e.Msg
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Msg)
}

// Load reads and validates the spec file at path.
func Load(path string) (*spec.Spec, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from user-specified spec args
	if err != nil {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: err.Error()}
	}
	return LoadBytes(path, data)
}

// LoadBytes parses and validates spec bytes, labeling errors with path.
func LoadBytes(path string, data []byte) (*spec.Spec, error) {
	var s spec.Spec
	dec := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())
	if err := dec.Decode(&s); err != nil {
		return nil, &Error{Path: path, Kind: KindParse, Msg: formatYAMLError(err)}
	}
	// Record each scenario's authored index before matrix expansion, so every
	// expanded instance can be traced back to its authored source location (#80).
	for i := range s.Scenarios {
		s.Scenarios[i].SourceIndex = i
	}
	// Validate matrix shape on the raw spec, then expand each matrix scenario into
	// concrete instances so the remaining validation and the engine only ever see
	// plain scenarios.
	if errs := validateMatrix(&s); len(errs) > 0 {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	expandMatrix(&s)
	// Expand the top-level defaults into the concrete scenario/step/service model
	// so validation and the engine only ever see fully-resolved scenarios (ADR-0039).
	applyDefaults(&s)
	if errs := validate(&s); len(errs) > 0 {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(errs)}
	}
	return &s, nil
}

// formatYAMLError renders goccy errors with position context when available,
// and appends a did-you-mean hint for misspelled field names.
func formatYAMLError(err error) string {
	var yerr yaml.Error
	if errors.As(err, &yerr) {
		return suggestScalarMatcher(suggestUnknownField(yaml.FormatError(err, false, true)))
	}
	return suggestScalarMatcher(suggestUnknownField(err.Error()))
}

func joinErrors(errs []string) string {
	if len(errs) == 1 {
		return errs[0]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d validation errors:", len(errs))
	for _, e := range errs {
		b.WriteString("\n  - ")
		b.WriteString(e)
	}
	return b.String()
}
