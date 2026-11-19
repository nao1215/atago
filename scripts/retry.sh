#!/usr/bin/env bash
# Run a command up to three times, waiting 10 and then 20 seconds between
# attempts, and exit with the last attempt's status.
#
# CI wraps every `go install` in this. A module download from proxy.golang.org
# fails now and then for reasons that have nothing to do with the module
# ("stream error: stream ID 41; INTERNAL_ERROR; received from peer" failed the
# kustomize leg of the third-party matrix on 2026-11-19), and the same command
# succeeds a few seconds later. A command that fails for a real reason, such as
# a version that does not exist, fails all three attempts and still fails the
# step, only 30 seconds later.
set -uo pipefail

if [ "$#" -eq 0 ]; then
	echo "usage: $0 <command> [<argument> ...]" >&2
	exit 64
fi

for attempt in 1 2 3; do
	"$@"
	status=$?
	if [ "$status" -eq 0 ]; then
		exit 0
	fi
	if [ "$attempt" -lt 3 ]; then
		echo "atago: '$*' exited $status (attempt $attempt/3); retrying in $((attempt * 10))s" >&2
		sleep $((attempt * 10))
	fi
done
echo "atago: '$*' failed three times" >&2
exit "$status"
