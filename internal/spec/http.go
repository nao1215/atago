package spec

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
	// file/image/pdf assertion targets — the http analog of run's stdout_to.
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
