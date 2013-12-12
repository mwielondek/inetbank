package main

import(
	"net"
	"fmt"
	"bufio"
	"os"
	"strconv"
	"strings"
	. "./tools"
)

const (
	// Control bits
	Failure = 0
	Success = 1
	Request = 2

	prompt_suffix = " >> "

	available_langs = "en,sv"
)

var (
	lang = "en" //default lang
	prompt = "Choose"+prompt_suffix // default prompt
)

func main() {
	c := new(Client)
	c.create()
	c.run()
	c.conn.Close()
}

// ================== MENU OPTIONS ==================

// placeholder function
func void() {}

func exit() {
	os.Exit(0)
}

func balance() {}

func withdraw() {}

func deposit() {}


// ===================== CHOICE =====================
type MenuFunc func()
type Choice struct {
	userInput int
	funcs []MenuFunc
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
		if err == nil && c.validateInput(d) {
			break
		}
		fmt.Println("Invalid input")
	}
	c.userInput=d
}

// Checks if given input is within range
func (c *Choice) validateInput(i int) bool {
	return i >= 1 && i <= len(c.funcs)
}

// executes function with index userInput-1 from funcs-array
func (c *Choice) exec() {
	c.funcs[c.userInput-1]()
}

// ================== TRANSLATOR ==================
type Translator struct {
	lang string
	dict map[string]string
}

// loads the locale file for translation
func createTranslator(lang string) (*Translator, error) {
	locale, err := os.Open("files/"+lang+"/locale.txt")
	defer locale.Close()
	if err != nil {
		return nil, err
	}
	dict := make(map[string]string)
	sc := bufio.NewScanner(locale)
	for sc.Scan() {
		word := strings.Split(sc.Text(), "::")
		key := strings.TrimSpace(word[0])
		value := strings.TrimSpace(word[1])
		dict[key] = value
	}

	t := &Translator{lang: lang, dict: dict}
	return t, nil
}

// give enum name and lang, return translation
func (t *Translator) translate(name string) string {
	return t.dict[name]
}

// Returns the menu prompt
func (c *Client) getMenu() (string, error) {
	menu, err := os.Open("files/"+c.lang.lang+"/menu.txt")
	defer menu.Close()
	reader := bufio.NewReader(menu)
	o, _ := reader.ReadString('\n')

	return o, err
}

// ==================== CLIENT =====================
type Client struct {
	conn net.Conn
	reader *bufio.Reader
	lang *Translator
}

func (c *Client) create() {
	c.conn = c.connect()
	reader := bufio.NewReader(os.Stdin)
	c.reader = reader
}

func (c *Client) connect() (conn net.Conn){
	conn, err := net.Dial("tcp4", "localhost:1337")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return conn
}

// Process response and return the response string/error
func (c *Client) processResponse(res []byte) (string, error) {
	// First bit is a control bit that indicates
	// 0: failure
	// 1: success
	// 2: request
	if res[0] == Failure {
		// return error with rest of response
		return "", fmt.Errorf("Failed: %s", string(res[1:]))
	}
	if res[0] == Request {
		req := BytesToString(res[1:])
		switch req {
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

func (c *Client) setLanguage(lang_index int) {
	lang = strings.Split(available_langs, ",")[lang_index]
	// set translator
	var err error
	c.lang, err = createTranslator(lang)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// set prompt
	prompt = Capitalize(c.lang.translate("choose")) + prompt_suffix
}

func (c *Client) run() {

	// Ask to choose language
	choice_lang := Choice{funcs: []MenuFunc{void, void}}
	fmt.Printf("Choose client language:\n%s\n", "(1) English (2) Swedish")
	choice_lang.getInput(c, prompt)
	c.setLanguage(choice_lang.userInput-1)

	// Request welcome message, (max 80 bytes)
	resp := make([]byte, 1000)
	c.conn.Write([]byte("get_wmsg"))
	var wmsg_str string
	var err error
	for wmsg_str == "" {
		c.conn.Read(resp)
		wmsg_str, err = c.processResponse(resp)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	fmt.Println(wmsg_str)


	// Print menu
	menufuncs := []MenuFunc{balance, withdraw, deposit, exit}
	choice := Choice{funcs: menufuncs}
	menu, err := c.getMenu()
	if err != nil { 
		fmt.Printf("Could not load menu. %s\n", err)
		os.Exit(1)
	}
	for {
		fmt.Println(menu)
		// Prompt user for input
		choice.getInput(c, prompt)
		choice.exec()
	}
}