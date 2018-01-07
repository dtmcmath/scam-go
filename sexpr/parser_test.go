package sexpr

import (
	"fmt"
	"testing"
)

func TestSimpleParse(t *testing.T) {
	atomone := mkAtomNumber("1")
	atomtwo := mkAtomNumber("2")

	primcons := mkAtomSymbol("cons")
	var tests = []struct {
		input string
		want []Sexpr
	}{
		{ "()", []Sexpr{ Nil } },
		{ "(1)", []Sexpr{ mkCons(atomone, Nil) } },
		{ "1", []Sexpr{ atomone } },
		{
			" (1) 2 ",
			[]Sexpr{
				mkCons(atomone, Nil),
				atomtwo,
			},
		},
		{
			"#t #f",
			[]Sexpr{True, False},
		},
		{
			"(cons 1 2)",
			[]Sexpr{
				mkCons(
					primcons,
					mkCons(atomone,
						mkCons(atomtwo,Nil),
					),
				),
			},
		},
		{
			"'()",
			[]Sexpr{
				mkList(Quote, Nil),
			},
		},
		{
			"'1",
			[]Sexpr{ mkList(Quote, atomone) },
		},
		{
			"(cons '1 ())",
			[]Sexpr{
				mkList(primcons, mkList(Quote, atomone), Nil),
			},
		},
		{
			"'(#t)",
			[]Sexpr{mkList(Quote, mkList(True))},
		},
		{
			"o o+",
			[]Sexpr{ mkAtomSymbol("o"), mkAtomSymbol("o+") },
		},
		// {
		// 	"(eq? (car (cons 1 2)) 1)",
		// 	// Bleh.  I've lost track!!!
		// 	[]Sexpr{
		// 		mkCons(
		// 			atomPrimitives[itemEqQ],
		// 			mkCons(
		// 				mkCons( atomPrimitives[itemCar],
		// 					mkCons(
		// 						mkCons(atomone, atomtwo},
		// 						Nil,
		// 					),
		// 				),
		// 				mkCons( atomone, Nil ),
		// 			),
		// 		),
		// 	},
		// },
	}

	for _, test := range tests {
		_, ch := Parse("test", mkRuneChannel(test.input))
		var got []Sexpr
		for sx := range ch {
			got = append(got, sx)
		}
		if !deepEqualSexpr(got, test.want) {
			t.Errorf("Parsed %q got '%v', wanted '%v'",
				test.input, got, test.want,
			)
		}
	}
}

func ExampleParse_nil() {
	s := "()"
	_, ch := Parse("test", mkRuneChannel(s))
	for sx := range ch {
		fmt.Println(sx)
	}
	// Output:
	// Nil
}

func ExampleParse_list() {
	s := "(+ 1 2)"
	_, ch := Parse("test", mkRuneChannel(s))
	for sx := range ch {
		fmt.Println(sx)
	}
	// Output:
	// Cons(Sym(+), Cons(1, Cons(2, Nil)))
}

func ExampleParse_multiple() {
	s := "#f (+)"
	_, ch := Parse("test", mkRuneChannel(s))
	for sx := range ch {
		fmt.Println(sx)
	}
	// Output:
	// #f
	// Cons(Sym(+), Nil)
}

func ExampleParse_errors() {
	bad_strings := []string{
		"(",
		"())",
		"(1",
	}
	for _, s := range bad_strings {
		_, ch := Parse("test", mkRuneChannel(s))
		for sx := range ch {
			fmt.Println(sx)
		}
	}
	// Output:
	// PARSE ERROR: Unexpected EOF
	// «TODO:  Better parse-error context»
	// Nil
	// PARSE ERROR: popUntil({LPAREN}) from a stack with no {"LPAREN"}
	// «TODO:  Better parse-error context»
	// PARSE ERROR: Unexpected EOF
	// «TODO:  Better parse-error context»
}
