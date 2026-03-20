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

# Recreate the demo gif.
demo: .github/demo.gif
.github/demo.gif: .github/demo.tape $(wildcard *.go)
	vhs .github/demo.tape

# Recreate the demo gif and open it in Safari.
open-demo: demo
	open .github/demo.gif -a Safari.app
