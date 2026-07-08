package engine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nao1215/atago/internal/platform"
	"github.com/nao1215/atago/internal/runner"
	browserrunner "github.com/nao1215/atago/internal/runner/browser"
	dbrunner "github.com/nao1215/atago/internal/runner/db"
	grpcrunner "github.com/nao1215/atago/internal/runner/grpc"
	mockrunner "github.com/nao1215/atago/internal/runner/mock"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	sshrunner "github.com/nao1215/atago/internal/runner/ssh"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// scenarioRun carries the mutable per-scenario execution state that the
// resource-lifecycle (lifecycle.go), step-executor and teardown (stepexec.go)
// phases share. Splitting runScenario across methods on this struct keeps each
// phase readable while the state stays in one place.
type scenarioRun struct {
	e       *Engine
	idx     int
	sc      *spec.Scenario
	rc      runConfig
	specDir string
	masker  *security.Masker

	workdir string
	st      *store.Store
	scEnv   map[string]string
	// current is the most recent step's raw result; assertions run against it.
	current *runner.Result

	mocks        []*mockrunner.Server
	services     []*servicerunner.Proc
	dbConns      map[string]*dbrunner.Runner
	sshConns     map[string]*sshrunner.Runner
	grpcConns    map[string]*grpcrunner.Runner
	browserConns map[string]*browserrunner.Runner

	out   ScenarioResult
	start time.Time
}

// runScenario executes one scenario end to end in an isolated temp workdir:
// skip check, resource setup (store, mock servers, leading fixtures, services),
// the step loop, teardown, and cleanup. It orchestrates the phases; the phases
// themselves live in lifecycle.go and stepexec.go.
func (e *Engine) runScenario(ctx context.Context, scenarioIdx int, sc *spec.Scenario, rc runConfig) ScenarioResult {
	if reason, skip := e.skipReason(ctx, sc); skip {
		return ScenarioResult{Name: sc.Name, Status: StatusSkipped, SkipReason: reason}
	}

	x := &scenarioRun{
		e:            e,
		idx:          scenarioIdx,
		sc:           sc,
		rc:           rc,
		specDir:      rc.specDir,
		masker:       rc.masker,
		dbConns:      map[string]*dbrunner.Runner{},
		sshConns:     map[string]*sshrunner.Runner{},
		grpcConns:    map[string]*grpcrunner.Runner{},
		browserConns: map[string]*browserrunner.Runner{},
		start:        time.Now(),
		out:          ScenarioResult{Name: sc.Name, Status: StatusPassed},
	}

	workdir, err := os.MkdirTemp("", "atago-")
	if err != nil {
		x.out.Status = StatusError
		x.out.Steps = append(x.out.Steps, StepResult{Kind: spec.StepNone, Setup: true, ErrMsg: fmt.Sprintf("could not create workdir: %v", err)})
		return x.out
	}
	x.workdir = workdir
	// Cleanup runs LIFO at scenario end, however it ends: stop services, then
	// mock servers, then close the lazy connections, then remove the workdir.
	// Each stop/close tolerates an empty set, so registering them up front is
	// safe even when an early resource-setup failure returns before they fill.
	defer os.RemoveAll(workdir)
	defer x.closeConns()
	defer x.stopMocks()
	defer x.stopServices()

	x.initStore()

	if !x.startMocks(ctx) {
		return x.out
	}
	leadingFixtures, ok := x.applyLeadingFixtures()
	if !ok {
		return x.out
	}
	if !x.startServices(ctx) {
		return x.out
	}

	x.runSteps(ctx, leadingFixtures)
	x.runTeardown(ctx)

	// Preserve running services' logs when a step failed or errored after the
	// services came up, so a post-readiness failure is just as inspectable as a
	// readiness failure (#51). Green runs write nothing (artifact-dir + failure
	// gated), keeping logs opt-in rather than mandatory noise.
	if x.out.Status == StatusFailed || x.out.Status == StatusError {
		x.e.writeServiceLogs(&x.out, x.masker, x.services, x.rc.specPath, x.sc.Name, x.idx)
		x.e.writeMockLogs(&x.out, x.masker, x.mocks, x.rc.specPath, x.sc.Name, x.idx)
	}

	x.out.Duration = time.Since(x.start)
	return x.out
}

// skipReason reports whether a scenario should be skipped given its skip/only
// conditions: the host OS, an environment variable's presence, and — last,
// because it spawns a process — a probe command's exit status. The cheap, side-effect-free checks run first so a probe only runs
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
	// A probe is a quick selection check, not a step, so bound it: a hanging
	// probe would otherwise stall the sequential selection phase indefinitely.
	if e.probeTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.probeTimeout)
		defer cancel()
	}
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
