package sexpr

// Utility/helper functions

// consify takes a list of S-expressions and returns a single
// S-expression that is the List (sexpr_cons's) represented by them.
func consify(slist []Sexpr) Sexpr {
	if len(slist) == 0 {
		return Nil
	}
	// else
	return sexpr_cons{slist[0], consify(slist[1:])}
}
