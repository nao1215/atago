// Package engine orchestrates spec execution: it plans scenarios, isolates each
// in its own temporary workdir, materializes fixtures, runs steps in order, and
// aggregates results.
package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nao1215/atago/internal/artifact"
	"github.com/nao1215/atago/internal/runner"
	runnercmd "github.com/nao1215/atago/internal/runner/cmd"
	mockrunner "github.com/nao1215/atago/internal/runner/mock"
	servicerunner "github.com/nao1215/atago/internal/runner/service"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
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
	// In-flight scenarios are allowed to finish. With RetryFailed it triggers
	// only on a FINAL failure (a recovered flaky scenario never trips it).
	FailFast bool

	// Repeat runs each selected scenario this many times (#29) to surface
	// flakiness: any failing iteration fails the scenario, and the per-
	// iteration statuses are recorded. Values < 2 disable it.
	Repeat int

	// RetryFailed re-runs a failed/errored scenario up to this many times in
	// a fresh workdir (#29); a recovering re-run yields StatusFlaky — counted
	// green for the exit code but reported loudly. Mutually exclusive with
	// Repeat (the CLI enforces it).
	RetryFailed int

	// Sem, if set, is a shared concurrency limiter acquired around every
	// scenario. It lets a caller run multiple suites concurrently while capping
	// the TOTAL number of in-flight scenarios across all of them (a global
	// worker pool). When nil, only this suite's own Parallel workers bound it.
	Sem chan struct{}

	// FilterNames, Tags, and SkipTags select which scenarios run.
	// FilterNames keeps a scenario whose name contains ANY listed substring
	// (OR semantics, mirroring Tags); Tags keeps only scenarios carrying at
	// least one listed tag; SkipTags drops scenarios carrying any listed tag.
	// Unselected scenarios are excluded entirely.
	FilterNames []string
	Tags        []string
	SkipTags    []string

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

	// probeTimeout bounds a skip/only probe command so a hanging probe cannot
	// stall the (sequential) selection phase. Zero means unbounded.
	probeTimeout time.Duration
}

// defaultProbeTimeout bounds a skip/only probe command: quick checks like
// `command -v tool` need far less, but a generous ceiling still stops a
// pathological probe (`sleep 9999`) from hanging the selection phase.
const defaultProbeTimeout = 30 * time.Second

// New returns an Engine with the default command runner.
func New() *Engine {
	return &Engine{cmd: runnercmd.New(), builtins: builtinVars(), probeTimeout: defaultProbeTimeout}
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
		specDir:      filepath.Dir(specPath),
		specPath:     specPath,
		masker:       security.NewMaskerForSpec(s),
		runners:      s.Runners,
		allow:        allowedHosts(s),
		suiteTimeout: s.Suite.Timeout,
	}
	if s.Defaults != nil && s.Defaults.Run != nil {
		rc.defaultsRunTimeout = s.Defaults.Run.Timeout
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
		rc.suiteServices = suiteRT.services
		rc.suiteMocks = suiteRT.mocks
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
				sc := e.runScenarioWithPolicy(ctx, idx, &s.Scenarios[idx], rc)
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
	if len(e.FilterNames) > 0 && !nameMatchesAny(sc.Name, e.FilterNames) {
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

// nameMatchesAny reports whether name contains any of the substrings (OR),
// giving --filter the same multi-value semantics as --tag.
func nameMatchesAny(name string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(name, sub) {
			return true
		}
	}
	return false
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
	// suiteTimeout / defaultsRunTimeout feed the step-timeout precedence
	// resolver (#17): step > runner > defaults.run > suite > built-in 60s.
	// They stay separate strings (instead of being merged into each step at
	// load time) so the resolver knows which level supplied the winning value
	// and can name it in the timeout-kill hint.
	suiteTimeout       string
	defaultsRunTimeout string
	// suiteServices are the suite-wide background processes started by
	// suite.setup service steps (#7), threaded here so a scenario's `signal:`
	// step (#23) can target them by name alongside its own services.
	suiteServices []*servicerunner.Proc
	// suiteMocks are the suite-wide stub HTTP servers (#24), threaded here so
	// scenario `mock:` asserts can read their recorded requests.
	suiteMocks []*mockrunner.Server
}

// worseStatus returns the more severe of two statuses (error > failed > passed,
// skipped is neutral at suite level).
func worseStatus(a, b Status) Status {
	rank := map[Status]int{StatusPassed: 0, StatusSkipped: 0, StatusFlaky: 0, StatusFailed: 1, StatusError: 2}
	if rank[b] > rank[a] {
		return b
	}
	return a
}
