// Package mock implements the declarative stub HTTP server behind
// `mock_servers:` (#24): canned routes served on an ephemeral loopback port,
// with every incoming request recorded for the `mock:` assertion target — so
// API-client CLIs (gh-style tools, cloud CLIs, webhook senders) are testable
// fully offline.
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
)

// maxRecordedBody caps a recorded request body so a runaway client cannot
// exhaust memory; the response is unaffected.
const maxRecordedBody = 8 << 20 // 8 MiB

// Record is one incoming request as the mock server observed it.
type Record struct {
	Method string
	Path   string
	Header http.Header
	Body   []byte
	// Status is the response status the mock answered with (404 for an
	// unmatched request — recorded so asserts can catch typo'd paths).
	Status int
}

// Server is a running mock server.
type Server struct {
	name    string
	specDir string
	routes  []spec.MockRoute
	lis     net.Listener
	srv     *http.Server

	mu      sync.Mutex
	records []Record
}

// Start launches the mock server on an ephemeral 127.0.0.1 port. The returned
// server is ready to accept connections when Start returns (the listener
// binds synchronously); ctx only scopes the bind itself.
func Start(ctx context.Context, ms *spec.MockServer, specDir string) (*Server, error) {
	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("mock server %q: listen: %w", ms.Name, err)
	}
	s := &Server{name: ms.Name, specDir: specDir, routes: ms.Routes, lis: lis}
	s.srv = &http.Server{Handler: s, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = s.srv.Serve(lis) }()
	return s, nil
}

// Name identifies the server in diagnostics and asserts.
func (s *Server) Name() string { return s.name }

// URL is the base URL clients hit, seeded as ${<name>.url}.
func (s *Server) URL() string { return "http://" + s.lis.Addr().String() }

// Port is the bound port, seeded as ${<name>.port}.
func (s *Server) Port() string {
	_, port, _ := net.SplitHostPort(s.lis.Addr().String())
	return port
}

// Stop shuts the listener down. Recorded requests remain readable.
func (s *Server) Stop() {
	if s == nil || s.srv == nil {
		return
	}
	_ = s.srv.Close()
}

// Records returns a snapshot of every request observed so far, in arrival
// order.
func (s *Server) Records() []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Record, len(s.records))
	copy(out, s.records)
	return out
}

// RequestLog renders the recorded requests as one line each — the durable
// artifact written next to service logs when a scenario fails.
func (s *Server) RequestLog() string {
	var b strings.Builder
	for i, r := range s.Records() {
		fmt.Fprintf(&b, "%d: %s %s -> %d (%d body bytes)\n", i+1, r.Method, r.Path, r.Status, len(r.Body))
	}
	return b.String()
}

// ServeHTTP matches the request against the routes (exact method+path, query
// string excluded), answers the canned response, and records the request —
// an unmatched request answers 404 and is still recorded.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, maxRecordedBody))
	rec := Record{Method: r.Method, Path: r.URL.Path, Header: r.Header.Clone(), Body: body}

	route := s.match(r.Method, r.URL.Path)
	if route == nil {
		rec.Status = http.StatusNotFound
		s.record(rec)
		http.Error(w, fmt.Sprintf("mock server %q has no route for %s %s", s.name, r.Method, r.URL.Path), http.StatusNotFound)
		return
	}

	if route.Delay != "" {
		if d, err := time.ParseDuration(route.Delay); err == nil { // validated at load time
			time.Sleep(d)
		}
	}

	status := route.Status
	if status == 0 {
		status = http.StatusOK
	}
	payload, contentType, err := routePayload(route, s.specDir)
	if err != nil {
		rec.Status = http.StatusInternalServerError
		s.record(rec)
		http.Error(w, fmt.Sprintf("mock server %q: %v", s.name, err), http.StatusInternalServerError)
		return
	}
	rec.Status = status
	s.record(rec)

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	for k, v := range route.Header {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

// match returns the first route with the request's exact method and path.
func (s *Server) match(method, path string) *spec.MockRoute {
	for i := range s.routes {
		rt := &s.routes[i]
		if strings.EqualFold(rt.Method, method) && rt.Path == path {
			return rt
		}
	}
	return nil
}

// routePayload builds the response body: inline JSON (marshaled), inline
// text, or a spec-relative file read at request time (confined to the spec
// directory like snapshot goldens).
func routePayload(rt *spec.MockRoute, specDir string) (body []byte, contentType string, err error) {
	switch {
	case rt.JSON != nil:
		b, err := json.Marshal(rt.JSON)
		if err != nil {
			return nil, "", fmt.Errorf("route %s %s: marshal json: %w", rt.Method, rt.Path, err)
		}
		return b, "application/json", nil
	case rt.BodyFile != "":
		abs, err := security.ResolveSpecPath("mock route body_file", specDir, rt.BodyFile)
		if err != nil {
			return nil, "", err
		}
		b, err := os.ReadFile(abs) //nolint:gosec // confined to the spec directory above
		if err != nil {
			return nil, "", fmt.Errorf("route %s %s: %w", rt.Method, rt.Path, err)
		}
		return b, "", nil
	default:
		return []byte(rt.Body), "", nil
	}
}

func (s *Server) record(r Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, r)
}
