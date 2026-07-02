package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/engine"
)

// writeSummary prints the final tally line. The uppercase status word
// (PASSED/FAILED) anchors the line and is part of the stable output contract.
func writeSummary(b *strings.Builder, color bool, c engine.Counts, total int, d time.Duration) {
	status, code := "PASSED", cGreen
	if c.Failed > 0 || c.Errored > 0 {
		status, code = "FAILED", cRed
	}
	plural := "scenarios"
	if total == 1 {
		plural = "scenario"
	}
	fmt.Fprintf(b, "\n%s  %d %s: %d passed, %d failed, %d errored, %d skipped (%s)\n",
		colorize(color, code+cBold, status), total, plural,
		c.Passed, c.Failed, c.Errored, c.Skipped, d.Round(time.Millisecond))
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
				if ck.Expected != "" {
					fmt.Fprintf(b, "\nExpected:\n%s\n", indent(ck.Expected))
				}
				if ck.Actual != "" {
					fmt.Fprintf(b, "\nActual:\n%s\n", indent(ck.Actual))
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
	// Reference any preserved background-service logs by path (#51), keeping the
	// console compact instead of dumping the captured output inline.
	if len(sc.ServiceLogs) > 0 {
		fmt.Fprintf(b, "\nService logs:\n")
		for _, sl := range sc.ServiceLogs {
			fmt.Fprintf(b, "  %s: %s\n", sl.Name, sl.Path)
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
