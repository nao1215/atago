package loader

import (
	"os"
	"path/filepath"

	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/yaml"
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
	root *yaml.Node
}

// LoadWithSource loads and validates the spec at path and also returns a Source
// locator for it. The spec is identical to what Load returns; the extra Source
// exposes authored line/column positions.
func LoadWithSource(path string) (*spec.Spec, *Source, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from user-specified spec args
	if err != nil {
		return nil, nil, &Error{Path: path, Kind: KindValidation, Msg: err.Error()}
	}
	// Discover the directory manifest exactly as Load does. Going straight to
	// LoadBytes skipped it, so a spec under an atago.project.yaml meant one
	// thing to run/explain/doc and another to `atago manifest` — which reported
	// no project_path and none of the configuration the file applies.
	proj, perr := FindProject(filepath.Dir(path))
	if perr != nil {
		return nil, nil, perr
	}
	s, doc, lerr := loadBytesWithProject(path, data, proj)
	if lerr != nil {
		return nil, nil, lerr
	}
	// The locator answers from the document the spec was decoded from.
	return s, &Source{root: doc}, nil
}

// newSource parses data for position lookups. A parse failure yields a Source
// that reports every position as unknown.
func newSource(data []byte) *Source {
	f, err := yaml.Parse(data)
	if err != nil {
		return &Source{}
	}
	return &Source{root: f.First()}
}

// Position is a 1-based source location. A zero Line means "unknown".
type Position struct {
	Line   int
	Column int
}

// at returns the Position of a node; a missing node is the zero Position.
func at(n *yaml.Node) Position {
	if n == nil {
		return Position{}
	}
	return Position{Line: n.Line, Column: n.Col}
}

// SuitePos returns the location of the suite declaration.
func (s *Source) SuitePos() (line, column int) {
	if s == nil {
		return 0, 0
	}
	suite := s.root.Get("suite")
	p := at(suite.Get("name"))
	if p.Line == 0 {
		p = at(suite)
	}
	return p.Line, p.Column
}

// RunnerPos returns the location of a named runner declaration.
func (s *Source) RunnerPos(name string) (line, column int) {
	if s == nil {
		return 0, 0
	}
	p := at(s.root.Get("runners").Get(name))
	return p.Line, p.Column
}

// ScenarioPos returns the location of the authored scenario at authoredIndex
// (its pre-matrix-expansion index). Every instance expanded from one matrix
// template shares this location.
func (s *Source) ScenarioPos(authoredIndex int) (line, column int) {
	if s == nil {
		return 0, 0
	}
	sc := s.root.Get("scenarios").Index(authoredIndex)
	p := at(sc.Get("name"))
	if p.Line == 0 {
		p = at(sc)
	}
	return p.Line, p.Column
}

// StepPos returns the location of step stepIndex within the authored scenario at
// authoredScenarioIndex.
func (s *Source) StepPos(authoredScenarioIndex, stepIndex int) (line, column int) {
	if s == nil {
		return 0, 0
	}
	p := at(s.root.Get("scenarios").Index(authoredScenarioIndex).Get("steps").Index(stepIndex))
	return p.Line, p.Column
}
