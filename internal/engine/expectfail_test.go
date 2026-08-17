package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/spec"
)

// TestExpectFail_UpdateSnapshotsKeepsTheGolden is a regression: the golden of an
// expect_fail scenario holds the output the tool SHOULD produce once the bug is
// fixed, which is why the scenario fails against it on purpose. A blanket
// --update-snapshots rewrote it with the buggy output, destroying the recorded
// expectation and then reporting XPASS — "this scenario documents a bug that is
// no longer there" — about a bug that is still there.
func TestExpectFail_UpdateSnapshotsKeepsTheGolden(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specPath := filepath.Join(dir, "s.atago.yaml")
	const want = "the CORRECT output once fixed\n"
	if err := os.WriteFile(filepath.Join(dir, "out.snap"), []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}
	src := `
version: "1"
suite:
  name: xfail-snap
scenarios:
  - name: a known bug is still there
    expect_fail:
      reason: "output drifted from the golden"
    steps:
      - run:
          shell: true
          command: echo current buggy output
      - assert:
          stdout:
            snapshot: out.snap
`
	if err := os.WriteFile(specPath, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loader.LoadBytes(specPath, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	eng := New()
	eng.UpdateSnapshots = true
	res := eng.Run(context.Background(), s, specPath)
	if res.Status != StatusPassed {
		t.Errorf("suite status = %s, want passed (the documented bug is still there, so the run stays green)", res.Status)
	}
	if got := res.Scenarios[0].Status; got != StatusXFail {
		t.Errorf("scenario status = %s, want %s", got, StatusXFail)
	}
	golden, err := os.ReadFile(filepath.Join(dir, "out.snap"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if string(golden) != want {
		t.Errorf("golden = %q, want the recorded expectation %q kept", golden, want)
	}
}

// TestExpectFail_UpdateSnapshotsStillReportsAFixedBug pins the other half: when
// the tool starts producing what the golden holds, the scenario passes and the
// run reports XPASS under --update-snapshots exactly as it does in a verify run.
// Keeping the golden must not hide the day the bug is fixed.
func TestExpectFail_UpdateSnapshotsStillReportsAFixedBug(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specPath := filepath.Join(dir, "s.atago.yaml")
	if err := os.WriteFile(filepath.Join(dir, "out.snap"), []byte("fixed output\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := `
version: "1"
suite:
  name: xfail-snap
scenarios:
  - name: the bug is fixed
    expect_fail:
      reason: "output drifted from the golden"
    steps:
      - run:
          shell: true
          command: echo fixed output
      - assert:
          stdout:
            snapshot: out.snap
`
	if err := os.WriteFile(specPath, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loader.LoadBytes(specPath, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	eng := New()
	eng.UpdateSnapshots = true
	res := eng.Run(context.Background(), s, specPath)
	if got := res.Scenarios[0].Status; got != StatusXPass {
		t.Errorf("scenario status = %s, want %s", got, StatusXPass)
	}
}

// TestUpdateSnapshots_OrdinaryScenarioStillWrites guards the accept side: only
// an expect_fail scenario's golden is frozen, and an ordinary one in the same
// run is re-recorded as before.
func TestUpdateSnapshots_OrdinaryScenarioStillWrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	specPath := filepath.Join(dir, "s.atago.yaml")
	if err := os.WriteFile(filepath.Join(dir, "ordinary.snap"), []byte("stale\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "frozen.snap"), []byte("desired\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := `
version: "1"
suite:
  name: mixed
scenarios:
  - name: ordinary
    steps:
      - run:
          shell: true
          command: echo fresh
      - assert:
          stdout:
            snapshot: ordinary.snap
  - name: documents a bug
    expect_fail:
      reason: "still broken"
    steps:
      - run:
          shell: true
          command: echo buggy
      - assert:
          stdout:
            snapshot: frozen.snap
`
	if err := os.WriteFile(specPath, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loader.LoadBytes(specPath, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	eng := New()
	eng.UpdateSnapshots = true
	res := eng.Run(context.Background(), s, specPath)
	if res.Status != StatusPassed {
		t.Fatalf("suite status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
	ordinary, err := os.ReadFile(filepath.Join(dir, "ordinary.snap"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(ordinary)) != "fresh" {
		t.Errorf("ordinary golden = %q, want the re-recorded output", ordinary)
	}
	frozen, err := os.ReadFile(filepath.Join(dir, "frozen.snap"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(frozen)) != "desired" {
		t.Errorf("expect_fail golden = %q, want it kept", frozen)
	}
}

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
