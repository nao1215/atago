# atago Behavior Specs
## Summary
1 suite · 7 scenarios
## Contents
- [sops + age (secrets encryption)](#sops--age-secrets-encryption) — 7 scenarios
  - [version prints a semantic version offline](#scenario-version-prints-a-semantic-version-offline)
  - [encryption hides values, keeps keys, and records metadata](#scenario-encryption-hides-values-keeps-keys-and-records-metadata)
  - [encrypt then decrypt recovers the original values](#scenario-encrypt-then-decrypt-recovers-the-original-values)
  - [extract returns a single decrypted value exactly](#scenario-extract-returns-a-single-decrypted-value-exactly)
  - [decrypting with the wrong key fails](#scenario-decrypting-with-the-wrong-key-fails)
  - [a tampered ciphertext fails the MAC](#scenario-a-tampered-ciphertext-fails-the-mac)
  - [an encrypted-regex scopes which keys are encrypted](#scenario-an-encrypted-regex-scopes-which-keys-are-encrypted)
## sops + age (secrets encryption)
Source: `test/e2e/thirdparty/sops/sops.atago.yaml`
### Scenario: version prints a semantic version offline
#### When
```shell
sops --version --disable-version-check
```
#### Then
- exit code is `0`
- stdout matches `/sops [0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: encryption hides values, keeps keys, and records metadata
#### Given
- Fixture file `secrets.yaml` is created.
#### Inputs
_Fixture `secrets.yaml`:_
```text
database:
  password: hunter2
  host: db.internal
region: us-west-1
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${recipient} from stdout
sops encrypt --age ${recipient} secrets.yaml
```
#### Then
- after `age-keygen -o key.txt`:
  - exit code is `0`
- after `sops encrypt --age ${recipient} secrets.yaml`:
  - exit code is `0`
  - stdout contains `password:`, `host:`, `ENC[AES256_GCM`, `sops:`, `mac:`
  - stdout does not contain `hunter2`, `db.internal`
### Scenario: encrypt then decrypt recovers the original values
#### Given
- Fixture file `secrets.yaml` is created.
- Environment variables are set: SOPS_AGE_KEY_FILE.
#### Inputs
_Fixture `secrets.yaml`:_
```text
database:
  password: hunter2
  host: db.internal
region: us-west-1
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${recipient} from stdout
sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml
sops decrypt secrets.enc.yaml
```
#### Then
- after `sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml`:
  - exit code is `0`
- after `sops decrypt secrets.enc.yaml`:
  - exit code is `0`
  - stdout contains `password: hunter2`, `host: db.internal`, `region: us-west-1`
### Scenario: extract returns a single decrypted value exactly
#### Given
- Fixture file `secrets.yaml` is created.
- Environment variables are set: SOPS_AGE_KEY_FILE.
#### Inputs
_Fixture `secrets.yaml`:_
```text
database:
  password: hunter2
  host: db.internal
region: us-west-1
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${recipient} from stdout
sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml
sops decrypt --extract '["region"]' secrets.enc.yaml
```
#### Then
- after `sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml`:
  - exit code is `0`
- after `sops decrypt --extract '["region"]' secrets.enc.yaml`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: decrypting with the wrong key fails
#### Given
- Fixture file `secrets.yaml` is created.
- Environment variables are set: SOPS_AGE_KEY_FILE.
#### Inputs
_Fixture `secrets.yaml`:_
```text
token: changeme
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${recipient} from stdout
sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml
age-keygen -o other.txt
sops decrypt secrets.enc.yaml
```
#### Then
- after `sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml`:
  - exit code is `0`
- after `age-keygen -o other.txt`:
  - exit code is `0`
- after `sops decrypt secrets.enc.yaml`:
  - exit code is `128`
  - stderr contains `Failed to get the data key`
### Scenario: a tampered ciphertext fails the MAC
#### Given
- Fixture file `secrets.yaml` is created.
- Environment variables are set: SOPS_AGE_KEY_FILE.
#### Inputs
_Fixture `secrets.yaml`:_
```text
token: changeme
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${recipient} from stdout
sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml
sed '0,/data:/ s/data:\([A-Za-z0-9]\)/data:Z/' secrets.enc.yaml > tampered.yaml
sops decrypt tampered.yaml
```
#### Then
- after `sops encrypt --age ${recipient} --output secrets.enc.yaml secrets.yaml`:
  - exit code is `0`
- after `sed '0,/data:/ s/data:\([A-Za-z0-9]\)/data:Z/' secrets.enc.yaml > tampered.yaml`:
  - exit code is `0`
- after `sops decrypt tampered.yaml`:
  - exit code is `25`
  - stderr contains `message authentication failed`
### Scenario: an encrypted-regex scopes which keys are encrypted
#### Given
- Fixture file `secrets.yaml` is created.
#### Inputs
_Fixture `secrets.yaml`:_
```text
database:
  password: hunter2
  host: db.internal
region: us-west-1
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${recipient} from stdout
sops encrypt --age ${recipient} --encrypted-regex '^password$' secrets.yaml
```
#### Then
- after `sops encrypt --age ${recipient} --encrypted-regex '^password$' secrets.yaml`:
  - exit code is `0`
  - stdout contains `host: db.internal`, `region: us-west-1`, `password: ENC[AES256_GCM`
  - stdout does not contain `password: hunter2`