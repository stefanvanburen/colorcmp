// Package colorcmp provides a [cmp.Reporter] that displays differences between compared values
// using ANSI terminal colors.
package colorcmp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"znkr.io/diff"
)

// Reporter is a [cmp.Reporter] that displays colored diffs using ANSI terminal colors. Single-line
// values are shown inline; multi-line values (e.g. structs formatted as JSON) use a line-by-line
// block diff via [znkr.io/diff].
//
// Example:
//
//	var r colorcmp.Reporter
//	cmp.Equal(x, y, cmp.Reporter(&r))
//	fmt.Print(r.String())
type Reporter struct {
	path  cmp.Path
	diffs []string
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
		var entry string
		if !strings.Contains(xs, "\n") && !strings.Contains(ys, "\n") {
			entry = fmt.Sprintf("%v: \033[31m-%s\033[m \033[32m+%s\033[m\n", r.path, xs, ys)
		} else {
			entry = fmt.Sprintf("%v:\n%s", r.path, colorDiff(x, y))
		}
		r.diffs = append(r.diffs, entry)
	}
}

func (r *Reporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

// String returns the accumulated colored diff output.
func (r *Reporter) String() string {
	return strings.Join(r.diffs, "")
}

// colorDiff returns a colored line-by-line diff of x and y.
func colorDiff(x, y string) string {
	xlines := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	ylines := strings.Split(strings.TrimSuffix(y, "\n"), "\n")
	var sb strings.Builder
	for _, edit := range diff.Edits(xlines, ylines) {
		switch edit.Op {
		case diff.Delete:
			fmt.Fprintf(&sb, "\033[31m-%s\033[m\n", edit.X)
		case diff.Insert:
			fmt.Fprintf(&sb, "\033[32m+%s\033[m\n", edit.Y)
		case diff.Match:
			fmt.Fprintf(&sb, " %s\n", edit.X)
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
