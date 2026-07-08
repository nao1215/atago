# atago Behavior Specs
## Summary
1 suite · 4 scenarios
## Contents
- [ntfy (self-hosted push notification service)](#ntfy-self-hosted-push-notification-service) — 4 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [a published notification comes back through the JSON poll feed](#scenario-a-published-notification-comes-back-through-the-json-poll-feed)
  - [the bundled CLI publishes against the same server](#scenario-the-bundled-cli-publishes-against-the-same-server)
  - [deny-all access control locks a topic until a user is granted](#scenario-deny-all-access-control-locks-a-topic-until-a-user-is-granted)
## ntfy (self-hosted push notification service)
Source: `test/e2e/thirdparty/ntfy/ntfy.atago.yaml`
### Scenario: the binary reports its version
#### When
```shell
ntfy --version
```
#### Then
- exit code is `0`
- stdout matches `/ntfy version [0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: a published notification comes back through the JSON poll feed
#### Given
- Background service `ntfy` is started: `ntfy serve --listen-http 127.0.0.1:18180 --base-url http://127.0.0.1:18180 --cache-file data/cache.db`.
- Fixture file `data/.keep` is created.
#### When
```shell
# HTTP GET /v1/health
# HTTP POST /atago-alerts
# HTTP GET /atago-alerts/json?poll=1
# HTTP GET /other-topic/json?poll=1
```
#### Then
- after `HTTP GET /v1/health`:
  - HTTP status is `200`
  - body at `$.healthy` equals `true`
- after `HTTP POST /atago-alerts`:
  - HTTP status is `200`
  - body at `$.event` equals `message`
- after `HTTP GET /atago-alerts/json?poll=1`:
  - HTTP status is `200`
  - body at `$.title` equals `Backup done`
  - body at `$.message` equals `nightly backup finished`
  - body at `$.priority` equals `4`
- after `HTTP GET /other-topic/json?poll=1`:
  - HTTP status is `200`
  - body is empty
### Scenario: the bundled CLI publishes against the same server
#### Given
- Background service `ntfy` is started: `ntfy serve --listen-http 127.0.0.1:18181 --base-url http://127.0.0.1:18181 --cache-file data/cache.db`.
- Fixture file `data/.keep` is created.
#### When
```shell
ntfy publish http://127.0.0.1:18181/atago-deploys "deploy one"
ntfy publish http://127.0.0.1:18181/atago-deploys "deploy two"
# HTTP GET /atago-deploys/json?poll=1&since=all
```
#### Then
- after `ntfy publish http://127.0.0.1:18181/atago-deploys "deploy one"`:
  - exit code is `0`
- after `ntfy publish http://127.0.0.1:18181/atago-deploys "deploy two"`:
  - exit code is `0`
- after `HTTP GET /atago-deploys/json?poll=1&since=all`:
  - HTTP status is `200`
  - body contains `deploy one`, `deploy two`
### Scenario: deny-all access control locks a topic until a user is granted
#### Given
- Background service `ntfy` is started: `ntfy serve --listen-http 127.0.0.1:18182 --base-url http://127.0.0.1:18182 --cache-file data/cache.db --auth-file data/auth.db --auth-default-access deny-all`.
- Fixture file `data/.keep` is created.
- Environment variables are set: NTFY_PASSWORD.
#### When
```shell
# HTTP POST /secure-topic
ntfy user add phil
ntfy access phil secure-topic rw
# HTTP POST /secure-topic
# HTTP GET /secure-topic/json?poll=1
```
#### Then
- after `HTTP POST /secure-topic`:
  - HTTP status is `403`
- after `ntfy user add phil`:
  - exit code is `0`
  - stdout contains `user phil added`
- after `ntfy access phil secure-topic rw`:
  - exit code is `0`
- after `HTTP POST /secure-topic`:
  - HTTP status is `200`
- after `HTTP GET /secure-topic/json?poll=1`:
  - HTTP status is `200`
  - body at `$.message` equals `authorized message`
