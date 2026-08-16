package diag

// ATG6xxx — security policy violations, which exit 6. These are operations
// atago refused on policy grounds rather than failures of the thing under
// test, and they take precedence over the execution-error code so a run cannot
// report a plain failure when a declared restriction was breached.
//
// The family is deliberately small. A refusal that the reader is likely to
// believe is atago being wrong needs room to explain the reasoning and to say
// how to do the thing legitimately, and a code is what gives it that room —
// one line at the terminal, the rest on a page it can be looked up on.
//
// Refusals that happen while loading a spec are reported as spec errors
// instead, since nothing has run: a path that leaves the workdir is ATG2310.
// Refusals that happen mid-run without a declared policy behind them —
// resolving a path through a symlink out of the scenario root, for instance —
// are execution errors, because they are atago protecting the isolation every
// scenario depends on rather than enforcing something the spec asked for.
const securitySince = "v0.21.0"

// NetworkPolicyDenied is a request to a host outside the declared allowlist.
var NetworkPolicyDenied = register(6001, "NetworkPolicyDenied", Entry{
	Summary: "a request targeted a host the spec's network policy does not allow",
	Detail:  "The spec declares `permissions.network.allow`, and this request went somewhere else. The check covers HTTP, gRPC, and SSH alike, and an HTTP redirect is re-checked against the same list before it is followed — otherwise an allowed host could bounce a request to a denied one and the restriction would mean nothing. Comparison is exact on the host, or on host and port when the entry names one; there is no wildcard.",
	Fix:     "Add the host to `permissions.network.allow` if the scenario is meant to reach it, or point the runner at a mock server or a scenario service instead. Removing the allowlist entirely turns the policy off, which is the default when no `permissions.network` block is declared.",
	Since:   securitySince,
})
