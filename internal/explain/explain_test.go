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

// TestExplain_SuiteSetupTeardownBlocks exercises explainSuiteBlock across every
// step kind that suite.setup/teardown accepts (run, service, mock_server,
// fixture, store, assert). The validator (validateSuiteBlock) permits exactly
// these kinds, so explain must render each — a dropped kind would hide part of
// the once-per-suite bootstrap from a reviewer.
func TestExplain_SuiteSetupTeardownBlocks(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: bootstrap
  setup:
    - fixture:
        file: seed.json
        content: "{}"
    - run:
        command: ./build-helper
    - store:
        name: token
        from:
          stdout:
            matches: "tok-[0-9]+"
    - service:
        name: broker
        command: ./broker
        ready:
          delay: 50ms
    - mock_server:
        name: api
        routes:
          - {method: GET, path: /health, body: ok}
    - assert:
        exit_code: 0
  teardown:
    - run:
        command: ./cleanup
    - store:
        name: bye
        from:
          stdout:
            matches: "bye-[0-9]+"
scenarios:
  - name: trivial
    steps:
      - run: {command: echo hi}
      - assert: {exit_code: 0}
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"Suite setup (runs once before any scenario):",
		"seed.json (inline content)",
		"./build-helper",
		"start suite service \"broker\": ./broker",
		"mock server api: 1 canned route(s), serves ${api.url}",
		"store token",
		"expect exit code is 0",
		"Suite teardown (always runs after the last scenario):",
		"./cleanup",
		"store bye",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain suite block missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_SignalStep covers describeSignal for both the fire-and-forget form
// and the wait form. The default wait timeout rendered here (5s) must match the
// engine's defaultSignalWait constant; a drift would mislead a reviewer about how
// long a teardown blocks.
func TestExplain_SignalStep(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: signals
scenarios:
  - name: send signals
    services:
      - name: server
        command: ./server
        ready:
          delay: 10ms
    steps:
      - run: {command: echo go}
      - signal:
          service: server
          signal: HUP
      - signal:
          service: server
          signal: TERM
          wait: {}
      - signal:
          service: server
          signal: INT
          wait:
            timeout: 2s
      - assert: {exit_code: 0}
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"send SIGHUP to service server",
		"send SIGTERM to service server  [wait up to 5s for exit]",
		"send SIGINT to service server  [wait up to 2s for exit]",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain signal missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_ChangesAssert covers describeChangesExplain: a populated category,
// an explicitly-empty category ("modified: []" renders "modified nothing"), and
// the mix. This is the workdir-delta assertion (#70).
func TestExplain_ChangesAssert(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: delta
scenarios:
  - name: workdir changes
    steps:
      - run: {command: ./generate}
      - assert:
          changes:
            created: [out.txt, log/run.log]
            modified: []
            deleted: [stale.tmp]
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"changed exactly created out.txt, log/run.log; modified nothing; deleted stale.tmp",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain changes missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_MockAssertVariants covers describeMockAssert (#24): a bare "received
// a request", a method+path filter, and a count. It also exercises describeMockServer
// via the scenario's mock_servers list.
func TestExplain_MockAssertVariants(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: mocks
scenarios:
  - name: assert against mock
    mock_servers:
      - name: up
        routes:
          - {method: POST, path: /events, status: 202}
          - {method: GET, path: /health, body: ok}
    steps:
      - run: {command: ./client}
      - assert:
          mock:
            name: up
      - assert:
          mock:
            name: up
            method: post
            path: /events
      - assert:
          mock:
            name: up
            path: /health
            count: 3
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"mock server up: 2 canned route(s), serves ${up.url}",
		"mock up received a request",
		"mock up received POST /events",
		"mock up received /health x3",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain mock assert missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_DirAssert covers describeDir (#74) across every constraint field so
// a directory/tree assertion is fully summarized.
func TestExplain_DirAssert(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: dirs
scenarios:
  - name: check a tree
    steps:
      - run: {command: ./build}
      - assert:
          dir:
            path: dist
            exists: true
            contains: [index.html]
            not_contains: [.DS_Store]
            min_count: 1
            max_count: 100
            glob: "*.html"
            recursive: true
            ignore: [tmp, cache]
      - assert:
          dir:
            path: empty
            exists: false
      - assert:
          dir:
            path: exact
            count: 2
      - assert:
          dir:
            path: snap
            snapshot: tree.snap
`
	out := mustExplain(t, src)
	for _, want := range []string{
		`dir "dist" exists, contains index.html, does not contain .DS_Store, has >= 1 entries, has <= 100 entries, matches glob *.html, (recursive), ignoring tmp, cache`,
		`dir "empty" does not exist`,
		`dir "exact" has 2 entries`,
		`dir "snap" tree matches snapshot tree.snap`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain dir assert missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_PDFAssert covers describePDF (#73) across page-count bounds, sorted
// metadata, and extracted-text matching.
func TestExplain_PDFAssert(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: pdfs
scenarios:
  - name: check a pdf
    steps:
      - run: {command: ./render}
      - assert:
          pdf:
            path: out.pdf
            pages: 3
            metadata:
              Title: Report
              Author: nao
            text:
              contains: Summary
      - assert:
          pdf:
            path: bounded.pdf
            min_pages: 1
            max_pages: 10
`
	out := mustExplain(t, src)
	for _, want := range []string{
		`pdf "out.pdf" 3 pages`,
		// Metadata keys are sorted, so Author precedes Title regardless of map order.
		`Author contains "nao", Title contains "Report"`,
		`text contains "Summary"`,
		`pdf "bounded.pdf" >= 1 pages, <= 10 pages`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain pdf assert missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_ExitCodeInAndNot covers intList (exit code in {…}) and the
// exit-code "not" / bare forms (#19) in describeTarget.
func TestExplain_ExitCodeInAndNot(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: codes
scenarios:
  - name: exit code sets
    steps:
      - run: {command: ./maybe-fail}
      - assert:
          exit_code:
            in: [0, 2, 3]
`
	out := mustExplain(t, src)
	if !strings.Contains(out, "exit code in [0, 2, 3]") {
		t.Errorf("explain exit-code-in missing\n--- got ---\n%s", out)
	}
}

// TestExplain_JSONMatcherOperators covers jsonMatcher for the numeric-bound
// operators (gte/lt/lte) not exercised elsewhere, so every comparison branch is
// rendered.
func TestExplain_JSONMatcherOperators(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: jsonops
scenarios:
  - name: numeric bounds
    steps:
      - run: {command: echo hi}
      - assert:
          stdout:
            json:
              path: $.gte
              gte: 5
      - assert:
          stdout:
            json:
              path: $.lt
              lt: 9
      - assert:
          stdout:
            json:
              path: $.lte
              lte: 7
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"$.gte >= 5",
		"$.lt < 9",
		"$.lte <= 7",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain json operator missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_FixtureBase64AndRunNotes covers describeFixture's base64 branch and
// describeRun with a timeout and env (the notes path), plus describeStream's
// empty:true/false via body asserts on an HTTP response.
func TestExplain_FixtureBase64AndRunNotes(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: misc
runners:
  api: {type: http, base_url: "http://127.0.0.1:8080"}
scenarios:
  - name: assorted rendering
    steps:
      - fixture:
          file: blob.bin
          base64: "aGVsbG8="
      - run:
          command: ./slow
          timeout: 30s
          env:
            MODE: fast
      - http:
          runner: api
          method: GET
          path: /health
      - assert:
          body:
            empty: false
      - assert:
          body:
            empty: true
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"blob.bin (base64 binary)",
		"./slow  (timeout 30s, env: MODE)",
		"body is not empty",
		"body is empty",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain misc missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_TeardownStepKinds covers the teardown branch of explainScenario
// across every step kind teardown accepts (run, http, query, grpc, cdp, fixture,
// assert, store, signal) plus the http header `matches` matcher (describeHeader)
// and a stream YAML matcher (describeStream). A teardown block always runs, so a
// reviewer must see what external cleanup a spec performs.
func TestExplain_TeardownStepKinds(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: teardowns
runners:
  api: {type: http, base_url: "http://127.0.0.1:8080"}
  d: {type: db, dsn: "sqlite::memory:"}
  g: {type: grpc, target: "localhost:50051"}
  web: {type: browser}
scenarios:
  - name: rich teardown
    services:
      - name: worker
        command: ./worker
        ready: {delay: 10ms}
    steps:
      - run: {command: echo go}
      - assert: {exit_code: 0}
    teardown:
      - run: {command: ./cleanup}
      - http: {runner: api, method: DELETE, path: /session}
      - assert:
          header:
            name: X-Trace
            matches: "^t-[0-9]+$"
      - assert:
          body:
            yaml:
              path: $.status
              equals: done
      - query: {runner: d, sql: "DELETE FROM tmp"}
      - grpc: {runner: g, method: pkg.Svc/Cleanup}
      - cdp:
          runner: web
          actions:
            - navigate: http://localhost:8080/logout
      - fixture: {file: marker.txt, content: bye}
      - store:
          name: final
          from:
            stdout:
              matches: "done-[0-9]+"
      - signal:
          service: worker
          signal: TERM
`
	out := mustExplain(t, src)
	for _, want := range []string{
		"Teardown (always runs):",
		"./cleanup",
		"HTTP DELETE /session",
		`"X-Trace" matches /^t-[0-9]+$/`,
		"body YAML $.status",
		"SQL query via d: DELETE FROM tmp",
		"gRPC pkg.Svc/Cleanup via g",
		"CDP via web: navigate http://localhost:8080/logout",
		"marker.txt (inline content)",
		"store final",
		"send SIGTERM to service worker",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("explain teardown missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestExplain_FileContains covers describeFile's contains branch (a multi-element
// contains list renders each quoted element).
func TestExplain_FileContains(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: edges
scenarios:
  - name: file contains
    steps:
      - run: {command: ./noop}
      - assert:
          file:
            path: log.txt
            contains: [started, finished]
`
	out := mustExplain(t, src)
	if want := `file "log.txt" contains "started", "finished"`; !strings.Contains(out, want) {
		t.Errorf("explain file-contains missing %q\n--- got ---\n%s", want, out)
	}
}

// TestExplain_RunStdinVariants covers describeRun's stdin-from-file and
// binary-stdin (base64) note branches (#…): a reviewer must see where a command's
// stdin comes from.
func TestExplain_RunStdinVariants(t *testing.T) {
	t.Parallel()
	fileSrc := `
version: "1"
suite:
  name: stdinfile
scenarios:
  - name: stdin from file
    steps:
      - fixture: {file: in.txt, content: "data"}
      - run:
          command: cat
          stdin:
            file: in.txt
      - assert: {exit_code: 0}
`
	out := mustExplain(t, fileSrc)
	if !strings.Contains(out, "stdin from file in.txt") {
		t.Errorf("explain stdin-file missing\n--- got ---\n%s", out)
	}

	b64Src := `
version: "1"
suite:
  name: stdinb64
scenarios:
  - name: binary stdin
    steps:
      - run:
          command: cat
          stdin:
            base64: "aGVsbG8="
      - assert: {exit_code: 0}
`
	out = mustExplain(t, b64Src)
	if !strings.Contains(out, "binary stdin (base64)") {
		t.Errorf("explain binary-stdin missing\n--- got ---\n%s", out)
	}
}
