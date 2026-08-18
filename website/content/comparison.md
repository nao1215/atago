---
toc: true
title: Comparison
description: How atago compares with Bats-core v1.14.0, ShellSpec 0.28.1, commander v2.5.0, goss v0.4.10, runn v1.9.4, and venom v1.3.0 — feature by feature, versions stated, sourced from official documentation.
---

Every tool here is good software, written by people who understood their
problem well; several of them shaped how atago thinks about testing. This page
exists to answer one practical question — *which tool owns which layer* — not
to rank projects.

Compared releases, checked on 2026-08-14 against each project's official
documentation and release pages (linked under [Sources](#sources)):

| Tool | Release compared | Released |
|------|------------------|----------|
| [atago](https://github.com/nao1215/atago) | v0.19.0 | 2026 |
| [Bats-core](https://github.com/bats-core/bats-core) | v1.14.0 | 2026-07-21 |
| [ShellSpec](https://github.com/shellspec/shellspec) | 0.28.1 | 2021-01-11 |
| [commander](https://github.com/commander-cli/commander) | v2.5.0 | 2023-03-28 |
| [goss](https://github.com/goss-org/goss) | v0.4.10 | 2026-07-26 |
| [runn](https://github.com/k1LoW/runn) | v1.9.4 | 2026-06-30 |
| [venom](https://github.com/ovh/venom) | v1.3.0 | 2026-01-06 |

In the tables below, — means the compared release does not provide the
feature itself, as far as its official documentation shows. If we got
something wrong, please
[open an issue](https://github.com/nao1215/atago/issues) — this page is meant
to stay accurate, not flattering.

## Tools that own a different layer

atago does not compete with these, and where their layer is the system under
test, they are the better choice:

- [runn](https://github.com/k1LoW/runn) owns scenario-based API
  testing. If the system under test is an HTTP/gRPC server, runn's
  runners, OpenAPI awareness, and Go test-helper integration are built for
  exactly that. atago points the other way: the CLI is the product, and
  servers appear only as peers the CLI talks to.
- [venom](https://github.com/ovh/venom) owns platform integration
  suites — one test reaching across HTTP, gRPC, Kafka, AMQP, SQL, Redis,
  and more through its executor catalog.
- [goss](https://github.com/goss-org/goss) owns server state
  validation: asserting that packages, services, ports, users, and mounts
  on a host are what they should be, and serving that as a health endpoint.
  Its `command` resource does check exit status and output, but validating a
  provisioned machine is a different job from testing a CLI you are
  developing. (goss is Linux-first; macOS and Windows support is alpha, per
  its README.)

## The same layer: black-box command testing

Bats, ShellSpec, and commander overlap with atago directly: all four run a
command and assert on what happened. The tables compare that shared job.
Columns are atago v0.19.0, Bats-core v1.14.0, ShellSpec 0.28.1, and
commander v2.5.0.

### Writing and running tests

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Tests are written in | YAML | Bash | shell DSL (POSIX) | YAML |
| Ships as | single binary | shell scripts (needs Bash) | shell scripts (needs a POSIX shell) | single binary |
| System under test | any executable | anything Bash can drive | shell scripts, functions, and commands | any executable |
| Windows | native, incl. ConPTY terminals | via Bash ports (e.g. Git Bash) | Git Bash, msys2, cygwin | native |
| CI setup on GitHub Actions | [setup-atago](https://github.com/nao1215/setup-atago): prebuilt binary, checksum-verified, Linux/macOS/Windows | [bats-action](https://github.com/bats-core/bats-action): installs Bats and its libraries | — (install script) | — (binary download) |
| Editor completion for the test format | JSON Schema | — | — | — |
| Record a first test from a real run | `atago record`, incl. `--pty` | — | — | `commander add` |
| Generate docs from tests | `atago doc` / `explain` / `list` | — | — | — |

### Assertions

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Exit code | exact, `not:`, `in:` | `run -N`, `run !`, `$status` | `The status should ...` | `exit-code` |
| stdout / stderr as separate streams | always captured separately | `run --separate-stderr` | `The output` / `The error` | `stdout` / `stderr` |
| contains / exact / regex | built in | Bash conditionals (richer via [bats-assert](https://github.com/bats-core/bats-assert)) | `include` / `equal` / `match pattern` | `contains` / `exactly` / `lines` / `line-count` / `not-contains` |
| JSON output | JSONPath matchers built in | via `jq` | via `jq` | GJSON paths built in |
| YAML output | built in | — | — | — |
| XML output | — | — | — | XPath built in |
| Files the command created | `file:` / `dir:` incl. recursive trees, permissions | via [bats-file](https://github.com/bats-core/bats-file) | file/path matchers built in | — (`file` compares output against a file) |
| Exact set of files a run touched | `changes:` workdir diff | — | — | — |
| Image / PDF content | format, dimensions, similarity / pages, text | — | — | — |
| Wall-clock duration bounds | `duration:` | — (`--timing` reports only) | — | `timeout` limit only |

### Interactive terminals and snapshots

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| PTY sessions (expect/send, named keys) | built in, Windows ConPTY too | — | — | — |
| Asserts on the rendered screen (TUIs) | `screen:` / `expect_screen:` | — | — | — |
| Snapshot testing with normalization and an update command | `snapshot:` + `atago snapshot update` | — (hand-written `diff`) | — (hand-written compare) | — |

### Environment and test doubles

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Isolated temp workdir per test | automatic, always | opt-in (`$BATS_TEST_TMPDIR`) | opt-in (`$SHELLSPEC_TMPBASE`) | `dir:` config |
| Env isolation | `clear_env` / `pass_env` / `sandbox_home` | manual | manual | `inherit-env` toggle |
| Mock HTTP servers with request asserts | built in | — | — | — |
| Mock/stub shell functions and commands | — | community patterns (e.g. PATH shims) | built in | — |
| Background services with readiness probes | built in | — | — | — |
| Steps on other systems | db / ssh / grpc / browser | — | — | ssh and docker nodes |

### Suite mechanics

| | atago | Bats | ShellSpec | commander |
|---|---|---|---|---|
| Parameterized tests | `matrix:` | — (per-case, or `bats_test_function`) | `Parameters` | — |
| Retry / polling | `retry:` re-runs the command until an assert passes | `$BATS_TEST_RETRIES` re-runs the whole test | — | `retries` + `interval` |
| Tags and filtering | `tags:`, `--tag` / `--skip-tag` / `--filter` | `# bats test_tags=`, `--filter-tags` (v1.8+) | `--tag`, focus (`fIt`), patterns | test-name filter |
| Parallel execution | `--parallel`, built in | `--jobs` (needs GNU parallel or rush) | built in | — |
| A known bug tracked in CI | `expect_fail:` (XFAIL, red on XPASS) | — | `Pending` | — |
| Flake detection | `--repeat`, `--retry-failed` reports flaky | — | — | — |
| Coverage of shell scripts | — | — | kcov integration built in | — |
| Report formats | console / JSON / JUnit / GHA / TAP | pretty / TAP / TAP13 / JUnit | documentation / TAP / JUnit / custom | — |

## Other tools you might be weighing

Adjacent tools that come up in the same search, each with an honest placement
(versions in this section checked on 2026-08-14):

- [aruba](https://github.com/cucumber/aruba) (v2.4.1) — CLI testing from the
  Cucumber family, driven from Cucumber-Ruby, RSpec, or Minitest. Same layer
  as atago, different axis: Gherkin scenarios read well, and you implement or
  reuse Ruby step definitions to make them run, where atago's YAML executes
  without glue code. If your team already writes Gherkin, aruba is the
  natural home.
- expect — the classic Tcl tool for scripting interactive programs, and the
  ancestor of every expect/send API including atago's `pty:`. Still fine for
  one-off automation; a spec runner adds assertions, isolation, and reporting
  around the same idea.
- [cram](https://github.com/aiiie/cram) and its fork
  [prysk](https://github.com/prysk/prysk) (0.20.0) — snapshot testing for
  command lines in `.t` files, the closest relatives of atago's `snapshot:`.
  They pin whole session transcripts; atago separates snapshots from
  structured asserts and normalizes volatile output.
- [testscript](https://github.com/rogpeppe/go-internal) (v1.15.0) — the
  txtar-based script runner extracted from `cmd/go`'s own tests, and the Go
  ecosystem's default for testing CLIs inside `go test`. Running in-process
  with your Go tests is its strength; atago runs any binary from outside, in
  any language.
- [shUnit2](https://github.com/kward/shunit2) (v2.1.8) — xUnit for shell
  scripts, in the same family as Bats and ShellSpec; the
  [migration guide](/migrate/)'s mappings apply the same way.
- [trycmd](https://github.com/assert-rs/trycmd) (v0.15.2) — snapshot-tests
  CLIs from Markdown files inside Rust's `cargo test`; roughly the Rust
  counterpart to testscript.
- General-purpose test frameworks — Cucumber itself, Jest, Vitest, pytest,
  and their peers own code-level testing for whatever language your CLI is
  written in. They can drive a CLI through a subprocess helper, but that
  helper is code you write and maintain. Jest's snapshot testing in
  particular is prior art for atago's `snapshot:`.

## Where the others are the better choice

An honest table needs the reverse direction spelled out:

- Bats is the standard for testing Bash code, with over a decade of
  history, active releases, and an ecosystem (bats-assert, bats-file,
  bats-mock, bats-detik) that atago has no equivalent of. If your tests *are*
  shell code — sourcing scripts, calling functions — Bats runs them
  in-process; a black-box runner cannot. atago does test the `bats` CLI
  itself from the outside — its TAP output, selection flags, formatters, and
  setup/teardown are pinned at [/real-world/bats/](/real-world/bats/).
- ShellSpec is the most featureful shell unit-testing framework:
  function and command mocking, parameterized examples, built-in parallel
  runs, and kcov coverage. When shell scripts are the product, it is the
  stronger tool; its released feature set has been stable since 2021.
- commander shares atago's YAML approach in a smaller package, and has
  XML assertions and a first-class docker execution node, which atago lacks.
- goss / runn / venom own their layers outright, as described
  [above](#tools-that-own-a-different-layer).

atago is also the youngest project on this page and still pre-1.0. The spec
format is versioned and every feature ships with a runnable, CI-tested
example, but the others have years more production mileage — that counts for
something, and pretending otherwise would defeat the point of this page.

If the table convinced you to try a migration, the
[migration guide](/migrate/) maps Bats and ShellSpec constructs to atago
one-to-one — and every mapping on that page runs in CI, originals and
migrations both.

## Sources

Feature claims come from each project's official documentation for the
release stated above:

- Bats-core: [writing tests](https://bats-core.readthedocs.io/en/stable/writing-tests.html), [usage](https://bats-core.readthedocs.io/en/stable/usage.html), [releases](https://github.com/bats-core/bats-core/releases)
- ShellSpec: [shellspec.info](https://shellspec.info/), [README](https://github.com/shellspec/shellspec), [releases](https://github.com/shellspec/shellspec/releases)
- commander: [README](https://github.com/commander-cli/commander), [releases](https://github.com/commander-cli/commander/releases)
- goss: [README](https://github.com/goss-org/goss), [releases](https://github.com/goss-org/goss/releases)
- runn: [README](https://github.com/k1LoW/runn), [releases](https://github.com/k1LoW/runn/releases)
- venom: [README](https://github.com/ovh/venom), [releases](https://github.com/ovh/venom/releases)
- atago: the [reference](/reference/) and the [runnable examples](/cookbook/#every-feature-has-a-runnable-spec) behind every claim
