# atago Behavior Specs
## Summary
1 suite · 7 scenarios
## Contents
- [rclone (self-hosted file sync program)](#rclone-self-hosted-file-sync-program) — 7 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [copy replicates a tree and check certifies the replica](#scenario-copy-replicates-a-tree-and-check-certifies-the-replica)
  - [check fails loudly once the replica is corrupted](#scenario-check-fails-loudly-once-the-replica-is-corrupted)
  - [lsjson emits a machine-readable listing](#scenario-lsjson-emits-a-machine-readable-listing)
  - [sync makes the destination mirror the source, deletions included](#scenario-sync-makes-the-destination-mirror-the-source-deletions-included)
  - [obscure and reveal round-trip a secret](#scenario-obscure-and-reveal-round-trip-a-secret)
  - [serve http publishes the tree over real HTTP](#scenario-serve-http-publishes-the-tree-over-real-http)
## rclone (self-hosted file sync program)
Source: `test/e2e/thirdparty/rclone/rclone.atago.yaml`
### Scenario: version prints a semantic version
#### When
```shell
rclone version
```
#### Then
- exit code is `0`
- stdout matches `/rclone v[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: copy replicates a tree and check certifies the replica
#### Given
- Fixture file `rclone.conf` is created.
- Fixture file `src/hello.txt` is created.
- Fixture file `src/sub/table.csv` is created.
#### Inputs
_Fixture `src/hello.txt`:_
```text
hello rclone
```
_Fixture `src/sub/table.csv`:_
```text
a,b
1,2
```
#### When
```shell
rclone copy src dst
rclone check src dst
```
#### Then
- after `rclone copy src dst`:
  - exit code is `0`
  - dir `dst` contains `hello.txt`, contains `sub/table.csv`
  - file `dst/hello.txt` contains `hello rclone`
- after `rclone check src dst`:
  - exit code is `0`
  - stderr contains `0 differences found`
### Scenario: check fails loudly once the replica is corrupted
#### Given
- Fixture file `rclone.conf` is created.
- Fixture file `src/hello.txt` is created.
- Fixture file `dst/hello.txt` is created.
#### Inputs
_Fixture `src/hello.txt`:_
```text
hello rclone
```
_Fixture `dst/hello.txt`:_
```text
corrupted replica
```
#### When
```shell
rclone copy src dst
rclone check src dst
```
#### Then
- after `rclone check src dst`:
  - exit code is not `0`
  - stderr contains `1 differences found`
### Scenario: lsjson emits a machine-readable listing
#### Given
- Fixture file `rclone.conf` is created.
- Fixture file `src/hello.txt` is created.
- Fixture file `src/sub/table.csv` is created.
#### Inputs
_Fixture `src/hello.txt`:_
```text
hello rclone
```
_Fixture `src/sub/table.csv`:_
```text
a,b
1,2
```
#### When
```shell
rclone lsjson src
rclone size --json src
```
#### Then
- after `rclone lsjson src`:
  - exit code is `0`
  - stdout at `$` has length 2
  - stdout at `$[0].Name` equals `hello.txt`
  - stdout at `$[0].Size` equals `12`
  - stdout at `$[1].IsDir` equals `true`
- after `rclone size --json src`:
  - exit code is `0`
  - stdout at `$.count` equals `2`
### Scenario: sync makes the destination mirror the source, deletions included
#### Given
- Fixture file `rclone.conf` is created.
- Fixture file `src/keep.txt` is created.
- Fixture file `dst/extraneous.txt` is created.
#### Inputs
_Fixture `src/keep.txt`:_
```text
stays
```
_Fixture `dst/extraneous.txt`:_
```text
must be deleted by sync
```
#### When
```shell
rclone sync src dst
```
#### Then
- exit code is `0`
- dir `dst` contains `keep.txt`, does not contain `extraneous.txt`
### Scenario: obscure and reveal round-trip a secret
#### Given
- Fixture file `rclone.conf` is created.
#### When
```shell
rclone obscure atago-test-secret
# capture ${obscured} from stdout
rclone reveal ${obscured}
```
#### Then
- after `rclone obscure atago-test-secret`:
  - exit code is `0`
- after `rclone reveal ${obscured}`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: serve http publishes the tree over real HTTP
#### Given
- Background service `rclone-http` is started: `rclone serve http src --addr 127.0.0.1:18110`.
- Fixture file `rclone.conf` is created.
- Fixture file `src/hello.txt` is created.
- Fixture file `src/api/data.json` is created.
#### Inputs
_Fixture `src/hello.txt`:_
```text
hello over http
```
_Fixture `src/api/data.json`:_
```text
{"source":"rclone","ok":true}
```
#### When
```shell
# HTTP GET /hello.txt
# HTTP GET /api/data.json
# HTTP GET /no-such-file.txt
```
#### Then
- after `HTTP GET /hello.txt`:
  - HTTP status is `200`
  - body contains `hello over http`
- after `HTTP GET /api/data.json`:
  - HTTP status is `200`
  - body at `$.source` equals `rclone`
- after `HTTP GET /no-such-file.txt`:
  - HTTP status is `404`
