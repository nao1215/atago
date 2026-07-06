# atago Behavior Specs
## Summary
1 suite · 4 scenarios
## Contents
- [htop (third-party CLI, full-screen TUI testbed)](#htop-third-party-cli-full-screen-tui-testbed) — 4 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [an unrecognized option is rejected without opening the TUI](#scenario-an-unrecognized-option-is-rejected-without-opening-the-tui)
  - [the finder loads its function-key bar and quits on q](#scenario-the-finder-loads-its-function-key-bar-and-quits-on-q)
  - [the rendered screen shows the live meters and column header](#scenario-the-rendered-screen-shows-the-live-meters-and-column-header)
## htop (third-party CLI, full-screen TUI testbed)
Source: `test/e2e/thirdparty/htop/htop.atago.yaml`
### Scenario: version prints a semantic version
#### When
```shell
htop --version
```
#### Then
- exit code is `0`
- stdout matches `/^htop [0-9]+\.[0-9]+/`
### Scenario: an unrecognized option is rejected without opening the TUI
#### When
```shell
htop --nonexistent-flag
```
#### Then
- exit code is not `0`
- stderr contains `unrecognized option`
### Scenario: the finder loads its function-key bar and quits on q
_skipped on windows_
#### When
```shell
# interactive (pty): htop
```
#### Then
- exit code is `0`
### Scenario: the rendered screen shows the live meters and column header
_skipped on windows_
#### When
```shell
# interactive (pty): htop
```
#### Then
- rendered screen contains `CPU%`
- rendered screen contains `Command`
- rendered screen contains `F10`