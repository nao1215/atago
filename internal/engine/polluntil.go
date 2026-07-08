package engine

import (
	"context"
	"time"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

// pollUntil is the single retry/until loop shared by run and http steps: it
// executes exec up to retry.Times, evaluating retry.Until against each result,
// and stops early when every check passes, the context ends, or exec errors
// (retry exists for "the target answered, but not what we want yet", not for a
// broken target — an exec error aborts the poll). It returns the last result
// and the last until-checks; the caller records the checks like assertions and
// fails the step if they never all passed. Keeping one loop means a future
// change to retry semantics (jitter, max-elapsed, ctx-err reporting) cannot
// silently apply to one step kind and not the other.
func pollUntil(ctx context.Context, retry *spec.Retry, st *store.Store, env assert.Env, exec func(context.Context) (*runner.Result, error)) (*runner.Result, []*assert.CheckResult, error) {
	interval, _ := time.ParseDuration(retry.Interval) // validated at load time
	until := expandAssert(st, retry.Until)

	var last *runner.Result
	var checks []*assert.CheckResult
	for attempt := 1; attempt <= retry.Times; attempt++ {
		r, err := exec(ctx)
		if err != nil {
			return nil, nil, err
		}
		last = r
		checks = assert.CheckAll(until, r, env)
		if assert.AllOK(checks) {
			break
		}
		if attempt < retry.Times && interval > 0 {
			select {
			case <-ctx.Done():
				return last, checks, nil
			case <-time.After(interval):
			}
		}
	}
	return last, checks, nil
}
