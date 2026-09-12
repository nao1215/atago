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

// ViaRunner renders the " via <name>" suffix that names the runner carrying a
// step, or "" when the step names none. explain, doc, and the security notes
// share it so a runner is named the same way wherever it appears.
func ViaRunner(name string) string {
	if name == "" {
		return ""
	}
	return " via " + name
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
// `image` and `pdf` assertions (the inspected output the tool wrote), redirect
// targets (run's stdout_to/stderr_to, http's body_to), and `cdp` screenshot
// actions — across the scenario's steps and its teardown, which always runs
// and whose redirects land in the workdir like any other step's. explain, doc,
// and manifest all consume this so a generated output can never appear in one
// spec summary but silently vanish from another (#56).
func GeneratedArtifacts(sc *Scenario) []string {
	a := &artifactSet{seen: map[string]bool{}}
	WalkScenarioSteps(sc, func(_ StepPhase, st *Step) { a.step(st) })
	return a.out
}

// SuiteGeneratedArtifacts is GeneratedArtifacts for the suite's once-per-run
// lifecycle. A redirect in suite.setup or suite.teardown writes a file exactly
// as a scenario's does, and it appeared in no summary at all: there was no
// suite-level list anywhere, so the path was unrecoverable from explain, doc,
// or the manifest.
func SuiteGeneratedArtifacts(su *Suite) []string {
	a := &artifactSet{seen: map[string]bool{}}
	WalkSuiteSteps(su, func(_ StepPhase, st *Step) { a.step(st) })
	return a.out
}

// artifactSet accumulates declared output paths in order, without repeats. The
// scenario and suite walks share it so a new way of declaring an output cannot
// be learned by one and missed by the other.
type artifactSet struct {
	out  []string
	seen map[string]bool
}

func (a *artifactSet) add(p string) {
	if p != "" && !a.seen[p] {
		a.seen[p] = true
		a.out = append(a.out, p)
	}
}

// step folds one step's declared outputs into the set. The kinds absent from
// the switch declare none: a fixture writes an INPUT, and query/grpc/store/
// pty/signal/service/mock_server produce no file path of their own — a claim
// TestGeneratedArtifacts_DecidesEveryStepKind pins per kind, so a new kind
// cannot be skipped silently.
func (a *artifactSet) step(step *Step) {
	switch step.Kind() {
	case StepRun:
		a.add(step.Run.StdoutTo)
		a.add(step.Run.StderrTo)
	case StepHTTP:
		a.add(step.HTTP.BodyTo)
	case StepAssert:
		as := step.Assert
		if as.File != nil && as.File.Exists != nil && *as.File.Exists {
			a.add(as.File.Path)
		}
		if as.Image != nil {
			a.add(as.Image.Path)
		}
		// A pdf assertion inspects an output the tool wrote, exactly like
		// image; it arrived later (#73) and was the one inspecting target
		// this list never learned.
		if as.PDF != nil {
			a.add(as.PDF.Path)
		}
	case StepCDP:
		for _, act := range step.CDP.Actions {
			if act.Screenshot != nil {
				a.add(act.Screenshot.Path)
			}
		}
	case StepFixture, StepQuery, StepGRPC, StepStore, StepService, StepPTY, StepSignal, StepMockServer:
		// No generated artifact. A fixture writes an INPUT, and the rest produce
		// no file path of their own.
	}
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
	// Teardown steps always run — whether the scenario passed or failed — so an
	// egress that only happens during cleanup is still egress the summary must
	// name. The walk used to stop at Steps, and a scenario whose only network
	// access was a cleanup call reported no security notes at all; walking
	// through WalkScenarioSteps makes stopping early impossible to re-introduce.
	WalkScenarioSteps(sc, func(_ StepPhase, st *Step) { n.step(st, runners) })
	return n.out
}

// SuiteSecurityNotes returns, in declaration order and de-duplicated, the
// security-relevant operations the suite's once-per-run lifecycle performs
// across setup and teardown. Only scenarios had a security summary, so a suite
// whose setup curls a seed file through the shell and starts an ssh-tunnel
// service — and whose teardown curls a purge endpoint — reported no security
// notes anywhere, while the identical steps inside a scenario produce one for
// each. explain and manifest share this the way they share SecurityNotes.
// It takes the whole spec because the lifecycle is not only suite.setup and
// suite.teardown: a directory manifest's subject build (#393) runs on the host
// before any scenario, with the invoking environment and optionally through the
// shell, and was described by nothing.
func SuiteSecurityNotes(s *Spec) []string {
	n := &noteSet{seen: map[string]bool{}}
	if sub := s.Subject; sub != nil {
		where := "subject build " + sub.Name
		if sub.Shell {
			n.add("shell execution enabled (" + where + "): " + sub.Command)
		}
		if NetworkCommand.MatchString(sub.Command) {
			n.add("network access (" + where + "): " + sub.Command)
		}
		n.envRefs(sub.Command)
	}
	WalkSuiteSteps(&s.Suite, func(_ StepPhase, st *Step) { n.step(st, s.Runners) })
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

// envNames notes each env:-prefixed variable name in an already-collected var
// set, in sorted order for determinism. It exists for the kinds whose
// variable-bearing field list lives in a collector (cdp's action walk):
// re-listing the fields here to run envRefs over them is exactly the drift the
// shared walkers prevent.
func (n *noteSet) envNames(set map[string]bool) {
	for _, name := range SortedKeys(set) {
		if strings.HasPrefix(name, "env:") {
			n.add("host environment read: ${" + name + "}")
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
	// A runner's own fields are ${name}-expanded when a step uses it, so a
	// ${env:} in a base_url, dsn, target, or ssh credential reads the invoking
	// host's environment as surely as one written in the step. The walk covered
	// steps and services only, so a database password taken from the
	// environment produced no note at all.
	if name := StepRunner(step); name != "" {
		if r, ok := runners[name]; ok {
			n.envRefs(RunnerVarFields(r)...)
		}
	}
	switch step.Kind() {
	case StepRun:
		n.runStep(step.Run, runners)
	case StepHTTP:
		// Name the runner the way every other kind does: the note was the
		// constant "network access: HTTP request", so requests to two different
		// hosts de-duplicated into one anonymous line.
		n.add("network access: HTTP request" + ViaRunner(step.HTTP.Runner))
	case StepQuery:
		if remoteDatabase(runners[step.Query.Runner]) {
			n.add("network access: SQL query via " + step.Query.Runner)
		}
	case StepGRPC:
		n.add("network access: gRPC " + step.GRPC.Method)
	case StepCDP:
		n.add("browser automation (CDP) via " + step.CDP.Runner)
	case StepPTY:
		n.ptyStep(step.PTY)
	case StepService:
		// A `service:` step appears only in suite.setup (the loader rejects it
		// elsewhere); the process it starts deserves the same scrutiny as a
		// scenario-level service.
		n.service(step.Service)
	case StepFixture, StepAssert, StepStore, StepSignal, StepMockServer:
		// No note of their own: none of these reaches the network or starts a
		// process. Their host-environment reads are still reported, by the
		// kind-independent CollectStepVars walk below.
	}
	// Host-environment reads are one rule for every kind: any ${env:} in a
	// field the engine expands, found through the same walkers the engine
	// expands with (via CollectStepVars). Per-kind field lists here are how
	// fixture, assert, store, signal, and cdp env reads each went unreported
	// while the identical reference in a run command produced a note. This
	// comes after the kind notes so a summary's first line stays the operation,
	// not an input dependency; duplicates de-duplicate through n.seen.
	vars := map[string]bool{}
	CollectStepVars(vars, step, runners)
	n.envNames(vars)
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
	// The ${env:} reads across the command, env, and session are reported by
	// the unified walker scan in step(); only the shell/network judgments live
	// here.
	for _, a := range p.Session {
		if a.Exec == nil {
			continue
		}
		if a.Exec.ShellEnabled() {
			n.add("shell execution enabled (pty exec): " + a.Exec.Command)
		}
		if NetworkCommand.MatchString(a.Exec.Command) {
			n.add("network access (pty exec): " + a.Exec.Command)
		}
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
	// The ${env:} reads are reported by the unified walker scan in step().
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
