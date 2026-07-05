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

// FailurePath composes the full relative path for a failed assertion's sidecar
// file: <suite-token>/<scenario-token>/<step-file>. It is deterministic and
// collision-free across suites, scenarios, steps, and parallel runs.
func FailurePath(specPath, scenario string, scenarioIdx, stepIdx int, kind, role, ext string) string {
	return path.Join(SuiteToken(specPath), ScenarioToken(scenario, scenarioIdx), StepFile(stepIdx, kind, role, ext))
}

// ServiceLogPath composes the relative path for a background service's preserved
// combined stdout/stderr log (#51): <suite-token>/<scenario-token>/service-<name>.log.
// It shares the scenario directory with failure sidecars and is collision-free
// across services, scenarios, suites, and parallel runs.
func ServiceLogPath(specPath, scenario string, scenarioIdx int, serviceName string) string {
	return path.Join(SuiteToken(specPath), ScenarioToken(scenario, scenarioIdx), "service-"+Slug(serviceName)+".log")
}
