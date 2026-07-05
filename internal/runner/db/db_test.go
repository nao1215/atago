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

// TestNormalizeValue covers the []byte→string conversion (so a TEXT/BLOB column
// serializes as a string, not a base64 byte array in JSON) and the passthrough
// for every other type.
func TestNormalizeValue(t *testing.T) {
	t.Parallel()
	if got := normalizeValue([]byte("hi")); got != "hi" {
		t.Errorf("normalizeValue([]byte) = %v, want \"hi\"", got)
	}
	if got := normalizeValue(int64(7)); got != int64(7) {
		t.Errorf("normalizeValue(int64) = %v, want 7", got)
	}
	if got := normalizeValue(nil); got != nil {
		t.Errorf("normalizeValue(nil) = %v, want nil", got)
	}
}

// TestSchemeOf covers dsn scheme extraction: a scheme is the lowercased text
// before the first ':', but only when that ':' is not the first character (a
// leading ':', like ":memory:", has no scheme).
func TestSchemeOf(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"sqlite:./a.db":      "sqlite",
		"POSTGRES://x":       "postgres",
		"mysql://u@tcp(h)/d": "mysql",
		"plainpath":          "",
		":memory:":           "",
		"":                   "",
	}
	for dsn, want := range cases {
		if got := schemeOf(dsn); got != want {
			t.Errorf("schemeOf(%q) = %q, want %q", dsn, got, want)
		}
	}
}

// TestValidateDriver covers the loader-facing driver check: empty is valid
// (inferred later), every alias is accepted, and an unknown/typo'd name is
// rejected with a helpful message.
func TestValidateDriver(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"", "sqlite", "sqlite3", "postgres", "postgresql", "pgx", "mysql", "  MySQL  "} {
		if err := ValidateDriver(ok); err != nil {
			t.Errorf("ValidateDriver(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"sqllite", "mariadb", "oracle", "x"} {
		if err := ValidateDriver(bad); err == nil {
			t.Errorf("ValidateDriver(%q) = nil, want error", bad)
		}
	}
}

// TestDataSource covers the driver-specific DSN rewriting: sqlite strips a
// sqlite:/sqlite3: prefix (and preserves ":memory:" and bare paths), postgres is
// passed through verbatim, and a mysql:// URL is converted to the native form.
func TestDataSource(t *testing.T) {
	t.Parallel()
	cases := []struct {
		driver, dsn, want string
	}{
		{"sqlite", "sqlite:./a.db", "./a.db"},
		{"sqlite", "sqlite3:./a.db", "./a.db"},
		{"sqlite", ":memory:", ":memory:"},
		{"sqlite", "/abs/path.db", "/abs/path.db"},
		{"postgres", "postgres://u@h/d", "postgres://u@h/d"},
		{"mysql", "user:pass@tcp(127.0.0.1:3306)/db", "user:pass@tcp(127.0.0.1:3306)/db"},
	}
	for _, c := range cases {
		got, err := dataSource(c.driver, c.dsn)
		if err != nil {
			t.Errorf("dataSource(%q, %q) error = %v", c.driver, c.dsn, err)
			continue
		}
		if got != c.want {
			t.Errorf("dataSource(%q, %q) = %q, want %q", c.driver, c.dsn, got, c.want)
		}
	}
	// A mysql:// URL is rewritten to the native user:pass@tcp(host)/db form.
	got, err := dataSource("mysql", "mysql://u:p@h:3306/d")
	if err != nil {
		t.Fatalf("dataSource(mysql url) error = %v", err)
	}
	if got == "mysql://u:p@h:3306/d" {
		t.Errorf("dataSource did not rewrite mysql:// URL: %q", got)
	}
}

// TestSkipQuotedEdges exercises skipQuoted through isRowReturning for a doubled
// (escaped) quote and an unterminated quote — the classifier must not panic and
// must treat the quoted region as opaque data.
func TestSkipQuotedEdges(t *testing.T) {
	t.Parallel()
	// Doubled single quote is an escaped literal; the string spans to the real
	// close, so the leading verb is still INSERT (row-returning false).
	if isRowReturning("INSERT INTO t VALUES ('it''s RETURNING x')") {
		t.Error("doubled-quote literal containing RETURNING should not route as row-returning")
	}
	// An unterminated quote must not panic and consumes to end of input.
	if isRowReturning("INSERT INTO t VALUES ('unterminated") {
		t.Error("unterminated quote after INSERT should not be row-returning")
	}
	// A SELECT with an empty doubled-quote pair still classifies as row-returning.
	if !isRowReturning("SELECT '''' AS q") {
		t.Error("SELECT with doubled-quote literal should be row-returning")
	}
}
