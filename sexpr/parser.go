package sexpr

import (
	"errors"
	"fmt"
	"log"
)

// The Parser consumes tokens and emits S-expressions.  An
// S-expression is either an Atom or a Cons of two S-expressions.

type nodeOfSexprLists struct {
	val []Sexpr
	next *nodeOfSexprLists
}
type stackOfSexprLists struct {
	head *nodeOfSexprLists
}
func (n nodeOfSexprLists) String() string {
	return fmt.Sprintf("%s", n.val)
}

func (s *stackOfSexprLists) push(l []Sexpr) {
	s.head = &nodeOfSexprLists{l, s.head}
}
func (s *stackOfSexprLists) pushHead(h Sexpr) {
	s.head.val = append(s.head.val, h)
}
func (s *stackOfSexprLists) pop() ([]Sexpr, error) {
	if s.head == nil {
		return nil, errors.New("Pop from an empty stack")
	}
	// else
	top := s.head
	s.head = s.head.next
	return top.val, nil
}
func (s *stackOfSexprLists) popHead() (Sexpr, error) {
	if s.head == nil {
		return nil, errors.New("PopHead from an empty stack")
	} else if len(s.head.val) == 0 {
		return nil, errors.New("PopHead but the head is empty")
	}
	// else
	first := s.head.val[0]
	s.head.val = s.head.val[1:]
	return first, nil
}
func (s stackOfSexprLists) String() string {
	if s.head == nil {
		return "<empty>"
	}
	// else
	return fmt.Sprintf("%s --> %s", s.head.val, s.head.next)
}

type parser struct {
	name string
	input string
	lex *lexer
	items <-chan item
	sexprs chan Sexpr
	stack *stackOfSexprLists
}

func (p *parser) emit(s Sexpr) {
	if p.stack.head == nil {
		// There is no context to roll up; we have a "final" S-expression
		p.sexprs <- s
	} else {
		p.stack.pushHead(s)
	}
}

func (p *parser) paniqf(format string, args ...interface{}) {
	// TODO:  The error needs to bubble up; it's like "panic", but
	// maybe we can do better.
	log.Printf(format, args...)
}
func (p *parser) unsupportedf(format string, args ...interface{}) {
	p.paniqf(format, args...)
}

func Parse(name, input string) (*parser, <-chan Sexpr) {
	p := &parser{
		name: name,
		input: input,
		sexprs: make(chan Sexpr),
		stack: &stackOfSexprLists{nil},
	}
	p.lex, p.items = Lex("lexer_"+name, input)
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
			var head []Sexpr
			p.stack.push(head)
		case itemRparen:
			slist, err := p.stack.pop()
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
			p.unsupportedf("We aren't ready for '%s' yet, «%s»",
				tok, p.input[p.lex.start:])
			return
		case itemWhitespace:
			continue
		default:
			p.paniqf("Unexpected token: %v", tok)
			return
		}
	}
}
