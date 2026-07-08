package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

func TestInit_CreatesValidSpec(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.atago.yaml")

	var out, errb bytes.Buffer
	if got := Main([]string{"init", path}, &out, &errb); got != ExitOK {
		t.Fatalf("init exit = %d, want %d (stderr=%s)", got, ExitOK, errb.String())
	}
	// The scaffold must itself be a valid, loadable spec.
	if _, err := loader.Load(path); err != nil {
		t.Fatalf("scaffold is not a valid spec: %v", err)
	}
	// The starter command must run through the shell so it is portable to
	// Windows, where `echo` is a shell builtin and not a real executable (#42).
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("shell: true")) {
		t.Errorf("starter spec should use shell: true for cross-platform portability:\n%s", data)
	}
}

// TestInit_RejectsExtraPaths proves init refuses more than one path argument
// rather than silently writing only the first and dropping the rest, which would
// make a user believe both files were scaffolded.
func TestInit_RejectsExtraPaths(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.atago.yaml")
	second := filepath.Join(dir, "second.atago.yaml")

	var out, errb bytes.Buffer
	if got := Main([]string{"init", first, second}, &out, &errb); got != ExitConfig {
		t.Fatalf("init with two paths exit = %d, want %d", got, ExitConfig)
	}
	if !bytes.Contains(errb.Bytes(), []byte("too many paths")) {
		t.Errorf("stderr should explain the extra-path error, got: %s", errb.String())
	}
	// Neither file is created when the invocation is rejected.
	if _, err := os.Stat(first); err == nil {
		t.Error("first path should not be written when init is rejected")
	}
	if _, err := os.Stat(second); err == nil {
		t.Error("second path should not be written when init is rejected")
	}
}

func TestInit_RefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.atago.yaml")
	if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if got := Main([]string{"init", path}, &out, &errb); got != ExitConfig {
		t.Fatalf("init exit = %d, want %d", got, ExitConfig)
	}
	// --force overwrites.
	out.Reset()
	errb.Reset()
	if got := Main([]string{"init", "--force", path}, &out, &errb); got != ExitOK {
		t.Fatalf("init --force exit = %d, want %d", got, ExitOK)
	}
	data, _ := os.ReadFile(path)
	if string(data) == "existing" {
		t.Error("--force did not overwrite the file")
	}
}
