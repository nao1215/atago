// Package artifact writes deterministic sidecar files for failed assertions and
// other durable review evidence (#48). It shapes collision-free relative paths
// from the suite, scenario, and step so CI, editors, and agents can jump
// directly to the captured payloads, and it is reused by the service-log and
// image-diff artifact features that build on the same directory mechanism.
package artifact

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Dir is a root directory into which durable review artifacts are written. A nil
// *Dir writes nothing, letting callers keep the artifact path optional.
type Dir struct {
	root string
}

// NewDir returns a Dir rooted at root. Callers gate creation on the presence of
// the --artifacts-dir flag, so an unset flag yields a nil *Dir.
func NewDir(root string) *Dir { return &Dir{root: filepath.Clean(root)} }

// Write stores content at relPath (a slash-separated path relative to the root)
// and returns the same relPath for embedding in reports. Parent directories are
// created as needed. Write is safe for concurrent use across distinct relPaths,
// which is how parallel scenarios each write their own sidecars.
func (d *Dir) Write(relPath string, content []byte) (string, error) {
	full := filepath.Join(d.root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return "", fmt.Errorf("create artifact dir: %w", err)
	}
	if err := os.WriteFile(full, content, 0o600); err != nil {
		return "", fmt.Errorf("write artifact %q: %w", relPath, err)
	}
	return relPath, nil
}

// Slug lowercases s and collapses every run of characters outside [a-z0-9] into a
// single '-'. An empty or all-separator input yields the stable token "artifact"
// so a filename is never empty.
func Slug(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "artifact"
	}
	return out
}

// SuiteToken returns a stable, collision-free directory token for a suite. It
// combines a readable slug of the spec path's base name with a short content
// hash of the full path, so two suites whose base names slug identically still
// land in distinct directories.
func SuiteToken(specPath string) string {
	base := filepath.Base(specPath)
	base = strings.TrimSuffix(base, ".yaml")
	base = strings.TrimSuffix(base, ".yml")
	base = strings.TrimSuffix(base, ".atago")
	sum := sha256.Sum256([]byte(specPath))
	return Slug(base) + "-" + hex.EncodeToString(sum[:])[:8]
}

// ScenarioToken returns a stable token for a scenario within a suite. The index
// disambiguates scenarios that share a name (e.g. matrix rows).
func ScenarioToken(name string, index int) string {
	return fmt.Sprintf("%s-%d", Slug(name), index)
}

// StepFile returns the sidecar filename for a step's failure payload, shaped as
// step-<NN>-<kind>.<role>.<ext>, e.g. step-02-stdout.actual.txt.
func StepFile(stepIndex int, kind, role, ext string) string {
	return fmt.Sprintf("step-%02d-%s.%s.%s", stepIndex, Slug(kind), Slug(role), ext)
}

// Scenario identifies one execution of one scenario: everything an artifact path
// needs apart from the step itself. Holding the fields together is what keeps a
// caller from composing a path for the wrong execution — the engine writes
// failure payloads, service logs, and mock request logs from four different
// places for the same run.
type Scenario struct {
	// SpecPath is the spec file the scenario was loaded from.
	SpecPath string
	// Name is the scenario name; Index disambiguates scenarios that share one
	// (matrix rows).
	Name  string
	Index int
	// Attempt counts executions of this same scenario: 1 (or 0) for a plain run,
	// then 2, 3, ... for each further --repeat iteration or --retry-failed
	// attempt. Every attempt after the first gets its own subdirectory, because a
	// report describes ONE attempt inline and points at that attempt's payloads:
	// sharing a path lets a later attempt overwrite evidence the report still
	// references, leaving the file disagreeing with the diff printed beside it.
	Attempt int
	// Phase names which block of the scenario produced the artifact. The empty
	// value is the scenario's own numbered steps and keeps the path a plain run
	// has always written; PhaseTeardown gets a segment of its own for the same
	// reason Attempt does — a teardown step's index counts from zero again, so a
	// teardown failure at index N would otherwise overwrite the evidence for the
	// scenario step at index N, under a filename still naming step N.
	Phase string
}

// PhaseTeardown is the Scenario.Phase value for a scenario's teardown block.
// Suite-level lifecycle blocks need no phase: they already run under their own
// pseudo-scenario directory.
const PhaseTeardown = "teardown"

// Dir is the directory every artifact of this execution lives in:
// <suite-token>/<scenario-token>, plus an attempt-<N> segment from the second
// attempt on and a phase segment for anything but the scenario's own steps. The
// first attempt of the steps phase keeps the plain path, so a run with neither
// --repeat nor --retry-failed nor a teardown failure writes exactly where it
// always has.
func (s Scenario) Dir() string {
	dir := path.Join(SuiteToken(s.SpecPath), ScenarioToken(s.Name, s.Index))
	if s.Attempt > 1 {
		dir = path.Join(dir, fmt.Sprintf("attempt-%d", s.Attempt))
	}
	if s.Phase != "" {
		dir = path.Join(dir, Slug(s.Phase))
	}
	return dir
}

// FailurePath composes the relative path for a failed assertion's sidecar file:
// <dir>/<step-file>. It is deterministic and collision-free across suites,
// scenarios, attempts, phases, steps, and parallel runs.
func (s Scenario) FailurePath(stepIdx int, kind, role, ext string) string {
	return path.Join(s.Dir(), StepFile(stepIdx, kind, role, ext))
}

// ServiceLogPath composes the relative path for a background service's preserved
// combined stdout/stderr log (#51): <dir>/service-<name>.log, sharing the
// scenario directory with failure sidecars.
func (s Scenario) ServiceLogPath(serviceName string) string {
	return path.Join(s.Dir(), "service-"+Slug(serviceName)+".log")
}

// MockLogPath composes the relative path for a mock server's preserved request
// log: <dir>/mock-<name>.log. The mock- prefix keeps it distinct from a service
// log even when a mock and a service share a name.
func (s Scenario) MockLogPath(mockName string) string {
	return path.Join(s.Dir(), "mock-"+Slug(mockName)+".log")
}
