package sexpr

import (
	"fmt"
)

// A symbolTable is used to look up values during evaluation.  When a
// new function is pushed into the stack for evaluation, the
// symbolTables get new frames that mask the lower ones.

type symbolTable map[sexpr_atom]Sexpr

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

func (e *evaluationContext) lookup(a sexpr_atom) (s Sexpr, ok bool) {
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
