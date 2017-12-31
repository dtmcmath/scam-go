package sexpr

import (
	"errors"
	"fmt"
)

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

func unconsify(list Sexpr) ([]Sexpr, error) {
	var ans []Sexpr
	for idx, lst := 0, list ; lst != Nil ; idx++ {
		switch l := lst.(type) {
		case sexpr_atom:
			errmsg := fmt.Sprintf(
				"Unexpected atom in position %d of %s",
				idx, list,
			)
			return nil, errors.New(errmsg)
		case sexpr_cons:
			ans = append(ans, l.car)
			lst = l.cdr
		default:
			panic(fmt.Sprintf("Unrecognized S-expression %v", l))
		}
	}
	return ans, nil
}
