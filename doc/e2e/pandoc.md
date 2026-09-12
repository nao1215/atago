# atago Behavior Specs
## Summary
2 suites · 8 scenarios
## Contents
- [pandoc + changes (a conversion writes exactly its output)](#pandoc--changes-a-conversion-writes-exactly-its-output) — 1 scenario
  - [markdown-to-html creates exactly the output file](#scenario-markdown-to-html-creates-exactly-the-output-file)
- [pandoc (document conversion filter)](#pandoc-document-conversion-filter) — 7 scenarios
  - [markdown converts to HTML and a binary docx](#scenario-markdown-converts-to-html-and-a-binary-docx)
  - [pandoc is a stdin-to-stdout filter](#scenario-pandoc-is-a-stdin-to-stdout-filter)
  - [the JSON AST is a queryable contract](#scenario-the-json-ast-is-a-queryable-contract)
  - [standalone HTML carries the metadata title](#scenario-standalone-html-carries-the-metadata-title)
  - [an unknown output format is rejected](#scenario-an-unknown-output-format-is-rejected)
  - [markdown survives a round-trip through HTML](#scenario-markdown-survives-a-round-trip-through-html)
  - [a missing input file fails cleanly](#scenario-a-missing-input-file-fails-cleanly)

## pandoc + changes (a conversion writes exactly its output)
Converting a document should produce the converted document — and nothing
besides. This suite states that exactly: with the source file present as the
baseline, a conversion creates precisely one new file, the one named on the
command line.

Anything else pandoc happened to write would fail the assertion, which is
what makes this a claim about its footprint rather than a check that the
output exists.

Source: `test/e2e/thirdparty/pandoc/changes.atago.yaml`
### Scenario: markdown-to-html creates exactly the output file
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `in.md` is created.

#### Inputs
_Fixture `in.md`:_
```text
# Title

Some text with *emphasis*.
```
#### When
```shell
pandoc in.md -o out.html
```
#### Then
- exit code is `0`
- the step changed exactly created `out.html`, modified nothing, deleted nothing
- file `out.html` contains `<em>emphasis</em>`

## pandoc (document conversion filter)
[Pandoc](https://pandoc.org/) converts documents between formats, and it can
be used two entirely different ways: as a file-to-file converter, or as a
filter in a pipeline reading stdin and writing stdout.

Both are pinned, because a script may depend on either, and they must agree:
the same input converted either way must produce the same document.

Beyond the rendered output, pandoc can emit its internal AST as JSON — a
structured oracle that says what pandoc *understood* the document to be, not
merely what it printed. Asserting on that catches a parsing regression that
happens to render plausibly.

Source: `test/e2e/thirdparty/pandoc/pandoc.atago.yaml`
### Scenario: markdown converts to HTML and a binary docx
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `doc.md` is created.

#### Inputs
_Fixture `doc.md`:_
```text
# Title

Some *emphasis* and a [link](https://example.org).

**bold text**
```
#### When
```shell
pandoc doc.md -o doc.html
pandoc doc.md -o doc.docx
unzip -p doc.docx word/document.xml
```
#### Then
- after `pandoc doc.md -o doc.html`:
  - exit code is `0`
  - file `doc.html` contains `<h1`, `<em>emphasis</em>`, `<strong>bold text</strong>`
- after `pandoc doc.md -o doc.docx`:
  - exit code is `0`
  - file `doc.docx` exists
- after `unzip -p doc.docx word/document.xml`:
  - exit code is `0`
  - stdout contains `Title`, `bold text`

#### Generated artifacts
- `doc.docx`

### Scenario: pandoc is a stdin-to-stdout filter
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `snippet.md` is created.

#### Inputs
_Fixture `snippet.md`:_
```text
**strong** and _italic_
```
_stdin for `pandoc`:_
```text
(read from file snippet.md)
```
#### When
```shell
pandoc -f markdown -t html
```
#### Then
- exit code is `0`
- stdout contains `<strong>strong</strong>`, `<em>italic</em>`

### Scenario: the JSON AST is a queryable contract
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `doc.md` is created.

#### Inputs
_Fixture `doc.md`:_
```text
# Heading One

A paragraph.
```
#### When
```shell
pandoc -t json doc.md
```
#### Then
- exit code is `0`
- stdout at `$.blocks[0].t` equals `Header`
- stdout at `$['pandoc-api-version']` has length 3

### Scenario: standalone HTML carries the metadata title
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `doc.md` is created.

#### Inputs
_Fixture `doc.md`:_
```text
# Body
```
#### When
```shell
pandoc --metadata title=Atago -s doc.md -o standalone.html
```
#### Then
- exit code is `0`
- file `standalone.html` contains `<title>Atago</title>`

### Scenario: an unknown output format is rejected
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `doc.md` is created.

#### Inputs
_Fixture `doc.md`:_
```text
# X
```
#### When
```shell
pandoc -t nosuchformat doc.md
```
#### Then
- exit code is one of `21`, `22`, `23`
- stderr contains `Unknown output format`

### Scenario: markdown survives a round-trip through HTML
_only when `pandoc --version` succeeds_
#### Given
- Fixture file `doc.md` is created.

#### Inputs
_Fixture `doc.md`:_
```text
Some *emphasis* and **bold text**.
```
#### When
```shell
pandoc doc.md -o rt.html
pandoc -f html -t markdown rt.html -o rt.md
```
#### Then
- after `pandoc doc.md -o rt.html`:
  - exit code is `0`
- after `pandoc -f html -t markdown rt.html -o rt.md`:
  - exit code is `0`
  - file `rt.md` contains `*emphasis*`, `**bold text**`

### Scenario: a missing input file fails cleanly
_only when `pandoc --version` succeeds_
#### When
```shell
pandoc no-such-file.md
```
#### Then
- exit code is not `0`
- stdout is empty
- stderr contains `does not exist`
