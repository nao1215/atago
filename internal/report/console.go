package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/engine"
)

// writeSummary prints the final tally line. The uppercase status word
// (PASSED/FAILED) anchors the line and is part of the stable output contract.
func writeSummary(b *strings.Builder, color bool, c engine.Counts, total int, d time.Duration, hardFail bool, loadFailures int) {
	status, code := "PASSED", cGreen
	// hardFail covers a suite that errored before producing any scenario row (#7):
	// the counts are all zero, but the verdict must still read FAILED to match the
	// non-zero exit code and the SUITE SETUP FAILED block printed above.
	// loadFailures covers spec files that could not be parsed/validated (#120):
	// they run no scenario, so without this the headline would read PASSED while
	// the process exits non-zero and the dropped file is silently uncounted.
	if c.Failed > 0 || c.Errored > 0 || hardFail || loadFailures > 0 {
		status, code = "FAILED", cRed
	}
	plural := "scenarios"
	if total == 1 {
		plural = "scenario"
	}
	// Flaky scenarios (#29) are green for the verdict but never hidden: the
	// count appears only when non-zero so steady-state output is unchanged.
	flaky := ""
	if c.Flaky > 0 {
		flaky = fmt.Sprintf(", %d flaky", c.Flaky)
	}
	// Spec-load failures are not scenarios, so they get their own count in the
	// tally line — otherwise the totals silently omit the dropped files (#120).
	loadFail := ""
	if loadFailures > 0 {
		specPlural := "specs"
		if loadFailures == 1 {
			specPlural = "spec"
		}
		loadFail = fmt.Sprintf(", %d %s failed to load", loadFailures, specPlural)
	}
	fmt.Fprintf(b, "\n%s  %d %s: %d passed, %d failed, %d errored, %d skipped%s%s (%s)\n",
		colorize(color, code+cBold, status), total, plural,
		c.Passed, c.Failed, c.Errored, c.Skipped, flaky, loadFail, d.Round(time.Millisecond))
}

// writeDetail prints the failure/error block for a scenario, or nothing if it
// passed or was skipped.
func writeDetail(b *strings.Builder, color bool, suite string, sc *engine.ScenarioResult) {
	switch sc.Status {
	case engine.StatusFailed:
		cmd := lastCommand(sc)
		for _, step := range sc.Steps {
			for _, ck := range step.Checks {
				if ck == nil || ck.OK {
					continue
				}
				fmt.Fprintf(b, "\n%s %s / %s\n", colorize(color, cRed+cBold, "FAILED:"), suite, sc.Name)
				fmt.Fprintf(b, "\nStep:\n  %s\n", ck.Desc)
				if cmd != "" {
					fmt.Fprintf(b, "\nCommand:\n  %s\n", cmd)
				}
				// Multi-line equals/snapshot failures render a unified diff
				// (#28): the one-character difference the two raw blocks hide.
				// Single-line failures keep the compact Expected/Actual form.
				if diff := checkDiff(ck); diff != "" {
					fmt.Fprintf(b, "\nDiff (-expected +actual):\n%s\n", indent(colorizeDiff(color, diff)))
				} else {
					if ck.Expected != "" {
						fmt.Fprintf(b, "\nExpected:\n%s\n", indent(ck.Expected))
					}
					if ck.Actual != "" {
						fmt.Fprintf(b, "\nActual:\n%s\n", indent(ck.Actual))
					}
				}
				if ck.Hint != "" {
					fmt.Fprintf(b, "\nHint:\n  %s\n", ck.Hint)
				}
				// Keep console output compact: reference the durable payloads by path
				// rather than dumping them inline (#48).
				if len(ck.ArtifactFiles) > 0 {
					fmt.Fprintf(b, "\nArtifacts:\n")
					for _, a := range ck.ArtifactFiles {
						fmt.Fprintf(b, "  %s: %s\n", a.Role, a.Path)
					}
				}
			}
		}
	case engine.StatusError:
		for _, step := range sc.Steps {
			if step.ErrMsg == "" {
				continue
			}
			// A setup-phase error comes from before any numbered step ran; rendering
			// it as "step 0 ()" is misleading (issue #19). Use the shared setup label
			// — the ErrMsg already names the service — so every report format agrees.
			var where string
			if isSetupError(step) {
				where = setupPhaseLabel
			} else {
				where = fmt.Sprintf("step %d (%s)", step.Index, step.Kind)
			}
			fmt.Fprintf(b, "\n%s %s / %s\n  %s: %s\n",
				colorize(color, cRed+cBold, "ERROR:"), suite, sc.Name, where, step.ErrMsg)
		}
	}
	// A failed teardown never flips the verdict — the steps decide that — but
	// incomplete cleanup of external resources must stay loud.
	if sc.TeardownFailed() {
		for _, step := range sc.Teardown {
			for _, ck := range step.Checks {
				if ck == nil || ck.OK {
					continue
				}
				fmt.Fprintf(b, "\n%s %s / %s\n", colorize(color, cYellow+cBold, "TEARDOWN FAILED:"), suite, sc.Name)
				fmt.Fprintf(b, "\nStep:\n  %s\n", ck.Desc)
				if ck.Hint != "" {
					fmt.Fprintf(b, "\nHint:\n  %s\n", ck.Hint)
				}
			}
			if step.ErrMsg != "" {
				fmt.Fprintf(b, "\n%s %s / %s\n  teardown step %d (%s): %s\n",
					colorize(color, cYellow+cBold, "TEARDOWN FAILED:"), suite, sc.Name, step.Index, step.Kind, step.ErrMsg)
			}
		}
	}
	// Reference any preserved background-service logs by path (#51), keeping the
	// console compact instead of dumping the captured output inline.
	if len(sc.ServiceLogs) > 0 {
		fmt.Fprintf(b, "\nService logs:\n")
		for _, sl := range sc.ServiceLogs {
			fmt.Fprintf(b, "  %s: %s\n", sl.Name, sl.Path)
		}
	}
}

// writeRepeatRates prints one per-scenario pass-rate line for --repeat runs
// (#29): "race prone: 18/20 passed". Silent when repeat was off.
func writeRepeatRates(b *strings.Builder, color bool, res *engine.SuiteResult) {
	for i := range res.Scenarios {
		sc := &res.Scenarios[i]
		if len(sc.Iterations) == 0 {
			continue
		}
		passed := 0
		for _, st := range sc.Iterations {
			if st == engine.StatusPassed {
				passed++
			}
		}
		code := cGreen
		if passed < len(sc.Iterations) {
			code = cRed
		}
		fmt.Fprintf(b, "\n%s %s: %d/%d passed\n",
			colorize(color, code+cBold, "REPEAT:"), sc.Name, passed, len(sc.Iterations))
	}
}

// writeFlaky prints one line per recovered scenario (#29): green for the
// verdict, but never hidden.
func writeFlaky(b *strings.Builder, color bool, res *engine.SuiteResult) {
	for i := range res.Scenarios {
		sc := &res.Scenarios[i]
		if sc.Status != engine.StatusFlaky {
			continue
		}
		fmt.Fprintf(b, "\n%s %s / %s: passed after %d attempts\n",
			colorize(color, cYellow+cBold, "FLAKY:"), res.Suite, sc.Name, sc.Attempts)
	}
}

// writeSuiteDetail prints suite-level setup/teardown failure blocks (#7).
// A setup failure already errors every scenario; the block explains why. A
// teardown failure never changes the verdict but must stay loud.
func writeSuiteDetail(b *strings.Builder, color bool, res *engine.SuiteResult) {
	writeSuiteSteps(b, color, res.Suite, "SUITE SETUP FAILED:", res.Setup)
	writeSuiteSteps(b, color, res.Suite, "SUITE TEARDOWN FAILED:", res.Teardown)
}

func writeSuiteSteps(b *strings.Builder, color bool, suite, label string, steps []engine.StepResult) {
	for _, step := range steps {
		for _, ck := range step.Checks {
			if ck == nil || ck.OK {
				continue
			}
			fmt.Fprintf(b, "\n%s %s\n", colorize(color, cRed+cBold, label), suite)
			fmt.Fprintf(b, "\nStep:\n  %s\n", ck.Desc)
			if diff := checkDiff(ck); diff != "" {
				fmt.Fprintf(b, "\nDiff (-expected +actual):\n%s\n", indent(colorizeDiff(color, diff)))
			} else {
				if ck.Expected != "" {
					fmt.Fprintf(b, "\nExpected:\n%s\n", indent(ck.Expected))
				}
				if ck.Actual != "" {
					fmt.Fprintf(b, "\nActual:\n%s\n", indent(ck.Actual))
				}
			}
			if ck.Hint != "" {
				fmt.Fprintf(b, "\nHint:\n  %s\n", ck.Hint)
			}
		}
		if step.ErrMsg != "" {
			fmt.Fprintf(b, "\n%s %s\n  step %d (%s): %s\n",
				colorize(color, cRed+cBold, label), suite, step.Index, step.Kind, step.ErrMsg)
		}
	}
}

// lastCommand returns the most recent run command in a scenario, for context.
func lastCommand(sc *engine.ScenarioResult) string {
	cmd := ""
	for _, step := range sc.Steps {
		if step.Run != nil {
			cmd = step.Run.Command
		}
	}
	return cmd
}

func indent(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}
