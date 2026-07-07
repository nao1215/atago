# Cookbook

Task-oriented YAML recipes: find what you want to do, copy the spec, replace
`mytool` with your binary. Each recipe links to a runnable, CI-tested spec
under [examples/](../examples/) that covers the same feature in full.

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
