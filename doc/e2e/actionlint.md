# atago Behavior Specs
## Summary
1 suite · 9 scenarios
## Contents
- [actionlint (GitHub Actions workflow linter)](#actionlint-github-actions-workflow-linter) — 9 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [a valid workflow passes silently](#scenario-a-valid-workflow-passes-silently)
  - [an undefined needs dependency is reported on stdout](#scenario-an-undefined-needs-dependency-is-reported-on-stdout)
  - [an unknown runner label is reported](#scenario-an-unknown-runner-label-is-reported)
  - [an invalid expression is reported](#scenario-an-invalid-expression-is-reported)
  - [multiple problems in one file are all reported](#scenario-multiple-problems-in-one-file-are-all-reported)
  - [the JSON format is a structured oracle over the findings](#scenario-the-json-format-is-a-structured-oracle-over-the-findings)
  - [-ignore removes a matching finding](#scenario--ignore-removes-a-matching-finding)
  - [stdin mode lints piped content under the given filename](#scenario-stdin-mode-lints-piped-content-under-the-given-filename)
## actionlint (GitHub Actions workflow linter)
Source: `test/e2e/thirdparty/actionlint/actionlint.atago.yaml`
### Scenario: version prints a semantic version
_only when `actionlint -version` succeeds_
#### When
```shell
actionlint -version
```
#### Then
- exit code is `0`
- stdout matches `/^v[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: a valid workflow passes silently
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `good.yml` is created.
#### Inputs
_Fixture `good.yml`:_
```text
name: good
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
```
#### When
```shell
actionlint -shellcheck= good.yml
```
#### Then
- exit code is `0`
- stdout is empty
- stderr is empty
### Scenario: an undefined needs dependency is reported on stdout
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `badneeds.yml` is created.
#### Inputs
_Fixture `badneeds.yml`:_
```text
name: badneeds
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    needs: [nonexistent]
    steps:
      - run: echo hi
```
#### When
```shell
actionlint -shellcheck= badneeds.yml
```
#### Then
- exit code is `1`
- stdout contains `needs job "nonexistent" which does not exist`, `[job-needs]`
- stderr is empty
### Scenario: an unknown runner label is reported
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `badlabel.yml` is created.
#### Inputs
_Fixture `badlabel.yml`:_
```text
name: badlabel
on: push
jobs:
  build:
    runs-on: [ubuntu-nonsense-9999]
    steps:
      - run: echo hi
```
#### When
```shell
actionlint -shellcheck= badlabel.yml
```
#### Then
- exit code is `1`
- stdout contains `label "ubuntu-nonsense-9999" is unknown`, `[runner-label]`
### Scenario: an invalid expression is reported
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `badexpr.yml` is created.
#### Inputs
_Fixture `badexpr.yml`:_
```text
name: badexpr
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo ${{ github.event.foo.bar( }}
```
#### When
```shell
actionlint -shellcheck= badexpr.yml
```
#### Then
- exit code is `1`
- stdout contains `[expression]`
### Scenario: multiple problems in one file are all reported
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `two.yml` is created.
#### Inputs
_Fixture `two.yml`:_
```text
name: two
on: push
jobs:
  build:
    runs-on: [ubuntu-nonsense-9999]
    needs: [ghost]
    steps:
      - run: echo hi
```
#### When
```shell
actionlint -shellcheck= two.yml
```
#### Then
- exit code is `1`
- stdout contains `[job-needs]`, `[runner-label]`
### Scenario: the JSON format is a structured oracle over the findings
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `two.yml` is created.
#### Inputs
_Fixture `two.yml`:_
```text
name: two
on: push
jobs:
  build:
    runs-on: [ubuntu-nonsense-9999]
    needs: [ghost]
    steps:
      - run: echo hi
```
#### When
```shell
actionlint -shellcheck= -format '{{json .}}' two.yml
```
#### Then
- exit code is `1`
- stdout at `$` has length 2
- stdout at `$[0].kind` equals `job-needs`
- stdout at `$[1].kind` equals `runner-label`
### Scenario: -ignore removes a matching finding
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `two.yml` is created.
#### Inputs
_Fixture `two.yml`:_
```text
name: two
on: push
jobs:
  build:
    runs-on: [ubuntu-nonsense-9999]
    needs: [ghost]
    steps:
      - run: echo hi
```
#### When
```shell
actionlint -shellcheck= -ignore 'label .* is unknown' two.yml
```
#### Then
- exit code is `1`
- stdout contains `[job-needs]`
- stdout does not contain `[runner-label]`
### Scenario: stdin mode lints piped content under the given filename
_only when `actionlint -version` succeeds_
#### Given
- Fixture file `piped.yml` is created.
#### Inputs
_Fixture `piped.yml`:_
```text
name: piped
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    needs: [ghost]
    steps:
      - run: echo hi
```
_stdin for `actionlint`:_
```text
(read from file piped.yml)
```
#### When
```shell
actionlint -shellcheck= -stdin-filename fromstdin.yml -
```
#### Then
- exit code is `1`
- stdout contains `fromstdin.yml`, `[job-needs]`
