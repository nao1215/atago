package diag

// ATG4xxx — execution errors, which exit 4. A scenario started and something
// stopped it from carrying out a step. These are distinct from a scenario
// failing: a failure means the assertions ran and disagreed with what happened,
// while an execution error means atago never got far enough to have an opinion.
//
// The sub-ranges group by what the reader has to do about it:
//
//	ATG40xx  starting the subject — the command, its arguments, its sandbox
//	ATG41xx  time — something took longer than its bound, or was interrupted
//	ATG42xx  reaching a peer — connecting, authenticating, addressing
//	ATG43xx  services and mock servers a scenario brought up
//	ATG44xx  session mechanics — terminals, signals, capture
//	ATG45xx  the data a step depends on — stored values, payloads, encodings
//
// Codes are assigned to the site that names the problem, not to every site an
// error passes through. "connection refused" from an HTTP runner and from a
// database runner share a code because the reader's next move is the same;
// a step that timed out and a service that never became ready do not, because
// one is fixed in `timeout:` and the other in `ready:`.
const execSince = "v0.21.0"

// ATG40xx — starting the subject.
var (
	// CommandUnparsable is a command line atago cannot split into arguments.
	CommandUnparsable = register(4001, "CommandUnparsable", Entry{
		Summary: "a command line could not be split into arguments",
		Detail:  "Without `shell: true`, atago splits the command itself using shell quoting rules rather than handing it to a shell, so that a spec means the same thing on every platform. An unbalanced quote leaves that split with no sensible answer. An empty command has the same problem for a different reason: there is nothing to run.",
		Fix:     "Balance the quotes, or set `shell: true` on the step when the command genuinely needs a shell to interpret it (pipes, redirection, globbing).",
		Since:   execSince,
	})

	// CommandNotStarted is a program that never began running.
	CommandNotStarted = register(4002, "CommandNotStarted", Entry{
		Summary: "the program could not be started",
		Detail:  "The operating system refused to launch it. Almost always the executable is not on `PATH`, or is present but not marked executable. In CI this usually means the build step has not run, or the binary landed somewhere the run does not look.",
		Fix:     "Check the program name and that it is on `PATH` for the run. A directory manifest's `subject:` block builds the binary under test before any scenario and puts it on `PATH`, which is the supported way to test something not installed system-wide.",
		Since:   execSince,
	})

	// StepFileUnusable is a file the step depends on that cannot be used.
	StepFileUnusable = register(4004, "StepFileUnusable", Entry{
		Summary: "a file the step depends on could not be read or written",
		Detail:  "The step named a file it had to read — a request body, an upload, an SSH key, a known_hosts list, a readiness probe's target — or one it had to write, such as a screenshot or a download. The filesystem refused. A path relative to the scenario workdir that was never created is the usual cause, since each scenario starts in a fresh directory.",
		Fix:     "Check the path, and check that whatever creates the file runs before the step that reads it. A `fixture:` step is the supported way to put a file into the scenario workdir.",
		Since:   execSince,
	})

	// SessionExecFailed is a helper command inside a session that failed.
	SessionExecFailed = register(4005, "SessionExecFailed", Entry{
		Summary: "a command run alongside a terminal session failed",
		Detail:  "A `pty:` session's `exec:` runs a command beside the terminal, usually to cause the change the session is waiting to observe. It exited non-zero, so that change was not made and whatever the session expects next can never arrive.",
		Fix:     "Run the command by hand to see why it fails. Its own output accompanies this message.",
		Since:   execSince,
	})

	// SandboxSetupFailed is the scenario's own environment failing to build.
	SandboxSetupFailed = register(4003, "SandboxSetupFailed", Entry{
		Summary: "the isolated environment for the step could not be prepared",
		Detail:  "Before a step runs, atago builds what it runs inside: the scenario workdir, a suite scratch directory, a sandboxed `HOME` when one was asked for, and any stdin the step feeds in. One of those could not be created or read. A full or read-only temp filesystem is the usual cause.",
		Fix:     "Check free space and permissions on the temp directory, and check any file the step names for its stdin.",
		Since:   execSince,
	})
)

// ATG41xx — time.
var (
	// StepTimeout is a step that outlasted its bound.
	StepTimeout = register(4101, "StepTimeout", Entry{
		Summary: "a step did not finish within its timeout",
		Detail:  "The step was still running when its bound elapsed, so atago stopped it. Every step has a timeout, from the step's own `timeout:`, the suite's, or atago's default, because a test that hangs forever is worse than one that fails — CI notices the failure and never notices the hang.",
		Fix:     "Raise `timeout:` if the work legitimately takes that long, or find out what the program is waiting for. A program waiting on input nobody sends is the usual cause: give the step a `stdin:`, or drive it as a `pty:` session.",
		Since:   execSince,
	})

	// ReadinessTimeout is a service that never came up.
	ReadinessTimeout = register(4102, "ReadinessTimeout", Entry{
		Summary: "a service did not become ready within its timeout",
		Detail:  "The service was started and its readiness probe never succeeded before the bound elapsed. atago waits rather than racing, so that a scenario cannot pass or fail depending on how fast a machine happens to be. The service's own output is included with this message, because what it printed while failing to start is usually the whole answer.",
		Fix:     "Read the service output above the message first. If the service is simply slow, raise `ready.timeout`; if the probe is wrong, check that `ready.port`, `ready.file`, or `ready.log` describes what the service actually does when it is up.",
		Since:   execSince,
	})

	// RunInterrupted is a run stopped from outside.
	RunInterrupted = register(4103, "RunInterrupted", Entry{
		Summary: "the run was interrupted before the step finished",
		Detail:  "A `Ctrl-C` or `SIGTERM` reached atago, which stops scheduling new work and unwinds what is in flight: processes are stopped, services torn down, sessions closed, and partial results reported. The step that was running when the signal arrived reports this rather than a verdict, because it never reached one.",
		Fix:     "Nothing, when the interruption was deliberate. In CI this usually means the job hit its own overall time limit, which is worth checking before the specs are.",
		Since:   execSince,
	})
)

// ATG42xx — reaching a peer.
var (
	// ConnectFailed is a peer that could not be reached.
	ConnectFailed = register(4201, "ConnectFailed", Entry{
		Summary: "a peer could not be reached",
		Detail:  "An HTTP, database, SSH, gRPC, or browser connection failed to establish. The same code covers all of them because the reader's next move is the same: find out whether the thing being connected to is actually up and actually at that address. Connection refused means nothing is listening there; a timeout usually means a firewall or a wrong host.",
		Fix:     "Check that the peer is running and that the runner's address matches. Where the peer is part of the test, declare it as a scenario `service:` so atago starts it and waits for readiness before the step runs.",
		Since:   execSince,
	})

	// RunnerConfigIncomplete is a runner missing what it needs to connect.
	RunnerConfigIncomplete = register(4202, "RunnerConfigIncomplete", Entry{
		Summary: "a runner is missing something it needs before it can connect",
		Detail:  "The runner declaration is short of a required piece: an SSH runner with no user or no credential, a database runner with no DSN, a gRPC runner with no target. SSH host-key verification belongs here too — a runner with no `known_hosts` is refused rather than connecting blind, since silently accepting any host key would make the check worthless.",
		Fix:     "Complete the `runners:` entry with what the message names. For SSH, either point `known_hosts` at a real file or set `insecure_host_key: true` deliberately, for a throwaway host you control.",
		Since:   execSince,
	})

	// RemoteRejected is a peer that answered by refusing.
	RemoteRejected = register(4203, "RemoteRejected", Entry{
		Summary: "the peer was reached but rejected what was asked of it",
		Detail:  "The connection worked and the request did not: a gRPC service or method that does not exist in the reflected schema, a method that streams where only unary calls are supported, a redirect chain that never settled. This is the peer disagreeing about what it offers, rather than a failure to reach it.",
		Fix:     "Check the name against what the peer actually exposes. For gRPC, the message lists what reflection reported, which is the authoritative answer for the server actually running.",
		Since:   execSince,
	})

	// BadEndpoint is an address atago cannot make sense of.
	BadEndpoint = register(4204, "BadEndpoint", Entry{
		Summary: "an address or connection string could not be understood",
		Detail:  "A URL, DSN, or method name is malformed, or a driver could not be inferred from a DSN whose scheme atago does not recognize. A relative request path with no `base_url` on the runner is the same problem seen from the other side: there is not enough there to form an address.",
		Fix:     "Correct the address to the form the message names. A database runner can also state its `driver:` explicitly rather than leaving it to be inferred.",
		Since:   execSince,
	})
)

// ATG43xx — services and mock servers.
var (
	// ServiceNotReady is a service that failed rather than started.
	ServiceNotReady = register(4301, "ServiceNotReady", Entry{
		Summary: "a service failed before it became ready",
		Detail:  "The service exited, or its readiness probe reported it unusable, before any step could depend on it. Unlike a readiness timeout this is not about waiting: the service made its answer clear early. Its captured output accompanies the message.",
		Fix:     "Read the service's own output first — a bad configuration file or an occupied port is usually visible there.",
		Since:   execSince,
	})

	// ServiceNotRunning is an operation on a service that is not up.
	ServiceNotRunning = register(4302, "ServiceNotRunning", Entry{
		Summary: "a service was addressed while it was not running",
		Detail:  "A step signaled or otherwise acted on a service that had already exited, or had not been started. A service that ignores a stop signal and outlives its grace period is reported here too, since what follows is the same question: what is that process doing.",
		Fix:     "Check whether the service exits on its own partway through the scenario, which its captured output usually shows. A service that will not stop on `TERM` may need `KILL`.",
		Since:   execSince,
	})

	// MockServerFailed is atago's own stub server failing.
	MockServerFailed = register(4303, "MockServerFailed", Entry{
		Summary: "a mock server could not be started or could not answer",
		Detail:  "The mock server atago runs on the scenario's behalf failed to bind a port, or could not build the response a route declares. Unlike the services above, this is atago's own process, so the cause is either the environment refusing the port or the route being impossible to render.",
		Fix:     "Check the route's declared response, and whether the port is already taken. Leaving the port unset lets atago choose a free one and expose it through the store.",
		Since:   execSince,
	})
)

// ATG44xx — session mechanics.
var (
	// PTYFailed is a pseudo-terminal that could not be set up or driven.
	PTYFailed = register(4401, "PTYFailed", Entry{
		Summary: "a pseudo-terminal could not be allocated or driven",
		Detail:  "A `pty:` step runs its program in a real terminal so that programs behaving differently when attached to one can be tested at all. Allocating that terminal, resizing it, or identifying it failed. A container with no `/dev/pts` mounted is the usual cause on Linux, and a terminal size outside what the kernel accepts is the usual cause everywhere.",
		Fix:     "Check that the environment provides pseudo-terminals, and that any `rows`/`cols` are within range. In a container, mounting `/dev/pts` is what makes `pty:` steps possible.",
		Since:   execSince,
	})

	// InputNotSupported is a key or signal atago does not know how to send.
	InputNotSupported = register(4402, "InputNotSupported", Entry{
		Summary: "atago does not know how to send that key or signal",
		Detail:  "Named keys and signals come from a fixed vocabulary, so that a spec transmits the same bytes everywhere rather than depending on what a terminal happens to map. The name given is not in it.",
		Fix:     "Use one of the names the message lists.",
		Since:   execSince,
	})

	// CaptureFailed is output that could not be collected.
	CaptureFailed = register(4403, "CaptureFailed", Entry{
		Summary: "the step's output could not be captured",
		Detail:  "The program ran but atago could not collect what it wrote. The usual cause is a program that leaves a child process holding the output pipe open after it exits: atago waits briefly and then reports this rather than blocking forever on a pipe nobody will close. Assertions on the output cannot be trusted when this happens, so the step errors instead of failing.",
		Fix:     "Check for a background process the command leaves behind. Redirecting the child's output inside the command, or waiting for it before exiting, closes the pipe.",
		Since:   execSince,
	})

	// BrowserActionFailed is a browser instruction that did not take effect.
	BrowserActionFailed = register(4405, "BrowserActionFailed", Entry{
		Summary: "a browser action could not be carried out",
		Detail:  "A `cdp:` action failed against the page: a selector matching nothing, a click that never landed, a download the browser would not begin. The browser was running and reachable — this is about the page rather than the connection.",
		Fix:     "Check the selector against the page as it is at that moment. A page still loading is the usual cause, and a `check:` action before the one that fails is how a scenario waits for it.",
		Since:   execSince,
	})

	// TerminalModeMismatch is input the program is not set up to receive.
	TerminalModeMismatch = register(4406, "TerminalModeMismatch", Entry{
		Summary: "the session sent input the program is not set up to receive",
		Detail:  "Mouse reporting and bracketed paste are modes a program turns on for itself, and a program that has not turned one on receives the bytes as ordinary keystrokes rather than as the event the session meant. atago refuses instead of sending them, because the result would be a scenario asserting against input the program interpreted as something else entirely.",
		Fix:     "Wait for the program to enable the mode before sending: an `expect_screen:` on whatever it draws once it is ready is usually enough. A program that never enables it cannot be driven this way.",
		Since:   execSince,
	})

	// UnsupportedOnPlatform is a step the running platform cannot perform.
	UnsupportedOnPlatform = register(4404, "UnsupportedOnPlatform", Entry{
		Summary: "the step is not supported on this platform",
		Detail:  "Some steps rest on facilities one platform has and another does not — POSIX signals being the clearest case. atago refuses rather than pretending, because a signal quietly not delivered would leave the scenario asserting against a program that never received it.",
		Fix:     "Gate the scenario with `skip: {os: ...}` so it runs where the facility exists, or express the same intent with something portable.",
		Since:   execSince,
	})
)

// ATG45xx — the data a step depends on.
var (
	// StoreSourceMissing is a capture with nothing behind it.
	StoreSourceMissing = register(4501, "StoreSourceMissing", Entry{
		Summary: "a store step has no result to capture from",
		Detail:  "`store:` captures a value out of the step before it — a command's stdout, a response body, a query's rows — and no such step has run in this scenario yet. The usual cause is a store placed before the step it means to read, or after a step of a different kind.",
		Fix:     "Move the store step directly after the step whose result it captures.",
		Since:   execSince,
	})

	// StoreExtractFailed is a capture that found nothing to store.
	StoreExtractFailed = register(4502, "StoreExtractFailed", Entry{
		Summary: "the value to store could not be extracted",
		Detail:  "The source step ran, and the selector found nothing usable in it: a JSON path matching no value or several, a regexp that did not match, a body that is not the JSON it was read as. atago errors rather than storing an empty value, because an empty value would flow silently into every later step that expands it.",
		Fix:     "Check the selector against what the step actually produced — `--verbose` prints each step's captured output, which is the quickest way to see it.",
		Since:   execSince,
	})

	// PayloadFailed is a request body that could not be built or read.
	PayloadFailed = register(4503, "PayloadFailed", Entry{
		Summary: "a request payload could not be built",
		Detail:  "The step's body could not be encoded or assembled: a JSON value that will not marshal, a multipart form that could not be written, a gRPC message that does not fit the schema the peer reflected. Nothing was sent, so the peer never saw the request.",
		Fix:     "Check the payload against the shape the peer expects. For gRPC, the reflected schema is what decides, not the local proto files.",
		Since:   execSince,
	})

	// VariableUnresolved is an expansion with nothing behind it.
	VariableUnresolved = register(4505, "VariableUnresolved", Entry{
		Summary: "a variable the step expands is not defined",
		Detail:  "The step referenced `${name}` or `${env:NAME}` and neither a stored value nor an environment variable of that name exists at that point. atago refuses rather than expanding to an empty string, which would silently send an empty password or an empty path and fail somewhere far from the cause.",
		Fix:     "Set the environment variable, or add the `store:` step that defines the name before the step that reads it. `$${...}` writes a literal dollar-brace where no expansion is wanted.",
		Since:   execSince,
	})

	// ResponseUnreadable is a peer's answer that could not be decoded.
	ResponseUnreadable = register(4504, "ResponseUnreadable", Entry{
		Summary: "the peer answered with something that could not be read",
		Detail:  "A response arrived and could not be decoded — a body that is not the JSON it claims to be, a row that could not be scanned into a value, a gRPC message that will not unmarshal. The exchange succeeded at the transport level and failed at the content level.",
		Fix:     "Look at the raw response, which `--verbose` prints. An HTML error page where JSON was expected is the most common version of this, and it usually means the request reached something other than the intended endpoint.",
		Since:   execSince,
	})
)
