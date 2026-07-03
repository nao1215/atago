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
	// element at load time (ADR-0039). It is authoring sugar only: the loader
	// expands it into the concrete scenario/step/service model before validation,
	// so nothing downstream (engine, manifest, explain) ever observes `defaults`.
	Defaults  *Defaults  `yaml:"defaults,omitempty"`
	Scenarios []Scenario `yaml:"scenarios"`
}

// Defaults holds the top-level `defaults:` block (ADR-0039). Each fragment is
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
	// driver is inferred from the scheme unless Driver is set (ADR-0026).
	DSN string `yaml:"dsn,omitempty"`
	// Driver, when set, names the database/sql driver explicitly (sqlite,
	// postgres, or mysql), overriding scheme inference.
	Driver string `yaml:"driver,omitempty"`

	// SSH runner fields (ADR-0027). A `run` step naming an ssh runner executes its
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

	// gRPC runner fields (ADR-0028). A `grpc` step calls a unary method on the
	// target, resolving the schema via server reflection (no compiled stubs).
	Target string `yaml:"target,omitempty"` // host:port of the gRPC server
	TLS    bool   `yaml:"tls,omitempty"`    // use TLS (default plaintext)

	// Browser runner fields (ADR-0038). A minimal, black-box configuration surface
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
	// one concrete scenario per row before validation (ADR-0020).
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
	// and terminated when the scenario ends (ADR-0031). They let a spec exercise a
	// CLI that talks to a peer (a TCP client, an API consumer) by standing up that
	// peer for the duration of the scenario.
	Services []Service `yaml:"services,omitempty"`
	Steps    []Step    `yaml:"steps"`
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
// terminated (with its whole process group) when the scenario ends (ADR-0031).
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
	// Ready declares how to wait until the service is accepting work before the
	// steps run. When omitted, the steps start as soon as the process is spawned.
	Ready *Ready `yaml:"ready,omitempty"`
}

// Ready is a service readiness probe (ADR-0031). Exactly one of File/Port/Log/
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
// only when X succeeds (ADR-0021). The probe runs through the shell.
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
}

// Query runs a SQL statement through a named db runner (ADR-0026). The result
// rows (for a SELECT) are captured as JSON for the `rows` assertion target and
// `store from.rows`; a non-row statement records its affected-row count.
type Query struct {
	Runner string `yaml:"runner"`
	SQL    string `yaml:"sql"`
}

// GRPC calls a unary gRPC method through a named grpc runner (ADR-0028). The
// response message is captured as JSON for the `message` assertion target and
// `store from.message`; the status code feeds the `grpc_status` target.
type GRPC struct {
	Runner string            `yaml:"runner"`
	Method string            `yaml:"method"` // "pkg.Service/Method"
	Header map[string]string `yaml:"header,omitempty"`
	JSON   any               `yaml:"json,omitempty"` // request message
}

// CDP drives a headless browser through a named browser runner (ADR-0029). The
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
	Shell   *bool             `yaml:"shell,omitempty"`
	Cwd     string            `yaml:"cwd,omitempty"`
	Timeout string            `yaml:"timeout,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	Stdin   string            `yaml:"stdin,omitempty"`
	// StdoutTo / StderrTo redirect the command's captured stdout / stderr to a
	// workdir-relative file (create/truncate), so a `shell: false` step can write
	// output to a file without borrowing the shell's `>` operator. The streams are
	// still captured internally, so stdout/stderr assertions on the same step keep
	// working. Paths follow the same workdir-confinement rule as assertion paths.
	StdoutTo string `yaml:"stdout_to,omitempty"`
	StderrTo string `yaml:"stderr_to,omitempty"`
	// Retry, when set, re-runs the command until the Until assertion passes,
	// polling declaratively for async behavior (ADR-0022).
	Retry *Retry `yaml:"retry,omitempty"`
}

// Bool returns a pointer to v — sugar for authoring optional booleans (Shell)
// in Go literals.
func Bool(v bool) *bool { return &v }

// ShellEnabled reports whether the step opts into shell execution.
func (r *Run) ShellEnabled() bool { return r.Shell != nil && *r.Shell }

// ShellEnabled reports whether the service opts into shell execution.
func (s *Service) ShellEnabled() bool { return s.Shell != nil && *s.Shell }

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
	// Send writes this string to the terminal verbatim (include "\n" to press
	// enter). The empty string sends EOF (^D). ${name} expansion applies.
	Send *string `yaml:"send,omitempty"`
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
	// step's retry (ADR-0022).
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
	// matched with the stream matchers (json path/length, contains, …) (ADR-0026).
	Rows *StreamAssert `yaml:"rows,omitempty"`

	// gRPC assertion targets (ADR-0028): GRPCStatus checks the numeric status
	// code; Message matches the response message (as JSON) with the stream matchers.
	GRPCStatus *int          `yaml:"grpc_status,omitempty"`
	Message    *StreamAssert `yaml:"message,omitempty"`

	// Value is the browser assertion target (ADR-0029): the value captured by the
	// last text/eval action, matched with the stream matchers.
	Value *StreamAssert `yaml:"value,omitempty"`

	// Image is the image assertion target (ADR-0030): it inspects a generated
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
}

// ImageAssert checks a generated image file (ADR-0030). Unlike the one-of
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

// ExitCode accepts either a bare integer or {not: <int>}.
type ExitCode struct {
	Equals *int
	Not    *int
}

// UnmarshalYAML decodes exit_code as a scalar int or a {not: int} map. Anything
// else gets a purpose-built error: the generic decoder message ("string was
// used where mapping is expected", positioned at the sub-node) reads like an
// internal failure, and a spec author needs to know the two accepted shapes.
func (e *ExitCode) UnmarshalYAML(b []byte) error {
	if n, err := strconv.Atoi(trimYAMLScalar(string(b))); err == nil {
		e.Equals = &n
		return nil
	}
	var m struct {
		Not *int `yaml:"not"`
	}
	if err := yaml.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("exit_code must be an integer (exit_code: 0) or a negation (exit_code: {not: 0}), got %q", strings.TrimSpace(string(b)))
	}
	e.Not = m.Not
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

// HeaderMatch checks an HTTP response header (post-MVP).
type HeaderMatch struct {
	Name     string  `yaml:"name"`
	Contains *string `yaml:"contains,omitempty"`
	Equals   *string `yaml:"equals,omitempty"`
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
