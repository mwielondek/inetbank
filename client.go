package main

import(
	"net"
	"fmt"
	"bufio"
	"os"
)

var (
	lang = "en" //default lang
)

type userInput int

type Client struct {
	conn net.Conn
	buff []byte
	reader *bufio.Reader
}

func main() {
	c := new(Client)
	c.create()
	c.run()
	c.conn.Close()
}

func (c *Client) create() {
	// setting up client
	buff := make([]byte, 1000)
	c.buff = buff
	c.conn = c.connect()
	reader := bufio.NewReader(os.Stdin)
	c.reader = reader
}

func (c *Client) run() {

	// Echo incoming msg from server
	// go io.Copy(os.Stdin, c.conn)

	// Print menu
	menu, err := getMenu(lang)
	if err != nil { 
		fmt.Printf("Could not load menu. %s\n", err)
		os.Exit(1)
	}
	fmt.Print(menu)

	// Prompt user for input
	var choice userInput
	choice.getInput("Choose: >> ")
	// fmt.Printf("C: %d\n", choice) // dbg
}

func (c *userInput) getInput(prompt string) {
	fmt.Print(prompt)
	_, err := fmt.Scanf("%d", c)
	if err != nil {
		fmt.Println(err)
	}
}

func getMenu(lang string) (string, error) {
	menu, err := os.Open("files/"+lang+"/menu.txt")
	scanner := bufio.NewScanner(menu)
	var o string
	for scanner.Scan() {
		o += scanner.Text() + "\n"
	}

	return o, err
}

func (c *Client) connect() (conn net.Conn){
	conn, err := net.Dial("tcp4", "localhost:1337")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return conn
}