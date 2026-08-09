---
toc: true
title: Getting started
description: Record a real run of your CLI, read the spec atago wrote, and keep it as a test — then assert on files, snapshots, interactive prompts, and server peers.
---

## Try it in 30 seconds

Before installing anything, prove the loop against a command you already have. If you have Go, paste this — it records a real run and replays it as a test:

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped
```

Open `demo.atago.yaml`: `record` captured the exit code, the version line on stdout, and an empty stderr, so you have a real test to tighten rather than YAML written from scratch. Swap `git --version` for any command you have (`go version`, `jq --version`, `ls -la`). Then [install atago](/install/) and point it at your own tool.

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
.....................................................................................................

PASSED  160 scenarios: 160 passed, 0 failed, 0 errored, 0 skipped (20.5s)
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

For full-screen TUIs, `expect_screen:` waits on the LIVE rendered frame during the session, and `screen:` asserts the final rendered frame after exit:

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

Named keys (`send: {key: enter}`) and rendered-screen checks cover full TUIs — including control-byte aliases like `ctrl-space`, `ctrl-[`, and `ctrl-_`, plus modified key events like `ctrl-hyphen`/`ctrl-minus` for apps that distinguish the physical `Ctrl+-` key — see [pty](https://github.com/nao1215/atago/blob/main/examples/pty.atago.yaml), [pty_screen](https://github.com/nao1215/atago/blob/main/examples/pty_screen.atago.yaml), and the cross-platform [pty_portable](https://github.com/nao1215/atago/blob/main/examples/pty_portable.atago.yaml).

The vocabulary covers the chords a real TUI binds, so a session never has to embed a raw `\x1b` escape: `shift-tab` (alias `backtab`) for reverse focus, `alt-a`..`alt-z` plus `alt-enter` and `alt-backspace` for the ESC-prefixed Meta chords readline word operations use, and the arrows with `ctrl-`/`shift-` modifiers (`ctrl-left`, `shift-up`, …) for word-wise and selection movement. A misspelled key is a load-time error that prints the whole accepted list, never a keystroke that silently does nothing. One caveat worth knowing: `alt-backspace` ends in DEL, which a cooked terminal claims as its ERASE character — a program that reads keys in raw mode (every full-screen TUI does) sees the chord intact, while a line-disciplined prompt sees an erase.

`exec: "..."` runs one command on the host while the program under test keeps running, so a session can reach the paths no keystroke can: a git client refreshing after a commit made outside it, a file manager showing a file another process created, a log viewer following a file that grows. It blocks until the command exits, which is the point — after it the change exists, so the `expect_screen` that follows is waiting on the program noticing rather than on a race. The command runs in the scenario workdir with the environment the pty child got, so `sandbox_home` and `clear_env` isolation still holds, and a non-zero exit or timeout ends the step with an error rather than leaving the next wait to blame the program for a change nobody made.

`resize: {rows: 40, cols: 120}` changes the terminal size while the program is running, delivered the way a real terminal delivers it — `SIGWINCH` on POSIX, a ConPTY notification on Windows. Relayout is where full-screen TUIs break (a stale right edge, a panel keeping its old width, a crash on a zero-width column), and without it that path is fixed at whatever the step started with. The rendered screen follows: each part of the transcript is drawn at the size it was produced under, so a frame written before the resize keeps the wrapping it was written with. Settle the screen with `expect_screen` on both sides of a resize — output still in flight when it lands is attributed to the old size, exactly as on a real terminal.

`send: {paste: "..."}` delivers text as a **bracketed paste** rather than as typing. A REPL or editor that enables the mode treats a pasted block as one unit — it must not run line by line, auto-indent, or fire completion — and that is a different code path from typing the same characters. atago refuses the send if the program has not enabled the mode (`ESC [?2004h`), because the markers would otherwise arrive as literal `[200~` text and surface as a puzzling failure much later; wait for the prompt with an `expect` first, since programs turn the mode on during startup. A tool that draws its interface on stderr and prints its result on stdout is testable as one step — the UI on the rendered screen, the result in the redirected file: [pty_stdout_split](https://github.com/nao1215/atago/blob/main/examples/pty_stdout_split.atago.yaml). `pty` steps and `atago record --pty` run on Linux, macOS, and Windows (where they drive a ConPTY pseudo-console); only `signal:` stays POSIX-only.

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

The [cookbook](/cookbook/) has a copyable spec for most jobs and indexes the runnable example for every feature, and [Use it in CI](/ci/) wires a suite into GitHub Actions.
