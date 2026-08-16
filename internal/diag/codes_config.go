package diag

// ATG3xxx — configuration errors, which exit 3. These are raised before any
// spec is read: the command line itself, the options on it, or the files and
// scenarios it selected. Nothing has been loaded or executed when one of these
// is reported.
//
// The sub-ranges group by what the reader has to do about it:
//
//	ATG30xx  the shape of the command — which subcommand, and what it takes
//	ATG31xx  options — unknown, wrong value, wrong combination, out of range
//	ATG32xx  what the command was pointed at — targets, selections, outputs
//
// The distinction from ATG2xxx is who wrote the mistake and where: a spec error
// is in a file under version control that a team reviews, while a configuration
// error is in the invocation, usually in a CI workflow or a shell history.
const configSince = "v0.21.0"

// ATG30xx — the shape of the command.
var (
	// UnknownCommand is a subcommand atago does not have.
	UnknownCommand = register(3001, "UnknownCommand", Entry{
		Summary: "there is no such atago subcommand",
		Detail:  "The first argument names a subcommand, and this one is not in the inventory. Running atago with no arguments at all reports the same thing, since there is no default subcommand — a bare `atago` that started running specs would be a surprising thing for a tool that writes files.",
		Fix:     "Run `atago help` for the list of subcommands.",
		Since:   configSince,
	})

	// BadUsage is a subcommand called in a shape it does not accept.
	BadUsage = register(3002, "BadUsage", Entry{
		Summary: "a subcommand was called in a shape it does not accept",
		Detail:  "The subcommand exists but its arguments are not what it takes: `atago snapshot` without `update`, `atago completion` without a shell to generate for, `atago record` with no command after the `--`, or `atago init` given several paths when it writes one file. This is about the arguments' shape rather than their values, which is why it is separate from an option given a value it does not accept.",
		Fix:     "Run the subcommand with `--help` for its usage line.",
		Since:   configSince,
	})
)

// ATG31xx — options.
var (
	// UnknownOption is a flag atago does not define.
	UnknownOption = register(3101, "UnknownOption", Entry{
		Summary: "the command line sets an option atago does not define",
		Detail:  "The option is not one this subcommand takes. Options are per-subcommand, so a flag that works for `atago run` is not necessarily accepted by `atago doc`, and a flag from a newer atago than the one running produces this too.",
		Fix:     "Check the option against the usage printed below the message, or run the subcommand with `--help`.",
		Since:   configSince,
	})

	// BadOptionValue is a value outside what an option accepts.
	BadOptionValue = register(3102, "BadOptionValue", Entry{
		Summary: "an option was given a value it does not accept",
		Detail:  "The option exists and takes a value from a fixed vocabulary, and this value is not in it — a report format, a shell to generate completion for, a scaffolding template. atago refuses rather than falling back to a default, because a misspelled `--report jnit` that quietly produced console output would leave a CI job collecting a report file that was never written.",
		Fix:     "Use one of the values the message lists.",
		Since:   configSince,
	})

	// OptionNeedsAnother is an option that only works alongside another.
	OptionNeedsAnother = register(3103, "OptionNeedsAnother", Entry{
		Summary: "an option needs another option that is not set",
		Detail:  "Some options only mean something in company: `--split-by-spec` writes one file per spec and needs the `--out-dir` to write them into, and `--snapshot` on `record` writes a golden next to the spec, which needs `--out` to say where the spec is going.",
		Fix:     "Add the option the message names, or drop the one that depends on it.",
		Since:   configSince,
	})

	// OptionsExclusive is a pair of options that cannot both apply.
	OptionsExclusive = register(3104, "OptionsExclusive", Entry{
		Summary: "two options that contradict each other are both set",
		Detail:  "The two options ask for incompatible things. `--repeat` and `--retry-failed` are the clearest pair: one re-runs scenarios to find out whether they flake, the other re-runs them so that flakiness does not fail the build, and a run cannot both detect and tolerate the same instability. `--out` and `--split-by-spec` disagree about whether the output is one file or many.",
		Fix:     "Keep the option that expresses what you want and drop the other.",
		Since:   configSince,
	})

	// OptionOutOfRange is a numeric option outside what it accepts.
	OptionOutOfRange = register(3105, "OptionOutOfRange", Entry{
		Summary: "a numeric option is outside the range it accepts",
		Detail:  "Counts are never negative. A negative `--parallel` would otherwise be clamped to sequential and exit 0, so the typo would run the suite in a way nobody asked for and report success; the same applies to `--repeat` and `--retry-failed`.",
		Fix:     "Use zero or a positive number. Zero means atago's default for these options rather than none.",
		Since:   configSince,
	})
)

// ATG32xx — what the command was pointed at.
var (
	// TargetNotFound is a path that cannot be reached.
	TargetNotFound = register(3201, "TargetNotFound", Entry{
		Summary: "a path on the command line cannot be reached",
		Detail:  "atago was pointed at a file or directory that does not exist, or that it is not allowed to look at. In CI this is usually a working directory that is not what the workflow assumed, or a checkout step that has not run yet.",
		Fix:     "Check the path as spelled, and the directory the command runs from.",
		Since:   configSince,
	})

	// NoSpecFiles is a search that found nothing to run.
	NoSpecFiles = register(3202, "NoSpecFiles", Entry{
		Summary: "no spec files were found under the paths given",
		Detail:  "The paths exist but hold no `*.atago.yaml` or `*.atago.yml` files. Directories are searched recursively, so this means there genuinely are none below them. Reporting success here would let a CI job that lost its specs go green forever.",
		Fix:     "Point atago at the directory holding the specs, or run `atago init` to scaffold one. The file name matters: a spec must end in `.atago.yaml` or `.atago.yml`.",
		Since:   configSince,
	})

	// EmptySelection is a filter that matched nothing under --ci.
	EmptySelection = register(3203, "EmptySelection", Entry{
		Summary: "the scenario selection matched nothing, and --ci refuses to call that success",
		Detail:  "Specs were found and loaded, but `--filter`, `--tag`, or `--skip-tag` selected none of their scenarios. Outside `--ci` this is a warning; under `--ci` it fails, because a selection that silently stops matching is how a suite disables itself without anyone noticing — a renamed tag keeps the build green while testing nothing. The two selectors match differently: `--filter` is a case-sensitive substring of the name, while `--tag` and `--skip-tag` compare tags for exact equality.",
		Fix:     "Run `atago list` to see the scenario names and tags actually present, and correct the selector.",
		Since:   configSince,
	})

	// RerunNothingMatched is a --rerun-failed run whose ledger no longer maps.
	RerunNothingMatched = register(3204, "RerunNothingMatched", Entry{
		Summary: "no recorded failing scenario matched the current specs",
		Detail:  "`--rerun-failed` re-runs what failed last time, and none of the recorded names exist any more — they were renamed or removed since the run that recorded them. The recorded failures are kept rather than cleared, because the work they represent is still unverified, and exiting 0 here would report a green run that tested nothing.",
		Fix:     "Run the full suite once to record failures against the current names, or drop the recorded state and start again.",
		Since:   configSince,
	})

	// OutputExists is a write that would destroy an existing file.
	OutputExists = register(3205, "OutputExists", Entry{
		Summary: "the output file already exists",
		Detail:  "`atago init` and `atago record` write a new spec, and neither overwrites one by default. Silently replacing a spec someone has edited is the kind of loss a tool should never cause without being asked.",
		Fix:     "Choose another path, or pass `--force` to overwrite the existing file deliberately.",
		Since:   configSince,
	})

	// OutputNotWritable is a destination that cannot be written.
	OutputNotWritable = register(3206, "OutputNotWritable", Entry{
		Summary: "a destination atago was asked to write cannot be written",
		Detail:  "The generated documentation, or the artifacts directory a run was told to use, could not be created or written. The artifacts directory is checked before the run starts rather than at the first failure: a directory that cannot be written would otherwise make every artifact write fail quietly, leaving a run that looks like it produced nothing to review when in fact nothing could be saved.",
		Fix:     "Check that the parent directory exists and is writable, and that nothing else already occupies the path as a file.",
		Since:   configSince,
	})

	// StateUnreadable is atago's own run-to-run state failing to load.
	StateUnreadable = register(3207, "StateUnreadable", Entry{
		Summary: "atago's recorded state from a previous run cannot be read",
		Detail:  "`--rerun-failed` reads a small ledger written by the previous run, and that file could not be read or decoded. A truncated or hand-edited ledger is the usual cause.",
		Fix:     "Delete the ledger and run the full suite once to record it again.",
		Since:   configSince,
	})
)
