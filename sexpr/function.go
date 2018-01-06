package sexpr

import (
	"strconv"
	"fmt"
)

// Functions are first class objects.  The file defines operations
// with functions and some with just macros.

type func_applicator func([]Sexpr) (Sexpr, sexpr_error)
type func_expr struct{
	definition string
	// A function is handed its arguments pre-evaluated
	apply func_applicator
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

func mkTodoApplicator(s string) func_applicator {
	return func(ignore []Sexpr) (Sexpr, sexpr_error) {
		return nil, evaluationError{s, "is not yet implemented"}
	}
}

var primitiveFunctions = map[string]func_applicator {
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

// requireArgCount checks various things about arguments.  Namely
// that the argument count matches required and, if "ctx" is non-nil,
// that the arguments are evaluable.
//
// It returns the arguments, evaluated if requested, or an error if
// there is a probblem.
func requireArgCount(
	lst Sexpr,
	context string,
	required int,
	ctx *evaluationContext) ([]Sexpr, sexpr_error) {
	var args []Sexpr
	var unconsify_err error
	if required >= 0 {
		args, unconsify_err = unconsifyN(lst, required)
	} else {
		args, unconsify_err = unconsify(lst)
	}
	if unconsify_err != nil {
		return nil, evaluationError{
			context, unconsify_err.Error(),
		}
	}
	if ctx != nil {
		var err sexpr_error
		for idx, arg := range args {
			if args[idx], err = evaluateWithContext(arg, ctx) ; err != nil {
				return nil, err
			}
		}
	}
	// else
	return args, nil
}

// mkNaryFn makes an n-ary "normal" function, one that operates on
// its arguments after they've been evaluated.
func mkNaryFn(name string, n int, fn func([]Sexpr) (Sexpr, sexpr_error)) func_applicator {
	return func(args []Sexpr) (Sexpr, sexpr_error) {
		ans, err := fn(args)
		if err != nil {
			// divide-by-zero, for instance
			return nil, err
		}
		return ans, nil
	}
}

func mkConsSelector(name string, sel func(sexpr_cons) Sexpr) func_applicator {
	return mkNaryFn(name, 1, func(args []Sexpr) (Sexpr, sexpr_error) {
		switch first := args[0].(type) {
		case sexpr_cons: return sel(first), nil
		default:
			return nil, evaluationError{
				"car",
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
	reducer func(acc intOrFloat, val Sexpr) (intOrFloat, sexpr_error, bool),
) func_applicator {
	return func(args []Sexpr) (Sexpr, sexpr_error) {
		ans := acc
		var ( // Declare outside the loop, else ans is shadowed
			err sexpr_error
			done bool
		)
		for _, val := range args {
			if ans, err, done = reducer(ans, val) ; done {
				break
			} else if err != nil {
				// Like, "divide by zero" or "incompatible types"
				return nil, evaluationError{name, err.Error()}
			}
		}
		if err != nil {
			// Pretend the reducer's error is our error.
			err = evaluationError{name, err.Error()}
		}
		return mkAtomNumber(fmt.Sprintf("%s", ans)), err
	}
}

func reducePlus(acc intOrFloat, summand Sexpr) (intOrFloat, sexpr_error, bool) {
	switch summand := summand.(type) {
	case sexpr_atom:
		if summand.typ == atomNumber {
			if acc.isInt {
				val, err := strconv.ParseInt(summand.name, 10, 64)
				if err != nil {
					// Time to give up on the idea that we have an int.
					acc.isInt = false
					acc.asfloat = float64(acc.asint)
				} else {
					acc.asint += val
					return acc, nil, false
				}
			}
			// NOT else; isInt might have changed!
			if !acc.isInt {
				val, err := strconv.ParseFloat(summand.name, 64)
				if err != nil {
					return acc, // should be ignored; can't be nil
					evaluationError{
						"",
						fmt.Sprintf("%s masquerades as a Number!!", summand),
					},
					true
				} else {
					acc.asfloat += val
					return acc, nil, false
				}
			}
		}
	}
	// else, it's either a non-atom or a non-number atom
	return acc, evaluationError{
		"",
		fmt.Sprintf("%s is not a number", summand),
	},
	true
}

// evalQuote is a macro; it does not evaluate all its arguments
func evalQuote(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "quote", 1, nil)
	if err != nil {
		return nil, err
	}
	// else
	// return the first argument, unevaluated
	return args[0], nil
}

// evalDefine is a macro; it does not evaluate all its arguments
func evalDefine(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "define", 2, nil)
	if err != nil {
		return nil, err
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
	args, err := requireArgCount(lst, "let", 2, nil)
	if err != nil {
		return nil, err
	}
	// else
	bindings, err := requireArgCount(args[0], "let(prep)", -1, nil)
	if err != nil {
		return nil, err
	}

	newCtx := evaluationContext{make(symbolTable), ctx}
	for _, b := range bindings {
		// log.Println("Create binding from", b)
		kv, err := requireArgCount(b, "let(binding)", 2, nil)
		if err != nil {
			return nil, err
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
	val, err := evaluateWithContext(args[1], &newCtx)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func evalLambda(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "lambda", 2, nil) // eventually "many"
	if err != nil {
		return nil, err
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
