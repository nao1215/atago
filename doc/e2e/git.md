# atago Behavior Specs
## Summary
4 suites · 20 scenarios
## Contents
- [git + changes (a staged blob touches exactly index + one object)](#git--changes-a-staged-blob-touches-exactly-index--one-object) — 2 scenarios
  - [staging one file creates the index and a single loose object (POSIX)](#scenario-staging-one-file-creates-the-index-and-a-single-loose-object-posix)
  - [git init's whole .git tree is pinned by one recursive glob (POSIX)](#scenario-git-inits-whole-git-tree-is-pinned-by-one-recursive-glob-posix)
- [git (third-party CLI, no build required)](#git-third-party-cli-no-build-required) — 5 scenarios
  - [init creates an empty repository](#scenario-init-creates-an-empty-repository)
  - [add and commit make the working tree clean](#scenario-add-and-commit-make-the-working-tree-clean)
  - [a captured commit hash flows into a later command](#scenario-a-captured-commit-hash-flows-into-a-later-command)
  - [checking out a missing ref fails with an explanation (no-such-branch)](#scenario-checking-out-a-missing-ref-fails-with-an-explanation-no-such-branch)
  - [checking out a missing ref fails with an explanation (v9.9.9)](#scenario-checking-out-a-missing-ref-fails-with-an-explanation-v999)
- [git + sandbox_home (global config in an isolated HOME)](#git--sandbox_home-global-config-in-an-isolated-home) — 1 scenario
  - [global user.name is written under the sandbox home and read back (POSIX)](#scenario-global-username-is-written-under-the-sandbox-home-and-read-back-posix)
- [git (branching, diffs, and failure contracts)](#git-branching-diffs-and-failure-contracts) — 12 scenarios
  - [diff --exit-code reports difference through the exit status](#scenario-diff---exit-code-reports-difference-through-the-exit-status)
  - [status --porcelain spells each state exactly](#scenario-status---porcelain-spells-each-state-exactly)
  - [a diverged branch fast-forwards into the base](#scenario-a-diverged-branch-fast-forwards-into-the-base)
  - [a conflicting merge stops with markers and a non-zero status](#scenario-a-conflicting-merge-stops-with-markers-and-a-non-zero-status)
  - [stash and pop restore the working tree byte for byte](#scenario-stash-and-pop-restore-the-working-tree-byte-for-byte)
  - [a diff applied back restores the discarded change](#scenario-a-diff-applied-back-restores-the-discarded-change)
  - [an ignored file is never staged](#scenario-an-ignored-file-is-never-staged)
  - [a local clone reproduces the same commit](#scenario-a-local-clone-reproduces-the-same-commit)
  - [committing with nothing staged fails and creates no commit](#scenario-committing-with-nothing-staged-fails-and-creates-no-commit)
  - [a duplicate tag is refused and the original is untouched](#scenario-a-duplicate-tag-is-refused-and-the-original-is-untouched)
  - [git outside a repository fails with its own exit code](#scenario-git-outside-a-repository-fails-with-its-own-exit-code)
  - [identical commits made twice hash the same](#scenario-identical-commits-made-twice-hash-the-same)

## git + changes (a staged blob touches exactly index + one object)
`git add` is documented as writing a blob and updating the index. This suite
holds git to that literally: after initializing a repository, staging one
file must create exactly two things on disk — the staging index, and one
loose object under `.git/objects` — and nothing else anywhere.

It is an exhaustive claim, not a spot check. Any extra file git wrote would
fail the assertion, which is what makes it a real statement about git's
footprint rather than a restatement of the documentation.

Source: `test/e2e/thirdparty/git/changes.atago.yaml`
### Scenario: staging one file creates the index and a single loose object (POSIX)
_skipped on Windows_
#### Given
- Fixture file `repo/f.txt` is created.

#### Inputs
_Fixture `repo/f.txt`:_
```text
hello from atago
```
#### When
```shell
git init -q repo
git -C repo add f.txt
```
#### Then
- after `git init -q repo`:
  - exit code is `0`
- after `git -C repo add f.txt`:
  - exit code is `0`
  - the step changed exactly created `repo/.git/index`, `repo/.git/objects/*/*`, modified nothing, deleted nothing

### Scenario: git init's whole .git tree is pinned by one recursive glob (POSIX)
_skipped on Windows_
#### When
```shell
git init -q repo
```
#### Then
- exit code is `0`
- the step changed exactly created `repo/.git/**`, modified nothing, deleted nothing

## git (third-party CLI, no build required)
The proof that testing someone else's CLI takes no cooperation from it. git
was not written with this test suite in mind, ships no hooks for it, and
exposes nothing but a command line — and that is enough.

What is guaranteed here is ordinary git as a script would meet it: a
repository is initialized, files are staged and committed, history and
status report the expected state, branches and tags behave, and the failure
cases (a bad revision, a dirty tree) exit non-zero with a message.

Every command uses git's own `-C`/`-c` flags instead of a shell, so this
exact spec runs unchanged on Linux, macOS, and Windows.

Source: `test/e2e/thirdparty/git/git.atago.yaml`
### Scenario: init creates an empty repository
#### When
```shell
git init -q repo
git -C repo rev-parse HEAD
```
#### Then
- after `git init -q repo`:
  - exit code is `0`
  - file `repo/.git/HEAD` contains `ref`
- after `git -C repo rev-parse HEAD`:
  - exit code is not `0`

### Scenario: add and commit make the working tree clean
#### Given
- Fixture file `repo-src/hello.txt` is created.

#### Inputs
_Fixture `repo-src/hello.txt`:_
```text
hello from atago
```
#### When
```shell
git init -q repo-src
git -C repo-src add hello.txt
git -C repo-src -c user.name=atago -c user.email=atago@example.com commit -q -m "add hello"
git -C repo-src status --porcelain
git -C repo-src log --oneline
```
#### Then
- after `git -C repo-src -c user.name=atago -c user.email=atago@example.com commit -q -m "add hello"`:
  - exit code is `0`
- after `git -C repo-src status --porcelain`:
  - exit code is `0`
  - stdout is empty
- after `git -C repo-src log --oneline`:
  - stdout contains `add hello`

### Scenario: a captured commit hash flows into a later command
#### Given
- Fixture file `r/f.txt` is created.

#### Inputs
_Fixture `r/f.txt`:_
```text
v1
```
#### When
```shell
git init -q r
git -C r add f.txt
git -C r -c user.name=atago -c user.email=atago@example.com commit -q -m "first"
git -C r rev-parse HEAD
# capture ${head} from stdout
git -C r show --no-patch --format=%s ${head}
```
#### Then
- after `git -C r show --no-patch --format=%s ${head}`:
  - exit code is `0`
  - stdout contains `first`

### Scenario: checking out a missing ref fails with an explanation (no-such-branch)
#### When
```shell
git init -q repo
git -C repo checkout no-such-branch
```
#### Then
- after `git -C repo checkout no-such-branch`:
  - exit code is not `0`
  - stderr contains `no-such-branch`

### Scenario: checking out a missing ref fails with an explanation (v9.9.9)
#### When
```shell
git init -q repo
git -C repo checkout v9.9.9
```
#### Then
- after `git -C repo checkout v9.9.9`:
  - exit code is not `0`
  - stderr contains `v9.9.9`

## git + sandbox_home (global config in an isolated HOME)
`git config --global` writes to the user's home directory — the one thing a
test must never do to the machine running it. This suite shows the
alternative: the command runs with its home redirected inside the scenario's
own workspace, so a genuine `--global` write happens, lands at a
deterministic path, and disappears with the workspace.

The isolated home persists across steps, not just within one: a later
command reads the same setting back, proving it is a real home directory
rather than a write that went nowhere.

Source: `test/e2e/thirdparty/git/sandbox_home.atago.yaml`
### Scenario: global user.name is written under the sandbox home and read back (POSIX)
_skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git config --global user.name atago-sandbox-user
git config --global user.name
```
#### Then
- after `git config --global user.name atago-sandbox-user`:
  - exit code is `0`
  - file `.atago-home/.gitconfig` contains `atago-sandbox-user`
- after `git config --global user.name`:
  - exit code is `0`
  - stdout equals an exact value

## git (branching, diffs, and failure contracts)
The second half of the git suite: the parts of git a script actually leans
on, driven as state transitions rather than single commands.

A branch is created, diverged, merged fast-forward, and then merged into a
real conflict; a stash and an `apply`ed patch are proven to restore the file
byte for byte; `--porcelain` and `diff --exit-code` are pinned as the
machine contracts they are; and each documented failure — no repository, a
duplicate tag, nothing staged, a conflicted merge — is pinned to its own
exit code with the message that explains it. Every command uses git's own
`-C`/`-c` flags, so no shell is involved except where a redirect is the
thing under test.

Source: `test/e2e/thirdparty/git/workflow.atago.yaml`
### Scenario: diff --exit-code reports difference through the exit status
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/a.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
one
```
_Fixture `repo/a.txt`:_
```text
one
two
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first
git -C repo diff --exit-code
git -C repo diff --exit-code
```
#### Then
- after `git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first`:
  - exit code is `0`
- after `git -C repo diff --exit-code`:
  - exit code is `0`
  - stdout is empty
- after `git -C repo diff --exit-code`:
  - exit code is `1`
  - stdout contains `diff --git a/a.txt b/a.txt`, `+two`

### Scenario: status --porcelain spells each state exactly
#### Given
- Fixture file `repo/tracked.txt` is created.
- Fixture file `repo/tracked.txt` is created.
- Fixture file `repo/untracked.txt` is created.
- Fixture file `repo/staged.txt` is created.

#### Inputs
_Fixture `repo/tracked.txt`:_
```text
v1
```
_Fixture `repo/tracked.txt`:_
```text
v2
```
_Fixture `repo/untracked.txt`:_
```text
new
```
_Fixture `repo/staged.txt`:_
```text
staged
```
#### When
```shell
git init -q repo
git -C repo add tracked.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first
git -C repo add staged.txt
git -C repo status --porcelain
```
#### Then
- after `git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first`:
  - exit code is `0`
- after `git -C repo add staged.txt`:
  - exit code is `0`
- after `git -C repo status --porcelain`:
  - exit code is `0`
  - stdout line `1` equals an exact value
  - stdout line `2` equals an exact value
  - stdout line `3` equals an exact value

### Scenario: a diverged branch fast-forwards into the base
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/b.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
base
```
_Fixture `repo/b.txt`:_
```text
feature
```
#### When
```shell
git init -q -b main repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m base
git -C repo checkout -q -b feature
git -C repo add b.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m feature
git -C repo checkout -q main
git -C repo merge --ff-only feature
git -C repo rev-parse main
# capture ${main_head} from stdout
git -C repo rev-parse feature
```
#### Then
- after `git -C repo checkout -q -b feature`:
  - exit code is `0`
- after `git -C repo merge --ff-only feature`:
  - exit code is `0`
- after `git -C repo rev-parse feature`:
  - exit code is `0`
  - stdout contains `${main_head}`
  - file `repo/b.txt` contains `feature`

### Scenario: a conflicting merge stops with markers and a non-zero status
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/a.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
base
```
_Fixture `repo/a.txt`:_
```text
from other
```
_Fixture `repo/a.txt`:_
```text
from main
```
#### When
```shell
git init -q -b main repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m base
git -C repo checkout -q -b other
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m other
git -C repo checkout -q main
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m main
git -C repo -c user.name=atago -c user.email=atago@example.com merge other
git -C repo status --porcelain
```
#### Then
- after `git -C repo -c user.name=atago -c user.email=atago@example.com merge other`:
  - exit code is `1`
  - stdout contains `CONFLICT`
  - file `repo/a.txt` contains `<<<<<<<`, `from main`, `from other`, `>>>>>>>`
- after `git -C repo status --porcelain`:
  - stdout contains `UU a.txt`

### Scenario: stash and pop restore the working tree byte for byte
_skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/a.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
committed
```
_Fixture `repo/a.txt`:_
```text
work in progress	tab and trailing space 
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first
cp repo/a.txt expected.txt
git -C repo -c user.name=atago -c user.email=atago@example.com stash
git -C repo status --porcelain
git -C repo stash pop
```
#### Then
- after `git -C repo -c user.name=atago -c user.email=atago@example.com stash`:
  - exit code is `0`
  - file `repo/a.txt` equals exact bytes
- after `git -C repo status --porcelain`:
  - stdout is empty
- after `git -C repo stash pop`:
  - exit code is `0`
  - file `repo/a.txt` is byte-identical to `expected.txt`

### Scenario: a diff applied back restores the discarded change
_skipped on Windows_
#### Given
- Fixture file `repo/a.txt` is created.
- Fixture file `repo/a.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
line1
```
_Fixture `repo/a.txt`:_
```text
line1
line2
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first
cp repo/a.txt expected.txt
git -C repo diff > change.patch
git -C repo checkout -- a.txt
git -C repo apply ../change.patch
```
#### Then
- after `git -C repo diff > change.patch`:
  - exit code is `0`
  - file `change.patch` contains `+line2`
- after `git -C repo checkout -- a.txt`:
  - exit code is `0`
  - file `repo/a.txt` equals exact bytes
- after `git -C repo apply ../change.patch`:
  - exit code is `0`
  - file `repo/a.txt` is byte-identical to `expected.txt`

### Scenario: an ignored file is never staged
#### Given
- Fixture file `repo/.gitignore` is created.
- Fixture file `repo/keep.txt` is created.
- Fixture file `repo/debug.log` is created.
- Fixture file `repo/build/out.bin` is created.

#### Inputs
_Fixture `repo/.gitignore`:_
```text
*.log
build/
```
_Fixture `repo/keep.txt`:_
```text
keep
```
_Fixture `repo/debug.log`:_
```text
noise
```
_Fixture `repo/build/out.bin`:_
```text
artifact
```
#### When
```shell
git init -q repo
git -C repo add .
git -C repo status --porcelain
git -C repo check-ignore -v debug.log
```
#### Then
- after `git -C repo add .`:
  - exit code is `0`
- after `git -C repo status --porcelain`:
  - exit code is `0`
  - stdout contains `A  .gitignore`, `A  keep.txt`
  - stdout does not contain `debug.log`, `build/out.bin`
- after `git -C repo check-ignore -v debug.log`:
  - exit code is `0`
  - stdout contains `*.log`

### Scenario: a local clone reproduces the same commit
#### Given
- Fixture file `origin/a.txt` is created.

#### Inputs
_Fixture `origin/a.txt`:_
```text
shared
```
#### When
```shell
git init -q origin
git -C origin add a.txt
git -C origin -c user.name=atago -c user.email=atago@example.com commit -q -m shared
git -C origin rev-parse HEAD
# capture ${origin_head} from stdout
git -c core.autocrlf=false clone -q origin copy
git -C copy rev-parse HEAD
```
#### Then
- after `git -c core.autocrlf=false clone -q origin copy`:
  - exit code is `0`
- after `git -C copy rev-parse HEAD`:
  - exit code is `0`
  - stdout contains `${origin_head}`
  - file `copy/a.txt` equals exact bytes

### Scenario: committing with nothing staged fails and creates no commit
#### Given
- Fixture file `repo/a.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
one
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first
git -C repo rev-parse HEAD
# capture ${head_before} from stdout
git -C repo -c user.name=atago -c user.email=atago@example.com commit -m "empty"
git -C repo rev-parse HEAD
```
#### Then
- after `git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first`:
  - exit code is `0`
- after `git -C repo -c user.name=atago -c user.email=atago@example.com commit -m "empty"`:
  - exit code is `1`
  - stdout contains `nothing to commit`
- after `git -C repo rev-parse HEAD`:
  - exit code is `0`
  - stdout contains `${head_before}`

### Scenario: a duplicate tag is refused and the original is untouched
#### Given
- Fixture file `repo/a.txt` is created.

#### Inputs
_Fixture `repo/a.txt`:_
```text
one
```
#### When
```shell
git init -q repo
git -C repo add a.txt
git -C repo -c user.name=atago -c user.email=atago@example.com commit -q -m first
git -C repo tag v1
git -C repo tag v1
git -C repo tag --list
```
#### Then
- after `git -C repo tag v1`:
  - exit code is `0`
- after `git -C repo tag v1`:
  - exit code is `128`
  - stderr contains `already exists`
- after `git -C repo tag --list`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
v1
```
### Scenario: git outside a repository fails with its own exit code
#### When
```shell
git status
```
#### Then
- exit code is `128`
- stdout is empty
- stderr contains `not a git repository`

### Scenario: identical commits made twice hash the same
#### Given
- Fixture file `one/a.txt` is created.
- Fixture file `two/a.txt` is created.
- Environment variables are set: GIT_AUTHOR_DATE, GIT_COMMITTER_DATE.
- Environment variables are set: GIT_AUTHOR_DATE, GIT_COMMITTER_DATE.

#### Inputs
_Fixture `one/a.txt`:_
```text
same bytes
```
_Fixture `two/a.txt`:_
```text
same bytes
```
#### When
```shell
git init -q one
git init -q two
git -C one add a.txt
git -C two add a.txt
git -C one -c user.name=atago -c user.email=atago@example.com commit -q -m same
git -C two -c user.name=atago -c user.email=atago@example.com commit -q -m same
git -C one rev-parse HEAD
# capture ${one_head} from stdout
git -C two rev-parse HEAD
```
#### Then
- after `git -C two rev-parse HEAD`:
  - exit code is `0`
  - stdout contains `${one_head}`
