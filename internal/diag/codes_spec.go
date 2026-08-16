package diag

// ATG2xxx — spec errors, which exit 2. These are raised while reading a spec
// file: the document could not be parsed, or it parsed but does not describe a
// runnable suite. Nothing has been executed when one of these is reported.
//
// The sub-ranges group by what the reader has to do about it:
//
//	ATG20xx  the document itself — syntax, encoding, the schema's outer shape
//	ATG21xx  a key that should not be there, or a combination that contradicts
//	ATG22xx  a key that should be there and is not
//	ATG23xx  a value that is present but wrong
//	ATG24xx  a name that does not resolve to anything the spec declares
//	ATG25xx  identity — names that collide
//
// Numbers are left with gaps so a later code joins its neighbours. A retired
// code is never reused: the registry records the version each shipped in, and
// a number that once meant something else would make an old CI log lie.
const specSince = "v0.21.0"

// ATG20xx — the document itself.
var (
	// SpecUnreadable is the file never being opened.
	SpecUnreadable = register(2001, "SpecUnreadable", Entry{
		Summary: "the spec file could not be read",
		Detail:  "atago was given a path and the operating system refused to open it. The file may not exist, may be a directory, or may not be readable by the user running atago. Nothing was parsed, so no other diagnostic can be reported about this file.",
		Fix:     "Check the path as spelled on the command line, and check the file's permissions. `atago list <dir>` shows which spec files atago can actually see in a directory.",
		Since:   specSince,
	})

	// SpecEmpty is a file with nothing in it to run.
	SpecEmpty = register(2002, "SpecEmpty", Entry{
		Summary: "the spec file contains no YAML document",
		Detail:  "The file was read but holds nothing a spec can be built from — it is empty, or contains only whitespace and comments. This is most often a file that was created but never filled in, or one whose contents were lost by an editor or a bad merge.",
		Fix:     "A spec needs at least `version`, `suite`, and `scenarios`. Run `atago init` to write a runnable starter spec and edit that.",
		Since:   specSince,
	})

	// YAMLSyntax is the document not being YAML at all.
	YAMLSyntax = register(2003, "YAMLSyntax", Entry{
		Summary: "the file is not valid YAML",
		Detail:  "Parsing stopped before atago could look at what the document means. The reported line and column are where the parser gave up, which is usually at or just after the real mistake. Inconsistent indentation, a missing colon after a key, and an unclosed quote or bracket account for most of these.",
		Fix:     "Fix the syntax at the reported position. Indentation in YAML must use spaces, never tabs, and every level must be indented consistently.",
		Since:   specSince,
	})

	// YAMLTag is an explicit tag written into a spec.
	YAMLTag = register(2004, "YAMLTag", Entry{
		Summary: "an explicit YAML tag is not supported in a spec",
		Detail:  "The document carries an explicit tag such as `!!str` or `!!map`. atago's schema is closed and fully typed: every field's type already fixes how its value is read, so a tag can only restate what the schema says or contradict it. The one exception is `!!binary`, which `atago record` writes when a captured stream is not valid UTF-8.",
		Fix:     "Remove the tag. If a value needs to be forced to text, quote it instead.",
		Since:   specSince,
	})

	// UnknownKey is a key the schema does not define.
	UnknownKey = register(2005, "UnknownKey", Entry{
		Summary: "the spec sets a key the schema does not define",
		Detail:  "Specs are decoded strictly, so a key atago does not know is an error rather than something quietly ignored — a silently dropped `assert:` block is a test that passes without testing anything. The usual causes are a typo, a key written at the wrong nesting level, and a key from a newer atago than the one running.",
		Fix:     "Check the spelling and the indentation against the message's suggestion. The JSON Schema under `schema/atago.schema.json` gives an editor live completion for every key.",
		Since:   specSince,
	})

	// WrongValueShape is a value written in a shape its key cannot take.
	WrongValueShape = register(2006, "WrongValueShape", Entry{
		Summary: "a value is written in a shape the key does not accept",
		Detail:  "The key exists but its value is the wrong kind of thing — a bare string where a mapping is wanted, a mapping where a list is wanted, or text where a number is wanted. The most common instance by far is a stream assertion written as a bare scalar: `stdout: hello` rather than `stdout: {contains: hello}`.",
		Fix:     "Write the value in the shape the key takes. Stream assertions (`stdout`, `stderr`, `body`, `rows`, `message`, `value`) always take a matcher mapping such as `{contains: \"...\"}` or `{equals: \"...\"}`.",
		Since:   specSince,
	})

	// SpecVersion is an unsupported spec format version.
	SpecVersion = register(2010, "SpecVersion", Entry{
		Summary: "the spec declares a format version atago does not support",
		Detail:  "The `version` key names the spec format, not the atago release, and `\"1\"` is the only format that exists. A version of `1` without quotes is read as a number and is also refused, because the field is a format name rather than a quantity.",
		Fix:     "Write `version: \"1\"`, with the quotes.",
		Since:   specSince,
	})
)

// ATG21xx — a key that should not be there, or a contradictory combination.
var (
	// StepNoAction is a step that does nothing.
	StepNoAction = register(2101, "StepNoAction", Entry{
		Summary: "a step sets no action",
		Detail:  "Every step does exactly one thing: run a command, make a request, drive a terminal, write a fixture, assert, or one of the other actions. A step with none of them is usually a block whose action key was lost to a bad indent, or a placeholder that was never filled in.",
		Fix:     "Give the step exactly one of `fixture`, `run`, `http`, `query`, `grpc`, `cdp`, `assert`, `store`, `pty`, or `signal` — or delete the step.",
		Since:   specSince,
	})

	// StepManyActions is a step that tries to do several things at once.
	StepManyActions = register(2102, "StepManyActions", Entry{
		Summary: "a step sets more than one action",
		Detail:  "A step is one action so that ordering, timing, and failure are unambiguous: a step that both ran a command and made a request has no single answer for which one a `duration` bounds or which one an assertion followed. The usual cause is a second action indented into the previous step instead of starting a new one.",
		Fix:     "Split the actions into separate list items under `steps:`, one `- ` per action.",
		Since:   specSince,
	})

	// ExclusiveKeys is a combination that contradicts itself.
	ExclusiveKeys = register(2103, "ExclusiveKeys", Entry{
		Summary: "keys that contradict each other are set together",
		Detail:  "Two keys were set whose meanings cannot both hold. A snapshot already pins every byte of what it captures, so pairing it with a size bound or with the `contains`/`count`/`glob` family asks the same question twice and invites the two answers to disagree. An exact `size` next to `min_size`/`max_size`, two lower bounds, and `exists: false` next to a size are the same mistake in other places.",
		Fix:     "Keep the key that expresses what you actually want to pin and drop the other. The message names both.",
		Since:   specSince,
	})

	// ChooseExactlyOne is a group where exactly one member is required.
	ChooseExactlyOne = register(2104, "ChooseExactlyOne", Entry{
		Summary: "a group of keys that takes exactly one member has none or several",
		Detail:  "Some blocks offer alternatives that are mutually exclusive by nature: a `pty` step is one of expect, send, expect_screen, resize, or exec; an `exit_code` is a bare number, a `not`, or an `in`. Setting several leaves the meaning undefined, and setting none leaves the block with nothing to do.",
		Fix:     "Set exactly one member of the group the message lists. To do two things, write two entries.",
		Since:   specSince,
	})

	// KeyNotHere is a real key used where it has no meaning.
	KeyNotHere = register(2105, "KeyNotHere", Entry{
		Summary: "a key is real but has no meaning in this position",
		Detail:  "The key exists elsewhere in the schema but not here, and accepting it silently would make it look effective when it is not. `defaults.run.command` is the clearest case: defaults describe how every step runs, while the command is what an individual step is, so a default command would be inherited by steps that already have one and quietly ignored. `trim` and `text` outside a store source, and `recursive` next to `snapshot`, are the same shape of mistake.",
		Fix:     "Move the key to where it applies, or drop it. The message says which of the two the key needs.",
		Since:   specSince,
	})

	// BlockNotHere is a whole block used at the wrong level.
	BlockNotHere = register(2106, "BlockNotHere", Entry{
		Summary: "a block is not allowed at this level of the spec",
		Detail:  "Some blocks are tied to a level by what they own. A `service` or `mock_server` step belongs in `suite.setup` because it outlives any one scenario; a scenario-scoped peer goes in that scenario's own `services:` or `mock_servers:` list instead. Others go the other way: steps needing a scenario workdir and runners cannot sit at suite level, and `assert.changes` is not available in teardown because the workdir delta is tracked only around a scenario's steps.",
		Fix:     "Move the block to the level the message names.",
		Since:   specSince,
	})

	// AssertNeedsStep is an assertion with nothing to assert about.
	AssertNeedsStep = register(2107, "AssertNeedsStep", Entry{
		Summary: "an assertion needs a preceding step it does not have",
		Detail:  "Some assertions describe the step before them rather than the scenario as a whole. `assert.screen` reads the terminal a `pty` step rendered, `assert.duration` bounds the wall-clock time of the step it follows, and `assert.changes` pins the workdir delta of the run or pty step it follows. Without that step there is nothing for the assertion to be about, so it would pass or fail for reasons unrelated to what it says.",
		Fix:     "Put the assertion directly after the step it describes. One assert block may set several keys at once, so an `exit_code`, a `stdout`, and a `changes` about the same step belong together.",
		Since:   specSince,
	})

	// KeyNeedsAnother is a key that only means something alongside another.
	KeyNeedsAnother = register(2108, "KeyNeedsAnother", Entry{
		Summary: "a key needs another key that is not set",
		Detail:  "The key is only meaningful in the presence of a second one: `max_diff` bounds the difference `similar_to` measures, `pass_env` selects which host variables survive a `clear_env`, and `count` counts the matches of a `contains`. Alone, the key has nothing to act on, so accepting it would leave a spec that looks like it constrains something and does not.",
		Fix:     "Add the key the message names, or drop the one that depends on it.",
		Since:   specSince,
	})
)

// ATG22xx — something required is missing.
var (
	// RequiredKey is a key the spec cannot do without.
	RequiredKey = register(2201, "RequiredKey", Entry{
		Summary: "a required key is missing",
		Detail:  "The block cannot be acted on without this key: a service with no command has nothing to start, an HTTP step with no method has no request to make, a scenario with no name cannot be selected, reported on, or documented. atago refuses these at load time rather than discovering them mid-run, so a broken spec fails before it starts changing anything.",
		Fix:     "Add the key the message names to the block it names.",
		Since:   specSince,
	})

	// EmptyList is a list that has to hold something and does not.
	EmptyList = register(2202, "EmptyList", Entry{
		Summary: "a list that must hold at least one entry is empty",
		Detail:  "A suite with no scenarios, a scenario with no steps, or a `cdp` block with no actions is accepted by YAML and means nothing to run. Left alone it would report success without having tested anything, which is worse than failing.",
		Fix:     "Add an entry to the list, or remove the block that holds it.",
		Since:   specSince,
	})

	// ChooseAtLeastOne is a group where at least one member is required.
	ChooseAtLeastOne = register(2203, "ChooseAtLeastOne", Entry{
		Summary: "a group of keys that needs at least one member has none",
		Detail:  "Unlike an exactly-one group, these accept any combination — but not the empty one, which would assert nothing. `assert.changes` needs at least one of `created`, `modified`, or `deleted`; a bounds block needs at least one of `lt`, `lte`, `gt`, or `gte`; a recursive directory assert needs at least one matcher.",
		Fix:     "Set one or more members of the group the message lists. To assert that a category changed nothing, write it explicitly as an empty list, e.g. `created: []`.",
		Since:   specSince,
	})

	// EmptyValue is a key written with nothing in it.
	EmptyValue = register(2204, "EmptyValue", Entry{
		Summary: "a value that must not be empty is empty",
		Detail:  "The key is present but carries nothing. This is worse than leaving the key out, because an empty value usually does not mean what it looks like: an empty ignore pattern excludes nothing, an empty name identifies nothing, an empty path names the workdir itself.",
		Fix:     "Give the key a value, or remove the key. Where removing it is the right answer the message says so.",
		Since:   specSince,
	})
)

// ATG23xx — a value that is present but wrong.
var (
	// BadDuration is text that does not parse as a duration.
	BadDuration = register(2301, "BadDuration", Entry{
		Summary: "a duration is not written in a form atago can read",
		Detail:  "Durations are written as a number with a unit — `500ms`, `2s`, `2m`, `1h30m` — and a bare number is refused because it does not say which unit was meant. A timeout that silently meant nanoseconds instead of seconds would look like a hang.",
		Fix:     "Write the value with a unit, e.g. `timeout: 2m`. Where a duration can be disabled, `\"0\"` does that.",
		Since:   specSince,
	})

	// NegativeValue is a quantity below zero that cannot be.
	NegativeValue = register(2302, "NegativeValue", Entry{
		Summary: "a value is negative where a negative has no meaning",
		Detail:  "Wall-clock bounds, sizes, counts, and page numbers are never below zero. A negative here is nearly always a sign slip or an arithmetic mistake in a generated spec, and treating it as zero would hide that.",
		Fix:     "Use zero or a positive value. Where the intent was to disable a bound rather than set it to nothing, the message names the way to do that.",
		Since:   specSince,
	})

	// NonPositiveValue is zero where zero has no meaning.
	NonPositiveValue = register(2303, "NonPositiveValue", Entry{
		Summary: "a value must be positive and is zero or below",
		Detail:  "Some values have no sensible zero: an interval that never elapses, a worker count of nobody, a retry that waits no time between attempts. Where atago has a working default, zero would be indistinguishable from not setting the key at all.",
		Fix:     "Set a positive value, or omit the key to take atago's default.",
		Since:   specSince,
	})

	// OutOfRange is a number outside what the field accepts.
	OutOfRange = register(2304, "OutOfRange", Entry{
		Summary: "a number is outside the range the key accepts",
		Detail:  "The value parsed but falls outside what the field can mean. Screen coordinates are 1-based cells, so row or column zero is not the first cell but a mistake, and it is the instance of this seen most often — a spec written against a 0-based mental model reads the wrong cell everywhere or nowhere.",
		Fix:     "Bring the value into the range the message states.",
		Since:   specSince,
	})

	// BadRegexp is a pattern that does not compile.
	BadRegexp = register(2305, "BadRegexp", Entry{
		Summary: "a regular expression does not compile",
		Detail:  "atago compiles every pattern in a spec at load time — `matches`, scrub patterns, readiness patterns — so a broken one is reported before the run rather than at the moment it would first have been used, halfway through a suite. The syntax is Go's RE2: no backreferences and no lookaround.",
		Fix:     "Correct the pattern at the position the message reports. Regex metacharacters that are meant literally need escaping, and a pattern inside a YAML double-quoted string needs its backslashes doubled — single quotes avoid that.",
		Since:   specSince,
	})

	// BadGlob is a path pattern that does not parse.
	BadGlob = register(2306, "BadGlob", Entry{
		Summary: "a glob pattern is not valid",
		Detail:  "Globs match paths by shape — `*` within a path segment, `**` across segments, `?` for one character, `[...]` for a class, `{a,b}` for alternatives. An unclosed bracket or brace is the usual cause.",
		Fix:     "Close the unbalanced bracket or brace, or escape it if it was meant literally.",
		Since:   specSince,
	})

	// NotAllowedValue is a value outside a closed set of choices.
	NotAllowedValue = register(2307, "NotAllowedValue", Entry{
		Summary: "a value is not one of the choices the key allows",
		Detail:  "The key takes a fixed vocabulary and the value is not in it. Platform names are the common case: `os` accepts `linux`, `darwin`, or `windows`, so `macos` or `osx` — reasonable words that Go's platform names do not use — are refused rather than quietly never matching.",
		Fix:     "Use one of the values the message lists.",
		Since:   specSince,
	})

	// EmptyInterval is a range that can never be satisfied.
	EmptyInterval = register(2308, "EmptyInterval", Entry{
		Summary: "the bounds of a range describe an interval nothing can fall in",
		Detail:  "A lower bound at or above the upper bound leaves no value that could satisfy both, so the assertion can only ever fail. This is almost always two bounds swapped, or a copied block whose second bound was not updated.",
		Fix:     "Order the bounds so the lower one is genuinely below the upper one, or use a single bound if only one side matters.",
		Since:   specSince,
	})

	// AbsolutePath is an absolute path where a relative one is required.
	AbsolutePath = register(2309, "AbsolutePath", Entry{
		Summary: "a path is absolute where it must be workdir-relative",
		Detail:  "Paths in a spec are relative to the scenario's own workdir, which is created fresh for each scenario. That is what lets the same spec run on another machine, in CI, and in parallel with itself. An absolute path names one machine's filesystem and breaks all three.",
		Fix:     "Write the path relative to the scenario workdir, e.g. `out/report.json` rather than `/home/you/project/out/report.json`.",
		Since:   specSince,
	})

	// PathEscapesWorkdir is a relative path pointing outside the workdir.
	PathEscapesWorkdir = register(2310, "PathEscapesWorkdir", Entry{
		Summary: "a path leads outside the scenario workdir",
		Detail:  "The path is relative but climbs out of the workdir with `../`. Scenario isolation is what makes a failing scenario reproducible and a parallel run safe, and a path that reaches outside it can read or overwrite another scenario's files, or the repository being tested. This is refused while loading; the equivalent refusal during a run is reported as a security violation and exits 6.",
		Fix:     "Keep the path inside the workdir. To work with a file that genuinely lives elsewhere, copy it in with a `fixture` step, which is the supported way to bring outside data into a scenario.",
		Since:   specSince,
	})

	// ControlCharacter is an unprintable byte in a name.
	ControlCharacter = register(2311, "ControlCharacter", Entry{
		Summary: "a name contains a control character",
		Detail:  "Suite and scenario names are printed in aligned tables, written into generated Markdown as headings, and turned into anchors and report fields. A newline, tab, or other control byte corrupts every one of those, and a name that renders differently in each place cannot be selected reliably with `--scenario`.",
		Fix:     "Remove the control character. A name spanning lines usually comes from a YAML block scalar (`|` or `>`) where a plain string was meant.",
		Since:   specSince,
	})

	// VacuousMatcher is a matcher that cannot do its job.
	VacuousMatcher = register(2312, "VacuousMatcher", Entry{
		Summary: "a matcher can never fail, or can never pass",
		Detail:  "The matcher is well-formed and asserts nothing. An empty substring is contained in every string, so `contains: \"\"` always passes and `not_contains: \"\"` never can; a regexp matching the empty string behaves the same way. A test that cannot fail is worse than no test, because it reports success and is counted as coverage.",
		Fix:     "Give the matcher a real substring or pattern, or remove it. Where a pattern is meant to match an empty field, anchor it (`^$`) rather than leaving it empty.",
		Since:   specSince,
	})

	// BadFormat is a value that does not follow the notation its key requires.
	BadFormat = register(2313, "BadFormat", Entry{
		Summary: "a value is not written in the notation the key requires",
		Detail:  "The key takes a value in a specific notation rather than from a fixed list of choices — an octal file mode, an RFC 3339 timestamp, base64 data, a JSON path, a color, a URL path beginning with a slash. The value given does not parse in that notation.",
		Fix:     "Rewrite the value in the notation the message names. Its example shows the accepted form.",
		Since:   specSince,
	})
)

// ATG24xx — a name that does not resolve.
var (
	// RunnerNotDeclared is a step naming a runner the spec never declared.
	RunnerNotDeclared = register(2401, "RunnerNotDeclared", Entry{
		Summary: "a step names a runner the spec does not declare",
		Detail:  "Steps that reach outside the local machine — `http`, `query`, `grpc`, `cdp`, `ssh` — address a runner by name, and runners are declared once under the spec's `runners:` block. A name with no declaration behind it is a typo or a block that was never added.",
		Fix:     "Declare the runner under `runners:`, or correct the name to one that is declared. The message lists the names the spec does declare.",
		Since:   specSince,
	})

	// RunnerTypeMismatch is a declared runner of the wrong kind.
	RunnerTypeMismatch = register(2402, "RunnerTypeMismatch", Entry{
		Summary: "a step names a runner of the wrong type",
		Detail:  "The runner is declared, but it is not the kind this step needs: a `query` step needs a database runner, an `http` step needs an HTTP runner. The usual cause is one name copied where another was meant, in a spec that declares several runners.",
		Fix:     "Name a runner of the type the message asks for.",
		Since:   specSince,
	})

	// NameNotDeclared is a reference to a service, mock, or stored value that
	// the spec never defines.
	NameNotDeclared = register(2403, "NameNotDeclared", Entry{
		Summary: "a reference names something the spec does not declare",
		Detail:  "Services, mock servers, and stored values are addressed by name from elsewhere in the spec — a signal targets a service, an assertion targets a mock, an expansion reads a stored value. A name nothing declares would be resolved at run time against nothing at all.",
		Fix:     "Declare the target, or correct the name. Suite-level services and mocks are visible to every scenario; a scenario's own are visible only within it.",
		Since:   specSince,
	})

	// ReservedName is a name atago has already given a meaning.
	ReservedName = register(2404, "ReservedName", Entry{
		Summary: "a name collides with one atago reserves",
		Detail:  "atago provides a few variables of its own — `atago`, `workdir`, `suitedir` — and a stored value taking one of those names would shadow it. Every later expansion would then read the spec's value where it meant atago's, silently and only in the specs that happen to declare the name.",
		Fix:     "Choose another name for the stored value.",
		Since:   specSince,
	})

	// PathNotUsable is a path that resolves to nothing, or to the wrong kind
	// of thing.
	PathNotUsable = register(2405, "PathNotUsable", Entry{
		Summary: "a path does not exist, or is not the kind of thing the key needs",
		Detail:  "Some paths are resolved while loading rather than during a run, because a typo caught here is one message about the spec instead of an identical failure in every scenario that uses it. A directory manifest's `fixtures_dir` is the usual case: a path that is missing, or that names a file where a directory is wanted.",
		Fix:     "Correct the path, or create what it points at. Relative paths are resolved against the file that declares them.",
		Since:   specSince,
	})
)

// ATG25xx — identity.
var (
	// DuplicateName is two things in one spec answering to the same name.
	DuplicateName = register(2501, "DuplicateName", Entry{
		Summary: "two entries in the same spec share a name",
		Detail:  "Names identify: `--scenario` selects by them, reports and generated docs are keyed by them, and services and mocks are addressed by them. With a duplicate, selecting one of the pair is impossible and a report cannot say which of the two it describes. Copy-and-edit that stopped before editing the name is the usual cause.",
		Fix:     "Rename one of the two. Where the duplicates were meant to be one thing parameterized, `matrix:` generates distinct names for each combination.",
		Since:   specSince,
	})

	// DuplicateEntry is a list repeating something it must list once.
	DuplicateEntry = register(2502, "DuplicateEntry", Entry{
		Summary: "a list repeats an entry that must appear once",
		Detail:  "Some lists are sets: the exit codes `in` accepts, the observables `deterministic.compare` watches, the modifiers a key press carries. A repeat has no second meaning, so it is either a merge artefact or a value that was meant to be different.",
		Fix:     "Remove the repeat, or change it to the value that was meant.",
		Since:   specSince,
	})
)
