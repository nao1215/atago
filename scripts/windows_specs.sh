#!/usr/bin/env bash
# Print the run targets in one Windows coverage bucket, space/newline separated
# for word-splitting by the caller.
#
# Usage: bash scripts/windows_specs.sh <cmd|bash|none>
#
# scripts/windows_specs.tsv is the single source of truth, shared by the
# push-gated Windows jobs (.github/workflows/e2e.yml) and the scheduled
# cross-platform drift check (.github/workflows/e2e-cross.yml), so the two
# cannot drift. See the header of that file for what each bucket means.
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <cmd|bash|none>" >&2
  exit 2
fi

case "$1" in
  cmd | bash | none) ;;
  *)
    echo "$0: unknown bucket $1 (want cmd, bash, or none)" >&2
    exit 2
    ;;
esac

dir="$(cd "$(dirname "$0")" && pwd)"
awk -F'\t' -v want="$1" '$0 !~ /^[[:space:]]*#/ && NF > 0 && $2 == want { print $1 }' "$dir/windows_specs.tsv"
