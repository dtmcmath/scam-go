package sexpr

import (
	"strconv"
	"fmt"
	"errors"
)

// Functions are first class objects.  The file defines operations
// with functions and some with just macros.

type func_expr struct{
	definition string
	bound []sexpr_atom // Need to be x.typ == atomSymbol
	body Sexpr
}
func (f func_expr) Sprint() string {
	return fmt.Sprintf("fn:%s", f.definition)
}
func declareFunction(description string, bound []sexpr_atom, body Sexpr) (*func_expr, error) {
	for _, v := range bound {
		if v.typ != atomSymbol { // TODO:  Symbols become their own things?
			msg := fmt.Sprintf("Invalid parameter-name %q", v)
			return nil, errors.New(msg)
		}
	}
	return &func_expr{description, bound, body}, nil
}
func (f func_expr) apply(args []Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	newCtx := &evaluationContext{make(symbolTable), ctx}
	for idx, sym := range f.bound {
		if val, err := evaluateWithContext(args[idx], ctx) ; err != nil {
			return nil, err
		} else {
			newCtx.bind(sym, val)
		}
	}
	return evaluateWithContext(f.body, newCtx)
}

type macro_expr struct{
	definition string
	applicator evaluator
}
func (m macro_expr) Sprint() string {
	return fmt.Sprintf("ma:%s", m.definition)
}
func (m macro_expr) apply(args Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	return m.applicator(args, ctx)
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
	"-":      mkTodoEvaluator("-"),
	"+":      evalPlus,
	"cons":   evalCons,
	"car":    evalCar,
	"cdr":    evalCdr,
	"eq?":    evalEqQ,
	"quote":  evalQuote,
	"null?":  evalNullQ,
	"atom?":  evalAtomQ, // non-primitive, soon
	"define": evalDefine,
	"let":    evalLet,
	"pair?":  mkTodoEvaluator("pair?"),
	"lambda": mkTodoEvaluator("lambda"),
	"and":    mkTodoEvaluator("and"),
	"or":     mkTodoEvaluator("or"),
	"if":     mkTodoEvaluator("if"),
	"cond":   mkTodoEvaluator("cond"),
	// "else": mkTodoEvaluator(),
}
/////
// Helpers
/////

// requireArgsCount checks various things about arguments.  Namely
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
	var err sexpr_error
	args, unconsify_err := unconsify(lst)
	if unconsify_err != nil {
		return nil, evaluationError{
			context,
			fmt.Sprintf("Strange arguments %q", lst),
		}
	}
	if required > 0 && len(args) != required {
		// return nil, evaluationError{
		// 	context: context,
		// 	message: fmt.Sprintf("Incorrect argument count (%d) in call %s",
		// 		len(args), lst)
		// }
		plural := ""
		if required > 1 {
			plural = "s"
		}
		return nil, evaluationError{
			context: context,
			message: fmt.Sprintf("Expected %d argument%s, got %d",
				required, plural, len(args),
			),
		}
	}
	if ctx != nil {
		for idx, arg := range args {
			if args[idx], err = evaluateWithContext(arg, ctx) ; err != nil {
				return nil, err
			}
		}
	}
	// else
	return args, nil
}
/////
// Definitions of evaluators
/////
func evalCons(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "cons", 2, ctx)
	if err != nil {
		return nil, err
	}
	// else
	return mkCons(args[0], args[1]), nil
}

func evalCar(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "car", 1, ctx)
	if err != nil {
		return nil, err
	}
	// else
	switch first := args[0].(type) {
	case sexpr_atom:
		return nil, evaluationError{
			"car",
			fmt.Sprintf("%s is not a pair", first),
		}
	case sexpr_cons:
		return first.car, nil // first.car has already been eval'd
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", first))
	}
}

func evalCdr(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "cdr", 1, ctx)
	if err != nil {
		return nil, err
	}
	// else
	switch first := args[0].(type) {
	case sexpr_atom:
		return nil, evaluationError{
			"cdr",
			fmt.Sprintf("%s is not a pair", first),
		}
	case sexpr_cons:
		return first.cdr, nil // first.cdr has already been eval'd
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", first))
	}
}

func evalEqQ(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "eq?", 2, ctx)
	if err != nil {
		return nil, err
	}
	// else
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
}

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

func evalPlus(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	context := "+"
	args, err := unconsify(lst)
	if err != nil {
		return nil, evaluationError{
			context,
			fmt.Sprintf("Strange arguments %q", lst),
		}
	}

	acc := intOrFloat{
		asint: 0,
		asfloat: 0,
		isInt: true,
	}

	for _, summand := range args {
		summand, err := evaluateWithContext(summand, ctx)
		if err != nil {
			return nil, err
		}
		switch summand := summand.(type) {
		case sexpr_atom:
			if summand.typ == atomNumber {
				if acc.isInt {
					val, err := strconv.ParseInt(summand.name, 10, 64)
					if err != nil {
						acc.isInt = false
						acc.asfloat = float64(acc.asint)
					} else {
						acc.asint += val
					}
				}
				// NOT else; isInt might have changed!
				if !acc.isInt {
					val, err := strconv.ParseFloat(summand.name, 64)
					if err != nil {
						return nil, evaluationError{
							context,
							fmt.Sprintf("%s masquerades as a Number!!", summand),
						}
					} else {
						acc.asfloat += val
					}
				}
			} else {
				return nil, evaluationError{
					context,
					fmt.Sprintf("%s is not a number", summand),
				}
			}
		default:
			return nil, evaluationError{
				context,
				fmt.Sprintf("%s is not a number", summand),
			}
		}
	}

	return mkAtomNumber(fmt.Sprintf("%s", acc)), nil
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

func evalNullQ(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "null?", 1, ctx)
	if err != nil {
		return nil, err
	}
	// else
	if args[0] == Nil {
		return True, nil
	} else {
		return False, nil
	}
}

func evalAtomQ(lst Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	args, err := requireArgCount(lst, "atom?", 1, ctx)
	if err != nil {
		return nil, err
	}
	// else
	switch val := args[0].(type) {
	case sexpr_atom:
		if val != Nil {
			return True, nil
		} else {
			return False, nil
		}
	default:
		return False, nil
	}
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
