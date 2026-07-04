package assert

import (
	"strings"
	"testing"
	"time"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// TestCheckDuration_Bounds proves the inclusive/exclusive boundary semantics
// and that a passing multi-bound interval requires ALL bounds (#31).
func TestCheckDuration_Bounds(t *testing.T) {
	t.Parallel()
	res := func(d time.Duration) *runner.Result { return &runner.Result{Duration: d} }
	cases := []struct {
		name string
		d    *spec.DurationAssert
		got  time.Duration
		pass bool
	}{
		{"lt pass", &spec.DurationAssert{LT: "2s"}, time.Second, true},
		{"lt fail equal", &spec.DurationAssert{LT: "2s"}, 2 * time.Second, false},
		{"lte pass equal", &spec.DurationAssert{LTE: "2s"}, 2 * time.Second, true},
		{"gt pass", &spec.DurationAssert{GT: "1s"}, 2 * time.Second, true},
		{"gt fail equal", &spec.DurationAssert{GT: "1s"}, time.Second, false},
		{"gte pass equal", &spec.DurationAssert{GTE: "1s"}, time.Second, true},
		{"interval pass", &spec.DurationAssert{GTE: "100ms", LT: "60s"}, 200 * time.Millisecond, true},
		{"interval fail lower", &spec.DurationAssert{GTE: "500ms", LT: "60s"}, 200 * time.Millisecond, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cr := checkDuration(tc.d, res(tc.got))
			if cr.OK != tc.pass {
				t.Errorf("OK = %v, want %v (got=%s bounds=%+v): %+v", cr.OK, tc.pass, tc.got, tc.d, cr)
			}
		})
	}
}

// TestCheckDuration_FailureText proves the measured duration and the
// CI-variance hint appear in failure output (#31).
func TestCheckDuration_FailureText(t *testing.T) {
	t.Parallel()
	cr := checkDuration(&spec.DurationAssert{LT: "1ms"}, &runner.Result{Duration: 3410 * time.Millisecond})
	if cr.OK {
		t.Fatal("3.41s < 1ms should fail")
	}
	if !strings.Contains(cr.Actual, "3.41") {
		t.Errorf("Actual = %q, want the measured duration", cr.Actual)
	}
	if !strings.Contains(cr.Expected, "< 1ms") {
		t.Errorf("Expected = %q, want the bound", cr.Expected)
	}
	if !strings.Contains(cr.Hint, "orders of magnitude") {
		t.Errorf("Hint = %q, want the CI-variance warning", cr.Hint)
	}
}

// TestCheckDuration_NoStep proves a duration assert with no measured step is a
// clean failure, not a panic.
func TestCheckDuration_NoStep(t *testing.T) {
	t.Parallel()
	cr := checkDuration(&spec.DurationAssert{LT: "1s"}, nil)
	if cr.OK {
		t.Error("no step should fail")
	}
}
