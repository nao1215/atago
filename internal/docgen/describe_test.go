package docgen

import (
	"testing"

	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/spectest"
)

// TestDescribeTarget_CoversEveryAssertTarget walks spec.AllAssertTargets and
// proves the doc renderer has a case for each one. The switch's default returns
// the bare target name, which reads as a heading with no sentence — a new target
// would silently publish that instead of describing what the scenario
// guarantees. The Assert values are minimal (the target's field allocated, no
// matcher set), so this covers the dispatch, not the phrasing.
func TestDescribeTarget_CoversEveryAssertTarget(t *testing.T) {
	t.Parallel()
	for _, target := range spec.AllAssertTargets() {
		got := describeTarget(spectest.AssertForTarget(target), target)
		if got == "" {
			t.Errorf("target %q renders as an empty bullet", target)
		}
		if got == string(target) {
			t.Errorf("target %q fell through to the default branch (rendered as the bare name)", target)
		}
	}
}
