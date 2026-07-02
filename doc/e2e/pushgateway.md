# atago Behavior Specs
## Summary
1 suite · 2 scenarios
## Contents
- [pushgateway (self-hosted metrics gateway)](#pushgateway-self-hosted-metrics-gateway) — 2 scenarios
  - [a pushed metric appears on the scrape endpoint](#scenario-a-pushed-metric-appears-on-the-scrape-endpoint)
  - [deleting a job group removes its metrics](#scenario-deleting-a-job-group-removes-its-metrics)
## pushgateway (self-hosted metrics gateway)
Source: `test/e2e/thirdparty/pushgateway/pushgateway.atago.yaml`
### Scenario: a pushed metric appears on the scrape endpoint
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18092`.
#### When
```shell
# HTTP POST /metrics/job/atago_e2e
# HTTP GET /metrics
```
#### Then
- after `HTTP POST /metrics/job/atago_e2e`:
  - HTTP status is `200`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body contains `job="atago_e2e"`
  - body matches `/atago_e2e_metric\{[^}]*job="atago_e2e"[^}]*\} 3.14/`
### Scenario: deleting a job group removes its metrics
#### Given
- Background service `pushgateway` is started: `pushgateway --web.listen-address=127.0.0.1:18093`.
#### When
```shell
# HTTP POST /metrics/job/ephemeral
# HTTP DELETE /metrics/job/ephemeral
# HTTP GET /metrics
```
#### Then
- after `HTTP POST /metrics/job/ephemeral`:
  - HTTP status is `200`
- after `HTTP DELETE /metrics/job/ephemeral`:
  - HTTP status is `202`
- after `HTTP GET /metrics`:
  - HTTP status is `200`
  - body does not contain `ephemeral_metric`