package report

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
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
