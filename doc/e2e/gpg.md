# atago Behavior Specs
## Summary
1 suite · 8 scenarios
## Contents
- [GnuPG (third-party CLI, cryptographic contracts)](#gnupg-third-party-cli-cryptographic-contracts) — 8 scenarios
  - [generating a key creates a keyring that lists the identity](#scenario-generating-a-key-creates-a-keyring-that-lists-the-identity)
  - [a message survives an encrypt and decrypt round trip](#scenario-a-message-survives-an-encrypt-and-decrypt-round-trip)
  - [armor produces text that decodes back to the same bytes](#scenario-armor-produces-text-that-decodes-back-to-the-same-bytes)
  - [a single flipped byte is refused instead of decoded](#scenario-a-single-flipped-byte-is-refused-instead-of-decoded)
  - [a keyring without the secret key cannot read the message](#scenario-a-keyring-without-the-secret-key-cannot-read-the-message)
  - [a detached signature stops verifying when the file changes](#scenario-a-detached-signature-stops-verifying-when-the-file-changes)
  - [symmetric encryption answers only to its own passphrase](#scenario-symmetric-encryption-answers-only-to-its-own-passphrase)
  - [usage mistakes are refused offline, with the input named](#scenario-usage-mistakes-are-refused-offline-with-the-input-named)

## GnuPG (third-party CLI, cryptographic contracts)
[GnuPG](https://gnupg.org/) is the reference OpenPGP implementation, and its
contract is one where being approximately right is worse than failing: a
decryption that returns subtly wrong plaintext, or a verification that
accepts a modified file, is a security bug rather than a cosmetic one.

So the scenarios are built around round trips and negative space. Encrypting
and decrypting returns the original bytes; the ciphertext never contains the
plaintext; a single flipped byte is refused rather than decoded; a keyring
without the secret key cannot read the message; and a signature stops
verifying the moment the signed file changes.

gpg is also one of the few CLIs whose exit codes carry three different
meanings, and the machine-readable `--status-fd` protocol — GOODSIG, BADSIG,
NO_SECKEY, DECRYPTION_OKAY — is what a program integrating gpg actually
reads. Both are pinned here, in preference to the human prose, which is
translated and therefore not a contract at all.

Every key is generated inside the scenario workdir with a throwaway
passphrase, nothing is committed, and no network is used: key lookup is
disabled so a missing recipient fails locally instead of reaching for a key
server.

Source: `test/e2e/thirdparty/gpg/gpg.atago.yaml`
### Scenario: generating a key creates a keyring that lists the identity
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never
gpg --batch --list-keys --with-colons
gpg --batch --list-secret-keys --with-colons
```
#### Then
- after `mkdir -m 700 -p "$GNUPGHOME"`:
  - exit code is `0`
- after `gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never`:
  - exit code is `0`
- after `gpg --batch --list-keys --with-colons`:
  - exit code is `0`
  - stdout contains `pub:`, `Test User <test@example.com>`, matches `/(?m)^fpr:{9}[0-9A-F]{40}:/`
- after `gpg --batch --list-secret-keys --with-colons`:
  - exit code is `0`
  - stdout matches `/(?m)^sec:/`
  - dir `gnupg/private-keys-v1.d` has >= 1 entry, matches glob `*.key`

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: a message survives an encrypt and decrypt round trip
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `msg.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `msg.txt`:_
```text
attack at dawn
second line
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never
gpg --batch --trust-model always --auto-key-locate clear --encrypt --recipient test@example.com --output msg.gpg msg.txt
gpg --batch --pinentry-mode loopback --passphrase "" --output back.txt --decrypt msg.gpg
```
#### Then
- after `gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never`:
  - exit code is `0`
- after `gpg --batch --trust-model always --auto-key-locate clear --encrypt --recipient test@example.com --output msg.gpg msg.txt`:
  - exit code is `0`
  - the step changed exactly created `msg.gpg`, modified nothing, deleted nothing, ignoring `.atago-home/**`, `gnupg/**`
  - file `msg.gpg` does not contain `attack at dawn`, `second line`
- after `gpg --batch --pinentry-mode loopback --passphrase "" --output back.txt --decrypt msg.gpg`:
  - exit code is `0`
  - file `back.txt` is byte-identical to `msg.txt`

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: armor produces text that decodes back to the same bytes
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `msg.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `msg.txt`:_
```text
attack at dawn
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never
gpg --batch --trust-model always --auto-key-locate clear --armor --encrypt --recipient test@example.com --output msg.asc msg.txt
gpg --batch --pinentry-mode loopback --passphrase "" --output back.txt --decrypt msg.asc
```
#### Then
- after `gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never`:
  - exit code is `0`
- after `gpg --batch --trust-model always --auto-key-locate clear --armor --encrypt --recipient test@example.com --output msg.asc msg.txt`:
  - exit code is `0`
  - file `msg.asc` contains `-----BEGIN PGP MESSAGE-----`, `-----END PGP MESSAGE-----`
  - file `msg.asc` does not contain `attack at dawn`
- after `gpg --batch --pinentry-mode loopback --passphrase "" --output back.txt --decrypt msg.asc`:
  - exit code is `0`
  - file `back.txt` is byte-identical to `msg.txt`

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: a single flipped byte is refused instead of decoded
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `msg.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `msg.txt`:_
```text
attack at dawn
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never
gpg --batch --trust-model always --auto-key-locate clear --encrypt --recipient test@example.com --output msg.gpg msg.txt
printf 'X' | dd of=msg.gpg bs=1 seek=40 conv=notrunc status=none
gpg --batch --pinentry-mode loopback --passphrase "" --output back.txt --decrypt msg.gpg
```
#### Then
- after `gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never`:
  - exit code is `0`
- after `gpg --batch --trust-model always --auto-key-locate clear --encrypt --recipient test@example.com --output msg.gpg msg.txt`:
  - exit code is `0`
- after `printf 'X' | dd of=msg.gpg bs=1 seek=40 conv=notrunc status=none`:
  - exit code is `0`
- after `gpg --batch --pinentry-mode loopback --passphrase "" --output back.txt --decrypt msg.gpg`:
  - exit code is `2`
  - stderr contains `decryption failed`
  - file `back.txt` does not exist

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: a keyring without the secret key cannot read the message
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `msg.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `msg.txt`:_
```text
attack at dawn
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME" && mkdir -m 700 -p "${workdir}/other"
gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never
gpg --batch --trust-model always --auto-key-locate clear --encrypt --recipient test@example.com --output msg.gpg msg.txt
gpg --batch --status-fd 1 --output stolen.txt --decrypt msg.gpg
```
#### Then
- after `gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never`:
  - exit code is `0`
- after `gpg --batch --trust-model always --auto-key-locate clear --encrypt --recipient test@example.com --output msg.gpg msg.txt`:
  - exit code is `0`
- after `gpg --batch --status-fd 1 --output stolen.txt --decrypt msg.gpg`:
  - exit code is `2`
  - stdout contains `NO_SECKEY`, `DECRYPTION_FAILED`
  - file `stolen.txt` does not exist

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: a detached signature stops verifying when the file changes
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `release.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `release.txt`:_
```text
v1.0.0 checksum 0123456789abcdef
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never
gpg --batch --pinentry-mode loopback --passphrase "" --output release.sig --detach-sign release.txt
gpg --batch --status-fd 1 --verify release.sig release.txt
printf '!' >> release.txt
gpg --batch --status-fd 1 --verify release.sig release.txt
```
#### Then
- after `gpg --batch --pinentry-mode loopback --passphrase "" --quick-generate-key "Test User <test@example.com>" default default never`:
  - exit code is `0`
- after `gpg --batch --pinentry-mode loopback --passphrase "" --output release.sig --detach-sign release.txt`:
  - exit code is `0`
- after `gpg --batch --status-fd 1 --verify release.sig release.txt`:
  - exit code is `0`
  - stdout contains `GOODSIG`, `VALIDSIG`
- after `printf '!' >> release.txt`:
  - exit code is `0`
- after `gpg --batch --status-fd 1 --verify release.sig release.txt`:
  - exit code is `1`
  - stdout contains `BADSIG`, does not contain `GOODSIG`

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: symmetric encryption answers only to its own passphrase
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `msg.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `msg.txt`:_
```text
attack at dawn
```
_stdin for `gpg`:_
```text
correct-horse-battery-staple
```
_stdin for `gpg`:_
```text
correct-horse-battery-staple
```
_stdin for `gpg`:_
```text
not-the-passphrase
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --yes --pinentry-mode loopback --passphrase-fd 0 --symmetric --output msg.gpg msg.txt
gpg --batch --pinentry-mode loopback --passphrase-fd 0 --output back.txt --decrypt msg.gpg
gpg --batch --pinentry-mode loopback --passphrase-fd 0 --output wrong.txt --decrypt msg.gpg
```
#### Then
- after `gpg --batch --yes --pinentry-mode loopback --passphrase-fd 0 --symmetric --output msg.gpg msg.txt`:
  - exit code is `0`
  - file `msg.gpg` does not contain `attack at dawn`
- after `gpg --batch --pinentry-mode loopback --passphrase-fd 0 --output back.txt --decrypt msg.gpg`:
  - exit code is `0`
  - file `back.txt` is byte-identical to `msg.txt`
- after `gpg --batch --pinentry-mode loopback --passphrase-fd 0 --output wrong.txt --decrypt msg.gpg`:
  - exit code is `2`
  - stderr contains `decryption failed`
  - file `wrong.txt` does not exist

#### Finally (teardown, always runs)
```shell
gpgconf --kill all
```
### Scenario: usage mistakes are refused offline, with the input named
_only when `gpg --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `msg.txt` is created.
- Environment variables are set: GNUPGHOME, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `msg.txt`:_
```text
attack at dawn
```
#### When
```shell
mkdir -m 700 -p "$GNUPGHOME"
gpg --batch --nosuchoption
gpg --batch --decrypt does-not-exist.gpg
gpg --batch --auto-key-locate clear --encrypt --recipient nobody@example.com --output msg.gpg msg.txt
```
#### Then
- after `gpg --batch --nosuchoption`:
  - exit code is `2`
  - stdout is empty
  - stderr contains `invalid option "--nosuchoption"`
- after `gpg --batch --decrypt does-not-exist.gpg`:
  - exit code is `2`
  - stderr contains `does-not-exist.gpg`
- after `gpg --batch --auto-key-locate clear --encrypt --recipient nobody@example.com --output msg.gpg msg.txt`:
  - exit code is `2`
  - stderr contains `nobody@example.com`
  - file `msg.gpg` does not exist
