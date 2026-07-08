# Cookbook

Task-oriented YAML recipes: find what you want to do, copy the spec, replace
`mytool` with your binary. Each recipe links to a runnable, CI-tested spec
under [examples/](../examples/) that covers the same feature in full.

## Start from a recorded run

The fastest way to a first spec is not writing one. Run the tool once under
`record` and atago writes the spec from what it observed:

```shell
atago record --out convert.atago.yaml -- mytool convert input.txt
atago run convert.atago.yaml
```

The generated spec is plain YAML — the same shape as every recipe below — so
the workflow is: record the happy path, then edit the file to tighten matchers,
add failure cases, and delete asserts you don't care about:

```yaml
version: "1"
suite:
  name: recorded
scenarios:
  - name: mytool convert input.txt
    steps:
      - run:
          command: mytool convert input.txt
      - assert:
          exit_code: 0
          stdout:
            contains: "wrote output.txt"     # recorded verbatim; loosen or tighten
      - assert:
          file:
            path: output.txt
            exists: true
```

Interactive tools record too: `atago record --pty -- mytool init` captures one
hand-driven session and writes a `pty:` step that replays it.

## Test a CLI that converts images

```yaml
version: "1"
suite:
  name: image conversion

scenarios:
  - name: png to jpeg keeps the dimensions
    steps:
      # Write the input image into the isolated workdir. base64 carries raw
      # bytes that inline YAML cannot.
      - fixture:
          file: in.png
          base64: iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR4nGNocFBoAAADhQFhC+q+qAAAAABJRU5ErkJggg==
      - run:
          command: mytool convert --format jpeg in.png out.jpg
      - assert:
          exit_code: 0
          # Decode the output and check observable properties. format is
          # sniffed from the file content, not the extension.
          image:
            path: out.jpg
            format: jpeg
            width: 1
            height: 1
      # Optionally pin pixel content against a committed baseline:
      #   image: {path: out.jpg, similar_to: golden/out.jpg, max_diff: 0.01}
```

Full spec: [image_and_pdf](../examples/image_and_pdf.atago.yaml)

## Test a CLI that generates a PDF

```yaml
version: "1"
suite:
  name: pdf generation

scenarios:
  - name: the report has the right pages, metadata, and text
    steps:
      - run:
          command: mytool report --out report.pdf
      - assert:
          exit_code: 0
          pdf:
            path: report.pdf
            pages: 2
            metadata:
              title: Quarterly       # substring match on the Info dictionary
            text:
              contains: Quarterly earnings   # extracted text
```

Full spec: [image_and_pdf](../examples/image_and_pdf.atago.yaml)

## Test a CLI that generates files

```yaml
version: "1"
suite:
  name: file generation

scenarios:
  - name: init scaffolds a project
    steps:
      - run:
          command: mytool init --name demo
      - assert:
          exit_code: 0
          # dir checks the tree shape; each listed path must exist.
          dir:
            path: demo
            contains:
              - demo.yaml
              - src/main.go
            not_contains:
              - secret.key
      - assert:
          # file checks one file's content.
          file:
            path: demo/demo.yaml
            contains: "name: demo"

  - name: a round-trip is byte-identical
    steps:
      - fixture:
          file: in.dat
          content: "DEADBEEF"
      - run:
          shell: true
          command: mytool encode in.dat enc.dat && mytool decode enc.dat out.dat
      - assert:
          # equals_file compares two runtime files byte for byte — no
          # CRLF/newline normalization.
          file:
            path: out.dat
            equals_file: in.dat
```

Full spec: [files_and_fixtures](../examples/files_and_fixtures.atago.yaml)

## Pin exactly which files a command touches

```yaml
version: "1"
suite:
  name: workdir delta

scenarios:
  - name: build touches only its outputs
    steps:
      - fixture:
          file: config.yaml
          content: "theme: dark\n"
      - run:
          command: mytool build
      - assert:
          exit_code: 0
          # Exhaustive in both directions: every observed change must match an
          # entry, and every entry must be observed. modified: [] asserts the
          # command modified nothing.
          changes:
            created:
              - site/index.html
              - site/assets/*.css   # /-globs work
            modified: []
            deleted: []
```

Full spec: [changes](../examples/changes.atago.yaml)

## Test error handling: exit codes and stderr

```yaml
version: "1"
suite:
  name: failure modes

scenarios:
  - name: a missing input fails loudly with a useful message
    steps:
      - run:
          command: mytool convert no-such-file.png out.jpg
      - assert:
          exit_code: 2          # or `not: 0`, or a documented set: `in: [1, 2]`
          stderr:
            contains: no-such-file.png   # the message names the missing file
          stdout:
            empty: true
      # The failed run created nothing.
      - assert:
          file:
            path: out.jpg
            exists: false
```

Full spec: [run_and_assert](../examples/run_and_assert.atago.yaml)

## Feed stdin to a filter CLI

```yaml
version: "1"
suite:
  name: stdin

scenarios:
  - name: filter a document from stdin to stdout
    steps:
      - fixture:
          file: input.txt
          content: |
            beta
            alpha
      - run:
          command: mytool sort
          stdin:
            file: input.txt      # or inline: `stdin: "text"`, or binary:
                                 # `stdin: {base64: AAEC/w==}`
      - assert:
          exit_code: 0
          stdout:
            contains:
              - alpha
              - beta
```

Full spec: [stdin](../examples/stdin.atago.yaml)

## Assert on JSON or YAML output

```yaml
version: "1"
suite:
  name: structured output

scenarios:
  - name: the status command reports healthy
    steps:
      - run:
          command: mytool status --json
      # Select values with JSONPath. A list of checks under one json: all must
      # hold; gt/gte/lt/lte bound numeric values that vary run to run.
      - assert:
          exit_code: 0
          stdout:
            json:
              - { path: "$.status", equals: ok }
              - { path: "$.jobs", length: 3 }
              - { path: "$.uptime_seconds", gt: 0 }

  - name: yaml output uses the same path syntax
    steps:
      - run:
          command: mytool config dump
      - assert:
          stdout:
            yaml:
              path: "$.theme"
              equals: dark
```

Full spec: [json_and_yaml](../examples/json_and_yaml.atago.yaml)

## Check what a CLI wrote to a database

```yaml
version: "1"
suite:
  name: database side effects

# A db runner with a workdir-scoped SQLite file: every scenario gets its own
# isolated database. Pure-Go drivers for SQLite, PostgreSQL, and MySQL are
# bundled.
runners:
  store:
    type: db
    dsn: sqlite:${workdir}/app.db   # or postgres://... / mysql://...

scenarios:
  - name: import lands the rows
    steps:
      - fixture:
          file: users.csv
          content: |
            id,name
            1,alice
      - run:
          command: mytool import --db app.db users.csv
      - assert:
          exit_code: 0
      # query runs SQL; rows asserts on the result set as a JSON array.
      - query:
          runner: store
          sql: "SELECT name FROM users ORDER BY id"
      - assert:
          rows:
            json:
              - { path: "$", length: 1 }
              - { path: "$[0].name", equals: alice }
```

Full spec: [db](../examples/db.atago.yaml)

## Test an interactive prompt

```yaml
version: "1"
suite:
  name: interactive prompt

scenarios:
  - name: the setup wizard accepts answers
    steps:
      # pty runs the command in a real pseudo-terminal; expect waits for the
      # regexp, send types (include \n to press enter, or use named keys).
      - pty:
          command: mytool setup
          timeout: 30s
          session:
            - expect: "Project name:"
            - send: "demo\n"
            - expect: "Use defaults\\? \\[Y/n\\]"
            - send: { key: enter }
            - expect: "Done"
      - assert:
          exit_code: 0
          stdout:
            contains: Done      # the transcript is the step's stdout
```

Full spec: [pty](../examples/pty.atago.yaml) (POSIX-only inner commands), [pty_portable](../examples/pty_portable.atago.yaml) (runs on Windows too)

## Test a full-screen TUI

```yaml
version: "1"
suite:
  name: tui screen

scenarios:
  - name: the dashboard renders its header
    steps:
      - pty:
          command: mytool dashboard
          rows: 24
          cols: 80
          timeout: 10s
          session:
            - expect: "Status"
            - send: "q"
      # screen replays the transcript through a vt100 emulator and asserts on
      # the final rendered frame — what the user actually sees.
      - assert:
          screen:
            line: 1              # 1-based screen row
            contains: Status
      # Or pin the whole frame: screen: {snapshot: snapshots/dashboard.txt}
```

Full spec: [pty_screen](../examples/pty_screen.atago.yaml)

## Test an API-client CLI without the network

```yaml
version: "1"
suite:
  name: offline api client

scenarios:
  - name: push sends the right request
    # A stub HTTP server on an ephemeral loopback port; ${api.url} points at it.
    mock_servers:
      - name: api
        routes:
          - method: POST
            path: /v1/reports
            status: 201
            json: { id: "r-1" }
    steps:
      - run:
          command: mytool push --endpoint ${api.url} report.txt
      - assert:
          exit_code: 0
      # The headline assertion: what did the CLI actually send?
      - assert:
          mock:
            name: api
            path: /v1/reports
            method: POST
            count: 1
            header: { name: Authorization, matches: "^Bearer " }
            body:
              json: { path: "$.title", equals: "report" }
```

Full spec: [mock_server](../examples/mock_server.atago.yaml)

## Test a CLI that starts a server

```yaml
version: "1"
suite:
  name: server lifecycle

scenarios:
  - name: serve answers requests once ready
    # services start before the steps and are torn down when the scenario ends.
    services:
      - name: api
        command: mytool serve --addr 127.0.0.1:8080
        ready:
          port: 127.0.0.1:8080   # gate the steps until the port accepts
          timeout: 10s           # ...or ready: {log: "listening on"}
    steps:
      - run:
          shell: true
          command: curl -sf http://127.0.0.1:8080/health
      - assert:
          exit_code: 0
          stdout:
            contains: ok
```

Full spec: [services](../examples/services.atago.yaml)

## Test graceful shutdown

```yaml
version: "1"
suite:
  name: graceful shutdown

scenarios:
  - name: the server cleans up on SIGTERM
    skip:
      os: windows        # signal steps are POSIX-only
    services:
      - name: server
        command: mytool serve
        ready:
          log: "listening on"
    steps:
      # signal targets a service atago started — race-free under --parallel,
      # unlike kill/killall by process name. wait fails loudly if the process
      # never exits.
      - signal:
          service: server
          signal: TERM
          wait:
            timeout: 5s
      # The evidence of a clean shutdown: whatever your server leaves behind.
      - assert:
          file:
            path: server.log
            contains: "graceful shutdown complete"
```

Full spec: [signal](../examples/signal.atago.yaml)

## Pin a generator's whole output tree

```yaml
version: "1"
suite:
  name: output tree

scenarios:
  - name: scaffold emits exactly the committed tree
    steps:
      - run:
          command: mytool new site
      - assert:
          exit_code: 0
          # One golden manifest line per entry (dirs, files with a sha256,
          # symlinks) replaces a ladder of per-path asserts. Refresh with:
          #   atago snapshot update spec.atago.yaml
          dir:
            path: site
            snapshot: snapshots/site_tree.txt
            ignore: ["*.log"]   # exclude noise from the walk and the manifest

  - name: recursive matchers when the exact tree is an implementation detail
    steps:
      - run:
          command: mytool new site
      - assert:
          dir:
            path: site
            recursive: true
            contains: [content/posts/hello.md]
            glob: "*.md"        # a "/"-less glob matches basenames at any depth
            min_count: 2        # bound the file total without pinning it
            max_count: 10
```

Full spec: [dir_tree](../examples/dir_tree.atago.yaml)

## Pin output with a golden file

```yaml
version: "1"
suite:
  name: golden output

# Rewrite volatile patterns the built-in normalizers (paths, UUIDs,
# timestamps, ports, ANSI colors) do not cover.
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}

scenarios:
  - name: help output stays stable
    steps:
      - run:
          command: mytool --help
      - assert:
          stdout:
            snapshot: snapshots/help.txt
```

Record or refresh with `atago snapshot update spec.atago.yaml`.
Full spec: [snapshot](../examples/snapshot.atago.yaml), [scrub](../examples/scrub.atago.yaml)

## Poll an async result

```yaml
version: "1"
suite:
  name: polling

scenarios:
  - name: the job finishes eventually
    steps:
      - run:
          command: mytool job start
      # Re-run until the until-assertion passes or the budget is spent.
      - run:
          command: mytool job status
          retry:
            times: 20
            interval: 250ms
            until:
              stdout:
                contains: "state: done"
      - assert:
          stdout:
            contains: "state: done"
```

Full spec: [retry](../examples/retry.atago.yaml)

## Bound how long a command may take

```yaml
version: "1"
suite:
  name: timing

scenarios:
  - name: status answers fast, backoff actually waits
    steps:
      - run:
          command: mytool status
      # duration bounds the immediately preceding step. Assert orders of
      # magnitude, not milliseconds — tight bounds flake on loaded CI runners.
      - assert:
          exit_code: 0
          duration:
            lt: 10s
      - run:
          command: mytool retry --backoff 200ms
      - assert:
          duration:
            gte: 200ms     # the backoff really waited
            lt: 60s
```

Full spec: [duration](../examples/duration.atago.yaml)

## Clean up external state even when a step fails

```yaml
version: "1"
suite:
  name: cleanup

scenarios:
  - name: the created resource never leaks
    steps:
      - run:
          command: mytool create
      # Capture the id the create printed; teardown shares the same store.
      - store:
          name: rid
          from:
            stdout:
              matches: "resource-[0-9]+"
      - assert:
          stdout:
            contains: created
    # teardown always runs — pass, fail, error, or interrupt. A teardown
    # failure is reported but never changes the scenario's verdict.
    teardown:
      - run:
          command: mytool delete ${rid}
```

Full spec: [teardown](../examples/teardown.atago.yaml)

## Run expensive setup once for the whole suite

```yaml
version: "1"
suite:
  name: shared setup
  # setup runs once, in order, before any scenario; ${suitedir} outlives every
  # scenario workdir. A suite-level teardown: would run once after the last
  # scenario, and a suite-level service: step starts a peer shared by all.
  setup:
    - run:
        shell: true
        command: go build -o ${suitedir}/mytool ./cmd/mytool

scenarios:
  - name: every scenario runs the binary built once
    steps:
      - run:
          command: ${suitedir}/mytool --version
      - assert:
          exit_code: 0
```

Full spec: [suite_setup](../examples/suite_setup.atago.yaml)

## Run a scenario only where it can pass

```yaml
version: "1"
suite:
  name: gating

scenarios:
  - name: skipped on windows
    skip:
      os: windows
    steps:
      - run:
          command: mytool daemon --check
      - assert:
          exit_code: 0

  - name: needs docker on the host
    only:
      command: "docker info"    # runs only when the probe exits 0
    steps:
      - run:
          command: mytool up
      - assert:
          exit_code: 0

  - name: tagged for --tag smoke / --skip-tag slow selection
    tags: [smoke]
    steps:
      - run:
          command: mytool --version
      - assert:
          exit_code: 0
```

Full spec: [select_skip_only](../examples/select_skip_only.atago.yaml)

## Run the same scenario over many inputs

```yaml
version: "1"
suite:
  name: matrix

scenarios:
  - name: "converts ${input} to ${format}"
    # One template scenario expands into one concrete scenario per row.
    matrix:
      - { input: photo.png, format: jpeg }
      - { input: photo.png, format: webp }
      - { input: icon.gif,  format: png }
    steps:
      - run:
          command: mytool convert --format ${format} ${input} out.${format}
      - assert:
          exit_code: 0
          # Matrix variables expand in paths; the conversion keeps dimensions.
          image:
            path: out.${format}
            width: 640
            height: 480
```

Full spec: [matrix](../examples/matrix.atago.yaml)

## Capture a value in one step and reuse it

```yaml
version: "1"
suite:
  name: capture and reuse

scenarios:
  - name: create a resource, then inspect it by id
    steps:
      - run:
          command: mytool create --json
      # Capture from stdout by JSONPath (or regex with matches:, or a whole
      # trimmed value with trim: true); later steps use ${id}.
      - store:
          name: id
          from:
            stdout:
              json:
                path: "$.id"
      - run:
          command: mytool show ${id}
      - assert:
          exit_code: 0
          stdout:
            contains: ${id}
```

Full spec: [store_and_variables](../examples/store_and_variables.atago.yaml)

## Isolate the test from the host environment

```yaml
version: "1"
suite:
  name: hermetic

scenarios:
  - name: config comes from the sandboxed home, not the host
    steps:
      # clear_env starts the child from an empty environment; pass_env
      # re-admits an allowlist. sandbox_home points HOME and the XDG dirs at a
      # fresh ${workdir}/.atago-home, so ~/.config reads/writes are isolated.
      - run:
          command: mytool config set theme dark
          clear_env: true
          pass_env: [PATH]
          sandbox_home: true
      - run:
          command: mytool config get theme
          clear_env: true
          pass_env: [PATH]
          sandbox_home: true
      - assert:
          exit_code: 0
          stdout:
            contains: dark
      # The sandboxed home is a normal workdir path, so file asserts reach it.
      - assert:
          file:
            path: .atago-home/.config/mytool/config.yaml
            contains: "theme: dark"
```

Full spec: [hermetic_env](../examples/hermetic_env.atago.yaml)

## Pin the help and misuse contract

`--help` and error handling for wrong invocations are user-facing behavior like
any other — and the cheapest place to catch an accidentally renamed subcommand
or a usage message that stopped naming the offending flag:

```yaml
version: "1"
suite:
  name: usage contract

scenarios:
  - name: --help documents every advertised subcommand
    steps:
      - run:
          command: mytool --help
      - assert:
          exit_code: 0
          stdout:
            contains:
              - "Usage:"
              - convert
              - inspect
          stderr:
            empty: true

  - name: an unknown flag fails loudly and names itself
    steps:
      - run:
          command: mytool --no-such-flag
      - assert:
          exit_code: 2                  # or `not: 0` if the code is not pinned
          stderr:
            contains: --no-such-flag    # the user learns WHAT was wrong
          stdout:
            empty: true                 # usage errors do not pollute stdout
```

Full spec: [run_and_assert](../examples/run_and_assert.atago.yaml)

## Test a CLI that reads environment variables

Set the variable on the step to test the override; use `clear_env` to prove
the documented default applies when nothing is set — otherwise a value
exported in your own shell can silently satisfy the test:

```yaml
version: "1"
suite:
  name: env config

scenarios:
  - name: the environment override wins
    steps:
      - run:
          command: mytool config show
          env:
            MYTOOL_PORT: "9090"
      - assert:
          exit_code: 0
          stdout:
            contains: "port: 9090"

  - name: without the variable the documented default applies
    steps:
      - run:
          command: mytool config show
          clear_env: true       # empty child environment: the default must hold
          pass_env: [PATH]      # keep only what launching the binary needs
      - assert:
          exit_code: 0
          stdout:
            contains: "port: 8080"
```

Full spec: [extend_host_env](../examples/extend_host_env.atago.yaml), [hermetic_env](../examples/hermetic_env.atago.yaml)

## Mask secrets in reports and snapshots

Name the variables under `secrets:` and every occurrence of their VALUES
renders as `***` in output, reports, and snapshots — so a CI token can drive a
test without ever being committed to a golden file or leaked in a failure log:

```yaml
version: "1"
suite:
  name: secret hygiene

secrets:
  - API_TOKEN               # env var names; the values get masked everywhere

scenarios:
  - name: the token drives the run but never appears in output
    steps:
      - run:
          shell: true
          command: mytool sync --verbose    # verbose mode echoes the token
      - assert:
          exit_code: 0
          stdout:
            contains: "token=***"   # masking replaced the raw value — seeing
                                    # the placeholder proves it was substituted
```

## Test a REPL

An `expect` waits for the transcript to match before the next `send` types, so
the session stays in lockstep with the program — and each `expect` scans only
past the previous match, so a recurring prompt waits for its NEXT occurrence:

```yaml
version: "1"
suite:
  name: repl

scenarios:
  - name: the REPL evaluates input and exits cleanly on EOF
    steps:
      - pty:
          command: mytool repl
          timeout: 20s
          session:
            - expect: ">>> "            # the first prompt
            - send: "1 + 2\n"
            - expect: "3"
            - expect: ">>> "            # the NEXT prompt, after the result
            - send: { key: ctrl-d }     # EOF ends the session
      - assert:
          exit_code: 0
          stdout:
            contains: "3"
```

Full spec: [pty](../examples/pty.atago.yaml), [pty_portable](../examples/pty_portable.atago.yaml)

## Prove a command is idempotent

Run it twice and pin the second run's workdir delta to nothing. `changes:`
fields are exhaustive, so empty lists assert "created nothing, modified
nothing" — a per-file `exists:` check cannot say that:

```yaml
version: "1"
suite:
  name: idempotency

scenarios:
  - name: running init twice changes nothing the second time
    steps:
      - run:
          command: mytool init --dir out
      - run:
          command: mytool init --dir out
      - assert:
          exit_code: 0
          changes:
            created: []     # the second run created nothing...
            modified: []    # ...rewrote nothing...
            deleted: []     # ...and removed nothing
```

Full spec: [changes](../examples/changes.atago.yaml)

## Compare two implementations of the same command

Capture the reference output with `store:`, then assert the other path
produces exactly the same bytes — an oracle test for a rewrite, a `--fast`
flag, or an optimization that must not change behavior:

```yaml
version: "1"
suite:
  name: differential

scenarios:
  - name: the fast engine matches the classic engine byte for byte
    steps:
      - fixture:
          file: scene.json
          content: |
            {"width": 4, "height": 4, "shapes": ["box"]}
      - run:
          command: mytool render --engine classic scene.json
      - store:
          name: reference
          from:
            stdout:
              trim: true
      - run:
          command: mytool render --engine fast scene.json
      - assert:
          exit_code: 0
          stdout:
            equals: ${reference}
```

Full spec: [store_and_variables](../examples/store_and_variables.atago.yaml)

## Pin the version and completion contracts

`--version` output is an interface: release tooling greps it, bug reports quote
it. Pin its shape with a regex, and check the completion script actually
mentions your subcommands:

```yaml
version: "1"
suite:
  name: version contract

scenarios:
  - name: --version prints a semver on one line
    steps:
      - run:
          command: mytool --version
      - assert:
          exit_code: 0
          stdout:
            matches: '^mytool \d+\.\d+\.\d+'

  - name: shell completion covers every subcommand
    steps:
      - run:
          command: mytool completion bash
      - assert:
          exit_code: 0
          stdout:
            contains:
              - convert
              - inspect
```

Full spec: [run_and_assert](../examples/run_and_assert.atago.yaml)

## Simulate API failures offline

A `mock_servers:` route can answer anything — a 500, a 429, malformed JSON —
so the error paths of an API-client CLI become deterministic tests instead of
"hope the staging server is down today":

```yaml
version: "1"
suite:
  name: api failure modes

scenarios:
  - name: a 500 from the API surfaces as a clean CLI error
    mock_servers:
      - name: api
        routes:
          - method: GET
            path: /v1/data
            status: 500
            body: internal error
    steps:
      - run:
          command: mytool fetch --endpoint ${api.url}/v1/data
      - assert:
          exit_code: 1
          stderr:
            contains: "500"      # the CLI reports the upstream status...
          stdout:
            empty: true          # ...and writes nothing it would have to undo
      # The CLI called once — it did not hammer a failing endpoint.
      - assert:
          mock:
            name: api
            path: /v1/data
            method: GET
            count: 1
```

Full spec: [mock_server](../examples/mock_server.atago.yaml)

## Test a download command offline

Serve the payload from a mock route, let the CLI download it, and compare the
written file against the same bytes shipped as a fixture:

```yaml
version: "1"
suite:
  name: download

scenarios:
  - name: fetch writes exactly the served bytes
    mock_servers:
      - name: files
        routes:
          - method: GET
            path: /release/tool.bin
            body: "binary-payload-v1"
    steps:
      - fixture:
          file: expected.bin
          content: "binary-payload-v1"
      - run:
          command: mytool fetch ${files.url}/release/tool.bin --out tool.bin
      - assert:
          exit_code: 0
          file:
            path: tool.bin
            equals_file: expected.bin   # byte-for-byte, no normalization
```

Full spec: [mock_server](../examples/mock_server.atago.yaml)

## Verify server state after the CLI acts

When the CLI's job is to change a server (deploy, publish, upload), assert the
server afterwards — an `http:` step queries it directly, as an observer:

```yaml
version: "1"
suite:
  name: server side effects

runners:
  api:
    type: http
    base_url: http://127.0.0.1:8080

scenarios:
  - name: publish makes the article visible over HTTP
    services:
      - name: server
        command: mytool serve --port 8080
        ready:
          file: server.pid
          timeout: 10s
    steps:
      - fixture:
          file: draft.md
          content: "# Hello"
      - run:
          command: mytool publish draft.md
      - assert:
          exit_code: 0
      - http:
          runner: api
          method: GET
          path: /articles/hello
      - assert:
          status: 200
          header:
            name: Content-Type
            contains: application/json
          body:
            json:
              path: "$.title"
              equals: Hello
```

Full spec: [services](../examples/services.atago.yaml), [http](../examples/http.atago.yaml)

## Send output streams to files

For output too large or too binary to inline, redirect each stream to a file
and use file assertions — the same matchers, applied at rest:

```yaml
version: "1"
suite:
  name: stream redirects

scenarios:
  - name: the export lands in the file, diagnostics stay on stderr
    steps:
      - run:
          command: mytool export --all
          stdout_to: export.csv
          stderr_to: warnings.log
      - assert:
          exit_code: 0
          file:
            path: export.csv
            contains: "id,name,role"
      - assert:
          file:
            path: warnings.log
            contains: "skipped 2 archived records"
```

Full spec: [shell_and_redirect](../examples/shell_and_redirect.atago.yaml)

## Ship binary test data with fixtures

`content:` carries text; `base64:` carries exact bytes; `from:` copies a real
file committed next to the spec; `mode:` pins permissions. Pick by what the
input is, not by what is easiest to paste:

```yaml
version: "1"
suite:
  name: fixture sources

scenarios:
  - name: a real sample file drives the parser
    steps:
      - fixture:
          file: sample.dat
          from: testdata/sample.dat     # committed beside the spec
      - fixture:
          file: header.bin
          base64: iVBORw0KGgo=          # exact bytes, safe for any content
      - fixture:
          file: run.sh
          content: "#!/bin/sh\necho ok\n"
          mode: "0755"                  # executable fixture
      - run:
          command: mytool inspect sample.dat
      - assert:
          exit_code: 0
```

Full spec: [files_and_fixtures](../examples/files_and_fixtures.atago.yaml)

## Test how the CLI treats symlinks

A `symlink:` fixture creates the link inside the isolated workdir, so
follow/no-follow behavior is testable without touching the host filesystem:

```yaml
version: "1"
suite:
  name: symlinks

scenarios:
  - name: the scanner follows links to files but reports the link path
    skip:
      os: windows      # symlink creation needs privileges on Windows
    steps:
      - fixture:
          file: real.txt
          content: "payload"
      - fixture:
          file: link.txt
          symlink: real.txt
      - run:
          command: mytool scan link.txt
      - assert:
          exit_code: 0
          stdout:
            contains: link.txt      # reported under the name it was given
```

Full spec: [files_and_fixtures](../examples/files_and_fixtures.atago.yaml)

## Test freshness logic with fixture timestamps

`mtime:` backdates a fixture, so "rebuild only what changed" logic has a
deterministic stale file to react to — no `sleep`, no clock games:

```yaml
version: "1"
suite:
  name: freshness

scenarios:
  - name: a stale output is rebuilt, a fresh one is left alone
    steps:
      - fixture:
          file: src/page.md
          content: "# v2"
      - fixture:
          file: out/page.html
          content: "<h1>v1</h1>"
          mtime: "2020-01-01T00:00:00Z"   # older than the source
      - run:
          command: mytool build
      - assert:
          exit_code: 0
          changes:
            modified:
              - out/page.html     # rebuilt — and nothing else was touched
            created: []
            deleted: []
```

Full spec: [files_and_fixtures](../examples/files_and_fixtures.atago.yaml), [changes](../examples/changes.atago.yaml)

## Prove the tool is binary-safe

Feed exact bytes on stdin, capture stdout to a file, and compare against the
expected bytes — any CRLF mangling, encoding pass, or stray log line breaks
the equality:

```yaml
version: "1"
suite:
  name: binary safety

scenarios:
  - name: pass-through preserves every byte
    steps:
      - fixture:
          file: expected.bin
          base64: AAEC/wDerg==
      - run:
          command: mytool passthrough
          stdin:
            base64: AAEC/wDerg==
          stdout_to: got.bin
      - assert:
          exit_code: 0
          file:
            path: got.bin
            equals_file: expected.bin
```

Full spec: [stdin](../examples/stdin.atago.yaml), [files_and_fixtures](../examples/files_and_fixtures.atago.yaml)

## Test the empty-input boundary

Empty file, empty stdin: the classic crash sites. Decide what the contract IS
(error? empty output? default?), then pin it:

```yaml
version: "1"
suite:
  name: empty inputs

scenarios:
  - name: an empty file is a clean no-op, not a crash
    steps:
      - fixture:
          file: empty.csv
          content: ""
      - run:
          command: mytool convert empty.csv out.json
      - assert:
          exit_code: 0
          stdout:
            contains: "0 records"
          file:
            path: out.json
            contains: "[]"

  - name: empty stdin fails with a pointer to --help
    steps:
      - run:
          command: mytool convert
          stdin: ""
      - assert:
          exit_code: 2
          stderr:
            contains: "no input"
```

Full spec: [stdin](../examples/stdin.atago.yaml)

## Prove multibyte text survives

Round-trip text that breaks naive byte handling — CJK, emoji, combining marks —
and require exact equality, not just "contains":

```yaml
version: "1"
suite:
  name: multibyte

scenarios:
  - name: unicode passes through unchanged
    steps:
      - run:
          command: mytool echo
          stdin: "café 😀 日本語\n"
      - assert:
          exit_code: 0
          stdout:
            equals: "café 😀 日本語\n"
```

Full spec: [stdin](../examples/stdin.atago.yaml)

## Test TTY detection

Many CLIs render progress bars on a terminal but plain lines when piped. Both
branches are contracts; `pty:` runs the real-terminal one, `run:` the piped one:

```yaml
version: "1"
suite:
  name: tty detection

scenarios:
  - name: a terminal gets the interactive rendering
    steps:
      - pty:
          command: mytool export --all
      - assert:
          exit_code: 0
          stdout:
            contains: "%"          # the progress indicator drew

  - name: a pipe gets plain machine-readable lines
    steps:
      - run:
          command: mytool export --all
      - assert:
          exit_code: 0
          stdout:
            not_matches: "[%\\r]"  # no progress art in piped output
```

Full spec: [pty](../examples/pty.atago.yaml)

## Inspect an archive the CLI produced

No archive assertion exists (yet) — list the archive with a real tool and
assert on the listing, which pins the member paths:

```yaml
version: "1"
suite:
  name: archives

scenarios:
  - name: the bundle contains exactly the advertised layout
    skip:
      os: windows     # relies on tar being on PATH
    steps:
      - run:
          command: mytool bundle --out dist.tar.gz
      - assert:
          exit_code: 0
      - run:
          shell: true
          command: tar -tzf dist.tar.gz
      - assert:
          exit_code: 0
          stdout:
            contains:
              - dist/bin/mytool
              - dist/LICENSE
            not_contains:
              - .git
```

Full spec: [shell_and_redirect](../examples/shell_and_redirect.atago.yaml)

## Test the unreadable-input failure mode

A `mode: "0000"` fixture is a file the CLI cannot open — the permission-denied
path becomes reproducible instead of needing a root-owned file on the host:

```yaml
version: "1"
suite:
  name: permission errors

scenarios:
  - name: an unreadable input names the file and fails cleanly
    skip:
      os: windows       # POSIX permission semantics
    steps:
      - fixture:
          file: locked.txt
          content: "cannot read me"
          mode: "0000"
      - run:
          command: mytool convert locked.txt out.txt
      - assert:
          exit_code: 2
          stderr:
            contains: locked.txt
          file:
            path: out.txt
            exists: false
```

Full spec: [files_and_fixtures](../examples/files_and_fixtures.atago.yaml)

## Run the command from a subdirectory

Config discovery, path resolution, and "am I inside a project?" checks all
depend on the working directory. `cwd:` runs the step from a subdirectory of
the isolated workdir:

```yaml
version: "1"
suite:
  name: working directory

scenarios:
  - name: the tool finds the project root from a nested directory
    steps:
      - fixture:
          file: mytool.yaml
          content: "name: demo"
      - fixture:
          file: src/deep/keep.txt
          content: ""
      - run:
          command: mytool status
          cwd: src/deep
      - assert:
          exit_code: 0
          stdout:
            contains: "project: demo"   # found the root two levels up
```

Full spec: [run_and_assert](../examples/run_and_assert.atago.yaml)

## Write one spec that runs on all three OSes

Prefer direct argv commands (no shell) so quoting is identical everywhere; use
`shell: true` only for shell builtins, and gate genuinely POSIX-only scenarios
instead of letting them fail on Windows:

```yaml
version: "1"
suite:
  name: portable

scenarios:
  - name: the core contract holds everywhere
    steps:
      # Direct execution: no /bin/sh vs cmd.exe quoting differences exist.
      - run:
          command: mytool convert input.txt
      - assert:
          exit_code: 0
          # contains/matches don't care about \n vs \r\n line endings within
          # a line; avoid asserting exact multi-line blocks with equals here.
          stdout:
            contains: "wrote output.txt"

  - name: the POSIX-only part is gated, not broken
    skip:
      os: windows
    steps:
      - run:
          shell: true
          command: 'test -x output.txt || echo not executable'
      - assert:
          stdout:
            contains: not executable
```

Full spec: [pty_portable](../examples/pty_portable.atago.yaml), [select_skip_only](../examples/select_skip_only.atago.yaml)

## Compare releases of your CLI with a matrix

A matrix variable can be the command itself — run the same contract against
two installed versions to catch behavior drift before shipping:

```yaml
version: "1"
suite:
  name: release comparison

scenarios:
  - name: "the convert contract holds in ${bin}"
    matrix:
      - bin: mytool-v1
      - bin: mytool-v2
    steps:
      - fixture:
          file: input.txt
          content: "hello"
      - run:
          command: ${bin} convert input.txt
      - assert:
          exit_code: 0
          file:
            path: output.txt
            contains: HELLO
```

Full spec: [matrix](../examples/matrix.atago.yaml)

## Tag scenarios and run a slice in CI

Tags split one suite into fast/slow, smoke/full, or per-feature slices —
selection happens at run time, so the spec files stay together:

```yaml
version: "1"
suite:
  name: tagged suite

scenarios:
  - name: version prints instantly
    tags: [smoke, fast]
    steps:
      - run:
          command: mytool --version
      - assert:
          exit_code: 0

  - name: a full export takes a while
    tags: [slow]
    steps:
      - run:
          command: mytool export --all
      - assert:
          exit_code: 0
```

```shell
atago run --tag smoke ./specs        # PR gate: the fast slice
atago run --skip-tag slow ./specs    # everything except the expensive ones
atago run --ci ./specs               # nightly: the whole suite
```

Full spec: [select_skip_only](../examples/select_skip_only.atago.yaml)

## Set defaults once for the whole suite

`defaults:` applies a setting to every step or scenario, so specs stop
repeating `shell: true` or a shared environment variable line by line:

```yaml
version: "1"
suite:
  name: suite defaults

defaults:
  run:
    shell: true            # every run step goes through the shell
  scenario:
    env:
      MYTOOL_NO_COLOR: "1" # every scenario gets deterministic plain output

scenarios:
  - name: steps inherit both defaults
    steps:
      - run:
          command: echo plain and shelled
      - assert:
          exit_code: 0
          stdout:
            contains: plain and shelled
```

Full spec: [defaults](../examples/defaults.atago.yaml)

## Assert a database migration's schema

Run the migration, then query the schema catalog with a `db` runner — the
bundled SQLite needs no server, and the `rows:` assertion reads the result:

```yaml
version: "1"
suite:
  name: migrations

runners:
  store:
    type: db
    dsn: sqlite:${workdir}/app.db

scenarios:
  - name: migrate creates the promised tables
    steps:
      - run:
          command: mytool migrate --db app.db
      - assert:
          exit_code: 0
      - query:
          runner: store
          sql: "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name"
      - assert:
          rows:
            json:
              - { path: "$[0].name", equals: jobs }
              - { path: "$[1].name", equals: users }
```

Full spec: [db](../examples/db.atago.yaml)

## Run the CLI on a remote host over SSH

An `ssh` runner executes run steps on another machine — for a CLI whose job is
provisioning or deployment, the remote is where the observable behavior lives:

```yaml
version: "1"
suite:
  name: remote

runners:
  box:
    type: ssh
    host: staging.example.com
    user: deploy
    key_file: ~/.ssh/id_ed25519
    known_hosts: ~/.ssh/known_hosts

scenarios:
  - name: deploy leaves the service running on the box
    steps:
      - run:
          command: mytool deploy --target staging
      - assert:
          exit_code: 0
      - run:
          runner: box
          command: systemctl is-active myservice
      - assert:
          exit_code: 0
          stdout:
            contains: active
```

Full spec: [ssh](../examples/ssh.atago.yaml)

## Check a gRPC dependency after the CLI acts

A `grpc:` step calls a method through server reflection — use it to observe
the state your CLI was supposed to change:

```yaml
version: "1"
suite:
  name: grpc side effects

runners:
  registry:
    type: grpc
    target: localhost:50051

scenarios:
  - name: register makes the entry queryable over gRPC
    steps:
      - run:
          command: mytool register --name demo
      - assert:
          exit_code: 0
      - grpc:
          runner: registry
          method: registry.v1.Registry/Get
          json:
            name: demo
      - assert:
          grpc_status: 0
      - assert:
          message:
            json:
              path: "$.entry.name"
              equals: demo
```

Full spec: [grpc](../examples/grpc.atago.yaml)

## Verify a generated page in a real browser

A `cdp:` step drives headless Chrome — when the CLI generates or serves HTML,
assert what a browser actually renders, not just what the bytes contain:

```yaml
version: "1"
suite:
  name: browser verification

runners:
  web:
    type: browser

scenarios:
  - name: the built site renders its title
    services:
      - name: server
        command: mytool serve --dir site --port 8080
        ready:
          file: server.pid
          timeout: 10s
    steps:
      - run:
          command: mytool build --out site
      - assert:
          exit_code: 0
      - cdp:
          runner: web
          actions:
            - navigate: http://127.0.0.1:8080/
            - wait_visible: h1
            - text: h1
      - assert:
          value:
            contains: Welcome
```

Full spec: [browser](../examples/browser.atago.yaml)

## Lay out specs for a growing suite

One file per contract area, a shared directory per product, tags for cost.
`atago run` takes directories, so CI never lists files:

```text
specs/
  cli/
    usage.atago.yaml        # help, version, misuse exit codes
    convert.atago.yaml      # the core command, happy + failure paths
    config.atago.yaml       # env vars, config files, defaults
  server/
    serve.atago.yaml        # services: + http asserts
    publish.atago.yaml      # side effects observed over HTTP
  interactive/
    wizard.atago.yaml       # pty: sessions
```

```shell
atago run ./specs                 # everything
atago run ./specs/cli             # one area
atago run --tag smoke ./specs     # one cost slice
atago list ./specs                # what exists, without running it
```

A spec file is also documentation: `atago doc --out docs/specs.md ./specs`
renders every scenario, fixture, and expectation as Markdown — this cookbook's
sibling [real-world docs](real-world.md) are generated exactly that way.

## Record an interactive session instead of scripting it

You don't hand-write expect/send choreography either. `record --pty` runs the
tool in a real terminal, you drive it once by hand, and the keystrokes become a
replayable `pty:` step — a password prompt turns into an `${env:...}`
placeholder instead of a recorded secret:

```shell
atago record --pty --out wizard.atago.yaml -- mytool init
```

```yaml
version: "1"
suite:
  name: recorded wizard
scenarios:
  - name: mytool init
    steps:
      - pty:
          command: mytool init
          session:
            - expect: "Project name:"
            - send: "demo\n"
            - expect: "Password:"
            - send: "${env:MYTOOL_PASSWORD}\n"   # record masked this for you
            - expect: "created demo/"
      - assert:
          exit_code: 0
```

Recording works on Linux, macOS, and Windows (ConPTY). Trim the session to the
exchanges that matter — fewer expects means less brittleness.

## Refresh snapshots when output legitimately changes

A golden file is only as good as its update workflow. When a change is
intentional, one command re-records every snapshot the spec owns; volatile
details (temp paths, UUIDs, timestamps, ports, ANSI colors) are normalized at
compare time, so refreshed goldens stay stable across machines:

```yaml
version: "1"
suite:
  name: golden workflow

scrub:
  - {pattern: 'build [0-9a-f]{8}', placeholder: 'build <HASH>'}  # your own volatiles

scenarios:
  - name: the report matches its golden
    steps:
      - run:
          command: mytool report
      - assert:
          exit_code: 0
          stdout:
            snapshot: snapshots/report.txt
```

```shell
atago run report.atago.yaml                     # fails with a colorized diff
atago snapshot update report.atago.yaml         # accept the new output
git diff snapshots/                             # review exactly what changed
```

The failure diff and the `git diff` after updating are the same review — a
snapshot turns "is this output right?" into a code-review question.

## Pin the final TUI frame with a screen snapshot

`stdout` of a pty step is the full transcript — every redraw, every spinner
frame. `screen:` asserts the RENDERED terminal instead: what a user actually
sees after cursor movement and clears are applied. Snapshot it like any text:

```yaml
version: "1"
suite:
  name: tui frame

scenarios:
  - name: the dashboard's final frame is stable
    steps:
      - pty:
          command: mytool dashboard --once
          rows: 24
          cols: 80          # pin the geometry, or the frame wraps differently
      - assert:
          screen:
            line: 1
            contains: "MYTOOL DASHBOARD"
      - assert:
          screen:
            snapshot: snapshots/dashboard_screen.txt
```

One caveat: a TUI on the alternate screen buffer restores the primary screen at
clean exit, so capture while the UI is up (`--once` style flags help) rather
than after quitting.

Full spec: [pty_screen](../examples/pty_screen.atago.yaml)

## Prove a dry run changes nothing

`--dry-run` has exactly one contract: describe what would happen, touch
nothing. Per-file asserts cannot state the second half ("...and nothing
else"), but an all-empty `changes:` delta can — the step created, modified,
and deleted no files at all:

```yaml
version: "1"
suite:
  name: dry run

scenarios:
  - name: --dry-run reports the plan but touches nothing
    steps:
      - fixture:
          file: notes/draft.txt
          content: "keep me\n"
      - run:
          command: mytool clean --dry-run .
      - assert:
          exit_code: 0
          stdout:
            contains: "would delete notes/draft.txt"
          # The dry-run contract, stated exhaustively: no file anywhere in the
          # workdir was created, modified, or deleted.
          changes:
            created: []
            modified: []
            deleted: []
```

Full spec: [changes](../examples/changes.atago.yaml)

## Test config precedence

Every layered-config CLI documents "flag beats environment beats file", and
each layer is usually tested alone — which is exactly how precedence bugs
hide. Pin each rung against the one below it:

```yaml
version: "1"
suite:
  name: config precedence

scenarios:
  - name: the config file sets the default
    steps:
      - fixture:
          file: config.toml
          content: |
            region = "eu-west-1"
      - run:
          command: mytool region --config config.toml
          # A stray MYTOOL_REGION on the host must not leak into this rung.
          clear_env: true
          pass_env: [PATH]
      - assert:
          exit_code: 0
          stdout:
            contains: eu-west-1

  - name: an environment variable overrides the file
    steps:
      - fixture:
          file: config.toml
          content: |
            region = "eu-west-1"
      - run:
          command: mytool region --config config.toml
          clear_env: true
          pass_env: [PATH]
          env:
            MYTOOL_REGION: us-east-2
      - assert:
          exit_code: 0
          stdout:
            contains: us-east-2

  - name: a flag overrides both
    steps:
      - fixture:
          file: config.toml
          content: |
            region = "eu-west-1"
      - run:
          command: mytool region --config config.toml --region ap-northeast-1
          clear_env: true
          pass_env: [PATH]
          env:
            MYTOOL_REGION: us-east-2
      - assert:
          exit_code: 0
          stdout:
            contains: ap-northeast-1
```

Full spec: [hermetic_env](../examples/hermetic_env.atago.yaml)

## Abort a destructive command at its confirmation prompt

The "yes" branch of a destructive command gets all the attention; the safety
branch — answer no, nothing happens — is the one users bet their data on.
Drive the prompt in a real terminal and pin the delta to empty:

```yaml
version: "1"
suite:
  name: confirmation prompt

scenarios:
  - name: answering no leaves every file in place
    steps:
      - fixture:
          file: data.db
          content: "precious\n"
      - pty:
          command: mytool purge
          timeout: 10s
          session:
            # [y/N] is a regex character class — escape it.
            - expect: 'delete 1 file\? \[y/N\]'
            - send: "n\n"
      - assert:
          # Exit-code conventions for an aborted command vary; the files ARE
          # the contract, so assert those exhaustively instead.
          changes:
            created: []
            modified: []
            deleted: []
      - assert:
          file:
            path: data.db
            contains: precious
```

Full spec: [pty](../examples/pty.atago.yaml)

## Prove color output turns off

Raw ANSI escapes leaking into piped output is a classic CLI bug. Under a
`run:` step stdout already IS a pipe, so a well-behaved tool should emit no
color even before you reach for [`NO_COLOR`](https://no-color.org/):

```yaml
version: "1"
suite:
  name: color contract

scenarios:
  - name: a pipe gets no color by default
    steps:
      - run:
          command: mytool status
      - assert:
          exit_code: 0
          stdout:
            not_matches: '\x1b\['    # no CSI escape sequence anywhere

  - name: NO_COLOR is honored even where color would be forced
    steps:
      - run:
          command: mytool status --color always
          env:
            NO_COLOR: "1"
      - assert:
          exit_code: 0
          stdout:
            not_matches: '\x1b\['
```

To test the opposite branch — color IS emitted on a terminal — run the same
command under `pty:` and assert `matches: '\x1b\['`.

Full spec: [run_and_assert](../examples/run_and_assert.atago.yaml)

## Assert a generated script is executable

A scaffolder that writes hook or build scripts makes two promises: the content
and the mode bit. A `file:` assert checks one property at a time, so stack two:

```yaml
version: "1"
suite:
  name: executable bit

scenarios:
  - name: scaffold writes a runnable build script
    skip:
      os: windows        # no executable bit to assert there
    steps:
      - run:
          command: mytool scaffold --with-scripts
      - assert:
          exit_code: 0
          file:
            path: bin/build.sh
            executable: true
      - assert:
          file:
            path: bin/build.sh
            contains: "#!/bin/sh"
```

Full spec: [files_and_fixtures](../examples/files_and_fixtures.atago.yaml)

## Hunt down a flaky scenario

Flakiness survives by hiding behind retries. atago's flake tooling points the
other way — it surfaces instability instead of absorbing it:

```shell
atago run --repeat 20 flaky.atago.yaml   # run each scenario 20 times; ONE bad iteration fails the run
atago run --retry-failed 2 ./specs      # retry failures, but a pass-after-fail is REPORTED as flaky, never hidden
atago run --rerun-failed ./specs        # while bisecting: re-run only what failed last time (from .atago/last-failed.json)
```

The usual culprits, and the recipe that fixes each: sleeping instead of
polling ([Poll an async result](#poll-an-async-result)), volatile output
reaching a snapshot ([Pin output with a golden
file](#pin-output-with-a-golden-file)), host environment leaking in ([Isolate
the test from the host environment](#isolate-the-test-from-the-host-environment)),
and unpinned terminal geometry in TUI tests — set `rows:`/`cols:` on the
`pty:` step so the frame cannot wrap differently between machines.

## Troubleshooting

The failures every new spec hits once, and the fast way out of each.

### "executable file not found", but the command works in your shell

`command:` runs the program directly (argv) — no shell parses the line. Shell
builtins (`echo`, `cd`, `type`, and `dir` on Windows), pipes, redirects, and
globs exist only under `shell: true`. If the binary genuinely exists on PATH,
check whether the scenario cleared its environment: `clear_env: true` drops
`PATH` too, so re-admit it with `pass_env: [PATH]`.

### atago expands a `${...}` you meant literally

`${name}` is atago's own expansion syntax — stores, matrix parameters,
`${env:NAME}`, `${workdir}`. To pass a literal `${...}` through to the command
(a shell variable, a template placeholder), escape it as `$${...}`.

### The command cannot find its input files

Every scenario runs in a fresh, isolated working directory — atago does not
run your command where you launched atago. Ship each input with a `fixture:`
step and reference it relatively (or as `${workdir}/...`); do not point at
files in your repository checkout. See
[Ship binary test data with fixtures](#ship-binary-test-data-with-fixtures).

### A snapshot passes locally and fails in CI

The built-in normalizers already cover ANSI colors, temp paths, UUIDs,
timestamps, ports, and CRLF. Anything else that varies per run — build hashes,
auto-increment IDs, durations — needs a spec-wide `scrub:` rule
([Pin output with a golden file](#pin-output-with-a-golden-file)). When output
changed on purpose, `atago snapshot update` re-records and `git diff
snapshots/` is the review.

### A pty session hangs until its timeout

Three usual causes. A `send:` without a trailing `"\n"` types the text but
never presses enter. An `expect:` is a Go regex, so a prompt like `[y/N]` is a
character class until you escape it: `\[y/N\]`. And each `expect` scans only
the transcript AFTER the previous match — expecting text that appeared earlier
in the session waits forever for a second occurrence. The session dies at
`timeout` (default 30s) with the transcript in the failure, so read what the
terminal actually showed.

### A `screen:` assert sees a blank screen

Most full-screen TUIs use the alternate screen buffer and restore the primary
screen when they exit cleanly — so after a clean quit there is nothing left to
assert. Capture the frame while the UI is still displayed: assert before
sending the quit key, or use a `--once`-style flag that leaves output on the
primary screen.

### The run exits 2 or 3 before any scenario runs

atago's exit codes separate "your CLI failed the test" from "the test could
not run": `1` means scenarios failed, `2` means a spec did not load (YAML
syntax or schema violation — the message points at the offending key), `3`
means atago itself was misused (unknown flag, or no spec files matched the
path). On `2`, `atago explain spec.atago.yaml` validates and describes the
spec without running anything.

### It passes on your machine and fails on a teammate's

Something from the host environment is leaking in — a config file in `$HOME`,
a locale, an exported variable. Make the scenario hermetic: `clear_env: true`
with a `pass_env:` allowlist, and `sandbox_home: true` to isolate `$HOME` and
the per-OS config/cache directories. See
[Isolate the test from the host environment](#isolate-the-test-from-the-host-environment).

### When all else fails, watch what atago actually did

`atago run --verbose` traces every command, captured stream, and per-assertion
verdict — for passing scenarios too, which is what you want while authoring.
`atago explain spec.atago.yaml` describes a spec without running it, and
`atago list` shows which scenarios and tags a path contains.
