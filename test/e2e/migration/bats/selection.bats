#!/usr/bin/env bats
# Original Bats suite for the migration guide: `skip` calls and
# `# bats test_tags=` comments (Bats >= 1.8.0), selected with
# `bats --filter-tags`. Maps 1:1 to ../selection.atago.yaml, where the same
# intent becomes declarative `skip:`/`only:` gates and `tags:` selected with
# `atago run --tag` / `--skip-tag`.

bats_require_minimum_version 1.8.0

@test "runs only where a POSIX filesystem exists" {
  if [ "$(uname)" = "Windows_NT" ]; then
    skip "POSIX-only"
  fi
  run -0 test -d /
}

# bats test_tags=smoke
@test "smoke-tagged, selectable with --filter-tags smoke" {
  run -0 echo smoke ok
  [[ "$output" == *ok* ]]
}

# bats test_tags=slow
@test "slow-tagged, droppable with --filter-tags !slow" {
  run -0 echo slow but still green
}

@test "runs only when an environment variable is set" {
  if [ -z "${CI_ONLY_UNSET_MARKER:-}" ]; then
    skip "CI_ONLY_UNSET_MARKER not set"
  fi
  run -0 echo would only run in CI
}

@test "skip decided by a probe command" {
  if false; then
    skip "probe matched"
  fi
  run -0 echo probe did not match, so we run
}
