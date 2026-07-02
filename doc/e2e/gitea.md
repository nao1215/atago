# atago Behavior Specs
## Summary
1 suite · 3 scenarios
## Contents
- [gitea (self-hosted git service)](#gitea-self-hosted-git-service) — 3 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [the server boots from an authored app.ini and reports healthy](#scenario-the-server-boots-from-an-authored-appini-and-reports-healthy)
  - [admin CLI, REST API, and a real git clone interoperate](#scenario-admin-cli-rest-api-and-a-real-git-clone-interoperate)
## gitea (self-hosted git service)
Source: `test/e2e/thirdparty/gitea/gitea.atago.yaml`
### Scenario: the binary reports its version
#### When
```shell
gitea --version
```
#### Then
- exit code is `0`
- stdout matches `/gitea version [0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: the server boots from an authored app.ini and reports healthy
#### Given
- Background service `gitea` is started: `gitea web --config app.ini`.
- Fixture file `app.ini` is created.
#### Inputs
_Fixture `app.ini`:_
```
[server]
HTTP_ADDR = 127.0.0.1
HTTP_PORT = 18140
ROOT_URL = http://127.0.0.1:18140/
DISABLE_SSH = true
OFFLINE_MODE = true

[database]
DB_TYPE = sqlite3
PATH = data/gitea.db

[security]
INSTALL_LOCK = true

[log]
MODE = console
LEVEL = Warn

[repository]
ROOT = data/repos
… (truncated, 3 more lines)
```
#### When
```shell
# HTTP GET /api/healthz
# HTTP GET /api/v1/version
```
#### Then
- after `HTTP GET /api/healthz`:
  - HTTP status is `200`
- after `HTTP GET /api/v1/version`:
  - HTTP status is `200`
  - body at `$.version` matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: admin CLI, REST API, and a real git clone interoperate
#### Given
- Background service `gitea` is started: `gitea web --config app.ini`.
- Fixture file `app.ini` is created.
#### Inputs
_Fixture `app.ini`:_
```
[server]
HTTP_ADDR = 127.0.0.1
HTTP_PORT = 18141
ROOT_URL = http://127.0.0.1:18141/
DISABLE_SSH = true
OFFLINE_MODE = true

[database]
DB_TYPE = sqlite3
PATH = data/gitea.db

[security]
INSTALL_LOCK = true

[log]
MODE = console
LEVEL = Warn

[repository]
ROOT = data/repos
… (truncated, 3 more lines)
```
#### When
```shell
# HTTP GET /api/healthz
gitea --config app.ini admin user create --username atago --password atago-e2e-pass1 --email atago@example.com --admin --must-change-password=false
gitea --config app.ini admin user generate-access-token --username atago --token-name e2e --scopes all --raw
# capture ${token} from stdout
# HTTP POST /api/v1/user/repos
# HTTP POST /api/v1/repos/atago/demo/contents/hello.txt
# HTTP POST /api/v1/repos/atago/demo/issues
# HTTP GET /api/v1/repos/atago/demo/issues/1
git clone http://atago:atago-e2e-pass1@127.0.0.1:18141/atago/demo.git checkout
```
#### Then
- after `gitea --config app.ini admin user create --username atago --password atago-e2e-pass1 --email atago@example.com --admin --must-change-password=false`:
  - exit code is `0`
  - stdout contains `successfully created`
- after `gitea --config app.ini admin user generate-access-token --username atago --token-name e2e --scopes all --raw`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]{40}\s*$/`
- after `HTTP POST /api/v1/user/repos`:
  - HTTP status is `201`
  - body at `$.name` equals `demo`
  - body at `$.owner.login` equals `atago`
- after `HTTP POST /api/v1/repos/atago/demo/contents/hello.txt`:
  - HTTP status is `201`
  - body at `$.content.path` equals `hello.txt`
- after `HTTP POST /api/v1/repos/atago/demo/issues`:
  - HTTP status is `201`
  - body at `$.number` equals `1`
- after `HTTP GET /api/v1/repos/atago/demo/issues/1`:
  - HTTP status is `200`
  - body at `$.state` equals `open`
- after `git clone http://atago:atago-e2e-pass1@127.0.0.1:18141/atago/demo.git checkout`:
  - exit code is `0`
  - file `checkout/hello.txt` contains `hello from atago`
  - dir `checkout` contains `README.md`, contains `hello.txt`