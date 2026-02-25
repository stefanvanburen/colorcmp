.github/demo.gif: .github/demo.tape $(wildcard *.go)
	vhs .github/demo.tape
