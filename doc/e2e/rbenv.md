# atago Behavior Specs
## Summary
2 suites · 16 scenarios
## Contents
- [rbenv (which version this directory selects)](#rbenv-which-version-this-directory-selects) — 9 scenarios
  - [the version, the usage text, and a command that does not exist](#scenario-the-version-the-usage-text-and-a-command-that-does-not-exist)
  - [an empty root has no versions but still answers with system](#scenario-an-empty-root-has-no-versions-but-still-answers-with-system)
  - [installed versions are listed and the selected one is marked](#scenario-installed-versions-are-listed-and-the-selected-one-is-marked)
  - [global writes one file in the root and nothing else](#scenario-global-writes-one-file-in-the-root-and-nothing-else)
  - [local wins over global, and the environment wins over both](#scenario-local-wins-over-global-and-the-environment-wins-over-both)
  - [unsetting the local version falls back to the global one](#scenario-unsetting-the-local-version-falls-back-to-the-global-one)
  - [a version that is not installed is refused and nothing is written](#scenario-a-version-that-is-not-installed-is-refused-and-nothing-is-written)
  - [a version file naming an uninstalled version says which file to fix](#scenario-a-version-file-naming-an-uninstalled-version-says-which-file-to-fix)
  - [which, prefix, version-name and version-origin agree on the selection](#scenario-which-prefix-version-name-and-version-origin-agree-on-the-selection)
- [rbenv (shims, exec, and shell integration)](#rbenv-shims-exec-and-shell-integration) — 7 scenarios
  - [rehash generates one shim per executable across every version](#scenario-rehash-generates-one-shim-per-executable-across-every-version)
  - [the shim answers to the directory, with PATH never changing](#scenario-the-shim-answers-to-the-directory-with-path-never-changing)
  - [a version added after the last rehash has no shim until the next one](#scenario-a-version-added-after-the-last-rehash-has-no-shim-until-the-next-one)
  - [exec runs the selected version, and a command that is not there is 127](#scenario-exec-runs-the-selected-version-and-a-command-that-is-not-there-is-127)
  - [a hook script runs before exec and sees the resolved version](#scenario-a-hook-script-runs-before-exec-and-sees-the-resolved-version)
  - [shell integration is printed on request and refused without it](#scenario-shell-integration-is-printed-on-request-and-refused-without-it)
  - [init edits the profile in the home it was given, once](#scenario-init-edits-the-profile-in-the-home-it-was-given-once)

## rbenv (which version this directory selects)
[rbenv](https://github.com/rbenv/rbenv) picks the Ruby version a command
should run under. It keeps its own test suite in Bats; what those tests
check is pinned here from outside, against a root the scenario builds
itself.

No Ruby is installed and none is needed: a version is a directory under
`versions/`, so each scenario writes a couple of two-line executables and
exercises the real resolution rules against them. What is asserted is the
state machine — the file `global` writes, the file `local` writes, the
environment variable that beats both, the origin rbenv reports for its
choice, and the refusals that leave every one of those files untouched.

Source: `test/e2e/thirdparty/rbenv/rbenv.atago.yaml`
### Scenario: the version, the usage text, and a command that does not exist
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
rbenv --version
rbenv
rbenv nosuchcommand
```
#### Then
- after `rbenv --version`:
  - exit code is `0`
  - stdout matches `/^rbenv [0-9]+\.[0-9]+/`
- after `rbenv`:
  - exit code is `1`
  - stdout is empty
  - stderr contains `Usage: rbenv <command> [<args>...]`, `version     Show the current Ruby version and its origin`
- after `rbenv nosuchcommand`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value

#### Expected output
_expected stderr:_
```text
rbenv: no such command `nosuchcommand'
```
### Scenario: an empty root has no versions but still answers with system
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
rbenv root
rbenv versions --bare
rbenv version
```
#### Then
- after `rbenv root`:
  - exit code is `0`
  - stdout matches `//root\n$/`
- after `rbenv versions --bare`:
  - exit code is `0`
  - stdout is empty
- after `rbenv version`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
system
```
### Scenario: installed versions are listed and the selected one is marked
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1"
```
#### When
```shell
rbenv versions --bare
rbenv global 3.3.1
rbenv versions
```
#### Then
- after `rbenv versions --bare`:
  - exit code is `0`
  - stdout equals an exact value
- after `rbenv global 3.3.1`:
  - exit code is `0`
- after `rbenv versions`:
  - exit code is `0`
  - stdout matches `/(?m)^\* 3\.3\.1 \(set by .*/root/version\)$/`
  - stdout matches `/(?m)^  3\.2\.0$/`

#### Expected output
_expected stdout:_
```text
3.2.0
3.3.1
```
### Scenario: global writes one file in the root and nothing else
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
#### When
```shell
rbenv global 3.2.0
rbenv global
rbenv version
```
#### Then
- after `rbenv global 3.2.0`:
  - exit code is `0`
  - stdout is empty
  - the step changed exactly created `root/version`, modified nothing, deleted nothing
  - file `root/version` equals exact bytes
- after `rbenv global`:
  - exit code is `0`
  - stdout equals an exact value
- after `rbenv version`:
  - exit code is `0`
  - stdout matches `/^3\.2\.0 \(set by .*/root/version\)\n$/`

#### Expected output
_expected stdout:_
```text
3.2.0
```
### Scenario: local wins over global, and the environment wins over both
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT, RBENV_VERSION.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1"
```
#### When
```shell
rbenv global 3.2.0
rbenv local 3.3.1
rbenv version
rbenv version
```
#### Then
- after `rbenv global 3.2.0`:
  - exit code is `0`
- after `rbenv local 3.3.1`:
  - exit code is `0`
  - the step changed exactly created `.ruby-version`, modified nothing, deleted nothing
  - file `.ruby-version` equals exact bytes
- after `rbenv version`:
  - exit code is `0`
  - stdout matches `/^3\.3\.1 \(set by .*/\.ruby-version\)\n$/`
- after `rbenv version`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
3.2.0 (set by RBENV_VERSION environment variable)
```
### Scenario: unsetting the local version falls back to the global one
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Fixture file `.ruby-version` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1"
```
_Fixture `.ruby-version`:_
```text
3.3.1
```
#### When
```shell
rbenv global 3.2.0
rbenv local --unset
rbenv version
```
#### Then
- after `rbenv global 3.2.0`:
  - exit code is `0`
- after `rbenv local --unset`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted `.ruby-version`
- after `rbenv version`:
  - exit code is `0`
  - stdout matches `/^3\.2\.0 \(set by .*/root/version\)\n$/`

### Scenario: a version that is not installed is refused and nothing is written
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `.ruby-version` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `.ruby-version`:_
```text
3.2.0
```
#### When
```shell
rbenv local 9.9.9
rbenv global 9.9.9
```
#### Then
- after `rbenv local 9.9.9`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
  - the step changed exactly created nothing, modified nothing, deleted nothing
  - file `.ruby-version` equals exact bytes
- after `rbenv global 9.9.9`:
  - exit code is `1`
  - stderr equals an exact value
  - the step changed exactly created nothing, modified nothing, deleted nothing

#### Expected output
_expected stderr:_
```text
rbenv: version `9.9.9' not installed
```
_expected stderr:_
```text
rbenv: version `9.9.9' not installed
```
### Scenario: a version file naming an uninstalled version says which file to fix
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `.ruby-version` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `.ruby-version`:_
```text
9.9.9
```
#### When
```shell
rbenv version
```
#### Then
- exit code is `1`
- stdout is empty
- stderr matches `` /^rbenv: version `9\.9\.9' is not installed \(set by .*/\.ruby-version\)\n$/ ``

### Scenario: which, prefix, version-name and version-origin agree on the selection
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Fixture file `.ruby-version` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1"
```
_Fixture `.ruby-version`:_
```text
3.3.1
```
#### When
```shell
rbenv version-name
rbenv version-origin
rbenv prefix
rbenv which ruby
```
#### Then
- after `rbenv version-name`:
  - exit code is `0`
  - stdout equals an exact value
- after `rbenv version-origin`:
  - exit code is `0`
  - stdout matches `//\.ruby-version\n$/`
- after `rbenv prefix`:
  - exit code is `0`
  - stdout matches `//root/versions/3\.3\.1\n$/`
- after `rbenv which ruby`:
  - exit code is `0`
  - stdout matches `//root/versions/3\.3\.1/bin/ruby\n$/`

#### Expected output
_expected stdout:_
```text
3.3.1
```
## rbenv (shims, exec, and shell integration)
The half of [rbenv](https://github.com/rbenv/rbenv) that actually runs
something: the shims it generates, the dispatch that makes `ruby` mean a
different program in a different directory without PATH ever changing,
`rbenv exec`, the hooks a plugin installs, and the shell integration
`rbenv init` writes into the user's home.

The dispatch is the whole point of the tool, so it is asserted the way a
user meets it — by running `ruby` with the shim directory on PATH and
reading what came back, before and after the directory's version changes.
`rbenv init` is asserted against a sandboxed home, which is what proves the
file it edits is the one it claims and that a second run edits nothing.

Source: `test/e2e/thirdparty/rbenv/shims.atago.yaml`
### Scenario: rehash generates one shim per executable across every version
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `root/versions/3.2.0/bin/gem` is created.
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `root/versions/3.2.0/bin/gem`:_
```text
#!/bin/sh
echo "gem 3.2.0"
```
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1"
```
#### When
```shell
rbenv rehash
rbenv shims --short
```
#### Then
- after `rbenv rehash`:
  - exit code is `0`
  - the step changed exactly created `root/shims/ruby`, `root/shims/gem`, modified nothing, deleted nothing
  - dir `root/shims` has 2 entries
  - file `root/shims/ruby` is executable
- after `rbenv shims --short`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
gem
ruby
```
### Scenario: the shim answers to the directory, with PATH never changing
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT, RBENV_VERSION.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1"
```
#### When
```shell
rbenv global 3.2.0
rbenv rehash
PATH="${workdir}/root/shims:$PATH" ruby
rbenv local 3.3.1
PATH="${workdir}/root/shims:$PATH" ruby
PATH="${workdir}/root/shims:$PATH" ruby
```
#### Then
- after `rbenv global 3.2.0`:
  - exit code is `0`
- after `rbenv rehash`:
  - exit code is `0`
- after `PATH="${workdir}/root/shims:$PATH" ruby`:
  - exit code is `0`
  - stdout equals an exact value
- after `rbenv local 3.3.1`:
  - exit code is `0`
- after `PATH="${workdir}/root/shims:$PATH" ruby`:
  - exit code is `0`
  - stdout equals an exact value
- after `PATH="${workdir}/root/shims:$PATH" ruby`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
ruby 3.2.0
```
_expected stdout:_
```text
ruby 3.3.1
```
_expected stdout:_
```text
ruby 3.2.0
```
### Scenario: a version added after the last rehash has no shim until the next one
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Fixture file `root/versions/3.3.1/bin/irb` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `root/versions/3.3.1/bin/irb`:_
```text
#!/bin/sh
echo "irb 3.3.1"
```
#### When
```shell
rbenv rehash
rbenv rehash
```
#### Then
- after `rbenv rehash`:
  - exit code is `0`
  - the step changed exactly created `root/shims/ruby`, modified nothing, deleted nothing
- after `rbenv rehash`:
  - exit code is `0`
  - the step changed exactly created `root/shims/irb`, modified nothing, deleted nothing

### Scenario: exec runs the selected version, and a command that is not there is 127
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.3.1/bin/ruby` is created.
- Fixture file `.ruby-version` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.3.1/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.3.1, argv: $*"
```
_Fixture `.ruby-version`:_
```text
3.3.1
```
#### When
```shell
rbenv exec ruby --flag argument
rbenv exec rake
rbenv which rake
```
#### Then
- after `rbenv exec ruby --flag argument`:
  - exit code is `0`
  - stdout equals an exact value
- after `rbenv exec rake`:
  - exit code is `127`
  - stdout is empty
  - stderr equals an exact value
- after `rbenv which rake`:
  - exit code is `127`
  - stderr equals an exact value

#### Expected output
_expected stdout:_
```text
ruby 3.3.1, argv: --flag argument
```
_expected stderr:_
```text
rbenv: rake: command not found
```
_expected stderr:_
```text
rbenv: rake: command not found
```
### Scenario: a hook script runs before exec and sees the resolved version
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Fixture file `.ruby-version` is created.
- Fixture file `hooks/exec/record.bash` is created.
- Environment variables are set: RBENV_HOOK_PATH, RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_HOOK_PATH, RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
_Fixture `.ruby-version`:_
```text
3.2.0
```
_Fixture `hooks/exec/record.bash`:_
```text
echo "hook saw $RBENV_VERSION" >> "$RBENV_ROOT/hook.log"
```
#### When
```shell
rbenv hooks exec
rbenv exec ruby
```
#### Then
- after `rbenv hooks exec`:
  - exit code is `0`
  - stdout matches `//hooks/exec/record\.bash/`
- after `rbenv exec ruby`:
  - exit code is `0`
  - stdout equals an exact value
  - the step changed exactly created `root/hook.log`, modified nothing, deleted nothing
  - file `root/hook.log` equals exact bytes

#### Expected output
_expected stdout:_
```text
ruby 3.2.0
```
### Scenario: shell integration is printed on request and refused without it
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.2.0/bin/ruby` is created.
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.2.0/bin/ruby`:_
```text
#!/bin/sh
echo "ruby 3.2.0"
```
#### When
```shell
rbenv shell 3.2.0
rbenv init - bash
```
#### Then
- after `rbenv shell 3.2.0`:
  - exit code is `1`
  - stdout is empty
  - stderr equals an exact value
- after `rbenv init - bash`:
  - exit code is `0`
  - stdout contains `export RBENV_SHELL=bash`, `command rbenv rehash 2>/dev/null`, `rbenv() {`, matches `/export PATH="[^"]*/root/shims:\$\{PATH\}"/`

#### Expected output
_expected stderr:_
```text
rbenv: shell integration not enabled. Run `rbenv init' for instructions.
```
### Scenario: init edits the profile in the home it was given, once
_only when `rbenv --version` succeeds · skipped on Windows_
#### Given
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: RBENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
rbenv init bash
rbenv init bash
```
#### Then
- after `rbenv init bash`:
  - exit code is `0`
  - stdout matches `/^writing ~/\.bash_profile: now configured for rbenv\.\n$/`
  - the step changed exactly created `.atago-home/.bash_profile`, modified nothing, deleted nothing
  - file `.atago-home/.bash_profile` contains `eval "$(rbenv init - --no-rehash bash)"`
- after `rbenv init bash`:
  - exit code is `0`
  - stdout matches `/^skipping ~/\.bash_profile: already configured for rbenv\.\n$/`
  - the step changed exactly created nothing, modified nothing, deleted nothing
