# atago Behavior Specs
## Summary
2 suites · 8 scenarios
## Contents
- [python3 + changes (bytecode cache footprint)](#python3--changes-bytecode-cache-footprint) — 2 scenarios
  - [importing a local module writes exactly one .pyc](#scenario-importing-a-local-module-writes-exactly-one-pyc)
  - [PYTHONDONTWRITEBYTECODE suppresses the cache entirely](#scenario-pythondontwritebytecode-suppresses-the-cache-entirely)
- [python3 REPL (interactive pty testbed)](#python3-repl-interactive-pty-testbed) — 6 scenarios
  - [version and -c contracts (non-interactive)](#scenario-version-and--c-contracts-non-interactive)
  - [a missing script exits 2 with a can't-open-file error](#scenario-a-missing-script-exits-2-with-a-cant-open-file-error)
  - [stdout is a pipe under run but a tty under pty](#scenario-stdout-is-a-pipe-under-run-but-a-tty-under-pty)
  - [an interactive session drives the REPL across exchanges](#scenario-an-interactive-session-drives-the-repl-across-exchanges)
  - [EOF (ctrl-d) ends the session cleanly](#scenario-eof-ctrl-d-ends-the-session-cleanly)
  - [a traceback is reported and the REPL recovers](#scenario-a-traceback-is-reported-and-the-repl-recovers)
## python3 + changes (bytecode cache footprint)
Source: `test/e2e/thirdparty/python/changes.atago.yaml`
### Scenario: importing a local module writes exactly one .pyc
#### Given
- Fixture file `mymod.py` is created.
- Fixture file `main.py` is created.
#### Inputs
_Fixture `mymod.py`:_
```text
def val():
    return 42
```
_Fixture `main.py`:_
```text
import mymod
print(mymod.val())
```
#### When
```shell
python3 main.py
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created `__pycache__/*.pyc`, modified nothing, deleted nothing
### Scenario: PYTHONDONTWRITEBYTECODE suppresses the cache entirely
#### Given
- Fixture file `mymod.py` is created.
- Fixture file `main.py` is created.
- Environment variables are set: PYTHONDONTWRITEBYTECODE.
#### Inputs
_Fixture `mymod.py`:_
```text
def val():
    return 42
```
_Fixture `main.py`:_
```text
import mymod
print(mymod.val())
```
#### When
```shell
python3 main.py
```
#### Then
- exit code is `0`
- stdout equals an exact value
- the step changed exactly created nothing, modified nothing, deleted nothing
## python3 REPL (interactive pty testbed)
Source: `test/e2e/thirdparty/python/python.atago.yaml`
### Scenario: version and -c contracts (non-interactive)
#### When
```shell
python3 --version
python3 -c "print('hello from python')"
```
#### Then
- after `python3 --version`:
  - exit code is `0`
  - stdout matches `/^Python 3\./`
- after `python3 -c "print('hello from python')"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: a missing script exits 2 with a can't-open-file error
#### When
```shell
python3 no_such_script.py
```
#### Then
- exit code is `2`
- stderr contains `can't open file`
### Scenario: stdout is a pipe under run but a tty under pty
_skipped on windows_
#### When
```shell
python3 -c "import sys; print(sys.stdout.isatty())"
# interactive (pty): python3 -c "import sys; print(sys.stdout.isatty())"
```
#### Then
- exit code is `0`
- stdout equals an exact value
- exit code is `0`
- stdout contains `True`
### Scenario: an interactive session drives the REPL across exchanges
_skipped on windows_
#### When
```shell
# interactive (pty): python3 -q
```
#### Then
- exit code is `0`
- stdout contains `2`, `120`
### Scenario: EOF (ctrl-d) ends the session cleanly
_skipped on windows_
#### When
```shell
# interactive (pty): python3 -q
```
#### Then
- exit code is `0`
### Scenario: a traceback is reported and the REPL recovers
_skipped on windows_
#### When
```shell
# interactive (pty): python3 -q
```
#### Then
- exit code is `0`
- stdout contains `ZeroDivisionError`, `recovered`