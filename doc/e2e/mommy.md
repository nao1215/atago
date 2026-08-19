# atago Behavior Specs
## Summary
1 suite · 8 scenarios
## Contents
- [mommy (a wrapper that must stay out of the way)](#mommy-a-wrapper-that-must-stay-out-of-the-way) — 8 scenarios
  - [the version and the manual are answered without running anything](#scenario-the-version-and-the-manual-are-answered-without-running-anything)
  - [the wrapped command's status comes back untouched](#scenario-the-wrapped-commands-status-comes-back-untouched)
  - [the message stays off stdout unless it is asked for](#scenario-the-message-stays-off-stdout-unless-it-is-asked-for)
  - [a status can be judged without running anything at all](#scenario-a-status-can-be-judged-without-running-anything-at-all)
  - [pipefail is a choice, and it changes the verdict](#scenario-pipefail-is-a-choice-and-it-changes-the-verdict)
  - [the message is a template the config fills in](#scenario-the-message-is-a-template-the-config-fills-in)
  - [the environment is not the configuration](#scenario-the-environment-is-not-the-configuration)
  - [the toggle is a file in the user's state directory](#scenario-the-toggle-is-a-file-in-the-users-state-directory)

## mommy (a wrapper that must stay out of the way)
[mommy](https://github.com/FWDekker/mommy) runs your command and says
something encouraging about how it went. It keeps its own test suite in
ShellSpec; what those tests check is pinned here from outside.

A wrapper has two obligations before anything else: pass the exit status
through untouched, and stay off the wrapped command's stdout. Both are
asserted here for success, for failure, and for an arbitrary code, together
with the flag that deliberately moves the message onto stdout. The messages
themselves are made deterministic by a config file the scenario writes, so
the assertions are about which list was chosen rather than about which
sentence was drawn from it.

Source: `test/e2e/thirdparty/mommy/mommy.atago.yaml`
### Scenario: the version and the manual are answered without running anything
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
mommy --version
mommy --help
```
#### Then
- after `mommy --version`:
  - exit code is `0`
  - stdout matches `/^mommy, v[0-9]+\.[0-9]+\.[0-9]+/`
- after `mommy --help`:
  - exit code is `0`
  - stdout contains `mommy - here to support you~`, `-s status, --status=status`, `-1     writes output to stdout instead of stderr~`

### Scenario: the wrapped command's status comes back untouched
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_COMPLIMENTS="that went well"
MOMMY_ENCOURAGEMENTS="that did not go well"
```
#### When
```shell
mommy -c ./mommy.conf true
mommy -c ./mommy.conf false
mommy -c ./mommy.conf sh -c "exit 42"
```
#### Then
- after `mommy -c ./mommy.conf true`:
  - exit code is `0`
  - stdout is empty
  - stderr equals an exact value
- after `mommy -c ./mommy.conf false`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `mommy -c ./mommy.conf sh -c "exit 42"`:
  - exit code is `42`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
that went well~
```
_expected stderr:_
```text
that did not go well~
```
_expected stderr:_
```text
that did not go well~
```
### Scenario: the message stays off stdout unless it is asked for
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_COMPLIMENTS="well done"
```
#### When
```shell
mommy -c ./mommy.conf -e "echo the-real-output | wc -l"
mommy -c ./mommy.conf -1 true
```
#### Then
- after `mommy -c ./mommy.conf -e "echo the-real-output | wc -l"`:
  - exit code is `0`
  - stdout equals an exact value
  - stderr equals an exact value
- after `mommy -c ./mommy.conf -1 true`:
  - exit code is `0`
  - stdout equals an exact value
  - stderr is empty

#### Expected output
_expected stdout:_
```text
1
```
_expected stderr:_
```text
well done~
```
_expected stdout:_
```text
well done~
```
### Scenario: a status can be judged without running anything at all
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_COMPLIMENTS="that went well"
MOMMY_ENCOURAGEMENTS="that did not go well"
```
#### When
```shell
mommy -c ./mommy.conf -s 3 sh -c "touch never-created"
mommy -c ./mommy.conf -s 0
```
#### Then
- after `mommy -c ./mommy.conf -s 3 sh -c "touch never-created"`:
  - exit code is `3`
  - stderr equals an exact value
  - the step changed exactly created nothing, modified nothing, deleted nothing
  - file `never-created` does not exist
- after `mommy -c ./mommy.conf -s 0`:
  - exit code is `0`
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
that did not go well~
```
_expected stderr:_
```text
that went well~
```
### Scenario: pipefail is a choice, and it changes the verdict
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_COMPLIMENTS="that went well"
MOMMY_ENCOURAGEMENTS="that did not go well"
```
#### When
```shell
mommy -c ./mommy.conf -e "false | true"
mommy -c ./mommy.conf -p -e "false | true"
```
#### Then
- after `mommy -c ./mommy.conf -e "false | true"`:
  - exit code is `0`
  - stderr equals an exact value
- after `mommy -c ./mommy.conf -p -e "false | true"`:
  - exit code is `1`
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
that went well~
```
_expected stderr:_
```text
that did not go well~
```
### Scenario: the message is a template the config fills in
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_CAREGIVER="daddy"
MOMMY_SWEETIE="champ"
MOMMY_SUFFIX="!"
MOMMY_COMPLIMENTS="%%CAREGIVER%% is proud of %%SWEETIE%%"
```
#### When
```shell
mommy -c ./mommy.conf true
```
#### Then
- exit code is `0`
- stderr equals an exact value

#### Expected output
_expected stderr:_
```text
daddy is proud of champ!
```
### Scenario: the environment is not the configuration
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, MOMMY_COMPLIMENTS.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_COMPLIMENTS="from the config file"
```
#### When
```shell
mommy -c ./mommy.conf true
mommy -c ./mommy.conf true
```
#### Then
- after `mommy -c ./mommy.conf true`:
  - exit code is `0`
  - stderr equals an exact value
- after `mommy -c ./mommy.conf true`:
  - exit code is `0`
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
from the config file~
```
_expected stderr:_
```text
from the config file~
```
### Scenario: the toggle is a file in the user's state directory
_only when `command -v mommy` succeeds · skipped on Windows_
#### Given
- Fixture file `mommy.conf` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `mommy.conf`:_
```text
MOMMY_COMPLIMENTS="that went well"
MOMMY_ENCOURAGEMENTS="that did not go well"
```
#### When
```shell
mommy -t
mommy -c ./mommy.conf false
mommy -t
mommy -c ./mommy.conf true
```
#### Then
- after `mommy -t`:
  - exit code is `0`
  - stdout contains `mommy has been disabled for this user.`
  - stderr is empty
  - the step changed exactly created `.atago-home/.local/state/mommy/toggle`, modified nothing, deleted nothing
- after `mommy -c ./mommy.conf false`:
  - exit code is `0`
  - stdout is empty
  - stderr is empty
- after `mommy -t`:
  - exit code is `0`
  - stdout contains `mommy has been enabled for this user.`
- after `mommy -c ./mommy.conf true`:
  - exit code is `0`
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
that went well~
```
