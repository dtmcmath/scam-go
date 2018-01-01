package sexpr

import (
	"fmt"
	"testing"
)

func TestEvaluateEqQ(t *testing.T) {
	atomone := mkAtomNumber("1")
	// atomtwo := mkAtomNumber("2")
	// atomthree := mkAtomNumber("3")

	var tests = []struct{
		input string
		want Sexpr
	} {
		{ "(eq? 1 2)", False },
		{ "(eq? 2 2)", True },
		{ "(+ 1 2)", atomthree },
		{ "(+ 1 (+ 1 1))", atomthree },
		{ "(eq? (+ 1 2) 3)", True },
		{ "(car (cons 1 2))", atomone },
		{ "(eq? (car (cons 1 2)) 1)", True },
		{ "(eq? (cons 1 2) (cons 1 2))", False },
		{ "'()", Nil },
		{ "()", Nil },
		{ "'1", atomone },
		{ "(cons '1 ())", mkList(atomone) },
	}

	for _, test := range tests {
		_, sexprs := Parse("test", mkRuneChannel(test.input))
		if sx, ok := <- sexprs ; !ok {
			t.Errorf("Parsing %q gave no S-expressions", test.input)
		} else {
			got := Evaluate(sx)
			if !equalSexpr(got, test.want) {
				t.Errorf("Evaluate[%s]=%v, want %v",
					test.input, got, test.want,
				)
			}
			_, ok = <- sexprs
			if ok {
				t.Errorf("Parsing %q save multiple S-expressions", test.input)
			}
		}
	}
}

func ExampleEvaluator() {
	s := "(cons 1 2) (eq? (car (cons 1 2)) 1)"

	_, ch := Parse("repl", mkRuneChannel(s))
	for sx := range ch {
		fmt.Println("Evaluating", sx)
		val := Evaluate(sx)
		fmt.Println("gave", val)
	}
	// Output:
	// Evaluating Cons(Sym(cons), Cons(1, Cons(2, Nil)))
	// gave Cons(1, 2)
	// Evaluating Cons(Sym(eq?), Cons(Cons(Sym(car), Cons(Cons(Sym(cons), Cons(1, Cons(2, Nil))), Nil)), Cons(1, Nil)))
	// gave #t
}
