# atago Behavior Specs
## Summary
1 suite · 43 scenarios
## Contents
- [yazi (third-party terminal file manager)](#yazi-third-party-terminal-file-manager) — 43 scenarios
  - [version prints a semantic version banner](#scenario-version-prints-a-semantic-version-banner)
  - [hidden files stay hidden by default](#scenario-hidden-files-stay-hidden-by-default)
  - [dot toggles hidden files into view](#scenario-dot-toggles-hidden-files-into-view)
  - [filter narrows the visible file list](#scenario-filter-narrows-the-visible-file-list)
  - [create makes a new file](#scenario-create-makes-a-new-file)
  - [create with a trailing slash makes a directory](#scenario-create-with-a-trailing-slash-makes-a-directory)
  - [rename changes the hovered file name and nothing else](#scenario-rename-changes-the-hovered-file-name-and-nothing-else)
  - [yank then paste copies a file into a sibling directory](#scenario-yank-then-paste-copies-a-file-into-a-sibling-directory)
  - [select-all then yank copies multiple files into a sibling directory](#scenario-select-all-then-yank-copies-multiple-files-into-a-sibling-directory)
  - [cut then paste moves a file into a sibling directory](#scenario-cut-then-paste-moves-a-file-into-a-sibling-directory)
  - [uppercase D permanently deletes the hovered file after confirmation](#scenario-uppercase-d-permanently-deletes-the-hovered-file-after-confirmation)
  - [canceling a yank leaves nothing to paste](#scenario-canceling-a-yank-leaves-nothing-to-paste)
  - [canceling a cut leaves nothing to paste](#scenario-canceling-a-cut-leaves-nothing-to-paste)
  - [select-all then cut moves multiple files into a sibling directory](#scenario-select-all-then-cut-moves-multiple-files-into-a-sibling-directory)
  - [select-all then permanent delete removes multiple files](#scenario-select-all-then-permanent-delete-removes-multiple-files)
  - [trash moves the hovered file into sandbox home trash](#scenario-trash-moves-the-hovered-file-into-sandbox-home-trash)
  - [trashing selected files moves both into sandbox home trash](#scenario-trashing-selected-files-moves-both-into-sandbox-home-trash)
  - [space-selected file is copied even when the cursor moves away](#scenario-space-selected-file-is-copied-even-when-the-cursor-moves-away)
  - [visual mode selects a range and copies both files](#scenario-visual-mode-selects-a-range-and-copies-both-files)
  - [lowercase p copies to item_1 when the destination already exists](#scenario-lowercase-p-copies-to-item_1-when-the-destination-already-exists)
  - [lowercase p moves to item_1 when the destination already exists](#scenario-lowercase-p-moves-to-item_1-when-the-destination-already-exists)
  - [uppercase P overwrites an existing file during copy](#scenario-uppercase-p-overwrites-an-existing-file-during-copy)
  - [uppercase P overwrites an existing file during move](#scenario-uppercase-p-overwrites-an-existing-file-during-move)
  - [alphabetical sort puts a then b then c at the cursor](#scenario-alphabetical-sort-puts-a-then-b-then-c-at-the-cursor)
  - [reverse alphabetical sort puts c then b then a at the cursor](#scenario-reverse-alphabetical-sort-puts-c-then-b-then-a-at-the-cursor)
  - [extension sort puts log then md then txt at the cursor](#scenario-extension-sort-puts-log-then-md-then-txt-at-the-cursor)
  - [reverse extension sort puts txt then md then log at the cursor](#scenario-reverse-extension-sort-puts-txt-then-md-then-log-at-the-cursor)
  - [size sort puts one then two then three at the cursor](#scenario-size-sort-puts-one-then-two-then-three-at-the-cursor)
  - [reverse size sort puts three then two then one at the cursor](#scenario-reverse-size-sort-puts-three-then-two-then-one-at-the-cursor)
  - [natural sort puts 3 then 20 then 100 at the cursor](#scenario-natural-sort-puts-3-then-20-then-100-at-the-cursor)
  - [reverse natural sort puts 100 then 20 then 3 at the cursor](#scenario-reverse-natural-sort-puts-100-then-20-then-3-at-the-cursor)
  - [chooser-file writes the hovered path and exits on enter](#scenario-chooser-file-writes-the-hovered-path-and-exits-on-enter)
  - [chooser-file writes all selected paths on enter](#scenario-chooser-file-writes-all-selected-paths-on-enter)
  - [chooser-file remains absent when quitting with uppercase Q](#scenario-chooser-file-remains-absent-when-quitting-with-uppercase-q)
  - [uppercase G moves to the bottom item before choosing](#scenario-uppercase-g-moves-to-the-bottom-item-before-choosing)
  - [gg moves to the top item before choosing](#scenario-gg-moves-to-the-top-item-before-choosing)
  - [quitting immediately writes the initial cwd to cwd-file](#scenario-quitting-immediately-writes-the-initial-cwd-to-cwd-file)
  - [quitting after entering a directory writes the final cwd](#scenario-quitting-after-entering-a-directory-writes-the-final-cwd)
  - [leaving a directory then quitting writes the parent cwd](#scenario-leaving-a-directory-then-quitting-writes-the-parent-cwd)
  - [tabs preserve per-tab cwd and q writes the active tab cwd](#scenario-tabs-preserve-per-tab-cwd-and-q-writes-the-active-tab-cwd)
  - [closing the current tab returns to the previous tab cwd](#scenario-closing-the-current-tab-returns-to-the-previous-tab-cwd)
  - [previous and next tab navigation preserve each tab cwd](#scenario-previous-and-next-tab-navigation-preserve-each-tab-cwd)
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
### Scenario: create with a trailing slash makes a directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
- dir `nested` exists
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
### Scenario: select-all then yank copies multiple files into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/b.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
_Fixture `src/b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/a.txt`, `dst/b.txt`, modified nothing, deleted nothing
- file `dst/a.txt` contains `a`
- file `dst/b.txt` contains `b`
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
### Scenario: uppercase D permanently deletes the hovered file after confirmation
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed.txt`:_
```text
doomed
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified nothing, deleted `doomed.txt`
- file `doomed.txt` does not exist
### Scenario: canceling a yank leaves nothing to paste
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/one.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/one.txt`:_
```text
one
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified nothing, deleted nothing
### Scenario: canceling a cut leaves nothing to paste
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/one.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/one.txt`:_
```text
one
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified nothing, deleted nothing
### Scenario: select-all then cut moves multiple files into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/b.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
_Fixture `src/b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `dst/a.txt` contains `a`
- file `dst/b.txt` contains `b`
- file `src/a.txt` does not exist
- file `src/b.txt` does not exist
### Scenario: select-all then permanent delete removes multiple files
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `a.txt`:_
```text
a
```
_Fixture `b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
- file `a.txt` does not exist
- file `b.txt` does not exist
### Scenario: trash moves the hovered file into sandbox home trash
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `trashme.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `trashme.txt`:_
```text
trash
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
- file `trashme.txt` does not exist
- file `.atago-home/.local/share/Trash/files/trashme.txt` contains `trash`
### Scenario: trashing selected files moves both into sandbox home trash
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `a.txt`:_
```text
a
```
_Fixture `b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
- file `a.txt` does not exist
- file `b.txt` does not exist
- file `.atago-home/.local/share/Trash/files/a.txt` contains `a`
- file `.atago-home/.local/share/Trash/files/b.txt` contains `b`
### Scenario: space-selected file is copied even when the cursor moves away
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/b.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
_Fixture `src/b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `dst/b.txt` contains `b`
- file `dst/a.txt` does not exist
### Scenario: visual mode selects a range and copies both files
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/b.txt` is created.
- Fixture file `src/c.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
_Fixture `src/b.txt`:_
```text
b
```
_Fixture `src/c.txt`:_
```text
c
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `dst/a.txt` contains `a`
- file `dst/b.txt` contains `b`
- file `dst/c.txt` does not exist
### Scenario: lowercase p copies to item_1 when the destination already exists
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/item.txt` is created.
- Fixture file `dst/item.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/item.txt`:_
```text
new
```
_Fixture `dst/item.txt`:_
```text
old
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `src/item.txt` contains `new`
- file `dst/item.txt` contains `old`
- file `dst/item_1.txt` contains `new`
### Scenario: lowercase p moves to item_1 when the destination already exists
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/item.txt` is created.
- Fixture file `dst/item.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/item.txt`:_
```text
new
```
_Fixture `dst/item.txt`:_
```text
old
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `src/item.txt` does not exist
- file `dst/item.txt` contains `old`
- file `dst/item_1.txt` contains `new`
### Scenario: uppercase P overwrites an existing file during copy
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/item.txt` is created.
- Fixture file `dst/item.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/item.txt`:_
```text
new
```
_Fixture `dst/item.txt`:_
```text
old
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `src/item.txt` contains `new`
- file `dst/item.txt` contains `new`
### Scenario: uppercase P overwrites an existing file during move
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/item.txt` is created.
- Fixture file `dst/item.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/item.txt`:_
```text
new
```
_Fixture `dst/item.txt`:_
```text
old
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- exit code is `0`
- file `src/item.txt` does not exist
- file `dst/item.txt` contains `new`
### Scenario: alphabetical sort puts a then b then c at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `b.txt` is created.
- Fixture file `a.txt` is created.
- Fixture file `c.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: reverse alphabetical sort puts c then b then a at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `b.txt` is created.
- Fixture file `a.txt` is created.
- Fixture file `c.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: extension sort puts log then md then txt at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `b.md` is created.
- Fixture file `a.txt` is created.
- Fixture file `c.log` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: reverse extension sort puts txt then md then log at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `b.md` is created.
- Fixture file `a.txt` is created.
- Fixture file `c.log` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: size sort puts one then two then three at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `one.txt` is created.
- Fixture file `three.txt` is created.
- Fixture file `two.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `one.txt`:_
```text
1
```
_Fixture `three.txt`:_
```text
333
```
_Fixture `two.txt`:_
```text
22
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: reverse size sort puts three then two then one at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `one.txt` is created.
- Fixture file `three.txt` is created.
- Fixture file `two.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `one.txt`:_
```text
1
```
_Fixture `three.txt`:_
```text
333
```
_Fixture `two.txt`:_
```text
22
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: natural sort puts 3 then 20 then 100 at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `20.txt` is created.
- Fixture file `3.txt` is created.
- Fixture file `100.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: reverse natural sort puts 100 then 20 then 3 at the cursor
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `20.txt` is created.
- Fixture file `3.txt` is created.
- Fixture file `100.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- exit code is `0`
### Scenario: chooser-file writes the hovered path and exits on enter
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `picked.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `picked.txt`:_
```text
picked
```
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- exit code is `0`
- file `chosen.txt` contains `picked.txt`
### Scenario: chooser-file writes all selected paths on enter
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `a.txt`:_
```text
a
```
_Fixture `b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- exit code is `0`
- file `chosen.txt` contains `a.txt`, `b.txt`
### Scenario: chooser-file remains absent when quitting with uppercase Q
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `picked.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `picked.txt`:_
```text
picked
```
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- exit code is `0`
- file `chosen.txt` does not exist
### Scenario: uppercase G moves to the bottom item before choosing
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- Fixture file `z.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- exit code is `0`
- file `chosen.txt` contains `z.txt`
### Scenario: gg moves to the top item before choosing
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- Fixture file `z.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- exit code is `0`
- file `chosen.txt` contains `a.txt`
### Scenario: quitting immediately writes the initial cwd to cwd-file
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `alpha.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` contains `/atago-`
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
### Scenario: leaving a directory then quitting writes the parent cwd
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a-dir/a.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` is checked
### Scenario: tabs preserve per-tab cwd and q writes the active tab cwd
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a-dir/a.txt` is created.
- Fixture file `z-dir/z.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` contains `a-dir`
### Scenario: closing the current tab returns to the previous tab cwd
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a-dir/a.txt` is created.
- Fixture file `z-dir/z.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` contains `a-dir`
### Scenario: previous and next tab navigation preserve each tab cwd
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a-dir/a.txt` is created.
- Fixture file `z-dir/z.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- exit code is `0`
- file `cwd.txt` contains `z-dir`
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
