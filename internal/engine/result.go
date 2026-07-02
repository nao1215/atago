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
)

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
	Name   string
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
	Duration  time.Duration
	// SecurityViolation is true when any scenario breached the security policy.
	SecurityViolation bool
}

// Counts summarizes scenario outcomes.
type Counts struct {
	Passed, Failed, Skipped, Errored int
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
		}
	}
	return c
}
