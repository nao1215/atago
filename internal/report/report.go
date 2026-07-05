// Package report renders SuiteResults in the supported output formats
// : console, JSON, JUnit XML, and GitHub Actions annotations.
package report

import (
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// suiteErroredWithoutScenarios reports a suite that failed or errored before it
// produced any scenario row: a suite.setup failure (#7) with nothing selected to
// run (all scenarios filtered out, or an empty scenario list), or a
// suite-runtime creation failure. exitForSuite maps such a suite to a non-zero
// code and console/JSON show the cause, but with no scenario rows the
// junit/tap/gha bodies would otherwise render an all-green empty suite that
// contradicts the exit code — so those formats synthesize a failure entry from
// suiteFailurePoints, and the console summary reads FAILED.
func suiteErroredWithoutScenarios(res *engine.SuiteResult) bool {
	return len(res.Scenarios) == 0 &&
		(res.Status == engine.StatusFailed || res.Status == engine.StatusError)
}

// suiteFailurePoint is a synthetic failure entry for a suite that errored
// without scenarios: one per recorded suite.setup failure, or a single generic
// entry when no setup detail was captured (the runtime-creation path).
type suiteFailurePoint struct {
	name    string // point/testcase name, e.g. "service setup"
	message string // one-line failure message
	body    string // optional detail block
}

// suiteFailurePoints builds the entries junit/tap/gha emit for a suite that
// errored without scenarios. It mirrors the suite.setup failures JSON already
// reports (setup_failures) and the console already shows (SUITE SETUP FAILED).
func suiteFailurePoints(res *engine.SuiteResult) []suiteFailurePoint {
	var pts []suiteFailurePoint
	for _, f := range suiteStepFailures(res.Suite, res.Setup) {
		pts = append(pts, suiteFailurePoint{
			name:    f.Step,
			message: suiteFailureMessage(f),
			body:    suiteFailureBody(f),
		})
	}
	if len(pts) == 0 {
		// The runtime-creation path records its message on a would-be scenario,
		// not res.Setup; with nothing selected there is no detail to show. Emit
		// one generic point so the failure is never rendered green.
		pts = append(pts, suiteFailurePoint{name: setupPhaseLabel, message: "suite errored before any scenario ran"})
	}
	return pts
}

// suiteFailureMessage picks the most informative one-line message from a
// suite-level failure record.
func suiteFailureMessage(f jsonFailure) string {
	switch {
	case f.Error != "":
		return f.Error
	case f.Hint != "":
		return f.Hint
	case f.Expected != "" || f.Actual != "":
		return "expected " + f.Expected + ", got " + f.Actual
	default:
		return f.Step
	}
}

// suiteFailureBody renders the multi-line detail block for a suite-level
// failure, matching the console/junit failure body shape.
func suiteFailureBody(f jsonFailure) string {
	var parts []string
	if f.Expected != "" {
		parts = append(parts, "Expected: "+f.Expected)
	}
	if f.Actual != "" {
		parts = append(parts, "Actual: "+f.Actual)
	}
	if f.Hint != "" {
		parts = append(parts, "Hint: "+f.Hint)
	}
	return strings.Join(parts, "\n")
}

// Format names a built-in report format.
type Format string

const (
	// FormatConsole is the default human-readable output.
	FormatConsole Format = "console"
	// FormatJSON is the machine-readable report.
	FormatJSON Format = "json"
	// FormatJUnit is a JUnit XML report.
	FormatJUnit Format = "junit"
	// FormatGHA emits GitHub Actions workflow-command annotations.
	FormatGHA Format = "gha"
	// FormatTAP emits a Test Anything Protocol (TAP 13) stream.
	FormatTAP Format = "tap"
)

// Valid reports whether f is a known report format. Render is the single entry
// point that dispatches on Format to the per-format build/write helpers.
func (f Format) Valid() bool {
	switch f {
	case FormatConsole, FormatJSON, FormatJUnit, FormatGHA, FormatTAP:
		return true
	default:
		return false
	}
}
