package spec

import (
	"reflect"
	"testing"
)

// TestWalkers_VisitEveryPhaseInOrder pins what the walkers exist to guarantee:
// every phase steps can live in is visited, in declaration order, and named
// correctly. The recurring bug this closes off is a summary that walked a
// scenario's numbered steps and stopped — the teardown that always runs, and
// later the suite lifecycle, were each forgotten by a different function in a
// different release.
func TestWalkers_VisitEveryPhaseInOrder(t *testing.T) {
	t.Parallel()
	sc := &Scenario{
		Steps:    []Step{{Run: &Run{Command: "a"}}, {Assert: &Assert{}}},
		Teardown: []Step{{Run: &Run{Command: "b"}}},
	}
	su := &Suite{
		Setup:    []Step{{Run: &Run{Command: "c"}}},
		Teardown: []Step{{Run: &Run{Command: "d"}}, {Store: &Store{}}},
	}

	type visitRecord struct {
		phase StepPhase
		kind  StepKind
	}
	var got []visitRecord
	record := func(phase StepPhase, st *Step) { got = append(got, visitRecord{phase, st.Kind()}) }
	WalkSuiteSteps(su, record)
	WalkScenarioSteps(sc, record)

	want := []visitRecord{
		{PhaseSuiteSetup, StepRun},
		{PhaseSuiteTeardown, StepRun},
		{PhaseSuiteTeardown, StepStore},
		{PhaseSteps, StepRun},
		{PhaseSteps, StepAssert},
		{PhaseTeardown, StepRun},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("visits = %+v, want %+v", got, want)
	}
}

// TestWalkers_CoverEveryPhase is the drift guard for the phase list itself: the
// union of phases the two walkers emit must be exactly AllStepPhases. A phase
// added to the list without a walker emitting it — or emitted by a walker but
// missing from the list — fails here, so "the phases" cannot mean different
// things to the walkers and to a coverage test iterating the list.
func TestWalkers_CoverEveryPhase(t *testing.T) {
	t.Parallel()
	sc := &Scenario{Steps: []Step{{Run: &Run{}}}, Teardown: []Step{{Run: &Run{}}}}
	su := &Suite{Setup: []Step{{Run: &Run{}}}, Teardown: []Step{{Run: &Run{}}}}
	emitted := map[StepPhase]bool{}
	record := func(phase StepPhase, _ *Step) { emitted[phase] = true }
	WalkScenarioSteps(sc, record)
	WalkSuiteSteps(su, record)

	listed := map[StepPhase]bool{}
	for _, p := range AllStepPhases() {
		if listed[p] {
			t.Errorf("phase %q is listed twice", p)
		}
		listed[p] = true
		if !emitted[p] {
			t.Errorf("phase %q is in AllStepPhases but no walker emits it", p)
		}
	}
	for p := range emitted {
		if !listed[p] {
			t.Errorf("walkers emit phase %q, which AllStepPhases does not list", p)
		}
	}
}
