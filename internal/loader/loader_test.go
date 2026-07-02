package loader

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadBytes_Valid(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
      - assert:
          exit_code:
            not: 1
      - assert:
          stdout:
            line: 1
            equals: hi
      - assert:
          stderr:
            not_contains: boom
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	if s.Suite.Name != "sample" {
		t.Errorf("suite name = %q, want sample", s.Suite.Name)
	}
	if len(s.Scenarios) != 1 || len(s.Scenarios[0].Steps) != 5 {
		t.Fatalf("unexpected scenario/steps shape: %+v", s.Scenarios)
	}
	if ln := s.Scenarios[0].Steps[3].Assert.Stdout.Line; ln == nil || *ln != 1 {
		t.Errorf("line selector not decoded: %+v", s.Scenarios[0].Steps[3].Assert.Stdout)
	}
	ec := s.Scenarios[0].Steps[1].Assert.ExitCode
	if ec.Equals == nil || *ec.Equals != 0 {
		t.Errorf("exit_code scalar not decoded: %+v", ec)
	}
	notEC := s.Scenarios[0].Steps[2].Assert.ExitCode
	if notEC.Not == nil || *notEC.Not != 1 {
		t.Errorf("exit_code {not:1} not decoded: %+v", notEC)
	}
}

// TestLoadBytes_BrowserRunnerConfig proves the minimal browser-runner
// configuration surface loads and round-trips: headless, exec_path, browser_args.
func TestLoadBytes_BrowserRunnerConfig(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
runners:
  web:
    type: browser
    headless: false
    exec_path: /usr/bin/chromium
    browser_args: ["disable-gpu", "window-size=1280,720"]
scenarios:
  - name: ok
    steps:
      - run: {command: echo hi}
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	r := s.Runners["web"]
	if r.Headless == nil || *r.Headless {
		t.Errorf("headless = %v, want explicit false", r.Headless)
	}
	if r.ExecPath != "/usr/bin/chromium" {
		t.Errorf("exec_path = %q, want /usr/bin/chromium", r.ExecPath)
	}
	if len(r.BrowserArgs) != 2 || r.BrowserArgs[0] != "disable-gpu" || r.BrowserArgs[1] != "window-size=1280,720" {
		t.Errorf("browser_args = %v, want [disable-gpu window-size=1280,720]", r.BrowserArgs)
	}
}

func TestLoadBytes_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		src      string
		wantKind Kind
		wantMsg  string
	}{
		{
			name:     "yaml syntax error",
			src:      "version: \"1\"\nsuite:\n  name: x\n  : bad",
			wantKind: KindParse,
		},
		{
			name:     "unknown field is strict-rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\nbogus: true\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo",
			wantKind: KindParse,
		},
		{
			name:     "wrong version",
			src:      "version: \"2\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo",
			wantKind: KindValidation,
			wantMsg:  "version must be",
		},
		{
			name:     "missing suite name",
			src:      "version: \"1\"\nsuite: {}\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo",
			wantKind: KindValidation,
			wantMsg:  "suite.name is required",
		},
		{
			name:     "no scenarios",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios: []",
			wantKind: KindValidation,
			wantMsg:  "at least one scenario",
		},
		{
			name:     "duplicate scenario names",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: dup\n    steps:\n      - run: {command: echo}\n  - name: dup\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "duplicate scenario name",
		},
		{
			name:     "step with two actions",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n        assert: {exit_code: 0}",
			wantKind: KindValidation,
			wantMsg:  "exactly one action",
		},
		{
			name:     "assert with two matchers",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          stdout:\n            contains: a\n            equals: b",
			wantKind: KindValidation,
			wantMsg:  "exactly one matcher",
		},
		{
			name:     "line below 1",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          stdout:\n            line: 0\n            equals: a",
			wantKind: KindValidation,
			wantMsg:  "line must be >= 1",
		},
		{
			name:     "run timeout must be a duration",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, timeout: \"30\"}",
			wantKind: KindValidation,
			wantMsg:  "run.timeout \"30\" is not a valid duration",
		},
		{
			name:     "runner timeout must be a duration",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://127.0.0.1:1\n    timeout: ten seconds\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "is not a valid duration",
		},
		{
			name:     "unknown runner reference fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://127.0.0.1:1\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, runner: sloww}",
			wantKind: KindValidation,
			wantMsg:  "runner \"sloww\" is not declared under runners: (declared: api)",
		},
		{
			name:     "runner type mismatch fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://127.0.0.1:1\nscenarios:\n  - name: a\n    steps:\n      - query: {runner: api, sql: select 1}",
			wantKind: KindValidation,
			wantMsg:  "runner \"api\" is a http runner; a query step needs a db runner",
		},
		{
			name:     "invalid stream regexp fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout:\n            matches: \"hi[\"",
			wantKind: KindValidation,
			wantMsg:  "is not a valid regexp",
		},
		{
			name:     "invalid not_matches regexp fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout:\n            not_matches: \"hi[\"",
			wantKind: KindValidation,
			wantMsg:  "not_matches \"hi[\" is not a valid regexp",
		},
		{
			name:     "matches and not_matches together are rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout:\n            matches: a\n            not_matches: b",
			wantKind: KindValidation,
			wantMsg:  "exactly one matcher",
		},
		{
			name:     "invalid json matches regexp fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout:\n            json: {path: \"$.x\", matches: \"a(\"}",
			wantKind: KindValidation,
			wantMsg:  "is not a valid regexp",
		},
		{
			name:     "invalid ready log regexp fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - name: svc\n        command: ./svc\n        ready:\n          log: \"up[\"\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "ready.log \"up[\" is not a valid regexp",
		},
		{
			name:     "invalid fixture mode fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - fixture: {file: f.txt, content: x, mode: \"rw-r--r--\"}\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "is not an octal file mode",
		},
		{
			name:     "service step inside a scenario is rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - service: {name: p, command: ./p}\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "service steps are only allowed in suite.setup",
		},
		{
			name:     "service step inside scenario teardown is rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n    teardown:\n      - service: {name: p, command: ./p}",
			wantKind: KindValidation,
			wantMsg:  "service steps are only allowed in suite.setup",
		},
		{
			name:     "service step inside suite teardown is rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\n  teardown:\n    - service: {name: p, command: ./p}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "service steps are only allowed in suite.setup",
		},
		{
			name:     "http step at suite level is rejected with a pointer",
			src:      "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - http: {method: GET, path: /}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "per-scenario",
		},
		{
			name:     "duplicate suite service names are rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {name: p, command: ./p}\n    - service: {name: p, command: ./q}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "duplicate suite service name",
		},
		{
			name:     "suite setup run without a command is rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - run: {shell: true}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "run.command is required",
		},
		{
			name:     "invalid fixture mtime fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - fixture: {file: f.txt, content: x, mtime: \"yesterday\"}\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "is not an RFC3339 timestamp",
		},
		{
			name:     "line with json is rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          stdout:\n            line: 1\n            json:\n              path: $.a\n              equals: 1",
			wantKind: KindValidation,
			wantMsg:  "line cannot be combined",
		},
		{
			name:     "service without a command",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - name: s\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "command is required",
		},
		{
			// Issue #44: an http runner carrying an ssh-only field is rejected.
			name:     "http runner with cross-type ssh field",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api: {type: http, base_url: 'http://x', host: example.com}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "cannot be set on a http runner",
		},
		{
			// Issue #44: a grpc runner carrying a db-only field is rejected.
			name:     "grpc runner with cross-type db field",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  svc: {type: grpc, target: 'localhost:50051', dsn: 'sqlite:./x.db'}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "cannot be set on a grpc runner",
		},
		{
			// A browser-only field on a non-browser runner is rejected like any other
			// cross-type field.
			name:     "http runner with cross-type browser field",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api: {type: http, base_url: 'http://x', headless: false}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "cannot be set on a http runner",
		},
		{
			name:     "browser runner with cross-type grpc field",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  web: {type: browser, target: 'localhost:50051'}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "cannot be set on a browser runner",
		},
		{
			name:     "service without a name",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - command: sleep 1\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "name is required",
		},
		{
			name:     "duplicate service names",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - {name: s, command: sleep 1}\n      - {name: s, command: sleep 1}\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "duplicate service name",
		},
		{
			name:     "service ready with two probes",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - name: s\n        command: sleep 1\n        ready: {file: r, port: 127.0.0.1:1}\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "set only one of file/port/log/delay",
		},
		{
			name:     "service ready store without file",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - name: s\n        command: sleep 1\n        ready: {log: up, store: addr}\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "ready.store requires file",
		},
		{
			name:     "service ready bad timeout",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - name: s\n        command: sleep 1\n        ready: {file: r, timeout: nope}\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "not a valid duration",
		},
		// Issue #27: runner-config validation error branches.
		{
			name:     "runner with an invalid type",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  r: {type: bogus}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "is invalid (want cmd",
		},
		{
			name:     "db runner without a dsn",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  d: {type: db}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "(db) requires a dsn",
		},
		{
			name:     "db runner with an unsupported driver",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  d: {type: db, driver: sqllite, dsn: \"sqlite:./app.db\"}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "unsupported runner.driver \"sqllite\"",
		},
		{
			name:     "ssh runner without a host",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  s: {type: ssh, user: bob}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "(ssh) requires a host",
		},
		{
			name:     "ssh runner without a user",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  s: {type: ssh, host: example.com}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "(ssh) requires a user",
		},
		{
			name:     "grpc runner without a target",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  g: {type: grpc}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "(grpc) requires a target",
		},
		{
			name:     "runner without a type",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  r: {cwd: .}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "type is required",
		},
		{
			name:     "cdp screenshot without a path",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  web: {type: browser}\nscenarios:\n  - name: a\n    steps:\n      - cdp:\n          runner: web\n          actions:\n            - screenshot: {selector: \"#x\"}",
			wantKind: KindValidation,
			wantMsg:  "screenshot requires a path",
		},
		{
			name:     "cdp attribute without a name",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  web: {type: browser}\nscenarios:\n  - name: a\n    steps:\n      - cdp:\n          runner: web\n          actions:\n            - attribute: {selector: \"#x\"}",
			wantKind: KindValidation,
			wantMsg:  "attribute requires selector and name",
		},
		{
			name:     "cdp press without a key",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  web: {type: browser}\nscenarios:\n  - name: a\n    steps:\n      - cdp:\n          runner: web\n          actions:\n            - press: {selector: \"#x\"}",
			wantKind: KindValidation,
			wantMsg:  "press requires selector and key",
		},
		{
			name:     "cdp action setting two keys",
			src:      "version: \"1\"\nsuite:\n  name: x\nrunners:\n  web: {type: browser}\nscenarios:\n  - name: a\n    steps:\n      - cdp:\n          runner: web\n          actions:\n            - {click: \"#a\", check: \"#b\"}",
			wantKind: KindValidation,
			wantMsg:  "set exactly one",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if err == nil {
				t.Fatalf("LoadBytes() error = nil, want error")
			}
			var lerr *Error
			if !errors.As(err, &lerr) {
				t.Fatalf("error type = %T, want *Error", err)
			}
			if lerr.Kind != tt.wantKind {
				t.Errorf("kind = %v, want %v (msg: %s)", lerr.Kind, tt.wantKind, lerr.Msg)
			}
			if tt.wantMsg != "" && !strings.Contains(lerr.Msg, tt.wantMsg) {
				t.Errorf("msg = %q, want substring %q", lerr.Msg, tt.wantMsg)
			}
		})
	}
}
