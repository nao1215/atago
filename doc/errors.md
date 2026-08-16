# Error codes

A coded error carries a name that can be searched for, linked to, and branched on, so the message beside it stays free to improve without breaking anyone. Assertion failures are not errors in this sense and carry no code: a spec that fails is a spec doing its job, and its message already says what was expected and what happened.

A code is `ATG` followed by four digits, and the first of those digits is the exit status the run produced. Reading `ATG2103` tells you the process exited 2 before it tells you anything else.

Codes are being assigned one family at a time. The table says which families carry them today; an error from a family not yet covered still exits the same way and still says what went wrong, it just has no code to look up here yet.

| Codes | Exit | Meaning | Assigned |
|---|---|---|---|
| `ATG2xxx` | 2 | spec error — the file could not be parsed, or does not describe a runnable suite | 39 codes |
| `ATG3xxx` | 3 | configuration error — the command line, its flags, or the spec files it selected | 14 codes |
| `ATG4xxx` | 4 | execution error — a step could not be carried out | not yet |
| `ATG5xxx` | 5 | internal error — a bug in atago | not yet |
| `ATG6xxx` | 6 | security policy violation — atago refused an operation on policy grounds | not yet |

`ATG1xxx` is never assigned. Exit 1 means one or more scenarios failed, which is a result rather than an error.

Codes are grouped by what you have to fix rather than by where in atago the error was raised, so one code can be reported from several places when the answer is the same in all of them.

Look a code up from the terminal with `atago explain ATG2201`, which prints this same text without a browser.

## Every code

| Code | Meaning | Exit | Since |
|---|---|---|---|
| [`ATG2001`](#atg2001--the-spec-file-could-not-be-read) | the spec file could not be read | 2 | v0.21.0 |
| [`ATG2002`](#atg2002--the-spec-file-contains-no-yaml-document) | the spec file contains no YAML document | 2 | v0.21.0 |
| [`ATG2003`](#atg2003--the-file-is-not-valid-yaml) | the file is not valid YAML | 2 | v0.21.0 |
| [`ATG2004`](#atg2004--an-explicit-yaml-tag-is-not-supported-in-a-spec) | an explicit YAML tag is not supported in a spec | 2 | v0.21.0 |
| [`ATG2005`](#atg2005--the-spec-sets-a-key-the-schema-does-not-define) | the spec sets a key the schema does not define | 2 | v0.21.0 |
| [`ATG2006`](#atg2006--a-value-is-written-in-a-shape-the-key-does-not-accept) | a value is written in a shape the key does not accept | 2 | v0.21.0 |
| [`ATG2010`](#atg2010--the-spec-declares-a-format-version-atago-does-not-support) | the spec declares a format version atago does not support | 2 | v0.21.0 |
| [`ATG2101`](#atg2101--a-step-sets-no-action) | a step sets no action | 2 | v0.21.0 |
| [`ATG2102`](#atg2102--a-step-sets-more-than-one-action) | a step sets more than one action | 2 | v0.21.0 |
| [`ATG2103`](#atg2103--keys-that-contradict-each-other-are-set-together) | keys that contradict each other are set together | 2 | v0.21.0 |
| [`ATG2104`](#atg2104--a-group-of-keys-that-takes-exactly-one-member-has-none-or-several) | a group of keys that takes exactly one member has none or several | 2 | v0.21.0 |
| [`ATG2105`](#atg2105--a-key-is-real-but-has-no-meaning-in-this-position) | a key is real but has no meaning in this position | 2 | v0.21.0 |
| [`ATG2106`](#atg2106--a-block-is-not-allowed-at-this-level-of-the-spec) | a block is not allowed at this level of the spec | 2 | v0.21.0 |
| [`ATG2107`](#atg2107--an-assertion-needs-a-preceding-step-it-does-not-have) | an assertion needs a preceding step it does not have | 2 | v0.21.0 |
| [`ATG2108`](#atg2108--a-key-needs-another-key-that-is-not-set) | a key needs another key that is not set | 2 | v0.21.0 |
| [`ATG2201`](#atg2201--a-required-key-is-missing) | a required key is missing | 2 | v0.21.0 |
| [`ATG2202`](#atg2202--a-list-that-must-hold-at-least-one-entry-is-empty) | a list that must hold at least one entry is empty | 2 | v0.21.0 |
| [`ATG2203`](#atg2203--a-group-of-keys-that-needs-at-least-one-member-has-none) | a group of keys that needs at least one member has none | 2 | v0.21.0 |
| [`ATG2204`](#atg2204--a-value-that-must-not-be-empty-is-empty) | a value that must not be empty is empty | 2 | v0.21.0 |
| [`ATG2301`](#atg2301--a-duration-is-not-written-in-a-form-atago-can-read) | a duration is not written in a form atago can read | 2 | v0.21.0 |
| [`ATG2302`](#atg2302--a-value-is-negative-where-a-negative-has-no-meaning) | a value is negative where a negative has no meaning | 2 | v0.21.0 |
| [`ATG2303`](#atg2303--a-value-must-be-positive-and-is-zero-or-below) | a value must be positive and is zero or below | 2 | v0.21.0 |
| [`ATG2304`](#atg2304--a-number-is-outside-the-range-the-key-accepts) | a number is outside the range the key accepts | 2 | v0.21.0 |
| [`ATG2305`](#atg2305--a-regular-expression-does-not-compile) | a regular expression does not compile | 2 | v0.21.0 |
| [`ATG2306`](#atg2306--a-glob-pattern-is-not-valid) | a glob pattern is not valid | 2 | v0.21.0 |
| [`ATG2307`](#atg2307--a-value-is-not-one-of-the-choices-the-key-allows) | a value is not one of the choices the key allows | 2 | v0.21.0 |
| [`ATG2308`](#atg2308--the-bounds-of-a-range-describe-an-interval-nothing-can-fall-in) | the bounds of a range describe an interval nothing can fall in | 2 | v0.21.0 |
| [`ATG2309`](#atg2309--a-path-is-absolute-where-it-must-be-workdir-relative) | a path is absolute where it must be workdir-relative | 2 | v0.21.0 |
| [`ATG2310`](#atg2310--a-path-leads-outside-the-scenario-workdir) | a path leads outside the scenario workdir | 2 | v0.21.0 |
| [`ATG2311`](#atg2311--a-name-contains-a-control-character) | a name contains a control character | 2 | v0.21.0 |
| [`ATG2312`](#atg2312--a-matcher-can-never-fail-or-can-never-pass) | a matcher can never fail, or can never pass | 2 | v0.21.0 |
| [`ATG2313`](#atg2313--a-value-is-not-written-in-the-notation-the-key-requires) | a value is not written in the notation the key requires | 2 | v0.21.0 |
| [`ATG2401`](#atg2401--a-step-names-a-runner-the-spec-does-not-declare) | a step names a runner the spec does not declare | 2 | v0.21.0 |
| [`ATG2402`](#atg2402--a-step-names-a-runner-of-the-wrong-type) | a step names a runner of the wrong type | 2 | v0.21.0 |
| [`ATG2403`](#atg2403--a-reference-names-something-the-spec-does-not-declare) | a reference names something the spec does not declare | 2 | v0.21.0 |
| [`ATG2404`](#atg2404--a-name-collides-with-one-atago-reserves) | a name collides with one atago reserves | 2 | v0.21.0 |
| [`ATG2405`](#atg2405--a-path-does-not-exist-or-is-not-the-kind-of-thing-the-key-needs) | a path does not exist, or is not the kind of thing the key needs | 2 | v0.21.0 |
| [`ATG2501`](#atg2501--two-entries-in-the-same-spec-share-a-name) | two entries in the same spec share a name | 2 | v0.21.0 |
| [`ATG2502`](#atg2502--a-list-repeats-an-entry-that-must-appear-once) | a list repeats an entry that must appear once | 2 | v0.21.0 |
| [`ATG3001`](#atg3001--there-is-no-such-atago-subcommand) | there is no such atago subcommand | 3 | v0.21.0 |
| [`ATG3002`](#atg3002--a-subcommand-was-called-in-a-shape-it-does-not-accept) | a subcommand was called in a shape it does not accept | 3 | v0.21.0 |
| [`ATG3101`](#atg3101--the-command-line-sets-an-option-atago-does-not-define) | the command line sets an option atago does not define | 3 | v0.21.0 |
| [`ATG3102`](#atg3102--an-option-was-given-a-value-it-does-not-accept) | an option was given a value it does not accept | 3 | v0.21.0 |
| [`ATG3103`](#atg3103--an-option-needs-another-option-that-is-not-set) | an option needs another option that is not set | 3 | v0.21.0 |
| [`ATG3104`](#atg3104--two-options-that-contradict-each-other-are-both-set) | two options that contradict each other are both set | 3 | v0.21.0 |
| [`ATG3105`](#atg3105--a-numeric-option-is-outside-the-range-it-accepts) | a numeric option is outside the range it accepts | 3 | v0.21.0 |
| [`ATG3201`](#atg3201--a-path-on-the-command-line-cannot-be-reached) | a path on the command line cannot be reached | 3 | v0.21.0 |
| [`ATG3202`](#atg3202--no-spec-files-were-found-under-the-paths-given) | no spec files were found under the paths given | 3 | v0.21.0 |
| [`ATG3203`](#atg3203--the-scenario-selection-matched-nothing-and---ci-refuses-to-call-that-success) | the scenario selection matched nothing, and --ci refuses to call that success | 3 | v0.21.0 |
| [`ATG3204`](#atg3204--no-recorded-failing-scenario-matched-the-current-specs) | no recorded failing scenario matched the current specs | 3 | v0.21.0 |
| [`ATG3205`](#atg3205--the-output-file-already-exists) | the output file already exists | 3 | v0.21.0 |
| [`ATG3206`](#atg3206--a-destination-atago-was-asked-to-write-cannot-be-written) | a destination atago was asked to write cannot be written | 3 | v0.21.0 |
| [`ATG3207`](#atg3207--atagos-recorded-state-from-a-previous-run-cannot-be-read) | atago's recorded state from a previous run cannot be read | 3 | v0.21.0 |

## ATG2xxx — exit 2

### ATG2001 — the spec file could not be read

atago was given a path and the operating system refused to open it. The file may not exist, may be a directory, or may not be readable by the user running atago. Nothing was parsed, so no other diagnostic can be reported about this file.

Fix: Check the path as spelled on the command line, and check the file's permissions. `atago list <dir>` shows which spec files atago can actually see in a directory.

Exits 2. Since v0.21.0.

### ATG2002 — the spec file contains no YAML document

The file was read but holds nothing a spec can be built from — it is empty, or contains only whitespace and comments. This is most often a file that was created but never filled in, or one whose contents were lost by an editor or a bad merge.

Fix: A spec needs at least `version`, `suite`, and `scenarios`. Run `atago init` to write a runnable starter spec and edit that.

Exits 2. Since v0.21.0.

### ATG2003 — the file is not valid YAML

Parsing stopped before atago could look at what the document means. The reported line and column are where the parser gave up, which is usually at or just after the real mistake. Inconsistent indentation, a missing colon after a key, and an unclosed quote or bracket account for most of these.

Fix: Fix the syntax at the reported position. Indentation in YAML must use spaces, never tabs, and every level must be indented consistently.

Exits 2. Since v0.21.0.

### ATG2004 — an explicit YAML tag is not supported in a spec

The document carries an explicit tag such as `!!str` or `!!map`. atago's schema is closed and fully typed: every field's type already fixes how its value is read, so a tag can only restate what the schema says or contradict it. The one exception is `!!binary`, which `atago record` writes when a captured stream is not valid UTF-8.

Fix: Remove the tag. If a value needs to be forced to text, quote it instead.

Exits 2. Since v0.21.0.

### ATG2005 — the spec sets a key the schema does not define

Specs are decoded strictly, so a key atago does not know is an error rather than something quietly ignored — a silently dropped `assert:` block is a test that passes without testing anything. The usual causes are a typo, a key written at the wrong nesting level, and a key from a newer atago than the one running.

Fix: Check the spelling and the indentation against the message's suggestion. The JSON Schema under `schema/atago.schema.json` gives an editor live completion for every key.

Exits 2. Since v0.21.0.

### ATG2006 — a value is written in a shape the key does not accept

The key exists but its value is the wrong kind of thing — a bare string where a mapping is wanted, a mapping where a list is wanted, or text where a number is wanted. The most common instance by far is a stream assertion written as a bare scalar: `stdout: hello` rather than `stdout: {contains: hello}`.

Fix: Write the value in the shape the key takes. Stream assertions (`stdout`, `stderr`, `body`, `rows`, `message`, `value`) always take a matcher mapping such as `{contains: "..."}` or `{equals: "..."}`.

Exits 2. Since v0.21.0.

### ATG2010 — the spec declares a format version atago does not support

The `version` key names the spec format, not the atago release, and `"1"` is the only format that exists. A version of `1` without quotes is read as a number and is also refused, because the field is a format name rather than a quantity.

Fix: Write `version: "1"`, with the quotes.

Exits 2. Since v0.21.0.

### ATG2101 — a step sets no action

Every step does exactly one thing: run a command, make a request, drive a terminal, write a fixture, assert, or one of the other actions. A step with none of them is usually a block whose action key was lost to a bad indent, or a placeholder that was never filled in.

Fix: Give the step exactly one of `fixture`, `run`, `http`, `query`, `grpc`, `cdp`, `assert`, `store`, `pty`, or `signal` — or delete the step.

Exits 2. Since v0.21.0.

### ATG2102 — a step sets more than one action

A step is one action so that ordering, timing, and failure are unambiguous: a step that both ran a command and made a request has no single answer for which one a `duration` bounds or which one an assertion followed. The usual cause is a second action indented into the previous step instead of starting a new one.

Fix: Split the actions into separate list items under `steps:`, one list item per action.

Exits 2. Since v0.21.0.

### ATG2103 — keys that contradict each other are set together

Two keys were set whose meanings cannot both hold. A snapshot already pins every byte of what it captures, so pairing it with a size bound or with the `contains`/`count`/`glob` family asks the same question twice and invites the two answers to disagree. An exact `size` next to `min_size`/`max_size`, two lower bounds, and `exists: false` next to a size are the same mistake in other places.

Fix: Keep the key that expresses what you actually want to pin and drop the other. The message names both.

Exits 2. Since v0.21.0.

### ATG2104 — a group of keys that takes exactly one member has none or several

Some blocks offer alternatives that are mutually exclusive by nature: a `pty` step is one of expect, send, expect_screen, resize, or exec; an `exit_code` is a bare number, a `not`, or an `in`. Setting several leaves the meaning undefined, and setting none leaves the block with nothing to do.

Fix: Set exactly one member of the group the message lists. To do two things, write two entries.

Exits 2. Since v0.21.0.

### ATG2105 — a key is real but has no meaning in this position

The key exists elsewhere in the schema but not here, and accepting it silently would make it look effective when it is not. `defaults.run.command` is the clearest case: defaults describe how every step runs, while the command is what an individual step is, so a default command would be inherited by steps that already have one and quietly ignored. `trim` and `text` outside a store source, and `recursive` next to `snapshot`, are the same shape of mistake.

Fix: Move the key to where it applies, or drop it. The message says which of the two the key needs.

Exits 2. Since v0.21.0.

### ATG2106 — a block is not allowed at this level of the spec

Some blocks are tied to a level by what they own. A `service` or `mock_server` step belongs in `suite.setup` because it outlives any one scenario; a scenario-scoped peer goes in that scenario's own `services:` or `mock_servers:` list instead. Others go the other way: steps needing a scenario workdir and runners cannot sit at suite level, and `assert.changes` is not available in teardown because the workdir delta is tracked only around a scenario's steps.

Fix: Move the block to the level the message names.

Exits 2. Since v0.21.0.

### ATG2107 — an assertion needs a preceding step it does not have

Some assertions describe the step before them rather than the scenario as a whole. `assert.screen` reads the terminal a `pty` step rendered, `assert.duration` bounds the wall-clock time of the step it follows, and `assert.changes` pins the workdir delta of the run or pty step it follows. Without that step there is nothing for the assertion to be about, so it would pass or fail for reasons unrelated to what it says.

Fix: Put the assertion directly after the step it describes. One assert block may set several keys at once, so an `exit_code`, a `stdout`, and a `changes` about the same step belong together.

Exits 2. Since v0.21.0.

### ATG2108 — a key needs another key that is not set

The key is only meaningful in the presence of a second one: `max_diff` bounds the difference `similar_to` measures, `pass_env` selects which host variables survive a `clear_env`, and `count` counts the matches of a `contains`. Alone, the key has nothing to act on, so accepting it would leave a spec that looks like it constrains something and does not.

Fix: Add the key the message names, or drop the one that depends on it.

Exits 2. Since v0.21.0.

### ATG2201 — a required key is missing

The block cannot be acted on without this key: a service with no command has nothing to start, an HTTP step with no method has no request to make, a scenario with no name cannot be selected, reported on, or documented. atago refuses these at load time rather than discovering them mid-run, so a broken spec fails before it starts changing anything.

Fix: Add the key the message names to the block it names.

Exits 2. Since v0.21.0.

### ATG2202 — a list that must hold at least one entry is empty

A suite with no scenarios, a scenario with no steps, or a `cdp` block with no actions is accepted by YAML and means nothing to run. Left alone it would report success without having tested anything, which is worse than failing.

Fix: Add an entry to the list, or remove the block that holds it.

Exits 2. Since v0.21.0.

### ATG2203 — a group of keys that needs at least one member has none

Unlike an exactly-one group, these accept any combination — but not the empty one, which would assert nothing. `assert.changes` needs at least one of `created`, `modified`, or `deleted`; a bounds block needs at least one of `lt`, `lte`, `gt`, or `gte`; a recursive directory assert needs at least one matcher.

Fix: Set one or more members of the group the message lists. To assert that a category changed nothing, write it explicitly as an empty list, e.g. `created: []`.

Exits 2. Since v0.21.0.

### ATG2204 — a value that must not be empty is empty

The key is present but carries nothing. This is worse than leaving the key out, because an empty value usually does not mean what it looks like: an empty ignore pattern excludes nothing, an empty name identifies nothing, an empty path names the workdir itself.

Fix: Give the key a value, or remove the key. Where removing it is the right answer the message says so.

Exits 2. Since v0.21.0.

### ATG2301 — a duration is not written in a form atago can read

Durations are written as a number with a unit — `500ms`, `2s`, `2m`, `1h30m` — and a bare number is refused because it does not say which unit was meant. A timeout that silently meant nanoseconds instead of seconds would look like a hang.

Fix: Write the value with a unit, e.g. `timeout: 2m`. Where a duration can be disabled, `"0"` does that.

Exits 2. Since v0.21.0.

### ATG2302 — a value is negative where a negative has no meaning

Wall-clock bounds, sizes, counts, and page numbers are never below zero. A negative here is nearly always a sign slip or an arithmetic mistake in a generated spec, and treating it as zero would hide that.

Fix: Use zero or a positive value. Where the intent was to disable a bound rather than set it to nothing, the message names the way to do that.

Exits 2. Since v0.21.0.

### ATG2303 — a value must be positive and is zero or below

Some values have no sensible zero: an interval that never elapses, a worker count of nobody, a retry that waits no time between attempts. Where atago has a working default, zero would be indistinguishable from not setting the key at all.

Fix: Set a positive value, or omit the key to take atago's default.

Exits 2. Since v0.21.0.

### ATG2304 — a number is outside the range the key accepts

The value parsed but falls outside what the field can mean. Screen coordinates are 1-based cells, so row or column zero is not the first cell but a mistake, and it is the instance of this seen most often — a spec written against a 0-based mental model reads the wrong cell everywhere or nowhere.

Fix: Bring the value into the range the message states.

Exits 2. Since v0.21.0.

### ATG2305 — a regular expression does not compile

atago compiles every pattern in a spec at load time — `matches`, scrub patterns, readiness patterns — so a broken one is reported before the run rather than at the moment it would first have been used, halfway through a suite. The syntax is Go's RE2: no backreferences and no lookaround.

Fix: Correct the pattern at the position the message reports. Regex metacharacters that are meant literally need escaping, and a pattern inside a YAML double-quoted string needs its backslashes doubled — single quotes avoid that.

Exits 2. Since v0.21.0.

### ATG2306 — a glob pattern is not valid

Globs match paths by shape — `*` within a path segment, `**` across segments, `?` for one character, `[...]` for a class, `{a,b}` for alternatives. An unclosed bracket or brace is the usual cause.

Fix: Close the unbalanced bracket or brace, or escape it if it was meant literally.

Exits 2. Since v0.21.0.

### ATG2307 — a value is not one of the choices the key allows

The key takes a fixed vocabulary and the value is not in it. Platform names are the common case: `os` accepts `linux`, `darwin`, or `windows`, so `macos` or `osx` — reasonable words that Go's platform names do not use — are refused rather than quietly never matching.

Fix: Use one of the values the message lists.

Exits 2. Since v0.21.0.

### ATG2308 — the bounds of a range describe an interval nothing can fall in

A lower bound at or above the upper bound leaves no value that could satisfy both, so the assertion can only ever fail. This is almost always two bounds swapped, or a copied block whose second bound was not updated.

Fix: Order the bounds so the lower one is genuinely below the upper one, or use a single bound if only one side matters.

Exits 2. Since v0.21.0.

### ATG2309 — a path is absolute where it must be workdir-relative

Paths in a spec are relative to the scenario's own workdir, which is created fresh for each scenario. That is what lets the same spec run on another machine, in CI, and in parallel with itself. An absolute path names one machine's filesystem and breaks all three.

Fix: Write the path relative to the scenario workdir, e.g. `out/report.json` rather than `/home/you/project/out/report.json`.

Exits 2. Since v0.21.0.

### ATG2310 — a path leads outside the scenario workdir

The path is relative but climbs out of the workdir with `../`. Scenario isolation is what makes a failing scenario reproducible and a parallel run safe, and a path that reaches outside it can read or overwrite another scenario's files, or the repository being tested. This is refused while loading; the equivalent refusal during a run is reported as a security violation and exits 6.

Fix: Keep the path inside the workdir. To work with a file that genuinely lives elsewhere, copy it in with a `fixture` step, which is the supported way to bring outside data into a scenario.

Exits 2. Since v0.21.0.

### ATG2311 — a name contains a control character

Suite and scenario names are printed in aligned tables, written into generated Markdown as headings, and turned into anchors and report fields. A newline, tab, or other control byte corrupts every one of those, and a name that renders differently in each place cannot be selected reliably with `--scenario`.

Fix: Remove the control character. A name spanning lines usually comes from a YAML block scalar (`|` or `>`) where a plain string was meant.

Exits 2. Since v0.21.0.

### ATG2312 — a matcher can never fail, or can never pass

The matcher is well-formed and asserts nothing. An empty substring is contained in every string, so `contains: ""` always passes and `not_contains: ""` never can; a regexp matching the empty string behaves the same way. A test that cannot fail is worse than no test, because it reports success and is counted as coverage.

Fix: Give the matcher a real substring or pattern, or remove it. Where a pattern is meant to match an empty field, anchor it (`^$`) rather than leaving it empty.

Exits 2. Since v0.21.0.

### ATG2313 — a value is not written in the notation the key requires

The key takes a value in a specific notation rather than from a fixed list of choices — an octal file mode, an RFC 3339 timestamp, base64 data, a JSON path, a color, a URL path beginning with a slash. The value given does not parse in that notation.

Fix: Rewrite the value in the notation the message names. Its example shows the accepted form.

Exits 2. Since v0.21.0.

### ATG2401 — a step names a runner the spec does not declare

Steps that reach outside the local machine — `http`, `query`, `grpc`, `cdp`, `ssh` — address a runner by name, and runners are declared once under the spec's `runners:` block. A name with no declaration behind it is a typo or a block that was never added.

Fix: Declare the runner under `runners:`, or correct the name to one that is declared. The message lists the names the spec does declare.

Exits 2. Since v0.21.0.

### ATG2402 — a step names a runner of the wrong type

The runner is declared, but it is not the kind this step needs: a `query` step needs a database runner, an `http` step needs an HTTP runner. The usual cause is one name copied where another was meant, in a spec that declares several runners.

Fix: Name a runner of the type the message asks for.

Exits 2. Since v0.21.0.

### ATG2403 — a reference names something the spec does not declare

Services, mock servers, and stored values are addressed by name from elsewhere in the spec — a signal targets a service, an assertion targets a mock, an expansion reads a stored value. A name nothing declares would be resolved at run time against nothing at all.

Fix: Declare the target, or correct the name. Suite-level services and mocks are visible to every scenario; a scenario's own are visible only within it.

Exits 2. Since v0.21.0.

### ATG2404 — a name collides with one atago reserves

atago provides a few variables of its own — `atago`, `workdir`, `suitedir` — and a stored value taking one of those names would shadow it. Every later expansion would then read the spec's value where it meant atago's, silently and only in the specs that happen to declare the name.

Fix: Choose another name for the stored value.

Exits 2. Since v0.21.0.

### ATG2405 — a path does not exist, or is not the kind of thing the key needs

Some paths are resolved while loading rather than during a run, because a typo caught here is one message about the spec instead of an identical failure in every scenario that uses it. A directory manifest's `fixtures_dir` is the usual case: a path that is missing, or that names a file where a directory is wanted.

Fix: Correct the path, or create what it points at. Relative paths are resolved against the file that declares them.

Exits 2. Since v0.21.0.

### ATG2501 — two entries in the same spec share a name

Names identify: `--scenario` selects by them, reports and generated docs are keyed by them, and services and mocks are addressed by them. With a duplicate, selecting one of the pair is impossible and a report cannot say which of the two it describes. Copy-and-edit that stopped before editing the name is the usual cause.

Fix: Rename one of the two. Where the duplicates were meant to be one thing parameterized, `matrix:` generates distinct names for each combination.

Exits 2. Since v0.21.0.

### ATG2502 — a list repeats an entry that must appear once

Some lists are sets: the exit codes `in` accepts, the observables `deterministic.compare` watches, the modifiers a key press carries. A repeat has no second meaning, so it is either a merge artifact or a value that was meant to be different.

Fix: Remove the repeat, or change it to the value that was meant.

Exits 2. Since v0.21.0.

## ATG3xxx — exit 3

### ATG3001 — there is no such atago subcommand

The first argument names a subcommand, and this one is not in the inventory. Running atago with no arguments at all reports the same thing, since there is no default subcommand — a bare `atago` that started running specs would be a surprising thing for a tool that writes files.

Fix: Run `atago help` for the list of subcommands.

Exits 3. Since v0.21.0.

### ATG3002 — a subcommand was called in a shape it does not accept

The subcommand exists but its arguments are not what it takes: `atago snapshot` without `update`, `atago completion` without a shell to generate for, `atago record` with no command after the `--`, or `atago init` given several paths when it writes one file. This is about the arguments' shape rather than their values, which is why it is separate from an option given a value it does not accept.

Fix: Run the subcommand with `--help` for its usage line.

Exits 3. Since v0.21.0.

### ATG3101 — the command line sets an option atago does not define

The option is not one this subcommand takes. Options are per-subcommand, so a flag that works for `atago run` is not necessarily accepted by `atago doc`, and a flag from a newer atago than the one running produces this too.

Fix: Check the option against the usage printed below the message, or run the subcommand with `--help`.

Exits 3. Since v0.21.0.

### ATG3102 — an option was given a value it does not accept

The option exists and takes a value from a fixed vocabulary, and this value is not in it — a report format, a shell to generate completion for, a scaffolding template. atago refuses rather than falling back to a default, because a misspelled `--report jnit` that quietly produced console output would leave a CI job collecting a report file that was never written.

Fix: Use one of the values the message lists.

Exits 3. Since v0.21.0.

### ATG3103 — an option needs another option that is not set

Some options only mean something in company: `--split-by-spec` writes one file per spec and needs the `--out-dir` to write them into, and `--snapshot` on `record` writes a golden next to the spec, which needs `--out` to say where the spec is going.

Fix: Add the option the message names, or drop the one that depends on it.

Exits 3. Since v0.21.0.

### ATG3104 — two options that contradict each other are both set

The two options ask for incompatible things. `--repeat` and `--retry-failed` are the clearest pair: one re-runs scenarios to find out whether they flake, the other re-runs them so that flakiness does not fail the build, and a run cannot both detect and tolerate the same instability. `--out` and `--split-by-spec` disagree about whether the output is one file or many.

Fix: Keep the option that expresses what you want and drop the other.

Exits 3. Since v0.21.0.

### ATG3105 — a numeric option is outside the range it accepts

Counts are never negative. A negative `--parallel` would otherwise be clamped to sequential and exit 0, so the typo would run the suite in a way nobody asked for and report success; the same applies to `--repeat` and `--retry-failed`.

Fix: Use zero or a positive number. Zero means atago's default for these options rather than none.

Exits 3. Since v0.21.0.

### ATG3201 — a path on the command line cannot be reached

atago was pointed at a file or directory that does not exist, or that it is not allowed to look at. In CI this is usually a working directory that is not what the workflow assumed, or a checkout step that has not run yet.

Fix: Check the path as spelled, and the directory the command runs from.

Exits 3. Since v0.21.0.

### ATG3202 — no spec files were found under the paths given

The paths exist but hold no `*.atago.yaml` or `*.atago.yml` files. Directories are searched recursively, so this means there genuinely are none below them. Reporting success here would let a CI job that lost its specs go green forever.

Fix: Point atago at the directory holding the specs, or run `atago init` to scaffold one. The file name matters: a spec must end in `.atago.yaml` or `.atago.yml`.

Exits 3. Since v0.21.0.

### ATG3203 — the scenario selection matched nothing, and --ci refuses to call that success

Specs were found and loaded, but `--filter`, `--tag`, or `--skip-tag` selected none of their scenarios. Outside `--ci` this is a warning; under `--ci` it fails, because a selection that silently stops matching is how a suite disables itself without anyone noticing — a renamed tag keeps the build green while testing nothing. The two selectors match differently: `--filter` is a case-sensitive substring of the name, while `--tag` and `--skip-tag` compare tags for exact equality.

Fix: Run `atago list` to see the scenario names and tags actually present, and correct the selector.

Exits 3. Since v0.21.0.

### ATG3204 — no recorded failing scenario matched the current specs

`--rerun-failed` re-runs what failed last time, and none of the recorded names exist any more — they were renamed or removed since the run that recorded them. The recorded failures are kept rather than cleared, because the work they represent is still unverified, and exiting 0 here would report a green run that tested nothing.

Fix: Run the full suite once to record failures against the current names, or drop the recorded state and start again.

Exits 3. Since v0.21.0.

### ATG3205 — the output file already exists

`atago init` and `atago record` write a new spec, and neither overwrites one by default. Silently replacing a spec someone has edited is the kind of loss a tool should never cause without being asked.

Fix: Choose another path, or pass `--force` to overwrite the existing file deliberately.

Exits 3. Since v0.21.0.

### ATG3206 — a destination atago was asked to write cannot be written

The generated documentation, or the artifacts directory a run was told to use, could not be created or written. The artifacts directory is checked before the run starts rather than at the first failure: a directory that cannot be written would otherwise make every artifact write fail quietly, leaving a run that looks like it produced nothing to review when in fact nothing could be saved.

Fix: Check that the parent directory exists and is writable, and that nothing else already occupies the path as a file.

Exits 3. Since v0.21.0.

### ATG3207 — atago's recorded state from a previous run cannot be read

`--rerun-failed` reads a small ledger written by the previous run, and that file could not be read or decoded. A truncated or hand-edited ledger is the usual cause.

Fix: Delete the ledger and run the full suite once to record it again.

Exits 3. Since v0.21.0.

