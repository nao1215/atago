<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

![Coverage](https://raw.githubusercontent.com/nao1215/octocovs-central-repo/main/badges/nao1215/atago/coverage.svg)
[![Build](https://github.com/nao1215/atago/actions/workflows/build.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/atago/actions/workflows/unit_test.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/unit_test.yml)
[![reviewdog](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml/badge.svg)](https://github.com/nao1215/atago/actions/workflows/reviewdog.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/nao1215/atago.svg)](https://pkg.go.dev/github.com/nao1215/atago)
![GitHub](https://img.shields.io/github/license/nao1215/atago)

<p align="center">
  <img src="./doc/img/atago-logo.jpg" alt="atago logo" width="400" />
</p>

**Language:** [English](./README.md) | 日本語

atago は、プレーンな YAML から CLI の実際の挙動を検証する black-box E2E テストツールです。コマンド、ファイル、snapshot、service、対話 terminal を扱い、実際の binary をそのまま動かして、ユーザーが観測する結果を assert します。テストコードも shell DSL も不要です。

Documentation: **https://nao1215.github.io/atago/ja/** · [English](https://nao1215.github.io/atago/)

![demo](./doc/img/demo.gif)

## 30 秒で試す

インストール不要です。Go が入っていれば、そのまま貼り付けてください。手元にある command を 1 回実行し、その結果を spec にして、すぐ test として再実行します。

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped
```

`record` は `git --version` を 1 回だけ実行し、終了コード、stdout の version 行、空の stderr を観測して spec を書きます。`run` はその spec を再生します。`demo.atago.yaml` を開けば、ゼロから YAML を書く代わりに、実行結果から始まる本物の test が手に入ります。`git --version` は `go version`、`jq --version`、`ls -la` などに置き換えても構いません。

自分の tool に向けるとこうなります。

```shell
atago record --out mytool.atago.yaml -- mytool convert input.txt  # 実行結果を spec にする
atago run mytool.atago.yaml                                       # spec として再生する
atago run --report junit specs/                                   # CI では suite ごと回す
```

## なぜ atago か

検証したい層ごとに道具を選びます。

| 何を検証するか | 使うもの |
|----------------|----------|
| HTTP/gRPC API server, シナリオベースの API テスト | [runn](https://github.com/k1LoW/runn) |
| HTTP, gRPC, Kafka, database などをまたぐ統合 suite | [venom](https://github.com/ovh/venom) |
| Shell 関数や script, BDD 風の unit test | [ShellSpec](https://shellspec.info/) |
| Bash script, TAP 風の test | [Bats](https://github.com/bats-core/bats-core) |
| CLI 製品そのもの, 終了コード, 出力, 生成 file, snapshot, 対話 prompt, TUI | atago |

server や platform 自体が system under test なら runn や venom を使います。atago は逆向きです。CLI が製品であり、HTTP、database、SSH、gRPC、browser、mock server はその CLI が相手にする peer として登場します。

## インストール

```shell
go install github.com/nao1215/atago@latest
```

macOS なら Homebrew でも入ります。

```shell
brew install --cask nao1215/tap/atago
```

Arch Linux なら AUR の [`atago-bin`](https://aur.archlinux.org/packages/atago-bin) を使えます。

```shell
yay -S atago-bin   # or: paru -S atago-bin
```

[release page](https://github.com/nao1215/atago/releases) には Linux、macOS、Windows 向けの prebuilt binary archive (amd64/arm64。Windows は `.zip`、それ以外は `.tar.gz`) に加え、Linux 向けの `.deb`、`.rpm`、`.apk` package があります。source build には Go 1.26 以降が必要です。

Linux、macOS、Windows で動作し、CI でも 3 つすべてを検証しています。

## はじめに

### 実際の実行から始める

最初の spec は手で書きません。`atago record -- <command>` が 1 回実行し、終了コード、出力、生成 file を観測して spec を作ります。

```shell
$ atago record --out mytool.atago.yaml -- mytool convert input.txt
recorded: exit 0, 2 stdout line(s), 1 file(s) created
wrote mytool.atago.yaml
$ atago run mytool.atago.yaml
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped (12ms)
```

対話 tool も record できます。`atago record --pty -- <command>` は本物の terminal で 1 session を操作させ、その keystroke を `pty:` step の expect/send pair として書き出します。Linux、macOS、Windows (ConPTY) で動きます。POSIX では password prompt が自動で `${env:...}` placeholder に変わり、Windows では ConPTY に echo 状態がないため secret send をあとで `${env:...}` に置き換えます。`--pty` session は `--timeout` (既定 30s) で打ち切られます。終了しない program で hang せず、capture 済みの内容を書いて fail させます。

```shell
$ atago record --pty --out wizard.atago.yaml -- mytool init
```

空の雛形から始めたいなら `atago init` でも構いません。どちらにせよ形は同じで、command を宣言し、実行し、結果を assert します。

### 1. 終了コード、stdout、stderr を検証する

```yaml
version: "1"
suite:
  name: example
scenarios:
  - name: echo greets the world
    steps:
      - run:
          shell: true            # Windows では echo が shell builtin なので portable
          command: echo "hello atago"
      - assert:
          exit_code: 0
          stdout:
            contains: atago
          stderr:
            empty: true
```

`atago run` は spec file と directory を受け取り、directory は `*.atago.yaml` を再帰的に探索します (`*.atago.yml` も可)。各 scenario は独立した temp directory で実行され、進捗は scenario ごとに 1 文字で流れます (`.` pass、`F` fail、`E` error、`s` skip)。

```shell
$ atago run ./specs
...............................................

PASSED  47 scenarios: 47 passed, 0 failed, 0 errored, 0 skipped (1.2s)
```

既定では scenario を並列実行します (`--parallel N`。既定値は CPU 数)。逐次実行したい場合は `--parallel 1` を使います。workdir は分離されますが、host network は共有です。複数 scenario がそれぞれ `service:` を起動するなら、port は分けてください。

失敗時は、期待値と実際の差分をそのまま出します。複数行の不一致は色付き unified diff になります。

```text
FAILED: demo / greeting matches its golden

Step:
  assert stdout snapshot

Diff (-expected +actual):
  --- snapshot (golden)
  +++ actual
  @@ -1,3 +1,3 @@
   hello
  -WORLD
  +world
   bye

Hint:
  stdout did not match snapshot "snaps/greeting.txt" (update with --update-snapshots if intended)
```

### 2. 生成 file と snapshot を検証する

`fixture:` は isolated workdir に input file を配置します。`file:` / `dir:` assert は command が生成したものを検証し、`snapshot:` は output を commit 済み golden file に固定します。fixture の source は `content:` (inline text)、`base64:` (inline bytes)、`from:` (既存 file の copy)、`symlink:` (target への link) のいずれかです。

```yaml
scenarios:
  - name: the generator writes the expected files
    steps:
      - run:
          command: mytool generate --out site
      - assert:
          file:
            path: site/index.html
            contains:
              - "<html"
      - assert:
          stdout:
            snapshot: snapshots/generate.txt   # `atago snapshot update` で記録/更新
```

例は [files_and_fixtures](examples/files_and_fixtures.atago.yaml)、[snapshot](examples/snapshot.atago.yaml)、[dir_tree](examples/dir_tree.atago.yaml) を参照してください。

### 3. 対話 prompt と TUI を操作する

`pty` step は command を本物の pseudo-terminal で起動し、宣言的な expect/send session で操作します。wizard、REPL、TTY 判定分岐まで `expect(1)` script なしで扱えます。

```yaml
scenarios:
  - name: the init wizard completes
    steps:
      - pty:
          command: mytool init
          session:
            - expect: "Project name:"
            - send: "demo\n"
            - expect: "created demo/"
      - assert:
          exit_code: 0
```

named key (`send: {key: enter}`) と、描画済み terminal 画面に対する assert で本格的な TUI も扱えます。例は [pty](examples/pty.atago.yaml)、[pty_screen](examples/pty_screen.atago.yaml)、cross-platform の [pty_portable](examples/pty_portable.atago.yaml) を参照してください。`pty` step と `atago record --pty` 自体は Linux、macOS、Windows で動き、POSIX 専用なのは `signal:` だけです。

### CLI が server と話すとき

同じ YAML で、HTTP、database、SSH、gRPC、headless browser、offline mock server も、CLI の依存先 peer として扱えます。`atago init --template <name>` でそれぞれの雛形を作れます。

```shell
$ atago init --list-templates
browser   drive a headless Chrome; assert page content (needs Chrome on PATH)
cli       run a command; assert exit code/stdout/stderr (runs as-is)
db        run SQL; assert on rows via bundled SQLite (runs as-is)
grpc      call a unary gRPC method via server reflection (edit target first)
http      call an HTTP API; assert status and JSON body (edit base_url first)
mock      stub an HTTP API offline and assert what the client sent (needs curl on PATH)
services  test against a background server: readiness, retry, teardown (runs as-is)
ssh       run a command on a remote host over SSH (edit host/user first)
```

## Example

全機能に対応する runnable example が [examples/](examples/) にあります。CI でも Linux、macOS、Windows で検証しています。copyable な recipe は [cookbook](https://nao1215.github.io/atago/ja/cookbook/) にまとめています。

scenario 選択 flag はどの spec にも組み合わせられます。`--filter NAME` (repeat 可、comma-separated OR 対応)、`--tag T`、`--skip-tag T`、`--parallel N`、`--fail-fast`、`--rerun-failed` を使えます。authoring 中は `--verbose` ですべての command、capture、assert verdict を trace できます。

## CI で使う

本物の E2E suite は flake ます。タイミング、port、外部 tool 依存があるからです。`--retry-failed N` は failed scenario だけ fresh workdir で再実行し、回復したものを flaky として report します。`--repeat N` は逆に flake 検出へ使います。

```shell
atago run --ci --retry-failed 2 ./specs          # CI を green に保ちつつ不安定さは loud に出す
atago run --repeat 20 --filter "race prone" ./specs
```

[setup-atago](https://github.com/nao1215/setup-atago) は released binary を GitHub Actions に入れます。

```yaml
name: behavior-specs
on: [push, pull_request]
jobs:
  atago:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: nao1215/setup-atago@v0
      - run: atago run --ci --report gha ./specs
```

container image から始める CI では、公開済みの GHCR image を使えます。

```yaml
image: ghcr.io/nao1215/atago:latest

stages: [test]

behavior-specs:
  stage: test
  script:
    - atago run --ci --report junit ./specs > junit.xml
  artifacts:
    when: always
    reports:
      junit: junit.xml
```

- `--report json|junit|gha|tap` で report format を選びます。JSON shape は stable かつ versioned です。
- `--ci` は deterministic かつ color-free な出力にし、空 selection を hard error (exit 3) に変えます。
- `--artifacts-dir DIR` は failed assert の比較 payload、background service log、mock server request を保持します。
- `secrets:` に列挙した environment variable 名は、すべての report と snapshot で `***` に mask されます。

## 実行せずに spec をレビューする

`explain` は spec の内容を説明し、`doc` は fixture、期待 payload、golden file を inline 展開した Markdown を生成し、`manifest` は tool 向けの stable JSON summary を出し、`list` は scenario、tag、artifact を列挙します。どれも最初に spec を load して validate するため、schema error なら exit 2 で fail し、そのまま lint step にも使えます。

![review](./doc/img/review.gif)

```shell
atago explain spec.atago.yaml
atago doc --out docs/specs.md ./specs
atago manifest ./specs
atago list ./specs
```

## Snapshot testing

`snapshot` matcher は output を commit 済み golden file と比較します。ANSI color、temp path、UUID、timestamp、port、CRLF は正規化されるので、machine 間で揺れにくくなります。記録・更新は次の command です。

![snapshot](./doc/img/snapshot.gif)

```shell
atago snapshot update spec.atago.yaml
```

built-in 正規化で足りない変動値には、spec 全体の `scrub:` rule を宣言します。

```yaml
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
```

例は [scrub](examples/scrub.atago.yaml) を参照してください。

## Editor support (JSON Schema)

JSON Schema は [schema/atago.schema.json](schema/atago.schema.json) にあります。YAML language server と組み合わせると、step type、matcher、`${workdir}` / `${env:NAME}` / `${name}` / `$${...}` の展開規則まで含めて completion と validation を得られます。`atago init` と `atago record` は生成する spec の 1 行目にこの header を入れるので、scaffold 直後から completion が効きます。既存 spec に追加する場合は absolute URL を使ってください。

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
```

## Shell completion

`atago completion <bash|zsh|fish|powershell>` は shell completion script を出力します。

## Exit codes

| Code | Meaning |
|------|---------|
| `0`  | すべての scenario が pass |
| `1`  | 1 件以上 fail |
| `2`  | spec error (YAML syntax または schema / semantic validation) |
| `3`  | CLI invocation error (unknown subcommand, bad flag, no matching spec file) |
| `4`  | execution error |
| `5`  | internal error |
| `6`  | security policy violation |

`Ctrl-C` / `SIGTERM` で run は clean に停止します。in-flight の process、service、session を teardown し、partial result を report して exit `4` で終了します。

## atago で検証している実在 CLI

これらの suite は、著者の Go tool (atago 自身を含む) から、未改変の third-party binary まで、本当にさまざまな program を動かします。git や jq、対話 TUI (fzf、htop)、python3 REPL、scenario service として起動する server (redis、gitea、grafana、prometheus)、offline で検証する cloud / IaC CLI (aws-cli、terraform、ecspresso)、crypto tool (openssl、age、sops)、document / media pipeline (pandoc、ffmpeg) などです。多くは ShellSpec から移行してきました。[Real CLIs tested with atago](https://nao1215.github.io/atago/ja/real-world/) に 40 以上の一覧と behavior docs があります。

## 名前

atago (愛宕) は、京都の愛宕山に由来します。愛宕神社が火伏せの神を祀るように、atago も project が炎上する前に火種を止める道具でありたい、という意図です。

## Contributing

Issue と pull request を歓迎します。詳細は [CONTRIBUTING.md](./CONTRIBUTING.md) を参照してください。code contribution だけでなく、GitHub Star も開発の継続に効きます。

## LICENSE

atago project は [MIT LICENSE](./LICENSE) の条件で提供されます。

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://debimate.jp/"><img src="https://avatars.githubusercontent.com/u/22737008?v=4?s=75" width="75px;" alt="CHIKAMATSU Naohiro"/><br /><sub><b>CHIKAMATSU Naohiro</b></sub></a><br /><a href="https://github.com/nao1215/atago/commits?author=nao1215" title="Code">💻</a> <a href="https://github.com/nao1215/atago/commits?author=nao1215" title="Documentation">📖</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
