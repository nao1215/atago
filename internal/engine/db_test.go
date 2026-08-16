package engine

import (
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

func TestEngine_DBWorkflow_SQLite(t *testing.T) {
	t.Parallel()
	// A per-scenario sqlite file under ${workdir} gives hermetic isolation.
	src := `
version: "1"
suite:
  name: db
runners:
  store:
    type: db
    dsn: "sqlite:${workdir}/app.db"
scenarios:
  - name: create insert select with row assertions and value binding
    steps:
      - query:
          runner: store
          sql: "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, role TEXT)"
      - query:
          runner: store
          sql: "INSERT INTO users (name, role) VALUES ('alice','admin'), ('bob','user')"
      - assert:
          rows:
            json:
              path: "$"
              length: 0
      - query:
          runner: store
          sql: "SELECT id, name, role FROM users ORDER BY id"
      - assert:
          rows:
            json:
              path: "$"
              length: 2
      - assert:
          rows:
            json:
              path: "$[0].name"
              equals: alice
      - store:
          name: admin_id
          from:
            rows:
              json:
                path: "$[0].id"
      - query:
          runner: store
          sql: "SELECT role FROM users WHERE id = ${admin_id}"
      - assert:
          rows:
            json:
              path: "$[0].role"
              equals: admin
`
	res := runSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

func TestEngine_DBQueryError(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: db
runners:
  store:
    type: db
    dsn: "sqlite:${workdir}/app.db"
scenarios:
  - name: query against a missing table errors
    steps:
      - query:
          runner: store
          sql: "SELECT * FROM does_not_exist"
`
	res := runSpec(t, src)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
}

func TestEngine_DBUnknownRunner(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: db
scenarios:
  - name: query references an undeclared runner
    steps:
      - query:
          runner: missing
          sql: "SELECT 1"
`
	// An undeclared runner is a load-time validation error (exit 2), not a
	// mid-run execution error; the engine keeps a runtime check as a backstop.
	if _, err := loader.LoadBytes("t.atago.yaml", []byte(src)); err == nil || !strings.Contains(err.Error(), "is not declared") {
		t.Fatalf("LoadBytes() error = %v, want an undeclared-runner validation error", err)
	}
}

func TestEngine_DBScenarioIsolation(t *testing.T) {
	t.Parallel()
	// Two scenarios share a runner whose dsn references ${workdir}; each gets its
	// own database file, so a table created in the first is absent in the second.
	src := `
version: "1"
suite:
  name: db
runners:
  store:
    type: db
    dsn: "sqlite:${workdir}/app.db"
scenarios:
  - name: first creates a table
    steps:
      - query:
          runner: store
          sql: "CREATE TABLE t (a INTEGER)"
      - query:
          runner: store
          sql: "INSERT INTO t VALUES (1)"
      - assert:
          rows:
            json:
              path: "$"
              length: 0
  - name: second sees a fresh database (table absent)
    steps:
      - query:
          runner: store
          sql: "SELECT * FROM t"
`
	res := runSpec(t, src)
	if res.Scenarios[0].Status != StatusPassed {
		t.Errorf("scenario[0] = %s, want passed", res.Scenarios[0].Status)
	}
	if res.Scenarios[1].Status != StatusError {
		t.Errorf("scenario[1] = %s, want error (table should not exist in a fresh db)", res.Scenarios[1].Status)
	}
}

// TestEngine_DBNetworkPolicyViolation is a regression: `permissions.network`
// says egress to an unlisted host is a policy violation, and the http, grpc,
// and ssh runners all check before dialing — the db runner did not. A `query:`
// against a denied host resolved it and connected, reporting a plain connection
// error (exit 4) if anything, so a spec could reach a database the policy was
// written to keep it away from.
func TestEngine_DBNetworkPolicyViolation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		dsn  string
	}{
		{"postgres url", "postgres://u:p@denied.example:5432/app?sslmode=disable"},
		{"mysql url", "mysql://u:p@denied.example:3306/app"},
		{"mysql native", "u:p@tcp(denied.example:3306)/app"},
		{"postgres keyword dsn", "host=denied.example port=5432 user=u dbname=app sslmode=disable"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			driver := "postgres"
			if strings.Contains(tt.name, "mysql") {
				driver = "mysql"
			}
			src := `
version: "1"
suite:
  name: db
permissions:
  network:
    allow:
      - db.allowed.example
runners:
  store:
    type: db
    driver: ` + driver + `
    dsn: "` + tt.dsn + `"
scenarios:
  - name: denied host
    steps:
      - query:
          runner: store
          sql: "SELECT 1"
`
			res := runSpec(t, src)
			if res.Status != StatusError {
				t.Errorf("status = %s, want error", res.Status)
			}
			if !res.SecurityViolation {
				t.Error("SecurityViolation = false, want true (denied host)")
			}
			if got := res.Scenarios[0].Steps[0].ErrMsg; !strings.Contains(got, "network policy denies") {
				t.Errorf("err = %q, want a network policy denial", got)
			}
		})
	}
}

// TestEngine_DBNetworkPolicyAllowsLocalAndListed proves the check does not
// deny what it should not: a sqlite dsn names a file rather than a host, so it
// reaches no network at all and must run under any allowlist, and a listed host
// is permitted.
func TestEngine_DBNetworkPolicyAllowsLocalAndListed(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: db
permissions:
  network:
    allow:
      - db.allowed.example
runners:
  store:
    type: db
    dsn: sqlite:${workdir}/app.db
scenarios:
  - name: a file-backed database is not egress
    steps:
      - query:
          runner: store
          sql: "CREATE TABLE t (a INTEGER)"
      - query:
          runner: store
          sql: "SELECT * FROM t"
      - assert:
          rows:
            json:
              path: "$"
              length: 0
`
	res := runSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed (sqlite reaches no host)", res.Status)
	}
	if res.SecurityViolation {
		t.Error("SecurityViolation = true for a sqlite dsn, which contacts no host")
	}
}
