#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -eq 0 ]; then
	echo "usage: $0 <package> [<package> ...]" >&2
	exit 64
fi

snapshot_id="${ATAGO_UBUNTU_SNAPSHOT_ID:-20260721T000000Z}"
codename="$(sed -n 's/^VERSION_CODENAME=//p' /etc/os-release)"

# The third-party matrix installs Ubuntu-packaged tools from one immutable
# snapshot so their versions do not drift as the runner image and the apt
# repositories move forward.
#
# The snapshot is the only source apt reads here. The runner image also lists
# the live archive mirror, security.ubuntu.com and packages.microsoft.com, and
# an earlier version of this script fetched all of them on every update even
# though nothing installed from them: security.ubuntu.com refused connections
# for whole runs (sqlite3 failed all three attempts on 2026-11-19) and the
# archive mirror once stalled one `apt-get update` until the job's 20-minute
# timeout cancelled it. A private source list keeps those hosts out of the
# install entirely and leaves the runner's own apt configuration untouched.
sources_dir=/etc/apt/atago-snapshot.d
sudo mkdir -p "$sources_dir"
printf 'Types: deb\nURIs: https://snapshot.ubuntu.com/ubuntu/%s\nSuites: %s %s-updates %s-backports %s-security\nComponents: main restricted universe multiverse\nSigned-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg\n' \
	"$snapshot_id" "$codename" "$codename" "$codename" "$codename" |
	sudo tee "$sources_dir/ubuntu-snapshot.sources" >/dev/null

# Fetch only the indexes an install needs. The command-not-found index exists
# to suggest a package for an unknown shell command, and the translations only
# localize package descriptions; nothing here uses either, and every request we
# skip is one the snapshot host cannot fail. snapshot.ubuntu.com once served a
# 500 for dists/noble-updates/universe/cnf/Commands-amd64 and five suites never
# ran. The index target is named CNF (see 50command-not-found).
#
# Acquire::*::Timeout bounds a connection that hangs instead of failing, and
# Acquire::Retries retries a failed fetch with backoff before apt gives up.
apt_options=(
	-o "Dir::Etc::SourceList=/dev/null"
	-o "Dir::Etc::SourceParts=$sources_dir"
	-o "Acquire::IndexTargets::deb::CNF::DefaultEnabled=false"
	-o "Acquire::Languages=none"
	-o "Acquire::Retries=5"
	-o "Acquire::http::Timeout=30"
	-o "Acquire::https::Timeout=30"
	-o "DPkg::Lock::Timeout=60"
)

# The snapshot host still 5xxes under load, so the whole update + download is
# retried as well. Each network phase also has a deadline: a transfer that
# trickles along never trips the connection timeout, and without a deadline it
# holds the step until the job timeout. Packages are unpacked only after every
# file is on disk, so a deadline never interrupts dpkg.
#
# `apt-get update` alone exits 0 when some indexes fail ("Some index files
# failed to download. They have been ignored"), and an install against that
# partial package list then fails with "has no installation candidate" however
# often it is retried. --error-on=any makes a partial update a failure, and the
# retry repeats the update rather than the install alone.
fetch_once() {
	sudo timeout --kill-after=10 120 apt-get "${apt_options[@]}" update --error-on=any &&
		sudo timeout --kill-after=10 240 apt-get "${apt_options[@]}" install -y --no-install-recommends --download-only "$@"
}

for attempt in 1 2 3; do
	if fetch_once "$@"; then
		sudo apt-get "${apt_options[@]}" install -y --no-install-recommends "$@"
		exit 0
	fi
	if [ "$attempt" -lt 3 ]; then
		echo "atago: apt update + download of '$*' failed or passed its deadline (attempt $attempt/3); retrying in $((attempt * 20))s" >&2
		sleep $((attempt * 20))
	fi
done
echo "atago: apt update + download of '$*' failed three times against snapshot $snapshot_id" >&2
exit 1
