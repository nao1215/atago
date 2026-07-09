package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/fixture"
	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/runner"
	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// unresolvedRunRefMsg explains a run field that references a ${name} no variable
// defines. field names the spec field ("run.command"/"run.cwd"). shellExpandable
// says whether enabling shell could expand it: true for the command argv, false
// for cwd, which Go passes to cmd.Dir verbatim so no shell ever touches it.
func unresolvedRunRefMsg(field, name string, shellExpandable bool) string {
	if envName, isEnv := strings.CutPrefix(name, "env:"); isEnv {
		return fmt.Sprintf(
			"%s references ${env:%s}, but the environment variable %s is not set", field, envName, envName)
	}
	if shellExpandable {
		return fmt.Sprintf(
			"%s references ${%s}, but no variable with that name is defined (builtins, matrix vars, store, ready.store, env:) and shell is not enabled, so nothing would expand it; define the variable, set shell: true for shell expansion, or write $${%s} for the literal text",
			field, name, name)
	}
	return fmt.Sprintf(
		"%s references ${%s}, but no variable with that name is defined (builtins, matrix vars, store, ready.store, env:); nothing expands a working directory, so define the variable or write $${%s} for the literal text",
		field, name, name)
}

// leakedRunRefMsg explains a run field whose value, after substitution, still
// contains a ${name} reference that leaked in from a store/matrix value.
// Expansion is single-pass, so a reference carried by a substituted value is
// never re-expanded and would reach argv (or cwd) verbatim. field names the
// spec field ("run.command"/"run.cwd") and name the leaked reference.
func leakedRunRefMsg(field, name string) string {
	return fmt.Sprintf(
		"%s expands to text that still contains ${%s}: a store or matrix value used here itself contains a ${%s} reference, and variable expansion is single-pass, so that inner reference is left verbatim and would leak into the command rather than being expanded. Reference ${%s} directly in the field instead of storing a value that contains it",
		field, name, name, name)
}

// runRefGuard checks a run step for a ${...} reference that nothing could ever
// expand and that would therefore leak verbatim into the child's argv. It
// returns an explained error message, or "" when the step is clean. The rules,
// all skipped for an ssh runner (a remote shell may expand a bare ${name}):
//
//   - no-shell command: an unresolved ${name}/${env:NAME} is a typo — with no
//     shell nothing expands it, so the literal text would run;
//   - shell command: a bare ${name} passes (the shell expands it), but an unset
//     ${env:NAME} is atago-only syntax no shell understands, so it still errors;
//   - cwd: passed to cmd.Dir verbatim, never shell-expanded, so any unresolved
//     ${name} is guarded regardless of shell;
//   - a ${...} dragged in by a substituted store/matrix value survives single-
//     pass expansion into a no-shell argv (or cwd) and is caught too (#249).
//
// It is shared by execStep (scenario steps) and runSuiteSteps (suite setup and
// teardown) so both enforce the identical guard instead of one loop drifting
// from the other (#243).
func runRefGuard(st *store.Store, run *spec.Run, runners map[string]spec.Runner) string {
	if isSSHRunner(run.Runner, runners) {
		return ""
	}
	if !run.ShellEnabled() {
		if names := st.Unresolved(run.Command); len(names) > 0 {
			return unresolvedRunRefMsg("run.command", names[0], true)
		}
	} else {
		for _, name := range st.Unresolved(run.Command) {
			if strings.HasPrefix(name, "env:") {
				return unresolvedRunRefMsg("run.command", name, true)
			}
		}
	}
	if names := st.Unresolved(run.Cwd); len(names) > 0 {
		return unresolvedRunRefMsg("run.cwd", names[0], false)
	}
	if !run.ShellEnabled() {
		if _, leaked := st.ExpandDetectingLeaks(run.Command); len(leaked) > 0 {
			return leakedRunRefMsg("run.command", leaked[0])
		}
	}
	if _, leaked := st.ExpandDetectingLeaks(run.Cwd); len(leaked) > 0 {
		return leakedRunRefMsg("run.cwd", leaked[0])
	}
	return ""
}

// execStep runs one step and returns its result, its status contribution
// (passed/failed/error), and whether it breached the security policy. It is
// shared by the Steps loop and the Teardown loop; only the caller decides
// whether the contribution affects the scenario's verdict.
//
// beforeAttempt, when non-nil, is invoked immediately before each execution
// attempt of a retried run step. runSteps uses it to re-capture the `changes:`
// baseline before every attempt, so the recorded delta reflects only the final
// (converged) attempt rather than the cumulative delta of all attempts (#251).
// It is nil for the teardown path and for every non-run step kind.
func (x *scenarioRun) execStep(ctx context.Context, i int, step *spec.Step, beforeAttempt func()) (StepResult, Status, bool) {
	sr := StepResult{Index: i, Kind: step.Kind()}
	status := StatusPassed
	secViolation := false

	switch step.Kind() {
	case spec.StepFixture:
		if err := fixture.Write(expandFixture(x.st, step.Fixture), x.workdir, x.specDir); err != nil {
			sr.ErrMsg = err.Error()
			status = StatusError
		}
	case spec.StepRun:
		if msg := runRefGuard(x.st, step.Run, x.rc.runners); msg != "" {
			sr.ErrMsg = msg
			return sr, StatusError, false
		}
		run := mergeScenarioEnv(x.scEnv, expandRun(x.st, step.Run), x.st)
		r, untilChecks, err := x.e.runStep(ctx, run, x.st, x.workdir, x.specDir, x.rc, x.sshConns, beforeAttempt)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, isPolicyViolation(err)
		}
		x.current = r
		// Assertions run against the real result (current); the copy kept for
		// reporting is masked so secrets never reach logs/reports.
		sr.Run = maskResult(x.masker, r)
		// A retry's `until` condition is reported like an assertion; if it never
		// passed within the budget, the run step fails.
		if len(untilChecks) > 0 {
			x.e.recordChecks(x.masker, untilChecks, x.rc.specPath, x.sc.Name, x.idx, i)
			sr.Checks = untilChecks
			if !assert.AllOK(untilChecks) {
				status = StatusFailed
			}
		}
	case spec.StepAssert:
		crs := assert.CheckAll(expandAssert(x.st, step.Assert), x.current, assert.Env{
			Workdir:         x.workdir,
			SpecDir:         x.specDir,
			UpdateSnapshots: x.e.UpdateSnapshots,
			Secrets:         x.masker.MaskBytes,
			Scrub:           x.rc.scrubber.Apply,
			MockRecords:     x.mockRecords,
		})
		x.e.recordChecks(x.masker, crs, x.rc.specPath, x.sc.Name, x.idx, i)
		sr.Checks = crs
		if !assert.AllOK(crs) {
			status = StatusFailed
		}
	case spec.StepStore:
		val, err := extractValue(expandStore(x.st, step.Store), x.current, x.workdir)
		if err != nil {
			sr.ErrMsg = err.Error()
			status = StatusError
		} else {
			x.st.Set(step.Store.Name, val)
		}
	case spec.StepHTTP:
		r, untilChecks, secViolation, err := x.e.runHTTPStep(ctx, expandHTTP(x.st, step.HTTP), x.st, x.rc, x.workdir, x.specDir)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, secViolation
		}
		x.current = r
		sr.Run = maskResult(x.masker, r)
		// As with run retries, a never-satisfied `until` fails the step and is
		// reported like an assertion.
		if len(untilChecks) > 0 {
			x.e.recordChecks(x.masker, untilChecks, x.rc.specPath, x.sc.Name, x.idx, i)
			sr.Checks = untilChecks
			if !assert.AllOK(untilChecks) {
				status = StatusFailed
			}
		}
	case spec.StepQuery:
		r, err := x.e.runQuery(ctx, step.Query, x.st, x.rc, x.dbConns)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, false
		}
		x.current = r
		sr.Run = maskResult(x.masker, r)
	case spec.StepGRPC:
		r, err := x.e.runGRPC(ctx, expandGRPC(x.st, step.GRPC), x.st, x.rc, x.grpcConns)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, isPolicyViolation(err)
		}
		x.current = r
		sr.Run = maskResult(x.masker, r)
	case spec.StepPTY:
		r, ef, err := x.e.runPTY(ctx, step.PTY, x.st, x.scEnv, x.workdir)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, false
		}
		x.current = r
		sr.Run = maskResult(x.masker, r)
		if ef != nil {
			// A never-matching expect fails like an assertion: the pattern
			// and the transcript excerpt land in the failure block.
			ck := ptyExpectCheck(ef)
			x.e.recordChecks(x.masker, []*assert.CheckResult{ck}, x.rc.specPath, x.sc.Name, x.idx, i)
			sr.Checks = []*assert.CheckResult{ck}
			status = StatusFailed
		}
	case spec.StepCDP:
		r, err := x.e.runCDP(ctx, expandCDP(x.st, step.CDP), x.workdir, x.st, x.rc, x.browserConns)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, false
		}
		x.current = r
		sr.Run = maskResult(x.masker, r)
	case spec.StepSignal:
		// Handle-based signaling (#23): the target is a service atago
		// itself started (scenario services first, then suite services),
		// so delivery is race-free under --parallel, unlike name-based
		// kill/killall shell hacks.
		if err := runSignal(step.Signal, x.st, x.services, x.rc.suiteServices); err != nil {
			sr.ErrMsg = err.Error()
			status = StatusError
		}
	default:
		sr.ErrMsg = "step has no recognized action"
		status = StatusError
	}
	return sr, status, secViolation
}

// runSteps executes the scenario steps after the leading fixtures, scanning the
// workdir around a run/pty step when a `changes:` assert follows it (#70) and
// stopping on the first execution error.
func (x *scenarioRun) runSteps(ctx context.Context, leadingFixtures int) {
	for i := leadingFixtures; i < len(x.sc.Steps); i++ {
		step := &x.sc.Steps[i]

		// Stop before running a step if the run was canceled (Ctrl-C / parent
		// cancel / deadline). Without this the loop would keep executing steps and
		// evaluating assertions after a cancellation (issue #30).
		if ctx.Err() != nil {
			x.out.Status = StatusError
			x.out.Steps = append(x.out.Steps, StepResult{Index: i, Kind: step.Kind(), ErrMsg: fmt.Sprintf("run canceled: %v", ctx.Err())})
			break
		}

		// The `changes:` assert (#70) pins the workdir delta of the immediately
		// preceding run/pty step. Scan the workdir just before that step runs —
		// and only then, so scenarios that never use it pay nothing — capturing a
		// baseline in which prior fixture writes already exist (they are inputs,
		// not changes). Fixtures written by THIS run/pty step's redirects
		// (stdout_to/stderr_to) land after the baseline and count as created.
		var preScan fsdelta.Snapshot
		scanChanges := measurableForChanges(step.Kind()) && changesFollows(x.sc.Steps, i)
		// rescan re-captures the baseline. For a retried run step it is invoked
		// before every attempt (via execStep → runStep → pollUntil), so the delta
		// below reflects only the final, converged attempt rather than the sum of
		// every attempt's writes (#251). The pre-loop scan here still covers pty
		// steps and the no-retry path.
		var rescan func()
		if scanChanges {
			rescan = func() { preScan, _ = fsdelta.Scan(x.workdir) }
			rescan()
		}

		sr, status, secViolation := x.execStep(ctx, i, step, rescan)
		if scanChanges && x.current != nil {
			post, _ := fsdelta.Scan(x.workdir)
			delta := fsdelta.Diff(preScan, post)
			x.current.Changes = &delta
		}
		if secViolation {
			x.out.SecurityViolation = true
		}
		// Error messages can embed captured output (e.g. a failed service probe's
		// raw stdout/stderr), so mask secrets before the message reaches any report
		// (issue #12).
		sr.ErrMsg = x.masker.Mask(sr.ErrMsg)
		x.out.Steps = append(x.out.Steps, sr)
		x.out.Status = worseStatus(x.out.Status, status)
		if x.out.Status == StatusError {
			break // stop the scenario on an execution error
		}
	}
}

// Teardown always runs — after a pass, a failure, an execution error, or an
// interrupt — because it exists for external side effects the isolated
// workdir cannot undo. It shares the scenario store (a `store`-captured
// resource id flows into the cleanup request) and runs while background
// services are still up. Failures are recorded on out.Teardown for the
// reports but never change the scenario's verdict: the behavior under test
// was decided by the steps above. Every teardown step runs even if an
// earlier one failed — cleanups of independent resources must not shadow
// each other.
func (x *scenarioRun) runTeardown(ctx context.Context) {
	if len(x.sc.Teardown) > 0 {
		tctx := ctx
		if ctx.Err() != nil {
			// The run was interrupted: give cleanup its own bounded context so an
			// interrupt still tears external resources down without letting a hung
			// teardown keep the process alive.
			var cancel context.CancelFunc
			tctx, cancel = context.WithTimeout(context.Background(), teardownInterruptTimeout)
			defer cancel()
		}
		for i := range x.sc.Teardown {
			sr, _, secViolation := x.execStep(tctx, i, &x.sc.Teardown[i], nil) // teardown never carries a changes assert
			// A teardown failure never changes the scenario verdict — the behavior
			// under test was decided by the steps above. But a security-policy
			// breach (e.g. a denied network host contacted during cleanup) is not a
			// verdict question: it must still set SecurityViolation so the run does
			// not report green after a declared egress rule was violated (#248).
			if secViolation {
				x.out.SecurityViolation = true
			}
			sr.ErrMsg = x.masker.Mask(sr.ErrMsg)
			x.out.Teardown = append(x.out.Teardown, sr)
		}
	}
}

// isSSHRunner reports whether a run step's named runner is an ssh runner, which
// executes remotely where a remote shell may still expand a ${name}. An empty
// name is the default local cmd runner, and an unknown name is not ssh; both run
// the command as local argv, so the unresolved-variable guard applies to them.
func isSSHRunner(name string, runners map[string]spec.Runner) bool {
	return name != "" && runners[name].Type == "ssh"
}

// runStep executes a run step, applying retry/until polling when requested. It
// returns the final observed result, the until CheckResult (nil when no retry is
// configured), and an execution error. With retry, the command is re-run until
// until passes or the attempt budget is spent; the last attempt's result is what
// later steps observe.
func (e *Engine) runStep(ctx context.Context, run *spec.Run, st *store.Store, workdir, specDir string, rc runConfig, sshConns map[string]*sshrunner.Runner, beforeAttempt func()) (*runner.Result, []*assert.CheckResult, error) {
	// A run step naming an ssh runner executes remotely; otherwise it
	// runs locally via the cmd runner. The runner is resolved once (not per
	// retry attempt) so the timeout precedence below sees the pristine authored
	// step value.
	remote := false
	var runnerTimeout string
	if run.Runner != "" {
		rdef, ok := rc.runners[run.Runner]
		if !ok {
			return nil, nil, fmt.Errorf("run step references unknown runner %q", run.Runner)
		}
		switch rdef.Type {
		case "ssh":
			remote = true
		case "cmd", "":
			// Layer the runner's cwd beneath the step's own value; the step
			// wins. run is the caller's expanded copy, so mutating it is safe;
			// cwd gets the same use-time ${name} expansion as the other runner
			// families' fields.
			if run.Cwd == "" {
				run.Cwd = st.Expand(rdef.Cwd)
			}
			runnerTimeout = rdef.Timeout
		default:
			return nil, nil, fmt.Errorf("runner %q (type %q) cannot run a command step; use a step matching its type", run.Runner, rdef.Type)
		}
	}
	if !remote {
		// Resolve the effective timeout across all five levels (#17) and
		// remember which level supplied it so a timeout kill can name the knob
		// in its hint. Remote (ssh) runs apply only the step's own explicit
		// timeout below — the other levels shape local execution, and the ssh
		// runner's dial-time timeout already bounds every remote command.
		run.Timeout, run.TimeoutSource = resolveTimeout(run.Timeout, runnerTimeout, rc.defaultsRunTimeout, rc.suiteTimeout)
	} else if run.Timeout != "" {
		run.TimeoutSource = "run.timeout"
	}
	exec := func(ctx context.Context) (*runner.Result, error) {
		if remote {
			conn, err := sshConn(run.Runner, st, rc, sshConns)
			if err != nil {
				return nil, err
			}
			// The loader whitelists `timeout` on ssh run steps because it is
			// honored remotely — so honor it: the step's own timeout arrives at
			// the runner as a context deadline and takes precedence over the
			// runner-level timeout applied inside conn.Run.
			if run.Timeout != "" {
				d, _ := time.ParseDuration(run.Timeout) // validated at load time
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d)
				defer cancel()
			}
			r, err := conn.Run(ctx, run.Command)
			if r != nil && r.TimedOut && r.TimeoutSource == "" {
				r.TimeoutSource = run.TimeoutSource
			}
			return r, err
		}
		return e.cmd.Run(ctx, run, workdir)
	}

	if run.Retry == nil {
		// No retry: the caller's single pre-step `changes:` baseline already
		// covers this one execution, so no per-attempt rebaselining is needed.
		if beforeAttempt != nil {
			beforeAttempt()
		}
		r, err := exec(ctx)
		return r, nil, err
	}
	env := assert.Env{Workdir: workdir, SpecDir: specDir, UpdateSnapshots: e.UpdateSnapshots, Secrets: rc.masker.MaskBytes, Scrub: rc.scrubber.Apply}
	return pollUntil(ctx, run.Retry, st, env, exec, beforeAttempt)
}

// measurableForChanges reports whether a step kind produces a workdir delta a
// following `changes:` assert can pin (#70): only run and pty steps touch the
// scenario workdir as their observable effect.
func measurableForChanges(k spec.StepKind) bool {
	return k == spec.StepRun || k == spec.StepPTY
}

// changesFollows reports whether the step at index i+1 is an assert carrying a
// `changes:` target — the trigger for scanning the workdir around step i (#70).
func changesFollows(steps []spec.Step, i int) bool {
	if i+1 >= len(steps) {
		return false
	}
	a := steps[i+1].Assert
	return a != nil && a.Changes != nil
}
