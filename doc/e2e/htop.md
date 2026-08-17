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
[htop](https://htop.dev/) is a full-screen program in the strict sense: it
switches the terminal to the alternate screen buffer, paints meters and a
process table with cursor movement and color, and restores what was there
before when it quits.

That is a harder thing to assert than scrolling output, and it is what this
suite pins. The rendered frame is reconstructed the way a terminal would
reconstruct it — following the alternate-screen switch and the cursor
movements — so the assertion is about what a user sees on screen, not about
the raw escape sequences that produced it. Quitting cleanly, and leaving the
terminal in a usable state, is part of the contract too.

Source: `test/e2e/thirdparty/htop/htop.atago.yaml`
### Scenario: version prints a semantic version
_only when `htop --version` succeeds_
#### When
```shell
htop --version
```
#### Then
- exit code is `0`
- stdout matches `/^htop [0-9]+\.[0-9]+/`

### Scenario: an unrecognized option is rejected without opening the TUI
_only when `htop --version` succeeds_
#### When
```shell
htop --nonexistent-flag
```
#### Then
- exit code is not `0`
- stderr contains `unrecognized option`

### Scenario: the finder loads its function-key bar and quits on q
_only when `htop --version` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): htop
```
#### Then
- exit code is `0`

### Scenario: the rendered screen shows the live meters and column header
_only when `htop --version` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): htop
```
#### Then
- rendered screen contains `CPU%`
- rendered screen contains `Command`
- rendered screen contains `F10`
