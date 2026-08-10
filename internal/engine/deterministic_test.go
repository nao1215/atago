package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// scriptedExec returns an exec function that hands back the given results in
// order, so a determinism check can be driven without a real command.
func scriptedExec(results ...*runner.Result) func(context.Context) (*runner.Result, error) {
	i := 0
	return func(context.Context) (*runner.Result, error) {
		r := results[min(i, len(results)-1)]
		i++
		return r, nil
	}
}

func out(s string) *runner.Result { return &runner.Result{Stdout: []byte(s)} }

// writeFile is the workdir side effect a self-mutating command would have.
func writeFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)
}

// TestCheckDeterminism_AgreeingRunsPass covers the green path: identical output
// across the reruns is the property holding, and nothing is reported.
func TestCheckDeterminism_AgreeingRunsPass(t *testing.T) {
	t.Parallel()
	first := out("a,b,c\n")
	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{}, first,
		t.TempDir(), scriptedExec(out("a,b,c\n")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr != nil {
		t.Errorf("identical runs must not fail: %+v", cr)
	}
}

// TestCheckDeterminism_DivergingStdoutFails is the bug this exists for: output
// that changes between runs while every loose matcher keeps passing — the shape
// a column order leaking out of a map iteration takes.
func TestCheckDeterminism_DivergingStdoutFails(t *testing.T) {
	t.Parallel()
	first := out("id,name\n1,alice\n")
	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{}, first,
		t.TempDir(), scriptedExec(out("name,id\nalice,1\n")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr == nil || cr.OK {
		t.Fatal("diverging stdout must fail the step")
	}
	if !strings.Contains(cr.Hint, "not deterministic") {
		t.Errorf("hint %q should say the command is not deterministic", cr.Hint)
	}
	// Both payloads have to be carried, or the console cannot render the diff
	// that shows WHICH field moved.
	if string(cr.ArtifactExpected) != string(first.Stdout) || len(cr.ArtifactActual) == 0 {
		t.Errorf("both run payloads must be attached for the diff (%q / %q)", cr.ArtifactExpected, cr.ArtifactActual)
	}
	if cr.ArtifactKind != "deterministic" {
		t.Errorf("ArtifactKind = %q, want deterministic", cr.ArtifactKind)
	}
}

// TestCheckDeterminism_ExitCodeDivergence pins the other default observable.
func TestCheckDeterminism_ExitCodeDivergence(t *testing.T) {
	t.Parallel()
	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{},
		&runner.Result{ExitCode: 0}, t.TempDir(),
		scriptedExec(&runner.Result{ExitCode: 1}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr == nil || cr.OK {
		t.Fatal("a changing exit code must fail the step")
	}
	if !strings.Contains(cr.Desc, "exit_code") {
		t.Errorf("desc %q should name the observable", cr.Desc)
	}
}

// TestCheckDeterminism_CompareSelection proves `compare` decides what is
// checked: stderr is opt-in, because progress output legitimately carries
// timings, and stdout can be dropped when only the exit code is the contract.
func TestCheckDeterminism_CompareSelection(t *testing.T) {
	t.Parallel()
	first := &runner.Result{Stdout: []byte("same"), Stderr: []byte("took 12ms")}
	later := &runner.Result{Stdout: []byte("same"), Stderr: []byte("took 31ms")}

	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{}, first, t.TempDir(), scriptedExec(later))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr != nil {
		t.Errorf("stderr must not be compared by default: %+v", cr)
	}

	cr, err = checkDeterminism(context.Background(),
		&spec.Deterministic{Compare: []string{spec.DeterministicStderr}}, first, t.TempDir(), scriptedExec(later))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr == nil || cr.OK {
		t.Fatal("stderr must be compared once it is listed")
	}
}

// TestCheckDeterminism_RunsCountsTotalExecutions pins that `runs: 3` means
// three executions in total — one already done plus two reruns — and that a
// divergence on the LAST one is still caught.
func TestCheckDeterminism_RunsCountsTotalExecutions(t *testing.T) {
	t.Parallel()
	calls := 0
	exec := func(context.Context) (*runner.Result, error) {
		calls++
		if calls == 2 { // the third execution overall
			return out("different"), nil
		}
		return out("same"), nil
	}
	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{Runs: 3}, out("same"), t.TempDir(), exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("runs: 3 must rerun twice after the first execution, got %d reruns", calls)
	}
	if cr == nil || cr.OK {
		t.Fatal("a divergence on the last run must still fail")
	}
	if !strings.Contains(cr.Actual, "run 3") {
		t.Errorf("actual %q should name the run that diverged", cr.Actual)
	}
}

// TestCheckDeterminism_SelfMutatingCommandIsCalledOut covers the honest
// caveat: the check is only sound for an effectively read-only command, and a
// spec pointed at one that rewrites its own input has to be told that rather
// than sent hunting for a map-order bug.
func TestCheckDeterminism_SelfMutatingCommandIsCalledOut(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	calls := 0
	exec := func(context.Context) (*runner.Result, error) {
		calls++
		// The rerun appends to the workdir, exactly as a command that logs would.
		if err := writeFile(dir, "log.txt", strings.Repeat("x", calls)); err != nil {
			return nil, err
		}
		return out("run " + strings.Repeat("!", calls)), nil
	}
	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{}, out("first"), dir, exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cr == nil {
		t.Fatal("diverging output must still fail")
	}
	if !strings.Contains(cr.Hint, "changed the workdir") {
		t.Errorf("hint %q should point at the command's own side effects", cr.Hint)
	}
}

// TestCheckDeterminism_RerunErrorIsExecutionError pins the split: a rerun that
// could not execute never tested the property, so it is an error rather than a
// verdict about the program.
func TestCheckDeterminism_RerunErrorIsExecutionError(t *testing.T) {
	t.Parallel()
	cr, err := checkDeterminism(context.Background(), &spec.Deterministic{}, out("x"), t.TempDir(),
		func(context.Context) (*runner.Result, error) { return nil, context.DeadlineExceeded })
	if err == nil {
		t.Fatal("a rerun that could not execute must be an execution error")
	}
	if cr != nil {
		t.Errorf("no verdict may be reported when the property was never tested: %+v", cr)
	}
	if !strings.Contains(err.Error(), "run 2 of 2") {
		t.Errorf("error %q should name which rerun failed", err)
	}
}
