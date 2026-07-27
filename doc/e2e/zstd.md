# atago Behavior Specs
## Summary
1 suite · 14 scenarios
## Contents
- [zstd (compression round trips and integrity)](#zstd-compression-round-trips-and-integrity) — 14 scenarios
  - [compressing keeps the input and adds one archive](#scenario-compressing-keeps-the-input-and-adds-one-archive)
  - [a compress and decompress round trip restores the bytes exactly](#scenario-a-compress-and-decompress-round-trip-restores-the-bytes-exactly)
  - [an empty file survives the round trip](#scenario-an-empty-file-survives-the-round-trip)
  - [NUL bytes and high bytes survive the round trip](#scenario-nul-bytes-and-high-bytes-survive-the-round-trip)
  - [the compression level changes the archive, never the content](#scenario-the-compression-level-changes-the-archive-never-the-content)
  - [--rm removes the input only after a successful compression](#scenario---rm-removes-the-input-only-after-a-successful-compression)
  - [an existing archive is not overwritten](#scenario-an-existing-archive-is-not-overwritten)
  - [-f overwrites the archive the refusal protected](#scenario--f-overwrites-the-archive-the-refusal-protected)
  - [-t accepts a real archive and rejects a corrupt one](#scenario--t-accepts-a-real-archive-and-rejects-a-corrupt-one)
  - [decompressing a corrupt archive leaves no output behind](#scenario-decompressing-a-corrupt-archive-leaves-no-output-behind)
  - [a missing input fails and names the file](#scenario-a-missing-input-fails-and-names-the-file)
  - [an unknown option fails without producing output](#scenario-an-unknown-option-fails-without-producing-output)
  - [the stdin to stdout pipeline round trips](#scenario-the-stdin-to-stdout-pipeline-round-trips)
  - [-l reports one frame and the checksum algorithm](#scenario--l-reports-one-frame-and-the-checksum-algorithm)
## zstd (compression round trips and integrity)
[zstd](https://facebook.github.io/zstd/) compresses and decompresses files.
A compressor has exactly one promise worth testing: what comes back out is
what went in, byte for byte.

Every round trip here is checked with `equals_file`, which compares the
restored bytes against the original with no newline or encoding
normalization — including the cases that break naive tooling, an empty file
and a file full of NUL bytes. The rest of the suite pins the safety
properties around that: the input is kept unless `--rm` is asked for, an
existing archive is never silently overwritten, a corrupt frame is rejected
by both `-t` and decompression, and a failed decompression leaves no partial
output behind. All inputs are written by the spec, so nothing is committed.

Source: `test/e2e/thirdparty/zstd/zstd.atago.yaml`
### Scenario: compressing keeps the input and adds one archive
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `payload.txt`:_
```text
the quick brown fox
```
#### When
```shell
zstd -q payload.txt
```
#### Then
- exit code is `0`
- the step changed exactly created `payload.txt.zst`, modified nothing, deleted nothing
### Scenario: a compress and decompress round trip restores the bytes exactly
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
line one
line two	tabbed
```
#### When
```shell
zstd -q payload.txt -o payload.zst
zstd -dq payload.zst -o restored.txt
```
#### Then
- after `zstd -q payload.txt -o payload.zst`:
  - exit code is `0`
- after `zstd -dq payload.zst -o restored.txt`:
  - exit code is `0`
  - file `restored.txt` is byte-identical to `payload.txt`
### Scenario: an empty file survives the round trip
_only when `zstd --version` succeeds_
#### Given
- Fixture file `empty.txt` is created.
#### When
```shell
zstd -q empty.txt -o empty.zst
zstd -dq empty.zst -o restored.txt
```
#### Then
- after `zstd -q empty.txt -o empty.zst`:
  - exit code is `0`
- after `zstd -dq empty.zst -o restored.txt`:
  - exit code is `0`
  - file `restored.txt` is byte-identical to `empty.txt`
### Scenario: NUL bytes and high bytes survive the round trip
_only when `zstd --version` succeeds_
#### Given
- Fixture file `binary.dat` is created.
#### When
```shell
zstd -q binary.dat -o binary.zst
zstd -dq binary.zst -o restored.dat
```
#### Then
- after `zstd -q binary.dat -o binary.zst`:
  - exit code is `0`
- after `zstd -dq binary.zst -o restored.dat`:
  - exit code is `0`
  - file `restored.dat` is byte-identical to `binary.dat`
### Scenario: the compression level changes the archive, never the content
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
repeatable payload for two levels
```
#### When
```shell
zstd -1 -q payload.txt -o fast.zst
zstd -19 -q payload.txt -o small.zst
zstd -dq fast.zst -o from-fast.txt
zstd -dq small.zst -o from-small.txt
```
#### Then
- after `zstd -1 -q payload.txt -o fast.zst`:
  - exit code is `0`
- after `zstd -19 -q payload.txt -o small.zst`:
  - exit code is `0`
- after `zstd -dq fast.zst -o from-fast.txt`:
  - exit code is `0`
  - file `from-fast.txt` is byte-identical to `payload.txt`
- after `zstd -dq small.zst -o from-small.txt`:
  - exit code is `0`
  - file `from-small.txt` is byte-identical to `payload.txt`
### Scenario: --rm removes the input only after a successful compression
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
throwaway
```
#### When
```shell
zstd -q --rm payload.txt -o payload.zst
```
#### Then
- exit code is `0`
- the step changed exactly created `payload.zst`, modified nothing, deleted `payload.txt`
### Scenario: an existing archive is not overwritten
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
- Fixture file `payload.zst` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
new data
```
_Fixture `payload.zst`:_
```text
PRECIOUS EXISTING FILE
```
#### When
```shell
zstd -q payload.txt -o payload.zst
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `already exists`
- the step changed exactly created nothing, modified nothing, deleted nothing
- file `payload.zst` contains `PRECIOUS EXISTING FILE`
### Scenario: -f overwrites the archive the refusal protected
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
- Fixture file `payload.zst` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
new data
```
_Fixture `payload.zst`:_
```text
OLD CONTENT
```
#### When
```shell
zstd -q -f payload.txt -o payload.zst
zstd -dq payload.zst -o restored.txt
```
#### Then
- after `zstd -q -f payload.txt -o payload.zst`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `payload.zst`, deleted nothing
- after `zstd -dq payload.zst -o restored.txt`:
  - exit code is `0`
  - file `restored.txt` is byte-identical to `payload.txt`
### Scenario: -t accepts a real archive and rejects a corrupt one
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
- Fixture file `broken.zst` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
integrity matters
```
_Fixture `broken.zst`:_
```text
this is not a zstd frame
```
#### When
```shell
zstd -q payload.txt -o good.zst
zstd -t good.zst
zstd -t broken.zst
```
#### Then
- after `zstd -q payload.txt -o good.zst`:
  - exit code is `0`
- after `zstd -t good.zst`:
  - exit code is `0`
- after `zstd -t broken.zst`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `broken.zst`
### Scenario: decompressing a corrupt archive leaves no output behind
_only when `zstd --version` succeeds_
#### Given
- Fixture file `broken.zst` is created.
#### Inputs
_Fixture `broken.zst`:_
```text
this is not a zstd frame
```
#### When
```shell
zstd -dq broken.zst -o restored.txt
```
#### Then
- exit code is `1`
- stdout is empty
- file `restored.txt` does not exist
### Scenario: a missing input fails and names the file
_only when `zstd --version` succeeds_
#### When
```shell
zstd -q no-such-input.txt
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `no-such-input.txt`
### Scenario: an unknown option fails without producing output
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
data
```
#### When
```shell
zstd --not-a-real-option payload.txt
```
#### Then
- exit code is `1`
- stderr contains `--not-a-real-option`
- file `payload.txt.zst` does not exist
### Scenario: the stdin to stdout pipeline round trips
_only when `zstd --version` succeeds_
#### When
```shell
printf 'piped payload\n' | zstd -q | zstd -dq
```
#### Then
- exit code is `0`
- stdout equals an exact value
#### Expected output
_expected stdout:_
```text
piped payload
```
### Scenario: -l reports one frame and the checksum algorithm
_only when `zstd --version` succeeds_
#### Given
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
listed archive
```
#### When
```shell
zstd -q payload.txt -o listed.zst
zstd -l listed.zst
```
#### Then
- after `zstd -q payload.txt -o listed.zst`:
  - exit code is `0`
- after `zstd -l listed.zst`:
  - exit code is `0`
  - stdout contains `listed.zst`, `XXH64`
