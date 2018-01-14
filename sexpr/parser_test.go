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
		{ "()", []Sexpr{ atomConstantNil } },
		{ "(1)", []Sexpr{ mkCons(atomone, atomConstantNil) } },
		{ "1", []Sexpr{ atomone } },
		{
			" (1) 2 ",
			[]Sexpr{
				mkCons(atomone, atomConstantNil),
				atomtwo,
			},
		},
		{
			"#t #f",
			[]Sexpr{atomConstantTrue, atomConstantFalse},
		},
		{
			"(cons 1 2)",
			[]Sexpr{
				mkCons(
					primcons,
					mkCons(atomone,
						mkCons(atomtwo,atomConstantNil),
					),
				),
			},
		},
		{
			"'()",
			[]Sexpr{
				mkList(atomConstantQuote, atomConstantNil),
			},
		},
		{
			"'1",
			[]Sexpr{ mkList(atomConstantQuote, atomone) },
		},
		{
			"(cons '1 ())",
			[]Sexpr{
				mkList(primcons, mkList(atomConstantQuote, atomone), atomConstantNil),
			},
		},
		{
			"'(#t)",
			[]Sexpr{mkList(atomConstantQuote, mkList(atomConstantTrue))},
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
		// 						atomConstantNil,
		// 					),
		// 				),
		// 				mkCons( atomone, atomConstantNil ),
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
