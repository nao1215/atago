package engine

import (
	"context"

	"github.com/nao1215/atago/internal/spec"
)

// runScenarioWithPolicy applies the flaky-test tooling (#29) around a single
// scenario execution: --repeat folds N iterations into one result (any
// failing iteration fails it), --retry-failed re-runs a failed/errored
// scenario and reports a recovery as StatusFlaky. Every attempt/iteration
// gets a fresh isolated workdir and runs its own teardown, exactly like a
// normal run; with neither knob set this is a plain runScenario call.
func (e *Engine) runScenarioWithPolicy(ctx context.Context, idx int, sc *spec.Scenario, rc runConfig) ScenarioResult {
	switch {
	case e.Repeat > 1:
		return e.runRepeated(ctx, idx, sc, rc)
	case e.RetryFailed > 0:
		return e.runWithRetries(ctx, idx, sc, rc)
	default:
		return e.runScenario(ctx, idx, sc, rc)
	}
}

// runRepeated executes the scenario Repeat times sequentially (iterations of
// one scenario must never race themselves) and folds the outcomes: the
// reported Steps belong to the FIRST failing iteration — the interesting one
// — or the last iteration when all passed. Duration sums the iterations.
func (e *Engine) runRepeated(ctx context.Context, idx int, sc *spec.Scenario, rc runConfig) ScenarioResult {
	var folded ScenarioResult
	var iterations []Status
	haveFailure := false
	total := folded.Duration

	for i := 0; i < e.Repeat; i++ {
		if ctx.Err() != nil && i > 0 {
			break // cancelled mid-repeat: report what actually ran
		}
		run := e.runScenario(ctx, idx, sc, rc)
		iterations = append(iterations, run.Status)
		total += run.Duration
		bad := run.Status == StatusFailed || run.Status == StatusError
		if i == 0 || (bad && !haveFailure) {
			folded = run
		}
		if bad {
			haveFailure = true
		}
	}
	folded.Iterations = iterations
	folded.Duration = total
	if haveFailure && folded.Status != StatusError {
		folded.Status = StatusFailed
	}
	// When the kept result failed with StatusError, keep it — error outranks
	// failed everywhere else too.
	for _, st := range iterations {
		if st == StatusError {
			folded.Status = StatusError
			break
		}
	}
	return folded
}

// runWithRetries executes the scenario and, on failure/error, re-runs it up
// to RetryFailed more times. A re-run that passes yields StatusFlaky with the
// recovering attempt's steps; exhausted retries keep the LAST attempt's
// failure. Skipped scenarios return immediately (a skip gate is not
// instability).
func (e *Engine) runWithRetries(ctx context.Context, idx int, sc *spec.Scenario, rc runConfig) ScenarioResult {
	run := e.runScenario(ctx, idx, sc, rc)
	attempts := 1
	for (run.Status == StatusFailed || run.Status == StatusError) && attempts <= e.RetryFailed {
		if ctx.Err() != nil {
			break
		}
		retry := e.runScenario(ctx, idx, sc, rc)
		attempts++
		if retry.Status == StatusPassed {
			retry.Status = StatusFlaky
			retry.Attempts = attempts
			return retry
		}
		run = retry
	}
	run.Attempts = attempts
	return run
}
