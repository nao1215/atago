# atago Behavior Specs
## Summary
1 suite · 15 scenarios
## Contents
- [gum (third-party shell-script TUI toolkit)](#gum-third-party-shell-script-tui-toolkit) — 15 scenarios
  - [version prints the pinned semantic version](#scenario-version-prints-the-pinned-semantic-version)
  - [version-check succeeds silently for a satisfied constraint](#scenario-version-check-succeeds-silently-for-a-satisfied-constraint)
  - [version-check rejects an unsatisfied constraint](#scenario-version-check-rejects-an-unsatisfied-constraint)
  - [join horizontal concatenates adjacent arguments](#scenario-join-horizontal-concatenates-adjacent-arguments)
  - [table print renders a static border and rows from stdin](#scenario-table-print-renders-a-static-border-and-rows-from-stdin)
  - [log writes a leveled message to a file](#scenario-log-writes-a-leveled-message-to-a-file)
  - [choose select-if-one returns the only matching stdin option without a tty](#scenario-choose-select-if-one-returns-the-only-matching-stdin-option-without-a-tty)
  - [filter select-if-one returns the only fuzzy match without a tty](#scenario-filter-select-if-one-returns-the-only-fuzzy-match-without-a-tty)
  - [interactive choose moves the cursor and writes the chosen option](#scenario-interactive-choose-moves-the-cursor-and-writes-the-chosen-option)
  - [interactive filter narrows to the typed query and submits the match](#scenario-interactive-filter-narrows-to-the-typed-query-and-submits-the-match)
  - [confirm exits 0 on affirmative input](#scenario-confirm-exits-0-on-affirmative-input)
  - [confirm exits 1 on negative input](#scenario-confirm-exits-1-on-negative-input)
  - [input starts from a seeded value and writes the edited result](#scenario-input-starts-from-a-seeded-value-and-writes-the-edited-result)
  - [file picker returns the selected file path](#scenario-file-picker-returns-the-selected-file-path)
  - [pager shows line numbers and quits on q](#scenario-pager-shows-line-numbers-and-quits-on-q)
## gum (third-party shell-script TUI toolkit)
[gum](https://github.com/charmbracelet/gum) exists to put a real interface
on a shell script, and it has two personalities to match.

Some subcommands are ordinary filters — `join`, `table`, `log`,
`version-check` — with output a script parses; those are pinned as plain
stdout contracts.

The rest are prompts that only exist in front of a person: `choose`,
`filter`, `confirm`, `input`, `file`, `pager`. Those are driven inside a
pseudo-terminal with real keystrokes, because they are the reason anyone
installs gum, and a test that skipped them would cover everything except the
product.

Source: `test/e2e/thirdparty/gum/gum.atago.yaml`
### Scenario: version prints the pinned semantic version
_only when `gum --version` succeeds_
#### When
```shell
gum --version
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: version-check succeeds silently for a satisfied constraint
_only when `gum --version` succeeds_
#### When
```shell
gum version-check ">=0.17.0"
```
#### Then
- exit code is `0`
- stdout is empty
- stderr is empty
### Scenario: version-check rejects an unsatisfied constraint
_only when `gum --version` succeeds_
#### When
```shell
gum version-check ">0.17.0"
```
#### Then
- exit code is `1`
- stderr contains `v0.17.0`, `>0.17.0`
### Scenario: join horizontal concatenates adjacent arguments
_only when `gum --version` succeeds_
#### When
```shell
gum join --horizontal alpha beta gamma
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: table print renders a static border and rows from stdin
_only when `gum --version` succeeds_
#### When
```shell
cat <<'EOF' | gum table --print -c Name,Lang
alpha,Go
beta,Rust
EOF

```
#### Then
- exit code is `0`
- stdout contains `Name`, `Lang`, `alpha`, `beta`
### Scenario: log writes a leveled message to a file
_only when `gum --version` succeeds_
#### When
```shell
gum log --file app.log --level info --prefix app hello world
```
#### Then
- exit code is `0`
- file `app.log` contains `INFO app: hello world`
### Scenario: choose select-if-one returns the only matching stdin option without a tty
_only when `gum --version` succeeds_
#### When
```shell
printf 'beta\n' | gum choose --select-if-one
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: filter select-if-one returns the only fuzzy match without a tty
_only when `gum --version` succeeds_
#### When
```shell
printf 'alpha\nbeta\ngamma\n' | gum filter --value=ga --select-if-one
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: interactive choose moves the cursor and writes the chosen option
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): gum choose alpha beta gamma > pick.txt
```
#### Then
- exit code is `0`
- file `pick.txt` contains `beta`
### Scenario: interactive filter narrows to the typed query and submits the match
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): gum filter alpha beta gamma > pick.txt
```
#### Then
- exit code is `0`
- file `pick.txt` contains `gamma`
### Scenario: confirm exits 0 on affirmative input
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): gum confirm "Deploy?"
```
#### Then
- exit code is `0`
### Scenario: confirm exits 1 on negative input
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): gum confirm "Deploy?"
```
#### Then
- exit code is `1`
### Scenario: input starts from a seeded value and writes the edited result
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): gum input --value seed > out.txt
```
#### Then
- exit code is `0`
- file `out.txt` contains `seed-tail`
### Scenario: file picker returns the selected file path
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- Fixture file `a.txt` is created.
- Fixture file `b.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
# interactive (pty): gum file --file . > pick.txt
```
#### Then
- exit code is `0`
- file `pick.txt` contains `/a.txt`
### Scenario: pager shows line numbers and quits on q
_only when `gum --version` succeeds · skipped on Windows_
#### Given
- Fixture file `sample.txt` is created.
- The command runs with a cleared environment (passing through: PATH).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `sample.txt`:_
```text
first
second
third
```
#### When
```shell
# interactive (pty): cat sample.txt | gum pager --show-line-numbers
```
#### Then
- exit code is `0`
