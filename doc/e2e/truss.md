# atago Behavior Specs
## Summary
2 suites · 9 scenarios
## Contents
- [truss convert (filesystem footprint)](#truss-convert-filesystem-footprint) — 2 scenarios
  - [convert PNG->JPEG creates only the output file](#scenario-convert-png-jpeg-creates-only-the-output-file)
  - [convert to a glob-matched output honors path.Match](#scenario-convert-to-a-glob-matched-output-honors-pathmatch)
- [truss image conversion](#truss-image-conversion) — 7 scenarios
  - [inspect reports PNG format and dimensions as JSON](#scenario-inspect-reports-png-format-and-dimensions-as-json)
  - [convert PNG to JPEG yields a same-size opaque JPEG](#scenario-convert-png-to-jpeg-yields-a-same-size-opaque-jpeg)
  - [resize with fit=fill produces the requested dimensions](#scenario-resize-with-fitfill-produces-the-requested-dimensions)
  - [convert to WebP yields a WebP of the same dimensions](#scenario-convert-to-webp-yields-a-webp-of-the-same-dimensions)
  - [a high-quality JPEG stays visually close to the source](#scenario-a-high-quality-jpeg-stays-visually-close-to-the-source)
  - [a lossless PNG re-encode is pixel-identical to the source](#scenario-a-lossless-png-re-encode-is-pixel-identical-to-the-source)
  - [a missing input file exits with the I/O error code](#scenario-a-missing-input-file-exits-with-the-io-error-code)
## truss convert (filesystem footprint)
Source: `test/e2e/tools/truss/changes.atago.yaml`
### Scenario: convert PNG->JPEG creates only the output file
_only when `truss --version` succeeds · skipped on Windows_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o out.jpg
```
#### Then
- exit code is `0`
- the step changed exactly created `out.jpg`, modified nothing, deleted nothing
- image `out.jpg` is `jpeg`, width 2, height 2
#### Generated artifacts
- `out.jpg`
### Scenario: convert to a glob-matched output honors path.Match
_only when `truss --version` succeeds · skipped on Windows_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o thumb.webp --format webp
```
#### Then
- exit code is `0`
- the step changed exactly created `*.webp`, modified nothing, deleted nothing
## truss image conversion
Source: `test/e2e/tools/truss/convert.atago.yaml`
### Scenario: inspect reports PNG format and dimensions as JSON
_only when `truss --version` succeeds_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss inspect in.png
```
#### Then
- exit code is `0`
- stdout at `$.format` equals `png`
- stdout at `$.width` equals `16`
- stdout at `$.height` equals `16`
### Scenario: convert PNG to JPEG yields a same-size opaque JPEG
_only when `truss --version` succeeds_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o out.jpg
```
#### Then
- exit code is `0`
- image `out.jpg` is `jpeg`, width 16, height 16, has no alpha
#### Generated artifacts
- `out.jpg`
### Scenario: resize with fit=fill produces the requested dimensions
_only when `truss --version` succeeds_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o thumb.png --width 8 --height 8 --fit fill
```
#### Then
- exit code is `0`
- image `thumb.png` is `png`, width 8, height 8
#### Generated artifacts
- `thumb.png`
### Scenario: convert to WebP yields a WebP of the same dimensions
_only when `truss --version` succeeds_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o out.webp --format webp
```
#### Then
- exit code is `0`
- image `out.webp` is `webp`, width 16, height 16
#### Generated artifacts
- `out.webp`
### Scenario: a high-quality JPEG stays visually close to the source
_only when `truss --version` succeeds_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o out.jpg --quality 100
```
#### Then
- exit code is `0`
- image `out.jpg` similar to `testdata/sample.png`
#### Expected output
_expected image `testdata/sample.png`:_
![expected image `testdata/sample.png`](../../test/e2e/tools/truss/testdata/sample.png)
#### Generated artifacts
- `out.jpg`
### Scenario: a lossless PNG re-encode is pixel-identical to the source
_only when `truss --version` succeeds_
#### Given
- Fixture file `in.png` is created.
#### When
```shell
truss convert in.png -o copy.png
```
#### Then
- exit code is `0`
- image `copy.png` similar to `testdata/sample.png`
#### Expected output
_expected image `testdata/sample.png`:_
![expected image `testdata/sample.png`](../../test/e2e/tools/truss/testdata/sample.png)
#### Generated artifacts
- `copy.png`
### Scenario: a missing input file exits with the I/O error code
_only when `truss --version` succeeds_
#### When
```shell
truss convert does-not-exist.png -o out.jpg
```
#### Then
- exit code is `2`
- stderr contains `does-not-exist.png`
