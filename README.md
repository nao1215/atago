<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

![Coverage](https://raw.githubusercontent.com/nao1215/octocovs-central-repo/main/badges/nao1215/atago/coverage.svg)
[![Build](https://github.com/nao1215/atago/actions/workflows/build.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/atago/actions/workflows/unit_test.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/unit_test.yml)
[![reviewdog](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nao1215/atago.svg)](https://pkg.go.dev/github.com/nao1215/atago)
![GitHub](https://img.shields.io/github/license/nao1215/atago)

<p align="center">
  <img src="./doc/img/atago-logo.jpg" alt="atago logo" width="400" />
</p>

atago tests real CLI behavior from plain YAML: commands, files, snapshots, services, and interactive terminals. It runs your actual binary — in any language — and asserts what a user observes. No test code, no shell DSL.

Documentation: **https://nao1215.github.io/atago/**

![demo](./doc/img/demo.gif)

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

The [release page](https://github.com/nao1215/atago/releases) contains prebuilt binary archives for Linux, macOS, and Windows (amd64/arm64; `.tar.gz`, or `.zip` on Windows), plus `.deb`, `.rpm`, and `.apk` packages for Linux. Requires Go 1.26 or later when building from source.

Runs on Linux, macOS, and Windows (CI tests all three).

## Verifying release integrity

Every release ships supply-chain metadata so you can verify what you download:

- Signed checksums: `checksums.txt` is signed with [cosign](https://github.com/sigstore/cosign) (keyless), producing `checksums.txt.sigstore.json`.
- SBOM: an SPDX Software Bill of Materials is attached to each release archive.
- Build provenance: SLSA build provenance is attested via GitHub OIDC.

```shell
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/nao1215/atago/\.github/workflows/release\.yml@refs/tags/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

```shell
gh attestation verify atago_<version>_<os>_<arch>.tar.gz --repo nao1215/atago  # .zip on Windows
```

## Getting started

### Start from a real run

You don't write the first spec — your tool does. `atago record -- <command>` runs it once and generates a spec from what it observed (exit code, output, created files):

```shell
$ atago record --out mytool.atago.yaml -- mytool convert input.txt
recorded: exit 0, 2 stdout line(s), 1 file(s) created
wrote mytool.atago.yaml
$ atago run mytool.atago.yaml
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped (12ms)
```

Interactive tools record too: `atago record --pty -- <command>` runs it in a real terminal, lets you drive one session by hand, and writes a `pty:` step that replays your keystrokes as expect/send pairs. It works on Linux, macOS, and Windows (a ConPTY); on POSIX a password prompt becomes an `${env:...}` placeholder automatically, while on Windows — where a ConPTY exposes no echo state — you convert a secret send to `${env:...}` by hand. A `--pty` session is bounded by `--timeout` (default 30s): if the program never exits, atago kills it, writes whatever was captured, and fails instead of hanging forever:

```shell
$ atago record --pty --out wizard.atago.yaml -- mytool init
```

Prefer a blank template? `atago init` scaffolds one. Either way, the shape is always the same: declare a command, run it, assert on what it produced.

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

`atago run` accepts spec files and directories (searched recursively for `*.atago.yaml`; the `*.atago.yml` spelling is accepted too). Each scenario runs in its own temporary directory, and progress streams as a dot per scenario (`.` pass, `F` fail, `E` error, `s` skip):

```shell
$ atago run ./specs
...............................................

PASSED  47 scenarios: 47 passed, 0 failed, 0 errored, 0 skipped (1.2s)
```

Scenarios run concurrently by default (`--parallel N`, defaulting to your CPU count; set `--parallel 1` to serialize). Workdirs are isolated, but the host network is shared — so if two scenarios each start a background `service:`, give them distinct ports, or one scenario's requests can reach the other's server.

When a check fails, atago prints exactly what was expected and what happened; multi-line mismatches render a colorized unified diff:

```text
FAILED: demo / greeting matches its golden

Step:
  assert stdout snapshot

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
```

### 2. Check generated files and snapshots

`fixture:` writes input files into the isolated workdir; `file:`/`dir:` assertions check what the command produced, and `snapshot:` pins output to a committed golden file (volatile details like temp paths, UUIDs, and timestamps are normalized). A fixture's source is one of `content:` (inline text), `base64:` (inline bytes), `from:` (copy an existing file), or `symlink:` (link to a target):

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

See [files_and_fixtures](examples/files_and_fixtures.atago.yaml), [snapshot](examples/snapshot.atago.yaml), and [dir_tree](examples/dir_tree.atago.yaml) for whole-tree golden manifests.

### 3. Drive interactive prompts and TUIs

A `pty` step runs the command in a real pseudo-terminal and drives it with a declarative expect/send session — wizards, REPLs, and TTY-detection branches, no `expect(1)` scripting:

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

Named keys (`send: {key: enter}`) and asserts on the RENDERED terminal screen cover full TUIs — see [pty](examples/pty.atago.yaml), [pty_screen](examples/pty_screen.atago.yaml), and the cross-platform [pty_portable](examples/pty_portable.atago.yaml). `pty` steps and `atago record --pty` run on Linux, macOS, and Windows (where they drive a ConPTY pseudo-console); only `signal:` stays POSIX-only. The `pty`/`pty_screen` examples skip on Windows because their inner commands (`[ -t 0 ]`, `cat -v`, a SIGINT trap) are POSIX-specific, not because the `pty` mechanism is.

### When your CLI talks to a server

The same YAML also drives HTTP, database, SSH, gRPC, headless-browser, and offline mock-server peers — as dependencies of the CLI under test. `atago init --template <name>` scaffolds each:

```shell
$ atago init --list-templates
browser   drive a headless Chrome; assert page content (needs Chrome on PATH)
cli       run a command; assert exit code/stdout/stderr (runs as-is)
db        run SQL; assert on rows via bundled SQLite (runs as-is)
grpc      call a unary gRPC method via server reflection (edit target first)
http      call an HTTP API; assert status and JSON body (edit base_url first)
mock      stub an HTTP API offline and assert what the client sent (needs curl on PATH)
services  test against a background server: readiness, retry, teardown (runs as-is)
ssh       run a command on a remote host over SSH (edit host/user first)
```

## Examples

Every feature has a commented, runnable spec under [examples/](examples/), tested in CI on Linux, macOS, and Windows. The **[cookbook](https://nao1215.github.io/atago/cookbook/)** collects 50+ copyable recipes for common jobs — converting images, driving prompts and TUIs, simulating API failures offline, proving idempotency — and the [examples index](https://nao1215.github.io/atago/examples/) lists every spec by task and by feature.

Selection flags compose with any spec: `--filter NAME` (repeatable, and comma-separated for OR — `--filter a,b` or `--filter a --filter b` runs scenarios whose name contains `a` or `b`), `--tag T`, `--skip-tag T`, `--parallel N`, `--fail-fast`, and `--rerun-failed`. While authoring, `--verbose` traces every command, capture, and assertion verdict — for passing scenarios too.

## Use it in CI

Real E2E suites flake (timing, ports, external tools). `--retry-failed N` re-runs failed scenarios in a fresh workdir and reports recovered ones as flaky — green for the exit code, but loud in every report format; silent retries are explicitly a non-goal. `--repeat N` does the opposite job: run each scenario N times to detect flakiness before it reaches CI.

```shell
atago run --ci --retry-failed 2 ./specs          # keep CI green, report instability loudly
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
      - uses: actions/checkout@v4
      - uses: nao1215/setup-atago@v0
      - run: atago run --ci --report gha ./specs
```

- `--report json|junit|gha|tap` picks the report format; the JSON shape is stable and versioned.
- `--ci` enables deterministic, color-free output.
- `--artifacts-dir DIR` persists the exact payloads a failed assertion compared, so a failure stays reviewable after the job ends.
- Environment variable names listed under `secrets:` are masked as `***` in all reports and snapshots.

## Review specs without running them

`explain` describes what a spec does, `doc` generates Markdown (with fixtures, expected payloads, and golden files inlined), `manifest` emits a stable JSON summary for tooling, and `list` shows scenarios, tags, and artifacts. All of them load and validate the spec first — exit code 2 on a schema error — so any of them doubles as a lint step in CI:

![review](./doc/img/review.gif)

```shell
atago explain spec.atago.yaml
atago doc --out docs/specs.md ./specs
atago manifest ./specs
atago list ./specs
```

## Snapshot testing

`snapshot` matchers compare output against committed golden files; ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF are normalized so snapshots stay stable across machines. Record or refresh them with:

![snapshot](./doc/img/snapshot.gif)

```shell
atago snapshot update spec.atago.yaml
```

For volatile patterns the built-ins do not cover — auto-increment IDs, request identifiers, epoch times — declare spec-wide `scrub:` rules that rewrite each regex match to a placeholder before the compare (applied after `secrets:` masking):

```yaml
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
```

See [scrub](examples/scrub.atago.yaml).

## Editor support (JSON Schema)

A JSON Schema lives at [schema/atago.schema.json](schema/atago.schema.json). With the YAML language server you get completion and validation as you type — step types, every matcher, and the `${workdir}` / `${env:NAME}` / `${name}` / `$${...}` expansion rules. `atago init` and `atago record` already emit this header as the first line of every generated spec, so scaffolded specs get completion out of the box. To add it to an existing spec, use the absolute URL (it resolves in any project, unlike a repo-relative path):

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
```

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

`Ctrl-C`/`SIGTERM` stops the run cleanly: in-flight processes, services, and sessions are torn down, partial results are reported, and the run exits `4`.

## Real CLIs tested with atago

These suites run real programs of every shape: the author's Go tools (atago tests itself) and unmodified third-party binaries — git and jq, interactive TUIs (fzf, htop), the python3 REPL, servers driven as scenario services (redis, gitea, grafana, prometheus), cloud and IaC CLIs tested offline (aws-cli, terraform, ecspresso), crypto tools (openssl, age, sops), and document/media pipelines (pandoc, ffmpeg). Most were migrated from ShellSpec. [doc/real-world.md](doc/real-world.md) lists all 40+ with specs and generated behavior docs.

## The name

atago (愛宕) is named for Mount Atago in Kyoto, whose shrine enshrines a deity of fire prevention. A test runner should do the same job: catch the sparks so a project never catches fire.

## Contributing

Issues and pull requests are welcome; see [CONTRIBUTING.md](./CONTRIBUTING.md). Contributions are not only about code: a GitHub Star also motivates development.

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
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
