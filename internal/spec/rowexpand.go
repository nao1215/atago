package spec

// ExpandRow substitutes a matrix row's bindings into s, leaving every other
// reference exactly as written. It is display-only: the row is the one binding
// that is part of the SPEC, so `${env:...}`, `${workdir}`, and a value a `store`
// captures at run time stay literal — a describer that resolved those would
// print host state into a document, and could put a secret into a committed
// page.
//
// The `$${name}` escape is honored the way it is everywhere else, so a literal
// `${who}` an author escaped survives.
func ExpandRow(s string, row map[string]string) string {
	if len(row) == 0 || !varRef.MatchString(s) {
		return s
	}
	return varRef.ReplaceAllStringFunc(s, func(m string) string {
		sub := varRef.FindStringSubmatch(m)
		escaped, name := sub[1], sub[2]
		if escaped != "" {
			return "${" + name + "}"
		}
		if v, ok := row[name]; ok {
			return v
		}
		return m
	})
}

// WalkStepStrings returns a copy of step with visit applied to every
// author-written string the engine expands for that kind — the one field list
// from stepstrings.go, dispatched per kind the way CollectStepVars reads it, so
// a describer cannot expand a different set of fields than the engine does.
//
// A retry's `until:` assertion is included: the poll loop expands it per
// attempt, so it carries live references exactly as the step's own fields do.
func WalkStepStrings(step *Step, visit func(string) string) *Step {
	if step == nil {
		return nil
	}
	c := *step
	switch step.Kind() {
	case StepFixture:
		c.Fixture = WalkFixtureStrings(step.Fixture, visit)
	case StepService:
		c.Service = WalkServiceStrings(step.Service, visit)
	case StepRun:
		c.Run = WalkRunStrings(step.Run, visit)
		c.Run.Retry = walkRetryStrings(step.Run.Retry, visit)
	case StepHTTP:
		c.HTTP = WalkHTTPStrings(step.HTTP, visit)
		c.HTTP.Retry = walkRetryStrings(step.HTTP.Retry, visit)
	case StepQuery:
		c.Query = WalkQueryStrings(step.Query, visit)
	case StepGRPC:
		c.GRPC = WalkGRPCStrings(step.GRPC, visit)
	case StepCDP:
		c.CDP = WalkCDPStrings(step.CDP, visit)
	case StepPTY:
		c.PTY = WalkPTYStrings(step.PTY, visit)
	case StepSignal:
		c.Signal = WalkSignalStrings(step.Signal, visit)
	case StepAssert:
		c.Assert = WalkAssertStrings(step.Assert, visit)
	case StepStore:
		c.Store = WalkStoreStrings(step.Store, visit)
	case StepMockServer:
		// A mock server is started verbatim; nothing in it is expanded.
	}
	return &c
}

// walkRetryStrings applies visit to a retry's until assertion, carrying the
// timing knobs unchanged.
func walkRetryStrings(r *Retry, visit func(string) string) *Retry {
	if r == nil {
		return nil
	}
	c := *r
	c.Until = WalkAssertStrings(r.Until, visit)
	return &c
}

// ExpandScenarioRow returns a copy of sc with its matrix row bound into every
// string the engine would expand — the view `atago doc`, `explain`, `manifest`,
// and `list` describe, so the four cannot answer differently about what one
// expanded row runs. A scenario without a row is returned unchanged.
func ExpandScenarioRow(sc *Scenario) *Scenario {
	if sc == nil || len(sc.Vars) == 0 {
		return sc
	}
	visit := func(s string) string { return ExpandRow(s, sc.Vars) }
	c := *sc
	c.Steps = walkSteps(sc.Steps, visit)
	c.Teardown = walkSteps(sc.Teardown, visit)
	if len(sc.Services) > 0 {
		c.Services = make([]Service, len(sc.Services))
		for i := range sc.Services {
			c.Services[i] = *WalkServiceStrings(&sc.Services[i], visit)
		}
	}
	return &c
}

func walkSteps(steps []Step, visit func(string) string) []Step {
	if len(steps) == 0 {
		return steps
	}
	out := make([]Step, len(steps))
	for i := range steps {
		out[i] = *WalkStepStrings(&steps[i], visit)
	}
	return out
}
