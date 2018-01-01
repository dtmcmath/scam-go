# SCAM, A Schemelike Computational Algorithm Machine

The goal of this project is to implement, in Go, enough of a Scheme
interpreter to get through the examples in
[_The Little Schemer_ (4th edition)](https://mitpress.mit.edu/books/little-schemer)
by Daniel Friedman and Matthias Felleisen.  Helpfully,
[all the examples from the book](https://github.com/pkrumins/the-little-schemer)
are available, so that will become a test suite.  Thanks, Peter
Krumins.


## Motivation

I wanted to learn Go, and I wanted to do something sort of hard.  I
probably could have written a web service that would have been good
for W$rk, but this was fun.  From the little I understand of Go so
far, "goroutines" are the most novel concept.  It did seem like the
read-eval-print loop from Scheme would be a good way to see goroutines
in action.

Browsing around, I found "Lexical Scanning in Go" by Rob Pike
([video](https://youtu.be/HxaD_trXwRE),
[slides](https://talks.golang.org/2011/lex.slide)) where he says "we
should write our own [lexer], 'cause it's easy, right?"  Huh.

Back when I was an undergraduate, the first CS class was in Scheme and
the second was in C++.  One of the big C++ projects was to write a
Scheme(-like) interpreter.  I did a terrible job.  I should probably
apologize to the professor, dougm@rice.edu for my particular blend of
a know-it-all overconfidence and a deep ignorance of how to manage
code.  Anyway, I remember having bugs in my Tokenizer that plagued me
all semester, so the idea that a lexer could be easy was appealing.

I've used `flex` to generate C code, but it turned out to be a mess.
So I was intrigued by the idea that a lexer could be "easy".  It
turned out that Mr. Pike was right.

## Installation

(TODO)

## Usage

To evaluate a file, just

    go run scam.go -in examples/arithmetic.ss

To use `scam` as a REPL, use

    go run scam.go

and type.

Or

    go run scam_server.go -port 8000 &
    echo "(eq? (car (cons 1 2) 2))" | nc localhost 8000

## For further reading

* We maintain a list of things [to do](./TODO.org).

<!--  LocalWords:  goroutines Schemelike Matthias Felleisen Krumins
 -->
<!--  LocalWords:  REPL
 -->
