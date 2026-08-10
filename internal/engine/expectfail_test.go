package engine

import (
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// TestApplyExpectFail_VerdictMapping pins the whole contract in one table
// (#395). The two remappings are the feature; the two non-remappings are what
// keeps a known-bug spec from rotting.
func TestApplyExpectFail_VerdictMapping(t *testing.T) {
	t.Parallel()
	sc := &spec.Scenario{Name: "known bug", ExpectFail: &spec.ExpectFail{Reason: "still broken"}}
	tests := map[string]struct {
		in   Status
		want Status
	}{
		"a failure is the expected outcome":    {StatusFailed, StatusXFail},
		"a pass means the bug is fixed":        {StatusPassed, StatusXPass},
		"an execution error is still an error": {StatusError, StatusError},
		"a skip gate still decides on its own": {StatusSkipped, StatusSkipped},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := applyExpectFail(sc, ScenarioResult{Name: sc.Name, Status: tt.in})
			if got.Status != tt.want {
				t.Errorf("status %s -> %s, want %s", tt.in, got.Status, tt.want)
			}
			if got.ExpectFail == nil || got.ExpectFail.Reason != "still broken" {
				t.Error("the declared reason must travel with the result so reports can name it")
			}
		})
	}
}

// TestExpectFail_ErrorIsNotAnExpectedFailure states the rule the table above
// covers, by name, because it is the one that decides whether a known-bug suite
// can live in CI at all. A spec whose command disappears, or whose YAML stops
// executing, must go red — folding that into XFAIL is exactly how the
// out-of-CI known-bugs directory it replaces used to rot unnoticed.
func TestExpectFail_ErrorIsNotAnExpectedFailure(t *testing.T) {
	t.Parallel()
	sc := &spec.Scenario{ExpectFail: &spec.ExpectFail{Reason: "known"}}
	got := applyExpectFail(sc, ScenarioResult{Status: StatusError})
	if got.Status != StatusError {
		t.Errorf("status = %s, want error: expect_fail says the program answers wrongly, not that the spec cannot run", got.Status)
	}
}

// TestCounts_ExpectFailTallies proves the new statuses are counted, since the
// run-level verdict is decided by the counts rather than by the worst scenario.
func TestCounts_ExpectFailTallies(t *testing.T) {
	t.Parallel()
	res := &SuiteResult{Scenarios: []ScenarioResult{
		{Status: StatusXFail}, {Status: StatusXFail}, {Status: StatusXPass}, {Status: StatusPassed},
	}}
	c := res.Counts()
	if c.XFail != 2 || c.XPass != 1 || c.Passed != 1 {
		t.Errorf("counts = %+v, want 2 xfail / 1 xpass / 1 passed", c)
	}
}

// TestWorseStatus_ExpectFailIsNeutral keeps an xfail from dragging a suite's
// status down: like flaky, it is a fact about one scenario, and the run-level
// decision belongs to the counts.
func TestWorseStatus_ExpectFailIsNeutral(t *testing.T) {
	t.Parallel()
	for _, st := range []Status{StatusXFail, StatusXPass} {
		if got := worseStatus(StatusPassed, st); got != StatusPassed {
			t.Errorf("worseStatus(passed, %s) = %s, want passed", st, got)
		}
		if got := worseStatus(st, StatusFailed); got != StatusFailed {
			t.Errorf("worseStatus(%s, failed) = %s, want failed", st, got)
		}
	}
}
