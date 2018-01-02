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
	ans := "==========\n"
	for key, val := range e.sym {
		ans += fmt.Sprintf("val(%s)\t<--\t%s\n", key, val)
	}
	ans += "----------\n"
	return ans
}

type evaluationStack struct{
	head *evaluationContext
}

func (e *evaluationStack) dump() string {
	ans := ""
	depth := 0
	for head := e.head ; head != nil ; {
		ans += fmt.Sprintf("Depth %d:\n", depth)
		ans += head.dump()
		depth += 1
		head = head.parent
	}
	return ans
}

func (e *evaluationStack) lookup(a sexpr_atom) (s Sexpr, ok bool) {
	if e.head == nil {
		return nil, false
	}
	// else
	head := e.head
	val, ok := head.sym[a]
	for !ok {
		// Check the parent context
		head = head.parent
		if head == nil {
			break
		}
		// else
		val, ok = head.sym[a]
	}
	return val, ok
}

func (e *evaluationStack) push(sym symbolTable) {
	newHead := &evaluationContext{sym, e.head}
	e.head = newHead
}

func (e *evaluationStack) pop() error {
	if e.head == nil {
		return emptyStackError
	}
	// else
	e.head = e.head.parent
	return nil
}
