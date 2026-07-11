---
title: インストール
description: atago を `go install`、Homebrew、AUR、Linux/macOS/Windows 向け配布 binary、または setup-atago GitHub Action で導入できます。検証方法も示します。
---

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

GitHub Actions では、[setup-atago](https://github.com/nao1215/setup-atago) が release binary を導入します。使い方は [CI で使う](/ci/) を参照してください。

## Release の整合性を検証する

各 release には、ダウンロード物を検証するための supply-chain metadata を添付しています。

- Signed checksums: `checksums.txt` は [cosign](https://github.com/sigstore/cosign) の keyless signing を使い、`checksums.txt.sigstore.json` を生成します。
- SBOM: 各 release archive に SPDX の Software Bill of Materials を添付します。
- Build provenance: GitHub OIDC を使った SLSA build provenance を付与します。

```shell
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/nao1215/atago/\.github/workflows/release\.yml@refs/tags/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

```shell
gh attestation verify atago_<version>_<os>_<arch>.tar.gz --repo nao1215/atago  # Windows は .zip
```
