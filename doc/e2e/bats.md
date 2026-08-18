# atago Behavior Specs
## Summary
4 suites · 43 scenarios
## Contents
- [bats (the verdict a test runner reports)](#bats-the-verdict-a-test-runner-reports) — 14 scenarios
  - [the version and the help text are answered on stdout](#scenario-the-version-and-the-help-text-are-answered-on-stdout)
  - [a passing file is TAP on stdout, exit 0, and no droppings](#scenario-a-passing-file-is-tap-on-stdout-exit-0-and-no-droppings)
  - [a failing test is exit 1 and names the file, the line, and the command](#scenario-a-failing-test-is-exit-1-and-names-the-file-the-line-and-the-command)
  - [one failure does not hide the tests that passed](#scenario-one-failure-does-not-hide-the-tests-that-passed)
  - [a command failing mid-test aborts the rest of the test body](#scenario-a-command-failing-mid-test-aborts-the-rest-of-the-test-body)
  - [skipped tests are ok and keep the run green](#scenario-skipped-tests-are-ok-and-keep-the-run-green)
  - [nothing to run is not an error](#scenario-nothing-to-run-is-not-an-error)
  - [a test file that does not exist reports to stderr and leaves stdout empty](#scenario-a-test-file-that-does-not-exist-reports-to-stderr-and-leaves-stdout-empty)
  - [an unknown option is rejected before any test runs](#scenario-an-unknown-option-is-rejected-before-any-test-runs)
  - [no arguments is a usage error, not an empty run](#scenario-no-arguments-is-a-usage-error-not-an-empty-run)
  - [a file Bash cannot parse fails as the gather-tests pseudo test](#scenario-a-file-bash-cannot-parse-fails-as-the-gather-tests-pseudo-test)
  - [a minimum-version guard the runner cannot meet stops the file](#scenario-a-minimum-version-guard-the-runner-cannot-meet-stops-the-file)
  - [UTF-8 test names survive into the report](#scenario-utf-8-test-names-survive-into-the-report)
  - [several files on one command line are one numbered run](#scenario-several-files-on-one-command-line-are-one-numbered-run)
- [bats (the harness around each test)](#bats-the-harness-around-each-test) — 10 scenarios
  - [setup_file runs once, setup and teardown run around every test](#scenario-setup_file-runs-once-setup-and-teardown-run-around-every-test)
  - [a failing setup keeps the body from running and teardown still runs](#scenario-a-failing-setup-keeps-the-body-from-running-and-teardown-still-runs)
  - [teardown still runs when the test itself fails](#scenario-teardown-still-runs-when-the-test-itself-fails)
  - [the run helper captures the status, the output, and the lines](#scenario-the-run-helper-captures-the-status-the-output-and-the-lines)
  - [an unmet expected status fails the test with both codes named](#scenario-an-unmet-expected-status-fails-the-test-with-both-codes-named)
  - [using run flags without declaring the minimum version warns on stderr](#scenario-using-run-flags-without-declaring-the-minimum-version-warns-on-stderr)
  - [each test gets its own temporary directory, removed when the run ends](#scenario-each-test-gets-its-own-temporary-directory-removed-when-the-run-ends)
  - [a test can read its own name, number, and file](#scenario-a-test-can-read-its-own-name-number-and-file)
  - [load pulls in a helper file next to the test](#scenario-load-pulls-in-a-helper-file-next-to-the-test)
  - [parallel jobs without the parallel binary fail as a raw shell error](#scenario-parallel-jobs-without-the-parallel-binary-fail-as-a-raw-shell-error)
- [bats (formatters, reports, and gathered output)](#bats-formatters-reports-and-gathered-output) — 10 scenarios
  - [TAP version 13 carries the failure as a YAML block](#scenario-tap-version-13-carries-the-failure-as-a-yaml-block)
  - [the JUnit formatter reports counts and the failure element](#scenario-the-junit-formatter-reports-counts-and-the-failure-element)
  - [a report formatter writes a file and leaves stdout as TAP](#scenario-a-report-formatter-writes-a-file-and-leaves-stdout-as-tap)
  - [a report directory that does not exist is refused before the run](#scenario-a-report-directory-that-does-not-exist-is-refused-before-the-run)
  - [gathered output is one file per test, named after it, holding what it printed](#scenario-gathered-output-is-one-file-per-test-named-after-it-holding-what-it-printed)
  - [gathering into a directory that already has files is refused](#scenario-gathering-into-a-directory-that-already-has-files-is-refused)
  - [test output is hidden when it passes and shown when it fails](#scenario-test-output-is-hidden-when-it-passes-and-shown-when-it-fails)
  - [timing and tracing add to the report without changing the verdict](#scenario-timing-and-tracing-add-to-the-report-without-changing-the-verdict)
  - [the quoting and file reference in a failure are configurable](#scenario-the-quoting-and-file-reference-in-a-failure-are-configurable)
  - [the pretty formatter exists only on a terminal](#scenario-the-pretty-formatter-exists-only-on-a-terminal)
- [bats (choosing which tests run)](#bats-choosing-which-tests-run) — 9 scenarios
  - [counting reports the number of tests and runs none of them](#scenario-counting-reports-the-number-of-tests-and-runs-none-of-them)
  - [the name filter is a regular expression, not a substring](#scenario-the-name-filter-is-a-regular-expression-not-a-substring)
  - [a filter that matches nothing is an empty plan, not a failure](#scenario-a-filter-that-matches-nothing-is-an-empty-plan-not-a-failure)
  - [tags select and deselect tests, and file tags apply to the whole file](#scenario-tags-select-and-deselect-tests-and-file-tags-apply-to-the-whole-file)
  - [rerunning failures requires a run-log directory the user must create](#scenario-rerunning-failures-requires-a-run-log-directory-the-user-must-create)
  - [with the run-log directory in place a run records it and the next run replays only the failures](#scenario-with-the-run-log-directory-in-place-a-run-records-it-and-the-next-run-replays-only-the-failures)
  - [replaying with nothing to replay says so and stays green](#scenario-replaying-with-nothing-to-replay-says-so-and-stays-green)
  - [a test the ledger has never seen counts as missed](#scenario-a-test-the-ledger-has-never-seen-counts-as-missed)
  - [a directory is walked flat unless recursion is asked for](#scenario-a-directory-is-walked-flat-unless-recursion-is-asked-for)

## bats (the verdict a test runner reports)
[Bats](https://github.com/bats-core/bats-core) runs test files written in
Bash and reports the result as TAP. Its own test suite is written in Bats,
so the runner is checked by the thing it runs; the contract those tests pin
is reproduced here from the outside, with atago watching the process rather
than Bash watching itself.

A test runner has one promise: the verdict it reports is the truth about
the tests, and the exit code carries that verdict to whatever ran it. This
file pins the verdict — the TAP plan and the ok/not ok lines, exit 0 for a
green run and exit 1 for a red one, which stream each kind of message goes
to, and the failures a runner meets before any test runs (no arguments, a
file that is not there, a file Bash cannot parse).

Every `.bats` file here is written by the scenario that runs it. No upstream
fixture is copied, and no test file is committed.

Source: `test/e2e/thirdparty/bats/bats.atago.yaml`
### Scenario: the version and the help text are answered on stdout
_only when `bats --version` succeeds_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
bats --version
bats --help
```
#### Then
- after `bats --version`:
  - exit code is `0`
  - stdout matches `/^Bats [0-9]+\.[0-9]+\.[0-9]+/`
  - stderr is empty
- after `bats --help`:
  - exit code is `0`
  - stdout contains `Usage: bats [OPTIONS] <tests>`, `-f, --filter <regex>`, `-F, --formatter <type>`

### Scenario: a passing file is TAP on stdout, exit 0, and no droppings
_only when `bats --version` succeeds_
#### Given
- Fixture file `pass.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `pass.bats`:_
```text
@test "adds numbers" {
  result=$(( 2 + 2 ))
  [ "$result" -eq 4 ]
}
```
#### When
```shell
bats pass.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stderr is empty
- the step changed exactly created nothing, modified nothing, deleted nothing

#### Expected output
_expected stdout:_
```text
1..1
ok 1 adds numbers
```
### Scenario: a failing test is exit 1 and names the file, the line, and the command
_only when `bats --version` succeeds_
#### Given
- Fixture file `fail.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `fail.bats`:_
```text
@test "fails" {
  false
}
```
#### When
```shell
bats fail.bats
```
#### Then
- exit code is `1`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing

#### Expected output
_expected stdout:_
```text
1..1
not ok 1 fails
# (in test file fail.bats, line 2)
#   `false' failed
```
### Scenario: one failure does not hide the tests that passed
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "alpha passes" { true; }
@test "beta fails" { false; }
@test "gamma passes" { true; }
```
#### When
```shell
bats mixed.bats
```
#### Then
- exit code is `1`
- stdout contains `1..3`, `ok 1 alpha passes`, `not ok 2 beta fails`, `ok 3 gamma passes`

### Scenario: a command failing mid-test aborts the rest of the test body
_only when `bats --version` succeeds_
#### Given
- Fixture file `mid.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mid.bats`:_
```text
@test "mid failure aborts the test" {
  echo before
  false
  echo after
}
```
#### When
```shell
bats mid.bats
```
#### Then
- exit code is `1`
- stdout contains `# (in test file mid.bats, line 3)`, `# before`, does not contain `after`

### Scenario: skipped tests are ok and keep the run green
_only when `bats --version` succeeds_
#### Given
- Fixture file `skip.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `skip.bats`:_
```text
@test "skipped with reason" {
  skip "not on this platform"
  false
}

@test "skipped without reason" {
  skip
  false
}
```
#### When
```shell
bats skip.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 skipped with reason # skip not on this platform
ok 2 skipped without reason # skip
```
### Scenario: nothing to run is not an error
_only when `bats --version` succeeds_
#### Given
- Fixture file `empty.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
bats empty.bats
mkdir -p notests && cp empty.bats notests/none.bats
bats notests
```
#### Then
- after `bats empty.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `mkdir -p notests && cp empty.bats notests/none.bats`:
  - exit code is `0`
- after `bats notests`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..0
```
_expected stdout:_
```text
1..0
```
### Scenario: a test file that does not exist reports to stderr and leaves stdout empty
_only when `bats --version` succeeds_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
bats missing.bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `missing.bats" does not exist.`, `not ok 1 bats-gather-tests`

### Scenario: an unknown option is rejected before any test runs
_only when `bats --version` succeeds_
#### Given
- Fixture file `pass.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `pass.bats`:_
```text
@test "would have passed" { true; }
```
#### When
```shell
bats --bogus pass.bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Error: Bad command line option '--bogus'`, `Usage: bats [OPTIONS] <tests>`

### Scenario: no arguments is a usage error, not an empty run
_only when `bats --version` succeeds_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Error: Must specify at least one <test>`, `Usage: bats [OPTIONS] <tests>`

### Scenario: a file Bash cannot parse fails as the gather-tests pseudo test
_only when `bats --version` succeeds_
#### Given
- Fixture file `broken.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `broken.bats`:_
```text
@test "unterminated" {
```
#### When
```shell
bats broken.bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `not ok 1 bats-gather-tests`, `syntax error: unexpected end of file`

### Scenario: a minimum-version guard the runner cannot meet stops the file
_only when `bats --version` succeeds_
#### Given
- Fixture file `future.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `future.bats`:_
```text
bats_require_minimum_version 9.9.9

@test "never runs" { true; }
```
#### When
```shell
bats future.bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr does not contain `ok 1 never runs`, matches `/does not meet required minimum 9\.9\.9/`

### Scenario: UTF-8 test names survive into the report
_only when `bats --version` succeeds_
#### Given
- Fixture file `utf8.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `utf8.bats`:_
```text
@test "日本語のテスト ✓" { true; }
```
#### When
```shell
bats utf8.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 日本語のテスト ✓
```
### Scenario: several files on one command line are one numbered run
_only when `bats --version` succeeds_
#### Given
- Fixture file `first.bats` is created.
- Fixture file `second.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `first.bats`:_
```text
@test "first file" { true; }
```
_Fixture `second.bats`:_
```text
@test "second file" { false; }
```
#### When
```shell
bats first.bats second.bats
```
#### Then
- exit code is `1`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 first file
not ok 2 second file
# (in test file second.bats, line 1)
#   `@test "second file" { false; }' failed
```
## bats (the harness around each test)
What [Bats](https://github.com/bats-core/bats-core) does around a test body:
the setup and teardown functions and the order they run in, the `run` helper
that captures a command's status and output, the per-test temporary
directory and its cleanup, the variables a test can read about itself, and
the helper files it can load.

These are the parts a test author relies on without asserting them, so they
are asserted from outside here. Ordering and "did this even run" are pinned
with an append-only log written by the fixture: the log content is the
oracle, so a teardown that stops running after a failure, or a body that
runs despite a broken setup, shows up as a diff rather than as a missing
line nobody reads.

Source: `test/e2e/thirdparty/bats/lifecycle.atago.yaml`
### Scenario: setup_file runs once, setup and teardown run around every test
_only when `bats --version` succeeds_
#### Given
- Fixture file `order.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `order.bats`:_
```text
setup_file() { echo "setup_file" >> "$BATS_TEST_DIRNAME/order.log"; }
setup() { echo "setup" >> "$BATS_TEST_DIRNAME/order.log"; }
teardown() { echo "teardown" >> "$BATS_TEST_DIRNAME/order.log"; }
teardown_file() { echo "teardown_file" >> "$BATS_TEST_DIRNAME/order.log"; }

@test "first" { echo "first" >> "$BATS_TEST_DIRNAME/order.log"; }
@test "second" { echo "second" >> "$BATS_TEST_DIRNAME/order.log"; }
```
#### When
```shell
bats order.bats
```
#### Then
- exit code is `0`
- the step changed exactly created `order.log`, modified nothing, deleted nothing
- file `order.log` equals exact bytes

### Scenario: a failing setup keeps the body from running and teardown still runs
_only when `bats --version` succeeds_
#### Given
- Fixture file `brokensetup.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `brokensetup.bats`:_
```text
setup() { echo "setup" >> "$BATS_TEST_DIRNAME/order.log"; false; }
teardown() { echo "teardown" >> "$BATS_TEST_DIRNAME/order.log"; }

@test "body" { echo "body" >> "$BATS_TEST_DIRNAME/order.log"; }
```
#### When
```shell
bats brokensetup.bats
```
#### Then
- exit code is `1`
- stdout contains `not ok 1 body`, `# (from function `setup' in test file brokensetup.bats, line 1)`
- file `order.log` equals exact bytes

### Scenario: teardown still runs when the test itself fails
_only when `bats --version` succeeds_
#### Given
- Fixture file `teardown.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `teardown.bats`:_
```text
teardown() { echo "teardown after $BATS_TEST_NAME" >> "$BATS_TEST_DIRNAME/order.log"; }

@test "fails" { false; }
```
#### When
```shell
bats teardown.bats
```
#### Then
- exit code is `1`
- file `order.log` equals exact bytes

### Scenario: the run helper captures the status, the output, and the lines
_only when `bats --version` succeeds_
#### Given
- Fixture file `runhelper.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `runhelper.bats`:_
```text
bats_require_minimum_version 1.5.0

@test "status and output are captured, not fatal" {
  run bash -c 'echo first; echo second; exit 3'
  [ "$status" -eq 3 ]
  [ "$output" = "first
second" ]
  [ "${lines[0]}" = "first" ]
  [ "${lines[1]}" = "second" ]
}

@test "an expected status is part of the call" {
  run -3 bash -c 'exit 3'
}

@test "separate stderr splits the streams" {
  run --separate-stderr bash -c 'echo to-stdout; echo to-stderr >&2'
  [ "$output" = "to-stdout" ]
  [ "$stderr" = "to-stderr" ]
}
```
#### When
```shell
bats runhelper.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stderr is empty

#### Expected output
_expected stdout:_
```text
1..3
ok 1 status and output are captured, not fatal
ok 2 an expected status is part of the call
ok 3 separate stderr splits the streams
```
### Scenario: an unmet expected status fails the test with both codes named
_only when `bats --version` succeeds_
#### Given
- Fixture file `expected.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `expected.bats`:_
```text
bats_require_minimum_version 1.5.0

@test "expects the wrong status" {
  run -0 bash -c 'exit 3'
}
```
#### When
```shell
bats expected.bats
```
#### Then
- exit code is `1`
- stdout contains `failed, expected exit code 0, got 3`

### Scenario: using run flags without declaring the minimum version warns on stderr
_only when `bats --version` succeeds_
#### Given
- Fixture file `unguarded.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `unguarded.bats`:_
```text
@test "uses a run flag without the version guard" {
  run -3 bash -c 'exit 3'
}
```
#### When
```shell
bats unguarded.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stderr contains `The following warnings were encountered during tests:`, `BW02: Using flags on `run` requires at least BATS_VERSION=1.5.0.`, `Use `bats_require_minimum_version 1.5.0` to fix this message.`

#### Expected output
_expected stdout:_
```text
1..1
ok 1 uses a run flag without the version guard
```
### Scenario: each test gets its own temporary directory, removed when the run ends
_only when `bats --version` succeeds_
#### Given
- Fixture file `tmpdir.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TMPDIR.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TMPDIR.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `tmpdir.bats`:_
```text
@test "writes into its own tmpdir" {
  [ -d "$BATS_TEST_TMPDIR" ]
  echo scratch > "$BATS_TEST_TMPDIR/scratch.txt"
  echo "$BATS_TEST_TMPDIR" >> "$BATS_TEST_DIRNAME/paths.log"
}
```
#### When
```shell
mkdir -p tmproot
bats tmpdir.bats
bats --no-tempdir-cleanup tmpdir.bats
```
#### Then
- after `mkdir -p tmproot`:
  - exit code is `0`
- after `bats tmpdir.bats`:
  - exit code is `0`
  - dir `tmproot` has 0 entries
- after `bats --no-tempdir-cleanup tmpdir.bats`:
  - exit code is `0`
  - stderr matches `/^BATS_RUN_TMPDIR: .*/tmproot/bats-run-/`
  - dir `tmproot` has 1 entry, matches glob `bats-run-*`

### Scenario: a test can read its own name, number, and file
_only when `bats --version` succeeds_
#### Given
- Fixture file `identity.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `identity.bats`:_
```text
@test "first" {
  echo "$BATS_TEST_NUMBER $BATS_TEST_NAME $(basename "$BATS_TEST_FILENAME")" >> "$BATS_TEST_DIRNAME/identity.log"
}

@test "second" {
  echo "$BATS_TEST_NUMBER $BATS_TEST_NAME" >> "$BATS_TEST_DIRNAME/identity.log"
}
```
#### When
```shell
bats identity.bats
```
#### Then
- exit code is `0`
- file `identity.log` equals exact bytes

### Scenario: load pulls in a helper file next to the test
_only when `bats --version` succeeds_
#### Given
- Fixture file `helper.bash` is created.
- Fixture file `uses_helper.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `helper.bash`:_
```text
greet() { echo "hello from the helper"; }
```
_Fixture `uses_helper.bats`:_
```text
load helper

@test "the helper is in scope" {
  run greet
  [ "$output" = "hello from the helper" ]
}
```
#### When
```shell
bats uses_helper.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 the helper is in scope
```
### Scenario: parallel jobs without the parallel binary fail as a raw shell error
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
@test "also passes" { true; }
```
#### When
```shell
bats --jobs 2 --parallel-binary-name definitely-not-a-parallel-binary mixed.bats
```
#### Then
- exit code is `1`
- stdout contains `1..2`, `definitely-not-a-parallel-binary: command not found`, `# bats warning: Executed 0 instead of expected 2 tests`, does not contain `ok 1 passes`

## bats (formatters, reports, and gathered output)
The same run, reported in every shape [Bats](https://github.com/bats-core/bats-core)
can report it: plain TAP, TAP version 13 with YAML diagnostics, JUnit XML on
stdout or written to a report directory, the per-test output files gathered
by `--gather-test-outputs-in`, and the pretty formatter that only appears
when a terminal is attached.

Formatters are where a test runner is easiest to break without noticing: the
tests still pass, and only the shape of the report changes. Each scenario
below runs one fixture — one passing test and one failing test — and pins
what a given flag makes of it, including which stream it lands on and which
files it leaves on disk. The pretty formatter is checked on a rendered
terminal screen, since that is the only place it exists.

Source: `test/e2e/thirdparty/bats/reporting.atago.yaml`
### Scenario: TAP version 13 carries the failure as a YAML block
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
@test "fails" { false; }
```
#### When
```shell
bats --formatter tap13 mixed.bats
```
#### Then
- exit code is `1`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
TAP version 13
1..2
ok 1 passes
not ok 2 fails
  ---
  message: |
    (in test file mixed.bats, line 2)
      `@test "fails" { false; }' failed
  ...
```
### Scenario: the JUnit formatter reports counts and the failure element
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
@test "fails" { false; }
```
#### When
```shell
bats --formatter junit mixed.bats
```
#### Then
- exit code is `1`
- stdout contains `<?xml version="1.0" encoding="UTF-8"?>`, `name="mixed.bats" tests="2" failures="1" errors="0" skipped="0"`, `<testcase classname="mixed.bats" name="passes"`, `<failure type="failure">(in test file mixed.bats, line 2)`

### Scenario: a report formatter writes a file and leaves stdout as TAP
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
@test "fails" { false; }
```
#### When
```shell
mkdir -p reports
bats --report-formatter junit --output reports mixed.bats
```
#### Then
- after `mkdir -p reports`:
  - exit code is `0`
- after `bats --report-formatter junit --output reports mixed.bats`:
  - exit code is `1`
  - stdout matches `/^1\.\.2\nok 1 passes # in [0-9]+ ms\n/`
  - the step changed exactly created `reports/report.xml`, modified nothing, deleted nothing
  - file `reports/report.xml` contains `tests="2" failures="1"`, `<testcase classname="mixed.bats" name="fails"`

### Scenario: a report directory that does not exist is refused before the run
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
```
#### When
```shell
bats --report-formatter junit --output nosuch mixed.bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Error: Output path nosuch is not writeable`, `Usage: bats [OPTIONS] <tests>`
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: gathered output is one file per test, named after it, holding what it printed
_only when `bats --version` succeeds_
#### Given
- Fixture file `chatty.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `chatty.bats`:_
```text
@test "prints and passes" { echo "passing output"; }
@test "prints and fails" { echo "failing output"; false; }
```
#### When
```shell
mkdir -p gathered
bats --gather-test-outputs-in gathered chatty.bats
```
#### Then
- after `mkdir -p gathered`:
  - exit code is `0`
- after `bats --gather-test-outputs-in gathered chatty.bats`:
  - exit code is `1`
  - dir `gathered` contains `1-prints and passes.log`, contains `2-prints and fails.log`, has 2 entries
  - file `gathered/1-prints and passes.log` equals exact bytes
  - file `gathered/2-prints and fails.log` equals exact bytes

### Scenario: gathering into a directory that already has files is refused
_only when `bats --version` succeeds_
#### Given
- Fixture file `chatty.bats` is created.
- Fixture file `gathered/leftover.log` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `chatty.bats`:_
```text
@test "prints and passes" { echo "passing output"; }
```
_Fixture `gathered/leftover.log`:_
```text
from an earlier run
```
#### When
```shell
bats --gather-test-outputs-in gathered chatty.bats
```
#### Then
- exit code is `1`
- stdout is empty
- stderr equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing
- dir `gathered` has 1 entry

#### Expected output
_expected stderr:_
```text
Error: Directory 'gathered' must be empty for --gather-test-outputs-in
```
### Scenario: test output is hidden when it passes and shown when it fails
_only when `bats --version` succeeds_
#### Given
- Fixture file `chatty.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `chatty.bats`:_
```text
@test "prints and passes" { echo "passing output"; }
@test "prints and fails" { echo "failing output"; false; }
```
#### When
```shell
bats chatty.bats
bats --show-output-of-passing-tests chatty.bats
```
#### Then
- after `bats chatty.bats`:
  - exit code is `1`
  - stdout contains `# failing output`, does not contain `# passing output`
- after `bats --show-output-of-passing-tests chatty.bats`:
  - exit code is `1`
  - stdout contains `# passing output`, `# failing output`

### Scenario: timing and tracing add to the report without changing the verdict
_only when `bats --version` succeeds_
#### Given
- Fixture file `mixed.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
@test "fails" { false; }
```
#### When
```shell
bats --timing mixed.bats
bats --trace mixed.bats
```
#### Then
- after `bats --timing mixed.bats`:
  - exit code is `1`
  - stdout matches `/^1\.\.2\nok 1 passes in [0-9]+ms\nnot ok 2 fails in [0-9]+ms\n/`
- after `bats --trace mixed.bats`:
  - exit code is `1`
  - stdout contains `# $ [mixed.bats, line 2]`, `# $ false`

### Scenario: the quoting and file reference in a failure are configurable
_only when `bats --version` succeeds_
#### Given
- Fixture file `fail.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `fail.bats`:_
```text
@test "fails" {
  false
}
```
#### When
```shell
bats fail.bats
bats --code-quote-style [] fail.bats
bats --line-reference-format colon fail.bats
```
#### Then
- after `bats fail.bats`:
  - exit code is `1`
  - stdout contains `# (in test file fail.bats, line 2)`, `#   `false' failed`
- after `bats --code-quote-style [] fail.bats`:
  - exit code is `1`
  - stdout contains `#   [false] failed`
- after `bats --line-reference-format colon fail.bats`:
  - exit code is `1`
  - stdout contains `# (in test file fail.bats:2)`

### Scenario: the pretty formatter exists only on a terminal
_only when `bats --version` succeeds · skipped on Windows_
#### Given
- Fixture file `mixed.bats` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mixed.bats`:_
```text
@test "passes" { true; }
@test "fails" { false; }
```
#### When
```shell
# interactive (pty): bats mixed.bats
```
#### Then
- rendered screen contains `mixed.bats`, `✓ passes`, `✗ fails`, `2 tests, 1 failure`
- rendered screen does not contain `1..2`

## bats (choosing which tests run)
Everything [Bats](https://github.com/bats-core/bats-core) offers for running
less than the whole suite: counting without running, `--filter` on the test
name, `--filter-tags` on the tags declared in the file, `--filter-status` on
what the previous run recorded, and a directory walked flat or recursively.

A selection flag has two halves, and only one of them shows up in the report:
the tests it runs, and the tests it does not. The scenarios below assert both
— the plan line counts what ran, and a test that writes a file proves what
stayed unexecuted. `--filter-status` is the interesting one, because it makes
the runner stateful: it reads a ledger the previous run wrote, and refuses,
with instructions, when the directory that ledger lives in does not exist.

Source: `test/e2e/thirdparty/bats/selection.atago.yaml`
### Scenario: counting reports the number of tests and runs none of them
_only when `bats --version` succeeds_
#### Given
- Fixture file `side_effect.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `side_effect.bats`:_
```text
@test "writes a file" { echo ran > "$BATS_TEST_DIRNAME/ran.txt"; }
@test "writes another" { echo ran >> "$BATS_TEST_DIRNAME/ran.txt"; }
```
#### When
```shell
bats --count side_effect.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing
- file `ran.txt` does not exist

#### Expected output
_expected stdout:_
```text
2
```
### Scenario: the name filter is a regular expression, not a substring
_only when `bats --version` succeeds_
#### Given
- Fixture file `names.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `names.bats`:_
```text
@test "alpha" { true; }
@test "alpha extended" { true; }
@test "beta" { true; }
```
#### When
```shell
bats --filter alpha names.bats
bats --filter ^alpha$ names.bats
bats --count --filter ^alpha$ names.bats
```
#### Then
- after `bats --filter alpha names.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --filter ^alpha$ names.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --count --filter ^alpha$ names.bats`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 alpha
ok 2 alpha extended
```
_expected stdout:_
```text
1..1
ok 1 alpha
```
_expected stdout:_
```text
1
```
### Scenario: a filter that matches nothing is an empty plan, not a failure
_only when `bats --version` succeeds_
#### Given
- Fixture file `names.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `names.bats`:_
```text
@test "alpha" { true; }
```
#### When
```shell
bats --filter zzz names.bats
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..0
```
### Scenario: tags select and deselect tests, and file tags apply to the whole file
_only when `bats --version` succeeds_
#### Given
- Fixture file `tagged.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `tagged.bats`:_
```text
# bats file_tags=in-file

# bats test_tags=fast,unit
@test "fast unit" { true; }

# bats test_tags=fast
@test "fast only" { true; }

# bats test_tags=slow
@test "slow only" { true; }
```
#### When
```shell
bats --filter-tags fast tagged.bats
bats --filter-tags fast,unit tagged.bats
bats --filter-tags unit --filter-tags slow tagged.bats
bats --filter-tags !fast tagged.bats
bats --count --filter-tags in-file tagged.bats
bats --filter-tags never-used tagged.bats
```
#### Then
- after `bats --filter-tags fast tagged.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --filter-tags fast,unit tagged.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --filter-tags unit --filter-tags slow tagged.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --filter-tags !fast tagged.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --count --filter-tags in-file tagged.bats`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --filter-tags never-used tagged.bats`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 fast unit
ok 2 fast only
```
_expected stdout:_
```text
1..1
ok 1 fast unit
```
_expected stdout:_
```text
1..2
ok 1 fast unit
ok 2 slow only
```
_expected stdout:_
```text
1..1
ok 1 slow only
```
_expected stdout:_
```text
3
```
_expected stdout:_
```text
1..0
```
### Scenario: rerunning failures requires a run-log directory the user must create
_only when `bats --version` succeeds_
#### Given
- Fixture file `mix.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mix.bats`:_
```text
@test "alpha passes" { true; }
@test "beta fails" { false; }
@test "gamma passes" { true; }
```
#### When
```shell
bats --filter-status failed mix.bats
```
#### Then
- exit code is `1`
- stdout contains `.bats/run-logs/' to save failed tests`, `Please create this folder, add it to .gitignore and try again.`
- stderr is empty
- dir `.bats` does not exist

### Scenario: with the run-log directory in place a run records it and the next run replays only the failures
_only when `bats --version` succeeds_
#### Given
- Fixture file `mix.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mix.bats`:_
```text
@test "alpha passes" { true; }
@test "beta fails" { false; }
@test "gamma passes" { true; }
```
#### When
```shell
mkdir -p .bats/run-logs
bats mix.bats
bats --filter-status failed mix.bats
```
#### Then
- after `mkdir -p .bats/run-logs`:
  - exit code is `0`
- after `bats mix.bats`:
  - exit code is `1`
  - stdout contains `1..3`, `not ok 2 beta fails`
  - the step changed exactly created `.bats/run-logs/*.log`, modified nothing, deleted nothing
- after `bats --filter-status failed mix.bats`:
  - exit code is `1`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
not ok 1 beta fails
# (in test file mix.bats, line 2)
#   `@test "beta fails" { false; }' failed
```
### Scenario: replaying with nothing to replay says so and stays green
_only when `bats --version` succeeds_
#### Given
- Fixture file `green.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `green.bats`:_
```text
@test "alpha passes" { true; }
@test "beta passes too" { true; }
```
#### When
```shell
mkdir -p .bats/run-logs
bats green.bats
bats --filter-status failed green.bats
```
#### Then
- after `mkdir -p .bats/run-logs`:
  - exit code is `0`
- after `bats green.bats`:
  - exit code is `0`
- after `bats --filter-status failed green.bats`:
  - exit code is `0`
  - stdout equals an exact value
  - stderr equals an exact value

#### Expected output
_expected stdout:_
```text
1..0
```
_expected stderr:_
```text
There were no tests of status 'failed' in the last recorded run.
```
### Scenario: a test the ledger has never seen counts as missed
_only when `bats --version` succeeds_
#### Given
- Fixture file `growing.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `growing.bats` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `growing.bats`:_
```text
@test "already recorded" { true; }
```
_Fixture `growing.bats`:_
```text
@test "already recorded" { true; }
@test "added afterwards" { true; }
```
#### When
```shell
mkdir -p .bats/run-logs
bats growing.bats
bats --filter-status missed growing.bats
```
#### Then
- after `mkdir -p .bats/run-logs`:
  - exit code is `0`
- after `bats growing.bats`:
  - exit code is `0`
- after `bats --filter-status missed growing.bats`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 added afterwards
```
### Scenario: a directory is walked flat unless recursion is asked for
_only when `bats --version` succeeds_
#### Given
- Fixture file `suite/top.bats` is created.
- Fixture file `suite/sub/nested.bats` is created.
- Fixture file `suite/notes.txt` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `suite/top.bats`:_
```text
@test "top level" { true; }
```
_Fixture `suite/sub/nested.bats`:_
```text
@test "nested" { true; }
```
_Fixture `suite/notes.txt`:_
```text
not a test file
```
#### When
```shell
bats suite
bats --recursive suite
```
#### Then
- after `bats suite`:
  - exit code is `0`
  - stdout equals an exact value
- after `bats --recursive suite`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 top level
```
_expected stdout:_
```text
1..2
ok 1 nested
ok 2 top level
```
