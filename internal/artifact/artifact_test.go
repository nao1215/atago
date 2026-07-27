package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"stdout":               "stdout",
		"Hello World":          "hello-world",
		"scenario: prints ok!": "scenario-prints-ok",
		"  spaced  ":           "spaced",
		"":                     "artifact",
		"____":                 "artifact",
		"MiXeD_Case-123":       "mixed-case-123",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSuiteTokenDeterministicAndCollisionFree(t *testing.T) {
	t.Parallel()
	// Same spec path → identical token (deterministic).
	a1 := SuiteToken("test/e2e/atago/run.atago.yaml")
	a2 := SuiteToken("test/e2e/atago/run.atago.yaml")
	if a1 != a2 {
		t.Fatalf("SuiteToken not deterministic: %q != %q", a1, a2)
	}
	// Two different spec paths that slug to the same base must not collide,
	// because the short hash suffix differs.
	b := SuiteToken("other/dir/run.atago.yaml")
	if a1 == b {
		t.Fatalf("distinct spec paths collided: both %q", a1)
	}
	if !strings.HasPrefix(a1, "run-") {
		t.Errorf("SuiteToken = %q, want readable prefix", a1)
	}
}

func TestFailurePathStableAndUnique(t *testing.T) {
	t.Parallel()
	run := Scenario{SpecPath: "test/e2e/atago/run.atago.yaml", Name: "prints hello"}
	p := run.FailurePath(2, "stdout", "actual", "txt")
	if !strings.HasSuffix(p, "step-02-stdout.actual.txt") {
		t.Errorf("FailurePath filename = %q", p)
	}
	if strings.Contains(p, "\\") {
		t.Errorf("FailurePath must use forward slashes: %q", p)
	}
	// Distinct scenarios in the same suite never collide.
	second := Scenario{SpecPath: "test/e2e/atago/run.atago.yaml", Name: "prints hello", Index: 1}
	q := second.FailurePath(2, "stdout", "actual", "txt")
	if p == q {
		t.Errorf("scenario index did not disambiguate path: %q", p)
	}
}

// TestMockLogPath_DistinctFromServiceLog proves a mock server's request log
// never collides with a service log even when both share a declared name.
func TestMockLogPath_DistinctFromServiceLog(t *testing.T) {
	t.Parallel()
	run := Scenario{SpecPath: "test/e2e/atago/mock.atago.yaml", Name: "client posts"}
	m := run.MockLogPath("api")
	if !strings.HasSuffix(m, "mock-api.log") {
		t.Errorf("MockLogPath = %q", m)
	}
	s := run.ServiceLogPath("api")
	if m == s {
		t.Errorf("mock and service logs for the same name collided: %q", m)
	}
}

func TestServiceLogPathStableAndUnique(t *testing.T) {
	t.Parallel()
	run := Scenario{SpecPath: "test/e2e/atago/services.atago.yaml", Name: "peer talks"}
	p := run.ServiceLogPath("api server")
	if !strings.HasSuffix(p, "service-api-server.log") {
		t.Errorf("ServiceLogPath = %q", p)
	}
	// Distinct services in the same scenario land in distinct files.
	q := run.ServiceLogPath("db server")
	if p == q {
		t.Errorf("distinct services collided: %q", p)
	}
	// Same scenario dir as failure sidecars.
	fp := run.FailurePath(1, "stdout", "actual", "txt")
	if dirOf(p) != dirOf(fp) {
		t.Errorf("service log %q not in the scenario dir of %q", p, fp)
	}
}

// TestScenarioDir_AttemptSeparatesRepeatedExecutions pins the attempt segment: a
// plain run keeps the path it always had, and every further --repeat iteration or
// --retry-failed attempt writes under its own directory so one attempt's payload
// cannot overwrite another's while a report still points at it.
func TestScenarioDir_AttemptSeparatesRepeatedExecutions(t *testing.T) {
	t.Parallel()
	base := Scenario{SpecPath: "test/e2e/atago/run.atago.yaml", Name: "prints hello"}
	// Attempt 0 (never set) and 1 (the first execution) are the same location.
	if got, want := base.Dir(), (Scenario{SpecPath: base.SpecPath, Name: base.Name, Attempt: 1}).Dir(); got != want {
		t.Errorf("attempt 0 dir = %q, attempt 1 dir = %q, want them identical", got, want)
	}
	if strings.Contains(base.Dir(), "attempt-") {
		t.Errorf("first attempt dir = %q, want no attempt segment", base.Dir())
	}

	seen := map[string]bool{base.Dir(): true}
	for attempt := 2; attempt <= 4; attempt++ {
		s := base
		s.Attempt = attempt
		dir := s.Dir()
		if !strings.HasSuffix(dir, fmt.Sprintf("attempt-%d", attempt)) {
			t.Errorf("attempt %d dir = %q, want it to name the attempt", attempt, dir)
		}
		if !strings.HasPrefix(dir, base.Dir()+"/") {
			t.Errorf("attempt %d dir = %q, want it under the scenario dir %q", attempt, dir, base.Dir())
		}
		if seen[dir] {
			t.Errorf("attempt %d reused directory %q", attempt, dir)
		}
		seen[dir] = true
	}

	// Every artifact of one attempt shares that attempt's directory, so a
	// scenario's failure payloads, service logs, and mock logs stay together.
	third := base
	third.Attempt = 3
	for _, p := range []string{
		third.FailurePath(1, "stdout", "actual", "txt"),
		third.ServiceLogPath("api"),
		third.MockLogPath("api"),
	} {
		if dirOf(p) != third.Dir() {
			t.Errorf("%q is not in the attempt dir %q", p, third.Dir())
		}
	}
}

func dirOf(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[:i]
	}
	return p
}

func TestDirWriteCreatesFileAndReturnsRelPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	d := NewDir(root)
	rel, err := d.Write("suite-x/scenario-0/step-01-stdout.actual.txt", []byte("hello"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if rel != "suite-x/scenario-0/step-01-stdout.actual.txt" {
		t.Errorf("returned rel path = %q", rel)
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q", got)
	}
}
