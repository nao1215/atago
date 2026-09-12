package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/diag"
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
//
// An XPASS is recorded: it is what turned an otherwise green run red, so
// leaving it out meant a red run wrote no ledger at all and the follow-up
// `--rerun-failed` reported "nothing to rerun" and exited 0 — a green answer to
// a run that failed. Re-running it reproduces the XPASS, which is the outcome
// that keeps saying "promote this spec" until someone does. --allow-xpass makes
// the same run green, and a green run must leave nothing behind.
//
// A flaky scenario stays out even though it also fails the run: its re-run most
// likely passes, and clearing the ledger is not the same as fixing the
// instability. What is worth re-running and what turns a run red are the same
// question for every status but that one.
func collectFailures(results []*engine.SuiteResult, allowXPass bool) []failedEntry {
	var failed []failedEntry
	for _, res := range results {
		if res == nil {
			continue
		}
		for _, sc := range res.Scenarios {
			if sc.Status == engine.StatusFlaky {
				continue
			}
			if engine.FailsRun(sc.Status, true, allowXPass) {
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
		fmt.Fprintf(stderr, "%s: %s\n", label, diag.StateUnreadable.Annotate(fmt.Sprintf("cannot read %s: %v", rerunStatePath(), lerr)))
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
// warnUnmatchedRerunEntries reports recorded failures that this --rerun-failed
// run did not execute. The ledger keeps those entries, but silence is the wrong
// answer: a rerun that shows "1 scenario" where two were recorded reads as
// "the other one is fixed". The all-gone case is handled by the caller (it
// verified nothing and exits non-zero); a partial mismatch stays a warning
// because the entry survives for the next rerun.
//
// The two reasons an entry goes unexecuted get separate sentences, because they
// send the reader to different places. A scenario the run LOOKED for and did not
// find was renamed or deleted while still broken — a spec change to go read. An
// entry under a spec this run never targeted is not evidence of any spec change
// at all: the user narrowed the run, and the recorded failure is simply out of
// scope. Blaming a rename there sends them hunting for an edit nobody made.
func warnUnmatchedRerunEntries(label string, stderr io.Writer, results []*engine.SuiteResult, unreached, targets []string) {
	prior, err := loadRerunState()
	if err != nil || len(prior.Failed) == 0 {
		return
	}
	// Selected, not executed: a recorded failure --fail-fast never got to does
	// still exist in the specs, so blaming a rename would send the reader after
	// a spec change that never happened. Same for a whole spec the run never
	// loaded, which contributes no scenarios to compare against at all.
	selected := selectedScenarioIDs(results)
	skip := make(map[string]bool, len(unreached))
	for _, p := range unreached {
		skip[absClean(p)] = true
	}
	inScope := make(map[string]bool, len(targets))
	for _, p := range targets {
		inScope[absClean(p)] = true
	}
	var unmatched, outOfScope []failedEntry
	for _, e := range prior.Failed {
		if selected[canonicalScenarioID(e.SpecPath, e.Scenario)] || skip[absClean(e.SpecPath)] {
			continue
		}
		if !inScope[absClean(e.SpecPath)] {
			outOfScope = append(outOfScope, e)
			continue
		}
		unmatched = append(unmatched, e)
	}
	if len(unmatched) > 0 {
		fmt.Fprintf(stderr, "%s: warning: %s did not match the current specs (renamed or removed?) and %s not rerun: %s; kept in %s for the next --rerun-failed\n",
			label, pluralScenarios(len(unmatched)), wasWere(len(unmatched)), namedRerunEntries(unmatched), rerunStatePath())
	}
	if len(outOfScope) > 0 {
		fmt.Fprintf(stderr, "%s: warning: %s outside this run's targets %s not rerun: %s; kept in %s for the next --rerun-failed\n",
			label, pluralScenarios(len(outOfScope)), wasWere(len(outOfScope)), namedRerunEntries(outOfScope), rerunStatePath())
	}
}

// namedRerunEntries renders recorded failures as "spec / scenario" pairs in a
// deterministic order, naming at most maxNamedRerunEntries of them. A ledger
// accumulates entries for every spec ever run from this directory, so an
// unbounded list buries the run's own result under thousands of characters of
// warning; the file named in the same sentence holds the full set.
func namedRerunEntries(entries []failedEntry) string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, fmt.Sprintf("%s / %s", e.SpecPath, e.Scenario))
	}
	sort.Strings(names)
	if len(names) <= maxNamedRerunEntries {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s, and %d more", strings.Join(names[:maxNamedRerunEntries], ", "), len(names)-maxNamedRerunEntries)
}

// wasWere picks the verb that agrees with a recorded-failure count.
func wasWere(n int) string {
	if n == 1 {
		return "was"
	}
	return "were"
}

// maxNamedRerunEntries bounds how many recorded failures a ledger warning names
// before summarizing the rest by count.
const maxNamedRerunEntries = 5

// pluralScenarios renders a recorded-failure count with the right noun.
func pluralScenarios(n int) string {
	if n == 1 {
		return "1 recorded failing scenario"
	}
	return fmt.Sprintf("%d recorded failing scenarios", n)
}

// executedScenarioIDs is the set of scenarios a run actually executed, and so
// the set whose ledger entries this run is entitled to rewrite.
//
// A selected scenario that never ran (NotRun) is deliberately left out:
// counting its skip as a verdict cleared a failure the ledger had already
// recorded, and the next --rerun-failed went green with the scenario still
// broken.
func executedScenarioIDs(results []*engine.SuiteResult) map[string]bool {
	executed := map[string]bool{}
	for _, r := range results {
		for _, sc := range r.Scenarios {
			if sc.NotRun {
				continue
			}
			executed[canonicalScenarioID(r.SpecPath, sc.Name)] = true
		}
	}
	return executed
}

// selectedScenarioIDs is the set of scenarios the run's selection matched,
// whether or not each one got to run. It answers "does this scenario still
// exist under these targets?", which is what the renamed-or-removed warning
// asks, not executedScenarioIDs' "did this run decide a verdict for it?".
func selectedScenarioIDs(results []*engine.SuiteResult) map[string]bool {
	selected := map[string]bool{}
	for _, r := range results {
		for _, sc := range r.Scenarios {
			selected[canonicalScenarioID(r.SpecPath, sc.Name)] = true
		}
	}
	return selected
}

func updateRerunLedger(label string, stderr io.Writer, results []*engine.SuiteResult, allowXPass bool) {
	prior, perr := loadRerunState()
	if perr != nil {
		fmt.Fprintf(stderr, label+": cannot read %s; leaving it untouched: %v\n", rerunStatePath(), perr)
		return
	}
	executed := executedScenarioIDs(results)
	var preserved []failedEntry
	for _, e := range prior.Failed {
		if !executed[canonicalScenarioID(e.SpecPath, e.Scenario)] {
			preserved = append(preserved, e)
		}
	}
	entries := dedupeEntries(portableEntries(append(collectFailures(results, allowXPass), preserved...)))
	if err := saveRerunState(entries); err != nil {
		fmt.Fprintf(stderr, label+": could not update %s: %v\n", rerunStatePath(), err)
	}
}
