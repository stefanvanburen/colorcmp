package colorcmp_test

import (
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

func TestReporter(t *testing.T) {
	x := Person{
		Name: "Alice",
		Age:  30,
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
			State:  "NY",
			Zip:    "10001",
		},
		Hobbies: []string{"reading", "hiking", "cooking"},
	}
	y := Person{
		Name: "Alice",
		Age:  31,
		Address: Address{
			Street: "123 Main St",
			City:   "Boston",
			State:  "MA",
			Zip:    "02101",
		},
		Hobbies: []string{"reading", "cycling", "cooking"},
	}

	var r colorcmp.Reporter
	if cmp.Equal(x, y, cmp.Reporter(&r)) {
		t.Fatal("expected not equal")
	}
	out := r.String()
	if out == "" {
		t.Fatal("expected non-empty diff output")
	}
	t.Log("\n" + out)
}

func TestReporterEqual(t *testing.T) {
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
