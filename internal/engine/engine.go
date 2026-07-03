// Package engine orchestrates spec execution: it plans scenarios, isolates each
// in its own temporary workdir, materializes fixtures, runs steps in order, and
// aggregates results.
package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/fixture"
	"github.com/nao1215/atago/internal/platform"
	"github.com/nao1215/atago/internal/runner"
	browserrunner "github.com/nao1215/atago/internal/runner/browser"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	dbrunner "github.com/nao1215/atago/internal/runner/db"
	grpcrunner "github.com/nao1215/atago/internal/runner/grpc"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// teardownInterruptTimeout bounds teardown execution after the run itself was
// cancelled (Ctrl-C / SIGTERM): cleanup of external resources still runs, but a
// hung teardown cannot keep an interrupted process alive indefinitely.
const teardownInterruptTimeout = 30 * time.Second

// Engine executes specs.
type Engine struct {
	cmd      runner.Runner
	builtins map[string]string

	// OnScenario, if set, is called as soon as each scenario finishes. It lets a
	// caller stream live progress (e.g. dot-style output) while a run is still
	// in flight. It must not retain the value beyond the call.
	OnScenario func(ScenarioResult)

	// UpdateSnapshots makes snapshot assertions write the snapshot file instead
	// of comparing against it.
	UpdateSnapshots bool

	// Parallel is the maximum number of scenarios to run concurrently. Values
	// < 1 mean sequential execution.
	Parallel int

	// FailFast stops scheduling new scenarios once one fails or errors.
	// In-flight scenarios are allowed to finish.
	FailFast bool

	// Sem, if set, is a shared concurrency limiter acquired around every
	// scenario. It lets a caller run multiple suites concurrently while capping
	// the TOTAL number of in-flight scenarios across all of them (a global
	// worker pool). When nil, only this suite's own Parallel workers bound it.
	Sem chan struct{}

	// FilterName, Tags, and SkipTags select which scenarios run.
	// FilterName is a substring match on the scenario name; Tags keeps only
	// scenarios carrying at least one listed tag; SkipTags drops scenarios
	// carrying any listed tag. Unselected scenarios are excluded entirely.
	FilterName string
	Tags       []string
	SkipTags   []string

	// Select, when non-nil, restricts execution to the exact scenario identities
	// it contains, keyed by ScenarioID(specPath, scenarioName). It composes with
	// (is intersected with) the filter/tag selection above and powers the
	// `--rerun-failed` red-green loop (#64). A nil map disables identity
	// selection; an empty (non-nil) map selects nothing.
	Select map[string]bool

	// Artifacts, when non-nil, is the directory into which failed text
	// assertions write durable sidecar files for review tooling (#48, the
	// --artifacts-dir flag). Nil disables artifact export.
	Artifacts *artifact.Dir
}

// New returns an Engine with the default command runner.
func New() *Engine {
	return &Engine{cmd: runnercmd.New(), builtins: builtinVars()}
}

// builtinVars are variables seeded into every scenario's store. ${atago} is the
// absolute path of the running atago binary, which lets self-hosted E2E specs
// invoke atago from inside their isolated temp workdir.
func builtinVars() map[string]string {
	m := make(map[string]string)
	if exe, err := os.Executable(); err == nil {
		m["atago"] = exe
	}
	return m
}

// Run executes every scenario in s and returns the aggregated suite result.
// Scenarios run with up to e.Parallel workers, but the returned result and the
// failure report stay in definition order for determinism.
func (e *Engine) Run(ctx context.Context, s *spec.Spec, specPath string) *SuiteResult {
	start := time.Now()
	res := &SuiteResult{Suite: s.Suite.Name, SpecPath: specPath, Status: StatusPassed}
	rc := runConfig{
		specDir:  filepath.Dir(specPath),
		specPath: specPath,
		masker:   security.NewMaskerForSpec(s),
		runners:  s.Runners,
		allow:    allowedHosts(s),
	}

	workers := e.Parallel
	if workers < 1 {
		workers = 1
	}

	selected := e.selectScenarios(s, specPath)

	// Suite lifecycle (#7): run suite.setup once before any scenario, expose
	// its scratch dir/stores/services to every scenario, and guarantee
	// suite.teardown runs after the last one (before services stop, LIFO).
	suiteRT, rtErr := e.newSuiteRuntime(s)
	if rtErr != nil {
		res.Status = StatusError
		for _, i := range selected {
			res.Scenarios = append(res.Scenarios, ScenarioResult{Name: s.Scenarios[i].Name, Suite: s.Suite.Name, Status: StatusError,
				Steps: []StepResult{{Kind: spec.StepNone, Setup: true, ErrMsg: suiteSetupLabel + ": " + rtErr.Error()}}})
		}
		res.Duration = time.Since(start)
		return res
	}
	if suiteRT != nil {
		defer suiteRT.stop()
		var setupOK bool
		res.Setup, setupOK = e.runSuiteSteps(ctx, s.Suite.Setup, suiteRT, rc, true)
		if !setupOK {
			// Every selected scenario is errored with the suite-setup phase
			// named; none of their steps run. Teardown still runs: a partially
			// executed setup may have created external state worth cleaning.
			res.Status = StatusError
			failure := suiteSetupFailure(res.Setup)
			for _, i := range selected {
				sr := ScenarioResult{Name: s.Scenarios[i].Name, Suite: s.Suite.Name, Status: StatusError,
					Steps: []StepResult{{Kind: spec.StepNone, Setup: true, ErrMsg: failure}}}
				if e.OnScenario != nil {
					e.OnScenario(sr)
				}
				res.Scenarios = append(res.Scenarios, sr)
			}
			res.Teardown = e.runSuiteTeardown(ctx, s, suiteRT, rc)
			res.Duration = time.Since(start)
			return res
		}
		rc.suiteVars = suiteRT.vars
		rc.suiteEnv = suiteRT.env
		defer func() {
			// Deferred after suiteRT.stop ⇒ runs before it, so teardown can
			// still reach the suite services.
			res.Teardown = e.runSuiteTeardown(ctx, s, suiteRT, rc)
		}()
	}

	results := make([]ScenarioResult, len(s.Scenarios))
	done := make([]bool, len(s.Scenarios))
	jobs := make(chan int)
	var mu sync.Mutex // guards OnScenario, failStop, results/done writes from the emit path
	failStop := false

	// Producer: feed selected scenario indices until fail-fast stops scheduling or
	// the run is cancelled (Ctrl-C / SIGTERM). On cancellation it stops scheduling
	// new scenarios; in-flight scenarios stop at their next step via ctx.Err().
	go func() {
		defer close(jobs)
		for _, i := range selected {
			if ctx.Err() != nil {
				return
			}
			mu.Lock()
			stop := failStop
			mu.Unlock()
			if stop {
				return
			}
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				// A job may have been queued just before another worker tripped
				// fail-fast; re-check so it is left as skipped rather than run.
				mu.Lock()
				stop := failStop
				mu.Unlock()
				if stop {
					continue
				}
				if e.Sem != nil {
					// The global semaphore is shared across every suite in the run.
					// On Ctrl-C, slots free up only as in-flight scenarios unwind, so
					// waiting for one here would both stall shutdown and then run a
					// scenario the user already cancelled. Bail instead; the scenario
					// is reported as "skipped after interrupt".
					select {
					case e.Sem <- struct{}{}:
					case <-ctx.Done():
						continue
					}
				}
				sc := e.runScenario(ctx, idx, &s.Scenarios[idx], rc)
				sc.Suite = s.Suite.Name
				if e.Sem != nil {
					<-e.Sem
				}
				mu.Lock()
				results[idx] = sc
				done[idx] = true
				if e.OnScenario != nil {
					e.OnScenario(sc)
				}
				if e.FailFast && (sc.Status == StatusFailed || sc.Status == StatusError) {
					failStop = true
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Build the report from the selected scenarios in definition order. Selected
	// scenarios that never ran (stopped by fail-fast) are recorded as skipped.
	// A scenario that was selected but never ran was stopped either by fail-fast or
	// by an interrupt; name the cause so the report is unambiguous.
	unrunReason := "skipped after fail-fast"
	if ctx.Err() != nil {
		unrunReason = "skipped after interrupt"
	}
	res.Scenarios = make([]ScenarioResult, 0, len(selected))
	for _, i := range selected {
		if !done[i] {
			results[i] = ScenarioResult{Name: s.Scenarios[i].Name, Suite: s.Suite.Name, Status: StatusSkipped, SkipReason: unrunReason}
		}
		res.Scenarios = append(res.Scenarios, results[i])
		res.Status = worseStatus(res.Status, results[i].Status)
		if results[i].SecurityViolation {
			res.SecurityViolation = true
		}
	}
	res.Duration = time.Since(start)
	return res
}

// selectScenarios returns the indices of scenarios that pass the filter/tag
// selection, in definition order.
func (e *Engine) selectScenarios(s *spec.Spec, specPath string) []int {
	var out []int
	for i := range s.Scenarios {
		if e.matches(&s.Scenarios[i], specPath) {
			out = append(out, i)
		}
	}
	return out
}

func (e *Engine) matches(sc *spec.Scenario, specPath string) bool {
	if e.FilterName != "" && !strings.Contains(sc.Name, e.FilterName) {
		return false
	}
	if len(e.Tags) > 0 && !hasAnyTag(sc, e.Tags) {
		return false
	}
	if len(e.SkipTags) > 0 && hasAnyTag(sc, e.SkipTags) {
		return false
	}
	if e.Select != nil && !e.Select[ScenarioID(specPath, sc.Name)] {
		return false
	}
	return true
}

// ScenarioID is the stable identity of a scenario for cross-run selection: its
// spec path joined with its (already matrix-expanded) name. It is the key used
// by the --rerun-failed state file and Engine.Select (#64).
func ScenarioID(specPath, scenarioName string) string {
	return specPath + "\x00" + scenarioName
}

func hasAnyTag(sc *spec.Scenario, tags []string) bool {
	for _, want := range tags {
		for _, have := range sc.Tags {
			if have == want {
				return true
			}
		}
	}
	return false
}

// runConfig carries the per-spec execution context shared by every scenario in a
// suite: where the spec lives, the secret masker, the named runners, and the
// network allowlist enforced for HTTP steps.
type runConfig struct {
	specDir  string
	specPath string
	masker   *security.Masker
	runners  map[string]spec.Runner
	allow    []string
	// suiteVars is the suite store snapshot (#7): ${suitedir} plus values
	// captured by suite.setup (store steps, service ready.store). Seeded into
	// every scenario's store before its own vars.
	suiteVars map[string]string
	// suiteEnv is the raw suite.env, layered beneath each scenario's env.
	suiteEnv map[string]string
}

func (e *Engine) runScenario(ctx context.Context, scenarioIdx int, sc *spec.Scenario, rc runConfig) ScenarioResult {
	specDir, masker := rc.specDir, rc.masker
	if reason, skip := e.skipReason(ctx, sc); skip {
		return ScenarioResult{Name: sc.Name, Status: StatusSkipped, SkipReason: reason}
	}

	start := time.Now()
	out := ScenarioResult{Name: sc.Name, Status: StatusPassed}

	workdir, err := os.MkdirTemp("", "atago-")
	if err != nil {
		out.Status = StatusError
		out.Steps = append(out.Steps, StepResult{Kind: spec.StepNone, Setup: true, ErrMsg: fmt.Sprintf("could not create workdir: %v", err)})
		return out
	}
	defer os.RemoveAll(workdir)

	st := store.New()
	for k, v := range e.builtins {
		st.Set(k, v)
	}
	// Suite-level values (#7) come before the scenario's own: ${suitedir} and
	// suite.setup captures are shared context every scenario may reference.
	for k, v := range rc.suiteVars {
		st.Set(k, v)
	}
	// ${workdir} is the absolute path of this scenario's isolated temp dir, so
	// specs can build absolute env paths (e.g. HOME=${workdir}/home,
	// GOBIN=${workdir}/gobin) that child toolchains require.
	st.Set("workdir", workdir)
	// Seed matrix row variables so ${name} references in commands, env, and
	// assertions resolve to this instance's values.
	for k, v := range sc.Vars {
		st.Set(k, v)
	}
	// suite.env layers beneath the scenario's own env (the scenario wins per
	// key); scEnv replaces sc.Env for every use below.
	scEnv := mergedEnv(rc.suiteEnv, sc.Env)
	// Database connections are opened lazily per scenario and closed at its end,
	// so a dsn referencing ${workdir} yields a fresh isolated DB each time.
	dbConns := map[string]*dbrunner.Runner{}
	sshConns := map[string]*sshrunner.Runner{}
	grpcConns := map[string]*grpcrunner.Runner{}
	browserConns := map[string]*browserrunner.Runner{}
	defer func() {
		for _, c := range dbConns {
			_ = c.Close()
		}
		for _, c := range sshConns {
			_ = c.Close()
		}
		for _, c := range grpcConns {
			_ = c.Close()
		}
		for _, c := range browserConns {
			_ = c.Close()
		}
	}()
	// Leading fixture steps — the uninterrupted prefix of steps that are all
	// fixtures — are applied BEFORE services start, so a background server can
	// consume authored input (its config file, seed data) the way a real daemon
	// does. Fixtures after the first non-fixture step keep their in-order,
	// after-services timing, so a scenario can still simulate files appearing
	// while the service runs.
	leadingFixtures := 0
	for leadingFixtures < len(sc.Steps) && sc.Steps[leadingFixtures].Kind() == spec.StepFixture {
		fx := sc.Steps[leadingFixtures].Fixture
		sr := StepResult{Index: leadingFixtures, Kind: spec.StepFixture}
		if err := fixture.Write(expandFixture(st, fx), workdir, specDir); err != nil {
			sr.ErrMsg = err.Error()
			out.Steps = append(out.Steps, sr)
			out.Status = StatusError
			out.Duration = time.Since(start)
			return out
		}
		out.Steps = append(out.Steps, sr)
		leadingFixtures++
	}

	// Background services (ADR-0031) start after the store is seeded — so their
	// commands can reference ${workdir} and matrix vars — and after the leading
	// fixtures, but before any other step runs. They are stopped in LIFO order
	// when the scenario ends, however it ends.
	var services []*servicerunner.Proc
	defer func() {
		for i := len(services) - 1; i >= 0; i-- {
			services[i].Stop()
		}
	}()
	for i := range sc.Services {
		proc, captured, err := servicerunner.Start(ctx, expandService(st, scEnv, &sc.Services[i]), workdir)
		if err != nil {
			out.Status = StatusError
			// Preserve the failed service's log (and any peers already started) as a
			// durable artifact before tearing down, so a readiness failure stays
			// inspectable (#51). proc is the stopped-but-readable failed service.
			if proc != nil {
				services = append(services, proc)
			}
			e.writeServiceLogs(&out, masker, services, rc.specPath, sc.Name, scenarioIdx)
			// The readiness error can embed the service's raw output, so mask secrets
			// before it reaches a report (issue #12).
			out.Steps = append(out.Steps, StepResult{Kind: spec.StepNone, Setup: true, ErrMsg: masker.Mask(err.Error())})
			out.Duration = time.Since(start)
			return out
		}
		services = append(services, proc)
		if sc.Services[i].Ready != nil && sc.Services[i].Ready.Store != "" {
			st.Set(sc.Services[i].Ready.Store, captured)
		}
	}

	var current *runner.Result

	// execStep runs one step and returns its result, its status contribution
	// (passed/failed/error), and whether it breached the security policy. It is
	// shared by the Steps loop and the Teardown loop; only the caller decides
	// whether the contribution affects the scenario's verdict.
	execStep := func(ctx context.Context, i int, step *spec.Step) (StepResult, Status, bool) {
		sr := StepResult{Index: i, Kind: step.Kind()}
		status := StatusPassed
		secViolation := false

		switch step.Kind() {
		case spec.StepFixture:
			if err := fixture.Write(expandFixture(st, step.Fixture), workdir, specDir); err != nil {
				sr.ErrMsg = err.Error()
				status = StatusError
			}
		case spec.StepRun:
			// A ${name} that no variable defines is left verbatim so a shell can
			// still expand it — but a local run without `shell: true`
			// has no shell, so nothing could ever expand it and the literal text
			// would leak into argv. That is almost always a typo; error with the
			// reference named instead of running a garbled command (#UX).
			if !step.Run.ShellEnabled() && step.Run.Runner == "" {
				if names := st.Unresolved(step.Run.Command); len(names) > 0 {
					if envName, isEnv := strings.CutPrefix(names[0], "env:"); isEnv {
						sr.ErrMsg = fmt.Sprintf(
							"run.command references ${env:%[1]s}, but the environment variable %[1]s is not set", envName)
					} else {
						sr.ErrMsg = fmt.Sprintf(
							"run.command references ${%[1]s}, but no variable with that name is defined (builtins, matrix vars, store, ready.store, env:) and shell is not enabled, so nothing would expand it; define the variable, set shell: true for shell expansion, or write $${%[1]s} for the literal text",
							names[0])
					}
					return sr, StatusError, false
				}
			}
			run := mergeScenarioEnv(scEnv, expandRun(st, step.Run), st)
			r, untilChecks, err := e.runStep(ctx, run, st, workdir, specDir, rc, sshConns)
			if err != nil {
				sr.ErrMsg = err.Error()
				return sr, StatusError, isPolicyViolation(err)
			}
			current = r
			// Assertions run against the real result (current); the copy kept for
			// reporting is masked so secrets never reach logs/reports.
			sr.Run = maskResult(masker, r)
			// A retry's `until` condition is reported like an assertion; if it never
			// passed within the budget, the run step fails.
			if len(untilChecks) > 0 {
				e.recordChecks(masker, untilChecks, rc.specPath, sc.Name, scenarioIdx, i)
				sr.Checks = untilChecks
				if !assert.AllOK(untilChecks) {
					status = StatusFailed
				}
			}
		case spec.StepAssert:
			crs := assert.CheckAll(expandAssert(st, step.Assert), current, assert.Env{
				Workdir:         workdir,
				SpecDir:         specDir,
				UpdateSnapshots: e.UpdateSnapshots,
				Secrets:         masker.MaskBytes,
			})
			e.recordChecks(masker, crs, rc.specPath, sc.Name, scenarioIdx, i)
			sr.Checks = crs
			if !assert.AllOK(crs) {
				status = StatusFailed
			}
		case spec.StepStore:
			val, err := extractValue(expandStore(st, step.Store), current, workdir)
			if err != nil {
				sr.ErrMsg = err.Error()
				status = StatusError
			} else {
				st.Set(step.Store.Name, val)
			}
		case spec.StepHTTP:
			r, untilChecks, secViolation, err := e.runHTTPStep(ctx, expandHTTP(st, step.HTTP), st, rc, workdir, specDir)
			if err != nil {
				sr.ErrMsg = err.Error()
				return sr, StatusError, secViolation
			}
			current = r
			sr.Run = maskResult(masker, r)
			// As with run retries, a never-satisfied `until` fails the step and is
			// reported like an assertion.
			if len(untilChecks) > 0 {
				e.recordChecks(masker, untilChecks, rc.specPath, sc.Name, scenarioIdx, i)
				sr.Checks = untilChecks
				if !assert.AllOK(untilChecks) {
					status = StatusFailed
				}
			}
		case spec.StepQuery:
			r, err := e.runQuery(ctx, step.Query, st, rc, dbConns)
			if err != nil {
				sr.ErrMsg = err.Error()
				return sr, StatusError, false
			}
			current = r
			sr.Run = maskResult(masker, r)
		case spec.StepGRPC:
			r, err := e.runGRPC(ctx, expandGRPC(st, step.GRPC), st, rc, grpcConns)
			if err != nil {
				sr.ErrMsg = err.Error()
				return sr, StatusError, isPolicyViolation(err)
			}
			current = r
			sr.Run = maskResult(masker, r)
		case spec.StepPTY:
			r, ef, err := e.runPTY(ctx, step.PTY, st, scEnv, workdir)
			if err != nil {
				sr.ErrMsg = err.Error()
				return sr, StatusError, false
			}
			current = r
			sr.Run = maskResult(masker, r)
			if ef != nil {
				// A never-matching expect fails like an assertion: the pattern
				// and the transcript excerpt land in the failure block.
				ck := ptyExpectCheck(ef)
				e.recordChecks(masker, []*assert.CheckResult{ck}, rc.specPath, sc.Name, scenarioIdx, i)
				sr.Checks = []*assert.CheckResult{ck}
				status = StatusFailed
			}
		case spec.StepCDP:
			r, err := e.runCDP(ctx, expandCDP(st, step.CDP), workdir, st, rc, browserConns)
			if err != nil {
				sr.ErrMsg = err.Error()
				return sr, StatusError, false
			}
			current = r
			sr.Run = maskResult(masker, r)
		default:
			sr.ErrMsg = "step has no recognized action"
			status = StatusError
		}
		return sr, status, secViolation
	}

	for i := leadingFixtures; i < len(sc.Steps); i++ {
		step := &sc.Steps[i]

		// Stop before running a step if the run was cancelled (Ctrl-C / parent
		// cancel / deadline). Without this the loop would keep executing steps and
		// evaluating assertions after a cancellation (issue #30).
		if ctx.Err() != nil {
			out.Status = StatusError
			out.Steps = append(out.Steps, StepResult{Index: i, Kind: step.Kind(), ErrMsg: fmt.Sprintf("run cancelled: %v", ctx.Err())})
			break
		}

		sr, status, secViolation := execStep(ctx, i, step)
		if secViolation {
			out.SecurityViolation = true
		}
		// Error messages can embed captured output (e.g. a failed service probe's
		// raw stdout/stderr), so mask secrets before the message reaches any report
		// (issue #12).
		sr.ErrMsg = masker.Mask(sr.ErrMsg)
		out.Steps = append(out.Steps, sr)
		out.Status = worseStatus(out.Status, status)
		if out.Status == StatusError {
			break // stop the scenario on an execution error
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
	if len(sc.Teardown) > 0 {
		tctx := ctx
		if ctx.Err() != nil {
			// The run was interrupted: give cleanup its own bounded context so an
			// interrupt still tears external resources down without letting a hung
			// teardown keep the process alive.
			var cancel context.CancelFunc
			tctx, cancel = context.WithTimeout(context.Background(), teardownInterruptTimeout)
			defer cancel()
		}
		for i := range sc.Teardown {
			sr, _, _ := execStep(tctx, i, &sc.Teardown[i])
			sr.ErrMsg = masker.Mask(sr.ErrMsg)
			out.Teardown = append(out.Teardown, sr)
		}
	}

	// Preserve running services' logs when a step failed or errored after the
	// services came up, so a post-readiness failure is just as inspectable as a
	// readiness failure (#51). Green runs write nothing (artifact-dir + failure
	// gated), keeping logs opt-in rather than mandatory noise.
	if out.Status == StatusFailed || out.Status == StatusError {
		e.writeServiceLogs(&out, masker, services, rc.specPath, sc.Name, scenarioIdx)
	}

	out.Duration = time.Since(start)
	return out
}

// skipReason reports whether a scenario should be skipped given its skip/only
// conditions: the host OS, an environment variable's presence, and — last,
// because it spawns a process — a probe command's exit status (ADR-0021). The cheap, side-effect-free checks run first so a probe only runs
// when nothing else already decided the outcome.
func (e *Engine) skipReason(ctx context.Context, sc *spec.Scenario) (string, bool) {
	if sc.Only != nil && sc.Only.OS != "" && !platform.Matches(sc.Only.OS) {
		return fmt.Sprintf("only on os=%s (host is %s)", sc.Only.OS, platform.OS()), true
	}
	if sc.Skip != nil && sc.Skip.OS != "" && platform.Matches(sc.Skip.OS) {
		return fmt.Sprintf("skip on os=%s", sc.Skip.OS), true
	}
	if sc.Only != nil && sc.Only.Env != "" && os.Getenv(sc.Only.Env) == "" {
		return fmt.Sprintf("only when env %s is set", sc.Only.Env), true
	}
	if sc.Skip != nil && sc.Skip.Env != "" && os.Getenv(sc.Skip.Env) != "" {
		return fmt.Sprintf("skip when env %s is set", sc.Skip.Env), true
	}
	if sc.Only != nil && sc.Only.Command != "" && !e.probeSucceeds(ctx, sc.Only.Command) {
		return fmt.Sprintf("only when command %q succeeds", sc.Only.Command), true
	}
	if sc.Skip != nil && sc.Skip.Command != "" && e.probeSucceeds(ctx, sc.Skip.Command) {
		return fmt.Sprintf("skip when command %q succeeds", sc.Skip.Command), true
	}
	return "", false
}

// probeSucceeds runs a skip/only probe command through the shell and reports
// whether it exited 0. A probe that cannot start at all counts as a failure (not
// succeeded), so `only: { command }` skips rather than erroring on a missing
// tool. The probe runs in a throwaway temp dir so it cannot touch the cwd; if
// that dir cannot be created it falls back to the process cwd.
func (e *Engine) probeSucceeds(ctx context.Context, command string) bool {
	dir, err := os.MkdirTemp("", "atago-probe-")
	if err == nil {
		defer os.RemoveAll(dir)
	}
	res, err := e.cmd.Run(ctx, &spec.Run{Command: command, Shell: spec.Bool(true)}, dir)
	if err != nil {
		return false
	}
	return res.ExitCode == 0
}

// runStep executes a run step, applying retry/until polling when requested. It
// returns the final observed result, the until CheckResult (nil when no retry is
// configured), and an execution error. With retry, the command is re-run until
// until passes or the attempt budget is spent; the last attempt's result is what
// later steps observe (ADR-0022).
func (e *Engine) runStep(ctx context.Context, run *spec.Run, st *store.Store, workdir, specDir string, rc runConfig, sshConns map[string]*sshrunner.Runner) (*runner.Result, []*assert.CheckResult, error) {
	// A run step naming an ssh runner executes remotely (ADR-0027); otherwise it
	// runs locally via the cmd runner.
	exec := func(ctx context.Context) (*runner.Result, error) {
		if run.Runner != "" {
			rdef, ok := rc.runners[run.Runner]
			if !ok {
				return nil, fmt.Errorf("run step references unknown runner %q", run.Runner)
			}
			switch rdef.Type {
			case "ssh":
				conn, err := sshConn(run.Runner, st, rc, sshConns)
				if err != nil {
					return nil, err
				}
				return conn.Run(ctx, run.Command)
			case "cmd", "":
				// Layer the runner's cwd/timeout beneath the step's own values
				//; the
				// step wins. run is the caller's expanded copy, so mutating it is
				// safe; cwd gets the same use-time ${name} expansion as the other
				// runner families' fields.
				if run.Cwd == "" {
					run.Cwd = st.Expand(rdef.Cwd)
				}
				if run.Timeout == "" {
					run.Timeout = rdef.Timeout
				}
				// fall through to the local cmd runner
			default:
				return nil, fmt.Errorf("runner %q (type %q) cannot run a command step; use a step matching its type", run.Runner, rdef.Type)
			}
		}
		return e.cmd.Run(ctx, run, workdir)
	}

	if run.Retry == nil {
		r, err := exec(ctx)
		return r, nil, err
	}

	interval, _ := time.ParseDuration(run.Retry.Interval) // validated at load time
	until := expandAssert(st, run.Retry.Until)
	env := assert.Env{Workdir: workdir, SpecDir: specDir, UpdateSnapshots: e.UpdateSnapshots, Secrets: rc.masker.MaskBytes}

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

// worseStatus returns the more severe of two statuses (error > failed > passed,
// skipped is neutral at suite level).
func worseStatus(a, b Status) Status {
	rank := map[Status]int{StatusPassed: 0, StatusSkipped: 0, StatusFailed: 1, StatusError: 2}
	if rank[b] > rank[a] {
		return b
	}
	return a
}
