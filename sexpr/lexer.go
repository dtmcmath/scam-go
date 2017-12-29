package sexpr

import (
	"fmt"
//	"log"
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
	case itemWhitespace: return "WHITESPACE"
	case itemQuotationMark: return "QUOTE"
	case itemError: return fmt.Sprintf("ERROR(%s)", i.val)
	default:
		panic(fmt.Sprintf("No way:  token %v", i))
	}
}

const (
	eof rune = utf8.RuneError // Find a better "eof" rune
)

const (
	leftParen = '('
	rightParen = ')'
	dot = '.'
	singleQuote = '\''
	doubleQuore = '"'
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
        l.width = 0
        return eof
    }
    r, l.width =
        utf8.DecodeRuneInString(l.input[l.pos:])
    l.pos += l.width
    return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
    l.start = l.pos
}
// backup steps back one rune.
// Can be called only once per call of next.
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

func Lex(name, input string) (*lexer, chan item) {
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
		switch r := l.next(); {
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
		case r == '+' || r == '-' || '0' <= r && r <= '9':
			l.backup()
			return lexNumber
		case r == '\'':
			l.ignore()
			return lexQuotedSymbol
		case unicode.IsLetter(r):
			l.backup()
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
		return l.errorf("bad quote syntax: %q",
			l.input[l.start:l.pos])
	}
	l.acceptRunPredicate(unicode.IsLetter, unicode.IsNumber)
	l.emit(itemQuotedSymbol)
	return lexText
}

func lexSymbol(l *lexer) stateFn {
	if !l.acceptPredicate(unicode.IsLetter) {
		return l.errorf("bad symbol syntax: %q",
			l.input[l.start:l.pos])
	}
	l.acceptRunPredicate(unicode.IsLetter, unicode.IsNumber)
	l.emit(itemSymbol)
	return lexText
}

// func lexLeftMeta(l *lexer) stateFn {
//     l.pos += len(leftMeta)
//     l.emit(itemLeftMeta)
//     return lexInsideAction    // Now inside {{ }}.
// }


// func lexInsideAction(l *lexer) stateFn {
//     // Either number, quoted string, or identifier.
//     // Spaces separate and are ignored.
//     // Pipe symbols separate and are emitted.
//     for {
//         if strings.HasPrefix(l.input[l.pos:], rightMeta) {
//             return lexRightMeta
//         }
//         switch r := l.next(); {
//         case r == eof || r == '\n':
//             return l.errorf("unclosed action")
//         case isSpace(r):
//             l.ignore()
//         case r == '|':
//             l.emit(itemPipe)
//         case r == '"':
//             return lexQuote
//         case r == '`':
//             return lexRawQuote
//         case r == '+' || r == '-' || '0' <= r && r <= '9':
//             l.backup()
//             return lexNumber
//         case isAlphaNumeric(r):
//             l.backup()
//             return lexIdentifier			
