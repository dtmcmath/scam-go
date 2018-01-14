package sexpr

import (
	"errors"
	"fmt"
	// "runtime/debug"
	"strings"
)

// The Parser consumes tokens and emits S-expressions.  An
// S-expression is either an sexpr_atom or a sexpr_cons of two S-expressions.

type sexpr_parse_artifact struct{
	name string
}
func (a sexpr_parse_artifact) Sprint() string {
	return fmt.Sprintf("MARKER(%s)", a.name)
}

var (
	emptyStackError error = errors.New("Pop from an empty stack")
	markerLPAREN sexpr_parse_artifact = sexpr_parse_artifact{"LPAREN"}
	markerQUOTE  sexpr_parse_artifact = sexpr_parse_artifact{"QUOTE"}
)

type stackOfSexprs struct {
	val sexpr_general
	next *stackOfSexprs
}

func (s stackOfSexprs) String() string {
	var stack []string
	for cur := &s ; cur != nil ; cur = cur.next {
		stack = append([]string{fmt.Sprintf("%s", cur.val)}, stack...)
	}
	return strings.Join(stack, " <-- ")
}

type parser struct {
	name string
	lex *lexer
	items <-chan item
	sexprs chan sexpr_general
	// State-type things
	stack *stackOfSexprs
}

// stack managment tools
func (p *parser) pushStack(l sexpr_general) {
	p.stack = &stackOfSexprs{l, p.stack}
}
func (p *parser) popStack() (sexpr_general, error) {
	if p.stack == nil {
		return nil, emptyStackError
	}
	// else
	top := p.stack
	p.stack = p.stack.next
	return top.val, nil
}
func (p *parser) mustPopStack() sexpr_general {
	ans, err := p.popStack()
	if err != nil { panic(err) }
	return ans
}
// peek looks at the top Sexpr.  If the stack is empty, just return
// nil (instead of throwing an error), which is different than Nil, so
// that's not ambigious
func (p *parser) peekStack() sexpr_general {
	if p.stack == nil {
		return nil
	}
	// else
	return p.stack.val
}

// parse qua parse tools

// popUntil removes elements from HEAD until reaching an object that
// equals (==) the marker.  It returns an array of S-expressions
// removed.  That array is reversed with respect to the order popped.
// So if the stack is
//
//   x -> y -> z -> marker
//
// the return value is
//
//   {z, y, x}
//
// If no marker is found, an error is returned.
func (p *parser) popStackUntil(marker sexpr_general) ([]sexpr_general, error) {
	var acc []sexpr_general
	for {
		head, err := p.popStack()
		if err == emptyStackError {
			return nil,  errors.New(fmt.Sprintf("popUntil(%s) from a stack with no %q", marker, marker))
		} else if err != nil {
			return nil, err
		} else if head == marker {
			return acc, nil
		} else {
			acc = append([]sexpr_general{head}, acc...)
		}
	}
}

func (p *parser) emit(s sexpr_general) {
	top := p.peekStack()
	if top == markerQUOTE {
		p.mustPopStack()
		s = mkList(atomConstantQuote, s)
	}
	if p.stack == nil {
		// There is no context to roll up; we have a "final" S-expression
		p.sexprs <- s
	} else {
		p.pushStack(s)
	}
}

func (p *parser) paniqf(format string, args ...interface{}) {
	// TODO:  The error needs to bubble up; it's like "panic", but
	// maybe we can do better.
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf("PARSE ERROR: " + format, args...)
	// TODO:  Give some indication of _where_!!
	fmt.Printf("«TODO:  Better parse-error context»\n")
	// debug.PrintStack()
}
func (p *parser) unsupportedf(format string, args ...interface{}) {
	p.paniqf(format, args...)
}

func Parse(name string, input <-chan rune) (*parser, <-chan sexpr_general) {
	p := &parser{
		name: name,
		sexprs: make(chan sexpr_general),
		stack: nil,
	}
	p.lex, p.items = lex("lexer_"+name, input)
	go p.run()
	return p, p.sexprs
}

func (p *parser) run() {
	defer close(p.sexprs)
	for {
		tok := <- p.items
		switch tok.typ {
		case itemEOF:
			// If there's something on the stack, we have a problem
			if p.peekStack() != nil {
				p.paniqf("Unexpected EOF")
			}
			return
		case itemLparen:
			p.pushStack(markerLPAREN)
		case itemRparen:
			slist, err := p.popStackUntil(markerLPAREN)
			if err != nil {
				p.paniqf(err.Error())
				return
			}
			// else
			s := consify(slist)
			p.emit(s)
		case itemSingleQuote:
			p.pushStack(markerQUOTE)
		case itemNumber:
			p.emit(mkAtomNumber(tok.val))
		case itemSymbol:
			p.emit(mkAtomSymbol(tok.val))
		case itemBoolean:
			switch tok.val {
			case "t": p.emit(atomConstantTrue)
			case "f": p.emit(atomConstantFalse)
			default:
				p.paniqf("Illegal boolean token %v", tok)
			}
		case itemQuotationMark, itemDot:
			p.unsupportedf("We aren't ready for '%s' yet", tok)
			return
		case itemWhitespace, itemComment:
			continue
		default:
			p.paniqf("Unexpected token: %v", tok)
			return
		}
	}
}
