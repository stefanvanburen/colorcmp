Set FontSize 24
Set Padding 4
# Note: the font must be installed locally.
Set FontFamily "Go Mono"
Set Width 1000
Set Height 900
Set CursorBlink false
Set TypingSpeed 75ms

Require go

Type "go test -v -run TestReporter ."
Enter
Sleep 3s

# Display the output.
Sleep 5s
