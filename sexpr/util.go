package sexpr

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Utility/helper functions

// consify takes a list of S-expressions and returns a single
// S-expression that is the List (sexpr_cons's) represented by them.
func consify(slist []Sexpr) Sexpr {
	if len(slist) == 0 {
		return Nil
	}
	// else
	return mkCons(slist[0], consify(slist[1:]))
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

// Test for equality (not eq?-ness) of expressions.  For everything
// except Cons-es, it's just identity in the normal Go-sense.  For
// Cons cells, we need car and cdr to be equal, but not serial number.
func equalSexpr(a Sexpr, b Sexpr) bool {
	switch a := a.(type) {
	case sexpr_cons:
		switch b := b.(type) {
		case sexpr_cons:
			return equalSexpr(a.car, b.car) &&
				equalSexpr(a.cdr, b.cdr)
		default:
			return false
		}
	default: return a == b
	}
}

func deepEqualSexpr(a []Sexpr, b []Sexpr) bool {
	for len(a) == len(b) {
		if len(a) == 0 {
			return true
		} else if !equalSexpr(a[0], b[0]) {
			return false
		}
		// else
		a = a[1:]
		b = b[1:]
	}
	return false
}

func mkRuneChannel(in string) <-chan rune {
	ans := make(chan rune)
	go func() {
		rdr := strings.NewReader(in)
		// This is what a while-loop looks like??
		r, _, err := rdr.ReadRune()
		for ; err != io.EOF ; r, _, err = rdr.ReadRune() {
			if err != nil {
				panic("Completely unexpected error " + err.Error())
			}
			ans <- r
		}
		close(ans)
	}()
	return ans
}
