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
//
// The folded Status distinguishes the two failure shapes --repeat exists to
// tell apart (#138): a scenario that failed EVERY iteration is a deterministic
// bug (StatusFailed, or StatusError when the failures were execution errors),
// while one that passed some iterations and failed others is unstable, not
// broken, and folds to StatusFlaky — surfaced with its flake rate but green for
// the exit code, exactly like a --retry-failed recovery. Collapsing a partial
// failure into StatusFailed (the old behavior) erased that distinction, so
// "3/10 flaked" was indistinguishable from "10/10 is a real bug".
func (e *Engine) runRepeated(ctx context.Context, idx int, sc *spec.Scenario, rc runConfig) ScenarioResult {
	var folded ScenarioResult
	var iterations []Status
	haveFailure := false
	total := folded.Duration

	for i := 0; i < e.Repeat; i++ {
		if ctx.Err() != nil && i > 0 {
			break // canceled mid-repeat: report what actually ran
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

	// Classify the fold by how many iterations came out clean. A skip gate is
	// deterministic (every iteration skips), so a skipped iteration counts as
	// "not a failure" alongside a pass.
	passed, errored := 0, 0
	for _, st := range iterations {
		switch st {
		case StatusPassed, StatusSkipped:
			passed++
		case StatusError:
			errored++
		}
	}
	switch bad := len(iterations) - passed; {
	case bad == 0:
		// Every iteration was clean; folded already holds a passing/skipped run.
	case passed == 0:
		// Never passed: a deterministic failure, not instability. Error outranks
		// failed (a step that could not execute is worse than an assertion miss).
		if errored > 0 {
			folded.Status = StatusError
		} else {
			folded.Status = StatusFailed
		}
	default:
		// Passed some, failed some: unstable. This is the flake --repeat is built
		// to catch; folded keeps the first failing iteration's steps.
		folded.Status = StatusFlaky
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
