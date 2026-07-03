package engine

import (
	"strings"
	"testing"
)

// TestEngine_MockServerRoundTrip proves the full #24 flow with the http
// runner as the client: routes answer canned JSON, ${name.url} seeds the
// store, and mock asserts check count/header/body of what was sent.
func TestEngine_MockServerRoundTrip(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
runners:
  api:
    type: http
    base_url: ${api.url}
scenarios:
  - name: client posts a report
    mock_servers:
      - name: api
        routes:
          - method: POST
            path: /v1/reports
            status: 201
            json: { id: "r-1", ok: true }
    steps:
      - http:
          runner: api
          method: POST
          path: /v1/reports
          header: { Authorization: "Bearer tok-123" }
          json: { title: "report" }
      - assert:
          status: 201
          body:
            json: { path: "$.id", equals: "r-1" }
      - assert:
          mock:
            name: api
            path: /v1/reports
            method: POST
            count: 1
            header: { name: Authorization, matches: "^Bearer " }
            body:
              json: { path: "$.title", equals: "report" }
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

// TestEngine_MockUnmatchedRecordedAs404 proves an unmatched request answers
// 404 and is still recorded, and a failing count summarizes what was sent.
func TestEngine_MockUnmatchedRecordedAs404(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
runners:
  api:
    type: http
    base_url: ${stub.url}
scenarios:
  - name: typo'd path is visible in the failure
    mock_servers:
      - name: stub
        routes:
          - method: GET
            path: /right
    steps:
      - http:
          runner: api
          method: GET
          path: /wrong
      - assert:
          status: 404
      - assert:
          mock:
            name: stub
            path: /right
            count: 1
`)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed: %+v", res.Status, res.Scenarios)
	}
	var actual string
	for _, sc := range res.Scenarios {
		for _, st := range sc.Steps {
			for _, c := range st.Checks {
				if c != nil && !c.OK {
					actual = c.Actual
				}
			}
		}
	}
	if !strings.Contains(actual, "GET /wrong -> 404") {
		t.Errorf("failing count Actual = %q, want the recorded-request summary", actual)
	}
}

// TestEngine_SuiteMockServer proves a suite.setup mock_server seeds every
// scenario with its URL and is assertable from scenario steps.
func TestEngine_SuiteMockServer(t *testing.T) {
	t.Parallel()
	res := runSpec(t, `
version: "1"
suite:
  name: s
  setup:
    - mock_server:
        name: shared
        routes:
          - method: GET
            path: /ping
            body: pong
runners:
  api:
    type: http
    base_url: ${shared.url}
scenarios:
  - name: scenario hits the suite-wide mock
    steps:
      - http:
          runner: api
          method: GET
          path: /ping
      - assert:
          status: 200
          body: { equals: pong }
      - assert:
          mock:
            name: shared
            path: /ping
            method: GET
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}
