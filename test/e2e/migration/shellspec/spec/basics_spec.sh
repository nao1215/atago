# Original ShellSpec suite for the migration guide. Each It here maps 1:1 to
# a scenario in ../../basics.atago.yaml; CI runs both (this file under a
# pinned ShellSpec, the atago spec under the freshly built binary) so the
# mapping the website shows is executed, not asserted from memory.

Describe 'migration basics'
  It 'exit code success'
    When run echo hello
    The status should be success
    The output should include "hello"
  End

  It 'exit code failure'
    When run false
    The status should be failure
  End

  It 'an exact exit code'
    When run sh -c 'exit 3'
    The status should equal 3
  End

  It 'stdout contains'
    When run echo hello world
    The output should include "world"
  End

  It 'stdout equals'
    When run echo exact
    The output should equal "exact"
  End

  It 'stdout does not contain'
    When run echo all good
    The output should not include "panic"
  End

  It 'stdout matches a pattern'
    When run echo v1.2.3
    The output should match pattern 'v[0-9]*.[0-9]*.[0-9]*'
  End

  It 'stderr contains'
    When run sh -c 'echo "warn: heads up" >&2'
    The error should include "warn"
  End

  It 'stderr is blank'
    When run echo quiet
    The output should include "quiet"
    The error should be blank
  End
End
