# SCAM, A Schemelike Computational Algorithm Machine

The goal of this project has been to implement, in Go, enough of a
Scheme interpreter to get through the examples in
[_The Little Schemer_ (4th edition)](https://mitpress.mit.edu/books/little-schemer)
by Daniel Friedman and Matthias Felleisen.  Helpfully,
[all the examples from the book](https://github.com/pkrumins/the-little-schemer)
are available, so that will become a test suite.  Thanks, Peter
Krumins.

This goal was accomplished around rev. c37065c.


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
apologize to the professor for my particular blend of
a know-it-all overconfidence and a deep ignorance of how to manage
code.  Anyway, I remember having bugs in my Tokenizer that plagued me
all semester, so the idea that a lexer could be easy was appealing.

I've used `flex` to generate C code, but it turned out to be a mess.
So I was intrigued by the idea that a lexer could be "easy".  It
turned out that Mr. Pike was right.

## Installation

### Mac OS X

Use Homebrew to install the latest version of Go (1.9 as of this writing). There are some prerequisities and some best practices. Here's the rundown:


1. Install the latest XCode
2. Install the XCode command line utilities (note: you may need to accept the license agreement for these to function)
3. `brew update`
4. `brew doctor` and correct any indicated issues
5. `brew update` for good measure
6. `brew install go`
7. `go version` should execute and indicate appropriate version

On Mac OS X, GOPATH environment variable is set to /Users/{username}. As a result, Go will be looking for packages included in your code someplace it is unlikely to find it. You might, on execution, see something like this:

```
scam.go:4:2: cannot find package "github.mheducation.com/dave-mcmath/scam/repl" in any of:
	/usr/local/Cellar/go/1.9.2/libexec/src/github.mheducation.com/dave-mcmath/scam/repl (from $GOROOT)
	/Users/andrewlippert/go/src/github.mheducation.com/dave-mcmath/scam/repl (from $GOPATH)
```

Drop a symlink in the appropriate location, pointing at your repo location, and everything will be right with the world.

You should now be able to successfully execute SCAM.

(PC: TODO)

## Usage

### ...without compiling

To evaluate a file, just

    go run cmd/scam/main.go -in examples/arithmetic.ss

To use `scam` as a REPL, use

    go run cms/scam/main.go

and type.  There is no "exit" command; send end-of-file with `C-d`.

Or

    go run cmd/scam_server/main.go -port 8000 &
    echo "(= (car (cons 1 2)) 2)" | nc localhost 8000

### ...with compiling

Install to `$GOPATH/bin` with

    go install cmd/scam/main.go
    go install cmd/scam_server/main.go

to create executables

    $GOPATH/bin/scam
    $GOPATH/bin/scam_server

respectively

## For further reading

* We maintain a list of things [to do](./TODO.org).

<!--  LocalWords:  goroutines Schemelike Matthias Felleisen Krumins
 -->
<!--  LocalWords:  REPL
 -->
