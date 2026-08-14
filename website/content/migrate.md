---
toc: true
title: Migrating from Bats or ShellSpec
description: Side-by-side mappings from Bats and ShellSpec tests to atago specs. Every pair on this page is committed to the repository and both sides run in CI, so the guide cannot drift from what actually works.
---

Bats and ShellSpec are mature, well-designed frameworks, and if they fit your
suite there is no reason to leave. This guide is for one specific situation:
the thing under test is a compiled binary and the shell code around it
exists only to run that binary and inspect what happened. That layer is
atago's job, and the mappings below show what each familiar construct becomes.

Everything on this page is executable. The original Bats and ShellSpec suites
live at [test/e2e/migration](https://github.com/nao1215/atago/blob/main/test/e2e/migration/)
next to the migrated atago specs, and the
[MigrationParity workflow](https://github.com/nao1215/atago/blob/main/.github/workflows/migration.yml)
runs all three on every pull request and on every push to `main`: the
originals under pinned Bats-core v1.14.0
and ShellSpec 0.28.1, the migrations under the freshly built atago binary.
A snippet you read here is an excerpt of a file CI ran.

For a feature-by-feature table (including what Bats and ShellSpec do that
atago does not), see the [comparison](/comparison/).

## The cheat sheet

| Bats | ShellSpec | atago |
|------|-----------|-------|
| `run cmd` + `[ "$status" -eq 0 ]` / `run -N` | `When run cmd` + `The status should ...` | `run:` step + `assert: exit_code:` |
| `[[ "$output" == *x* ]]` / `assert_output --partial` | `The output should include "x"` | `stdout: {contains: x}` |
| `[ "$output" = "x" ]` / `assert_output` | `The output should equal "x"` | `stdout: {equals: x}` |
| `[[ "$output" =~ regex ]]` | `The output should match pattern` | `stdout: {matches: regex}` |
| `run --separate-stderr` + `$stderr` | `The error should include` | `stderr: {contains: ...}` |
| `setup()` / `teardown()` | `BeforeEach` / `AfterEach` | `fixture:` step; scratch-file cleanup is automatic (fresh temp workdir per scenario) |
| pipe to `jq` | pipe to `jq` | `stdout: {json: {path: ...}}` |
| one `@test` per case | `Parameters` block | `matrix:` |
| retry loop in bash | retry loop in a helper | `run.retry` with `until:` |
| `skip "reason"` | `Skip if "reason" cmd` | `skip:` / `only:` gates |
| `# bats test_tags=` + `--filter-tags` | `--tag` | `tags:` + `--tag` / `--skip-tag` |
| `diff` against a committed file | `should equal "$(cat golden)"` | `snapshot:` + `atago snapshot update` |

## Exit codes, stdout, stderr

The everyday assertions translate one to one. In Bats:

```bash
@test "an exact exit code" {
  run -3 sh -c 'exit 3'
}

@test "stdout contains" {
  run echo hello world
  [[ "$output" == *world* ]]
}
```

In ShellSpec:

```sh
  It 'an exact exit code'
    When run sh -c 'exit 3'
    The status should equal 3
  End

  It 'stdout contains'
    When run echo hello world
    The output should include "world"
  End
```

In atago, each becomes a scenario with a `run:` step and declarative asserts
([basics.atago.yaml](https://github.com/nao1215/atago/blob/main/test/e2e/migration/basics.atago.yaml)):

```yaml
  # ShellSpec: The status should equal 3
  # Bats:      [ "$status" -eq 3 ]
  - name: an exact exit code maps to exit_code N
    skip: { os: windows }        # `exit N` phrasing below is POSIX shell
    steps:
      - run:
          shell: true
          command: "exit 3"
      - assert:
          exit_code: 3

  # ShellSpec: The output should include "world"
  # Bats:      assert_output --partial world
  # runn:      test: current.res.body contains "world"
  - name: stdout contains maps to a substring match
    steps:
      - run:
          shell: true
          command: echo hello world
      - assert:
          stdout:
            contains: world
```

stderr is a separate stream. Bats needs `run --separate-stderr`; ShellSpec has
`The error`; atago always captures the two streams independently:

```bash
@test "stderr contains" {
  run --separate-stderr sh -c 'echo "warn: heads up" >&2'
  [[ "$stderr" == *warn* ]]
}
```

```yaml
  # ShellSpec: The error should include "warn"
  # Bats:      run --separate-stderr; assert_stderr --partial warn
  - name: stderr contains maps to a stderr substring match
    skip: { os: windows }        # `>&2` redirect is POSIX-only
    steps:
      - run:
          shell: true
          command: "echo warn: heads up >&2"
      - assert:
          stderr:
            contains: warn
```

## setup/teardown become fixtures — and file cleanup disappears

A Bats `setup()` writes input files and `teardown()` removes them:

```bash
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
```

ShellSpec does the same with `BeforeEach`/`AfterEach` helpers:

```sh
  seed_input() {
    mkdir -p "$work" && printf 'id,name\n1,Alice\n' > "$work/input.txt"
  }
  remove_scratch() {
    rm -rf "$work"
  }

  BeforeEach 'seed_input'
  AfterEach 'remove_scratch'

  It 'a seeded input produces the expected output file'
    When run cp "$work/input.txt" "$work/output.txt"
    The status should be success
    The file "$work/output.txt" should be exist
    The contents of file "$work/output.txt" should include "Alice"
  End
```

In atago the setup is a declarative `fixture:` step, and for scratch files
there is no teardown to write: every scenario runs in its own temp workdir
that atago removes. (`teardown:` steps do exist — they are for external state
a workdir cannot carry away, like database rows or remote resources; see
[the cookbook](/cookbook/#clean-up-external-state-even-when-a-step-fails).)
From
[fixtures_and_files.atago.yaml](https://github.com/nao1215/atago/blob/main/test/e2e/migration/fixtures_and_files.atago.yaml):

```yaml
  # BeforeEach writes input.txt -> fixture step. The command consumes it and
  # writes an output; we assert on the produced file instead of a teardown rm.
  - name: a fixture seeds input and a file assertion checks the output
    skip: { os: windows }        # cp/redirect below are POSIX-only
    steps:
      - fixture:
          file: input.txt
          content: |
            id,name
            1,Alice
      - run:
          shell: true
          command: "cp input.txt output.txt"
      - assert:
          file:
            path: output.txt
            exists: true
      - assert:
          file:
            path: output.txt
            contains: Alice
```

## jq pipelines become JSONPath matchers

Both Bats and ShellSpec suites commonly shell out to `jq`:

```bash
@test "json output asserted via jq, value captured and reused" {
  run -0 sh -c "printf '%s' '{\"items\":[{\"id\":7,\"name\":\"Alice\"}],\"count\":1}' | jq -r '.items[0].name'"
  [ "$output" = "Alice" ]
  first_id="$(printf '%s' '{"items":[{"id":7,"name":"Alice"}],"count":1}' | jq -r '.items[0].id')"
  run -0 echo "picked $first_id"
  [[ "$output" == *"picked 7"* ]]
}
```

atago asserts JSON by path without an external tool, and `store:` replaces the
`var=$(... | jq)` capture:

```yaml
      - assert:
          stdout:
            json:
              path: "$.items[0].name"
              equals: Alice
      - assert:
          stdout:
            json:
              path: "$.count"
              gte: 1
      # ShellSpec/Bats would capture with `id=$(... | jq)`. atago's `store`
      # binds a captured value into ${name} for a later step — declaratively.
      - store:
          name: first_id
          from:
            stdout:
              json:
                path: "$.items[0].id"
      - run:
          shell: true
          command: echo "picked ${first_id}"
      - assert:
          stdout:
            contains: "picked 7"
```

## Parameterized tests become a matrix

Bats has no built-in parameterization, so the same test is written once per
case (or generated with `bats_test_function`):

```bash
@test "greets Alice in en" {
  run -0 echo "hello Alice (en)"
  [[ "$output" == *Alice* && "$output" == *en* ]]
}

@test "greets Bob in fr" {
  run -0 echo "hello Bob (fr)"
  [[ "$output" == *Bob* && "$output" == *fr* ]]
}
```

ShellSpec's `Parameters` block — a genuinely nice feature — feeds positional
parameters into one example:

```sh
  Parameters
    Alice en
    Bob   fr
  End

  It "greets $1 in $2"
    When run echo "hello $1 ($2)"
    The output should include "$1"
    The output should include "$2"
  End
```

atago's `matrix:` expands one template scenario into one concrete scenario per
row, each key available as `${name}`
([matrix.atago.yaml](https://github.com/nao1215/atago/blob/main/test/e2e/migration/matrix.atago.yaml)):

```yaml
  - name: "greets ${who} in ${lang}"
    matrix:
      - { who: Alice, lang: en }
      - { who: Bob,   lang: fr }
    steps:
      - run:
          shell: true
          command: echo "hello ${who} (${lang})"
      - assert:
          stdout:
            contains:
              - ${who}
              - ${lang}
```

## Polling loops become retry

Waiting for an async result is a hand-written loop in shell. The probed
command here is stateful, like a service warming up — the first attempt
creates a marker and prints "waiting", the second prints "ready" — so the
loop genuinely iterates:

```bash
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
```

atago re-runs the same stateful command until an `until:` assertion passes,
with the attempt budget and interval declared — the first attempt prints
"waiting", the retry sees "ready"
([retry.atago.yaml](https://github.com/nao1215/atago/blob/main/test/e2e/migration/retry.atago.yaml)):

```yaml
  - name: retry re-runs the command until until passes
    skip: { os: windows }        # the marker-file idiom below is POSIX shell
    steps:
      - run:
          shell: true
          command: "if [ -f marker ]; then echo ready; else touch marker; echo waiting; fi"
          retry:
            times: 5
            interval: 10ms
            until:
              stdout:
                contains: ready
      - assert:
          stdout:
            contains: ready
```

(Bats' own `BATS_TEST_RETRIES` re-runs the whole test on failure — useful, but
a different tool: it is for flaky tests, not for polling.)

## skip and tags become declarative gates

Bats skips imperatively and tags with comments:

```bash
# bats test_tags=smoke
@test "smoke-tagged, selectable with --filter-tags smoke" {
  run -0 echo smoke ok
  [[ "$output" == *ok* ]]
}

# bats test_tags=slow
@test "slow-tagged, droppable with --filter-tags !slow" {
  run -0 echo slow but still green
}
```

ShellSpec's `Skip if` takes a reason and a probe command:

```sh
  It 'runs only when an environment variable is set'
    Skip if "CI_ONLY_UNSET_MARKER not set" [ -z "${CI_ONLY_UNSET_MARKER:-}" ]
    When run echo would only run in CI
    The status should be success
  End
```

atago's gates are data, so `atago explain` and `atago list` can show them
without running anything
([selection.atago.yaml](https://github.com/nao1215/atago/blob/main/test/e2e/migration/selection.atago.yaml)):

```yaml
  # ShellSpec: Skip if "not linux" ... ; Bats: skip on non-linux
  # atago: skip: { os: windows } declaratively skips POSIX-only scenarios.
  - name: skip on an OS maps to skip os
    skip: { os: windows }
    steps:
      - run:
          shell: true
          command: "test -d /"        # POSIX filesystem probe
      - assert:
          exit_code: 0

  # ShellSpec: fIt / Bats: bats --filter-tags smoke
  # atago: tag a scenario, then `atago run --tag smoke` to focus on it.
  - name: a smoke-tagged scenario is selectable with --tag
    tags: [smoke]
    steps:
      - run:
          shell: true
          command: echo smoke ok
      - assert:
          stdout:
            contains: ok
```

The CI parity workflow runs the tag filters too: `bats --filter-tags smoke`
and `atago run --tag smoke` select the same test on both sides.

## Hand-maintained goldens become snapshots

Neither Bats nor ShellSpec ships snapshot testing, so a golden compare is
written by hand and refreshed by editing the file:

```bash
@test "stdout matches the committed golden file" {
  run -0 echo "hello from atago"
  diff <(printf '%s\n' "$output") "$BATS_TEST_DIRNAME/../snapshots/greeting.txt"
}
```

```sh
  It 'stdout matches the committed golden file'
    When run echo "hello from atago"
    The output should equal "$(cat "$SHELLSPEC_PROJECT_ROOT/../snapshots/greeting.txt")"
  End
```

atago's `snapshot:` matcher compares against the same kind of committed file,
but normalizes ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF so
the golden is stable across machines — and `atago snapshot update` refreshes
every intentionally changed golden in one command
([snapshot.atago.yaml](https://github.com/nao1215/atago/blob/main/test/e2e/migration/snapshot.atago.yaml)):

```yaml
  - name: stdout matches a committed snapshot
    steps:
      - run:
          shell: true
          command: echo hello from atago
      - assert:
          stdout:
            snapshot: snapshots/greeting.txt
```

## What has no source to migrate from

Some atago features have no Bats/ShellSpec counterpart, so after the port they
are simply new capability: `pty:` sessions for interactive prompts and TUIs,
`expect_screen:`/`screen:` asserts on the rendered terminal frame, `changes:`
for the exact set of files a command touched, `mock_servers:` for offline API
simulation, `services:` with readiness probes, and `atago record` to turn a
real run into a first spec. The [cookbook](/cookbook/) has a runnable recipe
for each.

## What should stay where it is

Migrate the black-box layer, not everything. Bats and ShellSpec remain the
right tool for what they were built for:

- Unit tests of shell functions. Both frameworks run shell code
  in-process and assert on it directly. atago only ever executes a program,
  so it cannot see inside your shell scripts.
- Mocking shell internals. ShellSpec's function and command mocks have no
  atago equivalent — atago's mocks are external HTTP servers, a different
  axis.
- Coverage of shell scripts. ShellSpec measures kcov coverage for shell
  code; a black-box runner cannot.

If your repository has both kinds of tests, keeping ShellSpec for the shell
library and moving the CLI contract tests to atago is a reasonable end state —
it is exactly how atago's own third-party suites were migrated.

## Migrate incrementally

Both runners coexist happily in one CI job, so port one file at a time:

```shell
bats test/                       # the suite you are migrating from
atago run ./specs                # the part that has moved
```

Start by recording a real run (`atago record -- mytool ...`) instead of
writing YAML from scratch, then tighten the generated spec with the mappings
above. [Use it in CI](/ci/) shows the GitHub Actions wiring, including
`--report junit` if your CI already consumes Bats' JUnit output.
