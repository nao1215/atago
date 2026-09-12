# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [pyenv (more than one version at a time)](#pyenv-more-than-one-version-at-a-time) — 5 scenarios
  - [a version file can name several versions, and order decides](#scenario-a-version-file-can-name-several-versions-and-order-decides)
  - [whence answers which of them has the command](#scenario-whence-answers-which-of-them-has-the-command)
  - [the shim reaches into whichever selected version has the command](#scenario-the-shim-reaches-into-whichever-selected-version-has-the-command)
  - [an alias is a version to list and not a version to skip](#scenario-an-alias-is-a-version-to-list-and-not-a-version-to-skip)
  - [a version file naming something that is not installed says which file](#scenario-a-version-file-naming-something-that-is-not-installed-says-which-file)

## pyenv (more than one version at a time)
[pyenv](https://github.com/pyenv/pyenv) selects the Python a command runs
under. It keeps its own test suite in Bats, and it shares its ancestry with
rbenv — so this suite deliberately does not repeat what the
[rbenv suite](rbenv.md) already pins. What is here is what pyenv does
differently: a version file that names several versions at once, aliases
that sit beside real versions, and `whence`, which answers the question
those two raise — which of the selected versions actually provides this
command.

No Python is installed and none is needed: a version is a directory under
the root, so each scenario writes a couple of two-line executables and
exercises the real resolution rules against them.

Source: `test/e2e/thirdparty/pyenv/pyenv.atago.yaml`
### Scenario: a version file can name several versions, and order decides
_only when `pyenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.11.9/bin/python` is created.
- Fixture file `root/versions/3.12.4/bin/python` is created.
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.11.9/bin/python`:_
```text
#!/bin/sh
echo "Python 3.11.9"
```
_Fixture `root/versions/3.12.4/bin/python`:_
```text
#!/bin/sh
echo "Python 3.12.4"
```
#### When
```shell
pyenv local 3.12.4 3.11.9
pyenv version
pyenv which python
```
#### Then
- after `pyenv local 3.12.4 3.11.9`:
  - exit code is `0`
  - the step changed exactly created `.python-version`, modified nothing, deleted nothing
  - file `.python-version` equals exact bytes
- after `pyenv version`:
  - exit code is `0`
  - stdout matches `/(?m)^3\.12\.4 \(set by .*/\.python-version\)$/`
  - stdout matches `/(?m)^3\.11\.9 \(set by .*/\.python-version\)$/`
- after `pyenv which python`:
  - exit code is `0`
  - stdout matches `//root/versions/3\.12\.4/bin/python\n$/`

### Scenario: whence answers which of them has the command
_only when `pyenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.11.9/bin/python` is created.
- Fixture file `root/versions/3.11.9/bin/tox` is created.
- Fixture file `root/versions/3.12.4/bin/python` is created.
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.11.9/bin/python`:_
```text
#!/bin/sh
echo "Python 3.11.9"
```
_Fixture `root/versions/3.11.9/bin/tox`:_
```text
#!/bin/sh
echo "tox from 3.11.9"
```
_Fixture `root/versions/3.12.4/bin/python`:_
```text
#!/bin/sh
echo "Python 3.12.4"
```
#### When
```shell
pyenv whence python
pyenv whence tox
pyenv whence pip
```
#### Then
- after `pyenv whence python`:
  - exit code is `0`
  - stdout equals an exact value
- after `pyenv whence tox`:
  - exit code is `0`
  - stdout equals an exact value
- after `pyenv whence pip`:
  - exit code is `1`
  - stdout is empty

#### Expected output
_expected stdout:_
```text
3.11.9
3.12.4
```
_expected stdout:_
```text
3.11.9
```
### Scenario: the shim reaches into whichever selected version has the command
_only when `pyenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.11.9/bin/python` is created.
- Fixture file `root/versions/3.11.9/bin/tox` is created.
- Fixture file `root/versions/3.12.4/bin/python` is created.
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.11.9/bin/python`:_
```text
#!/bin/sh
echo "Python 3.11.9"
```
_Fixture `root/versions/3.11.9/bin/tox`:_
```text
#!/bin/sh
echo "tox from 3.11.9"
```
_Fixture `root/versions/3.12.4/bin/python`:_
```text
#!/bin/sh
echo "Python 3.12.4"
```
#### When
```shell
pyenv local 3.12.4 3.11.9 && pyenv rehash
PATH="${workdir}/root/shims:$PATH" python
PATH="${workdir}/root/shims:$PATH" tox
```
#### Then
- after `pyenv local 3.12.4 3.11.9 && pyenv rehash`:
  - exit code is `0`
  - dir `root/shims` contains `python`, contains `tox`, has 2 entries
- after `PATH="${workdir}/root/shims:$PATH" python`:
  - exit code is `0`
  - stdout equals an exact value
- after `PATH="${workdir}/root/shims:$PATH" tox`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
Python 3.12.4
```
_expected stdout:_
```text
tox from 3.11.9
```
### Scenario: an alias is a version to list and not a version to skip
_only when `pyenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.11.9/bin/python` is created.
- Fixture file `root/versions/stable` is created.
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.11.9/bin/python`:_
```text
#!/bin/sh
echo "Python 3.11.9"
```
#### When
```shell
pyenv versions --bare
pyenv versions --bare --skip-aliases
pyenv local stable
pyenv which python
```
#### Then
- after `pyenv versions --bare`:
  - exit code is `0`
  - stdout equals an exact value
- after `pyenv versions --bare --skip-aliases`:
  - exit code is `0`
  - stdout equals an exact value
- after `pyenv local stable`:
  - exit code is `0`
- after `pyenv which python`:
  - exit code is `0`
  - stdout matches `//root/versions/stable/bin/python\n$/`

#### Expected output
_expected stdout:_
```text
3.11.9
stable
```
_expected stdout:_
```text
3.11.9
```
### Scenario: a version file naming something that is not installed says which file
_only when `pyenv --version` succeeds · skipped on Windows_
#### Given
- Fixture file `root/versions/3.11.9/bin/python` is created.
- Fixture file `.python-version` is created.
- Environment variables are set: PYENV_ROOT.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `root/versions/3.11.9/bin/python`:_
```text
#!/bin/sh
echo "Python 3.11.9"
```
_Fixture `.python-version`:_
```text
3.11.9
9.9.9
```
#### When
```shell
pyenv version
```
#### Then
- exit code is `1`
- stderr matches `` /pyenv: version `9\.9\.9' is not installed \(set by .*/\.python-version\)/ ``
