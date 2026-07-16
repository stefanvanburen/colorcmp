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

# Recreate the demo gifs.
demo: .github/demo-light.gif .github/demo-dark.gif

# Render each gif from the recorded session (see .github/demo.sh) with a theme.
.github/demo-%.gif: .github/demo.cast
	$(AGG) --theme github-$* $< $@

# Record the terminal session that the demo gifs are rendered from.
.github/demo.cast: .github/demo.sh $(wildcard *.go)
	asciinema rec --headless --overwrite --window-size 58x16 -c "sh .github/demo.sh" $@

# Recreate the demo gifs and open them in Safari.
open-demo: demo
	open .github/demo-light.gif -a Safari.app
	open .github/demo-dark.gif -a Safari.app

# Remove a target if its recipe fails, so a partial gif or cast isn't left
# behind looking up to date.
.DELETE_ON_ERROR:

.PHONY: default test lint check demo open-demo
