package tools

import (
	"bytes"
	"unicode"
)

// Creates a response with given status byte
func CreateResponse(msg string, status byte) []byte {
	b_msg := []byte(msg)
	return bytes.Join([][]byte{{status},b_msg}, []byte{})
}

// converts a byte array to a string, trimming trailing 0's
func BytesToString(b []byte) string {
	return string(bytes.TrimRight(b, string([]byte{0})))
}

// Capitalize first letter in string
func Capitalize(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}