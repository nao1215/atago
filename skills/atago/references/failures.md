# Reading a failure

## A failed assertion (exit 1)

<!-- verify: spec shout.atago.yaml -->
```yaml
version: "1"
suite:
  name: shout
scenarios:
  - name: greets in capitals
    steps:
      - run:
          shell: true
          command: echo hello
      - assert:
          stdout:
            equals: HELLO
```

<!-- verify: run exit=1 -->
```shell
atago run shout.atago.yaml
```

<!-- verify: output -->
```text
F

FAILED: shout / greets in capitals  (shout.atago.yaml)

Step:
  assert stdout equals exact text

Command:
  echo hello

Expected:
  HELLO

Actual:
  hello

Hint:
  stdout did not equal the expected text

FAILED  1 scenario: 0 passed, 1 failed, 0 errored, 0 skipped (3ms)
```

Read it top down: the suite and scenario, the step that failed, the command it ran, then expected against actual. Output of several lines is shown as a unified diff (`-expected +actual`). Decide which side is wrong. If the program is wrong, fix the program. If the expectation is wrong, fix the spec. Do not loosen the matcher until it passes.

To see every command, its output, and each assertion verdict, run `atago run --verbose shout.atago.yaml`.

## A spec error (exit 2)

<!-- verify: spec typo.atago.yaml invalid -->
```yaml
version: "1"
suite:
  name: typo
scenarios:
  - name: a misspelled key
    steps:
      - run:
          command: sort
          sell: true
```

<!-- verify: run exit=2 -->
```shell
atago run typo.atago.yaml
```

<!-- verify: output -->
```text
typo.atago.yaml: ATG2005: [9:11] unknown field "sell"
   6 |     steps:
   7 |       - run:
   8 |           command: sort
>  9 |           sell: true
                 ^
hint: did you mean "shell"?
```

The line is `file: code: [line:column] message`. The code explains itself:

<!-- verify: run -->
```shell
atago explain ATG2005
```

<!-- verify: output -->
```text
ATG2005  the spec sets a key the schema does not define
```

`atago explain` also takes a spec path and describes what the spec does without running it. `atago list`, `atago doc`, and `atago manifest` load and validate the spec first, so each of them exits 2 on a spec error too.

## Codes

- The first digit of `ATGxxxx` is the exit code: `ATG2xxx` is a spec error, `ATG3xxx` a command-line error, `ATG4xxx` an execution error.
- Common ones: `ATG2003` YAML syntax (quote a value that contains `: `), `ATG2005` unknown key, `ATG2102` a step with more than one action, `ATG2201` a required key is missing, `ATG2305` a regular expression that does not compile.
- Every code, with what to change: https://nao1215.github.io/atago/errors/.

## A missing snapshot

A `snapshot:` whose golden file does not exist fails with `Actual: missing`. Record it with `atago snapshot update <spec>`, read the new file, and commit it.
