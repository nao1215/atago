---
toc: true
title: Reference
description: atago subcommands, scenario selection flags, snapshot updating and scrub rules, the JSON Schema for editors, the skill for coding agents, shell completion, and the exit code contract.
---

## Subcommands

| Command | Does |
|---------|------|
| `atago run` | run specs and report results |
| `atago record` | run a command once and write a spec from what it observed (`--pty` for interactive sessions) |
| `atago init` | scaffold a spec (`--template` for browser, cli, db, grpc, http, mock, services, ssh; `cli` is the default) |
| `atago snapshot update` | record or refresh golden files |
| `atago explain` | describe what a spec does without running it, or what an `ATG` code means (`atago explain ATG2201`) |
| `atago doc` | generate Markdown from specs, with fixtures and golden files inlined |
| `atago manifest` | emit a stable JSON summary of specs for tooling |
| `atago list` | show scenarios, tags, and artifacts |
| `atago completion` | print a shell completion script |
| `atago version` | print the atago version |
| `atago help` | print usage (`atago <command> --help` for one command's flags) |

`explain`, `doc`, `manifest`, and `list` load and validate the spec first and exit 2 on a schema error, so any of them works as a lint step in CI.

## Selecting scenarios

| Flag | Selects or changes |
|------|--------------------|
| `--filter NAME` | scenarios whose name contains `NAME`; repeatable, and comma-separated for OR (`--filter a,b` equals `--filter a --filter b`) |
| `--tag T` / `--skip-tag T` | scenarios with or without tag `T`, matched exactly; comma-separated and repeatable (`atago list` shows the tags) |
| `--parallel N` | how many scenarios run at once |
| `--fail-fast` | stop scheduling new scenarios after the first outcome that fails the run |
| `--rerun-failed` | only the scenarios the previous run recorded as failed in `.atago/last-failed.json` |
| `--verbose` | trace every command, capture, and assertion verdict, for passing scenarios too |

Under `--ci`, a selection that matches no scenario fails the run (exit 3). Flags may go before, after, or between the paths (`atago run ./specs --report json` and `atago run --report json ./specs` are the same command), and `--` ends flag parsing so a path spelled like a flag can still be named. In `atago record`, everything after `--` is the recorded program's own command line.

## Snapshot testing

`snapshot` matchers compare output against committed golden files after normalizing ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF. Record or refresh them with:

```shell
atago snapshot update spec.atago.yaml
```

For other volatile values (auto-increment IDs, request identifiers, epoch times), spec-wide `scrub:` rules rewrite each regex match to a placeholder before the compare, after `secrets:` masking:

```yaml
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
```

See the [scrub example](https://github.com/nao1215/atago/blob/main/examples/scrub.atago.yaml).

## Spec file keys

Every key a spec file accepts, from the committed [JSON Schema](https://github.com/nao1215/atago/blob/main/schema/atago.schema.json) that also drives editor completion. Indentation shows nesting, and a linked type (like [step](#spec-step)) has its own section.

All keys belong to spec format version `1`, the only one so far, so every spec starts with `version: "1"`. Since is the atago release that introduced the key (`unreleased` means merged to main but not yet tagged).

{{< spec-reference >}}

## Directory manifest

An `atago.project.yaml` holds settings for a whole directory of specs:

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
      cwd: ".."
    env:
      GOCOVERDIR: "${env:GOCOVERDIR}"
```

`atago run` builds it once per invocation, before any scenario, and prepends the artifact's directory to `PATH`. `--profile NAME` swaps in that profile's whole `build` block, `cwd` included, and layers its `env`. A failing build, or one that exits 0 without writing `${artifact}`, is a run-level error and no scenario executes.

atago uses the nearest manifest walking up from a spec, so `atago run ./e2e` and `atago run ./e2e/one.atago.yaml` resolve the same configuration. Precedence is host < project < suite < scenario < step for `env`, and a spec file's own `defaults:` beat the manifest's. `fixtures_dir` resolves against the manifest's directory and must exist at load time. `atago explain` prints the manifest that applied and the resolved fixtures directory. Its own schema is [atago.project.schema.json](https://github.com/nao1215/atago/blob/main/schema/atago.project.schema.json).

| Variable | Is |
|----------|----|
| `${workdir}` | this scenario's isolated temp directory, the only one it owns |
| `${suitedir}` | the suite's scratch directory, shared by `suite.setup` and every scenario |
| `${specdir}` | the directory holding the spec file (read-only input) |
| `${fixtures}` | the manifest's `fixtures_dir` (read-only input); unset when no manifest declares one |

All four are absolute, because a scenario does not run where its spec lives.

## Editor support (JSON Schema)

With the YAML language server, [schema/atago.schema.json](https://github.com/nao1215/atago/blob/main/schema/atago.schema.json) gives completion and validation for step types, every matcher, and the `${workdir}` / `${env:NAME}` / `${name}` / `$${...}` expansion rules. `atago init` and `atago record` write this header as the first line of every spec they generate. For an existing spec, add it with the absolute URL, which resolves in any project:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
```

The report and manifest outputs have schemas too: [report.schema.json](https://github.com/nao1215/atago/blob/main/schema/report.schema.json) and [manifest.schema.json](https://github.com/nao1215/atago/blob/main/schema/manifest.schema.json).

## Coding agents

[skills/atago](https://github.com/nao1215/atago/tree/main/skills/atago) is an Agent Skill: a `SKILL.md` with short references for a minimal spec, the common matchers, and reading a failure. Its rule is the loop: write a spec, run `atago run`, fix until it passes. Its specs and commands run in CI.

- Claude Code: copy `skills/atago` to `~/.claude/skills/atago`.
- Codex: copy `skills/atago` to `~/.codex/skills/atago`.

## Platform support

atago runs on Linux, macOS, and Windows. CI runs the unit suite on every OS, the self-hosted E2E suite on Linux and macOS, and on Windows both under the native `cmd.exe` and under a POSIX shell. The differences are listed below.

`skip:` and `only:` gate on the host as Go names it: `linux`, `darwin`, `windows`, `freebsd`, `openbsd`, and `netbsd`. The BSDs are separate values because, for example, FreeBSD `ps` and OpenBSD `ps` print different things. `bsd` is not a value, and neither is `dragonfly`, where atago does not build. atago ships no binary for the three BSDs; `go install` builds one.

| Behavior | Linux / macOS | Windows |
|----------|---------------|---------|
| `shell: true` | `/bin/sh -c`, resolved absolutely so the program under test cannot supply it | `%SystemRoot%\System32\cmd.exe /S /C`, resolved the same way. `ATAGO_SHELL` overrides on both |
| `signal:` steps | delivers `TERM`, `INT`, `HUP`, `USR1`, `USR2`, `KILL` to the service's process group | not supported, since Windows has no POSIX signals. Gate with `skip: {os: windows}` |
| cancel / timeout teardown | kills the whole process group | kills the whole process tree (`taskkill /T`) |
| `pty:` steps and `atago record --pty` | a real pty | a ConPTY, which needs Windows 10 version 1809 or later. `record --pty` cannot auto-detect a password prompt there, because a ConPTY exposes no echo state, so convert a secret send to `${env:...}` by hand |
| `pty: {graphics: kitty}` | the same pty | a ConPTY hosted by the Windows Terminal console host (OpenConsole, from Microsoft's MIT-licensed ConPTY package) instead of the one built into Windows, which drops the kitty graphics protocol on its way through. atago embeds it for amd64 and arm64 and writes it to `%LOCALAPPDATA%\atago` on first use; a build for any other architecture reports the step as unsupported on the platform |
| `file: {executable: ...}` | the mode bits | the file extension against `PATHEXT`, which is what Windows uses to decide what it runs by name. There is no execute bit to read |
| `fixture: {mode: ...}` | sets the permission bits | no effect, since Windows has no POSIX permissions |
| `fixture: {symlink: ...}` | always available | needs Developer Mode or an elevated process |
| `changes:` | compares content, symlink target, kind, and permission bits | compares content, symlink target, and kind. Permissions are left out: Go synthesizes a mode from the read-only attribute, so including it would make one spec report a different delta per OS |
| `sandbox_home: true` | redirects `HOME` and the XDG base directories | redirects `USERPROFILE`, `APPDATA`, `LOCALAPPDATA`, `HOMEDRIVE`, `HOMEPATH` |
| `clear_env: true` | starts from an empty environment plus `pass_env` | the same, plus `SystemRoot`, `SystemDrive`, `TEMP`, `TMP`, and `PATHEXT`, without which a process cannot start at all |

### How a command is split into argv

Without `shell: true`, atago splits `run.command` itself. On every OS, both quote kinds group an argument and are stripped, and `\"` inside a double-quoted group is a literal quote. Other backslashes differ. On Windows they stay literal, because they are path separators, so `"C:\dir\x"`, `"\\server\share"` and `"a\\b"` arrive unchanged. On Linux and macOS the string goes through [go-shellwords](https://github.com/mattn/go-shellwords), whose escapes are C-style, so `\t` becomes a tab and `C:\dir\x` loses its backslashes. Write an argument with quotes or backslashes in single quotes, where nothing is an escape.

On Windows, a double-quoted path ending in a backslash (`"C:\dir\"`) turns its closing quote into a literal one, so the argument never closes and atago reports `ATG4001`. Write that path in single quotes.

### Choosing the shell

`shell: true` runs the platform's own interpreter, so a command written for `/bin/sh` does not run under `cmd.exe`. Either keep the command portable (`echo` and `exit` are builtins of both, and `run.env:`, `run.stdin:`, `run.stdout_to:` replace variable prefixes and redirects), or point atago at the shell you want:

```shell
ATAGO_SHELL='C:\Program Files\Git\bin\bash.exe' atago run ./e2e
```

`ATAGO_SHELL` takes an absolute path, or a name found on `PATH`, on either platform. atago picks the calling convention from the name: `/S /C` for `cmd.exe`, `-c` for anything else, which covers the bash that ships with Git for Windows and MSYS2 as well as PowerShell. With a POSIX shell on Windows, `${workdir}`, `${specdir}`, and `${atago}` expand to backslash paths that the shell reads as escapes, so interpolate them into argv-form commands (`shell: false`) instead of shell commands.

## Shell completion

`atago completion <bash|zsh|fish|powershell>` prints a completion script for your shell.

## Exit codes

| Code | Meaning                   |
|------|---------------------------|
| `0`  | no unexpected failures: every scenario passed, was skipped, or was an XFAIL (an `expect_fail:` scenario that failed as declared) |
| `1`  | one or more failed (including an XPASS: an `expect_fail:` scenario that passed, unless `--allow-xpass`) |
| `2`  | spec error (YAML syntax or schema/semantic validation) |
| `3`  | CLI-invocation error (unknown subcommand, bad flag, or no matching spec files) |
| `4`  | execution error           |
| `5`  | internal error            |
| `6`  | security policy violation |

`Ctrl-C`/`SIGTERM` stops the run cleanly: in-flight processes, services, and sessions are torn down, partial results are reported, and the run exits `4`.

Errors also carry a diagnostic code such as `ATG2201`, whose first digit is the exit code, so `ATG2xxx` always exits 2. Codes stay the same when a message is reworded. [Error codes](/errors/) lists what each one means, what to change, and which families carry codes today. Assertion failures (exit 1) carry no code. `atago explain ATG2201` prints the same entry in the terminal, and the JSON report has a `code` field on each failure for grouping by cause.
