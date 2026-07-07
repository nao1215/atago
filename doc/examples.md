# Examples

Two indexes into the same material: the [cookbook](cookbook.md) answers "how do
I test X" with a copyable spec, and the feature table links one commented,
runnable spec per feature. The specs under [examples/](../examples/) are
loaded, validated, and (where they need no external server) executed in CI on
Linux, macOS, and Windows, so they cannot drift from the implementation.

## By task

| You want to | Uses |
|-------------|------|
| [Test a CLI that converts images](cookbook.md#test-a-cli-that-converts-images) | `image:` format/dimension/similarity asserts |
| [Test a CLI that generates a PDF](cookbook.md#test-a-cli-that-generates-a-pdf) | `pdf:` page/metadata/text asserts |
| [Test a CLI that generates files](cookbook.md#test-a-cli-that-generates-files) | `file:`/`dir:` asserts, byte-exact `equals_file` |
| [Pin exactly which files a command touches](cookbook.md#pin-exactly-which-files-a-command-touches) | `changes:` workdir delta |
| [Test error handling: exit codes and stderr](cookbook.md#test-error-handling-exit-codes-and-stderr) | `exit_code` (`not`/`in`), `stderr`, `exists: false` |
| [Feed stdin to a filter CLI](cookbook.md#feed-stdin-to-a-filter-cli) | `stdin:` inline / file / base64 |
| [Test an interactive prompt](cookbook.md#test-an-interactive-prompt) | `pty:` expect/send sessions, named keys |
| [Test a full-screen TUI](cookbook.md#test-a-full-screen-tui) | `screen:` rendered-frame asserts and snapshots |
| [Test an API-client CLI without the network](cookbook.md#test-an-api-client-cli-without-the-network) | `mock_servers:` + `mock:` request asserts |
| [Test a CLI that starts a server](cookbook.md#test-a-cli-that-starts-a-server) | `services:` with readiness probes |
| [Pin output with a golden file](cookbook.md#pin-output-with-a-golden-file) | `snapshot:` + `scrub:` |
| [Poll an async result](cookbook.md#poll-an-async-result) | `retry:` until-assertions |
| [Run the same scenario over many inputs](cookbook.md#run-the-same-scenario-over-many-inputs) | `matrix:` expansion |
| [Capture a value in one step and reuse it](cookbook.md#capture-a-value-in-one-step-and-reuse-it) | `store:` + `${name}` |
| [Isolate the test from the host environment](cookbook.md#isolate-the-test-from-the-host-environment) | `clear_env`, `pass_env`, `sandbox_home` |

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
