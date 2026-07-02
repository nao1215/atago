# atago Behavior Specs
## Summary
1 suite · 1 scenario
## Contents
- [webhook (self-hosted webhook receiver)](#webhook-self-hosted-webhook-receiver) — 1 scenario
  - [a JSON post triggers the configured command](#scenario-a-json-post-triggers-the-configured-command)
## webhook (self-hosted webhook receiver)
Source: `test/e2e/thirdparty/webhook/webhook.atago.yaml`
### Scenario: a JSON post triggers the configured command
#### Given
- Background service `webhook` is started: `webhook -hooks hooks.json -ip 127.0.0.1 -port 18094`.
- Fixture file `hooks.json` is created.
#### Inputs
_Fixture `hooks.json`:_
```
[
  {
    "id": "greet",
    "execute-command": "/bin/echo",
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
  - body contains `Alice`
- after `HTTP POST /hooks/nope`:
  - HTTP status is `404`