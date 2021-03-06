package sexpr

import (
	"strconv"
	"fmt"
	"errors"
	"math"
)

// Functions are first class objects.  The file defines operations
// with functions and some with just macros.

type applicator func([]sexpr_general) (sexpr_general, sexpr_error)
type func_expr struct{
	definition string
	// A function is handed its arguments pre-evaluated
	apply applicator
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
	return func (ignore sexpr_general, i2 *evaluationContext) (sexpr_general, sexpr_error) {
		return nil, evaluationError{s, "is not yet implemented"}
	}
}

var primitiveMacros = map[string]evaluator {
	"quote":  evalQuote,
	"define": evalDefine,
	"let":    evalLet,
	"lambda": evalLambda,
	"and":    mkLazyReduce("and", atomConstantTrue, func(acc sexpr_general, val sexpr_general) (sexpr_general, bool) {
		if isFalsey(val) {
			return atomConstantFalse, true
		} else {
			return acc, false
		}
	}),
	"or":     mkLazyReduce("or", atomConstantFalse, func(acc sexpr_general, val sexpr_general) (sexpr_general, bool) {
		if !isFalsey(val) {
			return atomConstantTrue, true
		} else {
			return acc, false
		}
	}),
	"if":     mkTodoEvaluator("if"),
	"cond":   evalCond,
}

func mkTodoApplicator(s string) applicator {
	return func(ignore []sexpr_general) (sexpr_general, sexpr_error) {
		return nil, evaluationError{s, "is not yet implemented"}
	}
}

var primitiveFunctions = map[string]applicator {
	"+":      mkArithmeticReduce("+", zeroIntOrFloat, reducePlus),
	"*":      mkArithmeticReduce("*", oneIntOrFloat, reduceTimes),
	"expt":   mkNaryFn("expt", 2, fnExponent),
	"-":      mkNaryFn("-", 2, fnMinus),
	"/":      mkNaryFn("/", 2, fnDivide),
	"cons":   mkNaryFn("cons", 2, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		return mkCons(args[0], args[1]), nil
	}),
	"car":    mkConsSelector("car", func (c sexpr_cons) sexpr_general { return c.car }),
	"cdr":    mkConsSelector("cdr", func (c sexpr_cons) sexpr_general { return c.cdr }),
	"=":      mkNaryFn("=", 2, fnEqualNumber),
	"eq?":    mkNaryFn("eq?", 2, fnEqualAtom),
	"null?":  mkNaryFn("null?", 1, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		if args[0] == atomConstantNil {
			return atomConstantTrue, nil
		} else {
			return atomConstantFalse, nil
		}
	}),
	"pair?":  mkNaryFn("pair?", 1, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		switch args[0].(type) {
		case sexpr_cons:
			return atomConstantTrue, nil
		}
		// else
		return atomConstantFalse, nil
	}),
	"not":    mkNaryFn("not", 1, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		if args[0] == atomConstantTrue {
			return atomConstantFalse, nil
		} else {
			return atomConstantTrue, nil
		}
	}),
	"zero?":  mkNaryFn("pair?", 1, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		if args[0] == atomConstantZero {
			return atomConstantTrue, nil
		} else {
			return atomConstantFalse, nil
		}
	}),
	"number?":  mkNaryFn("pair?", 1, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		switch a := args[0].(type) {
		case sexpr_atom:
			if a.typ == atomNumber {
				return atomConstantTrue, nil
			}
		}
		// else
		return atomConstantFalse, nil
	}),
}
/////
// Helpers
/////

// mkNaryFn makes an n-ary "normal" function, one that operates on
// its arguments after they've been evaluated.
func mkNaryFn(name string, n int, fn func([]sexpr_general) (sexpr_general, sexpr_error)) applicator {
	return func(args []sexpr_general) (sexpr_general, sexpr_error) {
		if len(args) != n {
			msg := fmt.Sprintf("Expected %d arguments, got %d", n, len(args))
			return nil, evaluationError{name, msg}
		}
		ans, err := fn(args)
		if err != nil {
			// divide-by-zero, for instance
			return nil, evaluationError{name, err.Error()}
		}
		return ans, nil
	}
}

func mkConsSelector(name string, sel func(sexpr_cons) sexpr_general) applicator {
	return mkNaryFn(name, 1, func(args []sexpr_general) (sexpr_general, sexpr_error) {
		switch first := args[0].(type) {
		case sexpr_cons: return sel(first), nil
		default:
			return nil, evaluationError{
				name,
				fmt.Sprintf("%s is not a pair", first),
			}
		}
	})
}


func mkLazyReduce(
	name string,
	acc sexpr_general,
	reducer func(acc sexpr_general, val sexpr_general) (sexpr_general, bool),
) evaluator {
	return func(lst sexpr_general, ctx *evaluationContext) (sexpr_general, sexpr_error) {
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

var (
	zeroIntOrFloat = intOrFloat{0, 0.0, true}
	oneIntOrFloat  = intOrFloat{1, 1.0, true}
)

func parseIntOrFloat(s sexpr_general) (*intOrFloat, error) {
	var numberString string
	switch s := s.(type) {
	case sexpr_atom:
		if s.typ != atomNumber {
			msg := fmt.Sprintf("Atom %q is not a number", s)
			return nil, errors.New(msg)
		} else {
			numberString = s.name
		}
	default:
		msg := fmt.Sprintf("%q is not a number", s)
		return nil, errors.New(msg)
	}
	// else, numberString is the thing to parse
	ival, err := strconv.ParseInt(numberString, 10, 64)
	if err == nil {
		// It's an integer
		return &intOrFloat{ival, float64(ival), true}, nil
	}
	// else
	fval, err := strconv.ParseFloat(numberString, 64)
	if err == nil {
		// It's a float; the integer part is irrelevant
		return &intOrFloat{0, fval, false}, nil
	} else {
		return nil, err
	}
}

func (i intOrFloat) String() string {
	if i.isInt {
		return fmt.Sprintf("%d", i.asint)
	} else {
		return fmt.Sprintf("%f", i.asfloat)
	}
}
func (i intOrFloat) sexprize() sexpr_general {
	return mkAtomNumber(fmt.Sprintf("%s", i))
}

// add m to n, destructively modifying n.  Like +=
func (n *intOrFloat) increaseBy(m intOrFloat) {
	n.asint   += m.asint
	n.asfloat += m.asfloat
	if n.isInt {
		n.isInt = m.isInt
	}
}
// multiply m to n, destructively modifying n.  Like *=
func (n *intOrFloat) multiplyBy(m intOrFloat) {
	n.asint   *= m.asint
	n.asfloat *= m.asfloat
	if n.isInt {
		n.isInt = m.isInt
	}
}
// subtract m from n, destructively modifying n.  Like -=
func (n *intOrFloat) decreaseBy(m intOrFloat) {
	n.asint   -= m.asint
	n.asfloat -= m.asfloat
	if n.isInt {
		n.isInt = m.isInt
	}
}
// divide m into n, destructively modifying n.  Like /=
func (n *intOrFloat) divideBy(m intOrFloat) error {
	if (m.isInt && m.asint == 0) || m.asfloat == 0 {
		return errors.New("Divide by zero")
	}
	if n.isInt && m.isInt && n.asint % m.asint == 0 {
		// Integer division is still possible
		n.asint /= m.asint
	} else {
		n.isInt = false
		n.asfloat /= m.asfloat
	}
	return nil
}

// raise n to the power m, desructively modifying n
func (n *intOrFloat) toPower(m intOrFloat) {
	if m.isInt {
		intOrig := n.asint
		if m.asint == 0 {
			// TODO:  Assert n != 0
			n.asint = 1
			n.asfloat = 1
			n.isInt = true
			return
		} else if m.asint > 0 {
			for i := int64(1) ; i < m.asint ; i++ {
				n.asint *= intOrig
			}
		} else {
			// Fraction...
			// TODO:  Assert n != 0
			n.asfloat = 1/n.asfloat
			for i := int64(1) ; i < -m.asint ; i++ {
				n.asfloat /= float64(intOrig)
			}
			n.isInt = false
		}
	}
	n.asfloat = math.Pow(n.asfloat, m.asfloat)
	if n.isInt {
		n.isInt = m.isInt // Bleh... m is 1/2??
	}
}

type arithmeticReducerFunction func(*intOrFloat, intOrFloat)

func mkArithmeticReduce(
	name string,
	starter intOrFloat,
	reducer arithmeticReducerFunction,
) applicator {
	return func(args []sexpr_general) (sexpr_general, sexpr_error) {
		acc := starter // makes a copy!
		for _, val := range args {
			iorf, err := parseIntOrFloat(val)
			if err != nil {
				return nil, evaluationError{name, err.Error()}
			}
			// else, destructively modify the accumulator
			reducer(&acc, *iorf)
		}
		return acc.sexprize(), nil
	}
}

func reducePlus(acc *intOrFloat, addend intOrFloat) {
	acc.increaseBy(addend)
}
func reduceTimes(acc *intOrFloat, multiplicand intOrFloat) {
	acc.multiplyBy(multiplicand)
}

func fnMinus(args []sexpr_general) (sexpr_general, sexpr_error) {
	minuend, err := parseIntOrFloat(args[0])
	if err != nil {
		return nil, evaluationError{"-", err.Error()}
	}
	subtrahend, err := parseIntOrFloat(args[1])
	if err != nil {
		return nil, evaluationError{"-", err.Error()}
	}
	// else
	minuend.decreaseBy(*subtrahend)
	return minuend.sexprize(), nil
}
func fnDivide(args []sexpr_general) (sexpr_general, sexpr_error) {
	divisor, err := parseIntOrFloat(args[0])
	if err != nil {
		return nil, evaluationError{"/", err.Error()}
	}
	dividend, err := parseIntOrFloat(args[1])
	if err != nil {
		return nil, evaluationError{"/", err.Error()}
	}
	// else
	err = divisor.divideBy(*dividend)
	if err != nil {
		// Divide by zero
		return nil, evaluationError{"/", err.Error()}
	}
	return divisor.sexprize(), nil
}
func fnExponent(args []sexpr_general) (sexpr_general, sexpr_error) {
	base, err := parseIntOrFloat(args[0])
	if err != nil {
		return nil, evaluationError{"-", err.Error()}
	}
	exponent, err := parseIntOrFloat(args[1])
	if err != nil {
		return nil, evaluationError{"-", err.Error()}
	}
	// else
	base.toPower(*exponent)
	return base.sexprize(), nil
}

func mkEqualAtomChecker(typ atomType) applicator {
	return func(args []sexpr_general) (sexpr_general, sexpr_error) {
		var atoms []sexpr_atom
		for _, arg := range args {
			switch arg := arg.(type) {
			case sexpr_atom:
				atoms = append(atoms, arg)
			default:
				return atomConstantFalse, nil
			}
		}
		// reaching here, we have atoms
		// Assume at least one; n-ary 2 (probably)
		if atoms[0].typ != typ {
			return atomConstantFalse, nil
		}
		// now check all the rest are the same
		for i := 0 ; 1+i < len(atoms) ; i++ {
			if atoms[1+i] != atoms[i] {
				return atomConstantFalse, nil
			}
		}
		return atomConstantTrue, nil
	}
}
var fnEqualNumber = mkEqualAtomChecker(atomNumber)
var fnEqualSymbol = mkEqualAtomChecker(atomSymbol)
func fnEqualAtom(args []sexpr_general) (sexpr_general, sexpr_error) {
	if args[0] == atomConstantNil {
		for i := 1 ; i < len(args) ; i++ {
			if args[i] != atomConstantNil {
				return atomConstantFalse, nil
			}
		}
		return atomConstantTrue, nil
	}
	// else
	return fnEqualSymbol(args)
}

// evalQuote is a macro; it does not evaluate all its arguments
func evalQuote(lst sexpr_general, ctx *evaluationContext) (sexpr_general, sexpr_error) {
	args, err := unconsifyN(lst, 1)
	if err != nil {
		return nil, evaluationError{"quote", err.Error()}
	}
	// else
	// return the first argument, unevaluated
	return args[0], nil
}

// evalDefine is a macro; it does not evaluate all its arguments
func evalDefine(lst sexpr_general, ctx *evaluationContext) (sexpr_general, sexpr_error) {
	args, err := unconsifyN(lst, 2)
	if err != nil {
		return nil, evaluationError{"define", err.Error()}
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
		return atomConstantNil, nil
		// TODO:  "nil" isn't "Nil"; define has no value
		// return nil, nil
	default:
		return nil, evaluationError{
			"define",
			fmt.Sprintf("Cannot bind non-atom %q", key),
		}
	}
}

func evalLet(lst sexpr_general, ctx *evaluationContext) (sexpr_general, sexpr_error) {
	args, err := unconsifyN(lst, 2)
	if err != nil {
		return nil, evaluationError{"let", err.Error()}
	}
	// else
	bindings, err := unconsify(args[0])
	if err != nil {
		return nil, evaluationError{"let(args)", err.Error()}
	}

	newCtx := evaluationContext{make(symbolTable), ctx}
	for _, b := range bindings {
		// log.Println("Create binding from", b)
		kv, err := unconsifyN(b, 2)
		if err != nil {
			return nil, evaluationError{"let(binding)", err.Error()}
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
	return evaluateWithContext(args[1], &newCtx)
}

func evalLambda(lst sexpr_general, ctx *evaluationContext) (sexpr_general, sexpr_error) {
	args, err := unconsifyN(lst, 2) // eventually "many"
	if err != nil {
		return nil, evaluationError{"lambda", err.Error()}
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
	apply := func(args []sexpr_general) (sexpr_general, sexpr_error) {
		if len(bound) != len(args) {
			return nil, evaluationError{
				fmt.Sprintf("%s", definition),
				fmt.Sprintf("Evaluation with %d arguments, expected %d", len(args), len(bound)),
			}
		}
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
func evalCond(lst sexpr_general, ctx *evaluationContext) (s sexpr_general, serr sexpr_error) {
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
		if test[0] == atomConstantElse {
			return evaluateWithContext(test[1], ctx)
		}
		// else
		if predicate, serr := evaluateWithContext(test[0], ctx) ; serr != nil {
			return nil, serr
		} else if !isFalsey(predicate) {
			return evaluateWithContext(test[1], ctx)
		}
	}
	return atomConstantNil, nil // or nil, nil, if we get "define" doing that.
}
