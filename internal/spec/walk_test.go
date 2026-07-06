package spec

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestVarRefs(t *testing.T) {
	t.Parallel()
	got := VarRefs("${a}/x/${b_2}-${a}")
	sort.Strings(got)
	if want := []string{"a", "a", "b_2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("VarRefs = %v, want %v", got, want)
	}
	if VarRefs("no vars here") != nil {
		t.Error("VarRefs on plain text should be nil")
	}
	// Issue #37: an escaped $${name} is literal text, not a live reference, so
	// it must not be collected (the linter would otherwise flag it as undefined).
	if got := VarRefs("$${literal}"); got != nil {
		t.Errorf("VarRefs on escaped ref = %v, want nil", got)
	}
	if got := VarRefs("$${skip} but ${real}"); !reflect.DeepEqual(got, []string{"real"}) {
		t.Errorf("VarRefs mixed escape = %v, want [real]", got)
	}
	// Issue #24: namespaced built-ins use dotted names (${<mock>.url},
	// ${<mock>.port}). VarRefs must stay in lockstep with store.varRef and
	// collect them, or manifest/explain silently drop every dotted reference.
	if got := VarRefs("base_url: ${api.url} port ${api.port}"); !reflect.DeepEqual(got, []string{"api.url", "api.port"}) {
		t.Errorf("VarRefs dotted = %v, want [api.url api.port]", got)
	}
	if got := VarRefs("${env:HOME}"); !reflect.DeepEqual(got, []string{"env:HOME"}) {
		t.Errorf("VarRefs env = %v, want [env:HOME]", got)
	}
}

func sp(s string) *string { return &s }

// TestWalkAssertStrings_CollectAndExpand verifies the walker both records (with
// an identity visit) and substitutes (with a mutating visit) across every
// interpolatable field, and that it returns a copy without mutating the input.
func TestWalkAssertStrings_CollectAndExpand(t *testing.T) {
	t.Parallel()
	emptyList := StringList{}
	a := &Assert{
		Stdout:  &StreamAssert{Contains: StringList{"${a}"}},
		Rows:    &StreamAssert{JSON: JSONChecks{{Path: "$.${b}", Equals: "${c}"}}},
		Message: &StreamAssert{Equals: sp("${d}")},
		Value:   &StreamAssert{YAML: JSONChecks{{Path: "$.x", Matches: sp("${e}")}}},
		File:    &FileAssert{Path: "${f}", Contains: StringList{"${g}"}},
		Header:  &HeaderMatch{Name: "X", Equals: sp("${h}"), Matches: sp("${r}")},
		Image:   &ImageAssert{Path: "${i}", SimilarTo: "${j}"},
		Screen:  &StreamAssert{Contains: StringList{"${k}"}},
		Dir:     &DirAssert{Path: "${l}", Contains: []string{"${m}"}, NotContains: []string{"${n}"}, Glob: "${o}", Ignore: []string{"${p}"}},
		PDF:     &PDFAssert{Path: "${q}", Metadata: map[string]string{"title": "${s}"}, Text: &StreamAssert{Contains: StringList{"${t}"}}},
		Mock:    &MockAssert{Name: "api", Path: "${u}", Header: &HeaderMatch{Name: "Y", Contains: sp("${v}")}, Body: &StreamAssert{Contains: StringList{"${w}"}}},
		Changes: &ChangesAssert{Created: &StringList{"${x}"}, Modified: &emptyList},
	}

	// Collect: identity visit records every reference.
	seen := map[string]bool{}
	WalkAssertStrings(a, func(s string) string {
		for _, n := range VarRefs(s) {
			seen[n] = true
		}
		return s
	})
	for _, want := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x"} {
		if !seen[want] {
			t.Errorf("collect missed ${%s}; got %v", want, seen)
		}
	}

	// Expand: mutating visit substitutes, and the input is not mutated.
	out := WalkAssertStrings(a, func(s string) string { return strings.ReplaceAll(s, "${a}", "A") })
	if out.Stdout.Contains[0] != "A" {
		t.Errorf("expand stdout.contains = %q, want A", out.Stdout.Contains[0])
	}
	if a.Stdout.Contains[0] != "${a}" {
		t.Error("WalkAssertStrings mutated its input")
	}

	// The changes lists keep the nil-vs-empty distinction the exhaustive-set
	// semantics depend on: a visited empty list stays non-nil-and-empty, and an
	// omitted (nil) list stays nil.
	if out.Changes.Modified == nil || len(*out.Changes.Modified) != 0 {
		t.Errorf("changes.modified = %v, want a non-nil empty list", out.Changes.Modified)
	}
	if out.Changes.Deleted != nil {
		t.Errorf("changes.deleted = %v, want nil (omitted stays unconstrained)", out.Changes.Deleted)
	}
}

func TestWalkJSONValueStrings(t *testing.T) {
	t.Parallel()
	in := map[string]any{"s": "${x}", "n": 1, "arr": []any{"${y}", 2}}
	out, ok := WalkJSONValueStrings(in, func(s string) string { return strings.ToUpper(s) }).(map[string]any)
	if !ok {
		t.Fatalf("result is not a map")
	}
	if out["s"] != "${X}" {
		t.Errorf("string leaf = %v", out["s"])
	}
	if out["n"] != 1 {
		t.Errorf("non-string mutated: %v", out["n"])
	}
	arr, ok := out["arr"].([]any)
	if !ok || arr[0] != "${Y}" || arr[1] != 2 {
		t.Errorf("array = %v", out["arr"])
	}
	// Input not mutated.
	if in["s"] != "${x}" {
		t.Error("WalkJSONValueStrings mutated its input")
	}
}

// collectStep is a small helper: it runs CollectStepVars over one step and
// returns the referenced variable names, sorted, so a test can assert the exact
// set a step kind contributes.
func collectStep(step *Step) []string {
	set := map[string]bool{}
	CollectStepVars(set, step)
	return SortedKeys(set)
}

func hasVar(vars []string, name string) bool {
	for _, v := range vars {
		if v == name {
			return true
		}
	}
	return false
}

// TestCollectStepVars_UnscannedFields is a regression test for a variable
// under-reporting bug: CollectStepVars is documented as "the single source of
// truth for which fields of each step kind carry variables", yet several fields
// the engine actually ${name}-expands were never scanned. Each ${..._ref} below
// is referenced ONLY from a field the old code skipped, so a miss regresses the
// fix. The reference expansion for each lives in internal/engine/expand.go.
func TestCollectStepVars_UnscannedFields(t *testing.T) {
	t.Parallel()

	// fixture.from is expanded by expandFixture (from: st.Expand(f.From)).
	if got := collectStep(&Step{Fixture: &Fixture{File: "out", From: "${fixfrom_ref}"}}); !hasVar(got, "fixfrom_ref") {
		t.Errorf("fixture.from not collected; got %v", got)
	}

	// http header values and the JSON request body are expanded by expandHTTP.
	httpStep := &Step{HTTP: &HTTP{
		Method: "POST",
		Path:   "/x",
		Header: map[string]string{"Authorization": "Bearer ${httphdr_ref}"},
		JSON:   map[string]any{"id": "${httpjson_ref}", "nested": []any{"${httpjson_arr_ref}"}},
	}}
	got := collectStep(httpStep)
	for _, want := range []string{"httphdr_ref", "httpjson_ref", "httpjson_arr_ref"} {
		if !hasVar(got, want) {
			t.Errorf("http field %q not collected; got %v", want, got)
		}
	}

	// grpc header values and the JSON request message are expanded by expandGRPC.
	grpcStep := &Step{GRPC: &GRPC{
		Runner: "g",
		Method: "pkg.S/${grpcmethod_ref}",
		Header: map[string]string{"x-token": "${grpchdr_ref}"},
		JSON:   map[string]any{"q": "${grpcjson_ref}"},
	}}
	got = collectStep(grpcStep)
	for _, want := range []string{"grpcmethod_ref", "grpchdr_ref", "grpcjson_ref"} {
		if !hasVar(got, want) {
			t.Errorf("grpc field %q not collected; got %v", want, got)
		}
	}

	// An assert step's matcher arguments are expanded by expandAssert; the whole
	// StepAssert kind was previously uncounted.
	assertStep := &Step{Assert: &Assert{
		Stdout: &StreamAssert{Equals: strp("${assert_stdout_ref}")},
		File:   &FileAssert{Path: "${assert_path_ref}", Contains: StringList{"${assert_contains_ref}"}},
	}}
	got = collectStep(assertStep)
	for _, want := range []string{"assert_stdout_ref", "assert_path_ref", "assert_contains_ref"} {
		if !hasVar(got, want) {
			t.Errorf("assert field %q not collected; got %v", want, got)
		}
	}

	// A store step's file-source path is expanded by expandStore.
	storeStep := &Step{Store: &Store{Name: "v", From: &StoreFrom{File: &FileAssert{Path: "${store_path_ref}"}}}}
	if got := collectStep(storeStep); !hasVar(got, "store_path_ref") {
		t.Errorf("store from.file.path not collected; got %v", got)
	}
}

// TestCollectServiceVars_ReadyProbes is a regression test: the engine expands a
// service's readiness file/port/log probes (expandService), so a ${name} in any
// of them is a real variable reference the summaries must report.
func TestCollectServiceVars_ReadyProbes(t *testing.T) {
	t.Parallel()
	set := map[string]bool{}
	CollectServiceVars(set, &Service{
		Name:    "peer",
		Command: "./peer",
		Ready: &Ready{
			File: "${ready_file_ref}",
			Port: "${ready_port_ref}",
			Log:  "listening on ${ready_log_ref}",
		},
	})
	got := SortedKeys(set)
	for _, want := range []string{"ready_file_ref", "ready_port_ref", "ready_log_ref"} {
		if !hasVar(got, want) {
			t.Errorf("service ready probe %q not collected; got %v", want, got)
		}
	}
}

// TestCollectStepVars_KnownFields keeps the previously-fixed fields covered so a
// refactor cannot silently drop them.
func TestCollectStepVars_KnownFields(t *testing.T) {
	t.Parallel()
	run := &Step{Run: &Run{Command: "echo ${cmd}", Cwd: "${cwd}", Env: map[string]string{"T": "${env}"}, Stdin: Stdin{Inline: "${in}"}}}
	got := collectStep(run)
	for _, want := range []string{"cmd", "cwd", "env", "in"} {
		if !hasVar(got, want) {
			t.Errorf("run field %q not collected; got %v", want, got)
		}
	}
	pty := &Step{PTY: &PTY{Command: "${pcmd}", Cwd: "${pcwd}", Env: map[string]string{"K": "${penv}"}, Session: []PTYAction{{Expect: "${pexp}"}, {Send: SendText("${psend}")}}}}
	got = collectStep(pty)
	for _, want := range []string{"pcmd", "pcwd", "penv", "pexp", "psend"} {
		if !hasVar(got, want) {
			t.Errorf("pty field %q not collected; got %v", want, got)
		}
	}
}

// TestCollectStepVars_RemainingKinds covers the step kinds not exercised above
// (service, query, cdp, signal) so CollectStepVars is walked end to end, and
// drives every field collectCDPActionVars can carry.
func TestCollectStepVars_RemainingKinds(t *testing.T) {
	t.Parallel()

	if got := collectStep(&Step{Service: &Service{Name: "s", Command: "echo ${scmd}", Cwd: "${scwd}"}}); !hasVar(got, "scmd") || !hasVar(got, "scwd") {
		t.Errorf("service step vars = %v", got)
	}
	if got := collectStep(&Step{Query: &Query{Runner: "d", SQL: "SELECT ${col}"}}); !hasVar(got, "col") {
		t.Errorf("query step vars = %v", got)
	}
	if got := collectStep(&Step{Signal: &Signal{Service: "${sigsvc}", Signal: "TERM"}}); !hasVar(got, "sigsvc") {
		t.Errorf("signal step vars = %v", got)
	}

	// Every CDP action sub-field that carries a variable must be collected.
	cdp := &Step{CDP: &CDP{Runner: "web", Actions: []CDPAction{
		{Navigate: "${nav}"},
		{WaitVisible: "${wv}"},
		{WaitHidden: "${wh}"},
		{Click: "${clk}"},
		{Check: "${chk}"},
		{Uncheck: "${unchk}"},
		{Text: "${txt}"},
		{Eval: "${evl}"},
		{SendKeys: &CDPSendKeys{Selector: "${sk_sel}", Value: "${sk_val}"}},
		{Press: &CDPPress{Selector: "${pr_sel}", Key: "${pr_key}"}},
		{Select: &CDPSelect{Selector: "${se_sel}", Value: "${se_val}"}},
		{Screenshot: &CDPScreenshot{Path: "${ss_path}", Selector: "${ss_sel}"}},
		{Attribute: &CDPAttribute{Selector: "${at_sel}", Name: "${at_name}"}},
		{Upload: &CDPUpload{Selector: "${up_sel}", File: "${up_file}"}},
		{Download: &CDPDownload{Click: "${dl_click}", Dir: "${dl_dir}"}},
	}}}
	got := collectStep(cdp)
	for _, want := range []string{
		"nav", "wv", "wh", "clk", "chk", "unchk", "txt", "evl",
		"sk_sel", "sk_val", "pr_sel", "pr_key", "se_sel", "se_val",
		"ss_path", "ss_sel", "at_sel", "at_name", "up_sel", "up_file",
		"dl_click", "dl_dir",
	} {
		if !hasVar(got, want) {
			t.Errorf("cdp action var %q not collected; got %v", want, got)
		}
	}
}

// TestCollectStepVars_NoActionStep is a boundary case: a step with no action key
// (or more than one) has Kind StepNone and contributes no variables, never
// panicking.
func TestCollectStepVars_NoActionStep(t *testing.T) {
	t.Parallel()
	if got := collectStep(&Step{}); len(got) != 0 {
		t.Errorf("empty step collected %v, want none", got)
	}
}

// TestCollectJSONVars_NilAndScalar guards the helper's non-map/array inputs: a
// nil body or a bare scalar contributes nothing and never panics.
func TestCollectJSONVars_NilAndScalar(t *testing.T) {
	t.Parallel()
	set := map[string]bool{}
	collectJSONVars(set, nil)
	collectJSONVars(set, 42)
	collectJSONVars(set, "${top_ref}") // a bare string leaf is still scanned
	if len(set) != 1 || !set["top_ref"] {
		t.Errorf("collectJSONVars scalar handling = %v, want {top_ref}", SortedKeys(set))
	}
}

func TestDescribeDuration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		d    DurationAssert
		want string
	}{
		{"empty", DurationAssert{}, ""},
		{"lt", DurationAssert{LT: "2s"}, "in under 2s"},
		{"lte", DurationAssert{LTE: "2s"}, "in at most 2s"},
		{"gt", DurationAssert{GT: "1s"}, "in over 1s"},
		{"gte", DurationAssert{GTE: "1s"}, "in at least 1s"},
		{"interval gt+lt", DurationAssert{GT: "1s", LT: "2s"}, "in under 2s and in over 1s"},
		{"interval gte+lte", DurationAssert{GTE: "1s", LTE: "2s"}, "in at most 2s and in at least 1s"},
		{"all four", DurationAssert{LT: "4s", LTE: "3s", GT: "1s", GTE: "2s"}, "in under 4s and in at most 3s and in over 1s and in at least 2s"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.d.DescribeDuration(); got != tc.want {
				t.Errorf("DescribeDuration(%+v) = %q, want %q", tc.d, got, tc.want)
			}
		})
	}
}

func TestPTYSend_BytesAndLabel_Edges(t *testing.T) {
	t.Parallel()
	// A zero-value send (neither Text nor Key) resolves to no bytes and the
	// fallback label, never a panic.
	zero := &PTYSend{}
	if got := zero.Bytes(); got != nil {
		t.Errorf("zero send Bytes = %q, want nil", got)
	}
	if got := zero.Label(); got != "send" {
		t.Errorf("zero send Label = %q, want \"send\"", got)
	}
	// A named key resolves to its xterm sequence and a "press <key>" label.
	if got := (&PTYSend{Key: "ctrl-c"}).Bytes(); string(got) != "\x03" {
		t.Errorf("ctrl-c bytes = %q", got)
	}
	if got := (&PTYSend{Key: "tab"}).Label(); got != "press tab" {
		t.Errorf("tab label = %q", got)
	}
	// Non-empty text is quoted in the label.
	if got := SendText("hi").Label(); got != `type "hi"` {
		t.Errorf("text label = %q", got)
	}
}

func TestPTYSend_UnmarshalYAML_Errors(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"unknown key":    "{foo: bar}",
		"non-string key": "{key: 5}",
		"empty key name": "{key: \"  \"}",
		"sequence":       "[1, 2]",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var p PTYSend
			if err := yaml.Unmarshal([]byte(in), &p); err == nil {
				t.Errorf("expected error for %s (%q), got none: %+v", name, in, p)
			}
		})
	}
	// A valid {key: Enter} normalizes case and whitespace.
	var p PTYSend
	if err := yaml.Unmarshal([]byte("{key: \"  ENTER \"}"), &p); err != nil {
		t.Fatal(err)
	}
	if p.Key != "enter" {
		t.Errorf("key = %q, want normalized \"enter\"", p.Key)
	}
}

func TestStdin_UnmarshalYAML_Errors(t *testing.T) {
	t.Parallel()
	t.Run("unknown key", func(t *testing.T) {
		t.Parallel()
		var s Stdin
		if err := yaml.Unmarshal([]byte("{bogus: x}"), &s); err == nil {
			t.Error("expected error for unknown stdin key")
		}
	})
	t.Run("non-string value", func(t *testing.T) {
		t.Parallel()
		var s Stdin
		if err := yaml.Unmarshal([]byte("{file: 5}"), &s); err == nil {
			t.Error("expected error for non-string stdin.file")
		}
	})
	t.Run("empty mapping records the mapping form", func(t *testing.T) {
		t.Parallel()
		// An empty mapping ({}) decodes cleanly as the mapping form with no
		// source set; the loader is what rejects it. IsMapping lets the loader
		// tell it apart from "no stdin" (IsZero).
		var s Stdin
		if err := yaml.Unmarshal([]byte("{}"), &s); err != nil {
			t.Fatal(err)
		}
		if s.File != "" || s.Base64 != "" || s.Inline != "" {
			t.Errorf("empty mapping populated a source: %+v", s)
		}
		if !s.IsMapping() {
			t.Error("empty mapping should record the mapping form")
		}
		if s.IsZero() {
			t.Error("a {} mapping is authored stdin, so IsZero should be false")
		}
	})
	t.Run("sequence is neither string nor mapping", func(t *testing.T) {
		t.Parallel()
		// A YAML sequence decodes as neither the scalar nor the {file/base64}
		// mapping form, so it hits the shape error rather than silently accepting.
		var s Stdin
		if err := yaml.Unmarshal([]byte("[1, 2]"), &s); err == nil {
			t.Error("expected error decoding a sequence into stdin")
		}
	})
	t.Run("base64 mapping", func(t *testing.T) {
		t.Parallel()
		var s Stdin
		if err := yaml.Unmarshal([]byte("{base64: aGVsbG8=}"), &s); err != nil {
			t.Fatal(err)
		}
		if s.Base64 != "aGVsbG8=" || !s.IsMapping() {
			t.Errorf("base64 decode = %+v", s)
		}
	})
}

func TestExitCode_MarshalYAML_AllShapes(t *testing.T) {
	t.Parallel()
	// The default (no field set) marshals to nil rather than a bogus map.
	if v, err := (ExitCode{}).MarshalYAML(); err != nil || v != nil {
		t.Errorf("empty exit_code marshal = (%v, %v), want (nil, nil)", v, err)
	}
	// The `in` set round-trips through YAML.
	var e ExitCode
	if err := yaml.Unmarshal([]byte("{in: [0, 2]}"), &e); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(e.In, []int{0, 2}) {
		t.Errorf("in decode = %v", e.In)
	}
	b, err := yaml.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var back ExitCode
	if err := yaml.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back.In, []int{0, 2}) {
		t.Errorf("in round-trip = %v, want [0 2]", back.In)
	}
}

// TestExitCode_UnmarshalYAML_NotError hits the mapping-decode error path: a
// non-integer `not` value cannot decode into the {not int, in []int} struct.
func TestExitCode_UnmarshalYAML_NotError(t *testing.T) {
	t.Parallel()
	var e ExitCode
	if err := yaml.Unmarshal([]byte("{not: abc}"), &e); err == nil {
		t.Error("expected error for non-integer exit_code.not")
	}
}

// TestCDPActionLabel_AllVerbs walks the whole browser action vocabulary through
// the single label helper, including the unknown-action fallback.
func TestCDPActionLabel_AllVerbs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a    CDPAction
		want string
	}{
		{CDPAction{Navigate: "u"}, "navigate u"},
		{CDPAction{WaitVisible: "#a"}, "wait_visible #a"},
		{CDPAction{WaitHidden: "#a"}, "wait_hidden #a"},
		{CDPAction{Click: "#a"}, "click #a"},
		{CDPAction{Press: &CDPPress{Selector: "#a", Key: "Enter"}}, "press Enter on #a"},
		{CDPAction{Select: &CDPSelect{Selector: "#a", Value: "v"}}, "select v in #a"},
		{CDPAction{Check: "#a"}, "check #a"},
		{CDPAction{Uncheck: "#a"}, "uncheck #a"},
		{CDPAction{Screenshot: &CDPScreenshot{Path: "s.png"}}, "screenshot s.png"},
		{CDPAction{SendKeys: &CDPSendKeys{Selector: "#a", Value: "v"}}, "send_keys #a"},
		{CDPAction{Text: "#a"}, "text #a"},
		{CDPAction{Title: true}, "title"},
		{CDPAction{Attribute: &CDPAttribute{Selector: "#a", Name: "href"}}, "attribute href of #a"},
		{CDPAction{Eval: "1+1"}, "eval"},
		{CDPAction{Upload: &CDPUpload{Selector: "#a", File: "f"}}, "upload f to #a"},
		{CDPAction{Download: &CDPDownload{Click: "#dl"}}, "download via #dl"},
		{CDPAction{}, "(unknown action)"},
	}
	for _, tc := range cases {
		if got := CDPActionLabel(tc.a); got != tc.want {
			t.Errorf("CDPActionLabel = %q, want %q", got, tc.want)
		}
	}
}

// TestWalkListPtr_EmptyStaysNonNil pins the nil-vs-empty distinction the
// changes-assert exhaustive semantics depend on, exercised directly.
func TestWalkListPtr_EmptyStaysNonNil(t *testing.T) {
	t.Parallel()
	identity := func(s string) string { return s }
	if got := walkListPtr(nil, identity); got != nil {
		t.Errorf("nil list walked to %v, want nil", got)
	}
	empty := StringList{}
	got := walkListPtr(&empty, identity)
	if got == nil || len(*got) != 0 {
		t.Errorf("empty list walked to %v, want non-nil empty", got)
	}
	full := StringList{"${a}", "b"}
	got = walkListPtr(&full, func(s string) string { return s + "!" })
	if got == nil || !reflect.DeepEqual([]string(*got), []string{"${a}!", "b!"}) {
		t.Errorf("full list walk = %v", got)
	}
}

func TestStringList_UnmarshalYAML(t *testing.T) {
	t.Parallel()
	// A scalar decodes to a one-element list (byte-identical to the pre-list
	// format); a sequence decodes to its elements.
	var one StringList
	if err := yaml.Unmarshal([]byte("just one"), &one); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string(one), []string{"just one"}) {
		t.Errorf("scalar StringList = %v", one)
	}
	var many StringList
	if err := yaml.Unmarshal([]byte("[a, b, c]"), &many); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual([]string(many), []string{"a", "b", "c"}) {
		t.Errorf("sequence StringList = %v", many)
	}
	// A mapping is neither a scalar nor a string sequence and must error.
	var bad StringList
	if err := yaml.Unmarshal([]byte("{k: v}"), &bad); err == nil {
		t.Error("expected error decoding a mapping into StringList")
	}
}

func TestCDPActionSummary(t *testing.T) {
	t.Parallel()
	c := &CDP{Runner: "web", Actions: []CDPAction{{Navigate: "http://x"}, {Click: "#a"}, {Title: true}}}
	want := "CDP via web: navigate http://x → click #a → title"
	if got := CDPActionSummary(c); got != want {
		t.Errorf("CDPActionSummary = %q, want %q", got, want)
	}
	// No actions still renders the prefix.
	if got := CDPActionSummary(&CDP{Runner: "web"}); got != "CDP via web: " {
		t.Errorf("empty CDPActionSummary = %q", got)
	}
}

func TestPTY_EnvHelpers(t *testing.T) {
	t.Parallel()
	if (&PTY{}).ClearEnvEnabled() || (&PTY{}).SandboxHomeEnabled() {
		t.Error("zero PTY should not enable clear_env or sandbox_home")
	}
	if !(&PTY{ClearEnv: boolp(true)}).ClearEnvEnabled() {
		t.Error("ClearEnv=true not reported")
	}
	if !(&PTY{SandboxHome: boolp(true)}).SandboxHomeEnabled() {
		t.Error("SandboxHome=true not reported")
	}
	// An explicit false stays false (distinguishable from unset for defaults).
	if (&PTY{ClearEnv: boolp(false)}).ClearEnvEnabled() {
		t.Error("ClearEnv=false reported as enabled")
	}
}

func TestPTYKeyForSequence(t *testing.T) {
	t.Parallel()
	// Every named key's sequence reverse-maps back to a friendly name, with the
	// documented preference for the readable name over a ctrl-* alias when a byte
	// is shared (\r is enter, not ctrl-m).
	if name, ok := PTYKeyForSequence("\r"); !ok || name != "enter" {
		t.Errorf("\\r reverse = (%q, %v), want enter", name, ok)
	}
	if name, ok := PTYKeyForSequence("\x03"); !ok || name != "ctrl-c" {
		t.Errorf("\\x03 reverse = (%q, %v), want ctrl-c", name, ok)
	}
	if name, ok := PTYKeyForSequence("\x1b[A"); !ok || name != "up" {
		t.Errorf("up-arrow reverse = (%q, %v), want up", name, ok)
	}
	// An arbitrary byte string matches no named key.
	if _, ok := PTYKeyForSequence("not a key"); ok {
		t.Error("arbitrary bytes should not reverse-map to a key")
	}
	if PTYKeyNames() == "" {
		t.Error("PTYKeyNames should be non-empty")
	}
}

// TestSetTargets_All exhaustively hits every assertion target family so a newly
// added target that SetTargets forgets is caught.
func TestSetTargets_All(t *testing.T) {
	t.Parallel()
	all := Assert{
		ExitCode:   &ExitCode{},
		Stdout:     &StreamAssert{},
		Stderr:     &StreamAssert{},
		File:       &FileAssert{},
		Status:     intp(1),
		Header:     &HeaderMatch{},
		Body:       &StreamAssert{},
		Rows:       &StreamAssert{},
		GRPCStatus: intp(0),
		Message:    &StreamAssert{},
		Value:      &StreamAssert{},
		Image:      &ImageAssert{},
		Dir:        &DirAssert{},
		PDF:        &PDFAssert{},
		Mock:       &MockAssert{},
		Screen:     &StreamAssert{},
		Duration:   &DurationAssert{},
		Changes:    &ChangesAssert{},
	}
	got := all.SetTargets()
	if len(got) != 18 {
		t.Errorf("SetTargets found %d families, want 18: %v", len(got), got)
	}
}

func TestCollectServiceVars_EnvAndNoReady(t *testing.T) {
	t.Parallel()
	set := map[string]bool{}
	// A service with env values but no readiness probe exercises the env loop
	// and the nil-Ready skip.
	CollectServiceVars(set, &Service{Command: "run", Env: map[string]string{"URL": "${svcurl}"}})
	if !set["svcurl"] {
		t.Errorf("service env var not collected: %v", SortedKeys(set))
	}
}

func TestWalkJSONValueStrings_MapAnyAny(t *testing.T) {
	t.Parallel()
	// goccy can decode a YAML mapping into map[any]any; the walker must handle
	// that shape as well as map[string]any.
	in := map[any]any{"k": "${x}", 1: "${y}"}
	out, ok := WalkJSONValueStrings(in, func(s string) string { return s + "!" }).(map[any]any)
	if !ok {
		t.Fatalf("result type = %T, want map[any]any", out)
	}
	if out["k"] != "${x}!" || out[1] != "${y}!" {
		t.Errorf("map[any]any walk = %v", out)
	}
}

func TestSortedKeys(t *testing.T) {
	t.Parallel()
	in := map[string]bool{"c": true, "a": true, "b": true}
	got := SortedKeys(in)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
	if got := SortedKeys(map[string]bool{}); len(got) != 0 {
		t.Errorf("SortedKeys empty = %v, want []", got)
	}
	// Determinism: a second call over an equivalent map yields the same order.
	again := SortedKeys(map[string]bool{"b": true, "a": true, "c": true})
	sort.Strings(want)
	if !reflect.DeepEqual(again, want) {
		t.Errorf("SortedKeys not deterministic: %v", again)
	}
}
