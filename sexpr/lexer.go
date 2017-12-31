package sexpr

import (
	"errors"
	"fmt"
	// "log"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType int

// item represents a token returned from the scanner.
type item struct {
    typ itemType  // Type, such as itemNumber.
    val string    // Value, such as "23.2".
}

// itemType identifies the type of lex items.
const (
	itemError itemType = iota // error occurred;
                              // value is text of error
	itemEOF
	itemLparen                // (
	itemRparen                // )
	itemDot                   // .
	itemNumber                // a numeric thing
	itemQuotedSymbol          // 'abc
	itemSymbol                // abc
	itemBoolean               // #t or #f
	itemWhitespace            // ... maybe not needed
	itemQuotationMark         // " not yet implemented
)

func (i item) String() string {
	switch i.typ {
	case itemEOF: return "EOF"
	case itemLparen: return "LPAREN"
	case itemRparen: return "RPAREN"
	case itemDot: return "DOT"
	case itemNumber: return fmt.Sprintf("NUMBER(%s)", i.val)
	case itemQuotedSymbol: return fmt.Sprintf("QSYMBOL(%s)", i.val)
	case itemSymbol: return fmt.Sprintf("SYMBOL(%s)", i.val)
	case itemBoolean: return fmt.Sprintf("BOOL(%s)", i.val)
	case itemWhitespace: return "WHITESPACE"
	case itemQuotationMark: return "QUOTE"
	case itemError: return fmt.Sprintf("ERROR(%s)", i.val)
	default:
		panic(fmt.Sprintf("No way:  token {%v, $v}", i.typ, i.val))
	}
}

const (
	eof rune = utf8.RuneError // Find a better "eof" rune
)

type lexer struct {
    name  string    // used only for error reports.
    input string    // the string being scanned.
    start int       // start position of this item.
    pos   int       // current position in the input.
    width int       // width of last rune read from input.
    items chan item // channel of scanned items.
}

type stateFn func(*lexer) stateFn

// run lexes the input by executing state functions
// until the state is nil.
func (l *lexer) run() {
    for state := lexText; state != nil; {
        state = state(l)
    }
	close(l.items)
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
    l.items <- item{t, l.input[l.start:l.pos]}
    l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
    if l.pos >= len(l.input) {
		// BUG or FEATURE?  If used in "peek", this will make backup
		// impossible.
        l.width = 0
        return eof
    }
    r, l.width =
        utf8.DecodeRuneInString(l.input[l.pos:])
    l.pos += l.width
    return r
}

// consume positions the head (pos) at the end of s
// If s is not the next thing, return an error
func (l *lexer) consume(s string) error {
	for _, r := range s {
		tst := l.next()
		if tst != r {
			e := fmt.Sprintf("Mismatched attempt to consume %q; found %q",
				s, l.input[l.start:l.pos])
			return errors.New(e)
		}
	}
	return nil
}

func (l *lexer) mustConsume(s string) {
	err := l.consume(s)
	if err != nil {
		panic(err)
	}
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
    l.start = l.pos
}
// backup steps back one rune.
// Can be called only once per call of next and not after a peek.
func (l *lexer) backup() {
    l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() (r rune) {
    r = l.next()
    l.backup()
    return r
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	return l.acceptPredicate(func (r rune) bool {
		return strings.IndexRune(valid, r) >= 0
	})
    // if strings.IndexRune(valid, l.next()) >= 0 {
    //     return true
    // }
    // l.backup()
    // return false
}
func (l *lexer) acceptPredicate(preds ...func(rune) bool) bool {
	r := l.next()
	for _, pred := range preds {
		if pred(r) {
			return true
		}
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	l.acceptRunPredicate(func (r rune) bool {
		return strings.IndexRune(valid, r) >= 0
    })
}
func (l *lexer) acceptRunPredicate(preds ...func(rune) bool) {
	for l.acceptPredicate(preds...) {
	}
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
    l.items <- item{
        itemError,
        fmt.Sprintf(format, args...),
    }
    return nil
}

func lex(name, input string) (*lexer, chan item) {
    l := &lexer{
        name:  name,
        input: input,
        items: make(chan item),
    }
    go l.run()  // Concurrently run state machine.
    return l, l.items
}

// // lex creates a new scanner for the input string.
// func lex(name, input string) *lexer {
//     l := &lexer{
//         name:  name,
//         input: input,
//         state: lexText,
//         items: make(chan item, 2), // Two items sufficient.
//     }
//     return l
// }

// // nextItem returns the next item from the input.
// func (l *lexer) nextItem() item {
//     for {
//         select {
//         case item := <-l.items:
//             return item
//         default:
//             l.state = l.state(l)
//         }
//     }
//     panic("not reached")
// }

////
// The state functions
////

// lexText reads any initial whitespace up to the first interesting thing.
func lexText(l *lexer) stateFn {
    loop: for {
		switch r := l.next() ; {
		case r == eof:
			break loop
		case unicode.IsSpace(r):
			l.ignore() // Or emit some whitespace?
			// l.acceptRun(" \n\t")
			// if l.pos > l.start {
			// 	l.emit(itemWhitespace)
			// )
		case r == '(':
			l.emit(itemLparen)
		case r == ')':
			l.emit(itemRparen)
		case r == '+' || r == '-':
			peek := l.peek()
			if unicode.IsSpace(peek) || peek == eof || peek == ')' {
				return lexSymbol
			} else {
				l.backup()
				return lexNumber
			}
		case '0' <= r && r <= '9':
			l.backup()
			return lexNumber
		case r == '\'':
			l.ignore()
			return lexQuotedSymbol
		case unicode.IsLetter(r):
			// No backup; lexSymbol expects to have read one
			return lexSymbol
		case r == '#':
			l.ignore() // Consume
			return lexBoolean
		default:
			return l.errorf("unrecognized '%c' in '%q'", r, l.input[l.start:l.pos])
		}
    }
    // Correctly reached EOF.
    if l.pos > l.start {
        l.emit(itemWhitespace)
    }
    l.emit(itemEOF)  // Useful to make EOF a token.
    return nil       // Stop the run loop.
}

func lexNumber(l *lexer) stateFn {
    // Optional leading sign.
    l.accept("+-")
    // Is it hex?
    digits := "0123456789"
    if l.accept("0") && l.accept("xX") {
        digits = "0123456789abcdefABCDEF"
    }
    l.acceptRun(digits)
    if l.accept(".") {
        l.acceptRun(digits)
    }
    if l.accept("eE") {
        l.accept("+-")
        l.acceptRun("0123456789")
    }
    // Next thing mustn't be alphanumeric.
    if unicode.IsLetter(l.peek()) {
        l.next()
        return l.errorf("bad number syntax: %q",
            l.input[l.start:l.pos])
    }
    l.emit(itemNumber)
    return lexText
}	

func lexQuotedSymbol(l *lexer) stateFn {
	if !l.acceptPredicate(unicode.IsLetter) {
		return l.errorf("quoted symbols must start with a letter, not %q",
			l.input[-1+l.start:])
	}
	l.acceptRunPredicate(unicode.IsLetter, unicode.IsNumber)
	l.emit(itemQuotedSymbol)
	return lexText
}

func lexSymbol(l *lexer) stateFn {
	// We've already accepted a letter.  Go until the first
	// non-alphanumeric thing
	l.acceptRunPredicate(
		unicode.IsLetter,
		unicode.IsNumber,
		func(c rune) bool {
			return unicode.IsPunct(c) && c != ')'
		},
	)
	l.emit(itemSymbol)
	return lexText
}

func lexBoolean(l *lexer) stateFn {
	if !l.accept("tf") {
		return l.errorf("Unrecognized boolean %q",
			l.input[-1+l.start:l.pos])
	}
	// else
	peek := l.peek()
	if !(unicode.IsSpace(peek) || peek == eof) {
		return l.errorf("Unrecognized boolean %q",
			l.input[-1+l.start:1+l.pos])
	}
	// else
	l.emit(itemBoolean)
	return lexText
}
