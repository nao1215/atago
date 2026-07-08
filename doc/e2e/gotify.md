# atago Behavior Specs
## Summary
1 suite · 3 scenarios
## Contents
- [gotify (self-hosted notification server)](#gotify-self-hosted-notification-server) — 3 scenarios
  - [the server reports health and version](#scenario-the-server-reports-health-and-version)
  - [an application pushes and the admin reads it back](#scenario-an-application-pushes-and-the-admin-reads-it-back)
  - [the app icon round-trips through a multipart upload](#scenario-the-app-icon-round-trips-through-a-multipart-upload)
## gotify (self-hosted notification server)
Source: `test/e2e/thirdparty/gotify/gotify.atago.yaml`
### Scenario: the server reports health and version
#### Given
- Background service `gotify` is started: `gotify`.
- Fixture file `data/.keep` is created.
#### When
```shell
# HTTP GET /health
# HTTP GET /version
```
#### Then
- after `HTTP GET /health`:
  - HTTP status is `200`
  - body at `$.health` equals `green`
- after `HTTP GET /version`:
  - HTTP status is `200`
  - body at `$.version` matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: an application pushes and the admin reads it back
#### Given
- Background service `gotify` is started: `gotify`.
- Fixture file `data/.keep` is created.
#### When
```shell
# HTTP GET /message
# HTTP POST /application
# capture ${app_token} from the response body
# HTTP POST /message
# HTTP GET /message
```
#### Then
- after `HTTP GET /message`:
  - HTTP status is `401`
- after `HTTP POST /application`:
  - HTTP status is `200`
  - body at `$.name` equals `atago-pipeline`
- after `HTTP POST /message`:
  - HTTP status is `200`
  - body at `$.title` equals `Deploy done`
- after `HTTP GET /message`:
  - HTTP status is `200`
  - body at `$.messages` has length 1
  - body at `$.messages[0].priority` equals `5`
### Scenario: the app icon round-trips through a multipart upload
#### Given
- Background service `gotify` is started: `gotify`.
- Fixture file `data/.keep` is created.
- Fixture file `icon.png` is created.
#### When
```shell
# HTTP POST /application
# capture ${app_id} from the response body
# HTTP POST /application/${app_id}/image
# capture ${image_path} from the response body
# HTTP GET /${image_path}
```
#### Then
- after `HTTP POST /application`:
  - HTTP status is `200`
- after `HTTP POST /application/${app_id}/image`:
  - HTTP status is `200`
  - body at `$.image` matches `/^image//`
- after `HTTP GET /${image_path}`:
  - HTTP status is `200`
  - image `fetched-icon.png` is `png`, width 1, height 1
#### Generated artifacts
- `fetched-icon.png`
