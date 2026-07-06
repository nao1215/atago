package loader

import (
	"fmt"
	"time"

	"github.com/nao1215/atago/internal/runner/db"
	"github.com/nao1215/atago/internal/spec"
)

var validRunnerType = map[string]bool{"cmd": true, "http": true, "db": true, "ssh": true, "grpc": true, "browser": true}

func validateRunners(add func(string, ...any), runners map[string]spec.Runner) {
	for name, r := range runners {
		where := fmt.Sprintf("runner %q", name)
		if r.Type == "" {
			add("%s.type is required", where)
			continue
		}
		if !validRunnerType[r.Type] {
			add("%s.type %q is invalid (want cmd, http, db, ssh, grpc, or browser)", where, r.Type)
			continue
		}
		switch r.Type {
		case "db":
			if r.DSN == "" {
				add("%s (db) requires a dsn", where)
			}
			// A declared driver is authoritative: reject an unsupported value here so
			// a typo fails at load time instead of silently inferring from the dsn.
			if err := db.ValidateDriver(r.Driver); err != nil {
				add("%s: %v", where, err)
			}
		case "ssh":
			if r.Host == "" {
				add("%s (ssh) requires a host", where)
			}
			if r.User == "" {
				add("%s (ssh) requires a user", where)
			}
		case "grpc":
			if r.Target == "" {
				add("%s (grpc) requires a target", where)
			}
		case "browser":
			// no required fields; a browser runner launches a local headless Chrome.
		}
		// timeout is common to every runner type; catch a malformed value here
		// instead of when the first step opens the connection.
		if r.Timeout != "" {
			if d, err := time.ParseDuration(r.Timeout); err != nil {
				add("%s.timeout %q is not a valid duration (e.g. \"30s\")", where, r.Timeout)
			} else if d < 0 {
				add("%s.timeout must not be negative (got %q); a wall-clock bound is never below zero", where, r.Timeout)
			}
		}
		validateRunnerFields(add, where, &r)
	}
}

// runnerFields maps each runner field to the single runner type that owns it, so
// cross-type fields (an http runner with ssh fields, a grpc runner with db
// fields, ...) are rejected instead of silently accepted (#44). type/cwd/timeout
// are common to every runner and intentionally absent here.
func validateRunnerFields(add func(string, ...any), where string, r *spec.Runner) {
	type fieldOwner struct {
		owner string
		set   bool
		field string
	}
	fields := []fieldOwner{
		{"http", r.BaseURL != "", "base_url"},
		{"db", r.DSN != "", "dsn"},
		{"db", r.Driver != "", "driver"},
		{"ssh", r.Host != "", "host"},
		{"ssh", r.User != "", "user"},
		{"ssh", r.Password != "", "password"},
		{"ssh", r.KeyFile != "", "key_file"},
		{"ssh", r.KnownHosts != "", "known_hosts"},
		{"ssh", r.InsecureHostKey, "insecure_host_key"},
		{"grpc", r.Target != "", "target"},
		{"grpc", r.TLS, "tls"},
		{"browser", r.Headless != nil, "headless"},
		{"browser", r.ExecPath != "", "exec_path"},
		{"browser", len(r.BrowserArgs) > 0, "browser_args"},
	}
	for _, f := range fields {
		if f.set && f.owner != r.Type {
			add("%s: %q cannot be set on a %s runner (it is a %s-runner field)", where, f.field, r.Type, f.owner)
		}
	}
}
