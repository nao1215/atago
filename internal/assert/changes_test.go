package assert

import (
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
