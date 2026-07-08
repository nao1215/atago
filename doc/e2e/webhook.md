# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [webhook (self-hosted webhook receiver)](#webhook-self-hosted-webhook-receiver) — 5 scenarios
  - [a post runs the command, returns its output, and writes its file](#scenario-a-post-runs-the-command-returns-its-output-and-writes-its-file)
  - [a trigger-rule gates execution and blocks it when unsatisfied](#scenario-a-trigger-rule-gates-execution-and-blocks-it-when-unsatisfied)
  - [an HMAC signature is verified against an independent oracle](#scenario-an-hmac-signature-is-verified-against-an-independent-oracle)
  - [the http-methods allowlist rejects the wrong verb](#scenario-the-http-methods-allowlist-rejects-the-wrong-verb)
  - [a command that exits non-zero surfaces as a 500](#scenario-a-command-that-exits-non-zero-surfaces-as-a-500)
## webhook (self-hosted webhook receiver)
Source: `test/e2e/thirdparty/webhook/webhook.atago.yaml`
### Scenario: a post runs the command, returns its output, and writes its file
#### Given
- Background service `webhook` is started: `webhook -hooks hooks.json -ip 127.0.0.1 -port 18094`.
- Fixture file `handler.sh` is created.
- Fixture file `hooks.json` is created.
#### Inputs
_Fixture `handler.sh`:_
```text
#!/bin/sh
printf 'handled %s\n' "$1" > out.txt
printf 'ran for %s' "$1"
```
_Fixture `hooks.json`:_
```text
[
  {
    "id": "greet",
    "execute-command": "${workdir}/handler.sh",
    "command-working-directory": "${workdir}",
    "include-command-output-in-response": true,
    "pass-arguments-to-command": [
      { "source": "payload", "name": "name" }
    ]
  }
]
```
#### When
```shell
# HTTP POST /hooks/greet
# HTTP POST /hooks/nope
```
#### Then
- after `HTTP POST /hooks/greet`:
  - HTTP status is `200`
  - body contains `ran for Alice`
  - file `out.txt` contains `handled Alice`
- after `HTTP POST /hooks/nope`:
  - HTTP status is `404`
### Scenario: a trigger-rule gates execution and blocks it when unsatisfied
#### Given
- Background service `webhook` is started: `webhook -hooks hooks.json -ip 127.0.0.1 -port 18095`.
- Fixture file `handler.sh` is created.
- Fixture file `hooks.json` is created.
#### Inputs
_Fixture `handler.sh`:_
```text
#!/bin/sh
printf 'executed\n' > ran.txt
printf 'ok'
```
_Fixture `hooks.json`:_
```text
[
  {
    "id": "guarded",
    "execute-command": "${workdir}/handler.sh",
    "command-working-directory": "${workdir}",
    "include-command-output-in-response": true,
    "trigger-rule": {
      "match": {
        "type": "value",
        "value": "open-sesame",
        "parameter": { "source": "payload", "name": "token" }
      }
    }
  }
]
```
#### When
```shell
# HTTP POST /hooks/guarded
# HTTP POST /hooks/guarded
```
#### Then
- after `HTTP POST /hooks/guarded`:
  - HTTP status is `200`
  - body contains `Hook rules were not satisfied`
  - file `ran.txt` does not exist
- after `HTTP POST /hooks/guarded`:
  - HTTP status is `200`
  - file `ran.txt` contains `executed`
### Scenario: an HMAC signature is verified against an independent oracle
#### Given
- Background service `webhook` is started: `webhook -hooks hooks.json -ip 127.0.0.1 -port 18096`.
- Fixture file `handler.sh` is created.
- Fixture file `hooks.json` is created.
#### Inputs
_Fixture `handler.sh`:_
```text
#!/bin/sh
printf 'signed\n' > ran.txt
printf 'ok'
```
_Fixture `hooks.json`:_
```text
[
  {
    "id": "signed",
    "execute-command": "${workdir}/handler.sh",
    "command-working-directory": "${workdir}",
    "include-command-output-in-response": true,
    "trigger-rule": {
      "match": {
        "type": "payload-hmac-sha256",
        "secret": "atago-demo-secret",
        "parameter": { "source": "header", "name": "X-Hub-Signature-256" }
      }
    }
  }
]
```
#### When
```shell
# HTTP POST /hooks/signed
# HTTP POST /hooks/signed
```
#### Then
- after `HTTP POST /hooks/signed`:
  - HTTP status is `500`
  - body contains `Error occurred while evaluating hook rules`
  - file `ran.txt` does not exist
- after `HTTP POST /hooks/signed`:
  - HTTP status is `200`
  - file `ran.txt` contains `signed`
### Scenario: the http-methods allowlist rejects the wrong verb
#### Given
- Background service `webhook` is started: `webhook -hooks hooks.json -ip 127.0.0.1 -port 18097`.
- Fixture file `handler.sh` is created.
- Fixture file `hooks.json` is created.
#### Inputs
_Fixture `handler.sh`:_
```text
#!/bin/sh
printf 'ok'
```
_Fixture `hooks.json`:_
```text
[
  {
    "id": "postonly",
    "execute-command": "${workdir}/handler.sh",
    "command-working-directory": "${workdir}",
    "http-methods": ["POST"],
    "include-command-output-in-response": true
  }
]
```
#### When
```shell
# HTTP GET /hooks/postonly
```
#### Then
- HTTP status is `405`
### Scenario: a command that exits non-zero surfaces as a 500
#### Given
- Background service `webhook` is started: `webhook -hooks hooks.json -ip 127.0.0.1 -port 18098`.
- Fixture file `handler.sh` is created.
- Fixture file `hooks.json` is created.
#### Inputs
_Fixture `handler.sh`:_
```text
#!/bin/sh
printf 'boom\n' >&2
exit 3
```
_Fixture `hooks.json`:_
```text
[
  {
    "id": "failer",
    "execute-command": "${workdir}/handler.sh",
    "command-working-directory": "${workdir}",
    "include-command-output-in-response": true
  }
]
```
#### When
```shell
# HTTP POST /hooks/failer
```
#### Then
- HTTP status is `500`
- body contains `Error occurred while executing the hook's command`
