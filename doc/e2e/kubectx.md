# atago Behavior Specs
## Summary
2 suites · 31 scenarios
## Contents
- [kubectx (third-party CLI, kubeconfig state machine)](#kubectx-third-party-cli-kubeconfig-state-machine) — 18 scenarios
  - [usage is printed for --help and -h](#scenario-usage-is-printed-for---help-and--h)
  - [listing without a kubeconfig warns on stderr and still succeeds](#scenario-listing-without-a-kubeconfig-warns-on-stderr-and-still-succeeds)
  - [one context is listed and listing changes nothing](#scenario-one-context-is-listed-and-listing-changes-nothing)
  - [both contexts are listed, one per line](#scenario-both-contexts-are-listed-one-per-line)
  - [switching writes the choice into the kubeconfig and reports it on stderr](#scenario-switching-writes-the-choice-into-the-kubeconfig-and-reports-it-on-stderr)
  - [the dash toggles between the last two contexts](#scenario-the-dash-toggles-between-the-last-two-contexts)
  - [the dash fails when no previous context was recorded](#scenario-the-dash-fails-when-no-previous-context-was-recorded)
  - [an unknown context name fails and leaves the kubeconfig untouched](#scenario-an-unknown-context-name-fails-and-leaves-the-kubeconfig-untouched)
  - [--current fails while no context is selected](#scenario---current-fails-while-no-context-is-selected)
  - [renaming replaces the old name in the listing](#scenario-renaming-replaces-the-old-name-in-the-listing)
  - [a dot renames the current context](#scenario-a-dot-renames-the-current-context)
  - [deleting a context removes it from the listing](#scenario-deleting-a-context-removes-it-from-the-listing)
  - [a dot deletes the current context](#scenario-a-dot-deletes-the-current-context)
  - [deleting an unknown context fails and keeps the file intact](#scenario-deleting-an-unknown-context-fails-and-keeps-the-file-intact)
  - [several contexts are deleted in one call](#scenario-several-contexts-are-deleted-in-one-call)
  - [a batch delete naming one unknown context fails after deleting the rest](#scenario-a-batch-delete-naming-one-unknown-context-fails-after-deleting-the-rest)
  - [unsetting clears the selection](#scenario-unsetting-clears-the-selection)
  - [the interactive picker selects a context by typed query](#scenario-the-interactive-picker-selects-a-context-by-typed-query)
- [kubens (third-party CLI, namespace state machine)](#kubens-third-party-cli-namespace-state-machine) — 13 scenarios
  - [usage is printed for --help and -h](#scenario-usage-is-printed-for---help-and--h-1)
  - [listing without a kubeconfig fails](#scenario-listing-without-a-kubeconfig-fails)
  - [the namespaces of the current context are listed](#scenario-the-namespaces-of-the-current-context-are-listed)
  - [switching writes the namespace into the current context](#scenario-switching-writes-the-namespace-into-the-current-context)
  - [an unknown namespace is refused and changes nothing](#scenario-an-unknown-namespace-is-refused-and-changes-nothing)
  - [--force accepts a namespace the cluster does not have](#scenario---force-accepts-a-namespace-the-cluster-does-not-have)
  - [the dash toggles between the last two namespaces](#scenario-the-dash-toggles-between-the-last-two-namespaces)
  - [the dash fails when this context has no previous namespace](#scenario-the-dash-fails-when-this-context-has-no-previous-namespace)
  - [the dash fails when no context is selected](#scenario-the-dash-fails-when-no-context-is-selected)
  - [--current reports default before any namespace is chosen](#scenario---current-reports-default-before-any-namespace-is-chosen)
  - [--current reports the namespace after it is set](#scenario---current-reports-the-namespace-after-it-is-set)
  - [--current fails when no context is selected](#scenario---current-fails-when-no-context-is-selected)
  - [unsetting returns the context to the default namespace](#scenario-unsetting-returns-the-context-to-the-default-namespace)

## kubectx (third-party CLI, kubeconfig state machine)
[kubectx](https://github.com/ahmetb/kubectx) switches the active kubectl
context. It is a small program with a large contract: it rewrites the
kubeconfig in place, remembers the previous context in a file under the
user's home so that `kubectx -` can toggle, splits machine-readable output
(context names) from human status lines, and answers with exit 1 for every
mistake a user can make with a context name.

kubectx keeps its own test suite in Bats, and every one of those tests is
reproduced here, so the two can be compared. The scenarios then go past what
a Bats suite observes: which stream each line went to, what the kubeconfig
on disk looks like afterwards, that a failed command leaves it untouched,
that the previous-context state lands under the isolated home rather than
the real one, and what the interactive picker does when a terminal is
attached.

No cluster is contacted. Every kubeconfig here is a throwaway fixture
written in the scenario workdir, naming servers that are never dialed.

Source: `test/e2e/thirdparty/kubectx/kubectx.atago.yaml`
### Scenario: usage is printed for --help and -h
_only when `kubectx --help` succeeds_
#### Given
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
kubectx --help
kubectx -h
```
#### Then
- after `kubectx --help`:
  - exit code is `0`
  - stdout contains `USAGE:`, `kubectx <NEW_NAME>=<NAME>`
- after `kubectx -h`:
  - exit code is `0`
  - stdout contains `USAGE:`

### Scenario: listing without a kubeconfig warns on stderr and still succeeds
_only when `kubectx --help` succeeds_
#### Given
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
kubectx
```
#### Then
- exit code is `0`
- stdout is empty
- stderr equals an exact value
- file `config` does not exist

#### Expected output
_expected stderr:_
```text
warning: kubeconfig file not found
```
### Scenario: one context is listed and listing changes nothing
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubectx
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stderr is empty
- the step changed exactly created nothing, modified nothing, deleted nothing, ignoring `.atago-home/**`

#### Expected output
_expected stdout:_
```text
user1@cluster1
```
### Scenario: both contexts are listed, one per line
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
user1@cluster1
user2@cluster1
```
### Scenario: switching writes the choice into the kubeconfig and reports it on stderr
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx user1@cluster1
kubectx user2@cluster1
kubectx -c
kubectx --current
```
#### Then
- after `kubectx user1@cluster1`:
  - exit code is `0`
  - stdout is empty
  - stderr contains `Switched to context "user1@cluster1".`
  - file `config` contains `current-context: "user1@cluster1"`
- after `kubectx user2@cluster1`:
  - exit code is `0`
  - file `config` contains `current-context: "user2@cluster1"`
- after `kubectx -c`:
  - exit code is `0`
  - stdout equals an exact value
- after `kubectx --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
user2@cluster1
```
_expected stdout:_
```text
user2@cluster1
```
### Scenario: the dash toggles between the last two contexts
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx user1@cluster1
kubectx user2@cluster1
kubectx -
kubectx --current
kubectx -
kubectx --current
```
#### Then
- after `kubectx user1@cluster1`:
  - exit code is `0`
- after `kubectx user2@cluster1`:
  - exit code is `0`
  - file `.atago-home/.cache/kubectx` equals exact bytes
- after `kubectx -`:
  - exit code is `0`
  - stderr contains `Switched to context "user1@cluster1".`
- after `kubectx --current`:
  - exit code is `0`
  - stdout equals an exact value
- after `kubectx -`:
  - exit code is `0`
- after `kubectx --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
user1@cluster1
```
_expected stdout:_
```text
user2@cluster1
```
### Scenario: the dash fails when no previous context was recorded
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubectx -
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `no previous context found`
- the step changed exactly modified nothing, deleted nothing, ignoring `.atago-home/**`

### Scenario: an unknown context name fails and leaves the kubeconfig untouched
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubectx unknown-context
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `no context exists with the name: "unknown-context"`
- the step changed exactly modified nothing, deleted nothing, ignoring `.atago-home/**`

### Scenario: --current fails while no context is selected
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubectx -c
kubectx --current
```
#### Then
- after `kubectx -c`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `current-context is not set`
- after `kubectx --current`:
  - exit code is `1`

### Scenario: renaming replaces the old name in the listing
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx new-context=user1@cluster1
kubectx
```
#### Then
- after `kubectx new-context=user1@cluster1`:
  - exit code is `0`
  - stderr contains `Context user1@cluster1 renamed to new-context.`
- after `kubectx`:
  - exit code is `0`
  - stdout equals an exact value
  - file `config` does not contain `user1@cluster1`

#### Expected output
_expected stdout:_
```text
new-context
user2@cluster1
```
### Scenario: a dot renames the current context
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx user2@cluster1
kubectx new-context=.
kubectx
kubectx --current
```
#### Then
- after `kubectx user2@cluster1`:
  - exit code is `0`
- after `kubectx new-context=.`:
  - exit code is `0`
- after `kubectx`:
  - exit code is `0`
  - stdout equals an exact value
- after `kubectx --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
new-context
user1@cluster1
```
_expected stdout:_
```text
new-context
```
### Scenario: deleting a context removes it from the listing
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx -d user1@cluster1
kubectx
```
#### Then
- after `kubectx -d user1@cluster1`:
  - exit code is `0`
  - stderr contains `Deleted context user1@cluster1.`
- after `kubectx`:
  - exit code is `0`
  - stdout equals an exact value
  - file `config` contains `name: user1`, `name: cluster1`

#### Expected output
_expected stdout:_
```text
user2@cluster1
```
### Scenario: a dot deletes the current context
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx user2@cluster1
kubectx -d .
kubectx
```
#### Then
- after `kubectx user2@cluster1`:
  - exit code is `0`
- after `kubectx -d .`:
  - exit code is `0`
- after `kubectx`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
user1@cluster1
```
### Scenario: deleting an unknown context fails and keeps the file intact
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubectx -d unknown-context
```
#### Then
- exit code is `1`
- stderr contains `context does not exist`
- the step changed exactly modified nothing, deleted nothing, ignoring `.atago-home/**`

### Scenario: several contexts are deleted in one call
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx -d user1@cluster1 user2@cluster1
kubectx
```
#### Then
- after `kubectx -d user1@cluster1 user2@cluster1`:
  - exit code is `0`
- after `kubectx`:
  - exit code is `0`
  - stdout is empty

### Scenario: a batch delete naming one unknown context fails after deleting the rest
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx -d user1@cluster1 non-existent user2@cluster1
kubectx
```
#### Then
- after `kubectx -d user1@cluster1 non-existent user2@cluster1`:
  - exit code is `1`
- after `kubectx`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
user2@cluster1
```
### Scenario: unsetting clears the selection
_only when `kubectx --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
kubectx user1@cluster1
kubectx -u
kubectx --current
kubectx
```
#### Then
- after `kubectx user1@cluster1`:
  - exit code is `0`
- after `kubectx -u`:
  - exit code is `0`
  - stderr contains `Active context unset`
- after `kubectx --current`:
  - exit code is `1`
  - stderr contains `current-context is not set`
- after `kubectx`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
user1@cluster1
user2@cluster1
```
### Scenario: the interactive picker selects a context by typed query
_only when `fzf --version` succeeds · skipped on Windows_
#### Given
- Fixture file `config` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
  - name: user2
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
  - name: user2@cluster1
    context: {cluster: cluster1, user: user2}
current-context: ""
```
#### When
```shell
# interactive (pty): kubectx
```
#### Then
- file `config` contains `current-context: "user2@cluster1"`

## kubens (third-party CLI, namespace state machine)
[kubens](https://github.com/ahmetb/kubectx) is kubectx's sibling: it sets
the namespace of the current kubectl context. The contract has one more
dimension than kubectx's, because a namespace only means anything relative
to a context — the previous-namespace memory is kept per context, and every
command fails when no context is selected at all.

kubens ships its own Bats suite, and every test in it is reproduced here.
The scenarios then pin what that suite leaves open: which stream each line
goes to, that the chosen namespace really lands in the kubeconfig on disk,
that a rejected name changes nothing, and what `--force` does with a
namespace the cluster does not have.

Listing namespaces would need a cluster, so the suite sets the same
short-circuit environment variable kubens' own tests use and never opens a
socket.

Source: `test/e2e/thirdparty/kubectx/kubens.atago.yaml`
### Scenario: usage is printed for --help and -h
_only when `kubens --help` succeeds_
#### Given
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
kubens --help
kubens -h
```
#### Then
- after `kubens --help`:
  - exit code is `0`
  - stdout contains `USAGE:`, `kubens <NAME>`
- after `kubens -h`:
  - exit code is `0`
  - stdout contains `USAGE:`

### Scenario: listing without a kubeconfig fails
_only when `kubens --help` succeeds_
#### Given
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
kubens
```
#### Then
- exit code is `1`
- stdout is empty

### Scenario: the namespaces of the current context are listed
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens
```
#### Then
- exit code is `0`
- stdout contains `ns1`, `ns2`
- stderr is empty

### Scenario: switching writes the namespace into the current context
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens ns1
kubens --current
```
#### Then
- after `kubens ns1`:
  - exit code is `0`
  - stdout is empty
  - stderr contains `Active namespace is "ns1"`
  - the step changed exactly created nothing, modified `config`, deleted nothing, ignoring `.atago-home/**`
  - file `config` contains `namespace: ns1`
- after `kubens --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
ns1
```
### Scenario: an unknown namespace is refused and changes nothing
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens unknown-namespace
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `no namespace exists with name "unknown-namespace"`
- the step changed exactly modified nothing, deleted nothing, ignoring `.atago-home/**`

### Scenario: --force accepts a namespace the cluster does not have
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens --force not-yet-created
```
#### Then
- exit code is `0`
- stderr contains `Active namespace is "not-yet-created"`
- file `config` contains `namespace: not-yet-created`

### Scenario: the dash toggles between the last two namespaces
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens ns1
kubens ns2
kubens -
kubens --current
kubens -
kubens --current
```
#### Then
- after `kubens ns1`:
  - exit code is `0`
- after `kubens ns2`:
  - exit code is `0`
  - file `.atago-home/.cache/kubens/user1@cluster1` equals exact bytes
- after `kubens -`:
  - exit code is `0`
- after `kubens --current`:
  - exit code is `0`
  - stdout equals an exact value
- after `kubens -`:
  - exit code is `0`
- after `kubens --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
ns1
```
_expected stdout:_
```text
ns2
```
### Scenario: the dash fails when this context has no previous namespace
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens -
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `No previous namespace found for current context`

### Scenario: the dash fails when no context is selected
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubens -
```
#### Then
- exit code is `1`
- stderr contains `current-context is not set`

### Scenario: --current reports default before any namespace is chosen
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens -c
kubens --current
```
#### Then
- after `kubens -c`:
  - exit code is `0`
  - stdout equals an exact value
- after `kubens --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
default
```
_expected stdout:_
```text
default
```
### Scenario: --current reports the namespace after it is set
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens ns1
kubens -c
kubens --current
```
#### Then
- after `kubens ns1`:
  - exit code is `0`
- after `kubens -c`:
  - exit code is `0`
  - stdout equals an exact value
- after `kubens --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
ns1
```
_expected stdout:_
```text
ns1
```
### Scenario: --current fails when no context is selected
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: ""
```
#### When
```shell
kubens -c
kubens --current
```
#### Then
- after `kubens -c`:
  - exit code is `1`
- after `kubens --current`:
  - exit code is `1`

### Scenario: unsetting returns the context to the default namespace
_only when `kubens --help` succeeds_
#### Given
- Fixture file `config` is created.
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: KUBECONFIG, _MOCK_NAMESPACES.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config`:_
```text
apiVersion: v1
kind: Config
clusters:
  - name: cluster1
    cluster: {server: https://cluster1.example.com}
users:
  - name: user1
    user: {}
contexts:
  - name: user1@cluster1
    context: {cluster: cluster1, user: user1}
current-context: "user1@cluster1"
```
#### When
```shell
kubens ns2
kubens -u
kubens --current
```
#### Then
- after `kubens ns2`:
  - exit code is `0`
- after `kubens -u`:
  - exit code is `0`
  - stderr contains `Active namespace is "default"`
- after `kubens --current`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
default
```
