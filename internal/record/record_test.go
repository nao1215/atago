package record

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/nao1215/atago/internal/engine"
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

// TestGenerate_EmitsSchemaHeader proves a recorded spec starts with the
// resolvable yaml-language-server schema header, so `atago record` output gets
// editor completion for the DSL out of the box, and the spec still loads with
// the header present (it is an ignored YAML comment) (#121).
func TestGenerate_EmitsSchemaHeader(t *testing.T) {
	t.Parallel()
	out, err := Generate(Observation{
		Command:  "echo hi",
		ExitCode: 0,
		Stdout:   []byte("hi\n"),
	}, Options{SuiteName: "echo"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	first, _, _ := strings.Cut(string(out), "\n")
	if !strings.HasPrefix(first, "# yaml-language-server: $schema=https://") {
		t.Errorf("first line = %q, want an absolute schema header", first)
	}
	if strings.Contains(first, "./schema/") {
		t.Errorf("schema URL must be absolute, not repo-relative: %q", first)
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

// TestGenerate_ControlBytesRoundTrip is a regression: a recorded value carrying
// a tab or newline must survive a generate → reload round-trip exactly. A raw
// tab spliced into a plain YAML scalar is silently stripped on reparse (so the
// generated assertion could never match the real output), and a newline made
// yaml.Marshal emit a block scalar that broke the document and aborted record.
func TestGenerate_ControlBytesRoundTrip(t *testing.T) {
	t.Parallel()

	// Tab-separated first line (very common CLI output): must round-trip.
	out, err := Generate(Observation{
		Command:  "mytool list",
		ExitCode: 0,
		Stdout:   []byte("col1\tcol2\tcol3\n"),
	}, Options{SuiteName: "demo"})
	if err != nil {
		t.Fatalf("Generate(tab): %v", err)
	}
	s, err := loader.LoadBytes("g.atago.yaml", out)
	if err != nil {
		t.Fatalf("reload(tab): %v\n%s", err, out)
	}
	var contains []string
	for _, st := range s.Scenarios[0].Steps {
		if st.Assert != nil && st.Assert.Stdout != nil && st.Assert.Stdout.Contains != nil {
			contains = st.Assert.Stdout.Contains
		}
	}
	if len(contains) != 1 || contains[0] != "col1\tcol2\tcol3" {
		t.Errorf("tab-separated contains round-trip = %q, want [\"col1\\tcol2\\tcol3\"]", contains)
	}

	// A multi-line command must produce a valid spec, not abort with an
	// "atago bug" internal error.
	if _, err := Generate(Observation{
		Command:  "sh -c 'printf a\nprintf b'",
		Shell:    true,
		ExitCode: 0,
		Stdout:   []byte("ab\n"),
	}, Options{SuiteName: "demo"}); err != nil {
		t.Errorf("Generate(multi-line command) failed: %v", err)
	}

	// Invalid-UTF-8 first line (binary output, or Latin-1 / Shift-JIS text): the
	// generated contains anchor must reload byte-for-byte identical to the raw
	// output, or the anchor can never match the real (raw-byte) stdout on replay
	// and the recorded spec is RED by construction. A raw byte in a plain scalar,
	// or a \xNN double-quoted escape, is lossily transformed on reparse; only a
	// !!binary anchor round-trips the exact bytes.
	rawLine := "caf\xe9 ready"
	out2, err := Generate(Observation{
		Command:  "printf-tool",
		ExitCode: 0,
		Stdout:   []byte(rawLine + "\n"),
	}, Options{SuiteName: "demo"})
	if err != nil {
		t.Fatalf("Generate(invalid-utf8): %v", err)
	}
	s2, err := loader.LoadBytes("g.atago.yaml", out2)
	if err != nil {
		t.Fatalf("reload(invalid-utf8): %v\n%s", err, out2)
	}
	var got []string
	for _, st := range s2.Scenarios[0].Steps {
		if st.Assert != nil && st.Assert.Stdout != nil && st.Assert.Stdout.Contains != nil {
			got = st.Assert.Stdout.Contains
		}
	}
	if len(got) != 1 || got[0] != rawLine {
		t.Errorf("invalid-utf8 contains round-trip = %x, want [%x]", got, rawLine)
	}
}

// TestGenerate_RecordRunRoundTrip is the metamorphic law for record (#30): a
// spec generated from an observed run must itself replay green. The other tests
// check the generated text and that it loads; this replays the generated spec
// through the real engine against the same command and asserts it passes. The
// commands are cross-platform (`echo`/`true`/`false` via the shell, as the other
// engine tests use) so the round-trip is proven on every OS. The escape logic
// this feature depends on — including the ${1} regression where a ${ not
// followed by a valid name must stay literal — is exercised OS-independently by
// store.TestEscapeExpandRoundTrip, so it needs no POSIX-only replay here.
func TestGenerate_RecordRunRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		obs  Observation
	}{
		{"stdout and clean exit", Observation{Command: "echo hello", Shell: true, ExitCode: 0, Stdout: []byte("hello\n")}},
		{"multi-word stdout", Observation{Command: "echo one two three", Shell: true, ExitCode: 0, Stdout: []byte("one two three\n")}},
		{"nonzero exit", Observation{Command: "exit 1", Shell: true, ExitCode: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := Generate(tc.obs, Options{SuiteName: "rt"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			s, err := loader.LoadBytes("rt.atago.yaml", out)
			if err != nil {
				t.Fatalf("load generated: %v\n%s", err, out)
			}
			res := engine.New().Run(context.Background(), s, "rt.atago.yaml")
			if res.Status != engine.StatusPassed {
				t.Errorf("recorded spec did not replay green: status=%s\n--- spec ---\n%s", res.Status, out)
			}
		})
	}
}

// TestGenerate_MasksScratchWorkdirInAnchor is a regression for the record→run
// round-trip (#30): a command that prints an absolute path under the
// record-time scratch dir must not pin that path as a literal contains anchor.
// The replay runs in a *different* isolated workdir, so a literal scratch path
// can never match and the generated spec fails on its first replay. The
// record-time workdir is rewritten to the built-in ${workdir} reference, which
// expands to the replay workdir. This is deterministic and OS-independent
// (`atago record -- pwd` is the canonical trigger).
func TestGenerate_MasksScratchWorkdirInAnchor(t *testing.T) {
	t.Parallel()
	const wd = "/tmp/atago-record-4242"
	out, err := Generate(Observation{
		Command:  "pwd",
		ExitCode: 0,
		Stdout:   []byte("cwd is " + wd + "/logs\n"),
	}, Options{SuiteName: "pwd", Workdir: wd})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	got := string(out)
	if strings.Contains(got, wd) {
		t.Errorf("generated anchor pins the literal scratch workdir %q:\n%s", wd, got)
	}
	if !strings.Contains(got, "contains: cwd is ${workdir}/logs") {
		t.Errorf("anchor did not rewrite the scratch workdir to ${workdir}:\n%s", got)
	}
	// The rewrite must be a live reference: re-escaping must not neutralize it
	// into the $${workdir} literal the expander would leave untouched.
	if strings.Contains(got, "$${workdir}") {
		t.Errorf("workdir reference was escaped to a literal:\n%s", got)
	}
}

// TestGenerate_WorkdirAnchorReplaysGreen proves the metamorphic law end to end
// for the workdir case: a recorded `pwd` replays green because its anchor
// expands to the replay's workdir. POSIX-only (`pwd` is not a cmd.exe builtin);
// the rewrite itself is proven OS-independently by
// TestGenerate_MasksScratchWorkdirInAnchor.
func TestGenerate_WorkdirAnchorReplaysGreen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("pwd is POSIX-only; the rewrite is proven cross-platform elsewhere")
	}
	t.Parallel()
	const wd = "/tmp/atago-record-9999"
	out, err := Generate(Observation{
		Command:  "pwd",
		Shell:    true,
		ExitCode: 0,
		Stdout:   []byte(wd + "\n"),
	}, Options{SuiteName: "pwd", Workdir: wd})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	s, err := loader.LoadBytes("rt.atago.yaml", out)
	if err != nil {
		t.Fatalf("load generated: %v\n%s", err, out)
	}
	res := engine.New().Run(context.Background(), s, "rt.atago.yaml")
	if res.Status != engine.StatusPassed {
		t.Errorf("recorded pwd spec did not replay green: status=%s\n%s", res.Status, out)
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

// TestYAMLScalar_ExoticControlBytesRoundTrip is a metamorphic law for the record
// escaper: for any string, embedding yamlScalar(s) as a flow scalar and parsing
// the resulting YAML back must yield s byte-for-byte. It targets the escape
// branches yaml.Marshal cannot safely inline — an arbitrary C0 control byte
// (ESC), a backslash, and an embedded double quote — which route through
// yamlDoubleQuoted's `\xNN` / `\\` / `\"` cases. A regression here would make
// `atago record` emit a spec whose recorded assertion no longer matches the real
// output.
func TestYAMLScalar_ExoticControlBytesRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []string{
		"esc\x1bhere",                 // ESC → \x1b (the arbitrary-control-byte branch)
		"bell\aring",                  // BEL (0x07)
		"vtab\vand\ff",                // VT (0x0b) + FF (0x0c)
		"back\\slash",                 // literal backslash → \\
		`has "quotes" inside`,         // embedded double quote → \"
		"mixed\ttab\nnewline\rcr end", // the common control trio
		"plain punctuation: # * ? &",  // no control bytes → yaml.Marshal path
		"日本語と\x1bエスケープ",               // multibyte + control byte together
		"caf\xe9 ready",               // invalid UTF-8 (Latin-1 é): a lone 0x80–0xFF byte
		"\xff\xfe\x00",                // pure binary: a byte sequence that is not text at all
		"a\xe9b\x1bc",                 // invalid UTF-8 mixed with a C0 control byte
	}
	for _, s := range cases {
		scalar := yamlScalar(s)
		var doc map[string]string
		if err := yaml.Unmarshal([]byte("v: "+scalar+"\n"), &doc); err != nil {
			t.Errorf("yamlScalar(%q) = %s produced unparseable YAML: %v", s, scalar, err)
			continue
		}
		if doc["v"] != s {
			t.Errorf("round-trip mismatch: yamlScalar(%q) = %s parsed back as %q", s, scalar, doc["v"])
		}
	}
}
