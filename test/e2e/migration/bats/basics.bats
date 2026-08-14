#!/usr/bin/env bats
# Original Bats suite for the migration guide. Each @test here maps 1:1 to a
# scenario in ../basics.atago.yaml; CI runs both (this file under a pinned
# Bats, the atago spec under the freshly built binary) so the mapping the
# website shows is executed, not asserted from memory.

# `run -N`, `run !`, and --separate-stderr need Bats >= 1.5.0.
bats_require_minimum_version 1.5.0

@test "exit code success" {
  run -0 echo hello
}

@test "exit code failure" {
  run ! false
}

@test "an exact exit code" {
  run -3 sh -c 'exit 3'
}

@test "stdout contains" {
  run echo hello world
  [[ "$output" == *world* ]]
}

@test "stdout equals" {
  run echo exact
  [ "$output" = "exact" ]
}

@test "stdout does not contain" {
  run echo all good
  [[ "$output" != *panic* ]]
}

@test "stdout matches a regex" {
  run echo v1.2.3
  [[ "$output" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]
}

@test "stderr contains" {
  run --separate-stderr sh -c 'echo "warn: heads up" >&2'
  [[ "$stderr" == *warn* ]]
}

@test "stderr is empty" {
  run --separate-stderr echo quiet
  [ -z "$stderr" ]
}
