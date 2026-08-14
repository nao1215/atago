#!/usr/bin/env bats
# Original Bats suite for the migration guide: Bats has no snapshot testing,
# so the golden compare is hand-rolled with diff against a committed file —
# and refreshing it after an intended change is a manual edit. Maps to
# ../snapshot.atago.yaml, where the same golden is a `snapshot:` matcher with
# normalization and `atago snapshot update`.

bats_require_minimum_version 1.5.0

@test "stdout matches the committed golden file" {
  run -0 echo "hello from atago"
  diff <(printf '%s\n' "$output") "$BATS_TEST_DIRNAME/../snapshots/greeting.txt"
}
