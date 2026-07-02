package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

func TestRunner_Do_GET(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Trace", "abc123")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"hello":"world"}`)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL})
	res, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/greet"})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if !res.IsHTTP {
		t.Error("IsHTTP = false, want true")
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
	if got := res.Header.Get("X-Trace"); got != "abc123" {
		t.Errorf("X-Trace = %q, want abc123", got)
	}
	if string(res.Body) != `{"hello":"world"}` {
		t.Errorf("body = %q", res.Body)
	}
}

func TestRunner_Do_POSTJSONBody(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL})
	res, err := r.Do(context.Background(), &spec.HTTP{
		Method: "POST",
		Path:   "/users",
		JSON:   map[string]any{"name": "Alice"},
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", res.StatusCode)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody["name"] != "Alice" {
		t.Errorf("server received body %v, want name=Alice", gotBody)
	}
}

// TestRunner_Do_RawStringBody: `body:` sends the string verbatim with a
// text/plain default Content-Type — the shape of text-first APIs (metrics
// exposition, message publishing, paste endpoints).
func TestRunner_Do_RawStringBody(t *testing.T) {
	t.Parallel()
	var gotBody, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL})
	res, err := r.Do(context.Background(), &spec.HTTP{
		Method: "POST",
		Path:   "/metrics/job/atago",
		Body:   "some_metric 3.14\n",
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if res.StatusCode != http.StatusAccepted {
		t.Errorf("status = %d, want 202", res.StatusCode)
	}
	if gotBody != "some_metric 3.14\n" {
		t.Errorf("server received body %q, want the raw string verbatim", gotBody)
	}
	if gotCT != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want the text/plain default", gotCT)
	}
}

// TestRunner_Do_RawBodyHeaderOverride: an explicit Content-Type header wins
// over the text/plain default, exactly like the json default.
func TestRunner_Do_RawBodyHeaderOverride(t *testing.T) {
	t.Parallel()
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL})
	if _, err := r.Do(context.Background(), &spec.HTTP{
		Method: "POST",
		Path:   "/upload",
		Header: map[string]string{"Content-Type": "application/x-ndjson"},
		Body:   `{"a":1}` + "\n" + `{"a":2}`,
	}); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if gotCT != "application/x-ndjson" {
		t.Errorf("Content-Type = %q, explicit header should win", gotCT)
	}
}

func TestRunner_Do_ExplicitHeaderOverridesContentType(t *testing.T) {
	t.Parallel()
	var gotCT, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		gotAuth = r.Header.Get("Authorization")
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL})
	_, err := r.Do(context.Background(), &spec.HTTP{
		Method: "POST",
		Path:   "/x",
		Header: map[string]string{"Content-Type": "application/vnd.api+json", "Authorization": "Bearer t"},
		JSON:   map[string]any{"a": 1},
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if gotCT != "application/vnd.api+json" {
		t.Errorf("Content-Type = %q, explicit header should win", gotCT)
	}
	if gotAuth != "Bearer t" {
		t.Errorf("Authorization = %q, want Bearer t", gotAuth)
	}
}

func TestRunner_Do_AbsolutePathIgnoresBaseURL(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: "http://unused.example.com"})
	res, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: srv.URL + "/abs"})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if res.StatusCode != http.StatusTeapot {
		t.Errorf("status = %d, want 418 (absolute path should hit the test server)", res.StatusCode)
	}
}

func TestRunner_Do_RelativePathWithoutBaseURL(t *testing.T) {
	t.Parallel()
	r := New(Config{})
	_, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/x"})
	if err == nil {
		t.Fatal("Do() error = nil, want error for relative path with no base_url")
	}
}

func TestRunner_Do_NetworkPolicyDenies(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Allow: []string{"api.allowed.example"}})
	_, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/x"})
	if err == nil {
		t.Fatal("Do() error = nil, want PolicyError")
	}
	var pe *PolicyError
	if !errors.As(err, &pe) {
		t.Fatalf("error = %T (%v), want *PolicyError", err, err)
	}
}

func TestRunner_Do_NetworkPolicyAllowsListedHost(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// 127.0.0.1 is the httptest host; allow both the bare host and host:port form.
	r := New(Config{BaseURL: srv.URL, Allow: []string{"127.0.0.1"}})
	res, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/x"})
	if err != nil {
		t.Fatalf("Do() error = %v, want allowed", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
}

func TestRunner_Do_TimeoutErrors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Timeout: 10 * time.Millisecond})
	if _, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/slow"}); err == nil {
		t.Fatal("Do() error = nil, want timeout error")
	}
}

func boolPtr(b bool) *bool { return &b }

// TestRunner_Do_MultipartUpload proves form fields and a workdir file arrive as
// one multipart/form-data request — the browser-style upload most self-hosted
// web apps expect.
func TestRunner_Do_MultipartUpload(t *testing.T) {
	t.Parallel()
	workdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workdir, "avatar.png"), []byte("\x89PNG-fake-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotField, gotFileName, gotFileBody, gotPartType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotField = r.FormValue("title")
		f, hdr, err := r.FormFile("upload")
		if err != nil {
			t.Errorf("FormFile: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func() { _ = f.Close() }()
		data, _ := io.ReadAll(f)
		gotFileName, gotFileBody = hdr.Filename, string(data)
		gotPartType = hdr.Header.Get("Content-Type")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Workdir: workdir})
	res, err := r.Do(context.Background(), &spec.HTTP{
		Method: "POST",
		Path:   "/upload",
		Form:   map[string]string{"title": "profile picture"},
		Files:  []spec.FilePart{{Field: "upload", Path: "avatar.png", ContentType: "image/png"}},
	})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", res.StatusCode)
	}
	if gotField != "profile picture" {
		t.Errorf("form field = %q, want 'profile picture'", gotField)
	}
	if gotFileName != "avatar.png" {
		t.Errorf("file name = %q, want avatar.png", gotFileName)
	}
	if gotFileBody != "\x89PNG-fake-bytes" {
		t.Errorf("file body = %q", gotFileBody)
	}
	if gotPartType != "image/png" {
		t.Errorf("part Content-Type = %q, want image/png", gotPartType)
	}
}

// TestRunner_Do_FormURLEncoded proves form fields without files are sent as
// application/x-www-form-urlencoded.
func TestRunner_Do_FormURLEncoded(t *testing.T) {
	t.Parallel()
	var gotContentType, gotUser string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_ = r.ParseForm()
		gotUser = r.PostFormValue("user")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Workdir: t.TempDir()})
	if _, err := r.Do(context.Background(), &spec.HTTP{
		Method: "POST",
		Path:   "/login",
		Form:   map[string]string{"user": "alice"},
	}); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want urlencoded", gotContentType)
	}
	if gotUser != "alice" {
		t.Errorf("user = %q, want alice", gotUser)
	}
}

// TestRunner_Do_BodyFile streams a binary workdir file as the raw request body.
func TestRunner_Do_BodyFile(t *testing.T) {
	t.Parallel()
	workdir := t.TempDir()
	raw := []byte{0x00, 0x01, 0xFF, 0xFE, 'a', 't', 'a', 'g', 'o'}
	if err := os.WriteFile(filepath.Join(workdir, "blob.bin"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Workdir: workdir})
	if _, err := r.Do(context.Background(), &spec.HTTP{Method: "PUT", Path: "/blob", BodyFile: "blob.bin"}); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if !bytes.Equal(gotBody, raw) {
		t.Errorf("body = %v, want %v (binary must survive verbatim)", gotBody, raw)
	}
}

// TestRunner_Do_BodyFileEscapesWorkdir proves body_file cannot read outside the
// scenario workdir.
func TestRunner_Do_BodyFileEscapesWorkdir(t *testing.T) {
	t.Parallel()
	r := New(Config{BaseURL: "http://127.0.0.1:1", Workdir: t.TempDir()})
	_, err := r.Do(context.Background(), &spec.HTTP{Method: "PUT", Path: "/x", BodyFile: "../../etc/passwd"})
	if err == nil {
		t.Fatal("Do() error = nil, want a workdir-confinement error")
	}
}

// TestRunner_Do_BodyTo writes the response body into the workdir for later
// file/image/pdf assertions.
func TestRunner_Do_BodyTo(t *testing.T) {
	t.Parallel()
	workdir := t.TempDir()
	payload := []byte("downloaded artifact \x00\x01")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Workdir: workdir})
	res, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/file", BodyTo: "out/artifact.bin"})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if !bytes.Equal(res.Body, payload) {
		t.Errorf("in-memory body = %q, want %q (body_to must not consume it)", res.Body, payload)
	}
	got, err := os.ReadFile(filepath.Join(workdir, "out", "artifact.bin"))
	if err != nil {
		t.Fatalf("reading body_to file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("body_to file = %q, want %q", got, payload)
	}
}

// TestRunner_Do_FollowRedirects pins both redirect behaviors: the default
// follows a 302 to its target; follow_redirects false surfaces the 302 itself
// with its Location header intact.
func TestRunner_Do_FollowRedirects(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "login page")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	r := New(Config{BaseURL: srv.URL, Workdir: t.TempDir()})

	followed, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if followed.StatusCode != http.StatusOK || string(followed.Body) != "login page" {
		t.Errorf("default follow: status=%d body=%q, want the login page", followed.StatusCode, followed.Body)
	}

	raw, err := r.Do(context.Background(), &spec.HTTP{Method: "GET", Path: "/", FollowRedirects: boolPtr(false)})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if raw.StatusCode != http.StatusFound {
		t.Errorf("no-follow status = %d, want 302", raw.StatusCode)
	}
	if got := raw.Header.Get("Location"); got != "/login" {
		t.Errorf("Location = %q, want /login", got)
	}
}
