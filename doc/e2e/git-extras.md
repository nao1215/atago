# atago Behavior Specs
## Summary
2 suites · 10 scenarios
## Contents
- [git-extras (reporting on a repository)](#git-extras-reporting-on-a-repository) — 4 scenarios
  - [summary counts the commits, the files, and the authors](#scenario-summary-counts-the-commits-the-files-and-the-authors)
  - [count reports the commits per author](#scenario-count-reports-the-commits-per-author)
  - [effort lists each file with the commits that touched it](#scenario-effort-lists-each-file-with-the-commits-that-touched-it)
  - [authors writes the file it is named after, and can just print instead](#scenario-authors-writes-the-file-it-is-named-after-and-can-just-print-instead)
- [git-extras (changing a repository)](#git-extras-changing-a-repository) — 6 scenarios
  - [a branch is created, renamed, and deleted](#scenario-a-branch-is-created-renamed-and-deleted)
  - [the merged-branch sweep keeps the branches that still have work](#scenario-the-merged-branch-sweep-keeps-the-branches-that-still-have-work)
  - [undo takes back the last commit and leaves the work in place](#scenario-undo-takes-back-the-last-commit-and-leaves-the-work-in-place)
  - [ignore adds patterns to the file git reads](#scenario-ignore-adds-patterns-to-the-file-git-reads)
  - [alias writes a git alias and lists what is defined](#scenario-alias-writes-a-git-alias-and-lists-what-is-defined)
  - [changelog writes the commits since the last tag, and writes them again](#scenario-changelog-writes-the-commits-since-the-last-tag-and-writes-them-again)

## git-extras (reporting on a repository)
[git-extras](https://github.com/tj/git-extras) is a collection of git
subcommands, each a small shell script with its own Bats file upstream. The
reporting half is pinned here: what `summary`, `count`, `effort` and
`authors` say about a repository built by the scenario that asks.

Every repository is created in the workdir, commit by commit, so the numbers
are known before the command runs: two commits over two files, one author.
Times and dates are the only thing not asserted exactly, because they are
the only thing that changes between runs.

Source: `test/e2e/thirdparty/git-extras/reporting.atago.yaml`
### Scenario: summary counts the commits, the files, and the authors
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m 'first commit' && echo two >> a.txt && echo b > b.txt && git add . && git commit -q -m 'second commit'
git summary
echo three > c.txt
git summary
# interactive (pty): git summary
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m 'first commit' && echo two >> a.txt && echo b > b.txt && git add . && git commit -q -m 'second commit'`:
  - exit code is `0`
- after `git summary`:
  - exit code is `0`
  - stdout contains ` branch:     : master`, ` commits     : 2`, ` files       : 2`, ` uncommitted : 0`
  - stdout matches `/ authors     : \n$/`
- after `echo three > c.txt`:
  - exit code is `0`
- after `git summary`:
  - exit code is `0`
  - stdout contains ` uncommitted : 1`
- after `interactive (pty): git summary`:
  - rendered screen matches `/2\s+A Author\s+100\.0%/`

### Scenario: count reports the commits per author
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m one && git commit -q --allow-empty -m two && git commit -q --allow-empty -m three
git count
git -c user.email=other@example.com -c user.name='B Other' commit -q --allow-empty -m four
git count --all
# interactive (pty): git count --all
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m one && git commit -q --allow-empty -m two && git commit -q --allow-empty -m three`:
  - exit code is `0`
- after `git count`:
  - exit code is `0`
  - stdout equals an exact value
- after `git -c user.email=other@example.com -c user.name='B Other' commit -q --allow-empty -m four`:
  - exit code is `0`
- after `git count --all`:
  - exit code is `0`
  - stdout equals an exact value
- after `interactive (pty): git count --all`:
  - rendered screen contains `A Author (3)`, `B Other (1)`, `total 4`

#### Expected output
_expected stdout:_
```text
total 3
```
_expected stdout:_
```text

total 4
```
### Scenario: effort lists each file with the commits that touched it
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m first && echo two >> a.txt && echo b > b.txt && git add . && git commit -q -m second
git effort
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m first && echo two >> a.txt && echo b > b.txt && git add . && git commit -q -m second`:
  - exit code is `0`
- after `git effort`:
  - exit code is `0`
  - stdout contains `path       commits    active days`, matches `/a\.txt\.+ 2\s+1/`
  - stdout matches `/b\.txt\.+ 1\s+1/`

### Scenario: authors writes the file it is named after, and can just print instead
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first && git -c user.email=other@example.com -c user.name='B Other' commit -q --allow-empty -m second
git authors --list
git authors
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first && git -c user.email=other@example.com -c user.name='B Other' commit -q --allow-empty -m second`:
  - exit code is `0`
- after `git authors --list`:
  - exit code is `0`
  - stdout equals an exact value
  - the step changed exactly created nothing, modified nothing, deleted nothing
- after `git authors`:
  - exit code is `0`
  - the step changed exactly created `AUTHORS`, modified nothing, deleted nothing
  - file `AUTHORS` equals exact bytes

#### Expected output
_expected stdout:_
```text
A Author <author@example.com>
B Other <other@example.com>
```
## git-extras (changing a repository)
The half of [git-extras](https://github.com/tj/git-extras) that writes:
branch creation, renaming and deletion, the merged-branch sweep, `undo`,
and the three commands that edit files or configuration — `ignore`, `alias`
and `changelog`.

Each of these is asserted on the repository afterwards rather than on what
it printed: the branch list, the git config, the working tree, and the file
contents are the oracle. Two of them are pinned as observed rather than as
intended, and say so.

Source: `test/e2e/thirdparty/git-extras/workflow.atago.yaml`
### Scenario: a branch is created, renamed, and deleted
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first
git create-branch feature/one
git branch --show-current
git rename-branch feature/two
git branch --show-current
git checkout -q master
git delete-branch feature/two
git branch --format='%(refname:short)'
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first`:
  - exit code is `0`
- after `git create-branch feature/one`:
  - exit code is `0`
- after `git branch --show-current`:
  - exit code is `0`
  - stdout equals an exact value
- after `git rename-branch feature/two`:
  - exit code is `0`
- after `git branch --show-current`:
  - exit code is `0`
  - stdout equals an exact value
- after `git checkout -q master`:
  - exit code is `0`
- after `git delete-branch feature/two`:
  - exit code is `0`
  - stderr contains `error: remote-tracking branch 'origin/feature/two' not found`
- after `git branch --format='%(refname:short)'`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
feature/one
```
_expected stdout:_
```text
feature/two
```
_expected stdout:_
```text
master
```
### Scenario: the merged-branch sweep keeps the branches that still have work
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first
git checkout -q -b merged && git commit -q --allow-empty -m 'merged work' && git checkout -q master && git merge -q merged && git checkout -q -b unmerged && git commit -q --allow-empty -m 'unmerged work' && git checkout -q master
git delete-merged-branches
git branch --format='%(refname:short)'
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first`:
  - exit code is `0`
- after `git checkout -q -b merged && git commit -q --allow-empty -m 'merged work' && git checkout -q master && git merge -q merged && git checkout -q -b unmerged && git commit -q --allow-empty -m 'unmerged work' && git checkout -q master`:
  - exit code is `0`
- after `git delete-merged-branches`:
  - exit code is `0`
  - stdout contains `Deleted branch merged`
- after `git branch --format='%(refname:short)'`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
master
unmerged
```
### Scenario: undo takes back the last commit and leaves the work in place
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m first && echo two > b.txt && git add b.txt && git commit -q -m second
git undo
git log --oneline
git status --porcelain
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m first && echo two > b.txt && git add b.txt && git commit -q -m second`:
  - exit code is `0`
- after `git undo`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted nothing, ignoring `.git/**`
- after `git log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ first\n$/`
- after `git status --porcelain`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
A  b.txt
```
### Scenario: ignore adds patterns to the file git reads
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first
git ignore "*.log"
git ignore build/
touch debug.log && git status --porcelain
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first`:
  - exit code is `0`
- after `git ignore "*.log"`:
  - exit code is `0`
  - stdout contains `Adding pattern(s) to: .gitignore`, `... adding '*.log'`
  - the step changed exactly created `.gitignore`, modified nothing, deleted nothing
- after `git ignore build/`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `.gitignore`, deleted nothing
  - file `.gitignore` equals exact bytes
- after `touch debug.log && git status --porcelain`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
?? .gitignore
```
### Scenario: alias writes a git alias and lists what is defined
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first
git alias last "log -1 HEAD --oneline"
git alias
git last
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && git commit -q --allow-empty -m first`:
  - exit code is `0`
- after `git alias last "log -1 HEAD --oneline"`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `.git/config`, deleted nothing
- after `git alias`:
  - exit code is `0`
  - stdout equals an exact value
- after `git last`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ first\n$/`

#### Expected output
_expected stdout:_
```text
last = log -1 HEAD --oneline
```
### Scenario: changelog writes the commits since the last tag, and writes them again
_only when `command -v git-summary` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m 'first release commit' && git tag v1.0.0 && echo two > b.txt && git add b.txt && git commit -q -m 'add the second thing' && echo three > c.txt && git add c.txt && git commit -q -m 'add the third thing'
git changelog
git changelog
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && echo one > a.txt && git add a.txt && git commit -q -m 'first release commit' && git tag v1.0.0 && echo two > b.txt && git add b.txt && git commit -q -m 'add the second thing' && echo three > c.txt && git add c.txt && git commit -q -m 'add the third thing'`:
  - exit code is `0`
- after `git changelog`:
  - exit code is `0`
  - the step changed exactly created `History.md`, modified nothing, deleted nothing
  - file `History.md` contains `n.n.n / `, `  * add the third thing`, `  * add the second thing`
  - file `History.md` does not contain `first release commit`
- after `git changelog`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `History.md`, deleted nothing
  - file `History.md` contains `  * add the third thing` exactly 2 times
