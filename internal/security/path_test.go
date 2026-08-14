package security

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestResolve_AncestorSymlinkMayNotEscapeTheRoot pins the containment rule
// against a symlink ABOVE the leaf. The check was purely lexical, so a link the
// program under test planted at a directory component was invisible to it:
// `<root>/escape/secret.txt` compares as in-root while naming a host file, and
// every path-taking feature inherited the hole from this one resolver.
//
// A link that stays inside the root is ordinary and must keep resolving, and a
// path that is merely absent must not become an error — `exists: false` asks
// exactly that question.
func TestResolve_AncestorSymlinkMayNotEscapeTheRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "real"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(root, "alias")); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "through a link out of the root", in: "escape/secret.txt", wantErr: true},
		{name: "deeper through a link out of the root", in: "escape/sub/secret.txt", wantErr: true},
		{name: "the link itself is left to the read and write helpers", in: "escape"},
		{name: "through a link that stays inside the root", in: "alias/out.txt"},
		{name: "a path that simply does not exist", in: "absent/out.txt"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ResolveWorkdirPath("assert.file.path", root, tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ResolveWorkdirPath(%q) = nil error; the path names a location outside the root", tt.in)
				}
				if !strings.Contains(err.Error(), "symlink") || !strings.Contains(err.Error(), "scenario workdir") {
					t.Errorf("error %q should say the path resolves through a symlink out of the workdir", err)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolveWorkdirPath(%q) error = %v; want it accepted", tt.in, err)
			}
		})
	}
}

// TestResolve_DanglingAncestorJudgedByItsDeclaredTarget keeps the rule from
// depending on whether a link's target happens to exist yet: a link out of the
// root that currently dangles still declares where it points, and the program
// under test can create that target at any moment. A dangling link that stays
// inside the root is ordinary (a `latest ->` for a release not built yet).
func TestResolve_DanglingAncestorJudgedByItsDeclaredTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(t.TempDir(), "absent"), filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("releases/v3", filepath.Join(root, "latest")); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveWorkdirPath("fixture path", root, "escape/out.txt"); err == nil {
		t.Error("a dangling link out of the root must be refused, not written through")
	}
	if _, err := ResolveWorkdirPath("fixture path", root, "latest/out.txt"); err != nil {
		t.Errorf("a dangling link inside the root must still resolve: %v", err)
	}
}

// TestResolve_DanglingChainJudgedByItsWholeChain closes the one-hop reading of
// the rule above. Only the END of a chain says where a path goes: `hop ->
// escape` with `escape -> /outside/absent` looks in-root for one hop, and a
// check that stops there accepts a path the program under test completes into a
// host directory at will. A hop that leaves the root is refused wherever in the
// chain it sits, because the kernel resolves through that location.
func TestResolve_DanglingChainJudgedByItsWholeChain(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(t.TempDir(), "absent"), filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("escape", filepath.Join(root, "hop")); err != nil {
		t.Fatal(err)
	}
	// A chain that stays inside the root the whole way is ordinary and must keep
	// resolving, and a cycle must be answered rather than followed forever.
	if err := os.Symlink("latest", filepath.Join(root, "current")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("releases/v3", filepath.Join(root, "latest")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("b", filepath.Join(root, "a")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("a", filepath.Join(root, "b")); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolveWorkdirPath("fixture path", root, "hop/out.txt"); err == nil {
		t.Error("a chain whose far end leaves the root must be refused")
	}
	if _, err := ResolveWorkdirPath("fixture path", root, "current/out.txt"); err != nil {
		t.Errorf("a dangling chain that stays inside the root must still resolve: %v", err)
	}
	if _, err := ResolveWorkdirPath("fixture path", root, "a/out.txt"); err != nil {
		t.Errorf("a link cycle inside the root must be answered, not refused: %v", err)
	}
}

// TestResolve_RootBehindASymlink is the macOS spelling case. CI resolves the
// scenario workdir under /var and /tmp, which sit behind /private, so an
// ancestor EvalSymlinks resolved comes back in a spelling the root itself is
// never written with — comparing only the two literal forms refuses a path that
// never left the root.
func TestResolve_RootBehindASymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliably available on Windows CI")
	}
	t.Parallel()
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.MkdirAll(filepath.Join(real, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink("real", alias); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveWorkdirPath("assert.file.path", alias, "sub/out.txt"); err != nil {
		t.Errorf("a path inside a workdir reached through a symlink was refused: %v", err)
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

// TestReadFileNoFollow_RefusesANamedPipe is the hang regression: opening a named
// pipe for reading blocks until another process writes to it, and nothing bounds
// the assertion phase, so a program under test that left a pipe where a file was
// expected hung the run forever instead of failing it. The refusal has to name
// what the path is, since "could not read" alone sends the author looking for a
// permission problem.
func TestReadFileNoFollow_RefusesANamedPipe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no named pipes in the filesystem namespace on Windows")
	}
	bin, lookErr := exec.LookPath("mkfifo")
	if lookErr != nil {
		t.Skip("mkfifo not available")
	}
	t.Parallel()
	pipe := filepath.Join(t.TempDir(), "pipe")
	if out, err := exec.CommandContext(t.Context(), bin, pipe).CombinedOutput(); err != nil {
		t.Skipf("mkfifo: %v (%s)", err, out)
	}

	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	go func() {
		data, err := ReadFileNoFollow(pipe)
		done <- result{data, err}
	}()
	select {
	case got := <-done:
		if got.err == nil {
			t.Fatalf("ReadFileNoFollow(pipe) = %q, nil; want an error", got.data)
		}
		if !strings.Contains(got.err.Error(), "named pipe") {
			t.Errorf("error %q should name the entry as a named pipe", got.err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("ReadFileNoFollow blocked on a named pipe; it must refuse a non-regular file instead of opening it")
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

// TestWriteFileNoFollow_ConcurrentIdenticalContent is the regression for #250:
// several parallel scenarios that share one golden file call WriteFileNoFollow
// on the same path with byte-identical content (e.g. matrix rows producing the
// same output under --update-snapshots). The old non-atomic
// Lstat→Remove→OpenFile(O_EXCL) sequence raced — one goroutine's Remove hit the
// file another had already removed (ENOENT), or its O_EXCL open hit a file
// another had just created — so an update failed nondeterministically even
// though every writer produced the same bytes. An atomic write must let every
// concurrent identical write succeed and leave the expected content behind.
func TestWriteFileNoFollow_ConcurrentIdenticalContent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dest := filepath.Join(root, "shared.golden")
	content := []byte("same output from every row\n")

	const writers = 16
	var wg sync.WaitGroup
	errs := make([]error, writers)
	start := make(chan struct{})
	for i := range writers {
		wg.Go(func() {
			<-start // release all goroutines at once to maximize contention
			errs[i] = WriteFileNoFollow(dest, content, 0o600)
		})
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("writer %d: WriteFileNoFollow errored on identical concurrent write: %v", i, err)
		}
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("final content = %q, want %q", got, content)
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

// TestWriteWorkdirFile covers the confined-write helper the four
// workdir-scoped features share (#349): stdout_to/stderr_to, http.body_to,
// fixtures, and (via WriteConfinedFile) snapshots and CDP screenshots. Each of
// them used to open-code resolve → MkdirAll → WriteFileNoFollow, and the file
// mode had already drifted between them.
func TestWriteWorkdirFile(t *testing.T) {
	t.Parallel()

	t.Run("creates parent directories", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		dest, err := WriteWorkdirFile("run.stdout_to", root, filepath.Join("logs", "deep", "out.txt"), []byte("produced"))
		if err != nil {
			t.Fatalf("WriteWorkdirFile: %v", err)
		}
		if want := filepath.Join(root, "logs", "deep", "out.txt"); dest != want {
			t.Errorf("dest = %q, want %q", dest, want)
		}
		got, err := os.ReadFile(dest)
		if err != nil || string(got) != "produced" {
			t.Fatalf("read back = %q, %v", got, err)
		}
	})

	t.Run("rejects an escape and names the field", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		for _, rel := range []string{
			filepath.Join("sub", "..", "..", "escape.txt"),
			filepath.Join(t.TempDir(), "absolute-outside.txt"),
		} {
			_, err := WriteWorkdirFile("http.body_to", root, rel, []byte("x"))
			if err == nil {
				t.Fatalf("WriteWorkdirFile(%q) succeeded; want a containment error", rel)
			}
			if !strings.Contains(err.Error(), "http.body_to") {
				t.Errorf("error does not name the field: %v", err)
			}
			if !strings.Contains(err.Error(), "escapes the scenario workdir") {
				t.Errorf("error does not say what went wrong: %v", err)
			}
		}
	})

	t.Run("refuses to write through a planted leaf symlink", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("symlink creation is not reliably available on Windows CI")
		}
		t.Parallel()
		root := t.TempDir()
		victim := filepath.Join(t.TempDir(), "victim.txt")
		if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(victim, filepath.Join(root, "redirect.txt")); err != nil {
			t.Fatal(err)
		}
		if _, err := WriteWorkdirFile("run.stdout_to", root, "redirect.txt", []byte("pwned")); err == nil {
			t.Fatal("wrote through the planted symlink; want error")
		}
		if got, _ := os.ReadFile(victim); string(got) != "original" {
			t.Errorf("host file was modified through the symlink: %q", got)
		}
	})

	// One mode for every confined write. stdout_to used to be the odd one out at
	// 0o644 (#349); pin the agreed mode so it cannot drift back unnoticed.
	t.Run("uses the confined file mode", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows does not honor POSIX permission bits")
		}
		t.Parallel()
		root := t.TempDir()
		dest, err := WriteWorkdirFile("run.stdout_to", root, "out.txt", []byte("x"))
		if err != nil {
			t.Fatalf("WriteWorkdirFile: %v", err)
		}
		fi, err := os.Stat(dest)
		if err != nil {
			t.Fatal(err)
		}
		if got := fi.Mode().Perm(); got != confinedFileMode {
			t.Errorf("mode = %04o, want %04o", got, confinedFileMode)
		}
	})
}

// TestWriteConfinedFile is the resolved-path form used by snapshot writes, which
// resolve against the spec directory rather than the workdir, and by CDP
// screenshots, which resolve when the task is built and write after the run.
func TestWriteConfinedFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dest := filepath.Join(root, "golden", "out.snap")
	if err := WriteConfinedFile(dest, []byte("v1")); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteConfinedFile(dest, []byte("v2")); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "v2" {
		t.Fatalf("read back = %q, %v; want %q", got, err, "v2")
	}
}
