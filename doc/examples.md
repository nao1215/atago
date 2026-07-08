# Examples

## By task

| You want to | Uses |
|-------------|------|
| [Start from a recorded run](cookbook.md#start-from-a-recorded-run) | `atago record`, then edit the generated spec |
| [Test a CLI that converts images](cookbook.md#test-a-cli-that-converts-images) | `image:` format/dimension/similarity asserts |
| [Test a CLI that generates a PDF](cookbook.md#test-a-cli-that-generates-a-pdf) | `pdf:` page/metadata/text asserts |
| [Test a CLI that generates files](cookbook.md#test-a-cli-that-generates-files) | `file:`/`dir:` asserts, byte-exact `equals_file` |
| [Pin exactly which files a command touches](cookbook.md#pin-exactly-which-files-a-command-touches) | `changes:` workdir delta |
| [Test error handling: exit codes and stderr](cookbook.md#test-error-handling-exit-codes-and-stderr) | `exit_code` (`not`/`in`), `stderr`, `exists: false` |
| [Feed stdin to a filter CLI](cookbook.md#feed-stdin-to-a-filter-cli) | `stdin:` inline / file / base64 |
| [Assert on JSON or YAML output](cookbook.md#assert-on-json-or-yaml-output) | `json:`/`yaml:` JSONPath matchers, numeric bounds |
| [Check what a CLI wrote to a database](cookbook.md#check-what-a-cli-wrote-to-a-database) | db runner, `query:` steps, `rows:` asserts |
| [Test an interactive prompt](cookbook.md#test-an-interactive-prompt) | `pty:` expect/send sessions, named keys |
| [Test a full-screen TUI](cookbook.md#test-a-full-screen-tui) | `screen:` rendered-frame asserts and snapshots |
| [Test an API-client CLI without the network](cookbook.md#test-an-api-client-cli-without-the-network) | `mock_servers:` + `mock:` request asserts |
| [Test a CLI that starts a server](cookbook.md#test-a-cli-that-starts-a-server) | `services:` with readiness probes |
| [Test graceful shutdown](cookbook.md#test-graceful-shutdown) | `signal:` steps with `wait:` |
| [Pin a generator's whole output tree](cookbook.md#pin-a-generators-whole-output-tree) | recursive `dir:` asserts, tree snapshots |
| [Pin output with a golden file](cookbook.md#pin-output-with-a-golden-file) | `snapshot:` + `scrub:` |
| [Poll an async result](cookbook.md#poll-an-async-result) | `retry:` until-assertions |
| [Bound how long a command may take](cookbook.md#bound-how-long-a-command-may-take) | `duration:` wall-clock bounds |
| [Clean up external state even when a step fails](cookbook.md#clean-up-external-state-even-when-a-step-fails) | `teardown:` steps sharing the store |
| [Run expensive setup once for the whole suite](cookbook.md#run-expensive-setup-once-for-the-whole-suite) | `suite.setup`, `${suitedir}` |
| [Run a scenario only where it can pass](cookbook.md#run-a-scenario-only-where-it-can-pass) | `tags`, `skip:`/`only:` gates |
| [Run the same scenario over many inputs](cookbook.md#run-the-same-scenario-over-many-inputs) | `matrix:` expansion |
| [Capture a value in one step and reuse it](cookbook.md#capture-a-value-in-one-step-and-reuse-it) | `store:` + `${name}` |
| [Isolate the test from the host environment](cookbook.md#isolate-the-test-from-the-host-environment) | `clear_env`, `pass_env`, `sandbox_home` |
| [Pin the help and misuse contract](cookbook.md#pin-the-help-and-misuse-contract) | `--help` content, unknown-flag exit code and stderr |
| [Test a CLI that reads environment variables](cookbook.md#test-a-cli-that-reads-environment-variables) | step `env:`, `clear_env` to prove the default |
| [Mask secrets in reports and snapshots](cookbook.md#mask-secrets-in-reports-and-snapshots) | `secrets:` masking as `***` |
| [Test a REPL](cookbook.md#test-a-repl) | `pty:` prompt-gated expect/send, EOF via `ctrl-d` |
| [Prove a command is idempotent](cookbook.md#prove-a-command-is-idempotent) | second-run `changes:` pinned to empty |
| [Compare two implementations of the same command](cookbook.md#compare-two-implementations-of-the-same-command) | `store:` + `equals: ${reference}` oracle |
| [Record an interactive session instead of scripting it](cookbook.md#record-an-interactive-session-instead-of-scripting-it) | `atago record --pty`, secrets become `${env:...}` |
| [Refresh snapshots when output legitimately changes](cookbook.md#refresh-snapshots-when-output-legitimately-changes) | `atago snapshot update`, `scrub:`, git-reviewable goldens |
| [Pin the final TUI frame with a screen snapshot](cookbook.md#pin-the-final-tui-frame-with-a-screen-snapshot) | `screen:` line/contains asserts and snapshots, `rows:`/`cols:` |
| [Pin the version and completion contracts](cookbook.md#pin-the-version-and-completion-contracts) | `matches:` on `--version`, completion output |
| [Simulate API failures offline](cookbook.md#simulate-api-failures-offline) | `mock_servers:` error routes, `mock:` call counting |
| [Test a download command offline](cookbook.md#test-a-download-command-offline) | mock route body, `equals_file` byte compare |
| [Verify server state after the CLI acts](cookbook.md#verify-server-state-after-the-cli-acts) | `services:` + observer `http:` step, `status`/`header`/`body` |
| [Send output streams to files](cookbook.md#send-output-streams-to-files) | `stdout_to:`/`stderr_to:` + file asserts |
| [Ship binary test data with fixtures](cookbook.md#ship-binary-test-data-with-fixtures) | fixture `from:`/`base64:`/`mode:` |
| [Test how the CLI treats symlinks](cookbook.md#test-how-the-cli-treats-symlinks) | fixture `symlink:` |
| [Test freshness logic with fixture timestamps](cookbook.md#test-freshness-logic-with-fixture-timestamps) | fixture `mtime:` + `changes:` delta |
| [Prove the tool is binary-safe](cookbook.md#prove-the-tool-is-binary-safe) | `stdin: {base64:}`, `stdout_to:`, `equals_file` |
| [Test the empty-input boundary](cookbook.md#test-the-empty-input-boundary) | empty fixtures and `stdin: ""` |
| [Prove multibyte text survives](cookbook.md#prove-multibyte-text-survives) | exact `equals:` on unicode round-trips |
| [Test TTY detection](cookbook.md#test-tty-detection) | the same command under `pty:` and `run:` |
| [Inspect an archive the CLI produced](cookbook.md#inspect-an-archive-the-cli-produced) | listing via a real tool + `contains`/`not_contains` |
| [Test the unreadable-input failure mode](cookbook.md#test-the-unreadable-input-failure-mode) | fixture `mode: "0000"` |
| [Run the command from a subdirectory](cookbook.md#run-the-command-from-a-subdirectory) | `cwd:` inside the workdir |
| [Write one spec that runs on all three OSes](cookbook.md#write-one-spec-that-runs-on-all-three-oses) | direct argv commands, `skip: {os:}` gates |
| [Compare releases of your CLI with a matrix](cookbook.md#compare-releases-of-your-cli-with-a-matrix) | `matrix:` over the command itself |
| [Tag scenarios and run a slice in CI](cookbook.md#tag-scenarios-and-run-a-slice-in-ci) | `tags:` + `--tag`/`--skip-tag` |
| [Set defaults once for the whole suite](cookbook.md#set-defaults-once-for-the-whole-suite) | `defaults:` for run steps and scenario env |
| [Assert a database migration's schema](cookbook.md#assert-a-database-migrations-schema) | db runner, catalog `query:` + `rows:` |
| [Run the CLI on a remote host over SSH](cookbook.md#run-the-cli-on-a-remote-host-over-ssh) | `ssh` runner as the observation point |
| [Check a gRPC dependency after the CLI acts](cookbook.md#check-a-grpc-dependency-after-the-cli-acts) | `grpc:` step, `grpc_status`/`message` asserts |
| [Verify a generated page in a real browser](cookbook.md#verify-a-generated-page-in-a-real-browser) | `cdp:` actions + `value:` assert |
| [Lay out specs for a growing suite](cookbook.md#lay-out-specs-for-a-growing-suite) | directory layout, tags, `atago doc`/`list` |
| [Prove a dry run changes nothing](cookbook.md#prove-a-dry-run-changes-nothing) | all-empty `changes:` delta |
| [Test config precedence](cookbook.md#test-config-precedence) | flag > env > file, hermetic `clear_env` rungs |
| [Abort a destructive command at its confirmation prompt](cookbook.md#abort-a-destructive-command-at-its-confirmation-prompt) | `pty:` "no" branch + empty `changes:` |
| [Prove color output turns off](cookbook.md#prove-color-output-turns-off) | `not_matches:` on ANSI escapes, `NO_COLOR` |
| [Assert a generated script is executable](cookbook.md#assert-a-generated-script-is-executable) | `file:` `executable:` + `contains:` |
| [Hunt down a flaky scenario](cookbook.md#hunt-down-a-flaky-scenario) | `--repeat`, `--retry-failed` (flaky is reported), `atago rerun` |
| [Troubleshooting](cookbook.md#troubleshooting) | the failures every new spec hits once, and the fix for each |

## By feature

| Example | Shows |
|---------|-------|
| [run_and_assert](../examples/run_and_assert.atago.yaml) | exit code (exact, `not`, `in: [0, 2]` sets), stdout/stderr matchers (`contains`, `equals`, `matches`/`not_matches`, `empty: true`/`false`, lists, `line`), combining `contains`/`not_contains`/`matches`/`not_matches` in one block, multi-target asserts |
| [shell_and_redirect](../examples/shell_and_redirect.atago.yaml) | `shell: true` vs direct argv execution, `stdout_to`/`stderr_to` redirects |
| [json_and_yaml](../examples/json_and_yaml.atago.yaml) | JSONPath assertions, numeric bounds (`gt`/`lte`), a list of checks under one `json:`/`yaml:`, the `yaml` matcher |
| [files_and_fixtures](../examples/files_and_fixtures.atago.yaml) | input fixtures (inline `content:` and `base64:`), `file` and `dir` assertions, byte-exact `equals`/`equals_file` round-trip checks |
| [store_and_variables](../examples/store_and_variables.atago.yaml) | capturing values into `${name}`, `${workdir}`, `${env:NAME}` host-environment reads, the `$${...}` literal escape |
| [teardown](../examples/teardown.atago.yaml) | cleanup steps that always run — pass or fail — sharing the scenario's variables |
| [hermetic_env](../examples/hermetic_env.atago.yaml) | `clear_env: true` starts commands from an empty environment, `pass_env` re-admits an allowlist of host variables, `sandbox_home: true` isolates HOME and per-OS config/cache dirs |
| [extend_host_env](../examples/extend_host_env.atago.yaml) | extend an inherited variable in a scenario `env:` value with `${env:NAME}` (e.g. `PATH: "${workdir}/stub:${env:PATH}"`) instead of replacing it — put a stub binary earlier on PATH while real tools still resolve |
| [timeouts](../examples/timeouts.atago.yaml) | the built-in 60s default step timeout, `suite.timeout`, per-step overrides, and the `timeout: "0"` escape hatch |
| [stdin](../examples/stdin.atago.yaml) | stdin sources: inline text, `stdin: {file: ...}` from a workdir file, and binary input via `stdin: {base64: ...}` |
| [matrix](../examples/matrix.atago.yaml) | one template scenario expanded per parameter row |
| [mock_server](../examples/mock_server.atago.yaml) | test API-client CLIs offline: `mock_servers` serve canned routes, record every request, and `mock:` asserts what the client actually sent |
| [pty](../examples/pty.atago.yaml) | interactive testing in a real pseudo-terminal: expect/send sessions, named keys (`send: {key: enter}`), TTY-detection (scenarios use POSIX-only inner commands) |
| [pty_portable](../examples/pty_portable.atago.yaml) | the same `pty` mechanism on every OS — Linux, macOS, and Windows (ConPTY): drive a self-terminating command, match its output, assert the rendered screen |
| [pty_screen](../examples/pty_screen.atago.yaml) | TUI testing on the RENDERED terminal screen: vt100 emulation, row-addressed asserts, and screen snapshots (scenarios use POSIX-only inner commands) |
| [retry](../examples/retry.atago.yaml) | polling a command until an assertion passes |
| [snapshot](../examples/snapshot.atago.yaml) | golden-file testing with normalized output |
| [scrub](../examples/scrub.atago.yaml) | `scrub:` rewrites volatile output patterns (auto-increment IDs, request identifiers, epoch times) to a placeholder before a snapshot compares — the flake-killer the built-in normalizers do not cover |
| [duration](../examples/duration.atago.yaml) | bound a step's wall-clock time with `duration: {lt: 2s, gte: 100ms}` (use generous bounds — CI runners are slow) |
| [dir_tree](../examples/dir_tree.atago.yaml) | recursive dir assertions and directory-tree snapshots: pin a generator's whole output tree with one golden manifest |
| [changes](../examples/changes.atago.yaml) | `changes:` pins the exact workdir delta of a step — which files it created, modified, and deleted, and nothing else |
| [services](../examples/services.atago.yaml) | background servers: readiness probes, `ready.store`, teardown |
| [signal](../examples/signal.atago.yaml) | `signal:` steps deliver SIGTERM/SIGHUP/... to a managed service's process group for graceful-shutdown and reload testing (POSIX-only) |
| [defaults](../examples/defaults.atago.yaml) | sharing `shell`/`env`/`service` fragments across scenarios |
| [suite_setup](../examples/suite_setup.atago.yaml) | once-per-suite bootstrap: ordered setup steps, suite-wide `service:` steps, `${suitedir}`, suite env, always-run suite teardown |
| [select_skip_only](../examples/select_skip_only.atago.yaml) | tags, and gating scenarios by OS, env var, or a probe command |
| [db](../examples/db.atago.yaml) | SQL via the bundled SQLite driver, `rows` assertions, value binding |
| [image_and_pdf](../examples/image_and_pdf.atago.yaml) | image format/dimension/similarity checks, PDF page/metadata/text checks |
| [http](../examples/http.atago.yaml) | HTTP requests (`json:`, raw `body:`, form/multipart uploads, `body_file`), status/body assertions, token capture, `retry` polling, redirect assertions, `body_to` downloads, network allowlist |
| [ssh](../examples/ssh.atago.yaml) | running commands on a remote host |
| [grpc](../examples/grpc.atago.yaml) | unary gRPC calls via server reflection |
| [browser](../examples/browser.atago.yaml) | headless-Chrome flows and screenshots |
