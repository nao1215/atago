package spec

import "strings"

// Signal targets a declared service (scenario or suite) by name and delivers
// a POSIX signal to its whole process group (#23), consistent with the
// teardown kill semantics. Wait optionally blocks until the process exits.
type Signal struct {
	// Service names the target: a service declared in the scenario's
	// services: list or started by a suite.setup service: step. Unknown names
	// are a load-time error listing the declared services.
	Service string `yaml:"service"`
	// Signal is the POSIX signal name: TERM, INT, HUP, USR1, USR2, or KILL
	// (an optional SIG prefix is accepted).
	Signal string `yaml:"signal"`
	// Wait, when set, blocks until the signaled process exits or the timeout
	// elapses; a still-running process fails the step with a clear message.
	Wait *SignalWait `yaml:"wait,omitempty"`
}

// SignalWait bounds the wait for a signaled service's exit (#23).
type SignalWait struct {
	// Timeout is a Go duration (default "5s").
	Timeout string `yaml:"timeout,omitempty"`
}

// validSignalNames is the accepted `signal:` set (#23): the portable
// process-control signals. Anything more exotic (STOP/CONT/real-time) is out
// of scope for declarative CLI testing.
var validSignalNames = map[string]bool{
	"TERM": true, "INT": true, "HUP": true, "USR1": true, "USR2": true, "KILL": true,
}

// NormalizeSignalName upper-cases a signal name and strips an optional SIG
// prefix, so `term`, `TERM`, and `SIGTERM` all mean SIGTERM (#23).
func NormalizeSignalName(name string) string {
	return strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(name)), "SIG")
}

// ValidSignalName reports whether the (normalized) signal name is accepted.
func ValidSignalName(name string) bool { return validSignalNames[NormalizeSignalName(name)] }
