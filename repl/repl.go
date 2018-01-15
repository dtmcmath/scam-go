package repl

import (
	"github.mheducation.com/dave-mcmath/scam/sexpr"

	"bufio"
	"fmt"
	"io"
)

type repl struct{
	name    string

	in      io.Reader
	out     io.Writer
	err     io.Writer

	preface string
	prompt  string
}

func New(name string, in io.Reader, out io.Writer, err io.Writer) repl {
	return repl{name, in, out, err, "", "> "}
}

func (r *repl) SetPreface(p string) { r.preface = p }
func (r *repl) SetPrompt(p string) { r.prompt = p }

func (r *repl) Run() {
	ch := make(chan rune)

	go func(in *bufio.Scanner, ch chan<- rune) {
		err := fillRuneChannelFromScanner(in, ch)
		if err != nil {
			fmt.Fprintln(r.err, "reading input:", err)
		}
	}(bufio.NewScanner(r.in), ch)

	fmt.Fprintln(r.out, r.preface)
	fmt.Fprint(r.out, r.prompt)

	_, sexprs := sexpr.Parse("parser_"+r.name, ch)
	for sx := range sexprs {
		back := make(sexpr.SexprChannel)
		go func() {
			// Since "sexpr.sexpr_general" isn't exported,
			// we can't actually name the signature of
			// this method!  We can't give "sx" a type.
			// But it works as an anonymous function,
			// somehow.  That seems either spooky or
			// buggy, and I can't tell which.
			defer func() {
				if err := recover() ; err != nil {
					ans := sexpr.MkEvaluationError("root", fmt.Sprintf("%+v", err))
					back <- ans
				}
			}()
			back <- sexpr.Evaluate(sx)
		}()
		val := <- back
		if _, err := sexpr.Fprint(r.out, val) ; err != nil {
			fmt.Fprint(r.err, "!!ERROR: ", err)
		}
		fmt.Fprintln(r.out, "")
		fmt.Fprint(r.out, r.prompt)
	}
}
