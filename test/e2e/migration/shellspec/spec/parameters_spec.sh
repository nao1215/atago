# Original ShellSpec suite for the migration guide: a Parameters block feeding
# positional $1/$2 into one example. Maps to the single `matrix:` scenario in
# ../../matrix.atago.yaml, which expands one template into one scenario per
# row with each key available as ${name}.

Describe 'migration parameters'
  Parameters
    Alice en
    Bob   fr
  End

  It "greets $1 in $2"
    When run echo "hello $1 ($2)"
    The output should include "$1"
    The output should include "$2"
  End
End
