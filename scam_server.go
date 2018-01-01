package main

import (
	"github.mheducation.com/dave-mcmath/scam/sexpr"
	"github.mheducation.com/dave-mcmath/scam/scamutil"

	"flag"
	"log"
	"net"
	"fmt"
	"bufio"
)

var port = flag.Int("port", 8000, "where to listen")
func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on port %d", *port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan rune)

	_, sexprs := sexpr.Parse("repl", ch)

	go func(sc *bufio.Scanner) {
		err := scamutil.FillRuneChannelFromScanner(sc, ch)
		if err != nil {
			log.Print(err)
		}
	}(bufio.NewScanner(conn))

	for sx := range sexprs {
		log.Println("Evaluating", sx)
		val := sexpr.Evaluate(sx)
		// TODO:  Actually send it back?  Look at gopl.io/ch8/chat.go
		fmt.Println(val)
	}
}
	
