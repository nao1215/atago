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

# Do not fetch the command-not-found index. It exists to suggest a package for an
# unknown shell command, which nothing here does, and it is the index that made
# the pinned mirror fail the whole matrix: snapshot.ubuntu.com served a 500 for
# dists/noble-updates/universe/cnf/Commands-amd64, apt exited 100 for one index
# it could not fetch, and five suites never ran. Skipping it removes a request we
# have no use for rather than ignoring a failure that matters.
printf 'Acquire::IndexTargets::deb::Commands::DefaultEnabled "false";\n' |
	sudo tee /etc/apt/apt.conf.d/99atago-no-command-not-found >/dev/null

# The mirror still 5xxes under load. apt retries each acquire, and the whole
# step is retried too, because a snapshot host that is briefly unhealthy is not a
# reason to report the tool under test as broken.
apt_retry() {
	local attempt
	for attempt in 1 2 3; do
		if sudo "$@" -o Acquire::Retries=3; then
			return 0
		fi
		if [ "$attempt" -lt 3 ]; then
			echo "atago: '$*' failed (attempt $attempt/3); retrying in $((attempt * 15))s" >&2
			sleep $((attempt * 15))
		fi
	done
	echo "atago: '$*' failed three times against snapshot $snapshot_id" >&2
	return 1
}

apt_retry apt-get update
apt_retry apt-get install -y --no-install-recommends "$@"
