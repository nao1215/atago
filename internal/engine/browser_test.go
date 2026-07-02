package engine

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	browserrunner "github.com/nao1215/atago/internal/runner/browser"
)

// chromeAvailable reports whether a Chrome/Chromium binary can be found, and that
// it actually launches here (sandboxed CI may have the binary but block it).
//
// When ATAGO_REQUIRE_BROWSER is set (the dedicated browser CI job, #76), a
// missing or unusable Chrome is a hard failure instead of a skip, so browser
// coverage cannot silently demote itself to "best effort" — a broken browser
// environment turns the build red.
func chromeAvailable(t *testing.T) bool {
	t.Helper()
	required := os.Getenv("ATAGO_REQUIRE_BROWSER") != ""
	unusable := func(format string, args ...any) bool {
		if required {
			t.Fatalf("ATAGO_REQUIRE_BROWSER is set but "+format, args...)
		}
		return false
	}

	found := false
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "headless-shell"} {
		if _, err := exec.LookPath(name); err == nil {
			found = true
			break
		}
	}
	if !found {
		return unusable("no Chrome/Chromium binary was found on PATH")
	}
	// Probe an actual launch so a present-but-unusable Chrome skips cleanly.
	r, err := browserrunner.Open(browserrunner.Config{Headless: true, Timeout: 20 * time.Second})
	if err != nil {
		return unusable("Chrome is present but did not launch: %v", err)
	}
	_ = r.Close()
	return true
}

func htmlServer(t *testing.T) string {
	t.Helper()
	const page = `<!doctype html><html><body>
<h1 id="title">Hello atago</h1>
<input id="name" />
<div id="echo"></div>
<script>
  document.getElementById('name').addEventListener('input', function(e){
    document.getElementById('echo').textContent = 'echo:' + e.target.value;
  });
</script>
</body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, page)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestEngine_CDPWorkflow(t *testing.T) {
	t.Parallel()
	if !chromeAvailable(t) {
		t.Skip("no usable Chrome/Chromium for the browser runner")
	}
	url := htmlServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
    timeout: 30s
scenarios:
  - name: navigate, read text, type, and eval
    steps:
      - cdp:
          runner: web
          actions:
            - navigate: %s
            - wait_visible: "#title"
            - text: "#title"
      - assert:
          value:
            equals: Hello atago
      - cdp:
          runner: web
          actions:
            - send_keys: { selector: "#name", value: world }
            - text: "#echo"
      - assert:
          value:
            contains: "echo:world"
      - cdp:
          runner: web
          actions:
            - eval: "document.getElementById('title').id"
      - assert:
          value:
            contains: title
`, url)
	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

// richHTMLServer serves a page exercising the extended CDP actions (#50): a
// spinner that hides, an input reacting to Enter, a <select>, a checkbox, and a
// link with an attribute.
func richHTMLServer(t *testing.T) string {
	t.Helper()
	const page = `<!doctype html><html><head><title>atago Test Page</title></head><body>
<div id="spinner">loading…</div>
<input id="name" />
<div id="pressed"></div>
<select id="sel"><option value="a">A</option><option value="b">B</option></select>
<div id="selected"></div>
<input type="checkbox" id="chk" />
<div id="chkstate"></div>
<a id="lnk" href="https://example.com/target" data-x="42">link</a>
<script>
  setTimeout(function(){ document.getElementById('spinner').style.display='none'; }, 100);
  document.getElementById('name').addEventListener('keydown', function(e){
    if (e.key === 'Enter') { document.getElementById('pressed').textContent = 'pressed:enter'; }
  });
  document.getElementById('sel').addEventListener('change', function(e){
    document.getElementById('selected').textContent = 'selected:' + e.target.value;
  });
  document.getElementById('chk').addEventListener('change', function(e){
    document.getElementById('chkstate').textContent = 'checked:' + e.target.checked;
  });
</script>
</body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, page)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestEngine_CDPExtendedActions drives every new black-box action end to end:
// wait_hidden, press, select, check/uncheck, title, attribute, and screenshot
// (#50). It is skipped when no usable Chrome is present.
func TestEngine_CDPExtendedActions(t *testing.T) {
	t.Parallel()
	if !chromeAvailable(t) {
		t.Skip("no usable Chrome/Chromium for the browser runner")
	}
	url := richHTMLServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
    timeout: 30s
scenarios:
  - name: extended black-box actions
    steps:
      # wait_hidden: the spinner starts visible then hides; capture the title.
      - cdp:
          runner: web
          actions:
            - navigate: %s
            - wait_hidden: "#spinner"
            - title: true
      - assert:
          value:
            equals: atago Test Page
      # press Enter on the input triggers its keydown listener.
      - cdp:
          runner: web
          actions:
            - click: "#name"
            - press: { selector: "#name", key: Enter }
            - text: "#pressed"
      - assert:
          value:
            contains: "pressed:enter"
      # select an <option> by value.
      - cdp:
          runner: web
          actions:
            - select: { selector: "#sel", value: b }
            - text: "#selected"
      - assert:
          value:
            contains: "selected:b"
      # check then read the checkbox state.
      - cdp:
          runner: web
          actions:
            - check: "#chk"
            - text: "#chkstate"
      - assert:
          value:
            contains: "checked:true"
      - cdp:
          runner: web
          actions:
            - uncheck: "#chk"
            - text: "#chkstate"
      - assert:
          value:
            contains: "checked:false"
      # capture an element attribute into the value path.
      - cdp:
          runner: web
          actions:
            - attribute: { selector: "#lnk", name: "data-x" }
      - assert:
          value:
            equals: "42"
      # screenshot writes a PNG into the workdir; file/image assertions inspect it.
      - cdp:
          runner: web
          actions:
            - screenshot: { path: shot.png }
      - assert:
          file:
            path: shot.png
            exists: true
      - assert:
          image:
            path: shot.png
            format: png
`, url)
	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

func TestEngine_CDPUnknownRunner(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: browser
scenarios:
  - name: cdp references an undeclared runner
    steps:
      - cdp:
          runner: missing
          actions:
            - navigate: http://example.com
`
	res := runHTTPSpec(t, src)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error", res.Status)
	}
}

// uploadDownloadServer serves a page that exercises the file-upload and download
// CDP actions (#75): a file input that echoes the selected file's name, and a
// link to an attachment endpoint that triggers a deterministic download.
func uploadDownloadServer(t *testing.T) string {
	t.Helper()
	const page = `<!doctype html><html><body>
<input type="file" id="file" />
<div id="uploaded"></div>
<a id="dl" href="/download">download</a>
<script>
  document.getElementById('file').addEventListener('change', function(e){
    var f = e.target.files[0];
    document.getElementById('uploaded').textContent = 'uploaded:' + (f ? f.name : '');
  });
</script>
</body></html>`
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, page)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="report.txt"`)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "hello download")
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestEngine_CDPUploadDownload(t *testing.T) {
	t.Parallel()
	if !chromeAvailable(t) {
		t.Skip("no usable Chrome/Chromium for the browser runner")
	}
	url := uploadDownloadServer(t)
	src := fmt.Sprintf(`
version: "1"
suite:
  name: browser-io
runners:
  web:
    type: browser
    timeout: 30s
scenarios:
  - name: upload a file and capture a download
    steps:
      - fixture:
          file: to_upload.txt
          content: "some contents"
      - cdp:
          runner: web
          actions:
            - navigate: %s
            - upload: { selector: "#file", file: to_upload.txt }
            - text: "#uploaded"
      - assert:
          value:
            contains: "uploaded:to_upload.txt"
      - cdp:
          runner: web
          actions:
            - download: { click: "#dl", dir: downloads }
      - assert:
          value:
            equals: report.txt
      - assert:
          file:
            path: downloads/report.txt
            contains: "hello download"
`, url)
	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}
