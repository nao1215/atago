package loader

import (
	"encoding/base64"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/runner/db"
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

	// Suite services are legal signal targets from any scenario (#23).
	suiteServiceNames := map[string]bool{}
	for i := range s.Suite.Setup {
		if svc := s.Suite.Setup[i].Service; svc != nil && svc.Name != "" {
			suiteServiceNames[svc.Name] = true
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
		if len(sc.Steps) == 0 {
			add("%s: steps must contain at least one step", where)
			continue
		}
		for j := range sc.Steps {
			validateStep(add, fmt.Sprintf("%s.steps[%d]", where, j), &sc.Steps[j], s.Runners, serviceNames)
		}
		for j := range sc.Teardown {
			validateStep(add, fmt.Sprintf("%s.teardown[%d]", where, j), &sc.Teardown[j], s.Runners, serviceNames)
		}
	}
	return errs
}

// validateDefaults checks the top-level `defaults:` block (ADR-0039). The merge only
// covers non-identity, non-per-step fields, so a value the loader would silently
// ignore is reported here instead. Fields the loader does merge are validated on
// the concrete elements after applyDefaults (and, for a shared readiness probe,
// here too, so a wrong probe fails even when no scenario declares a service).
// validateSuiteTimeout checks the suite-level default step timeout (#17).
func validateSuiteTimeout(add func(string, ...any), s *spec.Suite) {
	if s.Timeout == "" {
		return
	}
	if _, err := time.ParseDuration(s.Timeout); err != nil {
		add("suite.timeout %q is not a valid duration (e.g. \"2m\"); use \"0\" to disable the built-in default", s.Timeout)
	}
}

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
			if _, err := time.ParseDuration(r.Timeout); err != nil {
				add("defaults.run.timeout %q is not a valid duration (e.g. \"30s\")", r.Timeout)
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

// validateExitCode checks the exit_code assertion (#19): exactly one of the
// bare-int / {not} / {in} forms, and an `in` set that is non-empty with unique
// values — a duplicated code is authoring confusion, not a wider contract.
func validateExitCode(add func(string, ...any), where string, e *spec.ExitCode) {
	set := 0
	if e.Equals != nil {
		set++
	}
	if e.Not != nil {
		set++
	}
	if e.In != nil {
		set++
	}
	if set == 0 {
		add("%s must be an int, {not: int}, or {in: [int, ...]}", where)
		return
	}
	if set > 1 {
		add("%s: set exactly one of a bare int, not, or in", where)
		return
	}
	if e.In != nil {
		if len(e.In) == 0 {
			add("%s.in must list at least one accepted exit code", where)
		}
		seen := make(map[int]bool, len(e.In))
		for _, n := range e.In {
			if seen[n] {
				add("%s.in lists %d more than once", where, n)
			}
			seen[n] = true
		}
	}
}

// validateStdin checks a run step's stdin source (#18): the mapping form must
// set exactly one of file/base64, and a base64 payload must decode — at load
// time, so a typo fails with a positioned message instead of mid-run.
func validateStdin(add func(string, ...any), where string, s spec.Stdin) {
	if s.IsMapping() {
		set := 0
		if s.File != "" {
			set++
		}
		if s.Base64 != "" {
			set++
		}
		if set != 1 {
			add("%s.stdin must set exactly one of file/base64 (or be a plain string for inline text)", where)
		}
	}
	if s.Base64 != "" {
		if _, err := base64.StdEncoding.DecodeString(s.Base64); err != nil {
			add("%s.stdin.base64 is not valid base64: %v", where, err)
		}
	}
}

// validateHermeticEnv checks the clear_env/pass_env pairing (#16): pass_env is
// only meaningful when clear_env starts the environment empty, so an
// allowlist without clear_env: true is authoring confusion and is rejected
// instead of silently ignored. Empty variable names are rejected too.
func validateHermeticEnv(add func(string, ...any), where string, clearEnv *bool, passEnv []string) {
	if len(passEnv) == 0 {
		return
	}
	if clearEnv == nil || !*clearEnv {
		add("%s.pass_env requires clear_env: true (pass_env selects host vars for a cleared environment)", where)
	}
	for i, name := range passEnv {
		if name == "" {
			add("%s.pass_env[%d] must not be an empty variable name", where, i)
		}
	}
}

var validRunnerType = map[string]bool{"cmd": true, "http": true, "db": true, "ssh": true, "grpc": true, "browser": true}

func validateRunners(add func(string, ...any), runners map[string]spec.Runner) {
	for name, r := range runners {
		where := fmt.Sprintf("runner %q", name)
		if r.Type == "" {
			add("%s.type is required", where)
			continue
		}
		if !validRunnerType[r.Type] {
			add("%s.type %q is invalid (want cmd, http, db, ssh, grpc, or browser)", where, r.Type)
			continue
		}
		switch r.Type {
		case "db":
			if r.DSN == "" {
				add("%s (db) requires a dsn", where)
			}
			// A declared driver is authoritative: reject an unsupported value here so
			// a typo fails at load time instead of silently inferring from the dsn.
			if err := db.ValidateDriver(r.Driver); err != nil {
				add("%s: %v", where, err)
			}
		case "ssh":
			if r.Host == "" {
				add("%s (ssh) requires a host", where)
			}
			if r.User == "" {
				add("%s (ssh) requires a user", where)
			}
		case "grpc":
			if r.Target == "" {
				add("%s (grpc) requires a target", where)
			}
		case "browser":
			// no required fields; a browser runner launches a local headless Chrome.
		}
		// timeout is common to every runner type; catch a malformed value here
		// instead of when the first step opens the connection.
		if r.Timeout != "" {
			if _, err := time.ParseDuration(r.Timeout); err != nil {
				add("%s.timeout %q is not a valid duration (e.g. \"30s\")", where, r.Timeout)
			}
		}
		validateRunnerFields(add, where, &r)
	}
}

// runnerFields maps each runner field to the single runner type that owns it, so
// cross-type fields (an http runner with ssh fields, a grpc runner with db
// fields, ...) are rejected instead of silently accepted (#44). type/cwd/timeout
// are common to every runner and intentionally absent here.
func validateRunnerFields(add func(string, ...any), where string, r *spec.Runner) {
	type fieldOwner struct {
		owner string
		set   bool
		field string
	}
	fields := []fieldOwner{
		{"http", r.BaseURL != "", "base_url"},
		{"db", r.DSN != "", "dsn"},
		{"db", r.Driver != "", "driver"},
		{"ssh", r.Host != "", "host"},
		{"ssh", r.User != "", "user"},
		{"ssh", r.Password != "", "password"},
		{"ssh", r.KeyFile != "", "key_file"},
		{"ssh", r.KnownHosts != "", "known_hosts"},
		{"ssh", r.InsecureHostKey, "insecure_host_key"},
		{"grpc", r.Target != "", "target"},
		{"grpc", r.TLS, "tls"},
		{"browser", r.Headless != nil, "headless"},
		{"browser", r.ExecPath != "", "exec_path"},
		{"browser", len(r.BrowserArgs) > 0, "browser_args"},
	}
	for _, f := range fields {
		if f.set && f.owner != r.Type {
			add("%s: %q cannot be set on a %s runner (it is a %s-runner field)", where, f.field, r.Type, f.owner)
		}
	}
}

func validateServices(add func(string, ...any), where string, services []spec.Service) {
	seen := make(map[string]bool, len(services))
	for i := range services {
		svc := &services[i]
		sw := fmt.Sprintf("%s.services[%d]", where, i)
		if svc.Name == "" {
			add("%s.name is required", sw)
		} else {
			if seen[svc.Name] {
				add("%s: duplicate service name %q", where, svc.Name)
			}
			seen[svc.Name] = true
			sw = fmt.Sprintf("%s service %q", where, svc.Name)
		}
		if svc.Command == "" {
			add("%s.command is required", sw)
		}
		validateHermeticEnv(add, sw, svc.ClearEnv, svc.PassEnv)
		validateReady(add, sw, svc.Ready)
	}
}

// validateSuiteBlock checks suite.setup / suite.teardown (#7): steps run once
// per suite in the ${suitedir} scratch dir, so only the suite-scoped kinds are
// allowed — fixture, run, store, assert, and (setup only) `service:`. The
// runner-backed kinds (http/query/grpc/cdp) are per-scenario machinery and are
// rejected with a pointer to where they belong.
func validateSuiteBlock(add func(string, ...any), where string, steps []spec.Step, runners map[string]spec.Runner, allowService bool) {
	seenService := map[string]bool{}
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
			validateRunnerRef(add, sw, "run", st.Run.Runner, runners)
			if st.Run.Timeout != "" {
				if _, err := time.ParseDuration(st.Run.Timeout); err != nil {
					add("%s.run.timeout %q is not a valid duration (e.g. \"30s\")", sw, st.Run.Timeout)
				}
			}
			if st.Run.Command == "" {
				add("%s.run.command is required", sw)
			}
			validateHermeticEnv(add, sw+".run", st.Run.ClearEnv, st.Run.PassEnv)
			validateStdin(add, sw+".run", st.Run.Stdin)
			validateRetry(add, sw+".run", st.Run.Retry)
		case spec.StepStore:
			validateStore(add, sw, st.Store)
		case spec.StepAssert:
			validateAssert(add, sw, st.Assert)
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
		default:
			add("%s: %s steps are per-scenario (they need a scenario workdir and runners); move it into a scenario", sw, st.Kind())
		}
	}
}

func validateReady(add func(string, ...any), where string, r *spec.Ready) {
	if r == nil {
		return
	}
	n := 0
	for _, set := range []bool{r.File != "", r.Port != "", r.Log != "", r.Delay != ""} {
		if set {
			n++
		}
	}
	if n > 1 {
		add("%s.ready: set only one of file/port/log/delay", where)
	}
	if r.Store != "" && r.File == "" {
		add("%s.ready.store requires file (the file whose content is captured)", where)
	}
	for _, d := range []struct {
		key, val string
	}{{"timeout", r.Timeout}, {"delay", r.Delay}} {
		if d.val != "" {
			if _, err := time.ParseDuration(d.val); err != nil {
				add("%s.ready.%s %q is not a valid duration", where, d.key, d.val)
			}
		}
	}
	if r.Log != "" {
		if _, err := regexp.Compile(r.Log); err != nil {
			add("%s.ready.log %q is not a valid regexp: %v", where, r.Log, err)
		}
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

func validateStep(add func(string, ...any), where string, st *spec.Step, runners map[string]spec.Runner, serviceNames map[string]bool) {
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
		validateRunnerRef(add, where, "run", st.Run.Runner, runners)
		if st.Run.Timeout != "" {
			if _, err := time.ParseDuration(st.Run.Timeout); err != nil {
				add("%s.run.timeout %q is not a valid duration (e.g. \"30s\")", where, st.Run.Timeout)
			}
		}
		if st.Run.Command == "" {
			add("%s.run.command is required", where)
		} else if !st.Run.ShellEnabled() {
			// Without shell, the command is tokenized into argv, so shell operators
			// (redirects, pipes, sequencing, substitution) are not honored. Rather
			// than silently pass them as literal argv, flag them with a fix-forward
			// hint (#: shell authoring UX).
			if tok := shellMetachar(st.Run.Command); tok != "" {
				add("%s.run.command contains the shell metacharacter %q but shell is not enabled; "+
					"set `shell: true` to run it through a shell, split it into multiple `run` steps, "+
					"or use `stdout_to` / `stderr_to` for redirection", where, tok)
			}
		}
		validateHermeticEnv(add, where+".run", st.Run.ClearEnv, st.Run.PassEnv)
		validateStdin(add, where+".run", st.Run.Stdin)
		validateRetry(add, where+".run", st.Run.Retry)
	case spec.StepAssert:
		validateAssert(add, where, st.Assert)
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
	case spec.StepPTY:
		validatePTY(add, where, st.PTY)
	case spec.StepSignal:
		validateSignal(add, where, st.Signal, serviceNames)
	}
}

// validateSignal checks a signal step (#23): a declared target service, an
// accepted signal name, and a parseable wait timeout. A ${name}-referencing
// target is resolved at run time and skips the declared-name check.
func validateSignal(add func(string, ...any), where string, sg *spec.Signal, serviceNames map[string]bool) {
	switch {
	case sg.Service == "":
		add("%s.signal.service is required (the scenario or suite service to signal)", where)
	case !strings.Contains(sg.Service, "${") && !serviceNames[sg.Service]:
		declared := "none"
		if len(serviceNames) > 0 {
			declared = strings.Join(slices.Sorted(maps.Keys(serviceNames)), ", ")
		}
		add("%s.signal.service %q is not a declared service (declared: %s)", where, sg.Service, declared)
	}
	switch {
	case sg.Signal == "":
		add("%s.signal.signal is required (TERM, INT, HUP, USR1, USR2, or KILL)", where)
	case !spec.ValidSignalName(sg.Signal):
		add("%s.signal.signal %q is not an accepted signal (TERM, INT, HUP, USR1, USR2, or KILL, with an optional SIG prefix)", where, sg.Signal)
	}
	if sg.Wait != nil && sg.Wait.Timeout != "" {
		d, err := time.ParseDuration(sg.Wait.Timeout)
		switch {
		case err != nil:
			add("%s.signal.wait.timeout %q is not a valid duration (e.g. \"5s\")", where, sg.Wait.Timeout)
		case d <= 0:
			add("%s.signal.wait.timeout must be positive (got %q); omit it for the 5s default", where, sg.Wait.Timeout)
		}
	}
}

// validatePTY checks a pty step (#8): a command, sane duration/size values,
// and a session whose entries each set exactly one of expect/send with
// compilable expect regexps.
func validatePTY(add func(string, ...any), where string, p *spec.PTY) {
	if p.Command == "" {
		add("%s.pty.command is required", where)
	}
	if p.Timeout != "" {
		d, err := time.ParseDuration(p.Timeout)
		switch {
		case err != nil:
			add("%s.pty.timeout %q is not a valid duration (e.g. \"30s\")", where, p.Timeout)
		case d <= 0:
			add("%s.pty.timeout must be positive (got %q); omit it for the 30s default", where, p.Timeout)
		}
	}
	validateHermeticEnv(add, where+".pty", p.ClearEnv, p.PassEnv)
	// A pty size is a uint16 on the wire; reject values the terminal cannot
	// represent instead of silently truncating.
	if p.Rows < 0 || p.Cols < 0 || p.Rows > 65535 || p.Cols > 65535 {
		add("%s.pty: rows/cols must be between 0 and 65535", where)
	}
	for i, a := range p.Session {
		aw := fmt.Sprintf("%s.pty.session[%d]", where, i)
		hasExpect := a.Expect != ""
		hasSend := a.Send != nil
		switch {
		case hasExpect && hasSend:
			add("%s: set exactly one of expect/send (got both)", aw)
		case !hasExpect && !hasSend:
			add("%s: set exactly one of expect/send (an empty send: \"\" transmits EOF)", aw)
		case hasExpect:
			if _, err := regexp.Compile(a.Expect); err != nil {
				add("%s.expect %q is not a valid regexp: %v", aw, a.Expect, err)
			}
		}
	}
}

// validateCDPActions checks that each browser action sets exactly one action key
// and supplies its required sub-fields, so a malformed cdp step fails at load
// time with a clear message rather than mid-run (#50).
func validateCDPActions(add func(string, ...any), where string, actions []spec.CDPAction) {
	for i, a := range actions {
		aw := fmt.Sprintf("%s.cdp.actions[%d]", where, i)
		if n := cdpActionCount(&a); n == 0 {
			add("%s sets no recognized action", aw)
			continue
		} else if n > 1 {
			add("%s sets multiple actions; set exactly one", aw)
		}
		switch {
		case a.Press != nil:
			if a.Press.Selector == "" || a.Press.Key == "" {
				add("%s.press requires selector and key", aw)
			}
		case a.Select != nil:
			if a.Select.Selector == "" {
				add("%s.select requires a selector", aw)
			}
		case a.Screenshot != nil:
			if a.Screenshot.Path == "" {
				add("%s.screenshot requires a path", aw)
			}
		case a.Attribute != nil:
			if a.Attribute.Selector == "" || a.Attribute.Name == "" {
				add("%s.attribute requires selector and name", aw)
			}
		case a.SendKeys != nil:
			if a.SendKeys.Selector == "" {
				add("%s.send_keys requires a selector", aw)
			}
		case a.Upload != nil:
			if a.Upload.Selector == "" || a.Upload.File == "" {
				add("%s.upload requires selector and file", aw)
			}
		case a.Download != nil:
			if a.Download.Click == "" {
				add("%s.download requires a click selector", aw)
			}
		}
	}
}

// cdpActionCount reports how many action keys are set on one browser action.
func cdpActionCount(a *spec.CDPAction) int {
	n := 0
	for _, set := range []bool{
		a.Navigate != "", a.WaitVisible != "", a.WaitHidden != "", a.Click != "",
		a.Press != nil, a.Select != nil, a.Check != "", a.Uncheck != "",
		a.Screenshot != nil, a.Text != "", a.Title, a.Attribute != nil,
		a.Eval != "", a.SendKeys != nil, a.Upload != nil, a.Download != nil,
	} {
		if set {
			n++
		}
	}
	return n
}

func validateStore(add func(string, ...any), where string, s *spec.Store) {
	if s.Name == "" {
		add("%s.store.name is required", where)
	}
	if s.From == nil {
		add("%s.store.from is required", where)
		return
	}
	n := 0
	if s.From.Stdout != nil {
		n++
	}
	if s.From.Body != nil {
		n++
	}
	if s.From.File != nil {
		n++
	}
	if s.From.Header != "" {
		n++
	}
	if s.From.Rows != nil {
		n++
	}
	if s.From.Message != nil {
		n++
	}
	if s.From.Value != nil {
		n++
	}
	switch n {
	case 0:
		add("%s.store.from must set one of stdout/body/file/header/rows/message/value", where)
	case 1:
	default:
		add("%s.store.from must set exactly one source", where)
	}
}

func validateFixture(add func(string, ...any), where string, f *spec.Fixture) {
	if f.File == "" {
		add("%s.fixture.file is required", where)
	}
	n := 0
	if f.Content != "" {
		n++
	}
	if f.Base64 != "" {
		n++
	}
	if f.From != "" {
		n++
	}
	if f.Symlink != "" {
		n++
	}
	if n > 1 {
		add("%s.fixture: set only one of content, base64, from, or symlink", where)
	}
	if f.Symlink != "" && f.Mode != "" {
		add("%s.fixture: mode cannot be applied to a symlink", where)
	}
	if f.Mode != "" {
		if _, err := strconv.ParseUint(f.Mode, 8, 32); err != nil {
			add("%s.fixture.mode %q is not an octal file mode (e.g. \"0444\")", where, f.Mode)
		}
	}
	if f.Mtime != "" {
		if _, err := time.Parse(time.RFC3339, f.Mtime); err != nil {
			add("%s.fixture.mtime %q is not an RFC3339 timestamp (e.g. \"2026-01-02T15:04:05Z\")", where, f.Mtime)
		}
	}
}

func validateAssert(add func(string, ...any), where string, a *spec.Assert) {
	targets := a.SetTargets()
	if len(targets) == 0 {
		add("%s.assert: must set at least one assertion target (got none)", where)
		return
	}
	// Each set target is an independent check and all must hold, so validate every
	// one of them (an assert may combine, e.g., exit_code + stdout + file).
	for _, t := range targets {
		validateAssertTarget(add, where, a, t)
	}
}

// validateAssertTarget checks the shape of a single assertion target family.
func validateAssertTarget(add func(string, ...any), where string, a *spec.Assert, target spec.AssertTarget) {
	switch target {
	case spec.AssertExitCode:
		// Scalar/mapping shape enforced by ExitCode.UnmarshalYAML; the one-of
		// rule and the `in` set contents are semantic checks (#19).
		validateExitCode(add, where+".assert.exit_code", a.ExitCode)
	case spec.AssertStdout:
		validateStream(add, where+".assert.stdout", a.Stdout)
	case spec.AssertStderr:
		validateStream(add, where+".assert.stderr", a.Stderr)
	case spec.AssertFile:
		validateFile(add, where+".assert.file", a.File)
	case spec.AssertBody:
		validateStream(add, where+".assert.body", a.Body)
	case spec.AssertHeader:
		validateHeaderMatch(add, where+".assert.header", a.Header)
	case spec.AssertRows:
		validateStream(add, where+".assert.rows", a.Rows)
	case spec.AssertMessage:
		validateStream(add, where+".assert.message", a.Message)
	case spec.AssertValue:
		validateStream(add, where+".assert.value", a.Value)
	case spec.AssertImage:
		validateImage(add, where+".assert.image", a.Image)
	case spec.AssertDir:
		validateDir(add, where+".assert.dir", a.Dir)
	case spec.AssertPDF:
		validatePDF(add, where+".assert.pdf", a.PDF)
	case spec.AssertGRPCStatus:
		// grpc_status is a bare int; no further shape to validate.
	}
}

// validatePDF checks a PDF assertion (#73): a path plus at least one constraint,
// sane page bounds, known metadata fields, and a well-formed text matcher.
func validatePDF(add func(string, ...any), where string, p *spec.PDFAssert) {
	if p.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	for _, c := range []*int{p.Pages, p.MinPages, p.MaxPages} {
		if c != nil {
			n++
			if *c < 0 {
				add("%s: page counts must be >= 0 (got %d)", where, *c)
			}
		}
	}
	if p.MinPages != nil && p.MaxPages != nil && *p.MinPages > *p.MaxPages {
		add("%s: min_pages %d exceeds max_pages %d", where, *p.MinPages, *p.MaxPages)
	}
	if len(p.Metadata) > 0 {
		n++
		for k := range p.Metadata {
			if !validPDFMetaField[strings.ToLower(k)] {
				add("%s.metadata: unknown field %q (want title/author/subject/keywords/creator/producer)", where, k)
			}
		}
	}
	if p.Text != nil {
		n++
		validateStream(add, where+".text", p.Text)
	}
	if n == 0 {
		add("%s: must set at least one of pages/min_pages/max_pages/metadata/text", where)
	}
}

var validPDFMetaField = map[string]bool{
	"title": true, "author": true, "subject": true,
	"keywords": true, "creator": true, "producer": true,
}

// validateDir checks a directory/tree assertion (#74): a path plus at least one
// constraint, with sane count bounds. Every set field is an independent
// constraint (like image), so no one-of rule applies beyond requiring at least
// one.
func validateDir(add func(string, ...any), where string, d *spec.DirAssert) {
	if d.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if d.Exists != nil {
		n++
	}
	if len(d.Contains) > 0 {
		n++
	}
	if len(d.NotContains) > 0 {
		n++
	}
	for _, c := range []*int{d.Count, d.MinCount, d.MaxCount} {
		if c != nil {
			n++
			if *c < 0 {
				add("%s: counts must be >= 0 (got %d)", where, *c)
			}
		}
	}
	if d.Glob != "" {
		n++
	}
	if d.MinCount != nil && d.MaxCount != nil && *d.MinCount > *d.MaxCount {
		add("%s: min_count %d exceeds max_count %d", where, *d.MinCount, *d.MaxCount)
	}
	if n == 0 {
		add("%s: must set at least one of exists/contains/not_contains/count/min_count/max_count/glob", where)
	}
}

// validateStringList rejects an explicitly-empty contains/not_contains list
// (`contains: []`), which would otherwise decode to a present-but-empty matcher
// that trivially passes. A scalar or non-empty list is accepted.
func validateStringList(add func(string, ...any), where, key string, l spec.StringList) {
	if l != nil && len(l) == 0 {
		add("%s.%s must not be empty", where, key)
	}
}

// shellMetachar returns the first shell metacharacter found outside quotes in a
// command, or "" if none. It backs the `shell: false` guard: with shell off the
// command is tokenized into argv, so operators like `>` or `|` would be passed
// as literal arguments instead of doing what the author expects. Single- and
// double-quoted regions are skipped so a quoted `">"` argument is not flagged.
func shellMetachar(cmd string) string {
	var quote rune // 0, '\'' or '"' when inside a quoted region
	runes := []rune(cmd)
	for i := range len(runes) {
		c := runes[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			quote = c
		case '`':
			return "`"
		case '|':
			if i+1 < len(runes) && runes[i+1] == '|' {
				return "||"
			}
			return "|"
		case '&':
			if i+1 < len(runes) && runes[i+1] == '&' {
				return "&&"
			}
			// A lone `&` (background) is not in the guarded set; ignore it.
		case ';':
			return ";"
		case '<':
			return "<"
		case '>':
			if i+1 < len(runes) && runes[i+1] == '>' {
				return ">>"
			}
			return ">"
		case '$':
			if i+1 < len(runes) && runes[i+1] == '(' {
				return "$("
			}
		}
	}
	return ""
}

func validateHeaderMatch(add func(string, ...any), where string, h *spec.HeaderMatch) {
	if h.Name == "" {
		add("%s.name is required", where)
	}
	n := 0
	if h.Contains != nil {
		n++
	}
	if h.Equals != nil {
		n++
	}
	switch n {
	case 0:
		add("%s: must set one of contains/equals", where)
	case 1:
	default:
		add("%s: must set exactly one of contains/equals", where)
	}
}

func validateStream(add func(string, ...any), where string, s *spec.StreamAssert) {
	matchers := s.SetMatchers()
	switch len(matchers) {
	case 0:
		add("%s: must set exactly one matcher (empty/contains/not_contains/matches/not_matches/equals/not_equals/json/yaml/snapshot)", where)
	case 1:
		if s.JSON != nil {
			validateJSON(add, where+".json", s.JSON)
		}
		if s.YAML != nil {
			validateJSON(add, where+".yaml", s.YAML)
		}
		if s.Matches != nil {
			if _, err := regexp.Compile(*s.Matches); err != nil {
				add("%s.matches %q is not a valid regexp: %v", where, *s.Matches, err)
			}
		}
		if s.NotMatches != nil {
			if _, err := regexp.Compile(*s.NotMatches); err != nil {
				add("%s.not_matches %q is not a valid regexp: %v", where, *s.NotMatches, err)
			}
		}
	default:
		add("%s: must set exactly one matcher, but set %v", where, matchers)
	}
	validateStringList(add, where, "contains", s.Contains)
	validateStringList(add, where, "not_contains", s.NotContains)

	if s.Line != nil {
		if *s.Line < 1 {
			add("%s.line must be >= 1 (got %d)", where, *s.Line)
		}
		// line narrows the stream to one line, so it only composes with the
		// text matchers — json/snapshot operate on the whole document.
		if s.JSON != nil || s.YAML != nil || s.Snapshot != "" {
			add("%s.line cannot be combined with json/yaml/snapshot (use contains/matches/equals/empty)", where)
		}
	}
}

func validateFile(add func(string, ...any), where string, f *spec.FileAssert) {
	if f.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if f.Exists != nil {
		n++
	}
	if f.Contains != nil {
		n++
		validateStringList(add, where, "contains", f.Contains)
	}
	if f.NotContains != nil {
		n++
		validateStringList(add, where, "not_contains", f.NotContains)
	}
	if f.Executable != nil {
		n++
	}
	if f.JSON != nil {
		n++
		validateJSON(add, where+".json", f.JSON)
	}
	if f.Snapshot != "" {
		n++
	}
	if n == 0 {
		add("%s: must set one of exists/contains/not_contains/executable/json/snapshot", where)
	} else if n > 1 {
		add("%s: must set exactly one of exists/contains/not_contains/executable/json/snapshot", where)
	}
}

var validImageFormat = map[string]bool{
	"png": true, "jpeg": true, "gif": true, "webp": true,
	"bmp": true, "tiff": true, "avif": true, "svg": true,
}

func validateImage(add func(string, ...any), where string, im *spec.ImageAssert) {
	if im.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if im.Format != "" {
		n++
		if !validImageFormat[im.Format] {
			add("%s.format %q is invalid (want png/jpeg/gif/webp/bmp/tiff/avif/svg)", where, im.Format)
		}
	}
	for _, d := range []*int{im.Width, im.Height, im.MinWidth, im.MaxWidth, im.MinHeight, im.MaxHeight} {
		if d != nil {
			n++
			if *d < 0 {
				add("%s: dimensions must be >= 0 (got %d)", where, *d)
			}
		}
	}
	if im.MinWidth != nil && im.MaxWidth != nil && *im.MinWidth > *im.MaxWidth {
		add("%s: min_width %d exceeds max_width %d", where, *im.MinWidth, *im.MaxWidth)
	}
	if im.MinHeight != nil && im.MaxHeight != nil && *im.MinHeight > *im.MaxHeight {
		add("%s: min_height %d exceeds max_height %d", where, *im.MinHeight, *im.MaxHeight)
	}
	if im.Alpha != nil {
		n++
	}
	if im.SimilarTo != "" {
		n++
	}
	if im.MaxDiff != nil {
		if im.SimilarTo == "" {
			add("%s.max_diff requires similar_to", where)
		}
		if *im.MaxDiff < 0 || *im.MaxDiff > 1 {
			add("%s.max_diff must be between 0 and 1 (got %g)", where, *im.MaxDiff)
		}
	}
	if n == 0 {
		add("%s: must set at least one of format/width/height/min_width/max_width/min_height/max_height/alpha/similar_to", where)
	}
	// AVIF and SVG cannot be decoded in pure Go, so only their format can be
	// asserted; reject measurement constraints up front for a clear error.
	if im.Format == "avif" || im.Format == "svg" {
		measures := im.Width != nil || im.Height != nil ||
			im.MinWidth != nil || im.MaxWidth != nil ||
			im.MinHeight != nil || im.MaxHeight != nil ||
			im.Alpha != nil || im.SimilarTo != ""
		if measures {
			add("%s: format %q cannot be measured (only format may be asserted for avif/svg)", where, im.Format)
		}
	}
}

// validateHTTPPayload enforces "a request has one payload": json, body,
// body_file, and form/files are mutually exclusive families (form and files
// combine into one multipart request, so they count as a single family).
func validateHTTPPayload(add func(string, ...any), where string, h *spec.HTTP) {
	var set []string
	if h.JSON != nil {
		set = append(set, "json")
	}
	if h.Body != "" {
		set = append(set, "body")
	}
	if h.BodyFile != "" {
		set = append(set, "body_file")
	}
	if len(h.Form) > 0 || len(h.Files) > 0 {
		set = append(set, "form/files")
	}
	if len(set) > 1 {
		add("%s.http sets %s; a request has one payload — use json for a structured value, body for raw text, body_file for a file's raw content, or form (+ files) for a form submission", where, strings.Join(set, " and "))
	}
}

// validateRetry validates a retry block; where already names the owning step
// action (".run" or ".http") so messages read e.g. "steps[0].http.retry.times".
func validateRetry(add func(string, ...any), where string, r *spec.Retry) {
	if r == nil {
		return
	}
	if r.Times < 1 {
		add("%s.retry.times must be >= 1 (got %d)", where, r.Times)
	}
	if r.Interval != "" {
		if _, err := time.ParseDuration(r.Interval); err != nil {
			add("%s.retry.interval %q is not a valid duration", where, r.Interval)
		}
	}
	if r.Until == nil {
		add("%s.retry.until is required", where)
		return
	}
	validateAssert(add, where+".retry.until", r.Until)
}

func validateJSON(add func(string, ...any), where string, j *spec.JSONAssert) {
	if j.Path == "" {
		add("%s.path is required", where)
	}
	n := 0
	if j.Equals != nil {
		n++
	}
	if j.Matches != nil {
		n++
		if _, err := regexp.Compile(*j.Matches); err != nil {
			add("%s.matches %q is not a valid regexp: %v", where, *j.Matches, err)
		}
	}
	if j.Length != nil {
		n++
	}
	if j.Gt != nil {
		n++
	}
	if j.Gte != nil {
		n++
	}
	if j.Lt != nil {
		n++
	}
	if j.Lte != nil {
		n++
	}
	if n == 0 {
		add("%s: must set one of equals/matches/length/gt/gte/lt/lte", where)
	} else if n > 1 {
		add("%s: must set exactly one of equals/matches/length/gt/gte/lt/lte", where)
	}
}
