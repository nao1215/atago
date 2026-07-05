package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/k1LoW/grpcstub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// pastDeadlineCtx models the race window Invoke must survive: the deadline has
// already elapsed, but the local timer goroutine has not yet run, so Err() is
// still nil while Deadline() is in the past. A real context.WithDeadline cannot
// be held in this state (its Err flips the instant the deadline passes), so the
// test uses a fake.
type pastDeadlineCtx struct {
	context.Context
	deadline time.Time
}

func (c pastDeadlineCtx) Deadline() (time.Time, bool) { return c.deadline, true }
func (c pastDeadlineCtx) Err() error                  { return nil }

// TestDeadlineFailure covers the timeout-detection logic that keeps a hung
// server from passing as Result{GRPCStatus:4}. The ctx.Err()-only guard flaked
// on a server-enforced deadline whose DeadlineExceeded status arrived over the
// wire before the local timer marked the context Done (main CI:
// TestInvoke_CallTimeoutIsError).
func TestDeadlineFailure(t *testing.T) {
	t.Parallel()

	deadlineErr := status.Error(codes.DeadlineExceeded, "context deadline exceeded")

	t.Run("no invoke error is not a timeout", func(t *testing.T) {
		t.Parallel()
		if err := deadlineFailure(context.Background(), nil); err != nil {
			t.Errorf("deadlineFailure(nil) = %v, want nil", err)
		}
	})

	t.Run("a canceled context reports its error", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := deadlineFailure(ctx, deadlineErr); !errors.Is(err, context.Canceled) {
			t.Errorf("deadlineFailure(canceled) = %v, want context.Canceled", err)
		}
	})

	t.Run("an elapsed deadline with a nil Err is still a timeout", func(t *testing.T) {
		t.Parallel()
		ctx := pastDeadlineCtx{Context: context.Background(), deadline: time.Now().Add(-time.Second)}
		if err := deadlineFailure(ctx, deadlineErr); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("deadlineFailure(past deadline, nil Err) = %v, want context.DeadlineExceeded; a hung server must not pass", err)
		}
	})

	t.Run("a live deadline leaves a real status alone", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		defer cancel()
		if err := deadlineFailure(ctx, deadlineErr); err != nil {
			t.Errorf("deadlineFailure(live deadline) = %v, want nil so the status is recorded as a Result", err)
		}
	})
}

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
	// NOT parallel: grpcstub.NewServer registers the proto into the global
	// protoregistry with a check-then-register that races another grpcstub server
	// spun up in parallel (TestInvoke_RepeatedReflectionCallsSucceed), panicking
	// with "file already registered". Running sequentially, the first registration
	// wins and the second is skipped.
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
