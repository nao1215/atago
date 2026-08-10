package assert

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// TestCheck_StreamCount covers the occurrence bounds on a stream matcher
// (#396). The bug class they exist for is duplicate output — an error the
// library logs once and main logs again — which every presence matcher passes.
func TestCheck_StreamCount(t *testing.T) {
	t.Parallel()
	// "warn" appears three times; "once" appears once.
	res := &runner.Result{Stdout: []byte("warn: a\nonce\nwarn: b\nwarn: c\n")}
	tests := []struct {
		name   string
		s      *spec.StreamAssert
		wantOK bool
	}{
		{"exact hit", &spec.StreamAssert{Contains: spec.StringList{"warn"}, Count: intp(3)}, true},
		{"exact miss", &spec.StreamAssert{Contains: spec.StringList{"warn"}, Count: intp(2)}, false},
		{"exactly once", &spec.StreamAssert{Contains: spec.StringList{"once"}, Count: intp(1)}, true},
		{"count zero is absence", &spec.StreamAssert{Contains: spec.StringList{"nope"}, Count: intp(0)}, true},
		{"count zero fails when present", &spec.StreamAssert{Contains: spec.StringList{"once"}, Count: intp(0)}, false},
		{"min hit", &spec.StreamAssert{Contains: spec.StringList{"warn"}, MinCount: intp(3)}, true},
		{"min miss", &spec.StreamAssert{Contains: spec.StringList{"warn"}, MinCount: intp(4)}, false},
		{"max hit", &spec.StreamAssert{Contains: spec.StringList{"warn"}, MaxCount: intp(3)}, true},
		{"max miss", &spec.StreamAssert{Contains: spec.StringList{"warn"}, MaxCount: intp(2)}, false},
		{"range hit", &spec.StreamAssert{Contains: spec.StringList{"warn"}, MinCount: intp(2), MaxCount: intp(4)}, true},
		{"range miss", &spec.StreamAssert{Contains: spec.StringList{"once"}, MinCount: intp(2), MaxCount: intp(4)}, false},
		{"regexp exact", &spec.StreamAssert{Matches: strp(`warn: [abc]`), Count: intp(3)}, true},
		{"regexp miss", &spec.StreamAssert{Matches: strp(`warn: [ab]`), Count: intp(3)}, false},
		{"line selector narrows the count", &spec.StreamAssert{Line: intp(1), Contains: spec.StringList{"warn"}, Count: intp(1)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{Stdout: tt.s}, res, Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_StreamCount_NonOverlapping pins the counting rule both matchers
// share: occurrences do not overlap, so "aa" occurs once in "aaa". A count that
// changed meaning depending on which matcher spelled it would make the number
// unreadable.
func TestCheck_StreamCount_NonOverlapping(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("aaa")}
	for name, s := range map[string]*spec.StreamAssert{
		"literal": {Contains: spec.StringList{"aa"}, Count: intp(1)},
		"regexp":  {Matches: strp("aa"), Count: intp(1)},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := Check(&spec.Assert{Stdout: s}, res, Env{}); !got.OK {
				t.Errorf("OK = false, want true (%s)", got.Hint)
			}
		})
	}
}

// TestCheck_StreamCount_FailureNamesWhere pins the part of the report that
// makes an off-by-one diagnosable: the count that was seen and the lines it was
// seen on. A bare "expected 1, got 2" leaves the author hunting for the extra.
func TestCheck_StreamCount_FailureNamesWhere(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("boom\nfine\nboom\n")}
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{
		Contains: spec.StringList{"boom"}, Count: intp(1),
	}}, res, Env{})
	if got.OK {
		t.Fatal("OK = true, want false")
	}
	for _, want := range []string{"occurs 2 times", "exactly 1 time", "line 1", "line 3"} {
		if !strings.Contains(got.Hint, want) {
			t.Errorf("hint %q should mention %q", got.Hint, want)
		}
	}
}

// TestCheck_StreamCount_CRLF keeps the count consistent with every other stream
// text matcher: line endings are an OS artifact, so a needle authored with LF
// counts the same against cmd.exe's CRLF output.
func TestCheck_StreamCount_CRLF(t *testing.T) {
	t.Parallel()
	res := &runner.Result{Stdout: []byte("a: 1\r\nb: 2\r\na: 3\r\n")}
	got := Check(&spec.Assert{Stdout: &spec.StreamAssert{
		Contains: spec.StringList{"a: "}, Count: intp(2),
	}}, res, Env{})
	if !got.OK {
		t.Errorf("OK = false, want true (%s)", got.Hint)
	}
}

// TestCheck_ScreenCount proves the bounds reach the rendered screen too, since
// ScreenAssert shares the stream matcher surface — "the selected marker is drawn
// once" is a real TUI claim.
func TestCheck_ScreenCount(t *testing.T) {
	t.Parallel()
	res := &runner.Result{IsPTY: true, Screen: []byte("> item a\n  item b\n  item c\n")}
	got := Check(&spec.Assert{Screen: &spec.ScreenAssert{
		StreamAssert: spec.StreamAssert{Contains: spec.StringList{"item "}, Count: intp(3)},
	}}, res, Env{})
	if !got.OK {
		t.Errorf("OK = false, want true (%s)", got.Hint)
	}
}

// TestCheck_FileCount covers the same bounds on a file's content (#396).
func TestCheck_FileCount(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "log.txt"), []byte("hit\nmiss\nhit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		f      *spec.FileAssert
		wantOK bool
	}{
		{"exact hit", &spec.FileAssert{Path: "log.txt", Contains: spec.StringList{"hit"}, Count: intp(2)}, true},
		{"exact miss", &spec.FileAssert{Path: "log.txt", Contains: spec.StringList{"hit"}, Count: intp(1)}, false},
		{"min", &spec.FileAssert{Path: "log.txt", Contains: spec.StringList{"hit"}, MinCount: intp(2)}, true},
		{"max", &spec.FileAssert{Path: "log.txt", Contains: spec.StringList{"hit"}, MaxCount: intp(1)}, false},
		{"missing file reports the read failure", &spec.FileAssert{Path: "nope.txt", Contains: spec.StringList{"hit"}, Count: intp(0)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{File: tt.f}, nil, Env{Workdir: dir})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_FileSize covers the byte-size bounds (#397), including the case
// they were added for: a failed run that must leave an empty file rather than a
// half-written one.
func TestCheck_FileSize(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "empty.csv"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "five.txt"), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "adir"), 0o750); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		f      *spec.FileAssert
		wantOK bool
	}{
		{"empty file", &spec.FileAssert{Path: "empty.csv", Size: int64p(0)}, true},
		{"exact hit", &spec.FileAssert{Path: "five.txt", Size: int64p(5)}, true},
		{"exact miss", &spec.FileAssert{Path: "five.txt", Size: int64p(4)}, false},
		{"non-empty via min", &spec.FileAssert{Path: "five.txt", MinSize: int64p(1)}, true},
		{"empty fails min", &spec.FileAssert{Path: "empty.csv", MinSize: int64p(1)}, false},
		{"max hit", &spec.FileAssert{Path: "five.txt", MaxSize: int64p(5)}, true},
		{"max miss", &spec.FileAssert{Path: "five.txt", MaxSize: int64p(4)}, false},
		{"range", &spec.FileAssert{Path: "five.txt", MinSize: int64p(1), MaxSize: int64p(10)}, true},
		{"missing file", &spec.FileAssert{Path: "nope.txt", Size: int64p(0)}, false},
		{"a directory is not a file", &spec.FileAssert{Path: "adir", MinSize: int64p(0)}, false},
		// Size composes with a content matcher: both must hold.
		{"size and contains both hold", &spec.FileAssert{Path: "five.txt", Contains: spec.StringList{"123"}, Size: int64p(5)}, true},
		{"size holds but contains fails", &spec.FileAssert{Path: "five.txt", Contains: spec.StringList{"nope"}, Size: int64p(5)}, false},
		{"contains holds but size fails", &spec.FileAssert{Path: "five.txt", Contains: spec.StringList{"123"}, Size: int64p(9)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Check(&spec.Assert{File: tt.f}, nil, Env{Workdir: dir})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (%s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheck_FileSize_NearMissNamesTheUsualCause pins the hint that turns a bare
// number into a diagnosis: one byte over an exact size is a trailing newline
// often enough that the report should say so.
func TestCheck_FileSize_NearMissNamesTheUsualCause(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := Check(&spec.Assert{File: &spec.FileAssert{Path: "x.txt", Size: int64p(3)}}, nil, Env{Workdir: dir})
	if got.OK {
		t.Fatal("OK = true, want false")
	}
	if !strings.Contains(got.Hint, "trailing newline") {
		t.Errorf("hint %q should name the trailing newline for a one-byte overshoot", got.Hint)
	}
}

func int64p(n int64) *int64 { return &n }

// TestCheck_FileSize_RefusesToStatThroughASymlink is the #16 rule applied to
// the size bounds: a program under test can plant a symlink at the assertion
// target, and following it would report the size of a host file outside the
// workdir. The read path already refuses; the stat path has to as well.
func TestCheck_FileSize_RefusesToStatThroughASymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating a symlink needs elevation on Windows; the POSIX path is the one at risk")
	}
	t.Parallel()
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("0123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "link.txt")); err != nil {
		t.Fatal(err)
	}
	got := Check(&spec.Assert{File: &spec.FileAssert{Path: "link.txt", Size: int64p(10)}}, nil, Env{Workdir: dir})
	if got.OK {
		t.Fatal("a size bound must not be satisfied through a symlink to a host file")
	}
	if !strings.Contains(got.Hint, "symlink") {
		t.Errorf("hint %q should name the refused symlink", got.Hint)
	}
}
