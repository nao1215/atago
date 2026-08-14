#!/usr/bin/env bats
# Original Bats suite for the migration guide: setup()/teardown() writing
# scratch files, and JSON checked by piping through jq. Maps 1:1 to
# ../fixtures_and_files.atago.yaml, where the same tests become a declarative
# `fixture:` step (no teardown: every atago scenario gets a fresh temp workdir)
# and JSONPath matchers (no jq dependency).

bats_require_minimum_version 1.5.0

setup() {
  cd "$BATS_TEST_TMPDIR"
  printf 'id,name\n1,Alice\n' > input.txt
}

teardown() {
  rm -f input.txt output.txt
}

@test "a seeded input produces the expected output file" {
  run -0 cp input.txt output.txt
  [ -f output.txt ]
  grep -q Alice output.txt
}

@test "json output asserted via jq, value captured and reused" {
  run -0 sh -c "printf '%s' '{\"items\":[{\"id\":7,\"name\":\"Alice\"}],\"count\":1}' | jq -r '.items[0].name'"
  [ "$output" = "Alice" ]
  first_id="$(printf '%s' '{"items":[{"id":7,"name":"Alice"}],"count":1}' | jq -r '.items[0].id')"
  run -0 echo "picked $first_id"
  [[ "$output" == *"picked 7"* ]]
}
