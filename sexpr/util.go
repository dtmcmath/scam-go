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
func consify(slist []sexpr_general) sexpr_general {
	if len(slist) == 0 {
		return atomConstantNil
	}
	// else
	return mkCons(slist[0], consify(slist[1:]))
}

func unconsify(list sexpr_general) ([]sexpr_general, error) {
	var ans []sexpr_general
	for idx, lst := 0, list ; lst != atomConstantNil ; idx++ {
		switch l := lst.(type) {
		case sexpr_atom:
			errmsg := fmt.Sprintf(
				"Unexpected atom %q in position %d of %s",
				l, 1+idx, list,
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
// Like unconsify, but throws an error if the resulting list is not of
// length n.
func unconsifyN(list sexpr_general, n int) ([]sexpr_general, error) {
	if ans, err := unconsify(list) ; err != nil {
		return nil, err
	} else if len(ans) != n {
		plural := ""
		if n > 1 {
			plural = "s"
		}
		msg := fmt.Sprintf("Expected %d argument%s, got %d",
			n, plural, len(ans),
		)
		return nil, errors.New(msg)
	} else {
		return ans, nil
	}
}

// Test for equality (not eq?-ness) of expressions.  For everything
// except Cons-es, it's just identity in the normal Go-sense.  For
// Cons cells, we need car and cdr to be equal, but not serial number.
func equalSexpr(a sexpr_general, b sexpr_general) bool {
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

func deepEqualSexpr(a []sexpr_general, b []sexpr_general) bool {
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

// isFalsey says whether "if" should skip it.  Only atomConstantFalse and atomConstantNil are falsey.
func isFalsey(s sexpr_general) bool { return s == atomConstantNil || s == atomConstantFalse }
