package spec

import (
	"reflect"
	"sort"
	"testing"
)

// visitedStrings runs walk with a recording visitor and returns every distinct
// non-empty string it was shown, sorted.
func visitedStrings(walk func(visit func(string) string)) []string {
	seen := map[string]bool{}
	walk(func(s string) string {
		if s != "" {
			seen[s] = true
		}
		return s
	})
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// TestWalkStepStrings_VisitsEveryExpandedField pins each walker's field list.
// The walkers are the ONE list of what a step kind ${name}-expands — the
// engine expands through them and the summaries collect through them — so
// widening or narrowing a kind's expansion surface must show up here as an
// explicit diff of this table, reviewed once, instead of as two hand lists
// drifting apart the way run's env, http's body_file/body_to/form/files, cdp's
// upload/download, and pty's paste/exec each did.
func TestWalkStepStrings_VisitsEveryExpandedField(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		walk func(visit func(string) string)
		want []string
	}{
		{
			"fixture",
			func(v func(string) string) {
				WalkFixtureStrings(&Fixture{File: "file", Content: "content", From: "from", Symlink: "symlink"}, v)
			},
			[]string{"content", "file", "from", "symlink"},
		},
		{
			"run",
			func(v func(string) string) {
				WalkRunStrings(&Run{
					Command:  "command",
					Cwd:      "cwd",
					Stdin:    Stdin{Inline: "stdin-inline", File: "stdin-file", Base64: "c3RkaW4="},
					StdoutTo: "stdout-to",
					StderrTo: "stderr-to",
					Env:      map[string]string{"K": "env-value"},
				}, v)
			},
			// Base64 is binary and stays byte-exact; env keys are names.
			[]string{"command", "cwd", "env-value", "stderr-to", "stdin-file", "stdin-inline", "stdout-to"},
		},
		{
			"service",
			func(v func(string) string) {
				WalkServiceStrings(&Service{
					Name:    "name-is-vocabulary",
					Command: "command",
					Cwd:     "cwd",
					Env:     map[string]string{"K": "env-value"},
					Ready:   &Ready{File: "ready-file", Port: "ready-port", Log: "ready-log", Delay: "1s"},
				}, v)
			},
			// The name identifies the service; delay is a duration.
			[]string{"command", "cwd", "env-value", "ready-file", "ready-log", "ready-port"},
		},
		{
			"store",
			func(v func(string) string) {
				WalkStoreStrings(&Store{
					Name: "name-is-vocabulary",
					From: &StoreFrom{File: &FileAssert{Path: "file-path", Text: boolp(true)}},
				}, v)
			},
			// Only the file source path is expanded; selectors match output.
			[]string{"file-path"},
		},
		{
			"http",
			func(v func(string) string) {
				WalkHTTPStrings(&HTTP{
					Runner:   "runner-is-vocabulary",
					Method:   "GET",
					Path:     "path",
					Header:   map[string]string{"K": "header-value"},
					JSON:     map[string]any{"k": "json-leaf"},
					Body:     "body",
					BodyFile: "body-file",
					BodyTo:   "body-to",
					Form:     map[string]string{"K": "form-value"},
					Files:    []FilePart{{Field: "field-is-vocabulary", Path: "file-part-path"}},
				}, v)
			},
			// The method and runner select behavior; a multipart field names a
			// form slot. Retry's until is expanded per attempt by the poll loop.
			[]string{"body", "body-file", "body-to", "file-part-path", "form-value", "header-value", "json-leaf", "path"},
		},
		{
			"query",
			func(v func(string) string) {
				WalkQueryStrings(&Query{Runner: "runner-is-vocabulary", SQL: "sql"}, v)
			},
			[]string{"sql"},
		},
		{
			"grpc",
			func(v func(string) string) {
				WalkGRPCStrings(&GRPC{
					Runner: "runner-is-vocabulary",
					Method: "method",
					Header: map[string]string{"K": "header-value"},
					JSON:   map[string]any{"k": "json-leaf"},
				}, v)
			},
			[]string{"header-value", "json-leaf", "method"},
		},
		{
			"cdp",
			func(v func(string) string) {
				WalkCDPStrings(&CDP{Runner: "runner-is-vocabulary", Actions: []CDPAction{
					{Navigate: "navigate"},
					{WaitVisible: "wait-visible"},
					{WaitHidden: "wait-hidden"},
					{Click: "click"},
					{Check: "check"},
					{Uncheck: "uncheck"},
					{Text: "text"},
					{Eval: "eval"},
					{SendKeys: &CDPSendKeys{Selector: "sendkeys-selector", Value: "sendkeys-value"}},
					{Press: &CDPPress{Selector: "press-selector", Key: "press-key"}},
					{Select: &CDPSelect{Selector: "select-selector", Value: "select-value"}},
					{Screenshot: &CDPScreenshot{Path: "screenshot-path", Selector: "screenshot-selector"}},
					{Attribute: &CDPAttribute{Selector: "attribute-selector", Name: "attribute-name"}},
					{Upload: &CDPUpload{Selector: "upload-selector", File: "upload-file"}},
					{Download: &CDPDownload{Click: "download-click", Dir: "download-dir"}},
				}}, v)
			},
			[]string{
				"attribute-name", "attribute-selector", "check", "click",
				"download-click", "download-dir", "eval", "navigate",
				"press-key", "press-selector", "screenshot-path", "screenshot-selector",
				"select-selector", "select-value", "sendkeys-selector", "sendkeys-value",
				"text", "uncheck", "upload-file", "upload-selector",
				"wait-hidden", "wait-visible",
			},
		},
		{
			"pty",
			func(v func(string) string) {
				text := "send-text"
				paste := "send-paste"
				WalkPTYStrings(&PTY{
					Command: "command",
					Cwd:     "cwd",
					Env:     map[string]string{"K": "env-value"},
					Session: []PTYAction{
						{Expect: "expect"},
						{Send: &PTYSend{Text: &text}},
						{Send: &PTYSend{Paste: &paste}},
						{Send: &PTYSend{Key: "key-is-vocabulary"}},
						{Exec: &PTYExec{Command: "exec-command"}},
						{Resize: &PTYResize{Rows: 24, Cols: 80}},
						{ExpectScreen: &PTYExpectScreen{ScreenAssert: ScreenAssert{StreamAssert: StreamAssert{Contains: StringList{"screen-contains"}}}}},
					},
				}, v)
			},
			// Named keys are fixed byte sequences; a resize carries integers.
			[]string{"command", "cwd", "env-value", "exec-command", "expect", "screen-contains", "send-paste", "send-text"},
		},
		{
			"signal",
			func(v func(string) string) {
				WalkSignalStrings(&Signal{Service: "service", Signal: "TERM", Wait: &SignalWait{Timeout: "1s"}}, v)
			},
			// The signal name is vocabulary and the wait timeout a duration.
			[]string{"service"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := visitedStrings(c.walk); !reflect.DeepEqual(got, c.want) {
				t.Errorf("visited = %v\nwant      %v\n(a difference here is a change to what the kind expands: make it on purpose, in the walker, and update this list)", got, c.want)
			}
		})
	}
}

// TestWalkStepStrings_IdentityKeepsEverything pins the copy-first rule: with an
// identity visitor, a walker's output is deeply equal to its fully-populated
// input. A walker that rebuilt its copy from the zero value would drop the
// fields it does not know — the pty expansion once dropped a newly added
// resize entry exactly that way — and this fails the moment it regresses.
func TestWalkStepStrings_IdentityKeepsEverything(t *testing.T) {
	t.Parallel()
	text := "t"
	paste := "p"
	identity := func(s string) string { return s }
	cases := []struct {
		name string
		got  any
		want any
	}{
		{"fixture", WalkFixtureStrings(&Fixture{File: "f", Content: "c", From: "fr", Symlink: "s", Mode: "0644"}, identity),
			&Fixture{File: "f", Content: "c", From: "fr", Symlink: "s", Mode: "0644"}},
		{"run", WalkRunStrings(&Run{Command: "c", Shell: boolp(true), Cwd: "d", Timeout: "1s", Env: map[string]string{"K": "v"}, Retry: &Retry{Times: 2}}, identity),
			&Run{Command: "c", Shell: boolp(true), Cwd: "d", Timeout: "1s", Env: map[string]string{"K": "v"}, Retry: &Retry{Times: 2}}},
		{"pty action", func() any {
			a := WalkPTYActionStrings(PTYAction{Send: &PTYSend{Text: &text, Times: 3}, Resize: &PTYResize{Rows: 24, Cols: 80}}, identity)
			return &a
		}(),
			&PTYAction{Send: &PTYSend{Text: &text, Times: 3}, Resize: &PTYResize{Rows: 24, Cols: 80}}},
		{"pty paste action", func() any {
			a := WalkPTYActionStrings(PTYAction{Send: &PTYSend{Paste: &paste}}, identity)
			return &a
		}(),
			&PTYAction{Send: &PTYSend{Paste: &paste}}},
		{"signal", WalkSignalStrings(&Signal{Service: "s", Signal: "TERM", Wait: &SignalWait{Timeout: "1s"}}, identity),
			&Signal{Service: "s", Signal: "TERM", Wait: &SignalWait{Timeout: "1s"}}},
		{"service", WalkServiceStrings(&Service{Name: "n", Command: "c", Cwd: "d", Env: map[string]string{"K": "v"}, Ready: &Ready{File: "f", Port: "p", Log: "l", Delay: "1s", Store: "sv", Timeout: "2s"}}, identity),
			&Service{Name: "n", Command: "c", Cwd: "d", Env: map[string]string{"K": "v"}, Ready: &Ready{File: "f", Port: "p", Log: "l", Delay: "1s", Store: "sv", Timeout: "2s"}}},
		{"store", WalkStoreStrings(&Store{Name: "n", From: &StoreFrom{File: &FileAssert{Path: "p", Text: boolp(true)}}}, identity),
			&Store{Name: "n", From: &StoreFrom{File: &FileAssert{Path: "p", Text: boolp(true)}}}},
		{"http", WalkHTTPStrings(&HTTP{Runner: "r", Method: "GET", Path: "p", Header: map[string]string{"K": "v"}, Body: "b", BodyFile: "bf", BodyTo: "bt", Form: map[string]string{"K": "v"}, Files: []FilePart{{Field: "f", Path: "fp"}}, Retry: &Retry{Times: 2}}, identity),
			&HTTP{Runner: "r", Method: "GET", Path: "p", Header: map[string]string{"K": "v"}, Body: "b", BodyFile: "bf", BodyTo: "bt", Form: map[string]string{"K": "v"}, Files: []FilePart{{Field: "f", Path: "fp"}}, Retry: &Retry{Times: 2}}},
		{"grpc", WalkGRPCStrings(&GRPC{Runner: "r", Method: "m", Header: map[string]string{"K": "v"}, JSON: map[string]any{"k": "v"}}, identity),
			&GRPC{Runner: "r", Method: "m", Header: map[string]string{"K": "v"}, JSON: map[string]any{"k": "v"}}},
		{"cdp", WalkCDPStrings(&CDP{Runner: "r", Actions: []CDPAction{{Title: true}, {Screenshot: &CDPScreenshot{Path: "p", Selector: "s"}}}}, identity),
			&CDP{Runner: "r", Actions: []CDPAction{{Title: true}, {Screenshot: &CDPScreenshot{Path: "p", Selector: "s"}}}}},
		{"pty", WalkPTYStrings(&PTY{Command: "c", Cwd: "d", Rows: 24, Cols: 80, Timeout: "5s", Env: map[string]string{"K": "v"}, Session: []PTYAction{{Resize: &PTYResize{Rows: 10, Cols: 20}}}}, identity),
			&PTY{Command: "c", Cwd: "d", Rows: 24, Cols: 80, Timeout: "5s", Env: map[string]string{"K": "v"}, Session: []PTYAction{{Resize: &PTYResize{Rows: 10, Cols: 20}}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if !reflect.DeepEqual(c.got, c.want) {
				t.Errorf("identity walk changed or dropped fields:\ngot  %+v\nwant %+v", c.got, c.want)
			}
		})
	}
}
