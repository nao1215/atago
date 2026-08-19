# atago Behavior Specs
## Summary
1 suite · 13 scenarios
## Contents
- [curl (what actually goes on the wire)](#curl-what-actually-goes-on-the-wire) — 13 scenarios
  - [a plain GET is a GET, with the headers curl adds on its own](#scenario-a-plain-get-is-a-get-with-the-headers-curl-adds-on-its-own)
  - [-d makes it a POST and names the form encoding](#scenario--d-makes-it-a-post-and-names-the-form-encoding)
  - [-d strips the newlines that --data-binary keeps](#scenario--d-strips-the-newlines-that---data-binary-keeps)
  - [--json sets both content types and sends the body untouched](#scenario---json-sets-both-content-types-and-sends-the-body-untouched)
  - [-F builds a multipart body with a boundary in the header](#scenario--f-builds-a-multipart-body-with-a-boundary-in-the-header)
  - [-H replaces a header curl would have sent, and can delete one](#scenario--h-replaces-a-header-curl-would-have-sent-and-can-delete-one)
  - [-u builds the Authorization header](#scenario--u-builds-the-authorization-header)
  - [-X renames the method and changes nothing else](#scenario--x-renames-the-method-and-changes-nothing-else)
  - [-G moves the data into the query string and leaves no body](#scenario--g-moves-the-data-into-the-query-string-and-leaves-no-body)
  - [-L follows the redirect, and the server sees both requests](#scenario--l-follows-the-redirect-and-the-server-sees-both-requests)
  - [a path the server does not serve is a 404 that curl reports only when asked](#scenario-a-path-the-server-does-not-serve-is-a-404-that-curl-reports-only-when-asked)
  - [a slow route is a timeout with its own exit code](#scenario-a-slow-route-is-a-timeout-with-its-own-exit-code)
  - [the same request twice produces the same bytes](#scenario-the-same-request-twice-produces-the-same-bytes)

## curl (what actually goes on the wire)
Every flag of [curl](https://curl.se/) that shapes a request is a promise
about bytes a server will receive, and the one place that promise can be
checked is at the server. These scenarios declare an atago mock server, run
curl against it, and then assert the recorded request — its method, its
headers, its body, and how many of them there were.

Nothing here asserts what curl printed, because that is not the contract:
`-d` versus `--data-binary` differ in what is sent, not in what is shown,
and `-X GET` with a body is a request no reasonable output would warn you
about. The mock is the only witness.

No network is reached: the server is atago's own, on a loopback port that
only lives as long as the scenario.

Source: `test/e2e/thirdparty/curl/requests.atago.yaml`
### Scenario: a plain GET is a GET, with the headers curl adds on its own
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s ${api.url}/hello
```
#### Then
- exit code is `0`
- stdout equals an exact value
- mock `api` received `GET /hello` exactly 1 time(s)
- mock `api` received `/hello`
- mock `api` received `/hello`

#### Expected output
_expected stdout:_
```text
hi
```
### Scenario: -d makes it a POST and names the form encoding
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s -d name=ada -d role=engineer ${api.url}/submit
```
#### Then
- exit code is `0`
- mock `api` received `POST /submit` exactly 1 time(s)
- mock `api` received `/submit`

### Scenario: -d strips the newlines that --data-binary keeps
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Fixture file `payload.txt` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `payload.txt`:_
```text
first line
second line
```
#### When
```shell
curl -s -d @payload.txt ${api.url}/upload
curl -s --data-binary @payload.txt ${api.url}/upload
```
#### Then
- after `curl -s -d @payload.txt ${api.url}/upload`:
  - exit code is `0`
  - mock `api` received `POST /upload`
- after `curl -s --data-binary @payload.txt ${api.url}/upload`:
  - exit code is `0`
  - mock `api` received `POST /upload`

### Scenario: --json sets both content types and sends the body untouched
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Fixture file `item.json` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `item.json`:_
```text
{"name": "ada", "tags": ["a", "b"]}
```
#### When
```shell
curl -s --json @item.json ${api.url}/v1/items
```
#### Then
- exit code is `0`
- mock `api` received `POST /v1/items` exactly 1 time(s)
- mock `api` received `/v1/items`
- mock `api` received `/v1/items`

### Scenario: -F builds a multipart body with a boundary in the header
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Fixture file `note.txt` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `note.txt`:_
```text
the file content
```
#### When
```shell
curl -s -F title=a-note -F file=@note.txt ${api.url}/files
```
#### Then
- exit code is `0`
- mock `api` received `POST /files`
- mock `api` received `/files`

### Scenario: -H replaces a header curl would have sent, and can delete one
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s -H "User-Agent: atago-suite/1.0" ${api.url}/hello
curl -s -H "Accept:" ${api.url}/hello
```
#### Then
- after `curl -s -H "User-Agent: atago-suite/1.0" ${api.url}/hello`:
  - exit code is `0`
  - mock `api` received `/hello`
- after `curl -s -H "Accept:" ${api.url}/hello`:
  - exit code is `0`
  - mock `api` received `/hello`

### Scenario: -u builds the Authorization header
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Fixture file `credential.txt` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `credential.txt`:_
```text
ada:test-password
```
#### When
```shell
curl -s -u "$(cat credential.txt)" ${api.url}/private
```
#### Then
- exit code is `0`
- mock `api` received `GET /private` exactly 1 time(s)

### Scenario: -X renames the method and changes nothing else
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s -X GET -d name=ada ${api.url}/odd
```
#### Then
- exit code is `0`
- mock `api` received `GET /odd` exactly 1 time(s)
- mock `api` received `/odd`

### Scenario: -G moves the data into the query string and leaves no body
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s -G -d q=ada -d limit=10 ${api.url}/search
```
#### Then
- exit code is `0`
- mock `api` received `GET /search` exactly 1 time(s)

### Scenario: -L follows the redirect, and the server sees both requests
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 2 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s -d name=ada ${api.url}/old
curl -s -L -d name=ada ${api.url}/old
```
#### Then
- after `curl -s -d name=ada ${api.url}/old`:
  - exit code is `0`
  - stdout is empty
  - mock `api` received `/old` exactly 1 time(s)
- after `curl -s -L -d name=ada ${api.url}/old`:
  - exit code is `0`
  - stdout equals an exact value
  - mock `api` received `GET /new` exactly 1 time(s)

#### Expected output
_expected stdout:_
```text
moved here
```
### Scenario: a path the server does not serve is a 404 that curl reports only when asked
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s -o /dev/null -w "%{http_code}" ${api.url}/nowhere
curl -s -f -o /dev/null ${api.url}/nowhere
```
#### Then
- after `curl -s -o /dev/null -w "%{http_code}" ${api.url}/nowhere`:
  - exit code is `0`
  - stdout equals an exact value
- after `curl -s -f -o /dev/null ${api.url}/nowhere`:
  - exit code is `22`
  - mock `api` received `/nowhere` exactly 2 time(s)

### Scenario: a slow route is a timeout with its own exit code
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s --max-time 1 ${api.url}/slow
```
#### Then
- exit code is `28`
- stdout is empty

### Scenario: the same request twice produces the same bytes
_only when `curl --version` succeeds · skipped on Windows_
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
curl -s ${api.url}/v1/item
```
#### Then
- exit code is `0`
- stdout at `$.name` equals `ada`
- mock `api` received `/v1/item` exactly 3 time(s)
