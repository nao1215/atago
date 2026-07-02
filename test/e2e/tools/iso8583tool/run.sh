#!/usr/bin/env bash
#
# Bootstrap for atago's iso8583tool dogfood. This replaces iso8583tool's
# ShellSpec harness (spec/ + spec_helper.sh): it builds the LATEST iso8583tool
# from its checkout, builds the single-shot TCP mock server it ships
# (spec/mock), builds atago, exports the environment the specs expect, and runs
# the atago specs in this directory against the real binary.
#
# The test DEFINITIONS are atago YAML (test/e2e/tools/iso8583tool/*.atago.yaml)
# — no ShellSpec. This script is only the environment bootstrap, exactly the
# role iso8583tool's `make build` + spec_helper.sh play.
#
# Environment contract used by the specs (mirrors spec/spec_helper.sh):
#   PATH            iso8583tool and iso-mock (the spec/mock server) resolve here
#   ISO_EXAMPLES    absolute path to the bundled examples/ fixtures
#   REPLY_HEX       hex of the 0810 network-echo response the mock replies with
#
# Usage: test/e2e/tools/iso8583tool/run.sh [extra atago args...]
#   ISO_REPO overrides the checkout (default: ~/ghq/github.com/nao1215/iso8583tool).
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../.." && pwd)"
ISO_REPO="${ISO_REPO:-$HOME/ghq/github.com/nao1215/iso8583tool}"

if [ ! -f "$ISO_REPO/main.go" ]; then
	echo "iso8583tool: checkout not found at '$ISO_REPO' (set ISO_REPO)" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/atago-iso8583tool.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT
mkdir -p "$TMP/bin"

echo "iso8583tool: building atago, iso8583tool, and the mock server..."
( cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/atago" . )
( cd "$ISO_REPO" && env CGO_ENABLED=0 go build -o "$TMP/bin/iso8583tool" main.go )
( cd "$ISO_REPO" && go build -o "$TMP/bin/iso-mock" ./spec/mock )

export PATH="$TMP/bin:$PATH"
export ISO_EXAMPLES="$ISO_REPO/examples"
REPLY_HEX="$(tr -d ' \t\n\r' < "$ISO_EXAMPLES/basei/0810-network-echo-response.hex")"
export REPLY_HEX

echo "iso8583tool: $("$TMP/bin/iso8583tool" version 2>/dev/null | head -1)"
# Extra args (e.g. --parallel 4) go before the path so the flag parser sees them.
"$TMP/bin/atago" run "$@" "$SCRIPT_DIR"
