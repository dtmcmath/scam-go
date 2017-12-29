// An S-expression is either an atom or the cons of two S-expressions.
package sexpr

type Atom struct {
	Name string
}

type Cons struct {
	car Sexpr
	cdr Sexpr
}

// A Sexpr includes Atom and Cons.  It's a discriminated union
type Sexpr interface{}


	
