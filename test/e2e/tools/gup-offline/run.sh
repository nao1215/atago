#!/usr/bin/env bash
#
# Bootstrap for atago's OFFLINE gup dogfood. This replaces gup's ShellSpec
# harness (e2e/run.sh + spec_helper.sh): it builds gup and the in-repo test
# module proxy, starts the proxy, exports the shared offline toolchain env, and
# then runs the atago specs in this directory against the real gup CLI.
#
# The test DEFINITIONS are atago YAML (test/e2e/tools/gup-offline/*.atago.yaml) — no
# ShellSpec. This script is only the environment bootstrap (a plain shell
# program, not a test framework), exactly the role gup's own run.sh plays.
#
# Each atago scenario builds its own isolated HOME/GOBIN/GOPATH inside its temp
# workdir via scenario-level `env:` + ${workdir}, and inherits GOPROXY and the
# shared module/build caches from here. Everything is offline and throwaway.
#
# Usage: test/e2e/tools/gup-offline/run.sh [extra atago args...]
#   GUP_REPO overrides the gup checkout (default: ~/ghq/github.com/nao1215/gup).
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../.." && pwd)"
GUP_REPO="${GUP_REPO:-$HOME/ghq/github.com/nao1215/gup}"

if [ ! -d "$GUP_REPO/e2e/testproxy" ]; then
	echo "offline-gup: gup checkout not found at '$GUP_REPO' (set GUP_REPO)" >&2
	echo "offline-gup: this dogfood needs gup's in-repo offline module proxy." >&2
	exit 127
fi

TMP="$(mktemp -d "${TMPDIR:-/tmp}/atago-gup.XXXXXX")"
PROXY_PID=""
cleanup() {
	if [ -n "$PROXY_PID" ]; then
		kill "$PROXY_PID" >/dev/null 2>&1 || true
		wait "$PROXY_PID" 2>/dev/null || true
	fi
	# The Go module cache is written read-only; make it removable.
	chmod -R u+w "$TMP" >/dev/null 2>&1 || true
	rm -rf "$TMP"
}
trap cleanup EXIT

mkdir -p "$TMP/bin" "$TMP/gomodcache" "$TMP/gocache"

echo "offline-gup: building atago, gup, and the test proxy..."
( cd "$REPO_ROOT" && env CGO_ENABLED=0 go build -o "$TMP/bin/atago" . )
( cd "$GUP_REPO" && go build -ldflags '-X github.com/nao1215/gup/internal/cmdinfo.Version=v0.0.0-e2e' -o "$TMP/bin/gup" . )
( cd "$GUP_REPO" && go build -o "$TMP/bin/testproxy" ./e2e/testproxy )

echo "offline-gup: starting offline module proxy..."
"$TMP/bin/testproxy" -dir "$TMP/proxy" -url-file "$TMP/proxy.url" -addr 127.0.0.1:0 &
PROXY_PID=$!
for _ in $(seq 1 50); do
	[ -s "$TMP/proxy.url" ] && break
	sleep 0.1
done
if [ ! -s "$TMP/proxy.url" ]; then
	echo "offline-gup: test proxy did not start" >&2
	exit 1
fi

# Shared, offline toolchain settings. Per-scenario HOME/GOBIN/GOPATH are set by
# each spec via scenario `env:` + ${workdir}; these are inherited by every run.
export GOPROXY GOSUMDB GOFLAGS GOTOOLCHAIN GOMODCACHE GOCACHE
GOPROXY="$(cat "$TMP/proxy.url")"
GOSUMDB="off"
GOFLAGS="-mod=mod"
GOTOOLCHAIN="local"
GOMODCACHE="$TMP/gomodcache"
GOCACHE="$TMP/gocache"
# Put the e2e-built gup first on PATH so the specs exercise that exact binary.
export PATH="$TMP/bin:$PATH"

echo "offline-gup: GOPROXY=$GOPROXY"
# Extra args (e.g. --parallel 8) go before the path so the flag parser sees them.
"$TMP/bin/atago" run "$@" "$SCRIPT_DIR"
