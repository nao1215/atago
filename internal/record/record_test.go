package record

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

// TestGenerate_Skeleton pins the conservative matcher policy (#30): exact
// exit code, first non-empty stdout line as contains, stderr empty only when
// it was, created files as exists asserts.
func TestGenerate_Skeleton(t *testing.T) {
	t.Parallel()
	out, err := Generate(Observation{
		Command:      "mytool convert input.txt",
		ExitCode:     0,
		Stdout:       []byte("\nconverted 3 records\ndetails follow\n"),
		Stderr:       nil,
		CreatedFiles: []string{"output.json"},
	}, Options{SuiteName: "mytool"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"suite:\n  name: mytool",
		"# TODO: describe the behavior",
		"command: mytool convert input.txt",
		"exit_code: 0",
		"contains: converted 3 records",
		"empty: true",
		"path: output.json",
		"exists: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("generated spec missing %q:\n%s", want, got)
		}
	}
}

// TestGenerate_EscapesVariableReferences proves the round-trip law for
// observed text containing ${...}: the engine expands ${name} in run.command
// and assert values at replay (and errors on an unresolved reference in a
// no-shell command), so record must emit the documented $${...} literal escape
// for the command, the stdout anchor, and created-file paths — otherwise the
// generated spec diverges from (or outright fails to replay) the recorded run.
func TestGenerate_EscapesVariableReferences(t *testing.T) {
	t.Parallel()
	out, err := Generate(Observation{
		Command:      "printf %s 'literal ${HOME} here'",
		ExitCode:     0,
		Stdout:       []byte("saw ${workdir} in output\n"),
		CreatedFiles: []string{"weird/${name}.txt"},
	}, Options{SuiteName: "esc"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		"command: printf %s 'literal $${HOME} here'",
		"contains: saw $${workdir} in output",
		"path: weird/$${name}.txt",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("generated spec missing escaped form %q:\n%s", want, got)
		}
	}
	// The scenario name is a human label the engine never expands, so it may
	// keep the raw text; the leak check applies to the expanded fields only.
	for _, line := range strings.Split(got, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "command:") || strings.HasPrefix(trimmed, "contains:") || strings.HasPrefix(trimmed, "path:") {
			if strings.Contains(strings.ReplaceAll(trimmed, "$${", ""), "${") {
				t.Errorf("live ${...} reference leaked into an expanded field: %s", trimmed)
			}
		}
	}
}

// TestGenerate_EdgeShapes proves generation stays valid across observed
// shapes: empty output, non-zero exit, noisy stderr, shell mode, file cap,
// and hostile strings that must not break YAML structure.
func TestGenerate_EdgeShapes(t *testing.T) {
	t.Parallel()
	files := make([]string, 15)
	for i := range files {
		files[i] = fmt.Sprintf("out/f%02d.txt", i)
	}
	cases := []struct {
		name string
		obs  Observation
		want []string
	}{
		{
			name: "empty output nonzero exit",
			obs:  Observation{Command: "false", ExitCode: 1},
			want: []string{"exit_code: 1", "empty: true"},
		},
		{
			name: "noisy stderr drops the empty assert",
			obs:  Observation{Command: "tool", ExitCode: 0, Stderr: []byte("warn\n")},
			want: []string{"exit_code: 0"},
		},
		{
			name: "shell mode",
			obs:  Observation{Command: "echo a | grep a", Shell: true, ExitCode: 0, Stdout: []byte("a\n")},
			want: []string{"shell: true", "command: echo a | grep a"},
		},
		{
			name: "file cap notes the rest",
			obs:  Observation{Command: "gen", ExitCode: 0, CreatedFiles: files},
			want: []string{"path: out/f09.txt", "and 5 more created files"},
		},
		{
			name: "hostile strings stay one scalar",
			obs:  Observation{Command: `tool --msg "a: b # c"`, ExitCode: 0, Stdout: []byte("line: with { yaml } chars\n")},
			want: []string{"exit_code: 0"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := Generate(tc.obs, Options{SuiteName: "s"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			if _, lerr := loader.LoadBytes("g.atago.yaml", out); lerr != nil {
				t.Fatalf("generated spec does not load: %v\n%s", lerr, out)
			}
			for _, w := range tc.want {
				if !strings.Contains(string(out), w) {
					t.Errorf("missing %q:\n%s", w, out)
				}
			}
			if strings.Contains(string(out), "out/f10.txt") {
				t.Errorf("file cap leaked an 11th assert:\n%s", out)
			}
		})
	}
}

// TestGenerate_Snapshot proves --snapshot switches stdout to the snapshot
// matcher referencing the given golden path.
func TestGenerate_Snapshot(t *testing.T) {
	t.Parallel()
	out, err := Generate(
		Observation{Command: "tool", ExitCode: 0, Stdout: []byte("big output\n")},
		Options{SuiteName: "s", Snapshot: true, SnapshotPath: "snapshots/tool.stdout.txt"},
	)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(string(out), "snapshot: snapshots/tool.stdout.txt") {
		t.Errorf("snapshot matcher missing:\n%s", out)
	}
	if strings.Contains(string(out), "contains:") {
		t.Errorf("snapshot mode must replace the contains matcher:\n%s", out)
	}
}
