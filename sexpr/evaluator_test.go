package sexpr

import (
	"fmt"
)

func ExampleEvaluator() {
	s := "(cons 1 2)"

	_, ch := Parse("repl", s)
	for sx := range ch {
		fmt.Println("Evaluating", sx)
		val := Evaluate(sx)
		fmt.Println("gave", val)
	}
	// Output:
	// Evaluating Cons(Sym(cons), Cons(N(1), Cons(N(2), Nil)))
	// gave Cons(N(1), N(2))
}
