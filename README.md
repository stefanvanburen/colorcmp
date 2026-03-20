# colorcmp

A [`cmp.Reporter`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Reporter) that uses [`znkr.io/diff`](https://pkg.go.dev/znkr.io/diff) to display differences between compared values with ANSI terminal colors.

![demo](.github/demo.gif)

## Installation

```console
$ go get github.com/stefanvanburen/colorcmp
```

## Usage

```go
reporter := colorcmp.New(os.Stdout)
cmp.Equal(x, y, cmp.Reporter(reporter))
fmt.Print(reporter.String())
```

In tests, pass `t.Output()` so color detection follows the test output stream:

```go
reporter := colorcmp.New(t.Output())
if !cmp.Equal(x, y, cmp.Reporter(reporter)) {
    t.Errorf("mismatch:\n%s", reporter.String())
}
```

## Environment variables

| Variable | Effect |
|---|---|
| [`NO_COLOR`](https://no-color.org) | Disables color output |
| [`FORCE_COLOR`](https://force-color.org) | Forces color output |
