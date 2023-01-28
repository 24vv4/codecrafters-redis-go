package main

import (
	"fmt"
	"net"
	"os"
    "sync"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
	    fmt.Println("Failed to bind to port 6379")
	    os.Exit(1)
	}
    wg := &sync.WaitGroup{}
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            conn, err := l.Accept()
            if err != nil {
                fmt.Println("Error accepting connection: ", err.Error())
                os.Exit(1)
            }

            defer conn.Close()

            for {
                buf := make([]byte, 1024)
                if _, err := conn.Read(buf); err != nil {
                    fmt.Println("error reading from client: ", err.Error())
                    break;
                }
                conn.Write([]byte("+PONG\r\n"))
            }
            wg.Done()
        }()
    }
    wg.Wait()
}
