---
title: atago
---

atago tests real CLI behavior from plain YAML: commands, files, snapshots, services, and interactive terminals. It runs your actual binary — in any language — and asserts what a user observes. No test code, no shell DSL.

![atago running a spec suite](/img/demo.gif)

```shell
atago record --out mytool.atago.yaml -- mytool convert input.txt  # turn a real run into a spec
atago run mytool.atago.yaml                                       # replay it as a test
atago run --report junit specs/                                   # or run a whole suite in CI
```

## Why atago?

Pick the tool that owns your layer:

| You are testing | Use |
|-----------------|-----|
| An HTTP/gRPC API server — scenario-based API testing | [runn](https://github.com/k1LoW/runn) |
| A whole platform — integration suites across HTTP, gRPC, Kafka, databases, and more | [venom](https://github.com/ovh/venom) |
| Shell functions and scripts — BDD-style unit tests | [ShellSpec](https://shellspec.info/) |
| Bash scripts — TAP-style tests | [Bats](https://github.com/bats-core/bats-core) |
| A CLI product — exit codes, output, generated files, snapshots, interactive prompts and TUIs | atago |

If the server or the platform is the system under test, use runn or venom. atago points the other way: the CLI is the product, and HTTP, database, SSH, gRPC, browser, and mock servers appear only as peers your CLI talks to.

## Record, don't write

The first spec comes from a real run: `atago record -- <command>` executes the tool once and writes a spec from what it observed — exit code, output, created files. Interactive tools record too: `record --pty` runs in a real terminal, you drive one session by hand, and the keystrokes become a replayable expect/send script, with password prompts masked into `${env:...}` placeholders automatically. Then you edit YAML, not write it from scratch.

## Snapshots built for CLI output

![snapshot testing: failure diff and one-command update](/img/snapshot.gif)

`snapshot:` pins output to a committed golden file. ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF are normalized so goldens stay stable across machines; your own volatile patterns get spec-wide `scrub:` rules. A failure renders a colorized unified diff, and `atago snapshot update` re-records everything intentionally changed — the review is just `git diff snapshots/`.

## Real terminals, not faked pipes

A `pty:` step runs the command in a real pseudo-terminal — on Windows too (ConPTY) — and drives it with declarative expect/send pairs and named keys, no `expect(1)` scripting. `screen:` asserts the RENDERED frame a user actually sees, after cursor movement and clears are applied, and screen snapshots pin full TUI layouts. TTY-detection branches, wizards, REPLs, htop-style dashboards: all of it is spec-able.

## Where to go

- [Install](/install/) — `go install`, Homebrew, AUR, prebuilt binaries, a GitHub Action.
- [Getting started](/getting-started/) — record a real run, read the spec it wrote, keep it as a test.
- [Cookbook](/cookbook/) — 50+ copyable recipes, from image conversion to API-failure simulation to TUI frames.
- [Examples](/examples/) — every feature has a commented, runnable spec, tested in CI on Linux, macOS, and Windows.
- [Use it in CI](/ci/) — report formats, retries, flake detection, artifacts, secret masking.
- [Reference](/reference/) — subcommands, selection flags, exit codes, editor schema.
- [Write specs with an LLM](/llm/) — a ready-made prompt that keeps generated specs honest.
- [Real CLIs tested with atago](/real-world/) — 40+ programs from jq to terraform to htop, each with executable specs and generated behavior docs.

Everything on this site comes from files committed in the [repository](https://github.com/nao1215/atago). The behavior docs are regenerated from executable specs, and drift tests fail the build when docs and specs disagree — if an example is on this site, it runs.
