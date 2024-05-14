package main

import (
	"bytes"
	"fmt"
	"strconv"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer conn.Close()

	// Read the request
	req := make([]byte, 1024)

	_, err = conn.Read(req)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		os.Exit(1)
	}

	// log the request
	fmt.Println(string(req))

	reqLines := bytes.Split(req, []byte("\r\n"))
	reqWords := bytes.Split(reqLines[0], []byte(" "))
	path := bytes.Split(reqWords[1], []byte("/"))

	statusLine := make([]byte, 0)
	var headers  bytes.Buffer
	body := make([]byte, 0)

	if bytes.Equal(path[1], []byte("echo")) && len(path) == 3 {
		statusLine = append(statusLine, []byte("HTTP/1.1 200 OK\r\n")...)

		headers.Write([]byte("Content-Type: text/plain\r\n"))
		headers.Write([]byte("Content-Length: "))
		headers.Write([]byte(strconv.Itoa(len(path[2]))))
		headers.Write([]byte("\r\n"))

		body = append(body, path[2]...)
	} else if bytes.Equal(reqWords[1], []byte("/")) {
		statusLine = append(statusLine, []byte("HTTP/1.1 200 OK\r\n\r\n")...)
	} else {
		statusLine = append(statusLine, []byte("HTTP/1.1 404 Not Found\r\n\r\n")...)
	}



	// bundle status line, headers and body into single response object
	res := append(statusLine, headers.Bytes()...)
	res = append(res, []byte("\r\n")...)
	res = append(res, body...)

	_, err = conn.Write(res)
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}
