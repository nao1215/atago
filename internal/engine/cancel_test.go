package engine

import (
	"context"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/loader"
)

// TestRun_CanceledBeforeStart stops the run before any scenario executes: an
// already-canceled context must not let a long-running command run to
// completion, and the suite reports promptly.
func TestRun_CanceledBeforeStart(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: cancel
scenarios:
  - name: slow
    steps:
      - run: {shell: true, command: "` + sleepCmd(30) + `"}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before Run starts

	start := time.Now()
	res := New().Run(ctx, s, "t.atago.yaml")
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("canceled run took %v; expected it to stop promptly, not run the 30s sleep", elapsed)
	}
	// The scenario must not have run to a passing completion: it is either skipped
	// (never scheduled) or errored (bailed at its first step).
	if got := res.Scenarios[0].Status; got != StatusSkipped && got != StatusError {
		t.Fatalf("scenario status = %q, want skipped or error under pre-cancellation", got)
	}
	if res.Scenarios[0].Status == StatusSkipped && res.Scenarios[0].SkipReason != "skipped after interrupt" {
		t.Errorf("skip reason = %q, want 'skipped after interrupt'", res.Scenarios[0].SkipReason)
	}
}

// TestRun_CanceledDuringExecution cancels while a long step is in flight and
// asserts the run unwinds well before the command's natural completion.
func TestRun_CanceledDuringExecution(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: cancel
scenarios:
  - name: a
    steps:
      - run: {shell: true, command: "echo one"}
      - run: {shell: true, command: "` + sleepCmd(30) + `"}
      - run: {shell: true, command: "echo three"}
`
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	res := New().Run(ctx, s, "t.atago.yaml")
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("run did not stop on cancellation: took %v", elapsed)
	}
	if res.Status == StatusPassed {
		t.Fatalf("run reported passed despite mid-run cancellation: %+v", res.Scenarios)
	}
}
