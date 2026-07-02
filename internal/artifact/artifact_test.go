package artifact

import (
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
	p := FailurePath("test/e2e/atago/run.atago.yaml", "prints hello", 0, 2, "stdout", "actual", "txt")
	if !strings.HasSuffix(p, "step-02-stdout.actual.txt") {
		t.Errorf("FailurePath filename = %q", p)
	}
	if strings.Contains(p, "\\") {
		t.Errorf("FailurePath must use forward slashes: %q", p)
	}
	// Distinct scenarios in the same suite never collide.
	q := FailurePath("test/e2e/atago/run.atago.yaml", "prints hello", 1, 2, "stdout", "actual", "txt")
	if p == q {
		t.Errorf("scenario index did not disambiguate path: %q", p)
	}
}

func TestServiceLogPathStableAndUnique(t *testing.T) {
	t.Parallel()
	p := ServiceLogPath("test/e2e/atago/services.atago.yaml", "peer talks", 0, "api server")
	if !strings.HasSuffix(p, "service-api-server.log") {
		t.Errorf("ServiceLogPath = %q", p)
	}
	// Distinct services in the same scenario land in distinct files.
	q := ServiceLogPath("test/e2e/atago/services.atago.yaml", "peer talks", 0, "db server")
	if p == q {
		t.Errorf("distinct services collided: %q", p)
	}
	// Same scenario dir as failure sidecars.
	fp := FailurePath("test/e2e/atago/services.atago.yaml", "peer talks", 0, 1, "stdout", "actual", "txt")
	if dirOf(p) != dirOf(fp) {
		t.Errorf("service log %q not in the scenario dir of %q", p, fp)
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
