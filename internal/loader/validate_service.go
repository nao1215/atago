package loader

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

func validateServices(add func(string, ...any), where string, services []spec.Service) {
	seen := make(map[string]bool, len(services))
	for i := range services {
		svc := &services[i]
		sw := fmt.Sprintf("%s.services[%d]", where, i)
		if svc.Name == "" {
			add("%s.name is required", sw)
		} else {
			if seen[svc.Name] {
				add("%s: duplicate service name %q", where, svc.Name)
			}
			seen[svc.Name] = true
			sw = fmt.Sprintf("%s service %q", where, svc.Name)
		}
		if svc.Command == "" {
			add("%s.command is required", sw)
		}
		validateHermeticEnv(add, sw, svc.ClearEnv, svc.PassEnv)
		validateReady(add, sw, svc.Ready)
	}
}

func validateReady(add func(string, ...any), where string, r *spec.Ready) {
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
		add("%s.ready: set only one of file/port/log/delay", where)
	}
	if r.Store != "" && r.File == "" {
		add("%s.ready.store requires file (the file whose content is captured)", where)
	}
	for _, d := range []struct {
		key, val string
	}{{"timeout", r.Timeout}, {"delay", r.Delay}} {
		if d.val != "" {
			if _, err := time.ParseDuration(d.val); err != nil {
				add("%s.ready.%s %q is not a valid duration", where, d.key, d.val)
			}
		}
	}
	if r.Log != "" {
		if _, err := regexp.Compile(r.Log); err != nil {
			add("%s.ready.log %q is not a valid regexp: %v", where, r.Log, err)
		}
	}
}

// validateMockServers checks a scenario's mock_servers block (#24) and adds
// every declared name to mockNames (which arrives pre-seeded with the
// suite-wide mock names).
func validateMockServers(add func(string, ...any), where string, servers []spec.MockServer, mockNames map[string]bool) {
	for i := range servers {
		ms := &servers[i]
		mw := fmt.Sprintf("%s.mock_servers[%d]", where, i)
		if ms.Name == "" {
			add("%s.name is required", mw)
		} else {
			if mockNames[ms.Name] {
				add("%s: duplicate mock server name %q", where, ms.Name)
			}
			mockNames[ms.Name] = true
			mw = fmt.Sprintf("%s mock server %q", where, ms.Name)
		}
		validateMockRoutes(add, mw, ms.Routes)
	}
}

// validateMockRoutes checks each canned route (#24): method+path required, at
// most one payload source, sane status, parseable delay.
func validateMockRoutes(add func(string, ...any), where string, routes []spec.MockRoute) {
	for i := range routes {
		rt := &routes[i]
		rw := fmt.Sprintf("%s.routes[%d]", where, i)
		if rt.Method == "" {
			add("%s.method is required", rw)
		}
		if rt.Path == "" {
			add("%s.path is required", rw)
		} else if !strings.HasPrefix(rt.Path, "/") {
			add("%s.path %q must start with \"/\"", rw, rt.Path)
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
			add("%s: set at most one of json/body/body_file", rw)
		}
		if rt.Status != 0 && (rt.Status < 100 || rt.Status > 599) {
			add("%s.status %d is not a valid HTTP status", rw, rt.Status)
		}
		if rt.Delay != "" {
			if _, err := time.ParseDuration(rt.Delay); err != nil {
				add("%s.delay %q is not a valid duration (e.g. \"500ms\")", rw, rt.Delay)
			}
		}
	}
}
