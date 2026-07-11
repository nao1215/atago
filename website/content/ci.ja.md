---
toc: true
title: CI で使う
description: atago suite を CI で回し、JUnit/JSON/TAP/GitHub report、flaky scenario の loud retry、repeat ベースの flake 検出、artifact 保持、secret masking を扱います。
---

本物の E2E suite は flake ます。タイミング、port、外部 tool 依存があるからです。`--retry-failed N` は failed scenario だけ fresh workdir で再実行し、回復したものを flaky として report します。exit code は green に保ちつつ、各 report format では loud に残します。silent retry は意図的にやりません。`--repeat N` は逆で、scenario を N 回繰り返して CI に入る前の flake 検出に使います。

```shell
atago run --ci --retry-failed 2 ./specs               # CI を green に保ちつつ不安定さは loud に出す
atago run --repeat 20 --filter "race prone" ./specs   # flake 検出
```

[setup-atago](https://github.com/nao1215/setup-atago) を使うと、released binary をそのまま GitHub Actions に入れられます。

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

GitLab CI のように container image から始める CI では、公開済みの GHCR image を使えます。

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

この image には `atago` と `ca-certificates` が入っています。scenario 内で `git`、`jq`、browser、独自 CLI binary などを使うなら、`ghcr.io/nao1215/atago:latest` を base にして必要な tool を上に足してください。

- `--report json|junit|gha|tap` で report format を選びます。JSON shape は stable かつ versioned です ([sample JSON](/samples/report.json), [JUnit](/samples/report.junit.xml), [TAP](/samples/report.tap))。
- `--ci` は deterministic かつ color-free な出力にします。同時に、空 selection を hard error に変えます。つまり `--filter` / `--tag` / `--skip-tag` が 1 scenario も選ばなければ exit 3 で落ち、typo で suite が silently disabled になるのを防ぎます。`--ci` なしでは warning を出しつつ exit 0 です。
- `--artifacts-dir DIR` は failed assert が比較した payload を保持します。failed scenario の background service log や mock server の recorded request も含むので、job 終了後も failure を review できます。
- `secrets:` に列挙した environment variable 名は、すべての report と snapshot で `***` に mask されます。

## 実行せずに spec を review する

`explain` は spec の内容を説明し、`doc` は fixture、期待 payload、golden file を inline 展開した Markdown を生成し、`manifest` は tool 向けの stable JSON summary を出し、`list` は scenario、tag、artifact を列挙します。どれも最初に spec を load して validate するため、schema error なら exit 2 で fail し、そのまま lint step にも使えます。

![atago explain and doc rendering a spec](/img/review.gif)

```shell
atago explain spec.atago.yaml
atago doc --out docs/specs.md ./specs
atago manifest ./specs
atago list ./specs
```

この site の [実在 CLI ページ](/real-world/) は `atago doc` の出力を commit し、drift test で同期を監視しています。
