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
		{ "(= 1 2)", []Sexpr{ atomConstantFalse } },
		{ "(= 2 2)", []Sexpr{ atomConstantTrue } },
		{ "(= (+ 1 2) 3)", []Sexpr{ atomConstantTrue } },
		{ "(expt 2 3)", []Sexpr{ mkAtomNumber("8") } },
		{ "(= (expt 2 3) 8)", []Sexpr{ atomConstantTrue } },
		{ "(expt 4 0.5)", []Sexpr{ mkAtomNumber("2.000000") } },
		{ "(= (expt 4 0.5) 2.000000)", []Sexpr{ atomConstantTrue } },
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
		{ "(= (car (cons 1 2)) 1)", []Sexpr{ atomConstantTrue } },
		{ "(eq? (cons 1 2) (cons 1 2))", []Sexpr{ atomConstantFalse } },
		{ "'()", []Sexpr{ atomConstantNil } },
		{ "()", []Sexpr{ atomConstantNil } },
		{ "'1", []Sexpr{ atomone } },
		{ "(cons '1 ())", []Sexpr{ mkList(atomone) } },
		{
			"(define a 2)(= 2 a)",
			[]Sexpr{ atomConstantNil, atomConstantTrue },
		},
		{
			"(define a 1)(define b 2)(cons a b)",
			[]Sexpr{ atomConstantNil, atomConstantNil, mkCons(atomone, atomtwo) },
		},
		{ "(eq? '() '())", []Sexpr{ atomConstantTrue } },
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
		{ "(and)", []Sexpr{ atomConstantTrue } },
		{ "(and #t)", []Sexpr{ atomConstantTrue } },
		{ "(and (eq? 'a 'a) #t)", []Sexpr{ atomConstantTrue } },
		{ "(and (eq? 'a 'b) #t)", []Sexpr{ atomConstantFalse } },
		{ "(and (eq? 'a 'b) unbound)", []Sexpr{ atomConstantFalse } },
		{ "(not #t)", []Sexpr{ atomConstantFalse } },
		{ "(not ())", []Sexpr{ atomConstantTrue } },
		{ "(not #f)", []Sexpr{ atomConstantTrue } },
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
			[]Sexpr{ atomConstantNil, mkList(atoma, atoma) },
		},
		{
			"(define a 3)(define b 2)(cons 'a (cons 'b '()))",
			[]Sexpr{ atomConstantNil, atomConstantNil, mkList(atoma, atomb) },
		},
		{
			"(define a 1) 'a",
			[]Sexpr{ atomConstantNil, atoma },
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
			[]Sexpr{ atomConstantNil, atomConstantNil, atomConstantTrue },
		},
		{ // Shadowing
			`
(define a 1)
(define b 2)
(let ([b 5]) (+ a b))
`,
			[]Sexpr{ atomConstantNil, atomConstantNil, mkAtomNumber("6") },
		},
		{
			"(let ([a 3] [b 4]) (= 7 (+ a b)))",
			[]Sexpr{ atomConstantTrue },
		},
		{ // Shadowing, no leaking
			`
(define a 1)
(define b 2)
(let ([b 5]) (+ a b))
(+ a b)
`,
			[]Sexpr{ atomConstantNil, atomConstantNil, mkAtomNumber("6"), atomthree },
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
	if s1_1 != atomConstantNil {
		t.Errorf("Binding: define got %v, want atomConstantNil", s1_1)
	}
	s1_2 := <- sexprs
	s1_2 = Evaluate(s1_2)
	if s1_2 != atomConstantTrue {
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
	if s3_2 != atomConstantNil {
		t.Errorf("Binding: define got %v, want atomConstantNil", s3_2)
	}
	s3_3 := <- sexprs
	s3_3 = Evaluate(s3_3)
	if s3_3 != atomConstantTrue {
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
			[]Sexpr{ atomConstantNil, atomConstantTrue },
		},
		{
			`
(define add1 (lambda (x) (+ x 1)))
(add1 2)
`,
			[]Sexpr{ atomConstantNil, atomthree },
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
			[]Sexpr{ atomConstantNil, atomConstantTrue, atomConstantFalse, atomConstantFalse },
		},
		{
			`
(define nonpair? (lambda (y) (not (pair? y))))
(nonpair? 'a)
(nonpair? '())
(nonpair? '(a b))
`,
			[]Sexpr{ atomConstantNil, atomConstantTrue, atomConstantTrue, atomConstantFalse },
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
					atomConstantNil,
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
