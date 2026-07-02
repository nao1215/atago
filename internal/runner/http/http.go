// Package http implements the HTTP runner: it issues a request described by an
// `http:` step and captures the response (status, headers, body) as a
// runner.Result. It is the atago counterpart to runn's HTTP
// runner, kept declarative — there is no expression language, only the `${name}`
// substitution the engine applies before a step reaches this package.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// PolicyError reports that a request targeted a host the spec's
// `permissions.network.allow` list does not permit. The engine
// maps it to exit code 6 (security policy violation).
type PolicyError struct {
	Host  string
	Allow []string
}

func (e *PolicyError) Error() string {
	return fmt.Sprintf("network policy denies host %q (allowed: %s)", e.Host, strings.Join(e.Allow, ", "))
}

// Config is the resolved configuration for an HTTP runner, derived from a named
// `runners:` entry and the spec's network policy.
type Config struct {
	// BaseURL is prepended to a relative request path; an absolute request path
	// (starting with http:// or https://) ignores it.
	BaseURL string
	// Timeout bounds the whole request/response exchange; zero means no timeout.
	Timeout time.Duration
	// Allow lists permitted hostnames (optionally host:port). Empty means no
	// network restriction is enforced at run time.
	Allow []string
	// Workdir is the scenario's isolated directory; body_file/files payloads
	// are read from it and body_to responses are written into it, all
	// workdir-confined like fixture and assertion paths.
	Workdir string
}

// Runner issues HTTP requests for `http:` steps.
type Runner struct {
	client  *http.Client
	baseURL string
	allow   []string
	workdir string
}

// New returns an HTTP runner for the given configuration.
func New(cfg Config) *Runner {
	return &Runner{
		client:  &http.Client{Timeout: cfg.Timeout},
		baseURL: cfg.BaseURL,
		allow:   cfg.Allow,
		workdir: cfg.Workdir,
	}
}

// Do builds and executes the request described by h and returns the captured
// response. A non-nil error means the request could not be made or completed; a
// response with any status code (including 4xx/5xx) is a successful Do with the
// status recorded on the Result. A *PolicyError means the host was denied.
func (r *Runner) Do(ctx context.Context, h *spec.HTTP) (*runner.Result, error) {
	target, err := r.resolveURL(h.Path)
	if err != nil {
		return nil, err
	}
	if err := r.checkPolicy(target); err != nil {
		return nil, err
	}

	body, contentType, err := r.encodeBody(h)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(strings.TrimSpace(h.Method))
	req, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return nil, fmt.Errorf("building %s %s: %w", method, target, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range h.Header { // explicit headers win over the default content-type
		req.Header.Set(k, v)
	}

	// follow_redirects defaults to true, matching every HTTP client a user
	// knows. `follow_redirects: false` surfaces the 3xx itself so a spec can
	// assert the redirect status and Location header (e.g. "/" -> "/login").
	client := r.client
	if h.FollowRedirects != nil && !*h.FollowRedirects {
		c := *r.client
		c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		client = &c
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("http %s %s: %w", method, target, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body from %s %s: %w", method, target, err)
	}

	// body_to persists the response for the file/image/pdf assertion targets —
	// the http analogue of run's stdout_to (workdir-confined, create/truncate).
	if h.BodyTo != "" {
		dst, perr := security.ResolveWorkdirPath("http.body_to", r.workdir, h.BodyTo)
		if perr != nil {
			return nil, perr
		}
		if perr := os.MkdirAll(filepath.Dir(dst), 0o750); perr != nil {
			return nil, fmt.Errorf("creating directory for http.body_to %q: %w", h.BodyTo, perr)
		}
		if perr := os.WriteFile(dst, data, 0o600); perr != nil {
			return nil, fmt.Errorf("writing http.body_to %q: %w", h.BodyTo, perr)
		}
	}

	return &runner.Result{
		Command:    method + " " + target.String(),
		IsHTTP:     true,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       data,
		Duration:   elapsed,
	}, nil
}

// resolveURL joins a request path with the runner's base_url. An absolute path
// (http:// or https://) is used verbatim and ignores base_url.
func (r *Runner) resolveURL(path string) (*url.URL, error) {
	raw := path
	if !isAbsURL(path) {
		if r.baseURL == "" {
			return nil, fmt.Errorf("http path %q is relative but the runner has no base_url", path)
		}
		raw = strings.TrimRight(r.baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("resolved URL %q has no host", raw)
	}
	return u, nil
}

// checkPolicy enforces the network allowlist when one is configured.
func (r *Runner) checkPolicy(u *url.URL) error {
	if len(r.allow) == 0 {
		return nil
	}
	for _, a := range r.allow {
		if a == u.Hostname() || a == u.Host {
			return nil
		}
	}
	return &PolicyError{Host: u.Hostname(), Allow: r.allow}
}

func isAbsURL(p string) bool {
	return strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://")
}

// encodeBody serializes the request payload. Exactly one payload family may be
// set (the loader enforces it): `json:` marshals a structured value, `body:`
// sends a string verbatim, `body_file:` streams a workdir file binary-safe,
// and `form:`/`files:` send form fields — urlencoded on their own, multipart
// once a file part is involved. With none set, the request has no payload.
func (r *Runner) encodeBody(h *spec.HTTP) (io.Reader, string, error) {
	switch {
	case len(h.Files) > 0:
		return r.encodeMultipart(h.Form, h.Files)
	case len(h.Form) > 0:
		v := url.Values{}
		for k, val := range h.Form {
			v.Set(k, val)
		}
		return strings.NewReader(v.Encode()), "application/x-www-form-urlencoded", nil
	case h.BodyFile != "":
		path, err := security.ResolveWorkdirPath("http.body_file", r.workdir, h.BodyFile)
		if err != nil {
			return nil, "", err
		}
		data, err := os.ReadFile(path) //nolint:gosec // path is confined to the workdir above
		if err != nil {
			return nil, "", fmt.Errorf("reading http.body_file %q: %w", h.BodyFile, err)
		}
		return bytes.NewReader(data), detectContentType(data), nil
	case h.Body != "":
		return strings.NewReader(h.Body), "text/plain; charset=utf-8", nil
	case h.JSON != nil:
		b, err := json.Marshal(h.JSON)
		if err != nil {
			return nil, "", fmt.Errorf("encoding json body: %w", err)
		}
		return bytes.NewReader(b), "application/json", nil
	default:
		return nil, "", nil
	}
}

// encodeMultipart builds a multipart/form-data payload: every form field
// becomes a regular part and every file entry a file part read from the
// workdir. Fields are written in sorted order so the payload is deterministic.
func (r *Runner) encodeMultipart(form map[string]string, files []spec.FilePart) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := w.WriteField(k, form[k]); err != nil {
			return nil, "", fmt.Errorf("writing form field %q: %w", k, err)
		}
	}

	for _, f := range files {
		path, err := security.ResolveWorkdirPath("http.files.path", r.workdir, f.Path)
		if err != nil {
			return nil, "", err
		}
		data, err := os.ReadFile(path) //nolint:gosec // path is confined to the workdir above
		if err != nil {
			return nil, "", fmt.Errorf("reading http.files %q: %w", f.Path, err)
		}
		ct := f.ContentType
		if ct == "" {
			ct = detectContentType(data)
		}
		hdr := make(map[string][]string)
		hdr["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name=%q; filename=%q`, f.Field, filepath.Base(f.Path))}
		hdr["Content-Type"] = []string{ct}
		part, err := w.CreatePart(hdr)
		if err != nil {
			return nil, "", fmt.Errorf("creating multipart part %q: %w", f.Field, err)
		}
		if _, err := part.Write(data); err != nil {
			return nil, "", fmt.Errorf("writing multipart part %q: %w", f.Field, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, "", fmt.Errorf("finalizing multipart body: %w", err)
	}
	return &buf, w.FormDataContentType(), nil
}

// detectContentType sniffs the payload's media type, falling back to
// application/octet-stream for content the standard library cannot classify.
func detectContentType(data []byte) string {
	return http.DetectContentType(data)
}
