package main

import(
	"net"
	"fmt"
	"bufio"
	"os"
	"strconv"
)

var (
	lang = "en" //default lang
)

type Options []int
type Choice struct {
	options Options
	userInput int
}

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

// Gets input from user
func (c *Choice) getInput(prompt string) {
	var s string
	var d int
	var err error
	for {
		fmt.Print(prompt)
		// scan for string so that we read the entire line
		// otherwise trailing chars wont get flushed
		fmt.Scanf("%s", &s)
		
		// convert to int and then to userInput
		d, err = strconv.Atoi(s)
		if err == nil && c.options.validateInput(d) {
			break
		}
		fmt.Println("Invalid input")
	}
	c.userInput=d
}

// Returns the menu prompt
func getMenu(lang string) (string, error) {
	menu, err := os.Open("files/"+lang+"/menu.txt")
	reader := bufio.NewReader(menu)
	o, _ := reader.ReadString('\n')

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

// Checks if given input is in the options list
func (o Options) validateInput(i int) bool {
	for _, x := range o {
		if x == i {
			return true
		}
	}
	return false
}

func (c *Client) run() {

	// Echo incoming msg from server
	// go io.Copy(os.Stdin, c.conn) // dbg

	// Request welcome message
	wmsg := make([]byte, 100)
	c.conn.Write([]byte("get_wmsg"))
	c.conn.Read(wmsg)
	fmt.Println(string(wmsg))

	// Print menu
	menu, err := getMenu(lang)
	if err != nil { 
		fmt.Printf("Could not load menu. %s\n", err)
		os.Exit(1)
	}
	fmt.Println(menu)

	// Prompt user for input
	choice := Choice{options: []int{1,2,3,4}}
	choice.getInput("Choose: >> ")
	fmt.Printf("C: %d\n", choice.userInput) // dbg
}