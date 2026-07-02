# atago Behavior Specs
## Summary
29 suites · 246 scenarios
## Contents
- [iso8583tool input auto-detection](#iso8583tool-input-auto-detection) — 4 scenarios
  - [views a raw binary capture without --encoding raw](#scenario-views-a-raw-binary-capture-without---encoding-raw)
  - [validates a raw binary capture without --encoding raw](#scenario-validates-a-raw-binary-capture-without---encoding-raw)
  - [converts a raw binary capture without --encoding raw](#scenario-converts-a-raw-binary-capture-without---encoding-raw)
  - [reads an all-numeric raw ASCII capture as raw, not packed hex](#scenario-reads-an-all-numeric-raw-ascii-capture-as-raw-not-packed-hex)
- [iso8583tool spec87bcd-starter preset](#iso8583tool-spec87bcd-starter-preset) — 6 scenarios
  - [packs and round-trips an EMV TLV tag (55.9F02)](#scenario-packs-and-round-trips-an-emv-tlv-tag-559f02)
  - [packs a raw PIN field (52)](#scenario-packs-a-raw-pin-field-52)
  - [packs and round-trips a raw MAC field (64)](#scenario-packs-and-round-trips-a-raw-mac-field-64)
  - [round-trips a variable-length field with a BCD length prefix (32)](#scenario-round-trips-a-variable-length-field-with-a-bcd-length-prefix-32)
  - [packs a secondary numeric field (71) as packed BCD, not ASCII](#scenario-packs-a-secondary-numeric-field-71-as-packed-bcd-not-ascii)
  - [round-trips secondary numeric fields (74, 99, 100)](#scenario-round-trips-secondary-numeric-fields-74-99-100)
- [iso8583tool convert with a UTF-8 BOM](#iso8583tool-convert-with-a-utf-8-bom) — 6 scenarios
  - [auto-detects BOM-prefixed JSON as a document to pack](#scenario-auto-detects-bom-prefixed-json-as-a-document-to-pack)
  - [packs BOM-prefixed JSON with an explicit --to hex](#scenario-packs-bom-prefixed-json-with-an-explicit---to-hex)
  - [views a BOM-prefixed hex file](#scenario-views-a-bom-prefixed-hex-file)
  - [doctors a BOM-prefixed hex file as hex, not raw](#scenario-doctors-a-bom-prefixed-hex-file-as-hex-not-raw)
  - [validates a BOM-prefixed hex file](#scenario-validates-a-bom-prefixed-hex-file)
  - [converts a BOM-prefixed hex file to JSON](#scenario-converts-a-bom-prefixed-hex-file-to-json)
- [iso8583tool canonical field values](#iso8583tool-canonical-field-values) — 4 scenarios
  - [shows F3 with canonical width in the full describe view](#scenario-shows-f3-with-canonical-width-in-the-full-describe-view)
  - [matches the filtered view for F4](#scenario-matches-the-filtered-view-for-f4)
  - [returns canonical decoded values in JSON for F3](#scenario-returns-canonical-decoded-values-in-json-for-f3)
  - [returns canonical decoded values from validate for F3](#scenario-returns-canonical-decoded-values-from-validate-for-f3)
- [iso8583tool CLI surface](#iso8583tool-cli-surface) — 7 scenarios
  - [root help prints help with no arguments](#scenario-root-help-prints-help-with-no-arguments)
  - [version prints the version](#scenario-version-prints-the-version)
  - [unknown command fails and shows the command list](#scenario-unknown-command-fails-and-shows-the-command-list)
  - [subcommand help describes convert and exits 0](#scenario-subcommand-help-describes-convert-and-exits-0)
  - [subcommand help describes view and lists --filter](#scenario-subcommand-help-describes-view-and-lists---filter)
  - [root flags reject --help with a trailing argument](#scenario-root-flags-reject---help-with-a-trailing-argument)
  - [root flags reject --version with a trailing argument](#scenario-root-flags-reject---version-with-a-trailing-argument)
- [iso8583tool convert](#iso8583tool-convert) — 21 scenarios
  - [auto-detected direction packs a JSON document to hex](#scenario-auto-detected-direction-packs-a-json-document-to-hex)
  - [auto-detected direction unpacks a message to a JSON document](#scenario-auto-detected-direction-unpacks-a-message-to-a-json-document)
  - [--to override forces json output from a message](#scenario---to-override-forces-json-output-from-a-message)
  - [--to override rejects an unknown direction](#scenario---to-override-rejects-an-unknown-direction)
  - [rejects a path present in both fields and binary_fields](#scenario-rejects-a-path-present-in-both-fields-and-binary_fields)
  - [rejects a parent path that also has nested children](#scenario-rejects-a-parent-path-that-also-has-nested-children)
  - [rejects field id 0 (reserved for the MTI)](#scenario-rejects-field-id-0-reserved-for-the-mti)
  - [rejects field id 1 (the bitmap)](#scenario-rejects-field-id-1-the-bitmap)
  - [rejects field id 0 set through binary_fields](#scenario-rejects-field-id-0-set-through-binary_fields)
  - [rejects an out-of-range field id](#scenario-rejects-an-out-of-range-field-id)
  - [rejects a non-numeric field id](#scenario-rejects-a-non-numeric-field-id)
  - [rejects a malformed dotted path](#scenario-rejects-a-malformed-dotted-path)
  - [rejects leading whitespace in a path key](#scenario-rejects-leading-whitespace-in-a-path-key)
  - [rejects a leading-zero duplicate alias (02 vs 2)](#scenario-rejects-a-leading-zero-duplicate-alias-02-vs-2)
  - [rejects a case-different duplicate TLV alias (9f02 vs 9F02)](#scenario-rejects-a-case-different-duplicate-tlv-alias-9f02-vs-9f02)
  - [rejects raw bytes routed to a text field via binary_fields](#scenario-rejects-raw-bytes-routed-to-a-text-field-via-binary_fields)
  - [round-trip is stable through hex -> json -> hex](#scenario-round-trip-is-stable-through-hex---json---hex)
  - [is stable through raw -> json -> raw with the packed-BCD starter preset](#scenario-is-stable-through-raw---json---raw-with-the-packed-bcd-starter-preset)
  - [to a file writes the result and reports it](#scenario-to-a-file-writes-the-result-and-reports-it)
  - [unmasked-output warning is documented in help](#scenario-unmasked-output-warning-is-documented-in-help)
  - [stays byte-clean on stderr when piped (stdout not a TTY)](#scenario-stays-byte-clean-on-stderr-when-piped-stdout-not-a-tty)
- [iso8583tool custom JSON spec import](#iso8583tool-custom-json-spec-import) — 6 scenarios
  - [loads a top-level Hex field and round-trips it](#scenario-loads-a-top-level-hex-field-and-round-trips-it)
  - [loads a Hex TLV subfield](#scenario-loads-a-hex-tlv-subfield)
  - [loads a Track1 field](#scenario-loads-a-track1-field)
  - [loads a Track3 field](#scenario-loads-a-track3-field)
  - [loads an IndexTag composite subfield](#scenario-loads-an-indextag-composite-subfield)
  - [loads a composite tag that omits sort](#scenario-loads-a-composite-tag-that-omits-sort)
- [iso8583tool detection messaging](#iso8583tool-detection-messaging) — 3 scenarios
  - [presents tied presets for an ambiguous message](#scenario-presents-tied-presets-for-an-ambiguous-message)
  - [flags a truncated capture instead of only "custom layout"](#scenario-flags-a-truncated-capture-instead-of-only-custom-layout)
  - [validate calls out a truncated capture rather than doctor](#scenario-validate-calls-out-a-truncated-capture-rather-than-doctor)
- [iso8583tool diff](#iso8583tool-diff) — 11 scenarios
  - [reports changed fields in text form](#scenario-reports-changed-fields-in-text-form)
  - [emits jq-compatible JSON](#scenario-emits-jq-compatible-json)
  - [reports no differences for identical messages](#scenario-reports-no-differences-for-identical-messages)
  - [filters to a field subtree](#scenario-filters-to-a-field-subtree)
  - [reads one side from stdin](#scenario-reads-one-side-from-stdin)
  - [rejects two stdin sides](#scenario-rejects-two-stdin-sides)
  - [masks track data by default](#scenario-masks-track-data-by-default)
  - [reveals raw values with --unsafe](#scenario-reveals-raw-values-with---unsafe)
  - [rejects an unknown --format value](#scenario-rejects-an-unknown---format-value)
  - [private-field safety - masks an embedded PAN by default](#scenario-private-field-safety---masks-an-embedded-pan-by-default)
  - [private-field safety - reveals the embedded PAN with --unsafe](#scenario-private-field-safety---reveals-the-embedded-pan-with---unsafe)
- [iso8583tool doctor](#iso8583tool-doctor) — 10 scenarios
  - [recommends the BASE I starter for an ASCII BASE I message](#scenario-recommends-the-base-i-starter-for-an-ascii-base-i-message)
  - [detects a packed-BCD raw message](#scenario-detects-a-packed-bcd-raw-message)
  - [auto-detects a raw .bin without --encoding](#scenario-auto-detects-a-raw-bin-without---encoding)
  - [emits a JSON report with --format json](#scenario-emits-a-json-report-with---format-json)
  - [exits non-zero when no preset fits](#scenario-exits-non-zero-when-no-preset-fits)
  - [is suggested by a wrong-spec validate failure](#scenario-is-suggested-by-a-wrong-spec-validate-failure)
  - [marks every tied preset recommended and confirms with each](#scenario-marks-every-tied-preset-recommended-and-confirms-with-each)
  - [explains how to choose between the tied basei-starter and spec87ascii presets](#scenario-explains-how-to-choose-between-the-tied-basei-starter-and-spec87ascii-presets)
  - [shell-safe confirm hint quotes a path that contains a space](#scenario-shell-safe-confirm-hint-quotes-a-path-that-contains-a-space)
  - [custom-spec validate hint does not steer a custom-spec failure to doctor](#scenario-custom-spec-validate-hint-does-not-steer-a-custom-spec-failure-to-doctor)
- [iso8583tool edge cases](#iso8583tool-edge-cases) — 13 scenarios
  - [rejects non-hex characters under --encoding hex](#scenario-rejects-non-hex-characters-under---encoding-hex)
  - [rejects odd-length hex](#scenario-rejects-odd-length-hex)
  - [reports the failing field for a truncated message](#scenario-reports-the-failing-field-for-a-truncated-message)
  - [fails on empty inline input](#scenario-fails-on-empty-inline-input)
  - [fails when the file does not exist](#scenario-fails-when-the-file-does-not-exist)
  - [fails when a directory is passed instead of a file](#scenario-fails-when-a-directory-is-passed-instead-of-a-file)
  - [refuses both a file argument and --raw](#scenario-refuses-both-a-file-argument-and---raw)
  - [fails to pack a document with no mti](#scenario-fails-to-pack-a-document-with-no-mti)
  - [fails to pack a document with an invalid TLV tag](#scenario-fails-to-pack-a-document-with-an-invalid-tlv-tag)
  - [oversized input - rejects an oversized file with a clear limit error](#scenario-oversized-input---rejects-an-oversized-file-with-a-clear-limit-error)
  - [oversized input - rejects oversized stdin with a clear limit error](#scenario-oversized-input---rejects-oversized-stdin-with-a-clear-limit-error)
  - [oversized input - cleanly fails a truncated field 55 TLV without panic](#scenario-oversized-input---cleanly-fails-a-truncated-field-55-tlv-without-panic)
  - [oversized input - does not panic on a length-spoofed variable-length field](#scenario-oversized-input---does-not-panic-on-a-length-spoofed-variable-length-field)
- [iso8583tool ergonomics](#iso8583tool-ergonomics) — 10 scenarios
  - [flag ordering - accepts the target after the flags](#scenario-flag-ordering---accepts-the-target-after-the-flags)
  - [flag ordering - accepts the target before the flags](#scenario-flag-ordering---accepts-the-target-before-the-flags)
  - [flag ordering - accepts flags interleaved around the target](#scenario-flag-ordering---accepts-flags-interleaved-around-the-target)
  - [color - is plain by default when not on a terminal](#scenario-color---is-plain-by-default-when-not-on-a-terminal)
  - [color - forces color with --color always](#scenario-color---forces-color-with---color-always)
  - [color - stays plain with --no-color even when forced elsewhere](#scenario-color---stays-plain-with---no-color-even-when-forced-elsewhere)
  - [color - rejects an unknown --color value instead of ignoring it](#scenario-color---rejects-an-unknown---color-value-instead-of-ignoring-it)
  - [end of options - treats a dash-leading filename after -- as a positional](#scenario-end-of-options---treats-a-dash-leading-filename-after----as-a-positional)
  - [config - applies an extension catalog from --config](#scenario-config---applies-an-extension-catalog-from---config)
  - [config - fails on a config with an invalid strategy](#scenario-config---fails-on-a-config-with-an-invalid-strategy)
- [iso8583tool extension strategy](#iso8583tool-extension-strategy) — 3 scenarios
  - [a custom positional composite spec does not apply the BASE I catalog](#scenario-a-custom-positional-composite-spec-does-not-apply-the-base-i-catalog)
  - [a built-in plain field documented as bitmap reports field 127 as opaque, matching the spec](#scenario-a-built-in-plain-field-documented-as-bitmap-reports-field-127-as-opaque-matching-the-spec)
  - [explains a dot-path set on a plain built-in field](#scenario-explains-a-dot-path-set-on-a-plain-built-in-field)
- [iso8583tool convert field count](#iso8583tool-convert-field-count) — 1 scenario
  - [reports the top-level field count matching doctor](#scenario-reports-the-top-level-field-count-matching-doctor)
- [iso8583tool filter normalization](#iso8583tool-filter-normalization) — 5 scenarios
  - [matches a lowercase EMV tag in view](#scenario-matches-a-lowercase-emv-tag-in-view)
  - [matches a lowercase EMV tag in diff](#scenario-matches-a-lowercase-emv-tag-in-diff)
  - [accepts "0" as an MTI alias in diff](#scenario-accepts-0-as-an-mti-alias-in-diff)
  - [reports an unmatched diff filter](#scenario-reports-an-unmatched-diff-filter)
  - [reports an unmatched filter in JSON](#scenario-reports-an-unmatched-filter-in-json)
- [iso8583tool sensitive-data masking](#iso8583tool-sensitive-data-masking) — 11 scenarios
  - [does not mask a non-PAN business identifier](#scenario-does-not-mask-a-non-pan-business-identifier)
  - [masks a dash-separated PAN](#scenario-masks-a-dash-separated-pan)
  - [masks a space-separated PAN](#scenario-masks-a-space-separated-pan)
  - [masks a PAN embedded in a non-private free-form field](#scenario-masks-a-pan-embedded-in-a-non-private-free-form-field)
  - [masks the extended PAN field 34](#scenario-masks-the-extended-pan-field-34)
  - [does not mask the country code field 20](#scenario-does-not-mask-the-country-code-field-20)
  - [shows the raw field 20 change in diff](#scenario-shows-the-raw-field-20-change-in-diff)
  - [masks a whole free-form track, not just its PAN](#scenario-masks-a-whole-free-form-track-not-just-its-pan)
  - [masks an underscore-labeled PAN (card_no)](#scenario-masks-an-underscore-labeled-pan-card_no)
  - [masks a spaced-label PAN (card number)](#scenario-masks-a-spaced-label-pan-card-number)
  - [a custom positional composite with a PAN-numbered subfield does not mask subfield 48.2 with the top-level PAN rule](#scenario-a-custom-positional-composite-with-a-pan-numbered-subfield-does-not-mask-subfield-482-with-the-top-level-pan-rule)
- [iso8583tool masking under custom specs](#iso8583tool-masking-under-custom-specs) — 8 scenarios
  - [masks a PAN in a binary field 63](#scenario-masks-a-pan-in-a-binary-field-63)
  - [masks a known 9F6B track2-equivalent tag in view](#scenario-masks-a-known-9f6b-track2-equivalent-tag-in-view)
  - [masks a known 9F6B track2-equivalent tag in redact](#scenario-masks-a-known-9f6b-track2-equivalent-tag-in-redact)
  - [masks a track2-equivalent tag in a non-55 container (127.57)](#scenario-masks-a-track2-equivalent-tag-in-a-non-55-container-12757)
  - [masks a sensitive tag nested in a constructed TLV (55.70.57)](#scenario-masks-a-sensitive-tag-nested-in-a-constructed-tlv-557057)
  - [does not over-mask a harmless custom field 35](#scenario-does-not-over-mask-a-harmless-custom-field-35)
  - [does not over-mask a harmless custom field 52](#scenario-does-not-over-mask-a-harmless-custom-field-52)
  - [still masks a real PAN in a custom field 2](#scenario-still-masks-a-real-pan-in-a-custom-field-2)
- [iso8583tool view nested composites](#iso8583tool-view-nested-composites) — 3 scenarios
  - [nested positional composite keeps the nested positional path](#scenario-nested-positional-composite-keeps-the-nested-positional-path)
  - [nested EMV tag annotation annotates a nested ARC tag as Approved](#scenario-nested-emv-tag-annotation-annotates-a-nested-arc-tag-as-approved)
  - [nested EMV tag annotation decodes nested leaf tags in validate --format json](#scenario-nested-emv-tag-annotation-decodes-nested-leaf-tags-in-validate---format-json)
- [iso8583tool nested TLV](#iso8583tool-nested-tlv) — 5 scenarios
  - [unpacks a nested TLV to its leaf path](#scenario-unpacks-a-nested-tlv-to-its-leaf-path)
  - [selects a nested TLV leaf with --filter](#scenario-selects-a-nested-tlv-leaf-with---filter)
  - [diffs at the nested TLV leaf tag](#scenario-diffs-at-the-nested-tlv-leaf-tag)
  - [keeps a top-level tag and a nested tag set on the same field](#scenario-keeps-a-top-level-tag-and-a-nested-tag-set-on-the-same-field)
  - [shows the full nested path in the describe output](#scenario-shows-the-full-nested-path-in-the-describe-output)
- [iso8583tool workflows](#iso8583tool-workflows) — 4 scenarios
  - [streams sample -> convert -> view](#scenario-streams-sample---convert---view)
  - [editing an EMV tag edits one tag and packs it back](#scenario-editing-an-emv-tag-edits-one-tag-and-packs-it-back)
  - [editing an EMV tag keeps an unknown tag through the round trip](#scenario-editing-an-emv-tag-keeps-an-unknown-tag-through-the-round-trip)
  - [extracts a single field value with --filter](#scenario-extracts-a-single-field-value-with---filter)
- [iso8583tool README examples](#iso8583tool-readme-examples) — 33 scenarios
  - [quick start - lists the bundled samples](#scenario-quick-start---lists-the-bundled-samples)
  - [quick start - views the BASE I auth response](#scenario-quick-start---views-the-base-i-auth-response)
  - [quick start - validates the unknown-TLV sample](#scenario-quick-start---validates-the-unknown-tlv-sample)
  - [quick start - converts the BASE I request to JSON](#scenario-quick-start---converts-the-base-i-request-to-json)
  - [view - shows JSON output](#scenario-view---shows-json-output)
  - [view - filters the requested fields](#scenario-view---filters-the-requested-fields)
  - [view - reads a message from stdin](#scenario-view---reads-a-message-from-stdin)
  - [view - is jq-compatible for fields](#scenario-view---is-jq-compatible-for-fields)
  - [diff - compares a request and a response](#scenario-diff---compares-a-request-and-a-response)
  - [diff - is jq-compatible for changes](#scenario-diff---is-jq-compatible-for-changes)
  - [redact - masks the PAN for safe sharing](#scenario-redact---masks-the-pan-for-safe-sharing)
  - [redact - supports a text format](#scenario-redact---supports-a-text-format)
  - [convert - packs the BASE I request to hex](#scenario-convert---packs-the-base-i-request-to-hex)
  - [convert - converts a sample through stdin](#scenario-convert---converts-a-sample-through-stdin)
  - [convert - writes converted output to a file](#scenario-convert---writes-converted-output-to-a-file)
  - [validate - reports a broken inline message as an error](#scenario-validate---reports-a-broken-inline-message-as-an-error)
  - [validate - emits JSON when asked](#scenario-validate---emits-json-when-asked)
  - [doctor - recommends a preset for the BASE I sample](#scenario-doctor---recommends-a-preset-for-the-base-i-sample)
  - [doctor - is jq-compatible for the recommendation](#scenario-doctor---is-jq-compatible-for-the-recommendation)
  - [specs - lists the presets](#scenario-specs---lists-the-presets)
  - [specs - is jq-compatible for preset names](#scenario-specs---is-jq-compatible-for-preset-names)
  - [sample - prints a sample as JSON](#scenario-sample---prints-a-sample-as-json)
  - [sample - writes a sample as hex](#scenario-sample---writes-a-sample-as-hex)
  - [send (default 2byte-binary framing) - sends a packed 0800 and decodes the 0810 reply](#scenario-send-default-2byte-binary-framing---sends-a-packed-0800-and-decodes-the-0810-reply)
  - [send (default 2byte-binary framing) - reads the message from stdin and is jq-compatible for the response MTI](#scenario-send-default-2byte-binary-framing---reads-the-message-from-stdin-and-is-jq-compatible-for-the-response-mti)
  - [send (default 2byte-binary framing) - asserts the reply with --expect-mti / --expect-field (no jq needed)](#scenario-send-default-2byte-binary-framing---asserts-the-reply-with---expect-mti----expect-field-no-jq-needed)
  - [send (default 2byte-binary framing) - accepts an inline message via --raw](#scenario-send-default-2byte-binary-framing---accepts-an-inline-message-via---raw)
  - [send (4-digit ASCII framing) - packs a JSON document and sends it with a 4-digit header](#scenario-send-4-digit-ascii-framing---packs-a-json-document-and-sends-it-with-a-4-digit-header)
  - [unknown TLV round-trip - preserves the unknown tag when unpacking and packing again](#scenario-unknown-tlv-round-trip---preserves-the-unknown-tag-when-unpacking-and-packing-again)
  - [other specs - validates the spec87ascii sample](#scenario-other-specs---validates-the-spec87ascii-sample)
  - [other specs - strict-validates the spec87ascii sample under its intended preset](#scenario-other-specs---strict-validates-the-spec87ascii-sample-under-its-intended-preset)
  - [other specs - views the spec87ascii sample](#scenario-other-specs---views-the-spec87ascii-sample)
  - [other specs - converts the spec87ascii sample to JSON](#scenario-other-specs---converts-the-spec87ascii-sample-to-json)
- [iso8583tool redact](#iso8583tool-redact) — 9 scenarios
  - [masks the PAN in JSON output](#scenario-masks-the-pan-in-json-output)
  - [never leaks the full PAN](#scenario-never-leaks-the-full-pan)
  - [fully masks the EMV application cryptogram](#scenario-fully-masks-the-emv-application-cryptogram)
  - [supports a human-readable text format](#scenario-supports-a-human-readable-text-format)
  - [orders text output by MTI then numeric field id](#scenario-orders-text-output-by-mti-then-numeric-field-id)
  - [reads from stdin for a Slack-safe pipe](#scenario-reads-from-stdin-for-a-slack-safe-pipe)
  - [masks a PAN embedded in a free-form private field (F63)](#scenario-masks-a-pan-embedded-in-a-free-form-private-field-f63)
  - [auto-detected input encoding redacts a raw binary capture without --encoding](#scenario-auto-detected-input-encoding-redacts-a-raw-binary-capture-without---encoding)
  - [auto-detected input encoding still masks the PAN in a raw binary capture](#scenario-auto-detected-input-encoding-still-masks-the-pan-in-a-raw-binary-capture)
- [iso8583tool control-byte sanitization](#iso8583tool-control-byte-sanitization) — 4 scenarios
  - [view escapes control bytes](#scenario-view-escapes-control-bytes)
  - [validate escapes control bytes](#scenario-validate-escapes-control-bytes)
  - [diff escapes control bytes](#scenario-diff-escapes-control-bytes)
  - [redact text escapes control bytes](#scenario-redact-text-escapes-control-bytes)
- [iso8583tool send](#iso8583tool-send) — 22 scenarios
  - [sends an 0800 and decodes the 0810 response (2byte-binary)](#scenario-sends-an-0800-and-decodes-the-0810-response-2byte-binary)
  - [lists every response field, not only annotated codes](#scenario-lists-every-response-field-not-only-annotated-codes)
  - [packs a JSON document and sends it (2byte-binary)](#scenario-packs-a-json-document-and-sends-it-2byte-binary)
  - [reads the message from stdin via -](#scenario-reads-the-message-from-stdin-via--)
  - [frames with a 4-digit ASCII length header](#scenario-frames-with-a-4-digit-ascii-length-header)
  - [sends with no length header and reads the reply until EOF](#scenario-sends-with-no-length-header-and-reads-the-reply-until-eof)
  - [decodes the response in describe output (none framing)](#scenario-decodes-the-response-in-describe-output-none-framing)
  - [exits non-zero with a clear error when the response times out](#scenario-exits-non-zero-with-a-clear-error-when-the-response-times-out)
  - [exits non-zero when a none-framing peer never replies](#scenario-exits-non-zero-when-a-none-framing-peer-never-replies)
  - [passes when --expect-mti and --expect-field match the response](#scenario-passes-when---expect-mti-and---expect-field-match-the-response)
  - [exits non-zero with a deterministic error on an MTI mismatch](#scenario-exits-non-zero-with-a-deterministic-error-on-an-mti-mismatch)
  - [exits non-zero when an expected field value differs](#scenario-exits-non-zero-when-an-expected-field-value-differs)
  - [rejects an --expect-field without PATH=VALUE](#scenario-rejects-an---expect-field-without-pathvalue)
  - [frames and prints the request without connecting](#scenario-frames-and-prints-the-request-without-connecting)
  - [emits a machine-readable dry-run record](#scenario-emits-a-machine-readable-dry-run-record)
  - [withholds the framed bytes by default](#scenario-withholds-the-framed-bytes-by-default)
  - [reveals the framed wire bytes under --unsafe](#scenario-reveals-the-framed-wire-bytes-under---unsafe)
  - [includes framed_hex in JSON only under --unsafe](#scenario-includes-framed_hex-in-json-only-under---unsafe)
  - [rejects expectations because there is no response to assert](#scenario-rejects-expectations-because-there-is-no-response-to-assert)
  - [rejects an invalid --framing value](#scenario-rejects-an-invalid---framing-value)
  - [rejects a HOST:PORT without a port](#scenario-rejects-a-hostport-without-a-port)
  - [prints usage for send --help](#scenario-prints-usage-for-send---help)
- [iso8583tool specs](#iso8583tool-specs) — 3 scenarios
  - [lists the built-in presets with the default marked](#scenario-lists-the-built-in-presets-with-the-default-marked)
  - [emits a JSON array with --format json](#scenario-emits-a-json-array-with---format-json)
  - [rejects an unexpected positional argument](#scenario-rejects-an-unexpected-positional-argument)
- [iso8583tool standard high-numbered fields](#iso8583tool-standard-high-numbered-fields) — 3 scenarios
  - [packs and round-trips fields 95/96/100/102/103/104](#scenario-packs-and-round-trips-fields-9596100102103104)
  - [packs the reserved fields 123-127](#scenario-packs-the-reserved-fields-123-127)
  - [packs and round-trips the binary MAC field 128](#scenario-packs-and-round-trips-the-binary-mac-field-128)
- [iso8583tool validate --strict advice and network rules](#iso8583tool-validate---strict-advice-and-network-rules) — 12 scenarios
  - [fails a hollow authorization advice (0120)](#scenario-fails-a-hollow-authorization-advice-0120)
  - [fails a hollow financial advice (0220)](#scenario-fails-a-hollow-financial-advice-0220)
  - [fails a hollow network advice (0820)](#scenario-fails-a-hollow-network-advice-0820)
  - [fails a hollow network response (0810)](#scenario-fails-a-hollow-network-response-0810)
  - [fails a hollow network advice response (0830)](#scenario-fails-a-hollow-network-advice-response-0830)
  - [still accepts the bundled network echo under --strict](#scenario-still-accepts-the-bundled-network-echo-under---strict)
  - [fails a hollow authorization notification (0140)](#scenario-fails-a-hollow-authorization-notification-0140)
  - [fails a hollow financial instruction ack (0270)](#scenario-fails-a-hollow-financial-instruction-ack-0270)
  - [fails a hollow file-action request (0300)](#scenario-fails-a-hollow-file-action-request-0300)
  - [requires a PAN source for a reversal request (0400)](#scenario-requires-a-pan-source-for-a-reversal-request-0400)
  - [warns that reconciliation (0500) rules are not implemented](#scenario-warns-that-reconciliation-0500-rules-are-not-implemented)
  - [rejects an alphabetic value in a numeric field (70)](#scenario-rejects-an-alphabetic-value-in-a-numeric-field-70)
- [iso8583tool validate](#iso8583tool-validate) — 7 scenarios
  - [passes a good message with exit 0](#scenario-passes-a-good-message-with-exit-0)
  - [reports unknown TLV tags as a warning but still exits 0](#scenario-reports-unknown-tlv-tags-as-a-warning-but-still-exits-0)
  - [fails a broken message with exit 1 and names the field](#scenario-fails-a-broken-message-with-exit-1-and-names-the-field)
  - [emits a JSON report with --format json](#scenario-emits-a-json-report-with---format-json-1)
  - [accepts a complete sample under --strict](#scenario-accepts-a-complete-sample-under---strict)
  - [flags a hollow response under --strict](#scenario-flags-a-hollow-response-under---strict)
  - [omits the Decoded Fields heading when only the MTI decoded](#scenario-omits-the-decoded-fields-heading-when-only-the-mti-decoded)
- [iso8583tool view](#iso8583tool-view) — 12 scenarios
  - [describe output - decodes codes and prints a summary](#scenario-describe-output---decodes-codes-and-prints-a-summary)
  - [describe output - masks the PAN](#scenario-describe-output---masks-the-pan)
  - [json output - emits a decoded array and stays uncolored](#scenario-json-output---emits-a-decoded-array-and-stays-uncolored)
  - [--filter prints only the requested fields](#scenario---filter-prints-only-the-requested-fields)
  - [--filter marks a field that is not present](#scenario---filter-marks-a-field-that-is-not-present)
  - [--filter emits object-shaped JSON with an explicit missing_filters list](#scenario---filter-emits-object-shaped-json-with-an-explicit-missing_filters-list)
  - [--filter always emits missing_filters as an array even when nothing is missing](#scenario---filter-always-emits-missing_filters-as-an-array-even-when-nothing-is-missing)
  - [stdin - reads a message piped in via -](#scenario-stdin---reads-a-message-piped-in-via--)
  - [stdin - reads from stdin when the target is omitted](#scenario-stdin---reads-from-stdin-when-the-target-is-omitted)
  - [raw binary + packed BCD - views a kanmu-like raw message with the packed-BCD starter preset](#scenario-raw-binary--packed-bcd---views-a-kanmu-like-raw-message-with-the-packed-bcd-starter-preset)
  - [private-field safety - masks a PAN embedded in a free-form private field by default](#scenario-private-field-safety---masks-a-pan-embedded-in-a-free-form-private-field-by-default)
  - [private-field safety - reveals the raw private-field value with --unsafe](#scenario-private-field-safety---reveals-the-raw-private-field-value-with---unsafe)
## iso8583tool input auto-detection
Source: `test/e2e/tools/iso8583tool/autodetect.atago.yaml`
### Scenario: views a raw binary capture without --encoding raw
#### Given
- Fixture file `raw.bin` is created.
#### When
```shell
iso8583tool view raw.bin --spec spec87bcd-starter
```
#### Then
- exit code is `0`
- stdout does not contain `decode hex`
- stdout contains `MTI`
### Scenario: validates a raw binary capture without --encoding raw
#### Given
- Fixture file `raw.bin` is created.
#### When
```shell
iso8583tool validate raw.bin --spec spec87bcd-starter
```
#### Then
- exit code is `0`
- stdout does not contain `decode hex`
### Scenario: converts a raw binary capture without --encoding raw
#### Given
- Fixture file `raw.bin` is created.
#### When
```shell
iso8583tool convert raw.bin --spec spec87bcd-starter
```
#### Then
- exit code is `0`
- stdout does not contain `decode hex`
- stdout contains `"mti"`
### Scenario: reads an all-numeric raw ASCII capture as raw, not packed hex
#### Given
- Fixture file `num.bin` is created.
#### When
```shell
iso8583tool view num.bin
```
#### Then
- exit code is `0`
- stdout does not contain `not enough data`
- stdout contains `0800`
## iso8583tool spec87bcd-starter preset
Source: `test/e2e/tools/iso8583tool/bcd_starter.atago.yaml`
### Scenario: packs and round-trips an EMV TLV tag (55.9F02)
#### Given
- Fixture file `emv.json` is created.
#### Inputs
_Fixture `emv.json`:_
```
{"mti":"0100","fields":{"2":"4019249999999999","3":"000000","4":"000000001000","7":"0605123456","11":"123456","41":"TERMID01"},"binary_fields":{"55.9F02":"000000001000"}}
```
#### When
```shell
iso8583tool convert emv.json --to hex --encoding hex --spec spec87bcd-starter > emv.hex
iso8583tool view emv.hex --encoding hex --spec spec87bcd-starter --format json

```
#### Then
- exit code is `0`
- stdout contains `"55.9F02"`, `000000001000`
### Scenario: packs a raw PIN field (52)
#### Given
- Fixture file `pin.json` is created.
#### Inputs
_Fixture `pin.json`:_
```
{"mti":"0100","fields":{"2":"4019249999999999","3":"000000","4":"000000001000","7":"0605123456","11":"123456","41":"TERMID01"},"binary_fields":{"52":"A1B2C3D4E5F60708"}}
```
#### When
```shell
iso8583tool convert pin.json --to hex --encoding hex --spec spec87bcd-starter
```
#### Then
- exit code is `0`
- stdout does not contain `should be fixed`
### Scenario: packs and round-trips a raw MAC field (64)
#### Given
- Fixture file `mac.json` is created.
#### Inputs
_Fixture `mac.json`:_
```
{"mti":"0100","fields":{"2":"4019249999999999","3":"000000","4":"000000001000","7":"0605123456","11":"123456","41":"TERMID01"},"binary_fields":{"64":"A1B2C3D4E5F60708"}}
```
#### When
```shell
iso8583tool convert mac.json --to hex --encoding hex --spec spec87bcd-starter > mac.hex
iso8583tool view mac.hex --encoding hex --spec spec87bcd-starter --unsafe --format json

```
#### Then
- exit code is `0`
- stdout contains `A1B2C3D4E5F60708`
### Scenario: round-trips a variable-length field with a BCD length prefix (32)
#### Given
- Fixture file `f32.json` is created.
#### Inputs
_Fixture `f32.json`:_
```
{"mti":"0100","fields":{"2":"4019249999999999","3":"000000","4":"000000001000","7":"0605123456","11":"123456","32":"123456","41":"TERMID01"}}
```
#### When
```shell
iso8583tool convert f32.json --to hex --encoding hex --spec spec87bcd-starter > f32.hex
iso8583tool view f32.hex --encoding hex --spec spec87bcd-starter --format json

```
#### Then
- exit code is `0`
- stdout contains `"32": "123456"`
### Scenario: packs a secondary numeric field (71) as packed BCD, not ASCII
#### Given
- Fixture file `f71.json` is created.
#### Inputs
_Fixture `f71.json`:_
```
{"mti":"0800","fields":{"11":"123456","70":"301","71":"1234"}}
```
#### When
```shell
iso8583tool convert f71.json --to hex --spec spec87bcd-starter
```
#### Then
- exit code is `0`
- stdout contains `1234`
- stdout does not contain `31323334`
### Scenario: round-trips secondary numeric fields (74, 99, 100)
#### Given
- Fixture file `sec.json` is created.
#### Inputs
_Fixture `sec.json`:_
```
{"mti":"0800","fields":{"11":"123456","70":"301","74":"0000000001","99":"12345678901","100":"98765432109"}}
```
#### When
```shell
iso8583tool convert sec.json --to hex --spec spec87bcd-starter > sec.hex
iso8583tool convert sec.hex --spec spec87bcd-starter --encoding hex --to json

```
#### Then
- exit code is `0`
- stdout contains `"74": "0000000001"`, `"99": "12345678901"`, `"100": "98765432109"`
## iso8583tool convert with a UTF-8 BOM
Source: `test/e2e/tools/iso8583tool/bom.atago.yaml`
### Scenario: auto-detects BOM-prefixed JSON as a document to pack
#### Given
- Fixture file `bom.json` is created.
#### When
```shell
iso8583tool convert bom.json
```
#### Then
- exit code is `0`
- stdout does not contain `invalid byte`, `decode hex`
### Scenario: packs BOM-prefixed JSON with an explicit --to hex
#### Given
- Fixture file `bom.json` is created.
#### When
```shell
iso8583tool convert bom.json --to hex
```
#### Then
- exit code is `0`
- stdout does not contain `invalid character`
### Scenario: views a BOM-prefixed hex file
#### When
```shell
{ printf '\357\273\277'; cat "$ISO_EXAMPLES/basei/0110-auth-response.hex"; } > bom.hex

iso8583tool view bom.hex --no-color
```
#### Then
- after `iso8583tool view bom.hex --no-color`:
  - exit code is `0`
  - stdout contains `0110`
### Scenario: doctors a BOM-prefixed hex file as hex, not raw
#### When
```shell
{ printf '\357\273\277'; cat "$ISO_EXAMPLES/basei/0110-auth-response.hex"; } > bom.hex

iso8583tool doctor bom.hex --no-color
```
#### Then
- after `iso8583tool doctor bom.hex --no-color`:
  - exit code is `0`
  - stdout contains `hex input`
### Scenario: validates a BOM-prefixed hex file
#### When
```shell
{ printf '\357\273\277'; cat "$ISO_EXAMPLES/basei/0110-auth-response.hex"; } > bom.hex

iso8583tool validate bom.hex --no-color
```
#### Then
- after `iso8583tool validate bom.hex --no-color`:
  - exit code is `0`
  - stdout contains `Validation: ok`
### Scenario: converts a BOM-prefixed hex file to JSON
#### When
```shell
{ printf '\357\273\277'; cat "$ISO_EXAMPLES/basei/0110-auth-response.hex"; } > bom.hex

iso8583tool convert bom.hex --to json
```
#### Then
- after `iso8583tool convert bom.hex --to json`:
  - exit code is `0`
  - stdout contains `"mti"`
## iso8583tool canonical field values
Source: `test/e2e/tools/iso8583tool/canonical.atago.yaml`
### Scenario: shows F3 with canonical width in the full describe view
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --no-color
```
#### Then
- exit code is `0`
- stdout is not empty
- stdout contains `Processing Code`
- stdout matches `/F3.*: 000000/`
### Scenario: matches the filtered view for F4
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 4 --no-color
```
#### Then
- exit code is `0`
- stdout contains `000000005000`
### Scenario: returns canonical decoded values in JSON for F3
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"3": "000000"`, `"value": "000000"`
### Scenario: returns canonical decoded values from validate for F3
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"value": "000000"`
## iso8583tool CLI surface
Source: `test/e2e/tools/iso8583tool/cli.atago.yaml`
### Scenario: root help prints help with no arguments
#### When
```shell
iso8583tool
```
#### Then
- exit code is `0`
- stdout contains `Commands:`, `view`, `convert`, `validate`
### Scenario: version prints the version
#### When
```shell
iso8583tool version
```
#### Then
- exit code is `0`
- stdout contains `iso8583tool`
### Scenario: unknown command fails and shows the command list
#### When
```shell
iso8583tool frobnicate
```
#### Then
- exit code is not `0`
- stderr contains `unknown command`, `Commands:`
### Scenario: subcommand help describes convert and exits 0
#### When
```shell
iso8583tool help convert
```
#### Then
- exit code is `0`
- stdout contains `Usage: iso8583tool convert`, `--to`
### Scenario: subcommand help describes view and lists --filter
#### When
```shell
iso8583tool view --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: iso8583tool view`, `--filter`
- stderr equals an exact value
### Scenario: root flags reject --help with a trailing argument
#### When
```shell
iso8583tool --help view
```
#### Then
- exit code is not `0`
- stderr contains `takes no arguments`
### Scenario: root flags reject --version with a trailing argument
#### When
```shell
iso8583tool --version view
```
#### Then
- exit code is not `0`
- stderr contains `takes no arguments`
## iso8583tool convert
Source: `test/e2e/tools/iso8583tool/convert.atago.yaml`
### Scenario: auto-detected direction packs a JSON document to hex
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.json
```
#### Then
- exit code is `0`
- stdout matches `/^3031/`
### Scenario: auto-detected direction unpacks a message to a JSON document
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.hex
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0100"`, `"55.9F02"`
### Scenario: --to override forces json output from a message
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0110-auth-response.hex --to json
```
#### Then
- exit code is `0`
- stdout contains `"mti"`
### Scenario: --to override rejects an unknown direction
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.json --to sideways
```
#### Then
- exit code is not `0`
- stderr contains `unsupported --to`
### Scenario: rejects a path present in both fields and binary_fields
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{"55.8A":"00"},"binary_fields":{"55.8A":"3035"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `55.8A`
### Scenario: rejects a parent path that also has nested children
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","binary_fields":{"55":"9F0206000000005000","55.9F02":"000000009999"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `55.9F02`
### Scenario: rejects field id 0 (reserved for the MTI)
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{"0":"9999"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `mti`
### Scenario: rejects field id 1 (the bitmap)
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{"1":"1234"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `bitmap`
### Scenario: rejects field id 0 set through binary_fields
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","binary_fields":{"0":"31323334"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `mti`
### Scenario: rejects an out-of-range field id
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{"129":"x"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `invalid field id`
### Scenario: rejects a non-numeric field id
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{"A.1":"x"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `invalid field id`
### Scenario: rejects a malformed dotted path
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","binary_fields":{"55..9F02":"00"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `empty segment`
### Scenario: rejects leading whitespace in a path key
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{" 2":"4111111111111111"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `whitespace`
### Scenario: rejects a leading-zero duplicate alias (02 vs 2)
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","fields":{"02":"4111111111111111","2":"4222222222222222"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `address field`
### Scenario: rejects a case-different duplicate TLV alias (9f02 vs 9F02)
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","binary_fields":{"55.9f02":"000000001000","55.9F02":"000000005000"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `address field`
### Scenario: rejects raw bytes routed to a text field via binary_fields
#### Inputs
_stdin for `iso8583tool`:_
```
{"mti":"0100","binary_fields":{"11":"000102030405"}}
```
#### When
```shell
iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `text field`
### Scenario: round-trip is stable through hex -> json -> hex
#### When
```shell
h="$(iso8583tool sample 0100-auth-request --format hex)"
back="$(printf "%s" "$h" | iso8583tool convert | iso8583tool convert)"
[ "$(printf "%s" "$h")" = "$(printf "%s" "$back")" ] && echo SAME || echo DIFF

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: is stable through raw -> json -> raw with the packed-BCD starter preset
#### Given
- Fixture file `message.bin` is created.
#### When
```shell
iso8583tool convert message.bin --encoding raw --spec spec87bcd-starter > doc.json &&
iso8583tool convert doc.json --to hex --encoding raw --spec spec87bcd-starter --output back.bin >/dev/null &&
cmp -s message.bin back.bin &&
echo SAME

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: to a file writes the result and reports it
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.json --output out.hex
```
#### Then
- exit code is `0`
- stdout contains `Converted with`, `unmasked`, `sensitive`
- file `out.hex` exists
#### Generated artifacts
- `out.hex`
### Scenario: unmasked-output warning is documented in help
#### When
```shell
iso8583tool convert --help
```
#### Then
- exit code is `0`
- stdout contains `UNMASKED`, `redact`
### Scenario: stays byte-clean on stderr when piped (stdout not a TTY)
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.hex
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0100"`
- stderr equals an exact value
## iso8583tool custom JSON spec import
Source: `test/e2e/tools/iso8583tool/custom_spec.atago.yaml`
### Scenario: loads a top-level Hex field and round-trips it
#### Given
- Fixture file `hex-top.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `hex-top.json`:_
```
{"name":"Hex top","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"52":{"type":"Hex","length":8,"description":"PIN Data","enc":"Binary","prefix":"Binary.Fixed"}}}
```
_Fixture `doc.json`:_
```
{"mti":"0100","binary_fields":{"52":"A1B2C3D4E5F60708"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec hex-top.json
```
#### Then
- exit code is `0`
- stdout does not contain `no constructor`
### Scenario: loads a Hex TLV subfield
#### Given
- Fixture file `hex-sub.json` is created.
#### Inputs
_Fixture `hex-sub.json`:_
```
{"name":"TLV Hex","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"55":{"type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL","tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"9F02":{"type":"Hex","length":6,"description":"Amount","enc":"Binary","prefix":"BerTLV"}}}}}
```
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0100-auth-request.hex --spec hex-sub.json
```
#### Then
- exit code is not `0`
- stderr does not contain `no constructor`
### Scenario: loads a Track1 field
#### Given
- Fixture file `track1.json` is created.
#### Inputs
_Fixture `track1.json`:_
```
{"name":"Track1","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"45":{"type":"Track1","length":76,"description":"Track 1","enc":"ASCII","prefix":"ASCII.LL"}}}
```
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0100-auth-request.hex --spec track1.json
```
#### Then
- exit code is not `0`
- stderr does not contain `no constructor`
### Scenario: loads a Track3 field
#### Given
- Fixture file `track3.json` is created.
#### Inputs
_Fixture `track3.json`:_
```
{"name":"Track3","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"36":{"type":"Track3","length":104,"description":"Track 3","enc":"ASCII","prefix":"ASCII.LLL"}}}
```
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0100-auth-request.hex --spec track3.json
```
#### Then
- exit code is not `0`
- stderr does not contain `no constructor`
### Scenario: loads an IndexTag composite subfield
#### Given
- Fixture file `indextag.json` is created.
#### Inputs
_Fixture `indextag.json`:_
```
{"name":"IndexTag","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"48":{"type":"Composite","length":999,"description":"IndexTag Composite","prefix":"ASCII.LLL","tag":{"sort":"StringsByInt","length":2,"enc":"ASCII"},"subfields":{"1":{"type":"IndexTag","length":2,"description":"Tag index","enc":"ASCII","prefix":"ASCII.Fixed"}}}}}
```
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0100-auth-request.hex --spec indextag.json
```
#### Then
- exit code is not `0`
- stderr does not contain `no constructor`
### Scenario: loads a composite tag that omits sort
#### Given
- Fixture file `nosort.json` is created.
#### Inputs
_Fixture `nosort.json`:_
```
{"name":"No sort","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"55":{"type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL","tag":{"enc":"BerTLVTag","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"9F02":{"type":"Binary","length":6,"description":"Amount","enc":"Binary","prefix":"BerTLV"}}}}}
```
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0100-auth-request.hex --spec nosort.json
```
#### Then
- exit code is not `0`
- stderr does not contain `unknown sort function`
## iso8583tool detection messaging
Source: `test/e2e/tools/iso8583tool/diagnostics.atago.yaml`
### Scenario: presents tied presets for an ambiguous message
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --no-color
```
#### Then
- exit code is `0`
- stdout contains `spec87ascii`, `fits equally well`
### Scenario: flags a truncated capture instead of only "custom layout"
#### When
```shell
iso8583tool doctor --raw 010000000000000008000103DF --no-color
```
#### Then
- exit code is not `0`
- stdout contains `truncated or malformed`
### Scenario: validate calls out a truncated capture rather than doctor
#### When
```shell
iso8583tool validate --raw 010000000000000008000103DF --no-color
```
#### Then
- exit code is not `0`
- stdout contains `truncated or malformed`
- stdout does not contain `iso8583tool doctor`
## iso8583tool diff
Source: `test/e2e/tools/iso8583tool/diff.atago.yaml`
### Scenario: reports changed fields in text form
#### When
```shell
cp "$ISO_EXAMPLES/basei/0100-auth-request.hex" before.hex
iso8583tool convert "$ISO_EXAMPLES/basei/0100-auth-request.hex" | sed 's/"000000005000"/"000000009999"/' | iso8583tool convert > after.hex

iso8583tool diff before.hex after.hex
```
#### Then
- after `iso8583tool diff before.hex after.hex`:
  - exit code is `0`
  - stdout contains `Field 4 changed`, `- 000000005000`, `+ 000000009999`
### Scenario: emits jq-compatible JSON
#### When
```shell
cp "$ISO_EXAMPLES/basei/0100-auth-request.hex" before.hex
iso8583tool convert "$ISO_EXAMPLES/basei/0100-auth-request.hex" | sed 's/"000000005000"/"000000009999"/' | iso8583tool convert > after.hex

iso8583tool diff before.hex after.hex --format json | jq -r ".changes[0].kind"
```
#### Then
- after `iso8583tool diff before.hex after.hex --format json | jq -r ".changes[0].kind"`:
  - exit code is `0`
  - stdout equals an exact value
### Scenario: reports no differences for identical messages
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0110-auth-response.hex $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `No differences.`
### Scenario: filters to a field subtree
#### When
```shell
cp "$ISO_EXAMPLES/basei/0100-auth-request.hex" before.hex
iso8583tool convert "$ISO_EXAMPLES/basei/0100-auth-request.hex" | sed 's/"000000005000"/"000000009999"/' | iso8583tool convert > after.hex

iso8583tool diff before.hex after.hex --filter 55
```
#### Then
- after `iso8583tool diff before.hex after.hex --filter 55`:
  - exit code is `0`
  - stdout contains `55.9F02`
  - stdout does not contain `Field 4 `
### Scenario: reads one side from stdin
#### When
```shell
cp "$ISO_EXAMPLES/basei/0100-auth-request.hex" before.hex
iso8583tool convert "$ISO_EXAMPLES/basei/0100-auth-request.hex" | sed 's/"000000005000"/"000000009999"/' | iso8583tool convert > after.hex

cat after.hex | iso8583tool diff before.hex -
```
#### Then
- after `cat after.hex | iso8583tool diff before.hex -`:
  - exit code is `0`
  - stdout contains `changed`
### Scenario: rejects two stdin sides
#### When
```shell
iso8583tool diff - -
```
#### Then
- exit code is not `0`
- stderr contains `stdin`
### Scenario: masks track data by default
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --color never
```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111D`
### Scenario: reveals raw values with --unsafe
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --color never --unsafe
```
#### Then
- exit code is `0`
- stdout contains `4111111111111111D`
### Scenario: rejects an unknown --format value
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --format bogus
```
#### Then
- exit code is not `0`
- stderr contains `unsupported format`
### Scenario: private-field safety - masks an embedded PAN by default
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4111111111111111"}}' | iso8583tool convert --to hex > pa.hex
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4222222222222222"}}' | iso8583tool convert --to hex > pb.hex

iso8583tool diff pa.hex pb.hex --color never
```
#### Then
- after `iso8583tool diff pa.hex pb.hex --color never`:
  - exit code is `0`
  - stdout does not contain `4111111111111111`
### Scenario: private-field safety - reveals the embedded PAN with --unsafe
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4111111111111111"}}' | iso8583tool convert --to hex > pa.hex
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4222222222222222"}}' | iso8583tool convert --to hex > pb.hex

iso8583tool diff pa.hex pb.hex --color never --unsafe
```
#### Then
- after `iso8583tool diff pa.hex pb.hex --color never --unsafe`:
  - exit code is `0`
  - stdout contains `4111111111111111`
## iso8583tool doctor
Source: `test/e2e/tools/iso8583tool/doctor.atago.yaml`
### Scenario: recommends the BASE I starter for an ASCII BASE I message
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `Recommended: --spec basei-starter`, `Confirm with: iso8583tool view`
### Scenario: detects a packed-BCD raw message
#### When
```shell
printf '\001\000\160\004\000\000\000\000\000\000\020\100\031\044\231\231\231\231\231\062\163\047\000\000\000\000\021\070\042\004' > message.bin

iso8583tool doctor message.bin --encoding raw
```
#### Then
- after `iso8583tool doctor message.bin --encoding raw`:
  - exit code is `0`
  - stdout contains `Recommended: --spec spec87bcd-starter`
### Scenario: auto-detects a raw .bin without --encoding
#### When
```shell
printf '\001\000\160\004\000\000\000\000\000\000\020\100\031\044\231\231\231\231\231\062\163\047\000\000\000\000\021\070\042\004' > message.bin

iso8583tool doctor message.bin
```
#### Then
- after `iso8583tool doctor message.bin`:
  - exit code is `0`
  - stdout contains `(raw input)`, `Recommended: --spec spec87bcd-starter`
### Scenario: emits a JSON report with --format json
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"recommended": "basei-starter"`, `"exact_round_trip": true`
### Scenario: exits non-zero when no preset fits
#### When
```shell
iso8583tool doctor --raw fffefd
```
#### Then
- exit code is not `0`
- stdout contains `No built-in preset could unpack`
### Scenario: is suggested by a wrong-spec validate failure
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --spec spec87bcd-starter
```
#### Then
- exit code is not `0`
- stdout contains `doctor`
### Scenario: marks every tied preset recommended and confirms with each
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --no-color
```
#### Then
- exit code is `0`
- stdout contains `view --spec basei-starter`, `view --spec spec87ascii`
### Scenario: explains how to choose between the tied basei-starter and spec87ascii presets
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --no-color
```
#### Then
- exit code is `0`
- stdout contains `Field 55`, `EMV`
### Scenario: shell-safe confirm hint quotes a path that contains a space
#### When
```shell
cp "$ISO_EXAMPLES/basei/0110-auth-response.hex" "with space.hex"
iso8583tool doctor "with space.hex" --no-color
```
#### Then
- after `iso8583tool doctor "with space.hex" --no-color`:
  - exit code is `0`
  - stdout contains `'`, `with space.hex`
### Scenario: custom-spec validate hint does not steer a custom-spec failure to doctor
#### Given
- Fixture file `spec.json` is created.
- Fixture file `other.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"F48 positional","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"48":{"type":"Composite","length":999,"description":"Private Data","prefix":"ASCII.LLL","tag":{"sort":"StringsByInt"},"subfields":{"1":{"type":"String","length":3,"description":"A","enc":"ASCII","prefix":"ASCII.Fixed"},"2":{"type":"String","length":2,"description":"B","enc":"ASCII","prefix":"ASCII.Fixed"}}}}}
```
_Fixture `other.json`:_
```
{"name":"F127 bitmap","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"127":{"type":"Composite","length":255,"description":"Private use field","prefix":"ASCII.LL","bitmap":{"type":"Bitmap","length":8,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed","disableAutoExpand":true},"subfields":{"1":{"type":"String","length":2,"description":"A","enc":"ASCII","prefix":"ASCII.Fixed"},"2":{"type":"String","length":2,"description":"B","enc":"ASCII","prefix":"ASCII.Fixed"}}}}}
```
#### When
```shell
printf "%s" "{\"mti\":\"0100\",\"fields\":{\"11\":\"123456\",\"127.1\":\"AA\",\"127.2\":\"BB\"}}" | iso8583tool convert --to hex --spec other.json > msg.hex
iso8583tool validate msg.hex --spec spec.json --encoding hex --no-color
```
#### Then
- after `iso8583tool validate msg.hex --spec spec.json --encoding hex --no-color`:
  - exit code is not `0`
  - stdout does not contain `doctor`
  - stdout contains `spec file`
## iso8583tool edge cases
Source: `test/e2e/tools/iso8583tool/edge_cases.atago.yaml`
### Scenario: rejects non-hex characters under --encoding hex
#### When
```shell
iso8583tool view --encoding hex --raw zzzz
```
#### Then
- exit code is not `0`
- stderr contains `hex`
### Scenario: rejects odd-length hex
#### When
```shell
iso8583tool view --raw 0100712
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: reports the failing field for a truncated message
#### When
```shell
iso8583tool validate --raw 01007220
```
#### Then
- exit code is not `0`
- stdout contains `[error]`
### Scenario: fails on empty inline input
#### When
```shell
iso8583tool view --raw ''
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: fails when the file does not exist
#### When
```shell
iso8583tool view /no/such/message.hex
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: fails when a directory is passed instead of a file
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: refuses both a file argument and --raw
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --raw 0100
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: fails to pack a document with no mti
#### When
```shell
printf "%s" "{\"fields\":{}}" | iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `mti`
### Scenario: fails to pack a document with an invalid TLV tag
#### When
```shell
printf "%s" "{\"mti\":\"0100\",\"binary_fields\":{\"55.ZZ\":\"00\"}}" | iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr is not empty
### Scenario: oversized input - rejects an oversized file with a clear limit error
#### When
```shell
head -c 1100000 /dev/zero | tr '\0' '0' > big.hex
iso8583tool view big.hex
```
#### Then
- after `iso8583tool view big.hex`:
  - exit code is not `0`
  - stderr contains `limit`
### Scenario: oversized input - rejects oversized stdin with a clear limit error
#### When
```shell
head -c 1100000 /dev/zero | tr '\0' '0' | iso8583tool view -
```
#### Then
- exit code is not `0`
- stderr contains `limit`
### Scenario: oversized input - cleanly fails a truncated field 55 TLV without panic
#### When
```shell
iso8583tool validate --raw 010000000000000008000103DF
```
#### Then
- exit code is not `0`
- stdout contains `[error]`
- stderr does not contain `panic`
### Scenario: oversized input - does not panic on a length-spoofed variable-length field
#### When
```shell
iso8583tool view --raw 0100400000000000000099
```
#### Then
- exit code is not `0`
- stderr does not contain `panic`
## iso8583tool ergonomics
Source: `test/e2e/tools/iso8583tool/ergonomics.atago.yaml`
### Scenario: flag ordering - accepts the target after the flags
#### When
```shell
iso8583tool view --format json $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `"mti"`
### Scenario: flag ordering - accepts the target before the flags
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"mti"`
### Scenario: flag ordering - accepts flags interleaved around the target
#### When
```shell
iso8583tool view --filter 39 $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 49
```
#### Then
- exit code is `0`
- stdout contains `Approved`, `JPY`
### Scenario: color - is plain by default when not on a terminal
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex | cat -v
```
#### Then
- exit code is `0`
- stdout does not contain `^[`
### Scenario: color - forces color with --color always
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --color always | cat -v
```
#### Then
- exit code is `0`
- stdout contains `^[`
### Scenario: color - stays plain with --no-color even when forced elsewhere
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --no-color | cat -v
```
#### Then
- exit code is `0`
- stdout does not contain `^[`
### Scenario: color - rejects an unknown --color value instead of ignoring it
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --color banana
```
#### Then
- exit code is not `0`
- stderr contains `invalid --color`
### Scenario: end of options - treats a dash-leading filename after -- as a positional
#### When
```shell
cp "$ISO_EXAMPLES/basei/0110-auth-response.hex" ./-response.hex
iso8583tool view -- -response.hex
```
#### Then
- after `iso8583tool view -- -response.hex`:
  - exit code is `0`
  - stdout contains `MTI`
### Scenario: config - applies an extension catalog from --config
#### Given
- Fixture file `cfg.json` is created.
#### Inputs
_Fixture `cfg.json`:_
```
{"spec":"basei-starter","extensions":[{"id":63,"name":"Acme Blob","strategy":"opaque"}]}
```
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0110-auth-response.hex --config cfg.json
```
#### Then
- exit code is `0`
- stdout contains `Acme Blob`
### Scenario: config - fails on a config with an invalid strategy
#### Given
- Fixture file `bad.json` is created.
#### Inputs
_Fixture `bad.json`:_
```
{"extensions":[{"id":1,"strategy":"nope"}]}
```
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --config bad.json
```
#### Then
- exit code is not `0`
- stderr contains `strategy`
## iso8583tool extension strategy
Source: `test/e2e/tools/iso8583tool/extension_strategy.atago.yaml`
### Scenario: a custom positional composite spec does not apply the BASE I catalog
#### Given
- Fixture file `spec.json` is created.
- Fixture file `msg.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "F48 positional",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "48": {
      "type":"Composite","length":999,"description":"Private Data","prefix":"ASCII.LLL",
      "tag":{"sort":"StringsByInt"},
      "subfields": {
        "1": {"type":"String","length":3,"description":"A","enc":"ASCII","prefix":"ASCII.Fixed"},
        "2": {"type":"String","length":2,"description":"B","enc":"ASCII","prefix":"ASCII.Fixed"}
      }
    }
  }
}
```
_Fixture `msg.json`:_
```
{"mti":"0100","fields":{"11":"123456","48.1":"ABC","48.2":"DE"}}
```
#### When
```shell
iso8583tool convert msg.json --to hex --spec spec.json --output msg.hex
iso8583tool view msg.hex --spec spec.json --encoding hex --no-color
```
#### Then
- after `iso8583tool view msg.hex --spec spec.json --encoding hex --no-color`:
  - exit code is `0`
  - stdout does not contain `Extension Field Strategy:`, `Additional Data - Private`, `[tlv]`
### Scenario: a built-in plain field documented as bitmap reports field 127 as opaque, matching the spec
#### Given
- Fixture file `msg.json` is created.
#### Inputs
_Fixture `msg.json`:_
```
{"mti":"0100","fields":{"11":"123456","127":"EEE"}}
```
#### When
```shell
iso8583tool convert msg.json --to hex --output msg.hex
iso8583tool view msg.hex --no-color
```
#### Then
- after `iso8583tool view msg.hex --no-color`:
  - exit code is `0`
  - stdout contains `F127 Reserved Private [opaque]`
  - stdout does not contain `F127 Reserved Private [bitmap]`
### Scenario: explains a dot-path set on a plain built-in field
#### When
```shell
printf "%s" "{\"mti\":\"0100\",\"fields\":{\"11\":\"123456\",\"48.1\":\"AB\"}}" | iso8583tool convert --to hex
```
#### Then
- exit code is not `0`
- stderr contains `dot-path subfields`
- stderr does not contain `PathMarshaler`
## iso8583tool convert field count
Source: `test/e2e/tools/iso8583tool/field_count.atago.yaml`
### Scenario: reports the top-level field count matching doctor
#### When
```shell
summary=$(iso8583tool convert "$ISO_EXAMPLES/basei/0100-auth-request.json" --output out.hex | head -1)
count=$(iso8583tool doctor out.hex --format json | sed -n 's/.*"field_count": \([0-9]*\).*/\1/p' | head -1)
[ "$summary" = "Converted with basei-starter (packed $count fields to hex)." ] && echo OK || echo MISMATCH

```
#### Then
- exit code is `0`
- stdout equals an exact value
## iso8583tool filter normalization
Source: `test/e2e/tools/iso8583tool/filter.atago.yaml`
### Scenario: matches a lowercase EMV tag in view
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0100-auth-request.hex --filter 55.9f02 --no-color
```
#### Then
- exit code is `0`
- stdout contains `55.9F02`
- stdout does not contain `<not present>`
### Scenario: matches a lowercase EMV tag in diff
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 55.8a
```
#### Then
- exit code is `0`
- stdout contains `55.8A`
- stdout does not contain `No field matched`
### Scenario: accepts "0" as an MTI alias in diff
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 0
```
#### Then
- exit code is `0`
- stdout contains `MTI changed`
### Scenario: reports an unmatched diff filter
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 999
```
#### Then
- exit code is `0`
- stdout contains `No field matched filter: 999`
### Scenario: reports an unmatched filter in JSON
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 999 --format json
```
#### Then
- exit code is `0`
- stdout contains `"missing_filters"`, `999`
## iso8583tool sensitive-data masking
Source: `test/e2e/tools/iso8583tool/masking.atago.yaml`
### Scenario: does not mask a non-PAN business identifier
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"ORDER_ID=1234567890123|TOKEN=ABC"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout contains `ORDER_ID=1234567890123`
### Scenario: masks a dash-separated PAN
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4111-1111-1111-1111"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `1111-1111-1111`
### Scenario: masks a space-separated PAN
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4111 1111 1111 1111"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `1111 1111 1111`
### Scenario: masks a PAN embedded in a non-private free-form field
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","44":"PAN=4111111111111111"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111`
### Scenario: masks the extended PAN field 34
#### When
```shell
printf '%s' '{"mti":"0100","fields":{"11":"123456","34":"411111111111111111111111"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `411111111111111111111111`
### Scenario: does not mask the country code field 20
#### When
```shell
printf '%s' '{"mti":"0100","fields":{"11":"123456","20":"840"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout contains `"20": "840"`
### Scenario: shows the raw field 20 change in diff
#### When
```shell
printf '%s' '{"mti":"0100","fields":{"11":"123456","20":"840"}}' | iso8583tool convert --to hex > a.hex
printf '%s' '{"mti":"0100","fields":{"11":"123456","20":"392"}}' | iso8583tool convert --to hex > b.hex

iso8583tool diff a.hex b.hex --no-color
```
#### Then
- after `iso8583tool diff a.hex b.hex --no-color`:
  - exit code is `0`
  - stdout contains `840`, `392`
### Scenario: masks a whole free-form track, not just its PAN
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"TRACK2=4111111111111111D29122011234567890"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `29122011234567890`
### Scenario: masks an underscore-labeled PAN (card_no)
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"card_no=4222222222222222"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `4222222222222222`
### Scenario: masks a spaced-label PAN (card number)
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"card number=4222222222222222"}}' | iso8583tool convert --to hex | iso8583tool view - --format json

```
#### Then
- exit code is `0`
- stdout does not contain `4222222222222222`
### Scenario: a custom positional composite with a PAN-numbered subfield does not mask subfield 48.2 with the top-level PAN rule
#### Given
- Fixture file `spec.json` is created.
- Fixture file `msg.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "F48 positional",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "48": {
      "type":"Composite","length":999,"description":"Private Data","prefix":"ASCII.LLL",
      "tag":{"sort":"StringsByInt"},
      "subfields": {
        "1": {"type":"String","length":3,"description":"A","enc":"ASCII","prefix":"ASCII.Fixed"},
        "2": {"type":"String","length":2,"description":"B","enc":"ASCII","prefix":"ASCII.Fixed"}
      }
    }
  }
}
```
_Fixture `msg.json`:_
```
{"mti":"0100","fields":{"11":"123456","48.1":"ABC","48.2":"DE"}}
```
#### When
```shell
iso8583tool convert msg.json --to hex --spec spec.json --output msg.hex
iso8583tool view msg.hex --spec spec.json --encoding hex --no-color
```
#### Then
- after `iso8583tool view msg.hex --spec spec.json --encoding hex --no-color`:
  - exit code is `0`
  - stdout contains `48.2`, `DE`
## iso8583tool masking under custom specs
Source: `test/e2e/tools/iso8583tool/masking_custom.atago.yaml`
### Scenario: masks a PAN in a binary field 63
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"B63","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"63":{"type":"Binary","length":999,"description":"Private","enc":"Binary","prefix":"ASCII.LLL"}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"63":"50414E3D34313131313131313131313131313131"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool view m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool view m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout does not contain `50414E3D`
### Scenario: masks a known 9F6B track2-equivalent tag in view
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"9F6B","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"55":{"type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL","tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"9F6B":{"type":"Binary","length":19,"description":"Track 2 Equivalent","enc":"Binary","prefix":"BerTLV"}}}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.9F6B":"4111111111111111D29122011234567890"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool view m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool view m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout does not contain `4111111111111111`
### Scenario: masks a known 9F6B track2-equivalent tag in redact
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"9F6B","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"55":{"type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL","tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"9F6B":{"type":"Binary","length":19,"description":"Track 2 Equivalent","enc":"Binary","prefix":"BerTLV"}}}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.9F6B":"4111111111111111D29122011234567890"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool redact m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool redact m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout does not contain `4111111111111111`
### Scenario: masks a track2-equivalent tag in a non-55 container (127.57)
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"T127","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"127":{"type":"Composite","length":999,"description":"Private TLV","prefix":"ASCII.LLL","tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"57":{"type":"Binary","length":18,"description":"Track2Eq","enc":"Binary","prefix":"BerTLV"}}}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"127.57":"4111111111111111D29122011234567890"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool view m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool view m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout does not contain `4111111111111111`
### Scenario: masks a sensitive tag nested in a constructed TLV (55.70.57)
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"C57","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"55":{"type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL","tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"70":{"type":"Composite","length":255,"description":"Template","prefix":"BerTLV","tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},"subfields":{"57":{"type":"Binary","length":18,"description":"Track2Eq","enc":"Binary","prefix":"BerTLV"}}}}}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.57":"4111111111111111D29122011234567890"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool redact m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool redact m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout does not contain `4111111111111111`
### Scenario: does not over-mask a harmless custom field 35
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"F35","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"35":{"type":"String","length":37,"description":"Partner Reference","enc":"ASCII","prefix":"ASCII.LL"}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456","35":"REF-ORDER-ABC-0001"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool view m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool view m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout contains `REF-ORDER-ABC-0001`
### Scenario: does not over-mask a harmless custom field 52
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"F52","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},"52":{"type":"String","length":8,"description":"Partner Status","enc":"ASCII","prefix":"ASCII.Fixed"}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"11":"123456","52":"ABCDEFGH"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool view m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool view m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout contains `ABCDEFGH`
### Scenario: still masks a real PAN in a custom field 2
#### Given
- Fixture file `spec.json` is created.
- Fixture file `doc.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{"name":"F2","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"2":{"type":"String","length":19,"description":"Account","enc":"ASCII","prefix":"ASCII.LL"},"11":{"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"}}}
```
_Fixture `doc.json`:_
```
{"mti":"0110","fields":{"2":"4111111111111111","11":"123456"}}
```
#### When
```shell
iso8583tool convert doc.json --to hex --spec spec.json --output m.hex
iso8583tool view m.hex --spec spec.json --format json
```
#### Then
- after `iso8583tool view m.hex --spec spec.json --format json`:
  - exit code is `0`
  - stdout does not contain `4111111111111111`
## iso8583tool view nested composites
Source: `test/e2e/tools/iso8583tool/nested_describe.atago.yaml`
### Scenario: nested positional composite keeps the nested positional path
#### Given
- Fixture file `spec.json` is created.
- Fixture file `msg.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "F48 nested positional",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "48": {
      "type":"Composite","length":999,"description":"Private Data","prefix":"ASCII.LLL",
      "tag":{"sort":"StringsByInt"},
      "subfields": {
        "1": {"type":"String","length":2,"description":"A","enc":"ASCII","prefix":"ASCII.Fixed"},
        "2": {
          "type":"Composite","length":6,"description":"B","prefix":"ASCII.Fixed",
          "tag":{"sort":"StringsByInt"},
          "subfields": {"1": {"type":"String","length":6,
… (truncated)
```
_Fixture `msg.json`:_
```
{"mti":"0100","fields":{"11":"123456","48.1":"AB","48.2.1":"260604"}}
```
#### When
```shell
iso8583tool convert msg.json --to hex --spec spec.json --output msg.hex
iso8583tool view msg.hex --spec spec.json --encoding hex --no-color
```
#### Then
- after `iso8583tool view msg.hex --spec spec.json --encoding hex --no-color`:
  - exit code is `0`
  - stdout contains `F48.2`, `48.2.1`
### Scenario: nested EMV tag annotation annotates a nested ARC tag as Approved
#### Given
- Fixture file `spec.json` is created.
- Fixture file `msg.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV 8A",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields
… (truncated, 2 more lines)
```
_Fixture `msg.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.8A":"3030","55.70.9A":"260605"}}
```
#### When
```shell
iso8583tool convert msg.json --to hex --spec spec.json --output msg.hex
iso8583tool view msg.hex --spec spec.json --encoding hex --no-color
```
#### Then
- after `iso8583tool view msg.hex --spec spec.json --encoding hex --no-color`:
  - exit code is `0`
  - stdout contains `55.70.8A`, `Approved`
### Scenario: nested EMV tag annotation decodes nested leaf tags in validate --format json
#### Given
- Fixture file `spec.json` is created.
- Fixture file `msg.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV 8A",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields
… (truncated, 2 more lines)
```
_Fixture `msg.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.8A":"3030","55.70.9A":"260605"}}
```
#### When
```shell
iso8583tool convert msg.json --to hex --spec spec.json --output msg.hex
iso8583tool validate msg.hex --spec spec.json --encoding hex --format json
```
#### Then
- after `iso8583tool validate msg.hex --spec spec.json --encoding hex --format json`:
  - exit code is `0`
  - stdout contains `55.70.9A`, `2026-06-05`
## iso8583tool nested TLV
Source: `test/e2e/tools/iso8583tool/nested_tlv.atago.yaml`
### Scenario: unpacks a nested TLV to its leaf path
#### Given
- Fixture file `spec.json` is created.
- Fixture file `a.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields": 
… (truncated)
```
_Fixture `a.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.9F02":"000000005000"}}
```
#### When
```shell
iso8583tool convert a.json --to hex --spec spec.json --output a.hex
iso8583tool convert a.hex --spec spec.json
```
#### Then
- after `iso8583tool convert a.hex --spec spec.json`:
  - exit code is `0`
  - stdout contains `"55.70.9F02"`
  - stdout does not contain `"55.70":`
### Scenario: selects a nested TLV leaf with --filter
#### Given
- Fixture file `spec.json` is created.
- Fixture file `a.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields": 
… (truncated)
```
_Fixture `a.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.9F02":"000000005000"}}
```
#### When
```shell
iso8583tool convert a.json --to hex --spec spec.json --output a.hex
iso8583tool view a.hex --spec spec.json --filter 55.70.9F02 --no-color
```
#### Then
- after `iso8583tool view a.hex --spec spec.json --filter 55.70.9F02 --no-color`:
  - exit code is `0`
  - stdout contains `55.70.9F02`
  - stdout does not contain `<not present>`
### Scenario: diffs at the nested TLV leaf tag
#### Given
- Fixture file `spec.json` is created.
- Fixture file `a.json` is created.
- Fixture file `b.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields": 
… (truncated)
```
_Fixture `a.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.9F02":"000000005000"}}
```
_Fixture `b.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.9F02":"000000009999"}}
```
#### When
```shell
iso8583tool convert a.json --to hex --spec spec.json --output a.hex
iso8583tool convert b.json --to hex --spec spec.json --output b.hex
iso8583tool diff a.hex b.hex --spec spec.json --no-color
```
#### Then
- after `iso8583tool diff a.hex b.hex --spec spec.json --no-color`:
  - exit code is `0`
  - stdout contains `Field 55.70.9F02 changed`
### Scenario: keeps a top-level tag and a nested tag set on the same field
#### Given
- Fixture file `spec.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields": 
… (truncated)
```
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.82":"3900","55.70.9F02":"000000005000"}}' > mix.json
iso8583tool convert mix.json --to hex --spec spec.json --output mix.hex
iso8583tool convert mix.hex --spec spec.json

```
#### Then
- exit code is `0`
- stdout contains `"55.82"`, `"55.70.9F02"`
### Scenario: shows the full nested path in the describe output
#### Given
- Fixture file `spec.json` is created.
- Fixture file `a.json` is created.
#### Inputs
_Fixture `spec.json`:_
```
{
  "name": "Constructed TLV",
  "fields": {
    "0": {"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},
    "1": {"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},
    "11": {"type":"String","length":6,"description":"STAN","enc":"ASCII","prefix":"ASCII.Fixed"},
    "55": {
      "type":"Composite","length":999,"description":"ICC","prefix":"ASCII.LLL",
      "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
      "subfields": {
        "70": {
          "type":"Composite","length":255,"description":"Template","prefix":"BerTLV",
          "tag":{"enc":"BerTLVTag","sort":"StringsByHex","skipUnknownTLVTags":true,"storeUnknownTLVTags":true},
          "subfields": 
… (truncated)
```
_Fixture `a.json`:_
```
{"mti":"0110","fields":{"11":"123456"},"binary_fields":{"55.70.9F02":"000000005000"}}
```
#### When
```shell
iso8583tool convert a.json --to hex --spec spec.json --output a.hex
iso8583tool view a.hex --spec spec.json --no-color
```
#### Then
- after `iso8583tool view a.hex --spec spec.json --no-color`:
  - exit code is `0`
  - stdout contains `F55.70`, `55.70.9F02`
## iso8583tool workflows
Source: `test/e2e/tools/iso8583tool/pipeline.atago.yaml`
### Scenario: streams sample -> convert -> view
#### When
```shell
iso8583tool sample 0100-auth-request --format hex | iso8583tool convert | iso8583tool convert | iso8583tool view -
```
#### Then
- exit code is `0`
- stdout contains `MTI`
### Scenario: editing an EMV tag edits one tag and packs it back
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.hex > msg.json
sed 's/"55.9F02": "000000005000"/"55.9F02": "000000010000"/' msg.json > edited.json

iso8583tool convert edited.json | iso8583tool view - --filter 55.9F02
```
#### Then
- after `iso8583tool convert edited.json | iso8583tool view - --filter 55.9F02`:
  - exit code is `0`
  - stdout contains `000000010000`
### Scenario: editing an EMV tag keeps an unknown tag through the round trip
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request-unknown-tlv.hex > u.json
iso8583tool convert u.json | iso8583tool validate -
```
#### Then
- after `iso8583tool convert u.json | iso8583tool validate -`:
  - exit code is `0`
  - stdout contains `55.DF8129`
### Scenario: extracts a single field value with --filter
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 39
```
#### Then
- exit code is `0`
- stdout contains `00`, `Approved`
## iso8583tool README examples
Source: `test/e2e/tools/iso8583tool/readme.atago.yaml`
### Scenario: quick start - lists the bundled samples
#### When
```shell
iso8583tool sample
```
#### Then
- exit code is `0`
- stdout contains `0100-auth-request`
### Scenario: quick start - views the BASE I auth response
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `Summary:`
### Scenario: quick start - validates the unknown-TLV sample
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0100-auth-request-unknown-tlv.hex
```
#### Then
- exit code is `0`
- stdout contains `55.DF8129`
### Scenario: quick start - converts the BASE I request to JSON
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.hex
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0100"`
### Scenario: view - shows JSON output
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"fields"`
### Scenario: view - filters the requested fields
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 39 --filter 55.8A
```
#### Then
- exit code is `0`
- stdout contains `Approved`
### Scenario: view - reads a message from stdin
#### When
```shell
cat $ISO_EXAMPLES/basei/0110-auth-response.hex | iso8583tool view -
```
#### Then
- exit code is `0`
- stdout contains `MTI`
### Scenario: view - is jq-compatible for fields
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --format json | jq -r '.fields["39"]'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: diff - compares a request and a response
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `changed`
### Scenario: diff - is jq-compatible for changes
#### When
```shell
iso8583tool diff $ISO_EXAMPLES/basei/0100-auth-request.hex $ISO_EXAMPLES/basei/0110-auth-response.hex --format json | jq -r '.changes[].path' | head -n1
```
#### Then
- exit code is `0`
- stdout is not empty
### Scenario: redact - masks the PAN for safe sharing
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex
```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111`
### Scenario: redact - supports a text format
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex --format text
```
#### Then
- exit code is `0`
- stdout contains `Redacted:`
### Scenario: convert - packs the BASE I request to hex
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.json
```
#### Then
- exit code is `0`
- stdout matches `/^3031/`
### Scenario: convert - converts a sample through stdin
#### When
```shell
iso8583tool sample 0100-auth-request --format hex | iso8583tool convert
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0100"`
### Scenario: convert - writes converted output to a file
#### When
```shell
tmp="$(mktemp)"; iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request.json --output "$tmp" && test -s "$tmp"
```
#### Then
- exit code is `0`
- stdout contains `Converted with`
### Scenario: validate - reports a broken inline message as an error
#### When
```shell
iso8583tool validate --raw 01007220
```
#### Then
- exit code is not `0`
- stdout contains `[error]`
### Scenario: validate - emits JSON when asked
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"valid": true`
### Scenario: doctor - recommends a preset for the BASE I sample
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `Recommended: --spec basei-starter`
### Scenario: doctor - is jq-compatible for the recommendation
#### When
```shell
iso8583tool doctor $ISO_EXAMPLES/basei/0110-auth-response.hex --format json | jq -r .recommended
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: specs - lists the presets
#### When
```shell
iso8583tool specs
```
#### Then
- exit code is `0`
- stdout contains `basei-starter (default)`
### Scenario: specs - is jq-compatible for preset names
#### When
```shell
iso8583tool specs --format json | jq -r '.[].name' | head -n1
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: sample - prints a sample as JSON
#### When
```shell
iso8583tool sample 0100-auth-request
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0100"`
### Scenario: sample - writes a sample as hex
#### When
```shell
tmp="$(mktemp)"; iso8583tool sample 0100-auth-request --format hex --output "$tmp" && test -s "$tmp"
```
#### Then
- exit code is `0`
- stdout contains `Wrote sample`
### Scenario: send (default 2byte-binary framing) - sends a packed 0800 and decodes the 0810 reply
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex
```
#### Then
- exit code is `0`
- stdout contains `0810`
### Scenario: send (default 2byte-binary framing) - reads the message from stdin and is jq-compatible for the response MTI
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool sample 0800-network-echo --format hex | iso8583tool send ${addr} - --format json | jq -r '.response_view.mti'
```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: send (default 2byte-binary framing) - asserts the reply with --expect-mti / --expect-field (no jq needed)
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --expect-mti 0810 --expect-field 39=00
```
#### Then
- exit code is `0`
- stdout contains `0810`
### Scenario: send (default 2byte-binary framing) - accepts an inline message via --raw
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} --raw '{"mti":"0800","fields":{"70":"301","11":"654321","41":"TERMNET1"}}' --format json
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0810"`
### Scenario: send (4-digit ASCII framing) - packs a JSON document and sends it with a 4-digit header
#### Given
- Background service `mock` is started: `iso-mock --framing 4digit-ascii --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0100-auth-request.json --framing 4digit-ascii --format json
```
#### Then
- exit code is `0`
- stdout contains `"framing": "4digit-ascii"`, `"mti": "0810"`
### Scenario: unknown TLV round-trip - preserves the unknown tag when unpacking and packing again
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/basei/0100-auth-request-unknown-tlv.hex | iso8583tool convert | iso8583tool view - --filter 55.DF8129
```
#### Then
- exit code is `0`
- stdout contains `DF8129`
### Scenario: other specs - validates the spec87ascii sample
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --spec spec87ascii
```
#### Then
- exit code is `0`
- stdout contains `Spec: spec87ascii`
### Scenario: other specs - strict-validates the spec87ascii sample under its intended preset
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --spec spec87ascii --strict
```
#### Then
- exit code is `0`
- stdout contains `ok`
### Scenario: other specs - views the spec87ascii sample
#### When
```shell
iso8583tool view $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --spec spec87ascii
```
#### Then
- exit code is `0`
- stdout contains `0800`
### Scenario: other specs - converts the spec87ascii sample to JSON
#### When
```shell
iso8583tool convert $ISO_EXAMPLES/spec87ascii/0800-network-echo.hex --spec spec87ascii
```
#### Then
- exit code is `0`
- stdout contains `"mti": "0800"`
## iso8583tool redact
Source: `test/e2e/tools/iso8583tool/redact.atago.yaml`
### Scenario: masks the PAN in JSON output
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex | jq -r '.fields["2"]'

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: never leaks the full PAN
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex
```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111`
### Scenario: fully masks the EMV application cryptogram
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex | jq -r '.binary_fields["55.9F26"]'

```
#### Then
- exit code is `0`
- stdout equals an exact value
### Scenario: supports a human-readable text format
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex --format text
```
#### Then
- exit code is `0`
- stdout contains `Redacted:`, `411111******1111`
### Scenario: orders text output by MTI then numeric field id
#### When
```shell
iso8583tool redact $ISO_EXAMPLES/basei/0100-auth-request.hex --format text --color never
```
#### Then
- exit code is `0`
- stdout contains `MTI:`
- stdout contains `F2 =`
### Scenario: reads from stdin for a Slack-safe pipe
#### When
```shell
cat $ISO_EXAMPLES/basei/0100-auth-request.hex | iso8583tool redact -
```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111`
### Scenario: masks a PAN embedded in a free-form private field (F63)
#### When
```shell
printf '%s' '{"mti":"0110","fields":{"11":"123456","39":"00","63":"PAN=4111111111111111"}}' | iso8583tool convert --to hex | iso8583tool redact -

```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111`
### Scenario: auto-detected input encoding redacts a raw binary capture without --encoding
#### When
```shell
printf '\001\000\160\004\000\000\000\000\000\000\020\100\031\044\231\231\231\231\231\062\163\047\000\000\000\000\021\070\042\004' > message.bin
iso8583tool redact message.bin --spec spec87bcd-starter
```
#### Then
- after `iso8583tool redact message.bin --spec spec87bcd-starter`:
  - exit code is `0`
  - stdout contains `"mti"`
### Scenario: auto-detected input encoding still masks the PAN in a raw binary capture
#### When
```shell
printf '\001\000\160\004\000\000\000\000\000\000\020\100\031\044\231\231\231\231\231\062\163\047\000\000\000\000\021\070\042\004' > message.bin
iso8583tool redact message.bin --spec spec87bcd-starter | jq -r '.fields["2"]'

```
#### Then
- after `iso8583tool redact message.bin --spec spec87bcd-starter | jq -r '.fields["2"]'
`:
  - exit code is `0`
  - stdout contains `*`
## iso8583tool control-byte sanitization
Source: `test/e2e/tools/iso8583tool/sanitize_output.atago.yaml`
### Scenario: view escapes control bytes
#### Given
- Fixture file `bin41.json` is created.
#### Inputs
_Fixture `bin41.json`:_
```
{"name":"Bin41","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"41":{"type":"Binary","length":8,"description":"Terminal","enc":"Binary","prefix":"Binary.Fixed"}}}
```
#### When
```shell
printf '%s' '{"mti":"0100","binary_fields":{"41":"1B5B324A20202020"}}' | iso8583tool convert --to hex --spec bin41.json > a.hex
iso8583tool view a.hex --no-color
```
#### Then
- after `iso8583tool view a.hex --no-color`:
  - exit code is `0`
  - stdout does not contain ``
  - stdout contains `^[`
### Scenario: validate escapes control bytes
#### Given
- Fixture file `bin41.json` is created.
#### Inputs
_Fixture `bin41.json`:_
```
{"name":"Bin41","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"41":{"type":"Binary","length":8,"description":"Terminal","enc":"Binary","prefix":"Binary.Fixed"}}}
```
#### When
```shell
printf '%s' '{"mti":"0100","binary_fields":{"41":"1B5B324A20202020"}}' | iso8583tool convert --to hex --spec bin41.json > a.hex
iso8583tool validate a.hex --no-color
```
#### Then
- after `iso8583tool validate a.hex --no-color`:
  - exit code is `0`
  - stdout does not contain ``
### Scenario: diff escapes control bytes
#### Given
- Fixture file `bin41.json` is created.
#### Inputs
_Fixture `bin41.json`:_
```
{"name":"Bin41","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"41":{"type":"Binary","length":8,"description":"Terminal","enc":"Binary","prefix":"Binary.Fixed"}}}
```
#### When
```shell
printf '%s' '{"mti":"0100","binary_fields":{"41":"1B5B324A20202020"}}' | iso8583tool convert --to hex --spec bin41.json > a.hex
printf '%s' '{"mti":"0100","binary_fields":{"41":"1B5B316D20202020"}}' | iso8583tool convert --to hex --spec bin41.json > b.hex
iso8583tool diff a.hex b.hex --no-color
```
#### Then
- after `iso8583tool diff a.hex b.hex --no-color`:
  - exit code is `0`
  - stdout does not contain ``
  - stdout contains `^[`
### Scenario: redact text escapes control bytes
#### Given
- Fixture file `bin41.json` is created.
#### Inputs
_Fixture `bin41.json`:_
```
{"name":"Bin41","fields":{"0":{"type":"String","length":4,"description":"MTI","enc":"ASCII","prefix":"ASCII.Fixed"},"1":{"type":"Bitmap","length":16,"description":"Bitmap","enc":"HexToASCII","prefix":"Hex.Fixed"},"41":{"type":"Binary","length":8,"description":"Terminal","enc":"Binary","prefix":"Binary.Fixed"}}}
```
#### When
```shell
printf '%s' '{"mti":"0100","binary_fields":{"41":"1B5B324A20202020"}}' | iso8583tool convert --to hex --spec bin41.json > a.hex
iso8583tool redact a.hex --format text --no-color
```
#### Then
- after `iso8583tool redact a.hex --format text --no-color`:
  - exit code is `0`
  - stdout does not contain ``
  - stdout contains `^[`
## iso8583tool send
Source: `test/e2e/tools/iso8583tool/send.atago.yaml`
### Scenario: sends an 0800 and decodes the 0810 response (2byte-binary)
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing 2byte-binary
```
#### Then
- exit code is `0`
- stdout contains `Framing:`, `2byte-binary`, `Request:`, `Response:`, `0810`
### Scenario: lists every response field, not only annotated codes
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing 2byte-binary --no-color
```
#### Then
- exit code is `0`
- stdout contains `41 = TERMNET1`, `48 = HEARTBEAT=BASEI`, `63 = ECHO=OK`
### Scenario: packs a JSON document and sends it (2byte-binary)
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.json --framing 2byte-binary --format json
```
#### Then
- exit code is `0`
- stdout at `$.framing` equals `2byte-binary`
- stdout at `$.response_view.mti` equals `0810`
- stdout contains `"rtt_ms"`, `"sent_bytes"`, `"received_bytes"`, `"request_view"`
### Scenario: reads the message from stdin via -
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} - --framing 2byte-binary < $ISO_EXAMPLES/basei/0800-network-echo.hex
```
#### Then
- exit code is `0`
- stdout contains `0810`, `Response:`
### Scenario: frames with a 4-digit ASCII length header
#### Given
- Background service `mock` is started: `iso-mock --framing 4digit-ascii --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing 4digit-ascii --format json
```
#### Then
- exit code is `0`
- stdout at `$.framing` equals `4digit-ascii`
- stdout at `$.response_view.mti` equals `0810`
### Scenario: sends with no length header and reads the reply until EOF
#### Given
- Background service `mock` is started: `iso-mock --framing none --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing none --format json
```
#### Then
- exit code is `0`
- stdout at `$.framing` equals `none`
- stdout at `$.response_view.mti` equals `0810`
### Scenario: decodes the response in describe output (none framing)
#### Given
- Background service `mock` is started: `iso-mock --framing none --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing none
```
#### Then
- exit code is `0`
- stdout contains `none`, `Response:`
### Scenario: exits non-zero with a clear error when the response times out
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt --no-reply`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing 2byte-binary --timeout 600ms
```
#### Then
- exit code is not `0`
- stderr contains `timed out`
### Scenario: exits non-zero when a none-framing peer never replies
#### Given
- Background service `mock` is started: `iso-mock --framing none --reply-hex $REPLY_HEX --ready-file ready.txt --no-reply`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --framing none --timeout 600ms
```
#### Then
- exit code is not `0`
- stderr contains `timed out`
### Scenario: passes when --expect-mti and --expect-field match the response
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --expect-mti 0810 --expect-field 39=00 --expect-field 70=301
```
#### Then
- exit code is `0`
- stdout contains `0810`
### Scenario: exits non-zero with a deterministic error on an MTI mismatch
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --expect-mti 0800
```
#### Then
- exit code is not `0`
- stdout contains `Response:`
- stderr contains `send expectation failed:`, `MTI: expected "0800", got "0810"`
### Scenario: exits non-zero when an expected field value differs
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --expect-field 39=99
```
#### Then
- exit code is not `0`
- stdout contains `Response:`
- stderr contains `send expectation failed:`, `F39: expected "99", got "00"`
### Scenario: rejects an --expect-field without PATH=VALUE
#### Given
- Background service `mock` is started: `iso-mock --framing 2byte-binary --reply-hex $REPLY_HEX --ready-file ready.txt`.
#### When
```shell
iso8583tool send ${addr} $ISO_EXAMPLES/basei/0800-network-echo.hex --expect-field 39
```
#### Then
- exit code is not `0`
- stderr contains `invalid --expect-field`
### Scenario: frames and prints the request without connecting
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --dry-run
```
#### Then
- exit code is `0`
- stdout contains `Dry run`, `Would send bytes:`, `Request:`, `0800`
### Scenario: emits a machine-readable dry-run record
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --dry-run --format json
```
#### Then
- exit code is `0`
- stdout at `$.dry_run` equals `true`
- stdout contains `"would_send_bytes"`
- stdout at `$.request_view.mti` equals `0800`
### Scenario: withholds the framed bytes by default
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --dry-run --no-color
```
#### Then
- exit code is `0`
- stdout does not contain `Framed bytes`
### Scenario: reveals the framed wire bytes under --unsafe
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --dry-run --unsafe --no-color
```
#### Then
- exit code is `0`
- stdout contains `Framed bytes:`
### Scenario: includes framed_hex in JSON only under --unsafe
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --dry-run --unsafe --format json
```
#### Then
- exit code is `0`
- stdout contains `"framed_hex"`
### Scenario: rejects expectations because there is no response to assert
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --dry-run --expect-mti 0810
```
#### Then
- exit code is not `0`
- stderr contains `dry-run`
### Scenario: rejects an invalid --framing value
#### When
```shell
iso8583tool send 127.0.0.1:1 $ISO_EXAMPLES/basei/0800-network-echo.hex --framing bogus
```
#### Then
- exit code is not `0`
- stderr contains `invalid --framing`
### Scenario: rejects a HOST:PORT without a port
#### When
```shell
iso8583tool send 127.0.0.1 $ISO_EXAMPLES/basei/0800-network-echo.hex --timeout 500ms
```
#### Then
- exit code is not `0`
- stderr contains `invalid address`
### Scenario: prints usage for send --help
#### When
```shell
iso8583tool send --help
```
#### Then
- exit code is `0`
- stdout contains `Usage: iso8583tool send`, `--framing`
## iso8583tool specs
Source: `test/e2e/tools/iso8583tool/specs.atago.yaml`
### Scenario: lists the built-in presets with the default marked
#### When
```shell
iso8583tool specs
```
#### Then
- exit code is `0`
- stdout contains `basei-starter (default)`, `spec87ascii`, `spec87bcd-starter`
### Scenario: emits a JSON array with --format json
#### When
```shell
iso8583tool specs --format json
```
#### Then
- exit code is `0`
- stdout contains `"name": "basei-starter"`, `"default": true`
### Scenario: rejects an unexpected positional argument
#### When
```shell
iso8583tool specs extra
```
#### Then
- exit code is not `0`
- stderr contains `Usage:`
## iso8583tool standard high-numbered fields
Source: `test/e2e/tools/iso8583tool/standard_fields.atago.yaml`
### Scenario: packs and round-trips fields 95/96/100/102/103/104
#### Given
- Fixture file `hi.json` is created.
#### Inputs
_Fixture `hi.json`:_
```
{"mti":"0200","fields":{"11":"123456","95":"000000000000000000000000000000000000000000","100":"12345678901","102":"1234567890123456789012345678","103":"8765432109876543210987654321","104":"DESCRIPTION"},"binary_fields":{"96":"A1B2C3D4E5F60708"}}
```
#### When
```shell
iso8583tool convert hi.json --to hex --output hi.hex
iso8583tool view hi.hex --format json
```
#### Then
- after `iso8583tool view hi.hex --format json`:
  - exit code is `0`
  - stdout contains `"100": "12345678901"`, `"104": "DESCRIPTION"`
### Scenario: packs the reserved fields 123-127
#### Given
- Fixture file `r.json` is created.
#### Inputs
_Fixture `r.json`:_
```
{"mti":"0100","fields":{"11":"123456","123":"AAA","124":"BBB","125":"CCC","126":"DDD","127":"EEE"}}
```
#### When
```shell
iso8583tool convert r.json --to hex
```
#### Then
- exit code is `0`
- stdout does not contain `not defined`
### Scenario: packs and round-trips the binary MAC field 128
#### Given
- Fixture file `m.json` is created.
#### Inputs
_Fixture `m.json`:_
```
{"mti":"0100","binary_fields":{"128":"A1B2C3D4E5F60708"}}
```
#### When
```shell
iso8583tool convert m.json --to hex --output m.hex
iso8583tool view m.hex --unsafe --format json
```
#### Then
- after `iso8583tool view m.hex --unsafe --format json`:
  - exit code is `0`
  - stdout contains `"128": "A1B2C3D4E5F60708"`
## iso8583tool validate --strict advice and network rules
Source: `test/e2e/tools/iso8583tool/strict_validate.atago.yaml`
### Scenario: fails a hollow authorization advice (0120)
#### Given
- Fixture file `h0120.json` is created.
#### Inputs
_Fixture `h0120.json`:_
```
{"mti":"0120","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0120.json --to hex --output h0120.hex
iso8583tool validate h0120.hex --strict
```
#### Then
- after `iso8583tool validate h0120.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: fails a hollow financial advice (0220)
#### Given
- Fixture file `h0220.json` is created.
#### Inputs
_Fixture `h0220.json`:_
```
{"mti":"0220","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0220.json --to hex --output h0220.hex
iso8583tool validate h0220.hex --strict
```
#### Then
- after `iso8583tool validate h0220.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: fails a hollow network advice (0820)
#### Given
- Fixture file `h0820.json` is created.
#### Inputs
_Fixture `h0820.json`:_
```
{"mti":"0820","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0820.json --to hex --output h0820.hex
iso8583tool validate h0820.hex --strict
```
#### Then
- after `iso8583tool validate h0820.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: fails a hollow network response (0810)
#### Given
- Fixture file `h0810.json` is created.
#### Inputs
_Fixture `h0810.json`:_
```
{"mti":"0810","fields":{"11":"123456","39":"00"}}
```
#### When
```shell
iso8583tool convert h0810.json --to hex --output h0810.hex
iso8583tool validate h0810.hex --strict
```
#### Then
- after `iso8583tool validate h0810.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: fails a hollow network advice response (0830)
#### Given
- Fixture file `h0830.json` is created.
#### Inputs
_Fixture `h0830.json`:_
```
{"mti":"0830","fields":{"11":"123456","39":"00"}}
```
#### When
```shell
iso8583tool convert h0830.json --to hex --output h0830.hex
iso8583tool validate h0830.hex --strict
```
#### Then
- after `iso8583tool validate h0830.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: still accepts the bundled network echo under --strict
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0800-network-echo.hex --strict
```
#### Then
- exit code is `0`
- stdout contains `ok`
### Scenario: fails a hollow authorization notification (0140)
#### Given
- Fixture file `h0140.json` is created.
#### Inputs
_Fixture `h0140.json`:_
```
{"mti":"0140","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0140.json --to hex --output h0140.hex
iso8583tool validate h0140.hex --strict
```
#### Then
- after `iso8583tool validate h0140.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: fails a hollow financial instruction ack (0270)
#### Given
- Fixture file `h0270.json` is created.
#### Inputs
_Fixture `h0270.json`:_
```
{"mti":"0270","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0270.json --to hex --output h0270.hex
iso8583tool validate h0270.hex --strict
```
#### Then
- after `iso8583tool validate h0270.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: fails a hollow file-action request (0300)
#### Given
- Fixture file `h0300.json` is created.
#### Inputs
_Fixture `h0300.json`:_
```
{"mti":"0300","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0300.json --to hex --output h0300.hex
iso8583tool validate h0300.hex --strict
```
#### Then
- after `iso8583tool validate h0300.hex --strict`:
  - exit code is not `0`
  - stdout contains `failed`
### Scenario: requires a PAN source for a reversal request (0400)
#### Given
- Fixture file `h0400.json` is created.
#### Inputs
_Fixture `h0400.json`:_
```
{"mti":"0400","fields":{"4":"000000001000","7":"0605123456","11":"123456","90":"020022334406041301050000000000000000000000"}}
```
#### When
```shell
iso8583tool convert h0400.json --to hex --output h0400.hex
iso8583tool validate h0400.hex --strict
```
#### Then
- after `iso8583tool validate h0400.hex --strict`:
  - exit code is not `0`
  - stdout contains `PAN source`
### Scenario: warns that reconciliation (0500) rules are not implemented
#### Given
- Fixture file `h0500.json` is created.
#### Inputs
_Fixture `h0500.json`:_
```
{"mti":"0500","fields":{"11":"123456"}}
```
#### When
```shell
iso8583tool convert h0500.json --to hex --output h0500.hex
iso8583tool validate h0500.hex --strict
```
#### Then
- after `iso8583tool validate h0500.hex --strict`:
  - exit code is `0`
  - stdout contains `class 5`
### Scenario: rejects an alphabetic value in a numeric field (70)
#### Given
- Fixture file `hc0800.json` is created.
#### Inputs
_Fixture `hc0800.json`:_
```
{"mti":"0800","fields":{"11":"123456","70":"ABC"}}
```
#### When
```shell
iso8583tool convert hc0800.json --to hex --output hc0800.hex
iso8583tool validate hc0800.hex --strict
```
#### Then
- after `iso8583tool validate hc0800.hex --strict`:
  - exit code is not `0`
  - stdout contains `must be numeric`
## iso8583tool validate
Source: `test/e2e/tools/iso8583tool/validate.atago.yaml`
### Scenario: passes a good message with exit 0
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `Validation: ok`, `MTI: 0110`
### Scenario: reports unknown TLV tags as a warning but still exits 0
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0100-auth-request-unknown-tlv.hex
```
#### Then
- exit code is `0`
- stdout contains `warning`, `55.DF8129`
### Scenario: fails a broken message with exit 1 and names the field
#### When
```shell
iso8583tool validate --raw 01007220
```
#### Then
- exit code is not `0`
- stdout contains `Validation: failed`, `[error]`, `input was`
### Scenario: emits a JSON report with --format json
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0110-auth-response.hex --format json
```
#### Then
- exit code is `0`
- stdout contains `"valid": true`, `"summary"`
### Scenario: accepts a complete sample under --strict
#### When
```shell
iso8583tool validate $ISO_EXAMPLES/basei/0110-auth-response.hex --strict
```
#### Then
- exit code is `0`
- stdout contains `Validation: ok`
### Scenario: flags a hollow response under --strict
#### When
```shell
printf "%s" "{\"mti\":\"0110\",\"fields\":{\"11\":\"123456\"}}" | iso8583tool convert --to hex | iso8583tool validate - --strict
```
#### Then
- exit code is not `0`
- stdout contains `Validation: failed`, `39`
### Scenario: omits the Decoded Fields heading when only the MTI decoded
#### When
```shell
printf "%s" "{\"mti\":\"0500\",\"fields\":{\"11\":\"123456\"}}" | iso8583tool convert --to hex | iso8583tool validate - --strict
```
#### Then
- exit code is `0`
- stdout does not contain `Decoded Fields:`
## iso8583tool view
Source: `test/e2e/tools/iso8583tool/view.atago.yaml`
### Scenario: describe output - decodes codes and prints a summary
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `Summary:`, `Approved`, `JPY 5000`, `06-04 12:34:56`
### Scenario: describe output - masks the PAN
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex
```
#### Then
- exit code is `0`
- stdout contains `411111******1111`
- stdout does not contain `4111111111111111`
### Scenario: json output - emits a decoded array and stays uncolored
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --format json --color always
```
#### Then
- exit code is `0`
- stdout contains `"decoded"`, `"meaning": "Approved"`
### Scenario: --filter prints only the requested fields
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 39 --filter 55.8A
```
#### Then
- exit code is `0`
- stdout contains `Approved`
- stdout does not contain `Primary Account Number`
### Scenario: --filter marks a field that is not present
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 90
```
#### Then
- exit code is `0`
- stdout contains `not present`
### Scenario: --filter emits object-shaped JSON with an explicit missing_filters list
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 39 --filter 90 --format json
```
#### Then
- exit code is `0`
- stdout contains `"mti"`, `"missing_filters"`, `"90"`, `"meaning": "Approved"`
### Scenario: --filter always emits missing_filters as an array even when nothing is missing
#### When
```shell
iso8583tool view $ISO_EXAMPLES/basei/0110-auth-response.hex --filter 39 --format json
```
#### Then
- exit code is `0`
- stdout contains `"missing_filters": []`
### Scenario: stdin - reads a message piped in via -
#### When
```shell
iso8583tool sample 0110-auth-response --format hex | iso8583tool view -
```
#### Then
- exit code is `0`
- stdout contains `MTI`, `Approved`
### Scenario: stdin - reads from stdin when the target is omitted
#### When
```shell
iso8583tool sample 0110-auth-response --format hex | iso8583tool view
```
#### Then
- exit code is `0`
- stdout contains `MTI`
### Scenario: raw binary + packed BCD - views a kanmu-like raw message with the packed-BCD starter preset
#### When
```shell
printf '\001\000\160\004\000\000\000\000\000\000\020\100\031\044\231\231\231\231\231\062\163\047\000\000\000\000\021\070\042\004' > message.bin
iso8583tool view message.bin --encoding raw --spec spec87bcd-starter
```
#### Then
- after `iso8583tool view message.bin --encoding raw --spec spec87bcd-starter`:
  - exit code is `0`
  - stdout contains `401924******9999`, `327327`, `1138`, `2204`
### Scenario: private-field safety - masks a PAN embedded in a free-form private field by default
#### When
```shell
printf "%s" "{\"mti\":\"0110\",\"fields\":{\"11\":\"123456\",\"39\":\"00\",\"63\":\"PAN=4111111111111111\"}}" | iso8583tool convert --to hex | iso8583tool view - --format json
```
#### Then
- exit code is `0`
- stdout does not contain `4111111111111111`
### Scenario: private-field safety - reveals the raw private-field value with --unsafe
#### When
```shell
printf "%s" "{\"mti\":\"0110\",\"fields\":{\"11\":\"123456\",\"39\":\"00\",\"63\":\"PAN=4111111111111111\"}}" | iso8583tool convert --to hex | iso8583tool view - --format json --unsafe
```
#### Then
- exit code is `0`
- stdout contains `4111111111111111`