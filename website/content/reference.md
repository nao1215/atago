---
toc: true
title: Reference
description: atago subcommands, scenario selection flags, snapshot updating and scrub rules, the JSON Schema for editors, shell completion, and the exit code contract.
---

## Subcommands

| Command | Does |
|---------|------|
| `atago run` | run specs and report results |
| `atago record` | run a command once and write a spec from what it observed (`--pty` for interactive sessions) |
| `atago init` | scaffold a spec (`--template` for browser, cli, db, grpc, http, mock, services, ssh; `cli` is the default) |
| `atago snapshot update` | record or refresh golden files |
| `atago explain` | describe what a spec does without running it |
| `atago doc` | generate Markdown from specs, with fixtures and golden files inlined |
| `atago manifest` | emit a stable JSON summary of specs for tooling |
| `atago list` | show scenarios, tags, and artifacts |
| `atago completion` | print a shell completion script |

`explain`, `doc`, `manifest`, and `list` all load and validate the spec first — exit code 2 on a schema error — so any of them doubles as a lint step in CI.

## Selecting scenarios

Selection flags compose with any spec: `--filter NAME` (repeatable, and comma-separated for OR — `--filter a,b` or `--filter a --filter b` runs scenarios whose name contains `a` or `b`), `--tag T` and `--skip-tag T` (tags match exactly, not by substring — `atago list` shows the available tags), `--parallel N`, `--fail-fast`, and `--rerun-failed`. `atago run --rerun-failed` re-runs only the scenarios the previous run recorded as failed in `.atago/last-failed.json`, so the fix-and-recheck loop replays just the failures instead of the whole suite. While authoring, `--verbose` traces every command, capture, and assertion verdict — for passing scenarios too. Under `--ci`, a selection that matches no scenario fails the run (exit 3) rather than passing an empty suite. Flags may be written before, after, or between the paths — `atago run ./specs --report json` and `atago run --report json ./specs` are the same command — and `--` ends flag parsing, so a path spelled like a flag can still be named. `atago record` is the exception: everything after its `--` is the recorded program's own command line.

## Snapshot testing

`snapshot` matchers compare output against committed golden files; ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF are normalized so snapshots stay stable across machines. Record or refresh them with:

```shell
atago snapshot update spec.atago.yaml
```

For volatile patterns the built-ins do not cover — auto-increment IDs, request identifiers, epoch times — declare spec-wide `scrub:` rules that rewrite each regex match to a placeholder before the compare (applied after `secrets:` masking):

```yaml
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
```

See the [scrub example](https://github.com/nao1215/atago/blob/main/examples/scrub.atago.yaml).

## Spec file keys

Every key a spec file accepts, generated from the committed
[JSON Schema](https://github.com/nao1215/atago/blob/main/schema/atago.schema.json)
— the same document that powers editor completion, so this reference cannot
drift from what the loader accepts. Indentation shows nesting; a type that
links (like [step](#spec-step)) is documented in its own section rather than
repeated inline.

All keys belong to spec format version `1` — the only format version so far;
`version: "1"` is the first line of every spec. The Since column is the atago release
that introduced the key (`unreleased` = merged to main, not yet in a tagged
release).

{{< spec-reference >}}

## Directory manifest

A directory of specs may carry an `atago.project.yaml` holding what belongs to the tree rather than to one file:

```yaml
env:
  MYTOOL_REGISTRY: "http://127.0.0.1:8080"
defaults:
  run:
    sandbox_home: true
fixtures_dir: testdata
```

It can also declare the binary under test:

```yaml
subject:
  name: mytool
  artifact: bin/mytool
  build:
    command: "go build -o ${artifact} ."
    cwd: ".."
profiles:
  cover:
    build:
      command: "go build -cover -covermode=atomic -coverpkg=./... -o ${artifact} ."
    env:
      GOCOVERDIR: "${env:GOCOVERDIR}"
```

`atago run` builds it once per invocation, before any scenario, and prepends the artifact's directory to `PATH`. `--profile NAME` swaps in that profile's build command (whole-command replacement) and layers its `env`. A failing build — or one that exits 0 without writing `${artifact}` — is a run-level error and no scenario executes.

It is discovered by walking up from a spec to the nearest one, so `atago run ./e2e` and `atago run ./e2e/one.atago.yaml` resolve the same configuration. Precedence is host < project < suite < scenario < step for `env`, and a spec file's own `defaults:` beat the manifest's. `fixtures_dir` resolves against the manifest's directory and must exist at load time. `atago explain` prints the manifest that applied and the resolved fixtures directory. Its own schema is [atago.project.schema.json](https://github.com/nao1215/atago/blob/main/schema/atago.project.schema.json).

| Variable | Is |
|----------|----|
| `${workdir}` | this scenario's isolated temp directory — the only one it owns |
| `${suitedir}` | the suite's scratch directory, shared by `suite.setup` and every scenario |
| `${specdir}` | the directory holding the spec file (read-only input) |
| `${fixtures}` | the manifest's `fixtures_dir` (read-only input); unset when no manifest declares one |

All four are absolute, because a scenario runs somewhere other than where its spec lives.

## Editor support (JSON Schema)

A JSON Schema lives at [schema/atago.schema.json](https://github.com/nao1215/atago/blob/main/schema/atago.schema.json). With the YAML language server you get completion and validation as you type — step types, every matcher, and the `${workdir}` / `${env:NAME}` / `${name}` / `$${...}` expansion rules. `atago init` and `atago record` already emit this header as the first line of every generated spec, so scaffolded specs get completion out of the box. To add it to an existing spec, use the absolute URL (it resolves in any project, unlike a repo-relative path):

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
```

The report and manifest outputs have schemas too: [report.schema.json](https://github.com/nao1215/atago/blob/main/schema/report.schema.json) and [manifest.schema.json](https://github.com/nao1215/atago/blob/main/schema/manifest.schema.json).

## Shell completion

`atago completion <bash|zsh|fish|powershell>` prints a completion script for your shell.

## Exit codes

| Code | Meaning                   |
|------|---------------------------|
| `0`  | no unexpected failures — every scenario passed, was skipped, or was an XFAIL (an `expect_fail:` scenario that failed as declared) |
| `1`  | one or more failed (including an XPASS: an `expect_fail:` scenario that passed, unless `--allow-xpass`) |
| `2`  | spec error (YAML syntax or schema/semantic validation) |
| `3`  | CLI-invocation error (unknown subcommand, bad flag, or no matching spec files) |
| `4`  | execution error           |
| `5`  | internal error            |
| `6`  | security policy violation |

`Ctrl-C`/`SIGTERM` stops the run cleanly: in-flight processes, services, and sessions are torn down, partial results are reported, and the run exits `4`.

Errors also carry a diagnostic code such as `ATG2201`, whose first digit is the exit code above — so `ATG2xxx` always exits 2. The codes are searchable and stable across rewordings of the message; [Error codes](/errors/) lists what each one means, what to change, and which families carry codes today. Assertion failures carry none: exit 1 is a result, not an error.
