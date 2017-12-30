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

// Pre-make all the primitive symbols.  Maybe these need to be their
// own things; we'll see how Evaluate goes
var atomPrimitives = make(map[itemType]sexpr_atom)
func init() {
	for typ, str := range primitives {
		atomPrimitives[typ] = mkAtomSymbol(str)
	}
}

type sexpr_cons struct {
	car Sexpr
	cdr Sexpr
}

// A Sexpr includes sexpr_atom and sexpr_cons.  It's a discriminated union
type Sexpr interface{}

func (c sexpr_cons) String() string {
	// TODO:  This will barf if c itself has a loop
	return fmt.Sprintf("Cons(%s, %s)", c.car, c.cdr)
}

// consify takes a list of S-expressions and returns a single
// S-expression that is the List (sexpr_cons's) represented by them.
func consify(slist []Sexpr) Sexpr {
	if len(slist) == 0 {
		return Nil
	}
	// else
	return sexpr_cons{slist[0], consify(slist[1:])}
}
