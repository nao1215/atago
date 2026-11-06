package report

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

func TestRender_ErrorRetainsFailedChecks(t *testing.T) {
	t.Parallel()
	res := &engine.SuiteResult{Suite: "errors", Status: engine.StatusError,
		Scenarios: []engine.ScenarioResult{{Name: "earlier checks", Status: engine.StatusError,
			Steps: []engine.StepResult{
				{Index: 0, Kind: spec.StepRun, Run: &runner.Result{Command: "first-command"}},
				{Index: 1, Kind: spec.StepAssert, Checks: []*assert.CheckResult{nil, {OK: true, Desc: "passing-check"}, {Desc: "first-failed-check", Expected: "one", Actual: "zero"}}},
				// A retry may carry checks and an execution error on the same step.
				{Index: 2, Kind: spec.StepRun, Run: &runner.Result{Command: "second-command"}, Checks: []*assert.CheckResult{{Desc: "second-failed-check"}}, ErrMsg: "execution-error"},
			},
		}},
	}
	for _, format := range AllFormats() {
		t.Run(string(format), func(t *testing.T) {
			t.Parallel()
			out := render(t, format, res)
			for _, evidence := range []string{"first-failed-check", "second-failed-check", "execution-error"} {
				if !strings.Contains(out, evidence) {
					t.Errorf("%s lost %s:\n%s", format, evidence, out)
				}
			}
			// JUnit/TAP summarize the error before the detailed body; their body
			// already includes all step evidence. Console and GHA print it in order.
			if format == FormatConsole || format == FormatGHA {
				first, second, last := strings.Index(out, "first-failed-check"), strings.Index(out, "second-failed-check"), strings.Index(out, "execution-error")
				if first < 0 || second <= first || last <= second {
					t.Errorf("evidence out of step order:\n%s", out)
				}
				for _, evidence := range []string{"first-failed-check", "second-failed-check", "execution-error"} {
					if strings.Count(out, evidence) != 1 {
						t.Errorf("evidence duplicated: %s\n%s", evidence, out)
					}
				}
				if strings.Contains(out, "passing-check") {
					t.Errorf("passing check reported as failure:\n%s", out)
				}
			}
			if format == FormatConsole && (!strings.Contains(out, "first-command") || !strings.Contains(out, "second-command")) {
				t.Errorf("failed checks lost command context:\n%s", out)
			}
		})
	}
}
