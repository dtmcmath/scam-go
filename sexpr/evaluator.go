package sexpr

import (
	"fmt"
	// "log"
)

var globalEvaluationContext evaluationContext // define always writes here

func init() {
	resetEvaluationContext()
}
// reset the rootSymbolTable.  This is useful for testing!
func resetEvaluationContext() {
	globalEvaluationContext = evaluationContext{
		make(map[sexpr_atom]sexpr_general),
		nil,
	}

	// Pre-make all the primitive symbols.  Maybe these need to be their
	// own things; we'll see how Evaluate goes
	for str, eva := range primitiveMacros {
		globalEvaluationContext.bind(
			mkAtomSymbol(str),
			macro_expr{str, eva},
		)
	}
	for str, eva := range primitiveFunctions {
		globalEvaluationContext.bind(
			mkAtomSymbol(str),
			func_expr{str, eva},
		)
	}
}

func Evaluate(s sexpr_general) sexpr_general {
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

func evaluateWithContext(s sexpr_general, ctx *evaluationContext) (sexpr_general, sexpr_error) {
	switch s := s.(type) {
	case sexpr_atom: return s.evaluate(ctx)
	case sexpr_cons:
		car, err := evaluateWithContext(s.car, ctx)
		if err != nil {
			return nil, err
		}
		switch car := car.(type) {
		case func_expr:
			// Functions evaluate their arguments in the current context
			terms, uerr := unconsify(s.cdr)
			if uerr != nil {
				return nil, evaluationError{"(eval)", uerr.Error()}
			}
			// else
			args := make([]sexpr_general, len(terms))
			var err sexpr_error
			for idx, term := range terms {
				if args[idx], err = evaluateWithContext(term, ctx) ; err != nil {
					return nil, err
				}
			}
			return car.apply(args)
		case macro_expr:
			// Macros might do anything; give it the context
			return car.apply(s.cdr, ctx)
		default:
			msg := fmt.Sprintf("Attempt to apply non-procedure %q", car)
			return nil, evaluationError{"(eval)", msg}
		}
	default:
		panic(fmt.Sprintf("(Evaluate) Unrecognized Sexpr (type=%T) %v", s, s))
	}
	return atomConstantNil, nil
}

// An evaluator is a decorated S-expression (probably an Atom) that
// can, when it appears in the Car of a Cons, evaluate the expression
// into a new S-expression
type evaluator func(sexpr_general, *evaluationContext) (sexpr_general, sexpr_error)

// A special kind of error to indicate something about evaluation
type evaluationError struct{
	context string
	message string
}
func MkEvaluationError(context string, message string) evaluationError {
	return evaluationError{context, message}
}
// A generic, nullable version of the evaluation error.  We need to
// use the real type (evaluationError) in return values but the
// interface (sexpr_error) in method signatures.
type sexpr_error interface{
	Error() string
	Sprint() string
}
func (e evaluationError) Error() string {
	return fmt.Sprintf("Exception in %s: %s", e.context, e.message)
}
func (e evaluationError) Sprint() string {
	// Implement this method if we want an evaluation error to look
	// like an S-expression
	return e.Error()
}
