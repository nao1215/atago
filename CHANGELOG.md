# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Fixed

- `json`/`yaml` `equals` no longer treats textually-different numeric strings as
  equal. A numeric-string coercion used `fmt.Sscanf("%g")`, which accepts a
  numeric PREFIX and ignores trailing bytes, so `"1.2.3"` parsed as `1.2` and
  `"3abc"` as `3` — making two different version strings compare equal and
  letting a string field silently satisfy a numeric matcher. Coercion now
  requires the whole string to be a valid number (`strconv.ParseFloat`) and only
  applies when at least one side is a genuine number, so `equals` on two strings
  is byte-exact (`"2"` ≠ `"2.0"`) while number/numeric-string equality is kept.
- Secret masking now masks the longest secret first. Sequential replacement let
  a short secret that is a substring of a longer one mask only its prefix and
  leak the remainder into reports and snapshots (masking `abcd` before
  `abcdefgh` left `efgh` visible).
- `atago manifest` and `atago explain` no longer drop dotted variable references:
  `spec.VarRefs` used a regexp that excluded the dotted names of namespaced
  built-ins (`${<mock>.url}`, `${<mock>.port}`, #24), so those references were
  silently omitted from the manifest's variable list and the explain output. Its
  pattern is now in lockstep with the expander again.
- `atago explain` now renders the `json`/`yaml` numeric comparison matchers
  (`gt`/`gte`/`lt`/`lte`); they previously showed as an empty matcher.
- A `line: N` selector (and `line`-scoped matchers) can now address a deliberate
  trailing blank line. Line splitting stripped EVERY trailing newline instead of
  the single phantom final one, so output ending in a blank line under-counted
  its lines and a spec pinning that blank line could never pass.
- `json`/`yaml` `length` on a string now counts characters, not bytes, so a
  multi-byte value like `"café"` has length 4 rather than 5.
- A spec that combines a shared `defaults.run` (env/shell/cwd/…) with an `ssh`
  step now loads: those defaults were layered onto the ssh step and then rejected
  by the ssh-runner validator, so any such spec failed to load even though the
  ssh step was authored bare. `defaults.run` no longer layers ssh-incompatible
  fields onto ssh steps.
- `contains: ""` / `not_contains: ""` (an empty-string element) is now rejected
  at load time: it is an always-true no-op (`contains`) or can never pass
  (`not_contains`), so it is caught like the empty-list case.
- A `store.from` stdout/body/rows/message/value/file selector is now shape- and
  syntax-checked at load time: it must set a `json` path or a `matches` regexp,
  and a malformed regexp or JSON path fails at load with a positioned message
  instead of aborting mid-run.
- `exit_code` accepts a YAML-quoted integer (`exit_code: "0"`) instead of
  rejecting it with a misleading "must be an integer" message.

### Docs

- Regenerated the snapshot-workflow README GIF (`doc/img/snapshot.gif`), which
  still showed the pre-rename `b3spec` command and `.b3spec.yaml` files instead
  of `atago` / `.atago.yaml`.
- Added `not_equals` to `examples/run_and_assert.atago.yaml` and `min_count` /
  `max_count` to `examples/dir_tree.atago.yaml`, filling the last matcher gaps in
  the runnable example suite.

## [0.3.0] - 2026-07-04

Exact workdir-delta assertions, hermetic per-OS home isolation, and interactive
terminal recording — closing the gap between "the command exited 0" and "the
command touched exactly these files, and nothing on the host."

### Added

- `changes` workdir-delta assertion (#70): `changes: {created, modified,
  deleted}` pins the EXACT set of files the immediately preceding run/pty step
  touched in the scenario workdir — the "this command writes only these files"
  contract no per-path `file:` assert can state. Each set field is exhaustive
  in both directions (every observed path must match an entry, every entry must
  match a path), so `modified: []` asserts "modified nothing"; an omitted field
  is unconstrained. The delta is content-hash based (a delete+rewrite of
  identical bytes shows in no list). Entries are doublestar globs, always
  `/`-separated: a single `*` stays within one path segment while `**` crosses
  `/` at any depth (`site/**`, `dist/**/*.css`), and a backslash escapes a
  literal metacharacter (#76). See `examples/changes.atago.yaml`.
- `sandbox_home` hermetic home isolation (#71): `sandbox_home: true` on `run`
  and `pty` steps points the child's home and per-OS config/cache/data/state
  directories (`$HOME`/`$XDG_*` on POSIX, `%APPDATA%` and friends on Windows)
  at a fresh `${workdir}/.atago-home`, so a CLI that reads or writes `~/.config`
  runs hermetically with one key. The isolated home is created once and reused
  across a scenario's steps, and its path is deterministic — an ordinary
  `file:` assert can inspect what the CLI wrote there. It composes with
  `clear_env`/`pass_env` (the sandbox home wins over a `pass_env: [HOME]` leak).
  It layers onto pty steps from `defaults.run` alongside the rest of the
  environment family (#77), so one `defaults.run.sandbox_home: true` governs
  run and pty steps alike.
- `atago record --pty -- <command>` (#69): records an INTERACTIVE session as a
  ready-to-replay `pty:` step. It runs the command in a real terminal, lets you
  drive one session by hand, and reconstructs an expect/send spec from the
  transcript — named keys where a lone control key was pressed, literal text
  otherwise (`${...}` escaped to `$${...}` so replay types it verbatim).
  Echo-off (password) input is NEVER recorded as its literal value: it becomes
  a live `${env:ATAGO_SECRET_n}` placeholder with a comment to set the variable
  and add it to `secrets:`. POSIX-only. The `--out`/`--force` existence check
  fires before any tty work, so a taken `--out` fails immediately.
- Hermetic environment control (#16): `clear_env: true` on `run`, `service`,
  and `pty` steps (and `defaults.run` / `defaults.service`) starts the child
  from an EMPTY environment instead of inheriting the host's, so host vars
  (`LANG`, `GIT_*`, proxies, ...) cannot silently change the behavior under
  test. `pass_env: [PATH, HOME]` re-admits an explicit allowlist (unset host
  vars are skipped); explicit `env:` overrides layer on top in the existing
  suite → scenario → step order. On Windows a system-critical set
  (`SystemRoot`, `SystemDrive`, `TEMP`, `TMP`, `PATHEXT`) is always retained.
  `pass_env` without `clear_env: true` is a load-time error (exit 2).
  See `examples/hermetic_env.atago.yaml`.
- `duration` assertion target (#31): `duration: {lt: 2s}` bounds the
  wall-clock time of the immediately preceding measurable step
  (run/http/query/grpc/pty) with lt/lte/gt/gte Go-duration bounds — "the
  command finishes within X", "the backoff actually waited" become
  declarative. At least one bound; lt/lte and gt/gte are mutually exclusive
  and any interval must be non-empty (validated at load, exit 2); a
  misplaced target (first in a scenario, or after fixture/store/assert) is a
  positioned load error. Failures show the measured duration with a
  CI-variance hint. See `examples/duration.atago.yaml`.
- `atago record -- <command>` (#30): run a command once in a scratch
  directory and generate a ready-to-edit spec skeleton from what it observed
  — exact exit code, the first non-empty stdout line as a `contains`
  matcher, `stderr: {empty: true}` when stderr was silent, and created
  files as `exists` asserts (capped at 10 with a note). `--out`/`--force`
  mirror `init`; `--shell` records shell command lines; `--snapshot` writes
  a stdout golden instead. The generated spec is validated in-process
  before it is written. Interactive (pty) and HTTP recording are explicit
  non-goals for now.
- Flaky-test tooling (#29): `--repeat N` runs each selected scenario N times
  (fresh workdir per iteration, sequential per scenario) to detect
  flakiness — any failing iteration fails the run and the console reports
  per-scenario pass rates; `--retry-failed N` re-runs failed/errored
  scenarios and reports recovered ones as **flaky** — counted green for the
  exit code but surfaced everywhere: `, N flaky` in the summary, `f`
  progress dots, JSON `status: "flaky"` + `attempts`/`iterations`
  (additive), JUnit `<flakyFailure>`, a TAP `# flaky` diagnostic, and a GHA
  `::warning`. The two flags are mutually exclusive (exit 3); fail-fast
  triggers only on a FINAL failure; flaky scenarios are not recorded by
  `--rerun-failed`.
- Colorized unified diff for equals/snapshot failures (#28): when both
  sides of a failed `equals`/`snapshot` comparison are multi-line, the
  console FAILED block renders a unified diff (3 context lines, removals
  red / additions green, hunk headers dimmed; snapshot sides labeled
  "snapshot (golden)" / "actual") instead of two raw Expected/Actual dumps.
  Color respects `--ci`/`NO_COLOR`; the uncolored diff text also lands in
  the JSON report's new additive `diff` field and the junit/tap/gha detail
  bodies. Inputs and rendered hunks are capped with explicit truncation
  notes; secrets are masked before diffing.
- PTY screen assertions (#27): the new `screen:` assertion target replays a
  pty step's transcript through a vt100 terminal emulator (hinshun/vt10x)
  sized by the step's rows/cols and asserts on the final RENDERED screen —
  what the user actually sees after every overwrite and erase — with the
  full stream matcher family (`line: N` addresses screen rows) including
  screen snapshots refreshed via `--update-snapshots`. Failures print the
  screen in a bordered block and export it to `--artifacts-dir`. The raw
  transcript stays on `stdout`. First declarative TUI E2E in the
  runn/ShellSpec/expect ecosystem. See `examples/pty_screen.atago.yaml`.
- PTY named keys (#26): `send: {key: enter}` presses a named key instead of
  embedding raw escape bytes — enter, tab, esc, space, backspace, delete,
  the arrows, home/end, pageup/pagedown, f1-f12, and ctrl-a..ctrl-z (mapped
  to standard xterm sequences; `{key: ctrl-d}` is the readable alias for the
  empty-send EOF rule). Unknown names are load-time errors listing the
  vocabulary; explain renders the keys symbolically.
- Recursive dir assertions and directory-tree snapshots (#25):
  `dir.recursive: true` makes `contains`/`not_contains` accept nested
  relative paths and `count`/`min_count`/`max_count` (files only) / `glob`
  walk the whole tree; `dir.snapshot:` pins the tree against a golden
  manifest (`dir <path>` / `file <path> sha256:<hash>` /
  `link <path> -> <target>`, sorted, /-separated, byte-exact hashes — CRLF
  differences ARE differences) refreshed with `--update-snapshots`, with a
  failure diff naming exactly the added/removed/changed paths;
  `dir.ignore:` glob patterns (`*.log`, `.git/**`) filter both. See
  `examples/dir_tree.atago.yaml`.
- Mock HTTP servers (#24): `mock_servers:` (scenario level) and
  suite.setup `mock_server:` steps start declarative stub HTTP servers on
  ephemeral loopback ports — canned routes matched on exact method+path
  (`json` / `body` / `body_file` payloads, optional `status`, `header`,
  `delay`), every incoming request recorded (unmatched ones answer 404 and
  stay visible). `${<name>.url}` / `${<name>.port}` are seeded into the
  store, and the new `mock:` assertion target checks what the CLI under test
  actually sent: request `count`, plus `header` / `body` matchers applied to
  the last matching request. Header matchers (http and mock) also gain
  `matches:` for regexp checks ("^Bearer "). Cross-platform (pure Go).
  Scaffold with `atago init --template mock`; see
  `examples/mock_server.atago.yaml`.
- `signal:` step (#23): send a named POSIX signal (TERM, INT, HUP, USR1,
  USR2, KILL) to a managed service's whole process group — scenario services
  and suite services both — for declarative graceful-shutdown and
  SIGHUP-reload testing. Handle-based targeting makes it race-free under
  `--parallel`, unlike `kill`/`killall` shell hacks. An optional
  `wait: {timeout: 5s}` blocks until the process exits and fails the step
  with a named message when it does not. POSIX-only (Windows reports a clear
  execution error, like pty). See `examples/signal.atago.yaml`.
- `exit_code: {in: [0, 2]}` (#19): assert the exit code against a set of
  accepted values — the contract shape of grep (0/1) or
  `terraform plan -detailed-exitcode` (0/2). Exactly one of the bare-int /
  `not` / `in` forms per assert; an empty or duplicated set is a load-time
  error. Failure output lists the accepted codes, and a timeout kill keeps
  its timeout hint.
- stdin sources (#18): `run.stdin` now also accepts `{file: path}` (a
  workdir-relative, `${name}`-expanded, path-confined file whose bytes are fed
  to the child) and `{base64: data}` (binary stdin, validated at load time; no
  `${name}` expansion, mirroring `fixture.base64`), alongside the historical
  inline string. The mapping form sets exactly one of file/base64 (exit 2
  otherwise). See `examples/stdin.atago.yaml`.
- Suite-level default step timeout (#17): `suite.timeout: 2m` bounds every
  `run`/`http`/`query`/`grpc` step that has no more specific timeout.
  Precedence: step > runner `timeout` > `defaults.run.timeout` >
  `suite.timeout` > built-in 60s. `timeout: "0"` (or `"0s"`) at any level
  disables the bound. A timeout kill names the level that supplied the bound
  in its failure hint. See `examples/timeouts.atago.yaml`.

### Changed

- **Steps are now bounded by a built-in 60s default timeout** (#17): a
  `run`/`http`/`query`/`grpc` step with no timeout configured at any level
  (step, runner, `defaults.run`, `suite.timeout`) now fails after 60s with
  "the command timed out ... and was killed" instead of hanging the run (or a
  CI job) forever. Specs relying on unbounded runs must either set a real
  bound (`suite.timeout: 10m`) or opt out explicitly with `timeout: "0"`.
- `defaults.run.stdin` is no longer accepted (#18): stdin is per-step input
  data, the same category as `command`. Declare it on each step.
- `defaults.run.timeout` is no longer string-merged into steps at load time;
  the engine resolves the timeout precedence chain itself so a runner-common
  `timeout` now correctly outranks `defaults.run.timeout`, and the failure
  hint can name the level that supplied the bound (#17).

### Fixed

- Unresolved-variable guard for pty sessions (#78): a `send`/`expect` entry
  referencing a `${name}` nothing defines (or an unset `${env:NAME}`) now fails
  the step with the same fix-forward error `run.command` produces — naming the
  entry, the reference, and the `$${...}` literal escape — BEFORE any terminal
  I/O, instead of silently typing (or matching) the literal reference text. A
  typo'd store name or a forgotten `env:` wiring, including a `record --pty`
  secret placeholder whose variable was never set, now fails loudly at the
  mistake. Escaped `$${...}` literals, named-key sends, and the empty-string
  EOF send are unaffected.
- E2E hardening (#75): every assert target's strings are expanded (not just a
  subset); `${` in recorded literal text is escaped so a generated spec never
  re-expands it; `shell: true` is rejected on ssh-runner steps (the command
  runs remotely, so a local shell flag is silently dropped) and the shell
  metacharacter hint is muted there; unsatisfiable spec shapes are rejected at
  load time; and `record --pty --out` is existence-checked up front.

## [0.2.0] - 2026-07-03

Suite-level bootstrap, interactive terminal testing, and verbose scenario
tracing — the three features real ShellSpec migrations were still missing.

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
