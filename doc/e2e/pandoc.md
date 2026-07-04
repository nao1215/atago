# atago Behavior Specs
## Summary
1 suite · 5 scenarios
## Contents
- [pandoc (document conversion filter)](#pandoc-document-conversion-filter) — 5 scenarios
  - [markdown converts to HTML and a binary docx](#scenario-markdown-converts-to-html-and-a-binary-docx)
  - [pandoc is a stdin-to-stdout filter](#scenario-pandoc-is-a-stdin-to-stdout-filter)
  - [the JSON AST is a queryable contract](#scenario-the-json-ast-is-a-queryable-contract)
  - [standalone HTML carries the metadata title](#scenario-standalone-html-carries-the-metadata-title)
  - [an unknown output format is rejected](#scenario-an-unknown-output-format-is-rejected)
## pandoc (document conversion filter)
Source: `test/e2e/thirdparty/pandoc/pandoc.atago.yaml`
### Scenario: markdown converts to HTML and a binary docx
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
```
#### Then
- after `pandoc doc.md -o doc.html`:
  - exit code is `0`
  - file `doc.html` contains `<h1`, `<em>emphasis</em>`, `<strong>bold text</strong>`
- after `pandoc doc.md -o doc.docx`:
  - exit code is `0`
  - file `doc.docx` exists
#### Generated artifacts
- `doc.docx`
### Scenario: pandoc is a stdin-to-stdout filter
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