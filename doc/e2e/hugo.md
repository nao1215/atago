# atago Behavior Specs
## Summary
2 suites · 10 scenarios
## Contents
- [hugo (scaffold + build CLI, tree-snapshot testbed)](#hugo-scaffold--build-cli-tree-snapshot-testbed) — 7 scenarios
  - [new site scaffolds the documented directory tree](#scenario-new-site-scaffolds-the-documented-directory-tree)
  - [new content plus a minimal layout builds the public tree](#scenario-new-content-plus-a-minimal-layout-builds-the-public-tree)
  - [building outside a site directory fails with a config hint](#scenario-building-outside-a-site-directory-fails-with-a-config-hint)
  - [scaffolding into a non-empty directory is refused until --force](#scenario-scaffolding-into-a-non-empty-directory-is-refused-until---force)
  - [creating content that already exists is refused](#scenario-creating-content-that-already-exists-is-refused)
  - [drafts stay out of the build until --buildDrafts asks for them](#scenario-drafts-stay-out-of-the-build-until---builddrafts-asks-for-them)
  - [rebuilding an unchanged site is byte-for-byte idempotent](#scenario-rebuilding-an-unchanged-site-is-byte-for-byte-idempotent)
- [hugo server (suite-wide service + http peer)](#hugo-server-suite-wide-service--http-peer) — 3 scenarios
  - [the home page lists the post](#scenario-the-home-page-lists-the-post)
  - [the post page renders its title](#scenario-the-post-page-renders-its-title)
  - [an unknown path is a 404](#scenario-an-unknown-path-is-a-404)
## hugo (scaffold + build CLI, tree-snapshot testbed)
[Hugo](https://gohugo.io/) is a scaffolder and a build tool in one binary,
and both produce their result as a directory tree rather than as output on
stdout.

That shape is what this suite is about. `hugo new site` must generate the
documented layout — the whole tree, asserted against a committed snapshot,
so an unexpected new file or a silently dropped one both show up. `hugo
--minify` must then turn content into a built site whose files exist and
contain what they should.

A generator is only trustworthy if the shape of what it generates is
trustworthy, which is why the assertion is the tree, not a spot check on one
file.

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
- dir `mysite` contains `archetypes`, contains `archetypes/default.md`, contains `assets`, contains `content`, contains `data`, contains `hugo.toml`, contains `i18n`, contains `layouts`, contains `static`, contains `themes`, has 2 entries, (recursive)
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
  - dir `mysite/public` contains `index.html`, contains `sitemap.xml`, contains `posts/hello/index.html`, matches glob `*.xml`, (recursive)
### Scenario: building outside a site directory fails with a config hint
_only when `hugo version` succeeds_
#### When
```shell
hugo
```
#### Then
- exit code is not `0`
- stderr matches `/(?i)config/`
### Scenario: scaffolding into a non-empty directory is refused until --force
_only when `hugo version` succeeds_
#### Given
- Fixture file `occupied/keep.txt` is created.
#### Inputs
_Fixture `occupied/keep.txt`:_
```text
not written by hugo
```
#### When
```shell
hugo new site occupied
hugo new site occupied --force
```
#### Then
- after `hugo new site occupied`:
  - exit code is `1`
  - stderr contains `already exists and is not empty`, `--force`
  - file `occupied/keep.txt` contains `not written by hugo`
  - dir `occupied` does not contain `hugo.toml`
- after `hugo new site occupied --force`:
  - exit code is `0`
  - stdout contains `Congratulations`
  - dir `occupied` contains `hugo.toml`, contains `keep.txt`
### Scenario: creating content that already exists is refused
_only when `hugo version` succeeds_
#### When
```shell
hugo new site mysite --quiet
hugo new content posts/dup.md
hugo new content posts/dup.md
```
#### Then
- after `hugo new content posts/dup.md`:
  - exit code is `0`
- after `hugo new content posts/dup.md`:
  - exit code is `1`
  - stderr contains `conflicts with existing content`
### Scenario: drafts stay out of the build until --buildDrafts asks for them
_only when `hugo version` succeeds_
#### Given
- Fixture file `mysite/layouts/home.html` is created.
- Fixture file `mysite/layouts/single.html` is created.
- Fixture file `mysite/layouts/list.html` is created.
#### Inputs
_Fixture `mysite/layouts/home.html`:_
```text
<!DOCTYPE html><html><body><h1>HOME</h1></body></html>
```
_Fixture `mysite/layouts/single.html`:_
```text
<html><body><h1>{{ .Title }}</h1></body></html>
```
_Fixture `mysite/layouts/list.html`:_
```text
<html><body>list</body></html>
```
#### When
```shell
hugo new site mysite --quiet
hugo new content posts/draft-post.md
hugo --quiet
hugo --quiet --buildDrafts
```
#### Then
- after `hugo new content posts/draft-post.md`:
  - exit code is `0`
  - file `mysite/content/posts/draft-post.md` contains `draft = true`
- after `hugo --quiet`:
  - exit code is `0`
  - dir `mysite/public/posts` does not contain `draft-post`
- after `hugo --quiet --buildDrafts`:
  - exit code is `0`
  - dir `mysite/public/posts` contains `draft-post`
  - file `mysite/public/posts/draft-post/index.html` contains `Draft Post`
### Scenario: rebuilding an unchanged site is byte-for-byte idempotent
_only when `hugo version` succeeds_
#### Given
- Fixture file `mysite/layouts/home.html` is created.
#### Inputs
_Fixture `mysite/layouts/home.html`:_
```text
<!DOCTYPE html><html><body><h1>HOME</h1></body></html>
```
#### When
```shell
hugo new site mysite --quiet
hugo --quiet
cd mysite && hugo --quiet
```
#### Then
- after `hugo --quiet`:
  - exit code is `0`
- after `cd mysite && hugo --quiet`:
  - exit code is `0`
  - the step changed exactly created nothing, modified nothing, deleted nothing
## hugo server (suite-wide service + http peer)
`hugo server` is the half of Hugo a developer actually looks at: the site,
served over HTTP, rendered. This suite scaffolds a site, starts the real
development server against it, and then asks for pages the way a browser
would — so what is asserted is the HTML that reached the client.

The ordering is the interesting constraint. A site must exist before a
server can serve it, so the scaffold runs first and the server starts at
exactly that point in the sequence, once, for the whole suite — the
build-then-serve bootstrap every "run my app and test it" setup needs.

Source: `test/e2e/thirdparty/hugo/hugo_server.atago.yaml`
Network policy: egress is allowed only to `127.0.0.1`.
### Suite setup (runs once before any scenario)
```shell
hugo new site mysite --quiet
# write fixture mysite/layouts/home.html
# write fixture mysite/layouts/single.html
# write fixture mysite/layouts/list.html
hugo new content posts/hello.md
# start service hugo: hugo server --source ${suitedir}/mysite --port 14140 --bind 127.0.0.1 --buildDrafts
```
### Scenario: the home page lists the post
_only when `hugo version` succeeds_
#### When
```shell
# HTTP GET / via site
```
#### Then
- HTTP status is `200`
- body contains `HOME`
- body contains `/posts/hello/`
### Scenario: the post page renders its title
_only when `hugo version` succeeds_
#### When
```shell
# HTTP GET /posts/hello/ via site
```
#### Then
- HTTP status is `200`
- body contains `<h1>Hello</h1>`
### Scenario: an unknown path is a 404
_only when `hugo version` succeeds_
#### When
```shell
# HTTP GET /no/such/page/ via site
```
#### Then
- HTTP status is `404`
