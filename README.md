# colorcmp

A [`cmp.Reporter`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Reporter) that uses [`znkr.io/diff`](https://pkg.go.dev/znkr.io/diff) to display differences between compared values with ANSI terminal colors.

![demo](.github/demo.gif)

## Installation

```
go get github.com/stefanvanburen/colorcmp
```

## Usage

```go
r := colorcmp.New(os.Stdout)
cmp.Equal(x, y, cmp.Reporter(r))
fmt.Print(r.String())
```

In tests, pass `t.Output()` so color detection follows the test output stream:

```go
r := colorcmp.New(t.Output())
if !cmp.Equal(x, y, cmp.Reporter(r)) {
    t.Errorf("mismatch:\n%s", r.String())
}
```

The zero value `var r colorcmp.Reporter` is also valid and produces output without ANSI color codes.

## Environment variables

| Variable | Effect |
|---|---|
| [`NO_COLOR`](https://no-color.org) | Disables color output |
| [`FORCE_COLOR`](https://force-color.org) | Forces color output |

## Note

`znkr.io/diff` is pre-stable (`v1.0.0-beta.x`); its output format may change across minor versions.
