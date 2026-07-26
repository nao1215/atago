# atago Behavior Specs
## Summary
1 suite · 10 scenarios
## Contents
- [yazi (third-party terminal file manager)](#yazi-third-party-terminal-file-manager) — 10 scenarios
  - [version prints a semantic version banner](#scenario-version-prints-a-semantic-version-banner)
  - [hidden files stay hidden by default](#scenario-hidden-files-stay-hidden-by-default)
  - [dot toggles hidden files into view](#scenario-dot-toggles-hidden-files-into-view)
  - [filter narrows the visible file list](#scenario-filter-narrows-the-visible-file-list)
  - [create makes a new file](#scenario-create-makes-a-new-file)
  - [rename changes the hovered file name and nothing else](#scenario-rename-changes-the-hovered-file-name-and-nothing-else)
  - [yank then paste copies a file into a sibling directory](#scenario-yank-then-paste-copies-a-file-into-a-sibling-directory)
  - [cut then paste moves a file into a sibling directory](#scenario-cut-then-paste-moves-a-file-into-a-sibling-directory)
  - [quitting after entering a directory writes the final cwd](#scenario-quitting-after-entering-a-directory-writes-the-final-cwd)
  - [uppercase Q exits without writing the cwd file](#scenario-uppercase-q-exits-without-writing-the-cwd-file)
## yazi (third-party terminal file manager)
Source: `test/e2e/thirdparty/yazi/yazi.atago.yaml`
### Scenario: version prints a semantic version banner
_only when `yazi --version` succeeds_
#### When
```shell
yazi --version
```
#### Then
- exit code is `0`
- stdout matches `/^Yazi [0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: hidden files stay hidden by default
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `alpha.txt` is created.
- Fixture file `.secret` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- rendered screen contains `alpha.txt`
- rendered screen does not contain `.secret`
### Scenario: dot toggles hidden files into view
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `alpha.txt` is created.
- Fixture file `.secret` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- rendered screen contains `alpha.txt`
- rendered screen contains `.secret`
### Scenario: filter narrows the visible file list
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `alpha.txt` is created.
- Fixture file `beta.txt` is created.
- Fixture file `gamma.log` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- rendered screen contains `beta.txt`
- rendered screen does not contain `alpha.txt`
- rendered screen does not contain `gamma.log`
### Scenario: create makes a new file
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
- the step changed exactly created `notes.txt`, modified nothing, deleted nothing
- file `notes.txt` exists
#### Generated artifacts
- `notes.txt`
### Scenario: rename changes the hovered file name and nothing else
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `old.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `old.txt`:_
```text
old
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- the step changed exactly created `old.bak.txt`, modified nothing, deleted `old.txt`
- file `old.txt` does not exist
- file `old.bak.txt` contains `old`
### Scenario: yank then paste copies a file into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/alpha.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/alpha.txt`:_
```text
alpha
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/alpha.txt`, modified nothing, deleted nothing
- file `src/alpha.txt` contains `alpha`
- file `dst/alpha.txt` contains `alpha`
### Scenario: cut then paste moves a file into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/beta.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/beta.txt`:_
```text
beta
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/beta.txt`, modified nothing, deleted `src/beta.txt`
- file `src/beta.txt` does not exist
- file `dst/beta.txt` contains `beta`
### Scenario: quitting after entering a directory writes the final cwd
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `z-nested/inside.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` contains `z-nested`
### Scenario: uppercase Q exits without writing the cwd file
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `z-nested/inside.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` does not exist
