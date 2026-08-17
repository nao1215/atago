package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/fixture"
	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/runner"
	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// assertReadsResult reports whether an assert inspects the runner result of the
// step it follows, rather than something that step left behind. file, dir,
// image, pdf, and mock targets read the filesystem or a mock server's log, so
// they say nothing about whether the command itself finished.
//
// A target not listed here counts as not reading the result, which is the loud
// direction: a new target added without updating this function leaves a timeout
// kill failing rather than silently green.
func assertReadsResult(a *spec.Assert) bool {
	if a == nil {
		return false
	}
	return a.ExitCode != nil || a.Stdout != nil || a.Stderr != nil ||
		a.Status != nil || a.Header != nil || a.Body != nil ||
		a.Rows != nil || a.GRPCStatus != nil || a.Message != nil ||
		a.Value != nil || a.Screen != nil ||
		a.Duration != nil || a.Changes != nil
}

// resultObserved reports whether an assert step inspects the result produced by
// the step at index i. Only the assert steps before the next result-producing
// step count: once another run/http/query/grpc/pty/cdp step lands, it replaces
// the current result, so a later assert is looking at that one instead.
func resultObserved(steps []spec.Step, i int) bool {
	for k := i + 1; k < len(steps); k++ {
		switch steps[k].Kind() {
		case spec.StepAssert:
			if assertReadsResult(steps[k].Assert) {
				return true
			}
		case spec.StepRun, spec.StepHTTP, spec.StepQuery, spec.StepGRPC, spec.StepPTY, spec.StepCDP:
			return false
		}
	}
	return false
}

// timeoutKillCheck turns an unobserved timeout kill into a failing check, or
// returns nil when the command exited on its own or an assert inspects it.
//
// A timeout is a guard, not an outcome to observe: issue #17 added the
// suite/built-in defaults so "an unconfigured hanging command can no longer
// stall a run", and a killed command that still reports passed delivers the
// same green verdict as before, only later. Asserting on the killed result
// catches it — that is the documented way to treat a timeout as an observable
// outcome, and it keeps working. But it requires the author to have anticipated
// the hang, and the bare `run:` step a first-time user writes had nothing to
// notice it.
//
// The escape hatch is unchanged: `timeout: "0"` at any level opts the step out,
// and a step that is never killed never reaches this check. Target is
// exit_code so the console failure block appends the captured output the
// command produced before it was killed, which is usually where the hang shows.
func timeoutKillCheck(res *runner.Result, steps []spec.Step, i int) *assert.CheckResult {
	if res == nil || !res.TimedOut || resultObserved(steps, i) {
		return nil
	}
	source := res.TimeoutSource
	if source == "" {
		source = "run.timeout"
	}
	elapsed := res.Duration.Round(time.Millisecond)
	if i < len(steps) && steps[i].Kind() == spec.StepPTY {
		// A pty step that gets here ran its whole session: a wait that never
		// completed is reported as its own expect failure instead (see
		// ptyExpectCheck), so reaching this point means the actions all went out
		// and the program stayed up. Saying "the command timed out" and offering a
		// bigger timeout is true and useless — the session is already over, so
		// waiting longer cannot change anything, and the reader who takes that
		// advice loses the time twice.
		return &assert.CheckResult{
			Target:   string(spec.AssertExitCode),
			Desc:     "pty session leaves the program exited",
			Expected: "the program to exit once the session ends",
			Actual:   fmt.Sprintf("the session finished its actions and the program was still running %s later, so it was killed", elapsed),
			Hint:     "a pty session does not close the program: send its quit key as the last action, and wait for the screen to show the program can take it (a dialog or prompt owns the keyboard until it closes). Raising the timeout does not help here, because the session already ran to the end",
		}
	}
	return &assert.CheckResult{
		Target:   string(spec.AssertExitCode),
		Desc:     "run completes before its timeout",
		Expected: "the command to exit on its own",
		Actual:   fmt.Sprintf("the command timed out after %s and was killed", elapsed),
		Hint: fmt.Sprintf(
			"the command hit its %s after %s and was killed before exiting; raise the timeout if the command is merely slow, or set timeout: \"0\" to let it run unbounded",
			source, elapsed),
	}
}

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
// adopt makes r the result later assertions read, and records the copy the
// report will show. Every protocol step (run, http, query, grpc, pty, cdp) ends
// this way, and the masked copy is what keeps a secret out of logs — a branch
// that forgot it would leak, so the pairing lives in one place.
func (x *scenarioRun) adopt(sr *StepResult, r *runner.Result) {
	x.current = r
	sr.Run = maskResult(x.masker, r)
}

// recordUntil reports a retry's `until` checks like assertions and says whether
// the step failed. A retry that never satisfied its condition within the budget
// is a failure, not a silent pass; run and http steps both retry this way.
func (x *scenarioRun) recordUntil(sr *StepResult, i int, checks []*assert.CheckResult) Status {
	if len(checks) == 0 {
		return StatusPassed
	}
	x.e.recordChecks(x.masker, checks, x.artifactScope(), i)
	sr.Checks = append(sr.Checks, checks...)
	if !assert.AllOK(checks) {
		return StatusFailed
	}
	return StatusPassed
}

// fail reports one engine-generated check (a killed timeout, a never-matching
// pty expect) the same way an assertion failure is reported.
func (x *scenarioRun) fail(sr *StepResult, i int, ck *assert.CheckResult) Status {
	x.e.recordChecks(x.masker, []*assert.CheckResult{ck}, x.artifactScope(), i)
	sr.Checks = append(sr.Checks, ck)
	return StatusFailed
}

func (x *scenarioRun) execStep(ctx context.Context, steps []spec.Step, i int, step *spec.Step, beforeAttempt func()) (StepResult, Status, bool) {
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
		return x.execRunStep(ctx, steps, i, step, beforeAttempt)
	case spec.StepAssert:
		crs := assert.CheckAll(expandAssert(x.st, step.Assert), x.current, assert.Env{
			Workdir:         x.workdir,
			SpecDir:         x.specDir,
			UpdateSnapshots: x.e.UpdateSnapshots,
			SnapshotWrites:  x.e.snapshotWrites,
			KeepSnapshots:   x.rc.keepSnapshots,
			Secrets:         x.masker.MaskBytes,
			Scrub:           x.rc.scrubber.Apply,
			MockRecords:     x.mockRecords,
		})
		x.e.recordChecks(x.masker, crs, x.artifactScope(), i)
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
		x.adopt(&sr, r)
		if x.recordUntil(&sr, i, untilChecks) == StatusFailed {
			status = StatusFailed
		}
	case spec.StepQuery:
		r, err := x.e.runQuery(ctx, step.Query, x.st, x.rc, x.dbConns)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, isPolicyViolation(err)
		}
		x.adopt(&sr, r)
	case spec.StepGRPC:
		r, err := x.e.runGRPC(ctx, expandGRPC(x.st, step.GRPC), x.st, x.rc, x.grpcConns)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, isPolicyViolation(err)
		}
		x.adopt(&sr, r)
	case spec.StepPTY:
		return x.execPTYStep(ctx, steps, i, step)
	case spec.StepCDP:
		r, err := x.e.runCDP(ctx, expandCDP(x.st, step.CDP), x.workdir, x.st, x.rc, x.browserConns)
		if err != nil {
			sr.ErrMsg = err.Error()
			return sr, StatusError, isPolicyViolation(err)
		}
		x.adopt(&sr, r)
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

// execRunStep executes one run step: the ${name} guard, the command itself,
// its retry `until` checks, and the silent-timeout-kill check.
func (x *scenarioRun) execRunStep(ctx context.Context, steps []spec.Step, i int, step *spec.Step, beforeAttempt func()) (StepResult, Status, bool) {
	sr := StepResult{Index: i, Kind: step.Kind()}
	status := StatusPassed
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
	// Assertions run against the real result (current); the copy kept for
	// reporting is masked so secrets never reach logs/reports.
	x.adopt(&sr, r)
	if x.recordUntil(&sr, i, untilChecks) == StatusFailed {
		status = StatusFailed
	}
	if ck := timeoutKillCheck(r, steps, i); ck != nil {
		status = x.fail(&sr, i, ck)
	}
	return sr, status, false
}

// execPTYStep executes one interactive pty step and folds an expect failure or
// a silent session-timeout kill into the step's checks.
func (x *scenarioRun) execPTYStep(ctx context.Context, steps []spec.Step, i int, step *spec.Step) (StepResult, Status, bool) {
	sr := StepResult{Index: i, Kind: step.Kind()}
	status := StatusPassed
	r, ef, err := x.e.runPTY(ctx, step.PTY, x.st, x.scEnv, x.workdir)
	if err != nil {
		sr.ErrMsg = err.Error()
		return sr, StatusError, false
	}
	x.adopt(&sr, r)
	if ef != nil {
		// A never-matching expect fails like an assertion: the pattern
		// and the transcript excerpt land in the failure block.
		status = x.fail(&sr, i, ptyExpectCheck(ef))
	} else if ck := timeoutKillCheck(r, steps, i); ck != nil {
		// No expect failed, so nothing else would notice that the program
		// never exited before the session timeout killed it.
		status = x.fail(&sr, i, ck)
	}
	return sr, status, false
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

		sr, status, secViolation := x.execStep(ctx, x.sc.Steps, i, step, rescan)
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
			sr, _, secViolation := x.execStep(tctx, x.sc.Teardown, i, &x.sc.Teardown[i], nil) // teardown never carries a changes assert
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
	remote, err := resolveRunTarget(run, st, rc)
	if err != nil {
		return nil, nil, err
	}
	exec := func(ctx context.Context) (*runner.Result, error) {
		if remote {
			return runRemote(ctx, run, st, rc, sshConns)
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
		if err != nil || run.Deterministic == nil {
			return r, nil, err
		}
		// The reruns happen AFTER the observable result is in hand, and r is what
		// the rest of the scenario sees. They add a claim about the command; they
		// do not change what the step means (#398).
		cr, derr := checkDeterminism(ctx, run.Deterministic, r, workdir, exec)
		if derr != nil {
			return nil, nil, derr
		}
		if cr != nil {
			return r, []*assert.CheckResult{cr}, nil
		}
		return r, nil, nil
	}
	env := assert.Env{Workdir: workdir, SpecDir: specDir, UpdateSnapshots: e.UpdateSnapshots, SnapshotWrites: e.snapshotWrites, KeepSnapshots: rc.keepSnapshots, Secrets: rc.masker.MaskBytes, Scrub: rc.scrubber.Apply}
	return pollUntil(ctx, run.Retry, st, env, exec, beforeAttempt)
}

// resolveRunTarget decides whether a run step executes remotely and settles
// its effective cwd and timeout. The runner is resolved once (not per retry
// attempt) so the timeout precedence sees the pristine authored step value.
// run is the caller's expanded copy, so mutating it is safe.
func resolveRunTarget(run *spec.Run, st *store.Store, rc runConfig) (remote bool, err error) {
	var runnerTimeout string
	if run.Runner != "" {
		rdef, ok := rc.runners[run.Runner]
		if !ok {
			return false, diag.InternalError.Errorf("run step references unknown runner %q", run.Runner)
		}
		switch rdef.Type {
		case "ssh":
			remote = true
		case "cmd", "":
			// Layer the runner's cwd beneath the step's own value; the step
			// wins. cwd gets the same use-time ${name} expansion as the other
			// runner families' fields.
			if run.Cwd == "" {
				run.Cwd = st.Expand(rdef.Cwd)
			}
			runnerTimeout = rdef.Timeout
		default:
			return false, diag.InternalError.Errorf("runner %q (type %q) cannot run a command step; use a step matching its type", run.Runner, rdef.Type)
		}
	}
	if !remote {
		// Resolve the effective timeout across all five levels (#17) and
		// remember which level supplied it so a timeout kill can name the knob
		// in its hint. Remote (ssh) runs apply only the step's own explicit
		// timeout — the other levels shape local execution, and the ssh
		// runner's dial-time timeout already bounds every remote command.
		run.Timeout, run.TimeoutSource = resolveTimeout(run.Timeout, runnerTimeout, rc.defaultsRunTimeout, rc.suiteTimeout)
	} else if run.Timeout != "" {
		run.TimeoutSource = "run.timeout"
	}
	return remote, nil
}

// runRemote executes one command over the step's ssh runner.
func runRemote(ctx context.Context, run *spec.Run, st *store.Store, rc runConfig, sshConns map[string]*sshrunner.Runner) (*runner.Result, error) {
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
