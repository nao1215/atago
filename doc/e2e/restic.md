# atago Behavior Specs
## Summary
1 suite · 7 scenarios
## Contents
- [restic (self-hosted backup program)](#restic-self-hosted-backup-program) — 7 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [init creates an encrypted repository on disk](#scenario-init-creates-an-encrypted-repository-on-disk)
  - [backup, list as JSON, and restore round-trip user data](#scenario-backup-list-as-json-and-restore-round-trip-user-data)
  - [diff names exactly the file added between two snapshots](#scenario-diff-names-exactly-the-file-added-between-two-snapshots)
  - [check reports a healthy repository as error-free](#scenario-check-reports-a-healthy-repository-as-error-free)
  - [forget --keep-last 1 --prune drops the older snapshot](#scenario-forget---keep-last-1---prune-drops-the-older-snapshot)
  - [the wrong password cannot unlock the repository](#scenario-the-wrong-password-cannot-unlock-the-repository)
## restic (self-hosted backup program)
Source: `test/e2e/thirdparty/restic/restic.atago.yaml`
### Scenario: version prints a semantic version
#### When
```shell
restic version
```
#### Then
- exit code is `0`
- stdout matches `/restic [0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: init creates an encrypted repository on disk
#### When
```shell
restic init
```
#### Then
- exit code is `0`
- stdout contains `created restic repository`
- dir `repo` exists, contains `config`, contains `keys`, contains `snapshots`
### Scenario: backup, list as JSON, and restore round-trip user data
#### Given
- Fixture file `data/hello.txt` is created.
- Fixture file `data/sub/notes.md` is created.
#### Inputs
_Fixture `data/hello.txt`:_
```text
hello from atago
```
_Fixture `data/sub/notes.md`:_
```text
# notes kept safe
```
#### When
```shell
restic init
restic backup data
restic snapshots --json
# capture ${snap} from stdout
restic restore ${snap} --target restored
```
#### Then
- after `restic backup data`:
  - exit code is `0`
  - stdout matches `/snapshot [0-9a-f]{8} saved/`
- after `restic snapshots --json`:
  - exit code is `0`
  - stdout at `$` has length 1
  - stdout at `$[0].summary.total_files_processed` equals `2`
- after `restic restore ${snap} --target restored`:
  - exit code is `0`
  - stdout contains `Restored`
  - file `restored/data/hello.txt` contains `hello from atago`
  - file `restored/data/sub/notes.md` contains `# notes kept safe`
### Scenario: diff names exactly the file added between two snapshots
#### Given
- Fixture file `data/base.txt` is created.
- Fixture file `data/added-later.txt` is created.
#### Inputs
_Fixture `data/base.txt`:_
```text
first generation
```
_Fixture `data/added-later.txt`:_
```text
second generation
```
#### When
```shell
restic init
restic backup data
restic backup data
restic snapshots --json
# capture ${first} from stdout
# capture ${second} from stdout
restic diff ${first} ${second}
```
#### Then
- after `restic snapshots --json`:
  - exit code is `0`
  - stdout at `$` has length 2
- after `restic diff ${first} ${second}`:
  - exit code is `0`
  - stdout contains `+    /data/added-later.txt`
  - stdout does not contain `/data/base.txt`
### Scenario: check reports a healthy repository as error-free
#### Given
- Fixture file `data/precious.txt` is created.
#### Inputs
_Fixture `data/precious.txt`:_
```text
verify me
```
#### When
```shell
restic init
restic backup data
restic check
```
#### Then
- after `restic check`:
  - exit code is `0`
  - stdout contains `no errors were found`
### Scenario: forget --keep-last 1 --prune drops the older snapshot
#### Given
- Fixture file `data/gen.txt` is created.
- Fixture file `data/gen.txt` is created.
#### Inputs
_Fixture `data/gen.txt`:_
```text
one
```
_Fixture `data/gen.txt`:_
```text
two
```
#### When
```shell
restic init
restic backup data
restic backup data
restic forget --keep-last 1 --prune
restic snapshots --json
```
#### Then
- after `restic forget --keep-last 1 --prune`:
  - exit code is `0`
  - stdout contains `remove 1 snapshots`
- after `restic snapshots --json`:
  - exit code is `0`
  - stdout at `$` has length 1
### Scenario: the wrong password cannot unlock the repository
#### Given
- Environment variables are set: RESTIC_PASSWORD.
#### When
```shell
restic init
restic snapshots
```
#### Then
- after `restic snapshots`:
  - exit code is not `0`
  - stderr contains `wrong password or no key found`