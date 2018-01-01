package sexpr

import (
	"fmt"
	"strconv"
	"log"
)

func Evaluate(s Sexpr) Sexpr {
	switch s := s.(type) {
	case sexpr_atom: return s.evaluate()
	case sexpr_error: return s
	case sexpr_cons:
		car := Evaluate(s.car)
		switch car := car.(type) {
		case sexpr_atom:
			if evaluator, ok := evaluators[car] ; ok {
				return evaluator(s.cdr)
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
type evaluator func(Sexpr) Sexpr

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
// Definitions of evaluators
/////
func evalCons(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"cons", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 2 {
		return sexpr_error{"cons", fmt.Sprintf("Expected 2 arguments, got %d", len(args))}
	}
	// else
	return mkCons(Evaluate(args[0]), Evaluate(args[1]))
}

func evalCar(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"car", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 1 {
		return sexpr_error{"car", fmt.Sprintf("Expected 1 argument, got %d", len(args))}
	}
	// else
	first := Evaluate(args[0])
	switch first := first.(type) {
	case sexpr_atom:  return sexpr_error{"car", "Cannot 'car' an Atom"}
	case sexpr_error: return first
	case sexpr_cons:  return first.car
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", first))
	}
}

func evalCdr(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"cdr", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 1 {
		return sexpr_error{"cdr", fmt.Sprintf("Expected 1 argument, got %d", len(args))}
	}
	// else
	first := Evaluate(args[0])
	switch first := first.(type) {
	case sexpr_atom:  return sexpr_error{"cdr", "Cannot 'cdr' an Atom"}
	case sexpr_error: return first
	case sexpr_cons:  return first.cdr
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", first))
	}
}

func evalEqQ(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"eq?", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 2 {
		return sexpr_error{"eq?", fmt.Sprintf("Expected 2 arguments, got %d", len(args))}
	}
	// else
	if Evaluate(args[0]) == Evaluate(args[1]) {
		return True
	} else {
		return False
	}
}

func evalPlus(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"+", fmt.Sprintf("Strange arguments %q", lst)}
	}
	var acc int64
	for _, summand := range args {
		switch summand := Evaluate(summand).(type) {
		case sexpr_atom:
			if summand.typ == atomNumber {
				val, err := strconv.ParseInt(summand.name, 10, 64)
				if err != nil {
					return sexpr_error{"+",
						fmt.Sprintf("Unexpected non-numeric number summand '%s'", summand),
					}
				}
				// else
				acc += val
			} else {
				return sexpr_error{"+",
					fmt.Sprintf("Unexpected non-numeric summand '%s'", summand),
				}
			}
		case sexpr_error: return summand
		default: return sexpr_error{"+", fmt.Sprintf("Unexpected summand '%s'", summand)}
		}
	}

	return mkAtomNumber(fmt.Sprintf("%d", acc))
}

func evalQuote(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"quote", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 1 {
		return sexpr_error{"eq?", fmt.Sprintf("Expected 1 argument, got %d", len(args))}
	}
	// else
	// return the first argument, unevaluated
	return args[0]
}

func evalNullQ(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"null?", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 1 {
		return sexpr_error{"null?", fmt.Sprintf("Expected 1 argument, got %d", len(args))}
	}
	// else
	if Evaluate(args[0]) == Nil {
		return True
	} else {
		return False
	}
}

func evalAtomQ(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"atom?", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 1 {
		return sexpr_error{"atom?", fmt.Sprintf("Expected 1 argument, got %d", len(args))}
	}
	// else
	val := Evaluate(args[0])
	switch val := val.(type) {
	case sexpr_atom:
		if val != Nil {
			return True
		} else {
			return False
		}
	default:         return False
	}
}

func evalDefine(lst Sexpr) Sexpr {
	args, err := unconsify(lst)
	if err != nil {
		return sexpr_error{"define", fmt.Sprintf("Strange arguments %q", lst)}
	} else if len(args) != 2 {
		return sexpr_error{"define?", fmt.Sprintf("Expected 2 arguments, got %d", len(args))}
	}
	// else
	key := args[0]
	switch key := key.(type) {
	case sexpr_atom:
		// TODO:  Check further; let's not redefine Nil, for instance
		val := Evaluate(args[1])
		log.Printf("DEFINE %q <-- %s", key, val)
		return Nil
	default:
		panic(fmt.Sprintf("Cannot define a non-symbol: %q", key))
	}
}
