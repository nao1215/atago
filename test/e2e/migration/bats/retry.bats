#!/usr/bin/env bats
# Original Bats suite for the migration guide: polling is a hand-rolled loop
# (Bats' own BATS_TEST_RETRIES re-runs the whole test, not the command). Maps
# to the `run.retry` block in ../retry.atago.yaml, which re-runs the command
# until an `until:` assertion passes.

bats_require_minimum_version 1.5.0

@test "poll until the command reports ready" {
  for _ in 1 2 3 4 5; do
    run echo ready
    if [[ "$output" == *ready* ]]; then
      return 0
    fi
    sleep 1
  done
  return 1
}
