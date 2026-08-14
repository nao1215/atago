# Migration parity suites

Backing tests for the website's [migration guide](https://nao1215.github.io/atago/migrate/).
The same behavior is specified three times:

- `bats/` — the original tests written for [Bats](https://github.com/bats-core/bats-core)
- `shellspec/` — the original tests written for [ShellSpec](https://github.com/shellspec/shellspec)
- `*.atago.yaml` — the migrated atago specs, one suite per topic

CI (`.github/workflows/migration.yml`) runs all three: the originals under
pinned releases of Bats and ShellSpec, the migrations under the freshly built
atago binary. A mapping shown on the website therefore exists as executable
code on both sides — if either side stops passing, the guide is wrong and CI
says so. The hermetic drift test in `drift_test.go` additionally runs the
atago specs through the engine on every platform the unit tests cover.

The topics mirror the guide's sections:

| Topic | Bats | ShellSpec | atago |
|-------|------|-----------|-------|
| exit code, stdout, stderr | `bats/basics.bats` | `shellspec/spec/basics_spec.sh` | `basics.atago.yaml` |
| setup/teardown → fixtures, JSON | `bats/fixtures.bats` | `shellspec/spec/fixtures_spec.sh` | `fixtures_and_files.atago.yaml` |
| parameterized tests → matrix | `bats/parameters.bats` | `shellspec/spec/parameters_spec.sh` | `matrix.atago.yaml` |
| polling → retry | `bats/retry.bats` | `shellspec/spec/retry_spec.sh` | `retry.atago.yaml` |
| skip and tags | `bats/selection.bats` | `shellspec/spec/selection_spec.sh` | `selection.atago.yaml` |
| golden files → snapshots | `bats/snapshot.bats` | `shellspec/spec/snapshot_spec.sh` | `snapshot.atago.yaml` |

Run them locally:

```shell
atago run ./test/e2e/migration                  # the migrated specs
bats test/e2e/migration/bats                    # needs bats-core >= 1.8.0
(cd test/e2e/migration/shellspec && shellspec)  # needs shellspec
```
