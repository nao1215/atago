# Original ShellSpec suite for the migration guide: Skip if guards. Maps 1:1
# to ../../selection.atago.yaml, where the same intent becomes declarative
# `skip:`/`only:` gates and `tags:` selected with `atago run --tag` /
# `--skip-tag`. (ShellSpec's own focusing/tagging lives in fIt/xIt and
# example tags selected with --tag.)

Describe 'migration selection'
  It 'runs only where a POSIX filesystem exists'
    Skip if "no POSIX filesystem" [ ! -d / ]
    When run test -d /
    The status should be success
  End

  It 'runs only when an environment variable is set'
    Skip if "CI_ONLY_UNSET_MARKER not set" [ -z "${CI_ONLY_UNSET_MARKER:-}" ]
    When run echo would only run in CI
    The status should be success
  End

  It 'skip decided by a probe command'
    Skip if "probe matched" false
    When run echo probe did not match, so we run
    The status should be success
    The output should include "we run"
  End
End
