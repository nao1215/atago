// Package spec defines the typed in-memory model for a atago YAML file.
//
// The model is intentionally decoupled from the raw YAML representation
// . Loading and validation live in internal/loader; this package
// only declares the shapes and the custom unmarshalers needed for the few
// polymorphic nodes in the format (see schema/atago.schema.json).
package spec

// Spec is a complete atago file.
type Spec struct {
	Version     string            `yaml:"version"`
	Suite       Suite             `yaml:"suite"`
	Runners     map[string]Runner `yaml:"runners,omitempty"`
	Permissions *Permissions      `yaml:"permissions,omitempty"`
	Secrets     []string          `yaml:"secrets,omitempty"`
	// Scrub declares spec-wide regex→placeholder rewrites applied to captured
	// output before it is compared against (or written to) a snapshot golden.
	// Where `secrets:` masks known values, `scrub:` normalizes volatile
	// patterns the built-in normalizers do not know about — auto-increment
	// IDs, request identifiers, custom timestamps — so a flaky snapshot becomes
	// deterministic (#137). Rules apply in order, after secret masking and
	// before the built-in ANSI/UUID/timestamp/port/path normalization.
	Scrub []ScrubRule `yaml:"scrub,omitempty"`
	// Defaults declares spec-wide default fragments merged into every matching
	// element at load time. It is authoring sugar only: the loader
	// expands it into the concrete scenario/step/service model before validation,
	// so nothing downstream (engine, manifest, explain) ever observes `defaults`.
	Defaults  *Defaults  `yaml:"defaults,omitempty"`
	Scenarios []Scenario `yaml:"scenarios"`

	// FixturesDir is the committed fixture tree a directory manifest points at
	// (#394), absolute. Set by the loader from atago.project.yaml, never
	// authored in a spec, and exposed to every scenario as ${fixtures}.
	FixturesDir string `yaml:"-"`
	// Subject is the binary under test declared by that manifest (#393), when
	// it declares one. It is recorded here — rather than staying inside the
	// loader — because the build is a command that runs on the host before any
	// scenario, with the invoking environment and optionally through the shell,
	// and the spec summaries have to be able to describe and flag it.
	Subject *Subject `yaml:"-"`
	// ProjectPath is the manifest that applied to this spec, if any (#392). Set
	// by the loader, never authored; the read-only commands print it, because
	// configuration that applies to a file without appearing in it has to be
	// visible somewhere.
	ProjectPath string `yaml:"-"`
}

// ScrubRule is one declarative output-normalization rule (#137): every substring
// matching Pattern (a Go RE2 regexp) is replaced with Placeholder, literally
// (no `$1` expansion), before snapshot comparison. Ordering is significant —
// earlier rules see the raw text, later rules see the output of earlier ones.
type ScrubRule struct {
	// Pattern is a Go regular expression (regexp/syntax). Required and must compile.
	Pattern string `yaml:"pattern"`
	// Placeholder is the literal replacement text, e.g. "<ID>". May be empty to
	// delete matches outright.
	Placeholder string `yaml:"placeholder"`
}

// Defaults holds the top-level `defaults:` block. Each fragment is
// merged into the concrete model at load time to cut repetition without adding a
// runtime model: `run` layers under every `run` step, `scenario.env` under every
// scenario env, and `service` under every service. The environment-shaping
// subset of `run` (env, clear_env, pass_env, sandbox_home) also layers onto
// `pty` steps — a pty step shares that surface — while run-only fields (runner,
// shell, cwd, timeout, stdin, redirects, retry) never reach pty steps (#77).
//
// Merge rules (applied by the loader): an explicitly-authored value always wins;
// maps shallow-merge (the authored key wins per key); a nil pointer / empty
// string counts as "unset" and takes the default; a boolean default is OR-ed in
// (a Go bool has no unset state, so a defaulted `shell: true` cannot be turned
// off per element). It is not a macro/include system: there is no substitution
// of one element by another and no expression language.
type Defaults struct {
	Run      *Run              `yaml:"run,omitempty"`
	Scenario *ScenarioDefaults `yaml:"scenario,omitempty"`
	Service  *Service          `yaml:"service,omitempty"`
}

// ScenarioDefaults holds the scenario-level default fragments: the env every
// scenario shares, and the gate every scenario is selected by.
//
// Env shallow-merges beneath each scenario's own `env`. Only and Skip are taken
// whole by a scenario that does not state its own, which is what a probe-first
// suite needs: "these scenarios exist only where the tool is installed" is a
// property of the file, not of each scenario in it, and repeating it on every
// scenario is how a suite ends up with some scenarios gated and some not — the
// ungated ones then error on a machine without the tool instead of skipping.
type ScenarioDefaults struct {
	Env map[string]string `yaml:"env,omitempty"`
	// Only is the default selection gate: a scenario without its own `only:`
	// runs only when this condition holds.
	Only *Condition `yaml:"only,omitempty"`
	// Skip is the default exclusion gate: a scenario without its own `skip:` is
	// skipped when this condition holds.
	Skip *Condition `yaml:"skip,omitempty"`
}

// Suite groups scenarios under a name.
type Suite struct {
	Name string `yaml:"name"`
	// Description is optional prose explaining what the suite as a whole
	// guarantees, rendered under the suite heading by `atago doc`. It is
	// documentation data, not behavior: nothing in the engine reads it, and it
	// never takes part in ${name} expansion, so it is exactly the literal text
	// the author wrote. Multi-line Markdown is allowed; an empty or
	// whitespace-only value renders nothing.
	//
	// It exists because the narrative every spec author already writes as a
	// file-top YAML comment — what this suite covers and why — is invisible to
	// every reader of the generated docs. Comments stay what they are (notes for
	// whoever maintains the spec: install steps, migration history, TODOs);
	// `description` is the part meant to be published.
	Description string `yaml:"description,omitempty"`
	// Timeout is the suite-level default step timeout (#17): every run, http,
	// query, and grpc step without a more specific timeout (step > runner >
	// defaults.run > suite) is bounded by this Go duration. "0"/"0s" disables
	// the bound. When no level configures one, a built-in 60s default applies
	// so a hanging command fails instead of stalling the run forever.
	Timeout string `yaml:"timeout,omitempty"`
	// Env is exported to every scenario, setup, and teardown step in the
	// suite (a scenario's own env wins per key). Values get ${name} expansion
	// against the suite store (${suitedir}, ${env:NAME}, setup stores).
	Env map[string]string `yaml:"env,omitempty"`
	// Setup steps run ONCE before any scenario, in declaration order, inside a
	// suite-scoped scratch directory exposed as ${suitedir} (#7). They exist
	// so bootstrap shell scripts (build a helper binary, start a shared peer,
	// warm a cache) become spec YAML. Allowed kinds: fixture, run, store,
	// assert, and — unique to this block — `service:`, which starts a
	// suite-wide background process at that exact point in the sequence.
	// Values captured here (store, ready.store) seed every scenario's store.
	// A failing setup step marks every scenario errored; no scenario runs.
	Setup []Step `yaml:"setup,omitempty"`
	// Teardown steps always run after the last scenario — pass, fail, error,
	// or interrupt — while suite services are still up (services stop last,
	// LIFO). Failures are reported but never change the suite verdict.
	Teardown []Step `yaml:"teardown,omitempty"`
}

// Runner declares a named runner. cmd is the implicit default for run steps;
// http supplies base_url for http steps; db supplies a dsn for query steps.
type Runner struct {
	Type    string `yaml:"type"`
	Cwd     string `yaml:"cwd,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
	BaseURL string `yaml:"base_url,omitempty"`
	// DSN is the data source for a db runner, e.g. "sqlite:./app.db",
	// "postgres://user:pass@host/db", or "mysql://user:pass@host:3306/db". The
	// driver is inferred from the scheme unless Driver is set.
	DSN string `yaml:"dsn,omitempty"`
	// Driver, when set, names the database/sql driver explicitly (sqlite,
	// postgres, or mysql), overriding scheme inference.
	Driver string `yaml:"driver,omitempty"`

	// SSH runner fields. A `run` step naming an ssh runner executes its
	// command on the remote host, capturing stdout/stderr/exit like a local run.
	Host       string `yaml:"host,omitempty"`        // host or host:port (default port 22)
	User       string `yaml:"user,omitempty"`        // login user
	Password   string `yaml:"password,omitempty"`    // password auth (prefer a secret)
	KeyFile    string `yaml:"key_file,omitempty"`    // path to a private key for key auth
	KnownHosts string `yaml:"known_hosts,omitempty"` // known_hosts file verifying the host key
	// InsecureHostKey must be set to true to connect without a known_hosts file
	// (disabling host-key verification). Without it, an empty known_hosts is a
	// configuration error rather than a silent MITM-able default (issue #17).
	InsecureHostKey bool `yaml:"insecure_host_key,omitempty"`

	// gRPC runner fields. A `grpc` step calls a unary method on the
	// target, resolving the schema via server reflection (no compiled stubs).
	Target string `yaml:"target,omitempty"` // host:port of the gRPC server
	TLS    bool   `yaml:"tls,omitempty"`    // use TLS (default plaintext)

	// Browser runner fields. A minimal, black-box configuration surface
	// for the CDP runner; all are optional and preserve the zero-config headless
	// default when omitted.
	//
	// Headless is a *bool so an unset field keeps the headless default while an
	// explicit `headless: false` runs headed for debugging.
	Headless *bool `yaml:"headless,omitempty"`
	// ExecPath points at a specific Chrome/Chromium binary instead of the one
	// chromedp discovers on PATH.
	ExecPath string `yaml:"exec_path,omitempty"`
	// BrowserArgs is a small escape hatch of extra Chrome launch flags (each a bare
	// flag name like "disable-gpu" or "window-size=1280,720", without the leading
	// "--") for CI environments that need them.
	BrowserArgs []string `yaml:"browser_args,omitempty"`
}

// Permissions carries the security policy for a spec.
type Permissions struct {
	Network *NetworkPolicy `yaml:"network,omitempty"`
}

// NetworkPolicy lists hosts that scenarios may contact.
type NetworkPolicy struct {
	Allow []string `yaml:"allow,omitempty"`
}

// Scenario is a single named behavior under test.
// ExpectFail marks a scenario as a documented known bug (#395).
//
// The mapping form is the only one: a bare boolean would let an expected
// failure into a suite with no record of WHY, and the next reader would have no
// way to tell a deliberate reproduction from a test somebody gave up on.
type ExpectFail struct {
	// Reason says what is broken, in the author's words. Required.
	Reason string `yaml:"reason"`
	// Issue is where the bug is tracked. Optional, but it is what turns an
	// XPASS into an actionable message ("this is fixed — close <url> and move
	// the scenario into the guarded suite").
	Issue string `yaml:"issue,omitempty"`
}

type Scenario struct {
	Name string `yaml:"name"`
	// Description is optional prose explaining the behavior this scenario pins
	// down and why it matters, rendered under the scenario heading by `atago
	// doc`. Same contract as Suite.Description: documentation only, never
	// expanded, never read by the engine. Use it where the name cannot carry the
	// background — the regression it guards, the failure mode it rules out — not
	// to restate the name.
	Description string     `yaml:"description,omitempty"`
	Tags        []string   `yaml:"tags,omitempty"`
	Skip        *Condition `yaml:"skip,omitempty"`
	Only        *Condition `yaml:"only,omitempty"`
	// ExpectFail declares that this scenario documents a KNOWN bug: it is
	// expected to fail, and failing is reported as XFAIL without failing the
	// run (#395). Passing is XPASS, which does fail the run — a spec that
	// silently starts passing is one nobody promotes, and the failure is what
	// says the fix landed and the scenario belongs in the guarded suite now.
	ExpectFail *ExpectFail       `yaml:"expect_fail,omitempty"`
	Env        map[string]string `yaml:"env,omitempty"`
	// Matrix, when set, makes this scenario a template: the loader expands it into
	// one concrete scenario per row before validation.
	// Each row's key/value pairs are seeded as ${name} variables for that instance.
	Matrix []map[string]string `yaml:"matrix,omitempty"`
	// Vars holds the bound matrix row for an expanded scenario instance. It is
	// populated by matrix expansion, never decoded from YAML, and seeded into the
	// scenario store before steps run.
	Vars map[string]string `yaml:"-"`
	// SourceIndex is the scenario's index in the authored (pre-matrix-expansion)
	// scenarios list. The loader sets it before matrix expansion, so every
	// instance expanded from one matrix template shares the template's authored
	// index — and therefore its source location (#80). Never decoded from YAML.
	SourceIndex int `yaml:"-"`
	// Services are background processes started before the scenario's steps run
	// and terminated when the scenario ends. They let a spec exercise a
	// CLI that talks to a peer (a TCP client, an API consumer) by standing up that
	// peer for the duration of the scenario.
	Services []Service `yaml:"services,omitempty"`
	// MockServers are declarative stub HTTP servers (#24): each serves canned
	// routes on an ephemeral loopback port, records every incoming request
	// for the `mock:` assertion target, and seeds ${<name>.url} /
	// ${<name>.port} into the store before steps run — so an API-client CLI
	// can be tested fully offline. Lifecycle mirrors services: started before
	// steps, stopped LIFO at scenario end.
	MockServers []MockServer `yaml:"mock_servers,omitempty"`
	Steps       []Step       `yaml:"steps"`
	// Teardown steps always run after Steps — whether the scenario passed,
	// failed, errored, or was interrupted — and share its variable store, so a
	// resource id captured with `store` flows into the cleanup request. They
	// exist for external side effects the isolated workdir cannot undo (rows in
	// a real database, resources created via an API, containers started by a run
	// step). A teardown failure is reported but does not change the scenario's
	// verdict: the behavior under test is decided by Steps alone.
	Teardown []Step `yaml:"teardown,omitempty"`
}

// Subject is the binary under test a directory manifest declares (#393),
// reduced to what a spec summary needs: the name specs invoke it by, and the
// command that builds it before any scenario runs.
//
// The build command is the reason this reaches the spec model at all. It runs
// on the host with the invoking environment, optionally through the shell, and
// nothing described or flagged it — a `curl … > ${artifact}` build was as
// invisible to review as an empty manifest.
type Subject struct {
	Name    string
	Command string
	Shell   bool
}

// Condition gates a scenario by platform, environment, or a probe command
// . For OS, skip/only compare against the host. For Env, the
// condition is true when the named environment variable is non-empty:
// `skip: { env: X }` skips when X is set; `only: { env: X }` runs only when X is
// set. For Command, the condition is true when the probe command succeeds (exits
// 0): `skip: { command: X }` skips when X succeeds; `only: { command: X }` runs
// only when X succeeds. The probe runs through the shell.
type Condition struct {
	OS      string `yaml:"os,omitempty"`
	Env     string `yaml:"env,omitempty"`
	Command string `yaml:"command,omitempty"`
}

// Step is exactly one of the action keys. Loader enforces the one-of rule.
type Step struct {
	Fixture *Fixture `yaml:"fixture,omitempty"`
	Run     *Run     `yaml:"run,omitempty"`
	HTTP    *HTTP    `yaml:"http,omitempty"`
	Query   *Query   `yaml:"query,omitempty"`
	GRPC    *GRPC    `yaml:"grpc,omitempty"`
	CDP     *CDP     `yaml:"cdp,omitempty"`
	Assert  *Assert  `yaml:"assert,omitempty"`
	Store   *Store   `yaml:"store,omitempty"`
	// Service starts a suite-wide background process (#7). It is valid only
	// inside suite.setup — the loader rejects it anywhere else — so its
	// position in the setup sequence controls ordering relative to the run
	// steps around it (build the binary, then serve it, then warm a cache).
	Service *Service `yaml:"service,omitempty"`
	// PTY runs one command inside a real pseudo-terminal and drives it with a
	// declarative expect/send session (#8) — for CLIs that branch on TTY-ness
	// or present an interactive prompt. Runs on POSIX (a real pty) and on Windows
	// (a ConPTY, Windows 10 1809+); the loader accepts the step everywhere.
	PTY *PTY `yaml:"pty,omitempty"`
	// Signal sends a named POSIX signal to a managed service (#23) — the
	// race-free, handle-based alternative to `kill`/`killall` shell hacks for
	// graceful-shutdown and signal-handling tests. POSIX-only like pty: the
	// loader accepts the step everywhere and the engine reports a clear error
	// on Windows.
	Signal *Signal `yaml:"signal,omitempty"`
	// MockServer starts a suite-wide stub HTTP server (#24). Like `service:`
	// it is valid only inside suite.setup, so its position in the sequence
	// controls ordering; its recorded requests and ${<name>.url} seed every
	// scenario.
	MockServer *MockServer `yaml:"mock_server,omitempty"`
}

// Query runs a SQL statement through a named db runner. The result
// rows (for a SELECT) are captured as JSON for the `rows` assertion target and
// `store from.rows`; a non-row statement records its affected-row count.
type Query struct {
	Runner string `yaml:"runner"`
	SQL    string `yaml:"sql"`
}

// GRPC calls a unary gRPC method through a named grpc runner. The
// response message is captured as JSON for the `message` assertion target and
// `store from.message`; the status code feeds the `grpc_status` target.
type GRPC struct {
	Runner string            `yaml:"runner"`
	Method string            `yaml:"method"` // "pkg.Service/Method"
	Header map[string]string `yaml:"header,omitempty"`
	JSON   any               `yaml:"json,omitempty"` // request message
}
