# atago Behavior Specs
## Summary
1 suite · 8 scenarios
## Contents
- [fzf (third-party CLI, pty testbed)](#fzf-third-party-cli-pty-testbed) — 8 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [filter mode matches fuzzily on stdin without a terminal](#scenario-filter-mode-matches-fuzzily-on-stdin-without-a-terminal)
  - [filter mode exits 1 when nothing matches](#scenario-filter-mode-exits-1-when-nothing-matches)
  - [interactive mode refuses to start without a terminal](#scenario-interactive-mode-refuses-to-start-without-a-terminal)
  - [interactive selection picks the queried line](#scenario-interactive-selection-picks-the-queried-line)
  - [multi-select accepts several lines at once](#scenario-multi-select-accepts-several-lines-at-once)
  - [aborting the finder exits 130](#scenario-aborting-the-finder-exits-130)
  - [the finder screen narrows to the typed query](#scenario-the-finder-screen-narrows-to-the-typed-query)
## fzf (third-party CLI, pty testbed)
Source: `test/e2e/thirdparty/fzf/fzf.atago.yaml`
### Scenario: version prints a semantic version
_only when `fzf --version` succeeds_
#### When
```shell
fzf --version
```
#### Then
- exit code is `0`
- stdout matches `/^[0-9]+\.[0-9]+/`
### Scenario: filter mode matches fuzzily on stdin without a terminal
_only when `fzf --version` succeeds_
#### Inputs
_stdin for `fzf`:_
```text
apple
banana
cherry
```
#### When
```shell
fzf --filter=bna
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: filter mode exits 1 when nothing matches
_only when `fzf --version` succeeds_
#### Inputs
_stdin for `fzf`:_
```text
apple
banana
cherry
```
#### When
```shell
fzf --filter=zzz
```
#### Then
- exit code is `1`
- stdout is empty
### Scenario: interactive mode refuses to start without a terminal
_only when env CI is set_
#### When
```shell
printf 'apple\n' | fzf
```
#### Then
- exit code is `2`
- stderr contains `ioctl`
### Scenario: interactive selection picks the queried line
_only when `fzf --version` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): printf 'apple\nbanana\ncherry\n' | fzf > pick.txt
```
#### Then
- exit code is `0`
- file `pick.txt` contains `cherry`
- file `pick.txt` is checked
### Scenario: multi-select accepts several lines at once
_only when `fzf --version` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): printf 'apple\nbanana\ncherry\n' | fzf -m --bind ctrl-a:select-all > pick.txt
```
#### Then
- exit code is `0`
- file `pick.txt` contains `apple`, `banana`, `cherry`
### Scenario: aborting the finder exits 130
_only when `fzf --version` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): printf 'apple\nbanana\ncherry\n' | fzf
```
#### Then
- exit code is `130`
### Scenario: the finder screen narrows to the typed query
_only when `fzf --version` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): printf 'apple\nbanana\ncherry\n' | fzf
```
#### Then
- rendered screen contains `banana`
- rendered screen does not contain `cherry`
