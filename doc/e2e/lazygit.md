# atago Behavior Specs
## Summary
1 suite · 11 scenarios
## Contents
- [lazygit (third-party git TUI)](#lazygit-third-party-git-tui) — 11 scenarios
  - [version prints the pinned release](#scenario-version-prints-the-pinned-release)
  - [opening a dirty repository shows unstaged changes and quits cleanly](#scenario-opening-a-dirty-repository-shows-unstaged-changes-and-quits-cleanly)
  - [space stages the selected unstaged file](#scenario-space-stages-the-selected-unstaged-file)
  - [space toggles a staged file back to unstaged](#scenario-space-toggles-a-staged-file-back-to-unstaged)
  - [capital C commits staged changes through git's editor](#scenario-capital-c-commits-staged-changes-through-gits-editor)
  - [escape cancels the discard dialog and preserves the diff](#scenario-escape-cancels-the-discard-dialog-and-preserves-the-diff)
  - [enter accepts the discard dialog and reverts the diff](#scenario-enter-accepts-the-discard-dialog-and-reverts-the-diff)
  - [n creates a new branch from the branch panel](#scenario-n-creates-a-new-branch-from-the-branch-panel)
  - [space checks out the selected branch](#scenario-space-checks-out-the-selected-branch)
  - [s opens the stash prompt and records a named stash entry](#scenario-s-opens-the-stash-prompt-and-records-a-named-stash-entry)
  - [g then enter pops the selected stash entry back into the worktree](#scenario-g-then-enter-pops-the-selected-stash-entry-back-into-the-worktree)
## lazygit (third-party git TUI)
Source: `test/e2e/thirdparty/lazygit/lazygit.atago.yaml`
### Scenario: version prints the pinned release
_only when `lazygit --version` succeeds_
#### When
```shell
lazygit --version
```
#### Then
- exit code is `0`
- stdout contains `version=0.63.1`
### Scenario: opening a dirty repository shows unstaged changes and quits cleanly
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
```
#### Then
- after `printf 'more\n' >> repo/a.txt`:
  - exit code is `0`
### Scenario: space stages the selected unstaged file
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/b.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `repo/b.txt`:_
```text
new
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
git -C repo status --porcelain
```
#### Then
- after `git -C repo status --porcelain`:
  - exit code is `0`
  - stdout contains `M  a.txt`, `?? b.txt`
### Scenario: space toggles a staged file back to unstaged
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt && git -C repo add a.txt
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
git -C repo status --porcelain
```
#### Then
- after `git -C repo status --porcelain`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: capital C commits staged changes through git's editor
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `editor.sh` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `editor.sh`:_
```text
#!/bin/sh
printf 'editor commit\n' > "$1"
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
git -C repo config user.name atago
git -C repo config user.email atago@example.com
printf 'more\n' >> repo/a.txt
chmod +x editor.sh
git -C repo config core.editor ${workdir}/editor.sh
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
git -C repo log --oneline -1
```
#### Then
- after `git -C repo log --oneline -1`:
  - exit code is `0`
  - stdout contains `editor commit`
### Scenario: escape cancels the discard dialog and preserves the diff
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
git -C repo diff -- a.txt
```
#### Then
- after `git -C repo diff -- a.txt`:
  - exit code is `0`
  - stdout contains `+more`
### Scenario: enter accepts the discard dialog and reverts the diff
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
git -C repo diff -- a.txt
```
#### Then
- after `git -C repo diff -- a.txt`:
  - exit code is `0`
  - stdout is empty
### Scenario: n creates a new branch from the branch panel
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo branch
git -C repo branch --show-current
```
#### Then
- after `git -C repo branch --show-current`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: space checks out the selected branch
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
git -C repo branch feature-one
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo branch
git -C repo branch --show-current
```
#### Then
- after `git -C repo branch --show-current`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: s opens the stash prompt and records a named stash entry
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo
git -C repo stash list
git -C repo diff -- a.txt
```
#### Then
- after `git -C repo stash list`:
  - exit code is `0`
  - stdout contains `stash@{0}`, `probe stash`
- after `git -C repo diff -- a.txt`:
  - exit code is `0`
  - stdout is empty
### Scenario: g then enter pops the selected stash entry back into the worktree
_only when `lazygit --version` succeeds · skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `cfg/config.yml` is created.
- The command runs with a cleared environment (passing through: PATH).
#### Inputs
_Fixture `repo/a.txt`:_
```text
hello
```
_Fixture `cfg/config.yml`:_
```text
gui:
  nerdFontsVersion: "3"
  showFileTree: false
  showBottomLine: false
  showCommandLog: false
  showRandomTip: false
  showFileIcons: false
  showListFooter: false
disableStartupPopups: true
promptToReturnFromSubprocess: false
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m initial
printf 'more\n' >> repo/a.txt
git -C repo stash push -m probe
# interactive (pty): lazygit -ucd ${workdir}/cfg -p ${workdir}/repo stash
git -C repo stash list
git -C repo diff -- a.txt
```
#### Then
- after `git -C repo stash list`:
  - exit code is `0`
  - stdout is empty
- after `git -C repo diff -- a.txt`:
  - exit code is `0`
  - stdout contains `+more`
