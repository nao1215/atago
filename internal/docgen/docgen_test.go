package docgen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/explain"
	"github.com/nao1215/atago/internal/loader"
	"github.com/nao1215/atago/internal/spec"
)

// TestGenerate_SnapshotGoldenInlined verifies a committed snapshot golden's
// content is inlined under "Expected output" so a reader sees the expected
// result without opening the snapshot file (#67).
func TestGenerate_SnapshotGoldenInlined(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "greeting.snap"), []byte("hello from atago\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: greets the world
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            snapshot: greeting.snap
`
	s := mustLoadSpec(t, "s.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: filepath.Join(dir, "s.atago.yaml"), Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, w := range []string{"#### Expected output", "stdout snapshot `greeting.snap`", "hello from atago"} {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestGenerateTo_ImageBaselineEmbedded verifies a committed image `similar_to`
// baseline is embedded as a Markdown image (relative to the output dir) so the
// expected result renders inline (#67).
func TestGenerateTo_ImageBaselineEmbedded(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	specDir := filepath.Join(root, "spec")
	outDir := filepath.Join(root, "docs")
	if err := os.MkdirAll(filepath.Join(specDir, "testdata"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "testdata", "base.png"), []byte("\x89PNG\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: output matches the baseline
    steps:
      - run: {command: gen}
      - assert:
          image:
            path: out.png
            similar_to: testdata/base.png
`
	s := mustLoadSpec(t, "s.atago.yaml", src)
	var b bytes.Buffer
	if err := GenerateTo(&b, []Source{{Path: filepath.Join(specDir, "s.atago.yaml"), Spec: s}}, outDir); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.Contains(out, "expected image `testdata/base.png`") {
		t.Errorf("missing image label\n--- got ---\n%s", out)
	}
	// The embed uses a relative link from the doc dir to the committed baseline.
	if !strings.Contains(out, "![expected image `testdata/base.png`](../spec/testdata/base.png)") {
		t.Errorf("missing/incorrect image embed\n--- got ---\n%s", out)
	}

	// Without an output dir (stdout), the baseline is not embedded (only described).
	var b2 bytes.Buffer
	if err := Generate(&b2, []Source{{Path: filepath.Join(specDir, "s.atago.yaml"), Spec: s}}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(b2.String(), "![expected image") {
		t.Errorf("stdout mode should not embed images\n--- got ---\n%s", b2.String())
	}
}

// TestGenerate_SnapshotGoldenMissingFallsBack verifies a snapshot with no
// committed file on disk (e.g. produced at runtime) renders as a reference label
// with no body rather than failing.
func TestGenerate_SnapshotGoldenMissingFallsBack(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sample
scenarios:
  - name: runtime snapshot
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            snapshot: nonexistent.snap
`
	s := mustLoadSpec(t, "s.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "s.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "stdout snapshot `nonexistent.snap`") {
		t.Errorf("missing snapshot reference label\n--- got ---\n%s", b.String())
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: gup
scenarios:
  - name: list installed tools as JSON
    steps:
      - fixture:
          file: config.yaml
          content: "{}"
      - run:
          command: gup list --json
          clear_env: true
          pass_env: [PATH]
          env:
            GOBIN: ./tmp/bin
      - assert:
          exit_code: 0
      - pty:
          command: gup interactive
          clear_env: true
          pass_env: [TERM]
          session:
            - send: ""
      - assert:
          stdout:
            json:
              path: "$[0].name"
              matches: ".+"
      - assert:
          file:
            path: out.json
            exists: true
`
	s := mustLoadSpec(t, "gup.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "gup.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	wants := []string{
		"# atago Behavior Specs",
		"## gup",
		"### Scenario: list installed tools as JSON",
		"#### Given",
		"Fixture file `config.yaml` is created.",
		"Environment variables are set: GOBIN.",
		"The command runs with a cleared environment (passing through: PATH).",
		"The command runs with a cleared environment (passing through: TERM).",
		"#### When",
		"```shell",
		"gup list --json",
		"#### Then",
		"exit code is `0`",
		"#### Generated artifacts",
		"`out.json`",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestGenerate_NonASCIIAnchorsResolve proves a non-ASCII scenario name yields a
// non-empty, unique GitHub-style anchor. Dropping non-ASCII letters collapsed
// every Japanese name to an empty "#scenario-" slug, so the TOC links all
// pointed at nothing and collided with one another.
func TestGenerate_NonASCIIAnchorsResolve(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: jp
scenarios:
  - name: "日本語シナリオ甲"
    steps:
      - run: {shell: true, command: echo a}
      - assert: {exit_code: 0}
  - name: "日本語シナリオ乙"
    steps:
      - run: {shell: true, command: echo b}
      - assert: {exit_code: 0}
`
	s := mustLoadSpec(t, "jp.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "jp.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{
		"(#scenario-日本語シナリオ甲)",
		"(#scenario-日本語シナリオ乙)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("TOC missing resolvable anchor %q\n--- got ---\n%s", want, out)
		}
	}
	if strings.Contains(out, "(#scenario-)") {
		t.Errorf("a non-ASCII name produced an empty anchor:\n%s", out)
	}
}

// TestGenerate_RendersEnvAndCommandGates proves the doc names env and command
// skip/only gates, not just the OS. A doc that showed only the OS made an
// env-gated scenario read as unconditional.
func TestGenerate_RendersEnvAndCommandGates(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: gates
scenarios:
  - name: skipped on CI
    skip: {env: CI}
    steps:
      - run: {shell: true, command: echo a}
      - assert: {exit_code: 0}
  - name: only when tool present
    only: {command: "which jq"}
    steps:
      - run: {shell: true, command: echo b}
      - assert: {exit_code: 0}
`
	s := mustLoadSpec(t, "gates.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "gates.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{
		"skipped when env CI is set",
		"only when `which jq` succeeds",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("doc missing gate prose %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestGenerate_SignalGroupsThenBullets proves a signal step is an action for
// Then-grouping (#23): a run + signal scenario (each with asserts) renders two
// "after ...:" groups instead of one flattened list — the writeThen counter
// and thenGroups must agree on what counts as an action.
func TestGenerate_SignalGroupsThenBullets(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: sig
scenarios:
  - name: shutdown flow
    services:
      - name: server
        command: ./server
    steps:
      - run:
          command: ./client warmup
      - assert:
          exit_code: 0
      - signal:
          service: server
          signal: TERM
          wait:
            timeout: 5s
      - assert:
          file:
            path: server.log
            contains: done
`
	s := mustLoadSpec(t, "sig.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "sig.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, w := range []string{
		"# send SIGTERM to service server and wait up to 5s for exit",
		"after `./client warmup`:",
		"after `SIGTERM to server`:",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestGenerate_MixedRunners verifies doc generation reaches parity with the
// supported step kinds: background services, HTTP/query/gRPC/CDP steps all
// appear in the narrative instead of only run steps (issue #41).
func TestGenerate_MixedRunners(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: mixed
runners:
  d: {type: db, dsn: "sqlite::memory:"}
  g: {type: grpc, target: "localhost:50051"}
  b: {type: browser}
scenarios:
  - name: exercises every runner
    services:
      - name: api
        command: ./serve --port 8080
    steps:
      - http:
          method: GET
          path: /health
      - assert:
          status: 200
      - assert:
          header:
            name: Content-Type
            contains: json
      - assert:
          body:
            contains: ok
      - query:
          runner: d
          sql: SELECT name FROM users
      - assert:
          rows:
            json:
              path: "$[0].name"
              equals: alice
      - grpc:
          runner: g
          method: pkg.Svc/Get
      - assert:
          grpc_status: 0
      - cdp:
          runner: b
          actions:
            - navigate: http://localhost:8080
            - wait_visible: "#title"
            - wait_hidden: "#spinner"
            - click: "#go"
            - press: {selector: "#name", key: Enter}
            - select: {selector: "#sel", value: b}
            - check: "#agree"
            - screenshot: {path: shot.png}
            - send_keys: {selector: "#name", value: alice}
            - text: "#title"
            - title: true
            - attribute: {selector: "#lnk", name: href}
            - eval: "1+1"
      - assert:
          value:
            contains: Hello
`
	s := mustLoadSpec(t, "mixed.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "mixed.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	wants := []string{
		"Background service `api` is started",
		"HTTP GET /health",
		"SQL via d: SELECT name FROM users",
		"gRPC pkg.Svc/Get via g",
		"CDP via b",
		// Extended CDP actions must appear in the generated doc (#50).
		"wait_hidden #spinner",
		"press Enter on #name",
		"select b in #sel",
		"screenshot shot.png",
		"attribute href of #lnk",
		"HTTP status is `200`",
		"gRPC status is `0`",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestGenerate_GeneratedArtifactsAcrossKinds is the regression from #56: the
// "Generated artifacts" section must include image outputs and browser
// screenshots, not only file exists:true assertions.
func TestGenerate_GeneratedArtifactsAcrossKinds(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: gen
runners:
  web: {type: browser}
scenarios:
  - name: produces several artifacts
    steps:
      - run: {command: make-thumb}
      - assert:
          image:
            path: thumb.png
            similar_to: baseline.png
            max_diff: 0.02
      - assert:
          file:
            path: out.txt
            exists: true
      - cdp:
          runner: web
          actions:
            - navigate: http://localhost:8080
            - screenshot: {path: home.png}
      - assert:
          value:
            contains: ok
`
	s := mustLoadSpec(t, "t.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "t.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, w := range []string{"Generated artifacts", "`thumb.png`", "`out.txt`", "`home.png`"} {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing generated artifact %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestGenerate_AssertVariants exercises describeAssert/describeStream/
// describeFile/describeImage/jsonMatcher across every matcher shape.
func TestGenerate_AssertVariants(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: variants
scenarios:
  - name: matchers
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code:
            not: 1
      - assert:
          stdout:
            matches: "h.+"
      - assert:
          stdout:
            not_contains: bye
      - assert:
          stdout:
            equals: hi
      - assert:
          stdout:
            not_equals: bye
      - assert:
          stdout:
            json:
              path: $.n
              equals: 1
      - assert:
          stderr:
            json:
              path: $.items
              length: 2
      - assert:
          stdout:
            snapshot: o.snap
      - assert:
          file:
            path: gone.txt
            exists: false
      - assert:
          file:
            path: x.txt
            contains: hello
      - assert:
          file:
            path: d.json
            json:
              path: $.ok
              matches: "y.+"
      - assert:
          file:
            path: o.txt
            snapshot: f.snap
      - assert:
          image:
            path: out.png
            format: png
            width: 800
            height: 600
            min_width: 1
            max_width: 2000
            min_height: 1
            max_height: 2000
            alpha: true
      - assert:
          image:
            path: t.webp
            similar_to: base.png
      - assert:
          image:
            path: flat.jpg
            alpha: false
`
	s := mustLoadSpec(t, "v.atago.yaml", src)
	var b bytes.Buffer
	if err := Generate(&b, []Source{{Path: "v.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	wants := []string{
		"exit code is not `1`",
		"stdout matches",
		"does not contain",
		"equals an exact value",
		"does not equal an exact value",
		"has length 2",
		"matches snapshot",
		"does not exist",
		"image `out.png` is `png`, width 800, height 600",
		"width >= 1, width <= 2000",
		"height >= 1, height <= 2000",
		"has alpha",
		"similar to `base.png`",
		"`flat.jpg` has no alpha",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("doc output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// mustLoadSpec loads src under the given spec path, failing the test on a load
// error. It collapses the load boilerplate shared by the docgen tests, leaving
// each test to invoke Generate/GenerateTo with its own Source wiring.
func mustLoadSpec(t *testing.T, path, src string) *spec.Spec {
	t.Helper()
	s, err := loader.LoadBytes(path, []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return s
}

func strptr(s string) *string   { return &s }
func intptr(n int) *int         { return &n }
func boolptr(b bool) *bool      { return &b }
func f64ptr(f float64) *float64 { return &f }

// oneAssertBullets renders a single assert's Then bullets via the doc renderer.
func oneAssertBullets(t *testing.T, a *spec.Assert) []string {
	t.Helper()
	return describeAsserts(a)
}

// TestDescribeTarget_MatcherMatrix drives describeTarget/describeStream/
// jsonMatcher/describeImage/describeDir/describePDF/describeChanges/describeHeader
// across every matcher and operator shape and pins the rendered phrase. A wrong
// operator symbol (> vs >=), a dropped matcher, or an off-by-one would show here.
func TestDescribeTarget_MatcherMatrix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		a    *spec.Assert
		want string
	}{
		// exit_code shapes
		{"exit not", &spec.Assert{ExitCode: &spec.ExitCode{Not: intptr(2)}}, "exit code is not `2`"},
		{"exit in", &spec.Assert{ExitCode: &spec.ExitCode{In: []int{0, 2}}}, "exit code is one of `0`, `2`"},
		{"exit equals", &spec.Assert{ExitCode: &spec.ExitCode{Equals: intptr(0)}}, "exit code is `0`"},
		{"exit bare", &spec.Assert{ExitCode: &spec.ExitCode{}}, "exit code is checked"},
		// status / grpc_status "checked" fallbacks
		{"status checked", &spec.Assert{Status: nil, Header: nil}, ""}, // placeholder, skipped below
		// stream matchers (via stdout)
		{"stdout matches", &spec.Assert{Stdout: &spec.StreamAssert{Matches: strptr("h.+")}}, "stdout matches `/h.+/`"},
		{"stdout not_matches", &spec.Assert{Stdout: &spec.StreamAssert{NotMatches: strptr("bye")}}, "stdout does not match `/bye/`"},
		{"stdout yaml", &spec.Assert{Stdout: &spec.StreamAssert{YAML: spec.JSONChecks{{Path: "$.n", Gte: f64ptr(3)}}}}, "stdout YAML at `$.n` is `>= 3`"},
		{"stdout json gt", &spec.Assert{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n", Gt: f64ptr(5)}}}}, "stdout at `$.n` is `> 5`"},
		{"stdout json lt", &spec.Assert{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n", Lt: f64ptr(9)}}}}, "stdout at `$.n` is `< 9`"},
		{"stdout json lte", &spec.Assert{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n", Lte: f64ptr(1)}}}}, "stdout at `$.n` is `<= 1`"},
		{"stdout json equals", &spec.Assert{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n", Equals: "ok"}}}}, "stdout at `$.n` equals `ok`"},
		{"stdout json matches", &spec.Assert{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n", Matches: strptr("y.+")}}}}, "stdout at `$.n` matches `/y.+/`"},
		{"stdout json bare", &spec.Assert{Stdout: &spec.StreamAssert{JSON: spec.JSONChecks{{Path: "$.n"}}}}, "stdout at `$.n` is checked"},
		{"stdout bare", &spec.Assert{Stdout: &spec.StreamAssert{}}, "stdout is checked"},
		// header matches (regression for the dropped-matcher bug)
		{"header matches", &spec.Assert{Header: &spec.HeaderMatch{Name: "Authorization", Matches: strptr("^Bearer ")}}, "header `Authorization` matches `/^Bearer /`"},
		{"header equals", &spec.Assert{Header: &spec.HeaderMatch{Name: "X", Equals: strptr("1")}}, "header `X` equals `1`"},
		{"header bare", &spec.Assert{Header: &spec.HeaderMatch{Name: "X"}}, "header `X` is checked"},
		// screen / message / value streams
		{"screen", &spec.Assert{Screen: &spec.StreamAssert{Contains: spec.StringList{"MENU"}}}, "rendered screen contains `MENU`"},
		{"message", &spec.Assert{Message: &spec.StreamAssert{Empty: boolptr(true)}}, "message is empty"},
		{"value", &spec.Assert{Value: &spec.StreamAssert{Empty: boolptr(false)}}, "value is not empty"},
		// duration
		{"duration", &spec.Assert{Duration: &spec.DurationAssert{LT: "2s", GTE: "100ms"}}, "completes in under 2s and in at least 100ms"},
		// changes
		{"changes", &spec.Assert{Changes: &spec.ChangesAssert{Created: &spec.StringList{"a.txt"}, Modified: &spec.StringList{}}}, "the step changed exactly created `a.txt`, modified nothing"},
		// mock
		{"mock", &spec.Assert{Mock: &spec.MockAssert{Name: "api", Method: "get", Path: "/v1", Count: intptr(2)}}, "mock `api` received `GET /v1` exactly 2 time(s)"},
		{"mock no route", &spec.Assert{Mock: &spec.MockAssert{Name: "api"}}, "mock `api` received a request"},
		// dir constraints
		{"dir", &spec.Assert{Dir: &spec.DirAssert{Path: "site", Exists: boolptr(true), Contains: []string{"index.html"}, NotContains: []string{"tmp"}, Count: intptr(3), MinCount: intptr(1), MaxCount: intptr(9), Glob: "*.html", Recursive: true, Snapshot: "tree.snap", Ignore: []string{"*.log"}}},
			"dir `site` exists, contains `index.html`, does not contain `tmp`, has 3 entries, has >= 1 entries, has <= 9 entries, matches glob `*.html`, tree matches snapshot `tree.snap`, (recursive), ignoring *.log"},
		{"dir absent", &spec.Assert{Dir: &spec.DirAssert{Path: "gone", Exists: boolptr(false)}}, "dir `gone` does not exist"},
		{"dir checked", &spec.Assert{Dir: &spec.DirAssert{Path: "d"}}, "dir `d` is checked"},
		// pdf
		{"pdf", &spec.Assert{PDF: &spec.PDFAssert{Path: "r.pdf", Pages: intptr(3), MinPages: intptr(1), MaxPages: intptr(9), Metadata: map[string]string{"title": "Q1", "author": "me"}, Text: &spec.StreamAssert{Contains: spec.StringList{"total"}}}},
			"pdf `r.pdf` 3 pages, >= 1 pages, <= 9 pages, author contains `me`, title contains `Q1`, text contains `total`"},
		{"pdf checked", &spec.Assert{PDF: &spec.PDFAssert{Path: "e.pdf"}}, "pdf `e.pdf` is checked"},
		// image min/max dims + no-alpha + similar
		{"image", &spec.Assert{Image: &spec.ImageAssert{Path: "o.png", MinWidth: intptr(1), MaxWidth: intptr(2), MinHeight: intptr(3), MaxHeight: intptr(4), Alpha: boolptr(false), SimilarTo: "base.png"}},
			"image `o.png` width >= 1, width <= 2, height >= 3, height <= 4, has no alpha, similar to `base.png`"},
	}
	for _, tc := range cases {
		if tc.name == "status checked" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bullets := oneAssertBullets(t, tc.a)
			if len(bullets) != 1 {
				t.Fatalf("expected 1 bullet, got %d: %v", len(bullets), bullets)
			}
			if bullets[0] != tc.want {
				t.Errorf("describe = %q\n           want %q", bullets[0], tc.want)
			}
		})
	}
}

// TestDescribeTarget_StatusAndGRPCChecked covers the nil-value "is checked"
// fallbacks that a normal spec (with a value) never reaches.
func TestDescribeTarget_StatusAndGRPCChecked(t *testing.T) {
	t.Parallel()
	if got := describeTarget(&spec.Assert{Status: intptr(0)}, spec.AssertStatus); got == "" {
		t.Error("status render empty")
	}
	// Force the nil-value branches directly (SetTargets would not surface them).
	if got := describeTarget(&spec.Assert{}, spec.AssertStatus); got != "HTTP status is checked" {
		t.Errorf("status checked = %q", got)
	}
	if got := describeTarget(&spec.Assert{}, spec.AssertGRPCStatus); got != "gRPC status is checked" {
		t.Errorf("grpc checked = %q", got)
	}
	if got := describeTarget(&spec.Assert{}, spec.AssertHeader); got != "header is checked" {
		t.Errorf("header nil = %q", got)
	}
	if got := describeTarget(&spec.Assert{Status: intptr(200)}, spec.AssertStatus); got != "HTTP status is `200`" {
		t.Errorf("status = %q", got)
	}
	if got := describeTarget(&spec.Assert{GRPCStatus: intptr(0)}, spec.AssertGRPCStatus); got != "gRPC status is `0`" {
		t.Errorf("grpc = %q", got)
	}
}

// TestDescribeChanges_Empty covers the "nothing" fallback (all categories nil).
func TestDescribeChanges_Empty(t *testing.T) {
	t.Parallel()
	if got := describeChanges(&spec.ChangesAssert{}); got != "nothing" {
		t.Errorf("empty changes = %q, want nothing", got)
	}
}

// TestDescribeAsserts_InvalidAndMultiTarget covers the no-target fallback and the
// multi-target expansion (one bullet per set target).
func TestDescribeAsserts_InvalidAndMultiTarget(t *testing.T) {
	t.Parallel()
	if got := describeAsserts(&spec.Assert{}); len(got) != 1 || got[0] != "_(invalid assertion)_" {
		t.Errorf("empty assert = %v", got)
	}
	multi := &spec.Assert{ExitCode: &spec.ExitCode{Equals: intptr(0)}, Stdout: &spec.StreamAssert{Empty: boolptr(true)}}
	if got := describeAsserts(multi); len(got) != 2 {
		t.Errorf("multi-target assert = %v, want 2 bullets", got)
	}
}

// TestStoreSourceLabel covers every StoreFrom branch.
func TestStoreSourceLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		st   *spec.Store
		want string
	}{
		{"nil from", &spec.Store{Name: "x"}, "the last result"},
		{"stdout", &spec.Store{From: &spec.StoreFrom{Stdout: &spec.StreamAssert{}}}, "stdout"},
		{"body", &spec.Store{From: &spec.StoreFrom{Body: &spec.StreamAssert{}}}, "the response body"},
		{"file", &spec.Store{From: &spec.StoreFrom{File: &spec.FileAssert{Path: "o.json"}}}, "file o.json"},
		{"header", &spec.Store{From: &spec.StoreFrom{Header: "ETag"}}, "response header ETag"},
		{"rows", &spec.Store{From: &spec.StoreFrom{Rows: &spec.StreamAssert{}}}, "the result rows"},
		{"message", &spec.Store{From: &spec.StoreFrom{Message: &spec.StreamAssert{}}}, "the response message"},
		{"value", &spec.Store{From: &spec.StoreFrom{Value: &spec.StreamAssert{}}}, "the captured value"},
		{"empty from", &spec.Store{From: &spec.StoreFrom{}}, "the last result"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := storeSourceLabel(tc.st); got != tc.want {
				t.Errorf("storeSourceLabel = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestActionLabel covers the query/grpc/cdp/default arms not hit by the run/http
// path in existing tests.
func TestActionLabel(t *testing.T) {
	t.Parallel()
	id := func(s string) string { return s }
	q := &spec.Step{Query: &spec.Query{Runner: "d", SQL: "SELECT 1"}}
	if got := actionLabel(q, id); got != "SELECT 1" {
		t.Errorf("query label = %q", got)
	}
	g := &spec.Step{GRPC: &spec.GRPC{Runner: "g", Method: "Svc/Get"}}
	if got := actionLabel(g, id); got != "gRPC Svc/Get" {
		t.Errorf("grpc label = %q", got)
	}
	c := &spec.Step{CDP: &spec.CDP{Runner: "b"}}
	if got := actionLabel(c, id); got != "the browser flow" {
		t.Errorf("cdp label = %q", got)
	}
	if got := actionLabel(&spec.Step{Fixture: &spec.Fixture{File: "x"}}, id); got != "" {
		t.Errorf("non-action label = %q, want empty", got)
	}
}

// TestFirstToken covers the empty-command fallback and a normal token.
func TestFirstToken(t *testing.T) {
	t.Parallel()
	if got := firstToken("   "); got != "   " {
		t.Errorf("blank cmd = %q, want the input", got)
	}
	if got := firstToken("git commit -m x"); got != "git" {
		t.Errorf("first token = %q, want git", got)
	}
}

// TestSplitBaseName sanitizes odd names and covers the empty-stem fallback.
func TestSplitBaseName(t *testing.T) {
	t.Parallel()
	if got := splitBaseName("dir/my spec!.atago.yaml"); got != "my-spec" {
		t.Errorf("sanitize = %q, want my-spec", got)
	}
	if got := splitBaseName("a/plain.atago.yml"); got != "plain" {
		t.Errorf("yml strip = %q, want plain", got)
	}
	// A name that sanitizes to all separators falls back to the "spec" stem.
	if got := splitBaseName("x/@@@"); got != "spec" {
		t.Errorf("all-hyphen stem = %q, want spec fallback", got)
	}
}

// TestRelPath computes a slash-separated relative link between two dirs.
func TestRelPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	base := filepath.Join(root, "docs")
	target := filepath.Join(root, "spec", "img.png")
	got, err := relPath(base, target)
	if err != nil {
		t.Fatal(err)
	}
	if got != "../spec/img.png" {
		t.Errorf("relPath = %q, want ../spec/img.png", got)
	}
}

// TestSnapshotPreview_MissingFallsBack proves an unreadable snapshot yields a
// label-only block (no body) rather than an error.
func TestSnapshotPreview_MissingFallsBack(t *testing.T) {
	t.Parallel()
	b := snapshotPreview("stdout snapshot", "nope.snap", t.TempDir())
	if b.body != "" {
		t.Errorf("missing snapshot should have no body, got %q", b.body)
	}
	if !strings.Contains(b.label, "nope.snap") {
		t.Errorf("label = %q", b.label)
	}
}

// TestImageBaselinePreview_Guards covers the early-return guards: no output dir,
// a ${workdir}-relative baseline, and a missing file.
func TestImageBaselinePreview_Guards(t *testing.T) {
	t.Parallel()
	if _, ok := imageBaselinePreview("base.png", t.TempDir(), ""); ok {
		t.Error("no output dir must return ok=false")
	}
	if _, ok := imageBaselinePreview("${workdir}/out.png", t.TempDir(), t.TempDir()); ok {
		t.Error("runtime baseline must return ok=false")
	}
	if _, ok := imageBaselinePreview("missing.png", t.TempDir(), t.TempDir()); ok {
		t.Error("missing file must return ok=false")
	}
}

// TestTruncatePreview_ByteBudget covers the byte-truncation arm (single long
// line under the line budget but over the byte budget).
func TestTruncatePreview_ByteBudget(t *testing.T) {
	t.Parallel()
	got := truncatePreview(strings.Repeat("x", previewMaxBytes+50))
	if !strings.HasSuffix(got, "… (truncated)") {
		t.Errorf("byte-truncated preview should end with the plain marker:\n%q", got[len(got)-40:])
	}
	if len(got) > previewMaxBytes+len("\n… (truncated)")+1 {
		t.Errorf("byte budget exceeded: %d", len(got))
	}
}

// --- differential: explain vs docgen must agree on the shared model ----------

// TestParity_HeaderMatchesRendered is the metamorphic regression: a header
// `matches` regexp must appear in BOTH the generated doc and the explain output.
// Before the fix, both silently dropped it (doc: "is checked", explain: name
// only), hiding a security-relevant auth-header assertion from every reader.
func TestParity_HeaderMatchesRendered(t *testing.T) {
	t.Parallel()
	src := `version: "1"
suite: {name: h}
scenarios:
  - name: auth
    steps:
      - http: {method: GET, path: /}
      - assert:
          header: {name: Authorization, matches: "^Bearer "}
`
	s, err := loader.LoadBytes("h.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var doc bytes.Buffer
	if err := Generate(&doc, []Source{{Path: "h.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.String(), "header `Authorization` matches `/^Bearer /`") {
		t.Errorf("doc dropped the header matches regexp:\n%s", doc.String())
	}
	if strings.Contains(doc.String(), "`Authorization` is checked") {
		t.Errorf("doc still renders the terse 'is checked' for a matches header:\n%s", doc.String())
	}
	var exp bytes.Buffer
	if err := explain.Explain(&exp, s, "h.atago.yaml"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(exp.String(), `"Authorization" matches /^Bearer /`) {
		t.Errorf("explain dropped the header matches regexp:\n%s", exp.String())
	}
}

// TestParity_JSONOperators pins that both renderers agree on the numeric operator
// symbols (>, >=, <, <=) for a json/yaml value bound — a swapped symbol in either
// would mislead a reader about the asserted bound.
func TestParity_JSONOperators(t *testing.T) {
	t.Parallel()
	src := `version: "1"
suite: {name: ops}
scenarios:
  - name: bounds
    steps:
      - run: {command: metrics}
      - assert: {stdout: {json: {path: "$.gt", gt: 1}}}
      - assert: {stdout: {json: {path: "$.gte", gte: 2}}}
      - assert: {stdout: {json: {path: "$.lt", lt: 3}}}
      - assert: {stdout: {json: {path: "$.lte", lte: 4}}}
`
	s, err := loader.LoadBytes("ops.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var doc bytes.Buffer
	if err := Generate(&doc, []Source{{Path: "ops.atago.yaml", Spec: s}}); err != nil {
		t.Fatal(err)
	}
	var exp bytes.Buffer
	if err := explain.Explain(&exp, s, "ops.atago.yaml"); err != nil {
		t.Fatal(err)
	}
	d, e := doc.String(), exp.String()
	// doc uses "is `> 1`"; explain uses "> 1". Both must carry the exact symbol.
	for _, sym := range []string{"> 1", ">= 2", "< 3", "<= 4"} {
		if !strings.Contains(d, sym) {
			t.Errorf("doc missing operator %q:\n%s", sym, d)
		}
		if !strings.Contains(e, sym) {
			t.Errorf("explain missing operator %q:\n%s", sym, e)
		}
	}
}

// TestGenerateSplit_WithOutputDir exercises GenerateSplit's outputDir path and
// the index generator, plus mdEscape of a suite name with markdown-active chars.
func TestGenerateSplit_WithOutputDir(t *testing.T) {
	t.Parallel()
	a := load(t, "one.atago.yaml", `version: "1"
suite: {name: "a[1]"}
scenarios: [{name: s, steps: [{run: {command: "true"}}, {assert: {exit_code: 0}}]}]
`)
	index, docs, err := GenerateSplit([]Source{*a}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].Name != "one.md" {
		t.Fatalf("docs = %+v", docs)
	}
	if !bytes.Contains(index, []byte("Documents")) {
		t.Errorf("index missing Documents section:\n%s", index)
	}
}
