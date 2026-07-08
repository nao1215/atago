# website/

The hosted documentation website, https://nao1215.github.io/atago/, built with
Hugo and deployed by `.github/workflows/website.yml` on every push to `main`
that touches `website/`, `doc/`, or `site/samples/`.

The committed docs stay the single source of truth. Hugo mounts them read-only
(`hugo.toml`) and `content/_content.gotmpl` turns them into pages, rewriting
repository-relative links to site pages or GitHub:

| Source | Page |
|--------|------|
| `doc/cookbook.md` | `/cookbook/` |
| `doc/examples.md` | `/examples/` |
| `doc/real-world.md` | `/real-world/` |
| `doc/e2e/<tool>.md` | `/real-world/<tool>/` |
| `doc/img/`, `site/samples/` | served as static files |

Only the landing page and the Install / Getting started / CI / Reference /
LLM pages under `content/` are authored here; their prose is adapted from
README.md. `/llms.txt` (a plain-text page index for LLM agents) is generated
by `layouts/home.llms.txt` via the custom `llms` output format.

```shell
make website        # build into website/public (requires hugo)
make website-serve  # live-reload server for local editing
```

There is no theme dependency: `layouts/` and `assets/css/` are the whole
design. `assets/css/chroma.css` is generated (`hugo gen chromastyles`, github
style light + github-dark wrapped in a dark-mode media query); regenerate it
only when changing the highlight style. `site/` (the repo-local Markdown index)
is unrelated to this directory and stays drift-guarded by `TestSite_InSync`.
