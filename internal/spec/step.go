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

// SetKeys returns the action keys that are present on the step. A valid step has
// exactly one; the loader uses this to enforce the one-of rule.
func (s *Step) SetKeys() []StepKind {
	var keys []StepKind
	if s.Fixture != nil {
		keys = append(keys, StepFixture)
	}
	if s.Run != nil {
		keys = append(keys, StepRun)
	}
	if s.HTTP != nil {
		keys = append(keys, StepHTTP)
	}
	if s.Query != nil {
		keys = append(keys, StepQuery)
	}
	if s.GRPC != nil {
		keys = append(keys, StepGRPC)
	}
	if s.CDP != nil {
		keys = append(keys, StepCDP)
	}
	if s.Assert != nil {
		keys = append(keys, StepAssert)
	}
	if s.Store != nil {
		keys = append(keys, StepStore)
	}
	if s.Service != nil {
		keys = append(keys, StepService)
	}
	if s.PTY != nil {
		keys = append(keys, StepPTY)
	}
	if s.Signal != nil {
		keys = append(keys, StepSignal)
	}
	if s.MockServer != nil {
		keys = append(keys, StepMockServer)
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
)

// SetTargets returns the assertion target families present. A valid assert has
// exactly one.
func (a *Assert) SetTargets() []AssertTarget {
	var t []AssertTarget
	if a.ExitCode != nil {
		t = append(t, AssertExitCode)
	}
	if a.Stdout != nil {
		t = append(t, AssertStdout)
	}
	if a.Stderr != nil {
		t = append(t, AssertStderr)
	}
	if a.File != nil {
		t = append(t, AssertFile)
	}
	if a.Status != nil {
		t = append(t, AssertStatus)
	}
	if a.Header != nil {
		t = append(t, AssertHeader)
	}
	if a.Body != nil {
		t = append(t, AssertBody)
	}
	if a.Rows != nil {
		t = append(t, AssertRows)
	}
	if a.GRPCStatus != nil {
		t = append(t, AssertGRPCStatus)
	}
	if a.Message != nil {
		t = append(t, AssertMessage)
	}
	if a.Value != nil {
		t = append(t, AssertValue)
	}
	if a.Image != nil {
		t = append(t, AssertImage)
	}
	if a.Dir != nil {
		t = append(t, AssertDir)
	}
	if a.PDF != nil {
		t = append(t, AssertPDF)
	}
	if a.Mock != nil {
		t = append(t, AssertMock)
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
