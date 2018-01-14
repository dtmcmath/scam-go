package sexpr

import (
	"errors"
	"fmt"
)

// A symbolTable is used to look up values during evaluation.  When a
// new function is pushed into the stack for evaluation, the
// symbolTables get new frames that mask the lower ones.

type symbolTable map[sexpr_atom]sexpr_general

// evaluationContext is really (currently) just a stack of symbol
// tables.  It might get more, though.
type evaluationContext struct{
	sym symbolTable
	parent *evaluationContext
}

func (e *evaluationContext) dump() string {
	return e.dump_helper(0)
}
func (e *evaluationContext) dump_helper(depth int) string {
	ans := fmt.Sprintf("== depth %02d ==\n", depth)
	for key, val := range e.sym {
		ans += fmt.Sprintf("val(%s)\t<--\t%s\n", key, val)
	}
	ans += "--------------"
	if e.parent != nil {
		ans += "\n" + e.parent.dump_helper(1+depth)
	}
	return ans
}

func (e *evaluationContext) bind(key sexpr_atom, val sexpr_general) error {
	if key.typ != atomSymbol {
		return errors.New(fmt.Sprintf("Cannot bind non-symbol %q", key))
	}
	// else
	// TODO:  Check further; let's not bind Nil, nor any of
	// the primitives, for instance.
	// So key needs to be a non-primitive symbol.
	// (if we _did_ accidentally redefine null?, the symbol-lookup
	// wouldn't honor the request anyway, but we would create
	// confusion and delay I'm sure)
	e.sym[key] = val
	return nil
}

func (e *evaluationContext) lookup(a sexpr_atom) (s sexpr_general, ok bool) {
	if e == nil {
		return nil, false
	}

	ptr := e
	val, ok := ptr.sym[a]
	for !ok {
		// Check the parent context
		ptr = ptr.parent
		if ptr == nil {
			break
		}
		// else
		val, ok = ptr.sym[a]
	}
	return val, ok
}
