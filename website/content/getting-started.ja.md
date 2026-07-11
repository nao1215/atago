---
toc: true
title: はじめに
description: CLI の実行を record し、atago が生成した spec を読み、そのまま test として残します。file、snapshot、対話 prompt、server peer まで段階的に広げられます。
---

## 30 秒で試す

何かを install する前に、まずは手元の command で 1 周します。Go が入っていれば、そのまま貼り付けてください。実行を record して、その spec を test として replay します。

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped
```

`demo.atago.yaml` を開くと、`record` が終了コード、stdout の version 行、空の stderr を記録しているのが分かります。ゼロから YAML を書くのでなく、実行結果から始まる test をそのまま絞り込めます。`git --version` は `go version`、`jq --version`、`ls -la` などに置き換えてもかまいません。そのあと [atago を install](/install/) して、自分の tool へ向けます。

## 実際の実行から始める

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
atago record --pty --out wizard.atago.yaml -- mytool init
```

空の雛形から始めたいなら `atago init` でも構いません。どちらにせよ形は同じで、command を宣言し、実行し、結果を assert します。

## 1. 終了コード、stdout、stderr を検証する

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

check が fail すると、期待値と実際の差分をそのまま出します。複数行の不一致は色付き unified diff になります。

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

## 2. 生成 file と snapshot を検証する

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

全体 manifest を含む例は [files_and_fixtures](https://github.com/nao1215/atago/blob/main/examples/files_and_fixtures.atago.yaml)、[snapshot](https://github.com/nao1215/atago/blob/main/examples/snapshot.atago.yaml)、[dir_tree](https://github.com/nao1215/atago/blob/main/examples/dir_tree.atago.yaml) を参照してください。

## 3. 対話 prompt と TUI を操作する

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

named key (`send: {key: enter}`) と、描画済み terminal 画面に対する assert で本格的な TUI も扱えます。例は [pty](https://github.com/nao1215/atago/blob/main/examples/pty.atago.yaml)、[pty_screen](https://github.com/nao1215/atago/blob/main/examples/pty_screen.atago.yaml)、cross-platform の [pty_portable](https://github.com/nao1215/atago/blob/main/examples/pty_portable.atago.yaml) を参照してください。`pty` step と `atago record --pty` 自体は Linux、macOS、Windows で動き、POSIX 専用なのは `signal:` だけです。

## CLI が server と話すとき

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

## 次

[クックブック](/cookbook/) には大半の仕事向けに copyable な spec があり、各機能の runnable example へ辿れます。[CI で使う](/ci/) では GitHub Actions への載せ方までまとめています。
