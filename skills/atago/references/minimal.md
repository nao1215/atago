# A minimal spec

Three ways to get a first spec. Each one runs green as written.

## Record a real run

`atago record` runs the command once in a scratch directory and writes a spec from what it observed: the exit code, the first stdout line, an empty stderr, and created files.

<!-- verify: run -->
```shell
atago record --out version.atago.yaml -- go version
atago run version.atago.yaml
```

For an interactive program, add `--pty` and drive one session by hand. For stdout pinned to a golden file, add `--snapshot`.

## Scaffold one

`atago init` writes a starter spec. `--template` picks `cli` (the default), `http`, `db`, `grpc`, `ssh`, `browser`, `mock`, or `services`.

<!-- verify: run -->
```shell
atago init --template cli starter.atago.yaml
atago run starter.atago.yaml
```

## Write one by hand

A fixture writes an input file into the scenario's temporary directory. The command runs there, and the asserts check what it produced. The first line gives an editor completion from the JSON Schema.

<!-- verify: spec sort.atago.yaml -->
```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
suite:
  name: sort
scenarios:
  - name: a fixture goes in and a file comes out
    steps:
      - fixture:
          file: names.txt
          content: |
            beta
            alpha
      - run:
          command: sort
          stdin:
            file: names.txt
          stdout_to: sorted.txt
      - assert:
          exit_code: 0
          file:
            path: sorted.txt
            contains: [alpha, beta]
```

<!-- verify: run -->
```shell
atago run sort.atago.yaml
```

Replace `sort` with your own program. If it is not on `PATH`, build it before the run, or declare it as the `subject:` of an `atago.project.yaml` next to the specs.
