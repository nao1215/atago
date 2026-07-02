package report

import (
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// TestVerbose_TracePassingScenario proves the verbose trace shows what a
// PASSING scenario actually did — the expanded command, its exit code, the
// captured output, and each assertion's one-line verdict — which is exactly
// the information a spec author otherwise only sees by breaking an assertion
// (#6).
func TestVerbose_TracePassingScenario(t *testing.T) {
	t.Parallel()
	res := engine.ScenarioResult{
		Suite:    "demo",
		Name:     "says hello",
		Status:   engine.StatusPassed,
		Duration: 12 * time.Millisecond,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "echo hello", ExitCode: 0, Stdout: []byte("hello\n")}},
			{Index: 1, Kind: spec.StepAssert, Checks: []*assert.CheckResult{{OK: true, Desc: `assert stdout contains "hello"`}}},
		},
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	out := b.String()

	for _, want := range []string{
		"demo / says hello",
		"echo hello",
		"exit 0",
		"hello",
		`ok   assert stdout contains "hello"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("trace = %q, want it to contain %q", out, want)
		}
	}
}

// TestVerbose_FailingCheckIsOneLine proves a failing assertion appears in the
// trace as a one-line verdict only — the full FAILED block (Expected/Actual/
// Hint) stays the console report's job, so it is never duplicated (#6).
func TestVerbose_FailingCheckIsOneLine(t *testing.T) {
	t.Parallel()
	res := engine.ScenarioResult{
		Suite:  "demo",
		Name:   "fails",
		Status: engine.StatusFailed,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepAssert, Checks: []*assert.CheckResult{{
				OK: false, Desc: `assert stdout contains "goodbye"`,
				Expected: "stdout contains goodbye", Actual: "hello", Hint: "the substring was not present",
			}}},
		},
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	out := b.String()

	if !strings.Contains(out, `FAIL assert stdout contains "goodbye"`) {
		t.Errorf("trace = %q, want a one-line FAIL verdict", out)
	}
	for _, forbidden := range []string{"Expected:", "Actual:", "Hint:"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("trace = %q, must not duplicate the failure block field %q", out, forbidden)
		}
	}
}

// TestVerbose_TruncatesLongOutput proves captured output is excerpted at the
// same deterministic limit failure output uses, ending with a marker (#6).
func TestVerbose_TruncatesLongOutput(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", 5000)
	res := engine.ScenarioResult{
		Suite:  "demo",
		Name:   "chatty",
		Status: engine.StatusPassed,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "yes", Stdout: []byte(long)}},
		},
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	out := b.String()

	if !strings.Contains(out, "(truncated)") {
		t.Errorf("trace should carry the truncation marker, got %d bytes without it", len(out))
	}
	if strings.Contains(out, long) {
		t.Error("trace contains the full 5000-byte payload; want it excerpted")
	}
}

// TestVerbose_SkippedScenarioShowsReason proves a skipped scenario's reason —
// currently visible only in the JSON report — is spelled out in the trace.
func TestVerbose_SkippedScenarioShowsReason(t *testing.T) {
	t.Parallel()
	res := engine.ScenarioResult{
		Suite:      "demo",
		Name:       "windows only",
		Status:     engine.StatusSkipped,
		SkipReason: "only on os=windows (host is linux)",
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	if !strings.Contains(b.String(), "only on os=windows") {
		t.Errorf("trace = %q, want the skip reason", b.String())
	}
}

// TestVerbose_TeardownStepsLabeled proves teardown steps are traced too, and
// are visually distinguishable from the scenario's own steps.
func TestVerbose_TeardownStepsLabeled(t *testing.T) {
	t.Parallel()
	res := engine.ScenarioResult{
		Suite:  "demo",
		Name:   "with cleanup",
		Status: engine.StatusPassed,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "echo main-step"}},
		},
		Teardown: []engine.StepResult{
			{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "echo cleanup-step"}},
		},
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	out := b.String()
	if !strings.Contains(out, "teardown") {
		t.Errorf("trace = %q, want a teardown label", out)
	}
	if !strings.Contains(out, "echo cleanup-step") {
		t.Errorf("trace = %q, want the teardown command", out)
	}
}
