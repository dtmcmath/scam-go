package sexpr

import (
	//"fmt"
	"testing"
)

var (
	atomfoo = mkAtomSymbol("foo")
	atomone = mkAtomNumber("1")
	atomtwo = mkAtomNumber("2")
	atomthree = mkAtomNumber("3")
)

func TestConsify(t *testing.T) {
	tests := []struct {
		input []sexpr_general
		want sexpr_general
	} {
		{
			[]sexpr_general{},
			atomConstantNil,
		},
		{
			[]sexpr_general{atomConstantNil},
			mkCons(atomConstantNil, atomConstantNil),
		},
		{
			[]sexpr_general{atomfoo},
			mkCons(atomfoo, atomConstantNil),
		},
		{
			[]sexpr_general{atomfoo, atomone},
			mkCons(atomfoo, mkCons(atomone, atomConstantNil)),
		},
	}

	for _, test := range tests {
		if got := consify(test.input); !equalSexpr(got, test.want) {
			t.Errorf("consify(%s) = %v, want %v",
				test.input, got, test.want,
			)
		}
	}
}

func TestUnconsify(t *testing.T) {
	input := "(1 2 3 2)"
	want  := []sexpr_general{
		atomone,
		atomtwo,
		atomthree,
		atomtwo,
	}
	_, sexprs := Parse("test", mkRuneChannel(input))
	list := <- sexprs
	if got, err := unconsify(list) ; err != nil {
		t.Error(err)
	} else if !deepEqualSexpr(got, want) {
		t.Errorf("unconsify[%s] = %v, want %v",
			input, got, want,
		)
	}
}
func TestUnconsifyErr(t *testing.T) {
	sinput := mkCons(atomone, atomtwo)
	if _, err := unconsify(sinput) ; err == nil {
		t.Error("unconsify[%s] did not give an error", sinput)
	}
}
