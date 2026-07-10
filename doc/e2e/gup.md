# atago Behavior Specs
## Summary
3 suites · 6 scenarios
## Contents
- [gup list (isolated empty environment)](#gup-list-isolated-empty-environment) — 2 scenarios
  - [reports a friendly note on an empty environment (#350)](#scenario-reports-a-friendly-note-on-an-empty-environment-350)
  - [emits a valid empty JSON array on an empty environment with --json](#scenario-emits-a-valid-empty-json-array-on-an-empty-environment-with---json)
- [gup read-only subcommands (no HOME writes)](#gup-read-only-subcommands-no-home-writes) — 2 scenarios
  - [gup version writes nothing to the workdir or its sandbox home](#scenario-gup-version-writes-nothing-to-the-workdir-or-its-sandbox-home)
  - [gup list is a pure read of GOBIN and writes nothing](#scenario-gup-list-is-a-pure-read-of-gobin-and-writes-nothing)
- [gup](#gup) — 2 scenarios
  - [version reports a semver string](#scenario-version-reports-a-semver-string)
  - [help describes the tool and exits zero](#scenario-help-describes-the-tool-and-exits-zero)
## gup list (isolated empty environment)
Source: `test/e2e/tools/gup/list.atago.yaml`
### Scenario: reports a friendly note on an empty environment (#350)
#### Given
- Fixture file `emptybin/.keep` is created.
- Environment variables are set: GOBIN.
#### When
```shell
gup list
```
#### Then
- exit code is `0`
- stdout contains `no binaries are installed`
### Scenario: emits a valid empty JSON array on an empty environment with --json
#### Given
- Fixture file `emptybin/.keep` is created.
- Environment variables are set: GOBIN.
#### When
```shell
gup list --json
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout at `$` has length 0
## gup read-only subcommands (no HOME writes)
Source: `test/e2e/tools/gup/sandbox_home.atago.yaml`
### Scenario: gup version writes nothing to the workdir or its sandbox home
_only when `gup version` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
gup version
```
#### Then
- exit code is `0`
- stdout contains `gup version`
- the step changed exactly created nothing, modified nothing, deleted nothing
### Scenario: gup list is a pure read of GOBIN and writes nothing
_only when `gup list` succeeds · skipped on Windows_
#### Given
- Fixture file `emptybin/.keep` is created.
- Environment variables are set: GOBIN.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
gup list
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified nothing, deleted nothing
## gup
Source: `test/e2e/tools/gup/smoke.atago.yaml`
### Scenario: version reports a semver string
#### When
```shell
gup version
```
#### Then
- exit code is `0`
- stdout matches `/gup version v[0-9]+\.[0-9]+\.[0-9]+/`
- stderr is empty
### Scenario: help describes the tool and exits zero
#### When
```shell
gup help
```
#### Then
- exit code is `0`
- stdout contains `gup`
