package spec

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

// CDPActionSummary renders a cdp step's action list as a single line —
// "CDP via <runner>: <label → label → …>" — shared by explain and manifest so
// the two never disagree about how a browser step reads (#56).
func CDPActionSummary(c *CDP) string {
	acts := make([]string, 0, len(c.Actions))
	for _, a := range c.Actions {
		acts = append(acts, CDPActionLabel(a))
	}
	return "CDP via " + c.Runner + ": " + strings.Join(acts, " → ")
}

// SortedKeys returns the keys of a set in lexicographic order — the shared
// helper behind the sorted variable/label lists in explain and manifest.
func SortedKeys(m map[string]bool) []string {
	return slices.Sorted(maps.Keys(m))
}

// NetworkCommand matches shell commands that reach the network. explain, doc, and
// manifest share this single heuristic so their security notes never disagree
// about whether a command touches the network (#56).
var NetworkCommand = regexp.MustCompile(`(?i)\b(curl|wget|nc|ncat|ssh|scp|telnet)\b|https?://`)

// GeneratedArtifacts returns, in declaration order and de-duplicated, the file
// paths a scenario declares it produces: from `file` exists:true assertions,
// `image` assertions (the inspected output the tool wrote), redirect targets
// (run's stdout_to/stderr_to, http's body_to), and `cdp` screenshot actions.
// explain, doc, and manifest all consume this so a generated output can never
// appear in one spec summary but silently vanish from another (#56).
func GeneratedArtifacts(sc *Scenario) []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case StepRun:
			add(step.Run.StdoutTo)
			add(step.Run.StderrTo)
		case StepHTTP:
			add(step.HTTP.BodyTo)
		case StepAssert:
			a := step.Assert
			if a.File != nil && a.File.Exists != nil && *a.File.Exists {
				add(a.File.Path)
			}
			if a.Image != nil {
				add(a.Image.Path)
			}
		case StepCDP:
			for _, act := range step.CDP.Actions {
				if act.Screenshot != nil {
					add(act.Screenshot.Path)
				}
			}
		}
	}
	return out
}

// SecurityNotes returns, in declaration order and de-duplicated, the
// security-relevant operations a scenario performs: shell execution, network
// access (via run/http/grpc), and browser automation. explain and manifest share
// this so their machine- and human-facing security summaries stay identical
// (#56).
func SecurityNotes(sc *Scenario) []string {
	var out []string
	seen := map[string]bool{}
	add := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	// A ${env:NAME} reference reads the invoking host environment — an input
	// dependency worth surfacing for review alongside shell/network use.
	addEnvRefs := func(fields ...string) {
		for _, f := range fields {
			for _, name := range VarRefs(f) {
				if strings.HasPrefix(name, "env:") {
					add("host environment read: ${" + name + "}")
				}
			}
		}
	}
	// Values in sorted-key order: explain/doc/manifest output must stay
	// deterministic, and Go map iteration is not.
	envValues := func(m map[string]string) []string {
		vals := make([]string, 0, len(m))
		for _, k := range slices.Sorted(maps.Keys(m)) {
			vals = append(vals, m[k])
		}
		return vals
	}
	for i := range sc.Services {
		svc := &sc.Services[i]
		if svc.ShellEnabled() {
			add("shell execution enabled (service " + svc.Name + "): " + svc.Command)
		}
		if NetworkCommand.MatchString(svc.Command) {
			add("network access (service " + svc.Name + "): " + svc.Command)
		}
		addEnvRefs(svc.Command)
		addEnvRefs(envValues(svc.Env)...)
	}
	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case StepRun:
			if step.Run.ShellEnabled() {
				add("shell execution enabled: " + step.Run.Command)
			}
			if NetworkCommand.MatchString(step.Run.Command) {
				add("network access: " + step.Run.Command)
			}
			addEnvRefs(step.Run.Command, step.Run.Stdin.Inline, step.Run.Stdin.File)
			addEnvRefs(envValues(step.Run.Env)...)
		case StepHTTP:
			add("network access: HTTP request")
			addEnvRefs(step.HTTP.Path, step.HTTP.Body)
			addEnvRefs(envValues(step.HTTP.Header)...)
		case StepQuery:
			addEnvRefs(step.Query.SQL)
		case StepGRPC:
			add("network access: gRPC " + step.GRPC.Method)
			addEnvRefs(envValues(step.GRPC.Header)...)
		case StepCDP:
			add("browser automation (CDP) via " + step.CDP.Runner)
		}
	}
	return out
}
