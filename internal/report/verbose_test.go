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

// TestVerbose_NonProcessFamiliesRenderTheirOwnSurface proves DB/gRPC/CDP/HTTP
// results are traced by family instead of as a misleading process-style
// "exit 0" block (CodeRabbit finding on #9).
func TestVerbose_NonProcessFamiliesRenderTheirOwnSurface(t *testing.T) {
	t.Parallel()
	res := engine.ScenarioResult{
		Suite:  "demo",
		Name:   "families",
		Status: engine.StatusPassed,
		Steps: []engine.StepResult{
			{Index: 0, Kind: spec.StepQuery, Run: &runner.Result{IsDB: true, RowsJSON: []byte(`[{"id":1}]`)}},
			{Index: 1, Kind: spec.StepGRPC, Run: &runner.Result{IsGRPC: true, GRPCStatus: 0, MessageJSON: []byte(`{"ok":true}`)}},
			{Index: 2, Kind: spec.StepCDP, Run: &runner.Result{IsCDP: true, CDPValue: []byte("page-title")}},
			{Index: 3, Kind: spec.StepHTTP, Run: &runner.Result{IsHTTP: true, StatusCode: 201, Body: []byte(`{"id":9}`)}},
		},
	}
	var b strings.Builder
	NewVerbose(&b).Scenario(res)
	out := b.String()

	for _, want := range []string{`[{"id":1}]`, "grpc status 0", `{"ok":true}`, "page-title", "status 201"} {
		if !strings.Contains(out, want) {
			t.Errorf("trace = %q, want %q", out, want)
		}
	}
	if strings.Contains(out, "exit 0") {
		t.Errorf("trace = %q, must not render a process-style exit for non-process families", out)
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
