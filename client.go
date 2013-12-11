package main

import(
	"net"
	"fmt"
	"bufio"
	"os"
	"strconv"
	"bytes"
)

const (
	// Control bits
	Failure = 0
	Success = 1
	Request = 2
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
func (c *Choice) getInput(cl *Client, prompt string) {
	var s string
	var d int
	var err error
	for {
		fmt.Print(prompt)
		// scan for string so that we read the entire line
		// otherwise trailing chars wont get flushed
		s, _ = cl.reader.ReadString('\n')
		
		// convert to int and remove trailing newline
		d, err = strconv.Atoi(s[:len(s)-1])
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

// Process response and return the response string/error
func (c *Client) processResponse(res []byte) (string, error) {
	// fmt.Println("Received: ", res)

	// First bit is a control bit that indicates
	// 0: failure
	// 1: success
	// 2: request
	if res[0] == Failure {
		// return error with rest of response
		return "", fmt.Errorf("Error: %s", string(res[1:]))
	}
	if res[0] == Request {
		req := bytes.TrimRight(res[1:], string([]byte{0}))
		switch string(req) {
		case "get_lang":
			c.conn.Write([]byte(lang))
			return "", nil
		default:
			return "", fmt.Errorf("Bad request from server: %s", string(req))
		}
	}
	if res[0] == Success {
		return string(res[1:]), nil
	}

	// Otherwise bad response
	return "",fmt.Errorf("Bad response: %s", string(res))
}

func (c *Client) run() {
	// Echo incoming msg from server
	// go io.Copy(os.Stdin, c.conn) // dbg

	// Ask to choose language
	choice_lang := Choice{options: []int{1}}
	fmt.Printf("Choose client language:\n%s\n", "(1) English (2) Swedish")
	choice_lang.getInput(c, "Choose: >> ")

	// Request welcome message, (max 80 bytes)
	resp := make([]byte, 80)
	c.conn.Write([]byte("get_wmsg2"))
	var wmsg_str string
	var err error
	for wmsg_str == "" {
		c.conn.Read(resp)
		wmsg_str, err = c.processResponse(resp)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println(wmsg_str)


	// Print menu
	menu, err := getMenu(lang)
	if err != nil { 
		fmt.Printf("Could not load menu. %s\n", err)
		os.Exit(1)
	}
	fmt.Println(menu)

	// Prompt user for input
	choice := Choice{options: []int{1,2,3,4}}
	choice.getInput(c, "Choose: >> ")
	fmt.Printf("C: %d\n", choice.userInput) // dbg
}