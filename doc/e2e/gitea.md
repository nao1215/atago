# atago Behavior Specs
## Summary
1 suite · 3 scenarios
## Contents
- [gitea (self-hosted git service)](#gitea-self-hosted-git-service) — 3 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [the server boots from an authored app.ini and reports healthy](#scenario-the-server-boots-from-an-authored-appini-and-reports-healthy)
  - [admin CLI, REST API, and a real git clone interoperate](#scenario-admin-cli-rest-api-and-a-real-git-clone-interoperate)
## gitea (self-hosted git service)
[Gitea](https://about.gitea.com/) is a whole git forge, and this suite meets
it the way an administrator would: a real server is booted from an authored
`app.ini` on SQLite, a user and an access token are provisioned through
Gitea's own CLI, and everything after that goes over the REST API —
creating a repository, committing files into it, opening issues.

The suite ends by cloning that repository with real `git` over HTTP. That
last step is the point: two independent third-party programs, neither aware
of this test, are shown to interoperate — the forge really serves git, not
just a JSON API that claims a repository exists.

Source: `test/e2e/thirdparty/gitea/gitea.atago.yaml`
Network policy: egress is allowed only to `127.0.0.1`.
### Scenario: the binary reports its version
_only when `command -v gitea` succeeds_
#### When
```shell
gitea --version
```
#### Then
- exit code is `0`
- stdout matches `/gitea version [0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: the server boots from an authored app.ini and reports healthy
_only when `command -v gitea` succeeds_
#### Given
- Background service `gitea` is started: `gitea web --config app.ini`.
- Fixture file `app.ini` is created.
- The step is retried up to 30 times every 1s until HTTP status is `200`.
#### Inputs
_Fixture `app.ini`:_
```text
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
# HTTP GET /api/healthz via gitea
# HTTP GET /api/v1/version via gitea
```
#### Then
- after `HTTP GET /api/healthz`:
  - HTTP status is `200`
- after `HTTP GET /api/v1/version`:
  - HTTP status is `200`
  - body at `$.version` matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: admin CLI, REST API, and a real git clone interoperate
_only when `command -v gitea` succeeds_
#### Given
- Background service `gitea` is started: `gitea web --config app.ini`.
- Fixture file `app.ini` is created.
- The step is retried up to 30 times every 1s until HTTP status is `200`.
#### Inputs
_Fixture `app.ini`:_
```text
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
# HTTP GET /api/healthz via gitea2
gitea --config app.ini admin user create --username atago --password atago-e2e-pass1 --email atago@example.com --admin --must-change-password=false
gitea --config app.ini admin user generate-access-token --username atago --token-name e2e --scopes all --raw
# capture ${token} from stdout
# HTTP POST /api/v1/user/repos via gitea2
# HTTP POST /api/v1/repos/atago/demo/contents/hello.txt via gitea2
# HTTP POST /api/v1/repos/atago/demo/issues via gitea2
# HTTP GET /api/v1/repos/atago/demo/issues/1 via gitea2
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
