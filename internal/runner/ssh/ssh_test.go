package ssh

import "testing"

func TestWithDefaultPort(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"example.com":      "example.com:22",
		"example.com:2222": "example.com:2222",
		"127.0.0.1":        "127.0.0.1:22",
		"127.0.0.1:22":     "127.0.0.1:22",
		// IPv6 literals must not be double-bracketed, and a bracketed literal
		// without a port must gain one.
		"::1":        "[::1]:22",
		"[::1]":      "[::1]:22",
		"[::1]:2222": "[::1]:2222",
		"[fe80::1]":  "[fe80::1]:22",
		// A trailing-colon host has an empty port that must still get the default.
		"host:": "host:22",
	}
	for in, want := range cases {
		if got := withDefaultPort(in); got != want {
			t.Errorf("withDefaultPort(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAuthMethods(t *testing.T) {
	t.Parallel()
	if _, err := authMethods(Config{}); err == nil {
		t.Error("authMethods with no password or key should error")
	}
	m, err := authMethods(Config{Password: "x"})
	if err != nil || len(m) != 1 {
		t.Errorf("password auth: methods = %d, err = %v", len(m), err)
	}
	if _, err := authMethods(Config{KeyFile: "/no/such/key"}); err == nil {
		t.Error("missing key file should error")
	}
}

func TestOpen_RequiresUser(t *testing.T) {
	t.Parallel()
	if _, err := Open(Config{Addr: "127.0.0.1:22", Password: "x"}); err == nil {
		t.Error("Open without a user should error")
	}
}

func TestHostKeyCallback(t *testing.T) {
	t.Parallel()
	// Issue #17: an empty known_hosts is now a configuration error unless the
	// insecure host-key mode is explicitly opted into.
	if _, err := hostKeyCallback("", false); err == nil {
		t.Error("empty known_hosts without insecure_host_key should error")
	}
	if _, err := hostKeyCallback("", true); err != nil {
		t.Errorf("empty known_hosts with insecure_host_key should be allowed, got %v", err)
	}
	if _, err := hostKeyCallback("/no/such/known_hosts", false); err == nil {
		t.Error("missing known_hosts file should error")
	}
}
