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
			sexpr_cons{Nil, Nil},
		},
		{
			[]Sexpr{atomfoo},
			sexpr_cons{atomfoo, Nil},
		},
		{
			[]Sexpr{atomfoo, atomone},
			sexpr_cons{atomfoo, sexpr_cons{atomone, Nil}},
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
