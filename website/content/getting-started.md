---
toc: true
title: Getting started
description: Record a real run of your CLI, read the spec atago wrote, and keep it as a test. Then assert on files, snapshots, interactive prompts, and server peers.
---

## First run

You can try the loop before installing anything. With Go on the machine, this records a real run and replays it as a test:

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped (4ms)
```

`record` printed `recorded: exit 0, 1 line of stdout, 0 files created` and wrote this spec:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"

# Recorded by `atago record` — a starting point, not a verdict:
# tighten the matchers to pin the behavior you actually care about.
suite:
  name: git

scenarios:
  - name: git --version # TODO: describe the behavior
    steps:
      - run:
          command: git --version
      - assert:
          exit_code: 0
      - assert:
          stdout:
            contains: git version 2.53.0 # first non-empty line, trimmed
      - assert:
          stderr:
            empty: true
```

A tool that writes a diagnostic to stderr gets that first line anchored instead of `empty: true`. Swap `git --version` for any command you have, then [install atago](/install/) and point it at your own tool.

## Start from a real run

`atago record -- <command>` runs your command once and writes a spec from what it observed (exit code, output, created files):

```shell
$ atago record --out mytool.atago.yaml -- mytool convert input.txt
recorded: exit 0, 2 lines of stdout, 1 file created
wrote mytool.atago.yaml
$ atago run mytool.atago.yaml
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped (12ms)
```

`atago record --pty -- <command>` records an interactive tool: you drive one session by hand in a real terminal, and atago writes a `pty:` step that replays your keystrokes as expect/send pairs. It works on Linux, macOS, and Windows (a ConPTY). On POSIX, a password prompt (a line-mode read with echo off) becomes an `${env:...}` placeholder automatically. A TUI's own password field is recorded literally, and on Windows, where a ConPTY exposes no echo state, so is every secret, so convert those sends to `${env:...}` by hand. `--timeout` (default 30s) bounds the session: if the program never exits, atago kills it, writes what it captured, and fails.

```shell
$ atago record --pty --out wizard.atago.yaml -- mytool init
```

`atago init` scaffolds a blank template instead.

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

`atago run` takes spec files and directories (searched recursively for `*.atago.yaml`; `*.atago.yml` works too). Each scenario runs in its own temporary directory, and progress prints one character per scenario (`.` pass, `F` fail, `E` error, `s` skip, `f` flaky, `x` XFAIL, `X` XPASS):

```shell
$ atago run ./specs
............

PASSED  12 scenarios: 12 passed, 0 failed, 0 errored, 0 skipped (1.4s)
```

Scenarios run concurrently, four per CPU by default (`--parallel N`; `--parallel 1` serializes). Workdirs are isolated but the host network is shared, so two scenarios that each start a background `service:` need distinct ports.

A failed check prints what was expected and what happened, with a unified diff for multi-line output:

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

## 2. Check generated files and snapshots

`fixture:` writes input files into the workdir, from `content:` (inline text), `base64:` (inline bytes), `from:` (a copy of an existing file), or `symlink:`. `file:` and `dir:` assertions check what the command produced, and `snapshot:` compares output with a committed golden file after normalizing temp paths, UUIDs, and timestamps:

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

Examples: [files_and_fixtures](https://github.com/nao1215/atago/blob/main/examples/files_and_fixtures.atago.yaml), [snapshot](https://github.com/nao1215/atago/blob/main/examples/snapshot.atago.yaml), and [dir_tree](https://github.com/nao1215/atago/blob/main/examples/dir_tree.atago.yaml) for whole-tree golden manifests.

### Directory-wide settings

An `atago.project.yaml` beside or above your specs holds settings for the whole directory: shared `env:`, shared `defaults:` (for example `sandbox_home: true` once instead of in every file), and `fixtures_dir:`, which every spec sees as `${fixtures}`. atago uses the nearest one walking up, so running the whole tree and re-running one spec inside it get the same configuration. A spec's own values win, and `atago explain` prints which manifest applied. `${specdir}` is the spec file's directory. Both paths are absolute and meant as input; atago does not make them read-only, so steps should write into `${workdir}`. See [project_manifest](https://github.com/nao1215/atago/blob/main/examples/project_manifest.atago.yaml) and [configuring a directory](/cookbook/#configure-a-whole-directory-of-specs-at-once).

The manifest can also build the binary under test. `subject:` says how to build it, and `atago run` builds it once before any scenario and puts it first on `PATH`, so specs keep calling `mytool convert ...`. `profiles:`, selected with `--profile`, are build variations such as a coverage build, a race build, or another toolchain. A build is any command (`go build`, `cargo build --release`, `make`), so coverage is a profile; merging the raw coverage files afterwards stays in a script. See [building the binary under test](/cookbook/#build-the-binary-under-test-before-the-suite-runs).

### Known bugs, counts, sizes, and determinism

`expect_fail:` keeps a known bug's reproduction in the suite. A scenario that declares it and fails is XFAIL and the run stays green. If it passes, it is XPASS and fails the run, because the bug is fixed and the scenario should move into the guarded suite (`--allow-xpass` keeps the run green meanwhile). An execution error is still an error. TAP renders both with `# TODO`, JUnit as `<skipped>` / `<failure>`, and GitHub Actions as a notice / an error annotation. See [expect_fail](https://github.com/nao1215/atago/blob/main/examples/expect_fail.atago.yaml) and [tracking a known bug](/cookbook/#track-a-known-bug-with-an-expected-failure-spec).

`count:`, `min_count:`, and `max_count:` say how many times the neighboring `contains` or `matches` matcher occurs, which catches duplicated output without `grep -c`. `size:`, `min_size:`, and `max_size:` check a file's byte count and also work alone: `size: 0` pins a failed run that left an empty file, and `max_size` catches a bundle that grew. Occurrences do not overlap, a failure names the lines each one landed on, and sizes are counted as written, with no CRLF or trailing-newline normalization. See [count_and_size](https://github.com/nao1215/atago/blob/main/examples/count_and_size.atago.yaml), [counting](/cookbook/#assert-an-error-is-printed-exactly-once), and [sizing](/cookbook/#assert-a-failed-run-leaves-no-partial-output).

`deterministic: {}` on a `run` step re-runs the command and requires the declared outputs to be byte-identical. That catches iteration order leaking into output (map column order, an unsorted listing, JSON keys that move), which loose matchers and `--repeat` never notice. A mismatch fails with a diff between the runs. Use it on commands that do not change the workdir; if a rerun does, the failure says so. See [deterministic](https://github.com/nao1215/atago/blob/main/examples/deterministic.atago.yaml) and [proving determinism](/cookbook/#prove-the-same-input-gives-the-same-output).

`suite.env` values can reference what `suite.setup` captured, such as a `store` step's value or the address a suite-wide service published through `ready: {store:}`. That hands a stub registry or proxy to every scenario without wrapping `atago run` in a shell script. An unresolved `${name}` is never passed to the child as literal text: the scenario fails before it starts, naming the key, the reference, and the names that are defined. See [suite_env_from_setup](https://github.com/nao1215/atago/blob/main/examples/suite_env_from_setup.atago.yaml) and [handing a service address to every scenario](/cookbook/#hand-a-suite-wide-services-address-to-every-scenario).

## 3. Drive interactive prompts and TUIs

A `pty` step runs the command in a real pseudo-terminal and drives it with an expect/send session, for wizards, REPLs, and TTY-detection branches:

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

For full-screen TUIs, `expect_screen:` waits on the live rendered frame during the session, and `screen:` asserts the final frame after exit:

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

The session can also send and check the following. Examples: [pty](https://github.com/nao1215/atago/blob/main/examples/pty.atago.yaml), [pty_screen](https://github.com/nao1215/atago/blob/main/examples/pty_screen.atago.yaml), and the cross-platform [pty_portable](https://github.com/nao1215/atago/blob/main/examples/pty_portable.atago.yaml).

- Named keys, `send: {key: enter}`: control-byte aliases (`ctrl-space`, `ctrl-[`, `ctrl-_`), `ctrl-hyphen`/`ctrl-minus` for apps that read the physical `Ctrl+-` key, `shift-tab` (alias `backtab`), `alt-a`..`alt-z`, `alt-enter`, `alt-backspace`, and arrows with `ctrl-`/`shift-` (`ctrl-left`, `shift-up`, …). A misspelled key fails at load time and prints the accepted list. `alt-backspace` ends in DEL, so a raw-mode TUI sees the chord while a cooked line prompt sees an erase.
- `attrs:` on `screen:` or `expect_screen:` checks how text is drawn: `- {text: "ERROR", fg: red, bold: true}`, `- {text: "README.md", reverse: true}` for a selected row, or `- {text: "ok", fg: default}` for a `--no-color` contract. An entry passes when at least one occurrence has every cell styled that way; add `row` to pin the position. Colors are ANSI names, 256-palette indices, or `default`, and `bold: false` is checked. Bold text is drawn in the bright variant of its color, so `fg: red` accepts a bold `SGR 31` while `fg: bright-red` stays exact.
- `send: {mouse: {row: 5, col: 12}}` clicks and scrolls in TUIs that accept a mouse (lazygit, yazi, htop, `fzf --mouse`, bubbletea apps). `row`/`col` are 1-based, `action` defaults to `click` (press and release in one write), and `wheel-up`/`wheel-down` send a single scroll report. Events are SGR (1006) reports, and atago refuses the send if the program has not enabled mouse tracking with SGR encoding.
- `exec: "..."` runs a host command while the program keeps running, for example a commit made outside a git client or a file another process creates. It blocks until the command exits, so the next `expect_screen` waits on the program noticing the change. It runs in the scenario workdir with the pty child's environment (`sandbox_home` and `clear_env` still apply), and a non-zero exit or timeout ends the step with an error.
- `resize: {rows: 40, cols: 120}` resizes the terminal (`SIGWINCH` on POSIX, a ConPTY notification on Windows) to test relayout. Each part of the transcript is rendered at the size it was produced under, so settle the screen with `expect_screen` before and after the resize.
- `send: {paste: "..."}` sends a bracketed paste, which REPLs and editors handle differently from typing. atago refuses the send if the program has not enabled the mode (`ESC [?2004h`), so `expect` the prompt first.

A tool that draws its UI on stderr and prints its result on stdout fits in one step: the UI on the rendered screen, the result in the redirected file ([pty_stdout_split](https://github.com/nao1215/atago/blob/main/examples/pty_stdout_split.atago.yaml)). `pty` steps and `atago record --pty` run on Linux, macOS, and Windows (ConPTY); only `signal:` is POSIX-only.

ConPTY does not report whether it echoes input. If output after a `send` starts with the sent text, atago treats that first copy as possible echo even when it arrives late, so write the next `expect` against the program's response.

## When your CLI talks to a server

The same YAML drives HTTP, database, SSH, gRPC, headless-browser, and offline mock-server peers as dependencies of the CLI under test. `atago init --template <name>` scaffolds each:

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

Scaffold one with: atago init --template <name>
```

## Next

The [cookbook](/cookbook/) has a copyable spec for most jobs and links the runnable example for every feature. [Use it in CI](/ci/) wires a suite into GitHub Actions.
