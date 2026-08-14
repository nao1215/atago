# Original ShellSpec suite for the migration guide: polling is a hand-rolled
# while/sleep loop in a helper function. The probed command is stateful, like
# a service warming up: the first attempt creates a marker and prints
# "waiting", the second prints "ready" — so the loop genuinely iterates. Maps
# to the `run.retry` block in ../../retry.atago.yaml, which re-runs the
# command until an `until:` assertion passes.

Describe 'migration retry'
  marker="$SHELLSPEC_TMPBASE/retry-marker"

  probe() {
    if [ -f "$marker" ]; then echo ready; else touch "$marker"; echo waiting; fi
  }
  wait_until_ready() {
    i=0
    while [ "$i" -lt 5 ]; do
      out=$(probe)
      case $out in (*ready*) return 0 ;; esac
      i=$((i + 1))
      sleep 1
    done
    return 1
  }
  remove_marker() { rm -f "$marker"; }
  AfterEach 'remove_marker'

  It 'polls until the command reports ready'
    When call wait_until_ready
    The status should be success
  End
End
