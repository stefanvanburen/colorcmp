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

# agg (https://github.com/asciinema/agg) renders an asciinema recording to a gif.
AGG = agg --font-family "Go Mono" --font-size 24

# Recreate the demo gifs by rendering the recorded session (see .github/demo.sh)
# with each theme.
demo: .github/demo.cast
	$(AGG) --theme github-light .github/demo.cast .github/demo-light.gif
	$(AGG) --theme github-dark .github/demo.cast .github/demo-dark.gif

# Record the terminal session that the demo gifs are rendered from.
.github/demo.cast: .github/demo.sh $(wildcard *.go)
	asciinema rec --headless --overwrite --window-size 58x16 -c "sh .github/demo.sh" $@

# Recreate the demo gifs and open them in Safari.
open-demo: demo
	open .github/demo-light.gif -a Safari.app
	open .github/demo-dark.gif -a Safari.app
