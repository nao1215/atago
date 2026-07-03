package manifest

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/spec"
)

// buildScenario summarizes one (already matrix-expanded) scenario. Variable
// references, generated artifacts, and security notes are collected across the
// scenario's services and steps so tooling can see them without replaying the run.
func buildScenario(sc *spec.Scenario, src SourceLocator) Scenario {
	out := Scenario{
		Name: sc.Name,
		Tags: append([]string(nil), sc.Tags...),
		Vars: copyMap(sc.Vars),
		Only: condition(sc.Only),
		Skip: condition(sc.Skip),
	}
	if src != nil {
		out.Source = sourceFrom(src.ScenarioPos(sc.SourceIndex))
	}

	vars := map[string]bool{}

	for i := range sc.Services {
		svc := &sc.Services[i]
		out.Services = append(out.Services, buildService(svc))
		collectVars(vars, svc.Command, svc.Cwd)
		for _, v := range svc.Env {
			collectVars(vars, v)
		}
	}

	for i := range sc.Steps {
		step := &sc.Steps[i]
		st := buildStep(i, step, vars)
		if src != nil {
			st.Source = sourceFrom(src.StepPos(sc.SourceIndex, i))
		}
		out.Steps = append(out.Steps, st)
	}

	// Teardown steps share the step shape; their variable references count
	// toward the scenario's referenced-variable set like any other step's.
	for i := range sc.Teardown {
		out.Teardown = append(out.Teardown, buildStep(i, &sc.Teardown[i], vars))
	}

	out.Variables = sortedKeys(vars)
	// Generated artifacts and security notes come from the shared spec model, so
	// the manifest and the human-facing explain/doc summaries never drift (#56).
	out.Generates = spec.GeneratedArtifacts(sc)
	out.Security = spec.SecurityNotes(sc)
	return out
}

func buildService(svc *spec.Service) Service {
	out := Service{Name: svc.Name, Command: svc.Command, Shell: svc.ShellEnabled()}
	if svc.Ready != nil {
		switch {
		case svc.Ready.File != "":
			out.Ready = "file"
		case svc.Ready.Port != "":
			out.Ready = "port"
		case svc.Ready.Log != "":
			out.Ready = "log"
		case svc.Ready.Delay != "":
			out.Ready = "delay"
		}
		out.Store = svc.Ready.Store
	}
	return out
}

// buildStep reduces one step to its declarative fields and folds its variable
// references into the scenario-level var set. Generated artifacts and security
// notes are derived separately from the shared spec model (#56).
func buildStep(index int, step *spec.Step, vars map[string]bool) Step {
	st := Step{Index: index, Kind: string(step.Kind())}
	switch step.Kind() {
	case spec.StepFixture:
		st.File = step.Fixture.File
		st.Action = "write fixture " + step.Fixture.File
		collectVars(vars, step.Fixture.File, step.Fixture.Content, step.Fixture.Symlink)

	case spec.StepService:
		svc := step.Service
		st.Command = svc.Command
		st.Shell = svc.ShellEnabled()
		st.Target = svc.Name
		st.Action = "start suite service " + svc.Name
		collectVars(vars, svc.Command, svc.Cwd)

	case spec.StepRun:
		r := step.Run
		st.Command = r.Command
		st.Shell = r.ShellEnabled()
		st.Runner = r.Runner
		st.Action = "run " + r.Command
		if r.Retry != nil {
			st.Retry = &Retry{Times: r.Retry.Times, Interval: r.Retry.Interval}
		}
		collectVars(vars, r.Command, r.Cwd, r.Stdin)
		for _, v := range r.Env {
			collectVars(vars, v)
		}

	case spec.StepHTTP:
		h := step.HTTP
		st.Method = h.Method
		st.Path = h.Path
		st.Runner = h.Runner
		st.Action = fmt.Sprintf("HTTP %s %s", h.Method, h.Path)
		if h.Retry != nil {
			st.Retry = &Retry{Times: h.Retry.Times, Interval: h.Retry.Interval}
		}
		collectVars(vars, h.Path)
		collectVars(vars, h.Body)
		collectVars(vars, h.BodyFile, h.BodyTo)
		for _, v := range h.Form {
			collectVars(vars, v)
		}
		for _, f := range h.Files {
			collectVars(vars, f.Path)
		}

	case spec.StepQuery:
		q := step.Query
		st.SQL = q.SQL
		st.Runner = q.Runner
		st.Action = "SQL query via " + q.Runner
		collectVars(vars, q.SQL)

	case spec.StepGRPC:
		g := step.GRPC
		st.Method = g.Method
		st.Runner = g.Runner
		st.Action = "gRPC " + g.Method + " via " + g.Runner
		collectVars(vars, g.Method)

	case spec.StepCDP:
		c := step.CDP
		st.Runner = c.Runner
		st.Action = describeCDPActions(c)
		for _, a := range c.Actions {
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
			if a.Upload != nil {
				collectVars(vars, a.Upload.Selector, a.Upload.File)
			}
			if a.Download != nil {
				collectVars(vars, a.Download.Click, a.Download.Dir)
			}
		}

	case spec.StepPTY:
		pt := step.PTY
		st.Command = pt.Command
		st.Shell = pt.Shell != nil && *pt.Shell
		st.Action = "interactive (pty) " + pt.Command
		collectVars(vars, pt.Command, pt.Cwd)
		for _, a := range pt.Session {
			if a.Send != nil {
				collectVars(vars, *a.Send)
			}
			collectVars(vars, a.Expect)
		}

	case spec.StepAssert:
		st.Target = assertTarget(step.Assert)
		st.Action = "assert " + st.Target

	case spec.StepStore:
		if step.Store != nil {
			st.Target = step.Store.Name
			st.Action = "store " + step.Store.Name
		}
	}
	return st
}

// assertTarget returns the assertion target name (stdout, file, status …). When
// an assert sets several targets (exit_code + stdout + …), it joins them with
// "+" so the manifest names every target the step checks.
func assertTarget(a *spec.Assert) string {
	targets := a.SetTargets()
	if len(targets) == 0 {
		return "invalid"
	}
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = string(t)
	}
	return strings.Join(names, "+")
}

func describeCDPActions(c *spec.CDP) string {
	acts := make([]string, 0, len(c.Actions))
	for _, a := range c.Actions {
		acts = append(acts, spec.CDPActionLabel(a))
	}
	return "CDP via " + c.Runner + ": " + strings.Join(acts, " → ")
}

func condition(c *spec.Condition) *Condition {
	if c == nil || (c.OS == "" && c.Env == "" && c.Command == "") {
		return nil
	}
	return &Condition{OS: c.OS, Env: c.Env, Command: c.Command}
}

func collectVars(set map[string]bool, fields ...string) {
	for _, f := range fields {
		for _, name := range spec.VarRefs(f) {
			set[name] = true
		}
	}
}

func copyMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
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
