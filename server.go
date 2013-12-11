package main

import (
	"fmt"
	"log"
	"net"
	"time"
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
			// set idle timeout
			c.SetDeadline(time.Now().Add(5*time.Second))

			// create request buffer
			request := make([]byte, 10)
			_, err := c.Read(request)

			// check for Timeout (err.(net.Error) => type assertion)
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				log.Printf("Connection from %s timed out.\n", conn.LocalAddr())
				return // exit goroutine and close conn
			}
			// for {
			// 	if n, err := fmt.Fscan(c, &request); err == nil && n > 0 {
			// 		if answer, err := processRequest(request); err != nil {
			// 			log.Printf("Bad request (%b)\n", request)
			// 		} else {
			// 			fmt.Fprint(c, answer)
			// 		}
			// 	}
			// }
		}(conn)
	}
}

func processRequest(req []byte) (ans []byte, err error) {
	// TODO
	return
}
