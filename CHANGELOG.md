# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Suite-level `setup:` / `teardown:` / `env:` (#7) — the bootstrap shell
  scripts real ShellSpec migrations could not shed (build a helper binary,
  start a shared peer, warm a cache) become spec YAML. `suite.setup` is an
  ordered list of steps run ONCE before any scenario inside a suite-scoped
  scratch dir (`${suitedir}`); a `service:` step — valid only there — starts a
  suite-wide background process at that exact point in the sequence, so
  build-then-serve-then-warm bootstraps keep their order. Setup stores and
  `ready.store` captures seed every scenario's store; `suite.env` is layered
  beneath each scenario's env. A failing setup step errors every scenario
  (labeled `suite setup`; nothing runs); `suite.teardown` always runs after
  the last scenario — pass, fail, error, or interrupt (bounded context) —
  while suite services are still up (services stop last, LIFO), and its
  failures are loud (console `SUITE TEARDOWN FAILED`, JSON
  `setup_failures`/`teardown_failures`) but never change the verdict.
  Surfaced in the JSON schema, `explain`, `manifest`
  (`suite_env`/`suite_setup`/`suite_teardown`), and a runnable example.
- Interactive terminal testing via `pty` steps (#8) — run one command inside a
  REAL pseudo-terminal and drive it with a declarative expect/send session,
  for CLIs that branch on TTY-ness or present interactive prompts (REPLs,
  wizards). `expect` waits until the transcript matches a regexp; `send` types
  into the terminal (an empty send transmits EOF/^D); the whole session is
  bounded by `timeout` (default 30s). The transcript (terminal echo included)
  becomes the step's stdout, so all stream matchers, snapshots (with their
  ANSI normalization), and `store from.stdout` work unchanged; a
  never-matching expect fails like an assertion with the pattern and
  transcript in the failure block (and as an --artifacts-dir sidecar).
  POSIX-only for now: the loader accepts the step everywhere, Windows reports
  a clear execution error (gate with `skip: {os: windows}`).

- `atago run --verbose` (#6): trace every scenario as it finishes — the
  expanded command, exit code / HTTP status, captured stdout/stderr (excerpted
  at the same limit as failure output), skip reasons, teardown steps, and each
  assertion's one-line verdict — for passing scenarios too, so authoring a
  spec no longer requires breaking an assertion to see what a command printed.
  Secrets stay masked; with a machine report (`--report json|junit|gha|tap`)
  the trace goes to stderr so stdout stays machine-readable; failing checks
  appear as one-line verdicts only (the full FAILED block is still rendered
  exactly once by the report).

## [0.1.0] - 2026-07-03

The first release of atago: an end-to-end test runner for command-line tools
driven by plain-YAML specs, with HTTP, DB, SSH, gRPC, and headless-browser
peers, snapshot testing, and Markdown doc generation.

### Added

- Scenario `teardown:` steps — cleanup that always runs after the steps (pass,
  fail, error, or interrupt), sharing the scenario's variable store so a
  `store`-captured resource id flows into the cleanup request. Built for
  external side effects the isolated workdir cannot undo (rows in a real
  database, resources created via an API, containers started by a run step).
  A teardown failure is reported — console `TEARDOWN FAILED` blocks and the
  JSON report's `teardown_failures` — but never changes the scenario's
  verdict; every teardown step runs even when an earlier one fails; after an
  interrupt, teardown gets its own bounded context. Surfaced in the JSON
  schema, `explain`, `doc`, and `manifest`.
- `${env:NAME}` interpolation — read a host environment variable anywhere
  `${name}` expands, including fields no shell ever touches (an http runner's
  base_url and headers, a db dsn, ssh credentials), so a CI-provided token or
  staging URL no longer needs a shell/store dance. An unset variable is an
  explicit error, not an empty string; `$${env:NAME}` stays literal; values
  listed under `secrets:` are masked as usual. `explain`/`manifest` surface
  host-environment reads as security notes.
- `not_matches` stream matcher — the regexp negation of `matches` on every
  stream target (stdout/stderr/body/rows/message/value), for "no
  warning/error lines" style assertions that `not_contains` (fixed strings)
  cannot express. Validated at load time like `matches`.

- Load-time validation for problems that previously escaped to runtime (exit 4)
  or silently misbehaved, all now exit 2 with a positioned message: a step's
  `runner:` must reference a declared runner of a compatible type (with the
  declared names listed on a miss), `run.timeout` / a runner's common `timeout`
  must be valid Go durations, every `matches:` (stream, json/yaml, `ready.log`)
  must compile as a regexp, and `fixture.mode` / `fixture.mtime` must parse.
- A hint for the single most common first-spec mistake: `stdout: hello` (a bare
  scalar where a matcher mapping is required) now explains the accepted shape
  (`stdout: {contains: ...}` / `{equals: ...}`) alongside the decoder's
  positioned error, for all stream targets.
- `atago run` warns on stderr when `--filter`/`--tag`/`--skip-tag` selects zero
  scenarios, so a typo'd selection in CI cannot greenlight silently (the exit
  code stays 0: nothing ran, nothing failed).
- Timed-out commands are now visible in failure output: an `exit_code`
  assertion against a command killed by `run.timeout` reports "the command
  timed out after Xms and was killed" instead of presenting the synthetic
  exit code -1 as a normal exit.
- `atago version` reports the module version for `go install`ed binaries (via
  the Go toolchain's embedded build info) instead of `dev`; release archives
  keep the exact tag injected at link time.
- A documented trust model (SECURITY.md): spec files
  are trusted input (a spec executes the commands it declares), and the
  network allowlist is enforced for the http/grpc/ssh runners — not for
  processes a `run` step spawns, a db DSN, or browser navigation.

### Fixed

- An authored `shell: false` now wins over a defaulted `shell: true`
  (`defaults.run.shell` / `defaults.service.shell`): `Run.Shell`/`Service.Shell`
  became a `*bool` so unset and false stay distinct. Previously the default was
  OR-ed in and could not be turned off per element, contradicting the
  documented "an explicitly authored value always wins" rule.
- A cmd runner's common `cwd`/`timeout` fields now reach the local command
  (they were documented as common to every runner but silently ignored).
  The step's own values still win.
- The JSON Schema accepts `version: 1` (a bare int) like the loader always did,
  so editor validation and runtime behavior agree.
- Flaky `TestEngine_ServiceLogPreservedOnLaterStepFailure`: the readiness probe
  now gates on the service's log output itself, closing the race where teardown
  could snapshot an empty output buffer under `-cover`.
- CI: gitleaks and reviewdog no longer fail on Dependabot PRs (Dependabot's
  read-only token 403s their PR API calls; both now skip Dependabot runs and
  declare least-privilege `permissions`).
- Docs: README's CI example pins the
  existing `setup-atago@v0` tag, documents Homebrew install and the `.zip`
  archive format on Windows; doc/RELEASE.md documents the `TAP_GITHUB_TOKEN`
  secret the release workflow needs.
- Release: the Homebrew cask strips macOS's quarantine attribute on install
  (the binary is unsigned, so Gatekeeper would otherwise block it).
