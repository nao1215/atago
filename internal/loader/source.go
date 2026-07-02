package loader

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/nao1215/atago/internal/spec"
)

// Source resolves stable source locations (line and column) for the declarations
// in a spec file, so tooling can jump from a manifest entry straight to the
// authored YAML (#80). It is built from the same bytes the spec was decoded from
// and answers queries by YAML path against the parsed AST.
//
// A location is 1-based and points at the value token for the queried node (for
// a scenario, the `name:` value; for a step, the step mapping). When a path
// cannot be resolved — e.g. an optional block is absent — the position is zero,
// which callers treat as "unknown" and omit.
type Source struct {
	file *ast.File
}

// LoadWithSource loads and validates the spec at path and also returns a Source
// locator for it. The spec is identical to what Load returns; the extra Source
// exposes authored line/column positions. A parse error for the position AST is
// non-fatal (the spec already decoded), so Source methods simply report unknown.
func LoadWithSource(path string) (*spec.Spec, *Source, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from user-specified spec args
	if err != nil {
		return nil, nil, &Error{Path: path, Kind: KindValidation, Msg: err.Error()}
	}
	s, lerr := LoadBytes(path, data)
	if lerr != nil {
		return nil, nil, lerr
	}
	return s, newSource(data), nil
}

// newSource parses data into an AST for position lookups. A parse failure yields
// a Source that reports every position as unknown rather than an error, since
// the caller already holds a successfully-decoded spec.
func newSource(data []byte) *Source {
	f, err := parser.ParseBytes(data, 0)
	if err != nil {
		return &Source{}
	}
	return &Source{file: f}
}

// Position is a 1-based source location. A zero Line means "unknown".
type Position struct {
	Line   int
	Column int
}

// pos resolves a YAML path (e.g. "$.scenarios[2].name") to a Position. Any error
// or a missing node yields the zero Position.
func (s *Source) pos(path string) Position {
	if s == nil || s.file == nil {
		return Position{}
	}
	p, err := yaml.PathString(path)
	if err != nil {
		return Position{}
	}
	node, err := p.FilterFile(s.file)
	if err != nil || node == nil {
		return Position{}
	}
	tk := node.GetToken()
	if tk == nil {
		return Position{}
	}
	return Position{Line: tk.Position.Line, Column: tk.Position.Column}
}

// SuitePos returns the location of the suite declaration.
func (s *Source) SuitePos() (line, column int) {
	p := s.pos("$.suite.name")
	if p.Line == 0 {
		p = s.pos("$.suite")
	}
	return p.Line, p.Column
}

// RunnerPos returns the location of a named runner declaration.
func (s *Source) RunnerPos(name string) (line, column int) {
	p := s.pos(fmt.Sprintf("$.runners.%s", yamlPathKey(name)))
	return p.Line, p.Column
}

// ScenarioPos returns the location of the authored scenario at authoredIndex
// (its pre-matrix-expansion index). Every instance expanded from one matrix
// template shares this location.
func (s *Source) ScenarioPos(authoredIndex int) (line, column int) {
	p := s.pos(fmt.Sprintf("$.scenarios[%d].name", authoredIndex))
	if p.Line == 0 {
		p = s.pos(fmt.Sprintf("$.scenarios[%d]", authoredIndex))
	}
	return p.Line, p.Column
}

// StepPos returns the location of step stepIndex within the authored scenario at
// authoredScenarioIndex.
func (s *Source) StepPos(authoredScenarioIndex, stepIndex int) (line, column int) {
	p := s.pos(fmt.Sprintf("$.scenarios[%d].steps[%d]", authoredScenarioIndex, stepIndex))
	return p.Line, p.Column
}

// yamlPathKey guards a runner name for use as a YAML path segment. goccy's path
// parser treats a bare identifier segment literally; names with path-significant
// characters (dots, brackets, spaces) cannot be expressed, so they resolve to
// unknown rather than mis-resolving. Returning the name unchanged keeps the
// common identifier case exact.
func yamlPathKey(name string) string {
	return name
}
