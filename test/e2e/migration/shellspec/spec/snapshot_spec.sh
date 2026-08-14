# Original ShellSpec suite for the migration guide: ShellSpec has no snapshot
# testing, so the golden compare is hand-rolled against a committed file — and
# refreshing it after an intended change is a manual edit. Maps to
# ../../snapshot.atago.yaml, where the same golden is a `snapshot:` matcher
# with normalization and `atago snapshot update`.

Describe 'migration snapshot'
  It 'stdout matches the committed golden file'
    When run echo "hello from atago"
    The output should equal "$(cat "$SHELLSPEC_PROJECT_ROOT/../snapshots/greeting.txt")"
  End
End
