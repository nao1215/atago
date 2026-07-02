# atago Specification

This is the reference specification for the **atago** spec-file format and CLI
behavior. It is the document the `spec.md §N` anchors throughout the Go source
refer to. It was reconstructed from — and is kept consistent with — the actual
implementation under `internal/` and the published JSON Schema at
`schema/atago.schema.json`.

atago is a black-box behavior spec runner for CLIs, APIs, and generated
artifacts. It runs a real command, in any language, and asserts what a user
would observe: the exit code, stdout, stderr, generated files, JSON/YAML output,
database rows, HTTP/gRPC responses, browser state, and normalized snapshots.
Specs are plain YAML. There is no embedded scripting or expression language;
variables expand only through `${name}` substitution.

The machine-readable spec file schema lives at
[`schema/atago.schema.json`](schema/atago.schema.json); the commented, runnable
examples under [`examples/`](examples/) cover every feature.

## Contents

- §1–4 Overview and model
- §5–6 Self-hosted testing (atago tested by atago)
- §14 Runner families and step kinds
- §15 Command execution; §15.1 stdin, retry, and readiness polling
- §16 Assertions (§16.1 exit_code, §16.2 stdout, §16.3 stderr/JSON, §16.6 file,
  §16.7 HTTP/gRPC targets, §16.10 snapshots)
- §17 Fixtures, store, and the workdir
- §18 Variables and value binding
- §19 Selection: tags, skip/only conditions
- §20 Subcommands and the CLI
- §21 `run` options
- §22 Matrix and retry
- §24 `explain`
- §25 / §35 Doc generation
- §26 / §27 Reports; §27.1 failure output, §27.2 machine-readable report
- §28 Security model (§28.2 network policy, §28.3 secret masking, §28.4 path confinement)
- §30 Isolation and concurrency
- §31 Execution pipeline (§31.1 validation layers, §31.3 run results)
- §34 Exit codes

---

## §1 Purpose

atago tests the *outside* of an executable by running it and checking the
result. It is not a shell testing framework and not a workflow engine. YAML stays
structured data.

## §2 Spec files

A spec file matches `*.atago.yaml` / `*.atago.yml`. It declares a `suite` of
`scenarios`, each a sequence of `steps`. Files are decoded strictly (unknown keys
are errors) and validated before any step runs.

## §3 Top-level shape

```yaml
version: "1"
suite:
  name: <string>
runners: { <name>: { type: ... } }   # optional
permissions: { network: { allow: [...] } }  # optional
secrets: [ ... ]                      # optional
scenarios: [ ... ]
```

`version` must be `"1"`; `suite.name` is required; `scenarios` must contain at
least one scenario. Full field reference:
[`schema/atago.schema.json`](schema/atago.schema.json).

## §4 Design constraints

Structured YAML only. No embedded scripting; variables expand only via `${name}`
substitution. Assertions are black-box: they observe output, not internals.

## §5 Self-hosted testing

atago must be tested by atago. The hermetic self-hosted suite under
`test/e2e/atago` is written in atago's own YAML and run by the freshly built
binary. `.github/workflows/e2e.yml` builds `atago` and runs this suite on every
push/PR to `main`; a release must not be tagged unless it passes.

## §6 The `${atago}` builtin

`${atago}` resolves to the absolute path of the running atago binary, so a
self-hosted spec can invoke atago from inside its isolated temporary workdir.
This is how atago exercises its own CLI end to end.

---

## §14 Runner families and step kinds

A step sets exactly one action key: `fixture`, `run`, `http`, `query`, `grpc`,
`cdp`, `assert`, or `store`. The runner families are:

- **cmd** — a local process (the implicit default for `run`).
- **http** — HTTP requests via a named `http` runner (`base_url`). The payload
  is one of `json:` (a structured value, Content-Type application/json),
  `body:` (a raw string sent verbatim, Content-Type text/plain unless a
  `header` overrides it — for text-first APIs such as metrics exposition or
  message publishing), `body_file:` (a workdir-relative file streamed
  binary-safe as the request body — for direct upload endpoints), or `form:`
  (+ `files:`) — form fields sent urlencoded on their own, or as
  multipart/form-data once a file part is attached (the browser-style upload
  most web apps expect). Setting more than one payload family is a load-time
  error. `body_to:` writes the response body to a workdir-relative file so a
  download can be checked with the file/image/pdf assertion targets.
  `follow_redirects: false` surfaces a 3xx response instead of following it,
  so the redirect's status and Location header are assertable.
- **db** — SQL via a named `db` runner (`dsn`; SQLite/PostgreSQL/MySQL, pure Go).
- **ssh** — a `run` step naming an `ssh` runner executes remotely.
- **grpc** — a unary gRPC call via a named `grpc` runner (server reflection).
- **browser** — a `cdp` step drives a headless Chrome via a `browser` runner.
- **service** — a background peer started for the duration of a scenario. It
  starts before the steps run — except the leading run of `fixture` steps,
  which is applied first so a service can read authored input (its config
  file, seed data). Fixtures after the first non-fixture step keep their
  in-order timing.

Runners are declared under the top-level `runners` map and referenced by name.
`type`/`cwd`/`timeout` are common to every runner; every other field belongs to
exactly one type, and a cross-type field is a load-time error.

## §15 Command execution

A `run` step runs `command` and captures its exit code, stdout, and stderr. By
default the command is tokenized and executed directly. `shell: true` opts into
POSIX-shell execution (pipes, redirects, `${}`). Other fields: `runner`, `cwd`,
`timeout`, `env`, `stdin`.

### §15.1 stdin, retry, and readiness polling

`stdin` feeds a string to the command's standard input within its timeout
budget; if the command never reads it, the run step fails. `run.retry:
{ times, interval, until }` re-runs the command until the `until` assertion
passes, polling declaratively for async behavior; the last attempt's result is
what later steps observe. `http` steps take the same `retry` block, re-issuing
the request until `until` passes — for eventually-consistent endpoints such as
a metric that appears after a scrape or an async job flipping to done. A
service `ready` probe polls until the peer is up (see §24).

---

## §16 Assertions

An `assert` step sets exactly one target family. Each `Check` returns structured
context (expected/actual/hint) for failure output (§27.1) and the JSON report
(§27.2).

### §16.1 exit_code

`exit_code` accepts either a bare integer (equals) or `{ not: <int> }`.

### §16.2 stdout (stream matchers)

A stream target (`stdout`, `stderr`, `body`, `rows`, `message`, `value`) sets
exactly one matcher: `empty`, `contains`, `not_contains`, `matches`, `equals`,
`not_equals`, `json`, `yaml`, or `snapshot`. An optional 1-based `line` narrows
the stream to a single line (only with the text matchers, not json/yaml/snapshot).

### §16.3 stderr and JSON/YAML matching

`stderr` uses the same stream matchers as §16.2. The `json`/`yaml` matchers
select a value by JSONPath `path` and apply one of `equals`, `matches`,
`length`, or a numeric bound `gt`/`gte`/`lt`/`lte`. `equals` compares structures
and is insensitive to map key ordering.

### §16.6 file

`file` sets `path` plus exactly one of `exists`, `contains`, `not_contains`,
`executable`, `json`, or `snapshot`. Relative paths resolve inside the scenario
workdir. `image` (a related file-inspecting target) decodes a generated image
and asserts `format`, dimensions/bounds, `alpha`, and `similar_to`; every set
field is a constraint and all must hold.

### §16.7 HTTP, gRPC, and stream targets

`status`/`header` assert an HTTP response; `grpc_status`/`message` assert a gRPC
response; `rows` asserts DB result rows; `value` asserts a captured browser
value. `message`/`rows`/`value` reuse the stream matchers. Header names are
matched case-insensitively per RFC 7230.

### §16.10 snapshots

A `snapshot` matcher compares captured output against a committed golden file.
Captured output is normalized — ANSI codes, temp paths, the home directory,
UUIDs, timestamps, and local ports are masked, and declared secrets are masked
first — so snapshots stay stable across machines and runs. `atago snapshot
update` (or `run --update-snapshots`) writes the snapshot instead of comparing.

---

## §17 Fixtures, store, and the workdir

A `fixture` step materializes an input file in the scenario workdir: exactly one
of inline `content`, `base64`, `from` (copy an existing file, resolved relative
to the spec directory), or `symlink`. Optional `mode` (octal) and `mtime`
(RFC3339). The scenario workdir is the isolated temporary directory each scenario
runs in; the store (§18) also seeds tool environment (e.g. `GOBIN` under the
workdir) that child toolchains require.

## §18 Variables and value binding

`${name}` references expand from the per-scenario store, which holds matrix-row
variables, `store`-captured values, service ready-file content, and builtins such
as `${workdir}` and `${atago}`. A `store` step captures a value (`stdout`,
`body`, `file`, `header`, `rows`, `message`, or `value` — exactly one source)
into a named variable for later steps: a login response's token flows into a
later request. Expansion is deliberately simple substitution. Write `$${name}`
for a literal `${name}`; a bare `$$` (a shell PID) is left untouched. Unknown
references are left verbatim so they surface as failures rather than empty
strings — a shell (`shell: true`, or a remote shell via an ssh runner) may
still expand them as its own variables. In a local `run` without `shell: true`
nothing could ever expand the reference, so the step errors naming it instead
of leaking the literal text into argv.

## §19 Selection: tags and conditions

`tags` label scenarios for `--tag`/`--skip-tag` filtering. `skip`/`only` gate a
scenario by a `Condition`: `os` (compared to the host: linux/darwin/windows),
`env` (true when the named variable is non-empty), or `command` (true when the
probe command exits 0; runs through the shell). `skip` gates off, `only` gates
on. Cheap OS/env checks run before a probe command so the probe only runs when
still relevant.

---

## §20 Subcommands and the CLI

```
atago run       Run spec files and assert behavior
atago init      Scaffold a starter spec file
atago explain   Describe what a spec does without running it
atago doc       Generate Markdown documentation from specs
atago manifest  Emit a stable machine-readable JSON summary of specs
atago snapshot  Manage snapshots (snapshot update <paths>)
atago version   Print the atago version
atago help      Show help
```

A run/doc/manifest target may be a spec file or a directory; a directory is
searched recursively for `*.atago.yaml`/`*.atago.yml`.

## §21 `run` options

- `--report console|json|junit|gha|tap` — output format (default `console`).
- `--update-snapshots` — write snapshots instead of comparing.
- `--parallel N` — run up to N scenarios concurrently (default: the number of
  CPUs; scenarios are isolated in per-scenario temp dirs, so concurrency is the
  default. `--parallel 1` pins serial scheduling, e.g. for deterministic
  `--fail-fast` skip counts).
- `--fail-fast` — stop scheduling new scenarios after the first failure.
- `--filter S` — run only scenarios whose name contains S.
- `--tag T` / `--skip-tag T` — comma-separated tag include/exclude.
- `--artifacts-dir DIR` — write deterministic failure artifacts under DIR.
- `--ci` — CI-safe defaults (deterministic, `NO_COLOR`, secret masking).

## §22 Matrix and retry

A `matrix:` scenario is a template: the loader expands it into one concrete
scenario per row before validation, in definition order, seeding each row's
key/value pairs as `${name}` variables for that instance. The instance name
substitutes `${var}` tokens in the template, or appends a deterministic `[k=v …]`
suffix. `run.retry` (§15.1) re-runs a command until an assertion passes.

---

## §24 `explain`

`atago explain` renders a human- and LLM-readable summary of what a spec does
without executing it: scenarios, commands, expected behavior, fixtures, generated
files, variables, and security-sensitive operations (declared secrets, network
policy). A spec with no allowlist is described as *unrestricted*.

## §25 Doc generation

`atago doc` renders Markdown documentation from specs: each spec becomes a
section with `Given`/`When`/`Then` subsections per scenario, covering run,
http, query, grpc, cdp steps, and background services. `--out` writes to a file
(default stdout). See also §35.

## §26 Reports

`atago run --report` supports five formats: `console` (default, human-readable
with live progress), `json` (machine-readable, §27.2), `junit` (JUnit XML for
CI), `gha` (GitHub Actions workflow-command annotations), and `tap` (Test
Anything Protocol, TAP 13).

## §27 Failure output and machine-readable report

### §27.1 Failure output

Each assertion produces structured expected/actual/hint context; the console
report prints compact failure context so a reader sees exactly what did not
match. Setup-phase failures (service readiness, workdir creation) are labeled
(e.g. `service setup`) rather than emitting a blank step.

### §27.2 Machine-readable report

`--report json` emits one stable top-level shape,
`{"schema_version":"1","suites":[...]}`, regardless of how many suites ran, with
enough failure context for an LLM agent to act on. Consumers branch on
`schema_version` for future format changes.

---

## §28 Security model

atago's security model is masking, network confinement, and path confinement.

### §28.2 Network policy

`permissions.network.allow` lists hosts scenarios may contact. An empty or absent
allowlist is unrestricted. A denied host is a policy violation (exit 6), enforced
for HTTP, gRPC, and SSH.

### §28.3 Secret masking

Declared `secrets` are masked in every report and snapshot — including values
injected via scenario `env` or a service's `env` — so a real credential never
reaches logs, reports, or a committed snapshot.

### §28.4 Path confinement

A user-declared path stays inside the root it is scoped to. A workdir-scoped path
(assertion file, store source, service ready file, CDP screenshot) may not escape
the scenario workdir; a snapshot path may not escape the spec directory. `../`
traversal is rejected; an absolute path is allowed only when it still resolves
inside the root. One shared resolver enforces this everywhere.

---

## §30 Isolation and concurrency

Each scenario runs in its own temporary workdir. Scenarios may run concurrently
(`--parallel N`), optionally under a shared global semaphore capping the total
in-flight count across suites, but results and the failure report always stay in
definition order for determinism. `--fail-fast` stops scheduling after the first
failure; in-flight scenarios finish. A run cancels on `Ctrl-C`/`SIGTERM`, killing
child process groups and marking unstarted scenarios *skipped after interrupt*.

## §31 Execution pipeline

### §31.1 Validation layers

Loading and validation happen in layers, decoupled from the raw YAML: (1) YAML
parse, (2) matrix-shape validation and expansion, (3) schema/semantic validation
that reports every problem in one pass. The typed model lives in `internal/spec`;
loading/validation live in `internal/loader`.

### §31.3 Run results

A `Result` is the externally observable outcome of a run step (exit code,
stdout, stderr, captured artifacts). A runner that fails to *start* a command is
an execution error (exit 4); a command that runs but exits non-zero is a normal
result whose exit code the assertions check.

---

## §34 Exit codes

Exit codes are a stable, user-facing contract (`internal/cli/cli.go`):

| Code | Name | Meaning |
| --- | --- | --- |
| 0 | OK | all scenarios passed |
| 1 | Failures | one or more scenarios failed |
| 2 | Parse | spec parse or schema/semantic validation error |
| 3 | Config | CLI-invocation error (unknown command, bad flag, no files) |
| 4 | Exec | execution error (a runner could not run a step) |
| 5 | Internal | internal error |
| 6 | Security | security policy violation (e.g. a denied host) |

A security violation takes precedence over the generic execution-error code. A
spec-content problem (e.g. a `db` runner missing its `dsn`) exits 2, while exit 3
is reserved for CLI-invocation problems.

---

## §35 Generated behavior docs

The committed behavior docs under `doc/e2e/*.md` are generated by `atago doc`
from the E2E/dogfood spec directories and kept in sync by `docs_test.go`, so the
documentation cannot silently rot. Regenerate with `make docs`. See §25 for the
generator.

---

## Architecture Decision Records

`ADR-NNNN` citations in the Go source refer to design decisions made during
development (ADR-0018 through ADR-0038). The decision context lives in the
code comments at each citation site.
