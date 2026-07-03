package engine

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// flakyOnceSpec fails while a marker file is absent, creates it, and passes
// once it exists — deterministic flakiness in a shared scratch dir. Each
// attempt runs in a fresh workdir, so the marker MUST live outside it.
func flakyOnceSpec(t *testing.T) string {
	t.Helper()
	marker := filepath.ToSlash(filepath.Join(t.TempDir(), "seen.txt"))
	return `
version: "1"
suite:
  name: s
scenarios:
  - name: flaky once
    steps:
      - run:
          shell: true
          command: "if [ -f '` + marker + `' ]; then echo recovered; else touch '` + marker + `'; exit 1; fi"
      - assert:
          exit_code: 0
`
}

// TestEngine_RetryFailedRecoversAsFlaky proves the retry loop (#29): a
// fail-then-pass scenario ends flaky with Attempts=2, the suite verdict stays
// green, and fail-fast is not tripped.
func TestEngine_RetryFailedRecoversAsFlaky(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	s, err := loader.LoadBytes("t.atago.yaml", []byte(flakyOnceSpec(t)))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.RetryFailed = 2
	eng.FailFast = true
	res := eng.Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusPassed {
		t.Fatalf("suite status = %s, want passed (flaky counts green): %+v", res.Status, res.Scenarios)
	}
	sc := res.Scenarios[0]
	if sc.Status != StatusFlaky {
		t.Fatalf("scenario status = %s, want flaky", sc.Status)
	}
	if sc.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", sc.Attempts)
	}
	if c := res.Counts(); c.Flaky != 1 || c.Failed != 0 {
		t.Errorf("counts = %+v, want 1 flaky, 0 failed", c)
	}
}

// TestEngine_RetryFailedExhaustedStaysFailed proves exhausted retries keep the
// failure with the attempt count recorded.
func TestEngine_RetryFailedExhaustedStaysFailed(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	s, err := loader.LoadBytes("t.atago.yaml", []byte(`
version: "1"
suite:
  name: s
scenarios:
  - name: always fails
    steps:
      - run: {shell: true, command: exit 1}
      - assert:
          exit_code: 0
`))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.RetryFailed = 2
	res := eng.Run(context.Background(), s, "t.atago.yaml")
	if res.Status != StatusFailed {
		t.Fatalf("suite status = %s, want failed", res.Status)
	}
	if sc := res.Scenarios[0]; sc.Status != StatusFailed || sc.Attempts != 3 {
		t.Errorf("scenario = %s attempts=%d, want failed after 3 attempts", sc.Status, sc.Attempts)
	}
}

// TestEngine_RepeatFoldsIterations proves --repeat (#29): every iteration
// runs, per-iteration statuses are recorded, and any failing iteration fails
// the fold while an all-green repeat passes.
func TestEngine_RepeatFoldsIterations(t *testing.T) {
	skipOnWindows(t)
	t.Parallel()
	green, err := loader.LoadBytes("t.atago.yaml", []byte(`
version: "1"
suite:
  name: s
scenarios:
  - name: steady
    steps:
      - run: {shell: true, command: echo ok}
      - assert:
          exit_code: 0
`))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng := New()
	eng.Repeat = 3
	res := eng.Run(context.Background(), green, "t.atago.yaml")
	if res.Status != StatusPassed {
		t.Fatalf("suite status = %s, want passed", res.Status)
	}
	if got := res.Scenarios[0].Iterations; len(got) != 3 {
		t.Fatalf("iterations = %v, want 3 entries", got)
	}

	// The flaky-once scenario passes on its second execution — under repeat
	// that is still a FAILURE (any iteration failing fails the run).
	flaky, err := loader.LoadBytes("t.atago.yaml", []byte(flakyOnceSpec(t)))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	eng2 := New()
	eng2.Repeat = 3
	res2 := eng2.Run(context.Background(), flaky, "t.atago.yaml")
	if res2.Status != StatusFailed {
		t.Fatalf("suite status = %s, want failed (one bad iteration taints the fold)", res2.Status)
	}
	iters := res2.Scenarios[0].Iterations
	if len(iters) != 3 || iters[0] != StatusFailed || iters[1] != StatusPassed {
		t.Errorf("iterations = %v, want [failed passed passed]", iters)
	}
	// The kept steps come from the first failing iteration.
	var joined strings.Builder
	for _, st := range res2.Scenarios[0].Steps {
		if st.Run != nil {
			joined.Write(st.Run.Stdout)
		}
	}
	if strings.Contains(joined.String(), "recovered") {
		t.Errorf("kept steps should belong to the failing iteration, got %q", joined.String())
	}
}
