package main

import (
	"bufio"
	"errors"
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
    //TODO: thread safe
    m := NewMemory()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, m)
	}
}

func handleConnection(conn net.Conn, m *Memory) {
	defer conn.Close()
	for {
		rd := bufio.NewReader(conn)
		buf := make([]byte, 1024)
		if _, err := rd.Read(buf); err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				fmt.Println("error reading from client: ", err.Error())
				os.Exit(1)
			}
		}
		tokens := tokenizer(buf)
		res := execCommand(tokens, m)
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
		start := next_cr + 2
		for i := 0; i < count; i++ {
			tokens := tokenizer(command[start:])
			token_len := 0
			for _, token := range tokens {
				res = append(res, token)
				token_len += len(token)
			}
			// TODO: this logic supports only Bulk String
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

func execCommand(tokens []string, memory *Memory) []byte {
	var res string
	switch tokens[0] {
	case "ping":
		res = "+PONG\r\n"
	case "echo":
		res = "$" + strconv.Itoa(len(tokens[1])) + "\r\n" + tokens[1] + "\r\n"
	case "set":
        if(len(tokens) > 3) {
            // PX
            ms, _ := strconv.ParseInt(tokens[4], 10, 64)
            memory.SetPX(tokens[1], tokens[2], ms)
        } else {
		    memory.Set(tokens[1], tokens[2])
        }
		res = "+OK\r\n"
	case "get":
        s, ok := memory.Get(tokens[1])
		if(ok) {
            res = "+" + s + "\r\n"
        } else {
            res = "$-1\r\n"
        }
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
