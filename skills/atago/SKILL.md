---
name: atago
description: Write, fix, or debug black-box E2E tests for a CLI with atago, whose tests are *.atago.yaml specs. Use it when you add or change an atago spec, read a failing atago run, or update its snapshots.
---

# Testing a CLI with atago

atago runs a real program and checks what a user observes: the exit code, stdout, stderr, files, and snapshots. A test is a YAML spec named `*.atago.yaml`. You write YAML, not test code.

## The loop

1. Look up the keys you need: https://nao1215.github.io/atago/reference/#spec-file-keys. Do not guess a key. Specs are decoded strictly, so an unknown key is an error (`ATG2005`).
2. Start from a real run or from a recipe. `atago record --out mytool.atago.yaml -- mytool --version` runs the command once in a scratch directory and writes a spec from what it observed. Tighten that spec. A command that reads a local file needs a `fixture:` step that writes the file into the scenario's workdir (examples/files_and_fixtures.atago.yaml).
3. Run the spec with `atago run`. Fix the spec or the program and run it again. Stop only when it exits 0.
4. When the output changed on purpose, update the snapshots and review the diff (see Snapshots below).

<!-- verify: spec greet.atago.yaml -->
```yaml
version: "1"
suite:
  name: greet
scenarios:
  - name: echo greets the world
    steps:
      - run:
          shell: true
          command: echo hello atago
      - assert:
          exit_code: 0
          stdout:
            contains: atago
          stderr:
            empty: true
```

<!-- verify: run -->
```shell
atago run greet.atago.yaml
```

<!-- verify: output -->
```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped (2ms)
```

If `atago` is not installed: `go install github.com/nao1215/atago@latest`.

## What every spec needs

- `version: "1"`, a `suite` with a `name`, and `scenarios`. Each scenario has a `name` and `steps`.
- A step does one thing: `run`, `fixture`, `assert`, `store`, `pty`, `service`, `signal`, `http`, `query`, `grpc`, `cdp`, or `mock_server`.
- Each scenario runs in its own temporary directory, not where you started atago. Ship every input with a `fixture:` step and refer to it by a relative path. Do not point at files in the repository checkout.
- Without `shell: true`, atago splits `run.command` into argv itself. There are no pipes or redirects. Use `run.stdin`, `run.env`, `run.stdout_to`, and `run.stderr_to` instead.
- Quote a whole command that contains `: `. Otherwise YAML reads it as a mapping (`ATG2003`).

A minimal spec and the commands that make one are in references/minimal.md. The common matchers are in references/matchers.md.

## Reading a failure

The exit code says what kind of problem it is.

| Code | Meaning |
|------|---------|
| `0` | no unexpected failure: every scenario passed, was skipped, or was an XFAIL (an `expect_fail:` scenario that failed as declared) |
| `1` | one or more scenarios failed, an XPASS (an `expect_fail:` scenario that passed), or a flaky scenario that passed only on a retry |
| `2` | spec error: YAML syntax or schema validation |
| `3` | command-line error: unknown subcommand, bad flag, or no matching spec files |
| `4` | execution error: a step could not be carried out |
| `5` | internal error: a bug in atago |
| `6` | security policy violation |

- Exit 1 is a result, not an error: a failed assertion, an XPASS (unless `--allow-xpass`), or a flaky pass (unless `--allow-flaky`). The report names the scenario, the step, the command, and a diff of expected and actual. It carries no code.
- An error carries a code such as `ATG2005`. Its first digit is the exit code. `atago explain ATG2005` says what it means and what to change. All codes: https://nao1215.github.io/atago/errors/.
- `atago run --verbose` traces every command, its output, and each assertion verdict, for passing scenarios too.
- `atago run --rerun-failed` runs only the scenarios that failed last time.
- `atago run --report json` gives the same results as JSON. A coded failure has a `code` field.

Worked examples are in references/failures.md.

## Snapshots

A `snapshot:` matcher compares output against a committed golden file. ANSI colors, temp paths, UUIDs, timestamps, ports, and CRLF are normalized. For any other volatile text, add a spec-wide `scrub:` rule.

<!-- verify: spec help.atago.yaml -->
```yaml
version: "1"
suite:
  name: help
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
scenarios:
  - name: help output stays stable
    steps:
      - run:
          shell: true
          command: echo usage id=42
      - assert:
          stdout:
            snapshot: snapshots/help.txt
```

A missing golden file fails the run. Record it, then run again:

<!-- verify: run -->
```shell
atago snapshot update help.atago.yaml
atago run help.atago.yaml
```

When a snapshot fails, read the diff first. Update only when the new output is what you want, then review exactly what changed:

```shell
atago run help.atago.yaml
atago snapshot update help.atago.yaml
git diff snapshots/
```

`atago run --update-snapshots` does the same as `atago snapshot update`.

## Recipes

- Cookbook, tested recipes for common jobs: https://nao1215.github.io/atago/cookbook/. Start from a recipe and change the names.
- One runnable spec per feature: https://github.com/nao1215/atago/tree/main/examples.
- `atago init --template <name>` scaffolds a spec for `cli`, `http`, `db`, `grpc`, `ssh`, `browser`, `mock`, or `services`. `atago init --list-templates` lists them.
- JSON Schema for editor completion: https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json. Put `# yaml-language-server: $schema=<that URL>` on the first line of a spec.

## Do not

- Do not update snapshots without reading the failure diff, and do not commit updated golden files without reviewing `git diff`.
- Do not weaken or delete an assertion to make a run pass. A known bug that is not fixed yet goes under `expect_fail:`, so the scenario still runs.
- Do not hide flakiness. `--retry-failed N` reports a recovered scenario as flaky and still fails the run. Add `--allow-flaky` only when the instability is known and accepted.
- Do not stop a background process by name (`kill`, `pkill`). Use a `signal:` step on a `service:`.
- Do not hard-code a port two parallel scenarios can share. Scenarios run concurrently and share the host network.
- Do not say the tests pass while `atago run` exits non-zero.
