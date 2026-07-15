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
	"unicode/utf8"

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
	// bytesSeen records byte-slice paths already reported. A []byte is compared
	// element by element, so many bytes may differ within one slice; it is
	// rendered as text once, keyed by its path here.
	bytesSeen map[string]bool
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
	if rs.Equal() {
		return
	}
	// []byte is compared byte by byte, so a differing byte would otherwise show
	// up as its decimal value; render the whole slice as text instead.
	if r.reportByteSlice() {
		return
	}

	vx, vy := r.path.Last().Values()
	path := pathString(r.path, vx, vy)

	var entry string
	switch {
	case !vx.IsValid():
		// Value present only in y: render as a pure insertion.
		entry = renderOneSided(path, formatValue(vy), true, r.colors)
	case !vy.IsValid():
		// Value present only in x: render as a pure deletion.
		entry = renderOneSided(path, formatValue(vx), false, r.colors)
	default:
		entry = renderChange(path, formatValue(vx), formatValue(vy), r.colors)
	}
	r.diffs = append(r.diffs, entry)
}

// reportByteSlice handles the case where the differing node is a byte within a
// []byte. Because cmp compares such slices element by element, it reports each
// differing byte separately; this renders the whole slice as text once (keyed
// by path in bytesSeen) when both sides are valid UTF-8, and reports nothing
// handled otherwise so the caller falls back to the default byte rendering.
func (r *Reporter) reportByteSlice() bool {
	if _, ok := r.path.Last().(cmp.SliceIndex); !ok || len(r.path) < 2 {
		return false
	}
	parent := r.path[len(r.path)-2]
	if t := parent.Type(); t.Kind() != reflect.Slice || t.Elem().Kind() != reflect.Uint8 {
		return false
	}
	px, py := parent.Values()
	if !px.IsValid() || !py.IsValid() {
		return false
	}
	bx, by := px.Bytes(), py.Bytes()
	if !utf8.Valid(bx) || !utf8.Valid(by) {
		return false
	}

	// The parent path drops the trailing byte index.
	path := pathString(r.path[:len(r.path)-1], px, py)
	if r.bytesSeen[path] {
		return true
	}
	if r.bytesSeen == nil {
		r.bytesSeen = make(map[string]bool)
	}
	r.bytesSeen[path] = true

	x := formatValue(reflect.ValueOf(string(bx)))
	y := formatValue(reflect.ValueOf(string(by)))
	r.diffs = append(r.diffs, renderChange(path, x, y, r.colors))
	return true
}

// renderChange renders a two-sided change between the formatted values x and y,
// inline for single-line values and as a line-by-line block otherwise.
func renderChange(path, x, y string, colors bool) string {
	xs := strings.TrimSuffix(x, "\n")
	ys := strings.TrimSuffix(y, "\n")
	if !strings.Contains(xs, "\n") && !strings.Contains(ys, "\n") {
		return fmt.Sprintf("%s: %s %s\n", path, colorize("-"+xs, colorDelete, colors), colorize("+"+ys, colorInsert, colors))
	}
	return fmt.Sprintf("%s:\n%s", path, renderDiff(x, y, colors))
}

// pathString returns a human-readable path to the last node in steps. Unlike
// [cmp.Path.String], which keeps only struct field names, it includes slice
// indices and map keys so the location of a diff is unambiguous. It falls back
// to the node's type when the node is the root of the comparison.
func pathString(steps cmp.Path, vx, vy reflect.Value) string {
	var sb strings.Builder
	if len(steps) > 0 {
		for _, step := range steps[1:] { // skip the root step, which carries only the type
			sb.WriteString(step.String())
		}
	}
	if path := strings.TrimPrefix(sb.String(), "."); path != "" {
		return path
	}
	switch {
	case vx.IsValid():
		return fmt.Sprintf("{%s}", typeName(vx.Type()))
	case vy.IsValid():
		return fmt.Sprintf("{%s}", typeName(vy.Type()))
	default:
		return "{}"
	}
}

// typeName renders t for display, preferring the byte alias over uint8 for a
// plain []byte, which reflect otherwise reports as []uint8.
func typeName(t reflect.Type) string {
	if t.Kind() == reflect.Slice && t.Name() == "" && t.Elem() == reflect.TypeFor[byte]() {
		return "[]byte"
	}
	return t.String()
}

func (r *Reporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

// String returns the accumulated diff output.
func (r *Reporter) String() string {
	return strings.Join(r.diffs, "")
}

// ANSI SGR color codes for deletions (red) and insertions (green).
const (
	colorDelete = 31
	colorInsert = 32
)

// colorize wraps s in the given ANSI color code when colors is true, and
// returns s unchanged otherwise.
func colorize(s string, code int, colors bool) string {
	if !colors {
		return s
	}
	return fmt.Sprintf("\033[%dm%s\033[m", code, s)
}

// renderOneSided renders value as an all-insertion (insert true) or
// all-deletion (insert false) block, one line at a time. It is used when a
// value is present on only one side of the comparison.
func renderOneSided(path, value string, insert, colors bool) string {
	sign, code := "-", colorDelete
	if insert {
		sign, code = "+", colorInsert
	}
	lines := strings.Split(strings.TrimSuffix(value, "\n"), "\n")
	if len(lines) == 1 {
		return fmt.Sprintf("%s: %s\n", path, colorize(sign+lines[0], code, colors))
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s:\n", path)
	for _, line := range lines {
		fmt.Fprintf(&sb, "%s\n", colorize(sign+line, code, colors))
	}
	return sb.String()
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
				fmt.Fprintf(&sb, "%s\n", colorize("-"+edit.X, colorDelete, colors))
			case diff.Insert:
				fmt.Fprintf(&sb, "%s\n", colorize("+"+edit.Y, colorInsert, colors))
			case diff.Match:
				fmt.Fprintf(&sb, " %s\n", edit.X)
			}
		}
	}
	return sb.String()
}

// formatValue formats a reflect.Value as a string for diffing. It encodes values as JSON so that
// complex types (structs, slices, maps) produce multi-line output that diffs well line-by-line.
// HTML escaping is disabled so that characters like <, >, and & appear literally rather than as
// \u00xx escapes. It falls back to %#v for types that cannot be JSON-encoded (e.g. channels,
// functions).
func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>\n"
	}
	// Multi-line strings diff far better line-by-line than as a single quoted,
	// escaped blob, so render them raw and let the block diff handle them. A
	// lone trailing newline doesn't make a string multi-line.
	if v.Kind() == reflect.String {
		if s := v.String(); strings.Contains(strings.TrimSuffix(s, "\n"), "\n") {
			return s
		}
	}
	// Render a []byte as its text rather than a JSON base64 blob when it is
	// valid UTF-8, reusing the string handling above.
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
		if b := v.Bytes(); utf8.Valid(b) {
			return formatValue(reflect.ValueOf(string(b)))
		}
	}
	if v.CanInterface() {
		var b strings.Builder
		enc := json.NewEncoder(&b)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "\t")
		if err := enc.Encode(v.Interface()); err == nil {
			return b.String() // Encode already appends a trailing newline.
		}
	}
	return fmt.Sprintf("%#v\n", v)
}
