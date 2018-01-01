package main

import (
	"github.mheducation.com/dave-mcmath/scam/sexpr"

	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode/utf8"
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

	go func(in *os.File, ch chan<- rune) {
		err := fillRuneChannelFromFile(in, ch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading input:", err)
		}
	}(infile, chin)

	for sx := range sexprs {
		log.Println("Evaluating", sx)
		val := sexpr.Evaluate(sx)
		fmt.Println(val)
	}
}

// fillRuneChannelFromFile reads the content of the file (in) and
// pushs all the runes into the channel (ch).  If there is an error
// reading the file, we return that value.  The channel is always
// closed when the method returns.
func fillRuneChannelFromFile(in *os.File, ch chan<- rune) error {
	defer close(ch)

	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		cur := scanner.Text()
		pos := 0
		for r, sz := utf8.DecodeRuneInString(cur[pos:]) ; sz > 0 ; {
			ch <- r
			pos += sz
			r, sz = utf8.DecodeRuneInString(cur[pos:])
		}
	}
	return scanner.Err()
}
