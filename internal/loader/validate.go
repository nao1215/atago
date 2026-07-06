package loader

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/spec"
)

var validOS = map[string]bool{"linux": true, "darwin": true, "windows": true}

// validate runs schema and semantic checks and
// returns all problems found so the user can fix them in one pass.
func validate(s *spec.Spec) []string {
	var errs []string
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	if s.Version != "1" {
		add("version must be \"1\" (got %q)", s.Version)
	}
	if s.Suite.Name == "" {
		add("suite.name is required")
	}
	validateSuiteTimeout(add, &s.Suite)
	if len(s.Scenarios) == 0 {
		add("scenarios must contain at least one scenario")
	}
	validateRunners(add, s.Runners)
	validateDefaults(add, s.Defaults)
	validateSuiteBlock(add, "suite.setup", s.Suite.Setup, s.Runners, true)
	validateSuiteBlock(add, "suite.teardown", s.Suite.Teardown, s.Runners, false)

	// Suite services are legal signal targets from any scenario (#23), and
	// suite mock servers are legal mock-assert targets (#24).
	suiteServiceNames := map[string]bool{}
	suiteMockNames := map[string]bool{}
	for i := range s.Suite.Setup {
		if svc := s.Suite.Setup[i].Service; svc != nil && svc.Name != "" {
			suiteServiceNames[svc.Name] = true
		}
		if ms := s.Suite.Setup[i].MockServer; ms != nil && ms.Name != "" {
			suiteMockNames[ms.Name] = true
		}
	}

	seen := make(map[string]bool, len(s.Scenarios))
	for i := range s.Scenarios {
		sc := &s.Scenarios[i]
		where := fmt.Sprintf("scenarios[%d]", i)
		if sc.Name == "" {
			add("%s.name is required", where)
		} else {
			if seen[sc.Name] {
				add("duplicate scenario name %q", sc.Name)
			}
			seen[sc.Name] = true
			where = fmt.Sprintf("scenario %q", sc.Name)
		}
		validateCondition(add, where, "skip", sc.Skip)
		validateCondition(add, where, "only", sc.Only)
		validateServices(add, where, sc.Services)
		serviceNames := maps.Clone(suiteServiceNames)
		for j := range sc.Services {
			if sc.Services[j].Name != "" {
				serviceNames[sc.Services[j].Name] = true
			}
		}
		mockNames := maps.Clone(suiteMockNames)
		validateMockServers(add, where, sc.MockServers, mockNames)
		if len(sc.Steps) == 0 {
			add("%s: steps must contain at least one step", where)
			continue
		}
		// A screen assert renders a pty step's terminal (#27) and a duration
		// assert bounds the immediately preceding measurable step (#31):
		// reject placements no step could feed.
		ptySeen := false
		prevMeasurable := false
		prevRunOrPTY := false
		for j := range sc.Steps {
			sw := fmt.Sprintf("%s.steps[%d]", where, j)
			st := &sc.Steps[j]
			if st.Kind() == spec.StepPTY {
				ptySeen = true
			}
			if st.Assert != nil && st.Assert.Screen != nil && !ptySeen {
				add("%s.assert.screen requires a preceding pty step (the screen is the pty step's rendered terminal)", sw)
			}
			if st.Assert != nil && st.Assert.Duration != nil && !prevMeasurable {
				add("%s.assert.duration requires an immediately preceding run/http/query/grpc/pty step (the step whose wall-clock time it bounds)", sw)
			}
			// changes bounds the workdir delta of the immediately preceding
			// run/pty step (#70): reject a placement no such step feeds.
			if st.Assert != nil && st.Assert.Changes != nil && !prevRunOrPTY {
				add("%s.assert.changes requires an immediately preceding run/pty step (the step whose workdir delta it pins); combine it with the assert block directly after the step (one assert may set exit_code, stdout, and changes together)", sw)
			}
			validateStep(add, sw, st, s.Runners, serviceNames, mockNames)
			prevMeasurable = measurableStep(st.Kind())
			prevRunOrPTY = st.Kind() == spec.StepRun || st.Kind() == spec.StepPTY
		}
		for j := range sc.Teardown {
			tw := fmt.Sprintf("%s.teardown[%d]", where, j)
			st := &sc.Teardown[j]
			if st.Assert != nil && st.Assert.Screen != nil && !ptySeen {
				add("%s.assert.screen requires a pty step in the scenario", tw)
			}
			// The workdir delta is only tracked around Steps, so a changes assert
			// in teardown could never be fed (#70).
			if st.Assert != nil && st.Assert.Changes != nil {
				add("%s.assert.changes is not supported in teardown (the workdir delta is tracked only around the scenario's steps)", tw)
			}
			validateStep(add, tw, st, s.Runners, serviceNames, mockNames)
		}
	}
	return errs
}

// validateSuiteTimeout checks the suite-level default step timeout (#17).
func validateSuiteTimeout(add func(string, ...any), s *spec.Suite) {
	if s.Timeout == "" {
		return
	}
	if d, err := time.ParseDuration(s.Timeout); err != nil {
		add("suite.timeout %q is not a valid duration (e.g. \"2m\"); use \"0\" to disable the built-in default", s.Timeout)
	} else if d < 0 {
		add("suite.timeout must not be negative (got %q); a wall-clock bound is never below zero", s.Timeout)
	}
}

// validateDefaults checks the top-level `defaults:` block. The merge only
// covers non-identity, non-per-step fields, so a value the loader would silently
// ignore is reported here instead. Fields the loader does merge are validated on
// the concrete elements after applyDefaults (and, for a shared readiness probe,
// here too, so a wrong probe fails even when no scenario declares a service).
func validateDefaults(add func(string, ...any), d *spec.Defaults) {
	if d == nil {
		return
	}
	if r := d.Run; r != nil {
		if r.Command != "" {
			add("defaults.run.command is not supported (command is per-step)")
		}
		if r.Retry != nil {
			add("defaults.run.retry is not supported (retry is per-step)")
		}
		if !r.Stdin.IsZero() {
			add("defaults.run.stdin is not supported (stdin is per-step input data, like command)")
		}
		if r.Timeout != "" {
			if d, err := time.ParseDuration(r.Timeout); err != nil {
				add("defaults.run.timeout %q is not a valid duration (e.g. \"30s\")", r.Timeout)
			} else if d < 0 {
				add("defaults.run.timeout must not be negative (got %q); a wall-clock bound is never below zero", r.Timeout)
			}
		}
		validateHermeticEnv(add, "defaults.run", r.ClearEnv, r.PassEnv)
	}
	if sv := d.Service; sv != nil {
		if sv.Name != "" {
			add("defaults.service.name is not supported (each service names itself)")
		}
		if sv.Command != "" {
			add("defaults.service.command is not supported (each service sets its own command)")
		}
		validateHermeticEnv(add, "defaults.service", sv.ClearEnv, sv.PassEnv)
		validateReady(add, "defaults.service", sv.Ready)
	}
}

func validateCondition(add func(string, ...any), where, key string, c *spec.Condition) {
	if c == nil {
		return
	}
	if c.OS != "" && !validOS[c.OS] {
		add("%s.%s.os %q is invalid (want linux, darwin, or windows)", where, key, c.OS)
	}
}

// stepRunnerTypes maps a step action to the runner types it accepts, mirroring
// the engine's dispatch so a wrong or missing runner reference fails at load
// time (exit 2) instead of surfacing mid-run as an execution error.
var stepRunnerTypes = map[string][]string{
	"run":   {"cmd", "ssh"},
	"http":  {"http"},
	"query": {"db"},
	"grpc":  {"grpc"},
	"cdp":   {"browser"},
}

// validateRunnerRef checks that a step's named runner exists and has a type the
// step can drive. An empty name is fine here: steps that require a runner
// enforce that separately.
func validateRunnerRef(add func(string, ...any), where, stepKind, name string, runners map[string]spec.Runner) {
	if name == "" {
		return
	}
	r, ok := runners[name]
	if !ok {
		declared := slices.Sorted(maps.Keys(runners))
		if len(declared) == 0 {
			add("%s.%s.runner %q is not declared (the spec has no runners: block)", where, stepKind, name)
			return
		}
		add("%s.%s.runner %q is not declared under runners: (declared: %s)", where, stepKind, name, strings.Join(declared, ", "))
		return
	}
	want := stepRunnerTypes[stepKind]
	// An unknown/empty type is reported by validateRunners already.
	if r.Type != "" && validRunnerType[r.Type] && !slices.Contains(want, r.Type) {
		add("%s: runner %q is a %s runner; a %s step needs a %s runner", where, name, r.Type, stepKind, strings.Join(want, " or "))
	}
}

// validateSuiteBlock checks suite.setup / suite.teardown (#7): steps run once
// per suite in the ${suitedir} scratch dir, so only the suite-scoped kinds are
// allowed — fixture, run, store, assert, and (setup only) `service:`. The
// runner-backed kinds (http/query/grpc/cdp) are per-scenario machinery and are
// rejected with a pointer to where they belong.
func validateSuiteBlock(add func(string, ...any), where string, steps []spec.Step, runners map[string]spec.Runner, allowService bool) {
	seenService := map[string]bool{}
	seenMock := map[string]bool{}
	for i := range steps {
		st := &steps[i]
		sw := fmt.Sprintf("%s[%d]", where, i)
		keys := st.SetKeys()
		if len(keys) != 1 {
			add("%s: step must set exactly one action (got %v)", sw, keys)
			continue
		}
		switch st.Kind() {
		case spec.StepFixture:
			validateFixture(add, sw, st.Fixture)
		case spec.StepRun:
			validateRunStep(add, sw, st.Run, runners, false)
		case spec.StepStore:
			validateStore(add, sw, st.Store)
		case spec.StepAssert:
			validateAssert(add, sw, st.Assert, nil)
		case spec.StepService:
			if !allowService {
				add("%s: service steps are only allowed in suite.setup", sw)
				continue
			}
			svc := st.Service
			if svc.Name == "" {
				add("%s.service.name is required", sw)
			} else if seenService[svc.Name] {
				add("%s: duplicate suite service name %q", where, svc.Name)
			} else {
				seenService[svc.Name] = true
			}
			if svc.Command == "" {
				add("%s.service.command is required", sw)
			}
			validateHermeticEnv(add, sw+".service", svc.ClearEnv, svc.PassEnv)
			validateReady(add, sw+".service", svc.Ready)
		case spec.StepMockServer:
			// Mock servers follow the service rule (#24): setup-only, so the
			// position in the sequence controls ordering.
			if !allowService {
				add("%s: mock_server steps are only allowed in suite.setup", sw)
				continue
			}
			ms := st.MockServer
			if ms.Name == "" {
				add("%s.mock_server.name is required", sw)
			} else if seenMock[ms.Name] {
				add("%s: duplicate suite mock server name %q", where, ms.Name)
			} else {
				seenMock[ms.Name] = true
			}
			validateMockRoutes(add, sw+".mock_server", ms.Routes)
		default:
			add("%s: %s steps are per-scenario (they need a scenario workdir and runners); move it into a scenario", sw, st.Kind())
		}
	}
}

func validateStep(add func(string, ...any), where string, st *spec.Step, runners map[string]spec.Runner, serviceNames, mockNames map[string]bool) {
	keys := st.SetKeys()
	switch len(keys) {
	case 0:
		add("%s: step must set exactly one of fixture/run/http/query/grpc/cdp/assert/store/pty/signal (got none)", where)
		return
	case 1:
	default:
		add("%s: step must set exactly one action, but set %v", where, keys)
		return
	}

	switch st.Kind() {
	case spec.StepFixture:
		validateFixture(add, where, st.Fixture)
	case spec.StepRun:
		validateRunStep(add, where, st.Run, runners, true)
	case spec.StepAssert:
		validateAssert(add, where, st.Assert, mockNames)
	case spec.StepHTTP:
		validateRunnerRef(add, where, "http", st.HTTP.Runner, runners)
		if st.HTTP.Method == "" {
			add("%s.http.method is required", where)
		}
		validateHTTPPayload(add, where, st.HTTP)
		for i, f := range st.HTTP.Files {
			if f.Field == "" {
				add("%s.http.files[%d].field is required (the multipart form field name)", where, i)
			}
			if f.Path == "" {
				add("%s.http.files[%d].path is required (the workdir-relative file to attach)", where, i)
			}
		}
		validateRetry(add, where+".http", st.HTTP.Retry)
	case spec.StepQuery:
		if st.Query.Runner == "" {
			add("%s.query.runner is required", where)
		}
		validateRunnerRef(add, where, "query", st.Query.Runner, runners)
		if st.Query.SQL == "" {
			add("%s.query.sql is required", where)
		}
	case spec.StepGRPC:
		if st.GRPC.Runner == "" {
			add("%s.grpc.runner is required", where)
		}
		validateRunnerRef(add, where, "grpc", st.GRPC.Runner, runners)
		if st.GRPC.Method == "" {
			add("%s.grpc.method is required", where)
		}
	case spec.StepCDP:
		if st.CDP.Runner == "" {
			add("%s.cdp.runner is required", where)
		}
		validateRunnerRef(add, where, "cdp", st.CDP.Runner, runners)
		if len(st.CDP.Actions) == 0 {
			add("%s.cdp.actions must contain at least one action", where)
		}
		validateCDPActions(add, where, st.CDP.Actions)
	case spec.StepStore:
		validateStore(add, where, st.Store)
	case spec.StepService:
		add("%s: service steps are only allowed in suite.setup (scenario-scoped peers go under the scenario's services: list)", where)
	case spec.StepMockServer:
		add("%s: mock_server steps are only allowed in suite.setup (a scenario-scoped stub goes under the scenario's mock_servers: list)", where)
	case spec.StepPTY:
		validatePTY(add, where, st.PTY)
	case spec.StepSignal:
		validateSignal(add, where, st.Signal, serviceNames)
	}
}

// measurableStep reports whether a step kind records a wall-clock duration a
// following duration assert can bound (#31).
func measurableStep(k spec.StepKind) bool {
	switch k {
	case spec.StepRun, spec.StepHTTP, spec.StepQuery, spec.StepGRPC, spec.StepPTY:
		return true
	default:
		return false
	}
}
