# mimixbox ShellSpec → atago migration guide

This directory holds atago ports of mimixbox's ShellSpec integration tests.
Source suite: `~/ghq/github.com/nao1215/mimixbox/test/it/` — the specs are
`test/it/spec/<applet>_spec.sh`, each `Include`ing a helper
`test/it/<category>/<applet>_test.sh` that defines the `Test*` shell functions.

## Output layout
One atago file per applet: `test/e2e/mimixbox/<category>/<applet>.atago.yaml`.
Use the SAME `<category>` directory the helper lives in (shellutils, textutils,
fileutils, util-linux, netutils, procps, console-tools, loginutils, archival,
securityutils, findutils, editors, debianutils, printutils, mailutils, compat,
embedded, runit, pmutils, jokeutils, games, coreutils).

## Translation rules
- `Describe '...'` + `It '...'` → one atago `scenario` with a descriptive `name`.
- `When call TestX` → inline the body of `TestX` (from the `_test.sh` helper) as
  the `run` step's command.
  - **Simple single command** (no pipe/redirect/heredoc/`;`/var-assignment):
    use **direct exec** — `run: { command: "<cmd>" }` (NO `shell: true`). This
    matches ShellSpec's forked `When call` and avoids a mimixbox quirk where its
    own `sh` changes the exit code of single commands.
  - **Needs a pipe `|`, redirect `>`, heredoc `<<`, `;`, or VAR=val**: use
    `run: { shell: true, command: "<pipeline>" }`.
- `${MIMIXBOX_IT_ROOT}` → `${workdir}` (each scenario has its own isolated
  workdir; atago expands `${workdir}` in commands AND in assertion values).
- Files the ShellSpec `Setup` creates: prefer declarative `fixture` steps:
  `- fixture: { file: foo.txt, content: "..." }`. Reference them by relative
  name (commands run in the workdir). Drop `CleanUp` (workdir is auto-removed).
- Assertions:
  - `The status should be success` → `assert: { exit_code: 0 }`
  - `The status should be failure` → `assert: { exit_code: { not: 0 } }`
  - `The status should equal N` → `assert: { exit_code: N }`
  - `The output should equal "X"` → `assert: { stdout: { equals: "X" } }`
    (atago `equals` ignores a single trailing newline, like ShellSpec)
  - `The output should include "X"` → `stdout: { contains: "X" }`
  - `The output should not include "X"` → `stdout: { not_contains: "X" }`
  - `The error should equal/include` → `stderr: { equals|contains: ... }`
  - `The output should be blank/empty` → `stdout: { empty: true }`
  - `The line N of output should equal "X"` → `stdout: { line: N, equals: "X" }`
  - `The output should match pattern 'GLOB'` → `stdout: { matches: "REGEX" }`
    (convert the shell GLOB to a Go regex: `*`→`.*`, `?`→`.`, anchor as needed)
  - `Skip if [ -z "$VAR" ]` / `Skip if`-on-env → `only: { env: VAR }` or
    `skip: { env: VAR }`. Platform skips → `skip: { os: windows }` etc.

## Whitespace-exact output
`stdout.equals` with a YAML `|` block scalar consumes the first line's leading
spaces. When the expected output has **leading spaces or tabs** (e.g. `cat -n`,
columnar output), use a **double-quoted scalar with explicit escapes** instead:
`equals: "     1\tsh\n     2\tash"`. Get the exact bytes with
`<applet> ... | cat -A`.

## NUL / binary output
For `-z`/`--zero` style NUL-separated output, atago captures raw bytes (the
shell's `$(...)` in ShellSpec strips NULs). Assert with a regex:
`matches: "^a\\x00b\\x00$"`.

## Verify every file (REQUIRED)
A prebuilt env is staged at `/tmp/mbenv` (atago binary + applet bin dir). After
writing a spec, run it and make it green:
```sh
PATH="/tmp/mbenv/applets:$PATH" /tmp/mbenv/atago run <path-to-your.atago.yaml>
```
If `/tmp/mbenv` is missing, rebuild it:
```sh
mkdir -p /tmp/mbenv/applets
( cd ~/ghq/github.com/nao1215/atago && go build -o /tmp/mbenv/atago . )
( cd ~/ghq/github.com/nao1215/mimixbox && go build -trimpath -o /tmp/mbenv/applets/mimixbox ./cmd/mimixbox && /tmp/mbenv/applets/mimixbox --full-install /tmp/mbenv/applets >/dev/null )
```

## When an applet can't run in this sandbox
Some applets need root, devices, or network (parts of util-linux/netutils/procps).
If the ShellSpec test only checks help/usage/error text, that ports fine. If it
needs privileges the sandbox lacks AND the ShellSpec didn't guard it, gate the
scenario with `skip:` and note why in a comment. Never assert output you can't
reproduce — match the real applet's behavior, captured from `/tmp/mbenv`.

## Run the whole suite
`make dogfood-mimixbox` (from the atago repo root) builds everything fresh and
runs all specs under this directory with `--parallel 8`.
