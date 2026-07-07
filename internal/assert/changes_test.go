package assert

import (
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
			got := checkChanges(tt.assert, changesResult(tt.delta))
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
	)
	if got.OK {
		t.Fatal("an unescaped [ never matches the literal filename, so it should fail")
	}
	wantNote := `note: "[" is a glob metacharacter — write "weird\[1\].txt" to match a literal filename`
	if !strings.Contains(got.Hint, wantNote) {
		t.Errorf("Hint should carry the glob-metacharacter note.\n got: %s\nwant substring: %s", got.Hint, wantNote)
	}
}

// TestCheckChanges_NoDelta proves a missing delta (no preceding run/pty step)
// is reported rather than silently passing.
func TestCheckChanges_NoDelta(t *testing.T) {
	t.Parallel()
	got := checkChanges(&spec.ChangesAssert{Created: list("x")}, &runner.Result{})
	if got.OK {
		t.Error("checkChanges with nil delta should not pass")
	}
	if got = checkChanges(&spec.ChangesAssert{Created: list("x")}, nil); got.OK {
		t.Error("checkChanges with nil result should not pass")
	}
}
