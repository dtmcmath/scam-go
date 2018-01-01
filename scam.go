package main

import (
	"github.mheducation.com/dave-mcmath/scam/repl"

	"flag"
	"fmt"
	"io"
	"os"
	"errors"
)

var infilename = flag.String("in", "-", "input file ('-' for stdin)")

type teeReader struct{
	in  io.Reader
	tee io.Writer
}
func (t teeReader) Read(p []byte) (n int, err error) {
	n, err = t.in.Read(p)
	if err == nil {
		// Success, so tee out the bytes
		_, err := t.tee.Write(p)
		if err != nil {
			// There was an error writing.
			// /me sighs
			err = errors.New("Tee-Error: " + err.Error())
			return 0, err
		}
	}
	return n, err
}


func main() {
	flag.Parse()

	//	var infile *os.Reader
	var infile io.Reader
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
		infile = teeReader{
			in: iii,
			tee: os.Stdout,
		}
	}

	r := repl.New("scam", infile, os.Stdout, os.Stderr)
	r.SetPreface(`SCAM Version 0.1
Please be gentle
`)
	r.SetPrompt("> ")

	r.Run()
}
