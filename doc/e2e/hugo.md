# atago Behavior Specs
## Summary
2 suites · 6 scenarios
## Contents
- [hugo (scaffold + build CLI, tree-snapshot testbed)](#hugo-scaffold--build-cli-tree-snapshot-testbed) — 3 scenarios
  - [new site scaffolds the documented directory tree](#scenario-new-site-scaffolds-the-documented-directory-tree)
  - [new content plus a minimal layout builds the public tree](#scenario-new-content-plus-a-minimal-layout-builds-the-public-tree)
  - [building outside a site directory fails with a config hint](#scenario-building-outside-a-site-directory-fails-with-a-config-hint)
- [hugo server (suite-wide service + http peer)](#hugo-server-suite-wide-service--http-peer) — 3 scenarios
  - [the home page lists the post](#scenario-the-home-page-lists-the-post)
  - [the post page renders its title](#scenario-the-post-page-renders-its-title)
  - [an unknown path is a 404](#scenario-an-unknown-path-is-a-404)
## hugo (scaffold + build CLI, tree-snapshot testbed)
Source: `test/e2e/thirdparty/hugo/hugo.atago.yaml`
### Scenario: new site scaffolds the documented directory tree
_only when `hugo version` succeeds_
#### When
```shell
hugo new site mysite
```
#### Then
- exit code is `0`
- stdout contains `Congratulations`
- dir `mysite` contains `archetypes`, contains `content`, contains `layouts`, contains `static`, contains `themes`
- file `mysite/hugo.toml` contains `baseURL`
### Scenario: new content plus a minimal layout builds the public tree
_only when `hugo version` succeeds_
#### Given
- Fixture file `mysite/layouts/home.html` is created.
- Fixture file `mysite/layouts/single.html` is created.
- Fixture file `mysite/layouts/list.html` is created.
#### Inputs
_Fixture `mysite/layouts/home.html`:_
```text
<!DOCTYPE html><html><body><h1>HOME</h1>{{ range site.RegularPages }}<a href="{{ .RelPermalink }}">{{ .Title }}</a>{{ end }}</body></html>
```
_Fixture `mysite/layouts/single.html`:_
```text
<!DOCTYPE html><html><body><article><h1>{{ .Title }}</h1>{{ .Content }}</article></body></html>
```
_Fixture `mysite/layouts/list.html`:_
```text
<html><body>{{ range .Pages }}{{ .Title }}{{ end }}</body></html>
```
#### When
```shell
hugo new site mysite --quiet
hugo new content posts/hello.md
hugo --minify --buildDrafts
```
#### Then
- after `hugo new content posts/hello.md`:
  - exit code is `0`
  - file `mysite/content/posts/hello.md` contains `draft`
- after `hugo --minify --buildDrafts`:
  - exit code is `0`
  - file `mysite/public/index.html` contains `HOME`
  - file `mysite/public/index.html` contains `/posts/hello/`
  - file `mysite/public/posts/hello/index.html` contains `<h1>Hello</h1>`
  - dir `mysite/public` contains `index.html`, contains `sitemap.xml`, contains `posts/hello/index.html`
### Scenario: building outside a site directory fails with a config hint
_only when `hugo version` succeeds_
#### When
```shell
hugo
```
#### Then
- exit code is not `0`
- stderr matches `/(?i)config/`
## hugo server (suite-wide service + http peer)
Source: `test/e2e/thirdparty/hugo/hugo_server.atago.yaml`
### Scenario: the home page lists the post
#### When
```shell
# HTTP GET /
```
#### Then
- HTTP status is `200`
- body contains `HOME`
- body contains `/posts/hello/`
### Scenario: the post page renders its title
#### When
```shell
# HTTP GET /posts/hello/
```
#### Then
- HTTP status is `200`
- body contains `<h1>Hello</h1>`
### Scenario: an unknown path is a 404
#### When
```shell
# HTTP GET /no/such/page/
```
#### Then
- HTTP status is `404`
