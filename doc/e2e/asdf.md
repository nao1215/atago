# atago Behavior Specs
## Summary
2 suites · 7 scenarios
## Contents
- [asdf (the plugin protocol)](#asdf-the-plugin-protocol) — 4 scenarios
  - [the version, and a command it does not have](#scenario-the-version-and-a-command-it-does-not-have)
  - [a plugin is a repository, added from wherever it lives](#scenario-a-plugin-is-a-repository-added-from-wherever-it-lives)
  - [installing runs the plugin and lands a real program](#scenario-installing-runs-the-plugin-and-lands-a-real-program)
  - [removing the plugin takes its versions with it](#scenario-removing-the-plugin-takes-its-versions-with-it)
- [asdf (which version this directory gets)](#asdf-which-version-this-directory-gets) — 3 scenarios
  - [global writes the home file and local writes the directory's](#scenario-global-writes-the-home-file-and-local-writes-the-directorys)
  - [the shim runs whatever the directory says, with PATH never changing](#scenario-the-shim-runs-whatever-the-directory-says-with-path-never-changing)
  - [exec and which agree, and a missing version explains itself](#scenario-exec-and-which-agree-and-a-missing-version-explains-itself)

## asdf (the plugin protocol)
[asdf](https://asdf-vm.com/) manages versions of any tool through plugins,
and a plugin is just a git repository holding a few executables that asdf
calls. It keeps its own test suite in Bats; what those tests check is pinned
here from outside.

Nothing is downloaded. Each scenario writes its own plugin — `bin/list-all`,
`bin/download`, `bin/install` — commits it in the workdir, and adds it from
that path, so the protocol itself is what is under test: which script asdf
calls, what it passes in the environment, and what it does with the result,
including a plugin that refuses a version.

Source: `test/e2e/thirdparty/asdf/plugins.atago.yaml`
### Scenario: the version, and a command it does not have
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### When
```shell
asdf version
asdf nosuchcommand
```
#### Then
- after `asdf version`:
  - exit code is `0`
  - stdout matches `/^v[0-9]+\.[0-9]+\.[0-9]+/`
- after `asdf nosuchcommand`:
  - exit code is `1`
  - stderr contains `` Unknown command: `asdf nosuchcommand` ``, `No plugin named nosuchcommand`

### Scenario: a plugin is a repository, added from wherever it lives
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Fixture file `myplugin/bin/list-all` is created.
- Fixture file `myplugin/bin/download` is created.
- Fixture file `myplugin/bin/install` is created.
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
_Fixture `myplugin/bin/list-all`:_
```text
#!/usr/bin/env bash
echo "1.0.0 1.1.0 2.0.0"
```
_Fixture `myplugin/bin/download`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_DOWNLOAD_PATH"
echo "downloaded $ASDF_INSTALL_VERSION" > "$ASDF_DOWNLOAD_PATH/payload"
```
_Fixture `myplugin/bin/install`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_INSTALL_PATH/bin"
printf '#!/bin/sh\necho "mytool %s"\n' "$ASDF_INSTALL_VERSION" > "$ASDF_INSTALL_PATH/bin/mytool"
chmod +x "$ASDF_INSTALL_PATH/bin/mytool"
```
#### When
```shell
cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin'
asdf plugin list
asdf plugin add mytool ${workdir}/myplugin
asdf plugin list
asdf list all mytool
```
#### Then
- after `cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin'`:
  - exit code is `0`
- after `asdf plugin list`:
  - exit code is `0`
  - stdout is empty
  - stderr equals an exact value
- after `asdf plugin add mytool ${workdir}/myplugin`:
  - exit code is `0`
  - dir `.atago-home/.asdf/plugins/mytool/bin` contains `list-all`, contains `download`, contains `install`
- after `asdf plugin list`:
  - exit code is `0`
  - stdout equals an exact value
- after `asdf list all mytool`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stderr:_
```text
No plugins installed
```
_expected stdout:_
```text
mytool
```
_expected stdout:_
```text
1.0.0
1.1.0
2.0.0
```
### Scenario: installing runs the plugin and lands a real program
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Fixture file `myplugin/bin/list-all` is created.
- Fixture file `myplugin/bin/download` is created.
- Fixture file `myplugin/bin/install` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `myplugin/bin/list-all`:_
```text
#!/usr/bin/env bash
echo "1.0.0 1.1.0"
```
_Fixture `myplugin/bin/download`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_DOWNLOAD_PATH"
echo "downloaded $ASDF_INSTALL_VERSION" > "$ASDF_DOWNLOAD_PATH/payload"
```
_Fixture `myplugin/bin/install`:_
```text
#!/usr/bin/env bash
case "$ASDF_INSTALL_VERSION" in
  1.0.0|1.1.0) ;;
  *) echo "mytool: no such version: $ASDF_INSTALL_VERSION" >&2; exit 1 ;;
esac
mkdir -p "$ASDF_INSTALL_PATH/bin"
printf '#!/bin/sh\necho "mytool %s"\n' "$ASDF_INSTALL_VERSION" > "$ASDF_INSTALL_PATH/bin/mytool"
chmod +x "$ASDF_INSTALL_PATH/bin/mytool"
```
#### When
```shell
cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin
asdf install mytool 1.1.0
asdf list mytool
asdf install mytool 9.9.9
```
#### Then
- after `cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin`:
  - exit code is `0`
- after `asdf install mytool 1.1.0`:
  - exit code is `0`
  - file `.atago-home/.asdf/installs/mytool/1.1.0/bin/mytool` is executable
  - dir `.atago-home/.asdf/downloads/mytool/1.1.0` does not exist
- after `asdf list mytool`:
  - exit code is `0`
  - stdout equals an exact value
- after `asdf install mytool 9.9.9`:
  - exit code is `1`
  - stderr contains `mytool: no such version: 9.9.9`
  - dir `.atago-home/.asdf/installs/mytool` does not contain `9.9.9`
  - file `.atago-home/.asdf/downloads/mytool/9.9.9/payload` equals exact bytes

#### Expected output
_expected stdout:_
```text
  1.1.0
```
### Scenario: removing the plugin takes its versions with it
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Fixture file `myplugin/bin/list-all` is created.
- Fixture file `myplugin/bin/download` is created.
- Fixture file `myplugin/bin/install` is created.
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- Environment variables are set: LC_ALL.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).

#### Inputs
_Fixture `myplugin/bin/list-all`:_
```text
#!/usr/bin/env bash
echo "1.0.0"
```
_Fixture `myplugin/bin/download`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_DOWNLOAD_PATH"
```
_Fixture `myplugin/bin/install`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_INSTALL_PATH/bin"
printf '#!/bin/sh\necho "mytool %s"\n' "$ASDF_INSTALL_VERSION" > "$ASDF_INSTALL_PATH/bin/mytool"
chmod +x "$ASDF_INSTALL_PATH/bin/mytool"
```
#### When
```shell
cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0
asdf plugin remove mytool
asdf plugin list
```
#### Then
- after `cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0`:
  - exit code is `0`
- after `asdf plugin remove mytool`:
  - exit code is `0`
  - dir `.atago-home/.asdf/plugins/mytool` does not exist
  - dir `.atago-home/.asdf/installs/mytool` does not exist
- after `asdf plugin list`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
*
```
## asdf (which version this directory gets)
The other half of [asdf](https://asdf-vm.com/): once a plugin has installed
a few versions, which one a command actually runs. The rules are a file —
`.tool-versions` — read from the current directory upwards, with the one in
the home directory as the fallback, and shims on PATH doing the dispatch.

Every scenario builds its own plugin and installs two versions of a program
that prints which one it is, so "the directory decided" is asserted by
running the program rather than by reading a report about it.

Source: `test/e2e/thirdparty/asdf/versions.atago.yaml`
### Scenario: global writes the home file and local writes the directory's
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Fixture file `myplugin/bin/list-all` is created.
- Fixture file `myplugin/bin/download` is created.
- Fixture file `myplugin/bin/install` is created.
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
_Fixture `myplugin/bin/list-all`:_
```text
#!/usr/bin/env bash
echo "1.0.0 2.0.0"
```
_Fixture `myplugin/bin/download`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_DOWNLOAD_PATH"
```
_Fixture `myplugin/bin/install`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_INSTALL_PATH/bin"
printf '#!/bin/sh\necho "mytool %s"\n' "$ASDF_INSTALL_VERSION" > "$ASDF_INSTALL_PATH/bin/mytool"
chmod +x "$ASDF_INSTALL_PATH/bin/mytool"
```
#### When
```shell
cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0 && asdf install mytool 2.0.0
asdf current mytool
asdf global mytool 1.0.0
asdf local mytool 2.0.0
asdf current mytool
```
#### Then
- after `cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0 && asdf install mytool 2.0.0`:
  - exit code is `0`
- after `asdf current mytool`:
  - exit code is `126`
  - stdout is empty
  - stderr contains `No version is set. Run "asdf <global|shell|local> mytool <version>"`
- after `asdf global mytool 1.0.0`:
  - exit code is `0`
  - the step changed exactly created `.atago-home/.tool-versions`, modified nothing, deleted nothing
  - file `.atago-home/.tool-versions` equals exact bytes
- after `asdf local mytool 2.0.0`:
  - exit code is `0`
  - the step changed exactly created `.tool-versions`, modified nothing, deleted nothing
- after `asdf current mytool`:
  - exit code is `0`
  - stdout matches `/mytool\s+2\.0\.0\s+.*/\.tool-versions/`

### Scenario: the shim runs whatever the directory says, with PATH never changing
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Fixture file `myplugin/bin/list-all` is created.
- Fixture file `myplugin/bin/download` is created.
- Fixture file `myplugin/bin/install` is created.
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
_Fixture `myplugin/bin/list-all`:_
```text
#!/usr/bin/env bash
echo "1.0.0 2.0.0"
```
_Fixture `myplugin/bin/download`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_DOWNLOAD_PATH"
```
_Fixture `myplugin/bin/install`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_INSTALL_PATH/bin"
printf '#!/bin/sh\necho "mytool %s"\n' "$ASDF_INSTALL_VERSION" > "$ASDF_INSTALL_PATH/bin/mytool"
chmod +x "$ASDF_INSTALL_PATH/bin/mytool"
```
#### When
```shell
cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0 && asdf install mytool 2.0.0 && asdf global mytool 1.0.0
PATH="${workdir}/.atago-home/.asdf/shims:$PATH" mytool
asdf local mytool 2.0.0
PATH="${workdir}/.atago-home/.asdf/shims:$PATH" mytool
PATH="${workdir}/.atago-home/.asdf/shims:$PATH" ASDF_MYTOOL_VERSION=1.0.0 mytool
```
#### Then
- after `cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0 && asdf install mytool 2.0.0 && asdf global mytool 1.0.0`:
  - exit code is `0`
  - file `.atago-home/.asdf/shims/mytool` is executable
- after `PATH="${workdir}/.atago-home/.asdf/shims:$PATH" mytool`:
  - exit code is `0`
  - stdout equals an exact value
- after `asdf local mytool 2.0.0`:
  - exit code is `0`
- after `PATH="${workdir}/.atago-home/.asdf/shims:$PATH" mytool`:
  - exit code is `0`
  - stdout equals an exact value
- after `PATH="${workdir}/.atago-home/.asdf/shims:$PATH" ASDF_MYTOOL_VERSION=1.0.0 mytool`:
  - exit code is `0`
  - stdout equals an exact value

#### Expected output
_expected stdout:_
```text
mytool 1.0.0
```
_expected stdout:_
```text
mytool 2.0.0
```
_expected stdout:_
```text
mytool 1.0.0
```
### Scenario: exec and which agree, and a missing version explains itself
_only when `command -v asdf` succeeds · skipped on Windows_
#### Given
- Fixture file `myplugin/bin/list-all` is created.
- Fixture file `myplugin/bin/download` is created.
- Fixture file `myplugin/bin/install` is created.
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

#### Inputs
_Fixture `myplugin/bin/list-all`:_
```text
#!/usr/bin/env bash
echo "1.0.0 2.0.0"
```
_Fixture `myplugin/bin/download`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_DOWNLOAD_PATH"
```
_Fixture `myplugin/bin/install`:_
```text
#!/usr/bin/env bash
mkdir -p "$ASDF_INSTALL_PATH/bin"
printf '#!/bin/sh\necho "mytool %s"\n' "$ASDF_INSTALL_VERSION" > "$ASDF_INSTALL_PATH/bin/mytool"
chmod +x "$ASDF_INSTALL_PATH/bin/mytool"
```
#### When
```shell
cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0 && asdf install mytool 2.0.0 && asdf local mytool 2.0.0
asdf which mytool
asdf where mytool
asdf exec mytool
asdf uninstall mytool 2.0.0
PATH="${workdir}/.atago-home/.asdf/shims:$PATH" mytool
```
#### Then
- after `cd myplugin && git init -q && git config user.email plugin@example.com && git config user.name 'A Plugin' && git add -A && git commit -q -m 'the plugin' && cd .. && asdf plugin add mytool ${workdir}/myplugin && asdf install mytool 1.0.0 && asdf install mytool 2.0.0 && asdf local mytool 2.0.0`:
  - exit code is `0`
- after `asdf which mytool`:
  - exit code is `0`
  - stdout matches `//\.asdf/installs/mytool/2\.0\.0/bin/mytool\n$/`
- after `asdf where mytool`:
  - exit code is `0`
  - stdout matches `//\.asdf/installs/mytool/2\.0\.0\n$/`
- after `asdf exec mytool`:
  - exit code is `0`
  - stdout equals an exact value
- after `asdf uninstall mytool 2.0.0`:
  - exit code is `0`
  - dir `.atago-home/.asdf/installs/mytool/2.0.0` does not exist
- after `PATH="${workdir}/.atago-home/.asdf/shims:$PATH" mytool`:
  - exit code is `126`
  - stdout is empty
  - stderr contains `No preset version installed for command mytool`, `asdf install mytool 2.0.0`, matches `//\.tool-versions/`

#### Expected output
_expected stdout:_
```text
mytool 2.0.0
```
