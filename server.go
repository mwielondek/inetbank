package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"bytes"
)

const (
	port = "1337"
	timeout = 30 // in seconds
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
			defer c.Close()

			// create request buffer, max 10 bytes
			request := make([]byte, 10)
			for {
				// set idle timeout
				c.SetDeadline(time.Now().Add(timeout*time.Second))

				// read
				_, err := c.Read(request)

				if err != nil {
					// check for Timeout (err.(net.Error) => type assertion)
					if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
						log.Printf("Connection from %s timed out.\n", conn.LocalAddr())
					} else {
						log.Printf("Connection from %s closed remotely.\n", conn.LocalAddr())
					}
					return // exit goroutine and close conn
				}

				// process request
				log.Printf("Request from %s: %s\n", conn.LocalAddr(), string(request))
				perr := processRequest(c, request)
				if perr != nil {
					fmt.Println(perr)
				}
			}
		}(conn)
	}
}

func processRequest(c net.Conn, req []byte) error {
	// trim trailing 0's
	req = bytes.TrimRight(req, string([]byte{0}))
	// conv to string
	sreq := string(req)
	
	switch sreq {

	// Get welcome message
	case "get_wmsg":
		c.Write([]byte("Knock knock! It's us!"))
	default:
		return fmt.Errorf("Request fallthrough (bad req: %s)", sreq)
	}
	return nil
}