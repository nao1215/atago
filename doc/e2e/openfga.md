# atago Behavior Specs
## Summary
1 suite · 7 scenarios
## Contents
- [openfga (fine-grained authorization CLI and server)](#openfga-fine-grained-authorization-cli-and-server) — 7 scenarios
  - [version prints something sane](#scenario-version-prints-something-sane)
  - [model test runs a local store file without a server](#scenario-model-test-runs-a-local-store-file-without-a-server)
  - [a store file's references are contained to its directory](#scenario-a-store-files-references-are-contained-to-its-directory)
  - [the type-count limit fails loudly and a sufficient limit passes](#scenario-the-type-count-limit-fails-loudly-and-a-sufficient-limit-passes)
  - [store create returns the new store as JSON](#scenario-store-create-returns-the-new-store-as-json)
  - [tuples write and delete against a live store](#scenario-tuples-write-and-delete-against-a-live-store)
  - [store import builds a store, a model, and tuples from one file](#scenario-store-import-builds-a-store-a-model-and-tuples-from-one-file)
## openfga (fine-grained authorization CLI and server)
[OpenFGA](https://openfga.dev/) is a relationship-based authorization
system: the `fga` CLI authors and tests authorization models, and the
`openfga` server answers the queries. The scenarios keep parity with the
CLI's own integration suite, in two halves.

The local half needs no server: `fga model test` runs a store file's
checks in-process, resolves `**` glob patterns, keeps a store file's
`model_file`/`tuple_file` references contained to the store file's
directory unless `--allow-external-files` says otherwise, and enforces
the type-count limit with a diagnostic naming the configured bound.

The server half boots a real `openfga` on scenario-owned ports and
drives the full authoring flow over its API: creating a store (JSON
out), writing and deleting relationship tuples from JSON and JSONL
files (including a conditional tuple), the `--hide-imported-tuples`
counts-only report, and `store import` building a store, a model, and
tuples from one file.

Source: `test/e2e/thirdparty/openfga/openfga.atago.yaml`
Network policy: egress is allowed only to `127.0.0.1`.
### Scenario: version prints something sane
_only when `fga --version` succeeds_
#### When
```shell
fga --version
```
#### Then
- exit code is `0`
- stdout contains `fga version`
### Scenario: model test runs a local store file without a server
_only when `fga --version` succeeds_
#### Given
- Fixture file `tests/fixtures/basic-model.fga` is created.
- Fixture file `tests/fixtures/basic-tuples.json` is created.
- Fixture file `tests/fixtures/basic-store.fga.yaml` is created.
#### Inputs
_Fixture `tests/fixtures/basic-model.fga`:_
```text
model
  schema 1.1

type user

type group
  relations
    define owner: [user, user with inOfficeIP]

condition inOfficeIP(ip_addr: ipaddress) {
  ip_addr.in_cidr("192.168.0.0/24")
}
```
_Fixture `tests/fixtures/basic-tuples.json`:_
```text
[
    {
        "user": "user:anne",
        "relation": "owner",
        "object": "group:foo"
    }
]
```
_Fixture `tests/fixtures/basic-store.fga.yaml`:_
```text
name: Basic Store
model_file: basic-model.fga
tuple_file: basic-tuples.json
tests:
  - name: test-1
    check:
      - user: user:anne
        object: group:foo
        assertions:
          owner: true
```
#### When
```shell
fga model test --tests ./tests/fixtures/basic-store.fga.yaml
fga model test --tests tests/fixtures/basic-store.fga.yaml
fga model test --tests tests/**/*-store.fga.yaml
```
#### Then
- after `fga model test --tests ./tests/fixtures/basic-store.fga.yaml`:
  - exit code is `0`
  - stderr contains `# Test Summary #`, `Tests 1/1 passing`
- after `fga model test --tests tests/fixtures/basic-store.fga.yaml`:
  - exit code is `0`
  - stderr contains `# Test Summary #`
- after `fga model test --tests tests/**/*-store.fga.yaml`:
  - exit code is `0`
  - stderr contains `# Test Summary #`
### Scenario: a store file's references are contained to its directory
_only when `fga --version` succeeds_
#### Given
- Fixture file `tests/fixtures/basic-model.fga` is created.
- Fixture file `tests/fixtures/basic-tuples.json` is created.
- Fixture file `tests/fixtures/relative-path/relative-path-store.fga.yaml` is created.
- Fixture file `tests/fixtures/traversal/traversal-store.fga.yaml` is created.
#### Inputs
_Fixture `tests/fixtures/basic-model.fga`:_
```text
model
  schema 1.1

type user

type group
  relations
    define owner: [user]
```
_Fixture `tests/fixtures/basic-tuples.json`:_
```text
[
    {
        "user": "user:anne",
        "relation": "owner",
        "object": "group:foo"
    }
]
```
_Fixture `tests/fixtures/relative-path/relative-path-store.fga.yaml`:_
```text
name: Relative Path Store
model_file: ../basic-model.fga
tuple_file: ../basic-tuples.json
tests:
  - name: test-1
    check:
      - user: user:anne
        object: group:foo
        assertions:
          owner: true
```
_Fixture `tests/fixtures/traversal/traversal-store.fga.yaml`:_
```text
name: Traversal Store
model_file: ../../../../../../../../etc/hosts
tuples: []
tests: []
```
#### When
```shell
fga model test --tests tests/fixtures/relative-path/relative-path-store.fga.yaml --allow-external-files
fga model test --tests tests/fixtures/relative-path/relative-path-store.fga.yaml
fga model test --tests tests/fixtures/traversal/traversal-store.fga.yaml
```
#### Then
- after `fga model test --tests tests/fixtures/relative-path/relative-path-store.fga.yaml --allow-external-files`:
  - exit code is `0`
  - stderr contains `# Test Summary #`
- after `fga model test --tests tests/fixtures/relative-path/relative-path-store.fga.yaml`:
  - exit code is `1`
  - stderr contains `is not accessible within`
- after `fga model test --tests tests/fixtures/traversal/traversal-store.fga.yaml`:
  - exit code is `1`
  - stderr contains `is not accessible within`
### Scenario: the type-count limit fails loudly and a sufficient limit passes
_only when `fga --version` succeeds_
#### Given
- Fixture file `tests/fixtures/many-types-model.fga` is created.
- Fixture file `tests/fixtures/many-types.fga.yaml` is created.
#### Inputs
_Fixture `tests/fixtures/many-types-model.fga`:_
```text
model
  schema 1.1

type user

type resource1
  relations
    define owner: [user]

type resource2
  relations
    define owner: [user]

type resource3
  relations
    define owner: [user]

type resource4
  relations
    define owner: [user]
… (truncated, 4 more lines)
```
_Fixture `tests/fixtures/many-types.fga.yaml`:_
```text
name: Many Types Store
model_file: many-types-model.fga
tests:
  - name: test-owner-check
    tuples:
      - user: user:anne
        relation: owner
        object: resource1:doc1
    check:
      - user: user:anne
        object: resource1:doc1
        assertions:
          owner: true
```
#### When
```shell
fga model test --tests tests/fixtures/many-types.fga.yaml --max-types-per-authorization-model 5
fga model test --tests tests/fixtures/many-types.fga.yaml --max-types-per-authorization-model 10
```
#### Then
- after `fga model test --tests tests/fixtures/many-types.fga.yaml --max-types-per-authorization-model 5`:
  - exit code is `1`
  - stderr contains `exceeds the allowed limit of 5`
- after `fga model test --tests tests/fixtures/many-types.fga.yaml --max-types-per-authorization-model 10`:
  - exit code is `0`
  - stderr contains `# Test Summary #`
### Scenario: store create returns the new store as JSON
_only when `fga --version && openfga version` succeeds_
#### Given
- Background service `openfga` is started: `openfga run --http-addr 127.0.0.1:18230 --grpc-addr 127.0.0.1:18231 --metrics-enabled=false --log-level warn`.
#### When
```shell
fga store create --name "FGA Demo Store" --api-url http://127.0.0.1:18230
```
#### Then
- exit code is `0`
- stdout at `$.store.name` equals `FGA Demo Store`; at `$.store.id` matches `/^[A-Z0-9]+$/`
### Scenario: tuples write and delete against a live store
_only when `fga --version && openfga version` succeeds_
#### Given
- Background service `openfga` is started: `openfga run --http-addr 127.0.0.1:18232 --grpc-addr 127.0.0.1:18233 --metrics-enabled=false --log-level warn`.
- Fixture file `tests/fixtures/basic-model.fga` is created.
- Fixture file `tests/fixtures/basic-tuples.json` is created.
- Fixture file `tests/fixtures/basic-tuples.jsonl` is created.
#### Inputs
_Fixture `tests/fixtures/basic-model.fga`:_
```text
model
  schema 1.1

type user

type group
  relations
    define owner: [user, user with inOfficeIP]

condition inOfficeIP(ip_addr: ipaddress) {
  ip_addr.in_cidr("192.168.0.0/24")
}
```
_Fixture `tests/fixtures/basic-tuples.json`:_
```text
[
    {
        "user": "user:anne",
        "relation": "owner",
        "object": "group:foo"
    }
]
```
_Fixture `tests/fixtures/basic-tuples.jsonl`:_
```text
{"user": "user:bob", "relation": "owner", "object": "group:foo", "condition": {"name": "inOfficeIP", "context": {"ip_addr": "10.0.0.1"}}}
```
#### When
```shell
fga store create --name integration-test-store --api-url http://127.0.0.1:18232
# capture ${store_id} from stdout
fga model write --file=./tests/fixtures/basic-model.fga --store-id=${store_id} --api-url http://127.0.0.1:18232
# capture ${model_id} from stdout
fga tuple write --file=./tests/fixtures/basic-tuples.json --max-tuples-per-write=1 --max-parallel-requests=1 --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232
fga tuple write --file=./tests/fixtures/basic-tuples.jsonl --max-tuples-per-write=1 --max-parallel-requests=1 --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232
fga tuple delete --file=./tests/fixtures/basic-tuples.json --max-tuples-per-write=1 --max-parallel-requests=1 --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232
fga tuple write --file=./tests/fixtures/basic-tuples.json --hide-imported-tuples --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232
```
#### Then
- after `fga tuple write --file=./tests/fixtures/basic-tuples.json --max-tuples-per-write=1 --max-parallel-requests=1 --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232`:
  - exit code is `0`
  - stdout at `$.successful[0].user` equals `user:anne`; at `$.successful_count` equals `1`; at `$.failed_count` equals `0`
- after `fga tuple write --file=./tests/fixtures/basic-tuples.jsonl --max-tuples-per-write=1 --max-parallel-requests=1 --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232`:
  - exit code is `0`
  - stdout at `$.successful[0].user` equals `user:bob`; at `$.successful[0].condition.name` equals `inOfficeIP`
- after `fga tuple delete --file=./tests/fixtures/basic-tuples.json --max-tuples-per-write=1 --max-parallel-requests=1 --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232`:
  - exit code is `0`
  - stdout at `$.successful[0].user` equals `user:anne`
- after `fga tuple write --file=./tests/fixtures/basic-tuples.json --hide-imported-tuples --store-id=${store_id} --model-id=${model_id} --api-url http://127.0.0.1:18232`:
  - exit code is `0`
  - stdout at `$.total_count` equals `1`; at `$.successful_count` equals `1`; at `$.failed_count` equals `0`
### Scenario: store import builds a store, a model, and tuples from one file
_only when `fga --version && openfga version` succeeds_
#### Given
- Background service `openfga` is started: `openfga run --http-addr 127.0.0.1:18234 --grpc-addr 127.0.0.1:18235 --metrics-enabled=false --log-level warn`.
- Fixture file `tests/fixtures/basic-model.fga` is created.
- Fixture file `tests/fixtures/basic-tuples.json` is created.
- Fixture file `tests/fixtures/basic-store.fga.yaml` is created.
- Fixture file `tests/fixtures/relative-path/relative-path-store.fga.yaml` is created.
#### Inputs
_Fixture `tests/fixtures/basic-model.fga`:_
```text
model
  schema 1.1

type user

type group
  relations
    define owner: [user]
```
_Fixture `tests/fixtures/basic-tuples.json`:_
```text
[
    {
        "user": "user:anne",
        "relation": "owner",
        "object": "group:foo"
    }
]
```
_Fixture `tests/fixtures/basic-store.fga.yaml`:_
```text
name: Basic Store
model_file: basic-model.fga
tuple_file: basic-tuples.json
tests:
  - name: test-1
    check:
      - user: user:anne
        object: group:foo
        assertions:
          owner: true
```
_Fixture `tests/fixtures/relative-path/relative-path-store.fga.yaml`:_
```text
name: Relative Path Store
model_file: ../basic-model.fga
tuple_file: ../basic-tuples.json
tests:
  - name: test-1
    check:
      - user: user:anne
        object: group:foo
        assertions:
          owner: true
```
#### When
```shell
fga store import --file=./tests/fixtures/basic-store.fga.yaml --max-parallel-requests=1 --max-tuples-per-write=1 --api-url http://127.0.0.1:18234
fga store import --file=./tests/fixtures/relative-path/relative-path-store.fga.yaml --allow-external-files --api-url http://127.0.0.1:18234
fga store import --file=./tests/fixtures/relative-path/relative-path-store.fga.yaml --api-url http://127.0.0.1:18234
```
#### Then
- after `fga store import --file=./tests/fixtures/basic-store.fga.yaml --max-parallel-requests=1 --max-tuples-per-write=1 --api-url http://127.0.0.1:18234`:
  - exit code is `0`
  - stdout at `$.store.name` equals `Basic Store`; at `$.model.authorization_model_id` matches `/^[A-Z0-9]+$/`
- after `fga store import --file=./tests/fixtures/relative-path/relative-path-store.fga.yaml --allow-external-files --api-url http://127.0.0.1:18234`:
  - exit code is `0`
  - stdout at `$.store.name` equals `Relative Path Store`
- after `fga store import --file=./tests/fixtures/relative-path/relative-path-store.fga.yaml --api-url http://127.0.0.1:18234`:
  - exit code is `1`
  - stderr contains `is not accessible within`
