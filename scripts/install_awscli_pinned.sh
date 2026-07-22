#!/usr/bin/env bash
set -euo pipefail

ver="2.36.5"
sha256="8e6c725ed2804bdbdd8d5d730998c19b3b09be06c2c31432673b5af12d2dff79"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

zip="$tmpdir/awscliv2.zip"
curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${ver}.zip" -o "$zip"
echo "${sha256}  ${zip}" | sha256sum -c -
unzip -q "$zip" -d "$tmpdir"

bin_dir="$(go env GOPATH)/bin"
install_dir="$HOME/.local/aws-cli-${ver}"
mkdir -p "$bin_dir"
"$tmpdir/aws/install" --bin-dir "$bin_dir" --install-dir "$install_dir"
