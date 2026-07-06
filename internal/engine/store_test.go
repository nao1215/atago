package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/spec"
)

func TestJSONValue(t *testing.T) {
	t.Parallel()
	data := []byte(`{"id":42,"name":"Alice","items":[{"k":"v"}]}`)
	tests := []struct {
		path, want string
		wantErr    bool
	}{
		{"$.id", "42", false},
		{"$.name", "Alice", false},
		{"$.items[0].k", "v", false},
		{"$.missing", "", true},     // selects nothing
		{"$.items[*].k", "", false}, // selects exactly one here
	}
	for _, tt := range tests {
		got, err := jsonValue(data, tt.path)
		if tt.wantErr {
			if err == nil {
				t.Errorf("jsonValue(%q) err = nil, want error", tt.path)
			}
			continue
		}
		if err != nil {
			t.Errorf("jsonValue(%q) err = %v", tt.path, err)
			continue
		}
		if tt.want != "" && got != tt.want {
			t.Errorf("jsonValue(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestRegexValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		data, pattern, want string
		wantErr             bool
	}{
		{"token=abc123", `token=(\w+)`, "abc123", false}, // capture group
		{"id-7", `\d+`, "7", false},                      // whole match (no group)
		{"nothing", `zzz`, "", true},                     // no match
	}
	for _, tt := range tests {
		got, err := regexValue([]byte(tt.data), tt.pattern)
		if tt.wantErr {
			if err == nil {
				t.Errorf("regexValue(%q,%q) err = nil, want error", tt.data, tt.pattern)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("regexValue(%q,%q) = %q,%v, want %q", tt.data, tt.pattern, got, err, tt.want)
		}
	}
}

// TestExtractStream_Trim covers the #158 whole-stream selector: trim captures
// the entire stream without a regex, optionally stripping surrounding
// whitespace.
func TestExtractStream_Trim(t *testing.T) {
	t.Parallel()
	yes, no := true, false
	multiline := "line one\nline two\n"
	tests := []struct {
		name string
		s    *spec.StreamAssert
		data string
		want string
		err  bool
	}{
		{"trim true strips trailing newline", &spec.StreamAssert{Trim: &yes}, "token-abc\n", "token-abc", false},
		{"trim true strips surrounding whitespace", &spec.StreamAssert{Trim: &yes}, "  spaced  ", "spaced", false},
		{"trim false keeps bytes verbatim", &spec.StreamAssert{Trim: &no}, "token-abc\n", "token-abc\n", false},
		{"trim true captures whole multi-line content", &spec.StreamAssert{Trim: &yes}, multiline, "line one\nline two", false},
		{"trim false captures whole multi-line content verbatim", &spec.StreamAssert{Trim: &no}, multiline, multiline, false},
		{"no selector errors", &spec.StreamAssert{}, "x", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := extractStream(tt.s, []byte(tt.data))
			if tt.err {
				if err == nil {
					t.Errorf("extractStream(%+v) err = nil, want error", tt.s)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("extractStream = %q,%v, want %q", got, err, tt.want)
			}
		})
	}
}

// TestExtractValue_FromFileText covers the #158 whole-file selector: text: true
// captures the entire file content verbatim, with no json path.
func TestExtractValue_FromFileText(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := "opaque\nmulti-line\nblob\n"
	if err := os.WriteFile(filepath.Join(dir, "blob.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	yes := true
	sp := &spec.Store{Name: "t", From: &spec.StoreFrom{
		File: &spec.FileAssert{Path: "blob.txt", Text: &yes},
	}}
	got, err := extractValue(sp, &runner.Result{}, dir)
	if err != nil || got != content {
		t.Fatalf("extractValue file text = %q,%v, want %q verbatim", got, err, content)
	}
}

func TestExtractValue_NoCommand(t *testing.T) {
	t.Parallel()
	sp := &spec.Store{Name: "x", From: &spec.StoreFrom{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.id"}}}}}
	if _, err := extractValue(sp, nil, ""); err == nil {
		t.Error("extractValue with nil result should error")
	}
}

func TestEngine_StoreFlow(t *testing.T) {
	t.Parallel()
	jsonEcho := `printf "{\"id\":7}"`
	if runtime.GOOS == "windows" {
		jsonEcho = `echo {"id":7}` // cmd.exe echoes braces and quotes literally
	}
	res := runSpec(t, `
version: "1"
suite:
  name: s
scenarios:
  - name: capture and reuse
    steps:
      - run:
          shell: true
          command: '`+jsonEcho+`'
      - store:
          name: id
          from:
            stdout:
              json:
                path: "$.id"
      - run:
          shell: true
          command: 'echo got=${id}'
      - assert:
          stdout:
            contains: "got=7"
`)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios)
	}
}

func TestExtractValue_FromFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "out.json"), []byte(`{"token":"xyz"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	sp := &spec.Store{Name: "t", From: &spec.StoreFrom{
		File: &spec.FileAssert{Path: "out.json", JSON: spec.JSONChecks{{Path: "$.token"}}},
	}}
	got, err := extractValue(sp, &runner.Result{}, dir)
	if err != nil || got != "xyz" {
		t.Fatalf("extractValue from file = %q,%v, want xyz", got, err)
	}
}

// TestExtractValue_FromFile_TraversalRejected proves a store file source may not
// read outside the scenario workdir, even when the escaping target exists.
func TestExtractValue_FromFile_TraversalRejected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	workdir := filepath.Join(root, "scn")
	if err := os.Mkdir(workdir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "secret.json"), []byte(`{"token":"leak"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	sp := &spec.Store{Name: "t", From: &spec.StoreFrom{
		File: &spec.FileAssert{Path: "../secret.json", JSON: spec.JSONChecks{{Path: "$.token"}}},
	}}
	_, err := extractValue(sp, &runner.Result{}, workdir)
	if err == nil {
		t.Fatal("store file source escaping the workdir was accepted")
	}
	if !strings.Contains(err.Error(), "escapes the scenario workdir") {
		t.Errorf("error %q should explain the containment failure", err)
	}
}
