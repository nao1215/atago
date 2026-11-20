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

`record` runs `git --version` once and writes a spec from what it observed: the exit code, the version line on stdout, and an empty stderr (when a tool writes a diagnostic there, the spec anchors its first line instead). `run` replays it. Open `demo.atago.yaml` to tighten the checks. Any command works in place of `git --version`, such as `go version`, `jq --version`, or `ls -la`.

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

The [feature-by-feature comparison](https://nao1215.github.io/atago/comparison/) lists the versions compared and its sources. The [migration guide](https://nao1215.github.io/atago/migrate/) maps Bats and ShellSpec constructs to atago one to one, and CI runs both the originals and the migrations.

## Install

```shell
go install github.com/nao1215/atago@latest
```

On macOS, Homebrew works too:

```shell
brew install --cask nao1215/tap/atago
```

On Arch Linux, install the [`atago-bin`](https://aur.archlinux.org/packages/atago-bin) package from the AUR:

```shell
yay -S atago-bin   # or: paru -S atago-bin
```

atago is in the [aqua](https://aquaproj.github.io/) standard registry. Add it to your `aqua.yaml`, then install it:

```shell
aqua g -i nao1215/atago
aqua i
```

[mise](https://mise.jdx.dev/) can install it through the same registry with the aqua backend:

```shell
mise use -g aqua:nao1215/atago
```

If your `mise` does not see `nao1215/atago` yet, update `mise` or enable floating registries with `mise settings registry_floating=true`.

The [release page](https://github.com/nao1215/atago/releases) has prebuilt archives for Linux, macOS, and Windows (amd64/arm64; `.tar.gz`, or `.zip` on Windows), plus `.deb`, `.rpm`, and `.apk` packages for Linux. Building from source requires Go 1.26.11 or later.

CI runs the unit suite and the self-hosted spec suite on Linux, macOS, and Windows; on Windows the spec suite runs under both the native `cmd.exe` and a POSIX shell. [Platform support](https://nao1215.github.io/atago/reference/#platform-support) lists the few behaviors that differ and why.

## Getting started

### Record an interactive session

`atago record --pty -- <command>` records an interactive tool in a real terminal (a ConPTY on Windows). You drive one session by hand, and atago writes a `pty:` step that replays your keystrokes as expect/send pairs. On POSIX, a password typed at an echo-off line-mode prompt becomes an `${env:...}` placeholder. A TUI's raw mode also turns echo off, but its keystrokes are recorded literally, so replace a password typed into a TUI's own field with `${env:...}` by hand; on Windows, where ConPTY exposes no echo state, do that for every secret. `--timeout` (default 30s) bounds the session: if the program never exits, atago kills it, writes what it captured, and fails.

```shell
$ atago record --pty --out wizard.atago.yaml -- mytool init
```

`atago init` scaffolds a blank spec instead. Every spec has the same shape: run a command, then assert on what it produced.

### 1. Check exit code, stdout, and stderr

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

`atago run` takes spec files and directories, searched recursively for `*.atago.yaml` (or `*.atago.yml`). Each scenario runs in its own temporary directory, and progress prints one character per scenario (`.` pass, `F` fail, `E` error, `s` skip, `f` flaky, `x` XFAIL, `X` XPASS):

```shell
$ atago run ./specs
............

PASSED  12 scenarios: 12 passed, 0 failed, 0 errored, 0 skipped (1.4s)
```

Scenarios run concurrently, four per CPU by default, because a scenario spends most of its time waiting on its program. `--parallel N` changes the count and `--parallel 1` serializes. Workdirs are isolated but the host network is shared, so two scenarios that each start a background `service:` need distinct ports.

When a check fails, atago prints what was expected and what happened, with a colorized unified diff for multi-line output:

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

### 2. Check generated files and snapshots

`fixture:` writes input files into the workdir, `file:` and `dir:` assertions check what the command produced, and `snapshot:` pins output to a committed golden file with temp paths, UUIDs, and timestamps normalized. A fixture's source is one of `content:` (inline text), `base64:` (inline bytes), `from:` (copy an existing file), or `symlink:` (link to a target):

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

See [files_and_fixtures](examples/files_and_fixtures.atago.yaml), [snapshot](examples/snapshot.atago.yaml), and [dir_tree](examples/dir_tree.atago.yaml) for whole-tree golden manifests. Other checks:

- `count`, `min_count`, and `max_count` say how many times a `contains` or `matches` matcher occurs, which catches duplicated output that a presence check passes. `size`, `min_size`, and `max_size` bound a file's byte count, with or without content matchers; `size: 0` pins a file that is created but empty. See [count_and_size](examples/count_and_size.atago.yaml).
- `deterministic: {}` runs a command again and requires byte-identical output. It catches map iteration order leaking into a report, which loose matchers pass on every run. See [deterministic](examples/deterministic.atago.yaml).
- `suite.env` can carry a value that `suite.setup` captured, such as the address a suite-wide service published through `ready: {store:}`, so every scenario reaches a stub server without a shell wrapper exporting it. See [suite_env_from_setup](examples/suite_env_from_setup.atago.yaml).
- An `atago.project.yaml` beside your specs configures the whole directory: shared `env:` and `defaults:`, and `fixtures_dir:` for a committed corpus that every spec reads as `${fixtures}`. Its `subject:` builds the binary under test once before the suite runs and puts it first on `PATH`, and `profiles:` (`--profile cover`) swap in an instrumented build. The build is a command, so any language works. See [project_manifest](examples/project_manifest.atago.yaml).
- `expect_fail:` keeps a known, unfixed bug in CI without turning it red. The scenario runs on every commit and an expected failure counts as XFAIL (green); when it starts passing, the run fails so you promote the spec. See [expect_fail](examples/expect_fail.atago.yaml).

### 3. Drive interactive prompts and TUIs

A `pty` step runs the command in a real pseudo-terminal and drives it with expect/send pairs, so wizards, REPLs, and TTY-detection branches need no `expect(1)` script:

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

`send: {key: enter}` sends named keys, including control-byte aliases (`ctrl-space`, `ctrl-[`, `ctrl-_`) and modified keys such as `ctrl-hyphen`/`ctrl-minus` for apps that tell the physical `Ctrl+-` key apart. See [pty](examples/pty.atago.yaml), [pty_screen](examples/pty_screen.atago.yaml), and the cross-platform [pty_portable](examples/pty_portable.atago.yaml). A tool that draws its UI on stderr and prints its result on stdout is one step, with the UI on the rendered screen and the result in the redirected file: [pty_stdout_split](examples/pty_stdout_split.atago.yaml).

`pty` steps and `atago record --pty` run on Linux, macOS, and Windows (through a ConPTY pseudo-console); only `signal:` is POSIX-only. The `pty` and `pty_screen` examples skip on Windows because their inner commands (`[ -t 0 ]`, `cat -v`, a SIGINT trap) are POSIX-specific. ConPTY also does not report whether it echoes input: if output after a `send` starts with the sent text, atago treats that copy as possible echo even when it arrives late, so write the next `expect` against the program's response.

### When your CLI talks to a server

The same YAML drives HTTP, database, SSH, gRPC, headless browser, and offline mock server peers as dependencies of the CLI under test. `atago init --template <name>` scaffolds each:

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

## Examples

Every feature has a commented, runnable spec under [examples/](examples/), tested in CI on Linux, macOS, and Windows. The [cookbook](https://nao1215.github.io/atago/cookbook/) has 50+ copyable recipes, such as converting images, driving prompts and TUIs, simulating API failures offline, and proving idempotency, and indexes every spec by task and by feature.

These selection flags work with any spec, before, after, or between the paths:

- `--filter NAME` selects scenarios whose name contains NAME. Repeat it or separate names with commas for OR: `--filter a,b` and `--filter a --filter b` both run scenarios whose name contains `a` or `b`.
- `--tag T` and `--skip-tag T` match tags exactly, not as substrings. `atago list` shows the available tags.
- `--fail-fast` stops at the first outcome that fails the run. `--parallel N` and `--rerun-failed` also apply.
- `--verbose` traces every command, capture, and assertion verdict, for passing scenarios too.

## Use it in CI

Real E2E suites flake on timing, ports, and external tools. `--retry-failed N` runs a failed scenario again in a fresh workdir and reports a recovered one as flaky in every report format; atago does not retry silently. A flaky scenario still fails the run, because a result that depends on how many times it ran proves nothing. Pass `--allow-flaky` for a suite whose instability you accept. `--repeat N` runs each scenario N times to find flakiness before it reaches CI.

```shell
atago run --ci --retry-failed 2 ./specs          # retry, report instability loudly, still fail
atago run --ci --retry-failed 2 --allow-flaky ./specs  # ...and keep CI green when that is accepted
atago run --repeat 20 --filter "race prone" ./specs   # flake detection
```

[setup-atago](https://github.com/nao1215/setup-atago) installs a released binary:

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

On GitLab CI, or any CI that starts from a container image, use the GHCR image. Its entrypoint is `atago`, so clear it to let the runner start its shell:

```yaml
image:
  name: ghcr.io/nao1215/atago:latest
  entrypoint: [""]

stages: [test]

behavior-specs:
  stage: test
  script:
    - atago run --ci --report junit ./specs > junit.xml
  artifacts:
    when: always
    reports:
      junit: junit.xml
```

The image has only `atago` and `ca-certificates`. If your scenarios use `git`, `jq`, a browser, or your own CLI binary, build FROM `ghcr.io/nao1215/atago:latest` and add them.

- `--report json|junit|gha|tap` picks the report format. The JSON shape is stable and versioned.
- `--ci` makes output deterministic and color-free. It also fails the run (exit 3) when a `--filter`, `--tag`, or `--skip-tag` matches no scenario, so a typo cannot disable your specs. Without `--ci` this is a warning and the run exits 0.
- `--artifacts-dir DIR` saves the payloads a failed assertion compared, plus a failed scenario's service logs and each mock server's recorded requests, so you can review a failure after the job ends.
- Environment variables named under `secrets:` are masked as `***` in all reports and snapshots.

## Review specs without running them

`explain` describes what a spec does, `doc` generates Markdown with fixtures, expected payloads, and golden files inlined, `manifest` prints a stable JSON summary for tooling, and `list` shows scenarios, tags, and artifacts. Each loads and validates the spec first and exits 2 on a schema error, so any of them works as a lint step in CI:

![review](./doc/img/review.gif)

```shell
atago explain spec.atago.yaml
atago doc --out docs/specs.md ./specs
atago manifest ./specs
atago list ./specs
```

A spec error has a code, as in `spec.yaml: ATG2201: suite.name is required`. The first digit is the exit code, so `ATG2xxx` always exits 2. Codes are searchable and survive rewording of the message, and the [error reference](https://nao1215.github.io/atago/errors/) explains each one and what to change. Codes are assigned one family at a time; the reference lists which families have them today. Assertion failures have no code, since exit 1 means a spec caught a problem.

Scenario names say what happens. An optional `description:` on the suite or a scenario says why it matters. It is rendered under the matching heading in the generated Markdown, never expanded, and has no effect on the run. Keep maintenance notes in YAML comments, which the generated page leaves out.

## Let atago pick the inputs

`forall:` describes the shape of an input instead of listing values and expands into one scenario per generated row. This is property-based testing in a spec file:

```yaml
scenarios:
  - name: parsing never crashes, whatever the argument is
    forall:
      vars: {text: {type: ascii, min: 0, max: 40}}
    steps:
      - run: {command: mytool parse, stdin: "${text}"}
      - assert: {exit_code: {in: [0, 1]}}
```

The generators are `int`, `bool`, `digits`, `alpha`, `alphanumeric`, `ascii`, `unicode`, and `one_of` for a fixed set. They are seeded from the spec, not the clock, so the same file tests the same inputs on every machine, `atago list` shows them before anything runs, and a failing instance has its input in its name. Boundary values come first, since that is where a CLI breaks. See [forall](examples/forall.atago.yaml).

## Snapshot testing

`snapshot` matchers compare output against committed golden files, with ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF normalized so snapshots match across machines. Record or refresh them with:

![snapshot](./doc/img/snapshot.gif)

```shell
atago snapshot update spec.atago.yaml
```

For other volatile values, such as auto-increment IDs, request identifiers, or epoch times, declare spec-wide `scrub:` rules. Each rewrites regex matches to a placeholder before the compare, after `secrets:` masking:

```yaml
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
```

See [scrub](examples/scrub.atago.yaml).

## Editor support (JSON Schema)

[schema/atago.schema.json](schema/atago.schema.json) is a JSON Schema for specs. With the YAML language server it gives completion and validation for step types, every matcher, and the `${workdir}` / `${env:NAME}` / `${name}` / `$${...}` expansion rules. `atago init` and `atago record` write this header as the first line of every spec they generate. To add it to an existing spec, use the absolute URL, which resolves in any project:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
```

## Coding agents

An [Agent Skill](skills/atago/SKILL.md) for Claude Code and Codex is in `skills/atago`. Copy it to `~/.claude/skills/atago` or `~/.codex/skills/atago`.

## Shell completion

`atago completion <bash|zsh|fish|powershell>` prints a completion script for your shell.

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

`Ctrl-C` or `SIGTERM` stops the run: atago tears down in-flight processes, services, and sessions, reports partial results, and exits `4`.

## Real CLIs tested with atago

These suites run the author's Go tools (atago tests itself) and unmodified third-party binaries: git and jq, TUIs (fzf, htop, yazi), the python3 REPL, servers run as scenario services (redis, gitea, grafana, prometheus), cloud and IaC CLIs tested offline (aws-cli, terraform, ecspresso), crypto tools (openssl, age, sops), and document and media pipelines (pandoc, ffmpeg). Most were migrated from ShellSpec. The generated [real-world index](https://nao1215.github.io/atago/real-world/) has behavior docs for those binaries, such as Yazi file operations, chooser/cwd handoff, tabs, and sort modes.

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
