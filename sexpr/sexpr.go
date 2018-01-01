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

func (a sexpr_atom) Sprint() (string, error) {
	switch a.typ {
	case atomNil:   return "()", nil
	case atomBoolean:
		switch a {
		case True:  return "#t", nil
		case False: return "#f", nil
		default:
			panic(fmt.Sprintf("The faux boolean atom %+v", a))
		}
	case atomNumber, atomSymbol:
		return a.name, nil
	default:
		msg := fmt.Sprintf("Unprintable atom of type %q: %s", a.typ, a)
		return "", errors.New(msg)
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
			panic(fmt.Sprintf("The false boolean atom %+v", a))
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

func (a sexpr_atom) evaluate() Sexpr {
	// TODO:  Actual lookup in the symbol table
	switch a.typ {
	case atomSymbol:
		// log.Printf("Pretend we looked up the value for %s", a)
		return a
	default:
		return a
	}
}

var (
	// These are really a constant, but we call them variables.
	// Please don't try to change them.
	Nil sexpr_atom = sexpr_atom{atomNil, "nil"}
	True sexpr_atom = sexpr_atom{atomBoolean, "t"}
	False sexpr_atom = sexpr_atom{atomBoolean, "f"}
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

func (c sexpr_cons) Sprint() (string, error) {
	str := "("
	switch car := c.car.(type) {
	case sexpr_atom:
		if scar, err := car.Sprint() ; err != nil {
			msg := fmt.Sprintf("Unprintable CAR %q: %s", car, err.Error())
			return "", errors.New(msg)
		} else {
			str += scar
		}
	case sexpr_cons:
		if scar, err := car.Sprint() ; err != nil {
			msg := fmt.Sprintf("Unprintable CAR %q: %s", car, err.Error())
			return "", errors.New(msg)
		} else {
			str += scar
		}
	default:
		msg := fmt.Sprintf("Unprintable CAR %q", car)
		return "", errors.New(msg)
	}

	if c.cdr == Nil {
		str += ")"
	} else {
		switch cdr := c.cdr.(type) {
		case sexpr_atom:
			if scdr, err := cdr.Sprint() ; err != nil {
				msg := fmt.Sprintf("Unprintable CDR %q: %s", cdr, err.Error())
				return "", errors.New(msg)
			} else {
				str += fmt.Sprintf(" . %s)", scdr)
			}
		case sexpr_cons:
			if scdr, err := cdr.Sprint() ; err != nil {
				msg := fmt.Sprintf("Unprintable CDR %q: %s", cdr, err.Error())
				return "", errors.New(msg)
			} else {
				str += fmt.Sprintf(" %s)", scdr)
			}
		default:
			msg := fmt.Sprintf("Unprintable CDR %q", cdr)
			return "", errors.New(msg)
		}
	}
	return str, nil
}

func (c sexpr_cons) String() string {
	// TODO:  This will barf if c itself has a loop
	return fmt.Sprintf("Cons(%s, %s)", c.car, c.cdr)
}

// A special kind of S-expression is the "error".  I think we're going
// to need it, but I haven't figured out quite where yet.
type sexpr_error struct {
	context string
	message string
}

// A Sexpr includes sexpr_atom and sexpr_cons.  It's a discriminated union
type Sexpr interface{}

func Sprint(s Sexpr) (string, error) {
	switch s := s.(type) {
	case sexpr_atom:
		return s.Sprint()
	case sexpr_cons:
		return s.Sprint()
	default:
		msg := fmt.Sprintf("Unprintable S-expression %T: %s", s, s)
		return "", errors.New(msg)
	}
}

// Fprint writes a pretty version of the S-expression to the named
// writer.  It returns the number of bytes written and any write-error
// encountered.
func Fprint(out io.Writer, s Sexpr) (n int, err error) {
	if str, err := Sprint(s) ; err != nil {
		return 0, err
	} else {
		return fmt.Fprint(out, str)
	}
}

func Print(s Sexpr) (n int, err error) {
	return Fprint(os.Stdout, s)
}
