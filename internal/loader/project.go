package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"

	"github.com/nao1215/atago/internal/diag"
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
	// Subject is the binary under test: how to build it, and what the specs
	// call it (#393).
	Subject *Subject `yaml:"subject,omitempty"`
	// Profiles are named build variations selected with `atago run --profile
	// NAME` — a coverage-instrumented build, a race build, a different
	// toolchain.
	Profiles map[string]Profile `yaml:"profiles,omitempty"`

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
	// Absolute from here on: fixtures_dir and a build cwd resolve against the
	// manifest's directory, and a relative manifest path would silently make
	// them resolve against the process's working directory instead — a
	// difference nobody would see until a run from another directory failed.
	if abs, aerr := filepath.Abs(path); aerr == nil {
		path = abs
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is the manifest atago itself located
	if err != nil {
		return nil, &Error{Path: path, Kind: KindValidation, Code: diag.SpecUnreadable, Msg: err.Error()}
	}
	var p Project
	if derr := yaml.UnmarshalWithOptions(data, &p, yaml.Strict()); derr != nil {
		msg := yaml.FormatError(derr, false, true)
		return nil, &Error{Path: path, Kind: KindParse, Code: classifyYAMLError(data, msg), Msg: msg}
	}
	p.Path = path

	var msgs []string
	add := func(code diag.Code, format string, args ...any) {
		msgs = append(msgs, code.Annotate(fmt.Sprintf(format, args...)))
	}
	if p.Defaults != nil {
		validateDefaults(add, p.Defaults)
	}
	validateSubject(add, p.Subject, p.Profiles)
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
			add(diag.PathNotUsable, "fixtures_dir %q does not exist (resolved to %s)", p.FixturesDir, resolved)
		case !st.IsDir():
			add(diag.PathNotUsable, "fixtures_dir %q is not a directory (resolved to %s)", p.FixturesDir, resolved)
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
	// Record the subject build on the spec: it is a command that runs on the
	// host before any scenario, and leaving it inside the loader is why nothing
	// described or flagged it.
	if p.Subject != nil && p.Subject.Build != nil {
		s.Subject = &spec.Subject{
			Name:    p.Subject.Name,
			Command: p.Subject.Build.Command,
			Shell:   p.Subject.Build.Shell != nil && *p.Subject.Build.Shell,
		}
	}
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

// Subject declares the binary under test (#393).
//
// Every downstream repo builds it in a shell wrapper before calling atago:
// `go build -o "$TMP/bin/gup" .`, `cargo build --release --locked`, and so on,
// then prepends the artifact directory to PATH so a bare `gup` in a spec
// resolves to the freshly built binary. The build is just a command, which is
// what keeps this language-neutral: nothing here knows about Go.
//
// The cookbook's older recipe — a suite.setup step building into ${suitedir} —
// does not scale to it: a suite is one FILE, so sqly's 89 specs would rebuild
// 89 times, the binary never lands on PATH, and there is no per-OS artifact
// name or profile switch. That is why all nine repos kept the bash.
type Subject struct {
	// Name is what specs call the binary. On Windows the artifact resolves with
	// a .exe suffix, which is the per-OS branch jose hand-rolls in bash today.
	Name string `yaml:"name"`
	// Build is the command that produces the artifact.
	Build *Build `yaml:"build"`
	// Artifact is where the build writes, relative to a run-scoped scratch
	// directory. ${artifact} in the build command expands to its absolute path.
	Artifact string `yaml:"artifact"`
}

// Build is one build command (#393).
type Build struct {
	// Command runs through the same argv parser a run step uses; set Shell for
	// a command that needs shell syntax.
	Command string `yaml:"command"`
	Shell   *bool  `yaml:"shell,omitempty"`
	// Cwd is where the build runs, relative to the manifest's directory. It
	// exists because the manifest usually sits under e2e/ while the module root
	// is above it — the `cd "$REPO_ROOT"` every wrapper script performs.
	Cwd string `yaml:"cwd,omitempty"`
	// Timeout bounds the build (Go duration). A build that hangs must fail the
	// run rather than the CI job's own wall clock.
	Timeout string `yaml:"timeout,omitempty"`
}

// Profile is a named build variation (#393).
//
// It is what keeps coverage out of atago: instrumenting a binary is an
// alternate build command plus some environment, and that shape is the same in
// every language — `go build -cover` with GOCOVERDIR, `RUSTFLAGS=-C
// instrument-coverage` with LLVM_PROFILE_FILE. Merging the raw profiles
// afterwards is a toolchain job (`go tool covdata`, `llvm-profdata`) and stays
// in a script, where it belongs.
type Profile struct {
	// Build replaces the subject's build command wholly when the profile is
	// selected. Whole-command replacement rather than a merge: a half-merged
	// build command is unreadable and unpredictable.
	Build *Build `yaml:"build,omitempty"`
	// Env is layered into every spec under the manifest while the profile is
	// active, the same way the manifest's own env is.
	Env map[string]string `yaml:"env,omitempty"`
}

// validateSubject checks the subject/profiles block (#393).
func validateSubject(add addFunc, sub *Subject, profiles map[string]Profile) {
	for name, prof := range profiles {
		if name == "" {
			add(diag.EmptyValue, "profiles has an empty name")
		}
		if prof.Build == nil && len(prof.Env) == 0 {
			add(diag.ChooseAtLeastOne, "profiles.%s sets neither build nor env, so selecting it would change nothing", name)
		}
		if prof.Build != nil {
			validateBuild(add, "profiles."+name+".build", prof.Build)
		}
		if prof.Build != nil && sub == nil {
			add(diag.KeyNeedsAnother, "profiles.%s declares a build, but there is no subject: to build", name)
		}
	}
	if sub == nil {
		return
	}
	if sub.Name == "" {
		add(diag.RequiredKey, "subject.name is required: it is what specs call the binary")
	}
	if sub.Artifact == "" {
		add(diag.RequiredKey, "subject.artifact is required: it is where the build writes, and what ${artifact} expands to")
	} else if filepath.IsAbs(sub.Artifact) {
		add(diag.AbsolutePath, "subject.artifact %q must be relative: it is written into a run-scoped scratch directory, not into the repository", sub.Artifact)
	}
	if sub.Build == nil {
		add(diag.RequiredKey, "subject.build is required: atago has to know how to produce the binary")
		return
	}
	validateBuild(add, "subject.build", sub.Build)
}

func validateBuild(add addFunc, where string, b *Build) {
	if b.Command == "" {
		add(diag.RequiredKey, "%s.command is required", where)
	}
	if b.Timeout != "" {
		// Two different mistakes, and the reader fixes them differently: text
		// that is not a duration at all, and a duration that parsed but leaves
		// the build no time to run.
		if d, err := time.ParseDuration(b.Timeout); err != nil {
			add(diag.BadDuration, "%s.timeout must be a positive Go duration (got %q)", where, b.Timeout)
		} else if d <= 0 {
			add(diag.NonPositiveValue, "%s.timeout must be a positive Go duration (got %q)", where, b.Timeout)
		}
	}
}
