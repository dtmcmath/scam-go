package main

import (
	"github.mheducation.com/dave-mcmath/scam/repl"

	"flag"
	"log"
	"net"
	"fmt"
	"os"
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
	log.Printf("Accepted a connection from %s", conn.RemoteAddr())

	r := repl.New(
		conn.RemoteAddr().String(),
		conn,
		conn,
		os.Stderr,
	)
	r.SetPrompt(fmt.Sprintf("write(%d)> ", *port))
	r.Run()
}
	
