# atago Behavior Specs
## Summary
47 suites · 397 scenarios
## Contents
- [sqly ACH/Fedwire native write-back](#sqly-achfedwire-native-write-back) — 4 scenarios
  - [round-trips an ACH file through --save --force after an UPDATE](#scenario-round-trips-an-ach-file-through---save---force-after-an-update)
  - [writes a reconstructed ACH file into a directory with --save-dir](#scenario-writes-a-reconstructed-ach-file-into-a-directory-with---save-dir)
  - [round-trips a Fedwire file through --save --force after an UPDATE](#scenario-round-trips-a-fedwire-file-through---save---force-after-an-update)
  - [still rejects a single-table --output to .ach](#scenario-still-rejects-a-single-table---output-to-ach)
- [sqly batch mode (piped stdin)](#sqly-batch-mode-piped-stdin) — 16 scenarios
  - [runs SQL read from stdin](#scenario-runs-sql-read-from-stdin)
  - [switches output mode and runs the following query](#scenario-switches-output-mode-and-runs-the-following-query)
  - [exits non-zero and names the failing statement and its line on error](#scenario-exits-non-zero-and-names-the-failing-statement-and-its-line-on-error)
  - [reports the line span of a failing multiline statement](#scenario-reports-the-line-span-of-a-failing-multiline-statement)
  - [stops at .exit with a success status](#scenario-stops-at-exit-with-a-success-status)
  - [still exits non-zero when a failure precedes .exit](#scenario-still-exits-non-zero-when-a-failure-precedes-exit)
  - [runs a multiline SELECT terminated by a semicolon](#scenario-runs-a-multiline-select-terminated-by-a-semicolon)
  - [runs a multiline UNION ALL across bare newlines as one statement](#scenario-runs-a-multiline-union-all-across-bare-newlines-as-one-statement)
  - [runs a multiline WITH (CTE) query](#scenario-runs-a-multiline-with-cte-query)
  - [runs multiple statements and helper commands in one payload](#scenario-runs-multiple-statements-and-helper-commands-in-one-payload)
  - [ignores a semicolon inside a leading line comment](#scenario-ignores-a-semicolon-inside-a-leading-line-comment)
  - [ignores a semicolon inside a block comment](#scenario-ignores-a-semicolon-inside-a-block-comment)
  - [ignores a semicolon inside a trailing line comment](#scenario-ignores-a-semicolon-inside-a-trailing-line-comment)
  - [does not split on a semicolon inside a bracket-quoted identifier](#scenario-does-not-split-on-a-semicolon-inside-a-bracket-quoted-identifier)
  - [does not split on a semicolon inside a backtick-quoted identifier](#scenario-does-not-split-on-a-semicolon-inside-a-backtick-quoted-identifier)
  - [reports an error for incomplete SQL](#scenario-reports-an-error-for-incomplete-sql)
- [sqly batch fail-fast](#sqly-batch-fail-fast) — 4 scenarios
  - [stops later statements after a SQL failure](#scenario-stops-later-statements-after-a-sql-failure)
  - [does not run .save --force after an earlier failure](#scenario-does-not-run-save---force-after-an-earlier-failure)
  - [does not run .dump after an earlier failure](#scenario-does-not-run-dump-after-an-earlier-failure)
  - [does not write back for empty stdin with --save --force](#scenario-does-not-write-back-for-empty-stdin-with---save---force)
- [sqly --cache import cache](#sqly---cache-import-cache) — 8 scenarios
  - [writes a cache on the cold run and reuses it on the warm run](#scenario-writes-a-cache-on-the-cold-run-and-reuses-it-on-the-warm-run)
  - [invalidates the cache when the source changes](#scenario-invalidates-the-cache-when-the-source-changes)
  - [rebuilds after --cache-clear](#scenario-rebuilds-after---cache-clear)
  - [falls back to a cold import when the cache path is unwritable](#scenario-falls-back-to-a-cold-import-when-the-cache-path-is-unwritable)
  - [invalidates the cache when content changes but size and mtime do not (#592)](#scenario-invalidates-the-cache-when-content-changes-but-size-and-mtime-do-not-592)
  - [reuses the cache that lives inside the imported directory and ignores its manifest](#scenario-reuses-the-cache-that-lives-inside-the-imported-directory-and-ignores-its-manifest)
  - [keeps the cache warm when only an unsupported sibling file changes](#scenario-keeps-the-cache-warm-when-only-an-unsupported-sibling-file-changes)
  - [still invalidates the cache when the supported file changes](#scenario-still-invalidates-the-cache-when-the-supported-file-changes)
- [sqly changes delta cross-checks (exhaustive-set semantics)](#sqly-changes-delta-cross-checks-exhaustive-set-semantics) — 2 scenarios
  - [omitting sqly's history DB from an exhaustive list is rejected](#scenario-omitting-sqlys-history-db-from-an-exhaustive-list-is-rejected)
  - [the exhaustive list including the history DB passes](#scenario-the-exhaustive-list-including-the-history-db-passes)
- [sqly CLI surface](#sqly-cli-surface) — 10 scenarios
  - [prints the version](#scenario-prints-the-version)
  - [prints usage with --help](#scenario-prints-usage-with---help)
  - [fails on a non-existent file](#scenario-fails-on-a-non-existent-file)
  - [fails on invalid SQL with --sql](#scenario-fails-on-invalid-sql-with---sql)
  - [reports an unknown flag as a CLI error](#scenario-reports-an-unknown-flag-as-a-cli-error)
  - [reports conflicting output mode flags as a CLI error](#scenario-reports-conflicting-output-mode-flags-as-a-cli-error)
  - [joins two imported files](#scenario-joins-two-imported-files)
  - [writes JSON results to the given path with --output](#scenario-writes-json-results-to-the-given-path-with---output)
  - [applies an output-mode flag placed after the file path](#scenario-applies-an-output-mode-flag-placed-after-the-file-path)
  - [fails fast on an unknown flag after the file path](#scenario-fails-fast-on-an-unknown-flag-after-the-file-path)
- [sqly table-name collision](#sqly-table-name-collision) — 1 scenario
  - [fails when two inputs sanitize to the same table name](#scenario-fails-when-two-inputs-sanitize-to-the-same-table-name)
- [sqly --compare workflow](#sqly---compare-workflow) — 10 scenarios
  - [reports schema, row count, and keyed rows as JSON](#scenario-reports-schema-row-count-and-keyed-rows-as-json)
  - [emits a human-readable summary with --compare-format text](#scenario-emits-a-human-readable-summary-with---compare-format-text)
  - [resolves an uppercase --compare-key against a lowercase column](#scenario-resolves-an-uppercase---compare-key-against-a-lowercase-column)
  - [rejects a missing key column](#scenario-rejects-a-missing-key-column)
  - [rejects a duplicate key value as ambiguous](#scenario-rejects-a-duplicate-key-value-as-ambiguous)
  - [rejects a genuinely missing --compare-tables name](#scenario-rejects-a-genuinely-missing---compare-tables-name)
  - [rejects an ambiguous single-table compare](#scenario-rejects-an-ambiguous-single-table-compare)
  - [follows CLI input order for left and right, not table-name sorting](#scenario-follows-cli-input-order-for-left-and-right-not-table-name-sorting)
  - [reverses left and right when the inputs are swapped](#scenario-reverses-left-and-right-when-the-inputs-are-swapped)
  - [keeps the keyed diff correct on a larger input](#scenario-keeps-the-keyed-diff-correct-on-a-larger-input)
- [sqly cross-format workflows](#sqly-cross-format-workflows) — 1 scenario
  - [joins a Parquet table to a CSV table and computes a column](#scenario-joins-a-parquet-table-to-a-csv-table-and-computes-a-column)
- [sqly empty command arguments](#sqly-empty-command-arguments) — 3 scenarios
  - [rejects .save "" and leaves the source unchanged](#scenario-rejects-save--and-leaves-the-source-unchanged)
  - [rejects .dump with an empty destination](#scenario-rejects-dump-with-an-empty-destination)
  - [rejects .import with an empty path](#scenario-rejects-import-with-an-empty-path)
- [sqly excel export](#sqly-excel-export) — 2 scenarios
  - [writes a non-executable .xlsx with --output](#scenario-writes-a-non-executable-xlsx-with---output)
  - [writes a non-executable .xlsx with the .dump command](#scenario-writes-a-non-executable-xlsx-with-the-dump-command)
- [sqly export format inference](#sqly-export-format-inference) — 11 scenarios
  - [infers parquet from the destination extension without a flag](#scenario-infers-parquet-from-the-destination-extension-without-a-flag)
  - [infers ndjson with gzip and writes a compressed file](#scenario-infers-ndjson-with-gzip-and-writes-a-compressed-file)
  - [re-imports a gzip-compressed csv it wrote](#scenario-re-imports-a-gzip-compressed-csv-it-wrote)
  - [writes the CSV fallback to an unknown extension path without rewriting it](#scenario-writes-the-csv-fallback-to-an-unknown-extension-path-without-rewriting-it)
  - [rejects --output to an existing directory](#scenario-rejects---output-to-an-existing-directory)
  - [rejects .dump to an existing directory](#scenario-rejects-dump-to-an-existing-directory)
  - [errors when an explicit mode flag disagrees with the path extension](#scenario-errors-when-an-explicit-mode-flag-disagrees-with-the-path-extension)
  - [rejects bzip2 output](#scenario-rejects-bzip2-output)
  - [rejects compression on parquet](#scenario-rejects-compression-on-parquet)
  - [infers tsv from the .dump destination path](#scenario-infers-tsv-from-the-dump-destination-path)
  - [keeps --json --output result.json writing json](#scenario-keeps---json---output-resultjson-writing-json)
- [sqly helper commands reject extra args](#sqly-helper-commands-reject-extra-args) — 7 scenarios
  - [rejects .schema with an extra argument](#scenario-rejects-schema-with-an-extra-argument)
  - [rejects .describe with an extra argument](#scenario-rejects-describe-with-an-extra-argument)
  - [rejects .tables with an extra argument](#scenario-rejects-tables-with-an-extra-argument)
  - [rejects .mode with an extra argument](#scenario-rejects-mode-with-an-extra-argument)
  - [rejects .pwd with an extra argument](#scenario-rejects-pwd-with-an-extra-argument)
  - [rejects .clear with an extra argument](#scenario-rejects-clear-with-an-extra-argument)
  - [does not let .exit with an extra argument silently terminate the batch](#scenario-does-not-let-exit-with-an-extra-argument-silently-terminate-the-batch)
- [sqly filesql integration](#sqly-filesql-integration) — 8 scenarios
  - [imports and queries a CSV file](#scenario-imports-and-queries-a-csv-file)
  - [imports and queries a JSONL file](#scenario-imports-and-queries-a-jsonl-file)
  - [imports and queries a Parquet file](#scenario-imports-and-queries-a-parquet-file)
  - [imports and queries an Excel file](#scenario-imports-and-queries-an-excel-file)
  - [imports and queries an ACH file](#scenario-imports-and-queries-an-ach-file)
  - [imports and queries a Fedwire file](#scenario-imports-and-queries-a-fedwire-file)
  - [preserves filesql-detected column types for schema inspection](#scenario-preserves-filesql-detected-column-types-for-schema-inspection)
  - [produces identical tables on repeated ACH imports](#scenario-produces-identical-tables-on-repeated-ach-imports)
- [sqly shell helper commands](#sqly-shell-helper-commands) — 9 scenarios
  - [.help groups commands, shows usage, and flags destructive save](#scenario-help-groups-commands-shows-usage-and-flags-destructive-save)
  - [.cd changes directory with a relative path and reports it](#scenario-cd-changes-directory-with-a-relative-path-and-reports-it)
  - [expands a bare ~ in .cd to the home directory](#scenario-expands-a-bare--in-cd-to-the-home-directory)
  - [expands ~/file in .import](#scenario-expands-file-in-import)
  - [.clear emits no ANSI escapes to stdout in batch mode](#scenario-clear-emits-no-ansi-escapes-to-stdout-in-batch-mode)
  - [keeps --json stdout parseable when .clear precedes a query](#scenario-keeps---json-stdout-parseable-when-clear-precedes-a-query)
  - [.import imports a quoted path containing a space as one argument](#scenario-import-imports-a-quoted-path-containing-a-space-as-one-argument)
  - [.mode fails on a missing mode name but still lists the modes](#scenario-mode-fails-on-a-missing-mode-name-but-still-lists-the-modes)
  - [.dump infers TSV from the .tsv extension in table mode, not CSV](#scenario-dump-infers-tsv-from-the-tsv-extension-in-table-mode-not-csv)
- [sqly hermetic environment](#sqly-hermetic-environment) — 3 scenarios
  - [runs with HOME inside the sandbox](#scenario-runs-with-home-inside-the-sandbox)
  - [routes the history DB into the sandbox](#scenario-routes-the-history-db-into-the-sandbox)
  - [still runs sqly normally inside the sandbox](#scenario-still-runs-sqly-normally-inside-the-sandbox)
- [sqly history tolerance](#sqly-history-tolerance) — 5 scenarios
  - [runs --sql and warns instead of failing](#scenario-runs---sql-and-warns-instead-of-failing)
  - [runs batch mode without a writable history DB](#scenario-runs-batch-mode-without-a-writable-history-db)
  - [runs --sql when the history DB is read-only after startup](#scenario-runs---sql-when-the-history-db-is-read-only-after-startup)
  - [runs --inspect when the history DB is read-only after startup](#scenario-runs---inspect-when-the-history-db-is-read-only-after-startup)
  - [runs batch mode and warns when a history write fails after startup](#scenario-runs-batch-mode-and-warns-when-a-history-write-fails-after-startup)
- [sqly import failure handling](#sqly-import-failure-handling) — 5 scenarios
  - [fails query mode on a partial import and keeps stdout clean](#scenario-fails-query-mode-on-a-partial-import-and-keeps-stdout-clean)
  - [names the failed count and first path when every input fails](#scenario-names-the-failed-count-and-first-path-when-every-input-fails)
  - [fails --inspect on a partial import](#scenario-fails---inspect-on-a-partial-import)
  - [fails batch .import on a partial import and stops later commands](#scenario-fails-batch-import-on-a-partial-import-and-stops-later-commands)
  - [keeps stdout clean when a stdin dataset fails to import](#scenario-keeps-stdout-clean-when-a-stdin-dataset-fails-to-import)
- [sqly .import with space-containing paths](#sqly-import-with-space-containing-paths) — 5 scenarios
  - [imports a backslash-escaped space path as a single argument](#scenario-imports-a-backslash-escaped-space-path-as-a-single-argument)
  - [imports a double-quoted space path as a single argument](#scenario-imports-a-double-quoted-space-path-as-a-single-argument)
  - [imports a single-quoted space path as a single argument](#scenario-imports-a-single-quoted-space-path-as-a-single-argument)
  - [imports a file inside a space-containing directory when escaped](#scenario-imports-a-file-inside-a-space-containing-directory-when-escaped)
  - [splits an unquoted space path into two failing arguments](#scenario-splits-an-unquoted-space-path-into-two-failing-arguments)
- [sqly --inspect](#sqly---inspect) — 7 scenarios
  - [prints a JSON report for a single file](#scenario-prints-a-json-report-for-a-single-file)
  - [maps every table from a multi-table ACH file to its source](#scenario-maps-every-table-from-a-multi-table-ach-file-to-its-source)
  - [keeps stdout as pure JSON for a directory and stays quiet on stderr](#scenario-keeps-stdout-as-pure-json-for-a-directory-and-stays-quiet-on-stderr)
  - [fails with a clear error when no input is given](#scenario-fails-with-a-clear-error-when-no-input-is-given)
  - [emits a schema-only report with --inspect-sample 0](#scenario-emits-a-schema-only-report-with---inspect-sample-0)
  - [limits sample rows with --inspect-sample](#scenario-limits-sample-rows-with---inspect-sample)
  - [rejects a negative --inspect-sample](#scenario-rejects-a-negative---inspect-sample)
- [sqly --inspect conflicts](#sqly---inspect-conflicts) — 3 scenarios
  - [rejects --inspect with --sql](#scenario-rejects---inspect-with---sql)
  - [rejects --inspect with --output and writes no file](#scenario-rejects---inspect-with---output-and-writes-no-file)
  - [rejects --inspect with --save-dir and creates no save dir](#scenario-rejects---inspect-with---save-dir-and-creates-no-save-dir)
- [sqly JSON/NDJSON NULL handling](#sqly-jsonndjson-null-handling) — 2 scenarios
  - [emits NULL as JSON null in --json](#scenario-emits-null-as-json-null-in---json)
  - [emits NULL as JSON null in --ndjson](#scenario-emits-null-as-json-null-in---ndjson)
- [sqly metamorphic relations](#sqly-metamorphic-relations) — 5 scenarios
  - [COUNT(*) equals the number of selected rows](#scenario-count-equals-the-number-of-selected-rows)
  - [WHERE 1=1 returns all rows and WHERE 1=0 returns none](#scenario-where-11-returns-all-rows-and-where-10-returns-none)
  - [ORDER BY preserves the row multiset](#scenario-order-by-preserves-the-row-multiset)
  - [csv and ndjson yield the same rows](#scenario-csv-and-ndjson-yield-the-same-rows)
  - [CSV dump reimported yields identical data](#scenario-csv-dump-reimported-yields-identical-data)
- [sqly .mode banner routing](#sqly-mode-banner-routing) — 3 scenarios
  - [keeps stdout pure JSON after .mode json](#scenario-keeps-stdout-pure-json-after-mode-json)
  - [keeps stdout pure NDJSON after .mode ndjson](#scenario-keeps-stdout-pure-ndjson-after-mode-ndjson)
  - [reports the typed mode by name and emits typed output after .mode json-typed](#scenario-reports-the-typed-mode-by-name-and-emits-typed-output-after-mode-json-typed)
- [sqly output formats](#sqly-output-formats) — 5 scenarios
  - [--json renders results as a JSON array](#scenario---json-renders-results-as-a-json-array)
  - [--json prints \[\] for an empty result](#scenario---json-prints--for-an-empty-result)
  - [--ndjson renders one JSON object per line](#scenario---ndjson-renders-one-json-object-per-line)
  - [--ndjson prints nothing for an empty result](#scenario---ndjson-prints-nothing-for-an-empty-result)
  - [--csv renders header and rows as CSV](#scenario---csv-renders-header-and-rows-as-csv)
- [sqly --output requires --sql](#sqly---output-requires---sql) — 2 scenarios
  - [rejects --output with no query](#scenario-rejects---output-with-no-query)
  - [rejects --output for batch SQL from stdin](#scenario-rejects---output-for-batch-sql-from-stdin)
- [sqly file-output status routing](#sqly-file-output-status-routing) — 3 scenarios
  - [keeps stdout empty for --output and reports on stderr](#scenario-keeps-stdout-empty-for---output-and-reports-on-stderr)
  - [keeps stdout free of the .dump status line](#scenario-keeps-stdout-free-of-the-dump-status-line)
  - [keeps the .save confirmation off stdout](#scenario-keeps-the-save-confirmation-off-stdout)
- [sqly parquet export](#sqly-parquet-export) — 6 scenarios
  - [writes a parquet file that re-imports with the same rows](#scenario-writes-a-parquet-file-that-re-imports-with-the-same-rows)
  - [appends the .parquet extension](#scenario-appends-the-parquet-extension)
  - [writes query results to the given parquet path with --output](#scenario-writes-query-results-to-the-given-parquet-path-with---output)
  - [preserves leading-zero codes through a parquet round-trip](#scenario-preserves-leading-zero-codes-through-a-parquet-round-trip)
  - [preserves SQL NULL through a parquet round-trip](#scenario-preserves-sql-null-through-a-parquet-round-trip)
  - [reports a clear error when exporting an empty result](#scenario-reports-a-clear-error-when-exporting-an-empty-result)
- [sqly input path validation](#sqly-input-path-validation) — 4 scenarios
  - [imports a deeply nested path](#scenario-imports-a-deeply-nested-path)
  - [imports a file whose name literally contains ..%2f](#scenario-imports-a-file-whose-name-literally-contains-2f)
  - [rejects a symlink alias that resolves to a blocked system path](#scenario-rejects-a-symlink-alias-that-resolves-to-a-blocked-system-path)
  - [imports a symlink alias that resolves to an ordinary user file](#scenario-imports-a-symlink-alias-that-resolves-to-an-ordinary-user-file)
- [sqly --profile workflow](#sqly---profile-workflow) — 10 scenarios
  - [reports per-column data quality as JSON](#scenario-reports-per-column-data-quality-as-json)
  - [profiles a stdin dataset](#scenario-profiles-a-stdin-dataset)
  - [profiles multiple tables in one run](#scenario-profiles-multiple-tables-in-one-run)
  - [emits a human-readable summary with --profile-format text](#scenario-emits-a-human-readable-summary-with---profile-format-text)
  - [counts a blank string as a distinct value in JSON output](#scenario-counts-a-blank-string-as-a-distinct-value-in-json-output)
  - [counts a blank string as a distinct value in text output](#scenario-counts-a-blank-string-as-a-distinct-value-in-text-output)
  - [flags a padded null-like placeholder and its whitespace together](#scenario-flags-a-padded-null-like-placeholder-and-its-whitespace-together)
  - [warns only about whitespace for a padded ordinary value](#scenario-warns-only-about-whitespace-for-a-padded-ordinary-value)
  - [counts comma-formatted numerals as numeric, matching table-mode](#scenario-counts-comma-formatted-numerals-as-numeric-matching-table-mode)
  - [right-aligns the same comma-formatted column in table-mode](#scenario-right-aligns-the-same-comma-formatted-column-in-table-mode)
- [README examples](#readme-examples) — 29 scenarios
  - [prints the full user table as an ASCII table](#scenario-prints-the-full-user-table-as-an-ascii-table)
  - [joins two files on a shared key](#scenario-joins-two-files-on-a-shared-key)
  - [runs the analytics script (CTE + window + GROUP BY)](#scenario-runs-the-analytics-script-cte--window--group-by)
  - [extracts JSON fields from a JSONL file](#scenario-extracts-json-fields-from-a-jsonl-file)
  - [reads a gzipped CSV transparently](#scenario-reads-a-gzipped-csv-transparently)
  - [queries a Parquet file](#scenario-queries-a-parquet-file)
  - [joins a compressed CSV with a plain CSV](#scenario-joins-a-compressed-csv-with-a-plain-csv)
  - [renders CSV with --csv](#scenario-renders-csv-with---csv)
  - [renders JSON with --json](#scenario-renders-json-with---json)
  - [renders NDJSON with --ndjson](#scenario-renders-ndjson-with---ndjson)
  - [renders a markdown table with --markdown](#scenario-renders-a-markdown-table-with---markdown)
  - [renders LTSV with --ltsv](#scenario-renders-ltsv-with---ltsv)
  - [writes CSV to the path given by --output](#scenario-writes-csv-to-the-path-given-by---output)
  - [queries piped CSV through the default stdin table](#scenario-queries-piped-csv-through-the-default-stdin-table)
  - [joins piped stdin with a file argument](#scenario-joins-piped-stdin-with-a-file-argument)
  - [runs a helper command and a query from piped stdin](#scenario-runs-a-helper-command-and-a-query-from-piped-stdin)
  - [runs SQL from join.sql while stdin carries a dataset](#scenario-runs-sql-from-joinsql-while-stdin-carries-a-dataset)
  - [prints a JSON inspect report with a stable source and column types](#scenario-prints-a-json-inspect-report-with-a-stable-source-and-column-types)
  - [omits sample rows with --inspect-sample 0](#scenario-omits-sample-rows-with---inspect-sample-0)
  - [prints the CREATE TABLE statement with .schema](#scenario-prints-the-create-table-statement-with-schema)
  - [prints column information with .describe](#scenario-prints-column-information-with-describe)
  - [writes the updated table to --save-dir, leaving the source untouched](#scenario-writes-the-updated-table-to---save-dir-leaving-the-source-untouched)
  - [rejects --save without --force](#scenario-rejects---save-without---force)
  - [rejects a schema-changing statement under --save-dir before writing anything](#scenario-rejects-a-schema-changing-statement-under---save-dir-before-writing-anything)
  - [overwrites the source in place with --save --force](#scenario-overwrites-the-source-in-place-with---save---force)
  - [imports every supported file under a directory](#scenario-imports-every-supported-file-under-a-directory)
  - [loads ACH records into multiple tables](#scenario-loads-ach-records-into-multiple-tables)
  - [queries the ACH entries table](#scenario-queries-the-ach-entries-table)
  - [loads a Fedwire file into a single message table](#scenario-loads-a-fedwire-file-into-a-single-message-table)
- [sqly interactive shell (pty)](#sqly-interactive-shell-pty) — 4 scenarios
  - [run a query and read its rendered result table over a pty](#scenario-run-a-query-and-read-its-rendered-result-table-over-a-pty)
  - [the .tables dot-command renders the imported table over a pty](#scenario-the-tables-dot-command-renders-the-imported-table-over-a-pty)
  - [a computed aggregate round-trips through the pty shell](#scenario-a-computed-aggregate-round-trips-through-the-pty-shell)
  - [the prompt reflects a live .mode switch over a pty](#scenario-the-prompt-reflects-a-live-mode-switch-over-a-pty)
- [sqly sandbox_home + changes (history DB isolation)](#sqly-sandbox_home--changes-history-db-isolation) — 2 scenarios
  - [sqly --sql writes exactly its history DB, only inside the sandbox home](#scenario-sqly---sql-writes-exactly-its-history-db-only-inside-the-sandbox-home)
  - [a second sqly batch run leaves its sandbox home byte-identical](#scenario-a-second-sqly-batch-run-leaves-its-sandbox-home-byte-identical)
- [sqly write-back](#sqly-write-back) — 8 scenarios
  - [writes to --save-dir without modifying the source](#scenario-writes-to---save-dir-without-modifying-the-source)
  - [refuses --save without --force](#scenario-refuses---save-without---force)
  - [overwrites the source in place with --save --force](#scenario-overwrites-the-source-in-place-with---save---force-1)
  - [re-imports a file rewritten in place (round-trip)](#scenario-re-imports-a-file-rewritten-in-place-round-trip)
  - [preserves gzip compression on in-place save](#scenario-preserves-gzip-compression-on-in-place-save)
  - [saves via the .save command in batch mode](#scenario-saves-via-the-save-command-in-batch-mode)
  - [guides a non-interactive --save with no input files toward passing input](#scenario-guides-a-non-interactive---save-with-no-input-files-toward-passing-input)
  - [guides a batch .save with no imported tables toward passing input](#scenario-guides-a-batch-save-with-no-imported-tables-toward-passing-input)
- [sqly schema inspection](#sqly-schema-inspection) — 7 scenarios
  - [.schema prints a CREATE TABLE statement for a CSV table](#scenario-schema-prints-a-create-table-statement-for-a-csv-table)
  - [.schema emits a structured object in json mode](#scenario-schema-emits-a-structured-object-in-json-mode)
  - [.schema returns the stored CREATE VIEW for a differently cased view name](#scenario-schema-returns-the-stored-create-view-for-a-differently-cased-view-name)
  - [.schema errors on a missing table](#scenario-schema-errors-on-a-missing-table)
  - [.describe lists columns and types for a CSV table](#scenario-describe-lists-columns-and-types-for-a-csv-table)
  - [.describe emits structured column metadata in json mode](#scenario-describe-emits-structured-column-metadata-in-json-mode)
  - [.describe errors on a missing table](#scenario-describe-errors-on-a-missing-table)
- [sqly --sheet validation](#sqly---sheet-validation) — 8 scenarios
  - [rejects --sheet with a non-Excel file and --sql](#scenario-rejects---sheet-with-a-non-excel-file-and---sql)
  - [rejects --sheet with a non-Excel file and --inspect](#scenario-rejects---sheet-with-a-non-excel-file-and---inspect)
  - [still imports an Excel file with --sheet](#scenario-still-imports-an-excel-file-with---sheet)
  - [rejects an explicit empty --sheet](#scenario-rejects-an-explicit-empty---sheet)
  - [rejects --sheet for a directory with no Excel files](#scenario-rejects---sheet-for-a-directory-with-no-excel-files)
  - [tells the user how to recover when --sheet has no Excel input](#scenario-tells-the-user-how-to-recover-when---sheet-has-no-excel-input)
  - [names the workbook and suggests recovery on a single-workbook sheet miss](#scenario-names-the-workbook-and-suggests-recovery-on-a-single-workbook-sheet-miss)
  - [names every checked workbook on a multi-workbook sheet miss](#scenario-names-every-checked-workbook-on-a-multi-workbook-sheet-miss)
- [sqly](#sqly) — 3 scenarios
  - [count rows in a CSV fixture](#scenario-count-rows-in-a-csv-fixture)
  - [filter rows and select a column](#scenario-filter-rows-and-select-a-column)
  - [markdown output format works](#scenario-markdown-output-format-works)
- [sqly --sql-file](#sqly---sql-file) — 7 scenarios
  - [runs a multiline query loaded from a file against a file input](#scenario-runs-a-multiline-query-loaded-from-a-file-against-a-file-input)
  - [joins a piped --stdin dataset with a query loaded from a file](#scenario-joins-a-piped---stdin-dataset-with-a-query-loaded-from-a-file)
  - [runs multiple statements from a file in order](#scenario-runs-multiple-statements-from-a-file-in-order)
  - [rejects --sql and --sql-file together](#scenario-rejects---sql-and---sql-file-together)
  - [fails for a missing SQL file](#scenario-fails-for-a-missing-sql-file)
  - [fails for an empty SQL file](#scenario-fails-for-an-empty-sql-file)
  - [locates a failing statement by its line in the SQL file](#scenario-locates-a-failing-statement-by-its-line-in-the-sql-file)
- [sqly --sql-file --output](#sqly---sql-file---output) — 4 scenarios
  - [exports a single-SELECT script to the output file with clean stdout](#scenario-exports-a-single-select-script-to-the-output-file-with-clean-stdout)
  - [exports a single result set even when the script first runs DDL/DML](#scenario-exports-a-single-result-set-even-when-the-script-first-runs-ddldml)
  - [rejects a script that produces no result set](#scenario-rejects-a-script-that-produces-no-result-set)
  - [rejects a script that produces multiple result sets](#scenario-rejects-a-script-that-produces-multiple-result-sets)
- [sqly --stdin dataset](#sqly---stdin-dataset) — 11 scenarios
  - [queries piped CSV through the default stdin table](#scenario-queries-piped-csv-through-the-default-stdin-table-1)
  - [queries piped TSV data](#scenario-queries-piped-tsv-data)
  - [queries piped JSONL data stored in a data column](#scenario-queries-piped-jsonl-data-stored-in-a-data-column)
  - [overrides the stdin table name with --stdin-name](#scenario-overrides-the-stdin-table-name-with---stdin-name)
  - [joins piped stdin with an imported file argument](#scenario-joins-piped-stdin-with-an-imported-file-argument)
  - [reports a stable stdin source in --inspect, not a temp path](#scenario-reports-a-stable-stdin-source-in---inspect-not-a-temp-path)
  - [rejects --save --force for a stdin-backed table](#scenario-rejects---save---force-for-a-stdin-backed-table)
  - [rejects a non-identifier --stdin-name so the name stays queryable](#scenario-rejects-a-non-identifier---stdin-name-so-the-name-stays-queryable)
  - [rejects a path-like --stdin-name](#scenario-rejects-a-path-like---stdin-name)
  - [reports a clear error for an unsupported stdin format](#scenario-reports-a-clear-error-for-an-unsupported-stdin-format)
  - [still reads stdin as SQL and helper commands without --stdin](#scenario-still-reads-stdin-as-sql-and-helper-commands-without---stdin)
- [sqly typed JSON output](#sqly-typed-json-output) — 7 scenarios
  - [emits native numbers, booleans, and null with --json-typed](#scenario-emits-native-numbers-booleans-and-null-with---json-typed)
  - [keeps the legacy string contract with plain --json](#scenario-keeps-the-legacy-string-contract-with-plain---json)
  - [emits native scalars per line with --ndjson-typed](#scenario-emits-native-scalars-per-line-with---ndjson-typed)
  - [keeps a large integer column lossless (no scientific notation)](#scenario-keeps-a-large-integer-column-lossless-no-scientific-notation)
  - [leaves a leading-zero value as a string](#scenario-leaves-a-leading-zero-value-as-a-string)
  - [uses the typed contract for --inspect sample rows](#scenario-uses-the-typed-contract-for---inspect-sample-rows)
  - [rejects plain --json combined with --inspect](#scenario-rejects-plain---json-combined-with---inspect)
- [sqly v0.18.0 binary bug fixes](#sqly-v0180-binary-bug-fixes) — 22 scenarios
  - [rejects an empty --output](#scenario-rejects-an-empty---output)
  - [rejects an empty --sql-file](#scenario-rejects-an-empty---sql-file)
  - [rejects an empty --save-dir](#scenario-rejects-an-empty---save-dir)
  - [rejects an empty --stdin](#scenario-rejects-an-empty---stdin)
  - [rejects conflicting output mode flags](#scenario-rejects-conflicting-output-mode-flags)
  - [prints rows for a DML RETURNING statement](#scenario-prints-rows-for-a-dml-returning-statement)
  - [rejects --output for a non-rowset DML statement](#scenario-rejects---output-for-a-non-rowset-dml-statement)
  - [exports RETURNING rows with --output](#scenario-exports-returning-rows-with---output)
  - [rejects a comment-only --sql-file](#scenario-rejects-a-comment-only---sql-file)
  - [strips a UTF-8 BOM from a --sql-file script](#scenario-strips-a-utf-8-bom-from-a---sql-file-script)
  - [strips a UTF-8 BOM from batch stdin](#scenario-strips-a-utf-8-bom-from-batch-stdin)
  - [rejects non-empty piped stdin with --sql-file](#scenario-rejects-non-empty-piped-stdin-with---sql-file)
  - [fails a --stdin dataset run with no query](#scenario-fails-a---stdin-dataset-run-with-no-query)
  - [reports per-file provenance for a sanitized basename](#scenario-reports-per-file-provenance-for-a-sanitized-basename)
  - [rejects duplicate basenames from different subdirectories](#scenario-rejects-duplicate-basenames-from-different-subdirectories)
  - [reports an overwrite when re-importing a directory](#scenario-reports-an-overwrite-when-re-importing-a-directory)
  - [rejects --output that aliases an imported source](#scenario-rejects---output-that-aliases-an-imported-source)
  - [rejects --save-dir that resolves to the source directory](#scenario-rejects---save-dir-that-resolves-to-the-source-directory)
  - [rejects a --save-dir destination that already exists](#scenario-rejects-a---save-dir-destination-that-already-exists)
  - [keeps stdout clean when write-back fails](#scenario-keeps-stdout-clean-when-write-back-fails)
  - [skips write-back for a read-only query under --save --force](#scenario-skips-write-back-for-a-read-only-query-under---save---force)
  - [skips workbooks lacking the requested sheet (multi-workbook --sheet)](#scenario-skips-workbooks-lacking-the-requested-sheet-multi-workbook---sheet)
- [sqly v0.19.0 binary bug fixes](#sqly-v0190-binary-bug-fixes) — 26 scenarios
  - [quotes a CSV value containing a comma](#scenario-quotes-a-csv-value-containing-a-comma)
  - [quotes a CSV value containing a double quote](#scenario-quotes-a-csv-value-containing-a-double-quote)
  - [rejects an LTSV value containing a tab](#scenario-rejects-an-ltsv-value-containing-a-tab)
  - [rejects duplicate JSON keys](#scenario-rejects-duplicate-json-keys)
  - [rejects duplicate NDJSON keys](#scenario-rejects-duplicate-ndjson-keys)
  - [keeps a Markdown row on one line when a value has a newline](#scenario-keeps-a-markdown-row-on-one-line-when-a-value-has-a-newline)
  - [accepts a leading block comment in direct --sql](#scenario-accepts-a-leading-block-comment-in-direct---sql)
  - [accepts PRAGMA in direct --sql](#scenario-accepts-pragma-in-direct---sql)
  - [accepts VALUES in direct --sql](#scenario-accepts-values-in-direct---sql)
  - [accepts the TABLE shorthand in direct --sql](#scenario-accepts-the-table-shorthand-in-direct---sql)
  - [accepts CREATE TABLE in direct --sql](#scenario-accepts-create-table-in-direct---sql)
  - [accepts ANALYZE in direct --sql](#scenario-accepts-analyze-in-direct---sql)
  - [runs WITH ... UPDATE without RETURNING as DML](#scenario-runs-with--update-without-returning-as-dml)
  - [rejects --stdin-name without --stdin](#scenario-rejects---stdin-name-without---stdin)
  - [rejects --inspect-sample without --inspect](#scenario-rejects---inspect-sample-without---inspect)
  - [rejects --force without --save](#scenario-rejects---force-without---save)
  - [rejects --inspect combined with --csv](#scenario-rejects---inspect-combined-with---csv)
  - [imports an empty JSON array as a zero-row table](#scenario-imports-an-empty-json-array-as-a-zero-row-table)
  - [imports an empty JSONL file as a zero-row table](#scenario-imports-an-empty-jsonl-file-as-a-zero-row-table)
  - [rejects an --output path ending with a slash](#scenario-rejects-an---output-path-ending-with-a-slash)
  - [rejects an --output ACH destination](#scenario-rejects-an---output-ach-destination)
  - [parses a helper command after a terminated statement](#scenario-parses-a-helper-command-after-a-terminated-statement)
  - [parses a helper command after a leading comment](#scenario-parses-a-helper-command-after-a-leading-comment)
  - [does not write back for an EXPLAIN under --save-dir](#scenario-does-not-write-back-for-an-explain-under---save-dir)
  - [does not write back for a zero-row DML under --save-dir](#scenario-does-not-write-back-for-a-zero-row-dml-under---save-dir)
  - [keeps stdout clean when parquet write-back fails](#scenario-keeps-stdout-clean-when-parquet-write-back-fails)
- [sqly v0.20.0 binary regressions](#sqly-v0200-binary-regressions) — 32 scenarios
  - [write-back rejects: ALTER TABLE RENAME COLUMN](#scenario-write-back-rejects-alter-table-rename-column)
  - [write-back rejects: DROP TABLE](#scenario-write-back-rejects-drop-table)
  - [write-back rejects: CREATE VIEW](#scenario-write-back-rejects-create-view)
  - [write-back rejects: CREATE INDEX](#scenario-write-back-rejects-create-index)
  - [write-back rejects: CREATE TABLE](#scenario-write-back-rejects-create-table)
  - [write-back rejects: REINDEX](#scenario-write-back-rejects-reindex)
  - [write-back rejects: ANALYZE](#scenario-write-back-rejects-analyze)
  - [rejects CREATE TABLE AS SELECT under --save-dir and writes nothing](#scenario-rejects-create-table-as-select-under---save-dir-and-writes-nothing)
  - [preflight rejects a CTAS+DML script before it executes](#scenario-preflight-rejects-a-ctasdml-script-before-it-executes)
  - [allows a .import + UPDATE batch under --save-dir and writes the change](#scenario-allows-a-import--update-batch-under---save-dir-and-writes-the-change)
  - [neutral success: CREATE VIEW](#scenario-neutral-success-create-view)
  - [neutral success: CREATE TABLE](#scenario-neutral-success-create-table)
  - [neutral success: ANALYZE](#scenario-neutral-success-analyze)
  - [runs a setter PRAGMA](#scenario-runs-a-setter-pragma)
  - [runs a command PRAGMA that returns no rows](#scenario-runs-a-command-pragma-that-returns-no-rows)
  - [rejects BEGIN in a --sql-file script](#scenario-rejects-begin-in-a---sql-file-script)
  - [rejects VACUUM](#scenario-rejects-vacuum)
  - [rejects VACUUM INTO and writes no file](#scenario-rejects-vacuum-into-and-writes-no-file)
  - [rejects ATTACH DATABASE and persists no external file](#scenario-rejects-attach-database-and-persists-no-external-file)
  - [runs a multiline CREATE TRIGGER from a --sql-file](#scenario-runs-a-multiline-create-trigger-from-a---sql-file)
  - [accepts a schema-qualified .schema name](#scenario-accepts-a-schema-qualified-schema-name)
  - [lists session-created views and temp tables in .tables](#scenario-lists-session-created-views-and-temp-tables-in-tables)
  - [prints CREATE VIEW for a view in .schema](#scenario-prints-create-view-for-a-view-in-schema)
  - [imports an empty compressed JSON array as a zero-row table](#scenario-imports-an-empty-compressed-json-array-as-a-zero-row-table)
  - [imports an empty compressed JSONL file as a zero-row table](#scenario-imports-an-empty-compressed-jsonl-file-as-a-zero-row-table)
  - [imports /dev/stdin as CSV](#scenario-imports-devstdin-as-csv)
  - [imports /proc/self/fd/0 as CSV](#scenario-imports-procselffd0-as-csv)
  - [rejects --output to a multi-compressed ACH destination](#scenario-rejects---output-to-a-multi-compressed-ach-destination)
  - [rejects .dump to a multi-compressed Fedwire destination](#scenario-rejects-dump-to-a-multi-compressed-fedwire-destination)
  - [rejects an invalid LTSV output label](#scenario-rejects-an-invalid-ltsv-output-label)
  - [rejects duplicate LTSV output labels](#scenario-rejects-duplicate-ltsv-output-labels)
  - [rejects an LTSV import with duplicate labels](#scenario-rejects-an-ltsv-import-with-duplicate-labels)
- [sqly v0.21.0 binary regressions](#sqly-v0210-binary-regressions) — 30 scenarios
  - [prefers a TEMP table over a same-named main table in .schema](#scenario-prefers-a-temp-table-over-a-same-named-main-table-in-schema)
  - [prefers a TEMP view over a same-named main table in .schema](#scenario-prefers-a-temp-view-over-a-same-named-main-table-in-schema)
  - [keeps both a main and a same-named TEMP object in .tables](#scenario-keeps-both-a-main-and-a-same-named-temp-object-in-tables)
  - [targets a literal dotted table name in .schema](#scenario-targets-a-literal-dotted-table-name-in-schema)
  - [targets a literal dotted table name in .describe](#scenario-targets-a-literal-dotted-table-name-in-describe)
  - [targets a literal dotted table name in .header](#scenario-targets-a-literal-dotted-table-name-in-header)
  - [targets a literal dotted table name in .dump](#scenario-targets-a-literal-dotted-table-name-in-dump)
  - [prints a paste-safe quoted identifier in .tables](#scenario-prints-a-paste-safe-quoted-identifier-in-tables)
  - [keeps the full spaced table name in .header](#scenario-keeps-the-full-spaced-table-name-in-header)
  - [keeps the TEMP keyword for a temp-qualified table in .schema](#scenario-keeps-the-temp-keyword-for-a-temp-qualified-table-in-schema)
  - [keeps the TEMP keyword for a temp-qualified view in .schema](#scenario-keeps-the-temp-keyword-for-a-temp-qualified-view-in-schema)
  - [direct --sql rejects multi-statement input: two SELECTs](#scenario-direct---sql-rejects-multi-statement-input-two-selects)
  - [direct --sql rejects multi-statement input: SELECT then UPDATE](#scenario-direct---sql-rejects-multi-statement-input-select-then-update)
  - [rejects multi-statement --sql --output before writing the file](#scenario-rejects-multi-statement---sql---output-before-writing-the-file)
  - [rejects under --save --force: PRAGMA user_version=1](#scenario-rejects-under---save---force-pragma-user_version1)
  - [rejects under --save --force: PRAGMA incremental_vacuum](#scenario-rejects-under---save---force-pragma-incremental_vacuum)
  - [rejects under --save --force: PRAGMA journal_mode=OFF](#scenario-rejects-under---save---force-pragma-journal_modeoff)
  - [rejects a setter PRAGMA under --save-dir and writes nothing](#scenario-rejects-a-setter-pragma-under---save-dir-and-writes-nothing)
  - [rejects a command PRAGMA under --save-dir and writes nothing](#scenario-rejects-a-command-pragma-under---save-dir-and-writes-nothing)
  - [rejects END in direct --sql](#scenario-rejects-end-in-direct---sql)
  - [rejects END in batch stdin](#scenario-rejects-end-in-batch-stdin)
  - [rejects END in a --sql-file script](#scenario-rejects-end-in-a---sql-file-script)
  - [--output rejects nested compression suffixes: out.csv.gz.zst](#scenario---output-rejects-nested-compression-suffixes-outcsvgzzst)
  - [--output rejects nested compression suffixes: out.parquet.gz.zst](#scenario---output-rejects-nested-compression-suffixes-outparquetgzzst)
  - [--output rejects nested compression suffixes: out.xlsx.gz.zst](#scenario---output-rejects-nested-compression-suffixes-outxlsxgzzst)
  - [rejects a nested .dump destination and writes nothing](#scenario-rejects-a-nested-dump-destination-and-writes-nothing)
  - [emits structured .tables output under .mode json](#scenario-emits-structured-tables-output-under-mode-json)
  - [emits structured .header output under .mode ndjson](#scenario-emits-structured-header-output-under-mode-ndjson)
  - [does not rewrite an unchanged source on .save --force](#scenario-does-not-rewrite-an-unchanged-source-on-save---force)
  - [writes no directory export for an unchanged session on .save DIR](#scenario-writes-no-directory-export-for-an-unchanged-session-on-save-dir)
- [sqly v0.22.0 binary regressions](#sqly-v0220-binary-regressions) — 18 scenarios
  - [inspects a literal "main.x" table with .schema](#scenario-inspects-a-literal-mainx-table-with-schema)
  - [inspects a literal "temp.x" table with .describe](#scenario-inspects-a-literal-tempx-table-with-describe)
  - [inspects a literal "main.v" view with .header](#scenario-inspects-a-literal-mainv-view-with-header)
  - [exports a literal "temp.v" view with .dump](#scenario-exports-a-literal-tempv-view-with-dump)
  - [prints a paste-safe literal "main.x" name in .tables](#scenario-prints-a-paste-safe-literal-mainx-name-in-tables)
  - [rejects a --output destination that stacks .gzip and .zst on a format suffix](#scenario-rejects-a---output-destination-that-stacks-gzip-and-zst-on-a-format-suffix)
  - [rejects a --output .json.gzip.zst destination](#scenario-rejects-a---output-jsongzipzst-destination)
  - [rejects a --output .ach.gzip.zst destination as input-only](#scenario-rejects-a---output-achgzipzst-destination-as-input-only)
  - [rejects a .dump destination that stacks .gzip and .zst](#scenario-rejects-a-dump-destination-that-stacks-gzip-and-zst)
  - [runs the SELECT after a leading empty statement in direct --sql](#scenario-runs-the-select-after-a-leading-empty-statement-in-direct---sql)
  - [runs the SELECT after multiple leading empty statements in direct --sql](#scenario-runs-the-select-after-multiple-leading-empty-statements-in-direct---sql)
  - [exports the SELECT after a leading empty statement with --output](#scenario-exports-the-select-after-a-leading-empty-statement-with---output)
  - [still rejects ATTACH after a leading empty statement in direct --sql](#scenario-still-rejects-attach-after-a-leading-empty-statement-in-direct---sql)
  - [does not rewrite an unchanged CSV when only a TEMP table changed](#scenario-does-not-rewrite-an-unchanged-csv-when-only-a-temp-table-changed)
  - [does not fail on an unchanged JSONL import when only a scratch table changed](#scenario-does-not-fail-on-an-unchanged-jsonl-import-when-only-a-scratch-table-changed)
  - [does not rewrite the source after net-zero CSV edits](#scenario-does-not-rewrite-the-source-after-net-zero-csv-edits)
  - [does not rewrite the source after net-zero edits via --sql-file --save --force](#scenario-does-not-rewrite-the-source-after-net-zero-edits-via---sql-file---save---force)
  - [still persists a genuine CSV change with .save --force](#scenario-still-persists-a-genuine-csv-change-with-save---force)
- [sqly v0.25.0 binary regressions](#sqly-v0250-binary-regressions) — 15 scenarios
  - [rejects an explicit empty --sql value](#scenario-rejects-an-explicit-empty---sql-value)
  - [reports a hint when non-interactive run gets empty stdin and no file](#scenario-reports-a-hint-when-non-interactive-run-gets-empty-stdin-and-no-file)
  - [reports a hint when non-interactive run gets empty stdin with a file](#scenario-reports-a-hint-when-non-interactive-run-gets-empty-stdin-with-a-file)
  - [reports a stable stdin reference instead of the staging temp path](#scenario-reports-a-stable-stdin-reference-instead-of-the-staging-temp-path)
  - [fails batch mode when .schema is missing its table name](#scenario-fails-batch-mode-when-schema-is-missing-its-table-name)
  - [fails batch mode when .header is missing its table name](#scenario-fails-batch-mode-when-header-is-missing-its-table-name)
  - [fails batch mode when .describe is missing its table name](#scenario-fails-batch-mode-when-describe-is-missing-its-table-name)
  - [fails batch mode when .mode is missing its mode name](#scenario-fails-batch-mode-when-mode-is-missing-its-mode-name)
  - [fails batch mode when .dump is missing its destination](#scenario-fails-batch-mode-when-dump-is-missing-its-destination)
  - [fails batch mode when .import is missing its path](#scenario-fails-batch-mode-when-import-is-missing-its-path)
  - [fails batch mode when .save is missing its argument](#scenario-fails-batch-mode-when-save-is-missing-its-argument)
  - [keeps --inspect quiet on stderr after a successful directory import](#scenario-keeps---inspect-quiet-on-stderr-after-a-successful-directory-import)
  - [keeps --profile quiet on stderr after a successful directory import](#scenario-keeps---profile-quiet-on-stderr-after-a-successful-directory-import)
  - [guides "sqly help" to --help instead of an import error](#scenario-guides-sqly-help-to---help-instead-of-an-import-error)
  - [guides "sqly version" to --version instead of an import error](#scenario-guides-sqly-version-to---version-instead-of-an-import-error)
## sqly ACH/Fedwire native write-back
Source: `test/e2e/tools/sqly/ach_fedwire_writeback.atago.yaml`
### Scenario: round-trips an ACH file through --save --force after an UPDATE
#### Given
- Fixture file `payment.ach` is created.
#### When
```shell
sqly --sql "UPDATE payment_entries SET individual_name='E2E Receiver' WHERE entry_index=0" --save --force payment.ach
sqly --json --sql "SELECT individual_name FROM payment_entries WHERE entry_index=0" payment.ach
```
#### Then
- after `sqly --sql "UPDATE payment_entries SET individual_name='E2E Receiver' WHERE entry_index=0" --save --force payment.ach`:
  - stderr contains `Saved`
- after `sqly --json --sql "SELECT individual_name FROM payment_entries WHERE entry_index=0" payment.ach`:
  - exit code is `0`
  - stdout contains `E2E Receiver`
### Scenario: writes a reconstructed ACH file into a directory with --save-dir
#### Given
- Fixture file `payment.ach` is created.
#### When
```shell
sqly --sql "UPDATE payment_entries SET individual_name='Dir Receiver' WHERE entry_index=0" --save-dir out payment.ach
sqly --json --sql "SELECT individual_name FROM payment_entries WHERE entry_index=0" out/payment.ach
```
#### Then
- after `sqly --sql "UPDATE payment_entries SET individual_name='Dir Receiver' WHERE entry_index=0" --save-dir out payment.ach`:
  - stderr contains `Saved`
- after `sqly --json --sql "SELECT individual_name FROM payment_entries WHERE entry_index=0" out/payment.ach`:
  - exit code is `0`
  - stdout contains `Dir Receiver`
### Scenario: round-trips a Fedwire file through --save --force after an UPDATE
#### Given
- Fixture file `transfer.fed` is created.
#### When
```shell
sqly --sql "UPDATE transfer_message SET sender_reference='E2EREF'" --save --force transfer.fed
sqly --json --sql "SELECT sender_reference FROM transfer_message" transfer.fed
```
#### Then
- after `sqly --sql "UPDATE transfer_message SET sender_reference='E2EREF'" --save --force transfer.fed`:
  - stderr contains `Saved`
- after `sqly --json --sql "SELECT sender_reference FROM transfer_message" transfer.fed`:
  - exit code is `0`
  - stdout contains `E2EREF`
### Scenario: still rejects a single-table --output to .ach
#### Given
- Fixture file `payment.ach` is created.
#### When
```shell
sqly --sql "SELECT * FROM payment_entries" --output out.ach payment.ach
```
#### Then
- exit code is `1`
- stderr contains `input-only`
## sqly batch mode (piped stdin)
Source: `test/e2e/tools/sqly/batch.atago.yaml`
### Scenario: runs SQL read from stdin
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT user_name FROM user ORDER BY identifier LIMIT 1
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `booker12`
### Scenario: switches output mode and runs the following query
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.mode ndjson
SELECT user_name FROM user ORDER BY identifier LIMIT 1
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `{"user_name":"booker12"}`
- stderr contains `Change output mode`
### Scenario: exits non-zero and names the failing statement and its line on error
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT user_name FROM user ORDER BY identifier LIMIT 1;
SELECT * FROM no_such_table;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stdout contains `booker12`
- stderr contains `batch statement 2 failed at line 2`, `no_such_table`
### Scenario: reports the line span of a failing multiline statement
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT user_name FROM user ORDER BY identifier LIMIT 1;
SELECT 1;
SELECT *
FROM no_such_table;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `batch statement 3 failed at lines 3-4`, `no_such_table`
### Scenario: stops at .exit with a success status
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.exit
SELECT * FROM no_such_table
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
### Scenario: still exits non-zero when a failure precedes .exit
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT * FROM no_such_table;
.exit
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `no_such_table`
### Scenario: runs a multiline SELECT terminated by a semicolon
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT user_name
FROM user
ORDER BY identifier
LIMIT 1;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `booker12`
### Scenario: runs a multiline UNION ALL across bare newlines as one statement
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.mode csv
SELECT 1 AS n
UNION ALL
SELECT 2;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
- stdout equals an exact value
- stderr contains `Change output mode`
### Scenario: runs a multiline WITH (CTE) query
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
WITH x AS (
  SELECT user_name FROM user ORDER BY identifier LIMIT 1
)
SELECT * FROM x;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `booker12`
### Scenario: runs multiple statements and helper commands in one payload
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.tables
SELECT COUNT(*) AS c FROM user;
SELECT user_name FROM user ORDER BY identifier LIMIT 1;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `TABLE NAME`, `booker12`
### Scenario: ignores a semicolon inside a leading line comment
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
-- comment ;
SELECT 'v' AS x;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `v`
### Scenario: ignores a semicolon inside a block comment
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
/* comment ; */
SELECT 'v' AS x;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `v`
### Scenario: ignores a semicolon inside a trailing line comment
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT 'first' AS x; -- trailing ; comment
SELECT 'second' AS y;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `first`, `second`
### Scenario: does not split on a semicolon inside a bracket-quoted identifier
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT 'v' AS [a;b];
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `a;b`, `v`
### Scenario: does not split on a semicolon inside a backtick-quoted identifier
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT 'v' AS `a;b`;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `a;b`, `v`
### Scenario: reports an error for incomplete SQL
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT * FROM (
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `batch statement`
## sqly batch fail-fast
Source: `test/e2e/tools/sqly/batch_failfast.atago.yaml`
### Scenario: stops later statements after a SQL failure
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
_stdin for `sqly`:_
```text
SELECT * FROM no_such_table;
SELECT 1 AS later;
```
#### When
```shell
sqly u.csv
```
#### Then
- exit code is `1`
- stdout does not contain `later`
- stderr contains `no_such_table`
### Scenario: does not run .save --force after an earlier failure
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
_stdin for `sqly`:_
```text
UPDATE u SET first_name = 'BROKEN' WHERE identifier = 1;
SELECT * FROM no_such_table;
.save --force
```
#### When
```shell
sqly u.csv
```
#### Then
- exit code is `1`
- stdout contains `affected`
- stderr does not contain `Saved`
- file `u.csv` is checked
### Scenario: does not run .dump after an earlier failure
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
_stdin for `sqly`:_
```text
SELECT * FROM no_such_table;
.dump u out.csv
```
#### When
```shell
sqly u.csv
```
#### Then
- exit code is `1`
- stderr contains `no_such_table`
- file `out.csv` does not exist
### Scenario: does not write back for empty stdin with --save --force
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
#### When
```shell
sqly u.csv --save --force
```
#### Then
- exit code is `1`
- stderr contains `no TTY detected`
- stderr does not contain `Saved`
- file `u.csv` contains `Rachel`
## sqly --cache import cache
Source: `test/e2e/tools/sqly/cache.atago.yaml`
### Scenario: writes a cache on the cold run and reuses it on the warm run
#### Given
- Fixture file `data.csv` is created.
#### Inputs
_Fixture `data.csv`:_
```text
id,name
1,Alice
2,Bob
3,Carol
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv
```
#### Then
- after `sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv`:
  - exit code is `0`
  - stdout contains `3`
  - stderr contains `cache: reused`
### Scenario: invalidates the cache when the source changes
#### Given
- Fixture file `data.csv` is created.
- Fixture file `data.csv` is created.
#### Inputs
_Fixture `data.csv`:_
```text
id,name
1,Alice
2,Bob
3,Carol
```
_Fixture `data.csv`:_
```text
id,name
1,Alice
2,Bob
3,Carol
4,Dave
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv
```
#### Then
- after `sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv`:
  - exit code is `0`
  - stdout contains `4`
  - stderr does not contain `cache: reused`
### Scenario: rebuilds after --cache-clear
#### Given
- Fixture file `data.csv` is created.
#### Inputs
_Fixture `data.csv`:_
```text
id,name
1,Alice
2,Bob
3,Carol
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv
sqly --cache snap.cache --cache-clear --sql "SELECT COUNT(*) AS n FROM data" data.csv
```
#### Then
- after `sqly --cache snap.cache --cache-clear --sql "SELECT COUNT(*) AS n FROM data" data.csv`:
  - exit code is `0`
  - stdout contains `3`
  - stderr does not contain `cache: reused`
### Scenario: falls back to a cold import when the cache path is unwritable
#### Given
- Fixture file `data.csv` is created.
- Fixture file `snap.cache/keep` is created.
#### Inputs
_Fixture `data.csv`:_
```text
id,name
1,Alice
2,Bob
3,Carol
```
_Fixture `snap.cache/keep`:_
```text
x
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" data.csv
```
#### Then
- exit code is `0`
- stdout contains `3`
- stderr contains `cache`
### Scenario: invalidates the cache when content changes but size and mtime do not (#592)
#### Given
- Fixture file `d.csv` is created.
- Fixture file `d.csv` is created.
#### Inputs
_Fixture `d.csv`:_
```text
id,name
1,Alice
2,Bob
```
_Fixture `d.csv`:_
```text
id,name
1,Carol
2,Eve
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT group_concat(name, ',') AS names FROM d" d.csv
sqly --cache snap.cache --sql "SELECT group_concat(name, ',') AS names FROM d" d.csv
```
#### Then
- after `sqly --cache snap.cache --sql "SELECT group_concat(name, ',') AS names FROM d" d.csv`:
  - exit code is `0`
  - stdout contains `Carol,Eve`
  - stdout does not contain `Alice,Bob`
  - stderr does not contain `cache: reused`
### Scenario: reuses the cache that lives inside the imported directory and ignores its manifest
#### Given
- Fixture file `indir/data.csv` is created.
#### Inputs
_Fixture `indir/data.csv`:_
```text
id,name
1,Alice
2,Bob
3,Carol
```
#### When
```shell
sqly --cache indir/snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir
sqly --cache indir/snap.cache --sql "SELECT group_concat(name, ',') AS t FROM sqlite_master WHERE type='table'" indir
```
#### Then
- after `sqly --cache indir/snap.cache --sql "SELECT group_concat(name, ',') AS t FROM sqlite_master WHERE type='table'" indir`:
  - exit code is `0`
  - stdout contains `data`
  - stderr contains `cache: reused`
  - stdout does not contain `manifest`, `snap`
### Scenario: keeps the cache warm when only an unsupported sibling file changes
#### Given
- Fixture file `indir/data.csv` is created.
- Fixture file `indir/ignore.txt` is created.
- Fixture file `indir/ignore.txt` is created.
#### Inputs
_Fixture `indir/data.csv`:_
```text
id,name
1,Alice
```
_Fixture `indir/ignore.txt`:_
```text
note
```
_Fixture `indir/ignore.txt`:_
```text
changed
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir
```
#### Then
- after `sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir`:
  - exit code is `0`
  - stdout contains `1`
  - stderr contains `cache: reused`
### Scenario: still invalidates the cache when the supported file changes
#### Given
- Fixture file `indir/data.csv` is created.
- Fixture file `indir/ignore.txt` is created.
- Fixture file `indir/data.csv` is created.
#### Inputs
_Fixture `indir/data.csv`:_
```text
id,name
1,Alice
```
_Fixture `indir/ignore.txt`:_
```text
note
```
_Fixture `indir/data.csv`:_
```text
id,name
1,Alice
2,Bob
```
#### When
```shell
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir
sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir
```
#### Then
- after `sqly --cache snap.cache --sql "SELECT COUNT(*) AS n FROM data" indir`:
  - exit code is `0`
  - stdout contains `2`
  - stderr does not contain `cache: reused`
## sqly changes delta cross-checks (exhaustive-set semantics)
Source: `test/e2e/tools/sqly/changes_crosscheck.atago.yaml`
### Scenario: omitting sqly's history DB from an exhaustive list is rejected
_skipped on windows_
#### Given
- Fixture file `inner.atago.yaml` is created.
#### Inputs
_Fixture `inner.atago.yaml`:_
```text
version: "1"
suite:
  name: inner
scenarios:
  - name: under-specified created list
    steps:
      - fixture:
          file: user.csv
          content: |
            n,v
            a,1
      - run:
          sandbox_home: true
          command: sqly --sql "SELECT 1" user.csv
      - assert:
          changes:
            created: []
```
#### When
```shell
${atago} run inner.atago.yaml
```
#### Then
- exit code is `1`
- stdout contains `unexpected created file`
- stdout contains `.atago-home/.config/sqly/history.db`
### Scenario: the exhaustive list including the history DB passes
_skipped on windows_
#### Given
- Fixture file `user.csv` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `user.csv`:_
```text
n,v
a,1
```
#### When
```shell
sqly --sql "SELECT 1" user.csv
```
#### Then
- exit code is `0`
- the step changed exactly created `.atago-home/.config/sqly/history.db`, modified nothing, deleted nothing
## sqly CLI surface
Source: `test/e2e/tools/sqly/cli.atago.yaml`
### Scenario: prints the version
#### When
```shell
sqly --version
```
#### Then
- exit code is `0`
- stdout contains `sqly`
### Scenario: prints usage with --help
#### When
```shell
sqly --help
```
#### Then
- exit code is `0`
- stdout contains `[Usage]`, `--json`
### Scenario: fails on a non-existent file
#### When
```shell
sqly --sql "SELECT 1" does_not_exist.csv
```
#### Then
- exit code is `1`
- stderr contains `does not exist`
### Scenario: fails on invalid SQL with --sql
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELEKT * FROM user" user.csv
```
#### Then
- exit code is `1`
- stderr contains `syntax error`
### Scenario: reports an unknown flag as a CLI error
#### When
```shell
sqly --no-such-flag
```
#### Then
- exit code is `1`
- stderr contains `unknown flag`
### Scenario: reports conflicting output mode flags as a CLI error
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --csv --json --sql "SELECT 1" user.csv
```
#### Then
- exit code is `1`
- stderr contains `conflicting output mode flags`
### Scenario: joins two imported files
#### Given
- Fixture file `user.csv` is created.
- Fixture file `identifier.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
```
#### When
```shell
sqly --csv --sql "SELECT user_name, position FROM user INNER JOIN identifier ON user.identifier = identifier.id ORDER BY user.identifier LIMIT 1" user.csv identifier.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `booker12`
### Scenario: writes JSON results to the given path with --output
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --json --output result.json --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv
```
#### Then
- exit code is `0`
- stderr contains `result.json`
- file `result.json` contains `"user_name":"booker12"`
### Scenario: applies an output-mode flag placed after the file path
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv --csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `booker12`
### Scenario: fails fast on an unknown flag after the file path
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly user.csv --nope
```
#### Then
- exit code is `1`
- stderr contains `unknown flag`
## sqly table-name collision
Source: `test/e2e/tools/sqly/collision.atago.yaml`
### Scenario: fails when two inputs sanitize to the same table name
#### Given
- Fixture file `a-b.csv` is created.
- Fixture file `a_b.csv` is created.
#### Inputs
_Fixture `a-b.csv`:_
```text
id,name
1,A
```
_Fixture `a_b.csv`:_
```text
id,name
2,B
```
#### When
```shell
sqly --inspect a-b.csv a_b.csv
```
#### Then
- exit code is `1`
- stderr contains `collision`
## sqly --compare workflow
Source: `test/e2e/tools/sqly/compare.atago.yaml`
### Scenario: reports schema, row count, and keyed rows as JSON
#### Given
- Fixture file `rev1.csv` is created.
- Fixture file `rev2.csv` is created.
#### Inputs
_Fixture `rev1.csv`:_
```text
id,name,age
1,Alice,30
2,Bob,25
3,Carol,40
```
_Fixture `rev2.csv`:_
```text
id,name,age
1,Alice,31
2,Bob,25
4,Dave,50
```
#### When
```shell
sqly --compare --compare-key id rev1.csv rev2.csv
```
#### Then
- exit code is `0`
- stdout at `$.schema.equal` equals `true`
- stdout at `$.row_count.delta` equals `0`
- stdout at `$.rows.key` equals `id`
- stdout contains `"4"`, `"3"`
### Scenario: emits a human-readable summary with --compare-format text
#### Given
- Fixture file `rev1.csv` is created.
- Fixture file `rev2.csv` is created.
#### Inputs
_Fixture `rev1.csv`:_
```text
id,name,age
1,Alice,30
2,Bob,25
3,Carol,40
```
_Fixture `rev2.csv`:_
```text
id,name,age
1,Alice,31
2,Bob,25
4,Dave,50
```
#### When
```shell
sqly --compare --compare-key id --compare-format text rev1.csv rev2.csv
```
#### Then
- exit code is `0`
- stdout contains `schema: identical`, `1 added, 1 removed, 1 modified`
### Scenario: resolves an uppercase --compare-key against a lowercase column
#### Given
- Fixture file `rev1.csv` is created.
- Fixture file `rev2.csv` is created.
#### Inputs
_Fixture `rev1.csv`:_
```text
id,name,age
1,Alice,30
2,Bob,25
3,Carol,40
```
_Fixture `rev2.csv`:_
```text
id,name,age
1,Alice,31
2,Bob,25
4,Dave,50
```
#### When
```shell
sqly --compare --compare-key ID rev1.csv rev2.csv
```
#### Then
- exit code is `0`
- stdout at `$.rows.key` equals `ID`
### Scenario: rejects a missing key column
#### Given
- Fixture file `rev1.csv` is created.
- Fixture file `rev2.csv` is created.
#### Inputs
_Fixture `rev1.csv`:_
```text
id,name,age
1,Alice,30
2,Bob,25
3,Carol,40
```
_Fixture `rev2.csv`:_
```text
id,name,age
1,Alice,31
2,Bob,25
4,Dave,50
```
#### When
```shell
sqly --compare --compare-key nope rev1.csv rev2.csv
```
#### Then
- exit code is `1`
- stderr contains `compare key`
### Scenario: rejects a duplicate key value as ambiguous
#### Given
- Fixture file `dupe.csv` is created.
- Fixture file `single.csv` is created.
#### Inputs
_Fixture `dupe.csv`:_
```text
id,name
1,Alice
1,Bob
```
_Fixture `single.csv`:_
```text
id,name
1,Alice
```
#### When
```shell
sqly --compare --compare-key id dupe.csv single.csv
```
#### Then
- exit code is `1`
- stderr contains `not unique`
### Scenario: rejects a genuinely missing --compare-tables name
#### Given
- Fixture file `rev1.csv` is created.
- Fixture file `rev2.csv` is created.
#### Inputs
_Fixture `rev1.csv`:_
```text
id,name,age
1,Alice,30
2,Bob,25
3,Carol,40
```
_Fixture `rev2.csv`:_
```text
id,name,age
1,Alice,31
2,Bob,25
4,Dave,50
```
#### When
```shell
sqly --compare --compare-tables "nope,rev2" rev1.csv rev2.csv
```
#### Then
- exit code is `1`
- stderr contains `compare table`
### Scenario: rejects an ambiguous single-table compare
#### Given
- Fixture file `rev1.csv` is created.
#### Inputs
_Fixture `rev1.csv`:_
```text
id,name,age
1,Alice,30
2,Bob,25
3,Carol,40
```
#### When
```shell
sqly --compare rev1.csv
```
#### Then
- exit code is `1`
- stderr contains `exactly two tables`
### Scenario: follows CLI input order for left and right, not table-name sorting
#### Given
- Fixture file `zebra.csv` is created.
- Fixture file `ant.csv` is created.
#### Inputs
_Fixture `zebra.csv`:_
```text
id,name
1,Alice
```
_Fixture `ant.csv`:_
```text
id,name
1,Alice
```
#### When
```shell
sqly --compare --compare-format text zebra.csv ant.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: reverses left and right when the inputs are swapped
#### Given
- Fixture file `zebra.csv` is created.
- Fixture file `ant.csv` is created.
#### Inputs
_Fixture `zebra.csv`:_
```text
id,name
1,Alice
```
_Fixture `ant.csv`:_
```text
id,name
1,Alice
```
#### When
```shell
sqly --compare --compare-format text ant.csv zebra.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: keeps the keyed diff correct on a larger input
#### Given
- Fixture file `big1.csv` is created.
- Fixture file `big2.csv` is created.
#### Inputs
_Fixture `big1.csv`:_
```text
id,name,score
0,name0,0
1,name1,1
2,name2,2
3,name3,3
4,name4,4
5,name5,5
6,name6,6
7,name7,7
8,name8,8
9,name9,9
10,name10,10
11,name11,11
12,name12,12
13,name13,13
14,name14,14
15,name15,15
16,name16,16
17,name17,17
18,name18,18
… (truncated, 31 more lines)
```
_Fixture `big2.csv`:_
```text
id,name,score
0,name0,1
1,name1,1
2,name2,2
3,name3,3
4,name4,4
5,name5,5
6,name6,6
7,name7,7
8,name8,8
9,name9,9
10,name10,10
11,name11,11
12,name12,12
13,name13,13
14,name14,14
15,name15,15
16,name16,16
17,name17,17
18,name18,18
… (truncated, 31 more lines)
```
#### When
```shell
sqly --compare --compare-key id --compare-format text big1.csv big2.csv
```
#### Then
- exit code is `0`
- stdout contains `1 added, 1 removed, 1 modified`
## sqly cross-format workflows
Source: `test/e2e/tools/sqly/cross_format.atago.yaml`
### Scenario: joins a Parquet table to a CSV table and computes a column
#### Given
- Fixture file `products.parquet` is created.
- Fixture file `sales.csv` is created.
#### Inputs
_Fixture `sales.csv`:_
```text
product_id,quantity
1,3
2,10
3,5
```
#### When
```shell
sqly --csv --sql "SELECT p.name, p.price, s.quantity, ROUND(p.price * s.quantity, 2) AS revenue FROM products p JOIN sales s ON p.id = s.product_id ORDER BY revenue DESC" products.parquet sales.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `Laptop`, `2999.97`
## sqly empty command arguments
Source: `test/e2e/tools/sqly/empty_args.atago.yaml`
### Scenario: rejects .save "" and leaves the source unchanged
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
```
_stdin for `sqly`:_
```text
UPDATE u SET first_name = 'EMPTY' WHERE identifier = 1;
.save ""
```
#### When
```shell
sqly u.csv
```
#### Then
- exit code is `1`
- stdout contains `affected`
- stderr contains `.save requires`
- file `u.csv` is checked
### Scenario: rejects .dump with an empty destination
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump user ""
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.dump requires`
### Scenario: rejects .import with an empty path
#### Inputs
_stdin for `sqly`:_
```text
.import ""
.tables
```
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stderr contains `empty import path`
## sqly excel export
Source: `test/e2e/tools/sqly/excel_export.atago.yaml`
### Scenario: writes a non-executable .xlsx with --output
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --excel --output out.xlsx --sql "SELECT * FROM user LIMIT 1" user.csv
```
#### Then
- exit code is `0`
- stderr contains `output mode=excel`
- file `out.xlsx` exists
- file `out.xlsx` is checked
#### Generated artifacts
- `out.xlsx`
### Scenario: writes a non-executable .xlsx with the .dump command
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode excel
.dump user dump.xlsx
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stderr contains `mode=excel`
- file `dump.xlsx` exists
- file `dump.xlsx` is checked
#### Generated artifacts
- `dump.xlsx`
## sqly export format inference
Source: `test/e2e/tools/sqly/export_inference.atago.yaml`
### Scenario: infers parquet from the destination extension without a flag
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv --output result.parquet
```
#### Then
- exit code is `0`
- stderr contains `output mode=parquet`
- file `result.parquet` exists
#### Generated artifacts
- `result.parquet`
### Scenario: infers ndjson with gzip and writes a compressed file
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv --output result.ndjson.gz
```
#### Then
- exit code is `0`
- stderr contains `output mode=ndjson`
- file `result.ndjson.gz` exists
- file `result.ndjson.gz` is checked
#### Generated artifacts
- `result.ndjson.gz`
### Scenario: re-imports a gzip-compressed csv it wrote
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
#### When
```shell
sqly --csv --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv --output result.csv.gz
sqly --csv --sql "SELECT user_name FROM result LIMIT 1" result.csv.gz
```
#### Then
- after `sqly --csv --sql "SELECT user_name FROM result LIMIT 1" result.csv.gz`:
  - exit code is `0`
  - stdout equals an exact value
  - stdout equals an exact value
### Scenario: writes the CSV fallback to an unknown extension path without rewriting it
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT user_name FROM user LIMIT 1" user.csv --output out.unknown
```
#### Then
- exit code is `0`
- stderr contains `out.unknown`
- file `out.unknown` exists
- file `out.csv` does not exist
#### Generated artifacts
- `out.unknown`
### Scenario: rejects --output to an existing directory
#### Given
- Fixture file `user.csv` is created.
- Fixture file `outdir/.keep` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT identifier FROM user LIMIT 1" user.csv --output outdir
```
#### Then
- exit code is `1`
- stderr contains `directory`
- file `outdir.csv` does not exist
### Scenario: rejects .dump to an existing directory
#### Given
- Fixture file `user.csv` is created.
- Fixture file `outdir/.keep` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump user outdir
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `directory`
### Scenario: errors when an explicit mode flag disagrees with the path extension
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --json --sql "SELECT user_name FROM user LIMIT 1" user.csv --output result.csv
```
#### Then
- exit code is `1`
- stderr contains `conflicts with destination path`
### Scenario: rejects bzip2 output
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT user_name FROM user LIMIT 1" user.csv --output result.csv.bz2
```
#### Then
- exit code is `1`
- stderr contains `bzip2`
### Scenario: rejects compression on parquet
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT user_name FROM user LIMIT 1" user.csv --output result.parquet.gz
```
#### Then
- exit code is `1`
- stderr contains `cannot be compressed`
### Scenario: infers tsv from the .dump destination path
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump user dump.tsv
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stderr contains `mode=tsv`
- file `dump.tsv` exists
#### Generated artifacts
- `dump.tsv`
### Scenario: keeps --json --output result.json writing json
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
#### When
```shell
sqly --json --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv --output result.json
```
#### Then
- exit code is `0`
- stderr contains `output mode=json`
- file `result.json` contains `"user_name":"booker12"`
## sqly helper commands reject extra args
Source: `test/e2e/tools/sqly/extra_args.atago.yaml`
### Scenario: rejects .schema with an extra argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.schema user extra
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.schema`
### Scenario: rejects .describe with an extra argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.describe user extra
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.describe`
### Scenario: rejects .tables with an extra argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.tables extra
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.tables`
### Scenario: rejects .mode with an extra argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode csv extra
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.mode`
### Scenario: rejects .pwd with an extra argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.pwd extra
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.pwd`
### Scenario: rejects .clear with an extra argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.clear extra
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.clear`
### Scenario: does not let .exit with an extra argument silently terminate the batch
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.exit extra
SELECT 1;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.exit`
## sqly filesql integration
Source: `test/e2e/tools/sqly/filesql_integration.atago.yaml`
### Scenario: imports and queries a CSV file
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
SELECT COUNT(*) FROM user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `3`
### Scenario: imports and queries a JSONL file
#### Given
- Fixture file `sample.jsonl` is created.
#### Inputs
_stdin for `sqly`:_
```text
SELECT COUNT(*) FROM sample
```
#### When
```shell
sqly sample.jsonl
```
#### Then
- exit code is `0`
- stdout contains `COUNT(*)`
### Scenario: imports and queries a Parquet file
#### Given
- Fixture file `products.parquet` is created.
#### Inputs
_stdin for `sqly`:_
```text
SELECT COUNT(*) FROM products
```
#### When
```shell
sqly products.parquet
```
#### Then
- exit code is `0`
- stdout contains `3`
### Scenario: imports and queries an Excel file
#### Given
- Fixture file `sample.xlsx` is created.
#### Inputs
_stdin for `sqly`:_
```text
SELECT COUNT(*) FROM sample_test_sheet
```
#### When
```shell
sqly sample.xlsx
```
#### Then
- exit code is `0`
- stdout contains `COUNT(*)`
### Scenario: imports and queries an ACH file
#### Given
- Fixture file `ppd-debit.ach` is created.
#### Inputs
_stdin for `sqly`:_
```text
SELECT COUNT(*) FROM ppd_debit_entries
```
#### When
```shell
sqly ppd-debit.ach
```
#### Then
- exit code is `0`
- stdout contains `COUNT(*)`
### Scenario: imports and queries a Fedwire file
#### Given
- Fixture file `customer-transfer.fed` is created.
#### Inputs
_stdin for `sqly`:_
```text
SELECT COUNT(*) FROM customer_transfer_message
```
#### When
```shell
sqly customer-transfer.fed
```
#### Then
- exit code is `0`
- stdout contains `COUNT(*)`
### Scenario: preserves filesql-detected column types for schema inspection
#### Given
- Fixture file `products.parquet` is created.
#### Inputs
_stdin for `sqly`:_
```text
.describe products
```
#### When
```shell
sqly products.parquet
```
#### Then
- exit code is `0`
- stdout contains `name`, `price`
### Scenario: produces identical tables on repeated ACH imports
#### Given
- Fixture file `ppd-debit.ach` is created.
#### Inputs
_stdin for `sqly`:_
```text
.tables
```
_stdin for `sqly`:_
```text
.tables
```
#### When
```shell
sqly ppd-debit.ach
sqly ppd-debit.ach
```
#### Then
- after `sqly ppd-debit.ach`:
  - stdout contains `ppd_debit_entries`
- after `sqly ppd-debit.ach`:
  - stdout contains `ppd_debit_entries`
## sqly shell helper commands
Source: `test/e2e/tools/sqly/helpers.atago.yaml`
### Scenario: .help groups commands, shows usage, and flags destructive save
#### Inputs
_stdin for `sqly`:_
```text
.help
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `Import / Export`, `.import PATH`, `.dump TABLE FILE`, `.save DIR`, `.save --force`, `destructive`
### Scenario: .cd changes directory with a relative path and reports it
#### Given
- Fixture file `testdata/x.csv` is created.
#### Inputs
_Fixture `testdata/x.csv`:_
```text
id,name
1,a
```
_stdin for `sqly`:_
```text
.cd testdata
.pwd
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `testdata`
### Scenario: expands a bare ~ in .cd to the home directory
#### Given
- Fixture file `home/sqly_tilde.csv` is created.
#### Inputs
_Fixture `home/sqly_tilde.csv`:_
```text
id,name
1,foo
```
_stdin for `sqly`:_
```text
.cd ~
.pwd
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `home`
### Scenario: expands ~/file in .import
#### Given
- Fixture file `home/sqly_tilde.csv` is created.
#### Inputs
_Fixture `home/sqly_tilde.csv`:_
```text
id,name
1,foo
```
_stdin for `sqly`:_
```text
.import ~/sqly_tilde.csv
.tables
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `sqly_tilde`
### Scenario: .clear emits no ANSI escapes to stdout in batch mode
#### Inputs
_stdin for `sqly`:_
```text
.clear
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout is empty
### Scenario: keeps --json stdout parseable when .clear precedes a query
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.clear
SELECT 1 AS x;
```
#### When
```shell
sqly --json user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: .import imports a quoted path containing a space as one argument
#### Given
- Fixture file `sqly_e2e space.csv` is created.
#### Inputs
_Fixture `sqly_e2e space.csv`:_
```text
id,name
1,foo
```
_stdin for `sqly`:_
```text
.import "sqly_e2e space.csv"
.tables
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `sqly_e2e_space`
### Scenario: .mode fails on a missing mode name but still lists the modes
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `json`, `ndjson`
### Scenario: .dump infers TSV from the .tsv extension in table mode, not CSV
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump user out.tsv
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stderr contains `mode=tsv`
- file `out.tsv` contains `user_name	identifier`
- file `out.tsv` is checked
## sqly hermetic environment
Source: `test/e2e/tools/sqly/hermetic.atago.yaml`
### Scenario: runs with HOME inside the sandbox
#### When
```shell
case "$HOME" in "$SQLY_E2E_SANDBOX"*) exit 0 ;; *) exit 1 ;; esac
```
#### Then
- exit code is `0`
### Scenario: routes the history DB into the sandbox
#### When
```shell
case "$SQLY_HISTORY_DB_PATH" in "$SQLY_E2E_SANDBOX"*) exit 0 ;; *) exit 1 ;; esac
```
#### Then
- exit code is `0`
### Scenario: still runs sqly normally inside the sandbox
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT 1 AS one" user.csv
```
#### Then
- exit code is `0`
- stdout contains `one`
## sqly history tolerance
Source: `test/e2e/tools/sqly/history_tolerance.atago.yaml`
### Scenario: runs --sql and warns instead of failing
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
#### When
```shell
sqly --csv --sql "SELECT actor FROM actor ORDER BY actor LIMIT 1" actor.csv
```
#### Then
- exit code is `0`
- stdout contains `Adam Sandler`
- stderr contains `history disabled`
### Scenario: runs batch mode without a writable history DB
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
_stdin for `sqly`:_
```text
.tables
SELECT actor FROM actor ORDER BY actor LIMIT 1
```
#### When
```shell
sqly actor.csv
```
#### Then
- exit code is `0`
- stdout contains `TABLE NAME`, `Adam Sandler`
- stderr contains `history disabled`
### Scenario: runs --sql when the history DB is read-only after startup
#### Given
- Fixture file `user.csv` is created.
- Fixture file `h.db` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
#### When
```shell
sqly --sql "SELECT 1" user.csv
sqly --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv
```
#### Then
- after `sqly --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv`:
  - exit code is `0`
  - stdout contains `booker12`
### Scenario: runs --inspect when the history DB is read-only after startup
#### Given
- Fixture file `user.csv` is created.
- Fixture file `h.db` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
#### When
```shell
sqly --sql "SELECT 1" user.csv
sqly --inspect user.csv
```
#### Then
- after `sqly --inspect user.csv`:
  - exit code is `0`
  - stdout equals an exact value
  - stdout contains `"name": "user"`
### Scenario: runs batch mode and warns when a history write fails after startup
#### Given
- Fixture file `user.csv` is created.
- Fixture file `h.db` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
_stdin for `sqly`:_
```text
SELECT user_name FROM user ORDER BY identifier LIMIT 1
```
#### When
```shell
sqly --sql "SELECT 1" user.csv
sqly user.csv
```
#### Then
- after `sqly user.csv`:
  - exit code is `0`
  - stdout contains `booker12`
  - stderr contains `history disabled`
## sqly import failure handling
Source: `test/e2e/tools/sqly/import_failure.atago.yaml`
### Scenario: fails query mode on a partial import and keeps stdout clean
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --json --sql "SELECT user_name FROM user LIMIT 1" user.csv /no/such/file.csv
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `failed to import`, `inputs failed to import: path does not exist`
### Scenario: names the failed count and first path when every input fails
#### When
```shell
sqly --sql "SELECT 1" /no/such/a.csv /no/such/b.csv
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `all 2 import(s) failed`, `/no/such/a.csv`, `+1 more`
### Scenario: fails --inspect on a partial import
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect user.csv /no/such/file.csv
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `failed to import`
### Scenario: fails batch .import on a partial import and stops later commands
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.import user.csv /no/such/file.csv
.tables
```
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stdout does not contain `TABLE NAME`
- stderr contains `failed to import`
### Scenario: keeps stdout clean when a stdin dataset fails to import
#### When
```shell
sqly --stdin csv --json --sql "SELECT COUNT(*) FROM stdin"
```
#### Then
- exit code is `1`
- stdout is empty
- stderr contains `Import failed`
## sqly .import with space-containing paths
Source: `test/e2e/tools/sqly/import_quoting.atago.yaml`
### Scenario: imports a backslash-escaped space path as a single argument
#### Given
- Fixture file `space name.csv` is created.
#### Inputs
_Fixture `space name.csv`:_
```text
label,score
alpha,1
beta,2
```
_stdin for `sqly`:_
```text
.import space\ name.csv
SELECT label FROM space_name ORDER BY score;
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `alpha`, `beta`
### Scenario: imports a double-quoted space path as a single argument
#### Given
- Fixture file `space name.csv` is created.
#### Inputs
_Fixture `space name.csv`:_
```text
label,score
alpha,1
beta,2
```
_stdin for `sqly`:_
```text
.import "space name.csv"
SELECT label FROM space_name ORDER BY score;
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `alpha`
### Scenario: imports a single-quoted space path as a single argument
#### Given
- Fixture file `space name.csv` is created.
#### Inputs
_Fixture `space name.csv`:_
```text
label,score
alpha,1
beta,2
```
_stdin for `sqly`:_
```text
.import 'space name.csv'
SELECT label FROM space_name ORDER BY score;
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `alpha`
### Scenario: imports a file inside a space-containing directory when escaped
#### Given
- Fixture file `space dir/nested.csv` is created.
#### Inputs
_Fixture `space dir/nested.csv`:_
```text
label,score
gamma,3
delta,4
```
_stdin for `sqly`:_
```text
.import space\ dir/nested.csv
SELECT label FROM nested ORDER BY score;
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `gamma`, `delta`
### Scenario: splits an unquoted space path into two failing arguments
#### Given
- Fixture file `space name.csv` is created.
#### Inputs
_Fixture `space name.csv`:_
```text
label,score
alpha,1
beta,2
```
_stdin for `sqly`:_
```text
.import space name.csv
```
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stderr contains `space`
## sqly --inspect
Source: `test/e2e/tools/sqly/inspect.atago.yaml`
### Scenario: prints a JSON report for a single file
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --inspect user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `"name": "user"`, `"row_count": 3`, `"user_name"`
### Scenario: maps every table from a multi-table ACH file to its source
#### Given
- Fixture file `ppd-debit.ach` is created.
#### When
```shell
sqly --inspect ppd-debit.ach
```
#### Then
- exit code is `0`
- stdout contains `"name": "ppd_debit_entries"`, `ppd-debit.ach`
### Scenario: keeps stdout as pure JSON for a directory and stays quiet on stderr
#### Given
- Fixture file `ins/a.csv` is created.
- Fixture file `ins/b.csv` is created.
#### Inputs
_Fixture `ins/a.csv`:_
```text
user_name,identifier
booker12,1
```
_Fixture `ins/b.csv`:_
```text
id,position
1,developrt
```
#### When
```shell
sqly --inspect ins
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `"name": "a"`, `"name": "b"`
- stderr does not contain `Successfully imported`
### Scenario: fails with a clear error when no input is given
#### When
```shell
sqly --inspect
```
#### Then
- exit code is `1`
- stderr contains `no tables to inspect`
### Scenario: emits a schema-only report with --inspect-sample 0
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --inspect --inspect-sample 0 user.csv
```
#### Then
- exit code is `0`
- stdout contains `"sample_rows": []`
- stdout does not contain `booker12`
### Scenario: limits sample rows with --inspect-sample
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --inspect --inspect-sample 1 user.csv
```
#### Then
- exit code is `0`
- stdout contains `booker12`
- stdout does not contain `jenkins46`
### Scenario: rejects a negative --inspect-sample
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --inspect --inspect-sample -1 user.csv
```
#### Then
- exit code is `1`
- stderr contains `inspect-sample`
## sqly --inspect conflicts
Source: `test/e2e/tools/sqly/inspect_conflicts.atago.yaml`
### Scenario: rejects --inspect with --sql
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --sql "SELECT * FROM user LIMIT 1" user.csv
```
#### Then
- exit code is `1`
- stderr contains `--inspect`
### Scenario: rejects --inspect with --output and writes no file
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --output out.csv user.csv
```
#### Then
- exit code is `1`
- stderr contains `--inspect`
- file `out.csv` does not exist
### Scenario: rejects --inspect with --save-dir and creates no save dir
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --save-dir save user.csv
```
#### Then
- exit code is `1`
- stderr contains `--inspect`
- file `save` does not exist
## sqly JSON/NDJSON NULL handling
Source: `test/e2e/tools/sqly/json_null.atago.yaml`
### Scenario: emits NULL as JSON null in --json
#### When
```shell
sqly --json --sql "SELECT NULL AS n, '' AS e, 1 AS x"
```
#### Then
- exit code is `0`
- stdout contains `"n":null`, `"e":""`
### Scenario: emits NULL as JSON null in --ndjson
#### When
```shell
sqly --ndjson --sql "SELECT NULL AS n, '' AS e, 1 AS x"
```
#### Then
- exit code is `0`
- stdout contains `"n":null`, `"e":""`
## sqly metamorphic relations
Source: `test/e2e/tools/sqly/metamorphic.atago.yaml`
### Scenario: COUNT(*) equals the number of selected rows
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS c FROM user" user.csv
sqly --csv --sql "SELECT user_name FROM user" user.csv
```
#### Then
- after `sqly --csv --sql "SELECT COUNT(*) AS c FROM user" user.csv`:
  - stdout equals an exact value
- after `sqly --csv --sql "SELECT user_name FROM user" user.csv`:
  - stdout contains `booker12`, `jenkins46`, `grey07`
### Scenario: WHERE 1=1 returns all rows and WHERE 1=0 returns none
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS c FROM user WHERE 1=1" user.csv
sqly --json --sql "SELECT user_name FROM user WHERE 1=0" user.csv
```
#### Then
- after `sqly --csv --sql "SELECT COUNT(*) AS c FROM user WHERE 1=1" user.csv`:
  - stdout equals an exact value
- after `sqly --json --sql "SELECT user_name FROM user WHERE 1=0" user.csv`:
  - stdout equals an exact value
### Scenario: ORDER BY preserves the row multiset
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --csv --sql "SELECT user_name FROM user ORDER BY user_name DESC" user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `grey07`, `booker12`
### Scenario: csv and ndjson yield the same rows
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --ndjson --sql "SELECT user_name FROM user ORDER BY identifier" user.csv
```
#### Then
- stdout equals an exact value
- stdout equals an exact value
### Scenario: CSV dump reimported yields identical data
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.dump user rt.csv
```
#### When
```shell
sqly user.csv
sqly --csv --sql "SELECT * FROM rt ORDER BY identifier" rt.csv
```
#### Then
- after `sqly --csv --sql "SELECT * FROM rt ORDER BY identifier" rt.csv`:
  - exit code is `0`
  - stdout equals an exact value
  - stdout equals an exact value
  - stdout equals an exact value
## sqly .mode banner routing
Source: `test/e2e/tools/sqly/mode_banner.atago.yaml`
### Scenario: keeps stdout pure JSON after .mode json
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode json
SELECT user_name FROM user LIMIT 1
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout does not contain `Change output mode`
- stderr contains `Change output mode from table to json`
### Scenario: keeps stdout pure NDJSON after .mode ndjson
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode ndjson
SELECT user_name FROM user LIMIT 1
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout does not contain `Change output mode`
- stderr contains `Change output mode from table to ndjson`
### Scenario: reports the typed mode by name and emits typed output after .mode json-typed
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode json-typed
SELECT 7 AS n, 'x' AS s
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout does not contain `Change output mode`
- stderr contains `Change output mode from table to json-typed`
- stdout contains `"n":7`, `"s":"x"`
## sqly output formats
Source: `test/e2e/tools/sqly/output_format.atago.yaml`
### Scenario: --json renders results as a JSON array
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --json --sql "SELECT user_name, identifier FROM user ORDER BY identifier LIMIT 2" user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout contains `{"user_name":"booker12","identifier":"1"}`
- stdout contains `{"user_name":"jenkins46","identifier":"2"}`
### Scenario: --json prints [] for an empty result
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --json --sql "SELECT user_name FROM user WHERE user_name = 'nobody'" user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: --ndjson renders one JSON object per line
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --ndjson --sql "SELECT user_name, identifier FROM user ORDER BY identifier LIMIT 2" user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
### Scenario: --ndjson prints nothing for an empty result
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --ndjson --sql "SELECT user_name FROM user WHERE user_name = 'nobody'" user.csv
```
#### Then
- exit code is `0`
- stdout is empty
### Scenario: --csv renders header and rows as CSV
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
#### When
```shell
sqly --csv --sql "SELECT user_name, identifier FROM user ORDER BY identifier LIMIT 1" user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
## sqly --output requires --sql
Source: `test/e2e/tools/sqly/output_requires_sql.atago.yaml`
### Scenario: rejects --output with no query
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly user.csv --output out.csv
```
#### Then
- exit code is `1`
- stderr contains `--output requires --sql`
- file `out.csv` does not exist
### Scenario: rejects --output for batch SQL from stdin
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
SELECT user_name FROM user ORDER BY identifier LIMIT 1
```
#### When
```shell
sqly user.csv --output out.csv
```
#### Then
- exit code is `1`
- stderr contains `--output requires --sql`
- file `out.csv` does not exist
## sqly file-output status routing
Source: `test/e2e/tools/sqly/output_status.atago.yaml`
### Scenario: keeps stdout empty for --output and reports on stderr
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT 1 AS x" --output out.csv user.csv
```
#### Then
- exit code is `0`
- stdout is empty
- stderr contains `Output sql result to`
- file `out.csv` exists
#### Generated artifacts
- `out.csv`
### Scenario: keeps stdout free of the .dump status line
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump user dump.csv
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout is empty
- stderr contains `dump `user` table to`
- file `dump.csv` exists
#### Generated artifacts
- `dump.csv`
### Scenario: keeps the .save confirmation off stdout
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
UPDATE u SET user_name = 'X' WHERE identifier = 1;
.save saved
```
#### When
```shell
sqly u.csv
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stdout does not contain `Saved`
- stderr contains `Saved u to`
## sqly parquet export
Source: `test/e2e/tools/sqly/parquet_export.atago.yaml`
### Scenario: writes a parquet file that re-imports with the same rows
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.mode parquet
.dump user user.parquet
```
#### When
```shell
sqly user.csv
sqly --csv --sql "SELECT COUNT(*) AS c FROM user" user.parquet
```
#### Then
- after `sqly user.csv`:
  - stderr contains `Change output mode`
  - file `user.parquet` exists
- after `sqly --csv --sql "SELECT COUNT(*) AS c FROM user" user.parquet`:
  - exit code is `0`
  - stdout equals an exact value
#### Generated artifacts
- `user.parquet`
### Scenario: appends the .parquet extension
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
_stdin for `sqly`:_
```text
.mode parquet
.dump user result
```
#### When
```shell
sqly user.csv
```
#### Then
- file `result.parquet` exists
#### Generated artifacts
- `result.parquet`
### Scenario: writes query results to the given parquet path with --output
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --parquet --output q.parquet --sql "SELECT user_name FROM user LIMIT 2" user.csv
```
#### Then
- exit code is `0`
- stderr contains `q.parquet`
- file `q.parquet` exists
#### Generated artifacts
- `q.parquet`
### Scenario: preserves leading-zero codes through a parquet round-trip
#### When
```shell
sqly --parquet --output codes.parquet --sql "SELECT '007' AS code, '00042' AS acct"
sqly --csv --sql "SELECT code, acct FROM codes" codes.parquet
```
#### Then
- after `sqly --csv --sql "SELECT code, acct FROM codes" codes.parquet`:
  - exit code is `0`
  - stdout contains `007`, `00042`
### Scenario: preserves SQL NULL through a parquet round-trip
#### When
```shell
sqly --parquet --output n.parquet --sql "SELECT CAST(NULL AS TEXT) AS id, 'A' AS name UNION ALL SELECT '1' AS id, 'B' AS name"
sqly --json-typed --sql "SELECT * FROM n" n.parquet
```
#### Then
- after `sqly --json-typed --sql "SELECT * FROM n" n.parquet`:
  - exit code is `0`
  - stdout contains `"id":null`
  - stdout does not contain `"id":""`
### Scenario: reports a clear error when exporting an empty result
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --parquet --output empty.parquet --sql "SELECT user_name FROM user WHERE 1=0" user.csv
```
#### Then
- exit code is `1`
- stderr contains `empty result`
## sqly input path validation
Source: `test/e2e/tools/sqly/path_validation.atago.yaml`
### Scenario: imports a deeply nested path
#### Given
- Fixture file `a/b/c/d/e/f/g/h/i/j/k/user.csv` is created.
#### Inputs
_Fixture `a/b/c/d/e/f/g/h/i/j/k/user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
grey07,3
```
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS c FROM user" a/b/c/d/e/f/g/h/i/j/k/user.csv
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: imports a file whose name literally contains ..%2f
#### Given
- Fixture file `..%2fuser.csv` is created.
#### Inputs
_Fixture `..%2fuser.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect "..%2fuser.csv"
```
#### Then
- exit code is `0`
- stdout contains `user_name`
- stderr does not contain `dangerous path pattern`
### Scenario: rejects a symlink alias that resolves to a blocked system path
_skipped on windows_
#### When
```shell
ln -s /etc/hosts hosts_alias.csv
sqly --inspect hosts_alias.csv
```
#### Then
- after `sqly --inspect hosts_alias.csv`:
  - exit code is `1`
  - stderr contains `system directory not allowed`
### Scenario: imports a symlink alias that resolves to an ordinary user file
_skipped on windows_
#### Given
- Fixture file `real.csv` is created.
- Fixture file `user_alias.csv` is created.
#### Inputs
_Fixture `real.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect user_alias.csv
```
#### Then
- exit code is `0`
- stdout contains `user_name`
## sqly --profile workflow
Source: `test/e2e/tools/sqly/profile.atago.yaml`
### Scenario: reports per-column data quality as JSON
#### Given
- Fixture file `messy.csv` is created.
#### Inputs
_Fixture `messy.csv`:_
```text
id,score,note
1,10, hi
2,abc,
3,30,N/A
```
#### When
```shell
sqly --profile messy.csv
```
#### Then
- exit code is `0`
- stdout contains `"row_count": 3`, `"column_count": 3`, `mixed numeric and non-numeric`, `null placeholders`
### Scenario: profiles a stdin dataset
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,Alice
2,Bob
```
#### When
```shell
sqly --stdin csv --profile
```
#### Then
- exit code is `0`
- stdout contains `"name": "stdin"`, `"row_count": 2`
### Scenario: profiles multiple tables in one run
#### Given
- Fixture file `messy.csv` is created.
- Fixture file `orders.csv` is created.
#### Inputs
_Fixture `messy.csv`:_
```text
id,score,note
1,10, hi
2,abc,
3,30,N/A
```
_Fixture `orders.csv`:_
```text
oid,amount
1,9.99
2,5.00
```
#### When
```shell
sqly --profile messy.csv orders.csv
```
#### Then
- exit code is `0`
- stdout contains `"name": "messy"`, `"name": "orders"`
### Scenario: emits a human-readable summary with --profile-format text
#### Given
- Fixture file `orders.csv` is created.
#### Inputs
_Fixture `orders.csv`:_
```text
oid,amount
1,9.99
2,5.00
```
#### When
```shell
sqly --profile --profile-format text orders.csv
```
#### Then
- exit code is `0`
- stdout contains `table orders: 2 rows, 2 columns`
### Scenario: counts a blank string as a distinct value in JSON output
#### Given
- Fixture file `blank.csv` is created.
#### Inputs
_Fixture `blank.csv`:_
```text
id,v
x,
x,A
```
#### When
```shell
sqly --profile blank.csv
```
#### Then
- stdout contains `"blank_count": 1`, `"distinct_count": 2`
### Scenario: counts a blank string as a distinct value in text output
#### Given
- Fixture file `blank.csv` is created.
#### Inputs
_Fixture `blank.csv`:_
```text
id,v
x,
x,A
```
#### When
```shell
sqly --profile --profile-format text blank.csv
```
#### Then
- stdout contains `blanks=1 distinct=2`
### Scenario: flags a padded null-like placeholder and its whitespace together
#### Given
- Fixture file `nullspace.csv` is created.
#### Inputs
_Fixture `nullspace.csv`:_
```text
v
" NULL "
```
#### When
```shell
sqly --profile nullspace.csv
```
#### Then
- stdout contains `null placeholders`, `leading or trailing whitespace`
### Scenario: warns only about whitespace for a padded ordinary value
#### Given
- Fixture file `padded.csv` is created.
#### Inputs
_Fixture `padded.csv`:_
```text
v
" hello "
```
#### When
```shell
sqly --profile padded.csv
```
#### Then
- stdout does not contain `null placeholders`
- stdout contains `leading or trailing whitespace`
### Scenario: counts comma-formatted numerals as numeric, matching table-mode
#### Given
- Fixture file `commas.csv` is created.
#### Inputs
_Fixture `commas.csv`:_
```text
amount
"1,000"
"2,500"
```
#### When
```shell
sqly --profile commas.csv
```
#### Then
- stdout contains `"numeric_count": 2`
- stdout does not contain `mixed numeric`
### Scenario: right-aligns the same comma-formatted column in table-mode
#### Given
- Fixture file `commas.csv` is created.
#### Inputs
_Fixture `commas.csv`:_
```text
amount
"1,000"
"2,500"
```
#### When
```shell
sqly --sql "SELECT * FROM commas" commas.csv
```
#### Then
- exit code is `0`
- stdout contains `|  1,000 |`
## README examples
Source: `test/e2e/tools/sqly/readme_examples.atago.yaml`
### Scenario: prints the full user table as an ASCII table
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --sql "SELECT * FROM user" user.csv
```
#### Then
- exit code is `0`
- stdout contains `user_name`, `booker12`, `smith79`
### Scenario: joins two files on a shared key
#### Given
- Fixture file `user.csv` is created.
- Fixture file `identifier.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
3,neet
```
#### When
```shell
sqly --sql "SELECT user_name, position FROM user JOIN identifier ON user.identifier = identifier.id" user.csv identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `position`, `developrt`
### Scenario: runs the analytics script (CTE + window + GROUP BY)
#### Given
- Fixture file `actor.csv` is created.
- Fixture file `analytics.sql` is created.
#### When
```shell
sqly --sql-file analytics.sql actor.csv
```
#### Then
- exit code is `0`
- stdout contains `Harrison Ford`, `50+ movies`
### Scenario: extracts JSON fields from a JSONL file
#### Given
- Fixture file `sample.jsonl` is created.
- Fixture file `json.sql` is created.
#### When
```shell
sqly --sql-file json.sql sample.jsonl
```
#### Then
- exit code is `0`
- stdout contains `Charlie`, `Nagoya`
### Scenario: reads a gzipped CSV transparently
#### Given
- Fixture file `user.csv.gz` is created.
#### When
```shell
sqly --csv --sql "SELECT user_name FROM user ORDER BY identifier LIMIT 1" user.csv.gz
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: queries a Parquet file
#### Given
- Fixture file `products.parquet` is created.
#### When
```shell
sqly --csv --sql "SELECT name FROM products ORDER BY CAST(price AS REAL) DESC LIMIT 1" products.parquet
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: joins a compressed CSV with a plain CSV
#### Given
- Fixture file `user.csv.gz` is created.
- Fixture file `identifier.csv` is created.
#### Inputs
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
3,neet
```
#### When
```shell
sqly --csv --sql "SELECT user_name, position FROM user JOIN identifier ON user.identifier = identifier.id ORDER BY user.identifier LIMIT 1" user.csv.gz identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `developrt`
### Scenario: renders CSV with --csv
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --csv --sql "SELECT user_name, identifier FROM user LIMIT 2" user.csv
```
#### Then
- stdout equals an exact value
- stdout equals an exact value
- stdout equals an exact value
### Scenario: renders JSON with --json
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --json --sql "SELECT user_name, identifier FROM user LIMIT 2" user.csv
```
#### Then
- stdout contains `{"user_name":"booker12","identifier":"1"}`
### Scenario: renders NDJSON with --ndjson
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --ndjson --sql "SELECT user_name, identifier FROM user LIMIT 2" user.csv
```
#### Then
- stdout equals an exact value
- stdout equals an exact value
### Scenario: renders a markdown table with --markdown
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --markdown --sql "SELECT user_name, identifier FROM user LIMIT 2" user.csv
```
#### Then
- stdout equals an exact value
- stdout contains `| booker12 | 1 |`
### Scenario: renders LTSV with --ltsv
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --ltsv --sql "SELECT user_name, identifier FROM user LIMIT 1" user.csv
```
#### Then
- stdout equals an exact value
### Scenario: writes CSV to the path given by --output
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --sql "SELECT * FROM user" --output out.csv user.csv
```
#### Then
- exit code is `0`
- stderr contains `out.csv`
- file `out.csv` contains `booker12`
### Scenario: queries piped CSV through the default stdin table
#### Inputs
_stdin for `sqly`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --stdin csv --sql "SELECT user_name FROM stdin LIMIT 1"
```
#### Then
- exit code is `0`
- stdout contains `booker12`
### Scenario: joins piped stdin with a file argument
#### Given
- Fixture file `identifier.csv` is created.
#### Inputs
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
3,neet
```
_stdin for `sqly`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --stdin csv --csv --sql "SELECT s.user_name, i.position FROM stdin s JOIN identifier i ON s.identifier = i.id" identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `developrt`
### Scenario: runs a helper command and a query from piped stdin
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
_stdin for `sqly`:_
```text
.tables
SELECT COUNT(*) FROM user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `TABLE NAME`, `user`, `3`
### Scenario: runs SQL from join.sql while stdin carries a dataset
#### Given
- Fixture file `identifier.csv` is created.
- Fixture file `join.sql` is created.
#### Inputs
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
3,neet
```
_stdin for `sqly`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --stdin csv --sql-file join.sql identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `developrt`
### Scenario: prints a JSON inspect report with a stable source and column types
#### Given
- Fixture file `identifier.csv` is created.
#### Inputs
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
3,neet
```
#### When
```shell
sqly --inspect --inspect-sample 1 identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `"name": "identifier"`, `identifier.csv`, `"type": "INTEGER"`, `"position": "developrt"`
### Scenario: omits sample rows with --inspect-sample 0
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --inspect --inspect-sample 0 user.csv
```
#### Then
- exit code is `0`
- stdout contains `"sample_rows": []`
### Scenario: prints the CREATE TABLE statement with .schema
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
_stdin for `sqly`:_
```text
.schema user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `CREATE TABLE "user"`, `"identifier" INTEGER`
### Scenario: prints column information with .describe
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
_stdin for `sqly`:_
```text
.describe user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `notnull`, `user_name`, `INTEGER`
### Scenario: writes the updated table to --save-dir, leaving the source untouched
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --sql "UPDATE user SET first_name = 'Rachelle' WHERE identifier = 1" --save-dir out user.csv
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `Saved user to`
- file `user.csv` is checked
- file `out/user.csv` contains `Rachelle`
### Scenario: rejects --save without --force
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --sql "UPDATE user SET identifier = identifier + 100" --save user.csv
```
#### Then
- exit code is `1`
- stderr contains `force`
- file `user.csv` is checked
### Scenario: rejects a schema-changing statement under --save-dir before writing anything
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --sql "CREATE TABLE backup AS SELECT * FROM user" --save-dir out user.csv
```
#### Then
- exit code is `1`
- stderr contains `cannot persist`
- file `out` does not exist
### Scenario: overwrites the source in place with --save --force
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
jenkins46,2,Mary,Jenkins
smith79,3,Jamie,Smith
```
#### When
```shell
sqly --sql "UPDATE user SET identifier = identifier + 100" --save --force user.csv
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `Saved user to`
- file `user.csv` contains `101`
### Scenario: imports every supported file under a directory
#### Given
- Fixture file `imp/user.csv` is created.
- Fixture file `imp/identifier.csv` is created.
#### Inputs
_Fixture `imp/user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `imp/identifier.csv`:_
```text
id,position
1,developrt
```
_stdin for `sqly`:_
```text
.tables
```
#### When
```shell
sqly imp
```
#### Then
- exit code is `0`
- stdout contains `user`, `identifier`
- stderr contains `Successfully imported`
### Scenario: loads ACH records into multiple tables
#### Given
- Fixture file `ppd-debit.ach` is created.
#### Inputs
_stdin for `sqly`:_
```text
.tables
```
#### When
```shell
sqly ppd-debit.ach
```
#### Then
- exit code is `0`
- stdout contains `ppd_debit_file_header`, `ppd_debit_entries`
### Scenario: queries the ACH entries table
#### Given
- Fixture file `ppd-debit.ach` is created.
#### When
```shell
sqly --csv --sql "SELECT amount FROM ppd_debit_entries" ppd-debit.ach
```
#### Then
- exit code is `0`
- stdout contains `amount`
### Scenario: loads a Fedwire file into a single message table
#### Given
- Fixture file `customer-transfer.fed` is created.
#### Inputs
_stdin for `sqly`:_
```text
.tables
```
#### When
```shell
sqly customer-transfer.fed
```
#### Then
- exit code is `0`
- stdout contains `customer_transfer_message`
## sqly interactive shell (pty)
Source: `test/e2e/tools/sqly/repl_pty.atago.yaml`
### Scenario: run a query and read its rendered result table over a pty
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor,gross
Harrison Ford,4871
Samuel L. Jackson,4772
```
#### When
```shell
# interactive (pty): sqly actor.csv
```
#### Then
- exit code is `0`
- stdout contains `Harrison Ford`, `+--`
- rendered screen contains `Harrison Ford`
### Scenario: the .tables dot-command renders the imported table over a pty
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor,gross
Harrison Ford,4871
Samuel L. Jackson,4772
```
#### When
```shell
# interactive (pty): sqly actor.csv
```
#### Then
- exit code is `0`
- stdout contains `TABLE NAME`, `actor`
### Scenario: a computed aggregate round-trips through the pty shell
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor,gross
Harrison Ford,4871
Samuel L. Jackson,4772
```
#### When
```shell
# interactive (pty): sqly actor.csv
```
#### Then
- exit code is `0`
- stdout contains `ROWS=2`
### Scenario: the prompt reflects a live .mode switch over a pty
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor,gross
Harrison Ford,4871
Samuel L. Jackson,4772
```
#### When
```shell
# interactive (pty): sqly actor.csv
```
#### Then
- exit code is `0`
## sqly sandbox_home + changes (history DB isolation)
Source: `test/e2e/tools/sqly/sandbox_home.atago.yaml`
### Scenario: sqly --sql writes exactly its history DB, only inside the sandbox home
_skipped on windows_
#### Given
- Fixture file `user.csv` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT identifier FROM user WHERE user_name = 'booker12'" user.csv
```
#### Then
- exit code is `0`
- stdout contains `1`
- the step changed exactly created `.atago-home/.config/sqly/history.db`, modified nothing, deleted nothing
- file `.atago-home/.config/sqly/history.db` exists
#### Generated artifacts
- `.atago-home/.config/sqly/history.db`
### Scenario: a second sqly batch run leaves its sandbox home byte-identical
_skipped on windows_
#### Given
- Fixture file `user.csv` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT 1" user.csv
sqly --sql "SELECT 2" user.csv
```
#### Then
- after `sqly --sql "SELECT 1" user.csv`:
  - exit code is `0`
- after `sqly --sql "SELECT 2" user.csv`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted nothing
## sqly write-back
Source: `test/e2e/tools/sqly/save.atago.yaml`
### Scenario: writes to --save-dir without modifying the source
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
#### When
```shell
sqly --sql "UPDATE u SET first_name = 'CHANGED' WHERE identifier = 1" u.csv --save-dir out
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `Saved u to`
- file `u.csv` is checked
- file `out/u.csv` contains `CHANGED`
### Scenario: refuses --save without --force
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
#### When
```shell
sqly --sql "UPDATE u SET first_name = 'X'" u.csv --save
```
#### Then
- exit code is `1`
- stderr contains `--force`
- file `u.csv` is checked
### Scenario: overwrites the source in place with --save --force
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
#### When
```shell
sqly --sql "DELETE FROM u WHERE identifier > 1" u.csv --save --force
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `Saved u to`
### Scenario: re-imports a file rewritten in place (round-trip)
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
#### When
```shell
sqly --sql "DELETE FROM u WHERE identifier > 1" u.csv --save --force
sqly --csv --sql "SELECT COUNT(*) AS c FROM u" u.csv
```
#### Then
- after `sqly --csv --sql "SELECT COUNT(*) AS c FROM u" u.csv`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: preserves gzip compression on in-place save
#### Given
- Fixture file `c.csv.gz` is created.
#### When
```shell
sqly --sql "UPDATE c SET first_name = 'GZ' WHERE identifier = 1" c.csv.gz --save --force
sqly --csv --sql "SELECT first_name FROM c WHERE identifier = 1" c.csv.gz
```
#### Then
- after `sqly --csv --sql "SELECT first_name FROM c WHERE identifier = 1" c.csv.gz`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: saves via the .save command in batch mode
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name
booker12,1,Rachel
jenkins46,2,Mary
```
_stdin for `sqly`:_
```text
UPDATE u SET first_name = 'BATCH' WHERE identifier = 1;
.save --force
```
#### When
```shell
sqly u.csv
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `Saved u to`
- file `u.csv` contains `BATCH`
### Scenario: guides a non-interactive --save with no input files toward passing input
#### When
```shell
sqly --save --force --sql "UPDATE foo SET x = 1"
```
#### Then
- exit code is `1`
- stderr contains `no tables to save`, `input files`
### Scenario: guides a batch .save with no imported tables toward passing input
#### Inputs
_stdin for `sqly`:_
```text
.save --force
```
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stderr contains `no tables to save`, `input files`
## sqly schema inspection
Source: `test/e2e/tools/sqly/schema.atago.yaml`
### Scenario: .schema prints a CREATE TABLE statement for a CSV table
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.schema user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `CREATE TABLE`, `user_name`
### Scenario: .schema emits a structured object in json mode
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode json
.schema user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout at `$[0].table` equals `user`
- stdout at `$[0].schema` matches `/^CREATE TABLE/`
- stderr contains `Change output mode`
### Scenario: .schema returns the stored CREATE VIEW for a differently cased view name
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
CREATE VIEW v AS SELECT 1 AS x;
.schema V
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `CREATE VIEW`
- stdout does not contain `CREATE TABLE`
### Scenario: .schema errors on a missing table
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.schema no_such_table
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `no such table`
### Scenario: .describe lists columns and types for a CSV table
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.describe user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `user_name`, `identifier`
### Scenario: .describe emits structured column metadata in json mode
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode json
.describe user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout at `$[0].name` equals `user_name`
- stdout at `$[0].type` equals `TEXT`
### Scenario: .describe errors on a missing table
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.describe no_such_table
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `no such table`
## sqly --sheet validation
Source: `test/e2e/tools/sqly/sheet_flag.atago.yaml`
### Scenario: rejects --sheet with a non-Excel file and --sql
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "SELECT * FROM user" --sheet "A test" user.csv
```
#### Then
- exit code is `1`
- stderr contains `--sheet`
### Scenario: rejects --sheet with a non-Excel file and --inspect
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --sheet "A test" user.csv
```
#### Then
- exit code is `1`
- stderr contains `--sheet`
### Scenario: still imports an Excel file with --sheet
#### Given
- Fixture file `sample.xlsx` is created.
#### When
```shell
sqly --csv --sql "SELECT * FROM sample_test_sheet" --sheet test_sheet sample.xlsx
```
#### Then
- exit code is `0`
- stdout contains `name`
### Scenario: rejects an explicit empty --sheet
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --sheet "" user.csv
```
#### Then
- exit code is `1`
- stderr contains `sheet`
### Scenario: rejects --sheet for a directory with no Excel files
#### Given
- Fixture file `dir/u.csv` is created.
#### Inputs
_Fixture `dir/u.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --sheet anything dir
```
#### Then
- exit code is `1`
- stderr contains `--sheet`
### Scenario: tells the user how to recover when --sheet has no Excel input
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --inspect --sheet "A test" user.csv
```
#### Then
- exit code is `1`
- stderr contains `Excel`, `remove --sheet`
### Scenario: names the workbook and suggests recovery on a single-workbook sheet miss
#### Given
- Fixture file `sample.xlsx` is created.
#### When
```shell
sqly --inspect --sheet no_such_sheet sample.xlsx
```
#### Then
- exit code is `1`
- stderr contains `sample.xlsx`, `without --sheet`
### Scenario: names every checked workbook on a multi-workbook sheet miss
#### Given
- Fixture file `sample.xlsx` is created.
- Fixture file `sheet_with_accents.xlsx` is created.
#### When
```shell
sqly --inspect --sheet no_such_sheet sample.xlsx sheet_with_accents.xlsx
```
#### Then
- exit code is `1`
- stderr contains `sample.xlsx`, `sheet_with_accents.xlsx`, `without --sheet`
## sqly
Source: `test/e2e/tools/sqly/smoke.atago.yaml`
### Scenario: count rows in a CSV fixture
#### Given
- Fixture file `users.csv` is created.
#### Inputs
_Fixture `users.csv`:_
```text
id,name
1,Alice
2,Bob
```
#### When
```shell
sqly --sql 'SELECT count(*) AS cnt FROM users' --csv users.csv
```
#### Then
- exit code is `0`
- stdout contains `2`
### Scenario: filter rows and select a column
#### Given
- Fixture file `users.csv` is created.
#### Inputs
_Fixture `users.csv`:_
```text
id,name
1,Alice
2,Bob
```
#### When
```shell
sqly --sql 'SELECT name FROM users WHERE id = 1' --csv users.csv
```
#### Then
- exit code is `0`
- stdout contains `Alice`
### Scenario: markdown output format works
#### Given
- Fixture file `users.csv` is created.
#### Inputs
_Fixture `users.csv`:_
```text
id,name
1,Alice
```
#### When
```shell
sqly --sql 'SELECT name FROM users' --markdown users.csv
```
#### Then
- exit code is `0`
- stdout matches `/\|-+\|/`
- stdout contains `Alice`
## sqly --sql-file
Source: `test/e2e/tools/sqly/sql_file.atago.yaml`
### Scenario: runs a multiline query loaded from a file against a file input
#### Given
- Fixture file `actor.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
_Fixture `q.sql`:_
```text
-- top actor by name
SELECT actor
FROM actor
ORDER BY actor
LIMIT 1;
```
#### When
```shell
sqly --csv --sql-file q.sql actor.csv
```
#### Then
- exit code is `0`
- stdout contains `Adam Sandler`
### Scenario: joins a piped --stdin dataset with a query loaded from a file
#### Given
- Fixture file `identifier.csv` is created.
- Fixture file `join.sql` is created.
#### Inputs
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
```
_Fixture `join.sql`:_
```text
SELECT s.name, i.position
FROM stdin s
JOIN identifier i ON s.id = i.id
ORDER BY s.id;
```
_stdin for `sqly`:_
```text
id,name
1,alice
2,bob
```
#### When
```shell
sqly --stdin csv --csv --sql-file join.sql identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `alice`, `developrt`
### Scenario: runs multiple statements from a file in order
#### Given
- Fixture file `actor.csv` is created.
- Fixture file `multi.sql` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
_Fixture `multi.sql`:_
```text
SELECT 'first' AS x;
SELECT 'second' AS x;
```
#### When
```shell
sqly --csv --sql-file multi.sql actor.csv
```
#### Then
- exit code is `0`
- stdout contains `first`, `second`
### Scenario: rejects --sql and --sql-file together
#### Given
- Fixture file `actor.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
_Fixture `q.sql`:_
```text
SELECT 1;
```
#### When
```shell
sqly --sql "SELECT 1" --sql-file q.sql actor.csv
```
#### Then
- exit code is `1`
- stderr contains `--sql-file`
### Scenario: fails for a missing SQL file
#### Given
- Fixture file `actor.csv` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
#### When
```shell
sqly --sql-file no_such.sql actor.csv
```
#### Then
- exit code is `1`
- stderr contains `sql-file`
### Scenario: fails for an empty SQL file
#### Given
- Fixture file `actor.csv` is created.
- Fixture file `empty.sql` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
#### When
```shell
sqly --sql-file empty.sql actor.csv
```
#### Then
- exit code is `1`
- stderr contains `empty`
### Scenario: locates a failing statement by its line in the SQL file
#### Given
- Fixture file `actor.csv` is created.
- Fixture file `bad.sql` is created.
#### Inputs
_Fixture `actor.csv`:_
```text
actor
Adam Sandler
Harrison Ford
```
_Fixture `bad.sql`:_
```text
SELECT 1;
SELECT 2;
SELECT * FROM no_such_table;
```
#### When
```shell
sqly --sql-file bad.sql actor.csv
```
#### Then
- exit code is `1`
- stderr contains `batch statement 3 failed at line 3`, `no_such_table`
## sqly --sql-file --output
Source: `test/e2e/tools/sqly/sqlfile_output.atago.yaml`
### Scenario: exports a single-SELECT script to the output file with clean stdout
#### Given
- Fixture file `user.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
_Fixture `q.sql`:_
```text
SELECT user_name FROM user ORDER BY identifier LIMIT 1;
```
#### When
```shell
sqly --sql-file q.sql --output out.csv user.csv
```
#### Then
- exit code is `0`
- stdout is empty
- stderr contains `Output sql result`
- file `out.csv` contains `user_name`
### Scenario: exports a single result set even when the script first runs DDL/DML
#### Given
- Fixture file `user.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
_Fixture `q.sql`:_
```text
CREATE TEMP TABLE picked AS SELECT user_name FROM user;
SELECT * FROM picked ORDER BY user_name LIMIT 1;
```
#### When
```shell
sqly --sql-file q.sql --output out.csv user.csv
```
#### Then
- exit code is `0`
- stdout is empty
- stderr contains `Output sql result`
- file `out.csv` exists
#### Generated artifacts
- `out.csv`
### Scenario: rejects a script that produces no result set
#### Given
- Fixture file `user.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
_Fixture `q.sql`:_
```text
CREATE TEMP TABLE t AS SELECT 1 AS x;
```
#### When
```shell
sqly --sql-file q.sql --output out.csv user.csv
```
#### Then
- exit code is `1`
- stderr contains `result set`
- file `out.csv` does not exist
### Scenario: rejects a script that produces multiple result sets
#### Given
- Fixture file `user.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
jenkins46,2
```
_Fixture `q.sql`:_
```text
SELECT user_name FROM user LIMIT 1;
SELECT identifier FROM user LIMIT 1;
```
#### When
```shell
sqly --sql-file q.sql --output out.csv user.csv
```
#### Then
- exit code is `1`
- stderr contains `single result set`
- file `out.csv` does not exist
## sqly --stdin dataset
Source: `test/e2e/tools/sqly/stdin_dataset.atago.yaml`
### Scenario: queries piped CSV through the default stdin table
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,alice
2,bob
```
#### When
```shell
sqly --stdin csv --csv --sql "SELECT name FROM stdin ORDER BY id"
```
#### Then
- exit code is `0`
- stdout equals an exact value
- stdout equals an exact value
- stdout equals an exact value
### Scenario: queries piped TSV data
#### Inputs
_stdin for `sqly`:_
```text
id	name
1	alice
```
#### When
```shell
sqly --stdin tsv --csv --sql "SELECT COUNT(*) AS c FROM stdin"
```
#### Then
- exit code is `0`
- stdout contains `1`
### Scenario: queries piped JSONL data stored in a data column
#### Inputs
_stdin for `sqly`:_
```text
{"id":1,"name":"alice"}
{"id":2,"name":"bob"}
```
#### When
```shell
sqly --stdin jsonl --csv --sql "SELECT COUNT(*) AS c FROM stdin"
```
#### Then
- exit code is `0`
- stdout contains `2`
### Scenario: overrides the stdin table name with --stdin-name
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,alice
2,bob
```
#### When
```shell
sqly --stdin csv --stdin-name people --csv --sql "SELECT COUNT(*) FROM people"
```
#### Then
- exit code is `0`
- stdout contains `2`
### Scenario: joins piped stdin with an imported file argument
#### Given
- Fixture file `identifier.csv` is created.
#### Inputs
_Fixture `identifier.csv`:_
```text
id,position
1,developrt
2,manager
```
_stdin for `sqly`:_
```text
id,name
1,alice
2,bob
```
#### When
```shell
sqly --stdin csv --csv --sql "SELECT s.name, i.position FROM stdin s JOIN identifier i ON s.id = i.id ORDER BY s.id" identifier.csv
```
#### Then
- exit code is `0`
- stdout contains `alice`, `developrt`
### Scenario: reports a stable stdin source in --inspect, not a temp path
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,alice
```
#### When
```shell
sqly --stdin csv --inspect
```
#### Then
- exit code is `0`
- stdout contains `"source": "stdin"`
- stdout does not contain `sqly-stdin-`
### Scenario: rejects --save --force for a stdin-backed table
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,alice
```
#### When
```shell
sqly --stdin csv --sql "UPDATE stdin SET name = 'x'" --save --force
```
#### Then
- exit code is `1`
- stdout does not contain `affected`
- stderr contains `stdin`
### Scenario: rejects a non-identifier --stdin-name so the name stays queryable
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,alice
```
#### When
```shell
sqly --stdin csv --stdin-name "my data" --sql 'SELECT * FROM "my data"'
```
#### Then
- exit code is `1`
- stderr contains `stdin-name`
### Scenario: rejects a path-like --stdin-name
#### Inputs
_stdin for `sqly`:_
```text
a
1
```
#### When
```shell
sqly --stdin csv --stdin-name "../escaped" --sql "SELECT 1"
```
#### Then
- exit code is `1`
- stderr contains `stdin-name`
- file `escaped.csv` does not exist
### Scenario: reports a clear error for an unsupported stdin format
#### Inputs
_stdin for `sqly`:_
```text
a,b
1,2
```
#### When
```shell
sqly --stdin xml --sql "SELECT 1"
```
#### Then
- exit code is `1`
- stderr contains `unsupported --stdin format`
### Scenario: still reads stdin as SQL and helper commands without --stdin
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.tables
SELECT user_name FROM user ORDER BY identifier LIMIT 1
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `TABLE NAME`, `booker12`
## sqly typed JSON output
Source: `test/e2e/tools/sqly/typed_json.atago.yaml`
### Scenario: emits native numbers, booleans, and null with --json-typed
#### When
```shell
sqly --json-typed --sql "SELECT 42 AS i, -1.5 AS f, NULL AS n, 'x' AS s"
```
#### Then
- exit code is `0`
- stdout contains `"i":42`, `"f":-1.5`, `"n":null`, `"s":"x"`
### Scenario: keeps the legacy string contract with plain --json
#### When
```shell
sqly --json --sql "SELECT 42 AS i"
```
#### Then
- stdout contains `"i":"42"`
### Scenario: emits native scalars per line with --ndjson-typed
#### When
```shell
sqly --ndjson-typed --sql "SELECT 7 AS n, 't' AS s"
```
#### Then
- stdout contains `"n":7`, `"s":"t"`
### Scenario: keeps a large integer column lossless (no scientific notation)
#### Given
- Fixture file `typed_bigint.csv` is created.
#### Inputs
_Fixture `typed_bigint.csv`:_
```text
id,amount,flag
1,9007199254740993,true
2,42,false
```
#### When
```shell
sqly --json-typed --sql "SELECT amount FROM typed_bigint WHERE id = 1" typed_bigint.csv
```
#### Then
- stdout contains `"amount":9007199254740993`
- stdout does not contain `e+`
### Scenario: leaves a leading-zero value as a string
#### When
```shell
sqly --json-typed --sql "SELECT '007' AS code"
```
#### Then
- stdout contains `"code":"007"`
### Scenario: uses the typed contract for --inspect sample rows
#### Given
- Fixture file `typed_bigint.csv` is created.
#### Inputs
_Fixture `typed_bigint.csv`:_
```text
id,amount,flag
1,9007199254740993,true
2,42,false
```
#### When
```shell
sqly --inspect --json-typed typed_bigint.csv
```
#### Then
- stdout contains `"amount": 9007199254740993`
### Scenario: rejects plain --json combined with --inspect
#### Given
- Fixture file `typed_bigint.csv` is created.
#### Inputs
_Fixture `typed_bigint.csv`:_
```text
id,amount,flag
1,9007199254740993,true
2,42,false
```
#### When
```shell
sqly --inspect --json typed_bigint.csv
```
#### Then
- exit code is `1`
- stderr contains `inspect`
## sqly v0.18.0 binary bug fixes
Source: `test/e2e/tools/sqly/v0_18_bugs.atago.yaml`
### Scenario: rejects an empty --output
#### When
```shell
sqly --sql "SELECT 1 AS x" --output ""
```
#### Then
- exit code is `1`
- stderr contains `--output`
### Scenario: rejects an empty --sql-file
#### When
```shell
sqly --sql-file ""
```
#### Then
- exit code is `1`
- stderr contains `--sql-file`
### Scenario: rejects an empty --save-dir
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "SELECT 1" --save-dir "" user.csv
```
#### Then
- exit code is `1`
- stderr contains `--save-dir`
### Scenario: rejects an empty --stdin
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,a
```
#### When
```shell
sqly --stdin "" --sql "SELECT 1 AS x"
```
#### Then
- exit code is `1`
- stderr contains `--stdin`
### Scenario: rejects conflicting output mode flags
#### When
```shell
sqly --csv --json --sql "SELECT 1 AS x"
```
#### Then
- exit code is `1`
- stderr contains `conflicting`
### Scenario: prints rows for a DML RETURNING statement
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql "UPDATE u SET first_name='X' WHERE identifier=1 RETURNING identifier, first_name" u.csv
```
#### Then
- exit code is `0`
- stdout contains `X`
- stdout does not contain `affected`
### Scenario: rejects --output for a non-rowset DML statement
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "UPDATE u SET first_name='X' WHERE identifier=1" --output out.csv u.csv
```
#### Then
- exit code is `1`
- stderr contains `--output`
- file `out.csv` does not exist
### Scenario: exports RETURNING rows with --output
#### Given
- Fixture file `u.csv` is created.
#### Inputs
_Fixture `u.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql "UPDATE u SET first_name='X' WHERE identifier=1 RETURNING identifier" --output out.csv u.csv
```
#### Then
- exit code is `0`
- stderr contains `Output sql result`
- file `out.csv` exists
#### Generated artifacts
- `out.csv`
### Scenario: rejects a comment-only --sql-file
#### Given
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `q.sql`:_
```text
-- header only
/* block */
```
#### When
```shell
sqly --sql-file q.sql
```
#### Then
- exit code is `1`
- stderr contains `no executable`
### Scenario: strips a UTF-8 BOM from a --sql-file script
#### Given
- Fixture file `user.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql-file q.sql user.csv
```
#### Then
- exit code is `0`
- stdout contains `2`
### Scenario: strips a UTF-8 BOM from batch stdin
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
﻿SELECT 7 AS z;
```
#### When
```shell
sqly --csv user.csv
```
#### Then
- exit code is `0`
- stdout contains `7`
### Scenario: rejects non-empty piped stdin with --sql-file
#### Given
- Fixture file `user.csv` is created.
- Fixture file `q.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `q.sql`:_
```text
SELECT 1 AS x;
```
_stdin for `sqly`:_
```text
SELECT 999 AS y;
```
#### When
```shell
sqly --sql-file q.sql user.csv
```
#### Then
- exit code is `1`
- stderr contains `stdin`
### Scenario: fails a --stdin dataset run with no query
#### Inputs
_stdin for `sqly`:_
```text
id,name
1,a
```
#### When
```shell
sqly --stdin csv
```
#### Then
- exit code is `1`
- stderr contains `--stdin`
### Scenario: reports per-file provenance for a sanitized basename
#### Given
- Fixture file `dir/2023-data.csv` is created.
#### Inputs
_Fixture `dir/2023-data.csv`:_
```text
id,name
1,a
```
#### When
```shell
sqly --inspect dir
```
#### Then
- exit code is `0`
- stdout contains `2023-data.csv`
- stderr does not contain `Successfully imported`
### Scenario: rejects duplicate basenames from different subdirectories
#### Given
- Fixture file `dir/a/user.csv` is created.
- Fixture file `dir/b/user.csv` is created.
#### Inputs
_Fixture `dir/a/user.csv`:_
```text
id,name
1,alpha
```
_Fixture `dir/b/user.csv`:_
```text
id,name
2,beta
```
#### When
```shell
sqly --inspect dir
```
#### Then
- exit code is `1`
- stderr contains `collision`
### Scenario: reports an overwrite when re-importing a directory
#### Given
- Fixture file `user.csv` is created.
- Fixture file `dir/user.csv` is created.
- Fixture file `cmds.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `dir/user.csv`:_
```text
user_name,identifier,first_name,last_name
alt1,1,ALT,One
```
_Fixture `cmds.sql`:_
```text
.import dir
SELECT user_name FROM user ORDER BY identifier;
```
#### When
```shell
sqly --sql-file cmds.sql user.csv
```
#### Then
- exit code is `0`
- stdout contains `alt1`
- stderr does not contain `No supported files`
### Scenario: rejects --output that aliases an imported source
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql "SELECT * FROM user WHERE identifier=1" --output user.csv user.csv
```
#### Then
- exit code is `1`
- stderr contains `--output`
### Scenario: rejects --save-dir that resolves to the source directory
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "UPDATE user SET first_name='P' WHERE identifier=1" --save-dir . user.csv
```
#### Then
- exit code is `1`
- stderr contains `source`
### Scenario: rejects a --save-dir destination that already exists
#### Given
- Fixture file `src/user.csv` is created.
- Fixture file `out/user.csv` is created.
#### Inputs
_Fixture `src/user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `out/user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "UPDATE user SET first_name='Q' WHERE identifier=1" --save-dir out src/user.csv
```
#### Then
- exit code is `1`
- stderr contains `already exists`
### Scenario: keeps stdout clean when write-back fails
#### Given
- Fixture file `user.csv` is created.
- Fixture file `sample.xlsx` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "UPDATE user SET first_name='X' WHERE identifier=1" --save-dir out user.csv sample.xlsx
```
#### Then
- exit code is `1`
- stdout does not contain `affected`
- stderr contains `cannot save`
### Scenario: skips write-back for a read-only query under --save --force
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql "SELECT * FROM user WHERE identifier=1" --save --force user.csv
```
#### Then
- exit code is `0`
- stdout contains `booker12`
- stderr does not contain `Saved`
### Scenario: skips workbooks lacking the requested sheet (multi-workbook --sheet)
#### Given
- Fixture file `sheet_with_spaces.xlsx` is created.
- Fixture file `sample.xlsx` is created.
- Fixture file `sheet_with_accents.xlsx` is created.
#### When
```shell
sqly --inspect --sheet "A test" sheet_with_spaces.xlsx sample.xlsx sheet_with_accents.xlsx
```
#### Then
- exit code is `0`
- stdout contains `sheet_with_spaces`
- stderr contains `Skipped`
## sqly v0.19.0 binary bug fixes
Source: `test/e2e/tools/sqly/v0_19_bugs.atago.yaml`
### Scenario: quotes a CSV value containing a comma
#### When
```shell
sqly --csv --sql "SELECT 'a,b' AS c"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: quotes a CSV value containing a double quote
#### When
```shell
sqly --csv --sql "SELECT 'a' || char(34) || 'b' AS c"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an LTSV value containing a tab
#### When
```shell
sqly --ltsv --sql "SELECT 'a' || char(9) || 'b' AS c"
```
#### Then
- exit code is `1`
- stderr contains `LTSV`
### Scenario: rejects duplicate JSON keys
#### When
```shell
sqly --json --sql "SELECT 1 AS x, 2 AS x"
```
#### Then
- exit code is `1`
- stderr contains `unique column names`
### Scenario: rejects duplicate NDJSON keys
#### When
```shell
sqly --ndjson --sql "SELECT 1 AS x, 2 AS x"
```
#### Then
- exit code is `1`
- stderr contains `unique column names`
### Scenario: keeps a Markdown row on one line when a value has a newline
#### When
```shell
sqly --markdown --sql "SELECT 'a' || char(10) || 'b' AS x"
```
#### Then
- exit code is `0`
- stdout contains `a<br>b`
### Scenario: accepts a leading block comment in direct --sql
#### When
```shell
sqly --csv --sql "/* note */ SELECT 1 AS x"
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: accepts PRAGMA in direct --sql
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql "PRAGMA table_info(user)" user.csv
```
#### Then
- exit code is `0`
- stdout contains `user_name`
### Scenario: accepts VALUES in direct --sql
#### When
```shell
sqly --csv --sql "VALUES (1), (2)"
```
#### Then
- exit code is `0`
- stdout contains `1`
### Scenario: accepts the TABLE shorthand in direct --sql
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --csv --sql "TABLE user" user.csv
```
#### Then
- exit code is `0`
- stdout contains `booker12`
### Scenario: accepts CREATE TABLE in direct --sql
#### When
```shell
sqly --sql "CREATE TABLE t(x)"
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
### Scenario: accepts ANALYZE in direct --sql
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "ANALYZE" user.csv
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
### Scenario: runs WITH ... UPDATE without RETURNING as DML
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "WITH s AS (SELECT 1 AS identifier) UPDATE user SET first_name='Z' WHERE identifier IN (SELECT identifier FROM s)" user.csv
```
#### Then
- exit code is `0`
- stdout contains `affected is 1 row`
### Scenario: rejects --stdin-name without --stdin
#### When
```shell
sqly --stdin-name weird --csv --sql "SELECT 1 AS x"
```
#### Then
- exit code is `1`
- stderr contains `stdin-name`
### Scenario: rejects --inspect-sample without --inspect
#### When
```shell
sqly --inspect-sample 0 --csv --sql "SELECT 1 AS x"
```
#### Then
- exit code is `1`
- stderr contains `inspect-sample`
### Scenario: rejects --force without --save
#### When
```shell
sqly --force --sql "SELECT 1 AS x"
```
#### Then
- exit code is `1`
- stderr contains `force`
### Scenario: rejects --inspect combined with --csv
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --inspect --csv user.csv
```
#### Then
- exit code is `1`
- stderr contains `inspect`
### Scenario: imports an empty JSON array as a zero-row table
#### Given
- Fixture file `empty.json` is created.
#### Inputs
_Fixture `empty.json`:_
```text
[]
```
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS n FROM empty" empty.json
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: imports an empty JSONL file as a zero-row table
#### Given
- Fixture file `empty.jsonl` is created.
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS n FROM empty" empty.jsonl
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects an --output path ending with a slash
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "SELECT 1 AS x" --output "outdir/" user.csv
```
#### Then
- exit code is `1`
- stderr contains `separator`
### Scenario: rejects an --output ACH destination
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "SELECT identifier FROM user LIMIT 1" --output out.ach user.csv
```
#### Then
- exit code is `1`
- stderr contains `input-only`
### Scenario: parses a helper command after a terminated statement
#### Inputs
_stdin for `sqly`:_
```text
SELECT 1 AS x;
.mode csv
SELECT 2 AS y;
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `2`
- stderr does not contain `arguments`
### Scenario: parses a helper command after a leading comment
#### Inputs
_stdin for `sqly`:_
```text
-- header
.mode csv
SELECT 1 AS x;
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `1`
- stderr does not contain `arguments`
### Scenario: does not write back for an EXPLAIN under --save-dir
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "EXPLAIN UPDATE user SET first_name='X' WHERE identifier=1" --save-dir out user.csv
```
#### Then
- exit code is `0`
- file `out/user.csv` does not exist
### Scenario: does not write back for a zero-row DML under --save-dir
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "UPDATE user SET first_name='X' WHERE identifier=999999" --save-dir out user.csv
```
#### Then
- exit code is `0`
- stdout contains `affected is 0`
- file `out/user.csv` does not exist
### Scenario: keeps stdout clean when parquet write-back fails
#### Given
- Fixture file `products.parquet` is created.
#### When
```shell
sqly --sql "DELETE FROM products" --save --force products.parquet
```
#### Then
- exit code is `1`
- stdout is empty
- stderr is not empty
## sqly v0.20.0 binary regressions
Source: `test/e2e/tools/sqly/v0_20_bugs.atago.yaml`
### Scenario: write-back rejects: ALTER TABLE RENAME COLUMN
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "ALTER TABLE user RENAME COLUMN first_name TO fname" --save --force user.csv
```
#### Then
- exit code is `1`
- stdout does not contain `affected is`
- stderr is not empty
### Scenario: write-back rejects: DROP TABLE
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "DROP TABLE user" --save --force user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: write-back rejects: CREATE VIEW
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "CREATE VIEW v AS SELECT user_name FROM user" --save --force user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: write-back rejects: CREATE INDEX
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "CREATE INDEX idx ON user(identifier)" --save --force user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: write-back rejects: CREATE TABLE
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "CREATE TABLE backup (id INTEGER)" --save --force user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: write-back rejects: REINDEX
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "REINDEX" --save --force user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: write-back rejects: ANALYZE
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "ANALYZE" --save --force user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: rejects CREATE TABLE AS SELECT under --save-dir and writes nothing
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "CREATE TABLE backup AS SELECT * FROM user" --save-dir out user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
- file `out` does not exist
### Scenario: preflight rejects a CTAS+DML script before it executes
#### Given
- Fixture file `user.csv` is created.
- Fixture file `s.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `s.sql`:_
```text
CREATE TABLE backup AS SELECT * FROM user;
UPDATE user SET first_name='Z' WHERE identifier=1;
```
#### When
```shell
sqly --sql-file s.sql --save-dir out user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: allows a .import + UPDATE batch under --save-dir and writes the change
#### Given
- Fixture file `testdata/user.csv` is created.
- Fixture file `imp.sql` is created.
#### Inputs
_Fixture `testdata/user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `imp.sql`:_
```text
.import testdata/user.csv
UPDATE user SET first_name='Batch' WHERE identifier=1;
```
#### When
```shell
sqly --sql-file imp.sql --save-dir out
```
#### Then
- exit code is `0`
- stdout contains `affected is 1 row(s)`
- file `out/user.csv` contains `Batch`
### Scenario: neutral success: CREATE VIEW
#### When
```shell
sqly --sql "CREATE VIEW v AS SELECT 1 AS x"
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
- stdout does not contain `affected is`
### Scenario: neutral success: CREATE TABLE
#### When
```shell
sqly --sql "CREATE TABLE t (id INTEGER)"
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
- stdout does not contain `affected is`
### Scenario: neutral success: ANALYZE
#### When
```shell
sqly --sql "ANALYZE"
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
### Scenario: runs a setter PRAGMA
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "PRAGMA user_version = 1" user.csv
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
### Scenario: runs a command PRAGMA that returns no rows
#### When
```shell
sqly --sql "PRAGMA incremental_vacuum"
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
### Scenario: rejects BEGIN in a --sql-file script
#### Given
- Fixture file `user.csv` is created.
- Fixture file `tx.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `tx.sql`:_
```text
BEGIN IMMEDIATE;
UPDATE user SET first_name='TX' WHERE identifier=1;
COMMIT;
```
#### When
```shell
sqly --sql-file tx.sql user.csv
```
#### Then
- exit code is `1`
- stderr contains `transaction`
### Scenario: rejects VACUUM
#### When
```shell
sqly --sql "VACUUM"
```
#### Then
- exit code is `1`
- stderr contains `VACUUM`
### Scenario: rejects VACUUM INTO and writes no file
#### When
```shell
sqly --sql "VACUUM INTO 'dump.db'"
```
#### Then
- exit code is `1`
- stderr contains `VACUUM`
- file `dump.db` does not exist
### Scenario: rejects ATTACH DATABASE and persists no external file
#### Given
- Fixture file `a.sql` is created.
#### Inputs
_Fixture `a.sql`:_
```text
ATTACH DATABASE 'aux.db' AS aux;
CREATE TABLE aux.t (id INTEGER);
```
#### When
```shell
sqly --sql-file a.sql
```
#### Then
- exit code is `1`
- stderr is not empty
- file `aux.db` does not exist
### Scenario: runs a multiline CREATE TRIGGER from a --sql-file
#### Given
- Fixture file `user.csv` is created.
- Fixture file `t.sql` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_Fixture `t.sql`:_
```text
CREATE TRIGGER trig_user AFTER UPDATE ON user BEGIN
  UPDATE user SET last_name='Triggered' WHERE identifier=2;
END;
```
#### When
```shell
sqly --sql-file t.sql user.csv
```
#### Then
- exit code is `0`
- stdout contains `statement executed successfully`
### Scenario: accepts a schema-qualified .schema name
#### Given
- Fixture file `testdata/user.csv` is created.
#### Inputs
_Fixture `testdata/user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
.import testdata/user.csv
.schema main.user
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `user_name`
### Scenario: lists session-created views and temp tables in .tables
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
CREATE TEMP TABLE temp_t (id INTEGER);
CREATE VIEW v_user AS SELECT user_name FROM user;
.tables
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `temp_t`, `v_user`
### Scenario: prints CREATE VIEW for a view in .schema
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
CREATE VIEW v_user AS SELECT user_name FROM user;
.schema v_user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `CREATE VIEW`
### Scenario: imports an empty compressed JSON array as a zero-row table
#### Given
- Fixture file `empty.json.gz` is created.
#### When
```shell
sqly --sql "SELECT COUNT(*) AS c FROM empty" empty.json.gz
```
#### Then
- exit code is `0`
- stdout contains `0`
### Scenario: imports an empty compressed JSONL file as a zero-row table
#### Given
- Fixture file `empty.jsonl.gz` is created.
#### When
```shell
sqly --sql "SELECT COUNT(*) AS c FROM empty" empty.jsonl.gz
```
#### Then
- exit code is `0`
- stdout contains `0`
### Scenario: imports /dev/stdin as CSV
_skipped on windows_
#### Inputs
_stdin for `sqly`:_
```text
name,score
a,1
b,2
c,3
```
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS c FROM stdin" /dev/stdin
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: imports /proc/self/fd/0 as CSV
_only on linux_
#### Inputs
_stdin for `sqly`:_
```text
name,score
a,1
b,2
c,3
```
#### When
```shell
sqly --csv --sql "SELECT COUNT(*) AS c FROM sheet_0" /proc/self/fd/0
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: rejects --output to a multi-compressed ACH destination
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "SELECT * FROM user LIMIT 1" --output out.ach.gz.zst user.csv
```
#### Then
- exit code is `1`
- stderr contains `ACH/Fedwire`
- file `out.ach.gz.zst` does not exist
### Scenario: rejects .dump to a multi-compressed Fedwire destination
#### Given
- Fixture file `testdata/user.csv` is created.
#### Inputs
_Fixture `testdata/user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.import testdata/user.csv
.dump user out.fed.gz.zst
```
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stderr contains `ACH/Fedwire`
- file `out.fed.gz.zst` does not exist
### Scenario: rejects an invalid LTSV output label
#### When
```shell
sqly --ltsv --sql 'SELECT 1 AS "foo:bar"' --output out.ltsv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: rejects duplicate LTSV output labels
#### When
```shell
sqly --ltsv --sql "SELECT 1 AS x, 2 AS x" --output out.ltsv
```
#### Then
- exit code is `1`
- stderr is not empty
### Scenario: rejects an LTSV import with duplicate labels
#### Given
- Fixture file `dup.ltsv` is created.
#### Inputs
_Fixture `dup.ltsv`:_
```text
x:1	x:2
```
#### When
```shell
sqly --sql "SELECT * FROM dup" dup.ltsv
```
#### Then
- exit code is `1`
- stderr is not empty
## sqly v0.21.0 binary regressions
Source: `test/e2e/tools/sqly/v0_21_bugs.atago.yaml`
### Scenario: prefers a TEMP table over a same-named main table in .schema
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
CREATE TEMP TABLE user(id TEXT);
.schema user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `TEMP`
- stdout does not contain `first_name`
### Scenario: prefers a TEMP view over a same-named main table in .schema
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
CREATE TEMP VIEW user AS SELECT 1 AS id;
.schema user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `TEMP VIEW`
### Scenario: keeps both a main and a same-named TEMP object in .tables
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
CREATE TEMP TABLE user(id TEXT);
.tables
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `temp.user`
### Scenario: targets a literal dotted table name in .schema
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "a.b"(id INTEGER);
.schema "a.b"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `id`
### Scenario: targets a literal dotted table name in .describe
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "a.b"(id INTEGER);
.describe "a.b"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `id`
### Scenario: targets a literal dotted table name in .header
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "a.b"(id INTEGER);
.header "a.b"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `id`
### Scenario: targets a literal dotted table name in .dump
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "a.b"(id INTEGER);
.dump "a.b" ab.csv
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- file `ab.csv` contains `id`
### Scenario: prints a paste-safe quoted identifier in .tables
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "two words"(id INTEGER);
.tables
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `"two words"`
### Scenario: keeps the full spaced table name in .header
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "two words"(id INTEGER);
.header "two words"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `two words`
### Scenario: keeps the TEMP keyword for a temp-qualified table in .schema
#### Inputs
_stdin for `sqly`:_
```text
CREATE TEMP TABLE t(id INTEGER PRIMARY KEY);
.schema temp.t
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `TEMP`
### Scenario: keeps the TEMP keyword for a temp-qualified view in .schema
#### Inputs
_stdin for `sqly`:_
```text
CREATE TEMP VIEW v AS SELECT 1 AS id;
.schema temp.v
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `TEMP VIEW`
### Scenario: direct --sql rejects multi-statement input: two SELECTs
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "SELECT 1 AS x; SELECT 2 AS y" user.csv
```
#### Then
- exit code is `1`
- stderr contains `single SQL statement`
### Scenario: direct --sql rejects multi-statement input: SELECT then UPDATE
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "SELECT 1 AS x; UPDATE user SET first_name='z'" user.csv
```
#### Then
- exit code is `1`
- stderr contains `single SQL statement`
### Scenario: rejects multi-statement --sql --output before writing the file
#### When
```shell
sqly --csv --sql "SELECT 1 AS x; SELECT 2 AS y" --output out.csv
```
#### Then
- exit code is `1`
- stderr contains `single SQL statement`
- file `out.csv` does not exist
### Scenario: rejects under --save --force: PRAGMA user_version=1
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "PRAGMA user_version=1" --save --force user.csv
```
#### Then
- exit code is `1`
- stdout does not contain `journal_mode`
- stderr is not empty
### Scenario: rejects under --save --force: PRAGMA incremental_vacuum
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "PRAGMA incremental_vacuum" --save --force user.csv
```
#### Then
- exit code is `1`
- stdout does not contain `journal_mode`
- stderr is not empty
### Scenario: rejects under --save --force: PRAGMA journal_mode=OFF
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "PRAGMA journal_mode=OFF" --save --force user.csv
```
#### Then
- exit code is `1`
- stdout does not contain `journal_mode`
- stderr is not empty
### Scenario: rejects a setter PRAGMA under --save-dir and writes nothing
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "PRAGMA user_version=1" --save-dir out user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
- file `out` does not exist
### Scenario: rejects a command PRAGMA under --save-dir and writes nothing
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
#### When
```shell
sqly --sql "PRAGMA incremental_vacuum" --save-dir out user.csv
```
#### Then
- exit code is `1`
- stderr is not empty
- file `out` does not exist
### Scenario: rejects END in direct --sql
#### When
```shell
sqly --sql "END"
```
#### Then
- exit code is `1`
- stderr contains `transaction`
### Scenario: rejects END in batch stdin
#### Inputs
_stdin for `sqly`:_
```text
END;
```
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stderr contains `transaction`
### Scenario: rejects END in a --sql-file script
#### Given
- Fixture file `end.sql` is created.
#### Inputs
_Fixture `end.sql`:_
```text
END;
```
#### When
```shell
sqly --sql-file end.sql
```
#### Then
- exit code is `1`
- stderr contains `transaction`
### Scenario: --output rejects nested compression suffixes: out.csv.gz.zst
#### When
```shell
sqly --sql "SELECT 1 AS x" --output out.csv.gz.zst
```
#### Then
- exit code is `1`
- file `out.csv.gz.zst` does not exist
### Scenario: --output rejects nested compression suffixes: out.parquet.gz.zst
#### When
```shell
sqly --sql "SELECT 1 AS x" --output out.parquet.gz.zst
```
#### Then
- exit code is `1`
- file `out.parquet.gz.zst` does not exist
### Scenario: --output rejects nested compression suffixes: out.xlsx.gz.zst
#### When
```shell
sqly --sql "SELECT 1 AS x" --output out.xlsx.gz.zst
```
#### Then
- exit code is `1`
- file `out.xlsx.gz.zst` does not exist
### Scenario: rejects a nested .dump destination and writes nothing
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
.dump user d.csv.gz.zst
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- file `d.csv.gz.zst` does not exist
### Scenario: emits structured .tables output under .mode json
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
.mode json
.tables
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `"name"`, `"schema"`
- stderr contains `Change output mode`
### Scenario: emits structured .header output under .mode ndjson
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier,first_name,last_name
booker12,1,Rachel,Booker
```
_stdin for `sqly`:_
```text
.mode ndjson
.header user
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `0`
- stdout contains `"column"`, `first_name`
- stderr contains `Change output mode`
### Scenario: does not rewrite an unchanged source on .save --force
#### Given
- Fixture file `ro.csv` is created.
#### Inputs
_Fixture `ro.csv`:_
```text
user_name,identifier
alice,1
```
_stdin for `sqly`:_
```text
.save --force
```
#### When
```shell
sqly ro.csv
```
#### Then
- exit code is `0`
- stderr contains `nothing to save`
### Scenario: writes no directory export for an unchanged session on .save DIR
#### Given
- Fixture file `ro2.csv` is created.
#### Inputs
_Fixture `ro2.csv`:_
```text
user_name,identifier
alice,1
```
_stdin for `sqly`:_
```text
SELECT 1;
.save out
```
#### When
```shell
sqly ro2.csv
```
#### Then
- exit code is `0`
- stdout contains `1`
- stderr contains `nothing to save`
- file `out` does not exist
## sqly v0.22.0 binary regressions
Source: `test/e2e/tools/sqly/v0_22_bugs.atago.yaml`
### Scenario: inspects a literal "main.x" table with .schema
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "main.x"(litcol INTEGER);
.schema "main.x"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `litcol`
### Scenario: inspects a literal "temp.x" table with .describe
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "temp.x"(litcol INTEGER);
.describe "temp.x"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `litcol`
### Scenario: inspects a literal "main.v" view with .header
#### Inputs
_stdin for `sqly`:_
```text
CREATE VIEW "main.v" AS SELECT 1 AS litcol;
.header "main.v"
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `litcol`
### Scenario: exports a literal "temp.v" view with .dump
#### Inputs
_stdin for `sqly`:_
```text
CREATE VIEW "temp.v" AS SELECT 1 AS litcol;
.dump "temp.v" tv.csv
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- file `tv.csv` contains `litcol`
### Scenario: prints a paste-safe literal "main.x" name in .tables
#### Inputs
_stdin for `sqly`:_
```text
CREATE TABLE "main.x"(litcol INTEGER);
.tables
```
#### When
```shell
sqly
```
#### Then
- exit code is `0`
- stdout contains `"main.x"`
### Scenario: rejects a --output destination that stacks .gzip and .zst on a format suffix
#### When
```shell
sqly --sql "SELECT 1 AS x" --output fake.parquet.gzip.zst
```
#### Then
- exit code is `1`
- file `fake.parquet.gzip.zst` does not exist
### Scenario: rejects a --output .json.gzip.zst destination
#### When
```shell
sqly --sql "SELECT 1 AS x" --output fake.json.gzip.zst
```
#### Then
- exit code is `1`
- file `fake.json.gzip.zst` does not exist
### Scenario: rejects a --output .ach.gzip.zst destination as input-only
#### When
```shell
sqly --sql "SELECT 1 AS x" --output fake.ach.gzip.zst
```
#### Then
- exit code is `1`
- file `fake.ach.gzip.zst` does not exist
### Scenario: rejects a .dump destination that stacks .gzip and .zst
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump user fake.parquet.gzip.zst
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- file `fake.parquet.gzip.zst` does not exist
### Scenario: runs the SELECT after a leading empty statement in direct --sql
#### When
```shell
sqly --sql ";SELECT 1 AS x"
```
#### Then
- exit code is `0`
- stdout contains `x`, `1`
### Scenario: runs the SELECT after multiple leading empty statements in direct --sql
#### When
```shell
sqly --sql ";;SELECT 2 AS y"
```
#### Then
- exit code is `0`
- stdout contains `2`
### Scenario: exports the SELECT after a leading empty statement with --output
#### When
```shell
sqly --sql ";SELECT 1 AS x" --output lead.csv
```
#### Then
- exit code is `0`
- file `lead.csv` contains `x`
### Scenario: still rejects ATTACH after a leading empty statement in direct --sql
#### When
```shell
sqly --sql ";ATTACH DATABASE 'x.db' AS aux"
```
#### Then
- exit code is `1`
- stderr contains `ATTACH`
### Scenario: does not rewrite an unchanged CSV when only a TEMP table changed
#### Given
- Fixture file `temp_only.csv` is created.
#### Inputs
_Fixture `temp_only.csv`:_
```text
name,age
alice,30
bob,25
```
_stdin for `sqly`:_
```text
CREATE TEMP TABLE scratch(id INTEGER);
INSERT INTO scratch VALUES (1);
.save --force
```
#### When
```shell
sqly temp_only.csv
```
#### Then
- exit code is `0`
- stdout does not contain `Saved`
- stderr contains `nothing to save`
- file `temp_only.csv` contains `alice,30`
### Scenario: does not fail on an unchanged JSONL import when only a scratch table changed
#### Given
- Fixture file `data.jsonl` is created.
#### Inputs
_Fixture `data.jsonl`:_
```text
{"id":1}
{"id":2}
```
_stdin for `sqly`:_
```text
CREATE TABLE scratch(id INTEGER);
INSERT INTO scratch VALUES (1);
.save --force
```
#### When
```shell
sqly data.jsonl
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `nothing to save`
- stderr does not contain `not loaded from a file`
### Scenario: does not rewrite the source after net-zero CSV edits
#### Given
- Fixture file `netzero.csv` is created.
#### Inputs
_Fixture `netzero.csv`:_
```text
name,age
alice,30
bob,25
```
_stdin for `sqly`:_
```text
UPDATE netzero SET age=99 WHERE name='alice';
UPDATE netzero SET age=30 WHERE name='alice';
.save --force
```
#### When
```shell
sqly netzero.csv
```
#### Then
- exit code is `0`
- stdout does not contain `Saved`
- stderr contains `nothing to save`
- file `netzero.csv` contains `alice,30`
### Scenario: does not rewrite the source after net-zero edits via --sql-file --save --force
#### Given
- Fixture file `netzero_file.csv` is created.
- Fixture file `netzero.sql` is created.
#### Inputs
_Fixture `netzero_file.csv`:_
```text
name,age
alice,30
bob,25
```
_Fixture `netzero.sql`:_
```text
UPDATE netzero_file SET age=99 WHERE name='alice';
UPDATE netzero_file SET age=30 WHERE name='alice';
```
#### When
```shell
sqly --sql-file netzero.sql --save --force netzero_file.csv
```
#### Then
- exit code is `0`
- stderr contains `nothing to save`
- file `netzero_file.csv` contains `alice,30`
### Scenario: still persists a genuine CSV change with .save --force
#### Given
- Fixture file `genuine.csv` is created.
#### Inputs
_Fixture `genuine.csv`:_
```text
name,age
alice,30
bob,25
```
_stdin for `sqly`:_
```text
UPDATE genuine SET age=999 WHERE name='alice';
.save --force
```
#### When
```shell
sqly genuine.csv
```
#### Then
- exit code is `0`
- stdout contains `affected`
- stderr contains `Saved`
- file `genuine.csv` contains `999`
## sqly v0.25.0 binary regressions
Source: `test/e2e/tools/sqly/v0_25_bugs.atago.yaml`
### Scenario: rejects an explicit empty --sql value
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
#### When
```shell
sqly --sql "" user.csv
```
#### Then
- exit code is `1`
- stderr contains `--sql requires a non-empty SQL statement`
### Scenario: reports a hint when non-interactive run gets empty stdin and no file
#### Inputs
_stdin for `sqly`:_
#### When
```shell
sqly
```
#### Then
- exit code is `1`
- stderr contains `no TTY detected`
### Scenario: reports a hint when non-interactive run gets empty stdin with a file
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `no TTY detected`
### Scenario: reports a stable stdin reference instead of the staging temp path
#### Inputs
_stdin for `sqly`:_
#### When
```shell
sqly --stdin csv --sql "SELECT COUNT(*) FROM stdin"
```
#### Then
- exit code is `1`
- stderr contains `stdin`
- stderr does not contain `sqly-stdin-`
### Scenario: fails batch mode when .schema is missing its table name
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.schema
SELECT 1;
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.schema requires`
### Scenario: fails batch mode when .header is missing its table name
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.header
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.header requires`
### Scenario: fails batch mode when .describe is missing its table name
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.describe
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.describe requires`
### Scenario: fails batch mode when .mode is missing its mode name
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.mode
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.mode requires`
### Scenario: fails batch mode when .dump is missing its destination
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.dump
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.dump requires`
### Scenario: fails batch mode when .import is missing its path
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.import
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.import requires`
### Scenario: fails batch mode when .save is missing its argument
#### Given
- Fixture file `user.csv` is created.
#### Inputs
_Fixture `user.csv`:_
```text
user_name,identifier
booker12,1
```
_stdin for `sqly`:_
```text
.save
```
#### When
```shell
sqly user.csv
```
#### Then
- exit code is `1`
- stderr contains `.save requires`
### Scenario: keeps --inspect quiet on stderr after a successful directory import
#### Given
- Fixture file `space dir/d.csv` is created.
#### Inputs
_Fixture `space dir/d.csv`:_
```text
a
1
```
#### When
```shell
sqly --inspect "space dir"
```
#### Then
- exit code is `0`
- stdout contains `"tables"`
- stderr does not contain `Successfully imported`
### Scenario: keeps --profile quiet on stderr after a successful directory import
#### Given
- Fixture file `space dir/d.csv` is created.
#### Inputs
_Fixture `space dir/d.csv`:_
```text
a
1
```
#### When
```shell
sqly --profile "space dir"
```
#### Then
- exit code is `0`
- stdout contains `"tables"`
- stderr does not contain `Successfully imported`
### Scenario: guides "sqly help" to --help instead of an import error
#### When
```shell
sqly help
```
#### Then
- exit code is `1`
- stderr contains `--help`, `no subcommands`
- stderr does not contain `path does not exist`
### Scenario: guides "sqly version" to --version instead of an import error
#### When
```shell
sqly version
```
#### Then
- exit code is `1`
- stderr contains `--version`
- stderr does not contain `path does not exist`