# atago Behavior Specs
## Summary
1 suite · 101 scenarios
## Contents
- [yazi (third-party terminal file manager)](#yazi-third-party-terminal-file-manager) — 101 scenarios
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
  - [canceling permanent delete returns to the file list and keeps the file](#scenario-canceling-permanent-delete-returns-to-the-file-list-and-keeps-the-file)
  - [canceling a yank leaves nothing to paste](#scenario-canceling-a-yank-leaves-nothing-to-paste)
  - [canceling a cut leaves nothing to paste](#scenario-canceling-a-cut-leaves-nothing-to-paste)
  - [select-all then cut moves multiple files into a sibling directory](#scenario-select-all-then-cut-moves-multiple-files-into-a-sibling-directory)
  - [select-all then permanent delete removes multiple files](#scenario-select-all-then-permanent-delete-removes-multiple-files)
  - [trash moves the hovered file into sandbox home trash](#scenario-trash-moves-the-hovered-file-into-sandbox-home-trash)
  - [trashing selected files moves both into sandbox home trash](#scenario-trashing-selected-files-moves-both-into-sandbox-home-trash)
  - [canceling trash returns to the file list and keeps the file](#scenario-canceling-trash-returns-to-the-file-list-and-keeps-the-file)
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
  - [natural sort is visible on the rendered file list before quitting](#scenario-natural-sort-is-visible-on-the-rendered-file-list-before-quitting)
  - [reverse natural sort puts 100 then 20 then 3 at the cursor](#scenario-reverse-natural-sort-puts-100-then-20-then-3-at-the-cursor)
  - [reverse natural sort is visible on the rendered file list before quitting](#scenario-reverse-natural-sort-is-visible-on-the-rendered-file-list-before-quitting)
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
  - [create accepts filenames with spaces](#scenario-create-accepts-filenames-with-spaces)
  - [create accepts directory names with spaces](#scenario-create-accepts-directory-names-with-spaces)
  - [rename accepts spaces in filenames](#scenario-rename-accepts-spaces-in-filenames)
  - [rename accepts spaces in directory names](#scenario-rename-accepts-spaces-in-directory-names)
  - [yank then dash creates an absolute symlink in the sibling directory](#scenario-yank-then-dash-creates-an-absolute-symlink-in-the-sibling-directory)
  - [yank then underscore creates a relative symlink in the sibling directory](#scenario-yank-then-underscore-creates-a-relative-symlink-in-the-sibling-directory)
  - [yanked directory can be absolute-symlinked into a sibling directory](#scenario-yanked-directory-can-be-absolute-symlinked-into-a-sibling-directory)
  - [yanked directory can be relative-symlinked into a sibling directory](#scenario-yanked-directory-can-be-relative-symlinked-into-a-sibling-directory)
  - [modified time sort reorders the rendered list](#scenario-modified-time-sort-reorders-the-rendered-list)
  - [reverse modified time sort reorders the rendered list](#scenario-reverse-modified-time-sort-reorders-the-rendered-list)
  - [modified time sort chooses the oldest file first](#scenario-modified-time-sort-chooses-the-oldest-file-first)
  - [reverse modified time sort chooses the newest file first](#scenario-reverse-modified-time-sort-chooses-the-newest-file-first)
  - [toggling the same file twice removes the explicit selection](#scenario-toggling-the-same-file-twice-removes-the-explicit-selection)
  - [inverse selection flips one picked file into the other two](#scenario-inverse-selection-flips-one-picked-file-into-the-other-two)
  - [direct tab switching with 1 and 2 preserves per-tab cwd](#scenario-direct-tab-switching-with-1-and-2-preserves-per-tab-cwd)
  - [block shell command creates a file and returns to the browser](#scenario-block-shell-command-creates-a-file-and-returns-to-the-browser)
  - [filter can be cleared after confirming it](#scenario-filter-can-be-cleared-after-confirming-it)
  - [yank then paste copies a directory subtree into a sibling directory](#scenario-yank-then-paste-copies-a-directory-subtree-into-a-sibling-directory)
  - [cut then paste moves a directory subtree into a sibling directory](#scenario-cut-then-paste-moves-a-directory-subtree-into-a-sibling-directory)
  - [lowercase p copies a directory to name_1 when the destination exists](#scenario-lowercase-p-copies-a-directory-to-name_1-when-the-destination-exists)
  - [lowercase p moves a directory to name_1 when the destination exists](#scenario-lowercase-p-moves-a-directory-to-name_1-when-the-destination-exists)
  - [uppercase D permanently deletes a directory subtree after confirmation](#scenario-uppercase-d-permanently-deletes-a-directory-subtree-after-confirmation)
  - [canceling uppercase D keeps a directory subtree intact](#scenario-canceling-uppercase-d-keeps-a-directory-subtree-intact)
  - [trash moves a directory subtree into sandbox home trash](#scenario-trash-moves-a-directory-subtree-into-sandbox-home-trash)
  - [canceling trash keeps a directory subtree intact](#scenario-canceling-trash-keeps-a-directory-subtree-intact)
  - [canceling trash for selected directories keeps both subtrees](#scenario-canceling-trash-for-selected-directories-keeps-both-subtrees)
  - [yank then paste copies a spaced filename into a sibling directory](#scenario-yank-then-paste-copies-a-spaced-filename-into-a-sibling-directory)
  - [cut then paste moves a spaced filename into a sibling directory](#scenario-cut-then-paste-moves-a-spaced-filename-into-a-sibling-directory)
  - [yank then paste copies a spaced directory subtree into a sibling directory](#scenario-yank-then-paste-copies-a-spaced-directory-subtree-into-a-sibling-directory)
  - [cut then paste moves a spaced directory subtree into a sibling directory](#scenario-cut-then-paste-moves-a-spaced-directory-subtree-into-a-sibling-directory)
  - [dash symlinks a spaced filename with an absolute target](#scenario-dash-symlinks-a-spaced-filename-with-an-absolute-target)
  - [underscore symlinks a spaced filename with a relative target](#scenario-underscore-symlinks-a-spaced-filename-with-a-relative-target)
  - [dash symlinks a spaced directory with an absolute target](#scenario-dash-symlinks-a-spaced-directory-with-an-absolute-target)
  - [underscore symlinks a spaced directory with a relative target](#scenario-underscore-symlinks-a-spaced-directory-with-a-relative-target)
  - [hidden toggle can be turned back off](#scenario-hidden-toggle-can-be-turned-back-off)
  - [block shell command creates a directory and returns to the browser](#scenario-block-shell-command-creates-a-directory-and-returns-to-the-browser)
  - [block shell command can rename a file in place](#scenario-block-shell-command-can-rename-a-file-in-place)
  - [block shell command creates a spaced filename](#scenario-block-shell-command-creates-a-spaced-filename)
  - [block shell command creates a spaced directory](#scenario-block-shell-command-creates-a-spaced-directory)
  - [block shell command deletes the hovered file](#scenario-block-shell-command-deletes-the-hovered-file)
  - [block shell command copies the hovered file to a new name](#scenario-block-shell-command-copies-the-hovered-file-to-a-new-name)
  - [block shell command runs inside the entered directory](#scenario-block-shell-command-runs-inside-the-entered-directory)
  - [block shell command can create nested files inside a new directory](#scenario-block-shell-command-can-create-nested-files-inside-a-new-directory)
  - [block shell command moves a file into a new directory](#scenario-block-shell-command-moves-a-file-into-a-new-directory)
  - [block shell command renames a directory](#scenario-block-shell-command-renames-a-directory)
  - [block shell command appends to the hovered file content](#scenario-block-shell-command-appends-to-the-hovered-file-content)
  - [block shell command can create a hidden file that dot then reveals](#scenario-block-shell-command-can-create-a-hidden-file-that-dot-then-reveals)
  - [block shell command deletes a directory subtree](#scenario-block-shell-command-deletes-a-directory-subtree)
  - [block shell command copies a directory subtree to a new name](#scenario-block-shell-command-copies-a-directory-subtree-to-a-new-name)
  - [block shell command creates a symlink to a file](#scenario-block-shell-command-creates-a-symlink-to-a-file)
  - [block shell command creates a symlink to a directory](#scenario-block-shell-command-creates-a-symlink-to-a-directory)
  - [block shell command renames a spaced filename](#scenario-block-shell-command-renames-a-spaced-filename)
  - [block shell command creates a symlink to a spaced filename](#scenario-block-shell-command-creates-a-symlink-to-a-spaced-filename)
  - [block shell command creates a symlink to a spaced directory](#scenario-block-shell-command-creates-a-symlink-to-a-spaced-directory)
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
### Scenario: canceling permanent delete returns to the file list and keeps the file
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
- file `keep.txt` contains `keep`
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
### Scenario: canceling trash returns to the file list and keeps the file
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
- file `keep.txt` contains `keep`
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
### Scenario: natural sort is visible on the rendered file list before quitting
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
### Scenario: reverse natural sort is visible on the rendered file list before quitting
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
### Scenario: create accepts filenames with spaces
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
- file `two words.txt` exists
#### Generated artifacts
- `two words.txt`
### Scenario: create accepts directory names with spaces
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
- dir `two words` exists
### Scenario: rename accepts spaces in filenames
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
- file `old two words.txt` contains `old`
### Scenario: rename accepts spaces in directory names
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `old-dir/keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `old-dir/keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `old-dir two words/keep.txt` contains `keep`
### Scenario: yank then dash creates an absolute symlink in the sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
readlink dst/a.txt
```
#### Then
- stdout matches `/^/tmp/atago-[0-9]+/src/a\.txt
?$/`
### Scenario: yank then underscore creates a relative symlink in the sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
readlink dst/a.txt
```
#### Then
- stdout equals an exact value
### Scenario: yanked directory can be absolute-symlinked into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/sub/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/sub/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
readlink dst/sub
```
#### Then
- stdout matches `/^/tmp/atago-[0-9]+/src/sub
?$/`
### Scenario: yanked directory can be relative-symlinked into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/sub/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/sub/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
readlink dst/sub
```
#### Then
- stdout equals an exact value
### Scenario: modified time sort reorders the rendered list
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `oldest.txt` is created.
- Fixture file `newest.txt` is created.
- Fixture file `middle.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `oldest.txt`:_
```text
old
```
_Fixture `newest.txt`:_
```text
new
```
_Fixture `middle.txt`:_
```text
mid
```
#### When
```shell
# interactive (pty): yazi .
```
### Scenario: reverse modified time sort reorders the rendered list
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `oldest.txt` is created.
- Fixture file `newest.txt` is created.
- Fixture file `middle.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `oldest.txt`:_
```text
old
```
_Fixture `newest.txt`:_
```text
new
```
_Fixture `middle.txt`:_
```text
mid
```
#### When
```shell
# interactive (pty): yazi .
```
### Scenario: modified time sort chooses the oldest file first
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `oldest.txt` is created.
- Fixture file `newest.txt` is created.
- Fixture file `middle.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `oldest.txt`:_
```text
old
```
_Fixture `newest.txt`:_
```text
new
```
_Fixture `middle.txt`:_
```text
mid
```
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- file `chosen.txt` contains `oldest.txt`
### Scenario: reverse modified time sort chooses the newest file first
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `oldest.txt` is created.
- Fixture file `newest.txt` is created.
- Fixture file `middle.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `oldest.txt`:_
```text
old
```
_Fixture `newest.txt`:_
```text
new
```
_Fixture `middle.txt`:_
```text
mid
```
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- file `chosen.txt` contains `newest.txt`
### Scenario: toggling the same file twice removes the explicit selection
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
- file `chosen.txt` contains `/b.txt`
### Scenario: inverse selection flips one picked file into the other two
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- Fixture file `c.txt` is created.
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
_Fixture `c.txt`:_
```text
c
```
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
cat chosen.txt
```
#### Then
- stdout matches `/(?m)^.*/b\.txt\n.*/c\.txt\n?$/`
### Scenario: direct tab switching with 1 and 2 preserves per-tab cwd
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `one/a.txt` is created.
- Fixture file `two/b.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `one/a.txt`:_
```text
a
```
_Fixture `two/b.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi --cwd-file ${workdir}/cwd.txt .
```
#### Then
- file `cwd.txt` contains `/two`
### Scenario: block shell command creates a file and returns to the browser
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
- file `shell-created.txt` contains `shell-created`
### Scenario: filter can be cleared after confirming it
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `alpha.txt` is created.
- Fixture file `beta.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `alpha.txt`:_
```text
a
```
_Fixture `beta.txt`:_
```text
b
```
#### When
```shell
# interactive (pty): yazi .
```
### Scenario: yank then paste copies a directory subtree into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/sub/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/sub/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `dst/src/sub/a.txt` contains `a`
- file `src/sub/a.txt` contains `a`
### Scenario: cut then paste moves a directory subtree into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/sub/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/sub/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `dst/src/sub/a.txt` contains `a`
- dir `src` does not exist
### Scenario: lowercase p copies a directory to name_1 when the destination exists
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/sub/a.txt` is created.
- Fixture file `dst/src/existing.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/sub/a.txt`:_
```text
src
```
_Fixture `dst/src/existing.txt`:_
```text
dst
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `dst/src/existing.txt` contains `dst`
- file `dst/src_1/sub/a.txt` contains `src`
### Scenario: lowercase p moves a directory to name_1 when the destination exists
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/sub/a.txt` is created.
- Fixture file `dst/src/existing.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/sub/a.txt`:_
```text
src
```
_Fixture `dst/src/existing.txt`:_
```text
dst
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `dst/src/existing.txt` contains `dst`
- file `dst/src_1/sub/a.txt` contains `src`
- dir `src` does not exist
### Scenario: uppercase D permanently deletes a directory subtree after confirmation
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed/sub/a.txt` is created.
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed/sub/a.txt`:_
```text
a
```
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- dir `doomed` does not exist
### Scenario: canceling uppercase D keeps a directory subtree intact
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed/sub/a.txt` is created.
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed/sub/a.txt`:_
```text
a
```
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `doomed/sub/a.txt` contains `a`
### Scenario: trash moves a directory subtree into sandbox home trash
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed/sub/a.txt` is created.
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed/sub/a.txt`:_
```text
a
```
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- dir `doomed` does not exist
- file `.atago-home/.local/share/Trash/files/doomed/sub/a.txt` contains `a`
### Scenario: canceling trash keeps a directory subtree intact
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed/sub/a.txt` is created.
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed/sub/a.txt`:_
```text
a
```
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `doomed/sub/a.txt` contains `a`
### Scenario: canceling trash for selected directories keeps both subtrees
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a/sub/x.txt` is created.
- Fixture file `b/sub/y.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `a/sub/x.txt`:_
```text
x
```
_Fixture `b/sub/y.txt`:_
```text
y
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `a/sub/x.txt` contains `x`
- file `b/sub/y.txt` contains `y`
### Scenario: yank then paste copies a spaced filename into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- file `dst/two words.txt` contains `spaced`
### Scenario: cut then paste moves a spaced filename into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- file `dst/two words.txt` contains `spaced`
- file `src/two words.txt` does not exist
### Scenario: yank then paste copies a spaced directory subtree into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words/a.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- file `dst/two words/a.txt` contains `spaced`
### Scenario: cut then paste moves a spaced directory subtree into a sibling directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words/a.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
```
#### Then
- file `dst/two words/a.txt` contains `spaced`
- dir `src/two words` does not exist
### Scenario: dash symlinks a spaced filename with an absolute target
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
readlink "dst/two words.txt"
```
#### Then
- stdout matches `/^/tmp/atago-[0-9]+/src/two words\.txt
?$/`
### Scenario: underscore symlinks a spaced filename with a relative target
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
readlink "dst/two words.txt"
```
#### Then
- stdout equals an exact value
### Scenario: dash symlinks a spaced directory with an absolute target
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words/a.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
readlink "dst/two words"
```
#### Then
- stdout matches `/^/tmp/atago-[0-9]+/src/two words
?$/`
### Scenario: underscore symlinks a spaced directory with a relative target
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/two words/a.txt` is created.
- Fixture file `dst/.keep` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/two words/a.txt`:_
```text
spaced
```
#### When
```shell
# interactive (pty): yazi src
readlink "dst/two words"
```
#### Then
- stdout equals an exact value
### Scenario: hidden toggle can be turned back off
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `alpha.txt` is created.
- Fixture file `.secret` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): yazi --chooser-file ${workdir}/chosen.txt .
```
#### Then
- file `chosen.txt` contains `alpha.txt`
### Scenario: block shell command creates a directory and returns to the browser
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
- dir `shell-dir` exists
### Scenario: block shell command can rename a file in place
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
- file `shell-renamed.txt` contains `old`
### Scenario: block shell command creates a spaced filename
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
- file `two words.txt` contains `spaced`
### Scenario: block shell command creates a spaced directory
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
- dir `two words` exists
### Scenario: block shell command deletes the hovered file
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed.txt` is created.
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed.txt`:_
```text
doomed
```
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `doomed.txt` does not exist
### Scenario: block shell command copies the hovered file to a new name
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
- file `shell-copy.txt` contains `old`
### Scenario: block shell command runs inside the entered directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `sub/keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `sub/keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `sub/inside.txt` contains `inside`
### Scenario: block shell command can create nested files inside a new directory
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
- file `shell-nested/file.txt` contains `nested`
### Scenario: block shell command moves a file into a new directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `move-me.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `move-me.txt`:_
```text
move
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `moved/move-me.txt` contains `move`
### Scenario: block shell command renames a directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `old-dir/a.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `old-dir/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `shell-dir/a.txt` contains `a`
### Scenario: block shell command appends to the hovered file content
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `note.txt`:_
```text
first
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `note.txt` contains `second`
### Scenario: block shell command can create a hidden file that dot then reveals
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
- file `.shell-hidden` contains `hidden`
### Scenario: block shell command deletes a directory subtree
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `doomed/a.txt` is created.
- Fixture file `keep.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doomed/a.txt`:_
```text
doomed
```
_Fixture `keep.txt`:_
```text
keep
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- dir `doomed` does not exist
### Scenario: block shell command copies a directory subtree to a new name
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `shell-copy-dir/a.txt` contains `a`
### Scenario: block shell command creates a symlink to a file
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `target.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `target.txt`:_
```text
target
```
#### When
```shell
# interactive (pty): yazi .
readlink linked.txt
```
#### Then
- stdout equals an exact value
### Scenario: block shell command creates a symlink to a directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `target-dir/a.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `target-dir/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
readlink linked-dir
```
#### Then
- stdout equals an exact value
### Scenario: block shell command renames a spaced filename
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `old name.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `old name.txt`:_
```text
old
```
#### When
```shell
# interactive (pty): yazi .
```
#### Then
- file `new name.txt` contains `old`
### Scenario: block shell command creates a symlink to a spaced filename
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `target file.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `target file.txt`:_
```text
target
```
#### When
```shell
# interactive (pty): yazi .
readlink "linked file.txt"
```
#### Then
- stdout equals an exact value
### Scenario: block shell command creates a symlink to a spaced directory
_only when `yazi --version` succeeds · skipped on Windows_
#### Given
- Fixture file `target dir/a.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `target dir/a.txt`:_
```text
a
```
#### When
```shell
# interactive (pty): yazi .
readlink "linked dir"
```
#### Then
- stdout equals an exact value
