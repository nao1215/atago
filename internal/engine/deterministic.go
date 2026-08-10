package engine

import (
	"bytes"
	"context"
	"fmt"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/plural"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkDeterminism re-runs a command and requires the declared observables to
// come back byte-identical (#398).
//
// The first result is the one everything else in the scenario sees: assertions,
// `store`, snapshots, and the `changes:` delta all describe run 1, and the later
// runs exist only to be compared against it. Anything else would make the
// feature change what a spec means rather than add a claim to it.
//
// It returns nil when the runs agree, and a failed check otherwise — a verdict
// rather than an execution error, because "this command is nondeterministic" is
// a fact about the program under test, exactly like a failed assertion.
func checkDeterminism(ctx context.Context, d *spec.Deterministic, first *runner.Result, workdir string, exec func(context.Context) (*runner.Result, error)) (*assert.CheckResult, error) {
	total := d.DeterministicRuns()
	observables := d.Comparables()
	// selfMutating records that a rerun changed the workdir, which is the one
	// shape this check cannot judge: a command whose own side effects feed its
	// next run (a log it appends to, an input it consumes, a file it rewrites)
	// legitimately differs the second time. Seeing it lets the failure point at
	// that instead of at an iteration-order bug the author does not have.
	selfMutating := false

	for i := 2; i <= total; i++ {
		before, scanErr := fsdelta.Scan(workdir)
		again, err := exec(ctx)
		if err != nil {
			// A rerun that could not execute at all is an execution error, not a
			// verdict: the property was never tested.
			return nil, fmt.Errorf("deterministic: run %d of %d could not execute: %w", i, total, err)
		}
		if scanErr == nil {
			if after, err := fsdelta.Scan(workdir); err == nil && deltaTouched(fsdelta.Diff(before, after)) {
				selfMutating = true
			}
		}
		for _, name := range observables {
			if cr := compareObservable(name, first, again, i, total, selfMutating); cr != nil {
				return cr, nil
			}
		}
	}
	return nil, nil
}

// deltaTouched reports whether a workdir delta recorded any change at all.
func deltaTouched(d fsdelta.Delta) bool {
	return len(d.Created)+len(d.Modified)+len(d.Deleted) > 0
}

// compareObservable returns a failed check when one observable diverged between
// run 1 and run i, or nil when it matched.
func compareObservable(name string, first, again *runner.Result, i, total int, selfMutating bool) *assert.CheckResult {
	desc := fmt.Sprintf("assert deterministic: %s identical across %s", name, plural.Count(total, "run", "runs"))
	switch name {
	case spec.DeterministicExitCode:
		if first.ExitCode == again.ExitCode {
			return nil
		}
		return &assert.CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("exit code %d on every run", first.ExitCode),
			Actual:   fmt.Sprintf("exit code %d on run %d", again.ExitCode, i),
			Hint:     nondeterminismHint(name, i, selfMutating),
		}
	case spec.DeterministicStdout, spec.DeterministicStderr:
		a, b := streamOf(first, name), streamOf(again, name)
		if bytes.Equal(a, b) {
			return nil
		}
		return &assert.CheckResult{
			Desc:     desc,
			Expected: fmt.Sprintf("run 1 %s, byte for byte", name),
			Actual:   fmt.Sprintf("run %d %s differs", i, name),
			Hint:     nondeterminismHint(name, i, selfMutating),
			// Carrying both payloads is what makes the console render a unified
			// diff: a divergence inside a hundred-line table is unreadable as two
			// excerpts side by side, and the whole point is to see WHICH field
			// moved — a column order, a JSON key, an unsorted row.
			ArtifactKind:     "deterministic",
			ArtifactExpected: a,
			ArtifactActual:   b,
		}
	}
	return nil
}

func streamOf(r *runner.Result, name string) []byte {
	if name == spec.DeterministicStderr {
		return r.Stderr
	}
	return r.Stdout
}

// nondeterminismHint says what diverged and names the cause worth checking
// first. When a rerun changed the workdir, that warning comes first: the check
// is only meaningful for an effectively read-only command, and a spec that
// points it at a command which rewrites its own input would otherwise read the
// failure as a bug in the program.
func nondeterminismHint(name string, i int, selfMutating bool) string {
	base := fmt.Sprintf("%s changed between run 1 and run %d, so the command is not deterministic for this input"+
		" — the usual cause is iteration order leaking into output (a map, an unsorted listing, a set)", name, i)
	if selfMutating {
		return base + ". Note that the rerun also changed the workdir: a command whose own side effects feed its" +
			" next run is expected to differ, and deterministic: is only meaningful for an effectively read-only command"
	}
	return base
}
