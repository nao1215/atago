// Package manifest builds a stable, machine-readable summary of what one or more
// specs declare, without executing them (#49). Unlike the presentation-oriented
// `explain` and `doc` outputs, a manifest is a deterministic JSON document with a
// schema_version, intended for review tooling, indexing, editor integration, and
// automation. It describes the authored spec — suites, scenarios, tags, runners,
// services, matrix bindings, variable references, generated artifacts, and
// security-relevant operations — and never embeds scripting or workflow
// semantics.
package manifest

import (
	"path/filepath"
	"sort"

	"github.com/nao1215/atago/internal/spec"
)

// SchemaVersion is the stable top-level version of the manifest document. It is
// bumped only on a breaking change to the document shape.
const SchemaVersion = "1"

// Input pairs a loaded spec with the path it was read from. Source, when
// non-nil, supplies authored line/column positions so the manifest can carry
// source locations for editor/review-tool integration (#80).
type Input struct {
	Spec   *spec.Spec
	Path   string
	Source SourceLocator
}

// SourceLocator resolves authored source positions for a spec's declarations. It
// is satisfied by loader.Source; the manifest package stays decoupled from the
// loader by depending only on this small interface. Every method returns a
// 1-based (line, column); a zero line means the position is unknown and the
// field is omitted from the manifest.
type SourceLocator interface {
	SuitePos() (line, column int)
	RunnerPos(name string) (line, column int)
	ScenarioPos(authoredIndex int) (line, column int)
	StepPos(authoredScenarioIndex, stepIndex int) (line, column int)
}

// Source is a resolved authored source location, emitted only when known.
type Source struct {
	Line   int `json:"line"`
	Column int `json:"column,omitempty"`
}

// sourceFrom builds a *Source from a (line, column) pair, returning nil when the
// position is unknown so callers can leave the field unset.
func sourceFrom(line, column int) *Source {
	if line <= 0 {
		return nil
	}
	return &Source{Line: line, Column: column}
}

// Document is the top-level manifest shape.
type Document struct {
	SchemaVersion string `json:"schema_version"`
	Specs         []Spec `json:"specs"`
}

// Spec is one spec file's declarative summary.
type Spec struct {
	SpecPath string `json:"spec_path"`
	Suite    string `json:"suite"`
	// SuiteTimeout mirrors suite.timeout, the suite-level default step timeout
	// (#17); empty when the built-in default applies.
	SuiteTimeout string `json:"suite_timeout,omitempty"`
	// Source is the authored location of the suite declaration (#80). Omitted
	// when positions are unavailable.
	Source  *Source  `json:"source,omitempty"`
	Secrets []string `json:"secrets,omitempty"`
	Network Network  `json:"network"`
	Runners []Runner `json:"runners,omitempty"`
	// SuiteEnv, SuiteSetup, and SuiteTeardown mirror the suite-level lifecycle
	// (#7): env exported to every scenario, steps run once before any scenario
	// (setup may contain service steps starting suite-wide peers), and steps
	// that always run after the last scenario.
	SuiteEnv      []string   `json:"suite_env,omitempty"`
	SuiteSetup    []Step     `json:"suite_setup,omitempty"`
	SuiteTeardown []Step     `json:"suite_teardown,omitempty"`
	Scenarios     []Scenario `json:"scenarios"`
}

// Network summarizes the spec's network policy.
type Network struct {
	// Policy is "allowlist" when hosts are restricted, else "unrestricted".
	Policy string   `json:"policy"`
	Allow  []string `json:"allow,omitempty"`
}

// Runner is a named runner declaration, reduced to its identifying fields.
type Runner struct {
	Name    string `json:"name"`
	Type    string `json:"type,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
	Target  string `json:"target,omitempty"`
	Host    string `json:"host,omitempty"`
	// HasDSN reports that a db runner declares a dsn, without exposing the dsn
	// itself (it may embed credentials).
	HasDSN bool `json:"has_dsn,omitempty"`
	// Browser-runner configuration (omitted for other runner types). Headless is a
	// pointer so an unset field is elided while an explicit `headless: false` is
	// preserved.
	Headless    *bool    `json:"headless,omitempty"`
	ExecPath    string   `json:"exec_path,omitempty"`
	BrowserArgs []string `json:"browser_args,omitempty"`
	// Source is the authored location of this runner declaration (#80).
	Source *Source `json:"source,omitempty"`
}

// Scenario is one scenario's declarative summary.
type Scenario struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
	// Vars holds the bound matrix row for a scenario expanded from a matrix,
	// so tooling can see which parameterized instance this is.
	Vars     map[string]string `json:"vars,omitempty"`
	Only     *Condition        `json:"only,omitempty"`
	Skip     *Condition        `json:"skip,omitempty"`
	Services []Service         `json:"services,omitempty"`
	Steps    []Step            `json:"steps"`
	// Teardown lists steps that always run after Steps (pass, fail, error, or
	// interrupt), sharing the scenario's variable store. Their failures never
	// change the scenario's verdict.
	Teardown  []Step   `json:"teardown,omitempty"`
	Variables []string `json:"variables,omitempty"`
	Generates []string `json:"generates,omitempty"`
	Security  []string `json:"security,omitempty"`
	// Source is the authored location of this scenario (#80). Every instance
	// expanded from one matrix template shares its template's location.
	Source *Source `json:"source,omitempty"`
}

// Condition mirrors a skip/only gate.
type Condition struct {
	OS      string `json:"os,omitempty"`
	Env     string `json:"env,omitempty"`
	Command string `json:"command,omitempty"`
}

// Service summarizes a background service.
type Service struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Shell   bool   `json:"shell,omitempty"`
	// ClearEnv / PassEnv mirror the hermetic-environment controls (#16).
	ClearEnv bool     `json:"clear_env,omitempty"`
	PassEnv  []string `json:"pass_env,omitempty"`
	// Ready is the readiness signal kind ("file", "port", "log", "delay") or
	// empty when the steps start as soon as the process spawns.
	Ready string `json:"ready,omitempty"`
	Store string `json:"store,omitempty"`
}

// Step is one step, reduced to its kind and salient declarative fields.
type Step struct {
	Index  int    `json:"index"`
	Kind   string `json:"kind"`
	Action string `json:"action,omitempty"` // one-line human-oriented summary

	// Kind-specific fields, all optional.
	Command string `json:"command,omitempty"`
	Shell   bool   `json:"shell,omitempty"`
	// ClearEnv / PassEnv mirror the hermetic-environment controls on run, pty,
	// and suite service steps (#16).
	ClearEnv bool     `json:"clear_env,omitempty"`
	PassEnv  []string `json:"pass_env,omitempty"`
	Runner   string   `json:"runner,omitempty"`
	Method   string   `json:"method,omitempty"`
	Path     string   `json:"path,omitempty"`
	SQL      string   `json:"sql,omitempty"`
	File     string   `json:"file,omitempty"`
	Target   string   `json:"target,omitempty"` // assert target / store name
	Retry    *Retry   `json:"retry,omitempty"`
	// Source is the authored location of this step (#80).
	Source *Source `json:"source,omitempty"`
}

// Retry mirrors a run or http step's retry policy.
type Retry struct {
	Times    int    `json:"times"`
	Interval string `json:"interval,omitempty"`
}

// Build assembles a deterministic manifest document from the given specs, in the
// order they were supplied. Every map is emitted in sorted key order so repeated
// runs over the same specs produce byte-identical output.
func Build(inputs []Input) Document {
	doc := Document{SchemaVersion: SchemaVersion, Specs: make([]Spec, 0, len(inputs))}
	for _, in := range inputs {
		doc.Specs = append(doc.Specs, buildSpec(in))
	}
	return doc
}

func buildSpec(in Input) Spec {
	s := in.Spec
	out := Spec{
		// Forward slashes keep spec_path stable across platforms (Windows
		// filepath.Clean uses backslashes), so the manifest is a portable contract.
		SpecPath:     filepath.ToSlash(in.Path),
		Suite:        s.Suite.Name,
		SuiteTimeout: s.Suite.Timeout,
		Secrets:      append([]string(nil), s.Secrets...),
		Network:      buildNetwork(s),
		Runners:      buildRunners(s.Runners, in.Source),
		Scenarios:    make([]Scenario, 0, len(s.Scenarios)),
	}
	if in.Source != nil {
		out.Source = sourceFrom(in.Source.SuitePos())
	}
	// Suite lifecycle (#7). Env is emitted as sorted key names only — values
	// may embed credentials, same rule as a db runner's dsn.
	if len(s.Suite.Env) > 0 {
		keys := make([]string, 0, len(s.Suite.Env))
		for k := range s.Suite.Env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out.SuiteEnv = keys
	}
	suiteVars := map[string]bool{}
	for i := range s.Suite.Setup {
		out.SuiteSetup = append(out.SuiteSetup, buildStep(i, &s.Suite.Setup[i], suiteVars))
	}
	for i := range s.Suite.Teardown {
		out.SuiteTeardown = append(out.SuiteTeardown, buildStep(i, &s.Suite.Teardown[i], suiteVars))
	}
	for i := range s.Scenarios {
		out.Scenarios = append(out.Scenarios, buildScenario(&s.Scenarios[i], in.Source))
	}
	return out
}

func buildNetwork(s *spec.Spec) Network {
	if s.Permissions != nil && s.Permissions.Network != nil && len(s.Permissions.Network.Allow) > 0 {
		return Network{Policy: "allowlist", Allow: append([]string(nil), s.Permissions.Network.Allow...)}
	}
	return Network{Policy: "unrestricted"}
}

func buildRunners(runners map[string]spec.Runner, src SourceLocator) []Runner {
	if len(runners) == 0 {
		return nil
	}
	names := make([]string, 0, len(runners))
	for name := range runners {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]Runner, 0, len(names))
	for _, name := range names {
		r := runners[name]
		mr := Runner{
			Name:        name,
			Type:        r.Type,
			BaseURL:     r.BaseURL,
			Target:      r.Target,
			Host:        r.Host,
			HasDSN:      r.DSN != "",
			Headless:    r.Headless,
			ExecPath:    r.ExecPath,
			BrowserArgs: r.BrowserArgs,
		}
		if src != nil {
			mr.Source = sourceFrom(src.RunnerPos(name))
		}
		out = append(out, mr)
	}
	return out
}
