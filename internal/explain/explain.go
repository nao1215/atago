// Package explain renders a human- and LLM-readable summary of what a spec does
// without executing it: scenarios, commands, expected behavior,
// fixtures, generated files, variables, and security-sensitive operations.
package explain

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/nao1215/atago/internal/assertdesc"
	"github.com/nao1215/atago/internal/spec"
)

// Explain writes a summary of s to w.
func Explain(w io.Writer, s *spec.Spec, path string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "Spec: %s\n", path)
	fmt.Fprintf(&b, "Suite: %s\n", s.Suite.Name)
	if s.Suite.Timeout != "" {
		fmt.Fprintf(&b, "Default step timeout: %s (suite.timeout; a step or runner timeout overrides it)\n", s.Suite.Timeout)
	}
	if len(s.Secrets) > 0 {
		fmt.Fprintf(&b, "Secrets declared: %s\n", strings.Join(s.Secrets, ", "))
	}
	fmt.Fprintf(&b, "Network policy: %s\n", networkPolicy(s))
	explainSuiteBlock(&b, "Suite setup (runs once before any scenario)", s.Suite.Setup)
	explainSuiteBlock(&b, "Suite teardown (always runs after the last scenario)", s.Suite.Teardown)

	for i := range s.Scenarios {
		explainScenario(&b, &s.Scenarios[i])
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func networkPolicy(s *spec.Spec) string {
	if s.Permissions != nil && s.Permissions.Network != nil && len(s.Permissions.Network.Allow) > 0 {
		return "allow " + strings.Join(s.Permissions.Network.Allow, ", ")
	}
	// Runtime semantics: an empty allowlist means no policy is configured and
	// every host is permitted (security.CheckHost). Say so plainly instead of
	// implying a restrictive default that does not exist (issue #41).
	return "unrestricted (no allowlist set; all hosts permitted)"
}

// explainSuiteBlock summarizes suite.setup / suite.teardown (#7) so a reviewer
// sees the once-per-suite bootstrap (built helpers, suite-wide services,
// cleanup) without reading YAML.
func explainSuiteBlock(b *strings.Builder, label string, steps []spec.Step) {
	if len(steps) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", label)
	for i := range steps {
		step := &steps[i]
		switch step.Kind() {
		case spec.StepRun:
			fmt.Fprintf(b, "  - %s\n", describeRun(step.Run))
		case spec.StepService:
			fmt.Fprintf(b, "  - start suite service %q: %s\n", step.Service.Name, step.Service.Command)
		case spec.StepMockServer:
			fmt.Fprintf(b, "  - %s\n", describeMockServer(step.MockServer))
		case spec.StepFixture:
			fmt.Fprintf(b, "  - %s\n", describeFixture(step.Fixture))
		case spec.StepStore:
			fmt.Fprintf(b, "  - store %s\n", step.Store.Name)
		case spec.StepAssert:
			for _, d := range describeAsserts(step.Assert) {
				fmt.Fprintf(b, "  - expect %s\n", d)
			}
		}
	}
}

func explainScenario(b *strings.Builder, sc *spec.Scenario) {
	fmt.Fprintf(b, "\nScenario: %s", sc.Name)
	if len(sc.Tags) > 0 {
		fmt.Fprintf(b, "  [tags: %s]", strings.Join(sc.Tags, ", "))
	}
	if sc.Only != nil && sc.Only.OS != "" {
		fmt.Fprintf(b, "  [only os=%s]", sc.Only.OS)
	}
	if sc.Skip != nil && sc.Skip.OS != "" {
		fmt.Fprintf(b, "  [skip os=%s]", sc.Skip.OS)
	}
	b.WriteByte('\n')

	var fixtures, commands, expects, stores, services []string
	vars := map[string]bool{}

	for i := range sc.Services {
		svc := &sc.Services[i]
		services = append(services, describeService(svc))
		spec.CollectServiceVars(vars, svc)
		if svc.Ready != nil && svc.Ready.Store != "" {
			stores = append(stores, svc.Ready.Store)
		}
	}
	for i := range sc.MockServers {
		services = append(services, describeMockServer(&sc.MockServers[i]))
	}

	for i := range sc.Steps {
		step := &sc.Steps[i]
		// Variable references are collected by the shared spec walk so explain and
		// manifest never disagree about which ${name}s a step uses; the switch
		// below only formats the human-facing summary lines.
		spec.CollectStepVars(vars, step)
		switch step.Kind() {
		case spec.StepFixture:
			fixtures = append(fixtures, describeFixture(step.Fixture))
		case spec.StepRun:
			commands = append(commands, describeRun(step.Run))
		case spec.StepAssert:
			expects = append(expects, describeAsserts(step.Assert)...)
		case spec.StepStore:
			if step.Store != nil {
				stores = append(stores, step.Store.Name)
			}
		case spec.StepHTTP:
			if step.HTTP != nil {
				commands = append(commands, fmt.Sprintf("HTTP %s %s", step.HTTP.Method, step.HTTP.Path))
			}
		case spec.StepQuery:
			if step.Query != nil {
				commands = append(commands, fmt.Sprintf("SQL query via %s: %s", step.Query.Runner, step.Query.SQL))
			}
		case spec.StepGRPC:
			if step.GRPC != nil {
				commands = append(commands, fmt.Sprintf("gRPC %s via %s", step.GRPC.Method, step.GRPC.Runner))
			}
		case spec.StepPTY:
			if step.PTY != nil {
				desc := fmt.Sprintf("interactive (pty): %s  [%d session actions]", step.PTY.Command, len(step.PTY.Session))
				if step.PTY.ClearEnvEnabled() {
					note := "  (cleared environment"
					if len(step.PTY.PassEnv) > 0 {
						note += ", passes: " + strings.Join(step.PTY.PassEnv, ", ")
					}
					desc += note + ")"
				}
				if step.PTY.SandboxHomeEnabled() {
					desc += "  (isolated home)"
				}
				commands = append(commands, desc)
				var keys []string
				for _, a := range step.PTY.Session {
					if a.Send != nil && a.Send.Key != "" {
						keys = append(keys, a.Send.Key)
					}
				}
				if len(keys) > 0 {
					commands[len(commands)-1] += "  [keys: " + strings.Join(keys, ", ") + "]"
				}
			}
		case spec.StepCDP:
			if step.CDP != nil {
				commands = append(commands, spec.CDPActionSummary(step.CDP))
			}
		case spec.StepSignal:
			if step.Signal != nil {
				commands = append(commands, describeSignal(step.Signal))
			}
		}
	}

	// Teardown steps always run after the scenario — summarize them separately
	// so a reviewer sees what cleanup a spec performs against external systems.
	var teardown []string
	for i := range sc.Teardown {
		step := &sc.Teardown[i]
		spec.CollectStepVars(vars, step)
		switch step.Kind() {
		case spec.StepRun:
			teardown = append(teardown, describeRun(step.Run))
		case spec.StepHTTP:
			teardown = append(teardown, fmt.Sprintf("HTTP %s %s", step.HTTP.Method, step.HTTP.Path))
		case spec.StepQuery:
			teardown = append(teardown, fmt.Sprintf("SQL query via %s: %s", step.Query.Runner, step.Query.SQL))
		case spec.StepGRPC:
			teardown = append(teardown, fmt.Sprintf("gRPC %s via %s", step.GRPC.Method, step.GRPC.Runner))
		case spec.StepCDP:
			teardown = append(teardown, spec.CDPActionSummary(step.CDP))
		case spec.StepFixture:
			teardown = append(teardown, describeFixture(step.Fixture))
		case spec.StepAssert:
			teardown = append(teardown, describeAsserts(step.Assert)...)
		case spec.StepStore:
			teardown = append(teardown, "store "+step.Store.Name)
		case spec.StepSignal:
			teardown = append(teardown, describeSignal(step.Signal))
		}
	}

	// Generated artifacts and security notes come from the shared spec model, so
	// explain, doc, and manifest describe the same runtime surface (#56).
	writeList(b, "Services", services)
	writeList(b, "Fixtures", fixtures)
	writeList(b, "Commands", commands)
	writeList(b, "Expects", expects)
	writeList(b, "Teardown (always runs)", teardown)
	writeList(b, "Generates", spec.GeneratedArtifacts(sc))
	writeList(b, "Stores", stores)
	if used := spec.SortedKeys(vars); len(used) > 0 {
		fmt.Fprintf(b, "  Variables used: %s\n", strings.Join(used, ", "))
	}
	if security := spec.SecurityNotes(sc); len(security) > 0 {
		writeList(b, "⚠ Security notes", security)
	}
}

// describeService renders a one-line summary of a background service and how its
// readiness is decided.
func describeService(svc *spec.Service) string {
	desc := svc.Name + ": " + svc.Command
	if svc.ClearEnvEnabled() {
		note := "  [cleared environment"
		if len(svc.PassEnv) > 0 {
			note += ", passes: " + strings.Join(svc.PassEnv, ", ")
		}
		desc += note + "]"
	}
	if svc.Ready == nil {
		return desc
	}
	switch {
	case svc.Ready.File != "":
		desc += "  [ready when file " + svc.Ready.File + " appears]"
		if svc.Ready.Store != "" {
			desc += " → ${" + svc.Ready.Store + "}"
		}
	case svc.Ready.Port != "":
		desc += "  [ready when port " + svc.Ready.Port + " accepts]"
	case svc.Ready.Log != "":
		desc += "  [ready when log matches /" + svc.Ready.Log + "/]"
	case svc.Ready.Delay != "":
		desc += "  [ready after " + svc.Ready.Delay + "]"
	}
	return desc
}

// describeMockServer renders a one-line summary of a stub HTTP server (#24).
func describeMockServer(ms *spec.MockServer) string {
	return fmt.Sprintf("mock server %s: %d canned route(s), serves ${%s.url}", ms.Name, len(ms.Routes), ms.Name)
}

// describeSignal renders a one-line summary of a signal step (#23).
func describeSignal(sg *spec.Signal) string {
	desc := "send SIG" + spec.NormalizeSignalName(sg.Signal) + " to service " + sg.Service
	if sg.Wait != nil {
		timeout := sg.Wait.Timeout
		if timeout == "" {
			timeout = "5s"
		}
		desc += "  [wait up to " + timeout + " for exit]"
	}
	return desc
}

// describeChangesExplain renders a workdir-delta assertion (#70) as a compact
// phrase for `atago explain`. `modified: []` renders as "modified nothing".
func describeChangesExplain(c *spec.ChangesAssert) string {
	return assertdesc.DescribeChanges(c, explainChangesStyle)
}

// describeMockAssert renders a one-line summary of a mock assertion (#24).
func describeMockAssert(m *spec.MockAssert) string {
	return assertdesc.DescribeMock(m, explainMockStyle)
}

func describeFixture(f *spec.Fixture) string {
	kind := "inline content"
	if f.Base64 != "" {
		kind = "base64 binary"
	}
	return fmt.Sprintf("%s (%s)", f.File, kind)
}

func describeRun(r *spec.Run) string {
	var notes []string
	if r.Timeout != "" {
		notes = append(notes, "timeout "+r.Timeout)
	}
	if len(r.Env) > 0 {
		notes = append(notes, "env: "+strings.Join(spec.SortedKeys(toSet(r.Env)), ", "))
	}
	if r.ClearEnvEnabled() {
		note := "cleared environment"
		if len(r.PassEnv) > 0 {
			note += " (passes: " + strings.Join(r.PassEnv, ", ") + ")"
		}
		notes = append(notes, note)
	}
	if r.SandboxHomeEnabled() {
		notes = append(notes, "isolated home")
	}
	switch {
	case r.Stdin.File != "":
		notes = append(notes, "stdin from file "+r.Stdin.File)
	case r.Stdin.Base64 != "":
		notes = append(notes, "binary stdin (base64)")
	}
	if r.ShellEnabled() {
		notes = append(notes, "shell")
	}
	desc := r.Command
	if len(notes) > 0 {
		desc += "  (" + strings.Join(notes, ", ") + ")"
	}
	return desc
}

// describeAsserts renders one line per assertion target; an assert may set
// several (exit_code + stdout + …), each an independent expectation.
func describeAsserts(a *spec.Assert) []string {
	targets := a.SetTargets()
	if len(targets) == 0 {
		return []string{"(invalid assertion)"}
	}
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, describeTarget(a, t))
	}
	return out
}

func describeTarget(a *spec.Assert, target spec.AssertTarget) string {
	switch target {
	case spec.AssertExitCode:
		if a.ExitCode.Not != nil {
			return fmt.Sprintf("exit code is not %d", *a.ExitCode.Not)
		}
		if len(a.ExitCode.In) > 0 {
			return "exit code in " + intList(a.ExitCode.In)
		}
		if a.ExitCode.Equals != nil {
			return fmt.Sprintf("exit code is %d", *a.ExitCode.Equals)
		}
		return "exit code"
	case spec.AssertMock:
		return describeMockAssert(a.Mock)
	case spec.AssertScreen:
		return "screen " + describeStream(a.Screen)
	case spec.AssertDuration:
		return "completes " + a.Duration.DescribeDuration()
	case spec.AssertChanges:
		return "changed exactly " + describeChangesExplain(a.Changes)
	case spec.AssertStdout:
		return "stdout " + describeStream(a.Stdout)
	case spec.AssertStderr:
		return "stderr " + describeStream(a.Stderr)
	case spec.AssertFile:
		return "file " + describeFile(a.File)
	case spec.AssertImage:
		return "image " + describeImage(a.Image)
	case spec.AssertDir:
		return "dir " + describeDir(a.Dir)
	case spec.AssertPDF:
		return "pdf " + describePDF(a.PDF)
	case spec.AssertStatus:
		if a.Status != nil {
			return fmt.Sprintf("HTTP status is %d", *a.Status)
		}
		return "HTTP status"
	case spec.AssertHeader:
		if a.Header != nil {
			return "header " + describeHeader(a.Header)
		}
		return "header"
	case spec.AssertBody:
		return "body " + describeStream(a.Body)
	case spec.AssertRows:
		return "rows " + describeStream(a.Rows)
	case spec.AssertGRPCStatus:
		if a.GRPCStatus != nil {
			return fmt.Sprintf("gRPC status is %d", *a.GRPCStatus)
		}
		return "gRPC status"
	case spec.AssertMessage:
		return "message " + describeStream(a.Message)
	case spec.AssertValue:
		return "value " + describeStream(a.Value)
	default:
		return string(target)
	}
}

func describeHeader(h *spec.HeaderMatch) string {
	return assertdesc.DescribeHeader(h, explainHeaderStyle)
}

func describeImage(im *spec.ImageAssert) string {
	return assertdesc.DescribeImage(im, explainImageStyle)
}

// describeDir renders a directory/tree assertion (#74) for explain output.
func describeDir(d *spec.DirAssert) string {
	return assertdesc.DescribeDir(d, explainDirStyle)
}

// describePDF renders a PDF assertion (#73) for explain output.
func describePDF(p *spec.PDFAssert) string {
	return assertdesc.DescribePDF(p, explainPDFStyle)
}

// quoteList renders a contains/not_contains matcher argument. A single element
// is rendered as %q (byte-identical to the pre-list format); a list joins its
// quoted elements with ", ".
func quoteList(subs spec.StringList) string {
	parts := make([]string, len(subs))
	for i, s := range subs {
		parts[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(parts, ", ")
}

var explainJSONStyle = assertdesc.JSONStyle{
	Prefix:  func(path string) string { return "JSON " + path },
	Equals:  func(v any) string { return fmt.Sprintf("== %v", v) },
	Matches: func(s string) string { return fmt.Sprintf("matches /%s/", s) },
	Length:  func(n int) string { return fmt.Sprintf("length %d", n) },
	Compare: func(op string, v any) string { return fmt.Sprintf("%s %v", op, v) },
	Default: "",
}

var explainYAMLStyle = assertdesc.JSONStyle{
	Prefix:  func(path string) string { return "YAML " + path },
	Equals:  explainJSONStyle.Equals,
	Matches: explainJSONStyle.Matches,
	Length:  explainJSONStyle.Length,
	Compare: explainJSONStyle.Compare,
	Default: explainJSONStyle.Default,
}

var explainStreamStyle = assertdesc.StreamStyle{
	List:      quoteList,
	Regex:     func(s string) string { return fmt.Sprintf("/%s/", s) },
	Equals:    "equals exact text",
	NotEquals: "does not equal exact text",
	JSON:      explainJSONStyle,
	YAML:      explainYAMLStyle,
	Snapshot:  func(s string) string { return s },
	NoMatcher: "(no matcher)",
}

var explainFileStyle = assertdesc.FileStyle{
	Path:       func(s string) string { return fmt.Sprintf("%q", s) },
	List:       quoteList,
	JSON:       explainJSONStyle,
	Snapshot:   func(s string) string { return s },
	Checked:    func(path string) string { return path },
	ExactBytes: "equals exact bytes",
}

var explainHeaderStyle = assertdesc.HeaderStyle{
	Name:  func(s string) string { return fmt.Sprintf("%q", s) },
	Value: func(s string) string { return fmt.Sprintf("%q", s) },
	Regex: func(s string) string { return fmt.Sprintf("/%s/", s) },
	Bare:  func(s string) string { return fmt.Sprintf("%q", s) },
}

var explainImageStyle = assertdesc.ImageStyle{
	Path:      func(s string) string { return fmt.Sprintf("%q", s) },
	Format:    func(s string) string { return s },
	SimilarTo: func(s string) string { return s },
	Checked:   func(path string) string { return fmt.Sprintf("%q is checked", path) },
}

var explainDirStyle = assertdesc.DirStyle{
	Path:    func(s string) string { return fmt.Sprintf("%q", s) },
	Item:    func(s string) string { return s },
	Token:   func(s string) string { return s },
	Checked: func(path string) string { return fmt.Sprintf("%q is checked", path) },
}

var explainPDFStyle = assertdesc.PDFStyle{
	Path:    func(s string) string { return fmt.Sprintf("%q", s) },
	Value:   func(s string) string { return fmt.Sprintf("%q", s) },
	Stream:  describeStream,
	Checked: func(path string) string { return fmt.Sprintf("%q is checked", path) },
}

var explainChangesStyle = assertdesc.ChangesStyle{
	Entry: func(s string) string { return s },
	Join:  "; ",
}

var explainMockStyle = assertdesc.MockStyle{
	Name:  func(s string) string { return s },
	Route: func(s string) string { return s },
	Count: func(n int) string { return fmt.Sprintf(" x%d", n) },
}

func describeStream(s *spec.StreamAssert) string {
	return assertdesc.DescribeStream(s, explainStreamStyle)
}

func describeFile(f *spec.FileAssert) string {
	return assertdesc.DescribeFile(f, explainFileStyle)
}

func writeList(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "  %s:\n", title)
	for _, it := range items {
		fmt.Fprintf(b, "    - %s\n", it)
	}
}

// intList renders an accepted exit-code set as "[0, 2]" (#19).
func intList(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func toSet(m map[string]string) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}
