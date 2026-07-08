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

## Where to go

- [Install](/install/) — `go install`, Homebrew, AUR, prebuilt binaries, a GitHub Action.
- [Getting started](/getting-started/) — record a real run, read the spec it wrote, keep it as a test.
- [Cookbook](/cookbook/) — copyable specs for common jobs, from image conversion to graceful shutdown.
- [Examples](/examples/) — every feature has a commented, runnable spec, tested in CI on Linux, macOS, and Windows.
- [Use it in CI](/ci/) — report formats, retries, flake detection, artifacts, secret masking.
- [Reference](/reference/) — subcommands, selection flags, exit codes, editor schema.
- [Real CLIs tested with atago](/real-world/) — 40+ programs from jq to terraform to htop, each with executable specs and generated behavior docs.

Everything on this site comes from files committed in the [repository](https://github.com/nao1215/atago). The behavior docs are regenerated from executable specs, and drift tests fail the build when docs and specs disagree — if an example is on this site, it runs.
