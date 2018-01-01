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
var (
	emptyStackError error = errors.New("Pop from an empty stack")
	markerLPAREN sexpr_parse_artifact = sexpr_parse_artifact{"LPAREN"}
)

type nodeOfSexprs struct {
	val Sexpr
	next *nodeOfSexprs
}
type stackOfSexprs struct {
	head *nodeOfSexprs
}
func (n nodeOfSexprs) String() string {
	return fmt.Sprintf("%s", n.val)
}

func (s *stackOfSexprs) push(l Sexpr) {
	s.head = &nodeOfSexprs{l, s.head}
}
func (s *stackOfSexprs) pop() (Sexpr, error) {
	if s.head == nil {
		return nil, emptyStackError
	}
	// else
	top := s.head
	s.head = s.head.next
	return top.val, nil
}
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
func (s *stackOfSexprs) popUntil(marker Sexpr) ([]Sexpr, error) {
	var acc []Sexpr
	for {
		head, err := s.pop()
		if err == emptyStackError {
			return nil,  errors.New(fmt.Sprintf("popUntil(%s) from a stack with no %q", marker, marker))
		} else if err != nil {
			return nil, err
		} else if head == marker {
			return acc, nil
		} else {
			acc = append([]Sexpr{head}, acc...)
		}
	}
}
func (s stackOfSexprs) String() string {
	if s.head == nil {
		return "<empty>"
	}
	// else
	return fmt.Sprintf("%s --> %s", s.head.val, s.head.next)
}

type parser struct {
	name string
	lex *lexer
	items <-chan item
	sexprs chan Sexpr
	stack *stackOfSexprs
}

func (p *parser) emit(s Sexpr) {
	if p.stack.head == nil {
		// There is no context to roll up; we have a "final" S-expression
		p.sexprs <- s
	} else {
		p.stack.push(s)
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

func Parse(name string, input <-chan rune) (*parser, <-chan Sexpr) {
	p := &parser{
		name: name,
		sexprs: make(chan Sexpr),
		stack: &stackOfSexprs{nil},
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
			if p.stack.head != nil {
				p.paniqf("Unexpected EOF")
			}
			return
		case itemLparen:
			p.stack.push(markerLPAREN)
		case itemRparen:
			slist, err := p.stack.popUntil(markerLPAREN)
			if err != nil {
				p.paniqf(err.Error())
				return
			}
			// else
			// TODO:  If a "dot" is in play, we have something different
			s := consify(slist)
			p.emit(s)
		case itemNumber:
			p.emit(mkAtomNumber(tok.val))
		case itemSymbol:
			p.emit(mkAtomSymbol(tok.val))
		case itemBoolean:
			switch tok.val {
			case "t": p.emit(True)
			case "f": p.emit(False)
			default:
				p.paniqf("Illegal boolean token %v", tok)
			}
		case itemQuotedSymbol:
			p.emit(mkAtomQuoted(tok.val))
		case itemQuotationMark, itemDot:
			p.unsupportedf("We aren't ready for '%s' yet", tok)
			return
		case itemWhitespace:
			continue
		default:
			p.paniqf("Unexpected token: %v", tok)
			return
		}
	}
}
