package report

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
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
		c.Passed, c.Failed, c.Errored, c.Skipped, flakySuffix(c), loadFail, d.Round(time.Millisecond))
}

// writeDetail prints the failure/error block for a scenario, or nothing if it
// passed or was skipped. specPath names the spec file in each block header so a
// multi-file directory run points at the YAML to edit instead of forcing a grep
// for the scenario name; it is empty in direct API use, which keeps the header
// in its historical shape.
func writeDetail(b *strings.Builder, color bool, suite, specPath string, sc *engine.ScenarioResult) {
	where := specSuffix(specPath)
	switch sc.Status {
	case engine.StatusFailed:
		var lastRun *runner.Result
		for _, step := range sc.Steps {
			// Track the most recent run result BEFORE rendering this step's
			// checks: a run step's own retry `until` checks assert against that
			// step's result, and every failure block names THAT command.
			if step.Run != nil {
				lastRun = step.Run
			}
			for _, ck := range step.Checks {
				if ck == nil || ck.OK {
					continue
				}
				fmt.Fprintf(b, "\n%s %s / %s%s\n", colorize(color, cRed+cBold, "FAILED:"), suite, sc.Name, where)
				writeCheckFailure(b, color, ck, commandOf(lastRun))
				writeFailureStreams(b, ck, lastRun)
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
			// it as "step 0 ()" is misleading. Use the phase label — distinguishing a
			// suite.setup failure from service readiness so neither is mislabeled — so
			// every report format agrees.
			var phase string
			if isSetupError(step) {
				phase = setupPhaseLabelFor(step)
			} else {
				phase = fmt.Sprintf("step %d (%s)", step.Index, step.Kind)
			}
			fmt.Fprintf(b, "\n%s %s / %s%s\n  %s: %s\n",
				colorize(color, cRed+cBold, "ERROR:"), suite, sc.Name, where, phase, step.ErrMsg)
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
				fmt.Fprintf(b, "\n%s %s / %s%s\n", colorize(color, cYellow+cBold, "TEARDOWN FAILED:"), suite, sc.Name, where)
				fmt.Fprintf(b, "\nStep:\n  %s\n", ck.Desc)
				if ck.Hint != "" {
					fmt.Fprintf(b, "\nHint:\n  %s\n", ck.Hint)
				}
			}
			if step.ErrMsg != "" {
				fmt.Fprintf(b, "\n%s %s / %s%s\n  teardown step %d (%s): %s\n",
					colorize(color, cYellow+cBold, "TEARDOWN FAILED:"), suite, sc.Name, where, step.Index, step.Kind, step.ErrMsg)
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
		// Color by the fold's verdict: all clean is green, a partial failure is
		// flaky (yellow, green for the exit code), and an all-failed repeat is a
		// deterministic red failure (#138).
		code := cGreen
		switch sc.Status {
		case engine.StatusFlaky:
			code = cYellow
		case engine.StatusFailed, engine.StatusError:
			code = cRed
		}
		fmt.Fprintf(b, "\n%s %s: %d/%d passed\n",
			colorize(color, code+cBold, "REPEAT:"), sc.Name, passed, len(sc.Iterations))
	}
}

// writeFlaky prints one line per --retry-failed recovery (#29): green for the
// verdict, but never hidden. A --repeat flake (Iterations set) is reported by
// writeRepeatRates instead, with its rate; skip it here so it is not
// double-reported with a meaningless "0 attempts".
func writeFlaky(b *strings.Builder, color bool, res *engine.SuiteResult) {
	for i := range res.Scenarios {
		sc := &res.Scenarios[i]
		if sc.Status != engine.StatusFlaky || len(sc.Iterations) > 0 {
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
			writeCheckFailure(b, color, ck, "")
		}
		if step.ErrMsg != "" {
			fmt.Fprintf(b, "\n%s %s\n  step %d (%s): %s\n",
				colorize(color, cRed+cBold, label), suite, step.Index, step.Kind, step.ErrMsg)
		}
	}
}

// commandOf names the command a failure was observed under: the result the
// failing check asserts on, not the scenario's last command. Using the last one
// mislabeled every failure in a scenario that keeps working afterwards — an
// interactive step that timed out was reported against the verification command
// that followed it and passed, which sends the reader to tune a step that never
// had a problem.
func commandOf(res *runner.Result) string {
	if res == nil {
		return ""
	}
	return res.Command
}

// specSuffix renders the spec-file annotation appended to failure headers, or
// "" when the result carries no path (direct API use).
func specSuffix(specPath string) string {
	if specPath == "" {
		return ""
	}
	return fmt.Sprintf("  (%s)", filepath.ToSlash(specPath))
}

// failureStreamTail bounds how much captured output an exit_code failure block
// inlines: enough to show the actual error a command printed on its way out,
// small enough to keep the console report scannable. --verbose still shows
// everything.
const failureStreamTail = 10

// writeFailureStreams appends the failing command's captured stderr tail (or
// stdout tail when stderr is empty) to an exit_code failure block. The WHY of
// a non-zero exit almost always sits on stderr, and hiding it behind --verbose
// made the most common first failure a dead end. Stream asserts (stdout,
// stderr, body, …) already display the stream they matched, so only exit_code
// failures grow this section. run.Stderr/Stdout arrive already masked by the
// engine, like every other StepResult field the reports render.
func writeFailureStreams(b *strings.Builder, ck *assert.CheckResult, run *runner.Result) {
	if ck.Target != string(spec.AssertExitCode) || run == nil {
		return
	}
	stream, label := run.Stderr, "Stderr"
	if len(stream) == 0 {
		stream, label = run.Stdout, "Stdout"
	}
	if len(stream) == 0 {
		// The command exited non-zero without printing a word. That is a notable
		// fact about the program under test — arguably the most notable one in
		// this report — and the block used to stay quiet about it, which is what
		// a reader also sees when there is no rule at all (#346). Say it instead,
		// in the one spelling internal/assert already uses.
		//
		// Only for a result family that HAS these streams: an HTTP response or a
		// DB query has no stdout, so "(empty)" there would describe a stream that
		// does not exist rather than one that was silent.
		if hasProcessStreams(run) {
			fmt.Fprintf(b, "\nStdout/Stderr:\n  %s\n", assert.EmptyExcerpt)
		}
		return
	}
	tail, kept, total := tailLines(stream, failureStreamTail)
	if kept < total {
		fmt.Fprintf(b, "\n%s (last %d of %d lines):\n%s\n", label, kept, total, indent(tail))
		return
	}
	fmt.Fprintf(b, "\n%s:\n%s\n", label, indent(tail))
}

// hasProcessStreams reports whether a Result belongs to a family that has
// stdout/stderr at all — a process run or a pty session. An HTTP, DB, gRPC, or
// CDP result has its own observable surface (body, rows, message, value) and
// leaves Stdout/Stderr permanently nil, so describing them as "(empty)" would
// name a stream that does not exist rather than one that was silent (#346).
func hasProcessStreams(run *runner.Result) bool {
	return !run.IsHTTP && !run.IsDB && !run.IsGRPC && !run.IsCDP
}

// tailLines returns the last n lines of buf (sans trailing newline), how many
// lines it kept, and how many the buffer held in total.
func tailLines(buf []byte, n int) (tail string, kept, total int) {
	lines := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	total = len(lines)
	if total > n {
		lines = lines[total-n:]
	}
	return strings.Join(lines, "\n"), len(lines), total
}

func indent(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}

// oneLineCommand renders a command for a surface where it shares a line with
// other text — the verbose trace's step line — so a newline in it can never
// split that line. A raw newline there is ambiguous with the start of the next
// step, and a grep for a step in a CI log stops matching at it (#348).
//
// A command with no embedded newline is returned untouched: that is nearly
// every command, and this text is what users grep for, so the common case must
// stay byte-identical. Only a multi-line command is quoted, and strconv.Quote is
// the escaping chosen because `\n` is the same escape the spec author typed in
// the YAML — the trace then reads recognizably like the source, and like the
// failure block's indented rendering of the same command. A trailing newline (a
// `>` block scalar carries one) is dropped first: it says nothing about the
// command, and quoting it would show a `\n` the author did not write.
func oneLineCommand(cmd string) string {
	cmd = strings.TrimRight(cmd, "\r\n")
	if !strings.ContainsAny(cmd, "\r\n") {
		return cmd
	}
	return strconv.Quote(cmd)
}

// writeCheckFailure renders the body every failure report shares: the step that
// failed, the command behind it when there is one, the comparison, and the
// hint. A scenario failure and a suite setup/teardown failure differ only in
// their heading and in whether captured streams follow, so the body lives here
// once — a change to how a comparison reads cannot apply to one and miss the
// other.
func writeCheckFailure(b *strings.Builder, color bool, ck *assert.CheckResult, cmd string) {
	fmt.Fprintf(b, "\nStep:\n  %s\n", ck.Desc)
	if cmd != "" {
		// Through indent(), like every other multi-line value in this block: a
		// command carrying a newline (a YAML block scalar, a `shell: true` script)
		// used to drop its continuation lines to column 0 (#348). indent() also
		// trims the trailing newline a `>` block scalar leaves, so the section
		// never ends in a dangling blank line.
		fmt.Fprintf(b, "\nCommand:\n%s\n", indent(cmd))
	}
	// Multi-line equals/snapshot failures render a unified diff (#28): the
	// one-character difference the two raw blocks hide. Single-line failures
	// keep the compact Expected/Actual form.
	if diff := checkDiff(ck); diff != "" {
		fmt.Fprintf(b, "\nDiff (-expected +actual):\n%s\n", indent(colorizeDiff(color, diff)))
	} else {
		expected, actual := visiblePair(ck.Expected, ck.Actual)
		if expected != "" {
			fmt.Fprintf(b, "\nExpected:\n%s\n", indent(expected))
		}
		if actual != "" {
			fmt.Fprintf(b, "\nActual:\n%s\n", indent(actual))
		}
	}
	if ck.Hint != "" {
		fmt.Fprintf(b, "\nHint:\n  %s\n", ck.Hint)
	}
}
