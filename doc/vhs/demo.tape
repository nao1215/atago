# VHS tape for the atago README demo.
# Regenerate with:  make demo     (builds the binary, then runs vhs)
# or directly:      vhs doc/vhs/demo.tape   (run from the repo root, after `go build -o ./dist/atago .`)

Output doc/img/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1180
Set Height 760
Set Padding 24
Set Theme "Dracula"
Set TypingSpeed 45ms

# --- setup (hidden): put the freshly built binary on PATH
Hide
Type "export PS1='$ ' PATH=$PWD/dist:$PATH; clear"
Enter
Show

# --- a whole suite streams by: one dot per scenario
Type "atago run --parallel 8 ./test/e2e/atago"
Sleep 600ms
Enter
Sleep 4.5s

# --- move to the demo dir (hidden) for the failure walkthrough
Hide
Type "cd doc/demo; clear"
Enter
Show

# --- show a tiny spec that is about to fail
Type "cat failing.atago.yaml"
Sleep 400ms
Enter
Sleep 2s

# --- run it: atago prints Expected / Actual / Hint, compact enough to act on
Type "atago run failing.atago.yaml"
Sleep 600ms
Enter
Sleep 3.5s
