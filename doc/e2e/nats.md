# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [nats (self-hosted messaging system)](#nats-self-hosted-messaging-system) — 5 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [the monitoring endpoint reports a healthy JetStream server](#scenario-the-monitoring-endpoint-reports-a-healthy-jetstream-server)
  - [request/reply round-trips through the broker](#scenario-requestreply-round-trips-through-the-broker)
  - [a JetStream stream persists, counts, and purges messages](#scenario-a-jetstream-stream-persists-counts-and-purges-messages)
  - [the JetStream KV bucket stores and serves configuration](#scenario-the-jetstream-kv-bucket-stores-and-serves-configuration)
## nats (self-hosted messaging system)
[NATS](https://nats.io/) is a message broker, which makes it the one
workload here that needs two live participants to mean anything.

Request/reply is tested with a real responder: a second background process
subscribes and answers, so a request is proven to have reached something and
come back — not merely to have been accepted by the broker.

JetStream, the persistence layer, is followed through its whole cycle:
create a stream, publish into it, read the stream's state back and count
what is stored, then purge it and confirm the count fell. The key-value
store built on top is exercised the same way, and the broker's monitoring
endpoint is checked over HTTP alongside.

Source: `test/e2e/thirdparty/nats/nats.atago.yaml`
Network policy: egress is allowed only to `127.0.0.1`.
### Scenario: the binary reports its version
_only when `nats-server --version` succeeds_
#### When
```shell
nats-server --version
```
#### Then
- exit code is `0`
- stdout matches `/nats-server: v[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: the monitoring endpoint reports a healthy JetStream server
_only when `nats-server --version` succeeds_
#### Given
- Background service `nats-server` is started: `nats-server -js -sd store -a 127.0.0.1 -p 18160 -m 18161`.
#### When
```shell
# HTTP GET /healthz via monitor
# HTTP GET /varz via monitor
```
#### Then
- after `HTTP GET /healthz`:
  - HTTP status is `200`
  - body at `$.status` equals `ok`
- after `HTTP GET /varz`:
  - HTTP status is `200`
  - body at `$.version` matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
  - body at `$.port` equals `18160`
### Scenario: request/reply round-trips through the broker
_only when `nats-server --version` succeeds_
#### Given
- Background service `nats-server` is started: `nats-server -a 127.0.0.1 -p 18162`.
- Background service `responder` is started: `nats -s nats://127.0.0.1:18162 reply help.please "OK I CAN HELP"`.
#### When
```shell
nats -s nats://127.0.0.1:18162 request help.please "help me"
```
#### Then
- exit code is `0`
- stdout contains `OK I CAN HELP`
### Scenario: a JetStream stream persists, counts, and purges messages
_only when `nats-server --version` succeeds_
#### Given
- Background service `nats-server` is started: `nats-server -js -sd store -a 127.0.0.1 -p 18163`.
#### When
```shell
nats -s nats://127.0.0.1:18163 stream add ORDERS --subjects "orders.>" --defaults
nats -s nats://127.0.0.1:18163 pub orders.new "order-1"
nats -s nats://127.0.0.1:18163 pub orders.priority "order-2"
nats -s nats://127.0.0.1:18163 stream info ORDERS --json
nats -s nats://127.0.0.1:18163 stream purge ORDERS --force
nats -s nats://127.0.0.1:18163 stream info ORDERS --json
```
#### Then
- after `nats -s nats://127.0.0.1:18163 stream add ORDERS --subjects "orders.>" --defaults`:
  - exit code is `0`
  - stdout contains `Stream ORDERS was created`
- after `nats -s nats://127.0.0.1:18163 pub orders.new "order-1"`:
  - exit code is `0`
- after `nats -s nats://127.0.0.1:18163 pub orders.priority "order-2"`:
  - exit code is `0`
- after `nats -s nats://127.0.0.1:18163 stream info ORDERS --json`:
  - exit code is `0`
  - stdout at `$.state.messages` equals `2`
  - stdout at `$.config.subjects[0]` equals `orders.>`
- after `nats -s nats://127.0.0.1:18163 stream purge ORDERS --force`:
  - exit code is `0`
- after `nats -s nats://127.0.0.1:18163 stream info ORDERS --json`:
  - exit code is `0`
  - stdout at `$.state.messages` equals `0`
### Scenario: the JetStream KV bucket stores and serves configuration
_only when `nats-server --version` succeeds_
#### Given
- Background service `nats-server` is started: `nats-server -js -sd store -a 127.0.0.1 -p 18164`.
#### When
```shell
nats -s nats://127.0.0.1:18164 kv add CONFIG
nats -s nats://127.0.0.1:18164 kv put CONFIG greeting "hello from atago"
nats -s nats://127.0.0.1:18164 kv get CONFIG greeting --raw
nats -s nats://127.0.0.1:18164 kv del CONFIG greeting --force
nats -s nats://127.0.0.1:18164 kv get CONFIG greeting --raw
```
#### Then
- after `nats -s nats://127.0.0.1:18164 kv add CONFIG`:
  - exit code is `0`
  - stdout contains `Bucket Name: CONFIG`
- after `nats -s nats://127.0.0.1:18164 kv put CONFIG greeting "hello from atago"`:
  - exit code is `0`
- after `nats -s nats://127.0.0.1:18164 kv get CONFIG greeting --raw`:
  - exit code is `0`
  - stdout equals an exact value
- after `nats -s nats://127.0.0.1:18164 kv del CONFIG greeting --force`:
  - exit code is `0`
- after `nats -s nats://127.0.0.1:18164 kv get CONFIG greeting --raw`:
  - exit code is not `0`
