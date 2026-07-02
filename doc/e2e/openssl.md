# atago Behavior Specs
## Summary
1 suite · 7 scenarios
## Contents
- [openssl (third-party CLI, no build required)](#openssl-third-party-cli-no-build-required) — 7 scenarios
  - [sha256 digest of a fixed input is exact and stable](#scenario-sha256-digest-of-a-fixed-input-is-exact-and-stable)
  - [base64 encode and decode round-trip via stdin](#scenario-base64-encode-and-decode-round-trip-via-stdin)
  - [rand -hex emits exactly the requested entropy](#scenario-rand--hex-emits-exactly-the-requested-entropy)
  - [a generated private key is valid and yields its public half](#scenario-a-generated-private-key-is-valid-and-yields-its-public-half)
  - [symmetric encryption round-trips and a wrong password fails loudly](#scenario-symmetric-encryption-round-trips-and-a-wrong-password-fails-loudly)
  - [a self-signed certificate carries the requested subject](#scenario-a-self-signed-certificate-carries-the-requested-subject)
  - [signing with the private key verifies with the public key](#scenario-signing-with-the-private-key-verifies-with-the-public-key)
## openssl (third-party CLI, no build required)
Source: `test/e2e/thirdparty/openssl/openssl.atago.yaml`
### Scenario: sha256 digest of a fixed input is exact and stable
#### Given
- Fixture file `msg.txt` is created.
#### Inputs
_Fixture `msg.txt`:_
```text
the quick brown fox
```
#### When
```shell
openssl dgst -sha256 -r msg.txt
```
#### Then
- exit code is `0`
- stdout contains `6e459fed18ddb06d57c8e9f0d000c302c7e01389926db6e89884bfbe91a2a5df`
### Scenario: base64 encode and decode round-trip via stdin
#### Inputs
_stdin for `openssl`:_
```text
round-trip me
```
_stdin for `openssl`:_
```text
${encoded}
```
#### When
```shell
openssl base64
# capture ${encoded} from stdout
openssl base64 -d
```
#### Then
- after `openssl base64 -d`:
  - exit code is `0`
  - stdout contains `round-trip me`
### Scenario: rand -hex emits exactly the requested entropy
#### When
```shell
openssl rand -hex 16
```
#### Then
- exit code is `0`
- stdout matches `/^[0-9a-f]{32}
?$/`
### Scenario: a generated private key is valid and yields its public half
#### When
```shell
openssl genpkey -algorithm ed25519 -out key.pem
openssl pkey -in key.pem -check -noout
openssl pkey -in key.pem -pubout -out pub.pem
```
#### Then
- after `openssl genpkey -algorithm ed25519 -out key.pem`:
  - exit code is `0`
  - file `key.pem` contains `BEGIN PRIVATE KEY`
- after `openssl pkey -in key.pem -check -noout`:
  - exit code is `0`
  - stdout contains `Key is valid`
- after `openssl pkey -in key.pem -pubout -out pub.pem`:
  - exit code is `0`
  - file `pub.pem` contains `BEGIN PUBLIC KEY`
### Scenario: symmetric encryption round-trips and a wrong password fails loudly
#### Given
- Fixture file `secret.txt` is created.
#### Inputs
_Fixture `secret.txt`:_
```text
attack at dawn
```
#### When
```shell
openssl enc -aes-256-cbc -pbkdf2 -pass pass:correct-horse -in secret.txt -out secret.enc
openssl enc -d -aes-256-cbc -pbkdf2 -pass pass:correct-horse -in secret.enc -out roundtrip.txt
openssl enc -d -aes-256-cbc -pbkdf2 -pass pass:wrong-password -in secret.enc -out garbage.txt
```
#### Then
- after `openssl enc -aes-256-cbc -pbkdf2 -pass pass:correct-horse -in secret.txt -out secret.enc`:
  - exit code is `0`
  - file `secret.enc` exists
  - file `secret.enc` is checked
- after `openssl enc -d -aes-256-cbc -pbkdf2 -pass pass:correct-horse -in secret.enc -out roundtrip.txt`:
  - exit code is `0`
  - file `roundtrip.txt` contains `attack at dawn`
- after `openssl enc -d -aes-256-cbc -pbkdf2 -pass pass:wrong-password -in secret.enc -out garbage.txt`:
  - exit code is not `0`
  - stderr contains `bad decrypt`
#### Generated artifacts
- `secret.enc`
### Scenario: a self-signed certificate carries the requested subject
#### When
```shell
openssl req -x509 -newkey ed25519 -keyout ca.key -out ca.crt -nodes -days 1 -subj /CN=atago-test
openssl x509 -in ca.crt -noout -subject
openssl verify -CAfile ca.crt ca.crt
```
#### Then
- after `openssl req -x509 -newkey ed25519 -keyout ca.key -out ca.crt -nodes -days 1 -subj /CN=atago-test`:
  - exit code is `0`
- after `openssl x509 -in ca.crt -noout -subject`:
  - exit code is `0`
  - stdout contains `atago-test`
- after `openssl verify -CAfile ca.crt ca.crt`:
  - exit code is `0`
  - stdout contains `OK`
### Scenario: signing with the private key verifies with the public key
#### Given
- Fixture file `payload.txt` is created.
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
sign me
```
_Fixture `payload.txt`:_
```text
sign me (tampered)
```
#### When
```shell
openssl genpkey -algorithm ed25519 -out sk.pem
openssl pkey -in sk.pem -pubout -out vk.pem
openssl pkeyutl -sign -inkey sk.pem -rawin -in payload.txt -out payload.sig
openssl pkeyutl -verify -pubin -inkey vk.pem -rawin -in payload.txt -sigfile payload.sig
openssl pkeyutl -verify -pubin -inkey vk.pem -rawin -in payload.txt -sigfile payload.sig
```
#### Then
- after `openssl pkeyutl -sign -inkey sk.pem -rawin -in payload.txt -out payload.sig`:
  - exit code is `0`
- after `openssl pkeyutl -verify -pubin -inkey vk.pem -rawin -in payload.txt -sigfile payload.sig`:
  - exit code is `0`
  - stdout contains `Signature Verified Successfully`
- after `openssl pkeyutl -verify -pubin -inkey vk.pem -rawin -in payload.txt -sigfile payload.sig`:
  - exit code is not `0`