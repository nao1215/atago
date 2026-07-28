package report

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/engine"
)

// traceExcerptLimit bounds captured output in a verbose trace, matching the
// failure-output excerpt limit so both surfaces truncate identically.
const traceExcerptLimit = 2000

// Verbose streams a per-scenario execution trace: every step's expanded
// command, exit code, captured output, and each assertion's one-line verdict —
// for passing scenarios too (#6). It answers "what did my command actually
// print?" without forcing the author to break an assertion. Failing checks are
// a one-line verdict only; the full Expected/Actual/Hint block remains the
// console report's job, so it is never duplicated.
//
// Everything rendered comes from the already-masked StepResult fields, so
// declared secrets never reach a trace.
type Verbose struct {
	mu    sync.Mutex
	w     io.Writer
	color bool
}

// NewVerbose returns a Verbose tracer writing to w. Color is enabled only when
// w is a terminal.
func NewVerbose(w io.Writer) *Verbose {
	return &Verbose{w: w, color: isTTY(w)}
}

// Scenario renders one finished scenario's trace. It is safe to use directly
// as engine.Engine.OnScenario, including from concurrent suites: the whole
// trace is built first and written in a single call so traces never interleave.
func (v *Verbose) Scenario(res engine.ScenarioResult) {
	var b strings.Builder

	status := string(res.Status)
	header := fmt.Sprintf("=== %s / %s (%s, %s)", res.Suite, res.Name, status, res.Duration.Round(time.Millisecond))
	fmt.Fprintln(&b, colorize(v.color, headerColor(res.Status), header))
	if res.Status == engine.StatusSkipped && res.SkipReason != "" {
		fmt.Fprintf(&b, "    skipped: %s\n", res.SkipReason)
	}

	// A trace of a scenario that did not pass is a trace the reader opened to
	// find something, so its empty streams are stated rather than omitted (#346).
	// See writeStream for why the split exists at all.
	interesting := res.Status == engine.StatusFailed || res.Status == engine.StatusError

	for i := range res.Steps {
		v.writeStep(&b, "", &res.Steps[i], interesting)
	}
	for i := range res.Teardown {
		v.writeStep(&b, "teardown ", &res.Teardown[i], interesting)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	fmt.Fprint(v.w, b.String())
}

// writeStep renders one step: its label line, the executed command and outcome
// for run-family steps, captured output, any execution error, and one verdict
// line per assertion check.
//
// scenarioFailed carries down the scenario's own verdict: a step that failed or
// errored is one the reader came to look at, and so is every step of a scenario
// that did not pass — the failing assert step carries no Result of its own, so
// the silent command's own step is where the stream has to be stated (#346).
func (v *Verbose) writeStep(b *strings.Builder, phase string, sr *engine.StepResult, scenarioFailed bool) {
	stateEmpty := scenarioFailed || sr.ErrMsg != "" || !assert.AllOK(sr.Checks)
	label := string(sr.Kind)
	if sr.Setup {
		label = setupPhaseLabelFor(*sr)
	}
	fmt.Fprintf(b, "  [%s%d] %s", phase, sr.Index, label)
	if sr.Run != nil && sr.Run.Command != "" {
		// The command shares this line with the step label, so it is rendered on
		// one line: a raw newline would be ambiguous with the next step's line and
		// would stop a CI-log grep mid-command (#348).
		fmt.Fprintf(b, ": %s", oneLineCommand(sr.Run.Command))
	}
	fmt.Fprintln(b)

	// Each result family renders its own observable surface; treating a DB,
	// gRPC, or CDP result like a process would print a misleading "exit 0".
	if sr.Run != nil {
		switch {
		case sr.Run.IsHTTP:
			fmt.Fprintf(b, "      status %d\n", sr.Run.StatusCode)
			writeStream(b, "body", sr.Run.Body, stateEmpty)
		case sr.Run.IsDB:
			writeStream(b, "rows", sr.Run.RowsJSON, stateEmpty)
		case sr.Run.IsGRPC:
			fmt.Fprintf(b, "      grpc status %d\n", sr.Run.GRPCStatus)
			writeStream(b, "message", sr.Run.MessageJSON, stateEmpty)
		case sr.Run.IsCDP:
			writeStream(b, "value", sr.Run.CDPValue, stateEmpty)
		default:
			fmt.Fprintf(b, "      exit %d\n", sr.Run.ExitCode)
			writeStream(b, "stdout", sr.Run.Stdout, stateEmpty)
			writeStream(b, "stderr", sr.Run.Stderr, stateEmpty)
		}
	}
	if sr.ErrMsg != "" {
		fmt.Fprintf(b, "      %s %s\n", colorize(v.color, cRed+cBold, "error:"), sr.ErrMsg)
	}
	for _, ck := range sr.Checks {
		if ck == nil {
			continue
		}
		if ck.OK {
			fmt.Fprintf(b, "      %s %s\n", colorize(v.color, cGreen, "ok  "), ck.Desc)
		} else {
			fmt.Fprintf(b, "      %s %s\n", colorize(v.color, cRed+cBold, "FAIL"), ck.Desc)
		}
	}
}

// writeStream renders one captured stream as an indented block, excerpted at
// the shared limit.
//
// An empty stream is omitted when stateEmpty is false and stated as (empty)
// when it is true. The split is deliberate (#346): compactness is a real
// argument for a 300-step green trace, where an empty stream is not what the
// reader is scanning for, and no argument at all for the step they opened the
// trace to read — there, an omitted block makes a command that printed nothing
// indistinguishable from one whose capture was lost. The rule applies to every
// result family, not just stdout/stderr: an HTTP 204 with an empty body carries
// exactly the same ambiguity.
func writeStream(b *strings.Builder, name string, data []byte, stateEmpty bool) {
	if len(data) == 0 {
		if stateEmpty {
			fmt.Fprintf(b, "      %s |\n        %s\n", name, assert.EmptyExcerpt)
		}
		return
	}
	s := string(data)
	if len(s) > traceExcerptLimit {
		s = s[:traceExcerptLimit] + "\n... (truncated)"
	}
	fmt.Fprintf(b, "      %s |\n", name)
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		fmt.Fprintf(b, "        %s\n", line)
	}
}

// headerColor picks the trace-header color for a scenario status.
func headerColor(s engine.Status) string {
	switch s {
	case engine.StatusPassed:
		return cGreen
	case engine.StatusFailed, engine.StatusError:
		return cRed + cBold
	case engine.StatusSkipped, engine.StatusFlaky:
		return cYellow
	default:
		return ""
	}
}
