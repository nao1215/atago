# Common matchers

An `assert:` step checks one or more targets: `exit_code`, `stdout`, `stderr`, `file`, `dir`, and more. Every target set in one step must hold. The full list of keys: https://nao1215.github.io/atago/reference/#spec-file-keys.

The spec below runs green as written. Each scenario shows one group of matchers.

<!-- verify: spec matchers.atago.yaml -->
```yaml
version: "1"
suite:
  name: matchers
scenarios:
  - name: text matchers on a stream
    steps:
      - run:
          shell: true
          command: echo release 1.2.3 ready
      - assert:
          exit_code: 0
          stderr:
            empty: true
          # contains, not_contains, matches, and not_matches can be set together.
          stdout:
            contains: ready
            not_contains: [error, panic]
            matches: '[0-9]+\.[0-9]+\.[0-9]+'
      # equals is the whole stream, without the trailing newline.
      - assert:
          stdout:
            equals: release 1.2.3 ready
      # line narrows the stream to one 1-based line before matching.
      - assert:
          stdout:
            line: 1
            matches: '^release '
      # count bounds how many times a contains or matches occurs.
      - assert:
          stdout:
            contains: ready
            count: 1

  - name: exit code sets
    steps:
      - run:
          shell: true
          command: exit 2
      - assert:
          exit_code:
            in: [0, 2]
      - assert:
          exit_code:
            not: 1

  - name: files, JSON, and directories
    steps:
      - fixture:
          file: out/report.json
          content: |
            {"name":"ada","items":[1,2,3]}
      - assert:
          file:
            path: out/report.json
            exists: true
      - assert:
          file:
            path: out/report.json
            contains: ada
      # json takes a JSONPath and a check, or a list of them.
      - assert:
          file:
            path: out/report.json
            json:
              - {path: "$.name", equals: ada}
              - {path: "$.items", length: 3}
      - assert:
          dir:
            path: out
            contains: [report.json]
            not_contains: [secret.key]
```

<!-- verify: run -->
```shell
atago run matchers.atago.yaml
```

More matchers, one runnable spec each:

- `snapshot:` and `scrub:`, golden files: examples/snapshot.atago.yaml, examples/scrub.atago.yaml.
- `permutation_of` and `subset_of`, lines in any order: examples/run_and_assert.atago.yaml.
- `size`, `min_count`, `max_count`: examples/count_and_size.atago.yaml.
- `json` with `gt`, `lt`, and `yaml`: examples/json_and_yaml.atago.yaml.
- `pty:` sessions and `screen:` for prompts and TUIs: examples/pty.atago.yaml, examples/pty_screen.atago.yaml.
- `store:` and `${name}`, a value from one step used in a later one: examples/store_and_variables.atago.yaml.

The examples live at https://github.com/nao1215/atago/tree/main/examples.

Traps:

- `matches`, `not_matches`, and a pty `expect:` are Go regular expressions. Escape `[`, `(`, `.`, and `?` when you mean them literally. Write them in single quotes so YAML keeps the backslashes.
- A `file:` assert takes one of `exists`, `contains`, `not_contains`, `equals`, `equals_file`, `json`, `snapshot`, or `executable` (`ATG2104`). Use one assert step for each.
- `empty: false` asserts the stream is not empty.
- A list under `contains` is one matcher: every element must be present.
