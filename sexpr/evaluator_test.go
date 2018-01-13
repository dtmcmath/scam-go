package sexpr

import (
	"fmt"
	"testing"
	"regexp"
)

// atomone, atomtwo, etc are defined in sexpr_test.go

func TestEvaluateArithmetic(t *testing.T) {
	var tests = []struct{
		input string
		want []Sexpr
	} {
		{ "(+ 1 2)", []Sexpr{ atomthree } },
		{ "(+ 1 (+ 1 1))", []Sexpr{ atomthree } },
		// TODO:  Precision, on numbers
		{
			"(+ 1 3.141593)",
			[]Sexpr{ mkAtomNumber("4.141593") },
		},
		{
			"(+ 2.718281 3.141593)",
			[]Sexpr{ mkAtomNumber("5.859874") },
		},
		{ "(+ 1 2 3 4 5)", []Sexpr{ mkAtomNumber("15") } },
		{ "(- 5 2)", []Sexpr{ atomthree } },
		{ "(= 1 2)", []Sexpr{ False } },
		{ "(= 2 2)", []Sexpr{ True } },
		{ "(= (+ 1 2) 3)", []Sexpr{ True } },
		{ "(expt 2 3)", []Sexpr{ mkAtomNumber("8") } },
		{ "(= (expt 2 3) 8)", []Sexpr{ True } },
		{ "(expt 4 0.5)", []Sexpr{ mkAtomNumber("2.000000") } },
		{ "(= (expt 4 0.5) 2.000000)", []Sexpr{ True } },
		{ "(* 4 3) (* 2.718281 3.141593)", []Sexpr{ mkAtomNumber("12"), mkAtomNumber("8.539733") } },
		{ "(/ 4 3) (/ (* 40 37) 40)", []Sexpr{ mkAtomNumber("1.333333"), mkAtomNumber("37") } },
	}

	for _, test := range tests {
		resetEvaluationContext()
		helpConfirmEvaluation(test.input, test.want, t)
	}
}

func ExampleArithmetic() {
	s := "(- 3 2 1)"

	_, ch := Parse("repl", mkRuneChannel(s))
	for sx := range ch {
		fmt.Println("Evaluating", sx.Sprint())
		val := Evaluate(sx)
		fmt.Println("gave", val)
	}
	// Output:
	// Evaluating (- 3 2 1)
	// gave Exception in -: Expected 2 arguments, got 3
}

func TestEvaluateEqQ(t *testing.T) {
	var tests = []struct{
		input string
		want []Sexpr
	} {
		{ "(car (cons 1 2))", []Sexpr{ atomone } },
		{ "(= (car (cons 1 2)) 1)", []Sexpr{ True } },
		{ "(eq? (cons 1 2) (cons 1 2))", []Sexpr{ False } },
		{ "'()", []Sexpr{ Nil } },
		{ "()", []Sexpr{ Nil } },
		{ "'1", []Sexpr{ atomone } },
		{ "(cons '1 ())", []Sexpr{ mkList(atomone) } },
		{
			"(define a 2)(= 2 a)",
			[]Sexpr{ Nil, True },
		},
		{
			"(define a 1)(define b 2)(cons a b)",
			[]Sexpr{ Nil, Nil, mkCons(atomone, atomtwo) },
		},
		{ "(eq? '() '())", []Sexpr{ True } },
	}

	for _, test := range tests {
		resetEvaluationContext()
		helpConfirmEvaluation(test.input, test.want, t)
	}
}

func TestEvaluateLogic(t *testing.T) {
	var tests = []struct{
		input string
		want []Sexpr
	} {
		{ "(and)", []Sexpr{ True } },
		{ "(and #t)", []Sexpr{ True } },
		{ "(and (eq? 'a 'a) #t)", []Sexpr{ True } },
		{ "(and (eq? 'a 'b) #t)", []Sexpr{ False } },
		{ "(and (eq? 'a 'b) unbound)", []Sexpr{ False } },
		{ "(not #t)", []Sexpr{ False } },
		{ "(not ())", []Sexpr{ True } },
		{ "(not #f)", []Sexpr{ True } },
		{
			`
(cond
 (#t 1)
 (#f 2)
 (else 3))`,
			[]Sexpr{ atomone },
		},
		{
			`
(cond
 (#f 1)
 (#t 2)
 (else 3))`,
			[]Sexpr{ atomtwo },
		},
		{
			`
(cond
 (#f 1)
 (#f 2)
 (else 3))`,
			[]Sexpr{ atomthree },
		},
		{
			`
(cond
 (#t 1)
 (#t 2)
 (else 3))`,
			[]Sexpr{ atomone },
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
(let ([a 3] [b 4]) (= 7 (+ a b)))
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
			"(let ([a 3] [b 4]) (= 7 (+ a b)))",
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
(= alpha 1)`
	_, sexprs := Parse("test", mkRuneChannel(program1))
	s1_1 := <- sexprs
	s1_1 = Evaluate(s1_1)
	if s1_1 != Nil {
		t.Errorf("Binding: define got %v, want Nil", s1_1)
	}
	s1_2 := <- sexprs
	s1_2 = Evaluate(s1_2)
	if s1_2 != True {
		t.Errorf("Binding: (= alpha 1) got %v, want #t", s1_2)
	}
	_, ok := <- sexprs
	if ok {
		t.Error("Binding: returned 3 (or more) expressions, want 2")
	}

	resetEvaluationContext()
	program2 := "(= alpha 1)"
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
(= alpha 1)
(define alpha 1)
(= alpha 1)
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
		t.Errorf("Binding: (= alpha 1) got %v, want #t", s3_3)
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
(= (add1 1) 2)
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
		{
			`
(define atom?
 (lambda (x)
    (and (not (pair? x)) (not (null? x)))))
(atom? 'a)
(atom? '())
(atom? '(a b))
`,
			[]Sexpr{ Nil, True, False, False },
		},
		{
			`
(define nonpair? (lambda (y) (not (pair? y))))
(nonpair? 'a)
(nonpair? '())
(nonpair? '(a b))
`,
			[]Sexpr{ Nil, True, True, False },
		},
		{ // Chapter 10
			`
((lambda (nothing)
   (cons nothing (quote ())))
 (quote (from nothing comes something)))
`,
			[]Sexpr{
				mkCons(
					consify([]Sexpr{mkAtomSymbol("from"), mkAtomSymbol("nothing"), mkAtomSymbol("comes"), mkAtomSymbol("something")}),
					Nil,
				),
			},
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
(= a 1)
(define a 1)
(= a 1)`

	_, sexprs := Parse("test", mkRuneChannel(program))
	for sx := range sexprs {
		val := Evaluate(sx)
		in := Sprint(sx)
		out := Sprint(val)
		fmt.Printf("> %s\n%s\n", in, out)
	}
	// Output:
	// > (= a 1)
	// Exception in lookup: Variable Sym(a) is not bound
	// > (define a 1)
	// ()
	// > (= a 1)
	// #t
}

func ExampleEvaluator() {
	s := "(cons 1 2) (= (car (cons 1 2)) 1)"

	_, ch := Parse("repl", mkRuneChannel(s))
	for sx := range ch {
		fmt.Println("Evaluating", sx)
		val := Evaluate(sx)
		fmt.Println("gave", val)
	}
	// Output:
	// Evaluating Cons(Sym(cons), Cons(1, Cons(2, Nil)))
	// gave Cons(1, 2)
	// Evaluating Cons(Sym(=), Cons(Cons(Sym(car), Cons(Cons(Sym(cons), Cons(1, Cons(2, Nil))), Nil)), Cons(1, Nil)))
	// gave #t
}

func TestLambdaError(t *testing.T) {
	s := "((lambda (x y) x) 'a)"
	_, ch := Parse("repl", mkRuneChannel(s))
	for sx := range ch {
		val := Evaluate(sx)
		switch val := val.(type) {
		case evaluationError:
			good, _ := regexp.MatchString("Exception.*1 arguments.*expected 2", val.Error())
			if !good {
				t.Errorf("Evaluating %q expected error, got %q", s, val.Error())
			}
		default:
			t.Errorf("Evaluating %q expected error, got %T", val)
		}
	}
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
