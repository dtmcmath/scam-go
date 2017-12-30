package sexpr

import (
	//"fmt"
	"testing"
)

func TestConsify(t *testing.T) {
	atomfoo := mkAtomSymbol("foo")
	atomone := mkAtomNumber("1")
	tests := []struct {
		input []Sexpr
		want Sexpr
	} {
		{
			[]Sexpr{},
			Nil,
		},
		{
			[]Sexpr{Nil},
			Cons{Nil, Nil},
		},
		{
			[]Sexpr{atomfoo},
			Cons{atomfoo, Nil},
		},
		{
			[]Sexpr{atomfoo, atomone},
			Cons{atomfoo, Cons{atomone, Nil}},
		},
	}

	for _, test := range tests {
		if got := consify(test.input); got != test.want {
			t.Errorf("consify(%s) = %v, want %v",
				test.input, got, test.want,
			)
		}
	}
}
