#!/usr/bin/env bash
#
# Bootstrap for atago's jose dogfood. This replaces jose's ShellSpec harness
# (spec/ + spec_helper.sh): it builds the LATEST jose from its checkout (jwx v4
# needs GOEXPERIMENT=jsonv2), builds atago, puts jose first on PATH, and runs
# the atago specs in this directory against the real binary.
#
# The test DEFINITIONS are atago YAML (test/e2e/tools/jose/*.atago.yaml) — no
# ShellSpec. The specs are fully hermetic: each scenario generates its own keys
# and payloads inside its isolated ${workdir}, so no external fixtures are
# needed (jose's spec_helper.sh ran the binary inside a per-test mktemp dir,
# which maps to atago's ${workdir}).
#
# Usage: test/e2e/tools/jose/run.sh [extra atago args...]
#   JOSE_REPO overrides the checkout (default: ~/ghq/github.com/nao1215/jose).
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../.." && pwd)"
JOSE_REPO="${JOSE_REPO:-$HOME/ghq/github.com/nao1215/jose}"

if [ ! -f "$JOSE_REPO/main.go" ]; then
	echo "jose: checkout not found at '$JOSE_REPO' (set JOSE_REPO)" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/atago-jose.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT
mkdir -p "$TMP/bin"

echo "jose: building atago and jose (GOEXPERIMENT=jsonv2)..."
( cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/atago" . )
( cd "$JOSE_REPO" && env GOEXPERIMENT=jsonv2 CGO_ENABLED=0 go build -o "$TMP/bin/jose" main.go )

export PATH="$TMP/bin:$PATH"

echo "jose: $("$TMP/bin/jose" version 2>/dev/null | head -1)"
"$TMP/bin/atago" run "$@" "$SCRIPT_DIR"
