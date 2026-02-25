// Package colorcmp provides a [cmp.Reporter] that displays differences between compared values
// using ANSI terminal colors.
package colorcmp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/term"
	"znkr.io/diff"
)

// Reporter is a [cmp.Reporter] that displays colored diffs using ANSI terminal colors. Single-line
// values are shown inline; multi-line values (e.g. structs formatted as JSON) use a line-by-line
// block diff via [znkr.io/diff].
//
// The zero value is valid and produces output without ANSI color codes, suitable for non-terminal
// output. Use [New] to create a Reporter that auto-detects terminal support.
//
// Example:
//
//	r := colorcmp.New(os.Stdout)
//	cmp.Equal(x, y, cmp.Reporter(r))
//	fmt.Print(r.String())
type Reporter struct {
	path   cmp.Path
	diffs  []string
	colors bool
}

// New returns a Reporter that uses ANSI colors if w is connected to a terminal.
// It respects the NO_COLOR and FORCE_COLOR environment variables.
func New(w io.Writer) *Reporter {
	return &Reporter{colors: isTTY(w)}
}

func isTTY(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" { // https://no-color.org
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" { // https://force-color.org
		return true
	}
	type fdGetter interface {
		Fd() uintptr
	}
	if fd, ok := w.(fdGetter); ok {
		return term.IsTerminal(int(fd.Fd()))
	}
	// Fall back to the TERM environment variable when the writer doesn't expose
	// a file descriptor (e.g. testing.T.Output). go test pipes all fds of the
	// test binary, so fd-based detection always returns false inside tests.
	// TERM is inherited from the shell and reflects the actual terminal, while
	// CI environments typically leave it unset.
	t := os.Getenv("TERM")
	return t != "" && t != "dumb"
}

func (r *Reporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *Reporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		x := formatValue(vx)
		y := formatValue(vy)
		xs := strings.TrimSuffix(x, "\n")
		ys := strings.TrimSuffix(y, "\n")

		path := r.path.String()
		if path == "" {
			path = fmt.Sprintf("{%v}", vx.Type())
		}

		var entry string
		if !strings.Contains(xs, "\n") && !strings.Contains(ys, "\n") {
			if r.colors {
				entry = fmt.Sprintf("%s: \033[31m-%s\033[m \033[32m+%s\033[m\n", path, xs, ys)
			} else {
				entry = fmt.Sprintf("%s: -%s +%s\n", path, xs, ys)
			}
		} else {
			entry = fmt.Sprintf("%s:\n%s", path, renderDiff(x, y, r.colors))
		}
		r.diffs = append(r.diffs, entry)
	}
}

func (r *Reporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

// String returns the accumulated diff output.
func (r *Reporter) String() string {
	return strings.Join(r.diffs, "")
}

// renderDiff returns a line-by-line diff of x and y, using ANSI colors if enabled.
// It uses [diff.Hunks] to show only changed lines with surrounding context.
func renderDiff(x, y string, colors bool) string {
	xlines := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	ylines := strings.Split(strings.TrimSuffix(y, "\n"), "\n")
	var sb strings.Builder
	for i, hunk := range diff.Hunks(xlines, ylines) {
		if i > 0 {
			sb.WriteString("...\n")
		}
		for _, edit := range hunk.Edits {
			switch edit.Op {
			case diff.Delete:
				if colors {
					fmt.Fprintf(&sb, "\033[31m-%s\033[m\n", edit.X)
				} else {
					fmt.Fprintf(&sb, "-%s\n", edit.X)
				}
			case diff.Insert:
				if colors {
					fmt.Fprintf(&sb, "\033[32m+%s\033[m\n", edit.Y)
				} else {
					fmt.Fprintf(&sb, "+%s\n", edit.Y)
				}
			case diff.Match:
				fmt.Fprintf(&sb, " %s\n", edit.X)
			}
		}
	}
	return sb.String()
}

// formatValue formats a reflect.Value as a string for diffing. It uses [json.MarshalIndent] for
// complex types (structs, slices, maps) to produce multi-line output that diffs well
// line-by-line. It falls back to %#v for types that cannot be JSON-marshaled (e.g. channels,
// functions).
func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>\n"
	}
	if v.CanInterface() {
		b, err := json.MarshalIndent(v.Interface(), "", "\t")
		if err == nil {
			return string(b) + "\n"
		}
	}
	return fmt.Sprintf("%#v\n", v)
}
