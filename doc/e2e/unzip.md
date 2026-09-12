# atago Behavior Specs
## Summary
2 suites · 16 scenarios
## Contents
- [zip and unzip (archive round trips and extracted trees)](#zip-and-unzip-archive-round-trips-and-extracted-trees) — 13 scenarios
  - [zipping a tree keeps the inputs and adds exactly one archive](#scenario-zipping-a-tree-keeps-the-inputs-and-adds-exactly-one-archive)
  - [an extracted tree has exactly the paths the archive holds](#scenario-an-extracted-tree-has-exactly-the-paths-the-archive-holds)
  - [the extracted tree matches its snapshot manifest](#scenario-the-extracted-tree-matches-its-snapshot-manifest)
  - [a round trip restores every byte, including the awkward ones](#scenario-a-round-trip-restores-every-byte-including-the-awkward-ones)
  - [unzip -j drops the directories and keeps the files](#scenario-unzip--j-drops-the-directories-and-keeps-the-files)
  - [listing an archive names every entry without writing anything](#scenario-listing-an-archive-names-every-entry-without-writing-anything)
  - [an integrity test passes on a good archive and fails on a truncated one](#scenario-an-integrity-test-passes-on-a-good-archive-and-fails-on-a-truncated-one)
  - [a missing archive is exit 9 and creates nothing](#scenario-a-missing-archive-is-exit-9-and-creates-nothing)
  - [an unknown option is exit 10 and a name that matches nothing is exit 11](#scenario-an-unknown-option-is-exit-10-and-a-name-that-matches-nothing-is-exit-11)
  - [zip with nothing to archive is exit 12 and writes no archive](#scenario-zip-with-nothing-to-archive-is-exit-12-and-writes-no-archive)
  - [overwriting is refused by default, forced by -o, and skipped by -n](#scenario-overwriting-is-refused-by-default-forced-by--o-and-skipped-by--n)
  - [an updated archive replaces one entry and leaves the rest alone](#scenario-an-updated-archive-replaces-one-entry-and-leaves-the-rest-alone)
  - [deleting an entry removes it from the extracted tree](#scenario-deleting-an-entry-removes-it-from-the-extracted-tree)
- [unzip (path traversal in a hostile archive)](#unzip-path-traversal-in-a-hostile-archive) — 3 scenarios
  - [a traversing entry is stripped and lands inside the destination](#scenario-a-traversing-entry-is-stripped-and-lands-inside-the-destination)
  - [a deeper traversal is stripped down to its basename](#scenario-a-deeper-traversal-is-stripped-down-to-its-basename)
  - [unzip -t reports the hostile archive as intact](#scenario-unzip--t-reports-the-hostile-archive-as-intact)

## zip and unzip (archive round trips and extracted trees)
[Info-ZIP](https://infozip.sourceforge.net/) `zip` and `unzip` are the
archivers a shell script reaches for. An archiver makes two promises worth
testing: what comes back out is what went in, byte for byte, and the tree it
writes is exactly the tree the archive describes — no more paths, no fewer.

This suite pins both. Round trips are compared with `equals_file`, so no
newline or encoding normalization can hide a difference. The extracted tree
is asserted with `dir:` in every mode the family offers — recursive
membership, a count over regular files, a glob, and a snapshot manifest that
fixes the whole shape at once — and `changes:` states exhaustively which
paths a command created, so "it also wrote something else" fails. The
documented exit codes get one scenario each: 0, the warning code 1, 9 for a
missing or unreadable archive, 10 for a bad option, 11 when a name pattern
matches nothing, and zip's 12 for nothing to do.

Every archive is built by the spec from files the spec writes, so nothing is
committed and no scenario depends on the host's contents.

Source: `test/e2e/thirdparty/unzip/unzip.atago.yaml`
### Scenario: zipping a tree keeps the inputs and adds exactly one archive
_only when `zip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
_Fixture `src/sub/b.txt`:_
```text
nested
```
#### When
```shell
zip -q -r ar.zip src
```
#### Then
- exit code is `0`
- the step changed exactly created `ar.zip`, modified nothing, deleted nothing

### Scenario: an extracted tree has exactly the paths the archive holds
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.
- Fixture file `src/sub/deep/c.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
_Fixture `src/sub/b.txt`:_
```text
nested
```
_Fixture `src/sub/deep/c.txt`:_
```text
deeper
```
#### When
```shell
zip -q -r ar.zip src
unzip -q ar.zip -d out
```
#### Then
- after `zip -q -r ar.zip src`:
  - exit code is `0`
- after `unzip -q ar.zip -d out`:
  - exit code is `0`
  - dir `out` contains `src/a.txt`, contains `src/sub/b.txt`, contains `src/sub/deep/c.txt`, (recursive)
  - dir `out` has 3 entries, (recursive)
  - dir `out` matches glob `src/sub/deep/*.txt`, (recursive)
  - dir `out` contains `src`, has 1 entry

### Scenario: the extracted tree matches its snapshot manifest
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
_Fixture `src/sub/b.txt`:_
```text
nested
```
#### When
```shell
zip -q -r ar.zip src
unzip -q ar.zip -d out
```
#### Then
- after `unzip -q ar.zip -d out`:
  - exit code is `0`
  - dir `out` tree matches snapshot `extracted_tree`

### Scenario: a round trip restores every byte, including the awkward ones
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/empty.bin` is created.
- Fixture file `src/mixed.txt` is created.
- Fixture file `src/nul.bin` is created.

#### Inputs
_Fixture `src/mixed.txt`:_
```text
unix
windows
no trailing newline
```
#### When
```shell
zip -q -r ar.zip src
unzip -q ar.zip -d out
```
#### Then
- after `zip -q -r ar.zip src`:
  - exit code is `0`
- after `unzip -q ar.zip -d out`:
  - exit code is `0`
  - file `out/src/mixed.txt` is byte-identical to `src/mixed.txt`
  - file `out/src/nul.bin` is byte-identical to `src/nul.bin`
  - file `out/src/empty.bin` is byte-identical to `src/empty.bin`

### Scenario: unzip -j drops the directories and keeps the files
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
_Fixture `src/sub/b.txt`:_
```text
nested
```
#### When
```shell
zip -q -r ar.zip src
unzip -j -q ar.zip -d flat
```
#### Then
- after `unzip -j -q ar.zip -d flat`:
  - exit code is `0`
  - dir `flat` contains `a.txt`, contains `b.txt`, does not contain `src`, does not contain `sub`, has 2 entries, (recursive)
  - file `flat/b.txt` is byte-identical to `src/sub/b.txt`

### Scenario: listing an archive names every entry without writing anything
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
_Fixture `src/sub/b.txt`:_
```text
nested
```
#### When
```shell
zip -q -r ar.zip src
unzip -l ar.zip
```
#### Then
- after `unzip -l ar.zip`:
  - exit code is `0`
  - stdout contains `src/a.txt`, `src/sub/b.txt`, `4 files`
  - the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: an integrity test passes on a good archive and fails on a truncated one
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `broken.zip` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
_Fixture `broken.zip`:_
```text
PK truncated garbage
```
#### When
```shell
zip -q -r ar.zip src
unzip -t ar.zip
unzip -t broken.zip
```
#### Then
- after `unzip -t ar.zip`:
  - exit code is `0`
  - stdout contains `No errors detected in compressed data`
- after `unzip -t broken.zip`:
  - exit code is `9`
  - stdout contains `End-of-central-directory signature not found`

### Scenario: a missing archive is exit 9 and creates nothing
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
unzip -q absent.zip -d out
```
#### Then
- exit code is `9`
- stderr contains `cannot find or open absent.zip`
- the step changed exactly created nothing, modified nothing, deleted nothing
- dir `out` does not exist

### Scenario: an unknown option is exit 10 and a name that matches nothing is exit 11
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
hello
```
#### When
```shell
zip -q -r ar.zip src
unzip -Q ar.zip
unzip -q ar.zip "no/such/*" -d out
```
#### Then
- after `unzip -Q ar.zip`:
  - exit code is `10`
  - stdout is empty
  - stderr contains `Usage: unzip`
- after `unzip -q ar.zip "no/such/*" -d out`:
  - exit code is `11`
  - stderr contains `filename not matched`
  - dir `out` exists, has 0 entries

### Scenario: zip with nothing to archive is exit 12 and writes no archive
_only when `zip -v` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
zip -q empty.zip
```
#### Then
- exit code is `12`
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: overwriting is refused by default, forced by -o, and skipped by -n
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `out/src/a.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
from the archive
```
_Fixture `out/src/a.txt`:_
```text
already here
```
#### When
```shell
zip -q -r ar.zip src
unzip -q ar.zip -d out < /dev/null
unzip -qn ar.zip -d out
unzip -qo ar.zip -d out
```
#### Then
- after `unzip -q ar.zip -d out < /dev/null`:
  - exit code is `1`
  - file `out/src/a.txt` equals exact bytes
- after `unzip -qn ar.zip -d out`:
  - exit code is `0`
  - file `out/src/a.txt` equals exact bytes
- after `unzip -qo ar.zip -d out`:
  - exit code is `0`
  - file `out/src/a.txt` is byte-identical to `src/a.txt`

### Scenario: an updated archive replaces one entry and leaves the rest alone
_only when `zip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/b.txt` is created.
- Fixture file `src/a.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `src/a.txt`:_
```text
first
```
_Fixture `src/b.txt`:_
```text
second
```
_Fixture `src/a.txt`:_
```text
rewritten
```
#### When
```shell
zip -q -r ar.zip src
zip -q ar.zip src/a.txt
unzip -q ar.zip -d out
```
#### Then
- after `zip -q -r ar.zip src`:
  - exit code is `0`
- after `zip -q ar.zip src/a.txt`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `ar.zip`, deleted nothing
- after `unzip -q ar.zip -d out`:
  - exit code is `0`
  - file `out/src/a.txt` equals exact bytes
  - file `out/src/b.txt` equals exact bytes

### Scenario: deleting an entry removes it from the extracted tree
_only when `zip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `src/keep.txt` is created.
- Fixture file `src/drop.txt` is created.

#### Inputs
_Fixture `src/keep.txt`:_
```text
keep
```
_Fixture `src/drop.txt`:_
```text
drop
```
#### When
```shell
zip -q -r ar.zip src
zip -q -d ar.zip "src/drop.txt"
unzip -q ar.zip -d out
```
#### Then
- after `zip -q -d ar.zip "src/drop.txt"`:
  - exit code is `0`
- after `unzip -q ar.zip -d out`:
  - exit code is `0`
  - dir `out` contains `src/keep.txt`, does not contain `src/drop.txt`, has 1 entry, (recursive)

## unzip (path traversal in a hostile archive)
An archive is untrusted input: the names inside it come from whoever built
it. The attack known as zip slip stores an entry called `../escaped.txt` (or
an absolute `/etc/passwd`) so that a careless extractor writes outside the
directory the user chose.

`unzip` refuses to. It strips the traversal, says so, extracts the entry
under the destination anyway, and reports the warning exit code 1 rather than
a silent 0. This suite pins all four parts of that contract, and proves the
filesystem claim rather than trusting the message: the destination's tree is
asserted exhaustively with `dir:`, and `changes:` states that the workdir
around the destination gained nothing.

The hostile archives cannot be produced by `zip` itself, which never stores a
traversing name, so each is written by the spec as base64 — bytes authored
here, not a sample taken from anywhere.

Source: `test/e2e/thirdparty/unzip/zip_slip.atago.yaml`
### Scenario: a traversing entry is stripped and lands inside the destination
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `hostile.zip` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
unzip -l hostile.zip
unzip hostile.zip -d dest
```
#### Then
- after `unzip -l hostile.zip`:
  - exit code is `0`
  - stdout contains `../escaped.txt`, `/abs.txt`
- after `unzip hostile.zip -d dest`:
  - exit code is `1`
  - stdout contains `skipped "../" path component(s) in ../escaped.txt`
  - stderr contains `stripped absolute path spec from /abs.txt`
  - the step changed exactly created `dest/escaped.txt`, `dest/abs.txt`, `dest/good.txt`, modified nothing, deleted nothing
  - dir `dest` contains `escaped.txt`, contains `abs.txt`, contains `good.txt`, has 3 entries, (recursive)
  - file `dest/escaped.txt` equals exact bytes

### Scenario: a deeper traversal is stripped down to its basename
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `deep.zip` is created.
- Fixture file `dest/placeholder.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `dest/placeholder.txt`:_
```text
kept
```
#### When
```shell
unzip -o deep.zip -d dest
```
#### Then
- exit code is `1`
- stdout contains `skipped "../" path component(s)`
- the step changed exactly created `dest/tmp/pwned.txt`, modified nothing, deleted nothing
- dir `dest` contains `tmp/pwned.txt`, contains `placeholder.txt`, has 2 entries, (recursive)
- file `dest/tmp/pwned.txt` equals exact bytes

### Scenario: unzip -t reports the hostile archive as intact
_only when `unzip -v` succeeds · skipped on Windows_
#### Given
- Fixture file `hostile.zip` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
unzip -t hostile.zip
```
#### Then
- exit code is `0`
- stdout contains `No errors detected in compressed data`
- the step changed exactly created nothing, modified nothing, deleted nothing
