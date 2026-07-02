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
```
hello object storage
```
#### When
```shell
mc alias set local http://127.0.0.1:18122 atago atago-secret-key
mc mb local/atago-bucket
mc cp upload.txt local/atago-bucket/
mc ls --json local/atago-bucket
mc cat local/atago-bucket/upload.txt
mc cp local/atago-bucket/upload.txt downloaded.txt
mc rm local/atago-bucket/upload.txt
mc ls local/atago-bucket
```
#### Then
- after `mc alias set local http://127.0.0.1:18122 atago atago-secret-key`:
  - exit code is `0`
- after `mc mb local/atago-bucket`:
  - exit code is `0`
  - stdout contains `Bucket created successfully`
- after `mc cp upload.txt local/atago-bucket/`:
  - exit code is `0`
- after `mc ls --json local/atago-bucket`:
  - exit code is `0`
  - stdout at `$.key` equals `upload.txt`
  - stdout at `$.size` equals `20`
- after `mc cat local/atago-bucket/upload.txt`:
  - exit code is `0`
  - stdout equals an exact value
- after `mc cp local/atago-bucket/upload.txt downloaded.txt`:
  - exit code is `0`
  - file `downloaded.txt` contains `hello object storage`
- after `mc rm local/atago-bucket/upload.txt`:
  - exit code is `0`
  - stdout contains `Removed`
- after `mc ls local/atago-bucket`:
  - exit code is `0`
  - stdout is empty
### Scenario: bucket versioning can be enabled and reported
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18123`.
#### When
```shell
mc alias set local http://127.0.0.1:18123 atago atago-secret-key
mc mb local/versioned
mc version enable local/versioned
mc version info --json local/versioned
```
#### Then
- after `mc version enable local/versioned`:
  - exit code is `0`
  - stdout contains `versioning is enabled`
- after `mc version info --json local/versioned`:
  - exit code is `0`
  - stdout at `$.versioning.status` equals `Enabled`
### Scenario: an anonymous download policy publishes a bucket read-only
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18124`.
- Fixture file `page.txt` is created.
#### Inputs
_Fixture `page.txt`:_
```
published via bucket policy
```
#### When
```shell
mc alias set local http://127.0.0.1:18124 atago atago-secret-key
mc mb local/public-bucket
mc cp page.txt local/public-bucket/
# HTTP GET /public-bucket/page.txt
mc anonymous set download local/public-bucket
# HTTP GET /public-bucket/page.txt
# HTTP PUT /public-bucket/forbidden.txt
```
#### Then
- after `HTTP GET /public-bucket/page.txt`:
  - HTTP status is `403`
- after `mc anonymous set download local/public-bucket`:
  - exit code is `0`
  - stdout contains `is set to `download``
- after `HTTP GET /public-bucket/page.txt`:
  - HTTP status is `200`
  - body contains `published via bucket policy`
- after `HTTP PUT /public-bucket/forbidden.txt`:
  - HTTP status is `403`