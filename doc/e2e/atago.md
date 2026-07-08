# atago Behavior Specs
## Summary
74 suites · 383 scenarios
## Contents
- [atago self-hosting / cross-platform no-shell argv tokenization (#154)](#atago-self-hosting--cross-platform-no-shell-argv-tokenization-154) — 4 scenarios
  - [a single-quoted JSON argument survives tokenization](#scenario-a-single-quoted-json-argument-survives-tokenization)
  - [a single-quoted argument with a space stays one argument](#scenario-a-single-quoted-argument-with-a-space-stays-one-argument)
  - [a block-scalar command splits on newlines like spaces](#scenario-a-block-scalar-command-splits-on-newlines-like-spaces)
  - [a folded-scalar command drops its trailing newline](#scenario-a-folded-scalar-command-drops-its-trailing-newline)
- [atago self-hosting / artifacts-dir failure payloads](#atago-self-hosting--artifacts-dir-failure-payloads) — 4 scenarios
  - [a failing stdout equals writes expected and actual sidecars](#scenario-a-failing-stdout-equals-writes-expected-and-actual-sidecars)
  - [a passing scenario writes no failure payload](#scenario-a-passing-scenario-writes-no-failure-payload)
  - [the artifacts directory is created when it does not exist](#scenario-the-artifacts-directory-is-created-when-it-does-not-exist)
  - [a file-content mismatch also writes a payload](#scenario-a-file-content-mismatch-also-writes-a-payload)
- [atago self-hosting / variable expansion in assertion matcher values](#atago-self-hosting--variable-expansion-in-assertion-matcher-values) — 6 scenarios
  - [stdout.equals expands ${workdir}](#scenario-stdoutequals-expands-workdir)
  - [stdout.contains and not_contains expand a stored variable](#scenario-stdoutcontains-and-not_contains-expand-a-stored-variable)
  - [file.contains expands ${workdir}](#scenario-filecontains-expands-workdir)
  - [dir.path expands a stored variable](#scenario-dirpath-expands-a-stored-variable)
  - [changes entries expand a stored variable](#scenario-changes-entries-expand-a-stored-variable)
  - [screen matcher expands a stored variable](#scenario-screen-matcher-expands-a-stored-variable)
- [atago self-hosting / browser (cdp) runner](#atago-self-hosting--browser-cdp-runner) — 8 scenarios
  - [a cdp step with no actions fails validation (exit 2)](#scenario-a-cdp-step-with-no-actions-fails-validation-exit-2)
  - [a cdp step naming an undeclared runner fails validation (exit 2)](#scenario-a-cdp-step-naming-an-undeclared-runner-fails-validation-exit-2)
  - [a screenshot action without a path fails validation (exit 2)](#scenario-a-screenshot-action-without-a-path-fails-validation-exit-2)
  - [explain lists the extended cdp actions](#scenario-explain-lists-the-extended-cdp-actions)
  - [a browser-only field on a non-browser runner fails validation (exit 2)](#scenario-a-browser-only-field-on-a-non-browser-runner-fails-validation-exit-2)
  - [manifest surfaces the browser-runner configuration](#scenario-manifest-surfaces-the-browser-runner-configuration)
  - [an upload action without a file fails validation (exit 2)](#scenario-an-upload-action-without-a-file-fails-validation-exit-2)
  - [a download action without a click selector fails validation (exit 2)](#scenario-a-download-action-without-a-click-selector-fails-validation-exit-2)
- [atago self-hosting / changes (workdir delta assertions)](#atago-self-hosting--changes-workdir-delta-assertions) — 12 scenarios
  - [a generator touches exactly the files it should (POSIX)](#scenario-a-generator-touches-exactly-the-files-it-should-posix)
  - [an unexpected creation breaks the exact contract (POSIX)](#scenario-an-unexpected-creation-breaks-the-exact-contract-posix)
  - [stdout_to counts as created, and modified nothing holds (portable)](#scenario-stdout_to-counts-as-created-and-modified-nothing-holds-portable)
  - [the delta over a retried step is cumulative across all attempts (POSIX)](#scenario-the-delta-over-a-retried-step-is-cumulative-across-all-attempts-posix)
  - [deleting and recreating a byte-identical file appears in no list (POSIX)](#scenario-deleting-and-recreating-a-byte-identical-file-appears-in-no-list-posix)
  - [deleting and recreating with different content is modified only (POSIX)](#scenario-deleting-and-recreating-with-different-content-is-modified-only-posix)
  - [stdout_to overwrites a fixture (modified) while stderr_to creates an empty file (POSIX)](#scenario-stdout_to-overwrites-a-fixture-modified-while-stderr_to-creates-an-empty-file-posix)
  - [a pty step feeds the delta scan just like a run step (POSIX)](#scenario-a-pty-step-feeds-the-delta-scan-just-like-a-run-step-posix)
  - [a doublestar glob pins an arbitrary-depth generated tree exactly (POSIX)](#scenario-a-doublestar-glob-pins-an-arbitrary-depth-generated-tree-exactly-posix)
  - [a stray file outside the doublestar prefix breaks the exact contract (POSIX)](#scenario-a-stray-file-outside-the-doublestar-prefix-breaks-the-exact-contract-posix)
  - [a doublestar glob matches a nested redirect target (portable)](#scenario-a-doublestar-glob-matches-a-nested-redirect-target-portable)
  - [a doublestar prefix covers both redirect streams (portable)](#scenario-a-doublestar-prefix-covers-both-redirect-streams-portable)
- [atago self-hosting / CLI scenario selection](#atago-self-hosting--cli-scenario-selection) — 7 scenarios
  - [filter selects by a name substring](#scenario-filter-selects-by-a-name-substring)
  - [filter is OR across a comma-separated list](#scenario-filter-is-or-across-a-comma-separated-list)
  - [tag selects scenarios carrying the tag](#scenario-tag-selects-scenarios-carrying-the-tag)
  - [a repeated tag flag is OR](#scenario-a-repeated-tag-flag-is-or)
  - [skip-tag removes scenarios carrying the tag](#scenario-skip-tag-removes-scenarios-carrying-the-tag)
  - [tag and skip-tag compose as selected minus skipped](#scenario-tag-and-skip-tag-compose-as-selected-minus-skipped)
  - [a filter that matches nothing selects an empty set and still exits zero](#scenario-a-filter-that-matches-nothing-selects-an-empty-set-and-still-exits-zero)
- [atago self-hosting / completion](#atago-self-hosting--completion) — 5 scenarios
  - [bash completion emits a recognizable script](#scenario-bash-completion-emits-a-recognizable-script)
  - [zsh completion emits a compdef script](#scenario-zsh-completion-emits-a-compdef-script)
  - [fish completion emits complete directives](#scenario-fish-completion-emits-complete-directives)
  - [powershell completion registers an argument completer](#scenario-powershell-completion-registers-an-argument-completer)
  - [unknown shell is a configuration error](#scenario-unknown-shell-is-a-configuration-error)
- [atago self-hosting / db runner](#atago-self-hosting--db-runner) — 2 scenarios
  - [query workflow (create, insert, select, row assert, value binding) passes](#scenario-query-workflow-create-insert-select-row-assert-value-binding-passes)
  - [a query against an undeclared runner fails validation (exit 2)](#scenario-a-query-against-an-undeclared-runner-fails-validation-exit-2)
- [atago self-hosting / top-level defaults](#atago-self-hosting--top-level-defaults) — 6 scenarios
  - [defaults.run.shell applies to every run step without repeating it](#scenario-defaultsrunshell-applies-to-every-run-step-without-repeating-it)
  - [defaults.scenario.env is merged and an explicit scenario env wins](#scenario-defaultsscenarioenv-is-merged-and-an-explicit-scenario-env-wins)
  - [defaults.run.sandbox_home governs a run step and a pty step alike (POSIX)](#scenario-defaultsrunsandbox_home-governs-a-run-step-and-a-pty-step-alike-posix)
  - [an unsupported defaults field is a load-time error (exit 2)](#scenario-an-unsupported-defaults-field-is-a-load-time-error-exit-2)
  - [defaults.run.env merges per key and a step env wins the collisions](#scenario-defaultsrunenv-merges-per-key-and-a-step-env-wins-the-collisions)
  - [a step opts out of defaults.run.shell with an explicit shell false](#scenario-a-step-opts-out-of-defaultsrunshell-with-an-explicit-shell-false)
- [atago self-hosting / dir assertion](#atago-self-hosting--dir-assertion) — 3 scenarios
  - [directory/tree assertions cover a multi-file generator](#scenario-directorytree-assertions-cover-a-multi-file-generator)
  - [a missing directory can be asserted absent](#scenario-a-missing-directory-can-be-asserted-absent)
  - [a dangling symlink is a present directory entry (membership uses Lstat)](#scenario-a-dangling-symlink-is-a-present-directory-entry-membership-uses-lstat)
- [atago self-hosting / recursive dir asserts + tree snapshots](#atago-self-hosting--recursive-dir-asserts--tree-snapshots) — 3 scenarios
  - [record, compare green, then a mutation names the changed paths](#scenario-record-compare-green-then-a-mutation-names-the-changed-paths)
  - [recursive matchers and ignore globs walk the tree](#scenario-recursive-matchers-and-ignore-globs-walk-the-tree)
  - [combining snapshot with matchers is a load-time error](#scenario-combining-snapshot-with-matchers-is-a-load-time-error)
- [atago self-hosting / doc](#atago-self-hosting--doc) — 5 scenarios
  - [doc generates Markdown to a file](#scenario-doc-generates-markdown-to-a-file)
  - [doc writes Markdown to stdout without --out](#scenario-doc-writes-markdown-to-stdout-without---out)
  - [doc emits a summary, table of contents, and input previews](#scenario-doc-emits-a-summary-table-of-contents-and-input-previews)
  - [doc --split-by-spec writes one file per spec and an index](#scenario-doc---split-by-spec-writes-one-file-per-spec-and-an-index)
  - [doc --split-by-spec requires --out-dir](#scenario-doc---split-by-spec-requires---out-dir)
- [atago self-hosting / duration assertion](#atago-self-hosting--duration-assertion) — 4 scenarios
  - [a fast step passes a generous upper bound](#scenario-a-fast-step-passes-a-generous-upper-bound)
  - [an impossible bound fails and shows the measured duration](#scenario-an-impossible-bound-fails-and-shows-the-measured-duration)
  - [a deliberate wait satisfies a lower bound](#scenario-a-deliberate-wait-satisfies-a-lower-bound)
  - [a duration assert with no preceding step is a load-time error](#scenario-a-duration-assert-with-no-preceding-step-is-a-load-time-error)
- [atago self-hosting / edge cases](#atago-self-hosting--edge-cases) — 5 scenarios
  - [JSON assertion on empty stdout reports an empty stream](#scenario-json-assertion-on-empty-stdout-reports-an-empty-stream)
  - [an unsupported matcher is a parse error](#scenario-an-unsupported-matcher-is-a-parse-error)
  - [a mixed valid+invalid run reads FAILED and counts the dropped spec](#scenario-a-mixed-validinvalid-run-reads-failed-and-counts-the-dropped-spec)
  - [a snapshot update error names the snapshot command, not run](#scenario-a-snapshot-update-error-names-the-snapshot-command-not-run)
  - [a json assertion on malformed input fails cleanly, without a crash](#scenario-a-json-assertion-on-malformed-input-fails-cleanly-without-a-crash)
- [atago self-hosting / workdir + scenario env + not_contains](#atago-self-hosting--workdir--scenario-env--not_contains) — 6 scenarios
  - [run.stdout_to redirects stdout to a workdir file without a shell](#scenario-runstdout_to-redirects-stdout-to-a-workdir-file-without-a-shell)
  - [scenario env is shared by every run step and overridable per step](#scenario-scenario-env-is-shared-by-every-run-step-and-overridable-per-step)
  - [scenario env can reference ${workdir} for isolated paths](#scenario-scenario-env-can-reference-workdir-for-isolated-paths)
  - [file not_contains passes when the substring is absent](#scenario-file-not_contains-passes-when-the-substring-is-absent)
  - [not_contains fails when the substring is present](#scenario-not_contains-fails-when-the-substring-is-present)
  - [a shell metacharacter without shell is a load-time error](#scenario-a-shell-metacharacter-without-shell-is-a-load-time-error)
- [atago self-hosting / exit_code in-set matcher](#atago-self-hosting--exit_code-in-set-matcher) — 4 scenarios
  - [a listed exit code passes](#scenario-a-listed-exit-code-passes)
  - [an unlisted exit code fails and the output lists the set](#scenario-an-unlisted-exit-code-fails-and-the-output-lists-the-set)
  - [mixing not and in is a load-time error](#scenario-mixing-not-and-in-is-a-load-time-error)
  - [an empty in list is a load-time error](#scenario-an-empty-in-list-is-a-load-time-error)
- [atago self-hosting / exit code semantics](#atago-self-hosting--exit-code-semantics) — 14 scenarios
  - [a clean exit is zero](#scenario-a-clean-exit-is-zero)
  - [a general error is one](#scenario-a-general-error-is-one)
  - [a usage error is two](#scenario-a-usage-error-is-two)
  - [an arbitrary code passes through unchanged](#scenario-an-arbitrary-code-passes-through-unchanged)
  - [the single-byte ceiling is 255](#scenario-the-single-byte-ceiling-is-255)
  - [the not matcher excludes a specific code](#scenario-the-not-matcher-excludes-a-specific-code)
  - [the in matcher accepts any listed code](#scenario-the-in-matcher-accepts-any-listed-code)
  - [an unlisted code fails the in matcher and names the set](#scenario-an-unlisted-code-fails-the-in-matcher-and-names-the-set)
  - [SIGKILL is reported as 137](#scenario-sigkill-is-reported-as-137)
  - [SIGTERM is reported as 143](#scenario-sigterm-is-reported-as-143)
  - [SIGINT is reported as 130](#scenario-sigint-is-reported-as-130)
  - [a signal exit composes with the in matcher alongside normal codes](#scenario-a-signal-exit-composes-with-the-in-matcher-alongside-normal-codes)
  - [a missing command is 127 under the shell](#scenario-a-missing-command-is-127-under-the-shell)
  - [POSIX exit codes wrap modulo 256](#scenario-posix-exit-codes-wrap-modulo-256)
- [atago self-hosting / explain](#atago-self-hosting--explain) — 1 scenario
  - [explain summarizes a spec without running it](#scenario-explain-summarizes-a-spec-without-running-it)
- [atago self-hosting / file equals and equals_file byte-equality (#155)](#atago-self-hosting--file-equals-and-equals_file-byte-equality-155) — 4 scenarios
  - [equals_file passes for two byte-identical files](#scenario-equals_file-passes-for-two-byte-identical-files)
  - [equals matches an inline literal byte-for-byte](#scenario-equals-matches-an-inline-literal-byte-for-byte)
  - [equals_file fails the inner spec when the two files differ by one byte](#scenario-equals_file-fails-the-inner-spec-when-the-two-files-differ-by-one-byte)
  - [equals_file is byte-exact — a CRLF vs LF difference fails](#scenario-equals_file-is-byte-exact--a-crlf-vs-lf-difference-fails)
- [atago self-hosting / fixture from (copy committed testdata)](#atago-self-hosting--fixture-from-copy-committed-testdata) — 2 scenarios
  - [a committed binary blob is copied verbatim into the workdir](#scenario-a-committed-binary-blob-is-copied-verbatim-into-the-workdir)
  - [copying from a missing source errors the scenario](#scenario-copying-from-a-missing-source-errors-the-scenario)
- [atago self-hosting / fixture symlink+mode+mtime, file executable, env skip](#atago-self-hosting--fixture-symlinkmodemtime-file-executable-env-skip) — 5 scenarios
  - [a symlink fixture resolves to its target](#scenario-a-symlink-fixture-resolves-to-its-target)
  - [fixture.mode sets permissions and file.executable reads them](#scenario-fixturemode-sets-permissions-and-fileexecutable-reads-them)
  - [fixture.mtime pins the modification time](#scenario-fixturemtime-pins-the-modification-time)
  - [only.env skips when the variable is unset](#scenario-onlyenv-skips-when-the-variable-is-unset)
  - [skip.env runs when the variable is unset](#scenario-skipenv-runs-when-the-variable-is-unset)
- [atago self-hosting / flaky tooling (--repeat, --retry-failed)](#atago-self-hosting--flaky-tooling---repeat---retry-failed) — 3 scenarios
  - [retry-failed recovers a flaky scenario and reports it loudly](#scenario-retry-failed-recovers-a-flaky-scenario-and-reports-it-loudly)
  - [repeat surfaces flakiness that a single run would miss](#scenario-repeat-surfaces-flakiness-that-a-single-run-would-miss)
  - [repeat and retry-failed are mutually exclusive](#scenario-repeat-and-retry-failed-are-mutually-exclusive)
- [atago self-hosting / grpc runner](#atago-self-hosting--grpc-runner) — 2 scenarios
  - [a grpc runner without a target fails validation (exit 2)](#scenario-a-grpc-runner-without-a-target-fails-validation-exit-2)
  - [a grpc step naming an undeclared runner fails validation (exit 2)](#scenario-a-grpc-step-naming-an-undeclared-runner-fails-validation-exit-2)
- [atago self-hosting / hermetic environment (clear_env + pass_env)](#atago-self-hosting--hermetic-environment-clear_env--pass_env) — 5 scenarios
  - [clear_env drops inherited host variables](#scenario-clear_env-drops-inherited-host-variables)
  - [pass_env re-admits an allowlist of host variables](#scenario-pass_env-re-admits-an-allowlist-of-host-variables)
  - [explicit env wins over a passed-through host variable](#scenario-explicit-env-wins-over-a-passed-through-host-variable)
  - [pass_env without clear_env is a load-time error](#scenario-pass_env-without-clear_env-is-a-load-time-error)
  - [unset host variables in pass_env are skipped, not an error](#scenario-unset-host-variables-in-pass_env-are-skipped-not-an-error)
- [atago self-hosting / http runner](#atago-self-hosting--http-runner) — 2 scenarios
  - [a denied host is a security policy violation (exit 6)](#scenario-a-denied-host-is-a-security-policy-violation-exit-6)
  - [an http step with an undeclared runner fails validation (exit 2)](#scenario-an-http-step-with-an-undeclared-runner-fails-validation-exit-2)
- [atago self-hosting / image](#atago-self-hosting--image) — 4 scenarios
  - [format, dimension and alpha assertions pass on a PNG](#scenario-format-dimension-and-alpha-assertions-pass-on-a-png)
  - [a pixel comparison against an identical baseline passes](#scenario-a-pixel-comparison-against-an-identical-baseline-passes)
  - [a wrong dimension assertion fails with a clear diff](#scenario-a-wrong-dimension-assertion-fails-with-a-clear-diff)
  - [a failing similar_to writes visual diff artifacts](#scenario-a-failing-similar_to-writes-visual-diff-artifacts)
- [atago self-hosting / init](#atago-self-hosting--init) — 3 scenarios
  - [init scaffolds a runnable spec](#scenario-init-scaffolds-a-runnable-spec)
  - [init emits a resolvable schema header for editor completion](#scenario-init-emits-a-resolvable-schema-header-for-editor-completion)
  - [init refuses to overwrite without --force](#scenario-init-refuses-to-overwrite-without---force)
- [atago self-hosting / init templates](#atago-self-hosting--init-templates) — 12 scenarios
  - [every template scaffolds a schema-valid spec \[template=cli\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templatecli)
  - [every template scaffolds a schema-valid spec \[template=http\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templatehttp)
  - [every template scaffolds a schema-valid spec \[template=db\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templatedb)
  - [every template scaffolds a schema-valid spec \[template=grpc\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templategrpc)
  - [every template scaffolds a schema-valid spec \[template=ssh\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templatessh)
  - [every template scaffolds a schema-valid spec \[template=browser\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templatebrowser)
  - [every template scaffolds a schema-valid spec \[template=services\]](#scenario-every-template-scaffolds-a-schema-valid-spec-templateservices)
  - [list-templates names every runner family with a description](#scenario-list-templates-names-every-runner-family-with-a-description)
  - [unknown template is a configuration error](#scenario-unknown-template-is-a-configuration-error)
  - [the default cli template runs green](#scenario-the-default-cli-template-runs-green)
  - [the db template runs green with the bundled sqlite driver](#scenario-the-db-template-runs-green-with-the-bundled-sqlite-driver)
  - [the services template runs green and exercises readiness + retry](#scenario-the-services-template-runs-green-and-exercises-readiness--retry)
- [atago self-hosting / json numeric comparators](#atago-self-hosting--json-numeric-comparators) — 6 scenarios
  - [gt and gte pass on a value at or above the bound](#scenario-gt-and-gte-pass-on-a-value-at-or-above-the-bound)
  - [lt and lte pass on a value at or below the bound](#scenario-lt-and-lte-pass-on-a-value-at-or-below-the-bound)
  - [comparators work on a numeric string](#scenario-comparators-work-on-a-numeric-string)
  - [comparators apply to rows and file json targets too](#scenario-comparators-apply-to-rows-and-file-json-targets-too)
  - [a value below the gt bound fails the inner spec](#scenario-a-value-below-the-gt-bound-fails-the-inner-spec)
  - [a non-numeric value cannot be compared and fails](#scenario-a-non-numeric-value-cannot-be-compared-and-fails)
- [atago self-hosting / json and yaml matcher lists (#156)](#atago-self-hosting--json-and-yaml-matcher-lists-156) — 5 scenarios
  - [a file json list asserts several paths at once](#scenario-a-file-json-list-asserts-several-paths-at-once)
  - [a single mapping still works (backward compatible)](#scenario-a-single-mapping-still-works-backward-compatible)
  - [a json list fails the inner spec when one listed path mismatches](#scenario-a-json-list-fails-the-inner-spec-when-one-listed-path-mismatches)
  - [a stdout json list against a JSON-producing command](#scenario-a-stdout-json-list-against-a-json-producing-command)
  - [a yaml list asserts several paths on one document](#scenario-a-yaml-list-asserts-several-paths-on-one-document)
- [atago self-hosting / json matcher boundary values](#atago-self-hosting--json-matcher-boundary-values) — 9 scenarios
  - [an array element is addressable by index](#scenario-an-array-element-is-addressable-by-index)
  - [a top-level array reports its length](#scenario-a-top-level-array-reports-its-length)
  - [an empty array has length zero](#scenario-an-empty-array-has-length-zero)
  - [the numeric comparators bound a value](#scenario-the-numeric-comparators-bound-a-value)
  - [a boolean value compares equal](#scenario-a-boolean-value-compares-equal)
  - [a floating-point value compares equal](#scenario-a-floating-point-value-compares-equal)
  - [a string carrying a quote compares equal](#scenario-a-string-carrying-a-quote-compares-equal)
  - [a deeply nested path resolves](#scenario-a-deeply-nested-path-resolves)
  - [a path that selects nothing fails with a clear message](#scenario-a-path-that-selects-nothing-fails-with-a-clear-message)
- [atago self-hosting / line selector](#atago-self-hosting--line-selector) — 3 scenarios
  - [line selector narrows stdout to a single 1-based line](#scenario-line-selector-narrows-stdout-to-a-single-1-based-line)
  - [a trailing newline does not add a phantom final line](#scenario-a-trailing-newline-does-not-add-a-phantom-final-line)
  - [an out-of-range line fails the inner spec](#scenario-an-out-of-range-line-fails-the-inner-spec)
- [atago self-hosting / stream text matchers fold CRLF](#atago-self-hosting--stream-text-matchers-fold-crlf) — 12 scenarios
  - [equals folds a CRLF body to its LF form](#scenario-equals-folds-a-crlf-body-to-its-lf-form)
  - [equals tolerates the phantom trailing CRLF](#scenario-equals-tolerates-the-phantom-trailing-crlf)
  - [contains folds CRLF for a multi-line needle](#scenario-contains-folds-crlf-for-a-multi-line-needle)
  - [contains authored with CRLF matches LF-folded output](#scenario-contains-authored-with-crlf-matches-lf-folded-output)
  - [contains list every multi-line element folds](#scenario-contains-list-every-multi-line-element-folds)
  - [matches anchors a line over CRLF with the multiline flag](#scenario-matches-anchors-a-line-over-crlf-with-the-multiline-flag)
  - [matches a literal newline in the pattern over CRLF](#scenario-matches-a-literal-newline-in-the-pattern-over-crlf)
  - [not_contains stays clear of an absent multi-line needle](#scenario-not_contains-stays-clear-of-an-absent-multi-line-needle)
  - [not_matches passes for an anchored line that is absent](#scenario-not_matches-passes-for-an-anchored-line-that-is-absent)
  - [the line selector strips the trailing CR](#scenario-the-line-selector-strips-the-trailing-cr)
  - [json parses a CRLF-formatted document](#scenario-json-parses-a-crlf-formatted-document)
  - [folding does not make an absent multi-line needle match](#scenario-folding-does-not-make-an-absent-multi-line-needle-match)
- [atago self-hosting / list](#atago-self-hosting--list) — 2 scenarios
  - [list surfaces suites, scenarios, tags, and gates](#scenario-list-surfaces-suites-scenarios-tags-and-gates)
  - [list --json is a stable machine contract](#scenario-list---json-is-a-stable-machine-contract)
- [atago self-hosting / loader rejects malformed specs](#atago-self-hosting--loader-rejects-malformed-specs) — 14 scenarios
  - [an empty scenario list is rejected](#scenario-an-empty-scenario-list-is-rejected)
  - [a wrong version string is rejected](#scenario-a-wrong-version-string-is-rejected)
  - [an unknown top-level field is rejected with its position](#scenario-an-unknown-top-level-field-is-rejected-with-its-position)
  - [a step that sets two actions is rejected](#scenario-a-step-that-sets-two-actions-is-rejected)
  - [a stream assertion with no matcher is rejected](#scenario-a-stream-assertion-with-no-matcher-is-rejected)
  - [combining equals with another matcher is rejected](#scenario-combining-equals-with-another-matcher-is-rejected)
  - [a line index below one is rejected](#scenario-a-line-index-below-one-is-rejected)
  - [combining a line selector with json is rejected](#scenario-combining-a-line-selector-with-json-is-rejected)
  - [a duplicate scenario name is rejected](#scenario-a-duplicate-scenario-name-is-rejected)
  - [an empty run command is rejected](#scenario-an-empty-run-command-is-rejected)
  - [an unparseable timeout is rejected with an example](#scenario-an-unparseable-timeout-is-rejected-with-an-example)
  - [a fixture with two content sources is rejected](#scenario-a-fixture-with-two-content-sources-is-rejected)
  - [an absolute changes glob is rejected as not workdir-relative](#scenario-an-absolute-changes-glob-is-rejected-as-not-workdir-relative)
  - [the inline stdin form is a scalar, not a mapping key](#scenario-the-inline-stdin-form-is-a-scalar-not-a-mapping-key)
- [atago self-hosting / manifest](#atago-self-hosting--manifest) — 2 scenarios
  - [manifest emits a stable JSON summary without running the spec](#scenario-manifest-emits-a-stable-json-summary-without-running-the-spec)
  - [manifest does not execute the spec's commands](#scenario-manifest-does-not-execute-the-specs-commands)
- [atago self-hosting / matrix scenarios](#atago-self-hosting--matrix-scenarios) — 4 scenarios
  - [matrix expands into one scenario per row](#scenario-matrix-expands-into-one-scenario-per-row)
  - [matrix without a templated name gets a deterministic suffix](#scenario-matrix-without-a-templated-name-gets-a-deterministic-suffix)
  - [stdout_to expands a matrix variable into the redirect target \[who=alice\]](#scenario-stdout_to-expands-a-matrix-variable-into-the-redirect-target-whoalice)
  - [stdout_to expands a matrix variable into the redirect target \[who=bob\]](#scenario-stdout_to-expands-a-matrix-variable-into-the-redirect-target-whobob)
- [atago self-hosting / matrix expansion boundary values](#atago-self-hosting--matrix-expansion-boundary-values) — 5 scenarios
  - [each row substitutes into the scenario name](#scenario-each-row-substitutes-into-the-scenario-name)
  - [a row with several variables substitutes all of them](#scenario-a-row-with-several-variables-substitutes-all-of-them)
  - [a single-row matrix expands to exactly one scenario](#scenario-a-single-row-matrix-expands-to-exactly-one-scenario)
  - [an empty matrix row list is a load-time error](#scenario-an-empty-matrix-row-list-is-a-load-time-error)
  - [rows that expand to the same name are rejected as duplicates](#scenario-rows-that-expand-to-the-same-name-are-rejected-as-duplicates)
- [atago self-hosting / mock http server (offline API-client testing)](#atago-self-hosting--mock-http-server-offline-api-client-testing) — 3 scenarios
  - [count, header, and body-json asserts pass against a real client](#scenario-count-header-and-body-json-asserts-pass-against-a-real-client)
  - [a failing count summarizes the recorded requests](#scenario-a-failing-count-summarizes-the-recorded-requests)
  - [an unknown mock name in an assert is a load-time error](#scenario-an-unknown-mock-name-in-an-assert-is-a-load-time-error)
- [atago self-hosting / combined stream matchers](#atago-self-hosting--combined-stream-matchers) — 6 scenarios
  - [contains and not_contains hold together](#scenario-contains-and-not_contains-hold-together)
  - [matches and not_matches hold together](#scenario-matches-and-not_matches-hold-together)
  - [all four text matchers compose](#scenario-all-four-text-matchers-compose)
  - [a combined matcher composes with a line selector](#scenario-a-combined-matcher-composes-with-a-line-selector)
  - [a failing member fails the inner spec and names the offender](#scenario-a-failing-member-fails-the-inner-spec-and-names-the-offender)
  - [mixing a whole-stream matcher with a text matcher is a load error](#scenario-mixing-a-whole-stream-matcher-with-a-text-matcher-is-a-load-error)
- [atago self-hosting / not_equals matcher](#atago-self-hosting--not_equals-matcher) — 4 scenarios
  - [not_equals passes when stdout differs from the given text](#scenario-not_equals-passes-when-stdout-differs-from-the-given-text)
  - [not_equals is trailing-newline tolerant like equals](#scenario-not_equals-is-trailing-newline-tolerant-like-equals)
  - [not_equals composes with a line selector](#scenario-not_equals-composes-with-a-line-selector)
  - [not_equals fails the inner spec when the text matches exactly](#scenario-not_equals-fails-the-inner-spec-when-the-text-matches-exactly)
- [atago self-hosting / parallel](#atago-self-hosting--parallel) — 2 scenarios
  - [parallel run passes and stays deterministic](#scenario-parallel-run-passes-and-stays-deterministic)
  - [fail-fast stops after the first failure](#scenario-fail-fast-stops-after-the-first-failure)
- [atago self-hosting / forward-slash spec paths resolve on every OS](#atago-self-hosting--forward-slash-spec-paths-resolve-on-every-os) — 7 scenarios
  - [stdout_to creates a nested parent directory](#scenario-stdout_to-creates-a-nested-parent-directory)
  - [stderr_to creates its own nested parent directory](#scenario-stderr_to-creates-its-own-nested-parent-directory)
  - [a fixture at a nested forward-slash path is created and addressable](#scenario-a-fixture-at-a-nested-forward-slash-path-is-created-and-addressable)
  - [a file assert reaches a deeply nested fixture by forward-slash path](#scenario-a-file-assert-reaches-a-deeply-nested-fixture-by-forward-slash-path)
  - [a dir assert addresses a nested tree and child by forward-slash path](#scenario-a-dir-assert-addresses-a-nested-tree-and-child-by-forward-slash-path)
  - [equals_file compares two files addressed by forward-slash paths](#scenario-equals_file-compares-two-files-addressed-by-forward-slash-paths)
  - [a redirect path may not escape the workdir via a nested traversal](#scenario-a-redirect-path-may-not-escape-the-workdir-via-a-nested-traversal)
- [atago self-hosting / pdf assertion](#atago-self-hosting--pdf-assertion) — 2 scenarios
  - [pdf assertions cover page count, metadata, and text](#scenario-pdf-assertions-cover-page-count-metadata-and-text)
  - [a non-pdf file fails the pdf target](#scenario-a-non-pdf-file-fails-the-pdf-target)
- [atago self-hosting / pty](#atago-self-hosting--pty) — 8 scenarios
  - [a pty step sees a terminal where a run step sees a pipe](#scenario-a-pty-step-sees-a-terminal-where-a-run-step-sees-a-pipe)
  - [a never-matching expect fails with the pattern in the block](#scenario-a-never-matching-expect-fails-with-the-pattern-in-the-block)
  - [named keys transmit their documented bytes and ctrl-c aborts](#scenario-named-keys-transmit-their-documented-bytes-and-ctrl-c-aborts)
  - [an unknown key name is a load-time error listing the vocabulary](#scenario-an-unknown-key-name-is-a-load-time-error-listing-the-vocabulary)
  - [screen asserts see the final frame where the transcript sees history](#scenario-screen-asserts-see-the-final-frame-where-the-transcript-sees-history)
  - [a screen snapshot round-trips through update and compare](#scenario-a-screen-snapshot-round-trips-through-update-and-compare)
  - [a screen assert without a pty step is a load-time error](#scenario-a-screen-assert-without-a-pty-step-is-a-load-time-error)
  - [a send referencing an undefined variable is an execution error, not typed literally](#scenario-a-send-referencing-an-undefined-variable-is-an-execution-error-not-typed-literally)
- [atago self-hosting / pty (portable)](#atago-self-hosting--pty-portable) — 8 scenarios
  - [a pty step starts a command, captures its output, and reports exit 0](#scenario-a-pty-step-starts-a-command-captures-its-output-and-reports-exit-0)
  - [a pty step surfaces a command's non-zero exit code](#scenario-a-pty-step-surfaces-a-commands-non-zero-exit-code)
  - [sequential expects match successive output in declaration order](#scenario-sequential-expects-match-successive-output-in-declaration-order)
  - [an expect pattern is a regular expression, not a literal](#scenario-an-expect-pattern-is-a-regular-expression-not-a-literal)
  - [a screen assert reads the rendered frame sized by rows and cols](#scenario-a-screen-assert-reads-the-rendered-frame-sized-by-rows-and-cols)
  - [a pty step drives the atago binary directly with no shell](#scenario-a-pty-step-drives-the-atago-binary-directly-with-no-shell)
  - [a pty drives atago running an inner spec to a green result](#scenario-a-pty-drives-atago-running-an-inner-spec-to-a-green-result)
  - [a never-matching expect fails and names the pattern in the transcript](#scenario-a-never-matching-expect-fails-and-names-the-pattern-in-the-transcript)
- [atago self-hosting / record (spec skeleton from an observed run)](#atago-self-hosting--record-spec-skeleton-from-an-observed-run) — 14 scenarios
  - [record then run round-trips green](#scenario-record-then-run-round-trips-green)
  - [refusing to overwrite without --force](#scenario-refusing-to-overwrite-without---force)
  - [record --pty refuses an existing --out before driving the session](#scenario-record---pty-refuses-an-existing---out-before-driving-the-session)
  - [created files become exists asserts (shell mode)](#scenario-created-files-become-exists-asserts-shell-mode)
  - [snapshot mode writes a golden the run then matches](#scenario-snapshot-mode-writes-a-golden-the-run-then-matches)
  - [no command is a usage error](#scenario-no-command-is-a-usage-error)
  - [argv boundaries survive spaced arguments](#scenario-argv-boundaries-survive-spaced-arguments)
  - [a shell metacharacter argument stays one token](#scenario-a-shell-metacharacter-argument-stays-one-token)
  - [record --pty records a live session and the generated spec replays green](#scenario-record---pty-records-a-live-session-and-the-generated-spec-replays-green)
  - [record --pty of a no-input command yields a session-less spec that replays green](#scenario-record---pty-of-a-no-input-command-yields-a-session-less-spec-that-replays-green)
  - [a prompt with regex metacharacters is escaped in the generated expect](#scenario-a-prompt-with-regex-metacharacters-is-escaped-in-the-generated-expect)
  - [recorded text containing dollar-brace round-trips as literal text](#scenario-recorded-text-containing-dollar-brace-round-trips-as-literal-text)
  - [a recorded secret placeholder replays green with the env set and is guarded when unset](#scenario-a-recorded-secret-placeholder-replays-green-with-the-env-set-and-is-guarded-when-unset)
  - [record --pty of a never-exiting program times out instead of hanging](#scenario-record---pty-of-a-never-exiting-program-times-out-instead-of-hanging)
- [atago self-hosting / report formats agree on outcomes](#atago-self-hosting--report-formats-agree-on-outcomes) — 7 scenarios
  - [json report carries per-scenario verdicts and a failures array](#scenario-json-report-carries-per-scenario-verdicts-and-a-failures-array)
  - [junit report tallies tests, failures, skipped, and errors](#scenario-junit-report-tallies-tests-failures-skipped-and-errors)
  - [tap report emits the plan, a not ok line, and a SKIP directive](#scenario-tap-report-emits-the-plan-a-not-ok-line-and-a-skip-directive)
  - [gha report annotates the failure and summarizes the counts](#scenario-gha-report-annotates-the-failure-and-summarizes-the-counts)
  - [console report prints the same counts in its summary line](#scenario-console-report-prints-the-same-counts-in-its-summary-line)
  - [an all-passing run reports a zero-failure suite and exits zero](#scenario-an-all-passing-run-reports-a-zero-failure-suite-and-exits-zero)
  - [an errored step is counted as an error, not a failure, across formats](#scenario-an-errored-step-is-counted-as-an-error-not-a-failure-across-formats)
- [atago self-hosting / reports](#atago-self-hosting--reports) — 5 scenarios
  - [JUnit report is XML with a testsuite and testcase](#scenario-junit-report-is-xml-with-a-testsuite-and-testcase)
  - [GitHub Actions annotations are emitted on failure](#scenario-github-actions-annotations-are-emitted-on-failure)
  - [TAP report is a numbered TAP 13 stream with ok / not ok points](#scenario-tap-report-is-a-numbered-tap-13-stream-with-ok--not-ok-points)
  - [failure artifacts are written and referenced in the JSON report](#scenario-failure-artifacts-are-written-and-referenced-in-the-json-report)
  - [a multi-line snapshot failure renders a unified diff with hunks](#scenario-a-multi-line-snapshot-failure-renders-a-unified-diff-with-hunks)
- [atago self-hosting / rerun-failed](#atago-self-hosting--rerun-failed) — 3 scenarios
  - [a failing run is recorded and rerun-failed selects only it](#scenario-a-failing-run-is-recorded-and-rerun-failed-selects-only-it)
  - [rerun-failed with nothing recorded is a no-op success](#scenario-rerun-failed-with-nothing-recorded-is-a-no-op-success)
  - [rerun-failed with a filter preserves the still-failing scenarios it did not run](#scenario-rerun-failed-with-a-filter-preserves-the-still-failing-scenarios-it-did-not-run)
- [atago self-hosting / retry until](#atago-self-hosting--retry-until) — 3 scenarios
  - [retry polls until the condition becomes true](#scenario-retry-polls-until-the-condition-becomes-true)
  - [retry fails the inner spec when until never holds](#scenario-retry-fails-the-inner-spec-when-until-never-holds)
  - [until with a changes target is a load-time error](#scenario-until-with-a-changes-target-is-a-load-time-error)
- [atago self-hosting / run](#atago-self-hosting--run) — 8 scenarios
  - [a passing spec exits zero and reports PASS](#scenario-a-passing-spec-exits-zero-and-reports-pass)
  - [a failing assertion exits one and reports the failure](#scenario-a-failing-assertion-exits-one-and-reports-the-failure)
  - [an exit_code failure surfaces the command's stderr](#scenario-an-exit_code-failure-surfaces-the-commands-stderr)
  - [an exit_code failure surfaces the command's stderr (windows)](#scenario-an-exit_code-failure-surfaces-the-commands-stderr-windows)
  - [an exit_code failure falls back to stdout when stderr is silent](#scenario-an-exit_code-failure-falls-back-to-stdout-when-stderr-is-silent)
  - [an exit_code failure falls back to stdout when stderr is silent (windows)](#scenario-an-exit_code-failure-falls-back-to-stdout-when-stderr-is-silent-windows)
  - [a parse error exits with code two](#scenario-a-parse-error-exits-with-code-two)
  - [JSON report is valid JSON with a passed status](#scenario-json-report-is-valid-json-with-a-passed-status)
- [atago self-hosting / sandbox_home (isolated per-OS home)](#atago-self-hosting--sandbox_home-isolated-per-os-home) — 3 scenarios
  - [Unix XDG family — write config, read it back, inspect it under the workdir](#scenario-unix-xdg-family--write-config-read-it-back-inspect-it-under-the-workdir)
  - [Windows APPDATA family — write config, read it back, inspect it under the workdir](#scenario-windows-appdata-family--write-config-read-it-back-inspect-it-under-the-workdir)
  - [cwd anchors the run, but sandbox_home stays at the workdir ROOT (Unix)](#scenario-cwd-anchors-the-run-but-sandbox_home-stays-at-the-workdir-root-unix)
- [atago self-hosting / security](#atago-self-hosting--security) — 3 scenarios
  - [declared secrets are masked in failure output](#scenario-declared-secrets-are-masked-in-failure-output)
  - [a file assertion path may not escape the scenario workdir](#scenario-a-file-assertion-path-may-not-escape-the-scenario-workdir)
  - [a snapshot path may not escape the spec directory](#scenario-a-snapshot-path-may-not-escape-the-spec-directory)
- [atago self-hosting / selection](#atago-self-hosting--selection) — 3 scenarios
  - [--filter runs only matching scenarios](#scenario---filter-runs-only-matching-scenarios)
  - [--filter selects multiple scenarios with OR (comma and repeated)](#scenario---filter-selects-multiple-scenarios-with-or-comma-and-repeated)
  - [--skip-tag drops tagged scenarios](#scenario---skip-tag-drops-tagged-scenarios)
- [atago self-hosting / background services](#atago-self-hosting--background-services) — 7 scenarios
  - [file readiness captures a dynamic value into a variable](#scenario-file-readiness-captures-a-dynamic-value-into-a-variable)
  - [log readiness waits for a line on the service output](#scenario-log-readiness-waits-for-a-line-on-the-service-output)
  - [delay readiness waits a fixed duration](#scenario-delay-readiness-waits-a-fixed-duration)
  - [multiple services start and capture independently](#scenario-multiple-services-start-and-capture-independently)
  - [a readiness failure preserves the service log as an artifact](#scenario-a-readiness-failure-preserves-the-service-log-as-an-artifact)
  - [a step failure after the service is ready preserves the service log](#scenario-a-step-failure-after-the-service-is-ready-preserves-the-service-log)
  - [a green run with a healthy service writes no service log](#scenario-a-green-run-with-a-healthy-service-writes-no-service-log)
- [atago self-hosting / harness shell is not shadowed by the program PATH](#atago-self-hosting--harness-shell-is-not-shadowed-by-the-program-path) — 2 scenarios
  - [a PATH-resident fake sh does not hijack shell:true](#scenario-a-path-resident-fake-sh-does-not-hijack-shelltrue)
  - [ATAGO_SHELL overrides the shell used for shell:true](#scenario-atago_shell-overrides-the-shell-used-for-shelltrue)
- [atago self-hosting / signal step (graceful shutdown)](#atago-self-hosting--signal-step-graceful-shutdown) — 4 scenarios
  - [SIGTERM reaches the trap handler and wait observes the exit](#scenario-sigterm-reaches-the-trap-handler-and-wait-observes-the-exit)
  - [SIGHUP triggers a reload without stopping the service](#scenario-sighup-triggers-a-reload-without-stopping-the-service)
  - [a wait timeout on a TERM-ignoring service fails with the documented message](#scenario-a-wait-timeout-on-a-term-ignoring-service-fails-with-the-documented-message)
  - [an unknown target service is a load-time error listing declared names](#scenario-an-unknown-target-service-is-a-load-time-error-listing-declared-names)
- [atago self-hosting / skip-only command predicate](#atago-self-hosting--skip-only-command-predicate) — 3 scenarios
  - [skip command that succeeds skips the scenario](#scenario-skip-command-that-succeeds-skips-the-scenario)
  - [only command that fails skips the scenario](#scenario-only-command-that-fails-skips-the-scenario)
  - [only command that succeeds runs the scenario](#scenario-only-command-that-succeeds-runs-the-scenario)
- [atago self-hosting / snapshot](#atago-self-hosting--snapshot) — 3 scenarios
  - [a snapshot assertion passes against a committed snapshot](#scenario-a-snapshot-assertion-passes-against-a-committed-snapshot)
  - [snapshot update creates the snapshot file](#scenario-snapshot-update-creates-the-snapshot-file)
  - [a snapshot mismatch writes the normalized actual as an artifact](#scenario-a-snapshot-mismatch-writes-the-normalized-actual-as-an-artifact)
- [atago self-hosting / snapshot normalization and round-trip](#atago-self-hosting--snapshot-normalization-and-round-trip) — 9 scenarios
  - [record then run round-trips green](#scenario-record-then-run-round-trips-green-1)
  - [a UUID is masked in the golden](#scenario-a-uuid-is-masked-in-the-golden)
  - [an ISO timestamp is masked in the golden](#scenario-an-iso-timestamp-is-masked-in-the-golden)
  - [a loopback host and port are masked in the golden](#scenario-a-loopback-host-and-port-are-masked-in-the-golden)
  - [the home directory is masked to a tilde in the golden](#scenario-the-home-directory-is-masked-to-a-tilde-in-the-golden)
  - [a golden verifies against a different volatile value](#scenario-a-golden-verifies-against-a-different-volatile-value)
  - [updating a snapshot is deterministic](#scenario-updating-a-snapshot-is-deterministic)
  - [a real content change still fails the snapshot](#scenario-a-real-content-change-still-fails-the-snapshot)
  - [a missing golden names the update flag](#scenario-a-missing-golden-names-the-update-flag)
- [atago self-hosting / ssh runner](#atago-self-hosting--ssh-runner) — 3 scenarios
  - [an ssh runner without host/user fails validation (exit 2)](#scenario-an-ssh-runner-without-hostuser-fails-validation-exit-2)
  - [a run step naming an undeclared runner fails validation (exit 2)](#scenario-a-run-step-naming-an-undeclared-runner-fails-validation-exit-2)
  - [a local-only run field on an ssh runner fails validation (exit 2)](#scenario-a-local-only-run-field-on-an-ssh-runner-fails-validation-exit-2)
- [atago self-hosting / stdin sources (file + base64)](#atago-self-hosting--stdin-sources-file--base64) — 5 scenarios
  - [base64 stdin delivers the exact byte count](#scenario-base64-stdin-delivers-the-exact-byte-count)
  - [stdin file is expanded and read from the workdir](#scenario-stdin-file-is-expanded-and-read-from-the-workdir)
  - [a stdin file outside the workdir is rejected at runtime](#scenario-a-stdin-file-outside-the-workdir-is-rejected-at-runtime)
  - [stdin with both file and base64 is a load-time error](#scenario-stdin-with-both-file-and-base64-is-a-load-time-error)
  - [invalid base64 stdin is a load-time error](#scenario-invalid-base64-stdin-is-a-load-time-error)
- [atago self-hosting / store](#atago-self-hosting--store) — 2 scenarios
  - [a stored JSON value is reusable in later commands](#scenario-a-stored-json-value-is-reusable-in-later-commands)
  - [storing from a missing JSON path is an execution error](#scenario-storing-from-a-missing-json-path-is-an-execution-error)
- [atago self-hosting / store capture boundary values](#atago-self-hosting--store-capture-boundary-values) — 6 scenarios
  - [a regex with a capture group stores the group](#scenario-a-regex-with-a-capture-group-stores-the-group)
  - [a regex without a group stores the whole match](#scenario-a-regex-without-a-group-stores-the-whole-match)
  - [a JSON path captures a scalar from stdout](#scenario-a-json-path-captures-a-scalar-from-stdout)
  - [a JSON path captures a value from a generated file](#scenario-a-json-path-captures-a-value-from-a-generated-file)
  - [a regex that matches nothing is an execution error](#scenario-a-regex-that-matches-nothing-is-an-execution-error)
  - [a stored value does not leak into the next scenario](#scenario-a-stored-value-does-not-leak-into-the-next-scenario)
- [atago self-hosting / store whole-content trim and text selectors (#158)](#atago-self-hosting--store-whole-content-trim-and-text-selectors-158) — 2 scenarios
  - [trim captures an opaque token and round-trips it as an argument](#scenario-trim-captures-an-opaque-token-and-round-trips-it-as-an-argument)
  - [text captures a whole multi-line file verbatim](#scenario-text-captures-a-whole-multi-line-file-verbatim)
- [atago self-hosting / stream matcher boundary values](#atago-self-hosting--stream-matcher-boundary-values) — 18 scenarios
  - [equals a multibyte and emoji line](#scenario-equals-a-multibyte-and-emoji-line)
  - [contains a multibyte substring inside a longer line](#scenario-contains-a-multibyte-substring-inside-a-longer-line)
  - [a regex matches across multibyte runes](#scenario-a-regex-matches-across-multibyte-runes)
  - [line selection returns a multibyte line intact](#scenario-line-selection-returns-a-multibyte-line-intact)
  - [not_contains a multibyte needle that is absent](#scenario-not_contains-a-multibyte-needle-that-is-absent)
  - [empty is true for a command that prints nothing](#scenario-empty-is-true-for-a-command-that-prints-nothing)
  - [empty is true for whitespace-only output](#scenario-empty-is-true-for-whitespace-only-output)
  - [equals tolerates output with no trailing newline](#scenario-equals-tolerates-output-with-no-trailing-newline)
  - [a deliberate trailing blank line is addressable by index](#scenario-a-deliberate-trailing-blank-line-is-addressable-by-index)
  - [contains treats a needle with regex metacharacters literally](#scenario-contains-treats-a-needle-with-regex-metacharacters-literally)
  - [matches requires escaping a literal metacharacter](#scenario-matches-requires-escaping-a-literal-metacharacter)
  - [not_matches passes when an unescaped-metachar pattern does not match](#scenario-not_matches-passes-when-an-unescaped-metachar-pattern-does-not-match)
  - [a tab-separated record contains the exact tab byte](#scenario-a-tab-separated-record-contains-the-exact-tab-byte)
  - [quotes and brackets survive an exact equals](#scenario-quotes-and-brackets-survive-an-exact-equals)
  - [the last of many lines is selectable by index](#scenario-the-last-of-many-lines-is-selectable-by-index)
  - [a line selector composes with contains](#scenario-a-line-selector-composes-with-contains)
  - [a line selector composes with a regex](#scenario-a-line-selector-composes-with-a-regex)
  - [stderr carries the same matcher semantics as stdout](#scenario-stderr-carries-the-same-matcher-semantics-as-stdout)
- [atago self-hosting / suite setup](#atago-self-hosting--suite-setup) — 4 scenarios
  - [setup runs once, shares stores and env, and teardown always runs](#scenario-setup-runs-once-shares-stores-and-env-and-teardown-always-runs)
  - [a failing setup errors every scenario and none runs (exit 4)](#scenario-a-failing-setup-errors-every-scenario-and-none-runs-exit-4)
  - [a suite service starts once and its store reaches every scenario](#scenario-a-suite-service-starts-once-and-its-store-reaches-every-scenario)
  - [a failing suite teardown is loud but does not flip the verdict](#scenario-a-failing-suite-teardown-is-loud-but-does-not-flip-the-verdict)
- [atago self-hosting / step timeouts (suite default + escape hatch)](#atago-self-hosting--step-timeouts-suite-default--escape-hatch) — 4 scenarios
  - [suite.timeout kills a hanging step and the hint names it](#scenario-suitetimeout-kills-a-hanging-step-and-the-hint-names-it)
  - [a step timeout beats the suite timeout and the hint says run.timeout](#scenario-a-step-timeout-beats-the-suite-timeout-and-the-hint-says-runtimeout)
  - [timeout zero disables a short suite bound](#scenario-timeout-zero-disables-a-short-suite-bound)
  - [an invalid suite.timeout is a load-time error](#scenario-an-invalid-suitetimeout-is-a-load-time-error)
- [atago self-hosting / tui](#atago-self-hosting--tui) — 4 scenarios
  - [a pty step exports a usable TERM by default](#scenario-a-pty-step-exports-a-usable-term-by-default)
  - [an explicit TERM overrides the default](#scenario-an-explicit-term-overrides-the-default)
  - [an expect does not re-match a consumed pattern](#scenario-an-expect-does-not-re-match-a-consumed-pattern)
  - [less -X renders a real pager onto the screen](#scenario-less--x-renders-a-real-pager-onto-the-screen)
- [atago self-hosting / variable resolution semantics](#atago-self-hosting--variable-resolution-semantics) — 7 scenarios
  - [a doubled dollar keeps the braces literal](#scenario-a-doubled-dollar-keeps-the-braces-literal)
  - [the workdir builtin expands to the scenario directory](#scenario-the-workdir-builtin-expands-to-the-scenario-directory)
  - [the atago builtin resolves to the binary under test](#scenario-the-atago-builtin-resolves-to-the-binary-under-test)
  - [an env reference expands from the host environment](#scenario-an-env-reference-expands-from-the-host-environment)
  - [shell true defers an unknown reference to the shell](#scenario-shell-true-defers-an-unknown-reference-to-the-shell)
  - [an unresolved variable is a hard error, not a silent empty](#scenario-an-unresolved-variable-is-a-hard-error-not-a-silent-empty)
  - [an unset env reference names the missing variable](#scenario-an-unset-env-reference-names-the-missing-variable)
- [atago self-hosting / verbose](#atago-self-hosting--verbose) — 4 scenarios
  - [verbose shows a passing scenario's command, output, and verdicts](#scenario-verbose-shows-a-passing-scenarios-command-output-and-verdicts)
  - [without --verbose the trace is absent](#scenario-without---verbose-the-trace-is-absent)
  - [verbose with a JSON report keeps stdout pure and traces to stderr](#scenario-verbose-with-a-json-report-keeps-stdout-pure-and-traces-to-stderr)
  - [a failing run under --verbose renders the FAILED block exactly once](#scenario-a-failing-run-under---verbose-renders-the-failed-block-exactly-once)
- [atago self-hosting / version](#atago-self-hosting--version) — 2 scenarios
  - [version command prints the binary name](#scenario-version-command-prints-the-binary-name)
  - [unknown command is a configuration error](#scenario-unknown-command-is-a-configuration-error)
- [atago self-hosting / yaml stream matcher](#atago-self-hosting--yaml-stream-matcher) — 2 scenarios
  - [a yaml stream matcher selects and asserts a decoded value (#9)](#scenario-a-yaml-stream-matcher-selects-and-asserts-a-decoded-value-9)
  - [a yaml matcher mismatch fails the inner spec (#9)](#scenario-a-yaml-matcher-mismatch-fails-the-inner-spec-9)
## atago self-hosting / cross-platform no-shell argv tokenization (#154)
Source: `test/e2e/atago/argv_quotes.atago.yaml`
### Scenario: a single-quoted JSON argument survives tokenization
#### When
```shell
${atago} run '{"k":"v"}'
```
#### Then
- exit code is `3`
- stderr contains `{"k":"v"}`
### Scenario: a single-quoted argument with a space stays one argument
#### When
```shell
${atago} run 'no such file.yaml'
```
#### Then
- exit code is `3`
- stderr contains `no such file.yaml`
### Scenario: a block-scalar command splits on newlines like spaces
#### When
```shell
${atago} run
no-such-file.yaml

```
#### Then
- exit code is `3`
- stderr contains `no-such-file.yaml`
### Scenario: a folded-scalar command drops its trailing newline
#### When
```shell
${atago} run no-such-file.yaml

```
#### Then
- exit code is `3`
- stderr contains `no-such-file.yaml`
## atago self-hosting / artifacts-dir failure payloads
Source: `test/e2e/atago/artifacts.atago.yaml`
### Scenario: a failing stdout equals writes expected and actual sidecars
#### Given
- Fixture file `fail.atago.yaml` is created.
#### Inputs
_Fixture `fail.atago.yaml`:_
```text
version: "1"
suite: {name: fail}
scenarios:
  - name: stdout drifts
    steps:
      - run: {shell: true, command: "printf 'actual-value\\n'"}
      - assert: {stdout: {equals: "expected-value"}}
```
#### When
```shell
${atago} run --artifacts-dir arts fail.atago.yaml
cat arts/*/*/*.actual.txt arts/*/*/*.expected.txt
```
#### Then
- after `${atago} run --artifacts-dir arts fail.atago.yaml`:
  - exit code is `1`
- after `cat arts/*/*/*.actual.txt arts/*/*/*.expected.txt`:
  - exit code is `0`
  - stdout contains `actual-value`, `expected-value`
### Scenario: a passing scenario writes no failure payload
#### Given
- Fixture file `pass.atago.yaml` is created.
#### Inputs
_Fixture `pass.atago.yaml`:_
```text
version: "1"
suite: {name: pass}
scenarios:
  - name: all good
    steps:
      - run: {shell: true, command: "printf 'ok\\n'"}
      - assert: {stdout: {equals: "ok"}}
```
#### When
```shell
${atago} run --artifacts-dir arts2 pass.atago.yaml
find arts2 -name '*.actual.txt' 2>/dev/null | wc -l
```
#### Then
- after `${atago} run --artifacts-dir arts2 pass.atago.yaml`:
  - exit code is `0`
- after `find arts2 -name '*.actual.txt' 2>/dev/null | wc -l`:
  - stdout matches `/^\s*0\s*$/`
### Scenario: the artifacts directory is created when it does not exist
#### Given
- Fixture file `nested.atago.yaml` is created.
#### Inputs
_Fixture `nested.atago.yaml`:_
```text
version: "1"
suite: {name: nested}
scenarios:
  - name: fails
    steps:
      - run: {shell: true, command: "printf 'x\\n'"}
      - assert: {stdout: {equals: "y"}}
```
#### When
```shell
${atago} run --artifacts-dir deep/made/arts nested.atago.yaml
```
#### Then
- exit code is `1`
- dir `deep/made/arts` exists
### Scenario: a file-content mismatch also writes a payload
#### Given
- Fixture file `filefail.atago.yaml` is created.
#### Inputs
_Fixture `filefail.atago.yaml`:_
```text
version: "1"
suite: {name: filefail}
scenarios:
  - name: file drifts
    steps:
      - fixture: {file: out.txt, content: "written-bytes"}
      - assert: {file: {path: out.txt, equals: "wanted-bytes"}}
```
#### When
```shell
${atago} run --artifacts-dir farts filefail.atago.yaml
find farts -type f | wc -l
```
#### Then
- after `${atago} run --artifacts-dir farts filefail.atago.yaml`:
  - exit code is `1`
- after `find farts -type f | wc -l`:
  - stdout does not match `/^\s*0\s*$/`
## atago self-hosting / variable expansion in assertion matcher values
Source: `test/e2e/atago/assert_expand.atago.yaml`
### Scenario: stdout.equals expands ${workdir}
#### When
```shell
printf "%s\n" "${workdir}/out.txt"
```
#### Then
- stdout equals an exact value
### Scenario: stdout.contains and not_contains expand a stored variable
#### When
```shell
printf "%s" hello-123
# capture ${token} from stdout
printf "%s\n" "hello-123 world"
```
#### Then
- after `printf "%s\n" "hello-123 world"`:
  - stdout contains `${token}`
  - stdout does not contain `${token}-absent`
### Scenario: file.contains expands ${workdir}
#### When
```shell
printf "%s\n" "${workdir}/marker" > note.txt
```
#### Then
- file `note.txt` contains `${workdir}/marker`
### Scenario: dir.path expands a stored variable
#### When
```shell
mkdir -p site && touch site/a.html && echo site
# capture ${outdir} from stdout
```
#### Then
- dir `${outdir}` contains `a.html`
### Scenario: changes entries expand a stored variable
#### When
```shell
echo out
# capture ${base} from stdout
echo w > ${base}2.txt
```
#### Then
- after `echo w > ${base}2.txt`:
  - the step changed exactly created `${base}2.txt`, modified nothing, deleted nothing
### Scenario: screen matcher expands a stored variable
_skipped on Windows_
#### When
```shell
echo needle
# capture ${pat} from stdout
# interactive (pty): echo needle
```
#### Then
- rendered screen contains `${pat}`
## atago self-hosting / browser (cdp) runner
Source: `test/e2e/atago/cdp.atago.yaml`
### Scenario: a cdp step with no actions fails validation (exit 2)
#### Given
- Fixture file `badcdp.atago.yaml` is created.
#### Inputs
_Fixture `badcdp.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
scenarios:
  - name: empty actions
    steps:
      - cdp:
          runner: web
          actions: []
```
#### When
```shell
${atago} run badcdp.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `at least one action`
### Scenario: a cdp step naming an undeclared runner fails validation (exit 2)
#### Given
- Fixture file `norunner.atago.yaml` is created.
#### Inputs
_Fixture `norunner.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
scenarios:
  - name: missing runner
    steps:
      - cdp:
          runner: missing
          actions:
            - navigate: http://example.com
```
#### When
```shell
${atago} run norunner.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `is not declared`
### Scenario: a screenshot action without a path fails validation (exit 2)
#### Given
- Fixture file `noshotpath.atago.yaml` is created.
#### Inputs
_Fixture `noshotpath.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
scenarios:
  - name: screenshot missing path
    steps:
      - cdp:
          runner: web
          actions:
            - navigate: http://example.com
            - screenshot:
                selector: "#x"
```
#### When
```shell
${atago} run noshotpath.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `screenshot requires a path`
### Scenario: explain lists the extended cdp actions
#### Given
- Fixture file `actions.atago.yaml` is created.
#### Inputs
_Fixture `actions.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
scenarios:
  - name: a black-box ui flow
    steps:
      - cdp:
          runner: web
          actions:
            - navigate: http://example.com
            - wait_hidden: "#spinner"
            - press: { selector: "#in", key: Enter }
            - select: { selector: "#s", value: b }
            - check: "#agree"
            - screenshot: { path: shot.png }
            - title: true
            - attribute: { selector: "#lnk", name: href }
```
#### When
```shell
${atago} explain actions.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `wait_hidden`, `screenshot shot.png`, `attribute href`
### Scenario: a browser-only field on a non-browser runner fails validation (exit 2)
#### Given
- Fixture file `crosstype.atago.yaml` is created.
#### Inputs
_Fixture `crosstype.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  api:
    type: http
    base_url: http://x
    headless: false
scenarios:
  - name: misuse
    steps:
      - run: { command: echo hi }
```
#### When
```shell
${atago} run crosstype.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `cannot be set on a http runner`
### Scenario: manifest surfaces the browser-runner configuration
#### Given
- Fixture file `cfg.atago.yaml` is created.
#### Inputs
_Fixture `cfg.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
    headless: false
    exec_path: /usr/bin/chromium
    browser_args: ["disable-gpu", "window-size=1280,720"]
scenarios:
  - name: ui
    steps:
      - cdp:
          runner: web
          actions:
            - navigate: http://example.com
            - title: true
```
#### When
```shell
${atago} manifest cfg.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.specs[0].runners[0].headless` equals `false`
- stdout at `$.specs[0].runners[0].exec_path` equals `/usr/bin/chromium`
### Scenario: an upload action without a file fails validation (exit 2)
#### Given
- Fixture file `badupload.atago.yaml` is created.
#### Inputs
_Fixture `badupload.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
scenarios:
  - name: upload missing file
    steps:
      - cdp:
          runner: web
          actions:
            - upload: { selector: "#f" }
```
#### When
```shell
${atago} run badupload.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `upload requires selector and file`
### Scenario: a download action without a click selector fails validation (exit 2)
#### Given
- Fixture file `baddownload.atago.yaml` is created.
#### Inputs
_Fixture `baddownload.atago.yaml`:_
```text
version: "1"
suite:
  name: browser
runners:
  web:
    type: browser
scenarios:
  - name: download missing click
    steps:
      - cdp:
          runner: web
          actions:
            - download: { dir: out }
```
#### When
```shell
${atago} run baddownload.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `download requires a click selector`
## atago self-hosting / changes (workdir delta assertions)
Source: `test/e2e/atago/changes.atago.yaml`
### Scenario: a generator touches exactly the files it should (POSIX)
_skipped on Windows_
#### Given
- Fixture file `config.yaml` is created.
- Fixture file `stale.html` is created.
#### Inputs
_Fixture `config.yaml`:_
```text
theme: light
```
_Fixture `stale.html`:_
```text
old
```
#### When
```shell
printf 'theme: dark\n' > config.yaml && rm stale.html && mkdir -p site/assets && printf '<html></html>' > site/index.html && printf 'body{}' > site/assets/app.css
```
#### Then
- exit code is `0`
- the step changed exactly created `site/index.html`, `site/assets/*.css`, modified `config.yaml`, deleted `stale.html`
### Scenario: an unexpected creation breaks the exact contract (POSIX)
_skipped on Windows_
#### Given
- Fixture file `check.atago.yaml` is created.
#### Inputs
_Fixture `check.atago.yaml`:_
```text
version: "1"
suite:
  name: unexpected
scenarios:
  - name: extra file
    steps:
      - run:
          shell: true
          command: 'echo a > a.txt && echo b > b.txt'
      - assert:
          changes:
            created:
              - a.txt
```
#### When
```shell
${atago} run check.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `unexpected created file`
### Scenario: stdout_to counts as created, and modified nothing holds (portable)
#### Given
- Fixture file `input.txt` is created.
#### Inputs
_Fixture `input.txt`:_
```text
seed
```
#### When
```shell
echo produced
```
#### Then
- exit code is `0`
- the step changed exactly created `result.txt`, modified nothing, deleted nothing
#### Generated artifacts
- `result.txt`
### Scenario: the delta over a retried step is cumulative across all attempts (POSIX)
_skipped on Windows_
#### When
```shell
touch a; [ -f b ] && touch c; touch b; [ -f c ]
```
#### Then
- exit code is `0`
- the step changed exactly created `a`, `b`, `c`, modified nothing, deleted nothing
### Scenario: deleting and recreating a byte-identical file appears in no list (POSIX)
_skipped on Windows_
#### Given
- Fixture file `f.txt` is created.
#### Inputs
_Fixture `f.txt`:_
```text
hello
```
#### When
```shell
rm f.txt && printf hello > f.txt
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified nothing, deleted nothing
### Scenario: deleting and recreating with different content is modified only (POSIX)
_skipped on Windows_
#### Given
- Fixture file `f.txt` is created.
#### Inputs
_Fixture `f.txt`:_
```text
hello
```
#### When
```shell
rm f.txt && printf world > f.txt
```
#### Then
- exit code is `0`
- the step changed exactly created nothing, modified `f.txt`, deleted nothing
### Scenario: stdout_to overwrites a fixture (modified) while stderr_to creates an empty file (POSIX)
_skipped on Windows_
#### Given
- Fixture file `existing.txt` is created.
#### Inputs
_Fixture `existing.txt`:_
```text
old
```
#### When
```shell
printf newcontent
```
#### Then
- exit code is `0`
- the step changed exactly created `err.txt`, modified `existing.txt`, deleted nothing
#### Generated artifacts
- `existing.txt`
- `err.txt`
### Scenario: a pty step feeds the delta scan just like a run step (POSIX)
_skipped on Windows_
#### When
```shell
# interactive (pty): sh -c "touch from-pty; echo done"
```
#### Then
- the step changed exactly created `from-pty`
### Scenario: a doublestar glob pins an arbitrary-depth generated tree exactly (POSIX)
_skipped on Windows_
#### When
```shell
mkdir -p out/a/b && printf '1' > out/top.txt && printf '2' > out/a/mid.txt && printf '3' > out/a/b/leaf.txt
```
#### Then
- exit code is `0`
- the step changed exactly created `out/**`, modified nothing, deleted nothing
### Scenario: a stray file outside the doublestar prefix breaks the exact contract (POSIX)
_skipped on Windows_
#### Given
- Fixture file `check.atago.yaml` is created.
#### Inputs
_Fixture `check.atago.yaml`:_
```text
version: "1"
suite:
  name: stray-outside-doublestar
scenarios:
  - name: extra file beside the tree
    steps:
      - run:
          shell: true
          command: 'mkdir -p out/a && printf x > out/a/leaf.txt && printf y > stray.txt'
      - assert:
          changes:
            created:
              - "out/**"
```
#### When
```shell
${atago} run check.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `unexpected created file`
### Scenario: a doublestar glob matches a nested redirect target (portable)
#### When
```shell
echo produced
```
#### Then
- exit code is `0`
- the step changed exactly created `out/**`, modified nothing, deleted nothing
#### Generated artifacts
- `out/deep/result.txt`
### Scenario: a doublestar prefix covers both redirect streams (portable)
#### When
```shell
echo out
```
#### Then
- exit code is `0`
- the step changed exactly created `logs/**`, modified nothing, deleted nothing
#### Generated artifacts
- `logs/out.txt`
- `logs/err.txt`
## atago self-hosting / CLI scenario selection
Source: `test/e2e/atago/cli_selection.atago.yaml`
### Scenario: filter selects by a name substring
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --filter alpha inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 1
- stdout contains `"alpha"`
### Scenario: filter is OR across a comma-separated list
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --filter alpha,beta inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 2
- stdout contains `"alpha"`, `"beta"`
### Scenario: tag selects scenarios carrying the tag
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --tag fast inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 2
- stdout contains `"alpha"`, `"gamma"`
### Scenario: a repeated tag flag is OR
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --tag fast --tag slow inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 3
### Scenario: skip-tag removes scenarios carrying the tag
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --skip-tag slow inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 1
- stdout contains `"alpha"`
### Scenario: tag and skip-tag compose as selected minus skipped
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --tag fast --skip-tag slow inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 1
- stdout contains `"alpha"`
### Scenario: a filter that matches nothing selects an empty set and still exits zero
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite: {name: inner}
scenarios:
  - name: alpha
    tags: [fast]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: beta
    tags: [slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
  - name: gamma
    tags: [fast, slow]
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json --filter no_such_name inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 0
## atago self-hosting / completion
Source: `test/e2e/atago/completion.atago.yaml`
### Scenario: bash completion emits a recognizable script
#### When
```shell
${atago} completion bash
```
#### Then
- exit code is `0`
- stdout contains `complete -F _atago atago`
### Scenario: zsh completion emits a compdef script
#### When
```shell
${atago} completion zsh
```
#### Then
- exit code is `0`
- stdout contains `#compdef atago`
### Scenario: fish completion emits complete directives
#### When
```shell
${atago} completion fish
```
#### Then
- exit code is `0`
- stdout contains `complete -c atago`
### Scenario: powershell completion registers an argument completer
#### When
```shell
${atago} completion powershell
```
#### Then
- exit code is `0`
- stdout contains `Register-ArgumentCompleter`
### Scenario: unknown shell is a configuration error
#### When
```shell
${atago} completion tcsh
```
#### Then
- exit code is `3`
- stderr contains `unknown shell`
## atago self-hosting / db runner
Source: `test/e2e/atago/db.atago.yaml`
### Scenario: query workflow (create, insert, select, row assert, value binding) passes
#### Given
- Fixture file `db.atago.yaml` is created.
#### Inputs
_Fixture `db.atago.yaml`:_
```text
version: "1"
suite:
  name: inner-db
runners:
  store:
    type: db
    dsn: "sqlite::memory:"
scenarios:
  - name: seed then query with row assertions and binding
    steps:
      - query:
          runner: store
          sql: "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, role TEXT)"
      - query:
          runner: store
          sql: "INSERT INTO users (name, role) VALUES ('alice','admin'), ('bob','user')"
      - query:
          runner: store
          sql: "SELECT id, name, role FROM users ORDER BY id"
      - assert:
… (truncated, 23 more lines)
```
#### When
```shell
${atago} run db.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `passed`
### Scenario: a query against an undeclared runner fails validation (exit 2)
#### Given
- Fixture file `norunner.atago.yaml` is created.
#### Inputs
_Fixture `norunner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner-db
scenarios:
  - name: references a db runner that is not declared
    steps:
      - query:
          runner: missing
          sql: "SELECT 1"
```
#### When
```shell
${atago} run norunner.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `is not declared`
## atago self-hosting / top-level defaults
Source: `test/e2e/atago/defaults.atago.yaml`
### Scenario: defaults.run.shell applies to every run step without repeating it
#### Given
- Fixture file `shell.atago.yaml` is created.
#### Inputs
_Fixture `shell.atago.yaml`:_
```text
version: "1"
suite:
  name: shelled
defaults:
  run:
    shell: true
scenarios:
  - name: uses a shell builtin and env expansion
    env:
      WHO: world
    steps:
      - run:
          command: echo "hello $WHO"
      - assert:
          stdout:
            contains: hello world
```
#### When
```shell
${atago} run shell.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: defaults.scenario.env is merged and an explicit scenario env wins
#### Given
- Fixture file `env.atago.yaml` is created.
#### Inputs
_Fixture `env.atago.yaml`:_
```text
version: "1"
suite:
  name: enved
defaults:
  run:
    shell: true
  scenario:
    env:
      FROM_DEFAULT: base
      SHARED: default
scenarios:
  - name: sees the default and overrides one key
    env:
      SHARED: own
    steps:
      - run:
          command: echo "$FROM_DEFAULT-$SHARED"
      - assert:
          stdout:
            contains: base-own
```
#### When
```shell
${atago} run env.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: defaults.run.sandbox_home governs a run step and a pty step alike (POSIX)
_skipped on Windows_
#### Given
- Fixture file `sandbox.atago.yaml` is created.
#### Inputs
_Fixture `sandbox.atago.yaml`:_
```text
version: "1"
suite:
  name: shared-sandbox
defaults:
  run:
    shell: true
    sandbox_home: true
scenarios:
  - name: a run step writes and a pty step reads under one sandbox home
    steps:
      - run:
          command: 'mkdir -p "$XDG_CONFIG_HOME/mytool" && printf editor=vim > "$XDG_CONFIG_HOME/mytool/config"'
      - assert:
          exit_code: 0
      - pty:
          shell: true
          command: 'cat "$XDG_CONFIG_HOME/mytool/config"'
          session:
            - expect: 'editor=vim'
      - assert:
… (truncated, 6 more lines)
```
#### When
```shell
${atago} run sandbox.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: an unsupported defaults field is a load-time error (exit 2)
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
defaults:
  run:
    command: echo nope
scenarios:
  - name: never runs
    steps:
      - run:
          command: echo hi
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `defaults.run.command is not supported`
### Scenario: defaults.run.env merges per key and a step env wins the collisions
#### Given
- Fixture file `env.atago.yaml` is created.
#### Inputs
_Fixture `env.atago.yaml`:_
```text
version: "1"
suite:
  name: envmerge
defaults:
  run:
    shell: true
    env: {BASE: from-default, SHARED: default-value}
scenarios:
  - name: step env overrides one key and inherits the other
    steps:
      - run:
          command: "echo [$BASE][$SHARED]"
          env: {SHARED: step-value}
      - assert:
          stdout:
            contains: "[from-default][step-value]"
```
#### When
```shell
${atago} run env.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASSED`
### Scenario: a step opts out of defaults.run.shell with an explicit shell false
#### Given
- Fixture file `optout.atago.yaml` is created.
#### Inputs
_Fixture `optout.atago.yaml`:_
```text
version: "1"
suite:
  name: optout
defaults:
  run:
    shell: true
scenarios:
  - name: this step runs without a shell
    steps:
      - run:
          shell: false
          command: "echo literal-$HOME"
      - assert:
          stdout:
            contains: "literal-$HOME"
```
#### When
```shell
${atago} run optout.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASSED`
## atago self-hosting / dir assertion
Source: `test/e2e/atago/dir.atago.yaml`
### Scenario: directory/tree assertions cover a multi-file generator
_skipped on Windows_
#### When
```shell
mkdir -p site/assets && printf '<html>' > site/index.html && printf '<html>' > site/about.html && printf 'body{}' > site/assets/app.css
```
#### Then
- exit code is `0`
- dir `site` exists, contains `index.html`, contains `about.html`, contains `assets/app.css`, does not contain `secret.txt`, has 3 entries, has >= 1 entries, has <= 10 entries, matches glob `*.html`
### Scenario: a missing directory can be asserted absent
#### Then
- dir `never-created` does not exist
### Scenario: a dangling symlink is a present directory entry (membership uses Lstat)
_skipped on Windows_
#### When
```shell
mkdir -p linkdir && ln -s /nonexistent-target-xyz linkdir/planted
```
#### Then
- exit code is `0`
- dir `linkdir` contains `planted`, does not contain `never-planted`
## atago self-hosting / recursive dir asserts + tree snapshots
Source: `test/e2e/atago/dir_tree.atago.yaml`
### Scenario: record, compare green, then a mutation names the changed paths
#### Given
- Fixture file `inner.atago.yaml` is created.
- Fixture file `inner_mutated.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: tree matches its golden
    steps:
      - fixture:
          file: site/hugo.toml
          content: "baseURL = 'x'\n"
      - fixture:
          file: site/content/posts/hello.md
          content: "# hello\n"
      - assert:
          dir:
            path: site
            snapshot: snapshots/site_tree.txt
```
_Fixture `inner_mutated.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: tree matches its golden
    steps:
      - fixture:
          file: site/hugo.toml
          content: "baseURL = 'x'\n"
      - fixture:
          file: site/content/posts/hello.md
          content: "# CHANGED\n"
      - fixture:
          file: site/extra.txt
          content: "new\n"
      - assert:
          dir:
            path: site
            snapshot: snapshots/site_tree.txt
```
#### When
```shell
${atago} run --update-snapshots inner.atago.yaml
${atago} run inner.atago.yaml
${atago} run inner_mutated.atago.yaml
```
#### Then
- after `${atago} run --update-snapshots inner.atago.yaml`:
  - exit code is `0`
  - file `snapshots/site_tree.txt` contains `dir content`, `file hugo.toml sha256:`
- after `${atago} run inner.atago.yaml`:
  - exit code is `0`
- after `${atago} run inner_mutated.atago.yaml`:
  - exit code is `1`
  - stdout contains `added:   file extra.txt`, `changed: content/posts/hello.md`
### Scenario: recursive matchers and ignore globs walk the tree
#### Given
- Fixture file `out/a/deep/nested.md` is created.
- Fixture file `out/top.txt` is created.
- Fixture file `out/noise.log` is created.
#### Inputs
_Fixture `out/a/deep/nested.md`:_
```text
nested
```
_Fixture `out/top.txt`:_
```text
top
```
_Fixture `out/noise.log`:_
```text
noise
```
#### Then
- dir `out` contains `a/deep/nested.md`, has 2 entries, matches glob `*.md`, (recursive), ignoring *.log
- dir `out` does not contain `a/deep/missing.md`, (recursive)
### Scenario: combining snapshot with matchers is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: over-specified
    steps:
      - run: {command: echo hi}
      - assert:
          dir:
            path: out
            snapshot: tree.txt
            contains: [a.txt]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `snapshot cannot be combined`
## atago self-hosting / doc
Source: `test/e2e/atago/doc.atago.yaml`
### Scenario: doc generates Markdown to a file
#### Given
- Fixture file `target.atago.yaml` is created.
#### Inputs
_Fixture `target.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: greet
    steps:
      - run:
          command: echo hello
      - assert:
          stdout:
            contains: hello
```
#### When
```shell
${atago} doc --out specs.md target.atago.yaml
```
#### Then
- exit code is `0`
- file `specs.md` contains `# atago Behavior Specs`
- file `specs.md` contains `### Scenario: greet`
### Scenario: doc writes Markdown to stdout without --out
#### Given
- Fixture file `t2.atago.yaml` is created.
#### Inputs
_Fixture `t2.atago.yaml`:_
```text
version: "1"
suite:
  name: sample2
scenarios:
  - name: g2
    steps:
      - run:
          command: echo hi
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} doc t2.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `## sample2`
### Scenario: doc emits a summary, table of contents, and input previews
#### Given
- Fixture file `rich.atago.yaml` is created.
#### Inputs
_Fixture `rich.atago.yaml`:_
```text
version: "1"
suite:
  name: rich
scenarios:
  - name: seeded query
    tags: [smoke]
    steps:
      - fixture:
          file: data.csv
          content: "id,name\n1,ada\n"
      - run:
          command: cat data.csv
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} doc rich.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `## Summary`, `1 suite · 1 scenario`, `## Contents`, `(#scenario-seeded-query)`, `#### Inputs`, `id,name`
### Scenario: doc --split-by-spec writes one file per spec and an index
#### Given
- Fixture file `one.atago.yaml` is created.
- Fixture file `two.atago.yaml` is created.
#### Inputs
_Fixture `one.atago.yaml`:_
```text
version: "1"
suite:
  name: suite-one
scenarios:
  - name: a
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
```
_Fixture `two.atago.yaml`:_
```text
version: "1"
suite:
  name: suite-two
scenarios:
  - name: b
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
```
#### When
```shell
${atago} doc --split-by-spec --out-dir generated one.atago.yaml two.atago.yaml
```
#### Then
- exit code is `0`
- file `generated/one.md` exists
- file `generated/two.md` exists
- file `generated/index.md` contains `atago Behavior Specs — Index`
- file `generated/index.md` contains `(one.md)`
#### Generated artifacts
- `generated/one.md`
- `generated/two.md`
### Scenario: doc --split-by-spec requires --out-dir
#### Given
- Fixture file `solo.atago.yaml` is created.
#### Inputs
_Fixture `solo.atago.yaml`:_
```text
version: "1"
suite:
  name: solo
scenarios:
  - name: only
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
```
#### When
```shell
${atago} doc --split-by-spec solo.atago.yaml
```
#### Then
- exit code is `3`
- stderr contains `requires --out-dir`
## atago self-hosting / duration assertion
Source: `test/e2e/atago/duration.atago.yaml`
### Scenario: a fast step passes a generous upper bound
#### When
```shell
${atago} version
```
#### Then
- exit code is `0`
- completes in under 60s
### Scenario: an impossible bound fails and shows the measured duration
#### Given
- Fixture file `slow.atago.yaml` is created.
#### Inputs
_Fixture `slow.atago.yaml`:_
```text
version: "1"
suite:
  name: slow
scenarios:
  - name: cannot beat 1ns
    steps:
      - run:
          command: ${atago} version
      - assert:
          duration:
            lt: 1ns
```
#### When
```shell
${atago} run slow.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `assert duration < 1ns`, `orders of magnitude`
### Scenario: a deliberate wait satisfies a lower bound
_skipped on Windows_
#### When
```shell
sleep 0.2
```
#### Then
- exit code is `0`
- completes in under 60s and in at least 100ms
### Scenario: a duration assert with no preceding step is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: no step to measure
    steps:
      - assert:
          duration: {lt: 2s}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `requires an immediately preceding`
## atago self-hosting / edge cases
Source: `test/e2e/atago/edge.atago.yaml`
### Scenario: JSON assertion on empty stdout reports an empty stream
#### Given
- Fixture file `empty.atago.yaml` is created.
#### Inputs
_Fixture `empty.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: json on empty
    steps:
      - run:
          shell: true
          command: "exit 0"
      - assert:
          stdout:
            json:
              path: "$.x"
              equals: 1
```
#### When
```shell
${atago} run empty.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `was empty`
### Scenario: an unsupported matcher is a parse error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: csv matcher
    steps:
      - run:
          shell: true
          command: echo hi
      - assert:
          stdout:
            csv:
              column: name
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
### Scenario: a mixed valid+invalid run reads FAILED and counts the dropped spec
#### Given
- Fixture file `good.atago.yaml` is created.
- Fixture file `broken.atago.yaml` is created.
#### Inputs
_Fixture `good.atago.yaml`:_
```text
version: "1"
suite:
  name: good
scenarios:
  - name: echo works
    steps:
      - run: {shell: true, command: echo hi}
      - assert: {exit_code: 0}
```
_Fixture `broken.atago.yaml`:_
```text
version: "1"
suite:
  name: broken
scenarios:
  - name: uses an unknown field
    steps:
      - nonsense_field: {command: echo nope}
```
#### When
```shell
${atago} run good.atago.yaml broken.atago.yaml
```
#### Then
- exit code is `2`
- stdout contains `1 spec failed to load`
- stdout does not contain `PASSED`
### Scenario: a snapshot update error names the snapshot command, not run
#### When
```shell
${atago} snapshot update no-such-spec.atago.yaml
```
#### Then
- exit code is `3`
- stderr contains `atago snapshot update: cannot access`
### Scenario: a json assertion on malformed input fails cleanly, without a crash
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: badjson}
scenarios:
  - name: json on broken bytes
    steps:
      - fixture: {file: broken.json, content: '{"":f,"":0 0'}
      - assert: {file: {path: broken.json, json: {path: "$.x", equals: 1}}}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `not valid JSON`
## atago self-hosting / workdir + scenario env + not_contains
Source: `test/e2e/atago/env_workdir.atago.yaml`
### Scenario: run.stdout_to redirects stdout to a workdir file without a shell
#### When
```shell
printf marked
```
#### Then
- exit code is `0`
- file `marker.txt` contains `marked`
#### Generated artifacts
- `marker.txt`
### Scenario: scenario env is shared by every run step and overridable per step
#### Given
- Environment variables are set: OVERRIDE_ME.
#### When
```shell
printf '%s\n' "$SHARED"
printf '%s\n' "$OVERRIDE_ME"
```
#### Then
- after `printf '%s\n' "$SHARED"`:
  - stdout equals an exact value
- after `printf '%s\n' "$OVERRIDE_ME"`:
  - stdout equals an exact value
### Scenario: scenario env can reference ${workdir} for isolated paths
#### When
```shell
printf '%s\n' "$ISO_HOME"
```
#### Then
- stdout contains `/home`
- stdout does not contain `${workdir}/home/extra`
### Scenario: file not_contains passes when the substring is absent
#### Given
- Fixture file `data.txt` is created.
#### Inputs
_Fixture `data.txt`:_
```text
alpha
beta
```
#### Then
- file `data.txt` is checked
### Scenario: not_contains fails when the substring is present
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: stdout unexpectedly contains hello
    steps:
      - run:
          command: echo hello world
      - assert:
          stdout:
            not_contains: hello
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `unexpectedly present`
### Scenario: a shell metacharacter without shell is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: forgot shell true
    steps:
      - run:
          command: echo hi > out.txt
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `shell is not enabled`, `shell: true`, `stdout_to`
## atago self-hosting / exit_code in-set matcher
Source: `test/e2e/atago/exit_code_in.atago.yaml`
### Scenario: a listed exit code passes
#### When
```shell
exit 2
```
#### Then
- exit code is one of `0`, `2`
### Scenario: an unlisted exit code fails and the output lists the set
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: exit 2 is not in the accepted set
    steps:
      - run:
          shell: true
          command: exit 2
      - assert:
          exit_code:
            in: [0, 1]
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `exit code in [0, 1]`
### Scenario: mixing not and in is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: both forms
    steps:
      - run:
          shell: true
          command: exit 0
      - assert:
          exit_code:
            not: 1
            in: [0, 2]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `exactly one of`
### Scenario: an empty in list is a load-time error
#### Given
- Fixture file `empty.atago.yaml` is created.
#### Inputs
_Fixture `empty.atago.yaml`:_
```text
version: "1"
suite:
  name: empty
scenarios:
  - name: empty set
    steps:
      - run:
          shell: true
          command: exit 0
      - assert:
          exit_code:
            in: []
```
#### When
```shell
${atago} run empty.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `at least one accepted exit code`
## atago self-hosting / exit code semantics
Source: `test/e2e/atago/exit_codes.atago.yaml`
### Scenario: a clean exit is zero
#### When
```shell
exit 0
```
#### Then
- exit code is `0`
### Scenario: a general error is one
#### When
```shell
exit 1
```
#### Then
- exit code is `1`
### Scenario: a usage error is two
#### When
```shell
exit 2
```
#### Then
- exit code is `2`
### Scenario: an arbitrary code passes through unchanged
#### When
```shell
exit 42
```
#### Then
- exit code is `42`
### Scenario: the single-byte ceiling is 255
#### When
```shell
exit 255
```
#### Then
- exit code is `255`
### Scenario: the not matcher excludes a specific code
#### When
```shell
exit 3
```
#### Then
- exit code is not `0`
### Scenario: the in matcher accepts any listed code
#### When
```shell
exit 2
```
#### Then
- exit code is one of `0`, `1`, `2`
### Scenario: an unlisted code fails the in matcher and names the set
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: five is not in the accepted set
    steps:
      - run: {shell: true, command: "exit 5"}
      - assert:
          exit_code:
            in: [0, 1, 2]
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `exit code in [0, 1, 2]`
### Scenario: SIGKILL is reported as 137
_skipped on Windows_
#### When
```shell
kill -KILL $$
```
#### Then
- exit code is `137`
### Scenario: SIGTERM is reported as 143
_skipped on Windows_
#### When
```shell
kill -TERM $$
```
#### Then
- exit code is `143`
### Scenario: SIGINT is reported as 130
_skipped on Windows_
#### When
```shell
kill -INT $$
```
#### Then
- exit code is `130`
### Scenario: a signal exit composes with the in matcher alongside normal codes
_skipped on Windows_
#### When
```shell
kill -TERM $$
```
#### Then
- exit code is one of `0`, `143`
### Scenario: a missing command is 127 under the shell
_skipped on Windows_
#### When
```shell
no_such_command_zzz
```
#### Then
- exit code is `127`
### Scenario: POSIX exit codes wrap modulo 256
_skipped on Windows_
#### When
```shell
exit 257
```
#### Then
- exit code is `1`
## atago self-hosting / explain
Source: `test/e2e/atago/explain.atago.yaml`
### Scenario: explain summarizes a spec without running it
#### Given
- Fixture file `target.atago.yaml` is created.
#### Inputs
_Fixture `target.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: list as json
    steps:
      - run:
          command: echo "[]"
      - assert:
          exit_code: 0
          stdout:
            contains: "["
```
#### When
```shell
${atago} explain target.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `Suite: sample`, `Scenario: list as json`, `Commands:`, `Network policy:`
## atago self-hosting / file equals and equals_file byte-equality (#155)
Source: `test/e2e/atago/file_equals.atago.yaml`
### Scenario: equals_file passes for two byte-identical files
#### Given
- Fixture file `in.hex` is created.
- Fixture file `out.hex` is created.
#### Inputs
_Fixture `in.hex`:_
```text
DEADBEEF
```
_Fixture `out.hex`:_
```text
DEADBEEF
```
#### Then
- file `out.hex` is byte-identical to `in.hex`
### Scenario: equals matches an inline literal byte-for-byte
#### Given
- Fixture file `token.txt` is created.
#### Inputs
_Fixture `token.txt`:_
```text
opaque-value-42
```
#### Then
- file `token.txt` equals exact bytes
### Scenario: equals_file fails the inner spec when the two files differ by one byte
#### Given
- Fixture file `neq.atago.yaml` is created.
#### Inputs
_Fixture `neq.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: two files differ by one byte
    steps:
      - fixture:
          file: a.hex
          content: "DEADBEEF"
      - fixture:
          file: b.hex
          content: "DEADBEEE"
      - assert:
          file:
            path: a.hex
            equals_file: b.hex
```
#### When
```shell
${atago} run neq.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `not byte-identical`
### Scenario: equals_file is byte-exact — a CRLF vs LF difference fails
#### Given
- Fixture file `crlf.atago.yaml` is created.
#### Inputs
_Fixture `crlf.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: LF and CRLF files are not byte-identical
    steps:
      - fixture:
          file: lf.txt
          base64: "bGluZQo="
      - fixture:
          file: crlf.txt
          base64: "bGluZQ0K"
      - assert:
          file:
            path: lf.txt
            equals_file: crlf.txt
```
#### When
```shell
${atago} run crlf.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `not byte-identical`
## atago self-hosting / fixture from (copy committed testdata)
Source: `test/e2e/atago/fixture_from.atago.yaml`
### Scenario: a committed binary blob is copied verbatim into the workdir
#### Given
- Fixture file `copied.bin` is created.
#### When
```shell
wc -c < copied.bin
```
#### Then
- exit code is `0`
- stdout contains `21`
- file `copied.bin` contains `binary-marker`
### Scenario: copying from a missing source errors the scenario
#### Given
- Fixture file `copied.bin` is created.
#### Inputs
_Fixture `copied.bin`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: missing source
    steps:
      - fixture:
          file: x.bin
          from: testdata/does-not-exist.bin
      - run:
          command: "true"
```
#### When
```shell
${atago} run copied.bin
```
#### Then
- exit code is `4`
- stdout contains `copy from`
## atago self-hosting / fixture symlink+mode+mtime, file executable, env skip
Source: `test/e2e/atago/fixture_modes.atago.yaml`
### Scenario: a symlink fixture resolves to its target
#### Given
- Fixture file `target.txt` is created.
- Fixture file `alias.txt` is created.
#### Inputs
_Fixture `target.txt`:_
```text
payload
```
#### When
```shell
cat alias.txt
```
#### Then
- stdout equals an exact value
### Scenario: fixture.mode sets permissions and file.executable reads them
#### Given
- Fixture file `run.sh` is created.
- Fixture file `data.txt` is created.
#### Inputs
_Fixture `run.sh`:_
```text
echo hi
```
_Fixture `data.txt`:_
```text
plain
```
#### Then
- file `run.sh` is checked
- file `data.txt` is checked
### Scenario: fixture.mtime pins the modification time
#### Given
- Fixture file `stamped.txt` is created.
#### Inputs
_Fixture `stamped.txt`:_
```text
x
```
#### When
```shell
date -u -r stamped.txt +%Y
```
#### Then
- stdout contains `2021`
### Scenario: only.env skips when the variable is unset
_only when env ATAGO_DEFINITELY_UNSET is set_
#### When
```shell
false
```
### Scenario: skip.env runs when the variable is unset
_skipped when env ATAGO_DEFINITELY_UNSET is set_
#### When
```shell
true
```
#### Then
- exit code is `0`
## atago self-hosting / flaky tooling (--repeat, --retry-failed)
Source: `test/e2e/atago/flaky.atago.yaml`
### Scenario: retry-failed recovers a flaky scenario and reports it loudly
_skipped on Windows_
#### Given
- Fixture file `flaky.atago.yaml` is created.
#### Inputs
_Fixture `flaky.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: flaky once
    steps:
      - run:
          shell: true
          command: "if [ -f '${workdir}/seen.txt' ]; then echo recovered; else touch '${workdir}/seen.txt'; exit 1; fi"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run flaky.atago.yaml
rm -f seen.txt
${atago} run --retry-failed 1 flaky.atago.yaml
rm -f seen.txt
${atago} run --retry-failed 1 --report json flaky.atago.yaml
```
#### Then
- after `${atago} run flaky.atago.yaml`:
  - exit code is `1`
- after `${atago} run --retry-failed 1 flaky.atago.yaml`:
  - exit code is `0`
  - stdout contains `FLAKY:`, `passed after 2 attempts`, `1 flaky`
- after `${atago} run --retry-failed 1 --report json flaky.atago.yaml`:
  - exit code is `0`
  - stdout contains `"status": "flaky"`, `"attempts": 2`
### Scenario: repeat surfaces flakiness that a single run would miss
_skipped on Windows_
#### Given
- Fixture file `green.atago.yaml` is created.
- Fixture file `flaky.atago.yaml` is created.
- Fixture file `broken.atago.yaml` is created.
#### Inputs
_Fixture `green.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: steady
    steps:
      - run: {shell: true, command: echo ok}
      - assert:
          exit_code: 0
```
_Fixture `flaky.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: flaky once
    steps:
      - run:
          shell: true
          command: "if [ -f '${workdir}/seen.txt' ]; then echo recovered; else touch '${workdir}/seen.txt'; exit 1; fi"
      - assert:
          exit_code: 0
```
_Fixture `broken.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: always fails
    steps:
      - run: {shell: true, command: exit 1}
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run --repeat 3 green.atago.yaml
${atago} run --repeat 3 flaky.atago.yaml
${atago} run --repeat 3 broken.atago.yaml
```
#### Then
- after `${atago} run --repeat 3 green.atago.yaml`:
  - exit code is `0`
  - stdout contains `steady: 3/3 passed`
- after `${atago} run --repeat 3 flaky.atago.yaml`:
  - exit code is `0`
  - stdout contains `flaky once: 2/3 passed`, `1 flaky`
- after `${atago} run --repeat 3 broken.atago.yaml`:
  - exit code is `1`
  - stdout contains `always fails: 0/3 passed`, `1 failed`
### Scenario: repeat and retry-failed are mutually exclusive
#### Given
- Fixture file `any.atago.yaml` is created.
#### Inputs
_Fixture `any.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: s
    steps:
      - run: {command: echo hi}
```
#### When
```shell
${atago} run --repeat 2 --retry-failed 1 any.atago.yaml
```
#### Then
- exit code is `3`
- stderr contains `mutually exclusive`
## atago self-hosting / grpc runner
Source: `test/e2e/atago/grpc.atago.yaml`
### Scenario: a grpc runner without a target fails validation (exit 2)
#### Given
- Fixture file `badgrpc.atago.yaml` is created.
#### Inputs
_Fixture `badgrpc.atago.yaml`:_
```text
version: "1"
suite:
  name: grpc
runners:
  g:
    type: grpc
scenarios:
  - name: call
    steps:
      - grpc:
          runner: g
          method: pkg.Service/Method
```
#### When
```shell
${atago} run badgrpc.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `requires a target`
### Scenario: a grpc step naming an undeclared runner fails validation (exit 2)
#### Given
- Fixture file `norunner.atago.yaml` is created.
#### Inputs
_Fixture `norunner.atago.yaml`:_
```text
version: "1"
suite:
  name: grpc
scenarios:
  - name: call
    steps:
      - grpc:
          runner: missing
          method: pkg.Service/Method
```
#### When
```shell
${atago} run norunner.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `is not declared`
## atago self-hosting / hermetic environment (clear_env + pass_env)
Source: `test/e2e/atago/hermetic_env.atago.yaml`
### Scenario: clear_env drops inherited host variables
_skipped on Windows_
#### Given
- The command runs with a cleared environment.
#### When
```shell
env
env
```
#### Then
- after `env`:
  - exit code is `0`
  - stdout contains `ATAGO_HERMETIC_CANARY`
- after `env`:
  - exit code is `0`
  - stdout contains `ATAGO_HERMETIC_CANARY=leaked-from-scenario`
  - stdout does not contain `PATH=/`
### Scenario: pass_env re-admits an allowlist of host variables
_skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH).
#### When
```shell
env
```
#### Then
- exit code is `0`
- stdout contains `PATH=`
- stdout does not contain `HOME=`
### Scenario: explicit env wins over a passed-through host variable
_skipped on Windows_
#### Given
- Environment variables are set: HOME.
- The command runs with a cleared environment (passing through: HOME).
#### When
```shell
printf '%s\n' "$HOME"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: pass_env without clear_env is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: forgot clear_env
    steps:
      - run:
          command: env
          pass_env: [PATH]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `pass_env requires clear_env: true`, `steps[0].run`
### Scenario: unset host variables in pass_env are skipped, not an error
_skipped on Windows_
#### Given
- The command runs with a cleared environment (passing through: PATH, ATAGO_SURELY_UNSET_VAR_2026).
#### When
```shell
env
```
#### Then
- exit code is `0`
- stdout contains `PATH=`
- stdout does not contain `ATAGO_SURELY_UNSET_VAR_2026`
## atago self-hosting / http runner
Source: `test/e2e/atago/http.atago.yaml`
### Scenario: a denied host is a security policy violation (exit 6)
#### Given
- Fixture file `denied.atago.yaml` is created.
#### Inputs
_Fixture `denied.atago.yaml`:_
```text
version: "1"
suite:
  name: api
permissions:
  network:
    allow:
      - allowed.example
runners:
  api:
    type: http
    base_url: http://127.0.0.1:1
scenarios:
  - name: request to a non-allowlisted host
    steps:
      - http:
          runner: api
          method: GET
          path: /ping
```
#### When
```shell
${atago} run denied.atago.yaml
```
#### Then
- exit code is `6`
- stdout contains `network policy denies`
### Scenario: an http step with an undeclared runner fails validation (exit 2)
#### Given
- Fixture file `norunner.atago.yaml` is created.
#### Inputs
_Fixture `norunner.atago.yaml`:_
```text
version: "1"
suite:
  name: api
scenarios:
  - name: references a runner that is not declared
    steps:
      - http:
          runner: missing
          method: GET
          path: /ping
```
#### When
```shell
${atago} run norunner.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `is not declared`
## atago self-hosting / image
Source: `test/e2e/atago/image.atago.yaml`
### Scenario: format, dimension and alpha assertions pass on a PNG
#### Given
- Fixture file `img.atago.yaml` is created.
#### Inputs
_Fixture `img.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: a red 2x2 png
    steps:
      - fixture:
          file: red.png
          base64: iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAIAAAD91JpzAAAAFElEQVR4nGL5zwACTGCSARAAAP//DUcBBkxXt5sAAAAASUVORK5CYII=
      - assert:
          image:
            path: red.png
            format: png
            width: 2
            height: 2
            min_width: 1
            max_width: 100
            # This is an opaque truecolor (color-type 2) PNG with no
            # alpha channel, so alpha is false (issue #13).
            alpha: false
```
#### When
```shell
${atago} run img.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASSED`
### Scenario: a pixel comparison against an identical baseline passes
#### Given
- Fixture file `baseline.png` is created.
- Fixture file `sim.atago.yaml` is created.
#### Inputs
_Fixture `sim.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: output matches the committed baseline
    steps:
      - fixture:
          file: out.png
          base64: iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAIAAABLbSncAAAAGUlEQVR4nGJJSUlhwAaYsIoOWglAAAAA///xaAE/jf0lQAAAAABJRU5ErkJggg==
      - assert:
          image:
            path: out.png
            similar_to: ${workdir}/baseline.png
```
#### When
```shell
${atago} run sim.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASSED`
### Scenario: a wrong dimension assertion fails with a clear diff
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: width does not match
    steps:
      - fixture:
          file: red.png
          base64: iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAIAAAD91JpzAAAAFElEQVR4nGL5zwACTGCSARAAAP//DUcBBkxXt5sAAAAASUVORK5CYII=
      - assert:
          image:
            path: red.png
            width: 999
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `FAILED`
### Scenario: a failing similar_to writes visual diff artifacts
#### Given
- Fixture file `baseline.png` is created.
- Fixture file `diff.atago.yaml` is created.
#### Inputs
_Fixture `diff.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: output differs from the baseline
    steps:
      - fixture:
          file: out.png
          base64: iVBORw0KGgoAAAANSUhEUgAAAAQAAAAECAIAAAAmkwkpAAAAGElEQVR4nGJhYPjPAANMDEgANwcQAAD//zJvAQo6fkx6AAAAAElFTkSuQmCC
      - assert:
          image:
            path: out.png
            similar_to: ${workdir}/baseline.png
```
#### When
```shell
${atago} run --report json --artifacts-dir arts diff.atago.yaml
ls arts/*/*/
cat arts/*/*/*image.metadata.json
```
#### Then
- after `${atago} run --report json --artifacts-dir arts diff.atago.yaml`:
  - exit code is `1`
  - stdout contains `"artifacts"`
  - file `arts` exists
- after `ls arts/*/*/`:
  - exit code is `0`
  - stdout contains `image.actual.png`, `image.baseline.png`, `image.diff.png`, `image.metadata.json`
- after `cat arts/*/*/*image.metadata.json`:
  - exit code is `0`
  - stdout at `$.diff_generated` equals `true`
#### Generated artifacts
- `arts`
## atago self-hosting / init
Source: `test/e2e/atago/init.atago.yaml`
### Scenario: init scaffolds a runnable spec
#### When
```shell
${atago} init starter.atago.yaml
${atago} run starter.atago.yaml
```
#### Then
- after `${atago} init starter.atago.yaml`:
  - exit code is `0`
  - file `starter.atago.yaml` exists
- after `${atago} run starter.atago.yaml`:
  - exit code is `0`
  - stdout contains `PASSED`
#### Generated artifacts
- `starter.atago.yaml`
### Scenario: init emits a resolvable schema header for editor completion
#### When
```shell
${atago} init headed.atago.yaml
head -1 headed.atago.yaml
```
#### Then
- after `${atago} init headed.atago.yaml`:
  - exit code is `0`
- after `head -1 headed.atago.yaml`:
  - exit code is `0`
  - stdout contains `# yaml-language-server: $schema=https://`
  - stdout does not contain `./schema/`
### Scenario: init refuses to overwrite without --force
#### Given
- Fixture file `taken.atago.yaml` is created.
#### Inputs
_Fixture `taken.atago.yaml`:_
```text
version: "1"
suite:
  name: keep
scenarios:
  - name: keep me
    steps:
      - run:
          shell: true
          command: "exit 0"
```
#### When
```shell
${atago} init taken.atago.yaml
```
#### Then
- exit code is `3`
- stderr contains `already exists`
## atago self-hosting / init templates
Source: `test/e2e/atago/init_templates.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=cli]
#### When
```shell
${atago} init --template cli gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template cli gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=http]
#### When
```shell
${atago} init --template http gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template http gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=db]
#### When
```shell
${atago} init --template db gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template db gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=grpc]
#### When
```shell
${atago} init --template grpc gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template grpc gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=ssh]
#### When
```shell
${atago} init --template ssh gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template ssh gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=browser]
#### When
```shell
${atago} init --template browser gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template browser gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: every template scaffolds a schema-valid spec [template=services]
#### When
```shell
${atago} init --template services gen.atago.yaml
${atago} explain gen.atago.yaml
```
#### Then
- after `${atago} init --template services gen.atago.yaml`:
  - exit code is `0`
  - file `gen.atago.yaml` exists
- after `${atago} explain gen.atago.yaml`:
  - exit code is `0`
#### Generated artifacts
- `gen.atago.yaml`
### Scenario: list-templates names every runner family with a description
#### When
```shell
${atago} init --list-templates
```
#### Then
- exit code is `0`
- stdout contains `cli`, `http`, `db`, `grpc`, `ssh`, `browser`, `services`
- stdout contains `runs as-is`, `edit base_url first`
### Scenario: unknown template is a configuration error
#### When
```shell
${atago} init --template nope gen.atago.yaml
```
#### Then
- exit code is `3`
- stderr contains `unknown template`
### Scenario: the default cli template runs green
#### When
```shell
${atago} init --template cli cli.atago.yaml
${atago} run cli.atago.yaml
```
#### Then
- after `${atago} init --template cli cli.atago.yaml`:
  - exit code is `0`
- after `${atago} run cli.atago.yaml`:
  - exit code is `0`
### Scenario: the db template runs green with the bundled sqlite driver
#### When
```shell
${atago} init --template db db.atago.yaml
${atago} run db.atago.yaml
```
#### Then
- after `${atago} init --template db db.atago.yaml`:
  - exit code is `0`
- after `${atago} run db.atago.yaml`:
  - exit code is `0`
### Scenario: the services template runs green and exercises readiness + retry
_skipped on Windows_
#### When
```shell
${atago} init --template services services.atago.yaml
${atago} run services.atago.yaml
```
#### Then
- after `${atago} init --template services services.atago.yaml`:
  - exit code is `0`
- after `${atago} run services.atago.yaml`:
  - exit code is `0`
## atago self-hosting / json numeric comparators
Source: `test/e2e/atago/json_compare.atago.yaml`
### Scenario: gt and gte pass on a value at or above the bound
#### When
```shell
echo '{"count":3,"rate":0.5}'
```
#### Then
- stdout at `$.count` is `> 2`
- stdout at `$.count` is `>= 3`
### Scenario: lt and lte pass on a value at or below the bound
#### When
```shell
echo '{"count":3,"rate":0.5}'
```
#### Then
- stdout at `$.rate` is `< 1`
- stdout at `$.count` is `<= 3`
### Scenario: comparators work on a numeric string
#### When
```shell
echo '{"n":"7"}'
```
#### Then
- stdout at `$.n` is `>= 7`
### Scenario: comparators apply to rows and file json targets too
#### Given
- Fixture file `metrics.json` is created.
#### Inputs
_Fixture `metrics.json`:_
```text
{"processed": 1200, "errors": 0}
```
#### Then
- file `metrics.json` at `$.processed` is `> 1000`
- file `metrics.json` at `$.errors` is `<= 0`
### Scenario: a value below the gt bound fails the inner spec
#### Given
- Fixture file `cmp.atago.yaml` is created.
#### Inputs
_Fixture `cmp.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: count must exceed 5
    steps:
      - run:
          shell: true
          command: "echo '{\"count\":3}'"
      - assert:
          stdout:
            json:
              path: "$.count"
              gt: 5
```
#### When
```shell
${atago} run cmp.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `is not gt`
### Scenario: a non-numeric value cannot be compared and fails
#### Given
- Fixture file `cmp.atago.yaml` is created.
#### Inputs
_Fixture `cmp.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: name is not a number
    steps:
      - run:
          shell: true
          command: "echo '{\"name\":\"alice\"}'"
      - assert:
          stdout:
            json:
              path: "$.name"
              gt: 0
```
#### When
```shell
${atago} run cmp.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `not numeric`
## atago self-hosting / json and yaml matcher lists (#156)
Source: `test/e2e/atago/json_list.atago.yaml`
### Scenario: a file json list asserts several paths at once
#### Given
- Fixture file `starters.json` is created.
#### Inputs
_Fixture `starters.json`:_
```text
[
  {"name": "basei-starter", "default": true},
  {"name": "spec-starter"},
  {"name": "spec87bcd-starter"}
]
```
#### Then
- file `starters.json` at `$[0].name` equals `basei-starter`; at `$[0].default` equals `true`; at `$[2].name` equals `spec87bcd-starter`
### Scenario: a single mapping still works (backward compatible)
#### Given
- Fixture file `one.json` is created.
#### Inputs
_Fixture `one.json`:_
```text
{"id": 7}
```
#### Then
- file `one.json` at `$.id` equals `7`
### Scenario: a json list fails the inner spec when one listed path mismatches
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: one listed path is wrong
    steps:
      - fixture:
          file: data.json
          content: '{"a": 1, "b": 2}'
      - assert:
          file:
            path: data.json
            json:
              - { path: "$.a", equals: 1 }
              - { path: "$.b", equals: 999 }
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `did not equal`
### Scenario: a stdout json list against a JSON-producing command
_skipped on Windows_
#### When
```shell
echo '{"count": 3, "name": "ok"}'
```
#### Then
- stdout at `$.count` is `>= 2`; at `$.name` equals `ok`
### Scenario: a yaml list asserts several paths on one document
_skipped on Windows_
#### When
```shell
printf 'name: ada\nid: 42\n'
```
#### Then
- stdout YAML at `$.name` equals `ada`; YAML at `$.id` equals `42`
## atago self-hosting / json matcher boundary values
Source: `test/e2e/atago/json_matcher_edges.atago.yaml`
### Scenario: an array element is addressable by index
#### When
```shell
printf '{"items":[10,20,30]}'
```
#### Then
- stdout at `$.items[0]` equals `10`; at `$.items[2]` equals `30`
### Scenario: a top-level array reports its length
#### When
```shell
printf '[1,2,3,4,5]'
```
#### Then
- stdout at `$` has length 5
### Scenario: an empty array has length zero
#### When
```shell
printf '{"rows":[]}'
```
#### Then
- stdout at `$.rows` has length 0
### Scenario: the numeric comparators bound a value
#### When
```shell
printf '{"n":50}'
```
#### Then
- stdout at `$.n` is `> 49`; at `$.n` is `>= 50`; at `$.n` is `<= 50`; at `$.n` is `< 51`
### Scenario: a boolean value compares equal
#### When
```shell
printf '{"ok":true,"off":false}'
```
#### Then
- stdout at `$.ok` equals `true`; at `$.off` equals `false`
### Scenario: a floating-point value compares equal
#### When
```shell
printf '{"pi":3.14}'
```
#### Then
- stdout at `$.pi` equals `3.14`
### Scenario: a string carrying a quote compares equal
#### Given
- Fixture file `quoted.json` is created.
#### Inputs
_Fixture `quoted.json`:_
```text
{"s":"a\"b"}
```
#### When
```shell
cat quoted.json
```
#### Then
- stdout at `$.s` equals `a"b`
### Scenario: a deeply nested path resolves
#### When
```shell
printf '{"x":{"y":{"z":{"w":42}}}}'
```
#### Then
- stdout at `$.x.y.z.w` equals `42`
### Scenario: a path that selects nothing fails with a clear message
#### Given
- Fixture file `nopath.atago.yaml` is created.
#### Inputs
_Fixture `nopath.atago.yaml`:_
```text
version: "1"
suite: {name: nopath}
scenarios:
  - name: an absent key is a clean failure
    steps:
      - run: {shell: true, command: 'printf ''{"a":1}'''}
      - assert: {stdout: {json: [{path: "$.missing", equals: 1}]}}
```
#### When
```shell
${atago} run nopath.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `selected no value`
## atago self-hosting / line selector
Source: `test/e2e/atago/line.atago.yaml`
### Scenario: line selector narrows stdout to a single 1-based line
#### When
```shell
printf '[\n  {"id":1}\n]\n'
```
#### Then
- stdout equals an exact value
- stdout contains `"id":1`
- stdout equals an exact value
### Scenario: a trailing newline does not add a phantom final line
#### When
```shell
printf 'only-line\n'
```
#### Then
- stdout equals an exact value
### Scenario: an out-of-range line fails the inner spec
#### Given
- Fixture file `oor.atago.yaml` is created.
#### Inputs
_Fixture `oor.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: line 5 of a one-line stream
    steps:
      - run:
          command: echo single
      - assert:
          stdout:
            line: 5
            equals: nope
```
#### When
```shell
${atago} run oor.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `out of range`
## atago self-hosting / stream text matchers fold CRLF
Source: `test/e2e/atago/line_endings.atago.yaml`
### Scenario: equals folds a CRLF body to its LF form
#### When
```shell
printf 'first\r\nsecond\r\n'
```
#### Then
- stdout equals an exact value
#### Expected output
_expected stdout:_
```text
first
second
```
### Scenario: equals tolerates the phantom trailing CRLF
#### When
```shell
printf 'only\r\n'
```
#### Then
- stdout equals an exact value
### Scenario: contains folds CRLF for a multi-line needle
#### When
```shell
printf 'alpha\r\nbeta\r\ngamma\r\n'
```
#### Then
- stdout contains `alpha
beta`
### Scenario: contains authored with CRLF matches LF-folded output
#### When
```shell
printf 'alpha\r\nbeta\r\n'
```
#### Then
- stdout contains `alpha
beta`
### Scenario: contains list every multi-line element folds
#### When
```shell
printf 'a\r\nb\r\nc\r\nd\r\n'
```
#### Then
- stdout contains `a
b`, `c
d`
### Scenario: matches anchors a line over CRLF with the multiline flag
#### When
```shell
printf 'hello\r\nworld\r\n'
```
#### Then
- stdout matches `/(?m)^world$/`
### Scenario: matches a literal newline in the pattern over CRLF
#### When
```shell
printf 'up\r\ndown\r\n'
```
#### Then
- stdout matches `/up
down/`
### Scenario: not_contains stays clear of an absent multi-line needle
#### When
```shell
printf 'red\r\ngreen\r\n'
```
#### Then
- stdout does not contain `red
blue`
### Scenario: not_matches passes for an anchored line that is absent
#### When
```shell
printf 'north\r\nsouth\r\n'
```
#### Then
- stdout does not match `/(?m)^east$/`
### Scenario: the line selector strips the trailing CR
#### When
```shell
printf 'header\r\npayload\r\n'
```
#### Then
- stdout equals an exact value
### Scenario: json parses a CRLF-formatted document
#### When
```shell
printf '{\r\n"count":3\r\n}\r\n'
```
#### Then
- stdout at `$.count` equals `3`
### Scenario: folding does not make an absent multi-line needle match
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: an absent multi-line needle still fails
    steps:
      - run:
          shell: true
          command: printf 'one\r\ntwo\r\n'
      - assert:
          stdout:
            contains: "one\nMISSING"
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `was not present`
## atago self-hosting / list
Source: `test/e2e/atago/list.atago.yaml`
### Scenario: list surfaces suites, scenarios, tags, and gates
#### Given
- Fixture file `sample.atago.yaml` is created.
#### Inputs
_Fixture `sample.atago.yaml`:_
```text
version: "1"
suite:
  name: demo
scenarios:
  - name: tagged scenario
    tags: [smoke, fast]
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
  - name: gated scenario
    skip: {env: ATAGO_SKIP_DEMO}
    steps:
      - run: {shell: true, command: "echo hi > out.txt"}
      - assert: {file: {path: out.txt, exists: true}}
```
#### When
```shell
${atago} list sample.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `tagged scenario`, `smoke`, `skip:env=ATAGO_SKIP_DEMO`
### Scenario: list --json is a stable machine contract
#### Given
- Fixture file `sample.atago.yaml` is created.
#### Inputs
_Fixture `sample.atago.yaml`:_
```text
version: "1"
suite:
  name: demo
scenarios:
  - name: only scenario
    steps:
      - run: {command: "true"}
      - assert: {exit_code: 0}
```
#### When
```shell
${atago} list --json sample.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.schema_version` equals `1`
- stdout at `$.scenarios[0].scenario` equals `only scenario`
## atago self-hosting / loader rejects malformed specs
Source: `test/e2e/atago/loader_errors.atago.yaml`
### Scenario: an empty scenario list is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios: []
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `must contain at least one scenario`
### Scenario: a wrong version string is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "9"
suite: {name: x}
scenarios: [{name: a, steps: [{run: {command: echo}}]}]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `version must be "1"`
### Scenario: an unknown top-level field is rejected with its position
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenariosss: []
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `unknown field "scenariosss"`
### Scenario: a step that sets two actions is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
        fixture: {file: f, content: c}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `must set exactly one action`
### Scenario: a stream assertion with no matcher is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {stdout: {}}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `must set at least one matcher`
### Scenario: combining equals with another matcher is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {stdout: {equals: hi, contains: h}}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `cannot be combined with another matcher`
### Scenario: a line index below one is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {stdout: {line: 0, equals: x}}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `line must be >= 1`
### Scenario: combining a line selector with json is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {stdout: {line: 1, json: [{path: "$.k", equals: 1}]}}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `cannot be combined with json/yaml/snapshot`
### Scenario: a duplicate scenario name is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - {name: dup, steps: [{run: {command: echo}}]}
  - {name: dup, steps: [{run: {command: echo}}]}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `duplicate scenario name "dup"`
### Scenario: an empty run command is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios: [{name: a, steps: [{run: {command: ""}}]}]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `run.command is required`
### Scenario: an unparseable timeout is rejected with an example
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios: [{name: a, steps: [{run: {command: echo, timeout: "soon"}}]}]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `is not a valid duration`
### Scenario: a fixture with two content sources is rejected
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios: [{name: a, steps: [{fixture: {file: f, content: c, base64: "QQ=="}}]}]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `set only one of content, base64, from, or symlink`
### Scenario: an absolute changes glob is rejected as not workdir-relative
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios:
  - name: a
    steps:
      - run: {command: echo}
      - assert: {changes: {created: ["/etc/passwd"]}}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `must be workdir-relative, not absolute`
### Scenario: the inline stdin form is a scalar, not a mapping key
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite: {name: x}
scenarios: [{name: a, steps: [{run: {command: cat, stdin: {inline: hi}}}]}]
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `unknown key "inline"`
## atago self-hosting / manifest
Source: `test/e2e/atago/manifest.atago.yaml`
### Scenario: manifest emits a stable JSON summary without running the spec
#### Given
- Fixture file `sample.atago.yaml` is created.
#### Inputs
_Fixture `sample.atago.yaml`:_
```text
version: "1"
suite:
  name: demo
secrets:
  - TOKEN
scenarios:
  - name: "greets ${who}"
    tags: [smoke]
    matrix:
      - { who: Alice }
      - { who: Bob }
    steps:
      - run:
          command: echo ${who}
      - assert:
          stdout:
            contains: ${who}
          file:
            path: out.txt
            exists: true
```
#### When
```shell
${atago} manifest sample.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.schema_version` equals `1`
- stdout at `$.specs[0].suite` equals `demo`
- stdout at `$.specs[0].secrets[0]` equals `TOKEN`
- stdout at `$.specs[0].scenarios` has length 2
- stdout at `$.specs[0].scenarios[0].vars.who` equals `Alice`
- stdout at `$.specs[0].scenarios[0].generates[0]` equals `out.txt`
- stdout at `$.specs[0].source.line` equals `3`
- stdout at `$.specs[0].scenarios[0].source.line` equals `7`
- stdout at `$.specs[0].scenarios[1].source.line` equals `7`
- stdout at `$.specs[0].scenarios[0].steps[0].source.line` equals `13`
### Scenario: manifest does not execute the spec's commands
#### Given
- Fixture file `side_effect.atago.yaml` is created.
#### Inputs
_Fixture `side_effect.atago.yaml`:_
```text
version: "1"
suite:
  name: side
scenarios:
  - name: would write a file
    steps:
      - run:
          shell: true
          command: touch executed.marker
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} manifest side_effect.atago.yaml
```
#### Then
- exit code is `0`
- file `executed.marker` does not exist
## atago self-hosting / matrix scenarios
Source: `test/e2e/atago/matrix.atago.yaml`
### Scenario: matrix expands into one scenario per row
#### Given
- Fixture file `matrix.atago.yaml` is created.
#### Inputs
_Fixture `matrix.atago.yaml`:_
```text
version: "1"
suite:
  name: greetings
scenarios:
  - name: "greets ${who}"
    matrix:
      - { who: Alice }
      - { who: Bob }
    steps:
      - run:
          shell: true
          command: echo ${who}
      - assert:
          stdout:
            contains: ${who}
```
#### When
```shell
${atago} run --report junit matrix.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `name="greets Alice"`, `name="greets Bob"`
### Scenario: matrix without a templated name gets a deterministic suffix
#### Given
- Fixture file `suffix.atago.yaml` is created.
#### Inputs
_Fixture `suffix.atago.yaml`:_
```text
version: "1"
suite:
  name: suffixed
scenarios:
  - name: row
    matrix:
      - { n: "1" }
      - { n: "2" }
    steps:
      - run:
          shell: true
          command: echo ${n}
      - assert:
          stdout:
            contains: ${n}
```
#### When
```shell
${atago} run --report junit suffix.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `name="row [n=1]"`, `name="row [n=2]"`
### Scenario: stdout_to expands a matrix variable into the redirect target [who=alice]
#### When
```shell
printf hello-alice
```
#### Then
- exit code is `0`
- file `out-alice.txt` contains `hello-alice`
#### Generated artifacts
- `out-${who}.txt`
### Scenario: stdout_to expands a matrix variable into the redirect target [who=bob]
#### When
```shell
printf hello-bob
```
#### Then
- exit code is `0`
- file `out-bob.txt` contains `hello-bob`
#### Generated artifacts
- `out-${who}.txt`
## atago self-hosting / matrix expansion boundary values
Source: `test/e2e/atago/matrix_edges.atago.yaml`
### Scenario: each row substitutes into the scenario name
#### Given
- Fixture file `names.atago.yaml` is created.
#### Inputs
_Fixture `names.atago.yaml`:_
```text
version: "1"
suite: {name: names}
scenarios:
  - name: "case ${n}"
    matrix:
      - {n: "1"}
      - {n: "2"}
      - {n: "3"}
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json names.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 3; at `$.suites[0].scenarios[0].name` equals `case 1`; at `$.suites[0].scenarios[2].name` equals `case 3`
### Scenario: a row with several variables substitutes all of them
#### Given
- Fixture file `multi.atago.yaml` is created.
#### Inputs
_Fixture `multi.atago.yaml`:_
```text
version: "1"
suite: {name: multi}
scenarios:
  - name: "${who} speaks ${lang}"
    matrix:
      - {who: Alice, lang: en}
      - {who: Bob, lang: fr}
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json multi.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios[0].name` equals `Alice speaks en`; at `$.suites[0].scenarios[1].name` equals `Bob speaks fr`
### Scenario: a single-row matrix expands to exactly one scenario
#### Given
- Fixture file `single.atago.yaml` is created.
#### Inputs
_Fixture `single.atago.yaml`:_
```text
version: "1"
suite: {name: single}
scenarios:
  - name: "only ${k}"
    matrix:
      - {k: solo}
    steps: [{run: {shell: true, command: "exit 0"}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --ci --report json single.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].scenarios` has length 1; at `$.suites[0].scenarios[0].name` equals `only solo`
### Scenario: an empty matrix row list is a load-time error
#### Given
- Fixture file `emptymx.atago.yaml` is created.
#### Inputs
_Fixture `emptymx.atago.yaml`:_
```text
version: "1"
suite: {name: emptymx}
scenarios:
  - name: "e ${n}"
    matrix: []
    steps: [{run: {shell: true, command: "exit 0"}}]
```
#### When
```shell
${atago} run emptymx.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `matrix must contain at least one row`
### Scenario: rows that expand to the same name are rejected as duplicates
#### Given
- Fixture file `dupmx.atago.yaml` is created.
#### Inputs
_Fixture `dupmx.atago.yaml`:_
```text
version: "1"
suite: {name: dupmx}
scenarios:
  - name: "dup ${n}"
    matrix:
      - {n: same}
      - {n: same}
    steps: [{run: {shell: true, command: "exit 0"}}]
```
#### When
```shell
${atago} run dupmx.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `duplicate scenario name "dup same"`
## atago self-hosting / mock http server (offline API-client testing)
Source: `test/e2e/atago/mock_server.atago.yaml`
### Scenario: count, header, and body-json asserts pass against a real client
#### Given
- Stub HTTP server `api` serves 1 canned route(s) at `${api.url}` and records every request (#24).
- Fixture file `client.atago.yaml` is created.
#### Inputs
_Fixture `client.atago.yaml`:_
```text
version: "1"
suite:
  name: client
runners:
  api:
    type: http
    base_url: ${api.url}
scenarios:
  - name: post a report
    steps:
      - http:
          runner: api
          method: POST
          path: /v1/reports
          header: { Authorization: "Bearer tok-123" }
          json: { title: "report" }
      - assert:
          status: 201
```
#### When
```shell
${atago} run client.atago.yaml
```
#### Then
- exit code is `0`
- mock `api` received `POST /v1/reports` exactly 1 time(s)
### Scenario: a failing count summarizes the recorded requests
#### Given
- Stub HTTP server `stub` serves 1 canned route(s) at `${stub.url}` and records every request (#24).
- Fixture file `outer.atago.yaml` is created.
#### Inputs
_Fixture `outer.atago.yaml`:_
```text
version: "1"
suite:
  name: outer
runners:
  api:
    type: http
    base_url: ${stub.url}
scenarios:
  - name: wrong path then failing count
    mock_servers:
      - name: inner
        routes:
          - method: GET
            path: /right
    steps:
      - http:
          runner: api
          method: GET
          path: /wrong
      - assert:
… (truncated, 6 more lines)
```
#### When
```shell
${atago} run outer.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `1 request for /right`, `0 matching of 0 recorded`
### Scenario: an unknown mock name in an assert is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: wrong name
    mock_servers:
      - name: api
        routes: [{method: GET, path: /}]
    steps:
      - run: {command: echo hi}
      - assert:
          mock: {name: apo, count: 1}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `not a declared mock server (declared: api)`
## atago self-hosting / combined stream matchers
Source: `test/e2e/atago/multi_matcher.atago.yaml`
### Scenario: contains and not_contains hold together
#### When
```shell
echo "hello world"
```
#### Then
- stdout contains `hello`
### Scenario: matches and not_matches hold together
#### When
```shell
echo "release 1.2.3"
```
#### Then
- stdout matches `/[0-9]+\.[0-9]+\.[0-9]+/`
### Scenario: all four text matchers compose
#### When
```shell
echo "Alice and Bob"
```
#### Then
- stdout contains `Alice`, `Bob`
### Scenario: a combined matcher composes with a line selector
#### When
```shell
printf 'first line\nsecond line\n'
```
#### Then
- stdout contains `second`
### Scenario: a failing member fails the inner spec and names the offender
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: contains holds but not_contains does not
    steps:
      - run:
          command: echo "hello goodbye"
      - assert:
          stdout:
            contains: hello
            not_contains: goodbye
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `goodbye`
### Scenario: mixing a whole-stream matcher with a text matcher is a load error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: equals cannot combine
    steps:
      - run:
          command: echo hi
      - assert:
          stdout:
            equals: hi
            contains: h
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `cannot be combined with another matcher`
## atago self-hosting / not_equals matcher
Source: `test/e2e/atago/not_equals.atago.yaml`
### Scenario: not_equals passes when stdout differs from the given text
#### When
```shell
echo Bob
```
#### Then
- stdout does not equal an exact value
### Scenario: not_equals is trailing-newline tolerant like equals
#### When
```shell
echo hello
```
#### Then
- stdout does not equal an exact value
### Scenario: not_equals composes with a line selector
#### When
```shell
printf 'first\nsecond\n'
```
#### Then
- stdout does not equal an exact value
### Scenario: not_equals fails the inner spec when the text matches exactly
#### Given
- Fixture file `ne.atago.yaml` is created.
#### Inputs
_Fixture `ne.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: stdout equals the forbidden text
    steps:
      - run:
          command: echo same
      - assert:
          stdout:
            not_equals: same
```
#### When
```shell
${atago} run ne.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `unexpectedly equaled`
## atago self-hosting / parallel
Source: `test/e2e/atago/parallel.atago.yaml`
### Scenario: parallel run passes and stays deterministic
#### Given
- Fixture file `many.atago.yaml` is created.
#### Inputs
_Fixture `many.atago.yaml`:_
```text
version: "1"
suite:
  name: many
scenarios:
  - name: a
    steps: [{run: {shell: true, command: echo a}}, {assert: {stdout: {contains: a}}}]
  - name: b
    steps: [{run: {shell: true, command: echo b}}, {assert: {stdout: {contains: b}}}]
  - name: c
    steps: [{run: {shell: true, command: echo c}}, {assert: {stdout: {contains: c}}}]
```
#### When
```shell
${atago} run --parallel 3 many.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `3 passed`
### Scenario: fail-fast stops after the first failure
#### Given
- Fixture file `ff.atago.yaml` is created.
#### Inputs
_Fixture `ff.atago.yaml`:_
```text
version: "1"
suite:
  name: ff
scenarios:
  - name: boom
    steps: [{run: {shell: true, command: "exit 1"}}, {assert: {exit_code: 0}}]
  - name: never
    steps: [{run: {shell: true, command: echo hi}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --parallel 1 --fail-fast ff.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `1 skipped`
## atago self-hosting / forward-slash spec paths resolve on every OS
Source: `test/e2e/atago/paths_portable.atago.yaml`
### Scenario: stdout_to creates a nested parent directory
#### When
```shell
echo produced
```
#### Then
- exit code is `0`
- file `out/logs/result.txt` contains `produced`
#### Generated artifacts
- `out/logs/result.txt`
### Scenario: stderr_to creates its own nested parent directory
#### When
```shell
echo oops 1>&2
```
#### Then
- exit code is `0`
- file `errs/deep/err.txt` contains `oops`
#### Generated artifacts
- `errs/deep/err.txt`
### Scenario: a fixture at a nested forward-slash path is created and addressable
#### Given
- Fixture file `data/config/app.json` is created.
#### Inputs
_Fixture `data/config/app.json`:_
```text
{"k":1}
```
#### Then
- file `data/config/app.json` at `$.k` equals `1`
### Scenario: a file assert reaches a deeply nested fixture by forward-slash path
#### Given
- Fixture file `a/b/c/leaf.txt` is created.
#### Inputs
_Fixture `a/b/c/leaf.txt`:_
```text
at the bottom
```
#### Then
- file `a/b/c/leaf.txt` contains `at the bottom`
### Scenario: a dir assert addresses a nested tree and child by forward-slash path
#### Given
- Fixture file `pkg/mod/one.go` is created.
- Fixture file `pkg/mod/two.go` is created.
- Fixture file `pkg/mod/sub/three.go` is created.
#### Inputs
_Fixture `pkg/mod/one.go`:_
```text
package mod
```
_Fixture `pkg/mod/two.go`:_
```text
package mod
```
_Fixture `pkg/mod/sub/three.go`:_
```text
package sub
```
#### Then
- dir `pkg/mod` exists, contains `one.go`, contains `two.go`, contains `sub/three.go`, does not contain `missing.go`
### Scenario: equals_file compares two files addressed by forward-slash paths
#### Given
- Fixture file `golden/expected.bin` is created.
- Fixture file `build/actual.bin` is created.
#### Then
- file `build/actual.bin` is byte-identical to `golden/expected.bin`
### Scenario: a redirect path may not escape the workdir via a nested traversal
#### Given
- Fixture file `probe.atago.yaml` is created.
#### Inputs
_Fixture `probe.atago.yaml`:_
```text
version: "1"
suite:
  name: escape
scenarios:
  - name: nested traversal is rejected
    steps:
      - run:
          shell: true
          command: echo x
          stdout_to: sub/../../escape.txt
```
#### When
```shell
${atago} run probe.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `escapes the scenario workdir`
## atago self-hosting / pdf assertion
Source: `test/e2e/atago/pdf.atago.yaml`
### Scenario: pdf assertions cover page count, metadata, and text
#### Given
- Fixture file `report.pdf` is created.
#### Inputs
_Fixture `report.pdf`:_
```text
%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 46 >>
stream
BT /F1 24 Tf 72 700 Td (Hello atago report) Tj ET
endstream
endobj
6 0 obj
<< /Title (Quarterly Report) /Author (atago) >>
endobj
trailer
… (truncated, 2 more lines)
```
#### Then
- pdf `report.pdf` 1 pages, >= 1 pages, <= 3 pages, author contains `atago`, title contains `Quarterly`, text contains `Hello atago report`
### Scenario: a non-pdf file fails the pdf target
#### Given
- Fixture file `notpdf.txt` is created.
#### Inputs
_Fixture `notpdf.txt`:_
```text
just text
```
#### When
```shell
${atago} version
```
#### Then
- exit code is `0`
## atago self-hosting / pty
Source: `test/e2e/atago/pty.atago.yaml`
### Scenario: a pty step sees a terminal where a run step sees a pipe
_skipped on Windows_
#### Given
- Fixture file `tty.atago.yaml` is created.
#### Inputs
_Fixture `tty.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: tty
    steps:
      - pty:
          shell: true
          command: 'if [ -t 0 ]; then echo saw-a-tty; else echo saw-a-pipe; fi'
      - assert:
          exit_code: 0
          stdout:
            contains: saw-a-tty
```
#### When
```shell
${atago} run tty.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: a never-matching expect fails with the pattern in the block
_skipped on Windows_
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: waits forever
    steps:
      - pty:
          command: cat
          timeout: 2s
          session:
            - expect: "prompt-that-never-comes"
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `pty expect /prompt-that-never-comes/`, `never appeared in the terminal transcript`
### Scenario: named keys transmit their documented bytes and ctrl-c aborts
_skipped on Windows_
#### When
```shell
# interactive (pty): cat -v
# interactive (pty): trap 'exit 130' INT; echo waiting; while true; do sleep 0.1; done
```
#### Then
- exit code is `0`
- stdout contains `^[[B`
- exit code is `130`
### Scenario: an unknown key name is a load-time error listing the vocabulary
#### Given
- Fixture file `badkey.atago.yaml` is created.
#### Inputs
_Fixture `badkey.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: typo'd key
    steps:
      - pty:
          command: cat
          session:
            - send: { key: entr }
```
#### When
```shell
${atago} run badkey.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `not a supported key (supported: enter, tab`
### Scenario: screen asserts see the final frame where the transcript sees history
_skipped on Windows_
#### When
```shell
# interactive (pty): printf 'loading...'; printf 'done.      
'
```
#### Then
- rendered screen equals an exact value
- rendered screen does not contain `loading`
- stdout contains `loading`
### Scenario: a screen snapshot round-trips through update and compare
_skipped on Windows_
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: menu frame
    steps:
      - pty:
          shell: true
          command: "printf '\\033[2J\\033[HMain Menu\\r\\n> Settings\\r\\n'"
          rows: 10
          cols: 40
      - assert:
          screen:
            snapshot: snapshots/menu.txt
```
#### When
```shell
${atago} run --update-snapshots inner.atago.yaml
${atago} run inner.atago.yaml
```
#### Then
- after `${atago} run --update-snapshots inner.atago.yaml`:
  - exit code is `0`
  - file `snapshots/menu.txt` contains `> Settings`
- after `${atago} run inner.atago.yaml`:
  - exit code is `0`
### Scenario: a screen assert without a pty step is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: no pty
    steps:
      - run: {command: echo hi}
      - assert:
          screen: {contains: hi}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `requires a preceding pty step`
### Scenario: a send referencing an undefined variable is an execution error, not typed literally
_skipped on Windows_
#### Given
- Fixture file `typo.atago.yaml` is created.
#### Inputs
_Fixture `typo.atago.yaml`:_
```text
version: "1"
suite:
  name: typo
scenarios:
  - name: typo in a send variable
    steps:
      - pty:
          command: cat
          timeout: 5s
          session:
            - send: "${no_such_var}\n"
```
#### When
```shell
${atago} run typo.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `no variable with that name is defined`, `$${no_such_var}`
## atago self-hosting / pty (portable)
Source: `test/e2e/atago/pty_portable.atago.yaml`
### Scenario: a pty step starts a command, captures its output, and reports exit 0
#### When
```shell
# interactive (pty): echo hello from a pty
```
#### Then
- exit code is `0`
- stdout contains `hello from a pty`
### Scenario: a pty step surfaces a command's non-zero exit code
#### When
```shell
# interactive (pty): echo bye && exit 3
```
#### Then
- exit code is `3`
### Scenario: sequential expects match successive output in declaration order
#### When
```shell
# interactive (pty): echo first && echo second && echo third
```
#### Then
- exit code is `0`
- stdout contains `first`, `third`
### Scenario: an expect pattern is a regular expression, not a literal
#### When
```shell
# interactive (pty): echo item-42-done
```
#### Then
- exit code is `0`
### Scenario: a screen assert reads the rendered frame sized by rows and cols
#### When
```shell
# interactive (pty): echo rendered line
```
#### Then
- exit code is `0`
- rendered screen contains `rendered line`
### Scenario: a pty step drives the atago binary directly with no shell
#### When
```shell
# interactive (pty): ${atago} version
```
#### Then
- exit code is `0`
- stdout contains `atago`
### Scenario: a pty drives atago running an inner spec to a green result
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: echo
    steps:
      - run:
          shell: true
          command: echo inner-ok
      - assert:
          exit_code: 0
          stdout:
            contains: inner-ok
```
#### When
```shell
# interactive (pty): ${atago} run inner.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: a never-matching expect fails and names the pattern in the transcript
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: waits for output that never comes
    steps:
      - pty:
          shell: true
          command: echo present
          timeout: 2s
          session:
            - expect: "absent-forever"
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `pty expect /absent-forever/`, `never appeared in the terminal transcript`
## atago self-hosting / record (spec skeleton from an observed run)
Source: `test/e2e/atago/record.atago.yaml`
### Scenario: record then run round-trips green
#### When
```shell
${atago} record --out recorded.atago.yaml -- ${atago} version
${atago} run recorded.atago.yaml
```
#### Then
- after `${atago} record --out recorded.atago.yaml -- ${atago} version`:
  - exit code is `0`
  - stderr contains `wrote recorded.atago.yaml`
  - file `recorded.atago.yaml` contains `contains: atago`
- after `${atago} run recorded.atago.yaml`:
  - exit code is `0`
  - stdout contains `1 passed`
### Scenario: refusing to overwrite without --force
#### Given
- Fixture file `existing.atago.yaml` is created.
#### Inputs
_Fixture `existing.atago.yaml`:_
```text
precious
```
#### When
```shell
${atago} record --out existing.atago.yaml -- ${atago} version
${atago} record --force --out existing.atago.yaml -- ${atago} version
```
#### Then
- after `${atago} record --out existing.atago.yaml -- ${atago} version`:
  - exit code is `3`
  - stderr contains `use --force to overwrite`
  - file `existing.atago.yaml` contains `precious`
- after `${atago} record --force --out existing.atago.yaml -- ${atago} version`:
  - exit code is `0`
  - file `existing.atago.yaml` contains `exit_code: 0`
### Scenario: record --pty refuses an existing --out before driving the session
#### Given
- Fixture file `taken.atago.yaml` is created.
#### Inputs
_Fixture `taken.atago.yaml`:_
```text
precious
```
#### When
```shell
${atago} record --pty --out taken.atago.yaml -- echo hi
```
#### Then
- exit code is `3`
- stderr contains `use --force to overwrite`
- file `taken.atago.yaml` contains `precious`
### Scenario: created files become exists asserts (shell mode)
_skipped on Windows_
#### When
```shell
${atago} record --shell --out gen.atago.yaml -- 'echo made > out.txt; echo done'
${atago} run gen.atago.yaml
```
#### Then
- after `${atago} record --shell --out gen.atago.yaml -- 'echo made > out.txt; echo done'`:
  - exit code is `0`
  - file `gen.atago.yaml` contains `path: out.txt`, `exists: true`
  - file `gen.atago.yaml` contains `shell: true`
- after `${atago} run gen.atago.yaml`:
  - exit code is `0`
### Scenario: snapshot mode writes a golden the run then matches
_skipped on Windows_
#### When
```shell
${atago} record --snapshot --out snapdemo.atago.yaml -- echo stable output
${atago} run snapdemo.atago.yaml
```
#### Then
- after `${atago} record --snapshot --out snapdemo.atago.yaml -- echo stable output`:
  - exit code is `0`
  - file `snapdemo.atago.yaml` contains `snapshot: snapshots/snapdemo.stdout.txt`
  - file `snapshots/snapdemo.stdout.txt` contains `stable output`
- after `${atago} run snapdemo.atago.yaml`:
  - exit code is `0`
  - stdout contains `1 passed`
### Scenario: no command is a usage error
#### When
```shell
${atago} record
```
#### Then
- exit code is `3`
- stderr contains `no command given`
### Scenario: argv boundaries survive spaced arguments
_skipped on Windows_
#### When
```shell
${atago} record --out spaced.atago.yaml -- printf %s 'hello world'
${atago} run spaced.atago.yaml
```
#### Then
- after `${atago} record --out spaced.atago.yaml -- printf %s 'hello world'`:
  - exit code is `0`
  - file `spaced.atago.yaml` contains `contains: hello world`
- after `${atago} run spaced.atago.yaml`:
  - exit code is `0`
  - stdout contains `1 passed`
### Scenario: a shell metacharacter argument stays one token
_skipped on Windows_
#### When
```shell
${atago} record --out meta.atago.yaml -- printf %s "foo|bar"
${atago} run meta.atago.yaml
```
#### Then
- after `${atago} record --out meta.atago.yaml -- printf %s "foo|bar"`:
  - exit code is `0`
  - file `meta.atago.yaml` contains `contains: foo|bar`
- after `${atago} run meta.atago.yaml`:
  - exit code is `0`
  - stdout contains `1 passed`
### Scenario: record --pty records a live session and the generated spec replays green
_skipped on Windows_
#### When
```shell
# interactive (pty): ${atago} record --pty --out generated.atago.yaml -- sh -c 'printf PROMPT; read n; echo hi-$n'
${atago} run generated.atago.yaml
```
#### Then
- exit code is `0`
- file `generated.atago.yaml` contains `- pty:`, `- send:`
- exit code is `0`
- stdout contains `1 passed`
### Scenario: record --pty of a no-input command yields a session-less spec that replays green
_skipped on Windows_
#### When
```shell
# interactive (pty): ${atago} record --pty --out echo.atago.yaml -- echo done
${atago} run echo.atago.yaml
```
#### Then
- exit code is `0`
- file `echo.atago.yaml` contains `- pty:`, `command: echo done`
- file `echo.atago.yaml` is checked
- exit code is `0`
- stdout contains `1 passed`
### Scenario: a prompt with regex metacharacters is escaped in the generated expect
_skipped on Windows_
#### When
```shell
# interactive (pty): ${atago} record --pty --out meta.atago.yaml -- sh -c 'printf "Continue? (y/n): "; read a; echo got-$a'
${atago} run meta.atago.yaml
```
#### Then
- exit code is `0`
- file `meta.atago.yaml` contains `expect: "Continue\\? \\(y/n\\):"`, `- send:`
- exit code is `0`
- stdout contains `1 passed`
### Scenario: recorded text containing dollar-brace round-trips as literal text
_skipped on Windows_
#### When
```shell
${atago} record --out dollar.atago.yaml -- printf %s 'literal $${HOME} here'
${atago} run dollar.atago.yaml
```
#### Then
- after `${atago} record --out dollar.atago.yaml -- printf %s 'literal $${HOME} here'`:
  - exit code is `0`
  - file `dollar.atago.yaml` contains `$$${HOME}`
- after `${atago} run dollar.atago.yaml`:
  - exit code is `0`
  - stdout contains `1 passed`
### Scenario: a recorded secret placeholder replays green with the env set and is guarded when unset
_skipped on Windows_
#### Given
- Environment variables are set: ATAGO_SECRET_1.
#### When
```shell
# interactive (pty): ${atago} record --pty --out sec.atago.yaml -- sh -c 'stty -echo; printf "Password: "; read pw; stty echo; printf "\naccepted\n"'
${atago} run sec.atago.yaml
${atago} run sec.atago.yaml
```
#### Then
- exit code is `0`
- file `sec.atago.yaml` contains `${env:ATAGO_SECRET_1}`
- file `sec.atago.yaml` is checked
- after `${atago} run sec.atago.yaml`:
  - exit code is `0`
  - stdout contains `1 passed`
- after `${atago} run sec.atago.yaml`:
  - exit code is `4`
  - stdout contains `ATAGO_SECRET_1 is not set`
### Scenario: record --pty of a never-exiting program times out instead of hanging
_skipped on Windows_
#### When
```shell
${atago} record --pty --timeout 2s --out wedged.atago.yaml -- tail -f /dev/null
```
#### Then
- exit code is `4`
- stderr contains `did not exit within 2s`
- stderr contains `use --timeout to adjust`
- file `wedged.atago.yaml` contains `- pty:`
## atago self-hosting / report formats agree on outcomes
Source: `test/e2e/atago/report_formats.atago.yaml`
### Scenario: json report carries per-scenario verdicts and a failures array
#### Given
- Fixture file `mixed.atago.yaml` is created.
#### Inputs
_Fixture `mixed.atago.yaml`:_
```text
version: "1"
suite:
  name: mixed
scenarios:
  - name: alpha passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: beta passes
    steps:
      - run: {shell: true, command: "echo hi"}
      - assert: {stdout: {contains: hi}}
  - name: gamma fails
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 9}
  - name: delta is skipped
    skip: {env: PATH}
    steps:
      - run: {shell: true, command: "exit 0"}
```
#### When
```shell
${atago} run --ci --report json mixed.atago.yaml
```
#### Then
- exit code is `1`
- stdout at `$.suites[0].status` equals `failed`; at `$.suites[0].scenarios[0].status` equals `passed`; at `$.suites[0].scenarios[2].status` equals `failed`; at `$.suites[0].scenarios[3].status` equals `skipped`; at `$.suites[0].failures[0].scenario` equals `gamma fails`
### Scenario: junit report tallies tests, failures, skipped, and errors
#### Given
- Fixture file `mixed.atago.yaml` is created.
#### Inputs
_Fixture `mixed.atago.yaml`:_
```text
version: "1"
suite:
  name: mixed
scenarios:
  - name: alpha passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: beta passes
    steps:
      - run: {shell: true, command: "echo hi"}
      - assert: {stdout: {contains: hi}}
  - name: gamma fails
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 9}
  - name: delta is skipped
    skip: {env: PATH}
    steps:
      - run: {shell: true, command: "exit 0"}
```
#### When
```shell
${atago} run --ci --report junit mixed.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `tests="4"`, `failures="1"`, `errors="0"`, `skipped="1"`, `<skipped`
### Scenario: tap report emits the plan, a not ok line, and a SKIP directive
#### Given
- Fixture file `mixed.atago.yaml` is created.
#### Inputs
_Fixture `mixed.atago.yaml`:_
```text
version: "1"
suite:
  name: mixed
scenarios:
  - name: alpha passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: beta passes
    steps:
      - run: {shell: true, command: "echo hi"}
      - assert: {stdout: {contains: hi}}
  - name: gamma fails
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 9}
  - name: delta is skipped
    skip: {env: PATH}
    steps:
      - run: {shell: true, command: "exit 0"}
```
#### When
```shell
${atago} run --ci --report tap mixed.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `TAP version 13`, `1..4`, `not ok 3 - mixed / gamma fails`, `# SKIP`
### Scenario: gha report annotates the failure and summarizes the counts
#### Given
- Fixture file `mixed.atago.yaml` is created.
#### Inputs
_Fixture `mixed.atago.yaml`:_
```text
version: "1"
suite:
  name: mixed
scenarios:
  - name: alpha passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: beta passes
    steps:
      - run: {shell: true, command: "echo hi"}
      - assert: {stdout: {contains: hi}}
  - name: gamma fails
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 9}
  - name: delta is skipped
    skip: {env: PATH}
    steps:
      - run: {shell: true, command: "exit 0"}
```
#### When
```shell
${atago} run --ci --report gha mixed.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `::error title=mixed / gamma fails::`, `::notice title=atago::4 scenarios: 2 passed, 1 failed, 0 errored, 1 skipped`
### Scenario: console report prints the same counts in its summary line
#### Given
- Fixture file `mixed.atago.yaml` is created.
#### Inputs
_Fixture `mixed.atago.yaml`:_
```text
version: "1"
suite:
  name: mixed
scenarios:
  - name: alpha passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: beta passes
    steps:
      - run: {shell: true, command: "echo hi"}
      - assert: {stdout: {contains: hi}}
  - name: gamma fails
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 9}
  - name: delta is skipped
    skip: {env: PATH}
    steps:
      - run: {shell: true, command: "exit 0"}
```
#### When
```shell
${atago} run --ci --report console mixed.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `4 scenarios: 2 passed, 1 failed, 0 errored, 1 skipped`
### Scenario: an all-passing run reports a zero-failure suite and exits zero
#### Given
- Fixture file `allpass.atago.yaml` is created.
#### Inputs
_Fixture `allpass.atago.yaml`:_
```text
version: "1"
suite:
  name: allpass
scenarios:
  - name: only passes
    steps:
      - run: {shell: true, command: "echo ok"}
      - assert: {stdout: {contains: ok}}
```
#### When
```shell
${atago} run --ci --report json allpass.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.suites[0].status` equals `passed`; at `$.suites[0].scenarios[0].status` equals `passed`
### Scenario: an errored step is counted as an error, not a failure, across formats
#### Given
- Fixture file `errored.atago.yaml` is created.
#### Inputs
_Fixture `errored.atago.yaml`:_
```text
version: "1"
suite:
  name: errored
scenarios:
  - name: missing stdin file errors
    steps:
      - run: {shell: true, command: "cat", stdin: {file: "nope.txt"}}
```
#### When
```shell
${atago} run --ci --report junit errored.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `errors="1"`, `failures="0"`, `<error`
## atago self-hosting / reports
Source: `test/e2e/atago/reports.atago.yaml`
### Scenario: JUnit report is XML with a testsuite and testcase
#### Given
- Fixture file `ok.atago.yaml` is created.
#### Inputs
_Fixture `ok.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: echo
    steps:
      - run:
          command: echo hi
      - assert:
          stdout:
            contains: hi
```
#### When
```shell
${atago} run --report junit ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `<testsuites`, `<testcase`
### Scenario: GitHub Actions annotations are emitted on failure
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: nope
    steps:
      - run:
          command: echo Bob
      - assert:
          stdout:
            contains: Alice
```
#### When
```shell
${atago} run --report gha bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `::error`
### Scenario: TAP report is a numbered TAP 13 stream with ok / not ok points
#### Given
- Fixture file `mixed.atago.yaml` is created.
#### Inputs
_Fixture `mixed.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: good
    steps:
      - run:
          command: echo hi
      - assert:
          stdout:
            contains: hi
  - name: bad
    steps:
      - run:
          command: echo Bob
      - assert:
          stdout:
            contains: Alice
```
#### When
```shell
${atago} run --report tap mixed.atago.yaml
```
#### Then
- exit code is `1`
- stdout equals an exact value
- stdout equals an exact value
- stdout contains `ok 1 - sample / good`, `not ok 2 - sample / bad`
### Scenario: failure artifacts are written and referenced in the JSON report
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: nope
    steps:
      - run:
          command: echo Bob
      - assert:
          stdout:
            contains: Alice
```
#### When
```shell
${atago} run --report json --artifacts-dir arts bad.atago.yaml
cat arts/*/*/step-*-stdout.actual.txt
```
#### Then
- after `${atago} run --report json --artifacts-dir arts bad.atago.yaml`:
  - exit code is `1`
  - stdout contains `"artifacts"`, `.actual.txt`
- after `cat arts/*/*/step-*-stdout.actual.txt`:
  - exit code is `0`
  - stdout contains `Bob`
### Scenario: a multi-line snapshot failure renders a unified diff with hunks
#### Given
- Fixture file `diffspec.atago.yaml` is created.
- Fixture file `diffspec.atago.yaml` is created.
#### Inputs
_Fixture `diffspec.atago.yaml`:_
```text
version: "1"
suite:
  name: diffdemo
scenarios:
  - name: multi-line output matches its golden
    steps:
      - run:
          shell: true
          command: "printf 'alpha\\nbeta\\ngamma\\n'"
      - assert:
          stdout:
            snapshot: snaps/out.txt
```
_Fixture `diffspec.atago.yaml`:_
```text
version: "1"
suite:
  name: diffdemo
scenarios:
  - name: multi-line output matches its golden
    steps:
      - run:
          shell: true
          command: "printf 'alpha\\nBETA\\ngamma\\n'"
      - assert:
          stdout:
            snapshot: snaps/out.txt
```
#### When
```shell
${atago} run --update-snapshots diffspec.atago.yaml
${atago} run diffspec.atago.yaml
${atago} run --report json diffspec.atago.yaml; true
```
#### Then
- after `${atago} run --update-snapshots diffspec.atago.yaml`:
  - exit code is `0`
- after `${atago} run diffspec.atago.yaml`:
  - exit code is `1`
  - stdout contains `Diff (-expected +actual):`, `--- snapshot (golden)`, `-beta`, `+BETA`, `snaps/out.txt`
- after `${atago} run --report json diffspec.atago.yaml; true`:
  - stdout contains `"diff":`
## atago self-hosting / rerun-failed
Source: `test/e2e/atago/rerun.atago.yaml`
### Scenario: a failing run is recorded and rerun-failed selects only it
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: always-green
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
  - name: always-red
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
```
#### When
```shell
${atago} run inner.atago.yaml
${atago} run --rerun-failed inner.atago.yaml
```
#### Then
- after `${atago} run inner.atago.yaml`:
  - exit code is `1`
  - file `.atago/last-failed.json` exists
  - file `.atago/last-failed.json` at `$.failed[0].scenario` equals `always-red`
- after `${atago} run --rerun-failed inner.atago.yaml`:
  - exit code is `1`
  - stdout contains `always-red`
#### Generated artifacts
- `.atago/last-failed.json`
### Scenario: rerun-failed with nothing recorded is a no-op success
#### Given
- Fixture file `green.atago.yaml` is created.
#### Inputs
_Fixture `green.atago.yaml`:_
```text
version: "1"
suite:
  name: green
scenarios:
  - name: passes
    steps:
      - run: {shell: true, command: "exit 0"}
      - assert: {exit_code: 0}
```
#### When
```shell
${atago} run --rerun-failed green.atago.yaml
```
#### Then
- exit code is `0`
- stderr contains `nothing to rerun`
### Scenario: rerun-failed with a filter preserves the still-failing scenarios it did not run
#### Given
- Fixture file `two.atago.yaml` is created.
#### Inputs
_Fixture `two.atago.yaml`:_
```text
version: "1"
suite:
  name: two
scenarios:
  - name: red-a
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
  - name: red-b
    steps:
      - run: {shell: true, command: "exit 1"}
      - assert: {exit_code: 0}
```
#### When
```shell
${atago} run two.atago.yaml
${atago} run --rerun-failed --filter red-a two.atago.yaml
```
#### Then
- after `${atago} run two.atago.yaml`:
  - exit code is `1`
- after `${atago} run --rerun-failed --filter red-a two.atago.yaml`:
  - exit code is `1`
  - file `.atago/last-failed.json` contains `red-a`, `red-b`
## atago self-hosting / retry until
Source: `test/e2e/atago/retry.atago.yaml`
### Scenario: retry polls until the condition becomes true
#### Given
- Fixture file `ready.atago.yaml` is created.
#### Inputs
_Fixture `ready.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: becomes ready on a later attempt
    steps:
      - run:
          shell: true
          command: "if [ -f marker ]; then echo ready; else touch marker; echo waiting; fi"
          retry:
            times: 5
            interval: 10ms
            until:
              stdout:
                contains: ready
      - assert:
          stdout:
            contains: ready
```
#### When
```shell
${atago} run ready.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `passed`
### Scenario: retry fails the inner spec when until never holds
#### Given
- Fixture file `never.atago.yaml` is created.
#### Inputs
_Fixture `never.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: never ready
    steps:
      - run:
          command: echo waiting
          retry:
            times: 3
            interval: 1ms
            until:
              stdout:
                contains: ready
```
#### When
```shell
${atago} run never.atago.yaml
```
#### Then
- exit code is `1`
### Scenario: until with a changes target is a load-time error
#### Given
- Fixture file `badchanges.atago.yaml` is created.
#### Inputs
_Fixture `badchanges.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: unsatisfiable until
    steps:
      - run:
          command: echo hi
          retry:
            times: 3
            until:
              changes:
                created: [out.txt]
```
#### When
```shell
${atago} run badchanges.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `retry.until.changes cannot be satisfied`, `steps[0].run`
## atago self-hosting / run
Source: `test/e2e/atago/run.atago.yaml`
### Scenario: a passing spec exits zero and reports PASS
#### Given
- Fixture file `passing.atago.yaml` is created.
#### Inputs
_Fixture `passing.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: exit 0 succeeds
    steps:
      - run:
          shell: true
          command: "exit 0"
      - assert:
          exit_code: 0
          stderr:
            empty: true
```
#### When
```shell
${atago} run passing.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASS`
- stderr is empty
### Scenario: a failing assertion exits one and reports the failure
#### Given
- Fixture file `failing.atago.yaml` is created.
#### Inputs
_Fixture `failing.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: exit 1 should pass
    steps:
      - run:
          shell: true
          command: "exit 1"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run failing.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `FAILED`, `expected exit code 0`, `(failing.atago.yaml)`
### Scenario: an exit_code failure surfaces the command's stderr
_skipped on Windows_
#### Given
- Fixture file `stderr_cause.atago.yaml` is created.
#### Inputs
_Fixture `stderr_cause.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: converter fails loudly
    steps:
      - run:
          shell: true
          command: "echo conversion aborted: bad header 1>&2; exit 3"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run stderr_cause.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `Stderr:`, `conversion aborted: bad header`
### Scenario: an exit_code failure surfaces the command's stderr (windows)
_only on Windows_
#### Given
- Fixture file `stderr_cause.atago.yaml` is created.
#### Inputs
_Fixture `stderr_cause.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: converter fails loudly
    steps:
      - run:
          shell: true
          command: "echo conversion aborted: bad header 1>&2 & exit 3"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run stderr_cause.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `Stderr:`, `conversion aborted: bad header`
### Scenario: an exit_code failure falls back to stdout when stderr is silent
_skipped on Windows_
#### Given
- Fixture file `stdout_cause.atago.yaml` is created.
#### Inputs
_Fixture `stdout_cause.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: converter fails on stdout
    steps:
      - run:
          shell: true
          command: "echo wrote 0 of 3 files; exit 3"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run stdout_cause.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `Stdout:`, `wrote 0 of 3 files`
### Scenario: an exit_code failure falls back to stdout when stderr is silent (windows)
_only on Windows_
#### Given
- Fixture file `stdout_cause.atago.yaml` is created.
#### Inputs
_Fixture `stdout_cause.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: converter fails on stdout
    steps:
      - run:
          shell: true
          command: "echo wrote 0 of 3 files & exit 3"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run stdout_cause.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `Stdout:`, `wrote 0 of 3 files`
### Scenario: a parse error exits with code two
#### Given
- Fixture file `broken.atago.yaml` is created.
#### Inputs
_Fixture `broken.atago.yaml`:_
```text
version: "1"
suite:
  : not valid yaml
```
#### When
```shell
${atago} run broken.atago.yaml
```
#### Then
- exit code is `2`
### Scenario: JSON report is valid JSON with a passed status
#### Given
- Fixture file `ok.atago.yaml` is created.
#### Inputs
_Fixture `ok.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: echo
    steps:
      - run:
          shell: true
          command: echo hello
      - assert:
          stdout:
            contains: hello
```
#### When
```shell
${atago} run --report json ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.schema_version` equals `1`
- stdout at `$.suites[0].status` equals `passed`
- stdout at `$.suites[0].scenarios` has length 1
## atago self-hosting / sandbox_home (isolated per-OS home)
Source: `test/e2e/atago/sandbox_home.atago.yaml`
### Scenario: Unix XDG family — write config, read it back, inspect it under the workdir
_skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
mkdir -p "$XDG_CONFIG_HOME/mytool" && printf editor=vim > "$XDG_CONFIG_HOME/mytool/config"
cat "$XDG_CONFIG_HOME/mytool/config"
```
#### Then
- after `mkdir -p "$XDG_CONFIG_HOME/mytool" && printf editor=vim > "$XDG_CONFIG_HOME/mytool/config"`:
  - exit code is `0`
- after `cat "$XDG_CONFIG_HOME/mytool/config"`:
  - exit code is `0`
  - stdout equals an exact value
  - file `.atago-home/.config/mytool/config` contains `editor=vim`
### Scenario: Windows APPDATA family — write config, read it back, inspect it under the workdir
_only on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
mkdir "%APPDATA%\mytool" & echo editor=vim>"%APPDATA%\mytool\config.txt"
type "%APPDATA%\mytool\config.txt"
```
#### Then
- after `mkdir "%APPDATA%\mytool" & echo editor=vim>"%APPDATA%\mytool\config.txt"`:
  - exit code is `0`
- after `type "%APPDATA%\mytool\config.txt"`:
  - exit code is `0`
  - stdout contains `editor=vim`
  - file `.atago-home/AppData/Roaming/mytool/config.txt` contains `editor=vim`
### Scenario: cwd anchors the run, but sandbox_home stays at the workdir ROOT (Unix)
_skipped on Windows_
#### Given
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### When
```shell
mkdir -p sub
mkdir -p "$XDG_CONFIG_HOME/mytool" && printf editor=vim > "$XDG_CONFIG_HOME/mytool/config"
```
#### Then
- after `mkdir -p "$XDG_CONFIG_HOME/mytool" && printf editor=vim > "$XDG_CONFIG_HOME/mytool/config"`:
  - exit code is `0`
  - file `.atago-home/.config/mytool/config` contains `editor=vim`
  - file `sub/.atago-home/.config/mytool/config` does not exist
## atago self-hosting / security
Source: `test/e2e/atago/security.atago.yaml`
### Scenario: declared secrets are masked in failure output
#### Given
- Fixture file `sec.atago.yaml` is created.
- Environment variables are set: DEMO_TOKEN.
#### Inputs
_Fixture `sec.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
secrets:
  - DEMO_TOKEN
scenarios:
  - name: leaky
    steps:
      - run:
          shell: true
          command: echo "token=$DEMO_TOKEN"
      - assert:
          stdout:
            contains: NOPE
```
#### When
```shell
${atago} run sec.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `token=***`
### Scenario: a file assertion path may not escape the scenario workdir
#### Given
- Fixture file `escape.atago.yaml` is created.
#### Inputs
_Fixture `escape.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: traversal
    steps:
      - run:
          shell: true
          command: echo hi
      - assert:
          file:
            path: ../../../etc/hosts
            exists: true
```
#### When
```shell
${atago} run escape.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `escapes the scenario workdir`
### Scenario: a snapshot path may not escape the spec directory
#### Given
- Fixture file `snap_escape.atago.yaml` is created.
#### Inputs
_Fixture `snap_escape.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: traversal
    steps:
      - run:
          shell: true
          command: echo hi
      - assert:
          stdout:
            snapshot: ../../../tmp/leak.snap
```
#### When
```shell
${atago} run snap_escape.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `escapes the spec directory`
## atago self-hosting / selection
Source: `test/e2e/atago/select.atago.yaml`
### Scenario: --filter runs only matching scenarios
#### Given
- Fixture file `many.atago.yaml` is created.
#### Inputs
_Fixture `many.atago.yaml`:_
```text
version: "1"
suite:
  name: many
scenarios:
  - name: keep this one
    steps: [{run: {shell: true, command: echo a}}, {assert: {exit_code: 0}}]
  - name: drop that one
    steps: [{run: {shell: true, command: echo b}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --filter keep many.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: --filter selects multiple scenarios with OR (comma and repeated)
#### Given
- Fixture file `three.atago.yaml` is created.
#### Inputs
_Fixture `three.atago.yaml`:_
```text
version: "1"
suite:
  name: three
scenarios:
  - name: alpha one
    steps: [{run: {shell: true, command: echo a}}, {assert: {exit_code: 0}}]
  - name: beta two
    steps: [{run: {shell: true, command: echo b}}, {assert: {exit_code: 0}}]
  - name: gamma three
    steps: [{run: {shell: true, command: echo c}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --filter alpha,beta three.atago.yaml
${atago} run --filter alpha --filter gamma three.atago.yaml
```
#### Then
- after `${atago} run --filter alpha,beta three.atago.yaml`:
  - exit code is `0`
  - stdout contains `2 passed`
- after `${atago} run --filter alpha --filter gamma three.atago.yaml`:
  - exit code is `0`
  - stdout contains `2 passed`
### Scenario: --skip-tag drops tagged scenarios
#### Given
- Fixture file `tagged.atago.yaml` is created.
#### Inputs
_Fixture `tagged.atago.yaml`:_
```text
version: "1"
suite:
  name: tagged
scenarios:
  - name: quick
    tags: [fast]
    steps: [{run: {shell: true, command: echo a}}, {assert: {exit_code: 0}}]
  - name: heavy
    tags: [slow]
    steps: [{run: {shell: true, command: echo b}}, {assert: {exit_code: 0}}]
```
#### When
```shell
${atago} run --skip-tag slow tagged.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
## atago self-hosting / background services
Source: `test/e2e/atago/services.atago.yaml`
### Scenario: file readiness captures a dynamic value into a variable
#### Given
- Background service `publisher` is started: `printf "127.0.0.1:5555" > ready.txt; sleep 30`.
#### When
```shell
echo ${addr}
```
#### Then
- stdout equals an exact value
### Scenario: log readiness waits for a line on the service output
#### Given
- Background service `logger` is started: `echo "ready: listening"; sleep 30`.
#### When
```shell
echo started
```
#### Then
- stdout contains `started`
### Scenario: delay readiness waits a fixed duration
#### Given
- Background service `slow` is started: `sleep 30`.
#### When
```shell
echo ok
```
#### Then
- stdout contains `ok`
### Scenario: multiple services start and capture independently
#### Given
- Background service `first` is started: `printf alpha > a.txt; sleep 30`.
- Background service `second` is started: `printf beta > b.txt; sleep 30`.
#### When
```shell
echo ${a}-${b}
```
#### Then
- stdout equals an exact value
### Scenario: a readiness failure preserves the service log as an artifact
#### Given
- Fixture file `notready.atago.yaml` is created.
#### Inputs
_Fixture `notready.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: service never ready
    services:
      - name: chatty
        shell: true
        command: 'echo booting-up; sleep 30'
        ready:
          file: never.txt
          timeout: 200ms
    steps:
      - run:
          command: echo unreached
```
#### When
```shell
${atago} run --artifacts-dir arts notready.atago.yaml
cat arts/*/*/service-chatty.log
```
#### Then
- after `${atago} run --artifacts-dir arts notready.atago.yaml`:
  - exit code is `4`
- after `cat arts/*/*/service-chatty.log`:
  - exit code is `0`
  - stdout contains `booting-up`
### Scenario: a step failure after the service is ready preserves the service log
#### Given
- Fixture file `readythenfail.atago.yaml` is created.
#### Inputs
_Fixture `readythenfail.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: ready then fail
    services:
      - name: peer
        shell: true
        command: 'echo peer-serving > ready.txt; echo peer-log-line; sleep 30'
        ready:
          file: ready.txt
          timeout: 2s
    steps:
      - run:
          command: echo hello
      - assert:
          stdout:
            contains: goodbye
```
#### When
```shell
${atago} run --report json --artifacts-dir arts readythenfail.atago.yaml
cat arts/*/*/service-peer.log
```
#### Then
- after `${atago} run --report json --artifacts-dir arts readythenfail.atago.yaml`:
  - exit code is `1`
  - stdout contains `service_logs`, `service-peer.log`
- after `cat arts/*/*/service-peer.log`:
  - exit code is `0`
  - stdout contains `peer-log-line`
### Scenario: a green run with a healthy service writes no service log
#### Given
- Fixture file `healthy.atago.yaml` is created.
#### Inputs
_Fixture `healthy.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: healthy service passes
    services:
      - name: ok
        shell: true
        command: 'echo up > ready.txt; echo serving; sleep 30'
        ready:
          file: ready.txt
          timeout: 2s
    steps:
      - run:
          command: echo hi
      - assert:
          stdout:
            contains: hi
```
#### When
```shell
${atago} run --artifacts-dir arts healthy.atago.yaml
ls arts 2>/dev/null | wc -l | tr -d " "
```
#### Then
- after `${atago} run --artifacts-dir arts healthy.atago.yaml`:
  - exit code is `0`
- after `ls arts 2>/dev/null | wc -l | tr -d " "`:
  - stdout equals an exact value
## atago self-hosting / harness shell is not shadowed by the program PATH
Source: `test/e2e/atago/shell_path.atago.yaml`
### Scenario: a PATH-resident fake sh does not hijack shell:true
#### Given
- Fixture file `sh` is created.
#### Inputs
_Fixture `sh`:_
```text
#!/bin/sh
echo HIJACKED
exit 0
```
#### When
```shell
echo real-shell
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout does not contain `HIJACKED`
### Scenario: ATAGO_SHELL overrides the shell used for shell:true
#### When
```shell
printf '%s\n' ok
```
#### Then
- stdout equals an exact value
## atago self-hosting / signal step (graceful shutdown)
Source: `test/e2e/atago/signal.atago.yaml`
### Scenario: SIGTERM reaches the trap handler and wait observes the exit
_skipped on Windows_
#### Given
- Background service `server` is started: `trap 'echo graceful shutdown complete > server.log; exit 0' TERM; echo booted; while true; do sleep 0.1; done`.
#### When
```shell
# send SIGTERM to service server and wait up to 5s for exit
```
#### Then
- file `server.log` contains `graceful shutdown complete`
### Scenario: SIGHUP triggers a reload without stopping the service
_skipped on Windows_
#### Given
- Background service `reloader` is started: `trap 'echo reloaded >> reload.log' HUP; echo booted; while true; do sleep 0.1; done`.
#### When
```shell
# send SIGHUP to service reloader
for i in 1 2 3 4 5 6 7 8 9 10; do [ -f reload.log ] && break; sleep 0.1; done; cat reload.log
```
#### Then
- after `for i in 1 2 3 4 5 6 7 8 9 10; do [ -f reload.log ] && break; sleep 0.1; done; cat reload.log`:
  - exit code is `0`
  - stdout contains `reloaded`
### Scenario: a wait timeout on a TERM-ignoring service fails with the documented message
_skipped on Windows_
#### Given
- Fixture file `stubborn.atago.yaml` is created.
#### Inputs
_Fixture `stubborn.atago.yaml`:_
```text
version: "1"
suite:
  name: stubborn
scenarios:
  - name: never exits on TERM
    services:
      - name: stubborn
        shell: true
        command: "trap '' TERM; echo booted; while true; do sleep 0.2; done"
        ready: {log: booted, timeout: 10s}
    steps:
      - signal:
          service: stubborn
          signal: TERM
          wait:
            timeout: 300ms
```
#### When
```shell
${atago} run stubborn.atago.yaml
```
#### Then
- exit code is not `0`
- stdout contains `did not exit within 300ms after SIGTERM`
### Scenario: an unknown target service is a load-time error listing declared names
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: wrong name
    services:
      - name: web
        command: ./web
    steps:
      - signal:
          service: cache
          signal: TERM
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `not a declared service (declared: web)`
## atago self-hosting / skip-only command predicate
Source: `test/e2e/atago/skip_command.atago.yaml`
### Scenario: skip command that succeeds skips the scenario
#### Given
- Fixture file `skip.atago.yaml` is created.
#### Inputs
_Fixture `skip.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: gated by a succeeding probe
    skip:
      command: "exit 0"
    steps:
      - run:
          shell: true
          command: echo should-not-run
      - assert:
          stdout:
            contains: never-checked
```
#### When
```shell
${atago} run skip.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `skipped`
### Scenario: only command that fails skips the scenario
#### Given
- Fixture file `only.atago.yaml` is created.
#### Inputs
_Fixture `only.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: gated by a failing probe
    only:
      command: "exit 1"
    steps:
      - run:
          shell: true
          command: echo should-not-run
      - assert:
          stdout:
            contains: never-checked
```
#### When
```shell
${atago} run only.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `skipped`
### Scenario: only command that succeeds runs the scenario
#### Given
- Fixture file `run.atago.yaml` is created.
#### Inputs
_Fixture `run.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: gated by a succeeding probe
    only:
      command: "exit 0"
    steps:
      - run:
          shell: true
          command: echo ran
      - assert:
          stdout:
            contains: ran
```
#### When
```shell
${atago} run run.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `passed`
## atago self-hosting / snapshot
Source: `test/e2e/atago/snapshot.atago.yaml`
### Scenario: a snapshot assertion passes against a committed snapshot
#### Given
- Fixture file `snap.atago.yaml` is created.
- Fixture file `out.snap` is created.
#### Inputs
_Fixture `snap.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: stable output
    steps:
      - run:
          command: echo stable
      - assert:
          stdout:
            snapshot: out.snap
```
_Fixture `out.snap`:_
```text
stable
```
#### When
```shell
${atago} run snap.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASSED`
### Scenario: snapshot update creates the snapshot file
#### Given
- Fixture file `gen.atago.yaml` is created.
#### Inputs
_Fixture `gen.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: stable output
    steps:
      - run:
          command: echo stable
      - assert:
          stdout:
            snapshot: created.snap
```
#### When
```shell
${atago} snapshot update gen.atago.yaml
```
#### Then
- exit code is `0`
- file `created.snap` contains `stable`
### Scenario: a snapshot mismatch writes the normalized actual as an artifact
#### Given
- Fixture file `drift.atago.yaml` is created.
- Fixture file `committed.snap` is created.
#### Inputs
_Fixture `drift.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: drifted output
    steps:
      - run:
          command: echo changed
      - assert:
          stdout:
            snapshot: committed.snap
```
_Fixture `committed.snap`:_
```text
original
```
#### When
```shell
${atago} run --artifacts-dir arts drift.atago.yaml
cat arts/*/*/step-*-snapshot.actual.txt
```
#### Then
- after `${atago} run --artifacts-dir arts drift.atago.yaml`:
  - exit code is `1`
  - stdout contains `Artifacts:`
- after `cat arts/*/*/step-*-snapshot.actual.txt`:
  - exit code is `0`
  - stdout contains `changed`
## atago self-hosting / snapshot normalization and round-trip
Source: `test/e2e/atago/snapshot_normalization.atago.yaml`
### Scenario: record then run round-trips green
#### Given
- Fixture file `rt.atago.yaml` is created.
#### Inputs
_Fixture `rt.atago.yaml`:_
```text
version: "1"
suite: {name: rt}
scenarios:
  - name: stable line
    steps:
      - run: {shell: true, command: "printf 'stable output line\\n'"}
      - assert: {stdout: {snapshot: g.txt}}
```
#### When
```shell
${atago} snapshot update rt.atago.yaml
${atago} run rt.atago.yaml
```
#### Then
- after `${atago} snapshot update rt.atago.yaml`:
  - exit code is `0`
- after `${atago} run rt.atago.yaml`:
  - exit code is `0`
### Scenario: a UUID is masked in the golden
#### Given
- Fixture file `uuid.atago.yaml` is created.
#### Inputs
_Fixture `uuid.atago.yaml`:_
```text
version: "1"
suite: {name: uuid}
scenarios:
  - name: emits a uuid
    steps:
      - run: {shell: true, command: "printf 'id=550e8400-e29b-41d4-a716-446655440000\\n'"}
      - assert: {stdout: {snapshot: g.txt}}
```
#### When
```shell
${atago} snapshot update uuid.atago.yaml
```
#### Then
- file `g.txt` contains `id=<uuid>`
- file `g.txt` is checked
### Scenario: an ISO timestamp is masked in the golden
#### Given
- Fixture file `ts.atago.yaml` is created.
#### Inputs
_Fixture `ts.atago.yaml`:_
```text
version: "1"
suite: {name: ts}
scenarios:
  - name: emits a timestamp
    steps:
      - run: {shell: true, command: "printf 'at 2024-01-15T10:30:00Z done\\n'"}
      - assert: {stdout: {snapshot: g.txt}}
```
#### When
```shell
${atago} snapshot update ts.atago.yaml
```
#### Then
- file `g.txt` contains `at <timestamp> done`
### Scenario: a loopback host and port are masked in the golden
#### Given
- Fixture file `port.atago.yaml` is created.
#### Inputs
_Fixture `port.atago.yaml`:_
```text
version: "1"
suite: {name: port}
scenarios:
  - name: emits a listen address
    steps:
      - run: {shell: true, command: "printf 'listening on 127.0.0.1:54321\\n'"}
      - assert: {stdout: {snapshot: g.txt}}
```
#### When
```shell
${atago} snapshot update port.atago.yaml
```
#### Then
- file `g.txt` contains `127.0.0.1:<port>`
- file `g.txt` is checked
### Scenario: the home directory is masked to a tilde in the golden
#### Given
- Fixture file `home.atago.yaml` is created.
#### Inputs
_Fixture `home.atago.yaml`:_
```text
version: "1"
suite: {name: home}
scenarios:
  - name: emits the home path
    steps:
      - run: {shell: true, command: "printf 'home=%s/x\\n' \"$HOME\""}
      - assert: {stdout: {snapshot: g.txt}}
```
#### When
```shell
${atago} snapshot update home.atago.yaml
```
#### Then
- file `g.txt` contains `home=~/x`
### Scenario: a golden verifies against a different volatile value
#### Given
- Fixture file `rec.atago.yaml` is created.
- Fixture file `ver.atago.yaml` is created.
#### Inputs
_Fixture `rec.atago.yaml`:_
```text
version: "1"
suite: {name: rec}
scenarios:
  - name: uuid A
    steps:
      - run: {shell: true, command: "printf 'token=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee\\n'"}
      - assert: {stdout: {snapshot: shared.txt}}
```
_Fixture `ver.atago.yaml`:_
```text
version: "1"
suite: {name: rec}
scenarios:
  - name: uuid A
    steps:
      - run: {shell: true, command: "printf 'token=11111111-2222-3333-4444-555555555555\\n'"}
      - assert: {stdout: {snapshot: shared.txt}}
```
#### When
```shell
${atago} snapshot update rec.atago.yaml
${atago} run ver.atago.yaml
```
#### Then
- after `${atago} snapshot update rec.atago.yaml`:
  - exit code is `0`
- after `${atago} run ver.atago.yaml`:
  - exit code is `0`
### Scenario: updating a snapshot is deterministic
#### Given
- Fixture file `d1.atago.yaml` is created.
- Fixture file `d2.atago.yaml` is created.
#### Inputs
_Fixture `d1.atago.yaml`:_
```text
version: "1"
suite: {name: d}
scenarios:
  - name: same command
    steps:
      - run: {shell: true, command: "printf 'line one\\nline two\\n'"}
      - assert: {stdout: {snapshot: a.txt}}
```
_Fixture `d2.atago.yaml`:_
```text
version: "1"
suite: {name: d}
scenarios:
  - name: same command
    steps:
      - run: {shell: true, command: "printf 'line one\\nline two\\n'"}
      - assert: {stdout: {snapshot: b.txt}}
```
#### When
```shell
${atago} snapshot update d1.atago.yaml
${atago} snapshot update d2.atago.yaml
```
#### Then
- after `${atago} snapshot update d2.atago.yaml`:
  - file `a.txt` is byte-identical to `b.txt`
### Scenario: a real content change still fails the snapshot
#### Given
- Fixture file `change.atago.yaml` is created.
- Fixture file `g.txt` is created.
#### Inputs
_Fixture `change.atago.yaml`:_
```text
version: "1"
suite: {name: change}
scenarios:
  - name: drifts from the golden
    steps:
      - run: {shell: true, command: "printf 'the committed line\\n'"}
      - assert: {stdout: {snapshot: g.txt}}
```
_Fixture `g.txt`:_
```text
a different committed line
```
#### When
```shell
${atago} run change.atago.yaml
```
#### Then
- exit code is `1`
### Scenario: a missing golden names the update flag
#### Given
- Fixture file `miss.atago.yaml` is created.
#### Inputs
_Fixture `miss.atago.yaml`:_
```text
version: "1"
suite: {name: miss}
scenarios:
  - name: no golden yet
    steps:
      - run: {shell: true, command: "printf 'hi\\n'"}
      - assert: {stdout: {snapshot: never.txt}}
```
#### When
```shell
${atago} run miss.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `--update-snapshots`
## atago self-hosting / ssh runner
Source: `test/e2e/atago/ssh.atago.yaml`
### Scenario: an ssh runner without host/user fails validation (exit 2)
#### Given
- Fixture file `badssh.atago.yaml` is created.
#### Inputs
_Fixture `badssh.atago.yaml`:_
```text
version: "1"
suite:
  name: ssh
runners:
  box:
    type: ssh
scenarios:
  - name: remote run
    steps:
      - run:
          runner: box
          command: uptime
```
#### When
```shell
${atago} run badssh.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `requires a host`
### Scenario: a run step naming an undeclared runner fails validation (exit 2)
#### Given
- Fixture file `norunner.atago.yaml` is created.
#### Inputs
_Fixture `norunner.atago.yaml`:_
```text
version: "1"
suite:
  name: ssh
scenarios:
  - name: remote run
    steps:
      - run:
          runner: missing
          command: uptime
```
#### When
```shell
${atago} run norunner.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `is not declared`
### Scenario: a local-only run field on an ssh runner fails validation (exit 2)
#### Given
- Fixture file `sshfield.atago.yaml` is created.
#### Inputs
_Fixture `sshfield.atago.yaml`:_
```text
version: "1"
suite:
  name: ssh
runners:
  box:
    type: ssh
    host: deploy.example.com
    user: deploy
scenarios:
  - name: remote run
    steps:
      - run:
          runner: box
          command: uptime
          cwd: /var/tmp
```
#### When
```shell
${atago} run sshfield.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `run.cwd has no effect on an ssh runner`, `steps[0].run`
## atago self-hosting / stdin sources (file + base64)
Source: `test/e2e/atago/stdin_sources.atago.yaml`
### Scenario: base64 stdin delivers the exact byte count
_skipped on Windows_
#### Inputs
_stdin for `wc`:_
```text
(binary, 8 base64 chars)
```
#### When
```shell
wc -c
```
#### Then
- exit code is `0`
- stdout matches `/^\s*4\s*$/`
### Scenario: stdin file is expanded and read from the workdir
_skipped on Windows_
#### Given
- Fixture file `payload.txt` is created.
#### Inputs
_Fixture `payload.txt`:_
```text
from-a-file
```
_stdin for `cat`:_
```text
(read from file ${workdir}/payload.txt)
```
#### When
```shell
cat
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: a stdin file outside the workdir is rejected at runtime
_skipped on Windows_
#### Given
- Fixture file `escape.atago.yaml` is created.
#### Inputs
_Fixture `escape.atago.yaml`:_
```text
version: "1"
suite:
  name: escape
scenarios:
  - name: tries to read outside the workdir
    steps:
      - run:
          command: cat
          stdin:
            file: ../outside.txt
```
#### When
```shell
${atago} run escape.atago.yaml
```
#### Then
- exit code is not `0`
- stdout contains `run.stdin.file`
### Scenario: stdin with both file and base64 is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
scenarios:
  - name: both sources
    steps:
      - run:
          command: cat
          stdin:
            file: in.txt
            base64: AAEC
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `exactly one of file/base64`
### Scenario: invalid base64 stdin is a load-time error
#### Given
- Fixture file `badb64.atago.yaml` is created.
#### Inputs
_Fixture `badb64.atago.yaml`:_
```text
version: "1"
suite:
  name: badb64
scenarios:
  - name: bad payload
    steps:
      - run:
          command: cat
          stdin:
            base64: "not!!base64"
```
#### When
```shell
${atago} run badb64.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `not valid base64`
## atago self-hosting / store
Source: `test/e2e/atago/store.atago.yaml`
### Scenario: a stored JSON value is reusable in later commands
#### Given
- Fixture file `store.atago.yaml` is created.
#### Inputs
_Fixture `store.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: capture id and greet
    steps:
      - run:
          shell: true
          command: printf '{"id":99,"name":"Bob"}'
      - store:
          name: user_id
          from:
            stdout:
              json:
                path: "$.id"
      - run:
          shell: true
          command: echo "user is ${user_id}"
      - assert:
          stdout:
… (truncated, 1 more line)
```
#### When
```shell
${atago} run store.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASSED`
### Scenario: storing from a missing JSON path is an execution error
#### Given
- Fixture file `bad-store.atago.yaml` is created.
#### Inputs
_Fixture `bad-store.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: capture nonexistent
    steps:
      - run:
          shell: true
          command: printf '{}'
      - store:
          name: missing
          from:
            stdout:
              json:
                path: "$.nope"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run bad-store.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `ERROR`
## atago self-hosting / store capture boundary values
Source: `test/e2e/atago/store_edges.atago.yaml`
### Scenario: a regex with a capture group stores the group
#### When
```shell
echo release v1.2.3-rc1
# capture ${ver} from stdout
echo shipping ${ver}
```
#### Then
- after `echo shipping ${ver}`:
  - stdout equals an exact value
### Scenario: a regex without a group stores the whole match
#### When
```shell
echo digest sha=abc123def456
# capture ${sha} from stdout
echo checksum ${sha}
```
#### Then
- after `echo checksum ${sha}`:
  - stdout equals an exact value
### Scenario: a JSON path captures a scalar from stdout
#### When
```shell
printf '{"port":8080}'
# capture ${port} from stdout
echo bound ${port}
```
#### Then
- after `echo bound ${port}`:
  - stdout equals an exact value
### Scenario: a JSON path captures a value from a generated file
#### Given
- Fixture file `meta.json` is created.
#### Inputs
_Fixture `meta.json`:_
```text
{"build":{"id":"b-42"}}
```
#### When
```shell
# capture ${build} from file meta.json
echo build ${build}
```
#### Then
- stdout equals an exact value
### Scenario: a regex that matches nothing is an execution error
#### Given
- Fixture file `nomatch.atago.yaml` is created.
#### Inputs
_Fixture `nomatch.atago.yaml`:_
```text
version: "1"
suite: {name: nomatch}
scenarios:
  - name: capture fails
    steps:
      - run: {shell: true, command: "echo hello"}
      - store: {name: x, from: {stdout: {matches: "ABSENT([0-9]+)"}}}
```
#### When
```shell
${atago} run nomatch.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `did not match`
### Scenario: a stored value does not leak into the next scenario
#### Given
- Fixture file `scope.atago.yaml` is created.
#### Inputs
_Fixture `scope.atago.yaml`:_
```text
version: "1"
suite: {name: scope}
scenarios:
  - name: captures a value
    steps:
      - run: {shell: true, command: "echo secret-value"}
      - store: {name: captured, from: {stdout: {trim: true}}}
      - assert: {stdout: {contains: secret-value}}
  - name: cannot reference the earlier value
    steps:
      - run: {command: "echo ${captured}"}
```
#### When
```shell
${atago} run scope.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `references ${captured}`, `no variable with that name is defined`
## atago self-hosting / store whole-content trim and text selectors (#158)
Source: `test/e2e/atago/store_whole.atago.yaml`
### Scenario: trim captures an opaque token and round-trips it as an argument
#### When
```shell
echo opaque-token-abc123
# capture ${token} from stdout
${atago} run ${token}
```
#### Then
- after `${atago} run ${token}`:
  - exit code is `3`
  - stderr contains `opaque-token-abc123`
### Scenario: text captures a whole multi-line file verbatim
#### Given
- Fixture file `blob.txt` is created.
- Fixture file `copy.txt` is created.
#### Inputs
_Fixture `blob.txt`:_
```text
first line
second line
```
_Fixture `copy.txt`:_
```text
${blob}
```
#### When
```shell
# capture ${blob} from file blob.txt
```
#### Then
- file `copy.txt` contains `first line`, `second line`
## atago self-hosting / stream matcher boundary values
Source: `test/e2e/atago/stream_edges.atago.yaml`
### Scenario: equals a multibyte and emoji line
#### When
```shell
printf 'テスト🎌\n'
```
#### Then
- stdout equals an exact value
### Scenario: contains a multibyte substring inside a longer line
#### When
```shell
printf 'café ☕ の résumé\n'
```
#### Then
- stdout contains `☕ の`
### Scenario: a regex matches across multibyte runes
#### When
```shell
printf 'αβγδ\n'
```
#### Then
- stdout matches `/β.δ/`
### Scenario: line selection returns a multibyte line intact
#### When
```shell
printf 'ひらがな\nカタカナ\n漢字\n'
```
#### Then
- stdout equals an exact value
### Scenario: not_contains a multibyte needle that is absent
#### When
```shell
printf '日本語\n'
```
#### Then
- stdout does not contain `中文`
### Scenario: empty is true for a command that prints nothing
#### When
```shell
true
```
#### Then
- stdout is empty
### Scenario: empty is true for whitespace-only output
#### When
```shell
printf '   \n\t\n'
```
#### Then
- stdout is empty
### Scenario: equals tolerates output with no trailing newline
#### When
```shell
printf 'noeol'
```
#### Then
- stdout equals an exact value
### Scenario: a deliberate trailing blank line is addressable by index
#### When
```shell
printf 'body\n\n'
```
#### Then
- stdout equals an exact value
### Scenario: contains treats a needle with regex metacharacters literally
#### When
```shell
printf 'price is $3.50 (approx)\n'
```
#### Then
- stdout contains `$3.50 (approx)`
### Scenario: matches requires escaping a literal metacharacter
#### When
```shell
printf 'v1.2.3\n'
```
#### Then
- stdout matches `/v1\.2\.3/`
### Scenario: not_matches passes when an unescaped-metachar pattern does not match
#### When
```shell
printf 'ab\n'
```
#### Then
- stdout does not match `/a.c/`
### Scenario: a tab-separated record contains the exact tab byte
#### When
```shell
printf 'name\tvalue\n'
```
#### Then
- stdout contains `name	value`
### Scenario: quotes and brackets survive an exact equals
#### When
```shell
printf '[{"id":1}]\n'
```
#### Then
- stdout equals an exact value
### Scenario: the last of many lines is selectable by index
#### When
```shell
seq 1 100
```
#### Then
- stdout equals an exact value
### Scenario: a line selector composes with contains
#### When
```shell
printf 'alpha\nbeta-gamma\n'
```
#### Then
- stdout contains `gamma`
### Scenario: a line selector composes with a regex
#### When
```shell
printf 'k=1\nk=2\n'
```
#### Then
- stdout matches `/^k=[0-9]$/`
### Scenario: stderr carries the same matcher semantics as stdout
#### When
```shell
printf 'to stderr\n' 1>&2
```
#### Then
- stdout is empty
- stderr equals an exact value
## atago self-hosting / suite setup
Source: `test/e2e/atago/suite_setup.atago.yaml`
### Scenario: setup runs once, shares stores and env, and teardown always runs
#### Given
- Fixture file `ok.atago.yaml` is created.
#### Inputs
_Fixture `ok.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
  env:
    FLAG: suite-env-flag
  setup:
    - run:
        shell: true
        command: "echo boot-42 > ${suitedir}/t.txt && cat ${suitedir}/t.txt"
    - store:
        name: bootid
        from:
          stdout:
            matches: "boot-[0-9]+"
  teardown:
    - run:
        shell: true
        command: "echo swept ${bootid}"
scenarios:
  - name: one
… (truncated, 12 more lines)
```
#### When
```shell
${atago} run --verbose ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `2 passed`
### Scenario: a failing setup errors every scenario and none runs (exit 4)
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
  setup:
    - run: {command: definitely-not-a-real-binary-xyz}
scenarios:
  - name: never runs
    steps:
      - run: {shell: true, command: echo unreached}
  - name: never runs either
    steps:
      - run: {shell: true, command: echo unreached}
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `suite setup`, `0 passed`, `2 errored`
- stdout does not contain `unreached`
### Scenario: a suite service starts once and its store reaches every scenario
#### Given
- Fixture file `svc.atago.yaml` is created.
#### Inputs
_Fixture `svc.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
  setup:
    - service:
        name: peer
        shell: true
        command: "echo addr-9999 > ${suitedir}/ready.txt && sleep 30"
        ready:
          file: "${suitedir}/ready.txt"
          store: addr
          timeout: 5s
scenarios:
  - name: dials
    steps:
      - run: {shell: true, command: "echo dial ${addr}"}
      - assert:
          stdout: {contains: dial addr-9999}
```
#### When
```shell
${atago} run svc.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`
### Scenario: a failing suite teardown is loud but does not flip the verdict
#### Given
- Fixture file `td.atago.yaml` is created.
#### Inputs
_Fixture `td.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
  setup:
    - run: {shell: true, command: echo fine}
  teardown:
    - run: {command: definitely-not-a-real-binary-xyz}
scenarios:
  - name: passes
    steps:
      - run: {shell: true, command: echo ok}
      - assert:
          stdout: {contains: ok}
```
#### When
```shell
${atago} run td.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `1 passed`, `SUITE TEARDOWN FAILED`
## atago self-hosting / step timeouts (suite default + escape hatch)
Source: `test/e2e/atago/timeouts.atago.yaml`
### Scenario: suite.timeout kills a hanging step and the hint names it
_skipped on Windows_
#### Given
- Fixture file `hang.atago.yaml` is created.
#### Inputs
_Fixture `hang.atago.yaml`:_
```text
version: "1"
suite:
  name: hang
  timeout: 1s
scenarios:
  - name: sleeps past the suite bound
    steps:
      - run:
          shell: true
          command: sleep 300
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run hang.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `timed out`, `suite.timeout`
### Scenario: a step timeout beats the suite timeout and the hint says run.timeout
_skipped on Windows_
#### Given
- Fixture file `step_wins.atago.yaml` is created.
#### Inputs
_Fixture `step_wins.atago.yaml`:_
```text
version: "1"
suite:
  name: step-wins
  timeout: 30s
scenarios:
  - name: the step's own 1s bound fires first
    steps:
      - run:
          shell: true
          command: sleep 300
          timeout: 1s
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run step_wins.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `timed out`, `run.timeout`
### Scenario: timeout zero disables a short suite bound
_skipped on Windows_
#### Given
- Fixture file `optout.atago.yaml` is created.
#### Inputs
_Fixture `optout.atago.yaml`:_
```text
version: "1"
suite:
  name: optout
  timeout: 1s
scenarios:
  - name: the opted-out step outlives the suite bound
    steps:
      - run:
          shell: true
          command: sleep 2
          timeout: "0"
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run optout.atago.yaml
```
#### Then
- exit code is `0`
### Scenario: an invalid suite.timeout is a load-time error
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: bad
  timeout: fast
scenarios:
  - name: never runs
    steps:
      - run:
          command: echo hi
```
#### When
```shell
${atago} run bad.atago.yaml
```
#### Then
- exit code is `2`
- stderr contains `suite.timeout`
## atago self-hosting / tui
Source: `test/e2e/atago/tui.atago.yaml`
### Scenario: a pty step exports a usable TERM by default
_skipped on Windows_
#### When
```shell
# interactive (pty): echo "TERM=[$TERM]"
```
#### Then
- exit code is `0`
- stdout contains `TERM=[xterm-256color]`
### Scenario: an explicit TERM overrides the default
_skipped on Windows_
#### When
```shell
# interactive (pty): echo "TERM=[$TERM]"
```
#### Then
- exit code is `0`
- stdout contains `TERM=[vt100]`
### Scenario: an expect does not re-match a consumed pattern
_skipped on Windows_
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: stale expect must not re-match
    steps:
      - pty:
          shell: true
          command: 'printf "AAA\n"'
          timeout: 1s
          session:
            - expect: "AAA"
            - expect: "AAA"
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `1 failed`
### Scenario: less -X renders a real pager onto the screen
_only when `command -v less` succeeds · skipped on Windows_
#### Given
- Fixture file `page.txt` is created.
#### Inputs
_Fixture `page.txt`:_
```text
Alpha line
Beta line
Gamma line
```
#### When
```shell
# interactive (pty): less -X page.txt
```
#### Then
- rendered screen contains `Alpha line`
- rendered screen contains `Gamma line`
## atago self-hosting / variable resolution semantics
Source: `test/e2e/atago/var_resolution.atago.yaml`
### Scenario: a doubled dollar keeps the braces literal
#### When
```shell
echo pre-$${keep}-post
```
#### Then
- stdout contains `${keep}`
### Scenario: the workdir builtin expands to the scenario directory
#### When
```shell
echo at=${workdir}
```
#### Then
- stdout does not contain `$${workdir}`
### Scenario: the atago builtin resolves to the binary under test
#### When
```shell
${atago} --version
```
#### Then
- exit code is `0`
- stdout contains `atago`
### Scenario: an env reference expands from the host environment
#### When
```shell
echo path=${env:PATH}
```
#### Then
- stdout matches `/path=.+/`
### Scenario: shell true defers an unknown reference to the shell
#### When
```shell
echo [${undefined_in_atago}]
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: an unresolved variable is a hard error, not a silent empty
#### Given
- Fixture file `typo.atago.yaml` is created.
#### Inputs
_Fixture `typo.atago.yaml`:_
```text
version: "1"
suite: {name: typo}
scenarios:
  - name: a misspelled variable stops the run
    steps:
      - run: {command: "echo ${reuslt}"}
```
#### When
```shell
${atago} run typo.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `references ${reuslt}`, `no variable with that name is defined`
### Scenario: an unset env reference names the missing variable
#### Given
- Fixture file `unsetenv.atago.yaml` is created.
#### Inputs
_Fixture `unsetenv.atago.yaml`:_
```text
version: "1"
suite: {name: unsetenv}
scenarios:
  - name: an unset env var stops the run
    steps:
      - run: {command: "echo ${env:ATAGO_SURELY_UNSET_VAR}"}
```
#### When
```shell
${atago} run unsetenv.atago.yaml
```
#### Then
- exit code is `4`
- stdout contains `environment variable ATAGO_SURELY_UNSET_VAR is not set`
## atago self-hosting / verbose
Source: `test/e2e/atago/verbose.atago.yaml`
### Scenario: verbose shows a passing scenario's command, output, and verdicts
#### Given
- Fixture file `ok.atago.yaml` is created.
#### Inputs
_Fixture `ok.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: greets
    steps:
      - run:
          shell: true
          command: echo hello-trace
      - assert:
          exit_code: 0
          stdout:
            contains: hello-trace
```
#### When
```shell
${atago} run --verbose ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `sample / greets`, `echo hello-trace`, `exit 0`, `ok   assert`
### Scenario: without --verbose the trace is absent
#### Given
- Fixture file `ok.atago.yaml` is created.
#### Inputs
_Fixture `ok.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: greets
    steps:
      - run:
          shell: true
          command: echo hello-trace
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout does not contain `echo hello-trace`
### Scenario: verbose with a JSON report keeps stdout pure and traces to stderr
#### Given
- Fixture file `ok.atago.yaml` is created.
#### Inputs
_Fixture `ok.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: greets
    steps:
      - run:
          shell: true
          command: echo hello-trace
      - assert:
          exit_code: 0
```
#### When
```shell
${atago} run --verbose --report json ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout at `$.schema_version` equals `1`
- stderr contains `exit 0`
### Scenario: a failing run under --verbose renders the FAILED block exactly once
#### Given
- Fixture file `bad.atago.yaml` is created.
#### Inputs
_Fixture `bad.atago.yaml`:_
```text
version: "1"
suite:
  name: sample
scenarios:
  - name: mismatch
    steps:
      - run:
          shell: true
          command: echo hello
      - assert:
          stdout:
            contains: goodbye
```
#### When
```shell
${atago} run --verbose bad.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `FAIL assert`, `FAILED:`
- stdout does not match `/(?s)FAILED:.*FAILED:/`
## atago self-hosting / version
Source: `test/e2e/atago/version.atago.yaml`
### Scenario: version command prints the binary name
#### When
```shell
${atago} version
```
#### Then
- exit code is `0`
- stdout contains `atago`
- stderr is empty
### Scenario: unknown command is a configuration error
#### When
```shell
${atago} frobnicate
```
#### Then
- exit code is `3`
- stderr contains `unknown command`
## atago self-hosting / yaml stream matcher
Source: `test/e2e/atago/yaml.atago.yaml`
### Scenario: a yaml stream matcher selects and asserts a decoded value (#9)
#### Given
- Fixture file `yaml_ok.atago.yaml` is created.
#### Inputs
_Fixture `yaml_ok.atago.yaml`:_
```text
version: "1"
suite:
  name: yaml-matcher
scenarios:
  - name: command emits yaml and the matcher reads it
    steps:
      - fixture:
          file: data.yaml
          content: |
            name: alice
            items:
              - id: 1
              - id: 2
      - run:
          command: cat data.yaml
      - assert:
          stdout:
            yaml:
              path: "$.name"
              equals: alice
… (truncated, 5 more lines)
```
#### When
```shell
${atago} run yaml_ok.atago.yaml
```
#### Then
- exit code is `0`
- stdout contains `PASS`
- stdout does not contain `matcher not supported yet`
### Scenario: a yaml matcher mismatch fails the inner spec (#9)
#### Given
- Fixture file `yaml_fail.atago.yaml` is created.
#### Inputs
_Fixture `yaml_fail.atago.yaml`:_
```text
version: "1"
suite:
  name: yaml-matcher-fail
scenarios:
  - name: wrong expected value
    steps:
      - fixture:
          file: data.yaml
          content: |
            name: alice
      - run:
          command: cat data.yaml
      - assert:
          stdout:
            yaml:
              path: "$.name"
              equals: bob
```
#### When
```shell
${atago} run yaml_fail.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `FAILED`
