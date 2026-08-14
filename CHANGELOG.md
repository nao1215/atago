# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Third-party E2E suite for [OpenFGA](https://openfga.dev/), in
  `test/e2e/thirdparty/openfga/`, in parity with the `fga` CLI's own
  integration tests: local `fga model test` runs (glob patterns, the
  store-file reference containment contract with its `--allow-external-files`
  opt-in, the type-count limit) and a real `openfga` server driven through
  store creation, JSON/JSONL tuple writes and deletes (including a
  conditional tuple), and `store import` — each response asserted as JSON.
- Two website pages: a [comparison](https://nao1215.github.io/atago/comparison/)
  with Bats-core v1.14.0, ShellSpec 0.28.1, commander v2.5.0, goss v0.4.10,
  runn v1.9.4, and venom v1.3.0 (versions stated, sourced from official
  documentation), and a [migration guide](https://nao1215.github.io/atago/migrate/)
  mapping Bats and ShellSpec constructs to atago one-to-one. The guide is backed
  by executable parity suites under `test/e2e/migration/` — the original Bats
  and ShellSpec tests plus the migrated specs — run by the new MigrationParity
  workflow under pinned releases of both tools, and drift tests keep every
  snippet on the guide a verbatim excerpt of a file CI ran.

### Changed

- `atago record --pty` now notes in the generated spec's header when a send has
  no expect before it. The recorder anchors a send on the plain text that
  preceded it, but a full-screen TUI that redraws with pure cursor addressing
  prints no text to anchor on, so two sends end up back to back — and on replay
  they are written as fast as the engine can, where a program that drains
  typeahead on a mode switch reads the first and drops the second. The comment
  names the sends (by position) so the author can add an expect or expect_screen
  before them; a session where every send is anchored is unchanged.
- `atago record` now anchors an observed stderr the way it has always anchored
  stdout: on its first non-empty line. Only an EMPTY stderr produced an assert
  before, so recording the case a CLI author most wants pinned — a failing run
  whose whole observable behavior is a diagnostic on stderr — generated a spec
  that asserted the exit code and silently dropped the message. A first line
  carrying an escape or a carriage return is still left alone: that is a
  progress bar or a redrawing status line, and anchoring on one would hand the
  author a generated spec that is red on arrival.
- `record --pty` now anchors a generated expect on the latest output run
  carrying a letter or digit, keeping a purely decorative run only as the
  fallback. A full-screen TUI's last painted line is often a decorative rule
  (fzf's "3/3 ─────" info line, a status bar), so every expect of a recorded
  fzf session anchored on the same run of box-drawing characters — an anchor
  that matches any redraw and can release the next send before the state it
  depends on is on screen. The same recording now anchors on the match counts
  ("3/3", then "1/3" once the query applied).

### Fixed

- The box a failing `screen:` assertion draws around the terminal screen is now
  measured in display columns. It counted bytes first and runes after, and
  neither is what a terminal draws: a CJK character or an emoji is one rune and
  two columns, so a Japanese TUI — one of the screens this assertion exists to
  check — came out with its rows hanging past a border closed half a screen
  early. Box drawing stays one column, which is what the terminals TUIs run in
  give it, and the measure ignores the host locale: go-runewidth reads `LANG` at
  init and would otherwise frame the same screen one way on a CJK machine and
  another in CI.
- Every workdir- and spec-scoped path is now checked against a symlink at a
  DIRECTORY component, not only at the leaf. The containment check was lexical,
  so it compared path components and could not see that `escape/secret.txt`
  names a host file once the program under test has run `ln -s /etc escape` —
  and each path-taking key inherited the hole from the one resolver behind them.
  A `file:` assertion read the host file into the report and a `size:` bound
  disclosed its length; `fixture:`, `run.stdout_to`/`stderr_to`, `http.body_to`,
  and the snapshot writer created files, and whole directory trees, in the host
  directory — all while the run stayed green. The leaf refusal that was already
  in place saw none of it. Ancestors now face the same containment test the
  declared path does, with the rules `dir:` already follows: a dangling link is
  judged by its declared target — to the far end of a chain, since only the end
  says where a path goes — so the answer does not depend on whether that target
  exists yet, either spelling of a root that itself sits behind a symlink
  is accepted (macOS puts CI's workdirs behind `/private`), and a link that stays
  inside the root keeps resolving. A path that is merely absent is untouched, so
  `exists: false` still answers.
- A `dir:` assertion whose `path:` is a symlink to a directory now reads the
  same tree in every mode. `filepath.WalkDir` lstats its root, so the recursive
  and snapshot modes walked the link itself and saw an empty tree, while the
  non-recursive matchers followed it and read the real contents — one assertion
  giving two answers. The snapshot case was the worst of it: `--update-snapshots`
  wrote an empty golden, and an empty golden matches an empty walk forever, so
  the assertion stayed green whatever the directory later held. The same
  resolution closes the confinement hole underneath: the path check is lexical
  and could not see through a link the program under test planted, so a `path:`
  naming a link out of the workdir passed it and the directory outside was
  listed. The resolved path now faces the same containment test, and a link that
  escapes the workdir is refused — a dangling one by its declared target, so the
  rule does not depend on whether that target happens to exist.
- `record --pty` no longer masks every keystroke of a full-screen TUI session
  (fzf, vim, htop) as an `${env:ATAGO_SECRET_n}` placeholder. A TUI's raw mode
  clears ECHO and ICANON together, and the recorder treated ECHO alone as a
  password prompt, so a recorded TUI session replayed nothing. Secret masking now
  requires the termios state of an actual password prompt — echo off while
  canonical (line) input stays on, as with `read -s`, sudo, and ssh — and
  raw-mode keystrokes are recorded literally, so a recorded TUI session
  round-trips green.
- A pty step or `record --pty` no longer loses the output of a command that
  prints and exits at once. macOS discards whatever the terminal still holds the
  moment its last slave handle closes — a read already parked in the kernel comes
  back EOF with no bytes — so handing the terminal to the child and closing
  atago's own handle immediately made the child's exit a race against the drain,
  whatever order the drain was started in. It surfaced as an empty transcript
  from a scenario whose command clearly printed, roughly once in a few hundred
  fast-exit runs. atago now keeps its slave handle until the child has been
  reaped, which puts the terminal's last close after the transcript has been
  read, on a schedule atago controls.

## [0.19.0] - 2026-08-10

A release about the two things a CLI suite still had to say in bash, and about
finishing the terminal.

Nine downstream repositories built the binary under test in a wrapper script
before calling atago, and exported a throwaway `HOME` and the XDG variables for
a whole tree of specs — even though `sandbox_home: true` already did that per
step — because the only way to apply anything to 89 spec files was to repeat
`defaults:` in all 89. `atago.project.yaml` says it once per directory, and
`subject:` inside it says how to build the binary and what specs call it, so the
wrapper script has nothing left to do.

The other half is the terminal. `{key: ...}` existed so a session could be read
without `\x1b` escapes, but the vocabulary had holes exactly where real TUIs
live. Seven additions close them: the chords a form-style TUI binds, a repeat
count, bracketed paste, a mid-session resize, a host command run while the
session is live, mouse clicks and scrolling, and color/attribute matchers on the
rendered screen.

Four assertions and run behaviors round it out: an expected-failure scenario
that keeps a known bug's reproduction in CI, occurrence counts, byte sizes, and
a determinism check that reruns a command and requires identical bytes.

### Added

Directory-level configuration:

- `atago.project.yaml`, a manifest for a TREE of specs rather than one file. It
  carries `env:`, `defaults:`, and `fixtures_dir:`, is discovered by walking up
  from a spec to the nearest one — so `atago run ./e2e` and `atago run
  ./e2e/cli/one.atago.yaml` resolve the same configuration — and a spec file's
  own values always win. `atago explain` prints which manifest applied, because
  configuration that applies to a file without appearing in it has to be visible
  somewhere.
- `subject:` declares the binary under test: `build:` produces it, `artifact:`
  says where, and `name:` is what specs call it (with the `.exe` suffix resolved
  per OS). `atago run` builds it once per invocation, before any scenario, and
  puts the artifact's directory first on `PATH` so a spec keeps invoking the tool
  the way a user does. `profiles:` are named build variations selected with
  `--profile NAME` — a coverage-instrumented build, a race build, a different
  toolchain — which is what keeps coverage out of atago: instrumenting is an
  alternate build command plus some environment in every language, and merging
  the raw profiles afterwards is a toolchain job.
- `${specdir}` and `${fixtures}` reach committed input trees. Scenarios run in an
  isolated temp workdir, so both are absolute, and both are inputs: steps write
  into `${workdir}`, never into them. `${fixtures}` comes from the manifest's
  `fixtures_dir`, validated at load time so a typo is one message about the
  manifest instead of N per-scenario failures about a missing file.
- `suite.env` values may reference what `suite.setup` captured — a `store` step's
  value, or the ephemeral address a suite-wide service published through `ready:
  {store:}` — so a stub registry or proxy reaches every scenario without a shell
  wrapper exporting it. A value that cannot resolve is never passed on as the
  literal text `${name}`: a child process does not fail on that, it uses it, and
  the error would then arrive from the tool under test rather than from the spec.

Assertions and run semantics:

- `expect_fail:` keeps a known bug's reproduction in the suite instead of exiling
  it to a directory CI never runs. A scenario that fails is an XFAIL and the run
  stays green; one that passes is an XPASS and the run fails, which is what gets
  the spec promoted into the guarded suite. An execution error stays an error —
  `expect_fail` says "the program gives the wrong answer", not "the spec cannot
  run". `reason:` is required, `issue:` makes the XPASS message actionable, and
  `--allow-xpass` keeps the run green while the spec is promoted, moving every
  failure-level signal with the exit code: no `<failure>` in JUnit, no `::error`
  in GitHub Actions, nothing in the JSON `failures[]`.
- `count` / `min_count` / `max_count` on stream and file matchers answer "how
  many times", which is the question a duplicate-output bug passes when the only
  matcher available is `contains`. Occurrences are non-overlapping and the
  failure names the lines each one landed on.
- `size` / `min_size` / `max_size` assert byte counts on files and streams, and
  compose with a content matcher — the `wc -c` a spec used to reach for.
- `deterministic:` reruns a command and requires the declared observables to come
  back byte-identical, which is the only cheap oracle for output whose order
  leaks a map iteration. Assertions, `store`, snapshots, and `changes:` still
  describe the first run. When a rerun changes the workdir, the failure says so:
  the check is only meaningful for an effectively read-only command.

Terminal sessions:

- The named-key vocabulary covers what real TUIs bind: `shift-tab`, `alt-a`..
  `alt-z` with `alt-enter`/`alt-backspace`, the modified arrows
  (`ctrl-`/`shift-` plus a direction), and `insert`.
- `times:` repeats a named key send, so moving sixteen columns is one step rather
  than sixteen.
- `send: {paste: ...}` delivers text as a bracketed paste, taking the program's
  paste path instead of its typing path — the difference a REPL's multi-line
  handling turns on. It fails if the program has not enabled paste mode, rather
  than silently typing the brackets.
- `resize: {rows, cols}` delivers a real size change mid-session, so the redraw a
  window resize triggers is reachable from a spec.
- `exec:` runs a host command while the session is live, which is how a watcher's
  reaction to an external change becomes testable.
- `send: {mouse: ...}` sends SGR 1006 mouse reports for clicks, drags, and
  scrolling, gated on the program having enabled mouse mode.
- `screen:` and `expect_screen:` gained `attrs:` — assert the foreground and
  background color, bold, and reverse video of a region, which is what makes a
  "selected row is highlighted" contract testable.

### Changed

- `--fail-fast` now stops for every outcome that fails the run, not just a
  failed or errored scenario. An XPASS (a known bug that is fixed) and a flaky
  recovery both decide the exit code, so a run whose verdict was already made
  kept scheduling scenarios — the flag stopping for one red signal and not
  another, with nothing in the reports to explain the difference. `--allow-flaky`
  and `--allow-xpass` make those statuses green again, and then there is nothing
  to stop for. One predicate now answers the question for the exit code,
  `--fail-fast`, and the `--rerun-failed` ledger.
- The `--rerun-failed` ledger records an XPASS. A run whose only red signal was
  an XPASS wrote no ledger at all, so the follow-up `atago run --rerun-failed`
  reported "nothing to rerun" and exited 0 — a green answer about a run that
  failed. Re-running an XPASS reproduces it, which is the outcome that keeps
  saying "promote this spec" until someone does; under `--allow-xpass` the run
  is green and the ledger stays empty. A flaky scenario is still not recorded:
  its re-run most likely passes, and clearing the ledger is not the same as
  fixing the instability.

### Fixed

- A flag written after the spec paths is now read as a flag. `atago run specs/
  --report json` exited 3 with `cannot access "--report": no such file or
  directory` — an error about a file nobody typed — because Go's flag package
  stops parsing at the first non-flag argument, and everything after it became a
  spec path. "What to run, then how to run it" is the order people reach for
  first, so the failure landed on a correct-looking command line and named the
  wrong problem. `run`, `list`, `doc`, `explain`, `manifest`, and `init` now
  accept flags in any position. `--` still ends flag parsing, which is how a file
  whose name starts with a dash is addressed; `atago record` is unchanged,
  because there the trailing arguments are another program's command line and its
  flags belong to it.
- A pty session whose program exits immediately no longer loses its transcript.
  The terminal handed the child a slave descriptor derived from the master, so a
  fast exit could race the drain and leave the capture empty; the master is now
  paired with its own slave, and the drain starts before the child does.
- Closing a Windows ConPTY ends a read already parked in the kernel. `Read`
  issues a synchronous `ReadFile`, which `CloseHandle` does not abort, so a
  timed-out capture waited out its whole drain grace and leaked the goroutine for
  as long as a surviving conhost held the pipe open. `Close` now cancels pending
  I/O on the handle first.
- A run step whose timeout expires before the process starts is reported as a
  timeout rather than as a generic execution error.
- `atago doc` renders invisible characters in assertion values instead of
  dropping them, so a spec asserting on a control character produces documentation
  that says which one.
- Snapshot normalization folds CRLF after stripping escape sequences, which makes
  `Normalize` idempotent: a golden recorded from output carrying both no longer
  depends on how many times it has been normalized.
- `atago record` takes the same reader handoff before starting the child that a
  run step does, closing the window where a fast-exiting recorded program lost
  output.

### Documentation

- A cookbook recipe for `--fail-fast`, the one run flag with no recipe and no
  page of its own: what counts as a red verdict, what happens to the scenarios
  that never started, and how it pairs with `--rerun-failed`. The CI page and
  the reference gained the same facts, and the progress-dot legend in the README
  and the getting-started page now lists the flaky and expected-failure markers
  it has been printing since 0.18.0.

### Upgrading

Two invocation-level changes can be noticed by an existing pipeline:

- A spec path spelled like a flag (`-weird.atago.yaml`) used to be collected as a
  path once it followed another path. Now it is a flag, so name it after `--`:
  `atago run -- -weird.atago.yaml`.
- `--fail-fast` stops for an XPASS and for a flaky recovery, so a run using it
  may execute fewer scenarios than before and report the rest as `skipped after
  fail-fast`. `--allow-xpass` and `--allow-flaky` restore the old scheduling for
  the status they accept.

## [0.18.0] - 2026-07-30

A release about the one verdict atago was getting wrong about itself. `--repeat`
and `--retry-failed` exist to surface instability, and a run that found some
still exited 0 — so a suite could rot with CI green, and `--repeat`'s own help
said "any failing iteration fails the run" while a partial failure did not. A
flaky scenario now fails the run, and `--allow-flaky` is there for a suite whose
instability you already know about.

### Changed

- **A flaky scenario now fails the run.** A scenario that failed and then passed
  on a `--retry-failed` re-run, or that passed some `--repeat` iterations and
  failed others, was reported loudly and exited 0. So the two knobs whose whole
  purpose is to surface instability went green when they found it — `--repeat`'s
  own help said "any failing iteration fails the run" while a partial failure
  exited 0 — and a suite could rot without CI ever noticing. Downstream, sqly's
  runner passed `--retry-failed 3` for the sake of its interactive-pty specs, and
  that silently extended tolerance to every other spec in the suite: a real
  nondeterminism bug in filesql, where an LTSV table's column order came out of a
  map, surfaced there as "flaky, PASSED" rather than as a failure.

  A flaky scenario is still reported as flaky and not as a hard failure — TAP
  keeps its `ok` point with a diagnostic, junit keeps `flakyFailure` inside a
  passed testcase, gha keeps its warning — because per scenario the recovery is a
  fact. What changed is the run-level verdict: the console summary now reads
  FAILED, matching the exit code, where before the headline and the exit code
  could disagree.

### Added

- `--allow-flaky` exits 0 when the only problem is flakiness, for a suite whose
  instability is already known and accepted. atago's own pty scenarios lose
  keystrokes when their sessions are starved of CPU, so its `dogfood`,
  `thirdparty`, and pty CI jobs pass it — which puts the tolerance at the command
  line instead of in a comment beside a flag that granted it invisibly.

### Upgrading

A run that was green only because a scenario recovered on retry now fails. That
is the point, but it can turn a passing pipeline red on the first run after
upgrading. Two ways forward: fix the flake, which is what the report already
names, or pass `--allow-flaky` for a suite whose instability is accepted.

## [0.17.0] - 2026-07-28

A minor release about reach, and about failures that describe themselves.
Windows was the one platform CI has always tested and no package manager could
install from; atago now ships a Scoop bucket. The rest continues what v0.16.0
started: an assertion that observed nothing, and a run whose output capture died
before its process did, now say so instead of leaving the reader to infer it.

### Added

- Scoop distribution for Windows. `scoop bucket add nao1215
  https://github.com/nao1215/atago`, then `scoop install nao1215/atago`. The
  bucket is this repository's own [`bucket/`](bucket/) directory — Scoop accepts
  any git repository with one — so the manifest is published by the release
  job's built-in token, with no second repository and no second credential to
  keep alive.
- Third-party E2E suite for [fx](https://fx.wtf/), the terminal JSON viewer, in
  `test/e2e/thirdparty/fx`. fx is a filter when it is given a reducer and a
  full-screen viewer when it is not, and both halves are pinned. The filter side
  covers the JavaScript semantics that surprise a jq user (a missing key is
  `undefined` on stderr with exit 0; dereferencing through it is exit 1),
  `exit()` choosing the process's exit code, the identity path keeping number
  literals a float64 round trip would rewrite, the YAML/TOML/raw/slurp parsers
  and the filename-driven detection that switches them, `--strict` turning the
  tolerated comments and trailing commas back into errors, and base64 and YAML
  round trips. `save` is pinned with exhaustive `changes:` deltas — in-place
  rewrite, idempotence, and the no-file and symlink refusals — alongside
  `.fxrc.js` read from the workdir and from a sandboxed home that stays
  unwritten. The viewer runs under the pty runner: folding, regexp search
  reporting the cursor's JSON path, and the split that makes
  `fx data.json > picked.txt` work, with the interface on stderr and only the
  chosen value on stdout.
- `examples/pty_stdout_split.atago.yaml` and a matching cookbook recipe for that
  last pattern: an interactive tool whose UI goes to stderr while stdout carries
  the result. One pty step with a shell redirect asserts both sides — the
  interface through `expect_screen:`/`screen:`, and what the pipeline received
  through a `file:` assert, including the empty file that proves nothing of the
  UI leaked.

### Fixed

- A failure whose observed value was empty now says so. The console block prints
  an `Actual:` section only for a non-empty value, so an assertion over a stream,
  file, or snapshot that captured nothing rendered with no `Actual:` at all — the
  same shape as a report that never recorded what it saw. Reading such a failure
  meant knowing that the omission itself meant "empty" (#339). An observed empty
  payload now renders as `(empty)`, the wording an empty JSON/YAML payload
  already used.
- A run step whose process ran to completion while its output capture did not no
  longer reports `failed to execute`, which says the opposite of what happened.
  It now names the exit code the process reached along with the capture error,
  and — for the one cause a spec author can act on, a background process that
  outlives the command and holds the pipes open past `WaitDelay` — points at the
  `service:` step. The bytes read before the cut are withheld rather than handed
  back as the observed stream, and `stdout_to`/`stderr_to` no longer write them
  to a file either: a stream atago only partly read must not be assertable as
  though it were complete, in the step that produced it or in a later one.

## [0.16.0] - 2026-07-27

A minor release about output you can act on. Three failure messages were true
and useless, in the way that costs the most time: they pointed at the wrong step,
or offered a knob that could not help. The spec language gains one key,
`changes: ignore:`, for the path a program writes only sometimes.

Two things a spec or a script may notice:

- A `--repeat` or `--retry-failed` run writes each attempt after the first into an
  `attempt-<N>` subdirectory of `--artifacts-dir`. A run with neither flag writes
  exactly where it always did; tooling that globs the artifact tree may see the
  extra level.
- A `pty` step killed with its program still running reports a different message
  and step name (`pty session leaves the program exited`), and a failure block
  names the command the failing check asserts on instead of the scenario's last
  command. A spec that pins atago's own console text will need updating; a spec
  that asserts on the program under test is unaffected.

### Added

- `changes: ignore:` drops matching paths from the observed workdir delta before
  the categories are compared, so a path the program under test writes only
  sometimes stops being a choice between a flaky spec and no assertion at all. A
  delta is exhaustive in both directions, which is what makes `modified: []` a
  proof; but a tool that starts a background service on some runs, or writes a
  cache into the sandboxed HOME, could not be pinned: naming the path failed on
  the runs that skipped it, and omitting the category asserted nothing. An ignored
  path is neither an unexpected change nor evidence for an entry — an entry that
  names an ignored path still fails — and an ignore glob that matches nothing is
  fine, since describing what may appear is the point. The globs are doublestar
  patterns, as in `dir: ignore:`. `atago doc` and `atago explain` name the ignored
  globs, so a generated page does not advertise a claim the assertion no longer
  makes. The yazi suite uses it for the state directory yazi writes on some runs.

- Third-party suite for Info-ZIP `zip` and `unzip`, the first real-world coverage
  of the whole `dir:` family: an extracted tree asserted by recursive membership,
  a file-only count, a glob, and a snapshot manifest that fixes the shape in one
  assertion. Round trips are proven byte for byte with `equals_file` (an empty
  file, mixed line endings with no trailing newline, NUL bytes), `-j` flattening
  is checked as the metamorphic pair of the tree case, and the overwrite decision
  is pinned in all three forms (declined by default, `-n`, `-o`). The documented
  exit codes get one scenario each — 0, the warning code 1, 9 for a missing or
  unreadable archive, 10 for a bad option, 11 for a name that matches nothing,
  and zip's 12 for nothing to do. A second spec covers path traversal: a hostile
  archive holding `../escaped.txt`, `/abs.txt`, and a four-level `../../../../`
  entry is stripped, reported, and kept inside the destination, which `changes:`
  proves against the filesystem instead of trusting the warning text.

### Changed

- A pty step killed because its program was still running now says that, instead
  of offering a longer timeout that cannot help. A session does not close the
  program: when the actions run out and it is still up, the step is killed, and
  the old message ("the command timed out ... raise the timeout if the command is
  merely slow") sent the reader to a knob that changes nothing, since the session
  had already run to the end. It now reports that the session finished with the
  program still running and names the fix — send its quit key last, and wait for
  the screen to show the program can take it, because a dialog or prompt owns the
  keyboard until it closes. A wait that never completed is a different failure and
  keeps its own message naming the pattern. This is the shape that broke thirteen
  scenarios in the lazygit suite and three in yazi.

- The scheduled third-party matrix no longer reports the runner's weather as a
  broken tool. Its pinned Ubuntu snapshot mirror served a 500 for the
  command-not-found index — a file nothing here uses — and `apt-get update` exits
  100 for any index it cannot fetch, so five suites never ran. That index is no
  longer requested, and the update and install are retried against a mirror that
  5xxes under load. The version pin itself is unchanged: the snapshot id and the
  SHA256-verified release archives still decide what gets installed.
- The scheduled matrix runs with `--retry-failed 1`: a flake is re-run once in a
  fresh workdir and reported as flaky, while a regression still fails both
  attempts. Raising the two terminal-UI suites' timeouts from 20s to 45s was tried
  in the same change and reverted — it made the lazygit leg worse (4 failed
  scenarios became 13, and the suite went from about a minute to six), because the
  failures land on the plain `git` step that follows the pty step rather than on
  the session itself. Tracked in #330.

- Adding an assertion target is now guarded instead of remembered. Five layers
  dispatch on a target — the loader's validation, the runtime check, `atago doc`,
  `atago explain`, and the published JSON Schema — and each had its own switch
  whose default silently degraded: doc and explain rendered the bare target name
  as a sentence, and the schema would have made a valid spec light up red in an
  editor. `spec.AllAssertTargets` now derives the list from one table, a test
  proves that table matches the fields on `Assert`, and each layer has a coverage
  test that walks the list. A half-wired target fails a test rather than shipping
  a weaker document.
- A `changes:` entry that names a path the delta does not track now says what the
  path is. A workdir delta scans regular files and symlinks, so a directory the
  step really did create — or a named pipe, or a socket — is invisible to
  `created`/`modified`/`deleted`, and the failure read "matched no file the step
  created" while the author could see the thing sitting in the workdir. The note
  names the kind and, for a directory, points at `dir:`.
- A service that dies during its readiness probe now reports the exit status.
  "service exited before it became ready" left the two cases an author checks
  first indistinguishable: a missing binary (127) and a command that ran and
  rejected its own configuration. The captured service output is still shown
  beneath it.

### Fixed

- A failure now names the command it was observed under. The console block and the
  JSON report's `failures[].command` used the scenario's LAST run command, so a
  scenario that keeps working after the failure blamed the wrong step: an
  interactive step killed by its timeout was reported against the verification
  command that followed it and passed. That is a diagnosis pointing the wrong way,
  and it cost real time — the reader tunes a step that never had a problem.

- `atago record --pty` now replays the bytes it recorded. A typed burst that is
  not valid UTF-8 — a paste in another encoding, a keyboard macro emitting a high
  byte — was rendered as text, so every such byte became U+FFFD and the generated
  send typed three different bytes than the session captured. Those bursts are now
  emitted as `!!binary`, the same escape hatch plain `record` already uses. An
  `expect` anchor built from output carrying such a byte was corrupted the same
  way, producing a pattern that could never match: the recorded session hung until
  its timeout. An anchor is now the longest stretch of the line that really exists
  in the transcript.
- `atago record --pty` no longer fails on a command carrying a control character.
  The scenario name was the recorded command verbatim, and the loader rejects a
  control character in a name, so recording `printf 'a\rb'` produced a spec that
  failed its own validation with "this is an atago bug, please report it". The
  name is now a single-line label, as plain `record` already did; the command
  itself is still replayed verbatim.

- An assertion no longer hangs on an entry that is not a regular file. A `dir:`
  walk fingerprinted every non-directory entry it found, and a `file:` matcher
  read its target: opening a named pipe for reading blocks until another process
  writes to it, and nothing bounds the assertion phase — a step timeout covers
  the command, not the checks — so a program under test that left a `mkfifo` pipe
  (or a socket, or a device node) where a file was expected froze the run until
  the CI job was killed, instead of failing it in milliseconds. A `dir:` tree now
  records such an entry by kind and never opens it: it stays a present path for
  `contains`, is not one of the regular files `count` counts, and appears in a
  tree snapshot as `fifo <path>` rather than a content hash. A `file:` matcher
  fails saying the path is a named pipe, not a regular file.

- `--artifacts-dir` no longer lets one attempt of a scenario overwrite another's
  payloads. A `--repeat` run reports the first failing iteration inline and
  points at its sidecar files, but every iteration wrote to the same paths, so
  the last failure overwrote the evidence the report still referenced: opening
  the artifact showed a payload from a different iteration than the diff printed
  beside it, and the other iterations' payloads were gone. Each attempt after the
  first now writes under an `attempt-<N>` directory, which also keeps a
  `--retry-failed` attempt's evidence instead of dropping it. A run with neither
  flag writes exactly where it always did.

- `json: equals: null` now asserts that a field is null instead of being
  rejected as a matcher-less check. A YAML null and an omitted `equals` key both
  decode to the same Go nil, so the loader could not tell "assert this value is
  null" from "no matcher was written" and turned the first into an error. It now
  records which one the spec authored: `equals: null` passes only for a JSON
  null — the string `"null"`, `0`, `false`, `""` all fail, and a path that
  selects nothing still reports that rather than passing — while a check with no
  `equals` key is rejected exactly as before. "This field is null" and "this
  field is absent" are different contracts, ordinary in API output, and neither
  could be pinned. `atago doc` and `atago explain` spell the expected value
  `null` rather than Go's `<nil>`, matching the failure messages.

## [0.15.0] - 2026-07-27

Two changes can turn a passing spec red, so read these first.

- `changes:` now tracks symlinks. A step that plants a link, retargets one, or
  removes one is now a creation, a modification, or a deletion. A spec that
  wrote `created: []` for such a step passed only because the link was
  invisible to the delta; it now fails and has to name the link.
- `file: exists` and `file: executable` no longer accept a directory. Both were
  satisfied by one — every directory carries the execute bit — so an assertion
  written to catch "the tool left a directory where a file belongs" passed
  instead of catching it. Both now fail and point at `dir:`.

The rest of the release is about failure messages that name the difference
rather than printing both sides identically, and about three new third-party
suites — Ghostscript, zstd, and ImageMagick — that put the `pdf:` and `image:`
matchers on the output of real programs, which is how the two `pdf:` extraction
bugs below were found.

### Added

- The git third-party suite gains a second spec covering what a script leans on:
  `diff --exit-code` as a status contract, the exact `--porcelain` status codes,
  a fast-forward merge and a real conflict, stash and patch round trips proven
  byte for byte, `.gitignore` keeping a file out of the index, a local clone
  reproducing the same commit, content-addressed identity for two identical
  commits, and the documented failures (no repository, duplicate tag, nothing
  staged) with their own exit codes.
- Third-party suite for Ghostscript, the first real-world coverage of the `pdf:`
  matcher: PostScript converted to PDF and asserted by page count, extracted
  text, and Info-dictionary metadata; the rasterized page asserted with
  `image:`; and the relationships pinned — a selected page keeps its own text
  and drops the other, merging sums the pages, `txtwrite` agrees with the
  extracted text, and a repeated conversion yields identical text. The failure
  side records what Ghostscript actually does, including a failed conversion
  that still leaves a PDF behind.
- Third-party suite for zstd: compression round trips proven byte for byte with
  `equals_file` (including an empty file and one full of NUL bytes), the input
  kept unless `--rm` is asked for, an existing archive never silently
  overwritten and `-f` overriding that, `-t` integrity verification of a good
  and a corrupt frame, a failed decompression leaving no partial output, the
  stdin-to-stdout pipeline, and the missing-input and unknown-option failure
  contracts.
- Third-party suite for ImageMagick: conversion output decoded and checked as an
  image (format sniffed from the bytes, dimensions, alpha), aspect-ratio versus
  forced resize, a PNG↔PPM round trip verified twice — once by atago's pixel
  comparison and once by `magick compare -metric AE` as an independent oracle —
  metadata stripping that leaves the picture untouched, the `identify` JSON
  contract, and the missing-input, corrupt-input, and unknown-extension
  behaviors. Inputs are synthesized with `xc:` canvases, so the suite carries no
  image assets.

### Changed

- The messages for a target atago cannot use now say what atago was doing.
  A missing spec path reads `cannot access "x.atago.yaml": no such file or
  directory` instead of repeating the path through Go's `stat` wrapper; an
  unreadable directory reads `cannot search "/etc" for spec files: ...` instead
  of naming a file the user never mentioned; and a directory with no specs names
  the directory it searched and points at `atago init`.

- Counted nouns in messages now agree with their number: "1 page" instead of
  "1 pages", "1 entry" instead of "1 entries", "1 line" instead of
  "1 line(s)", and `atago record` reports "1 file created" rather than
  "1 file(s) created". Three packages had spelled this three different ways;
  they now share one helper, so a reader never has to notice which part of
  atago wrote a sentence.

### Fixed

- `pdf: text:` no longer loses everything after the first stream that ends
  without a newline. ISO 32000 recommends an EOL before `endstream` but does not
  require one, and Ghostscript omits it, so the scan ran past the true end of a
  stream and swallowed the objects that followed: a two-page Ghostscript PDF
  reported only its first page's words.
- `pdf: metadata:` now finds the Info dictionary when the producer stores it in
  a compressed object stream, which every PDF 1.5+ writer does by default —
  Ghostscript 10, LaTeX, Word. Only the raw bytes were scanned, so a metadata
  assertion reported "field not present" for an ordinary modern PDF while page
  count and text extraction (which already go through the decompressed streams)
  worked. A value stored in the clear still wins over one repeated in a stream.
- A JSON `null` is now spelled `null` in matcher subjects and failure messages
  instead of Go's `<nil>`. The Go-ism made a message about someone else's
  document read as a message about atago's implementation, and left `matches`
  with nothing writable to match a null field against; the `store` path already
  refused to leak it.
- A failure whose two sides differ only in invisible characters now quotes both.
  A byte-exact `equals` mismatch between `shared\n` and `shared\r\n` printed the
  same word twice, and the same happened for a trailing space or a tab against
  spaces: the reader could see there was a difference but not what it was.
  Ordinary differences keep their plain, copy-pasteable form.
- A `file:` assertion is no longer satisfied by a directory. `exists: true`
  passed for a tool that produced a directory where a file was expected, and
  `executable: true` passed on any directory, since every directory carries the
  execute bit — two false passes in exactly the case the assertion exists to
  catch. Both now fail and say the path is a directory, pointing at `dir:`.
  atago's own image suite had one of these sloppy assertions, now corrected.
- `--rerun-failed` now names the recorded failures it could not run. When a
  scenario was renamed or deleted while still broken, the rerun quietly reported
  fewer scenarios than were recorded, which reads as "the rest are fixed" — the
  all-gone case already warned, but the partial case, where the miscount is
  easiest to miss, did not. The entries are still kept in the ledger for the
  next rerun.
- `image: similar_to` now falls back to the scenario workdir when no committed
  baseline sits next to the spec. Comparing two images the run itself produced —
  the two ends of an encoder round trip, the before and after of an in-place
  edit — failed with "could not read baseline image" unless the spec spelled out
  `${workdir}/...`. A committed golden still wins, so no existing baseline
  changes meaning, and the workdir path is confined to the scenario like every
  other runtime path. The failure message now names both places it looked.

- `atago doc` and `atago explain` now describe every matcher an assertion sets
  instead of the first one. A stream assertion that combines `contains` with
  `not_contains`, `matches`, or `not_matches` enforces all of them at run time,
  but the generated page advertised only the first — publishing a weaker
  contract than the suite guards. A `line: N` selector is named too, so a
  matcher scoped to one line no longer reads as one pinning the whole stream.
- `file:` assertions using `not_contains` or `executable` rendered as "the file
  is checked" in generated docs and `explain` output, hiding what the scenario
  guarantees. Both matchers now have their own phrasing.
- A failing `json: equals` now names the difference. Both sides were rendered
  bare, so a type or whitespace mismatch printed the same text twice — the
  boolean `true` against the string `"true"` read as "expected true, got true",
  and `" x "` against `"x"` hid the spaces. Strings are quoted and `null` is
  named; numbers and booleans stay bare.
- A `dir:` assertion on a path that holds a file reported a bare
  `exists=false`, which reads as "nothing is there". It now says the path
  exists as a regular file (or symlink, socket, …) and points at `file:`.
- `changes:` now tracks symlinks. Only regular files were scanned, so a step
  that planted a link — an installer wiring `bin/tool`, a release that swaps a
  `current` pointer — passed `created: []`, reporting "nothing happened" for a
  run that added a path. A link is compared by the target it names: planting one
  is a creation, retargeting it is a modification, removing it is a deletion,
  and a dangling link stays visible. Writing through a link is still a
  modification of the target file, and the delta never resolves a link, so a
  cycle or a target outside the workdir is not followed.

## [0.14.0] - 2026-07-26

A minor release about the specs as documentation. A spec has always described
what a CLI does; now it can also say what that is *for*. Suites and scenarios
take an optional `description:` that `atago doc` publishes, and every
third-party suite on the real-world pages has been given one.

### Added

- `suite.description` and `scenario.description`: optional prose that `atago
  doc` renders as Markdown under the matching heading, so a generated page
  explains what a suite guarantees and why a scenario exists instead of leaving
  the reader to infer it from step mechanics. Both are documentation-only —
  never expanded, never read by the engine, absent from `explain`, `manifest`,
  and the JSON report — and both are optional: a spec without them generates
  byte-identical output to before. YAML comments are unaffected and still never
  reach the generated docs, which is what keeps install notes and TODOs out of
  published pages.

### Changed

- A failed `dir:` assertion now lists what the directory actually holds instead
  of reporting only an abstraction (`missing`, `present`, `3 entries`,
  `no match among 2 entries`). Every matcher in the family is covered, in both
  direct and recursive mode, and the listing marks directories with a trailing
  slash and shows symlink targets. It is capped at 40 entries so a large tree
  cannot bury the rest of the report. This needs no spec change: every existing
  `dir:` assertion produces the better failure block as-is. A count failure also
  reads "has 1 entry" rather than "has 1 entries".

### Documentation

- Every third-party suite behind the
  [real-world pages](https://nao1215.github.io/atago/real-world/) now carries a
  suite description, so each of the 38 published pages opens by saying what the
  tool is and what the suite guarantees about it rather than starting cold at
  the first scenario. The maintenance notes those specs already carried —
  install steps, port allocation, platform gates — stay YAML comments and remain
  out of the published output.

## [0.13.0] - 2026-07-26

A minor release focused on removing false greens. A command killed by its
timeout no longer reports passed when nothing asserts on it, and the spec loader
turns away explicit YAML tags instead of provoking a decoder panic on the way to
a parse error. Three more third-party CLI suites join the real-world coverage.

### Added

- Third-party coverage grows by three pinned suites: gum (15 scenarios),
  lazygit (21), and helix (21). Each downloads a fixed release verified against
  its published SHA256, so the scheduled matrix stays reproducible.

### Changed

- A command killed by its timeout now fails the step when no assert reads its
  result. Asserting on a killed command still works and still reports the
  timeout as an observable outcome, and `timeout: "0"` still opts a step out of
  every timeout level.
- The spec loader now rejects explicit YAML tags (`!foo`, `!!str`, a bare `!`)
  with an error naming the tag and its line/column, instead of letting them
  reach the decoder. atago's schema is closed and fully typed, so a tag could
  only restate or contradict it, and no spec, example, or doc authored one.
  `!!binary` stays accepted because `atago record` emits it for captures that
  are not valid UTF-8.

### Fixed

- A hanging command no longer passes. A step killed by `run.timeout`,
  `runner.timeout`, `defaults.run.timeout`, `suite.timeout`, or the built-in 60s
  default reported passed whenever the spec did not assert on the result, so the
  bare `run:` step a first-time user writes went green after the timeout fired.
  That is the outcome the timeout defaults exist to prevent, only slower. The
  same false green applied to a `pty` step whose session never waited on the
  program.
- Loading a spec no longer provokes a runtime nil-pointer panic on the way to a
  parse error. A tagged node on a list field made goccy/go-yaml v1.19.2 nil-deref
  inside `decodeSlice`; the loader recovered from it, but `internal/loader` was
  then the only package raising real runtime panics on every test run, and the
  only one whose Windows CI job kept dying with Go GC heap-corruption fatal
  errors (`found pointer to free object`). Rejecting the tags up front takes the
  loader test suite from 9 recovered panics per run to zero.

## [0.12.0] - 2026-07-26

A minor release focused on PTY/TUI correctness, reproducible third-party CI,
and much deeper real-world coverage. atago can now wait on the rendered screen
of an interactive terminal, responds to the terminal capability probes that
full-screen TUIs emit, and ships a substantially expanded yazi suite backed by
pinned installer versions so the same checks stay stable in CI.

### Added

- PTY sessions gain `expect_screen`, a rendered-screen matcher surface for TUIs
  with the usual text/JSON/YAML checks plus optional `timeout` and `stable_for`
  waits. This makes specs assert what a user actually sees on screen instead of
  only matching the raw terminal transcript.
- PTY input/output handling now covers more real terminal behavior: atago can
  answer common terminal capability probes emitted by TUIs and accepts the
  additional named control-key aliases `ctrl-hyphen` / `ctrl-minus`.
- The third-party yazi coverage grows into a dedicated 7-suite, 160-scenario
  corpus covering selection semantics, chooser mode, tab navigation, link and
  overwrite flows, spaced-path handling, entry arguments, and shell-driven file
  mutations.

### Changed

- Third-party CI is now pinned more aggressively: external installers and
  releases used by scheduled suites (including yazi, awscli, and minio legs)
  are version-locked, and CI has guards that fail if those pins drift.
- The README, generated E2E docs, real-world index, and GitHub Pages site now
  describe the PTY screen-assert feature, the expanded yazi coverage, and the
  additional install paths via `aqua` and `mise`.

### Fixed

- PTY/TUI execution is more stable and cheaper under load: rendered-screen waits
  now cache the screen between transcript changes instead of re-rendering on
  every poll, and failure reports reuse the captured screen instead of rendering
  it repeatedly.
- Interactive TUIs that query terminal capabilities no longer misbehave under
  atago because the PTY runner now replies to those probes instead of leaving
  them hanging in the transcript.
- Assertion correctness tightened in a few hot paths: identical JSON values now
  short-circuit cleanly, slice-length mismatches are validated before identity
  fast paths claim equality, and fuzz coverage was extended to keep those cases
  pinned.
- CI and test stability improved across the board: scheduled dogfood/minio
  suites were stabilized, Windows loader validation tests were de-parallelized
  to avoid a Go runtime crash, and the yazi specs were tightened to assert
  distinctive screen states instead of broad/noisy matches.

## [0.11.0] - 2026-07-09

A minor release focused on manifest visibility, documentation discoverability,
and safer shipping paths. `atago manifest` now reports suite lifecycle variable
usage, the docs site is easier to search and browse, release automation now
publishes a multi-arch container image, and several correctness fixes harden
teardown security, unresolved-variable handling, snapshot updates, retry
`changes:` accounting, and `atago run` finalization.

### Added

- `atago manifest` now reports the suite lifecycle's variable references in a new
  `suite_variables` field (mirroring a scenario's `variables`), and the manifest
  schema gains the previously-undocumented `suite_env`, `suite_setup`,
  `suite_teardown`, and `suite_variables` fields (#244).

### Changed

- The README, landing page, and getting-started guide now open with a
  "Try it in 30 seconds" block: a copy-pasteable, zero-install
  `go run github.com/nao1215/atago@latest record … -- git --version` /
  `run` loop against a command everyone already has, producing a green
  `PASSED` in under a minute before any install or fictional binary (#230).
- Website SEO, social, and landing improvements (no tool behavior change): the
  landing page now leads with a visible `<h1>` tagline above a right-sized logo
  (with explicit `width`/`height` on the logo and embedded GIFs to eliminate
  layout shift); every page ships a 1.91:1 (1200x630) social card and a square
  favicon, both derived at build time from the committed logo, plus a
  `twitter:card` and `og:image:width`/`height` tags; the sitemap no longer marks
  the cookbook and real-world pages as priority 0, `robots.txt` advertises the
  sitemap, and the 44 generated real-world pages get search-intent titles
  ("Testing <tool> with atago — executable E2E specs") and coverage-based
  descriptions (#238, #239, #240).
- Website performance and accessibility (no tool behavior change): the heavy
  real-world pages skip off-screen layout via `content-visibility` on code
  blocks and tables, and copy buttons are now created lazily as blocks scroll
  into view (instead of building ~1,900 buttons at load); a skip-to-content link,
  keyboard/touch-visible copy buttons with `:focus-visible` outlines, and real
  `<table>` semantics (horizontal scroll moved to a wrapper so screen readers
  keep announcing the Reference tables) round out an accessibility pass (#237,
  #242).
- Website cookbook and reference navigation (no tool behavior change): on wide
  viewports the "On this page" list becomes a sticky sidebar next to the content
  (the collapsible box stays on narrow screens), and the cookbook gains a live
  recipe filter — typing a keyword hides the non-matching recipe sections and
  their sidebar entries — turning the 63-recipe single scroll into a browsable
  catalog without splitting the committed file or changing any anchor (#241).
- Website full-text search (no tool behavior change): a self-hosted Pagefind
  index is built in CI (pinned binary, SHA-256 verified, no external hosts) and
  a nav search box lazy-loads the search UI and WASM index only when focused, so
  the landing weight is unchanged. Every recipe and spec key becomes directly
  reachable across the cookbook, reference, and real-world pages (#236).
- Release artifacts now include a multi-arch GHCR image for `linux/amd64` and
  `linux/arm64`, published under both the release tag and `latest`, so
  container users no longer depend on a single-architecture image path (#264).

### Fixed

- A suite-level `service:` step's `env` values and `ready` file/port/log probes
  are now counted as variable references by the shared collector, so a
  `suite.setup` service like `env: {DSN: ${...}}` / `ready: {file: ${suitedir}/…}`
  is no longer under-reported by `atago manifest`/`explain` (#244).
- A failing assertion inside `suite.setup`/`suite.teardown` now masks declared
  secrets in its reported `Expected`/`Actual`/`Hint` fields and writes
  `--artifacts-dir` sidecars, matching scenario-level asserts. The suite path
  previously set the check payloads raw, so a secret printed by a setup command
  and echoed into an assert-failure block leaked into the console/JSON reports
  (the leak class #12 fixed for scenarios); and a suite run step with an
  unresolved `${name}` now errors with the same explained diagnostic instead of
  leaking the literal reference into argv (#243).
- A network-policy (`permissions.network.allow`) violation raised by a
  `teardown` step now sets `SecurityViolation` and exits 6, instead of being
  silently swallowed so a spec that contacted a denied host during cleanup
  reported a fully green run. Teardown failures still never change the pass/fail
  verdict — only the security signal is now honored regardless of phase (#248).
- A store- or matrix-provided value that itself contains a `${...}` reference no
  longer leaks that reference verbatim into a no-shell command's argv (or into
  `cwd`). Expansion is single-pass, so such a reference was never expanded; the
  unresolved-reference guard now catches it after substitution and errors with
  the leaked reference named instead of running a garbled command (#249).
- `--update-snapshots` no longer races under `--parallel` when several scenarios
  share one golden file: snapshot writes are now atomic (temp-file + rename), so
  identical-content scenarios update the shared golden deterministically instead
  of failing nondeterministically in a non-atomic remove/create window (#250).
- A `changes:` assert after a retried (`retry.until`) run step now reflects only
  the final, converged attempt's workdir delta rather than the cumulative delta
  of every attempt: the baseline is re-captured before each attempt, so a
  command with a per-attempt side effect reports the net effect the `until` gate
  accepted (#251).
- `atago run` no longer crashes or misreports when post-run finalization sees an
  incomplete result shape. The shared finish path now tolerates partial result
  payloads and still returns the best available status and diagnostics instead
  of panicking late in CLI teardown (#265).

## [0.10.0] - 2026-07-09

### Added

- Console failure blocks now say WHERE and WHY without a `--verbose` detour:
  every `FAILED:` / `ERROR:` / `TEARDOWN FAILED:` header names the spec file
  that produced it (`FAILED: suite / scenario  (specs/convert.atago.yaml)`),
  and an `exit_code` failure appends the failing command's captured stderr
  tail (or its stdout tail when stderr is silent) — the actual error a command
  printed on its way out, previously visible only under `--verbose`.
- A hosted documentation website, https://nao1215.github.io/atago/, built
  with Hugo from the committed docs (no content is duplicated: `doc/cookbook.md`,
  `doc/examples.md`, `doc/real-world.md`, and the generated behavior docs under
  `doc/e2e/` are mounted and turned into pages at build time). Deployed to
  GitHub Pages by `.github/workflows/website.yml` on every push to `main` that
  touches the docs; `make website` / `make website-serve` build it locally.
- 34 new cookbook recipes (23 → 57), all loader-validated by
  `TestCookbook_SnippetsValid` and indexed in [doc/examples.md](doc/examples.md):
  record-first workflows (including `record --pty`), snapshot refresh with
  `scrub:`, TUI screen-frame snapshots, help/version/completion contracts,
  environment and secrets handling, REPL and TTY-detection sessions,
  idempotency and differential oracles, offline API-failure simulation,
  fixture sources (`from:`/`base64:`/`mode:`/`mtime:`/`symlink:`), binary
  safety, boundary inputs, db/ssh/grpc/browser peers, matrix release
  comparison, tags, defaults, and suite layout guidance.
- Website polish: per-page "On this page" TOC, hover heading anchors, a copy
  button on code blocks, active-nav state, card-grid landing links, and zebra
  tables — plus landing sections spotlighting record, snapshot, and PTY/TUI
  strengths.
- A generated "Spec file keys" reference on the website's Reference page:
  every key a spec accepts, rendered from `schema/atago.schema.json` with its
  nesting, type, description, and the atago release that introduced it
  (`website/data/spec_keys.json`, regenerated from release tags by
  `website/tools/gen-spec-keys.py`).
- Descriptions for every property in `schema/atago.schema.json` (58 → all 329
  described), so YAML-language-server hovers and the generated key reference
  document each key, not just the well-known ones.

- Two spec-key inventory drift guards (`TestSpecSchema_StructParity`,
  `TestSpecSchema_SpecKeysComplete`): the key set a spec may contain is written
  down in the internal Go structs, the published JSON Schema, and the website's
  `spec_keys.json` "Since" table, with no mechanical link between them. The
  guards walk all three inventories into the same shape and fail the build on
  any mismatch, so a new struct field without its schema property (which would
  make schema-validating editors reject valid specs) — or a schema change
  without regenerating `spec_keys.json` — can no longer ship silently.

- New spec key `services[].max_log_bytes`: caps how much combined
  stdout/stderr atago retains per service (default 8 MiB), dropping the oldest
  bytes first with an explicit truncation notice. Suite-level services live for
  the whole run and `--parallel` multiplies scenario services, so an unbounded
  capture let one chatty server grow memory without limit; every consumer of
  the capture (readiness excerpts, preserved log artifacts) only ever needs the
  tail. The `ready.log` probe also stopped rescanning the whole capture on
  every idle 20ms tick, and a pending pty `expect` no longer copies the entire
  transcript per poll — both were O(output²) on busy programs.
- Failing scenarios now preserve each mock server's recorded requests as a
  durable artifact next to the service logs (`--artifacts-dir`), so the
  sharpest evidence a mock scenario has — the typo'd client path that 404'd —
  survives the CI job. Green runs and mocks that recorded nothing write no
  file, exactly like service logs.

### Changed

- README now links the hosted cookbook from its Examples section and points to
  the documentation site under the tagline.
- The website's Cookbook and Examples pages are merged into one `/cookbook/`
  page: the by-task index, the recipes, then the runnable per-feature example
  index (the committed `doc/cookbook.md` and `doc/examples.md` stay separate).
- [doc/real-world.md](doc/real-world.md) and every third-party page under
  `/real-world/` now state explicitly that the atago project wrote and runs
  these suites on its own initiative — they are not the upstream projects'
  official test suites, and those projects are not affiliated with atago.
- Under `--ci`, a `--filter`/`--tag`/`--skip-tag` selection that matches no
  scenario now fails the run (exit 3, `ExitConfig`) instead of exiting 0 with
  `PASSED 0 scenarios`, so a typo'd selector can no longer silently disable an
  entire suite in a pipeline. Without `--ci` the same case stays a warning that
  exits 0, leaving interactive workflows untouched.

### Fixed

- Union-shape spec errors (`exit_code`, `stdin`, pty `send`, `json`/`yaml`
  check lists) now carry the same `[line:col]` annotation, source excerpt, and
  caret a typo'd key gets, pointing at the offending value. Previously they
  were bare messages — in a 300-line spec with a dozen `exit_code` asserts,
  "exit_code must be an integer … got \"zero\"" named neither scenario, step,
  nor line.
- `retry.interval` now rejects negative durations at load time like every
  other duration field (a negative interval silently behaved as "no wait").
  All duration validation shares one helper, so bounds and wording can no
  longer drift between fields.

- A step-level `timeout:` on an ssh run step was parsed, validated — and
  silently ignored: the loader whitelists it "because it is honored remotely",
  but the engine only ever applied the runner-level timeout captured at dial.
  The step timeout now bounds the remote command (taking precedence over the
  runner-level value), and an ssh timeout of either level is an observable
  `TimedOut` result naming its source — matching how the cmd runner reports
  local timeouts — instead of a hard scenario error.
- An unset `${env:NAME}` in a `shell: true` run command now fails with the
  same explained error as the shell-less path ("the environment variable NAME
  is not set") instead of reaching the shell, where POSIX sh dies with a
  cryptic "Bad substitution" and cmd.exe forwards the literal text. A bare
  unresolved `${name}` is still left for the shell, which CAN expand it.

- The "no scenarios matched" warning is now selector-aware. `--tag`/`--skip-tag`
  say tags match exactly and point at `atago list`, while `--filter` keeps the
  case-sensitive-substring note. Previously every empty selection was blamed on a
  "case-sensitive substring", which is wrong for tags (they match by equality via
  `==`) and sent users fixing the wrong thing.

### Fixed

- A `screen:` assertion no longer crashes or hangs atago on adversarial
  terminal output. The vt10x emulator panicked on a negative CSI parameter
  (`CSI -10 P` reached slice arithmetic in delete/insert-chars) and spun for
  minutes on an oversized repeat count (`CSI 80111111110 Z` steps one tab stop
  at a time); both shapes — plus variants that hide behind bytes vt10x's
  decoder silently skips — are now defused before the transcript reaches the
  emulator, with a recover backstop so an unknown parser bug fails only the
  assertion, not the process. Found by the new `FuzzRenderScreen` fuzz target,
  which asserts the rendered screen's row/column budgets and UTF-8 validity
  across pathological ANSI input.

## [0.9.0] - 2026-07-08

A second hardening release. Parallel bug hunters and Go-native fuzzing swept the
surfaces that set atago apart — secret masking, snapshot normalization, the
changes-assert glob layer, generated docs, the loader, record, and the
`--rerun-failed` ledger — and every fix landed with a regression test, three of
them first found by new fuzz targets. Declared secrets no longer leak through a
pty step, suite env, or a multi-line value's CRLF variant; a mode-000 file is no
longer dropped from a `changes:` delta; `atago record` round-trips quoted
metacharacters and control-byte commands; and the last-failed ledger no longer
forgets a failure a green run of an unrelated spec did not execute. `atago doc`
also keeps non-ASCII anchors, names env and command gates, and ends with a
newline. A task-oriented cookbook and per-feature/real-world doc indexes round
out the docs.

### Added

- A task-oriented cookbook, [doc/cookbook.md](doc/cookbook.md): 23 copyable
  specs for common jobs — converting images, generating PDFs and files,
  driving interactive prompts and TUIs, JSON/YAML output, database side
  effects, mocking an HTTP API, starting a server, graceful shutdown, output
  trees, golden files, polling, timing bounds, teardown, suite setup, gating,
  matrices, captured variables, and hermetic environments.
  A drift test loads and schema-validates every snippet, and keeps the new
  index in lockstep with examples/ and the cookbook headings. The index,
  [doc/examples.md](doc/examples.md), pairs the cookbook's by-task table with
  the per-feature examples table, which moved there from the README; the
  "Real CLIs tested with atago" tables moved to
  [doc/real-world.md](doc/real-world.md) the same way, with the README keeping
  a summary of the shapes of CLIs covered.
- The README documents installing from the AUR
  ([atago-bin](https://aur.archlinux.org/packages/atago-bin)) on Arch Linux.

### Changed

- The package descriptions GoReleaser ships (Homebrew cask, deb/rpm/apk) were
  shortened to match the repository description.
- The "Why atago?" table names [venom](https://github.com/ovh/venom) for
  multi-protocol platform integration suites, and the README drops bold
  emphasis.

### Fixed

- The last-failed ledger now preserves a recorded failure the run did not
  execute, so a green run of an unrelated spec (or one whose `--filter`/`--tag`
  excludes the failing scenario) no longer clears it and lets the next
  `--rerun-failed` exit 0 while the failure is still real. A fully-green run of
  the same specs still clears the ledger; only scenarios that actually ran are
  re-decided. This is the rule the narrowed `--rerun-failed` path already used,
  now applied to every run.
- Secret masking now collects declared `secrets:` values from every
  env-bearing location — suite.env, defaults.scenario.env, pty steps, suite
  setup/teardown, and scenario teardown — not just run-step, scenario, and
  service env, so a secret injected through any of them no longer leaks into
  reports or `--verbose`.
- A multi-line secret is masked in both line-ending forms, so a value declared
  with LF (a PEM private key) is still masked when the program prints its CRLF
  variant, and the reverse.
- Snapshot normalization no longer writes a secret to a golden when an ANSI
  escape split it in the raw output (the escape strip reassembled it), and is
  idempotent for a run of carriage returns before a newline; a user `scrub`
  rule anchored to line boundaries now also fires on CRLF output.
- A `changes:` entry hints the doublestar `{ }` brace alternation as a glob
  metacharacter, and the suggested escaped spelling is correct for a filename
  that already contains a backslash.
- A workdir file that exists but cannot be read (mode 000) is tracked in the
  `changes:` delta instead of vanishing, so `created: []` no longer passes for
  a step that planted an unreadable file.
- `atago doc` keeps non-ASCII scenario names in their table-of-contents
  anchors (a Japanese name no longer collapses to an empty, colliding
  `#scenario-`), renders env and command skip/only gates rather than only the
  OS, and ends the document with a trailing newline.
- The loader rejects a suite or scenario name containing a newline or tab,
  which corrupted the `atago list` table and split generated doc headings.
- `atago init` rejects more than one path argument instead of silently
  creating the first and ignoring the rest.
- `atago record` generates a valid spec for a command whose quoted argument
  contains a shell metacharacter, for a multi-line or tab-carrying command, and
  for output whose first line begins with `?`.
- `--rerun-failed` stores portable, cwd-relative spec paths in the last-failed
  ledger, so moving the project no longer makes the next rerun silently pass;
  an unknown ledger `schema_version` is rejected rather than misread.
- `--artifacts-dir` pointing at an existing file (or an un-creatable path) is a
  clear config error instead of a run that silently writes no artifacts.

## [0.8.0] - 2026-07-07

A hardening release. A bug-hunting sweep across the surfaces that set atago
apart — assertions, snapshot normalization, report formats, the run engine,
record, and the loader — fixes verified correctness, security, and UX defects,
each landed with a regression test at the layer of the defect and the most
user-visible ones with a self-hosted e2e scenario. A broken symlink is no
longer a false pass for a negative directory assertion, `stdout_to`/`stderr_to`
expand variables, snapshot normalization no longer leaks a raw escape or a
CRLF-folded secret into a golden, `atago record` round-trips non-UTF-8 output,
`--rerun-failed` no longer forgets the failures a filter excluded, and a `pdf`
assertion bounds decompression so a zip bomb in the tested CLI's output cannot
exhaust memory. `atago record --pty` also gains a `--timeout` so it never waits
forever.

### Added

- `atago record --pty` takes a `--timeout` flag (default 30s) that bounds how
  long it waits for the recorded program to exit. Previously the recorder waited
  forever: a server-like process, or an interactive program whose quit keystroke
  was lost to a read-readiness race, would wedge `record --pty` with no output
  and no way out but Ctrl-C. When the timeout elapses atago now kills the child
  process tree (a process-group kill on POSIX, a `taskkill /T` tree kill on
  Windows), still writes whatever transcript was captured, and exits non-zero
  with a clear `did not exit within …` message — mirroring the `pty:` spec
  step's own `timeout:` bound. ([#194](https://github.com/nao1215/atago/issues/194))

### Fixed

- `run` no longer writes redirected output to a literal filename: `stdout_to`
  and `stderr_to` are now `${name}`-expanded like the assertion paths that read
  them, so a per-matrix-row or store-keyed target such as `out-${who}.txt`
  resolves instead of failing the following assert with a spurious "no such
  file".
- `dir` `contains`/`not_contains` judged membership by following the symlink, so
  a dangling symlink was reported absent — a dangerous false pass for
  `not_contains`. Membership now reflects the directory entry (`Lstat`), matching
  the recursive and glob paths.
- `atago record` output containing non-UTF-8 bytes (binary, Latin-1, Shift-JIS)
  now round-trips: the generated spec re-parsed lossily and failed on its first
  replay. The recorder emits a `!!binary` scalar for such lines so `record` →
  `run` is green.
- `snapshot` normalization is idempotent again and no longer leaks secrets. A
  stray escape spliced by OSC removal left a raw ANSI byte in the golden, and a
  multi-line secret whose value uses LF reappeared in the golden when the output
  used CRLF; both are now closed.
- `--rerun-failed` combined with `--filter`/`--tag`/`--skip-tag` silently dropped
  the still-failing scenarios it did not run from the ledger, greenlighting a
  later rerun; those failures are now preserved, the "renamed or removed?"
  warning is no longer shown for a filter exclusion, and an equivalent
  absolute/relative spec path now matches the recorded ledger.
- `pdf` assertions bound FlateDecode decompression at 64 MiB so a decompression
  bomb in the tested CLI's output can no longer exhaust memory, and PDF
  hex-string (`<…>`) metadata and text are now decoded instead of reported as a
  missing field.
- A byte-exact `file equals` failure that differs only in line endings now
  renders `(only the line endings differ: CRLF vs LF)` instead of a blank diff.
- The GitHub Actions `::notice::` summary counts a suite that errored before any
  scenario ran, matching the junit and tap counts and its own `::error::`
  annotation; a `suite.setup` failure is labeled `suite setup`, not
  `service setup`.
- `loader` rejects a `not_matches` pattern that also matches the empty string
  (it can never pass), and explains a matrix name-template collapse by naming the
  omitted row variable instead of a bare "duplicate scenario name".
- A negative `--parallel` is rejected with a config error like `--repeat` and
  `--retry-failed`, and `--repeat 1` no longer trips the `--retry-failed`
  mutual-exclusion guard (repeat activates only at `> 1`).

### Changed

- An unresolved `${var}` in a run step's `cwd` now fails with an explained error
  naming the reference, instead of the misleading "executable not found" the
  child raised when it could not start in a literal `${var}` directory. Bare
  `${…}` in `env` values and `stdin` keeps its documented literal passthrough.
- Explicit `--help` on `explain`, `doc`, `list`, `manifest`, and `init` prints to
  stdout, matching `completion` and the top-level help, so the usage text can be
  piped.

## [0.7.0] - 2026-07-07

Windows completes its interactive-terminal story. `pty:` steps and
`atago record --pty` now drive a ConPTY pseudo-console on Windows, so the
interactive specs that were gated `skip: {os: windows}` run everywhere, and a
session recorded on Windows produces a spec that also replays on Windows. A
background `service:` is now torn down with its whole process tree on Windows,
not just its leader. The ConPTY layer is implemented in-repo on
`golang.org/x/sys/windows` with no third-party dependency, and the POSIX pty
path is unchanged. Only `signal:` remains POSIX-only — Windows has no POSIX
signals to deliver.

### Added

- `pty:` steps run on Windows through a ConPTY pseudo-console (Windows 10 1809
  or later), not just on POSIX. The expect/send session, named keys, and
  rendered-screen asserts all work; the loader already accepted the step on
  every platform, and the engine now drives it instead of returning an
  "unsupported" error. The POSIX runner keeps using creack/pty unchanged — only
  the platform-neutral expect/send loop is shared between the two. See
  [examples/pty_portable.atago.yaml](examples/pty_portable.atago.yaml).
- `atago record --pty` records an interactive session on Windows over a ConPTY,
  and a session recorded there generates a spec whose `pty:` step also replays
  on Windows. The ConPTY wrapper is shared with the pty runner in a new
  `internal/conpty` package, built directly on `golang.org/x/sys/windows` (no
  new dependency). One difference from POSIX: a ConPTY exposes no echo state, so
  a typed password is not auto-masked on Windows — convert a secret send to a
  `${env:...}` placeholder by hand.

### Fixed

- A background `service:` on Windows is now torn down with its whole process
  tree, not just its leader. A service launched with `shell: true` that forked
  further children would orphan them on teardown; teardown now uses
  `taskkill /T` — a race-free kill of the live process tree — backed by a
  kill-on-close job object as crash-insurance, the Windows analog of the POSIX
  runner's process-group kill.

## [0.6.1] - 2026-07-07

A robustness patch: seven fixes, each landed test-first. Two panics on malformed
input — in the YAML loader and the JSON parser — were surfaced by fuzzing and now
yield clean errors. The rest close cross-platform gaps that only bit
on Windows: CRLF folding across every stream text matcher, newline tokenization
for block-scalar commands, nested redirect directories, and POSIX signal exit
codes, plus a command-name fix in a snapshot-update error. The self-hosted E2E
suite grew alongside them.

### Fixed

- A `json` assertion or a `store` capture no longer crashes on malformed JSON
  from the program under test. The third-party JSON parser (ojg) panics with
  "assignment to entry in nil map" on some inputs, and the panic propagated out
  and aborted atago. Both the matcher and the store capture now recover from a
  parser panic and report invalid JSON, so a broken payload is a normal
  assertion failure or capture error, not a crash. Found by FuzzValuesEqual.

- Loading a spec no longer crashes on malformed YAML. Some inputs (a bare `!`
  tag over a broken mapping) made the underlying goccy/go-yaml decoder panic
  with a nil pointer dereference, which propagated out of the loader and aborted
  atago. Loading now recovers from a decoder panic and reports a clean parse
  error, honoring the contract that loading untrusted spec bytes never crashes.
  Found by FuzzLoadBytes; the crasher is kept as a regression seed.

- An error from `atago snapshot update` now names that command instead of the
  `atago run` it delegates to internally. A missing target or a bad flag printed
  `atago run: ...` even though the user typed `snapshot update`; the diagnostic
  prefix now matches the invoked command.

- A no-shell command (`shell: false`, the default) authored as a YAML block
  scalar now tokenizes to the same argv on every OS. `windowsFields` split only
  on spaces and tabs while POSIX (`go-shellwords`) also splits on carriage
  returns and newlines, so a command written with `>` (a trailing newline) or
  `|` (interior newlines) — or authored in a CRLF file — glued a stray `\r`/`\n`
  onto an argument on Windows alone, silently breaking a cross-platform suite on
  `windows-latest`. Windows now treats carriage return and newline as field
  separators too, completing the tokenization parity begun in #154.
- Every stdout/stderr text matcher folds CRLF, matching the `equals` matcher.
  `contains`, `not_contains`, `matches`, and `not_matches` compared raw bytes, so
  a multi-line `contains` needle or a `(?m)`-anchored `matches` written with LF
  passed against POSIX output but failed against cmd.exe's CRLF output on
  Windows, while `equals` on the same output passed everywhere. Line endings are
  an OS artifact for all stream text matchers now; byte-exact comparison
  (including the exact line ending) stays with the file matchers (`equals_file`,
  `dir` `sha256`).
- `stdout_to` / `stderr_to` create the parent directory of a nested redirect
  target, mirroring the fixture writer. A redirect to `logs/out.txt` failed with
  a raw "no such file or directory" when `logs/` did not exist yet, while a
  fixture at the same path created it; the two path-taking features now behave
  the same. The parent stays inside the scenario workdir.
- A run command terminated by a POSIX signal reports the 128+signal exit code
  (130 for SIGINT, 137 for SIGKILL, 143 for SIGTERM) instead of a bare `-1`, so a
  spec can assert a CLI's signal-handling contract. Go's `ExitCode()` collapses
  every signal to `-1`, which is also indistinguishable from atago's
  timeout/cancel sentinel; the timeout and parent-cancel paths are resolved
  before this point, so a signal here is the program's own termination. Windows
  has no POSIX signals and is unaffected.

## [0.6.0] - 2026-07-07

An assertion-and-capture ergonomics release: byte-exact file round-trips,
regex-free whole-content capture, multi-check json/yaml matchers, and a
cross-platform tokenization fix. Every change landed test-first and expanded the
Windows-portable self-hosted E2E subset.

### Added

- Byte-exact file content matchers for round-trip and idempotence tests (#155).
  `file: { path: out.hex, equals_file: in.hex }` compares two runtime-produced
  files byte-for-byte, and `file: { path: out.hex, equals: "<literal>" }`
  compares a file's bytes to an inline literal. Neither normalizes CRLF or a
  trailing newline (matching the `dir` snapshot hashing semantics), so the
  natural "these two files the run just produced are equal" assertion — which a
  `shell: true` POSIX `cmp` cannot express portably — is now declarative. Both
  paths are confined to the workdir. See
  [examples/files_and_fixtures.atago.yaml](examples/files_and_fixtures.atago.yaml).
- `store` can capture a whole stream or file without a regex (#158). A stream
  source takes `trim` (`{ stdout: { trim: true } }` grabs the whole stdout and
  strips the trailing newline; `trim: false` keeps bytes verbatim) and a file
  source takes `text: true` (`{ file: { path: out.bin, text: true } }` captures
  the whole file verbatim). Previously the only way to capture an opaque value
  (a token, an id, a signed blob) was to invent a regex that matched the whole
  thing. Exactly one selector per source is still enforced. See
  [examples/store_and_variables.atago.yaml](examples/store_and_variables.atago.yaml).
- `json:` and `yaml:` matchers accept a list of checks (#156). A single response
  can now assert several JSONPaths in one block —
  `json: [ { path: "$[0].name", equals: … }, { path: "$[0].default", equals: … } ]` —
  where each entry is an independent check and all must hold, instead of
  repeating the whole `assert:` block once per path. The single-mapping form is
  unchanged and still valid. See
  [examples/json_and_yaml.atago.yaml](examples/json_and_yaml.atago.yaml).

### Fixed

- A no-shell command (`shell: false`, the default) now tokenizes to the same
  argv on every OS (#154). `windowsFields` treated a single quote as an ordinary
  character while POSIX (`go-shellwords`) grouped and stripped it, so a spec
  passing single-quoted inline JSON — `mycli '{"k":"v"}'` — reached the CLI as
  valid JSON on Linux/macOS but as `'{k:v}'` on Windows, silently breaking a
  cross-platform suite on `windows-latest` alone. Windows now groups single
  quotes too (keeping backslashes literal so `C:\` paths are unaffected), and an
  unmatched single quote is an error like an unmatched double quote.

### Docs

- The README names the real fixture inline-source key `content:` (and
  `base64:`/`from:`/`symlink:`) instead of describing it as "text", so a reader
  can no longer guess a non-existent `text:` key and hit an unknown-field error
  (#157).

## [0.5.1] - 2026-07-06

A security and robustness patch, each fix landed with a reproduction test first.

### Security

- Path containment is no longer purely lexical. `ResolveWorkdirPath` /
  `ResolveSpecPath` validated a spec-declared path with `Join`/`Clean`/`Rel`
  only, so a symlink planted **inside** the workdir or spec directory by the
  program under test passed the check and the feature then followed the link
  during I/O — reading or overwriting an arbitrary host file. Only `fixture`
  self-defended, so the guard was inconsistent across features. The no-follow
  guard is now centralized in the `security` package (`ReadFileNoFollow` /
  `WriteFileNoFollow`) and enforced at every path-taking read/write site: file
  assertions, `run.stdout_to` / `stderr_to`, `http.body_to`, the cdp
  `screenshot` output, and snapshot `update`/`assert` (#16).

### Fixed

- `ready.timeout: "0"` now behaves as the documented unbounded wait for every
  readiness probe. The `delay` probe already honored it, but the `file`/`port`/
  `log` probes handed `0` to a zero-duration timer and failed on the first tick.
  A non-positive timeout is now treated as unbounded (bounded only by the process
  staying alive or the scenario context), matching the delay probe.

## [0.5.0] - 2026-07-06

A correctness pass across `record`, `snapshot`, the report formats, the loader,
the assertion matchers, secret masking, and the run engine. Each defect was
surfaced by an edge-case bug hunt and fixed with a reproduction test first, plus
two flaky-test tooling improvements described below.

### Added

- Stream matchers now compose. `contains`, `not_contains`, `matches`, and
  `not_matches` can be set together on one `stdout`/`stderr`/`body` block and all
  have to hold — the common "output has X but not Y" shape, previously two
  separate `assert:` steps. The whole-stream matchers (`equals`, `empty`,
  `snapshot`, `json`, `yaml`) are still used alone; mixing one in with a text
  matcher is a load error. See [examples/run_and_assert.atago.yaml](examples/run_and_assert.atago.yaml).
- `scrub:` (#137) — spec-wide declarative output normalization for snapshots. Each rule
  rewrites every regex match in captured output to a literal placeholder
  (`{pattern: 'id=\d+', placeholder: 'id=<ID>'}`) before a snapshot is compared
  or written. Where the built-in normalizers fold a fixed set of volatile forms
  (ANSI, UUID, timestamp, port, path) and `secrets:` masks known values, `scrub:`
  handles the open set of volatile patterns only the author knows about —
  auto-increment IDs, request identifiers, custom timestamps — so a snapshot that
  flaked every run becomes deterministic. Rules apply after secret masking and
  before the built-in normalization. See [examples/scrub.atago.yaml](examples/scrub.atago.yaml).

### Changed

- `--repeat` (#138) now distinguishes an unstable scenario from a broken one. A run
  where some iterations passed and some failed folds to `flaky` (green for the
  exit code, surfaced with its flake rate, e.g. `2/10 iterations failed`) instead
  of a hard `failed`; a scenario that failed *every* iteration stays a
  deterministic `failed`. Previously any failing iteration collapsed the whole
  fold to `failed`, so "3/10 flaked" was indistinguishable from "10/10 is a real
  bug" — the exact signal `--repeat` exists to surface. This matches how
  `--retry-failed` already reports a recovery as `flaky`.

### Fixed

- `record --pty` now anchors its generated `expect`/`stdout: contains` on text
  that is a verbatim substring of the raw transcript. It stripped ANSI from the
  visible line and joined the result, so a colored prompt produced an anchor with
  mid-line color codes removed that never matched the raw pty stdout — the
  generated spec failed on replay. It now anchors on the longest color-free run
  of the last visible line.
- `record --pty` no longer explodes a typed password into one
  `${env:ATAGO_SECRET_n}` placeholder per character. Raw-mode capture delivers
  one keystroke per read, and consecutive input bursts were not coalesced, so a
  six-character password generated seven single-character secret sends.
  Consecutive input bursts with the same echo state are now merged.
- A directory-tree snapshot no longer reports a false match when a symlink name
  or target contains the ` -> ` that joins the two in the manifest. `>` is now
  escaped inside a manifest field, so a symlink named `a -> b` pointing at `c` no
  longer collides with one named `a` pointing at `b -> c`.
- Snapshot normalization no longer corrupts a path that merely shares a prefix
  with the home or scratch directory. The home dir `/home/nao` turned
  `/home/naoki` into `~ki`, a container `HOME=/` turned every absolute path into
  tildes, and the workdir `/tmp/run1` turned `/tmp/run10` into `<workdir>0`.
  Masking is now path-component aware and skips a root home.
- A unified diff now numbers a zero-line hunk side `0`, per the GNU convention
  patch(1) relies on: a pure insertion emitted `@@ -1,0 +1,2 @@` instead of
  `@@ -0,0 +1,2 @@`, and an empty side no longer carries a spurious
  "No newline at end of file" marker.
- A `--report tap` diagnostic message no longer injects a stray backslash before
  a `#`. The message was escaped for the bare `ok`/`not ok` line and then quoted,
  so `issue #42` reached the consumer as `issue \#42`; `#` is an ordinary
  character inside the quoted YAML scalar and is left alone.
- A `--report junit` `time` attribute is now a plain decimal. A sub-millisecond
  duration serialized as `1.5e-06`, which the JUnit/Surefire schema and strict
  parsers reject.
- The console summary headline duration now reports the run's real wall-clock
  time. It summed each suite's own duration, so `--parallel` running four
  one-second suites concurrently printed `(4s)` for a run that took about one.
- `--fail-fast` now stops scheduling across spec files, not only within a single
  suite. A failing first spec still let every later spec run to completion.
- `--tag` and `--skip-tag` are now repeatable and OR their values, like
  `--filter`. `--tag a --tag b` silently kept only `b`.
- `json:`/`yaml:` `gt`/`gte`/`lt`/`lte` now compare integers beyond 2^53 exactly,
  like `equals` already does. The comparison rounded the selected value through
  float64, so `9007199254740993` was reported not greater than
  `9007199254740992`.
- `json:`/`yaml:` `equals` no longer reports a JSON boolean as equal to the
  string spelling of its value. A fallback compared the two by their printed
  form, so `true` matched `equals: "true"`.
- Secret masking no longer leaks part of an overlapping secret. Masking one
  secret consumed bytes a second, overlapping secret needed, so its tail survived
  in cleartext even though it appeared verbatim; every occurrence is now masked
  in one pass over the original text.
- The loader now rejects a `mock_server` step placed in a scenario's `steps` or
  `teardown`. It is a `suite.setup`-only action, like `service`, but was silently
  accepted and never started.
- The loader now rejects an unknown key in the `exit_code` mapping form
  (`exit_code: {not: 0, bogus: 5}`). The `{not}`/`{in}` decode bypassed the
  document-wide strict decoding, silently dropping the typo.
- The loader now validates a `dir.glob` pattern at load time, like the sibling
  `dir.ignore` and `changes` globs, instead of only failing when the assertion
  runs.
- The loader now rejects a negative `timeout`/`delay` duration on the suite, run,
  defaults, runner, service readiness, and mock-route fields. A negative
  wall-clock bound yields an already-expired context that kills the step
  immediately — the same rule the pty and signal timeouts already enforce.

## [0.4.2] - 2026-07-05

A correctness pass over three surfaces — the JSON/YAML numeric matchers, the
loader's assertion validation, and the machine-readable report formats — each
defect surfaced by an edge-case bug hunt and fixed with a reproduction test
first. No new features.

Highlights: `json` `equals` no longer collapses two distinct 64-bit integers,
and `gt`/`lt` compare integers beyond int64; a whole-number float like
`1000000.0` now matches `^1000000$`; the loader rejects an empty
`matches`/`not_matches` regexp and negative `length`/`duration` bounds at load
time instead of at run time; and a captured ANSI escape can no longer break a
`--report tap` diagnostic or a `--report gha` annotation.

### Fixed

- `json:`/`yaml:` `equals` no longer reports two distinct large integers as
  equal. The comparison rounded both operands through float64, so a 64-bit id
  like `9007199254740993` matched `9007199254740992`. Integer operands are now
  compared at full precision.
- `json:`/`yaml:` `gt`/`gte`/`lt`/`lte` now compare integers larger than int64.
  Such a number decodes as a JSON number the numeric path did not recognize, so
  a valid large value failed with "not numeric, so it cannot be compared".
- `json:`/`yaml:` `matches:` now sees a whole-number float as its digits. A
  value like `1000000.0` rendered as `1e+06`, so a pattern written against the
  digits (`^1000000$`) never matched; it now renders as `1000000`, and the
  `equals` failure message reads the same way.
- The loader now rejects an empty `matches:`/`not_matches:` pattern on the
  `stdout`/`stderr`/`body`, `json:`/`yaml:`, and `header:` matchers. An empty
  regexp matches everything, so `matches: ""` is an always-true no-op and
  `not_matches: ""` can never pass — mirroring the existing empty-string
  `contains`/`not_contains` rejection.
- The loader now rejects a negative `json:`/`yaml:` `length:` and a negative
  `duration:` bound. No array, object, or string has a negative length, and a
  measured wall-clock duration is never below zero, so either bound could only
  ever be an authoring mistake.
- The TAP (`--report tap`) failure diagnostic stays valid YAML. A captured
  control byte from a command's output — most often an ANSI color escape —
  landed verbatim in the diagnostic's `data:` block, so a TAP consumer parsing
  it failed with "control characters are not allowed". Such bytes are now folded
  to the Unicode replacement character, keeping tab and newline, as junit
  already does through its XML encoder; the GitHub Actions (`--report gha`)
  annotations are cleaned the same way.

## [0.4.1] - 2026-07-05

A correctness pass over five independent surfaces — record, the equals matcher,
the rerun ledger, tree snapshots, and the loader — each surfaced by a systematic
bug hunt and fixed with a reproduction test first. No new features.

Highlights: `atago record -- pwd` (and any command printing a path under its
workdir) round-trips green again; `stdout equals` no longer ignores trailing
blank lines; a narrowed `--rerun-failed` no longer forgets still-failing specs
it did not run; a `dir:` tree snapshot can no longer be fooled by a newline in a
filename; and a spec saved with a UTF-8 BOM loads instead of failing with a
misleading unknown-field error.

### Fixed

- A spec file saved with a leading UTF-8 byte-order mark now loads. Windows and
  Notepad-family editors emit a BOM routinely, and the YAML decoder glued it onto
  the first key, so a correctly-authored spec failed with a confusing
  `unknown field "version"` that blamed a field the author wrote right. A single
  leading BOM is now stripped before parsing, as most YAML tooling does.
- A `dir:` tree `snapshot:` manifest now escapes control bytes in entry paths and
  link targets, so a filesystem name embedding a newline (legal on POSIX) can no
  longer masquerade as several manifest lines. Before, a single entry named
  `a<newline>dir b` produced the same manifest text as a structurally different
  two-entry tree and falsely matched its golden. Ordinary names — with no
  backslash, CR, or LF — render unchanged, so existing goldens are unaffected
  (#25).
- `atago run --rerun-failed` narrowed to a subset of the recorded specs no longer
  drops the recorded failures in the specs it did not run. It rewrote the whole
  `.atago/last-failed.json` ledger from only the scenarios that ran, so
  `--rerun-failed a.atago.yaml` forgot a still-broken `b.atago.yaml` — and if the
  narrowed target now passed, the ledger was wiped and the run exited green while
  other work stayed broken. Recorded failures for specs outside the run's target
  are now carried back into the saved ledger, since they were never re-verified;
  a full-scope rerun and a normal run are unchanged (#64).
- `stdout`/`stderr` `equals` and `not_equals` now tolerate only a single
  trailing newline, not an arbitrary run of trailing blank lines. The matcher
  trimmed every trailing newline, so `equals: "hello"` passed against output
  that was `hello` followed by extra blank lines, and `not_equals` reported a
  false "equal" for the same output — an exact-text assertion silently ignoring
  real trailing content. It now drops at most one trailing newline per side,
  consistent with the `line:` selector, which already keeps a deliberate
  trailing blank line addressable.
- `atago record` no longer pins its own scratch directory into the generated
  stdout anchor. A command that prints an absolute path under its workdir (the
  canonical case is `atago record -- pwd`) recorded a `contains:` matcher on the
  dead `/tmp/atago-record-NNN` scratch path, which the replay's own isolated
  workdir could never match — so the very spec `record` produced failed on its
  first `run`. The record-time workdir is now rewritten to the built-in
  `${workdir}` reference, which expands to the replay workdir, matching the
  masking `record --snapshot` already applied. The record→run round-trip is
  green again (#30).

## [0.4.0] - 2026-07-05

Runner-selection and reporting polish surfaced while migrating real E2E suites
(gup) to atago, plus editor-completion discoverability for the spec DSL.

Highlights: `--filter` now selects multiple scenarios with OR semantics (comma
list or repeated flag) instead of silently honoring only the last; a run that
drops a spec to a load error now reads `FAILED` and counts the drop instead of a
misleading `PASSED ... 0 errored`; and `atago init`/`atago record` emit a
resolvable `# yaml-language-server: $schema=...` header so scaffolded specs get
editor completion out of the box.

### Added

- `atago run --filter` accepts multiple name substrings with OR semantics,
  mirroring `--tag`: a comma list (`--filter a,b`) or a repeated flag
  (`--filter a --filter b`) runs every scenario whose name contains any term. A
  single substring is unchanged, and the "no scenarios matched" warning still
  fires when none match. Previously a comma list was one literal substring and
  repeated flags silently kept only the last — a green run could hide that half
  the selection never ran (#119).
- `atago init` and `atago record` emit a resolvable
  `# yaml-language-server: $schema=<url>` header as the first line of every
  generated spec, so a freshly scaffolded spec gets editor completion for step
  types, matchers, and `${...}` expansion forms out of the box. The URL pins to
  a tagged binary's own release tag and otherwise resolves against `main`; the
  README editor-support snippet now uses the same absolute URL instead of a
  repo-relative path that only resolved inside the atago repo (#121).

### Fixed

- `atago run` over a directory that mixes loadable specs with a spec that fails
  to load now prints a summary that reads `FAILED` (matching the non-zero exit
  code) and counts the dropped file (`N spec(s) failed to load`), instead of a
  misleading `PASSED ... 0 errored` that silently omitted it. A fully-valid run
  is unchanged (#120).
- PDF text extraction now decodes octal `\ddd`, `\b`, `\f`, and backslash-newline
  line-continuation escapes in literal strings (ISO 32000), so a `pdf: {text:
  {...}}` assertion matches non-ASCII text as written by pandoc/LaTeX/wkhtmltopdf
  instead of seeing the literal escape sequence (e.g. `caf\351` now yields the
  byte for `é`, not `caf351`).
- The `screen` assertion's failure box measures line width and padding in runes,
  not bytes, so a rendered pty/TUI screen with box-drawing characters or CJK text
  frames with an aligned right border instead of a ragged one.
- A store step capturing a JSON `null` (`from.stdout.json`) now returns a clear
  error instead of storing Go's `"<nil>"` string into the variable, so a null
  field can no longer masquerade as a captured value.
- Browser `cdp` upload and download actions now expand `${name}` references in
  their file path, capture directory, and selectors, matching every other cdp
  action; a stored or `${workdir}`-derived path previously reached the browser
  literally.
- `atago explain` and `atago doc` now render an HTTP header `matches:` regexp
  matcher; it was silently dropped, hiding a security-relevant header constraint
  from the generated summaries.
- The shared variable walk (`explain`/`manifest`/`doc`) now counts `${name}`
  references in `fixture.from`, HTTP/gRPC header values and JSON bodies, assert
  matcher arguments, a service `ready` probe, and a store file-source path — every
  field the engine expands — so the "Variables used" summary no longer
  under-reports them.

## [0.3.4] - 2026-07-05

A bug-fix release that bounds and guards the run engine. No new features.

Highlights: `defaults.run.timeout` now bounds `http`, `query`, and `grpc` steps,
not just `run` steps; a `skip.command` / `only.command` probe and a service
`ready.delay` are both time-bounded so a hanging probe or an over-long delay
fails fast instead of stalling the whole run; and the unresolved-variable guard
now fires for a named `cmd` runner, so a typo cannot leak `${...}` into argv and
run a garbled command.

### Fixed

- The unresolved-variable guard now fires for a named `cmd` runner, not only the
  default (unnamed) local runner. A no-shell run executes its command as argv, so
  a typo like `run: {runner: local, command: "echo ${typo}"}` leaked the literal
  `${typo}` into argv and ran a garbled command instead of erroring with the
  reference named. Any local (non-ssh) run is now guarded; only an ssh runner,
  where a remote shell may expand it, stays exempt.
- A service `ready.delay` is now bounded by `ready.timeout`, as the docs promise
  ("Timeout bounds the readiness wait"). The delay branch waited the full
  duration regardless of the timeout, so `delay: 30s` with `timeout: 500ms`
  stalled for 30 seconds instead of failing fast — a CI-hang hazard. A delay
  longer than the timeout now fails at the timeout with a message naming the
  misconfiguration; a delay within the timeout is unchanged.
- A `skip.command` / `only.command` probe is now time-bounded (30s). The probe
  ran unbounded, so a hanging probe (`sleep 9999`, an unreachable health check)
  stalled the sequential selection phase and with it the whole run. A probe that
  exceeds the bound is treated as "did not succeed".
- `defaults.run.timeout` now bounds `http`, `query`, and `grpc` steps, not just
  `run` steps. The precedence chain (step > runner > defaults.run > suite >
  built-in 60s) documents these steps as members, but the http and connection
  config builders passed an empty defaults.run level, so a spec relying on
  `defaults.run.timeout` to bound a slow request or query silently fell through
  to `suite.timeout` or the 60s default.

## [0.3.2] - 2026-07-05

A bug-fix release hardening the runners and the reporters. No new features.

Highlights: a `query`/`db` SELECT preceded by a SQL comment no longer misroutes
to ExecContext and drops its rows; `atago record` escapes only genuine variable
references, so a recorded spec whose output carries a literal `${1}` replays
green; and a hung `grpc` server no longer passes on a per-call timeout. Reporting
is steadier too — a suite that errors in `suite.setup` with nothing selected is
no longer a green empty junit/tap/gha result, and TAP marks a recovered flaky
scenario `ok`.

### Fixed

- A `query`/`db` step whose SELECT is preceded by a SQL comment (a `-- ...` line
  or a `/* ... */` block) no longer misroutes to ExecContext and returns no rows.
  The statement classifier read the leading verb without stripping comments, so a
  commented SELECT looked like a non-row statement and its result set was lost.
  Comments outside string literals are now stripped before classification; a
  comment marker inside a quoted value is still treated as data.
- `atago record` no longer generates a spec that cannot replay when the observed
  command or output carries a `${` not followed by a valid variable name (a tool
  that prints `${1}`, `${}`, or `${ }`). The recorder escaped every `${` to
  `$${` unconditionally, but the replay expander only restores `$${<valid-name>}`
  — so `$${1}` stayed literal and the generated `contains`/`command`/file-path
  matcher could never match the real `${1}`. Escaping now mirrors the expander
  exactly (only genuine references are escaped), so the recorded spec replays
  green. The same fix covers `atago record --pty` transcripts and sends.
- A `grpc` step against a hung server no longer passes as a normal
  `Result{grpc_status: 4}` when its per-call timeout fires. The timeout was
  detected only through `ctx.Err()`, which a server enforcing the propagated
  deadline could beat: its `DeadlineExceeded` status arrived over the wire, and
  the call returned, before the local deadline timer marked the context done, so
  the timed-out call was recorded as a captured status instead of a hard error.
  A call is now treated as a timeout whenever the per-call deadline has already
  elapsed, closing the race; a live deadline still records a real status as a
  Result. This also fixes a flaky `TestInvoke_CallTimeoutIsError` on CI.
- Loading an empty spec file (empty, whitespace-only, or comments only) now
  reports `spec is empty: expected a YAML document with version, suite, and
  scenarios` instead of the bare decoder `EOF`, which named neither the file's
  problem nor what a spec needs.
- A suite that fails in `suite.setup` with no scenario selected to run (every
  scenario filtered out by `--filter`/`--tag`, or an empty scenario list) is no
  longer rendered as a green, empty result by the junit, tap, and gha reports,
  and the console verdict now reads FAILED. The run already exits non-zero and
  the console/json output showed the failure, but junit emitted `tests="0"`, tap
  a bare `1..0` plan, and gha no error annotation, so a CI step gating on those
  reports read the errored suite as passing. Each format now surfaces the setup
  failure as a failing entry.
- The TAP report now emits a passing `ok` point for a flaky scenario (one that
  failed and then passed under `--retry-failed`), matching the exit code and the
  console, gha, and junit reports, which all treat a recovered scenario as green.
  TAP previously fell through to `not ok`, so a CI step consuming the TAP stream
  read the run as failed even though atago exited 0; the recovery now stays
  visible in the point's diagnostic instead of flipping the verdict.

## [0.3.1] - 2026-07-05

A bug-fix release — correctness and UX hardening across the assertion engine,
the loader, and every runner (command, pty/TUI, HTTP, gRPC, database, SSH,
background services), plus the record/snapshot/report tooling. No new features.

Highlights: two security fixes (an HTTP redirect could follow an allowed host
onto a denied one, bypassing `permissions.network.allow`; a `mode`/`mtime`-only
fixture could follow a planted symlink and re-permission a file outside the
workdir), and pty/TUI fixes that make real full-screen programs testable at last
(a usable default `TERM`, and an `expect` that no longer matches a stale earlier
prompt and races the session ahead).

### Fixed

- A `pty` step now exports `TERM=xterm-256color` by default (overridable via
  `env: {TERM: ...}`). Without a sane TERM, full-screen TUIs (less, vim, htop)
  print "terminal is not fully functional" and refuse to draw, so a pty/screen
  assertion never saw the real UI; the default matches the vt10x screen emulator
  and is deterministic regardless of the host's own TERM.
- A `pty` `expect` now scans only the transcript AFTER the previous match, so a
  recurring pattern (any shell prompt) waits for its NEXT occurrence instead of
  matching the stale earlier one. Previously every expect re-scanned the whole
  transcript, so a second `expect "PROMPT> "` matched instantly and the session
  raced ahead — a false pass for exactly the interactive flows the feature
  exists for.
- A `pty` step now surfaces a parent-context cancellation (Ctrl-C / suite
  cancel) as an execution error instead of a spurious expect failure or a benign
  timeout, so the scenario stops rather than asserting against a killed terminal
  (matching the command runner, #30).
- The gRPC runner treats a call that times out or drops (our per-call deadline
  fired) as a transport error, not a passing result. `status.FromError` maps a
  client-deadline `DeadlineExceeded`/`Unavailable` to `ok=true`, which was
  recorded as a normal Result and passed against a hung or unreachable server
  unless the spec happened to assert `grpc_status`.
- `grpc` `method` accepts the fully-qualified `/pkg.Service/Method` form (a
  single leading slash) and rejects a method with more than one internal slash,
  instead of silently mis-parsing `/pkg.Service/Method` as service
  `/pkg.Service` and failing later with a confusing reflection error.
- A `service` `ready.delay` probe now fails if the process exits during the
  delay window, instead of reporting the (dead) service ready and running the
  scenario against a dead peer — matching the file/port/log probes.
- A host-less `service` `ready.port` (e.g. `9997` instead of `:9997` /
  `127.0.0.1:9997`) now fails fast with a clear message instead of swallowing
  "missing port in address" and running to the full readiness timeout.
- The SSH runner no longer mangles a bracketed IPv6 host: `[::1]` resolves to
  `[::1]:22` rather than the malformed `[[::1]]:22`, and a trailing-colon host
  (`host:`) gains the default port instead of dialing an empty one.
- `atago run --rerun-failed` no longer reports a false green — and no longer
  deletes the recorded failures — when the recorded scenario names no longer
  match the specs (renamed or removed while still broken). It warns, keeps the
  state, and exits non-zero so the still-failing work is not silently forgotten.
- `atago doc` rejects `--out` combined with `--split-by-spec` instead of
  silently ignoring `--out` (the split branch only writes into `--out-dir`).

- The HTTP runner now re-enforces the network allowlist on every redirect hop.
  An allowed host could 3xx-redirect the client onto a denied host and the
  default client would follow it, silently defeating
  `permissions.network.allow` — the redirect target is now policy-checked and a
  denied hop fails with the same violation as a direct request.
- The db runner no longer misroutes a statement whose string literal contains the
  word `RETURNING` (e.g. `INSERT INTO logs (msg) VALUES ('order RETURNING to
  sender')`). The RETURNING check now skips quoted regions, so such an
  INSERT/UPDATE/DELETE runs through Exec and its affected-row count is captured
  instead of being lost.
- `atago record` no longer corrupts a spec when recorded output/command/path
  carries a control byte. A raw tab was spliced into a plain YAML scalar and
  silently stripped on reparse (so a recorded tab-separated line could never
  match), and a newline produced a block scalar that aborted `record` with an
  "atago bug" error; such values are now emitted as escaped double-quoted
  scalars that round-trip exactly.
- A `fixture` that sets only `mode`/`mtime` no longer follows a symlink the
  program under test may have planted at the destination: `chmod`/`chtimes`
  follow symlinks, so this could re-permission a file outside the workdir. An
  existing symlink at the target is now refused, matching the content path.
- Snapshot normalization strips private-mode and colon-subparameter CSI escapes
  (`\x1b[?25l` cursor hide/show, `\x1b[?1049h` alt-screen, `\x1b[38:2:…m`) and
  OSC sequences, which every spinner/TUI emits — they previously leaked raw
  escape bytes into golden files.
- A snapshot golden checked out with CRLF line endings (git `autocrlf`, a CRLF
  editor) now matches LF-folded actual output; CRLF is folded on both sides, not
  only the actual.
- Snapshot port masking consumes the whole port number, so a single-digit
  ephemeral port is masked and a >5-digit value no longer leaves an orphan
  trailing digit.
- A background service's `ready.log` regexp probe is now `${name}`-expanded like
  its `ready.file`/`ready.port` siblings; a probe referencing `${workdir}` was
  compiled verbatim and could never match, so the service always hit its
  readiness timeout and the scenario errored falsely.
- A `store.name` or `matrix` key that reuses a built-in variable name
  (`atago`/`workdir`/`suitedir`) is rejected at load time instead of silently
  shadowing the built-in and breaking scenario isolation.

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
