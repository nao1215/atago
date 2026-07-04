# atago Behavior Specs
## Summary
3 suites · 8 scenarios
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
## git + changes (a staged blob touches exactly index + one object)
Source: `test/e2e/thirdparty/git/changes.atago.yaml`
### Scenario: staging one file creates the index and a single loose object (POSIX)
_skipped on windows_
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
_skipped on windows_
#### When
```shell
git init -q repo
```
#### Then
- exit code is `0`
- the step changed exactly created `repo/.git/**`, modified nothing, deleted nothing
## git (third-party CLI, no build required)
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
Source: `test/e2e/thirdparty/git/sandbox_home.atago.yaml`
### Scenario: global user.name is written under the sandbox home and read back (POSIX)
_skipped on windows_
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