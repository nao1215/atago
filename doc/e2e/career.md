# atago Behavior Specs
## Summary
4 suites · 27 scenarios
## Contents
- [career CLI](#career-cli) — 6 scenarios
  - [prints root help with no arguments](#scenario-prints-root-help-with-no-arguments)
  - [prints the version](#scenario-prints-the-version)
  - [lists the available templates](#scenario-lists-the-available-templates)
  - [documents the aliases generate accepts](#scenario-documents-the-aliases-generate-accepts)
  - [describes the generate command](#scenario-describes-the-generate-command)
  - [unknown command fails with a helpful message](#scenario-unknown-command-fails-with-a-helpful-message)
- [career generate](#career-generate) — 13 scenarios
  - [cv renders a PDF (default template) with an accent color](#scenario-cv-renders-a-pdf-default-template-with-an-accent-color)
  - [cv writes a valid PDF header](#scenario-cv-writes-a-valid-pdf-header)
  - [japanese-resume renders a PDF from a positional input](#scenario-japanese-resume-renders-a-pdf-from-a-positional-input)
  - [japanese-resume accepts the 履歴書 alias](#scenario-japanese-resume-accepts-the-履歴書-alias)
  - [japanese-resume embeds the bundled sample portrait passed with --photo](#scenario-japanese-resume-embeds-the-bundled-sample-portrait-passed-with---photo)
  - [work-history renders a PDF using --input and the default output name](#scenario-work-history-renders-a-pdf-using---input-and-the-default-output-name)
  - [work-history still accepts the legacy career-history alias](#scenario-work-history-still-accepts-the-legacy-career-history-alias)
  - [work-history accepts the 職務経歴書 alias](#scenario-work-history-accepts-the-職務経歴書-alias)
  - [multiple templates renders every template with -t all](#scenario-multiple-templates-renders-every-template-with--t-all)
  - [errors fails when the input file is missing](#scenario-errors-fails-when-the-input-file-is-missing)
  - [errors fails on an unknown template](#scenario-errors-fails-on-an-unknown-template)
  - [errors fails on an invalid accent color](#scenario-errors-fails-on-an-invalid-accent-color)
  - [errors fails when no input is given](#scenario-errors-fails-when-no-input-is-given)
- [career init](#career-init) — 4 scenarios
  - [writes a starter file](#scenario-writes-a-starter-file)
  - [refuses to overwrite without --force](#scenario-refuses-to-overwrite-without---force)
  - [overwrites with --force](#scenario-overwrites-with---force)
  - [produces a file that generate accepts](#scenario-produces-a-file-that-generate-accepts)
- [career README examples](#career-readme-examples) — 4 scenarios
  - [cv example: career generate resume.yaml -t cv -o cv.pdf](#scenario-cv-example-career-generate-resumeyaml--t-cv--o-cvpdf)
  - [japanese-resume example: -t japanese-resume -o rirekisho.pdf](#scenario-japanese-resume-example--t-japanese-resume--o-rirekishopdf)
  - [work-history example: -t work-history -o shokumukeirekisho.pdf](#scenario-work-history-example--t-work-history--o-shokumukeirekishopdf)
  - [all example: -t all writes the three default file names](#scenario-all-example--t-all-writes-the-three-default-file-names)
## career CLI
Source: `test/e2e/tools/career/cli.atago.yaml`
### Scenario: prints root help with no arguments
#### When
```shell
career
```
#### Then
- exit code is `0`
- stdout contains `Usage:`, `generate`
### Scenario: prints the version
#### When
```shell
career version
```
#### Then
- exit code is `0`
- stdout contains `career`
### Scenario: lists the available templates
#### When
```shell
career templates
```
#### Then
- exit code is `0`
- stdout contains `cv`, `japanese-resume`, `work-history`
### Scenario: documents the aliases generate accepts
#### When
```shell
career templates
```
#### Then
- exit code is `0`
- stdout contains `履歴書`, `職務経歴書`, `career-history`
### Scenario: describes the generate command
#### When
```shell
career help generate
```
#### Then
- exit code is `0`
- stdout contains `--template`
### Scenario: unknown command fails with a helpful message
#### When
```shell
career frobnicate
```
#### Then
- exit code is not `0`
- stderr contains `unknown command`
## career generate
Source: `test/e2e/tools/career/generate.atago.yaml`
### Scenario: cv renders a PDF (default template) with an accent color
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml --accent "#2c6e6e" -o cv.pdf
```
#### Then
- after `career generate resume.yaml --accent "#2c6e6e" -o cv.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `cv.pdf` exists
#### Generated artifacts
- `cv.pdf`
### Scenario: cv writes a valid PDF header
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t cv -o cv.pdf
head -c 4 cv.pdf
```
#### Then
- after `head -c 4 cv.pdf`:
  - stdout equals an exact value
  - pdf `cv.pdf` >= 1 pages
### Scenario: japanese-resume renders a PDF from a positional input
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t japanese-resume -o out.pdf
```
#### Then
- after `career generate resume.yaml -t japanese-resume -o out.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `out.pdf` exists
#### Generated artifacts
- `out.pdf`
### Scenario: japanese-resume accepts the 履歴書 alias
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t 履歴書 -o out.pdf
```
#### Then
- after `career generate resume.yaml -t 履歴書 -o out.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `out.pdf` exists
#### Generated artifacts
- `out.pdf`
### Scenario: japanese-resume embeds the bundled sample portrait passed with --photo
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t japanese-resume --photo "$CAREER_EXAMPLES/../image/sample_japanese_man.jpg" -o out.pdf
```
#### Then
- after `career generate resume.yaml -t japanese-resume --photo "$CAREER_EXAMPLES/../image/sample_japanese_man.jpg" -o out.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `out.pdf` exists
#### Generated artifacts
- `out.pdf`
### Scenario: work-history renders a PDF using --input and the default output name
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate --input resume.yaml --template work-history --output ch.pdf
```
#### Then
- after `career generate --input resume.yaml --template work-history --output ch.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `ch.pdf` exists
#### Generated artifacts
- `ch.pdf`
### Scenario: work-history still accepts the legacy career-history alias
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t career-history -o legacy.pdf
```
#### Then
- after `career generate resume.yaml -t career-history -o legacy.pdf`:
  - exit code is `0`
  - stdout contains `work-history`
  - file `legacy.pdf` exists
#### Generated artifacts
- `legacy.pdf`
### Scenario: work-history accepts the 職務経歴書 alias
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t 職務経歴書 -o ja.pdf
```
#### Then
- after `career generate resume.yaml -t 職務経歴書 -o ja.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `ja.pdf` exists
#### Generated artifacts
- `ja.pdf`
### Scenario: multiple templates renders every template with -t all
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t all
```
#### Then
- after `career generate resume.yaml -t all`:
  - exit code is `0`
  - stdout contains `cv`, `japanese-resume`, `work-history`
  - file `cv.pdf` exists
  - file `japanese-resume.pdf` exists
  - file `work-history.pdf` exists
#### Generated artifacts
- `cv.pdf`
- `japanese-resume.pdf`
- `work-history.pdf`
### Scenario: errors fails when the input file is missing
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate nope.yaml -t cv -o out.pdf
```
#### Then
- after `career generate nope.yaml -t cv -o out.pdf`:
  - exit code is not `0`
  - stderr is not empty
### Scenario: errors fails on an unknown template
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t bogus -o out.pdf
```
#### Then
- after `career generate resume.yaml -t bogus -o out.pdf`:
  - exit code is not `0`
  - stderr contains `unknown template`
### Scenario: errors fails on an invalid accent color
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate resume.yaml -t cv --accent bogus -o out.pdf
```
#### Then
- after `career generate resume.yaml -t cv --accent bogus -o out.pdf`:
  - exit code is not `0`
  - stderr contains `hex color`
### Scenario: errors fails when no input is given
#### When
```shell
cp "$CAREER_EXAMPLES/minimal.yaml" resume.yaml
career generate -t cv
```
#### Then
- after `career generate -t cv`:
  - exit code is not `0`
  - stderr contains `no input file`
## career init
Source: `test/e2e/tools/career/init.atago.yaml`
### Scenario: writes a starter file
#### When
```shell
career init resume.yaml
```
#### Then
- exit code is `0`
- stdout contains `wrote`
- file `resume.yaml` exists
#### Generated artifacts
- `resume.yaml`
### Scenario: refuses to overwrite without --force
#### When
```shell
career init resume.yaml
career init resume.yaml
```
#### Then
- after `career init resume.yaml`:
  - exit code is not `0`
  - stderr contains `already exists`
### Scenario: overwrites with --force
#### When
```shell
career init resume.yaml
career init resume.yaml --force
```
#### Then
- after `career init resume.yaml --force`:
  - exit code is `0`
  - stdout contains `wrote`
### Scenario: produces a file that generate accepts
#### When
```shell
career init resume.yaml
career generate resume.yaml -t cv -o cv.pdf
```
#### Then
- after `career generate resume.yaml -t cv -o cv.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `cv.pdf` exists
#### Generated artifacts
- `cv.pdf`
## career README examples
Source: `test/e2e/tools/career/readme.atago.yaml`
### Scenario: cv example: career generate resume.yaml -t cv -o cv.pdf
#### When
```shell
cp "$CAREER_EXAMPLES/resume.yaml" resume.yaml
career generate resume.yaml -t cv -o cv.pdf
```
#### Then
- after `career generate resume.yaml -t cv -o cv.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `cv.pdf` exists
#### Generated artifacts
- `cv.pdf`
### Scenario: japanese-resume example: -t japanese-resume -o rirekisho.pdf
#### When
```shell
cp "$CAREER_EXAMPLES/resume.yaml" resume.yaml
career generate resume.yaml -t japanese-resume -o rirekisho.pdf
```
#### Then
- after `career generate resume.yaml -t japanese-resume -o rirekisho.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `rirekisho.pdf` exists
#### Generated artifacts
- `rirekisho.pdf`
### Scenario: work-history example: -t work-history -o shokumukeirekisho.pdf
#### When
```shell
cp "$CAREER_EXAMPLES/resume.yaml" resume.yaml
career generate resume.yaml -t work-history -o shokumukeirekisho.pdf
```
#### Then
- after `career generate resume.yaml -t work-history -o shokumukeirekisho.pdf`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `shokumukeirekisho.pdf` exists
#### Generated artifacts
- `shokumukeirekisho.pdf`
### Scenario: all example: -t all writes the three default file names
#### When
```shell
cp "$CAREER_EXAMPLES/resume.yaml" resume.yaml
career generate resume.yaml -t all
```
#### Then
- after `career generate resume.yaml -t all`:
  - exit code is `0`
  - stdout contains `wrote`
  - file `cv.pdf` exists
  - file `japanese-resume.pdf` exists
  - file `work-history.pdf` exists
#### Generated artifacts
- `cv.pdf`
- `japanese-resume.pdf`
- `work-history.pdf`
