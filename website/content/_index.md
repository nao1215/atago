---
title: atago
---

atago tests real CLI behavior from plain YAML: commands, files, snapshots, services, and interactive terminals. It runs your actual binary, written in any language, and asserts what a user observes. You write no test code and no shell DSL.

![atago running a spec suite](/img/demo.gif)

## First run

If you have Go, paste this. It records a real run of a command you already have and replays it as a test:

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped (4ms)
```

`record` runs `git --version` once and writes a spec from what it observed: the exit code, the version line on stdout, and an empty stderr. `run` replays it. Open `demo.atago.yaml` to tighten the checks. Any command works in place of `git --version`, such as `go version`, `jq --version`, or `ls -la`.

Then point it at your own tool:

```shell
atago record --out mytool.atago.yaml -- mytool convert input.txt  # turn a real run into a spec
atago run mytool.atago.yaml                                       # replay it as a test
atago run --report junit specs/                                   # or run a whole suite in CI
```

## Why atago?

| You are testing | Use |
|-----------------|-----|
| An HTTP/gRPC API server (scenario-based API testing) | [runn](https://github.com/k1LoW/runn) |
| A whole platform (integration suites across HTTP, gRPC, Kafka, databases, and more) | [venom](https://github.com/ovh/venom) |
| Shell functions and scripts (BDD-style unit tests) | [ShellSpec](https://shellspec.info/) |
| Bash scripts (TAP-style tests) | [Bats](https://github.com/bats-core/bats-core) |
| A CLI product: exit codes, output, generated files, snapshots, interactive prompts and TUIs | atago |

If the server or the platform is the system under test, use runn or venom. In atago the CLI is the product, and HTTP, database, SSH, gRPC, browser, and mock servers appear only as peers your CLI talks to.

The [feature-by-feature comparison](/comparison/) lists the versions compared and its sources. The [migration guide](/migrate/) shows example mappings from Bats and ShellSpec constructs to atago, and CI runs both the originals and the migrations.

## Record a run

`atago record -- <command>` runs the tool once and writes a spec from its exit code, output, and created files. `record --pty` does the same for an interactive tool: you drive one session by hand in a real terminal, and the keystrokes become a replayable expect/send script. On POSIX, a password typed at an echo-off prompt becomes an `${env:...}` placeholder; keystrokes in a TUI and every send on Windows are recorded literally, so replace secrets there by hand.

## Keep a known bug in CI

`expect_fail:` marks a scenario that documents a bug you have not fixed yet. It runs on every commit, so the reproduction cannot rot the way a CI-excluded directory of "known bug" specs does, and an expected failure keeps the run green. When it starts passing, the run turns red, so you promote the spec into the suite that guards against a regression. An execution error stays an error, so a spec that stops running is not mistaken for a bug that is still there.

## Let atago pick the inputs

`forall:` describes the shape of an input instead of listing values and expands into one scenario per generated row. This is property-based testing in a spec file:

```yaml
forall:
  vars: {text: {type: ascii, min: 0, max: 40}}
```

The generators are seeded from the spec, not the clock, so the same file tests the same inputs on every machine, `atago list` shows them before anything runs, and a failing instance has its input in its name. Boundary values come first (the empty argument, the longest one, both booleans, every `one_of` choice), since that is where a CLI breaks.

## Snapshots for CLI output

![snapshot testing: failure diff and one-command update](/img/snapshot.gif)

`snapshot:` pins output to a committed golden file. ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF are normalized so goldens stay stable across machines; your own volatile patterns get spec-wide `scrub:` rules. A failure renders a colorized unified diff, and `atago snapshot update` re-records what you changed on purpose, so the review is `git diff snapshots/`.

## Real terminals

A `pty:` step runs the command in a real pseudo-terminal (a ConPTY on Windows) and drives it with expect/send pairs and named keys, without `expect(1)` scripts. Named keys include control-byte aliases such as `ctrl-space`, `ctrl-[`, and `ctrl-_`, plus modified key events like `ctrl-hyphen` for TUIs that distinguish the physical `Ctrl+-` key. `expect_screen:` waits on the rendered frame while the program is still running, `screen:` asserts the final frame after cursor movement and clears are applied, and screen snapshots pin full TUI layouts. That covers TTY-detection branches, wizards, REPLs, and htop-style dashboards.

## The name

atago (愛宕) is named for Mount Atago in Kyoto, whose shrine enshrines a deity of fire prevention. Japanese has a word for a software project on fire, 炎上 ("going up in flames"), and the name carries a wish that atago stops those fires before they spread. atago is also written in Go, so the name ends in *go*.

## Where to go

- [Install](/install/): `go install`, Homebrew, AUR, aqua, mise, prebuilt binaries, a GitHub Action.
- [Getting started](/getting-started/): record a real run, read the spec it wrote, keep it as a test.
- [Cookbook](/cookbook/): 50+ copyable recipes indexed by task, plus the runnable, CI-tested example for every feature.
- [Use it in CI](/ci/): report formats, retries, flake detection, artifacts, secret masking.
- [Reference](/reference/): subcommands, exit codes, and every spec key with the release that introduced it.
- [Real CLIs tested with atago](/real-world/): 40+ programs, from jq to terraform to yazi, each with executable specs and generated behavior docs.

Everything on this site comes from files committed in the [repository](https://github.com/nao1215/atago). The behavior docs are regenerated from executable specs, and drift tests fail the build when docs and specs disagree, so every example on this site runs.
