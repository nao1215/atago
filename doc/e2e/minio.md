# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [minio (self-hosted object storage)](#minio-self-hosted-object-storage) — 5 scenarios
  - [the server reports itself alive over the health API](#scenario-the-server-reports-itself-alive-over-the-health-api)
  - [anonymous S3 access is denied with a proper S3 XML error](#scenario-anonymous-s3-access-is-denied-with-a-proper-s3-xml-error)
  - [a full object lifecycle through the mc client](#scenario-a-full-object-lifecycle-through-the-mc-client)
  - [bucket versioning can be enabled and reported](#scenario-bucket-versioning-can-be-enabled-and-reported)
  - [an anonymous download policy publishes a bucket read-only](#scenario-an-anonymous-download-policy-publishes-a-bucket-read-only)
## minio (self-hosted object storage)
Source: `test/e2e/thirdparty/minio/minio.atago.yaml`
### Scenario: the server reports itself alive over the health API
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18120`.
#### When
```shell
# HTTP GET /minio/health/live
# HTTP GET /minio/health/ready
```
#### Then
- after `HTTP GET /minio/health/live`:
  - HTTP status is `200`
- after `HTTP GET /minio/health/ready`:
  - HTTP status is `200`
### Scenario: anonymous S3 access is denied with a proper S3 XML error
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18121`.
#### When
```shell
# HTTP GET /
```
#### Then
- HTTP status is `403`
- body contains `<Code>AccessDenied</Code>`
### Scenario: a full object lifecycle through the mc client
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18122`.
- Fixture file `upload.txt` is created.
#### Inputs
_Fixture `upload.txt`:_
```text
hello object storage
```
#### When
```shell
mc alias set lifecycle http://127.0.0.1:18122 atago atago-secret-key
mc mb lifecycle/atago-bucket
mc cp upload.txt lifecycle/atago-bucket/
mc ls --json lifecycle/atago-bucket
mc cat lifecycle/atago-bucket/upload.txt
mc cp lifecycle/atago-bucket/upload.txt downloaded.txt
mc rm lifecycle/atago-bucket/upload.txt
mc ls lifecycle/atago-bucket
```
#### Then
- after `mc alias set lifecycle http://127.0.0.1:18122 atago atago-secret-key`:
  - exit code is `0`
- after `mc mb lifecycle/atago-bucket`:
  - exit code is `0`
  - stdout contains `Bucket created successfully`
- after `mc cp upload.txt lifecycle/atago-bucket/`:
  - exit code is `0`
- after `mc ls --json lifecycle/atago-bucket`:
  - exit code is `0`
  - stdout at `$.key` equals `upload.txt`
  - stdout at `$.size` equals `20`
- after `mc cat lifecycle/atago-bucket/upload.txt`:
  - exit code is `0`
  - stdout equals an exact value
- after `mc cp lifecycle/atago-bucket/upload.txt downloaded.txt`:
  - exit code is `0`
  - file `downloaded.txt` contains `hello object storage`
- after `mc rm lifecycle/atago-bucket/upload.txt`:
  - exit code is `0`
  - stdout contains `Removed`
- after `mc ls lifecycle/atago-bucket`:
  - exit code is `0`
  - stdout is empty
### Scenario: bucket versioning can be enabled and reported
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18123`.
#### When
```shell
mc alias set versioned http://127.0.0.1:18123 atago atago-secret-key
mc mb versioned/versioned
mc version enable versioned/versioned
mc version info --json versioned/versioned
```
#### Then
- after `mc version enable versioned/versioned`:
  - exit code is `0`
  - stdout contains `versioning is enabled`
- after `mc version info --json versioned/versioned`:
  - exit code is `0`
  - stdout at `$.versioning.status` equals `Enabled`
### Scenario: an anonymous download policy publishes a bucket read-only
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18124`.
- Fixture file `page.txt` is created.
#### Inputs
_Fixture `page.txt`:_
```text
published via bucket policy
```
#### When
```shell
mc alias set publichost http://127.0.0.1:18124 atago atago-secret-key
mc mb publichost/public-bucket
mc cp page.txt publichost/public-bucket/
# HTTP GET /public-bucket/page.txt
mc anonymous set download publichost/public-bucket
# HTTP GET /public-bucket/page.txt
# HTTP PUT /public-bucket/forbidden.txt
```
#### Then
- after `HTTP GET /public-bucket/page.txt`:
  - HTTP status is `403`
- after `mc anonymous set download publichost/public-bucket`:
  - exit code is `0`
  - stdout contains `is set to `download``
- after `HTTP GET /public-bucket/page.txt`:
  - HTTP status is `200`
  - body contains `published via bucket policy`
- after `HTTP PUT /public-bucket/forbidden.txt`:
  - HTTP status is `403`
