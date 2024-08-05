package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

type requestLine struct {
	method      []byte
	url         []byte
	httpVersion []byte
}

func newRequestLine() *requestLine {
	return &requestLine{
		method:      make([]byte, 16),
		url:         make([]byte, 512),
		httpVersion: make([]byte, 16),
	}
}

type httpMessage struct {
	startLine   requestLine
	headers     map[string]string
	messageBody []byte
}

func newHttpMessage() *httpMessage {

	return &httpMessage{
		startLine:   *newRequestLine(),
		headers:     make(map[string]string),
		messageBody: make([]byte, 0),
	}
}

func parseStatusLine(statLine string) []string {
	componenets := strings.Split(statLine, " ")
	if len(componenets) != 3 {
		panic("Invalid Header")
	}
	return componenets
}

func (h *httpMessage) parseRequest(request []byte) error {
	if len(request) == 0 {
		return fmt.Errorf("request is empty")
	}
	scanner := bufio.NewScanner(strings.NewReader(string(request)))
	scanner.Split(bufio.ScanLines)
	isBody := false
	isHeader := false
	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			isBody = true
			continue
		}
		if !isHeader && (strings.HasPrefix(scanner.Text(), "GET") || strings.HasPrefix(scanner.Text(), "POST")) {
			statTemp := parseStatusLine(scanner.Text())
			h.startLine.method = []byte(statTemp[0])
			h.startLine.url = []byte(statTemp[1])
			h.startLine.httpVersion = []byte(statTemp[2])
			isHeader = true
			continue
		}
		if isBody {
			h.messageBody = append(h.messageBody, scanner.Bytes()...)
			h.messageBody = bytes.Trim(h.messageBody, "\x00")
			continue
		}

		fieldKey, fieldValue, _ := strings.Cut(scanner.Text(), ":")

		fieldKey = strings.Trim(fieldKey, " ")
		fieldValue = strings.Trim(fieldValue, " ")
		h.headers[fieldKey] = fieldValue
	}
	return nil
}

func (req *httpMessage) getFileRequest(conn net.Conn) error {
	// ^/files/(?P<file>\w+.?\w+$)
	r := regexp.MustCompile(`^/files/(?P<file>\w+.?\w+$)`)
	res := r.FindStringSubmatch(string(req.startLine.url))

	args := os.Args
	currentDir := false
	if len(args) != 3 {
		currentDir = true
	}
	var fileData []byte
	var err error
	if currentDir {
		fileData, err = os.ReadFile(fmt.Sprintf("%s%s", "./", res[1]))
	} else {
		dir := args[2]
		fileData, err = os.ReadFile(fmt.Sprintf("%s%s", dir, res[1]))
	}

	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		conn.Close()
		return err
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(fileData), fileData)
	conn.Write([]byte(response))
	conn.Close()
	return nil
}

func (req *httpMessage) postFileRequest(conn net.Conn) error {
	r := regexp.MustCompile(`^/files/(?P<file>\w+.?\w+$)`)
	res := r.FindStringSubmatch(string(req.startLine.url))

	args := os.Args
	currentDir := false
	if len(args) != 3 {
		currentDir = true
	}
	var fd *os.File
	var err error
	if currentDir {
		fd, err = os.Create(fmt.Sprintf("%s%s", "./", res[1]))
	} else {
		dir := args[2]
		fd, err = os.Create(fmt.Sprintf("%s%s", dir, res[1]))
	}
	if err != nil {
		conn.Write([]byte("HTPP/1.1 500 Internal Server error"))
		conn.Close()
		return err
	}
	n, _ := fd.Write(req.messageBody)
	fmt.Println("Bytes written : ", n)
	response := "HTTP/1.1 201 Created\r\n\r\n"
	conn.Write([]byte(response))
	conn.Close()
	return nil

}

func (req *httpMessage) responseHanldle(conn net.Conn) error {
	if string(req.startLine.url) == "/" {

		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		conn.Close()
		return nil

	} else if ok, _ := regexp.Match(`^/echo/\w+$`, req.startLine.url); ok {

		r := regexp.MustCompile(`^/echo/(?P<echo>\w+)$`)
		res := r.FindStringSubmatch(string(req.startLine.url))
		// index 1 contains the value of str in /echo/{str}

		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(res[1]), res[1])
		conn.Write([]byte(response))
		conn.Close()
		return nil

	} else if string(req.startLine.url) == "/user-agent" {

		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(req.headers["User-Agent"]), req.headers["User-Agent"])
		conn.Write([]byte(response))
		conn.Close()
		return nil

	} else if ok, _ := regexp.Match(`^/files/\w+.?\w+$`, req.startLine.url); ok {
		if string(req.startLine.method) == "GET" {
			err := req.getFileRequest(conn)
			if err != nil {
				return err
			}
			return nil
		} else if string(req.startLine.method) == "POST" {
			err := req.postFileRequest(conn)
			if err != nil {
				return err
			}
			return nil
		}
	}

	_, err := conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	if err != nil {
		fmt.Println("Error : ", err.Error())
		conn.Close()
		return err
	}
	conn.Close()
	return nil
}

func AcceptConn(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go requestHandle(conn)
	}
}

func requestHandle(conn net.Conn) {
	var request = make([]byte, 2048)
	conn.Read(request)
	req := newHttpMessage()
	req.parseRequest(request)
	err := req.responseHanldle(conn)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	fmt.Println("Server has started!!!")
	err := AcceptConn("0.0.0.0:4221")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}
