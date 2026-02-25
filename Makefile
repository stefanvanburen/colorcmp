# Recreate the demo gif.
# To regenerate and view:
# make; and open .github/demo.gif -a Safari.app
.github/demo.gif: .github/demo.tape $(wildcard *.go)
	vhs .github/demo.tape
