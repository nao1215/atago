# atago Behavior Specs
## Summary
4 suites · 50 scenarios
## Contents
- [fx (side effects on disk)](#fx-side-effects-on-disk) — 7 scenarios
  - [save rewrites the file it was given and touches nothing else](#scenario-save-rewrites-the-file-it-was-given-and-touches-nothing-else)
  - [running the same normalization twice changes nothing the second time](#scenario-running-the-same-normalization-twice-changes-nothing-the-second-time)
  - [a reducer that only reads leaves the workdir alone](#scenario-a-reducer-that-only-reads-leaves-the-workdir-alone)
  - [save without a file argument fails and writes nothing](#scenario-save-without-a-file-argument-fails-and-writes-nothing)
  - [save refuses to write through a symlink and leaves the target intact](#scenario-save-refuses-to-write-through-a-symlink-and-leaves-the-target-intact)
  - [a .fxrc.js in the working directory adds a function reducers can call](#scenario-a-fxrcjs-in-the-working-directory-adds-a-function-reducers-can-call)
  - [a .fxrc.js in the home directory is loaded too, and fx writes nothing there](#scenario-a-fxrcjs-in-the-home-directory-is-loaded-too-and-fx-writes-nothing-there)
- [fx (input formats, strictness, round trips)](#fx-input-formats-strictness-round-trips) — 19 scenarios
  - [--yaml parses YAML into the JSON model](#scenario---yaml-parses-yaml-into-the-json-model)
  - [--toml parses TOML into the JSON model](#scenario---toml-parses-toml-into-the-json-model)
  - [a .yaml file argument switches parsers with no flag at all](#scenario-a-yaml-file-argument-switches-parsers-with-no-flag-at-all)
  - [a .toml file argument does the same](#scenario-a-toml-file-argument-does-the-same)
  - [extension detection ignores case](#scenario-extension-detection-ignores-case)
  - [an unknown extension stays on the JSON parser](#scenario-an-unknown-extension-stays-on-the-json-parser)
  - [--raw treats each input line as a string](#scenario---raw-treats-each-input-line-as-a-string)
  - [-s collects a stream of documents into one array](#scenario--s-collects-a-stream-of-documents-into-one-array)
  - [-rs combines raw lines into an array of strings](#scenario--rs-combines-raw-lines-into-an-array-of-strings)
  - [the default parser accepts comments and a trailing comma](#scenario-the-default-parser-accepts-comments-and-a-trailing-comma)
  - [--strict rejects a comment](#scenario---strict-rejects-a-comment)
  - [--strict rejects a trailing comma](#scenario---strict-rejects-a-trailing-comma)
  - [--raw and --yaml together are refused](#scenario---raw-and---yaml-together-are-refused)
  - [--yaml and --toml together are refused](#scenario---yaml-and---toml-together-are-refused)
  - [base64 encoding round-trips through a second fx](#scenario-base64-encoding-round-trips-through-a-second-fx)
  - [a document survives a JSON to YAML to JSON round trip byte for byte](#scenario-a-document-survives-a-json-to-yaml-to-json-round-trip-byte-for-byte)
  - [--no-inline expands the containers the default output packs onto one line](#scenario---no-inline-expands-the-containers-the-default-output-packs-onto-one-line)
  - [malformed YAML fails with a diagnostic that points at the line](#scenario-malformed-yaml-fails-with-a-diagnostic-that-points-at-the-line)
  - [malformed TOML fails without writing to stdout](#scenario-malformed-toml-fails-without-writing-to-stdout)
- [fx (third-party CLI, JSON reducer contract)](#fx-third-party-cli-json-reducer-contract) — 16 scenarios
  - [version prints a semantic version](#scenario-version-prints-a-semantic-version)
  - [the identity reducer pretty-prints the document and keeps key order](#scenario-the-identity-reducer-pretty-prints-the-document-and-keeps-key-order)
  - [a dotted path selects a nested value and prints strings unquoted](#scenario-a-dotted-path-selects-a-nested-value-and-prints-strings-unquoted)
  - [reducers compose left to right](#scenario-reducers-compose-left-to-right)
  - [an arrow function transforms the document](#scenario-an-arrow-function-transforms-the-document)
  - [a missing key reports undefined on stderr and still exits 0](#scenario-a-missing-key-reports-undefined-on-stderr-and-still-exits-0)
  - [dereferencing through a missing key is a hard error](#scenario-dereferencing-through-a-missing-key-is-a-hard-error)
  - [exit() hands the reducer's own code to the shell](#scenario-exit-hands-the-reducers-own-code-to-the-shell)
  - [invalid JSON fails loudly and keeps stdout clean](#scenario-invalid-json-fails-loudly-and-keeps-stdout-clean)
  - [a missing input file fails before anything is printed](#scenario-a-missing-input-file-fails-before-anything-is-printed)
  - [empty input succeeds with no output at all](#scenario-empty-input-succeeds-with-no-output-at-all)
  - [each document in a stream produces its own result](#scenario-each-document-in-a-stream-produces-its-own-result)
  - [filter drops documents from a stream instead of emitting them](#scenario-filter-drops-documents-from-a-stream-instead-of-emitting-them)
  - [the identity path keeps number literals JavaScript would rewrite](#scenario-the-identity-path-keeps-number-literals-javascript-would-rewrite)
  - [piped output carries no terminal escapes](#scenario-piped-output-carries-no-terminal-escapes)
  - [a directory given as input fails without writing to stdout](#scenario-a-directory-given-as-input-fails-without-writing-to-stdout)
- [fx (interactive viewer, pty testbed)](#fx-interactive-viewer-pty-testbed) — 8 scenarios
  - [the viewer renders the document and quits cleanly on q](#scenario-the-viewer-renders-the-document-and-quits-cleanly-on-q)
  - [esc leaves the viewer with a success code](#scenario-esc-leaves-the-viewer-with-a-success-code)
  - [the interface goes to stderr, so a redirected stdout stays empty](#scenario-the-interface-goes-to-stderr-so-a-redirected-stdout-stays-empty)
  - [P prints the value under the cursor to stdout](#scenario-p-prints-the-value-under-the-cursor-to-stdout)
  - [collapse-all folds the containers and expand-all brings them back](#scenario-collapse-all-folds-the-containers-and-expand-all-brings-them-back)
  - [search moves the cursor to the match and the status bar shows its path](#scenario-search-moves-the-cursor-to-the-match-and-the-status-bar-shows-its-path)
  - [the rendered frame shows the tree and the status bar together](#scenario-the-rendered-frame-shows-the-tree-and-the-status-bar-together)
  - [with a terminal and no file, fx prints its usage instead of opening the viewer](#scenario-with-a-terminal-and-no-file-fx-prints-its-usage-instead-of-opening-the-viewer)

## fx (side effects on disk)
A JSON viewer sounds like a read-only program, and mostly it is — which is
exactly why the places where fx writes deserve an exhaustive delta rather
than a per-file check.

Two of them exist. `save` rewrites the file fx was given, in place, and
refuses to do so when there is no file or when the path is a symlink. And
before any reducer runs, fx loads `.fxrc.js` from the working directory and
from the user's home, so a stray file in either place changes what a
reducer means. `changes:` states the whole picture in one assertion —
what was written and, just as importantly, that nothing else was — and
`sandbox_home` puts the home directory inside the workdir so a write there
would show up in the same delta.

Source: `test/e2e/thirdparty/fx/changes.atago.yaml`
### Scenario: save rewrites the file it was given and touches nothing else
_only when `fx --version` succeeds_
#### Given
- Fixture file `config.json` is created.
- Fixture file `untouched.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config.json`:_
```text
{"b":2,"a":1}
```
_Fixture `untouched.json`:_
```text
{"keep":true}
```
#### When
```shell
fx config.json sortKeys save
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified `config.json`, deleted nothing
- file `config.json` equals exact bytes

### Scenario: running the same normalization twice changes nothing the second time
_only when `fx --version` succeeds_
#### Given
- Fixture file `config.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config.json`:_
```text
{"b":2,"a":1}
```
#### When
```shell
fx config.json sortKeys save
fx config.json sortKeys save
```
#### Then
- after `fx config.json sortKeys save`:
  - exit code is `0`
  - the step changed exactly modified `config.json`
- after `fx config.json sortKeys save`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: a reducer that only reads leaves the workdir alone
_only when `fx --version` succeeds_
#### Given
- Fixture file `config.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config.json`:_
```text
{"a":1}
```
#### When
```shell
fx config.json .a
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: save without a file argument fails and writes nothing
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"a":1}
```
#### When
```shell
fx save
```
#### Then
- exit code is `1`
- stderr contains `specify a file as the first argument`
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: save refuses to write through a symlink and leaves the target intact
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `real.json` is created.
- Fixture file `link.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `real.json`:_
```text
{"b":2,"a":1}
```
#### When
```shell
fx link.json sortKeys save
```
#### Then
- exit code is `1`
- stderr contains `cannot save to a symbolic link`
- the step changed exactly created nothing, modified nothing, deleted nothing
- file `real.json` equals exact bytes

### Scenario: a .fxrc.js in the working directory adds a function reducers can call
_only when `fx --version` succeeds_
#### Given
- Fixture file `.fxrc.js` is created.
- Fixture file `data.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.fxrc.js`:_
```text
function double(x) {
  return x * 2
}
```
_Fixture `data.json`:_
```text
{"n":21}
```
#### When
```shell
fx data.json .n double
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: a .fxrc.js in the home directory is loaded too, and fx writes nothing there
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `.atago-home/.fxrc.js` is created.
- Fixture file `data.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.atago-home/.fxrc.js`:_
```text
function shout(x) {
  return String(x).toUpperCase()
}
```
_Fixture `data.json`:_
```text
{"word":"atago"}
```
#### When
```shell
fx data.json .word shout
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing

## fx (input formats, strictness, round trips)
fx accepts more than JSON, and how it decides what it is holding is a
contract of its own: `--yaml` and `--toml` switch parsers explicitly, but a
file argument ending in .yaml, .yml or .toml switches them implicitly, and
the two flags are mutually exclusive with each other and with `--raw`.

Strictness is the other half. By default fx reads JSON the way a human
writes it — comments and trailing commas included — and `--strict` turns
exactly those tolerances back into errors. The round-trip scenarios use fx
twice in a pipeline so the assertion is about a relationship (encode then
decode returns the original bytes) rather than about one recorded output.

Source: `test/e2e/thirdparty/fx/formats.atago.yaml`
### Scenario: --yaml parses YAML into the JSON model
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
service:
  name: api
  port: 8080
```
#### When
```shell
fx --yaml .service.port
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: --toml parses TOML into the JSON model
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
[package]
name = "atago"
version = "1.2.3"
```
#### When
```shell
fx --toml .package.name
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: a .yaml file argument switches parsers with no flag at all
_only when `fx --version` succeeds_
#### Given
- Fixture file `config.yaml` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config.yaml`:_
```text
service:
  name: api
  replicas: 3
```
#### When
```shell
fx config.yaml .service.replicas
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: a .toml file argument does the same
_only when `fx --version` succeeds_
#### Given
- Fixture file `config.toml` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `config.toml`:_
```text
[package]
name = "atago"
```
#### When
```shell
fx config.toml .package.name
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: extension detection ignores case
_only when `fx --version` succeeds_
#### Given
- Fixture file `CONFIG.YAML` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `CONFIG.YAML`:_
```text
service:
  name: api
```
#### When
```shell
fx CONFIG.YAML .service.name
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: an unknown extension stays on the JSON parser
_only when `fx --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `payload.txt`:_
```text
{"a":1}
```
#### When
```shell
fx payload.txt .a
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: --raw treats each input line as a string
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
hello
hi
```
#### When
```shell
fx -r len
```
#### Then
- exit code is `0`
- stdout line `1` equals an exact value
- stdout line `2` equals an exact value

### Scenario: -s collects a stream of documents into one array
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"n":1}
{"n":2}
{"n":3}
```
#### When
```shell
fx -s len
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: -rs combines raw lines into an array of strings
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
alpha
beta
```
#### When
```shell
fx -rs .
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
[
  "alpha",
  "beta"
]
```
### Scenario: the default parser accepts comments and a trailing comma
_only when `fx --version` succeeds_
#### Given
- Fixture file `loose.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `loose.json`:_
```text
{
  // the service this config belongs to
  "name": "api",
  "port": 8080,
}
```
#### When
```shell
fx loose.json .port
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: --strict rejects a comment
_only when `fx --version` succeeds_
#### Given
- Fixture file `loose.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `loose.json`:_
```text
{
  // the service this config belongs to
  "name": "api"
}
```
#### When
```shell
fx --strict loose.json .name
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Comments are not allowed in strict mode`

### Scenario: --strict rejects a trailing comma
_only when `fx --version` succeeds_
#### Given
- Fixture file `trailing.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `trailing.json`:_
```text
{
  "name": "api",
}
```
#### When
```shell
fx --strict trailing.json .name
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Trailing comma is not allowed in strict mode`

### Scenario: --raw and --yaml together are refused
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
a: 1
```
#### When
```shell
fx --yaml --raw .
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `can't use --yaml/--toml and --raw flags together`

### Scenario: --yaml and --toml together are refused
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
a: 1
```
#### When
```shell
fx --yaml --toml .
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `can't use both --yaml and --toml flags together`

### Scenario: base64 encoding round-trips through a second fx
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `secret.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `secret.json`:_
```text
{"token":"s3cr3t value"}
```
#### When
```shell
fx secret.json .token toBase64 | fx -r fromBase64
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: a document survives a JSON to YAML to JSON round trip byte for byte
_only when `fx --version` succeeds_
#### Given
- Fixture file `source.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `source.json`:_
```text
{"name":"atago","nums":[1,2,3],"nested":{"leaf":true}}
```
#### When
```shell
fx source.json .
fx source.json YAML.stringify
fx round_trip.yaml .
```
#### Then
- after `fx source.json .`:
  - exit code is `0`
- after `fx source.json YAML.stringify`:
  - exit code is `0`
- after `fx round_trip.yaml .`:
  - exit code is `0`
  - file `after.txt` is byte-identical to `before.txt`

#### Generated artifacts
- `before.txt`
- `round_trip.yaml`
- `after.txt`

### Scenario: --no-inline expands the containers the default output packs onto one line
_only when `fx --version` succeeds_
#### Given
- Fixture file `doc.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `doc.json`:_
```text
{"nums":[1,2],"nested":{"leaf":1}}
```
#### When
```shell
fx doc.json .
fx --no-inline doc.json .
```
#### Then
- after `fx doc.json .`:
  - exit code is `0`
  - stdout contains `"nums": [ 1, 2 ]`
- after `fx --no-inline doc.json .`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
{
  "nums": [
    1,
    2
  ],
  "nested": {
    "leaf": 1
  }
}
```
### Scenario: malformed YAML fails with a diagnostic that points at the line
_only when `fx --version` succeeds_
#### Given
- Fixture file `broken.yaml` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `broken.yaml`:_
```text
service:
 name: api
  port: 8080
```
#### When
```shell
fx broken.yaml .
```
#### Then
- exit code is `1`
- stdout contains `mapping value is not allowed`
- stderr is empty

### Scenario: malformed TOML fails without writing to stdout
_only when `fx --version` succeeds_
#### Given
- Fixture file `broken.toml` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `broken.toml`:_
```text
a = 
```
#### When
```shell
fx broken.toml .
```
#### Then
- exit code is not `0`
- stdout is empty

## fx (third-party CLI, JSON reducer contract)
[fx](https://fx.wtf/) is two programs wearing one name. Hand it a reducer
and it behaves like a filter — JSON in, a value out, an exit code for the
shell. Hand it nothing and it opens a full-screen viewer instead. This file
pins the filter half; the viewer lives in tui.atago.yaml and the parser
flags in formats.atago.yaml.

The filter contract is worth pinning precisely because it is not the one a
jq user expects. A reducer is JavaScript, not a query language, so an
absent key is `undefined` rather than an error: fx reports it on stderr and
still exits 0, while dereferencing THROUGH that undefined is a hard exit 1.
Strings print unquoted, `exit()` lets the reducer choose the process's exit
code, and the identity path hands back the original number literals instead
of what a float64 round trip would leave.

Source: `test/e2e/thirdparty/fx/fx.atago.yaml`
### Scenario: version prints a semantic version
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
fx --version
```
#### Then
- exit code is `0`
- stdout matches `/^[0-9]+\.[0-9]+\.[0-9]+/`
- stderr is empty

### Scenario: the identity reducer pretty-prints the document and keeps key order
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"b":2,"a":1}
```
#### When
```shell
fx .
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stderr is empty

#### Expected output
_expected stdout:_
```text
{
  "b": 2,
  "a": 1
}
```
### Scenario: a dotted path selects a nested value and prints strings unquoted
_only when `fx --version` succeeds_
#### Given
- Fixture file `user.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `user.json`:_
```text
{"profile":{"name":"alice","admin":true}}
```
#### When
```shell
fx user.json .profile.name
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: reducers compose left to right
_only when `fx --version` succeeds_
#### Given
- Fixture file `items.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `items.json`:_
```text
{"items":["a","b","c"]}
```
#### When
```shell
fx items.json .items len
```
#### Then
- exit code is `0`
- stdout equals an exact value

### Scenario: an arrow function transforms the document
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"nums":[1,2,3]}
```
#### When
```shell
fx '.nums.map(n => n * 2)'
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
[
  2,
  4,
  6
]
```
### Scenario: a missing key reports undefined on stderr and still exits 0
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"a":1}
```
#### When
```shell
fx .nope
```
#### Then
- exit code is `0`
- stdout is empty
- stderr equals an exact value

### Scenario: dereferencing through a missing key is a hard error
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"a":1}
```
#### When
```shell
fx .a.b.c
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `TypeError`

### Scenario: exit() hands the reducer's own code to the shell
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{}
```
_stdin for `fx`:_
```text
{}
```
#### When
```shell
fx 'exit(3)'
fx 'exit(0)'
```
#### Then
- after `fx 'exit(3)'`:
  - exit code is `3`
- after `fx 'exit(0)'`:
  - exit code is `0`

### Scenario: invalid JSON fails loudly and keeps stdout clean
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"bad"
```
#### When
```shell
fx .
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Expected colon after object key`

### Scenario: a missing input file fails before anything is printed
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
fx no-such.json .
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `no-such.json`

### Scenario: empty input succeeds with no output at all
_only when `fx --version` succeeds_
#### Given
- Fixture file `empty.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
(read from file empty.json)
```
#### When
```shell
fx .
```
#### Then
- exit code is `0`
- stdout is empty
- stderr is empty

### Scenario: each document in a stream produces its own result
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"n":1}
{"n":2}
```
#### When
```shell
fx .n
```
#### Then
- exit code is `0`
- stdout line `1` equals an exact value
- stdout line `2` equals an exact value

### Scenario: filter drops documents from a stream instead of emitting them
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"n":1}
{"n":2}
{"n":3}
```
#### When
```shell
fx 'filter(x => x.n > 2)'
```
#### Then
- exit code is `0`
- stdout equals an exact value

#### Expected output
_expected stdout:_
```text
{
  "n": 3
}
```
### Scenario: the identity path keeps number literals JavaScript would rewrite
_only when `fx --version` succeeds_
#### Given
- Fixture file `numbers.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `numbers.json`:_
```text
{"exact":1.0,"big":10000000000000000001}
```
#### When
```shell
fx numbers.json .
fx numbers.json 'x => x'
```
#### Then
- after `fx numbers.json .`:
  - exit code is `0`
  - stdout contains `"exact": 1.0`, `"big": 10000000000000000001`
- after `fx numbers.json 'x => x'`:
  - exit code is `0`
  - stdout contains `"exact": 1`, `"big": 10000000000000000001`
  - stdout does not contain `"exact": 1.0`

### Scenario: piped output carries no terminal escapes
_only when `fx --version` succeeds_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_stdin for `fx`:_
```text
{"a":"text","b":42,"c":null}
```
#### When
```shell
fx .
```
#### Then
- exit code is `0`
- stdout does not match `/\x1b\[/`

### Scenario: a directory given as input fails without writing to stdout
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `subdir/keep.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `subdir/keep.txt`:_
```text
placeholder
```
#### When
```shell
fx subdir .
```
#### Then
- exit code is not `0`
- stdout is empty

## fx (interactive viewer, pty testbed)
Given a file and no reducer, fx stops being a filter and becomes a
full-screen JSON viewer: a foldable tree, a regexp search, and a cursor
whose position is reported as a JSON path in the status bar.

What makes it an unusual TUI to test is where it draws. fx paints the whole
interface on stderr and keeps stdout for the one value the user asks it to
print, so `fx data.json > picked.txt` is a real workflow: the UI still
appears on the terminal and the file receives only the selected node. The
scenarios below drive the viewer inside a pseudo-terminal and assert both
sides of that split — the rendered screen, and what the pipeline behind it
received.

The waits are `expect_screen`, not `expect`: fx negotiates with the
terminal before its first paint, so the interesting condition is a rendered
frame rather than a byte on the wire.

Source: `test/e2e/thirdparty/fx/tui.atago.yaml`
### Scenario: the viewer renders the document and quits cleanly on q
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago","tags":["cli","e2e"],"nested":{"deep":{"leaf":42}}}
```
#### When
```shell
# interactive (pty): fx viewer.json
```
#### Then
- exit code is `0`

### Scenario: esc leaves the viewer with a success code
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago"}
```
#### When
```shell
# interactive (pty): fx viewer.json
```
#### Then
- exit code is `0`

### Scenario: the interface goes to stderr, so a redirected stdout stays empty
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago","tags":["cli","e2e"]}
```
#### When
```shell
# interactive (pty): fx viewer.json > picked.txt
```
#### Then
- exit code is `0`
- file `picked.txt` equals exact bytes

### Scenario: P prints the value under the cursor to stdout
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago","tags":["cli","e2e"]}
```
#### When
```shell
# interactive (pty): fx viewer.json > picked.txt
```
#### Then
- exit code is `0`
- file `picked.txt` equals exact bytes

### Scenario: collapse-all folds the containers and expand-all brings them back
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago","tags":["cli","e2e"],"nested":{"deep":{"leaf":42}}}
```
#### When
```shell
# interactive (pty): fx viewer.json
```
#### Then
- exit code is `0`

### Scenario: search moves the cursor to the match and the status bar shows its path
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago","tags":["cli","e2e"],"nested":{"deep":{"leaf":42}}}
```
#### When
```shell
# interactive (pty): fx viewer.json
```
#### Then
- exit code is `0`

### Scenario: the rendered frame shows the tree and the status bar together
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- Fixture file `viewer.json` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `viewer.json`:_
```text
{"name":"atago","tags":["cli","e2e"],"nested":{"deep":{"leaf":42}}}
```
#### When
```shell
# interactive (pty): fx viewer.json
```
#### Then
- rendered screen line `1` contains `{`
- rendered screen contains `"leaf": 42`
- rendered screen contains `viewer.json`

### Scenario: with a terminal and no file, fx prints its usage instead of opening the viewer
_only when `fx --version` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
# interactive (pty): fx
```
#### Then
- exit code is `0`
- stdout contains `https://fx.wtf`
