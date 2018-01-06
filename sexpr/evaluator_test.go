package sexpr

import (
	"fmt"
	"testing"
)

// atomone, atomtwo, etc are defined in sexpr_test.go

func TestEvaluateEqQ(t *testing.T) {
	var tests = []struct{
		input string
		want []Sexpr
	} {
		{ "(eq? 1 2)", []Sexpr{ False } },
		{ "(eq? 2 2)", []Sexpr{ True } },
		{ "(+ 1 2)", []Sexpr{ atomthree } },
		{ "(+ 1 (+ 1 1))", []Sexpr{ atomthree } },
		{ "(eq? (+ 1 2) 3)", []Sexpr{ True } },
		{ "(car (cons 1 2))", []Sexpr{ atomone } },
		{ "(eq? (car (cons 1 2)) 1)", []Sexpr{ True } },
		{ "(eq? (cons 1 2) (cons 1 2))", []Sexpr{ False } },
		{ "'()", []Sexpr{ Nil } },
		{ "()", []Sexpr{ Nil } },
		{ "'1", []Sexpr{ atomone } },
		{ "(cons '1 ())", []Sexpr{ mkList(atomone) } },
		// TODO:  Precision, on numbers
		{
			"(+ 1 3.141593)",
			[]Sexpr{ mkAtomNumber("4.141593") },
		},
		{
			"(define a 2)(eq? 2 a)",
			[]Sexpr{ Nil, True },
		},
		{
			"(define a 1)(define b 2)(cons a b)",
			[]Sexpr{ Nil, Nil, mkCons(atomone, atomtwo) },
		},
	}

	for _, test := range tests {
		resetEvaluationContext()
		helpConfirmEvaluation(test.input, test.want, t)
	}
}

func TestEvaluateQuote(t *testing.T) {
	atoma := mkAtomSymbol("a")
	atomb := mkAtomSymbol("b")

	var tests = []struct{
		input string
		want []Sexpr
	} {
		{
			"(define a 2)'(a a)",
			[]Sexpr{ Nil, mkList(atoma, atoma) },
		},
		{
			"(define a 3)(define b 2)(cons 'a (cons 'b '()))",
			[]Sexpr{ Nil, Nil, mkList(atoma, atomb) },
		},
		{
			"(define a 1) 'a",
			[]Sexpr{ Nil, atoma },
		},
	}

	for _, test := range tests {
		resetEvaluationContext()
		helpConfirmEvaluation(test.input, test.want, t)
	}
}

func TestEvaluatorBinding(t *testing.T) {
	var tests = []struct{
		input string
		want []Sexpr
	} {
		{
			`
(define a 1)
(define b 2)
(let ([a 3] [b 4]) (eq? 7 (+ a b)))
`,
			[]Sexpr{ Nil, Nil, True },
		},
		{ // Shadowing
			`
(define a 1)
(define b 2)
(let ([b 5]) (+ a b))
`,
			[]Sexpr{ Nil, Nil, mkAtomNumber("6") },
		},
		{
			"(let ([a 3] [b 4]) (eq? 7 (+ a b)))",
			[]Sexpr{ True },
		},
		{ // Shadowing, no leaking
			`
(define a 1)
(define b 2)
(let ([b 5]) (+ a b))
(+ a b)
`,
			[]Sexpr{ Nil, Nil, mkAtomNumber("6"), atomthree },
		},
	}

	for _, test := range tests {
		resetEvaluationContext()
		helpConfirmEvaluation(test.input, test.want, t)
	}
}

func TestEvaluatorBinding2(t *testing.T) {
	// Confirm we can clear values and that bindings work
	resetEvaluationContext()
	program1 := `
(define alpha 1)
(eq? alpha 1)`
	_, sexprs := Parse("test", mkRuneChannel(program1))
	s1_1 := <- sexprs
	s1_1 = Evaluate(s1_1)
	if s1_1 != Nil {
		t.Errorf("Binding: define got %v, want Nil", s1_1)
	}
	s1_2 := <- sexprs
	s1_2 = Evaluate(s1_2)
	if s1_2 != True {
		t.Errorf("Binding: (eq? alpha 1) got %v, want #t", s1_2)
	}
	_, ok := <- sexprs
	if ok {
		t.Error("Binding: returned 3 (or more) expressions, want 2")
	}

	resetEvaluationContext()
	program2 := "(eq? alpha 1)"
	_, sexprs = Parse("test", mkRuneChannel(program2))
	s2_1 := <- sexprs
	s2_1 = Evaluate(s2_1)
	switch s2_1 := s2_1.(type) {
	case evaluationError:
		// the one we want
		break
	default:
		t.Errorf("Unbound evaluation gave a %T: %s, want 'evaluationError'",
			s2_1, s2_1)
	}
	_, ok = <- sexprs
	if ok {
		t.Error("Unbound evaluation: returned 2 (or more) expressions, want 1")
	}
	
	resetEvaluationContext()
	program3 := `
(eq? alpha 1)
(define alpha 1)
(eq? alpha 1)
`
	_, sexprs = Parse("test", mkRuneChannel(program3))
	s3_1 := <- sexprs
	s3_1 = Evaluate(s3_1)
	switch s3_1 := s3_1.(type) {
	case evaluationError:
		// the one we want
		break
	default:
		t.Errorf("Unbound/bound evaluation gave a %T: %s, want 'evaluationError'",
			s3_1, s3_1)
	}
	s3_2 := <- sexprs
	s3_2 = Evaluate(s3_2)
	if s3_2 != Nil {
		t.Errorf("Binding: define got %v, want Nil", s3_2)
	}
	s3_3 := <- sexprs
	s3_3 = Evaluate(s3_3)
	if s3_3 != True {
		t.Errorf("Binding: (eq? alpha 1) got %v, want #t", s3_3)
	}
	_, ok = <- sexprs
	if ok {
		t.Error("Binding: returned 4 (or more) expressions, want 3")
	}
}

func TestEvaluatorLambda(t *testing.T) {
	// Confirm we can clear values and that bindings work
	resetEvaluationContext()
	tests := []struct{
		input string
		want []Sexpr
	} {
		{
			`
(define add1 (lambda (x) (+ x 1)))
(eq? (add1 1) 2)
`,
			[]Sexpr{ Nil, True },
		},
		{
			`
(define add1 (lambda (x) (+ x 1)))
(add1 2)
`,
			[]Sexpr{ Nil, atomthree },
		},
	}

	for _, test := range tests {
		resetEvaluationContext()
		helpConfirmEvaluation(test.input, test.want, t)
	}
}

func ExampleEvaluatorBinding() {
	resetEvaluationContext()
	program := `
(eq? a 1)
(define a 1)
(eq? a 1)`

	_, sexprs := Parse("test", mkRuneChannel(program))
	for sx := range sexprs {
		val := Evaluate(sx)
		in := Sprint(sx)
		out := Sprint(val)
		fmt.Printf("> %s\n%s\n", in, out)
	}
	// Output:
	// > (eq? a 1)
	// Exception in lookup: Variable Sym(a) is not bound
	// > (define a 1)
	// ()
	// > (eq? a 1)
	// #t
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

/////
// Helpers
/////
func helpConfirmEvaluation(input string, want []Sexpr, t *testing.T) {
		_, sexprs := Parse("test", mkRuneChannel(input))
		idx := 0
		for sx := range sexprs {
			got := Evaluate(sx)
			if !equalSexpr(got, want[idx]) {
				t.Errorf("Evaluate[%s][#%d]=%v, want %v",
					input, idx, got, want[idx],
				)
			}
			idx += 1
		}
		if idx != len(want) {
			t.Errorf("Evaluate[%s] got %d results, want %d",
				input, idx, len(want),
			)
		}
}
