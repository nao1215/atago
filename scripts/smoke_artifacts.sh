#!/bin/sh
# Smoke-test goreleaser snapshot artifacts before a release can be published.
# "An archive exists" is not release quality. This proves what a user actually
# does with a release: verify checksums, read the SBOM, extract an archive,
# and run the binary end-to-end (version -> init -> run).
#
# Set SMOKE_SKIP_SBOM=1 for local runs without syft on PATH (CI must not).
set -eu

dist_dir="${1:-dist}"

fail() {
	echo "smoke: $1" >&2
	exit 1
}

[ -d "$dist_dir" ] || fail "dist directory not found: $dist_dir"
[ -f "$dist_dir/checksums.txt" ] || fail "checksums.txt not found in $dist_dir"

archives="$(
	find "$dist_dir" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.zip' \) \
	| sort
)"
[ -n "$archives" ] || fail "no release archives found in $dist_dir"

# Every OS/arch the release page promises must actually be in the dist.
for want in linux_amd64.tar.gz linux_arm64.tar.gz darwin_amd64.tar.gz \
	darwin_arm64.tar.gz windows_amd64.zip windows_arm64.zip; do
	echo "$archives" | grep -q "_${want}\$" || fail "missing release archive for ${want}"
done

# 1. Integrity: every artifact listed in checksums.txt must hash-match, and
#    every archive must be listed (an unlisted archive would ship unverifiable).
(
	cd "$dist_dir"
	if command -v sha256sum >/dev/null; then
		sha256sum --check --strict --quiet checksums.txt
	else
		shasum -a 256 --check --strict --quiet checksums.txt # macOS
	fi
) || fail "checksum verification failed against checksums.txt"
for a in $archives; do
	grep -q " $(basename "$a")\$" "$dist_dir/checksums.txt" \
		|| fail "$(basename "$a") is not listed in checksums.txt"
done
echo "smoke: checksums.txt verifies every archive"

# 2. Supply-chain metadata: each archive ships a parseable SPDX SBOM, as the
#    README's "Verifying release integrity" section promises.
if [ "${SMOKE_SKIP_SBOM:-0}" = "1" ]; then
	echo "smoke: SBOM check skipped (SMOKE_SKIP_SBOM=1)"
else
	for a in $archives; do
		sbom="${a}.sbom.json"
		[ -f "$sbom" ] || fail "missing SBOM: ${sbom}"
		grep -q '"spdxVersion"' "$sbom" || fail "not an SPDX document: ${sbom}"
		if command -v python3 >/dev/null; then
			python3 -m json.tool "$sbom" >/dev/null || fail "SBOM is not valid JSON: ${sbom}"
		fi
	done
	echo "smoke: every archive has a valid SPDX SBOM"
fi

# 3. Layout: every archive carries the binary plus license/readme at its root,
#    so "extract and run" works the way the install docs describe.
for a in $archives; do
	case "$a" in
	*.zip)
		listing="$(unzip -l "$a")"
		echo "$listing" | grep -q ' atago\.exe$' || fail "atago.exe missing from $(basename "$a")"
		;;
	*)
		listing="$(tar -tzf "$a")"
		echo "$listing" | grep -qx 'atago' || fail "atago binary missing from $(basename "$a")"
		;;
	esac
	echo "$listing" | grep -q 'LICENSE' || fail "LICENSE missing from $(basename "$a")"
	echo "$listing" | grep -q 'README' || fail "README missing from $(basename "$a")"
done
echo "smoke: every archive contains the binary, LICENSE, and README"

# 4. The user journey: extract the archive matching this host and drive the
#    real released binary through version -> init -> run.
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
x86_64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
esac

host_archive=""
for a in $archives; do
	case "$a" in
	*"_${os}_${arch}.tar.gz") host_archive="$a" ;;
	esac
done
[ -n "$host_archive" ] || fail "no archive matches this host (${os}/${arch}); cannot run the binary"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT INT TERM
tar -xzf "$host_archive" -C "$workdir"

version_out="$("$workdir/atago" version)"
case "$version_out" in
atago\ *) ;;
*) fail "unexpected 'atago version' output: ${version_out}" ;;
esac

(
	cd "$workdir"
	./atago init >/dev/null || fail "'atago init' failed on the released binary"
	./atago run example.atago.yaml >/dev/null \
		|| fail "'atago run' failed on the spec the released binary scaffolded"
)
echo "smoke: extracted $(basename "$host_archive") and ran ${version_out} (init + run pass)"

echo "release artifacts look healthy"
