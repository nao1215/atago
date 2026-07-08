---
title: Getting started
description: Record a real run of your CLI, read the spec atago wrote, and keep it as a test — then assert on files, snapshots, interactive prompts, and server peers.
---

## Start from a real run

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

## 1. Check exit code, stdout, and stderr

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

## 2. Check generated files and snapshots

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

See [files_and_fixtures](https://github.com/nao1215/atago/blob/main/examples/files_and_fixtures.atago.yaml), [snapshot](https://github.com/nao1215/atago/blob/main/examples/snapshot.atago.yaml), and [dir_tree](https://github.com/nao1215/atago/blob/main/examples/dir_tree.atago.yaml) for whole-tree golden manifests.

## 3. Drive interactive prompts and TUIs

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

Named keys (`send: {key: enter}`) and asserts on the RENDERED terminal screen cover full TUIs — see [pty](https://github.com/nao1215/atago/blob/main/examples/pty.atago.yaml), [pty_screen](https://github.com/nao1215/atago/blob/main/examples/pty_screen.atago.yaml), and the cross-platform [pty_portable](https://github.com/nao1215/atago/blob/main/examples/pty_portable.atago.yaml). `pty` steps and `atago record --pty` run on Linux, macOS, and Windows (where they drive a ConPTY pseudo-console); only `signal:` stays POSIX-only.

## When your CLI talks to a server

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

## Next

The [cookbook](/cookbook/) has a copyable spec for most jobs, the [examples](/examples/) index covers every feature, and [Use it in CI](/ci/) wires a suite into GitHub Actions.
