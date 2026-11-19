# Testing

atago is tested at three levels: Go unit tests for each internal package, atago specs that run the built binary against itself and against real third-party programs, and techniques that generate or alter inputs to find cases nobody wrote down. This page says what each technique checks, where it lives, and how CI runs it.

Run the everyday checks from the repository root:

```sh
make test   # unit tests, including the generated properties
make lint   # golangci-lint for every target OS
make e2e    # the self-hosted specs and the git suite
```

## Unit tests

Every package under `internal/` has table-driven tests next to its code. `make test` runs them with coverage, and `make coverage` adds the runtime coverage of the self-hosted E2E run. The Coverage badge in the README is that merged number.

## Self-hosted E2E

[`test/e2e/atago/`](../test/e2e/atago/) holds atago specs that drive the freshly built binary: exit codes, every report format, snapshots, pty sessions, services, the loader's error codes. A behavior a user can observe is pinned here as well as in a unit test. `make e2e` runs them, and the E2E workflows run them on Linux, macOS and Windows.

## Third-party E2E

[`test/e2e/thirdparty/`](../test/e2e/thirdparty/) runs atago against real programs (git, bats, ShellSpec, rsync, lazygit, and others), and `make dogfood` against nao1215's own CLIs. These catch the gaps a spec language only shows when it meets a program it did not write. The ThirdParty and Dogfood workflows run them daily.

## Snapshots

atago's own `snapshot:` assertion is used by the self-hosted specs (for example [`snapshot_normalization.atago.yaml`](../test/e2e/atago/snapshot_normalization.atago.yaml)), so a change to rendered output shows up as a snapshot diff that has to be accepted with `--update-snapshots`. The snapshot store itself has unit and fuzz tests in [`internal/snapshot/`](../internal/snapshot/).

## Differential migrations

[`test/e2e/migration/`](../test/e2e/migration/) keeps the original Bats and ShellSpec suites next to the atago specs the migration guide turns them into. The Migration workflow runs both sides against the same programs, so the guide and atago's behavior cannot drift from what the original tools do.

## Generated properties

[`internal/report/property_test.go`](../internal/report/property_test.go) uses [rapid](https://github.com/flyingmutant/rapid) to generate whole runs: one to three suites, any mix of the seven scenario statuses, and names assembled from the characters each format has to escape (XML metacharacters, TAP's `#` and backslash, newlines, control bytes, wide runes). For every run it checks that

- the JSON report lists each scenario in order with its exact name and status, and `failures[]` holds exactly the failed, errored and xpass scenarios,
- the JUnit counters match the run and add up across suites, and an XML-legal name comes back unchanged,
- the TAP plan and points match the run one to one, and a TAP 14 harness reads back the right directive and the original description,
- the GHA notice and the console summary carry the same tally.

The properties run in `make test` with rapid's default 100 cases. A failure prints the seed and a shrunk counterexample; rerun it with `go test ./internal/report -run <Name> -rapid.seed=<seed>`, or search longer with `-rapid.checks=10000`.

## Fuzzing

Go fuzz targets cover the inputs atago parses: spec YAML and the loader, `${...}` expansion, globs and JSON checks, the snapshot store, recorded sessions, terminal rendering and kitty graphics (`go test -list '^Fuzz' ./...` lists them). The Fuzz workflow runs each target nightly. A crasher is kept under the package's `testdata/fuzz/` as a regression input.

## Mutation testing

[gremlins](https://github.com/go-gremlins/gremlins) v0.6.0 rewrites one expression at a time (a comparison, an arithmetic operator, an increment) and reruns the package tests. A mutant that no test notices "lived" and marks logic no assertion looks at; some of them are equivalent (the rewrite changes nothing observable), so the score is not expected to reach 100%.

The [Mutation workflow](../.github/workflows/mutation.yml) mutates every line a pull request changes, from the module root, and mutates the packages below in full every week, one job per package. The process runners, the engine and the CLI wiring are not in the weekly list: their tests start real processes and terminals, so a run per mutant would take hours. Neither run is a gate yet; the numbers below are the baseline a later gate must not fall under.

Reproduce one package locally with `gremlins unleash --workers 4 --timeout-coefficient 10 ./internal/<pkg>`, and a change with `gremlins unleash --diff origin/main .` (run from the root: started from `./internal/<pkg>`, `--diff` matches no file and reports 0 mutants).

First full measurement, 2026-11-19, on main between v0.26.1 and v0.26.2 (4 workers, timeout coefficient 10). Score is killed / (killed + lived); not covered mutants sit on lines no test executes.

| Package | Killed | Lived | Not covered | Score |
|---|---:|---:|---:|---:|
| `internal/artifact` | 12 | 0 | 8 | 100.0% |
| `internal/assert` | 615 | 103 | 119 | 85.7% |
| `internal/assertdesc` | 123 | 2 | 78 | 98.4% |
| `internal/diag` | 9 | 0 | 33 | 100.0% |
| `internal/explain` | 84 | 26 | 12 | 76.4% |
| `internal/fixture` | 18 | 4 | 3 | 81.8% |
| `internal/fsdelta` | 14 | 1 | 2 | 93.3% |
| `internal/generator` | 40 | 6 | 1 | 87.0% |
| `internal/loader` | 785 | 105 | 62 | 88.2% |
| `internal/manifest` | 47 | 3 | 5 | 94.0% |
| `internal/plural` | 6 | 0 | 4 | 100.0% |
| `internal/report` | 295 | 45 | 21 | 86.8% |
| `internal/scrub` | 4 | 0 | 0 | 100.0% |
| `internal/security` | 56 | 5 | 2 | 91.8% |
| `internal/snapshot` | 22 | 1 | 0 | 95.7% |
| `internal/spec` | 319 | 22 | 143 | 93.5% |
| `internal/store` | 16 | 3 | 0 | 84.2% |
| `internal/yaml` | 514 | 123 | 82 | 80.7% |
| Total | 2979 | 449 | 575 | 86.9% |

With the generated properties above, `internal/report` kills 297 and leaves 43 (87.4%): the two extra kills are the JUnit counters for unloadable specs and security violations.

## Benchmarks

[`bench/himorime.yaml`](../bench/himorime.yaml) measures the binary with [himorime](https://github.com/nao1215/himorime). On a pull request the Benchmark workflow measures the base and the head in the same job and fails on a significant slowdown.
