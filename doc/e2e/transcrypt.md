# atago Behavior Specs
## Summary
2 suites · 9 scenarios
## Contents
- [transcrypt (plaintext in the tree, ciphertext in the repository)](#transcrypt-plaintext-in-the-tree-ciphertext-in-the-repository) — 5 scenarios
  - [it refuses to work outside a repository, and says so before anything else](#scenario-it-refuses-to-work-outside-a-repository-and-says-so-before-anything-else)
  - [initializing writes the filter configuration and nothing else](#scenario-initializing-writes-the-filter-configuration-and-nothing-else)
  - [the working tree keeps the plaintext and the commit holds the ciphertext](#scenario-the-working-tree-keeps-the-plaintext-and-the-commit-holds-the-ciphertext)
  - [re-saving the same content produces no change at all](#scenario-re-saving-the-same-content-produces-no-change-at-all)
  - [two files with the same content do not encrypt to the same bytes](#scenario-two-files-with-the-same-content-do-not-encrypt-to-the-same-bytes)
- [transcrypt (cloning, the wrong password, and uninstalling)](#transcrypt-cloning-the-wrong-password-and-uninstalling) — 4 scenarios
  - [a clone carries ciphertext until it is given the password](#scenario-a-clone-carries-ciphertext-until-it-is-given-the-password)
  - [the wrong password is caught rather than silently producing garbage](#scenario-the-wrong-password-is-caught-rather-than-silently-producing-garbage)
  - [uninstalling asks first, and leaves the plaintext behind when told to go ahead](#scenario-uninstalling-asks-first-and-leaves-the-plaintext-behind-when-told-to-go-ahead)
  - [answering no to the uninstall changes nothing](#scenario-answering-no-to-the-uninstall-changes-nothing)

## transcrypt (plaintext in the tree, ciphertext in the repository)
[transcrypt](https://github.com/elasticdog/transcrypt) encrypts chosen files
inside a git repository: you keep editing plaintext, and what git stores is
ciphertext. It keeps its own test suite in Bats; what those tests check is
pinned here from outside, against repositories the scenarios create.

The claim is a difference between two views of the same file, so that is
what is asserted: the working tree on one side, `git show HEAD:<file>` on
the other. Around it are the properties that make the scheme usable — the
same content re-saved produces no diff at all, a change produces a diff in
plaintext, and two files with the same content do not encrypt to the same
bytes — and the ones that make it safe, above all a wrong password being
detected rather than quietly producing garbage.

Every password here is invented for the scenario that uses it.

Source: `test/e2e/thirdparty/transcrypt/transcrypt.atago.yaml`
### Scenario: it refuses to work outside a repository, and says so before anything else
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
transcrypt --display
transcrypt --version
```
#### Then
- after `transcrypt --display`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `transcrypt: you are not currently in a git repository; did you forget to run "git init"?`
- after `transcrypt --version`:
  - exit code is `0`
  - stdout matches `/^transcrypt [0-9]+\.[0-9]+/`

### Scenario: initializing writes the filter configuration and nothing else
_only when `command -v transcrypt` succeeds · skipped on Windows_
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
git init -q && git config user.email author@example.com && git config user.name 'A Author'
transcrypt --cipher aes-256-cbc --password test-password-one --yes
git config --get transcrypt.cipher
transcrypt --display
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author'`:
  - exit code is `0`
- after `transcrypt --cipher aes-256-cbc --password test-password-one --yes`:
  - exit code is `0`
  - stdout equals an exact value
  - the step changed exactly created `.gitattributes`, modified nothing, deleted nothing, ignoring `.git/**`
  - file `.gitattributes` equals exact bytes
- after `git config --get transcrypt.cipher`:
  - exit code is `0`
  - stdout equals an exact value
- after `transcrypt --display`:
  - exit code is `0`
  - stdout contains `CIPHER:   aes-256-cbc`, `PASSWORD: test-password-one`, `  transcrypt -c aes-256-cbc -p 'test-password-one'`

#### Expected output
_expected stdout:_
```text
The repository has been successfully configured by transcrypt.
```
_expected stdout:_
```text
aes-256-cbc
```
### Scenario: the working tree keeps the plaintext and the commit holds the ciphertext
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.gitattributes` is created.
- Fixture file `secret.txt` is created.
- Fixture file `public.txt` is created.
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

#### Inputs
_Fixture `.gitattributes`:_
```text
secret.txt filter=crypt diff=crypt merge=crypt
```
_Fixture `secret.txt`:_
```text
api_key = not-a-real-key
```
_Fixture `public.txt`:_
```text
nothing secret here
```
#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p test-password-two -y
git add .gitattributes secret.txt public.txt && git commit -q -m 'add both files'
git show HEAD:secret.txt
git show HEAD:public.txt
transcrypt --list
transcrypt --show-raw secret.txt
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p test-password-two -y`:
  - exit code is `0`
- after `git add .gitattributes secret.txt public.txt && git commit -q -m 'add both files'`:
  - exit code is `0`
  - file `secret.txt` equals exact bytes
- after `git show HEAD:secret.txt`:
  - exit code is `0`
  - stdout does not contain `not-a-real-key`, matches `/^U2FsdGVkX1/`
- after `git show HEAD:public.txt`:
  - exit code is `0`
  - stdout equals an exact value
- after `transcrypt --list`:
  - exit code is `0`
  - stdout equals an exact value
- after `transcrypt --show-raw secret.txt`:
  - exit code is `0`
  - stdout matches `/^U2FsdGVkX1/`

#### Expected output
_expected stdout:_
```text
nothing secret here
```
_expected stdout:_
```text
secret.txt
```
### Scenario: re-saving the same content produces no change at all
_only when `command -v transcrypt` succeeds · skipped on Windows_
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
git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p test-password-three -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'value = one\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m first
printf 'value = one\n' > secret.txt && git status --porcelain
printf 'value = two\n' > secret.txt && git status --porcelain
git diff
git diff --stat
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p test-password-three -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'value = one\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m first`:
  - exit code is `0`
- after `printf 'value = one\n' > secret.txt && git status --porcelain`:
  - exit code is `0`
  - stdout is empty
- after `printf 'value = two\n' > secret.txt && git status --porcelain`:
  - exit code is `0`
  - stdout equals an exact value
- after `git diff`:
  - exit code is `0`
  - stdout contains `-value = one`, `+value = two`
- after `git diff --stat`:
  - exit code is `0`
  - stdout contains `secret.txt | Bin`

#### Expected output
_expected stdout:_
```text
 M secret.txt
```
### Scenario: two files with the same content do not encrypt to the same bytes
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `.gitattributes` is created.
- Fixture file `first.secret` is created.
- Fixture file `second.secret` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `.gitattributes`:_
```text
*.secret filter=crypt diff=crypt merge=crypt
```
_Fixture `first.secret`:_
```text
the same content
```
_Fixture `second.secret`:_
```text
the same content
```
#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p test-password-four -y
git add .gitattributes first.secret second.secret && git commit -q -m 'two identical files'
git show HEAD:first.secret > first.raw && git show HEAD:second.secret > second.raw && cmp -s first.raw second.raw
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p test-password-four -y`:
  - exit code is `0`
- after `git add .gitattributes first.secret second.secret && git commit -q -m 'two identical files'`:
  - exit code is `0`
- after `git show HEAD:first.secret > first.raw && git show HEAD:second.secret > second.raw && cmp -s first.raw second.raw`:
  - exit code is `1`
  - file `first.raw` does not contain `the same content`

## transcrypt (cloning, the wrong password, and uninstalling)
The workflow around [transcrypt](https://github.com/elasticdog/transcrypt):
what a colleague sees when they clone the repository, what happens when they
type the wrong password, and what is left after uninstalling.

The clone is a real one — a second repository in the same workdir, cloned
from the first — so "the files stay encrypted until you are configured" is
asserted on an actual working tree rather than described. The uninstall is
driven in a pseudo-terminal, because it asks a yes/no question, and both
answers are pinned: the one that removes the configuration and the one that
leaves everything as it was.

Source: `test/e2e/thirdparty/transcrypt/workflow.atago.yaml`
### Scenario: a clone carries ciphertext until it is given the password
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
mkdir origin && cd origin && git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p clone-password -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = shared-with-the-team\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'
git clone -q origin copy
transcrypt -c aes-256-cbc -p clone-password -y
```
#### Then
- after `mkdir origin && cd origin && git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p clone-password -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = shared-with-the-team\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'`:
  - exit code is `0`
- after `git clone -q origin copy`:
  - exit code is `0`
  - file `copy/secret.txt` contains `U2FsdGVkX1`
  - file `copy/secret.txt` does not contain `shared-with-the-team`
- after `transcrypt -c aes-256-cbc -p clone-password -y`:
  - exit code is `0`
  - stdout equals an exact value
  - file `copy/secret.txt` equals exact bytes

#### Expected output
_expected stdout:_
```text
The repository has been successfully configured by transcrypt.
```
### Scenario: the wrong password is caught rather than silently producing garbage
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
mkdir origin && cd origin && git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p the-right-password -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = shared-with-the-team\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'
git clone -q origin copy
transcrypt -c aes-256-cbc -p the-wrong-password -y
```
#### Then
- after `mkdir origin && cd origin && git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p the-right-password -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = shared-with-the-team\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'`:
  - exit code is `0`
- after `git clone -q origin copy`:
  - exit code is `0`
- after `transcrypt -c aes-256-cbc -p the-wrong-password -y`:
  - exit code is `1`
  - stderr contains `transcrypt: Unexpected new dirty files in the repository when configured by transcrypt, please check your password.`
  - file `copy/secret.txt` does not contain `shared-with-the-team`

### Scenario: uninstalling asks first, and leaves the plaintext behind when told to go ahead
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p uninstall-password -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = kept-after-uninstall\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'
# interactive (pty): transcrypt --uninstall
git config --get transcrypt.password
git show HEAD:secret.txt
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p uninstall-password -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = kept-after-uninstall\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'`:
  - exit code is `0`
- after `interactive (pty): transcrypt --uninstall`:
  - rendered screen contains `You are about to remove all transcrypt configuration from your repository.`, `All previously encrypted files will remain decrypted in this working copy`, `The transcrypt configuration has been completely removed from the repository.`
- after `git config --get transcrypt.password`:
  - exit code is `1`
  - stdout is empty
  - file `.gitattributes` does not contain `filter=crypt`
  - file `secret.txt` equals exact bytes
- after `git show HEAD:secret.txt`:
  - exit code is `0`
  - stdout matches `/^U2FsdGVkX1/`

### Scenario: answering no to the uninstall changes nothing
_only when `command -v transcrypt` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p keep-me -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = still-encrypted\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'
# interactive (pty): transcrypt --uninstall
git config --get transcrypt.password
```
#### Then
- after `git init -q && git config user.email author@example.com && git config user.name 'A Author' && transcrypt -c aes-256-cbc -p keep-me -y && printf 'secret.txt filter=crypt diff=crypt merge=crypt\n' > .gitattributes && printf 'token = still-encrypted\n' > secret.txt && git add .gitattributes secret.txt && git commit -q -m 'add the secret'`:
  - exit code is `0`
- after `interactive (pty): transcrypt --uninstall`:
  - rendered screen contains `transcrypt: uninstallation has been aborted`
- after `git config --get transcrypt.password`:
  - exit code is `0`
  - stdout equals an exact value
  - file `.gitattributes` contains `filter=crypt`

#### Expected output
_expected stdout:_
```text
keep-me
```
