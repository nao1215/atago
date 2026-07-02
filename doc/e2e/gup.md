# atago Behavior Specs
## Summary
2 suites · 4 scenarios
## Contents
- [gup list (isolated empty environment)](#gup-list-isolated-empty-environment) — 2 scenarios
  - [reports a friendly note on an empty environment (#350)](#scenario-reports-a-friendly-note-on-an-empty-environment-350)
  - [emits a valid empty JSON array on an empty environment with --json](#scenario-emits-a-valid-empty-json-array-on-an-empty-environment-with---json)
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