# atago Behavior Specs
## Summary
2 suites · 17 scenarios
## Contents
- [shdotenv (dialects and output formats)](#shdotenv-dialects-and-output-formats) — 7 scenarios
  - [one file, six formats, each exactly as its consumer expects](#scenario-one-file-six-formats-each-exactly-as-its-consumer-expects)
  - [the output can be filtered and reduced to names, and sorted only partly](#scenario-the-output-can-be-filtered-and-reduced-to-names-and-sorted-only-partly)
  - [a backslash escape means different things in different dialects](#scenario-a-backslash-escape-means-different-things-in-different-dialects)
  - [a space after the key is a syntax error in some dialects and not in others](#scenario-a-space-after-the-key-is-a-syntax-error-in-some-dialects-and-not-in-others)
  - [a line that is not a definition is quoted back with what is wrong](#scenario-a-line-that-is-not-a-definition-is-quoted-back-with-what-is-wrong)
  - [the environment can be dumped as a .env file and read back](#scenario-the-environment-can-be-dumped-as-a-env-file-and-read-back)
  - [a file can come from standard input](#scenario-a-file-can-come-from-standard-input)
- [shdotenv (parsing a .env file safely)](#shdotenv-parsing-a-env-file-safely) — 10 scenarios
  - [the version, the help, and an option that does not exist](#scenario-the-version-the-help-and-an-option-that-does-not-exist)
  - [a file is read as data and printed as shell that can be evaluated](#scenario-a-file-is-read-as-data-and-printed-as-shell-that-can-be-evaluated)
  - [a file that tries to run a command runs nothing](#scenario-a-file-that-tries-to-run-a-command-runs-nothing)
  - [shell metacharacters inside a value stay text, even when a command runs](#scenario-shell-metacharacters-inside-a-value-stay-text-even-when-a-command-runs)
  - [a reference to something unset is an error unless it is allowed](#scenario-a-reference-to-something-unset-is-an-error-unless-it-is-allowed)
  - [the environment already in place is visible, and can be hidden](#scenario-the-environment-already-in-place-is-visible-and-can-be-hidden)
  - [the file is a default until overload is asked for](#scenario-the-file-is-a-default-until-overload-is-asked-for)
  - [two files that define the same key are refused rather than merged](#scenario-two-files-that-define-the-same-key-are-refused-rather-than-merged)
  - [a file that is not there is not an error](#scenario-a-file-that-is-not-there-is-not-an-error)
  - [quiet mode validates a file without printing it](#scenario-quiet-mode-validates-a-file-without-printing-it)

## shdotenv (dialects and output formats)
The two matrices [shdotenv](https://github.com/ko1nksm/shdotenv) exists for:
the .env syntax dialect it reads, and the format it writes.

A dialect is not a preference — it changes what a file means, so the
scenarios here feed one file to several dialects and assert that the answers
differ in the documented way: a backslash-n that stays two characters under
the POSIX rules and becomes a newline under Ruby's, a key with a space after
it that POSIX and Docker refuse and Ruby accepts. The format half is asserted
exactly, because these outputs are consumed by other programs: a shell, csh,
fish, JSON, JSONL, YAML.

Source: `test/e2e/thirdparty/shdotenv/dialects.atago.yaml`
### Scenario: one file, six formats, each exactly as its consumer expects
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
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
_Fixture `.env`:_
```text
NAME=world
GREETING="hello, ${NAME}"
```
#### When
```shell
shdotenv --format sh
shdotenv --format csh
shdotenv --format fish
shdotenv --format json
shdotenv --format jsonl
shdotenv --format yaml
```
#### Then
- after `shdotenv --format sh`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --format csh`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --format fish`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --format json`:
  - exit code is `0`
  - stdout at `$.NAME` equals `world`; at `$.GREETING` equals `hello, world`
- after `shdotenv --format jsonl`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --format yaml`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
export NAME='world'
export GREETING='hello, world'
```
_expected stdout:_
```text
setenv NAME 'world';
setenv GREETING 'hello, world';
```
_expected stdout:_
```text
set --export NAME 'world';
set --export GREETING 'hello, world';
```
_expected stdout:_
```text
{ "NAME": "world", "GREETING": "hello, world" }
```
_expected stdout:_
```text
NAME: "world"
GREETING: "hello, world"
```
### Scenario: the output can be filtered and reduced to names, and sorted only partly
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
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
_Fixture `.env`:_
```text
DELTA=4
BRAVO=2
ECHO=5
ALFA=1
CHARLIE=3
```
#### When
```shell
shdotenv
shdotenv --sort
shdotenv --grep '^[BC]'
shdotenv --name-only
shdotenv --name-only --sort
```
#### Then
- after `shdotenv`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --sort`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --grep '^[BC]'`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --name-only`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --name-only --sort`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
export DELTA='4'
export BRAVO='2'
export ECHO='5'
export ALFA='1'
export CHARLIE='3'
```
_expected stdout:_
```text
export ALFA='1'
export BRAVO='2'
export DELTA='4'
export ECHO='5'
export CHARLIE='3'
```
_expected stdout:_
```text
export BRAVO='2'
export CHARLIE='3'
```
_expected stdout:_
```text
DELTA
BRAVO
ECHO
ALFA
CHARLIE
```
_expected stdout:_
```text
DELTA
BRAVO
ECHO
ALFA
CHARLIE
```
### Scenario: a backslash escape means different things in different dialects
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
MESSAGE="first\nsecond"
```
#### When
```shell
shdotenv --dialect posix --format json
shdotenv --dialect ruby --format json
shdotenv --dialect node --format json
shdotenv --dialect go --format json
```
#### Then
- after `shdotenv --dialect posix --format json`:
  - exit code is `0`
  - stdout at `$.MESSAGE` equals `first\nsecond`
- after `shdotenv --dialect ruby --format json`:
  - exit code is `0`
  - stdout at `$.MESSAGE` equals `"first\nsecond"`
- after `shdotenv --dialect node --format json`:
  - exit code is `0`
  - stdout at `$.MESSAGE` equals `"first\nsecond"`
- after `shdotenv --dialect go --format json`:
  - exit code is `0`
  - stdout at `$.MESSAGE` equals `"first\nsecond"`

### Scenario: a space after the key is a syntax error in some dialects and not in others
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
KEY = value
```
#### When
```shell
shdotenv --dialect posix
shdotenv --dialect docker
shdotenv --dialect ruby
```
#### Then
- after `shdotenv --dialect posix`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `shdotenv --dialect docker`:
  - exit code is `1`
  - stderr equals an exact value
- after `shdotenv --dialect ruby`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stderr:_
```text
shdotenv: `KEY ': no space allowed after the key
```
_expected stderr:_
```text
shdotenv: `KEY ': no space allowed after the key
```
_expected stdout:_
```text
export KEY='value'
```
### Scenario: a line that is not a definition is quoted back with what is wrong
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
GOOD=value
THIS IS NOT A DEFINITION
```
_Fixture `.env`:_
```text
A-B=value
```
#### When
```shell
shdotenv
shdotenv
```
#### Then
- after `shdotenv`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `shdotenv`:
  - exit code is `1`
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
shdotenv: `THIS IS NOT A DEFINITION': not a variable definition
```
_expected stderr:_
```text
shdotenv: `A-B': the key is not a valid identifier
```
### Scenario: the environment can be dumped as a .env file and read back
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, OTHER_VALUE, ROUND_TRIP.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, ROUND_TRIP.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
shdotenv export ROUND_TRIP OTHER_VALUE
shdotenv -e dumped.env --format json
shdotenv export -n ROUND_TRIP
```
#### Then
- after `shdotenv export ROUND_TRIP OTHER_VALUE`:
  - exit code is `0`
  - the step changed exactly created `dumped.env`, modified nothing, deleted nothing
- after `shdotenv -e dumped.env --format json`:
  - exit code is `0`
  - stdout at `$.ROUND_TRIP` equals `a value with spaces`; at `$.OTHER_VALUE` equals `x'y`
- after `shdotenv export -n ROUND_TRIP`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
ROUND_TRIP
```
#### Generated artifacts
- `dumped.env`

### Scenario: a file can come from standard input
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `shdotenv`:_
```text
FROM_STDIN=yes
SECOND=2
```
#### When
```shell
shdotenv -e -
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing

#### Expected output
_expected stdout:_
```text
export FROM_STDIN='yes'
export SECOND='2'
```
## shdotenv (parsing a .env file safely)
[shdotenv](https://github.com/ko1nksm/shdotenv) loads .env files into the
environment without letting the file run anything. Its own test suite is
written in ShellSpec; what those tests check is pinned here from outside.

The safety promise is the center of this file, and it is asserted the only
way a security claim should be: a .env file that tries to run a command is
given every chance to, and the proof is that the file it would have created
does not exist. Around it are the parsing rules, the strict defaults that
make the tool usable in CI (an unset reference is an error, two files
defining the same key is an error), and the precedence between the file and
the environment already in place.

Source: `test/e2e/thirdparty/shdotenv/shdotenv.atago.yaml`
### Scenario: the version, the help, and an option that does not exist
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
shdotenv --version
shdotenv --help
shdotenv --nope
```
#### Then
- after `shdotenv --version`:
  - exit code is `0`
  - stdout matches `/^[0-9]+\.[0-9]+\.[0-9]+\n$/`
- after `shdotenv --help`:
  - exit code is `0`
  - stdout contains `Usage: shdotenv [OPTION]... [--] [[COMMAND | export] [ARG]...]`, `-d, --dialect DIALECT`, `-f, --format FORMAT`
- after `shdotenv --nope`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `--nope`

### Scenario: a file is read as data and printed as shell that can be evaluated
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
# a comment line
NAME=world
GREETING="hello, ${NAME}"
EMPTY=
QUOTED='no ${expansion} here'
```
#### When
```shell
shdotenv
shdotenv -- sh -c "echo [$GREETING]"
```
#### Then
- after `shdotenv`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv -- sh -c "echo [$GREETING]"`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
export NAME='world'
export GREETING='hello, world'
export EMPTY=''
export QUOTED='no ${expansion} here'
```
_expected stdout:_
```text
[hello, world]
```
### Scenario: a file that tries to run a command runs nothing
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
EVIL=$(touch pwned-unquoted)
```
_Fixture `.env`:_
```text
EVIL="$(touch pwned-quoted)"
```
_Fixture `.env`:_
```text
EVIL="`touch pwned-backtick`"
```
#### When
```shell
shdotenv
shdotenv
shdotenv
```
#### Then
- after `shdotenv`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `spaces are not allowed without quoting`
- after `shdotenv`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `` the following metacharacters must be escaped: $`"\ ``
- after `shdotenv`:
  - exit code is `1`
  - stderr contains `` the following metacharacters must be escaped: $`"\ ``
  - the step changed exactly created nothing, modified nothing, deleted nothing
  - file `pwned-unquoted` does not exist
  - file `pwned-quoted` does not exist
  - file `pwned-backtick` does not exist

### Scenario: shell metacharacters inside a value stay text, even when a command runs
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
PAYLOAD="x; touch pwned-semicolon"
```
#### When
```shell
shdotenv
shdotenv -- sh -c "echo [$PAYLOAD]"
```
#### Then
- after `shdotenv`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv -- sh -c "echo [$PAYLOAD]"`:
  - exit code is `0`
  - stdout equals an exact value
  - file `pwned-semicolon` does not exist

#### Expected output
_expected stdout:_
```text
export PAYLOAD='x; touch pwned-semicolon'
```
_expected stdout:_
```text
[x; touch pwned-semicolon]
```
### Scenario: a reference to something unset is an error unless it is allowed
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
VALUE="${NOT_SET_ANYWHERE}"
```
#### When
```shell
shdotenv
shdotenv --no-nounset
```
#### Then
- after `shdotenv`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `shdotenv --no-nounset`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stderr:_
```text
shdotenv: NOT_SET_ANYWHERE: the key is not set
```
_expected stdout:_
```text
export VALUE=''
```
### Scenario: the environment already in place is visible, and can be hidden
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL, OUTSIDE.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, OUTSIDE.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
DERIVED="from ${OUTSIDE}"
```
#### When
```shell
shdotenv
shdotenv --ignore-environment
```
#### Then
- after `shdotenv`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --ignore-environment`:
  - exit code is `1`
  - stderr equals an exact value

#### Expected output
_expected stdout:_
```text
export DERIVED='from the-host'
```
_expected stderr:_
```text
shdotenv: OUTSIDE: the key is not set
```
### Scenario: the file is a default until overload is asked for
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL, PRESET.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, PRESET.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, PRESET.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, PRESET.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
PRESET=from-the-file
```
#### When
```shell
shdotenv -- sh -c "echo [$PRESET]"
shdotenv --overload -- sh -c "echo [$PRESET]"
shdotenv
shdotenv --overload
```
#### Then
- after `shdotenv -- sh -c "echo [$PRESET]"`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv --overload -- sh -c "echo [$PRESET]"`:
  - exit code is `0`
  - stdout equals an exact value
- after `shdotenv`:
  - exit code is `0`
  - stdout is empty
- after `shdotenv --overload`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
[from-the-environment]
```
_expected stdout:_
```text
[from-the-file]
```
_expected stdout:_
```text
export PRESET='from-the-file'
```
### Scenario: two files that define the same key are refused rather than merged
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `first.env` is created.
- Fixture file `second.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `second.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `first.env`:_
```text
SHARED=one
ONLY_FIRST=1
```
_Fixture `second.env`:_
```text
SHARED=two
ONLY_SECOND=2
```
_Fixture `second.env`:_
```text
ONLY_SECOND=2
```
#### When
```shell
shdotenv -e first.env -e second.env
shdotenv -e first.env -e second.env
```
#### Then
- after `shdotenv -e first.env -e second.env`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `shdotenv -e first.env -e second.env`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stderr:_
```text
shdotenv: second.env: `SHARED' is already defined in the first.env
```
_expected stdout:_
```text
export SHARED='one'
export ONLY_FIRST='1'
export ONLY_SECOND='2'
```
### Scenario: a file that is not there is not an error
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
shdotenv -e missing.env
```
#### Then
- exit code is `0`
- stdout is empty
- stderr is empty

### Scenario: quiet mode validates a file without printing it
_only when `command -v shdotenv` succeeds · skipped on Windows_
#### Given
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.env` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.env`:_
```text
GOOD=value
```
_Fixture `.env`:_
```text
THIS IS NOT A DEFINITION
```
#### When
```shell
shdotenv --quiet
shdotenv --quiet
```
#### Then
- after `shdotenv --quiet`:
  - exit code is `0`
  - stdout is empty
- after `shdotenv --quiet`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
shdotenv: `THIS IS NOT A DEFINITION': not a variable definition
```
