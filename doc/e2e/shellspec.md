# atago Behavior Specs
## Summary
4 suites · 39 scenarios
## Contents
- [shellspec (what a specfile can do)](#shellspec-what-a-specfile-can-do) — 10 scenarios
  - [hooks run around every example, and once around them all](#scenario-hooks-run-around-every-example-and-once-around-them-all)
  - [a script is included and its functions are called or run](#scenario-a-script-is-included-and-its-functions-are-called-or-run)
  - [a data block is the example's standard input](#scenario-a-data-block-is-the-examples-standard-input)
  - [parameters turn one example into one per row](#scenario-parameters-turn-one-example-into-one-per-row)
  - [a mock replaces a command only inside the group that declares it](#scenario-a-mock-replaces-a-command-only-inside-the-group-that-declares-it)
  - [the sandbox takes PATH away so an unmocked command cannot be reached](#scenario-the-sandbox-takes-path-away-so-an-unmocked-command-cannot-be-reached)
  - [the scaffold init writes is runnable, and writing it twice changes nothing](#scenario-the-scaffold-init-writes-is-runnable-and-writing-it-twice-changes-nothing)
  - [support commands are generated as executables under the helper directory](#scenario-support-commands-are-generated-as-executables-under-the-helper-directory)
  - [the options file is read on every run](#scenario-the-options-file-is-read-on-every-run)
  - [the report is colored on a terminal and plain when colors are refused](#scenario-the-report-is-colored-on-a-terminal-and-plain-when-colors-are-refused)
- [shellspec (formatters, report files, and tracing)](#shellspec-formatters-report-files-and-tracing) — 10 scenarios
  - [the default formatter is dots, a count, and the failure in full](#scenario-the-default-formatter-is-dots-a-count-and-the-failure-in-full)
  - [the documentation formatter nests the names as written](#scenario-the-documentation-formatter-nests-the-names-as-written)
  - [the failures formatter is one line an editor can jump to](#scenario-the-failures-formatter-is-one-line-an-editor-can-jump-to)
  - [the null formatter reports nothing and still returns the verdict](#scenario-the-null-formatter-reports-nothing-and-still-returns-the-verdict)
  - [the JUnit formatter carries the counts, the message, and the output](#scenario-the-junit-formatter-carries-the-counts-the-message-and-the-output)
  - [report files are written beside the console report, where asked](#scenario-report-files-are-written-beside-the-console-report-where-asked)
  - [profiling ranks the slowest examples](#scenario-profiling-ranks-the-slowest-examples)
  - [tracing writes the evaluated commands to a log file](#scenario-tracing-writes-the-evaluated-commands-to-a-log-file)
  - [running in parallel does not change the report](#scenario-running-in-parallel-does-not-change-the-report)
  - [randomizing the order keeps the verdict and breaks the TAP plan line](#scenario-randomizing-the-order-keeps-the-verdict-and-breaks-the-tap-plan-line)
- [shellspec (choosing which examples run)](#shellspec-choosing-which-examples-run) — 8 scenarios
  - [counting and listing report the suite without running it](#scenario-counting-and-listing-report-the-suite-without-running-it)
  - [a dry run prints the report a run would print, and runs nothing](#scenario-a-dry-run-prints-the-report-a-run-would-print-and-runs-nothing)
  - [the name filter matches each name as a pattern, not as a substring](#scenario-the-name-filter-matches-each-name-as-a-pattern-not-as-a-substring)
  - [tags select by name and by value](#scenario-tags-select-by-name-and-by-value)
  - [a focused example narrows the run only when focus is asked for](#scenario-a-focused-example-narrows-the-run-only-when-focus-is-asked-for)
  - [a specfile can be addressed by line number or by example id](#scenario-a-specfile-can-be-addressed-by-line-number-or-by-example-id)
  - [which files count as specfiles is a pattern](#scenario-which-files-count-as-specfiles-is-a-pattern)
  - [quick mode remembers what did not pass and forgets it once it does](#scenario-quick-mode-remembers-what-did-not-pass-and-forgets-it-once-it-does)
- [shellspec (the verdict and the exit contract)](#shellspec-the-verdict-and-the-exit-contract) — 11 scenarios
  - [the version and the help text are answered on stdout](#scenario-the-version-and-the-help-text-are-answered-on-stdout)
  - [a passing suite is TAP on stdout, exit 0, and no droppings](#scenario-a-passing-suite-is-tap-on-stdout-exit-0-and-no-droppings)
  - [a failing example is exit 101 and hands back the command that reruns it](#scenario-a-failing-example-is-exit-101-and-hands-back-the-command-that-reruns-it)
  - [the failure exit code is the caller's to choose](#scenario-the-failure-exit-code-is-the-callers-to-choose)
  - [a command line it cannot use is exit 1, before any spec is read](#scenario-a-command-line-it-cannot-use-is-exit-1-before-any-spec-is-read)
  - [a specfile that does not parse is a fatal error at its own exit code](#scenario-a-specfile-that-does-not-parse-is-a-fatal-error-at-its-own-exit-code)
  - [the syntax check reports the same mistake without running anything](#scenario-the-syntax-check-reports-the-same-mistake-without-running-anything)
  - [finding no examples is success unless it is declared a failure](#scenario-finding-no-examples-is-success-unless-it-is-declared-a-failure)
  - [an example that asserts nothing fails, unless warnings are demoted](#scenario-an-example-that-asserts-nothing-fails-unless-warnings-are-demoted)
  - [skipped, pending, and excluded examples keep the run green](#scenario-skipped-pending-and-excluded-examples-keep-the-run-green)
  - [the same specfile gets the same verdict in sh and in bash](#scenario-the-same-specfile-gets-the-same-verdict-in-sh-and-in-bash)

## shellspec (what a specfile can do)
The parts of [ShellSpec](https://shellspec.info/) a specfile author leans on:
the hooks around each example, `Include` and the difference between calling a
function and running it, `Data` for standard input, `Parameters` for one
example per row, command mocks and their scope, and the sandbox that takes
PATH away so an unmocked command cannot be reached by accident.

Ordering and "did this even run" are pinned with an append-only log the
specfile writes, so a hook that stops firing is a diff rather than a missing
line nobody reads. The last scenarios cover the project side — the scaffold
`--init` writes, the support commands `--gen-bin` generates, the options file
every run reads — and the colors the reports are drawn in, checked on a real
terminal because that is the only place they exist.

Source: `test/e2e/thirdparty/shellspec/features.atago.yaml`
### Scenario: hooks run around every example, and once around them all
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/hooks_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/hooks_spec.sh`:_
```text
Describe 'hooks'
  BeforeAll 'echo before_all >> hooks.log'
  AfterAll 'echo after_all >> hooks.log'
  BeforeEach 'echo before_each >> hooks.log'
  AfterEach 'echo after_each >> hooks.log'

  It 'first'
    When call echo first
    The output should equal first
    echo body_first >> hooks.log
  End

  It 'second'
    When call echo second
    The output should equal second
    echo body_second >> hooks.log
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created `hooks.log`, modified nothing, deleted nothing
- file `hooks.log` equals exact bytes

#### Expected output
_expected stdout:_
```text
1..2
ok 1 - hooks first
ok 2 - hooks second
```
### Scenario: a script is included and its functions are called or run
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `mylib.sh` is created.
- Fixture file `spec/mylib_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mylib.sh`:_
```text
greet() {
  echo "hello, $1"
}

bail_out() {
  echo "boom" >&2
  exit 3
}
```
_Fixture `spec/mylib_spec.sh`:_
```text
Describe 'mylib.sh'
  Include ./mylib.sh

  It 'calls a function in the current shell'
    When call greet world
    The output should equal "hello, world"
    The status should be success
  End

  It 'runs a function that exits, in a subshell'
    When run bail_out
    The status should equal 3
    The stderr should equal boom
    The output should be blank
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 - mylib.sh calls a function in the current shell
ok 2 - mylib.sh runs a function that exits, in a subshell
```
### Scenario: a data block is the example's standard input
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/data_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/data_spec.sh`:_
```text
Describe 'standard input'
  Data
    #|first line
    #|second line
    #|third line
  End

  It 'reads the block'
    When call wc -l
    The output should match pattern '*3*'
  End

  It 'sees the lines in order'
    When call head -n 1
    The output should equal 'first line'
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 - standard input reads the block
ok 2 - standard input sees the lines in order
```
### Scenario: parameters turn one example into one per row
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/params_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/params_spec.sh`:_
```text
Describe 'addition'
  Parameters
    1 2 3
    4 5 9
    10 20 30
  End

  It "adds $1 and $2"
    When call expr "$1" + "$2"
    The output should equal "$3"
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..3
ok 1 - addition adds 1 and 2
ok 2 - addition adds 4 and 5
ok 3 - addition adds 10 and 20
```
### Scenario: a mock replaces a command only inside the group that declares it
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/mock_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/mock_spec.sh`:_
```text
Describe 'mocking'
  Describe 'inside the group'
    Mock date
      echo "2001-02-03"
    End

    It 'sees the mock'
      When call date
      The output should equal "2001-02-03"
    End
  End

  It 'sees the real command outside it'
    When call date +%Y
    The output should match pattern '[0-9][0-9][0-9][0-9]'
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 - mocking inside the group sees the mock
ok 2 - mocking sees the real command outside it
```
### Scenario: the sandbox takes PATH away so an unmocked command cannot be reached
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/sandbox_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/sandbox_spec.sh`:_
```text
Describe 'an external command'
  It 'is reachable without the sandbox'
    When run command uname
    The status should be success
    The output should not be blank
  End
End
```
#### When
```shell
shellspec --format tap
shellspec --sandbox --format tap
```
#### Then
- after `shellspec --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --sandbox --format tap`:
  - exit code is `101`
  - stdout contains `not ok 1 - an external command is reachable without the sandbox # FAILED`, `got: failure (non-zero) [status: 127]`

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - an external command is reachable without the sandbox
```
### Scenario: the scaffold init writes is runnable, and writing it twice changes nothing
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
shellspec --init spec git
shellspec --format tap
shellspec --init spec git
```
#### Then
- after `shellspec --init spec git`:
  - exit code is `0`
  - the step changed exactly created `.shellspec`, `spec/spec_helper.sh`, `spec/*_spec.sh`, `.gitignore`, modified nothing, deleted nothing
  - file `.gitignore` contains `/.shellspec-quick.log`, `/report/`
- after `shellspec --format tap`:
  - exit code is `0`
  - stdout matches `/^1\.\.1\nnot ok 1 - .* # TODO You should implement hello function\n/`
- after `shellspec --init spec git`:
  - exit code is `0`
  - stdout contains `   exist   `
  - the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: support commands are generated as executables under the helper directory
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `spec/spec_helper.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
shellspec --gen-bin @true @cat
shellspec --gen-bin @true @cat
```
#### Then
- after `shellspec --gen-bin @true @cat`:
  - exit code is `1`
  - stdout is empty
  - stderr matches `/^shellspec helper directory not found: .*/spec\n$/`
  - the step changed exactly created nothing, modified nothing, deleted nothing
- after `shellspec --gen-bin @true @cat`:
  - exit code is `0`
  - stdout contains `Generate @true (spec/support/bin/@true)`, `Generate @cat (spec/support/bin/@cat)`
  - the step changed exactly created `spec/support/bin/@true`, `spec/support/bin/@cat`, modified nothing, deleted nothing
  - file `spec/support/bin/@cat` is executable

### Scenario: the options file is read on every run
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/simple_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.shellspec`:_
```text
--format tap
```
_Fixture `spec/simple_spec.sh`:_
```text
Describe 'options file'
  It 'is applied without being asked for'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec
shellspec --format failures
```
#### Then
- after `shellspec`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --format failures`:
  - exit code is `0`
  - stdout is empty

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - options file is applied without being asked for
```
### Scenario: the report is colored on a terminal and plain when colors are refused
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/simple_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/simple_spec.sh`:_
```text
Describe 'colors'
  It 'passes'
    When call true
    The status should be success
  End
End
```
#### When
```shell
# interactive (pty): shellspec
# interactive (pty): shellspec
```
#### Then
- after `interactive (pty): shellspec`:
  - rendered screen is checked and shows "1 example, 0 failures" in green
- after `interactive (pty): shellspec`:
  - rendered screen contains `1 example, 0 failures` and shows "1 example, 0 failures" in default

## shellspec (formatters, report files, and tracing)
One suite — a passing example and a failing one — reported in every shape
[ShellSpec](https://shellspec.info/) offers: progress dots, nested
documentation, JUnit XML, the editor-oriented `failures` lines, and `null`,
which prints nothing and still returns the verdict. Alongside the console
formatter, `--output` writes report files, `--profile` ranks the slowest
examples, and `--xtrace` writes an evaluation trace to a log file.

The last two scenarios are about what a report is worth: parallel jobs must
not change the verdict, and `--random` must not change it either — though it
does damage the TAP plan line, which is pinned here as it behaves.

Source: `test/e2e/thirdparty/shellspec/reporting.atago.yaml`
### Scenario: the default formatter is dots, a count, and the failure in full
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/mixed_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/mixed_spec.sh`:_
```text
Describe 'mixed'
  It 'passes'
    When call expr 2 + 2
    The output should equal 4
  End

  It 'fails'
    When call expr 2 + 2
    The output should equal 5
  End
End
```
#### When
```shell
shellspec
```
#### Then
- exit code is `101`
- stdout contains `Running: `, `.F`, `1.1) The output should equal 5`, `expected: 5`, `got: 4`, `2 examples, 1 failure`

### Scenario: the documentation formatter nests the names as written
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/nested_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/nested_spec.sh`:_
```text
Describe 'outer'
  Describe 'inner'
    It 'passes'
      When call true
      The status should be success
    End
  End
End
```
#### When
```shell
shellspec --format documentation
```
#### Then
- exit code is `0`
- stdout contains `"outer\n  inner\n    passes\n"`

### Scenario: the failures formatter is one line an editor can jump to
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/mixed_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/mixed_spec.sh`:_
```text
Describe 'mixed'
  It 'passes'
    When call true
    The status should be success
  End

  It 'fails'
    When call true
    The status should equal 9
  End
End
```
#### When
```shell
shellspec --format failures
```
#### Then
- exit code is `101`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
./spec/mixed_spec.sh:7:FAILED:mixed fails
```
### Scenario: the null formatter reports nothing and still returns the verdict
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/mixed_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/mixed_spec.sh`:_
```text
Describe 'mixed'
  It 'fails'
    When call true
    The status should equal 9
  End
End
```
#### When
```shell
shellspec --format null
```
#### Then
- exit code is `101`
- stdout is empty

### Scenario: the JUnit formatter carries the counts, the message, and the output
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/mixed_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/mixed_spec.sh`:_
```text
Describe 'mixed'
  It 'passes'
    When call true
    The status should be success
  End

  It 'fails'
    When call echo unexpected
    The output should equal expected
  End
End
```
#### When
```shell
shellspec --format junit
```
#### Then
- exit code is `101`
- stdout contains `<?xml version="1.0" encoding="UTF-8"?>`, `tests="2"`, `failures="1"`, `<testcase time="0" classname="spec/mixed_spec.sh" name="mixed fails">`, `<failure message="The output should equal expected">`, `<![CDATA[expected: "expected"`, `got: "unexpected"`

### Scenario: report files are written beside the console report, where asked
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/mixed_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/mixed_spec.sh`:_
```text
Describe 'mixed'
  It 'passes'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --output tap --output junit --format progress
shellspec --output tap --reportdir elsewhere --format null
```
#### Then
- after `shellspec --output tap --output junit --format progress`:
  - exit code is `0`
  - stdout contains `1 example, 0 failures`
  - the step changed exactly created `report/results.tap`, `report/results_junit.xml`, modified nothing, deleted nothing
  - file `report/results.tap` equals exact bytes
  - file `report/results_junit.xml` contains `name="mixed passes"`
- after `shellspec --output tap --reportdir elsewhere --format null`:
  - exit code is `0`
  - the step changed exactly created `elsewhere/results.tap`, modified nothing, deleted nothing

### Scenario: profiling ranks the slowest examples
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/two_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/two_spec.sh`:_
```text
Describe 'two'
  It 'first'
    When call true
    The status should be success
  End

  It 'second'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --profile --profile-limit 1
```
#### Then
- exit code is `0`
- stdout contains `# Top 1 slowest examples of the 2 examples`, matches `/#\s+1 [0-9]+\.[0-9]+ spec/two_spec\.sh:[0-9]+-[0-9]+/`

### Scenario: tracing writes the evaluated commands to a log file
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/traced_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/traced_spec.sh`:_
```text
Describe 'traced'
  greet() { echo "hello, $1"; }

  It 'greets'
    When call greet world
    The output should equal "hello, world"
  End
End
```
#### When
```shell
shellspec --xtrace --log-file trace.log --format tap
```
#### Then
- exit code is `0`
- stderr contains `Fall back to trace-only mode. All expectations will be skipped.`
- the step changed exactly created `trace.log`, modified nothing, deleted nothing
- file `trace.log` contains `[spec/traced_spec.sh:5] When call greet world`, `+ greet world`, `+ echo hello, world`

### Scenario: running in parallel does not change the report
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/a_spec.sh` is created.
- Fixture file `spec/b_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/a_spec.sh`:_
```text
Describe 'a'
  It 'passes'
    When call true
    The status should be success
  End
End
```
_Fixture `spec/b_spec.sh`:_
```text
Describe 'b'
  It 'fails'
    When call true
    The status should equal 9
  End
End
```
#### When
```shell
shellspec --format tap
shellspec --jobs 2 --format tap
```
#### Then
- after `shellspec --format tap`:
  - exit code is `101`
  - stdout contains `1..2`, `ok 1 - a passes`, `not ok 2 - b fails # FAILED`
- after `shellspec --jobs 2 --format tap`:
  - exit code is `101`
  - stdout contains `1..2`, `ok 1 - a passes`, `not ok 2 - b fails # FAILED`

### Scenario: randomizing the order keeps the verdict and breaks the TAP plan line
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/a_spec.sh` is created.
- Fixture file `spec/b_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/a_spec.sh`:_
```text
Describe 'a'
  It 'passes'
    When call true
    The status should be success
  End
End
```
_Fixture `spec/b_spec.sh`:_
```text
Describe 'b'
  It 'also passes'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --random specfiles --format tap
shellspec --random none --format tap
```
#### Then
- after `shellspec --random specfiles --format tap`:
  - exit code is `0`
  - stdout contains `ok 1 - `, `ok 2 - `
  - stderr matches `/^Randomized with seed [0-9]+/`
  - stdout matches `/^1\.\.spec/[ab]_spec\.sh\n/`
- after `shellspec --random none --format tap`:
  - exit code is `0`
  - stdout matches `/^1\.\.2\n/`

## shellspec (choosing which examples run)
Every way [ShellSpec](https://shellspec.info/) narrows a run: counting and
listing without running, `--dry-run`, the name filter, tags, focus, line and
id ranges, the file pattern, and quick mode — which remembers what did not
pass last time in a ledger on disk.

A selection flag is asserted on both halves here. The plan line says what
ran; a specfile whose examples write a file says what did not, so "counted
but not executed" is proven rather than assumed. Quick mode gets the full
state machine: the ledger appearing, the replay narrowing to what failed,
the ledger emptying once it passes, and the next run going wide again.

Source: `test/e2e/thirdparty/shellspec/selection.atago.yaml`
### Scenario: counting and listing report the suite without running it
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/side_effect_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/side_effect_spec.sh`:_
```text
Describe 'side effects'
  It 'writes a file'
    When call touch ran.txt
    The status should be success
  End

  It 'writes another'
    When call touch ran-too.txt
    The status should be success
  End
End
```
#### When
```shell
shellspec --count
shellspec --list examples
shellspec --list examples:lineno
```
#### Then
- after `shellspec --count`:
  - exit code is `0`
  - stdout equals an exact value
  - the step changed exactly created nothing, modified nothing, deleted nothing
- after `shellspec --list examples`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --list examples:lineno`:
  - exit code is `0`
  - stdout equals an exact value
  - file `ran.txt` does not exist
  - file `ran-too.txt` does not exist

#### Expected output
_expected stdout:_
```text
1 2
```
_expected stdout:_
```text
spec/side_effect_spec.sh:@1-1
spec/side_effect_spec.sh:@1-2
```
_expected stdout:_
```text
spec/side_effect_spec.sh:2
spec/side_effect_spec.sh:7
```
### Scenario: a dry run prints the report a run would print, and runs nothing
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/side_effect_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/side_effect_spec.sh`:_
```text
Describe 'side effects'
  It 'writes a file and would fail'
    When call touch ran.txt
    The status should equal 9
  End
End
```
#### When
```shell
shellspec --dry-run --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing
- file `ran.txt` does not exist

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - side effects writes a file and would fail
```
### Scenario: the name filter matches each name as a pattern, not as a substring
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/names_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/names_spec.sh`:_
```text
Describe 'arithmetic'
  It 'adds numbers'
    When call expr 2 + 2
    The output should equal 4
  End

  It 'subtracts numbers'
    When call expr 4 - 2
    The output should equal 2
  End
End
```
#### When
```shell
shellspec --example 'adds numbers' --format tap
shellspec --example adds --format tap
shellspec --example '*adds*' --format tap
shellspec --example 'arithmetic adds numbers' --format tap
```
#### Then
- after `shellspec --example 'adds numbers' --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --example adds --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --example '*adds*' --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --example 'arithmetic adds numbers' --format tap`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - arithmetic adds numbers
```
_expected stdout:_
```text
1..0
```
_expected stdout:_
```text
1..1
ok 1 - arithmetic adds numbers
```
_expected stdout:_
```text
1..0
```
### Scenario: tags select by name and by value
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/tagged_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/tagged_spec.sh`:_
```text
Describe 'tagged'
  It 'fast one' speed:fast
    When call true
    The status should be success
  End

  It 'slow one' speed:slow
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --tag speed --format tap
shellspec --tag speed:fast --format tap
shellspec --tag never-used --format tap
```
#### Then
- after `shellspec --tag speed --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --tag speed:fast --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --tag never-used --format tap`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..2
ok 1 - tagged fast one
ok 2 - tagged slow one
```
_expected stdout:_
```text
1..1
ok 1 - tagged fast one
```
_expected stdout:_
```text
1..0
```
### Scenario: a focused example narrows the run only when focus is asked for
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/focused_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/focused_spec.sh`:_
```text
Describe 'focus'
  fIt 'the one being worked on'
    When call true
    The status should be success
  End

  It 'the other one'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --focus --format tap
shellspec --format tap
```
#### Then
- after `shellspec --focus --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --format tap`:
  - exit code is `0`
  - stdout contains `1..2`, `ok 1 - focus the one being worked on`, `ok 2 - focus the other one`, `You need to specify --focus option to run focused (underlined) example(s) only.`

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - focus the one being worked on
```
### Scenario: a specfile can be addressed by line number or by example id
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/ranges_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/ranges_spec.sh`:_
```text
Describe 'ranges'
  It 'first'
    When call true
    The status should be success
  End

  It 'second'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --format tap spec/ranges_spec.sh:7
shellspec --format tap spec/ranges_spec.sh:@1-1
shellspec --format tap spec/ranges_spec.sh:1
```
#### Then
- after `shellspec --format tap spec/ranges_spec.sh:7`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --format tap spec/ranges_spec.sh:@1-1`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --format tap spec/ranges_spec.sh:1`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - ranges second
```
_expected stdout:_
```text
1..1
ok 1 - ranges first
```
_expected stdout:_
```text
1..2
ok 1 - ranges first
ok 2 - ranges second
```
### Scenario: which files count as specfiles is a pattern
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/normal_spec.sh` is created.
- Fixture file `spec/unusual_test.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/normal_spec.sh`:_
```text
Describe 'normal'
  It 'is collected by default'
    When call true
    The status should be success
  End
End
```
_Fixture `spec/unusual_test.sh`:_
```text
Describe 'unusual'
  It 'is collected only under a matching pattern'
    When call true
    The status should be success
  End
End
```
#### When
```shell
shellspec --format tap
shellspec --pattern '*_test.sh' --format tap
```
#### Then
- after `shellspec --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --pattern '*_test.sh' --format tap`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - normal is collected by default
```
_expected stdout:_
```text
1..1
ok 1 - unusual is collected only under a matching pattern
```
### Scenario: quick mode remembers what did not pass and forgets it once it does
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/quick_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `spec/quick_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/quick_spec.sh`:_
```text
Describe 'quick'
  It 'always passes'
    When call true
    The status should be success
  End

  It 'fails at first'
    When call expr 2 + 2
    The output should equal 5
  End
End
```
_Fixture `spec/quick_spec.sh`:_
```text
Describe 'quick'
  It 'always passes'
    When call true
    The status should be success
  End

  It 'fails at first'
    When call expr 2 + 2
    The output should equal 4
  End
End
```
#### When
```shell
shellspec --quick --format tap
shellspec --quick --format tap
shellspec --quick --format tap
shellspec --quick --format tap
```
#### Then
- after `shellspec --quick --format tap`:
  - exit code is `101`
  - stdout contains `1..2`, `not ok 2 - quick fails at first # FAILED`
  - the step changed exactly created `.shellspec-quick.log`, modified nothing, deleted nothing
  - file `.shellspec-quick.log` equals exact bytes
- after `shellspec --quick --format tap`:
  - exit code is `101`
  - stdout equals an exact value
- after `shellspec --quick --format tap`:
  - exit code is `0`
  - stdout contains `1..1`, `ok 1 - quick fails at first`, `All examples have been passed. Rerun to prevent regression.`
  - file `.shellspec-quick.log` equals exact bytes
- after `shellspec --quick --format tap`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
1..1
not ok 1 - quick fails at first # FAILED
# (in specfile spec/quick_spec.sh, line 9)
# When call expr 2 + 2
# The output should equal 5
# 
#   expected: 5
#        got: 4
# 
```
_expected stdout:_
```text
1..2
ok 1 - quick always passes
ok 2 - quick fails at first
```
## shellspec (the verdict and the exit contract)
[ShellSpec](https://shellspec.info/) is a BDD test framework for POSIX
shells, and its own test suite is written in ShellSpec. What that suite
checks from the inside is pinned here from the outside: what the runner
reports, and what it returns to whatever started it.

ShellSpec answers with three levels rather than the usual two — 1 for a
command line it cannot use, 101 for specs that failed, 102 for a fatal error
while loading them — and both spec-level codes can be overridden, which is
the whole reason to separate them. Each level is asserted here, together
with the TAP the run prints, the rerun command a failure hands back, and the
two defaults worth knowing about: an example that asserts nothing is a
failure, and finding no examples at all is not.

Every specfile is written by the scenario that runs it. Nothing upstream is
copied.

Source: `test/e2e/thirdparty/shellspec/shellspec.atago.yaml`
### Scenario: the version and the help text are answered on stdout
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
shellspec --version
shellspec --help
```
#### Then
- after `shellspec --version`:
  - exit code is `0`
  - stdout matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
  - stderr is empty
- after `shellspec --help`:
  - exit code is `0`
  - stdout contains `Usage: shellspec`, `-f, --format FORMATTER`, `--failure-exit-code CODE`

### Scenario: a passing suite is TAP on stdout, exit 0, and no droppings
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/arithmetic_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/arithmetic_spec.sh`:_
```text
Describe 'arithmetic'
  It 'adds numbers'
    When call expr 2 + 2
    The output should equal 4
    The status should be success
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - arithmetic adds numbers
```
### Scenario: a failing example is exit 101 and hands back the command that reruns it
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/arithmetic_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/arithmetic_spec.sh`:_
```text
Describe 'arithmetic'
  It 'adds numbers'
    When call expr 2 + 2
    The output should equal 5
  End
End
```
#### When
```shell
shellspec --format tap
shellspec
```
#### Then
- after `shellspec --format tap`:
  - exit code is `101`
  - stdout equals an exact value
- after `shellspec`:
  - exit code is `101`
  - stdout contains `1 example, 1 failure`, `shellspec spec/arithmetic_spec.sh:2 # 1) arithmetic adds numbers FAILED`

#### Expected output
_expected stdout:_
```text
1..1
not ok 1 - arithmetic adds numbers # FAILED
# (in specfile spec/arithmetic_spec.sh, line 4)
# When call expr 2 + 2
# The output should equal 5
# 
#   expected: 5
#        got: 4
# 
```
### Scenario: the failure exit code is the caller's to choose
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/failing_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/failing_spec.sh`:_
```text
Describe 'failing'
  It 'does not hold'
    When call true
    The status should equal 9
  End
End
```
#### When
```shell
shellspec --failure-exit-code 42 --format tap
```
#### Then
- exit code is `42`
- stdout contains `not ok 1 - failing does not hold # FAILED`

### Scenario: a command line it cannot use is exit 1, before any spec is read
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/arithmetic_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/arithmetic_spec.sh`:_
```text
Describe 'arithmetic'
  It 'adds numbers'
    When call expr 2 + 2
    The output should equal 4
  End
End
```
#### When
```shell
shellspec --bogus
shellspec spec/missing_spec.sh
```
#### Then
- after `shellspec --bogus`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `shellspec spec/missing_spec.sh`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
Unrecognized option: --bogus
```
_expected stderr:_
```text
Not found a path: spec/missing_spec.sh.
```
### Scenario: a specfile that does not parse is a fatal error at its own exit code
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/broken_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/broken_spec.sh`:_
```text
Describe 'broken'
  It 'never closes'
```
#### When
```shell
shellspec
shellspec --error-exit-code 77
```
#### Then
- after `shellspec`:
  - exit code is `102`
  - stdout contains `Example aborted (exit status: 102)`
  - stderr contains `Syntax error: Unexpected end of file (expecting 'End') in spec/broken_spec.sh line 2`, `Fatal error occurred, terminated with exit status 102.`
- after `shellspec --error-exit-code 77`:
  - exit code is `77`
  - stdout contains `Example aborted (exit status: 77)`
  - stderr contains `Fatal error occurred, terminated with exit status 77.`

### Scenario: the syntax check reports the same mistake without running anything
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/good_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `spec/broken_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/good_spec.sh`:_
```text
Describe 'good'
  It 'passes'
    When call true
    The status should be success
  End
End
```
_Fixture `spec/broken_spec.sh`:_
```text
Describe 'broken'
  It 'never closes'
```
#### When
```shell
shellspec --syntax-check
shellspec --syntax-check
```
#### Then
- after `shellspec --syntax-check`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --syntax-check`:
  - exit code is `1`
  - stdout contains `spec/broken_spec.sh`, `spec/good_spec.sh`
  - stderr matches `/Syntax error: Unexpected end of file \(expecting 'End'\) in spec/broken_spec.sh line 2/`

#### Expected output
_expected stdout:_
```text
spec/good_spec.sh
```
### Scenario: finding no examples is success unless it is declared a failure
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/notes.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/notes.txt`:_
```text
no specfile matches the default pattern
```
#### When
```shell
shellspec
shellspec --fail-no-examples
```
#### Then
- after `shellspec`:
  - exit code is `0`
  - stdout contains `0 examples, 0 failures`
- after `shellspec --fail-no-examples`:
  - exit code is `101`
  - stdout contains `0 examples, 0 failures, no examples found`

### Scenario: an example that asserts nothing fails, unless warnings are demoted
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/silent_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/silent_spec.sh`:_
```text
Describe 'no expectation'
  It 'checks nothing'
    When call true
  End
End
```
#### When
```shell
shellspec --format tap
shellspec --no-warning-as-failure --format tap
```
#### Then
- after `shellspec --format tap`:
  - exit code is `101`
  - stdout contains `not ok 1 - no expectation checks nothing # WARNED`, `# Not found any expectation`
- after `shellspec --no-warning-as-failure --format tap`:
  - exit code is `0`
  - stdout contains `ok 1 - no expectation checks nothing # WARNED`, `# Not found any expectation`

### Scenario: skipped, pending, and excluded examples keep the run green
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/states_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/states_spec.sh`:_
```text
Describe 'states'
  It 'is skipped'
    Skip 'not applicable here'
    When call false
    The status should be success
  End

  It 'is pending'
    Pending 'not implemented yet'
    When call false
    The status should be success
  End

  xIt 'is excluded'
    When call false
    The status should be success
  End
End
```
#### When
```shell
shellspec --format tap
```
#### Then
- exit code is `0`
- stdout contains `ok 1 - states is skipped # SKIP not applicable here`, `not ok 2 - states is pending # TODO not implemented yet`, `ok 3 - states is excluded # SKIP Temporarily skipped`

### Scenario: the same specfile gets the same verdict in sh and in bash
_only when `shellspec --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.shellspec` is created.
- Fixture file `spec/portable_spec.sh` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `spec/portable_spec.sh`:_
```text
Describe 'portable shell'
  It 'expands parameters the same way everywhere'
    greet() { echo "hello, $1"; }
    When call greet world
    The output should equal "hello, world"
  End
End
```
#### When
```shell
shellspec --shell sh --format tap
shellspec --shell bash --format tap
shellspec --shell bash
```
#### Then
- after `shellspec --shell sh --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --shell bash --format tap`:
  - exit code is `0`
  - stdout equals an exact value
- after `shellspec --shell bash`:
  - exit code is `0`
  - stdout matches `/^Running: .*bash \[bash /`

#### Expected output
_expected stdout:_
```text
1..1
ok 1 - portable shell expands parameters the same way everywhere
```
_expected stdout:_
```text
1..1
ok 1 - portable shell expands parameters the same way everywhere
```
