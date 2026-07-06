#!/usr/bin/env bash
# Single source of truth for the Windows "portable subset" of the self-hosted
# E2E specs: specs whose commands are shell builtins of both /bin/sh and
# cmd.exe, or that only drive atago itself. Both the push-gated e2e-windows job
# (.github/workflows/e2e.yml) and the scheduled cross-platform drift check
# (.github/workflows/e2e-cross.yml) run `atago run $(bash scripts/windows_portable_specs.sh)`,
# so the list lives in exactly one place and the two workflows cannot drift.
#
# Print the targets space-separated for word-splitting by the caller. Every path
# below must resolve to a real file or directory; TestWindowsPortableSubset_Exists
# enforces that so a spec rename fails a unit test instead of only the Windows CI.
set -euo pipefail

printf '%s\n' \
  ./test/e2e/atago/version.atago.yaml \
  ./test/e2e/atago/completion.atago.yaml \
  ./test/e2e/atago/list.atago.yaml \
  ./test/e2e/atago/manifest.atago.yaml \
  ./test/e2e/atago/doc.atago.yaml \
  ./test/e2e/atago/explain.atago.yaml \
  ./test/e2e/atago/init.atago.yaml \
  ./test/e2e/atago/init_templates.atago.yaml \
  ./test/e2e/atago/edge.atago.yaml \
  ./test/e2e/atago/argv_quotes.atago.yaml \
  ./test/e2e/atago/paths_portable.atago.yaml \
  ./test/e2e/atago/exit_codes.atago.yaml \
  ./test/e2e/atago/file_equals.atago.yaml \
  ./test/e2e/atago/store_whole.atago.yaml \
  ./test/e2e/atago/json_list.atago.yaml \
  ./test/e2e/atago/matrix.atago.yaml \
  ./test/e2e/atago/parallel.atago.yaml \
  ./test/e2e/atago/run.atago.yaml \
  ./test/e2e/atago/select.atago.yaml \
  ./test/e2e/atago/skip_command.atago.yaml \
  ./test/e2e/atago/rerun.atago.yaml \
  ./test/e2e/atago/db.atago.yaml \
  ./test/e2e/atago/http.atago.yaml \
  ./test/e2e/atago/mock_server.atago.yaml \
  ./test/e2e/atago/dir_tree.atago.yaml \
  ./test/e2e/atago/changes.atago.yaml \
  ./test/e2e/atago/record.atago.yaml \
  ./test/e2e/atago/duration.atago.yaml \
  ./test/e2e/atago/sandbox_home.atago.yaml \
  ./test/e2e/atago/grpc.atago.yaml \
  ./test/e2e/atago/ssh.atago.yaml \
  ./test/e2e/atago/cdp.atago.yaml \
  ./test/e2e/atago/pdf.atago.yaml \
  ./test/e2e/thirdparty/git
