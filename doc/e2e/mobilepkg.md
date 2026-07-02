# atago Behavior Specs
## Summary
8 suites · 35 scenarios
## Contents
- [mobilepkg inspect --baseline](#mobilepkg-inspect---baseline) — 2 scenarios
  - [diffing a package against its own baseline reports no changes](#scenario-diffing-a-package-against-its-own-baseline-reports-no-changes)
  - [a missing baseline file is an error](#scenario-a-missing-baseline-file-is-an-error)
- [mobilepkg CLI](#mobilepkg-cli) — 5 scenarios
  - [version prints the tool version](#scenario-version-prints-the-tool-version)
  - [--version is an alias for version](#scenario---version-is-an-alias-for-version)
  - [help lists the commands](#scenario-help-lists-the-commands)
  - [no arguments prints usage and fails](#scenario-no-arguments-prints-usage-and-fails)
  - [unknown command fails with a helpful message](#scenario-unknown-command-fails-with-a-helpful-message)
- [mobilepkg compare](#mobilepkg-compare) — 3 scenarios
  - [comparing a package with itself reports no identity or version change](#scenario-comparing-a-package-with-itself-reports-no-identity-or-version-change)
  - [diff is an alias for compare](#scenario-diff-is-an-alias-for-compare)
  - [compare requires two files](#scenario-compare-requires-two-files)
- [mobilepkg error handling](#mobilepkg-error-handling) — 4 scenarios
  - [inspect with no file prints usage and fails](#scenario-inspect-with-no-file-prints-usage-and-fails)
  - [inspect on a missing file fails](#scenario-inspect-on-a-missing-file-fails)
  - [inspect rejects an unknown output format](#scenario-inspect-rejects-an-unknown-output-format)
  - [inspect rejects a file that is not a mobile package](#scenario-inspect-rejects-a-file-that-is-not-a-mobile-package)
- [mobilepkg inspect --fail-on](#mobilepkg-inspect---fail-on) — 4 scenarios
  - [--fail-on error exits 1 on an app with error findings](#scenario---fail-on-error-exits-1-on-an-app-with-error-findings)
  - [--fail-on warn also exits 1 (warnings present)](#scenario---fail-on-warn-also-exits-1-warnings-present)
  - [--fail-on embeds a failing verdict with triggering findings](#scenario---fail-on-embeds-a-failing-verdict-with-triggering-findings)
  - [an unknown --fail-on threshold still trips the gate (lenient today)](#scenario-an-unknown---fail-on-threshold-still-trips-the-gate-lenient-today)
- [mobilepkg inspect (AndroGoat APK)](#mobilepkg-inspect-androgoat-apk) — 9 scenarios
  - [reports platform and format for an APK](#scenario-reports-platform-and-format-for-an-apk)
  - [extracts identity and version from the manifest](#scenario-extracts-identity-and-version-from-the-manifest)
  - [reports the debuggable and allow_backup manifest flags](#scenario-reports-the-debuggable-and-allow_backup-manifest-flags)
  - [flags a debug-signed APK](#scenario-flags-a-debug-signed-apk)
  - [raises an error finding for the debuggable manifest](#scenario-raises-an-error-finding-for-the-debuggable-manifest)
  - [flags an unguarded exported content provider](#scenario-flags-an-unguarded-exported-content-provider)
  - [warns about backup being allowed](#scenario-warns-about-backup-being-allowed)
  - [emits schema_version and tool_version envelope fields](#scenario-emits-schema_version-and-tool_version-envelope-fields)
  - [does not emit a verdict without --fail-on](#scenario-does-not-emit-a-verdict-without---fail-on)
- [mobilepkg inspect --format markdown](#mobilepkg-inspect---format-markdown) — 3 scenarios
  - [renders the report heading and package table](#scenario-renders-the-report-heading-and-package-table)
  - [includes the Top Findings and Exported Components sections](#scenario-includes-the-top-findings-and-exported-components-sections)
  - [md is accepted as an alias for markdown](#scenario-md-is-accepted-as-an-alias-for-markdown)
- [mobilepkg README examples](#mobilepkg-readme-examples) — 5 scenarios
  - [inspect app.apk (default JSON)](#scenario-inspect-appapk-default-json)
  - [inspect --format markdown app.apk](#scenario-inspect---format-markdown-appapk)
  - [inspect --fail-on warn app.apk (CI gate trips on AndroGoat)](#scenario-inspect---fail-on-warn-appapk-ci-gate-trips-on-androgoat)
  - [inspect --baseline prev.json app.apk](#scenario-inspect---baseline-prevjson-appapk)
  - [compare old.apk new.apk](#scenario-compare-oldapk-newapk)
## mobilepkg inspect --baseline
Source: `test/e2e/tools/mobilepkg/baseline.atago.yaml`
### Scenario: diffing a package against its own baseline reports no changes
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk > base.json
mobilepkg inspect --baseline base.json app.apk
```
#### Then
- after `mobilepkg inspect --baseline base.json app.apk`:
  - exit code is `0`
  - stdout at `$.result.diff.identity_changed` equals `false`
  - stdout at `$.result.diff.version_changed` equals `false`
  - stdout at `$.result.diff.entry_changed` equals `false`
### Scenario: a missing baseline file is an error
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --baseline no_such_baseline.json app.apk
```
#### Then
- after `mobilepkg inspect --baseline no_such_baseline.json app.apk`:
  - exit code is not `0`
  - stderr contains `baseline`
## mobilepkg CLI
Source: `test/e2e/tools/mobilepkg/cli.atago.yaml`
### Scenario: version prints the tool version
#### When
```shell
mobilepkg version
```
#### Then
- exit code is `0`
- stdout contains `mobilepkg`
### Scenario: --version is an alias for version
#### When
```shell
mobilepkg --version
```
#### Then
- exit code is `0`
- stdout contains `mobilepkg`
### Scenario: help lists the commands
#### When
```shell
mobilepkg --help
```
#### Then
- exit code is `0`
- stdout contains `inspect`, `compare`
### Scenario: no arguments prints usage and fails
#### When
```shell
mobilepkg
```
#### Then
- exit code is not `0`
- stderr contains `Usage`
### Scenario: unknown command fails with a helpful message
#### When
```shell
mobilepkg bogus
```
#### Then
- exit code is not `0`
- stderr contains `unknown command`
## mobilepkg compare
Source: `test/e2e/tools/mobilepkg/compare.atago.yaml`
### Scenario: comparing a package with itself reports no identity or version change
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg compare app.apk app.apk
```
#### Then
- after `mobilepkg compare app.apk app.apk`:
  - exit code is `0`
  - stdout at `$.identity_changed` equals `false`
  - stdout at `$.version_changed` equals `false`
  - stdout at `$.old_platform` equals `android`
  - stdout at `$.new_format` equals `apk`
### Scenario: diff is an alias for compare
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg diff app.apk app.apk
```
#### Then
- after `mobilepkg diff app.apk app.apk`:
  - exit code is `0`
  - stdout at `$.identity_changed` equals `false`
### Scenario: compare requires two files
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg compare app.apk
```
#### Then
- after `mobilepkg compare app.apk`:
  - exit code is not `0`
  - stderr contains `Usage: mobilepkg compare`
## mobilepkg error handling
Source: `test/e2e/tools/mobilepkg/errors.atago.yaml`
### Scenario: inspect with no file prints usage and fails
#### When
```shell
mobilepkg inspect
```
#### Then
- exit code is not `0`
- stderr contains `Usage: mobilepkg inspect`
### Scenario: inspect on a missing file fails
#### When
```shell
mobilepkg inspect does_not_exist.apk
```
#### Then
- exit code is not `0`
- stderr contains `no such file`
### Scenario: inspect rejects an unknown output format
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --format xml app.apk
```
#### Then
- after `mobilepkg inspect --format xml app.apk`:
  - exit code is not `0`
  - stderr contains `unknown format`
### Scenario: inspect rejects a file that is not a mobile package
#### Given
- Fixture file `notapkg.apk` is created.
#### Inputs
_Fixture `notapkg.apk`:_
```
this is not a zip archive
```
#### When
```shell
mobilepkg inspect notapkg.apk
```
#### Then
- exit code is not `0`
- stderr contains `unsupported package format`
## mobilepkg inspect --fail-on
Source: `test/e2e/tools/mobilepkg/fail_on.atago.yaml`
### Scenario: --fail-on error exits 1 on an app with error findings
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --fail-on error app.apk
```
#### Then
- after `mobilepkg inspect --fail-on error app.apk`:
  - exit code is `1`
### Scenario: --fail-on warn also exits 1 (warnings present)
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --fail-on warn app.apk
```
#### Then
- after `mobilepkg inspect --fail-on warn app.apk`:
  - exit code is `1`
### Scenario: --fail-on embeds a failing verdict with triggering findings
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --fail-on error app.apk || true
```
#### Then
- after `mobilepkg inspect --fail-on error app.apk || true`:
  - stdout at `$.verdict.passed` equals `false`
  - stdout at `$.verdict.triggering_findings[?(@.id=='manifest.debuggable')].severity` equals `error`
### Scenario: an unknown --fail-on threshold still trips the gate (lenient today)
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --fail-on bogus app.apk
```
#### Then
- after `mobilepkg inspect --fail-on bogus app.apk`:
  - exit code is not `0`
## mobilepkg inspect (AndroGoat APK)
Source: `test/e2e/tools/mobilepkg/inspect.atago.yaml`
### Scenario: reports platform and format for an APK
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - exit code is `0`
  - stdout at `$.result.platform` equals `android`
  - stdout at `$.result.format` equals `apk`
### Scenario: extracts identity and version from the manifest
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.result.identity.identifier` equals `owasp.sat.agoat`
  - stdout at `$.result.version.marketing` equals `1.0`
  - stdout at `$.result.entry.name` equals `owasp.sat.agoat.SplashActivity`
### Scenario: reports the debuggable and allow_backup manifest flags
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.result.debuggable` equals `true`
  - stdout at `$.result.allow_backup` equals `true`
### Scenario: flags a debug-signed APK
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.result.signing.certificates[0].subject` equals `Android Debug`
  - stdout at `$.result.findings[?(@.id=='signing.debug_cert')].severity` equals `error`
### Scenario: raises an error finding for the debuggable manifest
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.result.findings[?(@.id=='manifest.debuggable')].severity` equals `error`
### Scenario: flags an unguarded exported content provider
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.result.findings[?(@.id=='exported.provider.owasp.sat.agoat.ContentProviderActivity')].severity` equals `error`
### Scenario: warns about backup being allowed
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.result.findings[?(@.id=='manifest.allow_backup')].severity` equals `warn`
### Scenario: emits schema_version and tool_version envelope fields
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout at `$.schema_version` matches `/^[0-9]+\.[0-9]+\.[0-9]+$/`
  - stdout at `$.tool_version` matches `/^v?[0-9]/`
### Scenario: does not emit a verdict without --fail-on
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - stdout does not contain `verdict`
## mobilepkg inspect --format markdown
Source: `test/e2e/tools/mobilepkg/inspect_markdown.atago.yaml`
### Scenario: renders the report heading and package table
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --format markdown app.apk
```
#### Then
- after `mobilepkg inspect --format markdown app.apk`:
  - exit code is `0`
  - stdout contains `# mobilepkg Inspection Report`, `| Platform | android |`, `| Identifier | owasp.sat.agoat |`
### Scenario: includes the Top Findings and Exported Components sections
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --format markdown app.apk
```
#### Then
- after `mobilepkg inspect --format markdown app.apk`:
  - stdout contains `## Top Findings`, `## Exported Components`
### Scenario: md is accepted as an alias for markdown
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --format md app.apk
```
#### Then
- after `mobilepkg inspect --format md app.apk`:
  - exit code is `0`
  - stdout contains `# mobilepkg Inspection Report`
## mobilepkg README examples
Source: `test/e2e/tools/mobilepkg/readme.atago.yaml`
### Scenario: inspect app.apk (default JSON)
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk
```
#### Then
- after `mobilepkg inspect app.apk`:
  - exit code is `0`
  - stdout at `$.result.platform` equals `android`
### Scenario: inspect --format markdown app.apk
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --format markdown app.apk
```
#### Then
- after `mobilepkg inspect --format markdown app.apk`:
  - exit code is `0`
  - stdout contains `# mobilepkg Inspection Report`
### Scenario: inspect --fail-on warn app.apk (CI gate trips on AndroGoat)
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect --fail-on warn app.apk
```
#### Then
- after `mobilepkg inspect --fail-on warn app.apk`:
  - exit code is `1`
### Scenario: inspect --baseline prev.json app.apk
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" app.apk
mobilepkg inspect app.apk > prev.json
mobilepkg inspect --baseline prev.json app.apk
```
#### Then
- after `mobilepkg inspect --baseline prev.json app.apk`:
  - exit code is `0`
  - stdout at `$.result.diff.identity_changed` equals `false`
### Scenario: compare old.apk new.apk
#### When
```shell
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" old.apk
cp "$MOBILEPKG_TESTDATA/android/androgoat_rich.apk" new.apk
mobilepkg compare old.apk new.apk
```
#### Then
- after `mobilepkg compare old.apk new.apk`:
  - exit code is `0`
  - stdout at `$.identity_changed` equals `false`