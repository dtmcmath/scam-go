// An S-expression is either an atom or the cons of two S-expressions.
package sexpr

import (
	"fmt"
)

type atomType int
const (
	atomNil atomType = iota
	atomNumber
	atomSymbol
	atomQuotedSymbol
	atomBoolean
)

type Atom struct {
	typ atomType
	name string
}

func (a Atom) String() string {
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

var (
	// These are really a constant, but we call them variables.
	// Please don't try to change them.
	Nil Atom = Atom{atomNil, "nil"}
	True Atom = Atom{atomBoolean, "t"}
	False Atom = Atom{atomBoolean, "f"}
)
// TODO:  Different string representations of the same number are
// different atoms; are they comparable?

var atomNumberPool = make(map[string]Atom)
var atomSymbolPool = make(map[string]Atom)
var atomQuotedPool = make(map[string]Atom)

func atomFactory(t atomType, pool map[string]Atom) func(string) Atom {
	return func (s string) Atom {
		atom, ok := pool[s]
		if !ok {
			atom = Atom{t, s}
			pool[s] = atom
		}
		return atom
	}
}

var mkAtomSymbol = atomFactory(atomSymbol, atomSymbolPool)
var mkAtomQuoted = atomFactory(atomQuotedSymbol, atomQuotedPool)
var mkAtomNumber = atomFactory(atomNumber, atomNumberPool)

type Cons struct {
	car Sexpr
	cdr Sexpr
}

// A Sexpr includes Atom and Cons.  It's a discriminated union
type Sexpr interface{}

func (c Cons) String() string {
	// TODO:  This will barf if c itself has a loop
	return fmt.Sprintf("Cons(%s, %s)", c.car, c.cdr)
}

// consify takes a list of S-expressions and returns a single
// S-expression that is the List (Cons's) represented by them.
func consify(slist []Sexpr) Sexpr {
	if len(slist) == 0 {
		return Nil
	}
	// else
	return Cons{slist[0], consify(slist[1:])}
}
