# atago Behavior Specs
## Summary
1 suite · 10 scenarios
## Contents
- [Ghostscript (PostScript and PDF pipeline)](#ghostscript-postscript-and-pdf-pipeline) — 10 scenarios
  - [PostScript becomes a PDF with the pages and text it declared](#scenario-postscript-becomes-a-pdf-with-the-pages-and-text-it-declared)
  - [a DOCINFO pdfmark reaches the Info dictionary](#scenario-a-docinfo-pdfmark-reaches-the-info-dictionary)
  - [selecting one page keeps that page's text and drops the other](#scenario-selecting-one-page-keeps-that-pages-text-and-drops-the-other)
  - [merging two documents yields the sum of their pages](#scenario-merging-two-documents-yields-the-sum-of-their-pages)
  - [txtwrite agrees with the text inside the PDF](#scenario-txtwrite-agrees-with-the-text-inside-the-pdf)
  - [rasterizing a page produces an image of the requested size](#scenario-rasterizing-a-page-produces-an-image-of-the-requested-size)
  - [converting the same input twice yields the same text](#scenario-converting-the-same-input-twice-yields-the-same-text)
  - [a missing input fails, naming the file on stdout](#scenario-a-missing-input-fails-naming-the-file-on-stdout)
  - [a failed conversion still leaves a PDF behind](#scenario-a-failed-conversion-still-leaves-a-pdf-behind)
  - [an unknown device is refused](#scenario-an-unknown-device-is-refused)
## Ghostscript (PostScript and PDF pipeline)
[Ghostscript](https://www.ghostscript.com/) turns PostScript into PDF, PDF
into text, and PDF into raster images. Its output is the kind a test usually
gives up on — a binary document — so this suite asserts the document itself:
page count, the text really inside it, and the Info-dictionary metadata,
with `pdf:`; and the rasterized page with `image:`.

The relationships are pinned, not just single outputs: selecting one page of
a two-page document yields exactly that page's text and not the other's,
merging two documents yields the sum of their pages, and converting the same
input twice yields the same extracted text. The input PostScript is written
by the spec, so nothing third-party is committed.

One honest surprise is recorded as a scenario: a failed conversion still
leaves a PDF behind. That is Ghostscript's behavior, not a recommendation.

Source: `test/e2e/thirdparty/ghostscript/ghostscript.atago.yaml`
### Scenario: PostScript becomes a PDF with the pages and text it declared
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
- The command runs with an isolated home under `${workdir}/.atago-home` (HOME/XDG or APPDATA redirected).
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (Quarterly earnings report) show
showpage
/Helvetica findfont 18 scalefont setfont
72 700 moveto (Appendix and totals) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf doc.ps
```
#### Then
- exit code is `0`
- the step changed exactly created `out.pdf`, modified nothing, deleted nothing
- pdf `out.pdf` 2 pages
- pdf `out.pdf` text contains `Quarterly earnings report`, `Appendix and totals`
### Scenario: a DOCINFO pdfmark reaches the Info dictionary
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (Report body) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o titled.pdf -c "[/Title (Quarterly Report) /Author (atago) /DOCINFO pdfmark" -f doc.ps
```
#### Then
- exit code is `0`
- pdf `titled.pdf` author contains `atago`, title contains `Quarterly Report`
- pdf `titled.pdf` producer contains `Ghostscript`
### Scenario: selecting one page keeps that page's text and drops the other
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (First page marker) show
showpage
72 720 moveto (Second page marker) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o whole.pdf doc.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -dFirstPage=2 -dLastPage=2 -o second.pdf whole.pdf
```
#### Then
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o whole.pdf doc.ps`:
  - exit code is `0`
  - pdf `whole.pdf` 2 pages
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -dFirstPage=2 -dLastPage=2 -o second.pdf whole.pdf`:
  - exit code is `0`
  - pdf `second.pdf` 1 pages
  - pdf `second.pdf` text contains `Second page marker`
  - pdf `second.pdf` text does not contain `First page marker`
### Scenario: merging two documents yields the sum of their pages
_only when `gs --version` succeeds_
#### Given
- Fixture file `one.ps` is created.
- Fixture file `two.ps` is created.
#### Inputs
_Fixture `one.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (Document one) show
showpage
```
_Fixture `two.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (Document two page A) show
showpage
72 700 moveto (Document two page B) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o one.pdf one.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o two.pdf two.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o merged.pdf one.pdf two.pdf
```
#### Then
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o one.pdf one.ps`:
  - exit code is `0`
  - pdf `one.pdf` 1 pages
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o two.pdf two.ps`:
  - exit code is `0`
  - pdf `two.pdf` 2 pages
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o merged.pdf one.pdf two.pdf`:
  - exit code is `0`
  - pdf `merged.pdf` 3 pages
  - pdf `merged.pdf` text contains `Document one`, `Document two page A`, `Document two page B`
### Scenario: txtwrite agrees with the text inside the PDF
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (Extractable sentence) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf doc.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=txtwrite -o extracted.txt out.pdf
```
#### Then
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf doc.ps`:
  - exit code is `0`
  - pdf `out.pdf` text contains `Extractable sentence`
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=txtwrite -o extracted.txt out.pdf`:
  - exit code is `0`
  - file `extracted.txt` contains `Extractable sentence`
### Scenario: rasterizing a page produces an image of the requested size
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
<< /PageSize [200 100] >> setpagedevice
/Helvetica findfont 12 scalefont setfont
20 50 moveto (Raster me) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf doc.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=png16m -r72 -o page.png out.pdf
gs -q -dNOPAUSE -dBATCH -sDEVICE=png16m -r144 -o page2x.png out.pdf
```
#### Then
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf doc.ps`:
  - exit code is `0`
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=png16m -r72 -o page.png out.pdf`:
  - exit code is `0`
  - image `page.png` is `png`, width 200, height 100
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=png16m -r144 -o page2x.png out.pdf`:
  - exit code is `0`
  - image `page2x.png` width 400, height 200
#### Generated artifacts
- `page.png`
- `page2x.png`
### Scenario: converting the same input twice yields the same text
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (Deterministic content) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o first.pdf doc.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o second.pdf doc.ps
gs -q -dNOPAUSE -dBATCH -sDEVICE=txtwrite -o first.txt first.pdf
gs -q -dNOPAUSE -dBATCH -sDEVICE=txtwrite -o second.txt second.pdf
```
#### Then
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o first.pdf doc.ps`:
  - exit code is `0`
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o second.pdf doc.ps`:
  - exit code is `0`
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=txtwrite -o first.txt first.pdf`:
  - exit code is `0`
- after `gs -q -dNOPAUSE -dBATCH -sDEVICE=txtwrite -o second.txt second.pdf`:
  - exit code is `0`
  - file `second.txt` is byte-identical to `first.txt`
### Scenario: a missing input fails, naming the file on stdout
_only when `gs --version` succeeds_
#### When
```shell
gs -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf no-such-input.ps
```
#### Then
- exit code is `1`
- stdout contains `undefinedfilename`, `no-such-input.ps`
- stderr contains `Unrecoverable error`
### Scenario: a failed conversion still leaves a PDF behind
_only when `gs --version` succeeds_
#### Given
- Fixture file `broken.ps` is created.
#### Inputs
_Fixture `broken.ps`:_
```text
this is not postscript {{{
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=pdfwrite -o out.pdf broken.ps
```
#### Then
- exit code is `1`
- stderr contains `Unrecoverable error`
- pdf `out.pdf` 1 pages
- pdf `out.pdf` text does not contain `this is not postscript`
### Scenario: an unknown device is refused
_only when `gs --version` succeeds_
#### Given
- Fixture file `doc.ps` is created.
#### Inputs
_Fixture `doc.ps`:_
```text
%!PS-Adobe-3.0
/Helvetica findfont 24 scalefont setfont
72 720 moveto (body) show
showpage
```
#### When
```shell
gs -q -dNOPAUSE -dBATCH -sDEVICE=notadevice -o out.any doc.ps
```
#### Then
- exit code is not `0`
- stdout contains `Unknown device: notadevice`
