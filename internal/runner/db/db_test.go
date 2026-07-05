package db

import (
	"context"
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		driver     string
		dsn        string
		wantDriver string
		wantSource string
		wantErr    bool
	}{
		{name: "sqlite scheme", dsn: "sqlite:./app.db", wantDriver: "sqlite", wantSource: "./app.db"},
		{name: "sqlite3 scheme", dsn: "sqlite3:/tmp/x.db", wantDriver: "sqlite", wantSource: "/tmp/x.db"},
		{name: "sqlite memory", dsn: "sqlite::memory:", wantDriver: "sqlite", wantSource: ":memory:"},
		{name: "postgres url kept verbatim", dsn: "postgres://u:p@h:5432/db?sslmode=disable", wantDriver: "postgres", wantSource: "postgres://u:p@h:5432/db?sslmode=disable"},
		{name: "postgresql alias", dsn: "postgresql://u@h/db", wantDriver: "postgres", wantSource: "postgresql://u@h/db"},
		{name: "mysql url to native", dsn: "mysql://u:p@h:3306/db?parseTime=true", wantDriver: "mysql", wantSource: "u:p@tcp(h:3306)/db?parseTime=true"},
		{name: "explicit driver with native mysql dsn", driver: "mysql", dsn: "u:p@tcp(h:3306)/db", wantDriver: "mysql", wantSource: "u:p@tcp(h:3306)/db"},
		{name: "explicit sqlite driver bare path", driver: "sqlite", dsn: "/tmp/x.db", wantDriver: "sqlite", wantSource: "/tmp/x.db"},
		{name: "empty dsn errors", dsn: "", wantErr: true},
		{name: "unknown scheme without driver errors", dsn: "weird:thing", wantErr: true},
		// An explicit but unsupported driver must fail fast instead of
		// silently falling back to DSN-scheme inference.
		{name: "invalid explicit driver with sqlite dsn errors", driver: "sqllite", dsn: "sqlite:./app.db", wantErr: true},
		{name: "invalid explicit driver with postgres dsn errors", driver: "postgre", dsn: "postgres://u@h/db", wantErr: true},
		{name: "invalid explicit driver with mysql dsn errors", driver: "mariadb", dsn: "mysql://u:p@h:3306/db", wantErr: true},
		{name: "invalid explicit driver with native dsn errors", driver: "nope", dsn: "u:p@tcp(h:3306)/db", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Resolve(tt.driver, tt.dsn)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Resolve(%q,%q) error = nil, want error", tt.driver, tt.dsn)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%q,%q) error = %v", tt.driver, tt.dsn, err)
			}
			if got.Driver != tt.wantDriver {
				t.Errorf("driver = %q, want %q", got.Driver, tt.wantDriver)
			}
			if got.DataSource != tt.wantSource {
				t.Errorf("source = %q, want %q", got.DataSource, tt.wantSource)
			}
		})
	}
}

func TestRunner_Query_SQLite(t *testing.T) {
	t.Parallel()
	cfg, err := Resolve("", "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()

	ctx := context.Background()
	if _, err := r.Query(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatalf("create: %v", err)
	}
	ins, err := r.Query(ctx, "INSERT INTO users (name) VALUES ('alice'), ('bob')")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if ins.RowsAffected != 2 {
		t.Errorf("RowsAffected = %d, want 2", ins.RowsAffected)
	}
	if string(ins.RowsJSON) != "[]" {
		t.Errorf("insert RowsJSON = %q, want []", ins.RowsJSON)
	}

	sel, err := r.Query(ctx, "SELECT id, name FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if !sel.IsDB {
		t.Error("IsDB = false, want true")
	}
	got := string(sel.RowsJSON)
	for _, want := range []string{`"name":"alice"`, `"name":"bob"`, `"id":1`} {
		if !strings.Contains(got, want) {
			t.Errorf("rows JSON %q missing %q", got, want)
		}
	}
}

// Regression for issue #31: a WITH ... INSERT CTE without RETURNING must run via
// ExecContext so RowsAffected is captured (previously it was routed as a query,
// losing the count and returning an empty result set).
func TestRunner_Query_ModifyingCTE(t *testing.T) {
	t.Parallel()
	cfg, err := Resolve("", "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()

	ctx := context.Background()
	if _, err := r.Query(ctx, "CREATE TABLE t (a INTEGER)"); err != nil {
		t.Fatalf("create: %v", err)
	}
	res, err := r.Query(ctx, "WITH src(v) AS (VALUES (1),(2),(3)) INSERT INTO t (a) SELECT v FROM src")
	if err != nil {
		t.Fatalf("modifying CTE: %v", err)
	}
	if res.RowsAffected != 3 {
		t.Errorf("RowsAffected = %d, want 3 (affected-row count lost?)", res.RowsAffected)
	}
	if string(res.RowsJSON) != "[]" {
		t.Errorf("RowsJSON = %q, want [] for a non-row statement", res.RowsJSON)
	}
}

// TestRunner_Query_CommentedSelect is a regression: a SELECT preceded by a SQL
// comment must still route through QueryContext and return its rows. A leading
// comment hid the SELECT verb, misrouting the statement to ExecContext, which
// returns no rows — so the row assertion saw nothing.
func TestRunner_Query_CommentedSelect(t *testing.T) {
	t.Parallel()
	cfg, err := Resolve("", "sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()

	ctx := context.Background()
	if _, err := r.Query(ctx, "CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := r.Query(ctx, "INSERT INTO t (name) VALUES ('alice')"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	for _, q := range []string{
		"-- fetch the row\nSELECT id, name FROM t",
		"/* preamble */ SELECT id, name FROM t",
	} {
		sel, err := r.Query(ctx, q)
		if err != nil {
			t.Fatalf("query %q: %v", q, err)
		}
		if !strings.Contains(string(sel.RowsJSON), `"name":"alice"`) {
			t.Errorf("commented SELECT %q returned no rows: RowsJSON=%q (misrouted to Exec?)", q, sel.RowsJSON)
		}
	}
}

func TestRunner_Query_SyntaxError(t *testing.T) {
	t.Parallel()
	cfg, _ := Resolve("", "sqlite::memory:")
	r, _ := Open(cfg)
	defer func() { _ = r.Close() }()
	if _, err := r.Query(context.Background(), "SELECT * FROM nonexistent_table"); err == nil {
		t.Fatal("Query() error = nil, want error for missing table")
	}
}

func TestIsRowReturning(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"SELECT 1":                              true,
		"  select * from t":                     true,
		"WITH x AS (...) SELECT *":              true,
		"PRAGMA table_info(t)":                  true,
		"INSERT INTO t VALUES (1)":              false,
		"UPDATE t SET a=1":                      false,
		"CREATE TABLE t (a int)":                false,
		"DELETE FROM t RETURNING id":            true,
		"INSERT INTO t VALUES (1) RETURNING id": true,
		// Issue #31: a data-modifying CTE without RETURNING must route to Exec.
		"WITH x AS (SELECT 1) INSERT INTO t SELECT * FROM x":              false,
		"WITH x AS (SELECT 1) UPDATE t SET a = 1":                         false,
		"WITH x AS (SELECT 1) DELETE FROM t":                              false,
		"WITH x AS (SELECT 1) INSERT INTO t SELECT * FROM x RETURNING id": true,
		"WITH x AS (SELECT 1) DELETE FROM t RETURNING *":                  true,
		"WITH RECURSIVE t(n) AS (SELECT 1) SELECT * FROM t":               true,
		// A RETURNING token INSIDE a string literal must not route the statement
		// as row-returning (it would lose the affected-row count).
		"INSERT INTO logs (msg) VALUES ('order RETURNING to sender')":    false,
		"UPDATE t SET note = 'see RETURNING policy' WHERE id = 1":        false,
		"WITH x AS (SELECT 1) INSERT INTO logs VALUES ('a RETURNING b')": false,
		// A leading or embedded SQL comment must not hide the main verb: a
		// commented SELECT was misrouted to Exec, losing its rows.
		"-- fetch users\nSELECT 1":              true,
		"/* preamble */ SELECT 1":               true,
		"SELECT/* inline */1":                   true,
		"-- note\nINSERT INTO t VALUES (1)":     false,
		"/* c */ WITH x AS (SELECT 1) SELECT 1": true,
		// A comment marker inside a string literal is data, not a comment: it must
		// not be stripped, and the statement still routes to Exec.
		"INSERT INTO logs (msg) VALUES ('a -- b')": false,
		"INSERT INTO logs (msg) VALUES ('a /* b')": false,
	}
	for q, want := range cases {
		if got := isRowReturning(q); got != want {
			t.Errorf("isRowReturning(%q) = %v, want %v", q, got, want)
		}
	}
}
