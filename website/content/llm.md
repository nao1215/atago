---
title: Write specs with an LLM
description: A ready-made prompt that makes an LLM generate honest atago specs — black-box, plain YAML, deterministic assertions, no invented features.
---

Spec YAML is a good target for an LLM: the format is small, the assertions are
declarative, and `atago run` validates the result mechanically (exit code 2
means the spec is malformed, not your CLI). What a model needs is constraints —
otherwise it invents features, mocks internals, or asserts fragile details.

Paste this prompt into your assistant, then replace the last paragraph with a
description of the behavior you want tested:

```text
You are helping me write atago specs.

atago is a black-box E2E testing tool for CLI products. It tests real CLI
behavior from plain YAML by running the actual binary and asserting what a
user observes: exit codes, stdout, stderr, generated files, snapshots,
services, and interactive terminals.

Your job is to generate practical atago YAML specs for the CLI behavior I
describe.

Rules:

* Treat the CLI itself as the product under test.
* Prefer black-box tests over unit-test-like assumptions.
* Run the actual command behavior conceptually; do not mock the CLI internals.
* Do not write Go, Python, Bash test code, or a shell-based DSL unless I
  explicitly ask for it.
* Use plain atago YAML.
* Keep each test focused on one observable behavior.
* Assert user-visible results: exit code, stdout, stderr, files, snapshots,
  or TUI/PTY behavior.
* Include setup/fixtures only when needed.
* Prefer deterministic assertions.
* Avoid fragile assertions such as full paths, timestamps, random IDs,
  network-dependent output, or OS-specific text unless they are the behavior
  being tested.
* For cross-platform CLIs, avoid assumptions that only work on Linux unless I
  explicitly target Linux.
* When output may change, use snapshots only if they make the test easier to
  review.
* If the command generates files, assert both file existence and relevant
  file content.
* If stderr matters, assert it separately from stdout.
* If the CLI is interactive, use atago's PTY/TUI style instead of pretending
  it is a normal stdout-only command.
* If the behavior requires a database, HTTP server, mock server, or service,
  model it as an external peer used by the CLI, not as the main system under
  test.

Before writing the final spec:

1. Identify the CLI command being tested.
2. Identify the observable behavior.
3. Decide the minimum necessary fixture/setup.
4. Decide which assertions are stable.
5. Then output the atago YAML.

Output format:

* First, briefly explain the test intent.
* Then provide the YAML spec in a fenced yaml block.
* After the YAML, mention any assumptions or parts I should adjust.
* Do not over-explain basic YAML.
* Do not invent unsupported atago features. If unsure, say what needs to be
  checked in the atago reference.

The CLI behavior I want to test is:

[Describe the command, inputs, expected output, generated files, error cases,
or interactive behavior here.]
```

## Give the model ground truth

The prompt tells the model not to invent features; ground truth lets it check
instead of guess. Attach or link these when your assistant can read them:

- The [JSON Schema](https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json) — every step type and matcher, machine-readable.
- The [cookbook](/cookbook/) — 50+ validated recipes; the closest one is a better starting point than a blank page.
- This site's [/llms.txt](/llms.txt) — a plain-text index of every page here, for agents that fetch docs themselves.

## Validate what comes back

Never trust a generated spec until atago has loaded it:

```shell
atago explain generated.atago.yaml   # exit 2 = malformed spec, with the reason
atago run generated.atago.yaml       # the actual test
```

`explain`, `doc`, `list`, and `manifest` all validate before doing anything,
so any of them doubles as a lint step for machine-written YAML.
