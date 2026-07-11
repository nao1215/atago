---
title: atago
description: atago はプレーンな YAML から CLI の実挙動を検証します。コマンド、ファイル、スナップショット、サービス、対話端末をそのまま扱えます。
---

atago はプレーンな YAML から CLI の実際の挙動を検証します。コマンド、ファイル、スナップショット、サービス、対話端末を扱い、実際のバイナリをそのまま動かして、ユーザーが観測する結果をアサートします。テストコードも shell DSL も不要です。

![atago running a spec suite](/img/demo.gif)

## 30 秒で試す

何もインストールせず、手元のコマンドでまず一周させます。Go が入っていれば、そのまま貼り付けてください。実行結果を記録し、その spec をその場で再実行します。

```shell
go run github.com/nao1215/atago@latest record --out demo.atago.yaml -- git --version
go run github.com/nao1215/atago@latest run demo.atago.yaml
```

```text
.

PASSED  1 scenario: 1 passed, 0 failed, 0 errored, 0 skipped
```

`record` は `git --version` を 1 回だけ実行し、終了コード、stdout の version 行、空の stderr を観測して spec にします。`run` はそれを再生します。`demo.atago.yaml` を開けば、ゼロから YAML を書く代わりに、実行結果から始まる本物の test が手に入ります。`git --version` の代わりに `go version`、`jq --version`、`ls -la` でもかまいません。

次は自分のツールに向けます。

```shell
atago record --out mytool.atago.yaml -- mytool convert input.txt  # 実行結果を spec にする
atago run mytool.atago.yaml                                       # spec として再生する
atago run --report junit specs/                                   # CI では suite ごと回す
```

## なぜ atago か

自分が検証したい層に合わせて道具を選びます。

| 何を検証するか | 使うもの |
|----------------|----------|
| HTTP/gRPC API サーバー, シナリオベースの API テスト | [runn](https://github.com/k1LoW/runn) |
| HTTP, gRPC, Kafka, DB などをまたぐ統合 suite | [venom](https://github.com/ovh/venom) |
| Shell 関数や script, BDD 風の unit test | [ShellSpec](https://shellspec.info/) |
| Bash script, TAP 風の test | [Bats](https://github.com/bats-core/bats-core) |
| CLI 製品そのもの, 終了コード, 出力, 生成ファイル, snapshot, 対話 prompt, TUI | atago |

server や platform 自体が test 対象なら runn や venom を使います。atago は逆向きです。CLI が製品であり、HTTP、DB、SSH、gRPC、browser、mock server はその CLI が相手にする peer として登場します。

## 書く前に record する

最初の spec は手書きしません。`atago record -- <command>` が 1 回実行し、終了コード、出力、生成ファイルを観測して spec にします。対話ツールも `record --pty` で記録でき、1 session 手で操作すると、keystroke が replayable な expect/send script になります。password prompt は自動で `${env:...}` placeholder へ mask されます。つまり最初から YAML を書くのでなく、record した YAML を絞り込んでいきます。

## CLI 出力向けの snapshot

![snapshot testing: failure diff and one-command update](/img/snapshot.gif)

`snapshot:` は output を commit 済み golden file に固定します。ANSI color、temp path、UUID、timestamp、port、CRLF は正規化されるので、環境差分で golden が揺れにくくなります。自前の変動値には spec 全体の `scrub:` rule を追加します。失敗時は unified diff が色付きで出て、`atago snapshot update` で意図的な変更だけを更新できます。review はそのまま `git diff snapshots/` です。

## 擬似化しない本物の terminal

`pty:` step は本物の pseudo-terminal を起動し、宣言的な expect/send pair と named key で駆動します。Windows でも ConPTY 経由で動きます。`screen:` は cursor 移動や clear を適用した後の、ユーザーが実際に見る描画結果を assert します。TTY 判定分岐、wizard、REPL、htop のような dashboard まで spec 化できます。

## 名前

atago (愛宕) は、京都の愛宕山に由来します。愛宕神社が火伏せの神を祀るように、atago も project が炎上する前に火種を止める道具でありたい、という意図です。Go 製なので名前の末尾が *go* で終わるのも都合がいいところです。

## 次に読む

- [インストール](/install/) — `go install`、Homebrew、AUR、配布 binary、GitHub Action。
- [はじめに](/getting-started/) — 実際の実行を record し、生成された spec を読み、test として残す。
- [クックブック](/cookbook/) — よくある仕事向けの copyable recipe と、各機能の runnable spec。
- [CI で使う](/ci/) — report format、retry、flake 検出、artifact、secret masking。
- [リファレンス](/reference/) — subcommand、終了コード、spec key 一覧。
- [atago で検証している実在 CLI](/real-world/) — jq から terraform、htop まで 40 以上の tool に対して実行している spec と behavior docs。

この site の内容はすべて [repository](https://github.com/nao1215/atago) に commit された file から生成されています。behavior docs は実行可能 spec から再生成され、drift test が spec と doc の不一致を検知します。ここに載っている例は、実際に回っています。
