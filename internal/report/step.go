package report

import (
	"strings"

	"github.com/nao1215/atago/internal/engine"
)

// setupPhaseLabel names the phase of a pre-step execution error — one that
// happened before any numbered step ran (a service-readiness failure, a
// workdir-creation failure). Such an error carries no step Kind, so every report
// format uses this single label instead of a blank field or a misleading
// "step 0 ()".
const setupPhaseLabel = "service setup"

// suiteSetupPhaseLabel names a suite.setup failure specifically, so it is not
// mislabeled as background-service readiness ("service setup"). The setup phase
// is otherwise overloaded across service readiness, suite setup, and workdir
// creation.
const suiteSetupPhaseLabel = "suite setup"

// isSetupError reports whether a step error belongs to the setup phase. The
// engine marks these explicitly with Setup; the empty-Kind fallback keeps older
// result values (and hand-built test fixtures) rendering correctly.
func isSetupError(step engine.StepResult) bool {
	return step.Setup || step.Kind == ""
}

// isSuiteSetupError reports whether a setup-phase error came from the suite.setup
// block rather than service readiness or workdir creation. The engine names that
// phase at the front of the message (suiteSetupPhaseLabel), which the generic
// "service setup" label would otherwise contradict.
func isSuiteSetupError(step engine.StepResult) bool {
	return isSetupError(step) && strings.HasPrefix(step.ErrMsg, suiteSetupPhaseLabel)
}

// setupPhaseLabelFor returns the phase label for a setup-phase error, telling a
// suite.setup failure apart from a service-readiness / workdir-creation failure
// so neither is rendered under the other's label.
func setupPhaseLabelFor(step engine.StepResult) string {
	if isSuiteSetupError(step) {
		return suiteSetupPhaseLabel
	}
	return setupPhaseLabel
}

// stepPhase returns the phase label for the machine-readable `step` field: the
// step kind, or the setup-phase label for a pre-step execution error.
func stepPhase(step engine.StepResult) string {
	if isSetupError(step) {
		return setupPhaseLabelFor(step)
	}
	return string(step.Kind)
}

// stepErrorContext renders the human phrase used in JUnit/TAP detail text
// ("in run step", "during service setup") so a setup-phase error never produces
// a blank "Error in  step".
func stepErrorContext(step engine.StepResult) string {
	if isSetupError(step) {
		return "during " + setupPhaseLabelFor(step)
	}
	return "in " + string(step.Kind) + " step"
}
