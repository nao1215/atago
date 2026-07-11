---
toc: true
title: リファレンス
description: atago の subcommand、scenario 選択 flag、snapshot 更新と scrub rule、editor 向け JSON Schema、shell completion、終了コードの契約をまとめます。
---

## Subcommand

| Command | 役割 |
|---------|------|
| `atago run` | spec を実行して結果を report する |
| `atago record` | command を 1 回実行し、観測結果から spec を書く (`--pty` で対話 session も記録) |
| `atago init` | spec の雛形を生成する (`--template` で browser, cli, db, grpc, http, mock, services, ssh。既定は `cli`) |
| `atago snapshot update` | golden file を記録・更新する |
| `atago explain` | 実行せずに spec の内容を説明する |
| `atago doc` | fixture や golden file を inline 展開した Markdown を生成する |
| `atago manifest` | tooling 向けの stable JSON summary を出力する |
| `atago list` | scenario、tag、artifact を表示する |
| `atago completion` | shell completion script を出力する |

`explain`、`doc`、`manifest`、`list` はすべて最初に spec を load して validate します。schema error なら exit 2 で fail するので、そのまま lint step にも使えます。

## Scenario を選ぶ

選択系 flag はどの spec にも組み合わせられます。`--filter NAME` (repeat 可、かつ comma-separated OR に対応。`--filter a,b` または `--filter a --filter b` で scenario 名に `a` または `b` を含むものを選択)、`--tag T`、`--skip-tag T` (tag は substring ではなく完全一致。利用可能な tag は `atago list` で確認)、`--parallel N`、`--fail-fast`、`--rerun-failed` を使えます。`atago run --rerun-failed` は前回 failed だった scenario だけを `.atago/last-failed.json` から再実行するので、修正後の再確認で suite 全体を回し直さずに済みます。authoring 中は `--verbose` で pass した scenario を含め、すべての command、capture、assert verdict を trace できます。`--ci` 下では、1 件も match しない selection は empty suite として pass せず、exit 3 で落ちます。

## Snapshot testing

`snapshot` matcher は output を commit 済み golden file と比較します。ANSI color、temp path、UUID、timestamp、port、CRLF は正規化されるので、machine 間で揺れにくくなります。記録・更新は次の command です。

```shell
atago snapshot update spec.atago.yaml
```

built-in 正規化で足りない変動値、たとえば auto-increment ID、request identifier、epoch time などには、spec 全体の `scrub:` rule を宣言します。各 regex match を placeholder に書き換えてから compare します (`secrets:` masking のあとに適用)。

```yaml
scrub:
  - {pattern: 'id=\d+', placeholder: 'id=<ID>'}
```

例は [scrub example](https://github.com/nao1215/atago/blob/main/examples/scrub.atago.yaml) を参照してください。

## Spec file key

以下の表は、コミット済みの [JSON Schema](https://github.com/nao1215/atago/blob/main/schema/atago.schema.json) から生成しています。editor completion を支える authoritative source と同じものなので、loader が受け付ける shape とずれません。表の `Description` と type 名は、その schema 原文に合わせて英語表記のまま掲載しています。

spec format version は現時点で `1` だけです。`version: "1"` がすべての spec の先頭行に入ります。**追加** 列は、その key を導入した atago release を示します (`unreleased` は main に入ったが tag はまだ、の意味です)。

{{< spec-reference >}}

## Editor support (JSON Schema)

JSON Schema は [schema/atago.schema.json](https://github.com/nao1215/atago/blob/main/schema/atago.schema.json) にあります。YAML language server と組み合わせると、step type、matcher、`${workdir}` / `${env:NAME}` / `${name}` / `$${...}` の展開規則まで含めて completion と validation を得られます。`atago init` と `atago record` は生成する spec の 1 行目にこの header を入れるので、scaffold 直後から completion が効きます。既存 spec に追加する場合は、repo-relative path ではなく absolute URL を使ってください。

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/nao1215/atago/main/schema/atago.schema.json
version: "1"
```

report と manifest にも schema があります。[report.schema.json](https://github.com/nao1215/atago/blob/main/schema/report.schema.json) と [manifest.schema.json](https://github.com/nao1215/atago/blob/main/schema/manifest.schema.json) です。

## Shell completion

`atago completion <bash|zsh|fish|powershell>` は shell completion script を出力します。

## Exit code

| Code | 意味 |
|------|------|
| `0`  | すべての scenario が pass |
| `1`  | 1 件以上 fail |
| `2`  | spec error (YAML syntax または schema / semantic validation) |
| `3`  | CLI invocation error (unknown subcommand, bad flag, no matching spec file) |
| `4`  | execution error |
| `5`  | internal error |
| `6`  | security policy violation |

`Ctrl-C` / `SIGTERM` で run は clean に停止します。in-flight の process、service、session を teardown し、partial result を report して exit `4` で終了します。
