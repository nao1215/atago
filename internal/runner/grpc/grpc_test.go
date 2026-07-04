package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/k1LoW/grpcstub"
)

func TestSplitMethod(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		svc     string
		method  string
		wantErr bool
	}{
		{in: "pkg.Service/Method", svc: "pkg.Service", method: "Method"},
		{in: "a.b.c.Svc/Do", svc: "a.b.c.Svc", method: "Do"},
		// A single leading slash (the fully-qualified gRPC form) is tolerated.
		{in: "/pkg.Service/Method", svc: "pkg.Service", method: "Method"},
		{in: "NoSlash", wantErr: true},
		{in: "/Method", wantErr: true},
		{in: "Svc/", wantErr: true},
		// More than one internal slash is a malformed method, not a nested service.
		{in: "a/b/c", wantErr: true},
		{in: "/pkg.Service/Method/extra", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		svc, method, err := splitMethod(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("splitMethod(%q) error = nil, want error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("splitMethod(%q) error = %v", tt.in, err)
			continue
		}
		if svc != tt.svc || method != tt.method {
			t.Errorf("splitMethod(%q) = %q, %q; want %q, %q", tt.in, svc, method, tt.svc, tt.method)
		}
	}
}

// TestInvoke_CallTimeoutIsError is a regression: a unary call that hangs past
// the per-call timeout must be a hard error, not a passing Result. status.From
// Error maps a client-deadline DeadlineExceeded to ok=true, which would
// otherwise be recorded as a normal Result{GRPCStatus:4} and pass against a hung
// server unless the spec happened to assert grpc_status.
func TestInvoke_CallTimeoutIsError(t *testing.T) {
	t.Parallel()
	ts := grpcstub.NewServer(t, "testdata/greeter.proto")
	t.Cleanup(func() { ts.Close() })
	ts.Method("SayHello").Handler(func(_ *grpcstub.Request) *grpcstub.Response {
		time.Sleep(2 * time.Second) // longer than the client's per-call timeout
		return grpcstub.NewResponse()
	})

	r, err := Open(Config{Target: ts.Addr(), Timeout: 300 * time.Millisecond})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })

	out, err := r.Invoke(context.Background(), "atago.test.Greeter/SayHello", nil, []byte(`{"name":"x"}`))
	if err == nil {
		t.Fatalf("Invoke against a hung handler returned no error; got Result %+v (a timed-out call must be an error, not a passing status)", out)
	}
}

func TestOpen_RequiresTarget(t *testing.T) {
	t.Parallel()
	if _, err := Open(Config{}); err == nil {
		t.Error("Open with empty target should error")
	}
}

// Regression for issue #29: resolveMethod now defers refClient.Reset() to drain
// the per-call reflection stream. This test guards that Reset does not break
// reflection for subsequent calls on the same connection: many sequential
// invocations on one Runner must all resolve their method and succeed.
func TestInvoke_RepeatedReflectionCallsSucceed(t *testing.T) {
	t.Parallel()
	ts := grpcstub.NewServer(t, "testdata/greeter.proto")
	t.Cleanup(func() { ts.Close() })
	ts.Method("SayHello").Response(map[string]any{"message": "hi"})

	r, err := Open(Config{Target: ts.Addr(), Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })

	const method = "atago.test.Greeter/SayHello"
	for i := 0; i < 20; i++ {
		out, err := r.Invoke(context.Background(), method, nil, []byte(`{"name":"x"}`))
		if err != nil {
			t.Fatalf("invoke %d after prior Reset: %v", i, err)
		}
		if out.GRPCStatus != 0 {
			t.Fatalf("invoke %d status = %d, want 0", i, out.GRPCStatus)
		}
	}
}
