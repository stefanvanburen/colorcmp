# Run lint and test.
default: check

# Run all tests.
test:
	go test -race ./...

# Run go vet.
lint:
	go vet ./...

# Run lint and test.
check: lint test

# Recreate the demo gifs.
demo: .github/demo-light.gif .github/demo-dark.gif
.github/demo-light.gif: .github/demo-light.tape .github/demo-base.tape $(wildcard *.go)
	vhs .github/demo-light.tape
.github/demo-dark.gif: .github/demo-dark.tape .github/demo-base.tape $(wildcard *.go)
	vhs .github/demo-dark.tape

# Recreate the demo gifs and open them in Safari.
open-demo: demo
	open .github/demo-light.gif -a Safari.app
	open .github/demo-dark.gif -a Safari.app
