package sexpr

import (
	"errors"
	"fmt"
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
	itemComment               // ; ... \n
	itemDot                   // .
	itemNumber                // a numeric thing
	itemSingleQuote           // '
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
	case itemSingleQuote: return "QUOTE"
	case itemSymbol: return fmt.Sprintf("SYMBOL(%s)", i.val)
	case itemBoolean: return fmt.Sprintf("BOOL(%s)", i.val)
	case itemWhitespace: return "WHITESPACE"
	case itemQuotationMark: return "DOUBLEQUOTE"
	case itemError: return fmt.Sprintf("ERROR(%s)", i.val)
	default:
		panic(fmt.Sprintf("Unrecognized token in 'String': {%v, $v}", i.typ, i.val))
	}
}

const (
	eof rune = rune(0) // what you get when reading from a closed channel
)

type lexer struct {
	name  string    // used only for error reports.
	src   <-chan rune // the source of runes
	buf   chan rune // Another source of runes (peek/back)
	input string    // the string that has been scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

type stateFn func(*lexer) stateFn

type runeTester func(rune) bool
func mkLookupFunc (s string) runeTester {
	return func(r rune) bool {
		return strings.IndexRune(s, r) >= 0
	}
}

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
	// Non-blocking check to see whether something is buffered
	var ok bool
	select {
	case r, ok = <- l.buf:
		if r == eof {
			ok = false
		}
	default:
		r, ok = <- l.src
		if ok {
			l.input += string(r)
		}
	}
	if !ok {
		// BUG or FEATURE?  If used in "peek", this will make backup
		// impossible.
		l.width = 0
		return eof
	}
	l.width = utf8.RuneLen(r)
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
	if l.width == 0 {
		l.buf <- eof
		return
	}
	// else
	cur := l.pos
    l.pos -= l.width
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:cur])
	l.buf <- r
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() (r rune) {
    r = l.next()
    l.backup()
    return r
}

func matchOneOf(r rune, preds ...runeTester) bool {
	for _, pred := range preds {
		if pred(r) {
			return true
		}
	}
	return false
}
// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	return l.acceptPredicate(mkLookupFunc(valid))
}
func (l *lexer) acceptPredicate(preds ...runeTester) bool {
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
	l.acceptRunPredicate(mkLookupFunc(valid))
}
func (l *lexer) acceptRunPredicate(preds ...runeTester) {
	for l.acceptPredicate(preds...) {
	}
}

// acceptUntilPredicate consumes a run of runes until a rune matches
// one of the predicate conditions.  Once there is a match, we back up
// one step (un-ingest the rune) and return.
//
// EOF matches automatically/implicitly.
func (l *lexer) acceptUntilPredicate(preds ...runeTester) {
	for {
		r := l.next()
		if m := r == eof || matchOneOf(r, preds...) ; m {
			l.backup()
			return
		}
	}
	panic("Predicate-test found the second dimension")
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

func lex_string(name, input string) (*lexer, chan item) {
	return lex(name, mkRuneChannel(input))
}

func lex(name string, src <-chan rune) (*lexer, chan item) {
	l := &lexer{
		name: name,
		src: src,
		buf: make(chan rune, 1), // could use more...
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
// Helpers that "define" the language
////
var (
	isPartOfASymbol runeTester
	looksLikeSymbolTerminator runeTester
	looksLikeNumberStart runeTester
)

func init() {
	isPartOfASymbol = func() runeTester {
		// Oversimplified from https://www.scheme.com/tspl4/grammar.html#grammar:symbols
		l := mkLookupFunc("*+-^$/><=?") // "-" and "?" are punctuation
		return func (r rune) bool {
			switch {
			case unicode.IsLetter(r), l(r), '0' <= r && r <= '9' :
				return true
			case unicode.IsPunct(r):
				return !looksLikeSymbolTerminator(r)
			default:
				return false
			}
		}
	}()
	// looksLikeSymbolTerminator matches runes that would end a symbol-run
	looksLikeSymbolTerminator = func() runeTester {
		l := mkLookupFunc(")];([")
		return func (r rune) bool {
			return r == eof || unicode.IsSpace(r) || l(r)
		}
	}()
	looksLikeNumberStart = mkLookupFunc("+-")
}
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
		case r == '(' || r == '[':
			l.emit(itemLparen)
		case r == ')' || r == ']':
			l.emit(itemRparen)
		case r == ';':
			return lexComment
		case looksLikeNumberStart(r):
			// Look ahead to see whether it's a number or a symbol
			switch peek := l.peek() ; {
			case looksLikeSymbolTerminator(peek):
				return lexSymbol
			default:
				l.backup() // We read the number start; put it back
				return lexNumber
			}
		case '0' <= r && r <= '9':
			// It's going to be a number
			l.backup()
			return lexNumber
		case r == '\'':
			if l.peek() == '\'' {
				return l.errorf("invalid quote sequence %q", l.input[l.start:])
			}
			l.emit(itemSingleQuote)
		case r == '#':
			l.ignore() // Consume
			return lexBoolean
		case isPartOfASymbol(r):
			// No backup; lexSymbol expects to have read one
			return lexSymbol
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
	// If the next thing is symbol-like, we've been looking for a
	// symbol this whole time!!!
	nxt := l.peek()
	if !looksLikeSymbolTerminator(nxt) && isPartOfASymbol(nxt) {
		return lexSymbol
	}
	// else, it's a number
    l.emit(itemNumber)
    return lexText
}	

func lexSymbol(l *lexer) stateFn {
	// We've already accepted part.  Go until we find something
	// that doesn't apply.
	l.acceptRunPredicate(isPartOfASymbol)
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
	if !(unicode.IsSpace(peek) || peek == eof ||
		peek == ')' || peek == ']') {
		return l.errorf("Unrecognized boolean %q",
			l.input[-1+l.start:1+l.pos])
	}
	// else
	l.emit(itemBoolean)
	return lexText
}

func lexComment(l *lexer) stateFn {
	l.acceptUntilPredicate(mkLookupFunc("\n"))
	l.emit(itemComment)
	return lexText
}
