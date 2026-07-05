package spec

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

// ShellEnabled reports whether the service opts into shell execution.
func (s *Service) ShellEnabled() bool { return s.Shell != nil && *s.Shell }

// ClearEnvEnabled reports whether the service opts into a cleared environment (#16).
func (s *Service) ClearEnvEnabled() bool { return s.ClearEnv != nil && *s.ClearEnv }
