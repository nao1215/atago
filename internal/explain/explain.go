// Package explain renders a human- and LLM-readable summary of what a spec does
// without executing it: scenarios, commands, expected behavior,
// fixtures, generated files, variables, and security-sensitive operations.
package explain

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

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
		collectVars(vars, svc.Command, svc.Cwd)
		if svc.Ready != nil && svc.Ready.Store != "" {
			stores = append(stores, svc.Ready.Store)
		}
	}

	for i := range sc.Steps {
		step := &sc.Steps[i]
		switch step.Kind() {
		case spec.StepFixture:
			fixtures = append(fixtures, describeFixture(step.Fixture))
			collectVars(vars, step.Fixture.File, step.Fixture.Content)
		case spec.StepRun:
			commands = append(commands, describeRun(step.Run))
			collectVars(vars, step.Run.Command, step.Run.Cwd, step.Run.Stdin.Inline, step.Run.Stdin.File)
		case spec.StepAssert:
			expects = append(expects, describeAsserts(step.Assert)...)
		case spec.StepStore:
			if step.Store != nil {
				stores = append(stores, step.Store.Name)
			}
		case spec.StepHTTP:
			if step.HTTP != nil {
				commands = append(commands, fmt.Sprintf("HTTP %s %s", step.HTTP.Method, step.HTTP.Path))
				collectVars(vars, step.HTTP.Path)
				collectVars(vars, step.HTTP.Body)
			}
		case spec.StepQuery:
			if step.Query != nil {
				commands = append(commands, fmt.Sprintf("SQL query via %s: %s", step.Query.Runner, step.Query.SQL))
				collectVars(vars, step.Query.SQL)
			}
		case spec.StepGRPC:
			if step.GRPC != nil {
				commands = append(commands, fmt.Sprintf("gRPC %s via %s", step.GRPC.Method, step.GRPC.Runner))
				collectVars(vars, step.GRPC.Method)
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
				commands = append(commands, desc)
				collectVars(vars, step.PTY.Command, step.PTY.Cwd)
				for _, v := range step.PTY.Env {
					collectVars(vars, v)
				}
				for _, a := range step.PTY.Session {
					collectVars(vars, a.Expect)
					if a.Send != nil {
						collectVars(vars, *a.Send)
					}
				}
			}
		case spec.StepCDP:
			if step.CDP != nil {
				commands = append(commands, describeCDP(step.CDP))
				for _, a := range step.CDP.Actions {
					collectVars(vars, a.Navigate, a.WaitVisible, a.WaitHidden, a.Click, a.Check, a.Uncheck, a.Text, a.Eval)
					if a.SendKeys != nil {
						collectVars(vars, a.SendKeys.Selector, a.SendKeys.Value)
					}
					if a.Press != nil {
						collectVars(vars, a.Press.Selector, a.Press.Key)
					}
					if a.Select != nil {
						collectVars(vars, a.Select.Selector, a.Select.Value)
					}
					if a.Screenshot != nil {
						collectVars(vars, a.Screenshot.Path, a.Screenshot.Selector)
					}
					if a.Attribute != nil {
						collectVars(vars, a.Attribute.Selector, a.Attribute.Name)
					}
				}
			}
		case spec.StepSignal:
			if step.Signal != nil {
				commands = append(commands, describeSignal(step.Signal))
				collectVars(vars, step.Signal.Service)
			}
		}
	}

	// Teardown steps always run after the scenario — summarize them separately
	// so a reviewer sees what cleanup a spec performs against external systems.
	var teardown []string
	for i := range sc.Teardown {
		step := &sc.Teardown[i]
		switch step.Kind() {
		case spec.StepRun:
			teardown = append(teardown, describeRun(step.Run))
			collectVars(vars, step.Run.Command, step.Run.Cwd, step.Run.Stdin.Inline, step.Run.Stdin.File)
		case spec.StepHTTP:
			teardown = append(teardown, fmt.Sprintf("HTTP %s %s", step.HTTP.Method, step.HTTP.Path))
			collectVars(vars, step.HTTP.Path, step.HTTP.Body)
		case spec.StepQuery:
			teardown = append(teardown, fmt.Sprintf("SQL query via %s: %s", step.Query.Runner, step.Query.SQL))
			collectVars(vars, step.Query.SQL)
		case spec.StepGRPC:
			teardown = append(teardown, fmt.Sprintf("gRPC %s via %s", step.GRPC.Method, step.GRPC.Runner))
			collectVars(vars, step.GRPC.Method)
		case spec.StepCDP:
			teardown = append(teardown, describeCDP(step.CDP))
		case spec.StepFixture:
			teardown = append(teardown, describeFixture(step.Fixture))
		case spec.StepAssert:
			teardown = append(teardown, describeAsserts(step.Assert)...)
		case spec.StepStore:
			teardown = append(teardown, "store "+step.Store.Name)
		case spec.StepSignal:
			teardown = append(teardown, describeSignal(step.Signal))
			collectVars(vars, step.Signal.Service)
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
	if used := sortedKeys(vars); len(used) > 0 {
		fmt.Fprintf(b, "  Variables used: %s\n", strings.Join(used, ", "))
	}
	if security := spec.SecurityNotes(sc); len(security) > 0 {
		writeList(b, "⚠ Security notes", security)
	}
}

// describeService renders a one-line summary of a background service and how its
// readiness is decided (ADR-0031).
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

// describeCDP renders a one-line summary of a cdp step's action list (ADR-0029),
// reusing the shared per-action labels so explain stays aligned with doc and
// manifest (#50).
func describeCDP(c *spec.CDP) string {
	acts := make([]string, 0, len(c.Actions))
	for _, a := range c.Actions {
		acts = append(acts, spec.CDPActionLabel(a))
	}
	return fmt.Sprintf("CDP via %s: %s", c.Runner, strings.Join(acts, " → "))
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
		notes = append(notes, "env: "+strings.Join(sortedKeys(toSet(r.Env)), ", "))
	}
	if r.ClearEnvEnabled() {
		note := "cleared environment"
		if len(r.PassEnv) > 0 {
			note += " (passes: " + strings.Join(r.PassEnv, ", ") + ")"
		}
		notes = append(notes, note)
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
	switch {
	case h.Contains != nil:
		return fmt.Sprintf("%q contains %q", h.Name, *h.Contains)
	case h.Equals != nil:
		return fmt.Sprintf("%q equals %q", h.Name, *h.Equals)
	default:
		return fmt.Sprintf("%q", h.Name)
	}
}

func describeImage(im *spec.ImageAssert) string {
	var parts []string
	if im.Format != "" {
		parts = append(parts, "is "+im.Format)
	}
	if im.Width != nil {
		parts = append(parts, fmt.Sprintf("width %d", *im.Width))
	}
	if im.Height != nil {
		parts = append(parts, fmt.Sprintf("height %d", *im.Height))
	}
	if im.MinWidth != nil {
		parts = append(parts, fmt.Sprintf("width >= %d", *im.MinWidth))
	}
	if im.MaxWidth != nil {
		parts = append(parts, fmt.Sprintf("width <= %d", *im.MaxWidth))
	}
	if im.MinHeight != nil {
		parts = append(parts, fmt.Sprintf("height >= %d", *im.MinHeight))
	}
	if im.MaxHeight != nil {
		parts = append(parts, fmt.Sprintf("height <= %d", *im.MaxHeight))
	}
	if im.Alpha != nil {
		if *im.Alpha {
			parts = append(parts, "has alpha")
		} else {
			parts = append(parts, "has no alpha")
		}
	}
	if im.SimilarTo != "" {
		parts = append(parts, "similar to "+im.SimilarTo)
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%q is checked", im.Path)
	}
	return fmt.Sprintf("%q %s", im.Path, strings.Join(parts, ", "))
}

// describeDir renders a directory/tree assertion (#74) for explain output.
func describeDir(d *spec.DirAssert) string {
	var parts []string
	if d.Exists != nil {
		if *d.Exists {
			parts = append(parts, "exists")
		} else {
			parts = append(parts, "does not exist")
		}
	}
	for _, c := range d.Contains {
		parts = append(parts, "contains "+c)
	}
	for _, c := range d.NotContains {
		parts = append(parts, "does not contain "+c)
	}
	if d.Count != nil {
		parts = append(parts, fmt.Sprintf("has %d entries", *d.Count))
	}
	if d.MinCount != nil {
		parts = append(parts, fmt.Sprintf("has >= %d entries", *d.MinCount))
	}
	if d.MaxCount != nil {
		parts = append(parts, fmt.Sprintf("has <= %d entries", *d.MaxCount))
	}
	if d.Glob != "" {
		parts = append(parts, "matches glob "+d.Glob)
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%q is checked", d.Path)
	}
	return fmt.Sprintf("%q %s", d.Path, strings.Join(parts, ", "))
}

// describePDF renders a PDF assertion (#73) for explain output.
func describePDF(p *spec.PDFAssert) string {
	var parts []string
	if p.Pages != nil {
		parts = append(parts, fmt.Sprintf("%d pages", *p.Pages))
	}
	if p.MinPages != nil {
		parts = append(parts, fmt.Sprintf(">= %d pages", *p.MinPages))
	}
	if p.MaxPages != nil {
		parts = append(parts, fmt.Sprintf("<= %d pages", *p.MaxPages))
	}
	keys := make([]string, 0, len(p.Metadata))
	for k := range p.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s contains %q", k, p.Metadata[k]))
	}
	if p.Text != nil {
		parts = append(parts, "text "+describeStream(p.Text))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%q is checked", p.Path)
	}
	return fmt.Sprintf("%q %s", p.Path, strings.Join(parts, ", "))
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

func describeStream(s *spec.StreamAssert) string {
	switch {
	case s.Empty != nil:
		if *s.Empty {
			return "is empty"
		}
		return "is not empty"
	case s.Contains != nil:
		return "contains " + quoteList(s.Contains)
	case s.NotContains != nil:
		return "does not contain " + quoteList(s.NotContains)
	case s.Matches != nil:
		return fmt.Sprintf("matches /%s/", *s.Matches)
	case s.NotMatches != nil:
		return fmt.Sprintf("does not match /%s/", *s.NotMatches)
	case s.Equals != nil:
		return "equals exact text"
	case s.NotEquals != nil:
		return "does not equal exact text"
	case s.JSON != nil:
		return fmt.Sprintf("JSON %s %s", s.JSON.Path, jsonMatcher(s.JSON))
	case s.YAML != nil:
		return fmt.Sprintf("YAML %s %s", s.YAML.Path, jsonMatcher(s.YAML))
	case s.Snapshot != "":
		return "matches snapshot " + s.Snapshot
	default:
		return "(no matcher)"
	}
}

func describeFile(f *spec.FileAssert) string {
	switch {
	case f.Exists != nil:
		if *f.Exists {
			return fmt.Sprintf("%q exists", f.Path)
		}
		return fmt.Sprintf("%q does not exist", f.Path)
	case f.Contains != nil:
		return fmt.Sprintf("%q contains %s", f.Path, quoteList(f.Contains))
	case f.JSON != nil:
		return fmt.Sprintf("%q JSON %s %s", f.Path, f.JSON.Path, jsonMatcher(f.JSON))
	case f.Snapshot != "":
		return fmt.Sprintf("%q matches snapshot %s", f.Path, f.Snapshot)
	default:
		return f.Path
	}
}

func jsonMatcher(j *spec.JSONAssert) string {
	switch {
	case j.Equals != nil:
		return fmt.Sprintf("== %v", j.Equals)
	case j.Matches != nil:
		return fmt.Sprintf("matches /%s/", *j.Matches)
	case j.Length != nil:
		return fmt.Sprintf("length %d", *j.Length)
	default:
		return ""
	}
}

func collectVars(set map[string]bool, fields ...string) {
	for _, f := range fields {
		for _, name := range spec.VarRefs(f) {
			set[name] = true
		}
	}
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

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
