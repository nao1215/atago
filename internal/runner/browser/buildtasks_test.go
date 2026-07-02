package browser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chromedp/chromedp/kb"

	"github.com/nao1215/atago/internal/spec"
)

// TestBuildTasks_NewActions verifies each extended CDP action produces a task and
// the capturing actions register a capture, without needing a live browser (#50).
func TestBuildTasks_NewActions(t *testing.T) {
	t.Parallel()
	actions := []spec.CDPAction{
		{Navigate: "https://example.com"},
		{WaitVisible: "#a"},
		{WaitHidden: "#spinner"},
		{Click: "#btn"},
		{Press: &spec.CDPPress{Selector: "#in", Key: "Enter"}},
		{Select: &spec.CDPSelect{Selector: "#s", Value: "b"}},
		{Check: "#c"},
		{Uncheck: "#c"},
		{Text: "#out"},
		{Title: true},
		{Attribute: &spec.CDPAttribute{Selector: "#a", Name: "href"}},
		{Eval: "1+1"},
	}
	tasks, caps, shots, err := buildTasks(actions, t.TempDir())
	if err != nil {
		t.Fatalf("buildTasks: %v", err)
	}
	if len(tasks) != len(actions) {
		t.Errorf("tasks = %d, want %d", len(tasks), len(actions))
	}
	// text, title, attribute, and eval each register a capture (4 total).
	if len(caps) != 4 {
		t.Errorf("captures = %d, want 4", len(caps))
	}
	if len(shots) != 0 {
		t.Errorf("shots = %d, want 0 (no screenshot action)", len(shots))
	}
}

func TestBuildTasks_ScreenshotResolvesPathAndRegistersShot(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	_, _, shots, err := buildTasks([]spec.CDPAction{
		{Navigate: "https://example.com"},
		{Screenshot: &spec.CDPScreenshot{Path: "shot.png"}},
	}, wd)
	if err != nil {
		t.Fatalf("buildTasks: %v", err)
	}
	if len(shots) != 1 {
		t.Fatalf("shots = %d, want 1", len(shots))
	}
	if shots[0].path != filepath.Join(wd, "shot.png") {
		t.Errorf("screenshot path = %q, want it resolved against the workdir", shots[0].path)
	}
	if shots[0].buf == nil {
		t.Errorf("screenshot buffer not registered")
	}
}

func TestBuildTasks_UnknownActionErrors(t *testing.T) {
	t.Parallel()
	if _, _, _, err := buildTasks([]spec.CDPAction{{}}, t.TempDir()); err == nil {
		t.Fatal("expected an error for an action with no fields set")
	}
}

// TestBuildTasks_ScreenshotTraversalRejected proves a screenshot output path may
// not escape the scenario workdir.
func TestBuildTasks_ScreenshotTraversalRejected(t *testing.T) {
	t.Parallel()
	_, _, _, err := buildTasks([]spec.CDPAction{
		{Navigate: "https://example.com"},
		{Screenshot: &spec.CDPScreenshot{Path: "../shot.png"}},
	}, t.TempDir())
	if err == nil {
		t.Fatal("screenshot path escaping the workdir was accepted")
	}
	if !strings.Contains(err.Error(), "escapes the scenario workdir") {
		t.Errorf("error %q should explain the containment failure", err)
	}
}

func TestResolveKey(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"Enter": kb.Enter,
		"Tab":   kb.Tab,
		"a":     "a",
		"1":     "1",
	}
	for in, want := range cases {
		got, err := resolveKey(in)
		if err != nil {
			t.Errorf("resolveKey(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("resolveKey(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := resolveKey("NotAKey"); err == nil {
		t.Errorf("resolveKey should reject an unsupported multi-char key")
	}
}

func TestSetChecked_BuildsSelectorSafeJS(t *testing.T) {
	t.Parallel()
	// A quoted/special-character selector must be JSON-encoded, not string-glued,
	// so the generated JS stays valid.
	_, _, _, err := buildTasks([]spec.CDPAction{
		{Navigate: "https://example.com"},
		{Check: `input[name="agree"]`},
	}, t.TempDir())
	if err != nil {
		t.Fatalf("buildTasks with a quoted selector: %v", err)
	}
}

func TestBuildTasks_TitleAndAttributeCaptureLast(t *testing.T) {
	t.Parallel()
	// The last capturing action's value feeds the `value` assertion; verify a
	// trailing attribute capture is present so its value would be the one used.
	_, caps, _, err := buildTasks([]spec.CDPAction{
		{Text: "#a"},
		{Attribute: &spec.CDPAttribute{Selector: "#b", Name: "data-id"}},
	}, t.TempDir())
	if err != nil {
		t.Fatalf("buildTasks: %v", err)
	}
	if len(caps) != 2 {
		t.Fatalf("captures = %d, want 2", len(caps))
	}
	// The trailing capture must be a string capture (attribute), not JS.
	if caps[len(caps)-1].str == nil {
		t.Errorf("last capture should be a string (attribute value)")
	}
}

// TestBuildTasks_Upload confirms an upload action requires an existing,
// workdir-confined file and registers a task (#75).
func TestBuildTasks_Upload(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	if err := os.WriteFile(filepath.Join(wd, "f.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Present file: a task is built.
	tasks, _, _, err := buildTasks([]spec.CDPAction{{Upload: &spec.CDPUpload{Selector: "#file", File: "f.txt"}}}, wd)
	if err != nil {
		t.Fatalf("buildTasks upload: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("tasks = %d, want 1", len(tasks))
	}
	// Missing file: an error, not a silent no-op.
	if _, _, _, err := buildTasks([]spec.CDPAction{{Upload: &spec.CDPUpload{Selector: "#file", File: "missing.txt"}}}, wd); err == nil {
		t.Error("expected an error for a missing upload file")
	}
	// Escaping file path: rejected by workdir confinement.
	if _, _, _, err := buildTasks([]spec.CDPAction{{Upload: &spec.CDPUpload{Selector: "#file", File: "../etc/passwd"}}}, wd); err == nil {
		t.Error("expected a confinement error for an escaping upload path")
	}
}

// TestBuildTasks_Download confirms a download action registers a task and a
// capture and confines its destination directory (#75).
func TestBuildTasks_Download(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	tasks, caps, _, err := buildTasks([]spec.CDPAction{{Download: &spec.CDPDownload{Click: "#dl", Dir: "downloads"}}}, wd)
	if err != nil {
		t.Fatalf("buildTasks download: %v", err)
	}
	if len(tasks) != 1 || len(caps) != 1 {
		t.Errorf("tasks=%d caps=%d, want 1 and 1", len(tasks), len(caps))
	}
	// A destination escaping the workdir is rejected.
	if _, _, _, err := buildTasks([]spec.CDPAction{{Download: &spec.CDPDownload{Click: "#dl", Dir: "../out"}}}, wd); err == nil {
		t.Error("expected a confinement error for an escaping download dir")
	}
}
