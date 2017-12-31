package sexpr

import (
	"fmt"
	"testing"
)

func TestSimpleParse(t *testing.T) {
	atomone := mkAtomNumber("1")
	atomtwo := mkAtomNumber("2")
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
					atomPrimitives["cons"],
					mkCons(atomone,
						mkCons(atomtwo,Nil),
					),
				),
			},
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
		_, ch := Parse("test", test.input)
		var got []Sexpr
		for sx := range ch {
			got = append(got, sx)
		}
		if !deepEqualSexpr(got, test.want) {
			t.Errorf("Parsed '%s' got '%v', wanted '%v'",
				test.input, got, test.want,
			)
		}
	}
}

func ExampleParse_nil() {
	s := "()"
	_, ch := Parse("test", s)
	for sx := range ch {
		fmt.Println(sx)
	}
	// Output:
	// Nil
}

func ExampleParse_list() {
	s := "(+ 1 2)"
	_, ch := Parse("test", s)
	for sx := range ch {
		fmt.Println(sx)
	}
	// Output:
	// Cons(Sym(+), Cons(N(1), Cons(N(2), Nil)))
}

func ExampleParse_multiple() {
	s := "#f (+)"
	_, ch := Parse("test", s)
	for sx := range ch {
		fmt.Println(sx)
	}
	// Output:
	// #f
	// Cons(Sym(+), Nil)
}
