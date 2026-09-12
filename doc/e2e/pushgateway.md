# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [pushgateway (self-hosted metrics gateway)](#pushgateway-self-hosted-metrics-gateway) — 5 scenarios
  - [a pushed metric appears on the scrape endpoint](#scenario-a-pushed-metric-appears-on-the-scrape-endpoint)
  - [deleting a job group removes its metrics](#scenario-deleting-a-job-group-removes-its-metrics)
  - [a malformed exposition body is rejected and never ingested](#scenario-a-malformed-exposition-body-is-rejected-and-never-ingested)
  - [POST merges into a group while PUT replaces it](#scenario-post-merges-into-a-group-while-put-replaces-it)
  - [a grouping label decorates every metric in the group](#scenario-a-grouping-label-decorates-every-metric-in-the-group)

## pushgateway (self-hosted metrics gateway)
[Pushgateway](https://github.com/prometheus/pushgateway) exists so that a
job too short-lived to be scraped can push its metrics somewhere Prometheus
will find them later.

Its API is unusual in that the payload is not JSON but Prometheus's own
plain-text exposition format, sent as a raw request body. This suite pushes
real metrics that way and then reads them back from the gateway's own
endpoint — so what is asserted is that the gateway stored and re-exposed the
samples, in the format a Prometheus scrape would consume.

Source: `test/e2e/thirdparty/pushgateway/pushgateway.atago.yaml`
Network policy: egress is allowed only to `127.0.0.1`.
### Scenario: a pushed metric appears on the scrape endpoint
_only when `pushgateway --version` succeeds_
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18092`.

#### When
```shell
# HTTP POST /metrics/job/atago_e2e via push
# HTTP GET /metrics via push
```
#### Then
- after `HTTP POST /metrics/job/atago_e2e`:
  - HTTP status is `200`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body contains `job="atago_e2e"`
  - body matches `/atago_e2e_metric\{[^}]*job="atago_e2e"[^}]*\} 3.14/`

### Scenario: deleting a job group removes its metrics
_only when `pushgateway --version` succeeds_
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18093`.

#### When
```shell
# HTTP POST /metrics/job/ephemeral via push2
# HTTP DELETE /metrics/job/ephemeral via push2
# HTTP GET /metrics via push2
```
#### Then
- after `HTTP POST /metrics/job/ephemeral`:
  - HTTP status is `200`
- after `HTTP DELETE /metrics/job/ephemeral`:
  - HTTP status is `202`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body does not contain `ephemeral_metric`

### Scenario: a malformed exposition body is rejected and never ingested
_only when `pushgateway --version` succeeds_
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18099`.

#### When
```shell
# HTTP POST /metrics/job/badjob via push3
# HTTP GET /metrics via push3
```
#### Then
- after `HTTP POST /metrics/job/badjob`:
  - HTTP status is `400`
  - body contains `text format parsing error`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body does not contain `job="badjob"`

### Scenario: POST merges into a group while PUT replaces it
_only when `pushgateway --version` succeeds_
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18100`.

#### When
```shell
# HTTP POST /metrics/job/svc via push4
# HTTP POST /metrics/job/svc via push4
# HTTP GET /metrics via push4
# HTTP PUT /metrics/job/svc via push4
# HTTP GET /metrics via push4
```
#### Then
- after `HTTP POST /metrics/job/svc`:
  - HTTP status is `200`
- after `HTTP POST /metrics/job/svc`:
  - HTTP status is `200`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body contains `metric_a`, `metric_b`
- after `HTTP PUT /metrics/job/svc`:
  - HTTP status is `200`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body contains `metric_c`
  - body does not contain `metric_a`, `metric_b`

### Scenario: a grouping label decorates every metric in the group
_only when `pushgateway --version` succeeds_
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18101`.

#### When
```shell
# HTTP POST /metrics/job/svc/instance/host1 via push5
# HTTP GET /metrics via push5
```
#### Then
- after `HTTP POST /metrics/job/svc/instance/host1`:
  - HTTP status is `200`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body matches `/labeled_metric\{[^}]*instance="host1"[^}]*job="svc"[^}]*\} 9/`
