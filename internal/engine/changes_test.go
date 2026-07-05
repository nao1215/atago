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
