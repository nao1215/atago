package fixture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nao1215/atago/internal/spec"
)

func TestWrite_Content(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := Write(&spec.Fixture{File: "data/users.csv", Content: "id,name\n1,Alice\n"}, dir, ""); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "data", "users.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "id,name\n1,Alice\n" {
		t.Errorf("content = %q", got)
	}
}

func TestWrite_Base64(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// "AAECAwQ=" decodes to bytes 00 01 02 03 04.
	if err := Write(&spec.Fixture{File: "data.bin", Base64: "AAECAwQ="}, dir, ""); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "data.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{0, 1, 2, 3, 4}; string(got) != string(want) {
		t.Errorf("bytes = %v, want %v", got, want)
	}
}

func TestWrite_Escape(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := Write(&spec.Fixture{File: "../escape.txt", Content: "x"}, dir, ""); err == nil {
		t.Fatal("Write() error = nil, want escape rejection")
	}
}

func TestWrite_From(t *testing.T) {
	t.Parallel()
	specDir := t.TempDir()
	// A committed binary fixture next to the spec, with NUL/high bytes that an
	// inline content fixture could not carry.
	srcBytes := []byte{0x50, 0x41, 0x52, 0x31, 0x00, 0xff} // "PAR1" + binary
	if err := os.WriteFile(filepath.Join(specDir, "data.parquet"), srcBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := Write(&spec.Fixture{File: "in/data.parquet", From: "data.parquet"}, dir, specDir); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "in", "data.parquet"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(srcBytes) {
		t.Errorf("copied bytes = %v, want %v", got, srcBytes)
	}
}

func TestWrite_FromMissing(t *testing.T) {
	t.Parallel()
	if err := Write(&spec.Fixture{File: "x", From: "nope.bin"}, t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("Write() error = nil, want missing-source error")
	}
}

func TestWrite_Symlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "target.txt"), []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Write(&spec.Fixture{File: "link.txt", Symlink: filepath.Join(dir, "target.txt")}, dir, ""); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	fi, err := os.Lstat(filepath.Join(dir, "link.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("link.txt is not a symlink (mode %s)", fi.Mode())
	}
	got, err := os.ReadFile(filepath.Join(dir, "link.txt")) // follows the link
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Errorf("read through link = %q, want payload", got)
	}
}

// Regression for issue #16: a symlink fixture whose target resolves outside the
// workdir must be rejected (relative ../ escape or an absolute outside path).
func TestWrite_SymlinkEscapeRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, target := range []string{"../../../etc/cron.d/x", "/etc/passwd"} {
		if err := Write(&spec.Fixture{File: "link", Symlink: target}, dir, ""); err == nil {
			t.Errorf("symlink to %q should be rejected", target)
			_ = os.Remove(filepath.Join(dir, "link"))
		}
	}
	// An in-workdir relative target is still allowed.
	if err := Write(&spec.Fixture{File: "ok.link", Symlink: "target.txt"}, dir, ""); err != nil {
		t.Errorf("in-workdir symlink target should be allowed: %v", err)
	}
}

// Regression for issue #16 (TOCTOU): the untrusted program under test plants a
// symlink in the workdir pointing at a sensitive outside file; a later file
// fixture writing that name must NOT follow the link and overwrite the target.
func TestWrite_RefusesToFollowPlantedSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Simulate the program-under-test planting out.txt -> outside.
	if err := os.Symlink(outside, filepath.Join(dir, "out.txt")); err != nil {
		t.Fatal(err)
	}
	err := Write(&spec.Fixture{File: "out.txt", Content: "overwritten"}, dir, "")
	if err == nil {
		t.Error("writing through a planted symlink should be refused")
	}
	got, rerr := os.ReadFile(outside)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(got) != "original" {
		t.Errorf("outside file was modified through the symlink: %q", got)
	}
}

// TestWrite_ModeMtimeRefusesPlantedSymlink is a security regression: a
// mode/mtime-only fixture operates in place and chmod/chtimes FOLLOW symlinks,
// so a link the program-under-test planted at the destination must be refused
// rather than re-permissioning a host file outside the workdir (issue #16).
func TestWrite_ModeMtimeRefusesPlantedSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "planted")); err != nil {
		t.Fatal(err)
	}
	err := Write(&spec.Fixture{File: "planted", Mode: "0777"}, dir, "")
	if err == nil {
		t.Error("chmod through a planted symlink should be refused")
	}
	fi, serr := os.Lstat(outside)
	if serr != nil {
		t.Fatal(serr)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("outside file perms changed through the symlink: %o, want 0600", perm)
	}
}

func TestWrite_Mode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := Write(&spec.Fixture{File: "ro.txt", Content: "x", Mode: "0444"}, dir, ""); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	fi, err := os.Stat(filepath.Join(dir, "ro.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o444 {
		t.Errorf("mode = %o, want 0444", perm)
	}
}

func TestWrite_Mtime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := Write(&spec.Fixture{File: "t.txt", Content: "x", Mtime: "2020-01-02T03:04:05Z"}, dir, ""); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	fi, err := os.Stat(filepath.Join(dir, "t.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fi.ModTime().UTC().Format("2006-01-02T15:04:05Z"); got != "2020-01-02T03:04:05Z" {
		t.Errorf("mtime = %q, want 2020-01-02T03:04:05Z", got)
	}
}

func TestWrite_BadModeAndMtime(t *testing.T) {
	t.Parallel()
	if err := Write(&spec.Fixture{File: "x", Content: "y", Mode: "rwx"}, t.TempDir(), ""); err == nil {
		t.Error("Write() with non-octal mode: error = nil, want error")
	}
	if err := Write(&spec.Fixture{File: "x", Content: "y", Mtime: "not-a-time"}, t.TempDir(), ""); err == nil {
		t.Error("Write() with bad mtime: error = nil, want error")
	}
}
