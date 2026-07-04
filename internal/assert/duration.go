package assert

import (
	"fmt"
	"strings"
	"time"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

// checkDuration bounds the measured wall-clock time of the preceding step
// (#31). Bounds are duration-validated at load time; a step killed by its own
// timeout still recorded a duration, so the assert evaluates normally against
// it. The CI-variance warning is always appended to the hint — shared runners
// are slow, so bounds should be orders of magnitude, not milliseconds.
func checkDuration(d *spec.DurationAssert, res *runner.Result) *CheckResult {
	if res == nil {
		return &CheckResult{Desc: "assert duration", Hint: "no measurable step has run in this scenario yet"}
	}
	got := res.Duration
	desc := "assert duration " + durationBoundsLabel(d)

	for _, b := range durationBounds(d) {
		if !b.holds(got) {
			return &CheckResult{
				Desc:     desc,
				Expected: "duration " + b.label,
				Actual:   got.Round(time.Microsecond).String(),
				Hint:     fmt.Sprintf("the step took %s, expected %s (CI runners are slow — assert orders of magnitude, not milliseconds)", got.Round(time.Microsecond), b.label),
			}
		}
	}
	return pass(desc)
}

// durationBound is one comparison the measured duration must satisfy.
type durationBound struct {
	label string
	holds func(time.Duration) bool
}

// durationBounds resolves the set bounds into comparisons. The strings parse
// cleanly because the loader already validated them.
func durationBounds(d *spec.DurationAssert) []durationBound {
	var out []durationBound
	if d.LT != "" {
		lim, _ := time.ParseDuration(d.LT)
		out = append(out, durationBound{"< " + d.LT, func(g time.Duration) bool { return g < lim }})
	}
	if d.LTE != "" {
		lim, _ := time.ParseDuration(d.LTE)
		out = append(out, durationBound{"<= " + d.LTE, func(g time.Duration) bool { return g <= lim }})
	}
	if d.GT != "" {
		lim, _ := time.ParseDuration(d.GT)
		out = append(out, durationBound{"> " + d.GT, func(g time.Duration) bool { return g > lim }})
	}
	if d.GTE != "" {
		lim, _ := time.ParseDuration(d.GTE)
		out = append(out, durationBound{">= " + d.GTE, func(g time.Duration) bool { return g >= lim }})
	}
	return out
}

// durationBoundsLabel renders the whole bound set for a check description.
func durationBoundsLabel(d *spec.DurationAssert) string {
	bounds := durationBounds(d)
	labels := make([]string, len(bounds))
	for i, b := range bounds {
		labels[i] = b.label
	}
	return strings.Join(labels, " and ")
}
