package main

import (
	"log"
	"github.mheducation.com/dave-mcmath/scam/sexpr"
)

func main() {
	s := "(cons 1 2)"

	_, ch := sexpr.Parse("repl", s)
	for sx := range ch {
		log.Println("Evaluating", sx)
		val := sexpr.Evaluate(sx)
		log.Println("...value=", val)
	}
}
