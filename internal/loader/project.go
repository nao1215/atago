package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/nao1215/atago/internal/spec"
)

// ProjectFileName is the directory-level manifest a spec tree may carry (#392).
const ProjectFileName = "atago.project.yaml"

// Project is the directory-level manifest: configuration that belongs to a
// TREE of specs rather than to one file (#392).
//
// It exists because the concerns it carries are per-directory by nature, and
// every downstream suite was expressing them in a shell wrapper around `atago
// run` instead. sqly and omokage export a throwaway HOME and the XDG variables
// in bash — even though `sandbox_home: true` already does exactly that per step
// — because the only way to apply it to 89 spec files was to repeat `defaults:`
// in all 89, where file number 90 silently forgets. Five other repos export an
// env var naming a committed fixture directory, identical for every file in the
// tree.
//
// A `suite:` block cannot say "once per directory". This can, and a spec file
// with no manifest above it behaves exactly as before.
type Project struct {
	// Env is the weakest environment layer: host < project < suite < scenario <
	// step.
	Env map[string]string `yaml:"env,omitempty"`
	// Defaults merge BENEATH each spec file's own `defaults:`, with the same
	// rules (#39) — an explicit value always wins.
	Defaults *spec.Defaults `yaml:"defaults,omitempty"`
	// FixturesDir points at a committed fixture tree, resolved relative to the
	// manifest's own directory and exposed to every spec under it as
	// ${fixtures} (#394).
	FixturesDir string `yaml:"fixtures_dir,omitempty"`

	// Path is where the manifest was found. It is never authored; explain, doc,
	// list, and manifest print it, because configuration that applies to a file
	// without appearing in it has to be visible somewhere.
	Path string `yaml:"-"`
	// ResolvedFixturesDir is FixturesDir made absolute against the manifest's
	// directory. Never authored.
	ResolvedFixturesDir string `yaml:"-"`
}

// FindProject returns the nearest manifest at or above dir, or nil when there
// is none.
//
// Nearest-ancestor is the rule every tool with a project file uses (.git,
// package.json, .editorconfig), and it is what makes a manifest work for a
// nested tree: `atago run ./e2e` and `atago run ./e2e/cli/one.atago.yaml` have
// to resolve the same configuration, or a developer re-running one failing spec
// gets a different environment from CI. The walk is bounded by the filesystem
// root, and every command that reads a spec prints which manifest applied, so
// "where did this come from" is answerable without guessing.
func FindProject(dir string) (*Project, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for {
		candidate := filepath.Join(abs, ProjectFileName)
		if st, serr := os.Stat(candidate); serr == nil && !st.IsDir() {
			return LoadProject(candidate)
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return nil, nil
		}
		abs = parent
	}
}

// LoadProject reads and validates one manifest.
func LoadProject(path string) (*Project, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is the manifest atago itself located
	if err != nil {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: err.Error()}
	}
	var p Project
	if derr := yaml.UnmarshalWithOptions(data, &p, yaml.Strict()); derr != nil {
		return nil, &Error{Path: path, Kind: KindParse, Msg: yaml.FormatError(derr, false, true)}
	}
	p.Path = path

	var msgs []string
	add := func(format string, args ...any) { msgs = append(msgs, fmt.Sprintf(format, args...)) }
	if p.Defaults != nil {
		validateDefaults(add, p.Defaults)
	}
	if p.FixturesDir != "" {
		resolved := p.FixturesDir
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(filepath.Dir(path), resolved)
		}
		// Checked at load time on purpose: a typo here would otherwise surface as
		// N confusing per-scenario failures about a missing file, one per spec in
		// the tree, instead of one message about the manifest.
		st, serr := os.Stat(resolved)
		switch {
		case serr != nil:
			add("fixtures_dir %q does not exist (resolved to %s)", p.FixturesDir, resolved)
		case !st.IsDir():
			add("fixtures_dir %q is not a directory (resolved to %s)", p.FixturesDir, resolved)
		default:
			p.ResolvedFixturesDir = resolved
		}
	}
	if len(msgs) > 0 {
		return nil, &Error{Path: path, Kind: KindValidation, Msg: joinErrors(msgs)}
	}
	return &p, nil
}

// applyProject layers a manifest beneath a loaded spec: the spec's own values
// always win. It runs BEFORE applyDefaults, so the merged defaults go through
// exactly the same expansion the authored ones do and the two cannot drift.
func applyProject(s *spec.Spec, p *Project) {
	if p == nil {
		return
	}
	s.ProjectPath = p.Path
	s.FixturesDir = p.ResolvedFixturesDir
	if len(p.Env) > 0 {
		s.Suite.Env = mergeStringMap(p.Env, s.Suite.Env)
	}
	if p.Defaults != nil {
		// The manifest's defaults are the weaker layer. Rather than a second
		// merge algorithm for Defaults (which would drift from applyDefaults'
		// rules the first time a field is added), the file's own block is
		// applied first and the manifest's is applied to whatever is still
		// unset — the same "an explicit value always wins" rule, one pass later.
		applyDefaults(s)
		s.Defaults = p.Defaults
	}
}
