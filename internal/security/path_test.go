package security

import (
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

func TestResolveWorkdirPath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific separator handling")
	}
	root := `C:\work\scn`
	if _, err := ResolveWorkdirPath("f", root, `sub\out.txt`); err != nil {
		t.Fatalf("in-workdir windows path rejected: %v", err)
	}
	if _, err := ResolveWorkdirPath("f", root, `..\escape.txt`); err == nil {
		t.Fatal("windows parent traversal accepted")
	}
}
