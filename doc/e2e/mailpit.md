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
Source: `test/e2e/thirdparty/mailpit/mailpit.atago.yaml`
### Scenario: the binary reports its version
#### When
```shell
mailpit version
```
#### Then
- exit code is `0`
- stdout contains `mailpit`
### Scenario: a message sent over real SMTP is captured and readable via the API
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18170 --listen 127.0.0.1:18171 --database data.db`.
- Fixture file `mail.txt` is created.
#### Inputs
_Fixture `mail.txt`:_
```
From: Alice <alice@example.test>
To: Bob <bob@example.test>
Subject: Deploy finished

The deploy pipeline completed successfully.
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18170 --mail-from alice@example.test --mail-rcpt bob@example.test --upload-file mail.txt
# HTTP GET /api/v1/messages
# capture ${msg_id} from the response body
# HTTP GET /api/v1/message/${msg_id}
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
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18172 --listen 127.0.0.1:18173 --database data.db`.
- Fixture file `mail1.txt` is created.
- Fixture file `mail2.txt` is created.
#### Inputs
_Fixture `mail1.txt`:_
```
From: ci@example.test
To: team@example.test
Subject: nightly build report

All tests were green tonight.
```
_Fixture `mail2.txt`:_
```
From: ci@example.test
To: team@example.test
Subject: invoice reminder

Please pay the hosting invoice.
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18172 --mail-from ci@example.test --mail-rcpt team@example.test --upload-file mail1.txt
curl -s --url smtp://127.0.0.1:18172 --mail-from ci@example.test --mail-rcpt team@example.test --upload-file mail2.txt
# HTTP GET /api/v1/messages
# HTTP GET /api/v1/search?query=nightly
```
#### Then
- after `HTTP GET /api/v1/search?query=nightly`:
  - HTTP status is `200`
  - body at `$.messages_count` equals `1`
  - body at `$.messages[0].Subject` equals `nightly build report`
### Scenario: a MIME attachment survives delivery intact
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18174 --listen 127.0.0.1:18175 --database data.db`.
- Fixture file `mail.txt` is created.
#### Inputs
_Fixture `mail.txt`:_
```
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
# HTTP GET /api/v1/messages
# capture ${msg_id} from the response body
# HTTP GET /api/v1/message/${msg_id}
```
#### Then
- after `HTTP GET /api/v1/message/${msg_id}`:
  - HTTP status is `200`
  - body at `$.Attachments` has length 1
  - body at `$.Attachments[0].FileName` equals `data.csv`
### Scenario: deleting all messages empties the mailbox
#### Given
- Background service `mailpit` is started: `mailpit --smtp 127.0.0.1:18176 --listen 127.0.0.1:18177 --database data.db`.
- Fixture file `mail.txt` is created.
#### Inputs
_Fixture `mail.txt`:_
```
From: temp@example.test
To: trash@example.test
Subject: ephemeral

Delete me.
```
#### When
```shell
curl -s --url smtp://127.0.0.1:18176 --mail-from temp@example.test --mail-rcpt trash@example.test --upload-file mail.txt
# HTTP GET /api/v1/messages
# HTTP DELETE /api/v1/messages
# HTTP GET /api/v1/messages
```
#### Then
- after `HTTP DELETE /api/v1/messages`:
  - HTTP status is `200`
- after `HTTP GET /api/v1/messages`:
  - HTTP status is `200`
  - body at `$.total` equals `0`