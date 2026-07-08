# atago Behavior Specs
## Summary
1 suite · 6 scenarios
## Contents
- [caddy (self-hosted web server)](#caddy-self-hosted-web-server) — 6 scenarios
  - [the standard module set is compiled in](#scenario-the-standard-module-set-is-compiled-in)
  - [adapt turns a Caddyfile into canonical JSON](#scenario-adapt-turns-a-caddyfile-into-canonical-json)
  - [fmt normalizes a messy Caddyfile in place](#scenario-fmt-normalizes-a-messy-caddyfile-in-place)
  - [validate rejects a broken Caddyfile](#scenario-validate-rejects-a-broken-caddyfile)
  - [a config-driven server boots from an authored Caddyfile](#scenario-a-config-driven-server-boots-from-an-authored-caddyfile)
  - [the file server serves fixtures over real HTTP](#scenario-the-file-server-serves-fixtures-over-real-http)
## caddy (self-hosted web server)
Source: `test/e2e/thirdparty/caddy/caddy.atago.yaml`
### Scenario: the standard module set is compiled in
#### When
```shell
caddy list-modules
```
#### Then
- exit code is `0`
- stdout contains `http.handlers.file_server`, `http.handlers.static_response`
### Scenario: adapt turns a Caddyfile into canonical JSON
#### Given
- Fixture file `Caddyfile` is created.
#### Inputs
_Fixture `Caddyfile`:_
```text
:18080
respond "hello from caddy"
```
#### When
```shell
caddy adapt --config Caddyfile
```
#### Then
- exit code is `0`
- stdout at `$.apps.http.servers.srv0.listen[0]` equals `:18080`
### Scenario: fmt normalizes a messy Caddyfile in place
#### Given
- Fixture file `Caddyfile` is created.
#### Inputs
_Fixture `Caddyfile`:_
```text
:18080   {
        respond    "ok"
  }
```
#### When
```shell
caddy fmt --overwrite Caddyfile
```
#### Then
- exit code is `0`
- file `Caddyfile` contains `respond "ok"`
### Scenario: validate rejects a broken Caddyfile
#### Given
- Fixture file `Caddyfile` is created.
#### Inputs
_Fixture `Caddyfile`:_
```text
:18080
no_such_directive_xyz
```
#### When
```shell
caddy validate --config Caddyfile --adapter caddyfile
```
#### Then
- exit code is not `0`
- stderr contains `no_such_directive_xyz`
### Scenario: a config-driven server boots from an authored Caddyfile
#### Given
- Background service `caddy` is started: `caddy run --config Caddyfile --adapter caddyfile`.
- Fixture file `Caddyfile` is created.
#### Inputs
_Fixture `Caddyfile`:_
```text
{
	admin off
}
http://127.0.0.1:18081

respond "configured response" 200
```
#### When
```shell
# HTTP GET /
```
#### Then
- HTTP status is `200`
- body contains `configured response`
### Scenario: the file server serves fixtures over real HTTP
#### Given
- Background service `caddy` is started: `caddy file-server --listen 127.0.0.1:18080 --root site`.
- Fixture file `site/index.html` is created.
- Fixture file `site/api/status.json` is created.
#### Inputs
_Fixture `site/index.html`:_
```text
<html><body>hello from caddy</body></html>
```
_Fixture `site/api/status.json`:_
```text
{"status":"ok","service":"caddy"}
```
#### When
```shell
# HTTP GET /index.html
# HTTP GET /api/status.json
# HTTP GET /no-such-file
```
#### Then
- after `HTTP GET /index.html`:
  - HTTP status is `200`
  - body contains `hello from caddy`
  - header `Content-Type` contains `text/html`
- after `HTTP GET /api/status.json`:
  - HTTP status is `200`
  - body at `$.status` equals `ok`
- after `HTTP GET /no-such-file`:
  - HTTP status is `404`
