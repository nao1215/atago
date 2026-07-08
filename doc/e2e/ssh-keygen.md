# atago Behavior Specs
## Summary
1 suite · 6 scenarios
## Contents
- [ssh-keygen (OpenSSH key generation)](#ssh-keygen-openssh-key-generation) — 6 scenarios
  - [non-interactive generation writes the key pair](#scenario-non-interactive-generation-writes-the-key-pair)
  - [the fingerprint contract is exact](#scenario-the-fingerprint-contract-is-exact)
  - [-y regenerates the public key from the private key](#scenario--y-regenerates-the-public-key-from-the-private-key)
  - [-y on a corrupted key file fails](#scenario--y-on-a-corrupted-key-file-fails)
  - [interactive passphrase generation prompts twice](#scenario-interactive-passphrase-generation-prompts-twice)
  - [the wrong passphrase is rejected](#scenario-the-wrong-passphrase-is-rejected)
## ssh-keygen (OpenSSH key generation)
Source: `test/e2e/thirdparty/ssh-keygen/ssh-keygen.atago.yaml`
### Scenario: non-interactive generation writes the key pair
_only when `command -v ssh-keygen` succeeds_
#### When
```shell
ssh-keygen -t ed25519 -N '' -f key -C test@atago
```
#### Then
- exit code is `0`
- file `key` exists
- file `key.pub` exists
- file `key.pub` contains `ssh-ed25519`
- file `key.pub` is checked
#### Generated artifacts
- `key`
- `key.pub`
### Scenario: the fingerprint contract is exact
_only when `command -v ssh-keygen` succeeds_
#### When
```shell
ssh-keygen -t ed25519 -N '' -f key -C test@atago
ssh-keygen -lf key.pub
```
#### Then
- after `ssh-keygen -t ed25519 -N '' -f key -C test@atago`:
  - exit code is `0`
- after `ssh-keygen -lf key.pub`:
  - exit code is `0`
  - stdout matches `/(?m)^256 SHA256:[A-Za-z0-9+/]+ test@atago \(ED25519\)$/`
### Scenario: -y regenerates the public key from the private key
_only when `command -v ssh-keygen` succeeds_
#### When
```shell
ssh-keygen -t ed25519 -N '' -f key -C test@atago
ssh-keygen -y -f key
```
#### Then
- after `ssh-keygen -t ed25519 -N '' -f key -C test@atago`:
  - exit code is `0`
- after `ssh-keygen -y -f key`:
  - exit code is `0`
  - stdout contains `ssh-ed25519`
### Scenario: -y on a corrupted key file fails
_only when `command -v ssh-keygen` succeeds_
#### Given
- Fixture file `corrupt` is created.
#### Inputs
_Fixture `corrupt`:_
```text
this is not a private key
```
#### When
```shell
chmod 600 corrupt && ssh-keygen -y -f corrupt
```
#### Then
- exit code is one of `255`, `1`
- stderr contains `error in libcrypto`
### Scenario: interactive passphrase generation prompts twice
_only when `command -v ssh-keygen` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): ssh-keygen -t ed25519 -f protected -C test@atago
ssh-keygen -y -P "$PASSPHRASE" -f protected
```
#### Then
- exit code is `0`
- file `protected` exists
- file `protected.pub` exists
- file `protected.pub` contains `ssh-ed25519`
- exit code is `0`
- stdout contains `ssh-ed25519`
#### Generated artifacts
- `protected`
- `protected.pub`
### Scenario: the wrong passphrase is rejected
_only when `command -v ssh-keygen` succeeds · skipped on Windows_
#### When
```shell
# interactive (pty): ssh-keygen -t ed25519 -f protected -C test@atago
ssh-keygen -y -P definitely-not-the-passphrase -f protected
```
#### Then
- exit code is `0`
- exit code is one of `255`, `1`
- stderr contains `incorrect passphrase`
