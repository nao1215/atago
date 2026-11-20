<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

[![Build](https://github.com/nao1215/atago/actions/workflows/build.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/atago/actions/workflows/unit_test.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/unit_test.yml)
![Coverage](https://raw.githubusercontent.com/nao1215/octocovs-central-repo/main/badges/nao1215/atago/coverage.svg)
[![Gitleaks](https://github.com/nao1215/atago/actions/workflows/gitleaks.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/gitleaks.yml)
[![Trivy](https://github.com/nao1215/atago/actions/workflows/trivy.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/trivy.yml)
[![govulncheck](https://github.com/nao1215/atago/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/govulncheck.yml)
[![Gosec](https://github.com/nao1215/atago/actions/workflows/gosec.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/gosec.yml)
[![reviewdog](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml)
[![tested with atago](https://img.shields.io/badge/tested%20with-atago-7c3aed?logo=data:image/svg%2Bxml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCI%2BPHBhdGggZmlsbD0iI2ZmZiIgZD0iTTMuNiA0LjIgMTEuOSAxMmwtOC4zIDcuOC0xLjktMi4yTDcuOSAxMiAxLjcgNi40eiIvPjxyZWN0IGZpbGw9IiNmZmYiIHg9IjEyLjYiIHk9IjE3LjIiIHdpZHRoPSI5LjciIGhlaWdodD0iMi44IiByeD0iMS40Ii8%2BPC9zdmc%2B&logoColor=white)](https://github.com/nao1215/atago)
[![measured with himorime](https://img.shields.io/badge/measured%20with-himorime-d9480f?logo=data:image/svg%2Bxml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCI%2BPHBhdGggZmlsbD0ibm9uZSIgc3Ryb2tlPSIjZmZmIiBzdHJva2Utd2lkdGg9IjIuNCIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIiBkPSJNNC4yIDE4LjVBOSA5IDAgMSAxIDE5LjggMTguNSIvPjxwYXRoIGZpbGw9Im5vbmUiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIyLjQiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgZD0iTTEyIDE0LjUgMTYuNSA5Ii8%2BPGNpcmNsZSBmaWxsPSIjZmZmIiBjeD0iMTIiIGN5PSIxNC41IiByPSIyLjIiLz48L3N2Zz4=&logoColor=white)](https://github.com/nao1215/himorime)
[![Go Reference](https://pkg.go.dev/badge/github.com/nao1215/atago.svg)](https://pkg.go.dev/github.com/nao1215/atago)
[![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/nao1215/atago/total)](https://github.com/nao1215/atago/releases)
[![License](https://img.shields.io/github/license/nao1215/atago)](./LICENSE)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/nao1215/atago/badge)](https://scorecard.dev/viewer/?uri=github.com/nao1215/atago)

<p align="center">
  <img src="./doc/img/atago-logo.jpg" alt="atago logo" width="400" />
</p>

atago tests real CLI behavior from plain YAML: commands, files, snapshots, services, and interactive terminals. It runs your actual binary, written in any language, and asserts what a user observes. You write no test code and no shell DSL.

Documentation: https://nao1215.github.io/atago/

![demo](./doc/img/demo.gif)

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

`record` runs `git --version` once and writes a spec from what it observed: the exit code, the version line on stdout, and an empty stderr. `run` replays it. Open `demo.atago.yaml` to tighten the checks.

Then point it at your own tool:

```shell
atago record --out mytool.atago.yaml -- mytool convert input.txt  # turn a real run into a spec
atago run mytool.atago.yaml                                       # replay it as a test
atago run --report junit specs/                                   # or run a whole suite in CI
```

`atago record --pty -- <command>` does the same for an interactive tool: you drive one session by hand in a real terminal, and atago writes a `pty:` step that replays it. `atago init` scaffolds a blank spec instead.

## Why atago?

| You are testing | Use |
|-----------------|-----|
| An HTTP/gRPC API server (scenario-based API testing) | [runn](https://github.com/k1LoW/runn) |
| A whole platform (integration suites across HTTP, gRPC, Kafka, databases, and more) | [venom](https://github.com/ovh/venom) |
| Shell functions and scripts (BDD-style unit tests) | [ShellSpec](https://shellspec.info/) |
| Bash scripts (TAP-style tests) | [Bats](https://github.com/bats-core/bats-core) |
| A CLI product: exit codes, output, generated files, snapshots, interactive prompts and TUIs | atago |

The [comparison](https://nao1215.github.io/atago/comparison/) goes feature by feature, and the [migration guide](https://nao1215.github.io/atago/migrate/) maps Bats and ShellSpec constructs to atago one to one.

## Install

```shell
go install github.com/nao1215/atago@latest
```

```shell
brew install --cask nao1215/tap/atago
```

```shell
yay -S atago-bin   # or: paru -S atago-bin
```

```shell
aqua g -i nao1215/atago
aqua i
```

```shell
mise use -g aqua:nao1215/atago
```

The [install page](https://nao1215.github.io/atago/install/) covers each of these, release verification, and the prebuilt archives and `.deb`/`.rpm`/`.apk` packages on the [release page](https://github.com/nao1215/atago/releases). Building from source requires Go 1.26.11 or later. CI runs atago on Linux, macOS, and Windows.

## Writing specs

### Exit code, stdout, and stderr

```yaml
version: "1"
suite:
  name: example
scenarios:
  - name: echo greets the world
    steps:
      - run:
          shell: true            # portable: echo is a shell builtin on Windows
          command: echo "hello atago"
      - assert:
          exit_code: 0
          stdout:
            contains: atago
          stderr:
            empty: true
```

`atago run` takes spec files and directories, searched recursively for `*.atago.yaml`. Each scenario runs concurrently in its own temporary directory:

```shell
$ atago run ./specs
............

PASSED  12 scenarios: 12 passed, 0 failed, 0 errored, 0 skipped (1.4s)
```

When a check fails, atago prints what was expected and what happened:

```text
FAILED: demo / greeting matches its golden  (demo.atago.yaml)

Step:
  assert stdout snapshot

Command:
  mytool greet

Diff (-expected +actual):
  --- snapshot (golden)
  +++ actual
  @@ -1,3 +1,3 @@
   hello
  -WORLD
  +world
   bye

Hint:
  stdout did not match snapshot "snaps/greeting.txt" (update with --update-snapshots if intended)

FAILED  1 scenario: 0 passed, 1 failed, 0 errored, 0 skipped (3ms)
```

### Generated files and snapshots

`fixture:` writes input files into the workdir from `content:` (inline text), `base64:`, `from:` (an existing file), or `symlink:`. `file:` and `dir:` assertions check what the command produced, and `snapshot:` pins output to a committed golden file:

```yaml
scenarios:
  - name: the generator writes the expected files
    steps:
      - run:
          command: mytool generate --out site
      - assert:
          file:
            path: site/index.html
            contains:
              - "<html"
      - assert:
          stdout:
            snapshot: snapshots/generate.txt   # record/refresh with `atago snapshot update`
```

![snapshot](./doc/img/snapshot.gif)

Snapshots normalize ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF, and spec-wide `scrub:` rules handle your own volatile values. `atago snapshot update` re-records them.

Other checks have a runnable example each: [count_and_size](examples/count_and_size.atago.yaml) (how often a match occurs, file sizes), [deterministic](examples/deterministic.atago.yaml) (byte-identical output across runs), [expect_fail](examples/expect_fail.atago.yaml) (a known bug kept in CI as XFAIL), [project_manifest](examples/project_manifest.atago.yaml) (shared settings and building the binary under test), and [forall](examples/forall.atago.yaml) (generated inputs, seeded from the spec so every machine tests the same ones).

### Interactive prompts and TUIs

A `pty` step runs the command in a real pseudo-terminal (a ConPTY on Windows) and drives it with expect/send pairs:

```yaml
scenarios:
  - name: the init wizard completes
    steps:
      - pty:
          command: mytool init
          session:
            - expect: "Project name:"
            - send: "demo\n"
            - expect: "created demo/"
      - assert:
          exit_code: 0
```

For full-screen TUIs, `expect_screen:` waits on the rendered frame while the program runs, and `screen:` asserts the final frame after exit:

```yaml
      - pty:
          command: mytool dashboard
          session:
            - expect_screen:
                contains: "Ready"
                stable_for: 100ms
            - send: "q"
      - assert:
          screen:
            contains: "Summary"
```

Named keys, control-byte aliases, and the Windows notes are in [getting started](https://nao1215.github.io/atago/getting-started/) and the [pty](examples/pty.atago.yaml) and [pty_portable](examples/pty_portable.atago.yaml) examples.

### When your CLI talks to a server

The same YAML drives HTTP, database, SSH, gRPC, headless browser, and offline mock server peers as dependencies of the CLI under test:

```shell
$ atago init --list-templates
browser   drive a headless Chrome; emulate viewports and assert page content (needs Chrome on PATH)
cli       run a command; assert exit code/stdout/stderr (runs as-is)
db        run SQL; assert on rows via bundled SQLite (runs as-is)
grpc      call a unary gRPC method via server reflection (edit target first)
http      call an HTTP API; assert status and JSON body (edit base_url first)
mock      stub an HTTP API offline and assert what the client sent (needs curl on PATH)
services  test against a background server: readiness, retry, teardown (runs as-is)
ssh       run a command on a remote host over SSH (edit host/user first)

Scaffold one with: atago init --template <name>
```

Every feature has a commented, runnable spec under [examples/](examples/), tested in CI on Linux, macOS, and Windows, and the [cookbook](https://nao1215.github.io/atago/cookbook/) has 50+ recipes indexed by task.

## Use it in CI

```yaml
name: behavior-specs
on: [push, pull_request]
jobs:
  atago:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: nao1215/setup-atago@v0
      - run: atago run --ci --report gha ./specs
```

[setup-atago](https://github.com/nao1215/setup-atago) installs a released binary; other CI systems can use the `ghcr.io/nao1215/atago` image. `--report` writes `json`, `junit`, `gha`, or `tap`. `--retry-failed N` reports a scenario that passed on retry as flaky and still fails the run unless `--allow-flaky` is passed, and `--repeat N` hunts for flakiness:

```shell
atago run --ci --retry-failed 2 ./specs          # retry, report instability loudly, still fail
atago run --ci --retry-failed 2 --allow-flaky ./specs  # ...and keep CI green when that is accepted
atago run --repeat 20 --filter "race prone" ./specs   # flake detection
```

`--filter`, `--tag`, `--skip-tag`, `--fail-fast`, and `--rerun-failed` select what runs; with `--ci`, a selection that matches nothing fails the run. The [CI page](https://nao1215.github.io/atago/ci/) covers these, GitLab, `--artifacts-dir`, and secret masking.

## Review specs without running them

![review](./doc/img/review.gif)

```shell
atago explain spec.atago.yaml
atago doc --out docs/specs.md ./specs
atago manifest ./specs
atago list ./specs
```

Each validates the spec first and exits 2 on a schema error, so any of them works as a lint step. Spec errors carry a code such as `ATG2201`, explained in the [error reference](https://nao1215.github.io/atago/errors/).

## More

- [schema/atago.schema.json](schema/atago.schema.json) gives YAML editors completion and validation. `atago init` and `atago record` add the header to the specs they write.
- An [Agent Skill](skills/atago/SKILL.md) for Claude Code and Codex is in `skills/atago`.
- `atago completion <bash|zsh|fish|powershell>` prints a completion script.
- The [reference](https://nao1215.github.io/atago/reference/) lists every subcommand, flag, and spec key.

## Exit codes

| Code | Meaning                   |
|------|---------------------------|
| `0`  | all scenarios passed      |
| `1`  | one or more failed        |
| `2`  | spec error (YAML syntax or schema/semantic validation) |
| `3`  | CLI-invocation error (unknown subcommand, bad flag, or no matching spec files) |
| `4`  | execution error           |
| `5`  | internal error            |
| `6`  | security policy violation |

## Real CLIs tested with atago

atago tests itself, the author's Go tools, and unmodified third-party binaries from git and jq to fzf, htop, yazi, redis, terraform, openssl, and ffmpeg. The [real-world index](https://nao1215.github.io/atago/real-world/) has their specs and generated behavior docs.

## The name

atago (愛宕) is named for Mount Atago in Kyoto, whose shrine enshrines a deity of fire prevention. A test runner has the same job: catch the sparks before a project catches fire.

## Testing

| Technique | Bug class caught here |
|---|---|
| [Unit tests](doc/testing.md#unit-tests) | Wrong results from one package |
| [Self-hosted E2E](doc/testing.md#self-hosted-e2e) | Behavior a user of the binary sees |
| [Third-party E2E](doc/testing.md#third-party-e2e) | Spec features that break on real programs |
| [Snapshots](doc/testing.md#snapshots) | Unreviewed changes to rendered output |
| [Differential migrations](doc/testing.md#differential-migrations) | Drift from what Bats and ShellSpec do |
| [Generated properties](doc/testing.md#generated-properties) | Report formats that disagree, miscount, or let a name forge structure |
| [Fuzzing](doc/testing.md#fuzzing) | Panics and broken invariants on malformed specs and output |
| [Mutation testing](doc/testing.md#mutation-testing) | Logic no assertion looks at |
| [Benchmarks](doc/testing.md#benchmarks) | Slowdowns between the base and the head |

## Contributing

Issues and pull requests are welcome; see [CONTRIBUTING.md](./CONTRIBUTING.md). A GitHub Star also motivates development.

## LICENSE

The atago project is licensed under the terms of [MIT LICENSE](./LICENSE).

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://debimate.jp/"><img src="https://avatars.githubusercontent.com/u/22737008?v=4?s=75" width="75px;" alt="CHIKAMATSU Naohiro"/><br /><sub><b>CHIKAMATSU Naohiro</b></sub></a><br /><a href="https://github.com/nao1215/atago/commits?author=nao1215" title="Code">💻</a> <a href="https://github.com/nao1215/atago/commits?author=nao1215" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://rafaeldominiquini.ddns.net/"><img src="https://avatars.githubusercontent.com/u/1180808?v=4?s=75" width="75px;" alt="Rafael Baboni Dominiquini"/><br /><sub><b>Rafael Baboni Dominiquini</b></sub></a><br /><a href="#platform-Dominiquini" title="Packaging/porting to new platform">📦</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/team-humaki"><img src="https://avatars.githubusercontent.com/u/327779464?v=4?s=75" width="75px;" alt="team-humaki"/><br /><sub><b>team-humaki</b></sub></a><br /><a href="https://github.com/nao1215/atago/pull/663" title="Code">💻</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
