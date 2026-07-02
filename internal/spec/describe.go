package spec

import "regexp"

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
	for i := range sc.Services {
		svc := &sc.Services[i]
		if svc.Shell {
			add("shell execution enabled (service " + svc.Name + "): " + svc.Command)
		}
		if NetworkCommand.MatchString(svc.Command) {
			add("network access (service " + svc.Name + "): " + svc.Command)
		}
	}
	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case StepRun:
			if step.Run.Shell {
				add("shell execution enabled: " + step.Run.Command)
			}
			if NetworkCommand.MatchString(step.Run.Command) {
				add("network access: " + step.Run.Command)
			}
		case StepHTTP:
			add("network access: HTTP request")
		case StepGRPC:
			add("network access: gRPC " + step.GRPC.Method)
		case StepCDP:
			add("browser automation (CDP) via " + step.CDP.Runner)
		}
	}
	return out
}
