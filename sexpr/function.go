package sexpr

import (
	"strconv"
	"fmt"
)

// Functions are first class objects.  The file defines operations
// with functions and some with just macros.

type applicator func([]Sexpr) (Sexpr, sexpr_error)
type func_expr struct{
	definition string
	// A function is handed its arguments pre-evaluated
	apply applicator
}
func (f func_expr) Sprint() string {
	return fmt.Sprintf("fn:%s", f.definition)
}

type macro_expr struct{
	definition string
	apply evaluator
}
func (m macro_expr) Sprint() string {
	return fmt.Sprintf("ma:%s", m.definition)
}

// Certain primitive Atom(symbol)s are built-in.  They can't be
// implemented given other things defineable in the SCAM language, so
// we define them here.

func mkTodoEvaluator(s string) evaluator {
	return func (ignore Sexpr, i2 *evaluationContext) (Sexpr, sexpr_error) {
		return nil, evaluationError{s, "is not yet implemented"}
	}
}

var primitiveMacros = map[string]evaluator {
	"quote":  evalQuote,
	"define": evalDefine,
	"let":    evalLet,
	"lambda": evalLambda,
	"and":    mkLazyReduce("and", True, func(acc Sexpr, val Sexpr) (Sexpr, bool) {
		if isFalsey(val) {
			return False, true
		} else {
			return acc, false
		}
	}),
	"or":     mkLazyReduce("or", False, func(acc Sexpr, val Sexpr) (Sexpr, bool) {
		if !isFalsey(val) {
			return True, true
		} else {
			return acc, false
		}
	}),
	"if":     mkTodoEvaluator("if"),
	"cond":   evalCond,
}

func mkTodoApplicator(s string) applicator {
	return func(ignore []Sexpr) (Sexpr, sexpr_error) {
		return nil, evaluationError{s, "is not yet implemented"}
	}
}

var primitiveFunctions = map[string]applicator {
	"-":      mkTodoApplicator("-"),
	"+":      mkArithmeticReduce(
		"+",
		intOrFloat{
			asint: 0,
			asfloat: 0,
			isInt: true,
		},
		reducePlus,
	),
	"cons":   mkNaryFn("cons", 2, func(args []Sexpr) (Sexpr, sexpr_error) {
		return mkCons(args[0], args[1]), nil
	}),
	"car":    mkConsSelector("car", func (c sexpr_cons) Sexpr { return c.car }),
	"cdr":    mkConsSelector("cdr", func (c sexpr_cons) Sexpr { return c.cdr }),
	"eq?":    mkNaryFn("eq?", 2, func(args []Sexpr) (Sexpr, sexpr_error) {
		if args[0] == args[1] {
			return True, nil
		} else {
			// TODO:  Be nice with numbers?  The Little Schemer says
			//
			// ; (eq? 5 6)
			// ; not-applicable because eq? works only on non-numeric atom
			//
			// so we're already being generous by saying True sometimes.
			// For the record, the problem is that "3.0" and "3" and
			// "3.00" are all different atoms.
			return False, nil
		}
	}),
	"null?":  mkNaryFn("null?", 1, func(args []Sexpr) (Sexpr, sexpr_error) {
		if args[0] == Nil {
			return True, nil
		} else {
			return False, nil
		}
	}),
	"pair?":  mkNaryFn("pair?", 1, func(args []Sexpr) (Sexpr, sexpr_error) {
		switch args[0].(type) {
		case sexpr_cons:
			return True, nil
		}
		// else
		return False, nil
	}),
	"not":    mkNaryFn("not", 1, func(args []Sexpr) (Sexpr, sexpr_error) {
		if args[0] == True {
			return False, nil
		} else {
			return True, nil
		}
	}),
}
/////
// Helpers
/////

// mkNaryFn makes an n-ary "normal" function, one that operates on
// its arguments after they've been evaluated.
func mkNaryFn(name string, n int, fn func([]Sexpr) (Sexpr, sexpr_error)) applicator {
	return func(args []Sexpr) (Sexpr, sexpr_error) {
		ans, err := fn(args)
		if err != nil {
			// divide-by-zero, for instance
			return nil, err
		}
		return ans, nil
	}
}

func mkConsSelector(name string, sel func(sexpr_cons) Sexpr) applicator {
	return mkNaryFn(name, 1, func(args []Sexpr) (Sexpr, sexpr_error) {
		switch first := args[0].(type) {
		case sexpr_cons: return sel(first), nil
		default:
			return nil, evaluationError{
				name,
				fmt.Sprintf("%s is not a pair", first),
			}
		}
	})
}


func mkLazyReduce(
	name string,
	acc Sexpr,
	reducer func(acc Sexpr, val Sexpr) (Sexpr, bool),
) evaluator {
	return func(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
		args, err := unconsify(lst)
		if err != nil {
			return nil, evaluationError{name, err.Error()}
		}
		ans := acc
		for _, term := range args {
			val, err := evaluateWithContext(term, ctx)
			if err != nil {
				return nil, err
			}
			// else
			if ans, done := reducer(ans, val) ; done {
				return ans, nil
			}
		}
		return ans, nil
	}
}

/////
// Definitions of complicated things, too simple for inlining.  Mostly
// macros, but a few others.
/////
type intOrFloat struct{
	asint   int64
	asfloat float64
	isInt   bool
}

func (i intOrFloat) String() string {
	if i.isInt {
		return fmt.Sprintf("%d", i.asint)
	} else {
		return fmt.Sprintf("%f", i.asfloat)
	}
}

func mkArithmeticReduce(
	name string,
	acc intOrFloat,
	reducer arithmeticReducerFunction,
) applicator {
	return func(args []Sexpr) (Sexpr, sexpr_error) {
		ans := acc // makes a copy!
		for _, val := range args {
			if err, done := reducer(&ans, val) ; done {
				break
			} else if err != nil {
				// Like, "divide by zero" or "incompatible types"
				return nil, evaluationError{name, err.Error()}
			}
		}
		return mkAtomNumber(fmt.Sprintf("%s", ans)), nil
	}
}

type arithmeticReducerFunction func(*intOrFloat, Sexpr) (sexpr_error, bool)

func mkArithmeticReducerFunction(
	inter func(int64, int64) (int64, error),
	floater func(float64, float64) (float64, error),
) arithmeticReducerFunction {

	asIntHandler := func(acc *intOrFloat, x sexpr_atom) error {
		if !acc.isInt {
			// This is not the accumulator you're looking for
			return nil
		}
		// else, let's try
		val, parseErr := strconv.ParseInt(x.name, 10, 64)
		if parseErr != nil {
			// Time to give up on the idea that we have an int.
			acc.isInt = false
			acc.asfloat = float64(acc.asint)
			/// and continue
			return nil
		}
		// else
		if nxt, err := inter(acc.asint, val) ; err != nil {
			// an arithmetic error; fatal
			return err
		} else {
			acc.asint = nxt
			return nil
		}
	}
	asFloatHandler := func(acc *intOrFloat, x sexpr_atom) error {
		if acc.isInt {
			return nil
		}
		// else
		val, err := strconv.ParseFloat(x.name, 64)
		if err != nil {
			return err
		}
		if nxt, err := floater(acc.asfloat, val) ; err != nil {
			// an arithmetic error; fatal
			return err
		} else {
			// else
			log.Printf("...made %f\n", nxt)
			acc.asfloat = nxt
			return nil
		}
	}

	return func(acc *intOrFloat, x Sexpr) (sexpr_error, bool) {
		switch x := x.(type) {
		case sexpr_atom:
			if x.typ == atomNumber {
				// If acc is already a float, asIntHandler is a no-op.
				// If acc is still an int, asIntHandler might turn it
				// into a float.
				// Either way, try asIntHandler.
				// Then try asFloatHandler; if acc is still an int by
				// then, it's a no-op, otherwise it's the right thing.
				if err := asIntHandler(acc, x); err != nil {
					return evaluationError{"int", err.Error()}, true
				}
				if err := asFloatHandler(acc, x) ; err != nil {
					return evaluationError{"float", err.Error()}, true
				}
				// else... we computed, and can contiune accumulating.
				return nil, false
			}
		}
		// else, it's either a non-atom or a non-number atom
		return evaluationError{
			"",
			fmt.Sprintf("%s is not a number", x),
		},
		true
	}
}

var reducePlus = mkArithmeticReducerFunction(
	func(x int64, y int64) (int64, error) { return x+y, nil },
	func(x float64, y float64) (float64, error) { return x+y, nil },
)

// evalQuote is a macro; it does not evaluate all its arguments
func evalQuote(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := unconsifyN(lst, 1)
	if err != nil {
		return nil, evaluationError{"quote", err.Error()}
	}
	// else
	// return the first argument, unevaluated
	return args[0], nil
}

// evalDefine is a macro; it does not evaluate all its arguments
func evalDefine(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := unconsifyN(lst, 2)
	if err != nil {
		return nil, evaluationError{"define", err.Error()}
	}
	// else
	key := args[0]
	switch key := key.(type) {
	case sexpr_atom:
		val, err := evaluateWithContext(args[1], ctx)
		if err != nil {
			return nil, err
		}
		// log.Printf("DEFINE %q <-- %s", key, val)
		err2 := globalEvaluationContext.bind(key, val)
		if err2 != nil {
			return nil, evaluationError{
				"define(binding)",
				err2.Error(),
			}
		}
		return Nil, nil
		// TODO:  "nil" isn't "Nil"; define has no value
		// return nil, nil
	default:
		return nil, evaluationError{
			"define",
			fmt.Sprintf("Cannot bind non-atom %q", key),
		}
	}
}

func evalLet(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := unconsifyN(lst, 2)
	if err != nil {
		return nil, evaluationError{"let", err.Error()}
	}
	// else
	bindings, err := unconsify(args[0])
	if err != nil {
		return nil, evaluationError{"let(args)", err.Error()}
	}

	newCtx := evaluationContext{make(symbolTable), ctx}
	for _, b := range bindings {
		// log.Println("Create binding from", b)
		kv, err := unconsifyN(b, 2)
		if err != nil {
			return nil, evaluationError{"let(binding)", err.Error()}
		}
		switch key := kv[0].(type) {
		case sexpr_atom:
			val, err := evaluateWithContext(kv[1], ctx)
			if err != nil {
				return nil, err
			}
			err2 := newCtx.bind(key, val)
			if err2 != nil {
				return nil, evaluationError{
					"let(binding)",
					err2.Error(),
				}
			}
		default:
			return nil, evaluationError{
				"let",
				fmt.Sprintf("Cannot bind non-atom %q", key),
			}
		}
	}
	return evaluateWithContext(args[1], &newCtx)
}

func evalLambda(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := unconsifyN(lst, 2) // eventually "many"
	if err != nil {
		return nil, evaluationError{"lambda", err.Error()}
	}
	var bound []sexpr_atom
	if decl, unconsify_err := unconsify(args[0]) ; unconsify_err != nil {
		return nil, evaluationError{
			"lambda",
			fmt.Sprintf("Strange arguments %q:", lst, unconsify_err.Error()),
		}
	} else {
		for _, v := range decl {
			switch v := v.(type) {
			case sexpr_atom:
				if v.typ == atomSymbol {
					bound = append(bound, v)
				} else {
					msg := fmt.Sprintf("Invalid parameter-name %q", v)
					return nil, evaluationError{"lambda", msg}
				}
			default:
				return nil, evaluationError{
					"lambda",
					fmt.Sprintf("invalid parameter list in (λ %s %s)",
						args[0], args[1]),
				}
			}
		}
	}

	body := args[1]
	definition := fmt.Sprintf("(λ (%s) %s)", bound, args[1].Sprint())
	apply := func(args []Sexpr) (Sexpr, sexpr_error) {
		newCtx := &evaluationContext{make(symbolTable), ctx}
		for idx, sym := range bound {
			if err := newCtx.bind(sym, args[idx]) ; err != nil {
				return nil, evaluationError{
					fmt.Sprintf("%s(bind %q)", definition, sym),
					err.Error(),
				}
			}
		}
		return evaluateWithContext(body, newCtx)
	}
	return func_expr{definition, apply}, nil
}

// If none of the first-elements to "cond" are truthy, the eventual
// value is undefined.  We'll call it "Nil" (I guess)
func evalCond(lst Sexpr, ctx *evaluationContext) (s Sexpr, serr sexpr_error) {
	args, err := unconsify(lst)
	if err != nil {
		return nil, evaluationError{"cond", err.Error()}
	}
	for _, pair := range args {
		test, err := unconsifyN(pair, 2)
		if err != nil {
			return nil, evaluationError{
				"cond",
				fmt.Sprint("Unrecognizable test %q (%s)", pair, err.Error()),
			}
		}
		if test[0] == Else {
			return evaluateWithContext(test[1], ctx)
		}
		// else
		if predicate, serr := evaluateWithContext(test[0], ctx) ; serr != nil {
			return nil, serr
		} else if !isFalsey(predicate) {
			return evaluateWithContext(test[1], ctx)
		}
	}
	return Nil, nil // or nil, nil, if we get "define" doing that.
}
