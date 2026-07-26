# atago Behavior Specs
## Summary
1 suite · 12 scenarios
## Contents
- [helix (third-party modal editor)](#helix-third-party-modal-editor) — 12 scenarios
  - [version prints the pinned release](#scenario-version-prints-the-pinned-release)
  - [opening an existing file reaches the editor and quits cleanly](#scenario-opening-an-existing-file-reaches-the-editor-and-quits-cleanly)
  - [append mode edits an existing file and write quit saves it](#scenario-append-mode-edits-an-existing-file-and-write-quit-saves-it)
  - [force quit discards unsaved changes](#scenario-force-quit-discards-unsaved-changes)
  - [saving a new file path creates the file on disk](#scenario-saving-a-new-file-path-creates-the-file-on-disk)
  - [enter in insert mode writes multiple lines](#scenario-enter-in-insert-mode-writes-multiple-lines)
  - [undo removes an appended edit before save](#scenario-undo-removes-an-appended-edit-before-save)
  - [search selects a match and change replaces only that match](#scenario-search-selects-a-match-and-change-replaces-only-that-match)
  - [x then d deletes the selected line and saves the shorter file](#scenario-x-then-d-deletes-the-selected-line-and-saves-the-shorter-file)
  - [buffer-next switches to the second file in a multi-file session](#scenario-buffer-next-switches-to-the-second-file-in-a-multi-file-session)
  - [unicode filenames and Japanese text save without mangling](#scenario-unicode-filenames-and-japanese-text-save-without-mangling)
  - [wide and combining characters round-trip through a save](#scenario-wide-and-combining-characters-round-trip-through-a-save)
## helix (third-party modal editor)
Source: `test/e2e/thirdparty/helix/helix.atago.yaml`
### Scenario: version prints the pinned release
_only when `hx --version` succeeds_
#### When
```shell
hx --version
```
#### Then
- exit code is `0`
- stdout contains `helix 25.07.1`
### Scenario: opening an existing file reaches the editor and quits cleanly
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `note.txt`:_
```text
alpha
```
#### When
```shell
# interactive (pty): hx note.txt
```
#### Then
- exit code is `0`
### Scenario: append mode edits an existing file and write quit saves it
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `note.txt`:_
```text
alpha
```
#### When
```shell
# interactive (pty): hx note.txt
cat note.txt
```
#### Then
- exit code is `0`
- stdout contains `alpha beta`
### Scenario: force quit discards unsaved changes
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `note.txt`:_
```text
alpha
```
#### When
```shell
# interactive (pty): hx note.txt
cat note.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: saving a new file path creates the file on disk
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
#### When
```shell
# interactive (pty): hx fresh.txt
cat fresh.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: enter in insert mode writes multiple lines
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
#### When
```shell
# interactive (pty): hx lines.txt
cat lines.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```text
first
second
```
### Scenario: undo removes an appended edit before save
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `note.txt`:_
```text
alpha
```
#### When
```shell
# interactive (pty): hx note.txt
cat note.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: search selects a match and change replaces only that match
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `note.txt`:_
```text
alpha beta gamma
```
#### When
```shell
# interactive (pty): hx note.txt
cat note.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: x then d deletes the selected line and saves the shorter file
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `note.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `note.txt`:_
```text
first
second
```
#### When
```shell
# interactive (pty): hx note.txt
cat note.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: buffer-next switches to the second file in a multi-file session
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- Fixture file `one.txt` is created.
- Fixture file `two.txt` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
_Fixture `one.txt`:_
```text
one-start
```
_Fixture `two.txt`:_
```text
two-start
```
#### When
```shell
# interactive (pty): hx one.txt two.txt
cat two.txt
```
#### Then
- exit code is `0`
- stdout contains `two-start updated`
### Scenario: unicode filenames and Japanese text save without mangling
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
#### When
```shell
# interactive (pty): hx 日本語メモ.txt
cat 日本語メモ.txt
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: wide and combining characters round-trip through a save
_only when `hx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `cfg/config.toml` is created.
- The command runs with a cleared environment (passing through: PATH, HELIX_RUNTIME).
#### Inputs
_Fixture `cfg/config.toml`:_
```text
theme = "default"
[editor]
mouse = false
auto-save = false
```
#### When
```shell
# interactive (pty): hx unicode.txt
od -An -tx1 -v unicode.txt
```
#### Then
- exit code is `0`
- stdout contains `65 cc 81`
