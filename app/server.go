package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	// Uncomment this block to pass the first stage
	"flag"
	"log"
	"net"
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

func parseValidEncoding(header string) []string {
	// parse the accept-encoding header and return a slice of valid encodings
	// valid encodings are gzip and brotli
	// if the header is empty, return an empty slice
	// if the header contains an invalid encoding, ignore it
	// if the header contains a valid encoding multiple times, return it only once

	encodings := strings.Split(header, ",")
	encodingSet := map[string]bool{}
	var validEncodings []string
	for i, encoding := range encodings {
		encodings[i] = strings.TrimSpace(encoding)
		if encodings[i] == "gzip" || encodings[i] == "brotli" {
			encodingSet[encodings[i]] = true
		}
	}
	for encoding := range encodingSet {
		validEncodings = append(validEncodings, encoding)
	}
	return validEncodings
}

func responseNotFound(conn net.Conn) {
	res := "HTTP/1.1 404 Not Found\r\n\r\n"
	_, err := conn.Write([]byte(res))
	if err != nil {
		log.Fatalln("Error writing to connection: ", err.Error())
	}

}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func handleConn(conn net.Conn, dir string) {

	defer conn.Close()

	// Read the request
	req := make([]byte, 1024)

	_, err := conn.Read(req)
	if err != nil {
		log.Fatalln("Error reading request: ", err.Error())
	}

	// log the request
	log.Println(string(req))

	reqLines := bytes.Split(req, []byte("\r\n"))
	reqWords := bytes.Split(reqLines[0], []byte(" "))
	method := reqWords[0]
	path := bytes.Split(reqWords[1], []byte("/"))
	reqHeaders := parseReqHeaders(reqLines[1 : len(reqLines)-2])

	var res strings.Builder

	log.Println(reqHeaders, len(reqHeaders))
	log.Println(parseValidEncoding(reqHeaders["accept-encoding"]), len(parseValidEncoding(reqHeaders["accept-encoding"])))
	log.Printf("request path: %s", path[1])

	switch {
	case bytes.Equal(path[1], []byte("echo")) && len(path) == 3:

		res.WriteString("HTTP/1.1 200 OK\r\n")

		res.WriteString("Content-Type: text/plain\r\n")
		acceptedEncoding := parseValidEncoding(reqHeaders["accept-encoding"])
		if acceptedEncoding != nil && len(acceptedEncoding) == 1 {
			res.WriteString("Content-Encoding: " + acceptedEncoding[0] + "\r\n")

			var b bytes.Buffer
			gz := gzip.NewWriter(&b)
			if _, err := gz.Write([]byte(path[2])); err != nil {
				log.Fatal(err)
			}
			if err = gz.Close(); err != nil {
				log.Fatal(err)
			}
			res.WriteString("Content-Length: " + fmt.Sprintf("%d", len(b.String())) + "\r\n\r\n")
			res.WriteString(b.String())
		} else {
			res.WriteString("Content-Length: " + fmt.Sprintf("%d", len(path[2])) + "\r\n\r\n")
			res.WriteString(string(path[2]))
		}

	case bytes.Equal(path[1], []byte("user-agent")) && len(path) == 2:
		res.WriteString("HTTP/1.1 200 OK\r\n")

		res.WriteString("Content-Type: text/plain\r\n")
		res.WriteString("Content-Length: " + fmt.Sprintf("%d", len(reqHeaders["user-agent"])) + "\r\n\r\n")

		res.WriteString(reqHeaders["user-agent"])
	case bytes.Equal(method, []byte("GET")) && bytes.Equal(path[1], []byte("files")) && len(path) == 3:

		file := dir + string(path[2])
		_, err = os.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			responseNotFound(conn)
		} else if err != nil {
			panic(err)
		}

		dat, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		res.WriteString("HTTP/1.1 200 OK\r\n")

		res.WriteString("Content-Type: application/octet-stream\r\n")
		res.WriteString("Content-Length: " + fmt.Sprintf("%d", len(dat)) + "\r\n\r\n")

		res.WriteString(string(dat))
	case bytes.Equal(method, []byte("POST")) && bytes.Equal(path[1], []byte("files")) && len(path) == 3:
		contentLen, err := strconv.Atoi(reqHeaders["content-length"])
		checkErr(err)
		reqBody := reqLines[len(reqLines)-1][:contentLen]

		f, err := os.Create(dir + string(path[2]))
		checkErr(err)

		defer f.Close()

		log.Println(reqBody)

		f.Write(reqBody)
		f.Sync()

		res.WriteString("HTTP/1.1 201 Created\r\n\r\n")

	case bytes.Equal(reqWords[1], []byte("/")):
		res.WriteString("HTTP/1.1 200 OK\r\n\r\n")
	default:
		responseNotFound(conn)
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

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	log.Println("Logs from your program will appear here!")

	dir := flag.String("directory", "", "Directory to serve files from")
	flag.Parse()

	if *dir != "" {
		log.Println("Serving files from directory:", *dir)
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatalln("Failed to bind to port 4221")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("Error accepting connection: ", err.Error())
		}

		go handleConn(conn, *dir)
	}

}
