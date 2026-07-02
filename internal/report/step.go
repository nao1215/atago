package report

import "github.com/nao1215/atago/internal/engine"

// setupPhaseLabel names the phase of a pre-step execution error — one that
// happened before any numbered step ran (a service-readiness failure, a
// workdir-creation failure). Such an error carries no step Kind, so every report
// format uses this single label instead of a blank field or a misleading
// "step 0 ()".
const setupPhaseLabel = "service setup"

// isSetupError reports whether a step error belongs to the setup phase. The
// engine marks these explicitly with Setup; the empty-Kind fallback keeps older
// result values (and hand-built test fixtures) rendering correctly.
func isSetupError(step engine.StepResult) bool {
	return step.Setup || step.Kind == ""
}

// stepPhase returns the phase label for the machine-readable `step` field: the
// step kind, or the setup-phase label for a pre-step execution error.
func stepPhase(step engine.StepResult) string {
	if isSetupError(step) {
		return setupPhaseLabel
	}
	return string(step.Kind)
}

// stepErrorContext renders the human phrase used in JUnit/TAP detail text
// ("in run step", "during service setup") so a setup-phase error never produces
// a blank "Error in  step".
func stepErrorContext(step engine.StepResult) string {
	if isSetupError(step) {
		return "during " + setupPhaseLabel
	}
	return "in " + string(step.Kind) + " step"
}
