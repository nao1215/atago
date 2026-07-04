// Package spec defines the typed in-memory model for a atago YAML file.
//
// The model is intentionally decoupled from the raw YAML representation
// . Loading and validation live in internal/loader; this package
// only declares the shapes and the custom unmarshalers needed for the few
// polymorphic nodes in the format (see schema/atago.schema.json).
package spec

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

// Spec is a complete atago file.
type Spec struct {
	Version     string            `yaml:"version"`
	Suite       Suite             `yaml:"suite"`
	Runners     map[string]Runner `yaml:"runners,omitempty"`
	Permissions *Permissions      `yaml:"permissions,omitempty"`
	Secrets     []string          `yaml:"secrets,omitempty"`
	// Defaults declares spec-wide default fragments merged into every matching
	// element at load time. It is authoring sugar only: the loader
	// expands it into the concrete scenario/step/service model before validation,
	// so nothing downstream (engine, manifest, explain) ever observes `defaults`.
	Defaults  *Defaults  `yaml:"defaults,omitempty"`
	Scenarios []Scenario `yaml:"scenarios"`
}

// Defaults holds the top-level `defaults:` block. Each fragment is
// merged into the concrete model at load time to cut repetition without adding a
// runtime model: `run` layers under every `run` step, `scenario.env` under every
// scenario env, and `service` under every service.
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

// ScenarioDefaults holds the scenario-level default fragments. Only `env` is
// supported — the highest-value duplication — and it shallow-merges beneath each
// scenario's own `env`.
type ScenarioDefaults struct {
	Env map[string]string `yaml:"env,omitempty"`
}

// Suite groups scenarios under a name.
type Suite struct {
	Name string `yaml:"name"`
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
type Scenario struct {
	Name string            `yaml:"name"`
	Tags []string          `yaml:"tags,omitempty"`
	Skip *Condition        `yaml:"skip,omitempty"`
	Only *Condition        `yaml:"only,omitempty"`
	Env  map[string]string `yaml:"env,omitempty"`
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

// Service is a background process started before a scenario's steps run and
// terminated (with its whole process group) when the scenario ends.
// It runs in the scenario workdir and gets the same ${name} expansion and
// scenario-env layering as a run step.
type Service struct {
	// Name identifies the service in diagnostics; it must be unique per scenario.
	Name string `yaml:"name"`
	// Command is the program to run. Like a run step it is tokenized and exec'd
	// directly unless Shell is set.
	Command string `yaml:"command"`
	// Shell runs Command through the POSIX shell (pipes/redirects/${}), matching
	// run.shell semantics. A pointer so an authored `shell: false` is
	// distinguishable from "unset" when `defaults.service.shell` is layered in.
	Shell *bool `yaml:"shell,omitempty"`
	// Cwd is the working directory relative to the scenario workdir (default: the
	// workdir itself).
	Cwd string `yaml:"cwd,omitempty"`
	// Env sets process environment variables (layered over the scenario env and
	// the inherited host environment), with ${name} expansion on the values.
	Env map[string]string `yaml:"env,omitempty"`
	// ClearEnv starts the service from an empty environment instead of
	// inheriting the host environment (#16). A pointer so an authored
	// `clear_env: false` is distinguishable from "unset" when
	// `defaults.service.clear_env` is layered in.
	ClearEnv *bool `yaml:"clear_env,omitempty"`
	// PassEnv copies the listed host variables into the cleared environment
	// (#16). Only meaningful with ClearEnv; unset host variables are skipped.
	PassEnv []string `yaml:"pass_env,omitempty"`
	// Ready declares how to wait until the service is accepting work before the
	// steps run. When omitted, the steps start as soon as the process is spawned.
	Ready *Ready `yaml:"ready,omitempty"`
}

// MockServer is a declarative stub HTTP server (#24). It listens on
// 127.0.0.1 with an ephemeral port, serves Routes matched on exact
// method+path, and records every incoming request (matched or not — an
// unmatched request answers 404 and is still recorded) for the `mock:`
// assertion target.
type MockServer struct {
	// Name identifies the server for ${<name>.url} seeding and mock asserts;
	// unique per scenario (and among suite-wide mock servers).
	Name string `yaml:"name"`
	// Routes are the canned responses, matched top-down on exact method+path
	// (query string excluded; deliberately no patterns).
	Routes []MockRoute `yaml:"routes,omitempty"`
}

// MockRoute is one canned response (#24). At most one of JSON/Body/BodyFile
// supplies the payload; Status defaults to 200.
type MockRoute struct {
	// Method is the HTTP method to match (case-insensitive).
	Method string `yaml:"method"`
	// Path is the exact request path to match (query string excluded).
	Path string `yaml:"path"`
	// Status is the response status (default 200).
	Status int `yaml:"status,omitempty"`
	// JSON is an inline response document, marshaled with
	// Content-Type: application/json.
	JSON any `yaml:"json,omitempty"`
	// Body is an inline text response body.
	Body string `yaml:"body,omitempty"`
	// BodyFile reads the response body from a spec-relative file (confined to
	// the spec directory, like snapshot goldens), read at request time.
	BodyFile string `yaml:"body_file,omitempty"`
	// Header sets extra response headers.
	Header map[string]string `yaml:"header,omitempty"`
	// Delay sleeps this Go duration before responding — for retry testing.
	Delay string `yaml:"delay,omitempty"`
}

// MockAssert checks what the CLI under test actually sent to a mock server
// (#24). Records are filtered by the optional Path/Method, then Count pins
// the exact number of matching requests (without Count, at least one must
// exist); Header and Body apply to the LAST matching request.
type MockAssert struct {
	// Name references a declared mock server.
	Name string `yaml:"name"`
	// Path filters recorded requests by exact path.
	Path string `yaml:"path,omitempty"`
	// Method filters recorded requests by method (case-insensitive).
	Method string `yaml:"method,omitempty"`
	// Count asserts the exact number of matching requests.
	Count *int `yaml:"count,omitempty"`
	// Header matches a header of the last matching request.
	Header *HeaderMatch `yaml:"header,omitempty"`
	// Body matches the body of the last matching request with the stream
	// matchers (json path, contains, ...).
	Body *StreamAssert `yaml:"body,omitempty"`
}

// Ready is a service readiness probe. Exactly one of File/Port/Log/
// Delay decides when the service is considered up; Timeout bounds the wait.
type Ready struct {
	// File waits until this workdir-relative file exists and is non-empty — the
	// canonical pattern for a server that publishes its (ephemeral) listen address
	// to a file once bound.
	File string `yaml:"file,omitempty"`
	// Port waits until this TCP address (host:port) accepts a connection.
	Port string `yaml:"port,omitempty"`
	// Log waits until the service's combined stdout/stderr matches this regexp.
	Log string `yaml:"log,omitempty"`
	// Delay simply waits this Go duration (e.g. "200ms") — a last resort when no
	// observable readiness signal exists.
	Delay string `yaml:"delay,omitempty"`
	// Timeout bounds the readiness wait as a Go duration (default "5s").
	Timeout string `yaml:"timeout,omitempty"`
	// Store, used with File, captures the ready file's trimmed content into
	// ${<Store>} so steps can reference a dynamically-chosen address or port.
	Store string `yaml:"store,omitempty"`
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
	// or present an interactive prompt. POSIX-only for now; the loader accepts
	// the step everywhere and the engine reports a clear error on Windows.
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

// Signal targets a declared service (scenario or suite) by name and delivers
// a POSIX signal to its whole process group (#23), consistent with the
// teardown kill semantics. Wait optionally blocks until the process exits.
type Signal struct {
	// Service names the target: a service declared in the scenario's
	// services: list or started by a suite.setup service: step. Unknown names
	// are a load-time error listing the declared services.
	Service string `yaml:"service"`
	// Signal is the POSIX signal name: TERM, INT, HUP, USR1, USR2, or KILL
	// (an optional SIG prefix is accepted).
	Signal string `yaml:"signal"`
	// Wait, when set, blocks until the signaled process exits or the timeout
	// elapses; a still-running process fails the step with a clear message.
	Wait *SignalWait `yaml:"wait,omitempty"`
}

// SignalWait bounds the wait for a signaled service's exit (#23).
type SignalWait struct {
	// Timeout is a Go duration (default "5s").
	Timeout string `yaml:"timeout,omitempty"`
}

// validSignalNames is the accepted `signal:` set (#23): the portable
// process-control signals. Anything more exotic (STOP/CONT/real-time) is out
// of scope for declarative CLI testing.
var validSignalNames = map[string]bool{
	"TERM": true, "INT": true, "HUP": true, "USR1": true, "USR2": true, "KILL": true,
}

// NormalizeSignalName upper-cases a signal name and strips an optional SIG
// prefix, so `term`, `TERM`, and `SIGTERM` all mean SIGTERM (#23).
func NormalizeSignalName(name string) string {
	return strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(name)), "SIG")
}

// ValidSignalName reports whether the (normalized) signal name is accepted.
func ValidSignalName(name string) bool { return validSignalNames[NormalizeSignalName(name)] }

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

// CDP drives a headless browser through a named browser runner. The
// action list runs in order against one browser session; the value captured by
// the last `text`/`eval` action feeds the `value` assertion target and
// `store from.value`.
type CDP struct {
	Runner  string      `yaml:"runner"`
	Actions []CDPAction `yaml:"actions"`
}

// CDPAction is one browser action; exactly one field is set. The action set is
// intentionally small and declarative — no conditions, loops, or expression
// language (#50).
type CDPAction struct {
	Navigate    string         `yaml:"navigate,omitempty"`     // load a URL
	WaitVisible string         `yaml:"wait_visible,omitempty"` // wait until a selector is visible
	WaitHidden  string         `yaml:"wait_hidden,omitempty"`  // wait until a selector is hidden/absent
	Click       string         `yaml:"click,omitempty"`        // click a selector
	Press       *CDPPress      `yaml:"press,omitempty"`        // press a key on a selector
	Select      *CDPSelect     `yaml:"select,omitempty"`       // choose an <option> in a <select>
	Check       string         `yaml:"check,omitempty"`        // tick a checkbox selector
	Uncheck     string         `yaml:"uncheck,omitempty"`      // untick a checkbox selector
	Screenshot  *CDPScreenshot `yaml:"screenshot,omitempty"`   // write a PNG into the workdir
	Text        string         `yaml:"text,omitempty"`         // capture a selector's text
	Title       bool           `yaml:"title,omitempty"`        // capture the page title
	Attribute   *CDPAttribute  `yaml:"attribute,omitempty"`    // capture an element attribute
	Eval        string         `yaml:"eval,omitempty"`         // evaluate JS, capture the result as JSON
	SendKeys    *CDPSendKeys   `yaml:"send_keys,omitempty"`    // type into a selector
	Upload      *CDPUpload     `yaml:"upload,omitempty"`       // set a file on an <input type=file>
	Download    *CDPDownload   `yaml:"download,omitempty"`     // click to trigger a download, capture the file
}

// CDPUpload sets File on the <input type=file> matched by Selector (#75). File is
// resolved against the scenario workdir and must exist there; the browser surface
// stays black-box (no scripted file dialogs).
type CDPUpload struct {
	Selector string `yaml:"selector"`
	File     string `yaml:"file"`
}

// CDPDownload triggers a download by clicking Click and captures the downloaded
// file into Dir (a workdir-relative directory, default the workdir root) using
// the server-suggested filename (#75). The captured value is the final filename,
// so existing file/dir/pdf/image assertions can validate the downloaded artifact.
type CDPDownload struct {
	Click string `yaml:"click"`
	Dir   string `yaml:"dir,omitempty"`
}

// CDPSendKeys types Value into the element matched by Selector.
type CDPSendKeys struct {
	Selector string `yaml:"selector"`
	Value    string `yaml:"value"`
}

// CDPPress presses a single key (e.g. "Enter", "Tab", or a printable character)
// on the element matched by Selector.
type CDPPress struct {
	Selector string `yaml:"selector"`
	Key      string `yaml:"key"`
}

// CDPSelect chooses the option whose value is Value in the <select> matched by
// Selector.
type CDPSelect struct {
	Selector string `yaml:"selector"`
	Value    string `yaml:"value"`
}

// CDPScreenshot writes a PNG of the page (or of Selector when set) to Path,
// resolved against the scenario workdir, so existing file/image assertions can
// inspect it.
type CDPScreenshot struct {
	Path     string `yaml:"path"`
	Selector string `yaml:"selector,omitempty"`
}

// CDPAttribute captures the Name attribute of the element matched by Selector
// into the value assertion path (like text/eval).
type CDPAttribute struct {
	Selector string `yaml:"selector"`
	Name     string `yaml:"name"`
}

// Fixture materializes an input file in the scenario workdir.
// Exactly one source is set: inline Content, inline Base64, or From (copy an
// existing file — e.g. committed binary testdata — resolved relative to the
// spec file's directory).
type Fixture struct {
	File    string `yaml:"file"`
	Content string `yaml:"content,omitempty"`
	Base64  string `yaml:"base64,omitempty"`
	From    string `yaml:"from,omitempty"`
	// Symlink, when set, makes File a symbolic link to this target instead of a
	// regular file (mutually exclusive with content/base64/from). The target is
	// written verbatim, so use ${workdir} for an absolute in-workdir target.
	Symlink string `yaml:"symlink,omitempty"`
	// Mode, when set, is an octal file mode (e.g. "0444") applied after writing.
	Mode string `yaml:"mode,omitempty"`
	// Mtime, when set, is an RFC3339 modification time applied after writing, so
	// specs can pin timestamps for content-vs-mtime change detection.
	Mtime string `yaml:"mtime,omitempty"`
}

// Run executes a command.
type Run struct {
	Command string `yaml:"command"`
	Runner  string `yaml:"runner,omitempty"`
	// Shell is a pointer so an authored `shell: false` is distinguishable from
	// "unset" when `defaults.run.shell` is layered in (an authored value always
	// wins over a default).
	Shell   *bool  `yaml:"shell,omitempty"`
	Cwd     string `yaml:"cwd,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
	// TimeoutSource names the level that supplied the effective Timeout
	// (run.timeout / runner.timeout / defaults.run.timeout / suite.timeout /
	// built-in default). It is set by the engine's precedence resolver (#17),
	// never authored, and lets a timeout-kill failure hint say which level to
	// adjust.
	TimeoutSource string            `yaml:"-"`
	Env           map[string]string `yaml:"env,omitempty"`
	// ClearEnv starts the child from an empty environment instead of inheriting
	// the host environment (#16), so host vars (LANG, GIT_*, proxies, ...) cannot
	// silently change the behavior under test. A pointer so an authored
	// `clear_env: false` is distinguishable from "unset" when
	// `defaults.run.clear_env` is layered in.
	ClearEnv *bool `yaml:"clear_env,omitempty"`
	// PassEnv copies the listed host variables into the cleared environment
	// (#16). Only meaningful with ClearEnv (the loader rejects it otherwise);
	// unset host variables are skipped, not an error.
	PassEnv []string `yaml:"pass_env,omitempty"`
	// SandboxHome points the child's home directory and per-OS config/cache/
	// data/state locations at a fresh `${workdir}/.atago-home` so a CLI that
	// reads or writes ~/.config, ~/.cache, or %APPDATA% runs hermetically. The
	// sandbox variables sit between pass_env and the step's own env in
	// precedence, and compose with clear_env (injected after the clear). A
	// pointer so an authored `sandbox_home: false` beats a defaulted true.
	SandboxHome *bool `yaml:"sandbox_home,omitempty"`
	// Stdin feeds the command's standard input: a scalar string (inline
	// text), {file: path} (workdir-relative, expanded and confined), or
	// {base64: data} for binary bytes (#18).
	Stdin Stdin `yaml:"stdin,omitempty"`
	// StdoutTo / StderrTo redirect the command's captured stdout / stderr to a
	// workdir-relative file (create/truncate), so a `shell: false` step can write
	// output to a file without borrowing the shell's `>` operator. The streams are
	// still captured internally, so stdout/stderr assertions on the same step keep
	// working. Paths follow the same workdir-confinement rule as assertion paths.
	StdoutTo string `yaml:"stdout_to,omitempty"`
	StderrTo string `yaml:"stderr_to,omitempty"`
	// Retry, when set, re-runs the command until the Until assertion passes,
	// polling declaratively for async behavior.
	Retry *Retry `yaml:"retry,omitempty"`
}

// Bool returns a pointer to v — sugar for authoring optional booleans (Shell)
// in Go literals.
func Bool(v bool) *bool { return &v }

// ShellEnabled reports whether the step opts into shell execution.
func (r *Run) ShellEnabled() bool { return r.Shell != nil && *r.Shell }

// ShellEnabled reports whether the service opts into shell execution.
func (s *Service) ShellEnabled() bool { return s.Shell != nil && *s.Shell }

// ClearEnvEnabled reports whether the step opts into a cleared environment (#16).
func (r *Run) ClearEnvEnabled() bool { return r.ClearEnv != nil && *r.ClearEnv }

// ClearEnvEnabled reports whether the service opts into a cleared environment (#16).
func (s *Service) ClearEnvEnabled() bool { return s.ClearEnv != nil && *s.ClearEnv }

// ClearEnvEnabled reports whether the pty step opts into a cleared environment (#16).
func (p *PTY) ClearEnvEnabled() bool { return p.ClearEnv != nil && *p.ClearEnv }

// SandboxHomeEnabled reports whether the run step opts into an isolated home (#71).
func (r *Run) SandboxHomeEnabled() bool { return r.SandboxHome != nil && *r.SandboxHome }

// SandboxHomeEnabled reports whether the pty step opts into an isolated home (#71).
func (p *PTY) SandboxHomeEnabled() bool { return p.SandboxHome != nil && *p.SandboxHome }

// Retry re-runs a command until Until passes or the attempt budget is exhausted.
// The last attempt's result is what subsequent steps observe.
type Retry struct {
	// Times is the maximum number of attempts (>= 1).
	Times int `yaml:"times"`
	// Interval is the wait between attempts as a Go duration (e.g. "200ms").
	// Empty means no wait.
	Interval string `yaml:"interval,omitempty"`
	// Until is a single assertion polled after each attempt; the loop stops as
	// soon as it passes. When it never passes, the run step fails.
	Until *Assert `yaml:"until"`
}

// PTY runs a command inside a pseudo-terminal (#8). The captured transcript
// (terminal echo included, ANSI intact) becomes the step's stdout, so every
// stream matcher, snapshot (with its ANSI normalization), and
// `store from.stdout` works unchanged.
type PTY struct {
	Command string `yaml:"command"`
	// Shell runs Command through the shell like run.shell.
	Shell *bool  `yaml:"shell,omitempty"`
	Cwd   string `yaml:"cwd,omitempty"`
	// Rows / Cols set the terminal size (default 24x80).
	Rows int `yaml:"rows,omitempty"`
	Cols int `yaml:"cols,omitempty"`
	// Timeout bounds the WHOLE session as a Go duration (default "30s"): a
	// prompt that never appears or a program that never exits fails loudly
	// instead of hanging the run.
	Timeout string            `yaml:"timeout,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	// ClearEnv starts the pty child from an empty environment instead of
	// inheriting the host environment (#16), mirroring run.clear_env.
	ClearEnv *bool `yaml:"clear_env,omitempty"`
	// PassEnv copies the listed host variables into the cleared environment
	// (#16). Only meaningful with ClearEnv; unset host variables are skipped.
	PassEnv []string `yaml:"pass_env,omitempty"`
	// SandboxHome isolates the pty child's home and per-OS config/cache/data/
	// state directories under `${workdir}/.atago-home`, mirroring run.sandbox_home.
	SandboxHome *bool `yaml:"sandbox_home,omitempty"`
	// Session is the ordered expect/send script. Each entry sets exactly one
	// of Expect (wait until the accumulated transcript matches the regexp) or
	// Send (write the string to the terminal; an empty send transmits EOF,
	// i.e. ^D). Deliberately no branching — atago is not a scripting language.
	Session []PTYAction `yaml:"session,omitempty"`
}

// PTYAction is one expect-or-send entry in a pty session (#8).
type PTYAction struct {
	// Expect waits until the transcript matches this regexp. A never-matching
	// expect fails the step (reported like an assertion) when the session
	// timeout elapses.
	Expect string `yaml:"expect,omitempty"`
	// Send writes to the terminal: a scalar string verbatim (the empty string
	// sends EOF/^D; ${name} expansion applies) or {key: <name>} for a named
	// key (#26) — enter, tab, esc, arrows, f1-f12, ctrl-a..ctrl-z — so
	// sessions stay readable instead of embedding \x1b escapes.
	Send *PTYSend `yaml:"send,omitempty"`
}

// PTYSend is the polymorphic pty send payload (#26): exactly one of Text
// (scalar form) or Key (mapping form) is set.
type PTYSend struct {
	// Text is sent verbatim; the empty string transmits EOF (^D).
	Text *string
	// Key is a named key, normalized to lower case.
	Key string
}

// SendText is sugar for authoring the scalar form in Go literals (tests).
func SendText(s string) *PTYSend { return &PTYSend{Text: &s} }

// UnmarshalYAML decodes send as a scalar string or a {key: name} mapping,
// rejecting unknown mapping keys (a custom unmarshaler bypasses the loader's
// strict decode).
func (p *PTYSend) UnmarshalYAML(unmarshal func(any) error) error {
	var one string
	if err := unmarshal(&one); err == nil {
		p.Text = &one
		return nil
	}
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return fmt.Errorf("send must be a string or {key: <name>} (e.g. {key: enter})")
	}
	for k, v := range raw {
		if k != "key" {
			return fmt.Errorf("send: unknown key %q (accepted: key)", k)
		}
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("send.key must be a string")
		}
		p.Key = strings.ToLower(strings.TrimSpace(str))
	}
	if p.Key == "" {
		return fmt.Errorf("send: {key: <name>} requires a key name (e.g. enter, tab, ctrl-c)")
	}
	return nil
}

// ptyKeySequences maps each named key (#26) to the xterm byte sequence it
// transmits. Documented bytes: enter=\r, tab=\t, esc=\x1b, space=" ",
// backspace=\x7f (DEL, the modern erase), delete=\x1b[3~, arrows
// up/down/right/left=\x1b[A/B/C/D, home=\x1b[H, end=\x1b[F,
// pageup=\x1b[5~, pagedown=\x1b[6~, f1-f4=\x1bOP..\x1bOS,
// f5..f12=\x1b[15~,[17~..[21~,[23~,[24~, ctrl-a..ctrl-z=0x01..0x1a
// (ctrl-d is therefore the readable alias for the empty-send EOF rule).
var ptyKeySequences = func() map[string]string {
	m := map[string]string{
		"enter":     "\r",
		"tab":       "\t",
		"esc":       "\x1b",
		"space":     " ",
		"backspace": "\x7f",
		"delete":    "\x1b[3~",
		"up":        "\x1b[A",
		"down":      "\x1b[B",
		"right":     "\x1b[C",
		"left":      "\x1b[D",
		"home":      "\x1b[H",
		"end":       "\x1b[F",
		"pageup":    "\x1b[5~",
		"pagedown":  "\x1b[6~",
		"f1":        "\x1bOP",
		"f2":        "\x1bOQ",
		"f3":        "\x1bOR",
		"f4":        "\x1bOS",
		"f5":        "\x1b[15~",
		"f6":        "\x1b[17~",
		"f7":        "\x1b[18~",
		"f8":        "\x1b[19~",
		"f9":        "\x1b[20~",
		"f10":       "\x1b[21~",
		"f11":       "\x1b[23~",
		"f12":       "\x1b[24~",
	}
	for c := byte('a'); c <= 'z'; c++ {
		m["ctrl-"+string(c)] = string([]byte{c - 'a' + 1})
	}
	return m
}()

// ValidPTYKey reports whether name is in the named-key vocabulary (#26).
func ValidPTYKey(name string) bool {
	_, ok := ptyKeySequences[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

// ptyKeyBySequence reverse-maps an xterm byte sequence to its friendly key name
// (#69), preferring the readable name over a ctrl-* alias when a byte is shared
// (e.g. \r is both enter and ctrl-m — enter wins). Built once at init: the
// ctrl-* aliases go in first, then the friendly names overwrite any collision.
var ptyKeyBySequence = func() map[string]string {
	m := make(map[string]string, len(ptyKeySequences))
	for c := byte('a'); c <= 'z'; c++ {
		m[string([]byte{c - 'a' + 1})] = "ctrl-" + string(c)
	}
	for _, name := range []string{
		"enter", "tab", "esc", "space", "backspace", "delete",
		"up", "down", "right", "left", "home", "end",
		"pageup", "pagedown",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
	} {
		m[ptyKeySequences[name]] = name
	}
	return m
}()

// PTYKeyForSequence returns the friendly named key whose xterm sequence exactly
// equals seq (#69), so `atago record --pty` can render a lone control key as
// {key: <name>} instead of an opaque escape. It reports false when no named key
// matches the bytes exactly.
func PTYKeyForSequence(seq string) (string, bool) {
	name, ok := ptyKeyBySequence[seq]
	return name, ok
}

// PTYKeyNames lists the vocabulary for error messages, compactly.
func PTYKeyNames() string {
	return "enter, tab, esc, space, backspace, delete, up, down, left, right, home, end, pageup, pagedown, f1-f12, ctrl-a..ctrl-z"
}

// Bytes resolves the send payload to the bytes written to the terminal: the
// named key's xterm sequence, the verbatim text, or 0x04 (VEOF, ^D) for the
// historical empty-string EOF rule.
func (p *PTYSend) Bytes() []byte {
	if p.Key != "" {
		return []byte(ptyKeySequences[p.Key])
	}
	if p.Text != nil && *p.Text == "" {
		return []byte{0x04}
	}
	if p.Text != nil {
		return []byte(*p.Text)
	}
	return nil
}

// Label renders the send symbolically for explain/doc (#26): "press Enter"
// for keys, a quoted excerpt for text, "EOF (^D)" for the empty string.
func (p *PTYSend) Label() string {
	switch {
	case p.Key != "":
		return "press " + p.Key
	case p.Text != nil && *p.Text == "":
		return "send EOF (^D)"
	case p.Text != nil:
		return fmt.Sprintf("type %q", *p.Text)
	default:
		return "send"
	}
}

// HTTP issues an HTTP request (post-MVP).
type HTTP struct {
	Runner string            `yaml:"runner,omitempty"`
	Method string            `yaml:"method"`
	Path   string            `yaml:"path"`
	Header map[string]string `yaml:"header,omitempty"`
	JSON   any               `yaml:"json,omitempty"`
	// Body sends a raw string payload verbatim (default Content-Type
	// text/plain, overridable via Header) — for text-first APIs such as metrics
	// exposition or message publishing. The payload fields (JSON, Body,
	// BodyFile, Form/Files) are mutually exclusive: a request has one payload.
	Body string `yaml:"body,omitempty"`
	// BodyFile streams a workdir-relative file as the raw request body —
	// binary-safe, for PUT/POST upload endpoints that take the file content
	// directly (file-sharing servers, artifact stores).
	BodyFile string `yaml:"body_file,omitempty"`
	// Form sends form fields: application/x-www-form-urlencoded on their own,
	// or multipart/form-data parts when Files is also set.
	Form map[string]string `yaml:"form,omitempty"`
	// Files attaches workdir-relative files as multipart/form-data parts; any
	// Form fields ride along as regular parts. This is the browser-style
	// file-upload request most self-hosted web apps expect.
	Files []FilePart `yaml:"files,omitempty"`
	// BodyTo writes the response body to this workdir-relative file
	// (create/truncate), so a downloaded artifact can be inspected with the
	// file/image/pdf assertion targets — the http analogue of run's stdout_to.
	BodyTo string `yaml:"body_to,omitempty"`
	// FollowRedirects controls whether 3xx responses are followed. It defaults
	// to true (matching every HTTP client a user knows); set false to assert on
	// the redirect itself — its status code and Location header.
	FollowRedirects *bool `yaml:"follow_redirects,omitempty"`
	// Retry, when set, re-issues the request until the Until assertion passes,
	// polling declaratively for eventually-consistent endpoints (a metric that
	// appears after a scrape, an async job flipping to done) exactly like a run
	// step's retry.
	Retry *Retry `yaml:"retry,omitempty"`
}

// FilePart is one file attached to a multipart/form-data request.
type FilePart struct {
	// Field is the multipart form field name the server reads the file from.
	Field string `yaml:"field"`
	// Path is the workdir-relative file whose content becomes the part body.
	Path string `yaml:"path"`
	// ContentType overrides the part's Content-Type (default: detected from
	// the file content, falling back to application/octet-stream).
	ContentType string `yaml:"content_type,omitempty"`
}

// Assert checks externally observable behavior. Exactly one target family is set.
type Assert struct {
	ExitCode *ExitCode     `yaml:"exit_code,omitempty"`
	Stdout   *StreamAssert `yaml:"stdout,omitempty"`
	Stderr   *StreamAssert `yaml:"stderr,omitempty"`
	File     *FileAssert   `yaml:"file,omitempty"`

	// HTTP assertion targets.
	Status *int          `yaml:"status,omitempty"`
	Header *HeaderMatch  `yaml:"header,omitempty"`
	Body   *StreamAssert `yaml:"body,omitempty"`

	// Rows is the db assertion target: the query result rows as a JSON array,
	// matched with the stream matchers (json path/length, contains, …).
	Rows *StreamAssert `yaml:"rows,omitempty"`

	// gRPC assertion targets: GRPCStatus checks the numeric status
	// code; Message matches the response message (as JSON) with the stream matchers.
	GRPCStatus *int          `yaml:"grpc_status,omitempty"`
	Message    *StreamAssert `yaml:"message,omitempty"`

	// Value is the browser assertion target: the value captured by the
	// last text/eval action, matched with the stream matchers.
	Value *StreamAssert `yaml:"value,omitempty"`

	// Image is the image assertion target: it inspects a generated
	// image file's decoded properties (format, dimensions, alpha) and can compare
	// its pixels against a baseline image.
	Image *ImageAssert `yaml:"image,omitempty"`

	// Dir is the directory/tree assertion target (#74): black-box checks over a
	// generated directory — existence, expected/forbidden children, entry counts,
	// and glob coverage — for multi-file generators (static sites, scaffolds,
	// extracted archives). It is deliberately declarative and non-recursive.
	Dir *DirAssert `yaml:"dir,omitempty"`

	// PDF is the PDF assertion target (#73): a small, black-box, content-oriented
	// surface for generated PDFs — page count, Info metadata fields, and extracted
	// text — without depending on a specific layout engine.
	PDF *PDFAssert `yaml:"pdf,omitempty"`

	// Mock is the mock-server assertion target (#24): what the CLI under test
	// actually sent to a declared mock server — request count, and header/body
	// matchers applied to the last matching recorded request.
	Mock *MockAssert `yaml:"mock,omitempty"`

	// Screen is the rendered-terminal assertion target (#27), valid after a
	// pty step: the transcript replayed through a vt10x emulator sized by the
	// step's rows/cols, asserted as plain text with the stream matchers
	// (line.n addresses screen rows 1-based). The raw transcript stays on
	// stdout.
	Screen *StreamAssert `yaml:"screen,omitempty"`

	// Duration is the wall-clock assertion target (#31), valid after a
	// measurable step (run/http/query/grpc/pty): it bounds how long that step
	// took with lt/lte/gt/gte Go-duration bounds.
	Duration *DurationAssert `yaml:"duration,omitempty"`

	// Changes is the workdir-delta assertion target (#70), valid after an
	// immediately preceding run/pty step: it pins exactly which files that step
	// created, modified, and deleted in the scenario workdir.
	Changes *ChangesAssert `yaml:"changes,omitempty"`
}

// ChangesAssert pins the exact delta the immediately preceding run/pty step
// made to the scenario workdir (#70). The delta is content-based (hash, not
// mtime). Each set field is EXHAUSTIVE in both directions: every observed path
// must be matched by an entry (an exact workdir-relative path or a /-glob) and
// every entry must match at least one observed path — so `modified: []` asserts
// "modified nothing". An omitted (nil) field is unconstrained.
type ChangesAssert struct {
	Created  *StringList `yaml:"created,omitempty"`
	Modified *StringList `yaml:"modified,omitempty"`
	Deleted  *StringList `yaml:"deleted,omitempty"`
}

// DurationAssert bounds a step's measured wall-clock time (#31). At least one
// bound must be set; lt/lte are mutually exclusive, as are gt/gte, and any
// pair must form a non-empty interval (validated at load time). Values are Go
// duration strings ("2s", "100ms").
type DurationAssert struct {
	// LT / LTE are the upper bound (exclusive / inclusive).
	LT  string `yaml:"lt,omitempty"`
	LTE string `yaml:"lte,omitempty"`
	// GT / GTE are the lower bound (exclusive / inclusive).
	GT  string `yaml:"gt,omitempty"`
	GTE string `yaml:"gte,omitempty"`
}

// DescribeDuration renders a duration assert's bounds as a human phrase (#31),
// shared by explain and doc so the two never drift.
func (d *DurationAssert) DescribeDuration() string {
	var parts []string
	if d.LT != "" {
		parts = append(parts, "in under "+d.LT)
	}
	if d.LTE != "" {
		parts = append(parts, "in at most "+d.LTE)
	}
	if d.GT != "" {
		parts = append(parts, "in over "+d.GT)
	}
	if d.GTE != "" {
		parts = append(parts, "in at least "+d.GTE)
	}
	return strings.Join(parts, " and ")
}

// PDFAssert checks a generated PDF file (#73). Like ImageAssert/DirAssert, every
// set field is an independent constraint and all must hold; at least one (besides
// Path) must be set. The surface is intentionally small and content-oriented:
// page count, Info dictionary metadata, and extracted text — not layout.
type PDFAssert struct {
	// Path is the PDF under test, resolved against the scenario workdir when
	// relative (like FileAssert.Path).
	Path string `yaml:"path"`
	// Pages asserts the exact page count; MinPages/MaxPages assert bounds.
	Pages    *int `yaml:"pages,omitempty"`
	MinPages *int `yaml:"min_pages,omitempty"`
	MaxPages *int `yaml:"max_pages,omitempty"`
	// Metadata maps an Info-dictionary field (title, author, subject, keywords,
	// creator, producer) to a substring the field's value must contain. Keys are
	// matched case-insensitively.
	Metadata map[string]string `yaml:"metadata,omitempty"`
	// Text applies the standard stream matchers (contains/matches/equals/…) to the
	// text extracted from the PDF's content streams (raw and FlateDecode-decoded).
	Text *StreamAssert `yaml:"text,omitempty"`
}

// DirAssert checks a generated directory tree (#74). Like ImageAssert, every
// field that is set is a separate constraint and all of them must hold, because
// a directory has several independent observable properties. At least one
// constraint (besides Path) must be set. Child paths are confined to the
// directory and may not escape it.
type DirAssert struct {
	// Path is the directory under test, resolved against the scenario workdir when
	// relative (like FileAssert.Path).
	Path string `yaml:"path"`
	// Exists asserts the path exists and is a directory (exists:false asserts it
	// is absent).
	Exists *bool `yaml:"exists,omitempty"`
	// Contains lists child paths (relative to Path) that must exist. A child may
	// name a nested path (e.g. "assets/app.css"); it must stay within Path.
	Contains []string `yaml:"contains,omitempty"`
	// NotContains lists child paths (relative to Path) that must NOT exist.
	NotContains []string `yaml:"not_contains,omitempty"`
	// Count asserts the exact number of direct entries in the directory.
	Count *int `yaml:"count,omitempty"`
	// MinCount / MaxCount assert bounds on the number of direct entries.
	MinCount *int `yaml:"min_count,omitempty"`
	MaxCount *int `yaml:"max_count,omitempty"`
	// Glob asserts that at least one direct entry matches this shell glob pattern
	// (filepath.Match semantics, e.g. "*.html").
	Glob string `yaml:"glob,omitempty"`
	// Recursive makes Contains/NotContains accept slash-separated relative
	// paths anywhere in the tree, and Count/MinCount/MaxCount/Glob apply to
	// the whole walk (counts see FILES only; Glob matches each entry's
	// relative path, or its basename for patterns without "/") (#25).
	Recursive bool `yaml:"recursive,omitempty"`
	// Snapshot compares the whole tree against a golden manifest (#25):
	// sorted /-separated relative paths, one line per entry — `dir <path>`,
	// `file <path> sha256:<hash>` (hashed byte-exact: CRLF is NOT normalized
	// inside file content), or `link <path> -> <target>` (not traversed).
	// No mode/mtime (not portable). Refresh with --update-snapshots.
	Snapshot string `yaml:"snapshot,omitempty"`
	// Ignore lists glob patterns excluded from the recursive walk and the
	// snapshot manifest ("*.log", ".git/**"). A pattern without "/" also
	// matches basenames at any depth; a "<dir>/**" pattern prunes the whole
	// subtree.
	Ignore []string `yaml:"ignore,omitempty"`
}

// ImageAssert checks a generated image file. Unlike the one-of
// stream/file targets, every field that is set is a separate constraint and all
// of them must hold, because an image has several independent observable
// attributes. At least one constraint must be set.
type ImageAssert struct {
	// Path is the image file under test, resolved against the scenario workdir
	// when relative (like FileAssert.Path).
	Path string `yaml:"path"`
	// Format asserts the encoded image format, detected from the file's content:
	// png, jpeg, gif, webp, bmp, tiff, avif, or svg.
	Format string `yaml:"format,omitempty"`
	// Width / Height assert the exact pixel dimensions.
	Width  *int `yaml:"width,omitempty"`
	Height *int `yaml:"height,omitempty"`
	// MinWidth / MaxWidth / MinHeight / MaxHeight assert dimension bounds.
	MinWidth  *int `yaml:"min_width,omitempty"`
	MaxWidth  *int `yaml:"max_width,omitempty"`
	MinHeight *int `yaml:"min_height,omitempty"`
	MaxHeight *int `yaml:"max_height,omitempty"`
	// Alpha asserts whether the image actually carries transparency (any
	// non-opaque pixel). It scans decoded pixels rather than the in-memory color
	// model, so an opaque truecolor PNG/BMP correctly reports alpha=false.
	Alpha *bool `yaml:"alpha,omitempty"`
	// SimilarTo compares the decoded pixels against a baseline image. A relative
	// path resolves against the spec file's directory (like a committed
	// snapshot); use an absolute or ${workdir}-prefixed path to compare against
	// another generated file. Both images must share dimensions.
	SimilarTo string `yaml:"similar_to,omitempty"`
	// MaxDiff is the maximum allowed normalized mean per-pixel difference (0..1)
	// for SimilarTo. It defaults to 0 (an exact pixel match); lossy formats need a
	// small tolerance such as 0.02.
	MaxDiff *float64 `yaml:"max_diff,omitempty"`
}

// Stdin is a run step's standard-input source (#18). It accepts either the
// historical scalar string (inline text) or a mapping with exactly one of
// `file` (a workdir-relative, ${name}-expanded, workdir-confined path whose
// bytes are fed to the child) or `base64` (binary bytes, validated at load
// time; no ${name} expansion, mirroring fixture.base64). The loader enforces
// the one-of rule.
type Stdin struct {
	// Inline is the scalar form, fed to stdin verbatim.
	Inline string
	// File names a workdir-relative file whose bytes become stdin.
	File string
	// Base64 carries binary stdin as base64.
	Base64 string

	// mapped records that the author used the mapping form, so the validator
	// can reject an empty mapping ({}), which is otherwise indistinguishable
	// from "no stdin".
	mapped bool
}

// IsZero reports whether no stdin was authored at all.
func (s Stdin) IsZero() bool {
	return s.Inline == "" && s.File == "" && s.Base64 == "" && !s.mapped
}

// IsMapping reports whether the author used the {file/base64} mapping form.
func (s Stdin) IsMapping() bool { return s.mapped }

// UnmarshalYAML decodes stdin as a scalar string or a {file}/{base64} mapping.
// It uses the interface-based decoder so escapes like "\x1b" in the inline
// form are resolved by goccy's parser, matching the historical behavior.
// Unknown mapping keys are rejected here (a custom unmarshaler bypasses the
// loader's strict-decode), with the accepted shapes spelled out.
func (s *Stdin) UnmarshalYAML(unmarshal func(any) error) error {
	var one string
	if err := unmarshal(&one); err == nil {
		s.Inline = one
		return nil
	}
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return fmt.Errorf("stdin must be a string, {file: path}, or {base64: data}")
	}
	s.mapped = true
	for k, v := range raw {
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("stdin.%s must be a string", k)
		}
		switch k {
		case "file":
			s.File = str
		case "base64":
			s.Base64 = str
		default:
			return fmt.Errorf("stdin: unknown key %q (accepted: file, base64)", k)
		}
	}
	return nil
}

// ExitCode accepts a bare integer, {not: <int>}, or {in: [<int>, ...]} (#19).
// The `in` set is the contract shape real CLIs document (grep's 0/1,
// terraform plan -detailed-exitcode's 0/2): membership in an accepted set.
type ExitCode struct {
	Equals *int
	Not    *int
	In     []int
}

// UnmarshalYAML decodes exit_code as a scalar int, a {not: int} map, or an
// {in: [int, ...]} map. Anything else gets a purpose-built error: the generic
// decoder message ("string was used where mapping is expected", positioned at
// the sub-node) reads like an internal failure, and a spec author needs to
// know the accepted shapes.
func (e *ExitCode) UnmarshalYAML(b []byte) error {
	if n, err := strconv.Atoi(trimYAMLScalar(string(b))); err == nil {
		e.Equals = &n
		return nil
	}
	var m struct {
		Not *int  `yaml:"not"`
		In  []int `yaml:"in"`
	}
	if err := yaml.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("exit_code must be an integer (exit_code: 0), a negation (exit_code: {not: 0}), or a set (exit_code: {in: [0, 2]}), got %q", strings.TrimSpace(string(b)))
	}
	e.Not = m.Not
	e.In = m.In
	return nil
}

// StreamAssert matches a captured text stream (stdout/stderr/body). One matcher.
//
// Line is an optional 1-based selector: when set, the matcher is
// applied to that single line of the stream instead of the whole stream. It is
// not itself a matcher, so exactly one of empty/contains/matches/equals must
// still be set. Line does not compose with json/snapshot (those operate on the
// whole document).
type StreamAssert struct {
	Line        *int        `yaml:"line,omitempty"`
	Empty       *bool       `yaml:"empty,omitempty"`
	Contains    StringList  `yaml:"contains,omitempty"`
	NotContains StringList  `yaml:"not_contains,omitempty"`
	Matches     *string     `yaml:"matches,omitempty"`
	NotMatches  *string     `yaml:"not_matches,omitempty"`
	Equals      *string     `yaml:"equals,omitempty"`
	NotEquals   *string     `yaml:"not_equals,omitempty"`
	JSON        *JSONAssert `yaml:"json,omitempty"`
	YAML        *JSONAssert `yaml:"yaml,omitempty"`
	Snapshot    string      `yaml:"snapshot,omitempty"`
}

// StringList is a matcher argument that accepts either a single YAML scalar
// string or a sequence of strings. It backs the `contains` / `not_contains`
// matchers on stream and file assertions so one matcher can require (or forbid)
// several substrings without repeating the assert block. A scalar decodes to a
// one-element list and keeps byte-identical behavior with the pre-list format;
// a sequence decodes to its elements. `contains` requires every element to be
// present, `not_contains` requires every element to be absent, and either way
// the whole list counts as a single matcher (the one-of matcher rule is
// unchanged).
type StringList []string

// UnmarshalYAML accepts a scalar string or a sequence of strings. It uses the
// interface-based decoder (not the raw-bytes form) so escapes like "\x1b" are
// resolved by goccy's parser once, rather than re-tokenized from node bytes.
func (l *StringList) UnmarshalYAML(unmarshal func(any) error) error {
	var one string
	if err := unmarshal(&one); err == nil {
		*l = StringList{one}
		return nil
	}
	var many []string
	if err := unmarshal(&many); err != nil {
		return err
	}
	*l = StringList(many)
	return nil
}

// FileAssert checks a generated file.
type FileAssert struct {
	Path        string      `yaml:"path"`
	Exists      *bool       `yaml:"exists,omitempty"`
	Contains    StringList  `yaml:"contains,omitempty"`
	NotContains StringList  `yaml:"not_contains,omitempty"`
	Executable  *bool       `yaml:"executable,omitempty"`
	JSON        *JSONAssert `yaml:"json,omitempty"`
	Snapshot    string      `yaml:"snapshot,omitempty"`
}

// JSONAssert matches a value selected by a JSONPath. One matcher.
//
// Gt/Gte/Lt/Lte assert a numeric bound on the selected value (which must be a
// number, or a numeric string). They exist because tools routinely emit
// non-deterministic-but-bounded metrics — a processed-record count, a coverage
// figure, a duration — where an exact `equals` is impossible but "at least N"
// or "below N" is exactly the contract worth pinning (surfaced dogfooding runn's
// coverage/loadt metrics).
type JSONAssert struct {
	Path    string   `yaml:"path"`
	Equals  any      `yaml:"equals,omitempty"`
	Matches *string  `yaml:"matches,omitempty"`
	Length  *int     `yaml:"length,omitempty"`
	Gt      *float64 `yaml:"gt,omitempty"`
	Gte     *float64 `yaml:"gte,omitempty"`
	Lt      *float64 `yaml:"lt,omitempty"`
	Lte     *float64 `yaml:"lte,omitempty"`
}

// HeaderMatch checks an HTTP header (response headers on the `header` target,
// recorded request headers on the `mock` target). Exactly one matcher.
type HeaderMatch struct {
	Name     string  `yaml:"name"`
	Contains *string `yaml:"contains,omitempty"`
	Equals   *string `yaml:"equals,omitempty"`
	// Matches applies a regexp — the natural shape for auth headers
	// ("^Bearer ") (#24).
	Matches *string `yaml:"matches,omitempty"`
}

// Store captures a value into the variable store (post-MVP).
type Store struct {
	Name string     `yaml:"name"`
	From *StoreFrom `yaml:"from"`
}

// StoreFrom selects where a stored value comes from. Exactly one source is set.
// Stdout/Body extract via a json/regex selector; File reads a generated file via
// a json selector; Header captures an HTTP response header value by name.
type StoreFrom struct {
	Stdout  *StreamAssert `yaml:"stdout,omitempty"`
	Body    *StreamAssert `yaml:"body,omitempty"`
	File    *FileAssert   `yaml:"file,omitempty"`
	Header  string        `yaml:"header,omitempty"`
	Rows    *StreamAssert `yaml:"rows,omitempty"`
	Message *StreamAssert `yaml:"message,omitempty"`
	Value   *StreamAssert `yaml:"value,omitempty"`
}

// trimYAMLScalar strips surrounding whitespace/newlines from a raw scalar node.
func trimYAMLScalar(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
