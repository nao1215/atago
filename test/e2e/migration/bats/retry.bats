#!/usr/bin/env bats
# Original Bats suite for the migration guide: polling is a hand-rolled loop
# (Bats' own BATS_TEST_RETRIES re-runs the whole test, not the command). The
# probed command is stateful, like a service warming up: the first attempt
# creates a marker and prints "waiting", the second prints "ready" — so the
# loop genuinely iterates. Maps to the `run.retry` block in
# ../retry.atago.yaml, which re-runs the command until an `until:` assertion
# passes.

bats_require_minimum_version 1.5.0

@test "poll until the command reports ready" {
  marker="$BATS_TEST_TMPDIR/marker"
  for _ in 1 2 3 4 5; do
    run sh -c "if [ -f '$marker' ]; then echo ready; else touch '$marker'; echo waiting; fi"
    if [[ "$output" == *ready* ]]; then
      return 0
    fi
    sleep 1
  done
  return 1
}
