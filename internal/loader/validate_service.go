package loader

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/spec"
)

func validateServices(add addFunc, where string, services []spec.Service) {
	seen := make(map[string]bool, len(services))
	for i := range services {
		svc := &services[i]
		sw := fmt.Sprintf("%s.services[%d]", where, i)
		if svc.Name == "" {
			add(diag.RequiredKey, "%s.name is required", sw)
		} else {
			if seen[svc.Name] {
				add(diag.DuplicateName, "%s: duplicate service name %q", where, svc.Name)
			}
			seen[svc.Name] = true
			sw = fmt.Sprintf("%s service %q", where, svc.Name)
		}
		if svc.Command == "" {
			add(diag.RequiredKey, "%s.command is required", sw)
		}
		workdirRelativeDir(add, sw+".cwd", svc.Cwd)
		validateHermeticEnv(add, sw, svc.ClearEnv, svc.PassEnv)
		if svc.MaxLogBytes < 0 {
			add(diag.NonPositiveValue, "%s.max_log_bytes must be positive (got %d); omit it for the 8 MiB default", sw, svc.MaxLogBytes)
		}
		validateReady(add, sw, svc.Ready)
	}
}

func validateReady(add addFunc, where string, r *spec.Ready) {
	if r == nil {
		return
	}
	n := 0
	for _, set := range []bool{r.File != "", r.Port != "", r.Log != "", r.Delay != ""} {
		if set {
			n++
		}
	}
	if n > 1 {
		add(diag.ExclusiveKeys, "%s.ready: set only one of file/port/log/delay", where)
	}
	if r.Store != "" && r.File == "" {
		add(diag.KeyNeedsAnother, "%s.ready.store requires file (the file whose content is captured)", where)
	}
	nonNegativeDuration(add, where+".ready.timeout", r.Timeout, "5s")
	nonNegativeDuration(add, where+".ready.delay", r.Delay, "500ms")
	if r.Log != "" {
		if _, err := regexp.Compile(r.Log); err != nil {
			add(diag.BadRegexp, "%s.ready.log %q is not a valid regexp: %v", where, r.Log, err)
		}
	}
}

// validateMockServers checks a scenario's mock_servers block (#24) and adds
// every declared name to mockNames (which arrives pre-seeded with the
// suite-wide mock names).
func validateMockServers(add addFunc, where string, servers []spec.MockServer, mockNames map[string]bool) {
	for i := range servers {
		ms := &servers[i]
		mw := fmt.Sprintf("%s.mock_servers[%d]", where, i)
		if ms.Name == "" {
			add(diag.RequiredKey, "%s.name is required", mw)
		} else {
			if mockNames[ms.Name] {
				add(diag.DuplicateName, "%s: duplicate mock server name %q", where, ms.Name)
			}
			mockNames[ms.Name] = true
			mw = fmt.Sprintf("%s mock server %q", where, ms.Name)
		}
		validateMockRoutes(add, mw, ms.Routes)
	}
}

// validateMockRoutes checks each canned route (#24): method+path required, at
// most one payload source, sane status, parseable delay.
func validateMockRoutes(add addFunc, where string, routes []spec.MockRoute) {
	// A route is reached by the FIRST declaration whose method and path match,
	// with the method compared case-insensitively — so a repeat of that pair is
	// dead configuration, and the spec claims a response it will never send.
	seen := map[string]bool{}
	for i := range routes {
		rt := &routes[i]
		rw := fmt.Sprintf("%s.routes[%d]", where, i)
		if rt.Method == "" {
			add(diag.RequiredKey, "%s.method is required", rw)
		}
		if rt.Path == "" {
			add(diag.RequiredKey, "%s.path is required", rw)
		} else if !strings.HasPrefix(rt.Path, "/") {
			add(diag.BadFormat, "%s.path %q must start with \"/\"", rw, rt.Path)
		} else if strings.Contains(rt.Path, "?") {
			// Matching compares the request's path, which the server has already
			// split from its query, so a declared query can never be part of a
			// match. Serve on the path and assert the query with the mock
			// target's own matchers instead.
			add(diag.BadFormat, "%s.path %q must not contain a query string; matching ignores the query, so this route can never answer", rw, rt.Path)
		}
		if rt.Method != "" && rt.Path != "" {
			key := strings.ToUpper(rt.Method) + " " + rt.Path
			if seen[key] {
				add(diag.DuplicateName, "%s: duplicate route %s; the first one declared always answers, so this one never does", rw, key)
			}
			seen[key] = true
		}
		payloads := 0
		if rt.JSON != nil {
			payloads++
		}
		if rt.Body != "" {
			payloads++
		}
		if rt.BodyFile != "" {
			payloads++
		}
		if payloads > 1 {
			add(diag.ExclusiveKeys, "%s: set at most one of json/body/body_file", rw)
		}
		if rt.Status != 0 && (rt.Status < 100 || rt.Status > 599) {
			add(diag.OutOfRange, "%s.status %d is not a valid HTTP status", rw, rt.Status)
		}
		nonNegativeDuration(add, rw+".delay", rt.Delay, "500ms")
	}
}
