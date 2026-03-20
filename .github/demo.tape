Output .github/demo.gif

Set FontSize 12
Set Padding 4
# Note: the font must be installed locally.
Set FontFamily "Berkeley Mono Variable"
Set Width 500
Set Height 500
Set Theme "rose-pine-moon"
Set WindowBar Colorful
Set CursorBlink false
Set TypingSpeed 75ms

Require go

Type "go test -v -run TestReporter ."
Enter
Sleep 3s

# Display the output.
Sleep 5s
