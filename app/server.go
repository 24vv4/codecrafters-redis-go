package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 256)
		if _, err := conn.Read(buf); err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("error reading from client: ", err.Error())
				os.Exit(1)
			}
		}
		tokens := tokenizer(buf)
		res := execCommand(tokens)
		conn.Write(res)
	}
}

func tokenizer(command []byte) []string {
	var res []string
	switch command[0] {
	case '+':
		// Simple String
		s := string(command[1 : len(command)-2]) // remove \r\n
		res = append(res, s)
	case '*':
		// Arrays
		// ex. *10\r\n...
		next_cr := findCR(command)
		count, _ := strconv.Atoi(string(command[1:next_cr]))
		token_len := 0
		start := next_cr + 2
		for i := 0; i < count; i++ {
			tokens := tokenizer(command[start:])
			for _, token := range tokens {
				res = append(res, token)
				token_len += len(token)
			}
			start += 5 // $ + \r\n + \r\n
			start += token_len
			start += len(strconv.Itoa(token_len))
		}
	case '$':
		// Bulk String
		// ex. $11\r\nHello,World\r\n
		next_cr := findCR(command)
		length, _ := strconv.Atoi(string((command[1:next_cr])))
		res = append(res, string(command[next_cr+2:next_cr+2+length]))
	}
	return res
}

func execCommand(tokens []string) []byte {
	var res string
	switch tokens[0] {
	case "ping":
		res = "+PONG\r\n"
	case "echo":
		res = "$" + strconv.Itoa(len(tokens[1])) + "\r\n" + tokens[1] + "\r\n"
	}
	return []byte(res)
}

func findCR(command []byte) int {
	for j, b := range command {
		if b == '\r' {
			return j
		}
	}
	return -1
}
