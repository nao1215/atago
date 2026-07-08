# website/

The hosted documentation website, https://nao1215.github.io/atago/, built with
Hugo and deployed by `.github/workflows/website.yml` on every push to `main`
that touches `website/`, `doc/`, `schema/`, or `site/samples/`.

The committed docs stay the single source of truth. Hugo mounts them read-only
(`hugo.toml`) and `content/_content.gotmpl` turns them into pages, rewriting
repository-relative links to site pages or GitHub:

| Source | Page |
|--------|------|
| `doc/cookbook.md` + `doc/examples.md` | `/cookbook/` (one merged page: by-task index, recipes, per-feature example index) |
| `doc/real-world.md` | `/real-world/` |
| `doc/e2e/<tool>.md` | `/real-world/<tool>/` (third-party pages get a not-affiliated note) |
| `schema/atago.schema.json` | the "Spec file keys" tables on `/reference/` |
| `doc/img/`, `site/samples/` | served as static files |

Only the landing page and the Install / Getting started / CI / Reference pages
under `content/` are authored here; their prose is adapted from README.md.

The spec-key reference is rendered by the `spec-reference` shortcode
(`layouts/_shortcodes/` + `layouts/_partials/spec-row*.html`) from the JSON
Schema, so key names, types, and descriptions cannot drift from what the
loader accepts. The **Since** column comes from `data/spec_keys.json`; after
tagging a release that changed the schema, regenerate it with:

```shell
python3 website/tools/gen-spec-keys.py
```

```shell
make website        # build into website/public (requires hugo)
make website-serve  # live-reload server for local editing
```

There is no theme dependency: `layouts/` and `assets/css/` are the whole
design, and the copy button on code blocks is the only JavaScript.
`assets/css/chroma.css` is generated (`hugo gen chromastyles`, github style
light + github-dark wrapped in a dark-mode media query); regenerate it only
when changing the highlight style. `site/` (the repo-local Markdown index) is
unrelated to this directory and stays drift-guarded by `TestSite_InSync`.
