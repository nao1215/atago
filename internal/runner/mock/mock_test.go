package mock

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

func startTest(t *testing.T, ms *spec.MockServer, specDir string) *Server {
	t.Helper()
	s, err := Start(context.Background(), ms, specDir)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(s.Stop)
	return s
}

// do issues one request and returns the status, headers, and drained body —
// the response never escapes the helper, so closing stays in one place.
func do(t *testing.T, method, url string, body io.Reader) (int, http.Header, []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("close body: %v", err)
		}
	}()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return res.StatusCode, res.Header, b
}

// TestServe_RouteMatchingAndRecording proves exact method+path matching,
// canned JSON responses, 404-with-recording for unmatched requests, and
// ephemeral-port URL/Port reporting (#24).
func TestServe_RouteMatchingAndRecording(t *testing.T) {
	t.Parallel()
	s := startTest(t, &spec.MockServer{
		Name: "api",
		Routes: []spec.MockRoute{
			{Method: http.MethodPost, Path: "/v1/reports", Status: http.StatusCreated, JSON: map[string]any{"id": "r-1", "ok": true}},
			{Method: http.MethodGet, Path: "/v1/reports/r-1", Body: "plain"},
		},
	}, t.TempDir())

	if s.URL() == "" || s.Port() == "" {
		t.Fatalf("URL/Port empty: %q %q", s.URL(), s.Port())
	}

	// The query string is excluded from matching.
	status, header, body := do(t, http.MethodPost, s.URL()+"/v1/reports?ignored=1", strings.NewReader(`{"title":"report"}`))
	if status != http.StatusCreated {
		t.Errorf("status = %d, want 201", status)
	}
	if !bytes.Contains(body, []byte(`"id":"r-1"`)) {
		t.Errorf("body = %s, want the canned json", body)
	}
	if ct := header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}

	status2, _, _ := do(t, http.MethodGet, s.URL()+"/nope", nil)
	if status2 != http.StatusNotFound {
		t.Errorf("unmatched status = %d, want 404", status2)
	}

	recs := s.Records()
	if len(recs) != 2 {
		t.Fatalf("records = %d, want 2 (unmatched requests are recorded too)", len(recs))
	}
	if recs[0].Method != http.MethodPost || recs[0].Path != "/v1/reports" || recs[0].Status != http.StatusCreated {
		t.Errorf("record[0] = %+v, want POST /v1/reports 201", recs[0])
	}
	if !bytes.Contains(recs[0].Body, []byte("report")) {
		t.Errorf("record[0].Body = %s, want the request payload", recs[0].Body)
	}
	if recs[1].Status != http.StatusNotFound {
		t.Errorf("record[1].Status = %d, want 404", recs[1].Status)
	}
}

// TestServe_BodyFileConfined proves body_file reads a spec-relative file at
// request time and rejects escapes from the spec directory.
func TestServe_BodyFileConfined(t *testing.T) {
	t.Parallel()
	specDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(specDir, "canned.json"), []byte(`{"from":"file"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s := startTest(t, &spec.MockServer{
		Name: "api",
		Routes: []spec.MockRoute{
			{Method: http.MethodGet, Path: "/ok", BodyFile: "canned.json"},
			{Method: http.MethodGet, Path: "/escape", BodyFile: "../outside.txt"},
		},
	}, specDir)

	status, _, body := do(t, http.MethodGet, s.URL()+"/ok", nil)
	if status != http.StatusOK || !bytes.Contains(body, []byte("from")) {
		t.Errorf("status/body = %d/%s, want 200 with file content", status, body)
	}

	status2, _, _ := do(t, http.MethodGet, s.URL()+"/escape", nil)
	if status2 != http.StatusInternalServerError {
		t.Errorf("escape status = %d, want 500 (spec-dir confinement)", status2)
	}
}

// TestServe_DelayAndHeaders proves the per-route delay and extra response
// headers.
func TestServe_DelayAndHeaders(t *testing.T) {
	t.Parallel()
	s := startTest(t, &spec.MockServer{
		Name: "slow",
		Routes: []spec.MockRoute{
			{Method: http.MethodGet, Path: "/", Delay: "100ms", Header: map[string]string{"X-Mock": "yes"}},
		},
	}, t.TempDir())
	start := time.Now()
	_, header, _ := do(t, http.MethodGet, s.URL()+"/", nil)
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Errorf("elapsed = %s, want >= 100ms (route delay)", elapsed)
	}
	if header.Get("X-Mock") != "yes" {
		t.Errorf("X-Mock header missing")
	}
}

// TestRecords_ThreadSafe proves concurrent clients record without races
// (run with -race in CI).
func TestRecords_ThreadSafe(t *testing.T) {
	t.Parallel()
	s := startTest(t, &spec.MockServer{
		Name:   "busy",
		Routes: []spec.MockRoute{{Method: http.MethodGet, Path: "/"}},
	}, t.TempDir())
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, s.URL()+"/", nil)
			if err != nil {
				return
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			_, _ = io.Copy(io.Discard, res.Body)
			_ = res.Body.Close()
		})
	}
	wg.Wait()
	if got := len(s.Records()); got != 20 {
		t.Errorf("records = %d, want 20", got)
	}
}

// TestRequestLog renders the per-request artifact line for a matched and an
// unmatched request, covering RequestLog and the 404 recording path.
func TestRequestLog(t *testing.T) {
	t.Parallel()
	s := startTest(t, &spec.MockServer{
		Name: "api",
		Routes: []spec.MockRoute{
			{Method: "GET", Path: "/health", Body: "ok"},
		},
	}, t.TempDir())

	do(t, http.MethodGet, s.URL()+"/health", nil)
	do(t, http.MethodPost, s.URL()+"/missing", strings.NewReader("payload"))

	log := s.RequestLog()
	if !strings.Contains(log, "1: GET /health -> 200") {
		t.Errorf("RequestLog missing matched line:\n%s", log)
	}
	if !strings.Contains(log, "2: POST /missing -> 404") {
		t.Errorf("RequestLog missing unmatched line:\n%s", log)
	}
	if !strings.Contains(log, "7 body bytes") {
		t.Errorf("RequestLog missing recorded body size:\n%s", log)
	}
}

// TestStop_NilSafe covers the nil/zero guards in Stop so a double-stop or a
// never-started server does not panic.
func TestStop_NilSafe(t *testing.T) {
	t.Parallel()
	var nilServer *Server
	nilServer.Stop() // must not panic
	(&Server{}).Stop()
}
