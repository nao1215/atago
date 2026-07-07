package report

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/runner"
)

func sampleResults() []*engine.SuiteResult {
	return []*engine.SuiteResult{
		{
			Suite:    "s1",
			Status:   engine.StatusFailed,
			Duration: 5 * time.Millisecond,
			Scenarios: []engine.ScenarioResult{
				{Name: "p", Status: engine.StatusPassed, Duration: time.Millisecond,
					Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
				{Name: "f", Status: engine.StatusFailed, Duration: time.Millisecond, Steps: []engine.StepResult{
					{Kind: "assert", Checks: []*assert.CheckResult{{
						OK: false, Desc: "assert stdout contains \"x\"", Expected: "x", Actual: "y", Hint: "missing x"}}},
				}},
				{Name: "s", Status: engine.StatusSkipped, SkipReason: "only on os=plan9"},
			},
		},
	}
}

func TestRender_Console(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatConsole, sampleResults()); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"FAILED:", "Expected:", "Actual:", "Hint:", "1 passed, 1 failed", "skipped"} {
		if !strings.Contains(out, want) {
			t.Errorf("console output missing %q\n%s", want, out)
		}
	}
}

// allPassingResults is a single all-green suite, used to prove the load-failure
// path (#120) flips an otherwise-PASSED summary to FAILED.
func allPassingResults() []*engine.SuiteResult {
	return []*engine.SuiteResult{{
		Suite:    "good",
		Status:   engine.StatusPassed,
		Duration: time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{Name: "ok", Status: engine.StatusPassed, Duration: time.Millisecond,
				Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
		},
	}}
}

// TestRender_Console_LoadFailures proves that when spec files failed to load
// (#120) the console summary reads FAILED — not a misleading PASSED that
// contradicts the non-zero exit code — and surfaces the dropped-file count so
// the totals are not silently short. With zero load failures a green run is
// unchanged.
func TestRender_Console_LoadFailures(t *testing.T) {
	t.Parallel()

	t.Run("mixed valid+invalid reads FAILED and counts the drop", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatConsole, allPassingResults(), WithLoadFailures(1)); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if !strings.Contains(out, "FAILED") {
			t.Errorf("summary should read FAILED when a spec failed to load:\n%s", out)
		}
		if strings.Contains(out, "PASSED") {
			t.Errorf("summary must not read PASSED when a spec failed to load:\n%s", out)
		}
		if !strings.Contains(out, "1 spec failed to load") {
			t.Errorf("summary should report the load failure count:\n%s", out)
		}
	})

	t.Run("plural spec count", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatConsole, allPassingResults(), WithLoadFailures(3)); err != nil {
			t.Fatal(err)
		}
		if out := b.String(); !strings.Contains(out, "3 specs failed to load") {
			t.Errorf("summary should pluralize the load failure count:\n%s", out)
		}
	})

	t.Run("no load failures is unchanged (green stays PASSED)", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatConsole, allPassingResults()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if !strings.Contains(out, "PASSED") {
			t.Errorf("all-green run should read PASSED:\n%s", out)
		}
		if strings.Contains(out, "failed to load") {
			t.Errorf("clean run must not mention load failures:\n%s", out)
		}
	})
}

// Regression for issue #19: a service/setup-phase error (StepResult with an
// empty Kind) must render as "service setup", not the misleading "step 0 ()".
func TestRender_Console_ServiceSetupError(t *testing.T) {
	t.Parallel()
	results := []*engine.SuiteResult{{
		Suite:  "s1",
		Status: engine.StatusError,
		Scenarios: []engine.ScenarioResult{
			{Name: "svc", Status: engine.StatusError, Steps: []engine.StepResult{
				{Index: 0, Kind: "", ErrMsg: `service "api" not ready: timed out after 5s waiting for readiness`},
			}},
		},
	}}
	var b bytes.Buffer
	if err := Render(&b, FormatConsole, results); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if strings.Contains(out, "step 0 ()") {
		t.Errorf("console rendered the misleading 'step 0 ()' label:\n%s", out)
	}
	if !strings.Contains(out, "service setup") {
		t.Errorf("console should label a setup-phase error 'service setup':\n%s", out)
	}
}

// setupErrorResults builds a suite whose scenario failed in the setup phase
// (before any numbered step), the shape a service-readiness or workdir-creation
// failure produces.
func setupErrorResults(msg string) []*engine.SuiteResult {
	return []*engine.SuiteResult{{
		Suite:  "s1",
		Status: engine.StatusError,
		Scenarios: []engine.ScenarioResult{
			{Name: "svc", Status: engine.StatusError, Steps: []engine.StepResult{
				{Index: 0, Kind: "", Setup: true, ErrMsg: msg},
			}},
		},
	}}
}

// TestRender_SetupErrorContext_JSON proves a setup-phase error carries an
// explicit, non-blank phase in the machine-readable JSON `step` field.
func TestRender_SetupErrorContext_JSON(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatJSON, setupErrorResults(`service "api" not ready: timed out`)); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, `"step": "service setup"`) {
		t.Errorf("JSON step field should be labeled 'service setup', got:\n%s", out)
	}
	if strings.Contains(out, `"step": ""`) {
		t.Errorf("JSON step field must not be blank for a setup-phase error:\n%s", out)
	}
}

// TestRender_SetupErrorContext_JUnitTAP proves neither JUnit nor TAP emits the
// blank "Error in  step" phrasing for a setup-phase error.
func TestRender_SetupErrorContext_JUnitTAP(t *testing.T) {
	t.Parallel()
	for _, f := range []Format{FormatJUnit, FormatTAP} {
		var b bytes.Buffer
		if err := Render(&b, f, setupErrorResults(`could not create workdir: no space left`)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		out := b.String()
		if strings.Contains(out, "Error in  step") {
			t.Errorf("%s emitted the blank 'Error in  step' phrasing:\n%s", f, out)
		}
		if !strings.Contains(out, "service setup") {
			t.Errorf("%s should attribute the error to 'service setup':\n%s", f, out)
		}
	}
}

// TestRender_SuiteSetupErrorNotMislabeled is a regression: a suite.setup step
// failure (the shape engine.go builds — Setup:true, empty Kind, and a message
// the engine already prefixes with "suite setup failed …") must not be rendered
// under the "service setup" label, which is for background-service readiness.
// The label was overloaded across service readiness, suite setup, and workdir
// creation, so a suite.setup failure printed the misleading "service setup:".
func TestRender_SuiteSetupErrorNotMislabeled(t *testing.T) {
	t.Parallel()
	results := []*engine.SuiteResult{{
		Suite:  "s1",
		Status: engine.StatusError,
		Scenarios: []engine.ScenarioResult{
			{Name: "svc", Status: engine.StatusError, Steps: []engine.StepResult{
				{Index: 0, Kind: "", Setup: true, ErrMsg: "suite setup failed at step 2 (run): boom"},
			}},
		},
	}}
	for _, f := range []Format{FormatConsole, FormatJSON, FormatJUnit, FormatTAP} {
		var b bytes.Buffer
		if err := Render(&b, f, results); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		out := b.String()
		if strings.Contains(out, "service setup") {
			t.Errorf("%s mislabeled a suite-setup failure as 'service setup':\n%s", f, out)
		}
		if !strings.Contains(out, "suite setup") {
			t.Errorf("%s dropped the suite-setup attribution:\n%s", f, out)
		}
	}
}

func TestRender_JUnitValidXML(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatJUnit, sampleResults()); err != nil {
		t.Fatal(err)
	}
	var root junitTestsuites
	if err := xml.Unmarshal(b.Bytes(), &root); err != nil {
		t.Fatalf("JUnit output is not valid XML: %v\n%s", err, b.String())
	}
	if root.Tests != 3 || root.Failures != 1 || root.Skipped != 1 {
		t.Errorf("counts = tests %d failures %d skipped %d, want 3/1/1", root.Tests, root.Failures, root.Skipped)
	}
}

func TestRender_GHA(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatGHA, sampleResults()); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, "::error title=s1 / f::") {
		t.Errorf("missing error annotation:\n%s", out)
	}
	if !strings.Contains(out, "::notice title=atago::") {
		t.Errorf("missing notice summary:\n%s", out)
	}
	// Newlines in the message body must be escaped, not literal.
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if !strings.HasPrefix(line, "::") {
			t.Errorf("annotation line not properly escaped (contains raw newline): %q", line)
		}
	}
}

func TestRender_TAP(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	if err := Render(&b, FormatTAP, sampleResults()); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{
		"TAP version 13",
		"1..3",
		"ok 1 - s1 / p",
		"not ok 2 - s1 / f",
		"ok 3 - s1 / s # SKIP only on os=plan9",
		"  ---",
		"  ...",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("TAP output missing %q\n%s", want, out)
		}
	}
	// The plan count must equal the number of emitted points.
	points := strings.Count(out, "\nok ") + strings.Count(out, "\nnot ok ")
	if points != 3 {
		t.Errorf("emitted %d test points, want 3\n%s", points, out)
	}
}

// flakyResults builds a suite whose single scenario failed once and then passed
// on retry (#29). atago treats a recovered scenario as green: the process exits 0
// (exitForSuite), the console verdict stays PASSED, gha warns instead of erroring,
// and junit records it as a passed testcase. Every report format must agree on
// that verdict, so a flaky scenario is a pass and never a failure.
func flakyResults() []*engine.SuiteResult {
	return []*engine.SuiteResult{{
		Suite:    "s1",
		Status:   engine.StatusPassed,
		Duration: 3 * time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{Name: "flake", Status: engine.StatusFlaky, Attempts: 2, Duration: time.Millisecond, Steps: []engine.StepResult{
				{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}},
			}},
		},
	}}
}

// TestRender_FlakyIsGreenAcrossFormats is a metamorphic parity check: a scenario
// that recovered on retry (StatusFlaky) reads as a pass in every report format,
// matching the exit code and the console verdict. TAP fell through to a `not ok`
// point, contradicting the green verdict the same run reported everywhere else.
func TestRender_FlakyIsGreenAcrossFormats(t *testing.T) {
	t.Parallel()

	t.Run("tap marks a flaky scenario ok, not failed", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatTAP, flakyResults()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if strings.Contains(out, "not ok") {
			t.Errorf("TAP reported a flaky scenario as failed; a recovered scenario is green:\n%s", out)
		}
		if !strings.Contains(out, "ok 1 - s1 / flake") {
			t.Errorf("TAP should mark the flaky scenario as an ok point:\n%s", out)
		}
		// Green for the verdict, but the recovery stays visible, matching gha's
		// warning and junit's flakyFailure element.
		if !strings.Contains(out, "flaky: passed after 2 attempts") {
			t.Errorf("TAP should keep the flaky recovery visible in a diagnostic:\n%s", out)
		}
	})

	t.Run("junit counts a flaky scenario as passed", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatJUnit, flakyResults()); err != nil {
			t.Fatal(err)
		}
		var root junitTestsuites
		if err := xml.Unmarshal(b.Bytes(), &root); err != nil {
			t.Fatalf("JUnit output is not valid XML: %v\n%s", err, b.String())
		}
		if root.Tests != 1 || root.Failures != 0 || root.Errors != 0 || root.Skipped != 0 {
			t.Errorf("counts = tests %d failures %d errors %d skipped %d, want 1/0/0/0",
				root.Tests, root.Failures, root.Errors, root.Skipped)
		}
	})

	t.Run("gha warns instead of erroring on a flaky scenario", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatGHA, flakyResults()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if strings.Contains(out, "::error") {
			t.Errorf("gha emitted an error annotation for a flaky (green) scenario:\n%s", out)
		}
		if !strings.Contains(out, "::warning title=s1 / flake::") {
			t.Errorf("gha should warn on a flaky scenario:\n%s", out)
		}
	})

	t.Run("console verdict stays PASSED with a flaky scenario", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatConsole, flakyResults()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if strings.Contains(out, "FAILED") {
			t.Errorf("console verdict flipped to FAILED for a flaky (green) scenario:\n%s", out)
		}
		if !strings.Contains(out, "1 flaky") {
			t.Errorf("console should surface the flaky count:\n%s", out)
		}
	})
}

// suiteSetupErrorEmpty builds a suite that errored in suite.setup (#7) with no
// scenario rows — the shape the engine produces when suite.setup fails and
// nothing was selected to run (all scenarios filtered out, or an empty scenario
// list). exitForSuite maps StatusError to a non-zero code, and console/JSON both
// show the failure, so no report format may render this as a green suite.
func suiteSetupErrorEmpty() []*engine.SuiteResult {
	return []*engine.SuiteResult{{
		Suite:     "s1",
		Status:    engine.StatusError,
		Scenarios: nil,
		Setup: []engine.StepResult{
			{Index: 0, Kind: "", Setup: true, ErrMsg: `service "api" not ready: timed out after 5s`},
		},
	}}
}

// TestRender_SuiteSetupErrorEmptyIsNotGreen is a metamorphic parity check: a
// suite that errored before producing any scenario row exits non-zero, so every
// report format must surface the failure. Regression: junit reported
// tests=0/errors=0, tap emitted a bare "1..0" plan, gha emitted no error
// annotation, and the console verdict read PASSED — each contradicting the
// non-zero exit code and the JSON "error" status.
func TestRender_SuiteSetupErrorEmptyIsNotGreen(t *testing.T) {
	t.Parallel()

	t.Run("junit records an error testcase, not an empty green suite", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatJUnit, suiteSetupErrorEmpty()); err != nil {
			t.Fatal(err)
		}
		var root junitTestsuites
		if err := xml.Unmarshal(b.Bytes(), &root); err != nil {
			t.Fatalf("JUnit output is not valid XML: %v\n%s", err, b.String())
		}
		if root.Errors == 0 && root.Failures == 0 {
			t.Errorf("junit rendered a green suite (errors=%d failures=%d tests=%d) for a setup-errored suite:\n%s",
				root.Errors, root.Failures, root.Tests, b.String())
		}
		if root.Tests == 0 {
			t.Errorf("junit tests=0 for a setup-errored suite; the failure has no testcase:\n%s", b.String())
		}
		if !strings.Contains(b.String(), "not ready") {
			t.Errorf("junit should carry the setup failure message:\n%s", b.String())
		}
	})

	t.Run("tap emits a not-ok point, not a bare 1..0 plan", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatTAP, suiteSetupErrorEmpty()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if !strings.Contains(out, "not ok") {
			t.Errorf("tap emitted no failing point for a setup-errored suite:\n%s", out)
		}
		if strings.Contains(out, "1..0") {
			t.Errorf("tap emitted an empty 1..0 plan for a setup-errored suite:\n%s", out)
		}
		// The plan count must equal the number of emitted points.
		points := strings.Count(out, "\nok ") + strings.Count(out, "\nnot ok ")
		if !strings.Contains(out, "1.."+itoa(points)) {
			t.Errorf("tap plan does not match the %d emitted points:\n%s", points, out)
		}
	})

	t.Run("gha emits an error annotation", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatGHA, suiteSetupErrorEmpty()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if !strings.Contains(out, "::error") {
			t.Errorf("gha emitted no error annotation for a setup-errored suite:\n%s", out)
		}
	})

	t.Run("console verdict reads FAILED", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatConsole, suiteSetupErrorEmpty()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		// The summary line's verdict word (not the SUITE SETUP FAILED block) must
		// read FAILED, matching the non-zero exit code.
		if !strings.Contains(out, "FAILED  0 scenarios") {
			t.Errorf("console summary verdict read PASSED for a setup-errored suite:\n%s", out)
		}
		if strings.Contains(out, "PASSED  0 scenarios") {
			t.Errorf("console summary verdict must not read PASSED for a setup-errored suite:\n%s", out)
		}
		if !strings.Contains(out, "SUITE SETUP FAILED") {
			t.Errorf("console should show the setup failure block:\n%s", out)
		}
	})

	t.Run("json reports an error status with setup_failures", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		if err := Render(&b, FormatJSON, suiteSetupErrorEmpty()); err != nil {
			t.Fatal(err)
		}
		out := b.String()
		if !strings.Contains(out, `"status": "error"`) {
			t.Errorf("json should report the suite status as error:\n%s", out)
		}
		if !strings.Contains(out, "setup_failures") {
			t.Errorf("json should carry setup_failures:\n%s", out)
		}
	})
}

// itoa avoids the strconv import churn for the small counts this test uses.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestProgress_Markers(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	p := NewProgress(&b) // not a TTY → no color
	p.Scenario(engine.ScenarioResult{Status: engine.StatusPassed})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusSkipped})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusFailed})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusError})
	p.Done()
	if got := b.String(); got != ".sFE\n" {
		t.Errorf("progress markers = %q, want %q", got, ".sFE\n")
	}
}

// suiteWithSetupAndTeardownFailures builds a suite that has real scenario rows
// (so it is NOT the "errored without scenarios" synthetic path) but whose
// suite.setup AND suite.teardown each carry a failed assertion check and an
// errored step. This is the almost-untested writeSuiteSteps / suiteStepFailures
// path (#7): a suite whose bootstrap or cleanup broke.
func suiteWithSetupAndTeardownFailures() *engine.SuiteResult {
	setupCheck := &assert.CheckResult{
		OK:   false,
		Desc: "assert file db.sock exists",
		// Multi-line artifacts trigger the unified-diff rendering branch.
		ArtifactExpected: []byte("ready\nport 5432\n"),
		ArtifactActual:   []byte("ready\nport 5433\n"),
		Hint:             "the service came up on the wrong port",
	}
	teardownCheck := &assert.CheckResult{
		OK:       false,
		Desc:     "assert dir /tmp/scratch does not exist",
		Expected: "absent",
		Actual:   "present",
		Hint:     "scratch dir was left behind",
	}
	return &engine.SuiteResult{
		Suite:    "svc-suite",
		SpecPath: "svc.atago.yaml",
		Status:   engine.StatusFailed,
		Duration: 4 * time.Millisecond,
		Setup: []engine.StepResult{
			{Index: 1, Kind: "assert", Checks: []*assert.CheckResult{setupCheck}},
			{Index: 2, Kind: "run", ErrMsg: "migrate: connection refused"},
		},
		Teardown: []engine.StepResult{
			{Index: 1, Kind: "assert", Checks: []*assert.CheckResult{teardownCheck}},
			{Index: 2, Kind: "run", ErrMsg: "docker rm: no such container"},
		},
		Scenarios: []engine.ScenarioResult{
			{Name: "ok", Status: engine.StatusPassed, Duration: time.Millisecond,
				Steps: []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}}},
		},
	}
}

// TestConsole_SuiteSetupAndTeardownBlocks exercises writeSuiteSteps for both the
// setup and teardown labels, including the unified-diff branch, the
// Expected/Actual branch, the Hint line, and the per-step ErrMsg line.
func TestConsole_SuiteSetupAndTeardownBlocks(t *testing.T) {
	t.Parallel()
	out := render(t, FormatConsole, suiteWithSetupAndTeardownFailures())
	for _, want := range []string{
		"SUITE SETUP FAILED:",
		"svc-suite",
		"assert file db.sock exists",
		"Diff (-expected +actual):",
		"-port 5432", // the differing line rendered as a removal
		"+port 5433", // and the actual as an addition
		"the service came up on the wrong port",
		"step 2 (run): migrate: connection refused",
		"SUITE TEARDOWN FAILED:",
		"assert dir /tmp/scratch does not exist",
		"Expected:",
		"Actual:",
		"scratch dir was left behind",
		"step 2 (run): docker rm: no such container",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("console suite block missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestJSON_SuiteSetupAndTeardownFailures verifies suiteStepFailures maps both
// suite.setup and suite.teardown failures into the machine-readable document,
// with the suite name as the scenario label and the diff embedded.
func TestJSON_SuiteSetupAndTeardownFailures(t *testing.T) {
	t.Parallel()
	var doc jsonDocument
	if err := json.Unmarshal([]byte(render(t, FormatJSON, suiteWithSetupAndTeardownFailures())), &doc); err != nil {
		t.Fatalf("json invalid: %v", err)
	}
	rep := doc.Suites[0]
	if len(rep.SetupFailures) != 2 {
		t.Fatalf("setup_failures = %d, want 2: %+v", len(rep.SetupFailures), rep.SetupFailures)
	}
	if len(rep.TeardownFailures) != 2 {
		t.Fatalf("suite teardown_failures = %d, want 2: %+v", len(rep.TeardownFailures), rep.TeardownFailures)
	}
	// The suite name labels a suite-level failure (there is no scenario for it).
	if rep.SetupFailures[0].Scenario != "svc-suite" {
		t.Errorf("setup failure scenario label = %q, want the suite name", rep.SetupFailures[0].Scenario)
	}
	if rep.SetupFailures[0].Diff == "" {
		t.Errorf("multi-line suite setup failure should carry a diff: %+v", rep.SetupFailures[0])
	}
	// The errored step is captured with its phase label and error text.
	var sawSetupErr, sawTeardownErr bool
	for _, f := range rep.SetupFailures {
		if f.Error == "migrate: connection refused" && f.Step == "run" {
			sawSetupErr = true
		}
	}
	for _, f := range rep.TeardownFailures {
		if f.Error == "docker rm: no such container" && f.Step == "run" {
			sawTeardownErr = true
		}
	}
	if !sawSetupErr || !sawTeardownErr {
		t.Errorf("suite setup/teardown step errors not both mapped: %+v / %+v", rep.SetupFailures, rep.TeardownFailures)
	}
}

// scenarioTeardownFailure builds a passing scenario whose TEARDOWN failed — a
// failed cleanup check (carrying a sidecar artifact) plus an errored teardown
// step. The verdict stays passed; the failure must still surface.
func scenarioTeardownFailure() *engine.SuiteResult {
	return &engine.SuiteResult{
		Suite:    "td",
		SpecPath: "td.atago.yaml",
		Status:   engine.StatusPassed,
		Duration: 2 * time.Millisecond,
		Scenarios: []engine.ScenarioResult{
			{
				Name:     "leaves-clean",
				Status:   engine.StatusPassed,
				Duration: time.Millisecond,
				Steps:    []engine.StepResult{{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}}},
				Teardown: []engine.StepResult{
					{Index: 1, Kind: "assert", Checks: []*assert.CheckResult{{
						OK: false, Desc: "assert file lock removed", Hint: "lockfile survived cleanup",
						ArtifactFiles: []assert.ArtifactFile{{Role: "actual", Path: "artifacts/td/lock.txt"}},
					}}},
					{Index: 2, Kind: "run", ErrMsg: "rm: permission denied"},
				},
				ServiceLogs: []engine.ServiceLog{{Name: "api", Path: "artifacts/td/api.log"}},
			},
		},
	}
}

// TestConsole_ScenarioTeardownFailure exercises writeDetail's teardown block and
// the service-logs footer for a scenario whose verdict stays passed.
func TestConsole_ScenarioTeardownFailure(t *testing.T) {
	t.Parallel()
	out := render(t, FormatConsole, scenarioTeardownFailure())
	for _, want := range []string{
		"TEARDOWN FAILED:",
		"td / leaves-clean",
		"assert file lock removed",
		"lockfile survived cleanup",
		"teardown step 2 (run): rm: permission denied",
		"Service logs:",
		"api: artifacts/td/api.log",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("console teardown block missing %q\n--- got ---\n%s", want, out)
		}
	}
	// A passing-with-failed-teardown scenario keeps the verdict green.
	if strings.Contains(out, "FAILED  ") {
		t.Errorf("a failed teardown must not flip the summary verdict:\n%s", out)
	}
}

// TestJSON_ScenarioTeardownFailures exercises teardownFailuresOf: the failed
// check (with its artifact) and the errored step must both appear under the
// scenario's teardown_failures, and never in the top-level failures[] bucket.
func TestJSON_ScenarioTeardownFailures(t *testing.T) {
	t.Parallel()
	var doc jsonDocument
	if err := json.Unmarshal([]byte(render(t, FormatJSON, scenarioTeardownFailure())), &doc); err != nil {
		t.Fatalf("json invalid: %v", err)
	}
	rep := doc.Suites[0]
	if len(rep.Failures) != 0 {
		t.Errorf("a passing scenario with a failed teardown must not populate failures[]: %+v", rep.Failures)
	}
	sc := rep.Scenarios[0]
	if len(sc.TeardownFailures) != 2 {
		t.Fatalf("scenario teardown_failures = %d, want 2: %+v", len(sc.TeardownFailures), sc.TeardownFailures)
	}
	var sawArtifact, sawErr bool
	for _, f := range sc.TeardownFailures {
		if len(f.Artifacts) == 1 && f.Artifacts[0].Path == "artifacts/td/lock.txt" {
			sawArtifact = true
		}
		if f.Error == "rm: permission denied" {
			sawErr = true
		}
	}
	if !sawArtifact {
		t.Errorf("teardown check artifact not mapped: %+v", sc.TeardownFailures)
	}
	if !sawErr {
		t.Errorf("teardown step error not mapped: %+v", sc.TeardownFailures)
	}
	if len(sc.ServiceLogs) != 1 || sc.ServiceLogs[0].Name != "api" {
		t.Errorf("service logs not mapped: %+v", sc.ServiceLogs)
	}
}

// TestConsole_FailedScenarioArtifacts exercises writeDetail's Artifacts footer
// for a failed check that wrote sidecar files (#48).
func TestConsole_FailedScenarioArtifacts(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite:  "art",
		Status: engine.StatusFailed,
		Scenarios: []engine.ScenarioResult{
			{Name: "diffy", Status: engine.StatusFailed, Steps: []engine.StepResult{
				{Kind: "run", Run: nil},
				{Kind: "assert", Checks: []*assert.CheckResult{{
					OK: false, Desc: "assert stdout equals golden",
					ArtifactExpected: []byte("a\nb\nc\n"), ArtifactActual: []byte("a\nX\nc\n"),
					ArtifactFiles: []assert.ArtifactFile{
						{Role: "expected", Path: "artifacts/art/expected.txt"},
						{Role: "actual", Path: "artifacts/art/actual.txt"},
					},
				}}},
			}},
		},
	}
	out := render(t, FormatConsole, res)
	for _, want := range []string{"Artifacts:", "expected: artifacts/art/expected.txt", "actual: artifacts/art/actual.txt", "Diff (-expected +actual):"} {
		if !strings.Contains(out, want) {
			t.Errorf("console artifacts footer missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestSuiteFailureMessage covers every branch of the one-line message picker.
func TestSuiteFailureMessage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   jsonFailure
		want string
	}{
		{"error wins", jsonFailure{Error: "boom", Hint: "h", Step: "s"}, "boom"},
		{"hint next", jsonFailure{Hint: "check the port", Step: "s"}, "check the port"},
		{"expected+actual", jsonFailure{Expected: "0", Actual: "3", Step: "s"}, "expected 0, got 3"},
		{"only expected", jsonFailure{Expected: "x", Step: "s"}, "expected x, got "},
		{"step fallback", jsonFailure{Step: "service setup"}, "service setup"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := suiteFailureMessage(tc.in); got != tc.want {
				t.Errorf("suiteFailureMessage(%+v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestSuiteFailureBody covers the multi-line detail assembly, including the
// empty (no detail) case.
func TestSuiteFailureBody(t *testing.T) {
	t.Parallel()
	if got := suiteFailureBody(jsonFailure{Expected: "e", Actual: "a", Hint: "h"}); got != "Expected: e\nActual: a\nHint: h" {
		t.Errorf("full body = %q", got)
	}
	if got := suiteFailureBody(jsonFailure{Hint: "only hint"}); got != "Hint: only hint" {
		t.Errorf("hint-only body = %q", got)
	}
	if got := suiteFailureBody(jsonFailure{Error: "e"}); got != "" {
		t.Errorf("error-only failure has no detail body, got %q", got)
	}
}

// TestJUnit_MessageFallbacks proves the firstFailureMessage/firstErrorMessage
// fallbacks fire when a failed/errored scenario carries no failing check or no
// step error (a shape a folded --retry result can produce).
func TestJUnit_MessageFallbacks(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite:  "fb",
		Status: engine.StatusFailed,
		Scenarios: []engine.ScenarioResult{
			// Failed status but every check OK -> "assertion failed" fallback.
			{Name: "f", Status: engine.StatusFailed, Steps: []engine.StepResult{
				{Kind: "assert", Checks: []*assert.CheckResult{{OK: true}}},
			}},
			// Errored status but no step ErrMsg -> "execution error" fallback.
			{Name: "e", Status: engine.StatusError, Steps: []engine.StepResult{
				{Kind: "run"},
			}},
		},
	}
	var root junitTestsuites
	if err := xml.Unmarshal([]byte(render(t, FormatJUnit, res)), &root); err != nil {
		t.Fatalf("junit invalid: %v", err)
	}
	out := render(t, FormatJUnit, res)
	if !strings.Contains(out, "assertion failed") {
		t.Errorf("missing firstFailureMessage fallback:\n%s", out)
	}
	if !strings.Contains(out, "execution error") {
		t.Errorf("missing firstErrorMessage fallback:\n%s", out)
	}
}

// TestDetailText_WithDiff proves the JUnit/TAP plain-text body embeds the
// uncolored unified diff for a multi-line equals/snapshot failure (#28).
func TestDetailText_WithDiff(t *testing.T) {
	t.Parallel()
	sc := &engine.ScenarioResult{
		Name:   "d",
		Status: engine.StatusFailed,
		Steps: []engine.StepResult{
			{Kind: "assert", Checks: []*assert.CheckResult{{
				OK: false, Desc: "assert stdout equals golden",
				ArtifactExpected: []byte("one\ntwo\n"), ArtifactActual: []byte("one\nTWO\n"),
			}}},
			{Index: 1, Kind: "run", ErrMsg: "later boom"},
		},
	}
	got := detailText(sc)
	for _, want := range []string{"Step: assert stdout equals golden", "Diff (-expected +actual):", "-two", "+TWO", "Error in run step: later boom"} {
		if !strings.Contains(got, want) {
			t.Errorf("detailText missing %q\n--- got ---\n%s", want, got)
		}
	}
	// The XML/TAP body must never carry an ANSI escape.
	if strings.Contains(got, "\x1b[") {
		t.Errorf("detailText leaked an ANSI escape:\n%q", got)
	}
}

// TestProgress_AllMarkers pins every marker glyph including the flaky 'f' and the
// unknown-status '?' fallback.
func TestProgress_AllMarkers(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	p := NewProgress(&b)
	p.Scenario(engine.ScenarioResult{Status: engine.StatusPassed})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusFailed})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusError})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusSkipped})
	p.Scenario(engine.ScenarioResult{Status: engine.StatusFlaky})
	p.Scenario(engine.ScenarioResult{Status: engine.Status("weird")})
	p.Done()
	if got := b.String(); got != ".FEsf?\n" {
		t.Errorf("markers = %q, want %q", got, ".FEsf?\n")
	}
}

// TestProgress_DoneNoMarkers proves Done prints nothing when no scenario ran.
func TestProgress_DoneNoMarkers(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	NewProgress(&b).Done()
	if b.Len() != 0 {
		t.Errorf("Done printed %q with no markers", b.String())
	}
}

// TestTAP_UnknownStatusIsNotOk proves the default arm of writeTAP emits a
// failing point for a status TAP has no dedicated arm for.
func TestTAP_UnknownStatusIsNotOk(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite:     "u",
		Status:    engine.StatusFailed,
		Scenarios: []engine.ScenarioResult{{Name: "weird", Status: engine.Status("bogus")}},
	}
	out := render(t, FormatTAP, res)
	if !strings.Contains(out, "not ok 1 - u / weird") {
		t.Errorf("unknown status should emit a not-ok point:\n%s", out)
	}
}

// TestVerbose_SetupStepAndErrorAndHeaderColors covers writeStep's setup label and
// error line, plus headerColor for the flaky and unknown-status arms.
func TestVerbose_SetupStepAndErrorAndHeaderColors(t *testing.T) {
	t.Parallel()
	// A setup step (Setup=true) is labeled with the shared "service setup" phrase,
	// and its ErrMsg renders on its own line.
	res := engine.ScenarioResult{
		Suite:  "v",
		Name:   "boot",
		Status: engine.StatusError,
		Steps: []engine.StepResult{
			{Index: 0, Setup: true, ErrMsg: "service never came up"},
		},
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	out := b.String()
	if !strings.Contains(out, setupPhaseLabel) {
		t.Errorf("verbose setup step should carry the setup label:\n%s", out)
	}
	if !strings.Contains(out, "error: service never came up") {
		t.Errorf("verbose setup error line missing:\n%s", out)
	}

	// headerColor arms: flaky and unknown both have deterministic returns.
	if headerColor(engine.StatusFlaky) != cYellow {
		t.Errorf("flaky header color = %q, want yellow", headerColor(engine.StatusFlaky))
	}
	if headerColor(engine.Status("weird")) != "" {
		t.Errorf("unknown status header color must be empty, got %q", headerColor(engine.Status("weird")))
	}
	if headerColor(engine.StatusPassed) != cGreen {
		t.Errorf("passed header color = %q, want green", headerColor(engine.StatusPassed))
	}
}

// TestIsTTY covers the NO_COLOR short-circuit and the non-*os.File path.
func TestIsTTY(t *testing.T) {
	// Not parallel: mutates the NO_COLOR environment variable.
	if isTTY(&bytes.Buffer{}) {
		t.Error("a bytes.Buffer is not a TTY")
	}
	t.Setenv("NO_COLOR", "1")
	// NO_COLOR forces false even for a real terminal-like file; os.Stdout is not a
	// char device under `go test`, but the env short-circuit returns first anyway.
	if isTTY(os.Stdout) {
		t.Error("NO_COLOR must force isTTY to false")
	}
}

// TestConsole_RepeatRates exercises writeRepeatRates for a --repeat run: a
// per-scenario pass-rate line. A partial-failure repeat is flaky (#138), and it
// must be reported ONCE via its rate line — not also via writeFlaky with a
// meaningless "0 attempts".
func TestConsole_RepeatRates(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite:  "rp",
		Status: engine.StatusPassed,
		Scenarios: []engine.ScenarioResult{
			{Name: "race-prone", Status: engine.StatusFlaky, Iterations: []engine.Status{
				engine.StatusPassed, engine.StatusFailed, engine.StatusPassed,
			}},
			{Name: "steady", Status: engine.StatusPassed, Iterations: []engine.Status{
				engine.StatusPassed, engine.StatusPassed,
			}},
		},
	}
	out := render(t, FormatConsole, res)
	if !strings.Contains(out, "REPEAT: race-prone: 2/3 passed") {
		t.Errorf("missing flaky repeat rate:\n%s", out)
	}
	if !strings.Contains(out, "REPEAT: steady: 2/2 passed") {
		t.Errorf("missing steady repeat rate:\n%s", out)
	}
	if strings.Contains(out, "passed after 0 attempts") || strings.Contains(out, "FLAKY: rp / race-prone") {
		t.Errorf("repeat-flaky scenario double-reported via writeFlaky:\n%s", out)
	}
}

// TestRepeatFlaky_MessageAcrossFormats proves a --repeat flake (Iterations set,
// zero retry Attempts) reports its flake RATE — not a "0 attempts" retry phrase
// — in every machine format, so a partial-failure repeat reads correctly in
// tap/gha/junit (#138).
func TestRepeatFlaky_MessageAcrossFormats(t *testing.T) {
	t.Parallel()
	res := []*engine.SuiteResult{{
		Suite:  "rp",
		Status: engine.StatusPassed,
		Scenarios: []engine.ScenarioResult{
			{Name: "race-prone", Status: engine.StatusFlaky, Duration: time.Millisecond, Iterations: []engine.Status{
				engine.StatusPassed, engine.StatusFailed, engine.StatusPassed,
			}},
		},
	}}
	for _, tc := range []struct {
		name   string
		format Format
	}{
		{"tap", FormatTAP},
		{"gha", FormatGHA},
		{"junit", FormatJUnit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			if err := Render(&b, tc.format, res); err != nil {
				t.Fatalf("render: %v", err)
			}
			out := b.String()
			if !strings.Contains(out, "flaky: 1/3 iterations failed") {
				t.Errorf("%s missing repeat flake rate:\n%s", tc.name, out)
			}
			if strings.Contains(out, "passed after 0 attempts") {
				t.Errorf("%s used the retry phrasing for a --repeat flake:\n%s", tc.name, out)
			}
		})
	}
}

// TestConsole_FailedScenarioShowsCommand proves writeDetail prints the last run
// command for a failed scenario so the reader has the invocation context.
func TestConsole_FailedScenarioShowsCommand(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{
		Suite:  "c",
		Status: engine.StatusFailed,
		Scenarios: []engine.ScenarioResult{
			{Name: "cmd", Status: engine.StatusFailed, Steps: []engine.StepResult{
				{Kind: "run", Run: &runner.Result{Command: "tool --flag"}},
				{Kind: "assert", Checks: []*assert.CheckResult{{
					OK: false, Desc: "assert exit_code is 0", Expected: "0", Actual: "1", Hint: "nonzero",
				}}},
			}},
		},
	}
	out := render(t, FormatConsole, res)
	if !strings.Contains(out, "Command:\n  tool --flag") {
		t.Errorf("failed-scenario command context missing:\n%s", out)
	}
}

// TestSuiteFailurePoints_GenericFallback proves the runtime-creation path (a
// suite that errored without scenarios AND without any recorded suite.setup
// detail) still synthesizes one generic failure point so the failure is never
// rendered green. Regression oracle for the len(pts)==0 branch.
func TestSuiteFailurePoints_GenericFallback(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{Suite: "rt", Status: engine.StatusError} // no scenarios, no Setup
	pts := suiteFailurePoints(res)
	if len(pts) != 1 || !strings.Contains(pts[0].message, "errored before any scenario ran") {
		t.Fatalf("generic fallback point = %+v", pts)
	}
	// It must surface in tap and junit as a failing/errored entry.
	tap := render(t, FormatTAP, res)
	if !strings.Contains(tap, "not ok 1 - rt / "+setupPhaseLabel) {
		t.Errorf("generic suite failure not emitted by tap:\n%s", tap)
	}
	var root junitTestsuites
	if err := xml.Unmarshal([]byte(render(t, FormatJUnit, res)), &root); err != nil {
		t.Fatalf("junit invalid: %v", err)
	}
	if root.Errors != 1 {
		t.Errorf("junit errors = %d, want 1 for the generic suite failure", root.Errors)
	}
}

// errWriter fails every write, to exercise the error-return paths of the
// writers that other tests always feed a bytes.Buffer.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errFakeWrite }

var errFakeWrite = &writeError{}

type writeError struct{}

func (*writeError) Error() string { return "boom" }

// TestWriters_PropagateWriteErrors proves each format's Render surfaces a write
// error instead of swallowing it.
func TestWriters_PropagateWriteErrors(t *testing.T) {
	t.Parallel()
	for _, f := range []Format{FormatConsole, FormatJSON, FormatJUnit, FormatGHA, FormatTAP} {
		if err := Render(errWriter{}, f, []*engine.SuiteResult{sampleResults()[0]}); err == nil {
			t.Errorf("%s: Render swallowed a write error", f)
		}
	}
}

// TestColorize covers both arms of the tiny colorize helper.
func TestColorize(t *testing.T) {
	t.Parallel()
	if got := colorize(false, cRed, "x"); got != "x" {
		t.Errorf("color off = %q, want plain", got)
	}
	if got := colorize(true, "", "x"); got != "x" {
		t.Errorf("empty code = %q, want plain", got)
	}
	if got := colorize(true, cRed, "x"); got != cRed+"x"+cReset {
		t.Errorf("colorized = %q", got)
	}
}
