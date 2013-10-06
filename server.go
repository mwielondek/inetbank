package main

import (
	"fmt"
	"log"
	"net"
)

const (
	port = "1337"
)

func main() {
	// start the server
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Bank started listening on port " + port)

	// shut down server before exiting
	defer l.Close()

	// loop forever
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Connection from %s\n", conn.LocalAddr())

		// handle conn in a new goroutine
		go func(c net.Conn) {
			defer func() {
				c.Close()
				log.Printf("Connection from %s closed.\n", conn.LocalAddr())
			}()

			fmt.Fprintln(c, "Welcome to Bank!")
			var request []byte
			for {
				if n, err := fmt.Fscan(c, &request); err == nil && n > 0 {
					if answer, err := processRequest(request); err != nil {
						log.Printf("Bad request (%b)\n", request)
					} else {
						fmt.Fprint(c, answer)
					}
				}
			}
		}(conn)
	}
}

func processRequest(req []byte) (ans []byte, err error) {
	return
}
