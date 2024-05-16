package main

import (
	"bytes"
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"log"
	"net"
	"os"
)

func parseReqHeaders(reqHeaders [][]byte) map[string]string {
	var headerMap map[string]string = make(map[string]string)
	for _, l := range reqHeaders {
		field := bytes.Split(l, []byte(":"))

		// make sure to lowercase the field name, and to trim field value
		headerMap[strings.ToLower(string(field[0]))] = strings.TrimSpace(string(field[1]))
	}

	return headerMap
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	log.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatalln("Failed to bind to port 4221")
	}

	conn, err := l.Accept()
	if err != nil {
		log.Fatalln("Error accepting connection: ", err.Error())
	}

	defer conn.Close()

	// Read the request
	req := make([]byte, 1024)

	_, err = conn.Read(req)
	if err != nil {
		log.Fatalln("Error reading request: ", err.Error())
	}

	// log the request
	log.Println(string(req))

	reqLines := bytes.Split(req, []byte("\r\n"))
	reqWords := bytes.Split(reqLines[0], []byte(" "))
	path := bytes.Split(reqWords[1], []byte("/"))
	reqHeaders := parseReqHeaders(reqLines[1 : len(reqLines)-2])

	var res strings.Builder

	log.Println(reqHeaders, len(reqHeaders))
	log.Printf("%s", path[1])

	if bytes.Equal(path[1], []byte("echo")) && len(path) == 3 {

		res.WriteString("HTTP/1.1 200 OK\r\n")

		res.WriteString("Content-Type: text/plain\r\n")
		res.WriteString("Content-Length: " + fmt.Sprintf("%d", len(path[2])) + "\r\n\r\n")

		res.WriteString(string(path[2]))
	} else if bytes.Equal(path[1], []byte("user-agent")) && len(path) == 2 {
		res.WriteString("HTTP/1.1 200 OK\r\n")

		res.WriteString("Content-Type: text/plain\r\n")
		res.WriteString("Content-Length: " + fmt.Sprintf("%d", len(reqHeaders["user-agent"])) + "\r\n\r\n")

		res.WriteString(reqHeaders["user-agent"])
	} else if bytes.Equal(reqWords[1], []byte("/")) {
		res.WriteString("HTTP/1.1 200 OK\r\n\r\n")
	} else {
		res.WriteString("HTTP/1.1 404 Not Found\r\n\r\n")
	}

	// bundle status line, headers and body into single response object
	// var res strings.Builder
	// res.WriteString(statusLine.String())
	// res.WriteString(headers.String())
	// res.WriteString(body.String())

	_, err = conn.Write([]byte(res.String()))
	if err != nil {
		log.Fatalln("Error writing to connection: ", err.Error())
	}
}
