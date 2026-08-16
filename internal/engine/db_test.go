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
		// #497: hostaddr is the address lib/pq dials, so a dsn that names only
		// it must be denied like any other peer instead of slipping through as
		// a dsn that names no host.
		{"postgres hostaddr dsn", "hostaddr=127.0.0.1 port=1 user=u dbname=app sslmode=disable"},
		// #497: a host beside a hostaddr is not a way to launder the address —
		// lib/pq dials the hostaddr and keeps the host only as a name.
		{"postgres allowed host with denied hostaddr", "host=db.allowed.example hostaddr=127.0.0.1 port=1 user=u sslmode=disable"},
		// #497: a failover list must have every entry checked, not only the
		// first one the driver would try.
		{"postgres host failover list", "host=db.allowed.example,denied.example port=5432 user=u sslmode=disable"},
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

// TestEngine_DBNetworkPolicyAllowsQuotedHost is a regression for #497: libpq
// lets a value be single-quoted, and the quotes belong to the dsn syntax rather
// than to the host. Keeping them made the check compare "'db.allowed.example'"
// against the allowlist and deny a host the policy names — so the spec failed
// with a policy violation instead of the connection error it should have
// reached.
func TestEngine_DBNetworkPolicyAllowsQuotedHost(t *testing.T) {
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
    driver: postgres
    dsn: "host='db.allowed.example' port=5432 user=u dbname=app sslmode=disable"
scenarios:
  - name: quoted allowed host
    steps:
      - query:
          runner: store
          sql: "SELECT 1"
`
	res := runSpec(t, src)
	if res.SecurityViolation {
		t.Errorf("SecurityViolation = true for a quoted allowlisted host: %q",
			res.Scenarios[0].Steps[0].ErrMsg)
	}
	// The connection itself is expected to fail (the host does not resolve);
	// what matters is that it failed as a connection, not as a policy denial.
	if got := res.Scenarios[0].Steps[0].ErrMsg; strings.Contains(got, "network policy denies") {
		t.Errorf("err = %q, want no policy denial for an allowlisted host", got)
	}
}
