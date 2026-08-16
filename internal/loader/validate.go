package loader

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/diag"
	"github.com/nao1215/atago/internal/scrub"
	"github.com/nao1215/atago/internal/spec"
)

var validOS = map[string]bool{"linux": true, "darwin": true, "windows": true}

// firstControlChar returns a readable label for the first control character in
// name (a newline, tab, or other C0/DEL byte), or "" when there is none. A name
// carrying one silently corrupts the aligned `atago list` table and the
// generated Markdown headings/anchors, so the loader rejects it up front like it
// rejects an empty or duplicate name.
func firstControlChar(name string) string {
	for _, r := range name {
		if r == '\t' || r == '\n' || r == '\r' || (r < 0x20) || r == 0x7f {
			switch r {
			case '\t':
				return `"\t"`
			case '\n':
				return `"\n"`
			case '\r':
				return `"\r"`
			default:
				return fmt.Sprintf("U+%04X", r)
			}
		}
	}
	return ""
}

// addFunc records one validation problem under the diagnostic code that names
// it. Taking the code as the first argument is what keeps the published error
// reference honest: a check added later cannot report a problem without
// deciding which diagnostic it is, because omitting the code does not compile.
type addFunc func(code diag.Code, format string, args ...any)

// validate runs schema and semantic checks and
// returns all problems found so the user can fix them in one pass.
func validate(s *spec.Spec) []string {
	var errs []string
	add := func(code diag.Code, format string, args ...any) {
		errs = append(errs, code.Annotate(fmt.Sprintf(format, args...)))
	}

	if s.Version != "1" {
		add(diag.SpecVersion, "version must be \"1\" (got %q)", s.Version)
	}
	if s.Suite.Name == "" {
		add(diag.RequiredKey, "suite.name is required")
	} else if c := firstControlChar(s.Suite.Name); c != "" {
		add(diag.ControlCharacter, "suite.name must not contain the control character %s (it breaks list output and generated docs)", c)
	}
	validateSuiteTimeout(add, &s.Suite)
	validateScrub(add, s.Scrub)
	if len(s.Scenarios) == 0 {
		add(diag.EmptyList, "scenarios must contain at least one scenario")
	}
	validateRunners(add, s.Runners)
	validateDefaults(add, s.Defaults)
	validateSuiteBlock(add, "suite.setup", s.Suite.Setup, s.Runners, true)
	validateSuiteBlock(add, "suite.teardown", s.Suite.Teardown, s.Runners, false)

	// Suite services are legal signal targets from any scenario (#23), and
	// suite mock servers are legal mock-assert targets (#24).
	suiteServiceNames, suiteMockNames := suiteResourceNames(s)

	seen := make(map[string]bool, len(s.Scenarios))
	for i := range s.Scenarios {
		validateScenario(add, s, i, seen, suiteServiceNames, suiteMockNames)
	}
	return errs
}

// suiteResourceNames collects the names of services and mock servers declared
// in suite.setup, which every scenario may target.
func suiteResourceNames(s *spec.Spec) (services, mocks map[string]bool) {
	services = map[string]bool{}
	mocks = map[string]bool{}
	for i := range s.Suite.Setup {
		if svc := s.Suite.Setup[i].Service; svc != nil && svc.Name != "" {
			services[svc.Name] = true
		}
		if ms := s.Suite.Setup[i].MockServer; ms != nil && ms.Name != "" {
			mocks[ms.Name] = true
		}
	}
	return services, mocks
}

// validateScenario checks one scenario: its identity, gates, services, and
// every step and teardown step. seen tracks scenario names across the suite
// for the duplicate check.
func validateScenario(add addFunc, s *spec.Spec, i int, seen, suiteServiceNames, suiteMockNames map[string]bool) {
	sc := &s.Scenarios[i]
	where := fmt.Sprintf("scenarios[%d]", i)
	if sc.Name == "" {
		add(diag.RequiredKey, "%s.name is required", where)
	} else {
		if c := firstControlChar(sc.Name); c != "" {
			add(diag.ControlCharacter, "%s.name must not contain the control character %s (it breaks list output and generated docs)", where, c)
		}
		if seen[sc.Name] {
			add(diag.DuplicateName, "duplicate scenario name %q", sc.Name)
		}
		seen[sc.Name] = true
		where = fmt.Sprintf("scenario %q", sc.Name)
	}
	// A tag list is a set: --tag and --skip-tag ask whether a scenario carries
	// one, so repeating an entry selects nothing extra. It is not harmless
	// though — the generated docs count tag occurrences, so a scenario listing
	// `smoke` twice made the summary claim two scenarios carry it.
	tags := map[string]bool{}
	for _, tag := range sc.Tags {
		if tags[tag] {
			add(diag.DuplicateName, "%s: duplicate tag %q; a tag list is a set, and the generated docs count it twice", where, tag)
		}
		tags[tag] = true
	}
	validateCondition(add, where, "skip", sc.Skip)
	validateCondition(add, where, "only", sc.Only)
	validateExpectFail(add, where, sc.ExpectFail)
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
		add(diag.EmptyList, "%s: steps must contain at least one step", where)
		return
	}
	ptySeen := validateScenarioSteps(add, where, sc, s.Runners, serviceNames, mockNames)
	validateScenarioTeardown(add, where, sc, s.Runners, serviceNames, mockNames, ptySeen)
}

// validateScenarioSteps checks the scenario's steps in order, enforcing the
// placement rules that tie an assert to the step that feeds it. It reports
// whether the scenario contains a pty step, which teardown asserts may render.
func validateScenarioSteps(add addFunc, where string, sc *spec.Scenario, runners map[string]spec.Runner, serviceNames, mockNames map[string]bool) (ptySeen bool) {
	// A screen assert renders a pty step's terminal (#27) and a duration
	// assert bounds the immediately preceding measurable step (#31):
	// reject placements no step could feed.
	prevMeasurable := false
	prevRunOrPTY := false
	for j := range sc.Steps {
		sw := fmt.Sprintf("%s.steps[%d]", where, j)
		st := &sc.Steps[j]
		if st.Kind() == spec.StepPTY {
			ptySeen = true
		}
		if st.Assert != nil && st.Assert.Screen != nil && !ptySeen {
			add(diag.AssertNeedsStep, "%s.assert.screen requires a preceding pty step (the screen is the pty step's rendered terminal)", sw)
		}
		if st.Assert != nil && st.Assert.Duration != nil && !prevMeasurable {
			add(diag.AssertNeedsStep, "%s.assert.duration requires an immediately preceding run/http/query/grpc/pty step (the step whose wall-clock time it bounds)", sw)
		}
		// changes bounds the workdir delta of the immediately preceding
		// run/pty step (#70): reject a placement no such step feeds.
		if st.Assert != nil && st.Assert.Changes != nil && !prevRunOrPTY {
			add(diag.AssertNeedsStep, "%s.assert.changes requires an immediately preceding run/pty step (the step whose workdir delta it pins); combine it with the assert block directly after the step (one assert may set exit_code, stdout, and changes together)", sw)
		}
		validateStep(add, sw, st, runners, serviceNames, mockNames)
		prevMeasurable = measurableStep(st.Kind())
		prevRunOrPTY = st.Kind() == spec.StepRun || st.Kind() == spec.StepPTY
	}
	return ptySeen
}

// validateScenarioTeardown checks the scenario's teardown steps, whose asserts
// may render a pty screen but can never be fed a workdir delta.
func validateScenarioTeardown(add addFunc, where string, sc *spec.Scenario, runners map[string]spec.Runner, serviceNames, mockNames map[string]bool, ptySeen bool) {
	for j := range sc.Teardown {
		tw := fmt.Sprintf("%s.teardown[%d]", where, j)
		st := &sc.Teardown[j]
		if st.Assert != nil && st.Assert.Screen != nil && !ptySeen {
			add(diag.AssertNeedsStep, "%s.assert.screen requires a pty step in the scenario", tw)
		}
		// The workdir delta is only tracked around Steps, so a changes assert
		// in teardown could never be fed (#70).
		if st.Assert != nil && st.Assert.Changes != nil {
			add(diag.BlockNotHere, "%s.assert.changes is not supported in teardown (the workdir delta is tracked only around the scenario's steps)", tw)
		}
		validateStep(add, tw, st, runners, serviceNames, mockNames)
	}
}

// validateSuiteTimeout checks the suite-level default step timeout (#17).
func validateSuiteTimeout(add addFunc, s *spec.Suite) {
	if s.Timeout == "" {
		return
	}
	if d, err := time.ParseDuration(s.Timeout); err != nil {
		add(diag.BadDuration, "suite.timeout %q is not a valid duration (e.g. \"2m\"); use \"0\" to disable the built-in default", s.Timeout)
	} else if d < 0 {
		add(diag.NegativeValue, "suite.timeout must not be negative (got %q); a wall-clock bound is never below zero", s.Timeout)
	}
}

// validateScrub checks the top-level `scrub:` rules (#137): every pattern must
// be non-empty and compile as a Go regexp, so a broken rule fails at load rather
// than silently normalizing nothing (or, for an empty pattern, matching between
// every byte). The compile check reuses scrub.New so validation and runtime
// agree on what a valid rule is.
func validateScrub(add addFunc, rules []spec.ScrubRule) {
	for i, r := range rules {
		if r.Pattern == "" {
			add(diag.RequiredKey, "scrub[%d].pattern is required (a regex to normalize; e.g. \"req-\\d+\")", i)
		}
	}
	if _, err := scrub.New(rules); err != nil {
		add(diag.BadRegexp, "%v", err)
	}
}

// validateDefaults checks the top-level `defaults:` block. The merge only
// covers non-identity, non-per-step fields, so a value the loader would silently
// ignore is reported here instead. Fields the loader does merge are validated on
// the concrete elements after applyDefaults (and, for a shared readiness probe,
// here too, so a wrong probe fails even when no scenario declares a service).
func validateDefaults(add addFunc, d *spec.Defaults) {
	if d == nil {
		return
	}
	if r := d.Run; r != nil {
		if r.Command != "" {
			add(diag.KeyNotHere, "defaults.run.command is not supported (command is per-step)")
		}
		if r.Retry != nil {
			add(diag.KeyNotHere, "defaults.run.retry is not supported (retry is per-step)")
		}
		if !r.Stdin.IsZero() {
			add(diag.KeyNotHere, "defaults.run.stdin is not supported (stdin is per-step input data, like command)")
		}
		nonNegativeDuration(add, "defaults.run.timeout", r.Timeout, "30s")
		validateHermeticEnv(add, "defaults.run", r.ClearEnv, r.PassEnv)
	}
	if scn := d.Scenario; scn != nil {
		// The gates are validated here as well as on each scenario, so a
		// malformed default fails even when every scenario states its own and
		// the default is therefore never applied.
		validateCondition(add, "defaults.scenario", "only", scn.Only)
		validateCondition(add, "defaults.scenario", "skip", scn.Skip)
	}
	if sv := d.Service; sv != nil {
		if sv.Name != "" {
			add(diag.KeyNotHere, "defaults.service.name is not supported (each service names itself)")
		}
		if sv.Command != "" {
			add(diag.KeyNotHere, "defaults.service.command is not supported (each service sets its own command)")
		}
		validateHermeticEnv(add, "defaults.service", sv.ClearEnv, sv.PassEnv)
		validateReady(add, "defaults.service", sv.Ready)
	}
}

func validateCondition(add addFunc, where, key string, c *spec.Condition) {
	if c == nil {
		return
	}
	if c.OS != "" && !validOS[c.OS] {
		add(diag.NotAllowedValue, "%s.%s.os %q is invalid (want linux, darwin, or windows)", where, key, c.OS)
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
func validateRunnerRef(add addFunc, where, stepKind, name string, runners map[string]spec.Runner) {
	if name == "" {
		return
	}
	r, ok := runners[name]
	if !ok {
		declared := slices.Sorted(maps.Keys(runners))
		if len(declared) == 0 {
			add(diag.RunnerNotDeclared, "%s.%s.runner %q is not declared (the spec has no runners: block)", where, stepKind, name)
			return
		}
		add(diag.RunnerNotDeclared, "%s.%s.runner %q is not declared under runners: (declared: %s)", where, stepKind, name, strings.Join(declared, ", "))
		return
	}
	want := stepRunnerTypes[stepKind]
	// An unknown/empty type is reported by validateRunners already.
	if r.Type != "" && validRunnerType[r.Type] && !slices.Contains(want, r.Type) {
		add(diag.RunnerTypeMismatch, "%s: runner %q is a %s runner; a %s step needs a %s runner", where, name, r.Type, stepKind, strings.Join(want, " or "))
	}
}

// validateSuiteBlock checks suite.setup / suite.teardown (#7): steps run once
// per suite in the ${suitedir} scratch dir, so only the suite-scoped kinds are
// allowed — fixture, run, store, assert, and (setup only) `service:`. The
// runner-backed kinds (http/query/grpc/cdp) are per-scenario machinery and are
// rejected with a pointer to where they belong.
func validateSuiteBlock(add addFunc, where string, steps []spec.Step, runners map[string]spec.Runner, allowService bool) {
	seenService := map[string]bool{}
	seenMock := map[string]bool{}
	for i := range steps {
		st := &steps[i]
		sw := fmt.Sprintf("%s[%d]", where, i)
		keys := st.SetKeys()
		if len(keys) != 1 {
			add(diag.StepManyActions, "%s: step must set exactly one action (got %v)", sw, keys)
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
				add(diag.BlockNotHere, "%s: service steps are only allowed in suite.setup", sw)
				continue
			}
			svc := st.Service
			if svc.Name == "" {
				add(diag.RequiredKey, "%s.service.name is required", sw)
			} else if seenService[svc.Name] {
				add(diag.DuplicateName, "%s: duplicate suite service name %q", where, svc.Name)
			} else {
				seenService[svc.Name] = true
			}
			if svc.Command == "" {
				add(diag.RequiredKey, "%s.service.command is required", sw)
			}
			validateHermeticEnv(add, sw+".service", svc.ClearEnv, svc.PassEnv)
			validateReady(add, sw+".service", svc.Ready)
		case spec.StepMockServer:
			// Mock servers follow the service rule (#24): setup-only, so the
			// position in the sequence controls ordering.
			if !allowService {
				add(diag.BlockNotHere, "%s: mock_server steps are only allowed in suite.setup", sw)
				continue
			}
			ms := st.MockServer
			if ms.Name == "" {
				add(diag.RequiredKey, "%s.mock_server.name is required", sw)
			} else if seenMock[ms.Name] {
				add(diag.DuplicateName, "%s: duplicate suite mock server name %q", where, ms.Name)
			} else {
				seenMock[ms.Name] = true
			}
			validateMockRoutes(add, sw+".mock_server", ms.Routes)
		default:
			add(diag.BlockNotHere, "%s: %s steps are per-scenario (they need a scenario workdir and runners); move it into a scenario", sw, st.Kind())
		}
	}
}

func validateStep(add addFunc, where string, st *spec.Step, runners map[string]spec.Runner, serviceNames, mockNames map[string]bool) {
	keys := st.SetKeys()
	switch len(keys) {
	case 0:
		add(diag.StepNoAction, "%s: step must set exactly one of fixture/run/http/query/grpc/cdp/assert/store/pty/signal (got none)", where)
		return
	case 1:
	default:
		add(diag.StepManyActions, "%s: step must set exactly one action, but set %v", where, keys)
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
			add(diag.RequiredKey, "%s.http.method is required", where)
		}
		validateHTTPPayload(add, where, st.HTTP)
		for i, f := range st.HTTP.Files {
			if f.Field == "" {
				add(diag.RequiredKey, "%s.http.files[%d].field is required (the multipart form field name)", where, i)
			}
			if f.Path == "" {
				add(diag.RequiredKey, "%s.http.files[%d].path is required (the workdir-relative file to attach)", where, i)
			}
		}
		validateRetry(add, where+".http", st.HTTP.Retry)
	case spec.StepQuery:
		if st.Query.Runner == "" {
			add(diag.RequiredKey, "%s.query.runner is required", where)
		}
		validateRunnerRef(add, where, "query", st.Query.Runner, runners)
		if st.Query.SQL == "" {
			add(diag.RequiredKey, "%s.query.sql is required", where)
		}
	case spec.StepGRPC:
		if st.GRPC.Runner == "" {
			add(diag.RequiredKey, "%s.grpc.runner is required", where)
		}
		validateRunnerRef(add, where, "grpc", st.GRPC.Runner, runners)
		if st.GRPC.Method == "" {
			add(diag.RequiredKey, "%s.grpc.method is required", where)
		}
	case spec.StepCDP:
		if st.CDP.Runner == "" {
			add(diag.RequiredKey, "%s.cdp.runner is required", where)
		}
		validateRunnerRef(add, where, "cdp", st.CDP.Runner, runners)
		if len(st.CDP.Actions) == 0 {
			add(diag.EmptyList, "%s.cdp.actions must contain at least one action", where)
		}
		validateCDPActions(add, where, st.CDP.Actions)
	case spec.StepStore:
		validateStore(add, where, st.Store)
	case spec.StepService:
		add(diag.BlockNotHere, "%s: service steps are only allowed in suite.setup (scenario-scoped peers go under the scenario's services: list)", where)
	case spec.StepMockServer:
		add(diag.BlockNotHere, "%s: mock_server steps are only allowed in suite.setup (a scenario-scoped stub goes under the scenario's mock_servers: list)", where)
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

// nonNegativeDuration validates a wall-clock duration field that may be zero
// (zero usually means "disabled" or "no wait") in one place, so the accepted
// bounds and the failure wording cannot drift between the many duration keys —
// they already had: retry.interval accepted negatives that every timeout field
// rejected. key is the fully-qualified field ("scenario X.steps[0].run.timeout"),
// example the hint value shown for an unparsable string.
func nonNegativeDuration(add addFunc, key, val, example string) {
	if val == "" {
		return
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		add(diag.BadDuration, "%s %q is not a valid duration (e.g. %q)", key, val, example)
		return
	}
	if d < 0 {
		add(diag.NegativeValue, "%s must not be negative (got %q); a wall-clock bound is never below zero", key, val)
	}
}

// workdirRelativeDir checks a `cwd:` — the one path field a step hands to the
// OS as a directory rather than opening as a file.
//
// A `../` one used to walk straight out of the scenario workdir: `cwd: ../../..`
// ran the command wherever that landed, up to the filesystem root. The isolation
// every scenario is built on assumes the command runs in its own temp
// directory — `changes:` diffs it, and `dir:` and `file:` read it — so a step
// running outside acted on the host while its assertions examined an untouched
// sandbox, and passed having done none of what it claimed. Every other
// workdir-relative field has rejected the same traversal all along.
//
// An absolute cwd is left alone. It is explicit in a way `../..` is not: the
// author wrote a full path, the way a spec pointing a step at a checked-out
// tree does. key is the fully-qualified field.
func workdirRelativeDir(add addFunc, key, cwd string) {
	// A leading "/" is absolute on every platform a spec is written for, and
	// filepath.IsAbs additionally catches the Windows drive form.
	if cwd == "" || strings.HasPrefix(cwd, "/") || filepath.IsAbs(cwd) {
		return
	}
	if pathEscapesWorkdir(filepath.ToSlash(cwd)) {
		add(diag.PathEscapesWorkdir, "%s %q escapes the scenario workdir (no ../ traversal)", key, cwd)
	}
}

// positiveDuration validates a duration knob whose zero value is meaningless
// because omitting the key already selects a documented default; def names
// that default in the failure message.
func positiveDuration(add addFunc, key, val, example, def string) {
	if val == "" {
		return
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		add(diag.BadDuration, "%s %q is not a valid duration (e.g. %q)", key, val, example)
		return
	}
	if d <= 0 {
		add(diag.NonPositiveValue, "%s must be positive (got %q); omit it for the %s default", key, val, def)
	}
}

// validateExpectFail checks a declared known bug (#395).
//
// `reason` is required, and that is the whole rule: an expected failure with no
// stated reason is indistinguishable from a test somebody gave up on, and the
// next reader has no way to tell whether the scenario documents a real defect
// or is dead weight. An `issue` URL is optional but is what makes an XPASS
// actionable, so its absence is worth nothing more than the missing link.
func validateExpectFail(add addFunc, where string, ef *spec.ExpectFail) {
	if ef == nil {
		return
	}
	if strings.TrimSpace(ef.Reason) == "" {
		add(diag.RequiredKey, "%s.expect_fail.reason is required: say what is broken, or a reader cannot tell a documented known bug from a test that was given up on", where)
	}
}
