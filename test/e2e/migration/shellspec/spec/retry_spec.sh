# Original ShellSpec suite for the migration guide: polling is a hand-rolled
# while/sleep loop in a helper function. Maps to the `run.retry` block in
# ../../retry.atago.yaml, which re-runs the command until an `until:`
# assertion passes.

Describe 'migration retry'
  wait_until_ready() {
    i=0
    while [ "$i" -lt 5 ]; do
      out=$(echo ready)
      case $out in (*ready*) return 0 ;; esac
      i=$((i + 1))
      sleep 1
    done
    return 1
  }

  It 'polls until the command reports ready'
    When call wait_until_ready
    The status should be success
  End
End
