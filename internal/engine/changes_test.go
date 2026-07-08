package engine

import (
	"fmt"
	"runtime"
	"testing"
)

// mutateWorkdirCmd returns a shell command that modifies change.txt, deletes
// gone.txt, and creates a nested out/new.txt — the multi-file delta the changes
// assertion below pins. The command is the only OS-specific part: atago
// normalizes recorded paths to forward slashes, so the assertion is portable.
func mutateWorkdirCmd() string {
	if runtime.GOOS == "windows" {
		return `echo after >change.txt& del gone.txt& mkdir out& echo hi >out\new.txt`
	}
	return "echo after > change.txt && rm gone.txt && mkdir -p out && echo hi > out/new.txt"
}

// TestEngine_Changes_CreatedModifiedDeleted proves the delta assertion pins
// exactly what a run step created, modified, and deleted in the workdir (#70).
func TestEngine_Changes_CreatedModifiedDeleted(t *testing.T) {
	t.Parallel()
	res := runSpec(t, fmt.Sprintf(`
version: "1"
suite:
  name: changes
scenarios:
  - name: a generator touches only its files
    steps:
      # Seed inputs as fixtures — these are inputs, not changes, because they
      # land before the measured step's pre-scan.
      - fixture:
          file: keep.txt
          content: same
      - fixture:
          file: change.txt
          content: before
      - fixture:
          file: gone.txt
          content: bye
      - run:
          shell: true
          command: '%s'
      - assert:
          exit_code: 0
          changes:
            created:
              - out/new.txt
            modified:
              - change.txt
            deleted:
              - gone.txt
`, mutateWorkdirCmd()))
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_Changes_FixtureWritesAreInputs proves a fixture written before the
// measured step is NOT counted as a change (its content is part of the
// baseline), while stdout_to created by the step IS counted as created (#70).
func TestEngine_Changes_FixtureWritesAreInputs(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: changes
scenarios:
  - name: only the step's own output counts
    steps:
      - fixture:
          file: input.txt
          content: seed
      - run:
          shell: true
          command: echo produced
          stdout_to: result.txt
      - assert:
          exit_code: 0
          changes:
            created:
              - result.txt
            modified: []
            deleted: []
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_Changes_NoScanWithoutAssert proves the pre-scan is skipped when no
// changes assert follows: a run step with no following changes assert leaves
// current.Changes nil, so a later (validly preceded) changes assert that DOES
// follow its own step still works — i.e. the scan is scoped per step. Here the
// first run has no changes assert; the assertion after the SECOND run must see
// only the second step's delta.
func TestEngine_Changes_NoScanWithoutAssert(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: changes
scenarios:
  - name: delta is scoped to the immediately preceding step
    steps:
      - run:
          shell: true
          command: echo one
          stdout_to: first.txt
      - run:
          shell: true
          command: echo two
          stdout_to: second.txt
      - assert:
          changes:
            created:
              - second.txt
            modified: []
            deleted: []
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_Changes_RetryReflectsLastAttempt is the regression for #251: a
// `changes:` assert after a retried run step must pin the delta of the final
// (converged) attempt, not the cumulative delta of every attempt. Here attempt
// 1 creates a and b but exits 1 (c does not exist yet); attempt 2 finds b
// present, creates c, and passes. Before the fix the baseline was scanned once
// before attempt 1, so the delta spanned a, b, and c and an author asking for
// the converged net effect (c only) saw a spurious "unexpected created file
// a/b" failure. With the fix the baseline is re-taken before attempt 2, so a
// and b are inputs and only c is reported.
func TestEngine_Changes_RetryReflectsLastAttempt(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("uses POSIX shell test/touch to make each retry attempt write its own file")
	}
	res := runSpec(t, `
version: "1"
suite:
  name: retry-delta
scenarios:
  - name: converged attempt delta
    steps:
      - run:
          shell: true
          command: 'touch a; [ -f b ] && touch c; touch b; [ -f c ]'
          retry:
            times: 5
            interval: 5ms
            until:
              exit_code: 0
      - assert:
          exit_code: 0
          changes:
            created:
              - c
            modified: []
            deleted: []
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed (changes must reflect the converged last attempt, not the cumulative delta): %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// TestEngine_Changes_UnexpectedFileFails proves an unexpected created file
// fails the assertion (the exhaustive contract) (#70).
func TestEngine_Changes_UnexpectedFileFails(t *testing.T) {
	t.Parallel()
	create := "echo a > a.txt && echo b > b.txt"
	if runtime.GOOS == "windows" {
		create = `echo a >a.txt& echo b >b.txt`
	}
	res := runSpec(t, fmt.Sprintf(`
version: "1"
suite:
  name: changes
scenarios:
  - name: an extra file breaks the exact contract
    steps:
      - run:
          shell: true
          command: '%s'
      - assert:
          changes:
            created:
              - a.txt
`, create))
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed (b.txt is an unexpected creation): %+v", res.Status, res.Scenarios[0].Steps)
	}
}
