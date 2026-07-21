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
sudo apt-get update
sudo apt-get install -y --no-install-recommends "$@"
