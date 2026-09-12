# atago Behavior Specs
## Summary
1 suite · 15 scenarios
## Contents
- [git-open (turning a remote into a web address)](#git-open-turning-a-remote-into-a-web-address) — 15 scenarios
  - [every spelling of a remote lands on the same web address \[name=ssh remote=git@github.com:example/repo.git want=https://github.com/example/repo/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namessh-remotegitgithubcomexamplerepogit-wanthttpsgithubcomexamplerepotreetopic)
  - [every spelling of a remote lands on the same web address \[name=https remote=https://github.com/example/repo.git want=https://github.com/example/repo/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namehttps-remotehttpsgithubcomexamplerepogit-wanthttpsgithubcomexamplerepotreetopic)
  - [every spelling of a remote lands on the same web address \[name=git protocol remote=git://github.com/example/repo.git want=https://github.com/example/repo/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namegit-protocol-remotegitgithubcomexamplerepogit-wanthttpsgithubcomexamplerepotreetopic)
  - [every spelling of a remote lands on the same web address \[name=ssh with a port remote=ssh://git@github.com:2222/example/repo.git want=https://github.com/example/repo/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namessh-with-a-port-remotesshgitgithubcom2222examplerepogit-wanthttpsgithubcomexamplerepotreetopic)
  - [every spelling of a remote lands on the same web address \[name=no .git suffix remote=git@github.com:example/repo want=https://github.com/example/repo/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-nameno-git-suffix-remotegitgithubcomexamplerepo-wanthttpsgithubcomexamplerepotreetopic)
  - [every spelling of a remote lands on the same web address \[name=gitlab subgroup remote=git@gitlab.com:group/subgroup/proj.git want=https://gitlab.com/group/subgroup/proj/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namegitlab-subgroup-remotegitgitlabcomgroupsubgroupprojgit-wanthttpsgitlabcomgroupsubgroupprojtreetopic)
  - [every spelling of a remote lands on the same web address \[name=bitbucket uses src instead of tree remote=git@bitbucket.org:example/repo.git want=https://bitbucket.org/example/repo/src/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namebitbucket-uses-src-instead-of-tree-remotegitbitbucketorgexamplerepogit-wanthttpsbitbucketorgexamplereposrctopic)
  - [every spelling of a remote lands on the same web address \[name=a gist remote=https://gist.github.com/abc123.git want=https://gist.github.com/abc123/tree/topic\]](#scenario-every-spelling-of-a-remote-lands-on-the-same-web-address-namea-gist-remotehttpsgistgithubcomabc123git-wanthttpsgistgithubcomabc123treetopic)
  - [the branch decides whether the page is the repository or the branch](#scenario-the-branch-decides-whether-the-page-is-the-repository-or-the-branch)
  - [a detached head opens the commit it is standing on](#scenario-a-detached-head-opens-the-commit-it-is-standing-on)
  - [the flags each choose a different page of the same repository](#scenario-the-flags-each-choose-a-different-page-of-the-same-repository)
  - [another remote, and another branch, can be named on the command line](#scenario-another-remote-and-another-branch-can-be-named-on-the-command-line)
  - [the browser is a program, and it is handed the address as an argument](#scenario-the-browser-is-a-program-and-it-is-handed-the-address-as-an-argument)
  - [a repository with no such remote still produces an address](#scenario-a-repository-with-no-such-remote-still-produces-an-address)
  - [an option it does not know is git's own usage error](#scenario-an-option-it-does-not-know-is-gits-own-usage-error)

## git-open (turning a remote into a web address)
[git-open](https://github.com/paulirish/git-open) works out the web page for
the repository you are standing in and opens it. It keeps its own test suite
in Bats; what those tests check is pinned here from outside.

Underneath the browser it is a pure translation — remote URL plus branch in,
web address out — so most of this file is that translation asserted over a
matrix of remote spellings with `--print`, which prints the address instead
of opening it. One scenario goes further and names a script as the browser,
so what the tool actually invokes, and with which argument, is proven by the
file that script writes rather than by what appeared on stdout.

Nothing here reaches the network: every repository is created in the
scenario's workdir with a remote that is never contacted.

Source: `test/e2e/thirdparty/git-open/git-open.atago.yaml`
### Scenario: every spelling of a remote lands on the same web address [name=ssh remote=git@github.com:example/repo.git want=https://github.com/example/repo/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@github.com:example/repo.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@github.com:example/repo.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo/tree/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=https remote=https://github.com/example/repo.git want=https://github.com/example/repo/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'https://github.com/example/repo.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'https://github.com/example/repo.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo/tree/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=git protocol remote=git://github.com/example/repo.git want=https://github.com/example/repo/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git://github.com/example/repo.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git://github.com/example/repo.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo/tree/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=ssh with a port remote=ssh://git@github.com:2222/example/repo.git want=https://github.com/example/repo/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'ssh://git@github.com:2222/example/repo.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'ssh://git@github.com:2222/example/repo.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo/tree/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=no .git suffix remote=git@github.com:example/repo want=https://github.com/example/repo/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@github.com:example/repo'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@github.com:example/repo'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo/tree/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=gitlab subgroup remote=git@gitlab.com:group/subgroup/proj.git want=https://gitlab.com/group/subgroup/proj/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@gitlab.com:group/subgroup/proj.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@gitlab.com:group/subgroup/proj.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://gitlab.com/group/subgroup/proj/tree/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=bitbucket uses src instead of tree remote=git@bitbucket.org:example/repo.git want=https://bitbucket.org/example/repo/src/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@bitbucket.org:example/repo.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'git@bitbucket.org:example/repo.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://bitbucket.org/example/repo/src/topic
```
### Scenario: every spelling of a remote lands on the same web address [name=a gist remote=https://gist.github.com/abc123.git want=https://gist.github.com/abc123/tree/topic]
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'https://gist.github.com/abc123.git'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git checkout -q -b topic && git remote add origin 'https://gist.github.com/abc123.git'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://gist.github.com/abc123/tree/topic
```
### Scenario: the branch decides whether the page is the repository or the branch
_only when `command -v git-open` succeeds · skipped on Windows_
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
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git
git-open --print
git checkout -q -b feature/some-work
git-open --print
git checkout -q -b 'issue#42'
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value
- after `git checkout -q -b feature/some-work`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value
- after `git checkout -q -b 'issue#42'`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo
```
_expected stdout:_
```text
https://github.com/example/repo/tree/feature/some-work
```
_expected stdout:_
```text
https://github.com/example/repo/tree/issue%2342
```
### Scenario: a detached head opens the commit it is standing on
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git && git checkout -q --detach
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git && git checkout -q --detach`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout matches `/^https://github\.com/example/repo/tree/[0-9a-f]{40}\n$/`

### Scenario: the flags each choose a different page of the same repository
_only when `command -v git-open` succeeds · skipped on Windows_
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
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && echo hello > README.md && git add README.md && git commit -q -m first && git remote add origin git@github.com:example/repo.git
git-open --print --issue
git checkout -q -b 42-fix-the-thing
git-open --print --issue
git checkout -q master
git-open --print --commit
git-open --print --file README.md
git-open --print --file missing.md
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && echo hello > README.md && git add README.md && git commit -q -m first && git remote add origin git@github.com:example/repo.git`:
  - exit code is `0`
- after `git-open --print --issue`:
  - exit code is `0`
  - stdout equals an exact value
- after `git checkout -q -b 42-fix-the-thing`:
  - exit code is `0`
- after `git-open --print --issue`:
  - exit code is `0`
  - stdout equals an exact value
- after `git checkout -q master`:
  - exit code is `0`
- after `git-open --print --commit`:
  - exit code is `0`
  - stdout matches `/^https://github\.com/example/repo/commit/[0-9a-f]{40}\n$/`
- after `git-open --print --file README.md`:
  - exit code is `0`
  - stdout equals an exact value
- after `git-open --print --file missing.md`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/repo
```
_expected stdout:_
```text
https://github.com/example/repo/issues/42
```
_expected stdout:_
```text
https://github.com/example/repo/tree/master/README.md
```
_expected stderr:_
```text
File missing.md is not in repository
```
### Scenario: another remote, and another branch, can be named on the command line
_only when `command -v git-open` succeeds · skipped on Windows_
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
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/fork.git && git remote add upstream git@gitlab.com:group/proj.git
git-open --print
git-open --print upstream
git-open --print upstream develop
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/fork.git && git remote add upstream git@gitlab.com:group/proj.git`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value
- after `git-open --print upstream`:
  - exit code is `0`
  - stdout equals an exact value
- after `git-open --print upstream develop`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://github.com/example/fork
```
_expected stdout:_
```text
https://gitlab.com/group/proj
```
_expected stdout:_
```text
https://gitlab.com/group/proj/tree/develop
```
### Scenario: the browser is a program, and it is handed the address as an argument
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `fake-browser.sh` is created.
- Environment variables are set: BROWSER, LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `fake-browser.sh`:_
```text
#!/bin/sh
printf '%s\n' "argc=$#" "argv1=$1" >> opened.log
```
#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git
git-open
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git`:
  - exit code is `0`
- after `git-open`:
  - exit code is `0`
  - stdout is empty
  - the step changed exactly created `opened.log`, modified nothing, deleted nothing
  - file `opened.log` equals exact bytes

### Scenario: a repository with no such remote still produces an address
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first
git-open --print
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first`:
  - exit code is `0`
- after `git-open --print`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
https://origin/origin
```
### Scenario: an option it does not know is git's own usage error
_only when `command -v git-open` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git
git-open --bogus
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git commit -q --allow-empty -m first && git remote add origin git@github.com:example/repo.git`:
  - exit code is `0`
- after `git-open --bogus`:
  - exit code is `129`
  - stdout is empty
  - stderr contains `` error: unknown option `bogus' ``, `usage: git open [options]`, `-p, --print           just print the url`
