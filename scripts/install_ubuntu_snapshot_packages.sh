#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -eq 0 ]; then
	echo "usage: $0 <package> [<package> ...]" >&2
	exit 64
fi

snapshot_id="${ATAGO_UBUNTU_SNAPSHOT_ID:-20260721T000000Z}"

# Ubuntu 24.04+ apt can pin every archive request to a specific immutable
# snapshot. The third-party matrix uses this so Ubuntu-packaged tools do not
# drift as the runner image and apt repositories move forward.
printf 'APT::Snapshot "%s";\n' "$snapshot_id" | sudo tee /etc/apt/apt.conf.d/99atago-snapshot >/dev/null

# Fetch only the indexes an install needs. The command-not-found index exists
# to suggest a package for an unknown shell command, and the translations only
# localize package descriptions; nothing here uses either, and every request we
# skip is one the snapshot host cannot fail. snapshot.ubuntu.com once served a
# 500 for dists/noble-updates/universe/cnf/Commands-amd64 and five suites never
# ran. The index target is named CNF (see 50command-not-found); an earlier
# version of this script disabled a target called "Commands", which does not
# exist, so the c-n-f indexes kept being fetched.
printf 'Acquire::IndexTargets::deb::CNF::DefaultEnabled "false";\nAcquire::Languages "none";\n' |
	sudo tee /etc/apt/apt.conf.d/99atago-minimal-indexes >/dev/null

# The mirror still 5xxes under load. apt retries each acquire, and the whole
# update + install is retried too, because a snapshot host that is briefly
# unhealthy is not a reason to report the tool under test as broken.
#
# `apt-get update` alone exits 0 when some indexes fail ("Some index files
# failed to download. They have been ignored"), and an install against that
# partial package list then fails with "has no installation candidate" however
# often it is retried. --error-on=any makes a partial update a failure, and the
# retry repeats the update rather than the install alone.
install_once() {
	sudo apt-get update --error-on=any -o Acquire::Retries=3 &&
		sudo apt-get install -y --no-install-recommends -o Acquire::Retries=3 "$@"
}

for attempt in 1 2 3; do
	if install_once "$@"; then
		exit 0
	fi
	if [ "$attempt" -lt 3 ]; then
		echo "atago: apt update + install of '$*' failed (attempt $attempt/3); retrying in $((attempt * 30))s" >&2
		sleep $((attempt * 30))
	fi
done
echo "atago: apt update + install of '$*' failed three times against snapshot $snapshot_id" >&2
exit 1
