package spec

// StepKind names the single action a Step carries.
type StepKind string

const (
	StepNone       StepKind = ""
	StepFixture    StepKind = "fixture"
	StepRun        StepKind = "run"
	StepHTTP       StepKind = "http"
	StepQuery      StepKind = "query"
	StepGRPC       StepKind = "grpc"
	StepCDP        StepKind = "cdp"
	StepAssert     StepKind = "assert"
	StepStore      StepKind = "store"
	StepService    StepKind = "service"
	StepPTY        StepKind = "pty"
	StepSignal     StepKind = "signal"
	StepMockServer StepKind = "mock_server"
)

// stepActions is the one list of step action kinds: the key each carries in a
// spec, paired with the question "does this Step set it?". It mirrors
// assertTargets below, and exists for the same reason: many layers switch on a
// step's kind — the loader's validation, the variable walk, the security and
// artifact summaries, the doc/explain/manifest renderers — and each hand-written
// switch silently skips a kind it does not know. Deriving SetKeys/Kind and
// AllStepKinds from this table makes it the anchor the coverage tests pull on:
// adding a field to Step without an entry here fails
// TestStepActions_CoverEveryStepField, and a walker that never learned the new
// kind fails its own coverage test against AllStepKinds.
var stepActions = []struct {
	kind StepKind
	set  func(*Step) bool
}{
	{StepFixture, func(s *Step) bool { return s.Fixture != nil }},
	{StepRun, func(s *Step) bool { return s.Run != nil }},
	{StepHTTP, func(s *Step) bool { return s.HTTP != nil }},
	{StepQuery, func(s *Step) bool { return s.Query != nil }},
	{StepGRPC, func(s *Step) bool { return s.GRPC != nil }},
	{StepCDP, func(s *Step) bool { return s.CDP != nil }},
	{StepAssert, func(s *Step) bool { return s.Assert != nil }},
	{StepStore, func(s *Step) bool { return s.Store != nil }},
	{StepService, func(s *Step) bool { return s.Service != nil }},
	{StepPTY, func(s *Step) bool { return s.PTY != nil }},
	{StepSignal, func(s *Step) bool { return s.Signal != nil }},
	{StepMockServer, func(s *Step) bool { return s.MockServer != nil }},
}

// AllStepKinds returns every step action kind, in SetKeys order. It exists so a
// layer that must handle all of them can be tested against the list instead of
// against a reader's memory — the step-kind twin of AllAssertTargets.
func AllStepKinds() []StepKind {
	out := make([]StepKind, len(stepActions))
	for i, e := range stepActions {
		out[i] = e.kind
	}
	return out
}

// SetKeys returns the action keys that are present on the step. A valid step has
// exactly one; the loader uses this to enforce the one-of rule.
func (s *Step) SetKeys() []StepKind {
	var keys []StepKind
	for _, e := range stepActions {
		if e.set(s) {
			keys = append(keys, e.kind)
		}
	}
	return keys
}

// Kind reports the step's single action kind, or StepNone if not exactly one.
func (s *Step) Kind() StepKind {
	keys := s.SetKeys()
	if len(keys) != 1 {
		return StepNone
	}
	return keys[0]
}

// AssertTarget names a single assertion target family.
type AssertTarget string

const (
	AssertNone       AssertTarget = ""
	AssertExitCode   AssertTarget = "exit_code"
	AssertStdout     AssertTarget = "stdout"
	AssertStderr     AssertTarget = "stderr"
	AssertFile       AssertTarget = "file"
	AssertStatus     AssertTarget = "status"
	AssertHeader     AssertTarget = "header"
	AssertBody       AssertTarget = "body"
	AssertRows       AssertTarget = "rows"
	AssertGRPCStatus AssertTarget = "grpc_status"
	AssertMessage    AssertTarget = "message"
	AssertValue      AssertTarget = "value"
	AssertImage      AssertTarget = "image"
	AssertDir        AssertTarget = "dir"
	AssertPDF        AssertTarget = "pdf"
	AssertMock       AssertTarget = "mock"
	AssertScreen     AssertTarget = "screen"
	AssertDuration   AssertTarget = "duration"
	AssertChanges    AssertTarget = "changes"
)

// assertTargets is the one list of assertion target families: the name each
// carries in a spec, in the order SetTargets reports them, paired with the
// question "does this Assert set it?". Five layers switch on a target — the
// loader's validation, the runtime check, the doc and explain descriptions, and
// the JSON schema — and each of those switches is drift-tested against this
// list, so a target cannot be half-wired: adding a field to Assert without an
// entry here fails TestAssertTargets_CoverEveryAssertField, and a layer that
// forgets the new target fails its own coverage test.
var assertTargets = []struct {
	target AssertTarget
	set    func(*Assert) bool
}{
	{AssertExitCode, func(a *Assert) bool { return a.ExitCode != nil }},
	{AssertStdout, func(a *Assert) bool { return a.Stdout != nil }},
	{AssertStderr, func(a *Assert) bool { return a.Stderr != nil }},
	{AssertFile, func(a *Assert) bool { return a.File != nil }},
	{AssertStatus, func(a *Assert) bool { return a.Status != nil }},
	{AssertHeader, func(a *Assert) bool { return a.Header != nil }},
	{AssertBody, func(a *Assert) bool { return a.Body != nil }},
	{AssertRows, func(a *Assert) bool { return a.Rows != nil }},
	{AssertGRPCStatus, func(a *Assert) bool { return a.GRPCStatus != nil }},
	{AssertMessage, func(a *Assert) bool { return a.Message != nil }},
	{AssertValue, func(a *Assert) bool { return a.Value != nil }},
	{AssertImage, func(a *Assert) bool { return a.Image != nil }},
	{AssertDir, func(a *Assert) bool { return a.Dir != nil }},
	{AssertPDF, func(a *Assert) bool { return a.PDF != nil }},
	{AssertMock, func(a *Assert) bool { return a.Mock != nil }},
	{AssertScreen, func(a *Assert) bool { return a.Screen != nil }},
	{AssertDuration, func(a *Assert) bool { return a.Duration != nil }},
	{AssertChanges, func(a *Assert) bool { return a.Changes != nil }},
}

// AllAssertTargets returns every assertion target family, in the order
// SetTargets reports them. It exists so a layer that must handle all of them can
// be tested against the list instead of against a reader's memory.
func AllAssertTargets() []AssertTarget {
	out := make([]AssertTarget, len(assertTargets))
	for i, e := range assertTargets {
		out[i] = e.target
	}
	return out
}

// SetTargets returns the assertion target families present. A valid assert has
// exactly one.
func (a *Assert) SetTargets() []AssertTarget {
	var t []AssertTarget
	for _, e := range assertTargets {
		if e.set(a) {
			t = append(t, e.target)
		}
	}
	return t
}

// SetMatchers returns the names of matchers present on a StreamAssert. A valid
// stream assertion has exactly one.
func (s *StreamAssert) SetMatchers() []string {
	var m []string
	if s.Empty != nil {
		m = append(m, "empty")
	}
	if s.Contains != nil {
		m = append(m, "contains")
	}
	if s.NotContains != nil {
		m = append(m, "not_contains")
	}
	if s.Matches != nil {
		m = append(m, "matches")
	}
	if s.NotMatches != nil {
		m = append(m, "not_matches")
	}
	if s.Equals != nil {
		m = append(m, "equals")
	}
	if s.NotEquals != nil {
		m = append(m, "not_equals")
	}
	if s.JSON != nil {
		m = append(m, "json")
	}
	if s.YAML != nil {
		m = append(m, "yaml")
	}
	if s.Snapshot != "" {
		m = append(m, "snapshot")
	}
	return m
}

// CDPActionLabel returns a short, canonical label for one browser action (#50).
// explain, doc, and manifest all render CDP actions through this single helper so
// the human- and machine-facing summaries never drift from the runtime action
// set. The label names the action verb and, where useful, its selector/target.
func CDPActionLabel(a CDPAction) string {
	switch {
	case a.Navigate != "":
		return "navigate " + a.Navigate
	case a.WaitVisible != "":
		return "wait_visible " + a.WaitVisible
	case a.WaitHidden != "":
		return "wait_hidden " + a.WaitHidden
	case a.Click != "":
		return "click " + a.Click
	case a.Press != nil:
		return "press " + a.Press.Key + " on " + a.Press.Selector
	case a.Select != nil:
		return "select " + a.Select.Value + " in " + a.Select.Selector
	case a.Check != "":
		return "check " + a.Check
	case a.Uncheck != "":
		return "uncheck " + a.Uncheck
	case a.Screenshot != nil:
		return "screenshot " + a.Screenshot.Path
	case a.SendKeys != nil:
		return "send_keys " + a.SendKeys.Selector
	case a.Text != "":
		return "text " + a.Text
	case a.Title:
		return "title"
	case a.Attribute != nil:
		return "attribute " + a.Attribute.Name + " of " + a.Attribute.Selector
	case a.Eval != "":
		return "eval"
	case a.Upload != nil:
		return "upload " + a.Upload.File + " to " + a.Upload.Selector
	case a.Download != nil:
		return "download via " + a.Download.Click
	default:
		return "(unknown action)"
	}
}
