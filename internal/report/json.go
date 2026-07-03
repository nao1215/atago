package report

import (
	"path/filepath"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
)

// jsonSchemaVersion is the stable top-level schema version for --report json.
// Bump it only on a breaking change to the document shape (#43).
const jsonSchemaVersion = "1"

// jsonDocument is the single, stable top-level shape emitted by --report json
// (#43). Machine consumers can always read `.suites` as an array regardless of
// whether one suite or many were run, and branch on `schema_version` for future
// format changes.
type jsonDocument struct {
	SchemaVersion string       `json:"schema_version"`
	Suites        []jsonReport `json:"suites"`
}

// A machine-readable report carrying enough failure context
// for an LLM agent to act on. Rendered by Render (FormatJSON) via buildJSON.
type jsonReport struct {
	Suite      string         `json:"suite"`
	SpecPath   string         `json:"spec_path"`
	Status     string         `json:"status"`
	DurationMS int64          `json:"duration_ms"`
	Scenarios  []jsonScenario `json:"scenarios"`
	Failures   []jsonFailure  `json:"failures"`
	// SetupFailures / TeardownFailures list failed suite.setup / suite.teardown
	// steps (#7). Setup failures also error every scenario; teardown failures
	// never change the suite status but incomplete cleanup must stay visible.
	SetupFailures    []jsonFailure `json:"setup_failures,omitempty"`
	TeardownFailures []jsonFailure `json:"teardown_failures,omitempty"`
}

type jsonScenario struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	SkipReason string `json:"skip_reason,omitempty"`
	// Attempts is the execution count under --retry-failed (#29); omitted
	// when the feature was off. Iterations lists each --repeat execution's
	// status. Both additive, so schema_version stays "1".
	Attempts   int      `json:"attempts,omitempty"`
	Iterations []string `json:"iterations,omitempty"`
	// TeardownFailures lists teardown steps that failed or errored. Teardown
	// outcomes never change the scenario's status — the verdict is decided by
	// the steps — but incomplete cleanup of external resources must stay
	// visible to machine consumers.
	TeardownFailures []jsonFailure    `json:"teardown_failures,omitempty"`
	ServiceLogs      []jsonServiceLog `json:"service_logs,omitempty"`
}

// jsonServiceLog references a preserved background-service log artifact (#51).
type jsonServiceLog struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type jsonFailure struct {
	Scenario string `json:"scenario"`
	Step     string `json:"step"`
	Command  string `json:"command,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
	// Diff is the uncolored unified diff for multi-line equals/snapshot
	// failures (#28) — additive, so schema_version stays "1".
	Diff      string         `json:"diff,omitempty"`
	Hint      string         `json:"hint,omitempty"`
	Error     string         `json:"error,omitempty"`
	Artifacts []jsonArtifact `json:"artifacts,omitempty"`
}

// jsonArtifact references a durable sidecar file written for a failed assertion
// when --artifacts-dir is set (#48). Path is stable and relative to the
// artifacts dir root, so CI, editors, and agents can jump directly to it.
type jsonArtifact struct {
	Role string `json:"role"`
	Path string `json:"path"`
}

// buildJSON converts a suite result into the serializable report shape.
func buildJSON(res *engine.SuiteResult) jsonReport {
	out := jsonReport{
		Suite: res.Suite,
		// Forward slashes keep spec_path portable across platforms (Windows uses
		// backslashes), matching the manifest's stable-contract convention.
		SpecPath:   filepath.ToSlash(res.SpecPath),
		Status:     string(res.Status),
		DurationMS: res.Duration.Milliseconds(),
		Scenarios:  make([]jsonScenario, 0, len(res.Scenarios)),
		Failures:   []jsonFailure{},
	}
	out.SetupFailures = suiteStepFailures(res.Suite, res.Setup)
	out.TeardownFailures = suiteStepFailures(res.Suite, res.Teardown)
	for i := range res.Scenarios {
		sc := &res.Scenarios[i]
		js := jsonScenario{
			Name:             sc.Name,
			Status:           string(sc.Status),
			DurationMS:       sc.Duration.Milliseconds(),
			SkipReason:       sc.SkipReason,
			TeardownFailures: teardownFailuresOf(sc),
			ServiceLogs:      serviceLogsOf(sc),
		}
		if sc.Attempts > 1 {
			js.Attempts = sc.Attempts
		}
		for _, it := range sc.Iterations {
			js.Iterations = append(js.Iterations, string(it))
		}
		out.Scenarios = append(out.Scenarios, js)
		out.Failures = append(out.Failures, failuresOf(sc)...)
	}
	return out
}

// artifactsOf maps a failed check's written sidecar files into the stable JSON
// artifact references (#48). It returns nil when no artifacts were written, so
// the `artifacts` field is omitted for runs without --artifacts-dir.
func artifactsOf(cr *assert.CheckResult) []jsonArtifact {
	if len(cr.ArtifactFiles) == 0 {
		return nil
	}
	out := make([]jsonArtifact, 0, len(cr.ArtifactFiles))
	for _, a := range cr.ArtifactFiles {
		out = append(out, jsonArtifact{Role: a.Role, Path: a.Path})
	}
	return out
}

// serviceLogsOf maps a scenario's preserved service-log artifacts into stable
// JSON references (#51). It returns nil when none were written.
func serviceLogsOf(sc *engine.ScenarioResult) []jsonServiceLog {
	if len(sc.ServiceLogs) == 0 {
		return nil
	}
	out := make([]jsonServiceLog, 0, len(sc.ServiceLogs))
	for _, sl := range sc.ServiceLogs {
		out = append(out, jsonServiceLog{Name: sl.Name, Path: sl.Path})
	}
	return out
}

// suiteStepFailures maps failed/errored suite-level steps (#7) into the
// jsonFailure shape, using the suite name as the scenario label.
func suiteStepFailures(suite string, steps []engine.StepResult) []jsonFailure {
	var fs []jsonFailure
	for _, step := range steps {
		for _, ck := range step.Checks {
			if ck == nil || ck.OK {
				continue
			}
			fs = append(fs, jsonFailure{
				Scenario: suite,
				Step:     ck.Desc,
				Expected: ck.Expected,
				Actual:   ck.Actual,
				Diff:     checkDiff(ck),
				Hint:     ck.Hint,
			})
		}
		if step.ErrMsg != "" {
			fs = append(fs, jsonFailure{
				Scenario: suite,
				Step:     stepPhase(step),
				Error:    step.ErrMsg,
			})
		}
	}
	return fs
}

// teardownFailuresOf maps failed/errored teardown steps into the jsonFailure
// shape. It returns nil for a clean (or absent) teardown so the field is
// omitted.
func teardownFailuresOf(sc *engine.ScenarioResult) []jsonFailure {
	var fs []jsonFailure
	for _, step := range sc.Teardown {
		for _, ck := range step.Checks {
			if ck == nil || ck.OK {
				continue
			}
			fs = append(fs, jsonFailure{
				Scenario:  sc.Name,
				Step:      ck.Desc,
				Expected:  ck.Expected,
				Actual:    ck.Actual,
				Diff:      checkDiff(ck),
				Hint:      ck.Hint,
				Artifacts: artifactsOf(ck),
			})
		}
		if step.ErrMsg != "" {
			fs = append(fs, jsonFailure{
				Scenario: sc.Name,
				Step:     stepPhase(step),
				Error:    step.ErrMsg,
			})
		}
	}
	return fs
}

func failuresOf(sc *engine.ScenarioResult) []jsonFailure {
	var fs []jsonFailure
	cmd := lastCommand(sc)
	for _, step := range sc.Steps {
		for _, ck := range step.Checks {
			if ck == nil || ck.OK {
				continue
			}
			fs = append(fs, jsonFailure{
				Scenario:  sc.Name,
				Step:      ck.Desc,
				Command:   cmd,
				Expected:  ck.Expected,
				Actual:    ck.Actual,
				Diff:      checkDiff(ck),
				Hint:      ck.Hint,
				Artifacts: artifactsOf(ck),
			})
		}
		if step.ErrMsg != "" {
			fs = append(fs, jsonFailure{
				Scenario: sc.Name,
				Step:     stepPhase(step),
				Command:  cmd,
				Error:    step.ErrMsg,
			})
		}
	}
	return fs
}
