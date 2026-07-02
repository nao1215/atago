# Samples

These artifacts are generated deterministically by `internal/sitegen` (run
`make site`) and drift-guarded by `TestSite_InSync`.

- `report.json` / `report.junit.xml` / `report.tap` — sample outputs of
  `atago run --report <format>`. They are built from a fixed result set with
  **all durations set to zero** so the committed files are byte-stable (a real run
  would report real durations).
- `sample.pdf` — a tiny generated PDF, the kind the `pdf` assertion inspects.
- `imagediff/` — a baseline image, a one-pixel-changed actual image, and a
  per-pixel difference heatmap, the kind produced for an image `similar_to`
  failure.
