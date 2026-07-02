# VHS tape for the review-commands GIF (explain / manifest / doc).
# Regenerate with: make demo  (or: vhs doc/vhs/review.tape from the repo root)

Output doc/img/review.gif

Set Shell "bash"
Set FontSize 17
Set Width 1180
Set Height 820
Set Padding 24
Set Theme "Dracula"
Set TypingSpeed 40ms

Hide
Type "export PS1='$ ' PATH=$PWD/dist:$PATH; cd doc/demo; clear"
Enter
Show

# explain: understand a spec without running it
Type "atago explain passing.atago.yaml"
Sleep 500ms
Enter
Sleep 3s

# manifest: a stable machine-readable summary, without running the spec
Type "atago manifest passing.atago.yaml | head -n 16"
Sleep 500ms
Enter
Sleep 2s

# doc: generate Markdown documentation
Type "atago doc passing.atago.yaml | head -n 16"
Sleep 500ms
Enter
Sleep 3s
