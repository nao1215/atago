package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Exercise the original POSIX-shell reproduction through --ci, including the
// error exit code. Only elapsed time and the temporary spec path are normalized.
func TestRunCmd_FailedAssertionBeforeError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("reproduction uses POSIX shell quoting")
	}
	const src = `version: "1"
suite:
  name: a failed assertion before a store error
scenarios:
  - name: the assertion fails, then the store errors
    steps:
      - run: {command: "echo '{\"jobs\": []}'", shell: true}
      - assert:
          stdout:
            json:
              - {path: "$.jobs", length: 1}
      - store: {name: job, from: {stdout: {json: {path: "$.jobs[0]"}}}}
`
	path := writeSpec(t, t.TempDir(), "repro.atago.yaml", src)
	var out, errb bytes.Buffer
	if code := Main([]string{"run", "--ci", path}, &out, &errb); code != ExitExec {
		t.Fatalf("exit = %d, stderr = %s, stdout = %s", code, &errb, &out)
	}
	got := strings.ReplaceAll(out.String(), filepath.ToSlash(path), "repro.atago.yaml")
	got = regexp.MustCompile(`\([0-9.]+(?:ns|µs|ms|s)\)`).ReplaceAllString(got, "(TIME)")
	golden := filepath.Join("testdata", "report-failure-before-error.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("console differs from golden:\n%s", got)
	}
}
