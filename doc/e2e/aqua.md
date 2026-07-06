# atago Behavior Specs
## Summary
1 suite · 8 scenarios
## Contents
- [aqua (declarative CLI version manager)](#aqua-declarative-cli-version-manager) — 8 scenarios
  - [version prints without error](#scenario-version-prints-without-error)
  - [init writes an aqua.yaml with the standard registry](#scenario-init-writes-an-aquayaml-with-the-standard-registry)
  - [init is idempotent on an existing config](#scenario-init-is-idempotent-on-an-existing-config)
  - [policy init writes an aqua-policy.yaml](#scenario-policy-init-writes-an-aqua-policyyaml)
  - [root-dir prints the aqua root path](#scenario-root-dir-prints-the-aqua-root-path)
  - [completion generates a bash script](#scenario-completion-generates-a-bash-script)
  - [which reports an unknown command as not found](#scenario-which-reports-an-unknown-command-as-not-found)
  - [an unknown subcommand is a usage error](#scenario-an-unknown-subcommand-is-a-usage-error)
## aqua (declarative CLI version manager)
Source: `test/e2e/thirdparty/aqua/aqua.atago.yaml`
### Scenario: version prints without error
#### When
```shell
aqua version
```
#### Then
- exit code is `0`
- stdout matches `/\S/`
### Scenario: init writes an aqua.yaml with the standard registry
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
aqua init
```
#### Then
- exit code is `0`
- the step changed exactly created `aqua.yaml`, modified nothing, deleted nothing
- file `aqua.yaml` contains `registries:`, `type: standard`, `packages:`
### Scenario: init is idempotent on an existing config
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
aqua init
aqua init
```
#### Then
- after `aqua init`:
  - exit code is `0`
- after `aqua init`:
  - exit code is `0`
  - stderr contains `configuration file already exists`
  - the step changed exactly created nothing, modified nothing, deleted nothing
### Scenario: policy init writes an aqua-policy.yaml
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
aqua policy init
```
#### Then
- exit code is `0`
- the step changed exactly created `aqua-policy.yaml`, modified nothing, deleted nothing
- file `aqua-policy.yaml` contains `registries:`
### Scenario: root-dir prints the aqua root path
#### When
```shell
aqua root-dir
```
#### Then
- exit code is `0`
- stdout contains `aquaproj-aqua`
### Scenario: completion generates a bash script
#### When
```shell
aqua completion bash
```
#### Then
- exit code is `0`
- stdout contains `completion script`
### Scenario: which reports an unknown command as not found
#### Given
- Fixture file `aqua.yaml` is created.
#### Inputs
_Fixture `aqua.yaml`:_
```text
registries:
- type: standard
  ref: v4.531.0
packages:
```
#### When
```shell
aqua which definitely-not-a-real-tool-xyz
```
#### Then
- exit code is `1`
- stderr contains `command is not found`
### Scenario: an unknown subcommand is a usage error
#### When
```shell
aqua bogus-subcommand-xyz
```
#### Then
- exit code is `3`
- stderr contains `No help topic`