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
		{in: "NoSlash", wantErr: true},
		{in: "/Method", wantErr: true},
		{in: "Svc/", wantErr: true},
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
