# atago Behavior Specs
## Summary
2 suites · 6 scenarios
## Contents
- [age (modern file encryption)](#age-modern-file-encryption) — 5 scenarios
  - [keygen writes a key and reports the public half](#scenario-keygen-writes-a-key-and-reports-the-public-half)
  - [encrypt then decrypt round-trips binary bytes exactly](#scenario-encrypt-then-decrypt-round-trips-binary-bytes-exactly)
  - [armored output is PEM-wrapped](#scenario-armored-output-is-pem-wrapped)
  - [decrypting with the wrong identity fails](#scenario-decrypting-with-the-wrong-identity-fails)
  - [passphrase mode encrypts and decrypts interactively](#scenario-passphrase-mode-encrypts-and-decrypts-interactively)
- [age + changes (single-artifact generator)](#age--changes-single-artifact-generator) — 1 scenario
  - [age-keygen writes exactly the key file (HOME untouched)](#scenario-age-keygen-writes-exactly-the-key-file-home-untouched)
## age (modern file encryption)
Source: `test/e2e/thirdparty/age/age.atago.yaml`
### Scenario: keygen writes a key and reports the public half
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
```
#### Then
- after `age-keygen -o key.txt`:
  - exit code is `0`
  - stderr contains `Public key:`
  - file `key.txt` contains `AGE-SECRET-KEY-1`
- after `age-keygen -y key.txt`:
  - exit code is `0`
  - stdout matches `/(?m)^age1[a-z0-9]+$/`
### Scenario: encrypt then decrypt round-trips binary bytes exactly
#### Given
- Fixture file `data.bin` is created.
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${pubkey} from stdout
age -r ${pubkey} -o secret.age data.bin
age -d -i key.txt -o out.bin secret.age
cmp data.bin out.bin
```
#### Then
- after `age-keygen -o key.txt`:
  - exit code is `0`
- after `age -r ${pubkey} -o secret.age data.bin`:
  - exit code is `0`
- after `age -d -i key.txt -o out.bin secret.age`:
  - exit code is `0`
- after `cmp data.bin out.bin`:
  - exit code is `0`
### Scenario: armored output is PEM-wrapped
#### Given
- Fixture file `msg.txt` is created.
#### Inputs
_Fixture `msg.txt`:_
```text
armor me
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${pubkey} from stdout
age -a -r ${pubkey} -o armored.age msg.txt
```
#### Then
- after `age-keygen -o key.txt`:
  - exit code is `0`
- after `age -a -r ${pubkey} -o armored.age msg.txt`:
  - exit code is `0`
  - file `armored.age` contains `-----BEGIN AGE ENCRYPTED FILE-----`
### Scenario: decrypting with the wrong identity fails
#### Given
- Fixture file `msg.txt` is created.
#### Inputs
_Fixture `msg.txt`:_
```text
for the right key only
```
#### When
```shell
age-keygen -o key.txt
age-keygen -y key.txt
# capture ${pubkey} from stdout
age -r ${pubkey} -o secret.age msg.txt
age-keygen -o other.txt
age -d -i other.txt secret.age
```
#### Then
- after `age-keygen -o key.txt`:
  - exit code is `0`
- after `age -r ${pubkey} -o secret.age msg.txt`:
  - exit code is `0`
- after `age-keygen -o other.txt`:
  - exit code is `0`
- after `age -d -i other.txt secret.age`:
  - exit code is `1`
  - stderr contains `no identity matched`
### Scenario: passphrase mode encrypts and decrypts interactively
_skipped on windows_
#### Given
- Fixture file `msg.txt` is created.
#### Inputs
_Fixture `msg.txt`:_
```text
passphrase protected
```
#### When
```shell
# interactive (pty): age -p -o secret.age msg.txt
# interactive (pty): age -d -o out.txt secret.age
```
#### Then
- exit code is `0`
- file `secret.age` exists
- exit code is `0`
- file `out.txt` contains `passphrase protected`
#### Generated artifacts
- `secret.age`
## age + changes (single-artifact generator)
Source: `test/e2e/thirdparty/age/changes.atago.yaml`
### Scenario: age-keygen writes exactly the key file (HOME untouched)
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
age-keygen -o key.txt
```
#### Then
- exit code is `0`
- the step changed exactly created `key.txt`, modified nothing, deleted nothing
- file `key.txt` contains `AGE-SECRET-KEY-1`