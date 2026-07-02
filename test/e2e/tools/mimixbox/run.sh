#!/usr/bin/env bash
#
# Bootstrap for atago's mimixbox dogfood. This replaces mimixbox's ShellSpec
# harness (test/it + spec_helper.sh): it builds the mimixbox multi-call binary,
# `--full-install`s every applet into an isolated bin dir, puts that dir FIRST on
# PATH, and then runs the atago specs in this directory against the real applets.
#
# The test DEFINITIONS are atago YAML (test/e2e/tools/mimixbox/**/*.atago.yaml) — no
# ShellSpec. This script is only the environment bootstrap (a plain shell
# program, not a test framework), exactly the role mimixbox's `make test-e2e`
# PATH/setup plays.
#
# Because the applet bin dir is first on PATH, bare commands like `cat`, `seq`,
# `split`, `sort`, `tr` resolve to mimixbox's own applets — identical to the
# ShellSpec suite, which runs with the same PATH. Each atago scenario runs in
# its own isolated temp workdir, so the per-run MIMIXBOX_IT_ROOT used by the
# ShellSpec helpers maps to atago's ${workdir}.
#
# Usage: test/e2e/tools/mimixbox/run.sh [extra atago args...]
#   MIMIXBOX_REPO overrides the checkout (default: ~/ghq/github.com/nao1215/mimixbox).
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../.." && pwd)"
MIMIXBOX_REPO="${MIMIXBOX_REPO:-$HOME/ghq/github.com/nao1215/mimixbox}"

if [ ! -f "$MIMIXBOX_REPO/cmd/mimixbox/main.go" ]; then
	echo "mimixbox: checkout not found at '$MIMIXBOX_REPO' (set MIMIXBOX_REPO)" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/atago-mimixbox.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

mkdir -p "$TMP/bin" "$TMP/applets"

echo "mimixbox: building atago and the mimixbox binary..."
( cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/atago" . )
( cd "$MIMIXBOX_REPO" && go build -trimpath -o "$TMP/applets/mimixbox" ./cmd/mimixbox )

echo "mimixbox: installing applets via --full-install..."
"$TMP/applets/mimixbox" --full-install "$TMP/applets" >/dev/null

# Put the applet dir FIRST on PATH so bare command names resolve to mimixbox's
# own applets (cat, seq, split, sort, ...), exactly as the ShellSpec suite does.
export PATH="$TMP/applets:$PATH"

echo "mimixbox: applets installed: $(ls "$TMP/applets" | wc -l)"
# Extra args (e.g. --parallel 8) go before the path so the flag parser sees them.
"$TMP/bin/atago" run "$@" "$SCRIPT_DIR"
