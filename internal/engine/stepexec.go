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

// execStep runs one step and returns its result, its status contribution
// (passed/failed/error), and whether it breached the security policy. It is
// shared by the Steps loop and the Teardown loop; only the caller decides
// whether the contribution affects the scenario's verdict.
func (x *scenarioRun) execStep(ctx context.Context, i int, step *spec.Step) (StepResult, Status, bool) {
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
		// A ${name} that no variable defines is left verbatim so a shell can
		// still expand it — but a local run without `shell: true`
		// has no shell, so nothing could ever expand it and the literal text
		// would leak into argv. That is almost always a typo; error with the
		// reference named instead of running a garbled command (#UX). A named
		// cmd runner runs the command as local argv too, so it is guarded like
		// the default runner; only an ssh runner (remote, where a remote shell
		// may expand it) is exempt.
		if !isSSHRunner(step.Run.Runner, x.rc.runners) {
			if !step.Run.ShellEnabled() {
				if names := x.st.Unresolved(step.Run.Command); len(names) > 0 {
					sr.ErrMsg = unresolvedRunRefMsg("run.command", names[0], true)
					return sr, StatusError, false
				}
			}
			// cwd is passed to cmd.Dir verbatim; no shell ever expands it, so an
			// unresolved ${name} is always a typo that would make the child fail to
			// start in a literal "${name}" directory and surface as a misleading
			// "executable not found". Guard it regardless of shell.
			if names := x.st.Unresolved(step.Run.Cwd); len(names) > 0 {
				sr.ErrMsg = unresolvedRunRefMsg("run.cwd", names[0], false)
				return sr, StatusError, false
			}
		}
		run := mergeScenarioEnv(x.scEnv, expandRun(x.st, step.Run), x.st)
		r, untilChecks, err := x.e.runStep(ctx, run, x.st, x.workdir, x.specDir, x.rc, x.sshConns)
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
		if scanChanges {
			preScan, _ = fsdelta.Scan(x.workdir)
		}

		sr, status, secViolation := x.execStep(ctx, i, step)
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
			sr, _, _ := x.execStep(tctx, i, &x.sc.Teardown[i])
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
func (e *Engine) runStep(ctx context.Context, run *spec.Run, st *store.Store, workdir, specDir string, rc runConfig, sshConns map[string]*sshrunner.Runner) (*runner.Result, []*assert.CheckResult, error) {
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
		// in its hint. Remote (ssh) runs are bounded by the ssh runner's own
		// connection timeout instead.
		run.Timeout, run.TimeoutSource = resolveTimeout(run.Timeout, runnerTimeout, rc.defaultsRunTimeout, rc.suiteTimeout)
	}
	exec := func(ctx context.Context) (*runner.Result, error) {
		if remote {
			conn, err := sshConn(run.Runner, st, rc, sshConns)
			if err != nil {
				return nil, err
			}
			return conn.Run(ctx, run.Command)
		}
		return e.cmd.Run(ctx, run, workdir)
	}

	if run.Retry == nil {
		r, err := exec(ctx)
		return r, nil, err
	}

	interval, _ := time.ParseDuration(run.Retry.Interval) // validated at load time
	until := expandAssert(st, run.Retry.Until)
	env := assert.Env{Workdir: workdir, SpecDir: specDir, UpdateSnapshots: e.UpdateSnapshots, Secrets: rc.masker.MaskBytes, Scrub: rc.scrubber.Apply}

	var last *runner.Result
	var checks []*assert.CheckResult
	for attempt := 1; attempt <= run.Retry.Times; attempt++ {
		r, err := exec(ctx)
		if err != nil {
			return nil, nil, err
		}
		last = r
		checks = assert.CheckAll(until, r, env)
		if assert.AllOK(checks) {
			break
		}
		if attempt < run.Retry.Times && interval > 0 {
			select {
			case <-ctx.Done():
				return last, checks, nil
			case <-time.After(interval):
			}
		}
	}
	return last, checks, nil
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
