package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveWorkdirPath(t *testing.T) {
	t.Parallel()
	// A real temp dir is an absolute path in the host's own form (a drive letter on
	// Windows, a leading slash on Unix), so the absolute-path cases are exercised
	// portably instead of with a Unix-only "/work/scn" literal.
	root := t.TempDir()
	tests := []struct {
		name    string
		in      string
		wantErr bool
		want    string
	}{
		{name: "plain relative", in: "out.txt", want: filepath.Join(root, "out.txt")},
		{name: "nested relative", in: "sub/out.txt", want: filepath.Join(root, "sub", "out.txt")},
		{name: "dot relative", in: "./out.txt", want: filepath.Join(root, "out.txt")},
		{name: "absolute inside root", in: filepath.Join(root, "out.txt"), want: filepath.Join(root, "out.txt")},
		{name: "parent traversal", in: "../escape.txt", wantErr: true},
		{name: "deep traversal", in: "sub/../../escape.txt", wantErr: true},
		{name: "absolute outside root", in: filepath.Join(filepath.Dir(root), "outside.txt"), wantErr: true},
		{name: "sibling prefix not contained", in: root + "-other", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ResolveWorkdirPath("assert.file.path", root, tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ResolveWorkdirPath(%q) error = nil, want error", tt.in)
				}
				if !strings.Contains(err.Error(), "assert.file.path") || !strings.Contains(err.Error(), "scenario workdir") {
					t.Errorf("error %q should name the field and root", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveWorkdirPath(%q) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ResolveWorkdirPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveSpecPath(t *testing.T) {
	t.Parallel()
	root := filepath.FromSlash("/repo/spec")
	if _, err := ResolveSpecPath("assert.snapshot", root, "snapshots/out.txt"); err != nil {
		t.Fatalf("in-spec snapshot path rejected: %v", err)
	}
	_, err := ResolveSpecPath("assert.snapshot", root, "../snapshots/out.txt")
	if err == nil {
		t.Fatal("snapshot path escaping the spec directory was accepted")
	}
	if !strings.Contains(err.Error(), "spec directory") {
		t.Errorf("error %q should name the spec directory", err)
	}
}

// TestResolve_RelativeRoot covers a spec loaded by a bare filename, where the
// spec/workdir root is "." — a relative in-root path must still be accepted, and
// a "../" escape must still be rejected.
func TestResolve_RelativeRoot(t *testing.T) {
	t.Parallel()
	if got, err := ResolveSpecPath("assert.snapshot", ".", "out.snap"); err != nil {
		t.Fatalf("relative in-root path rejected under root %q: %v", ".", err)
	} else if got != "out.snap" {
		t.Errorf("resolved = %q, want %q", got, "out.snap")
	}
	if _, err := ResolveSpecPath("assert.snapshot", ".", "../out.snap"); err == nil {
		t.Fatal("../ escape accepted under a relative root")
	}
	if _, err := ResolveWorkdirPath("f", ".", "sub/out.txt"); err != nil {
		t.Fatalf("nested relative path rejected under root %q: %v", ".", err)
	}
}

// TestReadFileNoFollow verifies a leaf symlink pointing outside the root is
// refused (issue #16): the untrusted program under test could plant such a link
// at an assertion/snapshot read target to disclose an arbitrary host file. A
// plain regular file inside the root is read normally.
func TestReadFileNoFollow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	root := t.TempDir()

	regular := filepath.Join(root, "real.txt")
	if err := os.WriteFile(regular, []byte("in-root"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadFileNoFollow(regular); err != nil || string(got) != "in-root" {
		t.Fatalf("ReadFileNoFollow(regular) = %q, %v; want %q, nil", got, err, "in-root")
	}

	// A host secret outside the root, and a link to it planted inside the root.
	secret := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secret, []byte("top-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "leak.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFileNoFollow(link)
	if err == nil {
		t.Fatalf("ReadFileNoFollow followed the symlink and read %q; want error", got)
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("error %q should name the refused symlink", err)
	}
}

// TestWriteFileNoFollow verifies a leaf symlink at the write target is refused
// (so a redirect/snapshot write cannot clobber a host file through a link the
// program under test planted), while a fresh write and an overwrite of a plain
// regular file both succeed.
func TestWriteFileNoFollow(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	root := t.TempDir()

	fresh := filepath.Join(root, "out.txt")
	if err := WriteFileNoFollow(fresh, []byte("v1"), 0o600); err != nil {
		t.Fatalf("fresh write: %v", err)
	}
	if err := WriteFileNoFollow(fresh, []byte("v2"), 0o600); err != nil {
		t.Fatalf("overwrite of regular file: %v", err)
	}
	if got, err := os.ReadFile(fresh); err != nil || string(got) != "v2" {
		t.Fatalf("after overwrite = %q, %v; want %q", got, err, "v2")
	}

	// A host file outside the root must not be clobbered through a planted link.
	victim := filepath.Join(t.TempDir(), "victim.txt")
	if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "redirect.txt")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileNoFollow(link, []byte("pwned"), 0o600); err == nil {
		t.Fatal("WriteFileNoFollow wrote through the symlink; want error")
	}
	if got, _ := os.ReadFile(victim); string(got) != "original" {
		t.Errorf("host file was modified through the symlink: %q", got)
	}
}

func TestResolveWorkdirPath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific separator handling")
	}
	root := `C:\work\scn`
	if _, err := ResolveWorkdirPath("f", root, `sub\out.txt`); err != nil {
		t.Fatalf("in-workdir windows path rejected: %v", err)
	}
	// A spec written for portability uses forward slashes; on Windows they must
	// resolve to the native separator and land inside the workdir.
	got, err := ResolveWorkdirPath("f", root, "sub/out.txt")
	if err != nil {
		t.Fatalf("forward-slash windows path rejected: %v", err)
	}
	if want := `C:\work\scn\sub\out.txt`; got != want {
		t.Errorf("forward-slash resolve = %q, want %q", got, want)
	}
	if _, err := ResolveWorkdirPath("f", root, `..\escape.txt`); err == nil {
		t.Fatal("windows parent traversal accepted")
	}
}

// TestResolveWorkdirPath_ForwardSlashRelative proves a forward-slash relative
// path (how a portable spec is authored) resolves to the host's native separator
// and stays inside the workdir on every OS — so the same spec addresses the same
// file on Windows as on POSIX. On windows-latest CI this is the positive proof
// that `/`-separated spec paths are normalized to `\`.
func TestResolveWorkdirPath_ForwardSlashRelative(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	got, err := ResolveWorkdirPath("f", root, "sub/deep/out.txt")
	if err != nil {
		t.Fatalf("forward-slash relative path rejected: %v", err)
	}
	if want := filepath.Join(root, "sub", "deep", "out.txt"); got != want {
		t.Errorf("resolve = %q, want %q (native separators)", got, want)
	}
	if !WithinRoot(root, got) {
		t.Errorf("resolved path %q is not within root %q", got, root)
	}
}
