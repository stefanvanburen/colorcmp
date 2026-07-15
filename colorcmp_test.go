package colorcmp_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stefanvanburen/colorcmp"
)

type Address struct {
	Street string
	City   string
	State  string
	Zip    string
}

type Person struct {
	Name    string
	Age     int
	Address Address
	Hobbies []string
}

// Server has an Equal method, so go-cmp treats it as an opaque leaf rather
// than traversing its fields. This triggers the multi-line JSON formatting
// path in colorcmp, producing a structured diff of the whole value.
type Server struct {
	Host    string
	Port    int
	Timeout int
	Debug   bool
}

func (s Server) Equal(other Server) bool { return s == other }

// unexported exercises the reflect.Value.CanInterface == false path in
// formatValue, where a value cannot be JSON-marshaled and %#v is used instead.
type unexported struct{ n int }

// lines joins its arguments with newlines and appends a trailing newline,
// keeping multi-line expectations (with embedded tabs) readable.
func lines(ls ...string) string {
	return strings.Join(ls, "\n") + "\n"
}

// diffOf reports the colorless diff between x and y using a zero-value Reporter.
func diffOf(x, y any, opts ...cmp.Option) string {
	var r colorcmp.Reporter
	cmp.Equal(x, y, append(opts, cmp.Reporter(&r))...)
	return r.String()
}

func TestReporterOutput(t *testing.T) {
	tests := []struct {
		name string
		x, y any
		opts []cmp.Option
		want string
	}{
		{
			name: "scalar",
			x:    1, y: 2,
			want: "{int}: -1 +2\n",
		},
		{
			name: "string quotes values",
			x:    "hello", y: "world",
			want: `{string}: -"hello" +"world"` + "\n",
		},
		{
			name: "nil pointer versus value",
			x:    (*int)(nil), y: new(int),
			want: "{*int}: -null +0\n",
		},
		{
			name: "struct field",
			x:    Address{City: "New York"}, y: Address{City: "Boston"},
			want: `City: -"New York" +"Boston"` + "\n",
		},
		{
			name: "nested field and slice index carry their path",
			x:    Person{Age: 30, Address: Address{City: "New York"}, Hobbies: []string{"reading", "hiking"}},
			y:    Person{Age: 31, Address: Address{City: "Boston"}, Hobbies: []string{"reading", "cycling"}},
			want: "Age: -30 +31\n" +
				`Address.City: -"New York" +"Boston"` + "\n" +
				`Hobbies[1]: -"hiking" +"cycling"` + "\n",
		},
		{
			name: "slice element",
			x:    []int{1, 2, 3}, y: []int{1, 9, 3},
			want: "[1]: -2 +9\n",
		},
		{
			// Regression: a value present only in y once panicked on
			// reflect.Value.Type of the zero (invalid) x value.
			name: "slice grows",
			x:    []int{1, 2}, y: []int{1, 2, 3},
			want: "[?->2]: +3\n",
		},
		{
			name: "slice shrinks",
			x:    []int{1, 2, 3}, y: []int{1, 2},
			want: "[2->?]: -3\n",
		},
		{
			// Regression: an added map key once panicked; it must not leak
			// <invalid> either.
			name: "map key added",
			x:    map[string]int{"a": 1}, y: map[string]int{"a": 1, "b": 2},
			want: `["b"]: +2` + "\n",
		},
		{
			name: "map key removed",
			x:    map[string]int{"a": 1, "b": 2}, y: map[string]int{"a": 1},
			want: `["b"]: -2` + "\n",
		},
		{
			name: "unexported field falls back to %#v",
			x:    unexported{n: 1}, y: unexported{n: 2},
			opts: []cmp.Option{cmp.AllowUnexported(unexported{})},
			want: "n: -1 +2\n",
		},
		{
			name: "multi-line leaf",
			x:    Server{Host: "localhost", Port: 8080, Timeout: 30, Debug: true},
			y:    Server{Host: "remotehost", Port: 9090, Timeout: 60, Debug: false},
			want: lines(
				"{colorcmp_test.Server}:",
				" {",
				"-\t\"Host\": \"localhost\",",
				"-\t\"Port\": 8080,",
				"-\t\"Timeout\": 30,",
				"-\t\"Debug\": true",
				"+\t\"Host\": \"remotehost\",",
				"+\t\"Port\": 9090,",
				"+\t\"Timeout\": 60,",
				"+\t\"Debug\": false",
				" }",
			),
		},
		{
			name: "multi-line value added on one side",
			x:    map[string]Server{},
			y:    map[string]Server{"s": {Host: "h", Port: 1, Timeout: 2, Debug: true}},
			want: lines(
				`["s"]:`,
				"+{",
				"+\t\"Host\": \"h\",",
				"+\t\"Port\": 1,",
				"+\t\"Timeout\": 2,",
				"+\t\"Debug\": true",
				"+}",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffOf(tt.x, tt.y, tt.opts...)
			if got != tt.want {
				t.Errorf("diff mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestReporterColors(t *testing.T) {
	// New consults these to decide whether to emit ANSI codes; force them on.
	t.Setenv("NO_COLOR", "")
	t.Setenv("FORCE_COLOR", "1")

	tests := []struct {
		name string
		x, y any
		want string
	}{
		{"change", 1, 2, "{int}: \033[31m-1\033[m \033[32m+2\033[m\n"},
		{"insertion", []int{1}, []int{1, 2}, "[?->1]: \033[32m+2\033[m\n"},
		{"deletion", []int{1, 2}, []int{1}, "[1->?]: \033[31m-2\033[m\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := colorcmp.New(io.Discard)
			cmp.Equal(tt.x, tt.y, cmp.Reporter(r))
			if got := r.String(); got != tt.want {
				t.Errorf("diff mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	x := "hello"
	y := "hello"

	var r colorcmp.Reporter
	if !cmp.Equal(x, y, cmp.Reporter(&r)) {
		t.Fatal("expected equal")
	}
	if r.String() != "" {
		t.Fatalf("expected empty diff, got: %s", r.String())
	}
}

func ExampleReporter() {
	type Point struct {
		X, Y int
	}
	x := Point{X: 1, Y: 2}
	y := Point{X: 1, Y: 3}

	var r colorcmp.Reporter
	cmp.Equal(x, y, cmp.Reporter(&r))
	fmt.Print(r.String())
	// Output:
	// Y: -2 +3
}
