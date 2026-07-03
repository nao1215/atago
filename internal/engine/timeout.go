package engine

// DefaultStepTimeout is the built-in bound applied to run/http/query/grpc
// steps when no level configures a timeout (#17), so an unconfigured hanging
// command fails loudly instead of stalling the run (or a CI job) forever.
const DefaultStepTimeout = "60s"

// defaultTimeoutSource is the hint label for the built-in bound; it names the
// value so the failure message tells the user what budget they exceeded.
const defaultTimeoutSource = "built-in 60s default timeout"

// resolveTimeout walks the timeout precedence chain (#17) — step >
// runner-common > defaults.run > suite > built-in 60s — and returns the
// winning duration string plus a source label for the timeout-kill hint. An
// explicit "0"/"0s" at any level stops the walk and disables the bound (the
// documented escape hatch); an empty string means "unset, keep looking". All
// candidate strings are duration-validated at load time.
func resolveTimeout(step, runnerCommon, defaultsRun, suite string) (value, source string) {
	for _, c := range []struct{ v, src string }{
		{step, "run.timeout"},
		{runnerCommon, "runner.timeout"},
		{defaultsRun, "defaults.run.timeout"},
		{suite, "suite.timeout"},
	} {
		if c.v != "" {
			return c.v, c.src
		}
	}
	return DefaultStepTimeout, defaultTimeoutSource
}
