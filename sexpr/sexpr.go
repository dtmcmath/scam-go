// An S-expression is either an atom or the cons of two S-expressions.
package sexpr

import (
	"fmt"
	"log"
	"errors"
)

type atomType int
const (
	atomNil atomType = iota
	atomNumber
	atomSymbol
	atomQuotedSymbol
	atomBoolean
)

type sexpr_atom struct {
	typ atomType
	name string
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
	case atomNumber: return fmt.Sprintf("N(%s)", a.name)
	case atomSymbol: return fmt.Sprintf("Sym(%s)", a.name)
	case atomQuotedSymbol: return fmt.Sprintf("Quote(%s)", a.name)
	default:
		panic(fmt.Sprintf("No way: atom %v", a))
	}
}

func (a sexpr_atom) evaluate() Sexpr {
	// TODO:  Actual lookup in the symbol table
	switch a.typ {
	case atomSymbol:
		log.Printf("Pretend we looked up the value for %s", a)
		return a
	case atomQuotedSymbol:
		return mkAtomSymbol(a.name)
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
var atomQuotedPool = make(map[string]sexpr_atom)

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
var mkAtomQuoted = atomFactory(atomQuotedSymbol, atomQuotedPool)
var mkAtomNumber = atomFactory(atomNumber, atomNumberPool)

type sexpr_cons struct {
	car Sexpr
	cdr Sexpr
}

func mkCons(car Sexpr, cdr Sexpr) sexpr_cons {
	return sexpr_cons{car, cdr}
}

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
