package engine

import (
	"testing"

	"github.com/nao1215/atago/internal/assert"
	"github.com/nao1215/atago/internal/runner"
	"github.com/nao1215/atago/internal/security"
	"github.com/nao1215/atago/internal/spec"
	"github.com/nao1215/atago/internal/store"
)

func ptr(s string) *string { return &s }

// TestExpandAssert_StreamAndFileMatchers covers that ${name}/${workdir}
// substitution reaches stream and file matcher values, not just file.path.
func TestExpandAssert_StreamAndFileMatchers(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("workdir", "/tmp/wd")
	st.Set("token", "abc")

	in := &spec.Assert{
		Stdout: &spec.StreamAssert{
			Equals:      ptr("${workdir}/out.txt"),
			Contains:    nil,
			NotContains: spec.StringList{"${token}-no"},
		},
		Stderr: &spec.StreamAssert{Matches: ptr("^${token}$")},
		File: &spec.FileAssert{
			Path:     "${workdir}/note.txt",
			Contains: spec.StringList{"${token}"},
		},
	}

	out := expandAssert(st, in)

	if got := *out.Stdout.Equals; got != "/tmp/wd/out.txt" {
		t.Errorf("stdout.equals = %q, want /tmp/wd/out.txt", got)
	}
	if got := out.Stdout.NotContains[0]; got != "abc-no" {
		t.Errorf("stdout.not_contains = %q, want abc-no", got)
	}
	if got := *out.Stderr.Matches; got != "^abc$" {
		t.Errorf("stderr.matches = %q, want ^abc$", got)
	}
	if got := out.File.Path; got != "/tmp/wd/note.txt" {
		t.Errorf("file.path = %q, want /tmp/wd/note.txt", got)
	}
	if got := out.File.Contains[0]; got != "abc" {
		t.Errorf("file.contains = %q, want abc", got)
	}

	// The original assert must be untouched (expandAssert returns a copy).
	if *in.Stdout.Equals != "${workdir}/out.txt" {
		t.Errorf("input mutated: %q", *in.Stdout.Equals)
	}
}

// TestExpandAssert_JSONAndYAMLMatchers covers #6: ${name} substitution must
// reach json/yaml matcher payloads (path, equals, matches) on stream and file
// targets, not just the plain text matchers.
func TestExpandAssert_JSONAndYAMLMatchers(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("got_id", "42")
	st.Set("re", "ab+c")
	st.Set("key", "id")

	in := &spec.Assert{
		Stdout: &spec.StreamAssert{JSON: &spec.JSONAssert{
			Path:   "$.${key}",
			Equals: "${got_id}",
		}},
		Stderr: &spec.StreamAssert{YAML: &spec.JSONAssert{
			Path:    "$.name",
			Matches: ptr("^${re}$"),
		}},
		File: &spec.FileAssert{Path: "out.json", JSON: &spec.JSONAssert{
			Path:   "$.token",
			Equals: []any{"${got_id}", 1},
		}},
	}
	out := expandAssert(st, in)

	if got := out.Stdout.JSON.Path; got != "$.id" {
		t.Errorf("stdout.json.path = %q, want $.id", got)
	}
	if got := out.Stdout.JSON.Equals; got != "42" {
		t.Errorf("stdout.json.equals = %v, want 42", got)
	}
	if got := *out.Stderr.YAML.Matches; got != "^ab+c$" {
		t.Errorf("stderr.yaml.matches = %q, want ^ab+c$", got)
	}
	arr, ok := out.File.JSON.Equals.([]any)
	if !ok || arr[0] != "42" || arr[1] != 1 {
		t.Errorf("file.json.equals = %v, want [42 1]", out.File.JSON.Equals)
	}
	// Original must be untouched.
	if in.Stdout.JSON.Equals != "${got_id}" {
		t.Errorf("input mutated: %v", in.Stdout.JSON.Equals)
	}
}

func TestExpandAssert_Image(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("workdir", "/tmp/wd")
	st.Set("name", "thumb")

	in := &spec.Assert{Image: &spec.ImageAssert{
		Path:      "${workdir}/${name}.png",
		SimilarTo: "${workdir}/base.png",
		Format:    "png",
	}}
	out := expandAssert(st, in)
	if got := out.Image.Path; got != "/tmp/wd/thumb.png" {
		t.Errorf("image.path = %q, want /tmp/wd/thumb.png", got)
	}
	if got := out.Image.SimilarTo; got != "/tmp/wd/base.png" {
		t.Errorf("image.similar_to = %q, want /tmp/wd/base.png", got)
	}
	if in.Image.Path != "${workdir}/${name}.png" {
		t.Errorf("input image mutated: %q", in.Image.Path)
	}
}

func TestExpandAssert_Header(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("v", "ok")
	in := &spec.Assert{Header: &spec.HeaderMatch{Name: "X", Contains: ptr("${v}"), Equals: ptr("${v}")}}
	out := expandAssert(st, in)
	if *out.Header.Contains != "ok" || *out.Header.Equals != "ok" {
		t.Errorf("header not expanded: %+v", out.Header)
	}
}

func TestExpandStore_File(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("workdir", "/wd")
	s := &spec.Store{Name: "x", From: &spec.StoreFrom{File: &spec.FileAssert{Path: "${workdir}/f.json"}}}
	out := expandStore(st, s)
	if got := out.From.File.Path; got != "/wd/f.json" {
		t.Errorf("store file path = %q, want /wd/f.json", got)
	}
	// A store without a file source is returned unchanged.
	noFile := &spec.Store{Name: "y", From: &spec.StoreFrom{Header: "X"}}
	if expandStore(st, noFile) != noFile {
		t.Error("store without file should be returned as-is")
	}
}

func TestMergeScenarioEnv(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("workdir", "/wd")
	r := &spec.Run{Env: map[string]string{"A": "step", "B": "stepb"}}
	merged := mergeScenarioEnv(map[string]string{"A": "scenario", "C": "${workdir}/c"}, r, st)
	if merged.Env["A"] != "step" {
		t.Errorf("step env should win: A=%q", merged.Env["A"])
	}
	if merged.Env["B"] != "stepb" {
		t.Errorf("step-only key lost: B=%q", merged.Env["B"])
	}
	if merged.Env["C"] != "/wd/c" {
		t.Errorf("scenario env not expanded: C=%q", merged.Env["C"])
	}
	// Empty scenario env returns the run untouched.
	if mergeScenarioEnv(nil, r, st) != r {
		t.Error("nil scenario env should return run as-is")
	}
}

func TestExpandAny(t *testing.T) {
	t.Parallel()
	st := store.New()
	st.Set("tok", "secret")
	in := map[string]any{
		"auth":  "Bearer ${tok}",
		"n":     42,
		"items": []any{"${tok}", 1},
		"meta":  map[any]any{"k": "${tok}"},
	}
	out, ok := spec.WalkJSONValueStrings(in, st.Expand).(map[string]any)
	if !ok {
		t.Fatalf("WalkJSONValueStrings did not return a map: %T", out)
	}
	if out["auth"] != "Bearer secret" {
		t.Errorf("auth = %v", out["auth"])
	}
	if out["n"] != 42 {
		t.Errorf("non-string mutated: %v", out["n"])
	}
	arr, ok := out["items"].([]any)
	if !ok || arr[0] != "secret" || arr[1] != 1 {
		t.Errorf("array expand = %v", out["items"])
	}
	m, ok := out["meta"].(map[any]any)
	if !ok || m["k"] != "secret" {
		t.Errorf("nested map expand = %v", out["meta"])
	}
}

func TestMaskResultAndCheck(t *testing.T) {
	t.Parallel()
	m := security.NewMasker([]string{"s3cret"})

	r := &runner.Result{
		Command:     "login --token s3cret",
		Stdout:      []byte("got s3cret"),
		Stderr:      []byte("err s3cret"),
		Body:        []byte("body s3cret"),
		RowsJSON:    []byte("rows s3cret"),
		MessageJSON: []byte("msg s3cret"),
		CDPValue:    []byte("cdp s3cret"),
	}
	masked := maskResult(m, r)
	for _, got := range []string{
		masked.Command, string(masked.Stdout), string(masked.Stderr),
		string(masked.Body), string(masked.RowsJSON), string(masked.MessageJSON), string(masked.CDPValue),
	} {
		if got == "" || containsSecret(got) {
			t.Errorf("masked field still leaks secret: %q", got)
		}
	}
	// Original result untouched.
	if string(r.Stdout) != "got s3cret" {
		t.Error("original result mutated")
	}

	cr := &assert.CheckResult{Desc: "d s3cret", Expected: "e s3cret", Actual: "a s3cret", Hint: "h s3cret"}
	maskCheck(m, cr)
	for _, got := range []string{cr.Desc, cr.Expected, cr.Actual, cr.Hint} {
		if containsSecret(got) {
			t.Errorf("masked check still leaks secret: %q", got)
		}
	}

	// An empty masker returns the same result pointer and leaves checks alone.
	empty := security.NewMasker(nil)
	if maskResult(empty, r) != r {
		t.Error("empty masker should return the same result pointer")
	}
}

func containsSecret(s string) bool {
	return len(s) >= 6 && (s == "s3cret" || indexOf(s, "s3cret") >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
