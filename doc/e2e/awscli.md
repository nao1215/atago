# atago Behavior Specs
## Summary
1 suite · 4 scenarios
## Contents
- [aws-cli against MinIO (offline cloud-CLI story)](#aws-cli-against-minio-offline-cloud-cli-story) — 4 scenarios
  - [bucket and object lifecycle round-trips byte-identically](#scenario-bucket-and-object-lifecycle-round-trips-byte-identically)
  - [head-object and list-objects expose a JSON contract](#scenario-head-object-and-list-objects-expose-a-json-contract)
  - [a presigned URL is fetchable without credentials](#scenario-a-presigned-url-is-fetchable-without-credentials)
  - [head-object on a missing key fails with Not Found](#scenario-head-object-on-a-missing-key-fails-with-not-found)
## aws-cli against MinIO (offline cloud-CLI story)
Source: `test/e2e/thirdparty/awscli/awscli.atago.yaml`
### Scenario: bucket and object lifecycle round-trips byte-identically
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18530`.
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
hello from atago via the aws cli
```
#### When
```shell
aws --endpoint-url http://127.0.0.1:18530 s3 mb s3://mybucket
aws --endpoint-url http://127.0.0.1:18530 s3 cp payload.txt s3://mybucket/key.txt
aws --endpoint-url http://127.0.0.1:18530 s3 ls s3://mybucket/
aws --endpoint-url http://127.0.0.1:18530 s3 cp s3://mybucket/key.txt roundtrip.txt
aws --endpoint-url http://127.0.0.1:18530 s3 rm s3://mybucket/key.txt
aws --endpoint-url http://127.0.0.1:18530 s3 ls s3://mybucket/
```
#### Then
- after `aws --endpoint-url http://127.0.0.1:18530 s3 mb s3://mybucket`:
  - exit code is `0`
  - stdout contains `make_bucket: mybucket`
- after `aws --endpoint-url http://127.0.0.1:18530 s3 cp payload.txt s3://mybucket/key.txt`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:18530 s3 ls s3://mybucket/`:
  - exit code is `0`
  - stdout contains `key.txt`
- after `aws --endpoint-url http://127.0.0.1:18530 s3 cp s3://mybucket/key.txt roundtrip.txt`:
  - exit code is `0`
  - file `roundtrip.txt` contains `hello from atago via the aws cli`
- after `aws --endpoint-url http://127.0.0.1:18530 s3 rm s3://mybucket/key.txt`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:18530 s3 ls s3://mybucket/`:
  - exit code is `0`
  - stdout does not contain `key.txt`
### Scenario: head-object and list-objects expose a JSON contract
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18531`.
- Fixture file `obj.txt` is created.
#### Inputs
_Fixture `obj.txt`:_
```text
twelve bytes
```
#### When
```shell
aws --endpoint-url http://127.0.0.1:18531 s3 mb s3://jsonbucket
aws --endpoint-url http://127.0.0.1:18531 s3 cp obj.txt s3://jsonbucket/obj.txt
aws --endpoint-url http://127.0.0.1:18531 s3api head-object --bucket jsonbucket --key obj.txt --output json
aws --endpoint-url http://127.0.0.1:18531 s3 ls s3://jsonbucket/
```
#### Then
- after `aws --endpoint-url http://127.0.0.1:18531 s3 mb s3://jsonbucket`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:18531 s3 cp obj.txt s3://jsonbucket/obj.txt`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:18531 s3api head-object --bucket jsonbucket --key obj.txt --output json`:
  - exit code is `0`
  - stdout at `$.ContentLength` equals `12`
  - stdout at `$.ETag` matches `/^"[0-9a-f]{32}"$/`
- after `aws --endpoint-url http://127.0.0.1:18531 s3 ls s3://jsonbucket/`:
  - exit code is `0`
  - stdout contains `obj.txt`
### Scenario: a presigned URL is fetchable without credentials
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18532`.
- Fixture file `signed.txt` is created.
#### Inputs
_Fixture `signed.txt`:_
```text
presigned body content
```
#### When
```shell
aws --endpoint-url http://127.0.0.1:18532 s3 mb s3://signbucket
aws --endpoint-url http://127.0.0.1:18532 s3 cp signed.txt s3://signbucket/signed.txt
aws --endpoint-url http://127.0.0.1:18532 s3 presign s3://signbucket/signed.txt
# capture ${url} from stdout
curl -s "${url}"
```
#### Then
- after `aws --endpoint-url http://127.0.0.1:18532 s3 mb s3://signbucket`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:18532 s3 cp signed.txt s3://signbucket/signed.txt`:
  - exit code is `0`
- after `curl -s "${url}"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: head-object on a missing key fails with Not Found
#### Given
- Background service `minio` is started: `minio server data --address 127.0.0.1:18533`.
#### When
```shell
aws --endpoint-url http://127.0.0.1:18533 s3 mb s3://errbucket
aws --endpoint-url http://127.0.0.1:18533 s3api head-object --bucket errbucket --key nope.txt
```
#### Then
- after `aws --endpoint-url http://127.0.0.1:18533 s3 mb s3://errbucket`:
  - exit code is `0`
- after `aws --endpoint-url http://127.0.0.1:18533 s3api head-object --bucket errbucket --key nope.txt`:
  - exit code is one of `254`, `255`
  - stderr contains `Not Found`