# colorcmp

A [`cmp.Reporter`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Reporter) that uses [`znkr.io/diff`](https://pkg.go.dev/znkr.io/diff) to display differences between compared values with ANSI terminal colors.

![demo](.github/demo.gif)

## Installation

```
go get github.com/stefanvanburen/colorcmp
```

## Usage

```go
var r colorcmp.Reporter
cmp.Equal(x, y, cmp.Reporter(&r))
fmt.Print(r.String())
```

## Note

`znkr.io/diff` is pre-stable (`v1.0.0-beta.x`); its output format may change across minor versions.
