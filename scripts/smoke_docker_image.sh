#!/bin/sh
# Smoke-test the GoReleaser-built container image before a tagged release can
# publish it. Snapshot builds used by release-smoke do not always load a local
# `latest` tag, so prove at least one local tag runs; when `latest` is present
# we still prefer it and inspect its multi-arch manifest too.
set -eu

repo="${1:-ghcr.io/nao1215/atago}"

fail() {
	echo "docker-smoke: $1" >&2
	exit 1
}

command -v docker >/dev/null || fail "docker is not installed"

images="$(docker image ls --format '{{.Repository}}:{{.Tag}}' "$repo")"
[ -n "$images" ] || fail "no local docker images found for $repo"

latest_image="$(printf '%s\n' "$images" | grep "^${repo}:latest\$" || true)"
host_image="$latest_image"
[ -n "$host_image" ] || host_image="$(printf '%s\n' "$images" | grep "^${repo}:" | head -n1)"
[ -n "$host_image" ] || fail "no runnable image tag found for $repo"

version_out="$(docker run --rm "$host_image" version)"
case "$version_out" in
atago\ *) ;;
*) fail "unexpected 'atago version' output from ${host_image}: ${version_out}" ;;
esac

if docker buildx imagetools inspect "${repo}:latest" >/tmp/atago-imagetools.txt 2>/dev/null; then
	grep -q 'linux/amd64' /tmp/atago-imagetools.txt || fail "latest manifest is missing linux/amd64"
	grep -q 'linux/arm64' /tmp/atago-imagetools.txt || fail "latest manifest is missing linux/arm64"
fi

echo "docker-smoke: ${host_image} runs ${version_out}"
