# atago Behavior Specs
## Summary
1 suite · 3 scenarios
## Contents
- [grafana (self-hosted observability platform)](#grafana-self-hosted-observability-platform) — 3 scenarios
  - [the server boots and reports health and build info](#scenario-the-server-boots-and-reports-health-and-build-info)
  - [anonymous visitors are redirected to the login page](#scenario-anonymous-visitors-are-redirected-to-the-login-page)
  - [a dashboard and datasource lifecycle over the REST API](#scenario-a-dashboard-and-datasource-lifecycle-over-the-rest-api)
## grafana (self-hosted observability platform)
Source: `test/e2e/thirdparty/grafana/grafana.atago.yaml`
### Scenario: the server boots and reports health and build info
#### Given
- Background service `grafana` is started: `grafana server --homepath "$GF_PATHS_HOME"`.
#### When
```shell
# HTTP GET /api/health
```
#### Then
- HTTP status is `200`
- body at `$.database` equals `ok`
- body at `$.version` matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: anonymous visitors are redirected to the login page
#### Given
- Background service `grafana` is started: `grafana server --homepath "$GF_PATHS_HOME"`.
#### When
```shell
# HTTP GET /api/health
# HTTP GET /
# HTTP GET /api/search
```
#### Then
- after `HTTP GET /`:
  - HTTP status is `302`
  - header `Location` contains `/login`
- after `HTTP GET /api/search`:
  - HTTP status is `401`
### Scenario: a dashboard and datasource lifecycle over the REST API
#### Given
- Background service `grafana` is started: `grafana server --homepath "$GF_PATHS_HOME"`.
#### When
```shell
# HTTP GET /api/health
# HTTP GET /api/org
# HTTP POST /api/dashboards/db
# capture ${dash_uid} from the response body
# HTTP GET /api/dashboards/uid/${dash_uid}
# HTTP GET /api/search?query=atago
# HTTP POST /api/datasources
```
#### Then
- after `HTTP GET /api/org`:
  - HTTP status is `200`
  - body at `$.name` equals `Main Org.`
- after `HTTP POST /api/dashboards/db`:
  - HTTP status is `200`
  - body at `$.status` equals `success`
- after `HTTP GET /api/dashboards/uid/${dash_uid}`:
  - HTTP status is `200`
  - body at `$.dashboard.title` equals `atago e2e dashboard`
- after `HTTP GET /api/search?query=atago`:
  - HTTP status is `200`
  - body at `$` has length 1
- after `HTTP POST /api/datasources`:
  - HTTP status is `200`
  - body at `$.message` equals `Datasource added`