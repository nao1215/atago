package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/nao1215/atago/internal/buildinfo"
)

// starterSpec is the default scaffold written by `atago init` (the `cli`
// template). It is a valid, runnable spec that demonstrates the common building
// blocks: a command run, exit-code and stdout assertions, and an empty-stderr
// check.
const starterSpec = `version: "1"

# Runs green as-is: atago run example.atago.yaml
# Replace the echo command with your own CLI and grow the assertions from there.

suite:
  name: example

scenarios:
  - name: echo greets the world
    steps:
      - run:
          # shell: true keeps this portable: on Windows echo is a shell
          # builtin, not a standalone executable.
          shell: true
          command: echo "hello atago"
      - assert:
          exit_code: 0
          stdout:
            contains: atago
          stderr:
            empty: true
`

// httpTemplate scaffolds an HTTP peer-testing spec: a named http runner and a
// request/response assertion against a JSON API. It is schema-valid out of the
// box; point base_url at a live service to run it.
const httpTemplate = `version: "1"

# Before running: point base_url at your API. As written, the request goes to
# https://api.example.com (which does not exist), so the first run fails with a
# connection error — that failure is expected until you edit base_url.
#   atago run http.atago.yaml

suite:
  name: http-example

runners:
  api:
    type: http
    base_url: https://api.example.com   # <-- edit: your API's base URL

scenarios:
  - name: health endpoint returns ok
    steps:
      - http:
          runner: api
          method: GET
          path: /health                  # <-- edit: a real endpoint of your API
      - assert:
          status: 200
      - assert:
          body:
            json:
              path: $.status
              equals: ok
`

// dbTemplate scaffolds a database spec: a sqlite runner (isolated per scenario
// under ${workdir}) and a query/rows assertion. It is runnable with the bundled
// sqlite driver.
const dbTemplate = `version: "1"

# Runs green as-is (the sqlite driver is bundled): atago run db.atago.yaml
# For a real database, change dsn to postgres://... or mysql://...

suite:
  name: db-example

runners:
  store:
    type: db
    # A fresh, isolated SQLite database is created per scenario under the
    # scenario workdir, so runs never touch each other's state.
    dsn: sqlite:${workdir}/app.db

scenarios:
  - name: seeded row is queryable
    steps:
      - query:
          runner: store
          sql: "CREATE TABLE users (id INTEGER, name TEXT)"
      - query:
          runner: store
          sql: "INSERT INTO users (id, name) VALUES (1, 'ada')"
      - query:
          runner: store
          sql: "SELECT name FROM users WHERE id = 1"
      - assert:
          rows:
            json:
              path: $[0].name
              equals: ada
`

// grpcTemplate scaffolds a gRPC spec: a named grpc runner (schema resolved via
// server reflection) and a unary call with a message assertion. It is
// schema-valid; point target at a live reflection-enabled server to run it.
const grpcTemplate = `version: "1"

# Before running: point target at a gRPC server with server reflection enabled,
# and set method to one of its services. Until then the call fails with a
# connection error.
#   atago run grpc.atago.yaml

suite:
  name: grpc-example

runners:
  greeter:
    type: grpc
    target: localhost:50051              # <-- edit: host:port of your server

scenarios:
  - name: unary call echoes the name
    steps:
      - grpc:
          runner: greeter
          method: helloworld.Greeter/SayHello   # <-- edit: package.Service/Method
          json:
            name: atago
      - assert:
          grpc_status: 0
      - assert:
          message:
            contains: atago
`

// sshTemplate scaffolds an SSH spec: a named ssh runner and a run step executed
// on the remote host. It is schema-valid; edit host/user/key_file to point at a
// reachable machine before running it.
const sshTemplate = `version: "1"

# Before running: set host, user, and key_file (or password) to a machine you
# can reach. Until then the connection fails.
#   atago run ssh.atago.yaml

suite:
  name: ssh-example

runners:
  box:
    type: ssh
    host: ssh.example.com                # <-- edit: host or host:port (default 22)
    user: deploy                         # <-- edit: login user
    key_file: ~/.ssh/id_ed25519          # <-- edit: private key (or password: ...)
    known_hosts: ~/.ssh/known_hosts      # verifies the host key (recommended)
    # insecure_host_key: true            # test/lab only: skip host-key checks

scenarios:
  - name: remote command reports the kernel
    steps:
      # runner: box sends the command over SSH; assertions work exactly like a
      # local run (exit_code / stdout / stderr).
      - run:
          runner: box
          command: uname -s
      - assert:
          exit_code: 0
          stdout:
            contains: Linux
`

// browserTemplate scaffolds a browser (CDP) spec: a named browser runner and a
// small navigate/capture flow with a value assertion. It is schema-valid; it
// runs where a Chrome/Chromium binary is available.
const browserTemplate = `version: "1"

# Needs a Chrome/Chromium binary on PATH and network access to example.com.
#   atago run browser.atago.yaml
# Point navigate at your own app (e.g. http://localhost:8080) to test it.

suite:
  name: browser-example

runners:
  web:
    type: browser
    # headless: false  # uncomment to watch the browser while debugging

scenarios:
  - name: homepage title is captured
    steps:
      - cdp:
          runner: web
          actions:
            - navigate: https://example.com
            - wait_visible: h1
            - text: h1
      - assert:
          value:
            contains: Example Domain
`

// servicesTemplate scaffolds a background-service spec that runs green as-is
// while teaching the real workflow: wait for readiness, capture the published
// address, poll the peer with retry, and assert on data the peer produced.
const servicesTemplate = `version: "1"

# Runs green as-is: atago run services.atago.yaml
# The "server" below is a shell stand-in so this works anywhere. For a real
# server: set command to your server binary, and prefer a real readiness probe —
#   ready: { port: 127.0.0.1:8080 }   waits until the TCP port accepts
#   ready: { log: "listening" }       waits for a line on the service output

suite:
  name: services-example

scenarios:
  - name: client waits for the service and reads its response
    services:
      - name: server
        shell: true
        # Like a real daemon: publish the bound address once ready, respond a
        # moment later, and keep running until the scenario tears it down.
        command: |
          echo "127.0.0.1:8000" > addr.txt
          sleep 1
          echo "pong" > response.txt
          sleep 30
        ready:
          file: addr.txt   # ready when this file exists and is non-empty
          store: addr      # its trimmed content becomes ${addr}
          timeout: 5s
    steps:
      # ${addr} carries the address the service published while becoming ready.
      - run:
          shell: true
          command: echo "client would connect to ${addr}"
      - assert:
          stdout:
            contains: "127.0.0.1:8000"
      # Poll the service's response the way you would poll a real endpoint:
      # retry re-runs the command until the until-assertion passes.
      - run:
          shell: true
          command: cat response.txt
          retry:
            times: 20
            interval: 100ms
            until:
              stdout:
                contains: pong
      - assert:
          exit_code: 0
          stdout:
            contains: pong
`

const mockTemplate = `version: "1"

# Test an API-client CLI OFFLINE (#24): the mock server serves canned routes
# on an ephemeral loopback port and records every request, so you can assert
# what your CLI actually sent. Runs as-is with curl on PATH:
#   atago run mock.atago.yaml
# For your own CLI: replace the curl command with e.g.
#   mycli push --endpoint ${api.url} report.txt

suite:
  name: mock-example

scenarios:
  - name: the client posts a report and the mock records it
    mock_servers:
      - name: api
        routes:
          - method: POST
            path: /v1/reports
            status: 201
            json: { id: "r-1", ok: true }
          - method: GET
            path: /v1/reports/r-1
            json: { id: "r-1", title: "report" }
    steps:
      # ${api.url} is the mock's base URL, seeded before steps run.
      - run:
          shell: true
          command: >-
            curl -sf -X POST -H 'Authorization: Bearer tok-123'
            -d '{"title":"report"}' ${api.url}/v1/reports
      - assert:
          exit_code: 0
          stdout:
            json: { path: "$.id", equals: "r-1" }
      # Assert what the CLI actually sent: exactly one POST, with the auth
      # header and the JSON body it was supposed to carry.
      - assert:
          mock:
            name: api
            path: /v1/reports
            method: POST
            count: 1
            header: { name: Authorization, matches: "^Bearer " }
            body:
              json: { path: "$.title", equals: "report" }
`

// initTemplate pairs a scaffold with the one-line summary shown by
// --list-templates, so a user can pick a starting point without opening the
// generated file first.
type initTemplate struct {
	body string
	// desc states what the template tests and what the first run needs
	// (runnable as-is, or the field to edit first).
	desc string
}

// initTemplates maps each template name to its scaffold. Every template is
// exercised by the init tests, which load and validate the rendered YAML so a
// broken template fails before it ships (#65).
var initTemplates = map[string]initTemplate{
	"cli": {
		body: starterSpec,
		desc: "run a command; assert exit code/stdout/stderr (runs as-is)",
	},
	"http": {
		body: httpTemplate,
		desc: "call an HTTP API; assert status and JSON body (edit base_url first)",
	},
	"db": {
		body: dbTemplate,
		desc: "run SQL; assert on rows via bundled SQLite (runs as-is)",
	},
	"grpc": {
		body: grpcTemplate,
		desc: "call a unary gRPC method via server reflection (edit target first)",
	},
	"ssh": {
		body: sshTemplate,
		desc: "run a command on a remote host over SSH (edit host/user first)",
	},
	"browser": {
		body: browserTemplate,
		desc: "drive a headless Chrome; assert page content (needs Chrome on PATH)",
	},
	"services": {
		body: servicesTemplate,
		desc: "test against a background server: readiness, retry, teardown (runs as-is)",
	},
	"mock": {
		body: mockTemplate,
		desc: "stub an HTTP API offline and assert what the client sent (needs curl on PATH)",
	},
}

// initTemplateNames returns the template names in stable, sorted order for help
// text and listing.
func initTemplateNames() []string {
	names := make([]string, 0, len(initTemplates))
	for name := range initTemplates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// defaultInitFilename picks the scaffold file name for a template so multiple
// templates can coexist in one directory without --force.
func defaultInitFilename(template string) string {
	if template == "cli" {
		return "example.atago.yaml"
	}
	return template + ".atago.yaml"
}

// initCmd implements `atago init`: scaffold a starter spec file.
// With --template it emits a runner-oriented starter (cli/http/db/grpc/ssh/
// browser/services); without it, the portable cli starter.
func initCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("atago init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	force := fs.Bool("force", false, "overwrite the file if it already exists")
	template := fs.String("template", "cli", "starter template: "+strings.Join(initTemplateNames(), "|"))
	list := fs.Bool("list-templates", false, "list the available templates and exit")
	printUsage := func(w io.Writer) {
		fmt.Fprintf(w, "Usage: atago init [--force] [--template %s] [path]\n", strings.Join(initTemplateNames(), "|"))
		fs.SetOutput(w)
		fs.PrintDefaults()
	}
	// Suppress the flag package's automatic usage print; usage is routed
	// explicitly below — to stdout for an explicit --help (so it can be piped),
	// to stderr for a genuine parse error.
	fs.Usage = func() {}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(stdout)
			return ExitOK
		}
		printUsage(stderr)
		return ExitConfig
	}

	if *list {
		for _, name := range initTemplateNames() {
			fmt.Fprintf(stdout, "%-9s %s\n", name, initTemplates[name].desc)
		}
		fmt.Fprintln(stdout, "\nScaffold one with: atago init --template <name>")
		return ExitOK
	}

	tmpl, ok := initTemplates[*template]
	if !ok {
		fmt.Fprintf(stderr, "atago init: unknown template %q (want %s)\n", *template, strings.Join(initTemplateNames(), ", "))
		return ExitConfig
	}

	if fs.NArg() > 1 {
		fmt.Fprintf(stderr, "atago init: too many paths — init writes one spec file, got %d (%s)\n", fs.NArg(), strings.Join(fs.Args(), ", "))
		return ExitConfig
	}
	path := defaultInitFilename(*template)
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	if _, err := os.Stat(path); err == nil && !*force {
		fmt.Fprintf(stderr, "atago init: %q already exists (use --force to overwrite)\n", path)
		return ExitConfig
	}

	// Prepend a resolvable schema header so a freshly scaffolded spec gets
	// editor completion for step types, matchers, and ${...} forms out of the
	// box — the only delivery path for the DSL reference (#121).
	body := buildinfo.SchemaHeader() + tmpl.body
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		fmt.Fprintf(stderr, "atago init: %v\n", err)
		return ExitConfig
	}
	fmt.Fprintf(stdout, "Created %s\n", path)
	return ExitOK
}
