package spec

// StepPhase names one of the four places a spec carries steps. It exists
// because "which phases does this walk visit" has been the recurring silent
// bug: a summary that walked a scenario's numbered steps and nothing else
// missed the teardown that always runs, and then missed the suite lifecycle —
// each phase forgotten by a different function in a different release. A
// walker that takes the phase as data cannot forget one, and a function that
// receives the phase can still branch on it when the phases genuinely differ.
type StepPhase string

const (
	// PhaseSuiteSetup is suite.setup: once per suite, before any scenario.
	PhaseSuiteSetup StepPhase = "suite.setup"
	// PhaseSuiteTeardown is suite.teardown: once per suite, after every
	// scenario, run even when the run failed.
	PhaseSuiteTeardown StepPhase = "suite.teardown"
	// PhaseSteps is a scenario's numbered steps.
	PhaseSteps StepPhase = "steps"
	// PhaseTeardown is a scenario's teardown: run however the scenario ended.
	PhaseTeardown StepPhase = "teardown"
)

// AllStepPhases returns every phase steps can live in, in execution order. Like
// AllStepKinds, it exists so a walker can be tested for phase coverage against
// the list instead of against a reader's memory.
func AllStepPhases() []StepPhase {
	return []StepPhase{PhaseSuiteSetup, PhaseSteps, PhaseTeardown, PhaseSuiteTeardown}
}

// WalkScenarioSteps visits every step a scenario carries — the numbered steps,
// then the teardown — in declaration order, naming each step's phase. Summaries
// that describe "what this scenario does" walk through here so none of them can
// stop at Steps again: the teardown always runs, so an effect that only happens
// during cleanup is still an effect of the scenario.
func WalkScenarioSteps(sc *Scenario, visit func(phase StepPhase, st *Step)) {
	for i := range sc.Steps {
		visit(PhaseSteps, &sc.Steps[i])
	}
	for i := range sc.Teardown {
		visit(PhaseTeardown, &sc.Teardown[i])
	}
}

// WalkSuiteSteps is WalkScenarioSteps for the suite's once-per-run lifecycle:
// suite.setup, then suite.teardown.
func WalkSuiteSteps(su *Suite, visit func(phase StepPhase, st *Step)) {
	for i := range su.Setup {
		visit(PhaseSuiteSetup, &su.Setup[i])
	}
	for i := range su.Teardown {
		visit(PhaseSuiteTeardown, &su.Teardown[i])
	}
}
