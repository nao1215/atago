# atago Behavior Specs
## Summary
5 suites · 45 scenarios
## Contents
- [jose CLI surface](#jose-cli-surface) — 10 scenarios
  - [root help prints usage with no arguments](#scenario-root-help-prints-usage-with-no-arguments)
  - [root help prints usage with --help](#scenario-root-help-prints-usage-with---help)
  - [version prints the version](#scenario-version-prints-the-version)
  - [unknown command fails and reports the unknown command](#scenario-unknown-command-fails-and-reports-the-unknown-command)
  - [jwa lists the key types](#scenario-jwa-lists-the-key-types)
  - [jwa lists only key types jose can generate](#scenario-jwa-lists-only-key-types-jose-can-generate)
  - [jwa lists only signature algorithms jose accepts](#scenario-jwa-lists-only-signature-algorithms-jose-accepts)
  - [jwa lists only elliptic curves jose can generate](#scenario-jwa-lists-only-elliptic-curves-jose-can-generate)
  - [jwa lists only key encryption algorithms jose accepts](#scenario-jwa-lists-only-key-encryption-algorithms-jose-accepts)
  - [jwa fails when no option is given](#scenario-jwa-fails-when-no-option-is-given)
- [jose completion](#jose-completion) — 6 scenarios
  - [writes a bash completion script to stdout](#scenario-writes-a-bash-completion-script-to-stdout)
  - [writes a zsh completion script to stdout](#scenario-writes-a-zsh-completion-script-to-stdout)
  - [writes a fish completion script to stdout](#scenario-writes-a-fish-completion-script-to-stdout)
  - [requires a shell argument](#scenario-requires-a-shell-argument)
  - [rejects an unknown shell](#scenario-rejects-an-unknown-shell)
  - [does not create files in the working directory](#scenario-does-not-create-files-in-the-working-directory)
- [jose jwe](#jose-jwe) — 5 scenarios
  - [round-trips a payload through encrypt and decrypt](#scenario-round-trips-a-payload-through-encrypt-and-decrypt)
  - [round-trips a payload piped through stdin](#scenario-round-trips-a-payload-piped-through-stdin)
  - [round-trips with compression enabled](#scenario-round-trips-with-compression-enabled)
  - [fails to encrypt without a key](#scenario-fails-to-encrypt-without-a-key)
  - [fails to encrypt with an invalid content encryption](#scenario-fails-to-encrypt-with-an-invalid-content-encryption)
- [jose jwk generate](#jose-jwk-generate) — 11 scenarios
  - [generates an RSA key as JSON](#scenario-generates-an-rsa-key-as-json)
  - [generates an EC key in PEM format](#scenario-generates-an-ec-key-in-pem-format)
  - [generates an OKP Ed25519 key](#scenario-generates-an-okp-ed25519-key)
  - [generates an oct key](#scenario-generates-an-oct-key)
  - [emits a public key without private fields](#scenario-emits-a-public-key-without-private-fields)
  - [rejects the unsupported OKP X448 curve](#scenario-rejects-the-unsupported-okp-x448-curve)
  - [rejects the unsupported OKP Ed448 curve](#scenario-rejects-the-unsupported-okp-ed448-curve)
  - [requires a curve for EC keys](#scenario-requires-a-curve-for-ec-keys)
  - [rejects PEM output for oct keys](#scenario-rejects-pem-output-for-oct-keys)
  - [rejects --public-key for oct keys](#scenario-rejects---public-key-for-oct-keys)
  - [leaves a parseable EC key after overwriting a longer RSA key](#scenario-leaves-a-parseable-ec-key-after-overwriting-a-longer-rsa-key)
- [jose jws](#jose-jws) — 13 scenarios
  - [sign requires an algorithm](#scenario-sign-requires-an-algorithm)
  - [sign signs a payload into a compact JWS](#scenario-sign-signs-a-payload-into-a-compact-jws)
  - [sign signs a payload read from stdin via a pipe](#scenario-sign-signs-a-payload-read-from-stdin-via-a-pipe)
  - [verify verifies and prints the payload](#scenario-verify-verifies-and-prints-the-payload)
  - [verify verifies a token passed directly as an argument](#scenario-verify-verifies-a-token-passed-directly-as-an-argument)
  - [verify verifies a token read from stdin via a pipe](#scenario-verify-verifies-a-token-read-from-stdin-via-a-pipe)
  - [verify reports a missing file instead of a parse error](#scenario-verify-reports-a-missing-file-instead-of-a-parse-error)
  - [verify fails with the wrong key](#scenario-verify-fails-with-the-wrong-key)
  - [parse prints the payload from a file](#scenario-parse-prints-the-payload-from-a-file)
  - [parse prints the payload from an inline token argument](#scenario-parse-prints-the-payload-from-an-inline-token-argument)
  - [parse prints the payload from stdin](#scenario-parse-prints-the-payload-from-stdin)
  - [parse prints all parts with --all](#scenario-parse-prints-all-parts-with---all)
  - [parse reports a missing file instead of a parse error](#scenario-parse-reports-a-missing-file-instead-of-a-parse-error)
## jose CLI surface
Source: `test/e2e/tools/jose/cli.atago.yaml`
### Scenario: root help prints usage with no arguments
#### When
```shell
jose
```
#### Then
- exit code is `0`
- stdout contains `JSON Object Signing and Encryption`, `Available Commands:`, `jwk`, `jws`, `jwe`
### Scenario: root help prints usage with --help
#### When
```shell
jose --help
```
#### Then
- exit code is `0`
- stdout contains `Usage:`, `jwa`
### Scenario: version prints the version
#### When
```shell
jose version
```
#### Then
- exit code is `0`
- stdout contains `jose version`, `MIT LICENSE`
### Scenario: unknown command fails and reports the unknown command
#### When
```shell
jose frobnicate
```
#### Then
- exit code is not `0`
- stderr contains `unknown command`
### Scenario: jwa lists the key types
#### When
```shell
jose jwa --key-type
```
#### Then
- exit code is `0`
- stdout contains `RSA`, `oct`
### Scenario: jwa lists only key types jose can generate
#### When
```shell
jose jwa --key-type
```
#### Then
- exit code is `0`
- stdout does not contain `AKP`
### Scenario: jwa lists only signature algorithms jose accepts
#### When
```shell
jose jwa --signature
```
#### Then
- exit code is `0`
- stdout contains `EdDSA`
- stdout does not contain `Ed25519`, `none`
### Scenario: jwa lists only elliptic curves jose can generate
#### When
```shell
jose jwa --elliptic-curve
```
#### Then
- exit code is `0`
- stdout contains `X25519`
- stdout does not contain `X448`
### Scenario: jwa lists only key encryption algorithms jose accepts
#### When
```shell
jose jwa --key-encryption
```
#### Then
- exit code is `0`
- stdout contains `RSA-OAEP`
- stdout does not contain `RSA-OAEP-384`, `HPKE`
### Scenario: jwa fails when no option is given
#### When
```shell
jose jwa
```
#### Then
- exit code is not `0`
- stderr is not empty
## jose completion
Source: `test/e2e/tools/jose/completion.atago.yaml`
### Scenario: writes a bash completion script to stdout
#### When
```shell
jose completion bash
```
#### Then
- exit code is `0`
- stdout contains `bash completion`
### Scenario: writes a zsh completion script to stdout
#### When
```shell
jose completion zsh
```
#### Then
- exit code is `0`
- stdout contains `compdef`
### Scenario: writes a fish completion script to stdout
#### When
```shell
jose completion fish
```
#### Then
- exit code is `0`
- stdout contains `fish`
### Scenario: requires a shell argument
#### When
```shell
jose completion
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: rejects an unknown shell
#### When
```shell
jose completion powershell
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: does not create files in the working directory
#### When
```shell
jose completion zsh > /dev/null
ls -A

```
#### Then
- exit code is `0`
- stdout is empty
## jose jwe
Source: `test/e2e/tools/jose/jwe.atago.yaml`
### Scenario: round-trips a payload through encrypt and decrypt
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null

jose jwe encrypt --key ec.jwk --key-encryption ECDH-ES \
  --content-encryption A256GCM payload.json > secret.jwe
jose jwe decrypt --key ec.jwk secret.jwe

```
#### Then
- after `jose jwe encrypt --key ec.jwk --key-encryption ECDH-ES \
  --content-encryption A256GCM payload.json > secret.jwe
jose jwe decrypt --key ec.jwk secret.jwe
`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: round-trips a payload piped through stdin
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null

cat payload.json | jose jwe encrypt --key ec.jwk \
  --key-encryption ECDH-ES --content-encryption A256GCM > secret.jwe
cat secret.jwe | jose jwe decrypt --key ec.jwk

```
#### Then
- after `cat payload.json | jose jwe encrypt --key ec.jwk \
  --key-encryption ECDH-ES --content-encryption A256GCM > secret.jwe
cat secret.jwe | jose jwe decrypt --key ec.jwk
`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: round-trips with compression enabled
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null

jose jwe encrypt --key ec.jwk --key-encryption ECDH-ES \
  --content-encryption A256GCM --compress payload.json > secret.jwe
jose jwe decrypt --key ec.jwk secret.jwe

```
#### Then
- after `jose jwe encrypt --key ec.jwk --key-encryption ECDH-ES \
  --content-encryption A256GCM --compress payload.json > secret.jwe
jose jwe decrypt --key ec.jwk secret.jwe
`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: fails to encrypt without a key
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null

jose jwe encrypt --key-encryption ECDH-ES --content-encryption A256GCM payload.json
```
#### Then
- after `jose jwe encrypt --key-encryption ECDH-ES --content-encryption A256GCM payload.json`:
  - exit code is not `0`
  - stderr contains `key file required`
### Scenario: fails to encrypt with an invalid content encryption
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null

jose jwe encrypt --key ec.jwk --key-encryption ECDH-ES --content-encryption BOGUS payload.json
```
#### Then
- after `jose jwe encrypt --key ec.jwk --key-encryption ECDH-ES --content-encryption BOGUS payload.json`:
  - exit code is not `0`
  - stderr contains `content encryption`
## jose jwk generate
Source: `test/e2e/tools/jose/jwk.atago.yaml`
### Scenario: generates an RSA key as JSON
#### When
```shell
jose jwk generate --type RSA --size 2048
```
#### Then
- exit code is `0`
- stdout contains `"kty"`, `RSA`
### Scenario: generates an EC key in PEM format
#### When
```shell
jose jwk generate --type EC --curve P-256 --output-format pem
```
#### Then
- exit code is `0`
- stdout contains `BEGIN`, `PRIVATE KEY`
### Scenario: generates an OKP Ed25519 key
#### When
```shell
jose jwk generate --type OKP --curve Ed25519
```
#### Then
- exit code is `0`
- stdout contains `OKP`
### Scenario: generates an oct key
#### When
```shell
jose jwk generate --type oct --size 256
```
#### Then
- exit code is `0`
- stdout contains `oct`
### Scenario: emits a public key without private fields
#### When
```shell
jose jwk generate --type EC --curve P-256 --public-key
```
#### Then
- exit code is `0`
- stdout contains `"kty"`
- stdout does not contain `"d"`
### Scenario: rejects the unsupported OKP X448 curve
#### When
```shell
jose jwk generate --type OKP --curve X448
```
#### Then
- exit code is not `0`
- stderr contains `OKP supports`
### Scenario: rejects the unsupported OKP Ed448 curve
#### When
```shell
jose jwk generate --type OKP --curve Ed448
```
#### Then
- exit code is not `0`
- stderr contains `OKP supports`
### Scenario: requires a curve for EC keys
#### When
```shell
jose jwk generate --type EC
```
#### Then
- exit code is not `0`
- stderr contains `require --curve`
### Scenario: rejects PEM output for oct keys
#### When
```shell
jose jwk generate --type oct --size 256 --output-format pem
```
#### Then
- exit code is not `0`
- stderr contains `oct`, `json`
### Scenario: rejects --public-key for oct keys
#### When
```shell
jose jwk generate --type oct --size 256 --public-key
```
#### Then
- exit code is not `0`
- stderr contains `public key`
### Scenario: leaves a parseable EC key after overwriting a longer RSA key
#### When
```shell
jose jwk generate --type RSA --size 4096 --output key.jwk
jose jwk generate --type EC --curve P-256 --output key.jwk
printf 'hello' > msg.txt
jose jws sign --algorithm ES256 --key key.jwk msg.txt

```
#### Then
- exit code is `0`
- stdout is not empty
## jose jws
Source: `test/e2e/tools/jose/jws.atago.yaml`
### Scenario: sign requires an algorithm
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws sign --key ec.jwk payload.json
```
#### Then
- after `jose jws sign --key ec.jwk payload.json`:
  - exit code is not `0`
  - stderr contains `signature algorithm`
### Scenario: sign signs a payload into a compact JWS
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws sign --algorithm ES256 --key ec.jwk payload.json
```
#### Then
- after `jose jws sign --algorithm ES256 --key ec.jwk payload.json`:
  - exit code is `0`
  - stdout contains `.`
### Scenario: sign signs a payload read from stdin via a pipe
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

cat payload.json | jose jws sign --algorithm ES256 --key ec.jwk
```
#### Then
- after `cat payload.json | jose jws sign --algorithm ES256 --key ec.jwk`:
  - exit code is `0`
  - stdout contains `.`
### Scenario: verify verifies and prints the payload
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws verify --algorithm ES256 --key ec.jwk token.jws
```
#### Then
- after `jose jws verify --algorithm ES256 --key ec.jwk token.jws`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: verify verifies a token passed directly as an argument
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws verify --algorithm ES256 --key ec.jwk "$(cat token.jws)"
```
#### Then
- after `jose jws verify --algorithm ES256 --key ec.jwk "$(cat token.jws)"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: verify verifies a token read from stdin via a pipe
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

cat token.jws | jose jws verify --algorithm ES256 --key ec.jwk
```
#### Then
- after `cat token.jws | jose jws verify --algorithm ES256 --key ec.jwk`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: verify reports a missing file instead of a parse error
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws verify --algorithm ES256 --key ec.jwk does-not-exist.jws
```
#### Then
- after `jose jws verify --algorithm ES256 --key ec.jwk does-not-exist.jws`:
  - exit code is not `0`
  - stderr contains `failed to open file`
### Scenario: verify fails with the wrong key
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jwk generate --type EC --curve P-256 --output other.jwk
jose jws verify --algorithm ES256 --key other.jwk token.jws
```
#### Then
- after `jose jws verify --algorithm ES256 --key other.jwk token.jws`:
  - exit code is not `0`
  - stderr contains `verify`
### Scenario: parse prints the payload from a file
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws parse token.jws
```
#### Then
- after `jose jws parse token.jws`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: parse prints the payload from an inline token argument
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws parse "$(cat token.jws)"
```
#### Then
- after `jose jws parse "$(cat token.jws)"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: parse prints the payload from stdin
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

cat token.jws | jose jws parse -
```
#### Then
- after `cat token.jws | jose jws parse -`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: parse prints all parts with --all
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws parse --all token.jws
```
#### Then
- after `jose jws parse --all token.jws`:
  - exit code is `0`
  - stdout contains `Payload:`, `Signature 0:`
### Scenario: parse reports a missing file instead of a parse error
#### When
```shell
printf '{"sub":"alice"}' > payload.json
jose jwk generate --type EC --curve P-256 --output ec.jwk >/dev/null
jose jws sign --algorithm ES256 --key ec.jwk payload.json > token.jws

jose jws parse does-not-exist.jws
```
#### Then
- after `jose jws parse does-not-exist.jws`:
  - exit code is not `0`
  - stderr contains `failed to open file`
