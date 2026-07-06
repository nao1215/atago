package loader

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
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

// TestLoadBytes_StripsLeadingBOM is a regression: a spec saved with a leading
// UTF-8 byte-order mark (routinely emitted by Windows/Notepad-family editors)
// must load. The raw bytes went straight to the YAML decoder, which glued the
// BOM onto the first key and failed with a confusing `unknown field "version"`
// that blamed a field the author wrote correctly. A single leading BOM is now
// stripped transparently, as most YAML tooling does.
func TestLoadBytes_StripsLeadingBOM(t *testing.T) {
	t.Parallel()
	src := "\ufeff" + `version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`
	s, err := LoadBytes("bom.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() with a leading BOM error = %v", err)
	}
	if s.Suite.Name != "sample" {
		t.Errorf("suite name = %q, want sample", s.Suite.Name)
	}
	if len(s.Scenarios) != 1 || s.Scenarios[0].Name != "ok" {
		t.Errorf("scenario not decoded through the BOM: %+v", s.Scenarios)
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

// TestLoadBytes_SandboxHome proves sandbox_home is accepted on run, pty, and
// defaults.run, decodes to the pointer, and is strict-rejected on a service (#71).
func TestLoadBytes_SandboxHome(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
defaults:
  run:
    sandbox_home: true
scenarios:
  - name: ok
    steps:
      - run:
          command: echo hi
      - pty:
          command: echo hi
          sandbox_home: true
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	run := s.Scenarios[0].Steps[0].Run
	if !run.SandboxHomeEnabled() {
		t.Errorf("defaults.run.sandbox_home did not layer onto the run step: %+v", run.SandboxHome)
	}
	pty := s.Scenarios[0].Steps[1].PTY
	if !pty.SandboxHomeEnabled() {
		t.Errorf("pty.sandbox_home not decoded: %+v", pty.SandboxHome)
	}

	// A service has no sandbox_home key: strict decode must reject it.
	bad := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    services:
      - name: peer
        command: sleep 1
        sandbox_home: true
    steps:
      - run: {command: echo hi}
`
	if _, err := LoadBytes("sample.atago.yaml", []byte(bad)); err == nil {
		t.Error("sandbox_home on a service should be strict-rejected, got nil error")
	}
}

// TestLoadBytes_Changes covers the load-time validation of the changes: assert
// target (#70): it must follow a run/pty step, entries must be workdir-relative
// and confined, and a valid placement loads.
func TestLoadBytes_Changes(t *testing.T) {
	t.Parallel()
	valid := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - run:
          command: echo hi
      - assert:
          changes:
            created:
              - out.txt
              - "site/*.html"
              - "dist/**"
              - "assets/**/*.css"
            modified: []
            deleted: []
`
	s, err := LoadBytes("sample.atago.yaml", []byte(valid))
	if err != nil {
		t.Fatalf("valid changes spec should load: %v", err)
	}
	ch := s.Scenarios[0].Steps[1].Assert.Changes
	if ch == nil || ch.Created == nil || len(*ch.Created) != 4 {
		t.Fatalf("changes.created not decoded: %+v", ch)
	}
	if ch.Modified == nil || len(*ch.Modified) != 0 {
		t.Errorf("modified: [] should decode to a non-nil empty list (assert nothing), got %+v", ch.Modified)
	}

	bad := []struct {
		name    string
		src     string
		wantMsg string
	}{
		{
			name:    "not preceded by run/pty",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          changes:\n            created: [out.txt]",
			wantMsg: "requires an immediately preceding run/pty step",
		},
		{
			name:    "preceded by http, not run/pty",
			src:     "version: \"1\"\nsuite:\n  name: x\nrunners:\n  api:\n    type: http\n    base_url: http://127.0.0.1:1\nscenarios:\n  - name: a\n    steps:\n      - http: {runner: api, method: GET, path: /}\n      - assert:\n          changes:\n            created: [out.txt]",
			wantMsg: "requires an immediately preceding run/pty step",
		},
		{
			name:    "absolute entry",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: [/etc/passwd]",
			wantMsg: "must be workdir-relative",
		},
		{
			name:    "escaping entry",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: [\"../escape.txt\"]",
			wantMsg: "escapes the scenario workdir",
		},
		{
			name:    "empty changes block",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes: {}",
			wantMsg: "set at least one of created/modified/deleted",
		},
		{
			name:    "malformed glob entry",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: [\"site/[unclosed\"]",
			wantMsg: "is not a valid glob",
		},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("sample.atago.yaml", []byte(tt.src))
			if err == nil || !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error = %v, want containing %q", err, tt.wantMsg)
			}
		})
	}
}

// TestLoadBytes_SSHRunFields covers the rejection of run-step fields that only
// shape local execution when the step names an ssh runner: the command runs
// remotely, so env/clear_env/pass_env/sandbox_home/stdin/stdout_to/stderr_to/cwd
// are silently dropped and must fail at load time. The same fields load fine on
// a cmd runner, and an ssh step limited to command/runner/timeout/retry loads.
func TestLoadBytes_SSHRunFields(t *testing.T) {
	t.Parallel()
	// sshSpec wraps a run: mapping naming ssh runner "box" (host+user set).
	sshSpec := func(run string) string {
		return "version: \"1\"\nsuite:\n  name: x\nrunners:\n  box: {type: ssh, host: h, user: u}\nscenarios:\n  - name: a\n    steps:\n      - run: {runner: box, " + run + "}"
	}
	cmdSpec := func(run string) string {
		return "version: \"1\"\nsuite:\n  name: x\nrunners:\n  local: {type: cmd}\nscenarios:\n  - name: a\n    steps:\n      - run: {runner: local, " + run + "}"
	}

	rejected := []struct {
		name  string
		run   string
		field string
	}{
		{"sandbox_home", "command: uptime, sandbox_home: true", "sandbox_home"},
		{"clear_env", "command: uptime, clear_env: true", "clear_env"},
		{"pass_env", "command: uptime, clear_env: true, pass_env: [PATH]", "pass_env"},
		{"env", "command: uptime, env: {A: b}", "env"},
		{"stdin", "command: cat, stdin: hello", "stdin"},
		{"stdout_to", "command: uptime, stdout_to: out.txt", "stdout_to"},
		{"stderr_to", "command: uptime, stderr_to: err.txt", "stderr_to"},
		{"cwd", "command: uptime, cwd: sub", "cwd"},
	}
	for _, tt := range rejected {
		t.Run("ssh rejects "+tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(sshSpec(tt.run)))
			want := "run." + tt.field + " has no effect on an ssh runner"
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("error = %v, want substring %q", err, want)
			}
		})
		t.Run("cmd accepts "+tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := LoadBytes("t.atago.yaml", []byte(cmdSpec(tt.run))); err != nil {
				t.Errorf("cmd runner should load %s: %v", tt.name, err)
			}
		})
	}

	t.Run("ssh with only command/runner/timeout/retry loads", func(t *testing.T) {
		t.Parallel()
		src := sshSpec("command: uptime, timeout: 30s, retry: {times: 3, until: {exit_code: 0}}")
		if _, err := LoadBytes("t.atago.yaml", []byte(src)); err != nil {
			t.Errorf("minimal ssh run step should load: %v", err)
		}
	})

	// shell is rejected with its own message: the remote login shell always
	// interprets the command, so the knob has nothing to switch.
	t.Run("ssh rejects shell", func(t *testing.T) {
		t.Parallel()
		_, err := LoadBytes("t.atago.yaml", []byte(sshSpec("command: uptime, shell: true")))
		want := "run.shell has no effect on an ssh runner (the remote login shell always interprets the command)"
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want substring %q", err, want)
		}
	})

	// A remote pipeline needs no shell: opt-in — the metacharacter hint (which
	// would suggest the now-rejected shell: true) must not fire for ssh steps.
	t.Run("ssh command with metacharacters loads without the shell hint", func(t *testing.T) {
		t.Parallel()
		src := sshSpec("command: \"ps aux | grep sshd > /tmp/out\"")
		if _, err := LoadBytes("t.atago.yaml", []byte(src)); err != nil {
			t.Errorf("ssh command with metacharacters should load (the remote shell interprets them): %v", err)
		}
	})
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
			// An empty file previously surfaced the raw decoder "EOF", which tells
			// the user nothing. It now names the problem and what a spec needs.
			name:     "empty file",
			src:      "",
			wantKind: KindParse,
			wantMsg:  "spec is empty",
		},
		{
			name:     "whitespace-only file",
			src:      "   \n\t\n  ",
			wantKind: KindParse,
			wantMsg:  "spec is empty",
		},
		{
			name:     "comments-only file",
			src:      "# just a comment\n# nothing else\n",
			wantKind: KindParse,
			wantMsg:  "spec is empty",
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
			name:     "whole-stream matcher cannot combine with a text matcher",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - assert:\n          stdout:\n            contains: a\n            equals: b",
			wantKind: KindValidation,
			wantMsg:  "cannot be combined with another matcher",
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
			name:     "invalid scrub pattern fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscrub:\n  - {pattern: \"a(\", placeholder: X}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "scrub[0]",
		},
		{
			name:     "empty scrub pattern fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscrub:\n  - {pattern: \"\", placeholder: X}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "scrub[0].pattern is required",
		},
		{
			name:     "invalid not_matches regexp fails at load",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          stdout:\n            not_matches: \"hi[\"",
			wantKind: KindValidation,
			wantMsg:  "not_matches \"hi[\" is not a valid regexp",
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
			name:     "pty session entry needs exactly one of expect and send",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty:\n          command: cat\n          session:\n            - expect: hi\n              send: \"yo\\n\"",
			wantKind: KindValidation,
			wantMsg:  "set exactly one of expect/send",
		},
		{
			name:     "pty expect must be a valid regexp",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty:\n          command: cat\n          session:\n            - expect: \"hi[\"",
			wantKind: KindValidation,
			wantMsg:  "is not a valid regexp",
		},
		{
			name:     "pty timeout must be positive",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty: {command: cat, timeout: \"0s\"}",
			wantKind: KindValidation,
			wantMsg:  "pty.timeout must be positive",
		},
		{
			name:     "pty rows above the terminal limit are rejected",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty: {command: cat, rows: 70000}",
			wantKind: KindValidation,
			wantMsg:  "rows/cols must be between 0 and 65535",
		},
		{
			name:     "pty timeout must be a duration",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty: {command: cat, timeout: \"soon\"}",
			wantKind: KindValidation,
			wantMsg:  "pty.timeout \"soon\" is not a valid duration",
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
		// PDF assertion validation (#73): every branch of validatePDF.
		{
			name:     "pdf assert without a path",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          pdf: {pages: 1}",
			wantKind: KindValidation,
			wantMsg:  "pdf.path is required",
		},
		{
			name:     "pdf assert with a negative page count",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          pdf: {path: r.pdf, pages: -1}",
			wantKind: KindValidation,
			wantMsg:  "page counts must be >= 0",
		},
		{
			name:     "pdf assert with min_pages above max_pages",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          pdf: {path: r.pdf, min_pages: 5, max_pages: 2}",
			wantKind: KindValidation,
			wantMsg:  "min_pages 5 exceeds max_pages 2",
		},
		{
			name:     "pdf assert with an unknown metadata field",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          pdf:\n            path: r.pdf\n            metadata: {bogus: v}",
			wantKind: KindValidation,
			wantMsg:  "unknown field \"bogus\"",
		},
		{
			name:     "pdf assert with no constraint",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          pdf: {path: r.pdf}",
			wantKind: KindValidation,
			wantMsg:  "must set at least one of pages/min_pages/max_pages/metadata/text",
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

// specSteps assembles a minimal one-scenario spec whose steps are the given
// flow-style step entries (each is the text after "- " in a steps list). It
// keeps the many one-off validation cases readable without hand-indenting YAML.
func specSteps(steps ...string) string {
	var b strings.Builder
	b.WriteString("version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n")
	for _, s := range steps {
		b.WriteString("      - " + s + "\n")
	}
	return b.String()
}

// mustReject asserts LoadBytes fails with a validation/parse error whose message
// contains want.
func mustReject(t *testing.T, name, src, want string) {
	t.Helper()
	_, err := LoadBytes("t.atago.yaml", []byte(src))
	if err == nil {
		t.Fatalf("%s: LoadBytes() error = nil, want error containing %q", name, want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("%s: error = %q, want substring %q", name, err.Error(), want)
	}
}

// mustAccept asserts LoadBytes loads src cleanly.
func mustAccept(t *testing.T, name, src string) {
	t.Helper()
	if _, err := LoadBytes("t.atago.yaml", []byte(src)); err != nil {
		t.Errorf("%s: LoadBytes() error = %v, want clean load", name, err)
	}
}

// TestBugHunt_Rejections drives the validation-error paths — the untrusted-input
// surface — asserting each malformed spec is rejected with an accurate message.
func TestBugHunt_Rejections(t *testing.T) {
	t.Parallel()

	dbRunner := "runners:\n  d: {type: db, dsn: \"sqlite::memory:\"}\n"
	grpcRunner := "runners:\n  g: {type: grpc, target: \"127.0.0.1:50051\"}\n"
	browserRunner := "runners:\n  b: {type: browser}\n"

	// withRunner prepends a runners: block to a specSteps body (which has no
	// runners of its own).
	withRunner := func(runner, body string) string {
		// body starts with `version: "1"\nsuite:\n  name: x\n...`; splice the
		// runner block in after the suite name line.
		const anchor = "  name: x\n"
		i := strings.Index(body, anchor) + len(anchor)
		return body[:i] + runner + body[i:]
	}

	// scenarioSpec builds a spec whose single scenario has extra scenario-level
	// blocks (services/mock_servers) plus steps.
	mockScenario := func(step string) string {
		return "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n" +
			"    mock_servers:\n      - name: api\n        routes:\n          - {method: GET, path: /, status: 200}\n" +
			"    steps:\n      - " + step + "\n"
	}
	svcScenario := func(step string) string {
		return "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n" +
			"    services:\n      - {name: s, command: sleep 10}\n" +
			"    steps:\n      - " + step + "\n"
	}
	scenarioTop := func(extra, step string) string {
		return "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    " + extra + "\n    steps:\n      - " + step + "\n"
	}

	tests := []struct{ name, src, want string }{
		// ---- store / validateStoreJSONPath / validateStore ----
		{"store name required", specSteps("store: {from: {header: X-Foo}}"), "store.name is required"},
		{"store shadows builtin", specSteps("store: {name: workdir, from: {header: X}}"), "shadows a built-in variable"},
		{"store from required", specSteps("store: {name: v}"), "store.from is required"},
		{"store from empty", specSteps("store: {name: v, from: {}}"), "must set one of stdout/body/file/header/rows/message/value"},
		{"store from two sources", specSteps("store: {name: v, from: {header: X, stdout: {json: {path: \"$.a\"}}}}"), "must set exactly one source"},
		{"store selector no json/matches", specSteps("store: {name: v, from: {stdout: {}}}"), "must set a json path, a matches regexp, or trim"},
		{"store selector bad matches", specSteps("store: {name: v, from: {stdout: {matches: \"a[\"}}}"), "is not a valid regexp"},
		{"store selector json path required", specSteps("store: {name: v, from: {stdout: {json: {}}}}"), "path is required"},
		{"store selector bad json path", specSteps("store: {name: v, from: {stdout: {json: {path: \"$[\"}}}}"), "is not a valid JSON path"},
		{"store selector json and trim", specSteps("store: {name: v, from: {stdout: {json: {path: \"$.a\"}, trim: true}}}"), "must set exactly one selector (json, matches, or trim)"},
		{"store file no json/text", specSteps("store: {name: v, from: {file: {path: out.txt}}}"), "must set a json path or text: true"},
		{"store file json and text", specSteps("store: {name: v, from: {file: {path: out.txt, json: {path: \"$.a\"}, text: true}}}"), "must set exactly one selector (json or text)"},
		{"assert stream rejects trim", specSteps("assert: {stdout: {trim: true}}"), "trim is only valid in a store source"},
		{"assert file rejects text", specSteps("assert: {file: {path: out.txt, text: true}}"), "text is only valid in a store source"},
		{"store file bad json path", specSteps("store: {name: v, from: {file: {path: out.txt, json: {path: \"$.\"}}}}"), "is not a valid JSON path"},

		// ---- validateAssert / validateAssertTarget ----
		{"assert no target", specSteps("assert: {}"), "must set at least one assertion target"},

		// ---- validateExitCode ----
		{"exit_code none set", specSteps("assert: {exit_code: {}}"), "must be an int, {not: int}, or {in: [int, ...]}"},
		{"exit_code two forms", specSteps("assert: {exit_code: {not: 1, in: [2]}}"), "set exactly one of a bare int, not, or in"},
		{"exit_code in empty", specSteps("assert: {exit_code: {in: []}}"), "in must list at least one accepted exit code"},
		{"exit_code in dup", specSteps("assert: {exit_code: {in: [0, 0]}}"), "in lists 0 more than once"},

		// ---- validateStream ----
		{"stream no matcher", specSteps("assert: {stdout: {}}"), "must set at least one matcher"},
		{"stream text plus whole-stream matcher", specSteps("assert: {stdout: {contains: a, equals: b}}"), "cannot be combined with another matcher"},
		{"stream line with snapshot", specSteps("assert: {stdout: {line: 1, snapshot: snap}}"), "line cannot be combined with json/yaml/snapshot"},
		{"stream contains empty list", specSteps("assert: {stdout: {contains: []}}"), "contains must not be empty"},
		{"stream contains empty element", specSteps("assert: {stdout: {contains: [\"\"]}}"), "is an empty string"},

		// ---- validateFile ----
		{"file path required", specSteps("assert: {file: {exists: true}}"), "file.path is required"},
		{"file no matcher", specSteps("assert: {file: {path: out.txt}}"), "must set one of exists/contains/not_contains/executable/equals/equals_file/json/snapshot"},
		{"file two matchers", specSteps("assert: {file: {path: out.txt, exists: true, snapshot: s}}"), "must set exactly one of exists/contains/not_contains/executable/equals/equals_file/json/snapshot"},
		{"file not_contains empty", specSteps("assert: {file: {path: out.txt, not_contains: []}}"), "not_contains must not be empty"},
		{"file equals and equals_file exclusive", specSteps("assert: {file: {path: out.txt, equals: x, equals_file: in.txt}}"), "must set exactly one of exists/contains/not_contains/executable/equals/equals_file/json/snapshot"},
		{"file equals_file empty", specSteps("assert: {file: {path: out.txt, equals_file: \"\"}}"), "equals_file must not be empty"},

		// ---- validateHeaderMatch ----
		{"header name required", specSteps("assert: {header: {equals: text/html}}"), "header.name is required"},
		{"header no matcher", specSteps("assert: {header: {name: Content-Type}}"), "must set one of contains/equals/matches"},
		{"header two matchers", specSteps("assert: {header: {name: X, contains: a, equals: b}}"), "must set exactly one of contains/equals/matches"},
		{"header bad matches", specSteps("assert: {header: {name: X, matches: \"a[\"}}"), "is not a valid regexp"},

		// ---- validateJSON ----
		{"json path required", specSteps("assert: {stdout: {json: {equals: 1}}}"), "json.path is required"},
		{"json no matcher", specSteps("assert: {stdout: {json: {path: \"$.a\"}}}"), "must set one of equals/matches/length/gt/gte/lt/lte"},
		{"json two matchers", specSteps("assert: {stdout: {json: {path: \"$.a\", equals: 1, length: 2}}}"), "must set exactly one of equals/matches/length/gt/gte/lt/lte"},
		{"json bad matches", specSteps("assert: {stdout: {json: {path: \"$.a\", matches: \"a[\"}}}"), "is not a valid regexp"},

		// ---- validateMockAssert ----
		{"mock name required", specSteps("assert: {mock: {count: 1}}"), "mock.name is required"},
		{"mock undeclared", specSteps("assert: {mock: {name: nope, count: 1}}"), "is not a declared mock server"},
		{"mock count negative", mockScenario("assert: {mock: {name: api, count: -1}}"), "count must be >= 0"},
		{"mock count zero with matcher", mockScenario("assert: {mock: {name: api, count: 0, header: {name: X, equals: y}}}"), "count: 0 cannot be combined"},
		{"mock header invalid", mockScenario("assert: {mock: {name: api, header: {name: X}}}"), "must set one of contains/equals/matches"},
		{"mock body invalid", mockScenario("assert: {mock: {name: api, body: {}}}"), "must set at least one matcher"},

		// ---- validateCondition ----
		{"skip bad os", scenarioTop("skip: {os: solaris}", "run: {command: echo}"), "skip.os \"solaris\" is invalid"},
		{"only bad os", scenarioTop("only: {os: bsd}", "run: {command: echo}"), "only.os \"bsd\" is invalid"},

		// ---- validateStep ----
		{"step no action", specSteps("{}"), "step must set exactly one of fixture/run/http/query/grpc/cdp/assert/store/pty/signal (got none)"},
		{"step two actions", specSteps("{run: {command: x}, store: {name: v, from: {header: X}}}"), "step must set exactly one action, but set"},
		{"query missing runner", withRunner(dbRunner, specSteps("query: {sql: \"SELECT 1\"}")), "query.runner is required"},
		{"query missing sql", withRunner(dbRunner, specSteps("query: {runner: d}")), "query.sql is required"},
		{"grpc missing runner", withRunner(grpcRunner, specSteps("grpc: {method: m}")), "grpc.runner is required"},
		{"grpc missing method", withRunner(grpcRunner, specSteps("grpc: {runner: g}")), "grpc.method is required"},
		{"cdp missing runner", withRunner(browserRunner, specSteps("cdp: {actions: [{navigate: \"https://x\"}]}")), "cdp.runner is required"},
		{"cdp empty actions", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: []}")), "cdp.actions must contain at least one action"},
		{"service step in scenario", specSteps("service: {name: s, command: x}"), "service steps are only allowed in suite.setup"},

		// ---- validateCDPActions ----
		{"cdp no action key", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{}]}")), "sets no recognized action"},
		{"cdp multiple actions", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{navigate: x, click: \"#a\"}]}")), "sets multiple actions"},
		{"cdp press incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{press: {selector: \"#a\"}}]}")), "press requires selector and key"},
		{"cdp select incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{select: {value: v}}]}")), "select requires a selector"},
		{"cdp screenshot incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{screenshot: {}}]}")), "screenshot requires a path"},
		{"cdp attribute incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{attribute: {selector: \"#a\"}}]}")), "attribute requires selector and name"},
		{"cdp send_keys incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{send_keys: {value: hi}}]}")), "send_keys requires a selector"},
		{"cdp upload incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{upload: {selector: \"#a\"}}]}")), "upload requires selector and file"},
		{"cdp download incomplete", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{download: {}}]}")), "download requires a click selector"},

		// ---- validateFixture ----
		{"fixture file required", specSteps("fixture: {content: hi}"), "fixture.file is required"},
		{"fixture two sources", specSteps("fixture: {file: a.txt, content: x, base64: eA==}"), "set only one of content, base64, from, or symlink"},
		{"fixture symlink with mode", specSteps("fixture: {file: a, symlink: b, mode: \"0644\"}"), "mode cannot be applied to a symlink"},
		{"fixture bad mode", specSteps("fixture: {file: a, mode: \"999\"}"), "is not an octal file mode"},
		{"fixture bad mtime", specSteps("fixture: {file: a, mtime: nope}"), "is not an RFC3339 timestamp"},

		// ---- validateSignal ----
		{"signal service required", svcScenario("signal: {signal: TERM}"), "signal.service is required"},
		{"signal undeclared", svcScenario("signal: {service: nope, signal: TERM}"), "is not a declared service"},
		{"signal signal required", svcScenario("signal: {service: s}"), "signal.signal is required"},
		{"signal bad name", svcScenario("signal: {service: s, signal: BOGUS}"), "is not an accepted signal"},
		{"signal wait bad duration", svcScenario("signal: {service: s, signal: TERM, wait: {timeout: abc}}"), "is not a valid duration"},
		{"signal wait nonpositive", svcScenario("signal: {service: s, signal: TERM, wait: {timeout: \"-1s\"}}"), "must be positive"},

		// ---- validatePTY ----
		{"pty command required", specSteps("pty: {session: [{expect: hi}]}"), "pty.command is required"},
		{"pty bad timeout", specSteps("pty: {command: sh, timeout: abc}"), "is not a valid duration"},
		{"pty nonpositive timeout", specSteps("pty: {command: sh, timeout: \"-1s\"}"), "must be positive"},
		{"pty size overflow", specSteps("pty: {command: sh, rows: 70000}"), "rows/cols must be between 0 and 65535"},
		{"pty expect and send", specSteps("pty: {command: sh, session: [{expect: hi, send: x}]}"), "set exactly one of expect/send (got both)"},
		{"pty neither expect nor send", specSteps("pty: {command: sh, session: [{}]}"), "set exactly one of expect/send"},
		{"pty bad expect regexp", specSteps("pty: {command: sh, session: [{expect: \"a[\"}]}"), "is not a valid regexp"},
		{"pty bad send key", specSteps("pty: {command: sh, session: [{send: {key: BOGUS}}]}"), "is not a supported key"},

		// ---- validateMockRoutes (scenario) ----
		{"route method required", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    mock_servers:\n      - name: m\n        routes:\n          - {path: /}\n    steps:\n      - run: {command: echo}\n", "method is required"},
		{"route path required", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    mock_servers:\n      - name: m\n        routes:\n          - {method: GET}\n    steps:\n      - run: {command: echo}\n", "path is required"},
		{"route path no slash", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    mock_servers:\n      - name: m\n        routes:\n          - {method: GET, path: foo}\n    steps:\n      - run: {command: echo}\n", "must start with \"/\""},
		{"route two payloads", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    mock_servers:\n      - name: m\n        routes:\n          - {method: GET, path: /, body: hi, body_file: f.txt}\n    steps:\n      - run: {command: echo}\n", "set at most one of json/body/body_file"},
		{"route bad status", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    mock_servers:\n      - name: m\n        routes:\n          - {method: GET, path: /, status: 700}\n    steps:\n      - run: {command: echo}\n", "is not a valid HTTP status"},
		{"route bad delay", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    mock_servers:\n      - name: m\n        routes:\n          - {method: GET, path: /, delay: xyz}\n    steps:\n      - run: {command: echo}\n", "is not a valid duration"},

		// ---- validateSuiteBlock ----
		{"suite http rejected", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - http: {runner: api, method: GET, path: /}\nrunners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "steps are per-scenario"},
		{"suite service in teardown", "version: \"1\"\nsuite:\n  name: x\n  teardown:\n    - service: {name: s, command: x}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "service steps are only allowed in suite.setup"},
		{"suite duplicate service", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {name: s, command: x}\n    - service: {name: s, command: y}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "duplicate suite service name"},
		{"suite service missing command", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {name: s}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "service.command is required"},
		{"suite service missing name", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {command: x}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "service.name is required"},
		{"suite mock in teardown", "version: \"1\"\nsuite:\n  name: x\n  teardown:\n    - mock_server: {name: m, routes: []}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "mock_server steps are only allowed in suite.setup"},
		{"suite duplicate mock", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - mock_server: {name: m, routes: []}\n    - mock_server: {name: m, routes: []}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "duplicate suite mock server name"},
		{"suite mock missing name", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - mock_server: {routes: []}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "mock_server.name is required"},
		{"suite step no action", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - {}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "step must set exactly one action"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mustReject(t, tt.name, tt.src, tt.want)
		})
	}
}

// TestBugHunt_Acceptances pins the accept side of each accept/reject boundary so
// a future over-eager validation rule that starts rejecting a legal spec is
// caught here.
func TestBugHunt_Acceptances(t *testing.T) {
	t.Parallel()

	browserRunner := "runners:\n  b: {type: browser}\n"
	withRunner := func(runner, body string) string {
		const anchor = "  name: x\n"
		i := strings.Index(body, anchor) + len(anchor)
		return body[:i] + runner + body[i:]
	}
	mockScenario := func(step string) string {
		return "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n" +
			"    mock_servers:\n      - name: api\n        routes:\n          - {method: GET, path: /, status: 200}\n" +
			"    steps:\n      - " + step + "\n"
	}
	svcScenario := func(step string) string {
		return "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n" +
			"    services:\n      - {name: s, command: sleep 10}\n" +
			"    steps:\n      - " + step + "\n"
	}
	scenarioTop := func(extra, step string) string {
		return "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    " + extra + "\n    steps:\n      - " + step + "\n"
	}

	tests := []struct{ name, src string }{
		{"store stdout json", specSteps("store: {name: v, from: {stdout: {json: {path: \"$.a\"}}}}")},
		{"store header", specSteps("store: {name: v, from: {header: X-Request-Id}}")},
		{"store stdout matches", specSteps("store: {name: v, from: {stdout: {matches: \"id=(\\\\d+)\"}}}")},
		{"store file json", specSteps("store: {name: v, from: {file: {path: out.json, json: {path: \"$.id\"}}}}")},
		{"exit_code in", specSteps("run: {command: echo}", "assert: {exit_code: {in: [0, 1, 2]}}")},
		{"exit_code not", specSteps("run: {command: echo}", "assert: {exit_code: {not: 1}}")},
		{"file exists", specSteps("assert: {file: {path: out.txt, exists: true}}")},
		{"header equals", specSteps("assert: {header: {name: Content-Type, equals: text/html}}")},
		{"json gt", specSteps("assert: {stdout: {json: {path: \"$.count\", gt: 5}}}")},
		{"mock count", mockScenario("assert: {mock: {name: api, count: 2}}")},
		{"signal valid", svcScenario("signal: {service: s, signal: TERM}")},
		{"signal var target", svcScenario("signal: {service: \"${svc}\", signal: KILL}")},
		{"cdp navigate", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{navigate: \"https://x\"}, {click: \"#go\"}]}"))},
		{"fixture content", specSteps("fixture: {file: a.txt, content: hello}")},
		{"pty valid", specSteps("pty: {command: sh, session: [{expect: \"[$] \"}, {send: \"ls\\n\"}]}")},
		{"assert message", specSteps("assert: {message: {equals: ok}}")},
		{"assert value", specSteps("assert: {value: {contains: hi}}")},
		{"assert grpc_status", specSteps("assert: {grpc_status: 0}")},
		{"assert screen after pty", specSteps("pty: {command: sh}", "assert: {screen: {contains: prompt}}")},
		{"assert duration after run", specSteps("run: {command: echo}", "assert: {duration: {lt: \"5s\"}}")},
		{"skip valid os", scenarioTop("skip: {os: darwin}", "run: {command: echo}")},
		{"only valid os", scenarioTop("only: {os: windows}", "run: {command: echo}")},
		{"suite setup kinds", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - fixture: {file: seed.txt, content: hi}\n    - run: {command: echo}\n    - store: {name: v, from: {stdout: {json: {path: \"$.a\"}}}}\n    - assert: {exit_code: 0}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mustAccept(t, tt.name, tt.src)
		})
	}
}

// TestBugHunt_RoundTrip is a metamorphic check: a spec that loads cleanly must,
// after being re-marshaled to YAML and reloaded, still load cleanly. It catches
// any decode/marshal asymmetry that would silently change validation outcome.
func TestBugHunt_RoundTrip(t *testing.T) {
	t.Parallel()
	srcs := []string{
		specSteps("run: {command: echo}", "assert: {exit_code: {in: [0, 1]}}"),
		specSteps("store: {name: v, from: {stdout: {json: {path: \"$.a\"}}}}"),
		specSteps("fixture: {file: a.txt, content: hello}"),
		specSteps("assert: {stdout: {contains: [\"a\", \"b\"]}}"),
	}
	for i, src := range srcs {
		s1, err := LoadBytes("t.atago.yaml", []byte(src))
		if err != nil {
			t.Fatalf("case %d: initial load failed: %v", i, err)
		}
		out, err := yaml.Marshal(s1)
		if err != nil {
			t.Fatalf("case %d: marshal failed: %v", i, err)
		}
		if _, err := LoadBytes("t.atago.yaml", out); err != nil {
			t.Errorf("case %d: reload after round-trip failed: %v\nmarshaled:\n%s", i, err, out)
		}
	}
}

// TestBugHunt_LoadFromDisk covers Load / LoadWithSource error and success paths
// against the filesystem (the entry points the CLI actually calls).
func TestBugHunt_LoadFromDisk(t *testing.T) {
	t.Parallel()

	// Missing file: Load and LoadWithSource both surface a path-annotated error.
	if _, err := Load(filepath.Join(t.TempDir(), "nope.atago.yaml")); err == nil {
		t.Error("Load(missing) error = nil, want error")
	}
	if _, _, err := LoadWithSource(filepath.Join(t.TempDir(), "nope.atago.yaml")); err == nil {
		t.Error("LoadWithSource(missing) error = nil, want error")
	}

	// Invalid spec on disk: LoadWithSource returns the validation error, not a source.
	bad := filepath.Join(t.TempDir(), "bad.atago.yaml")
	if err := os.WriteFile(bad, []byte("version: \"2\"\nsuite: {name: x}\nscenarios: []"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, src, err := LoadWithSource(bad); err == nil || src != nil {
		t.Errorf("LoadWithSource(bad) = src %v err %v, want nil src + error", src, err)
	}

	// Valid spec on disk loads.
	good := filepath.Join(t.TempDir(), "good.atago.yaml")
	if err := os.WriteFile(good, []byte(specSteps("run: {command: echo}")), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(good); err != nil {
		t.Errorf("Load(good) error = %v", err)
	}
}

// TestBugHunt_ErrorString exercises Error.Error's path-less branch.
func TestBugHunt_ErrorString(t *testing.T) {
	t.Parallel()
	if got := (&Error{Msg: "boom"}).Error(); got != "boom" {
		t.Errorf("path-less Error() = %q, want %q", got, "boom")
	}
	if got := (&Error{Path: "f.yaml", Msg: "boom"}).Error(); got != "f.yaml: boom" {
		t.Errorf("Error() = %q, want %q", got, "f.yaml: boom")
	}
}

// TestBugHunt_DirAssert drives validateDir's accept/reject boundary (#25/#74):
// the tree-snapshot vs matcher-family split, count sanity, and glob validity.
func TestBugHunt_DirAssert(t *testing.T) {
	t.Parallel()

	dir := func(body string) string { return specSteps("assert: {dir: " + body + "}") }

	reject := []struct{ name, src, want string }{
		{"dir path required", dir("{exists: true}"), "dir.path is required"},
		{"dir no matcher", dir("{path: out}"), "must set at least one of exists/contains/not_contains/count/min_count/max_count/glob/snapshot"},
		{"dir snapshot with matcher", dir("{path: out, snapshot: tree, exists: true}"), "snapshot cannot be combined with the matcher family"},
		{"dir snapshot with recursive", dir("{path: out, snapshot: tree, recursive: true}"), "recursive is implied by snapshot; drop it"},
		{"dir recursive without matcher", dir("{path: out, recursive: true}"), "recursive needs at least one of"},
		{"dir ignore without recursive", dir("{path: out, count: 1, ignore: [\"*.tmp\"]}"), "ignore only applies to recursive or snapshot"},
		{"dir negative count", dir("{path: out, count: -1}"), "counts must be >= 0"},
		{"dir min exceeds max", dir("{path: out, min_count: 5, max_count: 2}"), "min_count 5 exceeds max_count 2"},
		{"dir bad ignore glob", dir("{path: out, recursive: true, count: 1, ignore: [\"[\"]}"), "is not a valid glob"},
	}
	for _, tt := range reject {
		t.Run("reject/"+tt.name, func(t *testing.T) {
			t.Parallel()
			mustReject(t, tt.name, tt.src, tt.want)
		})
	}

	accept := []struct{ name, src string }{
		{"dir exists", dir("{path: out, exists: true}")},
		{"dir count", dir("{path: out, count: 3}")},
		{"dir min max", dir("{path: out, min_count: 1, max_count: 4}")},
		{"dir glob", dir("{path: out, glob: \"*.txt\"}")},
		{"dir snapshot", dir("{path: out, snapshot: tree}")},
		{"dir recursive with matcher and ignore", dir("{path: out, recursive: true, count: 2, ignore: [\"*.tmp\", \"logs/**\"]}")},
		// KNOWN GAP (not a crash, intentionally pinned): `recursive: true` alongside
		// only `exists` is accepted, because `exists` counts toward n so the
		// recursive-needs-a-matcher guard (n==0) never fires — even though that
		// guard's own message excludes exists from the composable matchers. The
		// recursive flag is effectively a silent no-op here. Pinned so a future
		// tightening is a deliberate, visible change.
		{"dir recursive with only exists (gap)", dir("{path: out, exists: true, recursive: true}")},
	}
	for _, tt := range accept {
		t.Run("accept/"+tt.name, func(t *testing.T) {
			t.Parallel()
			mustAccept(t, tt.name, tt.src)
		})
	}
}

// TestBugHunt_FileAndJSONExtras closes the remaining file/json matcher branches
// (executable, snapshot, the numeric json bounds) on the accept side.
func TestBugHunt_FileAndJSONExtras(t *testing.T) {
	t.Parallel()
	accept := []struct{ name, src string }{
		{"file executable", specSteps("assert: {file: {path: bin/tool, executable: true}}")},
		{"file snapshot", specSteps("assert: {file: {path: out.txt, snapshot: golden}}")},
		{"file contains list", specSteps("assert: {file: {path: out.txt, contains: [\"a\", \"b\"]}}")},
		{"file json", specSteps("assert: {file: {path: out.json, json: {path: \"$.id\", equals: 7}}}")},
		{"json lt", specSteps("assert: {stdout: {json: {path: \"$.n\", lt: 10}}}")},
		{"json lte", specSteps("assert: {stdout: {json: {path: \"$.n\", lte: 10}}}")},
		{"json gte", specSteps("assert: {stdout: {json: {path: \"$.n\", gte: 1}}}")},
		{"json length", specSteps("assert: {stdout: {json: {path: \"$.items\", length: 3}}}")},
		{"yaml matcher", specSteps("assert: {stdout: {yaml: {path: \"$.k\", equals: v}}}")},
		{"store from body matches", specSteps("store: {name: v, from: {body: {matches: \"tok=(\\\\w+)\"}}}")},
		{"store from rows json", specSteps("store: {name: v, from: {rows: {json: {path: \"$[0].id\"}}}}")},
		{"store from message json", specSteps("store: {name: v, from: {message: {json: {path: \"$.ok\"}}}}")},
		{"store from value matches", specSteps("store: {name: v, from: {value: {matches: \"^ok$\"}}}")},
	}
	for _, tt := range accept {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mustAccept(t, tt.name, tt.src)
		})
	}
}
