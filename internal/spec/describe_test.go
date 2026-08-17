package spec

import (
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

// TestGeneratedArtifacts covers every source a scenario can generate a file
// from: a file exists:true assertion, image and pdf assertions, a cdp
// screenshot action, redirect targets, and teardown redirects (#56).
func TestGeneratedArtifacts(t *testing.T) {
	t.Parallel()
	sc := &Scenario{
		Steps: []Step{
			{Assert: &Assert{File: &FileAssert{Path: "out.txt", Exists: boolPtr(true)}}},
			{Assert: &Assert{Image: &ImageAssert{Path: "thumb.png", SimilarTo: "base.png"}}},
			{CDP: &CDP{Runner: "web", Actions: []CDPAction{{Screenshot: &CDPScreenshot{Path: "shot.png"}}}}},
			// Redirect targets are declared outputs too.
			{Run: &Run{Command: "mycli", StdoutTo: "logs/out.log", StderrTo: "logs/err.log"}},
			{HTTP: &HTTP{Method: "GET", Path: "/report.pdf", BodyTo: "downloads/report.pdf"}},
			// A pdf assertion inspects an output the tool wrote, like image; it
			// used to be the one inspecting target left out.
			{Assert: &Assert{PDF: &PDFAssert{Path: "report.pdf"}}},
			// A non-generating assertion must not add anything.
			{Assert: &Assert{Stdout: &StreamAssert{Contains: StringList{"x"}}}},
			// exists:false is a negative check, not a generated artifact.
			{Assert: &Assert{File: &FileAssert{Path: "gone.txt", Exists: boolPtr(false)}}},
			// A duplicate path is de-duplicated.
			{Assert: &Assert{Image: &ImageAssert{Path: "thumb.png"}}},
		},
		// Teardown always runs and its redirects land in the workdir like any
		// other step's; they used to be invisible here.
		Teardown: []Step{
			{Run: &Run{Command: "mycli audit", StdoutTo: "logs/audit.log"}},
		},
	}
	got := GeneratedArtifacts(sc)
	want := []string{"out.txt", "thumb.png", "shot.png", "logs/out.log", "logs/err.log", "downloads/report.pdf", "report.pdf", "logs/audit.log"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("GeneratedArtifacts = %v, want %v", got, want)
	}
}

func TestSecurityNotes(t *testing.T) {
	t.Parallel()
	sc := &Scenario{
		Services: []Service{
			{Name: "peer", Shell: Bool(true), Command: "curl https://api.example.com/serve"},
		},
		Steps: []Step{
			{Run: &Run{Command: "echo hi", Shell: Bool(true)}},
			{Run: &Run{Command: "wget https://example.com/file"}},
			{HTTP: &HTTP{Method: "GET", Path: "/x"}},
			{GRPC: &GRPC{Runner: "g", Method: "pkg.S/M"}},
			{CDP: &CDP{Runner: "web", Actions: []CDPAction{{Navigate: "http://x"}}}},
			// Duplicate CDP note must be de-duplicated.
			{CDP: &CDP{Runner: "web", Actions: []CDPAction{{Click: "#a"}}}},
		},
	}
	got := SecurityNotes(sc, nil)
	for _, want := range []string{
		"shell execution enabled (service peer): curl https://api.example.com/serve",
		"network access (service peer): curl https://api.example.com/serve",
		"shell execution enabled: echo hi",
		"network access: wget https://example.com/file",
		"network access: HTTP request",
		"network access: gRPC pkg.S/M",
		"browser automation (CDP) via web",
	} {
		if !contains(got, want) {
			t.Errorf("SecurityNotes missing %q\n got: %v", want, got)
		}
	}
	// The duplicate CDP note appears once.
	n := 0
	for _, s := range got {
		if s == "browser automation (CDP) via web" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("CDP security note appears %d times, want 1", n)
	}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// TestSecurityNotes_RemoteRunnersAreNetworkAccess is a regression: the summary
// names network access for http, grpc, and a run whose COMMAND looks networky,
// and said nothing about the two runners that reach a host by construction. A
// `run:` through an ssh runner logs into another machine and appeared as a bare
// local command; a `query:` through a db runner pointed at postgres or mysql
// appeared as nothing at all. Both are egress the network allowlist governs, so
// a reviewer reading this summary to see what a spec touches was told less than
// the policy already knew.
func TestSecurityNotes_RemoteRunnersAreNetworkAccess(t *testing.T) {
	t.Parallel()
	runners := map[string]Runner{
		"box":    {Type: "ssh", Host: "shell.example", User: "deploy"},
		"pg":     {Type: "db", DSN: "postgres://u:p@db.example:5432/app"},
		"my":     {Type: "db", DSN: "mysql://u:p@db.example:3306/app"},
		"local":  {Type: "db", DSN: "sqlite:${workdir}/a.db"},
		"local2": {Type: "db", Driver: "sqlite", DSN: "./a.db"},
	}
	sc := &Scenario{Steps: []Step{
		{Run: &Run{Runner: "box", Command: "uptime"}},
		{Query: &Query{Runner: "pg", SQL: "SELECT 1"}},
		{Query: &Query{Runner: "my", SQL: "SELECT 2"}},
		{Query: &Query{Runner: "local", SQL: "SELECT 3"}},
		{Query: &Query{Runner: "local2", SQL: "SELECT 4"}},
	}}
	got := strings.Join(SecurityNotes(sc, runners), "\n")
	for _, want := range []string{
		"network access (ssh box): uptime",
		"network access: SQL query via pg",
		"network access: SQL query via my",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("security notes missing %q:\n%s", want, got)
		}
	}
	// A file-backed database reaches no host, so claiming it does would train a
	// reviewer to ignore the line.
	for _, absent := range []string{"via local", "via local2"} {
		if strings.Contains(got, absent) {
			t.Errorf("security notes call a sqlite runner network access (%q):\n%s", absent, got)
		}
	}
}

// TestSecurityNotes_PTYSteps is a regression: the step switch had no pty case,
// so an interactive step was invisible to the security summary no matter what
// it did — a shell-enabled command, a command that dials out over ssh, host
// commands run mid-session by exec actions, and ${env:} reads in the command,
// env, and typed input all produced zero notes while the equivalent run step
// produced one for each.
func TestSecurityNotes_PTYSteps(t *testing.T) {
	t.Parallel()
	sc := &Scenario{Steps: []Step{
		{PTY: &PTY{Command: "mytui --menu", Shell: Bool(true)}},
		{PTY: &PTY{Command: "ssh admin@backend.example"}},
		{PTY: &PTY{
			Command: "watcher ${env:WATCH_DIR}",
			Env:     map[string]string{"TOKEN": "${env:API_TOKEN}"},
			Session: []PTYAction{
				{Send: SendText("${env:PASSWORD}\r")},
				{Send: &PTYSend{Paste: strp("${env:PASTED_SECRET}")}},
				{Exec: &PTYExec{Command: "curl https://hook.example/fire"}},
				{Exec: &PTYExec{Command: "touch marker && date", Shell: Bool(true)}},
			},
		}},
	}}
	got := SecurityNotes(sc, nil)
	for _, want := range []string{
		"shell execution enabled: mytui --menu",
		"network access: ssh admin@backend.example",
		"host environment read: ${env:WATCH_DIR}",
		"host environment read: ${env:API_TOKEN}",
		"host environment read: ${env:PASSWORD}",
		"host environment read: ${env:PASTED_SECRET}",
		"network access (pty exec): curl https://hook.example/fire",
		"shell execution enabled (pty exec): touch marker && date",
	} {
		if !contains(got, want) {
			t.Errorf("SecurityNotes missing %q\n got: %v", want, got)
		}
	}
}

// TestSecurityNotes_TeardownSteps is a regression: the walk covered services
// and steps but never teardown, and teardown steps always run — so a scenario
// whose only egress was a cleanup call (an HTTP DELETE, a curl through the
// shell, a command sent over ssh, a DELETE against a remote database) reported
// no security notes at all.
func TestSecurityNotes_TeardownSteps(t *testing.T) {
	t.Parallel()
	runners := map[string]Runner{
		"box": {Type: "ssh", Host: "h.example", User: "u"},
		"pg":  {Type: "db", DSN: "postgres://u:p@db.example:5432/app"},
	}
	sc := &Scenario{
		Steps: []Step{{Run: &Run{Command: "echo hi"}}},
		Teardown: []Step{
			{HTTP: &HTTP{Method: "DELETE", Path: "/resource/1"}},
			{Run: &Run{Command: "curl https://api.example/cleanup", Shell: Bool(true)}},
			{Run: &Run{Runner: "box", Command: "rm -f /tmp/deploy.lock"}},
			{Query: &Query{Runner: "pg", SQL: "DELETE FROM sessions"}},
		},
	}
	got := SecurityNotes(sc, runners)
	for _, want := range []string{
		"network access: HTTP request",
		"shell execution enabled: curl https://api.example/cleanup",
		"network access: curl https://api.example/cleanup",
		"network access (ssh box): rm -f /tmp/deploy.lock",
		"network access: SQL query via pg",
	} {
		if !contains(got, want) {
			t.Errorf("SecurityNotes missing teardown note %q\n got: %v", want, got)
		}
	}
}

// TestSuiteSecurityNotes is a regression: only scenarios had a security
// summary, so a suite whose setup curls a seed file through the shell and
// starts an ssh-tunnel service — and whose teardown curls a purge endpoint —
// reported no security notes anywhere, while the identical steps inside a
// scenario produce one for each.
func TestSuiteSecurityNotes(t *testing.T) {
	t.Parallel()
	su := &Suite{
		Setup: []Step{
			{Run: &Run{Command: "curl https://seed.example/data", Shell: Bool(true)}},
			{Service: &Service{Name: "relay", Command: "ssh -N -L 8080:internal.example:80 jump.example"}},
		},
		Teardown: []Step{
			{Run: &Run{Command: "curl https://api.example/purge"}},
		},
	}
	got := SuiteSecurityNotes(su, nil)
	for _, want := range []string{
		"shell execution enabled: curl https://seed.example/data",
		"network access: curl https://seed.example/data",
		"network access (service relay): ssh -N -L 8080:internal.example:80 jump.example",
		"network access: curl https://api.example/purge",
	} {
		if !contains(got, want) {
			t.Errorf("SuiteSecurityNotes missing %q\n got: %v", want, got)
		}
	}
	// A quiet suite says nothing: the section only exists when there is
	// something to review.
	if got := SuiteSecurityNotes(&Suite{Setup: []Step{{Run: &Run{Command: "echo hi"}}}}, nil); len(got) != 0 {
		t.Errorf("SuiteSecurityNotes for a quiet suite = %v, want none", got)
	}
}

// TestSecurityNotes_RunnerEnvReads is a regression: the engine expands ${name}
// in a runner's base_url, dsn, target, cwd, and ssh credentials, so a
// ${env:NAME} in any of them reads the invoking host's environment — and the
// summary walked steps and services only, never the runner table. A database
// password pulled from the host environment produced no note at all.
func TestSecurityNotes_RunnerEnvReads(t *testing.T) {
	t.Parallel()
	runners := map[string]Runner{
		"api": {Type: "http", BaseURL: "https://${env:API_HOST}/v1"},
		"pg":  {Type: "db", DSN: "postgres://u:${env:DB_PASSWORD}@db.example:5432/app"},
		"rpc": {Type: "grpc", Target: "${env:RPC_TARGET}"},
		"box": {Type: "ssh", Host: "${env:SSH_HOST}", User: "deploy", KeyFile: "${env:SSH_KEY}"},
	}
	sc := &Scenario{Steps: []Step{
		{HTTP: &HTTP{Method: "GET", Path: "/health", Runner: "api"}},
		{Query: &Query{Runner: "pg", SQL: "SELECT 1"}},
		{GRPC: &GRPC{Runner: "rpc", Method: "pkg.S/M"}},
		{Run: &Run{Runner: "box", Command: "uptime"}},
	}}
	got := SecurityNotes(sc, runners)
	for _, want := range []string{
		"host environment read: ${env:API_HOST}",
		"host environment read: ${env:DB_PASSWORD}",
		"host environment read: ${env:RPC_TARGET}",
		"host environment read: ${env:SSH_HOST}",
		"host environment read: ${env:SSH_KEY}",
	} {
		if !contains(got, want) {
			t.Errorf("SecurityNotes missing %q\n got: %v", want, got)
		}
	}

	// A runner the scenario never uses contributes nothing: the notes describe
	// what this scenario touches.
	unused := map[string]Runner{"idle": {Type: "http", BaseURL: "https://${env:UNUSED}/"}}
	if notes := SecurityNotes(&Scenario{Steps: []Step{{Run: &Run{Command: "echo hi"}}}}, unused); len(notes) != 0 {
		t.Errorf("an unused runner contributed notes: %v", notes)
	}
}

// TestSecurityNotes_HTTPNamesItsRunner is a regression: every other step kind
// names the runner that carried it, while an http step's note was the constant
// "network access: HTTP request" — so requests to two different hosts
// de-duplicated into one anonymous line.
func TestSecurityNotes_HTTPNamesItsRunner(t *testing.T) {
	t.Parallel()
	runners := map[string]Runner{
		"internal": {Type: "http", BaseURL: "http://127.0.0.1:8080"},
		"billing":  {Type: "http", BaseURL: "https://billing.example.com"},
	}
	sc := &Scenario{Steps: []Step{
		{HTTP: &HTTP{Method: "GET", Path: "/health", Runner: "internal"}},
		{HTTP: &HTTP{Method: "POST", Path: "/charge", Runner: "billing"}},
	}}
	got := SecurityNotes(sc, runners)
	for _, want := range []string{
		"network access: HTTP request via internal",
		"network access: HTTP request via billing",
	} {
		if !contains(got, want) {
			t.Errorf("SecurityNotes missing %q\n got: %v", want, got)
		}
	}
}

func TestRunHost(t *testing.T) {
	t.Parallel()
	runners := map[string]Runner{
		"box":   {Type: "ssh", Host: "shell.example", User: "deploy"},
		"local": {Type: "cmd"},
	}
	cases := []struct {
		name string
		run  *Run
		want string
	}{
		{"ssh runner", &Run{Runner: "box", Command: "uptime"}, "ssh box"},
		{"cmd runner", &Run{Runner: "local", Command: "uptime"}, ""},
		{"no runner", &Run{Command: "uptime"}, ""},
		{"unknown runner", &Run{Runner: "gone", Command: "uptime"}, ""},
		{"nil run", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := RunHost(c.run, runners); got != c.want {
				t.Errorf("RunHost = %q, want %q", got, c.want)
			}
		})
	}
}

// TestSecurityNotes_InsecureHostKey is a regression: the loader refuses an ssh
// runner that decides neither way about host-key verification, which is how
// much the choice matters — and then the security summary said nothing about
// which way it was decided. A reviewer reading it to see what a spec does was
// told the scenario reaches another host, but not that it accepts whatever key
// that host presents.
func TestSecurityNotes_InsecureHostKey(t *testing.T) {
	t.Parallel()
	runners := map[string]Runner{
		"loose":  {Type: "ssh", Host: "h.example", User: "u", InsecureHostKey: true},
		"strict": {Type: "ssh", Host: "h.example", User: "u", KnownHosts: "known_hosts"},
	}
	sc := &Scenario{Steps: []Step{
		{Run: &Run{Runner: "loose", Command: "uptime"}},
		{Run: &Run{Runner: "strict", Command: "uname"}},
	}}
	got := strings.Join(SecurityNotes(sc, runners), "\n")
	if !strings.Contains(got, `ssh host key verification disabled (runner "loose")`) {
		t.Errorf("security notes do not mention the disabled host-key check:\n%s", got)
	}
	// A runner that verifies says nothing extra: the note is about the opt-out,
	// and one line per ordinary runner would train a reader to skip the section.
	if strings.Contains(got, `"strict"`) {
		t.Errorf("security notes flag a runner that does verify:\n%s", got)
	}
}
