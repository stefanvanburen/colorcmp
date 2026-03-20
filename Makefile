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
# To regenerate and view:
# make demo; and open .github/demo.gif -a Safari.app
demo: .github/demo.gif
.github/demo.gif: .github/demo.tape $(wildcard *.go)
	vhs .github/demo.tape
