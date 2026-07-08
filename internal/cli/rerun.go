package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// rerunStateDir and rerunStateFile locate the deterministic local state file
// that records the previous run's failing scenarios for `--rerun-failed` (#64).
// The state lives under the current directory so a red-green loop is scoped to
// the project being worked on, and is meant to be git-ignored.
const (
	rerunStateDir  = ".atago"
	rerunStateFile = "last-failed.json"
)

// rerunState is the on-disk shape of the last-failed record. It is a small,
// explicit, versioned JSON document so the format is documented and stable.
type rerunState struct {
	SchemaVersion string        `json:"schema_version"`
	Failed        []failedEntry `json:"failed"`
}

// failedEntry identifies one scenario that failed on the last run.
type failedEntry struct {
	SpecPath string `json:"spec_path"`
	Scenario string `json:"scenario"`
}

// RerunStateSchemaVersion versions the last-failed state file.
const RerunStateSchemaVersion = "1"

func rerunStatePath() string {
	return filepath.Join(rerunStateDir, rerunStateFile)
}

// loadRerunState reads the last-failed state file. A missing file is not an
// error: it yields an empty state so `--rerun-failed` degrades to "nothing to
// rerun" rather than failing.
func loadRerunState() (rerunState, error) {
	var st rerunState
	data, err := os.ReadFile(rerunStatePath())
	if err != nil {
		if os.IsNotExist(err) {
			return rerunState{SchemaVersion: RerunStateSchemaVersion}, nil
		}
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, err
	}
	// Reject an unknown schema version rather than interpreting a future format
	// under v1 assumptions: a later version may move or rename fields, and reading
	// it as v1 would silently drop or misread the recorded failures — turning a
	// still-red project green. An empty version predates the field and is treated
	// as v1 for backward compatibility.
	if st.SchemaVersion != "" && st.SchemaVersion != RerunStateSchemaVersion {
		return st, fmt.Errorf("%s has unsupported schema_version %q (this atago understands %q); delete it to reset the rerun state",
			rerunStatePath(), st.SchemaVersion, RerunStateSchemaVersion)
	}
	return st, nil
}

// selectSet turns the recorded failures into the engine identity set consumed by
// Engine.Select. It returns nil when nothing was recorded.
func (st rerunState) selectSet() map[string]bool {
	if len(st.Failed) == 0 {
		return nil
	}
	set := make(map[string]bool, len(st.Failed))
	for _, e := range st.Failed {
		set[engine.ScenarioID(e.SpecPath, e.Scenario)] = true
	}
	return set
}

// specPaths returns the deduplicated, lexically-sorted spec paths that hold a
// recorded failure, so `--rerun-failed` can load only the specs it needs.
func (st rerunState) specPaths() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range st.Failed {
		if !seen[e.SpecPath] {
			seen[e.SpecPath] = true
			out = append(out, e.SpecPath)
		}
	}
	sort.Strings(out)
	return out
}

// saveRerunState writes the failing scenarios recorded from a run to the state
// file. When there are no failures the file is removed so a fully-green run
// clears the red-green loop. Writes are best-effort: an unwritable directory
// (e.g. a read-only CI checkout) is not a run failure, so the error is returned
// for the caller to surface as a warning rather than to fail the run.
func saveRerunState(failed []failedEntry) error {
	if len(failed) == 0 {
		err := os.Remove(rerunStatePath())
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	// Deterministic ordering so the file is stable across runs.
	sort.Slice(failed, func(i, j int) bool {
		if failed[i].SpecPath != failed[j].SpecPath {
			return failed[i].SpecPath < failed[j].SpecPath
		}
		return failed[i].Scenario < failed[j].Scenario
	})
	st := rerunState{SchemaVersion: RerunStateSchemaVersion, Failed: failed}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(rerunStateDir, 0o750); err != nil {
		return err
	}
	return os.WriteFile(rerunStatePath(), data, 0o600)
}

// absClean returns a canonical form of p so --rerun-failed matches a spec
// regardless of how its path is spelled (relative vs absolute) between the
// recording run and the rerun; comparing raw strings would miss equivalent
// paths. Symlinks in the prefix are resolved too: os.Getwd (used by Abs) returns
// the symlink-resolved directory, so on a platform whose temp dir is a symlink
// (macOS /var -> /private/var) a relative recording and an explicit /var/...
// target would otherwise canonicalize differently. EvalSymlinks needs the path
// to exist; fall back to the absolute (then lexical) form when it does not.
func absClean(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	if resolved, rerr := filepath.EvalSymlinks(abs); rerr == nil {
		return resolved
	}
	return abs
}

// portableSpecPath returns a spelling of p to store in the ledger that survives
// the project being moved or checked out at a different absolute path (a CI
// cache restored under a different directory). It prefers a forward-slashed,
// cwd-relative form; a --rerun-failed absolutizes its spec paths in memory to
// match across spellings, and persisting that absolute form made the next rerun
// after a move find nothing and silently greenlight still-failing work. A path
// outside cwd (or one that cannot be relativized) keeps its canonical absolute
// form, which is the best portable spelling available for it.
func portableSpecPath(p string) string {
	abs := absClean(p)
	cwd, err := os.Getwd()
	if err != nil {
		return abs
	}
	rel, rerr := filepath.Rel(absClean(cwd), abs)
	if rerr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return abs
	}
	return filepath.ToSlash(rel)
}

// portableEntry rewrites an entry's spec path to its portable spelling.
func portableEntry(e failedEntry) failedEntry {
	e.SpecPath = portableSpecPath(e.SpecPath)
	return e
}

// portableEntries maps portableEntry over a slice.
func portableEntries(in []failedEntry) []failedEntry {
	out := make([]failedEntry, 0, len(in))
	for _, e := range in {
		out = append(out, portableEntry(e))
	}
	return out
}

// intersectPaths returns the members of paths that also appear in keep,
// preserving the order of paths. Both sides are absolutized before comparison so
// an equivalent-but-differently-spelled path (relative vs absolute) still
// matches; without this, a rerun target that names the same spec by a different
// spelling would find "nothing" and silently greenlight despite real failures.
func intersectPaths(paths, keep []string) []string {
	want := make(map[string]bool, len(keep))
	for _, k := range keep {
		want[absClean(k)] = true
	}
	var out []string
	for _, p := range paths {
		if want[absClean(p)] {
			out = append(out, p)
		}
	}
	return out
}

// canonicalScenarioID is engine.ScenarioID with the spec path canonicalized, so
// a recorded failure matches an executed scenario whether their paths are spelled
// relative or absolute (or reach the same file through a symlinked temp dir).
// Comparing raw ScenarioID strings would miss equivalent-but-differently-spelled
// paths and let a preserved failure be dropped or double-counted.
func canonicalScenarioID(specPath, scenario string) string {
	return engine.ScenarioID(absClean(specPath), scenario)
}

// dedupeEntries removes duplicate (spec_path, scenario) entries, keeping
// first-seen order, so a failure preserved from the prior ledger and the same
// failure freshly recorded this run never both land in the file.
func dedupeEntries(in []failedEntry) []failedEntry {
	seen := make(map[string]bool, len(in))
	var out []failedEntry
	for _, e := range in {
		k := e.SpecPath + "\x00" + e.Scenario
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, e)
	}
	return out
}

// collectFailures extracts the failing scenario identities from a set of suite
// results and their spec paths, in a deterministic order.
func collectFailures(results []*engine.SuiteResult) []failedEntry {
	var failed []failedEntry
	for _, res := range results {
		if res == nil {
			continue
		}
		for _, sc := range res.Scenarios {
			if sc.Status == engine.StatusFailed || sc.Status == engine.StatusError {
				failed = append(failed, failedEntry{SpecPath: res.SpecPath, Scenario: sc.Name})
			}
		}
	}
	return failed
}

// applyRerunSelection narrows a --rerun-failed run to the scenarios the ledger
// recorded as failing: it absolutizes the recorded spec paths AND the run
// targets so a spec matches regardless of how its path is spelled between the
// recording run and the rerun (relative vs absolute) — without this, a rerun
// addressed by an equivalent-but-different spelling finds "nothing" and
// silently greenlights while the failures are still real. It intersects the
// recorded spec paths with the collected targets so the usual path semantics
// still apply, and installs an identity selector so only the recorded
// scenarios execute. done reports that the run should end immediately with
// exitNow: an unreadable ledger is a config error, and nothing-to-rerun is
// reported and treated as success.
func applyRerunSelection(label string, stderr io.Writer, paths []string, eng *engine.Engine) (narrowed []string, exitNow int, done bool) {
	state, lerr := loadRerunState()
	if lerr != nil {
		fmt.Fprintf(stderr, label+": cannot read %s: %v\n", rerunStatePath(), lerr)
		return nil, ExitConfig, true
	}
	for i := range state.Failed {
		state.Failed[i].SpecPath = absClean(state.Failed[i].SpecPath)
	}
	for i := range paths {
		paths[i] = absClean(paths[i])
	}
	sel := state.selectSet()
	if len(sel) == 0 {
		fmt.Fprintln(stderr, label+": no previously failed scenarios recorded; nothing to rerun")
		return nil, ExitOK, true
	}
	paths = intersectPaths(paths, state.specPaths())
	if len(paths) == 0 {
		fmt.Fprintln(stderr, label+": no previously failed scenarios under the given targets")
		return nil, ExitOK, true
	}
	eng.Select = sel
	return paths, 0, false
}

// updateRerunLedger persists the last-failed ledger after a run. The ledger
// reflects what this run decided about the scenarios it EXECUTED — a failure is
// recorded, a pass is cleared — while PRESERVING every recorded failure the run
// did not execute. A run that touches only a subset of scenarios (a narrowed
// --rerun-failed, a --filter/--tag/--skip-tag selection, or simply running a
// different or smaller set of specs) must not drop still-failing work elsewhere:
// overwriting the ledger with only what ran would forget it and could greenlight
// a later --rerun-failed while real failures remain. So a fully-green run of the
// SAME specs clears the file, but a green run of an unrelated spec keeps the
// other spec's recorded failures. Entries are stored with portable
// (cwd-relative) spec paths so the ledger survives a project move — a
// --rerun-failed absolutizes paths in memory to match across spellings, and
// persisting that absolute form would break the next rerun after the directory
// moves. Writing is best-effort — a read-only checkout must not fail the run —
// so a write error is a warning, not a fatal exit; an UNREADABLE ledger is
// never overwritten, so recorded failures a newer atago wrote survive a plain
// run by an older one.
func updateRerunLedger(label string, stderr io.Writer, results []*engine.SuiteResult, ranScenarios int) {
	prior, perr := loadRerunState()
	if perr != nil {
		fmt.Fprintf(stderr, label+": cannot read %s; leaving it untouched: %v\n", rerunStatePath(), perr)
		return
	}
	executed := make(map[string]bool, ranScenarios)
	for _, r := range results {
		for _, sc := range r.Scenarios {
			executed[canonicalScenarioID(r.SpecPath, sc.Name)] = true
		}
	}
	var preserved []failedEntry
	for _, e := range prior.Failed {
		if !executed[canonicalScenarioID(e.SpecPath, e.Scenario)] {
			preserved = append(preserved, e)
		}
	}
	entries := dedupeEntries(portableEntries(append(collectFailures(results), preserved...)))
	if err := saveRerunState(entries); err != nil {
		fmt.Fprintf(stderr, label+": could not update %s: %v\n", rerunStatePath(), err)
	}
}
