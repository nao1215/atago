# atago Behavior Specs
## Summary
2 suites · 9 scenarios
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
- [aqua + install (declarative install of a real tool)](#aqua--install-declarative-install-of-a-real-tool) — 1 scenario
  - [install downloads the tool and makes it runnable](#scenario-install-downloads-the-tool-and-makes-it-runnable)
## aqua (declarative CLI version manager)
Source: `test/e2e/thirdparty/aqua/aqua.atago.yaml`
### Scenario: version prints without error
_only when `aqua version` succeeds_
#### When
```shell
aqua version
```
#### Then
- exit code is `0`
- stdout matches `/\S/`
### Scenario: init writes an aqua.yaml with the standard registry
_only when `aqua version` succeeds_
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
_only when `aqua version` succeeds_
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
_only when `aqua version` succeeds_
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
_only when `aqua version` succeeds_
#### When
```shell
aqua root-dir
```
#### Then
- exit code is `0`
- stdout contains `aquaproj-aqua`
### Scenario: completion generates a bash script
_only when `aqua version` succeeds_
#### When
```shell
aqua completion bash
```
#### Then
- exit code is `0`
- stdout contains `completion script`
### Scenario: which reports an unknown command as not found
_only when `aqua version` succeeds_
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
_only when `aqua version` succeeds_
#### When
```shell
aqua bogus-subcommand-xyz
```
#### Then
- exit code is `3`
- stderr contains `No help topic`
## aqua + install (declarative install of a real tool)
Source: `test/e2e/thirdparty/aqua/install.atago.yaml`
### Scenario: install downloads the tool and makes it runnable
_only when `aqua version` succeeds_
#### Given
- Background service `fileserver` is started: `python3 -m http.server 18595 --bind 127.0.0.1 --directory dist`.
- Fixture file `dist/mytool` is created.
- Fixture file `registry.yaml` is created.
- Fixture file `aqua.yaml` is created.
- Fixture file `aqua-policy.yaml` is created.
#### Inputs
_Fixture `dist/mytool`:_
```text
#!/bin/sh
echo installed-tool-ran-ok
```
_Fixture `registry.yaml`:_
```text
packages:
  - type: http
    name: mytool
    url: "http://127.0.0.1:18595/mytool"
    files:
      - name: mytool
```
_Fixture `aqua.yaml`:_
```text
checksum:
  enabled: false
registries:
  - type: local
    name: local
    path: registry.yaml
packages:
  - name: mytool@0.1.0
    registry: local
```
_Fixture `aqua-policy.yaml`:_
```text
---
registries:
  - name: local
    type: local
    path: registry.yaml
packages:
  - registry: local
```
#### When
```shell
aqua --config aqua.yaml install
aqua --config aqua.yaml which mytool
aqua --config aqua.yaml exec -- mytool
```
#### Then
- after `aqua --config aqua.yaml install`:
  - exit code is `0`
  - file `.aqua/pkgs/http/127.0.0.1:18595/mytool/mytool` is checked
- after `aqua --config aqua.yaml which mytool`:
  - exit code is `0`
  - stdout contains `.aqua/pkgs/http/127.0.0.1:18595/mytool/mytool`
- after `aqua --config aqua.yaml exec -- mytool`:
  - exit code is `0`
  - stdout contains `installed-tool-ran-ok`
