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
	got := SecurityNotes(sc)
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
