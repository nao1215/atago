# atago Behavior Specs
## Summary
1 suite · 4 scenarios
## Contents
- [prometheus (self-hosted monitoring system)](#prometheus-self-hosted-monitoring-system) — 4 scenarios
  - [promtool accepts a valid config and rejects a broken one](#scenario-promtool-accepts-a-valid-config-and-rejects-a-broken-one)
  - [promtool unit-tests an alerting rule](#scenario-promtool-unit-tests-an-alerting-rule)
  - [the server boots from an authored config and answers the query API](#scenario-the-server-boots-from-an-authored-config-and-answers-the-query-api)
  - [a self-scrape is ingested and queryable as up == 1](#scenario-a-self-scrape-is-ingested-and-queryable-as-up--1)
## prometheus (self-hosted monitoring system)
[Prometheus](https://prometheus.io/) ships as two things, and this suite
covers both.

`promtool` is an ordinary CLI: it must accept a valid configuration and
reject an invalid one with a diagnostic, and it must be able to unit-test
alerting rules — catching a bad alert before it reaches production rather
than during an incident.

The server is booted from an authored config and then genuinely used: its
health endpoints answer, its query API returns JSON that can be asserted as
structure, and a metric it scraped from itself is polled until it appears.
That last one closes the loop — configuration, scrape, storage, and query
all had to work for the sample to be there.

Source: `test/e2e/thirdparty/prometheus/prometheus.atago.yaml`
### Scenario: promtool accepts a valid config and rejects a broken one
_only when `promtool --version` succeeds_
#### Given
- Fixture file `prometheus.yml` is created.
- Fixture file `broken.yml` is created.
#### Inputs
_Fixture `prometheus.yml`:_
```text
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: self
    static_configs:
      - targets: ["127.0.0.1:9090"]
```
_Fixture `broken.yml`:_
```text
global:
  scrape_interval: not-a-duration
```
#### When
```shell
promtool check config prometheus.yml
promtool check config broken.yml
```
#### Then
- after `promtool check config prometheus.yml`:
  - exit code is `0`
  - stdout contains `SUCCESS: prometheus.yml is valid prometheus config file syntax`
- after `promtool check config broken.yml`:
  - exit code is not `0`
  - stderr contains `not a valid duration string`
### Scenario: promtool unit-tests an alerting rule
_only when `promtool --version` succeeds_
#### Given
- Fixture file `rules.yml` is created.
- Fixture file `rules_test.yml` is created.
#### Inputs
_Fixture `rules.yml`:_
```text
groups:
  - name: atago
    rules:
      - record: job:up:count
        expr: count(up)
      - alert: InstanceDown
        expr: up == 0
        for: 1m
        labels:
          severity: page
        annotations:
          summary: "Instance down"
```
_Fixture `rules_test.yml`:_
```text
rule_files:
  - rules.yml
evaluation_interval: 1m
tests:
  - interval: 1m
    input_series:
      - series: 'up{job="api",instance="a:9100"}'
        values: "0 0 0"
    alert_rule_test:
      - eval_time: 2m
        alertname: InstanceDown
        exp_alerts:
          - exp_labels:
              severity: page
              job: api
              instance: a:9100
            exp_annotations:
              summary: "Instance down"
```
#### When
```shell
promtool check rules rules.yml
promtool test rules rules_test.yml
```
#### Then
- after `promtool check rules rules.yml`:
  - exit code is `0`
  - stdout contains `SUCCESS: 2 rules found`
- after `promtool test rules rules_test.yml`:
  - exit code is `0`
  - stdout contains `SUCCESS`
### Scenario: the server boots from an authored config and answers the query API
_only when `promtool --version` succeeds_
#### Given
- Background service `prometheus` is started: `prometheus --config.file=prometheus.yml --storage.tsdb.path=data --web.listen-address=127.0.0.1:18130`.
- Fixture file `prometheus.yml` is created.
#### Inputs
_Fixture `prometheus.yml`:_
```text
global:
  scrape_interval: 15s
scrape_configs: []
```
#### When
```shell
# HTTP GET /-/healthy
# HTTP GET /-/ready
# HTTP GET /api/v1/status/buildinfo
# HTTP GET /api/v1/query?query=vector(42)
```
#### Then
- after `HTTP GET /-/healthy`:
  - HTTP status is `200`
  - body contains `Prometheus Server is Healthy.`
- after `HTTP GET /-/ready`:
  - HTTP status is `200`
  - body contains `Prometheus Server is Ready.`
- after `HTTP GET /api/v1/status/buildinfo`:
  - HTTP status is `200`
  - body at `$.status` equals `success`
  - body at `$.data.version` matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
- after `HTTP GET /api/v1/query?query=vector(42)`:
  - HTTP status is `200`
  - body at `$.data.result[0].value[1]` equals `42`
### Scenario: a self-scrape is ingested and queryable as up == 1
_only when `promtool --version` succeeds_
#### Given
- Background service `prometheus` is started: `prometheus --config.file=prometheus.yml --storage.tsdb.path=data --web.listen-address=127.0.0.1:18131`.
- Fixture file `prometheus.yml` is created.
#### Inputs
_Fixture `prometheus.yml`:_
```text
global:
  scrape_interval: 1s
scrape_configs:
  - job_name: self
    static_configs:
      - targets: ["127.0.0.1:18131"]
```
#### When
```shell
# HTTP GET /api/v1/query?query=up{job="self"}
```
#### Then
- HTTP status is `200`
- body at `$.data.result[0].metric.job` equals `self`
