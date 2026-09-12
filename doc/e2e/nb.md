# atago Behavior Specs
## Summary
2 suites · 12 scenarios
## Contents
- [nb (notes kept as git history)](#nb-notes-kept-as-git-history) — 7 scenarios
  - [without a git identity the first run explains itself and stops](#scenario-without-a-git-identity-the-first-run-explains-itself-and-stops)
  - [on a terminal the first run asks for the identity and keeps it](#scenario-on-a-terminal-the-first-run-asks-for-the-identity-and-keeps-it)
  - [the first command initializes the notebook instead of doing what it was asked](#scenario-the-first-command-initializes-the-notebook-instead-of-doing-what-it-was-asked)
  - [adding a note writes a file and commits it](#scenario-adding-a-note-writes-a-file-and-commits-it)
  - [editing appends to the note and commits the change](#scenario-editing-appends-to-the-note-and-commits-the-change)
  - [deleting removes the file and records that too](#scenario-deleting-removes-the-file-and-records-that-too)
  - [search reads the notes rather than their names](#scenario-search-reads-the-notes-rather-than-their-names)
- [nb (notebooks, targeting, and refusals)](#nb-notebooks-targeting-and-refusals) — 5 scenarios
  - [a new notebook is a new repository](#scenario-a-new-notebook-is-a-new-repository)
  - [a notebook can be named in the command instead of switched to](#scenario-a-notebook-can-be-named-in-the-command-instead-of-switched-to)
  - [use changes what the bare commands mean](#scenario-use-changes-what-the-bare-commands-mean)
  - [the same refusal arrives on two different streams](#scenario-the-same-refusal-arrives-on-two-different-streams)
  - [the listing is uncolored through a pipe and still carries control sequences](#scenario-the-listing-is-uncolored-through-a-pipe-and-still-carries-control-sequences)

## nb (notes kept as git history)
[nb](https://xwmx.github.io/nb/) is a note-taking CLI that keeps every
notebook as a git repository, so each thing it does to a note is a commit.
It keeps its own test suite in Bats; what those tests check is pinned here
from outside.

That design makes the oracle obvious: after add, edit and delete, the
assertion is the git log of the notebook, plus the files on disk. The first
run is asserted too, in both shapes — refused with an explanation when git
has no identity to attribute edits to, and completed on a terminal, where nb
asks for the name and email and writes them into the home it was given.

Everything lands under a sandboxed home, so the suite also proves nb keeps
its notes where it says it does.

Source: `test/e2e/thirdparty/nb/nb.atago.yaml`
### Scenario: without a git identity the first run explains itself and stops
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
nb add --content "a note"
```
#### Then
- exit code is `1`
- stdout contains `Welcome to nb!`, `Git requires some additional setup before using nb.`, `Enter the name and email address you'd like to use with Git.`

### Scenario: on a terminal the first run asks for the identity and keeps it
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
# interactive (pty): nb add --content "a note" --filename first.md
```
#### Then
- file `.atago-home/.gitconfig` contains `name = Test Person`, `email = test@example.com`
- dir `.atago-home/.nb/home` exists

### Scenario: the first command initializes the notebook instead of doing what it was asked
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author'
nb add --content "the first note" --filename note1.md
nb add --content "the first note" --filename note1.md
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author'`:
  - exit code is `0`
- after `nb add --content "the first note" --filename note1.md`:
  - exit code is `0`
  - stdout contains `Welcome to`, `0 items.`
  - file `.atago-home/.nb/home/note1.md` does not exist
- after `nb add --content "the first note" --filename note1.md`:
  - exit code is `0`
  - stdout contains `Added: [1] note1.md`
  - file `.atago-home/.nb/home/note1.md` exists

#### Generated artifacts
- `.atago-home/.nb/home/note1.md`

### Scenario: adding a note writes a file and commits it
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks
nb add --content "the first note" --filename note1.md
git -C .atago-home/.nb/home log --oneline
nb list
nb show 1 --print
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks`:
  - exit code is `0`
- after `nb add --content "the first note" --filename note1.md`:
  - exit code is `0`
  - stdout contains `Added: [1] note1.md`
  - file `.atago-home/.nb/home/note1.md` equals exact bytes
- after `git -C .atago-home/.nb/home log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ \[nb\] Add: note1\.md\n[0-9a-f]+ \[nb\] Initialize\n$/`
- after `nb list`:
  - exit code is `0`
  - stdout contains `[1] note1.md · "the first note"`
- after `nb show 1 --print`:
  - exit code is `0`
  - stdout contains `the first note`

### Scenario: editing appends to the note and commits the change
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "the first line" --filename note1.md
nb edit 1 --content "a second line"
git -C .atago-home/.nb/home log --oneline
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "the first line" --filename note1.md`:
  - exit code is `0`
- after `nb edit 1 --content "a second line"`:
  - exit code is `0`
  - stdout contains `Updated: [1] note1.md`
  - file `.atago-home/.nb/home/note1.md` contains `the first line`, `a second line`
- after `git -C .atago-home/.nb/home log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ \[nb\] Edit: note1\.md\n/`

### Scenario: deleting removes the file and records that too
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "a note" --filename note1.md
nb delete 1 --force
git -C .atago-home/.nb/home log --oneline
nb list
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "a note" --filename note1.md`:
  - exit code is `0`
- after `nb delete 1 --force`:
  - exit code is `0`
  - stdout contains `Deleted:  [1] note1.md`
  - file `.atago-home/.nb/home/note1.md` does not exist
- after `git -C .atago-home/.nb/home log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ \[nb\] Delete: note1\.md\n[0-9a-f]+ \[nb\] Add: note1\.md\n[0-9a-f]+ \[nb\] Initialize\n$/`
- after `nb list`:
  - exit code is `0`
  - stdout contains `0 items.`, `Add a note:`

### Scenario: search reads the notes rather than their names
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "buy milk and bread" --filename groceries.md && nb add --content "call the plumber" --filename chores.md
nb search "plumber"
nb search "not in any note"
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "buy milk and bread" --filename groceries.md && nb add --content "call the plumber" --filename chores.md`:
  - exit code is `0`
- after `nb search "plumber"`:
  - exit code is `0`
  - stdout contains `[2] chores.md`, `call the plumber`, does not contain `groceries.md`
- after `nb search "not in any note"`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `! Not found`

## nb (notebooks, targeting, and refusals)
How [nb](https://xwmx.github.io/nb/) keeps more than one set of notes: each
notebook is its own git repository, identifiers are scoped to the notebook
they belong to, a command can name a notebook without switching to it, and
`use` changes which one the bare commands mean.

The refusals are here too, because they are what a script meets first: an
identifier that does not exist and a subcommand that does not exist are the
same shape of answer, on stderr, at exit 1.

Source: `test/e2e/thirdparty/nb/notebooks.atago.yaml`
### Scenario: a new notebook is a new repository
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks
nb notebooks add work
git -C .atago-home/.nb/work log --oneline
nb notebooks
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks`:
  - exit code is `0`
- after `nb notebooks add work`:
  - exit code is `0`
  - stdout contains `Added notebook: work`
  - dir `.atago-home/.nb/work/.git` exists
- after `git -C .atago-home/.nb/work log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ \[nb\] Initialize\n$/`
- after `nb notebooks`:
  - exit code is `0`
  - stdout contains `home`, `work`

### Scenario: a notebook can be named in the command instead of switched to
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb notebooks add work && nb add --content "a home note" --filename home1.md
nb work:add --content "a work note" --filename work1.md
nb list
nb work:list
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb notebooks add work && nb add --content "a home note" --filename home1.md`:
  - exit code is `0`
- after `nb work:add --content "a work note" --filename work1.md`:
  - exit code is `0`
  - stdout contains `Added: [work:1] work:work1.md`
  - file `.atago-home/.nb/work/work1.md` equals exact bytes
- after `nb list`:
  - exit code is `0`
  - stdout contains `[1] home1.md`, does not contain `work1.md`
- after `nb work:list`:
  - exit code is `0`
  - stdout contains `[work:1] work1.md`

### Scenario: use changes what the bare commands mean
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb notebooks add work && nb add --content "a home note" --filename home1.md && nb work:add --content "a work note" --filename work1.md
nb use work
nb list
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb notebooks add work && nb add --content "a home note" --filename home1.md && nb work:add --content "a work note" --filename work1.md`:
  - exit code is `0`
- after `nb use work`:
  - exit code is `0`
  - stdout contains `Now using: work`
  - file `.atago-home/.nb/.current` equals exact bytes
- after `nb list`:
  - exit code is `0`
  - stdout contains `[1] work1.md`, does not contain `home1.md`

### Scenario: the same refusal arrives on two different streams
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "the only note" --filename note1.md
nb show 99
nb nosuchcommand
nb delete 99 --force
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "the only note" --filename note1.md`:
  - exit code is `0`
- after `nb show 99`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `! Not found: 99`
- after `nb nosuchcommand`:
  - exit code is `1`
  - stdout contains `! Not found: nosuchcommand`
  - stderr is empty
- after `nb delete 99 --force`:
  - exit code is `1`
  - stderr contains `! Not found: 99`
  - file `.atago-home/.nb/home/note1.md` exists

#### Generated artifacts
- `.atago-home/.nb/home/note1.md`

### Scenario: the listing is uncolored through a pipe and still carries control sequences
_only when `command -v nb` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "a note" --filename note1.md
nb list
# interactive (pty): nb list
```
#### Then
- after `git config --global user.email author@example.com && git config --global user.name 'A Author' && nb notebooks && nb add --content "a note" --filename note1.md`:
  - exit code is `0`
- after `nb list`:
  - exit code is `0`
  - stdout does not match `/\x1b\[38;5;/`
  - stdout matches `/\x1b\[\?7l/`
  - stdout contains `[1] note1.md`
- after `interactive (pty): nb list`:
  - stdout matches `/\x1b\[38;5;/`
  - rendered screen contains `[1] note1.md`
