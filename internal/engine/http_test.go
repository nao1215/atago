package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// loginServer is a tiny stand-in API: POST /login returns a token (in the body)
// and a session id (in a header); GET /me echoes the bearer token back so a test
// can prove the token bound from the login response flowed into the next request.
func loginServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("X-Session", "sess-42")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"token":"secret-xyz"}`)
	})
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		body, _ := json.Marshal(map[string]string{"authorization": auth})
		_, _ = w.Write(body)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func runHTTPSpec(t *testing.T, src string) *SuiteResult {
	t.Helper()
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return New().Run(context.Background(), s, "t.atago.yaml")
}

func TestEngine_HTTPWorkflow_Passing(t *testing.T) {
	t.Parallel()
	srv := loginServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: api
runners:
  api:
    type: http
    base_url: %s
scenarios:
  - name: login then call an authed endpoint with the bound token
    steps:
      - http:
          runner: api
          method: POST
          path: /login
          json:
            user: alice
      - assert:
          status: 200
      - assert:
          header:
            name: X-Session
            contains: sess
      - store:
          name: token
          from:
            body:
              json:
                path: $.token
      - http:
          runner: api
          method: GET
          path: /me
          header:
            Authorization: "Bearer ${token}"
      - assert:
          status: 200
      - assert:
          body:
            json:
              path: $.authorization
              equals: "Bearer secret-xyz"
`, srv.URL)

	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

func TestEngine_HTTPStatusAssertionFails(t *testing.T) {
	t.Parallel()
	srv := loginServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: api
runners:
  api:
    type: http
    base_url: %s
scenarios:
  - name: wrong status fails
    steps:
      - http:
          runner: api
          method: POST
          path: /login
      - assert:
          status: 500
`, srv.URL)

	res := runHTTPSpec(t, src)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
}

func TestEngine_HTTPNetworkPolicyViolation(t *testing.T) {
	t.Parallel()
	srv := loginServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: api
permissions:
  network:
    allow:
      - api.allowed.example
runners:
  api:
    type: http
    base_url: %s
scenarios:
  - name: denied host
    steps:
      - http:
          runner: api
          method: GET
          path: /me
`, srv.URL)

	res := runHTTPSpec(t, src)
	if res.Status != StatusError {
		t.Errorf("status = %s, want error", res.Status)
	}
	if !res.SecurityViolation {
		t.Error("SecurityViolation = false, want true (denied host)")
	}
	if !strings.Contains(res.Scenarios[0].Steps[0].ErrMsg, "network policy denies") {
		t.Errorf("err = %q, want network policy denial", res.Scenarios[0].Steps[0].ErrMsg)
	}
}

func TestEngine_HTTPUnknownRunner(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: api
scenarios:
  - name: missing runner
    steps:
      - http:
          runner: nope
          method: GET
          path: /x
`
	res := runHTTPSpec(t, src)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
}

// flakyServer answers /job with "pending" until the given attempt, then "done".
// It stands in for any eventually-consistent endpoint (an async job, a metric
// that appears after a scrape).
func flakyServer(t *testing.T, readyOn int) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	attempts := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		n := attempts
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if n < readyOn {
			_, _ = io.WriteString(w, `{"state":"pending"}`)
			return
		}
		_, _ = io.WriteString(w, `{"state":"done"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestEngine_HTTPRetryUntilPasses proves an http step polls with retry/until
// exactly like a run step: the endpoint answers "pending" twice, "done" on the
// third attempt, and the scenario passes with the last response observable by
// later asserts.
func TestEngine_HTTPRetryUntilPasses(t *testing.T) {
	t.Parallel()
	srv := flakyServer(t, 3)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: api
runners:
  api:
    type: http
    base_url: %s
scenarios:
  - name: poll until the async job completes
    steps:
      - http:
          runner: api
          method: GET
          path: /job
          retry:
            times: 5
            interval: 10ms
            until:
              body:
                json:
                  path: "$.state"
                  equals: done
      - assert:
          status: 200
          body:
            json:
              path: "$.state"
              equals: done
`, srv.URL)

	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_HTTPRetryBudgetExhausted proves a never-satisfied until fails the
// http step (reported like an assertion), matching run-retry semantics.
func TestEngine_HTTPRetryBudgetExhausted(t *testing.T) {
	t.Parallel()
	srv := flakyServer(t, 100) // never ready within the budget
	src := fmt.Sprintf(`
version: "1"
suite:
  name: api
runners:
  api:
    type: http
    base_url: %s
scenarios:
  - name: the job never completes
    steps:
      - http:
          runner: api
          method: GET
          path: /job
          retry:
            times: 3
            interval: 1ms
            until:
              body:
                json:
                  path: "$.state"
                  equals: done
`, srv.URL)

	res := runHTTPSpec(t, src)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	checks := res.Scenarios[0].Steps[0].Checks
	if len(checks) == 0 {
		t.Fatal("no until checks recorded on the http step")
	}
}
