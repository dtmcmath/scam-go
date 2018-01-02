package sexpr

import (
	"fmt"
	"strconv"
	"log"
)

func Evaluate(s Sexpr) Sexpr {
	switch s := s.(type) {
	case sexpr_atom: return s.evaluate()
	case sexpr_cons:
		car := Evaluate(s.car)
		switch car := car.(type) {
		case sexpr_atom:
			if evaluator, ok := evaluators[car] ; ok {
				if val, err := evaluator(s.cdr) ; err != nil {
					// We should really "panic"... but that won't help
					// debugging, so we're permissive (and abuse "atom")
					log.Print(err)
					return mkAtomSymbol(err.Error())
				} else {
					return val
				}
			}
		}
		// else
		return mkCons(car, Evaluate(s.cdr))
	default:
		panic(fmt.Sprintf("(Evaluate) Unrecognized Sexpr (type=%T) %v", s, s))
	}
	return Nil
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
}

// An evaluator is a decorated S-expression (probably an Atom) that
// can, when it appears in the Car of a Cons, evaluate the expression
// into a new S-expression
type evaluator func(Sexpr) (Sexpr, error)

// A special kind of error to indicate something about evaluation
type evaluationError struct{
	context string
	message string
}
func (e evaluationError) Error() string {
	return fmt.Sprintf("Exception in %s: %s)", e.context, e.message)
}
// func (e evaluationError) Sprint() (string, error) {
// 	// Implement this method if we want an evaluation error to look
// 	// like an S-expression
// }

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
}

/////
// Helpers
/////
func requireArgCount(lst Sexpr, context string, required int) ([]Sexpr, error) {
	args, err := unconsify(lst)
	if err != nil {
		return nil, evaluationError{
			context,
			fmt.Sprintf("Strange arguments %q", lst),
		}
	} else if len(args) != required {
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
	} else {
		return args, nil
	}
}
/////
// Definitions of evaluators
/////
func evalCons(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "cons", 2)
	if err != nil {
		return nil, err
	}
	// else
	return mkCons(Evaluate(args[0]), Evaluate(args[1])), nil
}

func evalCar(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "car", 1)
	if err != nil {
		return nil, err
	}
	// else
	first := Evaluate(args[0])
	switch first := first.(type) {
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

func evalCdr(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "cdr", 1)
	if err != nil {
		return nil, err
	}
	// else
	first := Evaluate(args[0])
	switch first := first.(type) {
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

func evalEqQ(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "eq?", 2)
	if err != nil {
		return nil, err
	}
	// else
	if Evaluate(args[0]) == Evaluate(args[1]) {
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

func evalPlus(lst Sexpr) (Sexpr, error) {
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
		switch summand := Evaluate(summand).(type) {
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
func evalQuote(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "quote", 1)
	if err != nil {
		return nil, err
	}
	// else
	// return the first argument, unevaluated
	return args[0], nil
}

func evalNullQ(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "null?", 1)
	if err != nil {
		return nil, err
	}
	// else
	if Evaluate(args[0]) == Nil {
		return True, nil
	} else {
		return False, nil
	}
}

func evalAtomQ(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "atom?", 1)
	if err != nil {
		return nil, err
	}
	// else
	val := Evaluate(args[0])
	switch val := val.(type) {
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
func evalDefine(lst Sexpr) (Sexpr, error) {
	args, err := requireArgCount(lst, "define", 2)
	if err != nil {
		return nil, err
	}
	// else
	key := args[0]
	switch key := key.(type) {
	case sexpr_atom:
		// TODO:  Check further; let's not redefine Nil, for instance
		val := Evaluate(args[1])
		log.Printf("DEFINE %q <-- %s", key, val)
		return Nil, nil
	default:
		panic(fmt.Sprintf("Cannot define a non-symbol: %q", key))
	}
}
