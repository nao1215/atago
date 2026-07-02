package engine

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/k1LoW/grpcstub"
	"github.com/nao1215/atago/internal/loader"
)

// grpcRegMu serializes grpcstub.NewServer calls. grpcstub registers the proto
// file's descriptors into the process-global protoregistry with a
// check-then-register sequence that is not atomic, so two parallel NewServer
// calls for the same proto can race and panic with "already registered" (#54).
// Serializing just the registration keeps the gRPC engine tests parallel-safe
// within the package without giving up t.Parallel().
var grpcRegMu sync.Mutex

// newGreeterStub starts a grpcstub server for the shared greeter.proto with its
// global protobuf registration serialized (see grpcRegMu). The returned server
// is closed via t.Cleanup.
func newGreeterStub(t *testing.T) *grpcstub.Server {
	t.Helper()
	grpcRegMu.Lock()
	defer grpcRegMu.Unlock()
	ts := grpcstub.NewServer(t, "testdata/greeter.proto")
	t.Cleanup(func() { ts.Close() })
	return ts
}

func TestEngine_GRPCWorkflow(t *testing.T) {
	t.Parallel()
	ts := newGreeterStub(t)
	ts.Method("SayHello").Response(map[string]any{"message": "hello alice"})

	src := fmt.Sprintf(`
version: "1"
suite:
  name: grpc
runners:
  greeter:
    type: grpc
    target: %s
scenarios:
  - name: unary call with status and message assertions plus binding
    steps:
      - grpc:
          runner: greeter
          method: atago.test.Greeter/SayHello
          json:
            name: alice
      - assert:
          grpc_status: 0
      - assert:
          message:
            json:
              path: "$.message"
              equals: hello alice
      - store:
          name: greeting
          from:
            message:
              json:
                path: "$.message"
      - grpc:
          runner: greeter
          method: atago.test.Greeter/SayHello
          json:
            name: ${greeting}
      - assert:
          grpc_status: 0
`, ts.Addr())

	res := runHTTPSpec(t, src)
	if res.Status != StatusPassed {
		t.Fatalf("status = %s, want passed: %+v", res.Status, res.Scenarios[0].Steps)
	}
}

func TestEngine_GRPCStatusError(t *testing.T) {
	t.Parallel()
	ts := newGreeterStub(t)
	ts.Method("SayHello").Response(map[string]any{"message": "hi"})

	// Asserting the wrong status (5 = NotFound) against an OK response fails.
	src := fmt.Sprintf(`
version: "1"
suite:
  name: grpc
runners:
  greeter:
    type: grpc
    target: %s
scenarios:
  - name: wrong status fails
    steps:
      - grpc:
          runner: greeter
          method: atago.test.Greeter/SayHello
          json:
            name: x
      - assert:
          grpc_status: 5
`, ts.Addr())

	res := runHTTPSpec(t, src)
	if res.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", res.Status)
	}
}

// Regression for issue #17: a grpc step to a host not on permissions.network.allow
// must be denied (StatusError + SecurityViolation), just like an HTTP step.
func TestEngine_GRPCNetworkAllowlistDenied(t *testing.T) {
	t.Parallel()
	ts := newGreeterStub(t)
	ts.Method("SayHello").Response(map[string]any{"message": "hi"})

	src := fmt.Sprintf(`
version: "1"
suite:
  name: grpc
permissions:
  network:
    allow:
      - allowed.example.com
runners:
  greeter:
    type: grpc
    target: %s
scenarios:
  - name: grpc egress to a denied host
    steps:
      - grpc:
          runner: greeter
          method: atago.test.Greeter/SayHello
`, ts.Addr())

	res := runHTTPSpec(t, src)
	if res.Status != StatusError {
		t.Fatalf("status = %s, want error (denied host)", res.Status)
	}
	if !res.SecurityViolation {
		t.Error("SecurityViolation = false, want true for a denied grpc host")
	}
}

// TestEngine_GRPCStubRegistrationParallelSafe is the regression for #54: several
// grpcstub servers for the same proto must be creatable from parallel subtests
// in one package without a global protobuf-registration panic. Each subtest runs
// in parallel, so the serialization in newGreeterStub is what keeps this green.
func TestEngine_GRPCStubRegistrationParallelSafe(t *testing.T) {
	t.Parallel()
	for i := 0; i < 6; i++ {
		t.Run(fmt.Sprintf("stub-%d", i), func(t *testing.T) {
			t.Parallel()
			ts := newGreeterStub(t)
			if ts.Addr() == "" {
				t.Fatal("stub server has no address")
			}
		})
	}
}

func TestEngine_GRPCUnknownRunner(t *testing.T) {
	t.Parallel()
	src := `
version: "1"
suite:
  name: grpc
scenarios:
  - name: grpc references an undeclared runner
    steps:
      - grpc:
          runner: missing
          method: pkg.Service/Method
`
	// An undeclared runner is a load-time validation error (exit 2), not a
	// mid-run execution error; the engine keeps a runtime check as a backstop.
	if _, err := loader.LoadBytes("t.atago.yaml", []byte(src)); err == nil || !strings.Contains(err.Error(), "is not declared") {
		t.Fatalf("LoadBytes() error = %v, want an undeclared-runner validation error", err)
	}
}
