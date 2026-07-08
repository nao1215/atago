// Package ssh implements the SSH runner: a `run` step naming an ssh runner
// executes its command on a remote host over SSH, capturing stdout, stderr, and
// the exit code as a runner.Result. It is the atago
// counterpart to runn's SSH runner and builds on golang.org/x/crypto/ssh — the
// same transport runn uses — driven entirely by the runner's declared
// connection fields (no ~/.ssh/config dependency, so runs stay reproducible).
package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/nao1215/atago/internal/runner"
)

// Config is the resolved configuration for an ssh runner.
type Config struct {
	// Addr is the remote endpoint; a bare host gets the default port 22.
	Addr string
	// User is the login user (required).
	User string
	// Password, when set, enables password authentication.
	Password string
	// KeyFile, when set, is the path to a private key for public-key auth.
	KeyFile string
	// KnownHosts, when set, is a known_hosts file used to verify the host key.
	KnownHosts string
	// InsecureHostKey must be set to connect with an empty KnownHosts (host-key
	// verification disabled, for test/lab infrastructure). Without it an empty
	// KnownHosts is a configuration error, so the insecure mode is opt-in rather
	// than a silent MITM-able default (issue #17).
	InsecureHostKey bool
	// Timeout bounds connection establishment and each command; zero means none.
	Timeout time.Duration
}

// Runner holds a live SSH connection for one ssh runner.
type Runner struct {
	client  *ssh.Client
	timeout time.Duration
}

// Open dials the host and authenticates, returning a connected runner.
func Open(cfg Config) (*Runner, error) {
	if cfg.User == "" {
		return nil, errors.New("ssh runner requires a user")
	}
	auth, err := authMethods(cfg)
	if err != nil {
		return nil, err
	}
	hostKey, err := hostKeyCallback(cfg.KnownHosts, cfg.InsecureHostKey)
	if err != nil {
		return nil, err
	}
	clientCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            auth,
		HostKeyCallback: hostKey,
		Timeout:         cfg.Timeout,
	}
	client, err := ssh.Dial("tcp", withDefaultPort(cfg.Addr), clientCfg)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", cfg.Addr, err)
	}
	return &Runner{client: client, timeout: cfg.Timeout}, nil
}

// Close ends the SSH connection.
func (r *Runner) Close() error { return r.client.Close() }

// Run executes command in a fresh session and captures the result. A non-zero
// remote exit status is a successful Run with Result.ExitCode set (mirroring the
// cmd runner); only a transport/session failure returns an error.
func (r *Runner) Run(ctx context.Context, command string) (*runner.Result, error) {
	sess, err := r.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh session: %w", err)
	}
	defer func() { _ = sess.Close() }()

	// A caller-supplied deadline (the engine applies the step's own
	// run.timeout that way) takes precedence; without one, the runner-level
	// timeout captured at dial bounds the command. Track which level armed the
	// deadline so a fired timeout can name its knob, mirroring the cmd runner.
	timeoutSource := ""
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && r.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.timeout)
		defer cancel()
		timeoutSource = "runner.timeout"
	}

	var stdout, stderr strings.Builder
	sess.Stdout = &stdout
	sess.Stderr = &stderr

	done := make(chan error, 1)
	start := time.Now()
	go func() { done <- sess.Run(command) }()

	var runErr error
	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		_ = sess.Close()
		// Wait (bounded) for the session goroutine so the output builders are
		// quiescent before they are read — a remote that ignores the kill must
		// not turn into a data race or a hang here.
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
		// A fired deadline is an OBSERVABLE outcome assertions can inspect —
		// TimedOut with the captured output so far — matching the cmd runner
		// (#17). Only a cancellation (Ctrl-C / suite teardown) is an error.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return &runner.Result{
				Command:       command,
				Stdout:        []byte(stdout.String()),
				Stderr:        []byte(stderr.String()),
				Duration:      time.Since(start),
				ExitCode:      -1,
				TimedOut:      true,
				TimeoutSource: timeoutSource,
			}, nil
		}
		return nil, ctx.Err()
	case runErr = <-done:
	}
	elapsed := time.Since(start)

	res := &runner.Result{
		Command:  command,
		Stdout:   []byte(stdout.String()),
		Stderr:   []byte(stderr.String()),
		Duration: elapsed,
	}
	var exitErr *ssh.ExitError
	switch {
	case runErr == nil:
		res.ExitCode = 0
	case errors.As(runErr, &exitErr):
		res.ExitCode = exitErr.ExitStatus()
	default:
		return nil, fmt.Errorf("ssh run %q: %w", command, runErr)
	}
	return res, nil
}

func authMethods(cfg Config) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if cfg.KeyFile != "" {
		key, err := os.ReadFile(cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("reading ssh key %q: %w", cfg.KeyFile, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parsing ssh key %q: %w", cfg.KeyFile, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}
	if len(methods) == 0 {
		return nil, errors.New("ssh runner requires a password or key_file")
	}
	return methods, nil
}

func hostKeyCallback(knownHosts string, insecure bool) (ssh.HostKeyCallback, error) {
	if knownHosts == "" {
		if !insecure {
			return nil, errors.New("ssh runner requires known_hosts to verify the host key; set insecure_host_key: true to connect without verification (test/lab only)")
		}
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // opt-in via insecure_host_key: disables checking for test infra
	}
	cb, err := knownhosts.New(knownHosts)
	if err != nil {
		return nil, fmt.Errorf("reading known_hosts %q: %w", knownHosts, err)
	}
	return cb, nil
}

// withDefaultPort appends the default SSH port 22 when addr carries none. It
// must handle a bare IPv6 literal without mangling it: net.JoinHostPort on an
// already-bracketed "[::1]" would double-bracket it ("[[::1]]:22"), and a
// trailing-colon "host:" parses as a present-but-empty port that must still get
// the default.
func withDefaultPort(addr string) string {
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if port != "" {
			return addr
		}
		return net.JoinHostPort(host, "22") // "host:" → "host:22"
	}
	// No port present. Strip surrounding brackets from a bracketed IPv6 literal
	// so JoinHostPort re-wraps it exactly once.
	host := addr
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	return net.JoinHostPort(host, "22")
}
