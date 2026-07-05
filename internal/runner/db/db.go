// Package db implements the database runner: it runs a SQL statement from a
// `query:` step through a named db runner and captures the result — rows (for a
// SELECT) as a JSON array, or the affected-row count (for INSERT/UPDATE/DDL) —
// as a runner.Result. It is the atago counterpart to
// runn's DB runner. Only pure-Go drivers are linked in (no cgo): SQLite via
// modernc.org/sqlite, PostgreSQL via lib/pq, MySQL via go-sql-driver/mysql.
package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/runner"

	_ "github.com/go-sql-driver/mysql" // mysql driver (pure Go)
	_ "github.com/lib/pq"              // postgres driver (pure Go)
	_ "modernc.org/sqlite"             // sqlite driver (pure Go, no cgo)
)

// Config is the resolved configuration for a db runner.
type Config struct {
	// Driver is a database/sql driver name: sqlite, postgres, or mysql.
	Driver string
	// DataSource is the driver-specific DSN (path for sqlite, URL/native for the
	// servers).
	DataSource string
	// Timeout bounds a single query; zero means no timeout.
	Timeout time.Duration
}

// Resolve derives a Config from a runner's optional driver and its dsn. When
// driver is empty it is inferred from the dsn scheme (sqlite:, postgres://,
// mysql://); a native DSN with no recognizable scheme requires an explicit
// driver.
func Resolve(driver, dsn string) (Config, error) {
	if strings.TrimSpace(dsn) == "" {
		return Config{}, fmt.Errorf("db runner requires a dsn")
	}
	drv, err := resolveDriver(driver, dsn)
	if err != nil {
		return Config{}, err
	}
	ds, err := dataSource(drv, dsn)
	if err != nil {
		return Config{}, err
	}
	return Config{Driver: drv, DataSource: ds}, nil
}

// Runner holds an open database/sql pool for one db runner.
type Runner struct {
	db      *sql.DB
	timeout time.Duration
}

// Open opens (lazily — database/sql does not connect until first use) a pool for
// the configuration.
func Open(cfg Config) (*Runner, error) {
	db, err := sql.Open(cfg.Driver, cfg.DataSource)
	if err != nil {
		return nil, fmt.Errorf("opening %s database: %w", cfg.Driver, err)
	}
	// A scenario runs its queries sequentially against its own pool, so a single
	// connection is sufficient — and it is required for correctness with an
	// in-memory SQLite dsn, where each distinct connection would otherwise get its
	// own separate database (a table created on one would be invisible to the next).
	db.SetMaxOpenConns(1)
	return &Runner{db: db, timeout: cfg.Timeout}, nil
}

// Close releases the pool.
func (r *Runner) Close() error { return r.db.Close() }

// Query runs one SQL statement. A row-returning statement (SELECT, WITH, …) has
// its rows captured as a JSON array; any other statement is executed and its
// affected-row count recorded.
func (r *Runner) Query(ctx context.Context, query string) (*runner.Result, error) {
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}
	start := time.Now()

	if isRowReturning(query) {
		rows, err := r.db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer func() { _ = rows.Close() }()
		data, err := rowsToJSON(rows)
		if err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reading rows: %w", err)
		}
		return &runner.Result{Command: query, IsDB: true, RowsJSON: data, Duration: time.Since(start)}, nil
	}

	res, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %w", err)
	}
	affected, _ := res.RowsAffected() // not all drivers report it; best effort
	return &runner.Result{Command: query, IsDB: true, RowsJSON: []byte("[]"), RowsAffected: affected, Duration: time.Since(start)}, nil
}

// rowsToJSON scans every row into an ordered map and marshals the set as a JSON
// array. Byte-slice columns (text/blob) are rendered as strings rather than
// base64 so assertions read natural values.
func rowsToJSON(rows *sql.Rows) ([]byte, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("reading columns: %w", err)
	}
	out := make([]map[string]any, 0)
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			m[c] = normalizeValue(vals[i])
		}
		out = append(out, m)
	}
	return json.Marshal(out)
}

func normalizeValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

// isRowReturning reports whether a statement yields a result set and so should be
// run with QueryContext rather than ExecContext.
func isRowReturning(query string) bool {
	head := strings.ToUpper(strings.TrimSpace(stripSQLComments(query)))
	// A WITH (CTE) statement is only row-returning when its main statement is a
	// SELECT/VALUES (or it has RETURNING). A data-modifying CTE like
	// `WITH x AS (...) INSERT/UPDATE/DELETE ...` without RETURNING must run through
	// ExecContext so its affected-row count is captured (issue #31).
	if head == "WITH" || strings.HasPrefix(head, "WITH ") || strings.HasPrefix(head, "WITH\n") || strings.HasPrefix(head, "WITH\t") {
		return withIsRowReturning(head)
	}
	for _, kw := range []string{"SELECT", "PRAGMA", "SHOW", "EXPLAIN", "VALUES", "TABLE", "DESCRIBE", "DESC"} {
		if head == kw || strings.HasPrefix(head, kw+" ") || strings.HasPrefix(head, kw+"\n") || strings.HasPrefix(head, kw+"\t") {
			return true
		}
	}
	return hasReturningKeyword(head)
}

// stripSQLComments replaces SQL comments outside string/identifier literals with
// a single space, so the statement classifier sees the real leading verb. A `--`
// line comment runs to end of line; a `/* */` block comment to its close (an
// unterminated one to end of input). Quoted regions are skipped with the same
// rule as the keyword scans, so a `--` or `/*` inside a string literal stays as
// data. A commented `SELECT` was otherwise misrouted to ExecContext, losing its
// rows.
func stripSQLComments(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			j := skipQuoted(s, i)
			b.WriteString(s[i:j])
			i = j
		case c == '-' && i+1 < len(s) && s[i+1] == '-':
			j := i + 2
			for j < len(s) && s[j] != '\n' {
				j++
			}
			b.WriteByte(' ')
			i = j
		case c == '/' && i+1 < len(s) && s[i+1] == '*':
			j := i + 2
			for j+1 < len(s) && (s[j] != '*' || s[j+1] != '/') {
				j++
			}
			if j+1 < len(s) {
				j += 2 // consume the closing */
			} else {
				j = len(s) // unterminated block comment
			}
			b.WriteByte(' ')
			i = j
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

// hasReturningKeyword reports whether head contains a RETURNING keyword OUTSIDE
// any string literal or quoted identifier. A plain strings.Contains scan would
// also match a data value like 'order RETURNING to sender', misrouting an
// INSERT/UPDATE/DELETE to QueryContext and losing its affected-row count; this
// scan skips quoted regions the way withIsRowReturning does.
func hasReturningKeyword(head string) bool {
	for i := 0; i < len(head); {
		c := head[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			i = skipQuoted(head, i)
		case isWordChar(c):
			j := i
			for j < len(head) && isWordChar(head[j]) {
				j++
			}
			if head[i:j] == "RETURNING" {
				return true
			}
			i = j
		default:
			i++
		}
	}
	return false
}

// withIsRowReturning classifies an upper-cased WITH statement by the first SQL
// verb at parenthesis depth 0 — the main statement, since every CTE body is
// parenthesized. SELECT/VALUES returns rows; INSERT/UPDATE/DELETE/MERGE returns
// rows only with a RETURNING clause. Quoted strings/identifiers are skipped so
// their contents never trip the scan.
func withIsRowReturning(head string) bool {
	depth := 0
	for i := 0; i < len(head); {
		c := head[i]
		switch {
		case c == '(':
			depth++
			i++
		case c == ')':
			if depth > 0 {
				depth--
			}
			i++
		case c == '\'' || c == '"' || c == '`':
			i = skipQuoted(head, i)
		case depth == 0 && isWordChar(c):
			j := i
			for j < len(head) && isWordChar(head[j]) {
				j++
			}
			switch head[i:j] {
			case "SELECT", "VALUES":
				return true
			case "INSERT", "UPDATE", "DELETE", "MERGE":
				return hasReturningKeyword(head)
			}
			i = j
		default:
			i++
		}
	}
	// Couldn't find a main verb (e.g. an unusual dialect); default to row-returning
	// so a plain WITH ... SELECT is never misrouted.
	return true
}

// skipQuoted returns the index just past a quoted string/identifier beginning at
// s[i] (quote char s[i]), honoring SQL's doubled-quote escape (” / "" / “).
func skipQuoted(s string, i int) int {
	q := s[i]
	i++
	for i < len(s) {
		if s[i] == q {
			if i+1 < len(s) && s[i+1] == q { // doubled quote is an escaped literal
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return i
}

// isWordChar reports whether c can appear in a bare SQL keyword/identifier word.
func isWordChar(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

// resolveDriver decides the database/sql driver for a runner. An empty driver is
// inferred from the dsn scheme; a non-empty driver is authoritative, so an
// unsupported value is rejected as a configuration mistake (e.g. a typo like
// "sqllite") rather than silently falling back to inference.
func resolveDriver(driver, dsn string) (string, error) {
	if strings.TrimSpace(driver) == "" {
		drv := driverForScheme(schemeOf(dsn))
		if drv == "" {
			return "", fmt.Errorf("cannot infer db driver from dsn %q; set runner.driver to sqlite, postgres, or mysql", dsn)
		}
		return drv, nil
	}
	drv := canonicalDriver(driver)
	if drv == "" {
		return "", fmt.Errorf("unsupported runner.driver %q; use sqlite, postgres, or mysql (aliases: sqlite3, postgresql, pgx)", driver)
	}
	return drv, nil
}

// ValidateDriver reports whether an explicitly declared runner.driver is one of
// the supported names/aliases. It is meant for the loader, which wants to reject
// a bad driver before execution begins; an empty driver is always valid because
// it is inferred from the dsn at resolution time.
func ValidateDriver(driver string) error {
	if strings.TrimSpace(driver) == "" {
		return nil
	}
	if canonicalDriver(driver) == "" {
		return fmt.Errorf("unsupported runner.driver %q; use sqlite, postgres, or mysql (aliases: sqlite3, postgresql, pgx)", driver)
	}
	return nil
}

// canonicalDriver maps a user-supplied driver name to the linked driver, or ""
// when the name is empty or unknown.
func canonicalDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql", "pgx":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return ""
	}
}

func driverForScheme(scheme string) string {
	return canonicalDriver(scheme)
}

// schemeOf returns the lowercased text before the first ':' in a dsn, or "" when
// there is none. A native (schemeless) DSN therefore yields "", forcing the user
// to name the driver explicitly.
func schemeOf(dsn string) string {
	if i := strings.Index(dsn, ":"); i > 0 {
		return strings.ToLower(dsn[:i])
	}
	return ""
}

// dataSource converts a dsn into the form its driver expects.
func dataSource(driver, dsn string) (string, error) {
	switch driver {
	case "sqlite":
		// Accept "sqlite:<path>", "sqlite3:<path>", or a bare path. A path may be
		// ":memory:" (in-memory) — note the scheme prefix is only the leading token.
		for _, p := range []string{"sqlite3:", "sqlite:"} {
			if strings.HasPrefix(dsn, p) {
				return strings.TrimPrefix(dsn, p), nil
			}
		}
		return dsn, nil
	case "postgres":
		// lib/pq accepts both postgres:// and postgresql:// URLs verbatim, as well
		// as a native key=value DSN.
		return dsn, nil
	case "mysql":
		if strings.HasPrefix(dsn, "mysql://") {
			return mysqlNativeDSN(dsn)
		}
		return dsn, nil
	default:
		return dsn, nil
	}
}

// mysqlNativeDSN converts a mysql:// URL into the go-sql-driver native form
// "user:pass@tcp(host:port)/db?params".
func mysqlNativeDSN(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid mysql dsn %q: %w", raw, err)
	}
	var b strings.Builder
	if u.User != nil {
		b.WriteString(u.User.Username())
		if pass, ok := u.User.Password(); ok {
			b.WriteByte(':')
			b.WriteString(pass)
		}
		b.WriteByte('@')
	}
	fmt.Fprintf(&b, "tcp(%s)/%s", u.Host, strings.TrimPrefix(u.Path, "/"))
	if u.RawQuery != "" {
		b.WriteByte('?')
		b.WriteString(u.RawQuery)
	}
	return b.String(), nil
}
