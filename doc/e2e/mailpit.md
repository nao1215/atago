# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [mailpit (self-hosted email testing server)](#mailpit-self-hosted-email-testing-server) — 5 scenarios
  - [the binary reports its version](#scenario-the-binary-reports-its-version)
  - [a message sent over real SMTP is captured and readable via the API](#scenario-a-message-sent-over-real-smtp-is-captured-and-readable-via-the-api)
  - [full-text search finds exactly the matching message](#scenario-full-text-search-finds-exactly-the-matching-message)
  - [a MIME attachment survives delivery intact](#scenario-a-mime-attachment-survives-delivery-intact)
  - [deleting all messages empties the mailbox](#scenario-deleting-all-messages-empties-the-mailbox)

## mailpit (self-hosted email testing server)
[Mailpit](https://mailpit.axllent.org/) catches email so it can be
inspected, and testing it means using two protocols at once: mail goes in
over SMTP and comes back out over HTTP.

Messages are delivered by the stock `curl` client speaking real SMTP — no
library shim, no injected fixture — and everything after that is asserted
through Mailpit's REST API: that the message was captured with its headers
and body intact, that full-text search finds it, that a MIME attachment
survives as a distinct part, and that clearing the mailbox really empties
it.

Source: `test/e2e/thirdparty/mailpit/mailpit.atago.yaml`
Network policy: egress is allowed only to `127.0.0.1`.
### Scenario: the binary reports its version
_only when `mailpit version --no-release-check` succeeds_
#### When
```shell
mailpit version --no-release-check
```
#### Then
- exit code is `0`
- stdout contains `mailpit`

### Scenario: a message sent over real SMTP is captured and readable via the API
_only when `mailpit version --no-release-check` succeeds_
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18170 --listen 127.0.0.1:18171 --database data.db`.
- Fixture file `mail.txt` is created.
- The step is retried up to 20 times every 250ms until body at `$.total` equals `1`.

#### Inputs
_Fixture `mail.txt`:_
```text
From: Alice <alice@example.test>
To: Bob <bob@example.test>
Subject: Deploy finished

The deploy pipeline completed successfully.
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18170 --mail-from alice@example.test --mail-rcpt bob@example.test --upload-file mail.txt
# HTTP GET /api/v1/messages via api
# capture ${msg_id} from the response body
# HTTP GET /api/v1/message/${msg_id} via api
```
#### Then
- after `curl -s --url smtp://127.0.0.1:18170 --mail-from alice@example.test --mail-rcpt bob@example.test --upload-file mail.txt`:
  - exit code is `0`
- after `HTTP GET /api/v1/messages`:
  - HTTP status is `200`
  - body at `$.messages[0].Subject` equals `Deploy finished`
  - body at `$.messages[0].From.Address` equals `alice@example.test`
- after `HTTP GET /api/v1/message/${msg_id}`:
  - HTTP status is `200`
  - body at `$.Text` matches `/deploy pipeline completed successfully/`

### Scenario: full-text search finds exactly the matching message
_only when `mailpit version --no-release-check` succeeds_
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18172 --listen 127.0.0.1:18173 --database data.db`.
- Fixture file `mail1.txt` is created.
- Fixture file `mail2.txt` is created.
- The step is retried up to 20 times every 250ms until body at `$.total` equals `2`.

#### Inputs
_Fixture `mail1.txt`:_
```text
From: ci@example.test
To: team@example.test
Subject: nightly build report

All tests were green tonight.
```
_Fixture `mail2.txt`:_
```text
From: ci@example.test
To: team@example.test
Subject: invoice reminder

Please pay the hosting invoice.
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18172 --mail-from ci@example.test --mail-rcpt team@example.test --upload-file mail1.txt
curl -s --url smtp://127.0.0.1:18172 --mail-from ci@example.test --mail-rcpt team@example.test --upload-file mail2.txt
# HTTP GET /api/v1/messages via api2
# HTTP GET /api/v1/search?query=nightly via api2
```
#### Then
- after `HTTP GET /api/v1/search?query=nightly`:
  - HTTP status is `200`
  - body at `$.messages_count` equals `1`
  - body at `$.messages[0].Subject` equals `nightly build report`

### Scenario: a MIME attachment survives delivery intact
_only when `mailpit version --no-release-check` succeeds_
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18174 --listen 127.0.0.1:18175 --database data.db`.
- Fixture file `mail.txt` is created.
- The step is retried up to 20 times every 250ms until body at `$.total` equals `1`.

#### Inputs
_Fixture `mail.txt`:_
```text
From: reports@example.test
To: audit@example.test
Subject: weekly numbers
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="atago-boundary"

--atago-boundary
Content-Type: text/plain

Numbers attached as CSV.
--atago-boundary
Content-Type: text/csv; name="data.csv"
Content-Disposition: attachment; filename="data.csv"

region,total
east,42
--atago-boundary--
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18174 --mail-from reports@example.test --mail-rcpt audit@example.test --upload-file mail.txt
# HTTP GET /api/v1/messages via api3
# capture ${msg_id} from the response body
# HTTP GET /api/v1/message/${msg_id} via api3
```
#### Then
- after `HTTP GET /api/v1/message/${msg_id}`:
  - HTTP status is `200`
  - body at `$.Attachments` has length 1
  - body at `$.Attachments[0].FileName` equals `data.csv`

### Scenario: deleting all messages empties the mailbox
_only when `mailpit version --no-release-check` succeeds_
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18176 --listen 127.0.0.1:18177 --database data.db`.
- Fixture file `mail.txt` is created.
- The step is retried up to 20 times every 250ms until body at `$.total` equals `1`.

#### Inputs
_Fixture `mail.txt`:_
```text
From: temp@example.test
To: trash@example.test
Subject: ephemeral

Delete me.
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18176 --mail-from temp@example.test --mail-rcpt trash@example.test --upload-file mail.txt
# HTTP GET /api/v1/messages via api4
# HTTP DELETE /api/v1/messages via api4
# HTTP GET /api/v1/messages via api4
```
#### Then
- after `HTTP DELETE /api/v1/messages`:
  - HTTP status is `200`
- after `HTTP GET /api/v1/messages`:
  - HTTP status is `200`
  - body at `$.total` equals `0`
