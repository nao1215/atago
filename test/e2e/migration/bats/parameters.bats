#!/usr/bin/env bats
# Original Bats suite for the migration guide: Bats has no built-in
# parameterization, so the same test is written once per case (the documented
# alternative is generating tests with bats_test_function). Maps to the single
# `matrix:` scenario in ../matrix.atago.yaml, which expands one template into
# one scenario per row.

bats_require_minimum_version 1.5.0

@test "greets Alice in en" {
  run -0 echo "hello Alice (en)"
  [[ "$output" == *Alice* && "$output" == *en* ]]
}

@test "greets Bob in fr" {
  run -0 echo "hello Bob (fr)"
  [[ "$output" == *Bob* && "$output" == *fr* ]]
}
