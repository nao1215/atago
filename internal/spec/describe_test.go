package spec

import (
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

// TestGeneratedArtifacts covers the three sources a scenario can generate a file
// from: a file exists:true assertion, an image assertion, and a cdp screenshot
// action (#56).
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
			// A non-generating assertion must not add anything.
			{Assert: &Assert{Stdout: &StreamAssert{Contains: StringList{"x"}}}},
			// exists:false is a negative check, not a generated artifact.
			{Assert: &Assert{File: &FileAssert{Path: "gone.txt", Exists: boolPtr(false)}}},
			// A duplicate path is de-duplicated.
			{Assert: &Assert{Image: &ImageAssert{Path: "thumb.png"}}},
		},
	}
	got := GeneratedArtifacts(sc)
	want := []string{"out.txt", "thumb.png", "shot.png", "logs/out.log", "logs/err.log", "downloads/report.pdf"}
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
