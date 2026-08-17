package report

import (
	"path/filepath"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/spec"
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
	// LoadFailures are the spec files the run could not read (#120). They ran no
	// scenario, so they appear in no suite above. Omitted when there are none, so
	// an ordinary run's document is unchanged.
	LoadFailures []jsonLoadFailure `json:"load_failures,omitempty"`
	// SnapshotsUpdated counts the snapshots this run WROTE (--update-snapshots)
	// instead of comparing against. Rewriting the committed expected results is
	// the one passing outcome a consumer has to be able to notice; omitted when
	// zero, so an ordinary verify run's document is unchanged.
	SnapshotsUpdated int `json:"snapshots_updated,omitempty"`
}

// jsonLoadFailure names one unreadable spec and why it could not be read.
type jsonLoadFailure struct {
	SpecPath string `json:"spec_path"`
	Error    string `json:"error"`
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
	// ExpectFail carries the declared known bug for an xfail/xpass scenario
	// (#395), so a dashboard can link the issue rather than re-reading the spec.
	ExpectFail *jsonExpectFail `json:"expect_fail,omitempty"`
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
	Diff  string `json:"diff,omitempty"`
	Hint  string `json:"hint,omitempty"`
	Error string `json:"error,omitempty"`
	// Code is the diagnostic code carried by Error, when it has one. It lets a
	// consumer group failures by cause without matching on prose that is free
	// to be reworded. An assertion failing carries none by design.
	Code      string         `json:"code,omitempty"`
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
func buildJSON(res *engine.SuiteResult, allowXPass bool) jsonReport {
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
			ExpectFail:       jsonExpectFailOf(sc.ExpectFail),
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
		// An XFAIL's checks did fail, but the scenario did not: listing it here
		// would put a green scenario in the failures bucket every consumer reads
		// as "what went wrong", and a dashboard would show a known bug as a
		// regression forever. Its verdict and reason are on the scenario row
		// (status + expect_fail), which is where a consumer looks for it (#395).
		switch sc.Status {
		case engine.StatusXFail:
			// Its checks did fail, but the scenario did not.
		case engine.StatusXPass:
			// The opposite: nothing failed, yet this IS the thing to act on, and
			// it has no failing check to describe itself with — so the bucket
			// gets a synthesized entry rather than staying silent about the one
			// verdict that turned the run red.
			//
			// Under --allow-xpass it did NOT turn the run red, and every
			// failure-level signal has to agree with the exit code: a consumer
			// that reads failures[] would otherwise report a failed test for a
			// green build. The scenario row still carries status "xpass" and the
			// expect_fail block, so nothing is hidden.
			if !allowXPass {
				out.Failures = append(out.Failures, jsonFailure{Scenario: sc.Name, Error: xpassMessage(sc)})
			}
		default:
			out.Failures = append(out.Failures, failuresOf(sc)...)
		}
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
				Code:     firstCode(step.ErrMsg),
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
				Code:     firstCode(step.ErrMsg),
			})
		}
	}
	return fs
}

func failuresOf(sc *engine.ScenarioResult) []jsonFailure {
	var fs []jsonFailure
	// The command a failure is reported under is the one the failing check
	// asserts on, tracked as the steps are walked; the scenario's last command
	// belongs to whatever ran after the failure.
	cmd := ""
	for _, step := range sc.Steps {
		if step.Run != nil {
			cmd = step.Run.Command
		}
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
				Code:     firstCode(step.ErrMsg),
			})
		}
	}
	return fs
}

// jsonExpectFail mirrors a scenario's declared known bug in the JSON report.
type jsonExpectFail struct {
	Reason string `json:"reason"`
	Issue  string `json:"issue,omitempty"`
}

func jsonExpectFailOf(ef *spec.ExpectFail) *jsonExpectFail {
	if ef == nil {
		return nil
	}
	return &jsonExpectFail{Reason: ef.Reason, Issue: ef.Issue}
}

// firstCode returns the diagnostic code a message carries, or "" when it has
// none. Assertion failures deliberately carry none: exit 1 is a spec doing its
// job, so there is nothing for a consumer to branch on beyond the verdict.
func firstCode(msg string) string {
	codes := diag.Codes(msg)
	if len(codes) == 0 {
		return ""
	}
	return codes[0].String()
}
