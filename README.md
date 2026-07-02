<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

![Coverage](https://raw.githubusercontent.com/nao1215/octocovs-central-repo/main/badges/nao1215/atago/coverage.svg)
[![Build](https://github.com/nao1215/atago/actions/workflows/build.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/atago/actions/workflows/unit_test.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/unit_test.yml)
[![reviewdog](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nao1215/atago.svg)](https://pkg.go.dev/github.com/nao1215/atago)
[![Go Report Card](https://goreportcard.com/badge/github.com/nao1215/atago)](https://goreportcard.com/report/github.com/nao1215/atago)
![GitHub](https://img.shields.io/github/license/nao1215/atago)

# atago

atago is an end-to-end test runner for command-line tools. It runs a real command, in any language, and checks what a user observes: the exit code, stdout, stderr, generated files, JSON output, and snapshots. Specs are plain YAML — no test code, no shell DSL, no embedded scripting. It also drives HTTP, database, SSH, gRPC, and headless-browser peers, and generates Markdown docs from specs.

![demo](./doc/img/demo.gif)

```shell
atago init                       # write a starter example.atago.yaml
atago run example.atago.yaml     # run the spec you just created
atago run --report junit specs/  # emit a JUnit report for CI
```

## Install

```shell
go install github.com/nao1215/atago@latest
```

On macOS, Homebrew works too:

```shell
brew install --cask nao1215/tap/atago
```

The [release page](https://github.com/nao1215/atago/releases) contains prebuilt binary archives for Linux, macOS, and Windows (amd64/arm64; `.tar.gz`, or `.zip` on Windows). Requires Go 1.26 or later when building from source.

Runs on Linux, macOS, and Windows. CI runs the unit tests on all three, the full E2E suite on Linux and macOS, and a portable E2E subset on Windows; specs that lean on POSIX-only shell tools are the remaining gap on Windows.

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

`atago init` scaffolds a runnable spec. The shape is always the same: declare a command, run it, assert on what it produced.

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

`atago run` accepts spec files and directories (searched recursively for `*.atago.yaml`). Each scenario runs in its own temporary directory, and progress streams as a dot per scenario (`.` pass, `F` fail, `E` error, `s` skip):

```shell
$ atago run ./specs
...............................................

PASSED  47 scenarios: 47 passed, 0 failed, 0 errored, 0 skipped (1.2s)
```

When a check fails, atago prints exactly what was expected and what happened:

```text
FAILED: demo / expect Alice but the command prints Bob

Step:
  assert stdout contains "Alice"

Command:
  echo Bob

Expected:
  stdout contains "Alice"

Actual:
  Bob

Hint:
  the substring "Alice" was not present in stdout
```

`atago init --template <name>` scaffolds a starter for the other runner families; `atago init --list-templates` describes each one and says whether it runs as-is or what to edit first:

```shell
$ atago init --list-templates
browser   drive a headless Chrome; assert page content (needs Chrome on PATH)
cli       run a command; assert exit code/stdout/stderr (runs as-is)
db        run SQL; assert on rows via bundled SQLite (runs as-is)
grpc      call a unary gRPC method via server reflection (edit target first)
http      call an HTTP API; assert status and JSON body (edit base_url first)
services  test against a background server: readiness, retry, teardown (runs as-is)
ssh       run a command on a remote host over SSH (edit host/user first)
```

## Examples

Every feature has a commented, runnable spec under [examples/](examples/). The examples are loaded, validated, and (where they need no external server) executed in CI on Linux, macOS, and Windows, so they cannot drift from the implementation.

| Example | Shows |
|---------|-------|
| [run_and_assert](examples/run_and_assert.atago.yaml) | exit code, stdout/stderr matchers (`contains`, `equals`, `matches`, lists, `line`), multi-target asserts |
| [shell_and_redirect](examples/shell_and_redirect.atago.yaml) | `shell: true` vs direct argv execution, `stdout_to`/`stderr_to` redirects |
| [json_and_yaml](examples/json_and_yaml.atago.yaml) | JSONPath assertions, numeric bounds (`gt`/`lte`), the `yaml` matcher |
| [files_and_fixtures](examples/files_and_fixtures.atago.yaml) | input fixtures (text and base64), `file` and `dir` assertions |
| [store_and_variables](examples/store_and_variables.atago.yaml) | capturing values into `${name}`, `${workdir}`, the `$${...}` literal escape |
| [matrix](examples/matrix.atago.yaml) | one template scenario expanded per parameter row |
| [retry](examples/retry.atago.yaml) | polling a command until an assertion passes |
| [snapshot](examples/snapshot.atago.yaml) | golden-file testing with normalized output |
| [services](examples/services.atago.yaml) | background servers: readiness probes, `ready.store`, teardown |
| [defaults](examples/defaults.atago.yaml) | sharing `shell`/`env`/`service` fragments across scenarios |
| [select_skip_only](examples/select_skip_only.atago.yaml) | tags, and gating scenarios by OS, env var, or a probe command |
| [db](examples/db.atago.yaml) | SQL via the bundled SQLite driver, `rows` assertions, value binding |
| [image_and_pdf](examples/image_and_pdf.atago.yaml) | image format/dimension/similarity checks, PDF page/metadata/text checks |
| [http](examples/http.atago.yaml) | HTTP requests (`json:`, raw `body:`, form/multipart uploads, `body_file`), status/body assertions, token capture, `retry` polling, redirect assertions, `body_to` downloads, network allowlist |
| [ssh](examples/ssh.atago.yaml) | running commands on a remote host |
| [grpc](examples/grpc.atago.yaml) | unary gRPC calls via server reflection |
| [browser](examples/browser.atago.yaml) | headless-Chrome flows and screenshots |

Selection flags compose with any spec: `--filter NAME`, `--tag T`, `--skip-tag T`, `--parallel N` (default: the number of CPUs — scenarios are isolated, so runs are concurrent out of the box), `--fail-fast`, and `--rerun-failed` (rerun only what failed last time).

## Use it in CI

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

## Editor support (JSON Schema)

A JSON Schema lives at [schema/atago.schema.json](schema/atago.schema.json). With the YAML language server you get completion and validation as you type:

```yaml
# yaml-language-server: $schema=./schema/atago.schema.json
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

These real programs run their end-to-end suites on atago; most were migrated from ShellSpec. The generated behavior docs live in [doc/e2e/](doc/e2e/).

Programs maintained by the author, starting with atago itself ([test/e2e/tools/](test/e2e/tools/)):

| Tool | Feature | Specs | Docs |
|------|---------|-------|------|
| [atago](https://github.com/nao1215/atago) | atago tested by atago: the self-hosted specs run the real built binary in CI (`make e2e`). | [specs](test/e2e/atago) | [docs](doc/e2e/atago.md) |
| [gup](https://github.com/nao1215/gup) | Updates and manages the Go command-line tools in `$GOBIN`. | [specs](test/e2e/tools/gup) | [docs](doc/e2e/gup.md) |
| [sqly](https://github.com/nao1215/sqly) | Runs SQL against CSV/TSV/LTSV/JSON/Parquet/Excel/ACH/Fedwire files. | [specs](test/e2e/tools/sqly) | [docs](doc/e2e/sqly.md) |
| [truss](https://github.com/nao1215/truss) | Image transformation (convert/resize/re-encode). | [specs](test/e2e/tools/truss) | [docs](doc/e2e/truss.md) |
| [iso8583tool](https://github.com/nao1215/iso8583tool) | Debugs and inspects ISO 8583 payment messages. | [specs](test/e2e/tools/iso8583tool) | [docs](doc/e2e/iso8583tool.md) |
| [jose](https://github.com/nao1215/jose) | Signs and encrypts with JOSE. | [specs](test/e2e/tools/jose) | [docs](doc/e2e/jose.md) |
| [career](https://github.com/nao1215/career) | Renders résumé PDFs from a single YAML file. | [specs](test/e2e/tools/career) | [docs](doc/e2e/career.md) |
| [mimixbox](https://github.com/nao1215/mimixbox) | Packs many Unix commands into one BusyBox-style binary. | [specs](test/e2e/tools/mimixbox) | [docs](doc/e2e/mimixbox.md) |
| [mobilepkg](https://github.com/nao1215/mobilepkg) | Inspects Android packages for metadata and security findings. | [specs](test/e2e/tools/mobilepkg) | [docs](doc/e2e/mobilepkg.md) |

Third-party programs ([test/e2e/thirdparty/](test/e2e/thirdparty/)). atago runs each as an unmodified binary and ships only its own spec YAML — no third-party code is copied or redistributed, which every listed license permits:

| Program | Feature | License | Specs | Docs |
|---------|---------|---------|-------|------|
| [git](https://git-scm.com/) | Version control — runs in CI on all three OSes. | GPL-2.0 | [specs](test/e2e/thirdparty/git) | [docs](doc/e2e/git.md) |
| [caddy](https://caddyserver.com/) | Self-hosted web server, booted from an authored Caddyfile and queried over HTTP. | Apache-2.0 | [specs](test/e2e/thirdparty/caddy) | [docs](doc/e2e/caddy.md) |
| [coredns](https://coredns.io/) | Self-hosted DNS server: an authored zone queried with real `dig` — authoritative answers, CNAME chasing, NXDOMAIN/REFUSED, and the health plugin over HTTP. | Apache-2.0 | [specs](test/e2e/thirdparty/coredns) | [docs](doc/e2e/coredns.md) |
| [gitea](https://about.gitea.com/) | Self-hosted git service: booted with SQLite, administered via its CLI, driven over the REST API (repos, commits, issues), then cloned with real git. | MIT | [specs](test/e2e/thirdparty/gitea) | [docs](doc/e2e/gitea.md) |
| [gotify](https://gotify.net/) | Self-hosted notification server: app provisioning, token-authenticated pushes, and the app icon uploaded as real multipart/form-data, downloaded back, and verified as a PNG. | MIT | [specs](test/e2e/thirdparty/gotify) | [docs](doc/e2e/gotify.md) |
| [grafana](https://grafana.com/oss/grafana/) | Self-hosted observability platform: health/build info, the login redirect asserted with `follow_redirects: false`, and a dashboard + datasource lifecycle over the REST API. | AGPL-3.0 | [specs](test/e2e/thirdparty/grafana) | [docs](doc/e2e/grafana.md) |
| [mailpit](https://mailpit.axllent.org/) | Self-hosted email testing: messages delivered over real SMTP (stock curl), then asserted via the REST API — capture, search, MIME attachments, teardown. | MIT | [specs](test/e2e/thirdparty/mailpit) | [docs](doc/e2e/mailpit.md) |
| [minio](https://min.io/) | Self-hosted S3-compatible object storage: full object lifecycle via `mc`, versioning, anonymous bucket policies, S3 XML error contract. | AGPL-3.0 | [specs](test/e2e/thirdparty/minio) | [docs](doc/e2e/minio.md) |
| [nats](https://nats.io/) | Self-hosted messaging: request/reply through the real broker, JetStream persistence (create → publish → count → purge), the KV store, and the monitoring endpoint. | Apache-2.0 | [specs](test/e2e/thirdparty/nats) | [docs](doc/e2e/nats.md) |
| [ntfy](https://ntfy.sh/) | Self-hosted push notifications: publish with headers, poll the JSON feed, topic isolation, and deny-all access control unlocked via the admin CLI. | Apache-2.0 | [specs](test/e2e/thirdparty/ntfy) | [docs](doc/e2e/ntfy.md) |
| [prometheus](https://prometheus.io/) | Self-hosted monitoring: promtool config/rule checks and rule unit tests, the query API, and a self-scrape polled with http `retry`. | Apache-2.0 | [specs](test/e2e/thirdparty/prometheus) | [docs](doc/e2e/prometheus.md) |
| [pushgateway](https://github.com/prometheus/pushgateway) | Self-hosted metrics gateway; exercises raw text `body:` payloads. | Apache-2.0 | [specs](test/e2e/thirdparty/pushgateway) | [docs](doc/e2e/pushgateway.md) |
| [rclone](https://rclone.org/) | Self-hosted file sync: copy/sync/check semantics (including the corruption failure mode), JSON listings, and `rclone serve http`. | MIT | [specs](test/e2e/thirdparty/rclone) | [docs](doc/e2e/rclone.md) |
| [restic](https://restic.net/) | Self-hosted backup: init → backup → restore round-trip, snapshot JSON, diff, integrity check, retention, wrong-password failure mode. | BSD-2-Clause | [specs](test/e2e/thirdparty/restic) | [docs](doc/e2e/restic.md) |
| [transfer.sh](https://github.com/dutchcoders/transfer.sh) | Self-hosted file sharing: a PNG uploaded as the raw request body (`body_file`), downloaded back to disk (`body_to`), and verified byte-for-byte; multipart uploads too. | MIT | [specs](test/e2e/thirdparty/transfersh) | [docs](doc/e2e/transfersh.md) |
| [webhook](https://github.com/adnanh/webhook) | Self-hosted webhook receiver driven by a config fixture. | MIT | [specs](test/e2e/thirdparty/webhook) | [docs](doc/e2e/webhook.md) |

## The name

atago (愛宕) is named for Mount Atago in Kyoto, whose shrine enshrines a deity of fire prevention. A test runner should do the same job: catch the sparks so a project never catches fire.

## Contributing

Issues and pull requests are welcome; see [CONTRIBUTING.md](./CONTRIBUTING.md). Contributions are not only about code: a GitHub Star also motivates development.

[![Star History Chart](https://api.star-history.com/svg?repos=nao1215/atago&type=Date)](https://star-history.com/#nao1215/atago&Date)

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
