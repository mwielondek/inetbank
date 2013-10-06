package main

import(
	"net"
	"fmt"
	"bufio"
	"os"
	"io"
)

type Client struct{
	conn net.Conn
	buff []byte
	reader *bufio.Reader
}

func main(){
	c := new(Client)
	c.run()
	c.conn.Close()
}

func (c *Client) run(){
	buff := make([]byte, 1000)
	c.buff = buff
	c.conn = c.connect()
	reader := bufio.NewReader(os.Stdin)
	c.reader = reader

	// Echo incoming msg from server
	go io.Copy(os.Stdin, c.conn)
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil { fmt.Println(err)}
	fmt.Printf("C: %d\n", choice)
	c.conn.Close()
}

func (c *Client) connect() (conn net.Conn){
	conn, err := net.Dial("tcp4", "localhost:1337")
	if err != nil {fmt.Println(err)}
	return conn
}
// ...