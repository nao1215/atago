# VHS tape for the snapshot-workflow GIF.
# Regenerate with: make demo  (or: vhs doc/vhs/snapshot.tape from the repo root)

Output doc/img/snapshot.gif

Set Shell "bash"
Set FontSize 18
Set Width 1180
Set Height 760
Set Padding 24
Set Theme "Dracula"
Set TypingSpeed 45ms

# setup (hidden): binary on PATH, enter demo dir, start with no snapshot yet
Hide
Type "export PS1='$ ' PATH=$PWD/dist:$PATH; cd doc/demo; rm -f snapshots/help.txt; clear"
Enter
Show

# 1) first run: the snapshot does not exist yet, so it fails with a clear hint
Type "atago run snapshot.atago.yaml"
Sleep 500ms
Enter
Sleep 3s

# 2) record the snapshot
Type "atago snapshot update snapshot.atago.yaml"
Sleep 500ms
Enter
Sleep 1.5s

# 3) the snapshot file is now committed alongside the spec
Type "cat snapshots/help.txt"
Sleep 400ms
Enter
Sleep 1.5s

# 4) re-run: now it passes
Type "atago run snapshot.atago.yaml"
Sleep 500ms
Enter
Sleep 2.5s
