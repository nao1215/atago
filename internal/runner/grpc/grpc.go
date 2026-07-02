// Package grpc implements the gRPC runner: a `grpc` step calls a unary method on
// a target server and captures the response message (as JSON) and status code as
// a runner.Result (ADR-0028). It is the atago counterpart to runn's
// gRPC runner and resolves the service schema dynamically via server reflection
// (jhump/protoreflect/v2 + google.golang.org/grpc) — no compiled stubs, keeping
// specs declarative.
package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jhump/protoreflect/v2/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/nao1215/atago/internal/runner"
)

// Config is the resolved configuration for a grpc runner.
type Config struct {
	// Target is the host:port of the gRPC server.
	Target string
	// TLS enables transport security; the default is plaintext.
	TLS bool
	// Timeout bounds a single call; zero means none.
	Timeout time.Duration
}

// Runner holds a gRPC client connection for one grpc runner.
type Runner struct {
	cc      *grpc.ClientConn
	timeout time.Duration
}

// Open establishes the gRPC client connection (lazily — grpc.NewClient does not
// dial until the first call).
func Open(cfg Config) (*Runner, error) {
	if strings.TrimSpace(cfg.Target) == "" {
		return nil, errors.New("grpc runner requires a target")
	}
	var creds credentials.TransportCredentials
	if cfg.TLS {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		creds = insecure.NewCredentials()
	}
	cc, err := grpc.NewClient(cfg.Target, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("grpc connect %s: %w", cfg.Target, err)
	}
	return &Runner{cc: cc, timeout: cfg.Timeout}, nil
}

// Close releases the connection.
func (r *Runner) Close() error { return r.cc.Close() }

// Invoke calls a unary method ("pkg.Service/Method") with an optional JSON
// request body and optional headers, returning the response. A non-OK gRPC
// status is a successful Invoke with the code recorded on the Result; only a
// schema-resolution or transport failure returns an error.
func (r *Runner) Invoke(ctx context.Context, method string, header map[string]string, reqJSON []byte) (*runner.Result, error) {
	if r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
	}
	service, methodName, err := splitMethod(method)
	if err != nil {
		return nil, err
	}
	md, err := r.resolveMethod(ctx, service, methodName)
	if err != nil {
		return nil, err
	}

	req := dynamicpb.NewMessage(md.Input())
	if len(reqJSON) > 0 {
		if err := protojson.Unmarshal(reqJSON, req); err != nil {
			return nil, fmt.Errorf("encoding grpc request for %s: %w", method, err)
		}
	}
	if len(header) > 0 {
		pairs := make([]string, 0, len(header)*2)
		for k, v := range header {
			pairs = append(pairs, k, v)
		}
		ctx = metadata.AppendToOutgoingContext(ctx, pairs...)
	}

	res := dynamicpb.NewMessage(md.Output())
	invErr := r.cc.Invoke(ctx, "/"+service+"/"+methodName, req, res)
	stat, ok := status.FromError(invErr)
	if !ok {
		return nil, fmt.Errorf("grpc invoke %s: %w", method, invErr)
	}

	out := &runner.Result{Command: method, IsGRPC: true, GRPCStatus: int(stat.Code())}
	if stat.Code() == codes.OK {
		b, err := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}.Marshal(res)
		if err != nil {
			return nil, fmt.Errorf("encoding grpc response for %s: %w", method, err)
		}
		out.MessageJSON = b
	} else {
		out.MessageJSON = []byte("{}")
	}
	return out, nil
}

// resolveMethod fetches the method descriptor for service/method via server
// reflection.
func (r *Runner) resolveMethod(ctx context.Context, service, method string) (protoreflect.MethodDescriptor, error) {
	refClient := grpcreflect.NewClientAuto(ctx, r.cc)
	// Reset CloseSends and drains the reflection stream and cancels its context;
	// without it every grpc step leaks an open reflection stream and its goroutine
	// for the lifetime of the scenario connection (issue #29).
	defer refClient.Reset()
	fd, err := refClient.FileContainingSymbol(protoreflect.FullName(service))
	if err != nil {
		return nil, fmt.Errorf("resolving grpc service %q via reflection (is server reflection enabled?): %w", service, err)
	}
	svcs := fd.Services()
	for i := 0; i < svcs.Len(); i++ {
		sd := svcs.Get(i)
		if string(sd.FullName()) != service {
			continue
		}
		md := sd.Methods().ByName(protoreflect.Name(method))
		if md == nil {
			return nil, fmt.Errorf("grpc method %q not found in service %q", method, service)
		}
		if md.IsStreamingClient() || md.IsStreamingServer() {
			return nil, fmt.Errorf("grpc method %q is streaming; only unary calls are supported", method)
		}
		return md, nil
	}
	return nil, fmt.Errorf("grpc service %q not found in the reflected schema", service)
}

// splitMethod parses "pkg.Service/Method" into its service and method parts.
func splitMethod(method string) (string, string, error) {
	i := strings.LastIndex(method, "/")
	if i <= 0 || i == len(method)-1 {
		return "", "", fmt.Errorf("grpc method %q must be in the form pkg.Service/Method", method)
	}
	return method[:i], method[i+1:], nil
}
