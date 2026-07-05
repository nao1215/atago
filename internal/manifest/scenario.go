package manifest

import (
	"fmt"
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
		spec.CollectVars(vars, svc.Command, svc.Cwd)
		for _, v := range svc.Env {
			spec.CollectVars(vars, v)
		}
	}
	for i := range sc.MockServers {
		out.MockServers = append(out.MockServers, MockServer{Name: sc.MockServers[i].Name, Routes: len(sc.MockServers[i].Routes)})
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

	out.Variables = spec.SortedKeys(vars)
	// Generated artifacts and security notes come from the shared spec model, so
	// the manifest and the human-facing explain/doc summaries never drift (#56).
	out.Generates = spec.GeneratedArtifacts(sc)
	out.Security = spec.SecurityNotes(sc)
	return out
}

func buildService(svc *spec.Service) Service {
	out := Service{Name: svc.Name, Command: svc.Command, Shell: svc.ShellEnabled(), ClearEnv: svc.ClearEnvEnabled(), PassEnv: svc.PassEnv}
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
		spec.CollectVars(vars, step.Fixture.File, step.Fixture.Content, step.Fixture.Symlink)

	case spec.StepService:
		svc := step.Service
		st.Command = svc.Command
		st.Shell = svc.ShellEnabled()
		st.ClearEnv = svc.ClearEnvEnabled()
		st.PassEnv = svc.PassEnv
		st.Target = svc.Name
		st.Action = "start suite service " + svc.Name
		spec.CollectVars(vars, svc.Command, svc.Cwd)

	case spec.StepRun:
		r := step.Run
		st.Command = r.Command
		st.Shell = r.ShellEnabled()
		st.ClearEnv = r.ClearEnvEnabled()
		st.PassEnv = r.PassEnv
		st.Runner = r.Runner
		st.Action = "run " + r.Command
		if r.Retry != nil {
			st.Retry = &Retry{Times: r.Retry.Times, Interval: r.Retry.Interval}
		}
		spec.CollectVars(vars, r.Command, r.Cwd, r.Stdin.Inline, r.Stdin.File)
		for _, v := range r.Env {
			spec.CollectVars(vars, v)
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
		spec.CollectVars(vars, h.Path)
		spec.CollectVars(vars, h.Body)
		spec.CollectVars(vars, h.BodyFile, h.BodyTo)
		for _, v := range h.Form {
			spec.CollectVars(vars, v)
		}
		for _, f := range h.Files {
			spec.CollectVars(vars, f.Path)
		}

	case spec.StepQuery:
		q := step.Query
		st.SQL = q.SQL
		st.Runner = q.Runner
		st.Action = "SQL query via " + q.Runner
		spec.CollectVars(vars, q.SQL)

	case spec.StepGRPC:
		g := step.GRPC
		st.Method = g.Method
		st.Runner = g.Runner
		st.Action = "gRPC " + g.Method + " via " + g.Runner
		spec.CollectVars(vars, g.Method)

	case spec.StepCDP:
		c := step.CDP
		st.Runner = c.Runner
		st.Action = spec.CDPActionSummary(c)
		for _, a := range c.Actions {
			spec.CollectVars(vars, a.Navigate, a.WaitVisible, a.WaitHidden, a.Click, a.Check, a.Uncheck, a.Text, a.Eval)
			if a.SendKeys != nil {
				spec.CollectVars(vars, a.SendKeys.Selector, a.SendKeys.Value)
			}
			if a.Press != nil {
				spec.CollectVars(vars, a.Press.Selector, a.Press.Key)
			}
			if a.Select != nil {
				spec.CollectVars(vars, a.Select.Selector, a.Select.Value)
			}
			if a.Screenshot != nil {
				spec.CollectVars(vars, a.Screenshot.Path, a.Screenshot.Selector)
			}
			if a.Attribute != nil {
				spec.CollectVars(vars, a.Attribute.Selector, a.Attribute.Name)
			}
			if a.Upload != nil {
				spec.CollectVars(vars, a.Upload.Selector, a.Upload.File)
			}
			if a.Download != nil {
				spec.CollectVars(vars, a.Download.Click, a.Download.Dir)
			}
		}

	case spec.StepMockServer:
		ms := step.MockServer
		st.Target = ms.Name
		st.Action = fmt.Sprintf("start suite mock server %s (%d routes)", ms.Name, len(ms.Routes))

	case spec.StepSignal:
		sg := step.Signal
		st.Target = sg.Service
		st.Action = "signal SIG" + spec.NormalizeSignalName(sg.Signal) + " to service " + sg.Service
		if sg.Wait != nil {
			timeout := sg.Wait.Timeout
			if timeout == "" {
				timeout = "5s"
			}
			st.Action += ", wait up to " + timeout + " for exit"
		}
		spec.CollectVars(vars, sg.Service)

	case spec.StepPTY:
		pt := step.PTY
		st.Command = pt.Command
		st.Shell = pt.Shell != nil && *pt.Shell
		st.ClearEnv = pt.ClearEnvEnabled()
		st.PassEnv = pt.PassEnv
		st.Action = "interactive (pty) " + pt.Command
		spec.CollectVars(vars, pt.Command, pt.Cwd)
		for _, v := range pt.Env {
			spec.CollectVars(vars, v)
		}
		for _, a := range pt.Session {
			if a.Send != nil && a.Send.Text != nil {
				spec.CollectVars(vars, *a.Send.Text)
			}
			spec.CollectVars(vars, a.Expect)
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

func condition(c *spec.Condition) *Condition {
	if c == nil || (c.OS == "" && c.Env == "" && c.Command == "") {
		return nil
	}
	return &Condition{OS: c.OS, Env: c.Env, Command: c.Command}
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
