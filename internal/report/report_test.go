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
