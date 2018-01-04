package sexpr

import (
	"fmt"
	"strconv"
	"log"
)

var globalEvaluationContext evaluationContext // define always writes here

func init() {
	resetEvaluationContext()
}
// reset the rootSymbolTable.  This is useful for testing!
func resetEvaluationContext() {
	globalEvaluationContext = evaluationContext{
		make(map[sexpr_atom]Sexpr),
		nil,
	}
}

func Evaluate(s Sexpr) Sexpr {
	if val, err := evaluateWithContext(s, &globalEvaluationContext) ; err != nil {
		// The error is an S-expression, because it can Sprint!
		return err
		// // We should really do something more dramatic ("panic"?) but
		// // that won't help debugging, so we're permissive (and abuse
		// // "atom")
		// log.Print(err)
		// return mkAtomSymbol(err.Error())
	} else {
		return val
	}
}

func evaluateWithContext(s Sexpr, ctx *evaluationContext) (Sexpr, sexpr_error) {
	switch s := s.(type) {
	case sexpr_atom: return s.evaluate(ctx)
	case sexpr_cons:
		car, err := evaluateWithContext(s.car, ctx)
		if err != nil {
			return nil, err
		}
		switch car := car.(type) {
		case sexpr_atom:
			if evaluator, ok := evaluators[car] ; ok {
				if val, err := evaluator(s.cdr, ctx) ; err != nil {
					return nil, err
				} else {
					return val, nil
				}
			}
		}
		// else
		log.Printf("Evaluating a generic 'cons': %q", s)
		if cdr, err := evaluateWithContext(s.cdr, ctx) ; err != nil {
			return nil, err
		} else {
			return mkCons(car, cdr), nil
		}
	default:
		panic(fmt.Sprintf("(Evaluate) Unrecognized Sexpr (type=%T) %v", s, s))
	}
	return Nil, nil
}

// Certain primitive Atom(symbol)s are built-in.  They can't be
// implemented given other things defineable in the SCAM language, so
// we define them here.

var primitiveStrings = []string {
	"-",
	"+",
	"cons",
	"car",
	"cdr",
	"eq?",
	"quote",
	"null?",
	"atom?",
	"define",
	"let",
}

// An evaluator is a decorated S-expression (probably an Atom) that
// can, when it appears in the Car of a Cons, evaluate the expression
// into a new S-expression
type evaluator func(Sexpr, *evaluationContext) (Sexpr, sexpr_error)

// A special kind of error to indicate something about evaluation
type evaluationError struct{
	context string
	message string
}
// A generic, nullable version of the evaluation error.  We need to
// use the real type (evaluationError) in return values but the
// interface (sexpr_error) in method signatures.
type sexpr_error interface{
	Error() string
	Sprint() (string, error)
}
func (e evaluationError) Error() string {
	return fmt.Sprintf("Exception in %s: %s", e.context, e.message)
}
func (e evaluationError) Sprint() (string, error) {
	// Implement this method if we want an evaluation error to look
	// like an S-expression
	return e.Error(), nil
}

// Pre-make all the primitive symbols.  Maybe these need to be their
// own things; we'll see how Evaluate goes
var atomPrimitives = make(map[string]sexpr_atom)
var evaluators = make(map[sexpr_atom]evaluator)
func init() {
	for _, str := range primitiveStrings {
		atomPrimitives[str] = mkAtomSymbol(str)
	}

	evaluators[atomPrimitives["cons"]] = evalCons
	evaluators[atomPrimitives["car"]]  = evalCar
	evaluators[atomPrimitives["cdr"]]  = evalCdr
	evaluators[atomPrimitives["eq?"]]  = evalEqQ
	evaluators[atomPrimitives["+"]]    = evalPlus
	evaluators[atomPrimitives["quote"]] = evalQuote

	evaluators[atomPrimitives["null?"]] = evalNullQ
	evaluators[atomPrimitives["atom?"]] = evalAtomQ
	evaluators[atomPrimitives["define"]] = evalDefine
	evaluators[atomPrimitives["let"]] = evalLet
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
		log.Printf("DEFINE %q <-- %s", key, val)
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
		log.Println("Create binding from", b)
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
