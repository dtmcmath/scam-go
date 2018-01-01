package main

import (
	"github.mheducation.com/dave-mcmath/scam/sexpr"
	"github.mheducation.com/dave-mcmath/scam/scamutil"

	"flag"
	"fmt"
	"log"
	"os"
	"bufio"
)

var infilename = flag.String("in", "-", "input file ('-' for stdin)")

func main() {
	flag.Parse()

	chin := make(chan rune)
	_, sexprs := sexpr.Parse("repl", chin)

	var infile *os.File
	switch *infilename {
	case "-":
		infile = os.Stdin
	default:
		// Some scoping thing keeping us from assigning
		// straight to "infile" here?
		iii, err := os.Open(*infilename)
		if err != nil {
			msg := fmt.Sprintf("Failed to open file %q", *infilename)
			fmt.Fprintln(os.Stderr, msg, err)
			os.Exit(1)
		}
		infile = iii
	}

	go func(in *bufio.Scanner, ch chan<- rune) {
		err := scamutil.FillRuneChannelFromScanner(in, ch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading input:", err)
		}
	}(bufio.NewScanner(infile), chin)

	for sx := range sexprs {
		log.Println("Evaluating", sx)
		val := sexpr.Evaluate(sx)
		fmt.Println(val)
	}
}
