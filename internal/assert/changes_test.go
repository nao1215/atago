package assert

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/fsdelta"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

func list(items ...string) *spec.StringList {
	l := spec.StringList(items)
	return &l
}

func changesResult(d fsdelta.Delta) *runner.Result {
	return &runner.Result{Changes: &d}
}

func TestCheckChanges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		assert *spec.ChangesAssert
		delta  fsdelta.Delta
		wantOK bool
	}{
		{
			name:   "exact created match passes",
			assert: &spec.ChangesAssert{Created: list("site/index.html")},
			delta:  fsdelta.Delta{Created: []string{"site/index.html"}},
			wantOK: true,
		},
		{
			name:   "glob covers created files",
			assert: &spec.ChangesAssert{Created: list("site/index.html", "site/assets/*.css")},
			delta:  fsdelta.Delta{Created: []string{"site/index.html", "site/assets/app.css"}},
			wantOK: true,
		},
		{
			name:   "glob does not cross a slash boundary",
			assert: &spec.ChangesAssert{Created: list("site/*.css")},
			delta:  fsdelta.Delta{Created: []string{"site/assets/app.css"}},
			wantOK: false, // * must not match "assets/app"
		},
		{
			name:   "unexpected created file fails",
			assert: &spec.ChangesAssert{Created: list("site/index.html")},
			delta:  fsdelta.Delta{Created: []string{"site/index.html", "site/extra.html"}},
			wantOK: false,
		},
		{
			name:   "entry matching nothing fails",
			assert: &spec.ChangesAssert{Created: list("site/index.html", "site/missing.html")},
			delta:  fsdelta.Delta{Created: []string{"site/index.html"}},
			wantOK: false,
		},
		{
			name:   "empty modified asserts modified nothing (passes)",
			assert: &spec.ChangesAssert{Created: list("out.txt"), Modified: list(), Deleted: list()},
			delta:  fsdelta.Delta{Created: []string{"out.txt"}},
			wantOK: true,
		},
		{
			name:   "empty modified fails when something was modified",
			assert: &spec.ChangesAssert{Modified: list()},
			delta:  fsdelta.Delta{Modified: []string{"config.txt"}},
			wantOK: false,
		},
		{
			name:   "omitted field is unconstrained",
			assert: &spec.ChangesAssert{Created: list("out.txt")},
			delta:  fsdelta.Delta{Created: []string{"out.txt"}, Modified: []string{"other.txt"}, Deleted: []string{"gone.txt"}},
			wantOK: true, // modified/deleted omitted → unconstrained
		},
		{
			name:   "deleted category matched exhaustively",
			assert: &spec.ChangesAssert{Deleted: list("tmp/*")},
			delta:  fsdelta.Delta{Deleted: []string{"tmp/a", "tmp/b"}},
			wantOK: true,
		},
		{
			name:   "doublestar matches at any depth",
			assert: &spec.ChangesAssert{Created: list("site/**")},
			delta:  fsdelta.Delta{Created: []string{"site/index.html", "site/a/b/c.txt"}},
			wantOK: true,
		},
		{
			name:   "doublestar matches at depth 1",
			assert: &spec.ChangesAssert{Created: list("site/**")},
			delta:  fsdelta.Delta{Created: []string{"site/index.html"}},
			wantOK: true,
		},
		{
			name:   "doublestar composes with a suffix pattern",
			assert: &spec.ChangesAssert{Created: list("dist/**/*.css")},
			delta:  fsdelta.Delta{Created: []string{"dist/app.css", "dist/a/b/theme.css"}},
			wantOK: true,
		},
		{
			name:   "doublestar composition rejects the wrong suffix",
			assert: &spec.ChangesAssert{Created: list("dist/**/*.css")},
			delta:  fsdelta.Delta{Created: []string{"dist/app.js"}},
			wantOK: false, // .js is not covered, and the entry itself matches nothing
		},
		{
			name:   "doublestar does not spill onto a sibling prefix",
			assert: &spec.ChangesAssert{Created: list("site/**")},
			delta:  fsdelta.Delta{Created: []string{"sitex/y.txt"}},
			wantOK: false, // site/** must not match sitex/...
		},
		{
			// Pinned semantics: `site/**` matches the bare `site` path itself
			// (doublestar's native behavior), not only paths strictly under it.
			name:   "doublestar matches the bare prefix path itself",
			assert: &spec.ChangesAssert{Created: list("site/**")},
			delta:  fsdelta.Delta{Created: []string{"site"}},
			wantOK: true,
		},
		{
			name:   "backslash escapes a literal metacharacter",
			assert: &spec.ChangesAssert{Created: list(`a\[1\].txt`)},
			delta:  fsdelta.Delta{Created: []string{"a[1].txt"}},
			wantOK: true,
		},
		{
			// Documented: an UNescaped `[` is still a character class, so
			// `a[1].txt` matches `a1.txt` (and not the literal `a[1].txt`).
			name:   "unescaped bracket stays a character class",
			assert: &spec.ChangesAssert{Created: list("a[1].txt")},
			delta:  fsdelta.Delta{Created: []string{"a1.txt"}},
			wantOK: true,
		},
		{
			// A leading "./" on an entry is stripped before matching: observed
			// paths are workdir-relative without "./", so "./out.txt" must match
			// "out.txt".
			name:   "leading ./ on entry is normalized before matching",
			assert: &spec.ChangesAssert{Created: list("./out.txt")},
			delta:  fsdelta.Delta{Created: []string{"out.txt"}},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := checkChanges(tt.assert, changesResult(tt.delta), Env{})
			if got.OK != tt.wantOK {
				t.Errorf("checkChanges OK = %v, want %v (hint: %s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheckChanges_GlobMetaHint proves that an entry containing an unescaped
// glob metacharacter that matches nothing gets a clarifying note appended to the
// Hint, rather than a byte-identical, self-contradictory Expected/Actual block.
func TestCheckChanges_GlobMetaHint(t *testing.T) {
	t.Parallel()
	got := checkChanges(
		&spec.ChangesAssert{Created: list("weird[1].txt")},
		changesResult(fsdelta.Delta{Created: []string{"weird[1].txt"}}),
		Env{},
	)
	if got.OK {
		t.Fatal("an unescaped [ never matches the literal filename, so it should fail")
	}
	wantNote := `note: "[" is a glob metacharacter — write "weird\[1\].txt" to match a literal filename`
	if !strings.Contains(got.Hint, wantNote) {
		t.Errorf("Hint should carry the glob-metacharacter note.\n got: %s\nwant substring: %s", got.Hint, wantNote)
	}
}

// TestCheckChanges_GlobBraceHint proves the doublestar `{a,b}` brace
// alternation is treated as a metacharacter too: a literal file named
// "a{1}.txt" never matches its own name, so the failure must carry the note
// suggesting the escaped spelling rather than a baffling identical block.
func TestCheckChanges_GlobBraceHint(t *testing.T) {
	t.Parallel()
	got := checkChanges(
		&spec.ChangesAssert{Created: list("a{1}.txt")},
		changesResult(fsdelta.Delta{Created: []string{"a{1}.txt"}}),
		Env{},
	)
	if got.OK {
		t.Fatal("an unescaped { never matches the literal filename, so it should fail")
	}
	wantNote := `note: "{" is a glob metacharacter — write "a\{1\}.txt" to match a literal filename`
	if !strings.Contains(got.Hint, wantNote) {
		t.Errorf("Hint should carry the brace-metacharacter note.\n got: %s\nwant substring: %s", got.Hint, wantNote)
	}
}

// TestCheckChanges_NoDelta proves a missing delta (no preceding run/pty step)
// is reported rather than silently passing.
func TestCheckChanges_NoDelta(t *testing.T) {
	t.Parallel()
	got := checkChanges(&spec.ChangesAssert{Created: list("x")}, &runner.Result{}, Env{})
	if got.OK {
		t.Error("checkChanges with nil delta should not pass")
	}
	if got = checkChanges(&spec.ChangesAssert{Created: list("x")}, nil, Env{}); got.OK {
		t.Error("checkChanges with nil result should not pass")
	}
}

// TestCheckChanges_UntrackedKindHint covers the other way an unmatched entry
// reads as a lie: the path is right there in the workdir, but as something a
// delta does not track. fsdelta scans regular files and symlinks, so a directory
// a generator created is invisible to created/modified/deleted, and "matched no
// file the step created" reads as "the tool never made it".
func TestCheckChanges_UntrackedKindHint(t *testing.T) {
	t.Parallel()
	wd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wd, "out"), 0o750); err != nil {
		t.Fatal(err)
	}
	got := checkChanges(
		&spec.ChangesAssert{Created: list("out")},
		changesResult(fsdelta.Delta{}),
		Env{Workdir: wd},
	)
	if got.OK {
		t.Fatal("a directory is not tracked by the delta, so the entry should fail")
	}
	for _, want := range []string{
		`"out" exists as a directory`,
		"tracks only regular files and symlinks",
		"assert it with dir:",
	} {
		if !strings.Contains(got.Hint, want) {
			t.Errorf("Hint should say what the path is.\n got: %s\nwant substring: %s", got.Hint, want)
		}
	}

	// An entry naming nothing at all keeps the plain message: there is no kind to
	// report, and inventing one would be worse than saying less.
	plain := checkChanges(
		&spec.ChangesAssert{Created: list("never-created.txt")},
		changesResult(fsdelta.Delta{}),
		Env{Workdir: wd},
	)
	if plain.OK {
		t.Fatal("an absent path should fail")
	}
	if strings.Contains(plain.Hint, "exists as") {
		t.Errorf("Hint for an absent path should not claim it exists: %s", plain.Hint)
	}
}

// TestCheckChanges_Ignore covers the escape hatch for a path the program writes
// only sometimes (#327): a state file in the sandboxed HOME, a cache, a socket
// from a background service it starts on some runs. Without it the exhaustive
// contract is unusable for such a tool — naming the path fails on the runs that
// do not write it, and omitting the category asserts nothing.
func TestCheckChanges_Ignore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		assert *spec.ChangesAssert
		delta  fsdelta.Delta
		wantOK bool
	}{
		{
			name:   "an ignored creation does not break an exhaustive list",
			assert: &spec.ChangesAssert{Ignore: spec.StringList{".atago-home/**"}, Created: list("out.txt"), Modified: list(), Deleted: list()},
			delta:  fsdelta.Delta{Created: []string{".atago-home/.local/state/tool/.sock", "out.txt"}},
			wantOK: true,
		},
		{
			name:   "the same spec passes when the incidental path is absent",
			assert: &spec.ChangesAssert{Ignore: spec.StringList{".atago-home/**"}, Created: list("out.txt")},
			delta:  fsdelta.Delta{Created: []string{"out.txt"}},
			wantOK: true,
		},
		{
			name:   "ignore applies to modified and deleted too",
			assert: &spec.ChangesAssert{Ignore: spec.StringList{"cache/**"}, Created: list(), Modified: list(), Deleted: list()},
			delta:  fsdelta.Delta{Created: []string{"cache/a"}, Modified: []string{"cache/b"}, Deleted: []string{"cache/c"}},
			wantOK: true,
		},
		{
			name:   "a path outside the ignore globs is still unexpected",
			assert: &spec.ChangesAssert{Ignore: spec.StringList{"cache/**"}, Created: list()},
			delta:  fsdelta.Delta{Created: []string{"cache/a", "surprise.txt"}},
			wantOK: false,
		},
		{
			name:   "an ignored path cannot satisfy an entry that names it",
			assert: &spec.ChangesAssert{Ignore: spec.StringList{"out.txt"}, Created: list("out.txt")},
			delta:  fsdelta.Delta{Created: []string{"out.txt"}},
			wantOK: false,
		},
		{
			name:   "an ignore glob that matches nothing is fine",
			assert: &spec.ChangesAssert{Ignore: spec.StringList{"never/**"}, Created: list("out.txt")},
			delta:  fsdelta.Delta{Created: []string{"out.txt"}},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := checkChanges(tt.assert, changesResult(tt.delta), Env{})
			if got.OK != tt.wantOK {
				t.Errorf("OK = %v, want %v (hint: %s)", got.OK, tt.wantOK, got.Hint)
			}
		})
	}
}

// TestCheckChanges_IgnoredPathsStayOutOfTheEvidence pins what the failure
// block presents. The Actual line must describe the delta the assertion
// actually judged — the one with the ignore globs applied — and name what
// ignore dropped. It used to render the raw delta, so an ignored debug.log sat
// inside "created [...]" as evidence while the exhaustive created: list that
// omits it raised no problem, and the reader could not reconcile the claim
// with the evidence printed under it.
func TestCheckChanges_IgnoredPathsStayOutOfTheEvidence(t *testing.T) {
	t.Parallel()
	got := checkChanges(&spec.ChangesAssert{
		Ignore:  spec.StringList{"*.log"},
		Created: list("out.txt", "missing.txt"),
	}, changesResult(fsdelta.Delta{Created: []string{"debug.log", "out.txt"}}), Env{})
	if got.OK {
		t.Fatal("assertion should have failed: missing.txt matched nothing")
	}
	if strings.Contains(got.Actual, "created [debug.log") || strings.Contains(got.Actual, ", debug.log]") {
		t.Errorf("Actual = %q, an ignored path must not appear as observed evidence", got.Actual)
	}
	if !strings.Contains(got.Actual, "created [out.txt]") {
		t.Errorf("Actual = %q, want the judged delta", got.Actual)
	}
	if !strings.Contains(got.Actual, "ignored [debug.log]") {
		t.Errorf("Actual = %q, want it to name what ignore dropped", got.Actual)
	}
}

// TestCheckChanges_UntrackedKindHintStaysInsideTheWorkdir pins the confinement of
// the hint's own filesystem probe. `changes:` entries are author-written patterns,
// and untrackedKindNote used to stat them through a bare filepath.Join, so an
// entry naming `../secret` reported the kind of a file outside the scenario
// workdir — a hint that answers "does this exist, and what is it" about a path the
// spec has no business reaching. Every other path-taking assertion resolves
// through security.ResolveWorkdirPath; this one has to as well.
func TestCheckChanges_UntrackedKindHintStaysInsideTheWorkdir(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	wd := filepath.Join(base, "workdir")
	if err := os.MkdirAll(wd, 0o750); err != nil {
		t.Fatal(err)
	}
	// A directory OUTSIDE the workdir: the kind the hint would happily report.
	if err := os.MkdirAll(filepath.Join(base, "secret"), 0o750); err != nil {
		t.Fatal(err)
	}
	got := checkChanges(
		&spec.ChangesAssert{Created: list("../secret")},
		changesResult(fsdelta.Delta{}),
		Env{Workdir: wd},
	)
	if got.OK {
		t.Fatal("an escaping entry cannot be satisfied by an empty delta")
	}
	if strings.Contains(got.Hint, "exists as a directory") {
		t.Errorf("the hint probed outside the scenario workdir: %s", got.Hint)
	}
}
