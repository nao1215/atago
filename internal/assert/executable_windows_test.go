//go:build windows

package assert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

// TestIsExecutable_ByExtension pins the Windows rule. Before this the answer
// came from the mode bits, which Windows does not have — Go synthesizes 0666,
// or 0444 for a read-only file — so every file on the platform answered "not
// executable": `executable: true` could not be satisfied by a real program and
// `executable: false` passed vacuously on one.
func TestIsExecutable_ByExtension(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, tt := range []struct {
		name string
		want bool
	}{
		{"mytool.exe", true},
		{"MYTOOL.EXE", true}, // Windows extensions are case-insensitive
		{"build.bat", true},
		{"build.cmd", true},
		{"notes.txt", false},
		{"script.sh", false}, // a POSIX script is not runnable by name here
		{"README", false},    // no extension at all
	} {
		path := filepath.Join(dir, tt.name)
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := isExecutable(info, path); got != tt.want {
			t.Errorf("isExecutable(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestIsExecutable_HonorsPATHEXT proves the list is read from the host rather
// than hard-coded, and that an empty PATHEXT falls back to the stock value
// instead of calling every file unrunnable.
func TestIsExecutable_HonorsPATHEXT(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "run.ps1")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT")
	if isExecutable(info, path) {
		t.Error("run.ps1 is executable with a PATHEXT that does not list .PS1")
	}
	t.Setenv("PATHEXT", ".EXE;.PS1")
	if !isExecutable(info, path) {
		t.Error("run.ps1 is not executable with a PATHEXT that lists .PS1")
	}
	t.Setenv("PATHEXT", "")
	exe := filepath.Join(dir, "t.exe")
	if err := os.WriteFile(exe, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	exeInfo, err := os.Stat(exe)
	if err != nil {
		t.Fatal(err)
	}
	if !isExecutable(exeInfo, exe) {
		t.Error("t.exe is not executable with PATHEXT unset; the stock fallback did not apply")
	}
}

// TestCheckFileExecutable_ExplainsTheWindowsRule: a spec author arriving from
// POSIX will reach for chmod, so the failure has to say that the extension is
// what decides here — and must not quote a mode Windows synthesized.
func TestCheckFileExecutable_ExplainsTheWindowsRule(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tool.sh"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	yes := true
	res := checkFileExecutable(&spec.FileAssert{Path: "tool.sh", Executable: &yes}, dir, filepath.Join(dir, "tool.sh"))
	if res == nil || res.Hint == "" {
		t.Fatalf("checkFileExecutable = %+v, want a failure with a hint", res)
	}
	if want := "PATHEXT"; !strings.Contains(res.Hint, want) {
		t.Errorf("hint = %q, want it to mention %q", res.Hint, want)
	}
	if strings.Contains(res.Actual, "mode ") {
		t.Errorf("actual = %q, want it to name the extension rather than a synthesized mode", res.Actual)
	}
	if want := `extension ".sh"`; !strings.Contains(res.Actual, want) {
		t.Errorf("actual = %q, want it to mention %q", res.Actual, want)
	}
}
