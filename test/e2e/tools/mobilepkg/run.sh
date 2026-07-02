#!/usr/bin/env bash
#
# Bootstrap for atago's mobilepkg dogfood. mobilepkg has no ShellSpec suite, so
# these atago specs ARE its first end-to-end suite: this script builds the
# LATEST mobilepkg from its checkout, builds atago, puts mobilepkg first on
# PATH, exports the bundled testdata directory, and runs the atago specs in this
# directory against the real binary.
#
# The test DEFINITIONS are atago YAML (test/e2e/tools/mobilepkg/*.atago.yaml).
# This script is only the environment bootstrap (a plain shell program, not a
# test framework), the same role career/jose/iso8583tool's run.sh plays.
#
# mobilepkg is fully hermetic: it inspects a package file in-process (no Android
# SDK / Xcode / device / network). The only committed fixture is the small
# intentionally-vulnerable AndroGoat APK; the specs run against it via
# $MOBILEPKG_TESTDATA so the documented CLI behaviour cannot silently rot.
#
# Environment contract used by the specs:
#   PATH               mobilepkg resolves here (built from the checkout)
#   MOBILEPKG_TESTDATA absolute path to mobilepkg's testdata/ fixtures
#                        (testdata/android/androgoat_rich.apk is committed)
#
# Usage: test/e2e/tools/mobilepkg/run.sh [extra atago args...]
#   MOBILEPKG_REPO overrides the checkout
#   (default: ~/ghq/github.com/nao1215/mobilepkg).
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../.." && pwd)"
MOBILEPKG_REPO="${MOBILEPKG_REPO:-$HOME/ghq/github.com/nao1215/mobilepkg}"

if [ ! -f "$MOBILEPKG_REPO/cmd/mobilepkg/main.go" ]; then
	echo "mobilepkg: checkout not found at '$MOBILEPKG_REPO' (set MOBILEPKG_REPO)" >&2
	exit 127
fi
if [ ! -f "$MOBILEPKG_REPO/testdata/android/androgoat_rich.apk" ]; then
	echo "mobilepkg: committed fixture testdata/android/androgoat_rich.apk missing" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/atago-mobilepkg.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT
mkdir -p "$TMP/bin"

echo "mobilepkg: building atago and mobilepkg..."
( cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/atago" . )
( cd "$MOBILEPKG_REPO" && env CGO_ENABLED=0 go build -o "$TMP/bin/mobilepkg" ./cmd/mobilepkg )

export PATH="$TMP/bin:$PATH"
export MOBILEPKG_TESTDATA="$MOBILEPKG_REPO/testdata"

echo "mobilepkg: $("$TMP/bin/mobilepkg" version 2>/dev/null | head -1)"
# Extra args (e.g. --parallel 8) go before the path so the flag parser sees them.
"$TMP/bin/atago" run "$@" "$SCRIPT_DIR"
