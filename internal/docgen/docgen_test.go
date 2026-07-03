package docgen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
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
	s, err := loader.LoadBytes("s.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("s.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("s.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("gup.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("sig.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("mixed.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
	s, err := loader.LoadBytes("v.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
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
