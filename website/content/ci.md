---
title: Use it in CI
description: Run atago suites in CI with JUnit/JSON/TAP/GitHub reports, loud retries for flaky scenarios, repeat-based flake detection, kept artifacts, and secret masking.
---

Real E2E suites flake (timing, ports, external tools). `--retry-failed N` re-runs failed scenarios in a fresh workdir and reports recovered ones as flaky — green for the exit code, but loud in every report format; silent retries are explicitly a non-goal. `--repeat N` does the opposite job: run each scenario N times to detect flakiness before it reaches CI.

```shell
atago run --ci --retry-failed 2 ./specs          # keep CI green, report instability loudly
atago run --repeat 20 --filter "race prone" ./specs   # flake detection
```

[setup-atago](https://github.com/nao1215/setup-atago) installs a released binary:

```yaml
name: behavior-specs
on: [push, pull_request]
jobs:
  atago:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: nao1215/setup-atago@v0
      - run: atago run --ci --report gha ./specs
```

- `--report json|junit|gha|tap` picks the report format; the JSON shape is stable and versioned ([sample JSON](/samples/report.json), [JUnit](/samples/report.junit.xml), [TAP](/samples/report.tap)).
- `--ci` enables deterministic, color-free output.
- `--artifacts-dir DIR` persists the exact payloads a failed assertion compared, so a failure stays reviewable after the job ends.
- Environment variable names listed under `secrets:` are masked as `***` in all reports and snapshots.

## Review specs without running them

`explain` describes what a spec does, `doc` generates Markdown (with fixtures, expected payloads, and golden files inlined), `manifest` emits a stable JSON summary for tooling, and `list` shows scenarios, tags, and artifacts. All of them load and validate the spec first — exit code 2 on a schema error — so any of them doubles as a lint step in CI:

![atago explain and doc rendering a spec](/img/review.gif)

```shell
atago explain spec.atago.yaml
atago doc --out docs/specs.md ./specs
atago manifest ./specs
atago list ./specs
```

The [real-world pages](/real-world/) on this site are `atago doc` output, committed and drift-tested.
