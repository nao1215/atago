package spec

import (
	"maps"
	"regexp"
	"slices"
	"strconv"
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

// RunHost names where a run step executes when that is not the machine running
// atago: "ssh <runner>" for a command an ssh runner sends elsewhere, and "" for
// anything local.
//
// explain, doc, and manifest share the decision because all three rendered a
// remote command as a bare command line — `uptime` reads as something that ran
// here, and every other step kind already says which runner carried it. Only
// the decision is shared; each surface keeps its own phrasing.
func RunHost(r *Run, runners map[string]Runner) string {
	if r == nil || runners[r.Runner].Type != "ssh" {
		return ""
	}
	return "ssh " + r.Runner
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
// access, and browser automation. explain and manifest share this so their
// machine- and human-facing security summaries stay identical (#56).
//
// runners is the spec's runner table, because two of the five ways a scenario
// reaches the network are properties of the runner rather than of the step: a
// `run:` through an ssh runner logs into another machine, and a `query:`
// through a db runner pointed at a server dials it. Both were silent here while
// http and grpc were named, so the summary said less about a spec's egress than
// permissions.network already enforced.
func SecurityNotes(sc *Scenario, runners map[string]Runner) []string {
	n := &noteSet{seen: map[string]bool{}}
	for i := range sc.Services {
		n.service(&sc.Services[i])
	}
	for i := range sc.Steps {
		n.step(&sc.Steps[i], runners)
	}
	// Teardown steps always run — whether the scenario passed or failed — so an
	// egress that only happens during cleanup is still egress the summary must
	// name. The walk used to stop at Steps, and a scenario whose only network
	// access was a cleanup call reported no security notes at all.
	for i := range sc.Teardown {
		n.step(&sc.Teardown[i], runners)
	}
	return n.out
}

// SuiteSecurityNotes returns, in declaration order and de-duplicated, the
// security-relevant operations the suite's once-per-run lifecycle performs
// across setup and teardown. Only scenarios had a security summary, so a suite
// whose setup curls a seed file through the shell and starts an ssh-tunnel
// service — and whose teardown curls a purge endpoint — reported no security
// notes anywhere, while the identical steps inside a scenario produce one for
// each. explain and manifest share this the way they share SecurityNotes.
func SuiteSecurityNotes(su *Suite, runners map[string]Runner) []string {
	n := &noteSet{seen: map[string]bool{}}
	for i := range su.Setup {
		n.step(&su.Setup[i], runners)
	}
	for i := range su.Teardown {
		n.step(&su.Teardown[i], runners)
	}
	return n.out
}

// noteSet accumulates the security notes in order, without repeats: two steps
// enabling the shell on the same command say it once.
type noteSet struct {
	out  []string
	seen map[string]bool
}

func (n *noteSet) add(s string) {
	if s != "" && !n.seen[s] {
		n.seen[s] = true
		n.out = append(n.out, s)
	}
}

// envRefs notes each ${env:NAME} reference in the given fields. Reading the
// invoking host's environment is an input dependency worth surfacing for review
// alongside shell and network use.
func (n *noteSet) envRefs(fields ...string) {
	for _, f := range fields {
		for _, name := range VarRefs(f) {
			if strings.HasPrefix(name, "env:") {
				n.add("host environment read: ${" + name + "}")
			}
		}
	}
}

// envValues returns a map's values in sorted-key order: explain, doc, and
// manifest output must stay deterministic, and Go map iteration is not.
func envValues(m map[string]string) []string {
	vals := make([]string, 0, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		vals = append(vals, m[k])
	}
	return vals
}

func (n *noteSet) service(svc *Service) {
	if svc.ShellEnabled() {
		n.add("shell execution enabled (service " + svc.Name + "): " + svc.Command)
	}
	if NetworkCommand.MatchString(svc.Command) {
		n.add("network access (service " + svc.Name + "): " + svc.Command)
	}
	n.envRefs(svc.Command)
	n.envRefs(envValues(svc.Env)...)
}

func (n *noteSet) step(step *Step, runners map[string]Runner) {
	switch step.Kind() {
	case StepRun:
		n.runStep(step.Run, runners)
	case StepHTTP:
		n.add("network access: HTTP request")
		n.envRefs(step.HTTP.Path, step.HTTP.Body)
		n.envRefs(envValues(step.HTTP.Header)...)
	case StepQuery:
		if remoteDatabase(runners[step.Query.Runner]) {
			n.add("network access: SQL query via " + step.Query.Runner)
		}
		n.envRefs(step.Query.SQL)
	case StepGRPC:
		n.add("network access: gRPC " + step.GRPC.Method)
		n.envRefs(envValues(step.GRPC.Header)...)
	case StepCDP:
		n.add("browser automation (CDP) via " + step.CDP.Runner)
	case StepPTY:
		n.ptyStep(step.PTY)
	case StepService:
		// A `service:` step appears only in suite.setup (the loader rejects it
		// elsewhere); the process it starts deserves the same scrutiny as a
		// scenario-level service.
		n.service(step.Service)
	}
}

// ptyStep mirrors runStep for the interactive form: the pty child is a process
// like any run step's, and the session's exec actions run commands on the HOST
// mid-session — so both get the same shell/network/environment scrutiny. The
// step used to be invisible here entirely, while the equivalent run step
// produced a note for each.
func (n *noteSet) ptyStep(p *PTY) {
	if p.ShellEnabled() {
		n.add("shell execution enabled: " + p.Command)
	}
	if NetworkCommand.MatchString(p.Command) {
		n.add("network access: " + p.Command)
	}
	n.envRefs(p.Command)
	n.envRefs(envValues(p.Env)...)
	for _, a := range p.Session {
		if a.Send != nil {
			if a.Send.Text != nil {
				n.envRefs(*a.Send.Text)
			}
			if a.Send.Paste != nil {
				n.envRefs(*a.Send.Paste)
			}
		}
		if a.Exec == nil {
			continue
		}
		if a.Exec.ShellEnabled() {
			n.add("shell execution enabled (pty exec): " + a.Exec.Command)
		}
		if NetworkCommand.MatchString(a.Exec.Command) {
			n.add("network access (pty exec): " + a.Exec.Command)
		}
		n.envRefs(a.Exec.Command)
	}
}

func (n *noteSet) runStep(r *Run, runners map[string]Runner) {
	if r.ShellEnabled() {
		n.add("shell execution enabled: " + r.Command)
	}
	if NetworkCommand.MatchString(r.Command) {
		n.add("network access: " + r.Command)
	}
	// An ssh runner does not need the command to look networky: the step runs
	// on another host by construction, which a reader of a bare "uptime" line
	// has no way to tell.
	if rdef := runners[r.Runner]; rdef.Type == "ssh" {
		n.add("network access (ssh " + r.Runner + "): " + r.Command)
		// The loader refuses an ssh runner that decides neither way about
		// host-key verification, which is how much the choice matters. Having
		// forced the decision, the summary has to report which way it went —
		// accepting whatever key the host presents is the point of the note,
		// so a runner that does verify says nothing extra.
		if rdef.InsecureHostKey {
			n.add("ssh host key verification disabled (runner " + strconv.Quote(r.Runner) + ")")
		}
	}
	n.envRefs(r.Command, r.Stdin.Inline, r.Stdin.File)
	n.envRefs(envValues(r.Env)...)
}

// remoteDatabase reports whether a db runner dials a server rather than opening
// a local file.
//
// The test is deliberately coarse — sqlite is the one bundled driver that names
// a path instead of a peer, so everything else is treated as remote. It answers
// "should a reviewer be told this scenario talks to a database over the
// network", where naming one that turns out to be local is a wasted line and
// staying silent about a real one is the failure. The exact peer, which the
// network allowlist needs, is resolved from the dsn by the db runner itself.
func remoteDatabase(r Runner) bool {
	if r.Type != "db" {
		return false
	}
	if r.Driver != "" {
		return r.Driver != "sqlite" && r.Driver != "sqlite3"
	}
	scheme, _, ok := strings.Cut(r.DSN, ":")
	if !ok {
		// A dsn with no scheme is a bare sqlite path (the loader requires an
		// explicit driver for anything else).
		return false
	}
	switch strings.ToLower(scheme) {
	case "sqlite", "sqlite3":
		return false
	default:
		return true
	}
}
