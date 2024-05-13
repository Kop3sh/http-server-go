package main

import (
	"bytes"
	"fmt"

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

	res := []byte("HTTP/1.1 200 OK\r\n\r\n")

	if !bytes.Equal(reqWords[1], []byte("/")) {
		res = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	}


	_, err = conn.Write(res)
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
}
