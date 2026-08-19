# atago Behavior Specs
## Summary
2 suites · 13 scenarios
## Contents
- [git-secrets (patterns, scanning, and allow-listing)](#git-secrets-patterns-scanning-and-allow-listing) — 8 scenarios
  - [a pattern is stored in the repository's git config](#scenario-a-pattern-is-stored-in-the-repositorys-git-config)
  - [a clean file passes silently and a match is reported on stderr](#scenario-a-clean-file-passes-silently-and-a-match-is-reported-on-stderr)
  - [a false positive is allowed either in the config or in a file](#scenario-a-false-positive-is-allowed-either-in-the-config-or-in-a-file)
  - [a literal pattern is escaped so its metacharacters mean themselves](#scenario-a-literal-pattern-is-escaped-so-its-metacharacters-mean-themselves)
  - [scanning a directory needs the recursive flag](#scenario-scanning-a-directory-needs-the-recursive-flag)
  - [a file that is not there is exit 2, not a clean scan](#scenario-a-file-that-is-not-there-is-exit-2-not-a-clean-scan)
  - [a command line git rejects is 129 with the usage](#scenario-a-command-line-git-rejects-is-129-with-the-usage)
  - [registering the AWS rules adds the provider and its patterns](#scenario-registering-the-aws-rules-adds-the-provider-and-its-patterns)
- [git-secrets (hooks, blocked commits, and history)](#git-secrets-hooks-blocked-commits-and-history) — 5 scenarios
  - [installing writes the three hooks that call back into the tool](#scenario-installing-writes-the-three-hooks-that-call-back-into-the-tool)
  - [the installer reports a failure it did not have](#scenario-the-installer-reports-a-failure-it-did-not-have)
  - [a commit carrying the pattern never happens](#scenario-a-commit-carrying-the-pattern-never-happens)
  - [a commit message carrying the pattern is refused too](#scenario-a-commit-message-carrying-the-pattern-is-refused-too)
  - [what the bypass lets through, the history scan finds](#scenario-what-the-bypass-lets-through-the-history-scan-finds)

## git-secrets (patterns, scanning, and allow-listing)
[git-secrets](https://github.com/awslabs/git-secrets) scans files for
patterns that must never be committed. It keeps its own test suite in Bats;
what those tests check is pinned here from outside, against repositories the
scenarios create and throw away.

No real credential appears anywhere in this suite. The patterns are invented
for the scenario that uses them, which is enough to exercise every rule the
tool has: where a pattern is stored, what a match looks like, which stream it
goes to, the two ways to allow a false positive, and the exit codes — 0 for
clean, 1 for a match, 2 when the file is not there, 129 for a command line
git itself rejects.

Source: `test/e2e/thirdparty/git-secrets/git-secrets.atago.yaml`
### Scenario: a pattern is stored in the repository's git config
_only when `command -v git-secrets` succeeds · skipped on Windows_
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
git init -q && git config user.email atago@example.com && git config user.name atago
git secrets --list
git secrets --add topsecret-[0-9]+
git secrets --list
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago`:
  - exit code is `0`
- after `git secrets --list`:
  - exit code is `1`
  - stdout is empty
- after `git secrets --add topsecret-[0-9]+`:
  - exit code is `0`
  - the step changed exactly created nothing, modified `.git/config`, deleted nothing
  - file `.git/config` contains `patterns = topsecret-[0-9]+`
- after `git secrets --list`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
secrets.patterns topsecret-[0-9]+
```
### Scenario: a clean file passes silently and a match is reported on stderr
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `clean.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `leaky.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `clean.txt`:_
```text
nothing of interest here
```
_Fixture `leaky.txt`:_
```text
first line is fine
token = topsecret-4242
```
#### When
```shell
git init -q && git secrets --add topsecret-[0-9]+
git secrets --scan clean.txt
git secrets --scan leaky.txt
```
#### Then
- after `git init -q && git secrets --add topsecret-[0-9]+`:
  - exit code is `0`
- after `git secrets --scan clean.txt`:
  - exit code is `0`
  - stdout is empty
  - stderr is empty
- after `git secrets --scan leaky.txt`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `leaky.txt:2:token = topsecret-4242`, `[ERROR] Matched one or more prohibited patterns`, `- Mark false positives as allowed using: git config --add secrets.allowed ...`, `- Use --no-verify if this is a one-time false positive`

### Scenario: a false positive is allowed either in the config or in a file
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `leaky.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.gitallowed` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `leaky.txt`:_
```text
token = topsecret-4242
```
_Fixture `.gitallowed`:_
```text
topsecret-42[0-9][0-9]
```
#### When
```shell
git init -q && git secrets --add topsecret-[0-9]+
git secrets --scan leaky.txt
git secrets --add --allowed topsecret-4242
git secrets --scan leaky.txt
git config --unset-all secrets.allowed
git secrets --scan leaky.txt
```
#### Then
- after `git init -q && git secrets --add topsecret-[0-9]+`:
  - exit code is `0`
- after `git secrets --scan leaky.txt`:
  - exit code is `1`
- after `git secrets --add --allowed topsecret-4242`:
  - exit code is `0`
- after `git secrets --scan leaky.txt`:
  - exit code is `0`
  - stderr is empty
- after `git config --unset-all secrets.allowed`:
  - exit code is `0`
- after `git secrets --scan leaky.txt`:
  - exit code is `0`
  - stderr is empty

### Scenario: a literal pattern is escaped so its metacharacters mean themselves
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `literal.txt` is created.
- Fixture file `regexy.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `literal.txt`:_
```text
a.b*c
```
_Fixture `regexy.txt`:_
```text
aXbbbc
```
#### When
```shell
git init -q
git secrets --add --literal a.b*c
git secrets --list
git secrets --scan literal.txt
git secrets --scan regexy.txt
```
#### Then
- after `git init -q`:
  - exit code is `0`
- after `git secrets --add --literal a.b*c`:
  - exit code is `0`
- after `git secrets --list`:
  - exit code is `0`
  - stdout equals an exact value
- after `git secrets --scan literal.txt`:
  - exit code is `1`
  - stderr contains `literal.txt:1:a.b*c`
- after `git secrets --scan regexy.txt`:
  - exit code is `0`
  - stderr is empty

#### Expected output
_expected stdout:_
```text
secrets.patterns a\.b\*c
```
### Scenario: scanning a directory needs the recursive flag
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `src/nested/leaky.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `src/nested/leaky.txt`:_
```text
token = topsecret-1
```
#### When
```shell
git init -q && git secrets --add topsecret-[0-9]+
git secrets --scan src
git secrets --scan -r src
```
#### Then
- after `git init -q && git secrets --add topsecret-[0-9]+`:
  - exit code is `0`
- after `git secrets --scan src`:
  - exit code is `0`
  - stdout is empty
  - stderr is empty
- after `git secrets --scan -r src`:
  - exit code is `1`
  - stderr contains `src/nested/leaky.txt:1:token = topsecret-1`

### Scenario: a file that is not there is exit 2, not a clean scan
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git secrets --add topsecret-[0-9]+
git secrets --scan missing.txt
```
#### Then
- after `git init -q && git secrets --add topsecret-[0-9]+`:
  - exit code is `0`
- after `git secrets --scan missing.txt`:
  - exit code is `2`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
grep: missing.txt: No such file or directory
```
### Scenario: a command line git rejects is 129 with the usage
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q
git secrets --bogus
```
#### Then
- after `git init -q`:
  - exit code is `0`
- after `git secrets --bogus`:
  - exit code is `129`
  - stdout is empty
  - stderr contains `` error: unknown option `bogus' ``, `usage: git secrets --scan [-r|--recursive] [--cached] [--no-index] [--untracked] [<files>...]`

### Scenario: registering the AWS rules adds the provider and its patterns
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q
git secrets --register-aws
git secrets --list
```
#### Then
- after `git init -q`:
  - exit code is `0`
- after `git secrets --register-aws`:
  - exit code is `0`
  - stdout equals an exact value
  - the step changed exactly created nothing, modified `.git/config`, deleted nothing
- after `git secrets --list`:
  - exit code is `0`
  - stdout contains `secrets.providers git secrets --aws-provider`, `[A-Z0-9]{16}`

#### Expected output
_expected stdout:_
```text
OK
```
## git-secrets (hooks, blocked commits, and history)
The half of [git-secrets](https://github.com/awslabs/git-secrets) that runs
without being asked: the hooks it installs into a repository, the commit
they refuse, the message they read, the bypass that gets past them, and the
history scan that finds what the bypass let through.

A blocked commit is asserted as a commit that does not exist — the log is
the oracle, not the error text — and the bypass scenario carries it through
to the end: the secret reaches history, and `--scan-history` is what finds
it there, quoting the commit it lives in.

Source: `test/e2e/thirdparty/git-secrets/hooks.atago.yaml`
### Scenario: installing writes the three hooks that call back into the tool
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git secrets --add topsecret-[0-9]+
git secrets --install
git secrets --install
```
#### Then
- after `git init -q && git secrets --add topsecret-[0-9]+`:
  - exit code is `0`
- after `git secrets --install`:
  - the step changed exactly created `.git/hooks/commit-msg`, `.git/hooks/pre-commit`, `.git/hooks/prepare-commit-msg`, modified nothing, deleted nothing
  - file `.git/hooks/pre-commit` equals exact bytes
  - file `.git/hooks/pre-commit` is executable
- after `git secrets --install`:
  - exit code is `1`
  - stderr matches `/\.git/hooks/commit-msg already exists\. Use -f to force/`
  - the step changed exactly created nothing, modified nothing, deleted nothing

### Scenario: the installer reports a failure it did not have
_only when `command -v git-secrets` succeeds · skipped on macOS_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q
git secrets --install
```
#### Then
- after `git init -q`:
  - exit code is `0`
- after `git secrets --install`:
  - exit code is `127`
  - stderr contains `say: command not found`
  - dir `.git/hooks` contains `commit-msg`, contains `pre-commit`, contains `prepare-commit-msg`

### Scenario: a commit carrying the pattern never happens
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `clean.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `leaky.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `clean.txt`:_
```text
nothing of interest here
```
_Fixture `leaky.txt`:_
```text
token = topsecret-4242
```
#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git secrets --add topsecret-[0-9]+ && { git secrets --install --force || true; } && test -x .git/hooks/pre-commit
git add clean.txt && git commit -q -m "first commit"
git add leaky.txt && git commit -m "add the file"
git log --oneline
git status --porcelain
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git secrets --add topsecret-[0-9]+ && { git secrets --install --force || true; } && test -x .git/hooks/pre-commit`:
  - exit code is `0`
- after `git add clean.txt && git commit -q -m "first commit"`:
  - exit code is `0`
- after `git add leaky.txt && git commit -m "add the file"`:
  - exit code is `1`
  - stderr contains `leaky.txt:1:token = topsecret-4242`, `[ERROR] Matched one or more prohibited patterns`
- after `git log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ first commit\n$/`
- after `git status --porcelain`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
A  leaky.txt
```
### Scenario: a commit message carrying the pattern is refused too
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `clean.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `clean.txt`:_
```text
nothing of interest here
```
#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git secrets --add topsecret-[0-9]+ && { git secrets --install --force || true; } && test -x .git/hooks/pre-commit
git add clean.txt && git commit -m "rotating topsecret-77 today"
git log --oneline
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git secrets --add topsecret-[0-9]+ && { git secrets --install --force || true; } && test -x .git/hooks/pre-commit`:
  - exit code is `0`
- after `git add clean.txt && git commit -m "rotating topsecret-77 today"`:
  - exit code is `1`
  - stderr contains `.git/COMMIT_EDITMSG:1:rotating topsecret-77 today`, `[ERROR] Matched one or more prohibited patterns`
- after `git log --oneline`:
  - exit code is `128`
  - stdout is empty
  - stderr contains `does not have any commits yet`

### Scenario: what the bypass lets through, the history scan finds
_only when `command -v git-secrets` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `leaky.txt` is created.
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL, TERM.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `leaky.txt`:_
```text
token = topsecret-4242
```
#### When
```shell
git init -q && git config user.email atago@example.com && git config user.name atago && git secrets --add topsecret-[0-9]+ && { git secrets --install --force || true; } && test -x .git/hooks/pre-commit
git add leaky.txt && git commit -q --no-verify -m "sneak it in"
git log --oneline
git secrets --scan-history
git secrets --scan-history
```
#### Then
- after `git init -q && git config user.email atago@example.com && git config user.name atago && git secrets --add topsecret-[0-9]+ && { git secrets --install --force || true; } && test -x .git/hooks/pre-commit`:
  - exit code is `0`
- after `git add leaky.txt && git commit -q --no-verify -m "sneak it in"`:
  - exit code is `0`
- after `git log --oneline`:
  - exit code is `0`
  - stdout matches `/^[0-9a-f]+ sneak it in\n$/`
- after `git secrets --scan-history`:
  - exit code is `1`
  - stderr matches `/^[0-9a-f]{40}:leaky\.txt:1:token = topsecret-4242/`
- after `git secrets --scan-history`:
  - exit code is `1`
