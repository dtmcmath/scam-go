package sexpr

import (
	"fmt"
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
}

/////
// Definitions of evaluators
/////
func evalCons(cdr Sexpr) Sexpr {
	switch cdr := cdr.(type) {
	case sexpr_atom: return sexpr_error{"cons", "Cannot 'cons' an Atom"}
	case sexpr_error: return cdr
	case sexpr_cons:
		first := cdr.car
		second, err := getCar(cdr.cdr)
		if err != nil {
			return sexpr_error{
				"cons",
				fmt.Sprintf("too few arguments (%s)", err.Error()),
			}
		}
		// else
		return mkCons(first, second)
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", cdr))
	}
}
