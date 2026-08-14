---
toc: true
title: Getting started
description: Record a real run of your CLI, read the spec atago wrote, and keep it as a test — then assert on files, snapshots, interactive prompts, and server peers.
---

## First run

Before installing anything, prove the loop against a command you already have. If you have Go, paste this — it records a real run and replays it as a test:

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped
```

Open `demo.atago.yaml`: `record` captured the exit code, the version line on stdout, and an empty stderr — a tool that writes a diagnostic there gets that first line anchored instead — so you have a real test to tighten rather than YAML written from scratch. Swap `git --version` for any command you have (`go version`, `jq --version`, `ls -la`). Then [install atago](/install/) and point it at your own tool.

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

Interactive tools record too: `atago record --pty -- <command>` runs it in a real terminal, lets you drive one session by hand, and writes a `pty:` step that replays your keystrokes as expect/send pairs. It works on Linux, macOS, and Windows (a ConPTY). On POSIX a password prompt (a line-mode read with echo off) becomes an `${env:...}` placeholder automatically; a TUI's raw mode also disables echo, but its keystrokes are recorded literally, so convert a password typed into a TUI's own password field to `${env:...}` by hand. On Windows — where a ConPTY exposes no echo state — every secret send needs the same hand conversion. A `--pty` session is bounded by `--timeout` (default 30s): if the program never exits, atago kills it, writes whatever was captured, and fails instead of hanging forever:

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

`atago run` accepts spec files and directories (searched recursively for `*.atago.yaml`; the `*.atago.yml` spelling is accepted too). Each scenario runs in its own temporary directory, and progress streams as a dot per scenario (`.` pass, `F` fail, `E` error, `s` skip, `f` flaky, `x` XFAIL, `X` XPASS):

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

Some configuration belongs to a *directory* of specs rather than to one file. An `atago.project.yaml` beside (or above) your specs carries it — shared `env:`, shared `defaults:` (a `sandbox_home: true` written once instead of repeated in every file), and `fixtures_dir:`, which exposes a committed corpus to every spec as `${fixtures}`. It is found by walking up to the nearest one, so running a whole tree and re-running a single failing spec inside it resolve the same configuration; a spec file's own values always win, and `atago explain` prints which manifest applied. `${specdir}` is the spec file's own directory, for inputs committed next to it. Both are absolute — a scenario runs in an isolated temp workdir — and both are read-only input: steps write into `${workdir}`.

See [project_manifest](https://github.com/nao1215/atago/blob/main/examples/project_manifest.atago.yaml) and the cookbook recipe for [configuring a directory](/cookbook/#configure-a-whole-directory-of-specs-at-once).

A manifest can also declare the binary under test: `subject:` says how to build it and what specs call it, and `atago run` builds it once before any scenario, putting the artifact first on `PATH` so a spec keeps saying `mytool convert ...` rather than an absolute path. `profiles:` are named build variations selected with `--profile` — a coverage-instrumented build, a race build, a different toolchain. It is language-neutral because a build is just a command (`go build`, `cargo build --release`, `make`); coverage in particular is a profile rather than an atago feature, since instrumenting is an alternate build plus some environment in every language, and merging the raw profiles afterwards is a toolchain job that stays in a script.

See the cookbook recipe for [building the binary under test](/cookbook/#build-the-binary-under-test-before-the-suite-runs).

`expect_fail:` keeps a known bug's reproduction inside the suite instead of exiling it to a directory CI never runs. A scenario that declares it and fails is XFAIL — the run stays green, and the reproduction keeps executing on every commit, so it cannot silently stop working. One that PASSES is XPASS and fails the run: the bug is fixed, and the scenario has to be promoted into the guarded suite (`--allow-xpass` keeps the run green while you do). An execution error is still an error, which is what makes this safe in CI: `expect_fail` says the program gives the wrong answer, not that the spec cannot run. TAP renders both with its `# TODO` directive, JUnit as `<skipped>` / `<failure>`, and GitHub Actions as a notice / an error annotation.

See [expect_fail](https://github.com/nao1215/atago/blob/main/examples/expect_fail.atago.yaml) and the cookbook recipe for [tracking a known bug](/cookbook/#track-a-known-bug-with-an-expected-failure-spec).

`count:`, `min_count:`, and `max_count:` attach to the `contains` or `matches` matcher next to them and say how MANY times it occurs — the question a duplicate-output bug passes when you only ask whether the text is present, and one that used to need `grep -c` in a shell step. `size:`, `min_size:`, and `max_size:` do the same for a file's byte count, compose with the content matchers, and can stand alone: `size: 0` is how a spec pins "the failed run created the file but left it empty rather than half-written", and a `max_size` ceiling is the regression shape of a bundling or compression bug. Occurrences are non-overlapping and a failure names the lines each one landed on; sizes are counted as written, with no CRLF or trailing-newline normalization.

See [count_and_size](https://github.com/nao1215/atago/blob/main/examples/count_and_size.atago.yaml) and the cookbook recipes for [counting](/cookbook/#assert-an-error-is-printed-exactly-once) and [sizing](/cookbook/#assert-a-failed-run-leaves-no-partial-output).

`deterministic: {}` on a `run` step re-runs the command and requires the declared observables to come back byte-identical — the same-input-same-output property that catches iteration order leaking into output (a column order from a map, an unsorted listing, a JSON object whose keys move). Every loose matcher passes such output on every run, so `--repeat` sees no instability; comparing one run's bytes against the next's is the only cheap oracle. A mismatch fails the step with a unified diff between the runs. It is meaningful for an effectively read-only command; when a rerun changes the workdir, the failure says so rather than blaming a bug you do not have.

See [deterministic](https://github.com/nao1215/atago/blob/main/examples/deterministic.atago.yaml) and the cookbook recipe for [proving determinism](/cookbook/#prove-the-same-input-gives-the-same-output).
`suite.env` values may reference variables `suite.setup` captured — a `store` step's value, or the ephemeral address a suite-wide service published through `ready: {store:}` — so a stub registry or proxy can be handed to every scenario as an environment variable without a shell wrapper around `atago run`. A value that cannot resolve is never passed on as the literal text `${name}`: a child process does not fail on that, it uses it, and the resulting error arrives from the tool under test rather than from the spec. A scenario whose env references an undefined name fails before it starts, naming the key, the reference, and the names that are defined.

See [suite_env_from_setup](https://github.com/nao1215/atago/blob/main/examples/suite_env_from_setup.atago.yaml) and the cookbook recipe for [handing a service address to every scenario](/cookbook/#hand-a-suite-wide-services-address-to-every-scenario).

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

`attrs:` on a `screen:` (or `expect_screen:`) assert checks how text is DRAWN, not only what it says: `- {text: "ERROR", fg: red, bold: true}`, `- {text: "README.md", reverse: true}` for the selected row, or `- {text: "ok", fg: default}` — which is how a `--no-color` contract becomes assertable at all. An entry is position-free by default (it passes when at least one occurrence has every one of its cells styled that way), so a styling claim does not break each time the layout shifts; add `row` when the position is the point. Colors are ANSI names, 256-palette indices, or `default`, and `bold: false` is a real claim rather than the absence of one. One terminal rule to know: bold text is drawn in the bright variant of its color, so `fg: red` accepts what a bold `SGR 31` actually puts on screen while `fg: bright-red` stays exact.

`send: {mouse: {row: 5, col: 12}}` clicks and scrolls, for the TUIs that accept a mouse — lazygit, yazi, htop, `fzf --mouse`, anything on bubbletea — and for the times a click beats walking to a pane with twenty keystrokes. `row`/`col` are 1-based screen cells, `action` defaults to `click` (the press and its release in one write, the way a real click arrives), and the buttons include `wheel-up`/`wheel-down` — a wheel notch has no release, so `click` on one sends a single scroll report. Events go out as SGR (1006) reports, and atago refuses the send if the program has not enabled mouse tracking together with SGR encoding, naming the request that never came rather than putting bytes on the wire the program would read as garbage.

`exec: "..."` runs one command on the host while the program under test keeps running, so a session can reach the paths no keystroke can: a git client refreshing after a commit made outside it, a file manager showing a file another process created, a log viewer following a file that grows. It blocks until the command exits, which is the point — after it the change exists, so the `expect_screen` that follows is waiting on the program noticing rather than on a race. The command runs in the scenario workdir with the environment the pty child got, so `sandbox_home` and `clear_env` isolation still holds, and a non-zero exit or timeout ends the step with an error rather than leaving the next wait to blame the program for a change nobody made.

`resize: {rows: 40, cols: 120}` changes the terminal size while the program is running, delivered the way a real terminal delivers it — `SIGWINCH` on POSIX, a ConPTY notification on Windows. Relayout is where full-screen TUIs break (a stale right edge, a panel keeping its old width, a crash on a zero-width column), and without it that path is fixed at whatever the step started with. The rendered screen follows: each part of the transcript is drawn at the size it was produced under, so a frame written before the resize keeps the wrapping it was written with. Settle the screen with `expect_screen` on both sides of a resize — output still in flight when it lands is attributed to the old size, exactly as on a real terminal.

`send: {paste: "..."}` delivers text as a bracketed paste rather than as typing. A REPL or editor that enables the mode treats a pasted block as one unit — it must not run line by line, auto-indent, or fire completion — and that is a different code path from typing the same characters. atago refuses the send if the program has not enabled the mode (`ESC [?2004h`), because the markers would otherwise arrive as literal `[200~` text and surface as a puzzling failure much later; wait for the prompt with an `expect` first, since programs turn the mode on during startup. A tool that draws its interface on stderr and prints its result on stdout is testable as one step — the UI on the rendered screen, the result in the redirected file: [pty_stdout_split](https://github.com/nao1215/atago/blob/main/examples/pty_stdout_split.atago.yaml). `pty` steps and `atago record --pty` run on Linux, macOS, and Windows (where they drive a ConPTY pseudo-console); only `signal:` stays POSIX-only.

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
