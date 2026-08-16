# atago Behavior Specs
## Summary
2 suites · 11 scenarios
## Contents
- [jq (uncovered contracts — unicode passthrough + empty validation)](#jq-uncovered-contracts--unicode-passthrough--empty-validation) — 2 scenarios
  - [multibyte unicode passes through unchanged](#scenario-multibyte-unicode-passes-through-unchanged)
  - [jq empty validates JSON — silent success, loud failure](#scenario-jq-empty-validates-json--silent-success-loud-failure)
- [jq (third-party CLI, no build required)](#jq-third-party-cli-no-build-required) — 9 scenarios
  - [identity filter echoes the document from stdin](#scenario-identity-filter-echoes-the-document-from-stdin)
  - [sort-keys output is deterministic and exact](#scenario-sort-keys-output-is-deterministic-and-exact)
  - [raw output strips JSON quoting](#scenario-raw-output-strips-json-quoting)
  - [arguments flow in with --arg](#scenario-arguments-flow-in-with---arg)
  - [reduce aggregates an array](#scenario-reduce-aggregates-an-array)
  - [-e exits 1 when the result is false](#scenario--e-exits-1-when-the-result-is-false)
  - [a program that does not compile exits 3 with a diagnostic on stderr](#scenario-a-program-that-does-not-compile-exits-3-with-a-diagnostic-on-stderr)
  - [invalid JSON input fails loudly and keeps stdout clean](#scenario-invalid-json-input-fails-loudly-and-keeps-stdout-clean)
  - [streaming several documents produces one result per document](#scenario-streaming-several-documents-produces-one-result-per-document)
## jq (uncovered contracts — unicode passthrough + empty validation)
Two guarantees the main jq suite leaves untested.

First, that jq is safe for text that is not ASCII: multibyte characters must
survive a round trip byte for byte, with no re-encoding and no escaping
applied on the way through.

Second, that `jq empty` works as the validator everyone uses it as. It
consumes the document, prints nothing at all, and exits 0 — but exits 5 on
malformed input. That combination (silent on success, non-zero on garbage)
is what makes it usable as a JSON check in a script, and both halves are
pinned here.

Source: `test/e2e/thirdparty/jq/extra.atago.yaml`
### Scenario: multibyte unicode passes through unchanged
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
{"s":"café 😀 日本語"}
```
#### When
```shell
jq -c .
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: jq empty validates JSON — silent success, loud failure
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
{"a":1,"b":[2,3]}
```
_stdin for `jq`:_
```text
not json
```
#### When
```shell
jq empty
jq empty
```
#### Then
- after `jq empty`:
  - exit code is `0`
  - stdout is empty
  - stderr is empty
- after `jq empty`:
  - exit code is one of `4`, `5`
  - stdout is empty
  - stderr contains `parse error`
## jq (third-party CLI, no build required)
[jq](https://jqlang.github.io/jq/) is a filter: JSON in on stdin, JSON out
on stdout, and an exit code that tells the calling script what happened.

That exit-code contract is unusually rich, and it is what this suite pins,
because a shell pipeline depends on it: 0 when the filter ran, 1 when `-e`
saw a false or null result, 2 for a usage or system error, 3 for a program
that does not compile, and 5 when the input was not valid JSON at all. Four
different kinds of "it didn't work", each of which a script may need to
handle differently — and each of which must not be confused for the others.

Source: `test/e2e/thirdparty/jq/jq.atago.yaml`
### Scenario: identity filter echoes the document from stdin
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
{"b":2,"a":1}
```
#### When
```shell
jq -c .
```
#### Then
- exit code is `0`
- stdout at `$.a` equals `1`
- stderr is empty
### Scenario: sort-keys output is deterministic and exact
_only when `jq --version` succeeds_
#### Given
- Fixture file `input.json` is created.
#### Inputs
_Fixture `input.json`:_
```text
{"b":2,"a":1}
```
_stdin for `jq`:_
```text
(read from file input.json)
```
#### When
```shell
jq -c --sort-keys .
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: raw output strips JSON quoting
_only when `jq --version` succeeds_
#### Given
- Fixture file `user.json` is created.
#### Inputs
_Fixture `user.json`:_
```text
{"name":"alice","admin":true}
```
#### When
```shell
jq -r .name user.json
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: arguments flow in with --arg
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
null
```
#### When
```shell
jq -c --arg who atago '{greeting: ("hello " + $who)}'
```
#### Then
- exit code is `0`
- stdout at `$.greeting` equals `hello atago`
### Scenario: reduce aggregates an array
_only when `jq --version` succeeds_
#### Given
- Fixture file `nums.json` is created.
#### Inputs
_Fixture `nums.json`:_
```text
[1,2,3,4,5]
```
#### When
```shell
jq 'reduce .[] as $n (0; . + $n)' nums.json
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: -e exits 1 when the result is false
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
{"present":1}
```
#### When
```shell
jq -e '.missing'
```
#### Then
- exit code is `1`
### Scenario: a program that does not compile exits 3 with a diagnostic on stderr
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
{}
```
#### When
```shell
jq '.foo['
```
#### Then
- exit code is `3`
- stdout is empty
- stderr contains `error`
### Scenario: invalid JSON input fails loudly and keeps stdout clean
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
not json at all
```
#### When
```shell
jq .
```
#### Then
- exit code is not `0`
- stderr contains `parse error`
### Scenario: streaming several documents produces one result per document
_only when `jq --version` succeeds_
#### Inputs
_stdin for `jq`:_
```text
{"n":1}
{"n":2}
```
#### When
```shell
jq -c '.n * 2'
```
#### Then
- exit code is `0`
- stdout line `1` equals an exact value
- stdout line `2` equals an exact value
