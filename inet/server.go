package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"bytes"
	"os"
	"strconv"
	"bufio"
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
	. "./tools"
)

const (
	port = "1337"
	timeout = 300 // in seconds

	// Control bits
	Failure = 0
	Success = 1
	Request = 2

	db_filename = "./db.sqlite3"
)

var (
	db *sql.DB
)

type ClientConn struct {
	conn net.Conn
	user_id string
}

func connectToDatabase(filename string) (db *sql.DB) {
	db, err := sql.Open("sqlite3",db_filename)
	if err != nil {
		log.Printf("Error connecting to database: %s\n", err)
		os.Exit(1)
	} else {
		log.Printf("Established connection with database %s\n", db_filename)
	}
	return
}

func initDatabase(db *sql.DB) error {
	query := "CREATE TABLE codes(id INTEGER PRIMARY KEY ASC AUTOINCREMENT, code INTEGER, "
	query += "ownerid INTEGER, FOREIGN KEY(ownerid) REFERENCES users(id));"
	err := dbexec(db, query)
	if err != nil {
		return err
	}

	query = "CREATE TABLE users(id INTEGER PRIMARY KEY ASC AUTOINCREMENT, cardnr STRING, "
	query += "pin INTEGER, balance INTEGER);"
	err = dbexec(db, query)
	if err != nil {
		return err
	}
	return nil
}

func dbexec(db *sql.DB, query string) error {
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("Could not execute query: %s", query)
	}
	return nil
}

func main() {
	// ============================= DATABASE =============================
	defer db.Close()
	// check if db exists
	if _, err := os.Stat(db_filename); os.IsNotExist(err) {
		// file doesnt exist -> create it
		log.Printf("Database not found, attempting to create %s\n", db_filename)
		_, ferr := os.Create(db_filename)
		if ferr != nil {
			log.Printf("Failed attempt to create database: %s\n", ferr)
			os.Exit(1)
		}
		// create db connection
		db = connectToDatabase(db_filename)
		// init new database
		err = initDatabase(db)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}

	// Connect to database if no conn established
	if db == nil {
		db = connectToDatabase(db_filename)
	}
	

	// ============================= SERVER =============================
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

		clc := new(ClientConn)
		clc.conn = conn

		// handle conn in a new goroutine
		go func(cl *ClientConn) {
			c := cl.conn
			defer c.Close()

			var request []byte
			for {
				// set idle timeout
				c.SetDeadline(time.Now().Add(timeout*time.Second))

				// create/clear request buffer, max 10 bytes
				request = make([]byte, 10)

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
				perr := cl.processRequest(request)
				if perr != nil {
					fmt.Println(perr)
					c.Write(CreateResponse("Error processing request: "+perr.Error(), Failure))
				}
			}
		}(clc)
	}
}

func (c *ClientConn) processRequest(req []byte) error {
	resp := make([]byte, 10)
	
	sreq := BytesToString(req)
	
	switch sreq {

	// Get welcome message
	case "get_wmsg":
		// Get client language
		resp = CreateResponse("get_lang", Request)
		c.conn.Write(resp)
		// clear resp between wr/rd
		// and set limit to 80 bytes
		resp = make([]byte, 80)
		c.conn.Read(resp)
		lang := BytesToString(resp)

		// Load welcome message in given language
		message, err := os.Open("files/"+lang+"/welcome_message.txt")
		defer message.Close()
		if err != nil {
			return fmt.Errorf("Could not load welcome message: %s", err)
		}
		reader := bufio.NewReader(message)
		welcome_message, _ := reader.ReadString('\n')
		resp = []byte(welcome_message)
		resp = bytes.Join([][]byte{{Success},resp}, []byte{})
		c.conn.Write(resp)
	case "login":
		// The size of username should be 20 chars and password only 4.
		// The size is left intentionally much bigger to protect from
		// buffer overflows. However, this should be changed, either
		// through use of a buffer or using some other assertion logic.
		username, password := make([]byte, 256), make([]byte, 256)
		
		// get credentials
		c.conn.Write(CreateResponse("get_user", Request))
		c.conn.Read(username)
		c.conn.Write(CreateResponse("get_passw", Request))
		c.conn.Read(password)

		// Check user -> account
		var user_id, pin string
		err := db.QueryRow("SELECT id FROM users WHERE cardnr = ?", BytesToString(username)).Scan(&user_id)
		if err != nil && err == sql.ErrNoRows {
			c.conn.Write(CreateResponse("No matching account found", Failure))
			break
		}
		c.user_id = user_id

		// Check user PIN
		db.QueryRow("SELECT pin FROM users WHERE id = ?", user_id).Scan(&pin)
		if pin == BytesToString(password) {
			// c.user_id = 
			c.conn.Write(CreateResponse("Authenticated", Success))
		} else {
			c.conn.Write(CreateResponse("Wrong PIN", Failure))
		}
	case "get_blnce":
		balance := c.get_balance()
		c.conn.Write(CreateResponse(strconv.Itoa(balance), Success))
	case "withdraw":
		// user must authenticate with a one-time code
		c.conn.Write(CreateResponse("authcode", Request))
		resp := make([]byte, 10)
		c.conn.Read(resp)

		res, _ := db.Exec("DELETE FROM codes WHERE ownerid = ? AND code = ?", c.user_id, BytesToString(resp))
		if ra, _ := res.RowsAffected(); ra != 1 {
			c.conn.Write(CreateResponse("Wrong code", Failure))
		} else {
			c.money_op("withdraw")
		}
	case "deposit":
		c.money_op("deposit")
	default:
		return fmt.Errorf("Request fallthrough (bad request: %s)", sreq)
	}
	return nil
}

func (c *ClientConn) get_balance() int {
	var balance string
	db.QueryRow("SELECT balance FROM users WHERE id = ?", c.user_id).Scan(&balance)
	i, _ := strconv.Atoi(balance)
	return i
}

func (c *ClientConn) money_op(op string) {
	amount := make([]byte, 10)
	c.conn.Write(CreateResponse("get_amnt", Request))
	c.conn.Read(amount)
	new_balance, _ := strconv.Atoi(BytesToString(amount))
	switch op {
	case "withdraw":
		if new_balance > c.get_balance() {
			c.conn.Write(CreateResponse("Insufficient funds", Failure))
			return
		} 
		new_balance = c.get_balance() - new_balance
	case "deposit":
		new_balance += c.get_balance()
	}
	err := dbexec(db, fmt.Sprintf("UPDATE users SET balance = %d WHERE id = %s", new_balance, c.user_id))
	if err != nil {
		c.conn.Write(CreateResponse("Updating balance failed", Failure))
	} else {
		c.conn.Write(CreateResponse("Balance updated", Success))
	}	
}