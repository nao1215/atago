---
title: Install
description: Install atago with go install, Homebrew, the AUR, prebuilt binaries for Linux/macOS/Windows, or the setup-atago GitHub Action — and verify what you download.
---

```shell
go install github.com/nao1215/atago@latest
```

On macOS, Homebrew works too:

```shell
brew install --cask nao1215/tap/atago
```

On Arch Linux, install the [`atago-bin`](https://aur.archlinux.org/packages/atago-bin) package from the AUR:

```shell
yay -S atago-bin   # or: paru -S atago-bin
```

atago is in the [aqua](https://aquaproj.github.io/) standard registry. Add it to your `aqua.yaml`, then install it:

```shell
aqua g -i nao1215/atago
aqua i
```

[mise](https://mise.jdx.dev/) can install it through the same registry with the aqua backend:

```shell
mise use -g aqua:nao1215/atago
```

If your installed `mise` release does not see `nao1215/atago` yet, update `mise` or enable floating registries with `mise settings registry_floating=true`.

The [release page](https://github.com/nao1215/atago/releases) contains prebuilt binary archives for Linux, macOS, and Windows (amd64/arm64; `.tar.gz`, or `.zip` on Windows), plus `.deb`, `.rpm`, and `.apk` packages for Linux. Requires Go 1.26 or later when building from source.

Runs on Linux, macOS, and Windows (CI tests all three).

In GitHub Actions, [setup-atago](https://github.com/nao1215/setup-atago) installs a released binary — see [Use it in CI](/ci/).

## Verifying release integrity

Every release ships supply-chain metadata so you can verify what you download:

- Signed checksums: `checksums.txt` is signed with [cosign](https://github.com/sigstore/cosign) (keyless), producing `checksums.txt.sigstore.json`.
- SBOM: an SPDX Software Bill of Materials is attached to each release archive.
- Build provenance: SLSA build provenance is attested via GitHub OIDC.

```shell
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/nao1215/atago/\.github/workflows/release\.yml@refs/tags/.*' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

```shell
gh attestation verify atago_<version>_<os>_<arch>.tar.gz --repo nao1215/atago  # .zip on Windows
```
