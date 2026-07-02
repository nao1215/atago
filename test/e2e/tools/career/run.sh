#!/usr/bin/env bash
#
# Bootstrap for atago's career dogfood. This replaces career's ShellSpec
# harness (spec/ + spec_helper.sh): it builds the LATEST career from its
# checkout, builds atago, puts career first on PATH, exports the bundled
# examples directory, and runs the atago specs in this directory against the
# real binary.
#
# The test DEFINITIONS are atago YAML (test/e2e/tools/career/*.atago.yaml) —
# no ShellSpec. Each scenario works inside its isolated ${workdir}; specs that
# need the bundled sample resume copy it from $CAREER_EXAMPLES (career's
# spec_helper.sh copied examples/minimal.yaml to $WORK/resume.yaml).
#
# Environment contract used by the specs:
#   PATH              career resolves here
#   CAREER_EXAMPLES   absolute path to the bundled examples/ fixtures
#
# Usage: test/e2e/tools/career/run.sh [extra atago args...]
#   CAREER_REPO overrides the checkout (default: ~/ghq/github.com/nao1215/career).
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../.." && pwd)"
CAREER_REPO="${CAREER_REPO:-$HOME/ghq/github.com/nao1215/career}"

if [ ! -f "$CAREER_REPO/main.go" ]; then
	echo "career: checkout not found at '$CAREER_REPO' (set CAREER_REPO)" >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/atago-career.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT
mkdir -p "$TMP/bin"

echo "career: building atago and career..."
( cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/atago" . )
( cd "$CAREER_REPO" && env CGO_ENABLED=0 go build -o "$TMP/bin/career" . )

export PATH="$TMP/bin:$PATH"
export CAREER_EXAMPLES="$CAREER_REPO/examples"

echo "career: $("$TMP/bin/career" version 2>/dev/null | head -1)"
"$TMP/bin/atago" run "$@" "$SCRIPT_DIR"
