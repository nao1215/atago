package engine

import (
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// Status is the outcome of a scenario or suite.
type Status string

const (
	// StatusPassed means every assertion passed.
	StatusPassed Status = "passed"
	// StatusFailed means at least one assertion failed.
	StatusFailed Status = "failed"
	// StatusSkipped means a skip/only condition excluded the scenario.
	StatusSkipped Status = "skipped"
	// StatusError means a step could not execute (e.g. command not found).
	StatusError Status = "error"
	// StatusXFail means a scenario declared `expect_fail:` and did fail — the
	// known bug it documents is still there (#395). It does not fail the run:
	// the whole point is to keep a reproduction in CI, executing on every
	// commit, without painting the suite red for something already known.
	StatusXFail Status = "xfail"
	// StatusXPass means a scenario declared `expect_fail:` and PASSED — the bug
	// it documents is fixed. That fails the run by default, because a spec that
	// silently starts passing is a spec nobody promotes: the failure is what
	// tells the author to move it into the suite that guards against a
	// regression. --allow-xpass downgrades it to a loud warning.
	StatusXPass Status = "xpass"
	// StatusFlaky means the scenario was unstable, not broken: it failed at
	// least once and passed at least once. Two paths produce it — a
	// --retry-failed re-run that recovered (#29), or a --repeat run where some
	// iterations passed and some failed (#138). Either way it fails the run,
	// because a scenario whose answer depends on how many times it ran has not
	// been shown to work; --allow-flaky exits 0 for a suite whose instability the
	// caller already knows about. It is surfaced everywhere with its attempt count
	// or flake rate, so the distinction from a deterministic failure is kept.
	StatusFlaky Status = "flaky"
)

// FailsRun reports whether a finished scenario turns the run red, given the two
// policies that decide the statuses whose severity is a choice rather than a
// fact: --allow-flaky and --allow-xpass.
//
// It exists so the answer is given in one place. The exit code has always known
// that a flaky recovery and an XPASS fail the run, but --fail-fast asked a
// narrower question of its own (failed or errored), so a run whose verdict was
// already decided kept scheduling scenarios — the flag stopping for one red
// signal and not another, with nothing in the reports to explain the difference.
func FailsRun(st Status, allowFlaky, allowXPass bool) bool {
	switch st {
	case StatusFailed, StatusError:
		return true
	case StatusFlaky:
		return !allowFlaky
	case StatusXPass:
		return !allowXPass
	default:
		return false
	}
}

// StepResult records what happened for a single step.
type StepResult struct {
	Index int
	Kind  spec.StepKind
	Run   *runner.Result // set for run steps
	// Checks holds one CheckResult per assertion target evaluated for this step.
	// An assert step may set several targets (exit_code + stdout + file …); each is
	// an independent check. A run step's retry `until` records its checks here too.
	// Empty for non-assert steps.
	Checks []*assert.CheckResult
	ErrMsg string // set when the step could not execute
	// Setup marks an execution error that happened before any numbered step ran
	// (a service-readiness failure, a workdir-creation failure). Such an error has
	// no step Kind, so reports label its phase explicitly ("service setup") rather
	// than emitting a blank step field or a misleading "step 0 ()".
	Setup bool
}

// ScenarioResult aggregates the steps of one scenario.
type ScenarioResult struct {
	Name string
	// Suite is the owning suite's name, so per-scenario consumers (the
	// OnScenario stream, verbose traces) can label output without threading
	// the SuiteResult alongside.
	Suite  string
	Status Status
	Steps  []StepResult
	// Teardown records the scenario's teardown steps. They always run (pass,
	// fail, error, or interrupt) and their failures are reported here, but they
	// never change Status: the verdict is decided by Steps alone.
	Teardown   []StepResult
	Duration   time.Duration
	SkipReason string
	// SecurityViolation marks a scenario that errored because it breached the
	// spec's security policy (e.g. a network-allowlist denial). It maps to exit
	// code 6 rather than the generic execution-error code.
	SecurityViolation bool
	// ServiceLogs lists the preserved combined stdout/stderr log artifacts for
	// this scenario's background services, written only when --artifacts-dir is
	// set and the scenario failed or a service never became ready (#51). Paths are
	// relative to the artifacts dir root.
	ServiceLogs []ServiceLog
	// Attempts is how many executions this result folds under --retry-failed
	// (#29): 1 for a normal run, >1 when re-runs happened (StatusFlaky when
	// one of them recovered). Zero means the feature was off.
	Attempts int
	// Iterations records each execution's status under --repeat (#29); the
	// visible Steps belong to the first failing iteration (or the last one
	// when all passed).
	Iterations []Status
	// ExpectFail carries the scenario's declared known bug when it had one
	// (#395), so every report can name the reason and issue next to the XFAIL
	// or XPASS verdict rather than making a reader open the spec.
	ExpectFail *spec.ExpectFail
}

// PassedIterations counts how many --repeat iterations came out clean (passed,
// or skipped by a deterministic OS/env gate). Paired with len(Iterations) it is
// the flake rate a repeat-flaky scenario reports ("7/10 passed"). Zero-valued
// when --repeat was not used (Iterations is empty).
func (s *ScenarioResult) PassedIterations() int {
	n := 0
	for _, st := range s.Iterations {
		if st == StatusPassed || st == StatusSkipped {
			n++
		}
	}
	return n
}

// TeardownFailed reports whether any teardown step failed or errored. Reports
// use it to stay loud about incomplete cleanup without flipping the verdict.
func (s *ScenarioResult) TeardownFailed() bool {
	for i := range s.Teardown {
		sr := &s.Teardown[i]
		if sr.ErrMsg != "" || !assert.AllOK(sr.Checks) {
			return true
		}
	}
	return false
}

// ServiceLog references one background service's preserved log artifact (#51).
type ServiceLog struct {
	Name string // the service's declared name
	Path string // relative to the artifacts dir root, slash-separated
}

// SuiteResult aggregates all scenarios of one spec file.
type SuiteResult struct {
	Suite     string
	SpecPath  string
	Status    Status
	Scenarios []ScenarioResult
	// Setup records the suite.setup steps (#7). A failed setup step errors
	// every scenario (none runs) and the failure is visible here.
	Setup []StepResult
	// Teardown records the suite.teardown steps (#7). They always run after
	// the last scenario; failures are reported but never change Status.
	Teardown []StepResult
	Duration time.Duration
	// SecurityViolation is true when any scenario breached the security policy.
	SecurityViolation bool
}

// Counts summarizes scenario outcomes.
type Counts struct {
	Passed, Failed, Skipped, Errored, Flaky int
	// XFail counts expected failures that did fail, XPass those that passed
	// (#395).
	XFail, XPass int
}

// Add returns the field-wise sum of two tallies. Every report format aggregates
// per-suite counts for its summary line; they all go through this so a new
// status is one edit here rather than a grep for open-coded additions that a
// summary line would silently disagree with the rows above it about.
func (c Counts) Add(o Counts) Counts {
	c.Passed += o.Passed
	c.Failed += o.Failed
	c.Skipped += o.Skipped
	c.Errored += o.Errored
	c.Flaky += o.Flaky
	c.XFail += o.XFail
	c.XPass += o.XPass
	return c
}

// SumCounts tallies every scenario of every suite in one call. An empty (or nil)
// slice sums to the zero value, which is what an all-filtered-out run must
// report.
func SumCounts(results []*SuiteResult) Counts {
	var agg Counts
	for _, res := range results {
		agg = agg.Add(res.Counts())
	}
	return agg
}

// Counts tallies scenario statuses.
func (s *SuiteResult) Counts() Counts {
	var c Counts
	for _, sc := range s.Scenarios {
		switch sc.Status {
		case StatusPassed:
			c.Passed++
		case StatusFailed:
			c.Failed++
		case StatusSkipped:
			c.Skipped++
		case StatusError:
			c.Errored++
		case StatusFlaky:
			c.Flaky++
		case StatusXFail:
			c.XFail++
		case StatusXPass:
			c.XPass++
		}
	}
	return c
}
