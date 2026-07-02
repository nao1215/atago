# atago Behavior Specs
## Summary
1 suite · 7 scenarios
## Contents
- [truss image conversion](#truss-image-conversion) — 7 scenarios
  - [inspect reports PNG format and dimensions as JSON](#scenario-inspect-reports-png-format-and-dimensions-as-json)
  - [convert PNG to JPEG yields a same-size opaque JPEG](#scenario-convert-png-to-jpeg-yields-a-same-size-opaque-jpeg)
  - [resize with fit=fill produces the requested dimensions](#scenario-resize-with-fitfill-produces-the-requested-dimensions)
  - [convert to WebP yields a WebP of the same dimensions](#scenario-convert-to-webp-yields-a-webp-of-the-same-dimensions)
  - [a high-quality JPEG stays visually close to the source](#scenario-a-high-quality-jpeg-stays-visually-close-to-the-source)
  - [a lossless PNG re-encode is pixel-identical to the source](#scenario-a-lossless-png-re-encode-is-pixel-identical-to-the-source)
  - [a missing input file exits with the I/O error code](#scenario-a-missing-input-file-exits-with-the-io-error-code)
## truss image conversion
Source: `test/e2e/tools/truss/convert.atago.yaml`
### Scenario: inspect reports PNG format and dimensions as JSON
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
#### When
```shell
truss convert does-not-exist.png -o out.jpg
```
#### Then
- exit code is `2`
- stderr contains `does-not-exist.png`