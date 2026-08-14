# Original ShellSpec suite for the migration guide: BeforeEach/AfterEach
# helpers writing scratch files, and JSON checked by piping through jq. Maps
# 1:1 to ../../fixtures_and_files.atago.yaml, where the same tests become a
# declarative `fixture:` step (no cleanup: every atago scenario gets a fresh
# temp workdir) and JSONPath matchers (no jq dependency).

Describe 'migration fixtures'
  work="$SHELLSPEC_TMPBASE/migration-fixtures"

  seed_input() {
    mkdir -p "$work" && printf 'id,name\n1,Alice\n' > "$work/input.txt"
  }
  remove_scratch() {
    rm -rf "$work"
  }

  BeforeEach 'seed_input'
  AfterEach 'remove_scratch'

  It 'a seeded input produces the expected output file'
    When run cp "$work/input.txt" "$work/output.txt"
    The status should be success
    The file "$work/output.txt" should be exist
    The contents of file "$work/output.txt" should include "Alice"
  End

  It 'json output asserted via jq'
    When run sh -c "printf '%s' '{\"items\":[{\"id\":7,\"name\":\"Alice\"}],\"count\":1}' | jq -r '.items[0].name'"
    The output should equal "Alice"
  End
End
