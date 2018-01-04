// An S-expression is either an atom or the cons of two S-expressions.
package sexpr

import (
	"errors"
	"fmt"
	"io"
	// "runtime/debug"
	"strconv"
	//"strings"
	"os"
)

type atomType int
const (
	atomNil atomType = iota
	atomNumber
	atomSymbol
	atomBoolean
)

type sexpr_atom struct {
	typ atomType
	name string
}

func (a sexpr_atom) Sprint() string {
	switch a.typ {
	case atomNil:   return "()"
	case atomBoolean:
		switch a {
		case True:  return "#t"
		case False: return "#f"
		default:
			panic(fmt.Sprintf("The faux boolean atom %+v", a))
		}
	case atomNumber, atomSymbol:
		return a.name
	default:
		msg := fmt.Sprintf("Unprintable atom of type %q: %s", a.typ, a)
		panic(msg)
	}
}

func (a sexpr_atom) String() string {
	switch a.typ {
	case atomNil: return "Nil"
	case atomBoolean:
		switch a {
		case True: return "#t"
		case False: return "#f"
		default:
			panic(fmt.Sprintf("The faux boolean atom %+v", a))
		}
	case atomNumber:
		// Try to keep things integer, if possible
		if i, err := strconv.ParseInt(a.name, 10, 64) ; err == nil {
			return fmt.Sprintf("%d", i)
		} else {
		}
		if f, err := strconv.ParseFloat(a.name, 64) ; err == nil {
			return fmt.Sprintf("%f", f)
		}
		// else
		msg := fmt.Sprintf(
			"Unprintable non-number %q posing as a number",
			a.name,
		)
		panic(msg)
	case atomSymbol: return fmt.Sprintf("Sym(%s)", a.name)
	default:
		panic(fmt.Sprintf("No way: atom %v", a))
	}
}

func (a sexpr_atom) evaluate(ctx *evaluationContext) (Sexpr, sexpr_error) {
	switch a.typ {
	case atomSymbol:
		// The primitives are just themselves (for now)
		if _, ok := atomPrimitives[a.name] ; ok {
			return a, nil
		}
		// else
		if val, ok := ctx.lookup(a) ; !ok {
			return nil, evaluationError{
				"lookup",
				fmt.Sprintf("Variable %s is not bound", a),
			}
		} else {
			return val, nil
		}
	default:
		// Hrm.  This is probbably just right.
		return a, nil
	}
}

var (
	// These are really a constant, but we call them variables.
	// Please don't try to change them.
	Nil sexpr_atom = sexpr_atom{atomNil, "nil"}
	True sexpr_atom = sexpr_atom{atomBoolean, "t"}
	False sexpr_atom = sexpr_atom{atomBoolean, "f"}
	Quote sexpr_atom = mkAtomSymbol("quote")
)
// TODO:  Different string representations of the same number are
// different atoms; are they comparable?

var atomNumberPool = make(map[string]sexpr_atom)
var atomSymbolPool = make(map[string]sexpr_atom)

func atomFactory(t atomType, pool map[string]sexpr_atom) func(string) sexpr_atom {
	return func (s string) sexpr_atom {
		atom, ok := pool[s]
		if !ok {
			atom = sexpr_atom{t, s}
			pool[s] = atom
		}
		return atom
	}
}

var mkAtomSymbol = atomFactory(atomSymbol, atomSymbolPool)
var mkAtomNumber = atomFactory(atomNumber, atomNumberPool)

var currentConsNumber int64
type sexpr_cons struct {
	car Sexpr
	cdr Sexpr
	serialNumber int64 // Keeps separate objects different
}

func mkCons(car Sexpr, cdr Sexpr) sexpr_cons {
	defer func() {currentConsNumber += 1}()
	return sexpr_cons{car, cdr, currentConsNumber}
}
// mkList is a helper method to replace
//
//   mkCons(a, mkCons(b, mkCons(c, Nil)))
//
// with just
//
//   mkList(a, b, c)
func mkList(s ...Sexpr) Sexpr { return consify(s) }

func getCar(s Sexpr) (Sexpr, error) {
	switch s := s.(type) {
	case sexpr_atom: return nil, errors.New("Cannot car an Atom")
	case sexpr_cons: return s.car, nil
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", s))
	}
}

func getCdr(s Sexpr) (Sexpr, error) {
	switch s := s.(type) {
	case sexpr_atom: return nil, errors.New("Cannot cdr an Atom")
	case sexpr_cons: return s.cdr, nil
	default:
		panic(fmt.Sprintf("Unrecognized Sexpr %v", s))
	}
}

func getCadr(s Sexpr) (Sexpr, error) {
	if cdr, err := getCdr(s) ; err != nil {
		return nil, err
	} else {
		return getCar(cdr)
	}
}
func getCddr(s Sexpr) (Sexpr, error) {
	if cdr, err := getCdr(s) ; err != nil {
		return nil, err
	} else {
		return getCdr(cdr)
	}
}

func (c sexpr_cons) Sprint() string {
	str := "("

	for ptr := c ; ; {
		str += ptr.car.Sprint()

		switch cdr := ptr.cdr.(type) {
		case sexpr_atom:
			if cdr == Nil {
				return str + ")"
			} else {
				return str + fmt.Sprintf(" . %s)", cdr.Sprint())
			}
		case sexpr_cons:
			str += " "
			ptr = cdr
			// and loop around agian
		default:
			msg := fmt.Sprintf("Unprintable CDR %q", cdr)
			panic(msg)
		}
	}
	return str
}

func (c sexpr_cons) String() string {
	// TODO:  This will barf if c itself has a loop
	return fmt.Sprintf("Cons(%s, %s)", c.car, c.cdr)
}

// A Sexpr includes sexpr_atom and sexpr_cons.  It's a discriminated union
type Sexpr interface{
	Sprint() string
}

func Sprint(s Sexpr) string {
	return s.Sprint()
}

// Fprint writes a pretty version of the S-expression to the named
// writer.  It returns the number of bytes written and any write-error
// encountered.
func Fprint(out io.Writer, s Sexpr) (n int, err error) {
	return fmt.Fprint(out, Sprint(s))
}

func Print(s Sexpr) (n int, err error) {
	return Fprint(os.Stdout, s)
}
