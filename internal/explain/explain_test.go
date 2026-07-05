package explain

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nao1215/atago/internal/loader"
)

func TestExplain(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: demo
secrets:
  - API_TOKEN
permissions:
  network:
    allow:
      - localhost
scenarios:
  - name: fetch and check
    tags: [network]
    steps:
      - fixture:
          file: in.json
          content: "{}"
      - run:
          shell: true
          command: curl http://localhost:8080/health
      - assert:
          exit_code: 0
      - assert:
          stdout:
            contains: ok
      - assert:
          file:
            path: out.txt
            exists: true
`
	out := mustExplain(t, src)
	wants := []string{
		"Suite: demo",
		"Secrets declared: API_TOKEN",
		"Network policy: allow localhost",
		"Scenario: fetch and check",
		"[tags: network]",
		"in.json (inline content)",
		"curl http://localhost:8080/health",
		"exit code is 0",
		`stdout contains "ok"`,
		"Generates:",
		"out.txt",
		"Security notes",
		"shell execution enabled",
		"network access",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestExplain_NoAllowlistIsUnrestricted verifies that when no network allowlist
// is set, explain describes the network as unrestricted, matching runtime
// semantics where an empty allowlist permits every host (issue #41).
func TestExplain_NoAllowlistIsUnrestricted(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: nolist
scenarios:
  - name: plain
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
`
	out := mustExplain(t, src)
	if strings.Contains(out, "restricted") && !strings.Contains(out, "unrestricted") {
		t.Errorf("no-allowlist network should not be described as restricted:\n%s", out)
	}
	if !strings.Contains(out, "unrestricted") {
		t.Errorf("expected network policy to say unrestricted:\n%s", out)
	}
}

// Regression for issue #15: explain must describe query/grpc/cdp steps and
// their assertion targets, not silently omit them.
func TestExplain_QueryGRPCCDPSteps(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: runners
runners:
  d: {type: db, dsn: "sqlite::memory:"}
  g: {type: grpc, target: "localhost:50051"}
  b: {type: browser}
scenarios:
  - name: exercises db, grpc, cdp
    steps:
      - query:
          runner: d
          sql: SELECT name FROM users WHERE id = ${uid}
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
      - assert:
          message:
            contains: ok
      - cdp:
          runner: b
          actions:
            - navigate: http://localhost:8080
            - wait_visible: "#title"
            - text: "#title"
      - assert:
          value:
            contains: Hello
`
	out := mustExplain(t, src)
	wants := []string{
		"SQL query via d: SELECT name FROM users",
		"gRPC pkg.Svc/Get via g",
		"CDP via b: navigate http://localhost:8080 → wait_visible #title → text #title",
		"rows JSON $[0].name",
		"gRPC status is 0",
		"message contains \"ok\"",
		"value contains \"Hello\"",
		"Variables used: uid",
		"network access: gRPC pkg.Svc/Get",
		"browser automation (CDP) via b",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestExplain_GeneratedArtifactsAcrossKinds is the regression from #56: explain
// must surface generated files from image assertions and cdp screenshots, not
// only from file exists:true — so its "Generates" list matches doc and manifest.
func TestExplain_GeneratedArtifactsAcrossKinds(t *testing.T) {
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
	out := mustExplain(t, src)
	for _, w := range []string{"Generates", "thumb.png", "out.txt", "home.png"} {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing generated artifact %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestExplain_ExtendedCDPActions is the regression from #50: explain must
// describe the extended black-box CDP actions so its static summary stays aligned
// with the runtime action set.
func TestExplain_ExtendedCDPActions(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: browser
runners:
  b: {type: browser}
scenarios:
  - name: a black-box ui flow
    steps:
      - cdp:
          runner: b
          actions:
            - navigate: http://localhost:8080
            - wait_hidden: "#spinner"
            - press: {selector: "#in", key: Enter}
            - select: {selector: "#s", value: b}
            - check: "#agree"
            - uncheck: "#news"
            - screenshot: {path: shot.png}
            - title: true
            - attribute: {selector: "#lnk", name: href}
      - assert:
          value:
            contains: ok
`
	out := mustExplain(t, src)
	for _, w := range []string{
		"wait_hidden #spinner",
		"press Enter on #in",
		"select b in #s",
		"check #agree",
		"uncheck #news",
		"screenshot shot.png",
		"title",
		"attribute href of #lnk",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// TestExplain_ServicesAndHeader covers describeService (every readiness mode),
// the header assertion branch (describeHeader), and env listing (toSet) so the
// service- and header-aware paths are exercised.
func TestExplain_ServicesAndHeader(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: svc
runners:
  api: {type: http, base_url: "http://127.0.0.1:8080"}
permissions:
  network:
    allow: [127.0.0.1]
scenarios:
  - name: service with ready-file and header assert
    services:
      - name: peer
        command: ./serve --addr ${workdir}/sock
        env:
          LOG_LEVEL: debug
        ready:
          file: sock
          store: addr
    steps:
      - http:
          runner: api
          method: GET
          path: /health
      - assert:
          header:
            name: Content-Type
            contains: json
      - assert:
          header:
            name: X-Exact
            equals: "1"
  - name: service ready by port
    services:
      - name: p
        command: ./p
        ready:
          port: "127.0.0.1:9000"
    steps:
      - run: {command: echo ok}
      - assert: {exit_code: 0}
  - name: service ready by log
    services:
      - name: l
        command: ./l
        ready:
          log: "listening on"
    steps:
      - run: {command: echo ok}
      - assert: {exit_code: 0}
  - name: service ready by delay
    services:
      - name: d
        command: ./d
        ready:
          delay: 100ms
    steps:
      - run: {command: echo ok}
      - assert: {exit_code: 0}
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"Services:",
		"peer: ./serve",
		"ready when file sock appears",
		"${addr}",
		"ready when port 127.0.0.1:9000 accepts",
		"ready when log matches /listening on/",
		"ready after 100ms",
		`"Content-Type" contains "json"`,
		`"X-Exact" equals "1"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain output missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_HermeticEnv proves explain surfaces clear_env/pass_env on run
// steps and services (#16) so a reviewer sees the hermetic-environment intent.
func TestExplain_HermeticEnv(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: hermetic
scenarios:
  - name: cleared
    services:
      - name: srv
        command: ./srv
        clear_env: true
        pass_env: [PATH]
    steps:
      - run:
          command: env
          clear_env: true
          pass_env: [PATH, HOME]
      - assert:
          exit_code: 0
      - run:
          command: env
          clear_env: true
      - assert:
          exit_code: 0
      - pty:
          command: cat
          clear_env: true
          pass_env: [TERM]
          session:
            - send: ""
      - assert:
          exit_code: 0
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"cleared environment (passes: PATH, HOME)",
		"cleared environment)",
		"[cleared environment, passes: PATH]",
		"(cleared environment, passes: TERM)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain output missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_AssertVariants exercises describeAssert/describeStream/
// describeFile/describeImage/jsonMatcher across every matcher shape.
func TestExplain_AssertVariants(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: variants
scenarios:
  - name: all matchers
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
              path: $.name
              equals: alice
      - assert:
          stderr:
            json:
              path: $.items
              length: 3
      - assert:
          stdout:
            json:
              path: $.id
              matches: "[0-9]+"
      - assert:
          stdout:
            json:
              path: $.count
              gt: 5
      - assert:
          stdout:
            snapshot: out.snap
      - assert:
          file:
            path: gone.txt
            exists: false
      - assert:
          file:
            path: data.json
            json:
              path: $.ok
              equals: true
      - assert:
          file:
            path: out.txt
            snapshot: file.snap
      - assert:
          image:
            path: out.png
            format: png
            width: 800
            height: 600
            min_width: 10
            max_width: 1000
            min_height: 10
            max_height: 1000
            alpha: true
      - assert:
          image:
            path: thumb.webp
            similar_to: baseline.png
            max_diff: 0.02
      - assert:
          image:
            path: flat.jpg
            alpha: false
`
	out := mustExplain(t, src)
	wants := []string{
		"exit code is not 1",
		"stdout matches /h.+/",
		`stdout does not contain "bye"`,
		"stdout equals exact text",
		"stdout does not equal exact text",
		"JSON $.name",
		"length 3",
		"matches /[0-9]+/",
		"$.count > 5",
		"matches snapshot out.snap",
		`"gone.txt" does not exist`,
		`"data.json" JSON $.ok`,
		`"out.txt" matches snapshot file.snap`,
		`"out.png" is png, width 800, height 600`,
		"width >= 10, width <= 1000",
		"height >= 10, height <= 1000",
		"has alpha",
		`"thumb.webp" similar to baseline.png`,
		`"flat.jpg" has no alpha`,
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing %q\n--- got ---\n%s", w, out)
		}
	}
}

// mustExplain loads src and returns the rendered explain text, failing the test
// on any load or render error. It collapses the load+Explain boilerplate shared
// by the explain tests.
func mustExplain(t *testing.T, src string) string {
	t.Helper()
	s, err := loader.LoadBytes("t.atago.yaml", []byte(src))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	var b bytes.Buffer
	if err := Explain(&b, s, "t.atago.yaml"); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// TestExplain_CollectsAllVariableBearingFields is a regression test for a bug
// where the explain "Variables used" summary under-reported ${name} references:
// it scanned only http path/body and part of the cdp action set, missing run
// env, http body_file/body_to/form/files, and cdp upload/download (the manifest
// already scanned them). explain and manifest now share spec.CollectStepVars,
// so every variable-bearing field is counted. Each variable below is referenced
// only from a field the old code skipped, so a miss regresses the fix.
func TestExplain_CollectsAllVariableBearingFields(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: vars
runners:
  api:
    type: http
    base_url: http://localhost:9999
  web:
    type: browser
scenarios:
  - name: exercises previously-unscanned variable-bearing fields
    steps:
      - run:
          shell: true
          command: echo hi
          env:
            TOKEN: "${runenv_ref}"
      - http:
          runner: api
          method: PUT
          path: /put
          body_file: "${bodyfile_ref}"
          body_to: "${bodyto_ref}"
      - http:
          runner: api
          method: POST
          path: /form
          form:
            field1: "${form_ref}"
          files:
            - field: upload
              path: "${files_ref}"
      - cdp:
          runner: web
          actions:
            - navigate: http://localhost
            - upload:
                selector: "#file"
                file: "${upload_ref}"
            - download:
                click: "#dl"
                dir: "${download_ref}"
`
	out := mustExplain(t, src)
	for _, v := range []string{
		"runenv_ref", "bodyfile_ref", "bodyto_ref",
		"form_ref", "files_ref", "upload_ref", "download_ref",
	} {
		if !strings.Contains(out, v) {
			t.Errorf("explain output missing variable %q from a previously-unscanned field\n--- got ---\n%s", v, out)
		}
	}
}
