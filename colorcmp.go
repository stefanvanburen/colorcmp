// Package colorcmp provides a cmp.Reporter that uses ANSI color diffing to
// display differences between compared values.
package colorcmp

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"znkr.io/diff/textdiff"
)

// Reporter is a [cmp.Reporter] that prints colored diffs using ANSI terminal
// colors. It is similar to the DiffReporter example in the go-cmp package but
// uses znkr.io/diff's textdiff for colorized unified diff output.
//
// Usage:
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
		x := fmt.Sprintf("%#v\n", vx)
		y := fmt.Sprintf("%#v\n", vy)
		diff := textdiff.Unified(x, y, textdiff.TerminalColors())
		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n%s", r.path, diff))
	}
}

func (r *Reporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

// String returns the accumulated colored diff output.
func (r *Reporter) String() string {
	return strings.Join(r.diffs, "\n")
}
