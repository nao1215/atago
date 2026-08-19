# atago Behavior Specs
## Summary
2 suites · 14 scenarios
## Contents
- [getoptions (generating an option parser)](#getoptions-generating-an-option-parser) — 8 scenarios
  - [the version, the help, and a command that does not exist](#scenario-the-version-the-help-and-a-command-that-does-not-exist)
  - [the printed example is a definition the generator accepts](#scenario-the-printed-example-is-a-definition-the-generator-accepts)
  - [the generated parser is a shell function that parses what the definition declared](#scenario-the-generated-parser-is-a-shell-function-that-parses-what-the-definition-declared)
  - [the help text is generated from the same definition](#scenario-the-help-text-is-generated-from-the-same-definition)
  - [generation options change the artifact, not the behavior](#scenario-generation-options-change-the-artifact-not-the-behavior)
  - [the library is generated on its own, with shellcheck directives on request](#scenario-the-library-is-generated-on-its-own-with-shellcheck-directives-on-request)
  - [embedding makes the script self-contained, and erasing gives the file back](#scenario-embedding-makes-the-script-self-contained-and-erasing-gives-the-file-back)
  - [an embed block that is never closed is refused](#scenario-an-embed-block-that-is-never-closed-is-refused)
- [getoptions (what the generated parser does)](#getoptions-what-the-generated-parser-does) — 6 scenarios
  - [every declared form parses, and the rest arguments survive](#scenario-every-declared-form-parses-and-the-rest-arguments-survive)
  - [a flag can be turned back off, and -- ends the options](#scenario-a-flag-can-be-turned-back-off-and----ends-the-options)
  - [an abbreviation is accepted while it is unambiguous](#scenario-an-abbreviation-is-accepted-while-it-is-unambiguous)
  - [the mistakes a user makes each have their own message](#scenario-the-mistakes-a-user-makes-each-have-their-own-message)
  - [a custom validator decides, and says so when it refuses](#scenario-a-custom-validator-decides-and-says-so-when-it-refuses)
  - [a validator that is not defined is a validation error, not a crash](#scenario-a-validator-that-is-not-defined-is-a-validation-error-not-a-crash)

## getoptions (generating an option parser)
[getoptions](https://github.com/ko1nksm/getoptions) turns a short definition
into a POSIX-shell option parser. Its own test suite is written in ShellSpec;
what those tests check is pinned here from outside, by generating parsers and
then running them.

A generator is only as good as what it emits, so nothing here stops at "a
file was written": every generated artifact is executed, and the assertions
are about what the generated program does. The embed workflow gets the round
trip it deserves — a script with the parser embedded runs with getoptions
nowhere on PATH, and erasing the embedded block gives back the original file
byte for byte.

Source: `test/e2e/thirdparty/getoptions/gengetoptions.atago.yaml`
### Scenario: the version, the help, and a command that does not exist
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
gengetoptions --version
gengetoptions --help
gengetoptions --nope
```
#### Then
- after `gengetoptions --version`:
  - exit code is `0`
  - stdout matches `/^v[0-9]+\.[0-9]+\.[0-9]+\n$/`
- after `gengetoptions --help`:
  - exit code is `0`
  - stdout contains `Usage: gengetoptions [options]... <command> [arguments]...`, `  library                Generate custom library`, `  parser                 Generate option parser`, `  embed                  Embed the generated library or parser`
- after `gengetoptions --nope`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
Unrecognized option: --nope
```
### Scenario: the printed example is a definition the generator accepts
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
gengetoptions example
gengetoptions parser -f definition.sh parser_definition parse prog
gengetoptions parser -f definition.sh parser_definition parse
```
#### Then
- after `gengetoptions example`:
  - exit code is `0`
  - the step changed exactly created `definition.sh`, modified nothing, deleted nothing
  - file `definition.sh` contains `parser_definition() {`, `flag    FLAG    -f +f --{no-}flag`
- after `gengetoptions parser -f definition.sh parser_definition parse prog`:
  - exit code is `0`
  - file `parser.sh` contains `# Generated by getoptions (BEGIN)`, `parse() {`, `# Generated by getoptions (END)`
- after `gengetoptions parser -f definition.sh parser_definition parse`:
  - exit code is `0`
  - stderr contains `parameter not set`
  - file `underspecified.sh` does not contain `parse() {`

#### Generated artifacts
- `definition.sh`
- `parser.sh`
- `underspecified.sh`

### Scenario: the generated parser is a shell function that parses what the definition declared
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `tool.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage -- "Usage: tool [options]... [args]..." ''
  msg -- 'Options:'
  flag    VERBOSE -v --verbose -- "be noisy"
  param   NAME    -n --name    -- "who to greet"
  disp    :usage  -h --help
}
```
_Fixture `tool.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
eval "set -- $REST"
echo "VERBOSE=[$VERBOSE] NAME=[$NAME] rest=[$*]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse
./tool.sh -v --name world alpha beta
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse`:
  - exit code is `0`
- after `./tool.sh -v --name world alpha beta`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
VERBOSE=[1] NAME=[world] rest=[alpha beta]
```
#### Generated artifacts
- `parser.sh`

### Scenario: the help text is generated from the same definition
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `tool.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage -- "Usage: tool [options]... [args]..." ''
  msg -- 'Options:'
  flag    VERBOSE -v --verbose -- "be noisy"
  param   NAME    -n --name    -- "who to greet"
  disp    :usage  -h --help
}
```
_Fixture `tool.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse > parser.sh
./tool.sh --help
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse > parser.sh`:
  - exit code is `0`
- after `./tool.sh --help`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
Usage: tool [options]... [args]...

Options:
  -v, --verbose               be noisy
  -n, --name NAME             who to greet
  -h, --help                  
```
### Scenario: generation options change the artifact, not the behavior
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `run.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage -- "Usage: tool [options]" ''
  flag    VERBOSE -v --verbose
}
```
_Fixture `run.sh`:_
```text
#!/bin/sh
set -eu
. "./$1"
parse -v
echo "VERBOSE=[$VERBOSE]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse > tabs.sh
gengetoptions parser -i2 -f definition.sh parser_definition parse > spaces.sh
gengetoptions parser --no-comments -f definition.sh parser_definition parse > bare.sh
./run.sh spaces.sh
./run.sh bare.sh
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse > tabs.sh`:
  - exit code is `0`
- after `gengetoptions parser -i2 -f definition.sh parser_definition parse > spaces.sh`:
  - exit code is `0`
  - file `tabs.sh` contains `"\tcase $1 in"`
  - file `spaces.sh` does not contain `"\tcase $1 in"`
- after `gengetoptions parser --no-comments -f definition.sh parser_definition parse > bare.sh`:
  - exit code is `0`
  - file `bare.sh` does not contain `# Generated by getoptions (BEGIN)`
- after `./run.sh spaces.sh`:
  - exit code is `0`
  - stdout equals an exact value
- after `./run.sh bare.sh`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
VERBOSE=[1]
```
_expected stdout:_
```text
VERBOSE=[1]
```
### Scenario: the library is generated on its own, with shellcheck directives on request
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
gengetoptions library > library.sh
gengetoptions library --shellcheck > checked.sh
```
#### Then
- after `gengetoptions library > library.sh`:
  - exit code is `0`
  - file `library.sh` contains `getoptions() {`, `getoptions_help() {`
  - file `library.sh` does not contain `# shellcheck disable=SC2016,SC2317`
- after `gengetoptions library --shellcheck > checked.sh`:
  - exit code is `0`
  - file `checked.sh` contains `# shellcheck disable=SC2016,SC2317` exactly 3 times

### Scenario: embedding makes the script self-contained, and erasing gives the file back
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `tool.sh` is created.
- Fixture file `tool.orig` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `tool.sh`:_
```text
#!/bin/sh
set -eu

# @getoptions
parser_definition() {
  setup   REST help:usage -- "Usage: tool.sh [options]... [args]..." ''
  msg -- 'Options:'
  flag    VERBOSE -v --verbose -- "be noisy"
  param   NAME    -n --name    -- "who to greet"
  disp    :usage  -h --help
}
# @end

# @gengetoptions parser -i parser_definition parse
# @end

parse "$@"
eval "set -- $REST"
echo "VERBOSE=[$VERBOSE] NAME=[$NAME] rest=[$*]"
```
_Fixture `tool.orig`:_
```text
#!/bin/sh
set -eu

# @getoptions
parser_definition() {
  setup   REST help:usage -- "Usage: tool.sh [options]... [args]..." ''
  msg -- 'Options:'
  flag    VERBOSE -v --verbose -- "be noisy"
  param   NAME    -n --name    -- "who to greet"
  disp    :usage  -h --help
}
# @end

# @gengetoptions parser -i parser_definition parse
# @end

parse "$@"
eval "set -- $REST"
echo "VERBOSE=[$VERBOSE] NAME=[$NAME] rest=[$*]"
```
#### When
```shell
gengetoptions embed --overwrite tool.sh
PATH=/usr/bin:/bin ./tool.sh -v --name world alpha
gengetoptions embed --erase --overwrite tool.sh
```
#### Then
- after `gengetoptions embed --overwrite tool.sh`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `tool.sh`, deleted nothing
  - file `tool.sh` contains `# @gengetoptions parser -i parser_definition parse`, `# Generated by getoptions (BEGIN)`, `# Generated by getoptions (END)`
- after `PATH=/usr/bin:/bin ./tool.sh -v --name world alpha`:
  - exit code is `0`
  - stdout equals an exact value
- after `gengetoptions embed --erase --overwrite tool.sh`:
  - exit code is `0`
  - file `tool.sh` is byte-identical to `tool.orig`

#### Expected output
_expected stdout:_
```text
VERBOSE=[1] NAME=[world] rest=[alpha]
```
### Scenario: an embed block that is never closed is refused
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `broken.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `broken.sh`:_
```text
#!/bin/sh
# @gengetoptions parser -i parser_definition parse
echo "no end directive here"
```
#### When
```shell
gengetoptions embed broken.sh
```
#### Then
- exit code is `1`
- stderr contains `Missing @end directive`
- the step changed exactly created nothing, modified nothing, deleted nothing

## getoptions (what the generated parser does)
One definition, generated once per scenario, and then exercised the way a
user exercises a command line: short and long forms, clustering, negation,
counters, optional arguments, `--`, abbreviations, and every way of getting
it wrong.

This is the half that matters — a generator's contract is the behavior of
what it emits — so every assertion here is about a program
[getoptions](https://github.com/ko1nksm/getoptions) wrote, not about the
text it wrote.

Source: `test/e2e/thirdparty/getoptions/parsers.atago.yaml`
### Scenario: every declared form parses, and the rest arguments survive
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Fixture file `demo.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST plus:true help:usage abbr:true -- "Usage: demo [options]... [args]..." ''
  msg -- 'Options:'
  flag    FLAG    -f +f --{no-}flag -- "on and off"
  flag    VERBOSE -v --verbose counter:true init:=0 -- "repeatable"
  param   NAME    -n --name -- "one argument"
  option  LEVEL   -l --level on:"default" -- "optional argument"
  disp    :usage  -h --help
}
```
_Fixture `demo.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
eval "set -- $REST"
echo "FLAG=[$FLAG] VERBOSE=[$VERBOSE] NAME=[$NAME] LEVEL=[$LEVEL] rest=[$*]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh
./demo.sh --flag --verbose --verbose --name world --level=3 alpha beta
./demo.sh -fvvn world -l3 alpha
./demo.sh -l
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh`:
  - exit code is `0`
- after `./demo.sh --flag --verbose --verbose --name world --level=3 alpha beta`:
  - exit code is `0`
  - stdout equals an exact value
- after `./demo.sh -fvvn world -l3 alpha`:
  - exit code is `0`
  - stdout equals an exact value
- after `./demo.sh -l`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
FLAG=[1] VERBOSE=[2] NAME=[world] LEVEL=[3] rest=[alpha beta]
```
_expected stdout:_
```text
FLAG=[1] VERBOSE=[2] NAME=[world] LEVEL=[3] rest=[alpha]
```
_expected stdout:_
```text
FLAG=[] VERBOSE=[0] NAME=[] LEVEL=[default] rest=[]
```
### Scenario: a flag can be turned back off, and -- ends the options
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Fixture file `demo.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST plus:true help:usage -- "Usage: demo [options]... [args]..." ''
  flag    FLAG    -f +f --{no-}flag
  disp    :usage  -h --help
}
```
_Fixture `demo.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
eval "set -- $REST"
echo "FLAG=[$FLAG] rest=[$*]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh
./demo.sh --flag --no-flag
./demo.sh -f +f
./demo.sh -- -f --no-such-option
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh`:
  - exit code is `0`
- after `./demo.sh --flag --no-flag`:
  - exit code is `0`
  - stdout equals an exact value
- after `./demo.sh -f +f`:
  - exit code is `0`
  - stdout equals an exact value
- after `./demo.sh -- -f --no-such-option`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
FLAG=[] rest=[]
```
_expected stdout:_
```text
FLAG=[] rest=[]
```
_expected stdout:_
```text
FLAG=[] rest=[-f --no-such-option]
```
### Scenario: an abbreviation is accepted while it is unambiguous
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Fixture file `demo.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage abbr:true -- "Usage: demo [options]..." ''
  flag    FLAG -f --{no-}flag
  param   NAME -n --name
  disp    :usage -h --help
}
```
_Fixture `demo.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
echo "FLAG=[$FLAG] NAME=[$NAME]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh
./demo.sh --nam world
./demo.sh --n world
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh`:
  - exit code is `0`
- after `./demo.sh --nam world`:
  - exit code is `0`
  - stdout equals an exact value
- after `./demo.sh --n world`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stdout:_
```text
FLAG=[] NAME=[world]
```
_expected stderr:_
```text
Ambiguous option: --n (could be --no-flag, --name)
```
### Scenario: the mistakes a user makes each have their own message
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Fixture file `demo.sh` is created.
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
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage -- "Usage: demo [options]..." ''
  param   NAME -n --name pattern:"alice | bob"
  disp    :usage -h --help
}
```
_Fixture `demo.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
echo "NAME=[$NAME]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh
./demo.sh --nope
./demo.sh --name
./demo.sh --name carol
./demo.sh --name bob
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh`:
  - exit code is `0`
- after `./demo.sh --nope`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `./demo.sh --name`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `./demo.sh --name carol`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `./demo.sh --name bob`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stderr:_
```text
Unrecognized option: --nope
```
_expected stderr:_
```text
Requires an argument: --name
```
_expected stderr:_
```text
Does not match the pattern (alice | bob): --name
```
_expected stdout:_
```text
NAME=[bob]
```
### Scenario: a custom validator decides, and says so when it refuses
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Fixture file `demo.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage -- "Usage: demo [options]..." ''
  param   PORT -p --port validate:"is_port" -- "1024-65535"
  disp    :usage -h --help
}
```
_Fixture `demo.sh`:_
```text
#!/bin/sh
set -eu
is_port() {
  case $OPTARG in
    (*[!0-9]*) return 1 ;;
  esac
  [ "$OPTARG" -ge 1024 ] && [ "$OPTARG" -le 65535 ]
}
. ./parser.sh
parse "$@"
echo "PORT=[$PORT]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh
./demo.sh --port 8080
./demo.sh --port 80
./demo.sh --port http
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh`:
  - exit code is `0`
- after `./demo.sh --port 8080`:
  - exit code is `0`
  - stdout equals an exact value
- after `./demo.sh --port 80`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `./demo.sh --port http`:
  - exit code is `1`
  - stderr equals an exact value

#### Expected output
_expected stdout:_
```text
PORT=[8080]
```
_expected stderr:_
```text
Validation error (is_port:1): --port
```
_expected stderr:_
```text
Validation error (is_port:1): --port
```
### Scenario: a validator that is not defined is a validation error, not a crash
_only when `command -v gengetoptions` succeeds · skipped on Windows_
#### Given
- Fixture file `definition.sh` is created.
- Fixture file `demo.sh` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `definition.sh`:_
```text
parser_definition() {
  setup   REST help:usage -- "Usage: demo [options]..." ''
  param   PORT -p --port validate:"number"
  disp    :usage -h --help
}
```
_Fixture `demo.sh`:_
```text
#!/bin/sh
set -eu
. ./parser.sh
parse "$@"
echo "PORT=[$PORT]"
```
#### When
```shell
gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh
./demo.sh --port 8080
```
#### Then
- after `gengetoptions parser -f definition.sh parser_definition parse demo > parser.sh`:
  - exit code is `0`
- after `./demo.sh --port 8080`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `not found`, `Validation error (number:127): --port`
