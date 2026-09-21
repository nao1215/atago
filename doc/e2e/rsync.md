# atago Behavior Specs
## Summary
3 suites · 31 scenarios
## Contents
- [rsync (filter rules, file lists, and symlinks)](#rsync-filter-rules-file-lists-and-symlinks) — 9 scenarios
  - [an unanchored pattern matches at any depth, an anchored one only at the root](#scenario-an-unanchored-pattern-matches-at-any-depth-an-anchored-one-only-at-the-root)
  - [including directories then excluding everything else leaves empty chains](#scenario-including-directories-then-excluding-everything-else-leaves-empty-chains)
  - [--prune-empty-dirs drops the directory chains that end up empty](#scenario---prune-empty-dirs-drops-the-directory-chains-that-end-up-empty)
  - [--delete spares excluded destination files; --delete-excluded removes them](#scenario---delete-spares-excluded-destination-files---delete-excluded-removes-them)
  - [--files-from copies exactly the listed paths and keeps their layout](#scenario---files-from-copies-exactly-the-listed-paths-and-keeps-their-layout)
  - [--files-from: a listed path that does not exist is exit 23](#scenario---files-from-a-listed-path-that-does-not-exist-is-exit-23)
  - [-a copies a symlink as a symlink](#scenario--a-copies-a-symlink-as-a-symlink)
  - [-L replaces a symlink with its target's bytes](#scenario--l-replaces-a-symlink-with-its-targets-bytes)
  - [-r without -l skips a symlink with a stdout notice and still exits 0](#scenario--r-without--l-skips-a-symlink-with-a-stdout-notice-and-still-exits-0)
- [rsync (exit codes and stream separation)](#rsync-exit-codes-and-stream-separation) — 10 scenarios
  - [a successful copy is silent and exits 0](#scenario-a-successful-copy-is-silent-and-exits-0)
  - [exit 1: an unknown option is a usage error that copies nothing](#scenario-exit-1-an-unknown-option-is-a-usage-error-that-copies-nothing)
  - [exit 1: no arguments prints the usage on stderr, not stdout](#scenario-exit-1-no-arguments-prints-the-usage-on-stderr-not-stdout)
  - [a forgotten destination turns a --delete sync into a listing that exits 0](#scenario-a-forgotten-destination-turns-a---delete-sync-into-a-listing-that-exits-0)
  - [exit 11: a destination whose parent is missing fails and creates nothing](#scenario-exit-11-a-destination-whose-parent-is-missing-fails-and-creates-nothing)
  - [--mkpath creates the missing parents that exit 11 refused](#scenario---mkpath-creates-the-missing-parents-that-exit-11-refused)
  - [exit 23: a missing source is reported as a partial transfer](#scenario-exit-23-a-missing-source-is-reported-as-a-partial-transfer)
  - [exit 23: an unreadable file is skipped and everything else still arrives](#scenario-exit-23-an-unreadable-file-is-skipped-and-everything-else-still-arrives)
  - [exit 25: --max-delete stops deleting at the limit and says how many it kept](#scenario-exit-25---max-delete-stops-deleting-at-the-limit-and-says-how-many-it-kept)
  - [listing a single source prints it and changes nothing](#scenario-listing-a-single-source-prints-it-and-changes-nothing)
- [rsync (what a sync changes and what it leaves alone)](#rsync-what-a-sync-changes-and-what-it-leaves-alone) — 12 scenarios
  - [a trailing slash copies the contents, no slash copies the directory](#scenario-a-trailing-slash-copies-the-contents-no-slash-copies-the-directory)
  - [a second identical run is a no-op](#scenario-a-second-identical-run-is-a-no-op)
  - [--dry-run reports the plan and changes nothing, then the real run does exactly that](#scenario---dry-run-reports-the-plan-and-changes-nothing-then-the-real-run-does-exactly-that)
  - [--delete removes extraneous files and directories and nothing else](#scenario---delete-removes-extraneous-files-and-directories-and-nothing-else)
  - [without --delete an extraneous destination file survives](#scenario-without---delete-an-extraneous-destination-file-survives)
  - [the quick check skips a same-size, same-mtime file that --checksum then fixes](#scenario-the-quick-check-skips-a-same-size-same-mtime-file-that---checksum-then-fixes)
  - [-a preserves the mtime, so the next run skips; -r alone does not](#scenario--a-preserves-the-mtime-so-the-next-run-skips--r-alone-does-not)
  - [--update keeps a destination file that is newer than the source](#scenario---update-keeps-a-destination-file-that-is-newer-than-the-source)
  - [--ignore-existing adds new files and never touches existing ones](#scenario---ignore-existing-adds-new-files-and-never-touches-existing-ones)
  - [--backup keeps the replaced destination file under the suffix](#scenario---backup-keeps-the-replaced-destination-file-under-the-suffix)
  - [--remove-source-files moves the files and leaves the directories](#scenario---remove-source-files-moves-the-files-and-leaves-the-directories)
  - [awkward bytes and file names arrive unchanged](#scenario-awkward-bytes-and-file-names-arrive-unchanged)

## rsync (filter rules, file lists, and symlinks)
Filter rules are where rsync surprises people most, and a wrong rule does not
fail: it silently copies too much or too little. This suite fixes the rules
by the tree they produce, asserted with `dir:` and `changes:`, so a pattern
that matches one path more or fewer turns a scenario red.

It covers an unanchored pattern against an anchored one; the
include-directories-then-exclude-everything idiom and the empty directory
chains it leaves unless `--prune-empty-dirs` is given; excluded files at the
destination surviving `--delete` and going away with `--delete-excluded`;
rule files read with `--exclude-from`; an explicit `--files-from` list; and
the three ways rsync treats a symlink: kept as a link by `-a`, replaced by
its target's bytes by `-L`, and skipped with only a stdout notice and exit 0
by a bare `-r`.

All trees are written by the spec, so nothing from rsync is committed.

Source: `test/e2e/thirdparty/rsync/filters.atago.yaml`
### Scenario: an unanchored pattern matches at any depth, an anchored one only at the root
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.log` is created.
- Fixture file `src/keep/b.log` is created.
- Fixture file `src/keep/c.txt` is created.
- Fixture file `src/cache/d.txt` is created.
- Fixture file `src/keep/cache/e.txt` is created.
- Fixture file `rules` is created.

#### Inputs
_Fixture `src/a.log`:_
```text
1
```
_Fixture `src/keep/b.log`:_
```text
2
```
_Fixture `src/keep/c.txt`:_
```text
3
```
_Fixture `src/cache/d.txt`:_
```text
4
```
_Fixture `src/keep/cache/e.txt`:_
```text
5
```
_Fixture `rules`:_
```text
*.log
/cache/
```
#### When
```shell
rsync -a --exclude-from=rules src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/keep/c.txt`, `dst/keep/cache/e.txt`, modified nothing, deleted nothing
- file `dst/cache` does not exist

### Scenario: including directories then excluding everything else leaves empty chains
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/top.log` is created.
- Fixture file `src/a/b/deep.log` is created.
- Fixture file `src/a/b/deep.txt` is created.
- Fixture file `src/only_txt/x.txt` is created.
- Fixture file `rules` is created.

#### Inputs
_Fixture `src/top.log`:_
```text
1
```
_Fixture `src/a/b/deep.log`:_
```text
2
```
_Fixture `src/a/b/deep.txt`:_
```text
3
```
_Fixture `src/only_txt/x.txt`:_
```text
4
```
_Fixture `rules`:_
```text
+ */
+ *.log
- *
```
#### When
```shell
rsync -a --filter=". rules" src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/top.log`, `dst/a/b/deep.log`, modified nothing, deleted nothing
- dir `dst/only_txt` has 0 entries

### Scenario: --prune-empty-dirs drops the directory chains that end up empty
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/top.log` is created.
- Fixture file `src/a/b/deep.log` is created.
- Fixture file `src/only_txt/nested/x.txt` is created.
- Fixture file `rules` is created.

#### Inputs
_Fixture `src/top.log`:_
```text
1
```
_Fixture `src/a/b/deep.log`:_
```text
2
```
_Fixture `src/only_txt/nested/x.txt`:_
```text
4
```
_Fixture `rules`:_
```text
+ */
+ *.log
- *
```
#### When
```shell
rsync -a -m --filter=". rules" src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/top.log`, `dst/a/b/deep.log`, modified nothing, deleted nothing
- dir `dst` contains `top.log`, contains `a`, has 2 entries
- file `dst/only_txt` does not exist

### Scenario: --delete spares excluded destination files; --delete-excluded removes them
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/app.conf` is created.
- Fixture file `dst/app.conf` is created.
- Fixture file `dst/local.secret` is created.
- Fixture file `dst/stray.txt` is created.

#### Inputs
_Fixture `src/app.conf`:_
```text
conf
```
_Fixture `dst/app.conf`:_
```text
conf
```
_Fixture `dst/local.secret`:_
```text
do not lose me
```
_Fixture `dst/stray.txt`:_
```text
stray
```
#### When
```shell
rsync -a --delete --exclude=*.secret src/ dst/
rsync -a --delete --delete-excluded --exclude=*.secret src/ dst/
```
#### Then
- after `rsync -a --delete --exclude=*.secret src/ dst/`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted `dst/stray.txt`
  - file `dst/local.secret` contains `do not lose me`
- after `rsync -a --delete --delete-excluded --exclude=*.secret src/ dst/`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted `dst/local.secret`

### Scenario: --files-from copies exactly the listed paths and keeps their layout
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/b.txt` is created.
- Fixture file `src/sub/c.txt` is created.
- Fixture file `src/sub/d.txt` is created.
- Fixture file `list` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
_Fixture `src/b.txt`:_
```text
b
```
_Fixture `src/sub/c.txt`:_
```text
c
```
_Fixture `src/sub/d.txt`:_
```text
d
```
_Fixture `list`:_
```text
a.txt
sub/c.txt
```
#### When
```shell
rsync -a --files-from=list src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/a.txt`, `dst/sub/c.txt`, modified nothing, deleted nothing

### Scenario: --files-from: a listed path that does not exist is exit 23
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `list` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
a
```
_Fixture `list`:_
```text
a.txt
ghost.txt
```
#### When
```shell
rsync -a --files-from=list src/ dst/
```
#### Then
- exit code is `23`
- stdout is empty
- stderr contains `ghost.txt" failed: No such file or directory`, `(code 23)`
- the step changed exactly created `dst/a.txt`, modified nothing, deleted nothing

### Scenario: -a copies a symlink as a symlink
_only when `rsync --mkpath --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/target.txt` is created.

#### Inputs
_Fixture `src/target.txt`:_
```text
target bytes
```
#### When
```shell
ln -s target.txt src/link
rsync -a src/ dst/
readlink dst/link
```
#### Then
- after `rsync -a src/ dst/`:
  - exit code is `0`
- after `readlink dst/link`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
target.txt
```
### Scenario: -L replaces a symlink with its target's bytes
_only when `rsync --mkpath --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/target.txt` is created.

#### Inputs
_Fixture `src/target.txt`:_
```text
target bytes
```
#### When
```shell
ln -s target.txt src/link
rsync -rL src/ dst/
test -L dst/link
```
#### Then
- after `rsync -rL src/ dst/`:
  - exit code is `0`
  - file `dst/link` is byte-identical to `src/target.txt`
- after `test -L dst/link`:
  - exit code is `1`

### Scenario: -r without -l skips a symlink with a stdout notice and still exits 0
_only when `rsync --mkpath --version` succeeds · skipped on Windows_
#### Given
- Fixture file `src/target.txt` is created.

#### Inputs
_Fixture `src/target.txt`:_
```text
target bytes
```
#### When
```shell
ln -s target.txt src/link
rsync -r src/ dst/
```
#### Then
- after `rsync -r src/ dst/`:
  - exit code is `0`
  - stdout equals an exact value
  - stderr is empty
  - the step changed exactly created `dst/target.txt`, modified nothing, deleted nothing
  - file `dst/link` does not exist

#### Expected output
_expected stdout:_
```text
skipping non-regular file "link"
```
## rsync (exit codes and stream separation)
[rsync](https://rsync.samba.org/) documents a table of exit codes, and a
backup script branches on them: 0 means the destination now mirrors the
source, 23 means some files did not make it, 25 means a safety limit
stopped deletions part way. This suite pins every code that a local copy
can reach — 0, 1 for a usage error, 11 for a file I/O error, 23 for a
partial transfer, and 25 for the `--max-delete` limit — together with what
each one leaves behind on disk, because a code is only useful when it tells
the truth about the tree.

Every failure is also checked for stream separation: the diagnosis goes to
stderr and stdout stays empty, so a script that parses stdout never
mistakes an error for a transfer log. The `rsync error: ... at main.c(N)`
trailer carries a source line number that moves between releases, so the
specs match the stable prefix only.

All trees are written by the spec, so nothing from rsync is committed.

Source: `test/e2e/thirdparty/rsync/rsync.atago.yaml`
### Scenario: a successful copy is silent and exits 0
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
_Fixture `src/sub/b.txt`:_
```text
bravo
```
#### When
```shell
rsync -a src/ dst/
```
#### Then
- exit code is `0`
- stdout is empty
- stderr is empty
- the step changed exactly created `dst/a.txt`, `dst/sub/b.txt`, modified nothing, deleted nothing

### Scenario: exit 1: an unknown option is a usage error that copies nothing
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
#### When
```shell
rsync -a --no-such-option src/ dst/
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `--no-such-option: unknown option`, `rsync error: syntax or usage error (code 1)`
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: exit 1: no arguments prints the usage on stderr, not stdout
_only when `rsync --mkpath --version` succeeds_
#### When
```shell
rsync
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Usage: rsync [OPTION]... SRC [SRC]... DEST`, `rsync error: syntax or usage error (code 1)`

### Scenario: a forgotten destination turns a --delete sync into a listing that exits 0
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
#### When
```shell
rsync -a --delete src/
```
#### Then
- exit code is `0`
- stdout matches `/(?m) a\.txt$/`
- stderr is empty
- the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: exit 11: a destination whose parent is missing fails and creates nothing
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
#### When
```shell
rsync -a src/ missing/parent/dst/
```
#### Then
- exit code is `11`
- stdout is empty
- stderr contains `mkdir "`, `missing/parent/dst" failed: No such file or directory`, `rsync error: error in file IO (code 11)`
- file `missing` does not exist

### Scenario: --mkpath creates the missing parents that exit 11 refused
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
#### When
```shell
rsync -a --mkpath src/ missing/parent/dst/
```
#### Then
- exit code is `0`
- stdout is empty
- the step changed exactly created `missing/parent/dst/a.txt`, modified nothing, deleted nothing
- file `missing/parent/dst/a.txt` is byte-identical to `src/a.txt`

### Scenario: exit 23: a missing source is reported as a partial transfer
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `present.txt` is created.

#### Inputs
_Fixture `present.txt`:_
```text
here
```
#### When
```shell
rsync -a present.txt absent.txt dst/
```
#### Then
- exit code is `23`
- stdout is empty
- stderr contains `link_stat "`, `absent.txt" failed: No such file or directory`, `some files/attrs were not transferred`, `(code 23)`
- the step changed exactly created `dst/present.txt`, modified nothing, deleted nothing

### Scenario: exit 23: an unreadable file is skipped and everything else still arrives
_only when `rsync --mkpath --version` succeeds · skipped when `test "$(id -u)" = 0` succeeds_
#### Given
- Fixture file `src/ok.txt` is created.
- Fixture file `src/locked.txt` is created.

#### Inputs
_Fixture `src/ok.txt`:_
```text
readable
```
_Fixture `src/locked.txt`:_
```text
secret
```
#### When
```shell
rsync -a src/ dst/
```
#### Then
- exit code is `23`
- stdout is empty
- stderr contains `locked.txt": Permission denied`, `(code 23)`
- the step changed exactly created `dst/ok.txt`, modified nothing, deleted nothing
- file `dst/locked.txt` does not exist

### Scenario: exit 25: --max-delete stops deleting at the limit and says how many it kept
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `dst/a.txt` is created.
- Fixture file `dst/extra1` is created.
- Fixture file `dst/extra2` is created.
- Fixture file `dst/extra3` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
_Fixture `dst/a.txt`:_
```text
alpha
```
_Fixture `dst/extra1`:_
```text
x
```
_Fixture `dst/extra2`:_
```text
x
```
_Fixture `dst/extra3`:_
```text
x
```
#### When
```shell
rsync -a --delete --max-delete=1 src/ dst/
```
#### Then
- exit code is `25`
- stdout is empty
- stderr contains `Deletions stopped due to --max-delete limit (2 skipped)`, `(code 25)`
- dir `dst` contains `a.txt`, has 3 entries

### Scenario: listing a single source prints it and changes nothing
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
_Fixture `src/sub/b.txt`:_
```text
bravo
```
#### When
```shell
rsync src/
```
#### Then
- exit code is `0`
- stdout matches `/(?m)^d[rwx-]{9} +[0-9,]+ [0-9/]+ [0-9:]+ \.$/`
- stderr is empty
- the step changed exactly created nothing, modified nothing, deleted nothing
- stdout matches `/(?m)^-[rwx-]{9} +6 [0-9/]+ [0-9:]+ a\.txt$/`
- stdout matches `/(?m)^d[rwx-]{9} +[0-9,]+ [0-9/]+ [0-9:]+ sub$/`
- stdout does not contain `b.txt`

## rsync (what a sync changes and what it leaves alone)
An rsync run is judged by the tree it leaves behind. This suite asserts that
tree with `changes:`, which is exhaustive in both directions, so a scenario
fails when rsync writes, rewrites, or deletes one path more or one path
fewer than the contract says.

The contracts covered are the ones a backup or deploy script leans on: the
trailing slash on the source decides whether the directory itself or only
its contents is copied; a second identical run is a no-op; `--dry-run`
reports the plan and changes nothing, and the real run then does exactly
that plan; `--delete` removes only what the source no longer has; the quick
check trusts size and modification time, so a same-size, same-time file with
different bytes is skipped unless `--checksum` is given; `--update`,
`--ignore-existing`, and `--backup` each protect a destination file in their
own way; `--remove-source-files` moves files but never directories; and the
bytes that arrive are the bytes that left, checked with `equals_file` for an
empty file, NUL bytes, CRLF without a final newline, and a file name with
spaces and non-ASCII characters.

All trees are written by the spec, so nothing from rsync is committed.

Source: `test/e2e/thirdparty/rsync/sync.atago.yaml`
### Scenario: a trailing slash copies the contents, no slash copies the directory
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
_Fixture `src/sub/b.txt`:_
```text
bravo
```
#### When
```shell
rsync -a src dir_itself
rsync -a src/ contents_only
```
#### Then
- after `rsync -a src dir_itself`:
  - exit code is `0`
  - the step changed exactly created `dir_itself/src/a.txt`, `dir_itself/src/sub/b.txt`, modified nothing, deleted nothing
- after `rsync -a src/ contents_only`:
  - exit code is `0`
  - the step changed exactly created `contents_only/a.txt`, `contents_only/sub/b.txt`, modified nothing, deleted nothing

### Scenario: a second identical run is a no-op
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/a.txt` is created.
- Fixture file `src/sub/b.txt` is created.

#### Inputs
_Fixture `src/a.txt`:_
```text
alpha
```
_Fixture `src/sub/b.txt`:_
```text
bravo
```
#### When
```shell
rsync -a src/ dst/
rsync -a -i src/ dst/
```
#### Then
- after `rsync -a src/ dst/`:
  - exit code is `0`
- after `rsync -a -i src/ dst/`:
  - exit code is `0`
  - stdout is empty
  - the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: --dry-run reports the plan and changes nothing, then the real run does exactly that
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/new.txt` is created.
- Fixture file `src/same.txt` is created.
- Fixture file `dst/same.txt` is created.
- Fixture file `dst/stale.txt` is created.

#### Inputs
_Fixture `src/new.txt`:_
```text
new
```
_Fixture `src/same.txt`:_
```text
same
```
_Fixture `dst/same.txt`:_
```text
same
```
_Fixture `dst/stale.txt`:_
```text
stale
```
#### When
```shell
touch -r src/same.txt dst/same.txt
rsync -a -i --delete --dry-run src/ dst/
rsync -a -i --delete src/ dst/
```
#### Then
- after `rsync -a -i --delete --dry-run src/ dst/`:
  - exit code is `0`
  - stdout contains `*deleting   stale.txt`, `>f+++++++++ new.txt`
  - the step changed exactly created nothing, modified nothing, deleted nothing
  - stdout does not contain `same.txt`
- after `rsync -a -i --delete src/ dst/`:
  - exit code is `0`
  - stdout contains `*deleting   stale.txt`, `>f+++++++++ new.txt`
  - the step changed exactly created `dst/new.txt`, modified nothing, deleted `dst/stale.txt`

### Scenario: --delete removes extraneous files and directories and nothing else
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/keep.txt` is created.
- Fixture file `src/sub/keep.txt` is created.
- Fixture file `dst/keep.txt` is created.
- Fixture file `dst/sub/keep.txt` is created.
- Fixture file `dst/sub/gone.txt` is created.
- Fixture file `dst/olddir/deep/gone.txt` is created.

#### Inputs
_Fixture `src/keep.txt`:_
```text
keep
```
_Fixture `src/sub/keep.txt`:_
```text
keep
```
_Fixture `dst/keep.txt`:_
```text
keep
```
_Fixture `dst/sub/keep.txt`:_
```text
keep
```
_Fixture `dst/sub/gone.txt`:_
```text
gone
```
_Fixture `dst/olddir/deep/gone.txt`:_
```text
gone
```
#### When
```shell
rsync -a --delete src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, deleted `dst/sub/gone.txt`, `dst/olddir/deep/gone.txt`
- file `dst/olddir` does not exist

### Scenario: without --delete an extraneous destination file survives
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/keep.txt` is created.
- Fixture file `dst/extra.txt` is created.

#### Inputs
_Fixture `src/keep.txt`:_
```text
keep
```
_Fixture `dst/extra.txt`:_
```text
mine
```
#### When
```shell
rsync -a src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/keep.txt`, modified nothing, deleted nothing

### Scenario: the quick check skips a same-size, same-mtime file that --checksum then fixes
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/f.txt` is created.
- Fixture file `dst/f.txt` is created.

#### Inputs
_Fixture `src/f.txt`:_
```text
AAAA
```
_Fixture `dst/f.txt`:_
```text
BBBB
```
#### When
```shell
touch -r src/f.txt dst/f.txt
rsync -a -i src/ dst/
rsync -a -i -c src/ dst/
```
#### Then
- after `rsync -a -i src/ dst/`:
  - exit code is `0`
  - stdout does not contain `f.txt`
  - the step changed exactly created nothing, modified nothing, deleted nothing
- after `rsync -a -i -c src/ dst/`:
  - exit code is `0`
  - stdout contains `>fc........ f.txt`
  - the step changed exactly created nothing, modified `dst/f.txt`, deleted nothing
  - file `dst/f.txt` is byte-identical to `src/f.txt`

### Scenario: -a preserves the mtime, so the next run skips; -r alone does not
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/f.txt` is created.

#### Inputs
_Fixture `src/f.txt`:_
```text
payload
```
#### When
```shell
touch -t 202001020304.05 src/f.txt
rsync -r src/ plain/
rsync -r -i src/ plain/
rsync -a src/ archived/
rsync -a -i src/ archived/
find archived/f.txt plain/f.txt -newer src/f.txt
```
#### Then
- after `rsync -r -i src/ plain/`:
  - exit code is `0`
  - stdout equals an exact value
- after `rsync -a -i src/ archived/`:
  - exit code is `0`
  - stdout is empty
- after `find archived/f.txt plain/f.txt -newer src/f.txt`:
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
>f..T...... f.txt
```
_expected stdout:_
```text
plain/f.txt
```
### Scenario: --update keeps a destination file that is newer than the source
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/f.txt` is created.
- Fixture file `dst/f.txt` is created.

#### Inputs
_Fixture `src/f.txt`:_
```text
old source
```
_Fixture `dst/f.txt`:_
```text
newer edit
```
#### When
```shell
touch -t 202001020304.05 src/f.txt
rsync -a -u src/ dst/
rsync -a src/ dst/
```
#### Then
- after `rsync -a -u src/ dst/`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted nothing
  - file `dst/f.txt` contains `newer edit`
- after `rsync -a src/ dst/`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `dst/f.txt`, deleted nothing

### Scenario: --ignore-existing adds new files and never touches existing ones
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/exists.txt` is created.
- Fixture file `src/fresh.txt` is created.
- Fixture file `dst/exists.txt` is created.

#### Inputs
_Fixture `src/exists.txt`:_
```text
from source
```
_Fixture `src/fresh.txt`:_
```text
fresh
```
_Fixture `dst/exists.txt`:_
```text
hand edited
```
#### When
```shell
rsync -a --ignore-existing src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/fresh.txt`, modified nothing, deleted nothing
- file `dst/exists.txt` contains `hand edited`

### Scenario: --backup keeps the replaced destination file under the suffix
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/f.txt` is created.
- Fixture file `dst/f.txt` is created.

#### Inputs
_Fixture `src/f.txt`:_
```text
version two, longer
```
_Fixture `dst/f.txt`:_
```text
version one
```
#### When
```shell
rsync -a --backup --suffix=.bak src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/f.txt.bak`, modified `dst/f.txt`, deleted nothing
- file `dst/f.txt.bak` contains `version one`
- file `dst/f.txt` is byte-identical to `src/f.txt`

### Scenario: --remove-source-files moves the files and leaves the directories
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/one.txt` is created.
- Fixture file `src/sub/two.txt` is created.

#### Inputs
_Fixture `src/one.txt`:_
```text
1
```
_Fixture `src/sub/two.txt`:_
```text
2
```
#### When
```shell
rsync -a --remove-source-files src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/one.txt`, `dst/sub/two.txt`, modified nothing, deleted `src/one.txt`, `src/sub/two.txt`
- dir `src` contains `sub`, has 1 entry

### Scenario: awkward bytes and file names arrive unchanged
_only when `rsync --mkpath --version` succeeds_
#### Given
- Fixture file `src/empty.bin` is created.
- Fixture file `src/nul.bin` is created.
- Fixture file `src/mixed.txt` is created.
- Fixture file `src/with space/日本語 ファイル.txt` is created.

#### Inputs
_Fixture `src/mixed.txt`:_
```text
unix
windows
no trailing newline
```
_Fixture `src/with space/日本語 ファイル.txt`:_
```text
名前
```
#### When
```shell
rsync -a src/ dst/
```
#### Then
- exit code is `0`
- the step changed exactly created `dst/empty.bin`, `dst/nul.bin`, `dst/mixed.txt`, `dst/with space/日本語 ファイル.txt`, modified nothing, deleted nothing
- file `dst/empty.bin` is byte-identical to `src/empty.bin`
- file `dst/nul.bin` is byte-identical to `src/nul.bin`
- file `dst/mixed.txt` is byte-identical to `src/mixed.txt`
- file `dst/with space/日本語 ファイル.txt` is byte-identical to `src/with space/日本語 ファイル.txt`
