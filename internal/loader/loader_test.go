package loader

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/nao1215/atago/internal/store"
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

// TestLoadBytes_PTYKeyFamilies proves the key families added in #376 survive
// the real YAML path, not just the byte table: shift-tab and the modified
// arrows contain a hyphen, and `alt-b` sits next to a `ctrl-b` that must stay a
// different key. It also pins that the vocabulary is closed — a Meta digit is
// rejected rather than silently sending nothing.
func TestLoadBytes_PTYKeyFamilies(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - pty:
          command: editor
          session:
            - send: {key: shift-tab}
            - send: {key: backtab}
            - send: {key: insert}
            - send: {key: alt-b}
            - send: {key: alt-enter}
            - send: {key: alt-backspace}
            - send: {key: ctrl-left}
            - send: {key: shift-up}
`
	s, err := LoadBytes("sample.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes() error = %v", err)
	}
	want := []string{
		"\x1b[Z", "\x1b[Z", "\x1b[2~", "\x1bb",
		"\x1b\r", "\x1b\x7f", "\x1b[1;5D", "\x1b[1;2A",
	}
	session := s.Scenarios[0].Steps[0].PTY.Session
	if len(session) != len(want) {
		t.Fatalf("session has %d actions, want %d", len(session), len(want))
	}
	for i, w := range want {
		if got := string(session[i].Send.Bytes()); got != w {
			t.Errorf("session[%d] (%s) sends %q, want %q", i, session[i].Send.Key, got, w)
		}
	}

	bad := `
version: "1"
suite:
  name: sample
scenarios:
  - name: ok
    steps:
      - pty:
          command: editor
          session:
            - send: {key: alt-1}
`
	if _, err := LoadBytes("sample.atago.yaml", []byte(bad)); err == nil {
		t.Error("alt-1 is not in the vocabulary and should be a load error, got nil")
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
            ignore: [".atago-home/**"]
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
	if len(ch.Ignore) != 1 || ch.Ignore[0] != ".atago-home/**" {
		t.Errorf("changes.ignore not decoded: %+v", ch.Ignore)
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
		{
			name:    "ignore alone asserts nothing",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            ignore: [\"cache/**\"]",
			wantMsg: "set at least one of created/modified/deleted",
		},
		{
			name:    "empty ignore list",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: []\n            ignore: []",
			wantMsg: "ignore must not be empty",
		},
		{
			name:    "absolute ignore glob",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: []\n            ignore: [\"/var/**\"]",
			wantMsg: "must be workdir-relative",
		},
		{
			name:    "ignore glob escaping the workdir",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: []\n            ignore: [\"../elsewhere/**\"]",
			wantMsg: "escapes the scenario workdir",
		},
		{
			name:    "malformed ignore glob",
			src:     "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert:\n          changes:\n            created: []\n            ignore: [\"cache/[unclosed\"]",
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
			name:     "scenario name with a newline",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: \"line\\nbreak\"\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  `must not contain the control character "\n"`,
		},
		{
			name:     "scenario name with a tab",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: \"tab\\there\"\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  `must not contain the control character "\t"`,
		},
		{
			name:     "suite name with a newline",
			src:      "version: \"1\"\nsuite:\n  name: \"bad\\nname\"\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  `suite.name must not contain the control character "\n"`,
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
			name:     "pty session entry needs exactly one of expect, send, and expect_screen",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty:\n          command: cat\n          session:\n            - expect: hi\n              send: \"yo\\n\"",
			wantKind: KindValidation,
			wantMsg:  "set exactly one of expect/send/expect_screen",
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
			name:     "retry interval must not be negative",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run:\n          command: echo\n          retry: {times: 3, interval: -1s, until: {exit_code: 0}}",
			wantKind: KindValidation,
			wantMsg:  "retry.interval must not be negative",
		},
		{
			name:     "service max_log_bytes must be positive",
			src:      "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - {name: s, command: sleep 1, max_log_bytes: -1}\n    steps:\n      - run: {command: echo}",
			wantKind: KindValidation,
			wantMsg:  "max_log_bytes must be positive",
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

// TestLoadBytes_MalformedYAMLDoesNotPanic pins the loader's no-panic contract on
// untrusted input: some malformed YAML makes the underlying goccy/go-yaml decoder
// nil-panic (found by FuzzLoadBytes). LoadBytes must recover and return a clean
// parse error instead of crashing the process.
func TestLoadBytes_MalformedYAMLDoesNotPanic(t *testing.T) {
	t.Parallel()
	// Reduced from the fuzz crasher testdata/fuzz/FuzzLoadBytes/230de42ba4751bda.
	inputs := []string{
		"A: 0\nrunners:\n 0: {0000000000000\"}\nscenarios:\n  ! 00",
		"scenarios:\n  ! 0",
	}
	for _, src := range inputs {
		s, err := LoadBytes("t.atago.yaml", []byte(src))
		if err == nil {
			t.Errorf("LoadBytes(%q) = nil error, want a parse error for malformed YAML", src)
		}
		if s != nil {
			t.Errorf("LoadBytes(%q) returned a non-nil spec with a parse error", src)
		}
		var lerr *Error
		if errors.As(err, &lerr) && lerr.Kind != KindParse {
			t.Errorf("LoadBytes(%q) kind = %v, want KindParse", src, lerr.Kind)
		}
	}
}

// TestLoadBytes_ExplicitTagRejected pins the tag rejection that keeps the
// decoder's nil-panic path unreachable. Every input here reaches
// ast.TagNode.ArrayRange -> nil *ArrayNodeIter -> decodeSlice panic in
// goccy/go-yaml v1.19.2 when it is allowed through, so each must now be
// rejected before decoding, with a message that names the tag and its position
// instead of the opaque "malformed YAML" the recover path produces.
func TestLoadBytes_ExplicitTagRejected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			name: "bare tag over a scalar where a list belongs",
			src:  "scenarios:\n  ! 0",
			want: []string{"explicit YAML tag", "[2:3]"},
		},
		{
			name: "fuzz crasher 230de42ba4751bda",
			src:  "A: 0\nrunners:\n 0: {0000000000000\"}\nscenarios:\n  ! 00",
			want: []string{"explicit YAML tag", "[5:3]"},
		},
		{
			name: "custom tag on a list field",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios: !foo 0\n",
			want: []string{"!foo", "[4:12]"},
		},
		{
			name: "std tag on a list field",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios: !!str 0\n",
			want: []string{"!!str", "[4:12]"},
		},
		{
			name: "tag on a nested list field",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps: ! 0\n",
			want: []string{"explicit YAML tag", "[6:12]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := LoadBytes("t.atago.yaml", []byte(tt.src))
			if err == nil {
				t.Fatalf("LoadBytes() = nil error, want a parse error naming the tag")
			}
			if s != nil {
				t.Errorf("LoadBytes() returned a non-nil spec alongside an error")
			}
			var lerr *Error
			if !errors.As(err, &lerr) {
				t.Fatalf("LoadBytes() error = %T, want *loader.Error", err)
			}
			if lerr.Kind != KindParse {
				t.Errorf("kind = %v, want KindParse", lerr.Kind)
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want substring %q", err.Error(), want)
				}
			}
			// The recover path must no longer be what produces this error.
			if strings.Contains(err.Error(), "malformed YAML") {
				t.Errorf("error = %q, want the tag to be rejected before decoding", err.Error())
			}
		})
	}
}

// TestLoadBytes_TagRejectionLeavesOtherErrorsAlone guards the blast radius of
// the tag check: a spec with no explicit tag must keep loading, and a malformed
// spec that the parser itself rejects must keep its original syntax error
// rather than being reported as a tag problem.
func TestLoadBytes_TagRejectionLeavesOtherErrorsAlone(t *testing.T) {
	t.Parallel()
	good := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n"
	if _, err := LoadBytes("t.atago.yaml", []byte(good)); err != nil {
		t.Errorf("LoadBytes(untagged spec) error = %v, want a clean load", err)
	}
	// An unclosed flow mapping is a syntax error the parser itself catches, so
	// the tag check never gets a usable AST to walk.
	broken := "version: \"1\"\nsuite: {name: x\n"
	_, err := LoadBytes("t.atago.yaml", []byte(broken))
	if err == nil {
		t.Fatalf("LoadBytes(unclosed flow mapping) = nil error, want a parse error")
	}
	if strings.Contains(err.Error(), "explicit YAML tag") {
		t.Errorf("error = %q, want the original syntax error, not a tag error", err.Error())
	}
}

// TestLoadBytes_BinaryTagAccepted pins the one tag a spec may carry. `atago
// record` emits !!binary whenever a captured stream is not valid UTF-8, so
// rejecting it would make recorded specs unloadable — a round trip the record
// package's own tests depend on.
func TestLoadBytes_BinaryTagAccepted(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n" +
		"      - run: {command: echo}\n      - assert:\n          stdout:\n            equals: !!binary \"//4K\"\n"
	if _, err := LoadBytes("t.atago.yaml", []byte(src)); err != nil {
		t.Errorf("LoadBytes(!!binary scalar) error = %v, want a clean load", err)
	}
	// A tag nested under a !!binary node must still be caught: skipping the
	// binary tag must not skip the rest of the subtree.
	nested := "version: \"1\"\nsuite:\n  name: x\nscenarios: !!binary !foo 0\n"
	if _, err := LoadBytes("t.atago.yaml", []byte(nested)); err == nil {
		t.Errorf("LoadBytes(tag nested under !!binary) = nil error, want the inner tag rejected")
	}
}

// TestLoadBytes_JSONEqualsNull pins the difference between `equals: null` — an
// assertion that the selected value IS JSON null — and an omitted `equals` key,
// which means no matcher was set (#309). Both decode the value to a nil
// interface, so only the recorded key presence separates them.
func TestLoadBytes_JSONEqualsNull(t *testing.T) {
	t.Parallel()

	for _, spelling := range []string{"equals: null", "equals: ~", "equals: Null"} {
		src := specSteps("run: {command: echo}", "assert: {stdout: {json: {path: \"$.v\", "+spelling+"}}}")
		s, err := LoadBytes("t.atago.yaml", []byte(src))
		if err != nil {
			t.Fatalf("LoadBytes(%s) error = %v, want a clean load", spelling, err)
		}
		checks := s.Scenarios[0].Steps[1].Assert.Stdout.JSON
		if len(checks) != 1 {
			t.Fatalf("%s: got %d checks, want 1", spelling, len(checks))
		}
		if checks[0].Equals != nil {
			t.Errorf("%s: Equals = %#v, want a nil value (the YAML null)", spelling, checks[0].Equals)
		}
		if !checks[0].HasEquals() {
			t.Errorf("%s: HasEquals() = false, want the matcher recorded as present", spelling)
		}
	}

	// An `equals` key written with a real value keeps reporting present, and a
	// check with no `equals` key at all is still matcher-less.
	src := specSteps("run: {command: echo}", "assert: {stdout: {json: [{path: \"$.a\", equals: 1}, {path: \"$.b\", equals: null}]}}")
	s, err := LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("LoadBytes(list form) error = %v", err)
	}
	for i, c := range s.Scenarios[0].Steps[1].Assert.Stdout.JSON {
		if !c.HasEquals() {
			t.Errorf("list check %d: HasEquals() = false, want true", i)
		}
	}
	mustReject(t, "json without equals", specSteps("assert: {stdout: {json: {path: \"$.v\"}}}"),
		"must set one of equals/matches/length/gt/gte/lt/lte")
	mustReject(t, "json equals null with a second matcher", specSteps("assert: {stdout: {json: {path: \"$.v\", equals: null, length: 1}}}"),
		"must set exactly one of equals/matches/length/gt/gte/lt/lte")
	// The presence check must not swallow the strict unknown-key rejection: a
	// custom UnmarshalYAML bypasses the loader's document-wide yaml.Strict(),
	// so a typo inside a check has to be caught by the check itself.
	mustReject(t, "json typo'd matcher key", specSteps("assert: {stdout: {json: {path: \"$.v\", equal: null}}}"),
		"equal")
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
	requireDiagnosticCodes(t, name, err)
}

// atgCode matches the code prefix every load error is supposed to carry.
var atgCode = regexp.MustCompile(`ATG\d{4}: `)

// requireDiagnosticCodes asserts every message inside a load error carries a
// diagnostic code. The published error reference can only explain what it can
// name, and the coverage gate proves the reference is provoked — nothing proved
// the converse, which is how the matrix validator emitted five uncoded messages
// beside coded ones in a single error list. Every rejection case in this file is
// a corpus entry for that check.
func requireDiagnosticCodes(t *testing.T, name string, err error) {
	t.Helper()
	lines := strings.Split(err.Error(), "\n")
	// One problem is rendered as a single message that may carry a multi-line
	// source excerpt below it, so only its first line is checked. Several are
	// rendered as a count header followed by one "  - " entry each, and every
	// entry has to name its own diagnostic.
	if !strings.HasSuffix(lines[0], "validation errors:") {
		if !atgCode.MatchString(lines[0]) {
			t.Errorf("%s: error %q carries no ATG code; every load error must name the diagnostic that explains it", name, lines[0])
		}
		return
	}
	for _, line := range lines[1:] {
		entry, listed := strings.CutPrefix(strings.TrimSpace(line), "- ")
		if !listed {
			continue // continuation of the entry above
		}
		if !atgCode.MatchString(entry) {
			t.Errorf("%s: error entry %q carries no ATG code; every load error must name the diagnostic that explains it", name, entry)
		}
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

		// ---- json/yaml list (#156) ----
		{"stream json list element two matchers", specSteps("assert: {stdout: {json: [{path: \"$.a\", equals: 1}, {path: \"$.b\", equals: 2, length: 3}]}}"), "must set exactly one of equals/matches/length/gt/gte/lt/lte"},
		{"stream json list missing matcher", specSteps("assert: {stdout: {json: [{path: \"$.a\", equals: 1}, {path: \"$.b\"}]}}"), "must set one of equals/matches/length/gt/gte/lt/lte"},
		{"stream json list element path required", specSteps("assert: {stdout: {json: [{path: \"$.a\", equals: 1}, {equals: 2}]}}"), "path is required"},
		{"stream json empty list", specSteps("assert: {stdout: {json: []}}"), "at least one check"},
		{"store json list rejected", specSteps("store: {name: v, from: {stdout: {json: [{path: \"$.a\", equals: 1}, {path: \"$.b\", equals: 2}]}}}"), "single value with one json path"},

		// ---- validateFile ----
		{"file path required", specSteps("assert: {file: {exists: true}}"), "file.path is required"},
		{"file no matcher", specSteps("assert: {file: {path: out.txt}}"), "must set one of exists/contains/not_contains/executable/equals/equals_file/json/snapshot"},
		{"file two matchers", specSteps("assert: {file: {path: out.txt, exists: true, snapshot: s}}"), "must set exactly one of exists/contains/not_contains/executable/equals/equals_file/json/snapshot"},
		{"file not_contains empty", specSteps("assert: {file: {path: out.txt, not_contains: []}}"), "not_contains must not be empty"},
		{"file equals and equals_file exclusive", specSteps("assert: {file: {path: out.txt, equals: x, equals_file: in.txt}}"), "must set exactly one of exists/contains/not_contains/executable/equals/equals_file/json/snapshot"},
		{"file equals_file empty", specSteps("assert: {file: {path: out.txt, equals_file: \"\"}}"), "equals_file must not be empty"},

		// ---- occurrence counts (#396) ----
		{"stream count without a countable matcher", specSteps("assert: {stdout: {equals: x, count: 1}}"), "need a contains or matches matcher to count"},
		{"stream count with both countable matchers", specSteps("assert: {stdout: {contains: a, matches: b, count: 1}}"), "exactly one countable matcher"},
		{"stream count with a contains list", specSteps("assert: {stdout: {contains: [a, b], count: 1}}"), "count applies to one substring"},
		{"stream count and range together", specSteps("assert: {stdout: {contains: a, count: 1, min_count: 2}}"), "count is the exact form"},
		{"stream count negative", specSteps("assert: {stdout: {contains: a, count: -1}}"), "count must be >= 0"},
		{"stream count unsatisfiable range", specSteps("assert: {stdout: {contains: a, min_count: 3, max_count: 2}}"), "greater than max_count"},
		{"stream count with snapshot", specSteps("assert: {stdout: {snapshot: s.txt, count: 1}}"), "cannot be combined with snapshot"},
		{"file count without contains", specSteps("assert: {file: {path: out.txt, exists: true, count: 1}}"), "need a contains matcher to count"},
		{"file count with a contains list", specSteps("assert: {file: {path: out.txt, contains: [a, b], count: 1}}"), "count applies to one substring"},

		// ---- file size bounds (#397) ----
		{"file size negative", specSteps("assert: {file: {path: out.txt, size: -1}}"), "size must be >= 0"},
		{"file size and range together", specSteps("assert: {file: {path: out.txt, size: 1, min_size: 2}}"), "size is the exact form"},
		{"file size unsatisfiable range", specSteps("assert: {file: {path: out.txt, min_size: 9, max_size: 2}}"), "greater than max_size"},
		{"file size with snapshot", specSteps("assert: {file: {path: out.txt, snapshot: s.txt, size: 3}}"), "cannot be combined with snapshot"},
		{"file size with exists false", specSteps("assert: {file: {path: out.txt, exists: false, size: 0}}"), "an absent file has no size"},
		{"file size with two content matchers", specSteps("assert: {file: {path: out.txt, exists: true, contains: a, size: 3}}"), "must set exactly one of exists/contains"},
		// ---- deterministic (#398) ----
		{"deterministic runs below two", specSteps("run: {command: echo hi, deterministic: {runs: 1}}"), "must be at least 2"},
		{"deterministic runs above the cap", specSteps("run: {command: echo hi, deterministic: {runs: 99}}"), "capped at 10"},
		{"deterministic unknown observable", specSteps("run: {command: echo hi, deterministic: {compare: [stdout, workdir]}}"), "unknown observable"},
		{"deterministic duplicate observable", specSteps("run: {command: echo hi, deterministic: {compare: [stdout, stdout]}}"), "more than once"},
		{"deterministic with retry", specSteps("run: {command: echo hi, deterministic: {}, retry: {times: 2, until: {exit_code: 0}}}"), "cannot be combined with retry"},

		// ---- expect_fail (#395) ----
		{"expect_fail without a reason", "version: \"1\"\nsuite: {name: s}\nscenarios:\n  - name: known\n    expect_fail: {issue: \"http://x\"}\n    steps:\n      - run: {command: echo hi}\n", "expect_fail.reason is required"},
		{"expect_fail with a blank reason", "version: \"1\"\nsuite: {name: s}\nscenarios:\n  - name: known\n    expect_fail: {reason: \"   \"}\n    steps:\n      - run: {command: echo hi}\n", "expect_fail.reason is required"},

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

		// ---- empty-matching regexps outside not_matches (#557) ----
		{"count bound on an empty-matching pattern", specSteps("run: {command: echo}", "assert: {stdout: {matches: \"q*\", max_count: 0}}"), "matches the empty string"},
		{"count bound on an optional group", specSteps("run: {command: echo}", "assert: {stdout: {matches: \"(foo)?\", count: 1}}"), "matches the empty string"},
		{"scrub rule matching the empty string", "version: \"1\"\nsuite: {name: s}\nscrub:\n  - {pattern: \"[0-9]*\", placeholder: \"<ID>\"}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n", "matches the empty string"},
		{"store capture matching the empty string", specSteps("run: {command: echo}", "store: {name: v, from: {stdout: {matches: \"[0-9]*\"}}}"), "matches the empty string"},

		// ---- matchers of one assert that contradict each other (#558) ----
		{"contains and not_contains share an entry", specSteps("run: {command: echo}", "assert: {stdout: {contains: [abc], not_contains: [abc]}}"), "contains and not_contains both list \"abc\""},
		{"matches equals not_matches", specSteps("run: {command: echo}", "assert: {stdout: {matches: \"a.c\", not_matches: \"a.c\"}}"), "matches and not_matches are the same pattern"},
		{"dir contains with count zero", specSteps("assert: {dir: {path: d, contains: [x], count: 0}}"), "count: 0 cannot hold together with contains"},
		{"dir contains above max_count", specSteps("assert: {dir: {path: d, contains: [x, y], max_count: 1}}"), "requires at least 2 entries"},
		{"dir contains and not_contains share an entry", specSteps("assert: {dir: {path: d, contains: [x], not_contains: [x]}}"), "contains and not_contains both list \"x\""},

		// ---- validateCondition ----
		{"skip bad os", scenarioTop("skip: {os: solaris}", "run: {command: echo}"), "skip.os \"solaris\" is invalid"},
		{"only bad os", scenarioTop("only: {os: bsd}", "run: {command: echo}"), "only.os \"bsd\" is invalid"},
		{"empty skip gate", scenarioTop("skip: {}", "run: {command: echo}"), "skip must name a condition"},
		{"empty only gate", scenarioTop("only: {}", "run: {command: echo}"), "only must name a condition"},
		{"canceling os gates", scenarioTop("skip: {os: linux}\n    only: {os: linux}", "run: {command: echo}"), "skip.os and only.os both name \"linux\""},
		{"canceling env gates", scenarioTop("skip: {env: FEATURE_X}\n    only: {env: FEATURE_X}", "run: {command: echo}"), "skip.env and only.env both name \"FEATURE_X\""},
		{"canceling command gates", scenarioTop("skip: {command: \"true\"}\n    only: {command: \"true\"}", "run: {command: echo}"), "skip.command and only.command both name \"true\""},

		// ---- validateDeterministic empty compare (#564) ----
		{"deterministic empty compare", specSteps("run: {command: echo, deterministic: {compare: []}}"), "deterministic.compare must not be empty"},

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
		{"pty expect and send", specSteps("pty: {command: sh, session: [{expect: hi, send: x}]}"), "set exactly one of expect/send/expect_screen/resize/exec (got more than one)"},
		{"pty neither expect nor send", specSteps("pty: {command: sh, session: [{}]}"), "set exactly one of expect/send/expect_screen/resize/exec"},
		{"pty bad expect regexp", specSteps("pty: {command: sh, session: [{expect: \"a[\"}]}"), "is not a valid regexp"},
		{"pty bad send key", specSteps("pty: {command: sh, session: [{send: {key: BOGUS}}]}"), "is not a supported key"},
		{"pty send times zero", specSteps("pty: {command: sh, session: [{send: {key: left, times: 0}}]}"), "times must be at least 1"},
		{"pty send times negative", specSteps("pty: {command: sh, session: [{send: {key: left, times: -3}}]}"), "times must be at least 1"},
		{"pty send times over cap", specSteps("pty: {command: sh, session: [{send: {key: left, times: 10001}}]}"), "times must not exceed 10000"},
		{"pty send times not an integer", specSteps("pty: {command: sh, session: [{send: {key: left, times: 1.5}}]}"), "times must be an integer"},
		{"pty send times without key", specSteps("pty: {command: sh, session: [{send: {times: 3}}]}"), "requires a key name"},
		{"pty send unknown mapping key", specSteps("pty: {command: sh, session: [{send: {key: left, nope: 1}}]}"), "accepted: key, times, paste"},
		{"pty send key and paste", specSteps("pty: {command: sh, session: [{send: {key: left, paste: hi}}]}"), "set exactly one of {key: <name>}, {paste: <text>}, or {mouse: {...}}"},
		{"pty send paste with times", specSteps("pty: {command: sh, session: [{send: {paste: hi, times: 2}}]}"), "a paste is delivered once"},
		{"pty send paste not a string", specSteps("pty: {command: sh, session: [{send: {paste: 5}}]}"), "paste must be a string"},
		{"pty resize missing cols", specSteps("pty: {command: sh, session: [{resize: {rows: 40}}]}"), "rows and cols are both required"},
		{"pty resize zero", specSteps("pty: {command: sh, session: [{resize: {rows: 0, cols: 80}}]}"), "rows and cols are both required"},
		{"pty resize overflow", specSteps("pty: {command: sh, session: [{resize: {rows: 70000, cols: 80}}]}"), "rows/cols must be between 1 and 65535"},
		{"pty resize with send", specSteps("pty: {command: sh, session: [{resize: {rows: 4, cols: 8}, send: x}]}"), "set exactly one of expect/send/expect_screen/resize/exec (got more than one)"},
		{"pty exec empty mapping", specSteps("pty: {command: sh, session: [{exec: {}}]}"), "exec.command is required"},
		{"pty exec bad timeout", specSteps("pty: {command: sh, session: [{exec: {command: touch x, timeout: abc}}]}"), "is not a valid duration"},
		{"pty exec nonpositive timeout", specSteps("pty: {command: sh, session: [{exec: {command: touch x, timeout: \"-1s\"}}]}"), "must be positive"},
		{"pty exec unknown key", specSteps("pty: {command: sh, session: [{exec: {command: touch x, cwd: sub}}]}"), "accepted: command, shell, timeout"},
		{"pty exec with expect", specSteps("pty: {command: sh, session: [{exec: touch x, expect: hi}]}"), "set exactly one of expect/send/expect_screen/resize/exec (got more than one)"},
		{"pty mouse missing col", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 5}}}]}"), "row and col are required 1-based screen cells"},
		{"pty mouse zero row", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 0, col: 3}}}]}"), "row and col are required 1-based screen cells"},
		{"pty mouse bad button", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 1, col: 1, button: scroll}}}]}"), "is not a supported button"},
		{"pty mouse bad action", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 1, col: 1, action: drag}}}]}"), "is not a supported action"},
		{"pty mouse bad mod", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 1, col: 1, mods: [super]}}}]}"), "is not a supported modifier"},
		{"pty mouse wheel release", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 1, col: 1, button: wheel-up, action: release}}}]}"), "a wheel button has no release event"},
		{"pty mouse with key", specSteps("pty: {command: sh, session: [{send: {key: enter, mouse: {row: 1, col: 1}}}]}"), "set exactly one of"},
		{"screen attrs text required", specSteps("assert: {screen: {attrs: [{fg: red}]}}"), "text is required"},
		{"screen attrs need a claim", specSteps("assert: {screen: {attrs: [{text: ERROR}]}}"), "needs at least one of fg/bg/bold"},
		{"screen attrs bad color", specSteps("assert: {screen: {attrs: [{text: E, fg: chartreuse}]}}"), "is not a color"},
		{"screen attrs color out of range", specSteps("assert: {screen: {attrs: [{text: E, fg: \"300\"}]}}"), "outside the 256-color palette"},
		{"expect_screen attrs validated too", specSteps("pty: {command: sh, session: [{expect_screen: {attrs: [{text: E}]}}]}"), "needs at least one of fg/bg/bold"},
		{"pty mouse unknown key", specSteps("pty: {command: sh, session: [{send: {mouse: {row: 1, col: 1, speed: 2}}}]}"), "accepted: row, col, button, action, mods"},
		{"pty expect_screen with snapshot rejected", specSteps("pty: {command: sh, session: [{expect_screen: {snapshot: snap.txt}}]}"), "snapshot is not supported in expect_screen"},
		{"pty expect_screen with trim rejected", specSteps("pty: {command: sh, session: [{expect_screen: {contains: hi, trim: true}}]}"), "trim is not supported in expect_screen"},
		{"pty expect_screen stable exceeds timeout", specSteps("pty: {command: sh, session: [{expect_screen: {contains: hi, timeout: \"20ms\", stable_for: \"30ms\"}}]}"), "must not exceed scenario \"a\".steps[0].pty.session[0].expect_screen.timeout"},
		// With no action-local timeout the session budget is what bounds the
		// wait, so a stable_for above it can never succeed either.
		{"pty expect_screen stable exceeds the session timeout", specSteps("pty: {command: sh, timeout: \"2s\", session: [{expect_screen: {contains: hi, stable_for: \"60s\"}}]}"), "must not exceed scenario \"a\".steps[0].pty.timeout"},
		// Same contradiction against the built-in session default when the step
		// sets no timeout of its own.
		{"pty expect_screen stable exceeds the default session timeout", specSteps("pty: {command: sh, session: [{expect_screen: {contains: hi, stable_for: \"60s\"}}]}"), "must not exceed the pty session timeout"},

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

	// Keep these bulk validation tables serial inside the parent test. The
	// loader work here is tiny, and aggressive nested parallelism has crashed
	// the Go runtime on Windows CI.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		{"store stdout json", specSteps("run: {command: echo}", "store: {name: v, from: {stdout: {json: {path: \"$.a\"}}}}")},
		{"store header", withRunner("runners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\n", specSteps("http: {runner: api, method: GET, path: /}", "store: {name: v, from: {header: X-Request-Id}}"))},
		{"store stdout matches", specSteps("run: {command: echo}", "store: {name: v, from: {stdout: {matches: \"id=(\\\\d+)\"}}}")},
		{"store file json", specSteps("store: {name: v, from: {file: {path: out.json, json: {path: \"$.id\"}}}}")},
		{"exit_code in", specSteps("run: {command: echo}", "assert: {exit_code: {in: [0, 1, 2]}}")},
		{"exit_code not", specSteps("run: {command: echo}", "assert: {exit_code: {not: 1}}")},
		{"file exists", specSteps("assert: {file: {path: out.txt, exists: true}}")},
		{"header equals", withRunner("runners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\n", specSteps("http: {runner: api, method: GET, path: /}", "assert: {header: {name: Content-Type, equals: text/html}}"))},
		{"json gt", specSteps("run: {command: echo}", "assert: {stdout: {json: {path: \"$.count\", gt: 5}}}")},
		{"mock count", mockScenario("assert: {mock: {name: api, count: 2}}")},
		{"signal valid", svcScenario("signal: {service: s, signal: TERM}")},
		{"signal var target", svcScenario("signal: {service: \"${svc}\", signal: KILL}")},
		{"cdp navigate", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{navigate: \"https://x\"}, {click: \"#go\"}]}"))},
		{"fixture content", specSteps("fixture: {file: a.txt, content: hello}")},
		{"pty valid", specSteps("pty: {command: sh, session: [{expect: \"[$] \"}, {send: \"ls\\n\"}]}")},
		{"pty expect_screen valid", specSteps("pty: {command: sh, session: [{expect_screen: {contains: hi, stable_for: \"20ms\"}}]}")},
		{"assert message", withRunner("runners:\n  rpc: {type: grpc, target: \"127.0.0.1:1\"}\n", specSteps("grpc: {runner: rpc, method: pkg.S/M}", "assert: {message: {equals: ok}}"))},
		{"assert value", withRunner(browserRunner, specSteps("cdp: {runner: b, actions: [{text: \"h1\"}]}", "assert: {value: {contains: hi}}"))},
		{"assert grpc_status", withRunner("runners:\n  rpc: {type: grpc, target: \"127.0.0.1:1\"}\n", specSteps("grpc: {runner: rpc, method: pkg.S/M}", "assert: {grpc_status: 0}"))},
		{"assert screen after pty", specSteps("pty: {command: sh}", "assert: {screen: {contains: prompt}}")},
		{"assert duration after run", specSteps("run: {command: echo}", "assert: {duration: {lt: \"5s\"}}")},
		{"skip valid os", scenarioTop("skip: {os: darwin}", "run: {command: echo}")},
		{"only valid os", scenarioTop("only: {os: windows}", "run: {command: echo}")},
		// Contradiction checks compare literally, so anything that can be
		// satisfied keeps loading.
		{"contains and not_contains that differ", specSteps("run: {command: echo}", "assert: {stdout: {contains: [abc], not_contains: [abcd]}}")},
		{"matches and not_matches that differ", specSteps("run: {command: echo}", "assert: {stdout: {matches: \"a.c\", not_matches: \"a\\\\.c\"}}")},
		{"dir contains repeated names one child", specSteps("assert: {dir: {path: d, contains: [x, x], count: 1}}")},
		{"dir contains a nested path under its ceiling", specSteps("assert: {dir: {path: d, contains: [\"assets/app.css\"], max_count: 1}}")},
		{"recursive dir contains with a count", specSteps("assert: {dir: {path: d, recursive: true, contains: [\"a/b.txt\"], count: 0}}")},
		{"empty-matching pattern without a count bound", specSteps("run: {command: echo}", "assert: {stdout: {matches: \"z*\"}}")},
		// Gates that name DIFFERENT fields compose normally: "this scenario is
		// POSIX-only and needs fzf" is an ordinary spec, and so are two gates on
		// one field with different values.
		{"gates on different fields", scenarioTop("skip: {os: windows}\n    only: {command: fzf}", "run: {command: echo}")},
		{"gates on one field with different values", scenarioTop("skip: {os: windows}\n    only: {os: linux}", "run: {command: echo}")},
		{"deterministic compare listed", specSteps("run: {command: echo, deterministic: {compare: [stdout]}}")},
		{"deterministic without compare", specSteps("run: {command: echo, deterministic: {runs: 3}}")},
		{"suite setup kinds", "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - fixture: {file: seed.txt, content: hi}\n    - run: {command: echo}\n    - store: {name: v, from: {stdout: {json: {path: \"$.a\"}}}}\n    - assert: {exit_code: 0}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		specSteps("run: {command: echo}", "store: {name: v, from: {stdout: {json: {path: \"$.a\"}}}}"),
		specSteps("fixture: {file: a.txt, content: hello}"),
		specSteps("run: {command: echo}", "assert: {stdout: {contains: [\"a\", \"b\"]}}"),
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
			mustAccept(t, tt.name, tt.src)
		})
	}
}

// TestBugHunt_FileAndJSONExtras closes the remaining file/json matcher branches
// (executable, snapshot, the numeric json bounds) on the accept side.
func TestBugHunt_FileAndJSONExtras(t *testing.T) {
	t.Parallel()
	// A store reads what a step produced, so the runner-backed sources need both
	// the runner and the step that fills the source.
	withRunners := func(body string) string {
		runners := "runners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\n  db: {type: db, dsn: \"sqlite:./a.db\"}\n  rpc: {type: grpc, target: \"127.0.0.1:1\"}\n  b: {type: browser}\n"
		return strings.Replace(body, "scenarios:", runners+"scenarios:", 1)
	}
	accept := []struct{ name, src string }{
		{"file executable", specSteps("assert: {file: {path: bin/tool, executable: true}}")},
		{"file snapshot", specSteps("assert: {file: {path: out.txt, snapshot: golden}}")},
		{"file contains list", specSteps("assert: {file: {path: out.txt, contains: [\"a\", \"b\"]}}")},
		{"file json", specSteps("assert: {file: {path: out.json, json: {path: \"$.id\", equals: 7}}}")},
		{"json lt", specSteps("run: {command: echo}", "assert: {stdout: {json: {path: \"$.n\", lt: 10}}}")},
		{"json lte", specSteps("run: {command: echo}", "assert: {stdout: {json: {path: \"$.n\", lte: 10}}}")},
		{"json gte", specSteps("run: {command: echo}", "assert: {stdout: {json: {path: \"$.n\", gte: 1}}}")},
		{"json length", specSteps("run: {command: echo}", "assert: {stdout: {json: {path: \"$.items\", length: 3}}}")},
		{"yaml matcher", specSteps("run: {command: echo}", "assert: {stdout: {yaml: {path: \"$.k\", equals: v}}}")},
		{"store from body matches", withRunners(specSteps("http: {runner: api, method: GET, path: /}", "store: {name: v, from: {body: {matches: \"tok=(\\\\w+)\"}}}"))},
		{"store from rows json", withRunners(specSteps("query: {runner: db, sql: \"SELECT 1\"}", "store: {name: v, from: {rows: {json: {path: \"$[0].id\"}}}}"))},
		{"store from message json", withRunners(specSteps("grpc: {runner: rpc, method: pkg.S/M}", "store: {name: v, from: {message: {json: {path: \"$.ok\"}}}}"))},
		{"store from value matches", withRunners(specSteps("cdp: {runner: b, actions: [{text: \"h1\"}]}", "store: {name: v, from: {value: {matches: \"^ok$\"}}}"))},
	}
	for _, tt := range accept {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mustAccept(t, tt.name, tt.src)
		})
	}
}

// TestLoadBytes_UnionErrorsCarryPosition guards the location quality of the
// hand-decoded union nodes (exit_code, stdin, pty send, json checks). Their
// custom unmarshalers used to return bare fmt.Errorf messages, so in a 300-line
// spec with a dozen exit_code asserts the error named neither scenario, step,
// nor line — a hunt, while a typo'd KEY got a [line:col], source excerpt, and
// caret. Every union-shape error must now carry the same [line:col] annotation
// pointing at the offending value, plus the source excerpt.
func TestLoadBytes_UnionErrorsCarryPosition(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		src     string
		wantMsg string // the human explanation, unchanged
		wantPos string // the exact [line:col] of the offending value
		excerpt string // a fragment of the offending source line
	}{
		{
			name: "exit_code wrong type",
			src: `version: "1"
suite:
  name: x
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert:
          exit_code: zero
`,
			wantMsg: "exit_code must be an integer",
			wantPos: "[9:22]",
			excerpt: "exit_code: zero",
		},
		{
			name: "stdin wrong shape",
			src: `version: "1"
suite:
  name: x
scenarios:
  - name: a
    steps:
      - run:
          command: cat
          stdin: [1, 2]
`,
			wantMsg: "stdin must be a string, {file: path}, or {base64: data}",
			wantPos: "[9:18]",
			excerpt: "stdin: [1, 2]",
		},
		{
			name: "pty send unknown key",
			src: `version: "1"
suite:
  name: x
scenarios:
  - name: a
    steps:
      - pty:
          command: cat
          session:
            - send: {keyy: enter}
`,
			wantMsg: `send: unknown key "keyy"`,
			wantPos: "[10:21]",
			excerpt: "keyy: enter",
		},
		{
			name: "json empty check list",
			src: `version: "1"
suite:
  name: x
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert:
          stdout:
            json: []
`,
			wantMsg: "a json/yaml matcher list must have at least one check",
			wantPos: "[10:19]",
			excerpt: "json: []",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("t.atago.yaml", []byte(tc.src))
			if err == nil {
				t.Fatal("LoadBytes accepted the malformed union value")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantMsg) {
				t.Errorf("error = %q, want the explanation %q", msg, tc.wantMsg)
			}
			if !strings.Contains(msg, tc.wantPos) {
				t.Errorf("error = %q, want the position %q pointing at the offending value", msg, tc.wantPos)
			}
			if !strings.Contains(msg, tc.excerpt) {
				t.Errorf("error = %q, want a source excerpt containing %q", msg, tc.excerpt)
			}
		})
	}
}

// TestLoadBytes_CwdEscapesWorkdir is a regression: run.cwd is documented as a
// working directory relative to the scenario workdir, and a `../` one walked
// straight out of it — `cwd: ../../../../../..` ran the command at the
// filesystem root. Nothing said so, and every assertion the scenario then made
// (changes:, dir:, file:) still looked at the untouched sandbox, so a scenario
// could act on the host and pass having done none of what it claimed. Other
// workdir-relative fields have rejected the same traversal all along.
func TestLoadBytes_CwdEscapesWorkdir(t *testing.T) {
	t.Parallel()
	bad := []struct {
		name string
		src  string
	}{
		{
			name: "run step",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: \"../elsewhere\"}",
		},
		{
			name: "bare parent",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: \"..\"}",
		},
		{
			name: "traversal that re-enters",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: \"sub/../../out\"}",
		},
		{
			name: "defaults.run",
			src:  "version: \"1\"\nsuite:\n  name: x\ndefaults:\n  run: {cwd: \"../elsewhere\"}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}",
		},
		{
			name: "service",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - {name: s, command: sleep 1, cwd: \"../elsewhere\"}\n    steps:\n      - run: {command: echo}",
		},
		{
			name: "pty step",
			src:  "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - pty: {command: cat, cwd: \"../elsewhere\"}",
		},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadBytes("s.atago.yaml", []byte(tt.src))
			if err == nil {
				t.Fatalf("cwd escaping the workdir was accepted:\n%s", tt.src)
			}
			if !strings.Contains(err.Error(), "escapes the scenario workdir") {
				t.Errorf("err = %v, want an escapes-the-scenario-workdir rejection", err)
			}
		})
	}

	// What must keep loading: a sub-directory, the workdir itself, and an
	// absolute path, which is explicit in a way `../..` is not.
	good := []struct {
		name string
		src  string
	}{
		{"sub-directory", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: sub}"},
		{"nested sub-directory", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: \"a/b/c\"}"},
		{"dot", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: \".\"}"},
		{"re-entering traversal that stays inside", "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo, cwd: \"a/../b\"}"},
	}
	for _, tt := range good {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := LoadBytes("s.atago.yaml", []byte(tt.src)); err != nil {
				t.Errorf("a workdir-relative cwd was rejected: %v\n%s", err, tt.src)
			}
		})
	}
}

// TestLoadBytes_EveryBuiltinIsReserved is a regression: the engine seeds five
// variables into every scenario store, and the guard that stops a store or a
// matrix key from shadowing one knew about three. A `store: {name: specdir}`
// was accepted and silently redefined ${specdir} for the rest of the scenario,
// so a later step reading a committed file through it read somewhere else —
// while `store: {name: workdir}` was rejected with advice to pick another name.
func TestLoadBytes_EveryBuiltinIsReserved(t *testing.T) {
	t.Parallel()
	for _, name := range store.Builtins {
		t.Run("store "+name, func(t *testing.T) {
			t.Parallel()
			src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - store: {name: " + name + ", from: {stdout: {trim: true}}}"
			_, err := LoadBytes("s.atago.yaml", []byte(src))
			if err == nil {
				t.Fatalf("store named %q was accepted; it shadows a built-in", name)
			}
			if !strings.Contains(err.Error(), "shadows a built-in variable") {
				t.Errorf("err = %v, want a shadowing rejection", err)
			}
		})
		t.Run("matrix "+name, func(t *testing.T) {
			t.Parallel()
			src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: \"m ${" + name + "}\"\n    matrix:\n      - { " + name + ": v }\n    steps:\n      - run: {command: echo}"
			_, err := LoadBytes("s.atago.yaml", []byte(src))
			if err == nil {
				t.Fatalf("matrix key %q was accepted; it shadows a built-in", name)
			}
			if !strings.Contains(err.Error(), "shadows a built-in variable") {
				t.Errorf("err = %v, want a shadowing rejection", err)
			}
		})
	}

	// A name that merely resembles one is still the author's to use.
	src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - store: {name: workdir_path, from: {stdout: {trim: true}}}"
	if _, err := LoadBytes("s.atago.yaml", []byte(src)); err != nil {
		t.Errorf("a non-builtin name was rejected: %v", err)
	}
}

// TestLoadBytes_DuplicateTag is a regression: a tag list is a set, so repeating
// an entry selects nothing extra — but `atago doc` counts tag occurrences, so a
// scenario listing `smoke` twice made the summary read "smoke (2)" over a suite
// where one scenario carries it. `atago list` showed "smoke,smoke" beside it.
func TestLoadBytes_DuplicateTag(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    tags: [smoke, smoke, slow]\n    steps:\n      - run: {command: echo}"
	_, err := LoadBytes("s.atago.yaml", []byte(src))
	if err == nil {
		t.Fatal("a repeated tag was accepted")
	}
	if !strings.Contains(err.Error(), `duplicate tag "smoke"`) {
		t.Errorf("err = %v, want a duplicate-tag rejection", err)
	}

	// Distinct tags, and the same tag on two different scenarios, are the
	// ordinary case and must keep loading.
	ok := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    tags: [smoke, slow]\n    steps:\n      - run: {command: echo}\n  - name: b\n    tags: [smoke]\n    steps:\n      - run: {command: echo}"
	if _, err := LoadBytes("s.atago.yaml", []byte(ok)); err != nil {
		t.Errorf("distinct tags were rejected: %v", err)
	}
}

// TestLoadBytes_AssertNeedsItsProducingStep is a regression: the same
// authoring mistake — an assertion whose producing step never ran — was
// classified three different ways. `screen:` before a pty step was refused at
// load (ATG2107), a `store from.header` after a run step errored at runtime
// (exit 4), and `status:` or a leading `exit_code:` merely FAILED (exit 1, no
// code), so a statically knowable mistake was counted as a product regression
// by every dashboard reading the exit status. The loader sees the step order —
// ATG2107 proves it — so the whole family is refused there.
func TestLoadBytes_AssertNeedsItsProducingStep(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			"exit_code with no command",
			specSteps("assert: {exit_code: 0}", "run: {command: echo}"),
			"assert.exit_code requires a preceding run/pty step",
		},
		{
			"stdout with no command",
			specSteps("assert: {stdout: {contains: hi}}"),
			"assert.stdout requires a preceding run/pty step",
		},
		{
			"status with no http step",
			specSteps("run: {command: echo}", "assert: {status: 200}"),
			"assert.status requires a preceding http step",
		},
		{
			"rows with no query",
			specSteps("run: {command: echo}", "assert: {rows: {contains: x}}"),
			"assert.rows requires a preceding query step",
		},
		{
			"grpc_status with no grpc step",
			specSteps("run: {command: echo}", "assert: {grpc_status: 0}"),
			"assert.grpc_status requires a preceding grpc step",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			mustReject(t, c.name, c.src, c.want)
		})
	}
}

// TestLoadBytes_AssertAfterItsProducingStepLoads pins the accept side: every
// ordering that CAN be fed must keep loading, including an assert fed by a step
// earlier in the scenario and a teardown assert fed by the scenario's steps.
func TestLoadBytes_AssertAfterItsProducingStepLoads(t *testing.T) {
	t.Parallel()
	ok := []string{
		specSteps("run: {command: echo}", "assert: {exit_code: 0}"),
		specSteps("pty: {command: sh, session: [{send: \"\"}]}", "assert: {stdout: {contains: hi}}"),
		specSteps("http: {runner: api, method: GET, path: /}", "assert: {status: 200}"),
		specSteps("query: {runner: db, sql: \"SELECT 1\"}", "assert: {rows: {contains: x}}"),
		specSteps("grpc: {runner: rpc, method: pkg.S/M}", "assert: {grpc_status: 0}"),
		// An assert fed by a step two positions earlier is still fed.
		specSteps("run: {command: echo}", "fixture: {file: f.txt, content: x}", "assert: {exit_code: 0}"),
		// A teardown assert is fed by the scenario's own steps.
		"version: \"1\"\nsuite:\n  name: x\nrunners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n    teardown:\n      - assert: {exit_code: 0}\n",
	}
	runners := "runners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\n  db: {type: db, dsn: \"sqlite:./a.db\"}\n  rpc: {type: grpc, target: \"127.0.0.1:1\"}\n"
	for i, src := range ok {
		if !strings.Contains(src, "runners:") {
			src = strings.Replace(src, "scenarios:", runners+"scenarios:", 1)
		}
		if _, err := LoadBytes("s.atago.yaml", []byte(src)); err != nil {
			t.Errorf("case %d was rejected: %v\n%s", i, err, src)
		}
	}
}

// suiteBlockSteps builds a spec whose suite.setup or suite.teardown holds the
// given steps, with one trivial scenario so the suite has something to run.
func suiteBlockSteps(block string, steps ...string) string {
	var b strings.Builder
	b.WriteString("version: \"1\"\nsuite:\n  name: x\n  " + block + ":\n")
	for _, s := range steps {
		b.WriteString("    - " + s + "\n")
	}
	b.WriteString("scenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert: {exit_code: 0}\n")
	return b.String()
}

// TestLoadBytes_SuiteBlockAssertNeedsItsProducingStep is a regression: the
// producing-step rule was applied to scenario steps and teardown but not to the
// suite lifecycle, where the mistake is worse. A suite block admits neither pty
// nor the runner-backed kinds, so a screen or status assertion there can never
// be fed by any ordering — and a suite.teardown failure preserves the verdict by
// contract, so such a spec printed SUITE TEARDOWN FAILED on every run while
// exiting 0 forever.
func TestLoadBytes_SuiteBlockAssertNeedsItsProducingStep(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			"setup stdout with no command",
			suiteBlockSteps("setup", "fixture: {file: f.txt, content: x}", "assert: {stdout: {contains: hi}}"),
			"suite.setup[1].assert.stdout requires a preceding run step",
		},
		{
			"setup exit_code with no command",
			suiteBlockSteps("setup", "assert: {exit_code: 0}"),
			"suite.setup[0].assert.exit_code requires a preceding run step",
		},
		{
			"setup screen can never be fed",
			suiteBlockSteps("setup", "assert: {screen: {contains: hi}}"),
			"pty steps are not allowed in a suite block",
		},
		{
			"teardown status can never be fed",
			suiteBlockSteps("teardown", "assert: {status: 200}"),
			"http steps are not allowed in a suite block",
		},
		{
			"teardown does not inherit the setup block's command",
			"version: \"1\"\nsuite:\n  name: x\n  setup:\n    - run: {command: echo}\n  teardown:\n    - assert: {stdout: {contains: hi}}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n      - assert: {exit_code: 0}\n",
			"suite.teardown[0].assert.stdout requires a preceding run step",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			mustReject(t, c.name, c.src, c.want)
		})
	}
}

// TestLoadBytes_SuiteBlockAssertAfterItsRunLoads pins the accept side: the
// build-then-check bootstrap a suite.setup exists for must keep loading, in both
// blocks, and filesystem-fed assertions stay unaffected.
func TestLoadBytes_SuiteBlockAssertAfterItsRunLoads(t *testing.T) {
	t.Parallel()
	ok := []string{
		suiteBlockSteps("setup", "run: {command: echo}", "assert: {exit_code: 0}"),
		suiteBlockSteps("setup", "run: {command: echo}", "assert: {stdout: {contains: hi}}"),
		suiteBlockSteps("teardown", "run: {command: echo}", "assert: {stdout: {contains: hi}}"),
		// A filesystem assertion has no unambiguous producer and is unaffected.
		suiteBlockSteps("setup", "fixture: {file: f.txt, content: x}", "assert: {file: {path: f.txt, contains: x}}"),
	}
	for i, src := range ok {
		if _, err := LoadBytes("s.atago.yaml", []byte(src)); err != nil {
			t.Errorf("case %d was rejected: %v\n%s", i, err, src)
		}
	}
}

// TestLoadBytes_StoreNeedsItsProducingStep is a regression: the commit that
// refused a context-less assertion named a `store from.header` after a run step
// as one of the three shapes of the same authoring mistake, and then fixed only
// the assertion. A store whose source no step produces stayed a runtime error
// (ATG4501, exit 4) while the identical assertion became a load error.
func TestLoadBytes_StoreNeedsItsProducingStep(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			"stdout with no command",
			specSteps("store: {name: v, from: {stdout: {matches: \"(.+)\"}}}", "run: {command: echo}"),
			"store.from.stdout requires a preceding run/pty step",
		},
		{
			"header with no http step",
			specSteps("run: {command: echo}", "store: {name: v, from: {header: X-Token}}"),
			"store.from.header requires a preceding http step",
		},
		{
			"body with no http step",
			specSteps("run: {command: echo}", "store: {name: v, from: {body: {matches: \"(.+)\"}}}"),
			"store.from.body requires a preceding http step",
		},
		{
			"rows with no query",
			specSteps("run: {command: echo}", "store: {name: v, from: {rows: {matches: \"(.+)\"}}}"),
			"store.from.rows requires a preceding query step",
		},
		{
			"message with no grpc call",
			specSteps("run: {command: echo}", "store: {name: v, from: {message: {matches: \"(.+)\"}}}"),
			"store.from.message requires a preceding grpc step",
		},
		{
			"suite block header can never be fed",
			suiteBlockSteps("setup", "store: {name: v, from: {header: X-Token}}"),
			"http steps are not allowed in a suite block",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			mustReject(t, c.name, c.src, c.want)
		})
	}
}

// TestLoadBytes_StoreAfterItsProducingStepLoads pins the accept side for stores,
// including the sources with no unambiguous producer.
func TestLoadBytes_StoreAfterItsProducingStepLoads(t *testing.T) {
	t.Parallel()
	ok := []string{
		specSteps("run: {command: echo}", "store: {name: v, from: {stdout: {matches: \"(.+)\"}}}"),
		specSteps("pty: {command: sh, session: [{send: \"\"}]}", "store: {name: v, from: {stdout: {matches: \"(.+)\"}}}"),
		specSteps("http: {runner: api, method: GET, path: /}", "store: {name: v, from: {header: X-Token}}"),
		specSteps("query: {runner: db, sql: \"SELECT 1\"}", "store: {name: v, from: {rows: {matches: \"(.+)\"}}}"),
		specSteps("grpc: {runner: rpc, method: pkg.S/M}", "store: {name: v, from: {message: {matches: \"(.+)\"}}}"),
		// A file source may read what a fixture wrote, so it needs no command.
		specSteps("fixture: {file: f.txt, content: \"{}\"}", "store: {name: v, from: {file: {path: f.txt, text: true}}}"),
		// A suite.setup store reading the block's own run output is the ordinary
		// bootstrap and must keep loading.
		suiteBlockSteps("setup", "run: {command: echo}", "store: {name: v, from: {stdout: {matches: \"(.+)\"}}}"),
		// A teardown store is fed by the scenario's steps.
		"version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n    teardown:\n      - store: {name: v, from: {stdout: {matches: \"(.+)\"}}}\n",
	}
	runners := "runners:\n  api: {type: http, base_url: \"http://127.0.0.1:1\"}\n  db: {type: db, dsn: \"sqlite:./a.db\"}\n  rpc: {type: grpc, target: \"127.0.0.1:1\"}\n"
	for i, src := range ok {
		if !strings.Contains(src, "runners:") {
			src = strings.Replace(src, "scenarios:", runners+"scenarios:", 1)
		}
		if _, err := LoadBytes("s.atago.yaml", []byte(src)); err != nil {
			t.Errorf("case %d was rejected: %v\n%s", i, err, src)
		}
	}
}

// TestLoadWithSource_AppliesTheProjectManifest is a regression: LoadWithSource
// is what `atago manifest` reads specs with, and it went straight to LoadBytes
// — skipping the directory-manifest discovery every other command performs. A
// spec under an atago.project.yaml therefore meant one thing to run/explain/doc
// and another to manifest, which reported no project_path and none of the
// configuration the file applies.
func TestLoadWithSource_AppliesTheProjectManifest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	proj := filepath.Join(dir, "atago.project.yaml")
	if err := os.WriteFile(proj, []byte("env:\n  FROM_PROJECT: \"1\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(dir, "s.atago.yaml")
	if err := os.WriteFile(specPath, []byte("version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	withSource, _, err := LoadWithSource(specPath)
	if err != nil {
		t.Fatalf("LoadWithSource: %v", err)
	}
	plain, err := Load(specPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if withSource.ProjectPath != plain.ProjectPath {
		t.Errorf("project path = %q via LoadWithSource, %q via Load; the two must agree", withSource.ProjectPath, plain.ProjectPath)
	}
	if withSource.Suite.Env["FROM_PROJECT"] != "1" {
		t.Errorf("the manifest's env was not applied: %v", withSource.Suite.Env)
	}
}

// TestLoadBytes_EmptyTag is a regression: an empty tag loaded cleanly and then
// corrupted every tag index — `atago list` rendered the column as ",smoke" and
// `atago doc` summarized "Tags: “ (1)" — while selecting nothing, since no
// usable --tag invocation names it. It is the remaining hole in the set
// semantics the duplicate-tag rule guards.
func TestLoadBytes_EmptyTag(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{`""`, `"   "`} {
		src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    tags: [" + tag + ", smoke]\n    steps:\n      - run: {command: echo}"
		_, err := LoadBytes("s.atago.yaml", []byte(src))
		if err == nil {
			t.Fatalf("tag %s was accepted", tag)
		}
		if !strings.Contains(err.Error(), "tag must not be empty") {
			t.Errorf("tag %s: err = %v, want an empty-tag rejection", tag, err)
		}
	}
}

// TestLoadBytes_ReadyStoreShadowsBuiltin is a regression: `store:` and `matrix:`
// are refused when they would bind a built-in name, and a service's ready.store
// is the third binding site the guard never learned — `ready: {file: f, store:
// workdir}` silently redefined ${workdir} for the rest of the scenario.
func TestLoadBytes_ReadyStoreShadowsBuiltin(t *testing.T) {
	t.Parallel()
	src := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - {name: s, command: ./srv, ready: {file: marker.txt, store: workdir}}\n    steps:\n      - run: {command: echo}"
	_, err := LoadBytes("s.atago.yaml", []byte(src))
	if err == nil {
		t.Fatal("a ready.store shadowing ${workdir} was accepted")
	}
	if !strings.Contains(err.Error(), "shadows a built-in variable") {
		t.Errorf("err = %v, want a built-in shadow rejection", err)
	}

	// A suite.setup service is the same binding site one scope up.
	suiteSrc := "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {name: s, command: ./srv, ready: {file: marker.txt, store: specdir}}\nscenarios:\n  - name: a\n    steps:\n      - run: {command: echo}"
	if _, err := LoadBytes("s.atago.yaml", []byte(suiteSrc)); err == nil {
		t.Error("a suite service ready.store shadowing ${specdir} was accepted")
	}

	// An ordinary capture name keeps loading.
	ok := "version: \"1\"\nsuite:\n  name: x\nscenarios:\n  - name: a\n    services:\n      - {name: s, command: ./srv, ready: {file: marker.txt, store: addr}}\n    steps:\n      - run: {command: echo}"
	if _, err := LoadBytes("s.atago.yaml", []byte(ok)); err != nil {
		t.Errorf("an ordinary ready.store name was rejected: %v", err)
	}
}

// TestLoadBytes_ScenarioServiceShadowsSuiteService is a regression: duplicate
// names are refused within a scope, but a scenario service could reuse a
// suite-level name, leaving a `signal:` step's target ambiguous with nothing
// said. Mock servers had the identical cross-scope hole.
func TestLoadBytes_ScenarioServiceShadowsSuiteService(t *testing.T) {
	t.Parallel()
	svc := "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {name: peer, command: ./peer}\nscenarios:\n  - name: a\n    services:\n      - {name: peer, command: ./peer}\n    steps:\n      - run: {command: echo}"
	_, err := LoadBytes("s.atago.yaml", []byte(svc))
	if err == nil {
		t.Fatal("a scenario service shadowing a suite service was accepted")
	}
	if !strings.Contains(err.Error(), "duplicate service name") {
		t.Errorf("err = %v, want a duplicate-service rejection", err)
	}

	mock := "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - mock_server: {name: api, routes: []}\nscenarios:\n  - name: a\n    mock_servers:\n      - {name: api, routes: []}\n    steps:\n      - run: {command: echo}"
	if _, err := LoadBytes("s.atago.yaml", []byte(mock)); err == nil {
		t.Error("a scenario mock server shadowing a suite mock server was accepted")
	}

	// Distinct names across the scopes stay legal.
	ok := "version: \"1\"\nsuite:\n  name: x\n  setup:\n    - service: {name: shared, command: ./peer}\nscenarios:\n  - name: a\n    services:\n      - {name: local, command: ./peer}\n    steps:\n      - run: {command: echo}"
	if _, err := LoadBytes("s.atago.yaml", []byte(ok)); err != nil {
		t.Errorf("distinct service names across scopes were rejected: %v", err)
	}
}
