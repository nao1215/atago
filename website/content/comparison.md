---
toc: true
title: Comparison
description: How atago compares with Bats-core v1.14.0, ShellSpec 0.28.1, commander v2.5.0, goss v0.4.10, runn v1.9.4, venom v1.3.0, and Microsoft TUI Test 0.0.4, feature by feature, with versions stated and claims sourced from official documentation.
---

This page shows which tool fits which layer of testing. Several of these tools shaped how atago approaches testing.

Releases compared, checked against each project's official documentation and release pages (linked under [Sources](#sources)) on 2026-08-14, and TUI Test on 2026-09-12:

| Tool | Release compared | Released |
|------|------------------|----------|
| [atago](https://github.com/nao1215/atago) | v0.23.0 | 2026-09-21 |
| [Bats-core](https://github.com/bats-core/bats-core) | v1.14.0 | 2026-07-21 |
| [ShellSpec](https://github.com/shellspec/shellspec) | 0.28.1 | 2021-01-11 |
| [commander](https://github.com/commander-cli/commander) | v2.5.0 | 2023-03-28 |
| [goss](https://github.com/goss-org/goss) | v0.4.10 | 2026-07-26 |
| [runn](https://github.com/k1LoW/runn) | v1.9.4 | 2026-06-30 |
| [venom](https://github.com/ovh/venom) | v1.3.0 | 2026-01-06 |
| [TUI Test](https://github.com/microsoft/tui-test) | 0.0.4 | 2026-04-04 |

In the tables, "no" means the compared release's official documentation shows no such feature. If something is wrong, please [open an issue](https://github.com/nao1215/atago/issues).

## Tools that own a different layer

When their layer is the system under test, use these instead of atago:

- [runn](https://github.com/k1LoW/runn) is for scenario-based API testing of an HTTP/gRPC server, with OpenAPI awareness and Go test-helper integration. In atago the CLI is the product, and servers appear only as peers the CLI talks to.
- [venom](https://github.com/ovh/venom) is for platform integration suites, where one test reaches across HTTP, gRPC, Kafka, AMQP, SQL, Redis, and more through its executor catalog.
- [goss](https://github.com/goss-org/goss) validates server state: packages, services, ports, users, and mounts on a host, also served as a health endpoint. Its `command` resource checks exit status and output, but it is built for provisioned machines, not for a CLI you are developing. goss is Linux-first; its README calls macOS and Windows support alpha.

## The same layer: black-box command testing

Bats, ShellSpec, and commander, like atago, run a command and assert on what happened. Columns are atago v0.23.0, Bats-core v1.14.0, ShellSpec 0.28.1, and commander v2.5.0.

### Writing and running tests

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Tests are written in | YAML | Bash | shell DSL (POSIX) | YAML |
| Ships as | single binary | shell scripts (needs Bash) | shell scripts (needs a POSIX shell) | single binary |
| System under test | any executable | anything Bash can drive | shell scripts, functions, and commands | any executable |
| Windows | native, incl. ConPTY terminals | via Bash ports (e.g. Git Bash) | Git Bash, msys2, cygwin | native |
| CI setup on GitHub Actions | [setup-atago](https://github.com/nao1215/setup-atago): prebuilt binary, checksum-verified, Linux/macOS/Windows | [bats-action](https://github.com/bats-core/bats-action): installs Bats and its libraries | no (install script) | no (binary download) |
| Editor completion for the test format | JSON Schema | no | no | no |
| Record a first test from a real run | `atago record`, incl. `--pty` | no | no | `commander add` |
| Generate docs from tests | `atago doc` / `explain` / `list` | no | no | no |

### Assertions

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Exit code | exact, `not:`, `in:` | `run -N`, `run !`, `$status` | `The status should ...` | `exit-code` |
| stdout / stderr as separate streams | always captured separately | `run --separate-stderr` | `The output` / `The error` | `stdout` / `stderr` |
| contains / exact / regex | built in | Bash conditionals (richer via [bats-assert](https://github.com/bats-core/bats-assert)) | `include` / `equal` / `match pattern` | `contains` / `exactly` / `lines` / `line-count` / `not-contains` |
| Output whose line order is not guaranteed | `permutation_of:` / `subset_of:` line relations | via `sort` in the test body | via `sort` in the test body | no |
| JSON output | JSONPath matchers built in | via `jq` | via `jq` | GJSON paths built in |
| YAML output | built in | no | no | no |
| XML output | no | no | no | XPath built in |
| Files the command created | `file:` / `dir:` incl. recursive trees and the executable bit; `changes:` for permission changes | via [bats-file](https://github.com/bats-core/bats-file) | file/path matchers built in | no (`file` compares output against a file) |
| Exact set of files a run touched | `changes:` workdir diff | no | no | no |
| Image / PDF content | format, dimensions, similarity / pages, text | no | no | no |
| Wall-clock duration bounds | `duration:` | no (`--timing` reports only) | no | `timeout` limit only |

### Interactive terminals and snapshots

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| PTY sessions (expect/send, named keys) | built in, Windows ConPTY too | no | no | no |
| Asserts on the rendered screen (TUIs) | `screen:` / `expect_screen:` | no | no | no |
| Snapshot testing with normalization and an update command | `snapshot:` + `atago snapshot update` | no (hand-written `diff`) | no (hand-written compare) | no |

### Environment and test doubles

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Isolated temp workdir per test | automatic, always | opt-in (`$BATS_TEST_TMPDIR`) | opt-in (`$SHELLSPEC_TMPBASE`) | `dir:` config |
| Env isolation | `clear_env` / `pass_env` / `sandbox_home` | manual | manual | `inherit-env` toggle |
| Mock HTTP servers with request asserts | built in | no | no | no |
| Mock/stub shell functions and commands | no | community patterns (e.g. PATH shims) | built in | no |
| Background services with readiness probes | built in | no | no | no |
| Steps on other systems | db / ssh / grpc / browser | no | no | ssh and docker nodes |

### Suite mechanics

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Parameterized tests | `matrix:` | no (per-case, or `bats_test_function`) | `Parameters` | no |
| Generated inputs (property-based) | `forall:` seeded generators with pinned `examples:`, one instance per row | no | no | no |
| Retry / polling | `retry:` re-runs the command until an assert passes | `$BATS_TEST_RETRIES` re-runs the whole test | no | `retries` + `interval` |
| Tags and filtering | `tags:`, `--tag` / `--skip-tag` / `--filter` | `# bats test_tags=`, `--filter-tags` (v1.8+) | `--tag`, focus (`fIt`), patterns | test-name filter |
| Parallel execution | `--parallel`, built in | `--jobs` (needs GNU parallel or rush) | built in | no |
| A known bug tracked in CI | `expect_fail:` (XFAIL, red on XPASS) | no | `Pending` | no |
| Flake detection | `--repeat`, `--retry-failed` reports flaky | no | no | no |
| Coverage of shell scripts | no | no | kcov integration built in | no |
| Report formats | console / JSON / JUnit / GHA / TAP | pretty / TAP / TAP13 / JUnit | documentation / TAP / JUnit / custom | no |

## The terminal layer: TUI applications

[TUI Test](https://github.com/microsoft/tui-test) from Microsoft is the closest tool to atago at driving a real terminal. Columns are atago v0.23.0 and TUI Test 0.0.4, the latest stable release. The [0.1.0 betas](https://github.com/microsoft/tui-test/releases) (beta.3 as of 2026-09-12) are a Rust rewrite with a different design, so read their README too.

| | atago v0.23.0 | TUI Test 0.0.4 |
|---|---|---|
| Tests are written in | YAML | TypeScript / JavaScript |
| Ships as | single binary | npm package (Node 16.6+, or Bun on macOS/Linux) |
| Terminal emulation | [charmbracelet/x/vt](https://github.com/charmbracelet/x), in-process | [xterm.js](https://xtermjs.org/) |
| Isolation | fresh workdir and PTY per scenario | fresh terminal context (terminal + PTY) per test |
| Named keys, typing, submit | `send:` with a named-key vocabulary | `keyPress` and per-key helpers, `submit`, `write` |
| Mouse | clicks, drags, wheel, as SGR 1006 reports | `mouseDown` / `mousePress` / `mouseTo` / `mouseUp` |
| Resize mid-session | `resize:` | `resize()` |
| Assert on visible text | `expect_screen:` / `screen:` | `getByText(...)` with `toBeVisible` |
| Assert on colors and styling | `attrs:`: fg, bg, bold, italic, underline, reverse, blink | `toHaveFgColor` / `toHaveBgColor` |
| Screen snapshots | `snapshot:` with normalization and `atago snapshot update` | `toMatchSnapshot` |
| Retry / flake handling | `expect:` / `expect_screen:` wait until the output or screen matches; `--repeat`, `--retry-failed`, flake reported | `retries`, auto-wait before assertions |
| Parallel execution | `--parallel` | `workers` |
| What a failure leaves behind | `--artifacts-dir` writes the actual and expected payloads | traces record everything the terminal received, replayed with `show-trace` |
| Replaying the session itself | no | traces, via `show-trace` |
| Authoring a first test from a real session | `atago record --pty` writes the spec | no |
| Windows | native, ConPTY | supported (Bun on macOS/Linux only) |
| Shells driven | any executable, shell or not | cmd, Windows PowerShell, PowerShell, bash, git-bash, fish, zsh, xonsh |

TUI Test can replay a failing session from a trace, which atago cannot, and its tests are TypeScript that sits beside the rest of a Node project.

## Other tools you might be weighing

Versions in this section were checked on 2026-08-14.

- [aruba](https://github.com/cucumber/aruba) (v2.4.1) tests CLIs from the Cucumber family, driven from Cucumber-Ruby, RSpec, or Minitest. You write Gherkin scenarios and implement or reuse Ruby step definitions to run them; atago's YAML runs without glue code. A team that already writes Gherkin will be at home in aruba.
- expect is the classic Tcl tool for scripting interactive programs and the ancestor of every expect/send API, atago's `pty:` included. It is fine for one-off automation; a spec runner adds assertions, isolation, and reporting.
- [cram](https://github.com/aiiie/cram) and its fork [prysk](https://github.com/prysk/prysk) (0.20.0) snapshot-test command lines in `.t` files and are the closest relatives of atago's `snapshot:`. They pin whole session transcripts, while atago keeps snapshots apart from structured asserts and normalizes volatile output.
- [testscript](https://github.com/rogpeppe/go-internal) (v1.15.0) is the txtar-based script runner extracted from `cmd/go`'s own tests and the usual way to test Go CLIs inside `go test`. It runs in-process with your Go tests; atago runs any binary from outside, in any language.
- [shUnit2](https://github.com/kward/shunit2) (v2.1.8) is xUnit for shell scripts, in the same family as Bats and ShellSpec. The [migration guide](/migrate/)'s mappings apply to it the same way.
- [trycmd](https://github.com/assert-rs/trycmd) (v0.15.2) snapshot-tests CLIs from Markdown files inside Rust's `cargo test`, roughly the Rust counterpart to testscript.
- General-purpose frameworks (Cucumber itself, Jest, Vitest, pytest) test code in whatever language your CLI is written in. They can drive a CLI through a subprocess helper that you write and maintain. Jest's snapshot testing is prior art for atago's `snapshot:`.

## Where the others are the better choice

- Bats is the standard for testing Bash code, with over a decade of history, active releases, and an ecosystem (bats-assert, bats-file, bats-mock, bats-detik) that atago has no equivalent of. When your tests are shell code that sources scripts and calls functions, Bats runs them in-process, which a black-box runner cannot. atago tests the `bats` CLI itself from the outside: its TAP output, selection flags, formatters, and setup/teardown are pinned at [/real-world/bats/](/real-world/bats/).
- ShellSpec is the most featureful shell unit-testing framework, with function and command mocking, parameterized examples, built-in parallel runs, and kcov coverage. When shell scripts are the product, it is the stronger tool. Its released feature set has been stable since 2021.
- commander takes the same YAML approach in a smaller package and has XML assertions and a docker execution node, which atago lacks.
- goss, runn, and venom own their layers, as described [above](#tools-that-own-a-different-layer).
- TUI Test replays a failing terminal session from a trace, which atago cannot do. See [the terminal layer](#the-terminal-layer-tui-applications).

atago is the youngest project on this page and still pre-1.0. The spec format is versioned and every feature ships with a runnable, CI-tested example, but the other tools have years more production use.

The [migration guide](/migrate/) maps Bats and ShellSpec constructs to atago one-to-one, and every mapping on it runs in CI, originals and migrations both.
## Sources

Feature claims come from each project's official documentation for the
release stated above:

- Bats-core: [writing tests](https://bats-core.readthedocs.io/en/stable/writing-tests.html), [usage](https://bats-core.readthedocs.io/en/stable/usage.html), [releases](https://github.com/bats-core/bats-core/releases)
- ShellSpec: [shellspec.info](https://shellspec.info/), [README](https://github.com/shellspec/shellspec), [releases](https://github.com/shellspec/shellspec/releases)
- commander: [README](https://github.com/commander-cli/commander), [releases](https://github.com/commander-cli/commander/releases)
- goss: [README](https://github.com/goss-org/goss), [releases](https://github.com/goss-org/goss/releases)
- runn: [README](https://github.com/k1LoW/runn), [releases](https://github.com/k1LoW/runn/releases)
- venom: [README](https://github.com/ovh/venom), [releases](https://github.com/ovh/venom/releases)
- TUI Test: [README at 0.0.4](https://github.com/microsoft/tui-test/blob/0.0.4/README.md), [beta README](https://github.com/microsoft/tui-test/blob/main/README.md), [releases](https://github.com/microsoft/tui-test/releases)
- atago: the [reference](/reference/) and the [runnable examples](/cookbook/#every-feature-has-a-runnable-spec) behind every claim
